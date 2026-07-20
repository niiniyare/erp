// Package demo is a minimal example module used for framework validation.
// It contains one entity (demo_customer) that exercises the full pipeline:
// registry, compiler, runtime hooks, repository, API, SDUI, and org scope.
//
// Do NOT use this module in production. It exists solely to prove that the
// framework works end-to-end before ERP modules are implemented.
package demo

import "awo.so/awo/def"

// CustomerDefinition is the platform_organization-scoped customer entity.
// It is intentionally simple: one of everything the framework supports.
var CustomerDefinition = def.SystemDefinition{
	Name:        "customer",
	Module:      "demo",
	Label:       "Customer",
	LabelPlural: "Customers",

	Fields: []def.FieldDef{
		{
			// tenant_id — framework standard; enforced by runtime + RLS.
			Name:       "tenant_id",
			Type:       def.FieldTypeLink,
			Label:      "Tenant",
			LinkTarget: "platform_tenant",
			Required:   true,
			Immutable:  true,
		},
		{
			// org_id — application-layer org scope; resolved by OrganizationService.
			Name:       "org_id",
			Type:       def.FieldTypeLink,
			Label:      "Organization",
			LinkTarget: "platform_organization",
			Required:   true,
		},
		{
			// customer_code: NamingSeries-like but simple Data for demo simplicity.
			Name:       "customer_code",
			Type:       def.FieldTypeData,
			Label:      "Code",
			Required:   true,
			Unique:     true,
			Immutable:  true,
			Searchable: true,
			MaxLen:     50,
		},
		{
			Name:       "name",
			Type:       def.FieldTypeData,
			Label:      "Name",
			Required:   true,
			Searchable: true,
			MaxLen:     255,
		},
		{
			Name:   "email",
			Type:   def.FieldTypeData,
			Label:  "Email",
			Unique: true,
			MaxLen: 254,
		},
		{
			Name:   "phone",
			Type:   def.FieldTypeData,
			Label:  "Phone",
			MaxLen: 30,
		},
		{
			Name:    "active",
			Type:    def.FieldTypeBool,
			Label:   "Active",
			Default: func() any { return true },
		},
	},

	Hooks: def.HookSet{
		BeforeCreate: []def.BeforeCreateHook{&CustomerValidator{}},
	},

	Permissions: def.PermissionSet{
		// Role-to-permission mappings seeded in iam_role_permissions:
		//   role:tenant.admin → demo.customer.{create,read,update,delete}
		//   role:tenant.user  → demo.customer.{create,read,update}
		Create: []string{"demo.customer.create"},
		Read:   []string{"demo.customer.read"},
		Write:  []string{"demo.customer.update"},
		Delete: []string{"demo.customer.delete"},
	},
}
