> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Entity Scope
portal: 3 — Platform Architecture
section: 01-multi-tenancy
audience: [architect, backend-engineer, tech-lead]
related:
  - "[Tenancy Model](01-tenancy-model.md)"
  - "[Tenant Integration](../../04-backend-engineering/00-module-development-guide/07-core-integrations/01-tenant-integration.md)"
---

# Entity Scope

Entity scope is the second tenancy layer. Within a tenant, users may be restricted to a subset of the entity (organisational unit) hierarchy.

## EntityScope Types

```go
type EntityScopeType string

const (
    EntityScopeAll      EntityScopeType = "all"      // see all entities in tenant
    EntityScopeSubtree  EntityScopeType = "subtree"  // see entity + all descendants
    EntityScopeEntity   EntityScopeType = "entity"   // see single entity only
)

type EntityScope struct {
    Type     EntityScopeType
    EntityID string // populated when Type != All
}
```

## How Scope is Resolved

Entity scope is computed at login and stored in the `ResolvedSession`. It is not re-fetched per request.

```
Login request
    │
    ▼
SessionService.Resolve(token)
    │
    ▼
Load user's entity assignments from DB
    │
    ├── No assignment → EntityScopeAll (admin users)
    ├── Single entity assignment → EntityScopeEntity
    └── Assignment with subtree flag → EntityScopeSubtree
    │
    ▼
ResolvedSession.EntityScope set
```

## Applying Scope in Service Layer

```go
func (s *contractService) List(ctx context.Context, req ListContractsRequest) (*ListResult, error) {
    params := repository.ListContractsParams{
        TenantID: req.TenantID,
        Limit:    req.PageSize,
        Offset:   (req.Page - 1) * req.PageSize,
    }

    switch req.EntityScope.Type {
    case iam.EntityScopeAll:
        // no entity filter
    case iam.EntityScopeEntity:
        id := uuid.MustParse(req.EntityScope.EntityID)
        params.EntityID = &id
    case iam.EntityScopeSubtree:
        id := uuid.MustParse(req.EntityScope.EntityID)
        // repository handles subtree expansion via hierarchy_paths table
        params.SubtreeRootID = &id
    }

    return s.repo.List(ctx, params)
}
```

## Entity Hierarchy

Entities form a tree (company → division → department → cost center). The `hierarchy_paths` table stores all ancestor-descendant pairs (closure table pattern) for O(1) subtree queries:

```sql
-- All contracts visible to a user scoped to entity E or its descendants
SELECT c.* FROM contracts c
JOIN hierarchy_paths hp ON hp.descendant_id = c.entity_id
WHERE c.tenant_id = current_setting('app.tenant_id')::uuid
  AND hp.ancestor_id = @subtree_root_id
  AND c.deleted_at IS NULL;
```

## EntityID in Handlers

Handlers never accept `entity_id` from the URL or body for scoped operations. They always read it from the session:

```go
entityID := session.EntityID()
// EntityID() returns uuid.Nil when scope is All — handled by service layer
```

For admin operations that create data for a specific entity, the `entity_id` in the request body is validated against the session scope before use.
