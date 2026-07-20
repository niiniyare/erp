> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Data Architecture Overview
portal: 3 — Platform Architecture
section: 03-data-architecture
audience: [architect, backend-engineer, tech-lead]
related:
  - "[Schema Conventions](02-schema-conventions.md)"
  - "[Migration Strategy](03-migration-strategy.md)"
  - "[RLS Enforcement](../01-multi-tenancy/02-rls-enforcement.md)"
---

# Data Architecture Overview

## Storage Tier

| Store | Purpose | Technology |
|-------|---------|-----------|
| Primary DB | All business data | PostgreSQL 15+ |
| Session store | User sessions | Redis (TTL) |
| Cache | Computed views, lookups | Redis |
| Object store | File uploads, exports | S3-compatible |
| Workflow state | Temporal workflow history | Temporal DB (PostgreSQL) |

## PostgreSQL Design Philosophy

All persistent business data lives in PostgreSQL. The schema is designed for:

- **Correctness over speed**: constraints, checks, foreign keys enforced at DB layer
- **Tenant isolation**: RLS on every table
- **Auditability**: `created_at`, `updated_at`, `deleted_at`, `created_by`, `updated_by` on every mutable table
- **Concurrency safety**: optimistic locking via `version` column
- **Exact arithmetic**: `numeric(20,6)` for monetary and quantity values

## Column Invariants (every table)

| Column | Type | Rule |
|--------|------|------|
| `id` | `uuid DEFAULT gen_random_uuid()` | Never SERIAL |
| `tenant_id` | `uuid NOT NULL REFERENCES tenants(id)` | Every table |
| `version` | `integer NOT NULL DEFAULT 1` | Every mutable table |
| `created_at` | `timestamptz NOT NULL DEFAULT now()` | Immutable after insert |
| `updated_at` | `timestamptz NOT NULL DEFAULT now()` | Managed by trigger |
| `deleted_at` | `timestamptz` (nullable) | NULL = not deleted |
| `created_by` | `uuid NOT NULL REFERENCES users(id)` | Who created |
| `updated_by` | `uuid NOT NULL REFERENCES users(id)` | Who last modified |

## Naming Conventions

| Item | Convention | Example |
|------|-----------|---------|
| Tables | `snake_case` plural | `contracts`, `contract_lines` |
| Columns | `snake_case` | `contract_number`, `total_value` |
| Indexes | `idx_{table}_{columns}` | `idx_contracts_tenant_status` |
| Constraints | `{table}_{description}_{type}` | `contracts_status_check` |
| Policies | `rls_{table}` | `rls_contracts` |
| Triggers | `{table}_{event}` | `contracts_updated_at` |
| Migration files | `{group}{seq}_{description}.{up\|down}.sql` | `011001_create_contracts.up.sql` |

## SQLC Code Generation

All database access is via SQLC-generated typed Go code. No reflection-based ORM, no string-concatenated SQL outside of `.sql` files.

```
db/queries/*.sql   →  make sqlc  →  db/sqlc/*.go (generated, never edited)
```

Changes to queries require:
1. Edit `.sql` file
2. Run `make sqlc`
3. Compile generated code

## Connection Pool

```
pgxpool settings:
  MaxConns:          25
  MinConns:          5
  MaxConnLifetime:   30 minutes
  MaxConnIdleTime:   5 minutes
  HealthCheckPeriod: 1 minute
```

Pool is shared across all modules. `WithTenant` acquires a connection, sets the GUC, runs the operation, and releases the connection back to the pool.

## Read vs. Write Patterns

| Pattern | Used for |
|---------|---------|
| `WithTenant` (transactional) | All writes, consistency-sensitive reads |
| Direct pool query with GUC | Read-heavy list operations (future optimization) |
| Redis cache | Computed aggregates, lookup tables, session data |
