-- =====================================================================
-- FINANCE MODULE - TRANSACTIONS QUERIES
-- SQLC queries for financial transactions with proper tenant isolation
-- =====================================================================
-- name: CreateTransaction :one
INSERT INTO
  finance_transactions (
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
    memo,
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
    attachment_ids,
    tags,
    created_by
  )
VALUES
  (
    current_tenant_id(),
    sqlc.arg('entity_id'),
    sqlc.arg('transaction_number'),
    sqlc.arg('transaction_type'),
    sqlc.arg('transaction_status'),
    sqlc.arg('transaction_date'),
    sqlc.arg('posting_date'),
    sqlc.arg('due_date'),
    sqlc.arg('description'),
    sqlc.arg('reference_number'),
    sqlc.arg('external_reference'),
    sqlc.arg('memo'),
    sqlc.arg('currency_code'),
    sqlc.arg('exchange_rate'),
    sqlc.arg('total_debit_amount'),
    sqlc.arg('total_credit_amount'),
    sqlc.arg('source_module'),
    sqlc.arg('source_document_type'),
    sqlc.arg('source_document_id'),
    sqlc.arg('batch_id'),
    sqlc.arg('approval_required'),
    sqlc.arg('approval_status'),
    sqlc.arg('is_recurring'),
    sqlc.arg('recurring_frequency'),
    sqlc.arg('next_recurring_date'),
    sqlc.arg('transaction_attributes'),
    sqlc.arg('attachment_ids'),
    sqlc.arg('tags'),
    sqlc.arg('created_by')
  )
RETURNING
  *;

-- name: GetTransactionByID :one
SELECT
  *
FROM
  finance_transactions
WHERE
  id = sqlc.arg('transaction_id')
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: GetTransactionByNumber :one
SELECT
  *
FROM
  finance_transactions
WHERE
  transaction_number = sqlc.arg('transaction_number')
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: ListTransactions :many
SELECT
  *
FROM
  finance_transactions
WHERE
  tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND (
    sqlc.narg('transaction_type')::VARCHAR IS NULL
    OR transaction_type = sqlc.narg('transaction_type')::VARCHAR
  )
  AND (
    sqlc.narg('transaction_status')::VARCHAR IS NULL
    OR transaction_status = sqlc.narg('transaction_status')::VARCHAR
  )
  AND (
    sqlc.narg('date_from')::date IS NULL
    OR transaction_date >= sqlc.narg('date_from')::date
  )
  AND (
    sqlc.narg('date_to')::date IS NULL
    OR transaction_date <= sqlc.narg('date_to')::date
  )
ORDER BY
  transaction_date DESC,
  created_at DESC
LIMIT
  sqlc.arg('limit_count') OFFSET sqlc.arg('offset_count');

-- name: CountTransactions :one
SELECT
  COUNT(*)
FROM
  finance_transactions
WHERE
  tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND (
    sqlc.narg('transaction_type')::VARCHAR IS NULL
    OR transaction_type = sqlc.narg('transaction_type')::VARCHAR
  )
  AND (
    sqlc.narg('transaction_status')::VARCHAR IS NULL
    OR transaction_status = sqlc.narg('transaction_status')::VARCHAR
  )
  AND (
    sqlc.narg('date_from')::date IS NULL
    OR transaction_date >= sqlc.narg('date_from')::date
  )
  AND (
    sqlc.narg('date_to')::date IS NULL
    OR transaction_date <= sqlc.narg('date_to')::date
  );

-- name: GetTransactionWithEntries :many
SELECT
  t.*,
  te.id AS entry_id,
  te.entry_number,
  te.account_id,
  te.debit_amount,
  te.credit_amount,
  te.description AS entry_description,
  te.reference AS entry_reference,
  te.cost_center,
  te.department,
  te.project_id,
  a.account_code,
  a.account_name,
  a.root_type,
  a.normal_balance
FROM
  finance_transactions t
  LEFT JOIN finance_transaction_entries te ON t.id = te.transaction_id
  LEFT JOIN finance_accounts a ON te.account_id = a.id
WHERE
  t.id = sqlc.arg('transaction_id')
  AND t.tenant_id = current_tenant_id()
  AND t.deleted_at IS NULL
ORDER BY
  te.entry_number ASC;

-- name: UpdateTransaction :one
UPDATE
  finance_transactions
SET
  transaction_status = COALESCE(
    sqlc.narg('transaction_status'),
    transaction_status
  ),
  posting_date = COALESCE(sqlc.narg('posting_date'), posting_date),
  due_date = COALESCE(sqlc.narg('due_date'), due_date),
  description = COALESCE(sqlc.narg('description'), description),
  reference_number = COALESCE(sqlc.narg('reference_number'), reference_number),
  external_reference = COALESCE(
    sqlc.narg('external_reference'),
    external_reference
  ),
  memo = COALESCE(sqlc.narg('memo'), memo),
  total_debit_amount = COALESCE(
    sqlc.narg('total_debit_amount'),
    total_debit_amount
  ),
  total_credit_amount = COALESCE(
    sqlc.narg('total_credit_amount'),
    total_credit_amount
  ),
  approval_status = COALESCE(sqlc.narg('approval_status'), approval_status),
  approved_by = COALESCE(sqlc.narg('approved_by'), approved_by),
  approved_at = COALESCE(sqlc.narg('approved_at'), approved_at),
  approval_notes = COALESCE(sqlc.narg('approval_notes'), approval_notes),
  transaction_attributes = COALESCE(
    sqlc.narg('transaction_attributes'),
    transaction_attributes
  ),
  attachment_ids = COALESCE(sqlc.narg('attachment_ids'), attachment_ids),
  tags = COALESCE(sqlc.narg('tags'), tags),
  updated_at = NOW(),
  updated_by = sqlc.arg('updated_by')
WHERE
  id = sqlc.arg('transaction_id')
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
RETURNING
  *;

-- name: PostTransaction :one
UPDATE
  finance_transactions
SET
  transaction_status = 'POSTED',
  posting_date = COALESCE(sqlc.narg('posting_date'), NOW()::date),
  posted_by = sqlc.arg('posted_by'),
  posted_at = NOW(),
  updated_at = NOW(),
  updated_by = sqlc.arg('posted_by')
WHERE
  id = sqlc.arg('transaction_id')
  AND tenant_id = current_tenant_id()
  AND transaction_status IN ('APPROVED', 'DRAFT')
  AND deleted_at IS NULL
RETURNING
  *;

-- name: ApproveTransaction :one
UPDATE
  finance_transactions
SET
  approval_status = 'APPROVED',
  approved_by = sqlc.arg('approved_by'),
  approved_at = NOW(),
  approval_notes = sqlc.arg('approval_notes'),
  updated_at = NOW(),
  updated_by = sqlc.arg('approved_by')
WHERE
  id = sqlc.arg('transaction_id')
  AND tenant_id = current_tenant_id()
  AND approval_status = 'PENDING'
  AND deleted_at IS NULL
RETURNING
  *;

-- name: RejectTransaction :one
UPDATE
  finance_transactions
SET
  approval_status = 'REJECTED',
  approved_by = sqlc.arg('approved_by'),
  approved_at = NOW(),
  approval_notes = sqlc.arg('approval_notes'),
  updated_at = NOW(),
  updated_by = sqlc.arg('approved_by')
WHERE
  id = sqlc.arg('transaction_id')
  AND tenant_id = current_tenant_id()
  AND approval_status = 'PENDING'
  AND deleted_at IS NULL
RETURNING
  *;

-- name: ReverseTransaction :one
UPDATE
  finance_transactions
SET
  is_reversed = TRUE,
  reversed_by_transaction_id = sqlc.arg('reversed_by_transaction_id'),
  reversal_reason = sqlc.arg('reversal_reason'),
  updated_at = NOW(),
  updated_by = sqlc.arg('updated_by')
WHERE
  id = sqlc.arg('transaction_id')
  AND tenant_id = current_tenant_id()
  AND transaction_status = 'POSTED'
  AND is_reversed = false
  AND deleted_at IS NULL
RETURNING
  *;

-- name: SoftDeleteTransaction :exec
UPDATE
  finance_transactions
SET
  deleted_at = NOW(),
  updated_at = NOW(),
  updated_by = sqlc.arg('updated_by')
WHERE
  id = sqlc.arg('transaction_id')
  AND tenant_id = current_tenant_id()
  AND transaction_status IN ('DRAFT', 'CANCELLED')
  AND deleted_at IS NULL;

-- name: GetTransactionsByBatch :many
SELECT
  *
FROM
  finance_transactions
WHERE
  batch_id = sqlc.arg('batch_id')
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
ORDER BY
  created_at ASC;

-- name: GetTransactionsBySourceDocument :many
SELECT
  *
FROM
  finance_transactions
WHERE
  source_document_type = sqlc.arg('source_document_type')
  AND source_document_id = sqlc.arg('source_document_id')
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
ORDER BY
  created_at ASC;

-- name: GetPendingApprovalTransactions :many
SELECT
  *
FROM
  finance_transactions
WHERE
  tenant_id = current_tenant_id()
  AND approval_status = 'PENDING'
  AND deleted_at IS NULL
ORDER BY
  created_at ASC
LIMIT
  sqlc.arg('limit_count') OFFSET sqlc.arg('offset_count');

-- name: GetRecurringTransactionsDue :many
SELECT
  *
FROM
  finance_transactions
WHERE
  tenant_id = current_tenant_id()
  AND is_recurring = TRUE
  AND next_recurring_date <= sqlc.arg('due_date')
  AND transaction_status = 'POSTED'
  AND deleted_at IS NULL
ORDER BY
  next_recurring_date ASC;

-- name: UpdateRecurringTransactionNextDate :exec
UPDATE
  finance_transactions
SET
  next_recurring_date = sqlc.arg('next_recurring_date'),
  updated_at = NOW()
WHERE
  id = sqlc.arg('transaction_id')
  AND tenant_id = current_tenant_id()
  AND is_recurring = TRUE
  AND deleted_at IS NULL;

-- name: GetTransactionSummaryByPeriod :many
SELECT
  transaction_type,
  transaction_status,
  COUNT(*) AS transaction_count,
  SUM(total_debit_amount) AS total_debit,
  SUM(total_credit_amount) AS total_credit
FROM
  finance_transactions
WHERE
  tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND transaction_date >= sqlc.arg('date_from')
  AND transaction_date <= sqlc.arg('date_to')
GROUP BY
  transaction_type,
  transaction_status
ORDER BY
  transaction_type,
  transaction_status;

-- name: ValidateTransactionBalance :one
SELECT
  id,
  total_debit_amount,
  total_credit_amount,
  (total_debit_amount = total_credit_amount) AS is_balanced
FROM
  finance_transactions
WHERE
  id = sqlc.arg('transaction_id')
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: SearchTransactions :many
SELECT
  *
FROM
  finance_transactions
WHERE
  tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND (
    transaction_number ILIKE '%' || sqlc.arg('search_term') || '%'
    OR description ILIKE '%' || sqlc.arg('search_term') || '%'
    OR reference_number ILIKE '%' || sqlc.arg('search_term') || '%'
    OR external_reference ILIKE '%' || sqlc.arg('search_term') || '%'
    OR memo ILIKE '%' || sqlc.arg('search_term') || '%'
  )
ORDER BY
  CASE
    WHEN transaction_number ILIKE sqlc.arg('search_term') || '%' THEN 1
    ELSE 2
  END,
  transaction_date DESC
LIMIT
  sqlc.arg('limit_count') OFFSET sqlc.arg('offset_count');

-- =====================================================================
-- NEW QUERIES FOR ENHANCED FUNCTIONALITY
-- =====================================================================
-- name: UpdateTransactionMemo :exec
UPDATE
  finance_transactions
SET
  memo = sqlc.arg('memo'),
  updated_at = NOW(),
  updated_by = sqlc.arg('updated_by')
WHERE
  id = sqlc.arg('transaction_id')
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: AddTransactionAttachment :exec
UPDATE
  finance_transactions
SET
  attachment_ids = array_append(
    COALESCE(attachment_ids, '{}'),
    sqlc.arg('attachment_id')
  ),
  updated_at = NOW(),
  updated_by = sqlc.arg('updated_by')
WHERE
  id = sqlc.arg('transaction_id')
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: RemoveTransactionAttachment :exec
UPDATE
  finance_transactions
SET
  attachment_ids = array_remove(attachment_ids, sqlc.arg('attachment_id')),
  updated_at = NOW(),
  updated_by = sqlc.arg('updated_by')
WHERE
  id = sqlc.arg('transaction_id')
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: GetTransactionsByAttachment :many
SELECT
  *
FROM
  finance_transactions
WHERE
  tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND attachment_ids @> ARRAY [sqlc.arg('attachment_id')]::TEXT []
ORDER BY
  created_at DESC;

-- name: AddTransactionTag :exec
UPDATE
  finance_transactions
SET
  tags = array_append(COALESCE(tags, '{}'), sqlc.arg('tag')),
  updated_at = NOW(),
  updated_by = sqlc.arg('updated_by')
WHERE
  id = sqlc.arg('transaction_id')
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND NOT (tags @> ARRAY [sqlc.arg('tag')]::VARCHAR []);

-- name: RemoveTransactionTag :exec
UPDATE
  finance_transactions
SET
  tags = array_remove(tags, sqlc.arg('tag')),
  updated_at = NOW(),
  updated_by = sqlc.arg('updated_by')
WHERE
  id = sqlc.arg('transaction_id')
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: GetTransactionsByTag :many
SELECT
  *
FROM
  finance_transactions
WHERE
  tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND tags @> ARRAY [sqlc.arg('tag')]::VARCHAR []
ORDER BY
  created_at DESC
LIMIT
  sqlc.arg('limit_count') OFFSET sqlc.arg('offset_count');

-- name: GetTransactionsByTags :many
SELECT
  *
FROM
  finance_transactions
WHERE
  tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND tags && sqlc.arg('tag_array')::VARCHAR []
ORDER BY
  created_at DESC
LIMIT
  sqlc.arg('limit_count') OFFSET sqlc.arg('offset_count');

-- name: GetAllTransactionTags :many
SELECT
  DISTINCT unnest(tags) AS tag
FROM
  finance_transactions
WHERE
  tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND tags IS NOT NULL
ORDER BY
  tag;

-- name: SearchTransactionsByMemo :many
SELECT
  *
FROM
  finance_transactions
WHERE
  tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND memo ILIKE '%' || sqlc.arg('search_term') || '%'
ORDER BY
  transaction_date DESC
LIMIT
  sqlc.arg('limit_count') OFFSET sqlc.arg('offset_count');

-- name: GetTransactionsWithAttachments :many
SELECT
  id,
  transaction_number,
  description,
  transaction_date,
  total_debit_amount,
  total_credit_amount,
  array_length(attachment_ids, 1) AS attachment_count,
  attachment_ids
FROM
  finance_transactions
WHERE
  tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND attachment_ids IS NOT NULL
  AND array_length(attachment_ids, 1) > 0
ORDER BY
  transaction_date DESC
LIMIT
  sqlc.arg('limit_count') OFFSET sqlc.arg('offset_count');

-- name: BulkUpdateTransactionTags :exec
UPDATE
  finance_transactions
SET
  tags = sqlc.arg('tag_array')::VARCHAR [],
  updated_at = NOW(),
  updated_by = sqlc.arg('updated_by')
WHERE
  id = ANY(sqlc.arg('transaction_ids')::UUID [])
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: GetTransactionCountByTag :many
SELECT
  unnest(tags) AS tag,
  COUNT(*) AS transaction_count
FROM
  finance_transactions
WHERE
  tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND tags IS NOT NULL
GROUP BY
  unnest(tags)
ORDER BY
  transaction_count DESC;

-- name: UpdateTransactionAttributes :exec
UPDATE
  finance_transactions
SET
  transaction_attributes = sqlc.arg('attributes')::JSONB,
  updated_at = NOW(),
  updated_by = sqlc.arg('updated_by')
WHERE
  id = sqlc.arg('transaction_id')
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: GetTransactionsByAttribute :many
SELECT
  *
FROM
  finance_transactions
WHERE
  tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND transaction_attributes @> sqlc.arg('attribute_filter')::JSONB
ORDER BY
  created_at DESC
LIMIT
  sqlc.arg('limit_count') OFFSET sqlc.arg('offset_count');

-- name: CreateTransactionWithDefaults :one
INSERT INTO
  finance_transactions (
    tenant_id,
    entity_id,
    transaction_number,
    transaction_type,
    transaction_date,
    description,
    currency_code,
    created_by
  )
VALUES
  (
    current_tenant_id(),
    sqlc.arg('entity_id'),
    sqlc.arg('transaction_number'),
    sqlc.arg('transaction_type'),
    sqlc.arg('transaction_date'),
    sqlc.arg('description'),
    'USD',
    sqlc.arg('created_by')
  )
RETURNING
  *;

-- name: GetRecentTransactions :many
SELECT
  *
FROM
  finance_transactions
WHERE
  tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND created_at >= NOW() - INTERVAL '30 days'
ORDER BY
  created_at DESC
LIMIT
  sqlc.arg('limit_count');

-- name: GetTransactionActivity :many
SELECT
  DATE(created_at) AS activity_date,
  COUNT(*) AS transaction_count,
  SUM(total_debit_amount) AS daily_total_debit,
  SUM(total_credit_amount) AS daily_total_credit
FROM
  finance_transactions
WHERE
  tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND created_at >= sqlc.arg('date_from')
  AND created_at <= sqlc.arg('date_to')
GROUP BY
  DATE(created_at)
ORDER BY
  activity_date DESC;
