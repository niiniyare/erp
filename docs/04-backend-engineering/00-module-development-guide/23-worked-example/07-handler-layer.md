---
title: Worked Example — Handler Layer
portal: 4 — Backend Engineering
section: 00-module-development-guide/23-worked-example
audience: [backend-engineer, tech-lead]
related:
  - path: ../15-fiber-handlers/02-handler-implementation.md
    title: Handler Implementation
---

# Worked Example — Handler Layer

## Handler Struct

```go
// internal/core/contracts/handler/contract_handler.go
package handler

import (
    "github.com/go-playground/validator/v10"
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"

    "awo.so/internal/core/contracts/domain"
    "awo.so/internal/core/contracts/service"
    middlewarePkg "awo.so/internal/platform/middleware"
    iam "awo.so/internal/core/iam"
)

type contractHandler struct {
    svc       service.ContractService
    validator *validator.Validate
}

func NewContractHandler(svc service.ContractService) *contractHandler {
    return &contractHandler{
        svc:       svc,
        validator: validator.New(),
    }
}
```

## Create Handler

```go
func (h *contractHandler) Create(c *fiber.Ctx) error {
    session := middlewarePkg.SessionFrom(c)

    var req dto.CreateContractRequest
    if err := c.BodyParser(&req); err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body")
    }
    if errs := h.validator.Struct(req); errs != nil {
        return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
            "error": formatValidationErrors(errs),
        })
    }

    vendorID, err := uuid.Parse(req.VendorID)
    if err != nil {
        return fiber.NewError(fiber.StatusUnprocessableEntity, "vendor_id must be a valid UUID")
    }
    startDate, _ := time.Parse("2006-01-02", req.StartDate)
    endDate, _   := time.Parse("2006-01-02", req.EndDate)
    totalValue, _ := decimal.NewFromString(req.TotalValue)

    contract, err := h.svc.Create(c.Context(), service.CreateContractRequest{
        TenantID:       session.TenantID,
        EntityID:       session.EntityID(),
        UserID:         session.UserID,
        Principal:      session.ToPrincipal(),
        ContractNumber: req.ContractNumber,
        Title:          req.Title,
        Description:    derefString(req.Description),
        ContractType:   domain.ContractType(req.ContractType),
        TotalValue:     totalValue,
        Currency:       req.Currency,
        StartDate:      startDate,
        EndDate:        endDate,
        VendorID:       vendorID,
    })
    if err != nil {
        return h.mapError(err)
    }

    c.Set("Location", "/api/v1/contracts/"+contract.ID.String())
    return c.Status(fiber.StatusCreated).JSON(dto.ContractToResponse(contract))
}
```

## GetByID Handler

```go
func (h *contractHandler) GetByID(c *fiber.Ctx) error {
    session := middlewarePkg.SessionFrom(c)

    id, err := uuid.Parse(c.Params("id"))
    if err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "id must be a valid UUID")
    }

    contract, err := h.svc.GetByID(c.Context(), service.GetContractRequest{
        ID:        id,
        TenantID:  session.TenantID,
        Principal: session.ToPrincipal(),
    })
    if err != nil {
        return h.mapError(err)
    }

    return c.JSON(dto.ContractToResponse(contract))
}
```

## Submit Handler

```go
func (h *contractHandler) Submit(c *fiber.Ctx) error {
    session := middlewarePkg.SessionFrom(c)

    id, err := uuid.Parse(c.Params("id"))
    if err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "id must be a valid UUID")
    }

    var req dto.TransitionRequest
    if err := c.BodyParser(&req); err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body")
    }
    if errs := h.validator.Struct(req); errs != nil {
        return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
            "error": formatValidationErrors(errs),
        })
    }

    contract, err := h.svc.Submit(c.Context(), service.SubmitContractRequest{
        ContractID: id,
        TenantID:   session.TenantID,
        Version:    req.Version,
        UserID:     session.UserID,
        Principal:  session.ToPrincipal(),
    })
    if err != nil {
        return h.mapError(err)
    }

    return c.JSON(dto.ContractToResponse(contract))
}
```

## Error Mapping

```go
func (h *contractHandler) mapError(err error) error {
    switch {
    case errors.Is(err, domain.ErrContractNotFound):
        return fiber.NewError(fiber.StatusNotFound, "contract not found")
    case errors.Is(err, domain.ErrContractAlreadyExists):
        return fiber.NewError(fiber.StatusConflict, "contract number already exists")
    case errors.Is(err, domain.ErrContractConflict):
        return fiber.NewError(fiber.StatusConflict, "contract was modified by another request — retry with latest version")
    case errors.Is(err, domain.ErrContractNotEditable):
        return fiber.NewError(fiber.StatusUnprocessableEntity, "contract is not in an editable state")
    case errors.Is(err, domain.ErrContractInvalidTransition):
        return fiber.NewError(fiber.StatusUnprocessableEntity, "invalid status transition")
    case errors.Is(err, iam.ErrForbidden):
        return fiber.NewError(fiber.StatusForbidden, "insufficient permissions")
    default:
        return fiber.NewError(fiber.StatusInternalServerError, "an unexpected error occurred")
    }
}
```
