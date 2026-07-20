> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Primary Table Migration
portal: 4 — Backend Engineering
section: 00-module-development-guide/03-database-design
audience: [backend-engineer, tech-lead]
related:
  - path: ./01-schema-overview.md
    title: Schema Overview
  - path: ./03-rls-policies.md
    title: RLS Policies
  - path: ./04-indexes.md
    title: Indexes
---

# Primary Table Migration

The first migration file creates the module's primary table, enables RLS, and creates the initial indexes. Every element is shown with explanation.

## Complete Migration: 011001_create_contracts.sql

```sql
-- +migrate Up

-- ============================================================
-- contracts
-- ============================================================
-- Stores procurement contracts. Each contract is scoped to a
-- tenant and organisational entity.

CREATE TABLE contracts (
    -- Identity
    id              uuid        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id       uuid        NOT NULL REFERENCES tenants(id)  ON DELETE RESTRICT,
    entity_id       uuid        NOT NULL REFERENCES entities(id) ON DELETE RESTRICT,

    -- Business fields
    contract_number VARCHAR(50)    NOT NULL,
    title           TEXT           NOT NULL,
    description     TEXT           NOT NULL DEFAULT '',
    vendor_id       uuid           NOT NULL,  -- FK to vendor module (cross-context, no DB FK)
    contract_type   VARCHAR(50)    NOT NULL CHECK (contract_type IN (
                                                'service',
                                                'supply',
                                                'framework',
                                                'lease'
                                            )),
    start_date      DATE           NOT NULL,
    end_date        DATE           NOT NULL,
    total_value     numeric(20,6)  NOT NULL DEFAULT 0,
    currency        VARCHAR(3)     NOT NULL DEFAULT 'USD',

    -- Lifecycle
    status          VARCHAR(50)    NOT NULL DEFAULT 'draft'
                                   CHECK (status IN (
                                       'draft',
                                       'submitted',
                                       'under_review',
                                       'approved',
                                       'active',
                                       'suspended',
                                       'terminated'
                                   )),
    version         integer        NOT NULL DEFAULT 1,

    -- Actor tracking
    created_by      uuid           NOT NULL,
    updated_by      uuid           NOT NULL,

    -- Soft delete + timestamps
    deleted_at      timestamptz,
    created_at      timestamptz    NOT NULL DEFAULT now(),
    updated_at      timestamptz    NOT NULL DEFAULT now(),

    -- Constraints
    CONSTRAINT uq_contracts_tenant_number UNIQUE (tenant_id, contract_number),
    CONSTRAINT chk_contracts_dates CHECK (end_date >= start_date),
    CONSTRAINT chk_contracts_value CHECK (total_value >= 0)
);

COMMENT ON TABLE contracts IS 'Procurement contracts: formal agreements between the tenant organisation and vendors.';
COMMENT ON COLUMN contracts.vendor_id IS 'Cross-context reference to the vendor module. No DB-level FK to allow module independence.';
COMMENT ON COLUMN contracts.version IS 'Optimistic lock counter. Incremented on every UPDATE. Used to detect concurrent writes.';
COMMENT ON COLUMN contracts.deleted_at IS 'Soft-delete sentinel. NULL = active record. Non-NULL = deleted at this timestamp.';

-- ============================================================
-- Row-Level Security
-- ============================================================

ALTER TABLE contracts ENABLE ROW LEVEL SECURITY;

-- Tenant isolation: every query implicitly filters by the current tenant.
-- The GUC app.tenant_id is set by store.WithTenant() before executing any query.
CREATE POLICY contracts_tenant_isolation ON contracts
    USING (tenant_id = current_setting('app.tenant_id')::uuid);

-- ============================================================
-- Indexes
-- ============================================================

-- Covering index for the most common list query: filter by tenant + status
CREATE INDEX idx_contracts_tenant_status
    ON contracts (tenant_id, status)
    WHERE deleted_at IS NULL;

-- Index for lookups by vendor within a tenant
CREATE INDEX idx_contracts_tenant_vendor
    ON contracts (tenant_id, vendor_id)
    WHERE deleted_at IS NULL;

-- Index for lookups by entity (org unit) within a tenant
CREATE INDEX idx_contracts_tenant_entity
    ON contracts (tenant_id, entity_id)
    WHERE deleted_at IS NULL;

-- Index for date-range queries (expiry reports, activation schedules)
CREATE INDEX idx_contracts_tenant_dates
    ON contracts (tenant_id, start_date, end_date)
    WHERE deleted_at IS NULL;

-- ============================================================
-- updated_at auto-update trigger
-- ============================================================

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER contracts_updated_at
    BEFORE UPDATE ON contracts
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- +migrate Down

DROP TRIGGER IF EXISTS contracts_updated_at ON contracts;
DROP POLICY IF EXISTS contracts_tenant_isolation ON contracts;
DROP TABLE IF EXISTS contracts;
```

## Key Decisions Explained

### No DB-Level FK for vendor_id

```sql
vendor_id uuid NOT NULL,  -- no REFERENCES vendors(id)
```

Cross-context references do not use database foreign keys. The vendor module could be on a different database in the future, or the vendor table may not exist yet when this migration runs. Application-level validation (service checks vendor exists before create) is sufficient.

This is intentional — not missing. Add a SQL comment explaining the absence to prevent "fixing" it later.

### Partial Indexes

```sql
WHERE deleted_at IS NULL
```

All indexes exclude soft-deleted rows. Soft-deleted records are rarely queried; including them in indexes wastes space and slows index maintenance. The `deleted_at IS NULL` partial index clause means the planner uses these indexes for all normal (non-deleted) queries.

### STATUS Check Constraint vs Enum Type

```sql
status VARCHAR(50) NOT NULL CHECK (status IN ('draft', 'submitted', ...))
```

Prefer `VARCHAR + CHECK` over `CREATE TYPE ... AS ENUM` for status columns. PostgreSQL enums are hard to alter — adding a value requires `ALTER TYPE` and a full table scan. `VARCHAR + CHECK` can be updated with a simple `ALTER TABLE ... DROP CONSTRAINT ...; ALTER TABLE ... ADD CONSTRAINT ...` in a subsequent migration, with zero downtime using a concurrent index rebuild.

### Monetary Values

```sql
total_value numeric(20,6) NOT NULL DEFAULT 0
```

`numeric(20,6)` stores up to 14 digits before the decimal and 6 after — sufficient for any real-world contract value. `DEFAULT 0` allows creating a contract before lines are added.

Never use `float`, `real`, or `double precision` for monetary columns. Floating-point arithmetic is inexact by design.

### Updated_at Trigger

The `set_updated_at()` trigger ensures `updated_at` is always correct even if a query is run directly in psql (bypassing the application's `updated_at = now()` in the SQLC query). This provides a safety net during migrations and debugging.
