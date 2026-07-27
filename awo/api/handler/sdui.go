package handler

import (
	"github.com/gofiber/fiber/v2"

	"awo.so/awo/sdui"
)

// SDUINav returns a Fiber handler that serves the sidebar navigation schema.
// GET /api/sdui/nav — returns []sdui.NavModule as JSON.
func SDUINav(gen *sdui.Generator) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(gen.Nav())
	}
}
