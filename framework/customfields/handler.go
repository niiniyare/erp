package customfields

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"awo.so/framework/definition"
)

// RegisterRoutes mounts CRUD routes for CustomField management under prefix.
//
//	GET    /prefix/:entity           → list custom fields for entity
//	POST   /prefix/:entity           → create custom field
//	PUT    /prefix/:entity/:id       → update custom field
//	DELETE /prefix/:entity/:id       → delete custom field
func RegisterRoutes(router fiber.Router, prefix string, store Store, viewerFn func(*fiber.Ctx) (string, error)) {
	g := router.Group(prefix)

	g.Get("/:entity", func(c *fiber.Ctx) error {
		tenantID, err := viewerFn(c)
		if err != nil {
			return fiber.ErrUnauthorized
		}
		entity := c.Params("entity")
		if definition.Lookup(entity) == nil {
			return fiber.NewError(fiber.StatusNotFound, "unknown entity: "+entity)
		}
		fields, err := store.ListForEntity(c.Context(), tenantID, entity)
		if err != nil {
			return err
		}
		return c.JSON(fields)
	})

	g.Post("/:entity", func(c *fiber.Ctx) error {
		tenantID, err := viewerFn(c)
		if err != nil {
			return fiber.ErrUnauthorized
		}
		entity := c.Params("entity")
		if definition.Lookup(entity) == nil {
			return fiber.NewError(fiber.StatusNotFound, "unknown entity: "+entity)
		}

		var cf CustomField
		if err := c.BodyParser(&cf); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid body")
		}
		cf.TenantID = tenantID
		cf.Entity = entity
		if cf.ID == "" {
			cf.ID = uuid.New().String()
		}

		if err := store.Save(c.Context(), &cf); err != nil {
			return err
		}
		return c.Status(fiber.StatusCreated).JSON(cf)
	})

	g.Put("/:entity/:id", func(c *fiber.Ctx) error {
		tenantID, err := viewerFn(c)
		if err != nil {
			return fiber.ErrUnauthorized
		}

		var cf CustomField
		if err := c.BodyParser(&cf); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid body")
		}
		cf.ID = c.Params("id")
		cf.TenantID = tenantID
		cf.Entity = c.Params("entity")

		if err := store.Save(c.Context(), &cf); err != nil {
			return err
		}
		return c.JSON(cf)
	})

	g.Delete("/:entity/:id", func(c *fiber.Ctx) error {
		_, err := viewerFn(c)
		if err != nil {
			return fiber.ErrUnauthorized
		}
		if err := store.Delete(c.Context(), c.Params("id")); err != nil {
			return err
		}
		return c.SendStatus(fiber.StatusNoContent)
	})
}
