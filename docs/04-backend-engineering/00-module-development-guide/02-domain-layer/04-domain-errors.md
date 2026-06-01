---
title: Domain Errors
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer]
related:
  - "[Domain Layer Overview](01-domain-overview.md)"
  - "[Service Layer Overview](../06-service-layer/01-service-overview.md)"
  - "[Handler Layer Overview](../07-handler-layer/01-handler-overview.md)"
---

# Domain Errors

## Error Types

Two kinds of errors flow up from the domain:

| Type | When to use |
|------|-------------|
| Sentinel errors | Fixed, well-known conditions (not found, forbidden, version conflict) |
| `*BusinessError` | Dynamic errors with message and HTTP status (validation failures, business rule violations) |

## Sentinel Errors

Define in the domain package:

```go
// internal/core/contracts/domain/errors.go
package domain

import "errors"

var (
    ErrContractNotFound        = errors.New("contract not found")
    ErrContractNotEditable     = errors.New("contract is not in an editable state")
    ErrContractInvalidTransition = errors.New("invalid contract status transition")
    ErrDuplicateContractNumber = errors.New("contract number already exists in this tenant")
    ErrVersionConflict         = errors.New("optimistic lock conflict — re-fetch and retry")
    ErrForbidden               = errors.New("permission denied")
)
```

Use `errors.Is` everywhere — never compare error strings:

```go
// Correct
if errors.Is(err, domain.ErrContractNotFound) {
    return fiber.NewError(404, "contract not found")
}

// Wrong — breaks when errors are wrapped
if err == domain.ErrContractNotFound { ... }
if err.Error() == "contract not found" { ... }
```

## BusinessError

For errors with dynamic content (reason strings, field names):

```go
// internal/core/contracts/domain/errors.go

type BusinessError struct {
    Status  int    // HTTP status code
    Code    string // Machine-readable code (e.g. "FEATURE_NOT_ENABLED")
    Message string // Human-readable message
    Fields  []FieldError // Optional: validation field errors
}

type FieldError struct {
    Field   string
    Message string
}

func (e *BusinessError) Error() string {
    return e.Message
}

// Constructors
func NewBusinessError(status int, code, message string) *BusinessError {
    return &BusinessError{Status: status, Code: code, Message: message}
}

func NewValidationError(fields []FieldError) *BusinessError {
    return &BusinessError{
        Status:  422,
        Code:    "VALIDATION_ERROR",
        Message: "validation failed",
        Fields:  fields,
    }
}
```

Usage in service:

```go
if !session.FeatureFlags["contracts.bulk_import"] {
    return nil, domain.NewBusinessError(403, "FEATURE_NOT_ENABLED",
        "bulk import is not enabled for this tenant")
}
```

## Error Wrapping

Always wrap errors with context using `%w`:

```go
contract, err := s.repo.GetByID(ctx, id, tenantID)
if err != nil {
    return nil, fmt.Errorf("contracts.service.approve: %w", err)
}
```

This preserves `errors.Is` and `errors.As` unwrapping up the call stack.

Never swallow errors:

```go
// Wrong
contract, _ := s.repo.GetByID(ctx, id, tenantID)

// Also wrong — loses the error type
if err != nil {
    return nil, errors.New("failed to get contract")
}
```

## Error Propagation Chain

```
Repository    →    Service    →    Handler    →    HTTP Response
repo error         wraps           maps to
(sentinel or  →    with %w    →    fiber.Error  →  JSON error body
pg error)          + context       via mapError()
```

Each layer adds context without changing the type:

```go
// Repository
if pgErr.Code == "23505" {
    return nil, domain.ErrDuplicateContractNumber
}

// Service
if err != nil {
    return nil, fmt.Errorf("contractService.Create: %w", err)
}

// Handler
if errors.Is(err, domain.ErrDuplicateContractNumber) {
    return fiber.NewError(409, "contract number already exists")
}
```

## Standard Error Map in Handler

```go
// handler/errors.go
func mapError(err error) error {
    switch {
    case errors.Is(err, domain.ErrContractNotFound):
        return fiber.NewError(404, "contract not found")
    case errors.Is(err, domain.ErrVersionConflict):
        return fiber.NewError(409, "version conflict — re-fetch and retry")
    case errors.Is(err, domain.ErrContractNotEditable):
        return fiber.NewError(422, "contract cannot be modified in its current state")
    case errors.Is(err, domain.ErrContractInvalidTransition):
        return fiber.NewError(422, "invalid status transition")
    case errors.Is(err, domain.ErrDuplicateContractNumber):
        return fiber.NewError(409, "contract number already exists")
    case errors.Is(err, domain.ErrForbidden):
        return fiber.NewError(403, "permission denied")
    }

    var bizErr *domain.BusinessError
    if errors.As(err, &bizErr) {
        return fiber.NewError(bizErr.Status, bizErr.Message)
    }

    return fiber.NewError(500, "internal server error")
}
```

## Module-Shared Errors

Errors used across modules live in `internal/shared/`:

```go
// internal/shared/iam/errors.go
package iam

import "errors"

var ErrForbidden = errors.New("permission denied")
var ErrUnauthenticated = errors.New("unauthenticated")
```

Domain packages import from `internal/shared/iam`, not from each other's domain packages.

## Testing Error Types

```go
func TestService_Create_Forbidden(t *testing.T) {
    svc := newServiceWithAuthz(&mockAuthz{allow: false})
    _, err := svc.Create(ctx, sess, params)

    require.Error(t, err)
    assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestService_Create_FeatureDisabled(t *testing.T) {
    svc := newService()
    sess.FeatureFlags["contracts.bulk_import"] = false

    _, err := svc.Import(ctx, sess, data)

    var bizErr *domain.BusinessError
    require.ErrorAs(t, err, &bizErr)
    assert.Equal(t, "FEATURE_NOT_ENABLED", bizErr.Code)
    assert.Equal(t, 403, bizErr.Status)
}
```
