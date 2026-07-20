> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Authorization in the Service Layer
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[Service Layer Overview](01-service-overview.md)"
  - "[Authorization Model](../../../03-platform-architecture/02-iam/03-authorization-model.md)"
  - "[Authz Middleware](../16-middleware-chain/03-authz-middleware.md)"
---

# Authorization in the Service Layer

The service layer enforces permissions at two levels: route-level (middleware) and service-level (authzSvc.Enforce). This page explains the service-level call.

## Route-Level vs Service-Level Authorization

| Level | Where | Mechanism |
|-------|-------|-----------|
| Route level | `routes.go` | `middlewarePkg.Authorize(cfg, "permission")` |
| Service level | `service/<noun>.go` | `authzSvc.Enforce(ctx, Request{...})` |

Route-level is sufficient for most operations — the middleware blocks the request before it reaches the handler. Service-level is needed when:
- The permission depends on runtime data (e.g., only the contract's creator can delete it)
- A service method is called from multiple entry points (e.g., both HTTP handler and a Temporal workflow)
- A composite operation needs separate permission checks for each sub-operation

For the contracts module, service-level checks are the primary authorization mechanism. Route-level middleware is also applied for defense in depth.

## authzSvc.Enforce Call

```go
allowed, err := s.authzSvc.Enforce(ctx, iam.Request{
	Subject: iam.TenantSubject(principal.UserID()),
	Domain:  iam.TenantDomain(tenantID),
	Object:  "contracts/contract/*",
	Action:  "create",
})
if err != nil {
	return nil, err
}
if !allowed {
	return nil, iam.ErrForbidden
}
```

### Request Fields

| Field | Value | Example |
|-------|-------|---------|
| `Subject` | `iam.TenantSubject(userID)` | `"tenant:aaaaaaaa-..."` |
| `Domain` | `iam.TenantDomain(tenantID)` | `"aaaaaaaa-..."` (tenant UUID) |
| `Object` | slash-separated resource path | `"contracts/contract/*"` |
| `Action` | lowercase verb | `"create"`, `"read"`, `"update"`, `"delete"`, `"submit"`, `"approve"` |

### Object Path Format

The Casbin model uses `keyMatch2` on the Object field. The object path in Enforce must match the policy string:

```
Policy:  contracts/contract/*  ←→  matches  →  contracts/contract/create
Subject has permission "contracts.contract.create"

Enforce Object: "contracts/contract/*"  ← broad (matches all sub-resources)
Enforce Object: "contracts/contract/123"  ← specific (matches one ID)
```

Use `"contracts/contract/*"` for all general CRUD operations. For resource-level checks (e.g., "can this user access this specific contract?"), pass `"contracts/contract/<id>"`.

### Principal from Session

The handler extracts the session and derives the principal before calling the service:

```go
// handlers/handler.go
sess, ok := c.Locals(iam.LocalsKeySession).(*iam.ResolvedSession)
if !ok || sess == nil {
	return fiber.ErrUnauthorized
}

principal := sess.ToPrincipal()
// principal.Subject = "tenant:userID"
// principal.Domain  = tenantID.String()
```

The `ResolvedSession` has no `Can()` method. Always use `authzSvc.Enforce()`. The session is used to get the principal; the principal is used for authorization.

## Permission String → Object/Action Mapping

| Permission string | Object in Enforce | Action |
|------------------|--------------------|--------|
| `contracts.contract.create` | `"contracts/contract/*"` | `"create"` |
| `contracts.contract.read` | `"contracts/contract/*"` | `"read"` |
| `contracts.contract.update` | `"contracts/contract/*"` | `"update"` |
| `contracts.contract.delete` | `"contracts/contract/*"` | `"delete"` |
| `contracts.contract.submit` | `"contracts/contract/*"` | `"submit"` |
| `contracts.contract.approve` | `"contracts/contract/*"` | `"approve"` |
| `contracts.contract.activate` | `"contracts/contract/*"` | `"activate"` |
| `contracts.contract.terminate` | `"contracts/contract/*"` | `"terminate"` |
| `contracts.contract_line.create` | `"contracts/contract_line/*"` | `"create"` |

## Handling authzSvc Errors

```go
allowed, err := s.authzSvc.Enforce(ctx, req)
if err != nil {
	// Enforcement infrastructure error — not a permission denial
	// Log and return; do NOT treat as allowed
	s.logger.Error().Err(err).Msg("authz enforce failed")
	return nil, err  // propagates as 500 to the handler
}
if !allowed {
	return nil, iam.ErrForbidden  // handler maps to 403
}
```

An error from `authzSvc.Enforce` is an infrastructure failure (Casbin unreachable, policy load error), not a permission denial. Do not allow the operation on error — fail closed.

## Avoiding Double Authorization

Route-level middleware and service-level `Enforce` check the same permission. This is intentional — defense in depth. The middleware stops unauthenticated/unauthorized requests at the edge. The service-level check ensures correctness even if the service is called outside the HTTP path (e.g., from a Temporal activity).

Do not remove either check on the grounds that "the other one handles it".
