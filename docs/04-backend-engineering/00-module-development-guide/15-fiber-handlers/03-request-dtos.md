---
title: Request DTOs
portal: 4 — Backend Engineering
section: 00-module-development-guide/15-fiber-handlers
audience: [backend-engineer, tech-lead]
related:
  - path: ./02-handler-struct.md
    title: Handler Struct
  - path: ./04-response-dtos.md
    title: Response DTOs
---

# Request DTOs

`request.go` declares the structs that Fiber parses from request bodies. Every field has a `json` tag and a `validate` tag. Request DTOs never cross the handler boundary — they are always converted to service request types before calling the service.

## request.go

```go
// internal/api/handlers/contracts/request.go
package contracts

import "github.com/google/uuid"

// CreateContractRequest is the request body for POST /contracts
type CreateContractRequest struct {
	ContractNumber string    `json:"contract_number" validate:"required,min=3,max=50"`
	Title          string    `json:"title"           validate:"required,min=1,max=255"`
	Description    string    `json:"description"     validate:"max=2000"`
	VendorID       uuid.UUID `json:"vendor_id"       validate:"required"`
	ContractType   string    `json:"contract_type"   validate:"required,oneof=service supply framework lease"`
	StartDate      string    `json:"start_date"      validate:"required,datetime=2006-01-02"`
	EndDate        string    `json:"end_date"        validate:"required,datetime=2006-01-02"`
	Currency       string    `json:"currency"        validate:"required,len=3"`
}

// UpdateContractRequest is the request body for PUT /contracts/:id
type UpdateContractRequest struct {
	Title        string    `json:"title"          validate:"required,min=1,max=255"`
	Description  string    `json:"description"    validate:"max=2000"`
	VendorID     uuid.UUID `json:"vendor_id"      validate:"required"`
	ContractType string    `json:"contract_type"  validate:"required,oneof=service supply framework lease"`
	StartDate    string    `json:"start_date"     validate:"required,datetime=2006-01-02"`
	EndDate      string    `json:"end_date"       validate:"required,datetime=2006-01-02"`
	Currency     string    `json:"currency"       validate:"required,len=3"`
	Version      int       `json:"version"        validate:"required,min=1"`
}

// VersionedRequest is used for status-transition endpoints that only need a version.
type VersionedRequest struct {
	Version int `json:"version" validate:"required,min=1"`
}

// TerminateContractRequest includes a reason for termination.
type TerminateContractRequest struct {
	Version int    `json:"version" validate:"required,min=1"`
	Reason  string `json:"reason"  validate:"required,min=5,max=500"`
}

// AddContractLineRequest is the request body for POST /contracts/:id/lines
type AddContractLineRequest struct {
	Description string `json:"description" validate:"required,min=1,max=500"`
	Quantity    int    `json:"quantity"    validate:"required,min=1"`
	UnitPrice   string `json:"unit_price"  validate:"required"`  // decimal string
}

// UpdateContractLineRequest is the request body for PUT /contracts/:id/lines/:lineId
type UpdateContractLineRequest struct {
	Description string `json:"description" validate:"required,min=1,max=500"`
	Quantity    int    `json:"quantity"    validate:"required,min=1"`
	UnitPrice   string `json:"unit_price"  validate:"required"`
	Version     int    `json:"version"     validate:"required,min=1"`
}
```

## Validate Tags Reference

| Tag | Meaning |
|-----|---------|
| `required` | Field must be present and non-zero |
| `min=N` | Minimum string length or numeric value |
| `max=N` | Maximum string length or numeric value |
| `len=N` | Exact string length (for ISO codes) |
| `oneof=a b c` | Must be one of the listed values |
| `datetime=layout` | Must parse as the given time layout |
| `uuid` | Must be a valid UUID |
| `email` | Must be a valid email address |

## Validation Error Response

When validation fails, return a structured error message:

```go
if err := validate.Struct(req); err != nil {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		fields := make([]map[string]string, 0, len(validationErrors))
		for _, fe := range validationErrors {
			fields = append(fields, map[string]string{
				"field":   fe.Field(),
				"message": validationMessage(fe),
			})
		}
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error":  "validation failed",
			"fields": fields,
		})
	}
	return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
}
```

## What NOT to Put in Request DTOs

| Do not | Reason |
|--------|--------|
| `tenant_id` field | Always from session |
| `entity_id` field (for create) | Always from session scope |
| `created_by` / `updated_by` | Always from session |
| Status fields (for create) | Always `draft` — not a request field |
| `id` in request body | Use URL param `:id` |
| UUID strings (prefer `uuid.UUID`) | Parse at DTO level for early validation |
