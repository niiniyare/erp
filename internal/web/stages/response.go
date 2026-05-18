package stages

import (
	"fmt"

	"awo.so/internal/pipeline"
	"awo.so/internal/web/ui"
)

// ─── TASK 8 — RESPONSE STAGE ─────────────────────────────────────────────────
//
// ResponseStage assembles the final UISchemaOutput from the pipeline data map.
// It is the terminal stage — always runs, whether the schema came from cache
// or was freshly compiled.
//
// DESCRIPTION:
// Reads DataKeySchema (set by CompileStage or CacheLookupStage on hit).
// Reads DataKeyCacheHit. Reads Input.Route.
// Constructs UISchemaOutput and writes it to DataKeyResponse.
// SchemaHandler reads DataKeyResponse after pipeline.Run() returns.
//
// WHY:
// The response stage decouples schema assembly from HTTP response writing.
// SchemaHandler is a thin adapter — it builds opCtx, runs the pipeline,
// reads DataKeyResponse, and calls c.JSON(). No schema logic in the handler.
//
// RISKS:
// If DataKeySchema is missing, the pipeline has an untracked bug in stage ordering.
// This stage returns an error in that case to surface the defect immediately.

// ResponseStage is Priority 90, Required true.
type ResponseStage struct {
	pipeline.BaseStage
}

// NewResponseStage constructs a ResponseStage.
func NewResponseStage() *ResponseStage {
	return &ResponseStage{
		BaseStage: pipeline.BaseStage{
			StageName:       "ui.response",
			StageOperations: []string{ui.OperationKey, ui.AppOperationKey},
			StagePriority:   ui.PriorityResponse,
			StageRequired:   true,
		},
	}
}

// Execute assembles UISchemaOutput and writes it to DataKeyResponse.
func (s *ResponseStage) Execute(opCtx *pipeline.OperationContext) (pipeline.StageResult, error) {
	schema, ok := opCtx.Data[ui.DataKeySchema].(ui.Schema)
	if !ok || len(schema) == 0 {
		return pipeline.StageResult{}, fmt.Errorf("ui.response: DataKeySchema missing or empty — pipeline stage ordering defect")
	}

	cacheHit, _ := opCtx.Data[ui.DataKeyCacheHit].(bool)

	var route string
	if input, ok := opCtx.Input.(ui.UISchemaInput); ok {
		route = input.Route
	}

	out := ui.UISchemaOutput{
		Schema:   schema,
		CacheHit: cacheHit,
		Route:    route,
	}

	return pipeline.StageResult{
		Status:  "completed",
		Message: fmt.Sprintf("response assembled for route %s (cacheHit=%v)", route, cacheHit),
		Outputs: map[string]any{
			ui.DataKeyResponse: out,
		},
	}, nil
}

var _ pipeline.Stage = (*ResponseStage)(nil)
