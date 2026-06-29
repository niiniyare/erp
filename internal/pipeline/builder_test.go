package pipeline

import (
	"context"
	"errors"
	"testing"

	db "awo.so/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── TxRunner stub ─────────────────────────────────────────────────────────────

type stubTxRunner struct {
	called bool
	err    error
}

func (s *stubTxRunner) RunInTx(_ context.Context, fn func(context.Context, db.Store) error) error {
	s.called = true
	if s.err != nil {
		return s.err
	}
	// Pass a nil store — hooks in tests must not use it
	return fn(context.Background(), nil)
}

// ── Compensatable stub ────────────────────────────────────────────────────────

type compensatableStage struct {
	testStage
	compensated   bool
	compensateErr error
}

func (s *compensatableStage) Compensate(_ *OperationContext, _ map[string]any) error {
	s.compensated = true
	return s.compensateErr
}

var _ Compensatable = &compensatableStage{}

// ── helpers ───────────────────────────────────────────────────────────────────

func opCtxForOp(key string) *OperationContext {
	o := newTestOpCtx()
	o.Ctx = context.Background()
	o.OperationKey = key
	return o
}

func okStage(name string, ops []string, priority int) *testStage {
	return &testStage{
		BaseStage: BaseStage{
			StageName:       name,
			StageOperations: ops,
			StagePriority:   priority,
		},
		result: StageResult{Status: "completed", Message: "ok"},
	}
}

func errStage(name string, ops []string, priority int, required bool, err error) *testStage {
	return &testStage{
		BaseStage: BaseStage{
			StageName:       name,
			StageOperations: ops,
			StagePriority:   priority,
			StageRequired:   required,
		},
		executeErr: err,
		result:     StageResult{Status: "failed"},
	}
}

// ── NewPipelineBuilder ────────────────────────────────────────────────────────

func TestNewPipelineBuilder_NotNil(t *testing.T) {
	pb := NewPipelineBuilder(NewStageRegistry(), nil)
	require.NotNil(t, pb)
}

// ── Run: happy path ───────────────────────────────────────────────────────────

func TestRun_NoStages_Succeeds(t *testing.T) {
	pb := NewPipelineBuilder(NewStageRegistry(), nil)
	opCtx := opCtxForOp("ap.invoice.process")

	err := pb.Run(opCtx)
	require.NoError(t, err)
	assert.Empty(t, opCtx.Log)
}

func TestRun_SingleStage_Logged(t *testing.T) {
	r := NewStageRegistry()
	r.Register(okStage("ap.validate", []string{"ap.invoice.process"}, 110))

	pb := NewPipelineBuilder(r, nil)
	opCtx := opCtxForOp("ap.invoice.process")

	require.NoError(t, pb.Run(opCtx))
	require.Len(t, opCtx.Log, 1)
	assert.Equal(t, "ap.validate", opCtx.Log[0].StageName)
	assert.Equal(t, "completed", opCtx.Log[0].Status)
}

func TestRun_StagesExecutedInPriorityOrder(t *testing.T) {
	r := NewStageRegistry()
	r.Register(
		okStage("c", []string{"op"}, 300),
		okStage("a", []string{"op"}, 100),
		okStage("b", []string{"op"}, 200),
	)

	pb := NewPipelineBuilder(r, nil)
	opCtx := opCtxForOp("op")

	require.NoError(t, pb.Run(opCtx))
	require.Len(t, opCtx.Log, 3)
	assert.Equal(t, "a", opCtx.Log[0].StageName)
	assert.Equal(t, "b", opCtx.Log[1].StageName)
	assert.Equal(t, "c", opCtx.Log[2].StageName)
}

func TestRun_OutputsPropagatedToData(t *testing.T) {
	s := &testStage{
		BaseStage: BaseStage{StageName: "budget.check", StageOperations: []string{"op"}, StagePriority: 100},
		result: StageResult{
			Status:  "completed",
			Outputs: map[string]any{"budget_check.available": "50000"},
		},
	}
	r := NewStageRegistry()
	r.Register(s)

	pb := NewPipelineBuilder(r, nil)
	opCtx := opCtxForOp("op")

	require.NoError(t, pb.Run(opCtx))
	assert.Equal(t, "50000", opCtx.Data["budget_check.available"])
}

// ── Run: feature flag filtering ───────────────────────────────────────────────

func TestRun_FeatureFlagOff_StageSkipped(t *testing.T) {
	s := &testStage{
		BaseStage: BaseStage{
			StageName:        "budget.check",
			StageOperations:  []string{"op"},
			StagePriority:    100,
			StageFeatureFlag: "budget",
		},
		result: StageResult{Status: "completed"},
	}
	r := NewStageRegistry()
	r.Register(s)

	pb := NewPipelineBuilder(r, nil)
	opCtx := opCtxForOp("op")
	// budget flag is off (not in session flags)

	require.NoError(t, pb.Run(opCtx))
	assert.Empty(t, opCtx.Log, "stage with disabled feature flag must not execute")
}

func TestRun_FeatureFlagOn_StageRuns(t *testing.T) {
	s := &testStage{
		BaseStage: BaseStage{
			StageName:        "budget.check",
			StageOperations:  []string{"op"},
			StagePriority:    100,
			StageFeatureFlag: "budget",
		},
		result: StageResult{Status: "completed"},
	}
	r := NewStageRegistry()
	r.Register(s)

	pb := NewPipelineBuilder(r, nil)
	opCtx := opCtxForOp("op")
	opCtx.Session = newTestSession(map[string]bool{"budget": true}, nil)

	require.NoError(t, pb.Run(opCtx))
	require.Len(t, opCtx.Log, 1)
	assert.Equal(t, "completed", opCtx.Log[0].Status)
}

// ── Run: error handling ───────────────────────────────────────────────────────

func TestRun_RequiredStageError_AbortsAndReturnsError(t *testing.T) {
	r := NewStageRegistry()
	r.Register(
		okStage("a", []string{"op"}, 100),
		errStage("b", []string{"op"}, 200, true, errors.New("validation failed")),
		okStage("c", []string{"op"}, 300), // must NOT run
	)

	pb := NewPipelineBuilder(r, nil)
	opCtx := opCtxForOp("op")

	err := pb.Run(opCtx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")

	names := loggedStageNames(opCtx)
	assert.Contains(t, names, "a")
	assert.Contains(t, names, "b")
	assert.NotContains(t, names, "c", "stage after required failure must not run")
}

func TestRun_NonRequiredStageError_ContinuesExecution(t *testing.T) {
	r := NewStageRegistry()
	r.Register(
		okStage("a", []string{"op"}, 100),
		errStage("b", []string{"op"}, 200, false, errors.New("enrichment failed")),
		okStage("c", []string{"op"}, 300),
	)

	pb := NewPipelineBuilder(r, nil)
	opCtx := opCtxForOp("op")

	require.NoError(t, pb.Run(opCtx))

	names := loggedStageNames(opCtx)
	assert.Contains(t, names, "a")
	assert.Contains(t, names, "b")
	assert.Contains(t, names, "c", "non-required failure must not stop subsequent stages")
}

// ── Run: compensation ─────────────────────────────────────────────────────────

func TestRun_RequiredFailure_CompensatesCompletedStages(t *testing.T) {
	comp := &compensatableStage{}
	comp.StageName = "gl.post"
	comp.StageOperations = []string{"op"}
	comp.StagePriority = 100
	comp.result = StageResult{Status: "completed"}

	r := NewStageRegistry()
	r.Register(
		comp,
		errStage("audit.log", []string{"op"}, 200, true, errors.New("audit db down")),
	)

	pb := NewPipelineBuilder(r, nil)
	opCtx := opCtxForOp("op")

	err := pb.Run(opCtx)
	require.Error(t, err)
	assert.True(t, comp.compensated, "completed compensatable stage must be compensated on abort")
}

// ── Run: suspension ───────────────────────────────────────────────────────────

func TestRun_SuspendedStage_StopsExecution(t *testing.T) {
	suspending := &testStage{
		BaseStage: BaseStage{StageName: "workflow.gate", StageOperations: []string{"op"}, StagePriority: 500},
		result:    StageResult{Status: "suspended"},
	}
	// Override Execute to actually suspend the context
	suspending.result = StageResult{Status: "suspended"}

	r := NewStageRegistry()
	r.Register(
		okStage("validate", []string{"op"}, 100),
		&suspendingStage{name: "workflow.gate", ops: []string{"op"}, priority: 500},
		okStage("gl.post", []string{"op"}, 600),
	)

	pb := NewPipelineBuilder(r, nil)
	opCtx := opCtxForOp("op")

	require.NoError(t, pb.Run(opCtx))
	assert.True(t, opCtx.Suspended)

	names := loggedStageNames(opCtx)
	assert.Contains(t, names, "validate")
	assert.Contains(t, names, "workflow.gate")
	assert.NotContains(t, names, "gl.post", "stages after suspension must not run")
}

// ── RunFrom ───────────────────────────────────────────────────────────────────

func TestRunFrom_ResumesFromNamedStage(t *testing.T) {
	r := NewStageRegistry()
	r.Register(
		okStage("validate", []string{"op"}, 100),
		okStage("enrich", []string{"op"}, 200),
		okStage("gl.post", []string{"op"}, 600),
	)

	pb := NewPipelineBuilder(r, nil)
	opCtx := opCtxForOp("op")
	opCtx.Suspended = true
	opCtx.ResumePoint = "enrich"

	require.NoError(t, pb.RunFrom(opCtx, "enrich"))
	assert.False(t, opCtx.Suspended)

	names := loggedStageNames(opCtx)
	assert.NotContains(t, names, "validate", "stages before resume point must not re-run")
	assert.Contains(t, names, "enrich")
	assert.Contains(t, names, "gl.post")
}

// ── TxHooks ───────────────────────────────────────────────────────────────────

func TestRun_TxHooks_FiredAfterStages(t *testing.T) {
	r := NewStageRegistry()
	r.Register(okStage("a", []string{"op"}, 100))

	hookCalled := false
	txRunner := &stubTxRunner{}

	pb := NewPipelineBuilder(r, txRunner)
	opCtx := opCtxForOp("op")
	opCtx.RegisterTxHook(TxHook{
		Name:     "domain_events.insert",
		Priority: 10,
		Fn: func(_ context.Context, _ db.Store, _ *OperationContext) error {
			hookCalled = true
			return nil
		},
	})

	require.NoError(t, pb.Run(opCtx))
	assert.True(t, txRunner.called)
	assert.True(t, hookCalled)
}

func TestRun_TxHooks_NoTxRunner_Skipped(t *testing.T) {
	r := NewStageRegistry()
	pb := NewPipelineBuilder(r, nil)
	opCtx := opCtxForOp("op")
	opCtx.RegisterTxHook(TxHook{Name: "hook", Priority: 1, Fn: func(_ context.Context, _ db.Store, _ *OperationContext) error {
		return nil
	}})

	require.NoError(t, pb.Run(opCtx))
	// Skipped hook logged
	assert.Equal(t, "skipped", opCtx.Log[0].Status)
}

func TestRun_TxHooks_FiredInPriorityOrder(t *testing.T) {
	r := NewStageRegistry()
	txRunner := &stubTxRunner{}
	pb := NewPipelineBuilder(r, txRunner)
	opCtx := opCtxForOp("op")

	var order []string
	for _, h := range []struct {
		name     string
		priority int
	}{{"c", 30}, {"a", 10}, {"b", 20}} {
		n := h.name
		opCtx.RegisterTxHook(TxHook{
			Name:     n,
			Priority: h.priority,
			Fn: func(_ context.Context, _ db.Store, _ *OperationContext) error {
				order = append(order, n)
				return nil
			},
		})
	}

	require.NoError(t, pb.Run(opCtx))
	assert.Equal(t, []string{"a", "b", "c"}, order)
}

func TestRun_TxHook_Error_ReturnedToCallerAndLogged(t *testing.T) {
	r := NewStageRegistry()
	txRunner := &stubTxRunner{}
	pb := NewPipelineBuilder(r, txRunner)
	opCtx := opCtxForOp("op")
	opCtx.RegisterTxHook(TxHook{
		Name:     "failing_hook",
		Priority: 1,
		Fn: func(_ context.Context, _ db.Store, _ *OperationContext) error {
			return errors.New("outbox insert failed")
		},
	})

	err := pb.Run(opCtx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "outbox insert failed")
}

// ── DryRun ────────────────────────────────────────────────────────────────────

func TestRun_DryRun_SimulatesSimulatableStages(t *testing.T) {
	sim := &testSimulatableStage{
		testStage: testStage{
			BaseStage: BaseStage{StageName: "gl.post", StageOperations: []string{"op"}, StagePriority: 600},
		},
		simResult: StageResult{Status: "simulated", Message: "would post"},
	}
	r := NewStageRegistry()
	r.Register(sim)

	pb := NewPipelineBuilder(r, nil)
	opCtx := opCtxForOp("op")
	opCtx.DryRun = true

	require.NoError(t, pb.Run(opCtx))
	require.Len(t, opCtx.Log, 1)
	assert.Equal(t, "simulated", opCtx.Log[0].Status)
}

func TestRun_DryRun_NonSimulatableStageSkipped(t *testing.T) {
	r := NewStageRegistry()
	r.Register(okStage("plain", []string{"op"}, 100))

	pb := NewPipelineBuilder(r, nil)
	opCtx := opCtxForOp("op")
	opCtx.DryRun = true

	require.NoError(t, pb.Run(opCtx))
	require.Len(t, opCtx.Log, 1)
	assert.Equal(t, "skipped", opCtx.Log[0].Status)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func loggedStageNames(opCtx *OperationContext) []string {
	names := make([]string, len(opCtx.Log))
	for i, l := range opCtx.Log {
		names[i] = l.StageName
	}
	return names
}

// suspendingStage is a Stage whose Execute suspends opCtx.
type suspendingStage struct {
	name     string
	ops      []string
	priority int
}

func (s *suspendingStage) Name() string         { return s.name }
func (s *suspendingStage) Operations() []string { return s.ops }
func (s *suspendingStage) FeatureFlag() string  { return "" }
func (s *suspendingStage) Priority() int        { return s.priority }
func (s *suspendingStage) Required() bool       { return false }
func (s *suspendingStage) RunCondition() string { return "" }
func (s *suspendingStage) DependsOn() []string  { return nil }
func (s *suspendingStage) Execute(opCtx *OperationContext) (StageResult, error) {
	opCtx.Suspend("pending_workflow_approval:test-id", "gl.post")
	return StageResult{Status: "suspended", Message: "awaiting approval"}, nil
}
