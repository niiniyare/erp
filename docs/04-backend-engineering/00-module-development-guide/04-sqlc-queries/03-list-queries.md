---
title: List Queries
portal: 4 — Backend Engineering
section: 00-module-development-guide/04-sqlc-queries
audience: [backend-engineer, tech-lead]
related:
  - path: ./02-read-queries.md
    title: Read Queries
  - path: ./04-update-queries.md
    title: Update Queries
---

# List Queries

List queries return paginated, filtered, sorted results. They always come in pairs: a `List<Nouns>` query that returns rows, and a `Count<Nouns>` query that returns the total count for pagination.

## Pattern: sqlc.narg() for Optional Filters

SQLC provides `sqlc.narg(param)` to declare a parameter that is nullable. When the value is NULL, the condition is skipped:

```sql
-- name: ListContracts :many
SELECT
    id, tenant_id, entity_id, contract_number, title, description,
    vendor_id, contract_type, start_date, end_date, total_value, currency,
    status, version, created_by, updated_by, deleted_at, created_at, updated_at
FROM contracts
WHERE tenant_id   = @tenant_id
  AND deleted_at  IS NULL
  AND (sqlc.narg(entity_id)::uuid     IS NULL OR entity_id     = sqlc.narg(entity_id))
  AND (sqlc.narg(vendor_id)::uuid     IS NULL OR vendor_id     = sqlc.narg(vendor_id))
  AND (sqlc.narg(status)::VARCHAR     IS NULL OR status        = sqlc.narg(status))
  AND (sqlc.narg(contract_type)::VARCHAR IS NULL OR contract_type = sqlc.narg(contract_type))
  AND (
        sqlc.narg(search)::TEXT IS NULL
        OR title            ILIKE '%' || sqlc.narg(search) || '%'
        OR contract_number  ILIKE '%' || sqlc.narg(search) || '%'
      )
ORDER BY
    CASE WHEN @sort_by = 'created_at' AND @sort_dir = 'desc' THEN created_at END DESC,
    CASE WHEN @sort_by = 'created_at' AND @sort_dir = 'asc'  THEN created_at END ASC,
    CASE WHEN @sort_by = 'total_value' AND @sort_dir = 'desc' THEN total_value END DESC,
    CASE WHEN @sort_by = 'total_value' AND @sort_dir = 'asc'  THEN total_value END ASC,
    created_at DESC  -- default: most recent first
LIMIT  @page_size
OFFSET @page_offset;
```

## Count Query (always paired with List)

```sql
-- name: CountContracts :one
SELECT COUNT(*) AS total
FROM contracts
WHERE tenant_id   = @tenant_id
  AND deleted_at  IS NULL
  AND (sqlc.narg(entity_id)::uuid     IS NULL OR entity_id     = sqlc.narg(entity_id))
  AND (sqlc.narg(vendor_id)::uuid     IS NULL OR vendor_id     = sqlc.narg(vendor_id))
  AND (sqlc.narg(status)::VARCHAR     IS NULL OR status        = sqlc.narg(status))
  AND (sqlc.narg(contract_type)::VARCHAR IS NULL OR contract_type = sqlc.narg(contract_type))
  AND (
        sqlc.narg(search)::TEXT IS NULL
        OR title            ILIKE '%' || sqlc.narg(search) || '%'
        OR contract_number  ILIKE '%' || sqlc.narg(search) || '%'
      );
```

The `WHERE` clause is **identical** to the `ListContracts` query (minus pagination). This is intentional — they must filter the same rows. If filters drift between the two queries, pagination is wrong.

## Params Struct in Go

SQLC generates a params struct from the query annotations. The repository's `List` method accepts a domain-level params struct, not the SQLC-generated one:

```go
// domain/contract.go (or repository/contract.go)
type ListContractsParams struct {
    TenantID     uuid.UUID
    EntityID     *uuid.UUID   // optional
    VendorID     *uuid.UUID   // optional
    Status       *string      // optional
    ContractType *string      // optional
    Search       *string      // optional
    SortBy       string       // "created_at" | "total_value"
    SortDir      string       // "asc" | "desc"
    PageSize     int
    PageOffset   int
}
```

The repository maps this to the SQLC-generated params struct. Using pointer types (`*uuid.UUID`, `*string`) for optional filters maps cleanly to `sqlc.narg()` nullable parameters.

## Pagination in the Response

The service and handler compute pagination metadata from the count and page params:

```go
type PaginatedContracts struct {
    Items      []*domain.Contract
    TotalCount int64
    PageSize   int
    PageOffset int
    TotalPages int
}
```

```go
func paginate(total int64, pageSize, pageOffset int) PaginatedContracts {
    totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
    return PaginatedContracts{
        TotalCount: total,
        PageSize:   pageSize,
        PageOffset: pageOffset,
        TotalPages: totalPages,
    }
}
```

## Cursor-Based Pagination (alternative)

For high-volume tables where offset pagination is slow (OFFSET 10000 scans 10000 rows), use cursor-based pagination:

```sql
-- name: ListContractsCursor :many
SELECT ...
FROM contracts
WHERE tenant_id  = @tenant_id
  AND deleted_at IS NULL
  AND (sqlc.narg(after_id)::uuid IS NULL OR id < sqlc.narg(after_id))
ORDER BY id DESC
LIMIT @page_size;
```

The client passes the last `id` seen as `after_id` to get the next page. This is O(1) regardless of how deep into the list the client is.

Use cursor pagination for lists expected to exceed 10,000 rows. Use offset pagination for smaller lists where "go to page 5" is a UI requirement.

## ILIKE Full-Text Search

The `search` filter uses `ILIKE` for simple prefix/substring search:

```sql
OR title ILIKE '%' || sqlc.narg(search) || '%'
```

This performs a sequential scan and is fine for small tables (< 50,000 rows per tenant). For larger datasets, use `pg_trgm` GIN indexes:

```sql
CREATE INDEX idx_contracts_title_trgm ON contracts USING GIN (title gin_trgm_ops) WHERE deleted_at IS NULL;
CREATE INDEX idx_contracts_number_trgm ON contracts USING GIN (contract_number gin_trgm_ops) WHERE deleted_at IS NULL;
```

With GIN indexes, `ILIKE` queries use the trigram index. Add this to the migration when search performance becomes a concern.
