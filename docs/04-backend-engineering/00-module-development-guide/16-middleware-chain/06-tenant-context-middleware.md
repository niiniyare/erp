---
title: Tenant Context Middleware
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[Auth Middleware](02-auth-middleware.md)"
  - "[Middleware Reference](09-middleware-reference.md)"
  - "[Tenancy Model](../../../03-platform-architecture/01-multi-tenancy/01-tenancy-model.md)"
---

# Tenant Context Middleware

## Purpose

The `TenantContext` middleware:

1. Extracts `tenant_id` from the resolved session (set by `Authenticate`)
2. Validates the tenant is active (not suspended or archived)
3. Sets the tenant ID in the request context for downstream use

## Implementation

```go
// internal/platform/middleware/tenant_context.go
package middleware

import (
    "github.com/gofiber/fiber/v2"

    "awo.so/internal/core/iam/domain"
    "awo.so/internal/core/tenant"
)

// TenantContext validates the tenant from the session is active.
// Must run AFTER Authenticate.
func TenantContext(tenantSvc tenant.TenantService) fiber.Handler {
    return func(c *fiber.Ctx) error {
        session := SessionFrom(c)

        t, err := tenantSvc.GetByID(c.Context(), session.TenantID)
        if err != nil {
            return fiber.NewError(fiber.StatusUnauthorized, "tenant not found")
        }

        if !t.IsActive() {
            return fiber.NewError(fiber.StatusForbidden, "tenant account is not active")
        }

        // Store tenant in locals for handlers that need it
        c.Locals(domain.LocalsKeyTenant, t)
        return c.Next()
    }
}
```

## When TenantContext is Required

Apply `TenantContext` to all routes that access tenant data. Routes that don't require a tenant (health checks, public endpoints) must NOT have `TenantContext`:

```go
// Public routes — no auth, no tenant context
app.Get("/health/live",  handlers.HealthLive)
app.Get("/health/ready", handlers.HealthReady)

// Protected routes — auth + tenant context required
api := app.Group("/api/v1",
    middleware.Authenticate(sessionSvc),
    middleware.TenantContext(tenantSvc),
)

api.Get("/contracts", middleware.Authorize(authzSvc, "contracts.contract.read"), handlers.List)
```

## Accessing Tenant in Handler

```go
func (h *contractHandler) Create(c *fiber.Ctx) error {
    session := middleware.SessionFrom(c)

    // TenantID always comes from session — never from URL or body
    tenantID := session.TenantID

    // EntityID comes from session scope (EntityScopeEntity) or request body
    entityID := session.EntityID()

    // ...
}
```

## What TenantContext Must Never Do

- Parse tenant ID from URL parameters — always from session
- Allow cross-tenant operations (one session accessing another tenant's data)
- Cache tenant status for longer than the request lifecycle
