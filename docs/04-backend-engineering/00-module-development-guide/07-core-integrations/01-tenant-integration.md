---
title: Tenant Integration
portal: 4 — Backend Engineering
section: 00-module-development-guide/07-core-integrations
audience: [backend-engineer, tech-lead]
related:
  - path: ./02-iam-integration.md
    title: IAM Integration
  - path: ../05-repository-layer/03-with-tenant-pattern.md
    title: WithTenant Pattern
---

# Tenant Integration

Every module is a tenant-scoped module. This page describes how tenancy flows from the session through to the database.

## Tenancy Layers

AwoERP enforces tenant isolation at two layers:

| Layer | Mechanism | Where enforced |
|-------|-----------|----------------|
| Layer 1 | PostgreSQL RLS (`app.tenant_id` GUC) | `store.WithTenant()` in repository |
| Layer 2 | Application EntityScope (branch/dept) | Service + repository filter params |

Layer 1 is the hard boundary. Layer 2 is the soft boundary — it restricts which records within a tenant the user can see based on their organisational unit.

## Session as Tenancy Source

The session (`ResolvedSession`) is the canonical source of tenancy information. It is populated at login and attached to the Fiber context:

```go
// Session fields relevant to tenancy
sess.TenantID    // Layer 1: PostgreSQL tenant isolation
sess.EntityScope // Layer 2: organisational unit scope
```

The handler always extracts tenancy from the session, never from URL parameters:

```go
// CORRECT
sess := c.Locals(iam.LocalsKeySession).(*iam.ResolvedSession)
tenantID := sess.TenantID

// WRONG — never trust URL params for identity
tenantID, _ := uuid.Parse(c.Params("tenantID"))
```

If the route has `:tenantID` in the URL path (for multi-tenant admin endpoints), the handler may use it to display data, but must verify it matches the session's `TenantID`:

```go
urlTenantID, _ := uuid.Parse(c.Params("tenantID"))
if urlTenantID != sess.TenantID && !sess.IsPlatform() {
	return fiber.ErrForbidden
}
```

## EntityScope

`EntityScope` defines which part of the organisational hierarchy the user can access:

| Scope type | Meaning | Filter behaviour |
|-----------|---------|-----------------|
| `EntityScopeAll` | Can see all entities in tenant | No entity_id filter |
| `EntityScopeSubtree` | Can see entity + descendants | Filter by path prefix |
| `EntityScopeEntity` | Can see only their entity | Filter by exact entity_id |

The handler translates EntityScope to repository params:

```go
func entityFilter(scope iam.EntityScope) *uuid.UUID {
	switch scope.Type {
	case iam.EntityScopeAll:
		return nil  // no entity filter
	case iam.EntityScopeEntity, iam.EntityScopeSubtree:
		entityID := uuid.MustParse(scope.EntityID)
		return &entityID
	default:
		return nil
	}
}
```

For `EntityScopeSubtree`, a path-prefix filter may be needed. Check the tenant module's hierarchy queries for the correct pattern.

## Setting EntityID on New Records

When creating a record, `entity_id` comes from the session's scope:

```go
// handler
entityID := uuid.MustParse(sess.EntityScope.EntityID)

contract, err := h.service.Create(ctx, service.CreateContractRequest{
	TenantID: sess.TenantID,
	EntityID: entityID,   // from session scope, not request body
	// ...
})
```

Never allow the client to specify an arbitrary `entity_id` unless the user explicitly has permission to create records in a different entity (admin cross-entity create).

## Tenant Context in Cross-Service Calls

When calling another platform service (audit, notifications) inside a request handler, pass the tenant ID from the session:

```go
s.auditSvc.Record(ctx, audit.Event{
	TenantID: tenantID,  // from session — not re-fetched
	// ...
})
```

The `ctx` carries the same tenant context. Platform services that query the database also use `WithTenant`, so they automatically scope to the correct tenant as long as the GUC is set.
