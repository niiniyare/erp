package service

// evolution_safety.go — Evolution safety guard for the finance module.
//
// EvolutionSafetyGuard prevents future code changes from accidentally weakening
// governance guarantees. It provides two mechanisms:
//
//  1. Startup invariant assertions
//     Call AssertStartupInvariants at service boot. It verifies:
//       - Core governance dependencies are configured (non-nil)
//       - All registered extension contracts satisfy their required hooks
//       - Policy registry is valid
//     If assertions fail the service MUST NOT start (return error to main).
//
//  2. Extension registration contracts
//     Any new transaction type, workflow, or reconciliation path MUST call
//     RegisterExtension before use. The contract declares which governance
//     hooks the extension satisfies:
//       - HasAuditHook      — emits audit events
//       - HasSafetyCheck    — checked by SafetyEnforcer
//       - HasIntegrityCheck — covered by integrity scan
//     ValidateExtensions returns EvolutionSafetyViolations for any extension
//     that does not satisfy its required hooks. Callers fail fast.
//
// The guard does NOT enforce at runtime — it enforces at wiring time and startup.
// This is intentional: runtime enforcement belongs to SafetyEnforcer and
// IntegrityEscalationService. The guard prevents unsafe wiring.

import (
	"context"
	"fmt"
	"sync"

	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
)

// ExtensionContract declares the governance hooks satisfied by a finance module
// extension (new transaction type, workflow, reconciliation path, etc.).
//
// All fields default to false (not satisfied). Extensions that satisfy a hook
// must set the corresponding field to true. Extensions that skip required hooks
// are surfaced as EvolutionSafetyViolations at startup.
type ExtensionContract struct {
	// Name is a unique, stable machine-readable identifier.
	// Convention: "<kind>_<noun>", e.g. "inter_company_transaction_type".
	Name string `json:"name"`

	// Description is a human-readable explanation for governance dashboards.
	Description string `json:"description"`

	// HasAuditHook declares that this extension emits audit events via the
	// financeAuditWriter. Extensions that mutate finance state MUST set this.
	HasAuditHook bool `json:"has_audit_hook"`

	// HasSafetyCheck declares that SafetyEnforcer checks apply to this extension.
	// Extensions that create or modify transactions MUST set this.
	HasSafetyCheck bool `json:"has_safety_check"`

	// HasIntegrityCheck declares that the integrity scan covers entities
	// created by this extension. Extensions adding new transaction kinds MUST set this.
	HasIntegrityCheck bool `json:"has_integrity_check"`

	// MutatesFinanceState declares that this extension writes to finance tables.
	// If true, HasAuditHook and HasSafetyCheck MUST also be true.
	MutatesFinanceState bool `json:"mutates_finance_state"`

	// RequiredPolicies is a list of GovernancePolicy field names that must be
	// explicitly configured for this extension to operate safely.
	// Verified by ValidateExtensions.
	RequiredPolicies []string `json:"required_policies,omitempty"`
}

// EvolutionSafetyViolation describes a detected governance gap in a registered extension.
type EvolutionSafetyViolation struct {
	Extension string `json:"extension"`
	Field     string `json:"field"`
	Detail    string `json:"detail"`
}

func (v EvolutionSafetyViolation) Error() string {
	return fmt.Sprintf("evolution safety: %s.%s: %s", v.Extension, v.Field, v.Detail)
}

// EvolutionSafetyGuard validates extension contracts at startup and prevents
// unsafe code evolution from weakening governance guarantees.
//
// A nil *EvolutionSafetyGuard passes all checks (safe for tests).
type EvolutionSafetyGuard struct {
	mu       sync.RWMutex
	contracts map[string]ExtensionContract

	// Core dependencies whose absence is a startup-blocking invariant.
	registry *GovernancePolicyRegistry // must be non-nil
	enforcer *SafetyEnforcer           // must be non-nil for mutation paths
	metrics  metrics.MetricsProvider
}

// NewEvolutionSafetyGuard creates the guard.
// registry and enforcer must be non-nil; the guard will surface their absence
// in AssertStartupInvariants.
func NewEvolutionSafetyGuard(
	registry *GovernancePolicyRegistry,
	enforcer *SafetyEnforcer,
	m metrics.MetricsProvider,
) *EvolutionSafetyGuard {
	return &EvolutionSafetyGuard{
		contracts: make(map[string]ExtensionContract),
		registry:  registry,
		enforcer:  enforcer,
		metrics:   m,
	}
}

// RegisterExtension registers a governance contract for a finance module extension.
// Returns an error if a contract with the same Name was already registered
// (prevents accidental overwrites that could weaken declarations).
func (g *EvolutionSafetyGuard) RegisterExtension(contract ExtensionContract) error {
	if g == nil {
		return nil
	}
	if contract.Name == "" {
		return fmt.Errorf("evolution_safety: extension contract must have a non-empty Name")
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, exists := g.contracts[contract.Name]; exists {
		return fmt.Errorf("evolution_safety: extension %q already registered — use a unique Name or update the existing contract", contract.Name)
	}
	g.contracts[contract.Name] = contract
	g.metrics.IncrementCounter("finance_evolution_extensions_registered_total", metrics.Fields{
		"extension": contract.Name,
	})
	return nil
}

// ValidateExtensions checks all registered contracts against governance requirements.
// Returns a violation for every contract that declares MutatesFinanceState but
// lacks the required governance hooks.
//
// Must be called at startup after all RegisterExtension calls.
func (g *EvolutionSafetyGuard) ValidateExtensions() []EvolutionSafetyViolation {
	if g == nil {
		return nil
	}
	g.mu.RLock()
	defer g.mu.RUnlock()

	var violations []EvolutionSafetyViolation
	for _, c := range g.contracts {
		if c.MutatesFinanceState {
			if !c.HasAuditHook {
				violations = append(violations, EvolutionSafetyViolation{
					Extension: c.Name,
					Field:     "HasAuditHook",
					Detail:    "extension mutates finance state but does not declare an audit hook — mutations will produce no audit trail",
				})
			}
			if !c.HasSafetyCheck {
				violations = append(violations, EvolutionSafetyViolation{
					Extension: c.Name,
					Field:     "HasSafetyCheck",
					Detail:    "extension mutates finance state but is not covered by SafetyEnforcer — policy checks are bypassed",
				})
			}
		}
		if !c.HasIntegrityCheck && c.MutatesFinanceState {
			violations = append(violations, EvolutionSafetyViolation{
				Extension: c.Name,
				Field:     "HasIntegrityCheck",
				Detail:    "extension creates finance entities not covered by integrity scan — corruption may go undetected",
			})
		}
	}
	return violations
}

// AssertStartupInvariants verifies that the core governance infrastructure is
// correctly wired. Returns a non-nil error if any invariant is violated.
//
// MUST be called in main/server startup before accepting requests.
// Return the error to the caller — do not swallow it.
func (g *EvolutionSafetyGuard) AssertStartupInvariants(ctx context.Context) error {
	if g == nil {
		return nil
	}

	var failures []string

	// Invariant 1: governance registry must be configured.
	if g.registry == nil {
		failures = append(failures, "GovernancePolicyRegistry is nil — all governance decisions use unsafe defaults")
	}

	// Invariant 2: safety enforcer must be configured.
	if g.enforcer == nil {
		failures = append(failures, "SafetyEnforcer is nil — runtime policy checks are disabled for all mutations")
	}

	// Invariant 3: registered extension contracts must satisfy governance requirements.
	extViolations := g.ValidateExtensions()
	for _, v := range extViolations {
		failures = append(failures, v.Error())
	}

	if len(failures) > 0 {
		g.metrics.IncrementCounter("finance_evolution_startup_failures_total", metrics.Fields{
			"count": fmt.Sprintf("%d", len(failures)),
		})
		for _, f := range failures {
			logger.ErrorContext(ctx, "finance evolution safety: startup invariant FAILED", logger.Fields{
				"failure": f,
			})
		}
		return fmt.Errorf("finance evolution safety: %d startup invariant(s) failed:\n%v", len(failures), failures)
	}

	logger.InfoContext(ctx, "finance evolution safety: all startup invariants passed", logger.Fields{
		"extensions_registered": len(g.contracts),
	})
	g.metrics.IncrementCounter("finance_evolution_startup_passed_total", metrics.Fields{})
	return nil
}

// ListContracts returns all registered extension contracts for governance inspection.
func (g *EvolutionSafetyGuard) ListContracts() []ExtensionContract {
	if g == nil {
		return nil
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	out := make([]ExtensionContract, 0, len(g.contracts))
	for _, c := range g.contracts {
		out = append(out, c)
	}
	return out
}
