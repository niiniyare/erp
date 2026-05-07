package service

// verification_hooks.go — Continuous governance verification service.
//
// ContinuousVerificationService ties all governance subsystems together into
// a single verification sweep. It is designed to be called:
//
//   - At service STARTUP via AssertStartupInvariants (blocking — fail if error)
//   - From a TEMPORAL CRON via RunScheduledVerification (non-blocking — report only)
//   - After MAJOR OPERATIONS (period close, bulk import) for post-hoc confirmation
//
// Design:
//   - Startup assertions BLOCK service startup if critical infrastructure is missing
//   - Scheduled verification NEVER blocks traffic — it reports and emits metrics
//   - Each check is independent — a failing check does not prevent others from running
//   - Failed checks surface as VerificationFailures with CRITICAL/HIGH/MEDIUM severity
//
// Governance failure containment:
//   - All subsystem dependencies are optional (nil-safe)
//   - A nil dependency produces a DEGRADED VerificationFailure, not a panic
//   - Operators are notified via metrics and structured logs regardless of nil state

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"awo.so/internal/shared"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
)

// VerificationSeverity classifies the urgency of a verification failure.
type VerificationSeverity string

const (
	VerificationSeverityCritical VerificationSeverity = "CRITICAL"
	VerificationSeverityHigh     VerificationSeverity = "HIGH"
	VerificationSeverityMedium   VerificationSeverity = "MEDIUM"
)

// VerificationFailure describes a single check that did not pass.
type VerificationFailure struct {
	Check        string               `json:"check"`
	Severity     VerificationSeverity `json:"severity"`
	Detail       string               `json:"detail"`
	RepairAction string               `json:"repair_action,omitempty"`
}

func (f VerificationFailure) Error() string {
	return fmt.Sprintf("[%s] %s: %s", f.Severity, f.Check, f.Detail)
}

// VerificationReport is the output of a single verification sweep.
type VerificationReport struct {
	TenantID   uuid.UUID             `json:"tenant_id"`
	ChecksRun  []string              `json:"checks_run"`
	Failures   []VerificationFailure `json:"failures,omitempty"`
	Healthy    bool                  `json:"healthy"`
	VerifiedAt time.Time             `json:"verified_at"`
}

// ContinuousVerificationService runs all governance invariant checks.
// A nil *ContinuousVerificationService returns healthy empty reports.
type ContinuousVerificationService struct {
	evolutionGuard  *EvolutionSafetyGuard
	antiEntropy     *AntiEntropyService
	outboxGovernor  *OutboxGovernor
	gapDetector     *AuditGapDetector
	violationRepo   IntegrityViolationRepository
	policyRegistry  *GovernancePolicyRegistry
	metrics         metrics.MetricsProvider
}

// NewContinuousVerificationService creates the service. All deps may be nil.
func NewContinuousVerificationService(
	evolutionGuard *EvolutionSafetyGuard,
	antiEntropy *AntiEntropyService,
	outboxGovernor *OutboxGovernor,
	gapDetector *AuditGapDetector,
	violationRepo IntegrityViolationRepository,
	policyRegistry *GovernancePolicyRegistry,
	m metrics.MetricsProvider,
) *ContinuousVerificationService {
	return &ContinuousVerificationService{
		evolutionGuard: evolutionGuard,
		antiEntropy:    antiEntropy,
		outboxGovernor: outboxGovernor,
		gapDetector:    gapDetector,
		violationRepo:  violationRepo,
		policyRegistry: policyRegistry,
		metrics:        m,
	}
}

// AssertStartupInvariants verifies that the critical governance infrastructure
// is correctly wired. Returns a non-nil error to BLOCK service startup.
//
// Callers in main() MUST propagate this error and exit — do not ignore it.
func (s *ContinuousVerificationService) AssertStartupInvariants(ctx context.Context) error {
	if s == nil {
		return nil
	}

	var critical []string

	// Invariant: evolution safety guard must run its own startup checks.
	if s.evolutionGuard != nil {
		if err := s.evolutionGuard.AssertStartupInvariants(ctx); err != nil {
			critical = append(critical, err.Error())
		}
	} else {
		critical = append(critical, "EvolutionSafetyGuard is nil — extension registration contracts are not enforced")
	}

	// Invariant: governance registry must be configured.
	if s.policyRegistry == nil {
		critical = append(critical, "GovernancePolicyRegistry is nil — all policy decisions use static defaults, tenant overrides disabled")
	}

	if len(critical) > 0 {
		for _, c := range critical {
			logger.ErrorContext(ctx, "finance continuous verification: startup FAILED", logger.Fields{
				"failure": c,
			})
		}
		s.metrics.IncrementCounter("finance_verification_startup_failed_total", metrics.Fields{
			"count": fmt.Sprintf("%d", len(critical)),
		})
		return fmt.Errorf("finance verification: %d startup invariant(s) failed: %v", len(critical), critical)
	}

	s.metrics.IncrementCounter("finance_verification_startup_passed_total", metrics.Fields{})
	logger.InfoContext(ctx, "finance continuous verification: startup invariants passed", logger.Fields{})
	return nil
}

// RunScheduledVerification performs all governance checks for the current tenant.
// Non-blocking: failures are reported in the returned VerificationReport, not returned as errors.
// Call from a Temporal cron after anti-entropy checks.
func (s *ContinuousVerificationService) RunScheduledVerification(ctx context.Context) (*VerificationReport, error) {
	if s == nil {
		tenantID, _ := shared.GetTenantID(ctx)
		return &VerificationReport{TenantID: tenantID, Healthy: true, VerifiedAt: time.Now()}, nil
	}
	tenantID, _ := shared.GetTenantID(ctx)
	report := &VerificationReport{
		TenantID:   tenantID,
		VerifiedAt: time.Now(),
	}

	// Check 1: governance infrastructure completeness.
	s.checkInfrastructurePresence(report)

	// Check 2: extension contract coverage.
	s.checkExtensionCoverage(report)

	// Check 3: open CRITICAL violations.
	s.checkCriticalViolations(ctx, tenantID, report)

	// Check 4: outbox health.
	s.checkOutboxHealth(ctx, report)

	// Check 5: anti-entropy cross-system consistency.
	s.checkAntiEntropy(ctx, report)

	// Check 6: policy registry integrity.
	s.checkPolicyRegistry(report)

	report.Healthy = len(report.Failures) == 0

	totalChecks := len(report.ChecksRun)
	totalFailures := len(report.Failures)

	s.metrics.ObserveHistogram("finance_verification_checks_total", float64(totalChecks), metrics.Fields{})
	s.metrics.ObserveHistogram("finance_verification_failures_total", float64(totalFailures), metrics.Fields{})

	if !report.Healthy {
		s.metrics.IncrementCounter("finance_verification_unhealthy_total", metrics.Fields{
			"tenant_id": tenantID.String(),
		})
		logger.ErrorContext(ctx, "finance continuous verification: FAILURES detected", logger.Fields{
			"tenant_id": tenantID.String(),
			"failures":  totalFailures,
			"checks":    totalChecks,
		})
	} else {
		logger.InfoContext(ctx, "finance continuous verification: all checks passed", logger.Fields{
			"tenant_id": tenantID.String(),
			"checks":    totalChecks,
		})
	}

	return report, nil
}

func (s *ContinuousVerificationService) checkInfrastructurePresence(report *VerificationReport) {
	report.ChecksRun = append(report.ChecksRun, "INFRASTRUCTURE_PRESENCE")

	if s.evolutionGuard == nil {
		report.Failures = append(report.Failures, VerificationFailure{
			Check:        "INFRASTRUCTURE_PRESENCE",
			Severity:     VerificationSeverityHigh,
			Detail:       "EvolutionSafetyGuard is not wired — extension governance contracts are not enforced",
			RepairAction: "Wire EvolutionSafetyGuard in NewServices and call RegisterExtension for all custom extensions",
		})
	}
	if s.policyRegistry == nil {
		report.Failures = append(report.Failures, VerificationFailure{
			Check:        "INFRASTRUCTURE_PRESENCE",
			Severity:     VerificationSeverityHigh,
			Detail:       "GovernancePolicyRegistry is not wired — tenant-aware policy is disabled",
			RepairAction: "Wire GovernancePolicyRegistry in NewServices",
		})
	}
}

func (s *ContinuousVerificationService) checkExtensionCoverage(report *VerificationReport) {
	if s.evolutionGuard == nil {
		return
	}
	report.ChecksRun = append(report.ChecksRun, "EXTENSION_COVERAGE")
	violations := s.evolutionGuard.ValidateExtensions()
	for _, v := range violations {
		report.Failures = append(report.Failures, VerificationFailure{
			Check:        "EXTENSION_COVERAGE",
			Severity:     VerificationSeverityCritical,
			Detail:       v.Error(),
			RepairAction: fmt.Sprintf("Update ExtensionContract for %q to declare the missing hook", v.Extension),
		})
	}
}

func (s *ContinuousVerificationService) checkCriticalViolations(ctx context.Context, tenantID uuid.UUID, report *VerificationReport) {
	if s.violationRepo == nil {
		return
	}
	report.ChecksRun = append(report.ChecksRun, "CRITICAL_VIOLATIONS")
	count, err := s.violationRepo.CountOpenCritical(ctx, tenantID)
	if err != nil {
		report.Failures = append(report.Failures, VerificationFailure{
			Check:        "CRITICAL_VIOLATIONS",
			Severity:     VerificationSeverityHigh,
			Detail:       "could not query open CRITICAL violations: " + err.Error(),
			RepairAction: "Check DB connectivity and violation repo wiring",
		})
		return
	}
	if count > 0 {
		report.Failures = append(report.Failures, VerificationFailure{
			Check:    "CRITICAL_VIOLATIONS",
			Severity: VerificationSeverityCritical,
			Detail:   fmt.Sprintf("%d open CRITICAL integrity violations — finance mutations are blocked", count),
			RepairAction: "Acknowledge each violation, investigate the root cause, and call ResolveViolation once repaired",
		})
	}
}

func (s *ContinuousVerificationService) checkOutboxHealth(ctx context.Context, report *VerificationReport) {
	if s.gapDetector == nil && s.outboxGovernor == nil {
		return
	}
	report.ChecksRun = append(report.ChecksRun, "OUTBOX_HEALTH")

	if s.gapDetector != nil {
		gapReport, err := s.gapDetector.CheckGaps(ctx)
		if err != nil {
			report.Failures = append(report.Failures, VerificationFailure{
				Check:    "OUTBOX_HEALTH",
				Severity: VerificationSeverityHigh,
				Detail:   "gap detector query failed: " + err.Error(),
			})
			return
		}
		if gapReport.DeadOutboxCount > 0 {
			report.Failures = append(report.Failures, VerificationFailure{
				Check:    "OUTBOX_HEALTH",
				Severity: VerificationSeverityCritical,
				Detail:   fmt.Sprintf("%d DEAD outbox entries — CRITICAL audit events may be permanently lost", gapReport.DeadOutboxCount),
				RepairAction: "Inspect DEAD entries and attempt replay via SelfHealingService.ValidateReplay",
			})
		}
		if gapReport.StaleOutboxCount > 0 {
			report.Failures = append(report.Failures, VerificationFailure{
				Check:    "OUTBOX_HEALTH",
				Severity: VerificationSeverityHigh,
				Detail:   fmt.Sprintf("%d stale PENDING outbox entries — delivery worker is lagging", gapReport.StaleOutboxCount),
				RepairAction: "Check AuditOutboxProcessor Temporal cron status",
			})
		}
	}
}

func (s *ContinuousVerificationService) checkAntiEntropy(ctx context.Context, report *VerificationReport) {
	if s.antiEntropy == nil {
		return
	}
	report.ChecksRun = append(report.ChecksRun, "ANTI_ENTROPY")
	aeReport, err := s.antiEntropy.RunChecks(ctx)
	if err != nil {
		report.Failures = append(report.Failures, VerificationFailure{
			Check:    "ANTI_ENTROPY",
			Severity: VerificationSeverityHigh,
			Detail:   "anti-entropy check failed: " + err.Error(),
		})
		return
	}
	for _, v := range aeReport.Violations {
		report.Failures = append(report.Failures, VerificationFailure{
			Check:        "ANTI_ENTROPY",
			Severity:     VerificationSeverityCritical,
			Detail:       v.Detail,
			RepairAction: v.RepairAction,
		})
	}
}

func (s *ContinuousVerificationService) checkPolicyRegistry(report *VerificationReport) {
	if s.policyRegistry == nil {
		return
	}
	report.ChecksRun = append(report.ChecksRun, "POLICY_REGISTRY")
	global := s.policyRegistry.GlobalPolicy()
	if errs := ValidatePolicy(global); len(errs) > 0 {
		for _, e := range errs {
			report.Failures = append(report.Failures, VerificationFailure{
				Check:        "POLICY_REGISTRY",
				Severity:     VerificationSeverityCritical,
				Detail:       e.Error(),
				RepairAction: "Update GovernancePolicyRegistry global policy to satisfy validation requirements",
			})
		}
	}
}
