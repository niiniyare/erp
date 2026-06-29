package definition

import "context"

// Op is the mutation operation that triggered a hook.
type Op uint8

const (
	OpRead Op = 1 << iota
	OpCreate
	OpUpdate
	OpDelete
)

// OpAll is a convenience mask matching any operation.
const OpAll = OpRead | OpCreate | OpUpdate | OpDelete

// OpWrite matches create, update, delete (not read).
const OpWrite = OpCreate | OpUpdate | OpDelete

// Is reports whether op includes the given operation flag.
func (o Op) Is(flag Op) bool { return o&flag != 0 }

func (o Op) String() string {
	switch o {
	case OpRead:
		return "read"
	case OpCreate:
		return "create"
	case OpUpdate:
		return "update"
	case OpDelete:
		return "delete"
	default:
		return "unknown"
	}
}

// Mutation carries the full context of an in-flight write operation.
// Hooks receive a *Mutation and may read or modify Before/After.
type Mutation struct {
	// Op is the operation being performed.
	Op Op

	// Before is the record state before the operation (nil on create).
	Before Record

	// After is the record state after the operation (nil on delete).
	// Hooks may mutate the underlying map to modify fields before persistence.
	After MutableRecord

	// TenantID is the tenant context for this operation.
	// Available to hooks that need tenant-scoped lookups or outbox writes.
	TenantID string

	// ActorID is the authenticated user performing the operation (may be empty for system ops).
	ActorID string
}

// MutableRecord extends Record with write access for before/after hooks.
type MutableRecord interface {
	Record

	// Set updates the value of the named field.
	Set(field string, value any)
}

// HookFunc is a lifecycle hook invoked within the write transaction.
//
// Rules:
//   - No external I/O (no HTTP calls, no message publishing) — use after_commit hooks for that.
//   - after_save errors roll back the entire transaction.
//   - Hooks are invoked in registration order.
type HookFunc func(ctx context.Context, m *Mutation) error

// HookTiming identifies when in the lifecycle a hook fires.
type HookTiming uint8

func (t HookTiming) String() string {
	switch t {
	case HookBeforeValidate:
		return "before_validate"
	case HookBeforeSave:
		return "before_save"
	case HookAfterSave:
		return "after_save"
	case HookBeforeDelete:
		return "before_delete"
	case HookOnSubmit:
		return "on_submit"
	case HookOnCancel:
		return "on_cancel"
	default:
		return "unknown"
	}
}

const (
	// HookBeforeValidate fires before field validation, outside the DB transaction.
	// Use to normalise input or compute derived fields that validators depend on.
	HookBeforeValidate HookTiming = iota + 1

	// HookBeforeSave fires after validation but before the DB write, outside the transaction.
	// Use to enforce business rules that require the fully validated record.
	HookBeforeSave

	// HookAfterSave fires after the DB write, still inside the transaction.
	// Errors cause a full rollback. Keep fast — no external I/O.
	HookAfterSave

	// HookBeforeDelete fires before a delete, outside the transaction.
	// Use to guard deletion and enforce referential integrity the DB cannot.
	HookBeforeDelete

	// HookOnSubmit fires when an entity's status transitions to its "submitted" state.
	// Runs inside the transaction. Use for GL postings, outbox writes, naming series lock.
	HookOnSubmit

	// HookOnCancel fires when a submitted document is cancelled.
	// Use to reverse GL postings and trigger compensating workflows.
	HookOnCancel
)

// HookDef binds a named HookFunc to a set of operations and a lifecycle timing.
type HookDef struct {
	// Name identifies this hook in logs and error messages.
	Name string

	// Ops is the bitmask of operations this hook fires for.
	// Defaults to OpWrite (create | update | delete) when zero.
	Ops Op

	// Timing determines when in the lifecycle the hook fires.
	// Defaults to HookBeforeSave when zero.
	Timing HookTiming

	// Fn is the hook implementation.
	Fn HookFunc
}

// effectiveTiming returns Timing defaulted to HookBeforeSave when zero.
func (h HookDef) EffectiveTiming() HookTiming {
	if h.Timing == 0 {
		return HookBeforeSave
	}
	return h.Timing
}

// effectiveOps returns Ops defaulted to OpWrite when zero.
func (h HookDef) EffectiveOps() Op {
	if h.Ops == 0 {
		return OpWrite
	}
	return h.Ops
}

// IsBefore reports whether the hook fires before persistence (before_save, before_validate, before_delete).
func (h HookDef) IsBefore() bool {
	switch h.EffectiveTiming() {
	case HookBeforeValidate, HookBeforeSave, HookBeforeDelete:
		return true
	}
	return false
}

// ── Constructor helpers ────────────────────────────────────────────────────────

// BeforeHook creates a HookBeforeSave for the given ops. Legacy alias.
func BeforeHook(ops Op, fn HookFunc) HookDef {
	return HookDef{Ops: ops, Timing: HookBeforeSave, Fn: fn}
}

// AfterHook creates a HookAfterSave for the given ops. Legacy alias.
func AfterHook(ops Op, fn HookFunc) HookDef {
	return HookDef{Ops: ops, Timing: HookAfterSave, Fn: fn}
}

// BeforeValidateHook creates a named HookBeforeValidate for the given ops.
func BeforeValidateHook(name string, ops Op, fn HookFunc) HookDef {
	return HookDef{Name: name, Ops: ops, Timing: HookBeforeValidate, Fn: fn}
}

// BeforeSaveHook creates a named HookBeforeSave for the given ops.
func BeforeSaveHook(name string, ops Op, fn HookFunc) HookDef {
	return HookDef{Name: name, Ops: ops, Timing: HookBeforeSave, Fn: fn}
}

// AfterSaveHook creates a named HookAfterSave for the given ops.
func AfterSaveHook(name string, ops Op, fn HookFunc) HookDef {
	return HookDef{Name: name, Ops: ops, Timing: HookAfterSave, Fn: fn}
}

// BeforeDeleteHook creates a named HookBeforeDelete (ops forced to OpDelete).
func BeforeDeleteHook(name string, fn HookFunc) HookDef {
	return HookDef{Name: name, Ops: OpDelete, Timing: HookBeforeDelete, Fn: fn}
}

// OnSubmitHook creates a named HookOnSubmit (ops forced to OpUpdate).
func OnSubmitHook(name string, fn HookFunc) HookDef {
	return HookDef{Name: name, Ops: OpUpdate, Timing: HookOnSubmit, Fn: fn}
}

// OnCancelHook creates a named HookOnCancel (ops forced to OpUpdate).
func OnCancelHook(name string, fn HookFunc) HookDef {
	return HookDef{Name: name, Ops: OpUpdate, Timing: HookOnCancel, Fn: fn}
}
