// Package iam provides the IAM (Identity and Access Management) entity
// definitions for the Awo platform.
//
// IAM consists of three system entities:
//
//   - platform_user   — the canonical user identity
//   - platform_role   — named collection of permissions within a tenant
//   - platform_session — active session tokens (Redis-backed, DB as source of truth)
//
// All IAM entities are mandatory system entities. Declaring any of them as
// CustomDefinition causes the registry validator to panic at startup.
package iam

import (
	"awo.so/awo/def"
)

// UserDefinition is the platform_user entity.
var UserDefinition = def.SystemDefinition{
	Name:        "iam_user",
	Module:      "platform",
	Label:       "User",
	LabelPlural: "Users",

	Fields: []def.FieldDef{
		{
			Name:     "email",
			Type:     def.FieldTypeData,
			Label:    "Email",
			Required: true,
			Unique:   true,
			MaxLen:   254,
		},
		{
			Name:      "password_hash",
			Type:      def.FieldTypeData,
			Label:     "Password Hash",
			Sensitive: true, // excluded from all logs and API responses
			Required:  true,
		},
		{
			Name:   "first_name",
			Type:   def.FieldTypeData,
			Label:  "First Name",
			MaxLen: 100,
		},
		{
			Name:   "last_name",
			Type:   def.FieldTypeData,
			Label:  "Last Name",
			MaxLen: 100,
		},
		{
			Name:    "status",
			Type:    def.FieldTypeSelect,
			Label:   "Status",
			Options: []string{"pending_verification", "active", "suspended", "deleted"},
			Default: func() any { return "pending_verification" },
		},
		{
			Name:  "tenant_id_fk",
			Type:  def.FieldTypeLink,
			Label: "Tenant",
			// LinkTarget is "platform_tenant" — resolved by compiler.
			LinkTarget: "platform_tenant",
			Required:   true,
		},
		{
			Name:      "email_verified_at",
			Type:      def.FieldTypeDateTime,
			Label:     "Email Verified At",
			Sensitive: true,
		},
		{
			Name:      "last_login_at",
			Type:      def.FieldTypeDateTime,
			Label:     "Last Login At",
			Sensitive: true,
		},
		{
			Name:      "mfa_secret",
			Type:      def.FieldTypeData,
			Label:     "MFA Secret",
			Sensitive: true,
		},
		{
			Name:    "mfa_enabled",
			Type:    def.FieldTypeBool,
			Label:   "MFA Enabled",
			Default: func() any { return false },
		},
		{
			Name:    "locale",
			Type:    def.FieldTypeData,
			Label:   "Locale",
			MaxLen:  10,
			Default: func() any { return "en-KE" },
		},
		{
			Name:    "timezone",
			Type:    def.FieldTypeData,
			Label:   "Timezone",
			MaxLen:  64,
			Default: func() any { return "Africa/Nairobi" },
		},
	},

	Hooks: def.HookSet{
		BeforeCreate: []def.BeforeCreateHook{&UserValidator{}},
	},

	Permissions: def.PermissionSet{
		Create: []string{"role:platform-admin", "role:tenant.admin"},
		Read:   []string{"role:platform-admin", "role:tenant.admin", "role:tenant.user"},
		Write:  []string{"role:platform-admin", "role:tenant.admin"},
		Delete: []string{"role:platform-admin"},
	},
}

// RoleDefinition is the platform_role entity.
// Roles are tenant-scoped named permission sets managed via Casbin policies.
var RoleDefinition = def.SystemDefinition{
	Name:        "iam_role",
	Module:      "platform",
	Label:       "Role",
	LabelPlural: "Roles",

	Fields: []def.FieldDef{
		{
			Name:      "name",
			Type:      def.FieldTypeData,
			Label:     "Role Name",
			Required:  true,
			Immutable: true, // role names appear in Casbin policies — rename is a migration
			MaxLen:    100,
		},
		{
			Name:   "label",
			Type:   def.FieldTypeData,
			Label:  "Display Label",
			MaxLen: 255,
		},
		{
			Name:  "description",
			Type:  def.FieldTypeSmallText,
			Label: "Description",
		},
		{
			Name:    "is_system",
			Type:    def.FieldTypeBool,
			Label:   "System Role",
			Default: func() any { return false },
		},
	},

	Permissions: def.PermissionSet{
		Create: []string{"role:platform-admin", "role:tenant.admin"},
		Read:   []string{"role:platform-admin", "role:tenant.admin", "role:tenant.user"},
		Write:  []string{"role:platform-admin", "role:tenant.admin"},
		Delete: []string{"role:platform-admin"},
	},
}

// SessionDefinition is the iam_session entity.
// Sessions are written to PostgreSQL as the durable source of truth; Redis
// caches the token for fast validation. Redis failure invalidates all sessions
// (correct security behaviour — cannot authenticate without session store).
var SessionDefinition = def.SystemDefinition{
	Name:        "iam_session",
	Module:      "platform",
	Label:       "Session",
	LabelPlural: "Sessions",

	Fields: []def.FieldDef{
		{
			Name:       "user_id",
			Type:       def.FieldTypeLink,
			Label:      "User",
			LinkTarget: "iam_user",
			Required:   true,
			Immutable:  true,
		},
		{
			Name:      "token_hash",
			Type:      def.FieldTypeData,
			Label:     "Token Hash",
			Sensitive: true,
			Required:  true,
			Immutable: true,
			MaxLen:    128,
		},
		{
			Name:   "ip_address",
			Type:   def.FieldTypeData,
			Label:  "IP Address",
			MaxLen: 45,
		},
		{
			Name:  "user_agent",
			Type:  def.FieldTypeSmallText,
			Label: "User Agent",
		},
		{
			Name:     "expires_at",
			Type:     def.FieldTypeDateTime,
			Label:    "Expires At",
			Required: true,
		},
		{
			Name:    "revoked",
			Type:    def.FieldTypeBool,
			Label:   "Revoked",
			Default: func() any { return false },
		},
		{
			Name:  "revoked_at",
			Type:  def.FieldTypeDateTime,
			Label: "Revoked At",
		},
	},

	Permissions: def.PermissionSet{
		Create: []string{"role:platform-admin", "role:tenant.user"},
		Read:   []string{"role:platform-admin", "role:tenant.admin"},
		Write:  []string{"role:platform-admin"},
		Delete: []string{"role:platform-admin"},
	},
}

func init() {
	def.Register(&UserDefinition)
	def.Register(&RoleDefinition)
	def.Register(&SessionDefinition)
}
