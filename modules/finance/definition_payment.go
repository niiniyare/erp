package finance

import "awo.so/awo/def"

// PaymentMethodDefinition — payment method master mapped to a GL account.
var PaymentMethodDefinition = def.SystemDefinition{
	Name:        "payment_method",
	Module:      "finance",
	Label:       "Payment Method",
	LabelPlural: "Payment Methods",
	Description: "Payment method master. Maps payment types (cash, bank transfer, card) to GL accounts.",
	Icon:        "credit-card",
	Scope:       def.ScopeTenant,

	Fields: []def.FieldDef{
		{
			Name:       "name",
			Type:       def.FieldTypeData,
			Label:      "Name",
			Required:   true,
			Searchable: true,
			MaxLen:     255,
		},
		{
			Name:     "method_type",
			Type:     def.FieldTypeSelect,
			Label:    "Method Type",
			Required: true,
			Options:  []string{"cash", "bank_transfer", "credit_card", "debit_card", "cheque", "digital_wallet", "other"},
		},
		{
			Name:       "gl_account",
			Type:       def.FieldTypeLink,
			Label:      "GL Account",
			Description: "GL account debited/credited when this payment method is used.",
			LinkTarget: "finance_account",
			Required:   true,
		},
		{
			Name:       "currency",
			Type:       def.FieldTypeLink,
			Label:      "Currency",
			LinkTarget: "finance_currency",
		},
		{
			Name:    "is_active",
			Type:    def.FieldTypeBool,
			Label:   "Active",
			Default: func() any { return true },
		},
	},

	Permissions: def.PermissionSet{
		Create: []string{"finance.payment_method.create"},
		Read:   []string{"finance.payment_method.read"},
		Write:  []string{"finance.payment_method.update"},
		Delete: []string{"finance.payment_method.delete"},
	},
}

func init() {
	def.Register(&PaymentMethodDefinition)
}

// PaymentDefinition — payment record with state machine.
// State: draft → submitted → processed → reconciled
var PaymentDefinition = def.SystemDefinition{
	Name:        "payment",
	Module:      "finance",
	Label:       "Payment",
	LabelPlural: "Payments",
	Description: "Payment record. Transitions: draft → submitted → processed → reconciled.",
	Icon:        "dollar-sign",
	Scope:       def.ScopeTenant,

	Fields: []def.FieldDef{
		{
			Name:      "payment_number",
			Type:      def.FieldTypeNamingSeries,
			Label:     "Payment Number",
			Series:    "PAY-{YYYY}-{SEQ:6}",
			Required:  true,
			Immutable: true,
			ReadOnly:  true,
		},
		{
			Name:     "payment_type",
			Type:     def.FieldTypeSelect,
			Label:    "Payment Type",
			Required: true,
			Immutable: true,
			Options:  []string{"inbound", "outbound"},
		},
		{
			Name:       "payment_method",
			Type:       def.FieldTypeLink,
			Label:      "Payment Method",
			LinkTarget: "finance_payment_method",
			Required:   true,
		},
		{
			Name:      "payment_date",
			Type:      def.FieldTypeDate,
			Label:     "Payment Date",
			Required:  true,
		},
		{
			Name:      "amount",
			Type:      def.FieldTypeCurrency,
			Label:     "Amount",
			Required:  true,
		},
		{
			Name:       "currency",
			Type:       def.FieldTypeLink,
			Label:      "Currency",
			LinkTarget: "finance_currency",
			Required:   true,
		},
		{
			Name:     "status",
			Type:     def.FieldTypeSelect,
			Label:    "Status",
			Required: true,
			Options:  []string{"draft", "submitted", "processed", "reconciled", "cancelled"},
			Default:  func() any { return "draft" },
			ReadOnly: true,
		},
		{
			Name:        "bank_account",
			Type:        def.FieldTypeLink,
			Label:       "Bank Account",
			Description: "Bank account used for this payment.",
			LinkTarget:  "finance_bank_account",
		},
		{
			Name:        "journal_entry",
			Type:        def.FieldTypeLink,
			Label:       "Journal Entry",
			Description: "Journal entry generated when payment is processed.",
			LinkTarget:  "finance_journal_entry",
			ReadOnly:    true,
		},
		{
			Name:   "reference",
			Type:   def.FieldTypeData,
			Label:  "Reference",
			MaxLen: 255,
		},
		{
			Name:  "memo",
			Type:  def.FieldTypeSmallText,
			Label: "Memo",
		},
	},

	Actions: []def.ActionDef{
		{
			Name:           "submit",
			Label:          "Submit",
			Description:    "Submit for approval. Transitions draft → submitted.",
			ConfirmMessage: "Submit this payment for approval?",
			Icon:           "send",
			WorkflowEvent:  def.EventType("submit"),
			HandlerFunc:    transitionStatus("finance_payment", "submit", "draft", "submitted"),
		},
		{
			Name:           "process",
			Label:          "Process",
			Description:    "Process payment and generate journal entry. Transitions submitted → processed.",
			ConfirmMessage: "Process this payment? A journal entry will be created.",
			Icon:           "check",
			WorkflowEvent:  def.EventType("process"),
			HandlerFunc:    transitionStatus("finance_payment", "process", "submitted", "processed"),
		},
		{
			Name:           "reconcile",
			Label:          "Reconcile",
			Description:    "Mark as reconciled after bank statement matching.",
			ConfirmMessage: "Mark this payment as reconciled?",
			Icon:           "check-circle",
			HandlerFunc:    transitionStatus("finance_payment", "reconcile", "processed", "reconciled"),
		},
		{
			Name:           "cancel",
			Label:          "Cancel",
			Description:    "Cancel a draft or submitted payment.",
			ConfirmMessage: "Cancel this payment? This action cannot be undone.",
			Icon:           "x-circle",
			HandlerFunc:    cancelAction("finance_payment", "processed", "reconciled"),
		},
	},

	Permissions: def.PermissionSet{
		Create: []string{"finance.payment.create"},
		Read:   []string{"finance.payment.read"},
		Write:  []string{"finance.payment.update"},
		Delete: []string{"finance.payment.delete"},
		Actions: map[string][]string{
			"submit":    {"finance.payment.submit"},
			"process":   {"finance.payment.process"},
			"reconcile": {"finance.payment.reconcile"},
			"cancel":    {"finance.payment.cancel"},
		},
	},
}

func init() {
	def.Register(&PaymentDefinition)
}
