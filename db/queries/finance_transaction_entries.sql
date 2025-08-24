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
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
) RETURNING *;


-- name: GetTransactionEntryByID :one
SELECT * FROM finance_transaction_entries
WHERE id = $1 
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: ListTransactionEntries :many
SELECT * FROM finance_transaction_entries
WHERE transaction_id = $1 
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
JOIN finance_chart_of_accounts a ON te.account_id = a.id
WHERE te.transaction_id = $1 
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
    t.description as transaction_description
FROM finance_transaction_entries te
JOIN finance_transactions t ON te.transaction_id = t.id
WHERE te.account_id = $1 
  AND te.tenant_id = current_tenant_id()
  AND te.deleted_at IS NULL
  AND ($2::date IS NULL OR t.transaction_date >= $2)
  AND ($3::date IS NULL OR t.transaction_date <= $3)
  AND ($4::text IS NULL OR t.transaction_status = $4)
ORDER BY t.transaction_date DESC, te.entry_number ASC
LIMIT $5 OFFSET $6;

-- name: CountAccountEntries :one
SELECT COUNT(*) FROM finance_transaction_entries te
JOIN finance_transactions t ON te.transaction_id = t.id
WHERE te.account_id = $1 
  AND te.tenant_id = current_tenant_id()
  AND te.deleted_at IS NULL
  AND ($2::date IS NULL OR t.transaction_date >= $2)
  AND ($3::date IS NULL OR t.transaction_date <= $3)
  AND ($4::text IS NULL OR t.transaction_status = $4);

-- name: UpdateTransactionEntry :one
UPDATE finance_transaction_entries
SET 
    account_id = COALESCE($2, account_id),
    debit_amount = COALESCE($3, debit_amount),
    credit_amount = COALESCE($4, credit_amount),
    description = COALESCE($5, description),
    reference = COALESCE($6, reference),
    cost_center = COALESCE($7, cost_center),
    department = COALESCE($8, department),
    project_id = COALESCE($9, project_id),
    tax_code = COALESCE($10, tax_code),
    tax_rate = COALESCE($11, tax_rate),
    tax_amount = COALESCE($12, tax_amount),
    updated_at = NOW()
WHERE id = $1 
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
RETURNING *;

-- name: DeleteTransactionEntry :exec
DELETE FROM finance_transaction_entries
WHERE id = $1 
  AND tenant_id = current_tenant_id();

-- name: DeleteTransactionEntries :exec
DELETE FROM finance_transaction_entries
WHERE transaction_id = $1 
  AND tenant_id = current_tenant_id();

-- name: GetAccountBalance :one
SELECT 
    account_id,
    SUM(CASE WHEN debit_amount > 0 THEN debit_amount ELSE 0 END) as total_debits,
    SUM(CASE WHEN credit_amount > 0 THEN credit_amount ELSE 0 END) as total_credits,
    SUM(CASE WHEN debit_amount > 0 THEN debit_amount ELSE -credit_amount END) as net_balance
FROM finance_transaction_entries te
JOIN finance_transactions t ON te.transaction_id = t.id
WHERE te.account_id = $1 
  AND te.tenant_id = current_tenant_id()
  AND te.deleted_at IS NULL
  AND t.transaction_status = 'POSTED'
  AND ($2::date IS NULL OR t.posting_date <= $2)
GROUP BY account_id;

-- name: GetTrialBalance :many
SELECT 
    a.id,
    a.account_code,
    a.account_name,
    a.root_type,
    a.account_type,
    a.normal_balance,
    COALESCE(SUM(CASE WHEN te.debit_amount > 0 THEN te.debit_amount ELSE 0 END), 0) as total_debits,
    COALESCE(SUM(CASE WHEN te.credit_amount > 0 THEN te.credit_amount ELSE 0 END), 0) as total_credits,
    COALESCE(SUM(CASE WHEN te.debit_amount > 0 THEN te.debit_amount ELSE -te.credit_amount END), 0) as net_balance
FROM finance_chart_of_accounts a
LEFT JOIN finance_transaction_entries te ON a.id = te.account_id 
    AND te.tenant_id = current_tenant_id()
    AND te.deleted_at IS NULL
LEFT JOIN finance_transactions t ON te.transaction_id = t.id 
    AND t.transaction_status = 'POSTED'
    AND ($1::date IS NULL OR t.posting_date <= $1)
WHERE a.tenant_id = current_tenant_id()
  AND a.deleted_at IS NULL
  AND a.is_active = true
GROUP BY a.id, a.account_code, a.account_name, a.root_type, a.account_type, a.normal_balance
HAVING 
    COALESCE(SUM(CASE WHEN te.debit_amount > 0 THEN te.debit_amount ELSE 0 END), 0) != 0 OR
    COALESCE(SUM(CASE WHEN te.credit_amount > 0 THEN te.credit_amount ELSE 0 END), 0) != 0 OR
    $2::boolean = true -- include_zero_balances parameter
ORDER BY a.account_code ASC;

-- name: GetEntriesByCostCenter :many
SELECT 
    te.*,
    a.account_code,
    a.account_name,
    t.transaction_number,
    t.transaction_date
FROM finance_transaction_entries te
JOIN finance_chart_of_accounts a ON te.account_id = a.id
JOIN finance_transactions t ON te.transaction_id = t.id
WHERE te.cost_center = $1 
  AND te.tenant_id = current_tenant_id()
  AND te.deleted_at IS NULL
  AND t.transaction_status = 'POSTED'
  AND ($2::date IS NULL OR t.transaction_date >= $2)
  AND ($3::date IS NULL OR t.transaction_date <= $3)
ORDER BY t.transaction_date DESC, te.entry_number ASC;

-- name: GetEntriesByDepartment :many
SELECT 
    te.*,
    a.account_code,
    a.account_name,
    t.transaction_number,
    t.transaction_date
FROM finance_transaction_entries te
JOIN finance_chart_of_accounts a ON te.account_id = a.id
JOIN finance_transactions t ON te.transaction_id = t.id
WHERE te.department = $1 
  AND te.tenant_id = current_tenant_id()
  AND te.deleted_at IS NULL
  AND t.transaction_status = 'POSTED'
  AND ($2::date IS NULL OR t.transaction_date >= $2)
  AND ($3::date IS NULL OR t.transaction_date <= $3)
ORDER BY t.transaction_date DESC, te.entry_number ASC;

-- name: GetEntriesByProject :many
SELECT 
    te.*,
    a.account_code,
    a.account_name,
    t.transaction_number,
    t.transaction_date
FROM finance_transaction_entries te
JOIN finance_chart_of_accounts a ON te.account_id = a.id
JOIN finance_transactions t ON te.transaction_id = t.id
WHERE te.project_id = $1 
  AND te.tenant_id = current_tenant_id()
  AND te.deleted_at IS NULL
  AND t.transaction_status = 'POSTED'
  AND ($2::date IS NULL OR t.transaction_date >= $2)
  AND ($3::date IS NULL OR t.transaction_date <= $3)
ORDER BY t.transaction_date DESC, te.entry_number ASC;

-- name: GetUnreconciledEntries :many
SELECT 
    te.*,
    a.account_code,
    a.account_name,
    t.transaction_number,
    t.transaction_date
FROM finance_transaction_entries te
JOIN finance_chart_of_accounts a ON te.account_id = a.id
JOIN finance_transactions t ON te.transaction_id = t.id
WHERE te.account_id = $1 
  AND te.reconciled = false
  AND te.tenant_id = current_tenant_id()
  AND te.deleted_at IS NULL
  AND t.transaction_status = 'POSTED'
ORDER BY t.transaction_date ASC;

-- name: MarkEntriesReconciled :exec
UPDATE finance_transaction_entries
SET 
    reconciled = true,
    reconciled_date = $2,
    reconciliation_reference = $3,
    updated_at = NOW()
WHERE id = ANY($1::uuid[])
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

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
  AND t.transaction_date >= $1
  AND t.transaction_date <= $2
GROUP BY te.tax_code, te.tax_rate
ORDER BY te.tax_code, te.tax_rate;

-- name: ValidateTransactionEntriesBalance :one
SELECT 
    transaction_id,
    SUM(debit_amount) as total_debits,
    SUM(credit_amount) as total_credits,
    (SUM(debit_amount) = SUM(credit_amount)) as is_balanced
FROM finance_transaction_entries
WHERE transaction_id = $1 
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
GROUP BY transaction_id;