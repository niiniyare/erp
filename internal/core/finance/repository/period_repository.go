package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	db "awo.so/db/sqlc"
	"awo.so/internal/core/finance/domain"
	"awo.so/internal/shared"
	"awo.so/internal/shared/tracing"
)

type periodRepository struct {
	store   db.Store
	tracing tracing.Service
}

// NewPeriodRepository returns a new PeriodRepository.
func NewPeriodRepository(store db.Store, tracing tracing.Service) domain.PeriodRepository {
	return &periodRepository{store: store, tracing: tracing}
}

// ── FiscalYear ───────────────────────────────────────────────────────────────

func (r *periodRepository) CreateFiscalYear(ctx context.Context, fy *domain.FiscalYear) error {
	ctx, span := r.tracing.StartSpan(ctx, "PeriodRepository.CreateFiscalYear")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
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

func (r *periodRepository) GetFiscalYearByID(ctx context.Context, id uuid.UUID) (*domain.FiscalYear, error) {
	ctx, span := r.tracing.StartSpan(ctx, "PeriodRepository.GetFiscalYearByID")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var result *domain.FiscalYear
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
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

func (r *periodRepository) GetFiscalYearByYear(ctx context.Context, tenantID uuid.UUID, year int) (*domain.FiscalYear, error) {
	ctx, span := r.tracing.StartSpan(ctx, "PeriodRepository.GetFiscalYearByYear")
	defer span.End()

	var result *domain.FiscalYear
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
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
	return result, err
}

func (r *periodRepository) ListFiscalYears(ctx context.Context, tenantID uuid.UUID) ([]*domain.FiscalYear, error) {
	ctx, span := r.tracing.StartSpan(ctx, "PeriodRepository.ListFiscalYears")
	defer span.End()

	var results []*domain.FiscalYear
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
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

func (r *periodRepository) UpdateFiscalYear(ctx context.Context, fy *domain.FiscalYear) error {
	ctx, span := r.tracing.StartSpan(ctx, "PeriodRepository.UpdateFiscalYear")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
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

		var updatedBy interface{}
		if fy.UpdatedBy != nil {
			updatedBy = nullUUID(*fy.UpdatedBy)
		}
		return tx.QueryRow(ctx, q,
			fy.ID, fy.Name, fy.StartDate, fy.EndDate,
			fy.IsClosed, fy.IsLocked,
			updatedBy,
		).Scan(&fy.UpdatedAt)
	})
}

// ── AccountingPeriod ─────────────────────────────────────────────────────────

func (r *periodRepository) CreatePeriod(ctx context.Context, period *domain.AccountingPeriod) error {
	ctx, span := r.tracing.StartSpan(ctx, "PeriodRepository.CreatePeriod")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
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

func (r *periodRepository) GetPeriodByID(ctx context.Context, id uuid.UUID) (*domain.AccountingPeriod, error) {
	ctx, span := r.tracing.StartSpan(ctx, "PeriodRepository.GetPeriodByID")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var result *domain.AccountingPeriod
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
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

func (r *periodRepository) GetPeriodForDate(ctx context.Context, tenantID uuid.UUID, date time.Time) (*domain.AccountingPeriod, error) {
	ctx, span := r.tracing.StartSpan(ctx, "PeriodRepository.GetPeriodForDate")
	defer span.End()

	var result *domain.AccountingPeriod
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
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
	return result, err
}

func (r *periodRepository) GetCurrentPeriod(ctx context.Context, tenantID uuid.UUID) (*domain.AccountingPeriod, error) {
	return r.GetPeriodForDate(ctx, tenantID, time.Now())
}

func (r *periodRepository) ListPeriods(ctx context.Context, tenantID, fiscalYearID uuid.UUID) ([]*domain.AccountingPeriod, error) {
	ctx, span := r.tracing.StartSpan(ctx, "PeriodRepository.ListPeriods")
	defer span.End()

	var results []*domain.AccountingPeriod
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
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

func (r *periodRepository) UpdatePeriod(ctx context.Context, period *domain.AccountingPeriod) error {
	ctx, span := r.tracing.StartSpan(ctx, "PeriodRepository.UpdatePeriod")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
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
}

