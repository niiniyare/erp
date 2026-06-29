package domain_test

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"awo.so/internal/core/finance/domain"
)

// ============================================================================
// FIN-TYP-001: TransactionStatus.IsEditable — only DRAFT and REJECTED editable
// ============================================================================

func TestTransactionStatusEditability(t *testing.T) {
	tests := []struct {
		status   domain.TransactionStatus
		editable bool
	}{
		{domain.TransactionStatusDraft, true},
		{domain.TransactionStatusRejected, true},
		{domain.TransactionStatusPendingApproval, false},
		{domain.TransactionStatusApproved, false},
		{domain.TransactionStatusPosted, false},
		{domain.TransactionStatusCancelled, false},
		{domain.TransactionStatusReversed, false},
	}

	for _, tc := range tests {
		t.Run(tc.status.String(), func(t *testing.T) {
			assert.Equal(t, tc.editable, tc.status.IsEditable())
		})
	}
}

// ============================================================================
// FIN-TYP-002: PENDING_APPROVAL is never editable (regression guard)
// ============================================================================

func TestPendingApproval_NotEditable(t *testing.T) {
	assert.False(t, domain.TransactionStatusPendingApproval.IsEditable(),
		"PENDING_APPROVAL must never be editable — editing during review bypasses approval controls")
}

// ============================================================================
// FIN-TYP-003: TransactionStatus state machine — valid and illegal transitions
// ============================================================================

func TestTransactionStatusTransitions(t *testing.T) {
	// Valid transitions that MUST exist in the map
	validPairs := [][2]domain.TransactionStatus{
		{domain.TransactionStatusDraft, domain.TransactionStatusPendingApproval},
		{domain.TransactionStatusDraft, domain.TransactionStatusApproved}, // direct approve without approval flow
		{domain.TransactionStatusDraft, domain.TransactionStatusCancelled},
		{domain.TransactionStatusPendingApproval, domain.TransactionStatusApproved},
		{domain.TransactionStatusPendingApproval, domain.TransactionStatusRejected},
		{domain.TransactionStatusPendingApproval, domain.TransactionStatusDraft}, // return to draft for edits
		{domain.TransactionStatusApproved, domain.TransactionStatusPosted},
		{domain.TransactionStatusApproved, domain.TransactionStatusCancelled},
		{domain.TransactionStatusApproved, domain.TransactionStatusDraft},
		{domain.TransactionStatusRejected, domain.TransactionStatusDraft},
		{domain.TransactionStatusRejected, domain.TransactionStatusCancelled},
		{domain.TransactionStatusPosted, domain.TransactionStatusReversed},
	}

	for _, pair := range validPairs {
		from, to := pair[0], pair[1]
		allowed, ok := domain.TransactionStatusTransitions[from]
		if !ok {
			t.Errorf("no transitions defined for status %s", from)
			continue
		}
		found := false
		for _, s := range allowed {
			if s == to {
				found = true
				break
			}
		}
		assert.True(t, found, "expected valid transition %s → %s in transitions table", from, to)
	}

	// Illegal transitions that must NOT exist
	illegalPairs := [][2]domain.TransactionStatus{
		{domain.TransactionStatusPosted, domain.TransactionStatusDraft},
		{domain.TransactionStatusReversed, domain.TransactionStatusPosted},
		{domain.TransactionStatusCancelled, domain.TransactionStatusApproved},
		{domain.TransactionStatusPosted, domain.TransactionStatusPendingApproval},
	}

	for _, pair := range illegalPairs {
		from, to := pair[0], pair[1]
		allowed := domain.TransactionStatusTransitions[from]
		for _, s := range allowed {
			assert.NotEqual(t, to, s,
				"illegal transition %s → %s must not be in the transitions table", from, to)
		}
	}
}

// ============================================================================
// FIN-TYP-004: TransactionType — JOURNAL_ENTRY exists; JOURNAL does not
// ============================================================================

func TestTransactionType_NoDuplicate(t *testing.T) {
	assert.Equal(t, domain.TransactionType("JOURNAL_ENTRY"), domain.TransactionTypeJournalEntry)

	_, err := domain.ParseTransactionType("JOURNAL")
	assert.Error(t, err, "JOURNAL must not be a valid TransactionType")

	parsed, err := domain.ParseTransactionType("JOURNAL_ENTRY")
	assert.NoError(t, err)
	assert.Equal(t, domain.TransactionTypeJournalEntry, parsed)
}

// ============================================================================
// FIN-TYP-010: RejectionReason — only accounting-relevant reasons are valid
// ============================================================================

func TestRejectionReasonValidity(t *testing.T) {
	valid := []domain.RejectionReason{
		domain.RejectionReasonInvalidAccount,
		domain.RejectionReasonPeriodClosed,
		domain.RejectionReasonBudgetExceeded,
		domain.RejectionReasonUnbalancedEntry,
		domain.RejectionReasonMissingDocumentation,
		domain.RejectionReasonDuplicateTransaction,
		domain.RejectionReasonAmountMismatch,
		domain.RejectionReasonUnauthorisedAccount,
		domain.RejectionReasonCurrencyMismatch,
		domain.RejectionReasonPolicyViolation,
		domain.RejectionReasonOther,
	}
	for _, r := range valid {
		assert.True(t, r.IsValid(), "expected rejection reason %q to be valid", r)
	}

	// Payment-gateway codes must be purged — these are NOT accounting reasons
	invalid := []domain.RejectionReason{
		"EXPIRED_CARD",
		"INVALID_MERCHANT",
		"FRAUD_SUSPECTED",
		"DAILY_LIMIT_EXCEEDED",
		"INSUFFICIENT_FUNDS",
		"", // empty string
	}
	for _, r := range invalid {
		assert.False(t, r.IsValid(), "expected rejection reason %q to be invalid", r)
	}
}

// ============================================================================
// FIN-TYP-020: GetNormalBalanceForRootType — correct side per double-entry rules
// ============================================================================

func TestNormalBalance_PerRootType(t *testing.T) {
	tests := []struct {
		rt     domain.RootType
		expect domain.NormalBalance
	}{
		{domain.RootTypeAsset, domain.NormalBalanceDebit},
		{domain.RootTypeExpense, domain.NormalBalanceDebit},
		{domain.RootTypeLiability, domain.NormalBalanceCredit},
		{domain.RootTypeEquity, domain.NormalBalanceCredit},
		{domain.RootTypeRevenue, domain.NormalBalanceCredit},
	}

	for _, tc := range tests {
		t.Run(tc.rt.String(), func(t *testing.T) {
			got := domain.GetNormalBalanceForRootType(tc.rt)
			assert.Equal(t, tc.expect, got,
				"wrong normal balance for root type %s", tc.rt)
		})
	}
}

// Also verify the method-level shortcut matches the function
func TestRootType_NormalBalanceMethod_MatchesFunction(t *testing.T) {
	for _, rt := range domain.ListRootTypes() {
		assert.Equal(t,
			domain.GetNormalBalanceForRootType(rt),
			rt.NormalBalance(),
			"RootType.NormalBalance() must match GetNormalBalanceForRootType() for %s", rt,
		)
	}
}

// ============================================================================
// FIN-TYP-021: RootType.IsValid() rejects unknown values
// ============================================================================

func TestRootType_InvalidValues(t *testing.T) {
	invalid := []domain.RootType{"CONTRA_ASSET", "DEFERRED", ""}
	for _, rt := range invalid {
		assert.False(t, rt.IsValid(), "expected root type %q to be invalid", rt)
	}

	valid := []domain.RootType{
		domain.RootTypeAsset,
		domain.RootTypeLiability,
		domain.RootTypeEquity,
		domain.RootTypeRevenue,
		domain.RootTypeExpense,
	}
	for _, rt := range valid {
		assert.True(t, rt.IsValid(), "expected root type %q to be valid", rt)
	}
}

// ============================================================================
// FIN-TYP-030: AccountStatus.CanTransitionTo — valid transitions from ACTIVE
// ============================================================================

func TestAccountStatus_TransitionsFromActive(t *testing.T) {
	acc := &domain.Accounts{Status: domain.AccountStatusActive}

	// Allowed from ACTIVE
	allowed := []domain.AccountStatus{
		domain.AccountStatusInactive,
		domain.AccountStatusSuspended,
		domain.AccountStatusRestricted,
		domain.AccountStatusFrozen,
		domain.AccountStatusUnderReview,
		domain.AccountStatusYearEndProcessing,
		domain.AccountStatusAuditLock,
		domain.AccountStatusComplianceHold,
		domain.AccountStatusSystemMaintenance,
		domain.AccountStatusDataError,
	}
	for _, s := range allowed {
		assert.True(t, acc.CanTransitionTo(s), "ACTIVE → %s should be allowed", s)
	}

	// NOT allowed from ACTIVE (must go through intermediate states)
	notAllowed := []domain.AccountStatus{
		domain.AccountStatusDraft,
		domain.AccountStatusArchived,
		domain.AccountStatusPendingApproval,
	}
	for _, s := range notAllowed {
		assert.False(t, acc.CanTransitionTo(s), "ACTIVE → %s should not be allowed", s)
	}
}

// ============================================================================
// FIN-TYP-031: AccountStatus re-activation paths
// ============================================================================

func TestAccountStatus_ReactivationPaths(t *testing.T) {
	tests := []struct {
		from        domain.AccountStatus
		canActivate bool
	}{
		{domain.AccountStatusInactive, true},
		{domain.AccountStatusSuspended, true},
		{domain.AccountStatusRestricted, true},
		{domain.AccountStatusFrozen, true},
		{domain.AccountStatusAuditLock, true},
		{domain.AccountStatusComplianceHold, true},
		{domain.AccountStatusClosed, false},   // closed is terminal
		{domain.AccountStatusArchived, false}, // archived is terminal
	}

	for _, tc := range tests {
		acc := &domain.Accounts{Status: tc.from}
		got := acc.CanTransitionTo(domain.AccountStatusActive)
		assert.Equal(t, tc.canActivate, got,
			"from %s → ACTIVE: expected canActivate=%v", tc.from, tc.canActivate)
	}
}

// ============================================================================
// FIN-TYP-032: AccountStatus CLOSED — only ARCHIVED is reachable, nothing else
// ============================================================================

func TestAccountStatus_ClosedIsTerminal(t *testing.T) {
	acc := &domain.Accounts{Status: domain.AccountStatusClosed}

	// Only ARCHIVED is the next valid state
	assert.True(t, acc.CanTransitionTo(domain.AccountStatusArchived),
		"CLOSED → ARCHIVED should be allowed (archival path)")

	// All other statuses are unreachable from CLOSED
	unreachable := []domain.AccountStatus{
		domain.AccountStatusDraft,
		domain.AccountStatusActive,
		domain.AccountStatusInactive,
		domain.AccountStatusSuspended,
		domain.AccountStatusPendingApproval,
		domain.AccountStatusAuditLock,
		domain.AccountStatusComplianceHold,
		domain.AccountStatusFrozen,
		domain.AccountStatusRestricted,
	}
	for _, s := range unreachable {
		assert.False(t, acc.CanTransitionTo(s),
			"CLOSED → %s should not be allowed", s)
	}
}

// ============================================================================
// FIN-TYP-050: BudgetStatus.IsEditable — only DRAFT and REJECTED allow edits
// ============================================================================

func TestBudgetStatus_Editability(t *testing.T) {
	tests := []struct {
		status   domain.BudgetStatus
		editable bool
	}{
		{domain.BudgetStatusDraft, true},
		{domain.BudgetStatusRejected, true}, // can be corrected after rejection
		{domain.BudgetStatusSubmitted, false},
		{domain.BudgetStatusApproved, false},
		{domain.BudgetStatusRevised, false}, // creates a new version
		{domain.BudgetStatusClosed, false},
	}

	for _, tc := range tests {
		t.Run(tc.status.String(), func(t *testing.T) {
			assert.Equal(t, tc.editable, tc.status.IsEditable())
		})
	}
}

// ============================================================================
// FIN-TYP-051: BudgetLineItem.IsOverBudget — threshold is exclusive
// ============================================================================

func TestBudgetLineItem_IsOverBudget(t *testing.T) {
	budget := decimal.NewFromInt(100_000)

	tests := []struct {
		desc   string
		actual decimal.Decimal
		over   bool
	}{
		{"under budget", decimal.NewFromInt(99_999), false},
		{"at exact limit", decimal.NewFromInt(100_000), false}, // at limit, not over
		{"one unit over", decimal.NewFromInt(100_001), true},
		{"zero spend", decimal.Zero, false},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			li := &domain.BudgetLineItem{
				BudgetedAmount: budget,
				ActualAmount:   tc.actual,
			}
			assert.Equal(t, tc.over, li.IsOverBudget(), "actual=%s", tc.actual)
		})
	}
}

// ============================================================================
// FIN-TYP-052: BudgetLineItem.ExceedsVarianceThreshold — percentage-based
// ============================================================================

func TestBudgetLineItem_VarianceThreshold(t *testing.T) {
	budget := decimal.NewFromInt(100_000)
	threshold10 := decimal.NewFromFloat(10.0)

	// 5% variance: (100,000 - 95,000) / 100,000 * 100 = 5%
	li5pct := &domain.BudgetLineItem{
		BudgetedAmount: budget,
		ActualAmount:   decimal.NewFromInt(95_000),
	}
	assert.False(t, li5pct.ExceedsVarianceThreshold(threshold10),
		"5%% variance should be within 10%% threshold")

	// 12% variance: (100,000 - 88,000) / 100,000 * 100 = 12%
	li12pct := &domain.BudgetLineItem{
		BudgetedAmount: budget,
		ActualAmount:   decimal.NewFromInt(88_000),
	}
	assert.True(t, li12pct.ExceedsVarianceThreshold(threshold10),
		"12%% variance should exceed 10%% threshold")

	// Over-spend: negative variance is treated as absolute value
	liOverSpend := &domain.BudgetLineItem{
		BudgetedAmount: budget,
		ActualAmount:   decimal.NewFromInt(115_000), // 15% over
	}
	assert.True(t, liOverSpend.ExceedsVarianceThreshold(threshold10),
		"15%% overspend should exceed 10%% threshold (absolute)")

	// Zero budget: no division by zero; returns false
	liZeroBudget := &domain.BudgetLineItem{
		BudgetedAmount: decimal.Zero,
		ActualAmount:   decimal.NewFromInt(1_000),
	}
	assert.False(t, liZeroBudget.ExceedsVarianceThreshold(threshold10),
		"zero budget should not panic and should return false")
}
