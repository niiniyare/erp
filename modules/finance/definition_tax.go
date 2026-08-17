package finance

import "awo.so/awo/def"

// TaxGroupDefinition — tax grouping for VAT, withholding, etc.
var TaxGroupDefinition = def.SystemDefinition{
	Name:        "tax_group",
	Module:      "finance",
	Label:       "Tax Group",
	LabelPlural: "Tax Groups",
	Description: "Tax group for organizing related tax rates (e.g., Standard VAT, Withholding Tax).",
	Icon:        "tag",
	Scope:       def.ScopeTenant,

	Fields: []def.FieldDef{
		{
			Name:       "name",
			Type:       def.FieldTypeData,
			Label:      "Group Name",
			Required:   true,
			Searchable: true,
			MaxLen:     255,
		},
		{
			Name:        "description",
			Type:        def.FieldTypeSmallText,
			Label:       "Description",
		},
		{
			Name:     "tax_type",
			Type:     def.FieldTypeSelect,
			Label:    "Tax Type",
			Required: true,
			Options:  []string{"vat", "gst", "withholding", "excise", "customs", "other"},
		},
		{
			Name:    "is_active",
			Type:    def.FieldTypeBool,
			Label:   "Active",
			Default: func() any { return true },
		},
	},

	Permissions: def.PermissionSet{
		Create: []string{"finance.tax_group.create"},
		Read:   []string{"finance.tax_group.read"},
		Write:  []string{"finance.tax_group.update"},
		Delete: []string{"finance.tax_group.delete"},
	},
}

func init() {
	def.Register(&TaxGroupDefinition)
}

// TaxDefinition — individual tax rate (percentage, fixed, compound).
var TaxDefinition = def.SystemDefinition{
	Name:        "tax",
	Module:      "finance",
	Label:       "Tax",
	LabelPlural: "Taxes",
	Description: "Individual tax rate. Supports percentage, fixed-amount, and compound calculation methods.",
	Icon:        "percent",
	Scope:       def.ScopeTenant,

	Fields: []def.FieldDef{
		{
			Name:       "name",
			Type:       def.FieldTypeData,
			Label:      "Tax Name",
			Required:   true,
			Searchable: true,
			MaxLen:     255,
		},
		{
			Name:       "tax_group",
			Type:       def.FieldTypeLink,
			Label:      "Tax Group",
			LinkTarget: "finance_tax_group",
			Required:   true,
		},
		{
			Name:     "computation",
			Type:     def.FieldTypeSelect,
			Label:    "Computation",
			Required: true,
			Options:  []string{"percentage", "fixed", "percentage_of_tax", "percentage_of_total"},
		},
		{
			Name:    "rate",
			Type:    def.FieldTypeCurrency,
			Label:   "Rate",
			Description: "Percentage (e.g., 20 for 20%) or fixed amount depending on computation method.",
			Required: true,
		},
		{
			Name:       "account_collected",
			Type:       def.FieldTypeLink,
			Label:      "Tax Collected Account",
			Description: "GL account for tax collected on sales.",
			LinkTarget: "finance_account",
		},
		{
			Name:       "account_paid",
			Type:       def.FieldTypeLink,
			Label:      "Tax Paid Account",
			Description: "GL account for tax paid on purchases.",
			LinkTarget: "finance_account",
		},
		{
			Name:     "includes_price",
			Type:     def.FieldTypeBool,
			Label:    "Included in Price",
			Description: "When true, tax is already included in the line price (not added on top).",
		},
		{
			Name:        "effective_date",
			Type:        def.FieldTypeDate,
			Label:       "Effective Date",
		},
		{
			Name:        "expiry_date",
			Type:        def.FieldTypeDate,
			Label:       "Expiry Date",
		},
		{
			Name:    "is_active",
			Type:    def.FieldTypeBool,
			Label:   "Active",
			Default: func() any { return true },
		},
	},

	Permissions: def.PermissionSet{
		Create: []string{"finance.tax.create"},
		Read:   []string{"finance.tax.read"},
		Write:  []string{"finance.tax.update"},
		Delete: []string{"finance.tax.delete"},
	},
}

func init() {
	def.Register(&TaxDefinition)
}
