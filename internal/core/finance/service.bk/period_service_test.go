package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
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
// Mock PeriodRepository
// ============================================================================

type mockPeriodRepo struct {
	mock.Mock
}

func (m *mockPeriodRepo) CreateFiscalYear(ctx context.Context, fy *domain.FiscalYear) error {
	return m.Called(ctx, fy).Error(0)
}

func (m *mockPeriodRepo) GetFiscalYearByID(ctx context.Context, id uuid.UUID) (*domain.FiscalYear, error) {
	args := m.Called(ctx, id)
	if v, ok := args.Get(0).(*domain.FiscalYear); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockPeriodRepo) GetFiscalYearByYear(ctx context.Context, tenantID uuid.UUID, year int) (*domain.FiscalYear, error) {
	args := m.Called(ctx, tenantID, year)
	if v, ok := args.Get(0).(*domain.FiscalYear); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockPeriodRepo) ListFiscalYears(ctx context.Context, tenantID uuid.UUID) ([]*domain.FiscalYear, error) {
	args := m.Called(ctx, tenantID)
	if v, ok := args.Get(0).([]*domain.FiscalYear); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockPeriodRepo) UpdateFiscalYear(ctx context.Context, fy *domain.FiscalYear) error {
	return m.Called(ctx, fy).Error(0)
}

func (m *mockPeriodRepo) CreatePeriod(ctx context.Context, period *domain.AccountingPeriod) error {
	return m.Called(ctx, period).Error(0)
}

func (m *mockPeriodRepo) GetPeriodByID(ctx context.Context, id uuid.UUID) (*domain.AccountingPeriod, error) {
	args := m.Called(ctx, id)
	if v, ok := args.Get(0).(*domain.AccountingPeriod); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockPeriodRepo) GetPeriodForDate(ctx context.Context, tenantID uuid.UUID, date time.Time) (*domain.AccountingPeriod, error) {
	args := m.Called(ctx, tenantID, date)
	if v, ok := args.Get(0).(*domain.AccountingPeriod); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockPeriodRepo) GetCurrentPeriod(ctx context.Context, tenantID uuid.UUID) (*domain.AccountingPeriod, error) {
	args := m.Called(ctx, tenantID)
	if v, ok := args.Get(0).(*domain.AccountingPeriod); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockPeriodRepo) ListPeriods(ctx context.Context, tenantID, fiscalYearID uuid.UUID) ([]*domain.AccountingPeriod, error) {
	args := m.Called(ctx, tenantID, fiscalYearID)
	if v, ok := args.Get(0).([]*domain.AccountingPeriod); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockPeriodRepo) UpdatePeriod(ctx context.Context, period *domain.AccountingPeriod) error {
	return m.Called(ctx, period).Error(0)
}

// ============================================================================
// Suite
// ============================================================================

type PeriodServiceSuite struct {
	suite.Suite
	req      *require.Assertions
	repo     *mockPeriodRepo
	svc      service.PeriodService
	tenantID uuid.UUID
	ctx      context.Context
}

func TestPeriodServiceSuite(t *testing.T) {
	suite.Run(t, new(PeriodServiceSuite))
}

func (s *PeriodServiceSuite) SetupTest() {
	s.req = require.New(s.T())
	s.tenantID = uuid.New()
	s.ctx = shared.WithTenantID(context.Background(), s.tenantID)

	s.repo = new(mockPeriodRepo)
	s.svc = service.NewPeriodService(
		s.repo,
		tracing.NewNoOpService(),
		metrics.NewNoOpMetricsProvider(),
	)
}

func (s *PeriodServiceSuite) TearDownTest() {
	s.repo.AssertExpectations(s.T())
}

// ============================================================================
// Helpers
// ============================================================================

func (s *PeriodServiceSuite) newFY(start, end time.Time) *domain.FiscalYear {
	return &domain.FiscalYear{
		ID:        uuid.New(),
		TenantID:  s.tenantID,
		Name:      "FY" + start.Format("2006"),
		StartDate: start,
		EndDate:   end,
		CreatedBy: uuid.New(),
	}
}

func (s *PeriodServiceSuite) openPeriod(start, end time.Time) *domain.AccountingPeriod {
	return &domain.AccountingPeriod{
		ID:           uuid.New(),
		TenantID:     s.tenantID,
		FiscalYearID: uuid.New(),
		Name:         start.Format("January 2006"),
		StartDate:    start,
		EndDate:      end,
		Status:       domain.PeriodStatusOpen,
		CreatedBy:    uuid.New(),
	}
}

// ============================================================================
// FIN-PER-001: FiscalYear — overlapping years rejected at repository level
// ============================================================================

func (s *PeriodServiceSuite) TestFiscalYear_OverlapRejectedByRepository() {
	// The service delegates overlap detection to the repository (DB unique constraint).
	// Simulate a repository that returns a conflict error on overlap.
	fy := s.newFY(
		time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC),
	)

	s.repo.On("CreateFiscalYear", s.ctx, fy).
		Return(domain.ErrInvalidDateRange) // simulate DB constraint / overlap error

	result, err := s.svc.CreateFiscalYear(s.ctx, fy)
	s.req.Error(err)
	s.req.Nil(result)
}

// ============================================================================
// FIN-PER-002: FiscalYear — end date must be after start date
// ============================================================================

func (s *PeriodServiceSuite) TestFiscalYear_DateRangeValidation_EndBeforeStart() {
	fy := s.newFY(
		time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC), // start = Dec 31
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),  // end = Jan 1 (before start)
	)

	// Repo must NOT be called when validation fails at service level
	result, err := s.svc.CreateFiscalYear(s.ctx, fy)
	s.req.Error(err, "inverted date range must be rejected before any repo call")
	s.req.Nil(result)
}

func (s *PeriodServiceSuite) TestFiscalYear_DateRangeValidation_EqualDates() {
	fy := s.newFY(
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), // start == end
	)

	// !EndDate.After(StartDate) catches equal dates too
	result, err := s.svc.CreateFiscalYear(s.ctx, fy)
	s.req.Error(err)
	s.req.Nil(result)
}

func (s *PeriodServiceSuite) TestFiscalYear_DateRangeValidation_ValidRange() {
	fy := s.newFY(
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
	)

	s.repo.On("CreateFiscalYear", s.ctx, fy).Return(nil)

	result, err := s.svc.CreateFiscalYear(s.ctx, fy)
	s.req.NoError(err)
	s.req.NotNil(result)
}

func (s *PeriodServiceSuite) TestFiscalYear_RequiresName() {
	fy := s.newFY(
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
	)
	fy.Name = "" // missing name

	result, err := s.svc.CreateFiscalYear(s.ctx, fy)
	s.req.Error(err)
	s.req.Nil(result)
}

// ============================================================================
// FIN-PER-012: PeriodService.GetCurrentPeriod — delegates to repo
// ============================================================================

func (s *PeriodServiceSuite) TestGetCurrentPeriod_ReturnsOpenPeriod() {
	period := s.openPeriod(
		time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 4, 30, 0, 0, 0, 0, time.UTC),
	)

	s.repo.On("GetCurrentPeriod", s.ctx, s.tenantID).Return(period, nil)

	result, err := s.svc.GetCurrentPeriod(s.ctx)
	s.req.NoError(err)
	s.req.NotNil(result)
	s.req.Equal(domain.PeriodStatusOpen, result.Status)
}

func (s *PeriodServiceSuite) TestGetCurrentPeriod_NoPeriodReturnsError() {
	s.repo.On("GetCurrentPeriod", s.ctx, s.tenantID).Return(nil, domain.ErrPeriodNotFound)

	result, err := s.svc.GetCurrentPeriod(s.ctx)
	s.req.Error(err)
	s.req.Nil(result)
}

// ============================================================================
// Period status gate: ChangePeriodStatus
// ============================================================================

func (s *PeriodServiceSuite) TestChangePeriodStatus_OpenToSoftClosed() {
	period := s.openPeriod(
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
	)
	byUser := uuid.New()

	s.repo.On("GetPeriodByID", s.ctx, period.ID).Return(period, nil)
	s.repo.On("UpdatePeriod", s.ctx, mock.MatchedBy(func(p *domain.AccountingPeriod) bool {
		return p.Status == domain.PeriodStatusSoftClosed
	})).Return(nil)

	result, err := s.svc.ChangePeriodStatus(s.ctx, period.ID, domain.PeriodStatusSoftClosed, byUser, true)
	s.req.NoError(err)
	s.req.Equal(domain.PeriodStatusSoftClosed, result.Status)
}

func (s *PeriodServiceSuite) TestChangePeriodStatus_InvalidTransitionReturnsError() {
	// OPEN → LOCKED is not a direct transition (must go OPEN→SOFT_CLOSED→HARD_CLOSED→LOCKED)
	period := s.openPeriod(
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
	)
	byUser := uuid.New()

	s.repo.On("GetPeriodByID", s.ctx, period.ID).Return(period, nil)

	result, err := s.svc.ChangePeriodStatus(s.ctx, period.ID, domain.PeriodStatusLocked, byUser, true)
	s.req.Error(err)
	s.req.Nil(result)
}

// ============================================================================
// ListPeriods — delegates to repo
// ============================================================================

func (s *PeriodServiceSuite) TestListPeriods_ReturnsPeriods() {
	fyID := uuid.New()
	periods := []*domain.AccountingPeriod{
		s.openPeriod(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)),
		s.openPeriod(time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC), time.Date(2025, 2, 28, 0, 0, 0, 0, time.UTC)),
	}

	s.repo.On("ListPeriods", s.ctx, s.tenantID, fyID).Return(periods, nil)

	result, err := s.svc.ListPeriods(s.ctx, fyID)
	s.req.NoError(err)
	s.req.Len(result, 2)
}

// ============================================================================
// FIN-PER-012 complement: HARD_CLOSED period cannot accept postings
// ============================================================================

func (s *PeriodServiceSuite) TestGetCurrentPeriod_HardClosedPeriodNotUsable() {
	period := &domain.AccountingPeriod{
		ID:           uuid.New(),
		TenantID:     s.tenantID,
		FiscalYearID: uuid.New(),
		Name:         "January 2025",
		StartDate:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:      time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
		Status:       domain.PeriodStatusHardClosed,
		CreatedBy:    uuid.New(),
	}

	s.repo.On("GetCurrentPeriod", s.ctx, s.tenantID).Return(period, nil)

	result, err := s.svc.GetCurrentPeriod(s.ctx)
	s.req.NoError(err)
	s.req.NotNil(result)
	s.req.Equal(domain.PeriodStatusHardClosed, result.Status)
	// HARD_CLOSED period must not allow posting — IsOpen() delegates to AllowsPosting().
	s.req.False(result.IsOpen(), "HARD_CLOSED period must not accept new postings")
}
