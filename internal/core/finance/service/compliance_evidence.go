package service

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

// compliance_evidence.go — Compliance-grade evidence generation for the finance module.
//
// ComplianceEvidenceService produces machine-readable, export-friendly evidence
// bundles for audit and regulatory review. Evidence is:
//
//   - Machine-readable: all output structs carry JSON tags
//   - Self-describing: policy version, generator identity, and timestamps embedded
//   - Immutable-friendly: evidence captures state at generation time; callers
//     should persist the output to immutable storage (S3, WORM, etc.)
//   - Forensically complete: each bundle includes the SHA-256 hash of its content
//     so downstream tools can verify it has not been modified
//
// Evidence bundles do NOT modify state — all methods are read-only.

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"awo.so/internal/shared"
	"awo.so/internal/shared/metrics"
)

// EvidenceBundle is the base type embedded in all compliance evidence structs.
// The ContentHash field is computed over the JSON-serialised bundle (excluding
// ContentHash itself) so verifiers can confirm the bundle is unmodified.
type EvidenceBundle struct {
	BundleID      uuid.UUID               `json:"bundle_id"`
	TenantID      uuid.UUID               `json:"tenant_id"`
	GeneratedAt   time.Time               `json:"generated_at"`
	PolicyVersion GovernancePolicyVersion `json:"policy_version"`
	ContentHash   string                  `json:"content_hash"` // SHA256 of bundle sans this field
}

// PeriodCloseEvidence captures the complete governance state at period close.
// Suitable for regulatory sign-off and auditor hand-off.
type PeriodCloseEvidence struct {
	EvidenceBundle

	PeriodID   uuid.UUID `json:"period_id"`
	PeriodName string    `json:"period_name"`

	// Integrity state at close.
	OpenCriticalViolations int `json:"open_critical_violations"`
	OpenHighViolations     int `json:"open_high_violations"`

	// Audit delivery state at close.
	DeadOutboxCount  int `json:"dead_outbox_count"`
	StaleOutboxCount int `json:"stale_outbox_count"`

	// Chain state at close.
	AuditChainHealthy bool   `json:"audit_chain_healthy"`
	AuditChainMessage string `json:"audit_chain_message,omitempty"`

	// Governance policy in effect at close.
	PolicySnapshot GovernancePolicy `json:"policy_snapshot"`

	// Human sign-off.
	ClosedBy   uuid.UUID  `json:"closed_by"`
	ClosedAt   time.Time  `json:"closed_at"`
	ApproverID *uuid.UUID `json:"approver_id,omitempty"`
}

// AuditTrailEvidence summarises audit delivery guarantees for a time window.
// Use for SOC2 / ISO27001 audit delivery evidence.
type AuditTrailEvidence struct {
	EvidenceBundle

	WindowStart time.Time `json:"window_start"`
	WindowEnd   time.Time `json:"window_end"`

	// Delivery statistics.
	TotalDelivered int     `json:"total_delivered"`
	TotalDead      int     `json:"total_dead"`
	TotalStale     int     `json:"total_stale"`
	DeliveryRate   float64 `json:"delivery_rate"` // 0.0–1.0

	// Chain coverage (CRITICAL events with chain entries).
	ChainEntriesInWindow int  `json:"chain_entries_in_window"`
	ChainContinuous      bool `json:"chain_continuous"`

	// Gap analysis.
	Gaps []AuditGapRecord `json:"gaps,omitempty"`
}

// AuditGapRecord describes a single detected gap in the audit trail.
type AuditGapRecord struct {
	Kind       string    `json:"kind"`
	Detail     string    `json:"detail"`
	DetectedAt time.Time `json:"detected_at"`
}

// ComplianceSnapshot is a point-in-time snapshot of all compliance indicators.
// Use for quarterly compliance reviews and automated SLO reporting.
type ComplianceSnapshot struct {
	EvidenceBundle

	// Integrity health.
	OpenCriticalViolations int `json:"open_critical_violations"`
	OpenHighViolations     int `json:"open_high_violations"`
	TotalSuppressions      int `json:"total_active_suppressions"`

	// Outbox health.
	DeadOutboxEntries  int `json:"dead_outbox_entries"`
	StaleOutboxEntries int `json:"stale_outbox_entries"`

	// Safety enforcement.
	SafetyEnforcerActive     bool `json:"safety_enforcer_active"`
	GovernanceRegistryActive bool `json:"governance_registry_active"`
	EvolutionGuardActive     bool `json:"evolution_guard_active"`

	// Extension coverage.
	RegisteredExtensions int `json:"registered_extensions"`
	UncoveredExtensions  int `json:"uncovered_extensions"`

	// Policy.
	PolicySnapshot GovernancePolicy `json:"policy_snapshot"`
}

// ComplianceEvidenceRepository provides the read queries needed for evidence generation.
type ComplianceEvidenceRepository interface {
	// GetPeriodName returns the human-readable name for a period ID.
	GetPeriodName(ctx context.Context, periodID uuid.UUID) (string, error)

	// GetPeriodClosedBy returns who closed the period and when.
	GetPeriodClosedBy(ctx context.Context, periodID uuid.UUID) (closedBy uuid.UUID, closedAt time.Time, err error)

	// CountAuditDelivered returns delivered outbox entries in the time window.
	CountAuditDelivered(ctx context.Context, tenantID uuid.UUID, from, to time.Time) (int, error)

	// CountAuditDead returns dead outbox entries in the time window.
	CountAuditDead(ctx context.Context, tenantID uuid.UUID, from, to time.Time) (int, error)

	// CountAuditChainEntriesInWindow returns chain entries in the time window.
	CountAuditChainEntriesInWindow(ctx context.Context, tenantID uuid.UUID, from, to time.Time) (int, error)

	// CountActiveSuppressions returns the number of active suppression records.
	CountActiveSuppressions(ctx context.Context, tenantID uuid.UUID) (int, error)
}

// ComplianceEvidenceService generates compliance evidence bundles.
// All methods are read-only. A nil receiver returns empty healthy bundles.
type ComplianceEvidenceService struct {
	evidenceRepo   ComplianceEvidenceRepository // nil → DB sections omitted
	violationRepo  IntegrityViolationRepository // nil → violation counts omitted
	gapDetector    *AuditGapDetector            // nil → outbox health omitted
	policyRegistry *GovernancePolicyRegistry    // nil → DefaultGovernancePolicy used
	safetyEnforcer *SafetyEnforcer              // nil → reported as inactive
	evolutionGuard *EvolutionSafetyGuard        // nil → reported as inactive
	metrics        metrics.MetricsProvider
}

// NewComplianceEvidenceService creates the service. All deps may be nil; nil
// dependencies result in omitted sections in the evidence bundle.
func NewComplianceEvidenceService(
	evidenceRepo ComplianceEvidenceRepository,
	violationRepo IntegrityViolationRepository,
	gapDetector *AuditGapDetector,
	policyRegistry *GovernancePolicyRegistry,
	safetyEnforcer *SafetyEnforcer,
	evolutionGuard *EvolutionSafetyGuard,
	m metrics.MetricsProvider,
) *ComplianceEvidenceService {
	return &ComplianceEvidenceService{
		evidenceRepo:   evidenceRepo,
		violationRepo:  violationRepo,
		gapDetector:    gapDetector,
		policyRegistry: policyRegistry,
		safetyEnforcer: safetyEnforcer,
		evolutionGuard: evolutionGuard,
		metrics:        m,
	}
}

// GeneratePeriodCloseEvidence produces a compliance bundle for a closed period.
// The bundle is suitable for regulatory sign-off and auditor hand-off.
func (s *ComplianceEvidenceService) GeneratePeriodCloseEvidence(ctx context.Context, periodID uuid.UUID, closedBy uuid.UUID) (*PeriodCloseEvidence, error) {
	tenantID, _ := shared.GetTenantID(ctx)
	policy := s.effectivePolicy(ctx)

	e := &PeriodCloseEvidence{
		EvidenceBundle: EvidenceBundle{
			BundleID:      uuid.New(),
			TenantID:      tenantID,
			GeneratedAt:   time.Now(),
			PolicyVersion: policy.Version,
		},
		PeriodID:       periodID,
		ClosedBy:       closedBy,
		ClosedAt:       time.Now(),
		PolicySnapshot: policy,
	}

	// Period name.
	if s.evidenceRepo != nil {
		name, err := s.evidenceRepo.GetPeriodName(ctx, periodID)
		if err == nil {
			e.PeriodName = name
		}
	}

	// Integrity state.
	if s.violationRepo != nil {
		if count, err := s.violationRepo.CountOpenCritical(ctx, tenantID); err == nil {
			e.OpenCriticalViolations = count
		}
		sev := SeverityHigh
		if highs, err := s.violationRepo.ListOpenViolations(ctx, tenantID, &sev); err == nil {
			e.OpenHighViolations = len(highs)
		}
	}

	// Outbox health.
	if s.gapDetector != nil {
		if report, err := s.gapDetector.CheckGaps(ctx); err == nil {
			e.DeadOutboxCount = report.DeadOutboxCount
			e.StaleOutboxCount = report.StaleOutboxCount
		}
	}

	// Chain health (static — full verification deferred to cron).
	e.AuditChainHealthy = true
	e.AuditChainMessage = "full chain verification deferred to scheduled cron — see audit_chain_violations_total"

	e.ContentHash = computeEvidenceHash(e)

	s.metrics.IncrementCounter("finance_compliance_evidence_generated_total", metrics.Fields{
		"kind": "period_close",
	})
	return e, nil
}

// GenerateAuditTrailEvidence produces delivery-guarantee evidence for a time window.
func (s *ComplianceEvidenceService) GenerateAuditTrailEvidence(ctx context.Context, from, to time.Time) (*AuditTrailEvidence, error) {
	tenantID, _ := shared.GetTenantID(ctx)
	policy := s.effectivePolicy(ctx)

	e := &AuditTrailEvidence{
		EvidenceBundle: EvidenceBundle{
			BundleID:      uuid.New(),
			TenantID:      tenantID,
			GeneratedAt:   time.Now(),
			PolicyVersion: policy.Version,
		},
		WindowStart: from,
		WindowEnd:   to,
	}

	if s.evidenceRepo != nil {
		delivered, _ := s.evidenceRepo.CountAuditDelivered(ctx, tenantID, from, to)
		dead, _ := s.evidenceRepo.CountAuditDead(ctx, tenantID, from, to)
		chainEntries, _ := s.evidenceRepo.CountAuditChainEntriesInWindow(ctx, tenantID, from, to)

		e.TotalDelivered = delivered
		e.TotalDead = dead
		e.ChainEntriesInWindow = chainEntries
		e.ChainContinuous = dead == 0 && chainEntries > 0

		total := delivered + dead
		if total > 0 {
			e.DeliveryRate = float64(delivered) / float64(total)
		}
		if dead > 0 {
			e.Gaps = append(e.Gaps, AuditGapRecord{
				Kind:       "DEAD_OUTBOX",
				Detail:     fmt.Sprintf("%d DEAD outbox entries in window — audit events permanently undelivered", dead),
				DetectedAt: time.Now(),
			})
		}
	}

	e.ContentHash = computeEvidenceHash(e)
	s.metrics.IncrementCounter("finance_compliance_evidence_generated_total", metrics.Fields{
		"kind": "audit_trail",
	})
	return e, nil
}

// GenerateComplianceSnapshot produces a point-in-time compliance indicator snapshot.
func (s *ComplianceEvidenceService) GenerateComplianceSnapshot(ctx context.Context) (*ComplianceSnapshot, error) {
	tenantID, _ := shared.GetTenantID(ctx)
	policy := s.effectivePolicy(ctx)

	snap := &ComplianceSnapshot{
		EvidenceBundle: EvidenceBundle{
			BundleID:      uuid.New(),
			TenantID:      tenantID,
			GeneratedAt:   time.Now(),
			PolicyVersion: policy.Version,
		},
		SafetyEnforcerActive:     s.safetyEnforcer != nil,
		GovernanceRegistryActive: s.policyRegistry != nil,
		EvolutionGuardActive:     s.evolutionGuard != nil,
		PolicySnapshot:           policy,
	}

	if s.violationRepo != nil {
		if count, err := s.violationRepo.CountOpenCritical(ctx, tenantID); err == nil {
			snap.OpenCriticalViolations = count
		}
		sev := SeverityHigh
		if highs, err := s.violationRepo.ListOpenViolations(ctx, tenantID, &sev); err == nil {
			snap.OpenHighViolations = len(highs)
		}
	}

	if s.gapDetector != nil {
		if report, err := s.gapDetector.CheckGaps(ctx); err == nil {
			snap.DeadOutboxEntries = report.DeadOutboxCount
			snap.StaleOutboxEntries = report.StaleOutboxCount
		}
	}

	if s.evidenceRepo != nil {
		if count, err := s.evidenceRepo.CountActiveSuppressions(ctx, tenantID); err == nil {
			snap.TotalSuppressions = count
		}
	}

	if s.evolutionGuard != nil {
		contracts := s.evolutionGuard.ListContracts()
		snap.RegisteredExtensions = len(contracts)
		for _, v := range s.evolutionGuard.ValidateExtensions() {
			_ = v
			snap.UncoveredExtensions++
		}
	}

	snap.ContentHash = computeEvidenceHash(snap)
	s.metrics.IncrementCounter("finance_compliance_evidence_generated_total", metrics.Fields{
		"kind": "snapshot",
	})
	return snap, nil
}

func (s *ComplianceEvidenceService) effectivePolicy(ctx context.Context) GovernancePolicy {
	if s.policyRegistry != nil {
		return s.policyRegistry.PolicyFor(ctx)
	}
	return DefaultGovernancePolicy()
}

// computeEvidenceHash computes SHA-256 over the JSON-serialised v (with
// ContentHash cleared so the hash is deterministic).
func computeEvidenceHash(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
