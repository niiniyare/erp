---
title: Database Layer Overview
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer]
related:
  - "[Domain Layer](../02-domain-layer/01-domain-overview.md)"
  - "[Repository Layer](../05-repository-layer/01-repository-overview.md)"
  - "[Schema Conventions](../../../03-platform-architecture/03-data-architecture/02-schema-conventions.md)"
  - "[Migration Strategy](../../../03-platform-architecture/03-data-architecture/03-migration-strategy.md)"
  - "[Worked Example: Database Layer](../23-worked-example/03-database-layer.md)"
---

# Database Layer Overview

## Migration Files

Every schema change is a migration file pair in `db/migration/`:

```
db/migration/
  011001_create_contracts.up.sql
  011001_create_contracts.down.sql
  011002_create_contract_lines.up.sql
  011002_create_contract_lines.down.sql
```

**Naming convention**: `{group}{seq:04d}_{description}.{up|down}.sql`

Module group numbers:
```
01xxxx  Core / platform
02xxxx  IAM
03xxxx  Tenant
05xxxx  Entities
11xxxx  Contracts
12xxxx  Finance
```

## Table Template

Every business table follows this structure:

```sql
CREATE TABLE contracts (
    -- Identity
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   uuid NOT NULL REFERENCES tenants(id),

    -- Business fields
    contract_number  text NOT NULL,
    title            text NOT NULL,
    status           text NOT NULL DEFAULT 'draft'
                     CHECK (status IN ('draft','under_review','approved','active','suspended','terminated')),
    contract_type    text NOT NULL
                     CHECK (contract_type IN ('service','supply','license','framework')),
    total_value      numeric(20,6) NOT NULL DEFAULT 0,
    currency         char(3) NOT NULL DEFAULT 'USD',
    start_date       date NOT NULL,
    end_date         date NOT NULL,
    description      text,

    -- Relationships
    vendor_id   uuid REFERENCES entities(id),
    entity_id   uuid REFERENCES entities(id),

    -- Control fields (always last)
    version     integer NOT NULL DEFAULT 1,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    deleted_at  timestamptz,

    -- Constraints
    CONSTRAINT contracts_number_tenant_unique UNIQUE (tenant_id, contract_number)
);
```

Column order: identity → business → relationships → control. Never deviate.

## Required Setup After CREATE TABLE

```sql
-- 1. Enable RLS (required for every tenant-scoped table)
ALTER TABLE contracts ENABLE ROW LEVEL SECURITY;

-- 2. Tenant isolation policy
CREATE POLICY tenant_isolation ON contracts
    USING (tenant_id = current_setting('app.tenant_id')::uuid);

-- 3. Auto-update updated_at
CREATE TRIGGER contracts_updated_at
    BEFORE UPDATE ON contracts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 4. Indexes
CREATE INDEX contracts_tenant_id_status ON contracts (tenant_id, status)
    WHERE deleted_at IS NULL;
CREATE INDEX contracts_tenant_id_created_at ON contracts (tenant_id, created_at DESC)
    WHERE deleted_at IS NULL;
```

## Monetary Columns

Always `numeric(20,6)`:

```sql
total_value   numeric(20,6) NOT NULL DEFAULT 0,
unit_price    numeric(20,6) NOT NULL DEFAULT 0,
discount_pct  numeric(5,4) NOT NULL DEFAULT 0,  -- percentage: 0.1500 = 15%
```

Never `decimal`, never `float`, never `real`, never `money`.

## Status Columns

Always `text NOT NULL` with a CHECK constraint:

```sql
status text NOT NULL DEFAULT 'draft'
    CHECK (status IN ('draft','under_review','approved','active','suspended','terminated'))
```

Never `integer` or an enum type. Text is portable, readable in audit logs, and trivially extended.

## Views

Create views for common query patterns:

```sql
-- Active contracts with entity info
CREATE VIEW v_active_contracts AS
SELECT c.*,
       e.name AS entity_name,
       e.code AS entity_code
FROM contracts c
LEFT JOIN entities e ON e.id = c.entity_id
WHERE c.deleted_at IS NULL
  AND c.status = 'active';
```

Views inherit RLS from the base table — no separate policy needed.

## Zero-Downtime Migration Rules

| Change | Approach |
|--------|---------|
| Add column | `ADD COLUMN nullable_col type` — always nullable first |
| Add NOT NULL column | Add nullable, backfill, then `SET NOT NULL` |
| Add index | `CREATE INDEX CONCURRENTLY` in `notransaction` migration |
| Rename column | Add new, dual-write, migrate reads, drop old (expand-contract) |
| Drop column | Only after verifying no app code references it |
| Add constraint | `NOT VALID` first, then `VALIDATE CONSTRAINT` separately |

Never `ALTER TABLE ... ADD COLUMN col type NOT NULL` without a default on a table with existing rows — takes an `ACCESS EXCLUSIVE` lock.
