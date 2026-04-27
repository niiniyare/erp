package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"awo.so/internal/core/finance/domain"
	"awo.so/internal/core/finance/service"
	"awo.so/internal/shared"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

// ============================================================================
// Mock TransactionRepository
// ============================================================================

type mockTransactionRepo struct {
	mock.Mock
}

func (m *mockTransactionRepo) Create(ctx context.Context, t *domain.Transaction) error {
	return m.Called(ctx, t).Error(0)
}

func (m *mockTransactionRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error) {
	args := m.Called(ctx, id)
	if v, ok := args.Get(0).(*domain.Transaction); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockTransactionRepo) GetByNumber(ctx context.Context, entityID *uuid.UUID, number string) (*domain.Transaction, error) {
	args := m.Called(ctx, entityID, number)
	if v, ok := args.Get(0).(*domain.Transaction); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockTransactionRepo) Update(ctx context.Context, t *domain.Transaction) error {
	return m.Called(ctx, t).Error(0)
}

func (m *mockTransactionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockTransactionRepo) List(ctx context.Context, filter *domain.TransactionFilter) ([]*domain.Transaction, error) {
	args := m.Called(ctx, filter)
	if v, ok := args.Get(0).([]*domain.Transaction); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockTransactionRepo) Count(ctx context.Context, filter *domain.TransactionFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockTransactionRepo) ListByAccount(ctx context.Context, accountID uuid.UUID, filter *domain.TransactionFilter) ([]*domain.Transaction, error) {
	args := m.Called(ctx, accountID, filter)
	if v, ok := args.Get(0).([]*domain.Transaction); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockTransactionRepo) ListByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*domain.Transaction, error) {
	args := m.Called(ctx, startDate, endDate)
	if v, ok := args.Get(0).([]*domain.Transaction); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockTransactionRepo) GetByStatus(ctx context.Context, status domain.TransactionStatus, limit int) ([]*domain.Transaction, error) {
	args := m.Called(ctx, status, limit)
	if v, ok := args.Get(0).([]*domain.Transaction); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockTransactionRepo) GetPendingApproval(ctx context.Context, userID *uuid.UUID) ([]*domain.Transaction, error) {
	args := m.Called(ctx, userID)
	if v, ok := args.Get(0).([]*domain.Transaction); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockTransactionRepo) GetRecurringTransactions(ctx context.Context, dueDate time.Time) ([]*domain.Transaction, error) {
	args := m.Called(ctx, dueDate)
	if v, ok := args.Get(0).([]*domain.Transaction); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockTransactionRepo) UpdateNextRecurringDate(ctx context.Context, transactionID uuid.UUID, nextDate time.Time) error {
	return m.Called(ctx, transactionID, nextDate).Error(0)
}

func (m *mockTransactionRepo) CreateEntry(ctx context.Context, entry *domain.TransactionEntry) error {
	return m.Called(ctx, entry).Error(0)
}

func (m *mockTransactionRepo) CreateEntries(ctx context.Context, entries []*domain.TransactionEntry) error {
	return m.Called(ctx, entries).Error(0)
}

func (m *mockTransactionRepo) GetEntryByID(ctx context.Context, id uuid.UUID) (*domain.TransactionEntry, error) {
	args := m.Called(ctx, id)
	if v, ok := args.Get(0).(*domain.TransactionEntry); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockTransactionRepo) GetEntriesByTransaction(ctx context.Context, transactionID uuid.UUID) ([]domain.TransactionEntry, error) {
	args := m.Called(ctx, transactionID)
	if v, ok := args.Get(0).([]domain.TransactionEntry); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockTransactionRepo) GetEntriesByAccount(ctx context.Context, accountID uuid.UUID, filter *domain.EntryFilter) ([]domain.TransactionEntry, error) {
	args := m.Called(ctx, accountID, filter)
	if v, ok := args.Get(0).([]domain.TransactionEntry); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockTransactionRepo) UpdateEntry(ctx context.Context, entry *domain.TransactionEntry) error {
	return m.Called(ctx, entry).Error(0)
}

func (m *mockTransactionRepo) DeleteEntry(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockTransactionRepo) SearchEntries(ctx context.Context, query string, filters *domain.EntryFilter, limit int, offset int) ([]*domain.TransactionEntry, error) {
	args := m.Called(ctx, query, filters, limit, offset)
	if v, ok := args.Get(0).([]*domain.TransactionEntry); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockTransactionRepo) UpdateReconciliationStatus(ctx context.Context, entryID uuid.UUID, reconciled bool, reconciledDate *time.Time, reconciliationRef *string) error {
	return m.Called(ctx, entryID, reconciled, reconciledDate, reconciliationRef).Error(0)
}

func (m *mockTransactionRepo) GetUnreconciledEntries(ctx context.Context, accountID uuid.UUID, cutoffDate *time.Time) ([]*domain.TransactionEntry, error) {
	args := m.Called(ctx, accountID, cutoffDate)
	if v, ok := args.Get(0).([]*domain.TransactionEntry); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockTransactionRepo) GetEntrySummary(ctx context.Context, accountID uuid.UUID, startDate, endDate time.Time) (*domain.TransactionSummary, error) {
	args := m.Called(ctx, accountID, startDate, endDate)
	if v, ok := args.Get(0).(*domain.TransactionSummary); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockTransactionRepo) CalculateAccountBalance(ctx context.Context, accountID uuid.UUID, asOfDate *time.Time) (decimal.Decimal, error) {
	args := m.Called(ctx, accountID, asOfDate)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

func (m *mockTransactionRepo) GetAccountTransactionSummary(ctx context.Context, accountID uuid.UUID, dateRange *domain.DateRange) (*domain.TransactionSummary, error) {
	args := m.Called(ctx, accountID, dateRange)
	if v, ok := args.Get(0).(*domain.TransactionSummary); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockTransactionRepo) Post(ctx context.Context, transactionID uuid.UUID, postedBy uuid.UUID, postedAt time.Time) error {
	return m.Called(ctx, transactionID, postedBy, postedAt).Error(0)
}

func (m *mockTransactionRepo) Approve(ctx context.Context, transactionID uuid.UUID, approvedBy uuid.UUID, approvedAt time.Time, notes *string) error {
	return m.Called(ctx, transactionID, approvedBy, approvedAt, notes).Error(0)
}

func (m *mockTransactionRepo) Reject(ctx context.Context, transactionID uuid.UUID, rejectedBy uuid.UUID, rejectedAt time.Time, reason domain.RejectionReason, notes *string) error {
	return m.Called(ctx, transactionID, rejectedBy, rejectedAt, reason, notes).Error(0)
}

func (m *mockTransactionRepo) Reverse(ctx context.Context, originalID, reversalID uuid.UUID, reason string) error {
	return m.Called(ctx, originalID, reversalID, reason).Error(0)
}

func (m *mockTransactionRepo) IsTransactionNumberUnique(ctx context.Context, entityID *uuid.UUID, transactionNumber string, excludeID *uuid.UUID) (bool, error) {
	args := m.Called(ctx, entityID, transactionNumber, excludeID)
	return args.Bool(0), args.Error(1)
}

func (m *mockTransactionRepo) ValidateAccountsExist(ctx context.Context, accountIDs []uuid.UUID) error {
	return m.Called(ctx, accountIDs).Error(0)
}

func (m *mockTransactionRepo) GetNextTransactionNumber(ctx context.Context, entityID *uuid.UUID, transactionType domain.TransactionType) (string, error) {
	args := m.Called(ctx, entityID, transactionType)
	return args.String(0), args.Error(1)
}

// ============================================================================
// Mock AccountsRepository
// ============================================================================

type mockAccountsRepo struct {
	mock.Mock
}

func (m *mockAccountsRepo) Create(ctx context.Context, account *domain.Accounts) error {
	return m.Called(ctx, account).Error(0)
}

func (m *mockAccountsRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Accounts, error) {
	args := m.Called(ctx, id)
	if v, ok := args.Get(0).(*domain.Accounts); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAccountsRepo) GetByCode(ctx context.Context, entityID *uuid.UUID, accountCode string) (*domain.Accounts, error) {
	args := m.Called(ctx, entityID, accountCode)
	if v, ok := args.Get(0).(*domain.Accounts); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAccountsRepo) Update(ctx context.Context, account *domain.Accounts) error {
	return m.Called(ctx, account).Error(0)
}

func (m *mockAccountsRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockAccountsRepo) List(ctx context.Context, filter *domain.AccountFilter) ([]*domain.Accounts, error) {
	args := m.Called(ctx, filter)
	if v, ok := args.Get(0).([]*domain.Accounts); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAccountsRepo) Count(ctx context.Context, filter *domain.AccountFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockAccountsRepo) ListByParent(ctx context.Context, parentID uuid.UUID) ([]*domain.Accounts, error) {
	args := m.Called(ctx, parentID)
	if v, ok := args.Get(0).([]*domain.Accounts); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAccountsRepo) ListByRootType(ctx context.Context, rootType domain.RootType) ([]*domain.Accounts, error) {
	args := m.Called(ctx, rootType)
	if v, ok := args.Get(0).([]*domain.Accounts); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAccountsRepo) GetAccountHierarchy(ctx context.Context, rootID uuid.UUID) ([]*domain.Accounts, error) {
	args := m.Called(ctx, rootID)
	if v, ok := args.Get(0).([]*domain.Accounts); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAccountsRepo) GetAccountPath(ctx context.Context, accountID uuid.UUID) ([]domain.Accounts, error) {
	args := m.Called(ctx, accountID)
	if v, ok := args.Get(0).([]domain.Accounts); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAccountsRepo) ValidateHierarchy(ctx context.Context, accountID, parentID uuid.UUID) error {
	return m.Called(ctx, accountID, parentID).Error(0)
}

func (m *mockAccountsRepo) GetAccountBalance(ctx context.Context, accountID uuid.UUID, asOfDate *time.Time) (*domain.AccountBalance, error) {
	args := m.Called(ctx, accountID, asOfDate)
	if v, ok := args.Get(0).(*domain.AccountBalance); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAccountsRepo) GetAccountBalances(ctx context.Context, accountIDs []uuid.UUID, asOfDate *time.Time) ([]*domain.AccountBalance, error) {
	args := m.Called(ctx, accountIDs, asOfDate)
	if v, ok := args.Get(0).([]*domain.AccountBalance); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAccountsRepo) GetTrialBalance(ctx context.Context, entityID *uuid.UUID, asOfDate *time.Time) ([]*domain.TrialBalanceEntry, error) {
	args := m.Called(ctx, entityID, asOfDate)
	if v, ok := args.Get(0).([]*domain.TrialBalanceEntry); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAccountsRepo) GetActiveAccounts(ctx context.Context, entityID *uuid.UUID) ([]*domain.Accounts, error) {
	args := m.Called(ctx, entityID)
	if v, ok := args.Get(0).([]*domain.Accounts); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAccountsRepo) GetControlAccounts(ctx context.Context, entityID *uuid.UUID) ([]*domain.Accounts, error) {
	args := m.Called(ctx, entityID)
	if v, ok := args.Get(0).([]*domain.Accounts); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAccountsRepo) GetAccountsByType(ctx context.Context, accountType string, rootType *domain.RootType) ([]*domain.Accounts, error) {
	args := m.Called(ctx, accountType, rootType)
	if v, ok := args.Get(0).([]*domain.Accounts); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAccountsRepo) Search(ctx context.Context, query string, limit int) ([]*domain.Accounts, error) {
	args := m.Called(ctx, query, limit)
	if v, ok := args.Get(0).([]*domain.Accounts); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAccountsRepo) ValidateAccountCode(ctx context.Context, code string, excludeID *uuid.UUID) error {
	return m.Called(ctx, code, excludeID).Error(0)
}

func (m *mockAccountsRepo) IsAccountCodeUnique(ctx context.Context, entityID *uuid.UUID, accountCode string, excludeID *uuid.UUID) (bool, error) {
	args := m.Called(ctx, entityID, accountCode, excludeID)
	return args.Bool(0), args.Error(1)
}

func (m *mockAccountsRepo) HasChildren(ctx context.Context, accountID uuid.UUID) (bool, error) {
	args := m.Called(ctx, accountID)
	return args.Bool(0), args.Error(1)
}

func (m *mockAccountsRepo) GetChildren(ctx context.Context, accountID uuid.UUID) ([]*domain.Accounts, error) {
	args := m.Called(ctx, accountID)
	if v, ok := args.Get(0).([]*domain.Accounts); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAccountsRepo) HasTransactions(ctx context.Context, accountID uuid.UUID) (bool, error) {
	args := m.Called(ctx, accountID)
	return args.Bool(0), args.Error(1)
}

func (m *mockAccountsRepo) UpdateBalance(ctx context.Context, accountID uuid.UUID, balance domain.AccountBalance) error {
	return m.Called(ctx, accountID, balance).Error(0)
}

func (m *mockAccountsRepo) CreateAccountGroup(ctx context.Context, group *domain.AccountGroup) error {
	return m.Called(ctx, group).Error(0)
}

func (m *mockAccountsRepo) GetAccountGroupByID(ctx context.Context, id uuid.UUID) (*domain.AccountGroup, error) {
	args := m.Called(ctx, id)
	if v, ok := args.Get(0).(*domain.AccountGroup); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAccountsRepo) GetAccountGroupByCode(ctx context.Context, code string, entityID *uuid.UUID) (*domain.AccountGroup, error) {
	args := m.Called(ctx, code, entityID)
	if v, ok := args.Get(0).(*domain.AccountGroup); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAccountsRepo) UpdateAccountGroup(ctx context.Context, id uuid.UUID, group *domain.AccountGroup) error {
	return m.Called(ctx, id, group).Error(0)
}

func (m *mockAccountsRepo) DeleteAccountGroup(ctx context.Context, id uuid.UUID, entityID *uuid.UUID) error {
	return m.Called(ctx, id, entityID).Error(0)
}

func (m *mockAccountsRepo) ListAccountGroups(ctx context.Context, filter *domain.AccountGroupFilter) ([]*domain.AccountGroup, error) {
	args := m.Called(ctx, filter)
	if v, ok := args.Get(0).([]*domain.AccountGroup); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAccountsRepo) CountAccountGroups(ctx context.Context, filter *domain.AccountGroupFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockAccountsRepo) GetAccountGroupHierarchy(ctx context.Context, rootGroupID *uuid.UUID, entityID *uuid.UUID) ([]*domain.AccountGroup, error) {
	args := m.Called(ctx, rootGroupID, entityID)
	if v, ok := args.Get(0).([]*domain.AccountGroup); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAccountsRepo) GetGroupsByFinancialStatement(ctx context.Context, statementType string, entityID *uuid.UUID) ([]*domain.AccountGroup, error) {
	args := m.Called(ctx, statementType, entityID)
	if v, ok := args.Get(0).([]*domain.AccountGroup); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAccountsRepo) GetGroupsByCashFlowCategory(ctx context.Context, category string, entityID *uuid.UUID) ([]*domain.AccountGroup, error) {
	args := m.Called(ctx, category, entityID)
	if v, ok := args.Get(0).([]*domain.AccountGroup); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAccountsRepo) ValidateAccountGroupCode(ctx context.Context, code string, excludeID *uuid.UUID, entityID *uuid.UUID) error {
	return m.Called(ctx, code, excludeID, entityID).Error(0)
}

func (m *mockAccountsRepo) GetAccountChildrenHierarchy(ctx context.Context, parentAccountID uuid.UUID) ([]*domain.AccountHierarchy, error) {
	args := m.Called(ctx, parentAccountID)
	if v, ok := args.Get(0).([]*domain.AccountHierarchy); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAccountsRepo) GetAccountSubtree(ctx context.Context, accountID uuid.UUID) ([]*domain.AccountHierarchy, error) {
	args := m.Called(ctx, accountID)
	if v, ok := args.Get(0).([]*domain.AccountHierarchy); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAccountsRepo) GetAccountsWithRecentActivity(ctx context.Context, filter *domain.AccountActivityFilter) ([]*domain.AccountActivity, error) {
	args := m.Called(ctx, filter)
	if v, ok := args.Get(0).([]*domain.AccountActivity); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAccountsRepo) GetStaleAccountBalances(ctx context.Context, filter *domain.AccountActivityFilter) ([]*domain.AccountActivity, error) {
	args := m.Called(ctx, filter)
	if v, ok := args.Get(0).([]*domain.AccountActivity); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAccountsRepo) GetAccountActivitySummary(ctx context.Context, filter *domain.AccountActivityFilter) ([]*domain.AccountActivitySummary, error) {
	args := m.Called(ctx, filter)
	if v, ok := args.Get(0).([]*domain.AccountActivitySummary); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

// ============================================================================
// Mock TransactionEntryService
// ============================================================================

type mockEntryService struct {
	mock.Mock
}

func (m *mockEntryService) CreateEntry(ctx context.Context, entry *domain.TransactionEntry) error {
	return m.Called(ctx, entry).Error(0)
}

func (m *mockEntryService) CreateEntries(ctx context.Context, entries []*domain.TransactionEntry) error {
	return m.Called(ctx, entries).Error(0)
}

func (m *mockEntryService) GetEntryByID(ctx context.Context, id uuid.UUID) (*domain.TransactionEntry, error) {
	args := m.Called(ctx, id)
	if v, ok := args.Get(0).(*domain.TransactionEntry); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockEntryService) GetEntriesByTransactionID(ctx context.Context, transactionID uuid.UUID) ([]*domain.TransactionEntry, error) {
	args := m.Called(ctx, transactionID)
	if v, ok := args.Get(0).([]*domain.TransactionEntry); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockEntryService) UpdateEntry(ctx context.Context, id uuid.UUID, req domain.TransactionEntry) (*domain.TransactionEntry, error) {
	args := m.Called(ctx, id, req)
	if v, ok := args.Get(0).(*domain.TransactionEntry); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockEntryService) DeleteEntry(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockEntryService) GetEntriesByAccountID(ctx context.Context, accountID uuid.UUID, limit int, offset int) ([]*domain.TransactionEntry, error) {
	args := m.Called(ctx, accountID, limit, offset)
	if v, ok := args.Get(0).([]*domain.TransactionEntry); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockEntryService) SearchEntries(ctx context.Context, query string, filters *domain.EntryFilter, limit int, offset int) ([]*domain.TransactionEntry, error) {
	args := m.Called(ctx, query, filters, limit, offset)
	if v, ok := args.Get(0).([]*domain.TransactionEntry); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockEntryService) ReconcileEntries(ctx context.Context, entryIDs []uuid.UUID, reconciliationRef string) error {
	return m.Called(ctx, entryIDs, reconciliationRef).Error(0)
}

func (m *mockEntryService) UnreconcileEntries(ctx context.Context, entryIDs []uuid.UUID) error {
	return m.Called(ctx, entryIDs).Error(0)
}

func (m *mockEntryService) GetUnreconciledEntries(ctx context.Context, accountID uuid.UUID, cutoffDate *time.Time) ([]*domain.TransactionEntry, error) {
	args := m.Called(ctx, accountID, cutoffDate)
	if v, ok := args.Get(0).([]*domain.TransactionEntry); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockEntryService) ValidateEntryConsistency(ctx context.Context, entry *domain.TransactionEntry) ([]domain.ValidationError, error) {
	args := m.Called(ctx, entry)
	if v, ok := args.Get(0).([]domain.ValidationError); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockEntryService) GetEntrySummary(ctx context.Context, accountID uuid.UUID, startDate, endDate time.Time) (*domain.TransactionSummary, error) {
	args := m.Called(ctx, accountID, startDate, endDate)
	if v, ok := args.Get(0).(*domain.TransactionSummary); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

// ============================================================================
// Suite
// ============================================================================

type TransactionServiceSuite struct {
	suite.Suite
	req         *require.Assertions
	repo        *mockTransactionRepo
	accountRepo *mockAccountsRepo
	entrySvc    *mockEntryService
	svc         service.TransactionService
	tenantID    uuid.UUID
	ctx         context.Context
}

func TestTransactionServiceSuite(t *testing.T) {
	suite.Run(t, new(TransactionServiceSuite))
}

func (s *TransactionServiceSuite) SetupTest() {
	s.req = require.New(s.T())
	s.tenantID = uuid.New()
	s.ctx = shared.WithTenantID(context.Background(), s.tenantID)

	s.repo = new(mockTransactionRepo)
	s.accountRepo = new(mockAccountsRepo)
	s.entrySvc = new(mockEntryService)

	s.svc = service.NewTransactionService(
		s.repo,
		s.accountRepo,
		s.entrySvc,
		tracing.NewNoOpService(),
		metrics.NewNoOpMetricsProvider(),
	)
}

func (s *TransactionServiceSuite) TearDownTest() {
	s.repo.AssertExpectations(s.T())
	s.accountRepo.AssertExpectations(s.T())
	s.entrySvc.AssertExpectations(s.T())
}

// ============================================================================
// Helpers
// ============================================================================

// newDraftTxn builds a minimal DRAFT transaction suitable for service tests.
func (s *TransactionServiceSuite) newDraftTxn() *domain.Transaction {
	return &domain.Transaction{
		ID:                uuid.New(),
		TenantID:          s.tenantID,
		TransactionNumber: "JE-2025-00001",
		TransactionType:   domain.TransactionTypeJournalEntry,
		TransactionStatus: domain.TransactionStatusDraft,
		TransactionDate:   time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		Description:       "Test journal entry",
		CurrencyCode:      "KES",
		ExchangeRate:      decimal.NewFromInt(1),
		ApprovalRequired:  false,
		ApprovalStatus:    domain.ApprovalStatusNotRequired,
		CreatedBy:         uuid.New(),
	}
}

// newApprovedTxn builds a minimal APPROVED transaction.
func (s *TransactionServiceSuite) newApprovedTxn() *domain.Transaction {
	txn := s.newDraftTxn()
	txn.TransactionStatus = domain.TransactionStatusApproved
	return txn
}

// newPostedTxn builds a minimal POSTED transaction.
func (s *TransactionServiceSuite) newPostedTxn() *domain.Transaction {
	txn := s.newDraftTxn()
	txn.TransactionStatus = domain.TransactionStatusPosted
	now := time.Now()
	txn.PostedAt = &now
	return txn
}

// balancedEntries returns two balanced journal entries for txnID across two accounts.
func (s *TransactionServiceSuite) balancedEntries(txnID, acct1ID, acct2ID uuid.UUID) []*domain.TransactionEntry {
	amount := decimal.NewFromInt(50_000)
	return []*domain.TransactionEntry{
		{
			ID:            uuid.New(),
			TenantID:      s.tenantID,
			TransactionID: txnID,
			EntryNumber:   1,
			AccountID:     acct1ID,
			DebitAmount:   amount,
			CreditAmount:  decimal.Zero,
			Description:   "Cash debit",
			ExchangeRate:  decimal.NewFromInt(1),
		},
		{
			ID:            uuid.New(),
			TenantID:      s.tenantID,
			TransactionID: txnID,
			EntryNumber:   2,
			AccountID:     acct2ID,
			DebitAmount:   decimal.Zero,
			CreditAmount:  amount,
			Description:   "Revenue credit",
			ExchangeRate:  decimal.NewFromInt(1),
		},
	}
}

// activeAccount builds a minimal active Accounts record.
func activeAccount(id uuid.UUID) *domain.Accounts {
	return &domain.Accounts{
		ID:       id,
		IsActive: true,
		Status:   domain.AccountStatusActive,
		RootType: domain.RootTypeAsset,
	}
}

// ============================================================================
// FIN-TXN-001: CreateTransaction — minimum valid request succeeds
// ============================================================================

func (s *TransactionServiceSuite) TestCreateTransaction_MinimalValid() {
	acct1 := uuid.New()
	acct2 := uuid.New()
	amount := decimal.NewFromInt(50_000)

	req := domain.CreateTransactionRequest{
		TransactionNumber: "JE-2025-00001",
		TransactionType:   domain.TransactionTypeJournalEntry,
		TransactionDate:   time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC),
		Description:       "Minimal valid journal entry",
		CurrencyCode:      "KES",
		ExchangeRate:      decimal.NewFromInt(1),
		Entries: []domain.CreateEntryRequest{
			{AccountID: acct1, DebitAmount: amount, CreditAmount: decimal.Zero, Description: "Cash debit", ExchangeRate: decimal.NewFromInt(1)},
			{AccountID: acct2, DebitAmount: decimal.Zero, CreditAmount: amount, Description: "Revenue credit", ExchangeRate: decimal.NewFromInt(1)},
		},
		CreatedBy: uuid.New(),
	}

	s.repo.On("IsTransactionNumberUnique", s.ctx, req.EntityID, req.TransactionNumber, (*uuid.UUID)(nil)).
		Return(true, nil)
	s.repo.On("Create", s.ctx, mock.AnythingOfType("*domain.Transaction")).
		Return(nil)

	result, err := s.svc.CreateTransaction(s.ctx, req)
	s.req.NoError(err)
	s.req.NotNil(result)
	s.req.Equal(domain.TransactionStatusDraft, result.TransactionStatus)
	s.req.Equal("JE-2025-00001", result.TransactionNumber)
}

// ============================================================================
// FIN-TXN-002: CreateTransaction — duplicate number rejected
// ============================================================================

func (s *TransactionServiceSuite) TestCreateTransaction_DuplicateNumber() {
	amount := decimal.NewFromInt(50_000)
	req := domain.CreateTransactionRequest{
		TransactionNumber: "JE-2025-00001",
		TransactionType:   domain.TransactionTypeJournalEntry,
		TransactionDate:   time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC),
		Description:       "Duplicate number test",
		CurrencyCode:      "KES",
		ExchangeRate:      decimal.NewFromInt(1),
		Entries: []domain.CreateEntryRequest{
			{AccountID: uuid.New(), DebitAmount: amount, CreditAmount: decimal.Zero, Description: "Dr", ExchangeRate: decimal.NewFromInt(1)},
			{AccountID: uuid.New(), DebitAmount: decimal.Zero, CreditAmount: amount, Description: "Cr", ExchangeRate: decimal.NewFromInt(1)},
		},
	}

	// Repo returns a conflict error — simulates a DB unique-constraint violation.
	s.repo.On("IsTransactionNumberUnique", s.ctx, req.EntityID, req.TransactionNumber, (*uuid.UUID)(nil)).
		Return(false, errors.New("duplicate transaction number"))

	result, err := s.svc.CreateTransaction(s.ctx, req)
	s.req.Error(err)
	s.req.Nil(result)
}

// ============================================================================
// FIN-TXN-010: PostTransaction — balanced APPROVED transaction posts
// ============================================================================

func (s *TransactionServiceSuite) TestPostTransaction_BalancedSuccess() {
	txn := s.newApprovedTxn()
	acct1ID := uuid.New()
	acct2ID := uuid.New()
	entries := s.balancedEntries(txn.ID, acct1ID, acct2ID)

	// After posting, GetByID returns a POSTED copy
	postedTxn := *txn
	postedTxn.TransactionStatus = domain.TransactionStatusPosted

	s.repo.On("GetByID", s.ctx, txn.ID).Return(txn, nil).Once()
	s.entrySvc.On("GetEntriesByTransactionID", s.ctx, txn.ID).Return(entries, nil)
	// ValidateTransaction calls accountRepo.GetByID for each entry's account
	s.accountRepo.On("GetByID", s.ctx, acct1ID).Return(activeAccount(acct1ID), nil)
	s.accountRepo.On("GetByID", s.ctx, acct2ID).Return(activeAccount(acct2ID), nil)
	s.repo.On("Post", s.ctx, txn.ID, txn.CreatedBy, mock.AnythingOfType("time.Time")).Return(nil)
	s.repo.On("GetByID", s.ctx, txn.ID).Return(&postedTxn, nil).Once()
	// updateAccountBalances calls GetAccountBalance then UpdateBalance — allow both.
	s.accountRepo.On("GetAccountBalance", mock.Anything, mock.Anything, mock.Anything).Return((*domain.AccountBalance)(nil), errors.New("balance unavailable")).Maybe()
	s.accountRepo.On("UpdateBalance", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

	result, err := s.svc.PostTransaction(s.ctx, txn.ID, nil)
	s.req.NoError(err)
	s.req.NotNil(result)
	s.req.Equal(domain.TransactionStatusPosted, result.TransactionStatus)
}

// ============================================================================
// FIN-TXN-011: PostTransaction — unbalanced transaction rejected
// ============================================================================

func (s *TransactionServiceSuite) TestPostTransaction_Unbalanced() {
	txn := s.newApprovedTxn()
	acct1ID := uuid.New()
	acct2ID := uuid.New()

	// Intentionally unbalanced: Dr 50000 vs Cr 40000
	unbalanced := []*domain.TransactionEntry{
		{
			ID:            uuid.New(),
			TenantID:      s.tenantID,
			TransactionID: txn.ID,
			EntryNumber:   1,
			AccountID:     acct1ID,
			DebitAmount:   decimal.NewFromInt(50_000),
			CreditAmount:  decimal.Zero,
			Description:   "Rent debit",
		},
		{
			ID:            uuid.New(),
			TenantID:      s.tenantID,
			TransactionID: txn.ID,
			EntryNumber:   2,
			AccountID:     acct2ID,
			DebitAmount:   decimal.Zero,
			CreditAmount:  decimal.NewFromInt(40_000),
			Description:   "Cash credit",
		},
	}

	s.repo.On("GetByID", s.ctx, txn.ID).Return(txn, nil)
	s.entrySvc.On("GetEntriesByTransactionID", s.ctx, txn.ID).Return(unbalanced, nil)
	// ValidateTransaction still calls accountRepo for each account
	s.accountRepo.On("GetByID", s.ctx, acct1ID).Return(activeAccount(acct1ID), nil)
	s.accountRepo.On("GetByID", s.ctx, acct2ID).Return(activeAccount(acct2ID), nil)

	result, err := s.svc.PostTransaction(s.ctx, txn.ID, nil)
	s.req.Error(err, "unbalanced transaction must be rejected")
	s.req.Nil(result)
}

// ============================================================================
// FIN-TXN-013: PostTransaction — already-posted transaction is rejected
// ============================================================================

func (s *TransactionServiceSuite) TestPostTransaction_AlreadyPosted() {
	txn := s.newPostedTxn()

	s.repo.On("GetByID", s.ctx, txn.ID).Return(txn, nil)

	result, err := s.svc.PostTransaction(s.ctx, txn.ID, nil)
	s.req.Error(err, "posting an already-POSTED transaction must return an error")
	s.req.Nil(result)
}

// ============================================================================
// FIN-TXN-014: PostTransaction — no entries → rejected
// ============================================================================

func (s *TransactionServiceSuite) TestPostTransaction_NoEntries() {
	txn := s.newApprovedTxn()

	s.repo.On("GetByID", s.ctx, txn.ID).Return(txn, nil)
	s.entrySvc.On("GetEntriesByTransactionID", s.ctx, txn.ID).Return([]*domain.TransactionEntry{}, nil)

	result, err := s.svc.PostTransaction(s.ctx, txn.ID, nil)
	s.req.Error(err)
	s.req.Nil(result)
}

// ============================================================================
// FIN-TXN-022: ReverseTransaction — only POSTED transactions can be reversed
// ============================================================================

func (s *TransactionServiceSuite) TestReverseTransaction_RequiresPosted_Draft() {
	txn := s.newDraftTxn()

	s.repo.On("GetByID", s.ctx, txn.ID).Return(txn, nil)

	result, err := s.svc.ReverseTransaction(s.ctx, txn.ID, "correcting error")
	s.req.Error(err, "DRAFT transaction must not be reversible")
	s.req.Nil(result)
}

func (s *TransactionServiceSuite) TestReverseTransaction_RequiresPosted_Approved() {
	txn := s.newApprovedTxn()

	s.repo.On("GetByID", s.ctx, txn.ID).Return(txn, nil)

	result, err := s.svc.ReverseTransaction(s.ctx, txn.ID, "correcting error")
	s.req.Error(err, "APPROVED transaction must not be reversible")
	s.req.Nil(result)
}

// ============================================================================
// FIN-TXN-023: ReverseTransaction — double reversal blocked (IsReversed=true)
// ============================================================================

func (s *TransactionServiceSuite) TestReverseTransaction_AlreadyReversed() {
	txn := s.newPostedTxn()
	txn.IsReversed = true
	reversalID := uuid.New()
	txn.ReversedByTransactionID = &reversalID

	s.repo.On("GetByID", s.ctx, txn.ID).Return(txn, nil)

	result, err := s.svc.ReverseTransaction(s.ctx, txn.ID, "duplicate attempt")
	s.req.Error(err, "already-reversed transaction must return an error")
	s.req.Nil(result)
}

// ============================================================================
// FIN-TXN-030: ApproveTransaction — requires ApprovalRequired+ApprovalStatus=Pending
// ============================================================================

func (s *TransactionServiceSuite) TestApproveTransaction_NotPending_ReturnsError() {
	// Transaction does not have approval required
	txn := s.newDraftTxn()
	txn.ApprovalRequired = false
	txn.ApprovalStatus = domain.ApprovalStatusNotRequired

	s.repo.On("GetByID", s.ctx, txn.ID).Return(txn, nil)

	result, err := s.svc.ApproveTransaction(s.ctx, txn.ID, "approve notes")
	s.req.Error(err, "transaction not pending approval must return an error")
	s.req.Nil(result)
}

func (s *TransactionServiceSuite) TestApproveTransaction_PendingApproval_Succeeds() {
	txn := s.newDraftTxn()
	txn.TransactionStatus = domain.TransactionStatusPendingApproval
	txn.ApprovalRequired = true
	txn.ApprovalStatus = domain.ApprovalStatusPending

	approvedTxn := *txn
	approvedTxn.TransactionStatus = domain.TransactionStatusApproved
	approvedTxn.ApprovalStatus = domain.ApprovalStatusApproved

	s.repo.On("GetByID", s.ctx, txn.ID).Return(txn, nil).Once()
	s.repo.On("Approve", s.ctx, txn.ID, txn.CreatedBy, mock.AnythingOfType("time.Time"), mock.Anything).Return(nil)
	s.repo.On("GetByID", s.ctx, txn.ID).Return(&approvedTxn, nil).Once()

	result, err := s.svc.ApproveTransaction(s.ctx, txn.ID, "looks good")
	s.req.NoError(err)
	s.req.NotNil(result)
	s.req.Equal(domain.TransactionStatusApproved, result.TransactionStatus)
}

// ============================================================================
// FIN-TXN-031: RejectTransaction — requires ApprovalRequired+ApprovalStatus=Pending
// ============================================================================

func (s *TransactionServiceSuite) TestRejectTransaction_NotPending_ReturnsError() {
	txn := s.newDraftTxn()
	txn.ApprovalRequired = false
	txn.ApprovalStatus = domain.ApprovalStatusNotRequired

	s.repo.On("GetByID", s.ctx, txn.ID).Return(txn, nil)

	result, err := s.svc.RejectTransaction(s.ctx, txn.ID, "reject notes")
	s.req.Error(err)
	s.req.Nil(result)
}

func (s *TransactionServiceSuite) TestRejectTransaction_PendingApproval_Succeeds() {
	txn := s.newDraftTxn()
	txn.TransactionStatus = domain.TransactionStatusPendingApproval
	txn.ApprovalRequired = true
	txn.ApprovalStatus = domain.ApprovalStatusPending

	rejectedTxn := *txn
	rejectedTxn.TransactionStatus = domain.TransactionStatusRejected
	rejectedTxn.ApprovalStatus = domain.ApprovalStatusRejected

	s.repo.On("GetByID", s.ctx, txn.ID).Return(txn, nil).Once()
	s.repo.On("Reject", s.ctx, txn.ID, txn.CreatedBy, mock.AnythingOfType("time.Time"), domain.RejectionReasonOther, mock.Anything).Return(nil)
	s.repo.On("GetByID", s.ctx, txn.ID).Return(&rejectedTxn, nil).Once()

	result, err := s.svc.RejectTransaction(s.ctx, txn.ID, "incorrect account")
	s.req.NoError(err)
	s.req.NotNil(result)
	s.req.Equal(domain.TransactionStatusRejected, result.TransactionStatus)
}

// ============================================================================
// GetTransactionWithEntries — assembles header and entries
// ============================================================================

func (s *TransactionServiceSuite) TestGetTransactionWithEntries_Success() {
	txn := s.newPostedTxn()
	entries := []domain.TransactionEntry{
		{ID: uuid.New(), TransactionID: txn.ID, DebitAmount: decimal.NewFromInt(50_000)},
		{ID: uuid.New(), TransactionID: txn.ID, CreditAmount: decimal.NewFromInt(50_000)},
	}

	s.repo.On("GetByID", s.ctx, txn.ID).Return(txn, nil)
	s.repo.On("GetEntriesByTransaction", s.ctx, txn.ID).Return(entries, nil)

	result, err := s.svc.GetTransactionWithEntries(s.ctx, txn.ID)
	s.req.NoError(err)
	s.req.NotNil(result)
	s.req.Equal(txn.ID, result.Transaction.ID)
	s.req.Len(result.Entries, 2)
}

// ============================================================================
// FIN-TXN-003: CreateTransaction — fewer than 2 entries rejected
// ============================================================================

func (s *TransactionServiceSuite) TestCreateTransaction_TooFewEntries() {
	amount := decimal.NewFromInt(50_000)
	req := domain.CreateTransactionRequest{
		TransactionNumber: "JE-2025-00002",
		TransactionType:   domain.TransactionTypeJournalEntry,
		TransactionDate:   time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC),
		Description:       "Only one entry",
		CurrencyCode:      "KES",
		ExchangeRate:      decimal.NewFromInt(1),
		Entries: []domain.CreateEntryRequest{
			{
				AccountID:    uuid.New(),
				DebitAmount:  amount,
				CreditAmount: decimal.Zero,
				Description:  "Cash debit",
				ExchangeRate: decimal.NewFromInt(1),
			},
		},
		CreatedBy: uuid.New(),
	}

	// req.Validate() fails with INSUFFICIENT_ENTRIES; no repo calls must be made.
	result, err := s.svc.CreateTransaction(s.ctx, req)
	s.req.Error(err)
	s.req.Nil(result)
}

// ============================================================================
// FIN-TXN-015: PostTransaction — inactive account blocks posting
// ============================================================================

func (s *TransactionServiceSuite) TestPostTransaction_InactiveAccount() {
	txn := s.newApprovedTxn()
	activeID := uuid.New()
	inactiveID := uuid.New()
	amount := decimal.NewFromInt(50_000)

	entries := []*domain.TransactionEntry{
		{
			ID:            uuid.New(),
			TenantID:      s.tenantID,
			TransactionID: txn.ID,
			EntryNumber:   1,
			AccountID:     activeID,
			DebitAmount:   amount,
			CreditAmount:  decimal.Zero,
			Description:   "Cash debit",
			ExchangeRate:  decimal.NewFromInt(1),
		},
		{
			ID:            uuid.New(),
			TenantID:      s.tenantID,
			TransactionID: txn.ID,
			EntryNumber:   2,
			AccountID:     inactiveID,
			DebitAmount:   decimal.Zero,
			CreditAmount:  amount,
			Description:   "Revenue credit",
			ExchangeRate:  decimal.NewFromInt(1),
		},
	}

	inactiveAcct := &domain.Accounts{
		ID:       inactiveID,
		IsActive: false,
		Status:   domain.AccountStatusInactive,
		RootType: domain.RootTypeAsset,
	}

	s.repo.On("GetByID", s.ctx, txn.ID).Return(txn, nil)
	s.entrySvc.On("GetEntriesByTransactionID", s.ctx, txn.ID).Return(entries, nil)
	s.accountRepo.On("GetByID", s.ctx, activeID).Return(activeAccount(activeID), nil)
	s.accountRepo.On("GetByID", s.ctx, inactiveID).Return(inactiveAcct, nil)
	// repo.Post must NOT be called; absence of an On registration causes panic if called.

	result, err := s.svc.PostTransaction(s.ctx, txn.ID, nil)
	s.req.Error(err, "posting to an inactive account must return an error")
	s.req.Nil(result)
}

// ============================================================================
// FIN-TXN-020: ReverseTransaction — creates mirror entries and posts reversal
// ============================================================================

func (s *TransactionServiceSuite) TestReverseTransaction_MirrorEntries() {
	origTxn := s.newPostedTxn()
	origTxn.IsReversed = false
	acct1ID := uuid.New()
	acct2ID := uuid.New()

	origEntries := s.balancedEntries(origTxn.ID, acct1ID, acct2ID)

	// Reversal entries returned by the mock when postTransactionInline validates the reversal.
	// Amounts are swapped: original Dr→Cr, original Cr→Dr.
	// TransactionID must be non-nil to pass entry.Validate().
	amount := decimal.NewFromInt(50_000)
	revTxnID := uuid.New() // placeholder; matches the dynamic reversal transaction ID
	revEntries := []*domain.TransactionEntry{
		{
			ID:            uuid.New(),
			TenantID:      s.tenantID,
			TransactionID: revTxnID,
			EntryNumber:   1,
			AccountID:     acct1ID,
			DebitAmount:   decimal.Zero,
			CreditAmount:  amount,
			Description:   "REVERSAL: Cash debit",
			ExchangeRate:  decimal.NewFromInt(1),
		},
		{
			ID:            uuid.New(),
			TenantID:      s.tenantID,
			TransactionID: revTxnID,
			EntryNumber:   2,
			AccountID:     acct2ID,
			DebitAmount:   amount,
			CreditAmount:  decimal.Zero,
			Description:   "REVERSAL: Revenue credit",
			ExchangeRate:  decimal.NewFromInt(1),
		},
	}

	draftReversal := &domain.Transaction{
		ID:                uuid.New(),
		TenantID:          s.tenantID,
		TransactionNumber: "REV-JE-2025-00001",
		TransactionType:   domain.TransactionTypeJournalEntry,
		TransactionStatus: domain.TransactionStatusDraft,
		CurrencyCode:      "KES",
		ExchangeRate:      decimal.NewFromInt(1),
		ApprovalRequired:  false,
		ApprovalStatus:    domain.ApprovalStatusNotRequired,
	}
	postedReversal := *draftReversal
	postedReversal.TransactionStatus = domain.TransactionStatusPosted

	// Step 1: load original transaction
	s.repo.On("GetByID", s.ctx, origTxn.ID).Return(origTxn, nil).Once()
	// Step 2: get original entries for cloning
	s.entrySvc.On("GetEntriesByTransactionID", s.ctx, origTxn.ID).Return(origEntries, nil).Once()
	// Step 3: persist the new reversal transaction header
	s.repo.On("Create", s.ctx, mock.AnythingOfType("*domain.Transaction")).Return(nil).Once()
	// Step 4: persist each reversal entry (2 entries)
	s.entrySvc.On("CreateEntry", s.ctx, mock.AnythingOfType("*domain.TransactionEntry")).Return(nil).Times(2)
	// Step 5: mark original as reversed
	s.repo.On("Reverse", s.ctx, origTxn.ID, mock.Anything, "correcting error").Return(nil).Once()
	// Step 6: PostTransaction(reversalTransaction.ID) → postTransactionInline
	//   6a: fetch reversal header (DRAFT)
	s.repo.On("GetByID", s.ctx, mock.Anything).Return(draftReversal, nil).Once()
	//   6b: fetch reversal entries for validation
	s.entrySvc.On("GetEntriesByTransactionID", s.ctx, mock.Anything).Return(revEntries, nil).Once()
	//   6c: validate each entry's account
	s.accountRepo.On("GetByID", s.ctx, acct1ID).Return(activeAccount(acct1ID), nil)
	s.accountRepo.On("GetByID", s.ctx, acct2ID).Return(activeAccount(acct2ID), nil)
	//   6d: write POSTED status to DB
	s.repo.On("Post", s.ctx, mock.Anything, mock.Anything, mock.AnythingOfType("time.Time")).Return(nil).Once()
	//   6e: re-fetch reversal as POSTED
	s.repo.On("GetByID", s.ctx, mock.Anything).Return(&postedReversal, nil).Once()
	//   6f: updateAccountBalances calls GetAccountBalance then UpdateBalance — both are best-effort
	s.accountRepo.On("GetAccountBalance", mock.Anything, mock.Anything, mock.Anything).Return((*domain.AccountBalance)(nil), errors.New("balance unavailable")).Maybe()
	s.accountRepo.On("UpdateBalance", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

	result, err := s.svc.ReverseTransaction(s.ctx, origTxn.ID, "correcting error")
	s.req.NoError(err)
	s.req.NotNil(result)
}

// ============================================================================
// FIN-TXN-041: GetTransactionWithEntries — IsBalanced reflects entry sums
// ============================================================================

func (s *TransactionServiceSuite) TestGetTransactionWithEntries_IsBalanced() {
	amount := decimal.NewFromInt(50_000)

	// Balanced transaction: set header totals so IsBalanced() returns true.
	txn := s.newPostedTxn()
	txn.TotalDebitAmount = amount
	txn.TotalCreditAmount = amount

	entries := []domain.TransactionEntry{
		{ID: uuid.New(), TransactionID: txn.ID, DebitAmount: amount, CreditAmount: decimal.Zero},
		{ID: uuid.New(), TransactionID: txn.ID, DebitAmount: decimal.Zero, CreditAmount: amount},
	}

	s.repo.On("GetByID", s.ctx, txn.ID).Return(txn, nil)
	s.repo.On("GetEntriesByTransaction", s.ctx, txn.ID).Return(entries, nil)

	result, err := s.svc.GetTransactionWithEntries(s.ctx, txn.ID)
	s.req.NoError(err)
	s.req.NotNil(result)

	// Verify header reports balanced via TotalDebitAmount/TotalCreditAmount.
	s.req.True(result.Transaction.IsBalanced(), "transaction header must report IsBalanced=true")

	// Also verify by summing entry amounts directly.
	var totalDr, totalCr decimal.Decimal
	for _, e := range result.Entries {
		totalDr = totalDr.Add(e.DebitAmount)
		totalCr = totalCr.Add(e.CreditAmount)
	}
	s.req.True(totalDr.Equal(totalCr), "entry debits must equal entry credits")
}

// ============================================================================
// FIN-TXN-050/051: SearchTransactions — current stub returns empty slice
// ============================================================================

func (s *TransactionServiceSuite) TestSearchTransactions_CurrentBehavior() {
	// TODO: update this test once Search is implemented in the repo interface.
	// Current implementation is a stub that returns an empty slice with no error.
	results, err := s.svc.SearchTransactions(s.ctx, "rent", 10, 0)
	s.req.NoError(err)
	s.req.NotNil(results)
	s.req.Empty(results)
}
