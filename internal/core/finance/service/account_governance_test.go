// Package service_test — AccountMutationGovernor adversarial tests.
//
// Tests (FIN-GOV-*):
//
//	001 — ValidateStatusTransition: allowed transition passes
//	002 — ValidateStatusTransition: disallowed transition blocked
//	003 — ValidateStatusTransition: SOD violation (creator == approver)
//	004 — ValidateStatusTransition: frozen account blocks non-permitted exit
//	005 — ValidateCurrencyChange: blocked after last transaction date
//	006 — ValidateCurrencyChange: empty currency rejected
//	007 — ValidateCurrencyChange: invalid length rejected
//	008 — ValidateCurrencyChange: allowed when no transactions
//	009 — ValidateRootTypeChange: blocked after last transaction date
//	010 — ValidateRootTypeChange: invalid root type rejected
//	011 — ValidateArchival: non-zero balance blocked
//	012 — ValidateArchival: has-children blocked
//	013 — ValidateArchival: zero balance + no children passes
//	014 — ValidateDeletion: system account blocked
//	015 — ValidateDeletion: account with transactions blocked
//	016 — ValidateDeletion: clean account passes
//	017 — RequiresApproval: CRITICAL always requires approval
//	018 — RequiresApproval: HIGH requires approval only when transactions exist
//	019 — ValidateMerge: self-merge blocked
//	020 — ValidateMerge: non-zero source balance blocked
//	021 — Nil governor — all methods return nil (nil-safe)
package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"awo.so/internal/core/finance/domain"
	"awo.so/internal/core/finance/service"
	"awo.so/internal/shared"
	"awo.so/internal/shared/metrics"
)

// =============================================================================
// Helpers
// =============================================================================

func govCtx() context.Context {
	return shared.WithTenantID(context.Background(), uuid.New())
}

// mockGetByIDRepo is a single-account AccountsRepository stub for governance tests.
type mockGetByIDRepo struct {
	p16StubAccountRepo
	target *domain.Accounts
}

func (r *mockGetByIDRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Accounts, error) {
	if r.target != nil && r.target.ID == id {
		return r.target, nil
	}
	return nil, errors.New("account not found")
}

func newGovAccount() *domain.Accounts {
	code := "10010001"
	curr := "USD"
	return &domain.Accounts{
		ID:               uuid.New(),
		TenantID:         uuid.New(),
		AccountCode:      code,
		AccountName:      "Gov Test Account",
		RootType:         domain.RootTypeAsset,
		NormalBalance:    domain.NormalBalanceDebit,
		AccountType:      "Current Asset",
		Status:           domain.AccountStatusDraft,
		ValidationStatus: domain.ValidationStatusValid,
		CurrencyCode:     &curr,
		CurrentBalance:   decimal.Zero,
		CreatedBy:        uuid.New(),
		CreatedAt:        time.Now(),
	}
}

func newGovernor(target *domain.Accounts) *service.AccountMutationGovernor {
	return service.NewAccountMutationGovernor(
		&mockGetByIDRepo{target: target},
		metrics.NewNoOpMetricsProvider(),
	)
}

// =============================================================================
// FIN-GOV-001: Allowed transition passes
// =============================================================================

func TestGovernor_ValidateStatusTransition_AllowedPasses(t *testing.T) {
	account := newGovAccount()
	account.Status = domain.AccountStatusDraft
	g := newGovernor(account)

	if err := g.ValidateStatusTransition(govCtx(), account, domain.AccountStatusPendingApproval, uuid.New()); err != nil {
		t.Errorf("FIN-GOV-001: expected nil for allowed transition, got %v", err)
	}
}

// =============================================================================
// FIN-GOV-002: Disallowed transition blocked
// =============================================================================

func TestGovernor_ValidateStatusTransition_DisallowedBlocked(t *testing.T) {
	account := newGovAccount()
	account.Status = domain.AccountStatusDraft
	g := newGovernor(account)

	// DRAFT → ARCHIVED is not in AllowedTransitions.
	err := g.ValidateStatusTransition(govCtx(), account, domain.AccountStatusArchived, uuid.New())
	if err == nil {
		t.Error("FIN-GOV-002: expected error for disallowed transition, got nil")
	}
	if !errors.Is(err, domain.ErrAccountStatusTransition) {
		t.Errorf("FIN-GOV-002: expected ErrAccountStatusTransition, got %v", err)
	}
}

// =============================================================================
// FIN-GOV-003: SOD violation blocks activation
// =============================================================================

func TestGovernor_ValidateStatusTransition_SODViolation(t *testing.T) {
	creator := uuid.New()
	account := newGovAccount()
	account.Status = domain.AccountStatusPendingApproval
	account.CreatedBy = creator
	g := newGovernor(account)

	err := g.ValidateStatusTransition(govCtx(), account, domain.AccountStatusActive, creator)
	if err == nil {
		t.Error("FIN-GOV-003: expected SOD error, got nil")
	}
	if !errors.Is(err, domain.ErrSelfApprovalNotAllowed) {
		t.Errorf("FIN-GOV-003: expected ErrSelfApprovalNotAllowed, got %v", err)
	}
}

// =============================================================================
// FIN-GOV-004: Frozen account — only ACTIVE and RESTRICTED exits are permitted
// =============================================================================

func TestGovernor_ValidateStatusTransition_FrozenBlocksNonPermitted(t *testing.T) {
	account := newGovAccount()
	account.Status = domain.AccountStatusFrozen
	g := newGovernor(account)

	// FROZEN → CLOSED: state machine blocks this before the frozen guard runs.
	// Both state machine and frozen guard agree: the transition must be rejected.
	err := g.ValidateStatusTransition(govCtx(), account, domain.AccountStatusClosed, uuid.New())
	if err == nil {
		t.Error("FIN-GOV-004: expected error for FROZEN → CLOSED, got nil")
	}
	// State machine fires first (ErrAccountStatusTransition); frozen guard fires for
	// any state-machine-allowed exit that is not ACTIVE or RESTRICTED.
	// Either sentinel is acceptable here — the important invariant is rejection.
	if !errors.Is(err, domain.ErrAccountStatusTransition) && !errors.Is(err, domain.ErrAccountFrozen) {
		t.Errorf("FIN-GOV-004: expected ErrAccountStatusTransition or ErrAccountFrozen, got %v", err)
	}
}

// =============================================================================
// FIN-GOV-005: ValidateCurrencyChange blocked after posting
// =============================================================================

func TestGovernor_ValidateCurrencyChange_BlockedAfterPosting(t *testing.T) {
	account := newGovAccount()
	posted := time.Now().Add(-24 * time.Hour)
	account.LastTransactionDate = &posted
	g := newGovernor(account)

	err := g.ValidateCurrencyChange(govCtx(), account, "EUR")
	if err == nil {
		t.Error("FIN-GOV-005: expected error for currency change after posting, got nil")
	}
	if !errors.Is(err, domain.ErrCurrencyChangeForbidden) {
		t.Errorf("FIN-GOV-005: expected ErrCurrencyChangeForbidden, got %v", err)
	}
}

// =============================================================================
// FIN-GOV-006: ValidateCurrencyChange empty currency rejected
// =============================================================================

func TestGovernor_ValidateCurrencyChange_EmptyCurrencyRejected(t *testing.T) {
	account := newGovAccount()
	g := newGovernor(account)

	err := g.ValidateCurrencyChange(govCtx(), account, "")
	if err == nil {
		t.Error("FIN-GOV-006: expected error for empty currency, got nil")
	}
}

// =============================================================================
// FIN-GOV-007: ValidateCurrencyChange invalid length rejected
// =============================================================================

func TestGovernor_ValidateCurrencyChange_InvalidLengthRejected(t *testing.T) {
	account := newGovAccount()
	g := newGovernor(account)

	err := g.ValidateCurrencyChange(govCtx(), account, "US") // 2 chars, must be 3
	if err == nil {
		t.Error("FIN-GOV-007: expected error for 2-char currency, got nil")
	}
}

// =============================================================================
// FIN-GOV-008: ValidateCurrencyChange allowed when no transactions
// =============================================================================

func TestGovernor_ValidateCurrencyChange_AllowedWhenNoTransactions(t *testing.T) {
	account := newGovAccount() // LastTransactionDate == nil
	g := newGovernor(account)

	if err := g.ValidateCurrencyChange(govCtx(), account, "EUR"); err != nil {
		t.Errorf("FIN-GOV-008: expected nil for first-time currency change, got %v", err)
	}
}

// =============================================================================
// FIN-GOV-009: ValidateRootTypeChange blocked after posting
// =============================================================================

func TestGovernor_ValidateRootTypeChange_BlockedAfterPosting(t *testing.T) {
	account := newGovAccount()
	posted := time.Now().Add(-24 * time.Hour)
	account.LastTransactionDate = &posted
	g := newGovernor(account)

	err := g.ValidateRootTypeChange(govCtx(), account, domain.RootTypeExpense)
	if err == nil {
		t.Error("FIN-GOV-009: expected error for root type change after posting, got nil")
	}
	if !errors.Is(err, domain.ErrRootTypeChangeForbidden) {
		t.Errorf("FIN-GOV-009: expected ErrRootTypeChangeForbidden, got %v", err)
	}
}

// =============================================================================
// FIN-GOV-010: ValidateRootTypeChange invalid root type rejected
// =============================================================================

func TestGovernor_ValidateRootTypeChange_InvalidRootTypeRejected(t *testing.T) {
	account := newGovAccount()
	g := newGovernor(account)

	err := g.ValidateRootTypeChange(govCtx(), account, domain.RootType("BOGUS"))
	if err == nil {
		t.Error("FIN-GOV-010: expected error for invalid root type, got nil")
	}
	if !errors.Is(err, domain.ErrInvalidRootType) {
		t.Errorf("FIN-GOV-010: expected ErrInvalidRootType, got %v", err)
	}
}

// =============================================================================
// FIN-GOV-011: ValidateArchival non-zero balance blocked
// =============================================================================

func TestGovernor_ValidateArchival_NonZeroBalanceBlocked(t *testing.T) {
	account := newGovAccount()
	account.CurrentBalance = decimal.NewFromFloat(100.00)
	g := newGovernor(account)

	err := g.ValidateArchival(govCtx(), account)
	if err == nil {
		t.Error("FIN-GOV-011: expected error for non-zero balance, got nil")
	}
	if !errors.Is(err, domain.ErrAccountHasNonZeroBalance) {
		t.Errorf("FIN-GOV-011: expected ErrAccountHasNonZeroBalance, got %v", err)
	}
}

// =============================================================================
// FIN-GOV-012: ValidateArchival has-children blocked
// =============================================================================

func TestGovernor_ValidateArchival_HasChildrenBlocked(t *testing.T) {
	account := newGovAccount()
	account.HasChildren = true
	g := newGovernor(account)

	err := g.ValidateArchival(govCtx(), account)
	if err == nil {
		t.Error("FIN-GOV-012: expected error for account with children, got nil")
	}
	if !errors.Is(err, domain.ErrAccountHasChildren) {
		t.Errorf("FIN-GOV-012: expected ErrAccountHasChildren, got %v", err)
	}
}

// =============================================================================
// FIN-GOV-013: ValidateArchival zero balance + no children passes
// =============================================================================

func TestGovernor_ValidateArchival_ZeroBalanceNoChildren_Passes(t *testing.T) {
	account := newGovAccount()
	account.CurrentBalance = decimal.Zero
	account.HasChildren = false
	g := newGovernor(account)

	if err := g.ValidateArchival(govCtx(), account); err != nil {
		t.Errorf("FIN-GOV-013: expected nil for archivable account, got %v", err)
	}
}

// =============================================================================
// FIN-GOV-014: ValidateDeletion system account blocked
// =============================================================================

func TestGovernor_ValidateDeletion_SystemAccountBlocked(t *testing.T) {
	account := newGovAccount()
	account.IsSystemAccount = true
	g := newGovernor(account)

	err := g.ValidateDeletion(govCtx(), account)
	if err == nil {
		t.Error("FIN-GOV-014: expected error for system account deletion, got nil")
	}
}

// =============================================================================
// FIN-GOV-015: ValidateDeletion account with transactions blocked
// =============================================================================

func TestGovernor_ValidateDeletion_AccountWithTransactionsBlocked(t *testing.T) {
	account := newGovAccount()
	posted := time.Now().Add(-24 * time.Hour)
	account.LastTransactionDate = &posted
	g := newGovernor(account)

	err := g.ValidateDeletion(govCtx(), account)
	if err == nil {
		t.Error("FIN-GOV-015: expected error for account with transactions, got nil")
	}
	if !errors.Is(err, domain.ErrAccountHasTransactions) {
		t.Errorf("FIN-GOV-015: expected ErrAccountHasTransactions, got %v", err)
	}
}

// =============================================================================
// FIN-GOV-016: ValidateDeletion clean account passes
// =============================================================================

func TestGovernor_ValidateDeletion_CleanAccountPasses(t *testing.T) {
	account := newGovAccount()
	// No transactions, not system.
	g := newGovernor(account)

	if err := g.ValidateDeletion(govCtx(), account); err != nil {
		t.Errorf("FIN-GOV-016: expected nil for deletable account, got %v", err)
	}
}

// =============================================================================
// FIN-GOV-017: RequiresApproval — CRITICAL always
// =============================================================================

func TestGovernor_RequiresApproval_CriticalAlwaysTrue(t *testing.T) {
	account := newGovAccount()
	g := newGovernor(account)

	criticalTypes := []string{"root_type_change", "merge", "close", "archive", "delete"}
	for _, mut := range criticalTypes {
		if !g.RequiresApproval(account, mut) {
			t.Errorf("FIN-GOV-017: CRITICAL mutation %q must always require approval", mut)
		}
	}
}

// =============================================================================
// FIN-GOV-018: RequiresApproval — HIGH only when transactions exist
// =============================================================================

func TestGovernor_RequiresApproval_HighOnlyWhenTransactions(t *testing.T) {
	account := newGovAccount()
	g := newGovernor(account)

	// No transactions — HIGH should NOT require approval.
	if g.RequiresApproval(account, "currency_change") {
		t.Error("FIN-GOV-018: HIGH mutation should not require approval when no transactions")
	}

	// With transactions — HIGH SHOULD require approval.
	posted := time.Now().Add(-24 * time.Hour)
	account.LastTransactionDate = &posted
	if !g.RequiresApproval(account, "currency_change") {
		t.Error("FIN-GOV-018: HIGH mutation should require approval when transactions exist")
	}
}

// =============================================================================
// FIN-GOV-019: ValidateMerge self-merge blocked
// =============================================================================

func TestGovernor_ValidateMerge_SelfMergeBlocked(t *testing.T) {
	account := newGovAccount()
	g := newGovernor(account)

	err := g.ValidateMerge(govCtx(), account, account.ID, uuid.New())
	if err == nil {
		t.Error("FIN-GOV-019: expected error for self-merge, got nil")
	}
}

// =============================================================================
// FIN-GOV-020: ValidateMerge non-zero source balance blocked
// =============================================================================

func TestGovernor_ValidateMerge_NonZeroSourceBlocked(t *testing.T) {
	source := newGovAccount()
	source.CurrentBalance = decimal.NewFromFloat(100.00)

	target := newGovAccount()
	target.RootType = source.RootType

	repo := &mockGetByIDRepo{target: target}
	g := service.NewAccountMutationGovernor(repo, metrics.NewNoOpMetricsProvider())

	err := g.ValidateMerge(govCtx(), source, target.ID, uuid.New())
	if err == nil {
		t.Error("FIN-GOV-020: expected error for non-zero source balance in merge, got nil")
	}
	if !errors.Is(err, domain.ErrAccountHasNonZeroBalance) {
		t.Errorf("FIN-GOV-020: expected ErrAccountHasNonZeroBalance, got %v", err)
	}
}

// =============================================================================
// FIN-GOV-021: Nil governor — nil-safe
// =============================================================================

func TestGovernor_NilReceiver_Safe(t *testing.T) {
	var g *service.AccountMutationGovernor
	account := newGovAccount()
	ctx := govCtx()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("FIN-GOV-021: nil governor panicked: %v", r)
		}
	}()

	_ = g.ValidateStatusTransition(ctx, account, domain.AccountStatusActive, uuid.New())
	_ = g.ValidateCurrencyChange(ctx, account, "EUR")
	_ = g.ValidateRootTypeChange(ctx, account, domain.RootTypeExpense)
	_ = g.ValidateArchival(ctx, account)
	_ = g.ValidateDeletion(ctx, account)
	_ = g.ValidateMerge(ctx, account, uuid.New(), uuid.New())
	if g.RequiresApproval(account, "delete") {
		t.Error("FIN-GOV-021: nil governor RequiresApproval should return false")
	}
}
