-- =====================================================================
-- FINANCE MODULE - TRANSACTIONS QUERIES
-- SQLC queries for financial transactions with proper tenant isolation
-- =====================================================================

-- name: CreateTransaction :one
INSERT INTO finance_transactions (
    tenant_id,
    entity_id,
    transaction_number,
    transaction_type,
    transaction_status,
    transaction_date,
    posting_date,
    due_date,
    description,
    reference_number,
    external_reference,
    currency_code,
    exchange_rate,
    total_debit_amount,
    total_credit_amount,
    source_module,
    source_document_type,
    source_document_id,
    batch_id,
    approval_required,
    approval_status,
    is_recurring,
    recurring_frequency,
    next_recurring_date,
    transaction_attributes,
    created_by
) VALUES (
    current_tenant_id(),
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25
) RETURNING *;

-- name: GetTransactionByID :one
SELECT * FROM finance_transactions
WHERE id = $1 
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: GetTransactionByNumber :one
SELECT * FROM finance_transactions
WHERE transaction_number = $1 
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: ListTransactions :many
SELECT * FROM finance_transactions
WHERE tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND ($1::text IS NULL OR transaction_type = $1)
  AND ($2::text IS NULL OR transaction_status = $2)
  AND ($3::date IS NULL OR transaction_date >= $3)
  AND ($4::date IS NULL OR transaction_date <= $4)
ORDER BY transaction_date DESC, created_at DESC
LIMIT $5 OFFSET $6;

-- name: CountTransactions :one
SELECT COUNT(*) FROM finance_transactions
WHERE tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND ($1::text IS NULL OR transaction_type = $1)
  AND ($2::text IS NULL OR transaction_status = $2)
  AND ($3::date IS NULL OR transaction_date >= $3)
  AND ($4::date IS NULL OR transaction_date <= $4);

-- name: GetTransactionWithEntries :many
SELECT 
    t.*,
    te.id as entry_id,
    te.entry_number,
    te.account_id,
    te.debit_amount,
    te.credit_amount,
    te.description as entry_description,
    te.reference as entry_reference,
    te.cost_center,
    te.department,
    te.project_id,
    a.account_code,
    a.account_name,
    a.root_type,
    a.normal_balance
FROM finance_transactions t
LEFT JOIN finance_transaction_entries te ON t.id = te.transaction_id
LEFT JOIN finance_chart_of_accounts a ON te.account_id = a.id
WHERE t.id = $1 
  AND t.tenant_id = current_tenant_id()
  AND t.deleted_at IS NULL
ORDER BY te.entry_number ASC;

-- name: UpdateTransaction :one
UPDATE finance_transactions
SET 
    transaction_status = COALESCE($2, transaction_status),
    posting_date = COALESCE($3, posting_date),
    due_date = COALESCE($4, due_date),
    description = COALESCE($5, description),
    reference_number = COALESCE($6, reference_number),
    external_reference = COALESCE($7, external_reference),
    total_debit_amount = COALESCE($8, total_debit_amount),
    total_credit_amount = COALESCE($9, total_credit_amount),
    approval_status = COALESCE($10, approval_status),
    approved_by = COALESCE($11, approved_by),
    approved_at = COALESCE($12, approved_at),
    approval_notes = COALESCE($13, approval_notes),
    transaction_attributes = COALESCE($14, transaction_attributes),
    updated_at = NOW(),
    updated_by = $15
WHERE id = $1 
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
RETURNING *;

-- name: PostTransaction :one
UPDATE finance_transactions
SET 
    transaction_status = 'POSTED',
    posting_date = COALESCE($2, NOW()::date),
    posted_by = $3,
    posted_at = NOW(),
    updated_at = NOW(),
    updated_by = $3
WHERE id = $1 
  AND tenant_id = current_tenant_id()
  AND transaction_status IN ('APPROVED', 'DRAFT')
  AND deleted_at IS NULL
RETURNING *;

-- name: ApproveTransaction :one
UPDATE finance_transactions
SET 
    approval_status = 'APPROVED',
    approved_by = $2,
    approved_at = NOW(),
    approval_notes = $3,
    updated_at = NOW(),
    updated_by = $2
WHERE id = $1 
  AND tenant_id = current_tenant_id()
  AND approval_status = 'PENDING'
  AND deleted_at IS NULL
RETURNING *;

-- name: RejectTransaction :one
UPDATE finance_transactions
SET 
    approval_status = 'REJECTED',
    approved_by = $2,
    approved_at = NOW(),
    approval_notes = $3,
    updated_at = NOW(),
    updated_by = $2
WHERE id = $1 
  AND tenant_id = current_tenant_id()
  AND approval_status = 'PENDING'
  AND deleted_at IS NULL
RETURNING *;

-- name: ReverseTransaction :one
UPDATE finance_transactions
SET 
    is_reversed = true,
    reversed_by_transaction_id = $2,
    reversal_reason = $3,
    updated_at = NOW(),
    updated_by = $4
WHERE id = $1 
  AND tenant_id = current_tenant_id()
  AND transaction_status = 'POSTED'
  AND is_reversed = false
  AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteTransaction :exec
UPDATE finance_transactions
SET 
    deleted_at = NOW(),
    updated_at = NOW(),
    updated_by = $2
WHERE id = $1 
  AND tenant_id = current_tenant_id()
  AND transaction_status IN ('DRAFT', 'CANCELLED')
  AND deleted_at IS NULL;

-- name: GetTransactionsByBatch :many
SELECT * FROM finance_transactions
WHERE batch_id = $1 
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
ORDER BY created_at ASC;

-- name: GetTransactionsBySourceDocument :many
SELECT * FROM finance_transactions
WHERE source_document_type = $1 
  AND source_document_id = $2
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
ORDER BY created_at ASC;

-- name: GetPendingApprovalTransactions :many
SELECT * FROM finance_transactions
WHERE tenant_id = current_tenant_id()
  AND approval_status = 'PENDING'
  AND deleted_at IS NULL
ORDER BY created_at ASC
LIMIT $1 OFFSET $2;

-- name: GetRecurringTransactionsDue :many
SELECT * FROM finance_transactions
WHERE tenant_id = current_tenant_id()
  AND is_recurring = true
  AND next_recurring_date <= $1
  AND transaction_status = 'POSTED'
  AND deleted_at IS NULL
ORDER BY next_recurring_date ASC;

-- name: UpdateRecurringTransactionNextDate :exec
UPDATE finance_transactions
SET 
    next_recurring_date = $2,
    updated_at = NOW()
WHERE id = $1 
  AND tenant_id = current_tenant_id()
  AND is_recurring = true
  AND deleted_at IS NULL;

-- name: GetTransactionSummaryByPeriod :many
SELECT 
    transaction_type,
    transaction_status,
    COUNT(*) as transaction_count,
    SUM(total_debit_amount) as total_debit,
    SUM(total_credit_amount) as total_credit
FROM finance_transactions
WHERE tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND transaction_date >= $1
  AND transaction_date <= $2
GROUP BY transaction_type, transaction_status
ORDER BY transaction_type, transaction_status;

-- name: ValidateTransactionBalance :one
SELECT 
    id,
    total_debit_amount,
    total_credit_amount,
    (total_debit_amount = total_credit_amount) as is_balanced
FROM finance_transactions
WHERE id = $1 
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: SearchTransactions :many
SELECT * FROM finance_transactions
WHERE tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND (
    transaction_number ILIKE '%' || $1 || '%' OR
    description ILIKE '%' || $1 || '%' OR
    reference_number ILIKE '%' || $1 || '%' OR
    external_reference ILIKE '%' || $1 || '%'
  )
ORDER BY 
  CASE WHEN transaction_number ILIKE $1 || '%' THEN 1 ELSE 2 END,
  transaction_date DESC
LIMIT $2 OFFSET $3;