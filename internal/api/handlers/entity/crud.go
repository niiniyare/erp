package entity

// crud.go — CRUD and hierarchy handler methods.
//
// Routes (registered in routes.go):
//
//	POST   /api/v1/entities               → Create
//	GET    /api/v1/entities               → List
//	GET    /api/v1/entities/:id           → GetByID
//	PUT    /api/v1/entities/:id           → Update
//	DELETE /api/v1/entities/:id           → Delete
//	GET    /api/v1/entities/:id/children  → GetChildren
//	GET    /api/v1/entities/tree          → GetTree

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"

	entityDomain "awo.so/internal/core/entity"
	sharedErrors "awo.so/internal/shared/errors"
)

// Create creates a new entity.
//
// POST /api/v1/entities
func (h *EntityHandler) Create(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.UserContext(), "entity.Create")
	defer span.End()
	c.SetUserContext(ctx)

	var req entityDomain.CreateEntityRequest
	if err := h.bind(c, &req); err != nil {
		return h.fail(c, err)
	}

	ent, err := h.service.CreateEntity(ctx, req)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	span.SetAttributes(
		attribute.String("entity.id", ent.ID.String()),
		attribute.String("entity.code", ent.Code),
		attribute.String("entity.type", string(ent.Type)),
	)
	return h.ok201(c, ent)
}

// List returns entities with optional filters.
//
// GET /api/v1/entities
func (h *EntityHandler) List(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.UserContext(), "entity.List")
	defer span.End()
	c.SetUserContext(ctx)

	offset, limit := h.pagination(c)

	req := entityDomain.ListEntitiesRequest{
		Offset: offset,
		Limit:  limit,
	}

	// Optional filters
	if t := c.Query("type"); t != "" {
		et := entityDomain.EntityType(t)
		req.Type = &et
	}
	if v := c.Query("is_active"); v == "true" {
		b := true
		req.IsActive = &b
	} else if v == "false" {
		b := false
		req.IsActive = &b
	}
	if v := c.Query("parent_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return h.fail(c, sharedErrors.NewBusinessError("INVALID_PARENT_ID", "parent_id must be a valid UUID").
				WithHTTPStatus(fiber.StatusBadRequest).
				WithCategory(sharedErrors.CategoryValidation))
		}
		req.ParentID = &id
	}

	entities, err := h.service.ListEntities(ctx, req)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	return h.ok200Meta(c, entities, map[string]any{
		"offset": offset,
		"limit":  limit,
		"count":  len(entities),
	})
}

// GetByID retrieves a single entity by UUID.
//
// GET /api/v1/entities/:id
func (h *EntityHandler) GetByID(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.UserContext(), "entity.GetByID")
	defer span.End()
	c.SetUserContext(ctx)

	id, err := parseUUID(c.Params("id"))
	if err != nil {
		return h.fail(c, err)
	}

	ent, err := h.service.GetEntityByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	return h.ok200(c, ent)
}

// Update updates an entity by UUID.
//
// PUT /api/v1/entities/:id
func (h *EntityHandler) Update(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.UserContext(), "entity.Update")
	defer span.End()
	c.SetUserContext(ctx)

	id, err := parseUUID(c.Params("id"))
	if err != nil {
		return h.fail(c, err)
	}

	var req entityDomain.UpdateEntityRequest
	if err := h.bind(c, &req); err != nil {
		return h.fail(c, err)
	}

	ent, err := h.service.UpdateEntity(ctx, id, req)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	span.SetAttributes(attribute.String("entity.id", id.String()))
	return h.ok200(c, ent)
}

// Delete soft-deletes an entity by UUID.
//
// DELETE /api/v1/entities/:id
func (h *EntityHandler) Delete(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.UserContext(), "entity.Delete")
	defer span.End()
	c.SetUserContext(ctx)

	id, err := parseUUID(c.Params("id"))
	if err != nil {
		return h.fail(c, err)
	}

	// ?permanent=true for hard delete (admin use)
	permanent := c.Query("permanent") == "true"

	if err := h.service.DeleteEntity(ctx, id, permanent); err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	return h.ok204(c)
}

// GetChildren returns direct children of an entity.
//
// GET /api/v1/entities/:id/children
func (h *EntityHandler) GetChildren(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.UserContext(), "entity.GetChildren")
	defer span.End()
	c.SetUserContext(ctx)

	id, err := parseUUID(c.Params("id"))
	if err != nil {
		return h.fail(c, err)
	}

	children, err := h.service.GetEntityChildren(ctx, id)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	return h.ok200(c, children)
}

// GetTree returns the full entity hierarchy tree.
//
// GET /api/v1/entities/tree
func (h *EntityHandler) GetTree(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.UserContext(), "entity.GetTree")
	defer span.End()
	c.SetUserContext(ctx)

	tree, err := h.service.GetEntityTree(ctx)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	return h.ok200(c, tree)
}

// ============================================================================
// Helpers
// ============================================================================

// parseUUID parses a UUID path param, returning a validation error on failure.
func parseUUID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, sharedErrors.NewBusinessError("INVALID_ID", "id must be a valid UUID").
			WithHTTPStatus(fiber.StatusBadRequest).
			WithCategory(sharedErrors.CategoryValidation)
	}
	return id, nil
}
