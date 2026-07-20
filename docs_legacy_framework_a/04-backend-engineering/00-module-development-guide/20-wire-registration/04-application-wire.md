> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Application Wire File
portal: 4 — Backend Engineering
section: 00-module-development-guide/20-wire-registration
audience: [backend-engineer, tech-lead]
related:
  - path: ./03-provider-sets.md
    title: Provider Sets
  - path: ./05-wire-troubleshooting.md
    title: Wire Troubleshooting
---

# Application Wire File

`internal/app/wire.go` is the central Wire declaration file. It lists all provider sets and declares the `InitializeApp` injection target.

## Adding the Contracts Module

When adding the contracts module, add `contracts.ContractsSet` to `wire.Build`:

```go
// internal/app/wire.go
//go:build wireinject

package app

import (
	"github.com/google/wire"

	// Platform providers
	"awo.so/internal/platform/cache"
	"awo.so/internal/platform/audit"
	"awo.so/internal/platform/events"
	"awo.so/internal/platform/notifications"

	// Core module providers
	"awo.so/internal/core/iam"
	"awo.so/internal/core/tenant"
	contracts "awo.so/internal/core/contracts"      // ← new module

	// API handlers
	contractsHandler "awo.so/internal/api/handlers/contracts"  // ← new handler
	handlersRoot "awo.so/internal/api/handlers"

	// Infrastructure
	db "awo.so/db/sqlc"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

// InitializeApp constructs the fully-wired application.
func InitializeApp(cfg *Config) (*App, error) {
	wire.Build(
		// Infrastructure
		ProvideStore,      // db.Store from cfg.DatabaseURL
		ProvideCache,      // cache.Service
		ProvideLogger,     // logger.Logger
		ProvideTracer,     // tracing.Service
		ProvideMetrics,    // metrics.MetricsProvider

		// Platform services
		audit.ProvideService,
		events.ProvideBus,
		notifications.ProvideService,

		// IAM
		iam.IAMSet,        // AuthzService, SessionService, UserService

		// Tenant
		tenant.TenantSet,

		// ---- NEW: Contracts module ----
		contracts.ContractsSet,
		contractsHandler.ProvideHandler,
		// --------------------------------

		// App struct provider
		NewApp,
	)
	return nil, nil
}
```

## Handler Registration in App

After Wire wires up the handler, the application registers its routes:

```go
// internal/app/app.go
type App struct {
	fiber           *fiber.App
	contractHandler *contractsHandler.Handler
	// ... other handlers
}

func (a *App) RegisterRoutes() {
	api := a.fiber.Group("/api/v1")

	// ... other modules ...

	contractsHandler.RegisterRoutes(api, a.deps)
}
```

## Handler Provider Function

The handler package needs a provider function for Wire:

```go
// internal/api/handlers/contracts/wire.go
package contracts

import (
	"github.com/google/wire"

	middlewarePkg "awo.so/internal/api/middleware"
	contractsSvc "awo.so/internal/core/contracts"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/tracing"
)

// ProvideHandler is the Wire provider for the contracts Handler.
func ProvideHandler(
	svc  contractsSvc.ContractService,
	auth *middlewarePkg.AuthConfig,
	log  logger.Logger,
	tr   tracing.Service,
) *Handler {
	return NewHandler(svc, auth, log, tr)
}

// HandlerSet is the Wire set for this handler.
var HandlerSet = wire.NewSet(ProvideHandler)
```

## After Adding the Module

```bash
make wire    # regenerate wire_gen.go
make build   # verify compilation
make test    # run tests
```

Check `wire_gen.go` to verify the contracts module's providers appear in the generated `initializeApp` function.
