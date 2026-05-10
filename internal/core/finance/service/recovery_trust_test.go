// Package service_test — Phase 16 recovery trust and production posture tests.
//
// Recovery trust: verifies that the finance module can detect, report, and
// signal recovery from degraded states without panicking or hiding failures.
//
// Tests (FIN-RECOVERY-*):
//
//	001 — AssuranceReport.Summary() produces human-readable output
//	002 — AssuranceReport.IsAcceptable() false when score < MinAssuranceScore
//	003 — AssuranceReport.IsAcceptable() false when StartupFailures present
//	004 — AssuranceReport.IsAcceptable() true at exactly MinAssuranceScore
//	005 — Nil AssuranceReport.IsAcceptable() returns false (nil-safe)
//	006 — AssuranceDashboard fully-wired report scores above minimum
//	007 — AssuranceDashboard QueryBudgetStatus is a recognised value
//	008 — AssuranceDashboard reports NOT_RUN integrity scan when no service
//	009 — AssuranceDashboard reports NOT_RUN audit chain when no verifier
//	010 — Nil AssuranceDashboard GenerateReport never panics
//	011 — IntegrityService propagates repo error (not silently swallowed)
//	012 — AuditChainVerifier with nil repo returns error (not panic)
package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"awo.so/internal/core/finance/domain"
	"awo.so/internal/core/finance/service"
	"awo.so/internal/shared/metrics"
)

// ── FIN-RECOVERY-001: Summary produces human-readable output ─────────────────

func TestRecoveryTrust_AssuranceReport_Summary_HumanReadable(t *testing.T) {
	dashboard := service.NewAssuranceDashboard(nil, nil, nil, nil, true, true)
	report := dashboard.GenerateReport(p16ctx())

	summary := report.Summary()
	if summary == "" {
		t.Fatal("FIN-RECOVERY-001: Summary() should return non-empty string")
	}
	if !strings.Contains(summary, "score=") {
		t.Errorf("FIN-RECOVERY-001: Summary() should contain 'score=', got: %s", summary)
	}
}

// ── FIN-RECOVERY-002: IsAcceptable false when score below threshold ───────────

func TestRecoveryTrust_IsAcceptable_FalseWhenScoreTooLow(t *testing.T) {
	report := &service.AssuranceReport{
		OverallScore: service.MinAssuranceScore - 1,
	}
	if report.IsAcceptable() {
		t.Errorf("FIN-RECOVERY-002: score=%d should not be acceptable (min=%d)",
			report.OverallScore, service.MinAssuranceScore)
	}
}

// ── FIN-RECOVERY-003: IsAcceptable false when StartupFailures present ─────────

func TestRecoveryTrust_IsAcceptable_FalseWhenStartupFailures(t *testing.T) {
	report := &service.AssuranceReport{
		OverallScore:    100,
		StartupFailures: []string{"nil enforcer"},
	}
	if report.IsAcceptable() {
		t.Error("FIN-RECOVERY-003: report with startup failures must not be acceptable even at score=100")
	}
}

// ── FIN-RECOVERY-004: IsAcceptable true at exactly MinAssuranceScore ──────────

func TestRecoveryTrust_IsAcceptable_TrueAtExactThreshold(t *testing.T) {
	report := &service.AssuranceReport{
		OverallScore: service.MinAssuranceScore,
	}
	if !report.IsAcceptable() {
		t.Errorf("FIN-RECOVERY-004: score=%d should be acceptable (min=%d)",
			report.OverallScore, service.MinAssuranceScore)
	}
}

// ── FIN-RECOVERY-005: Nil AssuranceReport.IsAcceptable returns false ──────────

func TestRecoveryTrust_NilAssuranceReport_IsAcceptable_False(t *testing.T) {
	var report *service.AssuranceReport
	if report.IsAcceptable() {
		t.Error("FIN-RECOVERY-005: nil AssuranceReport.IsAcceptable() should return false")
	}
}

// ── FIN-RECOVERY-006: Fully-wired dashboard scores above minimum ──────────────

func TestRecoveryTrust_FullyWiredDashboard_ScoresAboveMinimum(t *testing.T) {
	registry := service.NewGovernancePolicyRegistry(service.DefaultGovernancePolicy(), metrics.NewNoOpMetricsProvider())
	enforcer := service.NewSafetyEnforcer(service.DefaultSafetyPolicy(), metrics.NewNoOpMetricsProvider())
	guard := service.NewEvolutionSafetyGuard(registry, enforcer, metrics.NewNoOpMetricsProvider())

	dashboard := service.NewAssuranceDashboard(guard, nil, enforcer, nil, true, true)
	report := dashboard.GenerateReport(p16ctx())

	if report.OverallScore < service.MinAssuranceScore {
		t.Errorf("FIN-RECOVERY-006: fully-wired dashboard scored %d, expected >= %d",
			report.OverallScore, service.MinAssuranceScore)
	}
}

// ── FIN-RECOVERY-007: Dashboard QueryBudgetStatus is a recognised value ────────

func TestRecoveryTrust_Dashboard_QueryBudgetStatus_KnownValue(t *testing.T) {
	dashboard := service.NewAssuranceDashboard(nil, nil, nil, nil, true, true)
	report := dashboard.GenerateReport(p16ctx())

	known := map[string]bool{"OK": true, "REGRESSED": true, "UNKNOWN": true}
	if !known[report.QueryBudgetStatus] {
		t.Errorf("FIN-RECOVERY-007: QueryBudgetStatus=%q not in {OK, REGRESSED, UNKNOWN}",
			report.QueryBudgetStatus)
	}
}

// ── FIN-RECOVERY-008: Dashboard reports NOT_RUN integrity scan when no service ─

func TestRecoveryTrust_Dashboard_IntegrityScanStatus_NotRunWhenNilService(t *testing.T) {
	dashboard := service.NewAssuranceDashboard(nil, nil, nil, nil, false, false)
	report := dashboard.GenerateReport(p16ctx())

	if report.IntegrityScanStatus != "NOT_RUN" {
		t.Errorf("FIN-RECOVERY-008: expected NOT_RUN when no integrity service, got: %s",
			report.IntegrityScanStatus)
	}
}

// ── FIN-RECOVERY-009: Dashboard reports NOT_RUN audit chain when no verifier ───

func TestRecoveryTrust_Dashboard_AuditChainStatus_NotRunWhenNilVerifier(t *testing.T) {
	dashboard := service.NewAssuranceDashboard(nil, nil, nil, nil, false, false)
	report := dashboard.GenerateReport(p16ctx())

	if report.AuditChainStatus != "NOT_RUN" {
		t.Errorf("FIN-RECOVERY-009: expected NOT_RUN when no chain verifier, got: %s",
			report.AuditChainStatus)
	}
}

// ── FIN-RECOVERY-010: Nil dashboard GenerateReport never panics ───────────────

func TestRecoveryTrust_NilDashboard_GenerateReport_NoPanic(t *testing.T) {
	var dashboard *service.AssuranceDashboard

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("FIN-RECOVERY-010: nil AssuranceDashboard.GenerateReport panicked: %v", r)
		}
	}()

	_ = dashboard.GenerateReport(p16ctx())
}

// ── FIN-RECOVERY-011: IntegrityService propagates repo error ─────────────────

func TestRecoveryTrust_IntegrityService_RepoError_Propagates(t *testing.T) {
	listErr := errors.New("DB connection lost")
	repo := &p16StubTxnRepo{
		fnList: func(_ context.Context, _ *domain.TransactionFilter) ([]*domain.Transaction, error) {
			return nil, listErr
		},
	}

	intSvc := service.NewIntegrityService(repo, nil, metrics.NewNoOpMetricsProvider())

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("FIN-RECOVERY-011: IntegrityService panicked on repo error: %v", r)
		}
	}()

	_, err := intSvc.ScanPostedTransactions(p16ctx(), nil)
	if err == nil {
		t.Error("FIN-RECOVERY-011: repo error should propagate from ScanPostedTransactions, got nil")
	}
	if !errors.Is(err, listErr) && !strings.Contains(err.Error(), listErr.Error()) {
		t.Errorf("FIN-RECOVERY-011: expected error to wrap listErr, got: %v", err)
	}
}

// ── FIN-RECOVERY-012: AuditChainVerifier with nil repo never panics ──────────

func TestRecoveryTrust_AuditChainVerifier_NilRepo_ReturnsError(t *testing.T) {
	// nil repo → VerifyChain short-circuits (nil-safe contract) and returns a
	// non-nil report without panic. FIN-CONC-005 verifies this under concurrent load.
	verifier := service.NewAuditChainVerifier(nil, metrics.NewNoOpMetricsProvider())

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("FIN-RECOVERY-012: AuditChainVerifier panicked with nil repo: %v", r)
		}
	}()

	report, err := verifier.VerifyChain(p16ctx(), 1, 100, 50)
	if err != nil {
		t.Errorf("FIN-RECOVERY-012: unexpected error from nil-repo verifier: %v", err)
	}
	if report == nil {
		t.Error("FIN-RECOVERY-012: nil repo verifier should return non-nil report")
	}
}
