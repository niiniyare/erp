---
title: "Chapter 11: Database Migrations"
part: "Part II — The EntityDefinition System"
chapter: 11
section: "11-migrations"
related:
  - "[Chapter 3: Architecture Overview](../part-01-foundations/03-architecture.md)"
  - "[Chapter 38: Tenant Lifecycle](../part-07-multitenancy/38-tenant-lifecycle.md)"
  - "[Chapter 52: CLI Reference](../part-08-deployment/52-cli.md)"
---

# Chapter 11: Database Migrations

Awo uses `golang-migrate` with the PostgreSQL driver to manage all database schema changes. Every DDL change — tables, columns, indexes, views, functions, triggers, RLS policies — is captured in versioned migration files. Nothing is applied to the database that does not pass through a numbered migration file first.

This chapter covers the migration toolchain, file format, RLS-specific patterns, complex DDL (views, functions, triggers, policies), testing strategy, and rollback procedures.

---

## 11.1 Migration Strategy Overview

### 11.1.1 Why Manual Reviewed Migrations, Not Auto-Migrate

Auto-migrate tools (including ent's built-in auto-migration) diff the current Go schema against the live database and apply the minimum set of changes. This is convenient in development but dangerous in production because:

- It can silently drop columns or indexes that it believes are unused.
- It runs destructive DDL without a human reviewing it.
- There is no down-migration — rollback is not possible.
- It provides no audit trail of what changed and when.

Awo's approach: generate migration files, review them before committing, apply them through a controlled CLI command that records which migrations have run.

### 11.1.2 golang-migrate

`golang-migrate` is the migration runner. It:
- Reads `.up.sql` and `.down.sql` files from a directory.
- Records which migrations have been applied in a `schema_migrations` table.
- Applies pending migrations in filename order.
- Rolls back the most recent migration when asked.

The PostgreSQL driver is used directly — no ORM abstraction sits between `golang-migrate` and the database.

### 11.1.3 RLS-Aware Migrations

All tenants share a single PostgreSQL database and schema. Isolation is enforced by Row-Level Security (RLS) keyed on `tenant_id`. This has an important implication for migrations: **applying a migration once applies it for all tenants**. There is no per-tenant schema to migrate, no fleet of schemas to walk through, no per-tenant migration concurrency concern.

A migration that adds a `credit_limit` column to `customers` adds it for every tenant's customers simultaneously. A migration that creates a new RLS policy applies to everyone. This is simpler operationally than a schema-per-tenant model, but it means you must be more careful with zero-downtime techniques because the shared table has all tenants' production data in it.

Per-tenant differences are handled via seed data (inserted by the provisioning workflow), not via schema differences.

---

## 11.2 Migration File Format

### 11.2.1 File Naming

Migration files come in pairs, named with a Unix timestamp in seconds (14 digits) followed by a description slug:

```
migrations/
  20240101090000_initial_schema.up.sql
  20240101090000_initial_schema.down.sql
  20240615120000_add_customer_credit_limit.up.sql
  20240615120000_add_customer_credit_limit.down.sql
  20240720083000_add_audit_log_partition_2024_08.up.sql
  20240720083000_add_audit_log_partition_2024_08.down.sql
```

The timestamp is the migration's version. `golang-migrate` applies them in ascending version order.

**Do not use sequential integers** (001, 002). Integers cause conflicts when two developers create migrations concurrently on different branches. Timestamps are effectively conflict-free because they encode the creation time to the second.

### 11.2.2 Up Migration

The `.up.sql` file contains the forward DDL. It must be idempotent where possible. Use `IF NOT EXISTS` for `CREATE TABLE` and `CREATE INDEX`. For `ALTER TABLE ADD COLUMN`, if the column already exists the migration will fail — write your migrations carefully and test the round-trip.

```sql
-- 20240615120000_add_customer_credit_limit.up.sql

ALTER TABLE customers
    ADD COLUMN credit_limit numeric(20,4) NOT NULL DEFAULT 0,
    ADD COLUMN credit_currency varchar(3) NOT NULL DEFAULT 'KES';

COMMENT ON COLUMN customers.credit_limit IS 'Maximum outstanding balance allowed. 0 = no limit enforced.';
```

### 11.2.3 Down Migration

The `.down.sql` file contains the rollback DDL. Down migrations are **required**, not optional. A migration without a corresponding down file will fail the CI migration round-trip test.

```sql
-- 20240615120000_add_customer_credit_limit.down.sql

ALTER TABLE customers
    DROP COLUMN IF EXISTS credit_limit,
    DROP COLUMN IF EXISTS credit_currency;
```

Down migrations must leave the database in the exact state it was in before the up migration ran. If your up migration creates a table, your down migration drops it. If your up migration adds a column, your down migration drops it.

### 11.2.4 Editing Applied Migrations

**Never edit a migration file that has already been applied to any environment**. Once a migration is applied, its content is the canonical record of what happened at that version. Editing it creates a divergence between the recorded hash and the file on disk, which `golang-migrate` detects as data corruption.

To correct a mistake in an applied migration, write a new migration that makes the corrective change.

---

## 11.3 CLI Commands

All migration operations go through the `awo migrate` subcommand, which wraps `golang-migrate` with Awo's configuration loading.

### 11.3.1 `awo migrate up`

Apply all pending migrations:

```bash
$ awo migrate up
2024/06/15 12:00:05 Applying migration 20240101090000 (initial_schema) ... OK
2024/06/15 12:00:06 Applying migration 20240615120000 (add_customer_credit_limit) ... OK
2024/06/15 12:00:06 All migrations applied. Current version: 20240615120000
```

Running `migrate up` when there are no pending migrations is a no-op.

### 11.3.2 `awo migrate down [N]`

Roll back N migrations. Default is 1:

```bash
$ awo migrate down
2024/06/15 12:05:00 Rolling back migration 20240615120000 (add_customer_credit_limit) ... OK
Current version: 20240101090000

$ awo migrate down 2
# Rolls back 2 migrations
```

### 11.3.3 `awo migrate version`

Print the currently applied version without making any changes:

```bash
$ awo migrate version
Version: 20240615120000 (dirty: false)
```

If a migration failed partway through, the dirty flag will be `true` and no further migrations can run until the situation is resolved.

### 11.3.4 `awo migrate create <name>`

Create a new timestamped migration file pair:

```bash
$ awo migrate create add_inventory_batches
Created: migrations/20240720083000_add_inventory_batches.up.sql
Created: migrations/20240720083000_add_inventory_batches.down.sql
```

The files are created with placeholder comments. You edit them to add the actual DDL.

### 11.3.5 `awo migrate force <V>`

Mark a specific version as applied without running any SQL. Use only in emergencies, when a migration partially succeeded and you have manually completed the remaining DDL:

```bash
$ awo migrate force 20240720083000
Version forced to 20240720083000 (dirty: false)
```

After `force`, `migrate up` will begin applying migrations from the version after the forced one.

---

## 11.4 RLS Migration Pattern

Every new tenant-scoped table requires three things: a `tenant_id` column, an RLS enable statement, and an RLS policy. These all go in the same `.up.sql` file as the `CREATE TABLE`.

```sql
-- 20240801100000_create_purchase_orders.up.sql

CREATE TABLE purchase_orders (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id      uuid NOT NULL REFERENCES tenants(id),
    series         varchar(50) NOT NULL,
    supplier_id    uuid NOT NULL REFERENCES suppliers(id),
    status         varchar(20) NOT NULL DEFAULT 'DRAFT',
    order_date     date NOT NULL,
    expected_date  date,
    grand_total    numeric(20,4) NOT NULL DEFAULT 0,
    currency       varchar(3) NOT NULL DEFAULT 'KES',
    notes          text,
    custom_fields  jsonb NOT NULL DEFAULT '{}',
    created_by     uuid REFERENCES users(id),
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now()
);

-- RLS setup
ALTER TABLE purchase_orders ENABLE ROW LEVEL SECURITY;

CREATE POLICY purchase_orders_tenant_isolation ON purchase_orders
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

-- Trigger for updated_at
CREATE TRIGGER purchase_orders_updated_at
    BEFORE UPDATE ON purchase_orders
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Audit log trigger
CREATE TRIGGER purchase_orders_audit
    AFTER INSERT OR UPDATE OR DELETE ON purchase_orders
    FOR EACH ROW EXECUTE FUNCTION audit_log_trigger_fn();

-- Indexes
CREATE INDEX idx_purchase_orders_tenant ON purchase_orders(tenant_id);
CREATE INDEX idx_purchase_orders_supplier ON purchase_orders(tenant_id, supplier_id);
CREATE INDEX idx_purchase_orders_status ON purchase_orders(tenant_id, status);
```

The down migration reverses in the correct order (triggers → policies → table):

```sql
-- 20240801100000_create_purchase_orders.down.sql

DROP TABLE IF EXISTS purchase_orders;
-- RLS policies and triggers are dropped automatically when the table is dropped
```

### 11.4.1 No Per-Tenant Migration

Because all tenants share one schema, `migrate up` is run once at deploy time. There is no need to loop over tenants or track per-tenant migration state. The `schema_migrations` table is a single global table in the shared schema.

---

## 11.5 Rolling Migrations — Zero-Downtime Patterns

Making schema changes to a live system without downtime requires discipline. The shared schema model (all tenants, all production traffic) makes this important.

### 11.5.1 Expand-Contract Pattern

Never make a column `NOT NULL` in the same migration that adds it unless you also supply a default value. The correct sequence for adding a required column to a populated table:

**Phase 1 — Expand** (can deploy immediately):
```sql
ALTER TABLE invoices ADD COLUMN exchange_rate numeric(20,8) DEFAULT 1.0;
```

**Deploy application code** that dual-writes the new field.

**Phase 2 — Backfill** (run as a background job, not a migration):
```sql
UPDATE invoices SET exchange_rate = 1.0 WHERE exchange_rate IS NULL;
```

**Phase 3 — Contract** (after backfill is complete):
```sql
ALTER TABLE invoices ALTER COLUMN exchange_rate SET NOT NULL;
ALTER TABLE invoices ALTER COLUMN exchange_rate DROP DEFAULT;
```

### 11.5.2 Multi-Phase Column Renames

PostgreSQL cannot rename a column without taking a brief lock. For truly zero-downtime column renames across a large table:

1. Add the new column (nullable, no default yet).
2. Deploy code that writes both old and new columns.
3. Backfill new column from old.
4. Deploy code that reads from new column only.
5. Make new column NOT NULL.
6. Drop old column.

Each step is a separate migration (or a background job, for the backfill).

### 11.5.3 Concurrent Index Creation

Creating an index on a large table with `CREATE INDEX` takes a table lock and blocks writes for minutes. Always use `CREATE INDEX CONCURRENTLY`:

```sql
-- Safe for production — does not block writes
CREATE INDEX CONCURRENTLY idx_journal_lines_account_period
ON journal_entry_lines(tenant_id, account_id, period_id);
```

**Note**: `CREATE INDEX CONCURRENTLY` cannot run inside a transaction. If your migration file wraps everything in a transaction, you must use a separate migration file for concurrent index creation, or configure `golang-migrate` to run that file outside a transaction block.

Use the `-- migrate: notransaction` directive at the top of the migration file:

```sql
-- migrate: notransaction
CREATE INDEX CONCURRENTLY idx_journal_lines_account_period
ON journal_entry_lines(tenant_id, account_id, period_id);
```

### 11.5.4 Operations That Always Require Downtime

Some DDL operations take aggressive locks that cannot be made concurrent:
- `ALTER TABLE ... ALTER COLUMN TYPE` (changing a column's data type)
- `VACUUM FULL` (full table rewrite)
- Adding a `CHECK` constraint that requires table scan with validation

For these, schedule a maintenance window. Document the expected lock duration. Have the rollback migration ready.

---

## 11.6 Drift Detection

Schema drift occurs when someone applies DDL directly to the database — a hotfix, a debug index, an emergency column — without writing a corresponding migration. This breaks the contract between the migration files and the actual schema.

Drift detection compares the schema derived by applying all migration files to a fresh database against the current live schema. Differences indicate drift.

In CI, the drift detection job:
1. Creates a temporary database.
2. Runs `awo migrate up` on it.
3. Uses `pg_dump --schema-only` on both the temp and production databases.
4. Diffs the two dumps.
5. Fails the build if there are differences.

Resolving drift:
- If the drift was intentional and correct: write a migration that recreates the change and commit it.
- If the drift was a mistake: drop the drifted object from production (in a maintenance window if needed), do not write a migration for it.

---

## 11.7 Complex DDL — Views, Functions, and Triggers

### 11.7.1 SQL Views

Views are used for:
- Reporting denormalised data (flattening entity hierarchies for BI tools)
- Cross-entity joins that would be too expensive or complex to express in application code
- Read models for frequently queried aggregations

Views go in migration files like any other DDL. Use `CREATE OR REPLACE VIEW` so that updating a view's definition does not require dropping and recreating it (which would require dropping dependent objects first):

```sql
-- 20240901090000_create_supplier_balance_view.up.sql

CREATE OR REPLACE VIEW supplier_balances AS
SELECT
    s.tenant_id,
    s.id AS supplier_id,
    s.name AS supplier_name,
    COALESCE(SUM(CASE WHEN jl.type = 'CREDIT' THEN jl.amount ELSE 0 END), 0) AS total_credits,
    COALESCE(SUM(CASE WHEN jl.type = 'DEBIT'  THEN jl.amount ELSE 0 END), 0) AS total_debits,
    COALESCE(SUM(CASE WHEN jl.type = 'CREDIT' THEN jl.amount ELSE -jl.amount END), 0) AS balance
FROM suppliers s
LEFT JOIN journal_entry_lines jl ON jl.party_id = s.id
    AND jl.party_type = 'SUPPLIER'
GROUP BY s.tenant_id, s.id, s.name;

-- RLS on views: the view queries underlying tables which already have RLS policies.
-- The view inherits RLS from its base tables — no separate policy needed.
```

Down migration:
```sql
DROP VIEW IF EXISTS supplier_balances;
```

### 11.7.2 The `set_tenant_context` Stored Function

This function is the RLS context setter. It is defined in the initial schema migration and must never be dropped:

```sql
CREATE OR REPLACE FUNCTION set_tenant_context(p_tenant_id uuid)
RETURNS void AS $$
DECLARE
    v_status varchar;
BEGIN
    -- Verify tenant exists and is active
    SELECT status INTO v_status
    FROM tenants
    WHERE id = p_tenant_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'TENANT_NOT_FOUND: %', p_tenant_id
            USING ERRCODE = 'P0001';
    END IF;

    IF v_status != 'ACTIVE' THEN
        RAISE EXCEPTION 'TENANT_NOT_ACTIVE: % (status: %)', p_tenant_id, v_status
            USING ERRCODE = 'P0002';
    END IF;

    -- Set the session-level variable used by RLS policies
    PERFORM set_config('app.current_tenant_id', p_tenant_id::text, false);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
```

`SECURITY DEFINER` means the function runs with the privileges of its owner (the migration user), not the caller. The application role (`awo_app`) is granted execute permission:

```sql
GRANT EXECUTE ON FUNCTION set_tenant_context(uuid) TO awo_app;
```

### 11.7.3 Naming Series Counter Function

Atomic sequence generation for document naming series uses a function that increments a counter and returns the new value, all within a single atomic operation:

```sql
CREATE OR REPLACE FUNCTION next_series_value(
    p_tenant_id uuid,
    p_series_key varchar,
    p_reset_month boolean DEFAULT false
)
RETURNS bigint AS $$
DECLARE
    v_value bigint;
BEGIN
    INSERT INTO naming_series_counters (tenant_id, series_key, current_value, last_reset)
    VALUES (p_tenant_id, p_series_key, 1, date_trunc('month', now()))
    ON CONFLICT (tenant_id, series_key) DO UPDATE
        SET current_value = CASE
            WHEN p_reset_month
                AND naming_series_counters.last_reset < date_trunc('month', now())
            THEN 1
            ELSE naming_series_counters.current_value + 1
        END,
        last_reset = CASE
            WHEN p_reset_month
                AND naming_series_counters.last_reset < date_trunc('month', now())
            THEN date_trunc('month', now())
            ELSE naming_series_counters.last_reset
        END
    RETURNING current_value INTO v_value;

    RETURN v_value;
END;
$$ LANGUAGE plpgsql;
```

This function is safe for concurrent access — the `INSERT ... ON CONFLICT DO UPDATE ... RETURNING` is atomic.

### 11.7.4 The `updated_at` Auto-Update Trigger

Rather than requiring every table to define its own trigger, a shared trigger function updates `updated_at` on any row:

```sql
-- Defined once in the initial schema migration
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
```

Each table gets a trigger that calls this function:

```sql
CREATE TRIGGER {table_name}_updated_at
    BEFORE UPDATE ON {table_name}
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
```

This is included in every `CREATE TABLE` migration file as a boilerplate block.

### 11.7.5 RLS Policies in Migrations

Every tenant-scoped table gets the same RLS policy pattern:

```sql
ALTER TABLE {table_name} ENABLE ROW LEVEL SECURITY;

CREATE POLICY {table_name}_tenant_isolation ON {table_name}
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);
```

This is the only RLS policy most tables need. The `set_tenant_context()` function sets the session variable; the policy enforces it.

For tables accessible to superuser/admin roles that need to see all tenants' data (the platform admin panel), add a bypass policy:

```sql
CREATE POLICY {table_name}_superuser_bypass ON {table_name}
    USING (
        current_setting('app.current_tenant_id', true) IS NULL
        OR tenant_id = current_setting('app.current_tenant_id')::uuid
    );
```

The `true` argument to `current_setting` suppresses the error if the variable is not set, returning `NULL` instead. A null current tenant means the connection is in admin/migration context and sees all rows.

The application role is configured with `ROW SECURITY` enforced — it does not have the `BYPASSRLS` privilege. Only the superuser and migration role can bypass RLS.

### 11.7.6 Keep All DDL in Migration Files

The rule is absolute: **never apply DDL manually in production**. This includes:
- Creating or dropping indexes
- Creating or modifying views
- Creating or modifying functions
- Creating or modifying triggers
- Modifying RLS policies
- Adding or dropping columns

Any DDL not in a migration file creates drift. Drift makes future migrations unpredictable and creates a discrepancy between what the codebase says the schema is and what it actually is.

If there is an emergency that requires manual DDL (e.g., a blocking index must be dropped immediately), document it, write the corresponding migration file immediately after, and run it through `awo migrate force` to record the current state.

---

## 11.8 Testing Migrations

### 11.8.1 Migration Round-Trip Test

The CI pipeline runs a round-trip test on every migration:

1. Create a fresh database.
2. Run `awo migrate up` — all migrations must apply without error.
3. Compare schema against the Go entity definitions (checks for missed fields or columns).
4. Run `awo migrate down` to the beginning — all down migrations must run without error.
5. Verify the database is empty (only the `schema_migrations` table should remain).

This test catches:
- Missing down migrations
- Down migrations that leave residual objects
- Up migrations that conflict with each other
- Syntax errors in SQL files

### 11.8.2 Dangerous Migration Detection in CI

The CI pipeline runs a linter over pending migration files to flag dangerous patterns:

| Pattern | Classification | Required review |
|---------|---------------|-----------------|
| `DROP TABLE` | Destructive | Senior engineer + DBA |
| `DROP COLUMN` | Destructive | Senior engineer |
| `ALTER COLUMN TYPE` | Potentially blocking | DBA review |
| `CREATE INDEX` without `CONCURRENTLY` | Blocking on large tables | Flag for review |
| Table without `ENABLE ROW LEVEL SECURITY` | Security gap | Block merge |
| Table without `tenant_id` column | Architecture violation | Block merge |

---

## 11.9 Rollback Procedures

### 11.9.1 Standard Rollback

If a migration caused a problem in staging or production:

```bash
# Roll back the most recent migration
awo migrate down 1

# Verify the version
awo migrate version

# Fix the migration file
# ...

# Re-apply
awo migrate up
```

### 11.9.2 When Rollback Is Not Possible

Some up migrations cannot be reversed cleanly:
- A migration that deletes data (there is no way to recover the deleted rows from the down migration).
- A migration that renames a column while the application is already writing to the new name (the old code expecting the old name will break on rollback).
- A migration that changes column types with precision loss.

In these cases, the path forward is:
1. Fix the application code to handle both states.
2. Write a new forward migration that corrects the problem.
3. Apply the new migration.

Document all cases where rollback was not possible in the migration file's comments so future developers understand the constraint.

### 11.9.3 Emergency `migrate force`

If a migration partially succeeded (e.g., it timed out after applying some but not all statements), the database is in a `dirty` state and `golang-migrate` will refuse to run:

```bash
$ awo migrate up
error: Dirty database version 20240801100000. Fix and force version.
```

The procedure:
1. Inspect the database to determine which statements from the migration actually ran.
2. Manually apply the remaining statements.
3. Verify the schema is correct.
4. Run `awo migrate force 20240801100000` to clear the dirty flag.
5. Confirm with `awo migrate version`.

This procedure requires direct database access and should only be performed by a DBA or senior engineer with a clear understanding of the migration's intended effect.
