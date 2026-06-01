---
title: Schema Overview
portal: 4 — Backend Engineering
section: 00-module-development-guide/03-database-design
audience: [backend-engineer, tech-lead]
related:
  - path: ./02-primary-table.md
    title: Primary Table
  - path: ./03-rls-policies.md
    title: RLS Policies
  - path: ./04-indexes.md
    title: Indexes
---

# Schema Overview

The database schema for a module is split across three migration files. This document explains the organisation and the invariants every migration must satisfy.

## Migration File Structure

```
db/migration/
├── NNN_001_create_<module>.sql          # primary table + RLS + indexes
├── NNN_002_create_<module>_lines.sql    # child tables (skip if none)
└── NNN_003_create_<module>_views.sql    # views + materialised views (skip if none)
```

`NNN` is the three-digit module group number assigned at module creation. All files for one module share the same prefix so they sort and migrate together.

For contracts: `011`

```
db/migration/
├── 011001_create_contracts.sql
├── 011002_create_contract_lines.sql
└── 011003_create_contracts_views.sql
```

## Column Invariants

Every table must satisfy these invariants. They are checked in code review.

| Invariant | Rule |
|-----------|------|
| Primary key | `uuid DEFAULT gen_random_uuid()` — never `SERIAL` |
| Tenant isolation | `tenant_id uuid NOT NULL REFERENCES tenants(id)` |
| Org hierarchy | `entity_id uuid NOT NULL REFERENCES entities(id)` |
| Soft delete | `deleted_at timestamptz` — nullable; `NOT NULL` prohibited |
| Optimistic lock | `version integer NOT NULL DEFAULT 1` |
| Actor tracking | `created_by uuid NOT NULL`, `updated_by uuid NOT NULL` |
| Timestamps | `created_at timestamptz NOT NULL DEFAULT now()`, `updated_at timestamptz NOT NULL DEFAULT now()` |
| Monetary values | `numeric(20,6)` — never `float`, `real`, `double precision` |
| Status column | `VARCHAR(50) NOT NULL CHECK (status IN (...))` |
| RLS enabled | `ALTER TABLE ... ENABLE ROW LEVEL SECURITY` |
| RLS policy | `USING (tenant_id = current_setting('app.tenant_id')::uuid)` |

## Column Order Convention

```sql
CREATE TABLE contracts (
    -- 1. Identity
    id          uuid        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id   uuid        NOT NULL REFERENCES tenants(id),
    entity_id   uuid        NOT NULL REFERENCES entities(id),

    -- 2. Module-specific fields
    contract_number VARCHAR(50) NOT NULL,
    title           TEXT        NOT NULL,
    -- ...

    -- 3. Lifecycle
    status      VARCHAR(50) NOT NULL CHECK (status IN (...)),
    version     integer     NOT NULL DEFAULT 1,

    -- 4. Actor tracking
    created_by  uuid        NOT NULL,
    updated_by  uuid        NOT NULL,

    -- 5. Timestamps (always last)
    deleted_at  timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);
```

## What Goes in Each File

### `NNN_001_create_<module>.sql`

- Primary table `CREATE TABLE`
- Unique constraints on the primary table
- RLS enable + policy
- Indexes on the primary table
- Any triggers (e.g., `updated_at` auto-update trigger)

### `NNN_002_create_<module>_lines.sql`

- Child table `CREATE TABLE` (with FK to primary table)
- RLS enable + policy on child table
- Indexes on child table
- Only created if the module has child records

### `NNN_003_create_<module>_views.sql`

- Read-only views (`CREATE VIEW`)
- Materialised views (`CREATE MATERIALIZED VIEW`) with `REFRESH MATERIALIZED VIEW CONCURRENTLY` support
- Only created if consumers benefit from pre-joined views

## Rollback Requirement

Every migration file must be fully reversible. The `make migrate-down` target must succeed without data loss on a clean test database.

```sql
-- At the end of each migration file, include a DOWN comment block
-- that the migration tool uses for rollback:

-- +migrate Down
DROP VIEW IF EXISTS v_contract_summary;
DROP TABLE IF EXISTS contract_lines;
DROP TABLE IF EXISTS contracts;
```

The exact rollback syntax depends on your migration tool (`golang-migrate`, `goose`, etc.). Check `db/migration/` for the format used in existing files.

## Naming Conventions

| Object | Convention | Example |
|--------|-----------|---------|
| Table | `snake_case`, plural | `contracts`, `contract_lines` |
| Column | `snake_case` | `contract_number`, `total_value` |
| Primary key | `id` | `id` |
| Foreign key | `<referenced_table_singular>_id` | `contract_id`, `tenant_id` |
| Index | `idx_<table>_<columns>` | `idx_contracts_tenant_id_status` |
| Unique constraint | `uq_<table>_<columns>` | `uq_contracts_tenant_id_number` |
| Check constraint | `chk_<table>_<column>` | `chk_contracts_status` |
| RLS policy | `<table>_tenant_isolation` | `contracts_tenant_isolation` |
| View | `v_<description>` | `v_contract_summary` |
| Materialised view | `mv_<description>` | `mv_contract_monthly_totals` |
