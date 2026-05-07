package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"awo.so/internal/core/audit"
	"awo.so/internal/core/finance/domain"
	"awo.so/internal/shared/errors"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

// ReconciliationService manages bank statement import and reconciliation.
type ReconciliationService interface {
	// ImportStatement creates a new bank statement with its lines.
	ImportStatement(ctx context.Context, s *domain.BankStatement, lines []*domain.BankStatementLine) (*domain.BankStatement, error)
	// GetStatement returns a statement with its lines.
	GetStatement(ctx context.Context, id uuid.UUID) (*domain.BankStatement, error)
	// ListStatements lists statements, optionally filtered by GL account.
	ListStatements(ctx context.Context, accountID *uuid.UUID) ([]*domain.BankStatement, error)
	// MatchLine matches a statement line to a journal entry.
	MatchLine(ctx context.Context, statementID, lineID, entryID uuid.UUID, byUserID uuid.UUID) (*domain.BankStatement, error)
	// UnmatchLine removes a match between a statement line and a journal entry.
	UnmatchLine(ctx context.Context, statementID, lineID uuid.UUID) (*domain.BankStatement, error)
	// ListLines returns lines for a statement.
	ListLines(ctx context.Context, statementID uuid.UUID, unmatchedOnly bool) ([]*domain.BankStatementLine, error)
	// CompleteReconciliation finalises the reconciliation and marks matched entries reconciled.
	CompleteReconciliation(ctx context.Context, statementID uuid.UUID, byUserID uuid.UUID) (*domain.BankStatement, error)
}

type reconciliationService struct {
	repo        domain.ReconciliationRepository
	tracing     tracing.Service
	metrics     metrics.MetricsProvider
	auditWriter *financeAuditWriter // nil → audit skipped
}

// NewReconciliationService creates a new ReconciliationService.
func NewReconciliationService(
	repo domain.ReconciliationRepository,
	tracing tracing.Service,
	metrics metrics.MetricsProvider,
	aw *financeAuditWriter,
) ReconciliationService {
	return &reconciliationService{repo: repo, tracing: tracing, metrics: metrics, auditWriter: aw}
}

func (s *reconciliationService) ImportStatement(ctx context.Context, stmt *domain.BankStatement, lines []*domain.BankStatementLine) (*domain.BankStatement, error) {
	ctx, span := s.tracing.StartSpan(ctx, "reconciliation_service.import_statement")
	defer span.End()

	tenantID, err := tenantIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	stmt.TenantID = tenantID
	stmt.Status = domain.ReconciliationStatusDraft
	stmt.UnmatchedCount = len(lines)

	if errs := stmt.Validate(); len(errs) > 0 {
		return nil, errors.NewBusinessError("STATEMENT_VALIDATION_FAILED", errs[0].Message).
			WithHTTPStatus(http.StatusBadRequest).
			WithCategory(errors.CategoryValidation)
	}

	if err := s.repo.CreateStatement(ctx, stmt); err != nil {
		span.RecordError(err)
		return nil, mapReconError(err, "import statement")
	}

	if len(lines) > 0 {
		for _, li := range lines {
			li.StatementID = stmt.ID
			li.TenantID = tenantID
		}
		if err := s.repo.CreateLines(ctx, lines); err != nil {
			span.RecordError(err)
			return nil, mapReconError(err, "create statement lines")
		}
		stmt.Lines = lines
	}

	s.metrics.IncrementCounter("bank_statement_imported_total", metrics.Fields{})
	s.auditWriter.writeAudit(ctx, audit.CreateAuditEventRequest{
		UserID:        auditUserID(ctx),
		EventType:     auditTypeStmtImported,
		EventCategory: auditCategoryFinance,
		Severity:      auditSeverityInfo,
		ResourceID:    uuidPtr(stmt.ID),
		Context: auditCtx(map[string]any{
			"statement_id": stmt.ID.String(),
			"lines_count":  len(lines),
			"account_id":   stmt.AccountID.String(),
		}),
	}, auditTypeStmtImported+":"+stmt.ID.String())
	return stmt, nil
}

func (s *reconciliationService) GetStatement(ctx context.Context, id uuid.UUID) (*domain.BankStatement, error) {
	ctx, span := s.tracing.StartSpan(ctx, "reconciliation_service.get_statement")
	defer span.End()

	stmt, err := s.repo.GetStatementByID(ctx, id)
	if err != nil {
		return nil, mapReconError(err, "get statement")
	}

	lines, err := s.repo.ListLines(ctx, id, false)
	if err != nil {
		return nil, mapReconError(err, "get statement lines")
	}
	stmt.Lines = lines
	return stmt, nil
}

func (s *reconciliationService) ListStatements(ctx context.Context, accountID *uuid.UUID) ([]*domain.BankStatement, error) {
	ctx, span := s.tracing.StartSpan(ctx, "reconciliation_service.list_statements")
	defer span.End()

	tenantID, err := tenantIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	list, err := s.repo.ListStatements(ctx, tenantID, accountID)
	if err != nil {
		return nil, mapReconError(err, "list statements")
	}
	return list, nil
}

func (s *reconciliationService) MatchLine(ctx context.Context, statementID, lineID, entryID uuid.UUID, byUserID uuid.UUID) (*domain.BankStatement, error) {
	ctx, span := s.tracing.StartSpan(ctx, "reconciliation_service.match_line")
	defer span.End()

	// Verify statement exists and is not completed
	stmt, err := s.repo.GetStatementByID(ctx, statementID)
	if err != nil {
		return nil, mapReconError(err, "get statement")
	}
	if stmt.Status == domain.ReconciliationStatusCompleted || stmt.Status == domain.ReconciliationStatusVoided {
		return nil, errors.NewBusinessError("STATEMENT_CLOSED", "cannot modify a completed or voided statement").
			WithHTTPStatus(http.StatusConflict)
	}

	// Transition to IN_PROGRESS if still DRAFT
	if stmt.Status == domain.ReconciliationStatusDraft {
		stmt.Status = domain.ReconciliationStatusInProgress
	}

	if err := s.repo.MatchLine(ctx, lineID, entryID, byUserID); err != nil {
		span.RecordError(err)
		return nil, mapReconError(err, "match line")
	}

	// Update counters
	stmt.MatchedCount++
	if stmt.UnmatchedCount > 0 {
		stmt.UnmatchedCount--
	}
	stmt.UpdatedBy = &byUserID
	if updateErr := s.repo.UpdateStatement(ctx, stmt); updateErr != nil {
		// Non-fatal: counters are best-effort
		span.RecordError(updateErr)
	}

	s.auditWriter.writeAudit(ctx, audit.CreateAuditEventRequest{
		UserID:        &byUserID,
		EventType:     auditTypeLineMatched,
		EventCategory: auditCategoryFinance,
		Severity:      auditSeverityInfo,
		ResourceID:    uuidPtr(statementID),
		Context: auditCtx(map[string]any{
			"statement_id": statementID.String(),
			"line_id":      lineID.String(),
			"entry_id":     entryID.String(),
		}),
	}, auditTypeLineMatched+":"+lineID.String())
	return stmt, nil
}

func (s *reconciliationService) UnmatchLine(ctx context.Context, statementID, lineID uuid.UUID) (*domain.BankStatement, error) {
	ctx, span := s.tracing.StartSpan(ctx, "reconciliation_service.unmatch_line")
	defer span.End()

	stmt, err := s.repo.GetStatementByID(ctx, statementID)
	if err != nil {
		return nil, mapReconError(err, "get statement")
	}
	if stmt.Status == domain.ReconciliationStatusCompleted {
		return nil, errors.NewBusinessError("STATEMENT_CLOSED", "cannot modify a completed statement").
			WithHTTPStatus(http.StatusConflict)
	}

	if err := s.repo.UnmatchLine(ctx, lineID); err != nil {
		span.RecordError(err)
		return nil, mapReconError(err, "unmatch line")
	}

	// Update counters
	if stmt.MatchedCount > 0 {
		stmt.MatchedCount--
	}
	stmt.UnmatchedCount++
	if updateErr := s.repo.UpdateStatement(ctx, stmt); updateErr != nil {
		span.RecordError(updateErr)
	}

	s.auditWriter.writeAudit(ctx, audit.CreateAuditEventRequest{
		UserID:        auditUserID(ctx),
		EventType:     auditTypeLineUnmatched,
		EventCategory: auditCategoryFinance,
		Severity:      auditSeverityInfo,
		ResourceID:    uuidPtr(statementID),
		Context: auditCtx(map[string]any{
			"statement_id": statementID.String(),
			"line_id":      lineID.String(),
		}),
	}, auditTypeLineUnmatched+":"+lineID.String())
	return stmt, nil
}

func (s *reconciliationService) ListLines(ctx context.Context, statementID uuid.UUID, unmatchedOnly bool) ([]*domain.BankStatementLine, error) {
	ctx, span := s.tracing.StartSpan(ctx, "reconciliation_service.list_lines")
	defer span.End()

	lines, err := s.repo.ListLines(ctx, statementID, unmatchedOnly)
	if err != nil {
		return nil, mapReconError(err, "list lines")
	}
	return lines, nil
}

func (s *reconciliationService) CompleteReconciliation(ctx context.Context, statementID uuid.UUID, byUserID uuid.UUID) (*domain.BankStatement, error) {
	ctx, span := s.tracing.StartSpan(ctx, "reconciliation_service.complete")
	defer span.End()

	stmt, err := s.repo.GetStatementByID(ctx, statementID)
	if err != nil {
		return nil, mapReconError(err, "get statement")
	}
	if stmt.Status == domain.ReconciliationStatusCompleted {
		return nil, errors.NewBusinessError("STATEMENT_ALREADY_COMPLETED", "statement is already completed").
			WithHTTPStatus(http.StatusConflict)
	}
	if stmt.Status == domain.ReconciliationStatusVoided {
		return nil, errors.NewBusinessError("STATEMENT_VOIDED", "cannot complete a voided statement").
			WithHTTPStatus(http.StatusConflict)
	}

	// Compute final difference: closing balance - (opening + net matched movements)
	// For now store the difference as closing - opening; a full implementation would
	// subtract the GL balance at statement date.
	stmt.DifferenceAmount = stmt.ClosingBalance.Sub(stmt.OpeningBalance)

	if err := s.repo.CompleteReconciliation(ctx, statementID, byUserID); err != nil {
		span.RecordError(err)
		return nil, mapReconError(err, "complete reconciliation")
	}

	stmt.Status = domain.ReconciliationStatusCompleted
	stmt.UpdatedBy = &byUserID
	stmt.UpdatedAt = time.Now()

	s.metrics.IncrementCounter("reconciliation_completed_total", metrics.Fields{})
	s.auditWriter.writeAudit(ctx, audit.CreateAuditEventRequest{
		UserID:        &byUserID,
		EventType:     auditTypeReconciled,
		EventCategory: auditCategoryFinance,
		Severity:      auditSeverityMedium,
		ResourceID:    uuidPtr(statementID),
		Context: auditCtx(map[string]any{
			"statement_id":      statementID.String(),
			"matched_count":     stmt.MatchedCount,
			"unmatched_count":   stmt.UnmatchedCount,
			"difference_amount": stmt.DifferenceAmount.String(),
		}),
	}, auditTypeReconciled+":"+statementID.String())
	return stmt, nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func mapReconError(err error, op string) error {
	switch err {
	case domain.ErrStatementNotFound:
		return errors.NewBusinessError("STATEMENT_NOT_FOUND", "bank statement not found").
			WithHTTPStatus(http.StatusNotFound)
	case domain.ErrStatementAlreadyClosed:
		return errors.NewBusinessError("STATEMENT_ALREADY_COMPLETED", "bank statement is already completed").
			WithHTTPStatus(http.StatusConflict)
	case domain.ErrLineNotFound:
		return errors.NewBusinessError("LINE_NOT_FOUND", "statement line not found").
			WithHTTPStatus(http.StatusNotFound)
	case domain.ErrLineAlreadyMatched:
		return errors.NewBusinessError("LINE_ALREADY_MATCHED", "statement line is already matched").
			WithHTTPStatus(http.StatusConflict)
	case domain.ErrEntryAlreadyReconciled:
		return errors.NewBusinessError("ENTRY_ALREADY_RECONCILED", "journal entry is already reconciled").
			WithHTTPStatus(http.StatusConflict)
	default:
		return errors.NewBusinessError("RECONCILIATION_ERROR", fmt.Sprintf("%s: %v", op, err)).
			WithHTTPStatus(http.StatusInternalServerError)
	}
}
