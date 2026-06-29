package service

import (
	"context"
	"time"

	"awo.so/internal/shared"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
)

// AuditGapReport summarises detected forensic blind spots in the finance audit trail.
type AuditGapReport struct {
	// DeadOutboxCount is the number of CRITICAL audit events that failed max
	// retries and are now in DEAD state. Each represents a potential audit gap
	// that requires manual intervention.
	DeadOutboxCount int `json:"dead_outbox_count"`

	// StaleOutboxCount is the number of PENDING entries older than OutboxStaleDuration.
	// Non-zero indicates the outbox worker is lagging or stopped.
	StaleOutboxCount int `json:"stale_outbox_count"`

	// Healthy is true only when both counts are zero.
	Healthy bool `json:"healthy"`

	CheckedAt time.Time `json:"checked_at"`
}

// AuditGapDetector surfaces forensic blind spots in the finance audit trail.
type AuditGapDetector struct {
	outboxRepo AuditOutboxRepository
	metrics    metrics.MetricsProvider
}

// NewAuditGapDetector creates a detector. Call CheckGaps on a scheduled basis
// (e.g. every 5 minutes from a Temporal cron) to keep metrics current.
func NewAuditGapDetector(outboxRepo AuditOutboxRepository, m metrics.MetricsProvider) *AuditGapDetector {
	return &AuditGapDetector{outboxRepo: outboxRepo, metrics: m}
}

// CheckGaps queries the audit outbox for dead-letter and stale-pending entries,
// updates metrics, and returns a structured report. Safe to call concurrently.
func (d *AuditGapDetector) CheckGaps(ctx context.Context) (*AuditGapReport, error) {
	tenantID, _ := shared.GetTenantID(ctx)

	dead, err := d.outboxRepo.CountDead(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	stale, err := d.outboxRepo.CountStalePending(ctx, OutboxStaleDuration)
	if err != nil {
		return nil, err
	}

	d.metrics.ObserveHistogram("finance_audit_outbox_dead_count", float64(dead), metrics.Fields{})
	d.metrics.ObserveHistogram("finance_audit_outbox_stale_pending_count", float64(stale), metrics.Fields{})

	if dead > 0 {
		logger.ErrorContext(ctx, "finance audit gap: DEAD outbox entries — CRITICAL events may not have audit records", logger.Fields{
			"dead_count": dead,
			"tenant_id":  tenantID.String(),
		})
		d.metrics.IncrementCounter("finance_audit_gap_dead_total", metrics.Fields{})
	}

	if stale > 0 {
		logger.WarnContext(ctx, "finance audit gap: stale PENDING outbox entries — delivery worker may be lagging", logger.Fields{
			"stale_count": stale,
			"stale_after": OutboxStaleDuration.String(),
			"tenant_id":   tenantID.String(),
		})
		d.metrics.IncrementCounter("finance_audit_gap_stale_total", metrics.Fields{})
	}

	report := &AuditGapReport{
		DeadOutboxCount:  dead,
		StaleOutboxCount: stale,
		Healthy:          dead == 0 && stale == 0,
		CheckedAt:        time.Now(),
	}
	return report, nil
}
