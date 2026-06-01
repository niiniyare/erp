---
title: Error Mapping
portal: 4 — Backend Engineering
section: 00-module-development-guide/05-repository-layer
audience: [backend-engineer, tech-lead]
related:
  - path: ./02-sqlc-adapter.md
    title: SQLC Adapter
  - path: ../02-ddd-domain-design/05-domain-errors.md
    title: Domain Errors
---

# Error Mapping

The repository adapter is the translation boundary between PostgreSQL errors and domain errors. Callers above the repository never see `pgx` or `pgconn` error types.

## The Mapping Function

```go
// internal/core/contracts/repository/errors.go
package repository

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"awo.so/internal/core/contracts/domain"
)

// mapContractDBError translates low-level database errors into domain sentinels.
// op is the operation name for wrapping context (e.g., "Create", "Update").
func mapContractDBError(err error, op string) error {
	if err == nil {
		return nil
	}

	// pgx.ErrNoRows: no matching row for a :one query
	// Callers that need not-found vs conflict disambiguation call exists() first.
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrContractNotFound
	}

	// pgconn.PgError: structured PostgreSQL error with SQLSTATE code
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			// Which unique constraint was violated?
			if pgErr.ConstraintName == "uq_contracts_tenant_number" {
				return domain.ErrContractAlreadyExists
			}
			return fmt.Errorf("%s: unique constraint violation on %s: %w",
				op, pgErr.ConstraintName, domain.ErrContractAlreadyExists)

		case "23503": // foreign_key_violation
			return fmt.Errorf("%s: foreign key violation: %w", op, err)

		case "23514": // check_violation
			return fmt.Errorf("%s: check constraint violation on %s: %w",
				op, pgErr.ConstraintName, err)

		case "40001": // serialization_failure (retry-able)
			return fmt.Errorf("%s: serialization failure: %w", op, err)
		}
	}

	// Unknown DB error: wrap with context but do not translate
	return fmt.Errorf("%s: %w", op, err)
}
```

## SQLSTATE Codes Used

| Code | Meaning | Domain mapping |
|------|---------|----------------|
| `23505` | unique_violation | `ErrContractAlreadyExists` |
| `23503` | foreign_key_violation | Wrapped with context |
| `23514` | check_violation | Wrapped with context |
| `40001` | serialization_failure | Wrapped with context (service may retry) |
| pgx.ErrNoRows | zero rows from `:one` query | `ErrContractNotFound` (default) |

## Disambiguating ErrNoRows

A RETURNING UPDATE that matches zero rows produces `pgx.ErrNoRows`. This happens in two cases:
1. The ID does not exist (or is soft-deleted).
2. The ID exists but the version check failed (concurrent write).

The adapter must distinguish these to return the correct error:

```go
// In Update / UpdateStatus / Delete
if errors.Is(err, pgx.ErrNoRows) {
	if exists, _ := r.exists(ctx, params.ID, params.TenantID); exists {
		return nil, domain.ErrContractConflict  // record exists, version mismatch
	}
	return nil, domain.ErrContractNotFound     // record does not exist
}
```

The `exists()` helper is a cheap `SELECT COUNT(*)` by ID that ignores soft-delete. It is called only in error paths, so it does not affect happy-path performance.

## Handler's mapError Function

The handler maps domain errors to HTTP responses. It uses `errors.Is` — never a type switch:

```go
// internal/api/handlers/contracts/errors.go
package contracts

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"awo.so/internal/core/contracts/domain"
	iam "awo.so/internal/core/iam"
)

// mapError translates domain errors to Fiber HTTP errors.
func mapError(err error) error {
	switch {
	case errors.Is(err, domain.ErrContractNotFound):
		return fiber.NewError(fiber.StatusNotFound, "contract not found")
	case errors.Is(err, domain.ErrContractAlreadyExists):
		return fiber.NewError(fiber.StatusConflict, "a contract with this number already exists")
	case errors.Is(err, domain.ErrContractInvalidTransition):
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, domain.ErrContractConflict):
		return fiber.NewError(fiber.StatusConflict, "contract was modified by another user; please reload and retry")
	case errors.Is(err, domain.ErrContractValueExceedsLimit):
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, domain.ErrContractNotEditable):
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, domain.ErrContractLineNotFound):
		return fiber.NewError(fiber.StatusNotFound, "contract line not found")
	case errors.Is(err, iam.ErrForbidden):
		return fiber.NewError(fiber.StatusForbidden, "you do not have permission to perform this action")
	case errors.Is(err, iam.ErrUnauthorized):
		return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
	default:
		return fiber.NewError(fiber.StatusInternalServerError, "an unexpected error occurred")
	}
}
```

## Why errors.Is, Not Type Switch

```go
// CORRECT — works with wrapped errors
case errors.Is(err, domain.ErrContractNotFound):

// WRONG — breaks when error is wrapped with %w
case err == domain.ErrContractNotFound:
```

`errors.Is` unwraps the error chain. Repository errors are wrapped with `fmt.Errorf("GetByID: %w", domain.ErrContractNotFound)`. A direct `==` comparison fails on the wrapped form; `errors.Is` succeeds.

```go
// WRONG — type switch does not traverse error chains
switch e := err.(type) {
case *domain.ContractError:
    // This misses wrapped errors
}
```

Type switches only match the outermost type. `errors.Is` and `errors.As` traverse the full chain. Always use them.
