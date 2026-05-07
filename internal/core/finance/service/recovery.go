package service

// recovery.go — Post-disaster-recovery diagnostics for the finance ledger.
//
// # Purpose
//
// After a database restore, Temporal server restart, or infrastructure failure,
// operators need to VERIFY the ledger before returning traffic. This service
// provides the structured checks required for that verification.
//
// # Recovery procedure
//
//  1. Run PostRestoreCheck; if HasCriticalIssues == true → hold writes
//  2. Review each violation's RepairAction field
//  3. Resolve manually (reversal, compensating entry, workflow resubmit)
//  4. Re-run PostRestoreCheck to confirm clean
//  5. Return traffic
//
// All methods are READ-ONLY. They detect problems; they never auto-repair.

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"awo.so/internal/core/finance/domain"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
)

// RecoveryService provides post-disaster-recovery diagnostics for the finance ledger.
type RecoveryService interface {
	// PostRestoreCheck runs all integrity scans and orphan detection.
	// Must return HasCriticalIssues == false before returning traffic after a restore.
	PostRestoreCheck(ctx context.Context, tenantID uuid.UUID) (*RecoveryReport, error)

	// FindOrphanedWorkflows returns transactions stuck in PENDING_APPROVAL
	// with no active workflow record older than stuckAfter.
	// These require manual resubmission or cancellation.
	FindOrphanedWorkflows(ctx context.Context, tenantID uuid.UUID, stuckAfter time.Duration) ([]*OrphanedWorkflow, error)
}

// RecoveryReport is the combined output of PostRestoreCheck.
type RecoveryReport struct {
	TenantID          uuid.UUID          `json:"tenant_id"`
	CheckedAt         time.Time          `json:"checked_at"`
	PostedScanReport  *IntegrityReport   `json:"posted_scan"`
	ReversalReport    *IntegrityReport   `json:"reversal_scan"`
	DuplicateReport   *IntegrityReport   `json:"duplicate_scan"`
	OrphanedWorkflows []*OrphanedWorkflow `json:"orphaned_workflows"`
	HasCriticalIssues bool               `json:"has_critical_issues"`
	Summary           string             `json:"summary"`
}

// OrphanedWorkflow describes a transaction stuck in PENDING_APPROVAL
// with no corresponding active workflow record.
type OrphanedWorkflow struct {
	TransactionID     uuid.UUID               `json:"transaction_id"`
	TransactionNumber string                  `json:"transaction_number"`
	Status            domain.TransactionStatus `json:"status"`
	PendingSince      time.Time               `json:"pending_since"`
	HasWorkflowRecord bool                    `json:"has_workflow_record"`
	RecommendedAction string                  `json:"recommended_action"`
}

type recoveryService struct {
	integrityService IntegrityService
	txnRepo          domain.TransactionRepository
	approvalRepo     domain.ApprovalWorkflowRepository // nil → workflow check skipped
	metrics          metrics.MetricsProvider
}

// NewRecoveryService constructs a RecoveryService.
// approvalRepo may be nil — orphaned workflow detection will be skipped.
func NewRecoveryService(
	integrityService IntegrityService,
	txnRepo domain.TransactionRepository,
	approvalRepo domain.ApprovalWorkflowRepository,
	m metrics.MetricsProvider,
) RecoveryService {
	return &recoveryService{
		integrityService: integrityService,
		txnRepo:          txnRepo,
		approvalRepo:     approvalRepo,
		metrics:          m,
	}
}

// PostRestoreCheck runs all three integrity scans and orphaned-workflow detection.
// Returns a RecoveryReport. HasCriticalIssues == true means writes must not resume.
func (s *recoveryService) PostRestoreCheck(ctx context.Context, tenantID uuid.UUID) (*RecoveryReport, error) {
	report := &RecoveryReport{
		TenantID:  tenantID,
		CheckedAt: time.Now(),
	}

	// 1. Balance + no-entries scan across all posted transactions.
	postedReport, err := s.integrityService.ScanPostedTransactions(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("post-restore check: posted scan: %w", err)
	}
	report.PostedScanReport = postedReport

	// 2. Reversal chain integrity.
	reversalReport, err := s.integrityService.ScanReversalChains(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("post-restore check: reversal scan: %w", err)
	}
	report.ReversalReport = reversalReport

	// 3. Duplicate postings (same transaction number posted multiple times).
	dupReport, err := s.integrityService.ScanDuplicatePostings(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("post-restore check: duplicate scan: %w", err)
	}
	report.DuplicateReport = dupReport

	// 4. Orphaned workflow detection (non-fatal — approval repo may not be wired).
	orphans, orphanErr := s.FindOrphanedWorkflows(ctx, tenantID, time.Hour)
	if orphanErr != nil {
		logger.WarnContext(ctx, "post-restore check: orphaned workflow scan skipped",
			logger.Fields{"error": orphanErr.Error()})
	} else {
		report.OrphanedWorkflows = orphans
	}

	// Aggregate critical status.
	report.HasCriticalIssues = postedReport.HasCritical() || reversalReport.HasCritical() || dupReport.HasCritical()

	totalViolations := len(postedReport.Violations) + len(reversalReport.Violations) + len(dupReport.Violations)
	criticalCount := countCriticalViolations(postedReport, reversalReport, dupReport)

	switch {
	case report.HasCriticalIssues:
		report.Summary = fmt.Sprintf(
			"POST-RESTORE CHECK FAILED: %d total violations (%d critical). "+
				"DO NOT RETURN TRAFFIC until all CRITICAL violations are resolved.",
			totalViolations, criticalCount,
		)
		logger.ErrorContext(ctx, report.Summary, logger.Fields{
			"tenant_id":          tenantID.String(),
			"total_violations":   totalViolations,
			"critical_count":     criticalCount,
			"orphaned_workflows": len(report.OrphanedWorkflows),
		})
		s.metrics.IncrementCounter("recovery_check_failed_total", metrics.Fields{"reason": "critical_violations"})

	case totalViolations > 0:
		report.Summary = fmt.Sprintf(
			"Post-restore check: %d non-critical violations. Review before period close. Traffic may resume.",
			totalViolations,
		)
		logger.WarnContext(ctx, report.Summary, logger.Fields{
			"tenant_id":        tenantID.String(),
			"total_violations": totalViolations,
		})
		s.metrics.IncrementCounter("recovery_check_warnings_total", metrics.Fields{})

	default:
		scanned := postedReport.ScannedTransactions + reversalReport.ScannedTransactions
		report.Summary = fmt.Sprintf(
			"Post-restore check PASSED. Ledger clean. Scanned %d transactions. Traffic may resume.",
			scanned,
		)
		logger.InfoContext(ctx, report.Summary, logger.Fields{"tenant_id": tenantID.String()})
		s.metrics.IncrementCounter("recovery_check_passed_total", metrics.Fields{})
	}

	return report, nil
}

// FindOrphanedWorkflows finds transactions in PENDING_APPROVAL with no workflow
// record that are older than stuckAfter. Paginates to avoid OOM.
func (s *recoveryService) FindOrphanedWorkflows(ctx context.Context, tenantID uuid.UUID, stuckAfter time.Duration) ([]*OrphanedWorkflow, error) {
	if s.approvalRepo == nil {
		return nil, nil
	}

	pendingStatus := domain.TransactionStatusPendingApproval
	pageSize := domain.MaxIntegrityScanPage
	cutoff := time.Now().Add(-stuckAfter)

	var orphans []*OrphanedWorkflow
	offset := 0

	for {
		lim := pageSize
		off := offset
		page, err := s.txnRepo.List(ctx, &domain.TransactionFilter{
			Status: &pendingStatus,
			Limit:  &lim,
			Offset: &off,
		})
		if err != nil {
			return nil, fmt.Errorf("orphan scan: list pending transactions: %w", err)
		}
		if len(page) == 0 {
			break
		}

		for _, txn := range page {
			// Skip recently-submitted transactions — not yet stuck.
			if txn.CreatedAt.After(cutoff) {
				continue
			}

			_, wfErr := s.approvalRepo.GetWorkflowByTransaction(ctx, txn.ID)
			hasRecord := wfErr == nil

			if !hasRecord {
				orphans = append(orphans, &OrphanedWorkflow{
					TransactionID:     txn.ID,
					TransactionNumber: txn.TransactionNumber,
					Status:            txn.TransactionStatus,
					PendingSince:      txn.CreatedAt,
					HasWorkflowRecord: false,
					RecommendedAction: "Resubmit via SubmitForApproval to create a workflow record, or cancel if the transaction is stale.",
				})
			}
		}

		if len(page) < pageSize {
			break
		}
		offset += pageSize
	}

	if len(orphans) > 0 {
		s.metrics.IncrementCounter("orphaned_workflows_detected_total", metrics.Fields{
			"count": fmt.Sprintf("%d", len(orphans)),
		})
		logger.WarnContext(ctx, "Orphaned workflows detected",
			logger.Fields{
				"tenant_id": tenantID.String(),
				"count":     len(orphans),
			})
	}

	return orphans, nil
}

// countCriticalViolations counts CRITICAL violations across multiple reports.
func countCriticalViolations(reports ...*IntegrityReport) int {
	n := 0
	for _, r := range reports {
		for _, v := range r.Violations {
			if v.Severity == SeverityCritical {
				n++
			}
		}
	}
	return n
}
