> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Worked Example — Database Layer
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[Database Layer Overview](../03-database-layer/01-database-overview.md)"
  - "[Migration Cookbook](../03-database-layer/02-migration-cookbook.md)"
  - "[Worked Example Overview](01-worked-example-overview.md)"
---

# Worked Example — Database Layer

## Migration 011001 — contracts table

```sql
-- db/migration/011001_create_contracts.up.sql

CREATE TABLE contracts (
    id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid        NOT NULL REFERENCES tenants(id),
    entity_id       uuid        NOT NULL REFERENCES entities(id),
    contract_number text        NOT NULL,
    title           text        NOT NULL,
    description     text        NOT NULL DEFAULT '',
    status          text        NOT NULL DEFAULT 'draft',
    contract_type   text        NOT NULL,
    total_value     numeric(20,6) NOT NULL DEFAULT 0,
    currency        char(3)     NOT NULL DEFAULT 'USD',
    start_date      date        NOT NULL,
    end_date        date        NOT NULL,
    vendor_id       uuid        NOT NULL,
    assigned_to     uuid        REFERENCES users(id),
    version         integer     NOT NULL DEFAULT 1,
    created_by      uuid        NOT NULL REFERENCES users(id),
    updated_by      uuid        NOT NULL REFERENCES users(id),
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    deleted_at      timestamptz,
    CONSTRAINT contracts_status_check
        CHECK (status IN ('draft','submitted','under_review','approved','active','suspended','terminated')),
    CONSTRAINT contracts_type_check
        CHECK (contract_type IN ('service','goods','lease','other')),
    CONSTRAINT contracts_value_positive
        CHECK (total_value >= 0),
    CONSTRAINT contracts_dates_valid
        CHECK (end_date > start_date),
    CONSTRAINT contracts_number_unique
        UNIQUE (tenant_id, contract_number)
);

ALTER TABLE contracts ENABLE ROW LEVEL SECURITY;

CREATE POLICY rls_contracts ON contracts
    USING (tenant_id = current_setting('app.tenant_id')::uuid);

CREATE INDEX idx_contracts_tenant_status
    ON contracts (tenant_id, status)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_contracts_tenant_entity
    ON contracts (tenant_id, entity_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_contracts_end_date
    ON contracts (tenant_id, end_date)
    WHERE deleted_at IS NULL AND status = 'active';

CREATE TRIGGER contracts_updated_at
    BEFORE UPDATE ON contracts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

```sql
-- db/migration/011001_create_contracts.down.sql
DROP TABLE IF EXISTS contracts;
```

## Migration 011002 — contract_lines table

```sql
-- db/migration/011002_create_contract_lines.up.sql

CREATE TABLE contract_lines (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   uuid        NOT NULL REFERENCES tenants(id),
    contract_id uuid        NOT NULL REFERENCES contracts(id) ON DELETE RESTRICT,
    line_number integer     NOT NULL,
    description text        NOT NULL,
    quantity    numeric(20,6) NOT NULL DEFAULT 1,
    unit_price  numeric(20,6) NOT NULL DEFAULT 0,
    total_price numeric(20,6) GENERATED ALWAYS AS (quantity * unit_price) STORED,
    unit        text        NOT NULL DEFAULT 'unit',
    version     integer     NOT NULL DEFAULT 1,
    created_by  uuid        NOT NULL REFERENCES users(id),
    updated_by  uuid        NOT NULL REFERENCES users(id),
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    deleted_at  timestamptz,
    CONSTRAINT contract_lines_quantity_positive CHECK (quantity > 0),
    CONSTRAINT contract_lines_price_positive    CHECK (unit_price >= 0),
    CONSTRAINT contract_lines_number_unique     UNIQUE (contract_id, line_number)
);

ALTER TABLE contract_lines ENABLE ROW LEVEL SECURITY;

CREATE POLICY rls_contract_lines ON contract_lines
    USING (tenant_id = current_setting('app.tenant_id')::uuid);

CREATE INDEX idx_contract_lines_contract
    ON contract_lines (contract_id)
    WHERE deleted_at IS NULL;

CREATE TRIGGER contract_lines_updated_at
    BEFORE UPDATE ON contract_lines
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

```sql
-- db/migration/011002_create_contract_lines.down.sql
DROP TABLE IF EXISTS contract_lines;
```

## Migration 011003 — views

```sql
-- db/migration/011003_create_contracts_views.up.sql

CREATE VIEW v_contracts_with_lines AS
SELECT
    c.*,
    COUNT(cl.id) FILTER (WHERE cl.deleted_at IS NULL)        AS line_count,
    COALESCE(SUM(cl.total_price) FILTER (WHERE cl.deleted_at IS NULL), 0) AS lines_total_value
FROM contracts c
LEFT JOIN contract_lines cl ON cl.contract_id = c.id
WHERE c.deleted_at IS NULL
GROUP BY c.id;

COMMENT ON VIEW v_contracts_with_lines IS 'Contracts with aggregated line item totals.';
```

```sql
-- db/migration/011003_create_contracts_views.down.sql
DROP VIEW IF EXISTS v_contracts_with_lines;
```

**Gate**: Migrations written → run `make migrate-up` against a test database, verify all tables created, then `make migrate-down` to verify rollback.
