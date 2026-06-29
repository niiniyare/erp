// Package sdui exposes an HTTP endpoint that serves generated AMIS schemas
// for registered EntityDefinitions.
package sdui

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"awo.so/framework/definition"
	"awo.so/framework/sdui/amis"
)

// NavItem is one entry in the navigation tree returned by GET /sdui/nav.
type NavItem struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Module string `json:"module,omitempty"`
	Icon   string `json:"icon,omitempty"`
}

// NavGroup is a labelled section in the sidebar nav.
type NavGroup struct {
	Group string    `json:"group"`
	Items []NavItem `json:"items"`
}

// CustomFieldSource resolves tenant-specific custom fields for an entity.
// Typically backed by customfield.Registry.Get.
type CustomFieldSource func(ctx context.Context, tenantID uuid.UUID, entity string) ([]*definition.FieldDef, error)

// TenantFromCtx extracts the tenant UUID from a Fiber request context.
// Mirrors api.ViewerFromCtx — host registers middleware that calls c.Locals("viewer", …).
type TenantFromCtx func(c *fiber.Ctx) (uuid.UUID, error)

// Handler serves SDUI schemas, optionally enriched with per-tenant custom fields.
type Handler struct {
	apiBase      string
	customFields CustomFieldSource
	tenantFn     TenantFromCtx
}

// NewHandler creates a Handler.
func NewHandler(apiBase string, opts ...HandlerOption) *Handler {
	h := &Handler{apiBase: apiBase}
	for _, o := range opts {
		o(h)
	}
	return h
}

// HandlerOption configures a Handler.
type HandlerOption func(*Handler)

// WithCustomFields injects a custom-field source and a function to extract the
// tenant ID from the request context. Both must be provided together.
func WithCustomFields(source CustomFieldSource, tenantFn TenantFromCtx) HandlerOption {
	return func(h *Handler) {
		h.customFields = source
		h.tenantFn = tenantFn
	}
}

// Register mounts the SDUI schema endpoint on router.
//
//	GET /sdui/nav            → nav tree for all registered entities
//	GET /sdui/:entity        → CRUDPage schema
//	GET /sdui/:entity/form   → FormPage schema
func Register(router fiber.Router, apiBase string, opts ...HandlerOption) {
	h := NewHandler(apiBase, opts...)
	g := router.Group("/sdui")
	g.Get("/nav", func(c *fiber.Ctx) error {
		return serveNav(c)
	})
	g.Get("/:entity", func(c *fiber.Ctx) error {
		return h.servePage(c, false)
	})
	g.Get("/:entity/form", func(c *fiber.Ctx) error {
		return h.servePage(c, true)
	})
}

// serveNav returns the sidebar navigation structure derived from all registered
// EntityDefinitions, grouped by EntityDefinition.Module.
func serveNav(c *fiber.Ctx) error {
	defs := definition.All()

	// Group by module, preserving first-seen order.
	order := []string{}
	groups := map[string][]NavItem{}

	for _, d := range defs {
		if d.IsGlobal() {
			continue // global entities are not user-navigable by default
		}
		mod := d.Module
		if mod == "" {
			mod = "Other"
		}
		if _, ok := groups[mod]; !ok {
			order = append(order, mod)
		}
		label := d.Label
		if label == "" {
			label = d.Name
		}
		groups[mod] = append(groups[mod], NavItem{
			ID:     d.Name,
			Label:  label,
			Module: mod,
		})
	}

	nav := make([]NavGroup, 0, len(order))
	for _, mod := range order {
		nav = append(nav, NavGroup{Group: mod, Items: groups[mod]})
	}

	return c.JSON(nav)
}

func (h *Handler) servePage(c *fiber.Ctx, formOnly bool) error {
	name := c.Params("entity")
	def := definition.Lookup(name)
	if def == nil {
		return fiber.NewError(fiber.StatusNotFound, "unknown entity: "+name)
	}

	opts := amis.PageOpts{APIBase: h.apiBase}

	// Enrich with tenant-specific custom fields when a source is configured.
	if h.customFields != nil && h.tenantFn != nil {
		tenantID, err := h.tenantFn(c)
		if err == nil && tenantID != uuid.Nil {
			extra, err := h.customFields(c.Context(), tenantID, def.Name)
			if err == nil {
				opts.ExtraFields = extra
			}
			// Non-fatal: serve base schema on registry error.
		}
	}

	var schema map[string]any
	if formOnly {
		schema = amis.FormPage(def, opts)
	} else {
		schema = amis.CRUDPage(def, opts)
	}

	return c.JSON(schema)
}
