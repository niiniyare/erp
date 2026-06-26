package definition

import (
	"context"
	"errors"

	"github.com/google/uuid"

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

	// TenantID returns the root tenant UUID string for this request.
	// Used by the persistence layer to set the Postgres RLS context.
	// Always non-empty for authenticated requests.
	TenantID() string

	// OrgUnitID returns the org_units.uuid the viewer is operating as.
	// uuid.Nil means the viewer is tenant-wide (no specific unit scoping).
	//
	// Use this in privacy policies for tree-based access control:
	//
	//	ok, err := orgTree.IsAncestorOrEqual(ctx, tenantID, viewer.OrgUnitID(), record.OrgUnitID())
	//	if !ok { return definition.ErrDeny }
	OrgUnitID() uuid.UUID

	// OrgScope returns the combined tenant + unit scope for this viewer.
	// Convenience accessor; equivalent to org.WithUnit(tenantID, unitID).
	OrgScope() org.Scope

	// HasRole reports whether the actor holds the given named role. Roles
	// are resolved against the viewer's OrgUnit and its ancestors (role grants
	// at a parent unit propagate down to children).
	HasRole(role string) bool

	// IsSystem reports true when the operation originates from a machine token,
	// Temporal workflow, scheduled job, or internal service call. System callers
	// bypass user-facing privacy policies but are still subject to tenant-scope
	// enforcement (a valid tenant must still be present).
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
//
// Use Op (singular) for a single operation or Ops for a bitmask of multiple operations.
// If both are set, Ops takes precedence. If neither is set, OpAll is assumed.
type PolicyDef struct {
	// Op is a convenience field for a single operation (e.g. definition.OpCreate).
	// Equivalent to setting Ops with a single-bit value.
	Op Op

	// Ops is the bitmask of operations this policy governs.
	// Use OpAll to govern every operation.
	Ops Op

	// Fn is the policy implementation.
	Fn PolicyFunc
}

// EffectiveOps returns the operation bitmask, preferring Ops over Op.
// Falls back to OpAll when neither is set.
func (p PolicyDef) EffectiveOps() Op {
	if p.Ops != 0 {
		return p.Ops
	}
	if p.Op != 0 {
		return p.Op
	}
	return OpAll
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

// AllowWithinOrgScope grants access when the record's org unit is a descendant
// of (or equal to) the viewer's org unit. Requires the record to implement
// OrgScoped and an org.Tree to be provided for containment lookup.
//
// If the record is nil (list-level check) this policy abstains — pair it with
// a role-based or list-filter policy that handles the nil case.
//
// Usage:
//
//	definition.Policy(definition.OpAll, definition.AllowWithinOrgScope(myTree))
func AllowWithinOrgScope(tree org.Tree) PolicyFunc {
	return func(ctx context.Context, viewer ViewerContext, _ Op, record Record) error {
		if record == nil {
			return ErrSkip // list-level; defer to companion policy
		}
		scoped, ok := record.(OrgScoped)
		if !ok {
			return ErrSkip // record carries no unit; not our concern
		}
		viewerUnitID := viewer.OrgUnitID()
		recordUnitID := scoped.RecordOrgUnitID()
		if viewerUnitID == uuid.Nil {
			// Tenant-wide viewer: can access all units in this tenant.
			return ErrAllow
		}
		ok, err := tree.IsAncestorOrEqual(ctx, viewer.OrgScope().TenantID, viewerUnitID, recordUnitID)
		if err != nil {
			return err // treated as ErrDeny by enforcer
		}
		if !ok {
			return ErrDeny
		}
		return ErrAllow
	}
}

// OrgScoped is an optional interface that records may implement to expose their
// org unit to the AllowWithinOrgScope built-in policy.
type OrgScoped interface {
	// RecordOrgUnitID returns the org_unit_id of this record.
	RecordOrgUnitID() uuid.UUID
}

// OwnerOnly grants access only when the named field on the record equals the
// viewer's ActorID. Abstains when record is nil (list-level check) — the list
// handler must apply an equivalent filter-by-owner predicate separately.
//
// Usage:
//
//	definition.Policy(definition.OpAll, definition.OwnerOnly("created_by_id"))
func OwnerOnly(field string) PolicyFunc {
	return func(_ context.Context, viewer ViewerContext, _ Op, record Record) error {
		if record == nil {
			return ErrSkip
		}
		owner, _ := record.Get(field).(string)
		if owner == "" {
			return ErrSkip
		}
		if owner == viewer.ActorID() {
			return ErrAllow
		}
		return ErrDeny
	}
}

// OwnerOnlyUnless grants access like OwnerOnly but abstains (instead of denying)
// when the viewer holds any of the listed roles, allowing a subsequent policy to
// grant broader access (e.g. a manager seeing all records).
//
// Usage:
//
//	definition.Policy(definition.OpAll,
//	    definition.OwnerOnlyUnless("created_by_id", "finance_manager", "admin"))
func OwnerOnlyUnless(field string, roles ...string) PolicyFunc {
	return func(_ context.Context, viewer ViewerContext, _ Op, record Record) error {
		for _, r := range roles {
			if viewer.HasRole(r) {
				return ErrSkip // let the next policy decide
			}
		}
		if record == nil {
			return ErrSkip
		}
		owner, _ := record.Get(field).(string)
		if owner == "" {
			return ErrSkip
		}
		if owner == viewer.ActorID() {
			return ErrAllow
		}
		return ErrDeny
	}
}

// RequireRole grants access when the viewer holds any of the named roles,
// otherwise denies. Abstains on nil record.
//
// Usage:
//
//	definition.Policy(definition.OpWrite, definition.RequireRole("finance_manager", "admin"))
func RequireRole(roles ...string) PolicyFunc {
	return func(_ context.Context, viewer ViewerContext, _ Op, record Record) error {
		for _, r := range roles {
			if viewer.HasRole(r) {
				return ErrAllow
			}
		}
		return ErrDeny
	}
}
