---
title: Request and Response Design
portal: 4 — Backend Engineering
section: 00-module-development-guide/14-api-design
audience: [backend-engineer, tech-lead]
related:
  - path: ./01-api-design-overview.md
    title: API Design Overview
  - path: ../15-fiber-handlers/02-handler-implementation.md
    title: Handler Implementation
---

# Request and Response Design

## Request DTO Design

Request DTOs use snake_case JSON keys, validate tags, and explicit pointer types for optional fields:

```go
// internal/core/contracts/handler/dto/create_contract_request.go
package dto

import "github.com/google/uuid"

type CreateContractRequest struct {
    ContractNumber string  `json:"contract_number" validate:"required,min=3,max=50"`
    Title          string  `json:"title"           validate:"required,min=1,max=255"`
    VendorID       string  `json:"vendor_id"        validate:"required,uuid"`
    ContractType   string  `json:"contract_type"    validate:"required,oneof=service goods lease other"`
    StartDate      string  `json:"start_date"       validate:"required,datetime=2006-01-02"`
    EndDate        string  `json:"end_date"         validate:"required,datetime=2006-01-02"`
    TotalValue     string  `json:"total_value"      validate:"required,numeric"`
    Currency       string  `json:"currency"         validate:"required,len=3"`
    Description    *string `json:"description"`     // optional
}
```

## Response DTO Design

Response DTOs serialize domain types to JSON-safe types:

- `uuid.UUID` → `string`
- `decimal.Decimal` → `string` (never float — precision loss)
- `time.Time` → RFC3339 string
- `*time.Time` (nullable) → string or `null`
- Enum types → string

```go
// internal/core/contracts/handler/dto/contract_response.go
package dto

type ContractResponse struct {
    ID             string  `json:"id"`
    TenantID       string  `json:"tenant_id"`
    EntityID       string  `json:"entity_id"`
    ContractNumber string  `json:"contract_number"`
    Title          string  `json:"title"`
    Status         string  `json:"status"`
    ContractType   string  `json:"contract_type"`
    TotalValue     string  `json:"total_value"`  // decimal string
    Currency       string  `json:"currency"`
    StartDate      string  `json:"start_date"`   // YYYY-MM-DD
    EndDate        string  `json:"end_date"`
    VendorID       string  `json:"vendor_id"`
    CreatedBy      string  `json:"created_by"`
    UpdatedBy      string  `json:"updated_by"`
    Version        int     `json:"version"`
    CreatedAt      string  `json:"created_at"`   // RFC3339
    UpdatedAt      string  `json:"updated_at"`
    DeletedAt      *string `json:"deleted_at"`   // null if not deleted
}

func ContractToResponse(c *domain.Contract) ContractResponse {
    r := ContractResponse{
        ID:             c.ID.String(),
        TenantID:       c.TenantID.String(),
        ContractNumber: c.ContractNumber,
        Title:          c.Title,
        Status:         string(c.Status),
        ContractType:   string(c.ContractType),
        TotalValue:     c.TotalValue.String(),
        Currency:       c.Currency,
        StartDate:      c.StartDate.Format("2006-01-02"),
        EndDate:        c.EndDate.Format("2006-01-02"),
        VendorID:       c.VendorID.String(),
        CreatedBy:      c.CreatedBy.String(),
        UpdatedBy:      c.UpdatedBy.String(),
        Version:        c.Version,
        CreatedAt:      c.CreatedAt.UTC().Format(time.RFC3339),
        UpdatedAt:      c.UpdatedAt.UTC().Format(time.RFC3339),
    }
    if c.DeletedAt != nil {
        s := c.DeletedAt.UTC().Format(time.RFC3339)
        r.DeletedAt = &s
    }
    return r
}
```

## List Response Envelope

List responses include pagination metadata:

```go
type ListContractsResponse struct {
    Data       []ContractResponse `json:"data"`
    Pagination PaginationMeta     `json:"pagination"`
}

type PaginationMeta struct {
    Page      int `json:"page"`
    PageSize  int `json:"page_size"`
    Total     int `json:"total"`
    TotalPages int `json:"total_pages"`
}
```

```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 87,
    "total_pages": 5
  }
}
```

## Create Response

201 Created with `Location` header:

```go
func (h *contractHandler) Create(c *fiber.Ctx) error {
    // ...
    contract, err := h.svc.Create(c.Context(), /* ... */)
    if err != nil {
        return h.mapError(err)
    }

    c.Set("Location", "/api/v1/contracts/"+contract.ID.String())
    return c.Status(fiber.StatusCreated).JSON(dto.ContractToResponse(contract))
}
```

## Validation Error Response

422 with field-level details:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": [
      { "field": "contract_number", "message": "contract_number is required" },
      { "field": "start_date", "message": "must be YYYY-MM-DD format" }
    ]
  }
}
```

```go
// Validation in handler
func (h *contractHandler) Create(c *fiber.Ctx) error {
    var req dto.CreateContractRequest
    if err := c.BodyParser(&req); err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body")
    }

    if errs := h.validator.Struct(req); errs != nil {
        return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
            "error": fiber.Map{
                "code":    "VALIDATION_ERROR",
                "message": "Validation failed",
                "details": formatValidationErrors(errs),
            },
        })
    }
    // ...
}
```
