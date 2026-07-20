> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Error Handling Overview
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer]
related:
  - "[Service Layer](../06-service-layer/01-service-overview.md)"
  - "[Handler Layer](../07-handler-layer/01-handler-overview.md)"
  - "[API Design](../14-api-design/01-api-design-overview.md)"
---

# Error Handling Overview

## Error Flow

```
Repository
  └── pgx error
        └── parseDomainError()
              └── domain sentinel or *BusinessError

Service
  └── calls repo
        └── wraps errors with context
              └── returns domain error up

Handler
  └── calls service
        └── mapError()
              └── *BusinessError with HTTP status
                    └── JSON error response
```

## Domain Errors

Define sentinel errors and `*BusinessError` in `internal/core/{module}/domain/errors.go`:

```go
package domain

import "errors"

// Sentinel errors — use errors.Is() to check
var (
    ErrContractNotFound    = errors.New("contract not found")
    ErrContractNotEditable = errors.New("contract is not in editable state")
    ErrVersionConflict     = errors.New("version conflict")
    ErrForbidden           = errors.New("permission denied")
    ErrDuplicateNumber     = errors.New("contract number already exists")
)

// BusinessError carries HTTP status + client-facing message
type BusinessError struct {
    Code    string
    Message string
    Status  int
    Err     error
}

func (e *BusinessError) Error() string { return e.Message }
func (e *BusinessError) Unwrap() error { return e.Err }
```

## Repository: Map DB Errors

```go
// internal/core/contracts/repository/errors.go
func parseContractDBError(err error, op string) error {
    if err == nil {
        return nil
    }
    if errors.Is(err, pgx.ErrNoRows) {
        return domain.ErrContractNotFound
    }
    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) {
        switch pgErr.Code {
        case "23505":   // unique_violation
            return &domain.BusinessError{
                Code:    "DUPLICATE_CONTRACT_NUMBER",
                Message: "contract number already exists",
                Status:  409,
                Err:     domain.ErrDuplicateNumber,
            }
        case "23514":   // check_violation
            return &domain.BusinessError{
                Code:    "VALIDATION_ERROR",
                Message: "data validation failed: " + pgErr.ConstraintName,
                Status:  422,
            }
        }
    }
    return fmt.Errorf("%s: %w", op, err)
}
```

## Service: Wrap with Context

```go
func (s *ContractService) Create(ctx context.Context, sess ResolvedSession, req CreateParams) (*Contract, error) {
    contract, err := s.repo.Create(ctx, sess.TenantID, req)
    if err != nil {
        // Don't re-wrap BusinessError — let it propagate
        return nil, err
    }
    return contract, nil
}

func (s *ContractService) Submit(ctx context.Context, sess ResolvedSession, id uuid.UUID, version int) error {
    contract, err := s.repo.GetByID(ctx, id, sess.TenantID)
    if err != nil {
        return err  // ErrContractNotFound propagates unchanged
    }
    if contract.Status != domain.StatusDraft {
        return domain.ErrContractNotEditable  // business rule sentinel
    }
    // ...
}
```

## Handler: Map to HTTP

```go
// internal/core/contracts/handler/errors.go
func mapError(err error) *fiber.Error {
    // Check sentinel errors first
    if errors.Is(err, domain.ErrContractNotFound) {
        return fiber.NewError(404, "contract not found")
    }
    if errors.Is(err, domain.ErrVersionConflict) {
        return fiber.NewError(409, "version conflict — re-fetch and retry")
    }
    if errors.Is(err, domain.ErrContractNotEditable) {
        return fiber.NewError(422, "contract is not editable in its current state")
    }
    if errors.Is(err, domain.ErrForbidden) {
        return fiber.NewError(403, "permission denied")
    }

    // Check BusinessError (carries its own status)
    var bizErr *domain.BusinessError
    if errors.As(err, &bizErr) {
        return fiber.NewError(bizErr.Status, bizErr.Message)
    }

    // Unexpected error — log and return 500
    return fiber.NewError(500, "internal server error")
}
```

**Always use `errors.Is` and `errors.As`** — never type-switch on errors directly. Error wrapping (`%w`) must be preserved for `errors.Is`/`errors.As` to work through the chain.

## Handler: Use the Mapper

```go
func (h *ContractHandler) Submit(c *fiber.Ctx) error {
    sess := middleware.SessionFrom(c)
    id, err := parseUUID(c.Params("id"))
    if err != nil {
        return fiber.NewError(400, "invalid contract id")
    }

    var req SubmitRequest
    if err := c.BodyParser(&req); err != nil {
        return fiber.NewError(400, "invalid request body")
    }

    if err := h.svc.Submit(c.UserContext(), sess, id, req.Version); err != nil {
        return mapError(err)   // convert domain error → HTTP error
    }

    return c.SendStatus(fiber.StatusNoContent)
}
```

## Fiber Global Error Handler

Fiber's error handler renders the final response:

```go
// internal/server/fiber.go
func customErrorHandler(c *fiber.Ctx, err error) error {
    code := fiber.StatusInternalServerError
    msg := "internal server error"

    var fiberErr *fiber.Error
    if errors.As(err, &fiberErr) {
        code = fiberErr.Code
        msg = fiberErr.Message
    }

    return c.Status(code).JSON(fiber.Map{
        "error": fiber.Map{
            "code":    httpCodeToErrorCode(code),
            "message": msg,
        },
    })
}
```

## Do Not

```go
// ❌ Type switch loses wrapped errors
switch err.(type) {
case *domain.BusinessError:
    // ...
}

// ❌ Swallowing errors
result, _ := s.repo.Create(ctx, req)   // Never ignore errors

// ❌ Re-wrapping a BusinessError loses the original status
return fmt.Errorf("create failed: %w", bizErr)   // OK to wrap, but don't create a new BusinessError

// ❌ Returning raw pgx errors to the handler
return nil, pgErr    // Always map at the repo layer

// ✅ Correct
if errors.Is(err, domain.ErrContractNotFound) { ... }
var bizErr *domain.BusinessError
if errors.As(err, &bizErr) { ... }
```

## Validation Errors

Return field-level errors for 422:

```go
type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

// In handler, after body parsing
if errs := validate(req); len(errs) > 0 {
    return c.Status(422).JSON(fiber.Map{
        "error": fiber.Map{
            "code":    "VALIDATION_ERROR",
            "message": "Validation failed",
            "details": errs,
        },
    })
}
```
