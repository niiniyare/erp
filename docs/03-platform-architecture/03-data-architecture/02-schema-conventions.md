---
title: Schema Conventions
portal: 3 — Platform Architecture
section: 03-data-architecture
audience: [architect, backend-engineer, tech-lead]
related:
  - "[Data Overview](01-data-overview.md)"
  - "[Migration Strategy](03-migration-strategy.md)"
  - "[Primary Table Design](../../04-backend-engineering/00-module-development-guide/03-database-design/02-primary-table.md)"
---

# Schema Conventions

## Primary Key

Always `uuid DEFAULT gen_random_uuid()`. Never `SERIAL` or `BIGSERIAL`.

```sql
id uuid PRIMARY KEY DEFAULT gen_random_uuid()
```

UUIDs are safe across environments (no sequence clash on restore), work with distributed ID generation, and do not leak row counts to clients.

## Monetary Values

`numeric(20,6)` — never `float`, `double precision`, or `money`.

- 20 digits of precision
- 6 decimal places (sufficient for most currencies and exchange rates)
- Go type: `github.com/shopspring/decimal`
- JSON representation: string (never float)

## Status Columns

Status columns use `text` with a `CHECK` constraint, not enums.

```sql
status text NOT NULL DEFAULT 'draft',
CONSTRAINT contracts_status_check
    CHECK (status IN ('draft','submitted','under_review','approved','active','suspended','terminated'))
```

Text over enum: adding a new status requires no `ALTER TYPE` — only a constraint update and a migration.

## Timestamps

All timestamps are `timestamptz` (with timezone). The database stores UTC; the application formats for display.

- `created_at`: set on INSERT, never updated
- `updated_at`: updated on every UPDATE via trigger
- `deleted_at`: set on soft delete, never hard-deleted

## updated_at Trigger

One reusable trigger function handles all tables:

```sql
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Applied per table:
CREATE TRIGGER contracts_updated_at
    BEFORE UPDATE ON contracts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

## Soft Delete Pattern

```sql
deleted_at timestamptz  -- NULL = active, non-NULL = deleted
```

Every query filters `AND deleted_at IS NULL`. Partial indexes include this condition:

```sql
CREATE INDEX idx_contracts_tenant_status
    ON contracts (tenant_id, status)
    WHERE deleted_at IS NULL;
```

## Foreign Keys

All FK columns are `NOT NULL` unless the relationship is optional. `ON DELETE RESTRICT` is the default — never `CASCADE` for business data (prevents accidental data loss).

```sql
contract_id uuid NOT NULL REFERENCES contracts(id) ON DELETE RESTRICT
```

The exception: junction tables for many-to-many relationships may use `ON DELETE CASCADE`.

## Generated Columns

Use `GENERATED ALWAYS AS ... STORED` for derived values that must be queryable and indexed:

```sql
total_price numeric(20,6)
    GENERATED ALWAYS AS (quantity * unit_price) STORED
```

Never compute totals in application code and persist them manually — this creates consistency risks.

## Column Order Convention

```
1. id
2. tenant_id
3. parent FK (e.g. contract_id)
4. business columns (alphabetical within logical groups)
5. version
6. created_by, updated_by
7. created_at, updated_at, deleted_at
```
