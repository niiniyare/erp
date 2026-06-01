---
title: Provider Sets
portal: 4 — Backend Engineering
section: 00-module-development-guide/20-wire-registration
audience: [backend-engineer, tech-lead]
related:
  - path: ./02-module-facade.md
    title: Module Facade
  - path: ./04-application-wire.md
    title: Application Wire File
---

# Provider Sets

Provider sets group related constructors. The application wire file composes all module provider sets into one `wire.Build(...)` call.

## Repository Set

```go
// Provides both repo interfaces for the module
var ContractsRepositorySet = wire.NewSet(
	repository.NewContractRepository,       // ContractRepository
	repository.NewContractLineRepository,   // ContractLineRepository
)
```

## Service Set

```go
// Provides ContractService
// All dependencies (repos, authzSvc, etc.) must be provided by other sets
var ContractsServiceSet = wire.NewSet(
	service.NewContractService,
)
```

## Constructor Signatures Must Match

Wire matches function parameters by type. The repository constructors must accept `db.Store`:

```go
// CORRECT — Wire can inject db.Store
func NewContractRepository(store db.Store) repository.ContractRepository {

// WRONG — Wire can't distinguish two string params
func NewContractRepository(connStr string, schema string) repository.ContractRepository {
```

The service constructor must declare all its dependencies as parameters:

```go
func NewContractService(
	repo     repository.ContractRepository,
	lineRepo repository.ContractLineRepository,
	authzSvc iam.AuthzService,
	auditSvc audit.Service,
	notifSvc notifications.Service,
	eventBus events.Bus,
	logger   logger.Logger,
	tracer   tracing.Service,
	metrics  metrics.MetricsProvider,
) service.ContractService {
```

Wire satisfies each parameter from the provider graph.

## Named Provider for Disambiguation

If two providers return the same type (e.g., two `logger.Logger` instances), use Wire's value binding or named types:

```go
// Module-scoped logger via wire.Value or type wrapper
type ContractsLogger logger.Logger

func ProvideContractsLogger(base logger.Logger) ContractsLogger {
	return ContractsLogger(base.With().Str("module", "contracts").Logger())
}
```

## Checking Provider Sets

After adding a new module, run `make wire` and check for errors. Common errors:

| Error | Cause | Fix |
|-------|-------|-----|
| `missing provider for ContractRepository` | Forgot to include `ContractsRepositorySet` | Add it to `wire.Build` |
| `has no provider for "contracts/service".ContractService` | Missing `ContractsServiceSet` | Add it to `wire.Build` |
| `constructor has two providers of type *db.Store` | Duplicate provider | Remove duplicate |
| `parameter order prevents injection` | Wire can't resolve circular dependencies | Restructure dependencies |
