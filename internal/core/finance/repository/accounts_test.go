//go:build database
// +build database

package repository

import (
	"context"
	"fmt"
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

// AccountsRepositoryTestSuite defines test suite for chart of accounts repository operations
// Tests cover multi-tenant isolation, CRUD operations, hierarchical operations, and error handling
type AccountsRepositoryTestSuite struct {
	suite.Suite
	ctx     context.Context
	runner  *tenant.DatabaseTestRunner
	repo    domain.AccountsRepository
	tenantA *db.Tenant
	tenantB *db.Tenant

	// Test data cleanup tracking
	createdAccountIDs []uuid.UUID
}

// SetupSuite runs once before the entire test suite
func (s *AccountsRepositoryTestSuite) SetupSuite() {
	var err error
	s.runner, err = tenant.NewDatabaseTestRunner()
	s.Require().NoError(err, "Failed to connect to the database")
	s.ctx = context.Background()

	// Create test tenants
	s.tenantA, err = s.runner.CreateTestTenant(s.ctx, fmt.Sprintf("accounts-tenant-a-%s", uuid.New().String()[:8]))
	s.Require().NoError(err, "Failed to create tenant A")

	s.tenantB, err = s.runner.CreateTestTenant(s.ctx, fmt.Sprintf("accounts-tenant-b-%s", uuid.New().String()[:8]))
	s.Require().NoError(err, "Failed to create tenant B")

	// Setup repository
	traceService := tracing.NewNoOpTracingService()
	s.repo = NewAccountsRepository(s.runner.GetStore(), nil, traceService) // nil cache for tests
}

// SetupTest runs before each test
func (s *AccountsRepositoryTestSuite) SetupTest() {
	s.createdAccountIDs = make([]uuid.UUID, 0)
}

// TearDownTest runs after each test
func (s *AccountsRepositoryTestSuite) TearDownTest() {
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
func (s *AccountsRepositoryTestSuite) TearDownSuite() {
	if s.runner != nil {
		s.runner.Close()
	}
}

// Test runner
func TestAccountsRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(AccountsRepositoryTestSuite))
}

// TestCreateAccount tests account creation with various scenarios
func (s *AccountsRepositoryTestSuite) TestCreateAccount() {
	testCases := []struct {
		name        string
		account     *domain.Accounts
		expectError bool
		errorType   string
	}{
		{
			name: "Valid Asset Account Creation",
			account: &domain.Accounts{
				ID:                      uuid.New(),
				AccountCode:             "1000-TEST-CASH",
				AccountName:             "Test Cash Account",
				AccountDescription:      stringPtr("Test cash account for unit tests"),
				RootType:                domain.RootTypeAsset,
				AccountType:             "CASH",
				AccountSubtype:          stringPtr("PETTY_CASH"),
				NormalBalance:           domain.NormalBalanceDebit,
				CurrencyCode:            stringPtr("USD"),
				IsActive:                true,
				AllowManualEntries:      true,
				CurrentBalance:          decimal.Zero,
				YTDBalance:              decimal.Zero,
				BudgetVarianceThreshold: decimal.Zero,
				ValidationStatus:        domain.ValidationStatusPending,
				Version:                 1,
				AccountLevel:            0, // Root level account
				HasChildren:             false,
				IsLeafAccount:           true,
				// CreatedBy:               uuid.New(), // Skip for now to avoid foreign key issues
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectError: false,
		},
		{
			name: "Valid Liability Account Creation",
			account: &domain.Accounts{
				ID:                      uuid.New(),
				AccountCode:             "2000-TEST-PAYABLE",
				AccountName:             "Test Accounts Payable",
				AccountDescription:      stringPtr("Test accounts payable for unit tests"),
				RootType:                domain.RootTypeLiability,
				AccountType:             "ACCOUNTS_PAYABLE",
				NormalBalance:           domain.NormalBalanceCredit,
				CurrencyCode:            stringPtr("USD"),
				IsActive:                true,
				AllowManualEntries:      true,
				CurrentBalance:          decimal.Zero,
				YTDBalance:              decimal.Zero,
				BudgetVarianceThreshold: decimal.Zero,
				ValidationStatus:        domain.ValidationStatusPending,
				Version:                 1,
				AccountLevel:            0, // Root level account
				HasChildren:             false,
				IsLeafAccount:           true,
				// CreatedBy:               uuid.New(), // Skip for now to avoid foreign key issues
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectError: false,
		},
		{
			name: "Invalid Account - Missing Required Fields",
			account: &domain.Accounts{
				ID:          uuid.New(),
				AccountCode: "", // Missing required field
				AccountName: "Test Invalid Account",
				RootType:    domain.RootTypeAsset,
				Version:     1,
			},
			expectError: true,
			errorType:   "validation",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Test with tenant A context
			ctx := shared.WithTenantID(s.ctx, s.tenantA.ID)
			s.T().Logf("Creating account with ID: %s, Code: %s, TenantID in context: %s", tc.account.ID, tc.account.AccountCode, s.tenantA.ID)

			err := s.repo.Create(ctx, tc.account)
			s.T().Logf("After creation - ID: %s, TenantID: %s, Error: %v", tc.account.ID, tc.account.TenantID, err)

			if tc.expectError {
				s.Assert().Error(err, "Expected error for test case: %s", tc.name)
			} else {
				s.Assert().NoError(err, "Expected no error for test case: %s", tc.name)
				if err == nil {
					s.createdAccountIDs = append(s.createdAccountIDs, tc.account.ID)

					// Verify the account was created correctly
					// Try both GetByID and GetByCode for debugging
					retrieved, err := s.repo.GetByID(ctx, tc.account.ID)
					s.T().Logf("GetByID result: account=%v, error=%v", retrieved != nil, err)

					// Also try GetByCode
					retrievedByCode, errByCode := s.repo.GetByCode(ctx, nil, tc.account.AccountCode)
					s.T().Logf("GetByCode result: account=%v, error=%v", retrievedByCode != nil, errByCode)
					if err != nil {
						s.T().Logf("Failed to retrieve created account %s: %v", tc.account.ID, err)
						s.Assert().NoError(err)
					} else {
						s.Assert().Equal(tc.account.AccountCode, retrieved.AccountCode)
						s.Assert().Equal(tc.account.AccountName, retrieved.AccountName)
						s.Assert().Equal(tc.account.RootType, retrieved.RootType)
						s.Assert().Equal(tc.account.NormalBalance, retrieved.NormalBalance)
					}
				}
			}
		})
	}
}

// TestGetAccountByID tests account retrieval by ID
func (s *AccountsRepositoryTestSuite) TestGetAccountByID() {
	ctx := shared.WithTenantID(s.ctx, s.tenantA.ID)

	// Create test account
	account := &domain.Accounts{
		ID:                 uuid.New(),
		AccountCode:        "1100-TEST-CHECKING",
		AccountName:        "Test Checking Account",
		AccountDescription: stringPtr("Test checking account for retrieval test"),
		RootType:           domain.RootTypeAsset,
		AccountType:        "BANK",
		NormalBalance:      domain.NormalBalanceDebit,
		CurrencyCode:       stringPtr("USD"),
		IsActive:           true,
		AllowManualEntries: true,
		Version:            1,
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
func (s *AccountsRepositoryTestSuite) TestGetAccountByCode() {
	ctx := shared.WithTenantID(s.ctx, s.tenantA.ID)

	// Create test account
	account := &domain.Accounts{
		ID:                 uuid.New(),
		AccountCode:        "1200-TEST-SAVINGS",
		AccountName:        "Test Savings Account",
		AccountDescription: stringPtr("Test savings account for code retrieval test"),
		RootType:           domain.RootTypeAsset,
		AccountType:        "BANK",
		NormalBalance:      domain.NormalBalanceDebit,
		CurrencyCode:       stringPtr("USD"),
		IsActive:           true,
		AllowManualEntries: true,
		Version:            1,
	}

	err := s.repo.Create(ctx, account)
	s.Require().NoError(err)
	s.createdAccountIDs = append(s.createdAccountIDs, account.ID)

	// Test retrieval by code
	retrieved, err := s.repo.GetByCode(ctx, nil, account.AccountCode)
	s.Assert().NoError(err)
	s.Assert().Equal(account.ID, retrieved.ID)
	s.Assert().Equal(account.AccountCode, retrieved.AccountCode)

	// Test non-existent code
	_, err = s.repo.GetByCode(ctx, nil, "NON-EXISTENT-CODE")
	s.Assert().Error(err)
	s.Assert().Equal(domain.ErrAccountNotFound, err)
}

// TestTenantIsolation tests that tenant isolation is properly enforced
func (s *AccountsRepositoryTestSuite) TestTenantIsolation() {
	// Create account in tenant A
	ctxA := shared.WithTenantID(s.ctx, s.tenantA.ID)
	accountA := &domain.Accounts{
		ID:                 uuid.New(),
		AccountCode:        "1300-TENANT-A",
		AccountName:        "Tenant A Account",
		AccountDescription: stringPtr("Account that should only be visible to tenant A"),
		RootType:           domain.RootTypeAsset,
		AccountType:        "CASH",
		NormalBalance:      domain.NormalBalanceDebit,
		CurrencyCode:       stringPtr("USD"),
		IsActive:           true,
		AllowManualEntries: true,
		Version:            1,
	}

	err := s.repo.Create(ctxA, accountA)
	s.Require().NoError(err)
	s.createdAccountIDs = append(s.createdAccountIDs, accountA.ID)

	// Create account in tenant B with same code (should be allowed due to tenant isolation)
	ctxB := shared.WithTenantID(s.ctx, s.tenantB.ID)
	accountB := &domain.Accounts{
		ID:                 uuid.New(),
		AccountCode:        "1300-TENANT-A", // Same code but different tenant
		AccountName:        "Tenant B Account",
		AccountDescription: stringPtr("Account that should only be visible to tenant B"),
		RootType:           domain.RootTypeAsset,
		AccountType:        "CASH",
		NormalBalance:      domain.NormalBalanceDebit,
		CurrencyCode:       stringPtr("USD"),
		IsActive:           true,
		AllowManualEntries: true,
		Version:            1,
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
func (s *AccountsRepositoryTestSuite) TestListAccountsWithFiltering() {
	ctx := shared.WithTenantID(s.ctx, s.tenantA.ID)

	// Create multiple test accounts
	accounts := []*domain.Accounts{
		{
			ID:                 uuid.New(),
			AccountCode:        "1400-TEST-CASH1",
			AccountName:        "Test Cash Account 1",
			AccountDescription: stringPtr("First cash account"),
			RootType:           domain.RootTypeAsset,
			AccountType:        "CASH",
			NormalBalance:      domain.NormalBalanceDebit,
			CurrencyCode:       stringPtr("USD"),
			IsActive:           true,
			AllowManualEntries: true,
			Version:            1,
		},
		{
			ID:                 uuid.New(),
			AccountCode:        "1401-TEST-CASH2",
			AccountName:        "Test Cash Account 2",
			AccountDescription: stringPtr("Second cash account"),
			RootType:           domain.RootTypeAsset,
			AccountType:        "CASH",
			NormalBalance:      domain.NormalBalanceDebit,
			CurrencyCode:       stringPtr("USD"),
			IsActive:           false, // Inactive account
			AllowManualEntries: true,
			Version:            1,
		},
		{
			ID:                 uuid.New(),
			AccountCode:        "2400-TEST-PAYABLE",
			AccountName:        "Test Accounts Payable",
			AccountDescription: stringPtr("Payable account"),
			RootType:           domain.RootTypeLiability,
			AccountType:        "ACCOUNTS_PAYABLE",
			NormalBalance:      domain.NormalBalanceCredit,
			CurrencyCode:       stringPtr("USD"),
			IsActive:           true,
			AllowManualEntries: true,
			Version:            1,
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
	assetRootType := domain.RootTypeAsset
	assetFilter := &domain.AccountFilter{
		RootType: &assetRootType,
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
func (s *AccountsRepositoryTestSuite) TestUpdateAccount() {
	ctx := shared.WithTenantID(s.ctx, s.tenantA.ID)

	// Create test account
	account := &domain.Accounts{
		ID:                 uuid.New(),
		AccountCode:        "1500-TEST-UPDATE",
		AccountName:        "Original Account Name",
		AccountDescription: stringPtr("Original description"),
		RootType:           domain.RootTypeAsset,
		AccountType:        "CASH",
		NormalBalance:      domain.NormalBalanceDebit,
		CurrencyCode:       stringPtr("USD"),
		IsActive:           true,
		AllowManualEntries: true,
		Version:            1,
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
func (s *AccountsRepositoryTestSuite) TestDeleteAccount() {
	ctx := shared.WithTenantID(s.ctx, s.tenantA.ID)

	// Create test account
	account := &domain.Accounts{
		ID:                 uuid.New(),
		AccountCode:        "1600-TEST-DELETE",
		AccountName:        "Account to Delete",
		AccountDescription: stringPtr("Account for delete testing"),
		RootType:           domain.RootTypeAsset,
		AccountType:        "CASH",
		NormalBalance:      domain.NormalBalanceDebit,
		CurrencyCode:       stringPtr("USD"),
		IsActive:           true,
		AllowManualEntries: true,
		Version:            1,
		ParentAccountID:    nil, // Explicitly set to nil to ensure no foreign key reference
	}

	err := s.repo.Create(ctx, account)
	s.Require().NoError(err)
	s.createdAccountIDs = append(s.createdAccountIDs, account.ID)

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

// Helper functions are now in mappers.go
