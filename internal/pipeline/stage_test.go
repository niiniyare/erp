package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ── compile-time interface check ──────────────────────────────────────────────

// testStage embeds BaseStage and implements Execute, satisfying Stage.
type testStage struct {
	BaseStage
	executeErr error
	result     StageResult
}

func (s *testStage) Execute(opCtx *OperationContext) (StageResult, error) {
	return s.result, s.executeErr
}

var _ Stage = &testStage{}

// testSimulatableStage also implements Simulatable.
type testSimulatableStage struct {
	testStage
	simResult StageResult
	simErr    error
}

func (s *testSimulatableStage) Simulate(opCtx *OperationContext) (StageResult, error) {
	return s.simResult, s.simErr
}

var _ Stage      = &testSimulatableStage{}
var _ Simulatable = &testSimulatableStage{}

// ── BaseStage defaults ────────────────────────────────────────────────────────

func TestBaseStage_InterfaceSatisfied(t *testing.T) {
	// Compile-time assertion above handles this; this test documents intent.
	s := &testStage{
		BaseStage: BaseStage{StageName: "test.stage"},
	}
	assert.Implements(t, (*Stage)(nil), s)
}

func TestBaseStage_Name(t *testing.T) {
	s := &testStage{BaseStage: BaseStage{StageName: "budget.check"}}
	assert.Equal(t, "budget.check", s.Name())
}

func TestBaseStage_Defaults_Required(t *testing.T) {
	s := &testStage{}
	assert.False(t, s.Required(), "zero-value BaseStage should not be required")
}

func TestBaseStage_Defaults_RunCondition(t *testing.T) {
	s := &testStage{}
	assert.Equal(t, "", s.RunCondition(), "zero-value BaseStage should have empty run condition")
}

func TestBaseStage_Defaults_FeatureFlag(t *testing.T) {
	s := &testStage{}
	assert.Equal(t, "", s.FeatureFlag(), "zero-value BaseStage should have empty feature flag")
}

func TestBaseStage_Defaults_DependsOn(t *testing.T) {
	s := &testStage{}
	assert.Nil(t, s.DependsOn())
}

func TestBaseStage_Operations_Set(t *testing.T) {
	ops := []string{"ap.invoice.process", "ap.invoice.approve"}
	s := &testStage{BaseStage: BaseStage{StageOperations: ops}}
	assert.Equal(t, ops, s.Operations())
}

func TestBaseStage_Operations_Wildcard(t *testing.T) {
	s := &testStage{BaseStage: BaseStage{StageOperations: []string{"*"}}}
	assert.Equal(t, []string{"*"}, s.Operations())
}

func TestBaseStage_Priority(t *testing.T) {
	s := &testStage{BaseStage: BaseStage{StagePriority: 310}}
	assert.Equal(t, 310, s.Priority())
}

func TestBaseStage_Required_True(t *testing.T) {
	s := &testStage{BaseStage: BaseStage{StageRequired: true}}
	assert.True(t, s.Required())
}

func TestBaseStage_RunCondition_Set(t *testing.T) {
	cond := `budget_exceeded == true`
	s := &testStage{BaseStage: BaseStage{StageRunCond: cond}}
	assert.Equal(t, cond, s.RunCondition())
}

func TestBaseStage_DependsOn_Set(t *testing.T) {
	deps := []string{"ap.resolve_gl_accounts"}
	s := &testStage{BaseStage: BaseStage{StageDependsOn: deps}}
	assert.Equal(t, deps, s.DependsOn())
}

func TestBaseStage_FeatureFlag_Set(t *testing.T) {
	s := &testStage{BaseStage: BaseStage{StageFeatureFlag: "budget"}}
	assert.Equal(t, "budget", s.FeatureFlag())
}

// ── StageResult ───────────────────────────────────────────────────────────────

func TestStageResult_Fields(t *testing.T) {
	r := StageResult{
		Status:      "completed",
		Message:     "all good",
		NextStageID: "next.stage",
		Outputs: map[string]any{
			"budget_check.available_amount": "50000",
		},
	}

	assert.Equal(t, "completed", r.Status)
	assert.Equal(t, "all good", r.Message)
	assert.Equal(t, "next.stage", r.NextStageID)
	assert.Equal(t, "50000", r.Outputs["budget_check.available_amount"])
}

func TestStageResult_ValidStatuses(t *testing.T) {
	validStatuses := []string{"completed", "skipped", "suspended", "simulated", "failed"}
	for _, s := range validStatuses {
		r := StageResult{Status: s}
		assert.Equal(t, s, r.Status, "status %q should round-trip", s)
	}
}

// ── Simulatable ───────────────────────────────────────────────────────────────

func TestSimulatable_InterfaceSatisfied(t *testing.T) {
	s := &testSimulatableStage{}
	assert.Implements(t, (*Simulatable)(nil), s)
}

func TestSimulatable_Simulate_Called(t *testing.T) {
	opCtx := newTestOpCtx()
	expected := StageResult{Status: "simulated", Message: "dry run result"}

	s := &testSimulatableStage{
		simResult: expected,
	}

	got, err := s.Simulate(opCtx)
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}

// ── stageCheckpoint (internal) ────────────────────────────────────────────────

func TestStageCheckpoint_Fields(t *testing.T) {
	cp := stageCheckpoint{
		StageName: "gl.post_transaction",
		Output:    map[string]any{"gl_posting.transaction_id": "uuid-123"},
	}

	assert.Equal(t, "gl.post_transaction", cp.StageName)
	assert.Equal(t, "uuid-123", cp.Output["gl_posting.transaction_id"])
}
