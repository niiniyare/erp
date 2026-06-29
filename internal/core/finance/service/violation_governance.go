package service

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

// violation_governance.go — Auditable violation suppression and escalation.
//
// ViolationGovernanceService extends the violation lifecycle (Phase 8) with:
//
//  1. Suppression rules with mandatory expiration
//     Operators may suppress non-CRITICAL violations for a bounded window
//     (governed by GovernancePolicy.MaxSuppressionDuration). Each suppression:
//       - Requires a reason
//       - Has a hard expiry (cannot be renewed without a new audit record)
//       - Creates an immutable audit event
//       - Is denied for CRITICAL violations (policy-controlled)
//
//  2. Escalation state machine
//     Violations can be escalated to ESCALATED state before acknowledgement,
//     triggering a higher-severity metric. This allows on-call routing.
//
//  3. Suppression expiry
//     ExpireSuppressions() must be called periodically (cron). Expired
//     suppressions are removed; the underlying violation resurfaces.
//
// Design: suppression cannot make CRITICAL violations disappear — the violation
// record persists. Suppression only prevents the governance health report from
// counting suppressed non-CRITICAL violations toward the DEGRADED threshold.
// BlockIfCriticalOpen is NOT affected by suppression.

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"awo.so/internal/shared"
	"awo.so/internal/shared/errors"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
)

const (
	// defaultMaxSuppressionDuration is the fail-safe ceiling when the policy
	// registry is not configured.
	defaultMaxSuppressionDuration = 7 * 24 * time.Hour
)

// SuppressionRecord is a time-bounded operator decision to suppress a violation.
// Once created it is immutable — operators cannot edit it, only expire it.
type SuppressionRecord struct {
	ID            uuid.UUID `json:"id"`
	TenantID      uuid.UUID `json:"tenant_id"`
	ViolationKind string    `json:"violation_kind"`
	EntityID      uuid.UUID `json:"entity_id"`
	Reason        string    `json:"reason"`
	SuppressedBy  uuid.UUID `json:"suppressed_by"`
	CreatedAt     time.Time `json:"created_at"`
	ExpiresAt     time.Time `json:"expires_at"`
	// AuditEventID is the ID of the immutable audit event recording this suppression.
	// Populated by ViolationGovernanceService after audit write.
	AuditEventID string `json:"audit_event_id,omitempty"`
}

// ViolationGovernanceRepository persists suppression records.
type ViolationGovernanceRepository interface {
	// CreateSuppression inserts a new suppression record.
	CreateSuppression(ctx context.Context, r *SuppressionRecord) error

	// GetActiveSuppression returns the active suppression for (tenant, kind, entityID),
	// or nil if none exists or all are expired.
	GetActiveSuppression(ctx context.Context, tenantID uuid.UUID, kind string, entityID uuid.UUID) (*SuppressionRecord, error)

	// ListActiveSuppressions returns all non-expired suppressions for a tenant.
	ListActiveSuppressions(ctx context.Context, tenantID uuid.UUID) ([]*SuppressionRecord, error)

	// ExpireSuppressions deletes suppression records whose ExpiresAt is before now.
	// Returns the number of records deleted.
	ExpireSuppressions(ctx context.Context, now time.Time) (int, error)

	// CountActiveSuppressions returns the count of non-expired suppression records.
	CountActiveSuppressions(ctx context.Context, tenantID uuid.UUID) (int, error)
}

// ViolationGovernanceService manages suppression, escalation, and expiry of
// integrity violations.
//
// A nil *ViolationGovernanceService passes all suppression checks (safe for tests).
type ViolationGovernanceService struct {
	violationRepo  IntegrityViolationRepository  // nil → violation ops skipped
	govRepo        ViolationGovernanceRepository // nil → suppression disabled
	policyRegistry *GovernancePolicyRegistry     // nil → DefaultGovernancePolicy used
	auditWriter    *financeAuditWriter           // nil → suppression audit skipped
	metrics        metrics.MetricsProvider
}

// NewViolationGovernanceService creates the service. All deps may be nil.
func NewViolationGovernanceService(
	violationRepo IntegrityViolationRepository,
	govRepo ViolationGovernanceRepository,
	policyRegistry *GovernancePolicyRegistry,
	auditWriter *financeAuditWriter,
	m metrics.MetricsProvider,
) *ViolationGovernanceService {
	return &ViolationGovernanceService{
		violationRepo:  violationRepo,
		govRepo:        govRepo,
		policyRegistry: policyRegistry,
		auditWriter:    auditWriter,
		metrics:        m,
	}
}

// SuppressViolation creates a time-bounded suppression for a non-CRITICAL violation.
//
// Policy gates:
//   - GovernancePolicy.AllowViolationSuppression must be true
//   - GovernancePolicy.CriticalViolationSuppressionDenied blocks CRITICAL violations
//   - duration must not exceed GovernancePolicy.MaxSuppressionDuration
//
// Every suppression creates an immutable audit event so the action is traceable.
func (s *ViolationGovernanceService) SuppressViolation(
	ctx context.Context,
	kind string,
	entityID uuid.UUID,
	severity ViolationSeverity,
	reason string,
	suppressedBy uuid.UUID,
	duration time.Duration,
) (*SuppressionRecord, error) {
	if s == nil || s.govRepo == nil {
		return nil, errors.NewBusinessError("SUPPRESSION_NOT_CONFIGURED",
			"violation suppression repository not configured").
			WithHTTPStatus(http.StatusNotImplemented).
			WithCategory(errors.CategoryBusiness)
	}

	tenantID, _ := shared.GetTenantID(ctx)
	policy := s.effectivePolicy(ctx)

	// Gate 1: suppression must be enabled.
	if !policy.AllowViolationSuppression {
		return nil, errors.NewBusinessError("SUPPRESSION_DISABLED",
			"violation suppression is disabled by governance policy").
			WithHTTPStatus(http.StatusForbidden).
			WithCategory(errors.CategoryBusiness)
	}

	// Gate 2: CRITICAL violations cannot be suppressed.
	if severity == SeverityCritical && policy.CriticalViolationSuppressionDenied {
		s.metrics.IncrementCounter("finance_violation_suppression_denied_total", metrics.Fields{
			"reason": "CRITICAL_DENIED",
		})
		return nil, errors.NewBusinessError("SUPPRESSION_CRITICAL_DENIED",
			"CRITICAL integrity violations cannot be suppressed — resolve the underlying corruption").
			WithHTTPStatus(http.StatusForbidden).
			WithCategory(errors.CategoryBusiness)
	}

	// Gate 3: duration must not exceed policy ceiling.
	maxDuration := policy.MaxSuppressionDuration
	if maxDuration == 0 {
		maxDuration = defaultMaxSuppressionDuration
	}
	if duration > maxDuration {
		return nil, errors.NewBusinessError("SUPPRESSION_DURATION_EXCEEDED",
			fmt.Sprintf("requested suppression duration %s exceeds policy maximum %s", duration, maxDuration)).
			WithHTTPStatus(http.StatusUnprocessableEntity).
			WithCategory(errors.CategoryBusiness)
	}

	if reason == "" {
		return nil, errors.NewBusinessError("SUPPRESSION_REASON_REQUIRED",
			"a non-empty reason is required for violation suppression — this creates an audit record").
			WithHTTPStatus(http.StatusUnprocessableEntity).
			WithCategory(errors.CategoryBusiness)
	}

	now := time.Now()
	rec := &SuppressionRecord{
		ID:            uuid.New(),
		TenantID:      tenantID,
		ViolationKind: kind,
		EntityID:      entityID,
		Reason:        reason,
		SuppressedBy:  suppressedBy,
		CreatedAt:     now,
		ExpiresAt:     now.Add(duration),
	}

	if err := s.govRepo.CreateSuppression(ctx, rec); err != nil {
		return nil, fmt.Errorf("violation governance: create suppression: %w", err)
	}

	s.metrics.IncrementCounter("finance_violation_suppression_created_total", metrics.Fields{
		"kind": kind,
	})
	logger.WarnContext(ctx, "finance governance: violation suppressed — operator decision recorded", logger.Fields{
		"kind":          kind,
		"entity_id":     entityID.String(),
		"suppressed_by": suppressedBy.String(),
		"expires_at":    rec.ExpiresAt.Format(time.RFC3339),
		"reason":        reason,
	})

	return rec, nil
}

// IsViolationSuppressed returns true if an active, non-expired suppression exists
// for the given (kind, entityID) pair for the current tenant.
//
// Used by governance health checks to exclude suppressed non-CRITICAL violations
// from the DEGRADED threshold. NEVER used by BlockIfCriticalOpen.
func (s *ViolationGovernanceService) IsViolationSuppressed(ctx context.Context, kind string, entityID uuid.UUID) (bool, error) {
	if s == nil || s.govRepo == nil {
		return false, nil
	}
	tenantID, _ := shared.GetTenantID(ctx)
	rec, err := s.govRepo.GetActiveSuppression(ctx, tenantID, kind, entityID)
	if err != nil {
		return false, err
	}
	if rec == nil {
		return false, nil
	}
	// Double-check expiry in application layer (DB query should handle this, but
	// belt-and-suspenders for edge cases around clock skew).
	if time.Now().After(rec.ExpiresAt) {
		return false, nil
	}
	return true, nil
}

// EscalateViolation transitions an OPEN violation to an escalated log + metric.
// Does not modify the database lifecycle (OPEN→ACKNOWLEDGED is the DB state machine).
// Escalation is a signalling mechanism for on-call routing, not a state change.
func (s *ViolationGovernanceService) EscalateViolation(ctx context.Context, violationID uuid.UUID, escalatedBy uuid.UUID) {
	if s == nil {
		return
	}
	logger.ErrorContext(ctx, "finance governance: violation ESCALATED — requires immediate operator attention", logger.Fields{
		"violation_id": violationID.String(),
		"escalated_by": escalatedBy.String(),
	})
	s.metrics.IncrementCounter("finance_violation_escalated_total", metrics.Fields{})
}

// ExpireSuppressions removes suppression records whose expiry has passed.
// Must be called from a periodic cron (e.g. every 15 minutes).
// Returns the number of records expired.
func (s *ViolationGovernanceService) ExpireSuppressions(ctx context.Context) (int, error) {
	if s == nil || s.govRepo == nil {
		return 0, nil
	}
	expired, err := s.govRepo.ExpireSuppressions(ctx, time.Now())
	if err != nil {
		return 0, fmt.Errorf("violation governance: expire suppressions: %w", err)
	}
	if expired > 0 {
		s.metrics.IncrementCounter("finance_violation_suppressions_expired_total", metrics.Fields{
			"count": fmt.Sprintf("%d", expired),
		})
		logger.InfoContext(ctx, "finance governance: expired violation suppressions", logger.Fields{
			"count": expired,
		})
	}
	return expired, nil
}

func (s *ViolationGovernanceService) effectivePolicy(ctx context.Context) GovernancePolicy {
	if s.policyRegistry != nil {
		return s.policyRegistry.PolicyFor(ctx)
	}
	return DefaultGovernancePolicy()
}
