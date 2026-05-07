//go:build database
// +build database

// Package repository_test contains real-PostgreSQL integration tests for the
// finance transaction repository.
//
// Run with:
//
//	DB_URL="postgres://user:pass@localhost:5432/erp_test?sslmode=disable" \
//	  go test -tags database ./internal/core/finance/repository/... -v
package repository_test

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

type TransactionRepoSuite struct {
	suite.Suite
	req      *require.Assertions
	pool     *pgxpool.Pool
	store    db.Store
	repo     domain.TransactionRepository
	tenantA  uuid.UUID
	tenantB  uuid.UUID
	userID   uuid.UUID // valid user ID for created_by FK
	ctx      context.Context
}

func TestTransactionRepoSuite(t *testing.T) {
	suite.Run(t, new(TransactionRepoSuite))
}

func (s *TransactionRepoSuite) SetupSuite() {
	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		s.T().Skip("DB_URL not set — skipping transaction repository integration tests")
	}

	store, err := db.NewDB(dsn)
	if err != nil {
		s.T().Skipf("cannot connect to test DB: %v", err)
	}
	s.store = store
	s.pool = store.GetPool()

	s.repo = repository.NewTransactionRepository(store, tracing.NewNoOpService(), nil, nil)

	s.tenantA = uuid.New()
	s.tenantB = uuid.New()
	s.userID = uuid.New()
	s.ctx = context.Background()

	if err := s.seedTenant(s.tenantA, "txn-tenant-a"); err != nil {
		s.T().Skipf("cannot seed tenant A: %v", err)
	}
	if err := s.seedTenant(s.tenantB, "txn-tenant-b"); err != nil {
		s.T().Skipf("cannot seed tenant B: %v", err)
	}
	if err := s.seedUser(); err != nil {
		// Not fatal — if no FK constraint, random UUID works fine.
		s.T().Logf("note: could not seed test user (%v); using random UUID for created_by", err)
	}
}

func (s *TransactionRepoSuite) TearDownSuite() {
	if s.pool == nil {
		return
	}
	ctx := context.Background()
	_, _ = s.pool.Exec(ctx,
		`DELETE FROM finance_transactions WHERE tenant_id IN ($1, $2)`, s.tenantA, s.tenantB)
	_, _ = s.pool.Exec(ctx,
		`DELETE FROM tenants WHERE id IN ($1, $2)`, s.tenantA, s.tenantB)
	s.store.Close()
}

func (s *TransactionRepoSuite) SetupTest() {
	s.req = require.New(s.T())
}

// ============================================================================
// Seed helpers
// ============================================================================

func (s *TransactionRepoSuite) seedTenant(id uuid.UUID, slug string) error {
	_, err := s.pool.Exec(context.Background(), `
		INSERT INTO tenants (id, name, slug, company_size, country_code, status)
		VALUES ($1, $2, $3, 'SMALL', 'KE', 'ACTIVE')
		ON CONFLICT (id) DO NOTHING`,
		id,
		"Txn Tenant "+id.String()[:8],
		slug+"-"+id.String()[:8],
	)
	return err
}

// seedUser inserts a minimal user row so that created_by FK constraints are satisfied.
// If the schema has no such FK, this is a no-op and the random s.userID is used.
func (s *TransactionRepoSuite) seedUser() error {
	// Try to insert a user; ignore if table structure differs.
	entityID := uuid.New()
	_, err := s.pool.Exec(context.Background(), `
		INSERT INTO entities (uuid, name, type, is_active, hidden, accrual_method, fy_start_month,
		                      address, settings, metadata, tenant_id)
		VALUES ($1, $2, 'COMPANY', true, false, true, 1, '{}', '{}', '{}', $3)
		ON CONFLICT (uuid) DO NOTHING`,
		entityID,
		"Test Entity "+entityID.String()[:8],
		s.tenantA,
	)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(context.Background(), `
		INSERT INTO users (id, entity_id, username, email, user_type, user_attributes, settings, tenant_id)
		VALUES ($1, $2, $3, $4, 'INTERNAL', '{}', '{}', $5)
		ON CONFLICT (id) DO NOTHING`,
		s.userID,
		entityID,
		"testuser-"+s.userID.String()[:8],
		"test-txn-"+s.userID.String()[:8]+"@example.com",
		s.tenantA,
	)
	return err
}

func (s *TransactionRepoSuite) newTransaction(suffix string) *domain.Transaction {
	ref := "REF-" + suffix
	return &domain.Transaction{
		ID:                uuid.New(),
		TransactionDate:   time.Now(),
		TransactionType:   domain.TransactionTypeManual,
		TransactionStatus: domain.TransactionStatusDraft,
		TransactionNumber: "TXN-" + suffix,
		ReferenceNumber:   &ref,
		Description:       "Test transaction " + suffix,
		CurrencyCode:      "USD",
		ExchangeRate:      decimal.NewFromFloat(1.0),
		IsRecurring:       false,
		CreatedBy:         s.userID,
	}
}

// ============================================================================
// CRUD tests
// ============================================================================

// TestCreate_SetsTimestamps verifies DB populates timestamps on insert.
func (s *TransactionRepoSuite) TestCreate_SetsTimestamps() {
	ctx := shared.WithTenantID(s.ctx, s.tenantA)
	txn := s.newTransaction(uuid.New().String()[:8])

	err := s.repo.Create(ctx, txn)
	s.req.NoError(err)
	s.req.False(txn.CreatedAt.IsZero(), "CreatedAt must be set by DB")
	s.req.False(txn.UpdatedAt.IsZero(), "UpdatedAt must be set by DB")
}

// TestGetByID_RoundTrip verifies all fields persisted correctly.
func (s *TransactionRepoSuite) TestGetByID_RoundTrip() {
	ctx := shared.WithTenantID(s.ctx, s.tenantA)
	txn := s.newTransaction(uuid.New().String()[:8])

	s.req.NoError(s.repo.Create(ctx, txn))

	got, err := s.repo.GetByID(ctx, txn.ID)
	s.req.NoError(err)
	s.req.Equal(txn.ID, got.ID)
	s.req.Equal(txn.Description, got.Description)
	s.req.Equal(txn.TransactionType, got.TransactionType)
	s.req.Equal(txn.TransactionStatus, got.TransactionStatus)
	s.req.Equal(txn.CurrencyCode, got.CurrencyCode)
	s.req.True(txn.ExchangeRate.Equal(got.ExchangeRate))
}

// TestGetByID_NotFound returns ErrTransactionNotFound for unknown ID.
func (s *TransactionRepoSuite) TestGetByID_NotFound() {
	ctx := shared.WithTenantID(s.ctx, s.tenantA)
	_, err := s.repo.GetByID(ctx, uuid.New())
	s.req.ErrorIs(err, domain.ErrTransactionNotFound)
}

// TestUpdate_PersistsDescriptionChange verifies update is durable.
func (s *TransactionRepoSuite) TestUpdate_PersistsDescriptionChange() {
	ctx := shared.WithTenantID(s.ctx, s.tenantA)
	txn := s.newTransaction(uuid.New().String()[:8])
	s.req.NoError(s.repo.Create(ctx, txn))

	txn.Description = "updated description"
	s.req.NoError(s.repo.Update(ctx, txn))

	got, err := s.repo.GetByID(ctx, txn.ID)
	s.req.NoError(err)
	s.req.Equal("updated description", got.Description)
	s.req.NotEqual(got.CreatedAt, got.UpdatedAt, "UpdatedAt should advance after update")
}

// TestDelete_SoftDelete verifies soft-deleted row is invisible to GetByID.
func (s *TransactionRepoSuite) TestDelete_SoftDelete() {
	ctx := shared.WithTenantID(s.ctx, s.tenantA)
	txn := s.newTransaction(uuid.New().String()[:8])
	s.req.NoError(s.repo.Create(ctx, txn))

	s.req.NoError(s.repo.Delete(ctx, txn.ID))

	_, err := s.repo.GetByID(ctx, txn.ID)
	s.req.ErrorIs(err, domain.ErrTransactionNotFound, "soft-deleted row must not be returned")
}

// TestDelete_NonExistent returns ErrTransactionNotFound for unknown ID.
func (s *TransactionRepoSuite) TestDelete_NonExistent() {
	ctx := shared.WithTenantID(s.ctx, s.tenantA)
	err := s.repo.Delete(ctx, uuid.New())
	s.req.ErrorIs(err, domain.ErrTransactionNotFound)
}

// ============================================================================
// Tenant isolation tests (FIN-REPO-TXN-ISO)
// ============================================================================

// TestTenantIsolation_CrossTenantRead verifies tenant A cannot read tenant B's data.
func (s *TransactionRepoSuite) TestTenantIsolation_CrossTenantRead() {
	ctxA := shared.WithTenantID(s.ctx, s.tenantA)
	ctxB := shared.WithTenantID(s.ctx, s.tenantB)

	txnA := s.newTransaction("ISO-A-" + uuid.New().String()[:6])
	txnB := s.newTransaction("ISO-B-" + uuid.New().String()[:6])

	s.req.NoError(s.repo.Create(ctxA, txnA))
	s.req.NoError(s.repo.Create(ctxB, txnB))

	// A can see its own data.
	got, err := s.repo.GetByID(ctxA, txnA.ID)
	s.req.NoError(err)
	s.req.Equal(txnA.ID, got.ID)

	// A cannot see B's data.
	_, err = s.repo.GetByID(ctxA, txnB.ID)
	s.req.ErrorIs(err, domain.ErrTransactionNotFound,
		"tenant A must not read tenant B's transactions")

	// B cannot see A's data.
	_, err = s.repo.GetByID(ctxB, txnA.ID)
	s.req.ErrorIs(err, domain.ErrTransactionNotFound,
		"tenant B must not read tenant A's transactions")
}

// TestTenantIsolation_SameNumberDifferentTenants verifies same transaction number
// can exist across tenants (uniqueness is per-tenant).
func (s *TransactionRepoSuite) TestTenantIsolation_SameNumberDifferentTenants() {
	sharedNumber := "TXN-SHARED-" + uuid.New().String()[:6]
	refA := "REF-A"
	refB := "REF-B"

	txnA := &domain.Transaction{
		ID: uuid.New(), TransactionDate: time.Now(),
		TransactionType: domain.TransactionTypeManual, TransactionStatus: domain.TransactionStatusDraft,
		TransactionNumber: sharedNumber, ReferenceNumber: &refA,
		Description: "Tenant A shared number", CurrencyCode: "USD",
		ExchangeRate: decimal.NewFromFloat(1.0), CreatedBy: s.userID,
	}
	txnB := &domain.Transaction{
		ID: uuid.New(), TransactionDate: time.Now(),
		TransactionType: domain.TransactionTypeManual, TransactionStatus: domain.TransactionStatusDraft,
		TransactionNumber: sharedNumber, ReferenceNumber: &refB,
		Description: "Tenant B shared number", CurrencyCode: "USD",
		ExchangeRate: decimal.NewFromFloat(1.0), CreatedBy: s.userID,
	}

	ctxA := shared.WithTenantID(s.ctx, s.tenantA)
	ctxB := shared.WithTenantID(s.ctx, s.tenantB)

	s.req.NoError(s.repo.Create(ctxA, txnA),
		"same transaction number must be allowed in different tenants")
	s.req.NoError(s.repo.Create(ctxB, txnB),
		"same transaction number must be allowed in different tenants")
}

// TestTenantIsolation_MissingContext rejects operation with no tenant in context.
func (s *TransactionRepoSuite) TestTenantIsolation_MissingContext() {
	txn := s.newTransaction("NO-TENANT")
	err := s.repo.Create(context.Background(), txn)
	s.req.Error(err)
	s.req.Contains(err.Error(), "tenant ID not found")
}

// ============================================================================
// List / filter tests
// ============================================================================

// TestList_ByStatus verifies status filter returns only matching rows.
func (s *TransactionRepoSuite) TestList_ByStatus() {
	ctx := shared.WithTenantID(s.ctx, s.tenantA)
	suffix := uuid.New().String()[:6]

	txn := s.newTransaction("LIST-STATUS-" + suffix)
	s.req.NoError(s.repo.Create(ctx, txn))

	status := domain.TransactionStatusDraft
	limit := 50
	offset := 0
	results, err := s.repo.List(ctx, &domain.TransactionFilter{
		Status: &status, Limit: &limit, Offset: &offset,
	})
	s.req.NoError(err)
	for _, r := range results {
		s.req.Equal(domain.TransactionStatusDraft, r.TransactionStatus)
	}
}

// TestList_ByType verifies type filter returns only matching rows.
func (s *TransactionRepoSuite) TestList_ByType() {
	ctx := shared.WithTenantID(s.ctx, s.tenantA)
	suffix := uuid.New().String()[:6]

	ref := "SYS-REF-" + suffix
	sysTxn := &domain.Transaction{
		ID: uuid.New(), TransactionDate: time.Now(),
		TransactionType: domain.TransactionTypeSystem, TransactionStatus: domain.TransactionStatusDraft,
		TransactionNumber: "SYS-" + suffix, ReferenceNumber: &ref,
		Description: "System txn", CurrencyCode: "USD",
		ExchangeRate: decimal.NewFromFloat(1.0), CreatedBy: s.userID,
	}
	s.req.NoError(s.repo.Create(ctx, sysTxn))

	sysType := domain.TransactionTypeSystem
	limit := 50
	offset := 0
	results, err := s.repo.List(ctx, &domain.TransactionFilter{
		TransactionType: &sysType, Limit: &limit, Offset: &offset,
	})
	s.req.NoError(err)
	for _, r := range results {
		s.req.Equal(domain.TransactionTypeSystem, r.TransactionType)
	}
}

// ============================================================================
// Approval tests
// ============================================================================

// TestApprove_SetsApprovalFields verifies approval metadata is persisted.
func (s *TransactionRepoSuite) TestApprove_SetsApprovalFields() {
	ctx := shared.WithTenantID(s.ctx, s.tenantA)
	txn := s.newTransaction("APPROVAL-" + uuid.New().String()[:6])
	txn.ApprovalRequired = true
	txn.ApprovalStatus = domain.ApprovalStatusPending
	s.req.NoError(s.repo.Create(ctx, txn))

	approverID := s.userID
	approvedAt := time.Now()
	notes := "looks good"
	s.req.NoError(s.repo.Approve(ctx, txn.ID, approverID, approvedAt, &notes))

	got, err := s.repo.GetByID(ctx, txn.ID)
	s.req.NoError(err)
	s.req.Equal(domain.ApprovalStatusApproved, got.ApprovalStatus)
	s.req.Equal(approverID, *got.ApprovedBy)
	s.req.WithinDuration(approvedAt, *got.ApprovedAt, time.Second)
	s.req.Equal(notes, *got.ApprovalNotes)
}

// ============================================================================
// DB constraint tests
// ============================================================================

// TestDBConstraint_UniqueTransactionNumber verifies per-tenant uniqueness.
func (s *TransactionRepoSuite) TestDBConstraint_UniqueTransactionNumber() {
	ctx := shared.WithTenantID(s.ctx, s.tenantA)
	txn := s.newTransaction("DUP-" + uuid.New().String()[:6])
	s.req.NoError(s.repo.Create(ctx, txn))

	// Duplicate with same number, same tenant → must fail.
	dup := s.newTransaction(txn.TransactionNumber[4:]) // same suffix
	dup.TransactionNumber = txn.TransactionNumber
	err := s.repo.Create(ctx, dup)
	s.req.Error(err, "duplicate transaction number within same tenant must fail")
}

// TestDBConstraint_EntryAmounts verifies double-entry constraints at DB level.
func (s *TransactionRepoSuite) TestDBConstraint_EntryAmounts() {
	_, err := s.pool.Exec(context.Background(),
		`SELECT set_tenant_context($1)`, s.tenantA)
	s.req.NoError(err)

	txnID := uuid.New()
	_, err = s.pool.Exec(context.Background(), `
		INSERT INTO finance_transactions
		  (id, tenant_id, transaction_number, transaction_type, transaction_status,
		   transaction_date, description, currency_code, exchange_rate,
		   total_debit_amount, total_credit_amount, approval_status, created_by)
		VALUES
		  ($1, current_tenant_id(), $2, 'MANUAL', 'DRAFT',
		   NOW(), 'Constraint test', 'USD', 1.0, 0, 0, 'NOT_REQUIRED', $3)`,
		txnID,
		fmt.Sprintf("TXN-CONSTRAINT-%s", uuid.New().String()[:8]),
		s.userID,
	)
	if err != nil {
		s.T().Logf("skipping entry constraint test: cannot insert parent txn: %v", err)
		s.T().Skip()
	}
	defer func() {
		_, _ = s.pool.Exec(context.Background(),
			`DELETE FROM finance_transactions WHERE id = $1`, txnID)
	}()

	// Both debit AND credit non-zero must fail.
	_, err = s.pool.Exec(context.Background(), `
		INSERT INTO finance_transaction_entries
		  (tenant_id, transaction_id, entry_number, account_id,
		   debit_amount, credit_amount, description, exchange_rate)
		VALUES
		  (current_tenant_id(), $1, 1, $2, 500.00, 500.00, 'dual amounts', 1.0)`,
		txnID, uuid.New(),
	)
	s.req.Error(err, "both debit_amount > 0 AND credit_amount > 0 must be rejected")
}

// ============================================================================
// Recurring transaction tests
// ============================================================================

// TestCreate_RecurringTransaction verifies recurring metadata persisted.
func (s *TransactionRepoSuite) TestCreate_RecurringTransaction() {
	ctx := shared.WithTenantID(s.ctx, s.tenantA)
	freq := "MONTHLY"
	next := time.Now().AddDate(0, 1, 0)
	ref := "REC-REF-" + uuid.New().String()[:6]

	txn := &domain.Transaction{
		ID: uuid.New(), TransactionDate: time.Now(),
		TransactionType: domain.TransactionTypeManual, TransactionStatus: domain.TransactionStatusDraft,
		TransactionNumber:  "REC-" + uuid.New().String()[:8],
		ReferenceNumber:    &ref,
		Description:        "Recurring transaction",
		CurrencyCode:       "USD",
		ExchangeRate:       decimal.NewFromFloat(1.0),
		IsRecurring:        true,
		RecurringFrequency: &freq,
		NextRecurringDate:  &next,
		CreatedBy:          s.userID,
	}
	s.req.NoError(s.repo.Create(ctx, txn))

	got, err := s.repo.GetByID(ctx, txn.ID)
	s.req.NoError(err)
	s.req.True(got.IsRecurring)
	s.req.Equal(freq, *got.RecurringFrequency)
	s.req.WithinDuration(next, *got.NextRecurringDate, time.Second)
}
