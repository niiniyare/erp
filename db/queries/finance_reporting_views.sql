-- =====================================================================
-- FINANCE MODULE - REPORTING QUERIES USING VIEWS
-- SQLC queries leveraging v_financial_statement_builder and other reporting views
-- =====================================================================
-- name: GetFinancialStatementBuilder :many
SELECT
  *
FROM
  v_financial_statement_builder
WHERE
  tenant_id = current_tenant_id()
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR entity_id = sqlc.narg('entity_id')::uuid
  )
  AND (
    sqlc.narg('statement_section')::text IS NULL
    OR statement_section = sqlc.narg('statement_section')
  )
ORDER BY
  statement_section,
  header_order,
  group_code;

-- name: GetBalanceSheetData :many
SELECT
  *
FROM
  v_financial_statement_builder
WHERE
  tenant_id = current_tenant_id()
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR entity_id = sqlc.narg('entity_id')::uuid
  )
  AND statement_section = 'Balance Sheet'
ORDER BY
  header_order,
  group_code;

-- name: GetIncomeStatementData :many
SELECT
  *
FROM
  v_financial_statement_builder
WHERE
  tenant_id = current_tenant_id()
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR entity_id = sqlc.narg('entity_id')::uuid
  )
  AND statement_section = 'Income Statement'
ORDER BY
  header_order,
  group_code;

-- name: GetFinancialStatementStructure :many
SELECT
  *
FROM
  v_financial_statement_structure
WHERE
  tenant_id = current_tenant_id()
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR entity_id = sqlc.narg('entity_id')::uuid
  )
  AND (
    sqlc.narg('statement_section')::text IS NULL
    OR financial_statement_section = sqlc.narg('statement_section')
  )
  AND (
    sqlc.narg('show_summary_only')::bool IS NULL
    OR (
      sqlc.narg('show_summary_only') = false
      OR show_in_summary = TRUE
    )
  )
ORDER BY
  financial_statement_section,
  statement_order,
  group_code;

-- name: GetGroupBalanceSummary :many
SELECT
  group_code,
  group_name,
  financial_statement_section,
  account_count,
  active_account_count,
  group_balance,
  CASE
    WHEN group_balance = 0 THEN 0
    ELSE ROUND(
      (
        active_account_count::decimal / NULLIF(account_count, 0)
      ) * 100,
      2
    )
  END AS active_percentage
FROM
  v_financial_statement_structure
WHERE
  tenant_id = current_tenant_id()
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR entity_id = sqlc.narg('entity_id')::uuid
  )
  AND (
    sqlc.narg('root_type')::text IS NULL
    OR root_type = sqlc.narg('root_type')
  )
  AND account_count > 0
ORDER BY
  financial_statement_section,
  statement_order,
  group_code;

-- =====================================================================
-- ACCOUNT HIERARCHY VIEW QUERIES
-- =====================================================================
-- -- name: GetAccountHierarchyView :many
-- SELECT
--   *
-- FROM
--   v_finance_accounts_hierarchy
-- WHERE
--   tenant_id = current_tenant_id()
--   AND (
--     sqlc.narg('entity_id')::uuid IS NULL
--     OR tenant_id = current_tenant_id()
--   )
--   AND (
--     sqlc.narg('root_type')::text IS NULL
--     OR root_type = sqlc.narg('root_type')
--   )
--   AND (
--     sqlc.narg('parent_account_id')::uuid IS NULL
--     OR (
--       sqlc.narg('parent_account_id') IS NULL
--       AND parent_account_id IS NULL
--     )
--     OR full_path LIKE '%' || sqlc.narg('parent_account_id')::text || '%'
--   )
-- ORDER BY
--   full_path;

-- name: GetAccountHierarchyByLevel :many
SELECT
  *
FROM
  v_finance_accounts_hierarchy
WHERE
  tenant_id = current_tenant_id()
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR entity_id = sqlc.narg('entity_id')::uuid
  )
  AND LEVEL <= sqlc.arg('max_level')
  AND (
    sqlc.narg('root_type')::text IS NULL
    OR root_type = sqlc.narg('root_type')
  )
ORDER BY
  full_path;

-- name: GetRootAccountsView :many
SELECT
  *
FROM
  v_finance_accounts_hierarchy
WHERE
  tenant_id = current_tenant_id()
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR entity_id = sqlc.narg('entity_id')::uuid
  )
  AND LEVEL = 1
  AND (
    sqlc.narg('root_type')::text IS NULL
    OR root_type = sqlc.narg('root_type')
  )
ORDER BY
  account_code;

-- name: GetAccountChildren :many
SELECT
  *
FROM
  v_finance_accounts_hierarchy
WHERE
  tenant_id = current_tenant_id()
  AND parent_account_id = sqlc.arg('parent_account_id')
ORDER BY
  account_code;

-- name: GetAccountWithChildren :many
-- Prefix-safe subtree fetch: anchors match at the start of the path segment
-- so account_code "1000" never matches "10001" or "21000".
-- Pattern: exact match OR path starts with "code." (child separator).
SELECT
  *
FROM
  v_finance_accounts_hierarchy h
WHERE
  h.tenant_id = current_tenant_id()
  AND (
    h.id = sqlc.arg('account_id')
    OR h.full_path LIKE (
      SELECT account_code
      FROM   finance_accounts
      WHERE  id        = sqlc.arg('account_id')
        AND  tenant_id = current_tenant_id()
    ) || '.%'
    OR h.full_path = (
      SELECT account_code
      FROM   finance_accounts
      WHERE  id        = sqlc.arg('account_id')
        AND  tenant_id = current_tenant_id()
    )
  )
ORDER BY
  h.LEVEL,
  h.account_code;

-- =====================================================================
-- ACCOUNT ACTIVITY VIEW QUERIES
-- =====================================================================
-- name: GetAccountActivity :many
SELECT
  *
FROM
  v_finance_account_activity
WHERE
  tenant_id = current_tenant_id()
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR entity_id = sqlc.narg('entity_id')::uuid
  )
  AND (
    sqlc.narg('min_balance')::decimal IS NULL
    OR ABS(current_balance) >= sqlc.narg('min_balance')
  )
  AND (
    sqlc.narg('min_entries')::int IS NULL
    OR total_entries >= sqlc.narg('min_entries')
  )
ORDER BY
  total_entries DESC,
  ABS(current_balance) DESC;

-- name: GetActiveAccounts :many
SELECT
  *
FROM
  v_finance_account_activity
WHERE
  tenant_id = current_tenant_id()
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR entity_id = sqlc.narg('entity_id')::uuid
  )
  AND entries_last_30_days > 0
ORDER BY
  entries_last_30_days DESC;

-- name: GetInactiveAccounts :many
SELECT
  account_id,
  account_code,
  account_name,
  current_balance,
  last_transaction_date,
  total_entries
FROM
  v_finance_account_activity
WHERE
  tenant_id = current_tenant_id()
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR entity_id = sqlc.narg('entity_id')::uuid
  )
  AND entries_last_30_days = 0
  AND (
    sqlc.narg('min_days_inactive')::int IS NULL
    OR last_transaction_date < CURRENT_DATE - INTERVAL '1 day' * sqlc.narg('min_days_inactive')
  )
ORDER BY
  last_transaction_date ASC NULLS FIRST;

-- name: GetHighActivityAccounts :many
SELECT
  account_id,
  account_code,
  account_name,
  current_balance,
  entries_last_30_days,
  debits_last_30_days,
  credits_last_30_days,
  (debits_last_30_days + credits_last_30_days) AS total_activity_30_days
FROM
  v_finance_account_activity
WHERE
  tenant_id = current_tenant_id()
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR entity_id = sqlc.narg('entity_id')::uuid
  )
  AND entries_last_30_days >= sqlc.arg('min_entries')
ORDER BY
  entries_last_30_days DESC,
  (debits_last_30_days + credits_last_30_days) DESC;

-- =====================================================================
-- TRANSACTION SUMMARY VIEW QUERIES
-- =====================================================================
-- name: GetTransactionSummary :many
SELECT
  *
FROM
  v_finance_transaction_summary
WHERE
  tenant_id = current_tenant_id()
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR entity_id = sqlc.narg('entity_id')::uuid
  )
  AND (
    sqlc.narg('from_date')::date IS NULL
    OR transaction_date >= sqlc.narg('from_date')
  )
  AND (
    sqlc.narg('to_date')::date IS NULL
    OR transaction_date <= sqlc.narg('to_date')
  )
  AND (
    sqlc.narg('transaction_status')::text IS NULL
    OR transaction_status = sqlc.narg('transaction_status')
  )
ORDER BY
  transaction_date DESC,
  transaction_number DESC
LIMIT
  sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetUnreconciledTransactions :many
SELECT
  *
FROM
  v_finance_transaction_summary
WHERE
  tenant_id = current_tenant_id()
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR entity_id = sqlc.narg('entity_id')::uuid
  )
  AND all_entries_reconciled = false
  AND transaction_status = 'POSTED'
ORDER BY
  transaction_date ASC;

-- name: GetTransactionsByAccount :many
-- account_codes is a comma-separated list (e.g. "1000,2000,3100").
-- Use word-boundary anchors via regex to avoid "1000" matching "10001":
--   match at string start, after a comma, or as exact full string.
SELECT
  ts.*
FROM
  v_finance_transaction_summary ts
WHERE
  ts.tenant_id = current_tenant_id()
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR ts.entity_id = sqlc.narg('entity_id')::uuid
  )
  AND (
    ts.account_codes = sqlc.arg('account_code')
    OR ts.account_codes LIKE sqlc.arg('account_code') || ',%'
    OR ts.account_codes LIKE '%,' || sqlc.arg('account_code') || ',%'
    OR ts.account_codes LIKE '%,' || sqlc.arg('account_code')
  )
  AND (
    sqlc.narg('from_date')::date IS NULL
    OR ts.transaction_date >= sqlc.narg('from_date')
  )
  AND (
    sqlc.narg('to_date')::date IS NULL
    OR ts.transaction_date <= sqlc.narg('to_date')
  )
ORDER BY
  ts.transaction_date DESC;

-- =====================================================================
-- PERFORMANCE AND ANALYTICS QUERIES
-- =====================================================================
-- name: GetAccountUtilizationStats :many
SELECT
  v.group_name,
  v.group_category,
  COUNT(v.account_id) AS total_accounts,
  COUNT(
    CASE
      WHEN v.is_active THEN 1
    END
  ) AS active_accounts,
  COUNT(
    CASE
      WHEN v.current_balance != 0 THEN 1
    END
  ) AS accounts_with_balance,
  COUNT(
    CASE
      WHEN act.entries_last_30_days > 0 THEN 1
    END
  ) AS active_last_30_days,
  ROUND(AVG(ABS(v.current_balance)), 2) AS avg_balance,
  SUM(v.current_balance) AS total_balance
FROM
  v_chart_of_accounts_complete v
  LEFT JOIN v_finance_account_activity act ON v.account_id = act.account_id
WHERE
  v.tenant_id = current_tenant_id()
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR v.tenant_id = current_tenant_id()
  )
  AND v.group_name IS NOT NULL
GROUP BY
  v.group_name,
  v.group_category
ORDER BY
  total_accounts DESC;

-- name: GetTopAccountsByBalance :many
SELECT
  account_id,
  account_code,
  account_name,
  current_balance,
  group_name,
  statement_section,
  ABS(current_balance) AS abs_balance
FROM
  v_chart_of_accounts_complete
WHERE
  tenant_id = current_tenant_id()
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR entity_id = sqlc.narg('entity_id')::uuid
  )
  AND is_active = TRUE
  AND current_balance != 0
  AND (
    sqlc.narg('root_type')::text IS NULL
    OR root_type = sqlc.narg('root_type')
  )
ORDER BY
  ABS(current_balance) DESC
LIMIT
  sqlc.arg('limit');

-- name: GetAccountsRequiringAttention :many
SELECT
  v.account_id,
  v.account_code,
  v.account_name,
  v.current_balance,
  v.group_name,
  act.last_transaction_date,
  act.entries_last_30_days,
  CASE
    WHEN v.current_balance != 0
    AND act.entries_last_30_days = 0 THEN 'Stale Balance'
    WHEN v.current_balance = 0
    AND act.entries_last_30_days > 10 THEN 'High Activity, Zero Balance'
    WHEN ABS(v.current_balance) > 1000000 THEN 'High Balance'
    ELSE 'Other'
  END AS attention_reason
FROM
  v_chart_of_accounts_complete v
  LEFT JOIN v_finance_account_activity act ON v.account_id = act.account_id
WHERE
  v.tenant_id = current_tenant_id()
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR v.tenant_id = current_tenant_id()
  )
  AND v.is_active = TRUE
  AND (
    (
      v.current_balance != 0
      AND act.entries_last_30_days = 0
    )
    OR (
      v.current_balance = 0
      AND act.entries_last_30_days > 10
    )
    OR (ABS(v.current_balance) > 1000000)
  )
ORDER BY
  ABS(v.current_balance) DESC;
