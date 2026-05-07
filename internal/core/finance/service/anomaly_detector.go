package service

// anomaly_detector.go — Lightweight heuristic anomaly detection for the finance module.
//
// AnomalyDetector observes finance mutation events and emits structured metrics
// when suspicious patterns are detected. Detection is intentionally non-blocking:
// SafetyEnforcer handles hard limits; this layer handles statistical anomalies.
//
// All methods are nil-safe.

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"awo.so/internal/shared"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
)

// AnomalyDetector observes finance mutations and detects suspicious behavioral patterns.
// It is metrics-driven: each observation increments a counter and the alerting layer
// (Prometheus, Grafana, PagerDuty) decides when to page.
type AnomalyDetector struct {
	// largeTxnThreshold triggers an anomaly when a single posting exceeds this amount.
	largeTxnThreshold decimal.Decimal
	metrics           metrics.MetricsProvider
}

// NewAnomalyDetector creates a detector with production defaults.
func NewAnomalyDetector(m metrics.MetricsProvider) *AnomalyDetector {
	return &AnomalyDetector{
		// 1 M — suspicious but not necessarily blocked; SafetyEnforcer blocks at 10 M.
		largeTxnThreshold: decimal.NewFromInt(1_000_000),
		metrics:           m,
	}
}

// ObservePosting records a posting event and fires anomalies for large amounts.
func (d *AnomalyDetector) ObservePosting(ctx context.Context, totalDebit decimal.Decimal) {
	if d == nil {
		return
	}
	tenantID, _ := shared.GetTenantID(ctx)
	d.metrics.IncrementCounter("finance_anomaly_postings_total", metrics.Fields{
		"tenant_id": tenantID.String(),
	})
	if totalDebit.GreaterThan(d.largeTxnThreshold) {
		d.metrics.IncrementCounter("finance_anomaly_events_total", metrics.Fields{
			"kind":      "LARGE_TRANSACTION",
			"tenant_id": tenantID.String(),
		})
		logger.WarnContext(ctx, "finance anomaly: large transaction posted — review recommended", logger.Fields{
			"amount":    totalDebit.StringFixed(2),
			"threshold": d.largeTxnThreshold.StringFixed(2),
		})
	}
}

// ObserveReversal records a reversal event for churn detection.
func (d *AnomalyDetector) ObserveReversal(ctx context.Context, byUserID uuid.UUID) {
	if d == nil {
		return
	}
	d.metrics.IncrementCounter("finance_anomaly_reversals_total", metrics.Fields{
		"user_id": byUserID.String(),
	})
}

// ObserveApproval records a successful approval event.
func (d *AnomalyDetector) ObserveApproval(ctx context.Context, byUserID uuid.UUID) {
	if d == nil {
		return
	}
	d.metrics.IncrementCounter("finance_anomaly_approvals_total", metrics.Fields{
		"user_id": byUserID.String(),
	})
}

// ObserveApprovalFailure records a failed or rejected approval attempt.
// Repeated failures from the same user are a signal of abuse or misconfiguration.
func (d *AnomalyDetector) ObserveApprovalFailure(ctx context.Context, byUserID uuid.UUID) {
	if d == nil {
		return
	}
	tenantID, _ := shared.GetTenantID(ctx)
	d.metrics.IncrementCounter("finance_anomaly_approval_failures_total", metrics.Fields{
		"user_id":   byUserID.String(),
		"tenant_id": tenantID.String(),
	})
	d.metrics.IncrementCounter("finance_anomaly_events_total", metrics.Fields{
		"kind":      "APPROVAL_FAILURE",
		"tenant_id": tenantID.String(),
	})
}

// ObserveReconciliationUnmatch records a line-unmatch event (reconciliation churn).
func (d *AnomalyDetector) ObserveReconciliationUnmatch(ctx context.Context) {
	if d == nil {
		return
	}
	tenantID, _ := shared.GetTenantID(ctx)
	d.metrics.IncrementCounter("finance_anomaly_reconciliation_unmatches_total", metrics.Fields{
		"tenant_id": tenantID.String(),
	})
}

// ObserveRejection records a transaction rejection for approval-chain monitoring.
func (d *AnomalyDetector) ObserveRejection(ctx context.Context, byUserID uuid.UUID) {
	if d == nil {
		return
	}
	d.metrics.IncrementCounter("finance_anomaly_rejections_total", metrics.Fields{
		"user_id": byUserID.String(),
	})
}
