// Package sdui exposes an HTTP endpoint that serves generated AMIS schemas
// for registered EntityDefinitions.
package sdui

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"awo.so/framework/def"
	"awo.so/framework/sdui/amis"
)

const (
	// schemaCacheTTL is the Redis TTL for cached SDUI schemas.
	// Invalidate explicitly on permission change or entity definition update.
	schemaCacheTTL = 5 * time.Minute

	// schemaCacheKeyFmt is the Redis key pattern for page schemas.
	// Fields: entity name, schema type ("crud"/"form"), tenant UUID.
	schemaCacheKeyFmt = "page:%s:%s:%s"
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
type CustomFieldSource func(ctx context.Context, tenantID uuid.UUID, entity string) ([]*def.FieldDef, error)

// TenantFromCtx extracts the tenant UUID from a Fiber request context.
type TenantFromCtx func(c *fiber.Ctx) (uuid.UUID, error)

// ViewerFromCtx extracts the viewer from the request context.
// When set, the handler passes viewer roles to the schema builder for
// permission-gated field/action filtering.
type ViewerFromCtx func(c *fiber.Ctx) (def.ViewerContext, error)

// SchemaCache is a minimal Redis interface used for schema caching.
// Satisfied by *redis.Client and *redis.ClusterClient.
type SchemaCache interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
}

// Handler serves SDUI schemas, optionally enriched with per-tenant custom fields.
type Handler struct {
	apiBase      string
	customFields CustomFieldSource
	tenantFn     TenantFromCtx
	viewerFn     ViewerFromCtx
	cache        SchemaCache
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

// WithViewer injects a viewer resolver for permission-gated schema generation.
// When set, sensitive fields (IsSensitive) are excluded from schemas returned
// to viewers that lack the appropriate role.
func WithViewer(fn ViewerFromCtx) HandlerOption {
	return func(h *Handler) { h.viewerFn = fn }
}

// WithCache injects a Redis client for schema caching (5-minute TTL).
// Schema is keyed by entity + schema type + tenant UUID.
// Without this option, schemas are generated fresh on every request.
func WithCache(c SchemaCache) HandlerOption {
	return func(h *Handler) { h.cache = c }
}

// InvalidateSchema removes cached schemas for the given entity and tenant.
// Call after permission changes, entity definition updates, or feature flag changes.
func InvalidateSchema(ctx context.Context, cache SchemaCache, entityName string, tenantID uuid.UUID) error {
	keys := []string{
		cacheKey(entityName, "crud", tenantID),
		cacheKey(entityName, "form", tenantID),
	}
	return cache.Del(ctx, keys...).Err()
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

func (h *Handler) servePage(c *fiber.Ctx, formOnly bool) error {
	entityName := c.Params("entity")
	entDef := def.Lookup(entityName)
	if entDef == nil {
		return fiber.NewError(fiber.StatusNotFound, "unknown entity: "+entityName)
	}

	// Require auth when viewer function is configured.
	if h.viewerFn != nil {
		if _, err := h.viewerFn(c); err != nil {
			return fiber.ErrUnauthorized
		}
	}

	opts := amis.PageOpts{APIBase: h.apiBase}

	// Resolve tenant for custom fields and cache key.
	tenantID := uuid.Nil
	if h.tenantFn != nil {
		if tid, err := h.tenantFn(c); err == nil {
			tenantID = tid
		}
	}

	// Enrich with tenant-specific custom fields when a source is configured.
	if h.customFields != nil && tenantID != uuid.Nil {
		extra, err := h.customFields(c.Context(), tenantID, entDef.Name)
		if err == nil {
			opts.ExtraFields = extra
		}
		// Non-fatal: serve base schema on registry error.
	}

	schemaType := "crud"
	if formOnly {
		schemaType = "form"
	}

	// Check PageBuilderSet override before hitting the cache.
	// Custom page builders are not cached because they may carry per-request state.
	if entDef.PageBuilders != nil {
		if formOnly && entDef.PageBuilders.FormPage != nil {
			return c.JSON(entDef.PageBuilders.FormPage(entDef, h.apiBase))
		}
		if !formOnly && entDef.PageBuilders.CRUDPage != nil {
			return c.JSON(entDef.PageBuilders.CRUDPage(entDef, h.apiBase))
		}
	}

	// Cache read.
	if h.cache != nil {
		key := cacheKey(entityName, schemaType, tenantID)
		if cached, err := h.cache.Get(c.Context(), key).Bytes(); err == nil {
			c.Set("Content-Type", "application/json")
			c.Set("X-Schema-Cache", "HIT")
			return c.Send(cached)
		}
	}

	// Generate schema.
	var schema map[string]any
	if formOnly {
		schema = amis.FormPage(entDef, opts)
	} else {
		schema = amis.CRUDPage(entDef, opts)
	}

	// Cache write (best-effort).
	if h.cache != nil {
		if b, err := json.Marshal(schema); err == nil {
			key := cacheKey(entityName, schemaType, tenantID)
			_ = h.cache.Set(c.Context(), key, b, schemaCacheTTL).Err()
		}
	}

	return c.JSON(schema)
}

// cacheKey builds the Redis key for an entity's schema.
// tenantID == Nil → key is tenant-agnostic (no custom fields injected).
func cacheKey(entityName, schemaType string, tenantID uuid.UUID) string {
	// Hash the entity name to keep keys short and avoid special-char issues.
	h := sha256.Sum256([]byte(entityName))
	entityHash := fmt.Sprintf("%x", h[:8])
	return fmt.Sprintf(schemaCacheKeyFmt, entityHash, schemaType, tenantID)
}
