---
title: Wire Registration Overview
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer]
related:
  - "[Wire Overview](../../../03-platform-architecture/06-dependency-injection/01-wire-overview.md)"
  - "[Provider Patterns](../../../03-platform-architecture/06-dependency-injection/02-provider-patterns.md)"
  - "[Worked Example: Wire Registration](../23-worked-example/08-wire-registration.md)"
  - "[Startup Sequence](../21-server-startup/02-wire-app.md)"
---

# Wire Registration Overview

## Module wire.go

Every module has exactly one `wire.go` at the package root:

```go
// internal/core/contracts/wire.go
package contracts

import "github.com/google/wire"

// ContractsSet provides all HTTP-facing components
var ContractsSet = wire.NewSet(
    // Repository
    repository.NewContractRepo,
    wire.Bind(new(repository.ContractRepository), new(*repository.sqlcContractRepo)),

    // Service
    service.NewContractService,

    // Handler + routes
    handler.NewContractHandler,
    handler.NewContractRouteRegistrar,
)

// WorkerSet provides Temporal workers and activities
var WorkerSet = wire.NewSet(
    workflow.NewContractActivities,
    workflow.NewContractsWorker,
)
```

## Registering in the App Injector

After creating a new module's `ProviderSet`, add it to the server injector:

```go
// cmd/server/wire.go
//go:build wireinject

package main

import (
    "github.com/google/wire"
    "awo.so/internal/core/contracts"
    "awo.so/internal/core/finance"
    "awo.so/internal/core/iam"
    // ...
)

func InitializeApp(cfg Config) (*App, func(), error) {
    wire.Build(
        // Infrastructure
        db.DatabaseSet,
        redis.RedisSet,
        temporal.TemporalSet,
        events.EventsSet,
        audit.AuditSet,
        notifications.NotificationsSet,

        // IAM (must come before business modules — they depend on IAM interfaces)
        iam.IAMSet,

        // Business modules
        contracts.ContractsSet,
        contracts.WorkerSet,
        finance.FinanceSet,
        finance.WorkerSet,

        // Server
        NewFiberApp,
        NewRouteRegistry,
        NewApp,
    )
    return nil, nil, nil
}
```

After editing `wire.go` run:

```bash
make wire
```

This regenerates `cmd/server/wire_gen.go`. Commit both files together.

## Route Registrar Pattern

Each module implements a `RouteRegistrar` interface:

```go
// internal/server/routes.go
type RouteRegistrar interface {
    Register(api fiber.Router)
}
```

The module's route registrar registers all its routes:

```go
// internal/core/contracts/handler/routes.go
type ContractRouteRegistrar struct {
    handler *ContractHandler
    authSvc authz.Service
}

func NewContractRouteRegistrar(h *ContractHandler, authz authz.Service) *ContractRouteRegistrar {
    return &ContractRouteRegistrar{handler: h, authSvc: authz}
}

func (r *ContractRouteRegistrar) Register(api fiber.Router) {
    authenticate := middleware.Authenticate(sessionSvc)
    authorize := func(perm string) fiber.Handler {
        return middleware.Authorize(r.authSvc, perm)
    }

    g := api.Group("/contracts", authenticate)
    g.Get("",      authorize("contracts.contract.read"),   r.handler.List)
    g.Post("",     authorize("contracts.contract.create"), r.handler.Create)
    g.Get("/:id",  authorize("contracts.contract.read"),   r.handler.GetByID)
    g.Put("/:id",  authorize("contracts.contract.update"), r.handler.Update)
    g.Delete("/:id", authorize("contracts.contract.delete"), r.handler.Delete)

    g.Post("/:id/submit",    authorize("contracts.contract.submit"),    r.handler.Submit)
    g.Post("/:id/approve",   authorize("contracts.contract.approve"),   r.handler.Approve)
    g.Post("/:id/activate",  authorize("contracts.contract.activate"),  r.handler.Activate)
    g.Post("/:id/terminate", authorize("contracts.contract.terminate"), r.handler.Terminate)

    g.Post("/import", authorize("contracts.contract.create"), r.handler.Import)
    g.Get("/export",  authorize("contracts.contract.read"),   r.handler.Export)
}
```

The `NewRouteRegistry` in the server collects all registrars and calls `Register(api)` on each:

```go
// cmd/server/routes.go
type RouteRegistry struct {
    registrars []RouteRegistrar
}

func NewRouteRegistry(
    contracts *handler.ContractRouteRegistrar,
    finance   *financeHandler.FinanceRouteRegistrar,
    iam       *iamHandler.IAMRouteRegistrar,
    // ...
) *RouteRegistry {
    return &RouteRegistry{
        registrars: []RouteRegistrar{contracts, finance, iam},
    }
}

func (r *RouteRegistry) RegisterAll(api fiber.Router) {
    for _, reg := range r.registrars {
        reg.Register(api)
    }
}
```

## Checklist: Adding a New Module to Wire

1. Create `internal/core/{module}/wire.go` with `{Module}Set` and `WorkerSet`
2. Add provider bindings for all interfaces
3. Import the module package in `cmd/server/wire.go`
4. Add `{module}.{Module}Set` and `{module}.WorkerSet` to `wire.Build()`
5. Add the module's `RouteRegistrar` to `NewRouteRegistry` params
6. Run `make wire`
7. Verify `wire_gen.go` compiles (tell user to run `go build ./...`)
8. Add the module's task queue name to Temporal worker registration
