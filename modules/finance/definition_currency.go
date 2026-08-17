// Package finance declares all Finance module entity definitions.
//
// Entities are registered via init() and compiled by the framework at bootstrap.
// All entities use the "finance" module prefix per naming convention.
//
// Entity inventory:
//   - finance_currency          ISO 4217 currency master
//   - finance_exchange_rate     Point-in-time forex snapshots (immutable)
//   - finance_chart_of_accounts COA master header
//   - finance_account           GL account in COA hierarchy
//   - finance_bank_account      Tenant bank account mapped to GL
//   - finance_bank_transaction  Imported bank statement lines (immutable amount/date)
//   - finance_journal           Journal master
//   - finance_journal_entry     Double-entry header with state machine
//   - finance_payment           Payment record with state machine
//   - finance_payment_method    Payment method master mapped to GL account
//   - finance_tax_group         Tax grouping for VAT, withholding
//   - finance_tax               Individual tax rate
//   - finance_fiscal_year       Fiscal year with lifecycle locking
//   - finance_accounting_period Monthly/quarterly period
package finance

import "awo.so/awo/def"

// CurrencyDefinition — ISO 4217 currency master.
// One record per currency; shared across tenants (system scope).
var CurrencyDefinition = def.SystemDefinition{
	Name:        "currency",
	Module:      "finance",
	Label:       "Currency",
	LabelPlural: "Currencies",
	Description: "ISO 4217 currency master. Defines the currencies available for transactions across all finance entities.",
	Icon:        "money",
	Scope:       def.ScopeSystem,

	Fields: []def.FieldDef{
		{
			Name:        "code",
			Type:        def.FieldTypeData,
			Label:       "Currency Code",
			Description: "ISO 4217 three-letter currency code (e.g., USD, EUR, GBP).",
			Required:    true,
			Unique:      true,
			Immutable:   true,
			MaxLen:      3,
		},
		{
			Name:        "name",
			Type:        def.FieldTypeData,
			Label:       "Currency Name",
			Description: "Full name of the currency (e.g., United States Dollar).",
			Required:    true,
			Searchable:  true,
			MaxLen:      100,
		},
		{
			Name:        "symbol",
			Type:        def.FieldTypeData,
			Label:       "Symbol",
			Description: "Currency symbol (e.g., $, €, £).",
			Required:    true,
			MaxLen:      10,
		},
		{
			Name:        "decimal_places",
			Type:        def.FieldTypeInt,
			Label:       "Decimal Places",
			Description: "Number of decimal places for this currency (e.g., 2 for USD, 0 for JPY).",
			Required:    true,
		},
		{
			Name:        "is_active",
			Type:        def.FieldTypeBool,
			Label:       "Active",
			Description: "Inactive currencies cannot be selected in new transactions.",
			Default:     func() any { return true },
		},
	},

	Permissions: def.PermissionSet{
		Create: []string{"finance.currency.create"},
		Read:   []string{"finance.currency.read"},
		Write:  []string{"finance.currency.update"},
		Delete: []string{"finance.currency.delete"},
	},

	DisableAudit: false,
}

func init() {
	def.Register(&CurrencyDefinition)
}

// ExchangeRateDefinition — point-in-time forex snapshot.
// Records are immutable once created to preserve historical accuracy.
var ExchangeRateDefinition = def.SystemDefinition{
	Name:        "exchange_rate",
	Module:      "finance",
	Label:       "Exchange Rate",
	LabelPlural: "Exchange Rates",
	Description: "Point-in-time foreign exchange rate snapshot. Immutable after creation to preserve historical accuracy.",
	Icon:        "trending-up",
	Scope:       def.ScopeTenant,

	Fields: []def.FieldDef{
		{
			Name:        "from_currency",
			Type:        def.FieldTypeLink,
			Label:       "From Currency",
			Description: "Base currency being converted from.",
			LinkTarget:  "finance_currency",
			Required:    true,
			Immutable:   true,
		},
		{
			Name:        "to_currency",
			Type:        def.FieldTypeLink,
			Label:       "To Currency",
			Description: "Target currency being converted to.",
			LinkTarget:  "finance_currency",
			Required:    true,
			Immutable:   true,
		},
		{
			Name:        "rate",
			Type:        def.FieldTypeCurrency,
			Label:       "Rate",
			Description: "Exchange rate: 1 unit of from_currency = rate units of to_currency.",
			Required:    true,
			Immutable:   true,
		},
		{
			Name:        "effective_date",
			Type:        def.FieldTypeDate,
			Label:       "Effective Date",
			Description: "Date from which this rate is effective.",
			Required:    true,
			Immutable:   true,
		},
		{
			Name:        "source",
			Type:        def.FieldTypeData,
			Label:       "Source",
			Description: "Rate source identifier (e.g., 'manual', 'ECB', 'OANDA').",
			MaxLen:      100,
		},
	},

	Permissions: def.PermissionSet{
		Create: []string{"finance.exchange_rate.create"},
		Read:   []string{"finance.exchange_rate.read"},
		Delete: []string{"finance.exchange_rate.delete"},
	},

	DisableAudit: false,
}

func init() {
	def.Register(&ExchangeRateDefinition)
}
