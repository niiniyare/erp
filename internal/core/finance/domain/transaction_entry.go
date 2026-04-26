package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// TransactionEntry represents a journal entry line within a transaction
// Each entry affects one account and contains either a debit or credit amount (not both)
type TransactionEntry struct {
	ID            uuid.UUID `json:"id"`
	TenantID      uuid.UUID `json:"tenant_id"`
	TransactionID uuid.UUID `json:"transaction_id"`
	EntryNumber   int32     `json:"entry_number"` // Sequential number within transaction
	AccountID     uuid.UUID `json:"account_id"`   // Account being affected

	// Entry amounts - only one should be non-zero per entry
	DebitAmount  decimal.Decimal `json:"debit_amount"`  // Debit amount (left side)
	CreditAmount decimal.Decimal `json:"credit_amount"` // Credit amount (right side)

	// Entry details
	Description string  `json:"description"`         // Entry-specific description
	Reference   *string `json:"reference,omitempty"` // Entry-specific reference

	// Dimensional analysis - for cost accounting and reporting
	CostCenterID *uuid.UUID `json:"cost_center_id,omitempty"` // Cost centre FK (preferred)
	CostCenter   *string    `json:"cost_center,omitempty"`    // Cost centre code — deprecated; use CostCenterID
	Department   *string    `json:"department,omitempty"`     // Department code
	ProjectID    *uuid.UUID `json:"project_id,omitempty"`     // Project reference

	// Multi-currency support
	OriginalCurrency *string         `json:"original_currency,omitempty"` // Original currency if different
	OriginalAmount   decimal.Decimal `json:"original_amount"`             // Amount in original currency
	ExchangeRate     decimal.Decimal `json:"exchange_rate"`               // Rate used for conversion

	// Tax information
	TaxCode   *string         `json:"tax_code,omitempty"` // Tax code applied
	TaxRate   decimal.Decimal `json:"tax_rate"`           // Tax rate percentage
	TaxAmount decimal.Decimal `json:"tax_amount"`         // Calculated tax amount

	// Reconciliation tracking
	Reconciled              bool       `json:"reconciled"`                         // Has been reconciled
	ReconciledDate          *time.Time `json:"reconciled_date,omitempty"`          // When reconciled
	ReconciliationReference *string    `json:"reconciliation_reference,omitempty"` // Reconciliation batch/reference

	// Related account information (populated in queries for convenience)
	Account *Accounts `json:"account,omitempty"`

	// Standard audit timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"` // Soft delete support
}

// CreateEntryRequest represents the request to create a transaction entry
type CreateEntryRequest struct {
	AccountID        uuid.UUID       `json:"account_id" validate:"required"`
	DebitAmount      decimal.Decimal `json:"debit_amount" validate:"gte=0"`
	CreditAmount     decimal.Decimal `json:"credit_amount" validate:"gte=0"`
	Description      string          `json:"description" validate:"required,max=500"`
	Reference        *string         `json:"reference,omitempty" validate:"omitempty,max=100"`
	CostCenter       *string         `json:"cost_center,omitempty" validate:"omitempty,max=20"`
	Department       *string         `json:"department,omitempty" validate:"omitempty,max=20"`
	ProjectID        *uuid.UUID      `json:"project_id,omitempty"`
	OriginalCurrency *string         `json:"original_currency,omitempty" validate:"omitempty,len=3"`
	OriginalAmount   decimal.Decimal `json:"original_amount" validate:"gte=0"`
	ExchangeRate     decimal.Decimal `json:"exchange_rate" validate:"gt=0"`
	TaxCode          *string         `json:"tax_code,omitempty" validate:"omitempty,max=20"`
	TaxRate          decimal.Decimal `json:"tax_rate" validate:"gte=0,lte=100"`
	TaxAmount        decimal.Decimal `json:"tax_amount" validate:"gte=0"`
}

// Validate validates the transaction entry entity
func (e *TransactionEntry) Validate() []ValidationError {
	var errors []ValidationError

	// Validate required fields
	if e.TenantID == uuid.Nil {
		errors = append(errors, ValidationError{
			Field:   "tenant_id",
			Message: "Tenant ID is required",
			Code:    "REQUIRED_FIELD",
		})
	}

	if e.TransactionID == uuid.Nil {
		errors = append(errors, ValidationError{
			Field:   "transaction_id",
			Message: "Transaction ID is required",
			Code:    "REQUIRED_FIELD",
		})
	}

	if e.AccountID == uuid.Nil {
		errors = append(errors, ValidationError{
			Field:   "account_id",
			Message: "Account ID is required",
			Code:    "REQUIRED_FIELD",
		})
	}

	if strings.TrimSpace(e.Description) == "" {
		errors = append(errors, ValidationError{
			Field:   "description",
			Message: "Description is required",
			Code:    "REQUIRED_FIELD",
		})
	} else if len(e.Description) > 500 {
		errors = append(errors, ValidationError{
			Field:   "description",
			Message: "Description must be 500 characters or less",
			Code:    "MAX_LENGTH_EXCEEDED",
		})
	}

	// Validate amounts
	if e.DebitAmount.LessThan(decimal.Zero) {
		errors = append(errors, ValidationError{
			Field:   "debit_amount",
			Message: "Debit amount cannot be negative",
			Code:    "VALUE_OUT_OF_RANGE",
		})
	}

	if e.CreditAmount.LessThan(decimal.Zero) {
		errors = append(errors, ValidationError{
			Field:   "credit_amount",
			Message: "Credit amount cannot be negative",
			Code:    "VALUE_OUT_OF_RANGE",
		})
	}

	// Business rule: Entry must have either debit or credit amount, but not both
	bothAmountsZero := e.DebitAmount.IsZero() && e.CreditAmount.IsZero()
	bothAmountsNonZero := !e.DebitAmount.IsZero() && !e.CreditAmount.IsZero()

	if bothAmountsZero {
		errors = append(errors, ValidationError{
			Field:   "amount",
			Message: "Entry must have either a debit or credit amount",
			Code:    "MISSING_AMOUNT",
		})
	}

	if bothAmountsNonZero {
		errors = append(errors, ValidationError{
			Field:   "amount",
			Message: "Entry cannot have both debit and credit amounts",
			Code:    "DUAL_AMOUNTS",
		})
	}

	// Validate exchange rate
	if e.ExchangeRate.LessThanOrEqual(decimal.Zero) {
		errors = append(errors, ValidationError{
			Field:   "exchange_rate",
			Message: "Exchange rate must be greater than zero",
			Code:    "VALUE_OUT_OF_RANGE",
		})
	}

	// Validate original amount consistency
	if e.OriginalCurrency != nil {
		if len(*e.OriginalCurrency) != 3 {
			errors = append(errors, ValidationError{
				Field:   "original_currency",
				Message: "Original currency must be exactly 3 characters (ISO 4217)",
				Code:    "INVALID_FORMAT",
			})
		}

		if e.OriginalAmount.LessThan(decimal.Zero) {
			errors = append(errors, ValidationError{
				Field:   "original_amount",
				Message: "Original amount cannot be negative",
				Code:    "VALUE_OUT_OF_RANGE",
			})
		}
	} else if !e.OriginalAmount.IsZero() {
		errors = append(errors, ValidationError{
			Field:   "original_amount",
			Message: "Original amount should be zero when no original currency is specified",
			Code:    "INCONSISTENT_CURRENCY_DATA",
		})
	}

	// Validate tax information
	if e.TaxRate.LessThan(decimal.Zero) || e.TaxRate.GreaterThan(decimal.NewFromInt(100)) {
		errors = append(errors, ValidationError{
			Field:   "tax_rate",
			Message: "Tax rate must be between 0 and 100",
			Code:    "VALUE_OUT_OF_RANGE",
		})
	}

	if e.TaxAmount.LessThan(decimal.Zero) {
		errors = append(errors, ValidationError{
			Field:   "tax_amount",
			Message: "Tax amount cannot be negative",
			Code:    "VALUE_OUT_OF_RANGE",
		})
	}

	// Validate tax consistency
	if e.TaxCode != nil && e.TaxRate.IsZero() && !e.TaxAmount.IsZero() {
		errors = append(errors, ValidationError{
			Field:   "tax_rate",
			Message: "Tax rate must be specified when tax code and tax amount are provided",
			Code:    "INCOMPLETE_TAX_DATA",
		})
	}

	// Validate entry number
	if e.EntryNumber <= 0 {
		errors = append(errors, ValidationError{
			Field:   "entry_number",
			Message: "Entry number must be greater than zero",
			Code:    "VALUE_OUT_OF_RANGE",
		})
	}

	// Validate dimensional codes length if provided
	if e.CostCenter != nil && len(*e.CostCenter) > 20 {
		errors = append(errors, ValidationError{
			Field:   "cost_center",
			Message: "Cost center code must be 20 characters or less",
			Code:    "MAX_LENGTH_EXCEEDED",
		})
	}

	if e.Department != nil && len(*e.Department) > 20 {
		errors = append(errors, ValidationError{
			Field:   "department",
			Message: "Department code must be 20 characters or less",
			Code:    "MAX_LENGTH_EXCEEDED",
		})
	}

	return errors
}

// GetEffectiveAmount returns the non-zero amount (debit or credit)
func (e *TransactionEntry) GetEffectiveAmount() decimal.Decimal {
	if !e.DebitAmount.IsZero() {
		return e.DebitAmount
	}
	return e.CreditAmount
}

// IsDebit returns true if this is a debit entry
func (e *TransactionEntry) IsDebit() bool {
	return !e.DebitAmount.IsZero()
}

// IsCredit returns true if this is a credit entry
func (e *TransactionEntry) IsCredit() bool {
	return !e.CreditAmount.IsZero()
}

// GetSignedAmount returns the amount with appropriate sign (positive for debit, negative for credit)
func (e *TransactionEntry) GetSignedAmount() decimal.Decimal {
	if e.IsDebit() {
		return e.DebitAmount
	}
	return e.CreditAmount.Neg()
}

// HasMultiCurrency returns true if this entry involves currency conversion
func (e *TransactionEntry) HasMultiCurrency() bool {
	return e.OriginalCurrency != nil && *e.OriginalCurrency != ""
}

// HasTax returns true if this entry has tax information
func (e *TransactionEntry) HasTax() bool {
	return e.TaxCode != nil && !e.TaxAmount.IsZero()
}

// IsReconciled returns true if this entry has been reconciled
func (e *TransactionEntry) IsReconciled() bool {
	return e.Reconciled && e.ReconciledDate != nil
}

// CalculateConvertedAmount calculates the base currency amount from original currency
func (e *TransactionEntry) CalculateConvertedAmount() decimal.Decimal {
	if !e.HasMultiCurrency() {
		return e.GetEffectiveAmount()
	}
	return e.OriginalAmount.Mul(e.ExchangeRate)
}

// ValidateAmountConsistency validates that converted amounts are consistent
func (e *TransactionEntry) ValidateAmountConsistency() []ValidationError {
	var errors []ValidationError

	if e.HasMultiCurrency() {
		expectedAmount := e.CalculateConvertedAmount()
		actualAmount := e.GetEffectiveAmount()

		// Allow for small rounding differences (0.01)
		tolerance := decimal.NewFromFloat(0.01)
		if expectedAmount.Sub(actualAmount).Abs().GreaterThan(tolerance) {
			errors = append(errors, ValidationError{
				Field:   "amount",
				Message: fmt.Sprintf("Converted amount (%s) does not match expected amount (%s)", actualAmount, expectedAmount),
				Code:    "CURRENCY_CONVERSION_MISMATCH",
			})
		}
	}

	return errors
}

// Validate validates the create entry request
func (r *CreateEntryRequest) Validate() []ValidationError {
	var errors []ValidationError

	// Create a temporary entry for validation
	entry := TransactionEntry{
		TenantID:         uuid.New(), // Dummy value for validation
		TransactionID:    uuid.New(), // Dummy value for validation
		AccountID:        r.AccountID,
		EntryNumber:      1, // Dummy value for validation
		DebitAmount:      r.DebitAmount,
		CreditAmount:     r.CreditAmount,
		Description:      r.Description,
		Reference:        r.Reference,
		CostCenter:       r.CostCenter,
		Department:       r.Department,
		ProjectID:        r.ProjectID,
		OriginalCurrency: r.OriginalCurrency,
		OriginalAmount:   r.OriginalAmount,
		ExchangeRate:     r.ExchangeRate,
		TaxCode:          r.TaxCode,
		TaxRate:          r.TaxRate,
		TaxAmount:        r.TaxAmount,
	}

	// Use the main validation logic
	validationErrors := entry.Validate()

	// Filter out the dummy field errors
	for _, err := range validationErrors {
		if err.Field != "tenant_id" && err.Field != "transaction_id" && err.Field != "entry_number" {
			errors = append(errors, err)
		}
	}

	// Additional request-specific validations
	if r.AccountID == uuid.Nil {
		errors = append(errors, ValidationError{
			Field:   "account_id",
			Message: "Account ID is required",
			Code:    "REQUIRED_FIELD",
		})
	}

	return errors
}

// ValidateBusinessRules validates complex business rules that require external data
func (e *TransactionEntry) ValidateBusinessRules(account *Accounts) []ValidationError {
	var errors []ValidationError

	if account == nil {
		errors = append(errors, ValidationError{
			Field:   "account_id",
			Message: "Account not found",
			Code:    "INVALID_ACCOUNT_REFERENCE",
		})
		return errors
	}

	// Validate account is active
	if !account.IsActive {
		errors = append(errors, ValidationError{
			Field:   "account_id",
			Message: "Cannot post to inactive account",
			Code:    "INACTIVE_ACCOUNT",
		})
	}

	// Validate account allows the entry type (manual vs system)
	if !account.AllowManualEntries {
		// FIXME: This validation would need context about whether this is a manual entry
		// For now, we'll skip this check as it requires transaction context
	}

	// Validate currency consistency
	if account.CurrencyCode != nil && e.OriginalCurrency != nil {
		if *account.CurrencyCode != *e.OriginalCurrency {
			if !account.IsMultiCurrency {
				errors = append(errors, ValidationError{
					Field:   "original_currency",
					Message: fmt.Sprintf("Account currency (%s) does not match entry currency (%s) and account does not allow multi-currency", *account.CurrencyCode, *e.OriginalCurrency),
					Code:    "CURRENCY_MISMATCH",
				})
			}
		}
	}

	// Validate reference requirement
	if account.RequireReference && e.Reference == nil {
		errors = append(errors, ValidationError{
			Field:   "reference",
			Message: "Account requires a reference for all entries",
			Code:    "REFERENCE_REQUIRED",
		})
	}

	// Validate normal balance consistency (warning only)
	entryIncreasesBalance := false
	if account.NormalBalance == NormalBalanceDebit && e.IsDebit() {
		entryIncreasesBalance = true
	} else if account.NormalBalance == NormalBalanceCredit && e.IsCredit() {
		entryIncreasesBalance = true
	}

	// This is informational - not an error, but could be used for warnings
	_ = entryIncreasesBalance

	return errors
}

// CreateCounterEntry creates a balancing entry for this entry
// Useful for creating simple two-entry transactions
func (e *TransactionEntry) CreateCounterEntry(accountID uuid.UUID, description string) *TransactionEntry {
	counter := &TransactionEntry{
		ID:               uuid.New(),
		TenantID:         e.TenantID,
		TransactionID:    e.TransactionID,
		EntryNumber:      e.EntryNumber + 1,
		AccountID:        accountID,
		Description:      description,
		Reference:        e.Reference,
		CostCenter:       e.CostCenter,
		Department:       e.Department,
		ProjectID:        e.ProjectID,
		OriginalCurrency: e.OriginalCurrency,
		OriginalAmount:   e.OriginalAmount,
		ExchangeRate:     e.ExchangeRate,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Swap debit/credit amounts to create balancing entry
	if e.IsDebit() {
		counter.CreditAmount = e.DebitAmount
		counter.DebitAmount = decimal.Zero
	} else {
		counter.DebitAmount = e.CreditAmount
		counter.CreditAmount = decimal.Zero
	}

	return counter
}

// Clone creates a deep copy of the transaction entry
func (e *TransactionEntry) Clone() *TransactionEntry {
	clone := *e
	clone.ID = uuid.New()
	clone.CreatedAt = time.Now()
	clone.UpdatedAt = time.Now()
	clone.DeletedAt = nil

	// Deep copy pointers
	if e.Reference != nil {
		ref := *e.Reference
		clone.Reference = &ref
	}
	if e.CostCenter != nil {
		cc := *e.CostCenter
		clone.CostCenter = &cc
	}
	if e.Department != nil {
		dept := *e.Department
		clone.Department = &dept
	}
	if e.ProjectID != nil {
		pid := *e.ProjectID
		clone.ProjectID = &pid
	}
	if e.OriginalCurrency != nil {
		curr := *e.OriginalCurrency
		clone.OriginalCurrency = &curr
	}
	if e.TaxCode != nil {
		tax := *e.TaxCode
		clone.TaxCode = &tax
	}
	if e.ReconciledDate != nil {
		rd := *e.ReconciledDate
		clone.ReconciledDate = &rd
	}
	if e.ReconciliationReference != nil {
		rr := *e.ReconciliationReference
		clone.ReconciliationReference = &rr
	}

	return &clone
}

// MarkReconciled marks the entry as reconciled
func (e *TransactionEntry) MarkReconciled(reconciliationRef string) {
	now := time.Now()
	e.Reconciled = true
	e.ReconciledDate = &now
	e.ReconciliationReference = &reconciliationRef
	e.UpdatedAt = now
}

// UnmarkReconciled removes reconciliation status
func (e *TransactionEntry) UnmarkReconciled() {
	e.Reconciled = false
	e.ReconciledDate = nil
	e.ReconciliationReference = nil
	e.UpdatedAt = time.Now()
}

// GetDimensionalAnalysis returns a map of dimensional attributes
func (e *TransactionEntry) GetDimensionalAnalysis() map[string]any {
	dimensions := make(map[string]any)

	if e.CostCenter != nil {
		dimensions["cost_center"] = *e.CostCenter
	}
	if e.Department != nil {
		dimensions["department"] = *e.Department
	}
	if e.ProjectID != nil {
		dimensions["project_id"] = *e.ProjectID
	}

	return dimensions
}

// TODO: Add support for entry-level attachments and supporting documents
// TODO: Implement entry templates for common entry patterns
// TODO: Add support for entry-level approval workflow for high-value entries
// TODO: Implement automated entry matching for bank reconciliation
// NOTE: Consider adding entry-level tags for enhanced categorization and reporting
// NOTE: Future enhancement: Add support for entry-level analytics and AI-powered categorization
// NOTE: Consider implementing entry-level audit trail with detailed change history
