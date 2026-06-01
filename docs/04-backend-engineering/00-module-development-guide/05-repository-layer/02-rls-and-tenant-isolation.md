---
title: RLS and Tenant Isolation
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer]
related:
  - "[Repository Layer Overview](01-repository-overview.md)"
  - "[Tenancy Model](../../../03-platform-architecture/01-multi-tenancy/01-tenancy-model.md)"
  - "[Database Layer](../03-database-layer/01-database-overview.md)"
---

# RLS and Tenant Isolation

## How It Works

PostgreSQL Row-Level Security (RLS) enforces tenant isolation at the database layer. Every tenant-scoped table has a policy:

```sql
-- Migration: enable RLS on every tenant table
ALTER TABLE contracts ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON contracts
    USING (tenant_id = current_setting('app.tenant_id')::uuid);
```

The policy reads `app.tenant_id` from the session local variable. All queries against the table automatically filter by this value — no `WHERE tenant_id = ?` needed in SQL.

## Setting Tenant Context

The repository uses `WithTenant` to set the session variable before executing queries:

```go
// internal/platform/db/tenant.go
package db

import (
    "context"
    "fmt"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"

    "awo.so/internal/core/iam/domain"
)

// WithTenant executes fn inside a transaction with app.tenant_id set.
// Uses SET LOCAL so the setting is scoped to the transaction only.
func WithTenant(ctx context.Context, pool *pgxpool.Pool, tenantID string, fn func(pgx.Tx) error) error {
    tx, err := pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("begin tx: %w", err)
    }
    defer tx.Rollback(ctx)

    _, err = tx.Exec(ctx, "SET LOCAL app.tenant_id = $1", tenantID)
    if err != nil {
        return fmt.Errorf("set tenant context: %w", err)
    }

    if err = fn(tx); err != nil {
        return err
    }

    return tx.Commit(ctx)
}
```

`SET LOCAL` — not `SET` — is critical. `SET LOCAL` reverts when the transaction ends, so the next connection pool checkout starts clean.

## Repository Usage Pattern

```go
func (r *contractRepository) List(ctx context.Context, tenantID uuid.UUID, params ListParams) ([]Contract, error) {
    return db.WithTenant(ctx, r.pool, tenantID.String(), func(tx pgx.Tx) error {
        // RLS automatically filters by tenant_id — no WHERE clause needed
        rows, err := r.q.WithTx(tx).ListContracts(ctx, sqlc.ListContractsParams{
            Status: params.Status,
            Limit:  int32(params.Limit),
            Offset: int32(params.Offset),
        })
        // ...
    })
}
```

The SQLC query does not include `AND tenant_id = $1`:

```sql
-- query.sql (NO tenant_id filter — RLS handles it)
-- name: ListContracts :many
SELECT * FROM contracts
WHERE deleted_at IS NULL
  AND ($1::text = '' OR status = $1)
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
```

## Connection Pool Safety

Connection pool connections are reused across requests. Using `SET` (session-level) instead of `SET LOCAL` (transaction-level) would leak the tenant ID to the next request that reuses the connection. Always use `SET LOCAL` inside a transaction.

| Setting | Scope | Safe? |
|---------|-------|-------|
| `SET app.tenant_id = '...'` | Session (connection lifetime) | **No** — leaks across requests |
| `SET LOCAL app.tenant_id = '...'` | Transaction | **Yes** — reverts on commit/rollback |

## Cross-Tenant Reads (Admin Only)

Platform-level admin operations (super-admin viewing all tenants) bypass RLS by connecting as a role that bypasses the policy:

```sql
-- Admin role bypasses tenant RLS
ALTER TABLE contracts FORCE ROW LEVEL SECURITY;  -- even table owner is subject
CREATE POLICY tenant_isolation ON contracts
    USING (tenant_id = current_setting('app.tenant_id')::uuid)
    -- No USING clause for pg_bypassrls role
;
```

In Go, admin repositories use a separate pool configured with the `pg_bypassrls` role. Regular module repositories must never use the admin pool.

## Testing RLS Isolation

```go
func TestRLSIsolation(t *testing.T) {
    // Create two tenants
    tenantA := uuid.New()
    tenantB := uuid.New()

    // Insert contract for tenant A
    err := db.WithTenant(ctx, pool, tenantA.String(), func(tx pgx.Tx) error {
        _, err := tx.Exec(ctx,
            `INSERT INTO contracts (id, tenant_id, title, status) VALUES ($1, $2, $3, $4)`,
            uuid.New(), tenantA, "Contract A", "draft",
        )
        return err
    })
    require.NoError(t, err)

    // Query as tenant B — should see 0 rows
    var count int
    err = db.WithTenant(ctx, pool, tenantB.String(), func(tx pgx.Tx) error {
        return tx.QueryRow(ctx, `SELECT COUNT(*) FROM contracts`).Scan(&count)
    })
    require.NoError(t, err)
    assert.Equal(t, 0, count, "tenant B must not see tenant A's contracts")
}
```

## What Must Never Happen

- `SET app.tenant_id` without `LOCAL` — leaks tenant context
- Passing `tenant_id` in SQL `WHERE` clauses when RLS handles it — redundant and can mask RLS misconfiguration
- Sharing the admin pool with regular module repositories
- Calling queries outside `WithTenant` (unset `app.tenant_id` returns no rows or errors, depending on `missing_ok`)
