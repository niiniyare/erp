package finance

import "awo.so/awo/def"

// BankAccountDefinition — tenant bank account mapped to a GL account.
var BankAccountDefinition = def.SystemDefinition{
	Name:        "bank_account",
	Module:      "finance",
	Label:       "Bank Account",
	LabelPlural: "Bank Accounts",
	Description: "Tenant bank account. Maps a physical bank account to a GL account for reconciliation.",
	Icon:        "bank",
	Scope:       def.ScopeTenant,

	Fields: []def.FieldDef{
		{
			Name:       "account_name",
			Type:       def.FieldTypeData,
			Label:      "Account Name",
			Required:   true,
			Searchable: true,
			MaxLen:     255,
		},
		{
			Name:      "account_number",
			Type:      def.FieldTypeData,
			Label:     "Account Number",
			Required:  true,
			Sensitive: true,
			MaxLen:    100,
		},
		{
			Name:   "bank_name",
			Type:   def.FieldTypeData,
			Label:  "Bank Name",
			Required: true,
			MaxLen: 255,
		},
		{
			Name:   "bank_code",
			Type:   def.FieldTypeData,
			Label:  "Bank Code / BIC / SWIFT",
			MaxLen: 50,
		},
		{
			Name:   "iban",
			Type:   def.FieldTypeData,
			Label:  "IBAN",
			MaxLen: 34,
		},
		{
			Name:       "currency",
			Type:       def.FieldTypeLink,
			Label:      "Currency",
			LinkTarget: "finance_currency",
			Required:   true,
		},
		{
			Name:        "gl_account",
			Type:        def.FieldTypeLink,
			Label:       "GL Account",
			Description: "General Ledger account that mirrors this bank account.",
			LinkTarget:  "finance_account",
			Required:    true,
		},
		{
			Name:    "is_active",
			Type:    def.FieldTypeBool,
			Label:   "Active",
			Default: func() any { return true },
		},
		{
			Name:    "opening_balance",
			Type:    def.FieldTypeCurrency,
			Label:   "Opening Balance",
		},
		{
			Name:   "opening_date",
			Type:   def.FieldTypeDate,
			Label:  "Opening Date",
		},
	},

	Permissions: def.PermissionSet{
		Create: []string{"finance.bank_account.create"},
		Read:   []string{"finance.bank_account.read"},
		Write:  []string{"finance.bank_account.update"},
		Delete: []string{"finance.bank_account.delete"},
	},
}

func init() {
	def.Register(&BankAccountDefinition)
}

// BankTransactionDefinition — imported bank statement line.
// amount and transaction_date are immutable to preserve statement integrity.
var BankTransactionDefinition = def.SystemDefinition{
	Name:        "bank_transaction",
	Module:      "finance",
	Label:       "Bank Transaction",
	LabelPlural: "Bank Transactions",
	Description: "Imported bank statement line. Amount and transaction date are immutable to preserve statement integrity.",
	Icon:        "swap",
	Scope:       def.ScopeTenant,

	Fields: []def.FieldDef{
		{
			Name:       "bank_account",
			Type:       def.FieldTypeLink,
			Label:      "Bank Account",
			LinkTarget: "finance_bank_account",
			Required:   true,
			Immutable:  true,
		},
		{
			Name:      "transaction_date",
			Type:      def.FieldTypeDate,
			Label:     "Transaction Date",
			Required:  true,
			Immutable: true,
		},
		{
			Name:      "value_date",
			Type:      def.FieldTypeDate,
			Label:     "Value Date",
			Immutable: true,
		},
		{
			Name:        "amount",
			Type:        def.FieldTypeCurrency,
			Label:       "Amount",
			Description: "Positive = credit to account, negative = debit from account.",
			Required:    true,
			Immutable:   true,
		},
		{
			Name:       "description",
			Type:       def.FieldTypeData,
			Label:      "Description",
			Searchable: true,
			MaxLen:     500,
		},
		{
			Name:   "reference",
			Type:   def.FieldTypeData,
			Label:  "Reference",
			MaxLen: 255,
		},
		{
			Name:     "status",
			Type:     def.FieldTypeSelect,
			Label:    "Status",
			Required: true,
			Options:  []string{"unmatched", "matched", "reconciled", "excluded"},
			Default:  func() any { return "unmatched" },
		},
		{
			Name:       "journal_entry",
			Type:       def.FieldTypeLink,
			Label:      "Journal Entry",
			Description: "Journal entry created from this transaction during reconciliation.",
			LinkTarget: "finance_journal_entry",
		},
		{
			Name:   "external_id",
			Type:   def.FieldTypeData,
			Label:  "External ID",
			Description: "Bank-provided transaction identifier for deduplication.",
			Unique: true,
			MaxLen: 255,
		},
	},

	Permissions: def.PermissionSet{
		Create: []string{"finance.bank_transaction.create"},
		Read:   []string{"finance.bank_transaction.read"},
		Write:  []string{"finance.bank_transaction.update"},
		Actions: map[string][]string{
			"reconcile": {"finance.bank_transaction.reconcile"},
			"exclude":   {"finance.bank_transaction.reconcile"},
		},
	},

	Actions: []def.ActionDef{
		{
			Name:           "reconcile",
			Label:          "Reconcile",
			Description:    "Match this transaction to a journal entry and mark as reconciled.",
			ConfirmMessage: "Mark this transaction as reconciled?",
			Icon:           "check-circle",
			HandlerFunc:    stubAction("reconcile"),
		},
		{
			Name:           "exclude",
			Label:          "Exclude",
			Description:    "Exclude this transaction from reconciliation.",
			ConfirmMessage: "Exclude this transaction from reconciliation?",
			HandlerFunc:    stubAction("exclude"),
		},
	},
}

func init() {
	def.Register(&BankTransactionDefinition)
}
