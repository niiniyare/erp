package finance

import (
	"context"

	"awo.so/awo/def"
	"awo.so/awo/runtime"
)

// ─── Invoice hooks ────────────────────────────────────────────────────────────

// InvoiceValidator validates a new invoice before creation.
type InvoiceValidator struct{}

func (h *InvoiceValidator) BeforeCreate(ctx context.Context, r *def.EntityRecord) error {
	ve := &runtime.ValidationError{Fields: map[string]string{}}

	name, _ := r.Data["customer_name"].(string)
	if name == "" {
		ve.Fields["customer_name"] = "Customer name is required"
	}
	if r.Data["customer_id"] == nil || r.Data["customer_id"] == "" {
		ve.Fields["customer_id"] = "Customer ID is required"
	}
	if r.Data["invoice_date"] == nil {
		ve.Fields["invoice_date"] = "Invoice date is required"
	}
	if r.Data["accounting_period_id"] == nil {
		ve.Fields["accounting_period_id"] = "Accounting period is required"
	}
	if r.Data["currency_id"] == nil {
		ve.Fields["currency_id"] = "Currency is required"
	}

	// Ensure status defaults to draft.
	if r.Data["status"] == nil {
		r.Data["status"] = "draft"
	}

	if len(ve.Fields) > 0 {
		return ve
	}
	return nil
}

// InvoiceStatusGuard prevents modification of terminal-state invoices.
type InvoiceStatusGuard struct{}

func (h *InvoiceStatusGuard) BeforeUpdate(ctx context.Context, r *def.EntityRecord, prev *def.EntityRecord) error {
	prevStatus, _ := prev.Data["status"].(string)
	switch prevStatus {
	case "cancelled":
		return &runtime.BusinessError{
			Code:    "finance.invoice.cancelled",
			Message: "Cancelled invoices cannot be modified.",
			Status:  409,
		}
	case "paid":
		return &runtime.BusinessError{
			Code:    "finance.invoice.paid",
			Message: "Fully paid invoices cannot be modified.",
			Status:  409,
		}
	}
	return nil
}

// ─── Payment hooks ────────────────────────────────────────────────────────────

// PaymentValidator validates a new payment before creation.
type PaymentValidator struct{}

func (h *PaymentValidator) BeforeCreate(ctx context.Context, r *def.EntityRecord) error {
	ve := &runtime.ValidationError{Fields: map[string]string{}}

	if r.Data["payment_date"] == nil {
		ve.Fields["payment_date"] = "Payment date is required"
	}
	if r.Data["payment_method_id"] == nil {
		ve.Fields["payment_method_id"] = "Payment method is required"
	}
	if r.Data["currency_id"] == nil {
		ve.Fields["currency_id"] = "Currency is required"
	}
	if r.Data["accounting_period_id"] == nil {
		ve.Fields["accounting_period_id"] = "Accounting period is required"
	}
	if r.Data["gl_account_id"] == nil {
		ve.Fields["gl_account_id"] = "GL account is required"
	}

	if toFloat64(r.Data["amount"]) <= 0 {
		ve.Fields["amount"] = "Amount must be greater than zero"
	}

	// Ensure status defaults to draft.
	if r.Data["status"] == nil {
		r.Data["status"] = "draft"
	}

	if len(ve.Fields) > 0 {
		return ve
	}
	return nil
}

// PaymentStatusGuard prevents modification of processed, cancelled, or reconciled payments.
type PaymentStatusGuard struct{}

func (h *PaymentStatusGuard) BeforeUpdate(ctx context.Context, r *def.EntityRecord, prev *def.EntityRecord) error {
	prevStatus, _ := prev.Data["status"].(string)
	switch prevStatus {
	case "processed":
		return &runtime.BusinessError{
			Code:    "finance.payment.processed",
			Message: "Processed payments cannot be modified. Cancel and reprocess.",
			Status:  409,
		}
	case "cancelled":
		return &runtime.BusinessError{
			Code:    "finance.payment.cancelled",
			Message: "Cancelled payments cannot be modified.",
			Status:  409,
		}
	case "reconciled":
		return &runtime.BusinessError{
			Code:    "finance.payment.reconciled",
			Message: "Reconciled payments cannot be modified.",
			Status:  409,
		}
	}
	return nil
}

// ─── Allocation hooks ─────────────────────────────────────────────────────────

// AllocationValidator validates a new payment allocation.
type AllocationValidator struct{}

func (h *AllocationValidator) BeforeCreate(ctx context.Context, r *def.EntityRecord) error {
	ve := &runtime.ValidationError{Fields: map[string]string{}}

	if r.Data["payment_id"] == nil {
		ve.Fields["payment_id"] = "Payment is required"
	}
	if r.Data["invoice_id"] == nil {
		ve.Fields["invoice_id"] = "Invoice is required"
	}
	if r.Data["currency_id"] == nil {
		ve.Fields["currency_id"] = "Currency is required"
	}
	if r.Data["allocation_date"] == nil {
		ve.Fields["allocation_date"] = "Allocation date is required"
	}
	if toFloat64(r.Data["amount_allocated"]) <= 0 {
		ve.Fields["amount_allocated"] = "Allocated amount must be greater than zero"
	}

	if len(ve.Fields) > 0 {
		return ve
	}
	return nil
}
