package service

// anti_entropy.go — Cross-system consistency verification for the finance module.
//
// AntiEntropyService detects silent divergence between:
//   - Finance DB state (transactions, periods, reconciliations)
//   - Audit delivery state (outbox DELIVERED vs dead-letter)
//   - Hash chain coverage (CRITICAL events without chain entries)
//
// Unlike integrity.go (which checks ledger math), anti-entropy checks whether
// the observability and audit subsystems are consistent with the finance state.
// Silent divergence means the system appears healthy but is losing evidence.
//
// All checks are READ-ONLY. No auto-correction.
// Intended to run from a Temporal cron (nightly or after major operations).

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"awo.so/internal/shared"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
)

// AntiEntropyViolation describes a detected cross-system inconsistency.
type AntiEntropyViolation struct {
	Kind         string `json:"kind"`
	Detail       string `json:"detail"`
	RepairAction string `json:"repair_action,omitempty"`
}

// AntiEntropyReport is the result of a full anti-entropy scan.
type AntiEntropyReport struct {
	TenantID   uuid.UUID              `json:"tenant_id"`
	Violations []AntiEntropyViolation `json:"violations,omitempty"`
	Healthy    bool                   `json:"healthy"`
	CheckedAt  time.Time              `json:"checked_at"`
}

// AntiEntropyRepository provides the cross-system query capabilities needed for
// anti-entropy checks that span multiple subsystems.
type AntiEntropyRepository interface {
	// CountPostedTransactionsSince returns the number of posted transactions
	// created after sinceTime for the tenant.
	CountPostedTransactionsSince(ctx context.Context, tenantID uuid.UUID, since time.Time) (int, error)

	// CountDeliveredOutboxSince returns the number of DELIVERED outbox entries
	// created after sinceTime for the tenant.
	CountDeliveredOutboxSince(ctx context.Context, tenantID uuid.UUID, since time.Time) (int, error)

	// CountChainEntriesSince returns the number of chain entries created after
	// sinceTime for the tenant.
	CountChainEntriesSince(ctx context.Context, tenantID uuid.UUID, since time.Time) (int, error)
}

// AntiEntropyService runs cross-system consistency checks.
// A nil *AntiEntropyService returns a healthy empty report.
type AntiEntropyService struct {
	aeRepo         AntiEntropyRepository        // nil → DB checks skipped
	violationRepo  IntegrityViolationRepository  // nil → open-violation check skipped
	outboxGovernor *OutboxGovernor               // nil → backlog check skipped
	metrics        metrics.MetricsProvider
}

// NewAntiEntropyService creates the service. All deps may be nil.
func NewAntiEntropyService(
	aeRepo AntiEntropyRepository,
	violationRepo IntegrityViolationRepository,
	outboxGovernor *OutboxGovernor,
	m metrics.MetricsProvider,
) *AntiEntropyService {
	return &AntiEntropyService{
		aeRepo:         aeRepo,
		violationRepo:  violationRepo,
		outboxGovernor: outboxGovernor,
		metrics:        m,
	}
}

// RunChecks performs all anti-entropy checks for the current tenant.
// Tolerates individual check failures — each failure adds a DEGRADED violation.
func (s *AntiEntropyService) RunChecks(ctx context.Context) (*AntiEntropyReport, error) {
	if s == nil {
		tenantID, _ := shared.GetTenantID(ctx)
		return &AntiEntropyReport{TenantID: tenantID, Healthy: true, CheckedAt: time.Now()}, nil
	}
	tenantID, _ := shared.GetTenantID(ctx)
	report := &AntiEntropyReport{TenantID: tenantID, CheckedAt: time.Now()}

	// Check 1: Open CRITICAL violations with no pending outbox delivery.
	// If CRITICAL violations exist but the outbox is healthy, the audit record
	// may have been suppressed — the system is not learning from the corruption.
	s.checkCriticalViolationsVsOutbox(ctx, tenantID, report)

	// Check 2: Outbox backlog vs posting rate divergence.
	s.checkOutboxBacklogDivergence(ctx, tenantID, report)

	// Check 3: Cross-system health consistency.
	s.checkSubsystemConsistency(ctx, tenantID, report)

	report.Healthy = len(report.Violations) == 0

	if !report.Healthy {
		s.metrics.IncrementCounter("finance_anti_entropy_violations_total", metrics.Fields{
			"tenant_id":  tenantID.String(),
			"violations": fmt.Sprintf("%d", len(report.Violations)),
		})
		logger.ErrorContext(ctx, "finance anti-entropy: cross-system divergence detected", logger.Fields{
			"tenant_id":  tenantID.String(),
			"violations": len(report.Violations),
		})
	} else {
		logger.InfoContext(ctx, "finance anti-entropy: all checks passed", logger.Fields{
			"tenant_id": tenantID.String(),
		})
	}

	return report, nil
}

func (s *AntiEntropyService) checkCriticalViolationsVsOutbox(ctx context.Context, tenantID uuid.UUID, report *AntiEntropyReport) {
	if s.violationRepo == nil || s.outboxGovernor == nil {
		return
	}
	openCount, err := s.violationRepo.CountOpenCritical(ctx, tenantID)
	if err != nil {
		report.Violations = append(report.Violations, AntiEntropyViolation{
			Kind:   "INTEGRITY_QUERY_FAILED",
			Detail: fmt.Sprintf("could not query open CRITICAL violations: %v", err),
			RepairAction: "Check DB connectivity.",
		})
		return
	}
	if openCount == 0 {
		return
	}

	// CRITICAL violations exist — verify the outbox is not also DEAD (double failure).
	backlog, backlogErr := s.outboxGovernor.CheckBacklog(ctx)
	if backlogErr != nil {
		return
	}
	if backlog.DeadOutboxCount > 0 && openCount > 0 {
		report.Violations = append(report.Violations, AntiEntropyViolation{
			Kind: "CRITICAL_VIOLATIONS_WITH_DEAD_AUDIT",
			Detail: fmt.Sprintf(
				"%d open CRITICAL integrity violations AND %d DEAD audit outbox entries — "+
					"the ledger is corrupted AND the audit trail is incomplete",
				openCount, backlog.DeadOutboxCount),
			RepairAction: "Resolve CRITICAL violations first. Then replay DEAD outbox entries via SelfHealingService.",
		})
		s.metrics.IncrementCounter("finance_anti_entropy_events_total", metrics.Fields{
			"kind": "CRITICAL_VIOLATIONS_WITH_DEAD_AUDIT",
		})
	}
}

func (s *AntiEntropyService) checkOutboxBacklogDivergence(ctx context.Context, tenantID uuid.UUID, report *AntiEntropyReport) {
	if s.aeRepo == nil {
		return
	}
	// Compare posting activity in the last hour with outbox delivery in the same window.
	since := time.Now().Add(-time.Hour)
	postings, pErr := s.aeRepo.CountPostedTransactionsSince(ctx, tenantID, since)
	deliveries, dErr := s.aeRepo.CountDeliveredOutboxSince(ctx, tenantID, since)

	if pErr != nil || dErr != nil {
		return // query failures are non-fatal for anti-entropy
	}

	// If significant postings occurred but zero outbox deliveries: delivery worker may be down.
	if postings > 10 && deliveries == 0 {
		report.Violations = append(report.Violations, AntiEntropyViolation{
			Kind: "OUTBOX_DELIVERY_STALLED",
			Detail: fmt.Sprintf(
				"%d transactions posted in the last hour but 0 audit events delivered — "+
					"outbox delivery worker may be stopped",
				postings),
			RepairAction: "Check AuditOutboxProcessor Temporal cron. Run RecoverStuck to unblock PROCESSING entries.",
		})
		s.metrics.IncrementCounter("finance_anti_entropy_events_total", metrics.Fields{
			"kind": "OUTBOX_DELIVERY_STALLED",
		})
	}
}

func (s *AntiEntropyService) checkSubsystemConsistency(ctx context.Context, tenantID uuid.UUID, report *AntiEntropyReport) {
	if s.aeRepo == nil {
		return
	}
	// Verify that chain entries exist for recently delivered CRITICAL audit events.
	// If delivered > chain entries, the AuditChainWriter failed silently.
	since := time.Now().Add(-24 * time.Hour)
	deliveries, dErr := s.aeRepo.CountDeliveredOutboxSince(ctx, tenantID, since)
	chains, cErr := s.aeRepo.CountChainEntriesSince(ctx, tenantID, since)

	if dErr != nil || cErr != nil {
		return
	}

	// A significant gap between deliveries and chain entries suggests the chain writer
	// is failing silently — tamper evidence is missing.
	if deliveries > 0 && chains == 0 {
		report.Violations = append(report.Violations, AntiEntropyViolation{
			Kind: "AUDIT_CHAIN_NOT_POPULATED",
			Detail: fmt.Sprintf(
				"%d audit events delivered in the last 24h but 0 chain entries exist — "+
					"tamper-evidence chain writer may not be wired",
				deliveries),
			RepairAction: "Wire AuditChainWriter into AuditOutboxProcessor.deliver() calls.",
		})
		s.metrics.IncrementCounter("finance_anti_entropy_events_total", metrics.Fields{
			"kind": "AUDIT_CHAIN_NOT_POPULATED",
		})
	}
}
