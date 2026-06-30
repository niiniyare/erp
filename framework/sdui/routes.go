package sdui

import (
	"github.com/gofiber/fiber/v2"
)

// Register mounts both the current and legacy SDUI routes on router.
//
// Current routes (documented):
//
//	GET /api/v1/pages/nav
//	GET /api/v1/pages/:entity              → CRUDPage schema
//	GET /api/v1/pages/:entity/form         → create form schema
//	GET /api/v1/pages/:entity/form/edit    → edit form schema
//	GET /api/v1/pages/:entity/detail       → read-only detail schema
//
// Legacy routes (deprecated — Deprecation header set):
//
//	GET /sdui/nav
//	GET /sdui/:entity
//	GET /sdui/:entity/form
func Register(router fiber.Router, apiBase string, opts ...HandlerOption) {
	h := NewHandler(apiBase, opts...)

	// ── Current routes (/api/v1/pages/) ───────────────────────────────────────
	pages := router.Group("/api/v1/pages")
	pages.Get("/nav", h.serveNav)
	pages.Get("/:entity", h.servePage)
	pages.Get("/:entity/form", h.serveFormCreate)
	pages.Get("/:entity/form/edit", h.serveFormEdit)
	pages.Get("/:entity/detail", h.serveDetail)

	// ── Legacy routes (/sdui/) — deprecated ───────────────────────────────────
	// These routes remain for backward compatibility. Clients should migrate to
	// /api/v1/pages/ at the next major version boundary.
	deprecationHeader := func(c *fiber.Ctx) error {
		c.Set("Deprecation", "true")
		c.Set("Link", `</api/v1/pages>; rel="successor-version"`)
		return c.Next()
	}

	legacy := router.Group("/sdui")
	legacy.Use(deprecationHeader)
	legacy.Get("/nav", h.serveNav)
	legacy.Get("/:entity", h.servePage)
	legacy.Get("/:entity/form", h.serveFormCreate)
}
