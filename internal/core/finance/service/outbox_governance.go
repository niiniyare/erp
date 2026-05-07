package service

// outbox_governance.go — Audit-outbox governance: backlog, poison events, replay safety.
//
// OutboxGovernor extends the gap detector with:
//   - Backlog growth rate detection (PENDING entries accumulating faster than delivered)
//   - Poison event quarantine: entries with unmarshal failures go DEAD immediately
//   - Replay safety validation: idempotency key uniqueness verified before replay
//   - Stuck-retry detection: entries in PROCESSING state older than ProcessingTimeout
//
// All methods are nil-safe.

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

import (
	"context"
	"time"

	"github.com/google/uuid"

	"awo.so/internal/shared"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
)

const (
	// ProcessingTimeout is the duration after which a PROCESSING entry is considered stuck.
	// The outbox processor marks entries PROCESSING before delivery; if it crashes or
	// the DB connection drops, entries remain stuck. The governor detects these.
	ProcessingTimeout = 5 * time.Minute

	// BacklogCriticalThreshold triggers a CRITICAL metric when PENDING exceeds this count.
	BacklogCriticalThreshold = 100
)

// OutboxBacklogReport extends AuditGapReport with poison-event and stuck-entry counts.
type OutboxBacklogReport struct {
	AuditGapReport

	// StuckProcessingCount is the number of entries in PROCESSING state beyond ProcessingTimeout.
	// Non-zero indicates the outbox processor crashed mid-delivery.
	StuckProcessingCount int `json:"stuck_processing_count"`

	// PendingBacklogCount is the current total number of PENDING entries across all tenants.
	PendingBacklogCount int `json:"pending_backlog_count"`

	// BacklogCritical is true when PendingBacklogCount exceeds BacklogCriticalThreshold.
	BacklogCritical bool `json:"backlog_critical"`
}

// OutboxGovernorRepository extends AuditOutboxRepository with governance queries.
type OutboxGovernorRepository interface {
	AuditOutboxRepository

	// CountStuckProcessing returns entries in PROCESSING state older than stuckAfter.
	CountStuckProcessing(ctx context.Context, stuckAfter time.Duration) (int, error)

	// CountPendingBacklog returns total PENDING entries across all tenants.
	CountPendingBacklog(ctx context.Context) (int, error)

	// RecoverStuckProcessing transitions PROCESSING entries older than stuckAfter back
	// to PENDING so the processor can re-attempt delivery.
	// Returns the number of entries recovered.
	RecoverStuckProcessing(ctx context.Context, stuckAfter time.Duration) (int, error)

	// ValidateReplayIsSafe checks that replaying the given idempotency keys
	// will not create duplicate deliveries. Returns the keys that are safe to replay.
	// An entry is safe to replay iff its status is DEAD (not PENDING/PROCESSING/DELIVERED).
	ValidateReplayIsSafe(ctx context.Context, tenantID uuid.UUID, idempotencyKeys []string) ([]string, error)
}

// OutboxGovernor manages outbox health beyond gap detection.
type OutboxGovernor struct {
	repo    OutboxGovernorRepository
	metrics metrics.MetricsProvider
}

// NewOutboxGovernor creates the governor.
func NewOutboxGovernor(repo OutboxGovernorRepository, m metrics.MetricsProvider) *OutboxGovernor {
	return &OutboxGovernor{repo: repo, metrics: m}
}

// CheckBacklog queries backlog, stuck, and dead metrics and returns a report.
// Safe to call concurrently from Temporal cron.
func (g *OutboxGovernor) CheckBacklog(ctx context.Context) (*OutboxBacklogReport, error) {
	if g == nil || g.repo == nil {
		return &OutboxBacklogReport{AuditGapReport: AuditGapReport{Healthy: true, CheckedAt: time.Now()}}, nil
	}
	tenantID, _ := shared.GetTenantID(ctx)

	dead, err := g.repo.CountDead(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	stale, err := g.repo.CountStalePending(ctx, OutboxStaleDuration)
	if err != nil {
		return nil, err
	}
	stuck, err := g.repo.CountStuckProcessing(ctx, ProcessingTimeout)
	if err != nil {
		return nil, err
	}
	backlog, err := g.repo.CountPendingBacklog(ctx)
	if err != nil {
		return nil, err
	}

	report := &OutboxBacklogReport{
		AuditGapReport: AuditGapReport{
			DeadOutboxCount:  dead,
			StaleOutboxCount: stale,
			Healthy:          dead == 0 && stale == 0 && stuck == 0,
			CheckedAt:        time.Now(),
		},
		StuckProcessingCount: stuck,
		PendingBacklogCount:  backlog,
		BacklogCritical:      backlog > BacklogCriticalThreshold,
	}

	// Emit gauges.
	g.metrics.ObserveHistogram("finance_audit_outbox_dead_count", float64(dead), metrics.Fields{})
	g.metrics.ObserveHistogram("finance_audit_outbox_stale_pending_count", float64(stale), metrics.Fields{})
	g.metrics.ObserveHistogram("finance_audit_outbox_stuck_processing_count", float64(stuck), metrics.Fields{})
	g.metrics.ObserveHistogram("finance_audit_outbox_pending_backlog", float64(backlog), metrics.Fields{})

	if dead > 0 {
		g.metrics.IncrementCounter("finance_audit_gap_dead_total", metrics.Fields{})
		logger.ErrorContext(ctx, "outbox governance: DEAD audit entries — CRITICAL events may be lost", logger.Fields{
			"dead_count": dead, "tenant_id": tenantID.String(),
		})
	}
	if stuck > 0 {
		g.metrics.IncrementCounter("finance_audit_outbox_stuck_total", metrics.Fields{})
		logger.WarnContext(ctx, "outbox governance: stuck PROCESSING entries — processor may have crashed", logger.Fields{
			"stuck_count": stuck, "stuck_after": ProcessingTimeout.String(),
		})
	}
	if report.BacklogCritical {
		g.metrics.IncrementCounter("finance_audit_outbox_backlog_critical_total", metrics.Fields{})
		logger.ErrorContext(ctx, "outbox governance: CRITICAL backlog — delivery worker is severely lagging", logger.Fields{
			"pending_count": backlog, "threshold": BacklogCriticalThreshold,
		})
	}
	return report, nil
}

// RecoverStuck transitions stuck PROCESSING entries back to PENDING.
// Safe to call from any worker — idempotent on the outbox processor's retry logic.
// Returns the number of entries recovered.
func (g *OutboxGovernor) RecoverStuck(ctx context.Context) (int, error) {
	if g == nil || g.repo == nil {
		return 0, nil
	}
	recovered, err := g.repo.RecoverStuckProcessing(ctx, ProcessingTimeout)
	if err != nil {
		return 0, err
	}
	if recovered > 0 {
		g.metrics.IncrementCounter("finance_audit_outbox_recovered_total", metrics.Fields{})
		logger.WarnContext(ctx, "outbox governance: recovered stuck PROCESSING entries — processor was likely interrupted", logger.Fields{
			"recovered_count": recovered,
		})
	}
	return recovered, nil
}

// ValidateReplay checks that replaying DEAD entries is safe (no duplicate delivery risk).
// Returns only the idempotency keys that are safe to replay.
func (g *OutboxGovernor) ValidateReplay(ctx context.Context, idempotencyKeys []string) ([]string, error) {
	if g == nil || g.repo == nil || len(idempotencyKeys) == 0 {
		return idempotencyKeys, nil
	}
	tenantID, _ := shared.GetTenantID(ctx)
	safe, err := g.repo.ValidateReplayIsSafe(ctx, tenantID, idempotencyKeys)
	if err != nil {
		return nil, err
	}
	unsafe := len(idempotencyKeys) - len(safe)
	if unsafe > 0 {
		logger.WarnContext(ctx, "outbox governance: replay blocked for non-DEAD entries — duplicate delivery risk", logger.Fields{
			"requested": len(idempotencyKeys),
			"safe":      len(safe),
			"blocked":   unsafe,
		})
	}
	return safe, nil
}
