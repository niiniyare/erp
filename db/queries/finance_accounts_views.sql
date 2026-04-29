-- =====================================================================
-- FINANCE MODULE - ENHANCED QUERIES USING VIEWS (NON-DUPLICATE)
-- SQLC queries leveraging v_finance_accounts_with_groups and v_chart_of_accounts_complete
-- These are unique queries not present in finance_accounts.sql
-- =====================================================================

-- name: GetAccountWithGroupInfo :one
SELECT
  *
FROM
  v_finance_accounts_with_groups
WHERE
  id = sqlc.arg('account_id')
  AND tenant_id = current_tenant_id();

-- name: GetAccountByCodeWithGroups :one
SELECT
  *
FROM
  v_finance_accounts_with_groups
WHERE
  account_code = sqlc.arg('account_code')
  AND tenant_id = current_tenant_id();

-- name: CountAccountsWithGroups :one
SELECT
  COUNT(*)
FROM
  v_finance_accounts_with_groups
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
  AND (
    sqlc.narg('account_type')::text IS NULL
    OR account_type = sqlc.narg('account_type')
  )
  AND (
    sqlc.narg('group_category')::text IS NULL
    OR group_category = sqlc.narg('group_category')
  )
  AND (
    sqlc.narg('is_active')::bool IS NULL
    OR is_active = sqlc.narg('is_active')
  )
  AND (
    sqlc.narg('is_leaf_only')::bool IS NULL
    OR (
      sqlc.narg('is_leaf_only') = false
      OR is_leaf_account = TRUE
    )
  );

-- name: GetAccountsByFinancialStatement :many
SELECT
  *
FROM
  v_finance_accounts_with_groups
WHERE
  tenant_id = current_tenant_id()
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR entity_id = sqlc.narg('entity_id')::uuid
  )
  AND financial_statement_section = sqlc.arg('statement_section')
  AND show_in_reports = TRUE
  AND is_active = TRUE
ORDER BY
  effective_display_order,
  account_code;

-- =====================================================================
-- HIERARCHY AND REPORTING QUERIES
-- =====================================================================
-- name: GetAccountHierarchyComplete :many
WITH RECURSIVE account_hierarchy AS (
  -- Root accounts
  SELECT
    coa.*,
    1 AS hierarchy_level,
    FALSE AS truncated,
    coa.account_code::text AS full_path,
    coa.account_name::text AS full_name
  FROM
    v_chart_of_accounts_complete coa
  WHERE
    coa.parent_account_id IS NULL
    AND coa.tenant_id = current_tenant_id()
    AND (
      sqlc.narg('entity_id')::uuid IS NULL
      OR coa.entity_id = sqlc.narg('entity_id')::uuid
    )
  UNION ALL
  -- Child accounts; mark rows at depth limit so callers can detect truncation
  SELECT
    coa.*,
    ah.hierarchy_level + 1,
    (ah.hierarchy_level + 1 >= 20) AS truncated,
    (ah.full_path || '.' || coa.account_code)::text,
    (ah.full_name || ' > ' || coa.account_name)::text
  FROM
    v_chart_of_accounts_complete coa
    JOIN account_hierarchy ah ON coa.parent_account_id = ah.account_id
  WHERE
    coa.tenant_id = current_tenant_id()
    AND ah.hierarchy_level < 20  -- Raised from 10; flag truncation instead of silently dropping
)
SELECT
  *
FROM
  account_hierarchy
WHERE
  (
    sqlc.narg('root_type')::text IS NULL
    OR root_type = sqlc.narg('root_type')
  )
  AND (
    sqlc.narg('include_inactive')::bool IS NULL
    OR (
      sqlc.narg('include_inactive') = TRUE
      OR is_active = TRUE
    )
  )
ORDER BY
  full_path;

-- name: GetFinancialStatementData :many
SELECT
  statement_section,
  header_name,
  group_name,
  account_code,
  account_name,
  current_balance,
  normal_balance,
  display_order,
  cash_flow_classification
FROM
  v_chart_of_accounts_complete
WHERE
  tenant_id = current_tenant_id()
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR entity_id = sqlc.narg('entity_id')::uuid
  )
  AND include_in_reports = TRUE
  AND is_active = TRUE
  AND (
    sqlc.narg('statement_section')::text IS NULL
    OR statement_section = sqlc.narg('statement_section')
  )
ORDER BY
  CASE
    statement_section
    WHEN 'Balance Sheet' THEN 1
    WHEN 'Income Statement' THEN 2
    ELSE 3
  END,
  display_order,
  account_code;