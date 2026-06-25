// Package finance declares EntityDefinitions for the finance module.
package finance

import (
	"context"

	"awo.so/framework/definition"
)

func init() {
	definition.Register(Account)
	definition.Register(FiscalYear)
	definition.Register(AccountingPeriod)
	definition.Register(CostCenter)
	definition.Register(Currency)
	definition.Register(TaxCode)
	definition.Register(Budget)
}

// ── Shared policy ─────────────────────────────────────────────────

func allowFinance(_ context.Context, viewer definition.ViewerContext, _ definition.Op, _ definition.Record) error {
	if viewer.IsSystem() || viewer.HasRole("finance_manager") || viewer.HasRole("accountant") {
		return definition.ErrAllow
	}
	return definition.ErrDeny
}

func allowFinanceRead(_ context.Context, viewer definition.ViewerContext, op definition.Op, _ definition.Record) error {
	if op == definition.OpRead {
		if viewer.IsSystem() || viewer.HasRole("finance_manager") || viewer.HasRole("accountant") || viewer.HasRole("auditor") {
			return definition.ErrAllow
		}
		return definition.ErrDeny
	}
	// Writes require finance_manager or higher.
	if viewer.IsSystem() || viewer.HasRole("finance_manager") {
		return definition.ErrAllow
	}
	return definition.ErrDeny
}

// ── Definitions ───────────────────────────────────────────────────

// Account is the chart-of-accounts entity.
var Account = &definition.EntityDefinition{
	Name:       "finance_account",
	Label:      "Account",
	Table:      "finance_accounts",
	Module:     "Finance",
	SoftDelete: true,
	Audited:    true,
	Fields: []*definition.FieldDef{
		definition.Field("account_code").OfType(definition.FieldTypeSmallText).WithLabel("Code").
			RequiredField().UniqueField().SearchableField().WithMaxLen(30),
		definition.Field("account_name").OfType(definition.FieldTypeSmallText).WithLabel("Name").
			RequiredField().SearchableField().WithMaxLen(255),
		definition.Field("account_description").OfType(definition.FieldTypeLongText).WithLabel("Description"),
		definition.Field("root_type").OfType(definition.FieldTypeSelect).WithLabel("Root Type").
			RequiredField().ImmutableField().
			WithOptions("ASSET", "LIABILITY", "EQUITY", "REVENUE", "EXPENSE"),
		definition.Field("account_type").OfType(definition.FieldTypeSmallText).WithLabel("Account Type").
			WithMaxLen(50),
		definition.Field("normal_balance").OfType(definition.FieldTypeSelect).WithLabel("Normal Balance").
			RequiredField().WithOptions("DEBIT", "CREDIT"),
		definition.Field("currency_code").OfType(definition.FieldTypeSmallText).WithLabel("Currency").
			WithMaxLen(3).WithDefault("USD"),
		definition.Field("is_active").OfType(definition.FieldTypeBool).WithLabel("Active").WithDefault(true),
		definition.Field("is_system_account").OfType(definition.FieldTypeBool).WithLabel("System Account").
			WithDefault(false).ReadOnlyField(),
		definition.Field("allow_manual_entries").OfType(definition.FieldTypeBool).WithLabel("Allow Manual Entries").
			WithDefault(true),
		definition.Field("current_balance").OfType(definition.FieldTypeCurrency).WithLabel("Current Balance").
			ReadOnlyField(),
		definition.Field("parent_account_id").OfType(definition.FieldTypeLink).WithLabel("Parent Account").
			LinksTo("finance_account"),
	},
	Policies: []definition.PolicyDef{
		definition.Policy(definition.OpAll, allowFinanceRead),
	},
}

// FiscalYear represents an accounting year.
var FiscalYear = &definition.EntityDefinition{
	Name:   "finance_fiscal_year",
	Label:  "Fiscal Year",
	Table:  "finance_fiscal_years",
	Module: "Finance",
	Fields: []*definition.FieldDef{
		definition.Field("name").OfType(definition.FieldTypeSmallText).WithLabel("Name").
			RequiredField().WithMaxLen(100),
		definition.Field("start_date").OfType(definition.FieldTypeDate).WithLabel("Start Date").RequiredField(),
		definition.Field("end_date").OfType(definition.FieldTypeDate).WithLabel("End Date").RequiredField(),
		definition.Field("is_closed").OfType(definition.FieldTypeBool).WithLabel("Closed").WithDefault(false),
		definition.Field("is_locked").OfType(definition.FieldTypeBool).WithLabel("Locked").WithDefault(false),
	},
	Policies: []definition.PolicyDef{
		definition.Policy(definition.OpAll, allowFinanceRead),
	},
}

// AccountingPeriod is a sub-period within a fiscal year.
var AccountingPeriod = &definition.EntityDefinition{
	Name:   "finance_accounting_period",
	Label:  "Accounting Period",
	Table:  "finance_accounting_periods",
	Module: "Finance",
	Fields: []*definition.FieldDef{
		definition.Field("fiscal_year_id").OfType(definition.FieldTypeLink).WithLabel("Fiscal Year").
			RequiredField().LinksTo("finance_fiscal_year"),
		definition.Field("period_number").OfType(definition.FieldTypeInt).WithLabel("Period #").RequiredField(),
		definition.Field("name").OfType(definition.FieldTypeSmallText).WithLabel("Name").
			RequiredField().WithMaxLen(50),
		definition.Field("start_date").OfType(definition.FieldTypeDate).WithLabel("Start").RequiredField(),
		definition.Field("end_date").OfType(definition.FieldTypeDate).WithLabel("End").RequiredField(),
		definition.Field("status").OfType(definition.FieldTypeSelect).WithLabel("Status").
			WithOptions("OPEN", "CLOSED", "LOCKED").WithDefault("OPEN"),
	},
	Policies: []definition.PolicyDef{
		definition.Policy(definition.OpAll, allowFinanceRead),
	},
}

// CostCenter for expense allocation.
var CostCenter = &definition.EntityDefinition{
	Name:   "finance_cost_center",
	Label:  "Cost Center",
	Table:  "finance_cost_centers",
	Module: "Finance",
	Fields: []*definition.FieldDef{
		definition.Field("code").OfType(definition.FieldTypeSmallText).WithLabel("Code").
			RequiredField().UniqueField().WithMaxLen(30),
		definition.Field("name").OfType(definition.FieldTypeSmallText).WithLabel("Name").
			RequiredField().SearchableField().WithMaxLen(150),
		definition.Field("description").OfType(definition.FieldTypeLongText).WithLabel("Description"),
		definition.Field("parent_id").OfType(definition.FieldTypeLink).WithLabel("Parent").
			LinksTo("finance_cost_center"),
		definition.Field("is_active").OfType(definition.FieldTypeBool).WithLabel("Active").WithDefault(true),
	},
	Policies: []definition.PolicyDef{
		definition.Policy(definition.OpAll, allowFinance),
	},
}

// Currency is the supported currencies catalogue (global).
var Currency = &definition.EntityDefinition{
	Name:    "finance_currency",
	Label:   "Currency",
	Table:   "finance_currencies",
	Module:  "Finance",
	Global:  false, // tenant-scoped enabled currencies
	Fields: []*definition.FieldDef{
		definition.Field("code").OfType(definition.FieldTypeSmallText).WithLabel("Code").
			RequiredField().UniqueField().ImmutableField().WithMaxLen(3),
		definition.Field("name").OfType(definition.FieldTypeSmallText).WithLabel("Name").
			RequiredField().WithMaxLen(100),
		definition.Field("symbol").OfType(definition.FieldTypeSmallText).WithLabel("Symbol").
			WithMaxLen(10),
		definition.Field("decimal_places").OfType(definition.FieldTypeInt).WithLabel("Decimal Places").
			WithDefault(2),
		definition.Field("is_active").OfType(definition.FieldTypeBool).WithLabel("Active").WithDefault(true),
		definition.Field("is_base").OfType(definition.FieldTypeBool).WithLabel("Base Currency").WithDefault(false),
	},
	Policies: []definition.PolicyDef{
		definition.Policy(definition.OpRead, definition.AllowAll),
		definition.Policy(definition.OpWrite, allowFinance),
	},
}

// TaxCode defines tax rates and GL account mappings.
var TaxCode = &definition.EntityDefinition{
	Name:   "finance_tax_code",
	Label:  "Tax Code",
	Table:  "finance_tax_codes",
	Module: "Finance",
	Fields: []*definition.FieldDef{
		definition.Field("code").OfType(definition.FieldTypeSmallText).WithLabel("Code").
			RequiredField().UniqueField().WithMaxLen(20),
		definition.Field("name").OfType(definition.FieldTypeSmallText).WithLabel("Name").
			RequiredField().SearchableField().WithMaxLen(150),
		definition.Field("tax_type").OfType(definition.FieldTypeSelect).WithLabel("Tax Type").
			WithOptions("VAT", "GST", "SALES", "WITHHOLDING", "EXCISE"),
		definition.Field("tax_rate").OfType(definition.FieldTypeCurrency).WithLabel("Rate (%)").
			RequiredField(),
		definition.Field("calculation_method").OfType(definition.FieldTypeSelect).WithLabel("Calculation").
			WithOptions("INCLUSIVE", "EXCLUSIVE"),
		definition.Field("effective_date").OfType(definition.FieldTypeDate).WithLabel("Effective Date"),
		definition.Field("expiry_date").OfType(definition.FieldTypeDate).WithLabel("Expiry Date"),
		definition.Field("is_active").OfType(definition.FieldTypeBool).WithLabel("Active").WithDefault(true),
	},
	Policies: []definition.PolicyDef{
		definition.Policy(definition.OpAll, allowFinanceRead),
	},
}

// Budget tracks approved spending limits per fiscal year.
var Budget = &definition.EntityDefinition{
	Name:    "finance_budget",
	Label:   "Budget",
	Table:   "finance_budgets",
	Module:  "Finance",
	Audited: true,
	Fields: []*definition.FieldDef{
		definition.Field("name").OfType(definition.FieldTypeSmallText).WithLabel("Name").
			RequiredField().WithMaxLen(200),
		definition.Field("description").OfType(definition.FieldTypeLongText).WithLabel("Description"),
		definition.Field("fiscal_year_id").OfType(definition.FieldTypeLink).WithLabel("Fiscal Year").
			RequiredField().LinksTo("finance_fiscal_year"),
		definition.Field("budget_type").OfType(definition.FieldTypeSelect).WithLabel("Type").
			WithOptions("OPERATIONAL", "CAPITAL", "CASH_FLOW").WithDefault("OPERATIONAL"),
		definition.Field("status").OfType(definition.FieldTypeSelect).WithLabel("Status").
			WithOptions("DRAFT", "SUBMITTED", "APPROVED", "REJECTED").WithDefault("DRAFT"),
		definition.Field("currency_code").OfType(definition.FieldTypeSmallText).WithLabel("Currency").
			WithMaxLen(3).WithDefault("USD"),
	},
	Policies: []definition.PolicyDef{
		definition.Policy(definition.OpAll, allowFinanceRead),
	},
}
