> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Provider Patterns
portal: 3 — Platform Architecture
section: 06-dependency-injection
audience: [architect, backend-engineer]
related:
  - "[Wire Overview](01-wire-overview.md)"
  - "[Startup Sequence](../../04-backend-engineering/00-module-development-guide/21-server-startup/03-startup-sequence.md)"
  - "[Module Development Guide](../../04-backend-engineering/00-module-development-guide/01-overview/01-what-this-guide-covers.md)"
---

# Provider Patterns

## Provider Function Signatures

Wire providers are constructor functions. The convention:

```go
// Single dependency → direct constructor
func NewContractRepo(db *pgxpool.Pool) ContractRepository {
    return &sqlcContractRepo{db: db}
}

// Multiple dependencies → struct injection
func NewContractService(
    repo   ContractRepository,
    authz  authz.Service,
    events events.Publisher,
    notif  notifications.Service,
    audit  audit.Service,
    logger *slog.Logger,
) *ContractService {
    return &ContractService{
        repo:   repo,
        authz:  authz,
        events: events,
        notif:  notif,
        audit:  audit,
        logger: logger.With("module", "contracts"),
    }
}
```

## Interface Binding

When a provider returns a concrete type but the consumer expects an interface, use `wire.Bind`:

```go
var ContractsSet = wire.NewSet(
    NewContractRepo,
    wire.Bind(new(ContractRepository), new(*sqlcContractRepo)),
    NewContractService,
    NewContractHandler,
)
```

Or return the interface directly from the provider — simpler when no ambiguity:

```go
func NewContractRepo(db *pgxpool.Pool) ContractRepository {
    return &sqlcContractRepo{db: db}  // returns interface
}
```

## ProviderSet Boundaries

Each module owns a `ProviderSet`. Shared infrastructure has its own sets.

```
internal/
  core/
    contracts/
      wire.go          ← ContractsSet
    finance/
      wire.go          ← FinanceSet
    iam/
      wire.go          ← IAMSet
  shared/
    db/
      wire.go          ← DatabaseSet
    events/
      wire.go          ← EventsSet
    audit/
      wire.go          ← AuditSet
cmd/
  server/
    wire.go            ← App injector (assembles all sets)
    wire_gen.go        ← generated; never edit
```

`DatabaseSet` example:

```go
// internal/shared/db/wire.go
var DatabaseSet = wire.NewSet(
    NewPool,           // *pgxpool.Pool
    NewStore,          // *Store (wraps pool with WithTenant)
    RunMigrations,     // runs on startup
)
```

## Scoping Rules

| What | Rule |
|------|------|
| DB pool | Singleton — one pool per process |
| Services | Singleton — stateless, safe to share |
| Handlers | Singleton — stateless |
| Repositories | Singleton — no per-request state |
| Workers | Singleton — long-running goroutines |
| HTTP server | Singleton |

Wire creates each type once unless you use `wire.Value` with distinct keys. Since all types here are singletons, standard `wire.NewSet` works.

## Value Providers

For config values passed into providers:

```go
func NewPool(cfg Config) (*pgxpool.Pool, func(), error) {
    pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
    if err != nil {
        return nil, nil, err
    }
    cleanup := func() { pool.Close() }
    return pool, cleanup, nil
}
```

Wire handles cleanup functions automatically (called on shutdown in reverse order).

## WorkerSet Pattern

Temporal workers are registered separately from the HTTP server set:

```go
// internal/core/contracts/wire.go
var ContractsSet = wire.NewSet(
    NewContractRepo,
    NewContractService,
    NewContractHandler,
    NewContractRouteRegistrar,
)

var WorkerSet = wire.NewSet(
    NewContractActivities,
    NewContractWorker,
)
```

The main injector includes both:

```go
//go:build wireinject

func InitializeApp(cfg Config) (*App, func(), error) {
    wire.Build(
        DatabaseSet,
        EventsSet,
        IAMSet,
        ContractsSet,
        ContractsWorkerSet,
        FinanceSet,
        // ...
        NewApp,
    )
    return nil, nil, nil
}
```

## Common Errors

| Error | Cause | Fix |
|-------|-------|-----|
| `wire: no provider for X` | Missing provider in set | Add provider or import the set that has it |
| `wire: ambiguous binding` | Two providers return same type | Use distinct interface types or `wire.ProviderFunc` with tag |
| `wire: cycle detected` | A→B→A | Extract shared dep to third type |
| `wire_gen.go out of date` | Forgot `make wire` after changing providers | Run `make wire` |

## Practical Guidelines

- Prefer returning interfaces from providers, not concrete types — consumers code to the interface.
- Never add `//go:build wireinject` to `wire_gen.go` — the generator manages it.
- One `ProviderSet` per module package. Don't split into multiple sets within one module.
- Pass `*slog.Logger` as a provider from `DatabaseSet` / infrastructure — modules derive child loggers.
- `wire.Value(x)` for values that are already constructed (e.g., passing a pre-built config struct).
