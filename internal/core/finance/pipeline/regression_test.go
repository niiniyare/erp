// Package pipeline_test — pipeline stage regression and mutation tests.
//
// Verifies that stage execution contracts hold under adversarial conditions:
// missing context data, wrong input types, invalid transaction states, and
// stage priority ordering.
//
// These tests prove that each stage's guard cannot be silently bypassed.
package pipeline_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"awo.so/internal/core/finance/domain"
	"awo.so/internal/core/finance/pipeline"
	corePipeline "awo.so/internal/pipeline"
)

// ============================================================================
// FIN-PREG-001: BalanceCheckStage — wrong type in context fails gracefully
// ============================================================================
// If upstream code puts the wrong type under KeyEntries, the stage must fail
// with a clear error rather than panic.

func TestPipelineRegression_BalanceCheck_WrongEntryType_FailsGracefully(t *testing.T) {
	opCtx := newOpCtx(nil)
	opCtx.Data[pipeline.KeyEntries] = "this is not a []domain.TransactionEntry" // wrong type

	stage := pipeline.NewBalanceCheckStage()
	result, err := stage.Execute(opCtx)

	if err == nil {
		t.Fatal("expected error for wrong entry type, got nil")
	}
	if result.Status != "failed" {
		t.Errorf("expected status=failed, got %q", result.Status)
	}
}

// ============================================================================
// FIN-PREG-002: GLPostStage — nil input fails gracefully (no panic)
// ============================================================================

func TestPipelineRegression_GLPost_NilInput_FailsGracefully(t *testing.T) {
	repo := &stubGLRepo{}
	opCtx := newOpCtx(nil) // nil input, no PostTransactionInput

	stage := pipeline.NewGLPostStage(repo)
	result, err := stage.Execute(opCtx)

	if err == nil {
		t.Fatal("expected error for nil input, got nil")
	}
	if result.Status != "failed" {
		t.Errorf("expected status=failed, got %q", result.Status)
	}
}

// ============================================================================
// FIN-PREG-003: GLPostStage — CANCELLED status blocked by state machine
// ============================================================================

func TestPipelineRegression_GLPost_CancelledStatus_Blocked(t *testing.T) {
	txnID := uuid.New()
	txn := &domain.Transaction{
		ID:                txnID,
		TransactionStatus: domain.TransactionStatusCancelled,
		ValidationStatus:  domain.ValidationStatusValid,
	}

	postCalled := false
	repo := &stubGLRepo{
		fnPost: func(_ context.Context, _, _ uuid.UUID, _ time.Time) error {
			postCalled = true
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
	result, err := stage.Execute(opCtx)

	if err == nil {
		t.Fatal("expected error for CANCELLED→POSTED transition")
	}
	if result.Status != "failed" {
		t.Errorf("expected status=failed, got %q", result.Status)
	}
	if postCalled {
		t.Error("repo.Post called despite state machine block — guard bypassed")
	}
}

// ============================================================================
// FIN-PREG-004: PeriodCheckStage — missing input type fails gracefully
// ============================================================================
// When a period repo is configured but no PostTransactionInput is in the
// context, the stage must fail with a clear error rather than panic.

func TestPipelineRegression_PeriodCheck_MissingInput_FailsGracefully(t *testing.T) {
	// stubPeriodRepo to make the stage not skip (nil repo → skip).
	stubRepo := &stubPeriodRepo{}
	opCtx := newOpCtx("wrong input type") // not a *PostTransactionInput

	stage := pipeline.NewPeriodCheckStage(stubRepo)
	result, err := stage.Execute(opCtx)

	if err == nil {
		t.Fatal("expected error for missing input, got nil")
	}
	if result.Status != "failed" {
		t.Errorf("expected status=failed, got %q", result.Status)
	}
}

// ============================================================================
// FIN-PREG-005: Stage priorities enforce correct execution order
// ============================================================================
// Load(100) < Balance(200) < Period(300) < AccountValidate(400) < GLPost(600)
// This order ensures data is loaded before it is used, and posting happens last.

func TestPipelineRegression_StagePriorities_EnforceCorrectOrder(t *testing.T) {
	nilTxnRepo := &stubGLRepo{}
	stages := []struct {
		name     string
		priority int
	}{
		{"gl.load_transaction", pipeline.NewLoadTransactionStage(nilTxnRepo).Priority()},
		{"gl.balance_check", pipeline.NewBalanceCheckStage().Priority()},
		{"gl.period_check", pipeline.NewPeriodCheckStage(nil).Priority()},
		{"gl.post", pipeline.NewGLPostStage(nilTxnRepo).Priority()},
	}

	// Verify strict priority ordering.
	for i := 1; i < len(stages); i++ {
		prev := stages[i-1]
		curr := stages[i]
		if prev.priority >= curr.priority {
			t.Errorf("stage order violation: %s (priority=%d) must come before %s (priority=%d)",
				prev.name, prev.priority, curr.name, curr.priority)
		}
	}

	// Verify GL post has highest priority (runs last).
	glPostPriority := pipeline.NewGLPostStage(nilTxnRepo).Priority()
	for _, s := range stages[:len(stages)-1] {
		if s.priority >= glPostPriority {
			t.Errorf("stage %s (priority=%d) must run before GLPost (priority=%d)",
				s.name, s.priority, glPostPriority)
		}
	}
}

// ============================================================================
// FIN-PREG-006: All required stages report Required()=true
// ============================================================================
// If a stage incorrectly sets Required=false, pipeline aborts on error become
// soft failures — silent data corruption risk.

func TestPipelineRegression_AllFinanceStages_AreRequired(t *testing.T) {
	nilTxnRepo := &stubGLRepo{}
	stages := []struct {
		name  string
		stage interface{ Required() bool }
	}{
		{"LoadTransactionStage", pipeline.NewLoadTransactionStage(nilTxnRepo)},
		{"BalanceCheckStage", pipeline.NewBalanceCheckStage()},
		{"PeriodCheckStage", pipeline.NewPeriodCheckStage(nil)},
		{"GLPostStage", pipeline.NewGLPostStage(nilTxnRepo)},
	}

	for _, s := range stages {
		if !s.stage.Required() {
			t.Errorf("stage %s: Required()=false — errors would be silently swallowed", s.name)
		}
	}
}

// ── stub helpers ─────────────────────────────────────────────────────────────

// stubPeriodRepo implements domain.PeriodRepository minimally for regression tests.
type stubPeriodRepo struct{}

func (r *stubPeriodRepo) GetPeriodForDate(_ context.Context, _ uuid.UUID, _ time.Time) (*domain.AccountingPeriod, error) {
	return nil, nil
}
func (r *stubPeriodRepo) CreateFiscalYear(_ context.Context, _ *domain.FiscalYear) error {
	panic("not implemented")
}
func (r *stubPeriodRepo) GetFiscalYearByID(_ context.Context, _ uuid.UUID) (*domain.FiscalYear, error) {
	panic("not implemented")
}
func (r *stubPeriodRepo) GetFiscalYearByYear(_ context.Context, _ uuid.UUID, _ int) (*domain.FiscalYear, error) {
	panic("not implemented")
}
func (r *stubPeriodRepo) ListFiscalYears(_ context.Context, _ uuid.UUID) ([]*domain.FiscalYear, error) {
	panic("not implemented")
}
func (r *stubPeriodRepo) UpdateFiscalYear(_ context.Context, _ *domain.FiscalYear) error {
	panic("not implemented")
}
func (r *stubPeriodRepo) CreatePeriod(_ context.Context, _ *domain.AccountingPeriod) error {
	panic("not implemented")
}
func (r *stubPeriodRepo) GetPeriodByID(_ context.Context, _ uuid.UUID) (*domain.AccountingPeriod, error) {
	panic("not implemented")
}
func (r *stubPeriodRepo) GetCurrentPeriod(_ context.Context, _ uuid.UUID) (*domain.AccountingPeriod, error) {
	panic("not implemented")
}
func (r *stubPeriodRepo) ListPeriods(_ context.Context, _, _ uuid.UUID) ([]*domain.AccountingPeriod, error) {
	panic("not implemented")
}
func (r *stubPeriodRepo) UpdatePeriod(_ context.Context, _ *domain.AccountingPeriod) error {
	panic("not implemented")
}

// Ensure unused import doesn't cause compile error.
var _ = decimal.Zero
var _ = corePipeline.StageResult{}
