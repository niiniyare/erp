package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"awo.so/internal/core/audit"
	"awo.so/internal/core/finance/domain"
	"awo.so/internal/shared"
	"awo.so/internal/shared/errors"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

// PeriodService manages fiscal years and accounting periods.
type PeriodService interface {
	// Fiscal years
	CreateFiscalYear(ctx context.Context, fy *domain.FiscalYear) (*domain.FiscalYear, error)
	GetFiscalYearByID(ctx context.Context, id uuid.UUID) (*domain.FiscalYear, error)
	ListFiscalYears(ctx context.Context) ([]*domain.FiscalYear, error)

	// Accounting periods
	CreatePeriod(ctx context.Context, period *domain.AccountingPeriod) (*domain.AccountingPeriod, error)
	GetPeriodByID(ctx context.Context, id uuid.UUID) (*domain.AccountingPeriod, error)
	GetCurrentPeriod(ctx context.Context) (*domain.AccountingPeriod, error)
	ListPeriods(ctx context.Context, fiscalYearID uuid.UUID) ([]*domain.AccountingPeriod, error)

	// Period lifecycle transitions
	ChangePeriodStatus(ctx context.Context, id uuid.UUID, newStatus domain.PeriodStatus, byUserID uuid.UUID, checksPassed bool) (*domain.AccountingPeriod, error)
}

type periodService struct {
	repo             domain.PeriodRepository
	integrityService IntegrityService                // nil → hard-close runs without integrity gate
	escalation       *IntegrityEscalationService     // nil → violation-lifecycle blocking skipped
	tracing          tracing.Service
	metrics          metrics.MetricsProvider
	auditWriter      *financeAuditWriter // nil → audit skipped
}

// NewPeriodService creates a new PeriodService without an integrity gate.
// Use NewPeriodServiceWithIntegrity to enforce ledger health checks before
// hard-closing a period.
func NewPeriodService(
	repo domain.PeriodRepository,
	tracing tracing.Service,
	metrics metrics.MetricsProvider,
	aw *financeAuditWriter,
	escalation *IntegrityEscalationService,
) PeriodService {
	return &periodService{
		repo:        repo,
		tracing:     tracing,
		metrics:     metrics,
		auditWriter: aw,
		escalation:  escalation,
	}
}

// NewPeriodServiceWithIntegrity creates a PeriodService that gates HARD_CLOSED
// transitions on a clean integrity scan. If the scan finds any CRITICAL
// violations, the hard-close is refused until they are resolved.
//
// This is the recommended constructor for production deployments. The simpler
// NewPeriodService is provided for wiring contexts where IntegrityService is
// not yet available (e.g. integration tests without a full finance stack).
func NewPeriodServiceWithIntegrity(
	repo domain.PeriodRepository,
	integrityService IntegrityService,
	tracing tracing.Service,
	metrics metrics.MetricsProvider,
	aw *financeAuditWriter,
	escalation *IntegrityEscalationService,
) PeriodService {
	return &periodService{
		repo:             repo,
		integrityService: integrityService,
		tracing:          tracing,
		metrics:          metrics,
		auditWriter:      aw,
		escalation:       escalation,
	}
}

func (s *periodService) CreateFiscalYear(ctx context.Context, fy *domain.FiscalYear) (*domain.FiscalYear, error) {
	ctx, span := s.tracing.StartSpan(ctx, "period_service.create_fiscal_year")
	defer span.End()

	if fy.Name == "" {
		return nil, errors.NewBusinessError("FISCAL_YEAR_VALIDATION_FAILED", "fiscal year name is required").
			WithHTTPStatus(http.StatusBadRequest).
			WithCategory(errors.CategoryValidation)
	}
	if !fy.EndDate.After(fy.StartDate) {
		return nil, errors.NewBusinessError("FISCAL_YEAR_VALIDATION_FAILED", "end_date must be after start_date").
			WithHTTPStatus(http.StatusBadRequest).
			WithCategory(errors.CategoryValidation)
	}

	if err := s.repo.CreateFiscalYear(ctx, fy); err != nil {
		span.RecordError(err)
		return nil, mapPeriodError(err, "create fiscal year")
	}

	s.metrics.IncrementCounter("fiscal_year_created_total", metrics.Fields{})
	return fy, nil
}

func (s *periodService) GetFiscalYearByID(ctx context.Context, id uuid.UUID) (*domain.FiscalYear, error) {
	ctx, span := s.tracing.StartSpan(ctx, "period_service.get_fiscal_year")
	defer span.End()

	fy, err := s.repo.GetFiscalYearByID(ctx, id)
	if err != nil {
		return nil, mapPeriodError(err, "get fiscal year")
	}
	return fy, nil
}

func (s *periodService) ListFiscalYears(ctx context.Context) ([]*domain.FiscalYear, error) {
	ctx, span := s.tracing.StartSpan(ctx, "period_service.list_fiscal_years")
	defer span.End()

	// Tenant ID is carried in ctx via middleware; ListFiscalYears uses it internally.
	// We need a tenantID to pass — retrieve it from context via shared helper.
	tenantID, err := tenantIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	fys, err := s.repo.ListFiscalYears(ctx, tenantID)
	if err != nil {
		return nil, mapPeriodError(err, "list fiscal years")
	}
	return fys, nil
}

func (s *periodService) CreatePeriod(ctx context.Context, period *domain.AccountingPeriod) (*domain.AccountingPeriod, error) {
	ctx, span := s.tracing.StartSpan(ctx, "period_service.create_period")
	defer span.End()

	if errs := period.Validate(); len(errs) > 0 {
		return nil, errors.NewBusinessError("PERIOD_VALIDATION_FAILED", errs[0].Message).
			WithHTTPStatus(http.StatusBadRequest).
			WithCategory(errors.CategoryValidation)
	}

	if period.Status == "" {
		period.Status = domain.PeriodStatusOpen
	}

	if err := s.repo.CreatePeriod(ctx, period); err != nil {
		span.RecordError(err)
		return nil, mapPeriodError(err, "create period")
	}

	s.metrics.IncrementCounter("accounting_period_created_total", metrics.Fields{})
	return period, nil
}

func (s *periodService) GetPeriodByID(ctx context.Context, id uuid.UUID) (*domain.AccountingPeriod, error) {
	ctx, span := s.tracing.StartSpan(ctx, "period_service.get_period")
	defer span.End()

	p, err := s.repo.GetPeriodByID(ctx, id)
	if err != nil {
		return nil, mapPeriodError(err, "get period")
	}
	return p, nil
}

func (s *periodService) GetCurrentPeriod(ctx context.Context) (*domain.AccountingPeriod, error) {
	ctx, span := s.tracing.StartSpan(ctx, "period_service.get_current_period")
	defer span.End()

	tenantID, err := tenantIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	p, err := s.repo.GetCurrentPeriod(ctx, tenantID)
	if err != nil {
		return nil, mapPeriodError(err, "get current period")
	}
	return p, nil
}

func (s *periodService) ListPeriods(ctx context.Context, fiscalYearID uuid.UUID) ([]*domain.AccountingPeriod, error) {
	ctx, span := s.tracing.StartSpan(ctx, "period_service.list_periods")
	defer span.End()

	tenantID, err := tenantIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	periods, err := s.repo.ListPeriods(ctx, tenantID, fiscalYearID)
	if err != nil {
		return nil, mapPeriodError(err, "list periods")
	}
	return periods, nil
}

func (s *periodService) ChangePeriodStatus(ctx context.Context, id uuid.UUID, newStatus domain.PeriodStatus, byUserID uuid.UUID, checksPassed bool) (*domain.AccountingPeriod, error) {
	ctx, span := s.tracing.StartSpan(ctx, "period_service.change_period_status")
	defer span.End()

	p, err := s.repo.GetPeriodByID(ctx, id)
	if err != nil {
		return nil, mapPeriodError(err, "get period for status change")
	}

	now := time.Now()
	switch newStatus {
	case domain.PeriodStatusSoftClosed:
		if err := p.SoftClose(byUserID, now); err != nil {
			return nil, errors.NewBusinessError("INVALID_TRANSITION", err.Error()).
				WithHTTPStatus(http.StatusUnprocessableEntity)
		}
	case domain.PeriodStatusHardClosed:
		// Integrity gate: if an IntegrityService is wired, run a ledger scan
		// before allowing hard-close. A CRITICAL violation means the ledger
		// has unresolved corruption that must be corrected before the period
		// can be frozen. The scan is tenant-scoped via RLS — no filter needed.
		if s.integrityService != nil {
			scanReport, scanErr := s.integrityService.ScanPostedTransactions(ctx, nil)
			if scanErr != nil {
				s.metrics.IncrementCounter("period_close_integrity_scan_errors_total", metrics.Fields{})
				return nil, errors.NewBusinessError("INTEGRITY_SCAN_FAILED",
					fmt.Sprintf("could not verify ledger integrity before hard-close: %v", scanErr)).
					WithHTTPStatus(http.StatusInternalServerError)
			}
			if scanReport.HasCritical() {
				s.metrics.IncrementCounter("period_close_blocked_total", metrics.Fields{
					"reason": "critical_integrity_violations",
				})
				return nil, errors.NewBusinessError("PERIOD_CLOSE_BLOCKED",
					fmt.Sprintf(
						"hard-close refused: %d CRITICAL integrity violation(s) detected. "+
							"Resolve all CRITICAL violations and retry. "+
							"Run ScanPostedTransactions to view details.",
						len(scanReport.Violations),
					)).
					WithHTTPStatus(http.StatusUnprocessableEntity).
					WithCategory(errors.CategoryBusiness)
			}
			// Scan passed — set checksPassed=true so the domain method proceeds.
			checksPassed = true
		}
		// Escalation gate: block hard-close if unresolved CRITICAL violations exist
		// in the violation lifecycle tracker (persisted by IntegrityEscalationService).
		if blockErr := s.escalation.BlockIfCriticalOpen(ctx); blockErr != nil {
			return nil, blockErr
		}
		if err := p.HardClose(byUserID, now, checksPassed); err != nil {
			return nil, errors.NewBusinessError("INVALID_TRANSITION", err.Error()).
				WithHTTPStatus(http.StatusUnprocessableEntity)
		}
	case domain.PeriodStatusOpen:
		if err := p.Reopen(byUserID, now); err != nil {
			return nil, errors.NewBusinessError("INVALID_TRANSITION", err.Error()).
				WithHTTPStatus(http.StatusUnprocessableEntity)
		}
	case domain.PeriodStatusLocked:
		if err := p.Lock(byUserID, now); err != nil {
			return nil, errors.NewBusinessError("INVALID_TRANSITION", err.Error()).
				WithHTTPStatus(http.StatusUnprocessableEntity)
		}
	default:
		return nil, errors.NewBusinessError("INVALID_STATUS", fmt.Sprintf("unknown status: %s", newStatus)).
			WithHTTPStatus(http.StatusBadRequest)
	}

	if err := s.repo.UpdatePeriod(ctx, p); err != nil {
		span.RecordError(err)
		return nil, mapPeriodError(err, "update period status")
	}

	s.metrics.IncrementCounter("period_status_changed_total", metrics.Fields{
		"new_status": string(newStatus),
	})

	severity := auditSeverityMedium
	if newStatus == domain.PeriodStatusHardClosed {
		severity = auditSeverityHigh
	}
	s.auditWriter.writeAudit(ctx, audit.CreateAuditEventRequest{
		UserID:        &byUserID,
		EventType:     auditTypePeriodStatus,
		EventCategory: auditCategoryFinance,
		Severity:      severity,
		ResourceID:    uuidPtr(id),
		Context: auditCtx(map[string]any{
			"period_id":   id.String(),
			"new_status":  string(newStatus),
			"period_name": p.Name,
		}),
	}, auditTypePeriodStatus+":"+id.String()+":"+string(newStatus))
	return p, nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

// mapPeriodError converts domain sentinel errors to BusinessErrors with HTTP codes.
func mapPeriodError(err error, op string) error {
	if err == domain.ErrPeriodNotFound {
		return errors.NewBusinessError("PERIOD_NOT_FOUND", "accounting period not found").
			WithHTTPStatus(http.StatusNotFound)
	}
	return errors.NewBusinessError("PERIOD_ERROR", fmt.Sprintf("%s: %v", op, err)).
		WithHTTPStatus(http.StatusInternalServerError)
}

// tenantIDFromCtx extracts the tenant ID from the context.
func tenantIDFromCtx(ctx context.Context) (uuid.UUID, error) {
	id, ok := shared.GetTenantID(ctx)
	if !ok {
		return uuid.Nil, errors.NewBusinessError("MISSING_TENANT", "tenant ID not found in context").
			WithHTTPStatus(http.StatusUnauthorized)
	}
	return id, nil
}
