// Package org defines the organisational hierarchy enforced by the Awo Framework.
//
// # Hierarchy
//
// Every piece of data in the system belongs to exactly one node in a three-level
// tree:
//
//	Tenant  ──┐  Root organisation — the SaaS customer.
//	           │  Owns all companies and their data.
//	           │  Example: "Acme Corporation"
//	           │
//	           ├── Company  ──┐  Legal entity / subsidiary within a tenant.
//	           │               │  Has its own chart of accounts, fiscal calendar,
//	           │               │  and regulatory settings.
//	           │               │  Examples: "Acme UK Ltd", "Acme Kenya Ltd"
//	           │               │
//	           │               └── Division  ──  Business unit, branch, or
//	           │                                 department within a company.
//	           │                                 Inherits company settings but
//	           │                                 can carry its own cost centres
//	           │                                 and budgets.
//	           │                                 Examples: "East Africa Sales",
//	           │                                 "Nairobi Branch", "Engineering"
//	           │
//	           └── Company ── Division ── ...   (repeats for every company)
//
// # Scope rules
//
//   - TenantID is ALWAYS required for any scoped operation.
//   - CompanyID is required for company-scoped or division-scoped entities.
//   - DivisionID requires CompanyID to be set first.
//   - A nil pointer means "not scoped at this level" (operates at parent level).
//
// # Data visibility
//
// A viewer at company level can see all divisions within that company.
// A viewer at tenant level can see all companies.
// Cross-company data access is a policy violation — never implicit.
package org

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// NodeType identifies the level in the organisational hierarchy.
type NodeType string

const (
	// NodeTypeTenant is the root organisation — the platform customer.
	// Every other node belongs to exactly one tenant.
	NodeTypeTenant NodeType = "tenant"

	// NodeTypeCompany is a legal entity within a tenant.
	// Companies are the primary unit of financial reporting and regulatory compliance.
	// Each company has its own: chart of accounts, fiscal calendar, tax settings,
	// currency configuration, and employee roster.
	NodeTypeCompany NodeType = "company"

	// NodeTypeDivision is a business unit or branch within a company.
	// Divisions share the company's financial configuration but maintain their own
	// cost centres, project allocations, and operational budgets.
	NodeTypeDivision NodeType = "division"
)

// ScopeLevel defines which level of the hierarchy scopes an entity's data rows.
// The framework uses this to generate the correct WHERE clause and RLS policy.
type ScopeLevel string

const (
	// ScopeLevelGlobal — no tenant scoping.
	// Rows are shared across ALL tenants on the platform.
	// Use for: currencies, countries, language codes, public reference data.
	ScopeLevelGlobal ScopeLevel = "global"

	// ScopeLevelTenant — rows belong to a tenant, visible across all its companies.
	// Use for: users, roles, subscription features, tenant configuration.
	ScopeLevelTenant ScopeLevel = "tenant"

	// ScopeLevelCompany — rows belong to one company within a tenant.
	// Use for: GL accounts, employees, fiscal years, tax codes, budgets.
	// Queries automatically filter by company_id; cross-company reads are forbidden.
	ScopeLevelCompany ScopeLevel = "company"

	// ScopeLevelDivision — rows belong to one division within a company.
	// Use for: sales targets, project allocations, divisional budgets.
	// Queries filter by both company_id and division_id.
	ScopeLevelDivision ScopeLevel = "division"
)

// Scope represents the full organisational context for a request or data row.
//
// It forms a path through the hierarchy:
//
//	Tenant ──► Company ──► Division
//
// Scope is carried in the request context (see contextutil.WithOrgScope) and
// is attached to every database write by the persistence layer.
type Scope struct {
	// TenantID is the root organisation. Always required.
	TenantID uuid.UUID

	// CompanyID identifies the legal entity within the tenant.
	// nil means the operation is tenant-wide (not company-scoped).
	CompanyID *uuid.UUID

	// DivisionID identifies the business unit within the company.
	// nil means the operation is company-wide.
	// DivisionID MUST NOT be set without CompanyID.
	DivisionID *uuid.UUID
}

// TenantOnly returns a tenant-wide Scope with no company or division.
func TenantOnly(tenantID uuid.UUID) Scope {
	return Scope{TenantID: tenantID}
}

// WithCompany returns a company-scoped Scope.
func WithCompany(tenantID, companyID uuid.UUID) Scope {
	return Scope{TenantID: tenantID, CompanyID: &companyID}
}

// WithDivision returns a division-scoped Scope.
func WithDivision(tenantID, companyID, divisionID uuid.UUID) Scope {
	return Scope{TenantID: tenantID, CompanyID: &companyID, DivisionID: &divisionID}
}

// Level returns the deepest level this scope operates at.
func (s Scope) Level() NodeType {
	if s.DivisionID != nil {
		return NodeTypeDivision
	}
	if s.CompanyID != nil {
		return NodeTypeCompany
	}
	return NodeTypeTenant
}

// Validate reports whether this scope is internally consistent.
func (s Scope) Validate() error {
	if s.TenantID == uuid.Nil {
		return errors.New("org scope: tenant_id is required")
	}
	if s.DivisionID != nil && s.CompanyID == nil {
		return errors.New("org scope: division_id requires company_id to be set")
	}
	return nil
}

// Contains reports whether this scope encompasses the target scope.
//
// Examples:
//   - Tenant scope contains all company and division scopes within that tenant.
//   - Company scope contains all division scopes within that company.
//   - Division scope contains only itself.
func (s Scope) Contains(target Scope) bool {
	if s.TenantID != target.TenantID {
		return false // different tenant — never contained
	}
	if s.CompanyID == nil {
		return true // tenant-wide contains everything in this tenant
	}
	if target.CompanyID == nil || *s.CompanyID != *target.CompanyID {
		return false // different or missing company
	}
	if s.DivisionID == nil {
		return true // company-wide contains all divisions in this company
	}
	return target.DivisionID != nil && *s.DivisionID == *target.DivisionID
}

// String returns a human-readable representation for logging and debugging.
func (s Scope) String() string {
	switch s.Level() {
	case NodeTypeDivision:
		return fmt.Sprintf("tenant:%s/company:%s/division:%s",
			s.TenantID, *s.CompanyID, *s.DivisionID)
	case NodeTypeCompany:
		return fmt.Sprintf("tenant:%s/company:%s", s.TenantID, *s.CompanyID)
	default:
		return fmt.Sprintf("tenant:%s", s.TenantID)
	}
}

// Node is the interface implemented by any entity that participates in the
// organisational hierarchy (Tenant, Company, Division records themselves).
type Node interface {
	// OrgID returns the UUID of this node.
	OrgID() uuid.UUID

	// OrgType returns the level of this node in the hierarchy.
	OrgType() NodeType

	// TenantID returns the root tenant this node belongs to.
	TenantID() uuid.UUID

	// ParentID returns the UUID of the immediate parent node.
	// Returns nil for the root (Tenant) node.
	ParentID() *uuid.UUID
}

// ErrCrossTenantAccess is returned when an operation attempts to read or write
// data belonging to a different tenant. This is always a security violation.
var ErrCrossTenantAccess = errors.New("cross-tenant data access is forbidden")

// ErrCrossCompanyAccess is returned when a company-scoped viewer attempts to
// access data belonging to a different company within the same tenant.
var ErrCrossCompanyAccess = errors.New("cross-company data access is forbidden")

// ErrScopeInsufficient is returned when the request scope is narrower than
// required to perform the operation (e.g. division-scoped viewer requesting
// tenant-wide data without explicit elevation).
var ErrScopeInsufficient = errors.New("organisational scope is insufficient for this operation")

// AssertContains returns nil if parent scope contains child scope, or an
// appropriate error if the child is outside the parent's boundary.
// Use this in policy functions to enforce hierarchy boundaries.
func AssertContains(parent, child Scope) error {
	if parent.TenantID != child.TenantID {
		return fmt.Errorf("%w: viewer tenant %s != data tenant %s",
			ErrCrossTenantAccess, parent.TenantID, child.TenantID)
	}
	if !parent.Contains(child) {
		return fmt.Errorf("%w: viewer scope %s does not contain data scope %s",
			ErrCrossCompanyAccess, parent, child)
	}
	return nil
}
