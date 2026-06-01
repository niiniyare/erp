---
title: Route Registration
portal: 4 — Backend Engineering
section: 00-module-development-guide/15-fiber-handlers
audience: [backend-engineer, tech-lead]
related:
  - path: ./02-handler-struct.md
    title: Handler Struct
  - path: ../16-middleware-chain/01-middleware-overview.md
    title: Middleware Chain Overview
---

# Route Registration

`routes.go` declares the `RegisterRoutes` function. It wires handler methods to HTTP paths and applies authentication and authorization middleware.

## routes.go

```go
// internal/api/handlers/contracts/routes.go
package contracts

import (
	"github.com/gofiber/fiber/v2"

	"awo.so/internal/api/handlers"
	middlewarePkg "awo.so/internal/api/middleware"
)

// RegisterRoutes registers all contract routes on the given API router group.
// apiGroup is typically the /api/v1 group from the application router.
func RegisterRoutes(apiGroup fiber.Router, deps *handlers.Dependencies) {
	h := NewHandler(
		deps.ContractService,
		deps.AuthConfig,
		deps.Logger,
		deps.Tracer,
	)

	auth := middlewarePkg.Authenticate(*deps.AuthConfig)
	authz := func(permission string) fiber.Handler {
		return middlewarePkg.Authorize(*deps.AuthConfig, permission)
	}

	contracts := apiGroup.Group("/contracts")
	contracts.Use(auth)  // all contract routes require authentication

	// Collection routes
	contracts.Get("",    authz("contracts.contract.read"),   h.List)
	contracts.Post("",   authz("contracts.contract.create"), h.Create)

	// Resource routes
	contracts.Get("/:id",    authz("contracts.contract.read"),   h.GetByID)
	contracts.Put("/:id",    authz("contracts.contract.update"), h.Update)
	contracts.Delete("/:id", authz("contracts.contract.delete"), h.Delete)

	// Status transition routes (POST to sub-resource)
	contracts.Post("/:id/submit",    authz("contracts.contract.submit"),    h.Submit)
	contracts.Post("/:id/approve",   authz("contracts.contract.approve"),   h.Approve)
	contracts.Post("/:id/activate",  authz("contracts.contract.activate"),  h.Activate)
	contracts.Post("/:id/suspend",   authz("contracts.contract.suspend"),   h.Suspend)
	contracts.Post("/:id/terminate", authz("contracts.contract.terminate"), h.Terminate)

	// Audit history
	contracts.Get("/:id/history", authz("contracts.contract.read"), h.GetHistory)

	// Contract lines sub-resource
	lines := contracts.Group("/:id/lines")
	lines.Get("",        authz("contracts.contract_line.read"),   h.ListLines)
	lines.Post("",       authz("contracts.contract_line.create"), h.AddLine)
	lines.Put("/:lineId",    authz("contracts.contract_line.update"), h.UpdateLine)
	lines.Delete("/:lineId", authz("contracts.contract_line.delete"), h.RemoveLine)
}
```

## Route Registration in Application Router

`RegisterRoutes` is called from the application-level `routes.go`:

```go
// internal/api/handlers/routes.go
func (r *Router) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1")

	// ... other modules ...

	contracts.RegisterRoutes(api, r.deps)
}
```

## Route Design Rules

**Collection before resource.** `GET /contracts` before `GET /contracts/:id`. Fiber's router matches in registration order.

**Status transitions as POST to sub-resource.** `POST /contracts/:id/submit` — not `PATCH /contracts/:id` with a `{"status": "submitted"}` body. Each transition is a distinct action with its own permission and semantic meaning.

**Singular nouns for resource paths.** `/contracts/:id/lines` — not `/contracts/:id/line`.

**No nested IDs that aren't needed.** `DELETE /contracts/:id/lines/:lineId` — the `:id` is in the path even though the line repository could look up by lineId alone. It's here for semantic clarity and to satisfy the middleware's tenant + auth checks at the contract level.

**Auth middleware applied at group level.** `contracts.Use(auth)` applies authentication to all routes in the group. Individual `authz(...)` middleware applies authorization per route. This pattern means authentication is never forgotten for a new route.

## Route Table

| Method | Path | Permission | Handler |
|--------|------|-----------|---------|
| GET | `/api/v1/contracts` | `contracts.contract.read` | `List` |
| POST | `/api/v1/contracts` | `contracts.contract.create` | `Create` |
| GET | `/api/v1/contracts/:id` | `contracts.contract.read` | `GetByID` |
| PUT | `/api/v1/contracts/:id` | `contracts.contract.update` | `Update` |
| DELETE | `/api/v1/contracts/:id` | `contracts.contract.delete` | `Delete` |
| POST | `/api/v1/contracts/:id/submit` | `contracts.contract.submit` | `Submit` |
| POST | `/api/v1/contracts/:id/approve` | `contracts.contract.approve` | `Approve` |
| POST | `/api/v1/contracts/:id/activate` | `contracts.contract.activate` | `Activate` |
| POST | `/api/v1/contracts/:id/suspend` | `contracts.contract.suspend` | `Suspend` |
| POST | `/api/v1/contracts/:id/terminate` | `contracts.contract.terminate` | `Terminate` |
| GET | `/api/v1/contracts/:id/history` | `contracts.contract.read` | `GetHistory` |
| GET | `/api/v1/contracts/:id/lines` | `contracts.contract_line.read` | `ListLines` |
| POST | `/api/v1/contracts/:id/lines` | `contracts.contract_line.create` | `AddLine` |
| PUT | `/api/v1/contracts/:id/lines/:lineId` | `contracts.contract_line.update` | `UpdateLine` |
| DELETE | `/api/v1/contracts/:id/lines/:lineId` | `contracts.contract_line.delete` | `RemoveLine` |
