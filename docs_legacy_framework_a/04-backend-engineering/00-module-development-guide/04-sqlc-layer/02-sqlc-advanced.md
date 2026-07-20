> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: SQLC Advanced Patterns
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer]
related:
  - "[SQLC Overview](01-sqlc-overview.md)"
  - "[Repository Layer](../05-repository-layer/01-repository-overview.md)"
  - "[Query Patterns](../../../03-platform-architecture/03-data-architecture/04-query-patterns.md)"
---

# SQLC Advanced Patterns

## Optional Filters with COALESCE

Use `COALESCE(@param, column)` for optional filter params:

```sql
-- name: ListContracts :many
SELECT c.*, COUNT(*) OVER() AS total_count
FROM contracts c
WHERE c.tenant_id = current_setting('app.tenant_id')::uuid
  AND c.deleted_at IS NULL
  AND (@status::text IS NULL       OR c.status = @status)
  AND (@entity_id::uuid IS NULL    OR c.entity_id = @entity_id)
  AND (@date_from::date IS NULL    OR c.created_at::date >= @date_from)
  AND (@date_to::date IS NULL      OR c.created_at::date <= @date_to)
ORDER BY c.created_at DESC
LIMIT @limit_ OFFSET @offset_;
```

Pass `pgtype.Text{Valid: false}` for SQL NULL — SQLC handles the parameter binding.

## Batch Insert

For inserting multiple rows atomically:

```sql
-- name: BulkCreateContractLines :many
INSERT INTO contract_lines (contract_id, description, quantity, unit_price, line_order)
SELECT
    @contract_id,
    unnest(@descriptions::text[]),
    unnest(@quantities::numeric[]),
    unnest(@unit_prices::numeric[]),
    generate_series(1, array_length(@descriptions::text[], 1))
RETURNING *;
```

In Go:

```go
q.BulkCreateContractLines(ctx, sqlc.BulkCreateContractLinesParams{
    ContractID:   contractID,
    Descriptions: descriptions,    // []string
    Quantities:   quantities,      // []decimal.Decimal
    UnitPrices:   unitPrices,
})
```

## Count Query

For pagination without returning rows:

```sql
-- name: CountContracts :one
SELECT COUNT(*) AS count
FROM contracts
WHERE tenant_id = current_setting('app.tenant_id')::uuid
  AND deleted_at IS NULL
  AND (@status::text IS NULL OR status = @status);
```

## Soft Delete Scope

SQLC doesn't support named scopes. Add `deleted_at IS NULL` to every query, or use a view:

```sql
-- Use the view in queries instead of the base table
-- name: ListActiveContracts :many
SELECT * FROM v_active_contracts   -- view has WHERE deleted_at IS NULL
WHERE tenant_id = current_setting('app.tenant_id')::uuid
ORDER BY created_at DESC
LIMIT @limit_ OFFSET @offset_;
```

## Returning vs Querying Again

Always use `RETURNING *` on mutations to avoid a second query:

```sql
-- name: ActivateContract :one
UPDATE contracts
SET status     = 'active',
    updated_at = now(),
    version    = version + 1
WHERE id = @id
  AND version  = @version
  AND status   = 'approved'   -- guard: only activate from approved
  AND deleted_at IS NULL
RETURNING *;
```

No need for a separate `SELECT` after the update.

## JSON Aggregation for Nested Objects

```sql
-- name: GetContractWithLines :one
SELECT
    c.id, c.contract_number, c.title, c.status, c.total_value, c.version,
    COALESCE(
        json_agg(
            json_build_object(
                'id',          cl.id,
                'description', cl.description,
                'quantity',    cl.quantity,
                'unit_price',  cl.unit_price,
                'amount',      cl.amount,
                'line_order',  cl.line_order
            ) ORDER BY cl.line_order
        ) FILTER (WHERE cl.id IS NOT NULL),
        '[]'::json
    ) AS lines
FROM contracts c
LEFT JOIN contract_lines cl ON cl.contract_id = c.id AND cl.deleted_at IS NULL
WHERE c.id = @id
  AND c.deleted_at IS NULL
GROUP BY c.id;
```

SQLC maps `lines` to `json.RawMessage` — deserialize in the repository:

```go
type ContractWithLinesRow struct {
    sqlc.Contract
    Lines json.RawMessage
}

// In mapper:
var lines []domain.ContractLine
json.Unmarshal(row.Lines, &lines)
```

## Exec for Fire-and-Forget

```sql
-- name: MarkEventDelivered :exec
UPDATE event_outbox
SET delivered_at = now()
WHERE id = @id;
```

`:exec` — no return value. Use when you don't need the updated row.

## ExecResult for Affected Count

```sql
-- name: DeleteExpiredSessions :execresult
DELETE FROM user_sessions
WHERE expires_at < now();
```

```go
result, err := q.DeleteExpiredSessions(ctx)
n, _ := result.RowsAffected()
slog.Info("expired sessions purged", "count", n)
```

## Named Parameters vs Positional

SQLC supports both `@param` (named) and `$1` (positional). Prefer `@param` for readability:

```sql
-- ✅ Named
WHERE id = @id AND version = @version

-- Also valid — generated identically
WHERE id = $1 AND version = $2
```

Named params map to struct fields in the generated `Params` struct by converting `@param_name` to `ParamName` (PascalCase).

## sqlc.yaml Override: Arrays

For passing arrays as params:

```yaml
overrides:
  - go_type: "[]github.com/google/uuid.UUID"
    db_type: "uuid[]"
```

Then use `= ANY(@ids::uuid[])` in queries.
