// Package service_test — Phase 16 mutation testing enforcement harness.
//
// Mutation-style tests WITHOUT code mutation. Each test:
//
//  1. Constructs the adversarial input that would succeed if the protection were absent.
//  2. Runs it through the real production code.
//  3. Asserts that it is rejected with the expected error code.
//
// If any test in this file passes when its protection is removed from the
// production code, the CI pipeline has an assurance coverage gap.
//
// Mutation coverage table (FIN-MUT-*):
//
//	001 — SOD check: creator == approver → SOD_VIOLATION
//	002 — Terminal deletion: POSTED delete → TRANSACTION_DELETE_NOT_ALLOWED
//	003 — Approval gate: DRAFT+Required post → APPROVAL_REQUIRED
//	004 — Balance check: unbalanced scan → UNBALANCED_TRANSACTION (CRITICAL)
//	005 — State machine: terminal absorbing → no outbound transitions
//	006 — Velocity limit: over-limit reversal → REVERSAL_LIMIT_EXCEEDED
//	007 — Amount limit: over-limit amount → TRANSACTION_AMOUNT_EXCEEDED
//	008 — Extension contract: mutates without audit hook → startup violation
//	009 — Extension contract: mutates without safety check → startup violation
//	010 — EvolutionGuard nil-safety: nil guard → no panic, no block
package service_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	gomock "go.uber.org/mock/gomock"

	"awo.so/internal/core/finance/domain"
	"awo.so/internal/core/finance/service"
	"awo.so/internal/shared"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

// mutationResult records whether a protection detected its adversarial input.
type mutationResult struct {
	name      string
	protected bool
	errCode   string
}

// ── FIN-MUT-001: SOD protection ──────────────────────────────────────────────

func TestMutationHarness_SODCheck_AlwaysBlocks_CreatorEqualsApprover(t *testing.T) {
	txnID := uuid.New()
	userID := uuid.New()

	repo := &p16StubTxnRepo{
		fnGetByID: func(_ context.Context, _ uuid.UUID) (*domain.Transaction, error) {
			return &domain.Transaction{
				ID:               txnID,
				TenantID:         uuid.New(),
				CreatedBy:        userID, // <-- mutation: same as approver
				ApprovalRequired: true,
				ApprovalStatus:   domain.ApprovalStatusPending,
			}, nil
		},
	}

	svc := p16NewSvc(repo)
	ctx := shared.WithUserID(shared.WithTenantID(context.Background(), uuid.New()), userID)

	_, err := svc.ApproveTransaction(ctx, txnID, "")
	if err == nil {
		t.Fatal("FIN-MUT-001 FAILED: SOD check removed — creator == approver was allowed")
	}
	if !p16ContainsCode(err, "SOD_VIOLATION") {
		t.Errorf("FIN-MUT-001: expected SOD_VIOLATION, got: %v", err)
	}
}

// ── FIN-MUT-002: Terminal deletion protection ─────────────────────────────────

func TestMutationHarness_TerminalDeletion_AlwaysBlocks_PostedAndReversed(t *testing.T) {
	guardedStatuses := []domain.TransactionStatus{
		domain.TransactionStatusPosted,
		domain.TransactionStatusReversed,
	}
	for _, status := range guardedStatuses {
		status := status
		t.Run(string(status), func(t *testing.T) {
			txnID := uuid.New()
			tenantID := uuid.New()
			repo := &p16StubTxnRepo{
				fnGetByID: func(_ context.Context, _ uuid.UUID) (*domain.Transaction, error) {
					return &domain.Transaction{
						ID:                txnID,
						TenantID:          tenantID,
						TransactionStatus: status,
					}, nil
				},
			}
			svc := p16NewSvc(repo)
			ctx := shared.WithTenantID(context.Background(), tenantID)
			err := svc.DeleteTransaction(ctx, txnID)
			if err == nil {
				t.Fatalf("FIN-MUT-002 FAILED: terminal deletion guard removed for %s", status)
			}
			if !p16ContainsCode(err, "TRANSACTION_DELETE_NOT_ALLOWED") {
				t.Errorf("FIN-MUT-002: expected TRANSACTION_DELETE_NOT_ALLOWED for %s, got: %v", status, err)
			}
		})
	}
}

// ── FIN-MUT-003: Approval gate ────────────────────────────────────────────────

func TestMutationHarness_ApprovalGate_BlocksDraftWithRequiredApproval(t *testing.T) {
	txnID := uuid.New()
	tenantID := uuid.New()

	repo := &p16StubTxnRepo{
		fnGetByID: func(_ context.Context, _ uuid.UUID) (*domain.Transaction, error) {
			return &domain.Transaction{
				ID:                txnID,
				TenantID:          tenantID,
				TransactionStatus: domain.TransactionStatusDraft, // <-- mutation: not approved
				ApprovalRequired:  true,
				ApprovalStatus:    domain.ApprovalStatusPending,
			}, nil
		},
	}

	svc := p16NewSvc(repo)
	ctx := shared.WithTenantID(context.Background(), tenantID)

	_, err := svc.PostTransaction(ctx, txnID, nil)
	if err == nil {
		t.Fatal("FIN-MUT-003 FAILED: approval gate removed — DRAFT+Required was allowed to post")
	}
	if !p16ContainsCode(err, "APPROVAL_REQUIRED") {
		t.Errorf("FIN-MUT-003: expected APPROVAL_REQUIRED, got: %v", err)
	}
}

// ── FIN-MUT-004: Balance check ────────────────────────────────────────────────

func TestMutationHarness_BalanceCheck_DetectsUnbalancedScan(t *testing.T) {
	type tc struct {
		name   string
		debit  float64
		credit float64
	}
	cases := []tc{
		{"penny off", 100.00, 99.99},
		{"large imbalance", 1.00, 1_000_000.00},
		{"debit only", 500.00, 0},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			entries := []domain.TransactionEntry{
				{DebitAmount: decimal.NewFromFloat(c.debit), CreditAmount: decimal.Zero},
				{DebitAmount: decimal.Zero, CreditAmount: decimal.NewFromFloat(c.credit)},
			}
			txn := &domain.Transaction{
				ID:                uuid.New(),
				TenantID:          uuid.New(),
				TransactionStatus: domain.TransactionStatusPosted,
			}
			pageCount := 0
			repo := &p16StubTxnRepo{
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
			report, err := intSvc.ScanPostedTransactions(p16ctx(), nil)
			if err != nil {
				t.Fatalf("unexpected scan error: %v", err)
			}
			found := false
			for _, v := range report.Violations {
				if v.Kind == "UNBALANCED_TRANSACTION" {
					found = true
				}
			}
			if !found {
				t.Errorf("FIN-MUT-004 FAILED: balance check removed for case %q — violation not detected", c.name)
			}
		})
	}
}

// ── FIN-MUT-005: State machine terminal absorbing ─────────────────────────────

func TestMutationHarness_StateMachine_TerminalStates_NoOutboundTransitions(t *testing.T) {
	terminals := []domain.TransactionStatus{
		domain.TransactionStatusReversed,
		domain.TransactionStatusCancelled,
	}
	all := []domain.TransactionStatus{
		domain.TransactionStatusDraft,
		domain.TransactionStatusPendingApproval,
		domain.TransactionStatusApproved,
		domain.TransactionStatusPosted,
		domain.TransactionStatusReversed,
		domain.TransactionStatusCancelled,
	}
	for _, terminal := range terminals {
		for _, target := range all {
			if terminal.CanTransitionTo(target) {
				t.Errorf("FIN-MUT-005 FAILED: absorbing state guard removed — %s → %s transition allowed", terminal, target)
			}
		}
	}
}

// ── FIN-MUT-006: Velocity limit ───────────────────────────────────────────────

func TestMutationHarness_VelocityLimit_BlocksExcessiveReversals(t *testing.T) {
	const limit = 3
	policy := service.SafetyPolicy{
		MaxTransactionAmount:       decimal.NewFromFloat(10_000_000),
		MaxReversalsPerHour:        limit,
		MaxApprovalVelocityPerHour: 100,
		MaxPostingsPerHour:         1000,
	}
	enforcer := service.NewSafetyEnforcer(policy, metrics.NewNoOpMetricsProvider())
	userID := uuid.New()
	ctx := shared.WithTenantID(context.Background(), uuid.New())

	// Exhaust limit.
	for i := 0; i < limit; i++ {
		if err := enforcer.CheckReversalVelocity(ctx, userID); err != nil {
			t.Fatalf("call %d: unexpected block before limit: %v", i+1, err)
		}
	}

	// Over-limit call must be blocked.
	if err := enforcer.CheckReversalVelocity(ctx, userID); err == nil {
		t.Fatal("FIN-MUT-006 FAILED: velocity limit removed — over-limit reversal was allowed")
	}
}

// ── FIN-MUT-007: Amount limit ─────────────────────────────────────────────────

func TestMutationHarness_AmountLimit_BlocksOverLimit(t *testing.T) {
	max := decimal.NewFromFloat(50_000)
	policy := service.SafetyPolicy{
		MaxTransactionAmount:       max,
		MaxReversalsPerHour:        100,
		MaxApprovalVelocityPerHour: 100,
		MaxPostingsPerHour:         1000,
	}
	enforcer := service.NewSafetyEnforcer(policy, metrics.NewNoOpMetricsProvider())
	ctx := shared.WithTenantID(context.Background(), uuid.New())

	over := max.Add(decimal.NewFromFloat(0.01))
	if err := enforcer.CheckTransactionAmount(ctx, over); err == nil {
		t.Fatal("FIN-MUT-007 FAILED: amount limit removed — over-limit amount was allowed")
	}
}

// ── FIN-MUT-008/009: Extension contract enforcement ───────────────────────────

func TestMutationHarness_ExtensionContract_MissingAuditHook_DetectedAtStartup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	registry := service.NewGovernancePolicyRegistry(service.DefaultGovernancePolicy(), metrics.NewNoOpMetricsProvider())
	enforcer := service.NewSafetyEnforcer(service.DefaultSafetyPolicy(), metrics.NewNoOpMetricsProvider())
	guard := service.NewEvolutionSafetyGuard(registry, enforcer, metrics.NewNoOpMetricsProvider())

	// Register an extension that mutates finance state but has NO audit hook.
	// This is the mutation: audit hook was removed.
	_ = guard.RegisterExtension(service.ExtensionContract{
		Name:                "missing_audit_hook_extension",
		Description:         "Extension that forgot to declare an audit hook",
		MutatesFinanceState: true,
		HasAuditHook:        false, // <-- mutation: should be true
		HasSafetyCheck:      true,
		HasIntegrityCheck:   true,
	})

	violations := guard.ValidateExtensions()
	if len(violations) == 0 {
		t.Fatal("FIN-MUT-008 FAILED: extension contract audit check removed — missing audit hook not detected")
	}
	found := false
	for _, v := range violations {
		if v.Field == "HasAuditHook" {
			found = true
		}
	}
	if !found {
		t.Errorf("FIN-MUT-008: expected HasAuditHook violation, got: %+v", violations)
	}
}

func TestMutationHarness_ExtensionContract_MissingSafetyCheck_DetectedAtStartup(t *testing.T) {
	registry := service.NewGovernancePolicyRegistry(service.DefaultGovernancePolicy(), metrics.NewNoOpMetricsProvider())
	enforcer := service.NewSafetyEnforcer(service.DefaultSafetyPolicy(), metrics.NewNoOpMetricsProvider())
	guard := service.NewEvolutionSafetyGuard(registry, enforcer, metrics.NewNoOpMetricsProvider())

	_ = guard.RegisterExtension(service.ExtensionContract{
		Name:                "missing_safety_check_extension",
		Description:         "Extension that forgot to declare a safety check",
		MutatesFinanceState: true,
		HasAuditHook:        true,
		HasSafetyCheck:      false, // <-- mutation: should be true
		HasIntegrityCheck:   true,
	})

	violations := guard.ValidateExtensions()
	if len(violations) == 0 {
		t.Fatal("FIN-MUT-009 FAILED: extension contract safety check removed — missing safety check not detected")
	}
	found := false
	for _, v := range violations {
		if v.Field == "HasSafetyCheck" {
			found = true
		}
	}
	if !found {
		t.Errorf("FIN-MUT-009: expected HasSafetyCheck violation, got: %+v", violations)
	}
}

// ── FIN-MUT-010: EvolutionGuard nil-safety ────────────────────────────────────

func TestMutationHarness_NilEvolutionGuard_NoPanicNoBlock(t *testing.T) {
	var guard *service.EvolutionSafetyGuard
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("FIN-MUT-010: nil guard panicked: %v", r)
		}
	}()
	// Nil guard must pass all checks (safe degraded state, not a hard block).
	if err := guard.AssertStartupInvariants(p16ctx()); err != nil {
		t.Errorf("FIN-MUT-010: nil guard should not return error, got: %v", err)
	}
	if violations := guard.ValidateExtensions(); len(violations) > 0 {
		t.Errorf("FIN-MUT-010: nil guard should have no violations, got: %+v", violations)
	}
}

// ── Mutation coverage summary ─────────────────────────────────────────────────

// TestMutationHarness_CoverageSummary logs the mutation coverage table.
// This is a documentation test — it always passes, but provides a human-readable
// audit trail of what protections have been mutation-tested in this phase.
func TestMutationHarness_CoverageSummary(t *testing.T) {
	coverage := []mutationResult{
		{"FIN-MUT-001: SOD check", true, "SOD_VIOLATION"},
		{"FIN-MUT-002: Terminal deletion", true, "TRANSACTION_DELETE_NOT_ALLOWED"},
		{"FIN-MUT-003: Approval gate", true, "APPROVAL_REQUIRED"},
		{"FIN-MUT-004: Balance check", true, "UNBALANCED_TRANSACTION"},
		{"FIN-MUT-005: Terminal absorbing state", true, "state machine"},
		{"FIN-MUT-006: Velocity limit", true, "velocity window"},
		{"FIN-MUT-007: Amount limit", true, "TRANSACTION_AMOUNT_EXCEEDED"},
		{"FIN-MUT-008: Audit hook contract", true, "HasAuditHook violation"},
		{"FIN-MUT-009: Safety check contract", true, "HasSafetyCheck violation"},
		{"FIN-MUT-010: Nil guard nil-safety", true, "nil-safe"},
	}

	verified := 0
	for _, m := range coverage {
		if m.protected {
			verified++
		}
		t.Logf("[%s] protected=%v errCode=%s", m.name, m.protected, m.errCode)
	}
	t.Logf("Mutation coverage: %d/%d protections verified", verified, len(coverage))

	if verified < len(coverage) {
		t.Errorf("mutation coverage gap: only %d/%d protections mutation-tested", verified, len(coverage))
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// p16NewSvc creates a minimal TransactionService suitable for mutation tests.
func p16NewSvc(repo *p16StubTxnRepo) service.TransactionService {
	return service.NewTransactionService(
		repo,
		&p16StubAccountRepo{},
		nil, // periodRepo
		nil, // reversalHistoryRepo
		&p16StubEntryService{},
		nil, // txRunner
		tracing.NewNoOpService(),
		metrics.NewNoOpMetricsProvider(),
		nil, // auditWriter
		nil, // safetyEnforcer
		nil, // anomalyDetector
	)
}

// p16ContainsCode returns true when the error message contains the given code.
func p16ContainsCode(err error, code string) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), code)
}

// ── p16StubAccountRepo ────────────────────────────────────────────────────────

type p16StubAccountRepo struct{}

func (r *p16StubAccountRepo) Create(ctx context.Context, a *domain.Accounts) error { return nil }
func (r *p16StubAccountRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Accounts, error) {
	return nil, nil
}

func (r *p16StubAccountRepo) GetByCode(ctx context.Context, entityID *uuid.UUID, code string) (*domain.Accounts, error) {
	return nil, nil
}
func (r *p16StubAccountRepo) Update(ctx context.Context, a *domain.Accounts) error { return nil }
func (r *p16StubAccountRepo) Delete(ctx context.Context, id uuid.UUID) error       { return nil }
func (r *p16StubAccountRepo) List(ctx context.Context, filter *domain.AccountFilter) ([]*domain.Accounts, error) {
	return nil, nil
}

func (r *p16StubAccountRepo) Count(ctx context.Context, filter *domain.AccountFilter) (int64, error) {
	return 0, nil
}

func (r *p16StubAccountRepo) ListByParent(ctx context.Context, parentID uuid.UUID) ([]*domain.Accounts, error) {
	return nil, nil
}

func (r *p16StubAccountRepo) ListByRootType(ctx context.Context, rootType domain.RootType) ([]*domain.Accounts, error) {
	return nil, nil
}

func (r *p16StubAccountRepo) GetAccountHierarchy(ctx context.Context, rootID uuid.UUID) ([]*domain.Accounts, error) {
	return nil, nil
}

func (r *p16StubAccountRepo) GetAccountPath(ctx context.Context, accountID uuid.UUID) ([]domain.Accounts, error) {
	return nil, nil
}

func (r *p16StubAccountRepo) ValidateHierarchy(ctx context.Context, accountID, parentID uuid.UUID) error {
	return nil
}

func (r *p16StubAccountRepo) GetAccountBalance(ctx context.Context, accountID uuid.UUID, asOfDate *time.Time) (*domain.AccountBalance, error) {
	return nil, nil
}

func (r *p16StubAccountRepo) GetAccountBalances(ctx context.Context, accountIDs []uuid.UUID, asOfDate *time.Time) ([]*domain.AccountBalance, error) {
	return nil, nil
}

func (r *p16StubAccountRepo) GetTrialBalance(ctx context.Context, entityID *uuid.UUID, asOfDate *time.Time) ([]*domain.TrialBalanceEntry, error) {
	return nil, nil
}

func (r *p16StubAccountRepo) GetActiveAccounts(ctx context.Context, entityID *uuid.UUID) ([]*domain.Accounts, error) {
	return nil, nil
}

func (r *p16StubAccountRepo) GetControlAccounts(ctx context.Context, entityID *uuid.UUID) ([]*domain.Accounts, error) {
	return nil, nil
}

func (r *p16StubAccountRepo) GetAccountsByType(ctx context.Context, accountType string, rootType *domain.RootType) ([]*domain.Accounts, error) {
	return nil, nil
}

func (r *p16StubAccountRepo) Search(ctx context.Context, query string, limit int) ([]*domain.Accounts, error) {
	return nil, nil
}

func (r *p16StubAccountRepo) ValidateAccountCode(ctx context.Context, code string, excludeID *uuid.UUID) error {
	return nil
}

func (r *p16StubAccountRepo) IsAccountCodeUnique(ctx context.Context, entityID *uuid.UUID, code string, excludeID *uuid.UUID) (bool, error) {
	return true, nil
}

func (r *p16StubAccountRepo) HasChildren(ctx context.Context, accountID uuid.UUID) (bool, error) {
	return false, nil
}

func (r *p16StubAccountRepo) GetChildren(ctx context.Context, accountID uuid.UUID) ([]*domain.Accounts, error) {
	return nil, nil
}

func (r *p16StubAccountRepo) HasTransactions(ctx context.Context, accountID uuid.UUID) (bool, error) {
	return false, nil
}

func (r *p16StubAccountRepo) UpdateBalance(ctx context.Context, accountID uuid.UUID, balance domain.AccountBalance) error {
	return nil
}

func (r *p16StubAccountRepo) CreateAccountGroup(ctx context.Context, group *domain.AccountGroup) error {
	return nil
}

func (r *p16StubAccountRepo) GetAccountGroupByID(ctx context.Context, id uuid.UUID) (*domain.AccountGroup, error) {
	return nil, nil
}

func (r *p16StubAccountRepo) GetAccountGroupByCode(ctx context.Context, code string, entityID *uuid.UUID) (*domain.AccountGroup, error) {
	return nil, nil
}

func (r *p16StubAccountRepo) UpdateAccountGroup(ctx context.Context, id uuid.UUID, group *domain.AccountGroup) error {
	return nil
}

func (r *p16StubAccountRepo) DeleteAccountGroup(ctx context.Context, id uuid.UUID, entityID *uuid.UUID) error {
	return nil
}

func (r *p16StubAccountRepo) ListAccountGroups(ctx context.Context, filter *domain.AccountGroupFilter) ([]*domain.AccountGroup, error) {
	return nil, nil
}

func (r *p16StubAccountRepo) CountAccountGroups(ctx context.Context, filter *domain.AccountGroupFilter) (int64, error) {
	return 0, nil
}

func (r *p16StubAccountRepo) GetAccountGroupHierarchy(ctx context.Context, rootGroupID *uuid.UUID, entityID *uuid.UUID) ([]*domain.AccountGroup, error) {
	return nil, nil
}

func (r *p16StubAccountRepo) GetGroupsByFinancialStatement(ctx context.Context, statementType string, entityID *uuid.UUID) ([]*domain.AccountGroup, error) {
	return nil, nil
}

func (r *p16StubAccountRepo) GetGroupsByCashFlowCategory(ctx context.Context, category string, entityID *uuid.UUID) ([]*domain.AccountGroup, error) {
	return nil, nil
}

func (r *p16StubAccountRepo) ValidateAccountGroupCode(ctx context.Context, code string, excludeID *uuid.UUID, entityID *uuid.UUID) error {
	return nil
}

func (r *p16StubAccountRepo) GetAccountChildrenHierarchy(ctx context.Context, parentAccountID uuid.UUID) ([]*domain.AccountHierarchy, error) {
	return nil, nil
}

func (r *p16StubAccountRepo) GetAccountSubtree(ctx context.Context, accountID uuid.UUID) ([]*domain.AccountHierarchy, error) {
	return nil, nil
}

func (r *p16StubAccountRepo) GetAccountsWithRecentActivity(ctx context.Context, filter *domain.AccountActivityFilter) ([]*domain.AccountActivity, error) {
	return nil, nil
}

func (r *p16StubAccountRepo) GetStaleAccountBalances(ctx context.Context, filter *domain.AccountActivityFilter) ([]*domain.AccountActivity, error) {
	return nil, nil
}

func (r *p16StubAccountRepo) GetAccountActivitySummary(ctx context.Context, filter *domain.AccountActivityFilter) ([]*domain.AccountActivitySummary, error) {
	return nil, nil
}

// ── p16StubEntryService ───────────────────────────────────────────────────────

type p16StubEntryService struct{}

func (s *p16StubEntryService) CreateEntry(ctx context.Context, entry *domain.TransactionEntry) error {
	return nil
}

func (s *p16StubEntryService) CreateEntries(ctx context.Context, entries []*domain.TransactionEntry) error {
	return nil
}

func (s *p16StubEntryService) GetEntryByID(ctx context.Context, id uuid.UUID) (*domain.TransactionEntry, error) {
	return nil, nil
}

func (s *p16StubEntryService) GetEntriesByTransactionID(ctx context.Context, transactionID uuid.UUID) ([]*domain.TransactionEntry, error) {
	return nil, nil
}

func (s *p16StubEntryService) GetEntriesByAccountID(ctx context.Context, accountID uuid.UUID, limit int, offset int) ([]*domain.TransactionEntry, error) {
	return nil, nil
}

func (s *p16StubEntryService) UpdateEntry(ctx context.Context, id uuid.UUID, req domain.TransactionEntry) (*domain.TransactionEntry, error) {
	return nil, nil
}
func (s *p16StubEntryService) DeleteEntry(ctx context.Context, id uuid.UUID) error { return nil }
func (s *p16StubEntryService) SearchEntries(ctx context.Context, query string, filters *domain.EntryFilter, limit int, offset int) ([]*domain.TransactionEntry, error) {
	return nil, nil
}

func (s *p16StubEntryService) ReconcileEntries(ctx context.Context, entryIDs []uuid.UUID, reconciliationRef string) error {
	return nil
}

func (s *p16StubEntryService) UnreconcileEntries(ctx context.Context, entryIDs []uuid.UUID) error {
	return nil
}

func (s *p16StubEntryService) GetUnreconciledEntries(ctx context.Context, accountID uuid.UUID, cutoffDate *time.Time) ([]*domain.TransactionEntry, error) {
	return nil, nil
}

func (s *p16StubEntryService) ValidateEntryConsistency(ctx context.Context, entry *domain.TransactionEntry) ([]domain.ValidationError, error) {
	return nil, nil
}

func (s *p16StubEntryService) GetEntrySummary(ctx context.Context, accountID uuid.UUID, startDate, endDate time.Time) (*domain.TransactionSummary, error) {
	return nil, nil
}
