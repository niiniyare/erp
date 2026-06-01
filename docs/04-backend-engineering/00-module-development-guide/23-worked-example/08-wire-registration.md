---
title: Worked Example — Wire Registration
portal: 4 — Backend Engineering
section: 00-module-development-guide/23-worked-example
audience: [backend-engineer, tech-lead]
related:
  - path: ../20-wire-registration/02-contracts-wire.md
    title: Contracts Wire Setup
---

# Worked Example — Wire Registration

## contracts.go

```go
// internal/core/contracts/contracts.go
package contracts

import (
    "github.com/google/wire"

    "awo.so/internal/core/contracts/activities"
    "awo.so/internal/core/contracts/handler"
    "awo.so/internal/core/contracts/repository"
    "awo.so/internal/core/contracts/service"
    "awo.so/internal/core/contracts/worker"
)

// ProviderSet provides all contracts module dependencies.
var ProviderSet = wire.NewSet(
    repository.NewContractRepository,
    repository.NewContractLineRepository,
    service.NewContractService,
    handler.NewContractHandler,
    handler.NewContractRouteRegistrar,

    // Bind interfaces
    wire.Bind(new(service.ContractService), new(*service.contractService)),
    wire.Bind(new(repository.ContractRepository), new(*repository.contractRepository)),
)

// WorkerSet provides Temporal worker dependencies.
var WorkerSet = wire.NewSet(
    activities.NewContractActivities,
    worker.NewContractsWorker,
)
```

## Wire Injector Addition

```go
// cmd/server/wire.go — add contracts to InitializeApp
func InitializeApp(cfg Config) (*server.App, error) {
    wire.Build(
        db.ProviderSet,
        temporal.ProviderSet,

        iam.ProviderSet,
        tenant.ProviderSet,
        notifications.ProviderSet,
        audit.ProviderSet,
        eventbus.ProviderSet,

        contracts.ProviderSet,  // ← add this
        contracts.WorkerSet,    // ← and this

        server.ProviderSet,
    )
    return nil, nil
}
```

## RouteRegistrar in ProviderSet

The `ContractRouteRegistrar` must be contributed to the `[]RouteRegistrar` slice value:

```go
// internal/core/contracts/contracts.go (addition)
var RouteRegistrars = wire.NewSet(
    ProviderSet,
    wire.Bind(new(server.RouteRegistrar), new(*handler.ContractRouteRegistrar)),
)
```

Or use a value group:
```go
// In contracts.go
func ProvideRouteRegistrar(r *handler.ContractRouteRegistrar) server.RouteRegistrar {
    return r
}

var ProviderSet = wire.NewSet(
    // ... existing providers ...
    ProvideRouteRegistrar,
)
```

## After Wire Changes

```bash
# User runs:
make wire

# Verify wire_gen.go updated with no errors
# Then verify compilation:
go build ./...
```

**Gate**: Wire builds successfully. `InitializeApp` wires without provider errors.
