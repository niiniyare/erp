package service

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

// governance_registry.go — Centralized governance policy registry.
//
// GovernancePolicyRegistry is the single source of truth for ALL runtime
// governance rules in the finance module. Replaces scattered per-service
// policy fields with a formal, versioned, tenant-aware registry.
//
// Design principles:
//   - Every governance decision must reference a named policy from this registry
//   - No hidden policy logic outside this package
//   - Policies are machine-readable and export-friendly (JSON tags on all fields)
//   - Tenant overrides cascade from global defaults
//   - ValidatePolicy catches unsafe configurations at wiring time
//
// Usage:
//   registry := NewGovernancePolicyRegistry(DefaultGovernancePolicy(), metrics)
//   registry.SetTenantPolicy(tenantID, customPolicy) // for high-volume tenants
//   p := registry.PolicyFor(ctx)                     // in service methods

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"awo.so/internal/shared"
	"awo.so/internal/shared/metrics"
)

// GovernancePolicyVersion is monotonically incremented on every policy change.
type GovernancePolicyVersion int

// GovernancePolicy is a complete, versioned set of governance rules for one tenant
// (or the global default). All thresholds use zero-value = "disabled" semantics
// EXCEPT the boolean flags, which default to the safer value.
//
// Zero-value policies are intentionally permissive on numeric thresholds and
// strict on boolean gates, so callers always receive the right fail-safe
// behavior from DefaultGovernancePolicy().
type GovernancePolicy struct {
	Version GovernancePolicyVersion `json:"version"`

	// ── Period close requirements ─────────────────────────────────────────────

	// RequireZeroCriticalViolations blocks HardClose when any CRITICAL integrity
	// violation is OPEN or ACKNOWLEDGED.
	RequireZeroCriticalViolations bool `json:"require_zero_critical_violations"`

	// RequireReconciliationComplete blocks HardClose when open bank statements exist.
	RequireReconciliationComplete bool `json:"require_reconciliation_complete"`

	// PeriodCloseGracePeriod is the maximum window for late adjustments before
	// HardClose is permitted.
	PeriodCloseGracePeriod time.Duration `json:"period_close_grace_period"`

	// ── Approval thresholds ──────────────────────────────────────────────────

	// ApprovalRequiredAbove requires at least one approver for transactions above
	// this amount. Zero = no threshold.
	ApprovalRequiredAbove decimal.Decimal `json:"approval_required_above"`

	// DualApprovalAbove requires two distinct approvers above this amount.
	// Zero = no threshold.
	DualApprovalAbove decimal.Decimal `json:"dual_approval_above"`

	// ── Audit criticality ────────────────────────────────────────────────────

	// CriticalAuditEventTypes lists event types that must be written to the
	// tamper-evident audit chain. Any type in this list that lacks a chain entry
	// is reported by AntiEntropyService as AUDIT_CHAIN_NOT_POPULATED.
	CriticalAuditEventTypes []string `json:"critical_audit_event_types"`

	// AuditDeliveryMaxLatency is the SLO for outbox delivery. Entries older
	// than this in PENDING state are flagged as stale by AuditGapDetector.
	AuditDeliveryMaxLatency time.Duration `json:"audit_delivery_max_latency"`

	// RequireAuditChainForCritical causes the governance health report to show
	// DEGRADED when AuditChainVerifier is not configured.
	RequireAuditChainForCritical bool `json:"require_audit_chain_for_critical"`

	// ── Anomaly escalation ───────────────────────────────────────────────────

	// AnomalyEscalationThreshold is the number of anomaly events within
	// AnomalyWindowDuration before an escalation metric fires.
	AnomalyEscalationThreshold int `json:"anomaly_escalation_threshold"`

	// AnomalyWindowDuration is the sliding window for anomaly counting.
	AnomalyWindowDuration time.Duration `json:"anomaly_window_duration"`

	// ── Tenant resource ceilings ─────────────────────────────────────────────

	// MaxOpenTransactions is the ceiling on DRAFT+PENDING_APPROVAL transactions
	// per tenant at any given time. Zero = unlimited.
	MaxOpenTransactions int `json:"max_open_transactions"`

	// MaxMonthlyPostingVolume is the ceiling on total debit posted per tenant
	// per calendar month. Zero = unlimited.
	MaxMonthlyPostingVolume decimal.Decimal `json:"max_monthly_posting_volume"`

	// ── Reconciliation safety ────────────────────────────────────────────────

	// ReconciliationStaleDays is the number of days after which an IN_PROGRESS
	// bank statement is flagged stale by OperationsService.
	ReconciliationStaleDays int `json:"reconciliation_stale_days"`

	// MaxUnmatchedLineRatio is the maximum fraction of unmatched lines allowed
	// before CompleteReconciliation is blocked (0.0–1.0). Zero = no gate.
	MaxUnmatchedLineRatio float64 `json:"max_unmatched_line_ratio"`

	// ── Integrity violation suppression ──────────────────────────────────────

	// AllowViolationSuppression permits operators to suppress non-critical
	// violations. Suppression is always auditable regardless of this setting.
	AllowViolationSuppression bool `json:"allow_violation_suppression"`

	// MaxSuppressionDuration caps how long a suppression may be active.
	// Zero = 7 days (fail-safe default applied in ViolationGovernanceService).
	MaxSuppressionDuration time.Duration `json:"max_suppression_duration"`

	// CriticalViolationSuppressionDenied prevents suppression of CRITICAL
	// violations regardless of AllowViolationSuppression.
	CriticalViolationSuppressionDenied bool `json:"critical_violation_suppression_denied"`
}

// DefaultGovernancePolicy returns conservative, production-safe governance defaults.
// All boolean gates default to ENABLED. All numeric thresholds default to the
// values proven safe in phase 8 testing.
func DefaultGovernancePolicy() GovernancePolicy {
	return GovernancePolicy{
		Version:                       1,
		RequireZeroCriticalViolations: true,
		RequireReconciliationComplete: true,
		PeriodCloseGracePeriod:        24 * time.Hour,
		ApprovalRequiredAbove:         decimal.NewFromInt(50_000),
		DualApprovalAbove:             decimal.NewFromInt(500_000),
		CriticalAuditEventTypes: []string{
			"TRANSACTION_POSTED",
			"PERIOD_HARD_CLOSED",
			"RECONCILIATION_COMPLETED",
			"REVERSAL_POSTED",
		},
		AuditDeliveryMaxLatency:            5 * time.Minute,
		RequireAuditChainForCritical:       true,
		AnomalyEscalationThreshold:         10,
		AnomalyWindowDuration:              time.Hour,
		MaxOpenTransactions:                10_000,
		MaxMonthlyPostingVolume:            decimal.NewFromInt(1_000_000_000), // 1 B
		ReconciliationStaleDays:            30,
		MaxUnmatchedLineRatio:              0.05,
		AllowViolationSuppression:          true,
		MaxSuppressionDuration:             7 * 24 * time.Hour,
		CriticalViolationSuppressionDenied: true,
	}
}

// PolicyValidationError describes a detected unsafe policy configuration.
type PolicyValidationError struct {
	Field   string
	Message string
}

func (e PolicyValidationError) Error() string {
	return fmt.Sprintf("policy validation: %s: %s", e.Field, e.Message)
}

// ValidatePolicy returns all unsafe configurations found in p.
// An empty slice means p is safe to activate.
func ValidatePolicy(p GovernancePolicy) []PolicyValidationError {
	var errs []PolicyValidationError

	if !p.RequireZeroCriticalViolations {
		errs = append(errs, PolicyValidationError{
			Field:   "RequireZeroCriticalViolations",
			Message: "disabling this allows period close with active CRITICAL integrity violations — unsafe",
		})
	}
	if p.MaxSuppressionDuration > 30*24*time.Hour {
		errs = append(errs, PolicyValidationError{
			Field:   "MaxSuppressionDuration",
			Message: "suppression window exceeds 30 days — violates compliance minimum review frequency",
		})
	}
	if !p.CriticalViolationSuppressionDenied {
		errs = append(errs, PolicyValidationError{
			Field:   "CriticalViolationSuppressionDenied",
			Message: "allowing suppression of CRITICAL violations can hide ledger corruption",
		})
	}
	if p.AuditDeliveryMaxLatency > time.Hour {
		errs = append(errs, PolicyValidationError{
			Field:   "AuditDeliveryMaxLatency",
			Message: "latency SLO exceeds 1 hour — forensic trail gaps become difficult to explain",
		})
	}
	return errs
}

// GovernancePolicyRegistry is the tenant-aware, versioned store for all
// governance policies in the finance module.
//
// Thread-safe: all methods are safe for concurrent use.
type GovernancePolicyRegistry struct {
	mu      sync.RWMutex
	global  GovernancePolicy
	tenants map[uuid.UUID]GovernancePolicy
	metrics metrics.MetricsProvider
}

// NewGovernancePolicyRegistry creates the registry with the given global default.
// Panics if global fails ValidatePolicy — prevents silent misconfiguration at startup.
func NewGovernancePolicyRegistry(global GovernancePolicy, m metrics.MetricsProvider) *GovernancePolicyRegistry {
	if errs := ValidatePolicy(global); len(errs) > 0 {
		// Surface at construction time so wiring fails fast.
		panic(fmt.Sprintf("governance_registry: unsafe global policy: %v", errs))
	}
	return &GovernancePolicyRegistry{
		global:  global,
		tenants: make(map[uuid.UUID]GovernancePolicy),
		metrics: m,
	}
}

// SetTenantPolicy installs a per-tenant policy override. Returns an error if
// the policy fails validation — callers must handle this; panicking on
// per-tenant config would be too aggressive.
func (r *GovernancePolicyRegistry) SetTenantPolicy(tenantID uuid.UUID, p GovernancePolicy) []PolicyValidationError {
	if errs := ValidatePolicy(p); len(errs) > 0 {
		r.metrics.IncrementCounter("finance_governance_policy_invalid_total", metrics.Fields{
			"tenant_id": tenantID.String(),
		})
		return errs
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tenants[tenantID] = p
	r.metrics.IncrementCounter("finance_governance_policy_updated_total", metrics.Fields{
		"tenant_id": tenantID.String(),
	})
	return nil
}

// PolicyFor returns the effective policy for the tenant in ctx.
// Falls back to global if no per-tenant override is configured.
func (r *GovernancePolicyRegistry) PolicyFor(ctx context.Context) GovernancePolicy {
	if r == nil {
		return DefaultGovernancePolicy()
	}
	tenantID, _ := shared.GetTenantID(ctx)
	return r.policyForID(tenantID)
}

// PolicyForID returns the effective policy for an explicit tenant ID.
func (r *GovernancePolicyRegistry) PolicyForID(tenantID uuid.UUID) GovernancePolicy {
	if r == nil {
		return DefaultGovernancePolicy()
	}
	return r.policyForID(tenantID)
}

func (r *GovernancePolicyRegistry) policyForID(tenantID uuid.UUID) GovernancePolicy {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if p, ok := r.tenants[tenantID]; ok {
		return p
	}
	return r.global
}

// GlobalPolicy returns the global default policy.
func (r *GovernancePolicyRegistry) GlobalPolicy() GovernancePolicy {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.global
}

// Snapshot returns a copy of all active policies keyed by "global" or tenant UUID string.
// Used by governance dashboard and compliance evidence exports.
func (r *GovernancePolicyRegistry) Snapshot() map[string]GovernancePolicy {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]GovernancePolicy, len(r.tenants)+1)
	out["global"] = r.global
	for id, p := range r.tenants {
		out[id.String()] = p
	}
	return out
}
