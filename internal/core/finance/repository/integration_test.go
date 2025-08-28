//go:build database
// +build database

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"

	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/core/finance/domain"
	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/shared"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// FinanceRepositoryIntegrationTestSuite tests integration between chart of accounts and transactions
// with comprehensive tenant isolation verification
type FinanceRepositoryIntegrationTestSuite struct {
	suite.Suite
	ctx             context.Context
	runner          *tenant.DatabaseTestRunner
	accountRepo     domain.AccountsRepository
	transactionRepo domain.TransactionRepository
	tenantA         *db.Tenant
	tenantB         *db.Tenant

	// Test data cleanup tracking
	createdAccountIDs     []uuid.UUID
	createdTransactionIDs []uuid.UUID
}

// SetupSuite runs once before the entire test suite
func (s *FinanceRepositoryIntegrationTestSuite) SetupSuite() {
	var err error
	s.runner, err = tenant.NewDatabaseTestRunner()
	s.Require().NoError(err, "Failed to connect to the database")
	s.ctx = context.Background()

	// Create test tenants
	s.tenantA, err = s.runner.CreateTestTenant(s.ctx, "finance-integration-tenant-a")
	s.Require().NoError(err, "Failed to create tenant A")

	s.tenantB, err = s.runner.CreateTestTenant(s.ctx, "finance-integration-tenant-b")
	s.Require().NoError(err, "Failed to create tenant B")

	// Setup repositories
	traceService := tracing.NewNoOpTracingService()
	s.accountRepo = NewAccountsRepository(s.runner.GetStore(), traceService)
	s.transactionRepo = NewTransactionRepository(s.runner.GetStore(), traceService)
}

// SetupTest runs before each test
func (s *FinanceRepositoryIntegrationTestSuite) SetupTest() {
	s.createdAccountIDs = make([]uuid.UUID, 0)
	s.createdTransactionIDs = make([]uuid.UUID, 0)
}

// TearDownTest runs after each test
func (s *FinanceRepositoryIntegrationTestSuite) TearDownTest() {
	// Clean up created transactions first (foreign key dependencies)
	for _, transactionID := range s.createdTransactionIDs {
		ctxA := shared.WithTenantID(s.ctx, s.tenantA.ID)
		ctxB := shared.WithTenantID(s.ctx, s.tenantB.ID)

		_ = s.transactionRepo.Delete(ctxA, transactionID)
		_ = s.transactionRepo.Delete(ctxB, transactionID)
	}

	// Clean up created accounts
	for _, accountID := range s.createdAccountIDs {
		ctxA := shared.WithTenantID(s.ctx, s.tenantA.ID)
		ctxB := shared.WithTenantID(s.ctx, s.tenantB.ID)

		_ = s.accountRepo.Delete(ctxA, accountID)
		_ = s.accountRepo.Delete(ctxB, accountID)
	}
}

// TearDownSuite runs once after the entire test suite
func (s *FinanceRepositoryIntegrationTestSuite) TearDownSuite() {
	if s.runner != nil {
		s.runner.Close()
	}
}

// Test runner
func TestFinanceRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(FinanceRepositoryIntegrationTestSuite))
}

// TestCrossRepositoryTenantIsolation tests tenant isolation across both repositories
func (s *FinanceRepositoryIntegrationTestSuite) TestCrossRepositoryTenantIsolation() {
	// Create accounts in both tenants with same code (should be allowed due to tenant isolation)
	ctxA := shared.WithTenantID(s.ctx, s.tenantA.ID)
	accountA := &domain.Accounts{
		ID:                 uuid.New(),
		AccountCode:        "INTEGRATION-CASH",
		AccountName:        "Tenant A Cash Account",
		AccountDescription: stringPtr("Cash account for tenant A"),
		RootType:           domain.RootTypeAsset,
		AccountType:        "CASH",
		NormalBalance:      domain.NormalBalanceDebit,
		CurrencyCode:       "USD",
		IsActive:           true,
		AllowManualEntries: true,
	}

	err := s.accountRepo.Create(ctxA, accountA)
	s.Require().NoError(err)
	s.createdAccountIDs = append(s.createdAccountIDs, accountA.ID)

	ctxB := shared.WithTenantID(s.ctx, s.tenantB.ID)
	accountB := &domain.Accounts{
		ID:                 uuid.New(),
		AccountCode:        "INTEGRATION-CASH", // Same code, different tenant
		AccountName:        "Tenant B Cash Account",
		AccountDescription: stringPtr("Cash account for tenant B"),
		RootType:           domain.RootTypeAsset,
		AccountType:        "CASH",
		NormalBalance:      domain.NormalBalanceDebit,
		CurrencyCode:       "USD",
		IsActive:           true,
		AllowManualEntries: true,
	}

	err = s.accountRepo.Create(ctxB, accountB)
	s.Require().NoError(err)
	s.createdAccountIDs = append(s.createdAccountIDs, accountB.ID)

	// Create transactions in both tenants with same reference (should be allowed due to tenant isolation)
	transactionA := &domain.Transaction{
		ID:                uuid.New(),
		TransactionDate:   time.Now(),
		TransactionType:   domain.TransactionTypeManual,
		TransactionStatus: domain.TransactionStatusPending,
		ReferenceNumber:   stringPtr("INTEGRATION-TXN-001"),
		Description:       "Tenant A transaction",
		Amount:            decimal.NewFromFloat(100.00),
		CurrencyCode:      "USD",
		IsRecurring:       false,
		CreatedBy:         uuid.New(),
	}

	err = s.transactionRepo.Create(ctxA, transactionA)
	s.Require().NoError(err)
	s.createdTransactionIDs = append(s.createdTransactionIDs, transactionA.ID)

	transactionB := &domain.Transaction{
		ID:                uuid.New(),
		TransactionDate:   time.Now(),
		TransactionType:   domain.TransactionTypeManual,
		TransactionStatus: domain.TransactionStatusPending,
		ReferenceNumber:   stringPtr("INTEGRATION-TXN-001"), // Same reference, different tenant
		Description:       "Tenant B transaction",
		Amount:            decimal.NewFromFloat(200.00),
		CurrencyCode:      "USD",
		IsRecurring:       false,
		CreatedBy:         uuid.New(),
	}

	err = s.transactionRepo.Create(ctxB, transactionB)
	s.Require().NoError(err)
	s.createdTransactionIDs = append(s.createdTransactionIDs, transactionB.ID)

	// Verify tenant A can only see its data
	retrievedAccountA, err := s.accountRepo.GetByCode(ctxA, "INTEGRATION-CASH")
	s.Assert().NoError(err)
	s.Assert().Equal("Tenant A Cash Account", retrievedAccountA.AccountName)

	retrievedTransactionA, err := s.transactionRepo.GetByReference(ctxA, "INTEGRATION-TXN-001")
	s.Assert().NoError(err)
	s.Assert().Equal("Tenant A transaction", retrievedTransactionA.Description)

	// Verify tenant A cannot see tenant B's data
	_, err = s.accountRepo.GetByID(ctxA, accountB.ID)
	s.Assert().Error(err)
	s.Assert().Equal(domain.ErrAccountNotFound, err)

	_, err = s.transactionRepo.GetByID(ctxA, transactionB.ID)
	s.Assert().Error(err)
	s.Assert().Equal(domain.ErrTransactionNotFound, err)

	// Verify tenant B can only see its data
	retrievedAccountB, err := s.accountRepo.GetByCode(ctxB, "INTEGRATION-CASH")
	s.Assert().NoError(err)
	s.Assert().Equal("Tenant B Cash Account", retrievedAccountB.AccountName)

	retrievedTransactionB, err := s.transactionRepo.GetByReference(ctxB, "INTEGRATION-TXN-001")
	s.Assert().NoError(err)
	s.Assert().Equal("Tenant B transaction", retrievedTransactionB.Description)

	// Verify tenant B cannot see tenant A's data
	_, err = s.accountRepo.GetByID(ctxB, accountA.ID)
	s.Assert().Error(err)
	s.Assert().Equal(domain.ErrAccountNotFound, err)

	_, err = s.transactionRepo.GetByID(ctxB, transactionA.ID)
	s.Assert().Error(err)
	s.Assert().Equal(domain.ErrTransactionNotFound, err)
}

// TestAccountTransactionRelationship tests the relationship between accounts and transactions
func (s *FinanceRepositoryIntegrationTestSuite) TestAccountTransactionRelationship() {
	ctx := shared.WithTenantID(s.ctx, s.tenantA.ID)

	// Create test accounts
	cashAccount := &domain.Accounts{
		ID:                 uuid.New(),
		AccountCode:        "1000-CASH",
		AccountName:        "Cash Account",
		AccountDescription: stringPtr("Main cash account"),
		RootType:           domain.RootTypeAsset,
		AccountType:        "CASH",
		NormalBalance:      domain.NormalBalanceDebit,
		CurrencyCode:       "USD",
		IsActive:           true,
		AllowManualEntries: true,
	}

	err := s.accountRepo.Create(ctx, cashAccount)
	s.Require().NoError(err)
	s.createdAccountIDs = append(s.createdAccountIDs, cashAccount.ID)

	revenueAccount := &domain.Accounts{
		ID:                 uuid.New(),
		AccountCode:        "4000-REVENUE",
		AccountName:        "Sales Revenue",
		AccountDescription: stringPtr("Sales revenue account"),
		RootType:           domain.RootTypeRevenue,
		AccountType:        "SALES",
		NormalBalance:      domain.NormalBalanceCredit,
		CurrencyCode:       "USD",
		IsActive:           true,
		AllowManualEntries: true,
	}

	err = s.accountRepo.Create(ctx, revenueAccount)
	s.Require().NoError(err)
	s.createdAccountIDs = append(s.createdAccountIDs, revenueAccount.ID)

	// Create a transaction that would typically involve these accounts
	transaction := &domain.Transaction{
		ID:                uuid.New(),
		TransactionDate:   time.Now(),
		TransactionType:   domain.TransactionTypeManual,
		TransactionStatus: domain.TransactionStatusPending,
		ReferenceNumber:   stringPtr("SALE-001"),
		Description:       "Cash sale transaction",
		Amount:            decimal.NewFromFloat(500.00),
		CurrencyCode:      "USD",
		IsRecurring:       false,
		CreatedBy:         uuid.New(),
	}

	err = s.transactionRepo.Create(ctx, transaction)
	s.Require().NoError(err)
	s.createdTransactionIDs = append(s.createdTransactionIDs, transaction.ID)

	// Verify transaction was created and can be retrieved
	retrievedTransaction, err := s.transactionRepo.GetByID(ctx, transaction.ID)
	s.Assert().NoError(err)
	s.Assert().Equal(transaction.Description, retrievedTransaction.Description)
	s.Assert().True(transaction.Amount.Equal(retrievedTransaction.Amount))

	// Verify accounts are still accessible
	retrievedCashAccount, err := s.accountRepo.GetByID(ctx, cashAccount.ID)
	s.Assert().NoError(err)
	s.Assert().Equal(cashAccount.AccountName, retrievedCashAccount.AccountName)

	retrievedRevenueAccount, err := s.accountRepo.GetByID(ctx, revenueAccount.ID)
	s.Assert().NoError(err)
	s.Assert().Equal(revenueAccount.AccountName, retrievedRevenueAccount.AccountName)
}

// TestMultiTenantOperationsConsistency tests consistency of operations across tenants
func (s *FinanceRepositoryIntegrationTestSuite) TestMultiTenantOperationsConsistency() {
	// Setup test data in tenant A
	ctxA := shared.WithTenantID(s.ctx, s.tenantA.ID)

	// Create multiple accounts in tenant A
	accountsA := []*domain.Accounts{
		{
			ID:                 uuid.New(),
			AccountCode:        "1000-CASH-A",
			AccountName:        "Cash Account A",
			RootType:           domain.RootTypeAsset,
			AccountType:        "CASH",
			NormalBalance:      domain.NormalBalanceDebit,
			CurrencyCode:       "USD",
			IsActive:           true,
			AllowManualEntries: true,
		},
		{
			ID:                 uuid.New(),
			AccountCode:        "4000-REVENUE-A",
			AccountName:        "Revenue Account A",
			RootType:           domain.RootTypeRevenue,
			AccountType:        "SALES",
			NormalBalance:      domain.NormalBalanceCredit,
			CurrencyCode:       "USD",
			IsActive:           true,
			AllowManualEntries: true,
		},
	}

	for _, account := range accountsA {
		err := s.accountRepo.Create(ctxA, account)
		s.Require().NoError(err)
		s.createdAccountIDs = append(s.createdAccountIDs, account.ID)
	}

	// Create multiple transactions in tenant A
	transactionsA := []*domain.Transaction{
		{
			ID:                uuid.New(),
			TransactionDate:   time.Now(),
			TransactionType:   domain.TransactionTypeManual,
			TransactionStatus: domain.TransactionStatusPending,
			ReferenceNumber:   stringPtr("TXN-A-001"),
			Description:       "Transaction A1",
			Amount:            decimal.NewFromFloat(100.00),
			CurrencyCode:      "USD",
			IsRecurring:       false,
			CreatedBy:         uuid.New(),
		},
		{
			ID:                uuid.New(),
			TransactionDate:   time.Now(),
			TransactionType:   domain.TransactionTypeSystem,
			TransactionStatus: domain.TransactionStatusPosted,
			ReferenceNumber:   stringPtr("TXN-A-002"),
			Description:       "Transaction A2",
			Amount:            decimal.NewFromFloat(200.00),
			CurrencyCode:      "USD",
			IsRecurring:       false,
			CreatedBy:         uuid.New(),
		},
	}

	for _, transaction := range transactionsA {
		err := s.transactionRepo.Create(ctxA, transaction)
		s.Require().NoError(err)
		s.createdTransactionIDs = append(s.createdTransactionIDs, transaction.ID)
	}

	// Setup similar data in tenant B
	ctxB := shared.WithTenantID(s.ctx, s.tenantB.ID)

	accountsB := []*domain.Accounts{
		{
			ID:                 uuid.New(),
			AccountCode:        "1000-CASH-B",
			AccountName:        "Cash Account B",
			RootType:           domain.RootTypeAsset,
			AccountType:        "CASH",
			NormalBalance:      domain.NormalBalanceDebit,
			CurrencyCode:       "EUR",
			IsActive:           true,
			AllowManualEntries: true,
		},
		{
			ID:                 uuid.New(),
			AccountCode:        "4000-REVENUE-B",
			AccountName:        "Revenue Account B",
			RootType:           domain.RootTypeRevenue,
			AccountType:        "SALES",
			NormalBalance:      domain.NormalBalanceCredit,
			CurrencyCode:       "EUR",
			IsActive:           true,
			AllowManualEntries: true,
		},
	}

	for _, account := range accountsB {
		err := s.accountRepo.Create(ctxB, account)
		s.Require().NoError(err)
		s.createdAccountIDs = append(s.createdAccountIDs, account.ID)
	}

	transactionsB := []*domain.Transaction{
		{
			ID:                uuid.New(),
			TransactionDate:   time.Now(),
			TransactionType:   domain.TransactionTypeManual,
			TransactionStatus: domain.TransactionStatusPending,
			ReferenceNumber:   stringPtr("TXN-B-001"),
			Description:       "Transaction B1",
			Amount:            decimal.NewFromFloat(300.00),
			CurrencyCode:      "EUR",
			IsRecurring:       false,
			CreatedBy:         uuid.New(),
		},
		{
			ID:                uuid.New(),
			TransactionDate:   time.Now(),
			TransactionType:   domain.TransactionTypeSystem,
			TransactionStatus: domain.TransactionStatusPosted,
			ReferenceNumber:   stringPtr("TXN-B-002"),
			Description:       "Transaction B2",
			Amount:            decimal.NewFromFloat(400.00),
			CurrencyCode:      "EUR",
			IsRecurring:       false,
			CreatedBy:         uuid.New(),
		},
	}

	for _, transaction := range transactionsB {
		err := s.transactionRepo.Create(ctxB, transaction)
		s.Require().NoError(err)
		s.createdTransactionIDs = append(s.createdTransactionIDs, transaction.ID)
	}

	// Test listing operations in both tenants
	accountFilterA := &domain.AccountFilter{
		IsActive: boolPtr(true),
		Limit:    intPtr(10),
		Offset:   intPtr(0),
	}

	accountsResultA, err := s.accountRepo.List(ctxA, accountFilterA)
	s.Assert().NoError(err)
	s.Assert().GreaterOrEqual(len(accountsResultA), 2)

	accountsResultB, err := s.accountRepo.List(ctxB, accountFilterA)
	s.Assert().NoError(err)
	s.Assert().GreaterOrEqual(len(accountsResultB), 2)

	// Verify no cross-tenant data leakage in list operations
	for _, account := range accountsResultA {
		s.Assert().Equal("USD", account.CurrencyCode, "Tenant A should only see USD accounts")
	}

	for _, account := range accountsResultB {
		s.Assert().Equal("EUR", account.CurrencyCode, "Tenant B should only see EUR accounts")
	}

	// Test transaction listing
	transactionFilterA := &domain.TransactionFilter{
		Limit:  intPtr(10),
		Offset: intPtr(0),
	}

	transactionsResultA, err := s.transactionRepo.List(ctxA, transactionFilterA)
	s.Assert().NoError(err)
	s.Assert().GreaterOrEqual(len(transactionsResultA), 2)

	transactionsResultB, err := s.transactionRepo.List(ctxB, transactionFilterA)
	s.Assert().NoError(err)
	s.Assert().GreaterOrEqual(len(transactionsResultB), 2)

	// Verify no cross-tenant data leakage in transaction list operations
	for _, transaction := range transactionsResultA {
		s.Assert().Equal("USD", transaction.CurrencyCode, "Tenant A should only see USD transactions")
	}

	for _, transaction := range transactionsResultB {
		s.Assert().Equal("EUR", transaction.CurrencyCode, "Tenant B should only see EUR transactions")
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func intPtr(i int) *int {
	return &i
}
