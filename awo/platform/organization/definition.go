// Package organization provides the platform Organization entity definition.
//
// An Organization is a node in the business hierarchy within a tenant. The tree
// is arbitrary-depth: company → division → department → team → … Nodes are
// typed via the "type" field; valid types are metadata-driven and extensible by
// tenant admins.
//
// Design decisions:
//
//   - Tenant is the infrastructure boundary (platform/tenant). Organization is
//     the business boundary. They are orthogonal concepts.
//   - Materialized path (path column) enables O(1) ancestor/descendant queries
//     without recursive CTEs. Path format: "/root-id/child-id/leaf-id/".
//   - depth column is redundant with path but kept for cheap depth-limit checks.
//   - Cycles are prevented by the PathComputeHook at creation/move time.
//   - Code is immutable: embedded in Temporal IDs, Redis keys, and audit records.
//
// Visibility scopes are declared in scope.go.
package organization

import "awo.so/awo/def"

// Definition is the SystemDefinition for the platform_organization entity.
// Registered in init() — available immediately at process start.
var Definition = def.SystemDefinition{
	Name:        "platform_organization",
	Module:      "platform",
	Label:       "Organization",
	LabelPlural: "Organizations",

	Fields: []def.FieldDef{
		{
			Name:       "name",
			Type:       def.FieldTypeData,
			Label:      "Name",
			Required:   true,
			Searchable: true,
			MaxLen:     255,
		},
		{
			// code is immutable: used in workflow IDs, Redis keys, audit records.
			Name:      "code",
			Type:      def.FieldTypeData,
			Label:     "Code",
			Required:  true,
			Unique:    true,
			Immutable: true,
			MaxLen:    50,
		},
		{
			// tenant_id is the infrastructure isolation boundary — always present.
			Name:       "tenant_id",
			Type:       def.FieldTypeLink,
			Label:      "Tenant",
			LinkTarget: "platform_tenant",
			Required:   true,
			Immutable:  true,
		},
		{
			// parent_id is nil for root nodes (top-level organization in a tenant).
			Name:       "parent_id",
			Type:       def.FieldTypeLink,
			Label:      "Parent",
			LinkTarget: "platform_organization",
		},
		{
			// type is metadata-driven: tenant admins register valid types via settings.
			// Common values: company, division, department, branch, team, region, zone.
			// Not an enum — extensible without migration.
			Name:   "type",
			Type:   def.FieldTypeData,
			Label:  "Type",
			MaxLen: 50,
		},
		{
			// path is a materialized path: "/root-id/parent-id/self-id/".
			// Computed by PathComputeHook on create and move. Never set by callers.
			Name:      "path",
			Type:      def.FieldTypeData,
			Label:     "Path",
			Immutable: false, // updated by Move operation via service layer
			MaxLen:    4096,
		},
		{
			// depth is 0 for root nodes, 1 for direct children, etc.
			// Derived from path but stored for cheap depth-limit enforcement.
			Name:    "depth",
			Type:    def.FieldTypeInt,
			Label:   "Depth",
			Default: func() any { return 0 },
		},
		{
			Name:    "active",
			Type:    def.FieldTypeBool,
			Label:   "Active",
			Default: func() any { return true },
		},
		{
			// description is optional free-text for human context.
			Name:  "description",
			Type:  def.FieldTypeSmallText,
			Label: "Description",
		},
	},

	Hooks: def.HookSet{
		BeforeCreate: []def.BeforeCreateHook{&CodeValidator{}, &PathComputeHook{}},
		BeforeUpdate: []def.BeforeUpdateHook{&MoveGuard{}},
	},

	Permissions: def.PermissionSet{
		Create: []string{"role:platform-admin", "role:tenant.admin"},
		Read:   []string{"role:platform-admin", "role:tenant.admin", "role:tenant.user"},
		Write:  []string{"role:platform-admin", "role:tenant.admin"},
		Delete: []string{"role:platform-admin", "role:tenant.admin"},
	},
}
