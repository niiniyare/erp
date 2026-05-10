// Package service — account mutation governor.
//
// AccountMutationGovernor is the single enforcement point for all high-risk
// chart-of-accounts mutations. Every service method that modifies account
// structure MUST call the relevant governor method before persisting.
//
// Design principle:
//   - Fail closed: when in doubt, reject.
//   - All blocking logic is in one place — not scattered across CRUD methods.
//   - Risk levels (LOW/MEDIUM/HIGH/CRITICAL) drive approval requirements.
//
// Nil-safe: a nil *AccountMutationGovernor passes all checks silently (for
// tests that don't need governance enforcement).
package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"awo.so/internal/core/finance/domain"
	"awo.so/internal/shared"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
)

// AccountMutationGovernor enforces governance rules on high-risk account mutations.
// All methods are nil-safe.
type AccountMutationGovernor struct {
	accountRepo domain.AccountsRepository
	metrics     metrics.MetricsProvider
}

// NewAccountMutationGovernor creates a new governor.
func NewAccountMutationGovernor(
	accountRepo domain.AccountsRepository,
	m metrics.MetricsProvider,
) *AccountMutationGovernor {
	return &AccountMutationGovernor{
		accountRepo: accountRepo,
		metrics:     m,
	}
}

// =============================================================================
// Status transition enforcement
// =============================================================================

// ValidateStatusTransition checks that the proposed status transition is
// permitted by the state machine and that any SOD constraints are satisfied.
//
// Rules enforced:
//  1. Transition must appear in domain.AllowedTransitions.
//  2. actorID must not equal account.CreatedBy when approving (SOD).
//  3. account must not be frozen unless the target is a terminal/hold state.
func (g *AccountMutationGovernor) ValidateStatusTransition(
	ctx context.Context,
	account *domain.Accounts,
	to domain.AccountStatus,
	actorID uuid.UUID,
) error {
	if g == nil {
		return nil
	}

	// 1. State machine check.
	if _, err := account.TransitionTo(to); err != nil {
		g.recordViolation(ctx, "status_transition_blocked", account.ID)
		return fmt.Errorf("%w: %s → %s", domain.ErrAccountStatusTransition, account.Status, to)
	}

	// 2. SOD: approver cannot be the creator.
	if to == domain.AccountStatusActive && account.Status == domain.AccountStatusPendingApproval {
		if actorID == account.CreatedBy {
			g.recordViolation(ctx, "sod_violation", account.ID)
			return domain.ErrSelfApprovalNotAllowed
		}
	}

	// 3. Frozen accounts block all transitions except emergency holds.
	if account.Status == domain.AccountStatusFrozen {
		// FROZEN → ACTIVE and FROZEN → RESTRICTED are the only permitted exits.
		if to != domain.AccountStatusActive && to != domain.AccountStatusRestricted {
			g.recordViolation(ctx, "frozen_mutation_blocked", account.ID)
			return fmt.Errorf("%w: account is frozen; only ACTIVE or RESTRICTED transitions are permitted", domain.ErrAccountFrozen)
		}
	}

	return nil
}

// =============================================================================
// Currency mutation enforcement
// =============================================================================

// ValidateCurrencyChange blocks currency changes after any transaction has been
// posted to the account. Currency changes after postings corrupt historical
// balance reporting and FX revaluation calculations.
func (g *AccountMutationGovernor) ValidateCurrencyChange(
	ctx context.Context,
	account *domain.Accounts,
	newCurrency string,
) error {
	if g == nil {
		return nil
	}

	if account.LastTransactionDate != nil {
		g.recordViolation(ctx, "currency_change_after_posting", account.ID)
		return fmt.Errorf("%w: account %s has transactions posted since %s",
			domain.ErrCurrencyChangeForbidden, account.AccountCode, account.LastTransactionDate.Format("2006-01-02"))
	}

	if newCurrency == "" {
		return fmt.Errorf("new currency code is required")
	}
	if len(newCurrency) != domain.CurrencyCodeLen {
		return fmt.Errorf("currency code must be exactly %d characters (ISO 4217)", domain.CurrencyCodeLen)
	}

	return nil
}

// =============================================================================
// Root type mutation enforcement
// =============================================================================

// ValidateRootTypeChange blocks root type changes after any transaction has been
// posted. Root type determines the normal balance and financial statement
// placement — changing it after postings produces incorrect comparative reports.
func (g *AccountMutationGovernor) ValidateRootTypeChange(
	ctx context.Context,
	account *domain.Accounts,
	newRootType domain.RootType,
) error {
	if g == nil {
		return nil
	}

	if account.LastTransactionDate != nil {
		g.recordViolation(ctx, "root_type_change_after_posting", account.ID)
		return fmt.Errorf("%w: account %s has transactions posted since %s",
			domain.ErrRootTypeChangeForbidden, account.AccountCode, account.LastTransactionDate.Format("2006-01-02"))
	}

	if !newRootType.IsValid() {
		return domain.ErrInvalidRootType
	}

	return nil
}

// =============================================================================
// Archival / closure enforcement
// =============================================================================

// ValidateArchival blocks archival of accounts that still carry a balance.
// An account with a non-zero balance cannot be closed because that balance
// would be invisible to reporting once the account is archived.
func (g *AccountMutationGovernor) ValidateArchival(
	ctx context.Context,
	account *domain.Accounts,
) error {
	if g == nil {
		return nil
	}

	if !account.CurrentBalance.IsZero() {
		g.recordViolation(ctx, "archival_with_balance", account.ID)
		return fmt.Errorf("%w: balance is %s (must be zero before closing)",
			domain.ErrAccountHasNonZeroBalance, account.CurrentBalance.String())
	}

	if account.HasChildren {
		return fmt.Errorf("%w: cannot archive an account that has child accounts",
			domain.ErrAccountHasChildren)
	}

	return nil
}

// =============================================================================
// Deletion enforcement
// =============================================================================

// ValidateDeletion enforces the invariants that prevent silent account deletion:
//  1. System accounts are protected forever.
//  2. Accounts that have ever had a posting cannot be deleted (soft-delete allowed,
//     but the domain check requires no historical postings).
func (g *AccountMutationGovernor) ValidateDeletion(
	ctx context.Context,
	account *domain.Accounts,
) error {
	if g == nil {
		return nil
	}

	if !account.CanBeDeleted() {
		if account.IsSystemAccount {
			g.recordViolation(ctx, "system_account_deletion_blocked", account.ID)
			return fmt.Errorf("system accounts cannot be deleted: account %s is a protected system account", account.AccountCode)
		}
		if account.LastTransactionDate != nil {
			g.recordViolation(ctx, "posted_account_deletion_blocked", account.ID)
			return fmt.Errorf("%w: account %s has transactions posted (last: %s)",
				domain.ErrAccountHasTransactions, account.AccountCode, account.LastTransactionDate.Format("2006-01-02"))
		}
	}

	return nil
}

// =============================================================================
// Hierarchy reparenting enforcement
// =============================================================================

// ValidateHierarchyReparent enforces all hierarchy invariants before an account
// is moved to a new parent. This is a CRITICAL-risk operation.
//
// Invariants checked:
//  1. Account cannot be its own parent.
//  2. New parent must exist and be in the same tenant.
//  3. Root types must be compatible (parent == child).
//  4. The move must not create a cycle.
//  5. Resulting depth must not exceed MaxAccountHierarchyDepth.
func (g *AccountMutationGovernor) ValidateHierarchyReparent(
	ctx context.Context,
	account *domain.Accounts,
	newParentID uuid.UUID,
) error {
	if g == nil {
		return nil
	}

	tenantID, _ := shared.GetTenantID(ctx)

	// 1. Self-parent check.
	if account.ID == newParentID {
		g.recordViolation(ctx, "self_parent", account.ID)
		return domain.ErrHierarchySelfParent
	}

	// 2. Load and validate new parent.
	newParent, err := g.accountRepo.GetByID(ctx, newParentID)
	if err != nil {
		return fmt.Errorf("new parent account not found: %w", err)
	}

	// 3. Tenant isolation.
	if newParent.TenantID != tenantID {
		g.recordViolation(ctx, "tenant_mismatch", account.ID)
		return domain.ErrHierarchyTenantMismatch
	}

	// 4. Root type compatibility.
	if newParent.RootType != account.RootType {
		g.recordViolation(ctx, "root_type_mismatch", account.ID)
		return fmt.Errorf("%w: account is %s but new parent is %s",
			domain.ErrHierarchyRootTypeMismatch, account.RootType, newParent.RootType)
	}

	// 5. Cycle detection: the new parent must NOT be a descendant of the account.
	// If newParent.Path contains account.ID, moving account under newParent creates a cycle.
	if newParent.IsDescendantOf(account.ID) {
		g.recordViolation(ctx, "cycle_detected", account.ID)
		return fmt.Errorf("%w: moving account %s under %s creates a circular reference",
			domain.ErrHierarchyCycleDetected, account.AccountCode, newParent.AccountCode)
	}

	// 6. Depth check: new depth would be newParent.AccountLevel + 1.
	newDepth := int(newParent.AccountLevel) + 1
	if newDepth > domain.MaxAccountHierarchyDepth {
		g.recordViolation(ctx, "depth_exceeded", account.ID)
		return fmt.Errorf("%w: new depth %d exceeds maximum %d",
			domain.ErrHierarchyDepthExceeded, newDepth, domain.MaxAccountHierarchyDepth)
	}

	logger.InfoContext(ctx, "hierarchy reparent validated", logger.Fields{
		"account_id":    account.ID.String(),
		"new_parent_id": newParentID.String(),
		"new_depth":     newDepth,
	})

	return nil
}

// =============================================================================
// Risk classification
// =============================================================================

// ClassifyMutationRisk returns the risk level for a proposed mutation type.
// Service methods should call this to determine whether approval is required.
func (g *AccountMutationGovernor) ClassifyMutationRisk(mutationType string) domain.AccountMutationRisk {
	return domain.MutationRiskOf(mutationType)
}

// RequiresApproval returns true when the mutation type requires explicit approval
// before it can be persisted. CRITICAL mutations always require approval.
// HIGH mutations require approval when the account already has postings.
func (g *AccountMutationGovernor) RequiresApproval(
	account *domain.Accounts,
	mutationType string,
) bool {
	if g == nil {
		return false
	}
	risk := domain.MutationRiskOf(mutationType)
	switch risk {
	case domain.RiskCritical:
		return true
	case domain.RiskHigh:
		// HIGH mutations require approval when there is financial history.
		return account.LastTransactionDate != nil
	default:
		return false
	}
}

// ValidateNormalBalanceChange blocks changes to normal balance after postings.
// The normal balance is derived from root type; this guard prevents the root
// type and normal balance from being set independently post-hoc.
func (g *AccountMutationGovernor) ValidateNormalBalanceChange(
	ctx context.Context,
	account *domain.Accounts,
	newBalance domain.NormalBalance,
) error {
	if g == nil {
		return nil
	}

	if account.LastTransactionDate != nil {
		g.recordViolation(ctx, "normal_balance_change_after_posting", account.ID)
		return fmt.Errorf("%w: cannot change normal balance after transactions have been posted",
			domain.ErrMutationBlockedAfterPosting)
	}

	expected := domain.GetNormalBalanceForRootType(account.RootType)
	if newBalance != expected {
		return fmt.Errorf("normal balance %s is inconsistent with root type %s (expected %s)",
			newBalance, account.RootType, expected)
	}

	return nil
}

// ValidateMerge enforces preconditions for account merge operations.
// Merge is CRITICAL risk: source account must have zero balance after migration.
func (g *AccountMutationGovernor) ValidateMerge(
	ctx context.Context,
	source *domain.Accounts,
	targetID uuid.UUID,
	actorID uuid.UUID,
) error {
	if g == nil {
		return nil
	}

	if source.ID == targetID {
		return fmt.Errorf("cannot merge account into itself")
	}
	if !source.CurrentBalance.IsZero() {
		return fmt.Errorf("%w: source account %s must have zero balance before merge (current: %s)",
			domain.ErrAccountHasNonZeroBalance, source.AccountCode, source.CurrentBalance.String())
	}
	if source.IsSystemAccount {
		return fmt.Errorf("system accounts cannot be merged: account %s is protected", source.AccountCode)
	}

	target, err := g.accountRepo.GetByID(ctx, targetID)
	if err != nil {
		return fmt.Errorf("merge target account not found: %w", err)
	}
	if target.RootType != source.RootType {
		return fmt.Errorf("cannot merge accounts with different root types (%s → %s)",
			source.RootType, target.RootType)
	}
	if target.Status == domain.AccountStatusClosed || target.Status == domain.AccountStatusArchived {
		return fmt.Errorf("cannot merge into a closed/archived account (target status: %s)", target.Status)
	}

	g.recordMutation(ctx, "merge_validated", source.ID, decimal.Zero)
	return nil
}

// =============================================================================
// Internal helpers
// =============================================================================

func (g *AccountMutationGovernor) recordViolation(ctx context.Context, kind string, accountID uuid.UUID) {
	if g.metrics != nil {
		g.metrics.IncrementCounter("finance_account_governance_violations_total", metrics.Fields{
			"kind": kind,
		})
	}
	logger.WarnContext(ctx, "account mutation governance violation", logger.Fields{
		"kind":       kind,
		"account_id": accountID.String(),
	})
}

func (g *AccountMutationGovernor) recordMutation(ctx context.Context, kind string, accountID uuid.UUID, _ decimal.Decimal) {
	if g.metrics != nil {
		g.metrics.IncrementCounter("finance_account_mutations_total", metrics.Fields{
			"kind": kind,
		})
	}
	logger.InfoContext(ctx, "account mutation recorded", logger.Fields{
		"kind":       kind,
		"account_id": accountID.String(),
	})
}
