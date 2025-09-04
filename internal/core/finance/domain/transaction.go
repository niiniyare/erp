package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Transaction represents a financial transaction header
// A transaction contains one or more entries that must balance (debits = credits)
type Transaction struct {
	ID       uuid.UUID  `json:"id"`
	TenantID uuid.UUID  `json:"tenant_id"`
	EntityID *uuid.UUID `json:"entity_id,omitempty"`

	// Transaction identification
	TransactionNumber string            `json:"transaction_number"` // Unique sequential number
	TransactionType   TransactionType   `json:"transaction_type"`   // Type/source of transaction
	TransactionStatus TransactionStatus `json:"transaction_status"` // Current lifecycle status

	// Transaction dates - critical for accounting periods
	TransactionDate time.Time  `json:"transaction_date"`       // When transaction occurred
	PostingDate     *time.Time `json:"posting_date,omitempty"` // When posted to ledger
	DueDate         *time.Time `json:"due_date,omitempty"`     // Payment due date (if applicable)

	// Transaction details
	Description       string  `json:"description"`                  // Primary description
	ReferenceNumber   *string `json:"reference_number,omitempty"`   // Internal reference
	ExternalReference *string `json:"external_reference,omitempty"` // External system reference

	// Financial information
	CurrencyCode      string          `json:"currency_code"`       // ISO currency code
	ExchangeRate      decimal.Decimal `json:"exchange_rate"`       // Exchange rate to base currency
	TotalDebitAmount  decimal.Decimal `json:"total_debit_amount"`  // Sum of all debit entries
	TotalCreditAmount decimal.Decimal `json:"total_credit_amount"` // Sum of all credit entries

	// Source and traceability
	SourceModule       *string    `json:"source_module,omitempty"`        // Originating module
	SourceDocumentType *string    `json:"source_document_type,omitempty"` // Document type that created this
	SourceDocumentID   *uuid.UUID `json:"source_document_id,omitempty"`   // Source document reference
	BatchID            *uuid.UUID `json:"batch_id,omitempty"`             // Batch processing reference

	// Approval workflow
	ApprovalRequired bool           `json:"approval_required"`        // Requires approval before posting
	ApprovalStatus   ApprovalStatus `json:"approval_status"`          // Current approval state
	ApprovedBy       *uuid.UUID     `json:"approved_by,omitempty"`    // User who approved
	ApprovedAt       *time.Time     `json:"approved_at,omitempty"`    // When approved
	ApprovalNotes    *string        `json:"approval_notes,omitempty"` // Approval/rejection notes

	// Recurring transaction support
	IsRecurring        bool       `json:"is_recurring"`                  // Is this a recurring transaction
	RecurringFrequency *string    `json:"recurring_frequency,omitempty"` // Frequency pattern
	NextRecurringDate  *time.Time `json:"next_recurring_date,omitempty"` // Next occurrence date

	// Reversal tracking
	IsReversed              bool       `json:"is_reversed"`                          // Has been reversed
	ReversedByTransactionID *uuid.UUID `json:"reversed_by_transaction_id,omitempty"` // Reversing transaction
	ReversalReason          *string    `json:"reversal_reason,omitempty"`            // Reason for reversal

	// Audit and validation
	Version          int32             `json:"version"`                     // Optimistic locking
	ValidationStatus ValidationStatus  `json:"validation_status"`           // Current validation state
	ValidationErrors []ValidationError `json:"validation_errors,omitempty"` // Validation issues
	RejectionReason  *RejectionReason  `json:"rejection_reason,omitempty"`  // Reason for rejection

	// Metadata - flexible attributes for extensions
	TransactionAttributes map[string]any `json:"transaction_attributes,omitempty"`
	AttachmentIds         []string       `json:"attachment_ids"`
	Tags                  []string       `json:"tags"`

	// Transaction entries - the actual accounting entries
	Entries []TransactionEntry `json:"entries,omitempty"`

	// Standard audit timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"` // Soft delete support
	CreatedBy uuid.UUID  `json:"created_by"`
	UpdatedBy *uuid.UUID `json:"updated_by,omitempty"`
	PostedBy  *uuid.UUID `json:"posted_by,omitempty"` // User who posted the transaction
	PostedAt  *time.Time `json:"posted_at,omitempty"` // When transaction was posted
}

// TransactionWithEntries represents a transaction with its related transaction entries
type TransactionWithEntries struct {
	Transaction Transaction        `json:"transaction"`
	Entries     []TransactionEntry `json:"entries"`
}

// CreateTransactionRequest represents the request to create a new transaction
type CreateTransactionRequest struct {
	EntityID              *uuid.UUID           `json:"entity_id,omitempty"`
	TransactionNumber     string               `json:"transaction_number" validate:"required,max=50"`
	TransactionType       TransactionType      `json:"transaction_type" validate:"required"`
	TransactionStatus     TransactionStatus    `json:"transaction_status"` // Current lifecycle status
	TransactionDate       time.Time            `json:"transaction_date" validate:"required"`
	PostingDate           *time.Time           `json:"posting_date,omitempty"`
	DueDate               *time.Time           `json:"due_date,omitempty"`
	Description           string               `json:"description" validate:"required,max=500"`
	ReferenceNumber       *string              `json:"reference_number,omitempty" validate:"omitempty,max=100"`
	ExternalReference     *string              `json:"external_reference,omitempty" validate:"omitempty,max=100"`
	CurrencyCode          string               `json:"currency_code" validate:"required,len=3"`
	ExchangeRate          decimal.Decimal      `json:"exchange_rate" validate:"gt=0"`
	SourceModule          *string              `json:"source_module,omitempty" validate:"omitempty,max=100"`
	SourceDocumentType    *string              `json:"source_document_type,omitempty" validate:"omitempty,max=100"`
	SourceDocumentID      *uuid.UUID           `json:"source_document_id,omitempty"`
	BatchID               *uuid.UUID           `json:"batch_id,omitempty"`
	ApprovalRequired      bool                 `json:"approval_required"`
	IsRecurring           bool                 `json:"is_recurring"`
	RecurringFrequency    *string              `json:"recurring_frequency,omitempty" validate:"omitempty,max=50"`
	NextRecurringDate     *time.Time           `json:"next_recurring_date,omitempty"`
	TransactionAttributes map[string]any       `json:"transaction_attributes,omitempty"`
	Entries               []CreateEntryRequest `json:"entries" validate:"required,min=2,dive"`
	Memo                  string               `json:"memo"`
	AttachmentIds         []string             `json:"attachment_ids"`
	Tags                  []string             `json:"tags"`
	CreatedBy             uuid.UUID            `json:"created_by"`
	UpdatedBy             *uuid.UUID           `json:"updated_by,omitempty"`
	PostedBy              *uuid.UUID           `json:"posted_by,omitempty"` // User who posted the transaction
}

// TransactionSummary represents aggregated transaction data for reporting
type TransactionSummary struct {
	TransactionType   TransactionType   `json:"transaction_type"`
	TransactionStatus TransactionStatus `json:"transaction_status"`
	TransactionCount  int64             `json:"transaction_count"`
	TotalDebit        decimal.Decimal   `json:"total_debit"`
	TotalCredit       decimal.Decimal   `json:"total_credit"`
}

// Validate validates the transaction entity
func (t *Transaction) Validate() []ValidationError {
	var errors []ValidationError

	// Validate required fields
	if t.TenantID == uuid.Nil {
		errors = append(errors, ValidationError{
			Field:   "tenant_id",
			Message: "Tenant ID is required",
			Code:    "REQUIRED_FIELD",
		})
	}

	if strings.TrimSpace(t.TransactionNumber) == "" {
		errors = append(errors, ValidationError{
			Field:   "transaction_number",
			Message: "Transaction number is required",
			Code:    "REQUIRED_FIELD",
		})
	} else if len(t.TransactionNumber) > 50 {
		errors = append(errors, ValidationError{
			Field:   "transaction_number",
			Message: "Transaction number must be 50 characters or less",
			Code:    "MAX_LENGTH_EXCEEDED",
		})
	}

	if strings.TrimSpace(t.Description) == "" {
		errors = append(errors, ValidationError{
			Field:   "description",
			Message: "Description is required",
			Code:    "REQUIRED_FIELD",
		})
	} else if len(t.Description) > 500 {
		errors = append(errors, ValidationError{
			Field:   "description",
			Message: "Description must be 500 characters or less",
			Code:    "MAX_LENGTH_EXCEEDED",
		})
	}

	// Validate transaction type
	if !t.TransactionType.IsValid() {
		errors = append(errors, ValidationError{
			Field:   "transaction_type",
			Message: "Invalid transaction type",
			Code:    "INVALID_VALUE",
		})
	}

	// Validate transaction status
	if !t.TransactionStatus.IsValid() {
		errors = append(errors, ValidationError{
			Field:   "transaction_status",
			Message: "Invalid transaction status",
			Code:    "INVALID_VALUE",
		})
	}

	// Validate approval status
	if !t.ApprovalStatus.IsValid() {
		errors = append(errors, ValidationError{
			Field:   "approval_status",
			Message: "Invalid approval status",
			Code:    "INVALID_VALUE",
		})
	}

	// Validate currency code
	if len(t.CurrencyCode) != 3 {
		errors = append(errors, ValidationError{
			Field:   "currency_code",
			Message: "Currency code must be exactly 3 characters (ISO 4217)",
			Code:    "INVALID_FORMAT",
		})
	}

	// Validate exchange rate
	if t.ExchangeRate.LessThanOrEqual(decimal.Zero) {
		errors = append(errors, ValidationError{
			Field:   "exchange_rate",
			Message: "Exchange rate must be greater than zero",
			Code:    "VALUE_OUT_OF_RANGE",
		})
	}

	// Validate transaction date
	if t.TransactionDate.IsZero() {
		errors = append(errors, ValidationError{
			Field:   "transaction_date",
			Message: "Transaction date is required",
			Code:    "REQUIRED_FIELD",
		})
	}

	// Validate posting date if provided
	if t.PostingDate != nil && t.PostingDate.Before(t.TransactionDate) {
		errors = append(errors, ValidationError{
			Field:   "posting_date",
			Message: "Posting date cannot be before transaction date",
			Code:    "INVALID_DATE_SEQUENCE",
		})
	}

	// Business rule validations
	if t.ApprovalRequired && t.ApprovalStatus == ApprovalStatusNotRequired {
		errors = append(errors, ValidationError{
			Field:   "approval_status",
			Message: "Approval status cannot be 'NOT_REQUIRED' when approval is required",
			Code:    "INCONSISTENT_APPROVAL_SETTINGS",
		})
	}

	if !t.ApprovalRequired && t.ApprovalStatus != ApprovalStatusNotRequired {
		errors = append(errors, ValidationError{
			Field:   "approval_status",
			Message: "Approval status must be 'NOT_REQUIRED' when approval is not required",
			Code:    "INCONSISTENT_APPROVAL_SETTINGS",
		})
	}

	// Validate recurring transaction settings
	if t.IsRecurring {
		if t.RecurringFrequency == nil || strings.TrimSpace(*t.RecurringFrequency) == "" {
			errors = append(errors, ValidationError{
				Field:   "recurring_frequency",
				Message: "Recurring frequency is required for recurring transactions",
				Code:    "REQUIRED_FIELD_CONDITIONAL",
			})
		}

		if t.NextRecurringDate == nil {
			errors = append(errors, ValidationError{
				Field:   "next_recurring_date",
				Message: "Next recurring date is required for recurring transactions",
				Code:    "REQUIRED_FIELD_CONDITIONAL",
			})
		}
	}

	// Validate reversal settings
	if t.IsReversed && t.ReversedByTransactionID == nil {
		errors = append(errors, ValidationError{
			Field:   "reversed_by_transaction_id",
			Message: "Reversed by transaction ID is required when transaction is marked as reversed",
			Code:    "REQUIRED_FIELD_CONDITIONAL",
		})
	}

	// Validate entries exist and balance
	if len(t.Entries) < 2 {
		errors = append(errors, ValidationError{
			Field:   "entries",
			Message: "Transaction must have at least 2 entries",
			Code:    "INSUFFICIENT_ENTRIES",
		})
	}

	// Check if transaction balances
	if !t.IsBalanced() {
		errors = append(errors, ValidationError{
			Field:   "entries",
			Message: fmt.Sprintf("Transaction does not balance: debits=%s, credits=%s", t.TotalDebitAmount, t.TotalCreditAmount),
			Code:    "TRANSACTION_UNBALANCED",
		})
	}

	// Validate individual entries
	for i, entry := range t.Entries {
		entryErrors := entry.Validate()
		for _, err := range entryErrors {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("entries[%d].%s", i, err.Field),
				Message: err.Message,
				Code:    err.Code,
			})
		}
	}

	return errors
}

// IsBalanced checks if the transaction debits equal credits
func (t *Transaction) IsBalanced() bool {
	return t.TotalDebitAmount.Equal(t.TotalCreditAmount)
}

// CanBePosted checks if a transaction can be posted to the ledger
func (t *Transaction) CanBePosted() bool {
	// Must be balanced
	if !t.IsBalanced() {
		return false
	}

	// Must have valid status
	switch t.TransactionStatus {
	case TransactionStatusApproved, TransactionStatusDraft:
		// If approval is required, must be approved
		if t.ApprovalRequired && t.ApprovalStatus != ApprovalStatusApproved {
			return false
		}
		return true
	default:
		return false
	}
}

// CanBeReversed checks if a transaction can be reversed
func (t *Transaction) CanBeReversed() bool {
	return t.TransactionStatus == TransactionStatusPosted && !t.IsReversed
}

// CanBeEdited checks if a transaction can be modified
func (t *Transaction) CanBeEdited() bool {
	return t.TransactionStatus.IsEditable() && !t.IsReversed
}

// CalculateTotals recalculates total debit and credit amounts from entries
func (t *Transaction) CalculateTotals() {
	totalDebit := decimal.Zero
	totalCredit := decimal.Zero

	for _, entry := range t.Entries {
		totalDebit = totalDebit.Add(entry.DebitAmount)
		totalCredit = totalCredit.Add(entry.CreditAmount)
	}

	t.TotalDebitAmount = totalDebit
	t.TotalCreditAmount = totalCredit
}

// GetNetAmount returns the net amount of the transaction
func (t *Transaction) GetNetAmount() decimal.Decimal {
	return t.TotalDebitAmount.Sub(t.TotalCreditAmount).Abs()
}

// IsSystemGenerated returns true if the transaction was generated by the system
func (t *Transaction) IsSystemGenerated() bool {
	return t.TransactionType == TransactionTypeSystem ||
		t.TransactionType == TransactionTypeRecurring ||
		t.TransactionType == TransactionTypeClosing
}

// RequiresApproval determines if the transaction requires approval based on type
func (t *Transaction) RequiresApproval() bool {
	return t.TransactionType.RequiresApproval()
}

// MarshalTransactionAttributes marshals transaction attributes to JSON
func (t *Transaction) MarshalTransactionAttributes() ([]byte, error) {
	if t.TransactionAttributes == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(t.TransactionAttributes)
}

// UnmarshalTransactionAttributes unmarshals transaction attributes from JSON
func (t *Transaction) UnmarshalTransactionAttributes(data []byte) error {
	if len(data) == 0 {
		t.TransactionAttributes = make(map[string]any)
		return nil
	}
	return json.Unmarshal(data, &t.TransactionAttributes)
}

// Validate validates the create transaction request
func (r *CreateTransactionRequest) Validate() []ValidationError {
	var errors []ValidationError

	// Basic field validations
	if strings.TrimSpace(r.TransactionNumber) == "" {
		errors = append(errors, ValidationError{
			Field:   "transaction_number",
			Message: "Transaction number is required",
			Code:    "REQUIRED_FIELD",
		})
	}

	if strings.TrimSpace(r.Description) == "" {
		errors = append(errors, ValidationError{
			Field:   "description",
			Message: "Description is required",
			Code:    "REQUIRED_FIELD",
		})
	}

	if !r.TransactionType.IsValid() {
		errors = append(errors, ValidationError{
			Field:   "transaction_type",
			Message: "Invalid transaction type",
			Code:    "INVALID_VALUE",
		})
	}

	if len(r.CurrencyCode) != 3 {
		errors = append(errors, ValidationError{
			Field:   "currency_code",
			Message: "Currency code must be exactly 3 characters (ISO 4217)",
			Code:    "INVALID_FORMAT",
		})
	}

	if r.ExchangeRate.LessThanOrEqual(decimal.Zero) {
		errors = append(errors, ValidationError{
			Field:   "exchange_rate",
			Message: "Exchange rate must be greater than zero",
			Code:    "VALUE_OUT_OF_RANGE",
		})
	}

	// Validate entries
	if len(r.Entries) < 2 {
		errors = append(errors, ValidationError{
			Field:   "entries",
			Message: "Transaction must have at least 2 entries",
			Code:    "INSUFFICIENT_ENTRIES",
		})
	}

	// Check if entries balance
	totalDebit := decimal.Zero
	totalCredit := decimal.Zero
	for _, entry := range r.Entries {
		totalDebit = totalDebit.Add(entry.DebitAmount)
		totalCredit = totalCredit.Add(entry.CreditAmount)
	}

	if !totalDebit.Equal(totalCredit) {
		errors = append(errors, ValidationError{
			Field:   "entries",
			Message: fmt.Sprintf("Transaction entries do not balance: debits=%s, credits=%s", totalDebit, totalCredit),
			Code:    "TRANSACTION_UNBALANCED",
		})
	}

	// Validate recurring transaction settings
	if r.IsRecurring {
		if r.RecurringFrequency == nil || strings.TrimSpace(*r.RecurringFrequency) == "" {
			errors = append(errors, ValidationError{
				Field:   "recurring_frequency",
				Message: "Recurring frequency is required for recurring transactions",
				Code:    "REQUIRED_FIELD_CONDITIONAL",
			})
		}
	}

	// Validate individual entries
	for i, entry := range r.Entries {
		entryErrors := entry.Validate()
		for _, err := range entryErrors {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("entries[%d].%s", i, err.Field),
				Message: err.Message,
				Code:    err.Code,
			})
		}
	}

	return errors
}

// ValidateBusinessRules validates complex business rules that require external data
func (t *Transaction) ValidateBusinessRules(accounts map[uuid.UUID]*Accounts) []ValidationError {
	var errors []ValidationError

	// Validate that all referenced accounts exist and are active
	for i, entry := range t.Entries {
		account, exists := accounts[entry.AccountID]
		if !exists {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("entries[%d].account_id", i),
				Message: "Referenced account does not exist",
				Code:    "INVALID_ACCOUNT_REFERENCE",
			})
			continue
		}

		if !account.IsActive {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("entries[%d].account_id", i),
				Message: "Cannot post to inactive account",
				Code:    "INACTIVE_ACCOUNT",
			})
		}

		if !account.CanAcceptManualEntries() && t.TransactionType == TransactionTypeManual {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("entries[%d].account_id", i),
				Message: "Account does not allow manual entries",
				Code:    "MANUAL_ENTRIES_NOT_ALLOWED",
			})
		}

		if account.RequireReference && entry.Reference == nil {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("entries[%d].reference", i),
				Message: "Account requires a reference for all entries",
				Code:    "REFERENCE_REQUIRED",
			})
		}
	}

	// Validate posting date is within open accounting period
	if t.PostingDate != nil {
		// TODO: Add accounting period validation
		// This would require access to accounting period configuration
	}

	return errors
}

// CreateReversalTransaction creates a reversal transaction for this transaction
func (t *Transaction) CreateReversalTransaction(reason string, createdBy uuid.UUID) (*Transaction, error) {
	if !t.CanBeReversed() {
		return nil, errors.New("transaction cannot be reversed")
	}

	reversal := &Transaction{
		ID:                uuid.New(),
		TenantID:          t.TenantID,
		EntityID:          t.EntityID,
		TransactionNumber: fmt.Sprintf("REV-%s", t.TransactionNumber),
		TransactionType:   TransactionTypeAdjustment,
		TransactionStatus: TransactionStatusDraft,
		TransactionDate:   time.Now(),
		Description:       fmt.Sprintf("Reversal of %s - %s", t.TransactionNumber, reason),
		CurrencyCode:      t.CurrencyCode,
		ExchangeRate:      t.ExchangeRate,
		SourceModule:      t.SourceModule,
		ApprovalRequired:  true,
		ApprovalStatus:    ApprovalStatusPending,
		ValidationStatus:  ValidationStatusPending,
		CreatedBy:         createdBy,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Create reversed entries (swap debit/credit amounts)
	for _, entry := range t.Entries {
		reversalEntry := TransactionEntry{
			ID:            uuid.New(),
			TenantID:      t.TenantID,
			TransactionID: reversal.ID,
			EntryNumber:   entry.EntryNumber,
			AccountID:     entry.AccountID,
			DebitAmount:   entry.CreditAmount, // Swap amounts
			CreditAmount:  entry.DebitAmount,  // Swap amounts
			Description:   fmt.Sprintf("Reversal: %s", entry.Description),
			Reference:     entry.Reference,
			CostCenter:    entry.CostCenter,
			Department:    entry.Department,
			ProjectID:     entry.ProjectID,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		reversal.Entries = append(reversal.Entries, reversalEntry)
	}

	reversal.CalculateTotals()
	return reversal, nil
}

// TODO: Add support for transaction templates for common transaction patterns
// TODO: Implement transaction batching with batch-level validation
// TODO: Add support for multi-currency transactions with automatic conversion
// TODO: Implement transaction attachments and supporting documents
// NOTE: Consider adding transaction series/numbering configuration per entity
// NOTE: Future enhancement: Add support for workflow routing based on transaction amount
// NOTE: Consider implementing transaction approval delegation and escalation rules
