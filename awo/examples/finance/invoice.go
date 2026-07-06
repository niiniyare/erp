// Package finance demonstrates how a module author declares entities using the
// Awo framework def package. This is the canonical example referenced in
// framework documentation.
//
// This file would live in: internal/core/finance/definition.go
package finance

import (
	"context"
	"fmt"

	"awo.so/awo/def"
	"awo.so/awo/filter"
	"awo.so/awo/runtime"
)

// InvoiceDefinition is the canonical example of a SystemDefinition.
// One call to def.Register drives 5 subsystems simultaneously:
// persistence routing, API generation, SDUI, RBAC, and workflow triggering.
var InvoiceDefinition = def.SystemDefinition{
	Name:        "finance_invoice",
	Module:      "finance",
	Label:       "Invoice",
	LabelPlural: "Invoices",

	Fields: []def.FieldDef{
		{
			Name:   "number",
			Type:   def.FieldTypeNamingSeries,
			Label:  "Invoice #",
			Series: "INV-{YYYY}-{SEQ:5}",
			// TenantOverridable: true allows tenants to change the prefix
			// via the Settings module without a redeploy.
			TenantOverridable: true,
			ReadOnly:          true,
		},
		{
			Name:       "customer",
			Type:       def.FieldTypeLink,
			Label:      "Customer",
			LinkTarget: "crm_customer",
			Required:   true,
		},
		{
			Name:    "status",
			Type:    def.FieldTypeSelect,
			Label:   "Status",
			Options: []string{"Draft", "Submitted", "Paid", "Cancelled"},
			Default: func() any { return "Draft" },
		},
		{
			Name:     "total_kes",
			Type:     def.FieldTypeCurrency,
			Label:    "Total (KES)",
			Required: true,
		},
		{
			Name:  "notes",
			Type:  def.FieldTypeLongText,
			Label: "Notes",
		},
		{
			Name:   "due_date",
			Type:   def.FieldTypeDate,
			Label:  "Due Date",
		},
	},

	Edges: []def.EdgeDef{
		{
			Name:          "lines",
			Target:        "finance_invoice_line",
			Type:          def.EdgeOneToMany,
			Label:         "Line Items",
			CascadeDelete: true,
			OrderBy:       "created_at ASC",
		},
	},

	Hooks: def.HookSet{
		BeforeCreate: []def.BeforeCreateHook{&InvoiceValidator{}},
		AfterCreate:  []def.AfterCreateHook{&InvoiceAuditHook{}},
		BeforeDelete: []def.BeforeDeleteHook{&InvoiceDeleteGuard{}},
	},

	Permissions: def.PermissionSet{
		Create: []string{"role:finance.accounts_payable", "role:tenant.admin"},
		Read:   []string{"role:finance.viewer", "role:finance.accounts_payable", "role:tenant.admin"},
		Write:  []string{"role:finance.accounts_payable", "role:tenant.admin"},
		Delete: []string{"role:tenant.admin"},
		Policy: def.PolicyFunc(func(ctx context.Context) def.Filter {
			// Finance viewers only see invoices from the current month.
			// Admins and AP staff see all — return nil for them.
			// (In real code, check actor roles from session.ActorFromContext.)
			return filter.Eq("status", "Draft") // simplified example
		}),
	},

	Actions: []def.ActionDef{
		{
			Name:           "submit",
			Method:         def.ActionMethodPost,
			Label:          "Submit for Approval",
			Permission:     "role:finance.accounts_payable",
			ConfirmMessage: "Submit this invoice for approval?",
			HandlerFunc:    SubmitInvoiceAction,
			WorkflowEvent:  def.EventOnSubmit,
		},
		{
			Name:        "cancel",
			Method:      def.ActionMethodPost,
			Label:       "Cancel",
			Permission:  "role:tenant.admin",
			HandlerFunc: CancelInvoiceAction,
		},
	},

	WorkflowTriggers: []def.WorkflowTrigger{
		{
			On:         def.EventOnSubmit,
			WorkflowFn: "InvoiceApprovalWorkflow",
			TaskQueue:  "finance.invoice.submit",
			InputBuilder: func(rec *def.EntityRecord, tc def.TriggerContext) (any, error) {
				return InvoiceApprovalInput{
					TenantID:  tc.TenantID,
					InvoiceID: rec.ID,
					ActorID:   tc.Actor.UserID,
				}, nil
			},
		},
	},
}

func init() {
	def.Register(&InvoiceDefinition)
}

// --- Hook implementations ---

// InvoiceValidator enforces pre-create business rules.
type InvoiceValidator struct{}

func (v *InvoiceValidator) BeforeCreate(ctx context.Context, record *def.EntityRecord) error {
	// Business rule: cannot create a paid invoice directly.
	if record.GetString("status") == "Paid" {
		return &runtime.BusinessError{
			Code:    "finance_invoice.invalid_initial_status",
			Message: "New invoices must start in Draft status",
			Status:  422,
		}
	}
	return nil
}

// InvoiceAuditHook writes an audit entry after a successful create.
// Runs INSIDE the database transaction — error causes rollback.
type InvoiceAuditHook struct{}

func (h *InvoiceAuditHook) AfterCreate(ctx context.Context, record *def.EntityRecord) error {
	// TODO: write to audit_log via AuditService injected at construction time.
	// This is a placeholder — real implementation uses dependency injection.
	return nil
}

// InvoiceDeleteGuard prevents deletion of non-Draft invoices.
type InvoiceDeleteGuard struct{}

func (g *InvoiceDeleteGuard) BeforeDelete(ctx context.Context, record *def.EntityRecord) error {
	if record.GetString("status") != "Draft" {
		return &runtime.BusinessError{
			Code:    "finance_invoice.cannot_delete",
			Message: fmt.Sprintf("Cannot delete invoice in %q status", record.GetString("status")),
			Status:  409,
		}
	}
	return nil
}

// --- Action handlers ---

func SubmitInvoiceAction(ctx *def.ActionContext) (*def.ActionResult, error) {
	// TODO: validate invoice has at least one line item.
	// TODO: validate total_kes > 0.
	// Action context provides pre-scoped repository access.
	return &def.ActionResult{
		Message: "Invoice submitted for approval",
		// WorkflowID will be set by the runtime after the workflow starts.
	}, nil
}

func CancelInvoiceAction(ctx *def.ActionContext) (*def.ActionResult, error) {
	// TODO: implement cancellation logic.
	return &def.ActionResult{Message: "Invoice cancelled"}, nil
}

// --- Workflow input types ---

// InvoiceApprovalInput is the Temporal workflow input for invoice approval.
// Must be JSON-serialisable. Always include TenantID so the activity can
// construct its TenantContext without relying on Temporal headers.
type InvoiceApprovalInput struct {
	TenantID  interface{} `json:"tenant_id"`
	InvoiceID interface{} `json:"invoice_id"`
	ActorID   interface{} `json:"actor_id"`
}
