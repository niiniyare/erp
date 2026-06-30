package iam

import (
	"awo.so/framework/def"
	"awo.so/framework/platform/org"
)

// TenantUserDefinition drives CRUD API, SDUI, and audit for tenant users.
// The auth service uses direct SQL for login; this definition covers admin management.
var TenantUserDefinition = def.EntityDefinition{
	Name:        "tenant_user",
	Label:       "User",
	LabelPlural: "Users",
	Module:      "IAM",
	Table:       "tenant_users",
	OrgScope:    org.ScopeLevelTenant, // visible to all org units within the tenant
	SoftDelete:  false,                // status field controls visibility
	Audited:     true,

	Fields: []*def.FieldDef{
		def.Field("email").OfType(def.FieldTypeData).
			WithLabel("Email").
			RequiredField().
			UniqueField().
			SearchableField().
			WithMaxLen(320),

		def.Field("full_name").OfType(def.FieldTypeData).
			WithLabel("Full Name").
			RequiredField().
			SearchableField().
			WithMaxLen(255),

		def.Field("status").OfType(def.FieldTypeSelect).
			WithLabel("Status").
			RequiredField().
			WithOptions("PENDING_VERIFICATION", "ACTIVE", "SUSPENDED", "DEACTIVATED").
			WithDefault("PENDING_VERIFICATION"),

		def.Field("actor_type").OfType(def.FieldTypeSelect).
			WithLabel("Actor Type").
			RequiredField().
			WithOptions("user", "service_account").
			WithDefault("user"),

		// password_hash is write-only — never returned in API responses.
		def.Field("password_hash").OfType(def.FieldTypeLongText).
			WithLabel("Password Hash").
			SensitiveField().
			HiddenInList(),

		def.Field("org_unit_id").OfType(def.FieldTypeLink).
			WithLabel("Primary Org Unit").
			LinksTo("org_unit"),

		def.Field("must_change_password").OfType(def.FieldTypeBool).
			WithLabel("Must Change Password").
			WithDefault(false),

		def.Field("last_login_at").OfType(def.FieldTypeDateTime).
			WithLabel("Last Login").
			ReadOnlyField(),
	},

	Policies: []def.PolicyDef{
		// System callers (Temporal, internal services) bypass user policies.
		def.Policy(def.OpAll, def.AllowSystem),
		// Tenant admins can manage all users.
		def.Policy(def.OpAll, def.RequireRole("tenant-admin")),
		// Managers can read but not create/delete.
		def.Policy(def.OpRead, def.RequireRole("tenant-manager", "tenant-staff")),
	},
}

// RoleDefinition drives CRUD for role management.
var RoleDefinition = def.EntityDefinition{
	Name:        "role",
	Label:       "Role",
	LabelPlural: "Roles",
	Module:      "IAM",
	Table:       "roles",
	OrgScope:    org.ScopeLevelTenant,
	Audited:     true,

	Fields: []*def.FieldDef{
		def.Field("name").OfType(def.FieldTypeData).
			WithLabel("Name").
			RequiredField().
			SearchableField().
			WithMaxLen(100),

		def.Field("slug").OfType(def.FieldTypeData).
			WithLabel("Slug").
			RequiredField().
			ImmutableField().
			WithMaxLen(100),

		def.Field("description").OfType(def.FieldTypeSmallText).
			WithLabel("Description"),

		def.Field("is_system").OfType(def.FieldTypeBool).
			WithLabel("System Role").
			WithDefault(false).
			ReadOnlyField(),

		def.Field("is_active").OfType(def.FieldTypeBool).
			WithLabel("Active").
			WithDefault(true),

		def.Field("parent_role_id").OfType(def.FieldTypeLink).
			WithLabel("Inherits From").
			LinksTo("role"),
	},

	Policies: []def.PolicyDef{
		def.Policy(def.OpAll, def.AllowSystem),
		def.Policy(def.OpAll, def.RequireRole("tenant-admin")),
		// All authenticated users can read roles (for UI dropdowns).
		def.Policy(def.OpRead, def.RequireRole("tenant-manager", "tenant-staff", "tenant-api-readonly")),
	},
}

// PermissionDefinition drives CRUD for per-role permission strings.
var PermissionDefinition = def.EntityDefinition{
	Name:        "permission",
	Label:       "Permission",
	LabelPlural: "Permissions",
	Module:      "IAM",
	Table:       "permissions",
	OrgScope:    org.ScopeLevelTenant,
	Audited:     true,

	Fields: []*def.FieldDef{
		def.Field("role_id").OfType(def.FieldTypeLink).
			WithLabel("Role").
			RequiredField().
			LinksTo("role"),

		def.Field("permission").OfType(def.FieldTypeData).
			WithLabel("Permission String").
			RequiredField().
			WithMaxLen(200),

		def.Field("granted").OfType(def.FieldTypeBool).
			WithLabel("Granted").
			WithDefault(true),
	},

	Policies: []def.PolicyDef{
		def.Policy(def.OpAll, def.AllowSystem),
		def.Policy(def.OpAll, def.RequireRole("tenant-admin")),
		def.Policy(def.OpRead, def.RequireRole("tenant-manager")),
	},
}

// UserRoleAssignmentDefinition drives CRUD for user→role bindings.
var UserRoleAssignmentDefinition = def.EntityDefinition{
	Name:        "user_role_assignment",
	Label:       "Role Assignment",
	LabelPlural: "Role Assignments",
	Module:      "IAM",
	Table:       "user_role_assignments",
	OrgScope:    org.ScopeLevelTenant,
	Audited:     true,

	Fields: []*def.FieldDef{
		def.Field("user_id").OfType(def.FieldTypeLink).
			WithLabel("User").
			RequiredField().
			LinksTo("tenant_user"),

		def.Field("role_id").OfType(def.FieldTypeLink).
			WithLabel("Role").
			RequiredField().
			LinksTo("role"),

		def.Field("org_unit_id").OfType(def.FieldTypeLink).
			WithLabel("Scoped to Org Unit").
			LinksTo("org_unit"),

		def.Field("is_active").OfType(def.FieldTypeBool).
			WithLabel("Active").
			WithDefault(true),

		def.Field("expires_at").OfType(def.FieldTypeDateTime).
			WithLabel("Expires At"),

		def.Field("reason").OfType(def.FieldTypeSmallText).
			WithLabel("Reason"),
	},

	Policies: []def.PolicyDef{
		def.Policy(def.OpAll, def.AllowSystem),
		def.Policy(def.OpAll, def.RequireRole("tenant-admin")),
		def.Policy(def.OpRead, def.RequireRole("tenant-manager")),
	},
}
