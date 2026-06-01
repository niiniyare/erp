---
title: Delete Queries
portal: 4 — Backend Engineering
section: 00-module-development-guide/04-sqlc-queries
audience: [backend-engineer, tech-lead]
related:
  - path: ./04-update-queries.md
    title: Update Queries
  - path: ../03-database-design/08-soft-delete.md
    title: Soft Delete
---

# Delete Queries

AwoERP uses soft delete. "Delete" queries update `deleted_at` — they never run `DELETE FROM`. Hard DELETE is only used by scheduled maintenance scripts, never by the application layer.

## SoftDeleteContract

```sql
-- name: SoftDeleteContract :one
UPDATE contracts
SET
    deleted_at = now(),
    updated_by = @updated_by,
    updated_at = now()
WHERE id        = @id
  AND tenant_id = @tenant_id
  AND version   = @version
  AND deleted_at IS NULL            -- idempotency guard
RETURNING
    id, tenant_id, entity_id, contract_number, title, description,
    vendor_id, contract_type, start_date, end_date, total_value, currency,
    status, version, created_by, updated_by, deleted_at, created_at, updated_at;
```

### Idempotency Guard

`AND deleted_at IS NULL` prevents double-deleting. If the record is already soft-deleted:
- The query returns zero rows (`pgx.ErrNoRows`).
- The repository maps this to `ErrContractNotFound`.
- The handler returns `404`.

This is correct — a deleted record "does not exist" from the caller's perspective.

### Version Check on Delete

`AND version = @version` is required for delete. Without it:
- User A opens the contract, sees version 3.
- User B edits the contract, it becomes version 4.
- User A deletes with version 3 — should fail (stale view).

With the version check, User A's delete returns a conflict error and must re-fetch.

## SoftDeleteContractLinesByContractID

```sql
-- name: SoftDeleteContractLinesByContractID :exec
UPDATE contract_lines
SET
    deleted_at = now(),
    updated_by = @updated_by,
    updated_at = now()
WHERE contract_id  = @contract_id
  AND tenant_id    = @tenant_id
  AND deleted_at   IS NULL;
```

Used by the service when soft-deleting a contract (cascades to lines). This query uses `:exec` (not `:one`) because it affects zero or more rows — not checking RETURNING. The service calls this before deleting the parent contract.

No version check here — bulk child delete on parent deletion is not a user-facing edit; it is a system cascade.

## SoftDeleteContractLine (single line)

```sql
-- name: SoftDeleteContractLine :one
UPDATE contract_lines
SET
    deleted_at = now(),
    updated_by = @updated_by,
    updated_at = now()
WHERE id         = @id
  AND contract_id = @contract_id
  AND tenant_id   = @tenant_id
  AND version     = @version
  AND deleted_at  IS NULL
RETURNING
    id, tenant_id, contract_id, line_number, description,
    quantity, unit_price, total_price, status, version,
    created_by, updated_by, deleted_at, created_at, updated_at;
```

Deleting a single line requires the version check (user is deleting a specific line they were viewing). After the line is deleted, the service also calls `UpdateContractTotalValue` to recalculate the contract total.

## Hard Delete (maintenance only)

```sql
-- name: HardDeleteExpiredContracts :execresult
-- Used by scheduled maintenance only. Never called from HTTP handlers.
DELETE FROM contracts
WHERE tenant_id  = @tenant_id
  AND deleted_at < now() - @retention_period
  AND deleted_at IS NOT NULL;
```

Using `:execresult` returns a `pgconn.CommandTag` so the caller can log how many rows were deleted. This query is only invoked from a maintenance command, never from the request path.
