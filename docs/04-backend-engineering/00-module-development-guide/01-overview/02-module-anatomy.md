---
title: Module Anatomy
portal: 4 — Backend Engineering
section: 00-module-development-guide/01-overview
audience: [backend-engineer, tech-lead]
related:
  - path: ./03-development-sequence.md
    title: Development Sequence
  - path: ./04-conventions-cheatsheet.md
    title: Conventions Cheatsheet
---

# Module Anatomy

Every AwoERP business module follows this directory structure. Deviations break Wire wiring, SQLC generation, and test discovery.

## Directory Layout

```
internal/core/<module>/
├── domain/
│   ├── <noun>.go           # Primary entity struct, enums, value objects, state machine
│   ├── errors.go           # Package-level sentinel error variables
│   └── events.go           # Domain event structs (published after state changes)
├── repository/
│   ├── <noun>.go           # Repository interface (the port)
│   └── <noun>_sqlc.go      # SQLC adapter (the adapter — implements the port)
├── service/
│   ├── <noun>.go           # Service interface + implementation
│   ├── <noun>_workflow.go  # Temporal workflow function + activity struct (if needed)
│   └── <noun>_pipeline.go  # Pipeline Stage and Hook types (if needed)
└── <module>.go             # Module facade: Wire provider sets + public type re-exports

internal/api/handlers/<module>/
├── handler.go              # Handler struct, constructor, all handler methods
├── routes.go               # RegisterRoutes(apiGroup fiber.Router, deps *Dependencies) function
├── request.go              # Request DTOs with validate tags
├── response.go             # Response DTOs + MapXxxToResponse() functions
└── errors.go               # mapError(err error) → *fiber.Error

db/migration/
├── NNN_001_create_<module>.sql
├── NNN_002_create_<module>_lines.sql
└── NNN_003_create_<module>_views.sql

db/queries/
└── <module>.sql            # All SQLC annotated queries for the module
```

## File Responsibilities

### `domain/<noun>.go`

The heart of the module. Contains:

- **Entity struct**: the primary domain object with all fields
- **Status enum**: `type <Noun>Status string` with all valid states as constants
- **`CanTransitionTo(next <Noun>Status) bool`**: enforces the state machine — called by the service before every status update
- **Value objects**: immutable types with validation in their constructors
- **Required fields on every entity**: `ID`, `TenantID`, `EntityID`, `Status`, `Version`, `CreatedBy`, `UpdatedBy`, `DeletedAt`, `CreatedAt`, `UpdatedAt`

```go
// Canonical entity shape — every AwoERP entity follows this pattern.
type Contract struct {
    ID        uuid.UUID
    TenantID  uuid.UUID   // Layer 1: RLS anchor — every DB call scoped to this
    EntityID  uuid.UUID   // Layer 2: org hierarchy scope (branch/dept/subsidiary)
    // ... module-specific fields ...
    Status    ContractStatus
    Version   int          // Optimistic locking: UPDATE ... WHERE version = $n
    CreatedBy uuid.UUID
    UpdatedBy uuid.UUID
    DeletedAt *time.Time   // Soft delete; nil = not deleted
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### `domain/errors.go`

Package-level sentinel errors. One error per distinct business situation a caller handles differently.

```go
var (
    ErrContractNotFound          = errors.New("contract not found")
    ErrContractAlreadyExists     = errors.New("contract already exists")
    ErrContractInvalidTransition = errors.New("invalid contract status transition")
    ErrContractConflict          = errors.New("contract version conflict")
    ErrContractValueExceedsLimit = errors.New("contract value exceeds approved limit")
)
```

### `domain/events.go`

Immutable domain events emitted after successful state changes. Always include `TenantID`. Named in past tense.

```go
type ContractCreated struct {
    ContractID uuid.UUID
    TenantID   uuid.UUID
    EntityID   uuid.UUID
    CreatedBy  uuid.UUID
    OccurredAt time.Time
}
```

### `repository/<noun>.go`

The interface (port) that the service depends on. Pure persistence — no business logic, no permission checks.

```go
type ContractRepository interface {
    Create(ctx context.Context, params CreateContractParams) (*Contract, error)
    GetByID(ctx context.Context, id, tenantID uuid.UUID) (*Contract, error)
    List(ctx context.Context, params ListContractsParams) ([]*Contract, error)
    Count(ctx context.Context, params ListContractsParams) (int64, error)
    Update(ctx context.Context, params UpdateContractParams) (*Contract, error)
    UpdateStatus(ctx context.Context, params UpdateContractStatusParams) (*Contract, error)
    Delete(ctx context.Context, id, tenantID, deletedBy uuid.UUID) error
    ExistsWithNumber(ctx context.Context, number string, tenantID uuid.UUID) (bool, error)
}
```

### `repository/<noun>_sqlc.go`

The SQLC adapter — implements the repository interface by calling SQLC-generated methods wrapped in `store.WithTenant()`.

```go
type contractSQLCRepository struct {
    store db.Store
}

func NewContractRepository(store db.Store) ContractRepository {
    return &contractSQLCRepository{store: store}
}
```

### `service/<noun>.go`

The application service. Owns:
- Permission enforcement (via `authzSvc.Enforce()` for non-route checks)
- Business rule validation
- State machine transitions (checks `CanTransitionTo()` before calling repo)
- Calls to platform services (audit, notifications, settings, feature flags)
- Domain event publication

**Constructor injects**: repository, authzSvc, audit service, notification service, settings snapshot via session, logger, tracer, metrics.

### `<module>.go` — Module Facade

Wire provider sets for the module. External packages import only this file, never the sub-packages.

```go
package contracts

import "github.com/google/wire"

var ContractsRepositorySet = wire.NewSet(
    repository.NewContractRepository,
)

var ContractsServiceSet = wire.NewSet(
    service.NewContractService,
)

var ContractsSet = wire.NewSet(
    ContractsRepositorySet,
    ContractsServiceSet,
)
```

### Handler Files

| File | Purpose |
|------|---------|
| `handler.go` | Struct, constructor, all HTTP handler methods |
| `routes.go` | `RegisterRoutes(apiGroup fiber.Router, deps *handlers.Dependencies)` |
| `request.go` | DTOs that map HTTP body/query params to service params |
| `response.go` | DTOs that map domain entities to JSON responses |
| `errors.go` | `mapError(err error) error` — domain errors → HTTP errors |

### Database Files

| File | Purpose |
|------|---------|
| `db/migration/NNN_001_*.sql` | Primary table + RLS + indexes |
| `db/migration/NNN_002_*.sql` | Child tables |
| `db/migration/NNN_003_*.sql` | Views and materialised views |
| `db/queries/<module>.sql` | All SQLC annotated queries |

`NNN` is a three-digit module group number assigned at module creation (e.g., `011` for contracts). All migrations for one module share the same prefix so they sort together.
