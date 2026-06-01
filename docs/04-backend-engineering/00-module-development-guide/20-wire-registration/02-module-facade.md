---
title: Module Facade
portal: 4 — Backend Engineering
section: 00-module-development-guide/20-wire-registration
audience: [backend-engineer, tech-lead]
related:
  - path: ./01-wire-overview.md
    title: Wire Overview
  - path: ./03-provider-sets.md
    title: Provider Sets
---

# Module Facade

`internal/core/contracts/contracts.go` is the module facade. It declares Wire provider sets and re-exports public types. External packages import only this facade — never internal sub-packages.

## Complete contracts.go

```go
// internal/core/contracts/contracts.go
package contracts

import (
	"github.com/google/wire"

	"awo.so/internal/core/contracts/repository"
	"awo.so/internal/core/contracts/service"
)

// ============================================================
// Wire Provider Sets
// ============================================================

// ContractsRepositorySet provides ContractRepository and ContractLineRepository.
var ContractsRepositorySet = wire.NewSet(
	repository.NewContractRepository,
	repository.NewContractLineRepository,
)

// ContractsServiceSet provides ContractService.
var ContractsServiceSet = wire.NewSet(
	service.NewContractService,
)

// ContractsSet is the full module provider set: repository + service.
// Import ContractsSet in the application wire provider.
var ContractsSet = wire.NewSet(
	ContractsRepositorySet,
	ContractsServiceSet,
)

// ============================================================
// Public Type Re-exports
// (external packages use contracts.ContractService, etc.)
// ============================================================

type (
	ContractService          = service.ContractService
	ContractRepository       = repository.ContractRepository
	ContractLineRepository   = repository.ContractLineRepository
)
```

## Why Re-export Types

Re-exports allow external packages to reference module types without importing sub-packages:

```go
// handler.go — uses the facade type alias
import contracts "awo.so/internal/core/contracts"

type Dependencies struct {
	ContractService contracts.ContractService  // clean
}

// Without re-exports, they'd need:
import svc "awo.so/internal/core/contracts/service"
type Dependencies struct {
	ContractService svc.ContractService  // exposes internal sub-package
}
```

This keeps the module's internal structure an implementation detail. If you rename `service` to `usecase`, only `contracts.go` changes — not every import site.

## Handler Wire Set

The handler is NOT in the module's `ContractsSet`. It is in the API layer's wire set:

```go
// internal/api/handlers/wire.go (or similar)
var HandlersSet = wire.NewSet(
	contracts.NewHandler,  // constructor from handlers/contracts/
	// ... other module handlers
)
```

The handler depends on `ContractService` (from `ContractsSet`) and platform dependencies like `AuthConfig`, `Logger`, `Tracer`.
