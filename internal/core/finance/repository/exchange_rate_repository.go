package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	db "awo.so/db/sqlc"
	"awo.so/internal/core/finance/domain"
	"awo.so/internal/shared"
	"awo.so/internal/shared/tracing"
)

// exchangeRateRepository implements domain.ExchangeRateRepository.
// It executes raw SQL inside the tenant-aware transaction provided by
// db.TxStore.GetTx() — the table and SQLC typed helpers will be generated
// after `make sqlc` is run against the migration in db/migration.
type exchangeRateRepository struct {
	store   db.Store
	tracing tracing.Service
}

// NewExchangeRateRepository returns a new ExchangeRateRepository.
func NewExchangeRateRepository(store db.Store, tracing tracing.Service) domain.ExchangeRateRepository {
	return &exchangeRateRepository{store: store, tracing: tracing}
}

// txFrom extracts the underlying pgx.Tx from the store inside a WithTenant closure.
func txFrom(s db.Store) (pgx.Tx, error) {
	txs, ok := s.(db.TxStore)
	if !ok {
		return nil, fmt.Errorf("exchange_rate_repository: store is not a TxStore (must be called inside WithTenant)")
	}
	return txs.GetTx(), nil
}

// decimalToNumeric converts a decimal.Decimal to pgtype.Numeric.
func decimalToNumericER(d decimal.Decimal) pgtype.Numeric {
	n := pgtype.Numeric{}
	_ = n.Scan(d.String())
	return n
}

// numericToDecimalER converts pgtype.Numeric to decimal.Decimal.
func numericToDecimalER(n pgtype.Numeric) decimal.Decimal {
	if !n.Valid {
		return decimal.Zero
	}
	return pgTypeNumericToDecimal(n)
}

// Upsert inserts or updates a rate for (fromCurrency, toCurrency, effectiveDate, rateType).
func (r *exchangeRateRepository) Upsert(ctx context.Context, rate *domain.ExchangeRate) error {
	ctx, span := r.tracing.StartSpan(ctx, "ExchangeRateRepository.Upsert")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	var createdBy *uuid.UUID
	if uid, ok := shared.GetUserID(ctx); ok {
		createdBy = &uid
	}

	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}

		const q = `
INSERT INTO finance_exchange_rates
  (tenant_id, from_currency, to_currency, rate, rate_type,
   effective_date, expiry_date, source, created_by)
VALUES
  (current_tenant_id(), $1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (tenant_id, from_currency, to_currency, effective_date, rate_type)
  DO UPDATE SET
    rate        = EXCLUDED.rate,
    expiry_date = EXCLUDED.expiry_date,
    source      = EXCLUDED.source
RETURNING id, created_at`

		row := tx.QueryRow(ctx, q,
			rate.FromCurrency,
			rate.ToCurrency,
			decimalToNumericER(rate.Rate),
			string(rate.RateType),
			rate.EffectiveDate,
			rate.ExpiryDate,
			rate.Source,
			createdBy,
		)
		return row.Scan(&rate.ID, &rate.CreatedAt)
	})
}

// GetRate returns the most recent rate for the pair on or before asOfDate.
func (r *exchangeRateRepository) GetRate(ctx context.Context, fromCurrency, toCurrency string, rateType domain.RateType, asOfDate time.Time) (*domain.ExchangeRate, error) {
	ctx, span := r.tracing.StartSpan(ctx, "ExchangeRateRepository.GetRate")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var result *domain.ExchangeRate
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}

		const q = `
SELECT id, tenant_id, from_currency, to_currency, rate, rate_type,
       effective_date, expiry_date, source, created_at, created_by
FROM   finance_exchange_rates
WHERE  tenant_id      = current_tenant_id()
  AND  from_currency  = $1
  AND  to_currency    = $2
  AND  rate_type      = $3
  AND  effective_date <= $4
  AND  (expiry_date IS NULL OR expiry_date >= $4)
ORDER  BY effective_date DESC
LIMIT  1`

		var (
			id         uuid.UUID
			tid        uuid.UUID
			from, to   string
			rateNum    pgtype.Numeric
			rt, src    string
			effDate    time.Time
			expiryDate *time.Time
			createdAt  time.Time
			createdBy  *uuid.UUID
		)
		row := tx.QueryRow(ctx, q, fromCurrency, toCurrency, string(rateType), asOfDate)
		if err := row.Scan(&id, &tid, &from, &to, &rateNum, &rt, &effDate, &expiryDate, &src, &createdAt, &createdBy); err != nil {
			if err == pgx.ErrNoRows {
				return domain.ErrExchangeRateNotFound
			}
			return fmt.Errorf("get exchange rate: %w", err)
		}

		er := &domain.ExchangeRate{
			ID:            id,
			TenantID:      tid,
			FromCurrency:  from,
			ToCurrency:    to,
			Rate:          numericToDecimalER(rateNum),
			RateType:      domain.RateType(rt),
			EffectiveDate: effDate,
			ExpiryDate:    expiryDate,
			Source:        src,
			CreatedAt:     createdAt,
		}
		if createdBy != nil {
			er.CreatedBy = *createdBy
		}
		result = er
		return nil
	})
	return result, err
}

// ListRates lists rates for a currency pair within an optional date range.
func (r *exchangeRateRepository) ListRates(ctx context.Context, fromCurrency, toCurrency string, from, to *time.Time, limit int) ([]*domain.ExchangeRate, error) {
	ctx, span := r.tracing.StartSpan(ctx, "ExchangeRateRepository.ListRates")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	if limit <= 0 {
		limit = 100
	}

	var results []*domain.ExchangeRate
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}

		const q = `
SELECT id, tenant_id, from_currency, to_currency, rate, rate_type,
       effective_date, expiry_date, source, created_at, created_by
FROM   finance_exchange_rates
WHERE  tenant_id     = current_tenant_id()
  AND  from_currency = $1
  AND  to_currency   = $2
  AND  ($3::date IS NULL OR effective_date >= $3)
  AND  ($4::date IS NULL OR effective_date <= $4)
ORDER  BY effective_date DESC
LIMIT  $5`

		rows, err := tx.Query(ctx, q, fromCurrency, toCurrency, from, to, int32(limit))
		if err != nil {
			return fmt.Errorf("list exchange rates: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var (
				id         uuid.UUID
				tid        uuid.UUID
				fc, tc     string
				rateNum    pgtype.Numeric
				rt, src    string
				effDate    time.Time
				expiryDate *time.Time
				createdAt  time.Time
				createdBy  *uuid.UUID
			)
			if err := rows.Scan(&id, &tid, &fc, &tc, &rateNum, &rt, &effDate, &expiryDate, &src, &createdAt, &createdBy); err != nil {
				return fmt.Errorf("scan exchange rate: %w", err)
			}
			er := &domain.ExchangeRate{
				ID:            id,
				TenantID:      tid,
				FromCurrency:  fc,
				ToCurrency:    tc,
				Rate:          numericToDecimalER(rateNum),
				RateType:      domain.RateType(rt),
				EffectiveDate: effDate,
				ExpiryDate:    expiryDate,
				Source:        src,
				CreatedAt:     createdAt,
			}
			if createdBy != nil {
				er.CreatedBy = *createdBy
			}
			results = append(results, er)
		}
		return rows.Err()
	})
	return results, err
}

// DeleteExpired removes rates whose expiry_date is before cutoff.
func (r *exchangeRateRepository) DeleteExpired(ctx context.Context, cutoff time.Time) error {
	ctx, span := r.tracing.StartSpan(ctx, "ExchangeRateRepository.DeleteExpired")
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
DELETE FROM finance_exchange_rates
WHERE  tenant_id   = current_tenant_id()
  AND  expiry_date IS NOT NULL
  AND  expiry_date < $1`
		_, err = tx.Exec(ctx, q, cutoff)
		return err
	})
}
