> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Middleware Chain Overview
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[Middleware Reference](09-middleware-reference.md)"
  - "[CORS Middleware](08-cors-middleware.md)"
  - "[Handler Layer](../07-handler-layer/01-handler-overview.md)"
  - "[Security Guide](../../../07-security/03-authorization-guide.md)"
---

# Middleware Chain Overview

Every request passes through a stack of middleware before reaching a handler. Middleware runs in registration order. Understanding the chain is critical for security — a misconfigured order can skip authentication.

## Global Middleware (all routes)

```
Request
  │
  ▼
RequestID        — assigns X-Request-ID
  │
  ▼
Logger           — structured request/response log
  │
  ▼
Recover          — catches handler panics, returns 500
  │
  ▼
RateLimit        — per-IP and per-tenant limits
  │
  ▼
CORS             — permissive for API, strict for admin
  │
  ▼
Compression      — gzip for responses > 1 KB
  │
  ▼
[route handler]
```

## Per-Route Middleware (protected routes)

```
[global chain]
  │
  ▼
Authenticate     — validates session token, sets session in locals
  │
  ▼
TenantContext    — derives tenant_id from session
  │
  ▼
FeatureFlag      — short-circuits if module disabled for tenant
  │
  ▼
Authorize(perm)  — checks Casbin permission string
  │
  ▼
[handler]
```

## Registration Pattern

```go
// Protected contract routes
contracts := app.Group("/api/v1/contracts",
    middlewarePkg.Authenticate(sessionSvc),
    middlewarePkg.TenantContext(),
)

// Read-only endpoint — requires list permission
contracts.Get("/",
    middlewarePkg.Authorize(authzSvc, "contracts.contract.read"),
    handlers.List,
)

// Write endpoint — requires create permission
contracts.Post("/",
    middlewarePkg.FeatureFlag("contracts.enabled"),
    middlewarePkg.Authorize(authzSvc, "contracts.contract.create"),
    handlers.Create,
)
```

## Order Rules

1. `Recover` must be first (or second after RequestID) — handles panics in all downstream middleware
2. `Authenticate` must run before `Authorize` — authz depends on session
3. `TenantContext` must run after `Authenticate` — derives tenant from session
4. `FeatureFlag` runs after `TenantContext` — needs tenant ID to check flags
5. `Authorize` runs last before the handler — all context must be resolved

## Middleware vs. Handler Responsibilities

| Concern | Middleware | Handler |
|---------|-----------|---------|
| Parse session token | Authenticate | — |
| Derive tenant ID | TenantContext | — |
| Check feature flag | FeatureFlag | — |
| Route-level permission check | Authorize | — |
| Service-level permission check | — | Service.Enforce |
| Parse request body | — | Handler |
| Call service layer | — | Handler |
| Build response | — | Handler |
