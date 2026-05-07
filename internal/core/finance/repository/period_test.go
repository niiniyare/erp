//go:build database
// +build database

// Package repository_test contains real-PostgreSQL integration tests for the
// finance period repository.
//
// Run with:
//
//	DB_URL="postgres://..." go test -tags database ./internal/core/finance/repository/... -v
package repository_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	db "awo.so/db/sqlc"
	"awo.so/internal/core/finance/domain"
	"awo.so/internal/core/finance/repository"
	"awo.so/internal/shared"
	"awo.so/internal/shared/tracing"
)

// ============================================================================
// Suite
// ============================================================================

type PeriodRepoSuite struct {
	suite.Suite
	req      *require.Assertions
	pool     *pgxpool.Pool
	store    db.Store
	repo     domain.PeriodRepository
	tenantID uuid.UUID
	tenantB  uuid.UUID
	ctx      context.Context
}

func TestPeriodRepoSuite(t *testing.T) {
	suite.Run(t, new(PeriodRepoSuite))
}

func (s *PeriodRepoSuite) SetupSuite() {
	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		s.T().Skip("DB_URL not set — skipping period repository integration tests")
	}

	store, err := db.NewDB(dsn)
	if err != nil {
		s.T().Skipf("cannot connect to test DB: %v", err)
	}
	s.store = store
	s.pool = store.GetPool()
	s.repo = repository.NewPeriodRepository(store, tracing.NewNoOpService(), nil, nil, nil)
	s.tenantID = uuid.New()
	s.tenantB = uuid.New()
	s.ctx = context.Background()

	if err := s.seedTenant(s.tenantID, "period-a"); err != nil {
		s.T().Skipf("cannot seed tenant A: %v", err)
	}
	if err := s.seedTenant(s.tenantB, "period-b"); err != nil {
		s.T().Skipf("cannot seed tenant B: %v", err)
	}
}

func (s *PeriodRepoSuite) TearDownSuite() {
	if s.pool == nil {
		return
	}
	ctx := context.Background()
	_, _ = s.pool.Exec(ctx,
		`DELETE FROM finance_accounting_periods WHERE tenant_id IN ($1, $2)`, s.tenantID, s.tenantB)
	_, _ = s.pool.Exec(ctx,
		`DELETE FROM finance_fiscal_years WHERE tenant_id IN ($1, $2)`, s.tenantID, s.tenantB)
	_, _ = s.pool.Exec(ctx,
		`DELETE FROM tenants WHERE id IN ($1, $2)`, s.tenantID, s.tenantB)
	s.store.Close()
}

func (s *PeriodRepoSuite) SetupTest() {
	s.req = require.New(s.T())
}

// ============================================================================
// Seed helpers
// ============================================================================

func (s *PeriodRepoSuite) seedTenant(id uuid.UUID, slug string) error {
	_, err := s.pool.Exec(context.Background(), `
		INSERT INTO tenants (id, name, slug, company_size, country_code, status)
		VALUES ($1, $2, $3, 'SMALL', 'KE', 'ACTIVE')
		ON CONFLICT (id) DO NOTHING`,
		id,
		"Period Tenant "+id.String()[:8],
		slug+"-"+id.String()[:8],
	)
	return err
}

// seedFiscalYear inserts a fiscal year for the given tenant via raw SQL.
func (s *PeriodRepoSuite) seedFiscalYear(tenantID uuid.UUID, year int) uuid.UUID {
	fyID := uuid.New()
	_, err := s.pool.Exec(context.Background(), `
		INSERT INTO finance_fiscal_years
		  (id, tenant_id, name, start_date, end_date, is_closed, is_locked)
		VALUES
		  ($1, $2, $3, $4, $5, false, false)
		ON CONFLICT (id) DO NOTHING`,
		fyID,
		tenantID,
		fmt.Sprintf("FY%d", year),
		time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC),
	)
	s.req.NoError(err, "seedFiscalYear must succeed")
	return fyID
}

func (s *PeriodRepoSuite) newFiscalYear(year int) *domain.FiscalYear {
	return &domain.FiscalYear{
		Name:      fmt.Sprintf("FY%d", year),
		StartDate: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC),
		IsClosed:  false,
		IsLocked:  false,
	}
}

func (s *PeriodRepoSuite) newPeriod(fyID uuid.UUID, month time.Month, year int) *domain.AccountingPeriod {
	start := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, -1)
	return &domain.AccountingPeriod{
		FiscalYearID: fyID,
		PeriodNumber: int(month),
		Name:         start.Format("Jan 2006"),
		StartDate:    start,
		EndDate:      end,
		Status:       domain.PeriodStatusOpen,
	}
}

// ============================================================================
// FiscalYear CRUD
// ============================================================================

func (s *PeriodRepoSuite) TestCreateFiscalYear_SetsIDAndTimestamps() {
	ctx := shared.WithTenantID(s.ctx, s.tenantID)
	year := 2090 + int(uuid.New().ID()%100) // far-future to avoid collisions
	fy := s.newFiscalYear(year)

	err := s.repo.CreateFiscalYear(ctx, fy)
	s.req.NoError(err)
	s.req.NotEqual(uuid.Nil, fy.ID)
	s.req.False(fy.CreatedAt.IsZero())
}

func (s *PeriodRepoSuite) TestGetFiscalYearByID_RoundTrip() {
	ctx := shared.WithTenantID(s.ctx, s.tenantID)
	year := 2190 + int(uuid.New().ID()%100)
	fy := s.newFiscalYear(year)

	s.req.NoError(s.repo.CreateFiscalYear(ctx, fy))

	got, err := s.repo.GetFiscalYearByID(ctx, fy.ID)
	s.req.NoError(err)
	s.req.Equal(fy.ID, got.ID)
	s.req.Equal(fy.Name, got.Name)
	s.req.False(got.IsClosed)
	s.req.False(got.IsLocked)
}

func (s *PeriodRepoSuite) TestGetFiscalYearByID_NotFound() {
	ctx := shared.WithTenantID(s.ctx, s.tenantID)
	_, err := s.repo.GetFiscalYearByID(ctx, uuid.New())
	s.req.ErrorIs(err, domain.ErrPeriodNotFound)
}

func (s *PeriodRepoSuite) TestGetFiscalYearByYear_ReturnsCorrectYear() {
	ctx := shared.WithTenantID(s.ctx, s.tenantID)
	year := 2290 + int(uuid.New().ID()%100)
	fy := s.newFiscalYear(year)
	s.req.NoError(s.repo.CreateFiscalYear(ctx, fy))

	got, err := s.repo.GetFiscalYearByYear(ctx, s.tenantID, year)
	s.req.NoError(err)
	s.req.Equal(fy.ID, got.ID)
	s.req.Equal(year, got.Year)
}

func (s *PeriodRepoSuite) TestGetFiscalYearByYear_NotFound() {
	ctx := shared.WithTenantID(s.ctx, s.tenantID)
	_, err := s.repo.GetFiscalYearByYear(ctx, s.tenantID, 1800)
	s.req.ErrorIs(err, domain.ErrPeriodNotFound)
}

func (s *PeriodRepoSuite) TestUpdateFiscalYear_PersistsChanges() {
	ctx := shared.WithTenantID(s.ctx, s.tenantID)
	year := 2390 + int(uuid.New().ID()%100)
	fy := s.newFiscalYear(year)
	s.req.NoError(s.repo.CreateFiscalYear(ctx, fy))

	fy.IsClosed = true
	fy.IsLocked = false
	updatedBy := uuid.New()
	fy.UpdatedBy = &updatedBy

	s.req.NoError(s.repo.UpdateFiscalYear(ctx, fy))

	got, err := s.repo.GetFiscalYearByID(ctx, fy.ID)
	s.req.NoError(err)
	s.req.True(got.IsClosed)
}

func (s *PeriodRepoSuite) TestListFiscalYears_IncludesCreatedYear() {
	ctx := shared.WithTenantID(s.ctx, s.tenantID)
	year := 2490 + int(uuid.New().ID()%100)
	fy := s.newFiscalYear(year)
	s.req.NoError(s.repo.CreateFiscalYear(ctx, fy))

	list, err := s.repo.ListFiscalYears(ctx, s.tenantID)
	s.req.NoError(err)
	s.req.NotEmpty(list)

	found := false
	for _, item := range list {
		if item.ID == fy.ID {
			found = true
		}
	}
	s.req.True(found, "created fiscal year must appear in list")
}

// ============================================================================
// AccountingPeriod CRUD
// ============================================================================

func (s *PeriodRepoSuite) TestCreatePeriod_SetsIDAndTimestamps() {
	ctx := shared.WithTenantID(s.ctx, s.tenantID)
	fyID := s.seedFiscalYear(s.tenantID, 2590+int(uuid.New().ID()%100))

	p := s.newPeriod(fyID, time.January, 2590)
	err := s.repo.CreatePeriod(ctx, p)
	s.req.NoError(err)
	s.req.NotEqual(uuid.Nil, p.ID)
	s.req.False(p.CreatedAt.IsZero())
}

func (s *PeriodRepoSuite) TestGetPeriodByID_RoundTrip() {
	ctx := shared.WithTenantID(s.ctx, s.tenantID)
	fyID := s.seedFiscalYear(s.tenantID, 2690+int(uuid.New().ID()%100))

	p := s.newPeriod(fyID, time.February, 2690)
	s.req.NoError(s.repo.CreatePeriod(ctx, p))

	got, err := s.repo.GetPeriodByID(ctx, p.ID)
	s.req.NoError(err)
	s.req.Equal(p.ID, got.ID)
	s.req.Equal(p.Name, got.Name)
	s.req.Equal(domain.PeriodStatusOpen, got.Status)
}

func (s *PeriodRepoSuite) TestGetPeriodByID_NotFound() {
	ctx := shared.WithTenantID(s.ctx, s.tenantID)
	_, err := s.repo.GetPeriodByID(ctx, uuid.New())
	s.req.ErrorIs(err, domain.ErrPeriodNotFound)
}

func (s *PeriodRepoSuite) TestGetPeriodForDate_FindsOverlappingPeriod() {
	ctx := shared.WithTenantID(s.ctx, s.tenantID)
	baseYear := 2790 + int(uuid.New().ID()%100)
	fyID := s.seedFiscalYear(s.tenantID, baseYear)

	p := s.newPeriod(fyID, time.March, baseYear)
	s.req.NoError(s.repo.CreatePeriod(ctx, p))

	mid := time.Date(baseYear, time.March, 15, 0, 0, 0, 0, time.UTC)
	got, err := s.repo.GetPeriodForDate(ctx, s.tenantID, mid)
	s.req.NoError(err)
	s.req.Equal(p.ID, got.ID)
}

func (s *PeriodRepoSuite) TestGetPeriodForDate_NoMatch() {
	ctx := shared.WithTenantID(s.ctx, s.tenantID)
	// Far-future date with no period
	_, err := s.repo.GetPeriodForDate(ctx, s.tenantID, time.Date(2999, 6, 15, 0, 0, 0, 0, time.UTC))
	s.req.ErrorIs(err, domain.ErrPeriodNotFound)
}

func (s *PeriodRepoSuite) TestUpdatePeriod_StatusTransition() {
	ctx := shared.WithTenantID(s.ctx, s.tenantID)
	baseYear := 2890 + int(uuid.New().ID()%100)
	fyID := s.seedFiscalYear(s.tenantID, baseYear)

	p := s.newPeriod(fyID, time.April, baseYear)
	s.req.NoError(s.repo.CreatePeriod(ctx, p))

	now := time.Now()
	closedBy := uuid.New()
	p.Status = domain.PeriodStatusSoftClosed
	p.ClosedAt = &now
	p.ClosedBy = &closedBy

	s.req.NoError(s.repo.UpdatePeriod(ctx, p))

	got, err := s.repo.GetPeriodByID(ctx, p.ID)
	s.req.NoError(err)
	s.req.Equal(domain.PeriodStatusSoftClosed, got.Status)
	s.req.NotNil(got.ClosedAt)
	s.req.Equal(closedBy, *got.ClosedBy)
}

func (s *PeriodRepoSuite) TestListPeriods_FiltersByFiscalYear() {
	ctx := shared.WithTenantID(s.ctx, s.tenantID)
	baseYear := 2990 + int(uuid.New().ID()%100)
	fyID := s.seedFiscalYear(s.tenantID, baseYear)

	// Seed 3 periods in this FY.
	for m := time.January; m <= time.March; m++ {
		p := s.newPeriod(fyID, m, baseYear)
		p.PeriodNumber = int(m)
		s.req.NoError(s.repo.CreatePeriod(ctx, p))
	}

	list, err := s.repo.ListPeriods(ctx, s.tenantID, fyID)
	s.req.NoError(err)
	s.req.Len(list, 3)
	// Verify ordered by period_number ASC.
	s.req.True(list[0].PeriodNumber <= list[1].PeriodNumber)
	s.req.True(list[1].PeriodNumber <= list[2].PeriodNumber)
}

// ============================================================================
// Tenant isolation
// ============================================================================

func (s *PeriodRepoSuite) TestTenantIsolation_FiscalYear_CrossTenantRead() {
	ctxA := shared.WithTenantID(s.ctx, s.tenantID)
	ctxB := shared.WithTenantID(s.ctx, s.tenantB)

	year := 3090 + int(uuid.New().ID()%100)
	fyA := s.newFiscalYear(year)
	s.req.NoError(s.repo.CreateFiscalYear(ctxA, fyA))

	// Tenant B cannot read tenant A's fiscal year.
	_, err := s.repo.GetFiscalYearByID(ctxB, fyA.ID)
	s.req.ErrorIs(err, domain.ErrPeriodNotFound,
		"cross-tenant fiscal year read must be blocked by RLS")
}

func (s *PeriodRepoSuite) TestTenantIsolation_Period_CrossTenantRead() {
	ctxA := shared.WithTenantID(s.ctx, s.tenantID)
	ctxB := shared.WithTenantID(s.ctx, s.tenantB)

	baseYear := 3190 + int(uuid.New().ID()%100)
	fyID := s.seedFiscalYear(s.tenantID, baseYear)

	p := s.newPeriod(fyID, time.May, baseYear)
	s.req.NoError(s.repo.CreatePeriod(ctxA, p))

	// Tenant B cannot read tenant A's period.
	_, err := s.repo.GetPeriodByID(ctxB, p.ID)
	s.req.ErrorIs(err, domain.ErrPeriodNotFound,
		"cross-tenant period read must be blocked by RLS")
}

// ============================================================================
// DB constraint tests
// ============================================================================

// TestDBConstraint_Period_InvalidStatus verifies DB rejects unknown status.
func (s *PeriodRepoSuite) TestDBConstraint_Period_InvalidStatus() {
	_, err := s.pool.Exec(context.Background(),
		`SELECT set_tenant_context($1)`, s.tenantID)
	s.req.NoError(err)

	baseYear := 3290 + int(uuid.New().ID()%100)
	fyID := s.seedFiscalYear(s.tenantID, baseYear)

	_, err = s.pool.Exec(context.Background(), `
		INSERT INTO finance_accounting_periods
		  (tenant_id, fiscal_year_id, name, period_number, start_date, end_date, status)
		VALUES
		  (current_tenant_id(), $1, 'Bad Period', 99, $2, $3, 'BOGUS_STATUS')`,
		fyID,
		time.Date(baseYear, 6, 1, 0, 0, 0, 0, time.UTC),
		time.Date(baseYear, 6, 30, 0, 0, 0, 0, time.UTC),
	)
	s.req.Error(err, "DB must reject invalid period status via CHECK constraint")
}

// TestDBConstraint_FiscalYear_MissingContext verifies tenant context is required.
func (s *PeriodRepoSuite) TestDBConstraint_FiscalYear_MissingContext() {
	fy := s.newFiscalYear(1900)
	err := s.repo.CreateFiscalYear(context.Background(), fy)
	s.req.Error(err)
	s.req.Contains(err.Error(), "tenant ID not found")
}
