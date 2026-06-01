---
title: Migration Cookbook
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer]
related:
  - "[Database Overview](01-database-overview.md)"
  - "[Migration Strategy](../../../03-platform-architecture/03-data-architecture/03-migration-strategy.md)"
  - "[Migration Failure Runbook](../../../09-operations/05-migration-failure.md)"
---

# Migration Cookbook

Common migration patterns for safe schema changes.

## Adding a New Table

```sql
-- 011004_create_contract_attachments.up.sql

CREATE TABLE contract_attachments (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id      uuid NOT NULL REFERENCES tenants(id),
    contract_id    uuid NOT NULL REFERENCES contracts(id),
    filename       text NOT NULL,
    file_size      bigint NOT NULL,
    content_type   text NOT NULL,
    storage_key    text NOT NULL,
    uploaded_by_id uuid,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),
    deleted_at     timestamptz
);

ALTER TABLE contract_attachments ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON contract_attachments
    USING (tenant_id = current_setting('app.tenant_id')::uuid);

CREATE TRIGGER contract_attachments_updated_at
    BEFORE UPDATE ON contract_attachments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE INDEX contract_attachments_contract_id ON contract_attachments (tenant_id, contract_id)
    WHERE deleted_at IS NULL;
```

```sql
-- 011004_create_contract_attachments.down.sql
DROP TABLE IF EXISTS contract_attachments;
```

## Adding a Nullable Column (Safe)

```sql
-- 011005_add_contract_reference_code.up.sql
ALTER TABLE contracts ADD COLUMN reference_code text;
```

No lock concern — adding a nullable column is instant in PostgreSQL.

```sql
-- 011005_add_contract_reference_code.down.sql
ALTER TABLE contracts DROP COLUMN IF EXISTS reference_code;
```

## Adding a NOT NULL Column (Expand-Contract)

**Step 1: Add nullable (this migration)**

```sql
-- 011006_add_priority_step1.up.sql
ALTER TABLE contracts ADD COLUMN priority integer;
```

**Step 2: Backfill (separate migration)**

```sql
-- 011007_add_priority_step2_backfill.up.sql
UPDATE contracts SET priority = 1 WHERE priority IS NULL;
```

Run in batches for large tables:

```sql
DO $$
DECLARE
    batch_size INT := 10000;
    updated INT;
BEGIN
    LOOP
        UPDATE contracts SET priority = 1
        WHERE id IN (
            SELECT id FROM contracts WHERE priority IS NULL LIMIT batch_size
        );
        GET DIAGNOSTICS updated = ROW_COUNT;
        EXIT WHEN updated = 0;
        PERFORM pg_sleep(0.1);   -- brief pause between batches
    END LOOP;
END $$;
```

**Step 3: Add NOT NULL constraint (separate migration)**

```sql
-- 011008_add_priority_step3_notnull.up.sql
ALTER TABLE contracts ALTER COLUMN priority SET NOT NULL;
ALTER TABLE contracts ALTER COLUMN priority SET DEFAULT 1;
```

## Adding an Index (Without Locking)

```sql
-- 011009_add_contracts_vendor_idx.up.sql
-- This migration uses CONCURRENTLY which requires running outside a transaction

CREATE INDEX CONCURRENTLY IF NOT EXISTS contracts_vendor_id_idx
    ON contracts (tenant_id, vendor_id)
    WHERE deleted_at IS NULL;
```

**Important**: SQLC migration files with `CONCURRENTLY` must not be wrapped in a transaction. Use the `--disable-transactions` flag or add a comment for `golang-migrate`:

```sql
-- migrate:notransaction
CREATE INDEX CONCURRENTLY ...
```

```sql
-- 011009_add_contracts_vendor_idx.down.sql
DROP INDEX CONCURRENTLY IF EXISTS contracts_vendor_id_idx;
```

## Adding a Check Constraint (Non-Blocking)

Add as NOT VALID first, then validate separately:

```sql
-- 011010_add_value_positive_constraint.up.sql
ALTER TABLE contracts ADD CONSTRAINT contracts_value_positive
    CHECK (total_value >= 0) NOT VALID;
```

```sql
-- 011011_validate_value_positive.up.sql
ALTER TABLE contracts VALIDATE CONSTRAINT contracts_value_positive;
```

`NOT VALID` adds the constraint but skips checking existing rows — fast. `VALIDATE CONSTRAINT` checks existing rows using a share lock (not exclusive).

## Renaming a Column (Expand-Contract)

**Never rename a column directly** — it breaks code before the deployment.

Step 1: Add new column, deploy code that writes both old and new:

```sql
ALTER TABLE contracts ADD COLUMN contract_ref text;
UPDATE contracts SET contract_ref = reference_code WHERE contract_ref IS NULL;
```

Step 2: Deploy code that reads only new column.

Step 3: Drop old column in next release:

```sql
ALTER TABLE contracts DROP COLUMN reference_code;
```

## Adding a Foreign Key

```sql
ALTER TABLE contracts ADD CONSTRAINT contracts_entity_id_fk
    FOREIGN KEY (entity_id) REFERENCES entities(id)
    NOT VALID;

-- In a later migration after verifying no orphans:
ALTER TABLE contracts VALIDATE CONSTRAINT contracts_entity_id_fk;
```

## Creating a View

```sql
-- 011012_create_contracts_summary_view.up.sql
CREATE VIEW v_contracts_summary AS
SELECT
    c.id,
    c.tenant_id,
    c.contract_number,
    c.title,
    c.status,
    c.total_value,
    c.currency,
    c.start_date,
    c.end_date,
    e.name AS entity_name,
    ven.name AS vendor_name
FROM contracts c
LEFT JOIN entities e   ON e.id = c.entity_id   AND e.deleted_at IS NULL
LEFT JOIN entities ven ON ven.id = c.vendor_id AND ven.deleted_at IS NULL
WHERE c.deleted_at IS NULL;
```

Views inherit RLS from base tables — no separate policy required.

## Rolling Back a Migration

```bash
# Roll back one step
make migrate-down

# Or directly
migrate -database "$DATABASE_URL" -path db/migration down 1
```

The `.down.sql` file is the exact inverse of the `.up.sql`. Test rollback before merging.
