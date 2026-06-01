---
title: RLS Policies
portal: 4 — Backend Engineering
section: 00-module-development-guide/03-database-design
audience: [backend-engineer, tech-lead]
related:
  - path: ./02-primary-table.md
    title: Primary Table Migration
  - path: ../05-repository-layer/03-with-tenant-pattern.md
    title: WithTenant Pattern
---

# RLS Policies

Row-Level Security (RLS) is the hard tenant isolation boundary. It operates at the PostgreSQL level, below the application. Even if the service layer has a bug that passes the wrong tenant ID, RLS prevents cross-tenant data leakage.

## How It Works

Before executing any query, the application sets a PostgreSQL GUC (Grand Unified Configuration variable):

```sql
SET app.tenant_id = '<uuid>';
```

The RLS policy reads this GUC and applies it as an implicit `WHERE` clause to every query on that table:

```sql
-- This policy is transparent to SQLC-generated queries
CREATE POLICY contracts_tenant_isolation ON contracts
    USING (tenant_id = current_setting('app.tenant_id')::uuid);
```

A query like `SELECT * FROM contracts WHERE id = $1` automatically becomes:

```sql
SELECT * FROM contracts
WHERE id = $1
AND tenant_id = current_setting('app.tenant_id')::uuid
```

The application never writes this clause manually — it is injected by PostgreSQL.

## How the GUC Is Set

`store.WithTenant()` wraps every database operation. It sets `app.tenant_id` in a transaction-local GUC before executing the queries:

```go
// Usage in repository
func (r *contractSQLCRepository) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.Contract, error) {
    var row db.Contract
    err := r.store.WithTenant(ctx, tenantID, func(q *db.Queries) error {
        var err error
        row, err = q.GetContractByID(ctx, db.GetContractByIDParams{
            ID:       id,
            TenantID: tenantID,
        })
        return err
    })
    // ...
}
```

`WithTenant` executes:
1. `SET LOCAL app.tenant_id = '<tenantID>'` — local to the transaction
2. Your query function
3. Resets to previous value on exit

`SET LOCAL` scopes the GUC to the current transaction, so it cannot bleed into a subsequent pooled connection after the transaction commits.

## Every Table Requires RLS

```sql
-- Pattern for every table in every module
ALTER TABLE <table_name> ENABLE ROW LEVEL SECURITY;

CREATE POLICY <table_name>_tenant_isolation ON <table_name>
    USING (tenant_id = current_setting('app.tenant_id')::uuid);
```

No exceptions. Platform-internal tables (schema_migrations, etc.) are exempt because they do not have tenant_id, but every business table must have both `tenant_id` and the RLS policy.

## Child Tables

Child tables also require RLS. The child table inherits isolation from its parent via the policy:

```sql
-- contract_lines is a child of contracts
ALTER TABLE contract_lines ENABLE ROW LEVEL SECURITY;

CREATE POLICY contract_lines_tenant_isolation ON contract_lines
    USING (tenant_id = current_setting('app.tenant_id')::uuid);
```

Even though `contract_lines.contract_id` already implies the tenant scope via the FK to `contracts`, the explicit `tenant_id` column and RLS policy on the child table provides defense in depth.

## Superuser and Row Security

By default, superusers and table owners bypass RLS. The application's database role must be a regular role, not a superuser. Verify with:

```sql
SELECT rolbypassrls FROM pg_roles WHERE rolname = 'erp_app';
-- Should be: f (false)
```

If `rolbypassrls` is `t`, RLS is silently bypassed for all queries from that role — the entire tenant isolation model breaks.

## Testing RLS

Test RLS directly in the database with two test tenants:

```sql
-- Set up two tenants in test data
INSERT INTO tenants (id, name, ...) VALUES
    ('aaaaaaaa-0000-0000-0000-000000000000', 'Tenant A', ...),
    ('bbbbbbbb-0000-0000-0000-000000000000', 'Tenant B', ...);

-- Insert a contract for Tenant A
SET app.tenant_id = 'aaaaaaaa-0000-0000-0000-000000000000';
INSERT INTO contracts (id, tenant_id, ...) VALUES (gen_random_uuid(), 'aaaaaaaa-...', ...);

-- Switch to Tenant B — Tenant A's row must be invisible
SET app.tenant_id = 'bbbbbbbb-0000-0000-0000-000000000000';
SELECT count(*) FROM contracts;
-- Must return: 0
```

Integration tests for the repository layer run this check automatically. See §22 Testing Guide for the test helper.

## Common Mistakes

**Calling queries outside WithTenant:**

```go
// WRONG — no GUC set, RLS will reject all rows if app.tenant_id is unset
rows, err := r.store.Queries().ListContracts(ctx, ...)
```

If `app.tenant_id` is not set, `current_setting('app.tenant_id')` throws a PostgreSQL error (unset GUC with no default). The application must always call queries through `WithTenant`.

**Using `current_setting('app.tenant_id', true)`:**

The second argument `true` (missing_ok) suppresses the error and returns `NULL` if the GUC is unset. Do **not** use `missing_ok = true`. The hard error is a safety feature — it alerts you when a query runs outside a tenant context.

**Relying on application-level tenantID filter only:**

```go
// NOT SUFFICIENT — RLS is still required even with this filter
q.GetContractByID(ctx, db.GetContractByIDParams{
    ID:       id,
    TenantID: tenantID,  // application filter — good but not the only defense
})
```

Application-level `tenant_id = $n` filters in SQLC queries are a readability and performance aid. RLS is the security boundary. Both must be present.
