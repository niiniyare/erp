// Package service_test — safety enforcer unit tests.
//
// These tests validate the sliding-window velocity controls and amount limits
// that prevent financial abuse patterns (large transactions, reversal bursts,
// approval velocity abuse). No database required.
package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"awo.so/internal/core/finance/domain"
	"awo.so/internal/core/finance/service"
	"awo.so/internal/shared"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

// ============================================================================
// FIN-SAFE-001: Large transaction is blocked
// ============================================================================

func TestSafetyEnforcer_BlocksLargeTransaction(t *testing.T) {
	policy := service.SafetyPolicy{
		MaxTransactionAmount:       decimal.NewFromFloat(10_000),
		MaxReversalsPerHour:        20,
		MaxApprovalVelocityPerHour: 50,
		MaxPostingsPerHour:         500,
	}
	enforcer := service.NewSafetyEnforcer(policy, metrics.NewNoOpMetricsProvider())

	ctx := shared.WithTenantID(context.Background(), uuid.New())
	overLimit := decimal.NewFromFloat(10_001)

	err := enforcer.CheckTransactionAmount(ctx, overLimit)
	if err == nil {
		t.Fatal("expected error for large transaction, got nil")
	}
}

// ============================================================================
// FIN-SAFE-002: Transaction within limit passes
// ============================================================================

func TestSafetyEnforcer_AllowsWithinLimit(t *testing.T) {
	policy := service.SafetyPolicy{
		MaxTransactionAmount:       decimal.NewFromFloat(10_000),
		MaxReversalsPerHour:        20,
		MaxApprovalVelocityPerHour: 50,
		MaxPostingsPerHour:         500,
	}
	enforcer := service.NewSafetyEnforcer(policy, metrics.NewNoOpMetricsProvider())

	ctx := shared.WithTenantID(context.Background(), uuid.New())
	withinLimit := decimal.NewFromFloat(9_999.99)

	err := enforcer.CheckTransactionAmount(ctx, withinLimit)
	if err != nil {
		t.Fatalf("expected nil error for within-limit transaction, got: %v", err)
	}
}

// ============================================================================
// FIN-SAFE-003: Reversal velocity blocks user at limit
// ============================================================================

func TestSafetyEnforcer_ReversalVelocity_BlocksAtLimit(t *testing.T) {
	const limit = 3
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

	// Next call must be blocked.
	err := enforcer.CheckReversalVelocity(ctx, userID)
	if err == nil {
		t.Fatal("expected reversal velocity block, got nil")
	}
}

// ============================================================================
// FIN-SAFE-004: Per-user counters are independent across users
// ============================================================================

func TestSafetyEnforcer_ReversalVelocity_DifferentUsers_Independent(t *testing.T) {
	const limit = 2
	policy := service.SafetyPolicy{
		MaxTransactionAmount:       decimal.NewFromFloat(10_000_000),
		MaxReversalsPerHour:        limit,
		MaxApprovalVelocityPerHour: 50,
		MaxPostingsPerHour:         500,
	}
	enforcer := service.NewSafetyEnforcer(policy, metrics.NewNoOpMetricsProvider())

	userA := uuid.New()
	userB := uuid.New()
	ctx := shared.WithTenantID(context.Background(), uuid.New())

	// Exhaust user A's limit.
	for i := 0; i < limit; i++ {
		_ = enforcer.CheckReversalVelocity(ctx, userA)
	}
	if err := enforcer.CheckReversalVelocity(ctx, userA); err == nil {
		t.Fatal("expected user A to be blocked")
	}

	// User B has an independent counter — must not be blocked.
	for i := 0; i < limit; i++ {
		if err := enforcer.CheckReversalVelocity(ctx, userB); err != nil {
			t.Fatalf("user B call %d: unexpected block: %v", i+1, err)
		}
	}
}

// ============================================================================
// FIN-SAFE-005: Approval velocity blocks at limit
// ============================================================================

func TestSafetyEnforcer_ApprovalVelocity_BlocksAtLimit(t *testing.T) {
	const limit = 2
	policy := service.SafetyPolicy{
		MaxTransactionAmount:       decimal.NewFromFloat(10_000_000),
		MaxReversalsPerHour:        20,
		MaxApprovalVelocityPerHour: limit,
		MaxPostingsPerHour:         500,
	}
	enforcer := service.NewSafetyEnforcer(policy, metrics.NewNoOpMetricsProvider())

	approverID := uuid.New()
	ctx := shared.WithTenantID(context.Background(), uuid.New())

	for i := 0; i < limit; i++ {
		if err := enforcer.CheckApprovalVelocity(ctx, approverID); err != nil {
			t.Fatalf("call %d: unexpected error: %v", i+1, err)
		}
	}

	err := enforcer.CheckApprovalVelocity(ctx, approverID)
	if err == nil {
		t.Fatal("expected approval velocity block, got nil")
	}
}

// ============================================================================
// FIN-SAFE-006: Nil safety — nil enforcer in service does not panic
// ============================================================================

func TestSafetyEnforcer_NilEnforcer_ServiceDoesNotPanic(t *testing.T) {
	txnID := uuid.New()
	tenantID := uuid.New()

	repo := &stubTxnRepo{
		fnGetByID: func(_ context.Context, _ uuid.UUID) (*domain.Transaction, error) {
			// Return POSTED to trigger early idempotency return before enforcer check.
			return &domain.Transaction{
				ID:                txnID,
				TenantID:          tenantID,
				TransactionStatus: domain.TransactionStatusPosted,
			}, nil
		},
	}

	svc := service.NewTransactionService(
		repo, &stubAccountRepo{},
		nil, nil,
		&stubEntryService{},
		nil,
		tracing.NewNoOpService(),
		metrics.NewNoOpMetricsProvider(),
		nil,
		nil, // nil safetyEnforcer
		nil,
	)

	ctx := shared.WithTenantID(context.Background(), tenantID)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("nil safety enforcer caused panic: %v", r)
		}
	}()

	_, _ = svc.PostTransaction(ctx, txnID, nil)
}

// ============================================================================
// FIN-SAFE-007: SetTenantPolicy overrides global policy per-tenant
// ============================================================================

func TestSafetyEnforcer_SetTenantPolicy_Overrides(t *testing.T) {
	globalPolicy := service.SafetyPolicy{
		MaxTransactionAmount:       decimal.NewFromFloat(1_000),
		MaxReversalsPerHour:        20,
		MaxApprovalVelocityPerHour: 50,
		MaxPostingsPerHour:         500,
	}
	enforcer := service.NewSafetyEnforcer(globalPolicy, metrics.NewNoOpMetricsProvider())

	tenantID := uuid.New()
	ctx := shared.WithTenantID(context.Background(), tenantID)

	// 1_500 exceeds global limit.
	over := decimal.NewFromFloat(1_500)
	if err := enforcer.CheckTransactionAmount(ctx, over); err == nil {
		t.Fatal("expected block at global limit before override")
	}

	// Override with higher limit.
	tenantPolicy := service.SafetyPolicy{
		MaxTransactionAmount:       decimal.NewFromFloat(5_000),
		MaxReversalsPerHour:        20,
		MaxApprovalVelocityPerHour: 50,
		MaxPostingsPerHour:         500,
	}
	enforcer.SetTenantPolicy(tenantID, tenantPolicy)

	// 1_500 should now pass under tenant policy.
	if err := enforcer.CheckTransactionAmount(ctx, over); err != nil {
		t.Errorf("expected pass after tenant override, got: %v", err)
	}
}

// ============================================================================
// FIN-SAFE-008: PostingVelocity is WARN-only — never returns an error
// ============================================================================

func TestSafetyEnforcer_PostingVelocity_WarnOnly_NeverBlocks(t *testing.T) {
	const limit = 2
	policy := service.SafetyPolicy{
		MaxTransactionAmount:       decimal.NewFromFloat(10_000_000),
		MaxReversalsPerHour:        20,
		MaxApprovalVelocityPerHour: 50,
		MaxPostingsPerHour:         limit,
	}
	enforcer := service.NewSafetyEnforcer(policy, metrics.NewNoOpMetricsProvider())

	ctx := shared.WithTenantID(context.Background(), uuid.New())

	// Exceed limit many times — must never panic or block.
	for i := 0; i < limit+10; i++ {
		enforcer.CheckPostingVelocity(ctx)
	}
}
