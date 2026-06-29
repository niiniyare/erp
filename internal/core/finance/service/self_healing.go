package service

// self_healing.go — Safe automated recovery for the finance audit subsystem.
//
// SelfHealingService identifies and executes provably safe recovery actions:
//
//   SAFE (auto-execute):
//   - Recover stuck PROCESSING outbox entries → reset to PENDING for re-delivery.
//     Safe because idempotency keys prevent duplicate delivery.
//   - Identify orphaned DRAFT transactions → surface candidates only (no mutation).
//
//   UNSAFE (surface diagnostics, escalate to operators):
//   - Reversal chain repairs       → requires human approval
//   - Integrity violation repair   → requires human + finance controller sign-off
//   - GL entry corrections         → NEVER automated
//
// All methods are nil-safe and read-only except RecoverStuckOutbox.

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

const (
	// OrphanedDraftAge is the age after which a DRAFT transaction is considered
	// orphaned and surfaced as a recovery candidate.
	OrphanedDraftAge = 72 * time.Hour
)

// SelfHealingReport describes the outcome of a self-healing run.
type SelfHealingReport struct {
	TenantID uuid.UUID `json:"tenant_id"`

	// RecoveredOutboxEntries is the number of stuck PROCESSING entries reset to PENDING.
	RecoveredOutboxEntries int `json:"recovered_outbox_entries"`

	// OrphanedDraftCount is the number of DRAFT transactions older than OrphanedDraftAge.
	// These are candidates for manual cancellation — not automatically cancelled.
	OrphanedDraftCount int `json:"orphaned_draft_count"`

	// EscalationRequired lists actions the system cannot safely automate.
	EscalationRequired []string `json:"escalation_required,omitempty"`

	// Timestamp is when the healing run completed.
	Timestamp time.Time `json:"timestamp"`
}

// SelfHealingRepository provides the queries needed for safe recovery decisions.
type SelfHealingRepository interface {
	// CountOrphanedDrafts returns the number of DRAFT transactions older than maxAge.
	CountOrphanedDrafts(ctx context.Context, tenantID uuid.UUID, maxAge time.Duration) (int, error)
}

// SelfHealingService executes provably safe automated recovery actions.
type SelfHealingService struct {
	outboxGovernor *OutboxGovernor              // nil → outbox recovery skipped
	healingRepo    SelfHealingRepository        // nil → orphan check skipped
	violationRepo  IntegrityViolationRepository // nil → violation escalation skipped
	metrics        metrics.MetricsProvider
}

// NewSelfHealingService creates the service. All deps may be nil.
func NewSelfHealingService(
	outboxGovernor *OutboxGovernor,
	healingRepo SelfHealingRepository,
	violationRepo IntegrityViolationRepository,
	m metrics.MetricsProvider,
) *SelfHealingService {
	return &SelfHealingService{
		outboxGovernor: outboxGovernor,
		healingRepo:    healingRepo,
		violationRepo:  violationRepo,
		metrics:        m,
	}
}

// RunHealingCycle performs all safe automated recovery actions and returns a report.
// This is the primary entry point — call from a Temporal cron after each anti-entropy scan.
func (s *SelfHealingService) RunHealingCycle(ctx context.Context) (*SelfHealingReport, error) {
	if s == nil {
		tenantID, _ := shared.GetTenantID(ctx)
		return &SelfHealingReport{TenantID: tenantID, Timestamp: time.Now()}, nil
	}
	tenantID, _ := shared.GetTenantID(ctx)
	report := &SelfHealingReport{TenantID: tenantID, Timestamp: time.Now()}

	// Safe action 1: recover stuck outbox entries.
	recovered, err := s.recoverOutbox(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "self-healing: outbox recovery failed", logger.Fields{"error": err.Error()})
		report.EscalationRequired = append(report.EscalationRequired,
			fmt.Sprintf("Outbox recovery failed: %v — manual restart of AuditOutboxProcessor may be needed", err))
	} else {
		report.RecoveredOutboxEntries = recovered
	}

	// Safe action 2: surface orphaned drafts (read-only).
	orphaned, oErr := s.detectOrphanedDrafts(ctx, tenantID)
	if oErr != nil {
		logger.WarnContext(ctx, "self-healing: orphaned draft detection failed", logger.Fields{"error": oErr.Error()})
	} else {
		report.OrphanedDraftCount = orphaned
		if orphaned > 0 {
			report.EscalationRequired = append(report.EscalationRequired,
				fmt.Sprintf("%d orphaned DRAFT transactions (>%s old) require manual review and cancellation", orphaned, OrphanedDraftAge))
		}
	}

	// Identify (but not execute) unsafe actions requiring human intervention.
	s.identifyEscalations(ctx, tenantID, report)

	s.metrics.IncrementCounter("finance_self_healing_cycles_total", metrics.Fields{
		"recovered_outbox": fmt.Sprintf("%d", report.RecoveredOutboxEntries),
		"orphaned_drafts":  fmt.Sprintf("%d", report.OrphanedDraftCount),
	})

	return report, nil
}

// recoverOutbox resets stuck PROCESSING outbox entries to PENDING.
// Safe: idempotency keys prevent duplicate delivery even if processor re-runs these.
func (s *SelfHealingService) recoverOutbox(ctx context.Context) (int, error) {
	if s.outboxGovernor == nil {
		return 0, nil
	}
	return s.outboxGovernor.RecoverStuck(ctx)
}

// detectOrphanedDrafts counts DRAFT transactions older than OrphanedDraftAge.
// Returns count only — does NOT cancel them. Cancellation requires human approval.
func (s *SelfHealingService) detectOrphanedDrafts(ctx context.Context, tenantID uuid.UUID) (int, error) {
	if s.healingRepo == nil {
		return 0, nil
	}
	return s.healingRepo.CountOrphanedDrafts(ctx, tenantID, OrphanedDraftAge)
}

// identifyEscalations surfaces actions that require human intervention and appends
// them to report.EscalationRequired. No mutations performed here.
func (s *SelfHealingService) identifyEscalations(ctx context.Context, tenantID uuid.UUID, report *SelfHealingReport) {
	if s.violationRepo == nil {
		return
	}
	count, err := s.violationRepo.CountOpenCritical(ctx, tenantID)
	if err != nil {
		return
	}
	if count > 0 {
		report.EscalationRequired = append(report.EscalationRequired,
			fmt.Sprintf(
				"%d open CRITICAL integrity violations require manual resolution by a finance controller. "+
					"Use ReverseTransaction + correcting journal entries. DO NOT modify entries directly.", count))
	}
}
