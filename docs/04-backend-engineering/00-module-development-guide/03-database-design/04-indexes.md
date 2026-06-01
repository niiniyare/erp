---
title: Indexes
portal: 4 — Backend Engineering
section: 00-module-development-guide/03-database-design
audience: [backend-engineer, tech-lead]
related:
  - path: ./02-primary-table.md
    title: Primary Table Migration
  - path: ./05-child-tables.md
    title: Child Tables
---

# Indexes

Index design follows the query patterns. Write the SQLC queries first (§04), then create indexes that support them. Every index must have a documented reason.

## Required Indexes

Every module must create these indexes at minimum:

### 1. Tenant + Status (list query)

```sql
CREATE INDEX idx_contracts_tenant_status
    ON contracts (tenant_id, status)
    WHERE deleted_at IS NULL;
```

The most common list query filters by tenant and status. Composite index `(tenant_id, status)` satisfies this with an index-only scan for status-filtered counts.

### 2. Tenant + Entity (org hierarchy filter)

```sql
CREATE INDEX idx_contracts_tenant_entity
    ON contracts (tenant_id, entity_id)
    WHERE deleted_at IS NULL;
```

Users scoped to a branch/department need contracts within their entity. This index supports `WHERE tenant_id = $1 AND entity_id = $2`.

### 3. Unique Business Key

```sql
-- Already created as a UNIQUE constraint — adds a unique index automatically
CONSTRAINT uq_contracts_tenant_number UNIQUE (tenant_id, contract_number)
```

The `UNIQUE` constraint creates a B-tree index automatically. No separate `CREATE INDEX` needed.

## Optional Indexes (add only when needed)

### Vendor lookups

```sql
CREATE INDEX idx_contracts_tenant_vendor
    ON contracts (tenant_id, vendor_id)
    WHERE deleted_at IS NULL;
```

Add if the UI has a "contracts by vendor" view or the vendor module needs to count active contracts.

### Date range queries

```sql
CREATE INDEX idx_contracts_tenant_dates
    ON contracts (tenant_id, start_date, end_date)
    WHERE deleted_at IS NULL;
```

Add if the application runs expiry reports or activation schedules. Without this, the query scans all tenant rows to find contracts expiring in a date range.

### Foreign key to parent (child tables)

```sql
-- On contract_lines — FK to contracts
CREATE INDEX idx_contract_lines_contract
    ON contract_lines (contract_id)
    WHERE deleted_at IS NULL;
```

PostgreSQL does not automatically index foreign keys. Every FK column on a child table needs an explicit index.

## Partial Indexes

All indexes use `WHERE deleted_at IS NULL` where possible. Benefits:

- **Smaller index**: soft-deleted rows (rare) are excluded.
- **Faster index maintenance**: INSERT/UPDATE of deleted rows does not touch the index.
- **Correct semantics**: the query planner uses the partial index for `WHERE deleted_at IS NULL` queries, which is every non-archive query.

The WHERE clause must exactly match the query filter for the planner to use it. If your query is `WHERE deleted_at IS NULL AND status = 'active'`, the index `WHERE deleted_at IS NULL` is still used — the planner applies the remaining filter on top.

## Index Naming Convention

```
idx_<table>_<column1>_<column2>
```

Examples:
- `idx_contracts_tenant_status`
- `idx_contracts_tenant_entity`
- `idx_contract_lines_contract`
- `idx_contracts_tenant_dates`

Constraint-based indexes follow the constraint name:
- `uq_contracts_tenant_number` (unique)
- `contracts_pkey` (primary key — auto-named)

## What NOT to Index

**Every column:** Indexes cost write performance. Only index columns that appear in `WHERE`, `ORDER BY`, or `JOIN ON` clauses in measured queries.

**Low-cardinality columns alone:** An index on `status` alone is useless — the table has 7 status values. As a leading column it adds nothing over a full table scan. Always pair with `tenant_id` as the leading column.

**Columns never used in filters:** `description`, `title` are free-text searched with `ILIKE` — these need `pg_trgm` GIN indexes, not B-tree. Do not create B-tree indexes on full-text columns.

## CONCURRENTLY for Production

Adding an index on a table with live data locks the table. Always use `CONCURRENTLY` in production migrations:

```sql
-- Safe for live production table
CREATE INDEX CONCURRENTLY idx_contracts_tenant_vendor
    ON contracts (tenant_id, vendor_id)
    WHERE deleted_at IS NULL;
```

`CONCURRENTLY` cannot run inside a transaction block. If your migration tool wraps migrations in transactions, run the `CREATE INDEX CONCURRENTLY` in a separate migration file outside the transaction, or use the tool's `-- disable-tx` annotation.

## Verify Index Usage

After deploying, verify the planner uses your indexes:

```sql
EXPLAIN (ANALYZE, BUFFERS)
SELECT * FROM contracts
WHERE tenant_id = 'aaaaaaaa-...'
  AND status = 'active'
  AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT 20;
```

Look for `Index Scan using idx_contracts_tenant_status` in the output. If you see `Seq Scan`, the index is not being used — check the partial index predicate and the query filter match.
