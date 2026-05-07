package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"awo.so/internal/core/audit"
	"awo.so/internal/shared"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
)

// Finance audit event types — EventType values in audit.CreateAuditEventRequest.
const (
	auditTypeTxnCreated       = "FINANCE_TXN_CREATED"
	auditTypeTxnPosted        = "FINANCE_TXN_POSTED"
	auditTypeTxnApproved      = "FINANCE_TXN_APPROVED"
	auditTypeTxnRejected      = "FINANCE_TXN_REJECTED"
	auditTypeTxnReversed      = "FINANCE_TXN_REVERSED"
	auditTypeRecurringCreated = "FINANCE_RECURRING_TXN_CREATED"
	auditTypePeriodStatus     = "FINANCE_PERIOD_STATUS_CHANGED"
	auditTypeStmtImported     = "FINANCE_BANK_STMT_IMPORTED"
	auditTypeReconciled       = "FINANCE_RECONCILIATION_COMPLETED"
	auditTypeLineMatched      = "FINANCE_STMT_LINE_MATCHED"
	auditTypeLineUnmatched    = "FINANCE_STMT_LINE_UNMATCHED"
)

const (
	auditCategoryFinance = "FINANCE"
	auditSeverityInfo    = "INFO"
	auditSeverityMedium  = "MEDIUM"
	auditSeverityHigh    = "HIGH"
)

// ── financeAuditWriter ────────────────────────────────────────────────────────

// financeAuditWriter routes finance audit events by their criticality level:
//
//   - CRITICAL → durable outbox (same DB as mutations) → worker delivery
//   - IMPORTANT → direct write with bounded retry
//   - LOW       → direct write, errors suppressed at WARN
//
// All methods are nil-safe: a nil *financeAuditWriter silently no-ops. This
// allows tests to pass nil without any additional mock setup.
type financeAuditWriter struct {
	auditSvc   audit.Service
	outboxRepo AuditOutboxRepository // nil → degraded mode (no durable outbox)
	metrics    metrics.MetricsProvider
}

// newFinanceAuditWriter constructs the writer. Returns nil when auditSvc is nil
// (safe: all write methods check for nil receiver).
func newFinanceAuditWriter(svc audit.Service, outboxRepo AuditOutboxRepository, m metrics.MetricsProvider) *financeAuditWriter {
	if svc == nil {
		return nil
	}
	return &financeAuditWriter{auditSvc: svc, outboxRepo: outboxRepo, metrics: m}
}

// writeAudit routes an audit event by its criticality.
// idempotencyKey is used for CRITICAL outbox entries to prevent duplicate
// delivery on Temporal retries. Pass empty string for IMPORTANT/LOW events.
func (w *financeAuditWriter) writeAudit(ctx context.Context, req audit.CreateAuditEventRequest, idempotencyKey string) {
	if w == nil {
		return
	}
	level := financeEventCriticality[req.EventType]
	if level == 0 {
		level = AuditLow
	}
	switch level {
	case AuditCritical:
		w.writeCritical(ctx, req, idempotencyKey)
	case AuditImportant:
		w.writeImportant(ctx, req)
	default:
		w.writeLow(ctx, req)
	}
}

// writeOutboxInTx writes a CRITICAL audit entry to the outbox within the
// current DB transaction (via txCtx from RunInTx).
//
// Caller semantics:
//   - When outboxRepo IS configured: returns the outbox write error.
//     The caller MUST return this error from RunInTx so the transaction rolls back.
//     This guarantees the audit record is atomic with the financial mutation.
//   - When outboxRepo is NOT configured (degraded): returns nil, logs at WARN.
//     The financial mutation proceeds; audit durability is degraded but not blocked.
//
// Use this ONLY for CRITICAL events on the atomic (txRunner) path.
func (w *financeAuditWriter) writeOutboxInTx(ctx context.Context, req audit.CreateAuditEventRequest, idempotencyKey string) error {
	if w == nil {
		return nil
	}
	if w.outboxRepo == nil {
		logger.WarnContext(ctx, "CRITICAL audit outbox not configured — audit trail at risk for this mutation", logger.Fields{
			"event_type": req.EventType,
		})
		return nil // degraded: mutation proceeds, audit durability not guaranteed
	}
	return w.writeToOutbox(ctx, req, idempotencyKey)
}

// ── internal write paths ──────────────────────────────────────────────────────

func (w *financeAuditWriter) writeCritical(ctx context.Context, req audit.CreateAuditEventRequest, idempotencyKey string) {
	if w.outboxRepo != nil {
		if err := w.writeToOutbox(ctx, req, idempotencyKey); err != nil {
			w.metrics.IncrementCounter("finance_audit_outbox_write_failures_total", metrics.Fields{
				"event_type": req.EventType,
			})
			logger.ErrorContext(ctx, "CRITICAL: audit outbox write failed — audit trail at risk; attempting direct fallback", logger.Fields{
				"event_type":      req.EventType,
				"idempotency_key": idempotencyKey,
				"error":           err.Error(),
			})
			// Fallback: direct write may still succeed and leaves a trace.
			w.directWrite(ctx, req)
		}
		return
	}
	// Degraded mode: outbox not configured. Log escalation and attempt direct write.
	logger.WarnContext(ctx, "CRITICAL audit written without durable outbox — configure AuditOutboxRepo for forensic guarantees", logger.Fields{
		"event_type": req.EventType,
	})
	w.directWrite(ctx, req)
}

func (w *financeAuditWriter) writeImportant(ctx context.Context, req audit.CreateAuditEventRequest) {
	var lastErr error
	for i := 0; i <= MaxImportantRetries; i++ {
		if _, err := w.auditSvc.CreateAuditEvent(ctx, req); err == nil {
			return
		} else {
			lastErr = err
		}
	}
	w.metrics.IncrementCounter("finance_audit_important_failures_total", metrics.Fields{
		"event_type": req.EventType,
	})
	logger.ErrorContext(ctx, "IMPORTANT audit event failed after retries — event lost", logger.Fields{
		"event_type":  req.EventType,
		"retry_count": MaxImportantRetries,
		"last_error":  lastErr.Error(),
	})
}

func (w *financeAuditWriter) writeLow(ctx context.Context, req audit.CreateAuditEventRequest) {
	if _, err := w.auditSvc.CreateAuditEvent(ctx, req); err != nil {
		logger.WarnContext(ctx, "LOW audit event suppressed", logger.Fields{
			"event_type": req.EventType,
			"error":      err.Error(),
		})
	}
}

func (w *financeAuditWriter) writeToOutbox(ctx context.Context, req audit.CreateAuditEventRequest, idempotencyKey string) error {
	tenantID, _ := shared.GetTenantID(ctx)
	payload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal audit payload: %w", err)
	}
	entry := &AuditOutboxEntry{
		ID:             uuid.New(),
		TenantID:       tenantID,
		IdempotencyKey: idempotencyKey,
		EventType:      req.EventType,
		Payload:        payload,
		Status:         AuditOutboxPending,
		MaxRetries:     MaxOutboxRetries,
		CreatedAt:      time.Now(),
	}
	return w.outboxRepo.WriteOutbox(ctx, entry)
}

func (w *financeAuditWriter) directWrite(ctx context.Context, req audit.CreateAuditEventRequest) {
	if _, err := w.auditSvc.CreateAuditEvent(ctx, req); err != nil {
		w.metrics.IncrementCounter("finance_audit_direct_write_failures_total", metrics.Fields{
			"event_type": req.EventType,
		})
		logger.ErrorContext(ctx, "finance audit direct write failed", logger.Fields{
			"event_type": req.EventType,
			"error":      err.Error(),
		})
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// auditUserID extracts the caller's user ID for audit events.
// Returns nil when no authenticated user is in context (system calls).
func auditUserID(ctx context.Context) *uuid.UUID {
	id, ok := shared.GetUserID(ctx)
	if !ok || id == uuid.Nil {
		return nil
	}
	return &id
}

// auditCtx marshals arbitrary fields into JSON for the audit Context field.
// Returns nil on marshal error — audit still fires, just without context payload.
func auditCtx(fields map[string]any) json.RawMessage {
	b, err := json.Marshal(fields)
	if err != nil {
		return nil
	}
	return b
}

// uuidPtr returns a pointer to v. Returns nil for uuid.Nil.
func uuidPtr(v uuid.UUID) *uuid.UUID {
	if v == uuid.Nil {
		return nil
	}
	return &v
}

// derefUUIDPtr dereferences a *uuid.UUID safely. Returns nil for nil or uuid.Nil.
func derefUUIDPtr(p *uuid.UUID) *uuid.UUID {
	if p == nil || *p == uuid.Nil {
		return nil
	}
	return p
}
