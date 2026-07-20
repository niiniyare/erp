> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Update Queries
portal: 4 — Backend Engineering
section: 00-module-development-guide/04-sqlc-queries
audience: [backend-engineer, tech-lead]
related:
  - path: ./05-delete-queries.md
    title: Delete Queries
  - path: ../03-database-design/07-optimistic-locking.md
    title: Optimistic Locking
---

# Update Queries

Update queries always include the optimistic lock check (`AND version = @version`) and always use `RETURNING *` to return the updated row without a second SELECT.

## Create Query

```sql
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
    'draft',           -- status is always 'draft' on create — not a parameter
    @created_by,
    @created_by        -- updated_by = created_by on initial insert
)
RETURNING
    id, tenant_id, entity_id, contract_number, title, description,
    vendor_id, contract_type, start_date, end_date, total_value, currency,
    status, version, created_by, updated_by, deleted_at, created_at, updated_at;
```

### Why status is hardcoded to 'draft'

The initial status is not a parameter — it is always `'draft'`. This is a domain invariant encoded in the query. If the repository allowed `status` to be passed at create time, a future programmer might pass an arbitrary status from the handler, bypassing the state machine.

## Full Field Update

```sql
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
```

Fields that must never change after creation are **not** in the SET clause:
- `id` — never changes
- `tenant_id` — never changes
- `entity_id` — never changes (re-assigning org unit requires explicit reparent operation)
- `contract_number` — never changes after creation
- `status` — changed only via `UpdateContractStatus`
- `created_by`, `created_at` — never change

## Status-Only Update

```sql
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
```

Status updates are a separate query from field updates. This separation ensures the service always goes through `CanTransitionTo()` before calling `UpdateContractStatus`. If field updates and status changes were in one query, it would be possible to update both simultaneously in a way that bypasses the state machine check.

## Update Contract Total Value (after line changes)

```sql
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
```

Called by the service after adding/updating/removing a contract line. This query does not require the caller to pass the new total — it re-computes from the current lines, which avoids race conditions where two concurrent line additions each read the old total and add their own amount.

Note: this query does NOT have `AND version = @version`. Total value rollup is not a user-facing edit — it should always succeed. The version increment still happens to reflect the change.

## RETURNING Clause

Every write query uses `RETURNING` with the full column list. This allows:

1. The repository returns the persisted entity immediately — no second SELECT needed.
2. The returned `version` reflects the new (incremented) value — the client receives the correct version for its next update.
3. The returned `status` reflects any DB-side defaulting.

Always use the same column list in RETURNING as in SELECT queries for the same table. If the column list drifts, the repository mapping function will need separate paths for read vs write.
