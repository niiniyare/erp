// Package registry provides the module registry entity definitions.
//
// The module registry tracks which business modules are installed and active
// for each tenant. Module activation triggers a provisioning workflow that
// seeds module-specific data (default roles, settings, chart of accounts, etc).
//
// The registry is not to be confused with awo/registry (which manages
// EntityDefinition registration at startup). This is the runtime, per-tenant
// module activation record.
package registry

import (
	"awo.so/awo/def"
)

// ModuleDefinition is the platform_module entity.
// One record per installed module in the entire platform (not per tenant).
var ModuleDefinition = def.SystemDefinition{
	Name:        "module",
	Module:      "platform",
	Label:       "Module",
	LabelPlural: "Modules",

	Fields: []def.FieldDef{
		{
			Name:      "key",
			Type:      def.FieldTypeData,
			Label:     "Module Key",
			Required:  true,
			Unique:    true,
			Immutable: true,
			MaxLen:    50,
			// Convention: short lowercase slug e.g. "finance", "hr", "crm"
		},
		{
			Name:   "label",
			Type:   def.FieldTypeData,
			Label:  "Label",
			MaxLen: 255,
		},
		{
			Name:  "description",
			Type:  def.FieldTypeSmallText,
			Label: "Description",
		},
		{
			Name:    "version",
			Type:    def.FieldTypeData,
			Label:   "Version",
			MaxLen:  20,
			Default: func() any { return "1.0.0" },
		},
		{
			Name:    "active",
			Type:    def.FieldTypeBool,
			Label:   "Active",
			Default: func() any { return true },
		},
		{
			Name:  "dependencies",
			Type:  def.FieldTypeJSON,
			Label: "Dependencies",
			// Array of module keys this module requires.
		},
	},

	Permissions: def.PermissionSet{
		Create: []string{"role:platform-admin"},
		Read:   []string{"role:platform-admin", "role:tenant.admin"},
		Write:  []string{"role:platform-admin"},
		Delete: []string{"role:platform-admin"},
	},
}

// TenantModuleDefinition is the platform_tenant_module entity.
// One record per module per tenant. Tracks activation state.
var TenantModuleDefinition = def.SystemDefinition{
	Name:        "tenant_module",
	Module:      "platform",
	Label:       "Tenant Module",
	LabelPlural: "Tenant Modules",

	Fields: []def.FieldDef{
		{
			Name:      "module_key",
			Type:      def.FieldTypeData,
			Label:     "Module",
			Required:  true,
			Immutable: true,
			MaxLen:    50,
		},
		{
			Name:    "status",
			Type:    def.FieldTypeSelect,
			Label:   "Status",
			Options: []string{"installing", "active", "suspended", "uninstalling"},
			Default: func() any { return "installing" },
		},
		{
			Name:  "installed_at",
			Type:  def.FieldTypeDateTime,
			Label: "Installed At",
		},
		{
			Name:  "config",
			Type:  def.FieldTypeJSON,
			Label: "Module Config",
		},
	},

	Permissions: def.PermissionSet{
		Create: []string{"role:platform-admin", "role:tenant.admin"},
		Read:   []string{"role:platform-admin", "role:tenant.admin"},
		Write:  []string{"role:platform-admin"},
		Delete: []string{"role:platform-admin"},
	},
}

func init() {
	def.Register(&ModuleDefinition)
	def.Register(&TenantModuleDefinition)
}
