package repository

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	db "awo.so/db/sqlc"
	"awo.so/internal/core/finance/domain"
	"awo.so/internal/platform/cache"
	"awo.so/internal/shared"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

type periodRepository struct {
	store   db.Store
	tracing tracing.Service
	logger  logger.Logger
	metrics metrics.MetricsProvider
	cache   cache.Service
}

// NewPeriodRepository returns a new PeriodRepository.
func NewPeriodRepository(store db.Store, tracer tracing.Service, log logger.Logger, met metrics.MetricsProvider, cacheService cache.Service) domain.PeriodRepository {
	initFinanceRepoMetrics(met)
	return &periodRepository{
		store:   store,
		tracing: tracer,
		logger:  log,
		metrics: met,
		cache:   cacheService,
	}
}

// ── FiscalYear ───────────────────────────────────────────────────────────────

func (r *periodRepository) CreateFiscalYear(ctx context.Context, fy *domain.FiscalYear) (err error) {
	ctx, span := r.tracing.StartSpan(ctx, "PeriodRepository.CreateFiscalYear")
	defer span.End()
	start := time.Now()
	defer func() { observeOp(ctx, "CreateFiscalYear", "period", start, err, r.logger, r.metrics) }()

	if _, ok := shared.GetTenantID(ctx); !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	return r.store.WithTenantFromCtx(ctx, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}
		const q = `
INSERT INTO finance_fiscal_years
  (tenant_id, name, start_date, end_date, is_closed, is_locked, created_by)
VALUES
  (current_tenant_id(), $1, $2, $3, $4, $5, $6)
RETURNING id, created_at, updated_at`

		return tx.QueryRow(ctx, q,
			fy.Name,
			fy.StartDate,
			fy.EndDate,
			fy.IsClosed,
			fy.IsLocked,
			nullUUID(fy.CreatedBy),
		).Scan(&fy.ID, &fy.CreatedAt, &fy.UpdatedAt)
	})
}

func (r *periodRepository) GetFiscalYearByID(ctx context.Context, id uuid.UUID) (_ *domain.FiscalYear, err error) {
	ctx, span := r.tracing.StartSpan(ctx, "PeriodRepository.GetFiscalYearByID")
	defer span.End()
	start := time.Now()
	defer func() { observeOp(ctx, "GetFiscalYearByID", "period", start, err, r.logger, r.metrics) }()

	if _, ok := shared.GetTenantID(ctx); !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var result *domain.FiscalYear
	err = r.store.WithTenantFromCtx(ctx, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}
		const q = `
SELECT id, tenant_id, name, start_date, end_date, is_closed, is_locked,
       created_at, updated_at, created_by, updated_by
FROM   finance_fiscal_years
WHERE  tenant_id = current_tenant_id()
  AND  id        = $1`

		fy := &domain.FiscalYear{}
		var createdBy *uuid.UUID
		err = tx.QueryRow(ctx, q, id).Scan(
			&fy.ID, &fy.TenantID,
			&fy.Name, &fy.StartDate, &fy.EndDate,
			&fy.IsClosed, &fy.IsLocked,
			&fy.CreatedAt, &fy.UpdatedAt,
			&createdBy, &fy.UpdatedBy,
		)
		if err == pgx.ErrNoRows {
			return domain.ErrPeriodNotFound
		}
		if err != nil {
			return fmt.Errorf("get fiscal year by id: %w", err)
		}
		if createdBy != nil {
			fy.CreatedBy = *createdBy
		}
		fy.Year = fy.StartDate.Year()
		result = fy
		return nil
	})
	return result, err
}

func (r *periodRepository) GetFiscalYearByYear(ctx context.Context, tenantID uuid.UUID, year int) (_ *domain.FiscalYear, err error) {
	ctx, span := r.tracing.StartSpan(ctx, "PeriodRepository.GetFiscalYearByYear")
	defer span.End()
	start := time.Now()
	defer func() { observeOp(ctx, "GetFiscalYearByYear", "period", start, err, r.logger, r.metrics) }()

	// Cache check — fiscal years are immutable after creation.
	cacheKey := "fiscal_year:year:" + strconv.Itoa(year)
	if r.cache != nil {
		var cached domain.FiscalYear
		if cErr := r.cache.GetMemory(ctx, cacheKey, &cached); cErr == nil {
			return &cached, nil
		}
	}

	var result *domain.FiscalYear
	err = r.store.WithTenantFromCtx(ctx, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}
		const q = `
SELECT id, tenant_id, name, start_date, end_date, is_closed, is_locked,
       created_at, updated_at, created_by, updated_by
FROM   finance_fiscal_years
WHERE  tenant_id = current_tenant_id()
  AND  EXTRACT(YEAR FROM start_date)::INT = $1
LIMIT  1`

		fy := &domain.FiscalYear{}
		var createdBy *uuid.UUID
		err = tx.QueryRow(ctx, q, year).Scan(
			&fy.ID, &fy.TenantID,
			&fy.Name, &fy.StartDate, &fy.EndDate,
			&fy.IsClosed, &fy.IsLocked,
			&fy.CreatedAt, &fy.UpdatedAt,
			&createdBy, &fy.UpdatedBy,
		)
		if err == pgx.ErrNoRows {
			return domain.ErrPeriodNotFound
		}
		if err != nil {
			return fmt.Errorf("get fiscal year by year: %w", err)
		}
		if createdBy != nil {
			fy.CreatedBy = *createdBy
		}
		fy.Year = fy.StartDate.Year()
		result = fy
		return nil
	})
	// Populate cache on successful fetch — 30 min TTL; fiscal years are stable.
	if err == nil && result != nil && r.cache != nil {
		_ = r.cache.SetMemory(ctx, cacheKey, result, 30*time.Minute)
	}
	return result, err
}

func (r *periodRepository) ListFiscalYears(ctx context.Context, tenantID uuid.UUID) (_ []*domain.FiscalYear, err error) {
	ctx, span := r.tracing.StartSpan(ctx, "PeriodRepository.ListFiscalYears")
	defer span.End()
	start := time.Now()
	defer func() { observeOp(ctx, "ListFiscalYears", "period", start, err, r.logger, r.metrics) }()

	var results []*domain.FiscalYear
	err = r.store.WithTenantFromCtx(ctx, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}
		const q = `
SELECT id, tenant_id, name, start_date, end_date, is_closed, is_locked,
       created_at, updated_at, created_by, updated_by
FROM   finance_fiscal_years
WHERE  tenant_id = current_tenant_id()
ORDER  BY start_date DESC`

		rows, err := tx.Query(ctx, q)
		if err != nil {
			return fmt.Errorf("list fiscal years: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			fy := &domain.FiscalYear{}
			var createdBy *uuid.UUID
			if err := rows.Scan(
				&fy.ID, &fy.TenantID,
				&fy.Name, &fy.StartDate, &fy.EndDate,
				&fy.IsClosed, &fy.IsLocked,
				&fy.CreatedAt, &fy.UpdatedAt,
				&createdBy, &fy.UpdatedBy,
			); err != nil {
				return fmt.Errorf("scan fiscal year: %w", err)
			}
			if createdBy != nil {
				fy.CreatedBy = *createdBy
			}
			fy.Year = fy.StartDate.Year()
			results = append(results, fy)
		}
		return rows.Err()
	})
	return results, err
}

func (r *periodRepository) UpdateFiscalYear(ctx context.Context, fy *domain.FiscalYear) (err error) {
	ctx, span := r.tracing.StartSpan(ctx, "PeriodRepository.UpdateFiscalYear")
	defer span.End()
	start := time.Now()
	defer func() { observeOp(ctx, "UpdateFiscalYear", "period", start, err, r.logger, r.metrics) }()

	if _, ok := shared.GetTenantID(ctx); !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	err = r.store.WithTenantFromCtx(ctx, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}
		const q = `
UPDATE finance_fiscal_years
SET    name       = $2,
       start_date = $3,
       end_date   = $4,
       is_closed  = $5,
       is_locked  = $6,
       updated_at = NOW(),
       updated_by = $7
WHERE  tenant_id = current_tenant_id()
  AND  id        = $1
RETURNING updated_at`

		var updatedBy any
		if fy.UpdatedBy != nil {
			updatedBy = nullUUID(*fy.UpdatedBy)
		}
		return tx.QueryRow(ctx, q,
			fy.ID, fy.Name, fy.StartDate, fy.EndDate,
			fy.IsClosed, fy.IsLocked,
			updatedBy,
		).Scan(&fy.UpdatedAt)
	})
	// Invalidate fiscal-year cache on successful update.
	if err == nil && r.cache != nil {
		cacheKey := "fiscal_year:year:" + strconv.Itoa(fy.StartDate.Year())
		_ = r.cache.DeleteMemory(ctx, cacheKey)
	}
	return err
}

// ── AccountingPeriod ─────────────────────────────────────────────────────────

func (r *periodRepository) CreatePeriod(ctx context.Context, period *domain.AccountingPeriod) (err error) {
	ctx, span := r.tracing.StartSpan(ctx, "PeriodRepository.CreatePeriod")
	defer span.End()
	start := time.Now()
	defer func() { observeOp(ctx, "CreatePeriod", "period", start, err, r.logger, r.metrics) }()

	if _, ok := shared.GetTenantID(ctx); !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	return r.store.WithTenantFromCtx(ctx, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}
		const q = `
INSERT INTO finance_accounting_periods
  (tenant_id, fiscal_year_id, period_number, name, start_date, end_date, status)
VALUES
  (current_tenant_id(), $1, $2, $3, $4, $5, $6)
RETURNING id, created_at, updated_at`

		return tx.QueryRow(ctx, q,
			period.FiscalYearID,
			period.PeriodNumber,
			period.Name,
			period.StartDate,
			period.EndDate,
			string(period.Status),
		).Scan(&period.ID, &period.CreatedAt, &period.UpdatedAt)
	})
}

func (r *periodRepository) GetPeriodByID(ctx context.Context, id uuid.UUID) (_ *domain.AccountingPeriod, err error) {
	ctx, span := r.tracing.StartSpan(ctx, "PeriodRepository.GetPeriodByID")
	defer span.End()
	start := time.Now()
	defer func() { observeOp(ctx, "GetPeriodByID", "period", start, err, r.logger, r.metrics) }()

	if _, ok := shared.GetTenantID(ctx); !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var result *domain.AccountingPeriod
	err = r.store.WithTenantFromCtx(ctx, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}
		const q = `
SELECT id, tenant_id, fiscal_year_id, period_number, name,
       start_date, end_date, status,
       closed_at, closed_by, locked_at, locked_by,
       created_at, updated_at
FROM   finance_accounting_periods
WHERE  tenant_id = current_tenant_id()
  AND  id        = $1`

		p := &domain.AccountingPeriod{}
		var status string
		err = tx.QueryRow(ctx, q, id).Scan(
			&p.ID, &p.TenantID, &p.FiscalYearID,
			&p.PeriodNumber, &p.Name,
			&p.StartDate, &p.EndDate, &status,
			&p.ClosedAt, &p.ClosedBy,
			&p.LockedAt, &p.LockedBy,
			&p.CreatedAt, &p.UpdatedAt,
		)
		if err == pgx.ErrNoRows {
			return domain.ErrPeriodNotFound
		}
		if err != nil {
			return fmt.Errorf("get period by id: %w", err)
		}
		p.Status = domain.PeriodStatus(status)
		result = p
		return nil
	})
	return result, err
}

func (r *periodRepository) GetPeriodForDate(ctx context.Context, tenantID uuid.UUID, date time.Time) (_ *domain.AccountingPeriod, err error) {
	ctx, span := r.tracing.StartSpan(ctx, "PeriodRepository.GetPeriodForDate")
	defer span.End()
	start := time.Now()
	defer func() { observeOp(ctx, "GetPeriodForDate", "period", start, err, r.logger, r.metrics) }()

	// Cache check — called on every transaction post; short TTL prevents stale reads.
	cacheKey := "period:date:" + date.Format("2006-01-02")
	if r.cache != nil {
		var cached domain.AccountingPeriod
		if cErr := r.cache.GetMemory(ctx, cacheKey, &cached); cErr == nil {
			return &cached, nil
		}
	}

	var result *domain.AccountingPeriod
	err = r.store.WithTenantFromCtx(ctx, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}
		const q = `
SELECT id, tenant_id, fiscal_year_id, period_number, name,
       start_date, end_date, status,
       closed_at, closed_by, locked_at, locked_by,
       created_at, updated_at
FROM   finance_accounting_periods
WHERE  tenant_id   = current_tenant_id()
  AND  start_date <= $1::DATE
  AND  end_date   >= $1::DATE
LIMIT  1`

		p := &domain.AccountingPeriod{}
		var status string
		err = tx.QueryRow(ctx, q, date).Scan(
			&p.ID, &p.TenantID, &p.FiscalYearID,
			&p.PeriodNumber, &p.Name,
			&p.StartDate, &p.EndDate, &status,
			&p.ClosedAt, &p.ClosedBy,
			&p.LockedAt, &p.LockedBy,
			&p.CreatedAt, &p.UpdatedAt,
		)
		if err == pgx.ErrNoRows {
			return domain.ErrPeriodNotFound
		}
		if err != nil {
			return fmt.Errorf("get period for date: %w", err)
		}
		p.Status = domain.PeriodStatus(status)
		result = p
		return nil
	})
	// Populate cache on hit — 5 min TTL; short enough to pick up period changes.
	if err == nil && result != nil && r.cache != nil {
		_ = r.cache.SetMemory(ctx, cacheKey, result, 5*time.Minute)
	}
	return result, err
}

func (r *periodRepository) GetCurrentPeriod(ctx context.Context, tenantID uuid.UUID) (*domain.AccountingPeriod, error) {
	return r.GetPeriodForDate(ctx, tenantID, time.Now())
}

func (r *periodRepository) ListPeriods(ctx context.Context, tenantID, fiscalYearID uuid.UUID) (_ []*domain.AccountingPeriod, err error) {
	ctx, span := r.tracing.StartSpan(ctx, "PeriodRepository.ListPeriods")
	defer span.End()
	start := time.Now()
	defer func() { observeOp(ctx, "ListPeriods", "period", start, err, r.logger, r.metrics) }()

	var results []*domain.AccountingPeriod
	err = r.store.WithTenantFromCtx(ctx, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}
		const q = `
SELECT id, tenant_id, fiscal_year_id, period_number, name,
       start_date, end_date, status,
       closed_at, closed_by, locked_at, locked_by,
       created_at, updated_at
FROM   finance_accounting_periods
WHERE  tenant_id      = current_tenant_id()
  AND  fiscal_year_id = $1
ORDER  BY period_number ASC`

		rows, err := tx.Query(ctx, q, fiscalYearID)
		if err != nil {
			return fmt.Errorf("list periods: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			p := &domain.AccountingPeriod{}
			var status string
			if err := rows.Scan(
				&p.ID, &p.TenantID, &p.FiscalYearID,
				&p.PeriodNumber, &p.Name,
				&p.StartDate, &p.EndDate, &status,
				&p.ClosedAt, &p.ClosedBy,
				&p.LockedAt, &p.LockedBy,
				&p.CreatedAt, &p.UpdatedAt,
			); err != nil {
				return fmt.Errorf("scan period: %w", err)
			}
			p.Status = domain.PeriodStatus(status)
			results = append(results, p)
		}
		return rows.Err()
	})
	return results, err
}

func (r *periodRepository) UpdatePeriod(ctx context.Context, period *domain.AccountingPeriod) (err error) {
	ctx, span := r.tracing.StartSpan(ctx, "PeriodRepository.UpdatePeriod")
	defer span.End()
	start := time.Now()
	defer func() { observeOp(ctx, "UpdatePeriod", "period", start, err, r.logger, r.metrics) }()

	if _, ok := shared.GetTenantID(ctx); !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	err = r.store.WithTenantFromCtx(ctx, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}
		const q = `
UPDATE finance_accounting_periods
SET    status     = $2,
       closed_at  = $3,
       closed_by  = $4,
       locked_at  = $5,
       locked_by  = $6,
       updated_at = NOW()
WHERE  tenant_id = current_tenant_id()
  AND  id        = $1
RETURNING updated_at`

		return tx.QueryRow(ctx, q,
			period.ID,
			string(period.Status),
			period.ClosedAt,
			period.ClosedBy,
			period.LockedAt,
			period.LockedBy,
		).Scan(&period.UpdatedAt)
	})
	// Invalidate period-for-date cache entries spanning this period's dates.
	// Simplest safe approach: delete the cached entry for each date in the range.
	// For now, invalidate StartDate and EndDate keys (callers typically use these).
	if err == nil && r.cache != nil {
		_ = r.cache.DeleteMemory(ctx, "period:date:"+period.StartDate.Format("2006-01-02"))
		_ = r.cache.DeleteMemory(ctx, "period:date:"+period.EndDate.Format("2006-01-02"))
	}
	return err
}
