-- =====================================================================
-- FINANCE MODULE — EXCHANGE RATES QUERIES
-- Tenant-isolated exchange rate persistence.
-- Rate look-up uses on-date-or-before semantics (most recent rate
-- on or before the requested date for each currency pair).
-- =====================================================================

-- name: UpsertExchangeRate :one
INSERT INTO
  finance_exchange_rates (
    tenant_id,
    from_currency,
    to_currency,
    rate,
    rate_type,
    effective_date,
    expiry_date,
    source,
    created_by
  )
VALUES
  (
    current_tenant_id(),
    $1, $2, $3, $4, $5, $6, $7, $8
  )
ON CONFLICT (tenant_id, from_currency, to_currency, effective_date, rate_type)
  DO UPDATE SET
    rate       = EXCLUDED.rate,
    expiry_date = EXCLUDED.expiry_date,
    source     = EXCLUDED.source
RETURNING
  id, tenant_id, from_currency, to_currency, rate, rate_type,
  effective_date, expiry_date, source, created_at, created_by;

-- name: GetExchangeRate :one
-- Returns the most recent rate for a pair on or before asOfDate.
SELECT
  id, tenant_id, from_currency, to_currency, rate, rate_type,
  effective_date, expiry_date, source, created_at, created_by
FROM
  finance_exchange_rates
WHERE
  tenant_id      = current_tenant_id()
  AND from_currency = $1
  AND to_currency   = $2
  AND rate_type     = $3
  AND effective_date <= $4
  AND (expiry_date IS NULL OR expiry_date >= $4)
ORDER BY
  effective_date DESC
LIMIT 1;

-- name: ListExchangeRates :many
SELECT
  id, tenant_id, from_currency, to_currency, rate, rate_type,
  effective_date, expiry_date, source, created_at, created_by
FROM
  finance_exchange_rates
WHERE
  tenant_id      = current_tenant_id()
  AND from_currency = $1
  AND to_currency   = $2
  AND ($3::date IS NULL OR effective_date >= $3)
  AND ($4::date IS NULL OR effective_date <= $4)
ORDER BY
  effective_date DESC
LIMIT  $5
OFFSET $6;

-- name: DeleteExpiredExchangeRates :exec
DELETE FROM finance_exchange_rates
WHERE
  tenant_id   = current_tenant_id()
  AND expiry_date IS NOT NULL
  AND expiry_date < $1;
