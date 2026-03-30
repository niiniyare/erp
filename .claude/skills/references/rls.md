# RLS & Multi-Tenancy Reference

## Session variable

```sql
app.current_tenant_id   -- UUID, set per-transaction via SET LOCAL
```

## RLS policy pattern

```sql
-- Every tenant-scoped table must have this policy
ALTER TABLE <table> ENABLE ROW LEVEL SECURITY;
ALTER TABLE <table> FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON <table>
  USING (tenant_id = current_setting('app.current_tenant_id')::uuid);
```

## Request lifecycle — how tenant reaches the DB

```
X-Tenant-ID header  (highest priority)
?tenant_id= param
subdomain: tenant1.domain.com  (strips "bo." prefix for back-office)
        ↓
TenantMiddleware  (internal/api/middleware/tenant.go)
  1. Parse tenant identity
  2. TenantService.Get() — cache-first (Redis, TTL ~5 min)
  3. Assert tenant.Status == ACTIVE
  4. c.Locals(shared.TenantIDKey, tenantID)
  5. shared.WithTenantFromCtxID(ctx, tenantID)  → stores in Go context
  6. store.SetTenantContextFromCtx(ctx)  → SET LOCAL app.current_tenant_id
        ↓
Handler receives ctx — all SQLC queries filtered by RLS automatically
```

## Context helpers  (internal/shared/context.go)

```go
// Write (middleware layer)
ctx = shared.WithTenantID(ctx, tenantID)
ctx = shared.WithUserID(ctx, userID)
ctx = shared.WithSession(ctx, sess)

// Read (service / repository layer)
tenantID, ok := shared.GetTenantID(ctx)
userID, ok   := shared.GetUserID(ctx)
sess, ok     := shared.GetSession(ctx)
```

## WithTenant pattern (repository layer)

```go
// Always wrap queries in WithTenantFromCtx — never run raw queries without it
err := r.store.WithTenant(ctx, func(ctx context.Context, s db.Store) error {
    rows, err := s.ListFoo(ctx, db.ListFooParams{Limit: 50})
    if err != nil {
        return err
    }
    result = rows
    return nil
})
```

`WithTenantFromCtx` opens a transaction, runs `SET LOCAL app.current_tenant_id = '<uuid>'`, executes
the callback, and commits. RLS then enforces the filter automatically.

## Public endpoint bypass

`internal/api/middleware/whitelist.go` — `EndpointWhitelist` lists exact paths that skip
tenant resolution entirely (e.g. `POST /api/v1/tenants`, `GET /health/`).

Add new public paths here — do not remove the TenantMiddleware from the group.

## Tenant status transitions

```
PENDING → ACTIVE → SUSPENDED → ACTIVE
   │           │
   └───────────┴──► ARCHIVED
```

Enforced in `internal/core/tenant/service.go`. Validate status before any tenant-scoped operation.

## Entity-scoped visibility (ABAC layer)

Sessions embed `EntityScope` that narrows queries within the tenant:

```go
type EntityScopeType string
const (
    EntityScopeAll     EntityScopeType = "all"      // admins — no extra WHERE
    EntityScopeSubtree EntityScopeType = "subtree"  // ltree: entity_path <@ prefix
    EntityScopeEntity  EntityScopeType = "entity"   // exact: entity_id = $1
)
```

Check `sess.EntityScope.Type` in repositories that need sub-tenant scoping.

## Adding a new tenant-scoped table

1. Write migration in `db/migration/<next_seq>_add_<table>.up.sql`
2. Include `tenant_id UUID NOT NULL REFERENCES tenants(id)` column
3. Add RLS policy (pattern above)
4. Write SQLC query in `db/queries/<domain>.sql`
5. Run `make sqlc`
6. Implement repository interface + SQLC adapter
