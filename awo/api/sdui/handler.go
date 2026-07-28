// Package sdui provides the HTTP handler layer for the SDUI engine.
//
// # Responsibilities
//
// This package owns the HTTP surface of the SDUI subsystem. It:
//   - Extracts request parameters (entity name, view mode, locale, renderer ID).
//   - Bridges auth.ViewerContext → sduictx.ViewerContext via adapt.ViewerAdapter.
//   - Builds the sduictx.GeneratorContext for each request.
//   - Delegates all generation and rendering to engine.Engine.
//   - Serializes the rendered output as JSON.
//   - Sets ETag and Cache-Control headers.
//
// # Endpoints
//
// All endpoints are mounted at /api/v1/ui by the router:
//
//	GET /api/v1/ui/{module}/{entity}          → list view
//	GET /api/v1/ui/{module}/{entity}/create   → create form
//	GET /api/v1/ui/{module}/{entity}/{id}     → detail view
//	GET /api/v1/ui/{module}/{entity}/{id}/edit → edit form
//
// The renderer is selected via the Accept-SDUI-Renderer header (default: amis).
// Locale is selected via the Accept-Language header (default: en-US).
//
// # No business logic
//
// This package contains no SDUI logic. All generation, caching, and rendering
// decisions are delegated to engine.Engine. HTTP handlers are thin.
package sdui

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"

	"awo.so/awo/auth"
	"awo.so/awo/compiler"
	"awo.so/awo/sdui/adapt"
	"awo.so/awo/sdui/engine"
	"awo.so/awo/sdui/renderer"
	"awo.so/awo/sdui/sduictx"
)

const (
	// defaultRendererID is used when no renderer is specified in the request.
	defaultRendererID = "amis"

	// headerRenderer is the HTTP request header for renderer selection.
	headerRenderer = "Accept-SDUI-Renderer"

	// headerLocale is the HTTP request header for locale selection.
	headerLocale = "Accept-Language"

	// headerETag is the HTTP response header for conditional requests.
	headerETag = "ETag"

	// headerCacheControl is the HTTP response header for cache hints.
	headerCacheControl = "Cache-Control"

	// sduiPublicMaxAge is the max-age for SDUI responses (5 minutes).
	sduiPublicMaxAge = "private, max-age=300"
)

// Handler handles SDUI page requests for a compiled schema.
// Safe for concurrent use; holds no mutable state.
type Handler struct {
	schema    *compiler.CompiledSchema
	engine    *engine.Engine
	evaluator auth.PolicyEvaluator   // may be nil
	grants    adapt.GrantIndex        // pre-built permission index
	schemaFPs map[string]string       // entityName → fingerprint (precomputed)
}

// New constructs a Handler.
//
//   - schema is the compiled schema produced by compiler.Compile. Required.
//   - eng is the configured SDUI engine. Required.
//   - evaluator is the policy evaluator for field permission gating. May be nil.
//
// New precomputes schema fingerprints and the grant index for O(1) access
// during request handling.
func New(schema *compiler.CompiledSchema, eng *engine.Engine, evaluator auth.PolicyEvaluator) *Handler {
	fps := make(map[string]string, len(schema.Entities))
	for _, es := range schema.Entities {
		fps[es.QualifiedName] = adapt.SchemaFingerprint(es)
	}
	return &Handler{
		schema:    schema,
		engine:    eng,
		evaluator: evaluator,
		grants:    adapt.BuildGrantIndex(schema.CapabilityGrants),
		schemaFPs: fps,
	}
}

// Register mounts SDUI endpoints on group g.
// The caller is responsible for applying auth middleware before the group.
//
//	ui := app.Group("/api/v1/ui")
//	ui.Use(middleware.RequireAuth(...))
//	h.Register(ui)
func (h *Handler) Register(g fiber.Router) {
	// List view.
	g.Get("/:module/:resource", h.list)

	// Create form.
	g.Get("/:module/:resource/create", h.create)

	// Edit form.
	g.Get("/:module/:resource/:id/edit", h.edit)

	// Detail view. Register after /create and /:id/edit to avoid ambiguity.
	g.Get("/:module/:resource/:id", h.detail)
}

// ── view handlers ─────────────────────────────────────────────────────────────

func (h *Handler) list(c *fiber.Ctx) error {
	return h.handle(c, sduictx.ViewModeList, false)
}

func (h *Handler) create(c *fiber.Ctx) error {
	return h.handle(c, sduictx.ViewModeCreate, false)
}

func (h *Handler) detail(c *fiber.Ctx) error {
	return h.handle(c, sduictx.ViewModeDetail, true)
}

func (h *Handler) edit(c *fiber.Ctx) error {
	return h.handle(c, sduictx.ViewModeEdit, false)
}

// handle is the shared implementation for all view handlers.
func (h *Handler) handle(c *fiber.Ctx, mode sduictx.ViewMode, readOnly bool) error {
	module := c.Params("module")
	resource := c.Params("resource")

	// Resolve entity by module + resource path segment.
	es := h.findEntity(module, resource)
	if es == nil {
		return fiber.NewError(fiber.StatusNotFound,
			fmt.Sprintf("sdui: entity not found: %s/%s", module, resource))
	}

	viewer := auth.ViewerFromContext(c.UserContext())
	rendererID := h.rendererID(c)
	locale := h.locale(c)
	schemaFP := h.schemaFPs[es.QualifiedName]

	// Build viewer adapter (bridges auth.ViewerContext → sduictx.ViewerContext).
	sduiViewer := adapt.NewViewerAdapter(c.UserContext(), viewer, h.evaluator, h.grants)

	// Build GeneratorContext.
	builder := sduictx.NewGeneratorContext(
		viewer.TenantID(),
		sduiViewer,
		es.QualifiedName,
		mode,
		rendererID,
	).
		WithLocale(locale).
		WithSchemaFingerprint(schemaFP)

	if readOnly {
		builder = builder.WithReadOnly()
	}

	ctx, err := builder.Build()
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	// Convert compiled schema → generator schema.
	gSchema := adapt.FromCompiled(es)

	// Delegate to engine.
	resp, err := h.engine.Handle(c.UserContext(), engine.Request{
		Ctx:        ctx,
		Schema:     gSchema,
		RendererID: rendererID,
		RendererCtx: renderer.RendererContext{
			GenCtx: ctx,
		},
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// Set ETag from schema fingerprint + renderer + locale.
	etag := fmt.Sprintf(`"%s-%s-%s"`, schemaFP, rendererID, locale)
	c.Set(headerETag, etag)
	c.Set(headerCacheControl, sduiPublicMaxAge)

	// Check conditional request (If-None-Match).
	if c.Get("If-None-Match") == etag {
		return c.SendStatus(fiber.StatusNotModified)
	}

	return c.JSON(resp.Output.AMISSchema)
}

// ── helpers ───────────────────────────────────────────────────────────────────

// findEntity resolves an entity schema by module and resource (plural local name).
// Returns nil when not found.
func (h *Handler) findEntity(module, resource string) *compiler.EntitySchema {
	for _, es := range h.schema.Entities {
		if es.Module == module && es.APIResource == resource {
			return es
		}
	}
	return nil
}

// rendererID extracts the renderer ID from Accept-SDUI-Renderer header.
// Falls back to defaultRendererID.
func (h *Handler) rendererID(c *fiber.Ctx) string {
	r := strings.TrimSpace(c.Get(headerRenderer))
	if r == "" {
		return defaultRendererID
	}
	return r
}

// locale extracts the locale from Accept-Language header.
// Returns only the first language tag (e.g. "en-US" from "en-US,en;q=0.9").
// Falls back to "en-US".
func (h *Handler) locale(c *fiber.Ctx) string {
	raw := strings.TrimSpace(c.Get(headerLocale))
	if raw == "" {
		return "en-US"
	}
	// Take the first locale tag before any comma or semicolon.
	for _, sep := range []string{",", ";"} {
		if idx := strings.IndexByte(raw, sep[0]); idx >= 0 {
			raw = raw[:idx]
		}
	}
	return strings.TrimSpace(raw)
}
