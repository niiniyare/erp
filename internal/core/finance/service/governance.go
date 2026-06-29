package service

// governance.go — Finance governance dashboard readiness.
//
// FinanceGovernanceService aggregates health signals from all finance subsystems
// into a single machine-readable FinanceHealthReport suitable for:
//   - Governance dashboards
//   - Automated health checks
//   - Incident management tooling
//   - Compliance sign-off workflows
//
// All fields use consistent severity indicators (healthy/degraded/critical)
// so dashboard tooling can colour-code without parsing string values.

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// SubsystemStatus is a machine-readable subsystem health indicator used by the
// governance dashboard. Distinct from HealthStatus (GREEN/YELLOW/RED) which is
// used by OperationsService for the operator health report.
type SubsystemStatus string

const (
	SubsystemStatusHealthy  SubsystemStatus = "HEALTHY"
	SubsystemStatusDegraded SubsystemStatus = "DEGRADED"
	SubsystemStatusCritical SubsystemStatus = "CRITICAL"
)

// SubsystemHealth is the health status of a single finance subsystem.
type SubsystemHealth struct {
	Status  SubsystemStatus `json:"status"`
	Message string          `json:"message,omitempty"`
	Count   int             `json:"count,omitempty"` // relevant count (violations, dead entries, etc.)
}

// FinanceHealthReport is a structured, machine-readable summary of finance module health.
// Suitable for governance dashboards, health endpoints, and compliance workflows.
type FinanceHealthReport struct {
	TenantID  uuid.UUID       `json:"tenant_id"`
	Overall   SubsystemStatus `json:"overall"`
	Timestamp time.Time       `json:"timestamp"`

	// IntegrityHealth reflects open CRITICAL/HIGH integrity violations.
	IntegrityHealth SubsystemHealth `json:"integrity_health"`

	// AuditDeliveryHealth reflects audit outbox delivery state.
	AuditDeliveryHealth SubsystemHealth `json:"audit_delivery_health"`

	// OutboxBacklogHealth reflects outbox backlog, stuck entries, and dead-letter state.
	OutboxBacklogHealth SubsystemHealth `json:"outbox_backlog_health"`

	// AuditChainHealth reflects hash-chain tamper-evidence integrity.
	AuditChainHealth SubsystemHealth `json:"audit_chain_health"`

	// AnomalyHealth reflects recent anomaly detection activity.
	AnomalyHealth SubsystemHealth `json:"anomaly_health"`

	// SafetyPolicyHealth reflects whether the safety policy engine is configured.
	SafetyPolicyHealth SubsystemHealth `json:"safety_policy_health"`
}

// FinanceGovernanceService aggregates subsystem health into a dashboard-ready report.
type FinanceGovernanceService struct {
	violationRepo  IntegrityViolationRepository // nil → skipped
	outboxGovernor *OutboxGovernor              // nil → skipped
	chainVerifier  *AuditChainVerifier          // nil → skipped
	gapDetector    *AuditGapDetector            // nil → skipped
	safetyEnforcer *SafetyEnforcer              // nil → policy not configured
}

// NewFinanceGovernanceService creates the governance service.
// All dependencies are optional: nil dependencies produce DEGRADED status for that subsystem.
func NewFinanceGovernanceService(
	violationRepo IntegrityViolationRepository,
	outboxGovernor *OutboxGovernor,
	chainVerifier *AuditChainVerifier,
	gapDetector *AuditGapDetector,
	safetyEnforcer *SafetyEnforcer,
) *FinanceGovernanceService {
	return &FinanceGovernanceService{
		violationRepo:  violationRepo,
		outboxGovernor: outboxGovernor,
		chainVerifier:  chainVerifier,
		gapDetector:    gapDetector,
		safetyEnforcer: safetyEnforcer,
	}
}

// GetHealthReport returns a comprehensive finance module health report for the current tenant.
// Non-critical query failures degrade individual subsystems without failing the whole report.
func (g *FinanceGovernanceService) GetHealthReport(ctx context.Context) *FinanceHealthReport {
	if g == nil {
		tenantID, _ := tenantIDOrNil(ctx)
		return &FinanceHealthReport{
			TenantID:            tenantID,
			Overall:             SubsystemStatusDegraded,
			Timestamp:           time.Now(),
			IntegrityHealth:     SubsystemHealth{Status: SubsystemStatusDegraded, Message: "governance service not configured"},
			AuditDeliveryHealth: SubsystemHealth{Status: SubsystemStatusDegraded, Message: "governance service not configured"},
			OutboxBacklogHealth: SubsystemHealth{Status: SubsystemStatusDegraded, Message: "governance service not configured"},
			AuditChainHealth:    SubsystemHealth{Status: SubsystemStatusDegraded, Message: "governance service not configured"},
			AnomalyHealth:       SubsystemHealth{Status: SubsystemStatusHealthy, Message: "metrics-only; check Prometheus"},
			SafetyPolicyHealth:  SubsystemHealth{Status: SubsystemStatusDegraded, Message: "safety enforcer not configured"},
		}
	}

	tenantID, _ := tenantIDOrNil(ctx)
	report := &FinanceHealthReport{
		TenantID:  tenantID,
		Timestamp: time.Now(),
	}

	report.IntegrityHealth = g.checkIntegrityHealth(ctx, tenantID)
	report.AuditDeliveryHealth = g.checkAuditDeliveryHealth(ctx)
	report.OutboxBacklogHealth = g.checkOutboxBacklogHealth(ctx)
	report.AuditChainHealth = g.checkChainHealth(ctx)
	report.AnomalyHealth = SubsystemHealth{Status: SubsystemStatusHealthy, Message: "anomaly detection active — see finance_anomaly_* metrics"}
	report.SafetyPolicyHealth = g.checkSafetyHealth()

	report.Overall = aggregateStatus(
		report.IntegrityHealth.Status,
		report.AuditDeliveryHealth.Status,
		report.OutboxBacklogHealth.Status,
		report.AuditChainHealth.Status,
		report.SafetyPolicyHealth.Status,
	)
	return report
}

func (g *FinanceGovernanceService) checkIntegrityHealth(ctx context.Context, tenantID uuid.UUID) SubsystemHealth {
	if g.violationRepo == nil {
		return SubsystemHealth{Status: SubsystemStatusDegraded, Message: "violation repository not configured — integrity blocking disabled"}
	}
	count, err := g.violationRepo.CountOpenCritical(ctx, tenantID)
	if err != nil {
		return SubsystemHealth{Status: SubsystemStatusDegraded, Message: "query failed: " + err.Error()}
	}
	if count > 0 {
		return SubsystemHealth{
			Status: SubsystemStatusCritical, Count: count,
			Message: "unresolved CRITICAL integrity violations — finance mutations are blocked",
		}
	}

	// Check for open HIGH violations.
	sev := SeverityHigh
	highs, err := g.violationRepo.ListOpenViolations(ctx, tenantID, &sev)
	if err == nil && len(highs) > 0 {
		return SubsystemHealth{
			Status: SubsystemStatusDegraded, Count: len(highs),
			Message: "unresolved HIGH integrity violations — review before period close",
		}
	}
	return SubsystemHealth{Status: SubsystemStatusHealthy}
}

func (g *FinanceGovernanceService) checkAuditDeliveryHealth(ctx context.Context) SubsystemHealth {
	if g.gapDetector == nil {
		return SubsystemHealth{Status: SubsystemStatusDegraded, Message: "gap detector not configured"}
	}
	gapReport, err := g.gapDetector.CheckGaps(ctx)
	if err != nil {
		return SubsystemHealth{Status: SubsystemStatusDegraded, Message: "gap check failed: " + err.Error()}
	}
	if gapReport.DeadOutboxCount > 0 {
		return SubsystemHealth{
			Status: SubsystemStatusCritical, Count: gapReport.DeadOutboxCount,
			Message: "DEAD outbox entries — CRITICAL audit events may be permanently lost",
		}
	}
	if gapReport.StaleOutboxCount > 0 {
		return SubsystemHealth{
			Status: SubsystemStatusDegraded, Count: gapReport.StaleOutboxCount,
			Message: "stale PENDING outbox entries — delivery worker is lagging",
		}
	}
	return SubsystemHealth{Status: SubsystemStatusHealthy}
}

func (g *FinanceGovernanceService) checkOutboxBacklogHealth(ctx context.Context) SubsystemHealth {
	if g.outboxGovernor == nil {
		return SubsystemHealth{Status: SubsystemStatusDegraded, Message: "outbox governor not configured"}
	}
	backlogReport, err := g.outboxGovernor.CheckBacklog(ctx)
	if err != nil {
		return SubsystemHealth{Status: SubsystemStatusDegraded, Message: "backlog check failed: " + err.Error()}
	}
	if backlogReport.BacklogCritical {
		return SubsystemHealth{
			Status: SubsystemStatusCritical, Count: backlogReport.PendingBacklogCount,
			Message: "critical backlog — delivery worker is severely behind",
		}
	}
	if backlogReport.StuckProcessingCount > 0 {
		return SubsystemHealth{
			Status: SubsystemStatusDegraded, Count: backlogReport.StuckProcessingCount,
			Message: "stuck PROCESSING entries — processor may have crashed",
		}
	}
	return SubsystemHealth{Status: SubsystemStatusHealthy}
}

func (g *FinanceGovernanceService) checkChainHealth(ctx context.Context) SubsystemHealth {
	if g.chainVerifier == nil {
		return SubsystemHealth{Status: SubsystemStatusDegraded, Message: "audit chain verifier not configured — tamper detection disabled"}
	}
	// Full chain verification is too expensive for the health endpoint.
	// Return HEALTHY here and run VerifyChain from the Temporal cron separately.
	return SubsystemHealth{Status: SubsystemStatusHealthy, Message: "chain verification runs on scheduled cron — see audit_chain_violations_total metric"}
}

func (g *FinanceGovernanceService) checkSafetyHealth() SubsystemHealth {
	if g.safetyEnforcer == nil {
		return SubsystemHealth{Status: SubsystemStatusDegraded, Message: "safety enforcer not configured — runtime policy checks disabled"}
	}
	return SubsystemHealth{Status: SubsystemStatusHealthy, Message: "safety enforcer active — see finance_safety_violations_total metric"}
}

// aggregateStatus returns the worst status across a set of subsystem statuses.
func aggregateStatus(statuses ...SubsystemStatus) SubsystemStatus {
	worst := SubsystemStatusHealthy
	for _, s := range statuses {
		if s == SubsystemStatusCritical {
			return SubsystemStatusCritical
		}
		if s == SubsystemStatusDegraded {
			worst = SubsystemStatusDegraded
		}
	}
	return worst
}

// tenantIDOrNil is a nil-safe wrapper for the context tenant ID.
func tenantIDOrNil(ctx context.Context) (uuid.UUID, bool) {
	// Reuse the existing tenantIDFromCtx helper.
	// Inline here to avoid returning an error in a non-error context.
	id, err := tenantIDFromCtx(ctx)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}
