package domain

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// =============================================================================
// Field-length constants
// =============================================================================

const (
	AccountCodeMaxLen        = 20
	AccountNameMaxLen        = 200
	AccountDescriptionMaxLen = 1000
	AccountTypeMaxLen        = 100
	AccountSubtypeMaxLen     = 100
	CurrencyCodeLen          = 3 // ISO 4217
)

// accountCodePattern restricts codes to alphanumeric characters and hyphens,
// e.g. "1000", "REV-FUEL-PMS", "AR-FLEET-001".
var accountCodePattern = regexp.MustCompile(`^[A-Za-z0-9\-]+$`)

// =============================================================================
// AccountStatus & state machine
// =============================================================================

// AccountStatus represents the full operational lifecycle of a ledger account.
//
// State machine overview:
//
//	DRAFT ──► PENDING_APPROVAL ──► ACTIVE ──► INACTIVE ──► CLOSED ──► ARCHIVED
//	                              │    ▲
//	                    holds, freezes, reviews (all reversible except CLOSED/ARCHIVED)
//
// See AllowedTransitions for the complete, authoritative transition table.
type AccountStatus string

const (
	// Draft — account is being configured; not yet visible to posting workflows.
	AccountStatusDraft AccountStatus = "DRAFT"

	// PendingApproval — submitted for review; awaiting authorisation.
	AccountStatusPendingApproval AccountStatus = "PENDING_APPROVAL"

	// Active — fully operational; accepts transactions subject to per-account settings.
	AccountStatusActive AccountStatus = "ACTIVE"

	// Inactive — voluntarily suspended; no new postings; zero-balance required.
	AccountStatusInactive AccountStatus = "INACTIVE"

	// Suspended — administrative hold; temporary, expects resolution.
	AccountStatusSuspended AccountStatus = "SUSPENDED"

	// Restricted — limited to approval-only or read-only entries.
	AccountStatusRestricted AccountStatus = "RESTRICTED"

	// Frozen — hard freeze; no transactions of any kind.
	AccountStatusFrozen AccountStatus = "FROZEN"

	// UnderReview — flagged for internal audit or compliance review.
	AccountStatusUnderReview AccountStatus = "UNDER_REVIEW"

	// YearEndProcessing — locked for period close; only closing entries permitted.
	AccountStatusYearEndProcessing AccountStatus = "YEAR_END_PROCESSING"

	// AuditLock — external audit in progress; read-only for the duration.
	AccountStatusAuditLock AccountStatus = "AUDIT_LOCK"

	// ComplianceHold — regulatory or sanctions hold; no transactions.
	AccountStatusComplianceHold AccountStatus = "COMPLIANCE_HOLD"

	// SystemMaintenance — brief technical window (e.g. data migration).
	AccountStatusSystemMaintenance AccountStatus = "SYSTEM_MAINTENANCE"

	// DataError — automated detection of data integrity issue; pending correction.
	AccountStatusDataError AccountStatus = "DATA_ERROR"

	// Closed — terminal operational state; balance must be zero. Retained for history.
	AccountStatusClosed AccountStatus = "CLOSED"

	// Archived — moved to long-term storage after closure. Read-only forever.
	AccountStatusArchived AccountStatus = "ARCHIVED"
)

// IsValid returns true if s is a recognised status constant.
func (s AccountStatus) IsValid() bool {
	switch s {
	case AccountStatusDraft, AccountStatusPendingApproval, AccountStatusActive,
		AccountStatusInactive, AccountStatusSuspended, AccountStatusRestricted,
		AccountStatusFrozen, AccountStatusUnderReview, AccountStatusYearEndProcessing,
		AccountStatusAuditLock, AccountStatusComplianceHold, AccountStatusSystemMaintenance,
		AccountStatusDataError, AccountStatusClosed, AccountStatusArchived:
		return true
	default:
		return false
	}
}

// String implements fmt.Stringer.
func (s AccountStatus) String() string { return string(s) }

// StatusTransition encodes a single permitted move between two account statuses,
// together with the permission string that the service layer must verify and a
// flag indicating whether a full validation run is required before the move.
type StatusTransition struct {
	From AccountStatus
	To   AccountStatus

	// RequiredPermission is the permission key the caller must hold.
	// e.g. "finance.accounts.approve"
	RequiredPermission string

	// ValidationRequired means Accounts.Validate() must return no errors
	// before this transition may proceed.
	ValidationRequired bool
}

// AllowedTransitions is the authoritative state machine for account status.
// All service-layer transition logic must consult this table; ad-hoc if-chains
// are forbidden. Adding a new transition requires a single row here.
var AllowedTransitions = []StatusTransition{
	// ── Initial lifecycle ────────────────────────────────────────────────────
	{AccountStatusDraft, AccountStatusPendingApproval, "finance.accounts.submit", true},
	{AccountStatusDraft, AccountStatusActive, "finance.accounts.activate", true},

	// ── Approval flow ────────────────────────────────────────────────────────
	{AccountStatusPendingApproval, AccountStatusActive, "finance.accounts.approve", true},
	{AccountStatusPendingApproval, AccountStatusDraft, "finance.accounts.reject", false},

	// ── Normal operational transitions ───────────────────────────────────────
	{AccountStatusActive, AccountStatusInactive, "finance.accounts.deactivate", false},
	{AccountStatusActive, AccountStatusSuspended, "finance.accounts.suspend", false},
	{AccountStatusActive, AccountStatusRestricted, "finance.accounts.restrict", false},
	{AccountStatusActive, AccountStatusFrozen, "finance.accounts.freeze", false},
	{AccountStatusActive, AccountStatusUnderReview, "finance.accounts.review", false},
	{AccountStatusActive, AccountStatusYearEndProcessing, "finance.accounts.year_end", true},
	{AccountStatusActive, AccountStatusAuditLock, "finance.accounts.audit_lock", false},
	{AccountStatusActive, AccountStatusComplianceHold, "finance.accounts.compliance_hold", false},
	{AccountStatusActive, AccountStatusSystemMaintenance, "finance.accounts.maintenance", false},
	{AccountStatusActive, AccountStatusDataError, "finance.accounts.flag_error", false},

	// ── Releasing holds ───────────────────────────────────────────────────────
	{AccountStatusSuspended, AccountStatusActive, "finance.accounts.activate", true},
	{AccountStatusSuspended, AccountStatusInactive, "finance.accounts.deactivate", false},
	{AccountStatusRestricted, AccountStatusActive, "finance.accounts.activate", true},
	{AccountStatusRestricted, AccountStatusFrozen, "finance.accounts.freeze", false},
	{AccountStatusFrozen, AccountStatusActive, "finance.accounts.activate", true},
	{AccountStatusFrozen, AccountStatusRestricted, "finance.accounts.restrict", false},
	{AccountStatusUnderReview, AccountStatusActive, "finance.accounts.activate", true},
	{AccountStatusUnderReview, AccountStatusRestricted, "finance.accounts.restrict", false},
	{AccountStatusUnderReview, AccountStatusFrozen, "finance.accounts.freeze", false},
	{AccountStatusYearEndProcessing, AccountStatusActive, "finance.accounts.activate", true},
	{AccountStatusAuditLock, AccountStatusActive, "finance.accounts.activate", true},
	{AccountStatusComplianceHold, AccountStatusActive, "finance.accounts.activate", true},
	{AccountStatusSystemMaintenance, AccountStatusActive, "finance.accounts.activate", false},
	{AccountStatusDataError, AccountStatusActive, "finance.accounts.activate", true},
	{AccountStatusDataError, AccountStatusDraft, "finance.accounts.reset", true},

	// ── Terminal transitions ──────────────────────────────────────────────────
	{AccountStatusInactive, AccountStatusActive, "finance.accounts.activate", true},
	{AccountStatusInactive, AccountStatusSuspended, "finance.accounts.suspend", false},
	{AccountStatusInactive, AccountStatusClosed, "finance.accounts.close", true},
	{AccountStatusClosed, AccountStatusArchived, "finance.accounts.archive", false},
}

// =============================================================================
// Supporting enumerations
// =============================================================================

// RootType is the primary accounting classification of an account.
// It determines the normal balance, how the account appears on financial
// statements, and which rollup rules apply.
type RootType string

const (
	RootTypeAsset     RootType = "ASSET"
	RootTypeLiability RootType = "LIABILITY"
	RootTypeEquity    RootType = "EQUITY"
	RootTypeRevenue   RootType = "REVENUE"
	RootTypeExpense   RootType = "EXPENSE"
)

// IsValid returns true if r is a recognised root type.
func (r RootType) IsValid() bool {
	switch r {
	case RootTypeAsset, RootTypeLiability, RootTypeEquity,
		RootTypeRevenue, RootTypeExpense:
		return true
	default:
		return false
	}
}

// NormalBalance indicates which side of the double-entry equation increases
// this account's balance.
type NormalBalance string

const (
	NormalBalanceDebit  NormalBalance = "DEBIT"
	NormalBalanceCredit NormalBalance = "CREDIT"
)

// IsValid returns true if n is a recognised normal balance.
func (n NormalBalance) IsValid() bool {
	return n == NormalBalanceDebit || n == NormalBalanceCredit
}

// GetNormalBalanceForRootType returns the conventional normal balance for a
// given root type. Assets and Expenses are debit-normal; all others are
// credit-normal.
func GetNormalBalanceForRootType(rt RootType) NormalBalance {
	switch rt {
	case RootTypeAsset, RootTypeExpense:
		return NormalBalanceDebit
	default:
		return NormalBalanceCredit
	}
}

// ValidationStatus tracks whether the account has passed its most recent
// domain validation run.
type ValidationStatus string

const (
	ValidationStatusValid   ValidationStatus = "VALID"
	ValidationStatusWarning ValidationStatus = "WARNING"
	ValidationStatusInvalid ValidationStatus = "INVALID"
	ValidationStatusPending ValidationStatus = "PENDING"
)

// =============================================================================
// Account
// =============================================================================

// Account is a single node in the chart of accounts (ledger tree).
//
// # Ledger hierarchy
//
// Accounts form a tree via ParentAccountID. The hierarchy is purely structural
// for balance roll-up and posting control. RootType must be homogeneous within
// any subtree — a child must share its parent's RootType.
//
// Only leaf accounts (IsLeaf == true, i.e. HasChildren == false) may accept
// direct journal postings. Parent accounts aggregate their children's balances;
// posting to them directly is a business rule violation.
//
// Path is a MaterialisedPath maintained by the repository on create/reparent.
// Reparenting an account requires re-pathing all its descendants.
//
// # Reporting placement
//
// An account's position on the P&L, balance sheet, or cash flow statement is
// NOT encoded here. That mapping lives in AccountMapping and is owned by
// ReportingGroup. This separation means EPRA, KRA, and internal report
// structures can each have their own mapping without touching ledger accounts.
//
// # Reconciliation & tax
//
// RequiresReconciliation, LastReconciledAt, TaxCode, and TaxRate are first-class
// fields here because they are properties of the ledger account itself, not of
// how it appears in a report.
type Accounts struct {
	ID       uuid.UUID  `json:"id"`
	TenantID uuid.UUID  `json:"tenant_id"`
	EntityID *uuid.UUID `json:"entity_id,omitempty"` // Multi-entity support

	// -------------------------------------------------------------------------
	// Identity
	// -------------------------------------------------------------------------

	// AccountCode is a unique mnemonic within the tenant, e.g. "1000", "REV-FUEL-PMS".
	// Format: alphanumeric and hyphens only. Maximum AccountCodeMaxLen characters.
	// Immutable after creation; code changes require a dedicated service operation.
	AccountCode string `json:"account_code"`

	// AccountName is the display label used in reports and selectors.
	// Maximum AccountNameMaxLen characters.
	AccountName string `json:"account_name"`

	// AccountDescription is an optional long-form explanation of the account's
	// purpose, posting rules, or regulatory reference.
	AccountDescription *string `json:"account_description,omitempty"`

	// -------------------------------------------------------------------------
	// Ledger hierarchy
	// -------------------------------------------------------------------------

	// ParentAccountID is the direct parent in the ledger tree.
	// Nil for root accounts (AccountLevel == 0).
	ParentAccountID *uuid.UUID `json:"parent_account_id,omitempty"`

	// AccountLevel is the zero-based depth (0 = root).
	// Invariant: ParentAccountID == nil ↔ AccountLevel == 0.
	AccountLevel int32 `json:"account_level"`

	// Path is the materialised ancestor path maintained by the repository.
	// Use IsDescendantOf for subtree membership tests.
	Path MaterialisedPath `json:"path"`

	// HasChildren is denormalised onto the entity so that CanAcceptTransactions
	// can enforce the leaf-only posting rule without a database round-trip.
	// The repository layer is responsible for keeping this consistent.
	HasChildren bool `json:"has_children"`

	// IsLeaf mirrors !HasChildren and is stored explicitly for clarity in
	// query filters. Always == !HasChildren.
	IsLeaf bool `json:"is_leaf"`

	// -------------------------------------------------------------------------
	// Grouping references
	// -------------------------------------------------------------------------

	// AccountGroupID optionally links this account to an AccountGroup in the
	// legacy grouping scheme. Prefer AccountMapping for new report structures.
	AccountGroupID *uuid.UUID `json:"account_group_id,omitempty"`

	// ControlAccountID references the control account that summarises this
	// account (e.g. an AR control account for a customer sub-ledger account).
	// An account that IS the control account has this nil and IsControlAccount true.
	ControlAccountID *uuid.UUID `json:"control_account_id,omitempty"`

	// -------------------------------------------------------------------------
	// Classification
	// -------------------------------------------------------------------------

	// RootType is the primary classification. It is immutable after the first
	// journal entry is posted; changes require a reclassification workflow.
	RootType RootType `json:"root_type"`

	// AccountType is the secondary classification, e.g. "Current Asset",
	// "Operating Revenue". Maximum AccountTypeMaxLen characters.
	AccountType string `json:"account_type"`

	// AccountSubtype and AccountCategory provide tertiary and quaternary
	// granularity used by industry-specific reports (e.g. EPRA fuel grades).
	AccountSubtype  *string `json:"account_subtype,omitempty"`
	AccountCategory *string `json:"account_category,omitempty"`
	SubCategory     *string `json:"sub_category,omitempty"`

	// -------------------------------------------------------------------------
	// Financial attributes
	// -------------------------------------------------------------------------

	// NormalBalance determines how journal entries increase or decrease this
	// account. Must match GetNormalBalanceForRootType(RootType).
	NormalBalance NormalBalance `json:"normal_balance"`

	// IsControlAccount marks this as a summary account for a sub-ledger
	// (e.g. Accounts Receivable Control). Control accounts must not accept
	// manual entries; postings flow from the sub-ledger automatically.
	IsControlAccount bool `json:"is_control_account"`

	// -------------------------------------------------------------------------
	// Currency
	// -------------------------------------------------------------------------

	// CurrencyCode is the ISO 4217 three-letter code for the account's
	// functional currency. Nil means the tenant's default currency applies.
	CurrencyCode *string `json:"currency_code,omitempty"`

	// IsMultiCurrency allows transactions in currencies other than CurrencyCode.
	// Requires a revaluation account to be configured for FX gains/losses.
	IsMultiCurrency bool `json:"is_multi_currency"`

	// CurrencyRevaluationRequired means the account balance must be revalued
	// at each period-end using the closing exchange rate. Applies to monetary
	// foreign-currency accounts (receivables, payables, bank accounts).
	CurrencyRevaluationRequired bool `json:"currency_revaluation_required"`

	// -------------------------------------------------------------------------
	// Operational settings
	// -------------------------------------------------------------------------

	// IsActive is the fast-path operational toggle. Both Status == ACTIVE and
	// IsActive == true must hold for the account to accept transactions.
	IsActive bool `json:"is_active"`

	// IsSystemAccount marks a protected account provisioned by the chart of
	// accounts seed. System accounts cannot be renamed, reclassified, or deleted.
	IsSystemAccount bool `json:"is_system_account"`

	// AllowManualEntries controls whether human-authored journal lines may
	// reference this account. Set false for control accounts and automated
	// clearing accounts.
	AllowManualEntries bool `json:"allow_manual_entries"`

	// RequireReference mandates a non-empty reference string on every journal
	// line that posts to this account. Useful for bank reconciliation accounts.
	RequireReference bool `json:"require_reference"`

	// -------------------------------------------------------------------------
	// Reconciliation
	// -------------------------------------------------------------------------

	// RequiresReconciliation flags accounts (typically bank and cash) that must
	// be reconciled against external statements each period.
	RequiresReconciliation bool `json:"requires_reconciliation"`

	// LastReconciledAt records when the most recent reconciliation was completed.
	LastReconciledAt *time.Time `json:"last_reconciled_at,omitempty"`

	// -------------------------------------------------------------------------
	// Tax
	// -------------------------------------------------------------------------

	// TaxCode is the applicable tax code for transactions on this account
	// (e.g. "VAT16", "WHT5"). Nil means no tax applies by default.
	TaxCode *string `json:"tax_code,omitempty"`

	// TaxRate is the default tax rate percentage (0–100) associated with TaxCode.
	// Nil when TaxCode is nil.
	TaxRate *decimal.Decimal `json:"tax_rate,omitempty"`

	// -------------------------------------------------------------------------
	// Balance (denormalised cache)
	// -------------------------------------------------------------------------

	// CurrentBalance is the running balance maintained by the journal engine.
	// For debit-normal accounts, a positive value means net debit position.
	// For credit-normal accounts, a positive value means net credit position.
	CurrentBalance decimal.Decimal `json:"current_balance"`

	// YTDBalance is the year-to-date movement, reset at each financial year open.
	YTDBalance decimal.Decimal `json:"ytd_balance"`

	// LastTransactionDate is the date of the most recent posted journal entry.
	LastTransactionDate *time.Time `json:"last_transaction_date,omitempty"`

	// -------------------------------------------------------------------------
	// Budgeting
	// -------------------------------------------------------------------------

	// IsBudgetable allows budget lines to be created for this account.
	// Typically true for revenue and expense accounts, false for balance sheet.
	IsBudgetable bool `json:"is_budgetable"`

	// BudgetVarianceThreshold is the percentage deviation from budget that
	// triggers an alert. Range [0, 100]. Zero means alerts are disabled.
	BudgetVarianceThreshold decimal.Decimal `json:"budget_variance_threshold"`

	// -------------------------------------------------------------------------
	// Lifecycle & validation
	// -------------------------------------------------------------------------

	Status           AccountStatus     `json:"status"`
	ValidationStatus ValidationStatus  `json:"validation_status"`
	ValidationErrors []ValidationError `json:"validation_errors,omitempty"`

	// -------------------------------------------------------------------------
	// Extensions
	// -------------------------------------------------------------------------

	// AccountAttributes holds tenant-defined key-value metadata for
	// industry-specific extensions (e.g. EPRA product codes, KRA tax mappings).
	AccountAttributes map[string]any `json:"account_attributes,omitempty"`

	// -------------------------------------------------------------------------
	// Audit trail
	// -------------------------------------------------------------------------

	Version   int32      `json:"version"` // Optimistic locking counter
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	CreatedBy uuid.UUID  `json:"created_by"`
	UpdatedBy *uuid.UUID `json:"updated_by,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"` // Soft delete
}

// =============================================================================
// State machine methods
// =============================================================================

// CanTransitionTo reports whether a move from the current Status to newStatus
// is listed in AllowedTransitions.
func (a *Accounts) CanTransitionTo(newStatus AccountStatus) bool {
	for _, t := range AllowedTransitions {
		if t.From == a.Status && t.To == newStatus {
			return true
		}
	}
	return false
}

// TransitionTo returns the matching StatusTransition for the caller to enforce
// the required permission and validation check, or an error when the transition
// is not permitted.
func (a *Accounts) TransitionTo(newStatus AccountStatus) (StatusTransition, error) {
	for _, t := range AllowedTransitions {
		if t.From == a.Status && t.To == newStatus {
			return t, nil
		}
	}
	return StatusTransition{}, fmt.Errorf(
		"account status transition %s → %s is not permitted", a.Status, newStatus,
	)
}

// =============================================================================
// Behaviour methods
// =============================================================================

// CanAcceptTransactions returns true when a journal line may post to this account.
//
// All conditions must hold simultaneously:
//   - Status == ACTIVE
//   - IsActive == true
//   - IsLeaf == true  (only leaf accounts accept direct postings)
//   - DeletedAt == nil
//   - ValidationStatus == VALID
func (a *Accounts) CanAcceptTransactions() bool {
	return a.Status == AccountStatusActive &&
		a.IsActive &&
		a.IsLeaf &&
		a.DeletedAt == nil &&
		a.ValidationStatus == ValidationStatusValid
}

// CanAcceptManualEntries returns true when a human-authored journal entry may
// reference this account (as opposed to system-generated postings only).
func (a *Accounts) CanAcceptManualEntries() bool {
	return a.CanAcceptTransactions() &&
		a.AllowManualEntries &&
		!a.IsControlAccount
}

// CanBeDeactivated returns true when it is safe to move the account to INACTIVE.
// Balance must be zero and the account must not be a protected system account.
func (a *Accounts) CanBeDeactivated() bool {
	return a.CurrentBalance.IsZero() && !a.IsSystemAccount
}

// CanBeDeleted returns true when soft-deletion is permitted.
// An account may not be deleted if it has ever had transactions
// (indicated by a non-nil LastTransactionDate) or is a system account.
func (a *Accounts) CanBeDeleted() bool {
	return a.LastTransactionDate == nil && !a.IsSystemAccount
}

// IsDeleted returns true when the account has been soft-deleted.
func (a *Accounts) IsDeleted() bool { return a.DeletedAt != nil }

// IsDescendantOf reports whether this account is a descendant of parentID
// using the materialised Path.
func (a *Accounts) IsDescendantOf(parentID uuid.UUID) bool {
	return a.Path.IsDescendantOf(parentID)
}

// GetEffectiveBalance returns the balance with sign adjusted for presentation.
// Debit-normal accounts return the balance as-is; credit-normal accounts
// return the negated balance so that a positive value always means
// "this side of the balance sheet has a balance".
func (a *Accounts) GetEffectiveBalance() decimal.Decimal {
	if a.NormalBalance == NormalBalanceDebit {
		return a.CurrentBalance
	}
	return a.CurrentBalance.Neg()
}

// GetTransactionRestrictions returns human-readable restriction messages based
// on the account's current status and settings. Used to explain to UI users
// why a posting was rejected.
func (a *Accounts) GetTransactionRestrictions() []string {
	var r []string
	switch a.Status {
	case AccountStatusRestricted:
		r = append(r, "manual entries require approval in restricted mode")
	case AccountStatusFrozen:
		r = append(r, "account is frozen — no new transactions")
	case AccountStatusYearEndProcessing:
		r = append(r, "only period-closing entries are permitted during year-end processing")
	case AccountStatusAuditLock:
		r = append(r, "account is read-only during external audit")
	case AccountStatusComplianceHold:
		r = append(r, "account is under a regulatory compliance hold")
	case AccountStatusSystemMaintenance:
		r = append(r, "account is temporarily unavailable during system maintenance")
	}
	if !a.AllowManualEntries {
		r = append(r, "account accepts system-generated entries only")
	}
	if a.IsControlAccount {
		r = append(r, "control accounts do not accept direct manual postings")
	}
	if !a.IsLeaf {
		r = append(r, "only leaf accounts accept direct postings; post to a child account")
	}
	if a.RequireReference {
		r = append(r, "a reference number is mandatory for all entries on this account")
	}
	return r
}

// RequiresApproval returns true when the account is awaiting administrative
// action and should be surfaced in approval queues.
func (a *Accounts) RequiresApproval() bool {
	return a.Status == AccountStatusPendingApproval ||
		a.Status == AccountStatusUnderReview
}

// MarshalAccountAttributes serialises AccountAttributes to JSON.
func (a *Accounts) MarshalAccountAttributes() ([]byte, error) {
	if a.AccountAttributes == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(a.AccountAttributes)
}

// UnmarshalAccountAttributes deserialises AccountAttributes from JSON.
func (a *Accounts) UnmarshalAccountAttributes(data []byte) error {
	if len(data) == 0 {
		a.AccountAttributes = make(map[string]any)
		return nil
	}
	return json.Unmarshal(data, &a.AccountAttributes)
}

// =============================================================================
// Validation
// =============================================================================

// Validate performs entity-level consistency checks.
// Cross-entity rules (parent existence, RootType homogeneity, control account
// children) are in ValidateWithContext which requires external data.
func (a *Accounts) Validate() []ValidationError {
	var errs []ValidationError

	// -- Tenant ---------------------------------------------------------------
	if a.TenantID == uuid.Nil {
		errs = append(errs, ValidationError{
			Field: "tenant_id", Message: "tenant_id is required",
			Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
		})
	}

	// -- Code -----------------------------------------------------------------
	switch {
	case strings.TrimSpace(a.AccountCode) == "":
		errs = append(errs, ValidationError{
			Field: "account_code", Message: "account code is required",
			Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
		})
	case len(a.AccountCode) > AccountCodeMaxLen:
		errs = append(errs, ValidationError{
			Field: "account_code",
			Message: fmt.Sprintf(
				"account code must be %d characters or less", AccountCodeMaxLen,
			),
			Code: "MAX_LENGTH_EXCEEDED", Severity: ValidationSeverityError,
		})
	case !accountCodePattern.MatchString(a.AccountCode):
		errs = append(errs, ValidationError{
			Field:    "account_code",
			Message:  "account code may only contain alphanumeric characters and hyphens",
			Code:     "INVALID_FORMAT",
			Severity: ValidationSeverityError,
		})
	}

	// -- Name -----------------------------------------------------------------
	switch {
	case strings.TrimSpace(a.AccountName) == "":
		errs = append(errs, ValidationError{
			Field: "account_name", Message: "account name is required",
			Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
		})
	case len(a.AccountName) > AccountNameMaxLen:
		errs = append(errs, ValidationError{
			Field: "account_name",
			Message: fmt.Sprintf(
				"account name must be %d characters or less", AccountNameMaxLen,
			),
			Code: "MAX_LENGTH_EXCEEDED", Severity: ValidationSeverityError,
		})
	}

	// -- Description ----------------------------------------------------------
	if a.AccountDescription != nil && len(*a.AccountDescription) > AccountDescriptionMaxLen {
		errs = append(errs, ValidationError{
			Field: "account_description",
			Message: fmt.Sprintf(
				"account description must be %d characters or less", AccountDescriptionMaxLen,
			),
			Code: "MAX_LENGTH_EXCEEDED", Severity: ValidationSeverityError,
		})
	}

	// -- Classification -------------------------------------------------------
	if !a.RootType.IsValid() {
		errs = append(errs, ValidationError{
			Field:    "root_type",
			Message:  fmt.Sprintf("invalid root type: %q", a.RootType),
			Code:     "INVALID_VALUE",
			Severity: ValidationSeverityError,
		})
	}
	if !a.NormalBalance.IsValid() {
		errs = append(errs, ValidationError{
			Field:    "normal_balance",
			Message:  fmt.Sprintf("invalid normal balance: %q", a.NormalBalance),
			Code:     "INVALID_VALUE",
			Severity: ValidationSeverityError,
		})
	}
	// Normal balance must match the conventional balance for the root type.
	if a.RootType.IsValid() && a.NormalBalance.IsValid() {
		expected := GetNormalBalanceForRootType(a.RootType)
		if a.NormalBalance != expected {
			errs = append(errs, ValidationError{
				Field: "normal_balance",
				Message: fmt.Sprintf(
					"normal balance must be %s for root type %s", expected, a.RootType,
				),
				Code:     "INCONSISTENT_VALUE",
				Severity: ValidationSeverityError,
			})
		}
	}

	// -- Status ---------------------------------------------------------------
	if !a.Status.IsValid() {
		errs = append(errs, ValidationError{
			Field:    "status",
			Message:  fmt.Sprintf("invalid account status: %q", a.Status),
			Code:     "INVALID_VALUE",
			Severity: ValidationSeverityError,
		})
	}

	// -- Hierarchy consistency ------------------------------------------------
	if a.ParentAccountID == nil && a.AccountLevel != 0 {
		errs = append(errs, ValidationError{
			Field:    "account_level",
			Message:  "root account (no parent) must have level 0",
			Code:     "INCONSISTENT_VALUE",
			Severity: ValidationSeverityError,
		})
	}
	if a.ParentAccountID != nil && a.AccountLevel <= 0 {
		errs = append(errs, ValidationError{
			Field:    "account_level",
			Message:  "child account (has parent) must have level > 0",
			Code:     "INCONSISTENT_VALUE",
			Severity: ValidationSeverityError,
		})
	}
	// HasChildren and IsLeaf must be consistent.
	if a.HasChildren == a.IsLeaf {
		errs = append(errs, ValidationError{
			Field:    "is_leaf",
			Message:  "is_leaf must be the inverse of has_children",
			Code:     "INCONSISTENT_VALUE",
			Severity: ValidationSeverityError,
		})
	}
	// Validate the materialised path if set.
	if err := a.Path.Validate(); err != nil {
		errs = append(errs, ValidationError{
			Field:    "path",
			Message:  err.Error(),
			Code:     "INVALID_FORMAT",
			Severity: ValidationSeverityError,
		})
	}

	// -- Currency -------------------------------------------------------------
	if a.CurrencyCode != nil && len(*a.CurrencyCode) != CurrencyCodeLen {
		errs = append(errs, ValidationError{
			Field: "currency_code",
			Message: fmt.Sprintf(
				"currency code must be exactly %d characters (ISO 4217)", CurrencyCodeLen,
			),
			Code:     "INVALID_FORMAT",
			Severity: ValidationSeverityError,
		})
	}

	// -- Business rules (self-contained) -------------------------------------

	// Control accounts must not allow manual entries; they are driven by
	// sub-ledger postings only.
	if a.IsControlAccount && a.AllowManualEntries {
		errs = append(errs, ValidationError{
			Field:    "allow_manual_entries",
			Message:  "control accounts must not allow manual entries",
			Code:     "BUSINESS_RULE_VIOLATION",
			Severity: ValidationSeverityError,
		})
	}
	// An account cannot be both a control account and have a controlling account.
	if a.IsControlAccount && a.ControlAccountID != nil {
		errs = append(errs, ValidationError{
			Field:    "control_account_id",
			Message:  "a control account cannot itself be controlled by another account",
			Code:     "BUSINESS_RULE_VIOLATION",
			Severity: ValidationSeverityError,
		})
	}

	// -- Tax consistency ------------------------------------------------------
	if a.TaxRate != nil && a.TaxCode == nil {
		errs = append(errs, ValidationError{
			Field:    "tax_code",
			Message:  "tax_code is required when tax_rate is set",
			Code:     "REQUIRED_FIELD",
			Severity: ValidationSeverityError,
		})
	}
	if a.TaxRate != nil {
		zero := decimal.Zero
		hundred := decimal.NewFromInt(100)
		if a.TaxRate.LessThan(zero) || a.TaxRate.GreaterThan(hundred) {
			errs = append(errs, ValidationError{
				Field:    "tax_rate",
				Message:  "tax rate must be between 0 and 100",
				Code:     "OUT_OF_RANGE",
				Severity: ValidationSeverityError,
			})
		}
	}

	// -- Budget ---------------------------------------------------------------
	zero := decimal.Zero
	hundred := decimal.NewFromInt(100)
	if a.BudgetVarianceThreshold.LessThan(zero) || a.BudgetVarianceThreshold.GreaterThan(hundred) {
		errs = append(errs, ValidationError{
			Field:    "budget_variance_threshold",
			Message:  "budget variance threshold must be between 0 and 100",
			Code:     "OUT_OF_RANGE",
			Severity: ValidationSeverityError,
		})
	}

	return errs
}

// ValidateWithContext performs cross-entity business rule checks that require
// data fetched from outside the entity itself. Call this after Validate().
//
//   - parentAccount: the resolved parent, or nil if this is a root account.
//   - hasChildren: true if child account records exist in the repository.
//   - hasTransactions: true if any posted journal lines reference this account.
func (a *Accounts) ValidateWithContext(
	parentAccount *Accounts,
	hasChildren bool,
	hasTransactions bool,
) []ValidationError {
	var errs []ValidationError

	// Parent must share the same RootType.
	if a.ParentAccountID != nil && parentAccount != nil {
		if parentAccount.RootType != a.RootType {
			errs = append(errs, ValidationError{
				Field: "root_type",
				Message: fmt.Sprintf(
					"account root type %s must match parent root type %s",
					a.RootType, parentAccount.RootType,
				),
				Code:     "INCONSISTENT_ROOT_TYPE",
				Severity: ValidationSeverityError,
			})
		}
		if a.AccountLevel != parentAccount.AccountLevel+1 {
			errs = append(errs, ValidationError{
				Field: "account_level",
				Message: fmt.Sprintf(
					"account level must be exactly one more than parent level (%d)",
					parentAccount.AccountLevel,
				),
				Code:     "INVALID_HIERARCHY_LEVEL",
				Severity: ValidationSeverityError,
			})
		}
	}

	// Control accounts must have at least one child (sub-ledger account).
	if a.IsControlAccount && !hasChildren {
		errs = append(errs, ValidationError{
			Field:    "is_control_account",
			Message:  "control accounts must have at least one child sub-ledger account",
			Code:     "CONTROL_ACCOUNT_NO_CHILDREN",
			Severity: ValidationSeverityError,
		})
	}

	// Cannot deactivate an account with a non-zero balance or live transactions.
	if !a.IsActive && hasTransactions && !a.CurrentBalance.IsZero() {
		errs = append(errs, ValidationError{
			Field:    "is_active",
			Message:  "cannot deactivate an account with a non-zero balance",
			Code:     "CANNOT_DEACTIVATE_WITH_BALANCE",
			Severity: ValidationSeverityError,
		})
	}

	return errs
}

// =============================================================================
// Request DTOs
// =============================================================================

// CreateAccountRequest is the inbound payload for creating a new ledger account.
//
// AccountLevel and Path are derived by the service layer from ParentAccountID.
// Status is always initialised to DRAFT; the caller must explicitly submit for
// approval or activate via a TransitionTo call.
type CreateAccountRequest struct {
	EntityID           *uuid.UUID `json:"entity_id,omitempty"`
	AccountCode        string     `json:"account_code"        validate:"required,max=20"`
	AccountName        string     `json:"account_name"        validate:"required,max=200"`
	AccountDescription *string    `json:"account_description,omitempty" validate:"omitempty,max=1000"`

	// Hierarchy
	ParentAccountID *uuid.UUID `json:"parent_account_id,omitempty"`
	AccountGroupID  *uuid.UUID `json:"account_group_id,omitempty"`

	// Classification
	RootType        RootType      `json:"root_type"        validate:"required"`
	AccountType     string        `json:"account_type"     validate:"required,max=100"`
	AccountSubtype  *string       `json:"account_subtype,omitempty"  validate:"omitempty,max=100"`
	AccountCategory *string       `json:"account_category,omitempty" validate:"omitempty,max=100"`
	SubCategory     *string       `json:"sub_category,omitempty"     validate:"omitempty,max=100"`
	NormalBalance   NormalBalance `json:"normal_balance"  validate:"required"`

	// Control
	IsControlAccount bool       `json:"is_control_account"`
	ControlAccountID *uuid.UUID `json:"control_account_id,omitempty"`

	// Currency
	CurrencyCode                *string `json:"currency_code,omitempty"       validate:"omitempty,len=3"`
	IsMultiCurrency             bool    `json:"is_multi_currency"`
	CurrencyRevaluationRequired bool    `json:"currency_revaluation_required"`

	// Operational
	IsActive           bool `json:"is_active"`
	AllowManualEntries bool `json:"allow_manual_entries"`
	RequireReference   bool `json:"require_reference"`

	// Reconciliation & tax
	RequiresReconciliation bool             `json:"requires_reconciliation"`
	TaxCode                *string          `json:"tax_code,omitempty"`
	TaxRate                *decimal.Decimal `json:"tax_rate,omitempty" validate:"omitempty,min=0,max=100"`

	// Budgeting
	IsBudgetable            bool            `json:"is_budgetable"`
	BudgetVarianceThreshold decimal.Decimal `json:"budget_variance_threshold" validate:"min=0,max=100"`

	// Extensions
	AccountAttributes map[string]any `json:"account_attributes,omitempty"`
}

// Validate checks request-level consistency before the service layer applies
// cross-entity rules.
func (r *CreateAccountRequest) Validate() []ValidationError {
	var errs []ValidationError

	if strings.TrimSpace(r.AccountCode) == "" {
		errs = append(errs, ValidationError{
			Field: "account_code", Message: "account code is required",
			Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
		})
	} else if len(r.AccountCode) > AccountCodeMaxLen {
		errs = append(errs, ValidationError{
			Field:   "account_code",
			Message: fmt.Sprintf("account code must be %d characters or less", AccountCodeMaxLen),
			Code:    "MAX_LENGTH_EXCEEDED", Severity: ValidationSeverityError,
		})
	} else if !accountCodePattern.MatchString(r.AccountCode) {
		errs = append(errs, ValidationError{
			Field:    "account_code",
			Message:  "account code may only contain alphanumeric characters and hyphens",
			Code:     "INVALID_FORMAT",
			Severity: ValidationSeverityError,
		})
	}
	if strings.TrimSpace(r.AccountName) == "" {
		errs = append(errs, ValidationError{
			Field: "account_name", Message: "account name is required",
			Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
		})
	} else if len(r.AccountName) > AccountNameMaxLen {
		errs = append(errs, ValidationError{
			Field:   "account_name",
			Message: fmt.Sprintf("account name must be %d characters or less", AccountNameMaxLen),
			Code:    "MAX_LENGTH_EXCEEDED", Severity: ValidationSeverityError,
		})
	}
	if !r.RootType.IsValid() {
		errs = append(errs, ValidationError{
			Field:    "root_type",
			Message:  fmt.Sprintf("invalid root type: %q", r.RootType),
			Code:     "INVALID_VALUE",
			Severity: ValidationSeverityError,
		})
	}
	if !r.NormalBalance.IsValid() {
		errs = append(errs, ValidationError{
			Field:    "normal_balance",
			Message:  fmt.Sprintf("invalid normal balance: %q", r.NormalBalance),
			Code:     "INVALID_VALUE",
			Severity: ValidationSeverityError,
		})
	}
	if r.RootType.IsValid() && r.NormalBalance.IsValid() {
		expected := GetNormalBalanceForRootType(r.RootType)
		if r.NormalBalance != expected {
			errs = append(errs, ValidationError{
				Field: "normal_balance",
				Message: fmt.Sprintf(
					"normal balance must be %s for root type %s", expected, r.RootType,
				),
				Code: "INCONSISTENT_VALUE", Severity: ValidationSeverityError,
			})
		}
	}
	if r.IsControlAccount && r.AllowManualEntries {
		errs = append(errs, ValidationError{
			Field:    "allow_manual_entries",
			Message:  "control accounts must not allow manual entries",
			Code:     "BUSINESS_RULE_VIOLATION",
			Severity: ValidationSeverityError,
		})
	}
	if r.TaxRate != nil && r.TaxCode == nil {
		errs = append(errs, ValidationError{
			Field: "tax_code", Message: "tax_code is required when tax_rate is set",
			Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
		})
	}

	return errs
}

// UpdateAccountRequest is the partial-update payload.
//
// AccountCode and ParentAccountID are excluded:
//   - Code changes need a uniqueness re-check via a dedicated service operation.
//   - Parent changes require a full subtree re-path via a dedicated Reparent operation.
//
// RootType and NormalBalance are also excluded: reclassification is a controlled
// workflow that must zero the balance and move all mappings.
type UpdateAccountRequest struct {
	AccountName        *string `json:"account_name,omitempty"        validate:"omitempty,max=200"`
	AccountDescription *string `json:"account_description,omitempty" validate:"omitempty,max=1000"`
	AccountType        *string `json:"account_type,omitempty"        validate:"omitempty,max=100"`
	AccountSubtype     *string `json:"account_subtype,omitempty"     validate:"omitempty,max=100"`
	AccountCategory    *string `json:"account_category,omitempty"    validate:"omitempty,max=100"`
	SubCategory        *string `json:"sub_category,omitempty"        validate:"omitempty,max=100"`

	AccountGroupID *uuid.UUID `json:"account_group_id,omitempty"`

	IsActive           *bool `json:"is_active,omitempty"`
	AllowManualEntries *bool `json:"allow_manual_entries,omitempty"`
	RequireReference   *bool `json:"require_reference,omitempty"`

	RequiresReconciliation *bool            `json:"requires_reconciliation,omitempty"`
	TaxCode                *string          `json:"tax_code,omitempty"`
	TaxRate                *decimal.Decimal `json:"tax_rate,omitempty" validate:"omitempty,min=0,max=100"`

	IsBudgetable            *bool            `json:"is_budgetable,omitempty"`
	BudgetVarianceThreshold *decimal.Decimal `json:"budget_variance_threshold,omitempty" validate:"omitempty,min=0,max=100"`

	AccountAttributes map[string]any `json:"account_attributes,omitempty"`
}

// Validate checks the update request for internal consistency.
func (r *UpdateAccountRequest) Validate() []ValidationError {
	var errs []ValidationError

	if r.AccountName != nil {
		if strings.TrimSpace(*r.AccountName) == "" {
			errs = append(errs, ValidationError{
				Field: "account_name", Message: "account name cannot be empty",
				Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
			})
		} else if len(*r.AccountName) > AccountNameMaxLen {
			errs = append(errs, ValidationError{
				Field:   "account_name",
				Message: fmt.Sprintf("account name must be %d characters or less", AccountNameMaxLen),
				Code:    "MAX_LENGTH_EXCEEDED", Severity: ValidationSeverityError,
			})
		}
	}
	if r.TaxRate != nil && r.TaxCode == nil {
		errs = append(errs, ValidationError{
			Field: "tax_code", Message: "tax_code is required when tax_rate is set",
			Code: "REQUIRED_FIELD", Severity: ValidationSeverityError,
		})
	}
	if r.BudgetVarianceThreshold != nil {
		zero := decimal.Zero
		hundred := decimal.NewFromInt(100)
		if r.BudgetVarianceThreshold.LessThan(zero) || r.BudgetVarianceThreshold.GreaterThan(hundred) {
			errs = append(errs, ValidationError{
				Field:    "budget_variance_threshold",
				Message:  "budget variance threshold must be between 0 and 100",
				Code:     "OUT_OF_RANGE",
				Severity: ValidationSeverityError,
			})
		}
	}

	return errs
}

// =============================================================================
// Read-side projections (used by query handlers, not the write path)
// =============================================================================

// AccountBalance captures an account's balance position at a specific point in
// time, computed by the reporting engine from journal entries.
type AccountBalance struct {
	AccountID    uuid.UUID       `json:"account_id"`
	TotalDebits  decimal.Decimal `json:"total_debits"`
	TotalCredits decimal.Decimal `json:"total_credits"`
	NetBalance   decimal.Decimal `json:"net_balance"`
	AsOfDate     time.Time       `json:"as_of_date"`
}

// TrialBalanceEntry is a single line in the trial balance report.
type TrialBalanceEntry struct {
	AccountID     uuid.UUID       `json:"account_id"`
	AccountCode   string          `json:"account_code"`
	AccountName   string          `json:"account_name"`
	RootType      RootType        `json:"root_type"`
	NormalBalance NormalBalance   `json:"normal_balance"`
	TotalDebits   decimal.Decimal `json:"total_debits"`
	TotalCredits  decimal.Decimal `json:"total_credits"`
	NetBalance    decimal.Decimal `json:"net_balance"`
}

// =============================================================================
// Filter types
// =============================================================================

// AccountFilter is the query predicate for listing accounts.
type AccountFilter struct {
	EntityID    *uuid.UUID     `json:"entity_id,omitempty"`
	RootType    *RootType      `json:"root_type,omitempty"`
	AccountType *string        `json:"account_type,omitempty"`
	Status      *AccountStatus `json:"status,omitempty"`
	ParentID    *uuid.UUID     `json:"parent_account_id,omitempty"`
	IsActive    *bool          `json:"is_active,omitempty"`
	IsLeaf      *bool          `json:"is_leaf,omitempty"`
	NonZeroOnly *bool          `json:"non_zero_only,omitempty"`
	SearchQuery *string        `json:"search_query,omitempty"`

	// Hierarchy
	MaxLevel        *int32 `json:"max_level,omitempty"`
	IncludeChildren bool   `json:"include_children"`

	// Pagination
	Limit  *int `json:"limit,omitempty"`
	Offset *int `json:"offset,omitempty"`

	// Sorting
	SortBy    *string `json:"sort_by,omitempty"`
	SortOrder *string `json:"sort_order,omitempty"`
}

// Validate checks that the filter's values are self-consistent.
func (f *AccountFilter) Validate() []ValidationError {
	var errs []ValidationError
	if f.RootType != nil && !f.RootType.IsValid() {
		errs = append(errs, ValidationError{
			Field:    "root_type",
			Message:  fmt.Sprintf("invalid root type: %q", *f.RootType),
			Code:     "INVALID_VALUE",
			Severity: ValidationSeverityError,
		})
	}
	if f.Status != nil && !f.Status.IsValid() {
		errs = append(errs, ValidationError{
			Field:    "status",
			Message:  fmt.Sprintf("invalid account status: %q", *f.Status),
			Code:     "INVALID_VALUE",
			Severity: ValidationSeverityError,
		})
	}
	if f.Limit != nil && *f.Limit < 0 {
		errs = append(errs, ValidationError{
			Field: "limit", Message: "limit must be non-negative",
			Code: "OUT_OF_RANGE", Severity: ValidationSeverityError,
		})
	}
	if f.Offset != nil && *f.Offset < 0 {
		errs = append(errs, ValidationError{
			Field: "offset", Message: "offset must be non-negative",
			Code: "OUT_OF_RANGE", Severity: ValidationSeverityError,
		})
	}
	return errs
}

// NOTE: Consider adding account category groupings for enhanced reporting
// NOTE: Future enhancement: Add support for account-specific posting rules
