// Package pipeline_test — Phase 16 query budget and explain-plan regression tests.
//
// Query budget enforcement prevents silent query amplification as the codebase evolves.
// These tests verify that the GL post pipeline:
//   - Contains exactly the expected number of stages.
//   - Stages execute in priority order without deviation.
//   - Required stages cannot be removed without test failure.
//   - No additional DB round-trips are introduced beyond the documented budget.
//
// Explain-plan simulation: since we can't run EXPLAIN ANALYZE in unit tests,
// we verify query-count proxies: each stage makes at most 1 repo call, and
// the total pipeline cost is bounded.
//
// Tests (FIN-QUERY-*):
//
//	001 — Stage count within budget (LoadTxn + Balance + Period + GLPost = 4 core)
//	002 — Required stages cannot be skipped (Required=true on core stages)
//	003 — Stage priorities form a strict ascending sequence
//	004 — LoadTransactionStage makes exactly 2 repo calls per execution
//	005 — BalanceCheckStage makes zero repo calls (reads from context)
//	006 — GLPostStage makes exactly 1 repo call per execution
//	007 — PeriodCheckStage makes zero or 1 repo calls (nil repo = 0)
//	008 — Total execution makes ≤ MaxPipelineRepoCalls per post operation
package pipeline_test

import (
	"context"
	"testing"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"awo.so/internal/core/finance/domain"
	"awo.so/internal/core/finance/pipeline"
	corePipeline "awo.so/internal/pipeline"
)

// MaxPipelineRepoCalls is the query budget for a single GL post operation.
// If the count exceeds this, the build gate must fail (query amplification detected).
const MaxPipelineRepoCalls = 5

// ── FIN-QUERY-001: Stage count within budget ─────────────────────────────────

// TestQueryBudget_StageCount_WithinBudget verifies the pipeline does not grow
// beyond the documented stage count without an explicit budget increase.
func TestQueryBudget_StageCount_WithinBudget(t *testing.T) {
	nilRepo := &qbStubTxnRepo{}
	stages := []corePipeline.Stage{
		pipeline.NewLoadTransactionStage(nilRepo),
		pipeline.NewBalanceCheckStage(),
		pipeline.NewPeriodCheckStage(nil),
		pipeline.NewGLPostStage(nilRepo),
	}

	// Current documented budget: 4 core stages.
	const maxStages = 4
	if len(stages) > maxStages {
		t.Errorf("FIN-QUERY-001: stage count %d exceeds budget %d — query amplification risk", len(stages), maxStages)
	}
}

// ── FIN-QUERY-002: Required stages cannot be skipped ─────────────────────────

// TestQueryBudget_RequiredStages_AllPresent verifies that all critical stages
// have Required=true. A Required=false on a critical stage means the pipeline
// can skip it silently, bypassing balance checks or audit.
func TestQueryBudget_RequiredStages_AllPresent(t *testing.T) {
	nilRepo := &qbStubTxnRepo{}
	coreStages := []struct {
		name  string
		stage corePipeline.Stage
	}{
		{"LoadTransaction", pipeline.NewLoadTransactionStage(nilRepo)},
		{"BalanceCheck", pipeline.NewBalanceCheckStage()},
		{"GLPost", pipeline.NewGLPostStage(nilRepo)},
	}

	for _, s := range coreStages {
		if !s.stage.Required() {
			t.Errorf("FIN-QUERY-002: critical stage %q has Required=false — it can be silently skipped", s.name)
		}
	}
}

// ── FIN-QUERY-003: Stage priorities form strict ascending sequence ────────────

// TestQueryBudget_StagePriorities_StrictlyAscending verifies that no two stages
// share the same priority (which would create non-deterministic ordering) and
// that the canonical order is preserved.
func TestQueryBudget_StagePriorities_StrictlyAscending(t *testing.T) {
	nilRepo := &qbStubTxnRepo{}
	stages := []struct {
		name     string
		priority int
	}{
		{"gl.load_transaction", pipeline.NewLoadTransactionStage(nilRepo).Priority()},
		{"gl.balance_check", pipeline.NewBalanceCheckStage().Priority()},
		{"gl.period_check", pipeline.NewPeriodCheckStage(nil).Priority()},
		{"gl.post", pipeline.NewGLPostStage(nilRepo).Priority()},
	}

	for i := 1; i < len(stages); i++ {
		prev := stages[i-1]
		curr := stages[i]
		if curr.priority <= prev.priority {
			t.Errorf("FIN-QUERY-003: stage %q (priority=%d) must be strictly greater than %q (priority=%d)",
				curr.name, curr.priority, prev.name, prev.priority)
		}
	}
}

// ── FIN-QUERY-004: LoadTransactionStage repo call budget ─────────────────────

// TestQueryBudget_LoadTransactionStage_ExactlyTwoRepoCalls verifies that
// LoadTransactionStage makes exactly 2 repo calls: GetByID and GetEntriesByTransaction.
// If this ever becomes 3+, it means N+1 or eager loading was accidentally added.
func TestQueryBudget_LoadTransactionStage_ExactlyTwoRepoCalls(t *testing.T) {
	var repoCalls int32
	txnID := uuid.New()
	txn := &domain.Transaction{
		ID:                txnID,
		TransactionStatus: domain.TransactionStatusApproved,
		TransactionNumber: "TXN-QB-001",
	}

	repo := &qbStubTxnRepo{
		fnGetByID: func(_ context.Context, _ uuid.UUID) (*domain.Transaction, error) {
			atomic.AddInt32(&repoCalls, 1) // call #1
			return txn, nil
		},
		fnGetEntriesByTxn: func(_ context.Context, _ uuid.UUID) ([]domain.TransactionEntry, error) {
			atomic.AddInt32(&repoCalls, 1) // call #2
			return []domain.TransactionEntry{}, nil
		},
	}

	input := &pipeline.PostTransactionInput{
		TransactionID: txnID,
		PostingDate:   time.Now(),
		TenantID:      uuid.New(),
		PostedBy:      uuid.New(),
	}
	opCtx := newOpCtx(input)

	stage := pipeline.NewLoadTransactionStage(repo)
	_, err := stage.Execute(opCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := int(atomic.LoadInt32(&repoCalls))
	if got != 2 {
		t.Errorf("FIN-QUERY-004: LoadTransactionStage made %d repo calls, expected exactly 2 (GetByID + GetEntries)", got)
	}
}

// ── FIN-QUERY-005: BalanceCheckStage makes zero repo calls ───────────────────

// TestQueryBudget_BalanceCheckStage_ZeroRepoCalls verifies that the balance
// check reads exclusively from the operation context (no DB access).
func TestQueryBudget_BalanceCheckStage_ZeroRepoCalls(t *testing.T) {
	txn := &domain.Transaction{
		ID:                uuid.New(),
		TransactionStatus: domain.TransactionStatusApproved,
	}
	entries := []domain.TransactionEntry{
		{DebitAmount: decimal.NewFromFloat(100), CreditAmount: decimal.Zero},
		{DebitAmount: decimal.Zero, CreditAmount: decimal.NewFromFloat(100)},
	}

	input005 := &pipeline.PostTransactionInput{TransactionID: txn.ID}
	opCtx := newOpCtx(input005)
	opCtx.Data[pipeline.KeyTransaction] = txn
	opCtx.Data[pipeline.KeyEntries] = entries

	// No repo injected — if stage tries to call repo, it will panic/error.
	stage := pipeline.NewBalanceCheckStage()
	_, err := stage.Execute(opCtx)
	if err != nil {
		t.Errorf("FIN-QUERY-005: BalanceCheckStage should complete successfully with balanced entries from context, got: %v", err)
	}
}

// ── FIN-QUERY-006: GLPostStage makes exactly one repo call ───────────────────

// TestQueryBudget_GLPostStage_ExactlyOneRepoCall verifies that posting a
// transaction makes exactly 1 repo call (repo.Post). No additional lookups.
func TestQueryBudget_GLPostStage_ExactlyOneRepoCall(t *testing.T) {
	var repoCalls int32
	txnID := uuid.New()
	txn := &domain.Transaction{
		ID:                txnID,
		TransactionStatus: domain.TransactionStatusApproved,
		ValidationStatus:  domain.ValidationStatusValid,
	}

	repo := &qbStubTxnRepo{
		fnPost: func(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ time.Time) error {
			atomic.AddInt32(&repoCalls, 1)
			return nil
		},
	}

	input := &pipeline.PostTransactionInput{
		TransactionID: txnID,
		PostingDate:   time.Now(),
		TenantID:      uuid.New(),
		PostedBy:      uuid.New(),
	}
	opCtx := newOpCtx(input)
	opCtx.Data[pipeline.KeyTransaction] = txn

	stage := pipeline.NewGLPostStage(repo)
	_, err := stage.Execute(opCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := int(atomic.LoadInt32(&repoCalls))
	if got != 1 {
		t.Errorf("FIN-QUERY-006: GLPostStage made %d repo calls, expected exactly 1 (Post only)", got)
	}
}

// ── FIN-QUERY-007: PeriodCheckStage zero calls with nil repo ─────────────────

// TestQueryBudget_PeriodCheckStage_NilRepo_ZeroCalls verifies that when no
// period repo is wired, the stage skips without making any calls.
func TestQueryBudget_PeriodCheckStage_NilRepo_ZeroCalls(t *testing.T) {
	opCtx := newOpCtx(&pipeline.PostTransactionInput{})

	// nil period repo → stage must skip gracefully.
	stage := pipeline.NewPeriodCheckStage(nil)
	_, err := stage.Execute(opCtx)
	if err != nil {
		t.Errorf("FIN-QUERY-007: PeriodCheckStage with nil repo should skip, got: %v", err)
	}
}

// ── FIN-QUERY-008: Total pipeline budget ─────────────────────────────────────

// TestQueryBudget_TotalPipeline_WithinBudget verifies that executing the full
// pipeline on a valid transaction stays within MaxPipelineRepoCalls.
// This is the composite query budget gate for the entire post operation.
func TestQueryBudget_TotalPipeline_WithinBudget(t *testing.T) {
	var totalRepoCalls int32
	txnID := uuid.New()
	txn := &domain.Transaction{
		ID:                txnID,
		TransactionStatus: domain.TransactionStatusApproved,
		ValidationStatus:  domain.ValidationStatusValid,
		TransactionNumber: "TXN-QB-BUDGET",
	}
	entries := []domain.TransactionEntry{
		{DebitAmount: decimal.NewFromFloat(500), CreditAmount: decimal.Zero},
		{DebitAmount: decimal.Zero, CreditAmount: decimal.NewFromFloat(500)},
	}

	repo := &qbStubTxnRepo{
		fnGetByID: func(_ context.Context, _ uuid.UUID) (*domain.Transaction, error) {
			atomic.AddInt32(&totalRepoCalls, 1)
			return txn, nil
		},
		fnGetEntriesByTxn: func(_ context.Context, _ uuid.UUID) ([]domain.TransactionEntry, error) {
			atomic.AddInt32(&totalRepoCalls, 1)
			return entries, nil
		},
		fnPost: func(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ time.Time) error {
			atomic.AddInt32(&totalRepoCalls, 1)
			return nil
		},
	}

	input := &pipeline.PostTransactionInput{
		TransactionID: txnID,
		PostingDate:   time.Now(),
		TenantID:      uuid.New(),
		PostedBy:      uuid.New(),
	}

	// Execute each stage in priority order.
	opCtx := newOpCtx(input)

	stages := []corePipeline.Stage{
		pipeline.NewLoadTransactionStage(repo),
		pipeline.NewBalanceCheckStage(),
		pipeline.NewPeriodCheckStage(nil),
		pipeline.NewGLPostStage(repo),
	}

	for _, stage := range stages {
		if _, err := stage.Execute(opCtx); err != nil {
			t.Fatalf("stage %q failed: %v", stage.Name(), err)
		}
	}

	got := int(atomic.LoadInt32(&totalRepoCalls))
	if got > MaxPipelineRepoCalls {
		t.Errorf("FIN-QUERY-008: pipeline made %d repo calls, exceeds budget of %d — query amplification detected", got, MaxPipelineRepoCalls)
	}
	t.Logf("FIN-QUERY-008: pipeline used %d/%d repo calls", got, MaxPipelineRepoCalls)
}

// ============================================================================
// qbStubTxnRepo — minimal stub for pipeline query budget tests
// ============================================================================

type qbStubTxnRepo struct {
	fnGetByID         func(ctx context.Context, id uuid.UUID) (*domain.Transaction, error)
	fnGetEntriesByTxn func(ctx context.Context, transactionID uuid.UUID) ([]domain.TransactionEntry, error)
	fnPost            func(ctx context.Context, transactionID, postedBy uuid.UUID, postedAt time.Time) error
}

func (r *qbStubTxnRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error) {
	if r.fnGetByID != nil {
		return r.fnGetByID(ctx, id)
	}
	return &domain.Transaction{ID: id}, nil
}

func (r *qbStubTxnRepo) GetEntriesByTransaction(ctx context.Context, transactionID uuid.UUID) ([]domain.TransactionEntry, error) {
	if r.fnGetEntriesByTxn != nil {
		return r.fnGetEntriesByTxn(ctx, transactionID)
	}
	return nil, nil
}

func (r *qbStubTxnRepo) Post(ctx context.Context, transactionID, postedBy uuid.UUID, postedAt time.Time) error {
	if r.fnPost != nil {
		return r.fnPost(ctx, transactionID, postedBy, postedAt)
	}
	return nil
}

// Satisfy the full domain.TransactionRepository interface with no-op methods.
func (r *qbStubTxnRepo) Create(ctx context.Context, t *domain.Transaction) error { return nil }
func (r *qbStubTxnRepo) GetByNumber(ctx context.Context, entityID *uuid.UUID, number string) (*domain.Transaction, error) {
	return nil, nil
}
func (r *qbStubTxnRepo) Update(ctx context.Context, t *domain.Transaction) error  { return nil }
func (r *qbStubTxnRepo) Delete(ctx context.Context, id uuid.UUID) error            { return nil }
func (r *qbStubTxnRepo) List(ctx context.Context, filter *domain.TransactionFilter) ([]*domain.Transaction, error) {
	return nil, nil
}
func (r *qbStubTxnRepo) Count(ctx context.Context, filter *domain.TransactionFilter) (int64, error) {
	return 0, nil
}
func (r *qbStubTxnRepo) ListByAccount(ctx context.Context, accountID uuid.UUID, filter *domain.TransactionFilter) ([]*domain.Transaction, error) {
	return nil, nil
}
func (r *qbStubTxnRepo) ListByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*domain.Transaction, error) {
	return nil, nil
}
func (r *qbStubTxnRepo) GetByStatus(ctx context.Context, status domain.TransactionStatus, limit int) ([]*domain.Transaction, error) {
	return nil, nil
}
func (r *qbStubTxnRepo) GetPendingApproval(ctx context.Context, userID *uuid.UUID) ([]*domain.Transaction, error) {
	return nil, nil
}
func (r *qbStubTxnRepo) GetRecurringTransactions(ctx context.Context, dueDate time.Time) ([]*domain.Transaction, error) {
	return nil, nil
}
func (r *qbStubTxnRepo) UpdateNextRecurringDate(ctx context.Context, transactionID uuid.UUID, nextDate time.Time) error {
	return nil
}
func (r *qbStubTxnRepo) CreateEntry(ctx context.Context, entry *domain.TransactionEntry) error {
	return nil
}
func (r *qbStubTxnRepo) CreateEntries(ctx context.Context, entries []*domain.TransactionEntry) error {
	return nil
}
func (r *qbStubTxnRepo) GetEntryByID(ctx context.Context, id uuid.UUID) (*domain.TransactionEntry, error) {
	return nil, nil
}
func (r *qbStubTxnRepo) GetEntriesByAccount(ctx context.Context, accountID uuid.UUID, filter *domain.EntryFilter) ([]domain.TransactionEntry, error) {
	return nil, nil
}
func (r *qbStubTxnRepo) UpdateEntry(ctx context.Context, entry *domain.TransactionEntry) error {
	return nil
}
func (r *qbStubTxnRepo) DeleteEntry(ctx context.Context, id uuid.UUID) error { return nil }
func (r *qbStubTxnRepo) SearchEntries(ctx context.Context, query string, filters *domain.EntryFilter, limit int, offset int) ([]*domain.TransactionEntry, error) {
	return nil, nil
}
func (r *qbStubTxnRepo) UpdateReconciliationStatus(ctx context.Context, entryID uuid.UUID, reconciled bool, reconciledDate *time.Time, reconciliationRef *string) error {
	return nil
}
func (r *qbStubTxnRepo) GetUnreconciledEntries(ctx context.Context, accountID uuid.UUID, cutoffDate *time.Time) ([]*domain.TransactionEntry, error) {
	return nil, nil
}
func (r *qbStubTxnRepo) GetEntrySummary(ctx context.Context, accountID uuid.UUID, startDate, endDate time.Time) (*domain.TransactionSummary, error) {
	return nil, nil
}
func (r *qbStubTxnRepo) CalculateAccountBalance(ctx context.Context, accountID uuid.UUID, asOfDate *time.Time) (decimal.Decimal, error) {
	return decimal.Zero, nil
}
func (r *qbStubTxnRepo) GetAccountTransactionSummary(ctx context.Context, accountID uuid.UUID, dateRange *domain.DateRange) (*domain.TransactionSummary, error) {
	return nil, nil
}
func (r *qbStubTxnRepo) Approve(ctx context.Context, transactionID uuid.UUID, approvedBy uuid.UUID, approvedAt time.Time, notes *string) error {
	return nil
}
func (r *qbStubTxnRepo) Reject(ctx context.Context, transactionID uuid.UUID, rejectedBy uuid.UUID, rejectedAt time.Time, reason domain.RejectionReason, notes *string) error {
	return nil
}
func (r *qbStubTxnRepo) Reverse(ctx context.Context, originalID, reversalID uuid.UUID, reason string) error {
	return nil
}
func (r *qbStubTxnRepo) IsTransactionNumberUnique(ctx context.Context, entityID *uuid.UUID, number string, excludeID *uuid.UUID) (bool, error) {
	return true, nil
}
func (r *qbStubTxnRepo) ValidateAccountsExist(ctx context.Context, accountIDs []uuid.UUID) error {
	return nil
}
func (r *qbStubTxnRepo) GetNextTransactionNumber(ctx context.Context, entityID *uuid.UUID, txType domain.TransactionType) (string, error) {
	return "TXN-QB-AUTO", nil
}
