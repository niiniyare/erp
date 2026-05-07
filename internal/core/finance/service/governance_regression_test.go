// Package service_test — governance regression detection tests.
//
// Each test verifies that a specific protection is independently active and
// cannot be silently removed without causing test failures. This is mutation
// testing without code mutation: every adversarial input that would succeed
// if the protection were absent is tested to confirm it is correctly rejected.
//
// If any test in this file passes when its protection is removed, the CI
// pipeline has a governance coverage gap.
package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	gomock "go.uber.org/mock/gomock"

	"awo.so/internal/core/finance/domain"
	"awo.so/internal/core/finance/service"
	"awo.so/internal/shared"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

// ============================================================================
// FIN-GOV-001: SOD always blocks — verified across 20 random user IDs
// ============================================================================
// Proves SOD check is not hardcoded to a specific user ID. Every random
// user ID must be blocked when they attempt to approve their own transaction.

func TestGovernanceRegression_SOD_AlwaysBlocks_RandomUsers(t *testing.T) {
	txnID := uuid.New()

	for i := 0; i < 20; i++ {
		userID := uuid.New() // different random user each iteration

		repo := &stubTxnRepo{
			fnGetByID: func(_ context.Context, _ uuid.UUID) (*domain.Transaction, error) {
				return &domain.Transaction{
					ID:               txnID,
					TenantID:         uuid.New(),
					CreatedBy:        userID, // creator == approver — SOD violation
					ApprovalRequired: true,
					ApprovalStatus:   domain.ApprovalStatusPending,
				}, nil
			},
		}

		svc := newSvc(repo, &stubAccountRepo{}, &stubEntryService{})
		ctx := tenantAndUserCtx(userID)

		_, err := svc.ApproveTransaction(ctx, txnID, "")
		if err == nil {
			t.Errorf("iteration %d (userID=%v): SOD not enforced — approver==creator bypassed", i, userID)
			continue
		}
		if !containsCode(err, "SOD_VIOLATION") {
			t.Errorf("iteration %d: expected SOD_VIOLATION, got: %v", i, err)
		}
	}
}

// ============================================================================
// FIN-GOV-002: Terminal state deletion always blocked for ALL guarded statuses
// ============================================================================

func TestGovernanceRegression_TerminalDeletion_AllStatuses_AlwaysBlocked(t *testing.T) {
	guardedStatuses := []domain.TransactionStatus{
		domain.TransactionStatusPosted,
		domain.TransactionStatusReversed,
	}

	for _, status := range guardedStatuses {
		status := status
		t.Run(string(status), func(t *testing.T) {
			txnID := uuid.New()
			tenantID := uuid.New()

			repo := &stubTxnRepo{
				fnGetByID: func(_ context.Context, _ uuid.UUID) (*domain.Transaction, error) {
					return &domain.Transaction{
						ID:                txnID,
						TenantID:          tenantID,
						TransactionStatus: status,
					}, nil
				},
			}

			svc := newSvc(repo, &stubAccountRepo{}, &stubEntryService{})
			ctx := shared.WithTenantID(context.Background(), tenantID)

			err := svc.DeleteTransaction(ctx, txnID)
			if err == nil {
				t.Errorf("status=%s: deletion should be blocked, got nil", status)
			}
			if !containsCode(err, "TRANSACTION_DELETE_NOT_ALLOWED") {
				t.Errorf("status=%s: expected TRANSACTION_DELETE_NOT_ALLOWED, got: %v", status, err)
			}
		})
	}
}

// ============================================================================
// FIN-GOV-003: Approval gate always blocks DRAFT with ApprovalRequired=true
// ============================================================================

func TestGovernanceRegression_ApprovalGate_DraftBlocked(t *testing.T) {
	txnID := uuid.New()
	tenantID := uuid.New()

	repo := &stubTxnRepo{
		fnGetByID: func(_ context.Context, _ uuid.UUID) (*domain.Transaction, error) {
			return &domain.Transaction{
				ID:                txnID,
				TenantID:          tenantID,
				TransactionStatus: domain.TransactionStatusDraft,
				ApprovalRequired:  true,
				ApprovalStatus:    domain.ApprovalStatusPending,
			}, nil
		},
	}

	svc := newSvc(repo, &stubAccountRepo{}, &stubEntryService{})
	ctx := shared.WithTenantID(context.Background(), tenantID)

	_, err := svc.PostTransaction(ctx, txnID, nil)
	if err == nil {
		t.Fatal("approval gate not enforced — DRAFT+ApprovalRequired posted without approval")
	}
	if !containsCode(err, "APPROVAL_REQUIRED") {
		t.Errorf("expected APPROVAL_REQUIRED, got: %v", err)
	}
}

// ============================================================================
// FIN-GOV-004: Balance check always detects adversarial unbalanced inputs
// ============================================================================

func TestGovernanceRegression_BalanceCheck_AdversarialInputs_AlwaysDetected(t *testing.T) {
	type tc struct {
		name   string
		debit  decimal.Decimal
		credit decimal.Decimal
	}
	cases := []tc{
		{"off by 10", decimal.NewFromFloat(100), decimal.NewFromFloat(90)},
		{"debit only", decimal.NewFromFloat(500), decimal.NewFromFloat(0.01)}, // 0.01 credit to avoid no-entries path
		{"credit only", decimal.NewFromFloat(0.01), decimal.NewFromFloat(300)},
		{"large imbalance", decimal.NewFromFloat(1), decimal.NewFromFloat(1_000_000)},
		{"penny off", decimal.NewFromFloat(100.00), decimal.NewFromFloat(99.99)},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			entries := []domain.TransactionEntry{
				{DebitAmount: c.debit, CreditAmount: decimal.Zero},
				{DebitAmount: decimal.Zero, CreditAmount: c.credit},
			}
			txn := &domain.Transaction{
				ID:                uuid.New(),
				TenantID:          uuid.New(),
				TransactionStatus: domain.TransactionStatusPosted,
			}
			pageCount := 0
			repo := &stubTxnRepo{
				fnList: func(_ context.Context, _ *domain.TransactionFilter) ([]*domain.Transaction, error) {
					pageCount++
					if pageCount == 1 {
						return []*domain.Transaction{txn}, nil
					}
					return nil, nil
				},
				fnGetEntriesByTxn: func(_ context.Context, _ uuid.UUID) ([]domain.TransactionEntry, error) {
					return entries, nil
				},
			}
			intSvc := service.NewIntegrityService(repo, nil, metrics.NewNoOpMetricsProvider())
			report, err := intSvc.ScanPostedTransactions(context.Background(), nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			found := false
			for _, v := range report.Violations {
				if v.Kind == "UNBALANCED_TRANSACTION" {
					found = true
				}
			}
			if !found {
				t.Errorf("UNBALANCED_TRANSACTION not detected for %s (debit=%s credit=%s)",
					c.name, c.debit, c.credit)
			}
		})
	}
}

// ============================================================================
// FIN-GOV-005: Hash tamper always detected regardless of which field is tampered
// ============================================================================

func TestGovernanceRegression_HashTamper_AllFields_Detected(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tenantID := uuid.New()
	ctx := shared.WithTenantID(context.Background(), tenantID)

	baseEntry := buildEntry(1, "", tenantID)

	cases := []struct {
		name   string
		mutate func(*service.AuditChainEntry)
	}{
		{
			"payload tampered",
			func(e *service.AuditChainEntry) { e.Payload = []byte(`{"tampered":true}`) },
		},
		{
			"event type tampered",
			func(e *service.AuditChainEntry) { e.EventType = "INJECTED_EVENT" },
		},
		{
			"chain hash zeroed",
			func(e *service.AuditChainEntry) {
				e.ChainHash = "0000000000000000000000000000000000000000000000000000000000000000"
			},
		},
		{
			"created_at shifted",
			func(e *service.AuditChainEntry) { e.CreatedAt = e.CreatedAt.Add(time.Second) },
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			entry := *baseEntry
			c.mutate(&entry)

			mockRepo := service.NewMockAuditChainRepository(ctrl)
			mockRepo.EXPECT().GetChainStats(gomock.Any(), tenantID).Return(int64(1), int64(1), nil)
			mockRepo.EXPECT().ListChainRange(gomock.Any(), tenantID, int64(1), int64(1)).
				Return([]*service.AuditChainEntry{&entry}, nil)

			verifier := service.NewAuditChainVerifier(mockRepo, metrics.NewNoOpMetricsProvider())
			report, err := verifier.VerifyChain(ctx, 1, 1, 100)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if report.Healthy {
				t.Errorf("%s: tamper not detected — report is healthy", c.name)
			}
			found := false
			for _, v := range report.Violations {
				if v.Kind == "HASH_MISMATCH" {
					found = true
				}
			}
			if !found {
				t.Errorf("%s: HASH_MISMATCH violation not reported", c.name)
			}
		})
	}
}

// ============================================================================
// FIN-GOV-006: Velocity limit holds for multiple distinct limit values
// ============================================================================

func TestGovernanceRegression_VelocityLimit_MultipleValues(t *testing.T) {
	for _, limit := range []int{1, 3, 7, 15} {
		limit := limit
		t.Run(fmt.Sprintf("limit=%d", limit), func(t *testing.T) {
			policy := service.SafetyPolicy{
				MaxTransactionAmount:       decimal.NewFromFloat(10_000_000),
				MaxReversalsPerHour:        limit,
				MaxApprovalVelocityPerHour: 100,
				MaxPostingsPerHour:         1000,
			}
			enforcer := service.NewSafetyEnforcer(policy, metrics.NewNoOpMetricsProvider())
			userID := uuid.New()
			ctx := shared.WithTenantID(context.Background(), uuid.New())

			for i := 0; i < limit; i++ {
				if err := enforcer.CheckReversalVelocity(ctx, userID); err != nil {
					t.Fatalf("limit=%d call %d: unexpected block: %v", limit, i+1, err)
				}
			}
			if err := enforcer.CheckReversalVelocity(ctx, userID); err == nil {
				t.Errorf("limit=%d: call %d not blocked — limit not enforced", limit, limit+1)
			}
		})
	}
}

// ============================================================================
// FIN-GOV-007: Nil safety enforcer in service never panics
// ============================================================================

func TestGovernanceRegression_NilSafetyEnforcer_NoPanic(t *testing.T) {
	txnID := uuid.New()
	tenantID := uuid.New()

	repo := &stubTxnRepo{
		fnGetByID: func(_ context.Context, _ uuid.UUID) (*domain.Transaction, error) {
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
		nil, // nil safetyEnforcer — must not panic
		nil,
	)

	ctx := shared.WithTenantID(context.Background(), tenantID)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("nil safetyEnforcer caused panic: %v", r)
		}
	}()
	_, _ = svc.PostTransaction(ctx, txnID, nil)
}
