> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: RLS Enforcement
portal: 3 — Platform Architecture
section: 01-multi-tenancy
audience: [architect, backend-engineer, tech-lead]
related:
  - "[Tenancy Model](01-tenancy-model.md)"
  - "[WithTenant Pattern](../../04-backend-engineering/00-module-development-guide/05-repository-layer/03-with-tenant-pattern.md)"
---

# RLS Enforcement

## How RLS Works in AwoERP

Every tenant-scoped table has:

1. `tenant_id uuid NOT NULL REFERENCES tenants(id)` column
2. `ALTER TABLE t ENABLE ROW LEVEL SECURITY`
3. A policy that checks the session GUC:

```sql
CREATE POLICY rls_contracts ON contracts
    USING (tenant_id = current_setting('app.tenant_id')::uuid);
```

When PostgreSQL evaluates a query, the RLS policy acts as an invisible `WHERE` clause. No application code can query another tenant's rows — the database enforces it.

## GUC Mechanism

`SET LOCAL app.tenant_id = '<uuid>'` sets a transaction-local configuration variable. `SET LOCAL` is scoped to the current transaction — it is automatically cleared when the transaction commits or rolls back.

```sql
BEGIN;
SET LOCAL app.tenant_id = 'aaaaaaaa-0000-0000-0000-000000000001';
SELECT * FROM contracts;  -- only rows where tenant_id matches
COMMIT;
-- GUC cleared — next transaction has no tenant set
```

## WithTenant in Go

The `Store.WithTenant` helper wraps every repository operation:

```go
func (s *pgStore) WithTenant(ctx context.Context, tenantID uuid.UUID, fn func(*sqlc.Queries) error) error {
    return s.pool.AcquireFunc(ctx, func(conn *pgxpool.Conn) error {
        tx, err := conn.Begin(ctx)
        if err != nil {
            return err
        }
        defer tx.Rollback(ctx)

        _, err = tx.Exec(ctx, "SET LOCAL app.tenant_id = $1", tenantID.String())
        if err != nil {
            return fmt.Errorf("set tenant context: %w", err)
        }

        if err := fn(sqlc.New(tx)); err != nil {
            return err
        }
        return tx.Commit(ctx)
    })
}
```

**Rule**: `missing_ok` must never be used in RLS policies. `current_setting('app.tenant_id', true)` would return an empty string when the GUC is not set, allowing the policy to evaluate `tenant_id = ''::uuid` which would error silently or match nothing — a false sense of security. Without `missing_ok`, PostgreSQL raises an error if the GUC is not set, which is the correct failure mode.

## Zero UUID Safety

```go
// In repository — always validate before WithTenant
if tenantID == uuid.Nil {
    return nil, fmt.Errorf("tenantID must not be nil")
}
```

A zero UUID would set `app.tenant_id = '00000000-0000-0000-0000-000000000000'` and potentially match rows if any tenant has that UUID (impossible in practice, but still a programming error that should fail fast).

## RLS on Views

Views inherit RLS from their underlying tables when queried through a role that does not bypass RLS. Materialized views do not — they are populated at refresh time by a privileged role and must have their own RLS policies or be restricted to read-only roles.

## Superuser Bypass

PostgreSQL superusers bypass RLS. The application database user must NOT be a superuser. Use a dedicated application role with `BYPASSRLS` only for the migration runner, and only during migrations.

## Verifying RLS in Tests

```sql
-- Confirm policy exists
SELECT tablename, policyname, cmd, qual
FROM pg_policies
WHERE schemaname = 'public' AND tablename = 'contracts';

-- Confirm RLS is enabled
SELECT relname, relrowsecurity
FROM pg_class
WHERE relname = 'contracts';
```
