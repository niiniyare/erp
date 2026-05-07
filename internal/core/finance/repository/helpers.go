package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	db "awo.so/db/sqlc"
	"awo.so/internal/shared"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
)

// txFrom extracts the underlying pgx.Tx from a db.Store that has been wrapped
// inside a WithTenant / WithTenantFromCtx callback. The store passed to the
// callback is always a db.TxStore; if it is not, the call site has a bug.
func txFrom(s db.Store) (pgx.Tx, error) {
	ts, ok := s.(db.TxStore)
	if !ok {
		return nil, fmt.Errorf("finance repository: store is not a TxStore — must be called inside a WithTenant callback")
	}
	return ts.GetTx(), nil
}

// nullUUID converts a non-pointer UUID to *uuid.UUID for use as a nullable
// SQL parameter. uuid.Nil is returned as nil (SQL NULL).
func nullUUID(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	return &id
}

// nullUUID2 is an identity pass-through for *uuid.UUID: a nil pointer becomes
// SQL NULL and a non-nil pointer passes the UUID value through unchanged.
func nullUUID2(id *uuid.UUID) *uuid.UUID {
	return id
}

// ── Observability helpers ─────────────────────────────────────────────────────

// observeOp records operation duration (histogram) and error count (counter)
// for a finance repository method. It also logs errors with structured context.
//
// Usage pattern (named return + defer):
//
//	func (r *fooRepository) Op(ctx context.Context) (_ T, err error) {
//	    start := time.Now()
//	    defer func() { observeOp(ctx, "Op", "foo", start, err, r.logger, r.metrics) }()
//	    ...
//	}
func observeOp(ctx context.Context, op, repo string, start time.Time, err error, log logger.Logger, met metrics.MetricsProvider) {
	if err != nil && log != nil {
		tenantID, _ := shared.GetTenantID(ctx)
		log.ErrorContext(ctx, "repository operation failed", logger.Fields{
			"repository": repo,
			"operation":  op,
			"tenant_id":  tenantID.String(),
			"error":      err.Error(),
		})
	}
	if met == nil {
		return
	}
	status := "success"
	if err != nil {
		status = "error"
	}
	met.ObserveHistogram("finance_repo_op_duration_seconds",
		time.Since(start).Seconds(),
		metrics.Fields{
			"operation":  op,
			"repository": repo,
			"status":     status,
		})
	if err != nil {
		met.IncrementCounter("finance_repo_errors_total",
			metrics.Fields{
				"operation":  op,
				"repository": repo,
			})
	}
}

// initFinanceRepoMetrics registers the shared metric descriptors for finance
// repositories. Safe to call multiple times — the MetricsProvider deduplicates.
func initFinanceRepoMetrics(met metrics.MetricsProvider) {
	if met == nil {
		return
	}
	met.Histogram(
		"finance_repo_op_duration_seconds",
		"Finance repository operation duration in seconds",
		[]float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
		"operation", "repository", "status",
	)
	met.Counter(
		"finance_repo_errors_total",
		"Finance repository error total",
		"operation", "repository",
	)
}

// zeroTimeToNil converts a time.Time value to a *time.Time pointer.
// Zero-valued times (representing SQL NULL scanned by pgx) are returned as nil
// so that callers receive a proper nil pointer rather than a pointer to the
// zero instant (0001-01-01 00:00:00 UTC).
func zeroTimeToNil(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}
