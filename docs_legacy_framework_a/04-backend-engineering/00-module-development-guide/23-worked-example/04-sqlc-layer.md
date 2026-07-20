> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Worked Example — SQLC Layer
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[SQLC Layer Overview](../04-sqlc-layer/01-sqlc-overview.md)"
  - "[SQLC Advanced Patterns](../04-sqlc-layer/02-sqlc-advanced.md)"
  - "[Worked Example Overview](01-worked-example-overview.md)"
---

# Worked Example — SQLC Layer

## db/queries/contracts.sql (abridged)

```sql
-- ============================================================
-- CONTRACTS
-- ============================================================

-- name: GetContractByID :one
SELECT * FROM contracts
WHERE id = @id
  AND tenant_id = @tenant_id
  AND deleted_at IS NULL
LIMIT 1;

-- name: GetContractByNumber :one
SELECT * FROM contracts
WHERE contract_number = @contract_number
  AND tenant_id = @tenant_id
  AND deleted_at IS NULL
LIMIT 1;

-- name: ListContracts :many
SELECT * FROM contracts
WHERE tenant_id = @tenant_id
  AND deleted_at IS NULL
  AND (sqlc.narg('status')::text   IS NULL OR status    = sqlc.narg('status'))
  AND (sqlc.narg('entity_id')::uuid IS NULL OR entity_id = sqlc.narg('entity_id'))
ORDER BY created_at DESC
LIMIT @limit_count
OFFSET @offset_count;

-- name: CountContracts :one
SELECT COUNT(*) FROM contracts
WHERE tenant_id = @tenant_id
  AND deleted_at IS NULL
  AND (sqlc.narg('status')::text   IS NULL OR status    = sqlc.narg('status'))
  AND (sqlc.narg('entity_id')::uuid IS NULL OR entity_id = sqlc.narg('entity_id'));

-- name: CreateContract :one
INSERT INTO contracts (
    tenant_id, entity_id, contract_number, title, description,
    status, contract_type, total_value, currency,
    start_date, end_date, vendor_id, assigned_to,
    created_by, updated_by
) VALUES (
    @tenant_id, @entity_id, @contract_number, @title, @description,
    'draft', @contract_type, @total_value, @currency,
    @start_date, @end_date, @vendor_id, @assigned_to,
    @created_by, @created_by
)
RETURNING *;

-- name: UpdateContract :one
UPDATE contracts SET
    title       = @title,
    description = @description,
    total_value = @total_value,
    currency    = @currency,
    start_date  = @start_date,
    end_date    = @end_date,
    vendor_id   = @vendor_id,
    assigned_to = @assigned_to,
    updated_by  = @updated_by,
    version     = version + 1
WHERE id        = @id
  AND tenant_id = @tenant_id
  AND version   = @version
  AND deleted_at IS NULL
RETURNING *;

-- name: UpdateContractStatus :one
UPDATE contracts SET
    status     = @status,
    updated_by = @updated_by,
    version    = version + 1
WHERE id        = @id
  AND tenant_id = @tenant_id
  AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteContract :one
UPDATE contracts SET
    deleted_at = now(),
    updated_by = @updated_by,
    version    = version + 1
WHERE id        = @id
  AND tenant_id = @tenant_id
  AND deleted_at IS NULL
  AND version   = @version
RETURNING *;

-- name: ContractExistsWithNumber :one
SELECT EXISTS(
    SELECT 1 FROM contracts
    WHERE tenant_id = @tenant_id
      AND contract_number = @contract_number
      AND deleted_at IS NULL
) AS exists;

-- ============================================================
-- CONTRACT LINES
-- ============================================================

-- name: ListContractLinesByContract :many
SELECT * FROM contract_lines
WHERE contract_id = @contract_id
  AND tenant_id   = @tenant_id
  AND deleted_at IS NULL
ORDER BY line_number ASC;

-- name: CreateContractLine :one
INSERT INTO contract_lines (
    tenant_id, contract_id, line_number, description,
    quantity, unit_price, unit, created_by, updated_by
) VALUES (
    @tenant_id, @contract_id, @line_number, @description,
    @quantity, @unit_price, @unit, @created_by, @created_by
)
RETURNING *;

-- name: SoftDeleteContractLinesByContractID :exec
UPDATE contract_lines SET
    deleted_at = now(),
    updated_by = @updated_by
WHERE contract_id = @contract_id
  AND tenant_id   = @tenant_id
  AND deleted_at IS NULL;
```

**Gate**: After writing queries, run `make sqlc` to generate Go code. Fix any type errors before continuing.
