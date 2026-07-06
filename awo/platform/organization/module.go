// Package organization implements the Awo platform Organization module.
//
// Organization is a first-class platform module that manages the business
// hierarchy within a tenant. It is entirely separate from platform/tenant:
//
//   - platform_tenant  (platform/tenant)      — infrastructure isolation boundary
//   - platform_organization (platform/organization) — business hierarchy node
//
// A tenant has exactly one infrastructure record (platform_tenant) and zero or
// more organization nodes (platform_organization). An organization node always
// belongs to exactly one tenant via the tenant_id foreign key.
//
// Organization trees are arbitrary-depth. The tree shape is not constrained by
// the framework — tenant admins define the valid node types via settings.
// Common configurations: flat list, two-level (company/department), or deep
// hierarchies (company/region/branch/team).
//
// Scope filtering (OrganizationScope) is orthogonal to TenantContext:
// TenantContext isolates data across tenants; OrganizationScope restricts
// visibility within a single tenant's org tree.
package organization

import "awo.so/awo/def"

func init() {
	def.Register(&Definition)
}
