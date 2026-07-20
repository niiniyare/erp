> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Wire Overview
portal: 4 — Backend Engineering
section: 00-module-development-guide/20-wire-registration
audience: [backend-engineer, tech-lead]
related:
  - path: ./02-module-facade.md
    title: Module Facade
  - path: ./03-provider-sets.md
    title: Provider Sets
---

# Wire Overview

Google Wire generates compile-time dependency injection code. No reflection, no runtime container, no magic. After adding or changing providers, run `make wire` to regenerate `wire_gen.go`.

## How Wire Works

Wire reads provider functions (constructors) and builds a dependency graph. You declare what you need; Wire figures out the construction order.

```go
// Provider: a constructor function
func NewContractRepository(store db.Store) repository.ContractRepository {
	return &contractSQLCRepository{store: store}
}

// Provider set: group of providers for a module
var ContractsRepositorySet = wire.NewSet(
	NewContractRepository,
)

// Wire injection target: the whole application
func InitializeApp(cfg *config.Config) (*App, error) {
	wire.Build(
		ContractsRepositorySet,
		ContractsServiceSet,
		// ... all other sets
	)
	return nil, nil  // Wire replaces this body
}
```

Wire reads the `InitializeApp` function, traces all dependencies, and generates the real implementation in `wire_gen.go`.

## Key Rules

**Never edit `wire_gen.go` by hand.** It is regenerated on every `make wire` run. Hand edits are overwritten.

**Run `make wire` after every provider change.** If you add a new dependency to a constructor, Wire will fail to compile until you update the provider sets and regenerate.

**Constructor parameter order matters.** Wire matches parameters by type. If two parameters have the same type, Wire cannot distinguish them — use named wrapper types.

**Wire runs at compile time.** There is no runtime "not found" error for missing dependencies. Missing providers cause a compilation error.

## Wire File Structure

```
internal/app/
├── wire.go          # Wire provider declarations (hand-written)
├── wire_gen.go      # Wire-generated implementation (never edit)
└── app.go           # Application struct and Fiber app construction
```
