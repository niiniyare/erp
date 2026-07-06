// Package tenant implements the Awo platform Tenant module.
//
// The tenant module manages the top-level isolation boundary: every request
// is scoped to exactly one tenant. It owns three system entities:
//
//   - platform_tenant   — the tenant record itself (lifecycle state machine)
//   - platform_org_unit — organisational hierarchy (company/division/department/branch/team)
//   - platform_branch   — physical locations linked to org units
//
// All entities are SystemDefinitions because they are accessible before
// per-tenant schemas load and because they are FK targets for every other
// module in the system.
//
// Registration happens in init() so that the definitions are available to the
// compiler before any request is served.
package tenant

import "awo.so/awo/def"

func init() {
	def.Register(&Definition)
	def.Register(&OrgUnitDefinition)
	def.Register(&BranchDefinition)
}
