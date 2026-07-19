package finance

import (
	"context"
	"fmt"

	"awo.so/awo/def"
	"awo.so/awo/runtime"
)

// ─── JournalEntry hooks ───────────────────────────────────────────────────────

// JournalEntryValidator validates a new journal entry before creation.
// Line balance is NOT validated here — lines are added after entry creation.
// Balance validation happens in the "post" action via PostingService.
type JournalEntryValidator struct{}

func (h *JournalEntryValidator) BeforeCreate(ctx context.Context, r *def.EntityRecord) error {
	ve := &runtime.ValidationError{Fields: map[string]string{}}

	if r.Data["posting_date"] == nil {
		ve.Fields["posting_date"] = "Posting date is required"
	}
	if r.Data["journal_id"] == nil {
		ve.Fields["journal_id"] = "Journal is required"
	}
	if r.Data["accounting_period_id"] == nil {
		ve.Fields["accounting_period_id"] = "Accounting period is required"
	}
	if r.Data["currency_id"] == nil {
		ve.Fields["currency_id"] = "Currency is required"
	}

	// Ensure status defaults to draft on create.
	if r.Data["status"] == nil {
		r.Data["status"] = "draft"
	}

	if len(ve.Fields) > 0 {
		return ve
	}
	return nil
}

// PostingGuard prevents modifications to posted, cancelled, or reversed journal entries.
type PostingGuard struct{}

func (h *PostingGuard) BeforeUpdate(ctx context.Context, r *def.EntityRecord, prev *def.EntityRecord) error {
	status, _ := prev.Data["status"].(string)
	switch status {
	case "posted":
		return &runtime.BusinessError{
			Code:    "finance.journal_entry.immutable_posted",
			Message: "Posted journal entries cannot be modified. Create a reversal entry instead.",
			Status:  409,
		}
	case "cancelled":
		return &runtime.BusinessError{
			Code:    "finance.journal_entry.immutable_cancelled",
			Message: "Cancelled journal entries cannot be modified.",
			Status:  409,
		}
	case "reversed":
		return &runtime.BusinessError{
			Code:    "finance.journal_entry.immutable_reversed",
			Message: "Reversed journal entries cannot be modified.",
			Status:  409,
		}
	}
	return nil
}

// ─── JournalEntryLine hooks ───────────────────────────────────────────────────

// LineBalanceValidator ensures a journal entry line has at most one of debit/credit non-zero.
type LineBalanceValidator struct{}

func (h *LineBalanceValidator) BeforeCreate(ctx context.Context, r *def.EntityRecord) error {
	debitF := toFloat64(r.Data["debit_amount"])
	creditF := toFloat64(r.Data["credit_amount"])

	if debitF < 0 || creditF < 0 {
		return &runtime.ValidationError{Fields: map[string]string{
			"debit_amount": "Amounts must be non-negative",
		}}
	}
	if debitF > 0 && creditF > 0 {
		return &runtime.ValidationError{Fields: map[string]string{
			"debit_amount":  "A line cannot have both debit and credit amounts",
			"credit_amount": "A line cannot have both debit and credit amounts",
		}}
	}
	return nil
}

// ─── FiscalYear hooks ─────────────────────────────────────────────────────────

// FiscalYearTransitionGuard enforces valid fiscal year status transitions.
// open → closed: allowed
// closed → locked: allowed
// Any other change: rejected
type FiscalYearTransitionGuard struct{}

func (h *FiscalYearTransitionGuard) BeforeUpdate(ctx context.Context, r *def.EntityRecord, prev *def.EntityRecord) error {
	prevStatus, _ := prev.Data["status"].(string)
	newStatus, _ := r.Data["status"].(string)

	if newStatus == "" || prevStatus == newStatus {
		return nil
	}

	allowed := map[string]string{
		"open":   "closed",
		"closed": "locked",
	}
	if allowed[prevStatus] != newStatus {
		return &runtime.BusinessError{
			Code:    "finance.fiscal_year.invalid_transition",
			Message: fmt.Sprintf("Cannot transition fiscal year from %q to %q", prevStatus, newStatus),
			Status:  422,
		}
	}
	return nil
}

// ─── AccountingPeriod hooks ───────────────────────────────────────────────────

// PeriodTransitionGuard enforces valid accounting period status transitions.
// open ↔ closed: allowed; locked: terminal.
type PeriodTransitionGuard struct{}

func (h *PeriodTransitionGuard) BeforeUpdate(ctx context.Context, r *def.EntityRecord, prev *def.EntityRecord) error {
	prevStatus, _ := prev.Data["status"].(string)
	newStatus, _ := r.Data["status"].(string)

	if newStatus == "" || prevStatus == newStatus {
		return nil
	}
	if prevStatus == "locked" {
		return &runtime.BusinessError{
			Code:    "finance.accounting_period.locked",
			Message: "Locked accounting periods cannot be modified.",
			Status:  409,
		}
	}
	return nil
}

// ─── Shared helpers ───────────────────────────────────────────────────────────

// toFloat64 coerces common numeric representations to float64 for comparison.
// FieldTypeCurrency values arrive as decimal strings from the framework.
func toFloat64(v any) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int64:
		return float64(val)
	case int:
		return float64(val)
	case string:
		var f float64
		fmt.Sscanf(val, "%f", &f)
		return f
	}
	return 0
}
