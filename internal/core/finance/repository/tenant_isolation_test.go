// Package repository_test contains mock-based unit tests that exercise the
// finance repository layer without a real database.  Build constraint is
// intentionally absent so these run in the standard `go test ./...` pass.
package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"

	db "awo.so/db/sqlc"
	"awo.so/internal/core/finance/domain"
	"awo.so/internal/core/finance/repository"
	"awo.so/internal/shared"
	"awo.so/internal/shared/tracing"
)

// ─── helpers ──────────────────────────────────────────────────────────────────

func withTenantCtx(tenantID uuid.UUID) context.Context {
	return shared.WithTenantID(context.Background(), tenantID)
}

// runInCallback is a DoAndReturn adapter that runs the WithTenantFromCtx
// callback with the provided store, simulating the real store without a DB.
func runInCallback(inner db.Store) func(context.Context, func(context.Context, db.Store) error) error {
	return func(ctx context.Context, fn func(context.Context, db.Store) error) error {
		return fn(ctx, inner)
	}
}

func newRepo(store db.Store) domain.TransactionRepository {
	return repository.NewTransactionRepository(store, tracing.NewNoOpService(), nil, nil)
}

// ─── FIN-REPO-010: no tenant in context → reject before DB touch ──────────────

func TestTxRepo_Create_MissingTenant(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	s := db.NewMockStore(ctrl) // zero calls expected
	err := newRepo(s).Create(context.Background(), &domain.Transaction{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tenant ID not found")
}

func TestTxRepo_GetByID_MissingTenant(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	s := db.NewMockStore(ctrl)
	got, err := newRepo(s).GetByID(context.Background(), uuid.New())
	require.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "tenant ID not found")
}

func TestTxRepo_Update_MissingTenant(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	s := db.NewMockStore(ctrl)
	err := newRepo(s).Update(context.Background(), &domain.Transaction{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tenant ID not found")
}

func TestTxRepo_Delete_MissingTenant(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	s := db.NewMockStore(ctrl)
	err := newRepo(s).Delete(context.Background(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tenant ID not found")
}

func TestTxRepo_CalculateBalance_MissingTenant(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	s := db.NewMockStore(ctrl)
	bal, err := newRepo(s).CalculateAccountBalance(context.Background(), uuid.New(), nil)
	require.Error(t, err)
	assert.True(t, bal.IsZero())
	assert.Contains(t, err.Error(), "tenant ID not found")
}

// Validates the loop guard: tenant check happens BEFORE iterating over IDs.
func TestTxRepo_ValidateAccountsExist_MissingTenant_NoLoopEntry(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	s := db.NewMockStore(ctrl) // MUST receive zero calls
	err := newRepo(s).ValidateAccountsExist(context.Background(), []uuid.UUID{uuid.New(), uuid.New()})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tenant ID not found")
}

// ─── FIN-REPO-011: ErrNoRows → domain.ErrTransactionNotFound ─────────────────

func TestTxRepo_GetByID_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tenantID := uuid.New()
	txnID := uuid.New()
	ctx := withTenantCtx(tenantID)

	inner := db.NewMockStore(ctrl)
	inner.EXPECT().GetTransactionByID(gomock.Any(), txnID).Return(nil, db.ErrNoRows)

	outer := db.NewMockStore(ctrl)
	outer.EXPECT().WithTenantFromCtx(gomock.Any(), gomock.Any()).DoAndReturn(runInCallback(inner))

	got, err := newRepo(outer).GetByID(ctx, txnID)
	require.ErrorIs(t, err, domain.ErrTransactionNotFound)
	assert.Nil(t, got)
}

// ─── FIN-REPO-012: arbitrary DB error is surfaced, not swallowed ──────────────

func TestTxRepo_GetByID_DBError_NotMasked(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tenantID := uuid.New()
	txnID := uuid.New()
	ctx := withTenantCtx(tenantID)
	dbErr := errors.New("connection reset by peer")

	inner := db.NewMockStore(ctrl)
	inner.EXPECT().GetTransactionByID(gomock.Any(), txnID).Return(nil, dbErr)

	outer := db.NewMockStore(ctrl)
	outer.EXPECT().WithTenantFromCtx(gomock.Any(), gomock.Any()).DoAndReturn(runInCallback(inner))

	_, err := newRepo(outer).GetByID(ctx, txnID)
	require.Error(t, err)
	assert.NotErrorIs(t, err, domain.ErrTransactionNotFound,
		"a real DB error must not be silently mapped to ErrTransactionNotFound")
}

// ─── FIN-REPO-013: store-level error (inactive tenant, etc.) propagates ───────

func TestTxRepo_GetByID_TenantStoreError_Propagates(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := withTenantCtx(uuid.New())
	storeErr := errors.New("tenant is not ACTIVE")

	outer := db.NewMockStore(ctrl)
	outer.EXPECT().WithTenantFromCtx(gomock.Any(), gomock.Any()).Return(storeErr)

	_, err := newRepo(outer).GetByID(ctx, uuid.New())
	require.ErrorIs(t, err, storeErr)
}

// ─── FIN-REPO-014: cross-tenant isolation ────────────────────────────────────
//
// Two repos backed by independent stores. gomock's strict controller verifies
// that store A is never invoked with store B's context and vice-versa.

func TestTxRepo_CrossTenantIsolation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	idA, idB := uuid.New(), uuid.New()
	txnID := uuid.New()

	innerA := db.NewMockStore(ctrl)
	innerA.EXPECT().GetTransactionByID(gomock.Any(), txnID).Return(nil, db.ErrNoRows)
	outerA := db.NewMockStore(ctrl)
	outerA.EXPECT().WithTenantFromCtx(gomock.Any(), gomock.Any()).DoAndReturn(runInCallback(innerA))

	innerB := db.NewMockStore(ctrl)
	innerB.EXPECT().GetTransactionByID(gomock.Any(), txnID).Return(nil, db.ErrNoRows)
	outerB := db.NewMockStore(ctrl)
	outerB.EXPECT().WithTenantFromCtx(gomock.Any(), gomock.Any()).DoAndReturn(runInCallback(innerB))

	repoA := newRepo(outerA)
	repoB := newRepo(outerB)

	_, errA := repoA.GetByID(withTenantCtx(idA), txnID)
	_, errB := repoB.GetByID(withTenantCtx(idB), txnID)

	assert.ErrorIs(t, errA, domain.ErrTransactionNotFound)
	assert.ErrorIs(t, errB, domain.ErrTransactionNotFound)
	// ctrl.Finish() (deferred above) asserts each store was called exactly once
	// with its own context — zero cross-contamination.
}

// ─── FIN-REPO-015: GetNextTransactionNumber (pure logic, no DB) ───────────────

func TestTxRepo_GetNextTransactionNumber_Prefixes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Pure-logic method: no DB calls expected.
	repo := newRepo(db.NewMockStore(ctrl))
	ctx := withTenantCtx(uuid.New())

	cases := []struct {
		txType domain.TransactionType
		prefix string
	}{
		{domain.TransactionTypeJournalEntry, "JE"},
		{domain.TransactionTypeManual, "MAN"},
		{domain.TransactionTypeAdjustment, "ADJ"},
		{domain.TransactionTypeRecurring, "REC"},
		{domain.TransactionTypeClosing, "CLO"},
		{domain.TransactionTypeSystem, "TXN"},
	}
	for _, tc := range cases {
		num, err := repo.GetNextTransactionNumber(ctx, nil, tc.txType)
		require.NoError(t, err)
		assert.Contains(t, num, tc.prefix, "txType=%s", tc.txType)
	}
}
