// Package domain — Financial System Consistency Layer types.
//
// This file defines the system-wide invariant model that binds COA, transactions,
// reporting, and audit into ONE CONSISTENT TRUTH MODEL.
//
// # Core invariants (enforced by service layer, verified by scanners)
//
//  1. Every posted transaction entry MUST resolve to a valid COA node at posting time.
//  2. Every financial report MUST be derived from a time-bound COA snapshot + ledger.
//  3. No report may depend on runtime-only derived state.
//  4. Audit chain MUST be sufficient to reconstruct any report.
//  5. COA mutations MUST be time-stamped and leave a snapshot trail.
//  6. Balance computed from ledger replay MUST equal stored balance (drift = violation).
//
// # Snapshot model
//
// COASnapshot captures the full COA structure at a point in time.
// Reports bind to a snapshot, not the live COA, so historical reports are
// immune to future structural mutations.
//
// # Deterministic reporting
//
// FinancialReportHash captures the inputs and output of a report as a SHA-256
// checksum. Recomputing the report from the same inputs MUST produce the same
// hash — any divergence signals corruption or non-determinism.
package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// =============================================================================
// 1. COA Snapshot — point-in-time immutable COA structure
// =============================================================================

// COASnapshot is an immutable capture of the full Chart of Accounts structure
// at a specific point in time. Reports MUST bind to a snapshot; never to the
// live COA. This guarantees historical report reproducibility even after COA
// mutations.
type COASnapshot struct {
	ID         uuid.UUID `json:"id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	SnapshotAt time.Time `json:"snapshot_at"`

	// PeriodID links this snapshot to an accounting period (optional but recommended).
	// When set, the snapshot represents the COA structure at period open.
	PeriodID *uuid.UUID `json:"period_id,omitempty"`

	// Accounts is the complete, ordered set of COA nodes at SnapshotAt.
	// Immutable after creation; never updated.
	Accounts []COASnapshotEntry `json:"accounts"`

	// Hash is a deterministic SHA-256 of all account entries in canonical order.
	// Used to detect snapshot tampering.
	Hash string `json:"hash"`

	// EntryCount is len(Accounts), denormalised for fast validation.
	EntryCount int `json:"entry_count"`

	CreatedBy uuid.UUID `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

// COASnapshotEntry is a single account node frozen into a snapshot.
// It contains exactly the fields required for report generation and
// transaction-to-COA binding — no operational metadata.
type COASnapshotEntry struct {
	AccountID     uuid.UUID    `json:"account_id"`
	AccountCode   string       `json:"account_code"`
	AccountName   string       `json:"account_name"`
	RootType      RootType     `json:"root_type"`
	NormalBalance NormalBalance `json:"normal_balance"`
	Status        AccountStatus `json:"status"`
	ParentID      *uuid.UUID   `json:"parent_id,omitempty"`
	AccountLevel  int32        `json:"account_level"`
	CurrencyCode  *string      `json:"currency_code,omitempty"`
	IsActive      bool         `json:"is_active"`
}

// IsPostingEligible returns true when this snapshot entry was in a state that
// allowed journal postings at snapshot time. Used during transaction-to-COA
// binding validation.
func (e COASnapshotEntry) IsPostingEligible() bool {
	return e.Status == AccountStatusActive && e.IsActive
}

// COASnapshotRepository persists and retrieves COA snapshots.
type COASnapshotRepository interface {
	// Create persists a new snapshot. Snapshots are immutable after creation.
	Create(ctx context.Context, snap *COASnapshot) error

	// GetByID retrieves a snapshot by its ID.
	GetByID(ctx context.Context, id uuid.UUID) (*COASnapshot, error)

	// GetForPeriod retrieves the snapshot bound to a fiscal period.
	// Returns nil when no snapshot exists for the period.
	GetForPeriod(ctx context.Context, tenantID uuid.UUID, periodID uuid.UUID) (*COASnapshot, error)

	// GetAtTime retrieves the most recent snapshot taken on or before asOfTime.
	// Returns nil when no snapshot exists before asOfTime.
	GetAtTime(ctx context.Context, tenantID uuid.UUID, asOfTime time.Time) (*COASnapshot, error)

	// List returns all snapshots for a tenant in reverse chronological order.
	List(ctx context.Context, tenantID uuid.UUID) ([]*COASnapshot, error)
}

// =============================================================================
// 2. Financial Report Hash — deterministic report fingerprint
// =============================================================================

// FinancialReportKind identifies the type of financial report.
type FinancialReportKind string

const (
	ReportKindTrialBalance  FinancialReportKind = "TRIAL_BALANCE"
	ReportKindBalanceSheet  FinancialReportKind = "BALANCE_SHEET"
	ReportKindProfitAndLoss FinancialReportKind = "PROFIT_AND_LOSS"
	ReportKindCashFlow      FinancialReportKind = "CASH_FLOW"
)

// FinancialReportHash is a deterministic fingerprint of a financial report.
// Given the same COA snapshot, the same ledger transactions, and the same
// as-of date, the hash MUST be identical on every recomputation.
//
// A mismatch between a stored hash and a recomputed hash signals one of:
//   - Ledger mutation (transactions added, modified, or deleted)
//   - COA structural drift (snapshot tampered)
//   - Non-determinism in the reporting engine (bug)
type FinancialReportHash struct {
	ID         uuid.UUID           `json:"id"`
	TenantID   uuid.UUID           `json:"tenant_id"`
	ReportKind FinancialReportKind `json:"report_kind"`
	AsOfDate   time.Time           `json:"as_of_date"`

	// SnapshotID is the COA snapshot used as input.
	SnapshotID uuid.UUID `json:"snapshot_id"`

	// LedgerHash is a deterministic hash of all transaction entries
	// (in canonical order) that contributed to this report.
	LedgerHash string `json:"ledger_hash"`

	// ReportHash is the SHA-256 of the canonical report output (rows, amounts).
	// This is the primary reproducibility signal.
	ReportHash string `json:"report_hash"`

	// EntryCount is the number of transaction entries included.
	EntryCount int `json:"entry_count"`

	GeneratedAt time.Time `json:"generated_at"`
	GeneratedBy uuid.UUID `json:"generated_by"`
}

// =============================================================================
// 3. Financial Consistency Violation — cross-system invariant failure
// =============================================================================

// ConsistencyViolationKind identifies the class of cross-system invariant broken.
type ConsistencyViolationKind string

const (
	// ConsistencyOrphanPosting — a transaction entry references an account that
	// does not exist in the COA at all.
	ConsistencyOrphanPosting ConsistencyViolationKind = "ORPHAN_POSTING"

	// ConsistencyAccountIneligibleAtPosting — the account referenced by a
	// transaction entry was not in a posting-eligible state at posting time.
	ConsistencyAccountIneligibleAtPosting ConsistencyViolationKind = "ACCOUNT_INELIGIBLE_AT_POSTING"

	// ConsistencyMissingAuditEntry — a posted transaction has no corresponding
	// audit chain entry.
	ConsistencyMissingAuditEntry ConsistencyViolationKind = "MISSING_AUDIT_ENTRY"

	// ConsistencyBalanceDrift — the stored account balance diverges from the
	// balance computed by replaying posted transactions.
	ConsistencyBalanceDrift ConsistencyViolationKind = "BALANCE_DRIFT"

	// ConsistencySnapshotMismatch — a report was generated from a different COA
	// state than the stored snapshot hash indicates.
	ConsistencySnapshotMismatch ConsistencyViolationKind = "SNAPSHOT_MISMATCH"

	// ConsistencyReportHashMismatch — recomputing a stored report produces a
	// different hash; the report has drifted from the ledger.
	ConsistencyReportHashMismatch ConsistencyViolationKind = "REPORT_HASH_MISMATCH"

	// ConsistencyCOAMutationWithoutSnapshot — a COA structural change occurred
	// without a corresponding snapshot, breaking historical report reproducibility.
	ConsistencyCOAMutationWithoutSnapshot ConsistencyViolationKind = "COA_MUTATION_WITHOUT_SNAPSHOT"

	// ConsistencyUnreplayableLedger — replaying the transaction ledger does not
	// produce the stored account balances.
	ConsistencyUnreplayableLedger ConsistencyViolationKind = "UNREPLAYABLE_LEDGER"
)

// ConsistencyViolation is a single cross-system invariant failure.
type ConsistencyViolation struct {
	Kind      ConsistencyViolationKind `json:"kind"`
	Severity  COAViolationSeverity     `json:"severity"`
	AccountID *uuid.UUID               `json:"account_id,omitempty"`
	TxnID     *uuid.UUID               `json:"txn_id,omitempty"`
	Detail    string                   `json:"detail"`
}

// SystemConsistencyReport is the output of a full cross-system consistency scan.
// It covers COA ↔ transactions ↔ audit ↔ reports as a unified truth model.
type SystemConsistencyReport struct {
	TenantID   uuid.UUID `json:"tenant_id"`
	AsOfDate   time.Time `json:"as_of_date"`
	GeneratedAt time.Time `json:"generated_at"`

	// Violations contains all detected cross-system invariant failures.
	Violations []ConsistencyViolation `json:"violations,omitempty"`

	// Healthy is false when any CRITICAL or HIGH violations exist.
	Healthy bool `json:"healthy"`

	// Metrics — scan scope summary.
	LedgerTxnCount  int `json:"ledger_txn_count"`
	LedgerEntryCount int `json:"ledger_entry_count"`
	COAAccountCount int `json:"coa_account_count"`
}

// HasCritical returns true when any CRITICAL violation exists.
func (r *SystemConsistencyReport) HasCritical() bool {
	for _, v := range r.Violations {
		if v.Severity == COASeverityCritical {
			return true
		}
	}
	return false
}

// =============================================================================
// 4. Replay-based reconstruction types
// =============================================================================

// ReplayedBalance is a single account balance derived by replaying
// posted transaction entries. Used to validate stored balances.
type ReplayedBalance struct {
	AccountID      uuid.UUID       `json:"account_id"`
	AccountCode    string          `json:"account_code"`
	ReplayedDebit  decimal.Decimal `json:"replayed_debit"`
	ReplayedCredit decimal.Decimal `json:"replayed_credit"`
	// NetBalance = ReplayedDebit - ReplayedCredit (raw, not sign-adjusted)
	NetBalance decimal.Decimal `json:"net_balance"`
	TxnCount   int             `json:"txn_count"`
}

// ReplayReport is the output of a full ledger replay.
type ReplayReport struct {
	TenantID    uuid.UUID         `json:"tenant_id"`
	AsOfDate    time.Time         `json:"as_of_date"`
	GeneratedAt time.Time         `json:"generated_at"`
	Balances    []ReplayedBalance `json:"balances"`
	TxnsReplayed int              `json:"txns_replayed"`
	EntriesReplayed int           `json:"entries_replayed"`
	// DriftAccounts contains IDs of accounts where replayed balance ≠ stored balance.
	DriftAccounts []uuid.UUID `json:"drift_accounts,omitempty"`
	ReplayClean   bool        `json:"replay_clean"` // true = no drift
}

// =============================================================================
// 5. Deterministic report output types
// =============================================================================

// TrialBalanceReport is the canonical deterministic trial balance output.
// Binding: must specify AsOfDate and SnapshotID for reproducibility.
type TrialBalanceReport struct {
	TenantID   uuid.UUID           `json:"tenant_id"`
	AsOfDate   time.Time           `json:"as_of_date"`
	SnapshotID uuid.UUID           `json:"snapshot_id"`
	GeneratedAt time.Time          `json:"generated_at"`
	Lines      []TrialBalanceLine  `json:"lines"`
	TotalDebits  decimal.Decimal   `json:"total_debits"`
	TotalCredits decimal.Decimal   `json:"total_credits"`
	// Balanced is true when TotalDebits == TotalCredits.
	Balanced bool   `json:"balanced"`
	Hash     string `json:"hash"` // deterministic fingerprint
}

// TrialBalanceLine is a single row in the deterministic trial balance.
type TrialBalanceLine struct {
	AccountID    uuid.UUID       `json:"account_id"`
	AccountCode  string          `json:"account_code"`
	AccountName  string          `json:"account_name"`
	RootType     RootType        `json:"root_type"`
	NormalBalance NormalBalance  `json:"normal_balance"`
	Debit        decimal.Decimal `json:"debit"`
	Credit       decimal.Decimal `json:"credit"`
}
