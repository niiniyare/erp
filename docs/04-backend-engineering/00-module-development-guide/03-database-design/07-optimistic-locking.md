---
title: Optimistic Locking
portal: 4 — Backend Engineering
section: 00-module-development-guide/03-database-design
audience: [backend-engineer, tech-lead]
related:
  - path: ./02-primary-table.md
    title: Primary Table Migration
  - path: ../04-sqlc-queries/04-update-queries.md
    title: Update Queries
  - path: ../05-repository-layer/04-error-mapping.md
    title: Error Mapping
---

# Optimistic Locking

Optimistic locking prevents lost updates when two users edit the same record concurrently. AwoERP uses a `version` counter pattern — no row-level locks, no serialisable transactions.

## How It Works

1. Client fetches a contract. Response includes `"version": 3`.
2. Client edits the contract in the UI.
3. Client submits the edit, including `"version": 3` in the request body.
4. Service passes `version: 3` to the repository's `Update` call.
5. Repository executes: `UPDATE contracts SET ... WHERE id = $1 AND version = 3`.
   - If the record still has `version = 3`: update succeeds, `version` becomes `4`.
   - If another writer incremented `version` to `4` first: 0 rows affected → `ErrContractConflict`.
6. Handler returns `409 Conflict` to the client.
7. Client re-fetches and re-applies the edit.

No locks held between steps 1 and 5. The conflict only surfaces at write time.

## Schema

```sql
version integer NOT NULL DEFAULT 1
```

Every table has this column. Default is `1` (not `0`) — a version of `0` looks like "uninitialised" to readers.

## SQLC Update Query Pattern

```sql
-- name: UpdateContract :one
UPDATE contracts
SET
    title       = @title,
    description = @description,
    total_value = @total_value,
    -- ... other updatable fields ...
    version     = version + 1,
    updated_by  = @updated_by,
    updated_at  = now()
WHERE id        = @id
  AND tenant_id = @tenant_id
  AND version   = @version          -- optimistic lock check
  AND deleted_at IS NULL
RETURNING *;
```

The `WHERE version = @version` clause is the lock check. If zero rows are affected, the version was already advanced by a concurrent writer.

## Repository Error Mapping

When zero rows are affected, the repository must return `domain.ErrContractConflict`:

```go
// internal/core/contracts/repository/contract_sqlc.go
func (r *contractSQLCRepository) Update(ctx context.Context, params UpdateContractParams) (*domain.Contract, error) {
    var row db.Contract
    err := r.store.WithTenant(ctx, params.TenantID, func(q *db.Queries) error {
        var err error
        row, err = q.UpdateContract(ctx, db.UpdateContractParams{
            ID:          params.ID,
            TenantID:    params.TenantID,
            Title:       params.Title,
            Description: params.Description,
            // ...
            Version:     int32(params.Version),
            UpdatedBy:   params.UpdatedBy,
        })
        return err
    })
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            // Could be: contract does not exist, OR version mismatch.
            // Disambiguate by checking existence.
            exists, _ := r.exists(ctx, params.ID, params.TenantID)
            if exists {
                return nil, domain.ErrContractConflict
            }
            return nil, domain.ErrContractNotFound
        }
        return nil, fmt.Errorf("UpdateContract: %w", err)
    }
    return mapRowToDomain(row), nil
}
```

`pgx.ErrNoRows` is returned by `pgx` when RETURNING gives back zero rows. Since both "not found" and "version conflict" produce zero rows, the repository checks existence to distinguish them.

## Handler Response for Conflict

```go
// handlers/errors.go
case errors.Is(err, domain.ErrContractConflict):
    return fiber.NewError(fiber.StatusConflict,
        "contract was modified by another user; please reload and try again")
```

The UI on receiving a `409` should:
1. Re-fetch the contract to get the current state
2. Re-apply the user's changes on top of the current state
3. Re-submit with the new version number

## Status-Only Updates

Status transitions use a separate update query and also require version:

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
  AND version   = @version          -- still required for status updates
  AND deleted_at IS NULL
RETURNING *;
```

Status updates also check version. Two concurrent approval attempts for the same contract will conflict — only the first one wins.

## Version in the Response DTO

Every response DTO for a versioned entity must include `version`:

```go
// handlers/response.go
type ContractResponse struct {
    ID             string `json:"id"`
    // ...
    Version        int    `json:"version"`   // client must echo this on update
    // ...
}
```

If `version` is missing from the response, the client cannot perform updates — it has no version to send.

## What Version Does NOT Protect Against

**Sequential writes by the same user:** If a user opens two tabs and edits the same contract, the second tab's fetch gets the post-edit version, so both edits can succeed. This is correct behaviour — the second edit intentionally replaces the first.

**Delete conflicts:** Soft delete (`deleted_at = now()`) also uses `WHERE version = @version`. A concurrent delete and update will conflict correctly.

**Create conflicts:** The unique constraint on `(tenant_id, contract_number)` handles concurrent creates with the same number, returning `ErrContractAlreadyExists`.
