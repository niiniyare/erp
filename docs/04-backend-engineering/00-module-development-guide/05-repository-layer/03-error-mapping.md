---
title: Repository Error Mapping
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer]
related:
  - "[Repository Layer Overview](01-repository-overview.md)"
  - "[Error Handling Overview](../17-error-handling/01-error-handling-overview.md)"
  - "[Error Catalog](../17-error-handling/02-error-catalog.md)"
---

# Repository Error Mapping

The repository layer is the only place that knows about database error types. It translates `pgx` errors into domain errors so the service layer never imports `pgx` or `pgconn`.

## parseModuleDBError Pattern

Every module has exactly one error mapping function:

```go
// internal/core/contracts/repository/error_mapping.go
package repository

import (
    "errors"
    "fmt"
    "strings"

    "github.com/jackc/pgx/v5/pgconn"

    "awo.so/internal/core/contracts/domain"
    sharedErr "awo.so/internal/shared/errors"
)

// parseContractDBError translates pgx errors into domain errors.
// op is the calling function name for error context.
func parseContractDBError(err error, op string) error {
    if err == nil {
        return nil
    }

    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) {
        switch pgErr.Code {
        case "23505": // unique_violation
            return mapUniqueViolation(pgErr)
        case "23503": // foreign_key_violation
            return mapForeignKeyViolation(pgErr)
        case "23514": // check_violation
            return &sharedErr.BusinessError{
                Code:    "INVALID_DATA",
                Message: fmt.Sprintf("data violates constraint: %s", pgErr.ConstraintName),
                Status:  400,
                Err:     err,
            }
        case "40001": // serialization_failure
            return domain.ErrContractConflict
        case "40P01": // deadlock_detected
            return domain.ErrContractConflict
        }
    }

    return fmt.Errorf("contracts/%s: %w", op, err)
}

func mapUniqueViolation(pgErr *pgconn.PgError) error {
    if strings.Contains(pgErr.ConstraintName, "contract_number") {
        return domain.ErrContractAlreadyExists
    }
    return &sharedErr.BusinessError{
        Code:    "DUPLICATE",
        Message: fmt.Sprintf("duplicate value violates constraint %s", pgErr.ConstraintName),
        Status:  409,
        Err:     pgErr,
    }
}

func mapForeignKeyViolation(pgErr *pgconn.PgError) error {
    if strings.Contains(pgErr.ConstraintName, "vendor_id") {
        return &sharedErr.BusinessError{
            Code:    "VENDOR_NOT_FOUND",
            Message: "referenced vendor does not exist",
            Status:  422,
            Err:     pgErr,
        }
    }
    if strings.Contains(pgErr.ConstraintName, "entity_id") {
        return &sharedErr.BusinessError{
            Code:    "ENTITY_NOT_FOUND",
            Message: "referenced entity does not exist",
            Status:  422,
            Err:     pgErr,
        }
    }
    return fmt.Errorf("foreign key violation: %s: %w", pgErr.ConstraintName, pgErr)
}
```

## pgx.ErrNoRows Handling

`pgx.ErrNoRows` is returned when a query finds no rows. The correct sentinel depends on context:

```go
func (r *contractRepository) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.Contract, error) {
    var result *domain.Contract
    err := r.store.WithTenant(ctx, tenantID, func(q *db.Queries) error {
        row, err := q.GetContractByID(ctx, db.GetContractByIDParams{
            ID: id, TenantID: tenantID,
        })
        if err != nil {
            if errors.Is(err, pgx.ErrNoRows) {
                return domain.ErrContractNotFound  // specific sentinel
            }
            return parseContractDBError(err, "GetByID")
        }
        result = mapToDomain(row)
        return nil
    })
    return result, err
}
```

For update with optimistic locking, ErrNoRows could be "not found" OR "version conflict":

```go
func (r *contractRepository) Update(ctx context.Context, p UpdateParams) (*domain.Contract, error) {
    var result *domain.Contract
    err := r.store.WithTenant(ctx, p.TenantID, func(q *db.Queries) error {
        row, err := q.UpdateContract(ctx, toSQLCUpdateParams(p))
        if err != nil {
            if errors.Is(err, pgx.ErrNoRows) {
                // Ambiguous: could be not found OR version mismatch.
                // Return version conflict — caller can re-fetch to check.
                return domain.ErrContractConflict
            }
            return parseContractDBError(err, "Update")
        }
        result = mapToDomain(row)
        return nil
    })
    return result, err
}
```

Service callers should handle `ErrContractConflict` by returning 409 and asking the client to re-fetch.

## PostgreSQL Error Codes Reference

| SQLSTATE | Name | Domain Mapping |
|----------|------|----------------|
| `23505` | unique_violation | `ErrAlreadyExists` or `BusinessError{409}` |
| `23503` | foreign_key_violation | `BusinessError{422}` with resource name |
| `23514` | check_violation | `BusinessError{400}` with constraint name |
| `23502` | not_null_violation | `BusinessError{400}` |
| `40001` | serialization_failure | `ErrVersionConflict` |
| `40P01` | deadlock_detected | `ErrVersionConflict` |
| `22001` | string_data_right_truncation | `BusinessError{400}` |
| `08006` | connection_failure | Propagate as-is (infrastructure error) |

## What Must NOT Leak from Repository

| Type | Why |
|------|-----|
| `*pgconn.PgError` | DB-specific; service should not import pgconn |
| `pgx.ErrNoRows` | DB-specific sentinel |
| Raw SQL text | Exposes DB schema details |
| `pgtype.*` types | DB encoding types; use Go native types |

All `pgx` types must be converted to domain types or standard Go types before returning from any repository method.

## BusinessError vs. Sentinel

| Type | Use when |
|------|---------|
| Sentinel (`errors.New(...)`) | Fixed error without dynamic context |
| `*BusinessError` | Need HTTP status + client message + optional wrapping |

```go
// Sentinel — no context needed
var ErrContractNotFound = errors.New("contract not found")

// BusinessError — needs HTTP status
return &sharedErr.BusinessError{
    Code:    "CONTRACT_NUMBER_TAKEN",
    Message: fmt.Sprintf("contract number %q is already in use", number),
    Status:  409,
    Err:     pgErr,
}
```

In the handler, `mapError` handles both:

```go
func (h *contractHandler) mapError(err error) error {
    var bizErr *sharedErr.BusinessError
    if errors.As(err, &bizErr) {
        return c.Status(bizErr.Status).JSON(fiber.Map{
            "error": fiber.Map{
                "code":    bizErr.Code,
                "message": bizErr.Message,
            },
        })
    }
    switch {
    case errors.Is(err, domain.ErrContractNotFound):
        return fiber.NewError(fiber.StatusNotFound, "contract not found")
    // ...
    }
}
```
