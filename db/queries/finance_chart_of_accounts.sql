-- =====================================================================
-- FINANCE MODULE - CHART OF ACCOUNTS QUERIES
-- SQLC queries for chart of accounts with proper tenant isolation
-- =====================================================================

-- name: CreateAccount :one
INSERT INTO finance_chart_of_accounts (
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
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24
) RETURNING *;

-- name: GetAccountByID :one
SELECT * FROM finance_chart_of_accounts
WHERE id = $1 
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: GetAccountByCode :one
SELECT * FROM finance_chart_of_accounts
WHERE account_code = $1 
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: ListAccounts :many
SELECT * FROM finance_chart_of_accounts
WHERE tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND ($1::text IS NULL OR account_type = $1)
  AND ($2::text IS NULL OR root_type = $2)
  AND ($3::bool IS NULL OR is_active = $3)
ORDER BY account_code ASC
LIMIT $4 OFFSET $5;

-- name: CountAccounts :one
SELECT COUNT(*) FROM finance_chart_of_accounts
WHERE tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND ($1::text IS NULL OR account_type = $1)
  AND ($2::text IS NULL OR root_type = $2)
  AND ($3::bool IS NULL OR is_active = $3);

-- name: ListAccountsByParent :many
SELECT * FROM finance_chart_of_accounts
WHERE parent_account_id = $1
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
ORDER BY account_code ASC;

-- name: GetAccountHierarchy :many
SELECT * FROM finance_chart_of_accounts
WHERE tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND account_path LIKE $1 || '%'
ORDER BY account_path ASC;

-- name: GetRootAccounts :many
SELECT * FROM finance_chart_of_accounts
WHERE parent_account_id IS NULL
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
ORDER BY account_code ASC;

-- name: UpdateAccount :one
UPDATE finance_chart_of_accounts
SET 
    account_name = COALESCE($2, account_name),
    account_description = COALESCE($3, account_description),
    account_type = COALESCE($4, account_type),
    account_subtype = COALESCE($5, account_subtype),
    is_active = COALESCE($6, is_active),
    allow_manual_entries = COALESCE($7, allow_manual_entries),
    require_reference = COALESCE($8, require_reference),
    financial_statement_line = COALESCE($9, financial_statement_line),
    report_order = COALESCE($10, report_order),
    is_budgetable = COALESCE($11, is_budgetable),
    budget_variance_threshold = COALESCE($12, budget_variance_threshold),
    account_attributes = COALESCE($13, account_attributes),
    updated_at = NOW(),
    updated_by = $14
WHERE id = $1 
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
RETURNING *;

-- name: UpdateAccountBalance :exec
UPDATE finance_chart_of_accounts
SET 
    current_balance = $2,
    ytd_balance = $3,
    last_transaction_date = $4,
    updated_at = NOW()
WHERE id = $1 
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: SoftDeleteAccount :exec
UPDATE finance_chart_of_accounts
SET 
    deleted_at = NOW(),
    updated_at = NOW(),
    updated_by = $2
WHERE id = $1 
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: RestoreAccount :exec
UPDATE finance_chart_of_accounts
SET 
    deleted_at = NULL,
    updated_at = NOW(),
    updated_by = $2
WHERE id = $1 
  AND tenant_id = current_tenant_id();

-- name: GetAccountsForFinancialStatements :many
SELECT 
    a.*,
    COALESCE(SUM(CASE WHEN te.debit_amount > 0 THEN te.debit_amount ELSE -te.credit_amount END), 0) as calculated_balance
FROM finance_chart_of_accounts a
LEFT JOIN finance_transaction_entries te ON a.id = te.account_id
LEFT JOIN finance_transactions t ON te.transaction_id = t.id
WHERE a.tenant_id = current_tenant_id()
  AND a.deleted_at IS NULL
  AND a.is_active = true
  AND ($1::text IS NULL OR a.financial_statement_line = $1)
  AND (t.transaction_status = 'POSTED' OR t.id IS NULL)
  AND (t.posting_date <= $2 OR t.posting_date IS NULL)
GROUP BY a.id
ORDER BY a.report_order ASC, a.account_code ASC;

-- name: SearchAccounts :many
SELECT * FROM finance_chart_of_accounts
WHERE tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND (
    account_code ILIKE '%' || $1 || '%' OR
    account_name ILIKE '%' || $1 || '%' OR
    account_description ILIKE '%' || $1 || '%'
  )
ORDER BY 
  CASE WHEN account_code ILIKE $1 || '%' THEN 1 ELSE 2 END,
  account_code ASC
LIMIT $2 OFFSET $3;

-- name: GetAccountsWithNonZeroBalance :many
SELECT * FROM finance_chart_of_accounts
WHERE tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND current_balance != 0
ORDER BY ABS(current_balance) DESC;

-- name: GetControlAccounts :many
SELECT * FROM finance_chart_of_accounts
WHERE tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND is_control_account = true
ORDER BY account_code ASC;

-- name: GetAccountsByEntity :many
SELECT * FROM finance_chart_of_accounts
WHERE entity_id = $1
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
ORDER BY account_code ASC;

-- name: ValidateAccountHierarchy :one
SELECT 
    CASE 
        WHEN EXISTS(
            SELECT 1 FROM finance_chart_of_accounts 
            WHERE parent_account_id = $1 AND id = $1
        ) THEN false -- Self reference check
        ELSE true
    END as is_valid_hierarchy;