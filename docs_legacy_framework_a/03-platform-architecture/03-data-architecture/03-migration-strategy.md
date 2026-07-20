> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Migration Strategy
portal: 3 — Platform Architecture
section: 03-data-architecture
audience: [architect, backend-engineer, tech-lead]
related:
  - "[Schema Conventions](02-schema-conventions.md)"
  - "[Database Layer](../../04-backend-engineering/00-module-development-guide/03-database-layer/01-database-overview.md)"
  - "[Deployment Checklist](../../04-backend-engineering/00-module-development-guide/22-deployment-checklist/01-deployment-checklist.md)"
---

# Migration Strategy

## Tool

Migrations use `golang-migrate`. Files live in `db/migration/` and run sequentially at server startup.

## File Naming

```
{group}{sequence}_{description}.{up|down}.sql
```

| Part | Description | Example |
|------|-------------|---------|
| group | 3-digit module group | `011` |
| sequence | 3-digit sequence within group | `001` |
| description | lowercase snake_case | `create_contracts` |
| direction | `up` or `down` | `up` |

Full example: `011001_create_contracts.up.sql`

## Migration Rules

**Append-only**: never modify a migration file after it has been applied to any environment (dev, staging, prod). Create a new migration to fix it.

**Reversible**: every `up` migration must have a matching `down` migration that fully reverses the change. `down` files are required even if they only run in development.

**Idempotent-safe**: migrations run exactly once via golang-migrate's schema_migrations tracking table. Do not use `CREATE TABLE IF NOT EXISTS` — if a migration is being retried it indicates a bug.

**No data migrations in schema files**: if a migration needs to backfill data, create a separate migration file for the data change after the schema change.

## Zero-Downtime Migrations

For production deployments, follow the expand-contract pattern for breaking changes:

### Adding a column

```sql
-- Safe: adding nullable column with no default (no table rewrite)
ALTER TABLE contracts ADD COLUMN notes text;
```

### Adding a NOT NULL column

```sql
-- Step 1: add nullable
ALTER TABLE contracts ADD COLUMN notes text;

-- Step 2 (separate migration): backfill default
UPDATE contracts SET notes = '' WHERE notes IS NULL;

-- Step 3 (separate migration, after deploy): add NOT NULL constraint
ALTER TABLE contracts ALTER COLUMN notes SET NOT NULL;
ALTER TABLE contracts ALTER COLUMN notes SET DEFAULT '';
```

### Renaming a column

```sql
-- Step 1: add new column
ALTER TABLE contracts ADD COLUMN contract_title text;

-- Step 2: backfill
UPDATE contracts SET contract_title = title;

-- Step 3: add NOT NULL after backfill
ALTER TABLE contracts ALTER COLUMN contract_title SET NOT NULL;

-- Step 4 (after old column code is removed): drop old column
ALTER TABLE contracts DROP COLUMN title;
```

Never rename in one step — it breaks in-flight queries.

## Index Creation in Production

```sql
-- Use CONCURRENTLY to avoid table lock
CREATE INDEX CONCURRENTLY idx_contracts_vendor
    ON contracts (tenant_id, vendor_id)
    WHERE deleted_at IS NULL;
```

`CONCURRENTLY` cannot run inside a transaction. golang-migrate runs each migration in a transaction by default. Use the `-- migrate: notransaction` directive:

```sql
-- migrate:notransaction
CREATE INDEX CONCURRENTLY idx_contracts_vendor
    ON contracts (tenant_id, vendor_id)
    WHERE deleted_at IS NULL;
```

## Rollback Strategy

For production incidents, use `make migrate-down VERSION=N` to roll back to a specific version. Ensure the `down` migration fully reverses the change including data.
