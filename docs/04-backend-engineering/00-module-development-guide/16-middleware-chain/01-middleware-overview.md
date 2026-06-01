---
title: Middleware Chain Overview
portal: 4 — Backend Engineering
section: 00-module-development-guide/16-middleware-chain
audience: [backend-engineer, tech-lead]
related:
  - path: ./02-auth-middleware.md
    title: Auth Middleware
  - path: ./03-authz-middleware.md
    title: Authz Middleware
  - path: ../15-fiber-handlers/05-route-registration.md
    title: Route Registration
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
