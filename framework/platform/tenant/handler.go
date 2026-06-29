package tenant

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Handler exposes tenant lifecycle endpoints.
// Mount via Register(router, svc).
type Handler struct {
	svc *Service
}

// Register mounts tenant routes on router.
//
//	POST   /tenants          — provision new tenant (pending)
//	GET    /tenants/:id      — get tenant by ID
//	GET    /tenants/slug/:s  — get tenant by slug
//	PATCH  /tenants/:id      — update tenant fields
//	POST   /tenants/:id/activate   — pending → active
//	POST   /tenants/:id/suspend    — active  → suspended
//	POST   /tenants/:id/archive    — any     → archived
func Register(router fiber.Router, svc *Service) {
	h := &Handler{svc: svc}
	g := router.Group("/tenants")
	g.Post("/", h.create)
	g.Get("/slug/:slug", h.getBySlug)
	g.Get("/:id", h.getByID)
	g.Patch("/:id", h.update)
	g.Post("/:id/activate", h.transition(StatusActive))
	g.Post("/:id/suspend", h.transition(StatusSuspended))
	g.Post("/:id/archive", h.transition(StatusArchived))
}

func (h *Handler) create(c *fiber.Ctx) error {
	var t Tenant
	if err := c.BodyParser(&t); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.Create(c.Context(), &t); err != nil {
		return mapErr(err)
	}
	return c.Status(http.StatusCreated).JSON(&t)
}

func (h *Handler) getByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "invalid uuid")
	}
	t, err := h.svc.store.GetByID(c.Context(), id)
	if err != nil {
		return mapErr(err)
	}
	return c.JSON(t)
}

func (h *Handler) getBySlug(c *fiber.Ctx) error {
	t, err := h.svc.store.GetBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		return mapErr(err)
	}
	return c.JSON(t)
}

func (h *Handler) update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "invalid uuid")
	}
	t, err := h.svc.store.GetByID(c.Context(), id)
	if err != nil {
		return mapErr(err)
	}
	if err := c.BodyParser(t); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}
	t.ID = id // guard against body override
	if err := h.svc.store.Update(c.Context(), t); err != nil {
		return mapErr(err)
	}
	return c.JSON(t)
}

func (h *Handler) transition(status string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := uuid.Parse(c.Params("id"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, "invalid uuid")
		}
		if err := h.svc.Transition(c.Context(), id, status); err != nil {
			return mapErr(err)
		}
		return c.SendStatus(http.StatusNoContent)
	}
}

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	switch err {
	case ErrInvalidTransition:
		return fiber.NewError(http.StatusUnprocessableEntity, err.Error())
	}
	return fiber.NewError(http.StatusInternalServerError, err.Error())
}
