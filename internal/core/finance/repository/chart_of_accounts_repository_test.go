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

// ChartOfAccountsRepositoryTestSuite defines comprehensive test suite for chart of accounts repository operations
// Tests cover multi-tenant isolation, CRUD operations, hierarchical operations, and error handling
type ChartOfAccountsRepositoryTestSuite struct {
	suite.Suite
	ctx     context.Context
	runner  *tenant.DatabaseTestRunner
	repo    domain.ChartOfAccountsRepository
	tenantA *db.Tenant
	tenantB *db.Tenant

	// Test data cleanup tracking
	createdAccountIDs []uuid.UUID
}

// SetupSuite runs once before the entire test suite
func (s *ChartOfAccountsRepositoryTestSuite) SetupSuite() {
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
	s.repo = NewChartOfAccountsRepository(s.runner.GetStore(), traceService)
}

// SetupTest runs before each test
func (s *ChartOfAccountsRepositoryTestSuite) SetupTest() {
	s.createdAccountIDs = make([]uuid.UUID, 0)
}

// TearDownTest runs after each test
func (s *ChartOfAccountsRepositoryTestSuite) TearDownTest() {
	// Clean up created accounts
	for _, accountID := range s.createdAccountIDs {
		// Clean up in both tenant contexts to ensure cleanup
		ctxA := shared.WithTenantID(s.ctx, s.tenantA.ID)
		ctxB := shared.WithTenantID(s.ctx, s.tenantB.ID)

		_ = s.repo.Delete(ctxA, accountID)
		_ = s.repo.Delete(ctxB, accountID)
	}
}

// TearDownSuite runs once after the entire test suite
func (s *ChartOfAccountsRepositoryTestSuite) TearDownSuite() {
	if s.runner != nil {
		s.runner.Close()
	}
}

// Test runner
func TestChartOfAccountsRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(ChartOfAccountsRepositoryTestSuite))
}

// TestCreateAccount tests account creation with various scenarios
func (s *ChartOfAccountsRepositoryTestSuite) TestCreateAccount() {
	testCases := []struct {
		name        string
		account     *domain.ChartOfAccounts
		expectError bool
		errorType   string
	}{
		{
			name: "Valid Asset Account Creation",
			account: &domain.ChartOfAccounts{
				ID:                 uuid.New(),
				AccountCode:        "1000-TEST-CASH",
				AccountName:        "Test Cash Account",
				AccountDescription: stringPtr("Test cash account for unit tests"),
				RootType:           domain.RootTypeAsset,
				AccountType:        "CASH",
				AccountSubtype:     stringPtr("PETTY_CASH"),
				NormalBalance:      domain.NormalBalanceDebit,
				CurrencyCode:       "USD",
				IsActive:           true,
				AllowManualEntries: true,
			},
			expectError: false,
		},
		{
			name: "Valid Liability Account Creation",
			account: &domain.ChartOfAccounts{
				ID:                 uuid.New(),
				AccountCode:        "2000-TEST-PAYABLE",
				AccountName:        "Test Accounts Payable",
				AccountDescription: stringPtr("Test accounts payable for unit tests"),
				RootType:           domain.RootTypeLiability,
				AccountType:        "ACCOUNTS_PAYABLE",
				NormalBalance:      domain.NormalBalanceCredit,
				CurrencyCode:       "USD",
				IsActive:           true,
				AllowManualEntries: true,
			},
			expectError: false,
		},
		{
			name: "Invalid Account - Missing Required Fields",
			account: &domain.ChartOfAccounts{
				ID:          uuid.New(),
				AccountCode: "", // Missing required field
				AccountName: "Test Invalid Account",
				RootType:    domain.RootTypeAsset,
			},
			expectError: true,
			errorType:   "validation",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Test with tenant A context
			ctx := shared.WithTenantID(s.ctx, s.tenantA.ID)

			err := s.repo.Create(ctx, tc.account)

			if tc.expectError {
				s.Assert().Error(err, "Expected error for test case: %s", tc.name)
			} else {
				s.Assert().NoError(err, "Expected no error for test case: %s", tc.name)
				if err == nil {
					s.createdAccountIDs = append(s.createdAccountIDs, tc.account.ID)

					// Verify the account was created correctly
					retrieved, err := s.repo.GetByID(ctx, tc.account.ID)
					s.Assert().NoError(err)
					s.Assert().Equal(tc.account.AccountCode, retrieved.AccountCode)
					s.Assert().Equal(tc.account.AccountName, retrieved.AccountName)
					s.Assert().Equal(tc.account.RootType, retrieved.RootType)
					s.Assert().Equal(tc.account.NormalBalance, retrieved.NormalBalance)
				}
			}
		})
	}
}

// TestGetAccountByID tests account retrieval by ID
func (s *ChartOfAccountsRepositoryTestSuite) TestGetAccountByID() {
	ctx := shared.WithTenantID(s.ctx, s.tenantA.ID)

	// Create test account
	account := &domain.ChartOfAccounts{
		ID:                 uuid.New(),
		AccountCode:        "1100-TEST-CHECKING",
		AccountName:        "Test Checking Account",
		AccountDescription: stringPtr("Test checking account for retrieval test"),
		RootType:           domain.RootTypeAsset,
		AccountType:        "BANK",
		NormalBalance:      domain.NormalBalanceDebit,
		CurrencyCode:       "USD",
		IsActive:           true,
		AllowManualEntries: true,
	}

	err := s.repo.Create(ctx, account)
	s.Require().NoError(err)
	s.createdAccountIDs = append(s.createdAccountIDs, account.ID)

	// Test retrieval
	retrieved, err := s.repo.GetByID(ctx, account.ID)
	s.Assert().NoError(err)
	s.Assert().Equal(account.AccountCode, retrieved.AccountCode)
	s.Assert().Equal(account.AccountName, retrieved.AccountName)
	s.Assert().NotZero(retrieved.CreatedAt)

	// Test non-existent account
	nonExistentID := uuid.New()
	_, err = s.repo.GetByID(ctx, nonExistentID)
	s.Assert().Error(err)
	s.Assert().Equal(domain.ErrAccountNotFound, err)
}

// TestGetAccountByCode tests account retrieval by code
func (s *ChartOfAccountsRepositoryTestSuite) TestGetAccountByCode() {
	ctx := shared.WithTenantID(s.ctx, s.tenantA.ID)

	// Create test account
	account := &domain.ChartOfAccounts{
		ID:                 uuid.New(),
		AccountCode:        "1200-TEST-SAVINGS",
		AccountName:        "Test Savings Account",
		AccountDescription: stringPtr("Test savings account for code retrieval test"),
		RootType:           domain.RootTypeAsset,
		AccountType:        "BANK",
		NormalBalance:      domain.NormalBalanceDebit,
		CurrencyCode:       "USD",
		IsActive:           true,
		AllowManualEntries: true,
	}

	err := s.repo.Create(ctx, account)
	s.Require().NoError(err)
	s.createdAccountIDs = append(s.createdAccountIDs, account.ID)

	// Test retrieval by code
	retrieved, err := s.repo.GetByCode(ctx, account.AccountCode)
	s.Assert().NoError(err)
	s.Assert().Equal(account.ID, retrieved.ID)
	s.Assert().Equal(account.AccountCode, retrieved.AccountCode)

	// Test non-existent code
	_, err = s.repo.GetByCode(ctx, "NON-EXISTENT-CODE")
	s.Assert().Error(err)
	s.Assert().Equal(domain.ErrAccountNotFound, err)
}

// TestTenantIsolation tests that tenant isolation is properly enforced
func (s *ChartOfAccountsRepositoryTestSuite) TestTenantIsolation() {
	// Create account in tenant A
	ctxA := shared.WithTenantID(s.ctx, s.tenantA.ID)
	accountA := &domain.ChartOfAccounts{
		ID:                 uuid.New(),
		AccountCode:        "1300-TENANT-A",
		AccountName:        "Tenant A Account",
		AccountDescription: stringPtr("Account that should only be visible to tenant A"),
		RootType:           domain.RootTypeAsset,
		AccountType:        "CASH",
		NormalBalance:      domain.NormalBalanceDebit,
		CurrencyCode:       "USD",
		IsActive:           true,
		AllowManualEntries: true,
	}

	err := s.repo.Create(ctxA, accountA)
	s.Require().NoError(err)
	s.createdAccountIDs = append(s.createdAccountIDs, accountA.ID)

	// Create account in tenant B with same code (should be allowed due to tenant isolation)
	ctxB := shared.WithTenantID(s.ctx, s.tenantB.ID)
	accountB := &domain.ChartOfAccounts{
		ID:                 uuid.New(),
		AccountCode:        "1300-TENANT-A", // Same code but different tenant
		AccountName:        "Tenant B Account",
		AccountDescription: stringPtr("Account that should only be visible to tenant B"),
		RootType:           domain.RootTypeAsset,
		AccountType:        "CASH",
		NormalBalance:      domain.NormalBalanceDebit,
		CurrencyCode:       "USD",
		IsActive:           true,
		AllowManualEntries: true,
	}

	err = s.repo.Create(ctxB, accountB)
	s.Require().NoError(err)
	s.createdAccountIDs = append(s.createdAccountIDs, accountB.ID)

	// Verify tenant A can only see its account
	retrievedA, err := s.repo.GetByID(ctxA, accountA.ID)
	s.Assert().NoError(err)
	s.Assert().Equal(accountA.AccountName, retrievedA.AccountName)

	// Verify tenant A cannot see tenant B's account
	_, err = s.repo.GetByID(ctxA, accountB.ID)
	s.Assert().Error(err)
	s.Assert().Equal(domain.ErrAccountNotFound, err)

	// Verify tenant B can only see its account
	retrievedB, err := s.repo.GetByID(ctxB, accountB.ID)
	s.Assert().NoError(err)
	s.Assert().Equal(accountB.AccountName, retrievedB.AccountName)

	// Verify tenant B cannot see tenant A's account
	_, err = s.repo.GetByID(ctxB, accountA.ID)
	s.Assert().Error(err)
	s.Assert().Equal(domain.ErrAccountNotFound, err)
}

// TestListAccountsWithFiltering tests account listing with various filters
func (s *ChartOfAccountsRepositoryTestSuite) TestListAccountsWithFiltering() {
	ctx := shared.WithTenantID(s.ctx, s.tenantA.ID)

	// Create multiple test accounts
	accounts := []*domain.ChartOfAccounts{
		{
			ID:                 uuid.New(),
			AccountCode:        "1400-TEST-CASH1",
			AccountName:        "Test Cash Account 1",
			AccountDescription: stringPtr("First cash account"),
			RootType:           domain.RootTypeAsset,
			AccountType:        "CASH",
			NormalBalance:      domain.NormalBalanceDebit,
			CurrencyCode:       "USD",
			IsActive:           true,
			AllowManualEntries: true,
		},
		{
			ID:                 uuid.New(),
			AccountCode:        "1401-TEST-CASH2",
			AccountName:        "Test Cash Account 2",
			AccountDescription: stringPtr("Second cash account"),
			RootType:           domain.RootTypeAsset,
			AccountType:        "CASH",
			NormalBalance:      domain.NormalBalanceDebit,
			CurrencyCode:       "USD",
			IsActive:           false, // Inactive account
			AllowManualEntries: true,
		},
		{
			ID:                 uuid.New(),
			AccountCode:        "2400-TEST-PAYABLE",
			AccountName:        "Test Accounts Payable",
			AccountDescription: stringPtr("Payable account"),
			RootType:           domain.RootTypeLiability,
			AccountType:        "ACCOUNTS_PAYABLE",
			NormalBalance:      domain.NormalBalanceCredit,
			CurrencyCode:       "USD",
			IsActive:           true,
			AllowManualEntries: true,
		},
	}

	// Create all accounts
	for _, account := range accounts {
		err := s.repo.Create(ctx, account)
		s.Require().NoError(err)
		s.createdAccountIDs = append(s.createdAccountIDs, account.ID)
	}

	// Test list all active accounts
	filter := &domain.AccountFilter{
		IsActive: boolPtr(true),
		Limit:    intPtr(10),
		Offset:   intPtr(0),
	}

	results, err := s.repo.List(ctx, filter)
	s.Assert().NoError(err)
	s.Assert().GreaterOrEqual(len(results), 2) // At least our 2 active accounts

	// Test list by root type
	assetFilter := &domain.AccountFilter{
		RootType: &domain.RootTypeAsset,
		IsActive: boolPtr(true),
		Limit:    intPtr(10),
		Offset:   intPtr(0),
	}

	assetResults, err := s.repo.List(ctx, assetFilter)
	s.Assert().NoError(err)
	s.Assert().GreaterOrEqual(len(assetResults), 1) // At least our active asset account

	// Verify all returned accounts are assets
	for _, account := range assetResults {
		s.Assert().Equal(domain.RootTypeAsset, account.RootType)
	}
}

// TestUpdateAccount tests account updates
func (s *ChartOfAccountsRepositoryTestSuite) TestUpdateAccount() {
	ctx := shared.WithTenantID(s.ctx, s.tenantA.ID)

	// Create test account
	account := &domain.ChartOfAccounts{
		ID:                 uuid.New(),
		AccountCode:        "1500-TEST-UPDATE",
		AccountName:        "Original Account Name",
		AccountDescription: stringPtr("Original description"),
		RootType:           domain.RootTypeAsset,
		AccountType:        "CASH",
		NormalBalance:      domain.NormalBalanceDebit,
		CurrencyCode:       "USD",
		IsActive:           true,
		AllowManualEntries: true,
	}

	err := s.repo.Create(ctx, account)
	s.Require().NoError(err)
	s.createdAccountIDs = append(s.createdAccountIDs, account.ID)

	// Update account
	account.AccountName = "Updated Account Name"
	account.AccountDescription = stringPtr("Updated description")
	account.IsActive = false

	err = s.repo.Update(ctx, account)
	s.Assert().NoError(err)

	// Verify updates
	updated, err := s.repo.GetByID(ctx, account.ID)
	s.Assert().NoError(err)
	s.Assert().Equal("Updated Account Name", updated.AccountName)
	s.Assert().Equal("Updated description", *updated.AccountDescription)
	s.Assert().False(updated.IsActive)
	s.Assert().NotEqual(updated.CreatedAt, updated.UpdatedAt)
}

// TestDeleteAccount tests soft delete functionality
func (s *ChartOfAccountsRepositoryTestSuite) TestDeleteAccount() {
	ctx := shared.WithTenantID(s.ctx, s.tenantA.ID)

	// Create test account
	account := &domain.ChartOfAccounts{
		ID:                 uuid.New(),
		AccountCode:        "1600-TEST-DELETE",
		AccountName:        "Account to Delete",
		AccountDescription: stringPtr("Account for delete testing"),
		RootType:           domain.RootTypeAsset,
		AccountType:        "CASH",
		NormalBalance:      domain.NormalBalanceDebit,
		CurrencyCode:       "USD",
		IsActive:           true,
		AllowManualEntries: true,
	}

	err := s.repo.Create(ctx, account)
	s.Require().NoError(err)

	// Verify account exists
	_, err = s.repo.GetByID(ctx, account.ID)
	s.Assert().NoError(err)

	// Delete account
	err = s.repo.Delete(ctx, account.ID)
	s.Assert().NoError(err)

	// Verify account is soft deleted (should return not found error)
	_, err = s.repo.GetByID(ctx, account.ID)
	s.Assert().Error(err)
	s.Assert().Equal(domain.ErrAccountNotFound, err)

	// Verify deleting non-existent account returns appropriate error
	nonExistentID := uuid.New()
	err = s.repo.Delete(ctx, nonExistentID)
	s.Assert().Error(err)
	s.Assert().Equal(domain.ErrAccountNotFound, err)
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
