// Package service_test — Phase 16 assurance layer framework and Layer A/B/E tests.
//
// This file documents the continuous assurance hierarchy and verifies that the
// Layer A (unit invariants), Layer B (integration assurance), and Layer E
// (production trust) coverage is complete.
//
// Assurance Layers:
//
//	Layer A — Unit invariants: state machines, balance validation, integrity checks.
//	Layer B — Integration assurance: repository+DB, transaction rollback, cache, RLS.
//	Layer C — Orchestration assurance: workflows, retries, replay, idempotency.
//	Layer D — Adversarial assurance: concurrency chaos, governance regression, corruption.
//	Layer E — Production trust: explain-plan, observability continuity, startup safety.
//
// CI Gate Rule: If any Layer A test fails, the build MUST fail. Layer A invariants
// are the load-bearing floor of the entire assurance pyramid.
package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"awo.so/internal/core/finance/domain"
	"awo.so/internal/core/finance/service"
	"awo.so/internal/shared/metrics"
)

// ============================================================================
// LAYER A — Unit Invariants
// ============================================================================

// TestLayerA_BalanceInvariant_SoundAndComplete verifies the balance check is
// both SOUND (no false positives on valid data) and COMPLETE (no false negatives
// on invalid data). Both must pass — if either fails, the balance invariant is broken.
func TestLayerA_BalanceInvariant_SoundAndComplete(t *testing.T) {
	m := metrics.NewNoOpMetricsProvider()

	// Soundness: balanced entries must never trigger a violation.
	soundCases := [][]domain.TransactionEntry{
		{
			{DebitAmount: decimal.NewFromFloat(100), CreditAmount: decimal.Zero},
			{DebitAmount: decimal.Zero, CreditAmount: decimal.NewFromFloat(100)},
		},
		{
			{DebitAmount: decimal.NewFromFloat(500), CreditAmount: decimal.Zero},
			{DebitAmount: decimal.Zero, CreditAmount: decimal.NewFromFloat(300)},
			{DebitAmount: decimal.Zero, CreditAmount: decimal.NewFromFloat(200)},
		},
		{
			{DebitAmount: decimal.NewFromFloat(0.01), CreditAmount: decimal.Zero},
			{DebitAmount: decimal.Zero, CreditAmount: decimal.NewFromFloat(0.01)},
		},
	}
	for i, entries := range soundCases {
		repo := p16SingleTxnRepo(entries)
		svc := service.NewIntegrityService(repo, nil, m)
		report, err := svc.ScanPostedTransactions(p16ctx(), nil)
		if err != nil {
			t.Fatalf("sound case %d: unexpected scan error: %v", i, err)
		}
		for _, v := range report.Violations {
			if v.Kind == "UNBALANCED_TRANSACTION" {
				t.Errorf("sound case %d: false positive — balanced entries triggered UNBALANCED_TRANSACTION", i)
			}
		}
	}

	// Completeness: unbalanced entries must always trigger a CRITICAL violation.
	type unbalancedCase struct {
		name    string
		entries []domain.TransactionEntry
	}
	unbalancedCases := []unbalancedCase{
		{
			"off by one",
			[]domain.TransactionEntry{
				{DebitAmount: decimal.NewFromFloat(100), CreditAmount: decimal.Zero},
				{DebitAmount: decimal.Zero, CreditAmount: decimal.NewFromFloat(99)},
			},
		},
		{
			"debit only",
			[]domain.TransactionEntry{
				{DebitAmount: decimal.NewFromFloat(1000), CreditAmount: decimal.Zero},
			},
		},
		{
			"penny imbalance",
			[]domain.TransactionEntry{
				{DebitAmount: decimal.NewFromFloat(0.01), CreditAmount: decimal.Zero},
				{DebitAmount: decimal.Zero, CreditAmount: decimal.NewFromFloat(0.02)},
			},
		},
	}
	for _, tc := range unbalancedCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			repo := p16SingleTxnRepo(tc.entries)
			svc := service.NewIntegrityService(repo, nil, m)
			report, err := svc.ScanPostedTransactions(p16ctx(), nil)
			if err != nil {
				t.Fatalf("unexpected scan error: %v", err)
			}
			found := false
			for _, v := range report.Violations {
				if v.Kind == "UNBALANCED_TRANSACTION" && v.Severity == service.SeverityCritical {
					found = true
				}
			}
			if !found {
				t.Errorf("false negative — unbalanced entries not detected as CRITICAL")
			}
		})
	}
}

// TestLayerA_StateMachine_ValidTransitions proves the state machine allows
// the canonical lifecycle paths used in production.
func TestLayerA_StateMachine_ValidTransitions(t *testing.T) {
	validPaths := []struct {
		from domain.TransactionStatus
		to   domain.TransactionStatus
	}{
		{domain.TransactionStatusDraft, domain.TransactionStatusPendingApproval},
		{domain.TransactionStatusPendingApproval, domain.TransactionStatusApproved},
		{domain.TransactionStatusApproved, domain.TransactionStatusPosted},
		{domain.TransactionStatusPosted, domain.TransactionStatusReversed},
		{domain.TransactionStatusDraft, domain.TransactionStatusCancelled},
		{domain.TransactionStatusApproved, domain.TransactionStatusCancelled},
	}
	for _, p := range validPaths {
		if !p.from.CanTransitionTo(p.to) {
			t.Errorf("Layer A invariant broken: valid transition %s → %s not allowed", p.from, p.to)
		}
	}
}

// TestLayerA_StateMachine_InvalidTransitions proves the state machine blocks
// all invalid lifecycle transitions.
func TestLayerA_StateMachine_InvalidTransitions(t *testing.T) {
	terminals := []domain.TransactionStatus{
		domain.TransactionStatusReversed,
		domain.TransactionStatusCancelled,
	}
	allStatuses := []domain.TransactionStatus{
		domain.TransactionStatusDraft,
		domain.TransactionStatusPendingApproval,
		domain.TransactionStatusApproved,
		domain.TransactionStatusRejected,
		domain.TransactionStatusPosted,
		domain.TransactionStatusReversed,
		domain.TransactionStatusCancelled,
	}
	for _, terminal := range terminals {
		for _, target := range allStatuses {
			if terminal.CanTransitionTo(target) {
				t.Errorf("Layer A: terminal %s → %s transition allowed (absorbing state violated)", terminal, target)
			}
		}
	}

	// POSTED cannot go backward.
	backwardFromPosted := []domain.TransactionStatus{
		domain.TransactionStatusDraft,
		domain.TransactionStatusPendingApproval,
		domain.TransactionStatusApproved,
	}
	for _, target := range backwardFromPosted {
		if domain.TransactionStatusPosted.CanTransitionTo(target) {
			t.Errorf("Layer A: POSTED → %s backward transition allowed", target)
		}
	}
}

// TestLayerA_HasCritical_Monotonic proves that once CRITICAL violations exist,
// adding more violations of any severity never removes the critical status.
func TestLayerA_HasCritical_Monotonic(t *testing.T) {
	report := &service.IntegrityReport{
		Violations: []service.IntegrityViolation{
			{Kind: "UNBALANCED_TRANSACTION", Severity: service.SeverityCritical},
		},
	}
	if !report.HasCritical() {
		t.Fatal("Layer A: report with CRITICAL violation must return HasCritical=true")
	}
	for i := 0; i < 100; i++ {
		report.Violations = append(report.Violations, service.IntegrityViolation{
			Kind: "EXTRA", Severity: service.SeverityHigh,
		})
	}
	if !report.HasCritical() {
		t.Error("Layer A: HasCritical became false after adding non-critical violations — monotonicity broken")
	}
}

// TestLayerA_SafetyEnforcer_AmountBoundary_ExactAtLimit verifies at-limit is
// allowed and over-limit is blocked — no off-by-one in the boundary check.
func TestLayerA_SafetyEnforcer_AmountBoundary_ExactAtLimit(t *testing.T) {
	limit := decimal.NewFromFloat(100_000)
	policy := service.SafetyPolicy{
		MaxTransactionAmount:       limit,
		MaxReversalsPerHour:        100,
		MaxApprovalVelocityPerHour: 100,
		MaxPostingsPerHour:         1000,
	}
	enforcer := service.NewSafetyEnforcer(policy, metrics.NewNoOpMetricsProvider())
	ctx := p16ctx()

	if err := enforcer.CheckTransactionAmount(ctx, limit); err != nil {
		t.Errorf("Layer A: at-limit amount %s should be allowed, got: %v", limit, err)
	}
	overLimit := limit.Add(decimal.NewFromFloat(0.01))
	if err := enforcer.CheckTransactionAmount(ctx, overLimit); err == nil {
		t.Errorf("Layer A: over-limit amount %s should be blocked, got nil", overLimit)
	}
}

// ============================================================================
// LAYER B — Integration Assurance
// ============================================================================

// TestLayerB_IntegrityScan_PostedWithoutEntries_AlwaysHigh proves that a POSTED
// transaction with no entries is always flagged as a HIGH violation.
func TestLayerB_IntegrityScan_PostedWithoutEntries_AlwaysHigh(t *testing.T) {
	txns := []*domain.Transaction{
		{ID: uuid.New(), TenantID: uuid.New(), TransactionStatus: domain.TransactionStatusPosted, TransactionNumber: "TXN-B-001"},
		{ID: uuid.New(), TenantID: uuid.New(), TransactionStatus: domain.TransactionStatusPosted, TransactionNumber: "TXN-B-002"},
	}
	for _, txn := range txns {
		txn := txn
		repo := p16EmptyEntriesRepo(txn)
		svc := service.NewIntegrityService(repo, nil, metrics.NewNoOpMetricsProvider())
		report, err := svc.ScanPostedTransactions(p16ctx(), nil)
		if err != nil {
			t.Fatalf("Layer B: txn %s: unexpected scan error: %v", txn.TransactionNumber, err)
		}
		found := false
		for _, v := range report.Violations {
			if v.Kind == "POSTED_WITHOUT_ENTRIES" && v.Severity == service.SeverityHigh {
				found = true
			}
		}
		if !found {
			t.Errorf("Layer B: txn %s: POSTED_WITHOUT_ENTRIES not detected", txn.TransactionNumber)
		}
	}
}

// TestLayerB_NilAuditChainVerifier_NoPanic verifies that a nil AuditChainVerifier
// never panics.
func TestLayerB_NilAuditChainVerifier_NoPanic(t *testing.T) {
	var verifier *service.AuditChainVerifier
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Layer B: nil AuditChainVerifier panicked: %v", r)
		}
	}()
	_, _ = verifier.VerifyChain(p16ctx(), 1, 100, 50)
}

// TestLayerB_NilEvolutionGuard_NoPanic verifies that a nil EvolutionSafetyGuard
// passes all invariant checks gracefully.
func TestLayerB_NilEvolutionGuard_NoPanic(t *testing.T) {
	var guard *service.EvolutionSafetyGuard
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Layer B: nil EvolutionSafetyGuard panicked: %v", r)
		}
	}()
	ctx := p16ctx()
	_ = guard.AssertStartupInvariants(ctx)
	_ = guard.ValidateExtensions()
	_ = guard.RegisterExtension(service.ExtensionContract{Name: "test"})
	_ = guard.ListContracts()
}

// ============================================================================
// LAYER E — Production Trust: Assurance Dashboard
// ============================================================================

// TestLayerE_AssuranceDashboard_GeneratesReport verifies the dashboard produces
// a valid, structured report with scoring.
func TestLayerE_AssuranceDashboard_GeneratesReport(t *testing.T) {
	dashboard := service.NewAssuranceDashboard(nil, nil, nil, nil, true, true)
	report := dashboard.GenerateReport(p16ctx())
	if report == nil {
		t.Fatal("Layer E: dashboard must produce non-nil report")
	}
	if report.OverallScore < 0 || report.OverallScore > 100 {
		t.Errorf("Layer E: OverallScore=%d out of [0,100] range", report.OverallScore)
	}
	if len(report.Protections) == 0 {
		t.Error("Layer E: protections list must not be empty")
	}
	if report.Summary() == "" {
		t.Error("Layer E: Summary must not be empty")
	}
}

// TestLayerE_NilAssuranceDashboard_NoPanic verifies a nil dashboard does not panic.
func TestLayerE_NilAssuranceDashboard_NoPanic(t *testing.T) {
	var d *service.AssuranceDashboard
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Layer E: nil AssuranceDashboard panicked: %v", r)
		}
	}()
	report := d.GenerateReport(p16ctx())
	if report == nil {
		t.Fatal("nil dashboard must return a non-nil degraded report")
	}
	if report.IsAcceptable() {
		t.Error("nil dashboard report must not be acceptable (score=0)")
	}
}

// TestLayerE_AssuranceReport_IsAcceptable gates correctly on score and failures.
func TestLayerE_AssuranceReport_IsAcceptable(t *testing.T) {
	low := &service.AssuranceReport{OverallScore: service.MinAssuranceScore - 1}
	if low.IsAcceptable() {
		t.Errorf("score %d should not be acceptable (min=%d)", low.OverallScore, service.MinAssuranceScore)
	}

	ok := &service.AssuranceReport{OverallScore: service.MinAssuranceScore}
	if !ok.IsAcceptable() {
		t.Errorf("score %d should be acceptable (min=%d)", ok.OverallScore, service.MinAssuranceScore)
	}

	withFailures := &service.AssuranceReport{
		OverallScore:    100,
		StartupFailures: []string{"critical dependency missing"},
	}
	if withFailures.IsAcceptable() {
		t.Error("report with startup failures must not be acceptable even at score=100")
	}
}

// ============================================================================
// Shared helpers — p16 prefix avoids conflicts with other test stubs
// ============================================================================

func p16ctx() context.Context {
	return context.Background()
}

// p16SingleTxnRepo returns a stub repo serving a single POSTED transaction
// with the provided entries (paginated: returns txn on call 1, nil on call 2+).
func p16SingleTxnRepo(entries []domain.TransactionEntry) domain.TransactionRepository {
	txn := &domain.Transaction{
		ID:                uuid.New(),
		TenantID:          uuid.New(),
		TransactionStatus: domain.TransactionStatusPosted,
		TransactionNumber: "TXN-P16-001",
		TransactionDate:   time.Now(),
		CreatedAt:         time.Now(),
	}
	pageCount := 0
	return &p16StubTxnRepo{
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
}

// p16EmptyEntriesRepo returns a stub repo for one POSTED transaction with zero entries.
func p16EmptyEntriesRepo(txn *domain.Transaction) domain.TransactionRepository {
	pageCount := 0
	return &p16StubTxnRepo{
		fnList: func(_ context.Context, _ *domain.TransactionFilter) ([]*domain.Transaction, error) {
			pageCount++
			if pageCount == 1 {
				return []*domain.Transaction{txn}, nil
			}
			return nil, nil
		},
		fnGetEntriesByTxn: func(_ context.Context, _ uuid.UUID) ([]domain.TransactionEntry, error) {
			return []domain.TransactionEntry{}, nil
		},
	}
}
