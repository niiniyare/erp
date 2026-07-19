package finance

import (
	"context"

	"awo.so/awo/def"
	"awo.so/awo/filter"
	"awo.so/awo/runtime"
)

// InvoiceDefinition — customer invoice. Full lifecycle from draft through payment.
var InvoiceDefinition = def.SystemDefinition{
	Name:        "invoice",
	Module:      "finance",
	Label:       "Invoice",
	Description: "Customer invoice. Approve to create the GL journal entry. Track payments via Allocations.",
	Fields: []def.FieldDef{
		{
			Name:              "invoice_number",
			Type:              def.FieldTypeNamingSeries,
			Label:             "Invoice Number",
			Series:            "INV-{YYYY}-{SEQ:6}",
			Immutable:         true,
			TenantOverridable: true,
		},
		{
			Name:       "customer_id",
			Type:       def.FieldTypeData,
			Label:      "Customer ID",
			Required:   true,
			MaxLen:     36,
		},
		{
			Name:       "customer_name",
			Type:       def.FieldTypeData,
			Label:      "Customer Name",
			Required:   true,
			Searchable: true,
			MaxLen:     255,
		},
		{Name: "invoice_date", Type: def.FieldTypeDate, Label: "Invoice Date", Required: true},
		{Name: "due_date", Type: def.FieldTypeDate, Label: "Due Date"},
		{
			Name:    "status",
			Type:    def.FieldTypeSelect,
			Options: []string{"draft", "submitted", "approved", "sent", "partial", "paid", "overdue", "cancelled"},
			Default: func() any { return "draft" },
		},
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
		{Name: "subtotal", Type: def.FieldTypeCurrency, ReadOnly: true},
		{Name: "tax_amount", Type: def.FieldTypeCurrency, Label: "Tax Amount", ReadOnly: true},
		{Name: "total", Type: def.FieldTypeCurrency, ReadOnly: true},
		{
			Name:     "amount_paid",
			Type:     def.FieldTypeCurrency,
			Label:    "Amount Paid",
			ReadOnly: true,
			Default:  func() any { return "0.0000" },
		},
		{Name: "amount_outstanding", Type: def.FieldTypeCurrency, Label: "Amount Outstanding", ReadOnly: true},
		{Name: "payment_terms", Type: def.FieldTypeData, Label: "Payment Terms", MaxLen: 100},
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
		// organization_id is plain Data (not Link) — customers may belong to
		// external systems. Used only for org-scoped row-level policy.
		{Name: "organization_id", Type: def.FieldTypeData, Label: "Organization ID", MaxLen: 36},
	},
	Edges: []def.EdgeDef{
		{
			Name:          "lines",
			Target:        "finance_invoice_line",
			Type:          def.EdgeOneToMany,
			ForeignKey:    "invoice_id",
			CascadeDelete: true,
			Label:         "Invoice Lines",
			OrderBy:       "sort_order ASC, created_at ASC",
		},
	},
	Hooks: def.HookSet{
		BeforeCreate: []def.BeforeCreateHook{&InvoiceValidator{}},
		BeforeUpdate: []def.BeforeUpdateHook{&InvoiceStatusGuard{}},
	},
	Actions: []def.ActionDef{
		{
			Name:          "submit",
			Method:        def.ActionMethodPost,
			Label:         "Submit for Approval",
			Permission:    "role:finance.accountant",
			Icon:          "fa fa-paper-plane",
			WorkflowEvent: def.EventOnSubmit,
			HandlerFunc:   submitInvoice,
		},
		{
			Name:          "approve",
			Method:        def.ActionMethodPost,
			Label:         "Approve",
			Permission:    "role:finance.manager",
			Icon:          "fa fa-check",
			WorkflowEvent: def.EventOnApprove,
			HandlerFunc:   approveInvoice,
		},
		{
			Name:           "cancel",
			Method:         def.ActionMethodPost,
			Label:          "Cancel",
			Permission:     "role:finance.manager",
			Icon:           "fa fa-times",
			ConfirmMessage: "Cancel this invoice? This cannot be undone.",
			HandlerFunc:    cancelInvoice,
		},
	},
	WorkflowTriggers: []def.WorkflowTrigger{
		{
			On:         def.EventOnSubmit,
			WorkflowFn: "InvoiceApprovalWorkflow",
			TaskQueue:  "finance.invoice.approval",
			InputBuilder: func(r *def.EntityRecord, tc def.TriggerContext) (any, error) {
				return InvoiceApprovalInput{
					TenantID:  r.TenantID,
					InvoiceID: r.ID,
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
			"submit":  {"role:finance.accountant", "role:finance.manager", "role:tenant.admin"},
			"approve": {"role:finance.manager", "role:tenant.admin"},
			"cancel":  {"role:finance.manager", "role:tenant.admin"},
		},
		// Org-scoped: actors only see invoices belonging to their organization.
		Policy: invoiceOrgPolicy,
	},
}

// InvoiceLineDefinition — individual line item on an invoice.
var InvoiceLineDefinition = def.SystemDefinition{
	Name:        "invoice_line",
	Module:      "finance",
	Label:       "Invoice Line",
	Description: "Single product or service line on an invoice.",
	Fields: []def.FieldDef{
		{
			Name:       "invoice_id",
			Type:       def.FieldTypeLink,
			Label:      "Invoice",
			LinkTarget: "finance_invoice",
			Required:   true,
			Immutable:  true,
		},
		{
			Name:     "description",
			Type:     def.FieldTypeData,
			Required: true,
			MaxLen:   255,
		},
		{
			Name:    "quantity",
			Type:    def.FieldTypeCurrency,
			Required: true,
			Default: func() any { return "1.0000" },
		},
		{Name: "unit_price", Type: def.FieldTypeCurrency, Label: "Unit Price", Required: true},
		{
			Name:       "tax_id",
			Type:       def.FieldTypeLink,
			Label:      "Tax",
			LinkTarget: "finance_tax",
		},
		{Name: "tax_amount", Type: def.FieldTypeCurrency, Label: "Tax Amount", ReadOnly: true},
		{Name: "line_total", Type: def.FieldTypeCurrency, Label: "Line Total", ReadOnly: true},
		{
			Name:       "account_id",
			Type:       def.FieldTypeLink,
			Label:      "Revenue Account",
			LinkTarget: "finance_account",
		},
		{
			Name:       "cost_center_id",
			Type:       def.FieldTypeLink,
			Label:      "Cost Center",
			LinkTarget: "finance_cost_center",
		},
		{
			Name:    "sort_order",
			Type:    def.FieldTypeInt,
			Label:   "Sort Order",
			Default: func() any { return int64(0) },
		},
	},
	Permissions: def.PermissionSet{
		Create: []string{"role:tenant.admin", "role:finance.manager", "role:finance.accountant"},
		Read:   []string{"role:tenant.admin", "role:finance.manager", "role:finance.accountant", "role:finance.viewer"},
		Write:  []string{"role:tenant.admin", "role:finance.manager", "role:finance.accountant"},
		Delete: []string{"role:tenant.admin", "role:finance.manager"},
	},
}

// CreditNoteDefinition — credit note against an original invoice.
var CreditNoteDefinition = def.SystemDefinition{
	Name:        "credit_note",
	Module:      "finance",
	Label:       "Credit Note",
	Description: "Reduces amount owed on an original invoice. Approve to create the reversal journal entry.",
	Fields: []def.FieldDef{
		{
			Name:      "credit_number",
			Type:      def.FieldTypeNamingSeries,
			Label:     "Credit Note Number",
			Series:    "CN-{YYYY}-{SEQ:5}",
			Immutable: true,
		},
		{
			Name:       "original_invoice_id",
			Type:       def.FieldTypeLink,
			Label:      "Original Invoice",
			LinkTarget: "finance_invoice",
			Required:   true,
			Immutable:  true,
		},
		{Name: "reason", Type: def.FieldTypeSmallText, Required: true},
		{Name: "issue_date", Type: def.FieldTypeDate, Label: "Issue Date", Required: true},
		{
			Name:    "status",
			Type:    def.FieldTypeSelect,
			Options: []string{"draft", "approved", "applied", "cancelled"},
			Default: func() any { return "draft" },
		},
		{
			Name:       "currency_id",
			Type:       def.FieldTypeLink,
			Label:      "Currency",
			LinkTarget: "finance_currency",
			Required:   true,
		},
		{Name: "total_amount", Type: def.FieldTypeCurrency, Label: "Total Amount", Required: true},
		{
			Name:       "accounting_period_id",
			Type:       def.FieldTypeLink,
			Label:      "Accounting Period",
			LinkTarget: "finance_accounting_period",
			Required:   true,
		},
	},
	Actions: []def.ActionDef{
		{
			Name:        "approve",
			Method:      def.ActionMethodPost,
			Label:       "Approve Credit Note",
			Permission:  "role:finance.manager",
			Icon:        "fa fa-check",
			HandlerFunc: approveCreditNote,
		},
	},
	Permissions: def.PermissionSet{
		Create: []string{"role:tenant.admin", "role:finance.manager", "role:finance.accountant"},
		Read:   []string{"role:tenant.admin", "role:finance.manager", "role:finance.accountant", "role:finance.viewer"},
		Write:  []string{"role:tenant.admin", "role:finance.manager"},
		Delete: []string{"role:tenant.admin", "role:finance.manager"},
	},
}

// DebitNoteDefinition — debit note increasing amount owed.
var DebitNoteDefinition = def.SystemDefinition{
	Name:        "debit_note",
	Module:      "finance",
	Label:       "Debit Note",
	Description: "Increases amount owed. Typically issued by supplier to correct an under-billed invoice.",
	Fields: []def.FieldDef{
		{
			Name:      "debit_number",
			Type:      def.FieldTypeNamingSeries,
			Label:     "Debit Note Number",
			Series:    "DN-{YYYY}-{SEQ:5}",
			Immutable: true,
		},
		{
			Name:       "original_invoice_id",
			Type:       def.FieldTypeLink,
			Label:      "Original Invoice",
			LinkTarget: "finance_invoice",
			Immutable:  true,
		},
		{Name: "reason", Type: def.FieldTypeSmallText, Required: true},
		{Name: "issue_date", Type: def.FieldTypeDate, Label: "Issue Date", Required: true},
		{
			Name:    "status",
			Type:    def.FieldTypeSelect,
			Options: []string{"draft", "approved", "applied", "cancelled"},
			Default: func() any { return "draft" },
		},
		{
			Name:       "currency_id",
			Type:       def.FieldTypeLink,
			Label:      "Currency",
			LinkTarget: "finance_currency",
			Required:   true,
		},
		{Name: "total_amount", Type: def.FieldTypeCurrency, Label: "Total Amount", Required: true},
		{
			Name:       "accounting_period_id",
			Type:       def.FieldTypeLink,
			Label:      "Accounting Period",
			LinkTarget: "finance_accounting_period",
			Required:   true,
		},
	},
	Actions: []def.ActionDef{
		{
			Name:        "approve",
			Method:      def.ActionMethodPost,
			Label:       "Approve Debit Note",
			Permission:  "role:finance.manager",
			Icon:        "fa fa-check",
			HandlerFunc: approveDebitNote,
		},
	},
	Permissions: def.PermissionSet{
		Create: []string{"role:tenant.admin", "role:finance.manager", "role:finance.accountant"},
		Read:   []string{"role:tenant.admin", "role:finance.manager", "role:finance.accountant", "role:finance.viewer"},
		Write:  []string{"role:tenant.admin", "role:finance.manager"},
		Delete: []string{"role:tenant.admin", "role:finance.manager"},
	},
}

// ReceiptDefinition — payment receipt. Immutable after issuance.
var ReceiptDefinition = def.SystemDefinition{
	Name:        "receipt",
	Module:      "finance",
	Label:       "Receipt",
	Description: "Proof of payment. Created when a payment is processed. Immutable.",
	Fields: []def.FieldDef{
		{
			Name:      "receipt_number",
			Type:      def.FieldTypeNamingSeries,
			Label:     "Receipt Number",
			Series:    "RCT-{YYYY}-{SEQ:5}",
			Immutable: true,
		},
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
			Immutable:  true,
		},
		{Name: "receipt_date", Type: def.FieldTypeDate, Label: "Receipt Date", Required: true, Immutable: true},
		{Name: "amount", Type: def.FieldTypeCurrency, Required: true, Immutable: true},
		{
			Name:       "currency_id",
			Type:       def.FieldTypeLink,
			Label:      "Currency",
			LinkTarget: "finance_currency",
			Required:   true,
			Immutable:  true,
		},
		{Name: "notes", Type: def.FieldTypeSmallText},
	},
	Permissions: def.PermissionSet{
		Create: []string{"role:tenant.admin", "role:finance.manager", "role:finance.accountant"},
		Read:   []string{"role:tenant.admin", "role:finance.manager", "role:finance.accountant", "role:finance.viewer", "role:tenant.user"},
		Write:  []string{},  // receipts are immutable
		Delete: []string{},
	},
}

// invoiceOrgPolicy scopes invoice reads to the actor's assigned organization.
// Returns nil (no restriction) if no org context is available — tenant RLS remains sole gate.
func invoiceOrgPolicy(ctx context.Context) def.Filter {
	type orgKey struct{}
	orgID, _ := ctx.Value(orgKey{}).(string)
	if orgID == "" {
		return nil
	}
	return filter.Eq("organization_id", orgID)
}

func submitInvoice(ctx *def.ActionContext) (*def.ActionResult, error) {
	// TODO: validate line totals non-zero, set status="submitted",
	// trigger InvoiceApprovalWorkflow via Temporal
	return nil, &runtime.BusinessError{Code: "finance.not_implemented", Message: "submit not yet implemented", Status: 501}
}

func approveInvoice(ctx *def.ActionContext) (*def.ActionResult, error) {
	// TODO: set status="approved", create GL journal entry (debit AR, credit revenue per line),
	// apply taxes, send approval notification
	return nil, &runtime.BusinessError{Code: "finance.not_implemented", Message: "approve not yet implemented", Status: 501}
}

func cancelInvoice(ctx *def.ActionContext) (*def.ActionResult, error) {
	// TODO: verify no processed payments, reverse journal entry if exists, set status="cancelled"
	return nil, &runtime.BusinessError{Code: "finance.not_implemented", Message: "cancel not yet implemented", Status: 501}
}

func approveCreditNote(ctx *def.ActionContext) (*def.ActionResult, error) {
	// TODO: verify original invoice exists, set status="approved", create reversal journal entry
	return nil, &runtime.BusinessError{Code: "finance.not_implemented", Message: "approve credit note not yet implemented", Status: 501}
}

func approveDebitNote(ctx *def.ActionContext) (*def.ActionResult, error) {
	// TODO: set status="approved", create supplementary journal entry
	return nil, &runtime.BusinessError{Code: "finance.not_implemented", Message: "approve debit note not yet implemented", Status: 501}
}
