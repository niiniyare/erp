package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// BudgetStatus tracks the lifecycle of a budget.
type BudgetStatus string

const (
	BudgetStatusDraft     BudgetStatus = "DRAFT"     // Being prepared; not yet in effect
	BudgetStatusSubmitted BudgetStatus = "SUBMITTED" // Awaiting approval
	BudgetStatusApproved  BudgetStatus = "APPROVED"  // Active budget; enforced
	BudgetStatusRejected  BudgetStatus = "REJECTED"  // Returned for revision
	BudgetStatusRevised   BudgetStatus = "REVISED"   // Supplementary/revised version of an approved budget
	BudgetStatusClosed    BudgetStatus = "CLOSED"    // Period ended; no further changes
)

// IsValid returns true if the BudgetStatus is a recognised value.
func (bs BudgetStatus) IsValid() bool {
	switch bs {
	case BudgetStatusDraft, BudgetStatusSubmitted, BudgetStatusApproved,
		BudgetStatusRejected, BudgetStatusRevised, BudgetStatusClosed:
		return true
	default:
		return false
	}
}

// String returns the string representation of BudgetStatus.
func (bs BudgetStatus) String() string { return string(bs) }

// IsEditable returns true when new line-items may be added or amounts changed.
func (bs BudgetStatus) IsEditable() bool {
	return bs == BudgetStatusDraft || bs == BudgetStatusRejected
}

// BudgetType classifies the nature of a budget.
type BudgetType string

const (
	BudgetTypeAnnual      BudgetType = "ANNUAL"      // Full-year operational budget
	BudgetTypeQuarterly   BudgetType = "QUARTERLY"   // Quarterly rolling budget
	BudgetTypeProject     BudgetType = "PROJECT"      // Project / capex budget
	BudgetTypeDepartment  BudgetType = "DEPARTMENT"  // Department-level cost budget
	BudgetTypeCapex       BudgetType = "CAPEX"        // Capital expenditure budget
)

// IsValid returns true if the BudgetType is a recognised value.
func (bt BudgetType) IsValid() bool {
	switch bt {
	case BudgetTypeAnnual, BudgetTypeQuarterly, BudgetTypeProject,
		BudgetTypeDepartment, BudgetTypeCapex:
		return true
	default:
		return false
	}
}

// String returns the string representation of BudgetType.
func (bt BudgetType) String() string { return string(bt) }

// Budget is the header record for a budget.  It groups BudgetLineItems by fiscal year
// and cost centre, and tracks approval status.
type Budget struct {
	ID           uuid.UUID `json:"id"`
	TenantID     uuid.UUID `json:"tenant_id"`
	FiscalYearID uuid.UUID `json:"fiscal_year_id"`

	Name         string     `json:"name"`
	Description  *string    `json:"description,omitempty"`
	BudgetType   BudgetType `json:"budget_type"`
	Status       BudgetStatus `json:"status"`
	CurrencyCode string     `json:"currency_code"` // ISO 4217

	// Scope — budget may cover all cost centres or just one
	CostCenterID *uuid.UUID `json:"cost_center_id,omitempty"`

	// Approval metadata
	SubmittedAt *time.Time `json:"submitted_at,omitempty"`
	SubmittedBy *uuid.UUID `json:"submitted_by,omitempty"`
	ApprovedAt  *time.Time `json:"approved_at,omitempty"`
	ApprovedBy  *uuid.UUID `json:"approved_by,omitempty"`
	RejectedAt  *time.Time `json:"rejected_at,omitempty"`
	RejectedBy  *uuid.UUID `json:"rejected_by,omitempty"`
	RejectNote  *string    `json:"reject_note,omitempty"`

	// Revision tracking
	Version        int        `json:"version"`                    // 1 = original, 2+ = revisions
	OriginalBudgetID *uuid.UUID `json:"original_budget_id,omitempty"` // Points to v1 for revised budgets

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	CreatedBy uuid.UUID  `json:"created_by"`
	UpdatedBy *uuid.UUID `json:"updated_by,omitempty"`
}

// Validate returns ValidationErrors for the Budget header.
func (b *Budget) Validate() []ValidationError {
	var errs []ValidationError

	if b.TenantID == uuid.Nil {
		errs = append(errs, ValidationError{Field: "tenant_id", Message: "tenant_id is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	if b.FiscalYearID == uuid.Nil {
		errs = append(errs, ValidationError{Field: "fiscal_year_id", Message: "fiscal_year_id is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	if strings.TrimSpace(b.Name) == "" {
		errs = append(errs, ValidationError{Field: "name", Message: "budget name is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	if !b.BudgetType.IsValid() {
		errs = append(errs, ValidationError{Field: "budget_type", Message: fmt.Sprintf("invalid budget type: %s", b.BudgetType), Code: "INVALID_VALUE", Severity: ValidationSeverityError})
	}
	if !b.Status.IsValid() {
		errs = append(errs, ValidationError{Field: "status", Message: fmt.Sprintf("invalid budget status: %s", b.Status), Code: "INVALID_VALUE", Severity: ValidationSeverityError})
	}
	if len(b.CurrencyCode) != 3 {
		errs = append(errs, ValidationError{Field: "currency_code", Message: "currency_code must be a 3-character ISO 4217 code", Code: "INVALID_FORMAT", Severity: ValidationSeverityError})
	}
	if b.Version < 1 {
		errs = append(errs, ValidationError{Field: "version", Message: "version must be ≥ 1", Code: "VALUE_OUT_OF_RANGE", Severity: ValidationSeverityError})
	}

	return errs
}

// BudgetLineItem is a single account allocation within a Budget, optionally broken down
// by accounting period.
type BudgetLineItem struct {
	ID       uuid.UUID `json:"id"`
	BudgetID uuid.UUID `json:"budget_id"`
	TenantID uuid.UUID `json:"tenant_id"`

	AccountID    uuid.UUID  `json:"account_id"`
	CostCenterID *uuid.UUID `json:"cost_center_id,omitempty"`
	PeriodID     *uuid.UUID `json:"period_id,omitempty"` // nil = annual lump sum

	BudgetedAmount decimal.Decimal `json:"budgeted_amount"`
	ActualAmount   decimal.Decimal `json:"actual_amount"`   // Populated by GL queries
	VarianceAmount decimal.Decimal `json:"variance_amount"` // Computed: budgeted - actual
	VariancePct    decimal.Decimal `json:"variance_pct"`    // Computed: variance / budgeted * 100

	Notes *string `json:"notes,omitempty"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	CreatedBy uuid.UUID  `json:"created_by"`
	UpdatedBy *uuid.UUID `json:"updated_by,omitempty"`
}

// IsOverBudget returns true when actual spend exceeds the budgeted amount.
func (li *BudgetLineItem) IsOverBudget() bool {
	return li.ActualAmount.GreaterThan(li.BudgetedAmount)
}

// VariancePercent recalculates the variance percentage from the stored amounts.
// Returns zero when BudgetedAmount is zero to avoid division by zero.
func (li *BudgetLineItem) VariancePercent() decimal.Decimal {
	if li.BudgetedAmount.IsZero() {
		return decimal.Zero
	}
	return li.BudgetedAmount.Sub(li.ActualAmount).Div(li.BudgetedAmount).Mul(decimal.NewFromInt(100))
}

// ExceedsVarianceThreshold returns true when the absolute variance percentage exceeds
// the supplied threshold (e.g. from Account.BudgetVarianceThreshold).
func (li *BudgetLineItem) ExceedsVarianceThreshold(threshold decimal.Decimal) bool {
	return li.VariancePercent().Abs().GreaterThan(threshold)
}

// Validate returns ValidationErrors for the BudgetLineItem.
func (li *BudgetLineItem) Validate() []ValidationError {
	var errs []ValidationError

	if li.BudgetID == uuid.Nil {
		errs = append(errs, ValidationError{Field: "budget_id", Message: "budget_id is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	if li.TenantID == uuid.Nil {
		errs = append(errs, ValidationError{Field: "tenant_id", Message: "tenant_id is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	if li.AccountID == uuid.Nil {
		errs = append(errs, ValidationError{Field: "account_id", Message: "account_id is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	if li.BudgetedAmount.LessThan(decimal.Zero) {
		errs = append(errs, ValidationError{Field: "budgeted_amount", Message: "budgeted_amount must be ≥ 0", Code: "VALUE_OUT_OF_RANGE", Severity: ValidationSeverityError})
	}

	return errs
}
