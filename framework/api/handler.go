// Package api wires Entitydef metadata into Fiber HTTP handlers,
// providing generic CRUD endpoints for every registered entity without
// requiring hand-written controller code.
//
// # Architecture
//
// A [Handler] owns a single [def.Entitydef] and delegates
// persistence to a [persistence.TenantStore].  Privacy enforcement is
// performed by [privacy.Enforcer] using the policies declared on the
// Entitydef.  Business-logic hooks ([hooks.Runner]) fire at
// well-defined points in the request lifecycle (BeforeValidate, BeforeSave,
// AfterSave, BeforeDelete, AfterDelete).
//
// # Multi-tenancy
//
// Every request is scoped to a tenant UUID resolved from the viewer context.
// If the viewer carries a non-UUID value (e.g. a dev slug), an optional
// [TenantResolver] translates it.  All database access is performed inside a
// tenant-scoped transaction ([persistence.TenantTx]) so that PostgreSQL RLS
// policies apply unconditionally.
//
// # Org-unit scoping
//
// When an Entitydef is flagged IsUnitScoped, the list handler narrows
// results to the viewer's org-unit subtree (via [org.Tree]).  Individual
// record handlers additionally verify that the record's org_unit_id falls
// within the viewer's subtree, closing the ID-enumeration gap that a
// list-only filter would leave open.
//
// # Backward compatibility
//
// This package is a drop-in replacement for the original handler.go.
// [NewHandler] and [Register] have identical signatures.  All exported types
// and option functions are preserved.
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

// ──────────────────────────────────────────────────────────────────────────────
// Public function types
// ──────────────────────────────────────────────────────────────────────────────

// ViewerFromCtx extracts a [def.ViewerContext] from a Fiber request
// context.  The host application must register middleware that calls
// c.Locals("viewer", ...) before any handler in this package runs.
type ViewerFromCtx func(c *fiber.Ctx) (def.ViewerContext, error)

// TenantResolver translates a human-readable tenant slug (or any non-UUID
// string) to a canonical tenant UUID.  It is only invoked when
// viewer.TenantID() cannot be parsed as a UUID.
//
// Return (uuid.Nil, error) to reject the request with 401.
type TenantResolver func(ctx context.Context, slugOrID string) (uuid.UUID, error)

// ──────────────────────────────────────────────────────────────────────────────
// Handler
// ──────────────────────────────────────────────────────────────────────────────

// Handler is a generic, tenant-aware CRUD controller for one
// [def.EntityDefinition].  Use [NewHandler] to construct one and
// [Register] to mount it on a Fiber router.
type Handler struct {
	// def is the entity metadata: fields, policies, scope flags.
	def *def.EntityDefinition

	// store provides tenant-scoped, transactional persistence.
	store persistence.TenantStore

	// hooks executes lifecycle callbacks registered against the entity.
	hooks *hooks.Runner

	// enforcer evaluates privacy policies declared on the entity def.
	enforcer *privacy.Enforcer

	// viewer extracts the authenticated viewer from a Fiber context.
	viewer ViewerFromCtx

	// tenantResolver is optional; translates slugs to UUIDs when needed.
	tenantResolver TenantResolver

	// orgTree is used to resolve org-unit subtrees for scoped list / record
	// access.  Nil when the entity is not unit-scoped or when org-tree
	// enforcement is disabled at the handler level.
	orgTree org.Tree
}

// NewHandler creates a Handler for def backed by store.
//
// Sensible defaults are applied:
//   - hooks.DefaultRunner is used unless overridden with [WithHooksRunner].
//   - A fresh [privacy.Enforcer] is created from the definition's policies.
//   - No TenantResolver is configured (pure-UUID tenant IDs only).
//   - No org.Tree is configured (unit-scope list narrowing disabled).
//
// Override any default by passing [HandlerOption] values.
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

// ──────────────────────────────────────────────────────────────────────────────
// Handler options
// ──────────────────────────────────────────────────────────────────────────────

// HandlerOption is a functional option for [NewHandler].
type HandlerOption func(*Handler)

// WithTenantResolver configures fn to translate tenant slugs → UUIDs.
// Without this option, any non-UUID TenantID in the viewer context causes a
// 401 response.
func WithTenantResolver(fn TenantResolver) HandlerOption {
	return func(h *Handler) { h.tenantResolver = fn }
}

// WithOrgTree configures the org.Tree used to evaluate unit-scope access.
// Required for [def.Entitydef.IsUnitScoped] enforcement on both
// list and individual record endpoints.
func WithOrgTree(tree org.Tree) HandlerOption {
	return func(h *Handler) { h.orgTree = tree }
}

// WithHooksRunner replaces the default [hooks.Runner].  Useful in tests to
// inject a no-op runner or a runner with pre-registered test hooks.
func WithHooksRunner(r *hooks.Runner) HandlerOption {
	return func(h *Handler) { h.hooks = r }
}

// ──────────────────────────────────────────────────────────────────────────────
// Route registration
// ──────────────────────────────────────────────────────────────────────────────

// Register mounts the five standard CRUD routes for h under prefix on router.
//
//	GET    {prefix}/        → list all records (paginated)
//	POST   {prefix}/        → create a new record
//	GET    {prefix}/:id     → fetch one record by UUID
//	PUT    {prefix}/:id     → replace / update a record
//	DELETE {prefix}/:id     → soft- or hard-delete a record
//
// The trailing slash on collection routes is intentional; Fiber treats
// "/prefix" and "/prefix/" as distinct paths by default.
func Register(router fiber.Router, prefix string, h *Handler) {
	g := router.Group(prefix)
	g.Get("/", h.list)
	g.Post("/", h.create)
	g.Get("/:id", h.findByID)
	g.Put("/:id", h.update)
	g.Delete("/:id", h.delete)
}

// ──────────────────────────────────────────────────────────────────────────────
// HTTP handlers (unexported — accessed only via Register)
// ──────────────────────────────────────────────────────────────────────────────

// findByID handles GET /:id.
//
// Privacy note: a 403 is returned (not 404) when the record exists but the
// viewer lacks read permission.  This intentionally hides existence from
// unauthorised callers — the same record that a privileged user sees as 200
// appears as 403, not 404, to a lower-privilege user.  This differs from the
// "404 to hide existence" pattern because our threat model assumes IDs are not
// secret (they are UUIDs v4 and may appear in URLs shared between users of the
// same tenant).  Use a dedicated privacy policy if 404 masking is required.
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
			return err // ErrNotFound → 404 via fiberErr
		}

		// Enforce read permission before returning any data.
		if err := h.enforcer.Allow(c.Context(), h.def, viewer, def.OpRead, rec); err != nil {
			return errForbidden
		}

		// For unit-scoped entities, additionally verify that the record belongs
		// to a unit within the viewer's permitted subtree.  This closes the
		// gap where a scoped viewer could fetch records outside their subtree
		// by guessing UUIDs — the list filter alone does not protect against
		// direct ID lookups.
		if err := h.enforceOrgScope(c.Context(), tenantID, viewer, rec); err != nil {
			return err
		}

		result = recordToMap(rec, h.def)
		return nil
	}); err != nil {
		return fiberErr(err)
	}

	return c.JSON(result)
}

// list handles GET /.
//
// Supports query parameters:
//
//	q        — full-text / prefix search string forwarded to the store
//	order_by — column name to sort by (store validates allowlist)
//	dir      — "asc" or "desc" (default "desc")
//	limit    — page size, default 50, capped by store
//	offset   — zero-based record offset for pagination
//
// For unit-scoped entities the result set is narrowed to records whose
// org_unit_id falls within the viewer's org-unit subtree.  Viewers with a
// nil OrgUnitID (i.e. tenant-wide administrators) see all records.
func (h *Handler) list(c *fiber.Ctx) error {
	viewer, tenantID, err := h.extractContext(c)
	if err != nil {
		return err
	}

	// List-level policy check: rec is nil because no specific record exists yet.
	// Policies MUST handle a nil record for list checks; a non-nil record is
	// only available for per-record operations.  See [privacy.Enforcer] docs.
	if err := h.enforcer.Allow(c.Context(), h.def, viewer, def.OpRead, nil); err != nil {
		return fiber.ErrForbidden
	}

	opts := listOptsFromQuery(c)

	// Narrow to the viewer's org-unit subtree when the entity is unit-scoped
	// and the viewer is not a tenant-wide principal.
	if h.def.IsUnitScoped() && h.orgTree != nil && viewer.OrgUnitID() != uuid.Nil {
		descendants, err := h.orgTree.Descendants(c.Context(), tenantID, viewer.OrgUnitID())
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "org tree lookup failed")
		}
		opts.OrgUnitIDs = descendants
	}

	var result fiber.Map

	if err := h.store.WithTx(c.Context(), tenantID, func(tx persistence.TenantTx) error {
		es := tx.ForEntity(h.def.Name)

		page, err := es.List(c.Context(), opts)
		if err != nil {
			return err
		}

		rows := make([]map[string]any, len(page.Records))
		for i, r := range page.Records {
			rows[i] = recordToMap(r, h.def)
		}

		result = fiber.Map{
			"data":   rows,
			"total":  page.Total,
			"limit":  page.Limit,
			"offset": page.Offset,
		}
		return nil
	}); err != nil {
		return fiberErr(err)
	}

	return c.JSON(result)
}

// create handles POST /.
//
// Flow:
//  1. Parse JSON body.
//  2. Inject tenant & org-unit from viewer (user-supplied values are ignored).
//  3. Assign a new UUID if the body does not include one.
//  4. Enforce create permission.
//  5. Run HookBeforeValidate (outside the transaction — may normalise fields).
//  6. Validate the record.
//  7. Inside a transaction: run HookBeforeSave → persist → run HookAfterSave.
//  8. Return 201 with the created record.
//
// HookBeforeValidate runs outside the transaction deliberately: hooks at that
// stage are expected to normalise / derive values (e.g. format phone numbers,
// set computed fields) and must not rely on transactional isolation.  Any
// external side-effects performed by such hooks are not rolled back on later
// failure — hook authors must account for this.
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

	// Always stamp the viewer's org unit onto the record, overwriting any
	// client-supplied value.  This prevents privilege escalation where a user
	// assigns a record to a unit outside their permitted subtree.
	//
	// For non-unit-scoped entities org_unit_id is stored in the record map but
	// omitted from the SQL INSERT by the persistence layer (scopeColumns logic).
	rec.Set("org_unit_id", viewer.OrgUnitID())

	if err := h.enforcer.Allow(c.Context(), h.def, viewer, def.OpCreate, rec); err != nil {
		return fiber.ErrForbidden
	}

	mut := &def.Mutation{
		Op:       def.OpCreate,
		After:    rec,
		TenantID: tenantID.String(),
		ActorID:  viewer.ActorID(),
	}

	// Run normalisation hooks before validation so that derived fields are
	// present when validation rules execute.
	if err := h.hooks.Run(c.Context(), h.def, mut, def.HookBeforeValidate); err != nil {
		return fiberErr(err)
	}

	if verrs := validate.Run(h.def, rec); verrs != nil {
		return validationResponse(c, verrs)
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

	return c.Status(fiber.StatusCreated).JSON(recordToMap(rec, h.def))
}

// update handles PUT /:id.
//
// Flow:
//  1. Load the existing record (authoritative source for unchanged fields).
//  2. Enforce update permission against the pre-mutation record.
//  3. Overlay body changes onto a copy of the record (newMutableFromRecord).
//  4. Enforce org-unit scope (the viewer cannot move a record out of scope).
//  5. Run HookBeforeValidate → validate → HookBeforeSave → persist → HookAfterSave.
//  6. Return 200 with the updated record.
//
// All steps after (1) run inside a single transaction so that the pre-mutation
// record read and the subsequent write are serialised.
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

	txErr := h.store.WithTx(c.Context(), tenantID, func(tx persistence.TenantTx) error {
		es := tx.ForEntity(h.def.Name)

		// Load the current state before applying changes.
		before, err := es.FindByID(c.Context(), id)
		if err != nil {
			return err // ErrNotFound → 404
		}

		// Authorise against the pre-mutation record so that a policy can
		// inspect current field values (e.g. only the owner may edit).
		if err := h.enforcer.Allow(c.Context(), h.def, viewer, def.OpUpdate, before); err != nil {
			return errForbidden
		}

		// Verify org-unit scope before applying changes.
		if err := h.enforceOrgScope(c.Context(), tenantID, viewer, before); err != nil {
			return err
		}

		// Overlay the body changes onto a copy of the existing record so that
		// fields omitted from the request body retain their current values.
		rec, err := newMutableFromRecord(h.def, before, body)
		if err != nil {
			return err
		}

		// Prevent the client from re-assigning org_unit_id to a unit outside
		// their permitted subtree.  Stamp the existing value back unconditionally
		// so that a rogue body field cannot escalate scope.
		if h.def.IsUnitScoped() {
			if scoped, ok := before.(interface{ OrgUnitID() uuid.UUID }); ok {
				rec.Set("org_unit_id", scoped.OrgUnitID())
			}
		}

		mut := &def.Mutation{
			Op:       def.OpUpdate,
			Before:   before,
			After:    rec,
			TenantID: tenantID.String(),
			ActorID:  viewer.ActorID(),
		}

		if err := h.hooks.Run(c.Context(), h.def, mut, def.HookBeforeValidate); err != nil {
			return err
		}

		if verrs := validate.Run(h.def, rec); verrs != nil {
			// Return verrs as an error so the transaction is rolled back.
			// The outer error handler detects validate.ValidationErrors and
			// converts them to a 422 response.
			return verrs
		}

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
		// Validation errors bubble up from inside the transaction.  Unwrap them
		// here so they produce a 422 rather than a 500.
		var verrs validate.ValidationErrors
		if errors.As(txErr, &verrs) {
			return validationResponse(c, verrs)
		}
		return fiberErr(txErr)
	}

	return c.JSON(result)
}

// delete handles DELETE /:id.
//
// Flow:
//  1. Load the record.
//  2. Enforce delete permission.
//  3. Enforce org-unit scope.
//  4. Run HookBeforeDelete → persist deletion → run HookAfterDelete.
//  5. Return 204 No Content.
//
// Previously this method incorrectly fired HookAfterSave post-deletion;
// it now correctly fires HookAfterDelete.
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

		if err := h.enforcer.Allow(c.Context(), h.def, viewer, def.OpDelete, rec); err != nil {
			return errForbidden
		}

		if err := h.enforceOrgScope(c.Context(), tenantID, viewer, rec); err != nil {
			return err
		}

		mut := &def.Mutation{
			Op:       def.OpDelete,
			Before:   rec,
			TenantID: tenantID.String(),
			ActorID:  viewer.ActorID(),
		}

		if err := h.hooks.Run(c.Context(), h.def, mut, def.HookBeforeDelete); err != nil {
			return err
		}
		if err := es.Delete(c.Context(), id); err != nil {
			return err
		}
		// NOTE: previously HookAfterSave was fired here — that was a bug.
		// HookAfterDelete is the correct stage for post-deletion side-effects
		// (audit trail, search index removal, cascade cleanup).
		return h.hooks.Run(c.Context(), h.def, mut, def.HookAfterDelete)
	}); err != nil {
		return fiberErr(err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
