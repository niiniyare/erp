package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// PaymentRunStatus tracks the lifecycle of a batch payment run.
type PaymentRunStatus string

const (
	PaymentRunStatusDraft     PaymentRunStatus = "DRAFT"      // Being constructed; invoices are being selected
	PaymentRunStatusPending   PaymentRunStatus = "PENDING"    // Ready for approval / bank submission
	PaymentRunStatusApproved  PaymentRunStatus = "APPROVED"   // Approved and queued for processing
	PaymentRunStatusProcessing PaymentRunStatus = "PROCESSING" // Submitted to bank / payment processor
	PaymentRunStatusCompleted PaymentRunStatus = "COMPLETED"  // All payments confirmed settled
	PaymentRunStatusPartial   PaymentRunStatus = "PARTIAL"    // Some payments succeeded, some failed
	PaymentRunStatusFailed    PaymentRunStatus = "FAILED"     // Entire run rejected by bank
	PaymentRunStatusCancelled PaymentRunStatus = "CANCELLED"  // Cancelled before processing
)

// IsValid returns true if the status is a recognised value.
func (s PaymentRunStatus) IsValid() bool {
	switch s {
	case PaymentRunStatusDraft, PaymentRunStatusPending, PaymentRunStatusApproved,
		PaymentRunStatusProcessing, PaymentRunStatusCompleted, PaymentRunStatusPartial,
		PaymentRunStatusFailed, PaymentRunStatusCancelled:
		return true
	default:
		return false
	}
}

// String returns the string representation.
func (s PaymentRunStatus) String() string { return string(s) }

// IsEditable returns true when invoice selection may still be modified.
func (s PaymentRunStatus) IsEditable() bool {
	return s == PaymentRunStatusDraft
}

// IsTerminal returns true when the run has reached a final state.
func (s PaymentRunStatus) IsTerminal() bool {
	switch s {
	case PaymentRunStatusCompleted, PaymentRunStatusFailed, PaymentRunStatusCancelled:
		return true
	default:
		return false
	}
}

// PaymentMethod classifies how the funds are transferred.
type PaymentMethod string

const (
	PaymentMethodEFT     PaymentMethod = "EFT"      // Electronic Funds Transfer
	PaymentMethodCheque  PaymentMethod = "CHEQUE"   // Physical cheque
	PaymentMethodCash    PaymentMethod = "CASH"      // Cash payment
	PaymentMethodBECS    PaymentMethod = "BECS"      // Bulk Electronic Clearing System (AU)
	PaymentMethodSEPA    PaymentMethod = "SEPA"      // Single Euro Payments Area
	PaymentMethodSWIFT   PaymentMethod = "SWIFT"     // International wire
	PaymentMethodInternal PaymentMethod = "INTERNAL" // Internal transfer between own accounts
)

// IsValid returns true if the PaymentMethod is a recognised value.
func (m PaymentMethod) IsValid() bool {
	switch m {
	case PaymentMethodEFT, PaymentMethodCheque, PaymentMethodCash,
		PaymentMethodBECS, PaymentMethodSEPA, PaymentMethodSWIFT, PaymentMethodInternal:
		return true
	default:
		return false
	}
}

// String returns the string representation.
func (m PaymentMethod) String() string { return string(m) }

// PaymentRun is the header record for a batch payment.
// It groups individual PaymentRunItems and drives the approval + GL posting flow.
type PaymentRun struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`

	// Identity
	RunNumber   string `json:"run_number"`   // System-generated sequential number
	Description string `json:"description"`  // Human-readable label
	Reference   string `json:"reference"`    // Bank payment reference (appears on statements)

	// Funding account (the GL account funds are drawn from)
	BankAccountID uuid.UUID `json:"bank_account_id"`
	CurrencyCode  string    `json:"currency_code"` // ISO 4217

	PaymentDate   time.Time     `json:"payment_date"`   // Value date for all items in this run
	PaymentMethod PaymentMethod `json:"payment_method"`

	// Aggregated amounts (maintained by the service layer)
	TotalAmount    decimal.Decimal `json:"total_amount"`    // Sum of all item amounts
	ItemCount      int             `json:"item_count"`      // Number of items in the run
	FailedCount    int             `json:"failed_count"`    // Items that failed processing

	Status PaymentRunStatus `json:"status"`

	// Approval metadata
	SubmittedAt *time.Time `json:"submitted_at,omitempty"`
	SubmittedBy *uuid.UUID `json:"submitted_by,omitempty"`
	ApprovedAt  *time.Time `json:"approved_at,omitempty"`
	ApprovedBy  *uuid.UUID `json:"approved_by,omitempty"`

	// Processing metadata
	ProcessedAt    *time.Time `json:"processed_at,omitempty"`
	BankReference  *string    `json:"bank_reference,omitempty"`  // Reference returned by the bank
	BankResponseAt *time.Time `json:"bank_response_at,omitempty"`

	// GL posting reference (created once the run is confirmed)
	TransactionID *uuid.UUID `json:"transaction_id,omitempty"` // The GL transaction that records the payment

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	CreatedBy uuid.UUID  `json:"created_by"`
	UpdatedBy *uuid.UUID `json:"updated_by,omitempty"`
}

// Validate returns ValidationErrors for the PaymentRun header.
func (r *PaymentRun) Validate() []ValidationError {
	var errs []ValidationError

	if r.TenantID == uuid.Nil {
		errs = append(errs, ValidationError{Field: "tenant_id", Message: "tenant_id is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	if r.BankAccountID == uuid.Nil {
		errs = append(errs, ValidationError{Field: "bank_account_id", Message: "bank_account_id is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	if strings.TrimSpace(r.RunNumber) == "" {
		errs = append(errs, ValidationError{Field: "run_number", Message: "run_number is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	if len(r.CurrencyCode) != 3 {
		errs = append(errs, ValidationError{Field: "currency_code", Message: "currency_code must be a 3-character ISO 4217 code", Code: "INVALID_FORMAT", Severity: ValidationSeverityError})
	}
	if r.PaymentDate.IsZero() {
		errs = append(errs, ValidationError{Field: "payment_date", Message: "payment_date is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	if !r.PaymentMethod.IsValid() {
		errs = append(errs, ValidationError{Field: "payment_method", Message: fmt.Sprintf("invalid payment method: %s", r.PaymentMethod), Code: "INVALID_VALUE", Severity: ValidationSeverityError})
	}
	if !r.Status.IsValid() {
		errs = append(errs, ValidationError{Field: "status", Message: fmt.Sprintf("invalid status: %s", r.Status), Code: "INVALID_VALUE", Severity: ValidationSeverityError})
	}
	if r.TotalAmount.LessThan(decimal.Zero) {
		errs = append(errs, ValidationError{Field: "total_amount", Message: "total_amount must be ≥ 0", Code: "VALUE_OUT_OF_RANGE", Severity: ValidationSeverityError})
	}

	return errs
}

// PaymentRunItemStatus tracks the outcome of a single item.
type PaymentRunItemStatus string

const (
	PaymentRunItemStatusPending   PaymentRunItemStatus = "PENDING"   // Not yet submitted
	PaymentRunItemStatusSubmitted PaymentRunItemStatus = "SUBMITTED" // Sent to bank
	PaymentRunItemStatusCleared   PaymentRunItemStatus = "CLEARED"   // Confirmed settled by bank
	PaymentRunItemStatusFailed    PaymentRunItemStatus = "FAILED"    // Rejected by bank
	PaymentRunItemStatusReversed  PaymentRunItemStatus = "REVERSED"  // Reversed after settlement
)

// IsValid returns true if the status is a recognised value.
func (s PaymentRunItemStatus) IsValid() bool {
	switch s {
	case PaymentRunItemStatusPending, PaymentRunItemStatusSubmitted,
		PaymentRunItemStatusCleared, PaymentRunItemStatusFailed, PaymentRunItemStatusReversed:
		return true
	default:
		return false
	}
}

// PaymentRunItem is a single payable in a PaymentRun.
// It typically corresponds to one supplier invoice or a group of invoices for the same payee.
type PaymentRunItem struct {
	ID           uuid.UUID `json:"id"`
	PaymentRunID uuid.UUID `json:"payment_run_id"`
	TenantID     uuid.UUID `json:"tenant_id"`

	// Payee (supplier / employee)
	PayeeID   uuid.UUID `json:"payee_id"`   // Tenant user / contact ID
	PayeeName string    `json:"payee_name"` // Denormalised for display

	// Source document(s)
	SourceDocumentType string    `json:"source_document_type"` // e.g. "INVOICE", "EXPENSE"
	SourceDocumentID   uuid.UUID `json:"source_document_id"`
	SourceReference    string    `json:"source_reference"` // Invoice number / expense claim ref

	// Amounts
	GrossAmount   decimal.Decimal `json:"gross_amount"`   // Total amount on source document
	DiscountAmount decimal.Decimal `json:"discount_amount"` // Early-payment discount applied
	NetAmount     decimal.Decimal `json:"net_amount"`      // Amount actually paid (gross - discount)

	// Bank details snapshot (copied at run creation to protect against changes)
	BankAccountNumber *string `json:"bank_account_number,omitempty"`
	BankSortCode      *string `json:"bank_sort_code,omitempty"`
	BankIBAN          *string `json:"bank_iban,omitempty"`
	BankSWIFT         *string `json:"bank_swift,omitempty"`

	Status PaymentRunItemStatus `json:"status"`

	// Processing outcome
	FailureReason  *string    `json:"failure_reason,omitempty"`
	ClearedAt      *time.Time `json:"cleared_at,omitempty"`
	BankReference  *string    `json:"bank_reference,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate returns ValidationErrors for the PaymentRunItem.
func (i *PaymentRunItem) Validate() []ValidationError {
	var errs []ValidationError

	if i.PaymentRunID == uuid.Nil {
		errs = append(errs, ValidationError{Field: "payment_run_id", Message: "payment_run_id is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	if i.PayeeID == uuid.Nil {
		errs = append(errs, ValidationError{Field: "payee_id", Message: "payee_id is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	if i.SourceDocumentID == uuid.Nil {
		errs = append(errs, ValidationError{Field: "source_document_id", Message: "source_document_id is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	if i.NetAmount.LessThanOrEqual(decimal.Zero) {
		errs = append(errs, ValidationError{Field: "net_amount", Message: "net_amount must be greater than zero", Code: "VALUE_OUT_OF_RANGE", Severity: ValidationSeverityError})
	}
	if i.DiscountAmount.LessThan(decimal.Zero) {
		errs = append(errs, ValidationError{Field: "discount_amount", Message: "discount_amount must be ≥ 0", Code: "VALUE_OUT_OF_RANGE", Severity: ValidationSeverityError})
	}
	if expected := i.GrossAmount.Sub(i.DiscountAmount); !expected.Equal(i.NetAmount) {
		errs = append(errs, ValidationError{
			Field:    "net_amount",
			Message:  fmt.Sprintf("net_amount (%s) must equal gross_amount (%s) minus discount_amount (%s)", i.NetAmount, i.GrossAmount, i.DiscountAmount),
			Code:     "AMOUNT_MISMATCH",
			Severity: ValidationSeverityError,
		})
	}
	if !i.Status.IsValid() {
		errs = append(errs, ValidationError{Field: "status", Message: fmt.Sprintf("invalid status: %s", i.Status), Code: "INVALID_VALUE", Severity: ValidationSeverityError})
	}

	return errs
}
