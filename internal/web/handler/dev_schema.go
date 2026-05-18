package handler

import (
	"github.com/gofiber/fiber/v2"

	"awo.so/internal/shared/logger"
	"awo.so/internal/web/registry"
	"awo.so/internal/web/ui"
)

// DevSchemaHandler serves AMIS page schemas directly from the registry,
// without authentication or the UI pipeline. Intended for the standalone UI
// dev server (cmd/ui) and for API server routes where the pipeline has not
// yet been wired.
//
// Schemas are built with a zero-value UISessionContext — suitable for
// static/layout-only schemas that do not gate content behind permissions.
type DevSchemaHandler struct {
	log logger.Logger
}

// NewDevSchemaHandler constructs a DevSchemaHandler. log is optional.
func NewDevSchemaHandler(log logger.Logger) *DevSchemaHandler {
	return &DevSchemaHandler{log: log}
}

// Handle serves the schema for the requested path.
// Path: everything after /schema, e.g. /schema/finance/invoices → /finance/invoices
func (h *DevSchemaHandler) Handle(c *fiber.Ctx) error {
	route := "/" + c.Params("*")

	fn := registry.Get(route)
	if fn == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status": 404,
			"msg":    "schema not found: " + route,
		})
	}

	schema := fn(ui.UISessionContext{})
	return c.JSON(fiber.Map{
		"status": 0,
		"data":   schema,
	})
}
