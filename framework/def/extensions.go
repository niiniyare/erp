package def

import "context"

// ── Workflow triggers ──────────────────────────────────────────────────────────

// WorkflowTrigger binds a Temporal workflow type to a set of mutation operations.
// Declared on EntityDefinition.WorkflowTriggers; registered automatically at
// bootstrap time when a Temporal client is configured.
//
// The trigger fires via HookAfterSave so the entity row is committed before the
// workflow reads it. Temporal unavailability must NOT roll back the entity save —
// callers should use an outbox or accept at-least-once delivery semantics.
type WorkflowTrigger struct {
	// Ops is the bitmask of operations that launch this workflow.
	// Example: OpCreate | OpUpdate
	Ops Op

	// WorkflowType is the Temporal workflow type name passed to
	// client.ExecuteWorkflow. Must match the workflow registered on the worker.
	WorkflowType string

	// TaskQueue overrides the default task queue for this trigger.
	// Empty string uses the application-wide default task queue.
	TaskQueue string

	// IDFunc generates the workflow ID from the mutation context.
	// Defaults to "{tenant}.{entity}.{record-id}.{op}" when nil.
	// Provide a custom function to deduplicate or control workflow identity.
	IDFunc func(m *Mutation) string
}

// ── Custom actions ─────────────────────────────────────────────────────────────

// ActionDef declares a custom action button in the SDUI and a corresponding
// POST route registered at /api/{entity}/{id}/{Name}.
//
// Actions are distinct from hooks: they are explicitly invoked by the user
// (e.g. "Approve", "Submit", "Cancel") rather than firing automatically on CRUD.
type ActionDef struct {
	// Name is the machine identifier used in URLs and hook dispatch.
	// Convention: snake_case verb, e.g. "submit", "approve", "cancel".
	Name string

	// Label is the display label for the action button in SDUI.
	Label string

	// Icon is an optional icon name (amis icon class) for the action button.
	Icon string

	// ConfirmMessage is shown in a confirmation dialog before the action fires.
	// Empty string skips the confirmation dialog.
	ConfirmMessage string

	// Level controls the button visual style: "primary", "success", "warning", "danger", "link".
	// Defaults to "default" when empty.
	Level string

	// Ops restricts which ops trigger this action hook. Defaults to OpUpdate.
	Ops Op

	// Fn is the action handler invoked when the action route is called.
	// It receives the Mutation with Op = OpUpdate and must return nil on success.
	Fn func(ctx context.Context, m *Mutation) error
}

// ── Page builders ──────────────────────────────────────────────────────────────

// PageBuilderSet allows module code to override the default AMIS schema
// generators produced by the sdui/amis package for a specific EntityDefinition.
//
// Set only the fields you want to override; nil fields fall back to the
// default generators. Useful for entities with complex layouts, embedded
// sub-tables, or custom toolbar actions beyond the standard CRUD controls.
type PageBuilderSet struct {
	// CRUDPage overrides the default CRUD list+form page schema generator.
	// Return value must be a valid amis page schema (map[string]any).
	CRUDPage func(def *EntityDefinition, apiBase string) map[string]any

	// FormPage overrides the default create/edit form schema generator.
	FormPage func(def *EntityDefinition, apiBase string) map[string]any

	// DetailPage overrides the read-only record detail page schema generator.
	// When nil, the framework does not register a detail-page SDUI route.
	DetailPage func(def *EntityDefinition, apiBase string) map[string]any
}
