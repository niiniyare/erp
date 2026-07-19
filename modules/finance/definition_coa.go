package finance

import "awo.so/awo/def"

// ChartOfAccountsDefinition — COA master header.
var ChartOfAccountsDefinition = def.SystemDefinition{
	Name:        "chart_of_accounts",
	Module:      "finance",
	Label:       "Chart of Accounts",
	LabelPlural: "Charts of Accounts",
	Description: "COA master. Multiple COAs per tenant supported; mark one as default.",
	Fields: []def.FieldDef{
		{
			Name:       "name",
			Type:       def.FieldTypeData,
			Required:   true,
			Searchable: true,
			MaxLen:     255,
		},
		{
			Name:      "code",
			Type:      def.FieldTypeData,
			Required:  true,
			Unique:    true,
			Immutable: true,
			MaxLen:    50,
		},
		{Name: "description", Type: def.FieldTypeSmallText},
		{
			Name:    "is_default",
			Type:    def.FieldTypeBool,
			Default: func() any { return false },
		},
		{
			Name:     "root_account_type",
			Type:     def.FieldTypeSelect,
			Options:  []string{"Asset", "Liability", "Equity", "Income", "Expense"},
			Required: true,
		},
	},
	Permissions: def.PermissionSet{
		Create: []string{"role:tenant.admin", "role:finance.manager"},
		Read:   []string{"role:tenant.admin", "role:finance.manager", "role:finance.accountant", "role:finance.viewer", "role:tenant.user"},
		Write:  []string{"role:tenant.admin", "role:finance.manager"},
		Delete: []string{"role:tenant.admin"},
	},
}

// AccountDefinition — individual account in the chart of accounts hierarchy.
var AccountDefinition = def.SystemDefinition{
	Name:        "account",
	Module:      "finance",
	Label:       "Account",
	Description: "GL account in the COA hierarchy. Group accounts have children; leaf accounts hold ledger entries.",
	Fields: []def.FieldDef{
		{
			Name:       "code",
			Type:       def.FieldTypeData,
			Required:   true,
			Unique:     true,
			Immutable:  true,
			Searchable: true,
			MaxLen:     20,
		},
		{
			Name:       "name",
			Type:       def.FieldTypeData,
			Required:   true,
			Searchable: true,
			MaxLen:     255,
		},
		{
			Name:     "account_type",
			Type:     def.FieldTypeSelect,
			Label:    "Account Type",
			Options:  []string{"Asset", "Liability", "Equity", "Income", "Expense"},
			Required: true,
		},
		{
			Name:   "account_subtype",
			Type:   def.FieldTypeData,
			Label:  "Account Subtype",
			MaxLen: 50,
		},
		{
			Name:       "parent_id",
			Type:       def.FieldTypeLink,
			Label:      "Parent Account",
			LinkTarget: "finance_account",
		},
		{
			Name:       "chart_of_accounts_id",
			Type:       def.FieldTypeLink,
			Label:      "Chart of Accounts",
			LinkTarget: "finance_chart_of_accounts",
			Required:   true,
			Immutable:  true,
		},
		{
			Name:       "currency_id",
			Type:       def.FieldTypeLink,
			Label:      "Currency",
			LinkTarget: "finance_currency",
		},
		{
			Name:      "is_group",
			Type:      def.FieldTypeBool,
			Label:     "Is Group",
			Immutable: true,
			Default:   func() any { return false },
		},
		{
			Name:     "balance_type",
			Type:     def.FieldTypeSelect,
			Label:    "Normal Balance",
			Options:  []string{"Debit", "Credit"},
			Required: true,
		},
		{
			Name:    "is_reconcilable",
			Type:    def.FieldTypeBool,
			Label:   "Reconcilable",
			Default: func() any { return false },
		},
		{
			Name:    "active",
			Type:    def.FieldTypeBool,
			Default: func() any { return true },
		},
		{Name: "description", Type: def.FieldTypeSmallText},
	},
	Edges: []def.EdgeDef{
		{
			Name:      "children",
			Target:    "finance_account",
			Type:      def.EdgeOneToMany,
			ForeignKey: "parent_id",
			Label:     "Sub-Accounts",
			OrderBy:   "code ASC",
		},
	},
	Permissions: def.PermissionSet{
		Create: []string{"role:tenant.admin", "role:finance.manager"},
		Read:   []string{"role:tenant.admin", "role:finance.manager", "role:finance.accountant", "role:finance.viewer", "role:tenant.user"},
		Write:  []string{"role:tenant.admin", "role:finance.manager"},
		Delete: []string{"role:tenant.admin"},
	},
}

// CostCenterDefinition — cost center for analytical accounting.
var CostCenterDefinition = def.SystemDefinition{
	Name:        "cost_center",
	Module:      "finance",
	Label:       "Cost Center",
	Description: "Analytical dimension for tracking costs and revenues by department or project.",
	Fields: []def.FieldDef{
		{
			Name:       "name",
			Type:       def.FieldTypeData,
			Required:   true,
			Searchable: true,
			MaxLen:     255,
		},
		{
			Name:      "code",
			Type:      def.FieldTypeData,
			Required:  true,
			Unique:    true,
			Immutable: true,
			MaxLen:    20,
		},
		{
			Name:       "parent_id",
			Type:       def.FieldTypeLink,
			Label:      "Parent Cost Center",
			LinkTarget: "finance_cost_center",
		},
		{
			Name:      "is_group",
			Type:      def.FieldTypeBool,
			Label:     "Is Group",
			Immutable: true,
			Default:   func() any { return false },
		},
		{
			Name:    "active",
			Type:    def.FieldTypeBool,
			Default: func() any { return true },
		},
		{Name: "description", Type: def.FieldTypeSmallText},
	},
	Edges: []def.EdgeDef{
		{
			Name:      "children",
			Target:    "finance_cost_center",
			Type:      def.EdgeOneToMany,
			ForeignKey: "parent_id",
			Label:     "Sub-Cost Centers",
			OrderBy:   "code ASC",
		},
	},
	Permissions: def.PermissionSet{
		Create: []string{"role:tenant.admin", "role:finance.manager"},
		Read:   []string{"role:tenant.admin", "role:finance.manager", "role:finance.accountant", "role:finance.viewer", "role:tenant.user"},
		Write:  []string{"role:tenant.admin", "role:finance.manager"},
		Delete: []string{"role:tenant.admin"},
	},
}
