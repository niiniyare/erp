package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"awo.so/internal/core/featureflag"
	"awo.so/internal/core/finance/domain"
	"awo.so/internal/core/finance/service"
	"awo.so/internal/core/iam"
	"awo.so/internal/shared"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

// ============================================================================
// Mock AccountsRepository — only methods exercised by tests are wired;
// the rest are satisfied by embedding the interface (panic if unexpectedly called).
// ============================================================================

type mockAccountsRepo struct {
	mock.Mock
	domain.AccountsRepository // satisfies unimplemented methods
}

func (m *mockAccountsRepo) ValidateAccountCode(ctx context.Context, code string, excludeID *uuid.UUID) error {
	return m.Called(ctx, code, excludeID).Error(0)
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

func (m *mockAccountsRepo) Create(ctx context.Context, account *domain.Accounts) error {
	return m.Called(ctx, account).Error(0)
}

func (m *mockAccountsRepo) Update(ctx context.Context, account *domain.Accounts) error {
	return m.Called(ctx, account).Error(0)
}

func (m *mockAccountsRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockAccountsRepo) HasTransactions(ctx context.Context, accountID uuid.UUID) (bool, error) {
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

func (m *mockAccountsRepo) GetAccountBalance(ctx context.Context, accountID uuid.UUID, asOfDate *time.Time) (*domain.AccountBalance, error) {
	args := m.Called(ctx, accountID, asOfDate)
	if v, ok := args.Get(0).(*domain.AccountBalance); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAccountsRepo) UpdateBalance(ctx context.Context, accountID uuid.UUID, balance domain.AccountBalance) error {
	return m.Called(ctx, accountID, balance).Error(0)
}

func (m *mockAccountsRepo) List(ctx context.Context, filter *domain.AccountFilter) ([]*domain.Accounts, error) {
	args := m.Called(ctx, filter)
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

// ============================================================================
// Mock AccountGroupRepository — none of the tested methods call it;
// embedding is sufficient.
// ============================================================================

type mockAccountGroupRepo struct {
	mock.Mock
	domain.AccountGroupRepository
}

// ============================================================================
// Mock iam.Service (AuthzService) — all IAM calls are TODO in account_service.go;
// embedding is sufficient.
// ============================================================================

type mockIAMService struct {
	mock.Mock
	iam.Service
}

// ============================================================================
// Mock featureflag.Service — IsEnabled is called in CreateAccount and
// DeleteAccount; the rest are satisfied by the embedded interface.
// ============================================================================

type mockFeatureFlagService struct {
	mock.Mock
	featureflag.Service
}

func (m *mockFeatureFlagService) IsEnabled(ctx context.Context, name string, evalCtx *featureflag.EvaluationContext) (bool, error) {
	args := m.Called(ctx, name, evalCtx)
	return args.Bool(0), args.Error(1)
}

// ============================================================================
// Suite
// ============================================================================

type AccountServiceSuite struct {
	suite.Suite
	req        *require.Assertions
	repo       *mockAccountsRepo
	groupRepo  *mockAccountGroupRepo
	iamSvc     *mockIAMService
	featureSvc *mockFeatureFlagService
	svc        service.AccountService
	tenantID   uuid.UUID
	ctx        context.Context
}

func TestAccountServiceSuite(t *testing.T) {
	suite.Run(t, new(AccountServiceSuite))
}

func (s *AccountServiceSuite) SetupTest() {
	s.req = require.New(s.T())
	s.tenantID = uuid.New()
	s.ctx = shared.WithTenantID(context.Background(), s.tenantID)

	s.repo = new(mockAccountsRepo)
	s.groupRepo = new(mockAccountGroupRepo)
	s.iamSvc = new(mockIAMService)
	s.featureSvc = new(mockFeatureFlagService)

	s.svc = service.NewAccountService(
		s.repo,
		s.groupRepo,
		tracing.NewNoOpService(),
		metrics.NewNoOpMetricsProvider(),
		// FIXNE: what is the correct deps for this
		// s.iamSvc,
		// s.featureSvc,
	)
}

func (s *AccountServiceSuite) TearDownTest() {
	s.repo.AssertExpectations(s.T())
	s.featureSvc.AssertExpectations(s.T())
}

// ============================================================================
// Helpers
// ============================================================================

// validCreateReq returns a minimal valid CreateAccountRequest that passes Validate().
func (s *AccountServiceSuite) validCreateReq() domain.CreateAccountRequest {
	return domain.CreateAccountRequest{
		AccountCode:        "10010101",
		AccountName:        "Cash",
		AccountType:        "Current Asset",
		RootType:           domain.RootTypeAsset,
		NormalBalance:      domain.NormalBalanceDebit,
		AllowManualEntries: true,
		IsActive:           true,
	}
}

// newAccount builds a minimal *domain.Accounts for use as a mock return value.
func (s *AccountServiceSuite) newAccount(code, name string) *domain.Accounts {
	return &domain.Accounts{
		ID:                 uuid.New(),
		TenantID:           s.tenantID,
		AccountCode:        code,
		AccountName:        name,
		AccountType:        "Current Asset",
		RootType:           domain.RootTypeAsset,
		NormalBalance:      domain.NormalBalanceDebit,
		Status:             domain.AccountStatusDraft,
		AllowManualEntries: true,
		IsActive:           true,
	}
}

// ============================================================================
// FIN-ACC-001: CreateAccount — happy path (no parent, feature flag disabled)
// ============================================================================

func (s *AccountServiceSuite) TestCreateAccount_HappyPath() {
	req := s.validCreateReq()
	created := s.newAccount(req.AccountCode, req.AccountName)

	s.featureSvc.On("IsEnabled", s.ctx, "enhanced_account_validation", mock.Anything).
		Return(false, nil).Once()
	s.repo.On("ValidateAccountCode", s.ctx, req.AccountCode, (*uuid.UUID)(nil)).
		Return(nil).Once()
	s.repo.On("Create", s.ctx, mock.Anything).Return(nil).Once()
	s.repo.On("GetByCode", s.ctx, (*uuid.UUID)(nil), req.AccountCode).
		Return(created, nil).Once()

	result, err := s.svc.CreateAccount(s.ctx, req)
	s.req.NoError(err)
	s.req.NotNil(result)
	s.req.Equal(req.AccountCode, result.AccountCode)
}

// ============================================================================
// FIN-ACC-002: CreateAccount — domain validation failure (bad account code)
// ============================================================================

func (s *AccountServiceSuite) TestCreateAccount_ValidationFailure_BadCode() {
	req := domain.CreateAccountRequest{
		AccountCode:   "BADCODE", // not 8 numeric digits
		AccountName:   "Cash",
		AccountType:   "Current Asset",
		RootType:      domain.RootTypeAsset,
		NormalBalance: domain.NormalBalanceDebit,
	}

	// Feature flag IS evaluated before Validate(); no repo calls expected.
	s.featureSvc.On("IsEnabled", s.ctx, "enhanced_account_validation", mock.Anything).
		Return(false, nil).Once()

	result, err := s.svc.CreateAccount(s.ctx, req)
	s.req.Error(err, "invalid account code must be rejected before any repo call")
	s.req.Nil(result)
}

// ============================================================================
// FIN-ACC-003: CreateAccount — duplicate code rejected by repository
// ============================================================================

func (s *AccountServiceSuite) TestCreateAccount_DuplicateCode() {
	req := s.validCreateReq()

	s.featureSvc.On("IsEnabled", s.ctx, "enhanced_account_validation", mock.Anything).
		Return(false, nil).Once()
	s.repo.On("ValidateAccountCode", s.ctx, req.AccountCode, (*uuid.UUID)(nil)).
		Return(domain.ErrAccountCodeExists).Once()

	result, err := s.svc.CreateAccount(s.ctx, req)
	s.req.Error(err)
	s.req.Nil(result)
}

// ============================================================================
// FIN-ACC-004: CreateAccount — parent account is inactive
// ============================================================================

func (s *AccountServiceSuite) TestCreateAccount_ParentInactive() {
	parentID := uuid.New()
	req := s.validCreateReq()
	req.ParentAccountID = &parentID

	inactiveParent := s.newAccount("10010000", "Current Assets")
	inactiveParent.ID = parentID
	inactiveParent.IsActive = false // parent must be active

	s.featureSvc.On("IsEnabled", s.ctx, "enhanced_account_validation", mock.Anything).
		Return(false, nil).Once()
	s.repo.On("ValidateAccountCode", s.ctx, req.AccountCode, (*uuid.UUID)(nil)).
		Return(nil).Once()
	s.repo.On("GetByID", s.ctx, parentID).Return(inactiveParent, nil).Once()

	result, err := s.svc.CreateAccount(s.ctx, req)
	s.req.Error(err, "should reject account creation under an inactive parent")
	s.req.Nil(result)
}

// ============================================================================
// FIN-ACC-005: CreateAccount — root type mismatch with parent
// ============================================================================

func (s *AccountServiceSuite) TestCreateAccount_RootTypeMismatch() {
	parentID := uuid.New()
	req := s.validCreateReq() // ASSET
	req.ParentAccountID = &parentID

	// Parent is LIABILITY — mismatch with child's ASSET root type
	parent := s.newAccount("20010000", "Current Liabilities")
	parent.ID = parentID
	parent.IsActive = true
	parent.RootType = domain.RootTypeLiability // mismatch

	s.featureSvc.On("IsEnabled", s.ctx, "enhanced_account_validation", mock.Anything).
		Return(false, nil).Once()
	s.repo.On("ValidateAccountCode", s.ctx, req.AccountCode, (*uuid.UUID)(nil)).
		Return(nil).Once()
	s.repo.On("GetByID", s.ctx, parentID).Return(parent, nil).Once()

	result, err := s.svc.CreateAccount(s.ctx, req)
	s.req.Error(err, "ROOT_TYPE_MISMATCH must be rejected before repo.Create is called")
	s.req.Nil(result)
}

// ============================================================================
// FIN-ACC-010: GetAccountByID — account exists
// ============================================================================

func (s *AccountServiceSuite) TestGetAccountByID_Found() {
	accountID := uuid.New()
	account := s.newAccount("10010101", "Cash")
	account.ID = accountID

	s.repo.On("GetByID", s.ctx, accountID).Return(account, nil).Once()

	result, err := s.svc.GetAccountByID(s.ctx, accountID)
	s.req.NoError(err)
	s.req.NotNil(result)
	s.req.Equal(accountID, result.ID)
}

// ============================================================================
// FIN-ACC-011: GetAccountByID — account not found
// ============================================================================

func (s *AccountServiceSuite) TestGetAccountByID_NotFound() {
	accountID := uuid.New()

	s.repo.On("GetByID", s.ctx, accountID).
		Return((*domain.Accounts)(nil), domain.ErrAccountNotFound).Once()

	result, err := s.svc.GetAccountByID(s.ctx, accountID)
	s.req.Error(err)
	s.req.Nil(result)
}

// ============================================================================
// FIN-ACC-012: GetAccountByCode — account found
// ============================================================================

func (s *AccountServiceSuite) TestGetAccountByCode_Found() {
	code := "10010101"
	account := s.newAccount(code, "Cash")

	// Service calls GetByCode(ctx, nil, code) — entityID is nil (no entity set in ctx)
	s.repo.On("GetByCode", s.ctx, (*uuid.UUID)(nil), code).Return(account, nil).Once()

	result, err := s.svc.GetAccountByCode(s.ctx, code)
	s.req.NoError(err)
	s.req.NotNil(result)
	s.req.Equal(code, result.AccountCode)
}

// ============================================================================
// FIN-ACC-020: UpdateAccount — happy path
// ============================================================================

func (s *AccountServiceSuite) TestUpdateAccount_HappyPath() {
	accountID := uuid.New()
	existing := s.newAccount("10010101", "Cash")
	existing.ID = accountID
	existing.Version = 1

	newName := "Petty Cash"
	req := domain.UpdateAccountRequest{
		Version:     1,
		AccountName: &newName,
	}

	s.repo.On("GetByID", s.ctx, accountID).Return(existing, nil).Once()
	// Update receives the modified `existing` pointer; use Anything to avoid
	// matching against the pre-mutation state.
	s.repo.On("Update", s.ctx, mock.Anything).Return(nil).Once()

	result, err := s.svc.UpdateAccount(s.ctx, accountID, req)
	s.req.NoError(err)
	s.req.NotNil(result)
	s.req.Equal("Petty Cash", result.AccountName)
}

// ============================================================================
// FIN-ACC-021: UpdateAccount — account not found
// ============================================================================

func (s *AccountServiceSuite) TestUpdateAccount_NotFound() {
	accountID := uuid.New()
	req := domain.UpdateAccountRequest{Version: 1}

	s.repo.On("GetByID", s.ctx, accountID).
		Return((*domain.Accounts)(nil), domain.ErrAccountNotFound).Once()

	result, err := s.svc.UpdateAccount(s.ctx, accountID, req)
	s.req.Error(err)
	s.req.Nil(result)
}

// ============================================================================
// FIN-ACC-030: DeleteAccount — happy path
// ============================================================================

func (s *AccountServiceSuite) TestDeleteAccount_HappyPath() {
	accountID := uuid.New()
	account := s.newAccount("10010101", "Cash")
	account.ID = accountID
	account.IsSystemAccount = false

	s.featureSvc.On("IsEnabled", mock.Anything, "enhanced_account_deletion", mock.Anything).
		Return(false, nil).Once()
	s.repo.On("GetByID", mock.Anything, accountID).Return(account, nil).Once()
	s.repo.On("HasTransactions", mock.Anything, accountID).Return(false, nil).Once()
	s.repo.On("GetChildren", mock.Anything, accountID).
		Return([]*domain.Accounts{}, nil).Once()
	s.repo.On("Delete", mock.Anything, accountID).Return(nil).Once()

	err := s.svc.DeleteAccount(s.ctx, accountID)
	s.req.NoError(err)
}

// ============================================================================
// FIN-ACC-031: DeleteAccount — system accounts cannot be deleted
// ============================================================================

func (s *AccountServiceSuite) TestDeleteAccount_SystemAccount() {
	accountID := uuid.New()
	account := s.newAccount("10010101", "Cash")
	account.ID = accountID
	account.IsSystemAccount = true // protected

	s.featureSvc.On("IsEnabled", mock.Anything, "enhanced_account_deletion", mock.Anything).
		Return(false, nil).Once()
	s.repo.On("GetByID", mock.Anything, accountID).Return(account, nil).Once()

	err := s.svc.DeleteAccount(s.ctx, accountID)
	s.req.Error(err, "SYSTEM_ACCOUNT must be blocked before HasTransactions/GetChildren/Delete are called")
}

// ============================================================================
// FIN-ACC-032: DeleteAccount — account with transactions cannot be deleted
// ============================================================================

func (s *AccountServiceSuite) TestDeleteAccount_HasTransactions() {
	accountID := uuid.New()
	account := s.newAccount("10010101", "Cash")
	account.ID = accountID
	account.IsSystemAccount = false

	s.featureSvc.On("IsEnabled", mock.Anything, "enhanced_account_deletion", mock.Anything).
		Return(false, nil).Once()
	s.repo.On("GetByID", mock.Anything, accountID).Return(account, nil).Once()
	s.repo.On("HasTransactions", mock.Anything, accountID).Return(true, nil).Once()

	err := s.svc.DeleteAccount(s.ctx, accountID)
	s.req.Error(err, "HAS_TRANSACTIONS must block deletion before GetChildren/Delete are called")
}

// ============================================================================
// FIN-ACC-033: DeleteAccount — account with children cannot be deleted
// ============================================================================

func (s *AccountServiceSuite) TestDeleteAccount_HasChildren() {
	accountID := uuid.New()
	account := s.newAccount("10010101", "Cash")
	account.ID = accountID
	account.IsSystemAccount = false

	child := s.newAccount("10010102", "Petty Cash")
	children := []*domain.Accounts{child}

	s.featureSvc.On("IsEnabled", mock.Anything, "enhanced_account_deletion", mock.Anything).
		Return(false, nil).Once()
	s.repo.On("GetByID", mock.Anything, accountID).Return(account, nil).Once()
	s.repo.On("HasTransactions", mock.Anything, accountID).Return(false, nil).Once()
	s.repo.On("GetChildren", mock.Anything, accountID).Return(children, nil).Once()

	err := s.svc.DeleteAccount(s.ctx, accountID)
	s.req.Error(err, "HAS_CHILDREN must block deletion before Delete is called")
}

// ============================================================================
// FIN-ACC-005: CreateAccount — currency defaults to DefaultBaseCurrency
// ============================================================================

func (s *AccountServiceSuite) TestCreateAccount_DefaultCurrency() {
	req := s.validCreateReq() // CurrencyCode is nil

	created := s.newAccount(req.AccountCode, req.AccountName)
	defCurrency := domain.DefaultBaseCurrency
	created.CurrencyCode = &defCurrency

	s.featureSvc.On("IsEnabled", s.ctx, "enhanced_account_validation", mock.Anything).
		Return(false, nil).Once()
	s.repo.On("ValidateAccountCode", s.ctx, req.AccountCode, (*uuid.UUID)(nil)).
		Return(nil).Once()
	// Verify the service sets CurrencyCode before calling Create.
	s.repo.On("Create", s.ctx, mock.MatchedBy(func(a *domain.Accounts) bool {
		return a.CurrencyCode != nil && *a.CurrencyCode == domain.DefaultBaseCurrency
	})).Return(nil).Once()
	s.repo.On("GetByCode", s.ctx, (*uuid.UUID)(nil), req.AccountCode).
		Return(created, nil).Once()

	result, err := s.svc.CreateAccount(s.ctx, req)
	s.req.NoError(err)
	s.req.NotNil(result)
	s.req.NotNil(result.CurrencyCode, "currency must be set on the returned account")
	s.req.Equal(domain.DefaultBaseCurrency, *result.CurrencyCode)
}

// ============================================================================
// FIN-ACC-004: CreateAccount — path is derived from parent
// ============================================================================

func (s *AccountServiceSuite) TestCreateAccount_PathFromParent() {
	parentID := uuid.New()
	req := s.validCreateReq()
	req.ParentAccountID = &parentID

	parent := s.newAccount("10010000", "Current Assets")
	parent.ID = parentID
	parent.IsActive = true
	parent.RootType = domain.RootTypeAsset
	parent.AccountLevel = 1
	parent.Path = domain.MaterialisedPath("") // root sentinel

	created := s.newAccount(req.AccountCode, req.AccountName)
	created.ParentAccountID = &parentID
	created.AccountLevel = 2

	s.featureSvc.On("IsEnabled", s.ctx, "enhanced_account_validation", mock.Anything).
		Return(false, nil).Once()
	s.repo.On("ValidateAccountCode", s.ctx, req.AccountCode, (*uuid.UUID)(nil)).
		Return(nil).Once()
	s.repo.On("GetByID", s.ctx, parentID).Return(parent, nil).Once()
	s.repo.On("Create", s.ctx, mock.MatchedBy(func(a *domain.Accounts) bool {
		return a.ParentAccountID != nil && *a.ParentAccountID == parentID && a.AccountLevel == 2
	})).Return(nil).Once()
	s.repo.On("GetByCode", s.ctx, (*uuid.UUID)(nil), req.AccountCode).
		Return(created, nil).Once()

	result, err := s.svc.CreateAccount(s.ctx, req)
	s.req.NoError(err)
	s.req.NotNil(result)
	s.req.NotNil(result.ParentAccountID)
	s.req.Equal(parentID, *result.ParentAccountID)
}

// ============================================================================
// FIN-ACC-040: ListAccounts — filter is passed through to repository
// ============================================================================

func (s *AccountServiceSuite) TestListAccounts_FilterByRootType() {
	rt := domain.RootTypeAsset
	filter := &domain.AccountFilter{RootType: &rt}

	accounts := []*domain.Accounts{
		s.newAccount("10010101", "Cash"),
		s.newAccount("10020101", "Bank"),
		s.newAccount("10030101", "AR"),
	}

	s.featureSvc.On("IsEnabled", s.ctx, "enhanced_account_listing", mock.Anything).
		Return(false, nil).Once()
	s.repo.On("List", s.ctx, mock.Anything).Return(accounts, nil).Once()

	result, err := s.svc.ListAccounts(s.ctx, filter)
	s.req.NoError(err)
	s.req.Len(result, 3)
	for _, a := range result {
		s.req.Equal(domain.RootTypeAsset, a.RootType)
	}
}

// ============================================================================
// FIN-ACC-050: GetAccountHierarchy — returns full subtree from repository
// ============================================================================

func (s *AccountServiceSuite) TestGetAccountHierarchy_FullSubtree() {
	rootID := uuid.New()
	hierarchy := []*domain.Accounts{
		s.newAccount("10000000", "Assets"),
		s.newAccount("10010000", "Current Assets"),
		s.newAccount("10010100", "Cash & Equivalents"),
		s.newAccount("10010101", "Petty Cash"),
		s.newAccount("10010102", "Main Account"),
		s.newAccount("10020000", "Non-Current Assets"),
	}

	s.repo.On("GetAccountHierarchy", s.ctx, rootID).Return(hierarchy, nil).Once()

	result, err := s.svc.GetAccountHierarchy(s.ctx, rootID)
	s.req.NoError(err)
	s.req.Len(result, 6, "all 6 nodes in the subtree must be returned")
}
