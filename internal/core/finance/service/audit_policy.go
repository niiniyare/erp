package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"awo.so/internal/core/audit"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
)

// ── Criticality Classification ────────────────────────────────────────────────

// AuditCriticality defines the required durability level for a finance audit event.
type AuditCriticality int

const (
	// AuditCritical events must never be silently lost.
	//
	// Write path (atomic txRunner operations):
	//   Outbox written inside the DB transaction. If outbox write fails, the
	//   financial mutation rolls back and the caller retries. Guarantees that
	//   the audit record and financial state are always in sync.
	//
	// Write path (non-atomic operations):
	//   Outbox written immediately after the successful repo call. There is a
	//   narrow window if the process crashes between the two writes; the gap
	//   detector surfaces such events for manual resolution.
	//
	// Delivery: background worker (Temporal cron) → audit.Service.
	// On failure: retry up to MaxOutboxRetries, then DEAD state + alert.
	AuditCritical AuditCriticality = iota + 1

	// AuditImportant events are preferred but a single lost event is tolerable.
	// Written directly to audit.Service with up to MaxImportantRetries attempts.
	// On sustained failure: logged at ERROR; event is not queued for retry.
	AuditImportant

	// AuditLow events are diagnostic/operational only.
	// Single direct write; errors logged at WARN and suppressed.
	AuditLow
)

// financeEventCriticality maps every finance audit event type to its required
// durability. Unlisted types default to AuditLow.
var financeEventCriticality = map[string]AuditCriticality{
	// ── CRITICAL: forensic evidence, must not lose ────────────────────────────
	// Ledger entry created — reversal traceability depends on this record.
	auditTypeTxnPosted: AuditCritical,
	// Approval evidence — segregation-of-duties accountability.
	auditTypeTxnApproved: AuditCritical,
	// Rejection evidence — same accountability requirement as approval.
	auditTypeTxnRejected: AuditCritical,
	// Reversal chain — must be traceable for every posted reversal.
	auditTypeTxnReversed: AuditCritical,
	// Period-close evidence — required for period-close sign-off and compliance.
	auditTypePeriodStatus: AuditCritical,
	// Reconciliation sign-off — required for bank reconciliation audit trail.
	auditTypeReconciled: AuditCritical,

	// ── IMPORTANT: preferred, bounded retry, tolerable loss ──────────────────
	// Transaction draft creation — useful for forensics but not primary evidence.
	auditTypeTxnCreated: AuditImportant,
	// Recurring generation — operational; idempotent by transaction number.
	auditTypeRecurringCreated: AuditImportant,
	// Statement import — useful for reconciliation history.
	auditTypeStmtImported: AuditImportant,

	// ── LOW: diagnostic, safe to suppress ────────────────────────────────────
	auditTypeLineMatched:   AuditLow,
	auditTypeLineUnmatched: AuditLow,
}

const (
	// MaxImportantRetries is the number of direct-write attempts for IMPORTANT events
	// before the failure is logged and the event is abandoned.
	MaxImportantRetries = 2

	// MaxOutboxRetries is the number of delivery attempts for an outbox entry before
	// it transitions to DEAD state and triggers an alert.
	MaxOutboxRetries = 5

	// OutboxStaleDuration is the age beyond which a PENDING outbox entry is considered
	// stale (delivery lag) by the gap detector.
	OutboxStaleDuration = 10 * time.Minute
)

// ── Outbox Domain Types ───────────────────────────────────────────────────────

// AuditOutboxStatus tracks the delivery lifecycle of a finance audit outbox entry.
type AuditOutboxStatus string

const (
	AuditOutboxPending    AuditOutboxStatus = "PENDING"
	AuditOutboxProcessing AuditOutboxStatus = "PROCESSING"
	AuditOutboxDelivered  AuditOutboxStatus = "DELIVERED"
	AuditOutboxDead       AuditOutboxStatus = "DEAD"
)

// AuditOutboxEntry is a CRITICAL finance audit event awaiting durable delivery
// to audit.Service by the background outbox worker.
type AuditOutboxEntry struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`

	// IdempotencyKey prevents duplicate delivery on Temporal/worker retries.
	// Format: "{event_type}:{primary_resource_id}[:{qualifier}]"
	// Example: "FINANCE_TXN_REVERSED:550e8400-e29b-41d4-a716-446655440000"
	IdempotencyKey string `json:"idempotency_key"`

	EventType string          `json:"event_type"`
	Payload   json.RawMessage `json:"payload"` // JSON-encoded audit.CreateAuditEventRequest

	Status     AuditOutboxStatus `json:"status"`
	RetryCount int               `json:"retry_count"`
	MaxRetries int               `json:"max_retries"`

	CreatedAt   time.Time  `json:"created_at"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
	ErrorMsg    *string    `json:"error_msg,omitempty"`
}

// ── Repository Interface ──────────────────────────────────────────────────────

// AuditOutboxRepository persists and manages the finance audit outbox.
type AuditOutboxRepository interface {
	// WriteOutbox inserts a PENDING entry. The context may carry a DB transaction —
	// when it does, the write is atomic with the enclosing financial mutation.
	// Idempotent: silently succeeds if idempotency_key already exists for tenant.
	WriteOutbox(ctx context.Context, entry *AuditOutboxEntry) error

	// ListPending returns up to limit PENDING entries ordered by created_at ASC.
	// Called by the outbox worker on each delivery cycle.
	ListPending(ctx context.Context, limit int) ([]*AuditOutboxEntry, error)

	// MarkDelivered transitions entry to DELIVERED and sets processed_at.
	MarkDelivered(ctx context.Context, id uuid.UUID) error

	// MarkFailed increments retry_count and stores errMsg.
	// Transitions to DEAD when retry_count exceeds entry.MaxRetries.
	MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error

	// CountDead returns the number of DEAD entries for a tenant (for alerting).
	CountDead(ctx context.Context, tenantID uuid.UUID) (int, error)

	// CountStalePending returns the count of PENDING entries older than staleAfter.
	// Used by the gap detector to surface delivery lag.
	CountStalePending(ctx context.Context, staleAfter time.Duration) (int, error)
}

// ── Outbox Processor ─────────────────────────────────────────────────────────

// AuditOutboxProcessor delivers pending outbox entries to audit.Service.
// Intended to be called from a Temporal cron activity on a short interval (e.g. 30s).
type AuditOutboxProcessor struct {
	outboxRepo AuditOutboxRepository
	auditSvc   audit.Service
	metrics    metrics.MetricsProvider
}

// NewAuditOutboxProcessor creates the processor used by the Temporal delivery worker.
func NewAuditOutboxProcessor(
	outboxRepo AuditOutboxRepository,
	auditSvc audit.Service,
	m metrics.MetricsProvider,
) *AuditOutboxProcessor {
	return &AuditOutboxProcessor{outboxRepo: outboxRepo, auditSvc: auditSvc, metrics: m}
}

// ProcessBatch delivers up to batchSize PENDING entries. Returns after processing
// all fetched entries regardless of individual delivery outcomes.
// Safe to call concurrently — idempotency_key prevents duplicate delivery.
func (p *AuditOutboxProcessor) ProcessBatch(ctx context.Context, batchSize int) error {
	entries, err := p.outboxRepo.ListPending(ctx, batchSize)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		p.deliver(ctx, entry)
	}
	return nil
}

func (p *AuditOutboxProcessor) deliver(ctx context.Context, entry *AuditOutboxEntry) {
	var req audit.CreateAuditEventRequest
	if err := json.Unmarshal(entry.Payload, &req); err != nil {
		errMsg := "payload unmarshal: " + err.Error()
		_ = p.outboxRepo.MarkFailed(ctx, entry.ID, errMsg)
		logger.ErrorContext(ctx, "audit outbox: payload unmarshal failed — entry marked failed", logger.Fields{
			"outbox_id":  entry.ID.String(),
			"event_type": entry.EventType,
			"error":      errMsg,
		})
		return
	}

	if _, err := p.auditSvc.CreateAuditEvent(ctx, req); err != nil {
		errMsg := err.Error()
		_ = p.outboxRepo.MarkFailed(ctx, entry.ID, errMsg)
		p.metrics.IncrementCounter("finance_audit_outbox_delivery_failures_total", metrics.Fields{
			"event_type": entry.EventType,
		})
		logger.ErrorContext(ctx, "audit outbox: delivery failed", logger.Fields{
			"outbox_id":   entry.ID.String(),
			"event_type":  entry.EventType,
			"retry_count": entry.RetryCount,
			"error":       errMsg,
		})
		return
	}

	_ = p.outboxRepo.MarkDelivered(ctx, entry.ID)
	p.metrics.IncrementCounter("finance_audit_outbox_delivered_total", metrics.Fields{
		"event_type": entry.EventType,
	})
}
