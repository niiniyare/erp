> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Soft Delete
portal: 4 — Backend Engineering
section: 00-module-development-guide/03-database-design
audience: [backend-engineer, tech-lead]
related:
  - path: ./02-primary-table.md
    title: Primary Table Migration
  - path: ../04-sqlc-queries/05-delete-queries.md
    title: Delete Queries
---

# Soft Delete

AwoERP never hard-deletes business records. Every business table uses soft delete: a `deleted_at timestamptz` column that is set to the deletion timestamp instead of removing the row.

## Why Soft Delete

- **Audit integrity**: finance and compliance records must never disappear.
- **Accidental deletion recovery**: support can restore a soft-deleted record.
- **Cross-reference safety**: other modules may hold FKs to this record. Hard delete breaks those references.
- **Event sourcing**: deleted records can still be replayed in event streams.

## Column Definition

```sql
deleted_at timestamptz,  -- nullable; NOT NULL not permitted
```

`deleted_at` is always nullable. A NULL value means the record is live. A non-NULL value means the record was soft-deleted at that timestamp.

Never add `NOT NULL` to `deleted_at`. A non-null `deleted_at` on every row would make every record appear deleted.

## Soft Delete Query

```sql
-- name: SoftDeleteContract :one
UPDATE contracts
SET
    deleted_at = now(),
    updated_by = @updated_by,
    updated_at = now()
WHERE id        = @id
  AND tenant_id = @tenant_id
  AND version   = @version           -- optimistic lock: concurrent deletes conflict
  AND deleted_at IS NULL             -- idempotency: already-deleted is a no-op
RETURNING *;
```

`AND deleted_at IS NULL` on the delete query means trying to delete an already-deleted record returns zero rows (treated as not found). This is correct — double-delete is not an error the caller needs to handle specially.

## Filtering Soft-Deleted Rows

Every SELECT query must filter out soft-deleted rows:

```sql
-- name: GetContractByID :one
SELECT * FROM contracts
WHERE id        = @id
  AND tenant_id = @tenant_id
  AND deleted_at IS NULL;           -- always present
```

```sql
-- name: ListContracts :many
SELECT * FROM contracts
WHERE tenant_id = @tenant_id
  AND deleted_at IS NULL            -- always present
  -- optional filters follow
```

If `deleted_at IS NULL` is omitted, deleted records appear in results. Partial indexes (`WHERE deleted_at IS NULL`) on the table mean the planner uses the index for these queries.

## Archive Query (admin only)

For admin screens that need to show deleted records:

```sql
-- name: ListDeletedContracts :many
SELECT * FROM contracts
WHERE tenant_id  = @tenant_id
  AND deleted_at IS NOT NULL        -- deleted records only
ORDER BY deleted_at DESC
LIMIT @limit OFFSET @offset;
```

This query is separate and explicit — normal queries cannot accidentally fetch deleted records.

## Cascade Soft Delete

When a parent is soft-deleted, child records should also be soft-deleted. The service handles this, not a DB trigger:

```go
// service/contract.go
func (s *contractService) Delete(ctx context.Context, id, tenantID, deletedBy uuid.UUID, version int) error {
    // 1. Soft-delete all lines first
    if err := s.lineRepo.SoftDeleteByContractID(ctx, id, tenantID, deletedBy); err != nil {
        return err
    }
    // 2. Soft-delete the contract
    if _, err := s.repo.Delete(ctx, id, tenantID, deletedBy, version); err != nil {
        return err
    }
    return nil
}
```

Run both operations inside a transaction to prevent partial cascades. The repository methods should accept a transaction context via `store.WithTenant`, which runs inside a `BEGIN`/`COMMIT` block.

## Entity Helper Method

The domain entity has a helper that services and handlers use:

```go
// domain/contract.go
func (c *Contract) IsDeleted() bool {
    return c.DeletedAt != nil
}
```

Use `contract.IsDeleted()` in service code rather than `contract.DeletedAt != nil` directly — it reads as domain language.

## Hard Delete (data retention / GDPR)

For regulatory data erasure (GDPR right to erasure), a separate maintenance operation uses true hard delete:

```sql
-- Run by a scheduled job, not application-request code
DELETE FROM contracts
WHERE tenant_id  = $1
  AND deleted_at < now() - INTERVAL '7 years'
  AND deleted_at IS NOT NULL;
```

This is never called from the HTTP API layer. It runs via a scheduled maintenance task with explicit operator approval.
