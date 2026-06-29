// Package platform declares EntityDefinitions for tenant management and IAM.
package platform

import (
	"context"

	"awo.so/framework/def"
)

func init() {
	def.Register(Tenant)
	def.Register(Role)
	def.Register(User)
}

// ── Shared policy helpers ──────────────────────────────────────────

// allowTenantAdmin grants access to system ops and tenant_admin role holders.
func allowTenantAdmin(_ context.Context, viewer def.ViewerContext, _ def.Op, _ def.Record) error {
	if viewer.IsSystem() || viewer.HasRole("tenant_admin") {
		return def.ErrAllow
	}
	return def.ErrDeny
}

// allowIAMManager grants access to system, tenant_admin, and iam_manager.
func allowIAMManager(_ context.Context, viewer def.ViewerContext, _ def.Op, _ def.Record) error {
	if viewer.IsSystem() || viewer.HasRole("tenant_admin") || viewer.HasRole("iam_manager") {
		return def.ErrAllow
	}
	return def.ErrDeny
}

// ── Definitions ───────────────────────────────────────────────────

// Tenant is the root multi-tenancy entity.
var Tenant = &def.EntityDefinition{
	Name:        "tenant",
	Label:       "Tenant",
	Description: "Workspace / organisation root. Lifecycle: PENDING→ACTIVE→SUSPENDED→ARCHIVED.",
	Table:       "tenants",
	Module:      "Platform",
	SoftDelete:  true,
	Audited:     true,
	Fields: []*def.FieldDef{
		def.Field("slug").OfType(def.FieldTypeSmallText).WithLabel("Slug").
			RequiredField().UniqueField().ImmutableField().WithMaxLen(63),
		def.Field("name").OfType(def.FieldTypeSmallText).WithLabel("Name").
			RequiredField().SearchableField().WithMaxLen(255),
		def.Field("email").OfType(def.FieldTypeSmallText).WithLabel("Admin Email").
			RequiredField().UniqueField().WithMaxLen(320),
		def.Field("status").OfType(def.FieldTypeSelect).WithLabel("Status").
			RequiredField().WithOptions("PENDING", "ACTIVE", "SUSPENDED", "ARCHIVED"),
		def.Field("plan_tier").OfType(def.FieldTypeSelect).WithLabel("Plan").
			WithOptions("FREE", "STARTER", "PROFESSIONAL", "ENTERPRISE").WithDefault("FREE"),
		def.Field("timezone").OfType(def.FieldTypeSmallText).WithLabel("Timezone").
			WithDefault("UTC"),
		def.Field("currency_code").OfType(def.FieldTypeSmallText).WithLabel("Currency").
			WithMaxLen(3).WithDefault("USD"),
		def.Field("industry").OfType(def.FieldTypeSmallText).WithLabel("Industry").WithMaxLen(100),
		def.Field("company_size").OfType(def.FieldTypeSelect).WithLabel("Company Size").
			WithOptions("MICRO", "SMALL", "MEDIUM", "LARGE", "ENTERPRISE"),
		def.Field("billing_email").OfType(def.FieldTypeSmallText).WithLabel("Billing Email").
			WithMaxLen(320),
		def.Field("tax_id").OfType(def.FieldTypeSmallText).WithLabel("Tax ID").
			SensitiveField().WithMaxLen(50),
		def.Field("subdomain").OfType(def.FieldTypeSmallText).WithLabel("Subdomain").
			UniqueField().WithMaxLen(63),
	},
	Policies: []def.PolicyDef{
		def.Policy(def.OpCreate|def.OpDelete, def.AllowSystem),
		def.Policy(def.OpRead|def.OpUpdate, allowTenantAdmin),
	},
}

// Role is the RBAC role entity.
var Role = &def.EntityDefinition{
	Name:   "role",
	Label:  "Role",
	Table:  "roles",
	Module: "Platform",
	Fields: []*def.FieldDef{
		def.Field("name").OfType(def.FieldTypeSmallText).WithLabel("Name").
			RequiredField().WithMaxLen(100),
		def.Field("display_name").OfType(def.FieldTypeSmallText).WithLabel("Display Name").
			WithMaxLen(150),
		def.Field("description").OfType(def.FieldTypeLongText).WithLabel("Description"),
		def.Field("role_type").OfType(def.FieldTypeSelect).WithLabel("Type").
			WithOptions("SYSTEM", "TENANT", "ENTITY", "CUSTOM", "FUNCTIONAL").WithDefault("CUSTOM"),
		def.Field("is_active").OfType(def.FieldTypeBool).WithLabel("Active").WithDefault(true),
	},
	Policies: []def.PolicyDef{
		def.Policy(def.OpAll, allowTenantAdmin),
	},
}

// User is the IAM user entity.
var User = &def.EntityDefinition{
	Name:       "user",
	Label:      "User",
	Table:      "users",
	Module:     "Platform",
	SoftDelete: true,
	Audited:    true,
	Fields: []*def.FieldDef{
		def.Field("email").OfType(def.FieldTypeSmallText).WithLabel("Email").
			RequiredField().UniqueField().WithMaxLen(320),
		def.Field("display_name").OfType(def.FieldTypeSmallText).WithLabel("Display Name").
			SearchableField().WithMaxLen(150),
		def.Field("username").OfType(def.FieldTypeSmallText).WithLabel("Username").
			UniqueField().WithMaxLen(100),
		def.Field("user_type").OfType(def.FieldTypeSelect).WithLabel("Type").
			WithOptions("INTERNAL", "CUSTOMER", "VENDOR", "PARTNER", "API", "SERVICE").WithDefault("INTERNAL"),
		def.Field("account_status").OfType(def.FieldTypeSelect).WithLabel("Status").
			WithOptions("ACTIVE", "INACTIVE", "LOCKED", "PENDING").WithDefault("PENDING"),
		def.Field("is_active").OfType(def.FieldTypeBool).WithLabel("Active").WithDefault(true),
		def.Field("mfa_enabled").OfType(def.FieldTypeBool).WithLabel("MFA Enabled").WithDefault(false),
		def.Field("last_login_at").OfType(def.FieldTypeDateTime).WithLabel("Last Login").ReadOnlyField(),
		def.Field("password_hash").OfType(def.FieldTypeData).WithLabel("Password Hash").
			SensitiveField().HiddenInList(),
	},
	Policies: []def.PolicyDef{
		def.Policy(def.OpAll, allowIAMManager),
	},
}
