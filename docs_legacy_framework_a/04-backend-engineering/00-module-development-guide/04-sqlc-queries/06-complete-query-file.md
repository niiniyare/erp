> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Complete Query File
portal: 4 — Backend Engineering
section: 00-module-development-guide/04-sqlc-queries
audience: [backend-engineer, tech-lead]
related:
  - path: ./01-query-overview.md
    title: Query Overview
  - path: ../05-repository-layer/01-repository-interface.md
    title: Repository Interface
---

# Complete Query File

The full `db/queries/contracts.sql` for the contracts module. Copy this as the starting point for a new module and rename accordingly.

## db/queries/contracts.sql

```sql
-- =============================================================
-- contracts.sql
-- SQLC annotated queries for the contracts module.
-- Module group: 011
-- Run `make sqlc` after any change to regenerate Go code.
-- =============================================================

-- ============================================================
-- contracts
-- ============================================================

-- name: CreateContract :one
INSERT INTO contracts (
    tenant_id,
    entity_id,
    contract_number,
    title,
    description,
    vendor_id,
    contract_type,
    start_date,
    end_date,
    total_value,
    currency,
    status,
    created_by,
    updated_by
) VALUES (
    @tenant_id,
    @entity_id,
    @contract_number,
    @title,
    @description,
    @vendor_id,
    @contract_type,
    @start_date,
    @end_date,
    @total_value,
    @currency,
    'draft',
    @created_by,
    @created_by
)
RETURNING
    id, tenant_id, entity_id, contract_number, title, description,
    vendor_id, contract_type, start_date, end_date, total_value, currency,
    status, version, created_by, updated_by, deleted_at, created_at, updated_at;

-- name: GetContractByID :one
SELECT
    id, tenant_id, entity_id, contract_number, title, description,
    vendor_id, contract_type, start_date, end_date, total_value, currency,
    status, version, created_by, updated_by, deleted_at, created_at, updated_at
FROM contracts
WHERE id        = @id
  AND tenant_id = @tenant_id
  AND deleted_at IS NULL;

-- name: GetContractByNumber :one
SELECT
    id, tenant_id, entity_id, contract_number, title, description,
    vendor_id, contract_type, start_date, end_date, total_value, currency,
    status, version, created_by, updated_by, deleted_at, created_at, updated_at
FROM contracts
WHERE contract_number = @contract_number
  AND tenant_id       = @tenant_id
  AND deleted_at      IS NULL;

-- name: ContractExistsWithNumber :one
SELECT EXISTS (
    SELECT 1 FROM contracts
    WHERE contract_number = @contract_number
      AND tenant_id       = @tenant_id
      AND deleted_at      IS NULL
) AS exists;

-- name: ListContracts :many
SELECT
    id, tenant_id, entity_id, contract_number, title, description,
    vendor_id, contract_type, start_date, end_date, total_value, currency,
    status, version, created_by, updated_by, deleted_at, created_at, updated_at
FROM contracts
WHERE tenant_id   = @tenant_id
  AND deleted_at  IS NULL
  AND (sqlc.narg(entity_id)::uuid       IS NULL OR entity_id     = sqlc.narg(entity_id))
  AND (sqlc.narg(vendor_id)::uuid       IS NULL OR vendor_id     = sqlc.narg(vendor_id))
  AND (sqlc.narg(status)::VARCHAR       IS NULL OR status        = sqlc.narg(status))
  AND (sqlc.narg(contract_type)::VARCHAR IS NULL OR contract_type = sqlc.narg(contract_type))
  AND (
        sqlc.narg(search)::TEXT IS NULL
        OR title           ILIKE '%' || sqlc.narg(search) || '%'
        OR contract_number ILIKE '%' || sqlc.narg(search) || '%'
      )
ORDER BY
    CASE WHEN @sort_by = 'created_at'  AND @sort_dir = 'desc' THEN created_at  END DESC,
    CASE WHEN @sort_by = 'created_at'  AND @sort_dir = 'asc'  THEN created_at  END ASC,
    CASE WHEN @sort_by = 'total_value' AND @sort_dir = 'desc' THEN total_value END DESC,
    CASE WHEN @sort_by = 'total_value' AND @sort_dir = 'asc'  THEN total_value END ASC,
    created_at DESC
LIMIT  @page_size
OFFSET @page_offset;

-- name: CountContracts :one
SELECT COUNT(*) AS total
FROM contracts
WHERE tenant_id  = @tenant_id
  AND deleted_at IS NULL
  AND (sqlc.narg(entity_id)::uuid       IS NULL OR entity_id     = sqlc.narg(entity_id))
  AND (sqlc.narg(vendor_id)::uuid       IS NULL OR vendor_id     = sqlc.narg(vendor_id))
  AND (sqlc.narg(status)::VARCHAR       IS NULL OR status        = sqlc.narg(status))
  AND (sqlc.narg(contract_type)::VARCHAR IS NULL OR contract_type = sqlc.narg(contract_type))
  AND (
        sqlc.narg(search)::TEXT IS NULL
        OR title           ILIKE '%' || sqlc.narg(search) || '%'
        OR contract_number ILIKE '%' || sqlc.narg(search) || '%'
      );

-- name: UpdateContract :one
UPDATE contracts
SET
    title           = @title,
    description     = @description,
    vendor_id       = @vendor_id,
    contract_type   = @contract_type,
    start_date      = @start_date,
    end_date        = @end_date,
    total_value     = @total_value,
    currency        = @currency,
    version         = version + 1,
    updated_by      = @updated_by,
    updated_at      = now()
WHERE id        = @id
  AND tenant_id = @tenant_id
  AND version   = @version
  AND deleted_at IS NULL
RETURNING
    id, tenant_id, entity_id, contract_number, title, description,
    vendor_id, contract_type, start_date, end_date, total_value, currency,
    status, version, created_by, updated_by, deleted_at, created_at, updated_at;

-- name: UpdateContractStatus :one
UPDATE contracts
SET
    status     = @status,
    version    = version + 1,
    updated_by = @updated_by,
    updated_at = now()
WHERE id        = @id
  AND tenant_id = @tenant_id
  AND version   = @version
  AND deleted_at IS NULL
RETURNING
    id, tenant_id, entity_id, contract_number, title, description,
    vendor_id, contract_type, start_date, end_date, total_value, currency,
    status, version, created_by, updated_by, deleted_at, created_at, updated_at;

-- name: UpdateContractTotalValue :one
UPDATE contracts
SET
    total_value = (
        SELECT COALESCE(SUM(total_price), 0)
        FROM contract_lines
        WHERE contract_id = @id
          AND tenant_id   = @tenant_id
          AND deleted_at  IS NULL
    ),
    version    = version + 1,
    updated_by = @updated_by,
    updated_at = now()
WHERE id        = @id
  AND tenant_id = @tenant_id
  AND deleted_at IS NULL
RETURNING
    id, tenant_id, entity_id, contract_number, title, description,
    vendor_id, contract_type, start_date, end_date, total_value, currency,
    status, version, created_by, updated_by, deleted_at, created_at, updated_at;

-- name: SoftDeleteContract :one
UPDATE contracts
SET
    deleted_at = now(),
    updated_by = @updated_by,
    updated_at = now()
WHERE id        = @id
  AND tenant_id = @tenant_id
  AND version   = @version
  AND deleted_at IS NULL
RETURNING
    id, tenant_id, entity_id, contract_number, title, description,
    vendor_id, contract_type, start_date, end_date, total_value, currency,
    status, version, created_by, updated_by, deleted_at, created_at, updated_at;

-- ============================================================
-- contract_lines
-- ============================================================

-- name: CreateContractLine :one
INSERT INTO contract_lines (
    tenant_id,
    contract_id,
    line_number,
    description,
    quantity,
    unit_price,
    status,
    created_by,
    updated_by
) VALUES (
    @tenant_id,
    @contract_id,
    @line_number,
    @description,
    @quantity,
    @unit_price,
    'active',
    @created_by,
    @created_by
)
RETURNING
    id, tenant_id, contract_id, line_number, description,
    quantity, unit_price, total_price, status, version,
    created_by, updated_by, deleted_at, created_at, updated_at;

-- name: ListContractLinesByContract :many
SELECT
    id, tenant_id, contract_id, line_number, description,
    quantity, unit_price, total_price, status, version,
    created_by, updated_by, deleted_at, created_at, updated_at
FROM contract_lines
WHERE contract_id = @contract_id
  AND tenant_id   = @tenant_id
  AND deleted_at  IS NULL
ORDER BY line_number ASC;

-- name: UpdateContractLine :one
UPDATE contract_lines
SET
    description = @description,
    quantity    = @quantity,
    unit_price  = @unit_price,
    version     = version + 1,
    updated_by  = @updated_by,
    updated_at  = now()
WHERE id          = @id
  AND contract_id = @contract_id
  AND tenant_id   = @tenant_id
  AND version     = @version
  AND deleted_at  IS NULL
RETURNING
    id, tenant_id, contract_id, line_number, description,
    quantity, unit_price, total_price, status, version,
    created_by, updated_by, deleted_at, created_at, updated_at;

-- name: SoftDeleteContractLine :one
UPDATE contract_lines
SET
    deleted_at = now(),
    updated_by = @updated_by,
    updated_at = now()
WHERE id          = @id
  AND contract_id = @contract_id
  AND tenant_id   = @tenant_id
  AND version     = @version
  AND deleted_at  IS NULL
RETURNING
    id, tenant_id, contract_id, line_number, description,
    quantity, unit_price, total_price, status, version,
    created_by, updated_by, deleted_at, created_at, updated_at;

-- name: SoftDeleteContractLinesByContractID :exec
UPDATE contract_lines
SET
    deleted_at = now(),
    updated_by = @updated_by,
    updated_at = now()
WHERE contract_id = @contract_id
  AND tenant_id   = @tenant_id
  AND deleted_at  IS NULL;

-- name: NextContractLineNumber :one
SELECT COALESCE(MAX(line_number), 0) + 1 AS next_line_number
FROM contract_lines
WHERE contract_id = @contract_id
  AND tenant_id   = @tenant_id;
```
