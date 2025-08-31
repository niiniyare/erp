-- =====================================================================
-- FINANCE MODULE - CHART OF ACCOUNTS QUERIES
-- SQLC queries for chart of accounts with proper tenant isolation
-- Updated with proper sqlc.narg and sqlc.arg usage
-- =====================================================================

-- name: CreateAccount :one
INSERT INTO finance_accounts (
    tenant_id,
    entity_id,
    account_code,
    account_name,
    account_description,
    parent_account_id,
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
    financial_statement_line,
    report_order,
    is_budgetable,
    budget_variance_threshold,
    account_attributes,
    created_by
) VALUES (
    current_tenant_id(),
    sqlc.narg('entity_id'), 
    sqlc.arg('account_code'), 
    sqlc.arg('account_name'), 
    sqlc.arg('account_description'), 
    sqlc.narg('parent_account_id'), 
    sqlc.arg('root_type'), 
    sqlc.arg('account_type'), 
    sqlc.narg('account_subtype'), 
    sqlc.arg('normal_balance'), 
    sqlc.arg('is_control_account'), 
    sqlc.narg('control_account_id'), 
    sqlc.narg('currency_code'), 
    sqlc.narg('is_multi_currency'), 
    sqlc.narg('currency_revaluation_required'), 
    sqlc.arg('is_active'), 
    sqlc.arg('is_system_account'), 
    sqlc.arg('allow_manual_entries'), 
    sqlc.arg('require_reference'), 
    sqlc.narg('financial_statement_line'), 
    sqlc.narg('report_order'), 
    sqlc.narg('is_budgetable'), 
    sqlc.arg('budget_variance_threshold'), 
    sqlc.arg('account_attributes'), 
    sqlc.narg('created_by')
) RETURNING *;

-- name: GetAccountByID :one
SELECT * FROM finance_accounts
WHERE id = sqlc.arg('account_id') 
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: GetAccountByCode :one
SELECT * FROM finance_accounts
WHERE account_code = sqlc.arg('account_code') 
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: ListAccounts :many
SELECT * 
FROM finance_accounts
WHERE tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND (sqlc.narg('account_type')::account_type_enum IS NULL OR account_type = sqlc.narg('account_type')::account_type_enum)
  AND (sqlc.narg('root_type')::root_type_enum IS NULL OR root_type = sqlc.narg('root_type')::root_type_enum)
  AND (sqlc.narg('is_active')::bool IS NULL OR is_active = sqlc.narg('is_active')::bool)
ORDER BY account_code ASC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountAccounts :one
SELECT COUNT(*) 
FROM finance_accounts
WHERE tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND (sqlc.narg('account_type')::account_type_enum IS NULL OR account_type = sqlc.narg('account_type')::account_type_enum)
  AND (sqlc.narg('root_type')::root_type_enum IS NULL OR root_type = sqlc.narg('root_type')::root_type_enum)
  AND (sqlc.narg('is_active')::bool IS NULL OR is_active = sqlc.narg('is_active')::bool);

-- name: ListAccountsByParent :many
SELECT * FROM finance_accounts
WHERE parent_account_id = sqlc.narg('parent_account_id')
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
ORDER BY account_code ASC;

-- name: GetAccountHierarchy :many
SELECT * FROM finance_accounts
WHERE tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND account_path LIKE sqlc.arg('account_path_prefix') || '%'
ORDER BY account_path ASC;

-- name: GetRootAccounts :many
SELECT * FROM finance_accounts
WHERE parent_account_id IS NULL
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
ORDER BY account_code ASC;

-- name: UpdateAccount :one
UPDATE finance_accounts
SET 
    account_name = COALESCE(sqlc.narg('account_name'), account_name),
    account_description = COALESCE(sqlc.narg('account_description'), account_description),
    account_type = COALESCE(sqlc.narg('account_type'), account_type),
    account_subtype = COALESCE(sqlc.narg('account_subtype'), account_subtype),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    allow_manual_entries = COALESCE(sqlc.narg('allow_manual_entries'), allow_manual_entries),
    require_reference = COALESCE(sqlc.narg('require_reference'), require_reference),
    financial_statement_line = COALESCE(sqlc.narg('financial_statement_line'), financial_statement_line),
    report_order = COALESCE(sqlc.narg('report_order'), report_order),
    is_budgetable = COALESCE(sqlc.narg('is_budgetable'), is_budgetable),
    budget_variance_threshold = COALESCE(sqlc.narg('budget_variance_threshold'), budget_variance_threshold),
    account_attributes = COALESCE(sqlc.narg('account_attributes'), account_attributes),
    updated_at = NOW(),
    updated_by = sqlc.narg('updated_by')
WHERE id = sqlc.arg('account_id') 
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
RETURNING *;

-- name: UpdateAccountBalance :exec
UPDATE finance_accounts
SET 
    current_balance = sqlc.arg('current_balance'),
    ytd_balance = sqlc.arg('ytd_balance'),
    last_transaction_date = sqlc.arg('last_transaction_date'),
    updated_at = NOW()
WHERE id = sqlc.arg('account_id') 
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: SoftDeleteAccount :exec
UPDATE finance_accounts
SET 
    deleted_at = NOW(),
    updated_at = NOW(),
    updated_by = sqlc.narg('updated_by')
WHERE id = sqlc.arg('account_id') 
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: RestoreAccount :exec
UPDATE finance_accounts
SET 
    deleted_at = NULL,
    updated_at = NOW(),
    updated_by = sqlc.narg('updated_by')
WHERE id = sqlc.arg('account_id') 
  AND tenant_id = current_tenant_id();

-- name: GetAccountsForFinancialStatements :many
SELECT 
    a.*,
    COALESCE(SUM(CASE WHEN te.debit_amount > 0 THEN te.debit_amount ELSE -te.credit_amount END), 0) as calculated_balance
FROM finance_accounts a
LEFT JOIN finance_transaction_entries te ON a.id = te.account_id
LEFT JOIN finance_transactions t ON te.transaction_id = t.id
WHERE a.tenant_id = current_tenant_id()
  AND a.deleted_at IS NULL
  AND a.is_active = true
  AND (sqlc.narg('financial_statement_line')::text IS NULL OR a.financial_statement_line = sqlc.narg('financial_statement_line'))
  AND (t.transaction_status = 'POSTED' OR t.id IS NULL)
  AND (t.posting_date <= sqlc.arg('posting_date') OR t.posting_date IS NULL)
GROUP BY a.id
ORDER BY a.report_order ASC, a.account_code ASC;

-- name: SearchAccounts :many
SELECT * FROM finance_accounts
WHERE tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND (
    account_code ILIKE '%' || sqlc.arg('search_term') || '%' OR
    account_name ILIKE '%' || sqlc.arg('search_term') || '%' OR
    account_description ILIKE '%' || sqlc.arg('search_term') || '%'
  )
ORDER BY 
  CASE WHEN account_code ILIKE sqlc.arg('search_term') || '%' THEN 1 ELSE 2 END,
  account_code ASC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetAccountsWithNonZeroBalance :many
SELECT * FROM finance_accounts
WHERE tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND current_balance != 0
ORDER BY ABS(current_balance) DESC;

-- name: GetControlAccounts :many
SELECT * FROM finance_accounts
WHERE tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND is_control_account = true
ORDER BY account_code ASC;

-- name: GetAccountsByEntity :many
SELECT * FROM finance_accounts
WHERE entity_id = sqlc.narg('entity_id')
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
ORDER BY account_code ASC;

-- name: ValidateAccountHierarchy :one
SELECT 
    CASE 
        WHEN EXISTS(
            SELECT 1 FROM finance_accounts 
            WHERE parent_account_id = sqlc.narg('parent_account_id') AND id = sqlc.narg('parent_account_id')
        ) THEN false -- Self reference check
        ELSE true
    END as is_valid_hierarchy;
