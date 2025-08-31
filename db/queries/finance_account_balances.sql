-- name: CreateAccountBalance :one
INSERT INTO
  finance_account_balances (
    tenant_id,
    entity_id,
    account_id,
    balance_date,
    fiscal_year,
    fiscal_period,
    opening_balance,
    period_debits,
    period_credits,
    closing_balance,
    created_by
  )
VALUES
  (
    current_tenant_id(),
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10
  )
RETURNING
  *;

-- name: GetAccountBalance :one
SELECT
  *
FROM
  finance_account_balances
WHERE
  id = $1
  AND tenant_id = current_tenant_id();

-- name: GetAccountBalanceByDate :one
SELECT
  *
FROM
  finance_account_balances
WHERE
  account_id = $1
  AND balance_date = $2
  AND tenant_id = current_tenant_id()
ORDER BY
  created_at DESC
LIMIT
  1;

-- name: GetLatestAccountBalance :one
SELECT
  *
FROM
  finance_account_balances
WHERE
  account_id = $1
  AND tenant_id = current_tenant_id()
ORDER BY
  balance_date DESC,
  created_at DESC
LIMIT
  1;

-- name: UpdateAccountBalance :one
UPDATE
  finance_account_balances
SET
  opening_balance = COALESCE($2, opening_balance),
  period_debits = COALESCE($3, period_debits),
  period_credits = COALESCE($4, period_credits),
  closing_balance = COALESCE($5, closing_balance)
WHERE
  id = $1
  AND tenant_id = current_tenant_id()
RETURNING
  *;

-- name: DeleteAccountBalance :exec
DELETE FROM
  finance_account_balances
WHERE
  id = $1
  AND tenant_id = current_tenant_id();

-- name: ListAccountBalances :many
SELECT
  *
FROM
  finance_account_balances
WHERE
  tenant_id = current_tenant_id()
  AND (
    $1::uuid IS NULL
    OR entity_id = $1
  )
  AND (
    $2::uuid IS NULL
    OR account_id = $2
  )
  AND (
    $3::date IS NULL
    OR balance_date >= $3
  )
  AND (
    $4::date IS NULL
    OR balance_date <= $4
  )
  AND (
    $5::integer IS NULL
    OR fiscal_year = $5
  )
  AND (
    $6::integer IS NULL
    OR fiscal_period = $6
  )
ORDER BY
  account_id,
  balance_date DESC
LIMIT
  $7 OFFSET $8;

-- name: CountAccountBalances :one
SELECT
  count(*)
FROM
  finance_account_balances
WHERE
  tenant_id = current_tenant_id()
  AND (
    $1::uuid IS NULL
    OR entity_id = $1
  )
  AND (
    $2::uuid IS NULL
    OR account_id = $2
  )
  AND (
    $3::date IS NULL
    OR balance_date >= $3
  )
  AND (
    $4::date IS NULL
    OR balance_date <= $4
  )
  AND (
    $5::integer IS NULL
    OR fiscal_year = $5
  )
  AND (
    $6::integer IS NULL
    OR fiscal_period = $6
  );

-- name: GetAccountBalanceHistory :many
SELECT
  *
FROM
  finance_account_balances
WHERE
  account_id = $1
  AND tenant_id = current_tenant_id()
  AND (
    $2::uuid IS NULL
    OR entity_id = $2
  )
  AND (
    $3::date IS NULL
    OR balance_date >= $3
  )
  AND (
    $4::date IS NULL
    OR balance_date <= $4
  )
ORDER BY
  balance_date DESC,
  created_at DESC
LIMIT
  $5 OFFSET $6;

-- name: GetPeriodEndBalances :many
SELECT
  *
FROM
  finance_account_balances
WHERE
  tenant_id = current_tenant_id()
  AND (
    $1::uuid IS NULL
    OR entity_id = $1
  )
  AND (
    $2::integer IS NULL
    OR fiscal_year = $2
  )
  AND (
    $3::integer IS NULL
    OR fiscal_period = $3
  )
  AND (
    $4::date IS NULL
    OR balance_date = $4
  )
ORDER BY
  account_id;

-- name: GetBalancesByYear :many
SELECT
  *
FROM
  finance_account_balances
WHERE
  tenant_id = current_tenant_id()
  AND (
    $1::uuid IS NULL
    OR entity_id = $1
  )
  AND (
    $2::integer IS NULL
    OR fiscal_year = $2
  )
ORDER BY
  account_id;

-- name: UpsertAccountBalance :one
INSERT INTO
  finance_account_balances (
    tenant_id,
    entity_id,
    account_id,
    balance_date,
    fiscal_year,
    fiscal_period,
    opening_balance,
    period_debits,
    period_credits,
    closing_balance,
    created_by
  )
VALUES
  (
    current_tenant_id(),
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10
  ) ON CONFLICT (tenant_id, account_id, balance_date) DO
UPDATE
SET
  opening_balance = EXCLUDED.opening_balance,
  period_debits = EXCLUDED.period_debits,
  period_credits = EXCLUDED.period_credits,
  closing_balance = EXCLUDED.closing_balance
RETURNING
  *;

-- name: GetBalancesByFiscalPeriod :many
SELECT
  fab.*,
  fa.account_code,
  fa.account_name,
  fa.root_type,
  fa.normal_balance
FROM
  finance_account_balances fab
  JOIN finance_accounts fa ON fab.account_id = fa.id
WHERE
  fab.tenant_id = current_tenant_id()
  AND (
    $1::uuid IS NULL
    OR fab.entity_id = $1
  )
  AND fab.fiscal_year = $2
  AND fab.fiscal_period = $3
  AND (
    $4::text IS NULL
    OR fa.root_type = $4
  )
ORDER BY
  fa.account_code;

-- name: GetBalanceTrend :many
SELECT
  balance_date,
  opening_balance,
  period_debits,
  period_credits,
  closing_balance
FROM
  finance_account_balances
WHERE
  account_id = $1
  AND tenant_id = current_tenant_id()
  AND (
    $2::uuid IS NULL
    OR entity_id = $2
  )
  AND balance_date BETWEEN $3 AND $4
ORDER BY
  balance_date;

-- name: UpdatePeriodDebitsCredits :one
UPDATE
  finance_account_balances
SET
  period_debits = period_debits + $2,
  period_credits = period_credits + $3,
  closing_balance = opening_balance + period_debits + $2 - period_credits - $3
WHERE
  id = $1
  AND tenant_id = current_tenant_id()
RETURNING
  *;

-- name: RecalculateClosingBalance :one
UPDATE
  finance_account_balances
SET
  closing_balance = opening_balance + period_debits - period_credits
WHERE
  id = $1
  AND tenant_id = current_tenant_id()
RETURNING
  *;

-- name: GetBalancesForTrialBalance :many
SELECT
  fab.account_id,
  fa.account_code,
  fa.account_name,
  fa.root_type,
  fa.normal_balance,
  fab.closing_balance,
  fab.period_debits,
  fab.period_credits
FROM
  finance_account_balances fab
  JOIN finance_accounts fa ON fab.account_id = fa.id
WHERE
  fab.tenant_id = current_tenant_id()
  AND (
    $1::uuid IS NULL
    OR fab.entity_id = $1
  )
  AND fab.balance_date = $2
  AND fa.is_active = TRUE
ORDER BY
  fa.account_code;
