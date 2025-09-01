//go:build database
// +build database

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/core/finance/domain"
	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/shared"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// TransactionRepositoryTestSuite defines comprehensive test suite for transaction repository operations
// Tests cover multi-tenant isolation, CRUD operations, transaction management, and error handling
type TransactionRepositoryTestSuite struct {
	suite.Suite
	ctx     context.Context
	runner  *tenant.DatabaseTestRunner
	repo    domain.TransactionRepository
	tenantA *db.Tenant
	tenantB *db.Tenant

	// Test data cleanup tracking
	createdTransactionIDs []uuid.UUID
}

// SetupSuite runs once before the entire test suite
func (s *TransactionRepositoryTestSuite) SetupSuite() {
	var err error
	s.runner, err = tenant.NewDatabaseTestRunner()
	s.Require().NoError(err, "Failed to connect to the database")
	s.ctx = context.Background()

	// Create test tenants
	s.tenantA, err = s.runner.CreateTestTenant(s.ctx, "finance-test-tenant-a")
	s.Require().NoError(err, "Failed to create tenant A")

	s.tenantB, err = s.runner.CreateTestTenant(s.ctx, "finance-test-tenant-b")
	s.Require().NoError(err, "Failed to create tenant B")

	// Setup repository
	traceService := tracing.NewNoOpTracingService()
	s.repo = NewTransactionRepository(s.runner.GetStore(), traceService)
}

// SetupTest runs before each test
func (s *TransactionRepositoryTestSuite) SetupTest() {
	s.createdTransactionIDs = make([]uuid.UUID, 0)
}

// TearDownTest runs after each test
func (s *TransactionRepositoryTestSuite) TearDownTest() {
	// Clean up created transactions
	for _, transactionID := range s.createdTransactionIDs {
		// Clean up in both tenant contexts to ensure cleanup
		ctxA := shared.WithTenantID(s.ctx, s.tenantA.ID)
		ctxB := shared.WithTenantID(s.ctx, s.tenantB.ID)

		_ = s.repo.Delete(ctxA, transactionID)
		_ = s.repo.Delete(ctxB, transactionID)
	}
}

// TearDownSuite runs once after the entire test suite
func (s *TransactionRepositoryTestSuite) TearDownSuite() {
	if s.runner != nil {
		s.runner.Close()
	}
}

// Test runner
func TestTransactionRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(TransactionRepositoryTestSuite))
}

// TestCreateTransaction tests transaction creation with various scenarios
func (s *TransactionRepositoryTestSuite) TestCreateTransaction() {
	testCases := []struct {
		name        string
		transaction *domain.Transaction
		expectError bool
		errorType   string
	}{
		{
			name: "Valid Manual Transaction Creation",
			transaction: &domain.Transaction{
				ID:                uuid.New(),
				TransactionDate:   time.Now(),
				TransactionType:   domain.TransactionTypeManual,
				TransactionStatus: domain.TransactionStatusPending,
				ReferenceNumber:   stringPtr("TXN-001"),
				Description:       "Test manual transaction",
				Amount:            decimal.NewFromFloat(100.50),
				CurrencyCode:      "USD",
				IsRecurring:       false,
				CreatedBy:         uuid.New(),
			},
			expectError: false,
		},
		{
			name: "Valid System Transaction Creation",
			transaction: &domain.Transaction{
				ID:                uuid.New(),
				TransactionDate:   time.Now(),
				TransactionType:   domain.TransactionTypeSystem,
				TransactionStatus: domain.TransactionStatusPosted,
				ReferenceNumber:   stringPtr("SYS-TXN-001"),
				Description:       "Test system transaction",
				Amount:            decimal.NewFromFloat(250.75),
				CurrencyCode:      "USD",
				IsRecurring:       false,
				CreatedBy:         uuid.New(),
			},
			expectError: false,
		},
		{
			name: "Transaction with Recurring Pattern",
			transaction: &domain.Transaction{
				ID:                 uuid.New(),
				TransactionDate:    time.Now(),
				TransactionType:    domain.TransactionTypeManual,
				TransactionStatus:  domain.TransactionStatusPending,
				ReferenceNumber:    stringPtr("REC-TXN-001"),
				Description:        "Test recurring transaction",
				Amount:             decimal.NewFromFloat(500.00),
				CurrencyCode:       "USD",
				IsRecurring:        true,
				RecurringFrequency: &domain.RecurringFrequencyMonthly,
				RecurringEndDate:   timePtr(time.Now().AddDate(1, 0, 0)),
				CreatedBy:          uuid.New(),
			},
			expectError: false,
		},
		{
			name: "Invalid Transaction - Missing Required Fields",
			transaction: &domain.Transaction{
				ID:                uuid.New(),
				TransactionType:   domain.TransactionTypeManual,
				TransactionStatus: domain.TransactionStatusPending,
				Description:       "",                      // Missing required field
				Amount:            decimal.NewFromFloat(0), // Invalid amount
			},
			expectError: true,
			errorType:   "validation",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Test with tenant A context
			ctx := shared.WithTenantID(s.ctx, s.tenantA.ID)

			err := s.repo.Create(ctx, tc.transaction)

			if tc.expectError {
				s.Assert().Error(err, "Expected error for test case: %s", tc.name)
			} else {
				s.Assert().NoError(err, "Expected no error for test case: %s", tc.name)
				if err == nil {
					s.createdTransactionIDs = append(s.createdTransactionIDs, tc.transaction.ID)

					// Verify the transaction was created correctly
					retrieved, err := s.repo.GetByID(ctx, tc.transaction.ID)
					s.Assert().NoError(err)
					s.Assert().Equal(tc.transaction.Description, retrieved.Description)
					s.Assert().Equal(tc.transaction.TransactionType, retrieved.TransactionType)
					s.Assert().Equal(tc.transaction.TransactionStatus, retrieved.TransactionStatus)
					s.Assert().True(tc.transaction.Amount.Equal(retrieved.Amount))
				}
			}
		})
	}
}

// TestGetTransactionByID tests transaction retrieval by ID
func (s *TransactionRepositoryTestSuite) TestGetTransactionByID() {
	ctx := shared.WithTenantID(s.ctx, s.tenantA.ID)

	// Create test transaction
	transaction := &domain.Transaction{
		ID:                uuid.New(),
		TransactionDate:   time.Now(),
		TransactionType:   domain.TransactionTypeManual,
		TransactionStatus: domain.TransactionStatusPending,
		ReferenceNumber:   stringPtr("GET-TEST-001"),
		Description:       "Test transaction for retrieval test",
		Amount:            decimal.NewFromFloat(150.25),
		CurrencyCode:      "USD",
		IsRecurring:       false,
		CreatedBy:         uuid.New(),
	}

	err := s.repo.Create(ctx, transaction)
	s.Require().NoError(err)
	s.createdTransactionIDs = append(s.createdTransactionIDs, transaction.ID)

	// Test retrieval
	retrieved, err := s.repo.GetByID(ctx, transaction.ID)
	s.Assert().NoError(err)
	s.Assert().Equal(transaction.Description, retrieved.Description)
	s.Assert().Equal(transaction.TransactionType, retrieved.TransactionType)
	s.Assert().True(transaction.Amount.Equal(retrieved.Amount))
	s.Assert().NotZero(retrieved.CreatedAt)

	// Test non-existent transaction
	nonExistentID := uuid.New()
	_, err = s.repo.GetByID(ctx, nonExistentID)
	s.Assert().Error(err)
	s.Assert().Equal(domain.ErrTransactionNotFound, err)
}

// TestTenantIsolation tests that tenant isolation is properly enforced
func (s *TransactionRepositoryTestSuite) TestTenantIsolation() {
	// Create transaction in tenant A
	ctxA := shared.WithTenantID(s.ctx, s.tenantA.ID)
	transactionA := &domain.Transaction{
		ID:                uuid.New(),
		TransactionDate:   time.Now(),
		TransactionType:   domain.TransactionTypeManual,
		TransactionStatus: domain.TransactionStatusPending,
		ReferenceNumber:   stringPtr("TENANT-A-001"),
		Description:       "Tenant A transaction",
		Amount:            decimal.NewFromFloat(100.00),
		CurrencyCode:      "USD",
		IsRecurring:       false,
		CreatedBy:         uuid.New(),
	}

	err := s.repo.Create(ctxA, transactionA)
	s.Require().NoError(err)
	s.createdTransactionIDs = append(s.createdTransactionIDs, transactionA.ID)

	// Create transaction in tenant B with same reference (should be allowed due to tenant isolation)
	ctxB := shared.WithTenantID(s.ctx, s.tenantB.ID)
	transactionB := &domain.Transaction{
		ID:                uuid.New(),
		TransactionDate:   time.Now(),
		TransactionType:   domain.TransactionTypeManual,
		TransactionStatus: domain.TransactionStatusPending,
		ReferenceNumber:   stringPtr("TENANT-A-001"), // Same reference but different tenant
		Description:       "Tenant B transaction",
		Amount:            decimal.NewFromFloat(200.00),
		CurrencyCode:      "USD",
		IsRecurring:       false,
		CreatedBy:         uuid.New(),
	}

	err = s.repo.Create(ctxB, transactionB)
	s.Require().NoError(err)
	s.createdTransactionIDs = append(s.createdTransactionIDs, transactionB.ID)

	// Verify tenant A can only see its transaction
	retrievedA, err := s.repo.GetByID(ctxA, transactionA.ID)
	s.Assert().NoError(err)
	s.Assert().Equal(transactionA.Description, retrievedA.Description)

	// Verify tenant A cannot see tenant B's transaction
	_, err = s.repo.GetByID(ctxA, transactionB.ID)
	s.Assert().Error(err)
	s.Assert().Equal(domain.ErrTransactionNotFound, err)

	// Verify tenant B can only see its transaction
	retrievedB, err := s.repo.GetByID(ctxB, transactionB.ID)
	s.Assert().NoError(err)
	s.Assert().Equal(transactionB.Description, retrievedB.Description)

	// Verify tenant B cannot see tenant A's transaction
	_, err = s.repo.GetByID(ctxB, transactionA.ID)
	s.Assert().Error(err)
	s.Assert().Equal(domain.ErrTransactionNotFound, err)
}

// TestListTransactionsWithFiltering tests transaction listing with various filters
func (s *TransactionRepositoryTestSuite) TestListTransactionsWithFiltering() {
	ctx := shared.WithTenantID(s.ctx, s.tenantA.ID)

	// Create multiple test transactions with different statuses
	transactions := []*domain.Transaction{
		{
			ID:                uuid.New(),
			TransactionDate:   time.Now(),
			TransactionType:   domain.TransactionTypeManual,
			TransactionStatus: domain.TransactionStatusPending,
			ReferenceNumber:   stringPtr("FILTER-001"),
			Description:       "Pending transaction 1",
			Amount:            decimal.NewFromFloat(100.00),
			CurrencyCode:      "USD",
			IsRecurring:       false,
			CreatedBy:         uuid.New(),
		},
		{
			ID:                uuid.New(),
			TransactionDate:   time.Now(),
			TransactionType:   domain.TransactionTypeManual,
			TransactionStatus: domain.TransactionStatusPosted,
			ReferenceNumber:   stringPtr("FILTER-002"),
			Description:       "Posted transaction 1",
			Amount:            decimal.NewFromFloat(200.00),
			CurrencyCode:      "USD",
			IsRecurring:       false,
			CreatedBy:         uuid.New(),
		},
		{
			ID:                uuid.New(),
			TransactionDate:   time.Now(),
			TransactionType:   domain.TransactionTypeSystem,
			TransactionStatus: domain.TransactionStatusPosted,
			ReferenceNumber:   stringPtr("FILTER-003"),
			Description:       "System transaction",
			Amount:            decimal.NewFromFloat(300.00),
			CurrencyCode:      "USD",
			IsRecurring:       false,
			CreatedBy:         uuid.New(),
		},
	}

	// Create all transactions
	for _, transaction := range transactions {
		err := s.repo.Create(ctx, transaction)
		s.Require().NoError(err)
		s.createdTransactionIDs = append(s.createdTransactionIDs, transaction.ID)
	}

	// Test list all transactions with limit
	filter := &domain.TransactionFilter{
		Limit:  intPtr(10),
		Offset: intPtr(0),
	}

	results, err := s.repo.List(ctx, filter)
	s.Assert().NoError(err)
	s.Assert().GreaterOrEqual(len(results), 3) // At least our 3 transactions

	// Test list by status
	statusFilter := &domain.TransactionFilter{
		Status: &domain.TransactionStatusPosted,
		Limit:  intPtr(10),
		Offset: intPtr(0),
	}

	statusResults, err := s.repo.List(ctx, statusFilter)
	s.Assert().NoError(err)
	s.Assert().GreaterOrEqual(len(statusResults), 2) // At least our 2 posted transactions

	// Verify all returned transactions have posted status
	for _, transaction := range statusResults {
		s.Assert().Equal(domain.TransactionStatusPosted, transaction.TransactionStatus)
	}

	// Test list by transaction type
	typeFilter := &domain.TransactionFilter{
		TransactionType: &domain.TransactionTypeSystem,
		Limit:           intPtr(10),
		Offset:          intPtr(0),
	}

	typeResults, err := s.repo.List(ctx, typeFilter)
	s.Assert().NoError(err)
	s.Assert().GreaterOrEqual(len(typeResults), 1) // At least our 1 system transaction

	// Verify all returned transactions are system type
	for _, transaction := range typeResults {
		s.Assert().Equal(domain.TransactionTypeSystem, transaction.TransactionType)
	}
}

// TestUpdateTransaction tests transaction updates
func (s *TransactionRepositoryTestSuite) TestUpdateTransaction() {
	ctx := shared.WithTenantID(s.ctx, s.tenantA.ID)

	// Create test transaction
	transaction := &domain.Transaction{
		ID:                uuid.New(),
		TransactionDate:   time.Now(),
		TransactionType:   domain.TransactionTypeManual,
		TransactionStatus: domain.TransactionStatusPending,
		ReferenceNumber:   stringPtr("UPDATE-001"),
		Description:       "Original transaction description",
		Amount:            decimal.NewFromFloat(100.00),
		CurrencyCode:      "USD",
		IsRecurring:       false,
		CreatedBy:         uuid.New(),
	}

	err := s.repo.Create(ctx, transaction)
	s.Require().NoError(err)
	s.createdTransactionIDs = append(s.createdTransactionIDs, transaction.ID)

	// Update transaction
	transaction.Description = "Updated transaction description"
	transaction.Amount = decimal.NewFromFloat(150.00)
	transaction.TransactionStatus = domain.TransactionStatusPosted

	err = s.repo.Update(ctx, transaction)
	s.Assert().NoError(err)

	// Verify updates
	updated, err := s.repo.GetByID(ctx, transaction.ID)
	s.Assert().NoError(err)
	s.Assert().Equal("Updated transaction description", updated.Description)
	s.Assert().True(decimal.NewFromFloat(150.00).Equal(updated.Amount))
	s.Assert().Equal(domain.TransactionStatusPosted, updated.TransactionStatus)
	s.Assert().NotEqual(updated.CreatedAt, updated.UpdatedAt)
}

// TestTransactionApproval tests transaction approval functionality
func (s *TransactionRepositoryTestSuite) TestTransactionApproval() {
	ctx := shared.WithTenantID(s.ctx, s.tenantA.ID)

	// Create test transaction
	transaction := &domain.Transaction{
		ID:                uuid.New(),
		TransactionDate:   time.Now(),
		TransactionType:   domain.TransactionTypeManual,
		TransactionStatus: domain.TransactionStatusPending,
		ReferenceNumber:   stringPtr("APPROVAL-001"),
		Description:       "Transaction requiring approval",
		Amount:            decimal.NewFromFloat(1000.00),
		CurrencyCode:      "USD",
		IsRecurring:       false,
		CreatedBy:         uuid.New(),
	}

	err := s.repo.Create(ctx, transaction)
	s.Require().NoError(err)
	s.createdTransactionIDs = append(s.createdTransactionIDs, transaction.ID)

	// Test approval
	approverID := uuid.New()
	approvedAt := time.Now()
	notes := "Approved for processing"

	err = s.repo.Approve(ctx, transaction.ID, approverID, approvedAt, &notes)
	s.Assert().NoError(err)

	// Verify approval
	approved, err := s.repo.GetByID(ctx, transaction.ID)
	s.Assert().NoError(err)
	s.Assert().Equal(domain.ApprovalStatusApproved, *approved.ApprovalStatus)
	s.Assert().Equal(approverID, *approved.ApprovedBy)
	s.Assert().WithinDuration(approvedAt, *approved.ApprovedAt, time.Second)
	s.Assert().Equal(notes, *approved.ApprovalNotes)
}

// TestDeleteTransaction tests soft delete functionality
func (s *TransactionRepositoryTestSuite) TestDeleteTransaction() {
	ctx := shared.WithTenantID(s.ctx, s.tenantA.ID)

	// Create test transaction
	transaction := &domain.Transaction{
		ID:                uuid.New(),
		TransactionDate:   time.Now(),
		TransactionType:   domain.TransactionTypeManual,
		TransactionStatus: domain.TransactionStatusPending,
		ReferenceNumber:   stringPtr("DELETE-001"),
		Description:       "Transaction to delete",
		Amount:            decimal.NewFromFloat(50.00),
		CurrencyCode:      "USD",
		IsRecurring:       false,
		CreatedBy:         uuid.New(),
	}

	err := s.repo.Create(ctx, transaction)
	s.Require().NoError(err)

	// Verify transaction exists
	_, err = s.repo.GetByID(ctx, transaction.ID)
	s.Assert().NoError(err)

	// Delete transaction
	err = s.repo.Delete(ctx, transaction.ID)
	s.Assert().NoError(err)

	// Verify transaction is soft deleted (should return not found error)
	_, err = s.repo.GetByID(ctx, transaction.ID)
	s.Assert().Error(err)
	s.Assert().Equal(domain.ErrTransactionNotFound, err)

	// Verify deleting non-existent transaction returns appropriate error
	nonExistentID := uuid.New()
	err = s.repo.Delete(ctx, nonExistentID)
	s.Assert().Error(err)
	s.Assert().Equal(domain.ErrTransactionNotFound, err)
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func timePtr(t time.Time) *time.Time {
	return &t
}
