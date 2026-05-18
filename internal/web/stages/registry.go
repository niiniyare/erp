package stages

import (
	"fmt"

	"awo.so/internal/pipeline"
	"awo.so/internal/web/registry"
	"awo.so/internal/web/ui"
)

// ─── TASK 4 — REGISTRY STAGE ─────────────────────────────────────────────────
//
// RegistryStage resolves the PageFn for the requested route from the page registry.
// If no PageFn is registered for the route, the pipeline aborts with ErrPageNotFound.
//
// DESCRIPTION:
// Reads UISchemaInput.Route from opCtx.Input. Calls registry.Get(route).
// If nil, returns PageNotFoundError which SchemaHandler maps to HTTP 404.
// If found, writes the PageFn to DataKeyPageFn for CompileStage.
//
// WHY:
// Separating registry resolution from compilation makes each step independently
// testable. A 404 from this stage is a clean client error — not an internal error.
//
// RISKS:
// If the route has a trailing slash mismatch ("/finance/invoices/" vs "/finance/invoices"),
// the lookup returns nil. Normalise routes before calling registry.Register.

// RegistryStage is Priority 40, Required true.
type RegistryStage struct {
	pipeline.BaseStage
}

// NewRegistryStage constructs a RegistryStage.
func NewRegistryStage() *RegistryStage {
	return &RegistryStage{
		BaseStage: pipeline.BaseStage{
			StageName:       "ui.registry",
			StageOperations: []string{ui.OperationKey},
			StagePriority:   ui.PriorityRegistry,
			StageRequired:   true,
			StageDependsOn:  []string{"ui.authz"},
		},
	}
}

// Execute resolves the PageFn from the registry. Aborts on miss.
func (s *RegistryStage) Execute(opCtx *pipeline.OperationContext) (pipeline.StageResult, error) {
	input, ok := opCtx.Input.(ui.UISchemaInput)
	if !ok || input.Route == "" {
		return pipeline.StageResult{}, fmt.Errorf("ui.registry: no UISchemaInput in opCtx.Input")
	}

	fn := registry.Get(input.Route)
	if fn == nil {
		return pipeline.StageResult{}, &ui.PageNotFoundError{Route: input.Route}
	}

	return pipeline.StageResult{
		Status:  "completed",
		Message: fmt.Sprintf("resolved PageFn for route %s", input.Route),
		Outputs: map[string]any{
			ui.DataKeyPageFn: fn,
		},
	}, nil
}

var _ pipeline.Stage = (*RegistryStage)(nil)
