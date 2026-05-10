// Package domain — COA governance types.
//
// This file extends the Accounts domain with the governance primitives needed
// to treat chart-of-accounts mutations as first-class financial events:
//
//   - AccountLifecycleEvent  records every state transition with actor + reason
//   - OpeningBalance          governed opening-balance request (requires approval)
//   - AccountMutationRisk     classifies how dangerous a mutation is
//   - HierarchyViolation      structural invariant failures (cycle, depth, type)
//   - COAIntegrityViolation   severity-tagged integrity scan result
//
// These types are the contracts between the service layer and the governance
// infrastructure (lifecycle service, opening balance engine, integrity scanner).
package domain

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// =============================================================================
// 1. AccountLifecycleEvent — immutable record of every status transition
// =============================================================================

// AccountLifecycleEventKind classifies a lifecycle event.
type AccountLifecycleEventKind string

const (
	// EventKindSubmitted — account submitted for approval (DRAFT → PENDING_APPROVAL).
	EventKindSubmitted AccountLifecycleEventKind = "SUBMITTED"

	// EventKindApproved — account approved and activated (PENDING_APPROVAL → ACTIVE).
	EventKindApproved AccountLifecycleEventKind = "APPROVED"

	// EventKindRejected — account approval rejected, returned to draft.
	EventKindRejected AccountLifecycleEventKind = "REJECTED"

	// EventKindFrozen — account frozen by operator.
	EventKindFrozen AccountLifecycleEventKind = "FROZEN"

	// EventKindUnfrozen — freeze lifted, account returned to ACTIVE.
	EventKindUnfrozen AccountLifecycleEventKind = "UNFROZEN"

	// EventKindSuspended — administrative hold placed.
	EventKindSuspended AccountLifecycleEventKind = "SUSPENDED"

	// EventKindRestricted — account restricted to approval-only entries.
	EventKindRestricted AccountLifecycleEventKind = "RESTRICTED"

	// EventKindDeactivated — account moved to INACTIVE.
	EventKindDeactivated AccountLifecycleEventKind = "DEACTIVATED"

	// EventKindClosed — account permanently closed.
	EventKindClosed AccountLifecycleEventKind = "CLOSED"

	// EventKindArchived — account moved to long-term storage.
	EventKindArchived AccountLifecycleEventKind = "ARCHIVED"

	// EventKindComplianceHold — regulatory hold applied.
	EventKindComplianceHold AccountLifecycleEventKind = "COMPLIANCE_HOLD"

	// EventKindAuditLock — external audit lock applied.
	EventKindAuditLock AccountLifecycleEventKind = "AUDIT_LOCK"

	// EventKindHierarchyReparent — account moved to a different parent.
	EventKindHierarchyReparent AccountLifecycleEventKind = "HIERARCHY_REPARENT"

	// EventKindCurrencyChange — account currency changed (controlled workflow).
	EventKindCurrencyChange AccountLifecycleEventKind = "CURRENCY_CHANGE"

	// EventKindMerged — account merged into another account.
	EventKindMerged AccountLifecycleEventKind = "MERGED"

	// EventKindSuperseded — account superseded by a new account.
	EventKindSuperseded AccountLifecycleEventKind = "SUPERSEDED"
)

// AccountLifecycleEvent is an immutable record of a state transition or
// high-risk mutation on an account. Written by the lifecycle service and
// the mutation governor; never updated after creation.
//
// This is NOT a generic audit log. It is the FORENSIC ACCOUNT STRUCTURE HISTORY
// required to reconstruct the COA at any point in time.
type AccountLifecycleEvent struct {
	ID        uuid.UUID                 `json:"id"`
	TenantID  uuid.UUID                 `json:"tenant_id"`
	AccountID uuid.UUID                 `json:"account_id"`
	Kind      AccountLifecycleEventKind `json:"kind"`

	// FromStatus / ToStatus capture the state machine transition.
	// Empty for non-status-changing events (e.g. currency change).
	FromStatus AccountStatus `json:"from_status,omitempty"`
	ToStatus   AccountStatus `json:"to_status,omitempty"`

	// ActorID is the user who triggered this transition.
	ActorID uuid.UUID `json:"actor_id"`

	// Reason is the mandatory justification for high-risk transitions.
	Reason string `json:"reason,omitempty"`

	// Metadata holds event-specific additional data (e.g. target account ID for
	// MERGED events, old/new currency codes for CURRENCY_CHANGE).
	Metadata map[string]any `json:"metadata,omitempty"`

	OccurredAt time.Time `json:"occurred_at"`
}

// AccountLifecycleEventRepository persists lifecycle events.
// Implementations must be append-only; updates and deletes are not permitted.
type AccountLifecycleEventRepository interface {
	// Append writes a single event. Must be idempotent on ID (duplicate ID = no-op).
	Append(ctx context.Context, event *AccountLifecycleEvent) error

	// ListForAccount returns events for a single account in chronological order.
	ListForAccount(ctx context.Context, accountID uuid.UUID) ([]*AccountLifecycleEvent, error)

	// ListForTenant returns all lifecycle events for a tenant in the given window.
	ListForTenant(ctx context.Context, tenantID uuid.UUID, from, to time.Time) ([]*AccountLifecycleEvent, error)
}

// =============================================================================
// 2. OpeningBalance — governed opening-balance request
// =============================================================================

// OpeningBalanceStatus tracks the approval lifecycle of an opening balance.
type OpeningBalanceStatus string

const (
	// OpeningBalanceStatusPending — submitted, awaiting approval.
	OpeningBalanceStatusPending OpeningBalanceStatus = "PENDING"

	// OpeningBalanceStatusApproved — approved but not yet posted.
	OpeningBalanceStatusApproved OpeningBalanceStatus = "APPROVED"

	// OpeningBalanceStatusPosted — posted as a journal entry; immutable.
	OpeningBalanceStatusPosted OpeningBalanceStatus = "POSTED"

	// OpeningBalanceStatusRejected — rejected; a new submission is required.
	OpeningBalanceStatusRejected OpeningBalanceStatus = "REJECTED"

	// OpeningBalanceStatusVoided — voided before posting (e.g. period reopened).
	OpeningBalanceStatusVoided OpeningBalanceStatus = "VOIDED"
)

// OpeningBalance is a governed request to set the opening balance of an
// account for a given fiscal period. It is NOT a simple field update.
//
// The workflow:
//  1. Submitter calls SetOpeningBalance → status = PENDING
//  2. Approver (different person) calls ApproveOpeningBalance → status = APPROVED
//  3. Accountant calls PostOpeningBalance → status = POSTED (creates journal entry)
//
// Once POSTED, the balance is immutable. Corrections require a new journal entry
// (not a new opening balance submission).
type OpeningBalance struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	AccountID uuid.UUID `json:"account_id"`

	// PeriodID links to the accounting period this opening balance applies to.
	// The engine validates that the period exists and is not closed.
	PeriodID uuid.UUID `json:"period_id"`

	// EffectiveDate is the date from which this opening balance applies.
	// Must fall within the referenced period.
	EffectiveDate time.Time `json:"effective_date"`

	// DebitAmount / CreditAmount encode the opening position.
	// Exactly one must be non-zero; which one must match the account's NormalBalance.
	DebitAmount  decimal.Decimal `json:"debit_amount"`
	CreditAmount decimal.Decimal `json:"credit_amount"`

	// CurrencyCode must match the account's CurrencyCode (or tenant default).
	CurrencyCode string `json:"currency_code"`

	// Status tracks the approval lifecycle.
	Status OpeningBalanceStatus `json:"status"`

	// SubmittedBy / ApprovedBy / PostedBy capture the separation-of-duties chain.
	// ApproveOpeningBalance enforces SubmittedBy != ApprovedBy.
	SubmittedBy uuid.UUID  `json:"submitted_by"`
	ApprovedBy  *uuid.UUID `json:"approved_by,omitempty"`
	PostedBy    *uuid.UUID `json:"posted_by,omitempty"`

	// TransactionID is populated after PostOpeningBalance creates the journal entry.
	// Nil while status is PENDING or APPROVED.
	TransactionID *uuid.UUID `json:"transaction_id,omitempty"`

	// Reason is the mandatory justification provided by the submitter.
	Reason string `json:"reason"`

	// RejectionReason is populated when the approver rejects the request.
	RejectionReason *string `json:"rejection_reason,omitempty"`

	SubmittedAt time.Time  `json:"submitted_at"`
	ApprovedAt  *time.Time `json:"approved_at,omitempty"`
	PostedAt    *time.Time `json:"posted_at,omitempty"`
}

// IsPosted returns true when the opening balance has been committed to the ledger.
func (ob *OpeningBalance) IsPosted() bool {
	return ob.Status == OpeningBalanceStatusPosted
}

// IsImmutable returns true when no further changes to the balance are permitted.
func (ob *OpeningBalance) IsImmutable() bool {
	return ob.Status == OpeningBalanceStatusPosted || ob.Status == OpeningBalanceStatusVoided
}

// Amount returns the signed opening balance amount (positive = debit, negative = credit).
// Used for internal calculations; use DebitAmount/CreditAmount for display.
func (ob *OpeningBalance) Amount() decimal.Decimal {
	if !ob.DebitAmount.IsZero() {
		return ob.DebitAmount
	}
	return ob.CreditAmount.Neg()
}

// OpeningBalanceRequest is the input DTO for SetOpeningBalance.
type OpeningBalanceRequest struct {
	AccountID     uuid.UUID       `json:"account_id"`
	PeriodID      uuid.UUID       `json:"period_id"`
	EffectiveDate time.Time       `json:"effective_date"`
	DebitAmount   decimal.Decimal `json:"debit_amount"`
	CreditAmount  decimal.Decimal `json:"credit_amount"`
	CurrencyCode  string          `json:"currency_code"`
	Reason        string          `json:"reason"`
	SubmittedBy   uuid.UUID       `json:"submitted_by"`
}

// Validate returns an error describing all validation failures in the request.
func (r *OpeningBalanceRequest) Validate() error {
	var errs []string

	if r.AccountID == uuid.Nil {
		errs = append(errs, "account_id is required")
	}
	if r.PeriodID == uuid.Nil {
		errs = append(errs, "period_id is required")
	}
	if r.EffectiveDate.IsZero() {
		errs = append(errs, "effective_date is required")
	}
	if r.SubmittedBy == uuid.Nil {
		errs = append(errs, "submitted_by is required")
	}
	if r.Reason == "" {
		errs = append(errs, "reason is required for opening balance submissions")
	}

	bothZero := r.DebitAmount.IsZero() && r.CreditAmount.IsZero()
	bothNonZero := !r.DebitAmount.IsZero() && !r.CreditAmount.IsZero()
	if bothZero {
		errs = append(errs, "either debit_amount or credit_amount must be non-zero")
	}
	if bothNonZero {
		errs = append(errs, "only one of debit_amount or credit_amount may be non-zero")
	}
	if !r.DebitAmount.IsZero() && r.DebitAmount.IsNegative() {
		errs = append(errs, "debit_amount must be positive")
	}
	if !r.CreditAmount.IsZero() && r.CreditAmount.IsNegative() {
		errs = append(errs, "credit_amount must be positive")
	}
	if r.CurrencyCode == "" {
		errs = append(errs, "currency_code is required")
	}

	if len(errs) > 0 {
		return fmt.Errorf("opening balance validation: %v", errs)
	}
	return nil
}

// OpeningBalanceRepository persists opening balance requests.
type OpeningBalanceRepository interface {
	Create(ctx context.Context, ob *OpeningBalance) error
	GetByID(ctx context.Context, id uuid.UUID) (*OpeningBalance, error)
	GetByAccount(ctx context.Context, accountID uuid.UUID) ([]*OpeningBalance, error)
	Update(ctx context.Context, ob *OpeningBalance) error
}

// =============================================================================
// 3. AccountMutationRisk — risk classification for COA mutations
// =============================================================================

// AccountMutationRisk classifies how dangerous a proposed account mutation is.
// Higher risk mutations require approval, audit trail, and impact analysis.
type AccountMutationRisk string

const (
	// RiskLow — routine mutation (name change, description update).
	// No approval required; audited but not blocked.
	RiskLow AccountMutationRisk = "LOW"

	// RiskMedium — operational change (activation, deactivation, freeze).
	// Requires authorization but no dual-approval.
	RiskMedium AccountMutationRisk = "MEDIUM"

	// RiskHigh — structural change (reparenting, currency change).
	// Requires approval and impact analysis; blocked after postings in some cases.
	RiskHigh AccountMutationRisk = "HIGH"

	// RiskCritical — irreversible structural mutation (root type change, merge, close).
	// Always requires dual-approval and audit chain entry. Blocked after postings.
	RiskCritical AccountMutationRisk = "CRITICAL"
)

// MutationRiskOf returns the risk level for a given type of account mutation.
// This is the single source of truth — service layer must not hard-code risk
// levels in individual methods.
func MutationRiskOf(mutationType string) AccountMutationRisk {
	switch mutationType {
	// LOW — safe to make without approval
	case "name_change", "description_change", "attributes_update",
		"report_order_change", "require_reference_toggle":
		return RiskLow

	// MEDIUM — operational changes that require authorization
	case "activate", "deactivate", "suspend", "restrict", "unfreeze",
		"allow_manual_entries_toggle", "budgetable_toggle":
		return RiskMedium

	// HIGH — structural changes that risk data integrity
	case "freeze", "compliance_hold", "audit_lock",
		"currency_change", "hierarchy_reparent", "control_account_change":
		return RiskHigh

	// CRITICAL — irreversible mutations that corrupt history if wrong
	case "root_type_change", "normal_balance_change", "close", "archive",
		"merge", "supersede", "delete":
		return RiskCritical

	default:
		return RiskHigh // unknown mutation types default to HIGH
	}
}

// =============================================================================
// 4. HierarchyViolation — structural invariant failures
// =============================================================================

// HierarchyViolationKind identifies the type of hierarchy constraint violated.
type HierarchyViolationKind string

const (
	// HierarchyViolationCycle — proposed reparenting would create a cycle.
	HierarchyViolationCycle HierarchyViolationKind = "CYCLE"

	// HierarchyViolationRootTypeMismatch — parent and child have incompatible root types.
	HierarchyViolationRootTypeMismatch HierarchyViolationKind = "ROOT_TYPE_MISMATCH"

	// HierarchyViolationDepthExceeded — proposed level exceeds MaxAccountHierarchyDepth.
	HierarchyViolationDepthExceeded HierarchyViolationKind = "DEPTH_EXCEEDED"

	// HierarchyViolationSelfParent — account references itself as its own parent.
	HierarchyViolationSelfParent HierarchyViolationKind = "SELF_PARENT"

	// HierarchyViolationOrphan — parent account does not exist or is deleted.
	HierarchyViolationOrphan HierarchyViolationKind = "ORPHAN"

	// HierarchyViolationTenantMismatch — parent and child belong to different tenants.
	HierarchyViolationTenantMismatch HierarchyViolationKind = "TENANT_MISMATCH"

	// HierarchyViolationEntityMismatch — parent and child belong to different legal entities.
	HierarchyViolationEntityMismatch HierarchyViolationKind = "ENTITY_MISMATCH"
)

// HierarchyViolation describes a single structural invariant failure.
type HierarchyViolation struct {
	Kind      HierarchyViolationKind `json:"kind"`
	AccountID uuid.UUID              `json:"account_id"`
	Detail    string                 `json:"detail"`
}

// Error implements the error interface so violations can be returned directly.
func (v HierarchyViolation) Error() string {
	return fmt.Sprintf("hierarchy violation [%s] on account %s: %s", v.Kind, v.AccountID, v.Detail)
}

// =============================================================================
// 5. COAIntegrityViolation — severity-tagged COA integrity scan result
// =============================================================================

// COAViolationSeverity mirrors the transaction IntegrityService severity model
// so governance dashboards can use the same severity rendering.
type COAViolationSeverity string

const (
	COASeverityCritical COAViolationSeverity = "CRITICAL"
	COASeverityHigh     COAViolationSeverity = "HIGH"
	COASeverityMedium   COAViolationSeverity = "MEDIUM"
	COASeverityLow      COAViolationSeverity = "LOW"
)

// COAViolationKind identifies the type of COA integrity issue.
type COAViolationKind string

const (
	// COAViolationOrphanAccount — account whose parent does not exist.
	COAViolationOrphanAccount COAViolationKind = "ORPHAN_ACCOUNT"

	// COAViolationCyclicHierarchy — cycle detected in the account tree.
	COAViolationCyclicHierarchy COAViolationKind = "CYCLIC_HIERARCHY"

	// COAViolationRootTypeMismatch — parent/child root types are incompatible.
	COAViolationRootTypeMismatch COAViolationKind = "ROOT_TYPE_MISMATCH"

	// COAViolationInvalidPath — materialised path is inconsistent with parent chain.
	COAViolationInvalidPath COAViolationKind = "INVALID_PATH"

	// COAViolationDepthExceeded — account depth exceeds MaxAccountHierarchyDepth.
	COAViolationDepthExceeded COAViolationKind = "DEPTH_EXCEEDED"

	// COAViolationDeadAccount — ACTIVE account with no transactions in long time.
	COAViolationDeadAccount COAViolationKind = "DEAD_ACCOUNT"

	// COAViolationBalanceInClosedAccount — CLOSED/ARCHIVED account with non-zero balance.
	COAViolationBalanceInClosedAccount COAViolationKind = "BALANCE_IN_CLOSED_ACCOUNT"

	// COAViolationNormalBalanceMismatch — normal balance does not match root type.
	COAViolationNormalBalanceMismatch COAViolationKind = "NORMAL_BALANCE_MISMATCH"

	// COAViolationDuplicateCode — two accounts share the same account code.
	COAViolationDuplicateCode COAViolationKind = "DUPLICATE_CODE"

	// COAViolationInvalidStatusForBalance — non-zero balance on DRAFT/PENDING account.
	COAViolationInvalidStatusForBalance COAViolationKind = "INVALID_STATUS_FOR_BALANCE"
)

// COAIntegrityViolation is a single finding from a COA integrity scan.
type COAIntegrityViolation struct {
	Kind      COAViolationKind     `json:"kind"`
	Severity  COAViolationSeverity `json:"severity"`
	AccountID uuid.UUID            `json:"account_id"`
	Detail    string               `json:"detail"`
}

// COAIntegrityReport is the output of a full COA integrity scan.
// It mirrors IntegrityReport from the transaction subsystem.
type COAIntegrityReport struct {
	TenantID       uuid.UUID              `json:"tenant_id"`
	GeneratedAt    time.Time              `json:"generated_at"`
	AccountsScanned int                   `json:"accounts_scanned"`
	Violations     []COAIntegrityViolation `json:"violations,omitempty"`
	Healthy        bool                   `json:"healthy"` // true when no CRITICAL or HIGH violations
}

// HasCritical returns true when the report contains any CRITICAL violations.
func (r *COAIntegrityReport) HasCritical() bool {
	for _, v := range r.Violations {
		if v.Severity == COASeverityCritical {
			return true
		}
	}
	return false
}

// HasHigh returns true when the report contains any HIGH or CRITICAL violations.
func (r *COAIntegrityReport) HasHigh() bool {
	for _, v := range r.Violations {
		if v.Severity == COASeverityCritical || v.Severity == COASeverityHigh {
			return true
		}
	}
	return false
}

// =============================================================================
// 6. New domain errors for governance enforcement
// =============================================================================

var (
	// Lifecycle errors
	ErrAccountNotInDraft          = errors.New("account is not in DRAFT status")
	ErrAccountNotPendingApproval  = errors.New("account is not in PENDING_APPROVAL status")
	ErrSelfApprovalNotAllowed     = errors.New("the approver cannot be the same person who submitted the account (SOD)")
	ErrAccountStatusTransition    = errors.New("account status transition not permitted")
	ErrAccountHasNonZeroBalance   = errors.New("account has a non-zero balance; cannot proceed")
	ErrAccountFrozen              = errors.New("account is frozen; no mutations permitted")

	// Mutation governance errors
	ErrMutationBlockedAfterPosting = errors.New("this mutation is not permitted after transactions have been posted")
	ErrRootTypeChangeForbidden     = errors.New("root type cannot be changed after transactions have been posted")
	ErrCurrencyChangeForbidden     = errors.New("currency cannot be changed after transactions have been posted")
	ErrHighRiskMutationRequiresApproval = errors.New("high-risk account mutation requires explicit approval")

	// Hierarchy errors
	ErrHierarchyCycleDetected     = errors.New("proposed hierarchy change creates a circular reference")
	ErrHierarchyDepthExceeded     = errors.New("proposed hierarchy depth exceeds maximum allowed level")
	ErrHierarchyRootTypeMismatch  = errors.New("parent and child accounts have incompatible root types")
	ErrHierarchySelfParent        = errors.New("account cannot be its own parent")
	ErrHierarchyTenantMismatch    = errors.New("parent and child accounts belong to different tenants")

	// Opening balance errors
	ErrOpeningBalanceAlreadyPosted  = errors.New("opening balance has already been posted; use a journal entry for corrections")
	ErrOpeningBalanceSelfApproval   = errors.New("opening balance approver cannot be the same as the submitter (SOD)")
	ErrOpeningBalanceCurrencyMismatch = errors.New("opening balance currency does not match account currency")
	ErrOpeningBalanceDirectionMismatch = errors.New("opening balance debit/credit direction does not match account normal balance")
	ErrOpeningBalanceInClosedPeriod = errors.New("opening balance cannot be set in a closed accounting period")
	ErrOpeningBalanceNotApproved    = errors.New("opening balance must be approved before posting")
)
