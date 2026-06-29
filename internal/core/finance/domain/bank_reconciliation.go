package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// BankReconciliationStatus tracks where a reconciliation session is in its lifecycle.
type BankReconciliationStatus string

const (
	BankReconciliationStatusDraft      BankReconciliationStatus = "DRAFT"       // Being prepared; no items matched yet
	BankReconciliationStatusInProgress BankReconciliationStatus = "IN_PROGRESS" // Items are being matched
	BankReconciliationStatusCompleted  BankReconciliationStatus = "COMPLETED"   // All items matched or explained; awaiting sign-off
	BankReconciliationStatusApproved   BankReconciliationStatus = "APPROVED"    // Signed off by an authorised approver
	BankReconciliationStatusVoided     BankReconciliationStatus = "VOIDED"      // Cancelled without completion
)

// IsValid returns true if the status is a recognised value.
func (s BankReconciliationStatus) IsValid() bool {
	switch s {
	case BankReconciliationStatusDraft, BankReconciliationStatusInProgress,
		BankReconciliationStatusCompleted, BankReconciliationStatusApproved,
		BankReconciliationStatusVoided:
		return true
	default:
		return false
	}
}

// String returns the string representation.
func (s BankReconciliationStatus) String() string { return string(s) }

// IsEditable returns true when items may be matched or unmatched.
func (s BankReconciliationStatus) IsEditable() bool {
	return s == BankReconciliationStatusDraft || s == BankReconciliationStatusInProgress
}

// ReconciliationItemStatus classifies whether a bank statement line has been matched.
type ReconciliationItemStatus string

const (
	ReconciliationItemStatusUnmatched  ReconciliationItemStatus = "UNMATCHED"   // Not yet matched to a GL entry
	ReconciliationItemStatusMatched    ReconciliationItemStatus = "MATCHED"     // Matched to one or more GL entries
	ReconciliationItemStatusExplained  ReconciliationItemStatus = "EXPLAINED"   // Timing difference; noted but not matched
	ReconciliationItemStatusDisputed   ReconciliationItemStatus = "DISPUTED"    // Disputed with the bank
	ReconciliationItemStatusWrittenOff ReconciliationItemStatus = "WRITTEN_OFF" // Difference written off via journal
)

// IsValid returns true if the status is a recognised value.
func (s ReconciliationItemStatus) IsValid() bool {
	switch s {
	case ReconciliationItemStatusUnmatched, ReconciliationItemStatusMatched,
		ReconciliationItemStatusExplained, ReconciliationItemStatusDisputed,
		ReconciliationItemStatusWrittenOff:
		return true
	default:
		return false
	}
}

// BankReconciliation is the header record for a bank account reconciliation session.
type BankReconciliation struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`

	// The GL bank account being reconciled
	AccountID uuid.UUID `json:"account_id"`

	// Statement details
	BankStatementDate   time.Time       `json:"bank_statement_date"`   // Date on the bank statement
	StatementOpeningBal decimal.Decimal `json:"statement_opening_bal"` // Opening balance per bank
	StatementClosingBal decimal.Decimal `json:"statement_closing_bal"` // Closing balance per bank
	CurrencyCode        string          `json:"currency_code"`         // ISO 4217

	// GL balances (populated by the system)
	GLOpeningBalance decimal.Decimal `json:"gl_opening_balance"` // GL balance at period start
	GLClosingBalance decimal.Decimal `json:"gl_closing_balance"` // GL balance at reconciliation date

	// Reconciled difference
	Difference decimal.Decimal `json:"difference"` // statement_closing - gl_closing; should be 0 when complete

	Status BankReconciliationStatus `json:"status"`

	// Timing differences (populated as items are matched)
	OutstandingDeposits   decimal.Decimal `json:"outstanding_deposits"`   // Deposits in transit
	OutstandingPayments   decimal.Decimal `json:"outstanding_payments"`   // Cheques not yet cleared
	UnexplainedDifference decimal.Decimal `json:"unexplained_difference"` // Remaining difference after adjustments

	// Approval metadata
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CompletedBy *uuid.UUID `json:"completed_by,omitempty"`
	ApprovedAt  *time.Time `json:"approved_at,omitempty"`
	ApprovedBy  *uuid.UUID `json:"approved_by,omitempty"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	CreatedBy uuid.UUID  `json:"created_by"`
	UpdatedBy *uuid.UUID `json:"updated_by,omitempty"`
}

// IsBalanced returns true when the reconciliation difference is zero.
func (r *BankReconciliation) IsBalanced() bool {
	return r.Difference.IsZero()
}

// Validate returns ValidationErrors for the BankReconciliation header.
func (r *BankReconciliation) Validate() []ValidationError {
	var errs []ValidationError

	if r.TenantID == uuid.Nil {
		errs = append(errs, ValidationError{Field: "tenant_id", Message: "tenant_id is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	if r.AccountID == uuid.Nil {
		errs = append(errs, ValidationError{Field: "account_id", Message: "account_id is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	if r.BankStatementDate.IsZero() {
		errs = append(errs, ValidationError{Field: "bank_statement_date", Message: "bank_statement_date is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	if len(r.CurrencyCode) != 3 {
		errs = append(errs, ValidationError{Field: "currency_code", Message: "currency_code must be a 3-character ISO 4217 code", Code: "INVALID_FORMAT", Severity: ValidationSeverityError})
	}
	if !r.Status.IsValid() {
		errs = append(errs, ValidationError{Field: "status", Message: fmt.Sprintf("invalid status: %s", r.Status), Code: "INVALID_VALUE", Severity: ValidationSeverityError})
	}

	return errs
}

// BankStatementItem represents a single line from a bank statement.
type BankStatementItem struct {
	ID               uuid.UUID `json:"id"`
	ReconciliationID uuid.UUID `json:"reconciliation_id"`
	TenantID         uuid.UUID `json:"tenant_id"`

	// Statement line data
	ValueDate   time.Time       `json:"value_date"`          // Date funds cleared
	Description string          `json:"description"`         // Bank's own narration
	Reference   *string         `json:"reference,omitempty"` // Bank reference / cheque number
	Amount      decimal.Decimal `json:"amount"`              // Positive = credit (inflow); negative = debit (outflow)
	RunningBal  decimal.Decimal `json:"running_bal"`         // Balance after this entry

	Status ReconciliationItemStatus `json:"status"`

	// Matching metadata (populated when status → MATCHED)
	MatchedEntryIDs []uuid.UUID `json:"matched_entry_ids,omitempty"` // GL TransactionEntry IDs
	MatchedAt       *time.Time  `json:"matched_at,omitempty"`
	MatchedBy       *uuid.UUID  `json:"matched_by,omitempty"`
	MatchNote       *string     `json:"match_note,omitempty"` // Explanation for EXPLAINED / WRITTEN_OFF items

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// IsCredit returns true when this line represents funds entering the account.
func (i *BankStatementItem) IsCredit() bool {
	return i.Amount.GreaterThan(decimal.Zero)
}

// IsDebit returns true when this line represents funds leaving the account.
func (i *BankStatementItem) IsDebit() bool {
	return i.Amount.LessThan(decimal.Zero)
}

// Validate returns ValidationErrors for the BankStatementItem.
func (i *BankStatementItem) Validate() []ValidationError {
	var errs []ValidationError

	if i.ReconciliationID == uuid.Nil {
		errs = append(errs, ValidationError{Field: "reconciliation_id", Message: "reconciliation_id is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	if i.TenantID == uuid.Nil {
		errs = append(errs, ValidationError{Field: "tenant_id", Message: "tenant_id is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	if i.ValueDate.IsZero() {
		errs = append(errs, ValidationError{Field: "value_date", Message: "value_date is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	if i.Amount.IsZero() {
		errs = append(errs, ValidationError{Field: "amount", Message: "amount must be non-zero", Code: "INVALID_VALUE", Severity: ValidationSeverityError})
	}
	if !i.Status.IsValid() {
		errs = append(errs, ValidationError{Field: "status", Message: fmt.Sprintf("invalid status: %s", i.Status), Code: "INVALID_VALUE", Severity: ValidationSeverityError})
	}

	return errs
}
