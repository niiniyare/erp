// Package workflow integrates EntityDefinition lifecycle hooks with Temporal.
// It provides a HookFunc factory that enqueues a Temporal workflow signal or
// start-workflow call whenever a matching mutation occurs, decoupling the
// write path from async processing.
package workflow

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/client"

	"awo.so/framework/def"
)

// TriggerKind controls how the workflow is launched.
type TriggerKind string

const (
	// TriggerSignal sends a signal to a running workflow identified by WorkflowID.
	// If no running workflow matches, the signal is dropped (use TriggerSignalWithStart
	// to avoid this).
	TriggerSignal TriggerKind = "signal"

	// TriggerSignalWithStart signals the workflow if running, or starts it if not.
	TriggerSignalWithStart TriggerKind = "signal_with_start"

	// TriggerStart always starts a new workflow execution.
	TriggerStart TriggerKind = "start"
)

// TriggerConfig describes one workflow trigger attached to an entity.
type TriggerConfig struct {
	// Ops is the bitmask of operations that fire this trigger.
	Ops def.Op

	// Kind controls how the Temporal client is called.
	Kind TriggerKind

	// WorkflowType is the registered Temporal workflow function name.
	WorkflowType string

	// TaskQueue is the Temporal task queue to route the workflow to.
	TaskQueue string

	// WorkflowIDFn derives the workflow ID from the mutation.
	// For TriggerSignal / TriggerSignalWithStart this identifies the target run.
	// For TriggerStart a unique run ID is appended automatically.
	// If nil, defaults to "<entity>/<record-id>/<workflow-type>".
	WorkflowIDFn func(m *def.Mutation) string

	// SignalName is the Temporal signal name (required for TriggerSignal and
	// TriggerSignalWithStart).
	SignalName string

	// PayloadFn builds the signal/workflow input from the mutation.
	// If nil, the entire mutation is passed as-is (serialised by Temporal's codec).
	PayloadFn func(m *def.Mutation) any
}

// Executor holds a Temporal client and issues workflow calls from hook invocations.
type Executor struct {
	client client.Client
}

// NewExecutor creates a workflow Executor backed by a Temporal client.
func NewExecutor(c client.Client) *Executor {
	return &Executor{client: c}
}

// Hook returns a def.HookFunc that fires the configured workflow trigger.
//
// The returned hook is intended for HookAfterCommit timing so Temporal calls
// execute AFTER the DB transaction commits. This guarantees:
//   - DB failure → no orphaned workflow started
//   - Temporal failure → DB data is preserved (hook error is logged, not fatal)
//
// Example usage:
//
//	def.AfterCommitHook("start_approval", def.OpCreate, executor.Hook(cfg))
func (e *Executor) Hook(cfg TriggerConfig) def.HookFunc {
	return func(ctx context.Context, m *def.Mutation) error {
		if !cfg.Ops.Is(m.Op) {
			return nil
		}

		workflowID := defaultWorkflowID(m, cfg)
		payload := buildPayload(m, cfg)

		switch cfg.Kind {
		case TriggerStart:
			opts := client.StartWorkflowOptions{
				ID:        workflowID,
				TaskQueue: cfg.TaskQueue,
			}
			_, err := e.client.ExecuteWorkflow(ctx, opts, cfg.WorkflowType, payload)
			if err != nil {
				return fmt.Errorf("workflow trigger start %s: %w", cfg.WorkflowType, err)
			}

		case TriggerSignal:
			err := e.client.SignalWorkflow(ctx, workflowID, "", cfg.SignalName, payload)
			if err != nil {
				return fmt.Errorf("workflow trigger signal %s/%s: %w", workflowID, cfg.SignalName, err)
			}

		case TriggerSignalWithStart:
			opts := client.StartWorkflowOptions{
				ID:        workflowID,
				TaskQueue: cfg.TaskQueue,
			}
			_, err := e.client.SignalWithStartWorkflow(ctx, workflowID, cfg.SignalName, payload, opts, cfg.WorkflowType, payload)
			if err != nil {
				return fmt.Errorf("workflow trigger signal-with-start %s/%s: %w", workflowID, cfg.SignalName, err)
			}

		default:
			return fmt.Errorf("workflow trigger: unknown kind %q", cfg.Kind)
		}

		return nil
	}
}

// ──────────────────────────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────────────────────────

func defaultWorkflowID(m *def.Mutation, cfg TriggerConfig) string {
	if cfg.WorkflowIDFn != nil {
		return cfg.WorkflowIDFn(m)
	}
	var rec def.Record = m.After
	if rec == nil {
		rec = m.Before
	}
	recordID := ""
	entityName := ""
	if rec != nil {
		recordID = rec.ID().String()
		entityName = rec.EntityName()
	}
	// Convention: "{tenantID}.{entity}.{recordID}.{op}"
	// Tenant prefix ensures uniqueness across tenants in a shared Temporal namespace.
	return fmt.Sprintf("%s.%s.%s.%s", m.TenantID, entityName, recordID, m.Op.String())
}

func buildPayload(m *def.Mutation, cfg TriggerConfig) any {
	if cfg.PayloadFn != nil {
		return cfg.PayloadFn(m)
	}
	return m
}
