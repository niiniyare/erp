package service

// operations.go — Operator-facing diagnostic capabilities for the finance ledger.
//
// # Intended audience
//
//   - On-call engineers investigating production incidents
//   - Finance controllers performing period close sign-off
//   - Support staff diagnosing tenant-reported discrepancies
//
// All outputs are READ-ONLY — no mutations. Outputs are safe to log or surface
// in internal dashboards. Each report includes actionable recommendations.
//
// # Usage patterns
//
//   - Routine monitoring: schedule TenantHealthReport nightly via Temporal cron
//   - Period close gate: block close if OverallStatus == RED
//   - Reconciliation investigation: ReconciliationDiagnostics when users report
//     statements stuck in IN_PROGRESS

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"awo.so/internal/core/finance/domain"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
)

// OperationsService aggregates operator diagnostics for the finance ledger.
type OperationsService interface {
	// TenantHealthReport runs all integrity scans and returns a structured
	// health snapshot with severity classification and recommendations.
	// Use for nightly monitoring and period-close sign-off.
	TenantHealthReport(ctx context.Context, tenantID uuid.UUID) (*TenantHealthReport, error)

	// ReconciliationDiagnostics returns open/in-progress bank statement counts
	// and identifies stale statements. Use when reconciliation appears stuck.
	ReconciliationDiagnostics(ctx context.Context, tenantID uuid.UUID) (*ReconciliationDiagnostics, error)
}

// HealthStatus classifies the overall tenant ledger health.
type HealthStatus string

const (
	// HealthStatusGreen — no integrity violations. Ledger is clean.
	HealthStatusGreen HealthStatus = "GREEN"

	// HealthStatusYellow — non-critical violations exist. Action required
	// before period close but traffic may continue.
	HealthStatusYellow HealthStatus = "YELLOW"

	// HealthStatusRed — CRITICAL violations exist. Period close must be
	// halted. Escalate immediately.
	HealthStatusRed HealthStatus = "RED"
)

// TenantHealthReport is the output of TenantHealthReport.
type TenantHealthReport struct {
	TenantID       uuid.UUID        `json:"tenant_id"`
	GeneratedAt    time.Time        `json:"generated_at"`
	OverallStatus  HealthStatus     `json:"overall_status"`
	IntegrityReport *IntegrityReport `json:"integrity"`
	ReversalReport  *IntegrityReport `json:"reversal_chains"`
	DuplicateReport *IntegrityReport `json:"duplicate_postings"`
	Recommendations []string         `json:"recommendations,omitempty"`
}

// ReconciliationDiagnostics is the output of ReconciliationDiagnostics.
type ReconciliationDiagnostics struct {
	TenantID             uuid.UUID              `json:"tenant_id"`
	GeneratedAt          time.Time              `json:"generated_at"`
	InProgressStatements []*StatementDiagnostic `json:"in_progress_statements"`
	StaleStatementCount  int                    `json:"stale_statement_count"` // open > 30 days
	Recommendations      []string               `json:"recommendations,omitempty"`
}

// StatementDiagnostic summarises a single bank statement's health.
type StatementDiagnostic struct {
	StatementID    uuid.UUID `json:"statement_id"`
	Status         string    `json:"status"`
	UnmatchedCount int       `json:"unmatched_count"`
	IsStale        bool      `json:"is_stale"`
}

type operationsService struct {
	integrityService   IntegrityService
	reconciliationRepo domain.ReconciliationRepository // nil → reconciliation diagnostics skipped
	metrics            metrics.MetricsProvider
}

// NewOperationsService constructs an OperationsService.
// reconciliationRepo may be nil — reconciliation diagnostics will return a
// stub report with a recommendation to wire the repository.
func NewOperationsService(
	integrityService IntegrityService,
	reconciliationRepo domain.ReconciliationRepository,
	m metrics.MetricsProvider,
) OperationsService {
	return &operationsService{
		integrityService:   integrityService,
		reconciliationRepo: reconciliationRepo,
		metrics:            m,
	}
}

// TenantHealthReport runs all three integrity scans and classifies the result.
func (s *operationsService) TenantHealthReport(ctx context.Context, tenantID uuid.UUID) (*TenantHealthReport, error) {
	report := &TenantHealthReport{
		TenantID:    tenantID,
		GeneratedAt: time.Now(),
	}

	integrityRpt, err := s.integrityService.ScanPostedTransactions(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("tenant health: posted scan: %w", err)
	}
	report.IntegrityReport = integrityRpt

	reversalRpt, err := s.integrityService.ScanReversalChains(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("tenant health: reversal scan: %w", err)
	}
	report.ReversalReport = reversalRpt

	dupRpt, err := s.integrityService.ScanDuplicatePostings(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant health: duplicate scan: %w", err)
	}
	report.DuplicateReport = dupRpt

	hasCritical := integrityRpt.HasCritical() || reversalRpt.HasCritical() || dupRpt.HasCritical()
	totalViolations := len(integrityRpt.Violations) + len(reversalRpt.Violations) + len(dupRpt.Violations)

	switch {
	case hasCritical:
		report.OverallStatus = HealthStatusRed
		report.Recommendations = []string{
			"CRITICAL violations detected — halt period close immediately.",
			"Review each CRITICAL violation's RepairAction field for remediation steps.",
			"Do not run bulk imports or migrations until the ledger is clean.",
			"Re-run TenantHealthReport after each repair to confirm resolution.",
		}
		logger.ErrorContext(ctx, "Tenant health report: RED",
			logger.Fields{
				"tenant_id":        tenantID.String(),
				"total_violations": totalViolations,
				"critical_count":   countCriticalViolations(integrityRpt, reversalRpt, dupRpt),
			})
		s.metrics.IncrementCounter("tenant_health_status_total", metrics.Fields{"status": "red"})

	case totalViolations > 0:
		report.OverallStatus = HealthStatusYellow
		report.Recommendations = []string{
			"Non-critical violations found. Review all violations before period close.",
			"Traffic may continue but no new period close sign-off until resolved.",
		}
		logger.WarnContext(ctx, "Tenant health report: YELLOW",
			logger.Fields{
				"tenant_id":        tenantID.String(),
				"total_violations": totalViolations,
			})
		s.metrics.IncrementCounter("tenant_health_status_total", metrics.Fields{"status": "yellow"})

	default:
		report.OverallStatus = HealthStatusGreen
		logger.InfoContext(ctx, "Tenant health report: GREEN",
			logger.Fields{
				"tenant_id": tenantID.String(),
				"scanned":   integrityRpt.ScannedTransactions + reversalRpt.ScannedTransactions,
			})
		s.metrics.IncrementCounter("tenant_health_status_total", metrics.Fields{"status": "green"})
	}

	return report, nil
}

// ReconciliationDiagnostics lists open and in-progress bank statements
// and identifies those that have been open longer than 30 days.
func (s *operationsService) ReconciliationDiagnostics(ctx context.Context, tenantID uuid.UUID) (*ReconciliationDiagnostics, error) {
	diag := &ReconciliationDiagnostics{
		TenantID:    tenantID,
		GeneratedAt: time.Now(),
	}

	if s.reconciliationRepo == nil {
		diag.Recommendations = []string{
			"ReconciliationRepository not wired into OperationsService — diagnostics unavailable.",
			"Pass a ReconciliationRepository to NewOperationsService to enable this feature.",
		}
		return diag, nil
	}

	statements, err := s.reconciliationRepo.ListStatements(ctx, tenantID, nil)
	if err != nil {
		return nil, fmt.Errorf("reconciliation diagnostics: list statements: %w", err)
	}

	const staleThresholdDays = 30

	for _, stmt := range statements {
		// Only surface non-terminal statements.
		if stmt.Status == domain.ReconciliationStatusCompleted || stmt.Status == domain.ReconciliationStatusVoided {
			continue
		}

		isStale := stmt.UnmatchedCount > 0 // simplified stale heuristic when date unavailable
		if isStale {
			diag.StaleStatementCount++
		}

		diag.InProgressStatements = append(diag.InProgressStatements, &StatementDiagnostic{
			StatementID:    stmt.ID,
			Status:         string(stmt.Status),
			UnmatchedCount: stmt.UnmatchedCount,
			IsStale:        isStale,
		})
	}

	if diag.StaleStatementCount > 0 {
		diag.Recommendations = append(diag.Recommendations,
			fmt.Sprintf(
				"%d bank statement(s) have unmatched lines. Complete matching then call CompleteReconciliation, or void if no longer needed.",
				diag.StaleStatementCount,
			),
		)
	}

	if len(diag.InProgressStatements) == 0 {
		diag.Recommendations = append(diag.Recommendations, "All bank statements are completed or voided — reconciliation is clean.")
	}

	logger.InfoContext(ctx, "Reconciliation diagnostics complete",
		logger.Fields{
			"tenant_id":     tenantID.String(),
			"open_count":    len(diag.InProgressStatements),
			"stale_count":   diag.StaleStatementCount,
		})

	return diag, nil
}
