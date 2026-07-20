> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Handler Layer Overview
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer]
related:
  - "[Service Layer](../06-service-layer/01-service-overview.md)"
  - "[API Design](../14-api-design/01-api-design-overview.md)"
  - "[Middleware Chain](../16-middleware-chain/01-middleware-overview.md)"
  - "[Worked Example: Handler Layer](../23-worked-example/07-handler-layer.md)"
---

# Handler Layer Overview

## Purpose

The handler layer:
- Parses HTTP request (path params, query params, body)
- Calls the service with parsed domain types
- Maps domain errors to HTTP errors
- Renders the HTTP response

Handlers contain **no business logic**. No permission checks. No DB calls. No business rules. Those belong in the service.

## Handler Struct

```go
// internal/core/contracts/handler/contract_handler.go
type ContractHandler struct {
    svc *service.ContractService
}

func NewContractHandler(svc *service.ContractService) *ContractHandler {
    return &ContractHandler{svc: svc}
}
```

## Method Template

```go
func (h *ContractHandler) Create(c *fiber.Ctx) error {
    // 1. Extract session (panics if missing — intentional)
    sess := middleware.SessionFrom(c)

    // 2. Parse and validate request
    var req CreateContractRequest
    if err := c.BodyParser(&req); err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
    }
    if errs := validate(req); len(errs) > 0 {
        return c.Status(422).JSON(fiber.Map{
            "error": fiber.Map{
                "code":    "VALIDATION_ERROR",
                "message": "Validation failed",
                "details": errs,
            },
        })
    }

    // 3. Call service
    contract, err := h.svc.Create(c.UserContext(), sess, service.CreateParams{
        ContractNumber: req.ContractNumber,
        Title:          req.Title,
        TotalValue:     req.TotalValue,
        Currency:       req.Currency,
        StartDate:      req.StartDate.Time,
        EndDate:        req.EndDate.Time,
    })
    if err != nil {
        return mapError(err)
    }

    // 4. Render response
    c.Set("Location", "/api/v1/contracts/"+contract.ID.String())
    return c.Status(fiber.StatusCreated).JSON(ContractToResponse(contract))
}
```

## Request DTOs

Define in the handler package:

```go
type CreateContractRequest struct {
    ContractNumber string          `json:"contract_number" validate:"required,max=50"`
    Title          string          `json:"title"           validate:"required,max=255"`
    VendorID       uuid.UUID       `json:"vendor_id"       validate:"required"`
    ContractType   string          `json:"contract_type"   validate:"required,oneof=service supply license framework"`
    TotalValue     decimal.Decimal `json:"total_value"     validate:"required"`
    Currency       string          `json:"currency"        validate:"required,len=3"`
    StartDate      civil.Date      `json:"start_date"      validate:"required"`
    EndDate        civil.Date      `json:"end_date"        validate:"required"`
    Description    string          `json:"description"`
}
```

DTOs live in the handler package — never in domain. Domain types don't know about JSON tags or HTTP.

## Response DTOs

```go
type ContractResponse struct {
    ID             uuid.UUID `json:"id"`
    ContractNumber string    `json:"contract_number"`
    Title          string    `json:"title"`
    Status         string    `json:"status"`
    ContractType   string    `json:"contract_type"`
    TotalValue     string    `json:"total_value"`  // decimal string, never float
    Currency       string    `json:"currency"`
    StartDate      string    `json:"start_date"`   // YYYY-MM-DD
    EndDate        string    `json:"end_date"`
    VendorID       uuid.UUID `json:"vendor_id"`
    Version        int       `json:"version"`
    CreatedAt      time.Time `json:"created_at"`
    UpdatedAt      time.Time `json:"updated_at"`
    DeletedAt      *time.Time `json:"deleted_at"`
}

func ContractToResponse(c *domain.Contract) ContractResponse {
    return ContractResponse{
        ID:             c.ID,
        ContractNumber: c.ContractNumber,
        Title:          c.Title,
        Status:         string(c.Status),
        ContractType:   string(c.ContractType),
        TotalValue:     c.TotalValue.String(),  // decimal.Decimal → string
        Currency:       c.Currency,
        StartDate:      c.StartDate.Format("2006-01-02"),
        EndDate:        c.EndDate.Format("2006-01-02"),
        VendorID:       c.VendorID,
        Version:        c.Version,
        CreatedAt:      c.CreatedAt,
        UpdatedAt:      c.UpdatedAt,
        DeletedAt:      c.DeletedAt,
    }
}
```

## Error Mapper

```go
// handler/errors.go
func mapError(err error) error {
    if errors.Is(err, domain.ErrContractNotFound) {
        return fiber.NewError(404, "contract not found")
    }
    if errors.Is(err, domain.ErrVersionConflict) {
        return fiber.NewError(409, "version conflict — re-fetch and retry")
    }
    if errors.Is(err, domain.ErrContractNotEditable) {
        return fiber.NewError(422, "contract cannot be modified in its current state")
    }
    if errors.Is(err, domain.ErrForbidden) {
        return fiber.NewError(403, "permission denied")
    }
    if errors.Is(err, domain.ErrDuplicateNumber) {
        return fiber.NewError(409, "contract number already exists")
    }
    var bizErr *domain.BusinessError
    if errors.As(err, &bizErr) {
        return fiber.NewError(bizErr.Status, bizErr.Message)
    }
    return fiber.NewError(500, "internal server error")
}
```

**Always use `errors.Is`/`errors.As`** — not type switch. Wrapping with `%w` must be preserved for unwrapping to work.

## UUID Parsing

```go
func parseUUID(s string) (uuid.UUID, error) {
    id, err := uuid.Parse(s)
    if err != nil {
        return uuid.Nil, fmt.Errorf("invalid uuid: %w", err)
    }
    return id, nil
}

// In handler:
id, err := parseUUID(c.Params("id"))
if err != nil {
    return fiber.NewError(400, "invalid id format")
}
```

## List Handler Pattern

```go
func (h *ContractHandler) List(c *fiber.Ctx) error {
    sess := middleware.SessionFrom(c)

    page := c.QueryInt("page", 1)
    pageSize := c.QueryInt("page_size", 20)
    if pageSize > 100 {
        pageSize = 100
    }

    status := c.Query("status")

    contracts, total, err := h.svc.List(c.UserContext(), sess, service.ListParams{
        Status: parseOptionalStatus(status),
        Limit:  int32(pageSize),
        Offset: int32((page - 1) * pageSize),
    })
    if err != nil {
        return mapError(err)
    }

    totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
    responses := make([]ContractResponse, len(contracts))
    for i, c := range contracts {
        responses[i] = ContractToResponse(&c)
    }

    return c.JSON(fiber.Map{
        "data": responses,
        "pagination": fiber.Map{
            "page":        page,
            "page_size":   pageSize,
            "total":       total,
            "total_pages": totalPages,
        },
    })
}
```

## Route Registration

```go
// internal/core/contracts/handler/routes.go
type ContractRouteRegistrar struct {
    handler *ContractHandler
    authz   authz.Service
}

func (r *ContractRouteRegistrar) Register(api fiber.Router, auth, authorize func(...) fiber.Handler) {
    contracts := api.Group("/contracts", auth())
    contracts.Get("",     authorize("contracts.contract.read"),   r.handler.List)
    contracts.Post("",    authorize("contracts.contract.create"), r.handler.Create)
    contracts.Get("/:id", authorize("contracts.contract.read"),   r.handler.GetByID)
    contracts.Put("/:id", authorize("contracts.contract.update"), r.handler.Update)
    contracts.Delete("/:id", authorize("contracts.contract.delete"), r.handler.Delete)

    contracts.Post("/:id/submit",    authorize("contracts.contract.submit"),    r.handler.Submit)
    contracts.Post("/:id/approve",   authorize("contracts.contract.approve"),   r.handler.Approve)
    contracts.Post("/:id/activate",  authorize("contracts.contract.activate"),  r.handler.Activate)
    contracts.Post("/:id/terminate", authorize("contracts.contract.terminate"), r.handler.Terminate)

    contracts.Post("/import", authorize("contracts.contract.create"), r.handler.Import)
    contracts.Get("/export",  authorize("contracts.contract.read"),   r.handler.Export)
}
```
