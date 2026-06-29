package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ReconciliationStatus tracks the lifecycle of a bank statement reconciliation.
type ReconciliationStatus string

const (
	ReconciliationStatusDraft      ReconciliationStatus = "DRAFT"       // Statement imported, not yet started
	ReconciliationStatusInProgress ReconciliationStatus = "IN_PROGRESS" // Matching in progress
	ReconciliationStatusCompleted  ReconciliationStatus = "COMPLETED"   // Fully reconciled
	ReconciliationStatusVoided     ReconciliationStatus = "VOIDED"      // Cancelled
)

// BankStatement represents an imported bank statement for a specific GL bank account.
type BankStatement struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`

	// GL account this statement belongs to (must be a bank/cash type account)
	AccountID uuid.UUID `json:"account_id"`

	// Statement details
	StatementReference string    `json:"statement_reference"` // Bank's statement number
	StatementDate      time.Time `json:"statement_date"`      // End date of statement period
	StartDate          time.Time `json:"start_date"`          // Beginning of statement period
	CurrencyCode       string    `json:"currency_code"`       // ISO 4217

	// Balances as per bank statement
	OpeningBalance decimal.Decimal `json:"opening_balance"`
	ClosingBalance decimal.Decimal `json:"closing_balance"`

	Status ReconciliationStatus `json:"status"`

	// Reconciled totals (updated as lines are matched)
	MatchedCount     int             `json:"matched_count"`
	UnmatchedCount   int             `json:"unmatched_count"`
	DifferenceAmount decimal.Decimal `json:"difference_amount"` // ClosingBalance - GL balance at statement date

	// Lines (loaded on demand)
	Lines []*BankStatementLine `json:"lines,omitempty"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	CreatedBy uuid.UUID  `json:"created_by"`
	UpdatedBy *uuid.UUID `json:"updated_by,omitempty"`
}

// Validate returns validation errors for the BankStatement.
func (s *BankStatement) Validate() []ValidationError {
	var errs []ValidationError
	if s.TenantID == uuid.Nil {
		errs = append(errs, ValidationError{Field: "tenant_id", Message: "tenant_id is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	if s.AccountID == uuid.Nil {
		errs = append(errs, ValidationError{Field: "account_id", Message: "account_id is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	if strings.TrimSpace(s.StatementReference) == "" {
		errs = append(errs, ValidationError{Field: "statement_reference", Message: "statement_reference is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	if s.StatementDate.IsZero() {
		errs = append(errs, ValidationError{Field: "statement_date", Message: "statement_date is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	if len(s.CurrencyCode) != 3 {
		errs = append(errs, ValidationError{Field: "currency_code", Message: "currency_code must be a 3-letter ISO code", Code: "INVALID_FORMAT", Severity: ValidationSeverityError})
	}
	return errs
}

// BankStatementLine represents a single line on an imported bank statement.
type BankStatementLine struct {
	ID          uuid.UUID `json:"id"`
	StatementID uuid.UUID `json:"statement_id"`
	TenantID    uuid.UUID `json:"tenant_id"`

	// Dates
	TransactionDate time.Time  `json:"transaction_date"` // Date on bank statement
	ValueDate       *time.Time `json:"value_date,omitempty"`

	// Description from bank
	Description string  `json:"description"`
	Reference   *string `json:"reference,omitempty"` // Bank's own reference / cheque number

	// Amount: positive = money in (deposit/credit), negative = money out (withdrawal/debit)
	// Stored as two separate columns to mirror journal entry style.
	DebitAmount  decimal.Decimal `json:"debit_amount"`  // Money out of the account
	CreditAmount decimal.Decimal `json:"credit_amount"` // Money into the account
	Balance      decimal.Decimal `json:"balance"`       // Running balance after this line

	// Reconciliation
	IsReconciled bool       `json:"is_reconciled"`
	ReconciledAt *time.Time `json:"reconciled_at,omitempty"`
	ReconciledBy *uuid.UUID `json:"reconciled_by,omitempty"`

	// Matched journal entry (set when this line is matched to a TransactionEntry)
	MatchedEntryID *uuid.UUID `json:"matched_entry_id,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}
