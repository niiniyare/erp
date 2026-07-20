> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: SQLC Layer Overview
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer]
related:
  - "[Database Layer](../03-database-layer/01-database-overview.md)"
  - "[Repository Layer](../05-repository-layer/01-repository-overview.md)"
  - "[Worked Example: SQLC Layer](../23-worked-example/04-sqlc-layer.md)"
---

# SQLC Layer Overview

## What SQLC Does

SQLC reads your SQL query files and generates:
- Strongly-typed Go structs matching query result rows
- Type-safe query functions with exact parameter and return types
- No reflection, no ORM magic — compiles to direct pgx calls

## Configuration

`sqlc.yaml` in the repo root:

```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "db/queries/"
    schema: "db/migration/"
    gen:
      go:
        package: "sqlc"
        out: "internal/shared/db/sqlc"
        emit_json_tags: true
        emit_pointers_for_null_types: true
        overrides:
          - db_type: "uuid"
            go_type: "github.com/google/uuid.UUID"
          - db_type: "numeric"
            go_type: "github.com/shopspring/decimal.Decimal"
          - db_type: "text"
            go_type: "string"
```

Key overrides:
- `uuid` → `uuid.UUID` (not string)
- `numeric` → `decimal.Decimal` (not float64)

## Query File Layout

```
db/queries/
  contracts.sql       ← Contract CRUD
  contract_lines.sql  ← Contract line CRUD
  finance_accounts.sql
  users.sql
  ...
```

One file per domain entity (not one file per table).

## Query Annotations

SQLC requires annotations to name queries and specify return type:

```sql
-- name: GetContractByID :one
SELECT id, tenant_id, contract_number, title, status, total_value, currency,
       start_date, end_date, vendor_id, version, created_at, updated_at, deleted_at
FROM contracts
WHERE id = @id
  AND tenant_id = current_setting('app.tenant_id')::uuid
  AND deleted_at IS NULL;

-- name: ListContracts :many
SELECT id, tenant_id, contract_number, title, status, total_value, currency,
       start_date, end_date, vendor_id, version, created_at, updated_at, deleted_at
FROM contracts
WHERE tenant_id = current_setting('app.tenant_id')::uuid
  AND deleted_at IS NULL
  AND (@status::text IS NULL OR status = @status)
ORDER BY created_at DESC
LIMIT @limit_ OFFSET @offset_;

-- name: CreateContract :one
INSERT INTO contracts (
    tenant_id, contract_number, title, status, contract_type,
    total_value, currency, start_date, end_date, vendor_id, description
) VALUES (
    current_setting('app.tenant_id')::uuid,
    @contract_number, @title, 'draft', @contract_type,
    @total_value, @currency, @start_date, @end_date, @vendor_id, @description
)
RETURNING *;
```

Annotation types:
- `:one` — returns a single row (errors if 0 or >1)
- `:many` — returns a slice
- `:exec` — no return value
- `:execresult` — returns `sql.Result` (for affected row count)

## Using `current_setting` for Tenant Filtering

Always use `current_setting('app.tenant_id')::uuid` in queries instead of a `@tenant_id` parameter. This ensures the GUC set by `WithTenant` is used, even if the caller forgets to pass a tenant ID parameter.

```sql
-- ✅ Tenant from GUC — enforced by WithTenant
WHERE tenant_id = current_setting('app.tenant_id')::uuid

-- ❌ Tenant from parameter — caller could pass wrong value
WHERE tenant_id = @tenant_id
```

## Optimistic Lock Update

SQLC's generated `UPDATE ... RETURNING *` with version check:

```sql
-- name: UpdateContract :one
UPDATE contracts
SET title       = COALESCE(@title, title),
    total_value = COALESCE(@total_value, total_value),
    updated_at  = now(),
    version     = version + 1
WHERE id        = @id
  AND tenant_id = current_setting('app.tenant_id')::uuid
  AND version   = @version
  AND deleted_at IS NULL
RETURNING *;
```

If no rows returned (`pgx.ErrNoRows`), the repository maps to `ErrVersionConflict` or `ErrContractNotFound`.

## Soft Delete

```sql
-- name: SoftDeleteContract :one
UPDATE contracts
SET deleted_at = now(),
    updated_at = now(),
    version    = version + 1
WHERE id        = @id
  AND tenant_id = current_setting('app.tenant_id')::uuid
  AND version   = @version
  AND deleted_at IS NULL
RETURNING *;
```

## Regenerating After Query Changes

After any change to `db/queries/`:

```bash
make sqlc
```

This regenerates `internal/shared/db/sqlc/`. Commit both the query file and the generated files together.

If SQLC reports an error:
1. Check column names against the latest migration
2. Ensure the migration has been applied: `make migrate-up`
3. Verify SQLC config overrides for custom types

## Generated Code — Never Edit

Files in `internal/shared/db/sqlc/` are generated. Never edit them directly. Changes go in `db/queries/*.sql`, then regenerate.

The generated `Queries` struct is the only DB-facing API the repository layer uses:

```go
// Repository uses generated queries
type sqlcContractRepo struct {
    q *sqlc.Queries   // generated
}

func (r *sqlcContractRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.Contract, error) {
    row, err := r.q.GetContractByID(ctx, id)  // generated function
    if errors.Is(err, pgx.ErrNoRows) {
        return nil, domain.ErrContractNotFound
    }
    if err != nil {
        return nil, err
    }
    return toDomain(row), nil
}
```
