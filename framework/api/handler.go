// Package api wires EntityDefinition metadata into Fiber HTTP handlers.
package api

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"awo.so/framework/def"
	"awo.so/framework/hooks"
	"awo.so/framework/persistence"
	"awo.so/framework/platform/org"
	"awo.so/framework/privacy"
	"awo.so/framework/validate"
)

// ViewerFromCtx extracts a ViewerContext from a Fiber request context.
// Host application must register a middleware that calls c.Locals("viewer", ...).
type ViewerFromCtx func(c *fiber.Ctx) (def.ViewerContext, error)

// TenantResolver resolves a tenant slug or UUID string to a UUID.
// Called when viewer.TenantID() is not a valid UUID (e.g. dev slug headers).
// Return uuid.Nil + error to reject the request.
type TenantResolver func(ctx context.Context, slugOrID string) (uuid.UUID, error)

// Handler is a generic CRUD handler for one EntityDefinition.
type Handler struct {
	def            *def.EntityDefinition
	store          persistence.TenantStore
	hooks          *hooks.Runner
	enforcer       *privacy.Enforcer
	viewer         ViewerFromCtx
	tenantResolver TenantResolver
	orgTree        org.Tree
}

// NewHandler creates a Handler for def.
func NewHandler(
	def *def.EntityDefinition,
	store persistence.TenantStore,
	viewer ViewerFromCtx,
	opts ...HandlerOption,
) *Handler {
	h := &Handler{
		def:      def,
		store:    store,
		hooks:    hooks.DefaultRunner,
		enforcer: privacy.New(),
		viewer:   viewer,
	}
	for _, o := range opts {
		o(h)
	}
	return h
}

// HandlerOption configures a Handler.
type HandlerOption func(*Handler)

// WithTenantResolver sets a function that resolves tenant slug → UUID.
func WithTenantResolver(fn TenantResolver) HandlerOption {
	return func(h *Handler) { h.tenantResolver = fn }
}

// WithOrgTree sets the org.Tree used by AllowWithinOrgScope policies.
func WithOrgTree(tree org.Tree) HandlerOption {
	return func(h *Handler) { h.orgTree = tree }
}

// Register mounts CRUD routes under prefix on router.
//
//	GET    /prefix          → List
//	POST   /prefix          → Create
//	GET    /prefix/:id      → FindByID
//	PATCH  /prefix/:id      → Update (partial)
//	PUT    /prefix/:id      → Update (alias — full-replace semantics not yet enforced)
//	DELETE /prefix/:id      → Delete
func Register(router fiber.Router, prefix string, h *Handler) {
	g := router.Group(prefix)
	g.Get("/", h.list)
	g.Post("/", h.create)
	g.Get("/:id", h.findByID)
	g.Patch("/:id", h.update) // canonical per docs
	g.Put("/:id", h.update)   // backward-compat alias
	g.Delete("/:id", h.delete)

	// Mount declared custom actions: POST /{prefix}/:id/{action.Name}
	for _, action := range h.def.Actions {
		action := action
		g.Post("/:id/"+action.Name, makeActionHandler(h, action))
	}
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

	var result map[string]any
	if err := h.store.WithTx(c.Context(), tenantID, func(tx persistence.TenantTx) error {
		es := tx.ForEntity(h.def.Name)
		rec, err := es.FindByID(c.Context(), id)
		if err != nil {
			return err
		}
		// Org scope check before policy evaluation so cross-org reads look
		// identical to not-found (prevents existence probing across units).
		if err := h.assertOrgScope(c.Context(), tenantID, viewer, rec); err != nil {
			return errForbidden
		}
		if err := h.enforcer.Allow(c.Context(), h.def, viewer, def.OpRead, rec); err != nil {
			return errForbidden // hide existence from unauthorized viewers
		}
		result = recordToMap(rec, h.def)
		return nil
	}); err != nil {
		return fiberErr(err)
	}

	return OK(c, result)
}

func (h *Handler) list(c *fiber.Ctx) error {
	viewer, tenantID, err := h.extractContext(c)
	if err != nil {
		return err
	}
	// Policy check against nil record — policies must handle nil for list-level checks.
	if err := h.enforcer.Allow(c.Context(), h.def, viewer, def.OpRead, nil); err != nil {
		return fiber.ErrForbidden
	}

	opts := persistence.ListOptions{
		Search:    c.Query("q"),
		OrderBy:   c.Query("order_by"),
		Ascending: c.Query("dir") == "asc",
	}
	if lim := c.QueryInt("limit", 50); lim > 0 {
		opts.Limit = lim
	}
	if off := c.QueryInt("offset", 0); off >= 0 {
		opts.Offset = off
	}

	// For unit-scoped entities, restrict results to the viewer's org subtree.
	// Tenant-wide viewers (OrgUnitID == Nil) see all rows — no restriction.
	if h.def.IsUnitScoped() && h.orgTree != nil && viewer.OrgUnitID() != uuid.Nil {
		descendants, err := h.orgTree.Descendants(c.Context(), tenantID, viewer.OrgUnitID())
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "org tree lookup failed")
		}
		opts.OrgUnitIDs = descendants
	}

	var (
		rows  []map[string]any
		total int64
		lim   int
		off   int
	)
	if err := h.store.WithTx(c.Context(), tenantID, func(tx persistence.TenantTx) error {
		es := tx.ForEntity(h.def.Name)
		page, err := es.List(c.Context(), opts)
		if err != nil {
			return err
		}
		rows = make([]map[string]any, len(page.Records))
		for i, r := range page.Records {
			rows[i] = recordToMap(r, h.def)
		}
		total = page.Total
		lim = page.Limit
		off = page.Offset
		return nil
	}); err != nil {
		return fiberErr(err)
	}

	return List(c, rows, total, lim, off)
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

	// Resolve the target org unit. For unit-scoped entities the caller may
	// specify an org_unit_id in the body to create in a child unit; if absent
	// we default to the viewer's own unit. Viewers with a non-nil org unit must
	// be an ancestor-or-equal of the target unit (hierarchical create).
	// For non-unit-scoped entities this is a no-op (scopeColumns omits
	// org_unit_id; uuid.Nil is stored but never written to the DB).
	if h.def.IsUnitScoped() {
		orgID := resolveOrgUnitID(body, viewer)
		if h.orgTree != nil && !viewer.IsSystem() &&
			viewer.OrgUnitID() != uuid.Nil && orgID != viewer.OrgUnitID() {
			if err := org.AssertAncestor(c.Context(), h.orgTree, tenantID, viewer.OrgUnitID(), orgID); err != nil {
				return fiber.ErrForbidden
			}
		}
		rec.Set("org_unit_id", orgID)
	} else {
		rec.Set("org_unit_id", viewer.OrgUnitID())
	}

	if err := h.enforcer.Allow(c.Context(), h.def, viewer, def.OpCreate, rec); err != nil {
		return fiber.ErrForbidden
	}

	mut := &def.Mutation{Op: def.OpCreate, After: rec, TenantID: tenantID.String(), ActorID: viewer.ActorID()}

	// RunBeforeValidate fires outside the transaction so hooks can normalise
	// input before validation (e.g. derive computed fields, trim whitespace).
	if err := h.hooks.Run(c.Context(), h.def, mut, def.HookBeforeValidate); err != nil {
		return fiberErr(err)
	}

	if verrs := validate.Run(h.def, rec); len(verrs) > 0 {
		return ValidationErr(c, verrFields(verrs))
	}

	if err := h.store.WithTx(c.Context(), tenantID, func(tx persistence.TenantTx) error {
		es := tx.ForEntity(h.def.Name)

		if err := h.hooks.Run(c.Context(), h.def, mut, def.HookBeforeSave); err != nil {
			return err
		}
		if err := es.Create(c.Context(), rec); err != nil {
			return err
		}
		return h.hooks.Run(c.Context(), h.def, mut, def.HookAfterSave)
	}); err != nil {
		return fiberErr(err)
	}

	// Fire after_commit hooks outside the transaction (Temporal triggers, outbox, etc.).
	h.hooks.RunAfterCommit(c.Context(), h.def, mut)

	return Created(c, recordToMap(rec, h.def))
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

	// Fetch existing record and validate outside the transaction so HookBeforeValidate
	// can perform external lookups without holding a DB connection.
	before, err := func() (def.Record, error) {
		var rec def.Record
		if err := h.store.WithTx(c.Context(), tenantID, func(tx persistence.TenantTx) error {
			var e error
			rec, e = tx.ForEntity(h.def.Name).FindByID(c.Context(), id)
			return e
		}); err != nil {
			return nil, err
		}
		return rec, nil
	}()
	if err != nil {
		return fiberErr(err)
	}

	// Org scope check: viewer must be ancestor-or-equal of the record's unit.
	// Return ErrNotFound (not ErrForbidden) to hide cross-unit existence.
	if err := h.assertOrgScope(c.Context(), tenantID, viewer, before); err != nil {
		return fiberErr(persistence.ErrNotFound)
	}
	if err := h.enforcer.Allow(c.Context(), h.def, viewer, def.OpUpdate, before); err != nil {
		return fiber.ErrForbidden
	}

	rec := newMutableFromRecord(before, body)
	mut := &def.Mutation{
		Op:       def.OpUpdate,
		Before:   before,
		After:    rec,
		TenantID: tenantID.String(),
		ActorID:  viewer.ActorID(),
	}

	// HookBeforeValidate and validation run outside the transaction (consistent with create).
	if err := h.hooks.Run(c.Context(), h.def, mut, def.HookBeforeValidate); err != nil {
		return fiberErr(err)
	}
	if verrs := validate.Run(h.def, rec); len(verrs) > 0 {
		return ValidationErr(c, verrFields(verrs))
	}

	var result map[string]any
	txErr := h.store.WithTx(c.Context(), tenantID, func(tx persistence.TenantTx) error {
		es := tx.ForEntity(h.def.Name)

		if err := h.hooks.Run(c.Context(), h.def, mut, def.HookBeforeSave); err != nil {
			return err
		}
		if err := es.Update(c.Context(), rec); err != nil {
			return err
		}
		if err := h.hooks.Run(c.Context(), h.def, mut, def.HookAfterSave); err != nil {
			return err
		}
		result = recordToMap(rec, h.def)
		return nil
	})
	if txErr != nil {
		var verrs validate.ValidationErrors
		if errors.As(txErr, &verrs) {
			return ValidationErr(c, verrFields(verrs))
		}
		return fiberErr(txErr)
	}

	// Fire after_commit hooks outside the transaction.
	h.hooks.RunAfterCommit(c.Context(), h.def, mut)

	return OK(c, result)
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

	var deleteMut *def.Mutation
	if err := h.store.WithTx(c.Context(), tenantID, func(tx persistence.TenantTx) error {
		es := tx.ForEntity(h.def.Name)

		rec, err := es.FindByID(c.Context(), id)
		if err != nil {
			return err
		}
		if err := h.assertOrgScope(c.Context(), tenantID, viewer, rec); err != nil {
			return errForbidden
		}
		if err := h.enforcer.Allow(c.Context(), h.def, viewer, def.OpDelete, rec); err != nil {
			return errForbidden
		}

		deleteMut = &def.Mutation{Op: def.OpDelete, Before: rec, TenantID: tenantID.String(), ActorID: viewer.ActorID()}

		if err := h.hooks.Run(c.Context(), h.def, deleteMut, def.HookBeforeDelete); err != nil {
			return err
		}
		if err := es.Delete(c.Context(), id); err != nil {
			return err
		}
		return h.hooks.Run(c.Context(), h.def, deleteMut, def.HookAfterDelete)
	}); err != nil {
		return fiberErr(err)
	}

	// Fire after_commit hooks outside the transaction.
	if deleteMut != nil {
		h.hooks.RunAfterCommit(c.Context(), h.def, deleteMut)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ──────────────────────────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────────────────────────

var errForbidden = errors.New("forbidden")

// assertOrgScope returns an error when the viewer's org unit is NOT an
// ancestor-or-equal of the record's org unit.
//
// No-op when:
//   - entity is not unit-scoped
//   - no orgTree configured on the handler
//   - viewer is a system caller (bypasses unit checks)
//   - viewer is tenant-wide (OrgUnitID == Nil — all units visible)
//   - record does not implement def.OrgScoped, or its unit is Nil
func (h *Handler) assertOrgScope(ctx context.Context, tenantID uuid.UUID, viewer def.ViewerContext, rec def.Record) error {
	if !h.def.IsUnitScoped() || h.orgTree == nil || viewer.IsSystem() || viewer.OrgUnitID() == uuid.Nil {
		return nil
	}
	scoped, ok := rec.(def.OrgScoped)
	if !ok {
		return nil
	}
	recordUnitID := scoped.RecordOrgUnitID()
	if recordUnitID == uuid.Nil {
		return nil
	}
	return org.AssertAncestor(ctx, h.orgTree, tenantID, viewer.OrgUnitID(), recordUnitID)
}

// resolveOrgUnitID picks the target org_unit_id for a new record.
// Prefers the body's "org_unit_id" value if it parses as a non-nil UUID;
// falls back to the viewer's own org unit.
func resolveOrgUnitID(body map[string]any, viewer def.ViewerContext) uuid.UUID {
	if raw, ok := body["org_unit_id"]; ok {
		switch v := raw.(type) {
		case string:
			if id, err := uuid.Parse(v); err == nil && id != uuid.Nil {
				return id
			}
		case uuid.UUID:
			if v != uuid.Nil {
				return v
			}
		}
	}
	return viewer.OrgUnitID()
}

func (h *Handler) extractContext(c *fiber.Ctx) (def.ViewerContext, uuid.UUID, error) {
	viewer, err := h.viewer(c)
	if err != nil {
		return nil, uuid.Nil, fiber.ErrUnauthorized
	}

	tenantID, err := uuid.Parse(viewer.TenantID())
	if err != nil {
		// Not a UUID — try resolving as a slug if a resolver is configured.
		if h.tenantResolver == nil {
			return nil, uuid.Nil, fiber.NewError(fiber.StatusUnauthorized,
				"tenant ID is not a valid UUID and no resolver is configured")
		}
		tenantID, err = h.tenantResolver(c.Context(), viewer.TenantID())
		if err != nil {
			return nil, uuid.Nil, fiber.NewError(fiber.StatusUnauthorized,
				"could not resolve tenant: "+err.Error())
		}
	}

	if tenantID == uuid.Nil {
		return nil, uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "missing tenant")
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

// verrFields converts ValidationErrors to a field→message map for ValidationErr.
func verrFields(verrs validate.ValidationErrors) map[string]string {
	fields := make(map[string]string, len(verrs))
	for _, ve := range verrs {
		fields[ve.Field] = ve.Message
	}
	return fields
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

func recordToMap(rec def.Record, def *def.EntityDefinition) map[string]any {
	m := map[string]any{
		"id":         rec.ID(),
		"created_at": rec.Get("created_at"),
		"updated_at": rec.Get("updated_at"),
	}
	if !def.IsGlobal() {
		m["tenant_id"] = rec.TenantID()
	}
	if def.IsUnitScoped() {
		if scoped, ok := rec.(interface{ OrgUnitID() uuid.UUID }); ok {
			m["org_unit_id"] = scoped.OrgUnitID()
		}
	}
	for _, f := range def.Fields {
		if !f.IsSensitive {
			m[f.Name] = rec.Get(f.Name)
		}
	}
	return m
}

// newMutableFromMap creates a MutableRecord from a JSON body map.
// tenantID is injected; id taken from body["id"] if present.
func newMutableFromMap(entityName string, tenantID uuid.UUID, body map[string]any) def.MutableRecord {
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
func newMutableFromRecord(base def.Record, changes map[string]any) def.MutableRecord {
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
	def := def.Lookup(base.EntityName())
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

func (r *simpleRecord) Get(field string) any      { return r.data[field] }
func (r *simpleRecord) Set(field string, val any) { r.data[field] = val }
func (r *simpleRecord) ID() uuid.UUID             { return r.id }
func (r *simpleRecord) TenantID() uuid.UUID       { return r.tenantID }
func (r *simpleRecord) EntityName() string        { return r.entityName }

var _ def.MutableRecord = (*simpleRecord)(nil)
