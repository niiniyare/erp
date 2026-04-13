package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	db "awo.so/db/sqlc"
	"awo.so/internal/core/finance/domain"
	financeService "awo.so/internal/core/finance/service"
	"awo.so/internal/shared"
	"awo.so/internal/shared/tracing"
)

type currencyRepository struct {
	store   db.Store
	tracing tracing.Service
}

// NewCurrencyRepository returns an implementation of service.CurrencyRepository.
func NewCurrencyRepository(store db.Store, tracing tracing.Service) financeService.CurrencyRepository {
	return &currencyRepository{store: store, tracing: tracing}
}

func (r *currencyRepository) Create(ctx context.Context, c *domain.Currency) error {
	ctx, span := r.tracing.StartSpan(ctx, "CurrencyRepository.Create")
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
INSERT INTO finance_currencies
  (tenant_id, code, name, symbol, decimal_places, is_active, is_base, created_by)
VALUES
  (current_tenant_id(), $1, $2, $3, $4, $5, $6, $7)
RETURNING id, created_at, updated_at`

		return tx.QueryRow(ctx, q,
			c.Code,
			c.Name,
			c.Symbol,
			c.DecimalPlaces,
			c.IsActive,
			c.IsBase,
			nullUUID(c.CreatedBy),
		).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
	})
}

func (r *currencyRepository) GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (*domain.Currency, error) {
	ctx, span := r.tracing.StartSpan(ctx, "CurrencyRepository.GetByCode")
	defer span.End()

	var result *domain.Currency
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}
		const q = `
SELECT id, tenant_id, code, name, symbol, decimal_places, is_active, is_base,
       created_at, updated_at, created_by, updated_by
FROM   finance_currencies
WHERE  tenant_id = current_tenant_id()
  AND  code      = $1`

		c := &domain.Currency{}
		err = tx.QueryRow(ctx, q, code).Scan(
			&c.ID, &c.TenantID,
			&c.Code, &c.Name, &c.Symbol,
			&c.DecimalPlaces, &c.IsActive, &c.IsBase,
			&c.CreatedAt, &c.UpdatedAt,
			&c.CreatedBy, &c.UpdatedBy,
		)
		if err == pgx.ErrNoRows {
			return fmt.Errorf("currency not found: %s", code)
		}
		if err != nil {
			return fmt.Errorf("get currency by code: %w", err)
		}
		result = c
		return nil
	})
	return result, err
}

func (r *currencyRepository) List(ctx context.Context, tenantID uuid.UUID, activeOnly bool) ([]*domain.Currency, error) {
	ctx, span := r.tracing.StartSpan(ctx, "CurrencyRepository.List")
	defer span.End()

	var results []*domain.Currency
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}
		const q = `
SELECT id, tenant_id, code, name, symbol, decimal_places, is_active, is_base,
       created_at, updated_at, created_by, updated_by
FROM   finance_currencies
WHERE  tenant_id = current_tenant_id()
  AND  ($1 = FALSE OR is_active = TRUE)
ORDER  BY code ASC`

		rows, err := tx.Query(ctx, q, activeOnly)
		if err != nil {
			return fmt.Errorf("list currencies: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			c := &domain.Currency{}
			if err := rows.Scan(
				&c.ID, &c.TenantID,
				&c.Code, &c.Name, &c.Symbol,
				&c.DecimalPlaces, &c.IsActive, &c.IsBase,
				&c.CreatedAt, &c.UpdatedAt,
				&c.CreatedBy, &c.UpdatedBy,
			); err != nil {
				return fmt.Errorf("scan currency: %w", err)
			}
			results = append(results, c)
		}
		return rows.Err()
	})
	return results, err
}

func (r *currencyRepository) Update(ctx context.Context, c *domain.Currency) error {
	ctx, span := r.tracing.StartSpan(ctx, "CurrencyRepository.Update")
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
UPDATE finance_currencies
SET    name           = $2,
       symbol         = $3,
       decimal_places = $4,
       is_active      = $5,
       is_base        = $6,
       updated_at     = NOW(),
       updated_by     = $7
WHERE  tenant_id = current_tenant_id()
  AND  id        = $1
RETURNING updated_at`

		var updatedBy interface{}
		if c.UpdatedBy != nil {
			updatedBy = nullUUID(*c.UpdatedBy)
		}
		return tx.QueryRow(ctx, q,
			c.ID, c.Name, c.Symbol,
			c.DecimalPlaces, c.IsActive, c.IsBase,
			updatedBy,
		).Scan(&c.UpdatedAt)
	})
}
