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

// HookDef binds a HookFunc to a set of operations and a timing.
type HookDef struct {
	// Ops is the bitmask of operations this hook fires for.
	Ops Op

	// Before indicates the hook runs before persistence (before_save).
	// If false, runs after persistence but still within the transaction (after_save).
	Before bool

	// Fn is the hook implementation.
	Fn HookFunc
}

// BeforeHook creates a HookDef that fires before persistence for the given ops.
func BeforeHook(ops Op, fn HookFunc) HookDef {
	return HookDef{Ops: ops, Before: true, Fn: fn}
}

// AfterHook creates a HookDef that fires after persistence (still in TX) for the given ops.
func AfterHook(ops Op, fn HookFunc) HookDef {
	return HookDef{Ops: ops, Before: false, Fn: fn}
}
