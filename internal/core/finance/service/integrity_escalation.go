package service

// integrity_escalation.go — Violation lifecycle tracking and automatic blocking.
//
// IntegrityEscalationService wraps IntegrityService with:
//   - Persistent violation records (OPEN → ACKNOWLEDGED → RESOLVED lifecycle)
//   - Automatic blocking of PostTransaction, CompleteReconciliation, and HardClose
//     when unresolved CRITICAL violations exist for the tenant.
//   - Re-open detection: a resolved violation that is re-detected triggers a
//     separate metric to surface recurring corruption.
//
// All methods degrade gracefully when deps are nil: scans are skipped and blocks
// are not enforced, so this can be wired as nil in tests.

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

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

// ViolationLifecycle tracks the resolution state of a persisted integrity violation.
type ViolationLifecycle string

const (
	ViolationOpen         ViolationLifecycle = "OPEN"
	ViolationAcknowledged ViolationLifecycle = "ACKNOWLEDGED"
	ViolationResolved     ViolationLifecycle = "RESOLVED"
)

// PersistedViolation is a stored record of a detected integrity violation with
// full lifecycle tracking for operator acknowledgement and resolution.
type PersistedViolation struct {
	ID         uuid.UUID          `json:"id"`
	TenantID   uuid.UUID          `json:"tenant_id"`
	Kind       string             `json:"kind"`
	Severity   ViolationSeverity  `json:"severity"`
	EntityID   uuid.UUID          `json:"entity_id"`
	Detail     string             `json:"detail"`
	Lifecycle  ViolationLifecycle `json:"lifecycle"`
	DetectedAt time.Time          `json:"detected_at"`
	AckedAt    *time.Time         `json:"acked_at,omitempty"`
	ResolvedAt *time.Time         `json:"resolved_at,omitempty"`
	AckedBy    *uuid.UUID         `json:"acked_by,omitempty"`
	ResolvedBy *uuid.UUID         `json:"resolved_by,omitempty"`
	RepairNote *string            `json:"repair_note,omitempty"`
}

// IntegrityViolationRepository persists and queries integrity violation records.
type IntegrityViolationRepository interface {
	// UpsertViolation inserts or updates the violation keyed on
	// (tenant_id, kind, entity_id). Returns whether a prior record existed and
	// its previous lifecycle so callers can detect re-opens.
	UpsertViolation(ctx context.Context, v *PersistedViolation) (existed bool, prior ViolationLifecycle, err error)

	// ListOpenViolations returns all OPEN or ACKNOWLEDGED violations for a tenant.
	// severity may be nil to return all severities.
	ListOpenViolations(ctx context.Context, tenantID uuid.UUID, severity *ViolationSeverity) ([]*PersistedViolation, error)

	// CountOpenCritical returns the number of OPEN CRITICAL violations for a tenant.
	CountOpenCritical(ctx context.Context, tenantID uuid.UUID) (int, error)

	// AcknowledgeViolation transitions a violation to ACKNOWLEDGED.
	AcknowledgeViolation(ctx context.Context, id uuid.UUID, byUserID uuid.UUID) error

	// ResolveViolation transitions a violation to RESOLVED with an optional note.
	ResolveViolation(ctx context.Context, id uuid.UUID, byUserID uuid.UUID, note string) error
}

// IntegrityEscalationService combines integrity scanning with violation persistence
// and automatic blocking of finance mutations.
type IntegrityEscalationService struct {
	integrityService IntegrityService             // nil → scans skipped
	violationRepo    IntegrityViolationRepository // nil → persistence skipped
	metrics          metrics.MetricsProvider
}

// NewIntegrityEscalationService creates the escalation service.
// Both integrityService and violationRepo may be nil (degrades gracefully).
func NewIntegrityEscalationService(
	integrityService IntegrityService,
	violationRepo IntegrityViolationRepository,
	m metrics.MetricsProvider,
) *IntegrityEscalationService {
	return &IntegrityEscalationService{
		integrityService: integrityService,
		violationRepo:    violationRepo,
		metrics:          m,
	}
}

// ScanAndEscalate runs a full integrity scan and persists newly detected violations.
// Returns the scan report. This is the scheduled path — call from Temporal cron.
func (s *IntegrityEscalationService) ScanAndEscalate(ctx context.Context) (*IntegrityReport, error) {
	if s == nil || s.integrityService == nil {
		return &IntegrityReport{}, nil
	}
	tenantID, _ := shared.GetTenantID(ctx)

	report, err := s.integrityService.ScanPostedTransactions(ctx, nil)
	if err != nil {
		return nil, err
	}

	if s.violationRepo == nil || len(report.Violations) == 0 {
		return report, nil
	}

	now := time.Now()
	for i := range report.Violations {
		v := report.Violations[i]
		pv := &PersistedViolation{
			ID:         uuid.New(),
			TenantID:   tenantID,
			Kind:       v.Kind,
			Severity:   v.Severity,
			EntityID:   v.EntityID,
			Detail:     v.Detail,
			Lifecycle:  ViolationOpen,
			DetectedAt: now,
		}
		existed, prior, upsertErr := s.violationRepo.UpsertViolation(ctx, pv)
		if upsertErr != nil {
			logger.ErrorContext(ctx, "integrity escalation: failed to persist violation", logger.Fields{
				"kind":      v.Kind,
				"entity_id": v.EntityID.String(),
				"error":     upsertErr.Error(),
			})
			continue
		}
		if !existed {
			s.metrics.IncrementCounter("finance_integrity_violations_persisted_total", metrics.Fields{
				"kind":     v.Kind,
				"severity": string(v.Severity),
			})
			logger.ErrorContext(ctx, "finance integrity: NEW violation detected and recorded", logger.Fields{
				"kind":      v.Kind,
				"severity":  string(v.Severity),
				"entity_id": v.EntityID.String(),
			})
		} else if prior == ViolationResolved {
			// Previously resolved violation re-detected — recurring corruption.
			s.metrics.IncrementCounter("finance_integrity_violations_reopened_total", metrics.Fields{
				"kind": v.Kind,
			})
			logger.WarnContext(ctx, "finance integrity: violation re-opened after resolution — recurring corruption", logger.Fields{
				"kind":      v.Kind,
				"entity_id": v.EntityID.String(),
			})
		}
	}
	return report, nil
}

// BlockIfCriticalOpen returns a blocking error when there are unresolved CRITICAL
// violations for the current tenant. Call before PostTransaction,
// CompleteReconciliation, and HardClose transitions.
//
// If the violation repo is unavailable the check is skipped (non-blocking degraded mode).
func (s *IntegrityEscalationService) BlockIfCriticalOpen(ctx context.Context) error {
	if s == nil || s.violationRepo == nil {
		return nil
	}
	tenantID, _ := shared.GetTenantID(ctx)
	count, err := s.violationRepo.CountOpenCritical(ctx, tenantID)
	if err != nil {
		// Query failure is non-blocking: log, allow operation.
		logger.WarnContext(ctx, "integrity escalation: cannot query open critical violations — check skipped", logger.Fields{
			"error": err.Error(),
		})
		return nil
	}
	if count > 0 {
		s.metrics.IncrementCounter("finance_integrity_block_total", metrics.Fields{
			"tenant_id": tenantID.String(),
		})
		return errors.NewBusinessError("INTEGRITY_CRITICAL_OPEN",
			fmt.Sprintf(
				"%d unresolved CRITICAL integrity violation(s) are blocking this operation. "+
					"Resolve all CRITICAL violations and retry. "+
					"Use ListOpenViolations to view details.", count)).
			WithHTTPStatus(http.StatusUnprocessableEntity).
			WithCategory(errors.CategoryBusiness)
	}
	return nil
}
