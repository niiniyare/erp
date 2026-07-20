package finance

import "awo.so/awo/def"

// TaxGroupDefinition — logical grouping of taxes (e.g. "VAT", "Withholding").
var TaxGroupDefinition = def.SystemDefinition{
	Name:        "tax_group",
	Module:      "finance",
	Label:       "Tax Group",
	Description: "Groups related taxes for reporting (e.g. VAT, Withholding Tax).",
	Fields: []def.FieldDef{
		{
			Name:       "name",
			Type:       def.FieldTypeData,
			Required:   true,
			Searchable: true,
			MaxLen:     100,
		},
		{Name: "description", Type: def.FieldTypeSmallText},
		{
			Name:    "active",
			Type:    def.FieldTypeBool,
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

// TaxDefinition — individual tax rate configuration.
var TaxDefinition = def.SystemDefinition{
	Name:        "tax",
	Module:      "finance",
	Label:       "Tax",
	Description: "Tax rate definition. Rate is numeric (e.g. 16.0000 = 16% VAT).",
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
			Name:       "tax_group_id",
			Type:       def.FieldTypeLink,
			Label:      "Tax Group",
			LinkTarget: "finance_tax_group",
		},
		{
			Name:     "tax_type",
			Type:     def.FieldTypeSelect,
			Label:    "Tax Type",
			Options:  []string{"percentage", "fixed", "compound"},
			Required: true,
			Default:  func() any { return "percentage" },
		},
		{
			// rate stores numeric value: 16.0000 = 16%, or 100.0000 = KES 100 fixed
			Name:     "rate",
			Type:     def.FieldTypeCurrency,
			Required: true,
		},
		{
			Name:       "account_id",
			Type:       def.FieldTypeLink,
			Label:      "Tax Account",
			LinkTarget: "finance_account",
			Required:   true,
		},
		{
			Name:    "is_inclusive",
			Type:    def.FieldTypeBool,
			Label:   "Inclusive of Tax",
			Default: func() any { return false },
		},
		{
			Name:    "applies_to",
			Type:    def.FieldTypeSelect,
			Label:   "Applies To",
			Options: []string{"sales", "purchase", "both"},
			Default: func() any { return "both" },
		},
		{
			Name:    "active",
			Type:    def.FieldTypeBool,
			Default: func() any { return true },
		},
		{Name: "description", Type: def.FieldTypeSmallText},
	},
	Permissions: def.PermissionSet{
		Create: []string{"finance.tax.create"},
		Read:   []string{"finance.tax.read"},
		Write:  []string{"finance.tax.update"},
		Delete: []string{"finance.tax.delete"},
	},
}

// TaxEntryDefinition — immutable tax record per transaction. MANDATORY system entity.
//
// Written exclusively by the framework when tax is applied on invoice/payment posting.
// Required for KRA eTIMS compliance.
var TaxEntryDefinition = def.SystemDefinition{
	Name:        "tax_entry",
	Module:      "finance",
	Label:       "Tax Entry",
	LabelPlural: "Tax Entries",
	Description: "Immutable tax record created at posting time. KRA eTIMS compliance record.",
	Fields: []def.FieldDef{
		{
			Name:      "entry_number",
			Type:      def.FieldTypeNamingSeries,
			Label:     "Entry Number",
			Series:    "TAX-{YYYY}-{SEQ:6}",
			Immutable: true,
		},
		{
			Name:       "tax_id",
			Type:       def.FieldTypeLink,
			Label:      "Tax",
			LinkTarget: "finance_tax",
			Required:   true,
			Immutable:  true,
		},
		{
			Name:       "journal_entry_id",
			Type:       def.FieldTypeLink,
			Label:      "Journal Entry",
			LinkTarget: "finance_journal_entry",
			Required:   true,
			Immutable:  true,
		},
		{
			Name:      "reference_type",
			Type:      def.FieldTypeData,
			Label:     "Document Type",
			MaxLen:    100,
			Immutable: true,
		},
		{
			Name:      "reference_id",
			Type:      def.FieldTypeData,
			Label:     "Document ID",
			MaxLen:    36,
			Immutable: true,
		},
		{
			Name:      "tax_base_amount",
			Type:      def.FieldTypeCurrency,
			Label:     "Tax Base Amount",
			Required:  true,
			Immutable: true,
		},
		{
			Name:      "tax_amount",
			Type:      def.FieldTypeCurrency,
			Label:     "Tax Amount",
			Required:  true,
			Immutable: true,
		},
		{Name: "posting_date", Type: def.FieldTypeDate, Required: true, Immutable: true},
		{
			Name:      "is_reverse_charge",
			Type:      def.FieldTypeBool,
			Label:     "Reverse Charge",
			Immutable: true,
			Default:   func() any { return false },
		},
	},
	// Empty = framework writes; API cannot create/modify/delete tax entries
	Permissions: def.PermissionSet{
		Create: []string{},
		Read:   []string{"finance.tax_entry.read"},
		Write:  []string{},
		Delete: []string{},
	},
}
