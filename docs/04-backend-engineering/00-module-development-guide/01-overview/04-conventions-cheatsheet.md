---
title: Conventions Cheatsheet
portal: 4 — Backend Engineering
section: 00-module-development-guide/01-overview
audience: [backend-engineer, tech-lead]
related:
  - path: ./02-module-anatomy.md
    title: Module Anatomy
  - path: ./03-development-sequence.md
    title: Development Sequence
---

# Conventions Cheatsheet

Quick reference for naming, formatting, and structural rules. Deviations break Wire, SQLC, linting, and test discovery. Keep this open while coding.

---

## Naming

### Module Key

| Thing | Convention | Example |
|-------|-----------|---------|
| Module key | lowercase, plural noun | `contracts` |
| Primary noun | singular | `contract` |
| Child noun | singular, underscore | `contract_line` |
| Go package | same as module key | `package contracts` |

### Go Identifiers

| Thing | Convention | Example |
|-------|-----------|---------|
| Entity struct | `PascalCase` singular | `Contract` |
| Status type | `<Entity>Status` | `ContractStatus` |
| Status constant | `<Entity>Status<State>` | `ContractStatusDraft` |
| Repository interface | `<Entity>Repository` | `ContractRepository` |
| SQLC adapter struct | `<entity>SQLCRepository` (unexported) | `contractSQLCRepository` |
| Service interface | `<Entity>Service` | `ContractService` |
| Service struct | `<entity>Service` (unexported) | `contractService` |
| Handler struct | `Handler` | `Handler` |
| Wire set (repo) | `<Module>RepositorySet` | `ContractsRepositorySet` |
| Wire set (service) | `<Module>ServiceSet` | `ContractsServiceSet` |
| Wire set (all) | `<Module>Set` | `ContractsSet` |

### Domain Errors

```go
// Pattern: Err<Entity><Situation>
var (
    ErrContractNotFound          = errors.New("contract not found")
    ErrContractAlreadyExists     = errors.New("contract already exists")
    ErrContractInvalidTransition = errors.New("invalid contract status transition")
    ErrContractConflict          = errors.New("contract version conflict")
    ErrContractValueExceedsLimit = errors.New("contract value exceeds approved limit")
)
```

### Domain Events

```go
// Pattern: <Entity><PastTense> — always past tense
type ContractCreated   struct { ... }
type ContractSubmitted struct { ... }
type ContractApproved  struct { ... }
type ContractActivated struct { ... }
```

### Params Structs

```go
// Pattern: <Verb><Entity>Params (not request/input/args)
type CreateContractParams struct { ... }
type UpdateContractParams struct { ... }
type ListContractsParams  struct { ... }   // plural for list
type UpdateContractStatusParams struct { ... }
```

### SQLC Query Names

```sql
-- name: CreateContract       :one
-- name: GetContractByID      :one
-- name: ListContracts        :many
-- name: CountContracts       :one
-- name: UpdateContract       :one
-- name: UpdateContractStatus :one
-- name: SoftDeleteContract   :one
-- name: ContractExistsWithNumber :one
```

### Permission Strings

Format: `module.resource.action` — all lowercase, dot-separated.

```
contracts.contract.create
contracts.contract.read
contracts.contract.update
contracts.contract.delete
contracts.contract.submit
contracts.contract.approve
contracts.contract.activate
contracts.contract.terminate
contracts.contract_line.create
contracts.contract_line.read
```

### Migration Files

```
db/migration/NNN_001_create_<module>.sql          # primary table
db/migration/NNN_002_create_<module>_lines.sql    # child tables
db/migration/NNN_003_create_<module>_views.sql    # views
```

`NNN` = three-digit module group number (e.g., `011`). Assigned at module creation. Never reuse.

---

## Database Conventions

### Required Columns (every table)

```sql
id          uuid        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
tenant_id   uuid        NOT NULL REFERENCES tenants(id),
entity_id   uuid        NOT NULL REFERENCES entities(id),
-- module-specific columns here --
status      VARCHAR(50) NOT NULL CHECK (status IN ('draft', 'active', ...)),
version     integer     NOT NULL DEFAULT 1,
created_by  uuid        NOT NULL,
updated_by  uuid        NOT NULL,
deleted_at  timestamptz,                  -- nullable: soft-delete
created_at  timestamptz NOT NULL DEFAULT now(),
updated_at  timestamptz NOT NULL DEFAULT now()
```

### RLS Block (every table)

```sql
ALTER TABLE contracts ENABLE ROW LEVEL SECURITY;

CREATE POLICY contracts_tenant_isolation ON contracts
    USING (tenant_id = current_setting('app.tenant_id')::uuid);
```

### Monetary Columns

```sql
-- CORRECT
total_value numeric(20,6) NOT NULL DEFAULT 0,

-- WRONG — never use float
total_value float NOT NULL DEFAULT 0,
total_value double precision NOT NULL DEFAULT 0,
```

### Optimistic Lock Update Pattern

```sql
-- name: UpdateContract :one
UPDATE contracts
SET
    -- fields --
    version    = version + 1,
    updated_by = @updated_by,
    updated_at = now()
WHERE id        = @id
  AND tenant_id = @tenant_id
  AND version   = @version          -- optimistic lock check
  AND deleted_at IS NULL
RETURNING *;
```

### Status Update (separate query)

```sql
-- name: UpdateContractStatus :one
UPDATE contracts
SET
    status     = @status,
    version    = version + 1,
    updated_by = @updated_by,
    updated_at = now()
WHERE id        = @id
  AND tenant_id = @tenant_id
  AND version   = @version
  AND deleted_at IS NULL
RETURNING *;
```

---

## Go Conventions

### Entity Required Fields Order

```go
type Contract struct {
    ID        uuid.UUID
    TenantID  uuid.UUID    // always second
    EntityID  uuid.UUID    // always third
    // ...module fields...
    Status    ContractStatus
    Version   int
    CreatedBy uuid.UUID
    UpdatedBy uuid.UUID
    DeletedAt *time.Time   // pointer — nil means not deleted
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### Session Extraction (handlers only)

```go
// CORRECT — extract in handler, pass principal to service
sess, ok := c.Locals(domain.LocalsKeySession).(*iam.ResolvedSession)
if !ok || sess == nil {
    return fiber.ErrUnauthorized
}

// WRONG — never do this
tenantID := c.Params("tenantID")  // never trust URL for identity
userID   := c.Get("X-User-ID")   // never trust headers for identity
```

### Service Authorization Call

```go
// CORRECT
allowed, err := s.authzSvc.Enforce(ctx, iam.Request{
    Subject: iam.TenantSubject(principal.UserID),
    Domain:  iam.TenantDomain(tenantID),
    Object:  "contracts/contract/*",
    Action:  "create",
})
if err != nil {
    return nil, err
}
if !allowed {
    return nil, iam.ErrForbidden
}

// WRONG — session has no Can() or HasRole() methods
if !sess.Can("contracts.contract.create") { ... }  // does not exist
if sess.HasRole("admin") { ... }                   // does not exist
```

### Feature Flag Check

```go
// CORRECT — snapshotted at login, no DB call
if sess.FeatureEnabled("contracts.multi_currency") {
    // ...
}

// Settings
approvalThreshold := sess.SettingDecimal("contracts.approval_threshold", 50000.0)
```

### Store WithTenant Pattern

```go
// CORRECT — always scope DB calls to tenant
var contract *domain.Contract
err = r.store.WithTenant(ctx, tenantID, func(q *db.Queries) error {
    row, err := q.GetContractByID(ctx, db.GetContractByIDParams{
        ID:       id,
        TenantID: tenantID,
    })
    if err != nil {
        return err
    }
    contract = mapRowToDomain(row)
    return nil
})

// WRONG — never call queries directly
row, err := r.store.GetContractByID(ctx, ...)  // no tenant isolation
```

### Error Mapping (repository)

```go
func mapContractDBError(err error, op string) error {
    if err == nil {
        return nil
    }
    if errors.Is(err, pgx.ErrNoRows) {
        return domain.ErrContractNotFound
    }
    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) {
        switch pgErr.Code {
        case "23505": // unique_violation
            return domain.ErrContractAlreadyExists
        case "23503": // foreign_key_violation
            return errors.New("contract references invalid entity")
        }
    }
    return fmt.Errorf("%s: %w", op, err)
}
```

### Error Mapping (handler)

```go
func mapError(err error) error {
    switch {
    case errors.Is(err, domain.ErrContractNotFound):
        return fiber.NewError(fiber.StatusNotFound, err.Error())
    case errors.Is(err, domain.ErrContractAlreadyExists):
        return fiber.NewError(fiber.StatusConflict, err.Error())
    case errors.Is(err, domain.ErrContractInvalidTransition):
        return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
    case errors.Is(err, domain.ErrContractConflict):
        return fiber.NewError(fiber.StatusConflict, err.Error())
    case errors.Is(err, iam.ErrForbidden):
        return fiber.NewError(fiber.StatusForbidden, "forbidden")
    case errors.Is(err, iam.ErrUnauthorized):
        return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
    default:
        return fiber.NewError(fiber.StatusInternalServerError, "internal error")
    }
}
```

---

## File Headers

Every Go file has the standard package declaration and NO file-level comment unless it is the package doc comment.

```go
package domain  // or repository, service, contracts, handlers

import (
    // stdlib first
    "context"
    "errors"
    "time"

    // third-party second
    "github.com/google/uuid"

    // internal last
    "awo.so/internal/core/iam/domain"
    db "awo.so/db/sqlc"
)
```

---

## What Never Goes Where

| Thing | Never in |
|-------|---------|
| Business logic | `repository/` |
| Permission checks | `repository/` |
| SQLC types | `service/` interface signatures |
| `db.*` types | handler response DTOs |
| `c.Params()` for tenantID/userID | `service/` |
| `float64` for money | anywhere |
| `serial` / `int` PK | migrations |
| `init()` functions | module packages |
| Manual edits | `wire_gen.go` |
| Raw SQL | anywhere except `db/queries/*.sql` |

---

## Make Targets Reference

| Target | When to run |
|--------|------------|
| `make wire` | After any Wire provider change |
| `make sqlc` | After any `db/queries/*.sql` change |
| `make migrate-up` | After adding a migration file |
| `make migrate-down` | To verify rollback |
| `make test` | Before every commit |
| `make lint` | Before every commit |
| `make generate` | Runs `wire` + `sqlc` together |
