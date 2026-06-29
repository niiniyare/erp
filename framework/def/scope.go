package def

import (
	"fmt"

	"github.com/google/uuid"
)

// ScopeLevel determines how entity rows are isolated within the org hierarchy.
// Used on EntityDefinition.OrgScope to drive persistence column selection,
// WHERE clause generation, and RLS policy application.
type ScopeLevel string

const (
	// ScopeLevelGlobal — no tenant_id or org_unit_id.
	// Rows shared across all tenants (currencies, countries, language codes).
	ScopeLevelGlobal ScopeLevel = "global"

	// ScopeLevelTenant — rows carry tenant_id only.
	// Visible to all org units within the tenant.
	ScopeLevelTenant ScopeLevel = "tenant"

	// ScopeLevelUnit — rows carry tenant_id + org_unit_id.
	// Access is tree-based: viewer's unit must be ancestor-or-equal of record's unit.
	ScopeLevelUnit ScopeLevel = "unit"
)

// Scope is the organisational context carried on every authenticated request.
// Combines tenant identity with the viewer's position in the org tree.
type Scope struct {
	TenantID uuid.UUID
	UnitID   uuid.UUID
}

func TenantOnly(tenantID uuid.UUID) Scope       { return Scope{TenantID: tenantID} }
func WithUnit(tenantID, unitID uuid.UUID) Scope { return Scope{TenantID: tenantID, UnitID: unitID} }
func (s Scope) IsTenantWide() bool              { return s.UnitID == uuid.Nil }

func (s Scope) String() string {
	if s.IsTenantWide() {
		return fmt.Sprintf("tenant:%s", s.TenantID)
	}
	return fmt.Sprintf("tenant:%s/unit:%s", s.TenantID, s.UnitID)
}
