package org

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// UnitStore is the persistence interface used by the HTTP handler.
type UnitStore interface {
	Create(c *fiber.Ctx, u *Unit) error
	GetByID(c *fiber.Ctx, tenantID, id uuid.UUID) (*Unit, error)
	List(c *fiber.Ctx, tenantID uuid.UUID) ([]*Unit, error)
	Update(c *fiber.Ctx, u *Unit) error
	Delete(c *fiber.Ctx, tenantID, id uuid.UUID) error
}

// TenantFromCtx extracts the tenant UUID from a request context.
type TenantFromCtx func(c *fiber.Ctx) (uuid.UUID, error)

// Handler serves org unit CRUD endpoints.
type Handler struct {
	store    UnitStore
	tenantFn TenantFromCtx
	tree     Tree
}

// Register mounts org unit routes.
//
//	POST   /org/units           — create unit
//	GET    /org/units           — list all units for tenant
//	GET    /org/units/:id       — get by ID
//	PATCH  /org/units/:id       — update name/code
//	DELETE /org/units/:id       — delete unit
func Register(router fiber.Router, store UnitStore, tenantFn TenantFromCtx, tree Tree) {
	h := &Handler{store: store, tenantFn: tenantFn, tree: tree}
	g := router.Group("/org/units")
	g.Post("/", h.create)
	g.Get("/", h.list)
	g.Get("/:id", h.getByID)
	g.Patch("/:id", h.update)
	g.Delete("/:id", h.delete)
}

func (h *Handler) tenantID(c *fiber.Ctx) (uuid.UUID, error) {
	id, err := h.tenantFn(c)
	if err != nil {
		return uuid.Nil, fiber.NewError(http.StatusUnauthorized, "tenant context missing")
	}
	return id, nil
}

func (h *Handler) create(c *fiber.Ctx) error {
	tenantID, err := h.tenantID(c)
	if err != nil {
		return err
	}
	var u Unit
	if err := c.BodyParser(&u); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}
	u.TenantID = tenantID
	if err := h.store.Create(c, &u); err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}
	if h.tree != nil {
		if err := h.tree.InsertPaths(c.Context(), &u); err != nil {
			return fiber.NewError(http.StatusInternalServerError, err.Error())
		}
	}
	return c.Status(http.StatusCreated).JSON(&u)
}

func (h *Handler) list(c *fiber.Ctx) error {
	tenantID, err := h.tenantID(c)
	if err != nil {
		return err
	}
	units, err := h.store.List(c, tenantID)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(units)
}

func (h *Handler) getByID(c *fiber.Ctx) error {
	tenantID, err := h.tenantID(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "invalid uuid")
	}
	u, err := h.store.GetByID(c, tenantID, id)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, err.Error())
	}
	return c.JSON(u)
}

func (h *Handler) update(c *fiber.Ctx) error {
	tenantID, err := h.tenantID(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "invalid uuid")
	}
	u, err := h.store.GetByID(c, tenantID, id)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, err.Error())
	}
	if err := c.BodyParser(u); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}
	u.ID = id
	u.TenantID = tenantID
	if err := h.store.Update(c, u); err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(u)
}

func (h *Handler) delete(c *fiber.Ctx) error {
	tenantID, err := h.tenantID(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "invalid uuid")
	}
	if err := h.store.Delete(c, tenantID, id); err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}
	return c.SendStatus(http.StatusNoContent)
}
