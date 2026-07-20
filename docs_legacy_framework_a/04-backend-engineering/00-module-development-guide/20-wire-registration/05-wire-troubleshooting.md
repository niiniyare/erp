> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Wire Troubleshooting
portal: 4 — Backend Engineering
section: 00-module-development-guide/20-wire-registration
audience: [backend-engineer, tech-lead]
related:
  - path: ./04-application-wire.md
    title: Application Wire File
---

# Wire Troubleshooting

Common Wire errors and their fixes.

## Error: Missing Provider

```
wire: contracts/service.NewContractService needs type repository.ContractRepository,
      but no provider was found
```

**Fix**: Add `contracts.ContractsRepositorySet` to `wire.Build`. The repository set provides `ContractRepository`.

## Error: Duplicate Provider

```
wire: contracts/repository.NewContractRepository and contracts/repository.NewContractRepositoryV2
      both provide repository.ContractRepository
```

**Fix**: Remove the duplicate. Only one provider per type is allowed.

## Error: Unused Provider

```
wire: provider repository.NewContractLineRepository is not used
```

**Fix**: The line repository is not consumed by any constructor in the graph. Check that `service.NewContractService` takes `ContractLineRepository` as a parameter. If the service doesn't need it, remove it from the repository set.

## Error: Cycle Detected

```
wire: dependency cycle: ContractService → AuditService → ContractService
```

**Fix**: A circular dependency exists. One of the services must be refactored. Common pattern: use an event bus instead of direct service-to-service calls to break the cycle. ContractService publishes an event; AuditService subscribes to it.

## Error: Multiple Providers for Same Type

If two different provider sets provide the same interface (e.g., both `iam.IAMSet` and a test set provide `AuthzService`), Wire will error.

**Fix**: Use Wire's build tags to separate production and test providers:

```go
//go:build !test

package app

// Production wire.go
```

```go
//go:build test

package app

// Test wire.go with mock providers
```

## Error: Pointer vs Interface Mismatch

```
wire: need type service.ContractService, have *service.contractService
```

**Fix**: The constructor must return the interface type, not the concrete type:

```go
// CORRECT
func NewContractService(...) service.ContractService {
	return &contractService{...}
}

// WRONG — returns concrete type
func NewContractService(...) *contractService {
	return &contractService{...}
}
```

## Verifying wire_gen.go

After `make wire`, check that `wire_gen.go` contains:

1. A call to `repository.NewContractRepository` with `store` as the argument.
2. A call to `service.NewContractService` with all dependencies.
3. A call to `contractsHandler.NewHandler` with the service and other deps.

If any are missing, the corresponding provider set was not added to `wire.Build`.
