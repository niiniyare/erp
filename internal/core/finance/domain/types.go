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

type AccountStatus string

const (
	AccountStatusDraft             AccountStatus = "DRAFT"
	AccountStatusPendingApproval   AccountStatus = "PENDING_APPROVAL"
	AccountStatusActive            AccountStatus = "ACTIVE"
	AccountStatusInactive          AccountStatus = "INACTIVE"
	AccountStatusSuspended         AccountStatus = "SUSPENDED"
	AccountStatusFrozen            AccountStatus = "FROZEN"
	AccountStatusClosed            AccountStatus = "CLOSED"
	AccountStatusArchived          AccountStatus = "ARCHIVED"
	AccountStatusUnderReview       AccountStatus = "UNDER_REVIEW"
	AccountStatusRestricted        AccountStatus = "RESTRICTED"
	AccountStatusYearEndProcessing AccountStatus = "YEAR_END_PROCESSING"
	AccountStatusAuditLock         AccountStatus = "AUDIT_LOCK"
	AccountStatusDataError         AccountStatus = "DATA_ERROR"
	AccountStatusComplianceHold    AccountStatus = "COMPLIANCE_HOLD"
	AccountStatusSystemMaintenance AccountStatus = "SYSTEM_MAINTENANCE"
)

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
	TransactionTypeManual       TransactionType = "MANUAL"     // User-created transactions
	TransactionTypeSystem       TransactionType = "SYSTEM"     // System-generated transactions
	TransactionTypeImported     TransactionType = "IMPORTED"   // Imported from external systems
	TransactionTypeRecurring    TransactionType = "RECURRING"  // Auto-generated recurring transactions
	TransactionTypeAdjustment   TransactionType = "ADJUSTMENT" // Correcting/adjusting entries
	TransactionTypeClosing      TransactionType = "CLOSING"       // Period-end closing entries
	TransactionTypeJournalEntry TransactionType = "JOURNAL_ENTRY" // Manual journal entry transactions
	TransactionTypeOpening      TransactionType = "OPENING"       // Opening balance transactions
	TransactionTypeInvoice      TransactionType = "INVOICE"
	TransactionTypePayment      TransactionType = "PAYMENT"
	TransactionTypePurchase     TransactionType = "PURCHASE"
)

// IsValid validates if the TransactionType is one of the defined constants
func (tt TransactionType) IsValid() bool {
	switch tt {
	case TransactionTypeManual, TransactionTypeSystem, TransactionTypeImported,
		TransactionTypeRecurring, TransactionTypeAdjustment, TransactionTypeClosing,
		TransactionTypeJournalEntry, TransactionTypeOpening,
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

// IsEditable returns true if the transaction can be modified in this status.
// Only DRAFT and REJECTED transactions are editable — a pending-approval transaction
// is locked until the approver acts; editing it would invalidate the in-flight request.
func (ts TransactionStatus) IsEditable() bool {
	switch ts {
	case TransactionStatusDraft, TransactionStatusRejected:
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

// ApprovalStatus represents the approval status of a transaction.
//
// Lifecycle:
//
//	NOT_REQUIRED → (posting proceeds directly)
//	PENDING → APPROVED | REJECTED | PARTIALLY_APPROVED | EXPIRED
//	PARTIALLY_APPROVED → APPROVED | REJECTED | EXPIRED
//	EXPIRED → (transaction reverts to DRAFT; submitter notified by Temporal workflow)
type ApprovalStatus string

const (
	// ApprovalStatusNotRequired — transaction does not require approval before posting.
	ApprovalStatusNotRequired ApprovalStatus = "NOT_REQUIRED"

	// ApprovalStatusPending — approval request submitted; awaiting action from the first
	// (or only) approver in the configured approval chain.
	ApprovalStatusPending ApprovalStatus = "PENDING"

	// ApprovalStatusApproved — all required approval tiers have approved.
	// The transaction may now be posted.
	ApprovalStatusApproved ApprovalStatus = "APPROVED"

	// ApprovalStatusRejected — at least one approver has rejected the transaction.
	// The transaction returns to DRAFT so the submitter can correct and resubmit.
	ApprovalStatusRejected ApprovalStatus = "REJECTED"

	// ApprovalStatusPartiallyApproved — used in sequential multi-tier approval.
	// Tier N has approved but tier N+1 has not yet acted. The transaction is still
	// locked (not editable) until the final tier approves or rejects.
	ApprovalStatusPartiallyApproved ApprovalStatus = "PARTIALLY_APPROVED"

	// ApprovalStatusExpired — the approval request was not acted on within the
	// configured SLA (default 48 hours). A Temporal timer triggers this transition;
	// the transaction reverts to DRAFT and the submitter receives a notification.
	ApprovalStatusExpired ApprovalStatus = "EXPIRED"
)

// IsValid validates if the ApprovalStatus is one of the defined constants
func (as ApprovalStatus) IsValid() bool {
	switch as {
	case ApprovalStatusNotRequired, ApprovalStatusPending, ApprovalStatusApproved,
		ApprovalStatusRejected, ApprovalStatusPartiallyApproved, ApprovalStatusExpired:
		return true
	default:
		return false
	}
}

// String returns the string representation of ApprovalStatus
func (as ApprovalStatus) String() string {
	return string(as)
}

// RejectionReason represents the reason a transaction was rejected during the approval workflow.
// These are accounting-domain reasons, not payment-gateway codes.
type RejectionReason string

const (
	RejectionReasonInvalidAccount      RejectionReason = "INVALID_ACCOUNT"       // Account does not exist, is inactive, or rejects transactions
	RejectionReasonPeriodClosed        RejectionReason = "PERIOD_CLOSED"         // Accounting period is closed; no further postings allowed
	RejectionReasonBudgetExceeded      RejectionReason = "BUDGET_EXCEEDED"       // Transaction would breach the approved budget
	RejectionReasonUnbalancedEntry     RejectionReason = "UNBALANCED_ENTRY"      // Debits ≠ credits; double-entry principle violated
	RejectionReasonMissingDocumentation RejectionReason = "MISSING_DOCUMENTATION" // Required supporting document not attached
	RejectionReasonDuplicateTransaction RejectionReason = "DUPLICATE_TRANSACTION" // Suspected duplicate of an existing transaction
	RejectionReasonAmountMismatch      RejectionReason = "AMOUNT_MISMATCH"       // Amount differs from approved purchase order / contract
	RejectionReasonUnauthorisedAccount RejectionReason = "UNAUTHORISED_ACCOUNT"  // Submitter lacks permission to post to this account
	RejectionReasonCurrencyMismatch    RejectionReason = "CURRENCY_MISMATCH"     // Transaction currency inconsistent with account currency
	RejectionReasonPolicyViolation     RejectionReason = "POLICY_VIOLATION"      // Violates a company policy (e.g. segregation of duties)
	RejectionReasonOther               RejectionReason = "OTHER"                 // Catch-all; rejector must supply a note
)

// ValidReasons contains all supported rejection reasons for validation
var ValidReasons = []RejectionReason{
	RejectionReasonInvalidAccount,
	RejectionReasonPeriodClosed,
	RejectionReasonBudgetExceeded,
	RejectionReasonUnbalancedEntry,
	RejectionReasonMissingDocumentation,
	RejectionReasonDuplicateTransaction,
	RejectionReasonAmountMismatch,
	RejectionReasonUnauthorisedAccount,
	RejectionReasonCurrencyMismatch,
	RejectionReasonPolicyViolation,
	RejectionReasonOther,
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

// ValidationSeverity classifies how serious a validation finding is.
type ValidationSeverity string

const (
	ValidationSeverityError   ValidationSeverity = "ERROR"   // Must be fixed before saving
	ValidationSeverityWarning ValidationSeverity = "WARNING" // Should be reviewed; save is allowed
	ValidationSeverityInfo    ValidationSeverity = "INFO"    // Informational only
)

// ValidationError represents a validation finding.
// Field supports nested paths such as "entries[0].amount".
type ValidationError struct {
	Field    string             `json:"field"`    // Dot/bracket path to the failing field
	Message  string             `json:"message"`  // Human-readable description
	Code     string             `json:"code"`     // Machine-readable code for i18n
	Severity ValidationSeverity `json:"severity"` // ERROR, WARNING, or INFO
}

// Error implements the error interface
func (ve ValidationError) Error() string {
	return ve.Message
}

// EffectiveSeverity returns the severity, treating the zero value as ERROR so that
// existing ValidationError literals without an explicit Severity field remain blockers.
func (ve ValidationError) EffectiveSeverity() ValidationSeverity {
	if ve.Severity == "" {
		return ValidationSeverityError
	}
	return ve.Severity
}

// IsBlocker returns true when the finding must be resolved before the record can be saved.
func (ve ValidationError) IsBlocker() bool {
	return ve.EffectiveSeverity() == ValidationSeverityError
}

// StatementSection represents financial statement sections
type StatementSection string

const (
	StatementSectionAssets        StatementSection = "ASSETS"
	StatementSectionLiabilities   StatementSection = "LIABILITIES"
	StatementSectionEquity        StatementSection = "EQUITY"
	StatementSectionRevenue       StatementSection = "REVENUE"
	StatementSectionExpenses      StatementSection = "EXPENSES"
	StatementSectionOtherIncome   StatementSection = "OTHER_INCOME"
	StatementSectionOtherExpenses StatementSection = "OTHER_EXPENSES"
)

// IsValid validates if the StatementSection is one of the defined constants
func (ss StatementSection) IsValid() bool {
	switch ss {
	case StatementSectionAssets, StatementSectionLiabilities, StatementSectionEquity,
		StatementSectionRevenue, StatementSectionExpenses, StatementSectionOtherIncome,
		StatementSectionOtherExpenses:
		return true
	default:
		return false
	}
}

// String returns the string representation of StatementSection
func (ss StatementSection) String() string {
	return string(ss)
}

// CashFlowCategory represents cash flow statement classifications
type CashFlowCategory string

const (
	CashFlowOperating CashFlowCategory = "OPERATING"
	CashFlowInvesting CashFlowCategory = "INVESTING"
	CashFlowFinancing CashFlowCategory = "FINANCING"
)

// IsValid validates if the CashFlowCategory is one of the defined constants
func (cfc CashFlowCategory) IsValid() bool {
	switch cfc {
	case CashFlowOperating, CashFlowInvesting, CashFlowFinancing:
		return true
	default:
		return false
	}
}

// String returns the string representation of CashFlowCategory
func (cfc CashFlowCategory) String() string {
	return string(cfc)
}

