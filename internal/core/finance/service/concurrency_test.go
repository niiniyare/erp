// Package service_test — adversarial concurrency simulation.
//
// Verifies that safety controls, integrity scans, and idempotency guards
// hold under concurrent load. All tests are race-detector compatible
// (run with: go test -race ./internal/core/finance/service/...).
package service_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"awo.so/internal/core/finance/domain"
	"awo.so/internal/core/finance/service"
	"awo.so/internal/shared"
	"awo.so/internal/shared/metrics"
)

// ============================================================================
// FIN-CONC-001: Reversal velocity limit never exceeded under 100 concurrent calls
// ============================================================================
// Proves the velocity window's internal mutex is correct — the exact limit
// is never exceeded even under goroutine contention.

func TestConcurrent_ReversalVelocity_LimitNeverExceeded(t *testing.T) {
	const limit = 10
	const goroutines = 100

	policy := service.SafetyPolicy{
		MaxTransactionAmount:       decimal.NewFromFloat(10_000_000),
		MaxReversalsPerHour:        limit,
		MaxApprovalVelocityPerHour: 50,
		MaxPostingsPerHour:         500,
	}
	enforcer := service.NewSafetyEnforcer(policy, metrics.NewNoOpMetricsProvider())

	userID := uuid.New()
	ctx := shared.WithTenantID(context.Background(), uuid.New())

	var (
		allowed int32
		wg      sync.WaitGroup
	)
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := enforcer.CheckReversalVelocity(ctx, userID); err == nil {
				atomic.AddInt32(&allowed, 1)
			}
		}()
	}
	wg.Wait()

	got := int(atomic.LoadInt32(&allowed))
	if got > limit {
		t.Errorf("reversal velocity limit %d exceeded under concurrency: %d allowed", limit, got)
	}
	if got == 0 {
		t.Errorf("expected at least 1 request allowed (limit=%d), got 0 — enforcer is broken", limit)
	}
}

// ============================================================================
// FIN-CONC-002: Approval velocity limit never exceeded under 100 concurrent calls
// ============================================================================

func TestConcurrent_ApprovalVelocity_LimitNeverExceeded(t *testing.T) {
	const limit = 5
	const goroutines = 80

	policy := service.SafetyPolicy{
		MaxTransactionAmount:       decimal.NewFromFloat(10_000_000),
		MaxReversalsPerHour:        100,
		MaxApprovalVelocityPerHour: limit,
		MaxPostingsPerHour:         500,
	}
	enforcer := service.NewSafetyEnforcer(policy, metrics.NewNoOpMetricsProvider())

	approverID := uuid.New()
	ctx := shared.WithTenantID(context.Background(), uuid.New())

	var (
		allowed int32
		wg      sync.WaitGroup
	)
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := enforcer.CheckApprovalVelocity(ctx, approverID); err == nil {
				atomic.AddInt32(&allowed, 1)
			}
		}()
	}
	wg.Wait()

	got := int(atomic.LoadInt32(&allowed))
	if got > limit {
		t.Errorf("approval velocity limit %d exceeded under concurrency: %d allowed", limit, got)
	}
}

// ============================================================================
// FIN-CONC-003: Concurrent PostTransaction replays on POSTED txn — zero mutations
// ============================================================================
// Proves idempotency guard is race-safe: 50 goroutines all see POSTED status
// and return early, calling repo.Post exactly 0 times.

func TestConcurrent_PostTransaction_Replays_ZeroMutations(t *testing.T) {
	txnID := uuid.New()
	tenantID := uuid.New()
	posted := &domain.Transaction{
		ID:                txnID,
		TenantID:          tenantID,
		TransactionStatus: domain.TransactionStatusPosted,
	}

	var mutationCount int32
	repo := &stubTxnRepo{
		fnGetByID: func(_ context.Context, _ uuid.UUID) (*domain.Transaction, error) {
			return posted, nil
		},
		fnPost: func(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ time.Time) error {
			atomic.AddInt32(&mutationCount, 1)
			return nil
		},
	}

	svc := newSvc(repo, &stubAccountRepo{}, &stubEntryService{})
	ctx := shared.WithTenantID(context.Background(), tenantID)

	const replays = 50
	errs := make([]error, replays)
	var wg sync.WaitGroup
	for i := 0; i < replays; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, errs[idx] = svc.PostTransaction(ctx, txnID, nil)
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("replay %d: unexpected error: %v", i, err)
		}
	}
	if got := int(atomic.LoadInt32(&mutationCount)); got != 0 {
		t.Errorf("expected 0 DB mutations on replay storm, got %d — idempotency broken", got)
	}
}

// ============================================================================
// FIN-CONC-004: Concurrent integrity scans produce no panics and no data races
// ============================================================================
// The IntegrityService creates per-call reports — shared state risk is in the
// repo stub and the enforcer. Run with -race to verify.

func TestConcurrent_IntegrityScan_NoPanicNoDataRace(t *testing.T) {
	txn := &domain.Transaction{
		ID:                uuid.New(),
		TenantID:          uuid.New(),
		TransactionStatus: domain.TransactionStatusPosted,
	}
	entries := []domain.TransactionEntry{
		{DebitAmount: decimal.NewFromFloat(100), CreditAmount: decimal.Zero},
		{DebitAmount: decimal.Zero, CreditAmount: decimal.NewFromFloat(100)},
	}

	// Read-only stubs — safe for concurrent access.
	repo := &stubTxnRepo{
		fnList: func(_ context.Context, f *domain.TransactionFilter) ([]*domain.Transaction, error) {
			// Return one page then stop. Offset-based pagination: second page returns empty.
			if f != nil && f.Offset != nil && *f.Offset > 0 {
				return nil, nil
			}
			return []*domain.Transaction{txn}, nil
		},
		fnGetEntriesByTxn: func(_ context.Context, _ uuid.UUID) ([]domain.TransactionEntry, error) {
			return entries, nil
		},
	}

	intSvc := service.NewIntegrityService(repo, nil, metrics.NewNoOpMetricsProvider())

	const goroutines = 20
	var wg sync.WaitGroup
	panics := make([]interface{}, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					panics[idx] = r
				}
			}()
			_, _ = intSvc.ScanPostedTransactions(context.Background(), nil)
		}(i)
	}
	wg.Wait()

	for i, p := range panics {
		if p != nil {
			t.Errorf("goroutine %d panicked: %v", i, p)
		}
	}
}

// ============================================================================
// FIN-CONC-005: Concurrent nil AuditChainVerifier calls never panic
// ============================================================================

func TestConcurrent_NilAuditChainVerifier_NoPanic(t *testing.T) {
	var verifier *service.AuditChainVerifier

	tenantID := uuid.New()
	ctx := shared.WithTenantID(context.Background(), tenantID)

	const goroutines = 30
	var wg sync.WaitGroup
	panics := make([]interface{}, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					panics[idx] = r
				}
			}()
			_, _ = verifier.VerifyChain(ctx, 1, 100, 50)
		}(i)
	}
	wg.Wait()

	for i, p := range panics {
		if p != nil {
			t.Errorf("goroutine %d: nil verifier panicked: %v", i, p)
		}
	}
}
