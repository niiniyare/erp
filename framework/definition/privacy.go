package definition

import (
	"context"
	"errors"

	"awo.so/framework/org"
)

// ErrDeny is returned by a PolicyFunc to deny access immediately.
// The framework converts this to:
//   - 403 Forbidden on write operations
//   - 404 Not Found on reads (to prevent existence probing)
var ErrDeny = errors.New("access denied")

// ErrAllow is returned by a PolicyFunc to grant access immediately,
// short-circuiting the remaining policy chain.
var ErrAllow = errors.New("access allowed")

// ErrSkip is returned by a PolicyFunc to abstain from the decision.
// Evaluation continues to the next policy in the chain.
var ErrSkip = errors.New("policy skipped")

// ViewerContext carries the authenticated principal's identity, org scope, and
// role memberships. Every framework handler receives one; privacy policies
// inspect it to decide whether to allow or deny an operation.
//
// # Org scope
//
// The viewer knows exactly where in the org hierarchy they are operating:
//
//	OrgScope().TenantID   — always set
//	OrgScope().CompanyID  — set when the request is company-scoped
//	OrgScope().DivisionID — set when the request is division-scoped
//
// A policy can call viewer.OrgScope().Contains(recordScope) to verify that the
// record the viewer is trying to read/write belongs to their scope.
type ViewerContext interface {
	// ActorID returns the authenticated user UUID.
	// Returns "anonymous" for unauthenticated requests.
	ActorID() string

	// TenantID returns the root tenant UUID for this request.
	// Always non-empty for authenticated requests.
	TenantID() string

	// CompanyID returns the active company UUID for this request.
	// Empty string if the request is not company-scoped (tenant-wide viewer).
	CompanyID() string

	// DivisionID returns the active division UUID for this request.
	// Empty string if the request is not division-scoped.
	DivisionID() string

	// OrgScope returns the full parsed org.Scope for this viewer.
	// Use this for hierarchy-aware policy checks:
	//
	//	if err := org.AssertContains(viewer.OrgScope(), recordScope); err != nil {
	//	    return definition.ErrDeny
	//	}
	OrgScope() org.Scope

	// HasRole reports whether the actor holds the given named role within
	// the active org scope. Roles are always tenant-scoped; company/division
	// scoping of roles is the responsibility of the host application.
	HasRole(role string) bool

	// IsSystem reports true when the operation originates from a machine token,
	// Temporal workflow, scheduled job, or internal service call. System callers
	// bypass user-facing privacy policies but are still subject to org-scope
	// enforcement (they must still supply a valid tenant).
	IsSystem() bool
}

// PolicyFunc determines whether the viewer may perform op on record.
//
// Return values:
//   - ErrAllow  — grant immediately; remaining policies are skipped
//   - ErrDeny   — deny immediately; remaining policies are skipped
//   - ErrSkip   — this policy abstains; continue to the next policy
//   - any other — treated as ErrDeny; the error is logged
//
// The framework is fail-closed: if the chain exhausts without ErrAllow, access
// is denied. This means entities with no policies are inaccessible to everyone
// (including system callers unless AllowSystem is included).
//
// Policies receive a nil record for list/collection-level checks. Policy
// implementations MUST handle nil gracefully.
type PolicyFunc func(ctx context.Context, viewer ViewerContext, op Op, record Record) error

// PolicyDef binds a PolicyFunc to one or more operations via a bitmask.
type PolicyDef struct {
	// Ops is the bitmask of operations this policy governs.
	// Use OpAll to govern every operation.
	Ops Op

	// Fn is the policy implementation.
	Fn PolicyFunc
}

// Policy creates a PolicyDef for the given operations and function.
//
// Example:
//
//	definition.Policy(definition.OpAll, func(ctx context.Context, v definition.ViewerContext, op definition.Op, rec definition.Record) error {
//	    if v.IsSystem() || v.HasRole("admin") {
//	        return definition.ErrAllow
//	    }
//	    return definition.ErrDeny
//	})
func Policy(ops Op, fn PolicyFunc) PolicyDef {
	return PolicyDef{Ops: ops, Fn: fn}
}

// ── Built-in policies ──────────────────────────────────────────────────────────

// AllowAll grants access to everyone, including anonymous viewers.
// Use only for genuinely public reference data (e.g. country codes, currencies).
func AllowAll(_ context.Context, _ ViewerContext, _ Op, _ Record) error {
	return ErrAllow
}

// DenyAll denies access to everyone without exception.
// Use as a safe placeholder during development to make an entity intentionally
// inaccessible until real policies are implemented.
func DenyAll(_ context.Context, _ ViewerContext, _ Op, _ Record) error {
	return ErrDeny
}

// AllowSystem grants access when IsSystem() is true, otherwise abstains.
// Chain after more specific policies so system callers bypass user checks:
//
//	Policies: []definition.PolicyDef{
//	    definition.Policy(definition.OpAll, definition.AllowSystem),
//	    definition.Policy(definition.OpAll, myUserPolicy),
//	}
func AllowSystem(_ context.Context, viewer ViewerContext, _ Op, _ Record) error {
	if viewer.IsSystem() {
		return ErrAllow
	}
	return ErrSkip
}

// AllowWithinOrgScope grants access when the record's org scope is contained
// within the viewer's org scope. Requires the record to implement OrgScoped.
// Use this as a base policy for any company- or division-scoped entity.
//
// If the record is nil (list check) this policy abstains — pair it with a
// role-based policy that handles the nil case.
func AllowWithinOrgScope(_ context.Context, viewer ViewerContext, _ Op, record Record) error {
	if record == nil {
		return ErrSkip // list-level check; defer to companion policy
	}
	scoped, ok := record.(OrgScoped)
	if !ok {
		return ErrSkip // record does not carry org scope; not our concern
	}
	if err := org.AssertContains(viewer.OrgScope(), scoped.RecordOrgScope()); err != nil {
		return ErrDeny
	}
	return ErrAllow
}

// OrgScoped is an optional interface that records may implement to expose their
// org scope to the AllowWithinOrgScope built-in policy.
type OrgScoped interface {
	RecordOrgScope() org.Scope
}
