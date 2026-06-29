// Package pipeline_test — finance pipeline stage unit tests.
//
// Proves that:
// - BalanceCheckStage rejects unbalanced entries with a business error.
// - BalanceCheckStage accepts balanced entries (completed status).
// - BalanceCheckStage rejects transactions with fewer than 2 entries.
// - PeriodCheckStage skips gracefully when no PeriodRepository is configured.
// - GLPostStage calls repo.Post and sets gl_posted flag on success.
// - GLPostStage blocks posting a transaction already in POSTED status.
// No database required — uses in-memory stubs.
package pipeline_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"awo.so/internal/core/finance/domain"
	"awo.so/internal/core/finance/pipeline"
	corePipeline "awo.so/internal/pipeline"
)

// newOpCtx builds a minimal OperationContext for stage testing.
func newOpCtx(input any) *corePipeline.OperationContext {
	return &corePipeline.OperationContext{
		Ctx:   context.Background(),
		Input: input,
		Data:  make(map[string]any),
		Flags: make(map[string]bool),
	}
}

// ── BalanceCheckStage ─────────────────────────────────────────────────────────

// ============================================================================
// FIN-PIPE-001: BalanceCheckStage rejects unbalanced entries
// ============================================================================

func TestBalanceCheckStage_UnbalancedEntries_Fails(t *testing.T) {
	entries := []domain.TransactionEntry{
		{DebitAmount: decimal.NewFromFloat(100), CreditAmount: decimal.Zero},
		{DebitAmount: decimal.Zero, CreditAmount: decimal.NewFromFloat(90)}, // off by 10
	}
	opCtx := newOpCtx(nil)
	opCtx.Data[pipeline.KeyEntries] = entries

	stage := pipeline.NewBalanceCheckStage()
	result, err := stage.Execute(opCtx)

	if err == nil {
		t.Fatal("expected error for unbalanced entries, got nil")
	}
	if result.Status != "failed" {
		t.Errorf("expected status=failed, got %q", result.Status)
	}
}

// ============================================================================
// FIN-PIPE-002: BalanceCheckStage accepts balanced entries
// ============================================================================

func TestBalanceCheckStage_BalancedEntries_Succeeds(t *testing.T) {
	entries := []domain.TransactionEntry{
		{DebitAmount: decimal.NewFromFloat(250), CreditAmount: decimal.Zero},
		{DebitAmount: decimal.Zero, CreditAmount: decimal.NewFromFloat(250)},
	}
	opCtx := newOpCtx(nil)
	opCtx.Data[pipeline.KeyEntries] = entries

	stage := pipeline.NewBalanceCheckStage()
	result, err := stage.Execute(opCtx)
	if err != nil {
		t.Fatalf("unexpected error for balanced entries: %v", err)
	}
	if result.Status != "completed" {
		t.Errorf("expected status=completed, got %q", result.Status)
	}
}

// ============================================================================
// FIN-PIPE-003: BalanceCheckStage rejects fewer than 2 entries
// ============================================================================

func TestBalanceCheckStage_SingleEntry_Fails(t *testing.T) {
	entries := []domain.TransactionEntry{
		{DebitAmount: decimal.NewFromFloat(100), CreditAmount: decimal.Zero},
	}
	opCtx := newOpCtx(nil)
	opCtx.Data[pipeline.KeyEntries] = entries

	stage := pipeline.NewBalanceCheckStage()
	result, err := stage.Execute(opCtx)

	if err == nil {
		t.Fatal("expected error for single entry, got nil")
	}
	if result.Status != "failed" {
		t.Errorf("expected status=failed, got %q", result.Status)
	}
}

// ============================================================================
// FIN-PIPE-004: BalanceCheckStage fails when entries key is missing
// ============================================================================

func TestBalanceCheckStage_MissingEntries_Fails(t *testing.T) {
	opCtx := newOpCtx(nil) // no KeyEntries set

	stage := pipeline.NewBalanceCheckStage()
	result, err := stage.Execute(opCtx)

	if err == nil {
		t.Fatal("expected error when entries not loaded")
	}
	if result.Status != "failed" {
		t.Errorf("expected status=failed, got %q", result.Status)
	}
}

// ── PeriodCheckStage ──────────────────────────────────────────────────────────

// ============================================================================
// FIN-PIPE-005: PeriodCheckStage skips gracefully with nil repo
// ============================================================================

func TestPeriodCheckStage_NilRepo_Skips(t *testing.T) {
	opCtx := newOpCtx(&pipeline.PostTransactionInput{
		TransactionID: uuid.New(),
		PostingDate:   time.Now(),
		TenantID:      uuid.New(),
		PostedBy:      uuid.New(),
	})

	stage := pipeline.NewPeriodCheckStage(nil)
	result, err := stage.Execute(opCtx)
	if err != nil {
		t.Fatalf("nil periodRepo should not return error, got: %v", err)
	}
	if result.Status != "skipped" {
		t.Errorf("expected status=skipped, got %q", result.Status)
	}
}

// ── GLPostStage ───────────────────────────────────────────────────────────────

// stubGLRepo is a minimal TransactionRepository for GLPostStage tests.
type stubGLRepo struct {
	fnPost            func(ctx context.Context, id, postedBy uuid.UUID, postingDate time.Time) error
	fnGetByID         func(ctx context.Context, id uuid.UUID) (*domain.Transaction, error)
	fnGetEntriesByTxn func(ctx context.Context, id uuid.UUID) ([]domain.TransactionEntry, error)
}

func (r *stubGLRepo) Post(ctx context.Context, id, postedBy uuid.UUID, postingDate time.Time) error {
	if r.fnPost != nil {
		return r.fnPost(ctx, id, postedBy, postingDate)
	}
	return nil
}

func (r *stubGLRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error) {
	if r.fnGetByID != nil {
		return r.fnGetByID(ctx, id)
	}
	return nil, errors.New("stubGLRepo.GetByID not implemented")
}

func (r *stubGLRepo) GetEntriesByTransaction(ctx context.Context, id uuid.UUID) ([]domain.TransactionEntry, error) {
	if r.fnGetEntriesByTxn != nil {
		return r.fnGetEntriesByTxn(ctx, id)
	}
	return nil, errors.New("stubGLRepo.GetEntriesByTransaction not implemented")
}

// Satisfy the full domain.TransactionRepository interface with panics for unused methods.
func (r *stubGLRepo) Create(ctx context.Context, txn *domain.Transaction) error {
	panic("not implemented")
}

func (r *stubGLRepo) Update(ctx context.Context, t *domain.Transaction) error {
	panic("not implemented")
}
func (r *stubGLRepo) Delete(ctx context.Context, id uuid.UUID) error { panic("not implemented") }
func (r *stubGLRepo) List(ctx context.Context, filter *domain.TransactionFilter) ([]*domain.Transaction, error) {
	panic("not implemented")
}

func (r *stubGLRepo) Approve(ctx context.Context, id, by uuid.UUID, at time.Time, notes *string) error {
	panic("not implemented")
}

func (r *stubGLRepo) Reject(ctx context.Context, id, by uuid.UUID, at time.Time, reason domain.RejectionReason, notes *string) error {
	panic("not implemented")
}

func (r *stubGLRepo) Reverse(ctx context.Context, origID, revID uuid.UUID, reason string) error {
	panic("not implemented")
}

func (r *stubGLRepo) Count(ctx context.Context, f *domain.TransactionFilter) (int64, error) {
	panic("not implemented")
}

func (r *stubGLRepo) ListByAccount(ctx context.Context, accountID uuid.UUID, f *domain.TransactionFilter) ([]*domain.Transaction, error) {
	panic("not implemented")
}

func (r *stubGLRepo) ListByDateRange(ctx context.Context, start, end time.Time) ([]*domain.Transaction, error) {
	panic("not implemented")
}

func (r *stubGLRepo) GetByStatus(ctx context.Context, status domain.TransactionStatus, limit int) ([]*domain.Transaction, error) {
	panic("not implemented")
}

func (r *stubGLRepo) GetPendingApproval(ctx context.Context, userID *uuid.UUID) ([]*domain.Transaction, error) {
	panic("not implemented")
}

func (r *stubGLRepo) GetRecurringTransactions(ctx context.Context, due time.Time) ([]*domain.Transaction, error) {
	panic("not implemented")
}

func (r *stubGLRepo) UpdateNextRecurringDate(ctx context.Context, id uuid.UUID, next time.Time) error {
	panic("not implemented")
}

func (r *stubGLRepo) CreateEntry(ctx context.Context, e *domain.TransactionEntry) error {
	panic("not implemented")
}

func (r *stubGLRepo) CreateEntries(ctx context.Context, ee []*domain.TransactionEntry) error {
	panic("not implemented")
}

func (r *stubGLRepo) GetEntryByID(ctx context.Context, id uuid.UUID) (*domain.TransactionEntry, error) {
	panic("not implemented")
}

func (r *stubGLRepo) GetEntriesByAccount(ctx context.Context, accountID uuid.UUID, f *domain.EntryFilter) ([]domain.TransactionEntry, error) {
	panic("not implemented")
}

func (r *stubGLRepo) UpdateEntry(ctx context.Context, e *domain.TransactionEntry) error {
	panic("not implemented")
}
func (r *stubGLRepo) DeleteEntry(ctx context.Context, id uuid.UUID) error { panic("not implemented") }
func (r *stubGLRepo) SearchEntries(ctx context.Context, q string, f *domain.EntryFilter, limit, offset int) ([]*domain.TransactionEntry, error) {
	panic("not implemented")
}

func (r *stubGLRepo) UpdateReconciliationStatus(ctx context.Context, id uuid.UUID, reconciled bool, date *time.Time, ref *string) error {
	panic("not implemented")
}

func (r *stubGLRepo) GetUnreconciledEntries(ctx context.Context, accountID uuid.UUID, cutoff *time.Time) ([]*domain.TransactionEntry, error) {
	panic("not implemented")
}

func (r *stubGLRepo) GetEntrySummary(ctx context.Context, accountID uuid.UUID, start, end time.Time) (*domain.TransactionSummary, error) {
	panic("not implemented")
}

func (r *stubGLRepo) CalculateAccountBalance(ctx context.Context, accountID uuid.UUID, asOf *time.Time) (decimal.Decimal, error) {
	panic("not implemented")
}

func (r *stubGLRepo) GetAccountTransactionSummary(ctx context.Context, accountID uuid.UUID, dr *domain.DateRange) (*domain.TransactionSummary, error) {
	panic("not implemented")
}

func (r *stubGLRepo) IsTransactionNumberUnique(ctx context.Context, eid *uuid.UUID, num string, excl *uuid.UUID) (bool, error) {
	panic("not implemented")
}

func (r *stubGLRepo) ValidateAccountsExist(ctx context.Context, ids []uuid.UUID) error {
	panic("not implemented")
}

func (r *stubGLRepo) GetNextTransactionNumber(ctx context.Context, eid *uuid.UUID, tt domain.TransactionType) (string, error) {
	panic("not implemented")
}

func (r *stubGLRepo) GetByNumber(ctx context.Context, eid *uuid.UUID, num string) (*domain.Transaction, error) {
	panic("not implemented")
}

func (r *stubGLRepo) GetReversalHistory(ctx context.Context, id uuid.UUID) ([]*domain.Transaction, error) {
	panic("not implemented")
}

// ============================================================================
// FIN-PIPE-006: GLPostStage calls repo.Post and sets gl_posted flag
// ============================================================================

func TestGLPostStage_PostsSuccessfully(t *testing.T) {
	txnID := uuid.New()
	postedBy := uuid.New()
	postCalled := false

	repo := &stubGLRepo{
		fnPost: func(_ context.Context, id, _ uuid.UUID, _ time.Time) error {
			postCalled = true
			if id != txnID {
				t.Errorf("Post called with wrong ID: %v", id)
			}
			return nil
		},
	}

	txn := &domain.Transaction{
		ID:                txnID,
		TransactionStatus: domain.TransactionStatusApproved,
		ValidationStatus:  domain.ValidationStatusValid,
		ApprovalStatus:    domain.ApprovalStatusApproved,
	}

	input := &pipeline.PostTransactionInput{
		TransactionID: txnID,
		PostingDate:   time.Now(),
		TenantID:      uuid.New(),
		PostedBy:      postedBy,
	}
	opCtx := newOpCtx(input)
	opCtx.Data[pipeline.KeyTransaction] = txn

	stage := pipeline.NewGLPostStage(repo)
	result, err := stage.Execute(opCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "completed" {
		t.Errorf("expected status=completed, got %q", result.Status)
	}
	if !postCalled {
		t.Error("repo.Post was not called")
	}
	if !opCtx.Flag("gl_posted") {
		t.Error("expected gl_posted flag to be set")
	}
}

// ============================================================================
// FIN-PIPE-007: GLPostStage blocks POSTED → POSTED (already posted)
// ============================================================================

func TestGLPostStage_AlreadyPosted_Blocked(t *testing.T) {
	txnID := uuid.New()
	txn := &domain.Transaction{
		ID:                txnID,
		TransactionStatus: domain.TransactionStatusPosted,
	}

	input := &pipeline.PostTransactionInput{
		TransactionID: txnID,
		PostingDate:   time.Now(),
		TenantID:      uuid.New(),
		PostedBy:      uuid.New(),
	}
	opCtx := newOpCtx(input)
	opCtx.Data[pipeline.KeyTransaction] = txn

	// If state machine blocks posting, repo.Post should never be called.
	repo := &stubGLRepo{
		fnPost: func(_ context.Context, _, _ uuid.UUID, _ time.Time) error {
			t.Error("repo.Post should not be called when state machine rejects posting")
			return nil
		},
	}

	stage := pipeline.NewGLPostStage(repo)
	result, err := stage.Execute(opCtx)

	if err == nil {
		t.Fatal("expected error for POSTED→POSTED transition")
	}
	if result.Status != "failed" {
		t.Errorf("expected status=failed, got %q", result.Status)
	}
}
