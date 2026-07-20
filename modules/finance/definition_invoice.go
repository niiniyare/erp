package finance

import (
	"context"
	"fmt"

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
	rt := ctx.Runtime
	invoiceRepo := rt.Repo("finance_invoice")

	invoice, err := invoiceRepo.Get(ctx.Ctx, ctx.RecordID)
	if err != nil {
		return nil, fmt.Errorf("finance.submit_invoice: %w", err)
	}
	if s := invoice.GetString("status"); s != "draft" {
		return nil, &runtime.BusinessError{
			Code:    "finance.invoice.invalid_status",
			Message: "only draft invoices can be submitted; status: " + s,
			Status:  400,
		}
	}
	if invoice.GetDecimal("total").IsZero() {
		return nil, &runtime.BusinessError{
			Code:    "finance.invoice.zero_total",
			Message: "cannot submit an invoice with zero total",
			Status:  400,
		}
	}

	if _, err := invoiceRepo.Update(ctx.Ctx, ctx.RecordID, map[string]any{"status": "submitted"}); err != nil {
		return nil, fmt.Errorf("finance.submit_invoice: update status: %w", err)
	}

	// Trigger approval workflow.
	wfID, _ := rt.StartWorkflow(ctx.Ctx, def.ActionWorkflowSpec{
		WorkflowFn: "InvoiceApprovalWorkflow",
		TaskQueue:  "finance.invoice.approval",
		Input: InvoiceApprovalInput{
			TenantID:  rt.TenantID(),
			InvoiceID: ctx.RecordID,
			ActorID:   rt.Actor().UserID,
		},
	})

	_ = rt.Publish(ctx.Ctx, def.ActionEvent{
		Topic:   "finance.invoice.submitted",
		Payload: map[string]any{"invoice_id": ctx.RecordID},
	})

	return &def.ActionResult{
		Message:    "Invoice submitted for approval.",
		Data:       map[string]any{"invoice_id": ctx.RecordID, "status": "submitted"},
		WorkflowID: wfID,
	}, nil
}

func approveInvoice(ctx *def.ActionContext) (*def.ActionResult, error) {
	rt := ctx.Runtime
	invoiceRepo := rt.Repo("finance_invoice")

	invoice, err := invoiceRepo.Get(ctx.Ctx, ctx.RecordID)
	if err != nil {
		return nil, fmt.Errorf("finance.approve_invoice: %w", err)
	}
	if s := invoice.GetString("status"); s != "submitted" {
		return nil, &runtime.BusinessError{
			Code:    "finance.invoice.invalid_status",
			Message: "only submitted invoices can be approved; status: " + s,
			Status:  400,
		}
	}

	if err := rt.Tx(ctx.Ctx, func(txCtx context.Context) error {
		_, err := invoiceRepo.Update(txCtx, ctx.RecordID, map[string]any{"status": "approved"})
		return err
	}); err != nil {
		return nil, fmt.Errorf("finance.approve_invoice: %w", err)
	}

	_ = rt.Notify(ctx.Ctx, def.ActionNotification{
		Subject: "Invoice Approved",
		Body:    "Invoice " + invoice.GetString("invoice_number") + " has been approved.",
		Channel: "in_app",
	})

	_ = rt.Publish(ctx.Ctx, def.ActionEvent{
		Topic:   "finance.invoice.approved",
		Payload: map[string]any{"invoice_id": ctx.RecordID},
	})

	return &def.ActionResult{
		Message: "Invoice approved.",
		Data:    map[string]any{"invoice_id": ctx.RecordID, "status": "approved"},
	}, nil
}

func cancelInvoice(ctx *def.ActionContext) (*def.ActionResult, error) {
	rt := ctx.Runtime
	invoiceRepo := rt.Repo("finance_invoice")
	paymentRepo := rt.Repo("finance_payment")

	invoice, err := invoiceRepo.Get(ctx.Ctx, ctx.RecordID)
	if err != nil {
		return nil, fmt.Errorf("finance.cancel_invoice: %w", err)
	}

	s := invoice.GetString("status")
	if s == "cancelled" || s == "paid" {
		return nil, &runtime.BusinessError{
			Code:    "finance.invoice.cannot_cancel",
			Message: "cannot cancel an invoice with status: " + s,
			Status:  400,
		}
	}

	// Verify no processed payments exist.
	count, err := paymentRepo.Count(ctx.Ctx, filter.And(
		filter.Eq("status", "processed"),
	))
	if err != nil {
		return nil, fmt.Errorf("finance.cancel_invoice: check payments: %w", err)
	}
	if count > 0 {
		return nil, &runtime.BusinessError{
			Code:    "finance.invoice.has_processed_payments",
			Message: "cannot cancel an invoice with processed payments",
			Status:  400,
		}
	}

	if _, err := invoiceRepo.Update(ctx.Ctx, ctx.RecordID, map[string]any{"status": "cancelled"}); err != nil {
		return nil, fmt.Errorf("finance.cancel_invoice: update: %w", err)
	}

	_ = rt.Publish(ctx.Ctx, def.ActionEvent{
		Topic:   "finance.invoice.cancelled",
		Payload: map[string]any{"invoice_id": ctx.RecordID},
	})

	return &def.ActionResult{
		Message: "Invoice cancelled.",
		Data:    map[string]any{"invoice_id": ctx.RecordID, "status": "cancelled"},
	}, nil
}

func approveCreditNote(ctx *def.ActionContext) (*def.ActionResult, error) {
	rt := ctx.Runtime
	creditRepo := rt.Repo("finance_credit_note")

	note, err := creditRepo.Get(ctx.Ctx, ctx.RecordID)
	if err != nil {
		return nil, fmt.Errorf("finance.approve_credit_note: %w", err)
	}
	if s := note.GetString("status"); s != "draft" {
		return nil, &runtime.BusinessError{
			Code:    "finance.credit_note.invalid_status",
			Message: "only draft credit notes can be approved; status: " + s,
			Status:  400,
		}
	}

	if _, err := creditRepo.Update(ctx.Ctx, ctx.RecordID, map[string]any{"status": "approved"}); err != nil {
		return nil, fmt.Errorf("finance.approve_credit_note: %w", err)
	}

	_ = rt.Publish(ctx.Ctx, def.ActionEvent{
		Topic:   "finance.credit_note.approved",
		Payload: map[string]any{"credit_note_id": ctx.RecordID},
	})

	return &def.ActionResult{
		Message: "Credit note approved.",
		Data:    map[string]any{"credit_note_id": ctx.RecordID, "status": "approved"},
	}, nil
}

func approveDebitNote(ctx *def.ActionContext) (*def.ActionResult, error) {
	rt := ctx.Runtime
	debitRepo := rt.Repo("finance_debit_note")

	note, err := debitRepo.Get(ctx.Ctx, ctx.RecordID)
	if err != nil {
		return nil, fmt.Errorf("finance.approve_debit_note: %w", err)
	}
	if s := note.GetString("status"); s != "draft" {
		return nil, &runtime.BusinessError{
			Code:    "finance.debit_note.invalid_status",
			Message: "only draft debit notes can be approved; status: " + s,
			Status:  400,
		}
	}

	if _, err := debitRepo.Update(ctx.Ctx, ctx.RecordID, map[string]any{"status": "approved"}); err != nil {
		return nil, fmt.Errorf("finance.approve_debit_note: %w", err)
	}

	_ = rt.Publish(ctx.Ctx, def.ActionEvent{
		Topic:   "finance.debit_note.approved",
		Payload: map[string]any{"debit_note_id": ctx.RecordID},
	})

	return &def.ActionResult{
		Message: "Debit note approved.",
		Data:    map[string]any{"debit_note_id": ctx.RecordID, "status": "approved"},
	}, nil
}
