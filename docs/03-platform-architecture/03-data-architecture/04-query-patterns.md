---
title: Query Patterns
portal: 3 — Platform Architecture
section: 03-platform-architecture
audience: [backend-engineer, architect]
related:
  - "[Data Overview](01-data-overview.md)"
  - "[Schema Conventions](02-schema-conventions.md)"
  - "[SQLC Layer](../../04-backend-engineering/00-module-development-guide/04-sqlc-layer/01-sqlc-overview.md)"
---

# Query Patterns

## Tenant-Scoped Queries

All queries are implicitly tenant-scoped via RLS. The `current_setting('app.tenant_id')` GUC is set by `WithTenant()` before any query runs.

```sql
-- In SQLC query files, rely on GUC for tenant isolation
SELECT * FROM contracts
WHERE id = @id
  AND deleted_at IS NULL;
-- tenant_id filter NOT needed — RLS policy handles it
```

However, add explicit `tenant_id` filter as belt-and-suspenders for large tables:

```sql
SELECT * FROM contracts
WHERE id = @id
  AND tenant_id = current_setting('app.tenant_id')::uuid   -- belt
  AND deleted_at IS NULL;
```

## Soft Delete Filter

Always filter `deleted_at IS NULL` in queries:

```sql
WHERE deleted_at IS NULL
```

Partial indexes include this condition for efficiency:

```sql
CREATE INDEX contracts_tenant_status ON contracts (tenant_id, status)
    WHERE deleted_at IS NULL;
```

## Pagination Pattern

Standard offset-based pagination:

```sql
-- name: ListContracts :many
SELECT *, COUNT(*) OVER() AS total_count
FROM contracts
WHERE tenant_id = current_setting('app.tenant_id')::uuid
  AND deleted_at IS NULL
  AND (@status::text IS NULL OR status = @status)
ORDER BY created_at DESC
LIMIT @limit_ OFFSET @offset_;
```

`COUNT(*) OVER()` is a window function that returns the total count in every row — avoids a separate COUNT query.

In SQLC, `limit_` is used instead of `limit` to avoid conflict with the SQL `LIMIT` keyword.

```go
// Repository builds params
params := sqlc.ListContractsParams{
    Status:  pgtype.Text{String: string(*filter.Status), Valid: filter.Status != nil},
    Limit_: int32(pageSize),
    Offset: int32((page - 1) * pageSize),
}
rows, err := q.ListContracts(ctx, params)
// rows[0].TotalCount gives the full count
```

## Hierarchy Queries (Closure Table)

Entity hierarchy uses `hierarchy_paths` closure table for efficient subtree queries:

```sql
-- Get all entities in a subtree
SELECT e.*
FROM entities e
JOIN hierarchy_paths hp ON hp.descendant_id = e.id
WHERE hp.ancestor_id = @root_entity_id
  AND e.tenant_id = current_setting('app.tenant_id')::uuid
  AND e.is_active = true;

-- Get path from root to a specific entity
SELECT e.*
FROM entities e
JOIN hierarchy_paths hp ON hp.ancestor_id = e.id
WHERE hp.descendant_id = @entity_id
ORDER BY hp.depth;
```

## Upsert Pattern

For idempotent operations (e.g., event deduplication, feature flag sync):

```sql
INSERT INTO tenant_feature_overrides (tenant_id, feature_name, enabled)
VALUES (@tenant_id, @feature_name, @enabled)
ON CONFLICT (tenant_id, feature_name)
DO UPDATE SET enabled = EXCLUDED.enabled, updated_at = now();
```

## Optimistic Lock Update

```sql
UPDATE contracts
SET title       = @title,
    updated_at  = now(),
    version     = version + 1
WHERE id        = @id
  AND version   = @version
  AND deleted_at IS NULL
RETURNING *;
```

Map `pgx.ErrNoRows` to `ErrVersionConflict` in the repository. The caller cannot distinguish "not found" from "version mismatch" — treat both as conflict (the client must re-fetch).

## Aggregation

Financial aggregations use `SUM` with `COALESCE` to handle empty result sets:

```sql
SELECT
    COALESCE(SUM(debit_amount), 0) AS total_debits,
    COALESCE(SUM(credit_amount), 0) AS total_credits
FROM finance_transaction_entries fte
JOIN finance_transactions ft ON ft.id = fte.transaction_id
WHERE fte.account_id = @account_id
  AND ft.status = 'posted'
  AND ft.tenant_id = current_setting('app.tenant_id')::uuid;
```

## JSON Aggregation

For nested result sets (contract with lines):

```sql
SELECT
    c.id, c.contract_number, c.title, c.status,
    COALESCE(
        json_agg(
            json_build_object(
                'id', cl.id,
                'description', cl.description,
                'quantity', cl.quantity,
                'unit_price', cl.unit_price
            ) ORDER BY cl.line_order
        ) FILTER (WHERE cl.id IS NOT NULL),
        '[]'
    ) AS lines
FROM contracts c
LEFT JOIN contract_lines cl ON cl.contract_id = c.id AND cl.deleted_at IS NULL
WHERE c.id = @id
  AND c.deleted_at IS NULL
GROUP BY c.id;
```

`FILTER (WHERE cl.id IS NOT NULL)` prevents `[null]` when there are no lines.

## Index Strategy

| Query pattern | Index type |
|--------------|-----------|
| `WHERE tenant_id = X AND status = Y` | Composite: `(tenant_id, status)` |
| `ORDER BY created_at DESC` | Composite: `(tenant_id, created_at DESC)` |
| Full-text search on title | `GIN` on `to_tsvector(title)` |
| Hierarchy traversal | `(ancestor_id, depth)` on `hierarchy_paths` |
| UUID lookup by PK | Default btree on `id` |

Always partial-index with `WHERE deleted_at IS NULL` for business tables to keep index size small.
