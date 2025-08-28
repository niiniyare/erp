-- =====================================================================
-- FINANCE MODULE - TRANSACTION ENTRIES QUERIES
-- SQLC queries for journal entries with proper tenant isolation
-- =====================================================================

-- name: CreateTransactionEntry :one
INSERT INTO finance_transaction_entries (
    tenant_id,
    transaction_id,
    entry_number,
    account_id,
    debit_amount,
    credit_amount,
    description,
    reference,
    cost_center,
    department,
    project_id,
    original_currency,
    original_amount,
    exchange_rate,
    tax_code,
    tax_rate,
    tax_amount
) VALUES (
    current_tenant_id(),
    sqlc.arg('transaction_id'), 
    sqlc.arg('entry_number'), 
    sqlc.arg('account_id'), 
    sqlc.arg('debit_amount'), 
    sqlc.arg('credit_amount'), 
    sqlc.arg('description'), 
    sqlc.narg('reference'), 
    sqlc.narg('cost_center'), 
    sqlc.narg('department'), 
    sqlc.narg('project_id'), 
    sqlc.narg('original_currency'), 
    sqlc.arg('original_amount'), 
    sqlc.arg('exchange_rate'), 
    sqlc.narg('tax_code'), 
    sqlc.arg('tax_rate'), 
    sqlc.arg('tax_amount')
) RETURNING *;

-- name: GetTransactionEntryByID :one
SELECT * FROM finance_transaction_entries
WHERE id = sqlc.arg('id') 
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: ListTransactionEntries :many
SELECT * FROM finance_transaction_entries
WHERE transaction_id = sqlc.arg('transaction_id') 
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
ORDER BY entry_number ASC;

-- name: GetTransactionEntriesWithAccounts :many
SELECT 
    te.*,
    a.account_code,
    a.account_name,
    a.root_type,
    a.account_type,
    a.normal_balance
FROM finance_transaction_entries te
JOIN finance_accounts a ON te.account_id = a.id
WHERE te.transaction_id = sqlc.arg('transaction_id') 
  AND te.tenant_id = current_tenant_id()
  AND te.deleted_at IS NULL
ORDER BY te.entry_number ASC;

-- name: GetAccountEntries :many
SELECT 
    te.*,
    t.transaction_number,
    t.transaction_date,
    t.transaction_type,
    t.transaction_status,
    t.description AS transaction_description
FROM finance_transaction_entries te
JOIN finance_transactions t ON te.transaction_id = t.id
WHERE te.account_id = sqlc.arg('account_id')
  AND te.tenant_id = current_tenant_id()
  AND te.deleted_at IS NULL
  AND (sqlc.narg('date_from')::date IS NULL OR t.transaction_date >= sqlc.narg('date_from')::date)
  AND (sqlc.narg('date_to')::date IS NULL OR t.transaction_date <= sqlc.narg('date_to')::date)
  AND (sqlc.narg('transaction_status')::transaction_status_enum IS NULL OR t.transaction_status = sqlc.narg('transaction_status')::transaction_status_enum)
ORDER BY t.transaction_date DESC, te.entry_number ASC
LIMIT sqlc.arg('limit')
OFFSET sqlc.arg('offset');

-- name: CountAccountEntries :one
SELECT COUNT(*)
FROM finance_transaction_entries te
JOIN finance_transactions t ON te.transaction_id = t.id
WHERE te.account_id = sqlc.arg('account_id')
  AND te.tenant_id = current_tenant_id()
  AND te.deleted_at IS NULL
  AND (sqlc.narg('date_from')::date IS NULL OR t.transaction_date >= sqlc.narg('date_from')::date)
  AND (sqlc.narg('date_to')::date IS NULL OR t.transaction_date <= sqlc.narg('date_to')::date)
  AND (sqlc.narg('transaction_status')::transaction_status_enum IS NULL OR t.transaction_status = sqlc.narg('transaction_status')::transaction_status_enum);

-- name: UpdateTransactionEntry :one
UPDATE finance_transaction_entries
SET 
    account_id    = COALESCE(sqlc.narg('account_id'), account_id),
    debit_amount  = COALESCE(sqlc.narg('debit_amount'), debit_amount),
    credit_amount = COALESCE(sqlc.narg('credit_amount'), credit_amount),
    description   = COALESCE(sqlc.narg('description'), description),
    reference     = COALESCE(sqlc.narg('reference'), reference),
    cost_center   = COALESCE(sqlc.narg('cost_center'), cost_center),
    department    = COALESCE(sqlc.narg('department'), department),
    project_id    = COALESCE(sqlc.narg('project_id'), project_id),
    tax_code      = COALESCE(sqlc.narg('tax_code'), tax_code),
    tax_rate      = COALESCE(sqlc.narg('tax_rate'), tax_rate),
    tax_amount    = COALESCE(sqlc.narg('tax_amount'), tax_amount),
    updated_at    = NOW()
WHERE id = sqlc.arg('id')
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
RETURNING *;

-- name: DeleteTransactionEntry :exec
DELETE FROM finance_transaction_entries
WHERE id = sqlc.arg('id') 
  AND tenant_id = current_tenant_id();

-- name: DeleteTransactionEntries :exec
DELETE FROM finance_transaction_entries
WHERE transaction_id = sqlc.arg('transaction_id') 
  AND tenant_id = current_tenant_id();

-- name: GetAccountBalance :one
SELECT 
    account_id,
    SUM(CASE WHEN debit_amount > 0 THEN debit_amount ELSE 0 END) as total_debits,
    SUM(CASE WHEN credit_amount > 0 THEN credit_amount ELSE 0 END) as total_credits,
    SUM(CASE WHEN debit_amount > 0 THEN debit_amount ELSE -credit_amount END) as net_balance
FROM finance_transaction_entries te
JOIN finance_transactions t ON te.transaction_id = t.id
WHERE te.account_id = sqlc.arg('account_id') 
  AND te.tenant_id = current_tenant_id()
  AND te.deleted_at IS NULL
  AND t.transaction_status = 'POSTED'
  AND (sqlc.narg('as_of_date')::date IS NULL OR t.posting_date <= sqlc.narg('as_of_date'))
GROUP BY account_id;

-- name: GetEntriesByCostCenter :many
SELECT 
    te.*,
    a.account_code,
    a.account_name,
    t.transaction_number,
    t.transaction_date
FROM finance_transaction_entries te
JOIN finance_accounts a ON te.account_id = a.id
JOIN finance_transactions t ON te.transaction_id = t.id
WHERE te.cost_center = sqlc.narg('cost_center') 
  AND te.tenant_id = current_tenant_id()
  AND te.deleted_at IS NULL
  AND t.transaction_status = 'POSTED'
  AND (sqlc.narg('date_from')::date IS NULL OR t.transaction_date >= sqlc.narg('date_from'))
  AND (sqlc.narg('date_to')::date IS NULL OR t.transaction_date <= sqlc.narg('date_to'))
ORDER BY t.transaction_date DESC, te.entry_number ASC;

-- name: GetEntriesByProject :many
SELECT 
    te.*,
    a.account_code,
    a.account_name,
    t.transaction_number,
    t.transaction_date
FROM finance_transaction_entries te
JOIN finance_accounts a ON te.account_id = a.id
JOIN finance_transactions t ON te.transaction_id = t.id
WHERE te.project_id = sqlc.narg('project_id') 
  AND te.tenant_id = current_tenant_id()
  AND te.deleted_at IS NULL
  AND t.transaction_status = 'POSTED'
  AND (sqlc.narg('date_from')::date IS NULL OR t.transaction_date >= sqlc.narg('date_from'))
  AND (sqlc.narg('date_to')::date IS NULL OR t.transaction_date <= sqlc.narg('date_to'))
ORDER BY t.transaction_date DESC, te.entry_number ASC;

-- name: GetUnreconciledEntries :many
SELECT 
    te.*,
    a.account_code,
    a.account_name,
    t.transaction_number,
    t.transaction_date
FROM finance_transaction_entries te
JOIN finance_accounts a ON te.account_id = a.id
JOIN finance_transactions t ON te.transaction_id = t.id
WHERE te.account_id = sqlc.arg('account_id') 
  AND te.reconciled = false
  AND te.tenant_id = current_tenant_id()
  AND te.deleted_at IS NULL
  AND t.transaction_status = 'POSTED'
ORDER BY t.transaction_date ASC;

-- name: GetEntryTaxSummary :many
SELECT 
    te.tax_code,
    te.tax_rate,
    COUNT(*) as entry_count,
    SUM(te.debit_amount + te.credit_amount) as taxable_amount,
    SUM(te.tax_amount) as total_tax
FROM finance_transaction_entries te
JOIN finance_transactions t ON te.transaction_id = t.id
WHERE te.tenant_id = current_tenant_id()
  AND te.deleted_at IS NULL
  AND te.tax_code IS NOT NULL
  AND t.transaction_status = 'POSTED'
  AND t.transaction_date >= sqlc.arg('date_from')
  AND t.transaction_date <= sqlc.arg('date_to')
GROUP BY te.tax_code, te.tax_rate
ORDER BY te.tax_code, te.tax_rate;

-- name: ValidateTransactionEntriesBalance :one
SELECT 
    transaction_id,
    SUM(debit_amount) as total_debits,
    SUM(credit_amount) as total_credits,
    (SUM(debit_amount) = SUM(credit_amount)) as is_balanced
FROM finance_transaction_entries
WHERE transaction_id = sqlc.arg('transaction_id') 
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
GROUP BY transaction_id;

-- name: GetEntriesByDepartment :many
SELECT
    te.*,
    a.account_code,
    a.account_name,
    t.transaction_number,
    t.transaction_date
FROM finance_transaction_entries te
JOIN finance_accounts a ON te.account_id = a.id
JOIN finance_transactions t ON te.transaction_id = t.id
WHERE te.department = sqlc.narg('department')
  AND te.tenant_id = current_tenant_id()
  AND te.deleted_at IS NULL
  AND t.transaction_status = 'POSTED'
  AND (sqlc.narg('date_from')::date IS NULL OR t.transaction_date >= sqlc.narg('date_from')::date)
  AND (sqlc.narg('date_to')::date IS NULL OR t.transaction_date <= sqlc.narg('date_to')::date)
ORDER BY t.transaction_date DESC, te.entry_number ASC;

-- name: GetTrialBalance :many
SELECT
    a.id,
    a.account_code,
    a.account_name,
    a.root_type,
    a.account_type,
    a.normal_balance,
    COALESCE(SUM(CASE WHEN te.debit_amount > 0 THEN te.debit_amount ELSE 0 END), 0) AS total_debits,
    COALESCE(SUM(CASE WHEN te.credit_amount > 0 THEN te.credit_amount ELSE 0 END), 0) AS total_credits,
    COALESCE(SUM(CASE WHEN te.debit_amount > 0 THEN te.debit_amount ELSE -te.credit_amount END), 0) AS net_balance
FROM finance_accounts a
LEFT JOIN finance_transaction_entries te ON a.id = te.account_id
    AND te.tenant_id = current_tenant_id()
    AND te.deleted_at IS NULL
LEFT JOIN finance_transactions t ON te.transaction_id = t.id
    AND t.transaction_status = 'POSTED'
    AND (sqlc.narg('as_of_date')::date IS NULL OR t.posting_date <= sqlc.narg('as_of_date')::date)
WHERE a.tenant_id = current_tenant_id()
  AND a.deleted_at IS NULL
  AND a.is_active = true
GROUP BY a.id, a.account_code, a.account_name, a.root_type, a.account_type, a.normal_balance
HAVING
    COALESCE(SUM(CASE WHEN te.debit_amount > 0 THEN te.debit_amount ELSE 0 END), 0) != 0 OR
    COALESCE(SUM(CASE WHEN te.credit_amount > 0 THEN te.credit_amount ELSE 0 END), 0) != 0 OR
    sqlc.arg('include_zero_balances') = true
ORDER BY a.account_code ASC;

-- name: MarkEntriesReconciled :exec
UPDATE finance_transaction_entries
SET
    reconciled = true,
    reconciled_date = sqlc.arg('reconciled_date'),
    reconciliation_reference = sqlc.narg('reconciliation_reference'),
    updated_at = NOW()
WHERE id = ANY(sqlc.arg('entry_ids')::uuid[])
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;
