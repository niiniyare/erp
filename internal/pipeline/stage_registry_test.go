package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func makeStage(name string, ops []string, priority int) Stage {
	return &testStage{
		BaseStage: BaseStage{
			StageName:       name,
			StageOperations: ops,
			StagePriority:   priority,
		},
	}
}

// ── Register ──────────────────────────────────────────────────────────────────

func TestStageRegistry_Register_Single(t *testing.T) {
	r := NewStageRegistry()
	s := makeStage("ap.validate_invoice", []string{"ap.invoice.process"}, 110)

	require.NotPanics(t, func() { r.Register(s) })
	assert.Len(t, r.All(), 1)
}

func TestStageRegistry_Register_Multiple(t *testing.T) {
	r := NewStageRegistry()
	r.Register(
		makeStage("ap.validate_invoice", []string{"ap.invoice.process"}, 110),
		makeStage("budget.check", []string{"ap.invoice.process"}, 310),
		makeStage("audit.log", []string{"*"}, 910),
	)
	assert.Len(t, r.All(), 3)
}

func TestStageRegistry_Register_DuplicateNamePanics(t *testing.T) {
	r := NewStageRegistry()
	r.Register(makeStage("ap.validate_invoice", []string{"ap.invoice.process"}, 110))

	assert.Panics(t, func() {
		r.Register(makeStage("ap.validate_invoice", []string{"ap.invoice.approve"}, 110))
	})
}

// ── ForOperation ──────────────────────────────────────────────────────────────

func TestStageRegistry_ForOperation_ExactMatch(t *testing.T) {
	r := NewStageRegistry()
	r.Register(
		makeStage("ap.validate_invoice", []string{"ap.invoice.process"}, 110),
		makeStage("budget.check", []string{"ap.invoice.process", "ap.invoice.approve"}, 310),
		makeStage("gl.post_transaction", []string{"gl.journal.post"}, 610),
	)

	result := r.ForOperation("ap.invoice.process")

	names := stageNames(result)
	assert.Contains(t, names, "ap.validate_invoice")
	assert.Contains(t, names, "budget.check")
	assert.NotContains(t, names, "gl.post_transaction")
}

func TestStageRegistry_ForOperation_Wildcard(t *testing.T) {
	r := NewStageRegistry()
	r.Register(
		makeStage("audit.log", []string{"*"}, 910),
		makeStage("ap.validate_invoice", []string{"ap.invoice.process"}, 110),
	)

	// Wildcard should appear for any operation key
	result := r.ForOperation("ap.invoice.process")
	assert.Contains(t, stageNames(result), "audit.log")

	result2 := r.ForOperation("gl.journal.post")
	assert.Contains(t, stageNames(result2), "audit.log")

	result3 := r.ForOperation("some.unknown.op")
	assert.Contains(t, stageNames(result3), "audit.log")
}

func TestStageRegistry_ForOperation_MultipleOpsOnStage(t *testing.T) {
	r := NewStageRegistry()
	r.Register(makeStage("budget.check", []string{"ap.invoice.process", "ap.invoice.approve"}, 310))

	assert.Len(t, r.ForOperation("ap.invoice.process"), 1)
	assert.Len(t, r.ForOperation("ap.invoice.approve"), 1)
	assert.Len(t, r.ForOperation("gl.journal.post"), 0)
}

func TestStageRegistry_ForOperation_UnknownKey_Empty(t *testing.T) {
	r := NewStageRegistry()
	r.Register(makeStage("ap.validate_invoice", []string{"ap.invoice.process"}, 110))

	result := r.ForOperation("nonexistent.operation")
	assert.Empty(t, result)
}

func TestStageRegistry_ForOperation_EmptyRegistry(t *testing.T) {
	r := NewStageRegistry()
	assert.Empty(t, r.ForOperation("ap.invoice.process"))
}

func TestStageRegistry_ForOperation_ReturnsCopy(t *testing.T) {
	r := NewStageRegistry()
	r.Register(makeStage("ap.validate_invoice", []string{"ap.invoice.process"}, 110))

	result1 := r.ForOperation("ap.invoice.process")
	result2 := r.ForOperation("ap.invoice.process")

	// Mutating one slice must not affect the other or the registry
	result1 = append(result1, makeStage("injected", []string{"*"}, 999))
	assert.Len(t, result2, 1, "ForOperation should return independent slices")
	assert.Len(t, r.All(), 1, "registry must not be mutated via returned slice")
}

// ── All ───────────────────────────────────────────────────────────────────────

func TestStageRegistry_All_ReturnsAllStages(t *testing.T) {
	r := NewStageRegistry()
	r.Register(
		makeStage("stage-a", []string{"op.a"}, 100),
		makeStage("stage-b", []string{"op.b"}, 200),
		makeStage("stage-c", []string{"*"}, 300),
	)

	all := r.All()
	assert.Len(t, all, 3)
	names := stageNames(all)
	assert.Contains(t, names, "stage-a")
	assert.Contains(t, names, "stage-b")
	assert.Contains(t, names, "stage-c")
}

func TestStageRegistry_All_EmptyRegistry(t *testing.T) {
	r := NewStageRegistry()
	assert.Empty(t, r.All())
}

// ── matchesOperation (internal helper) ───────────────────────────────────────

func TestMatchesOperation_ExactMatch(t *testing.T) {
	assert.True(t, matchesOperation([]string{"ap.invoice.process"}, "ap.invoice.process"))
}

func TestMatchesOperation_NoMatch(t *testing.T) {
	assert.False(t, matchesOperation([]string{"ap.invoice.process"}, "gl.journal.post"))
}

func TestMatchesOperation_Wildcard(t *testing.T) {
	assert.True(t, matchesOperation([]string{"*"}, "anything.at.all"))
}

func TestMatchesOperation_EmptyOps(t *testing.T) {
	assert.False(t, matchesOperation([]string{}, "ap.invoice.process"))
}

func TestMatchesOperation_MultipleOps(t *testing.T) {
	ops := []string{"ap.invoice.process", "ap.invoice.approve"}
	assert.True(t, matchesOperation(ops, "ap.invoice.approve"))
	assert.False(t, matchesOperation(ops, "gl.journal.post"))
}

// ── helpers ───────────────────────────────────────────────────────────────────

func stageNames(stages []Stage) []string {
	names := make([]string, len(stages))
	for i, s := range stages {
		names[i] = s.Name()
	}
	return names
}
