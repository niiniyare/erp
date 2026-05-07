// Package service_test — property-based financial invariant tests.
//
// Uses testing/quick to verify invariants hold across randomized inputs.
// Each test encodes a mathematical or state-machine property that must hold
// for ALL valid inputs, not just the cherry-picked cases in unit tests.
package service_test

import (
	"context"
	"testing"
	"testing/quick"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"awo.so/internal/core/finance/domain"
	"awo.so/internal/core/finance/service"
	"awo.so/internal/shared"
	"awo.so/internal/shared/metrics"
)

// ============================================================================
// FIN-PROP-001: Balanced entry construction NEVER produces UNBALANCED violation
// ============================================================================
// Invariant: any set of entries where ∑debit == ∑credit passes the integrity check.
// This proves the check is sound (no false positives on valid data).

func TestProperty_LedgerBalance_BalancedEntries_NeverViolated(t *testing.T) {
	prop := func(amounts []uint16) bool {
		if len(amounts) == 0 {
			return true // skip degenerate input
		}
		var entries []domain.TransactionEntry
		for _, a := range amounts {
			if a == 0 {
				a = 1
			}
			amt := decimal.NewFromInt(int64(a))
			entries = append(entries,
				domain.TransactionEntry{DebitAmount: amt, CreditAmount: decimal.Zero},
				domain.TransactionEntry{DebitAmount: decimal.Zero, CreditAmount: amt},
			)
		}
		txn := &domain.Transaction{
			ID:                uuid.New(),
			TenantID:          uuid.New(),
			TransactionStatus: domain.TransactionStatusPosted,
		}
		pageCount := 0
		repo := &stubTxnRepo{
			fnList: func(_ context.Context, _ *domain.TransactionFilter) ([]*domain.Transaction, error) {
				pageCount++
				if pageCount == 1 {
					return []*domain.Transaction{txn}, nil
				}
				return nil, nil
			},
			fnGetEntriesByTxn: func(_ context.Context, _ uuid.UUID) ([]domain.TransactionEntry, error) {
				return entries, nil
			},
		}
		intSvc := service.NewIntegrityService(repo, nil, metrics.NewNoOpMetricsProvider())
		report, err := intSvc.ScanPostedTransactions(context.Background(), nil)
		if err != nil {
			return false
		}
		for _, v := range report.Violations {
			if v.Kind == "UNBALANCED_TRANSACTION" {
				return false // false positive — invariant violated
			}
		}
		return true
	}
	if err := quick.Check(prop, &quick.Config{MaxCount: 300}); err != nil {
		t.Errorf("balance invariant false positive: %v", err)
	}
}

// ============================================================================
// FIN-PROP-002: Unbalanced entries ALWAYS produce UNBALANCED_TRANSACTION violation
// ============================================================================
// Invariant: any debit != credit pair fails the integrity check.
// This proves the check is complete (no false negatives on invalid data).

func TestProperty_LedgerBalance_UnbalancedEntries_AlwaysViolated(t *testing.T) {
	prop := func(debit, credit uint32) bool {
		if debit == credit || debit == 0 || credit == 0 {
			return true // skip: equal or zero amounts are degenerate
		}
		entries := []domain.TransactionEntry{
			{DebitAmount: decimal.NewFromInt(int64(debit)), CreditAmount: decimal.Zero},
			{DebitAmount: decimal.Zero, CreditAmount: decimal.NewFromInt(int64(credit))},
		}
		txn := &domain.Transaction{
			ID:                uuid.New(),
			TenantID:          uuid.New(),
			TransactionStatus: domain.TransactionStatusPosted,
		}
		pageCount := 0
		repo := &stubTxnRepo{
			fnList: func(_ context.Context, _ *domain.TransactionFilter) ([]*domain.Transaction, error) {
				pageCount++
				if pageCount == 1 {
					return []*domain.Transaction{txn}, nil
				}
				return nil, nil
			},
			fnGetEntriesByTxn: func(_ context.Context, _ uuid.UUID) ([]domain.TransactionEntry, error) {
				return entries, nil
			},
		}
		intSvc := service.NewIntegrityService(repo, nil, metrics.NewNoOpMetricsProvider())
		report, err := intSvc.ScanPostedTransactions(context.Background(), nil)
		if err != nil {
			return false
		}
		for _, v := range report.Violations {
			if v.Kind == "UNBALANCED_TRANSACTION" {
				return true // correctly detected
			}
		}
		return false // missed — invariant violated
	}
	if err := quick.Check(prop, &quick.Config{MaxCount: 300}); err != nil {
		t.Errorf("balance invariant false negative: %v", err)
	}
}

// ============================================================================
// FIN-PROP-003: Terminal states have NO outbound transitions
// ============================================================================
// Invariant: REVERSED and CANCELLED are absorbing states. Once reached,
// no further status transition is possible regardless of target.

func TestProperty_StateMachine_TerminalStates_NoOutTransitions(t *testing.T) {
	allStatuses := []domain.TransactionStatus{
		domain.TransactionStatusDraft,
		domain.TransactionStatusPendingApproval,
		domain.TransactionStatusApproved,
		domain.TransactionStatusRejected,
		domain.TransactionStatusPosted,
		domain.TransactionStatusReversed,
		domain.TransactionStatusCancelled,
	}
	terminals := []domain.TransactionStatus{
		domain.TransactionStatusReversed,
		domain.TransactionStatusCancelled,
	}
	for _, terminal := range terminals {
		for _, target := range allStatuses {
			if terminal.CanTransitionTo(target) {
				t.Errorf("terminal %q → %q allowed — absorbing state invariant violated", terminal, target)
			}
		}
	}
}

// ============================================================================
// FIN-PROP-004: IntegrityReport.HasCritical is monotonic
// ============================================================================
// Invariant: once a CRITICAL violation is in the report, adding more violations
// of any severity never removes the critical status.

func TestProperty_IntegrityReport_HasCritical_Monotonic(t *testing.T) {
	prop := func(extraCount uint8) bool {
		report := &service.IntegrityReport{
			Violations: []service.IntegrityViolation{
				{Kind: "UNBALANCED_TRANSACTION", Severity: service.SeverityCritical},
			},
		}
		if !report.HasCritical() {
			return false // must be critical before adding extras
		}
		for i := uint8(0); i < extraCount; i++ {
			report.Violations = append(report.Violations, service.IntegrityViolation{
				Kind:     "EXTRA_VIOLATION",
				Severity: service.SeverityHigh,
			})
		}
		return report.HasCritical() // must remain critical after additions
	}
	if err := quick.Check(prop, &quick.Config{MaxCount: 200}); err != nil {
		t.Errorf("HasCritical monotonicity violated: %v", err)
	}
}

// ============================================================================
// FIN-PROP-005: SafetyEnforcer blocks any amount strictly above MaxTransactionAmount
// ============================================================================

func TestProperty_SafetyEnforcer_AmountAboveMax_AlwaysBlocked(t *testing.T) {
	prop := func(maxAmount uint32) bool {
		if maxAmount == 0 {
			return true // zero max is degenerate
		}
		policy := service.SafetyPolicy{
			MaxTransactionAmount:       decimal.NewFromInt(int64(maxAmount)),
			MaxReversalsPerHour:        100,
			MaxApprovalVelocityPerHour: 100,
			MaxPostingsPerHour:         1000,
		}
		enforcer := service.NewSafetyEnforcer(policy, metrics.NewNoOpMetricsProvider())
		ctx := shared.WithTenantID(context.Background(), uuid.New())
		// amount = max + 1 — always above the limit
		over := decimal.NewFromInt(int64(maxAmount) + 1)
		return enforcer.CheckTransactionAmount(ctx, over) != nil
	}
	if err := quick.Check(prop, &quick.Config{MaxCount: 300}); err != nil {
		t.Errorf("SafetyEnforcer failed to block over-limit amount: %v", err)
	}
}

// ============================================================================
// FIN-PROP-006: SafetyEnforcer allows any amount at or below MaxTransactionAmount
// ============================================================================

func TestProperty_SafetyEnforcer_AmountAtOrBelowMax_AlwaysAllowed(t *testing.T) {
	prop := func(maxAmount uint32) bool {
		if maxAmount == 0 {
			return true
		}
		policy := service.SafetyPolicy{
			MaxTransactionAmount:       decimal.NewFromInt(int64(maxAmount)),
			MaxReversalsPerHour:        100,
			MaxApprovalVelocityPerHour: 100,
			MaxPostingsPerHour:         1000,
		}
		enforcer := service.NewSafetyEnforcer(policy, metrics.NewNoOpMetricsProvider())
		ctx := shared.WithTenantID(context.Background(), uuid.New())
		// amount = max — exactly at the limit must be allowed
		atLimit := decimal.NewFromInt(int64(maxAmount))
		return enforcer.CheckTransactionAmount(ctx, atLimit) == nil
	}
	if err := quick.Check(prop, &quick.Config{MaxCount: 300}); err != nil {
		t.Errorf("SafetyEnforcer incorrectly blocked at-limit amount: %v", err)
	}
}

// ============================================================================
// FIN-PROP-007: IntegrityReport with no violations is always healthy
// ============================================================================

func TestProperty_IntegrityReport_Empty_AlwaysHealthy(t *testing.T) {
	prop := func(unused uint8) bool {
		report := &service.IntegrityReport{}
		return !report.HasCritical()
	}
	if err := quick.Check(prop, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("empty report incorrectly marked critical: %v", err)
	}
}

// ============================================================================
// FIN-PROP-008: State machine DRAFT can always transition to CANCELLED (escape hatch)
// ============================================================================
// Invariant: DRAFT transactions must always be cancellable regardless of other state.

func TestProperty_StateMachine_Draft_AlwaysCancellable(t *testing.T) {
	if !domain.TransactionStatusDraft.CanTransitionTo(domain.TransactionStatusCancelled) {
		t.Error("DRAFT → CANCELLED not allowed — cancellation escape hatch missing")
	}
	if !domain.TransactionStatusApproved.CanTransitionTo(domain.TransactionStatusCancelled) {
		t.Error("APPROVED → CANCELLED not allowed — pre-post cancel missing")
	}
}
