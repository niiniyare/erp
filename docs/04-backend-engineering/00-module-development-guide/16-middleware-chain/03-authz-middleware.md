---
title: Authorization Middleware
portal: 4 — Backend Engineering
section: 00-module-development-guide/16-middleware-chain
audience: [backend-engineer, tech-lead]
related:
  - path: ./02-auth-middleware.md
    title: Authentication Middleware
  - path: ../06-service-layer/04-authorization.md
    title: Service-Level Authorization
---

# Authorization Middleware

## Route-Level Authorization

The `Authorize` middleware performs a coarse-grained permission check at the route level. It answers: "does this user have any permission for this action on this resource type?"

Service-level authorization (section §6.4) then performs a fine-grained check on the specific resource instance.

```go
// internal/platform/middleware/authorize.go
package middleware

import (
    "github.com/gofiber/fiber/v2"
    "awo.so/internal/core/iam"
    "awo.so/internal/core/iam/domain"
)

// Authorize returns a middleware that checks if the session has the given permission.
// permission format: "module.resource.action"  e.g. "contracts.contract.create"
func Authorize(authzSvc iam.AuthzService, permission string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        session := SessionFrom(c)

        allowed, err := authzSvc.Enforce(c.Context(), iam.Request{
            Subject: session.ToPrincipal().Subject,
            Domain:  session.ToPrincipal().Domain,
            Object:  permission,
            Action:  "allow",
        })
        if err != nil {
            // Authz service error — fail closed (deny)
            return fiber.NewError(fiber.StatusForbidden, "authorization check failed")
        }
        if !allowed {
            return fiber.NewError(fiber.StatusForbidden, "insufficient permissions")
        }

        return c.Next()
    }
}
```

## Permission String Convention

```
{module}.{resource}.{action}
```

| Part | Example | Description |
|------|---------|-------------|
| module | `contracts` | Module identifier |
| resource | `contract` | Singular resource name |
| action | `create`, `read`, `update`, `delete`, `submit`, `approve` | CRUD or business action |

Full examples:
```
contracts.contract.create
contracts.contract.read
contracts.contract.update
contracts.contract.delete
contracts.contract.submit
contracts.contract.approve
contracts.contract.terminate
contracts.line.create
contracts.line.delete
```

## Casbin Policy Example

```
p, role:contracts_viewer,   tenant:acme, contracts.contract.read,   allow
p, role:contracts_editor,   tenant:acme, contracts.contract.create,  allow
p, role:contracts_editor,   tenant:acme, contracts.contract.update,  allow
p, role:contracts_reviewer, tenant:acme, contracts.contract.approve, allow
p, role:contracts_admin,    tenant:acme, contracts.*,                allow
```

## Two-Level Authorization

Route-level (middleware) vs. service-level authorization serve different purposes:

```
Request arrives
    │
    ▼
Authorize("contracts.contract.read")   ← route-level: "can user read ANY contract?"
    │
    ▼
Handler calls svc.GetByID(contractID)
    │
    ▼
svc.authzSvc.Enforce(Object: "tenants/X/contracts/contractID")  ← service-level: "can user read THIS contract?"
```

Route-level denials return 403 before the handler runs. Service-level denials propagate as `iam.ErrForbidden` → `mapError` → 403.

## FeatureFlag Middleware

```go
// Authorize runs AFTER FeatureFlag
func FeatureFlag(flag string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        session := SessionFrom(c)
        if !session.FeatureEnabled(flag) {
            return fiber.NewError(fiber.StatusForbidden, "feature not available")
        }
        return c.Next()
    }
}
```

Usage:
```go
contracts.Post("/",
    middlewarePkg.FeatureFlag("contracts.enabled"),   // check flag first
    middlewarePkg.Authorize(authzSvc, "contracts.contract.create"),
    handlers.Create,
)
```
