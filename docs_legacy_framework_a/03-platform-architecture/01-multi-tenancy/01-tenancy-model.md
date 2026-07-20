> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Tenancy Model
portal: 3 — Platform Architecture
section: 01-multi-tenancy
audience: [architect, backend-engineer, tech-lead]
related:
  - "[RLS Enforcement](02-rls-enforcement.md)"
  - "[Entity Scope](03-entity-scope.md)"
  - "[System Overview](../00-overview/01-system-overview.md)"
---

# Tenancy Model

AwoERP uses **shared-database, shared-schema** multi-tenancy. Every table carries a `tenant_id` column and PostgreSQL Row-Level Security enforces isolation at the database layer.

## Tenancy Layers

```
Layer 1 — PostgreSQL RLS
    Every query runs inside a transaction where:
    SET LOCAL app.tenant_id = '<uuid>'
    RLS policy filters: WHERE tenant_id = current_setting('app.tenant_id')::uuid

Layer 2 — Application EntityScope
    Within a tenant, users may be scoped to:
    - All entities (EntityScopeAll)
    - A subtree of the entity hierarchy (EntityScopeSubtree)
    - A single entity (EntityScopeEntity)
```

Layer 1 prevents cross-tenant data access at the database level. Layer 2 restricts which entities within a tenant a user can see.

## Tenant Lifecycle

```
PENDING → ACTIVE → SUSPENDED → ARCHIVED
              ↑         │
              └─────────┘  (reactivate from suspended)
```

| Status | What it means |
|--------|--------------|
| PENDING | Provisioned, not yet activated — no user login allowed |
| ACTIVE | Fully operational |
| SUSPENDED | Temporarily blocked — data preserved, login denied |
| ARCHIVED | Soft-terminated — data retained for compliance, no access |

## Tenant in the Session

Every authenticated request carries a `ResolvedSession`. The tenant ID is resolved once at login and embedded in the session token — it is never re-fetched per request.

```go
type ResolvedSession struct {
    TenantID    uuid.UUID
    UserID      uuid.UUID
    EntityScope EntityScope
    Features    map[string]bool
    Settings    map[string]string
    ExpiresAt   time.Time
}
```

The session is the **canonical source** of `tenant_id` for all operations. Never accept `tenant_id` from URL parameters or request bodies.

## Tenant Provisioning

Tenant creation is a transactional operation that initializes:

1. `tenants` row (PENDING status)
2. Default roles and permissions (from config templates)
3. Root entity in the entity hierarchy
4. Default feature flag overrides (from global defaults)
5. Initial admin user invitation

After provisioning completes, the tenant transitions PENDING → ACTIVE.

## Cross-Tenant Operations

Cross-tenant operations are prohibited for regular users. Platform admins (with `superadmin` principal) may perform cross-tenant queries using a special context that bypasses RLS:

```go
// ONLY for superadmin platform operations — never in module code
store.WithoutTenant(ctx, func(q *sqlc.Queries) error {
    // RLS bypassed — use with extreme care
})
```

Module code never calls `WithoutTenant`. Only platform admin tooling uses it.
