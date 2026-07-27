package finance

import "awo.so/awo/def"

// CurrencyDefinition — ISO 4217 currency master.
var CurrencyDefinition = def.SystemDefinition{
	Name:        "currency",
	Module:      "finance",
	Label:       "Currency",
	LabelPlural: "Currencies",
	Description: "ISO 4217 currency master. One record per supported currency per tenant.",
	Fields: []def.FieldDef{
		{
			Name:       "code",
			Type:       def.FieldTypeData,
			Label:      "ISO Code",
			Required:   true,
			Unique:     true,
			Immutable:  true,
			Searchable: true,
			MaxLen:     3,
		},
		{
			Name:       "name",
			Type:       def.FieldTypeData,
			Required:   true,
			Searchable: true,
			MaxLen:     100,
		},
		{Name: "symbol", Type: def.FieldTypeData, MaxLen: 10},
		{
			Name:    "decimal_places",
			Type:    def.FieldTypeInt,
			Default: func() any { return int64(2) },
		},
		{
			Name:    "active",
			Type:    def.FieldTypeBool,
			Default: func() any { return true },
		},
	},
	Permissions: def.PermissionSet{
		Create: []string{"finance.currency.create"},
		Read:   []string{"finance.currency.read"},
		Write:  []string{"finance.currency.update"},
		Delete: []string{"finance.currency.delete"},
	},
}

// ExchangeRateDefinition — point-in-time forex snapshot. Immutable after creation.
var ExchangeRateDefinition = def.SystemDefinition{
	Name:        "exchange_rate",
	Module:      "finance",
	Label:       "Exchange Rate",
	Description: "Point-in-time exchange rate snapshot. Immutable — create a new rate to update.",
	Fields: []def.FieldDef{
		{
			Name:       "from_currency",
			Type:       def.FieldTypeLink,
			Label:      "From Currency",
			LinkTarget: "finance_currency",
			Required:   true,
			Immutable:  true,
		},
		{
			Name:       "to_currency",
			Type:       def.FieldTypeLink,
			Label:      "To Currency",
			LinkTarget: "finance_currency",
			Required:   true,
			Immutable:  true,
		},
		{
			Name:      "rate",
			Type:      def.FieldTypeCurrency,
			Required:  true,
			Immutable: true,
		},
		{
			Name:      "effective_date",
			Type:      def.FieldTypeDate,
			Required:  true,
			Immutable: true,
		},
		{
			Name:      "source",
			Type:      def.FieldTypeData,
			Label:     "Rate Source",
			MaxLen:    100,
			Immutable: true,
		},
	},
	Permissions: def.PermissionSet{
		Create: []string{"finance.exchange_rate.create"},
		Read:   []string{"finance.exchange_rate.read"},
		Write:  []string{}, // immutable — create a new rate instead
		Delete: []string{"finance.exchange_rate.delete"},
	},
}
