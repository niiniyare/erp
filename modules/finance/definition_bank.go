package finance

import "awo.so/awo/def"

// BankAccountDefinition — bank account master linked to a GL account.
var BankAccountDefinition = def.SystemDefinition{
	Name:        "bank_account",
	Module:      "finance",
	Label:       "Bank Account",
	Description: "Tenant bank account. Each account maps to a GL account for double-entry.",
	Fields: []def.FieldDef{
		{
			Name:       "account_name",
			Type:       def.FieldTypeData,
			Label:      "Account Name",
			Required:   true,
			Searchable: true,
			MaxLen:     255,
		},
		{Name: "bank_name", Type: def.FieldTypeData, Label: "Bank Name", Required: true, MaxLen: 255},
		{
			Name:      "account_number",
			Type:      def.FieldTypeData,
			Label:     "Account Number",
			Required:  true,
			Unique:    true,
			Sensitive: true,
			MaxLen:    50,
		},
		{Name: "swift_code", Type: def.FieldTypeData, Label: "SWIFT Code", MaxLen: 11},
		{Name: "branch_code", Type: def.FieldTypeData, Label: "Branch Code", MaxLen: 20},
		{
			Name:       "currency_id",
			Type:       def.FieldTypeLink,
			Label:      "Currency",
			LinkTarget: "finance_currency",
			Required:   true,
		},
		{
			Name:       "gl_account_id",
			Type:       def.FieldTypeLink,
			Label:      "GL Account",
			LinkTarget: "finance_account",
			Required:   true,
		},
		{
			Name:    "active",
			Type:    def.FieldTypeBool,
			Default: func() any { return true },
		},
		{
			Name:     "balance",
			Type:     def.FieldTypeCurrency,
			ReadOnly: true,
			Default:  func() any { return "0.0000" },
		},
	},
	Permissions: def.PermissionSet{
		Create: []string{"role:tenant.admin", "role:finance.manager"},
		Read:   []string{"role:tenant.admin", "role:finance.manager", "role:finance.accountant", "role:finance.viewer"},
		Write:  []string{"role:tenant.admin", "role:finance.manager"},
		Delete: []string{"role:tenant.admin"},
	},
}

// BankTransactionDefinition — imported bank statement line.
// Immutable amount/date once created; status updated during reconciliation.
var BankTransactionDefinition = def.SystemDefinition{
	Name:        "bank_transaction",
	Module:      "finance",
	Label:       "Bank Transaction",
	Description: "Imported bank statement line. Reconcile by matching to a payment or journal entry.",
	Fields: []def.FieldDef{
		{
			Name:      "transaction_number",
			Type:      def.FieldTypeNamingSeries,
			Label:     "Transaction Number",
			Series:    "BT-{YYYY}-{SEQ:6}",
			Immutable: true,
		},
		{
			Name:       "bank_account_id",
			Type:       def.FieldTypeLink,
			Label:      "Bank Account",
			LinkTarget: "finance_bank_account",
			Required:   true,
			Immutable:  true,
		},
		{Name: "transaction_date", Type: def.FieldTypeDate, Label: "Transaction Date", Required: true, Immutable: true},
		{Name: "value_date", Type: def.FieldTypeDate, Label: "Value Date", Immutable: true},
		{
			Name:      "transaction_type",
			Type:      def.FieldTypeSelect,
			Label:     "Type",
			Options:   []string{"credit", "debit"},
			Required:  true,
			Immutable: true,
		},
		{
			Name:      "amount",
			Type:      def.FieldTypeCurrency,
			Required:  true,
			Immutable: true,
		},
		{
			Name:       "currency_id",
			Type:       def.FieldTypeLink,
			Label:      "Currency",
			LinkTarget: "finance_currency",
			Required:   true,
			Immutable:  true,
		},
		{Name: "description", Type: def.FieldTypeData, MaxLen: 255, Immutable: true},
		{Name: "reference", Type: def.FieldTypeData, Label: "Bank Reference", MaxLen: 100, Immutable: true},
		{
			Name:    "status",
			Type:    def.FieldTypeSelect,
			Options: []string{"unreconciled", "matched", "reconciled", "ignored"},
			Default: func() any { return "unreconciled" },
		},
		{
			Name:       "payment_id",
			Type:       def.FieldTypeLink,
			Label:      "Matched Payment",
			LinkTarget: "finance_payment",
		},
		{
			Name:       "journal_entry_id",
			Type:       def.FieldTypeLink,
			Label:      "Journal Entry",
			LinkTarget: "finance_journal_entry",
			ReadOnly:   true,
		},
	},
	Permissions: def.PermissionSet{
		Create: []string{"role:tenant.admin", "role:finance.manager", "role:finance.accountant"},
		Read:   []string{"role:tenant.admin", "role:finance.manager", "role:finance.accountant", "role:finance.viewer"},
		Write:  []string{"role:tenant.admin", "role:finance.manager", "role:finance.accountant"},
		Delete: []string{},
	},
}
