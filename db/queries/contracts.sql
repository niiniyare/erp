-- =====================================================
-- CONTRACT QUERIES (RLS-aware: current_tenant_id())
-- =====================================================

-- name: CreateContract :one
INSERT INTO contracts (
    tenant_id,
    entity_id,
    number,
    title,
    status,
    contract_type,
    counterparty_name,
    counterparty_email,
    start_date,
    end_date,
    value,
    currency_code,
    description,
    terms,
    metadata,
    created_by
) VALUES (
    current_tenant_id(),
    @entity_id,
    @number,
    @title,
    @status,
    @contract_type,
    @counterparty_name,
    sqlc.narg('counterparty_email'),
    @start_date,
    sqlc.narg('end_date'),
    @value,
    @currency_code,
    sqlc.narg('description'),
    sqlc.narg('terms'),
    @metadata,
    sqlc.narg('created_by')
)
RETURNING *;

-- name: GetContractByID :one
SELECT * FROM contracts
WHERE id = @id
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: GetContractByNumber :one
SELECT * FROM contracts
WHERE number = @number
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: GetContractByIDForUpdate :one
SELECT * FROM contracts
WHERE id = @id
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
FOR UPDATE;

-- name: UpdateContract :one
UPDATE contracts SET
    title              = COALESCE(sqlc.narg('title'), title),
    status             = COALESCE(sqlc.narg('status'), status),
    contract_type      = COALESCE(sqlc.narg('contract_type'), contract_type),
    counterparty_name  = COALESCE(sqlc.narg('counterparty_name'), counterparty_name),
    counterparty_email = COALESCE(sqlc.narg('counterparty_email'), counterparty_email),
    start_date         = COALESCE(sqlc.narg('start_date'), start_date),
    end_date           = COALESCE(sqlc.narg('end_date'), end_date),
    value              = COALESCE(sqlc.narg('value'), value),
    currency_code      = COALESCE(sqlc.narg('currency_code'), currency_code),
    description        = COALESCE(sqlc.narg('description'), description),
    terms              = COALESCE(sqlc.narg('terms'), terms),
    signed_by          = COALESCE(sqlc.narg('signed_by'), signed_by),
    signed_at          = COALESCE(sqlc.narg('signed_at'), signed_at),
    metadata           = COALESCE(sqlc.narg('metadata'), metadata),
    version            = version + 1,
    updated_at         = NOW()
WHERE id = @id
  AND tenant_id = current_tenant_id()
  AND version   = @version
  AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteContract :exec
UPDATE contracts SET
    deleted_at = NOW(),
    updated_at = NOW()
WHERE id = @id
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: ListContracts :many
SELECT * FROM contracts
WHERE tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND (sqlc.narg('entity_id')::uuid IS NULL       OR entity_id       = sqlc.narg('entity_id'))
  AND (sqlc.narg('status')::text   IS NULL       OR status           = sqlc.narg('status'))
  AND (sqlc.narg('contract_type')::text IS NULL  OR contract_type    = sqlc.narg('contract_type'))
  AND (sqlc.narg('search')::text IS NULL
       OR title              ILIKE '%' || sqlc.narg('search') || '%'
       OR counterparty_name  ILIKE '%' || sqlc.narg('search') || '%'
       OR number             ILIKE '%' || sqlc.narg('search') || '%')
ORDER BY created_at DESC
LIMIT  sqlc.arg('limit_count')
OFFSET sqlc.arg('offset_count');

-- name: CountContracts :one
SELECT COUNT(*) FROM contracts
WHERE tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND (sqlc.narg('entity_id')::uuid IS NULL       OR entity_id       = sqlc.narg('entity_id'))
  AND (sqlc.narg('status')::text   IS NULL       OR status           = sqlc.narg('status'))
  AND (sqlc.narg('contract_type')::text IS NULL  OR contract_type    = sqlc.narg('contract_type'))
  AND (sqlc.narg('search')::text IS NULL
       OR title              ILIKE '%' || sqlc.narg('search') || '%'
       OR counterparty_name  ILIKE '%' || sqlc.narg('search') || '%'
       OR number             ILIKE '%' || sqlc.narg('search') || '%');

-- name: CheckContractNumberExists :one
SELECT EXISTS(
    SELECT 1 FROM contracts
    WHERE number = @number
      AND tenant_id = current_tenant_id()
      AND deleted_at IS NULL
);
