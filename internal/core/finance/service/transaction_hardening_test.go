package service_test

// Hardening tests — invariant, regression, and concurrency guarantees.
//
// All mock types reuse those defined in other test files (same package):
//   mockTransactionRepo, mockReversalHistoryRepo, mockEntryService — transaction_test.go
//   mockAccountsRepo — account_service_test.go
//   mockPeriodRepo   — period_test.go

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"awo.so/internal/core/finance/domain"
	"awo.so/internal/core/finance/service"
	"awo.so/internal/shared"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func hardenCtx(t *testing.T) context.Context {
	t.Helper()
	ctx := shared.WithTenantID(context.Background(), uuid.New())
	return shared.WithUserID(ctx, uuid.New())
}

func hardenBalancedEntries(txnID, tenantID uuid.UUID, amount int64) []*domain.TransactionEntry {
	return []*domain.TransactionEntry{
		{ID: uuid.New(), TenantID: tenantID, TransactionID: txnID, EntryNumber: 1,
			AccountID: uuid.New(), DebitAmount: decimal.NewFromInt(amount),
			Description: "DR", ExchangeRate: decimal.NewFromInt(1)},
		{ID: uuid.New(), TenantID: tenantID, TransactionID: txnID, EntryNumber: 2,
			AccountID: uuid.New(), CreditAmount: decimal.NewFromInt(amount),
			Description: "CR", ExchangeRate: decimal.NewFromInt(1)},
	}
}

func hardenActiveAccount() *domain.Accounts {
	return &domain.Accounts{ID: uuid.New(), IsActive: true, AllowManualEntries: true}
}

func newHardenSvc(
	repo *mockTransactionRepo,
	acctRepo *mockAccountsRepo,
	periodRepo domain.PeriodRepository,
	histRepo *mockReversalHistoryRepo,
	entrySvc *mockEntryService,
) service.TransactionService {
	return service.NewTransactionService(
		repo, acctRepo, periodRepo, histRepo, entrySvc, nil,
		tracing.NewNoOpService(),
		metrics.NewNoOpMetricsProvider(),
	)
}

// ---------------------------------------------------------------------------
// INV-001: Unbalanced transaction must never post
// ---------------------------------------------------------------------------

func TestINV001_PostTransaction_UnbalancedEntries_MustFail(t *testing.T) {
	ctx := hardenCtx(t)
	tenantID, _ := shared.GetTenantID(ctx)

	txnID := uuid.New()
	acctID1, acctID2 := uuid.New(), uuid.New()
	txn := &domain.Transaction{
		ID: txnID, TenantID: tenantID,
		TransactionNumber: "TXN-INV-001",
		TransactionType:   domain.TransactionTypeManual,
		TransactionStatus: domain.TransactionStatusDraft,
		ApprovalRequired:  false, ApprovalStatus: domain.ApprovalStatusNotRequired,
		TransactionDate: time.Now(), Description: "Unbalanced",
		CurrencyCode: "USD", ExchangeRate: decimal.NewFromInt(1),
		TotalDebitAmount: decimal.NewFromInt(500), TotalCreditAmount: decimal.NewFromInt(400),
		CreatedBy: uuid.New(),
	}
	entries := []*domain.TransactionEntry{
		{ID: uuid.New(), TenantID: tenantID, TransactionID: txnID, EntryNumber: 1,
			AccountID: acctID1, DebitAmount: decimal.NewFromInt(500),
			Description: "DR", ExchangeRate: decimal.NewFromInt(1)},
		{ID: uuid.New(), TenantID: tenantID, TransactionID: txnID, EntryNumber: 2,
			AccountID: acctID2, CreditAmount: decimal.NewFromInt(400),
			Description: "CR", ExchangeRate: decimal.NewFromInt(1)},
	}

	repo := &mockTransactionRepo{}
	acctRepo := &mockAccountsRepo{}
	entrySvc := &mockEntryService{}

	repo.On("GetByID", mock.Anything, txnID).Return(txn, nil)
	// Service auto-approves DRAFT+no-approval-required before validating; stub it.
	repo.On("Approve", mock.Anything, txnID, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	entrySvc.On("GetEntriesByTransactionID", mock.Anything, txnID).Return(entries, nil)
	acctRepo.On("GetByID", mock.Anything, acctID1).Return(hardenActiveAccount(), nil)
	acctRepo.On("GetByID", mock.Anything, acctID2).Return(hardenActiveAccount(), nil)

	svc := newHardenSvc(repo, acctRepo, nil, nil, entrySvc)
	_, err := svc.PostTransaction(ctx, txnID, nil)

	require.Error(t, err, "unbalanced transaction must not post")
}

// ---------------------------------------------------------------------------
// INV-002: Cannot reverse a reversal
// ---------------------------------------------------------------------------

func TestINV002_ReverseTransaction_CannotReverseReversal(t *testing.T) {
	ctx := hardenCtx(t)
	tenantID, _ := shared.GetTenantID(ctx)

	txnID := uuid.New()
	txn := &domain.Transaction{
		ID: txnID, TenantID: tenantID,
		TransactionType:   domain.TransactionTypeReversal,
		TransactionStatus: domain.TransactionStatusPosted,
		IsReversed:        false,
		ApprovalRequired:  false, ApprovalStatus: domain.ApprovalStatusNotRequired,
		CurrencyCode: "USD", ExchangeRate: decimal.NewFromInt(1),
		CreatedBy: uuid.New(),
	}

	repo := &mockTransactionRepo{}
	histRepo := &mockReversalHistoryRepo{}
	entrySvc := &mockEntryService{}
	acctRepo := &mockAccountsRepo{}

	repo.On("GetByID", mock.Anything, txnID).Return(txn, nil)
	histRepo.On("IsReversal", mock.Anything, txnID).Return(true, nil)

	svc := newHardenSvc(repo, acctRepo, nil, histRepo, entrySvc)
	_, err := svc.ReverseTransaction(ctx, txnID, "test")

	require.Error(t, err, "reversing a reversal must fail")
}

// ---------------------------------------------------------------------------
// INV-003: Posting to closed period must fail
// ---------------------------------------------------------------------------

func TestINV003_PostTransaction_ClosedPeriod_MustFail(t *testing.T) {
	ctx := hardenCtx(t)
	tenantID, _ := shared.GetTenantID(ctx)

	txnID := uuid.New()
	txn := &domain.Transaction{
		ID: txnID, TenantID: tenantID,
		TransactionNumber: "TXN-PERIOD-001",
		TransactionType:   domain.TransactionTypeManual,
		TransactionStatus: domain.TransactionStatusDraft,
		ApprovalRequired:  false, ApprovalStatus: domain.ApprovalStatusNotRequired,
		TransactionDate: time.Now(), Description: "Period test",
		CurrencyCode: "USD", ExchangeRate: decimal.NewFromInt(1),
		TotalDebitAmount: decimal.NewFromInt(100), TotalCreditAmount: decimal.NewFromInt(100),
		CreatedBy: uuid.New(),
	}
	entries := hardenBalancedEntries(txnID, tenantID, 100)
	closedPeriod := &domain.AccountingPeriod{
		ID: uuid.New(), TenantID: tenantID, Name: "Jan 2025",
		Status: domain.PeriodStatusHardClosed,
	}

	repo := &mockTransactionRepo{}
	acctRepo := &mockAccountsRepo{}
	entrySvc := &mockEntryService{}
	periodRepo := &mockPeriodRepo{}

	repo.On("GetByID", mock.Anything, txnID).Return(txn, nil)
	repo.On("Approve", mock.Anything, txnID, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	entrySvc.On("GetEntriesByTransactionID", mock.Anything, txnID).Return(entries, nil)
	for _, e := range entries {
		acctRepo.On("GetByID", mock.Anything, e.AccountID).Return(hardenActiveAccount(), nil)
	}
	periodRepo.On("GetPeriodForDate", mock.Anything, tenantID, mock.Anything).Return(closedPeriod, nil)

	svc := newHardenSvc(repo, acctRepo, periodRepo, nil, entrySvc)
	_, err := svc.PostTransaction(ctx, txnID, nil)

	require.Error(t, err, "posting to closed period must fail")
	assert.Contains(t, err.Error(), "PERIOD_CLOSED")
}

// ---------------------------------------------------------------------------
// INV-004: Posting to inactive account must fail
// ---------------------------------------------------------------------------

func TestINV004_PostTransaction_InactiveAccount_MustFail(t *testing.T) {
	ctx := hardenCtx(t)
	tenantID, _ := shared.GetTenantID(ctx)

	txnID := uuid.New()
	inactiveAcctID, activeAcctID := uuid.New(), uuid.New()
	txn := &domain.Transaction{
		ID: txnID, TenantID: tenantID,
		TransactionNumber: "TXN-ACCT-001",
		TransactionType:   domain.TransactionTypeManual,
		TransactionStatus: domain.TransactionStatusDraft,
		ApprovalRequired:  false, ApprovalStatus: domain.ApprovalStatusNotRequired,
		TransactionDate: time.Now(), Description: "Inactive acct",
		CurrencyCode: "USD", ExchangeRate: decimal.NewFromInt(1),
		TotalDebitAmount: decimal.NewFromInt(100), TotalCreditAmount: decimal.NewFromInt(100),
		CreatedBy: uuid.New(),
	}
	entries := []*domain.TransactionEntry{
		{ID: uuid.New(), TenantID: tenantID, TransactionID: txnID, EntryNumber: 1,
			AccountID: inactiveAcctID, DebitAmount: decimal.NewFromInt(100),
			Description: "DR", ExchangeRate: decimal.NewFromInt(1)},
		{ID: uuid.New(), TenantID: tenantID, TransactionID: txnID, EntryNumber: 2,
			AccountID: activeAcctID, CreditAmount: decimal.NewFromInt(100),
			Description: "CR", ExchangeRate: decimal.NewFromInt(1)},
	}

	repo := &mockTransactionRepo{}
	acctRepo := &mockAccountsRepo{}
	entrySvc := &mockEntryService{}

	repo.On("GetByID", mock.Anything, txnID).Return(txn, nil)
	repo.On("Approve", mock.Anything, txnID, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	entrySvc.On("GetEntriesByTransactionID", mock.Anything, txnID).Return(entries, nil)
	acctRepo.On("GetByID", mock.Anything, inactiveAcctID).Return(
		&domain.Accounts{ID: inactiveAcctID, IsActive: false, AllowManualEntries: true}, nil)
	acctRepo.On("GetByID", mock.Anything, activeAcctID).Return(hardenActiveAccount(), nil)

	svc := newHardenSvc(repo, acctRepo, nil, nil, entrySvc)
	_, err := svc.PostTransaction(ctx, txnID, nil)

	require.Error(t, err, "posting to inactive account must fail")
}

// ---------------------------------------------------------------------------
// INV-005: domain.CreateReversalTransaction must produce TransactionTypeReversal
// Regression: was TransactionTypeAdjustment — double-reversal guard could be bypassed
// ---------------------------------------------------------------------------

func TestINV005_DomainCreateReversalTransaction_TypeMustBeReversal(t *testing.T) {
	original := &domain.Transaction{
		ID: uuid.New(), TenantID: uuid.New(),
		TransactionNumber: "TXN-001",
		TransactionType:   domain.TransactionTypeManual,
		TransactionStatus: domain.TransactionStatusPosted,
		CurrencyCode:      "USD", ExchangeRate: decimal.NewFromInt(1),
		TotalDebitAmount: decimal.NewFromInt(100), TotalCreditAmount: decimal.NewFromInt(100),
		Entries: []domain.TransactionEntry{
			{ID: uuid.New(), AccountID: uuid.New(), DebitAmount: decimal.NewFromInt(100),
				Description: "DR", ExchangeRate: decimal.NewFromInt(1)},
			{ID: uuid.New(), AccountID: uuid.New(), CreditAmount: decimal.NewFromInt(100),
				Description: "CR", ExchangeRate: decimal.NewFromInt(1)},
		},
	}

	reversal, err := original.CreateReversalTransaction("test", uuid.New())
	require.NoError(t, err)
	assert.Equal(t, domain.TransactionTypeReversal, reversal.TransactionType,
		"reversal type MUST be REVERSAL — double-reversal guard depends on this")
}

// ---------------------------------------------------------------------------
// INV-006: Submitter cannot reject their own transaction (SOD)
// ---------------------------------------------------------------------------

func TestINV006_RejectTransaction_SubmitterCannotReject_SODViolation(t *testing.T) {
	submitterID := uuid.New()
	ctx := shared.WithTenantID(context.Background(), uuid.New())
	ctx = shared.WithUserID(ctx, submitterID) // caller IS the submitter

	txnID := uuid.New()
	txn := &domain.Transaction{
		ID: txnID, TenantID: uuid.New(),
		ApprovalRequired: true, ApprovalStatus: domain.ApprovalStatusPending,
		CreatedBy: submitterID,
		CurrencyCode: "USD", ExchangeRate: decimal.NewFromInt(1),
	}

	repo := &mockTransactionRepo{}
	acctRepo := &mockAccountsRepo{}
	entrySvc := &mockEntryService{}

	repo.On("GetByID", mock.Anything, txnID).Return(txn, nil)

	svc := newHardenSvc(repo, acctRepo, nil, nil, entrySvc)
	_, err := svc.RejectTransaction(ctx, txnID, domain.RejectionReasonPolicyViolation, "self-reject")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "SOD_VIOLATION")
}

// ---------------------------------------------------------------------------
// INV-007: RejectTransaction records approver identity, not submitter
// Regression: was always using transaction.CreatedBy as rejectedBy
// ---------------------------------------------------------------------------

func TestINV007_RejectTransaction_UsesApproverNotSubmitter(t *testing.T) {
	submitterID := uuid.New()
	approverID := uuid.New()

	ctx := shared.WithTenantID(context.Background(), uuid.New())
	ctx = shared.WithUserID(ctx, approverID)

	txnID := uuid.New()
	txn := &domain.Transaction{
		ID: txnID, TenantID: uuid.New(),
		ApprovalRequired: true, ApprovalStatus: domain.ApprovalStatusPending,
		CreatedBy: submitterID,
		CurrencyCode: "USD", ExchangeRate: decimal.NewFromInt(1),
	}
	rejected := *txn
	rejected.ApprovalStatus = domain.ApprovalStatusRejected

	repo := &mockTransactionRepo{}
	acctRepo := &mockAccountsRepo{}
	entrySvc := &mockEntryService{}

	repo.On("GetByID", mock.Anything, txnID).Return(txn, nil).Once()
	repo.On("Reject", mock.Anything, txnID, approverID,
		mock.Anything, domain.RejectionReasonPolicyViolation, mock.Anything).Return(nil)
	repo.On("GetByID", mock.Anything, txnID).Return(&rejected, nil).Once()

	svc := newHardenSvc(repo, acctRepo, nil, nil, entrySvc)
	_, _ = svc.RejectTransaction(ctx, txnID, domain.RejectionReasonPolicyViolation, "budget")

	repo.AssertCalled(t, "Reject", mock.Anything, txnID, approverID,
		mock.Anything, domain.RejectionReasonPolicyViolation, mock.Anything)
}

// ---------------------------------------------------------------------------
// INV-008: ReconcileEntries must fail on not-found, not silently skip
// Regression: was continuing on not-found, reporting success with missing entries
// ---------------------------------------------------------------------------

func TestINV008_ReconcileEntries_NotFound_MustFailNotSkip(t *testing.T) {
	ctx := hardenCtx(t)
	missingID := uuid.New()

	repo := &mockTransactionRepo{}
	acctRepo := &mockAccountsRepo{}

	repo.On("GetEntryByID", mock.Anything, missingID).Return(nil, assert.AnError)

	svc := service.NewTransactionEntryService(
		repo, acctRepo,
		tracing.NewNoOpService(),
		metrics.NewNoOpMetricsProvider(),
	)
	err := svc.ReconcileEntries(ctx, []uuid.UUID{missingID}, "REF-001")

	require.Error(t, err, "not-found entry during reconciliation must fail — not silently skip")
}

// ---------------------------------------------------------------------------
// INV-009: Cross-tenant posting must be blocked
// ---------------------------------------------------------------------------

func TestINV009_PostTransaction_CrossTenantAccess_MustFail(t *testing.T) {
	callerTenantID := uuid.New()
	otherTenantID := uuid.New()

	ctx := shared.WithTenantID(context.Background(), callerTenantID)
	ctx = shared.WithUserID(ctx, uuid.New())

	txnID := uuid.New()
	txn := &domain.Transaction{
		ID: txnID, TenantID: otherTenantID, // different tenant
		TransactionNumber: "TXN-XTEN-001",
		TransactionType:   domain.TransactionTypeManual,
		TransactionStatus: domain.TransactionStatusDraft,
		ApprovalRequired:  false, ApprovalStatus: domain.ApprovalStatusNotRequired,
		TransactionDate: time.Now(), CurrencyCode: "USD", ExchangeRate: decimal.NewFromInt(1),
		TotalDebitAmount: decimal.NewFromInt(100), TotalCreditAmount: decimal.NewFromInt(100),
		CreatedBy: uuid.New(),
	}

	repo := &mockTransactionRepo{}
	acctRepo := &mockAccountsRepo{}
	entrySvc := &mockEntryService{}

	repo.On("GetByID", mock.Anything, txnID).Return(txn, nil)

	svc := newHardenSvc(repo, acctRepo, nil, nil, entrySvc)
	_, err := svc.PostTransaction(ctx, txnID, nil)

	require.Error(t, err, "cross-tenant posting must be blocked")
	assert.Contains(t, err.Error(), "TENANT_MISMATCH")
}

// ---------------------------------------------------------------------------
// CONC-001: Concurrent PostTransaction on same ID — no panic
// ---------------------------------------------------------------------------

func TestCONC001_ConcurrentPost_SameTxn_NoPanic(t *testing.T) {
	ctx := hardenCtx(t)
	tenantID, _ := shared.GetTenantID(ctx)

	txnID := uuid.New()
	txn := &domain.Transaction{
		ID: txnID, TenantID: tenantID,
		TransactionNumber: "TXN-CONC-001",
		TransactionType:   domain.TransactionTypeManual,
		TransactionStatus: domain.TransactionStatusDraft,
		ApprovalRequired:  false, ApprovalStatus: domain.ApprovalStatusNotRequired,
		TransactionDate: time.Now(), Description: "concurrent",
		CurrencyCode: "USD", ExchangeRate: decimal.NewFromInt(1),
		TotalDebitAmount: decimal.NewFromInt(100), TotalCreditAmount: decimal.NewFromInt(100),
		CreatedBy: uuid.New(),
	}
	entries := hardenBalancedEntries(txnID, tenantID, 100)

	repo := &mockTransactionRepo{}
	acctRepo := &mockAccountsRepo{}
	entrySvc := &mockEntryService{}

	repo.On("GetByID", mock.Anything, txnID).Return(txn, nil)
	repo.On("Approve", mock.Anything, txnID, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	repo.On("Post", mock.Anything, txnID, mock.Anything, mock.Anything).Return(nil)
	entrySvc.On("GetEntriesByTransactionID", mock.Anything, txnID).Return(entries, nil)
	for _, e := range entries {
		acctRepo.On("GetByID", mock.Anything, e.AccountID).Return(hardenActiveAccount(), nil)
	}
	// updateAccountBalances is non-fatal; return error so it logs and continues
	acctRepo.On("GetAccountBalance", mock.Anything, mock.Anything, mock.Anything).
		Return((*domain.AccountBalance)(nil), assert.AnError)

	svc := newHardenSvc(repo, acctRepo, nil, nil, entrySvc)

	const n = 10
	var wg sync.WaitGroup
	panics := make(chan interface{}, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					panics <- r
				}
				wg.Done()
			}()
			//nolint:errcheck
			svc.PostTransaction(ctx, txnID, nil)
		}()
	}

	wg.Wait()
	close(panics)
	for p := range panics {
		t.Errorf("concurrent PostTransaction panicked: %v", p)
	}
}

// mockPeriodRepo is defined in period_test.go (same package).
