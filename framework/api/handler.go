// Package api wires EntityDefinition metadata into Fiber HTTP handlers.
package api

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"awo.so/framework/definition"
	"awo.so/framework/hooks"
	"awo.so/framework/persistence"
	"awo.so/framework/privacy"
)

// ViewerFromCtx extracts a ViewerContext from a Fiber request context.
// Host application must register a middleware that calls c.Locals("viewer", ...).
type ViewerFromCtx func(c *fiber.Ctx) (definition.ViewerContext, error)

// Handler is a generic CRUD handler for one EntityDefinition.
type Handler struct {
	def      *definition.EntityDefinition
	store    persistence.TenantStore
	hooks    *hooks.Executor
	enforcer *privacy.Enforcer
	viewer   ViewerFromCtx
}

// NewHandler creates a Handler for def.
func NewHandler(
	def *definition.EntityDefinition,
	store persistence.TenantStore,
	viewer ViewerFromCtx,
) *Handler {
	return &Handler{
		def:      def,
		store:    store,
		hooks:    hooks.New(),
		enforcer: privacy.New(),
		viewer:   viewer,
	}
}

// Register mounts CRUD routes under prefix on router.
//
//	GET    /prefix          → List
//	POST   /prefix          → Create
//	GET    /prefix/:id      → FindByID
//	PUT    /prefix/:id      → Update
//	DELETE /prefix/:id      → Delete
func Register(router fiber.Router, prefix string, h *Handler) {
	g := router.Group(prefix)
	g.Get("/", h.list)
	g.Post("/", h.create)
	g.Get("/:id", h.findByID)
	g.Put("/:id", h.update)
	g.Delete("/:id", h.delete)
}

// ──────────────────────────────────────────────────────────────────
// Handlers
// ──────────────────────────────────────────────────────────────────

func (h *Handler) findByID(c *fiber.Ctx) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}
	viewer, tenantID, err := h.extractContext(c)
	if err != nil {
		return err
	}

	es, err := h.store.ForEntity(c.Context(), tenantID, h.def.Name)
	if err != nil {
		return fiberErr(err)
	}

	rec, err := es.FindByID(c.Context(), id)
	if err != nil {
		return fiberErr(err)
	}

	if err := h.enforcer.Allow(c.Context(), h.def, viewer, definition.OpRead, rec); err != nil {
		return fiber.ErrNotFound // hide existence from unauthorized viewers
	}

	return c.JSON(recordToMap(rec, h.def))
}

func (h *Handler) list(c *fiber.Ctx) error {
	viewer, tenantID, err := h.extractContext(c)
	if err != nil {
		return err
	}
	// Policy check against nil record — policies must handle nil for list-level checks.
	if err := h.enforcer.Allow(c.Context(), h.def, viewer, definition.OpRead, nil); err != nil {
		return fiber.ErrForbidden
	}

	opts := persistence.ListOptions{
		Search: c.Query("q"),
		OrderBy: c.Query("order_by"),
	}
	if lim := c.QueryInt("limit", 50); lim > 0 {
		opts.Limit = lim
	}
	if off := c.QueryInt("offset", 0); off >= 0 {
		opts.Offset = off
	}
	opts.Ascending = c.Query("dir") == "asc"

	es, err := h.store.ForEntity(c.Context(), tenantID, h.def.Name)
	if err != nil {
		return fiberErr(err)
	}

	page, err := es.List(c.Context(), opts)
	if err != nil {
		return fiberErr(err)
	}

	rows := make([]map[string]any, len(page.Records))
	for i, r := range page.Records {
		rows[i] = recordToMap(r, h.def)
	}
	return c.JSON(fiber.Map{
		"data":   rows,
		"total":  page.Total,
		"limit":  page.Limit,
		"offset": page.Offset,
	})
}

func (h *Handler) create(c *fiber.Ctx) error {
	viewer, tenantID, err := h.extractContext(c)
	if err != nil {
		return err
	}

	body := make(map[string]any)
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body")
	}

	rec := newMutableFromMap(h.def.Name, tenantID, body)

	if err := h.enforcer.Allow(c.Context(), h.def, viewer, definition.OpCreate, rec); err != nil {
		return fiber.ErrForbidden
	}

	mut := &definition.Mutation{Op: definition.OpCreate, After: rec, TenantID: tenantID.String(), ActorID: viewer.ActorID()}

	if err := h.store.WithTx(c.Context(), tenantID, func(tx persistence.TenantTx) error {
		es := tx.ForEntity(h.def.Name)

		if err := h.hooks.RunBefore(c.Context(), h.def, mut); err != nil {
			return err
		}
		if err := es.Create(c.Context(), rec); err != nil {
			return err
		}
		return h.hooks.RunAfter(c.Context(), h.def, mut)
	}); err != nil {
		return fiberErr(err)
	}

	return c.Status(fiber.StatusCreated).JSON(recordToMap(rec, h.def))
}

func (h *Handler) update(c *fiber.Ctx) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}
	viewer, tenantID, err := h.extractContext(c)
	if err != nil {
		return err
	}

	body := make(map[string]any)
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body")
	}

	var result map[string]any
	if err := h.store.WithTx(c.Context(), tenantID, func(tx persistence.TenantTx) error {
		es := tx.ForEntity(h.def.Name)

		before, err := es.FindByID(c.Context(), id)
		if err != nil {
			return err
		}

		if err := h.enforcer.Allow(c.Context(), h.def, viewer, definition.OpUpdate, before); err != nil {
			return errForbidden
		}

		rec := newMutableFromRecord(before, body)
		mut := &definition.Mutation{
			Op:       definition.OpUpdate,
			Before:   before,
			After:    rec,
			TenantID: tenantID.String(),
			ActorID:  viewer.ActorID(),
		}

		if err := h.hooks.RunBefore(c.Context(), h.def, mut); err != nil {
			return err
		}
		if err := es.Update(c.Context(), rec); err != nil {
			return err
		}
		if err := h.hooks.RunAfter(c.Context(), h.def, mut); err != nil {
			return err
		}
		result = recordToMap(rec, h.def)
		return nil
	}); err != nil {
		return fiberErr(err)
	}

	return c.JSON(result)
}

func (h *Handler) delete(c *fiber.Ctx) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}
	viewer, tenantID, err := h.extractContext(c)
	if err != nil {
		return err
	}

	if err := h.store.WithTx(c.Context(), tenantID, func(tx persistence.TenantTx) error {
		es := tx.ForEntity(h.def.Name)

		rec, err := es.FindByID(c.Context(), id)
		if err != nil {
			return err
		}
		if err := h.enforcer.Allow(c.Context(), h.def, viewer, definition.OpDelete, rec); err != nil {
			return errForbidden
		}

		mut := &definition.Mutation{Op: definition.OpDelete, Before: rec, TenantID: tenantID.String(), ActorID: viewer.ActorID()}

		if err := h.hooks.RunBefore(c.Context(), h.def, mut); err != nil {
			return err
		}
		if err := es.Delete(c.Context(), id); err != nil {
			return err
		}
		return h.hooks.RunAfter(c.Context(), h.def, mut)
	}); err != nil {
		return fiberErr(err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ──────────────────────────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────────────────────────

var errForbidden = errors.New("forbidden")

func (h *Handler) extractContext(c *fiber.Ctx) (definition.ViewerContext, uuid.UUID, error) {
	viewer, err := h.viewer(c)
	if err != nil {
		return nil, uuid.Nil, fiber.ErrUnauthorized
	}
	tenantID, err := uuid.Parse(viewer.TenantID())
	if err != nil {
		return nil, uuid.Nil, fiber.ErrUnauthorized
	}
	return viewer, tenantID, nil
}

func parseUUID(c *fiber.Ctx, param string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params(param))
	if err != nil {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "invalid UUID: "+c.Params(param))
	}
	return id, nil
}

func fiberErr(err error) error {
	switch {
	case errors.Is(err, persistence.ErrNotFound):
		return fiber.ErrNotFound
	case errors.Is(err, persistence.ErrConflict):
		return fiber.NewError(fiber.StatusConflict, err.Error())
	case errors.Is(err, errForbidden):
		return fiber.ErrForbidden
	default:
		return err
	}
}

func recordToMap(rec definition.Record, def *definition.EntityDefinition) map[string]any {
	m := map[string]any{
		"id":         rec.ID(),
		"tenant_id":  rec.TenantID(),
		"created_at": rec.Get("created_at"),
		"updated_at": rec.Get("updated_at"),
	}
	for _, f := range def.Fields {
		if !f.Hidden {
			m[f.Name] = rec.Get(f.Name)
		}
	}
	return m
}

// newMutableFromMap creates a MutableRecord from a JSON body map.
// tenantID is injected; id taken from body["id"] if present.
func newMutableFromMap(entityName string, tenantID uuid.UUID, body map[string]any) definition.MutableRecord {
	rec := &simpleRecord{
		entityName: entityName,
		tenantID:   tenantID,
		data:       make(map[string]any, len(body)),
	}
	if raw, ok := body["id"]; ok {
		if s, ok := raw.(string); ok {
			rec.id, _ = uuid.Parse(s)
		}
	}
	for k, v := range body {
		rec.data[k] = v
	}
	return rec
}

// newMutableFromRecord overlays body changes onto an existing record.
func newMutableFromRecord(base definition.Record, changes map[string]any) definition.MutableRecord {
	// We need access to base's full field set — reconstruct via Get.
	// This works because base is a *mapRecord from pgstore.
	rec := &simpleRecord{
		id:         base.ID(),
		tenantID:   base.TenantID(),
		entityName: base.EntityName(),
		data:       make(map[string]any),
	}
	// Copy existing values first.
	// (Record has no Iterate; we iterate known fields via EntityName lookup.)
	def := definition.Lookup(base.EntityName())
	if def != nil {
		for _, f := range def.Fields {
			rec.data[f.Name] = base.Get(f.Name)
		}
	}
	// Apply changes.
	for k, v := range changes {
		rec.data[k] = v
	}
	return rec
}

// simpleRecord is an api-internal MutableRecord backed by a map.
type simpleRecord struct {
	id         uuid.UUID
	tenantID   uuid.UUID
	entityName string
	data       map[string]any
}

func (r *simpleRecord) Get(field string) any        { return r.data[field] }
func (r *simpleRecord) Set(field string, val any)   { r.data[field] = val }
func (r *simpleRecord) ID() uuid.UUID               { return r.id }
func (r *simpleRecord) TenantID() uuid.UUID         { return r.tenantID }
func (r *simpleRecord) EntityName() string          { return r.entityName }

var _ definition.MutableRecord = (*simpleRecord)(nil)

