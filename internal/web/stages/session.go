// Package stages contains all pipeline stages for UI schema compilation.
// Every stage operates exclusively on pipeline.OperationContext — no Fiber
// context, no direct IAM domain imports (except via contract).
package stages

import (
	"fmt"

	"awo.so/internal/core/iam/contract"
	"awo.so/internal/pipeline"
	"awo.so/internal/web/ui"
)

// ─── TASK 1 — SESSION STAGE ──────────────────────────────────────────────────
//
// SessionStage extracts and validates the contract.SessionContext from the
// Go context. It is the pipeline's first gate: if no authenticated session
// exists in the context, the pipeline aborts immediately.
//
// DESCRIPTION:
// Reads contract.SessionContext via contract.FromContext(opCtx.Ctx).
// The InjectSessionContext() Fiber middleware must have run before the handler
// builds opCtx to guarantee this key is present.
//
// WHY:
// Centralises the "is the caller authenticated?" check into one required stage.
// All downstream stages can safely call contract.FromContext() and assert ok==true.
//
// IMPLEMENTATION:
// Validates zero check on the returned SessionContext. Sets no Data — subsequent
// stages call contract.FromContext directly. SessionStage is the contract boundary
// guard: if it passes, the caller is authenticated.
//
// RISKS:
// If contract.InjectSessionContext() is absent from the route chain, FromContext
// returns false and this stage aborts with ErrUnauthenticated. This is correct
// behaviour — do not add fallback local extraction here.

// SessionStage is Priority 10, Required true.
// Aborts pipeline (returns ErrUnauthenticated) if no contract.SessionContext.
type SessionStage struct {
	pipeline.BaseStage
}

// NewSessionStage constructs a SessionStage.
func NewSessionStage() *SessionStage {
	return &SessionStage{
		BaseStage: pipeline.BaseStage{
			StageName:       "ui.session",
			StageOperations: []string{ui.OperationKey, ui.AppOperationKey},
			StagePriority:   ui.PrioritySession,
			StageRequired:   true,
		},
	}
}

// Execute validates that a contract.SessionContext is present in opCtx.Ctx.
// Returns ErrUnauthenticated (wrapped) if absent or zero.
func (s *SessionStage) Execute(opCtx *pipeline.OperationContext) (pipeline.StageResult, error) {
	sc, ok := contract.FromContext(opCtx.Ctx)
	if !ok || sc.IsZero() {
		return pipeline.StageResult{}, fmt.Errorf("%w: no session in context — ensure contract.InjectSessionContext() middleware is in the route chain", ui.ErrUnauthenticated)
	}

	return pipeline.StageResult{
		Status:  "completed",
		Message: fmt.Sprintf("session validated for user %s tenant %s", sc.UserID(), sc.TenantID()),
	}, nil
}

// Ensure compile-time conformance.
var _ pipeline.Stage = (*SessionStage)(nil)
