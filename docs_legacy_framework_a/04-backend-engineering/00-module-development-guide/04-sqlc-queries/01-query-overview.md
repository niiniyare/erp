> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: SQLC Query Overview
portal: 4 — Backend Engineering
section: 00-module-development-guide/04-sqlc-queries
audience: [backend-engineer, tech-lead]
related:
  - path: ./02-read-queries.md
    title: Read Queries
  - path: ./03-list-queries.md
    title: List Queries
  - path: ./04-update-queries.md
    title: Update Queries
---

# SQLC Query Overview

SQLC reads annotated SQL files and generates type-safe Go functions. All queries for a module live in `db/queries/<module>.sql`. No SQL anywhere else.

## File Location

```
db/queries/
└── contracts.sql    # all SQLC-annotated queries for the contracts module
```

## Annotation Format

```sql
-- name: <QueryName> :<return_type>
SELECT ...
```

| Return type | What it means |
|------------|---------------|
| `:one` | Returns one row. Zero rows → `pgx.ErrNoRows`. More than one → error. |
| `:many` | Returns a slice. Zero rows → empty slice (no error). |
| `:exec` | No rows returned (INSERT/UPDATE/DELETE without RETURNING). |
| `:execresult` | Returns `pgconn.CommandTag` with affected row count. |

AwoERP nearly always uses `RETURNING *` with `:one`, so the repository immediately gets the updated row without a second SELECT.

## Query Naming Convention

| Pattern | Example |
|---------|---------|
| `Create<Noun>` | `CreateContract` |
| `Get<Noun>ByID` | `GetContractByID` |
| `List<Nouns>` | `ListContracts` |
| `Count<Nouns>` | `CountContracts` |
| `Update<Noun>` | `UpdateContract` |
| `UpdateStatus<Noun>` or `Update<Noun>Status` | `UpdateContractStatus` |
| `SoftDelete<Noun>` | `SoftDeleteContract` |
| `<Noun>ExistsWith<Field>` | `ContractExistsWithNumber` |
| `List<Child>By<Parent>` | `ListContractLinesByContract` |
| `SoftDelete<Child>By<Parent>ID` | `SoftDeleteContractLinesByContractID` |
| `Update<Parent>TotalValue` | `UpdateContractTotalValue` |

## Parameter Placeholder Syntax

SQLC supports PostgreSQL `$1, $2` positional parameters or named `@param_name` syntax. Use named syntax (`@param_name`) for clarity in complex queries:

```sql
-- PREFERRED: named parameters
WHERE id = @id AND tenant_id = @tenant_id AND version = @version

-- ACCEPTABLE: positional parameters
WHERE id = $1 AND tenant_id = $2 AND version = $3
```

Named parameters make diffs clearer when parameters are reordered or added.

## Required Queries for Every Module

Every module must implement this minimum set:

| Query | Return | Purpose |
|-------|--------|---------|
| `Create<Noun>` | `:one` | Insert new record, return persisted row |
| `Get<Noun>ByID` | `:one` | Fetch by primary key within tenant |
| `List<Nouns>` | `:many` | Paginated list with filters |
| `Count<Nouns>` | `:one` | Pagination total count |
| `Update<Noun>` | `:one` | Full field update with optimistic lock |
| `Update<Noun>Status` | `:one` | Status-only update with optimistic lock |
| `SoftDelete<Noun>` | `:one` | Soft delete with optimistic lock |
| `<Noun>ExistsWithNumber` | `:one` | Unique key existence check |

## SQLC Configuration

SQLC is configured in `sqlc.yaml` at the repo root. After adding or changing queries, run:

```bash
make sqlc
```

Generated code lands in `db/sqlc/`. Never edit generated files manually — they are overwritten by `make sqlc`.

## What Not to Put in Query Files

| Prohibited | Reason |
|-----------|--------|
| Business logic | Queries are persistence; logic belongs in service layer |
| Permission checks | Queries execute after auth; no policy checks in SQL |
| `SET app.tenant_id` | Set by `store.WithTenant()`, never in query files |
| Dynamic SQL built by string concatenation | Use SQLC's `sqlc.narg()` for optional filters |
| `SELECT *` in production list queries | Enumerate columns; prevents schema drift issues |

## File Header

```sql
-- =============================================================
-- contracts.sql
-- SQLC annotated queries for the contracts module.
-- Module group: 011
-- Run `make sqlc` after any change to regenerate Go code.
-- =============================================================
```
