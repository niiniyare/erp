//go:build unit
// +build unit

package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/core/finance/domain"
	"github.com/niiniyare/erp/internal/shared"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// AccountsRepositoryUnitTestSuite defines unit test suite using mocks
type AccountsRepositoryUnitTestSuite struct {
	suite.Suite
	ctrl       *gomock.Controller
	mockStore  *db.MockStore
	mockTracer *tracing.MockTracingService
	repo       *chartOfAccountsRepository
	ctx        context.Context
	tenantID   uuid.UUID
}

// SetupTest runs before each test
func (s *AccountsRepositoryUnitTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockStore = db.NewMockStore(s.ctrl)
	s.mockTracer = tracing.NewMockTracingService(s.ctrl)
	s.repo = NewAccountRepository(s.mockStore, s.mockTracer)
	s.tenantID = uuid.New()
	s.ctx = shared.WithTenantID(context.Background(), s.tenantID)
}

// TearDownTest runs after each test
func (s *AccountsRepositoryUnitTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

// Test runner
func TestAccountsRepositoryUnitTestSuite(t *testing.T) {
	suite.Run(t, new(AccountsRepositoryUnitTestSuite))
}

// TestCreateAccountSuccess tests successful account creation
func (s *AccountsRepositoryUnitTestSuite) TestCreateAccountSuccess() {
	account := &domain.Accounts{
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
		Version:            1,
	}

	// Expected SQLC parameters
	expectedParams := db.CreateAccountParams{
		EntityID:                    account.EntityID,
		AccountCode:                 account.AccountCode,
		AccountName:                 account.AccountName,
		AccountDescription:          getStringValue(account.AccountDescription),
		ParentAccountID:             account.ParentAccountID,
		RootType:                    db.RootTypeEnumASSET,
		AccountType:                 account.AccountType,
		AccountSubtype:              account.AccountSubtype,
		NormalBalance:               db.NormalBalanceEnumDEBIT,
		IsControlAccount:            account.IsControlAccount,
		ControlAccountID:            account.ControlAccountID,
		CurrencyCode:                account.CurrencyCode,
		IsMultiCurrency:             &account.IsMultiCurrency,
		CurrencyRevaluationRequired: &account.CurrencyRevaluationRequired,
		IsActive:                    account.IsActive,
		IsSystemAccount:             false,
		AllowManualEntries:          account.AllowManualEntries,
		RequireReference:            account.RequireReference,
		FinancialStatementLine:      account.FinancialStatementLine,
		ReportOrder:                 getInt32Ptr(account.ReportOrder),
		IsBudgetable:                getBoolPtr(account.IsBudgetable),
		BudgetVarianceThreshold:     decimalToPgNumeric(&account.BudgetVarianceThreshold),
		AccountAttributes:           []byte("{}"),
	}

	// Mock successful creation
	returnedAccount := db.FinanceChartOfAccount{
		ID:                          account.ID,
		TenantID:                    s.tenantID,
		EntityID:                    account.EntityID,
		AccountCode:                 account.AccountCode,
		AccountName:                 account.AccountName,
		AccountDescription:          getStringValue(account.AccountDescription),
		ParentAccountID:             account.ParentAccountID,
		AccountLevel:                1,
		AccountPath:                 account.AccountCode,
		RootType:                    db.RootTypeEnumASSET,
		AccountType:                 account.AccountType,
		AccountSubtype:              account.AccountSubtype,
		NormalBalance:               db.NormalBalanceEnumDEBIT,
		IsControlAccount:            account.IsControlAccount,
		ControlAccountID:            account.ControlAccountID,
		CurrencyCode:                account.CurrencyCode,
		IsMultiCurrency:             &account.IsMultiCurrency,
		CurrencyRevaluationRequired: &account.CurrencyRevaluationRequired,
		IsActive:                    account.IsActive,
		IsSystemAccount:             false,
		AllowManualEntries:          account.AllowManualEntries,
		RequireReference:            account.RequireReference,
		CreatedAt:                   time.Now(),
		UpdatedAt:                   time.Now(),
		CreatedBy:                   &uuid.UUID{},
	}

	s.mockStore.EXPECT().
		WithTenant(gomock.Any(), s.tenantID, gomock.Any()).
		DoAndReturn(func(ctx context.Context, tenantID uuid.UUID, fn func(context.Context, db.Store) error) error {
			// Mock the inner store call
			s.mockStore.EXPECT().
				CreateAccount(gomock.Any(), gomock.Eq(expectedParams)).
				Return(&returnedAccount, nil).
				Times(1)

			return fn(ctx, s.mockStore)
		}).
		Times(1)

	// Execute test
	err := s.repo.Create(s.ctx, account)

	// Assert results
	s.Assert().NoError(err)
	s.Assert().Equal(s.tenantID, account.TenantID)
	s.Assert().NotZero(account.CreatedAt)
	s.Assert().NotZero(account.UpdatedAt)
}

// TestCreateAccountDuplicateCode tests handling of duplicate account codes
func (s *AccountsRepositoryUnitTestSuite) TestCreateAccountDuplicateCode() {
	account := &domain.Accounts{
		ID:                 uuid.New(),
		AccountCode:        "1000-CASH",
		AccountName:        "Cash Account",
		RootType:           domain.RootTypeAsset,
		AccountType:        "CASH",
		NormalBalance:      domain.NormalBalanceDebit,
		CurrencyCode:       "USD",
		IsActive:           true,
		AllowManualEntries: true,
		Version:            1,
	}

	// Mock duplicate key violation
	s.mockStore.EXPECT().
		WithTenant(gomock.Any(), s.tenantID, gomock.Any()).
		DoAndReturn(func(ctx context.Context, tenantID uuid.UUID, fn func(context.Context, db.Store) error) error {
			s.mockStore.EXPECT().
				CreateAccount(gomock.Any(), gomock.Any()).
				Return(nil, &mockPgError{code: "23505"}). // Unique violation
				Times(1)

			return fn(ctx, s.mockStore)
		}).
		Times(1)

	// Execute test
	err := s.repo.Create(s.ctx, account)

	// Assert results
	s.Assert().Error(err)
	s.Assert().Equal(domain.ErrAccountCodeExists, err)
}

// TestGetByIDSuccess tests successful account retrieval by ID
func (s *AccountsRepositoryUnitTestSuite) TestGetByIDSuccess() {
	accountID := uuid.New()
	expectedAccount := db.FinanceChartOfAccount{
		ID:                 accountID,
		TenantID:           s.tenantID,
		AccountCode:        "1000-CASH",
		AccountName:        "Cash Account",
		AccountDescription: "Main cash account",
		RootType:           db.RootTypeEnumASSET,
		AccountType:        "CASH",
		NormalBalance:      db.NormalBalanceEnumDEBIT,
		CurrencyCode:       "USD",
		IsActive:           true,
		AllowManualEntries: true,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	s.mockStore.EXPECT().
		WithTenant(gomock.Any(), s.tenantID, gomock.Any()).
		DoAndReturn(func(ctx context.Context, tenantID uuid.UUID, fn func(context.Context, db.Store) error) error {
			s.mockStore.EXPECT().
				GetAccountByID(gomock.Any(), accountID).
				Return(&expectedAccount, nil).
				Times(1)

			return fn(ctx, s.mockStore)
		}).
		Times(1)

	// Execute test
	result, err := s.repo.GetByID(s.ctx, accountID)

	// Assert results
	s.Assert().NoError(err)
	s.Assert().NotNil(result)
	s.Assert().Equal(accountID, result.ID)
	s.Assert().Equal("1000-CASH", result.AccountCode)
	s.Assert().Equal("Cash Account", result.AccountName)
	s.Assert().Equal(domain.RootTypeAsset, result.RootType)
	s.Assert().Equal(domain.NormalBalanceDebit, result.NormalBalance)
}

// TestGetByIDNotFound tests handling of non-existent account
func (s *AccountsRepositoryUnitTestSuite) TestGetByIDNotFound() {
	accountID := uuid.New()

	s.mockStore.EXPECT().
		WithTenant(gomock.Any(), s.tenantID, gomock.Any()).
		DoAndReturn(func(ctx context.Context, tenantID uuid.UUID, fn func(context.Context, db.Store) error) error {
			s.mockStore.EXPECT().
				GetAccountByID(gomock.Any(), accountID).
				Return(nil, sql.ErrNoRows).
				Times(1)

			return fn(ctx, s.mockStore)
		}).
		Times(1)

	// Execute test
	result, err := s.repo.GetByID(s.ctx, accountID)

	// Assert results
	s.Assert().Error(err)
	s.Assert().Nil(result)
	s.Assert().Equal(domain.ErrAccountNotFound, err)
}

// TestGetByCodeSuccess tests successful account retrieval by code
func (s *AccountsRepositoryUnitTestSuite) TestGetByCodeSuccess() {
	accountCode := "1000-CASH"
	expectedAccount := db.FinanceChartOfAccount{
		ID:            uuid.New(),
		TenantID:      s.tenantID,
		AccountCode:   accountCode,
		AccountName:   "Cash Account",
		RootType:      db.RootTypeEnumASSET,
		NormalBalance: db.NormalBalanceEnumDEBIT,
		CurrencyCode:  "USD",
		IsActive:      true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	s.mockStore.EXPECT().
		WithTenant(gomock.Any(), s.tenantID, gomock.Any()).
		DoAndReturn(func(ctx context.Context, tenantID uuid.UUID, fn func(context.Context, db.Store) error) error {
			s.mockStore.EXPECT().
				GetAccountByCode(gomock.Any(), accountCode).
				Return(&expectedAccount, nil).
				Times(1)

			return fn(ctx, s.mockStore)
		}).
		Times(1)

	// Execute test
	result, err := s.repo.GetByCode(s.ctx, accountCode)

	// Assert results
	s.Assert().NoError(err)
	s.Assert().NotNil(result)
	s.Assert().Equal(accountCode, result.AccountCode)
	s.Assert().Equal("Cash Account", result.AccountName)
}

// TestUpdateAccountSuccess tests successful account update
func (s *AccountsRepositoryUnitTestSuite) TestUpdateAccountSuccess() {
	account := &domain.Accounts{
		ID:                 uuid.New(),
		TenantID:           s.tenantID,
		AccountCode:        "1000-CASH",
		AccountName:        "Updated Cash Account",
		AccountDescription: stringPtr("Updated description"),
		RootType:           domain.RootTypeAsset,
		AccountType:        "CASH",
		NormalBalance:      domain.NormalBalanceDebit,
		CurrencyCode:       "USD",
		IsActive:           false, // Updated to inactive
		AllowManualEntries: true,
		Version:            1,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
		UpdatedBy:          &uuid.UUID{},
	}

	expectedParams := db.UpdateAccountParams{
		ID:                          account.ID,
		AccountName:                 account.AccountName,
		AccountDescription:          getStringValue(account.AccountDescription),
		AccountSubtype:              account.AccountSubtype,
		IsMultiCurrency:             &account.IsMultiCurrency,
		CurrencyRevaluationRequired: &account.CurrencyRevaluationRequired,
		IsActive:                    account.IsActive,
		AllowManualEntries:          account.AllowManualEntries,
		RequireReference:            account.RequireReference,
		FinancialStatementLine:      account.FinancialStatementLine,
		ReportOrder:                 getInt32Ptr(account.ReportOrder),
		IsBudgetable:                getBoolPtr(account.IsBudgetable),
		BudgetVarianceThreshold:     decimalToPgNumeric(&account.BudgetVarianceThreshold),
		AccountAttributes:           mapAttributesToJSON(account.AccountAttributes),
		UpdatedBy:                   account.UpdatedBy,
	}

	s.mockStore.EXPECT().
		WithTenant(gomock.Any(), s.tenantID, gomock.Any()).
		DoAndReturn(func(ctx context.Context, tenantID uuid.UUID, fn func(context.Context, db.Store) error) error {
			s.mockStore.EXPECT().
				UpdateAccount(gomock.Any(), gomock.Eq(expectedParams)).
				Return(nil).
				Times(1)

			return fn(ctx, s.mockStore)
		}).
		Times(1)

	// Execute test
	err := s.repo.Update(s.ctx, account)

	// Assert results
	s.Assert().NoError(err)
}

// TestDeleteAccountSuccess tests successful account soft delete
func (s *AccountsRepositoryUnitTestSuite) TestDeleteAccountSuccess() {
	accountID := uuid.New()

	s.mockStore.EXPECT().
		WithTenant(gomock.Any(), s.tenantID, gomock.Any()).
		DoAndReturn(func(ctx context.Context, tenantID uuid.UUID, fn func(context.Context, db.Store) error) error {
			s.mockStore.EXPECT().
				SoftDeleteAccount(gomock.Any(), gomock.Any()).
				Return(nil).
				Times(1)

			return fn(ctx, s.mockStore)
		}).
		Times(1)

	// Execute test
	err := s.repo.Delete(s.ctx, accountID)

	// Assert results
	s.Assert().NoError(err)
}

// TestListAccountsSuccess tests successful account listing
func (s *AccountsRepositoryUnitTestSuite) TestListAccountsSuccess() {
	filter := &domain.AccountFilter{
		RootType: &domain.RootTypeAsset,
		IsActive: boolPtr(true),
		Limit:    intPtr(10),
		Offset:   intPtr(0),
	}

	expectedAccounts := []*db.FinanceChartOfAccount{
		{
			ID:            uuid.New(),
			TenantID:      s.tenantID,
			AccountCode:   "1000-CASH",
			AccountName:   "Cash Account",
			RootType:      db.RootTypeEnumASSET,
			NormalBalance: db.NormalBalanceEnumDEBIT,
			IsActive:      true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		{
			ID:            uuid.New(),
			TenantID:      s.tenantID,
			AccountCode:   "1100-BANK",
			AccountName:   "Bank Account",
			RootType:      db.RootTypeEnumASSET,
			NormalBalance: db.NormalBalanceEnumDEBIT,
			IsActive:      true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}

	s.mockStore.EXPECT().
		WithTenant(gomock.Any(), s.tenantID, gomock.Any()).
		DoAndReturn(func(ctx context.Context, tenantID uuid.UUID, fn func(context.Context, db.Store) error) error {
			s.mockStore.EXPECT().
				ListAccounts(gomock.Any(), gomock.Any()).
				Return(expectedAccounts, nil).
				Times(1)

			return fn(ctx, s.mockStore)
		}).
		Times(1)

	// Execute test
	results, err := s.repo.List(s.ctx, filter)

	// Assert results
	s.Assert().NoError(err)
	s.Assert().Len(results, 2)
	s.Assert().Equal("1000-CASH", results[0].AccountCode)
	s.Assert().Equal("1100-BANK", results[1].AccountCode)
}

// Mock PostgreSQL error for testing
type mockPgError struct {
	code string
}

func (e *mockPgError) Error() string {
	return "mock pg error"
}

func (e *mockPgError) SQLState() string {
	return e.code
}

// Helper functions specific to unit tests
func boolPtr(b bool) *bool {
	return &b
}

func intPtr(i int) *int {
	return &i
}
