package finance

import (
	"awo.so/awo/def"
	"awo.so/awo/runtime"
)

// PaymentMethodDefinition — payment method master (cash, bank transfer, M-Pesa, etc.).
var PaymentMethodDefinition = def.SystemDefinition{
	Name:        "payment_method",
	Module:      "finance",
	Label:       "Payment Method",
	Description: "Defines how payments are made. Each method maps to a GL account for posting.",
	Fields: []def.FieldDef{
		{
			Name:       "name",
			Type:       def.FieldTypeData,
			Required:   true,
			Searchable: true,
			MaxLen:     100,
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
			Name:     "payment_type",
			Type:     def.FieldTypeSelect,
			Label:    "Payment Type",
			Options:  []string{"cash", "bank_transfer", "cheque", "mobile_money", "card", "other"},
			Required: true,
		},
		{
			Name:       "bank_account_id",
			Type:       def.FieldTypeLink,
			Label:      "Bank Account",
			LinkTarget: "finance_bank_account",
		},
		{
			Name:       "gl_account_id",
			Type:       def.FieldTypeLink,
			Label:      "GL Account",
			LinkTarget: "finance_account",
		},
		{
			Name:    "active",
			Type:    def.FieldTypeBool,
			Default: func() any { return true },
		},
	},
	Permissions: def.PermissionSet{
		Create: []string{"role:tenant.admin", "role:finance.manager"},
		Read:   []string{"role:tenant.admin", "role:finance.manager", "role:finance.accountant", "role:finance.viewer", "role:tenant.user"},
		Write:  []string{"role:tenant.admin", "role:finance.manager"},
		Delete: []string{"role:tenant.admin"},
	},
}

// PaymentDefinition — payment record. MANDATORY system entity.
//
// Covers both incoming (receive) and outgoing (send) payments.
// Lifecycle: draft → submitted → processed → reconciled (or cancelled).
var PaymentDefinition = def.SystemDefinition{
	Name:        "payment",
	Module:      "finance",
	Label:       "Payment",
	Description: "Money movement record. Process to create the GL journal entry.",
	Fields: []def.FieldDef{
		{
			Name:              "payment_number",
			Type:              def.FieldTypeNamingSeries,
			Label:             "Payment Number",
			Series:            "PAY-{YYYY}-{SEQ:6}",
			Immutable:         true,
			TenantOverridable: true,
		},
		{
			Name:     "payment_type",
			Type:     def.FieldTypeSelect,
			Label:    "Payment Direction",
			Options:  []string{"receive", "send", "internal"},
			Required: true,
		},
		{Name: "payment_date", Type: def.FieldTypeDate, Label: "Payment Date", Required: true},
		{
			Name:       "payment_method_id",
			Type:       def.FieldTypeLink,
			Label:      "Payment Method",
			LinkTarget: "finance_payment_method",
			Required:   true,
		},
		{Name: "party_type", Type: def.FieldTypeSelect, Label: "Party Type", Options: []string{"customer", "supplier", "employee", "other"}},
		{Name: "party_id", Type: def.FieldTypeData, Label: "Party ID", MaxLen: 36},
		{Name: "party_name", Type: def.FieldTypeData, Label: "Party Name", MaxLen: 255},
		{Name: "amount", Type: def.FieldTypeCurrency, Required: true},
		{
			Name:       "currency_id",
			Type:       def.FieldTypeLink,
			Label:      "Currency",
			LinkTarget: "finance_currency",
			Required:   true,
		},
		{
			Name:    "exchange_rate",
			Type:    def.FieldTypeCurrency,
			Label:   "Exchange Rate",
			Default: func() any { return "1.0000" },
		},
		{Name: "reference", Type: def.FieldTypeData, Label: "External Reference", MaxLen: 100},
		{
			Name:    "status",
			Type:    def.FieldTypeSelect,
			Options: []string{"draft", "submitted", "processed", "cancelled", "reconciled"},
			Default: func() any { return "draft" },
		},
		{
			Name:       "bank_account_id",
			Type:       def.FieldTypeLink,
			Label:      "Bank Account",
			LinkTarget: "finance_bank_account",
		},
		{
			Name:       "gl_account_id",
			Type:       def.FieldTypeLink,
			Label:      "GL Account",
			LinkTarget: "finance_account",
			Required:   true,
		},
		{Name: "notes", Type: def.FieldTypeSmallText},
		{
			Name:       "accounting_period_id",
			Type:       def.FieldTypeLink,
			Label:      "Accounting Period",
			LinkTarget: "finance_accounting_period",
			Required:   true,
		},
		{
			Name:       "journal_entry_id",
			Type:       def.FieldTypeLink,
			Label:      "Journal Entry",
			LinkTarget: "finance_journal_entry",
			ReadOnly:   true,
		},
	},
	Hooks: def.HookSet{
		BeforeCreate: []def.BeforeCreateHook{&PaymentValidator{}},
		BeforeUpdate: []def.BeforeUpdateHook{&PaymentStatusGuard{}},
	},
	Actions: []def.ActionDef{
		{
			Name:        "process",
			Method:      def.ActionMethodPost,
			Label:       "Process Payment",
			Description: "Create GL journal entry and mark payment as processed.",
			Permission:  "role:finance.accountant",
			Icon:        "fa fa-check",
			HandlerFunc: processPayment,
		},
		{
			Name:           "cancel",
			Method:         def.ActionMethodPost,
			Label:          "Cancel Payment",
			Permission:     "role:finance.manager",
			Icon:           "fa fa-times",
			ConfirmMessage: "Cancel this payment?",
			HandlerFunc:    cancelPayment,
		},
	},
	WorkflowTriggers: []def.WorkflowTrigger{
		{
			On:         def.EventOnSubmit,
			WorkflowFn: "PaymentProcessingWorkflow",
			TaskQueue:  "finance.payment.process",
			InputBuilder: func(r *def.EntityRecord, tc def.TriggerContext) (any, error) {
				return PaymentProcessingInput{
					TenantID:  r.TenantID,
					PaymentID: r.ID,
					ActorID:   tc.Actor.UserID,
				}, nil
			},
		},
	},
	Permissions: def.PermissionSet{
		Create: []string{"role:tenant.admin", "role:finance.manager", "role:finance.accountant"},
		Read:   []string{"role:tenant.admin", "role:finance.manager", "role:finance.accountant", "role:finance.viewer"},
		Write:  []string{"role:tenant.admin", "role:finance.manager", "role:finance.accountant"},
		Delete: []string{"role:tenant.admin", "role:finance.manager"},
		Actions: map[string][]string{
			"process": {"role:finance.accountant", "role:finance.manager", "role:tenant.admin"},
			"cancel":  {"role:finance.manager", "role:tenant.admin"},
		},
	},
}

// AllocationDefinition — maps a payment to one or more invoices.
var AllocationDefinition = def.SystemDefinition{
	Name:        "allocation",
	Module:      "finance",
	Label:       "Payment Allocation",
	Description: "Links a payment to an invoice. Multiple allocations allowed per payment.",
	Fields: []def.FieldDef{
		{
			Name:       "payment_id",
			Type:       def.FieldTypeLink,
			Label:      "Payment",
			LinkTarget: "finance_payment",
			Required:   true,
			Immutable:  true,
		},
		{
			Name:       "invoice_id",
			Type:       def.FieldTypeLink,
			Label:      "Invoice",
			LinkTarget: "finance_invoice",
			Required:   true,
			Immutable:  true,
		},
		{Name: "amount_allocated", Type: def.FieldTypeCurrency, Label: "Amount Allocated", Required: true},
		{
			Name:       "currency_id",
			Type:       def.FieldTypeLink,
			Label:      "Currency",
			LinkTarget: "finance_currency",
			Required:   true,
			Immutable:  true,
		},
		{Name: "allocation_date", Type: def.FieldTypeDate, Label: "Allocation Date", Required: true},
	},
	Hooks: def.HookSet{
		BeforeCreate: []def.BeforeCreateHook{&AllocationValidator{}},
	},
	Permissions: def.PermissionSet{
		Create: []string{"role:tenant.admin", "role:finance.manager", "role:finance.accountant"},
		Read:   []string{"role:tenant.admin", "role:finance.manager", "role:finance.accountant", "role:finance.viewer"},
		Write:  []string{"role:tenant.admin", "role:finance.manager"},
		Delete: []string{"role:tenant.admin", "role:finance.manager"},
	},
}

func processPayment(ctx *def.ActionContext) (*def.ActionResult, error) {
	// TODO: load payment, verify status=="submitted", verify period open,
	// determine GL accounts from payment method, create JournalEntry+lines,
	// post the journal entry, set payment status="processed"
	return nil, &runtime.BusinessError{
		Code:    "finance.not_implemented",
		Message: "process payment requires PostingService — wire EntityRepository first",
		Status:  501,
	}
}

func cancelPayment(ctx *def.ActionContext) (*def.ActionResult, error) {
	// TODO: verify status in ("submitted","processed"), reverse journal entry if exists,
	// set status="cancelled"
	return nil, &runtime.BusinessError{
		Code:    "finance.not_implemented",
		Message: "cancel payment not yet implemented",
		Status:  501,
	}
}
