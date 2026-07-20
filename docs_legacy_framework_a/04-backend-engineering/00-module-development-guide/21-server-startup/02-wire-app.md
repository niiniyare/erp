> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Wire App Initialization
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[Startup Overview](01-startup-overview.md)"
  - "[Wire Registration](../09-wire-registration/01-wire-registration-overview.md)"
---

# Wire App Initialization

## App Struct

The `App` struct holds all initialized dependencies that the server needs post-startup:

```go
// internal/server/app.go
package server

import (
    "github.com/gofiber/fiber/v2"
    "go.temporal.io/sdk/worker"

    "awo.so/internal/platform/eventbus"
)

// App holds all initialized application dependencies.
type App struct {
    Fiber           *fiber.App
    TemporalWorkers []worker.Worker
    EventBus        eventbus.Subscriber
    Subscribers     []EventSubscriber
    DBPool          *pgxpool.Pool
    TracerProvider  *sdktrace.TracerProvider
}

// EventSubscriber is any type that can register subscriptions on the event bus.
type EventSubscriber interface {
    Register(bus eventbus.Subscriber) error
}
```

## Wire Injector

```go
// cmd/server/wire.go
//go:build wireinject

package main

import (
    "github.com/google/wire"
    "awo.so/internal/server"
    "awo.so/internal/platform/db"
    "awo.so/internal/platform/temporal"
    "awo.so/internal/core/contracts"
    "awo.so/internal/core/finance"
    "awo.so/internal/core/iam"
    "awo.so/internal/core/tenant"
    "awo.so/internal/core/notifications"
)

func InitializeApp(cfg Config) (*server.App, error) {
    wire.Build(
        // Platform
        db.ProviderSet,
        temporal.ProviderSet,

        // Core modules
        iam.ProviderSet,
        tenant.ProviderSet,
        notifications.ProviderSet,
        contracts.ProviderSet,
        contracts.WorkerSet,
        finance.ProviderSet,

        // Server setup
        server.ProviderSet,
    )
    return nil, nil
}
```

## Contracts ProviderSet

```go
// internal/core/contracts/contracts.go
package contracts

import (
    "github.com/google/wire"

    "awo.so/internal/core/contracts/handler"
    "awo.so/internal/core/contracts/repository"
    "awo.so/internal/core/contracts/service"
)

var ProviderSet = wire.NewSet(
    repository.NewContractRepository,
    repository.NewContractLineRepository,
    service.NewContractService,
    handler.NewContractHandler,
    handler.NewContractRouteRegistrar,
)

var WorkerSet = wire.NewSet(
    activities.NewContractActivities,
    worker.NewContractsWorker,
)
```

## Route Registrar Pattern

Each module provides a `RouteRegistrar` that wires its handlers to the Fiber router:

```go
// internal/core/contracts/handler/routes.go
package handler

import (
    "github.com/gofiber/fiber/v2"
    middlewarePkg "awo.so/internal/platform/middleware"
    "awo.so/internal/core/iam"
)

type ContractRouteRegistrar struct {
    handler  *contractHandler
    authzSvc iam.AuthzService
    sessionSvc iam.SessionService
    tenantSvc  tenant.TenantService
}

func NewContractRouteRegistrar(
    h *contractHandler,
    authzSvc iam.AuthzService,
    sessionSvc iam.SessionService,
    tenantSvc tenant.TenantService,
) *ContractRouteRegistrar {
    return &ContractRouteRegistrar{h, authzSvc, sessionSvc, tenantSvc}
}

func (r *ContractRouteRegistrar) Register(app *fiber.App) {
    auth := []fiber.Handler{
        middlewarePkg.Authenticate(r.sessionSvc),
        middlewarePkg.TenantContext(r.tenantSvc),
    }

    contracts := app.Group("/api/v1/contracts", auth...)

    contracts.Get("/",    middlewarePkg.Authorize(r.authzSvc, "contracts.contract.read"),   r.handler.List)
    contracts.Post("/",   middlewarePkg.Authorize(r.authzSvc, "contracts.contract.create"), r.handler.Create)
    contracts.Get("/:id", middlewarePkg.Authorize(r.authzSvc, "contracts.contract.read"),   r.handler.GetByID)
    contracts.Put("/:id", middlewarePkg.Authorize(r.authzSvc, "contracts.contract.update"), r.handler.Update)
    contracts.Delete("/:id", middlewarePkg.Authorize(r.authzSvc, "contracts.contract.delete"), r.handler.Delete)

    contracts.Post("/:id/submit",    middlewarePkg.Authorize(r.authzSvc, "contracts.contract.submit"),    r.handler.Submit)
    contracts.Post("/:id/approve",   middlewarePkg.Authorize(r.authzSvc, "contracts.contract.approve"),   r.handler.Approve)
    contracts.Post("/:id/reject",    middlewarePkg.Authorize(r.authzSvc, "contracts.contract.approve"),   r.handler.Reject)
    contracts.Post("/:id/activate",  middlewarePkg.Authorize(r.authzSvc, "contracts.contract.activate"),  r.handler.Activate)
    contracts.Post("/:id/terminate", middlewarePkg.Authorize(r.authzSvc, "contracts.contract.terminate"), r.handler.Terminate)

    contracts.Post("/import", middlewarePkg.Authorize(r.authzSvc, "contracts.contract.create"), r.handler.Import)
    contracts.Get("/export",  middlewarePkg.Authorize(r.authzSvc, "contracts.contract.read"),   r.handler.Export)
}
```

## Server Setup

```go
// internal/server/setup.go
package server

type RouteRegistrar interface {
    Register(app *fiber.App)
}

func NewFiberApp(registrars []RouteRegistrar, log zerolog.Logger) *fiber.App {
    app := fiber.New(fiber.Config{
        ErrorHandler: customErrorHandler(log),
    })

    // Global middleware
    app.Use(middleware.Recover(log))
    app.Use(middleware.RequestID())
    app.Use(middleware.Logger(log))

    // Register module routes
    for _, r := range registrars {
        r.Register(app)
    }

    return app
}
```
