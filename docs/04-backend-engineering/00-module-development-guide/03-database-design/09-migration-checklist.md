---
title: Migration Checklist
portal: 4 — Backend Engineering
section: 00-module-development-guide/03-database-design
audience: [backend-engineer, tech-lead]
related:
  - path: ./01-schema-overview.md
    title: Schema Overview
  - path: ./02-primary-table.md
    title: Primary Table Migration
---

# Migration Checklist

Review every migration file against this checklist before running `make migrate-up` or opening a pull request.

## File-Level Checks

- [ ] **Filename follows convention**: `NNN_SSS_create_<name>.sql` where NNN = module group, SSS = sequence within module.
- [ ] **Both UP and DOWN sections present**: `-- +migrate Up` and `-- +migrate Down` (or tool-specific equivalents).
- [ ] **DOWN section reverses UP completely**: `DROP TABLE IF EXISTS`, `DROP VIEW IF EXISTS`, etc.
- [ ] **File is idempotent on UP**: uses `CREATE TABLE IF NOT EXISTS`, `CREATE INDEX IF NOT EXISTS`, `CREATE POLICY IF NOT EXISTS` where applicable.

## Table Checks

- [ ] **Primary key**: `uuid NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY` — no `SERIAL`, no `BIGSERIAL`.
- [ ] **tenant_id**: `uuid NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT`.
- [ ] **entity_id**: `uuid NOT NULL REFERENCES entities(id) ON DELETE RESTRICT`.
- [ ] **version**: `integer NOT NULL DEFAULT 1`.
- [ ] **created_by / updated_by**: `uuid NOT NULL`.
- [ ] **deleted_at**: `timestamptz` (no `NOT NULL`).
- [ ] **created_at**: `timestamptz NOT NULL DEFAULT now()`.
- [ ] **updated_at**: `timestamptz NOT NULL DEFAULT now()`.
- [ ] **Status column**: `VARCHAR(50) NOT NULL CHECK (status IN (...))` with all valid states listed.
- [ ] **Monetary columns**: `numeric(20,6)` — not `float`, `real`, `double precision`.
- [ ] **Cross-context FKs** (vendor_id, etc.): no `REFERENCES` to foreign module tables; has a `COMMENT` explaining the omission.
- [ ] **Table COMMENT**: `COMMENT ON TABLE ... IS '...'`.

## RLS Checks

- [ ] **RLS enabled**: `ALTER TABLE ... ENABLE ROW LEVEL SECURITY;`.
- [ ] **Isolation policy**: `CREATE POLICY <table>_tenant_isolation ON <table> USING (tenant_id = current_setting('app.tenant_id')::uuid);`.
- [ ] **Policy name**: `<tablename>_tenant_isolation`.
- [ ] No `missing_ok` on `current_setting` — must be `current_setting('app.tenant_id')::uuid` not `current_setting('app.tenant_id', true)`.

## Index Checks

- [ ] **No unindexed FKs**: every FK column has a corresponding `CREATE INDEX`.
- [ ] **Partial index on status+tenant**: `CREATE INDEX idx_<table>_tenant_status ON <table> (tenant_id, status) WHERE deleted_at IS NULL`.
- [ ] **Partial index predicate**: `WHERE deleted_at IS NULL` on all standard indexes.
- [ ] **Unique constraints** for natural keys: `(tenant_id, <business_key>)`.

## Constraint Checks

- [ ] **Date range**: if table has `start_date`/`end_date`, `CHECK (end_date >= start_date)`.
- [ ] **Non-negative quantities**: `CHECK (quantity > 0)`.
- [ ] **Non-negative amounts**: `CHECK (amount >= 0)`.

## Child Table Additional Checks

- [ ] **FK to parent table**: `REFERENCES <parent>(id) ON DELETE CASCADE` (or RESTRICT with documented reason).
- [ ] **tenant_id on child**: same tenant_id requirement as parent — must be set to parent's tenant_id by service.
- [ ] **RLS on child table**: separate `ALTER TABLE` and `CREATE POLICY` — not inherited from parent.
- [ ] **FK index**: `CREATE INDEX idx_<child>_<parent_fk>` on the parent FK column.

## Trigger Checks

- [ ] **updated_at trigger**: `CREATE TRIGGER <table>_updated_at BEFORE UPDATE ON <table> FOR EACH ROW EXECUTE FUNCTION set_updated_at()`.
- [ ] `set_updated_at()` function is created in the module's first migration; subsequent migrations reuse it — do not redefine.

## Testing Before PR

```bash
# Apply migration
make migrate-up

# Verify rollback
make migrate-down

# Re-apply (must succeed)
make migrate-up

# Run integration tests (uses test DB)
make test
```

All three steps must succeed. A migration that applies cleanly but fails to roll back is rejected in review.

## Common Review Failures

| Issue | Fix |
|-------|-----|
| `SERIAL` or `BIGSERIAL` PK | Replace with `uuid DEFAULT gen_random_uuid()` |
| Missing `tenant_id` on child table | Add `tenant_id uuid NOT NULL REFERENCES tenants(id)` |
| Missing RLS policy | Add `ALTER TABLE ... ENABLE ROW LEVEL SECURITY` + `CREATE POLICY` |
| `float` monetary column | Replace with `numeric(20,6)` |
| Missing FK index | Add `CREATE INDEX idx_<table>_<fk_col>` |
| No DOWN section | Add all DROP statements in reverse CREATE order |
| Missing `deleted_at IS NULL` partial predicate on indexes | Add `WHERE deleted_at IS NULL` |
| `current_setting('app.tenant_id', true)` | Remove the `true` — errors must not be suppressed |
