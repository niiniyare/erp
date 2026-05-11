// Package service_test — replay storm and idempotency survival tests.
//
// Simulates Temporal worker restarts and replay storms: the same operation
// is executed many times concurrently or sequentially to verify that
// idempotency guards prevent duplicate financial effects.
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
	"awo.so/internal/shared/tracing"
)

// ============================================================================
// FIN-REPLAY-001: PostTransaction — 50 concurrent replays on POSTED txn
// ============================================================================
// Simulates a Temporal replay storm after worker restart.
// The already-POSTED transaction must be returned identically each time with
// zero calls to repo.Post (idempotent guard fires on every replay).

func TestReplayStorm_PostTransaction_50Replays_ZeroMutations(t *testing.T) {
	txnID := uuid.New()
	tenantID := uuid.New()
	postedTxn := &domain.Transaction{
		ID:                txnID,
		TenantID:          tenantID,
		TransactionStatus: domain.TransactionStatusPosted,
		TransactionNumber: "TXN-REPLAY-001",
	}

	var postCalls int32
	repo := &stubTxnRepo{
		fnGetByID: func(_ context.Context, _ uuid.UUID) (*domain.Transaction, error) {
			return postedTxn, nil
		},
		fnPost: func(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ time.Time) error {
			atomic.AddInt32(&postCalls, 1)
			return nil
		},
	}

	svc := newSvc(repo, &stubAccountRepo{}, &stubEntryService{})
	ctx := shared.WithTenantID(context.Background(), tenantID)

	const replays = 50
	results := make([]*domain.Transaction, replays)
	errors := make([]error, replays)
	var wg sync.WaitGroup

	for i := 0; i < replays; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx], errors[idx] = svc.PostTransaction(ctx, txnID, nil)
		}(i)
	}
	wg.Wait()

	for i := 0; i < replays; i++ {
		if errors[i] != nil {
			t.Errorf("replay %d: expected nil error, got: %v", i, errors[i])
		}
		if results[i] == nil || results[i].TransactionStatus != domain.TransactionStatusPosted {
			t.Errorf("replay %d: expected POSTED transaction", i)
		}
	}
	if got := int(atomic.LoadInt32(&postCalls)); got != 0 {
		t.Errorf("replay storm: expected 0 Post mutations, got %d — duplicate posting occurred", got)
	}
}

// ============================================================================
// FIN-REPLAY-002: ApproveTransaction — 50 concurrent replays on APPROVED txn
// ============================================================================

func TestReplayStorm_ApproveTransaction_50Replays_ZeroMutations(t *testing.T) {
	txnID := uuid.New()
	approverID := uuid.New()
	approvedTxn := &domain.Transaction{
		ID:             txnID,
		TenantID:       uuid.New(),
		ApprovalStatus: domain.ApprovalStatusApproved,
	}

	var approveCalls int32
	repo := &stubTxnRepo{
		fnGetByID: func(_ context.Context, _ uuid.UUID) (*domain.Transaction, error) {
			return approvedTxn, nil
		},
		fnApprove: func(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ time.Time, _ *string) error {
			atomic.AddInt32(&approveCalls, 1)
			return nil
		},
	}

	svc := newSvc(repo, &stubAccountRepo{}, &stubEntryService{})
	ctx := tenantAndUserCtx(approverID)

	const replays = 50
	errors := make([]error, replays)
	var wg sync.WaitGroup

	for i := 0; i < replays; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, errors[idx] = svc.ApproveTransaction(ctx, txnID, "idempotent replay")
		}(i)
	}
	wg.Wait()

	for i, err := range errors {
		if err != nil {
			t.Errorf("replay %d: unexpected error: %v", i, err)
		}
	}
	if got := int(atomic.LoadInt32(&approveCalls)); got != 0 {
		t.Errorf("replay storm: expected 0 Approve mutations on already-APPROVED, got %d", got)
	}
}

// ============================================================================
// FIN-REPLAY-003: SOD violation never bypassed under concurrent replay attempts
// ============================================================================
// Simulates adversary repeatedly retrying approval with creator == approver.
// Every attempt must be rejected regardless of concurrency.

func TestReplayStorm_SODViolation_NeverBypassed(t *testing.T) {
	txnID := uuid.New()
	userID := uuid.New() // same user as creator

	repo := &stubTxnRepo{
		fnGetByID: func(_ context.Context, _ uuid.UUID) (*domain.Transaction, error) {
			return &domain.Transaction{
				ID:               txnID,
				TenantID:         uuid.New(),
				CreatedBy:        userID, // creator == approver
				ApprovalRequired: true,
				ApprovalStatus:   domain.ApprovalStatusPending,
			}, nil
		},
	}

	svc := newSvc(repo, &stubAccountRepo{}, &stubEntryService{})
	ctx := tenantAndUserCtx(userID) // same user

	const attempts = 30
	bypasses := make([]bool, attempts)
	var wg sync.WaitGroup

	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, err := svc.ApproveTransaction(ctx, txnID, "bypass attempt")
			bypasses[idx] = (err == nil) // true means bypass succeeded (BAD)
		}(i)
	}
	wg.Wait()

	for i, bypassed := range bypasses {
		if bypassed {
			t.Errorf("attempt %d: SOD bypass succeeded — protection is broken", i)
		}
	}
}

// ============================================================================
// FIN-REPLAY-004: SafetyEnforcer — velocity not bypassed under rapid sequential replay
// ============================================================================
// Simulates a retry loop rapidly exhausting and probing the velocity window.

func TestReplayStorm_SafetyEnforcer_VelocityWindowHolds(t *testing.T) {
	const limit = 5
	policy := service.SafetyPolicy{
		MaxTransactionAmount:       decimal.NewFromFloat(10_000_000),
		MaxReversalsPerHour:        limit,
		MaxApprovalVelocityPerHour: 50,
		MaxPostingsPerHour:         500,
	}
	enforcer := service.NewSafetyEnforcer(policy, metrics.NewNoOpMetricsProvider())

	userID := uuid.New()
	ctx := shared.WithTenantID(context.Background(), uuid.New())

	// Exhaust the limit.
	for i := 0; i < limit; i++ {
		if err := enforcer.CheckReversalVelocity(ctx, userID); err != nil {
			t.Fatalf("call %d: unexpected block before limit: %v", i+1, err)
		}
	}

	// Rapid replay: every subsequent call must be blocked.
	const extraAttempts = 50
	for i := 0; i < extraAttempts; i++ {
		if err := enforcer.CheckReversalVelocity(ctx, userID); err == nil {
			t.Errorf("replay attempt %d: expected block after limit=%d, got nil — limit bypassed", i, limit)
		}
	}
}

// ============================================================================
// FIN-REPLAY-005: Nil AuditChainWriter survives concurrent replay calls
// ============================================================================
// Simulates worker restart replaying outbox delivery events when writer is nil.

func TestReplayStorm_NilAuditChainWriter_NoPanic(t *testing.T) {
	var writer *service.AuditChainWriter

	tenantID := uuid.New()
	ctx := shared.WithTenantID(context.Background(), tenantID)

	const goroutines = 40
	panics := make([]any, goroutines)
	var wg sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					panics[idx] = r
				}
			}()
			writer.AppendDelivered(ctx, uuid.New(), "TRANSACTION_POSTED", []byte(`{}`), time.Now())
		}(i)
	}
	wg.Wait()

	for i, p := range panics {
		if p != nil {
			t.Errorf("goroutine %d: nil AuditChainWriter panicked: %v", i, p)
		}
	}
}

// ============================================================================
// FIN-REPLAY-006: Terminal state deletion blocked under sequential retry loop
// ============================================================================
// Simulates an operator retrying a delete on a POSTED transaction many times.

func TestReplayStorm_TerminalStateDeletion_AlwaysBlocked(t *testing.T) {
	txnID := uuid.New()
	tenantID := uuid.New()

	var deleteCalls int32
	repo := &stubTxnRepo{
		fnGetByID: func(_ context.Context, _ uuid.UUID) (*domain.Transaction, error) {
			return &domain.Transaction{
				ID:                txnID,
				TenantID:          tenantID,
				TransactionStatus: domain.TransactionStatusPosted,
			}, nil
		},
		fnDelete: func(_ context.Context, _ uuid.UUID) error {
			atomic.AddInt32(&deleteCalls, 1)
			return nil
		},
	}

	svc := service.NewTransactionService(
		repo, &stubAccountRepo{},
		nil, nil,
		&stubEntryService{},
		nil,
		tracing.NewNoOpService(),
		metrics.NewNoOpMetricsProvider(),
		nil, nil, nil,
	)
	ctx := shared.WithTenantID(context.Background(), tenantID)

	const attempts = 20
	for i := 0; i < attempts; i++ {
		err := svc.DeleteTransaction(ctx, txnID)
		if err == nil {
			t.Errorf("attempt %d: expected deletion blocked on POSTED, got nil", i)
		}
	}
	if got := int(atomic.LoadInt32(&deleteCalls)); got != 0 {
		t.Errorf("repo.Delete called %d times despite protection — guard bypassed", got)
	}
}
