// Package service_test contains behavioural unit tests for the finance service
// layer. Tests prove orchestration correctness, idempotency, state-machine
// enforcement, and financial invariants without a live database.
//
// All stubs are hand-crafted (no mockgen required) and panic on unexpected
// method calls so test failures surface immediately.
package service_test

import (
	"context"
	"errors"
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
// Shared stub types
// ============================================================================

// stubTxnRepo is a minimal TransactionRepository stub. Only methods populated
// via function fields are live; all others panic so unexpected calls fail loudly.
type stubTxnRepo struct {
	fnGetByID              func(ctx context.Context, id uuid.UUID) (*domain.Transaction, error)
	fnCreate               func(ctx context.Context, t *domain.Transaction) error
	fnUpdate               func(ctx context.Context, t *domain.Transaction) error
	fnDelete               func(ctx context.Context, id uuid.UUID) error
	fnList                 func(ctx context.Context, f *domain.TransactionFilter) ([]*domain.Transaction, error)
	fnPost                 func(ctx context.Context, id uuid.UUID, by uuid.UUID, at time.Time) error
	fnApprove              func(ctx context.Context, id uuid.UUID, by uuid.UUID, at time.Time, notes *string) error
	fnReject               func(ctx context.Context, id uuid.UUID, by uuid.UUID, at time.Time, r domain.RejectionReason, notes *string) error
	fnReverse              func(ctx context.Context, origID, revID uuid.UUID, reason string) error
	fnIsUnique             func(ctx context.Context, eid *uuid.UUID, num string, excl *uuid.UUID) (bool, error)
	fnGetByNumber          func(ctx context.Context, eid *uuid.UUID, num string) (*domain.Transaction, error)
	fnGetEntriesByTxn      func(ctx context.Context, id uuid.UUID) ([]domain.TransactionEntry, error)
}

func (s *stubTxnRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error) {
	if s.fnGetByID != nil {
		return s.fnGetByID(ctx, id)
	}
	panic("stubTxnRepo.GetByID not expected")
}
func (s *stubTxnRepo) Create(ctx context.Context, t *domain.Transaction) error {
	if s.fnCreate != nil {
		return s.fnCreate(ctx, t)
	}
	panic("stubTxnRepo.Create not expected")
}
func (s *stubTxnRepo) Update(ctx context.Context, t *domain.Transaction) error {
	if s.fnUpdate != nil {
		return s.fnUpdate(ctx, t)
	}
	panic("stubTxnRepo.Update not expected")
}
func (s *stubTxnRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if s.fnDelete != nil {
		return s.fnDelete(ctx, id)
	}
	panic("stubTxnRepo.Delete not expected")
}
func (s *stubTxnRepo) List(ctx context.Context, f *domain.TransactionFilter) ([]*domain.Transaction, error) {
	if s.fnList != nil {
		return s.fnList(ctx, f)
	}
	panic("stubTxnRepo.List not expected")
}
func (s *stubTxnRepo) Count(ctx context.Context, f *domain.TransactionFilter) (int64, error) {
	panic("stubTxnRepo.Count not expected")
}
func (s *stubTxnRepo) ListByAccount(ctx context.Context, accountID uuid.UUID, f *domain.TransactionFilter) ([]*domain.Transaction, error) {
	panic("stubTxnRepo.ListByAccount not expected")
}
func (s *stubTxnRepo) ListByDateRange(ctx context.Context, start, end time.Time) ([]*domain.Transaction, error) {
	panic("stubTxnRepo.ListByDateRange not expected")
}
func (s *stubTxnRepo) GetByStatus(ctx context.Context, status domain.TransactionStatus, limit int) ([]*domain.Transaction, error) {
	panic("stubTxnRepo.GetByStatus not expected")
}
func (s *stubTxnRepo) GetPendingApproval(ctx context.Context, userID *uuid.UUID) ([]*domain.Transaction, error) {
	panic("stubTxnRepo.GetPendingApproval not expected")
}
func (s *stubTxnRepo) GetRecurringTransactions(ctx context.Context, due time.Time) ([]*domain.Transaction, error) {
	panic("stubTxnRepo.GetRecurringTransactions not expected")
}
func (s *stubTxnRepo) UpdateNextRecurringDate(ctx context.Context, id uuid.UUID, next time.Time) error {
	panic("stubTxnRepo.UpdateNextRecurringDate not expected")
}
func (s *stubTxnRepo) CreateEntry(ctx context.Context, e *domain.TransactionEntry) error {
	panic("stubTxnRepo.CreateEntry not expected")
}
func (s *stubTxnRepo) CreateEntries(ctx context.Context, ee []*domain.TransactionEntry) error {
	panic("stubTxnRepo.CreateEntries not expected")
}
func (s *stubTxnRepo) GetEntryByID(ctx context.Context, id uuid.UUID) (*domain.TransactionEntry, error) {
	panic("stubTxnRepo.GetEntryByID not expected")
}
func (s *stubTxnRepo) GetEntriesByTransaction(ctx context.Context, id uuid.UUID) ([]domain.TransactionEntry, error) {
	if s.fnGetEntriesByTxn != nil {
		return s.fnGetEntriesByTxn(ctx, id)
	}
	panic("stubTxnRepo.GetEntriesByTransaction not expected")
}
func (s *stubTxnRepo) GetEntriesByAccount(ctx context.Context, accountID uuid.UUID, f *domain.EntryFilter) ([]domain.TransactionEntry, error) {
	panic("stubTxnRepo.GetEntriesByAccount not expected")
}
func (s *stubTxnRepo) UpdateEntry(ctx context.Context, e *domain.TransactionEntry) error {
	panic("stubTxnRepo.UpdateEntry not expected")
}
func (s *stubTxnRepo) DeleteEntry(ctx context.Context, id uuid.UUID) error {
	panic("stubTxnRepo.DeleteEntry not expected")
}
func (s *stubTxnRepo) SearchEntries(ctx context.Context, q string, f *domain.EntryFilter, limit, offset int) ([]*domain.TransactionEntry, error) {
	panic("stubTxnRepo.SearchEntries not expected")
}
func (s *stubTxnRepo) UpdateReconciliationStatus(ctx context.Context, id uuid.UUID, reconciled bool, date *time.Time, ref *string) error {
	panic("stubTxnRepo.UpdateReconciliationStatus not expected")
}
func (s *stubTxnRepo) GetUnreconciledEntries(ctx context.Context, accountID uuid.UUID, cutoff *time.Time) ([]*domain.TransactionEntry, error) {
	panic("stubTxnRepo.GetUnreconciledEntries not expected")
}
func (s *stubTxnRepo) GetEntrySummary(ctx context.Context, accountID uuid.UUID, start, end time.Time) (*domain.TransactionSummary, error) {
	panic("stubTxnRepo.GetEntrySummary not expected")
}
func (s *stubTxnRepo) CalculateAccountBalance(ctx context.Context, accountID uuid.UUID, asOf *time.Time) (decimal.Decimal, error) {
	panic("stubTxnRepo.CalculateAccountBalance not expected")
}
func (s *stubTxnRepo) GetAccountTransactionSummary(ctx context.Context, accountID uuid.UUID, dr *domain.DateRange) (*domain.TransactionSummary, error) {
	panic("stubTxnRepo.GetAccountTransactionSummary not expected")
}
func (s *stubTxnRepo) Post(ctx context.Context, id uuid.UUID, by uuid.UUID, at time.Time) error {
	if s.fnPost != nil {
		return s.fnPost(ctx, id, by, at)
	}
	panic("stubTxnRepo.Post not expected")
}
func (s *stubTxnRepo) Approve(ctx context.Context, id uuid.UUID, by uuid.UUID, at time.Time, notes *string) error {
	if s.fnApprove != nil {
		return s.fnApprove(ctx, id, by, at, notes)
	}
	panic("stubTxnRepo.Approve not expected")
}
func (s *stubTxnRepo) Reject(ctx context.Context, id uuid.UUID, by uuid.UUID, at time.Time, r domain.RejectionReason, notes *string) error {
	if s.fnReject != nil {
		return s.fnReject(ctx, id, by, at, r, notes)
	}
	panic("stubTxnRepo.Reject not expected")
}
func (s *stubTxnRepo) Reverse(ctx context.Context, origID, revID uuid.UUID, reason string) error {
	if s.fnReverse != nil {
		return s.fnReverse(ctx, origID, revID, reason)
	}
	panic("stubTxnRepo.Reverse not expected")
}
func (s *stubTxnRepo) IsTransactionNumberUnique(ctx context.Context, eid *uuid.UUID, num string, excl *uuid.UUID) (bool, error) {
	if s.fnIsUnique != nil {
		return s.fnIsUnique(ctx, eid, num, excl)
	}
	panic("stubTxnRepo.IsTransactionNumberUnique not expected")
}
func (s *stubTxnRepo) ValidateAccountsExist(ctx context.Context, ids []uuid.UUID) error {
	panic("stubTxnRepo.ValidateAccountsExist not expected")
}
func (s *stubTxnRepo) GetNextTransactionNumber(ctx context.Context, eid *uuid.UUID, tt domain.TransactionType) (string, error) {
	panic("stubTxnRepo.GetNextTransactionNumber not expected")
}
func (s *stubTxnRepo) GetByNumber(ctx context.Context, eid *uuid.UUID, num string) (*domain.Transaction, error) {
	if s.fnGetByNumber != nil {
		return s.fnGetByNumber(ctx, eid, num)
	}
	panic("stubTxnRepo.GetByNumber not expected")
}

// stubAccountRepo is a minimal AccountsRepository stub.
type stubAccountRepo struct {
	fnGetByID          func(ctx context.Context, id uuid.UUID) (*domain.Accounts, error)
	fnGetAccountBalance func(ctx context.Context, id uuid.UUID, asOf *time.Time) (*domain.AccountBalance, error)
	fnUpdateBalance     func(ctx context.Context, id uuid.UUID, b domain.AccountBalance) error
}

func (s *stubAccountRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Accounts, error) {
	if s.fnGetByID != nil {
		return s.fnGetByID(ctx, id)
	}
	panic("stubAccountRepo.GetByID not expected")
}

// All other AccountsRepository methods panic — not needed for these tests.
func (s *stubAccountRepo) Create(ctx context.Context, a *domain.Accounts) error              { panic("not expected") }
func (s *stubAccountRepo) GetByCode(ctx context.Context, eid *uuid.UUID, code string) (*domain.Accounts, error) { panic("not expected") }
func (s *stubAccountRepo) Update(ctx context.Context, a *domain.Accounts) error              { panic("not expected") }
func (s *stubAccountRepo) Delete(ctx context.Context, id uuid.UUID) error                    { panic("not expected") }
func (s *stubAccountRepo) List(ctx context.Context, f *domain.AccountFilter) ([]*domain.Accounts, error) { panic("not expected") }
func (s *stubAccountRepo) Count(ctx context.Context, f *domain.AccountFilter) (int64, error) { panic("not expected") }
func (s *stubAccountRepo) ListByParent(ctx context.Context, pid uuid.UUID) ([]*domain.Accounts, error) { panic("not expected") }
func (s *stubAccountRepo) ListByRootType(ctx context.Context, rt domain.RootType) ([]*domain.Accounts, error) { panic("not expected") }
func (s *stubAccountRepo) GetAccountHierarchy(ctx context.Context, rootID uuid.UUID) ([]*domain.Accounts, error) { panic("not expected") }
func (s *stubAccountRepo) GetAccountPath(ctx context.Context, id uuid.UUID) ([]domain.Accounts, error) { panic("not expected") }
func (s *stubAccountRepo) ValidateHierarchy(ctx context.Context, id, pid uuid.UUID) error    { panic("not expected") }
func (s *stubAccountRepo) GetAccountBalance(ctx context.Context, id uuid.UUID, asOf *time.Time) (*domain.AccountBalance, error) {
	if s.fnGetAccountBalance != nil {
		return s.fnGetAccountBalance(ctx, id, asOf)
	}
	panic("not expected")
}
func (s *stubAccountRepo) GetAccountBalances(ctx context.Context, ids []uuid.UUID, asOf *time.Time) ([]*domain.AccountBalance, error) { panic("not expected") }
func (s *stubAccountRepo) GetTrialBalance(ctx context.Context, eid *uuid.UUID, asOf *time.Time) ([]*domain.TrialBalanceEntry, error) { panic("not expected") }
func (s *stubAccountRepo) GetActiveAccounts(ctx context.Context, eid *uuid.UUID) ([]*domain.Accounts, error) { panic("not expected") }
func (s *stubAccountRepo) GetControlAccounts(ctx context.Context, eid *uuid.UUID) ([]*domain.Accounts, error) { panic("not expected") }
func (s *stubAccountRepo) GetAccountsByType(ctx context.Context, t string, rt *domain.RootType) ([]*domain.Accounts, error) { panic("not expected") }
func (s *stubAccountRepo) Search(ctx context.Context, q string, limit int) ([]*domain.Accounts, error) { panic("not expected") }
func (s *stubAccountRepo) ValidateAccountCode(ctx context.Context, code string, excl *uuid.UUID) error { panic("not expected") }
func (s *stubAccountRepo) IsAccountCodeUnique(ctx context.Context, eid *uuid.UUID, code string, excl *uuid.UUID) (bool, error) { panic("not expected") }
func (s *stubAccountRepo) HasChildren(ctx context.Context, id uuid.UUID) (bool, error)        { panic("not expected") }
func (s *stubAccountRepo) GetChildren(ctx context.Context, id uuid.UUID) ([]*domain.Accounts, error) { panic("not expected") }
func (s *stubAccountRepo) HasTransactions(ctx context.Context, id uuid.UUID) (bool, error)    { panic("not expected") }
func (s *stubAccountRepo) UpdateBalance(ctx context.Context, id uuid.UUID, b domain.AccountBalance) error {
	if s.fnUpdateBalance != nil {
		return s.fnUpdateBalance(ctx, id, b)
	}
	panic("not expected")
}
func (s *stubAccountRepo) CreateAccountGroup(ctx context.Context, g *domain.AccountGroup) error { panic("not expected") }
func (s *stubAccountRepo) GetAccountGroupByID(ctx context.Context, id uuid.UUID) (*domain.AccountGroup, error) { panic("not expected") }
func (s *stubAccountRepo) GetAccountGroupByCode(ctx context.Context, code string, eid *uuid.UUID) (*domain.AccountGroup, error) { panic("not expected") }
func (s *stubAccountRepo) UpdateAccountGroup(ctx context.Context, id uuid.UUID, g *domain.AccountGroup) error { panic("not expected") }
func (s *stubAccountRepo) DeleteAccountGroup(ctx context.Context, id uuid.UUID, eid *uuid.UUID) error { panic("not expected") }
func (s *stubAccountRepo) ListAccountGroups(ctx context.Context, f *domain.AccountGroupFilter) ([]*domain.AccountGroup, error) { panic("not expected") }
func (s *stubAccountRepo) CountAccountGroups(ctx context.Context, f *domain.AccountGroupFilter) (int64, error) { panic("not expected") }
func (s *stubAccountRepo) GetAccountGroupHierarchy(ctx context.Context, rootID *uuid.UUID, eid *uuid.UUID) ([]*domain.AccountGroup, error) { panic("not expected") }
func (s *stubAccountRepo) GetGroupsByFinancialStatement(ctx context.Context, st string, eid *uuid.UUID) ([]*domain.AccountGroup, error) { panic("not expected") }
func (s *stubAccountRepo) GetGroupsByCashFlowCategory(ctx context.Context, cat string, eid *uuid.UUID) ([]*domain.AccountGroup, error) { panic("not expected") }
func (s *stubAccountRepo) ValidateAccountGroupCode(ctx context.Context, code string, excl *uuid.UUID, eid *uuid.UUID) error { panic("not expected") }
func (s *stubAccountRepo) GetAccountChildrenHierarchy(ctx context.Context, pid uuid.UUID) ([]*domain.AccountHierarchy, error) { panic("not expected") }
func (s *stubAccountRepo) GetAccountSubtree(ctx context.Context, id uuid.UUID) ([]*domain.AccountHierarchy, error) { panic("not expected") }
func (s *stubAccountRepo) GetAccountsWithRecentActivity(ctx context.Context, f *domain.AccountActivityFilter) ([]*domain.AccountActivity, error) { panic("not expected") }
func (s *stubAccountRepo) GetStaleAccountBalances(ctx context.Context, f *domain.AccountActivityFilter) ([]*domain.AccountActivity, error) { panic("not expected") }
func (s *stubAccountRepo) GetAccountActivitySummary(ctx context.Context, f *domain.AccountActivityFilter) ([]*domain.AccountActivitySummary, error) { panic("not expected") }

// stubEntryService is a minimal TransactionEntryService stub.
type stubEntryService struct {
	fnGetEntriesByTxnID func(ctx context.Context, id uuid.UUID) ([]*domain.TransactionEntry, error)
}

func (s *stubEntryService) CreateEntry(ctx context.Context, e *domain.TransactionEntry) error { panic("not expected") }
func (s *stubEntryService) CreateEntries(ctx context.Context, ee []*domain.TransactionEntry) error { panic("not expected") }
func (s *stubEntryService) GetEntryByID(ctx context.Context, id uuid.UUID) (*domain.TransactionEntry, error) { panic("not expected") }
func (s *stubEntryService) GetEntriesByTransactionID(ctx context.Context, id uuid.UUID) ([]*domain.TransactionEntry, error) {
	if s.fnGetEntriesByTxnID != nil {
		return s.fnGetEntriesByTxnID(ctx, id)
	}
	panic("stubEntryService.GetEntriesByTransactionID not expected")
}
func (s *stubEntryService) UpdateEntry(ctx context.Context, id uuid.UUID, req domain.TransactionEntry) (*domain.TransactionEntry, error) { panic("not expected") }
func (s *stubEntryService) DeleteEntry(ctx context.Context, id uuid.UUID) error { panic("not expected") }
func (s *stubEntryService) GetEntriesByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]*domain.TransactionEntry, error) { panic("not expected") }
func (s *stubEntryService) SearchEntries(ctx context.Context, q string, f *domain.EntryFilter, limit, offset int) ([]*domain.TransactionEntry, error) { panic("not expected") }
func (s *stubEntryService) ReconcileEntries(ctx context.Context, ids []uuid.UUID, ref string) error { panic("not expected") }
func (s *stubEntryService) UnreconcileEntries(ctx context.Context, ids []uuid.UUID) error { panic("not expected") }
func (s *stubEntryService) GetUnreconciledEntries(ctx context.Context, accountID uuid.UUID, cutoff *time.Time) ([]*domain.TransactionEntry, error) { panic("not expected") }
func (s *stubEntryService) ValidateEntryConsistency(ctx context.Context, entry *domain.TransactionEntry) ([]domain.ValidationError, error) {
	panic("not expected")
}
func (s *stubEntryService) GetEntrySummary(ctx context.Context, accountID uuid.UUID, start, end time.Time) (*domain.TransactionSummary, error) { panic("not expected") }

// newSvc builds a TransactionService with stub deps.
// nil for unused optional params (auditWriter, safetyEnforcer, anomalyDetector).
func newSvc(repo domain.TransactionRepository, accountRepo domain.AccountsRepository, entryService service.TransactionEntryService) service.TransactionService {
	return service.NewTransactionService(
		repo,
		accountRepo,
		nil, // periodRepo — skips period check
		nil, // reversalHistoryRepo — skips reversal-of-reversal check
		entryService,
		nil, // txRunner — best-effort path
		tracing.NewNoOpService(),
		metrics.NewNoOpMetricsProvider(),
		nil, // auditWriter — nil-safe, audit skipped
		nil, // safetyEnforcer — nil-safe, skipped
		nil, // anomalyDetector — nil-safe, skipped
	)
}

// tenantCtx returns a context carrying a tenant ID.
func tenantCtx() context.Context {
	return shared.WithTenantID(context.Background(), uuid.New())
}

// tenantAndUserCtx returns a context with both tenant and user IDs.
func tenantAndUserCtx(userID uuid.UUID) context.Context {
	ctx := shared.WithTenantID(context.Background(), uuid.New())
	return shared.WithUserID(ctx, userID)
}

// ============================================================================
// FIN-SVC-001: PostTransaction idempotency — already POSTED returns existing
// ============================================================================

func TestPostTransaction_Idempotent_AlreadyPosted(t *testing.T) {
	txnID := uuid.New()
	tenantID := uuid.New()
	want := &domain.Transaction{
		ID:                txnID,
		TenantID:          tenantID,
		TransactionStatus: domain.TransactionStatusPosted,
	}

	repo := &stubTxnRepo{
		fnGetByID: func(_ context.Context, id uuid.UUID) (*domain.Transaction, error) {
			return want, nil
		},
	}

	svc := newSvc(repo, &stubAccountRepo{}, &stubEntryService{})
	ctx := shared.WithTenantID(context.Background(), tenantID)

	got, err := svc.PostTransaction(ctx, txnID, nil)
	if err != nil {
		t.Fatalf("expected nil error on idempotent post, got: %v", err)
	}
	if got.ID != txnID {
		t.Errorf("expected returned transaction ID %v, got %v", txnID, got.ID)
	}
	if got.TransactionStatus != domain.TransactionStatusPosted {
		t.Errorf("expected status POSTED, got %v", got.TransactionStatus)
	}
}

// ============================================================================
// FIN-SVC-002: ApproveTransaction idempotency — already APPROVED returns existing
// ============================================================================

func TestApproveTransaction_Idempotent_AlreadyApproved(t *testing.T) {
	txnID := uuid.New()
	approverID := uuid.New()
	want := &domain.Transaction{
		ID:             txnID,
		ApprovalStatus: domain.ApprovalStatusApproved,
		TenantID:       uuid.New(),
	}

	repo := &stubTxnRepo{
		fnGetByID: func(_ context.Context, _ uuid.UUID) (*domain.Transaction, error) {
			return want, nil
		},
		// Approve must NOT be called on an idempotent retry.
	}

	svc := newSvc(repo, &stubAccountRepo{}, &stubEntryService{})
	ctx := tenantAndUserCtx(approverID)

	got, err := svc.ApproveTransaction(ctx, txnID, "retry notes")
	if err != nil {
		t.Fatalf("expected nil error on idempotent approve, got: %v", err)
	}
	if got.ApprovalStatus != domain.ApprovalStatusApproved {
		t.Errorf("expected ApprovalStatusApproved, got %v", got.ApprovalStatus)
	}
}

// ============================================================================
// FIN-SVC-003: ApproveTransaction SOD — submitter cannot approve own transaction
// ============================================================================

func TestApproveTransaction_SODViolation(t *testing.T) {
	userID := uuid.New() // same user submits and tries to approve
	txnID := uuid.New()

	repo := &stubTxnRepo{
		fnGetByID: func(_ context.Context, _ uuid.UUID) (*domain.Transaction, error) {
			return &domain.Transaction{
				ID:              txnID,
				TenantID:        uuid.New(),
				CreatedBy:       userID, // same as caller
				ApprovalRequired: true,
				ApprovalStatus:  domain.ApprovalStatusPending,
			}, nil
		},
		// Approve must NOT be reached.
	}

	svc := newSvc(repo, &stubAccountRepo{}, &stubEntryService{})
	ctx := tenantAndUserCtx(userID) // same user ID

	_, err := svc.ApproveTransaction(ctx, txnID, "")
	if err == nil {
		t.Fatal("expected SOD_VIOLATION error, got nil")
	}
	if !containsCode(err, "SOD_VIOLATION") {
		t.Errorf("expected SOD_VIOLATION error, got: %v", err)
	}
}

// ============================================================================
// FIN-SVC-004: RejectTransaction SOD — submitter cannot reject own transaction
// ============================================================================

func TestRejectTransaction_SODViolation(t *testing.T) {
	userID := uuid.New()
	txnID := uuid.New()

	repo := &stubTxnRepo{
		fnGetByID: func(_ context.Context, _ uuid.UUID) (*domain.Transaction, error) {
			return &domain.Transaction{
				ID:              txnID,
				TenantID:        uuid.New(),
				CreatedBy:       userID, // same user
				ApprovalRequired: true,
				ApprovalStatus:  domain.ApprovalStatusPending,
			}, nil
		},
	}

	svc := newSvc(repo, &stubAccountRepo{}, &stubEntryService{})
	ctx := tenantAndUserCtx(userID)

	_, err := svc.RejectTransaction(ctx, txnID, domain.RejectionReasonInvalidAccount, "notes")
	if err == nil {
		t.Fatal("expected SOD_VIOLATION error, got nil")
	}
	if !containsCode(err, "SOD_VIOLATION") {
		t.Errorf("expected SOD_VIOLATION error, got: %v", err)
	}
}

// ============================================================================
// FIN-SVC-005: DeleteTransaction blocks POSTED (terminal) transactions
// ============================================================================

func TestDeleteTransaction_PostedStatus_Blocked(t *testing.T) {
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
		// Delete must NOT be called.
	}

	svc := newSvc(repo, &stubAccountRepo{}, &stubEntryService{})
	ctx := shared.WithTenantID(context.Background(), tenantID)

	err := svc.DeleteTransaction(ctx, txnID)
	if err == nil {
		t.Fatal("expected error when deleting POSTED transaction, got nil")
	}
	if !containsCode(err, "TRANSACTION_DELETE_NOT_ALLOWED") {
		t.Errorf("expected TRANSACTION_DELETE_NOT_ALLOWED, got: %v", err)
	}
}

// ============================================================================
// FIN-SVC-006: DeleteTransaction blocks REVERSED (terminal) transactions
// ============================================================================

func TestDeleteTransaction_ReversedStatus_Blocked(t *testing.T) {
	txnID := uuid.New()
	tenantID := uuid.New()

	repo := &stubTxnRepo{
		fnGetByID: func(_ context.Context, _ uuid.UUID) (*domain.Transaction, error) {
			return &domain.Transaction{
				ID:                txnID,
				TenantID:          tenantID,
				TransactionStatus: domain.TransactionStatusReversed,
			}, nil
		},
	}

	svc := newSvc(repo, &stubAccountRepo{}, &stubEntryService{})
	ctx := shared.WithTenantID(context.Background(), tenantID)

	err := svc.DeleteTransaction(ctx, txnID)
	if err == nil {
		t.Fatal("expected error when deleting REVERSED transaction, got nil")
	}
	if !containsCode(err, "TRANSACTION_DELETE_NOT_ALLOWED") {
		t.Errorf("expected TRANSACTION_DELETE_NOT_ALLOWED, got: %v", err)
	}
}

// ============================================================================
// FIN-SVC-007: UpdateTransaction blocks POSTED (not editable) transactions
// ============================================================================

func TestUpdateTransaction_PostedStatus_NotEditable(t *testing.T) {
	txnID := uuid.New()
	tenantID := uuid.New()

	repo := &stubTxnRepo{
		fnGetByID: func(_ context.Context, _ uuid.UUID) (*domain.Transaction, error) {
			return &domain.Transaction{
				ID:                txnID,
				TenantID:          tenantID,
				TransactionStatus: domain.TransactionStatusPosted,
				TransactionType:   domain.TransactionTypeManual,
				CurrencyCode:      "KES",
				TransactionDate:   time.Now(),
			}, nil
		},
	}

	svc := newSvc(repo, &stubAccountRepo{}, &stubEntryService{})
	ctx := shared.WithTenantID(context.Background(), tenantID)

	_, err := svc.UpdateTransaction(ctx, txnID, domain.Transaction{
		TransactionType: domain.TransactionTypeManual,
		CurrencyCode:    "KES",
		TransactionDate: time.Now(),
	})
	if err == nil {
		t.Fatal("expected error updating POSTED transaction, got nil")
	}
	if !containsCode(err, "TRANSACTION_NOT_EDITABLE") {
		t.Errorf("expected TRANSACTION_NOT_EDITABLE, got: %v", err)
	}
}

// ============================================================================
// FIN-SVC-008: PostTransaction blocks DRAFT requiring approval
// ============================================================================

func TestPostTransaction_ApprovalRequired_DraftBlocked(t *testing.T) {
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
		t.Fatal("expected APPROVAL_REQUIRED error, got nil")
	}
	if !containsCode(err, "APPROVAL_REQUIRED") {
		t.Errorf("expected APPROVAL_REQUIRED, got: %v", err)
	}
}

// ============================================================================
// FIN-SVC-009: PostTransaction auto-approves DRAFT with no approval required
// ============================================================================

func TestPostTransaction_AutoApproves_NoApprovalRequired(t *testing.T) {
	txnID := uuid.New()
	tenantID := uuid.New()
	createdBy := uuid.New()

	approveCallCount := 0
	postCallCount := 0

	txn := &domain.Transaction{
		ID:                txnID,
		TenantID:          tenantID,
		TransactionStatus: domain.TransactionStatusDraft,
		ApprovalRequired:  false,
		ApprovalStatus:    domain.ApprovalStatusNotRequired,
		CreatedBy:         createdBy,
		TransactionType:   domain.TransactionTypeSystem,
		Description:       "test auto-approve transaction",
		ExchangeRate:      decimal.NewFromFloat(1),
		ValidationStatus:  domain.ValidationStatusValid,
		CurrencyCode:      "KES",
		TransactionDate:   time.Now(),
	}

	repo := &stubTxnRepo{
		fnGetByID: func(_ context.Context, _ uuid.UUID) (*domain.Transaction, error) {
			return txn, nil
		},
		fnApprove: func(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ time.Time, _ *string) error {
			approveCallCount++
			return nil
		},
		fnPost: func(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ time.Time) error {
			postCallCount++
			return nil
		},
	}

	entryService := &stubEntryService{
		fnGetEntriesByTxnID: func(_ context.Context, _ uuid.UUID) ([]*domain.TransactionEntry, error) {
			// Return two balanced entries so ValidateTransaction passes.
			acc1, acc2 := uuid.New(), uuid.New()
			return []*domain.TransactionEntry{
				{
					ID: uuid.New(), TransactionID: txnID, AccountID: acc1,
					TenantID: tenantID, Description: "debit line",
					DebitAmount: decimal.NewFromFloat(100), CreditAmount: decimal.Zero,
					ExchangeRate: decimal.NewFromFloat(1), EntryNumber: 1,
				},
				{
					ID: uuid.New(), TransactionID: txnID, AccountID: acc2,
					TenantID: tenantID, Description: "credit line",
					CreditAmount: decimal.NewFromFloat(100), DebitAmount: decimal.Zero,
					ExchangeRate: decimal.NewFromFloat(1), EntryNumber: 2,
				},
			}, nil
		},
	}

	accountRepo := &stubAccountRepo{
		fnGetByID: func(_ context.Context, id uuid.UUID) (*domain.Accounts, error) {
			return &domain.Accounts{
				ID:                 id,
				IsActive:           true,
				Status:             domain.AccountStatusActive,
				AllowManualEntries: true,
				HasChildren:        false,
				ValidationStatus:   domain.ValidationStatusValid,
			}, nil
		},
		fnGetAccountBalance: func(_ context.Context, id uuid.UUID, _ *time.Time) (*domain.AccountBalance, error) {
			return &domain.AccountBalance{AccountID: id}, nil
		},
		fnUpdateBalance: func(_ context.Context, _ uuid.UUID, _ domain.AccountBalance) error {
			return nil
		},
	}

	svc := newSvc(repo, accountRepo, entryService)
	ctx := shared.WithTenantID(context.Background(), tenantID)

	// GetByID is called twice: once for the initial load and once after posting.
	// Override to return POSTED on the second call.
	callCount := 0
	repo.fnGetByID = func(_ context.Context, _ uuid.UUID) (*domain.Transaction, error) {
		callCount++
		if callCount == 1 {
			return txn, nil
		}
		posted := *txn
		posted.TransactionStatus = domain.TransactionStatusPosted
		return &posted, nil
	}

	got, err := svc.PostTransaction(ctx, txnID, nil)
	if err != nil {
		t.Fatalf("PostTransaction failed: %v", err)
	}
	if approveCallCount != 1 {
		t.Errorf("expected Approve called once (auto-approve), got %d", approveCallCount)
	}
	if postCallCount != 1 {
		t.Errorf("expected Post called once, got %d", postCallCount)
	}
	_ = got
}

// ============================================================================
// FIN-SVC-010: ReverseTransaction blocks already-reversed transactions
// ============================================================================

func TestReverseTransaction_AlreadyReversed(t *testing.T) {
	txnID := uuid.New()
	tenantID := uuid.New()

	repo := &stubTxnRepo{
		fnGetByID: func(_ context.Context, _ uuid.UUID) (*domain.Transaction, error) {
			return &domain.Transaction{
				ID:                txnID,
				TenantID:          tenantID,
				TransactionStatus: domain.TransactionStatusPosted,
				IsReversed:        true,
			}, nil
		},
	}

	svc := newSvc(repo, &stubAccountRepo{}, &stubEntryService{})
	ctx := shared.WithTenantID(context.Background(), tenantID)

	_, err := svc.ReverseTransaction(ctx, txnID, "duplicate reversal attempt")
	if err == nil {
		t.Fatal("expected ALREADY_REVERSED error, got nil")
	}
	if !containsCode(err, "ALREADY_REVERSED") {
		t.Errorf("expected ALREADY_REVERSED, got: %v", err)
	}
}

// ============================================================================
// FIN-SVC-011: ReverseTransaction blocks non-POSTED transactions
// ============================================================================

func TestReverseTransaction_NotPosted_Blocked(t *testing.T) {
	txnID := uuid.New()
	tenantID := uuid.New()

	for _, status := range []domain.TransactionStatus{
		domain.TransactionStatusDraft,
		domain.TransactionStatusApproved,
		domain.TransactionStatusCancelled,
	} {
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

		_, err := svc.ReverseTransaction(ctx, txnID, "attempt")
		if err == nil {
			t.Errorf("status=%s: expected NOT_POSTED error, got nil", status)
		}
	}
}

// ============================================================================
// FIN-SVC-012: ApproveTransaction rejects non-pending transactions
// ============================================================================

func TestApproveTransaction_NotPendingApproval_Rejected(t *testing.T) {
	txnID := uuid.New()
	approverID := uuid.New()

	for _, approvalStatus := range []domain.ApprovalStatus{
		domain.ApprovalStatusNotRequired,
		domain.ApprovalStatusRejected,
	} {
		repo := &stubTxnRepo{
			fnGetByID: func(_ context.Context, _ uuid.UUID) (*domain.Transaction, error) {
				return &domain.Transaction{
					ID:              txnID,
					TenantID:        uuid.New(),
					CreatedBy:       uuid.New(), // different from approver
					ApprovalRequired: approvalStatus != domain.ApprovalStatusNotRequired,
					ApprovalStatus:  approvalStatus,
				}, nil
			},
		}

		svc := newSvc(repo, &stubAccountRepo{}, &stubEntryService{})
		ctx := tenantAndUserCtx(approverID)

		_, err := svc.ApproveTransaction(ctx, txnID, "")
		if err == nil {
			t.Errorf("approvalStatus=%v: expected error, got nil", approvalStatus)
		}
	}
}

// ============================================================================
// FIN-SVC-013: ListTransactions enforces default pagination
// ============================================================================

func TestListTransactions_DefaultPaginationApplied(t *testing.T) {
	wantLimit := 50
	capturedLimit := 0

	repo := &stubTxnRepo{
		fnList: func(_ context.Context, f *domain.TransactionFilter) ([]*domain.Transaction, error) {
			if f.Limit != nil {
				capturedLimit = *f.Limit
			}
			return nil, nil
		},
	}

	svc := newSvc(repo, &stubAccountRepo{}, &stubEntryService{})
	ctx := tenantCtx()

	_, err := svc.ListTransactions(ctx, &domain.TransactionFilter{}) // no limit set
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedLimit != wantLimit {
		t.Errorf("expected default limit %d, got %d", wantLimit, capturedLimit)
	}
}

// ============================================================================
// FIN-SVC-014: ListTransactions clamps excessive limit
// ============================================================================

func TestListTransactions_ExcessiveLimitClamped(t *testing.T) {
	const excessive = 9999
	const maxLimit = 1000
	capturedLimit := 0

	repo := &stubTxnRepo{
		fnList: func(_ context.Context, f *domain.TransactionFilter) ([]*domain.Transaction, error) {
			if f.Limit != nil {
				capturedLimit = *f.Limit
			}
			return nil, nil
		},
	}

	svc := newSvc(repo, &stubAccountRepo{}, &stubEntryService{})
	ctx := tenantCtx()

	limit := excessive
	_, err := svc.ListTransactions(ctx, &domain.TransactionFilter{Limit: &limit})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedLimit > maxLimit {
		t.Errorf("expected limit clamped to %d, got %d", maxLimit, capturedLimit)
	}
}

// ============================================================================
// Helpers
// ============================================================================

// containsCode returns true if err's message contains code or err implements
// an interface with a Code() string method matching code.
func containsCode(err error, code string) bool {
	if err == nil {
		return false
	}
	type coder interface{ Code() string }
	var c coder
	if errors.As(err, &c) {
		return c.Code() == code
	}
	return containsString(err.Error(), code)
}

func containsString(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsRune(s, sub))
}

func containsRune(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
