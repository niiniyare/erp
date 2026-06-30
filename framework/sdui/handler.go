// Package sdui exposes HTTP endpoints that serve generated AMIS page schemas
// for registered EntityDefinitions.
package sdui

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"awo.so/framework/def"
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
type CustomFieldSource func(ctx context.Context, tenantID uuid.UUID, entity string) ([]*def.FieldDef, error)

// TenantFromCtx extracts the tenant UUID from a Fiber request context.
type TenantFromCtx func(c *fiber.Ctx) (uuid.UUID, error)

// ViewerFromCtx extracts the viewer from the request context.
// When provided, schemas are permission-filtered per viewer role set.
type ViewerFromCtx func(c *fiber.Ctx) (def.ViewerContext, error)

// Handler serves SDUI schemas for all registered entities.
type Handler struct {
	apiBase      string
	customFields CustomFieldSource
	tenantFn     TenantFromCtx
	viewerFn     ViewerFromCtx
	cache        Cache
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

// WithCustomFields injects a custom-field source and tenant resolver.
func WithCustomFields(source CustomFieldSource, tenantFn TenantFromCtx) HandlerOption {
	return func(h *Handler) {
		h.customFields = source
		h.tenantFn = tenantFn
	}
}

// WithViewer injects a viewer resolver for permission-gated schema generation.
// Sensitive fields are excluded from schemas returned to viewers that lack admin roles.
// Routes return 401 when the viewer resolver returns an error.
func WithViewer(fn ViewerFromCtx) HandlerOption {
	return func(h *Handler) { h.viewerFn = fn }
}

// WithCache injects a Redis-backed schema cache (5-minute TTL).
// Schemas are cached tenant-wide (viewer filtering is applied after cache retrieval).
func WithCache(c Cache) HandlerOption {
	return func(h *Handler) { h.cache = c }
}

// serveNav returns the sidebar nav grouped by module.
func (h *Handler) serveNav(c *fiber.Ctx) error {
	defs := def.All()
	order := []string{}
	groups := map[string][]NavItem{}

	for _, d := range defs {
		if d.IsGlobal() {
			continue
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

// servePage serves the CRUDPage schema for the entity.
func (h *Handler) servePage(c *fiber.Ctx) error {
	return h.serveView(c, ViewCRUD)
}

// serveFormCreate serves the create form schema.
func (h *Handler) serveFormCreate(c *fiber.Ctx) error {
	return h.serveView(c, ViewFormCreate)
}

// serveFormEdit serves the edit form schema.
func (h *Handler) serveFormEdit(c *fiber.Ctx) error {
	return h.serveView(c, ViewFormEdit)
}

// serveDetail serves the read-only detail page schema.
func (h *Handler) serveDetail(c *fiber.Ctx) error {
	return h.serveView(c, ViewDetail)
}

// serveView is the shared handler for all schema view types.
func (h *Handler) serveView(c *fiber.Ctx, viewType ViewType) error {
	entityName := c.Params("entity")
	entDef := def.Lookup(entityName)
	if entDef == nil {
		return fiber.NewError(fiber.StatusNotFound, "unknown entity: "+entityName)
	}

	// Require authenticated viewer when viewerFn is configured.
	var viewer def.ViewerContext
	if h.viewerFn != nil {
		v, err := h.viewerFn(c)
		if err != nil {
			return fiber.ErrUnauthorized
		}
		viewer = v
	}

	// Resolve tenant for cache key and custom fields.
	tenantID := uuid.Nil
	if h.tenantFn != nil {
		if tid, err := h.tenantFn(c); err == nil {
			tenantID = tid
		}
	}

	// Resolve extra (tenant custom) fields.
	var extraFields []*def.FieldDef
	if h.customFields != nil && tenantID != uuid.Nil {
		if extra, err := h.customFields(c.Context(), tenantID, entDef.Name); err == nil {
			extraFields = extra
		}
	}

	b := NewBuilder(entDef, h.apiBase,
		WithBuilderCache(h.cache),
		WithExtraFields(extraFields),
	)

	schema, err := b.Build(c.Context(), viewType, viewer, tenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(schema)
}
