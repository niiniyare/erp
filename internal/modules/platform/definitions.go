// Package platform declares EntityDefinitions for tenant management and IAM.
package platform

import (
	"context"

	"awo.so/framework/definition"
)

func init() {
	definition.Register(Tenant)
	definition.Register(Role)
	definition.Register(User)
}

// ── Shared policy helpers ──────────────────────────────────────────

// allowTenantAdmin grants access to system ops and tenant_admin role holders.
func allowTenantAdmin(_ context.Context, viewer definition.ViewerContext, _ definition.Op, _ definition.Record) error {
	if viewer.IsSystem() || viewer.HasRole("tenant_admin") {
		return definition.ErrAllow
	}
	return definition.ErrDeny
}

// allowIAMManager grants access to system, tenant_admin, and iam_manager.
func allowIAMManager(_ context.Context, viewer definition.ViewerContext, _ definition.Op, _ definition.Record) error {
	if viewer.IsSystem() || viewer.HasRole("tenant_admin") || viewer.HasRole("iam_manager") {
		return definition.ErrAllow
	}
	return definition.ErrDeny
}

// ── Definitions ───────────────────────────────────────────────────

// Tenant is the root multi-tenancy entity.
var Tenant = &definition.EntityDefinition{
	Name:        "tenant",
	Label:       "Tenant",
	Description: "Workspace / organisation root. Lifecycle: PENDING→ACTIVE→SUSPENDED→ARCHIVED.",
	Table:       "tenants",
	Module:      "Platform",
	SoftDelete:  true,
	Audited:     true,
	Fields: []*definition.FieldDef{
		definition.Field("slug").OfType(definition.FieldTypeSmallText).WithLabel("Slug").
			RequiredField().UniqueField().ImmutableField().WithMaxLen(63),
		definition.Field("name").OfType(definition.FieldTypeSmallText).WithLabel("Name").
			RequiredField().SearchableField().WithMaxLen(255),
		definition.Field("email").OfType(definition.FieldTypeSmallText).WithLabel("Admin Email").
			RequiredField().UniqueField().WithMaxLen(320),
		definition.Field("status").OfType(definition.FieldTypeSelect).WithLabel("Status").
			RequiredField().WithOptions("PENDING", "ACTIVE", "SUSPENDED", "ARCHIVED"),
		definition.Field("plan_tier").OfType(definition.FieldTypeSelect).WithLabel("Plan").
			WithOptions("FREE", "STARTER", "PROFESSIONAL", "ENTERPRISE").WithDefault("FREE"),
		definition.Field("timezone").OfType(definition.FieldTypeSmallText).WithLabel("Timezone").
			WithDefault("UTC"),
		definition.Field("currency_code").OfType(definition.FieldTypeSmallText).WithLabel("Currency").
			WithMaxLen(3).WithDefault("USD"),
		definition.Field("industry").OfType(definition.FieldTypeSmallText).WithLabel("Industry").WithMaxLen(100),
		definition.Field("company_size").OfType(definition.FieldTypeSelect).WithLabel("Company Size").
			WithOptions("MICRO", "SMALL", "MEDIUM", "LARGE", "ENTERPRISE"),
		definition.Field("billing_email").OfType(definition.FieldTypeSmallText).WithLabel("Billing Email").
			WithMaxLen(320),
		definition.Field("tax_id").OfType(definition.FieldTypeSmallText).WithLabel("Tax ID").
			SensitiveField().WithMaxLen(50),
		definition.Field("subdomain").OfType(definition.FieldTypeSmallText).WithLabel("Subdomain").
			UniqueField().WithMaxLen(63),
	},
	Policies: []definition.PolicyDef{
		definition.Policy(definition.OpCreate|definition.OpDelete, definition.AllowSystem),
		definition.Policy(definition.OpRead|definition.OpUpdate, allowTenantAdmin),
	},
}

// Role is the RBAC role entity.
var Role = &definition.EntityDefinition{
	Name:   "role",
	Label:  "Role",
	Table:  "roles",
	Module: "Platform",
	Fields: []*definition.FieldDef{
		definition.Field("name").OfType(definition.FieldTypeSmallText).WithLabel("Name").
			RequiredField().WithMaxLen(100),
		definition.Field("display_name").OfType(definition.FieldTypeSmallText).WithLabel("Display Name").
			WithMaxLen(150),
		definition.Field("description").OfType(definition.FieldTypeLongText).WithLabel("Description"),
		definition.Field("role_type").OfType(definition.FieldTypeSelect).WithLabel("Type").
			WithOptions("SYSTEM", "TENANT", "ENTITY", "CUSTOM", "FUNCTIONAL").WithDefault("CUSTOM"),
		definition.Field("is_active").OfType(definition.FieldTypeBool).WithLabel("Active").WithDefault(true),
	},
	Policies: []definition.PolicyDef{
		definition.Policy(definition.OpAll, allowTenantAdmin),
	},
}

// User is the IAM user entity.
var User = &definition.EntityDefinition{
	Name:       "user",
	Label:      "User",
	Table:      "users",
	Module:     "Platform",
	SoftDelete: true,
	Audited:    true,
	Fields: []*definition.FieldDef{
		definition.Field("email").OfType(definition.FieldTypeSmallText).WithLabel("Email").
			RequiredField().UniqueField().WithMaxLen(320),
		definition.Field("display_name").OfType(definition.FieldTypeSmallText).WithLabel("Display Name").
			SearchableField().WithMaxLen(150),
		definition.Field("username").OfType(definition.FieldTypeSmallText).WithLabel("Username").
			UniqueField().WithMaxLen(100),
		definition.Field("user_type").OfType(definition.FieldTypeSelect).WithLabel("Type").
			WithOptions("INTERNAL", "CUSTOMER", "VENDOR", "PARTNER", "API", "SERVICE").WithDefault("INTERNAL"),
		definition.Field("account_status").OfType(definition.FieldTypeSelect).WithLabel("Status").
			WithOptions("ACTIVE", "INACTIVE", "LOCKED", "PENDING").WithDefault("PENDING"),
		definition.Field("is_active").OfType(definition.FieldTypeBool).WithLabel("Active").WithDefault(true),
		definition.Field("mfa_enabled").OfType(definition.FieldTypeBool).WithLabel("MFA Enabled").WithDefault(false),
		definition.Field("last_login_at").OfType(definition.FieldTypeDateTime).WithLabel("Last Login").ReadOnlyField(),
		definition.Field("password_hash").OfType(definition.FieldTypeData).WithLabel("Password Hash").
			SensitiveField().HiddenInList(),
	},
	Policies: []definition.PolicyDef{
		definition.Policy(definition.OpAll, allowIAMManager),
	},
}
