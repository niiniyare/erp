//go:build database
// +build database

package repository_test

// Integration tests for BudgetRepository.
//
// These tests require a running PostgreSQL instance with the full schema applied.
// Set DB_URL to run them, e.g.:
//
//	DB_URL="postgres://user:pass@localhost:5432/erp_test?sslmode=disable" go test ./internal/core/finance/repository/...
//
// The repository manages its own transactions via WithTenant internally.
// Tests call repo methods directly on top of the shared s.ctx (which carries the tenantID).

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	db "awo.so/db/sqlc"
	"awo.so/internal/core/finance/domain"
	"awo.so/internal/core/finance/repository"
	"awo.so/internal/shared"
	"awo.so/internal/shared/tracing"
)

// ============================================================================
// Suite setup
// ============================================================================

type BudgetRepoSuite struct {
	suite.Suite
	req      *require.Assertions
	pool     *pgxpool.Pool
	store    db.Store
	repo     domain.BudgetRepository
	tenantID uuid.UUID
	fyID     uuid.UUID
	ctx      context.Context
}

func TestBudgetRepoSuite(t *testing.T) {
	suite.Run(t, new(BudgetRepoSuite))
}

// SetupSuite connects to the test DB and creates seed data once for the run.
func (s *BudgetRepoSuite) SetupSuite() {
	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		s.T().Skip("DB_URL not set — skipping repository integration tests")
	}

	store, err := db.NewDB(dsn)
	if err != nil {
		s.T().Skipf("cannot connect to test DB: %v", err)
	}
	s.store = store
	s.pool = store.GetPool()

	s.repo = repository.NewBudgetRepository(store, tracing.NewNoOpService())

	// Create a stable test tenant so set_tenant_context() succeeds.
	s.tenantID = uuid.New()
	s.fyID = uuid.New()
	s.ctx = shared.WithTenantID(context.Background(), s.tenantID)

	if err := s.seedTenant(); err != nil {
		s.T().Skipf("cannot seed test tenant: %v", err)
	}
	if err := s.seedFiscalYear(); err != nil {
		s.T().Skipf("cannot seed fiscal year: %v", err)
	}
}

// TearDownSuite removes all test data so the suite can be re-run safely.
func (s *BudgetRepoSuite) TearDownSuite() {
	if s.pool == nil {
		return
	}
	ctx := context.Background()

	// Remove budget line items and budgets created under the test tenant.
	_, _ = s.pool.Exec(ctx,
		`DELETE FROM finance_budget_line_items WHERE tenant_id = $1`, s.tenantID)
	_, _ = s.pool.Exec(ctx,
		`DELETE FROM finance_budgets WHERE tenant_id = $1`, s.tenantID)
	_, _ = s.pool.Exec(ctx,
		`DELETE FROM finance_fiscal_years WHERE tenant_id = $1`, s.tenantID)
	_, _ = s.pool.Exec(ctx,
		`DELETE FROM tenants WHERE id = $1`, s.tenantID)

	s.store.Close()
}

// SetupTest ensures each test uses a fresh Assertions instance.
func (s *BudgetRepoSuite) SetupTest() {
	s.req = require.New(s.T())
}

// ============================================================================
// Seed helpers
// ============================================================================

// seedTenant inserts a minimal ACTIVE tenant so the RLS stored-proc accepts it.
func (s *BudgetRepoSuite) seedTenant() error {
	_, err := s.pool.Exec(context.Background(), `
		INSERT INTO tenants
		  (id, name, slug, company_size, country_code, status)
		VALUES
		  ($1, $2, $3, 'SMALL', 'KE', 'ACTIVE')
		ON CONFLICT (id) DO NOTHING`,
		s.tenantID,
		"Test Tenant "+s.tenantID.String()[:8],
		"test-"+s.tenantID.String()[:8],
	)
	return err
}

// seedFiscalYear inserts a fiscal year so budget FK constraints are satisfied.
func (s *BudgetRepoSuite) seedFiscalYear() error {
	year := time.Now().Year()
	_, err := s.pool.Exec(context.Background(), `
		INSERT INTO finance_fiscal_years
		  (id, tenant_id, name, start_date, end_date, status)
		VALUES
		  ($1, $2, $3, $4, $5, 'open')
		ON CONFLICT (id) DO NOTHING`,
		s.fyID,
		s.tenantID,
		fmt.Sprintf("FY%d", year),
		time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC),
	)
	return err
}

// newBudget returns a valid, unsaved budget header tied to the suite's tenant.
func (s *BudgetRepoSuite) newBudget() *domain.Budget {
	return &domain.Budget{
		TenantID:     s.tenantID,
		FiscalYearID: s.fyID,
		Name:         "Test Budget " + uuid.New().String()[:8],
		BudgetType:   domain.BudgetTypeAnnual,
		Status:       domain.BudgetStatusDraft,
		CurrencyCode: "KES",
		Version:      1,
		CreatedBy:    uuid.New(),
	}
}

// newLineItem returns a valid BudgetLineItem for the given budget.
func (s *BudgetRepoSuite) newLineItem(budgetID uuid.UUID) *domain.BudgetLineItem {
	return &domain.BudgetLineItem{
		BudgetID:       budgetID,
		TenantID:       s.tenantID,
		AccountID:      uuid.New(),
		BudgetedAmount: decimal.NewFromInt(100_000),
		CreatedBy:      uuid.New(),
	}
}

// ============================================================================
// CreateBudget
// ============================================================================

func (s *BudgetRepoSuite) TestCreateBudget_SetsIDAndTimestamps() {
	b := s.newBudget()
	s.req.Equal(uuid.Nil, b.ID, "ID should be nil before insert")

	err := s.repo.CreateBudget(s.ctx, b)
	s.req.NoError(err)
	s.req.NotEqual(uuid.Nil, b.ID, "DB should assign an ID")
	s.req.False(b.CreatedAt.IsZero(), "CreatedAt should be set")
}

func (s *BudgetRepoSuite) TestCreateBudget_ThenGetByID_RoundTrips() {
	b := s.newBudget()

	err := s.repo.CreateBudget(s.ctx, b)
	s.req.NoError(err)

	fetched, err := s.repo.GetBudgetByID(s.ctx, b.ID)
	s.req.NoError(err)
	s.req.Equal(b.ID, fetched.ID)
	s.req.Equal(b.Name, fetched.Name)
	s.req.Equal(domain.BudgetStatusDraft, fetched.Status)
	s.req.Equal(domain.BudgetTypeAnnual, fetched.BudgetType)
	s.req.Equal("KES", fetched.CurrencyCode)
}

// ============================================================================
// DB Constraint: budget_type CHECK
// ============================================================================

func (s *BudgetRepoSuite) TestCreateBudget_InvalidBudgetType_DBConstraintViolation() {
	// The DB should enforce a CHECK constraint on the budget_type column.
	// Attempting to insert an invalid type must return an error.
	_, err := s.pool.Exec(context.Background(), `SELECT set_tenant_context($1)`, s.tenantID)
	s.req.NoError(err, "set_tenant_context must succeed")

	_, err = s.pool.Exec(context.Background(), `
		INSERT INTO finance_budgets
		  (tenant_id, fiscal_year_id, name, budget_type, status, currency_code, version, created_by)
		VALUES
		  (current_tenant_id(), $1, $2, $3, $4, $5, 1, $6)`,
		s.fyID,
		"Invalid Type Budget",
		"ROLLING", // not a valid BudgetType
		"DRAFT",
		"KES",
		uuid.New(),
	)
	s.req.Error(err, "DB should reject invalid budget_type via CHECK constraint")
}

// ============================================================================
// DB Constraint: status CHECK
// ============================================================================

func (s *BudgetRepoSuite) TestCreateBudget_InvalidStatus_DBConstraintViolation() {
	_, err := s.pool.Exec(context.Background(), `SELECT set_tenant_context($1)`, s.tenantID)
	s.req.NoError(err)

	_, err = s.pool.Exec(context.Background(), `
		INSERT INTO finance_budgets
		  (tenant_id, fiscal_year_id, name, budget_type, status, currency_code, version, created_by)
		VALUES
		  (current_tenant_id(), $1, $2, $3, $4, $5, 1, $6)`,
		s.fyID,
		"Invalid Status Budget",
		"ANNUAL",
		"LIMBO", // not a valid BudgetStatus
		"KES",
		uuid.New(),
	)
	s.req.Error(err, "DB should reject invalid status via CHECK constraint")
}

// ============================================================================
// DB Constraint: budgeted_amount >= 0
// ============================================================================

func (s *BudgetRepoSuite) TestCreateLineItem_NegativeAmount_DBConstraintViolation() {
	// Create a valid budget first.
	b := s.newBudget()
	err := s.repo.CreateBudget(s.ctx, b)
	s.req.NoError(err)

	// Now try to insert a line item with a negative amount directly via SQL.
	_, err = s.pool.Exec(context.Background(), `SELECT set_tenant_context($1)`, s.tenantID)
	s.req.NoError(err)

	_, err = s.pool.Exec(context.Background(), `
		INSERT INTO finance_budget_line_items
		  (budget_id, tenant_id, account_id, budgeted_amount, created_by)
		VALUES
		  ($1, current_tenant_id(), $2, $3, $4)`,
		b.ID,
		uuid.New(),
		"-500.00", // negative — should be rejected
		uuid.New(),
	)
	s.req.Error(err, "DB should reject negative budgeted_amount via CHECK constraint")
}

// ============================================================================
// GetBudgetByID — not found
// ============================================================================

func (s *BudgetRepoSuite) TestGetBudgetByID_NonExistentID_ErrBudgetNotFound() {
	_, err := s.repo.GetBudgetByID(s.ctx, uuid.New())
	s.req.ErrorIs(err, domain.ErrBudgetNotFound)
}

// ============================================================================
// Tenant isolation
// ============================================================================

func (s *BudgetRepoSuite) TestGetBudgetByID_DifferentTenant_CannotSeeOthersBudget() {
	// Insert a budget under our test tenant.
	b := s.newBudget()
	err := s.repo.CreateBudget(s.ctx, b)
	s.req.NoError(err)

	// Create a second, unrelated tenant and confirm it cannot read the first's budget.
	otherTenantID := uuid.New()
	_, err = s.pool.Exec(context.Background(), `
		INSERT INTO tenants (id, name, slug, company_size, country_code, status)
		VALUES ($1, $2, $3, 'SMALL', 'KE', 'ACTIVE')`,
		otherTenantID,
		"Other Tenant "+otherTenantID.String()[:8],
		"other-"+otherTenantID.String()[:8],
	)
	s.req.NoError(err)
	defer func() {
		_, _ = s.pool.Exec(context.Background(),
			`DELETE FROM tenants WHERE id = $1`, otherTenantID)
	}()

	otherCtx := shared.WithTenantID(context.Background(), otherTenantID)
	otherRepo := repository.NewBudgetRepository(s.store, tracing.NewNoOpService())
	_, fetchErr := otherRepo.GetBudgetByID(otherCtx, b.ID)
	s.req.ErrorIs(fetchErr, domain.ErrBudgetNotFound,
		"other tenant must not be able to read our budget")
}

// ============================================================================
// UpdateBudget — status transitions
// ============================================================================

func (s *BudgetRepoSuite) TestUpdateBudget_SubmitTransition_PersistsTimestamps() {
	b := s.newBudget()
	err := s.repo.CreateBudget(s.ctx, b)
	s.req.NoError(err)

	userID := uuid.New()
	now := time.Now()
	b.Status = domain.BudgetStatusSubmitted
	b.SubmittedAt = &now
	b.SubmittedBy = &userID
	b.UpdatedBy = &userID

	err = s.repo.UpdateBudget(s.ctx, b)
	s.req.NoError(err)

	fetched, err := s.repo.GetBudgetByID(s.ctx, b.ID)
	s.req.NoError(err)
	s.req.Equal(domain.BudgetStatusSubmitted, fetched.Status)
	s.req.NotNil(fetched.SubmittedAt)
	s.req.Equal(userID, *fetched.SubmittedBy)
}

func (s *BudgetRepoSuite) TestUpdateBudget_ApproveTransition_PersistsApprover() {
	b := s.newBudget()
	b.Status = domain.BudgetStatusSubmitted
	err := s.repo.CreateBudget(s.ctx, b)
	s.req.NoError(err)

	approverID := uuid.New()
	now := time.Now()
	b.Status = domain.BudgetStatusApproved
	b.ApprovedAt = &now
	b.ApprovedBy = &approverID
	b.UpdatedBy = &approverID

	err = s.repo.UpdateBudget(s.ctx, b)
	s.req.NoError(err)

	fetched, err := s.repo.GetBudgetByID(s.ctx, b.ID)
	s.req.NoError(err)
	s.req.Equal(domain.BudgetStatusApproved, fetched.Status)
	s.req.Equal(approverID, *fetched.ApprovedBy)
}

// ============================================================================
// ListBudgets
// ============================================================================

func (s *BudgetRepoSuite) TestListBudgets_FilterByFiscalYear_OnlyReturnsThatYear() {
	otherFyID := uuid.New()
	_, err := s.pool.Exec(context.Background(), `
		INSERT INTO finance_fiscal_years
		  (id, tenant_id, name, start_date, end_date, status)
		VALUES
		  ($1, $2, 'Other FY', $3, $4, 'open')`,
		otherFyID, s.tenantID,
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
	)
	s.req.NoError(err)
	defer func() {
		_, _ = s.pool.Exec(context.Background(),
			`DELETE FROM finance_fiscal_years WHERE id = $1`, otherFyID)
	}()

	// Insert one budget under main fyID, one under otherFyID.
	b1 := s.newBudget()
	b2 := s.newBudget()
	b2.FiscalYearID = otherFyID

	err = s.repo.CreateBudget(s.ctx, b1)
	s.req.NoError(err)
	err = s.repo.CreateBudget(s.ctx, b2)
	s.req.NoError(err)

	listed, err := s.repo.ListBudgets(s.ctx, s.tenantID, &s.fyID)
	s.req.NoError(err)

	ids := make(map[uuid.UUID]bool)
	for _, b := range listed {
		ids[b.ID] = true
	}
	s.req.True(ids[b1.ID], "b1 should be in the list for main fiscal year")
	s.req.False(ids[b2.ID], "b2 should NOT be in the list for main fiscal year")
}

// ============================================================================
// CreateLineItems / GetLineItems
// ============================================================================

func (s *BudgetRepoSuite) TestCreateAndGetLineItems_RoundTrip() {
	b := s.newBudget()
	err := s.repo.CreateBudget(s.ctx, b)
	s.req.NoError(err)

	lines := []*domain.BudgetLineItem{
		s.newLineItem(b.ID),
		s.newLineItem(b.ID),
	}
	lines[0].BudgetedAmount = decimal.NewFromInt(50_000)
	lines[1].BudgetedAmount = decimal.NewFromInt(30_000)

	err = s.repo.CreateLineItems(s.ctx, lines)
	s.req.NoError(err)
	s.req.NotEqual(uuid.Nil, lines[0].ID, "line item IDs must be set by DB")
	s.req.NotEqual(uuid.Nil, lines[1].ID)

	fetched, err := s.repo.GetLineItems(s.ctx, b.ID)
	s.req.NoError(err)
	s.req.Len(fetched, 2)
}

func (s *BudgetRepoSuite) TestGetLineItems_ComputesVarianceAmountAndPct() {
	b := s.newBudget()
	err := s.repo.CreateBudget(s.ctx, b)
	s.req.NoError(err)

	li := s.newLineItem(b.ID)
	li.BudgetedAmount = decimal.NewFromInt(100_000)
	// actual_amount defaults to 0 in DB; variance should be 100_000

	err = s.repo.CreateLineItems(s.ctx, []*domain.BudgetLineItem{li})
	s.req.NoError(err)

	fetched, err := s.repo.GetLineItems(s.ctx, b.ID)
	s.req.NoError(err)
	s.req.Len(fetched, 1)
	// VarianceAmount = BudgetedAmount - ActualAmount = 100_000 - 0 = 100_000
	s.req.True(fetched[0].VarianceAmount.Equal(decimal.NewFromInt(100_000)),
		"VarianceAmount should be 100_000 when actual=0")
	// VariancePct = 100_000 / 100_000 * 100 = 100
	s.req.True(fetched[0].VariancePct.Equal(decimal.NewFromInt(100)),
		"VariancePct should be 100%% when actual=0")
}

// ============================================================================
// DeleteLineItems
// ============================================================================

func (s *BudgetRepoSuite) TestDeleteLineItems_RemovesAllLinesForBudget() {
	b := s.newBudget()
	err := s.repo.CreateBudget(s.ctx, b)
	s.req.NoError(err)

	lines := []*domain.BudgetLineItem{s.newLineItem(b.ID), s.newLineItem(b.ID)}
	err = s.repo.CreateLineItems(s.ctx, lines)
	s.req.NoError(err)

	err = s.repo.DeleteLineItems(s.ctx, b.ID)
	s.req.NoError(err)

	remaining, err := s.repo.GetLineItems(s.ctx, b.ID)
	s.req.NoError(err)
	s.req.Empty(remaining)
}

// ============================================================================
// DB: double-entry transaction constraints (raw SQL)
// ============================================================================
// These tests exercise CHECK constraints that exist on the
// finance_transactions and finance_transaction_entries tables.
// They bypass the service layer intentionally to verify the DB itself
// enforces double-entry integrity — not just the application.

func (s *BudgetRepoSuite) insertTestTransaction(txnNumber string) (uuid.UUID, error) {
	txnID := uuid.New()
	_, err := s.pool.Exec(context.Background(), `
		INSERT INTO finance_transactions
		  (id, tenant_id, transaction_number, transaction_type, transaction_status,
		   transaction_date, description, currency_code, exchange_rate,
		   total_debit_amount, total_credit_amount, approval_status, created_by)
		VALUES
		  ($1, current_tenant_id(), $2, 'MANUAL', 'DRAFT',
		   NOW(), 'Test transaction', 'KES', 1.0, 1000, 1000, 'NOT_REQUIRED', $3)`,
		txnID, txnNumber, uuid.New(),
	)
	return txnID, err
}

func (s *BudgetRepoSuite) TestTransactionEntry_BothAmountsNonZero_DBRejects() {
	_, err := s.pool.Exec(context.Background(), `SELECT set_tenant_context($1)`, s.tenantID)
	s.req.NoError(err)

	txnID, err := s.insertTestTransaction("TXN-DUAL-" + uuid.New().String()[:8])
	if err != nil {
		s.T().Logf("skipping double-entry constraint test: cannot insert parent transaction: %v", err)
		s.T().Skip()
	}
	defer func() {
		_, _ = s.pool.Exec(context.Background(),
			`DELETE FROM finance_transactions WHERE id = $1`, txnID)
	}()

	// Both debit AND credit are non-zero — DUAL_AMOUNTS violation.
	_, err = s.pool.Exec(context.Background(), `
		INSERT INTO finance_transaction_entries
		  (tenant_id, transaction_id, entry_number, account_id,
		   debit_amount, credit_amount, description, exchange_rate)
		VALUES
		  (current_tenant_id(), $1, 1, $2, 500.00, 500.00, 'Bad entry', 1.0)`,
		txnID, uuid.New(),
	)
	s.req.Error(err,
		"DB must reject entry with both debit_amount > 0 AND credit_amount > 0")
}

func (s *BudgetRepoSuite) TestTransactionEntry_BothAmountsZero_DBRejects() {
	_, err := s.pool.Exec(context.Background(), `SELECT set_tenant_context($1)`, s.tenantID)
	s.req.NoError(err)

	txnID, err := s.insertTestTransaction("TXN-ZERO-" + uuid.New().String()[:8])
	if err != nil {
		s.T().Logf("skipping: cannot insert parent transaction: %v", err)
		s.T().Skip()
	}
	defer func() {
		_, _ = s.pool.Exec(context.Background(),
			`DELETE FROM finance_transactions WHERE id = $1`, txnID)
	}()

	// Both amounts = 0 → MISSING_AMOUNT violation.
	_, err = s.pool.Exec(context.Background(), `
		INSERT INTO finance_transaction_entries
		  (tenant_id, transaction_id, entry_number, account_id,
		   debit_amount, credit_amount, description, exchange_rate)
		VALUES
		  (current_tenant_id(), $1, 1, $2, 0.00, 0.00, 'Zero entry', 1.0)`,
		txnID, uuid.New(),
	)
	s.req.Error(err,
		"DB must reject entry where both debit_amount and credit_amount are zero")
}

func (s *BudgetRepoSuite) TestTransactionEntry_NegativeDebitAmount_DBRejects() {
	_, err := s.pool.Exec(context.Background(), `SELECT set_tenant_context($1)`, s.tenantID)
	s.req.NoError(err)

	txnID, err := s.insertTestTransaction("TXN-NEGDB-" + uuid.New().String()[:8])
	if err != nil {
		s.T().Logf("skipping: cannot insert parent transaction: %v", err)
		s.T().Skip()
	}
	defer func() {
		_, _ = s.pool.Exec(context.Background(),
			`DELETE FROM finance_transactions WHERE id = $1`, txnID)
	}()

	_, err = s.pool.Exec(context.Background(), `
		INSERT INTO finance_transaction_entries
		  (tenant_id, transaction_id, entry_number, account_id,
		   debit_amount, credit_amount, description, exchange_rate)
		VALUES
		  (current_tenant_id(), $1, 1, $2, -100.00, 0.00, 'Neg debit', 1.0)`,
		txnID, uuid.New(),
	)
	s.req.Error(err, "DB must reject negative debit_amount (debit_amount >= 0 CHECK)")
}

func (s *BudgetRepoSuite) TestTransactionEntry_NegativeCreditAmount_DBRejects() {
	_, err := s.pool.Exec(context.Background(), `SELECT set_tenant_context($1)`, s.tenantID)
	s.req.NoError(err)

	txnID, err := s.insertTestTransaction("TXN-NEGCR-" + uuid.New().String()[:8])
	if err != nil {
		s.T().Logf("skipping: cannot insert parent transaction: %v", err)
		s.T().Skip()
	}
	defer func() {
		_, _ = s.pool.Exec(context.Background(),
			`DELETE FROM finance_transactions WHERE id = $1`, txnID)
	}()

	_, err = s.pool.Exec(context.Background(), `
		INSERT INTO finance_transaction_entries
		  (tenant_id, transaction_id, entry_number, account_id,
		   debit_amount, credit_amount, description, exchange_rate)
		VALUES
		  (current_tenant_id(), $1, 1, $2, 0.00, -100.00, 'Neg credit', 1.0)`,
		txnID, uuid.New(),
	)
	s.req.Error(err, "DB must reject negative credit_amount (credit_amount >= 0 CHECK)")
}

// ============================================================================
// DB: period status CHECK constraint
// ============================================================================

func (s *BudgetRepoSuite) TestAccountingPeriod_InvalidStatus_DBRejects() {
	_, err := s.pool.Exec(context.Background(), `SELECT set_tenant_context($1)`, s.tenantID)
	s.req.NoError(err)

	_, err = s.pool.Exec(context.Background(), `
		INSERT INTO finance_accounting_periods
		  (tenant_id, fiscal_year_id, name, period_number, start_date, end_date, status)
		VALUES
		  (current_tenant_id(), $1, 'Jan 2026', 1, '2026-01-01', '2026-01-31', $2)`,
		s.fyID,
		"FAKE_STATUS",
	)
	s.req.Error(err, "DB must reject invalid period status via CHECK constraint")
}
