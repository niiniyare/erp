package finance

import (
	"awo.so/awo/def"
	"awo.so/awo/runtime"
)

// FiscalYearDefinition — fiscal year master with lifecycle locking.
var FiscalYearDefinition = def.SystemDefinition{
	Name:        "fiscal_year",
	Module:      "finance",
	Label:       "Fiscal Year",
	Description: "Fiscal year boundary. Closing a year locks all its accounting periods.",
	Fields: []def.FieldDef{
		{
			Name:       "name",
			Type:       def.FieldTypeData,
			Required:   true,
			Searchable: true,
			MaxLen:     100,
		},
		{Name: "start_date", Type: def.FieldTypeDate, Required: true},
		{Name: "end_date", Type: def.FieldTypeDate, Required: true},
		{
			Name:    "status",
			Type:    def.FieldTypeSelect,
			Options: []string{"open", "closed", "locked"},
			Default: func() any { return "open" },
		},
		{
			Name:    "is_current",
			Type:    def.FieldTypeBool,
			Default: func() any { return false },
		},
	},
	Hooks: def.HookSet{
		BeforeUpdate: []def.BeforeUpdateHook{&FiscalYearTransitionGuard{}},
	},
	Actions: []def.ActionDef{
		{
			Name:           "close",
			Method:         def.ActionMethodPost,
			Label:          "Close Fiscal Year",
			Description:    "Close all periods and lock this fiscal year.",
			Permission:     "finance.fiscal_year.close",
			Icon:           "fa fa-lock",
			ConfirmMessage: "Close fiscal year? This will lock all accounting periods.",
			HandlerFunc:    closeFiscalYear,
		},
	},
	Permissions: def.PermissionSet{
		Create: []string{"finance.fiscal_year.create"},
		Read:   []string{"finance.fiscal_year.read"},
		Write:  []string{"finance.fiscal_year.update"},
		Delete: []string{"finance.fiscal_year.delete"},
	},
}

// AccountingPeriodDefinition — monthly/quarterly accounting period with lock support.
var AccountingPeriodDefinition = def.SystemDefinition{
	Name:        "accounting_period",
	Module:      "finance",
	Label:       "Accounting Period",
	Description: "Monthly or quarterly period within a fiscal year. Journal entries require an open period.",
	Fields: []def.FieldDef{
		{
			Name:       "fiscal_year_id",
			Type:       def.FieldTypeLink,
			Label:      "Fiscal Year",
			LinkTarget: "finance_fiscal_year",
			Required:   true,
			Immutable:  true,
		},
		{
			Name:       "name",
			Type:       def.FieldTypeData,
			Required:   true,
			Searchable: true,
			MaxLen:     100,
		},
		{
			Name:      "period_number",
			Type:      def.FieldTypeInt,
			Label:     "Period Number",
			Required:  true,
			Immutable: true,
		},
		{Name: "start_date", Type: def.FieldTypeDate, Required: true, Immutable: true},
		{Name: "end_date", Type: def.FieldTypeDate, Required: true, Immutable: true},
		{
			Name:    "status",
			Type:    def.FieldTypeSelect,
			Options: []string{"open", "closed", "locked"},
			Default: func() any { return "open" },
		},
		{
			Name:    "is_adjustment",
			Type:    def.FieldTypeBool,
			Label:   "Adjustment Period",
			Default: func() any { return false },
		},
	},
	Hooks: def.HookSet{
		BeforeUpdate: []def.BeforeUpdateHook{&PeriodTransitionGuard{}},
	},
	Actions: []def.ActionDef{
		{
			Name:           "close",
			Method:         def.ActionMethodPost,
			Label:          "Close Period",
			Permission:     "finance.accounting_period.close",
			Icon:           "fa fa-lock",
			ConfirmMessage: "Close this accounting period? No new entries will be allowed.",
			HandlerFunc:    closeAccountingPeriod,
		},
		{
			Name:           "reopen",
			Method:         def.ActionMethodPost,
			Label:          "Reopen Period",
			Permission:     "finance.accounting_period.reopen",
			Icon:           "fa fa-unlock",
			ConfirmMessage: "Reopen this accounting period?",
			HandlerFunc:    reopenAccountingPeriod,
		},
	},
	Permissions: def.PermissionSet{
		Create: []string{"finance.accounting_period.create"},
		Read:   []string{"finance.accounting_period.read"},
		Write:  []string{"finance.accounting_period.update"},
		Delete: []string{"finance.accounting_period.delete"},
	},
}

func closeFiscalYear(ctx *def.ActionContext) (*def.ActionResult, error) {
	// TODO: load fiscal year, close all open periods, set status="closed"
	return nil, &runtime.BusinessError{Code: "not_implemented", Message: "close fiscal year not yet implemented", Status: 501}
}

func closeAccountingPeriod(ctx *def.ActionContext) (*def.ActionResult, error) {
	// TODO: verify no draft journal entries, set status="closed"
	return nil, &runtime.BusinessError{Code: "not_implemented", Message: "close period not yet implemented", Status: 501}
}

func reopenAccountingPeriod(ctx *def.ActionContext) (*def.ActionResult, error) {
	// TODO: verify period is "closed" (not "locked"), set status="open"
	return nil, &runtime.BusinessError{Code: "not_implemented", Message: "reopen period not yet implemented", Status: 501}
}
