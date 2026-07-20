> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: IAM Integration
portal: 4 — Backend Engineering
section: 00-module-development-guide/07-core-integrations
audience: [backend-engineer, tech-lead]
related:
  - path: ./01-tenant-integration.md
    title: Tenant Integration
  - path: ../06-service-layer/04-authorization.md
    title: Authorization in the Service Layer
---

# IAM Integration

The IAM module provides authentication (sessions) and authorization (Casbin-backed Enforce). This page explains how to integrate both in a business module.

## Session Extraction

Sessions are set by the authentication middleware and available in every handler via Fiber Locals:

```go
// Key constant from IAM domain package
// iam.LocalsKeySession = "resolved_session"

sess, ok := c.Locals(iam.LocalsKeySession).(*iam.ResolvedSession)
if !ok || sess == nil {
	return fiber.ErrUnauthorized
}
```

The session is pre-populated at login with:
- `TenantID` — tenant scope
- `UserID` — actor identity
- `EntityScope` — organisational scope
- `Configuration.Flags` — feature flag snapshot
- `Configuration.Settings` — settings snapshot
- `Configuration.Prefs` — user preferences snapshot

None of these are fetched from the database during the request — they are all read from the session object (which may be in a fast cache).

## ResolvedSession API

```go
// Identity
sess.TenantID     uuid.UUID
sess.UserID       uuid.UUID
sess.DisplayName  string

// Org hierarchy
sess.EntityScope  iam.EntityScope  // .Type, .EntityID, .PathPrefix

// Feature flags (snapshotted at login)
if sess.FeatureEnabled("contracts.multi_currency") {
	// ...
}

// Settings (snapshotted at login)
threshold := sess.SettingDecimal("contracts.approval_threshold", 50000.0)
currency  := sess.SettingString("contracts.default_currency", "USD")

// Principal for authz calls
principal := sess.ToPrincipal()
// principal is iam.Principal{Subject: "tenant:userID", Domain: "tenantID"}
```

## What Session Does NOT Have

The `ResolvedSession` does **not** have:
- `Can(permission string) bool` — does not exist
- `HasRole(role string) bool` — does not exist
- `IsAdmin() bool` — does not exist

All authorization decisions go through `authzSvc.Enforce()`. The session provides the principal identity; the authz service evaluates the policy.

## Route-Level Authorization (Middleware)

Applied in `routes.go`, before the handler:

```go
// routes.go
contracts := api.Group("/contracts")
contracts.Use(deps.AuthenticateMiddleware(), deps.AuthorizeMiddleware("contracts.contract.read"))
contracts.Get("", h.List)

contracts.Post("", deps.AuthorizeMiddleware("contracts.contract.create"), h.Create)
```

`AuthorizeMiddleware("contracts.contract.create")` calls `middlewarePkg.Authorize(*deps.AuthConfig, "contracts.contract.create")` which internally calls `authzSvc.Enforce` with the session's principal.

## Service-Level Authorization

Applied inside each service method as the first operation:

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

## AuthConfig in Routes

`deps.AuthConfig` is a `*middlewarePkg.AuthConfig` injected via Wire. It holds the `AuthzService` reference used by the `Authorize` middleware. The handler struct stores a reference to it:

```go
type Handler struct {
	service    service.ContractService
	authConfig *middlewarePkg.AuthConfig
}
```

The `AuthorizeMiddleware` helper wraps it:

```go
func (h *Handler) AuthorizeMiddleware(permission string) fiber.Handler {
	return middlewarePkg.Authorize(*h.authConfig, permission)
}
```

## Error Responses

| IAM error | HTTP response |
|-----------|--------------|
| `iam.ErrUnauthorized` | `401 Unauthorized` |
| `iam.ErrForbidden` | `403 Forbidden` |
| `authzSvc.Enforce` returns error | `500 Internal Server Error` |

The handler's `mapError` function must handle all three cases. Do not return `403` when `authzSvc.Enforce` returns an error — that signals an infrastructure failure, not a permission denial.
