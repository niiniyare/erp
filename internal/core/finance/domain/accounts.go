// Package domain contains the core ledger account entity and all directly
// related types: status constants, the state machine transition table,
// classification enumerations, validation helpers, request/response DTOs,
// and read-side projections.
//
// # File organisation (read top-to-bottom)
//
//  1. Account-code constants & pattern       – 8-digit segmentation rules
//  2. AccountStatus & state machine          – lifecycle constants + transition table
//  3. Supporting enumerations                – RootType, NormalBalance, ValidationStatus
//  4. Accounts entity                        – the ledger account aggregate root
//  5. State machine methods                  – CanTransitionTo / TransitionTo
//  6. Behaviour methods                      – posting guards, balance helpers
//  7. Shared validation helpers (private)    – reusable building blocks
//  8. Validate / ValidateWithContext         – entity-level validation entry points
//  9. Request DTOs                           – Create / Update payloads
//
// 10. Read-side projections                  – AccountBalance, TrialBalanceEntry
// 11. AccountFilter                          – list/query predicate
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
// 1. Account-code constants & segmentation pattern
// =============================================================================

// Account codes use a fixed 8-digit numeric format that encodes the account's
// position in the chart of accounts hierarchy. This makes codes both
// human-readable and programmatically parseable without a database round-trip.
//
// Segment layout (all segments are 2 digits, zero-padded):
//
//	┌──────────────────────────────────────────────────────┐
//	│  Pos 1-2  │  Pos 3-4  │  Pos 5-6  │  Pos 7-8       │
//	│  RootSeg  │  Category │ SubCat    │  Leaf detail    │
//	└──────────────────────────────────────────────────────┘
//
// Recommended root-segment values (enforced by convention, not code):
//
//	10xxxxxx  → ASSET
//	20xxxxxx  → LIABILITY
//	30xxxxxx  → EQUITY
//	40xxxxxx  → REVENUE
//	50xxxxxx  → EXPENSE
//
// Examples:
//
//	10000000  Asset root (level 0, parent of all asset accounts)
//	10010000  Current Assets (level 1, category 01)
//	10010100  Cash & Cash Equivalents (level 2, sub-category 01)
//	10010101  Petty Cash (level 3, leaf account 01)
//	10010102  Main Operating Account (level 3, leaf account 02)
//	20010101  Accounts Payable – Trade (liability leaf)
//
// Rules:
//   - Parent code is always the code of the same length with the child's
//     right-most non-zero segment zeroed out, e.g. 10010101's parent is 10010100.
//   - Root accounts have the form XX000000 (only the root segment is non-zero).
//   - Leaf accounts have all 8 digits non-zero (or at minimum the last segment non-zero).
//
// These are conventions documented here for developers; they are NOT enforced
// by the domain layer because the service layer resolves ParentAccountID
// explicitly and does not derive it from the code alone.

const (
	// AccountCodeLen is the exact required length of every account code.
	// Fixed length enables prefix-based hierarchy queries (LIKE '1001%').
	AccountCodeLen = 8

	// AccountNameMaxLen caps the display label used in reports and selectors.
	AccountNameMaxLen = 200

	// AccountDescriptionMaxLen caps the long-form purpose description.
	AccountDescriptionMaxLen = 1000

	// AccountTypeMaxLen caps the secondary classification label.
	AccountTypeMaxLen = 100

	// AccountSubtypeMaxLen caps the tertiary classification label.
	AccountSubtypeMaxLen = 100

	// CurrencyCodeLen is the ISO 4217 fixed-width currency code length.
	CurrencyCodeLen = 3
)

// accountCodePattern matches a string that is exactly AccountCodeLen (8) digits.
// Compiled once at package initialisation; reused for every validation call.
//
// Why digits only (not alphanumeric)?
// Numeric codes support:
//   - Prefix-range queries:  WHERE code >= '10000000' AND code < '20000000'
//   - Arithmetic derivation: parent = (code / 100) * 100
//   - Sortable display order matching the natural accounting sequence
var accountCodePattern = regexp.MustCompile(`^[0-9]{8}$`)

// AccountCodeSegments breaks an 8-digit code into its four 2-digit segments.
// Returns an error if the code does not match the required format.
// Useful for display, hierarchy derivation, and debugging.
//
// Example: AccountCodeSegments("10010101") → [10, 01, 01, 01], nil
func AccountCodeSegments(code string) ([4]string, error) {
	var segs [4]string
	if !accountCodePattern.MatchString(code) {
		return segs, fmt.Errorf(
			"account code %q is not a valid 8-digit code", code,
		)
	}
	segs[0] = code[0:2] // Root segment  (e.g. "10" = Asset)
	segs[1] = code[2:4] // Category      (e.g. "01" = Current Assets)
	segs[2] = code[4:6] // Sub-category  (e.g. "01" = Cash & Equivalents)
	segs[3] = code[6:8] // Leaf detail   (e.g. "01" = Petty Cash)
	return segs, nil
}

// DeriveParentCode returns the code of the immediate parent account by zeroing
// out the right-most non-zero 2-digit segment.
//
// Examples:
//
//	"10010101" → "10010100"  (leaf → sub-category)
//	"10010100" → "10010000"  (sub-category → category)
//	"10010000" → "10000000"  (category → root)
//	"10000000" → ""          (root → no parent)
//
// Returns ("", false) when code is a root-level code (XX000000).
func DeriveParentCode(code string) (string, bool) {
	if !accountCodePattern.MatchString(code) {
		return "", false
	}
	// Walk segments right-to-left (positions 6-7, 4-5, 2-3).
	// Zero out the first non-zero segment we encounter.
	b := []byte(code)
	for start := 6; start >= 2; start -= 2 {
		if b[start] != '0' || b[start+1] != '0' {
			b[start] = '0'
			b[start+1] = '0'
			return string(b), true
		}
	}
	// All segments below the root segment are already "00" → this is a root.
	return "", false
}

// AccountHierarchyLevel returns the zero-based depth implied by an 8-digit
// account code. Root accounts (XX000000) are level 0; fully-qualified leaf
// accounts (XXXXXXXX with no trailing zeroed pair) are level 3.
//
// This is a convenience function for display and debugging. The authoritative
// level is stored in Accounts.AccountLevel and maintained by the repository.
func AccountHierarchyLevel(code string) (int, error) {
	if !accountCodePattern.MatchString(code) {
		return 0, fmt.Errorf("account code %q is not a valid 8-digit code", code)
	}
	level := 3
	// Each trailing "00" pair reduces the level by one.
	for start := 6; start >= 2; start -= 2 {
		if code[start] == '0' && code[start+1] == '0' {
			level--
		} else {
			break
		}
	}
	return level, nil
}

// =============================================================================
// 2. AccountStatus & state machine
// =============================================================================

// AccountStatus represents the full operational lifecycle of a ledger account.
//
// State machine overview (simplified):
//
//	DRAFT ──► PENDING_APPROVAL ──► ACTIVE ──► INACTIVE ──► CLOSED ──► ARCHIVED
//	                                  │
//	                    reversible holds: SUSPENDED, RESTRICTED, FROZEN,
//	                    UNDER_REVIEW, YEAR_END_PROCESSING, AUDIT_LOCK,
//	                    COMPLIANCE_HOLD, SYSTEM_MAINTENANCE, DATA_ERROR
//
// Authoritative transitions are defined in AllowedTransitions below.
// Never add ad-hoc if-chains; extend the table instead.
type AccountStatus string

const (
	// AccountStatusDraft — account is being configured; invisible to posting workflows.
	AccountStatusDraft AccountStatus = "DRAFT"

	// AccountStatusPendingApproval — submitted for review; awaiting authorisation.
	AccountStatusPendingApproval AccountStatus = "PENDING_APPROVAL"

	// AccountStatusActive — fully operational; accepts transactions subject to
	// per-account settings (AllowManualEntries, IsLeaf, etc.).
	AccountStatusActive AccountStatus = "ACTIVE"

	// AccountStatusInactive — voluntarily suspended; no new postings allowed;
	// a zero balance is required to enter this state.
	AccountStatusInactive AccountStatus = "INACTIVE"

	// AccountStatusSuspended — administrative hold; temporary, expects resolution.
	AccountStatusSuspended AccountStatus = "SUSPENDED"

	// AccountStatusRestricted — limited to approval-only or read-only entries.
	AccountStatusRestricted AccountStatus = "RESTRICTED"

	// AccountStatusFrozen — hard freeze; no transactions of any kind permitted.
	AccountStatusFrozen AccountStatus = "FROZEN"

	// AccountStatusUnderReview — flagged for internal audit or compliance review.
	AccountStatusUnderReview AccountStatus = "UNDER_REVIEW"

	// AccountStatusYearEndProcessing — locked for period close; only closing
	// entries are permitted.
	AccountStatusYearEndProcessing AccountStatus = "YEAR_END_PROCESSING"

	// AccountStatusAuditLock — external audit in progress; read-only for the duration.
	AccountStatusAuditLock AccountStatus = "AUDIT_LOCK"

	// AccountStatusComplianceHold — regulatory or sanctions hold; no transactions.
	AccountStatusComplianceHold AccountStatus = "COMPLIANCE_HOLD"

	// AccountStatusSystemMaintenance — brief technical window (e.g. data migration).
	AccountStatusSystemMaintenance AccountStatus = "SYSTEM_MAINTENANCE"

	// AccountStatusDataError — automated detection of a data integrity issue;
	// pending manual correction before the account can be reactivated.
	AccountStatusDataError AccountStatus = "DATA_ERROR"

	// AccountStatusClosed — terminal operational state; balance must be zero.
	// Retained in the system for historical reporting. Cannot be reopened.
	AccountStatusClosed AccountStatus = "CLOSED"

	// AccountStatusArchived — moved to long-term storage after closure.
	// Read-only forever; the only terminal state after CLOSED.
	AccountStatusArchived AccountStatus = "ARCHIVED"
)

// accountStatusSet is the single source of truth for valid status values.
// IsValid() derives from this map; adding a new status here is sufficient.
var accountStatusSet = map[AccountStatus]struct{}{
	AccountStatusDraft:             {},
	AccountStatusPendingApproval:   {},
	AccountStatusActive:            {},
	AccountStatusInactive:          {},
	AccountStatusSuspended:         {},
	AccountStatusRestricted:        {},
	AccountStatusFrozen:            {},
	AccountStatusUnderReview:       {},
	AccountStatusYearEndProcessing: {},
	AccountStatusAuditLock:         {},
	AccountStatusComplianceHold:    {},
	AccountStatusSystemMaintenance: {},
	AccountStatusDataError:         {},
	AccountStatusClosed:            {},
	AccountStatusArchived:          {},
}

// IsValid returns true if s is a recognised status constant.
func (s AccountStatus) IsValid() bool {
	_, ok := accountStatusSet[s]
	return ok
}

// String implements fmt.Stringer.
func (s AccountStatus) String() string { return string(s) }

// StatusTransition encodes a single permitted move between two account statuses.
//
// Fields:
//   - From / To              the source and destination statuses.
//   - RequiredPermission     the permission key the caller MUST hold (checked
//     by the service layer, not the domain).
//   - ValidationRequired     when true, Accounts.Validate() must return no errors
//     before the transition is allowed to proceed.
type StatusTransition struct {
	From               AccountStatus
	To                 AccountStatus
	RequiredPermission string
	ValidationRequired bool
}

// AllowedTransitions is the authoritative state machine for account status.
//
// Design rules:
//   - All service-layer transition logic MUST consult this table.
//   - Ad-hoc if/switch chains that duplicate this logic are forbidden.
//   - To add a new transition, append a single row here; no other code changes
//     are needed (CanTransitionTo and TransitionTo iterate this slice).
//   - Rows are grouped by source state for readability; within each group they
//     are ordered from least-privileged to most-privileged operation.
var AllowedTransitions = []StatusTransition{
	// ── Initial lifecycle ─────────────────────────────────────────────────────
	// A DRAFT account can be submitted for approval or activated directly
	// (the latter requires the higher "activate" permission).
	{AccountStatusDraft, AccountStatusPendingApproval, "finance.accounts.submit", true},
	{AccountStatusDraft, AccountStatusActive, "finance.accounts.activate", true},

	// ── Approval flow ─────────────────────────────────────────────────────────
	// An approver can approve (→ ACTIVE) or reject (→ DRAFT for rework).
	{AccountStatusPendingApproval, AccountStatusActive, "finance.accounts.approve", true},
	{AccountStatusPendingApproval, AccountStatusDraft, "finance.accounts.reject", false},

	// ── Normal operational transitions ────────────────────────────────────────
	// From ACTIVE, an operator can place a variety of holds. Validation is only
	// required for transitions that lock the account for a structural reason
	// (year-end), not for emergency holds (freeze, compliance).
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

	// ── Releasing holds ────────────────────────────────────────────────────────
	// Most holds resolve back to ACTIVE (requires re-validation to ensure the
	// account is still structurally sound after the hold period).
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

	// System maintenance resolves automatically; no re-validation needed.
	{AccountStatusSystemMaintenance, AccountStatusActive, "finance.accounts.activate", false},

	// DATA_ERROR can either be corrected and reactivated, or reset to DRAFT for
	// a full reconfiguration cycle.
	{AccountStatusDataError, AccountStatusActive, "finance.accounts.activate", true},
	{AccountStatusDataError, AccountStatusDraft, "finance.accounts.reset", true},

	// ── Terminal transitions ───────────────────────────────────────────────────
	// INACTIVE → ACTIVE: the account was dormant but is needed again.
	// INACTIVE → CLOSED: permanent closure; balance must be zero (enforced by
	//   CanBeDeactivated which the service calls before TransitionTo).
	{AccountStatusInactive, AccountStatusActive, "finance.accounts.activate", true},
	{AccountStatusInactive, AccountStatusSuspended, "finance.accounts.suspend", false},
	{AccountStatusInactive, AccountStatusClosed, "finance.accounts.close", true},

	// CLOSED → ARCHIVED: moves the account to cold storage. Cannot be undone.
	{AccountStatusClosed, AccountStatusArchived, "finance.accounts.archive", false},
}

// =============================================================================
// 3. Supporting enumerations
// =============================================================================

// RootType is the primary accounting classification of an account. It
// determines the normal balance (debit/credit), the financial statement section
// (balance sheet vs P&L), and aggregation behaviour in reporting hierarchies.
type RootType string

const (
	RootTypeAsset     RootType = "ASSET"
	RootTypeLiability RootType = "LIABILITY"
	RootTypeEquity    RootType = "EQUITY"
	RootTypeRevenue   RootType = "REVENUE"
	RootTypeExpense   RootType = "EXPENSE"
)

// rootTypeSet is the single source of truth for valid root types.
// IsValid() and ListRootTypes() both derive from this map.
// Adding a new root type here is sufficient; no switch statement needs updating.
//
// NOTE: GetNormalBalanceForRootType contains a separate switch that DOES need
// manual updating when a new root type is added, because the debit/credit
// assignment is a business decision that must be made explicitly.
var rootTypeSet = map[RootType]struct{}{
	RootTypeAsset:     {},
	RootTypeLiability: {},
	RootTypeEquity:    {},
	RootTypeRevenue:   {},
	RootTypeExpense:   {},
}

// IsValid returns true if r is a recognised root type constant.
func (r RootType) IsValid() bool {
	_, ok := rootTypeSet[r]
	return ok
}

// String implements fmt.Stringer.
func (r RootType) String() string { return string(r) }

// NormalBalance returns the expected conventional balance side for this root
// type. Assets and Expenses increase on the debit side; all others increase on
// the credit side. Used internally and exposed for display purposes.
func (r RootType) NormalBalance() NormalBalance {
	return GetNormalBalanceForRootType(r)
}

// ParseRootType converts a raw string into a RootType, returning an error if
// the value is not recognised. Prefer this over direct RootType(s) casting in
// service and HTTP layers.
func ParseRootType(s string) (RootType, error) {
	rt := RootType(s)
	if rt.IsValid() {
		return rt, nil
	}
	return "", fmt.Errorf("invalid root type %q; valid values: ASSET, LIABILITY, EQUITY, REVENUE, EXPENSE", s)
}

// ListRootTypes returns all valid root type constants in an arbitrary order.
// The slice is freshly allocated on each call; callers may sort or modify it.
func ListRootTypes() []RootType {
	types := make([]RootType, 0, len(rootTypeSet))
	for rt := range rootTypeSet {
		types = append(types, rt)
	}
	return types
}

// NormalBalance indicates which side of the double-entry equation increases an
// account's balance.
type NormalBalance string

const (
	// NormalBalanceDebit — the account increases with debits (ASSET, EXPENSE).
	NormalBalanceDebit NormalBalance = "DEBIT"

	// NormalBalanceCredit — the account increases with credits (LIABILITY, EQUITY, REVENUE).
	NormalBalanceCredit NormalBalance = "CREDIT"
)

// IsValid returns true if n is a recognised normal balance constant.
func (n NormalBalance) IsValid() bool {
	return n == NormalBalanceDebit || n == NormalBalanceCredit
}

// GetNormalBalanceForRootType returns the conventional normal balance for a
// given root type.
//
// Double-entry rule:
//   - ASSET and EXPENSE accounts have a debit normal balance: they increase
//     when debited and decrease when credited.
//   - LIABILITY, EQUITY, and REVENUE accounts have a credit normal balance:
//     they increase when credited and decrease when debited.
//
// If a new RootType is added to rootTypeSet, this function MUST be updated to
// assign its normal balance explicitly.
func GetNormalBalanceForRootType(rt RootType) NormalBalance {
	switch rt {
	case RootTypeAsset, RootTypeExpense:
		return NormalBalanceDebit
	default:
		// LIABILITY, EQUITY, REVENUE — and any future credit-normal type.
		return NormalBalanceCredit
	}
}

// ValidationStatus tracks whether an account has passed its most recent
// domain validation run. The repository updates this field after every
// call to Accounts.Validate().
type ValidationStatus string

const (
	// ValidationStatusValid — all validation checks passed; account may be activated.
	ValidationStatusValid ValidationStatus = "VALID"

	// ValidationStatusWarning — passed with non-blocking warnings.
	ValidationStatusWarning ValidationStatus = "WARNING"

	// ValidationStatusInvalid — one or more errors; account cannot be activated.
	ValidationStatusInvalid ValidationStatus = "INVALID"

	// ValidationStatusPending — not yet validated since last change.
	ValidationStatusPending ValidationStatus = "PENDING"
)

// =============================================================================
// 4. Accounts entity — the aggregate root
// =============================================================================

// Accounts is a single node in the chart of accounts (ledger tree).
//
// # Account code & hierarchy
//
// AccountCode is a fixed 8-digit numeric string encoding the account's position
// in the tree (see segment layout at the top of this file). Structural
// relationships are also maintained explicitly via ParentAccountID, AccountLevel,
// and Path so the hierarchy can be queried without parsing codes.
//
// Only leaf accounts (HasChildren == false) may accept direct journal postings.
// Parent accounts aggregate their children's balances; posting to them directly
// is a business rule violation enforced by CanAcceptTransactions().
//
// # IsLeaf — method, not field
//
// IsLeaf() is a pure derived method (return !HasChildren). It is NOT stored as
// a struct field to prevent the inconsistency bug where HasChildren and IsLeaf
// contradict each other. If you need a queryable column in the database, add a
// generated/computed column:
//
//	ALTER TABLE accounts
//	    ADD COLUMN is_leaf boolean GENERATED ALWAYS AS (NOT has_children) STORED;
//
// # Reporting placement
//
// An account's position on the P&L, balance sheet, or cash flow statement is
// NOT encoded here. That mapping lives in AccountMapping and is owned by
// ReportingGroup. This decoupling means EPRA, KRA, and internal report
// structures can each have their own mapping without touching ledger accounts.
//
// # Reconciliation & tax
//
// RequiresReconciliation, LastReconciledAt, TaxCode, and TaxRate are
// first-class fields because they are intrinsic properties of the ledger
// account itself, not of how it appears in a specific report.
type Accounts struct {
	ID       uuid.UUID  `json:"id"`
	TenantID uuid.UUID  `json:"tenant_id"`
	EntityID *uuid.UUID `json:"entity_id,omitempty"` // Multi-entity support; nil = tenant-wide

	// -------------------------------------------------------------------------
	// Identity
	// -------------------------------------------------------------------------

	// AccountCode is the 8-digit numeric identifier for this account.
	//
	// Segment layout (see top of file for full documentation):
	//   Digits 1-2: Root segment (10=Asset, 20=Liability, 30=Equity, 40=Revenue, 50=Expense)
	//   Digits 3-4: Category
	//   Digits 5-6: Sub-category
	//   Digits 7-8: Leaf detail
	//
	// The code is immutable after the first journal entry is posted. A
	// recode operation requires a dedicated service workflow.
	AccountCode string `json:"account_code" db:"account_code"`

	// AccountName is the human-readable display label used in reports and
	// account selectors. Maximum AccountNameMaxLen characters.
	AccountName string `json:"account_name"`

	// AccountDescription is an optional long-form explanation of the account's
	// purpose, posting rules, or regulatory reference. Maximum AccountDescriptionMaxLen.
	AccountDescription *string `json:"account_description,omitempty"`

	// -------------------------------------------------------------------------
	// Ledger hierarchy
	// -------------------------------------------------------------------------

	// ParentAccountID is the direct parent in the ledger tree.
	// Nil only for root accounts (AccountLevel == 0).
	ParentAccountID *uuid.UUID `json:"parent_account_id,omitempty"`

	// AccountLevel is the zero-based depth of this account in the tree.
	// Invariant (enforced by Validate): ParentAccountID == nil ↔ AccountLevel == 0.
	AccountLevel int32 `json:"account_level"`

	// Path is the materialised ancestor path maintained by the repository on
	// create and reparent. Use IsDescendantOf() for subtree membership tests;
	// never parse the path string directly in business logic.
	Path MaterialisedPath `json:"path"`

	// HasChildren is denormalised onto this entity so that IsLeaf() and
	// CanAcceptTransactions() can enforce the leaf-only posting rule without
	// an extra database round-trip. The repository is responsible for keeping
	// this consistent whenever a child account is created or deleted.
	HasChildren bool `json:"has_children"`

	// -------------------------------------------------------------------------
	// Grouping references
	// -------------------------------------------------------------------------

	// AccountGroupID optionally links this account to a legacy AccountGroup.
	// Prefer AccountMapping for new report structures.
	AccountGroupID *uuid.UUID `json:"account_group_id,omitempty"`

	// ControlAccountID references the control account that summarises this
	// account (e.g. an AR control account for a customer sub-ledger account).
	// An account that IS the control account has this nil and IsControlAccount true.
	ControlAccountID *uuid.UUID `json:"control_account_id,omitempty"`

	// -------------------------------------------------------------------------
	// Classification
	// -------------------------------------------------------------------------

	// RootType is the primary accounting classification (ASSET, LIABILITY, …).
	// Immutable after the first journal entry is posted; changes require a
	// reclassification workflow that zeros the balance and migrates mappings.
	RootType RootType `json:"root_type"`

	// AccountType is the secondary classification, e.g. "Current Asset",
	// "Operating Revenue". Maximum AccountTypeMaxLen characters.
	AccountType string `json:"account_type"`

	// AccountSubtype, AccountCategory, and SubCategory provide tertiary and
	// quaternary granularity used by industry-specific reports
	// (e.g. EPRA fuel grades, KRA tax mapping codes).
	AccountSubtype  *string `json:"account_subtype,omitempty"`
	AccountCategory *string `json:"account_category,omitempty"`
	SubCategory     *string `json:"sub_category,omitempty"`

	// -------------------------------------------------------------------------
	// Financial attributes
	// -------------------------------------------------------------------------

	// NormalBalance determines how journal entries increase or decrease this
	// account. Must equal GetNormalBalanceForRootType(RootType); enforced by
	// Validate(). Stored explicitly to avoid repeated derivation at query time.
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

	// CurrencyRevaluationRequired means the balance must be revalued at each
	// period-end using the closing exchange rate. Applies to monetary
	// foreign-currency accounts (receivables, payables, bank accounts).
	CurrencyRevaluationRequired bool `json:"currency_revaluation_required"`

	// -------------------------------------------------------------------------
	// Operational settings
	// -------------------------------------------------------------------------

	// IsActive is the fast-path operational toggle. Both Status == ACTIVE and
	// IsActive == true must hold for the account to accept transactions. The
	// two flags allow an account to be administratively locked (Status != ACTIVE)
	// while retaining the IsActive setting for when the lock is lifted.
	IsActive bool `json:"is_active"`

	// IsSystemAccount marks a protected account provisioned by the chart of
	// accounts seed. System accounts cannot be renamed, reclassified, or deleted
	// by regular users.
	IsSystemAccount bool `json:"is_system_account"`

	// AllowManualEntries controls whether human-authored journal lines may
	// reference this account. False for control accounts and automated
	// clearing accounts.
	AllowManualEntries bool `json:"allow_manual_entries"`

	// RequireReference mandates a non-empty reference string on every journal
	// line that posts to this account. Useful for bank reconciliation accounts.
	RequireReference bool `json:"require_reference"`

	// FinancialStatementLine is an optional label that ties this account to a
	// specific line in a statutory financial statement (e.g. "IFRS 15 Revenue").
	// Nil means no statutory line mapping.
	FinancialStatementLine *string `json:"financial_statement_line,omitempty"`

	// ReportOrder controls the sort position of this account within its
	// financial statement line. Lower values appear first.
	ReportOrder int32 `json:"report_order"`

	// -------------------------------------------------------------------------
	// Reconciliation
	// -------------------------------------------------------------------------

	// RequiresReconciliation flags accounts (typically bank and cash) that must
	// be reconciled against external statements each period.
	RequiresReconciliation bool `json:"requires_reconciliation"`

	// LastReconciledAt records the timestamp of the most recent completed
	// reconciliation. Nil means never reconciled.
	LastReconciledAt *time.Time `json:"last_reconciled_at,omitempty"`

	// -------------------------------------------------------------------------
	// Tax
	// -------------------------------------------------------------------------

	// TaxCode is the applicable tax code for transactions on this account
	// (e.g. "VAT16", "WHT5"). Nil means no default tax applies.
	TaxCode *string `json:"tax_code,omitempty"`

	// TaxRate is the default tax rate percentage (0–100). Nil when TaxCode is nil.
	// Stored as a decimal to avoid floating-point precision issues.
	TaxRate *decimal.Decimal `json:"tax_rate,omitempty"`

	// -------------------------------------------------------------------------
	// Balance (denormalised cache — updated by the journal engine)
	// -------------------------------------------------------------------------

	// CurrentBalance is the running balance maintained by the journal engine.
	//   Debit-normal accounts: positive value = net debit position.
	//   Credit-normal accounts: positive value = net credit position.
	// Use GetEffectiveBalance() for sign-adjusted presentation values.
	CurrentBalance decimal.Decimal `json:"current_balance"`

	// YTDBalance is the year-to-date movement, reset at each financial year open.
	YTDBalance decimal.Decimal `json:"ytd_balance"`

	// LastTransactionDate is the posting date of the most recent journal entry.
	// Nil means no transactions have ever been posted to this account.
	LastTransactionDate *time.Time `json:"last_transaction_date,omitempty"`

	// -------------------------------------------------------------------------
	// Budgeting
	// -------------------------------------------------------------------------

	// IsBudgetable allows budget lines to be created for this account.
	// Typically true for revenue and expense accounts, false for balance sheet.
	IsBudgetable bool `json:"is_budgetable"`

	// BudgetVarianceThreshold is the percentage deviation from budget that
	// triggers an alert. Range [0, 100]. Zero disables budget alerts.
	BudgetVarianceThreshold decimal.Decimal `json:"budget_variance_threshold"`

	// -------------------------------------------------------------------------
	// Lifecycle & validation
	// -------------------------------------------------------------------------

	// Status is the current lifecycle state. All transitions must use
	// TransitionTo() which consults AllowedTransitions.
	Status AccountStatus `json:"status"`

	// ValidationStatus tracks the result of the last Validate() call.
	// The repository updates this field; do not set it directly in service code.
	ValidationStatus ValidationStatus `json:"validation_status"`

	// ValidationErrors holds the errors from the most recent Validate() call.
	// Empty when ValidationStatus == VALID.
	ValidationErrors []ValidationError `json:"validation_errors,omitempty"`

	// -------------------------------------------------------------------------
	// Extensions
	// -------------------------------------------------------------------------

	// AccountAttributes holds tenant-defined key-value metadata for
	// industry-specific extensions (e.g. EPRA product codes, KRA tax mappings).
	// Serialised as JSONB in the database; use MarshalAccountAttributes and
	// UnmarshalAccountAttributes for explicit JSONB column handling.
	AccountAttributes map[string]any `json:"account_attributes,omitempty"`

	// -------------------------------------------------------------------------
	// Audit trail
	// -------------------------------------------------------------------------

	// Version is an optimistic-locking counter. The repository increments it on
	// every UPDATE and rejects writes where the caller's version does not match
	// the stored version (returns ErrConflict). Always pass Version in update
	// requests so stale writes are caught.
	Version   int32      `json:"version"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	CreatedBy uuid.UUID  `json:"created_by"`
	UpdatedBy *uuid.UUID `json:"updated_by,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"` // Soft-delete; nil = not deleted
}

// =============================================================================
// 5. State machine methods
// =============================================================================

// CanTransitionTo reports whether a move from the current Status to newStatus
// is listed in AllowedTransitions. It is a read-only check; it does NOT
// verify permissions or run validation.
//
// Use TransitionTo when you need the full StatusTransition (permission key,
// validation flag) to enforce the transition.
func (a *Accounts) CanTransitionTo(newStatus AccountStatus) bool {
	for _, t := range AllowedTransitions {
		if t.From == a.Status && t.To == newStatus {
			return true
		}
	}
	return false
}

// TransitionTo returns the matching StatusTransition for the service layer to
// enforce the required permission and run validation if needed.
//
// Returns an error when the (current → newStatus) pair is not in AllowedTransitions.
// The caller is responsible for:
//  1. Verifying the caller holds t.RequiredPermission.
//  2. Running Accounts.Validate() and checking for errors when t.ValidationRequired is true.
//  3. Updating a.Status to newStatus and persisting via the repository.
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
// 6. Behaviour methods
// =============================================================================

// IsLeaf returns true when this account has no children and can therefore
// accept direct journal postings.
//
// This is a derived method rather than a stored field to eliminate the
// HasChildren / IsLeaf inconsistency bug. The repository maintains HasChildren;
// this method derives IsLeaf from it at zero cost.
//
// For database queries, add a generated column:
//
//	ALTER TABLE accounts
//	    ADD COLUMN is_leaf boolean GENERATED ALWAYS AS (NOT has_children) STORED;
func (a *Accounts) IsLeaf() bool {
	return !a.HasChildren
}

// CanAcceptTransactions returns true when a journal engine may post a line to
// this account.
//
// All five conditions must hold simultaneously:
//  1. Status == ACTIVE — not in any hold, close, or draft state.
//  2. IsActive == true — not administratively toggled off.
//  3. IsLeaf() == true — only leaf accounts accept direct postings.
//  4. DeletedAt == nil — not soft-deleted.
//  5. ValidationStatus == VALID — passed last validation run.
//
// When this returns false, call GetTransactionRestrictions() for a human-readable
// explanation suitable for error responses.
func (a *Accounts) CanAcceptTransactions() bool {
	return a.Status == AccountStatusActive &&
		a.IsActive &&
		a.IsLeaf() &&
		a.DeletedAt == nil &&
		a.ValidationStatus == ValidationStatusValid
}

// CanAcceptManualEntries returns true when a human-authored journal entry may
// reference this account (as opposed to system-generated postings only).
//
// This is a stricter superset of CanAcceptTransactions: the account must also
// have AllowManualEntries == true and must not be a control account.
func (a *Accounts) CanAcceptManualEntries() bool {
	return a.CanAcceptTransactions() &&
		a.AllowManualEntries &&
		!a.IsControlAccount
}

// CanBeDeactivated returns true when it is safe to move the account to INACTIVE.
//
// Three conditions must all hold:
//  1. The current balance is zero (no outstanding position to carry).
//  2. The account is not a protected system account.
//  3. The current status permits a transition to INACTIVE (consults AllowedTransitions).
//
// Condition 3 prevents DRAFT, CLOSED, or ARCHIVED accounts from appearing
// deactivatable simply because their balance happens to be zero.
func (a *Accounts) CanBeDeactivated() bool {
	if a.IsSystemAccount {
		return false
	}
	if !a.CurrentBalance.IsZero() {
		return false
	}
	// Delegate the "is this a valid source state?" question to the state machine
	// so the logic lives in exactly one place.
	return a.CanTransitionTo(AccountStatusInactive)
}

// CanBeDeleted returns true when soft-deletion is permitted.
//
// An account may not be deleted if:
//   - It is a system account (provisioned by the chart of accounts seed).
//   - It has ever had transactions posted (LastTransactionDate is non-nil).
//
// A soft-deleted account is invisible to posting workflows but retained for
// audit and historical reporting.
func (a *Accounts) CanBeDeleted() bool {
	return !a.IsSystemAccount && a.LastTransactionDate == nil
}

// IsDeleted returns true when the account has been soft-deleted.
func (a *Accounts) IsDeleted() bool { return a.DeletedAt != nil }

// IsDescendantOf reports whether this account is a descendant of parentID
// using the materialised Path. Prefer this over comparing AccountCode prefixes,
// as the path is maintained transactionally by the repository.
func (a *Accounts) IsDescendantOf(parentID uuid.UUID) bool {
	return a.Path.IsDescendantOf(parentID)
}

// GetEffectiveBalance returns the balance sign-adjusted for financial statement
// presentation.
//
// Convention:
//   - Debit-normal accounts (ASSET, EXPENSE): returned as-is; a positive value
//     means a net asset/expense position.
//   - Credit-normal accounts (LIABILITY, EQUITY, REVENUE): negated; a positive
//     value means a net liability/equity/revenue position on the credit side.
//
// Use this method in report renderers and UI display; use CurrentBalance
// directly in journal engine arithmetic.
func (a *Accounts) GetEffectiveBalance() decimal.Decimal {
	if a.NormalBalance == NormalBalanceDebit {
		return a.CurrentBalance
	}
	return a.CurrentBalance.Neg()
}

// GetTransactionRestrictions returns a slice of human-readable messages
// explaining why a journal posting would be rejected for this account.
//
// Returns an empty slice when the account can accept unrestricted transactions.
// The messages are suitable for inclusion in API error responses and UI tooltips.
func (a *Accounts) GetTransactionRestrictions() []string {
	var msgs []string

	// Status-driven restrictions.
	switch a.Status {
	case AccountStatusRestricted:
		msgs = append(msgs, "manual entries require approval while the account is restricted")
	case AccountStatusFrozen:
		msgs = append(msgs, "account is frozen — no new transactions of any kind are permitted")
	case AccountStatusYearEndProcessing:
		msgs = append(msgs, "only period-closing entries are permitted during year-end processing")
	case AccountStatusAuditLock:
		msgs = append(msgs, "account is read-only while an external audit is in progress")
	case AccountStatusComplianceHold:
		msgs = append(msgs, "account is under a regulatory compliance hold — no transactions")
	case AccountStatusSystemMaintenance:
		msgs = append(msgs, "account is temporarily unavailable during system maintenance")
	case AccountStatusDraft, AccountStatusPendingApproval:
		msgs = append(msgs, fmt.Sprintf("account has not been activated (current status: %s)", a.Status))
	case AccountStatusInactive, AccountStatusSuspended:
		msgs = append(msgs, fmt.Sprintf("account is not active (current status: %s)", a.Status))
	case AccountStatusClosed, AccountStatusArchived:
		msgs = append(msgs, fmt.Sprintf("account is permanently closed (status: %s)", a.Status))
	}

	// Structural restrictions (independent of status).
	if !a.AllowManualEntries {
		msgs = append(msgs, "account accepts system-generated entries only; manual postings are disabled")
	}
	if a.IsControlAccount {
		msgs = append(msgs, "control accounts do not accept direct manual postings; post to a child account")
	}
	if !a.IsLeaf() {
		msgs = append(msgs, "only leaf accounts accept direct postings; post to a child account instead")
	}
	if a.RequireReference {
		msgs = append(msgs, "a non-empty reference number is mandatory for all entries on this account")
	}
	if a.ValidationStatus != ValidationStatusValid {
		msgs = append(msgs, fmt.Sprintf(
			"account failed validation (%s); resolve errors before posting", a.ValidationStatus,
		))
	}
	return msgs
}

// RequiresApproval returns true when the account is awaiting administrative
// action and should be surfaced in approval queues or dashboards.
func (a *Accounts) RequiresApproval() bool {
	return a.Status == AccountStatusPendingApproval ||
		a.Status == AccountStatusUnderReview
}

// MarshalAccountAttributes serialises AccountAttributes to a JSON byte slice.
// Call this when writing the attributes to a dedicated JSONB column rather than
// relying on the struct-level json.Marshal of the whole entity.
func (a *Accounts) MarshalAccountAttributes() ([]byte, error) {
	if a.AccountAttributes == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(a.AccountAttributes)
}

// UnmarshalAccountAttributes deserialises AccountAttributes from a JSON byte
// slice read from a dedicated JSONB column.
func (a *Accounts) UnmarshalAccountAttributes(data []byte) error {
	if len(data) == 0 {
		a.AccountAttributes = make(map[string]any)
		return nil
	}
	return json.Unmarshal(data, &a.AccountAttributes)
}

// =============================================================================
// 7. Shared validation helpers (package-private)
// =============================================================================
//
// These helpers are extracted from Accounts.Validate() and
// CreateAccountRequest.Validate() to prevent copy-paste drift. Any change to a
// validation rule need only be made once here.
//
// Naming convention: validate<FieldOrConcept>(args...) []ValidationError

// validateAccountCode checks that code conforms to the 8-digit numeric format.
// Returns one or more ValidationErrors; returns nil when the code is valid.
func validateAccountCode(code string) []ValidationError {
	var errs []ValidationError
	switch {
	case strings.TrimSpace(code) == "":
		errs = append(errs, ValidationError{
			Field:    "account_code",
			Message:  "account code is required",
			Code:     "REQUIRED_FIELD",
			Severity: ValidationSeverityError,
		})
	case len(code) != AccountCodeLen:
		// Check length before the regex so the error message is more specific.
		errs = append(errs, ValidationError{
			Field: "account_code",
			Message: fmt.Sprintf(
				"account code must be exactly %d digits (e.g. 10010101)", AccountCodeLen,
			),
			Code:     "INVALID_LENGTH",
			Severity: ValidationSeverityError,
		})
	case !accountCodePattern.MatchString(code):
		// Catches non-digit characters; length was already verified above.
		errs = append(errs, ValidationError{
			Field:    "account_code",
			Message:  "account code must contain digits only (0-9)",
			Code:     "INVALID_FORMAT",
			Severity: ValidationSeverityError,
		})
	}
	return errs
}

// validateAccountName checks that name is non-empty and within the length cap.
func validateAccountName(name string) []ValidationError {
	var errs []ValidationError
	switch {
	case strings.TrimSpace(name) == "":
		errs = append(errs, ValidationError{
			Field:    "account_name",
			Message:  "account name is required",
			Code:     "REQUIRED_FIELD",
			Severity: ValidationSeverityError,
		})
	case len(name) > AccountNameMaxLen:
		errs = append(errs, ValidationError{
			Field: "account_name",
			Message: fmt.Sprintf(
				"account name must be %d characters or less", AccountNameMaxLen,
			),
			Code:     "MAX_LENGTH_EXCEEDED",
			Severity: ValidationSeverityError,
		})
	}
	return errs
}

// validateNormalBalanceConsistency checks that nb is the conventional balance
// for rt. Both values must already be individually valid (IsValid() == true)
// before calling this; invalid individual values are checked separately.
func validateNormalBalanceConsistency(rt RootType, nb NormalBalance) []ValidationError {
	if !rt.IsValid() || !nb.IsValid() {
		// Individual field validators catch these; nothing to do here.
		return nil
	}
	expected := GetNormalBalanceForRootType(rt)
	if nb != expected {
		return []ValidationError{{
			Field: "normal_balance",
			Message: fmt.Sprintf(
				"normal balance must be %s for root type %s (got %s)",
				expected, rt, nb,
			),
			Code:     "INCONSISTENT_VALUE",
			Severity: ValidationSeverityError,
		}}
	}
	return nil
}

// validateTaxConsistency checks that TaxCode is present when TaxRate is set and
// that the rate value is within the valid [0, 100] range.
func validateTaxConsistency(taxCode *string, taxRate *decimal.Decimal) []ValidationError {
	var errs []ValidationError
	if taxRate == nil {
		return errs // Nothing to validate when no rate is set.
	}
	if taxCode == nil {
		errs = append(errs, ValidationError{
			Field:    "tax_code",
			Message:  "tax_code is required when tax_rate is set",
			Code:     "REQUIRED_FIELD",
			Severity: ValidationSeverityError,
		})
	}
	zero, hundred := decimal.Zero, decimal.NewFromInt(100)
	if taxRate.LessThan(zero) || taxRate.GreaterThan(hundred) {
		errs = append(errs, ValidationError{
			Field:    "tax_rate",
			Message:  "tax rate must be between 0 and 100",
			Code:     "OUT_OF_RANGE",
			Severity: ValidationSeverityError,
		})
	}
	return errs
}

// validateBudgetVarianceThreshold checks that the threshold is within [0, 100].
func validateBudgetVarianceThreshold(threshold decimal.Decimal) []ValidationError {
	zero, hundred := decimal.Zero, decimal.NewFromInt(100)
	if threshold.LessThan(zero) || threshold.GreaterThan(hundred) {
		return []ValidationError{{
			Field:    "budget_variance_threshold",
			Message:  "budget variance threshold must be between 0 and 100",
			Code:     "OUT_OF_RANGE",
			Severity: ValidationSeverityError,
		}}
	}
	return nil
}

// validateControlAccountRules checks the mutual-exclusion constraints between
// IsControlAccount, AllowManualEntries, and ControlAccountID.
func validateControlAccountRules(isControl bool, allowManual bool, controlID *uuid.UUID) []ValidationError {
	var errs []ValidationError
	if isControl && allowManual {
		errs = append(errs, ValidationError{
			Field:    "allow_manual_entries",
			Message:  "control accounts must not allow manual entries; postings flow from the sub-ledger automatically",
			Code:     "BUSINESS_RULE_VIOLATION",
			Severity: ValidationSeverityError,
		})
	}
	if isControl && controlID != nil {
		errs = append(errs, ValidationError{
			Field:    "control_account_id",
			Message:  "a control account cannot itself be controlled by another account",
			Code:     "BUSINESS_RULE_VIOLATION",
			Severity: ValidationSeverityError,
		})
	}
	return errs
}

// =============================================================================
// 8. Validate / ValidateWithContext
// =============================================================================

// Validate performs all entity-level consistency checks that require only the
// data stored on this Accounts instance. Cross-entity rules (parent existence,
// RootType homogeneity, control account child counts) live in
// ValidateWithContext which accepts externally fetched data.
//
// Call order in service layer:
//  1. errs := account.Validate()
//  2. if len(errs) == 0 { errs = account.ValidateWithContext(parent, hasChildren, hasTxns) }
//  3. Persist account.ValidationStatus based on whether errs is empty.
func (a *Accounts) Validate() []ValidationError {
	var errs []ValidationError

	// -- Tenant ---------------------------------------------------------------
	if a.TenantID == uuid.Nil {
		errs = append(errs, ValidationError{
			Field:    "tenant_id",
			Message:  "tenant_id is required",
			Code:     "REQUIRED_FIELD",
			Severity: ValidationSeverityError,
		})
	}

	// -- Account code (delegates to shared helper) ----------------------------
	errs = append(errs, validateAccountCode(a.AccountCode)...)

	// -- Account name (delegates to shared helper) ----------------------------
	errs = append(errs, validateAccountName(a.AccountName)...)

	// -- Description length ---------------------------------------------------
	if a.AccountDescription != nil && len(*a.AccountDescription) > AccountDescriptionMaxLen {
		errs = append(errs, ValidationError{
			Field: "account_description",
			Message: fmt.Sprintf(
				"account description must be %d characters or less", AccountDescriptionMaxLen,
			),
			Code:     "MAX_LENGTH_EXCEEDED",
			Severity: ValidationSeverityError,
		})
	}

	// -- Classification -------------------------------------------------------
	if !a.RootType.IsValid() {
		errs = append(errs, ValidationError{
			Field:    "root_type",
			Message:  fmt.Sprintf("invalid root type %q; valid values: ASSET, LIABILITY, EQUITY, REVENUE, EXPENSE", a.RootType),
			Code:     "INVALID_VALUE",
			Severity: ValidationSeverityError,
		})
	}
	if !a.NormalBalance.IsValid() {
		errs = append(errs, ValidationError{
			Field:    "normal_balance",
			Message:  fmt.Sprintf("invalid normal balance %q; valid values: DEBIT, CREDIT", a.NormalBalance),
			Code:     "INVALID_VALUE",
			Severity: ValidationSeverityError,
		})
	}
	// NormalBalance must match the conventional balance for the RootType.
	errs = append(errs, validateNormalBalanceConsistency(a.RootType, a.NormalBalance)...)

	// -- Status ---------------------------------------------------------------
	if !a.Status.IsValid() {
		errs = append(errs, ValidationError{
			Field:    "status",
			Message:  fmt.Sprintf("invalid account status %q", a.Status),
			Code:     "INVALID_VALUE",
			Severity: ValidationSeverityError,
		})
	}

	// -- Hierarchy consistency ------------------------------------------------
	// ParentAccountID == nil ↔ AccountLevel == 0.
	if a.ParentAccountID == nil && a.AccountLevel != 0 {
		errs = append(errs, ValidationError{
			Field:    "account_level",
			Message:  "a root account (no parent) must have level 0",
			Code:     "INCONSISTENT_VALUE",
			Severity: ValidationSeverityError,
		})
	}
	if a.ParentAccountID != nil && a.AccountLevel <= 0 {
		errs = append(errs, ValidationError{
			Field:    "account_level",
			Message:  "a child account (has parent) must have level > 0",
			Code:     "INCONSISTENT_VALUE",
			Severity: ValidationSeverityError,
		})
	}

	// Materialised path format check (content correctness is in ValidateWithContext).
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
				"currency code must be exactly %d characters (ISO 4217), got %d",
				CurrencyCodeLen, len(*a.CurrencyCode),
			),
			Code:     "INVALID_FORMAT",
			Severity: ValidationSeverityError,
		})
	}

	// -- Control account rules ------------------------------------------------
	errs = append(errs, validateControlAccountRules(
		a.IsControlAccount, a.AllowManualEntries, a.ControlAccountID,
	)...)

	// -- Tax consistency ------------------------------------------------------
	errs = append(errs, validateTaxConsistency(a.TaxCode, a.TaxRate)...)

	// -- Budget variance threshold --------------------------------------------
	errs = append(errs, validateBudgetVarianceThreshold(a.BudgetVarianceThreshold)...)

	return errs
}

// ValidateWithContext performs cross-entity business rule checks that require
// data fetched from outside this entity. Call this after Validate() to avoid
// redundant work when the entity itself has errors.
//
// Parameters:
//   - parentAccount: the resolved parent Accounts, or nil for a root account.
//   - hasChildren:   true when child account rows exist in the repository.
//   - hasTransactions: true when any posted journal lines reference this account.
func (a *Accounts) ValidateWithContext(
	parentAccount *Accounts,
	hasChildren bool,
	hasTransactions bool,
) []ValidationError {
	var errs []ValidationError

	// -- Hierarchy homogeneity ------------------------------------------------
	// A child must share its parent's RootType. Mixing types in a subtree
	// (e.g. an ASSET child under a REVENUE parent) is a structural error that
	// would corrupt roll-up balances and financial statement placements.
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
		// AccountLevel must be exactly one more than the parent's level.
		if a.AccountLevel != parentAccount.AccountLevel+1 {
			errs = append(errs, ValidationError{
				Field: "account_level",
				Message: fmt.Sprintf(
					"account level must be parent level + 1 (parent is %d, this account is %d)",
					parentAccount.AccountLevel, a.AccountLevel,
				),
				Code:     "INVALID_HIERARCHY_LEVEL",
				Severity: ValidationSeverityError,
			})
		}
	}

	// -- Control account child requirement ------------------------------------
	// A control account without any sub-ledger children cannot summarise anything.
	// This check is deferred to ValidateWithContext because the repository must
	// confirm whether children exist.
	if a.IsControlAccount && !hasChildren {
		errs = append(errs, ValidationError{
			Field:    "is_control_account",
			Message:  "a control account must have at least one child sub-ledger account",
			Code:     "CONTROL_ACCOUNT_NO_CHILDREN",
			Severity: ValidationSeverityError,
		})
	}

	// -- Deactivation with outstanding balance --------------------------------
	// Deactivating an account that carries a non-zero balance would leave the
	// trial balance out of balance. The service layer should call CanBeDeactivated()
	// first, but this provides a second line of defence.
	if !a.IsActive && hasTransactions && !a.CurrentBalance.IsZero() {
		errs = append(errs, ValidationError{
			Field:    "is_active",
			Message:  "cannot deactivate an account that has a non-zero balance",
			Code:     "CANNOT_DEACTIVATE_WITH_BALANCE",
			Severity: ValidationSeverityError,
		})
	}

	return errs
}

// =============================================================================
// 9. Request DTOs
// =============================================================================

// CreateAccountRequest is the inbound payload for creating a new ledger account.
//
// Fields NOT accepted here (derived or set by the service layer):
//   - AccountLevel and Path — derived from ParentAccountID.
//   - Status — always initialised to DRAFT; callers use TransitionTo to advance.
//   - NormalBalance — can be provided, but is also derivable from RootType.
//   - CurrentBalance / YTDBalance — always initialised to 0.
//   - Version — set to 1 on create.
//
// AccountCode must be an 8-digit numeric string following the segment convention
// described at the top of this file.
type CreateAccountRequest struct {
	EntityID *uuid.UUID `json:"entity_id,omitempty"`

	// 8-digit numeric code; see AccountCodeSegments for segment layout.
	AccountCode        string  `json:"account_code"                      validate:"required,len=8"`
	AccountName        string  `json:"account_name"                      validate:"required,max=200"`
	AccountDescription *string `json:"account_description,omitempty"     validate:"omitempty,max=1000"`

	// Hierarchy
	ParentAccountID *uuid.UUID `json:"parent_account_id,omitempty"`
	AccountGroupID  *uuid.UUID `json:"account_group_id,omitempty"`

	// Classification
	RootType        RootType `json:"root_type"                         validate:"required"`
	AccountType     string   `json:"account_type"                      validate:"required,max=100"`
	AccountSubtype  *string  `json:"account_subtype,omitempty"         validate:"omitempty,max=100"`
	AccountCategory *string  `json:"account_category,omitempty"        validate:"omitempty,max=100"`
	SubCategory     *string  `json:"sub_category,omitempty"            validate:"omitempty,max=100"`

	// NormalBalance may be omitted; the service derives it from RootType if nil.
	// Include it if the caller wants an explicit consistency check.
	NormalBalance NormalBalance `json:"normal_balance"                    validate:"required"`

	// Control account
	IsControlAccount bool       `json:"is_control_account"`
	ControlAccountID *uuid.UUID `json:"control_account_id,omitempty"`

	// Currency
	CurrencyCode                *string `json:"currency_code,omitempty"           validate:"omitempty,len=3"`
	IsMultiCurrency             bool    `json:"is_multi_currency"`
	CurrencyRevaluationRequired bool    `json:"currency_revaluation_required"`

	// Operational
	IsActive           bool `json:"is_active"`
	AllowManualEntries bool `json:"allow_manual_entries"`
	RequireReference   bool `json:"require_reference"`

	// Reconciliation & tax
	RequiresReconciliation bool             `json:"requires_reconciliation"`
	TaxCode                *string          `json:"tax_code,omitempty"`
	TaxRate                *decimal.Decimal `json:"tax_rate,omitempty"                validate:"omitempty,min=0,max=100"`

	// Budgeting
	IsBudgetable            bool            `json:"is_budgetable"`
	BudgetVarianceThreshold decimal.Decimal `json:"budget_variance_threshold"         validate:"min=0,max=100"`

	// Tenant-defined extension metadata.
	AccountAttributes map[string]any `json:"account_attributes,omitempty"`
}

// Validate checks request-level consistency using the shared helpers.
// The service layer must also call ValidateWithContext on the resulting entity
// after resolving the parent account.
func (r *CreateAccountRequest) Validate() []ValidationError {
	var errs []ValidationError

	errs = append(errs, validateAccountCode(r.AccountCode)...)
	errs = append(errs, validateAccountName(r.AccountName)...)

	if !r.RootType.IsValid() {
		errs = append(errs, ValidationError{
			Field:    "root_type",
			Message:  fmt.Sprintf("invalid root type %q; valid values: ASSET, LIABILITY, EQUITY, REVENUE, EXPENSE", r.RootType),
			Code:     "INVALID_VALUE",
			Severity: ValidationSeverityError,
		})
	}
	if !r.NormalBalance.IsValid() {
		errs = append(errs, ValidationError{
			Field:    "normal_balance",
			Message:  fmt.Sprintf("invalid normal balance %q; valid values: DEBIT, CREDIT", r.NormalBalance),
			Code:     "INVALID_VALUE",
			Severity: ValidationSeverityError,
		})
	}

	errs = append(errs, validateNormalBalanceConsistency(r.RootType, r.NormalBalance)...)
	errs = append(errs, validateControlAccountRules(r.IsControlAccount, r.AllowManualEntries, r.ControlAccountID)...)
	errs = append(errs, validateTaxConsistency(r.TaxCode, r.TaxRate)...)
	errs = append(errs, validateBudgetVarianceThreshold(r.BudgetVarianceThreshold)...)

	return errs
}

// UpdateAccountRequest is the partial-update payload.
//
// Intentionally excluded fields — each requires a dedicated service operation:
//   - AccountCode         — needs uniqueness re-check + journal audit trail entry.
//   - ParentAccountID     — needs full subtree re-path (Reparent operation).
//   - RootType            — reclassification must zero the balance and migrate all mappings.
//   - NormalBalance       — follows RootType; cannot be changed independently.
//
// All pointer fields are optional; a nil value means "do not update this field".
// Version must match the stored optimistic-lock counter or the update is rejected.
type UpdateAccountRequest struct {
	// Version must match Accounts.Version; guards against lost-update anomalies.
	Version int32 `json:"version" validate:"required"`

	AccountName        *string `json:"account_name,omitempty"            validate:"omitempty,max=200"`
	AccountDescription *string `json:"account_description,omitempty"     validate:"omitempty,max=1000"`
	AccountType        *string `json:"account_type,omitempty"            validate:"omitempty,max=100"`
	AccountSubtype     *string `json:"account_subtype,omitempty"         validate:"omitempty,max=100"`
	AccountCategory    *string `json:"account_category,omitempty"        validate:"omitempty,max=100"`
	SubCategory        *string `json:"sub_category,omitempty"            validate:"omitempty,max=100"`

	AccountGroupID *uuid.UUID `json:"account_group_id,omitempty"`

	IsActive           *bool `json:"is_active,omitempty"`
	AllowManualEntries *bool `json:"allow_manual_entries,omitempty"`
	RequireReference   *bool `json:"require_reference,omitempty"`

	RequiresReconciliation *bool            `json:"requires_reconciliation,omitempty"`
	TaxCode                *string          `json:"tax_code,omitempty"`
	TaxRate                *decimal.Decimal `json:"tax_rate,omitempty"                validate:"omitempty,min=0,max=100"`

	IsBudgetable            *bool            `json:"is_budgetable,omitempty"`
	BudgetVarianceThreshold *decimal.Decimal `json:"budget_variance_threshold,omitempty" validate:"omitempty,min=0,max=100"`

	AccountAttributes map[string]any `json:"account_attributes,omitempty"`
}

// Validate checks the update request for internal consistency.
// Called by the service layer before applying changes to the entity.
func (r *UpdateAccountRequest) Validate() []ValidationError {
	var errs []ValidationError

	if r.AccountName != nil {
		errs = append(errs, validateAccountName(*r.AccountName)...)
	}
	if r.AccountDescription != nil && len(*r.AccountDescription) > AccountDescriptionMaxLen {
		errs = append(errs, ValidationError{
			Field: "account_description",
			Message: fmt.Sprintf(
				"account description must be %d characters or less", AccountDescriptionMaxLen,
			),
			Code:     "MAX_LENGTH_EXCEEDED",
			Severity: ValidationSeverityError,
		})
	}

	// Tax consistency: only validate when at least one tax field is provided.
	if r.TaxRate != nil {
		errs = append(errs, validateTaxConsistency(r.TaxCode, r.TaxRate)...)
	}

	if r.BudgetVarianceThreshold != nil {
		errs = append(errs, validateBudgetVarianceThreshold(*r.BudgetVarianceThreshold)...)
	}

	return errs
}

// =============================================================================
// 10. Read-side projections
// =============================================================================

// AccountBalance captures an account's debit/credit position at a specific
// point in time, computed by the reporting engine from posted journal entries.
// This is a read-only projection; it is never persisted directly.
type AccountBalance struct {
	AccountID    uuid.UUID       `json:"account_id"`
	TotalDebits  decimal.Decimal `json:"total_debits"`
	TotalCredits decimal.Decimal `json:"total_credits"`
	// NetBalance = TotalDebits - TotalCredits (raw arithmetic, not sign-adjusted).
	// Use GetEffectiveBalance on the Accounts entity for sign-adjusted values.
	NetBalance decimal.Decimal `json:"net_balance"`
	AsOfDate   time.Time       `json:"as_of_date"`
}

// TrialBalanceEntry is a single line in the trial balance report, produced by
// the reporting engine. It carries the codes and names needed for a
// self-contained report row without joining back to the accounts table.
type TrialBalanceEntry struct {
	AccountID     uuid.UUID       `json:"account_id"`
	AccountCode   string          `json:"account_code"`
	AccountName   string          `json:"account_name"`
	RootType      RootType        `json:"root_type"`
	NormalBalance NormalBalance   `json:"normal_balance"`
	TotalDebits   decimal.Decimal `json:"total_debits"`
	TotalCredits  decimal.Decimal `json:"total_credits"`
	// NetBalance = TotalDebits - TotalCredits (raw; not sign-adjusted for presentation).
	NetBalance decimal.Decimal `json:"net_balance"`
}

// =============================================================================
// 11. AccountFilter
// =============================================================================

// AccountFilter is the query predicate for listing accounts.
// All fields are optional; a nil field means "no constraint on this dimension".
type AccountFilter struct {
	EntityID    *uuid.UUID     `json:"entity_id,omitempty"`
	RootType    *RootType      `json:"root_type,omitempty"`
	AccountType *string        `json:"account_type,omitempty"`
	Status      *AccountStatus `json:"status,omitempty"`
	ParentID    *uuid.UUID     `json:"parent_account_id,omitempty"`
	IsActive    *bool          `json:"is_active,omitempty"`

	// IsLeaf filters to leaf-only (true) or parent-only (false) accounts.
	// Nil means return both. Maps to the generated is_leaf column in the DB.
	IsLeaf *bool `json:"is_leaf,omitempty"`

	// NonZeroOnly omits accounts whose CurrentBalance is exactly zero.
	NonZeroOnly *bool `json:"non_zero_only,omitempty"`

	// SearchQuery performs a case-insensitive substring match on AccountCode
	// and AccountName.
	SearchQuery *string `json:"search_query,omitempty"`

	// Hierarchy: MaxLevel caps the depth returned; IncludeChildren expands a
	// single ParentID query to include all descendants (requires a recursive
	// CTE or materialised-path LIKE query in the repository).
	MaxLevel        *int32 `json:"max_level,omitempty"`
	IncludeChildren bool   `json:"include_children"`

	// Pagination (zero values mean "no limit / start from beginning").
	Limit  *int `json:"limit,omitempty"`
	Offset *int `json:"offset,omitempty"`

	// Sorting: SortBy is the field name; SortOrder is "asc" or "desc".
	SortBy    *string `json:"sort_by,omitempty"`
	SortOrder *string `json:"sort_order,omitempty"`
}

// Validate checks that the filter's values are individually self-consistent.
// Returns errors for invalid enum values or negative pagination parameters.
func (f *AccountFilter) Validate() []ValidationError {
	var errs []ValidationError

	if f.RootType != nil && !f.RootType.IsValid() {
		errs = append(errs, ValidationError{
			Field:    "root_type",
			Message:  fmt.Sprintf("invalid root type %q", *f.RootType),
			Code:     "INVALID_VALUE",
			Severity: ValidationSeverityError,
		})
	}
	if f.Status != nil && !f.Status.IsValid() {
		errs = append(errs, ValidationError{
			Field:    "status",
			Message:  fmt.Sprintf("invalid account status %q", *f.Status),
			Code:     "INVALID_VALUE",
			Severity: ValidationSeverityError,
		})
	}
	if f.Limit != nil && *f.Limit < 0 {
		errs = append(errs, ValidationError{
			Field:    "limit",
			Message:  "limit must be non-negative",
			Code:     "OUT_OF_RANGE",
			Severity: ValidationSeverityError,
		})
	}
	if f.Offset != nil && *f.Offset < 0 {
		errs = append(errs, ValidationError{
			Field:    "offset",
			Message:  "offset must be non-negative",
			Code:     "OUT_OF_RANGE",
			Severity: ValidationSeverityError,
		})
	}
	if f.SortOrder != nil {
		switch strings.ToLower(*f.SortOrder) {
		case "asc", "desc":
			// valid
		default:
			errs = append(errs, ValidationError{
				Field:    "sort_order",
				Message:  fmt.Sprintf("sort_order must be 'asc' or 'desc', got %q", *f.SortOrder),
				Code:     "INVALID_VALUE",
				Severity: ValidationSeverityError,
			})
		}
	}

	return errs
}

// NOTE: Consider adding account category groupings for enhanced reporting
// NOTE: Future enhancement: Add support for account-specific posting rules
