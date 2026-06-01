---
title: Child Tables
portal: 4 — Backend Engineering
section: 00-module-development-guide/03-database-design
audience: [backend-engineer, tech-lead]
related:
  - path: ./02-primary-table.md
    title: Primary Table Migration
  - path: ./03-rls-policies.md
    title: RLS Policies
---

# Child Tables

Child tables store records that belong to a primary entity — in this module, `contract_lines` belongs to `contracts`. The second migration file creates child tables.

## Complete Migration: 011002_create_contract_lines.sql

```sql
-- +migrate Up

-- ============================================================
-- contract_lines
-- ============================================================
-- Individual deliverables or scope items within a contract.
-- Each line has its own unit price, quantity, and derived total.

CREATE TABLE contract_lines (
    -- Identity
    id           uuid        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id    uuid        NOT NULL REFERENCES tenants(id)   ON DELETE RESTRICT,
    contract_id  uuid        NOT NULL REFERENCES contracts(id) ON DELETE CASCADE,

    -- Business fields
    line_number  integer     NOT NULL,
    description  TEXT        NOT NULL,
    quantity     integer     NOT NULL CHECK (quantity > 0),
    unit_price   numeric(20,6) NOT NULL CHECK (unit_price >= 0),
    total_price  numeric(20,6) NOT NULL GENERATED ALWAYS AS (quantity * unit_price) STORED,

    -- Lifecycle
    status       VARCHAR(50) NOT NULL DEFAULT 'active'
                             CHECK (status IN ('active', 'cancelled')),
    version      integer     NOT NULL DEFAULT 1,

    -- Actor tracking
    created_by   uuid        NOT NULL,
    updated_by   uuid        NOT NULL,

    -- Soft delete + timestamps
    deleted_at   timestamptz,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),

    -- Constraints
    CONSTRAINT uq_contract_lines_contract_number UNIQUE (contract_id, line_number),
    CONSTRAINT chk_contract_lines_total CHECK (total_price >= 0)
);

COMMENT ON TABLE contract_lines IS 'Individual line items within a contract. Totals roll up to contracts.total_value.';
COMMENT ON COLUMN contract_lines.total_price IS 'Derived column: quantity * unit_price. Stored (not virtual) for index and query support.';

-- ============================================================
-- Row-Level Security
-- ============================================================

ALTER TABLE contract_lines ENABLE ROW LEVEL SECURITY;

CREATE POLICY contract_lines_tenant_isolation ON contract_lines
    USING (tenant_id = current_setting('app.tenant_id')::uuid);

-- ============================================================
-- Indexes
-- ============================================================

-- FK index — PostgreSQL does not auto-index FKs
CREATE INDEX idx_contract_lines_contract
    ON contract_lines (contract_id)
    WHERE deleted_at IS NULL;

-- Tenant-scoped queries (needed for repo list-by-tenant queries)
CREATE INDEX idx_contract_lines_tenant
    ON contract_lines (tenant_id)
    WHERE deleted_at IS NULL;

-- updated_at trigger (reuse the function created in migration 011001)
CREATE TRIGGER contract_lines_updated_at
    BEFORE UPDATE ON contract_lines
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- +migrate Down

DROP TRIGGER IF EXISTS contract_lines_updated_at ON contract_lines;
DROP POLICY IF EXISTS contract_lines_tenant_isolation ON contract_lines;
DROP TABLE IF EXISTS contract_lines;
```

## Key Decisions for Child Tables

### ON DELETE CASCADE vs RESTRICT on the FK to Parent

```sql
contract_id uuid NOT NULL REFERENCES contracts(id) ON DELETE CASCADE
```

Lines are existentially dependent on their contract — there is no purpose to a line without a parent contract. `CASCADE` means deleting a contract hard-deletes all its lines. However, since `contracts` uses soft-delete (`deleted_at`), this `CASCADE` only fires on a hard `DELETE FROM contracts` — which only happens in data cleanup scripts, not in application code.

For child tables where records might outlive their parent (e.g., audit records), use `ON DELETE RESTRICT` or `ON DELETE SET NULL`.

### Generated Stored Column for total_price

```sql
total_price numeric(20,6) NOT NULL GENERATED ALWAYS AS (quantity * unit_price) STORED
```

`total_price` is derived data — it should never be set by the application independently. Using `GENERATED ALWAYS AS ... STORED` delegates the calculation to PostgreSQL. This eliminates the class of bugs where application code and database disagree on line totals.

`STORED` means PostgreSQL writes the value to disk at write time (not computed on read). This allows indexing and avoids recalculation on every SELECT.

### tenant_id on Child Table

Even though `contract_lines.contract_id` implies the tenant scope via the FK to `contracts`, the explicit `tenant_id` column is required. Two reasons:

1. The RLS policy requires `tenant_id` directly on the table — it cannot follow a FK chain.
2. Some queries on child tables need to be tenant-scoped without joining the parent.

The application sets `contract_lines.tenant_id = contracts.tenant_id` on insert. This is enforced at the service level:

```go
func (s *contractService) AddLine(ctx context.Context, params AddContractLineParams) (*domain.ContractLine, error) {
    contract, err := s.repo.GetByID(ctx, params.ContractID, params.TenantID)
    if err != nil {
        return nil, err
    }
    // Ensure line inherits parent's tenant_id
    return s.lineRepo.Create(ctx, repository.CreateContractLineParams{
        ContractID: params.ContractID,
        TenantID:   contract.TenantID,  // from parent, not from request
        // ...
    })
}
```

### version on Child Table

Child tables have their own `version` column for optimistic locking on line-level edits. This is independent of the parent contract's version.

## Rollup to Parent

When lines are added or updated, the parent contract's `total_value` must be kept in sync. Two approaches:

**Option A — Application-level rollup** (preferred for most modules):

After adding/updating a line, the service calls `repo.UpdateTotalValue(ctx, contractID, tenantID)` which runs:

```sql
-- name: UpdateContractTotalValue :one
UPDATE contracts
SET
    total_value = (
        SELECT COALESCE(SUM(total_price), 0)
        FROM contract_lines
        WHERE contract_id = @contract_id
          AND deleted_at IS NULL
    ),
    version    = version + 1,
    updated_by = @updated_by,
    updated_at = now()
WHERE id        = @contract_id
  AND tenant_id = @tenant_id
  AND deleted_at IS NULL
RETURNING *;
```

**Option B — Database trigger** (use sparingly):

A trigger on `contract_lines` that updates `contracts.total_value` on every line insert/update/delete. Triggers are hard to observe, test, and debug. Prefer Option A unless performance demands it.
