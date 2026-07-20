> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Guide Conventions
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer]
related:
  - "[What This Guide Covers](01-what-this-guide-covers.md)"
  - "[Domain Layer](../02-domain-layer/01-domain-overview.md)"
  - "[Worked Example](../23-worked-example/01-worked-example-overview.md)"
---

# Guide Conventions

## The Running Example

All 23 sections use the **Contracts module** as the single running example. Code in every section is taken from or consistent with the contracts codebase. When you apply the patterns to a new module, substitute your module name, entity names, and domain rules.

## Module Skeleton

The contracts module lives at `internal/core/contracts/`. Every module follows this layout:

```
internal/core/{module}/
  domain/
    {entity}.go             ← domain struct
    status.go               ← status type + transition table
    errors.go               ← sentinel errors + BusinessError
    events.go               ← domain event types
    value_objects.go        ← small typed values
  repository/
    interface.go            ← ContractRepository interface
    {entity}_sqlc.go        ← SQLC implementation
    error_mapping.go        ← pgx errors → domain errors
    mappers.go              ← sqlc row → domain struct
  service/
    {entity}_service.go     ← business logic
  handler/
    {entity}_handler.go     ← HTTP request/response
    routes.go               ← route registration
    errors.go               ← domain error → fiber.Error
    dto.go                  ← request/response structs
  wire.go                   ← ProviderSet, WorkerSet
```

## Import Rules

| Package | May import |
|---------|-----------|
| `domain` | stdlib + `uuid` + `decimal` only |
| `repository` | `domain`, `sqlc`, `pgx`, `db.Store` |
| `service` | `domain`, `repository`, `iam`, `events`, `notifications`, `audit` |
| `handler` | `domain`, `service`, `middleware`, `fiber` |
| `wire.go` | All of the above |

**Never** import handler from service, or service from repository, or domain from anything in the module.

## Naming Conventions

| Thing | Convention | Example |
|-------|-----------|---------|
| Module package | lowercase single word | `contracts` |
| Domain struct | PascalCase noun | `Contract`, `ContractLine` |
| Status type | `{Entity}Status` string type | `ContractStatus` |
| Status constants | `Status{State}` | `StatusDraft`, `StatusActive` |
| Repository interface | `{Entity}Repository` | `ContractRepository` |
| Service struct | `{Entity}Service` | `ContractService` |
| Handler struct | `{Entity}Handler` | `ContractHandler` |
| Wire set | `{Module}Set` | `ContractsSet` |
| Worker set | `WorkerSet` | `WorkerSet` |
| DB table | plural snake_case | `contracts`, `contract_lines` |
| DB column | snake_case | `contract_number`, `total_value` |
| Migration group | two-digit prefix | `11xxxx` for contracts |
| Permission string | `{module}.{resource}.{action}` | `contracts.contract.approve` |
| Task queue | `awoerp.{module}` | `awoerp.contracts` |
| Event topic | `{module}.{entity_state}` | `contracts.submitted` |

## Error Handling Rules

1. Map DB errors in the repository — never let `pgx.PgError` escape to the service
2. Map domain errors in the handler — never let `*domain.BusinessError` escape to the HTTP response as-is
3. Always use `errors.Is`/`errors.As` — never type-switch on errors
4. Wrap errors with `%w` to preserve the chain
5. Async goroutines must `recover()` — panics in goroutines crash the process

## Async Side Effects

Any operation that should not block the HTTP response (audit, events, notifications):

```go
go func() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    defer func() { recover() }()
    // ... side effect
}()
```

Rules:
- Use `context.Background()` — not the request context (which cancels when the request completes)
- Always set a timeout (5s)
- Always recover panics
- Log failures, never return them to the caller

## Session Rules

`ResolvedSession` (from `iam` package):

| Use | Don't use |
|-----|-----------|
| `sess.TenantID` for DB calls | URL params for tenant ID |
| `sess.UserID` for audit | Session data from request body |
| `sess.ToPrincipal()` for authz | `sess.Can(...)` — doesn't exist |
| `sess.FeatureEnabled(...)` | `sess.HasRole(...)` — doesn't exist |
| `sess.SettingString(...)` | Hard-coded config values |

## Version in Mutations

All update/delete/transition operations require a `version` field from the client:

```go
// PUT /contracts/:id
type UpdateContractRequest struct {
    Title   string `json:"title"`
    Version int    `json:"version" validate:"required,min=1"`
}
```

The repo query uses `WHERE version = @version`. If the row was updated between the client's GET and PUT, the UPDATE returns 0 rows → `ErrVersionConflict` → 409.

## Monetary Values

| Layer | Type |
|-------|------|
| DB column | `numeric(20,6)` |
| Go domain/service | `decimal.Decimal` |
| JSON request body | parsed from string |
| JSON response | `string` (`.String()` on decimal) |

Never use `float64` for money. Ever.
