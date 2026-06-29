package service_test

import (
	"context"
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
// Mock BudgetRepository
// ============================================================================

type mockBudgetRepo struct {
	mock.Mock
}

func (m *mockBudgetRepo) CreateBudget(ctx context.Context, b *domain.Budget) error {
	args := m.Called(ctx, b)
	return args.Error(0)
}

func (m *mockBudgetRepo) GetBudgetByID(ctx context.Context, id uuid.UUID) (*domain.Budget, error) {
	args := m.Called(ctx, id)
	if b, ok := args.Get(0).(*domain.Budget); ok {
		return b, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockBudgetRepo) ListBudgets(ctx context.Context, tenantID uuid.UUID, fiscalYearID *uuid.UUID) ([]*domain.Budget, error) {
	args := m.Called(ctx, tenantID, fiscalYearID)
	if v, ok := args.Get(0).([]*domain.Budget); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockBudgetRepo) UpdateBudget(ctx context.Context, b *domain.Budget) error {
	args := m.Called(ctx, b)
	return args.Error(0)
}

func (m *mockBudgetRepo) CreateLineItems(ctx context.Context, lines []*domain.BudgetLineItem) error {
	args := m.Called(ctx, lines)
	return args.Error(0)
}

func (m *mockBudgetRepo) GetLineItems(ctx context.Context, budgetID uuid.UUID) ([]*domain.BudgetLineItem, error) {
	args := m.Called(ctx, budgetID)
	if v, ok := args.Get(0).([]*domain.BudgetLineItem); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockBudgetRepo) DeleteLineItems(ctx context.Context, budgetID uuid.UUID) error {
	args := m.Called(ctx, budgetID)
	return args.Error(0)
}

// ============================================================================
// Suite
// ============================================================================

type BudgetServiceSuite struct {
	suite.Suite
	req      *require.Assertions
	repo     *mockBudgetRepo
	svc      service.BudgetService
	tenantID uuid.UUID
	ctx      context.Context
}

func TestBudgetServiceSuite(t *testing.T) {
	suite.Run(t, new(BudgetServiceSuite))
}

func (s *BudgetServiceSuite) SetupTest() {
	s.req = require.New(s.T())
	s.tenantID = uuid.New()
	s.ctx = shared.WithTenantID(context.Background(), s.tenantID)

	s.repo = new(mockBudgetRepo)
	s.svc = service.NewBudgetService(
		s.repo,
		tracing.NewNoOpService(),
		metrics.NewNoOpMetricsProvider(),
	)
}

func (s *BudgetServiceSuite) TearDownTest() {
	s.repo.AssertExpectations(s.T())
}

// ============================================================================
// Helpers
// ============================================================================

func (s *BudgetServiceSuite) newDraftBudget() *domain.Budget {
	return &domain.Budget{
		TenantID:     s.tenantID,
		FiscalYearID: uuid.New(),
		Name:         "FY2026 Budget",
		BudgetType:   domain.BudgetTypeAnnual,
		Status:       domain.BudgetStatusDraft,
		CurrencyCode: "USD",
		Version:      1,
	}
}

func (s *BudgetServiceSuite) seedBudget(status domain.BudgetStatus) *domain.Budget {
	b := s.newDraftBudget()
	b.ID = uuid.New()
	b.Status = status
	b.CreatedAt = time.Now()
	b.UpdatedAt = time.Now()
	return b
}

// ============================================================================
// CreateBudget
// ============================================================================

func (s *BudgetServiceSuite) TestCreateBudget_ValidInput_SetsDefaultsAndPersists() {
	b := s.newDraftBudget()
	b.Status = "" // service should default to DRAFT
	b.Version = 0 // service should default to 1

	s.repo.On("CreateBudget", s.ctx, mock.MatchedBy(func(arg *domain.Budget) bool {
		return arg.Status == domain.BudgetStatusDraft && arg.Version == 1
	})).Return(nil)

	result, err := s.svc.CreateBudget(s.ctx, b, nil)
	s.req.NoError(err)
	s.req.NotNil(result)
	s.req.Equal(domain.BudgetStatusDraft, result.Status)
	s.req.Equal(1, result.Version)
}

func (s *BudgetServiceSuite) TestCreateBudget_WithLineItems_CreatesLineItems() {
	b := s.newDraftBudget()
	b.ID = uuid.New()
	lines := []*domain.BudgetLineItem{
		{AccountID: uuid.New(), BudgetedAmount: decimal.NewFromInt(10_000)},
	}

	s.repo.On("CreateBudget", s.ctx, mock.AnythingOfType("*domain.Budget")).Return(nil)
	s.repo.On("CreateLineItems", s.ctx, mock.MatchedBy(func(arg []*domain.BudgetLineItem) bool {
		return len(arg) == 1 && arg[0].TenantID == s.tenantID
	})).Return(nil)

	result, err := s.svc.CreateBudget(s.ctx, b, lines)
	s.req.NoError(err)
	s.req.Len(result.Lines, 1)
}

func (s *BudgetServiceSuite) TestCreateBudget_InvalidName_ReturnsValidationError() {
	b := s.newDraftBudget()
	b.Name = "" // required

	// repo should NOT be called when validation fails
	result, err := s.svc.CreateBudget(s.ctx, b, nil)
	s.req.Error(err)
	s.req.Nil(result)
	// no repo.On() call → AssertExpectations verifies nothing was called
}

func (s *BudgetServiceSuite) TestCreateBudget_InvalidCurrencyCode_ReturnsValidationError() {
	b := s.newDraftBudget()
	b.CurrencyCode = "TOOLONG"

	result, err := s.svc.CreateBudget(s.ctx, b, nil)
	s.req.Error(err)
	s.req.Nil(result)
}

func (s *BudgetServiceSuite) TestCreateBudget_NoTenantInContext_ReturnsError() {
	ctx := context.Background() // no tenant ID
	b := s.newDraftBudget()

	result, err := s.svc.CreateBudget(ctx, b, nil)
	s.req.Error(err)
	s.req.Nil(result)
}

func (s *BudgetServiceSuite) TestCreateBudget_RepoError_ReturnsWrappedError() {
	b := s.newDraftBudget()

	s.repo.On("CreateBudget", s.ctx, mock.AnythingOfType("*domain.Budget")).
		Return(domain.ErrBudgetNotFound) // simulate unexpected repo error

	result, err := s.svc.CreateBudget(s.ctx, b, nil)
	s.req.Error(err)
	s.req.Nil(result)
}

// ============================================================================
// GetBudget
// ============================================================================

func (s *BudgetServiceSuite) TestGetBudget_Found_ReturnsWithLineItems() {
	b := s.seedBudget(domain.BudgetStatusApproved)
	lines := []*domain.BudgetLineItem{
		{ID: uuid.New(), BudgetID: b.ID, AccountID: uuid.New(), BudgetedAmount: decimal.NewFromInt(5_000)},
	}

	s.repo.On("GetBudgetByID", s.ctx, b.ID).Return(b, nil)
	s.repo.On("GetLineItems", s.ctx, b.ID).Return(lines, nil)

	result, err := s.svc.GetBudget(s.ctx, b.ID)
	s.req.NoError(err)
	s.req.Equal(b.ID, result.ID)
	s.req.Len(result.Lines, 1)
}

func (s *BudgetServiceSuite) TestGetBudget_NotFound_ReturnsNotFoundError() {
	id := uuid.New()
	s.repo.On("GetBudgetByID", s.ctx, id).Return(nil, domain.ErrBudgetNotFound)

	result, err := s.svc.GetBudget(s.ctx, id)
	s.req.Error(err)
	s.req.Nil(result)
}

// ============================================================================
// ListBudgets
// ============================================================================

func (s *BudgetServiceSuite) TestListBudgets_NoFilter_ReturnsAll() {
	budgets := []*domain.Budget{s.seedBudget(domain.BudgetStatusApproved), s.seedBudget(domain.BudgetStatusDraft)}
	s.repo.On("ListBudgets", s.ctx, s.tenantID, (*uuid.UUID)(nil)).Return(budgets, nil)

	result, err := s.svc.ListBudgets(s.ctx, nil)
	s.req.NoError(err)
	s.req.Len(result, 2)
}

func (s *BudgetServiceSuite) TestListBudgets_WithFiscalYearFilter_PassesFilter() {
	fyID := uuid.New()
	expected := []*domain.Budget{s.seedBudget(domain.BudgetStatusApproved)}
	s.repo.On("ListBudgets", s.ctx, s.tenantID, &fyID).Return(expected, nil)

	result, err := s.svc.ListBudgets(s.ctx, &fyID)
	s.req.NoError(err)
	s.req.Len(result, 1)
}

// ============================================================================
// SubmitBudget
// ============================================================================

func (s *BudgetServiceSuite) TestSubmitBudget_FromDraft_TransitionsToSubmitted() {
	b := s.seedBudget(domain.BudgetStatusDraft)
	userID := uuid.New()

	s.repo.On("GetBudgetByID", s.ctx, b.ID).Return(b, nil)
	s.repo.On("UpdateBudget", s.ctx, mock.MatchedBy(func(arg *domain.Budget) bool {
		return arg.Status == domain.BudgetStatusSubmitted &&
			arg.SubmittedBy != nil && *arg.SubmittedBy == userID
	})).Return(nil)

	result, err := s.svc.SubmitBudget(s.ctx, b.ID, userID)
	s.req.NoError(err)
	s.req.Equal(domain.BudgetStatusSubmitted, result.Status)
	s.req.NotNil(result.SubmittedAt)
	s.req.Equal(userID, *result.SubmittedBy)
}

func (s *BudgetServiceSuite) TestSubmitBudget_FromRejected_AllowedTransition() {
	b := s.seedBudget(domain.BudgetStatusRejected)
	userID := uuid.New()

	s.repo.On("GetBudgetByID", s.ctx, b.ID).Return(b, nil)
	s.repo.On("UpdateBudget", s.ctx, mock.Anything).Return(nil)

	result, err := s.svc.SubmitBudget(s.ctx, b.ID, userID)
	s.req.NoError(err)
	s.req.Equal(domain.BudgetStatusSubmitted, result.Status)
}

func (s *BudgetServiceSuite) TestSubmitBudget_FromApproved_ReturnsConflictError() {
	b := s.seedBudget(domain.BudgetStatusApproved)
	userID := uuid.New()

	s.repo.On("GetBudgetByID", s.ctx, b.ID).Return(b, nil)

	result, err := s.svc.SubmitBudget(s.ctx, b.ID, userID)
	s.req.Error(err)
	s.req.Nil(result)
}

func (s *BudgetServiceSuite) TestSubmitBudget_FromClosed_ReturnsConflictError() {
	b := s.seedBudget(domain.BudgetStatusClosed)
	userID := uuid.New()

	s.repo.On("GetBudgetByID", s.ctx, b.ID).Return(b, nil)

	_, err := s.svc.SubmitBudget(s.ctx, b.ID, userID)
	s.req.Error(err)
}

// ============================================================================
// ApproveBudget
// ============================================================================

func (s *BudgetServiceSuite) TestApproveBudget_FromSubmitted_TransitionsToApproved() {
	b := s.seedBudget(domain.BudgetStatusSubmitted)
	approverID := uuid.New()

	s.repo.On("GetBudgetByID", s.ctx, b.ID).Return(b, nil)
	s.repo.On("UpdateBudget", s.ctx, mock.MatchedBy(func(arg *domain.Budget) bool {
		return arg.Status == domain.BudgetStatusApproved &&
			arg.ApprovedBy != nil && *arg.ApprovedBy == approverID
	})).Return(nil)

	result, err := s.svc.ApproveBudget(s.ctx, b.ID, approverID)
	s.req.NoError(err)
	s.req.Equal(domain.BudgetStatusApproved, result.Status)
	s.req.NotNil(result.ApprovedAt)
	s.req.Equal(approverID, *result.ApprovedBy)
}

func (s *BudgetServiceSuite) TestApproveBudget_FromDraft_ReturnsConflictError() {
	// SOD: a budget cannot be self-approved from draft — must go through submit flow.
	b := s.seedBudget(domain.BudgetStatusDraft)
	approverID := uuid.New()

	s.repo.On("GetBudgetByID", s.ctx, b.ID).Return(b, nil)

	result, err := s.svc.ApproveBudget(s.ctx, b.ID, approverID)
	s.req.Error(err)
	s.req.Nil(result)
}

func (s *BudgetServiceSuite) TestApproveBudget_AlreadyApproved_ReturnsConflictError() {
	b := s.seedBudget(domain.BudgetStatusApproved)
	approverID := uuid.New()

	s.repo.On("GetBudgetByID", s.ctx, b.ID).Return(b, nil)

	result, err := s.svc.ApproveBudget(s.ctx, b.ID, approverID)
	s.req.Error(err)
	s.req.Nil(result)
}

// ============================================================================
// RejectBudget
// ============================================================================

func (s *BudgetServiceSuite) TestRejectBudget_FromSubmitted_TransitionsToRejected() {
	b := s.seedBudget(domain.BudgetStatusSubmitted)
	reviewerID := uuid.New()
	note := "Missing supporting documentation"

	s.repo.On("GetBudgetByID", s.ctx, b.ID).Return(b, nil)
	s.repo.On("UpdateBudget", s.ctx, mock.MatchedBy(func(arg *domain.Budget) bool {
		return arg.Status == domain.BudgetStatusRejected &&
			arg.RejectedBy != nil && *arg.RejectedBy == reviewerID &&
			arg.RejectNote != nil && *arg.RejectNote == note
	})).Return(nil)

	result, err := s.svc.RejectBudget(s.ctx, b.ID, reviewerID, note)
	s.req.NoError(err)
	s.req.Equal(domain.BudgetStatusRejected, result.Status)
	s.req.NotNil(result.RejectedAt)
	s.req.NotNil(result.RejectNote)
	s.req.Equal(note, *result.RejectNote)
}

func (s *BudgetServiceSuite) TestRejectBudget_FromDraft_ReturnsConflictError() {
	b := s.seedBudget(domain.BudgetStatusDraft)
	reviewerID := uuid.New()

	s.repo.On("GetBudgetByID", s.ctx, b.ID).Return(b, nil)

	result, err := s.svc.RejectBudget(s.ctx, b.ID, reviewerID, "bad budget")
	s.req.Error(err)
	s.req.Nil(result)
}

// ============================================================================
// CloseBudget
// ============================================================================

func (s *BudgetServiceSuite) TestCloseBudget_FromApproved_TransitionsToClosed() {
	b := s.seedBudget(domain.BudgetStatusApproved)

	s.repo.On("GetBudgetByID", s.ctx, b.ID).Return(b, nil)
	s.repo.On("UpdateBudget", s.ctx, mock.MatchedBy(func(arg *domain.Budget) bool {
		return arg.Status == domain.BudgetStatusClosed
	})).Return(nil)

	result, err := s.svc.CloseBudget(s.ctx, b.ID)
	s.req.NoError(err)
	s.req.Equal(domain.BudgetStatusClosed, result.Status)
}

func (s *BudgetServiceSuite) TestCloseBudget_FromDraft_ReturnsConflictError() {
	b := s.seedBudget(domain.BudgetStatusDraft)

	s.repo.On("GetBudgetByID", s.ctx, b.ID).Return(b, nil)

	result, err := s.svc.CloseBudget(s.ctx, b.ID)
	s.req.Error(err)
	s.req.Nil(result)
}

// ============================================================================
// GetLineItems
// ============================================================================

func (s *BudgetServiceSuite) TestGetLineItems_ReturnsItems() {
	budgetID := uuid.New()
	lines := []*domain.BudgetLineItem{
		{ID: uuid.New(), BudgetID: budgetID, AccountID: uuid.New(), BudgetedAmount: decimal.NewFromInt(1_000)},
		{ID: uuid.New(), BudgetID: budgetID, AccountID: uuid.New(), BudgetedAmount: decimal.NewFromInt(2_000)},
	}

	s.repo.On("GetLineItems", s.ctx, budgetID).Return(lines, nil)

	result, err := s.svc.GetLineItems(s.ctx, budgetID)
	s.req.NoError(err)
	s.req.Len(result, 2)
}
