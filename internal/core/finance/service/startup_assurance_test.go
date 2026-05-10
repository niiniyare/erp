// Package service_test — Phase 16 startup assurance contracts.
//
// Startup assurance: the finance module MUST fail to start (return an error from
// AssertStartupInvariants) when critical governance infrastructure is absent.
// Unsafe runtime startup must be impossible.
//
// Tests (FIN-STARTUP-*):
//
//	001 — Full wiring passes startup
//	002 — Nil registry detected at startup
//	003 — Nil enforcer detected at startup
//	004 — Extension with missing audit hook blocks startup
//	005 — Extension with missing safety check blocks startup
//	006 — Extension with all hooks passes startup
//	007 — Multiple violations all surfaced simultaneously
//	008 — Duplicate extension registration rejected
//	009 — ListContracts returns registered extensions
//	010 — AssuranceDashboard integrates startup result into report
package service_test

import (
	"testing"

	"awo.so/internal/core/finance/service"
	"awo.so/internal/shared/metrics"
)

// ── FIN-STARTUP-001: Full wiring passes ──────────────────────────────────────

func TestStartupAssurance_FullWiring_Passes(t *testing.T) {
	registry := service.NewGovernancePolicyRegistry(service.DefaultGovernancePolicy(), metrics.NewNoOpMetricsProvider())
	enforcer := service.NewSafetyEnforcer(service.DefaultSafetyPolicy(), metrics.NewNoOpMetricsProvider())
	guard := service.NewEvolutionSafetyGuard(registry, enforcer, metrics.NewNoOpMetricsProvider())

	if err := guard.AssertStartupInvariants(p16ctx()); err != nil {
		t.Errorf("FIN-STARTUP-001: full wiring should pass startup, got: %v", err)
	}
}

// ── FIN-STARTUP-002: Nil registry detected ───────────────────────────────────

func TestStartupAssurance_NilRegistry_BlocksStartup(t *testing.T) {
	enforcer := service.NewSafetyEnforcer(service.DefaultSafetyPolicy(), metrics.NewNoOpMetricsProvider())
	// Pass nil registry.
	guard := service.NewEvolutionSafetyGuard(nil, enforcer, metrics.NewNoOpMetricsProvider())

	err := guard.AssertStartupInvariants(p16ctx())
	if err == nil {
		t.Fatal("FIN-STARTUP-002 FAILED: nil registry should block startup")
	}
}

// ── FIN-STARTUP-003: Nil enforcer detected ───────────────────────────────────

func TestStartupAssurance_NilEnforcer_BlocksStartup(t *testing.T) {
	registry := service.NewGovernancePolicyRegistry(service.DefaultGovernancePolicy(), metrics.NewNoOpMetricsProvider())
	// Pass nil enforcer.
	guard := service.NewEvolutionSafetyGuard(registry, nil, metrics.NewNoOpMetricsProvider())

	err := guard.AssertStartupInvariants(p16ctx())
	if err == nil {
		t.Fatal("FIN-STARTUP-003 FAILED: nil enforcer should block startup")
	}
}

// ── FIN-STARTUP-004: Extension with missing audit hook blocked ───────────────

func TestStartupAssurance_ExtensionMissingAuditHook_BlocksStartup(t *testing.T) {
	registry := service.NewGovernancePolicyRegistry(service.DefaultGovernancePolicy(), metrics.NewNoOpMetricsProvider())
	enforcer := service.NewSafetyEnforcer(service.DefaultSafetyPolicy(), metrics.NewNoOpMetricsProvider())
	guard := service.NewEvolutionSafetyGuard(registry, enforcer, metrics.NewNoOpMetricsProvider())

	_ = guard.RegisterExtension(service.ExtensionContract{
		Name:                "p16_bad_extension_no_audit",
		Description:         "Extension missing audit hook",
		MutatesFinanceState: true,
		HasAuditHook:        false, // missing — must be detected
		HasSafetyCheck:      true,
		HasIntegrityCheck:   true,
	})

	err := guard.AssertStartupInvariants(p16ctx())
	if err == nil {
		t.Fatal("FIN-STARTUP-004 FAILED: extension without audit hook should block startup")
	}
	if !p16ContainsCode(err, "HasAuditHook") {
		t.Errorf("FIN-STARTUP-004: expected HasAuditHook in error message, got: %v", err)
	}
}

// ── FIN-STARTUP-005: Extension with missing safety check blocked ─────────────

func TestStartupAssurance_ExtensionMissingSafetyCheck_BlocksStartup(t *testing.T) {
	registry := service.NewGovernancePolicyRegistry(service.DefaultGovernancePolicy(), metrics.NewNoOpMetricsProvider())
	enforcer := service.NewSafetyEnforcer(service.DefaultSafetyPolicy(), metrics.NewNoOpMetricsProvider())
	guard := service.NewEvolutionSafetyGuard(registry, enforcer, metrics.NewNoOpMetricsProvider())

	_ = guard.RegisterExtension(service.ExtensionContract{
		Name:                "p16_bad_extension_no_safety",
		Description:         "Extension missing safety check",
		MutatesFinanceState: true,
		HasAuditHook:        true,
		HasSafetyCheck:      false, // missing — must be detected
		HasIntegrityCheck:   true,
	})

	err := guard.AssertStartupInvariants(p16ctx())
	if err == nil {
		t.Fatal("FIN-STARTUP-005 FAILED: extension without safety check should block startup")
	}
	if !p16ContainsCode(err, "HasSafetyCheck") {
		t.Errorf("FIN-STARTUP-005: expected HasSafetyCheck in error message, got: %v", err)
	}
}

// ── FIN-STARTUP-006: Extension with all hooks passes ─────────────────────────

func TestStartupAssurance_FullyCompliantExtension_Passes(t *testing.T) {
	registry := service.NewGovernancePolicyRegistry(service.DefaultGovernancePolicy(), metrics.NewNoOpMetricsProvider())
	enforcer := service.NewSafetyEnforcer(service.DefaultSafetyPolicy(), metrics.NewNoOpMetricsProvider())
	guard := service.NewEvolutionSafetyGuard(registry, enforcer, metrics.NewNoOpMetricsProvider())

	_ = guard.RegisterExtension(service.ExtensionContract{
		Name:                "p16_compliant_extension",
		Description:         "Extension with all required hooks",
		MutatesFinanceState: true,
		HasAuditHook:        true,
		HasSafetyCheck:      true,
		HasIntegrityCheck:   true,
	})

	if err := guard.AssertStartupInvariants(p16ctx()); err != nil {
		t.Errorf("FIN-STARTUP-006: fully compliant extension should pass startup, got: %v", err)
	}
}

// ── FIN-STARTUP-007: Multiple violations all surfaced simultaneously ──────────

func TestStartupAssurance_MultipleViolations_AllSurfaced(t *testing.T) {
	registry := service.NewGovernancePolicyRegistry(service.DefaultGovernancePolicy(), metrics.NewNoOpMetricsProvider())
	enforcer := service.NewSafetyEnforcer(service.DefaultSafetyPolicy(), metrics.NewNoOpMetricsProvider())
	guard := service.NewEvolutionSafetyGuard(registry, enforcer, metrics.NewNoOpMetricsProvider())

	// Register two bad extensions.
	_ = guard.RegisterExtension(service.ExtensionContract{
		Name: "p16_bad_ext_1", MutatesFinanceState: true,
		HasAuditHook: false, HasSafetyCheck: false, HasIntegrityCheck: true,
	})
	_ = guard.RegisterExtension(service.ExtensionContract{
		Name: "p16_bad_ext_2", MutatesFinanceState: true,
		HasAuditHook: true, HasSafetyCheck: false, HasIntegrityCheck: false,
	})

	violations := guard.ValidateExtensions()
	// Expect at least 3 violations: 2 for ext_1 (audit+safety) and 2 for ext_2 (safety+integrity).
	if len(violations) < 3 {
		t.Errorf("FIN-STARTUP-007: expected >=3 violations, got %d: %+v", len(violations), violations)
	}
}

// ── FIN-STARTUP-008: Duplicate extension registration rejected ────────────────

func TestStartupAssurance_DuplicateExtension_Rejected(t *testing.T) {
	registry := service.NewGovernancePolicyRegistry(service.DefaultGovernancePolicy(), metrics.NewNoOpMetricsProvider())
	enforcer := service.NewSafetyEnforcer(service.DefaultSafetyPolicy(), metrics.NewNoOpMetricsProvider())
	guard := service.NewEvolutionSafetyGuard(registry, enforcer, metrics.NewNoOpMetricsProvider())

	contract := service.ExtensionContract{
		Name:                "p16_duplicate_ext",
		MutatesFinanceState: true,
		HasAuditHook:        true,
		HasSafetyCheck:      true,
		HasIntegrityCheck:   true,
	}

	if err := guard.RegisterExtension(contract); err != nil {
		t.Fatalf("FIN-STARTUP-008: first registration should succeed, got: %v", err)
	}

	// Second registration of same name must be rejected.
	if err := guard.RegisterExtension(contract); err == nil {
		t.Fatal("FIN-STARTUP-008 FAILED: duplicate extension registration should be rejected")
	}
}

// ── FIN-STARTUP-009: ListContracts returns registered extensions ──────────────

func TestStartupAssurance_ListContracts_ReturnsRegistered(t *testing.T) {
	registry := service.NewGovernancePolicyRegistry(service.DefaultGovernancePolicy(), metrics.NewNoOpMetricsProvider())
	enforcer := service.NewSafetyEnforcer(service.DefaultSafetyPolicy(), metrics.NewNoOpMetricsProvider())
	guard := service.NewEvolutionSafetyGuard(registry, enforcer, metrics.NewNoOpMetricsProvider())

	names := []string{"p16_ext_a", "p16_ext_b", "p16_ext_c"}
	for _, name := range names {
		_ = guard.RegisterExtension(service.ExtensionContract{Name: name})
	}

	contracts := guard.ListContracts()
	if len(contracts) != len(names) {
		t.Errorf("FIN-STARTUP-009: expected %d contracts, got %d", len(names), len(contracts))
	}
}

// ── FIN-STARTUP-010: AssuranceDashboard integrates startup result ─────────────

// TestStartupAssurance_AssuranceDashboard_ReflectsStartupFailures verifies that
// startup invariant failures propagate into the assurance report.
// Operators can see startup failures in the dashboard, not just in logs.
func TestStartupAssurance_AssuranceDashboard_ReflectsStartupFailures(t *testing.T) {
	// Guard with nil enforcer — will fail startup.
	registry := service.NewGovernancePolicyRegistry(service.DefaultGovernancePolicy(), metrics.NewNoOpMetricsProvider())
	badGuard := service.NewEvolutionSafetyGuard(registry, nil, metrics.NewNoOpMetricsProvider())

	dashboard := service.NewAssuranceDashboard(badGuard, nil, nil, nil, true, true)
	report := dashboard.GenerateReport(p16ctx())

	if len(report.StartupFailures) == 0 {
		t.Error("FIN-STARTUP-010: startup failures should be reflected in assurance report")
	}
	if report.IsAcceptable() {
		t.Error("FIN-STARTUP-010: report with startup failures must not be acceptable")
	}
}
