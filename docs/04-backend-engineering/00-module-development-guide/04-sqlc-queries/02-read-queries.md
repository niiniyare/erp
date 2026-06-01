---
title: Read Queries
portal: 4 — Backend Engineering
section: 00-module-development-guide/04-sqlc-queries
audience: [backend-engineer, tech-lead]
related:
  - path: ./01-query-overview.md
    title: Query Overview
  - path: ./03-list-queries.md
    title: List Queries
---

# Read Queries

Read queries fetch single records by primary key or unique constraint. They always filter by `tenant_id` and exclude soft-deleted rows.

## GetContractByID

```sql
-- name: GetContractByID :one
SELECT
    id,
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
    version,
    created_by,
    updated_by,
    deleted_at,
    created_at,
    updated_at
FROM contracts
WHERE id        = @id
  AND tenant_id = @tenant_id
  AND deleted_at IS NULL;
```

### Key Rules

**Enumerate columns explicitly** — do not use `SELECT *`. When a migration adds a column, `SELECT *` silently changes the return type. Explicit column lists fail at `make sqlc` time if a column is removed, giving a compile-time error rather than a runtime one.

**Both id AND tenant_id in WHERE** — the RLS policy enforces the tenant boundary in the database, but the explicit `tenant_id = @tenant_id` in the query provides defense in depth and makes the query plan use the `(id, tenant_id)` index more efficiently.

**`deleted_at IS NULL`** — never return soft-deleted rows from a normal read query.

## GetContractByNumber

```sql
-- name: GetContractByNumber :one
SELECT
    id, tenant_id, entity_id, contract_number, title, description,
    vendor_id, contract_type, start_date, end_date, total_value, currency,
    status, version, created_by, updated_by, deleted_at, created_at, updated_at
FROM contracts
WHERE contract_number = @contract_number
  AND tenant_id       = @tenant_id
  AND deleted_at      IS NULL;
```

Used by the existence check and when resolving contracts by their human-readable number.

## ContractExistsWithNumber

```sql
-- name: ContractExistsWithNumber :one
SELECT EXISTS (
    SELECT 1 FROM contracts
    WHERE contract_number = @contract_number
      AND tenant_id       = @tenant_id
      AND deleted_at      IS NULL
) AS exists;
```

Returns a single boolean. Called by the service before create to check for duplicates before attempting an INSERT (which would fail with a unique constraint violation).

The service also handles the unique constraint violation from the DB as `ErrContractAlreadyExists` — this is the belt-and-suspenders check.

## GetContractLinesByContractID

```sql
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
```

Returns all lines for a contract, ordered by line number. Used when fetching the full contract detail view.

## Error Handling for Read Queries

The repository maps `pgx.ErrNoRows` to the module's not-found sentinel:

```go
func (r *contractSQLCRepository) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.Contract, error) {
    var row db.GetContractByIDRow
    err := r.store.WithTenant(ctx, tenantID, func(q *db.Queries) error {
        var e error
        row, e = q.GetContractByID(ctx, db.GetContractByIDParams{
            ID:       id,
            TenantID: tenantID,
        })
        return e
    })
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, domain.ErrContractNotFound
        }
        return nil, fmt.Errorf("GetContractByID: %w", err)
    }
    return mapRowToDomain(row), nil
}
```

`pgx.ErrNoRows` is a clean sentinel that `pgx` returns when a `:one` query returns zero rows. It is never returned for `:many` queries (which return an empty slice).
