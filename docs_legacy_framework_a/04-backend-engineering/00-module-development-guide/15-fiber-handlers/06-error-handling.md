> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Handler Error Handling
portal: 4 — Backend Engineering
section: 00-module-development-guide/15-fiber-handlers
audience: [backend-engineer, tech-lead]
related:
  - path: ./02-handler-struct.md
    title: Handler Struct
  - path: ../05-repository-layer/04-error-mapping.md
    title: Error Mapping
---

# Handler Error Handling

`errors.go` translates domain errors to HTTP responses. Every handler uses `mapError(err)` as the last step before returning an error.

## errors.go

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
// Always use errors.Is — never == or type switch — to traverse error chains.
func mapError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	// Not Found
	case errors.Is(err, domain.ErrContractNotFound):
		return fiber.NewError(fiber.StatusNotFound, "contract not found")
	case errors.Is(err, domain.ErrContractLineNotFound):
		return fiber.NewError(fiber.StatusNotFound, "contract line not found")

	// Conflict — unique key / concurrent write
	case errors.Is(err, domain.ErrContractAlreadyExists):
		return fiber.NewError(fiber.StatusConflict, "a contract with this number already exists")
	case errors.Is(err, domain.ErrContractConflict):
		return fiber.NewError(fiber.StatusConflict, "contract was modified by another user; please reload and retry")

	// Unprocessable — business rule violation / invalid state
	case errors.Is(err, domain.ErrContractInvalidTransition):
		return fiber.NewError(fiber.StatusUnprocessableEntity, "this status change is not allowed in the current state")
	case errors.Is(err, domain.ErrContractNotEditable):
		return fiber.NewError(fiber.StatusUnprocessableEntity, "contract cannot be edited in its current status")
	case errors.Is(err, domain.ErrContractValueExceedsLimit):
		return fiber.NewError(fiber.StatusUnprocessableEntity, "contract value exceeds the approval threshold; escalated approval required")

	// Auth
	case errors.Is(err, iam.ErrForbidden):
		return fiber.NewError(fiber.StatusForbidden, "you do not have permission to perform this action")
	case errors.Is(err, iam.ErrUnauthorized):
		return fiber.NewError(fiber.StatusUnauthorized, "authentication required")

	// Default: internal server error — do not leak error details to client
	default:
		return fiber.NewError(fiber.StatusInternalServerError, "an unexpected error occurred")
	}
}
```

## Global Error Handler

The application registers a global Fiber error handler that serialises `*fiber.Error` to JSON:

```go
// app.go
app.Use(func(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "internal server error"

	var fe *fiber.Error
	if errors.As(err, &fe) {
		code = fe.Code
		message = fe.Message
	}

	return c.Status(code).JSON(fiber.Map{
		"error": message,
		"code":  code,
	})
})
```

`mapError` returns `*fiber.Error`. The global handler serialises it. No module needs its own error response formatting.

## HTTP Status Code Reference

| Code | Meaning | When to use |
|------|---------|------------|
| `200` | OK | Successful GET, PUT, PATCH |
| `201` | Created | Successful POST that creates a resource |
| `204` | No Content | Successful DELETE |
| `400` | Bad Request | Malformed input (parse errors, invalid UUID) |
| `401` | Unauthorized | No valid session / token expired |
| `403` | Forbidden | Valid session, insufficient permission |
| `404` | Not Found | Resource doesn't exist or is deleted |
| `409` | Conflict | Duplicate key or optimistic lock failure |
| `422` | Unprocessable Entity | Valid format but business rule violation |
| `500` | Internal Server Error | Unexpected failures |

## Client-Facing Error Messages

Error messages in `mapError` must be:
- Human-readable and actionable
- Free of internal system details (no stack traces, no SQL error codes)
- Safe to log by clients

```go
// CORRECT — actionable, no internal details
fiber.NewError(fiber.StatusConflict, "contract was modified by another user; please reload and retry")

// WRONG — leaks internal error
fiber.NewError(fiber.StatusInternalServerError, err.Error())

// WRONG — too vague
fiber.NewError(fiber.StatusConflict, "conflict")
```

Never call `err.Error()` in a fiber.NewError message — the error chain may contain SQL errors, internal IDs, or other sensitive implementation details.
