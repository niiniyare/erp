package stages

import (
	"fmt"
	"runtime/debug"

	"awo.so/internal/pipeline"
	"awo.so/internal/web/ast"
	"awo.so/internal/web/ui"
)

// ─── COMPILE STAGE ────────────────────────────────────────────────────────────
//
// CompileStage executes the PageFn (or ASTPageFn) with the UISessionContext to
// produce the compiled AMIS schema.
//
// Dispatch order:
//  1. ASTPageFn (DataKeyASTPageFn) — typed AST path.
//     fn(sess) returns ast.Node → ast.CompileTree(node) → Schema.
//     Validation happens inside CompileTree before JSON emission.
//     Sets DataKeyASTCompiled = true so NormalizeStage can skip redundant checks.
//
//  2. PageFn (DataKeyPageFn) — legacy path.
//     fn(sess) returns Schema directly.
//     NormalizeStage applies full structural rule set.
//
// Recovers from panics in both paths — a panicking function returns an error
// that aborts the pipeline with HTTP 500 but does not crash the server.

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
			StageDependsOn:  []string{"ui.registry"},
		},
	}
}

// Execute dispatches to ASTPageFn or PageFn and writes the compiled schema to
// DataKeySchema. Sets DataKeyASTCompiled when the AST path was used.
func (s *CompileStage) Execute(opCtx *pipeline.OperationContext) (pipeline.StageResult, error) {
	sess, ok := opCtx.Data[ui.DataKeySessionCtx].(ui.UISessionContext)
	if !ok {
		return pipeline.StageResult{}, fmt.Errorf("ui.compile: DataKeySessionCtx missing — AuthzStage must run first")
	}

	// ── AST path ──────────────────────────────────────────────────────────────
	if astFn, ok := opCtx.Data[ui.DataKeyASTPageFn].(ui.ASTPageFn); ok && astFn != nil {
		schema, err := safeCompileAST(astFn, sess)
		if err != nil {
			return pipeline.StageResult{}, fmt.Errorf("ui.compile: ASTPageFn failed: %w", err)
		}
		if len(schema) == 0 {
			return pipeline.StageResult{}, fmt.Errorf("ui.compile: ASTPageFn produced empty schema")
		}
		return pipeline.StageResult{
			Status:  "completed",
			Message: fmt.Sprintf("AST-compiled schema with %d top-level keys", len(schema)),
			Outputs: map[string]any{
				ui.DataKeySchema:      schema,
				ui.DataKeyASTCompiled: true,
			},
		}, nil
	}

	// ── Legacy PageFn path ────────────────────────────────────────────────────
	pageFn, ok := opCtx.Data[ui.DataKeyPageFn].(ui.PageFn)
	if !ok || pageFn == nil {
		return pipeline.StageResult{}, fmt.Errorf("ui.compile: neither DataKeyASTPageFn nor DataKeyPageFn found — RegistryStage must run first")
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

// safeCompileAST calls the ASTPageFn, asserts the return value to ast.Node,
// runs CompileTree, and recovers any panic.
func safeCompileAST(fn ui.ASTPageFn, sess ui.UISessionContext) (schema ui.Schema, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v\nstack:\n%s", r, debug.Stack())
		}
	}()

	raw := fn(sess)
	if raw == nil {
		return nil, fmt.Errorf("ASTPageFn returned nil")
	}

	node, ok := raw.(ast.Node)
	if !ok {
		return nil, fmt.Errorf("ASTPageFn return value does not implement ast.Node (got %T)", raw)
	}

	return ast.CompileTree(node)
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
