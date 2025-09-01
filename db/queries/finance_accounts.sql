-- =====================================================================
-- FINANCE MODULE - CHART OF ACCOUNTS QUERIES
-- SQLC queries for chart of accounts with proper tenant isolation
-- Updated with proper sqlc.narg and sqlc.arg usage
-- =====================================================================
-- name: CreateAccount :one
INSERT INTO
  finance_accounts (
    tenant_id,
    entity_id,
    account_code,
    account_name,
    account_description,
    account_group_id,
    account_header_id,
    parent_account_id,
    account_level,
    account_path,
    account_category,
    sub_category,
    display_order,
    show_in_reports,
    consolidation_account,
    cash_flow_type,
    root_type,
    account_type,
    account_subtype,
    normal_balance,
    is_control_account,
    control_account_id,
    currency_code,
    is_multi_currency,
    currency_revaluation_required,
    is_active,
    is_system_account,
    allow_manual_entries,
    require_reference,
    current_balance,
    ytd_balance,
    last_transaction_date,
    financial_statement_line,
    report_order,
    is_budgetable,
    budget_variance_threshold,
    version,
    last_validation_run,
    validation_status,
    validation_errors,
    account_attributes,
    has_children,
    is_leaf_account,
    created_by
  )
VALUES
  (
    current_tenant_id(),
    sqlc.narg('entity_id'),
    sqlc.arg('account_code'),
    sqlc.arg('account_name'),
    sqlc.narg('account_description'),
    sqlc.narg('account_group_id'),
    sqlc.narg('account_header_id'),
    sqlc.narg('parent_account_id'),
    sqlc.narg('account_level'),
    sqlc.narg('account_path'),
    sqlc.narg('account_category'),
    sqlc.narg('sub_category'),
    sqlc.narg('display_order'),
    sqlc.narg('show_in_reports'),
    sqlc.narg('consolidation_account'),
    sqlc.narg('cash_flow_type'),
    sqlc.arg('root_type'),
    sqlc.arg('account_type'),
    sqlc.narg('account_subtype'),
    sqlc.arg('normal_balance'),
    sqlc.narg('is_control_account'),
    sqlc.narg('control_account_id'),
    sqlc.narg('currency_code'),
    sqlc.narg('is_multi_currency'),
    sqlc.narg('currency_revaluation_required'),
    sqlc.narg('is_active'),
    sqlc.narg('is_system_account'),
    sqlc.narg('allow_manual_entries'),
    sqlc.narg('require_reference'),
    sqlc.narg('current_balance'),
    sqlc.narg('ytd_balance'),
    sqlc.narg('last_transaction_date'),
    sqlc.narg('financial_statement_line'),
    sqlc.narg('report_order'),
    sqlc.narg('is_budgetable'),
    sqlc.narg('budget_variance_threshold'),
    sqlc.narg('version'),
    sqlc.narg('last_validation_run'),
    sqlc.narg('validation_status'),
    sqlc.narg('validation_errors'),
    sqlc.narg('account_attributes'),
    sqlc.narg('has_children'),
    sqlc.narg('is_leaf_account'),
    sqlc.narg('created_by')
  )
RETURNING
  *;

-- name: GetAccountByID :one
SELECT
  *
FROM
  finance_accounts
WHERE
  id = sqlc.arg('account_id')
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: GetAccountWithGroupsByID :one
SELECT
  *
FROM
  v_finance_accounts_with_groups
WHERE
  id = sqlc.arg('account_id')
  AND tenant_id = current_tenant_id();

-- name: GetAccountByCode :one
SELECT
  *
FROM
  finance_accounts
WHERE
  account_code = sqlc.arg('account_code')
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: GetAccountWithGroupsByCode :one
SELECT
  *
FROM
  v_finance_accounts_with_groups
WHERE
  account_code = sqlc.arg('account_code')
  AND tenant_id = current_tenant_id();

-- name: ListAccounts :many
SELECT
  *
FROM
  finance_accounts
WHERE
  tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR entity_id = sqlc.narg('entity_id')::uuid
  )
  AND (
    sqlc.narg('account_type')::VARCHAR IS NULL
    OR account_type = sqlc.narg('account_type')::VARCHAR
  )
  AND (
    sqlc.narg('root_type')::VARCHAR IS NULL
    OR root_type = sqlc.narg('root_type')::VARCHAR
  )
  AND (
    sqlc.narg('is_active')::bool IS NULL
    OR is_active = sqlc.narg('is_active')::bool
  )
ORDER BY
  account_code ASC
LIMIT
  sqlc.arg('limit') OFFSET sqlc.arg('offset');

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
    sqlc.narg('group_code')::text IS NULL
    OR group_code = sqlc.narg('group_code')
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

-- name: CountAccounts :one
SELECT
  COUNT(*)
FROM
  finance_accounts
WHERE
  tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR entity_id = sqlc.narg('entity_id')::uuid
  )
  AND (
    sqlc.narg('account_type')::VARCHAR IS NULL
    OR account_type = sqlc.narg('account_type')::VARCHAR
  )
  AND (
    sqlc.narg('root_type')::VARCHAR IS NULL
    OR root_type = sqlc.narg('root_type')::VARCHAR
  )
  AND (
    sqlc.narg('is_active')::bool IS NULL
    OR is_active = sqlc.narg('is_active')::bool
  );

-- name: ListAccountsByParent :many
SELECT
  *
FROM
  finance_accounts
WHERE
  parent_account_id = sqlc.narg('parent_account_id')
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
ORDER BY
  account_code ASC;

-- name: GetAccountHierarchy :many
SELECT
  *
FROM
  finance_accounts
WHERE
  tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND account_path LIKE sqlc.arg('account_path_prefix') || '%'
ORDER BY
  account_path ASC;

-- name: GetRootAccounts :many
SELECT
  *
FROM
  finance_accounts
WHERE
  parent_account_id IS NULL
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
ORDER BY
  account_code ASC;

-- name: UpdateAccount :one
UPDATE
  finance_accounts
SET
  account_name = COALESCE(sqlc.narg('account_name'), account_name),
  account_description = COALESCE(
    sqlc.narg('account_description'),
    account_description
  ),
  account_group_id = COALESCE(sqlc.narg('account_group_id'), account_group_id),
  account_header_id = COALESCE(
    sqlc.narg('account_header_id'),
    account_header_id
  ),
  account_level = COALESCE(sqlc.narg('account_level'), account_level),
  account_path = COALESCE(sqlc.narg('account_path'), account_path),
  account_category = COALESCE(sqlc.narg('account_category'), account_category),
  sub_category = COALESCE(sqlc.narg('sub_category'), sub_category),
  display_order = COALESCE(sqlc.narg('display_order'), display_order),
  show_in_reports = COALESCE(sqlc.narg('show_in_reports'), show_in_reports),
  consolidation_account = COALESCE(
    sqlc.narg('consolidation_account'),
    consolidation_account
  ),
  cash_flow_type = COALESCE(sqlc.narg('cash_flow_type'), cash_flow_type),
  account_type = COALESCE(sqlc.narg('account_type'), account_type),
  account_subtype = COALESCE(sqlc.narg('account_subtype'), account_subtype),
  is_active = COALESCE(sqlc.narg('is_active'), is_active),
  allow_manual_entries = COALESCE(
    sqlc.narg('allow_manual_entries'),
    allow_manual_entries
  ),
  require_reference = COALESCE(
    sqlc.narg('require_reference'),
    require_reference
  ),
  financial_statement_line = COALESCE(
    sqlc.narg('financial_statement_line'),
    financial_statement_line
  ),
  report_order = COALESCE(sqlc.narg('report_order'), report_order),
  is_budgetable = COALESCE(sqlc.narg('is_budgetable'), is_budgetable),
  budget_variance_threshold = COALESCE(
    sqlc.narg('budget_variance_threshold'),
    budget_variance_threshold
  ),
  validation_status = COALESCE(
    sqlc.narg('validation_status'),
    validation_status
  ),
  validation_errors = COALESCE(
    sqlc.narg('validation_errors'),
    validation_errors
  ),
  account_attributes = COALESCE(
    sqlc.narg('account_attributes'),
    account_attributes
  ),
  has_children = COALESCE(sqlc.narg('has_children'), has_children),
  is_leaf_account = COALESCE(sqlc.narg('is_leaf_account'), is_leaf_account),
  updated_at = NOW(),
  updated_by = sqlc.narg('updated_by')
WHERE
  id = sqlc.arg('account_id')
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
RETURNING
  *;

-- name: UpdateAccountCurrentBalance :exec
UPDATE
  finance_accounts
SET
  current_balance = sqlc.arg('current_balance'),
  ytd_balance = sqlc.arg('ytd_balance'),
  last_transaction_date = sqlc.arg('last_transaction_date'),
  updated_at = NOW()
WHERE
  id = sqlc.arg('account_id')
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: SoftDeleteAccount :exec
UPDATE
  finance_accounts
SET
  deleted_at = NOW(),
  updated_at = NOW(),
  updated_by = sqlc.narg('updated_by')
WHERE
  id = sqlc.arg('account_id')
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: RestoreAccount :exec
UPDATE
  finance_accounts
SET
  deleted_at = NULL,
  updated_at = NOW(),
  updated_by = sqlc.narg('updated_by')
WHERE
  id = sqlc.arg('account_id')
  AND tenant_id = current_tenant_id();

-- name: GetAccountsForFinancialStatements :many
SELECT
  a.*,
  COALESCE(
    SUM(
      CASE
        WHEN te.debit_amount > 0 THEN te.debit_amount
        ELSE - te.credit_amount
      END
    ),
    0
  ) AS calculated_balance
FROM
  finance_accounts a
  LEFT JOIN finance_transaction_entries te ON a.id = te.account_id
  LEFT JOIN finance_transactions t ON te.transaction_id = t.id
WHERE
  a.tenant_id = current_tenant_id()
  AND a.deleted_at IS NULL
  AND a.is_active = TRUE
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR a.entity_id = sqlc.narg('entity_id')::uuid
  )
  AND (
    sqlc.narg('financial_statement_line')::text IS NULL
    OR a.financial_statement_line = sqlc.narg('financial_statement_line')
  )
  AND (
    t.transaction_status = 'POSTED'
    OR t.id IS NULL
  )
  AND (
    t.posting_date <= sqlc.arg('posting_date')
    OR t.posting_date IS NULL
  )
GROUP BY
  a.id
ORDER BY
  a.report_order ASC,
  a.account_code ASC;

-- name: SearchAccounts :many
SELECT
  *
FROM
  finance_accounts
WHERE
  tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR entity_id = sqlc.narg('entity_id')::uuid
  )
  AND (
    account_code ILIKE '%' || sqlc.arg('search_term') || '%'
    OR account_name ILIKE '%' || sqlc.arg('search_term') || '%'
    OR account_description ILIKE '%' || sqlc.arg('search_term') || '%'
  )
ORDER BY
  CASE
    WHEN account_code ILIKE sqlc.arg('search_term') || '%' THEN 1
    ELSE 2
  END,
  account_code ASC
LIMIT
  sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: SearchAccountsWithGroupInfo :many
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

-- name: GetAccountsWithNonZeroBalance :many
SELECT
  *
FROM
  finance_accounts
WHERE
  tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR entity_id = sqlc.narg('entity_id')::uuid
  )
  AND current_balance != 0
ORDER BY
  ABS(current_balance) DESC;

-- name: GetControlAccounts :many
SELECT
  *
FROM
  finance_accounts
WHERE
  tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND (
    sqlc.narg('entity_id')::uuid IS NULL
    OR entity_id = sqlc.narg('entity_id')::uuid
  )
  AND is_control_account = TRUE
ORDER BY
  account_code ASC;

-- name: GetAccountsByEntity :many
SELECT
  *
FROM
  finance_accounts
WHERE
  entity_id = sqlc.narg('entity_id')
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
ORDER BY
  account_code ASC;

-- name: ValidateAccountHierarchy :one
SELECT
  CASE
    WHEN EXISTS(
      SELECT
        1
      FROM
        finance_accounts
      WHERE
        parent_account_id = sqlc.narg('parent_account_id')
        AND id = sqlc.narg('parent_account_id')
    ) THEN false -- Self reference check
    ELSE TRUE
  END AS is_valid_hierarchy;

-- =====================================================================
-- ENHANCED QUERIES USING v_chart_of_accounts_complete VIEW
-- =====================================================================

-- name: GetChartOfAccountsComplete :many
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

-- name: GetAccountReportingInfo :one
SELECT
  *
FROM
  v_chart_of_accounts_complete
WHERE
  account_id = sqlc.arg('account_id')
  AND tenant_id = current_tenant_id();

-- name: GetAccountsByStatement :many
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

-- name: GetTrialBalanceData :many
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
  AND (
    sqlc.narg('non_zero_only')::bool IS NULL
    OR (
      sqlc.narg('non_zero_only') = false
      OR current_balance != 0
    )
  )
ORDER BY
  display_order,
  account_code;

-- name: GetAccountsByGroupCode :many
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

-- name: GetAccountsByHeaderCode :many
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

-- name: GetAccountBalancesList :many
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

-- name: GetLeafAccountsWithGroups :many
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

-- name: GetCashFlowAccountsList :many
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

-- name: GetAccountGroupSummary :many
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
