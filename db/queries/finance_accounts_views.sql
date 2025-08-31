-- =====================================================================
-- FINANCE MODULE - ENHANCED QUERIES USING VIEWS
-- SQLC queries leveraging v_finance_accounts_with_groups and v_chart_of_accounts_complete
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

-- name: ListAccountsWithGroups :many
SELECT
  *
FROM
  v_finance_accounts_with_groups
WHERE
  tenant_id = current_tenant_id()
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR tenant_id = current_tenant_id()
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
  )
ORDER BY
  effective_display_order,
  account_code
LIMIT
  sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountAccountsWithGroups :one
SELECT
  COUNT(*)
FROM
  v_finance_accounts_with_groups
WHERE
  tenant_id = current_tenant_id()
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR tenant_id = current_tenant_id()
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

-- name: GetAccountsByGroup :many
SELECT
  *
FROM
  v_finance_accounts_with_groups
WHERE
  tenant_id = current_tenant_id()
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR tenant_id = current_tenant_id()
  )
  AND group_code = sqlc.arg('group_code')
  AND is_active = TRUE
ORDER BY
  account_code;

-- name: GetAccountsByFinancialStatement :many
SELECT
  *
FROM
  v_finance_accounts_with_groups
WHERE
  tenant_id = current_tenant_id()
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR tenant_id = current_tenant_id()
  )
  AND financial_statement_section = sqlc.arg('statement_section')
  AND show_in_reports = TRUE
  AND is_active = TRUE
ORDER BY
  effective_display_order,
  account_code;

-- name: SearchAccountsWithGroups :many
SELECT
  *
FROM
  v_finance_accounts_with_groups
WHERE
  tenant_id = current_tenant_id()
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR tenant_id = current_tenant_id()
  )
  AND (
    account_code ILIKE '%' || sqlc.arg('search_term') || '%'
    OR account_name ILIKE '%' || sqlc.arg('search_term') || '%'
    OR account_description ILIKE '%' || sqlc.arg('search_term') || '%'
    OR group_name ILIKE '%' || sqlc.arg('search_term') || '%'
  )
ORDER BY
  CASE
    WHEN account_code ILIKE sqlc.arg('search_term') || '%' THEN 1
    ELSE 2
  END,
  effective_display_order,
  account_code
LIMIT
  sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetLeafAccountsOnly :many
SELECT
  *
FROM
  v_finance_accounts_with_groups
WHERE
  tenant_id = current_tenant_id()
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR tenant_id = current_tenant_id()
  )
  AND is_leaf_account = TRUE
  AND is_active = TRUE
  AND (
    sqlc.narg('root_type')::text IS NULL
    OR root_type = sqlc.narg('root_type')
  )
ORDER BY
  effective_display_order,
  account_code;

-- =====================================================================
-- COMPLETE CHART OF ACCOUNTS VIEW QUERIES
-- =====================================================================
-- name: GetCompleteChartOfAccounts :many
SELECT
  *
FROM
  v_chart_of_accounts_complete
WHERE
  tenant_id = current_tenant_id()
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR tenant_id = current_tenant_id()
  )
  AND (
    sqlc.narg('statement_section')::text IS NULL
    OR statement_section = sqlc.narg('statement_section')
  )
  AND (
    sqlc.narg('include_inactive')::bool IS NULL
    OR (
      sqlc.narg('include_inactive') = TRUE
      OR is_active = TRUE
    )
  )
  AND (
    sqlc.narg('include_in_reports')::bool IS NULL
    OR (
      sqlc.narg('include_in_reports') = false
      OR include_in_reports = TRUE
    )
  )
ORDER BY
  display_order,
  account_code;

-- name: GetAccountForReporting :one
SELECT
  *
FROM
  v_chart_of_accounts_complete
WHERE
  account_id = sqlc.arg('account_id')
  AND tenant_id = current_tenant_id();

-- name: GetAccountsByStatementSection :many
SELECT
  *
FROM
  v_chart_of_accounts_complete
WHERE
  tenant_id = current_tenant_id()
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR tenant_id = current_tenant_id()
  )
  AND statement_section = sqlc.arg('statement_section')
  AND include_in_reports = TRUE
  AND is_active = TRUE
ORDER BY
  display_order,
  account_code;

-- name: GetAccountsForTrialBalance :many
SELECT
  account_id,
  account_code,
  account_name,
  root_type,
  normal_balance,
  current_balance,
  group_name,
  statement_section
FROM
  v_chart_of_accounts_complete
WHERE
  tenant_id = current_tenant_id()
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR tenant_id = current_tenant_id()
  )
  AND is_active = TRUE
  AND include_in_reports = TRUE
  AND current_balance != 0
ORDER BY
  display_order,
  account_code;

-- name: GetAccountsByGroup :many
SELECT
  *
FROM
  v_chart_of_accounts_complete
WHERE
  tenant_id = current_tenant_id()
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR tenant_id = current_tenant_id()
  )
  AND group_code = sqlc.arg('group_code')
  AND is_active = TRUE
ORDER BY
  display_order,
  account_code;

-- name: GetAccountsByHeader :many
SELECT
  *
FROM
  v_chart_of_accounts_complete
WHERE
  tenant_id = current_tenant_id()
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR tenant_id = current_tenant_id()
  )
  AND header_code = sqlc.arg('header_code')
  AND is_active = TRUE
ORDER BY
  display_order,
  account_code;

-- name: GetAccountsWithBalances :many
SELECT
  account_id,
  account_code,
  account_name,
  current_balance,
  normal_balance,
  group_name,
  statement_section,
  cash_flow_classification
FROM
  v_chart_of_accounts_complete
WHERE
  tenant_id = current_tenant_id()
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR tenant_id = current_tenant_id()
  )
  AND is_active = TRUE
  AND (
    sqlc.narg('non_zero_only')::bool IS NULL
    OR (
      sqlc.narg('non_zero_only') = false
      OR current_balance != 0
    )
  )
  AND (
    sqlc.narg('root_type')::text IS NULL
    OR root_type = sqlc.narg('root_type')
  )
ORDER BY
  display_order,
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
    coa.account_code::text AS full_path,
    coa.account_name::text AS full_name
  FROM
    v_chart_of_accounts_complete coa
  WHERE
    coa.parent_account_id IS NULL
    AND coa.tenant_id = current_tenant_id()
    AND (
      sqlc.narg('entity_id')::uuid IS NULL
      OR coa.tenant_id = current_tenant_id()
    )
  UNION
  ALL
  -- Child accounts
  SELECT
    coa.*,
    ah.hierarchy_level + 1,
    (ah.full_path || '.' || coa.account_code)::text,
    (ah.full_name || ' > ' || coa.account_name)::text
  FROM
    v_chart_of_accounts_complete coa
    JOIN account_hierarchy ah ON coa.parent_account_id = ah.account_id
  WHERE
    coa.tenant_id = current_tenant_id()
    AND ah.hierarchy_level < 10 -- Prevent infinite recursion
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
  header_order,
  cash_flow_classification
FROM
  v_chart_of_accounts_complete
WHERE
  tenant_id = current_tenant_id()
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR tenant_id = current_tenant_id()
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
  header_order,
  display_order,
  account_code;

-- name: GetCashFlowAccounts :many
SELECT
  account_id,
  account_code,
  account_name,
  current_balance,
  cash_flow_classification,
  group_name
FROM
  v_chart_of_accounts_complete
WHERE
  tenant_id = current_tenant_id()
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR tenant_id = current_tenant_id()
  )
  AND cash_flow_classification IS NOT NULL
  AND is_active = TRUE
  AND include_in_reports = TRUE
ORDER BY
  CASE
    cash_flow_classification
    WHEN 'OPERATING' THEN 1
    WHEN 'INVESTING' THEN 2
    WHEN 'FINANCING' THEN 3
    ELSE 4
  END,
  display_order,
  account_code;

-- name: GetAccountSummaryByGroup :many
SELECT
  group_code,
  group_name,
  group_category,
  statement_section,
  COUNT(account_id) AS account_count,
  COUNT(
    CASE
      WHEN is_active THEN 1
    END
  ) AS active_account_count,
  SUM(current_balance) AS total_balance,
  SUM(
    CASE
      WHEN is_active THEN current_balance
      ELSE 0
    END
  ) AS active_balance
FROM
  v_chart_of_accounts_complete
WHERE
  tenant_id = current_tenant_id()
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR tenant_id = current_tenant_id()
  )
  AND group_code IS NOT NULL
GROUP BY
  group_code,
  group_name,
  group_category,
  statement_section
ORDER BY
  statement_section,
  group_code;
