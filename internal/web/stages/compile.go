package stages

import (
	"fmt"
	"runtime/debug"

	"awo.so/internal/pipeline"
	"awo.so/internal/web/ui"
)

// ─── TASK 5 — COMPILE STAGE ──────────────────────────────────────────────────
//
// CompileStage executes the PageFn with the UISessionContext to produce the
// AMIS schema. This is the only stage that calls PageFn.
//
// DESCRIPTION:
// Reads DataKeyPageFn (ui.PageFn) and DataKeySessionCtx (ui.UISessionContext)
// from opCtx.Data. Calls fn(sess) and writes the result to DataKeySchema.
// Recovers from panics in PageFn — a panicking page function returns an error
// that aborts the pipeline with HTTP 500, but does not crash the server.
//
// WHY:
// PageFn is user-land code. Isolating its execution in a recoverable stage
// prevents a buggy page function from taking down the process.
//
// INVARIANTS enforced by this stage (not checked here — NormalizeStage checks):
//   - PageFn must return a non-nil, non-empty Schema.
//   - PageFn must be pure: same UISessionContext → same Schema.
//
// RISKS:
// Panics are recovered but the underlying PageFn bug is not fixed — it will
// panic on every cache miss until fixed. Log the stack trace at ERROR level.

// CompileStage is Priority 50, Required true.
type CompileStage struct {
	pipeline.BaseStage
}

// NewCompileStage constructs a CompileStage.
func NewCompileStage() *CompileStage {
	return &CompileStage{
		BaseStage: pipeline.BaseStage{
			StageName:       "ui.compile",
			StageOperations: []string{ui.OperationKey},
			StagePriority:   ui.PriorityCompile,
			StageRequired:   true,
		},
	}
}

// Execute calls the PageFn with the UISessionContext. Recovers panics.
func (s *CompileStage) Execute(opCtx *pipeline.OperationContext) (pipeline.StageResult, error) {
	pageFn, ok := opCtx.Data[ui.DataKeyPageFn].(ui.PageFn)
	if !ok || pageFn == nil {
		return pipeline.StageResult{}, fmt.Errorf("ui.compile: DataKeyPageFn missing or wrong type — RegistryStage must run first")
	}

	sess, ok := opCtx.Data[ui.DataKeySessionCtx].(ui.UISessionContext)
	if !ok {
		return pipeline.StageResult{}, fmt.Errorf("ui.compile: DataKeySessionCtx missing — AuthzStage must run first")
	}

	schema, err := safeCompile(pageFn, sess)
	if err != nil {
		return pipeline.StageResult{}, fmt.Errorf("ui.compile: PageFn panicked: %w", err)
	}

	if len(schema) == 0 {
		return pipeline.StageResult{}, fmt.Errorf("ui.compile: PageFn returned empty schema")
	}

	return pipeline.StageResult{
		Status:  "completed",
		Message: fmt.Sprintf("compiled schema with %d top-level keys", len(schema)),
		Outputs: map[string]any{
			ui.DataKeySchema: schema,
		},
	}, nil
}

// safeCompile calls fn(sess) and recovers any panic into an error.
func safeCompile(fn ui.PageFn, sess ui.UISessionContext) (schema ui.Schema, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v\nstack:\n%s", r, debug.Stack())
		}
	}()
	schema = fn(sess)
	return schema, nil
}

var _ pipeline.Stage = (*CompileStage)(nil)
