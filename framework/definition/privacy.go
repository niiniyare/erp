package definition

import (
	"context"
	"errors"
)

// ErrDeny is returned by a PolicyFunc to deny access.
// The framework converts this to a 403 Forbidden (or 404 for reads, to prevent
// information leakage about record existence).
var ErrDeny = errors.New("access denied")

// ErrAllow is returned by a PolicyFunc to short-circuit the policy chain and
// grant access unconditionally, skipping remaining policies.
var ErrAllow = errors.New("access allowed")

// ErrSkip is returned by a PolicyFunc to indicate this policy makes no
// decision; evaluation continues to the next policy in the chain.
var ErrSkip = errors.New("policy skipped")

// ViewerContext carries the authenticated principal's identity and role
// information for use by privacy policies.
type ViewerContext interface {
	// ActorID returns the authenticated user ID (empty for unauthenticated requests).
	ActorID() string

	// TenantID returns the active tenant scope.
	TenantID() string

	// HasRole reports whether the actor holds the given role within the tenant.
	HasRole(role string) bool

	// IsSystem reports whether the operation is a system-initiated call
	// (e.g. Temporal workflow, background job) that bypasses user-facing policies.
	IsSystem() bool
}

// PolicyFunc is evaluated to determine whether the viewer may perform op on record.
//
// Return values:
//   - ErrAllow  — grant immediately, skip remaining policies
//   - ErrDeny   — deny immediately, skip remaining policies
//   - ErrSkip   — this policy abstains; continue to next
//   - any other error — treated as ErrDeny with the error logged
//
// The framework is fail-closed: if no policy returns ErrAllow, access is denied.
type PolicyFunc func(ctx context.Context, viewer ViewerContext, op Op, record Record) error

// PolicyDef binds a PolicyFunc to a set of operations.
type PolicyDef struct {
	// Ops is the bitmask of operations this policy governs.
	Ops Op

	// Fn is the policy implementation.
	Fn PolicyFunc
}

// Policy creates a PolicyDef for the given operations.
func Policy(ops Op, fn PolicyFunc) PolicyDef {
	return PolicyDef{Ops: ops, Fn: fn}
}

// AllowAll is a PolicyFunc that grants access to everyone.
// Useful as a baseline for public entities (e.g. timezone lists).
func AllowAll(_ context.Context, _ ViewerContext, _ Op, _ Record) error {
	return ErrAllow
}

// DenyAll is a PolicyFunc that denies access to everyone.
// Useful as a safe placeholder during development.
func DenyAll(_ context.Context, _ ViewerContext, _ Op, _ Record) error {
	return ErrDeny
}

// AllowSystem is a PolicyFunc that allows only system-initiated operations.
func AllowSystem(_ context.Context, viewer ViewerContext, _ Op, _ Record) error {
	if viewer.IsSystem() {
		return ErrAllow
	}
	return ErrSkip
}
