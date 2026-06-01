---
title: Handler Struct
portal: 4 — Backend Engineering
section: 00-module-development-guide/15-fiber-handlers
audience: [backend-engineer, tech-lead]
related:
  - path: ./01-handler-overview.md
    title: Handler Overview
  - path: ./05-routes.md
    title: Routes
---

# Handler Struct

`handler.go` contains the Handler struct, its constructor, and all handler methods. One file for the entire module's handlers.

## Complete handler.go

```go
// internal/api/handlers/contracts/handler.go
package contracts

import (
	"context"
	"math"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"

	iam "awo.so/internal/core/iam"
	"awo.so/internal/core/contracts/service"
	"awo.so/internal/core/contracts/repository"
	middlewarePkg "awo.so/internal/api/middleware"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/tracing"
)

// Handler handles all HTTP requests for the contracts module.
type Handler struct {
	service    service.ContractService
	authConfig *middlewarePkg.AuthConfig
	logger     logger.Logger
	tracer     tracing.Service
}

// NewHandler constructs a contracts Handler.
func NewHandler(
	service    service.ContractService,
	authConfig *middlewarePkg.AuthConfig,
	logger     logger.Logger,
	tracer     tracing.Service,
) *Handler {
	return &Handler{
		service:    service,
		authConfig: authConfig,
		logger:     logger.With().Str("handler", "contracts").Logger(),
		tracer:     tracer,
	}
}

// ============================================================
// List
// ============================================================

func (h *Handler) List(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "ContractsHandler.List")
	defer span.End()

	sess, ok := c.Locals(iam.LocalsKeySession).(*iam.ResolvedSession)
	if !ok || sess == nil {
		return fiber.ErrUnauthorized
	}

	// Parse filters from query params
	var entityID *uuid.UUID
	if raw := c.Query("entity_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err == nil {
			entityID = &id
		}
	}

	var status *string
	if s := c.Query("status"); s != "" {
		status = &s
	}

	var search *string
	if q := c.Query("q"); q != "" {
		search = &q
	}

	pageSize := c.QueryInt("page_size", 20)
	if pageSize > 100 { pageSize = 100 }
	if pageSize < 1  { pageSize = 1  }
	pageOffset := c.QueryInt("page_offset", 0)
	if pageOffset < 0 { pageOffset = 0 }

	params := repository.ListContractsParams{
		TenantID:   sess.TenantID,
		EntityID:   entityID,
		Status:     (*contractStatus)(status),
		Search:     search,
		SortBy:     c.Query("sort_by", "created_at"),
		SortDir:    c.Query("sort_dir", "desc"),
		PageSize:   pageSize,
		PageOffset: pageOffset,
	}

	result, err := h.service.List(ctx, params, sess.ToPrincipal())
	if err != nil {
		span.RecordError(err)
		return mapError(err)
	}

	return c.JSON(MapListToResponse(result))
}

// ============================================================
// GetByID
// ============================================================

func (h *Handler) GetByID(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "ContractsHandler.GetByID")
	defer span.End()

	sess, ok := c.Locals(iam.LocalsKeySession).(*iam.ResolvedSession)
	if !ok || sess == nil {
		return fiber.ErrUnauthorized
	}

	contractID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid contract id")
	}

	span.SetAttributes(attribute.String("contract.id", contractID.String()))

	contract, err := h.service.GetByID(ctx, contractID, sess.TenantID, sess.ToPrincipal())
	if err != nil {
		span.RecordError(err)
		return mapError(err)
	}

	return c.JSON(MapContractToResponse(contract))
}

// ============================================================
// Create
// ============================================================

func (h *Handler) Create(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "ContractsHandler.Create")
	defer span.End()

	sess, ok := c.Locals(iam.LocalsKeySession).(*iam.ResolvedSession)
	if !ok || sess == nil {
		return fiber.ErrUnauthorized
	}

	var req CreateContractRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if err := validate.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	contract, err := h.service.Create(ctx, service.CreateContractRequest{
		TenantID:       sess.TenantID,
		EntityID:       sess.EntityScopeEntityID(),
		ContractNumber: req.ContractNumber,
		Title:          req.Title,
		Description:    req.Description,
		VendorID:       req.VendorID,
		ContractType:   domain.ContractType(req.ContractType),
		StartDate:      req.StartDate,
		EndDate:        req.EndDate,
		Currency:       req.Currency,
		CreatedBy:      sess.UserID,
		Principal:      sess.ToPrincipal(),
	})
	if err != nil {
		span.RecordError(err)
		return mapError(err)
	}

	span.SetAttributes(attribute.String("contract.id", contract.ID.String()))
	return c.Status(fiber.StatusCreated).JSON(MapContractToResponse(contract))
}

// ============================================================
// Update
// ============================================================

func (h *Handler) Update(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "ContractsHandler.Update")
	defer span.End()

	sess, ok := c.Locals(iam.LocalsKeySession).(*iam.ResolvedSession)
	if !ok || sess == nil {
		return fiber.ErrUnauthorized
	}

	contractID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid contract id")
	}

	var req UpdateContractRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if err := validate.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	contract, err := h.service.Update(ctx, service.UpdateContractRequest{
		ID:           contractID,
		TenantID:     sess.TenantID,
		Title:        req.Title,
		Description:  req.Description,
		VendorID:     req.VendorID,
		ContractType: domain.ContractType(req.ContractType),
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		Currency:     req.Currency,
		Version:      req.Version,
		UpdatedBy:    sess.UserID,
		Principal:    sess.ToPrincipal(),
	})
	if err != nil {
		span.RecordError(err)
		return mapError(err)
	}

	return c.JSON(MapContractToResponse(contract))
}

// ============================================================
// Submit (status transition)
// ============================================================

func (h *Handler) Submit(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "ContractsHandler.Submit")
	defer span.End()

	sess, ok := c.Locals(iam.LocalsKeySession).(*iam.ResolvedSession)
	if !ok || sess == nil {
		return fiber.ErrUnauthorized
	}

	contractID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid contract id")
	}

	var req VersionedRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	contract, err := h.service.Submit(ctx, contractID, sess.TenantID, req.Version, sess.UserID, sess.ToPrincipal())
	if err != nil {
		span.RecordError(err)
		return mapError(err)
	}

	return c.JSON(MapContractToResponse(contract))
}

// ============================================================
// Delete
// ============================================================

func (h *Handler) Delete(c *fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "ContractsHandler.Delete")
	defer span.End()

	sess, ok := c.Locals(iam.LocalsKeySession).(*iam.ResolvedSession)
	if !ok || sess == nil {
		return fiber.ErrUnauthorized
	}

	contractID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid contract id")
	}

	version := c.QueryInt("version", 0)
	if version <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "version query parameter is required")
	}

	if err := h.service.Delete(ctx, contractID, sess.TenantID, version, sess.UserID, sess.ToPrincipal()); err != nil {
		span.RecordError(err)
		return mapError(err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ============================================================
// Helper
// ============================================================

// contractStatus is a local type alias for use in filter params.
type contractStatus = domain.ContractStatus

// validate is the shared validator instance.
var validate = validator.New()
```
