// Package finance contains the domain models and business logic for the finance module
package domain

import (
	"fmt"
	"slices"
)

// RootType represents the high-level account classification following
// standard accounting principles (Assets = Liabilities + Equity)
type RootType string

const (
	RootTypeAsset     RootType = "ASSET"     // Economic resources owned by the entity
	RootTypeLiability RootType = "LIABILITY" // Obligations owed to external parties
	RootTypeEquity    RootType = "EQUITY"    // Owner's residual interest in assets
	RootTypeRevenue   RootType = "REVENUE"   // Income from business operations
	RootTypeExpense   RootType = "EXPENSE"   // Costs incurred to generate revenue
)

// IsValid validates if the RootType is one of the defined constants
func (rt RootType) IsValid() bool {
	switch rt {
	case RootTypeAsset, RootTypeLiability, RootTypeEquity, RootTypeRevenue, RootTypeExpense:
		return true
	default:
		return false
	}
}

// String returns the string representation of RootType
func (rt RootType) String() string {
	return string(rt)
}

// ParseRootType parses a string into a RootType and validates it
func ParseRootType(s string) (RootType, error) {
	rt := RootType(s)
	if !rt.IsValid() {
		return "", fmt.Errorf("invalid root type: %s", s)
	}
	return rt, nil
}

// NormalBalance represents the normal balance type for accounts
// This determines which side (debit/credit) increases the account balance
type NormalBalance string

const (
	NormalBalanceDebit  NormalBalance = "DEBIT"  // Left side increases balance
	NormalBalanceCredit NormalBalance = "CREDIT" // Right side increases balance
)

// IsValid validates if the NormalBalance is one of the defined constants
func (nb NormalBalance) IsValid() bool {
	switch nb {
	case NormalBalanceDebit, NormalBalanceCredit:
		return true
	default:
		return false
	}
}

// String returns the string representation of NormalBalance
func (nb NormalBalance) String() string {
	return string(nb)
}

// ParseNormalBalance parses a string into a NormalBalance and validates it
func ParseNormalBalance(s string) (NormalBalance, error) {
	nb := NormalBalance(s)
	if !nb.IsValid() {
		return "", fmt.Errorf("invalid normal balance: %s", s)
	}
	return nb, nil
}

// TransactionType represents the type of financial transaction
// This helps categorize transactions by their source and purpose
type TransactionType string

const (
	TransactionTypeManual        TransactionType = "MANUAL"         // User-created transactions
	TransactionTypeSystem        TransactionType = "SYSTEM"         // System-generated transactions
	TransactionTypeImported      TransactionType = "IMPORTED"       // Imported from external systems
	TransactionTypeRecurring     TransactionType = "RECURRING"      // Auto-generated recurring transactions
	TransactionTypeAdjustment    TransactionType = "ADJUSTMENT"     // Correcting/adjusting entries
	TransactionTypeClosing       TransactionType = "CLOSING"        // Period-end closing entries
	TransactionTypeJournal       TransactionType = "JOURNAL"
	TransactionTypeJournalEntry  TransactionType = "JOURNAL_ENTRY"  // Journal entry transactions
	TransactionTypeOpening       TransactionType = "OPENING"        // Opening balance transactions
	TransactionTypeInvoice       TransactionType = "INVOICE"
	TransactionTypePayment       TransactionType = "PAYMENT"
	TransactionTypePurchase      TransactionType = "PURCHASE"
)

// IsValid validates if the TransactionType is one of the defined constants
func (tt TransactionType) IsValid() bool {
	switch tt {
	case TransactionTypeManual, TransactionTypeSystem, TransactionTypeImported,
		TransactionTypeRecurring, TransactionTypeAdjustment, TransactionTypeClosing,
		TransactionTypeJournal, TransactionTypeJournalEntry, TransactionTypeOpening,
		TransactionTypeInvoice, TransactionTypePayment, TransactionTypePurchase:
		return true
	default:
		return false
	}
}

// RequiresApproval returns true if this transaction type typically requires approval
func (tt TransactionType) RequiresApproval() bool {
	switch tt {
	case TransactionTypeManual, TransactionTypeAdjustment, TransactionTypeClosing:
		return true
	default:
		return false
	}
}

// String returns the string representation of TransactionType
func (tt TransactionType) String() string {
	return string(tt)
}

// ParseTransactionType converts a string to a valid TransactionType
// Returns an error if the string is not a valid transaction type
func ParseTransactionType(s string) (TransactionType, error) {
	tt := TransactionType(s)
	if !tt.IsValid() {
		return "", fmt.Errorf("invalid transaction type: %s", s)
	}
	return tt, nil
}

// TransactionStatus represents the current status of a transaction
// This tracks the transaction through its lifecycle
type TransactionStatus string

const (
	TransactionStatusDraft           TransactionStatus = "DRAFT"            // Initial state, can be edited
	TransactionStatusPendingApproval TransactionStatus = "PENDING_APPROVAL" // Waiting for approval
	TransactionStatusApproved        TransactionStatus = "APPROVED"         // Approved but not posted
	TransactionStatusPosted          TransactionStatus = "POSTED"           // Posted to ledger, affects balances
	TransactionStatusCancelled       TransactionStatus = "CANCELLED"        // Cancelled before posting
	TransactionStatusReversed        TransactionStatus = "REVERSED"         // Posted but later reversed
	TransactionStatusRejected        TransactionStatus = "REJECTED"         // Rejected during approval
)

// IsValid validates if the TransactionStatus is one of the defined constants
func (ts TransactionStatus) IsValid() bool {
	switch ts {
	case TransactionStatusDraft, TransactionStatusPendingApproval, TransactionStatusApproved,
		TransactionStatusPosted, TransactionStatusCancelled, TransactionStatusReversed, TransactionStatusRejected:
		return true
	default:
		return false
	}
}

// IsEditable returns true if the transaction can be modified in this status
func (ts TransactionStatus) IsEditable() bool {
	switch ts {
	case TransactionStatusDraft, TransactionStatusPendingApproval:
		return true
	default:
		return false
	}
}

// IsActive returns true if the transaction affects account balances
func (ts TransactionStatus) IsActive() bool {
	return ts == TransactionStatusPosted
}

// String returns the string representation of TransactionStatus
func (ts TransactionStatus) String() string {
	return string(ts)
}

// ParseTransactionStatus converts a string to a valid TransactionStatus
// Returns an error if the string is not a valid transaction status
func ParseTransactionStatus(s string) (TransactionStatus, error) {
	ts := TransactionStatus(s)
	if !ts.IsValid() {
		return "", fmt.Errorf("invalid transaction status: %s", s)
	}
	return ts, nil
}

// ApprovalStatus represents the approval status of a transaction
type ApprovalStatus string

const (
	ApprovalStatusNotRequired        ApprovalStatus = "NOT_REQUIRED"        // No approval needed
	ApprovalStatusPending            ApprovalStatus = "PENDING"             // Waiting for approval
	ApprovalStatusApproved           ApprovalStatus = "APPROVED"            // Approved by authorized user
	ApprovalStatusRejected           ApprovalStatus = "REJECTED"            // Rejected by approver
	ApprovalStatusPartiallyApproved  ApprovalStatus = "PARTIALLY_APPROVED"  // Partially approved (multi-level approval)
)

// IsValid validates if the ApprovalStatus is one of the defined constants
func (as ApprovalStatus) IsValid() bool {
	switch as {
	case ApprovalStatusNotRequired, ApprovalStatusPending, ApprovalStatusApproved, 
		ApprovalStatusRejected, ApprovalStatusPartiallyApproved:
		return true
	default:
		return false
	}
}

// String returns the string representation of ApprovalStatus
func (as ApprovalStatus) String() string {
	return string(as)
}

// RejectionReason represents the reason for transaction rejection
type RejectionReason string

// Enumeration of supported rejection reasons
const (
	InsufficientFunds     RejectionReason = "INSUFFICIENT_FUNDS"
	InvalidAccount        RejectionReason = "INVALID_ACCOUNT"
	TransactionLimit      RejectionReason = "TRANSACTION_LIMIT_EXCEEDED"
	DailyLimit            RejectionReason = "DAILY_LIMIT_EXCEEDED"
	FraudSuspected        RejectionReason = "FRAUD_SUSPECTED"
	AuthorizationRequired RejectionReason = "AUTHORIZATION_REQUIRED"
	ExpiredCard           RejectionReason = "EXPIRED_CARD"
	InvalidMerchant       RejectionReason = "INVALID_MERCHANT"
	TechnicalError        RejectionReason = "TECHNICAL_ERROR"
	InvalidAmount         RejectionReason = "INVALID_AMOUNT"
	DuplicateTransaction  RejectionReason = "DUPLICATE_TRANSACTION"
)

// ValidReasons contains all supported rejection reasons for validation
var ValidReasons = []RejectionReason{
	InsufficientFunds,
	InvalidAccount,
	TransactionLimit,
	DailyLimit,
	FraudSuspected,
	AuthorizationRequired,
	ExpiredCard,
	InvalidMerchant,
	TechnicalError,
	InvalidAmount,
	DuplicateTransaction,
}

// IsValid checks if the rejection reason is supported
func (r RejectionReason) IsValid() bool {
	return slices.Contains(ValidReasons, r)
}

// String returns the string representation of the rejection reason
func (r RejectionReason) String() string {
	return string(r)
}

// ParseRejectionReason converts a string to a valid RejectionReason
// Returns an error if the string is not a valid reason
func (r RejectionReason) ParseRejectionReason(s string) (RejectionReason, error) {
	rr := RejectionReason(s)
	if !rr.IsValid() {
		return "", fmt.Errorf("invalid rejection reason: %s", s)
	}
	return rr, nil
}

// ValidationStatus represents the validation status of an entity
type ValidationStatus string

const (
	ValidationStatusPending ValidationStatus = "PENDING" // Validation not yet performed
	ValidationStatusValid   ValidationStatus = "VALID"   // Passed all validations
	ValidationStatusWarning ValidationStatus = "WARNING" // Has warnings but valid
	ValidationStatusError   ValidationStatus = "ERROR"   // Has validation errors
)

// IsValid validates if the ValidationStatus is one of the defined constants
func (vs ValidationStatus) IsValid() bool {
	switch vs {
	case ValidationStatusPending, ValidationStatusValid, ValidationStatusWarning, ValidationStatusError:
		return true
	default:
		return false
	}
}

// HasErrors returns true if the validation status indicates errors
func (vs ValidationStatus) HasErrors() bool {
	return vs == ValidationStatusError
}

// String returns the string representation of ValidationStatus
func (vs ValidationStatus) String() string {
	return string(vs)
}

// GetNormalBalanceForRootType returns the normal balance for a root type
// This follows standard accounting principles:
// - Assets and Expenses have debit normal balance
// - Liabilities, Equity, and Revenue have credit normal balance
func GetNormalBalanceForRootType(rootType RootType) NormalBalance {
	switch rootType {
	case RootTypeAsset, RootTypeExpense:
		return NormalBalanceDebit
	case RootTypeLiability, RootTypeEquity, RootTypeRevenue:
		return NormalBalanceCredit
	default:
		// NOTE: Default to debit for unknown types, but this should be validated
		return NormalBalanceDebit
	}
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`   // The field that failed validation
	Message string `json:"message"` // Human-readable error message
	Code    string `json:"code"`    // Machine-readable error code for i18n
}

// Error implements the error interface
func (ve ValidationError) Error() string {
	return ve.Message
}

// TODO: Add severity levels to ValidationError (ERROR, WARNING, INFO)
// TODO: Add support for nested field paths (e.g., "entries[0].amount")
