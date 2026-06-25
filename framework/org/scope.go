package org

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// ScopeLevel defines how an entity's data rows are scoped within the org hierarchy.
// Used on EntityDefinition to tell the persistence layer which columns to include
// and how to generate WHERE clauses and RLS policies.
type ScopeLevel string

const (
	// ScopeLevelGlobal — no tenant or unit scoping.
	// Rows are shared across ALL tenants (currencies, countries, language codes).
	// Generated columns: none of tenant_id / org_unit_id.
	ScopeLevelGlobal ScopeLevel = "global"

	// ScopeLevelTenant — rows belong to a tenant but are not unit-scoped.
	// Use for: users, roles, subscription features, tenant configuration.
	// Generated columns: tenant_id only.
	ScopeLevelTenant ScopeLevel = "tenant"

	// ScopeLevelUnit — rows belong to one org unit within a tenant.
	// Access is tree-based: viewer's unit must be an ancestor-or-equal of the record's unit.
	// Use for: GL accounts, invoices, employees, cost centre budgets — any entity that
	// lives at a specific node in the org hierarchy.
	// Generated columns: tenant_id + org_unit_id.
	ScopeLevelUnit ScopeLevel = "unit"
)

// Scope is the organisational context carried on every authenticated request.
//
// It combines the RLS tenant identity (TenantID) with the viewer's position in
// the org tree (UnitID). The pair is stored in the request context via
// contextutil.WithOrgScope and read by privacy policies and the persistence layer.
//
// # Invariants
//
//   - TenantID is always non-nil for authenticated requests.
//   - UnitID == uuid.Nil means the viewer is operating at the tenant level without
//     being scoped to a specific org unit (e.g. a platform admin or a system actor).
//   - When UnitID is set it MUST belong to TenantID (enforced by auth middleware).
type Scope struct {
	// TenantID is the root organisation. Required for RLS context.
	TenantID uuid.UUID

	// UnitID is the org_unit the viewer is operating as. uuid.Nil means tenant-wide.
	// When set, it points to a row in org_units where org_units.tenant_id = TenantID.
	UnitID uuid.UUID
}

// TenantOnly returns a tenant-wide Scope with no unit scoping.
// Used for system actors, platform admins, and Temporal workflow contexts that
// need tenant isolation without a specific org unit.
func TenantOnly(tenantID uuid.UUID) Scope {
	return Scope{TenantID: tenantID}
}

// WithUnit returns a unit-scoped Scope.
// unitID must be the UUID of a live org_units row belonging to tenantID.
func WithUnit(tenantID, unitID uuid.UUID) Scope {
	return Scope{TenantID: tenantID, UnitID: unitID}
}

// IsTenantWide reports true when no specific org unit is set.
func (s Scope) IsTenantWide() bool { return s.UnitID == uuid.Nil }

// Validate reports whether this scope is internally consistent.
func (s Scope) Validate() error {
	if s.TenantID == uuid.Nil {
		return errors.New("org scope: tenant_id is required")
	}
	return nil
}

// String returns a human-readable representation for logging and debugging.
func (s Scope) String() string {
	if s.IsTenantWide() {
		return fmt.Sprintf("tenant:%s", s.TenantID)
	}
	return fmt.Sprintf("tenant:%s/unit:%s", s.TenantID, s.UnitID)
}

// ── Legacy 3-level hierarchy ───────────────────────────────────────────────
// The following types and functions are retained for migration compatibility.
// New code should use Scope + Tree instead.

// NodeType identifies a level in the organisational hierarchy.
// Deprecated: use UnitType for org_units rows.
type NodeType = UnitType

// Sentinel errors for scope boundary violations.
var (
	// ErrCrossTenantAccess is returned when an operation attempts to read or write
	// data belonging to a different tenant. Always a security violation.
	ErrCrossTenantAccess = errors.New("cross-tenant data access is forbidden")

	// ErrCrossUnitAccess is returned when a unit-scoped viewer attempts to access
	// data belonging to an org unit outside their subtree.
	ErrCrossUnitAccess = errors.New("cross-unit data access is forbidden")

	// ErrScopeInsufficient is returned when the request scope is narrower than
	// required to perform the operation.
	ErrScopeInsufficient = errors.New("organisational scope is insufficient for this operation")
)
