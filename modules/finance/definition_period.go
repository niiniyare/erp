package finance

import "awo.so/awo/def"

// FiscalYearDefinition — fiscal year with lifecycle locking.
// Closing a fiscal year locks all accounting periods within it.
var FiscalYearDefinition = def.SystemDefinition{
	Name:        "fiscal_year",
	Module:      "finance",
	Label:       "Fiscal Year",
	LabelPlural: "Fiscal Years",
	Description: "Fiscal year with lifecycle locking. Closing a fiscal year locks all accounting periods within it.",
	Icon:        "calendar",
	Scope:       def.ScopeTenant,

	Fields: []def.FieldDef{
		{
			Name:       "name",
			Type:       def.FieldTypeData,
			Label:      "Fiscal Year Name",
			Required:   true,
			Unique:     true,
			Searchable: true,
			MaxLen:     100,
		},
		{
			Name:      "start_date",
			Type:      def.FieldTypeDate,
			Label:     "Start Date",
			Required:  true,
			Immutable: true,
		},
		{
			Name:      "end_date",
			Type:      def.FieldTypeDate,
			Label:     "End Date",
			Required:  true,
			Immutable: true,
		},
		{
			Name:     "status",
			Type:     def.FieldTypeSelect,
			Label:    "Status",
			Required: true,
			Options:  []string{"draft", "active", "closing", "closed"},
			Default:  func() any { return "draft" },
			ReadOnly: true,
		},
		{
			Name:        "closed_at",
			Type:        def.FieldTypeDateTime,
			Label:       "Closed At",
			ReadOnly:    true,
		},
	},

	Actions: []def.ActionDef{
		{
			Name:           "activate",
			Label:          "Activate",
			Description:    "Mark the fiscal year as active and open for accounting period creation.",
			ConfirmMessage: "Activate this fiscal year?",
			Icon:           "play",
			HandlerFunc:    stubAction("activate"),
		},
		{
			Name:           "close",
			Label:          "Close",
			Description:    "Close the fiscal year. All accounting periods will be locked.",
			ConfirmMessage: "Close this fiscal year? All accounting periods will be locked and no further entries will be accepted.",
			Icon:           "lock",
			WorkflowEvent:  def.EventType("close"),
			HandlerFunc:    stubAction("close"),
		},
	},

	Permissions: def.PermissionSet{
		Create: []string{"finance.fiscal_year.create"},
		Read:   []string{"finance.fiscal_year.read"},
		Write:  []string{"finance.fiscal_year.update"},
		Delete: []string{"finance.fiscal_year.delete"},
		Actions: map[string][]string{
			"activate": {"finance.fiscal_year.activate"},
			"close":    {"finance.fiscal_year.close"},
		},
	},
}

func init() {
	def.Register(&FiscalYearDefinition)
}

// AccountingPeriodDefinition — monthly/quarterly period within a fiscal year.
// Must be open for journal entries to post.
var AccountingPeriodDefinition = def.SystemDefinition{
	Name:        "accounting_period",
	Module:      "finance",
	Label:       "Accounting Period",
	LabelPlural: "Accounting Periods",
	Description: "Accounting period (monthly/quarterly). Journal entries can only post to open periods.",
	Icon:        "calendar-days",
	Scope:       def.ScopeTenant,

	Fields: []def.FieldDef{
		{
			Name:       "name",
			Type:       def.FieldTypeData,
			Label:      "Period Name",
			Required:   true,
			Searchable: true,
			MaxLen:     100,
		},
		{
			Name:       "fiscal_year",
			Type:       def.FieldTypeLink,
			Label:      "Fiscal Year",
			LinkTarget: "finance_fiscal_year",
			Required:   true,
			Immutable:  true,
		},
		{
			Name:      "start_date",
			Type:      def.FieldTypeDate,
			Label:     "Start Date",
			Required:  true,
			Immutable: true,
		},
		{
			Name:      "end_date",
			Type:      def.FieldTypeDate,
			Label:     "End Date",
			Required:  true,
			Immutable: true,
		},
		{
			Name:     "period_type",
			Type:     def.FieldTypeSelect,
			Label:    "Period Type",
			Required: true,
			Immutable: true,
			Options:  []string{"monthly", "quarterly", "annual", "custom"},
			Default:  func() any { return "monthly" },
		},
		{
			Name:     "status",
			Type:     def.FieldTypeSelect,
			Label:    "Status",
			Required: true,
			Options:  []string{"draft", "open", "closed", "locked"},
			Default:  func() any { return "draft" },
			ReadOnly: true,
		},
		{
			Name:     "closed_at",
			Type:     def.FieldTypeDateTime,
			Label:    "Closed At",
			ReadOnly: true,
		},
	},

	Actions: []def.ActionDef{
		{
			Name:           "open",
			Label:          "Open",
			Description:    "Open the period for journal entry posting.",
			ConfirmMessage: "Open this accounting period?",
			Icon:           "unlock",
			HandlerFunc:    stubAction("open"),
		},
		{
			Name:           "close",
			Label:          "Close",
			Description:    "Close the period. No further journal entries will be accepted.",
			ConfirmMessage: "Close this accounting period? No further journal entries will be accepted.",
			Icon:           "lock",
			HandlerFunc:    stubAction("close"),
		},
	},

	Permissions: def.PermissionSet{
		Create: []string{"finance.accounting_period.create"},
		Read:   []string{"finance.accounting_period.read"},
		Write:  []string{"finance.accounting_period.update"},
		Delete: []string{"finance.accounting_period.delete"},
		Actions: map[string][]string{
			"open":  {"finance.accounting_period.open"},
			"close": {"finance.accounting_period.close"},
		},
	},
}

func init() {
	def.Register(&AccountingPeriodDefinition)
}
