package pipeline

import (
	"context"
	"fmt"
	"sort"
	"time"

	db "awo.so/db/sqlc"
)

// TxRunner abstracts running work inside a tenant-scoped database transaction.
// The pipeline fires TxHooks through this interface so it has no dependency on
// a concrete pool or driver.
//
// Implementations must call set_tenant_context() before executing fn — the
// db.Store.WithTenantFromCtx method satisfies this contract out of the box.
type TxRunner interface {
	RunInTx(ctx context.Context, fn func(context.Context, db.Store) error) error
}

// Compensatable is an optional extension to Stage for stages with side effects
// that can be reversed. When the pipeline aborts (a required stage fails), it
// calls Compensate on completed stages in LIFO order.
type Compensatable interface {
	Compensate(opCtx *OperationContext, output map[string]any) error
}

// PipelineBuilder assembles and runs a pipeline for a single operation.
// Create one per application startup via NewPipelineBuilder and reuse it.
type PipelineBuilder struct {
	registry *StageRegistry
	txRunner TxRunner // nil → TxHooks are logged as warnings and skipped
}

// NewPipelineBuilder creates a PipelineBuilder backed by the given registry.
// txRunner may be nil when TxHook support is not required.
func NewPipelineBuilder(registry *StageRegistry, txRunner TxRunner) *PipelineBuilder {
	return &PipelineBuilder{
		registry: registry,
		txRunner: txRunner,
	}
}

// Run executes the pipeline for opCtx.OperationKey.
//
// Lifecycle:
//  1. Collect stages from the registry that match the operation key.
//  2. Filter stages whose FeatureFlag is off (per opCtx session).
//  3. Sort by Priority (ascending); within the same priority band,
//     stages without dependencies on each other are run sequentially
//     (parallel execution is a future enhancement).
//  4. For each stage:
//     a. If DryRun and stage implements Simulatable → call Simulate.
//     b. Otherwise call Execute.
//     c. If the stage returns an error and Required() == true → abort,
//        compensate completed stages in LIFO order, return the error.
//     d. If Required() == false → log the error and continue.
//     e. Record a stageCheckpoint for compensation tracking.
//     f. Write Outputs back to opCtx.Data.
//  5. If opCtx.Suspended after any stage → return nil (caller must persist
//     opCtx and resume later via RunFrom).
//  6. Fire registered TxHooks inside a database transaction (if txRunner set).
func (pb *PipelineBuilder) Run(opCtx *OperationContext) error {
	stages := pb.collect(opCtx)
	return pb.execute(opCtx, stages, 0)
}

// RunFrom resumes a suspended pipeline from the named stage.
// Stages before resumeFromStage are skipped; all subsequent stages run normally.
func (pb *PipelineBuilder) RunFrom(opCtx *OperationContext, resumeFromStage string) error {
	stages := pb.collect(opCtx)

	startIdx := 0
	for i, s := range stages {
		if s.Name() == resumeFromStage {
			startIdx = i
			break
		}
	}

	// Clear suspension so the resumed run proceeds normally
	opCtx.Suspended = false
	opCtx.SuspendReason = ""
	opCtx.ResumePoint = ""

	return pb.execute(opCtx, stages, startIdx)
}

// ── internals ─────────────────────────────────────────────────────────────────

// collect fetches and filters stages for the current operation.
func (pb *PipelineBuilder) collect(opCtx *OperationContext) []Stage {
	candidates := pb.registry.ForOperation(opCtx.OperationKey)

	// Filter out feature-flagged stages that are disabled for this tenant
	active := candidates[:0]
	for _, s := range candidates {
		if flag := s.FeatureFlag(); flag != "" && !opCtx.FeatureEnabled(flag) {
			continue
		}
		active = append(active, s)
	}

	// Sort by priority ascending so lower numbers run first
	sort.Slice(active, func(i, j int) bool {
		return active[i].Priority() < active[j].Priority()
	})

	return active
}

// execute runs stages[startIdx:] and fires TxHooks on success.
func (pb *PipelineBuilder) execute(opCtx *OperationContext, stages []Stage, startIdx int) error {
	var checkpoints []stageCheckpoint

	for i := startIdx; i < len(stages); i++ {
		s := stages[i]
		started := time.Now()

		result, err := pb.runStage(opCtx, s)

		entry := StageLog{
			StageName: s.Name(),
			Status:    result.Status,
			Message:   result.Message,
			StartedAt: started,
			Duration:  time.Since(started),
			Output:    result.Outputs,
		}

		if err != nil {
			entry.Status = "failed"
			entry.Error = err.Error()
			opCtx.Log = append(opCtx.Log, entry)

			if s.Required() {
				// Abort: compensate completed stages in LIFO order
				pb.compensate(opCtx, checkpoints)
				return fmt.Errorf("pipeline: required stage %q failed: %w", s.Name(), err)
			}
			// Non-required failure: log and continue
			continue
		}

		opCtx.Log = append(opCtx.Log, entry)

		// Write stage outputs into shared data map
		for k, v := range result.Outputs {
			opCtx.Data[k] = v
		}

		// Record checkpoint for potential compensation
		checkpoints = append(checkpoints, stageCheckpoint{
			StageName: s.Name(),
			Output:    result.Outputs,
		})

		// Suspension: stop executing; caller must persist and resume later
		if opCtx.Suspended {
			return nil
		}

		// Branch jump: skip ahead to the named stage if set
		if result.NextStageID != "" {
			for j := i + 1; j < len(stages); j++ {
				if stages[j].Name() == result.NextStageID {
					i = j - 1 // loop will increment
					break
				}
			}
		}
	}

	// All stages completed — fire TxHooks inside a DB transaction
	return pb.fireTxHooks(opCtx)
}

// runStage dispatches to Simulate or Execute based on DryRun mode.
func (pb *PipelineBuilder) runStage(opCtx *OperationContext, s Stage) (StageResult, error) {
	if opCtx.DryRun {
		if sim, ok := s.(Simulatable); ok {
			return sim.Simulate(opCtx)
		}
		// Stage does not support simulation: return a synthetic skipped result
		return StageResult{
			Status:  "skipped",
			Message: fmt.Sprintf("stage %q does not implement Simulatable; skipped in dry-run", s.Name()),
		}, nil
	}
	return s.Execute(opCtx)
}

// compensate calls Compensate on completed stages in LIFO order.
// Compensation errors are logged but do not prevent other compensations.
func (pb *PipelineBuilder) compensate(opCtx *OperationContext, checkpoints []stageCheckpoint) {
	for i := len(checkpoints) - 1; i >= 0; i-- {
		cp := checkpoints[i]
		s, exists := pb.registry.stages[cp.StageName]
		if !exists {
			continue
		}
		comp, ok := s.(Compensatable)
		if !ok {
			continue
		}
		if err := comp.Compensate(opCtx, cp.Output); err != nil {
			opCtx.Log = append(opCtx.Log, StageLog{
				StageName: cp.StageName,
				Status:    "compensation_failed",
				Error:     err.Error(),
				StartedAt: time.Now(),
			})
		} else {
			opCtx.Log = append(opCtx.Log, StageLog{
				StageName: cp.StageName,
				Status:    "compensated",
				StartedAt: time.Now(),
			})
		}
	}
}

// fireTxHooks runs all registered TxHooks inside a single DB transaction.
// Hooks are sorted by Priority (ascending) before execution.
// If txRunner is nil and there are hooks, they are skipped with a log entry.
func (pb *PipelineBuilder) fireTxHooks(opCtx *OperationContext) error {
	if len(opCtx.TxHooks) == 0 {
		return nil
	}

	if pb.txRunner == nil {
		opCtx.Log = append(opCtx.Log, StageLog{
			StageName: "pipeline.tx_hooks",
			Status:    "skipped",
			Message:   fmt.Sprintf("%d TxHook(s) registered but no TxRunner provided", len(opCtx.TxHooks)),
			StartedAt: time.Now(),
		})
		return nil
	}

	// Sort hooks by priority ascending
	hooks := make([]TxHook, len(opCtx.TxHooks))
	copy(hooks, opCtx.TxHooks)
	sort.Slice(hooks, func(i, j int) bool {
		return hooks[i].Priority < hooks[j].Priority
	})

	return pb.txRunner.RunInTx(opCtx.Ctx, func(ctx context.Context, store db.Store) error {
		for _, h := range hooks {
			started := time.Now()
			if err := h.Fn(ctx, store, opCtx); err != nil {
				opCtx.Log = append(opCtx.Log, StageLog{
					StageName: "tx_hook." + h.Name,
					Status:    "failed",
					Error:     err.Error(),
					StartedAt: started,
					Duration:  time.Since(started),
				})
				return fmt.Errorf("pipeline: tx hook %q failed: %w", h.Name, err)
			}
			opCtx.Log = append(opCtx.Log, StageLog{
				StageName: "tx_hook." + h.Name,
				Status:    "completed",
				StartedAt: started,
				Duration:  time.Since(started),
			})
		}
		return nil
	})
}
