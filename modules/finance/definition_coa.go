package finance

import "awo.so/awo/def"

// ChartOfAccountsDefinition — COA master header.
// A tenant may have multiple COA setups (e.g., one per reporting standard).
var ChartOfAccountsDefinition = def.SystemDefinition{
	Name:        "chart_of_accounts",
	Module:      "finance",
	Label:       "Chart of Accounts",
	LabelPlural: "Charts of Accounts",
	Description: "Chart of Accounts master header. Groups GL accounts under a reporting standard (e.g., GAAP, IFRS).",
	Icon:        "list-tree",
	Scope:       def.ScopeTenant,

	Fields: []def.FieldDef{
		{
			Name:       "code",
			Type:       def.FieldTypeData,
			Label:      "COA Code",
			Required:   true,
			Unique:     true,
			Immutable:  true,
			MaxLen:     50,
			Searchable: true,
		},
		{
			Name:       "name",
			Type:       def.FieldTypeData,
			Label:      "COA Name",
			Required:   true,
			Searchable: true,
			MaxLen:     255,
		},
		{
			Name:        "description",
			Type:        def.FieldTypeSmallText,
			Label:       "Description",
			Description: "Purpose and scope of this chart of accounts.",
		},
		{
			Name:        "currency",
			Type:        def.FieldTypeLink,
			Label:       "Default Currency",
			Description: "Functional currency for this COA.",
			LinkTarget:  "finance_currency",
			Required:    true,
		},
		{
			Name:    "is_active",
			Type:    def.FieldTypeBool,
			Label:   "Active",
			Default: func() any { return true },
		},
	},

	Permissions: def.PermissionSet{
		Create: []string{"finance.chart_of_accounts.create"},
		Read:   []string{"finance.chart_of_accounts.read"},
		Write:  []string{"finance.chart_of_accounts.update"},
		Delete: []string{"finance.chart_of_accounts.delete"},
	},
}

func init() {
	def.Register(&ChartOfAccountsDefinition)
}

// AccountDefinition — GL account in a COA hierarchy.
// Account type drives debit/credit normal balance and P&L vs balance sheet placement.
var AccountDefinition = def.SystemDefinition{
	Name:        "account",
	Module:      "finance",
	Label:       "Account",
	LabelPlural: "Accounts",
	Description: "General Ledger account within a Chart of Accounts. Account type determines debit/credit normal balance.",
	Icon:        "book-open",
	Scope:       def.ScopeTenant,

	Fields: []def.FieldDef{
		{
			Name:       "code",
			Type:       def.FieldTypeData,
			Label:      "Account Code",
			Required:   true,
			Unique:     true,
			MaxLen:     50,
			Searchable: true,
		},
		{
			Name:       "name",
			Type:       def.FieldTypeData,
			Label:      "Account Name",
			Required:   true,
			Searchable: true,
			MaxLen:     255,
		},
		{
			Name:       "chart_of_accounts",
			Type:       def.FieldTypeLink,
			Label:      "Chart of Accounts",
			LinkTarget: "finance_chart_of_accounts",
			Required:   true,
		},
		{
			Name:    "parent_account",
			Type:    def.FieldTypeLink,
			Label:   "Parent Account",
			Description: "Parent account for hierarchical COA structure. Nil = root account.",
			LinkTarget: "finance_account",
		},
		{
			Name:     "account_type",
			Type:     def.FieldTypeSelect,
			Label:    "Account Type",
			Required: true,
			Immutable: true,
			Options:  []string{"asset", "liability", "equity", "income", "expense"},
		},
		{
			Name:     "account_subtype",
			Type:     def.FieldTypeSelect,
			Label:    "Account Subtype",
			Options:  []string{"current_asset", "fixed_asset", "current_liability", "long_term_liability", "retained_earnings", "revenue", "cost_of_goods_sold", "operating_expense", "other"},
		},
		{
			Name:        "currency",
			Type:        def.FieldTypeLink,
			Label:       "Currency",
			Description: "Account currency. Defaults to COA functional currency.",
			LinkTarget:  "finance_currency",
		},
		{
			Name:    "is_active",
			Type:    def.FieldTypeBool,
			Label:   "Active",
			Default: func() any { return true },
		},
		{
			Name:        "allow_direct_posting",
			Type:        def.FieldTypeBool,
			Label:       "Allow Direct Posting",
			Description: "When false, this account is a grouping account only; journal entries cannot post directly to it.",
			Default:     func() any { return true },
		},
		{
			Name:        "description",
			Type:        def.FieldTypeSmallText,
			Label:       "Description",
		},
	},

	Permissions: def.PermissionSet{
		Create: []string{"finance.account.create"},
		Read:   []string{"finance.account.read"},
		Write:  []string{"finance.account.update"},
		Delete: []string{"finance.account.delete"},
	},
}

func init() {
	def.Register(&AccountDefinition)
}
