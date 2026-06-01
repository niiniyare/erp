---
title: Domain Errors
portal: 4 — Backend Engineering
section: 00-module-development-guide/02-ddd-domain-design
audience: [backend-engineer, tech-lead]
related:
  - path: ./06-domain-events.md
    title: Domain Events
  - path: ../05-repository-layer/04-error-mapping.md
    title: Error Mapping
---

# Domain Errors

`domain/errors.go` declares sentinel error variables. Each variable represents one distinct business situation that a caller (service, handler) handles differently. The set of domain errors defines the error vocabulary for the module.

## The File

```go
// internal/core/contracts/domain/errors.go
package domain

import "errors"

// Sentinel errors for the contracts bounded context.
// Callers use errors.Is() to identify these conditions.
var (
	// ErrContractNotFound is returned when a contract ID has no matching record
	// in the tenant's data, or when the record is soft-deleted.
	ErrContractNotFound = errors.New("contract not found")

	// ErrContractAlreadyExists is returned when creating a contract whose
	// ContractNumber is already in use within the tenant.
	ErrContractAlreadyExists = errors.New("contract already exists")

	// ErrContractInvalidTransition is returned when a requested status change
	// is not permitted by the state machine (CanTransitionTo returned false).
	ErrContractInvalidTransition = errors.New("invalid contract status transition")

	// ErrContractConflict is returned when an optimistic lock check fails —
	// the contract was modified by another writer between fetch and update.
	ErrContractConflict = errors.New("contract version conflict")

	// ErrContractValueExceedsLimit is returned when a contract's total value
	// exceeds the tenant-configured approval threshold and no approval has
	// been granted.
	ErrContractValueExceedsLimit = errors.New("contract value exceeds approved limit")

	// ErrContractNotEditable is returned when attempting to modify a contract
	// that is not in draft status.
	ErrContractNotEditable = errors.New("contract is not editable in its current status")

	// ErrContractLineNotFound is returned when a contract line ID has no match.
	ErrContractLineNotFound = errors.New("contract line not found")
)
```

## One Error Per Distinct Situation

The rule: if a caller (handler) would handle two errors identically (same HTTP status, same user message), they should be one error. If the handler would respond differently, they must be separate errors.

| Error | Handler response |
|-------|-----------------|
| `ErrContractNotFound` | `404 Not Found` |
| `ErrContractAlreadyExists` | `409 Conflict` |
| `ErrContractInvalidTransition` | `422 Unprocessable Entity` |
| `ErrContractConflict` | `409 Conflict` |
| `ErrContractValueExceedsLimit` | `422 Unprocessable Entity` |
| `ErrContractNotEditable` | `422 Unprocessable Entity` |

`ErrContractAlreadyExists` and `ErrContractConflict` both map to 409, but they carry different messages and may trigger different client behaviour (retry vs. change the contract number), so they are separate errors.

`ErrContractInvalidTransition` and `ErrContractValueExceedsLimit` both map to 422, but they describe different business failures that UI code may handle differently (show state machine hint vs. show approval threshold info).

## Wrapping Domain Errors

When a repository adapts a database error, it wraps using `fmt.Errorf("... %w", domainErr)` so callers can use `errors.Is`:

```go
// repository/<noun>_sqlc.go
if errors.Is(err, pgx.ErrNoRows) {
	return fmt.Errorf("GetContractByID: %w", domain.ErrContractNotFound)
}
```

The handler then uses `errors.Is` — never a type switch or string comparison:

```go
// handlers/errors.go
func mapError(err error) error {
	switch {
	case errors.Is(err, domain.ErrContractNotFound):
		return fiber.NewError(fiber.StatusNotFound, "contract not found")
	case errors.Is(err, domain.ErrContractConflict):
		return fiber.NewError(fiber.StatusConflict, "contract was modified by another user, please retry")
	// ...
	}
}
```

## What Not to Put in Domain Errors

**Do not include HTTP status codes in domain errors.** The domain knows nothing about HTTP. Status code mapping happens in `handlers/errors.go`.

**Do not include tenant or user context in error messages.** Error messages are strings — no IDs, no email addresses. Structured context goes in log fields, not error messages.

**Do not use `fmt.Errorf` with `%w` at the call site in the domain.** Domain errors are declared as package-level variables. Only the repository and service layer wrap them.

**Do not create a catch-all `ErrContractInternal` or similar.** Unknown errors propagate as-is up to the handler, which maps them to 500. Adding an internal sentinel adds noise without benefit.

## Cross-Context Error Handling

When a service calls another module's service (e.g., calling `vendorSvc.GetByID`), do not wrap the foreign error in a domain error:

```go
// CORRECT — propagate foreign error as-is
vendor, err := s.vendorSvc.GetByID(ctx, contract.VendorID, tenantID)
if err != nil {
	return nil, err  // handler's mapError will catch unknown → 500
}

// WRONG — wrapping foreign error in this module's sentinel
if err != nil {
	return nil, domain.ErrContractNotFound  // misleading — the vendor wasn't found
}
```

The calling handler should only need to recognise errors from its own domain. Cross-module errors that reach the handler without matching any sentinel will correctly produce a 500 response.
