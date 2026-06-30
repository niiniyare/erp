package api

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"awo.so/framework/def"
	"awo.so/framework/persistence"
)

// ActionContext is passed to ActionDef.Fn by the action handler.
// It provides pre-resolved, permission-checked access to the entity and actor.
type ActionContext struct {
	// RecordID is the UUID of the target record from the URL parameter.
	RecordID uuid.UUID

	// Actor is the authenticated viewer performing the action.
	Actor def.ViewerContext

	// TenantID is the resolved tenant UUID for this request.
	TenantID uuid.UUID

	// Input is the parsed JSON body of the action request (may be nil for no-body actions).
	Input map[string]any

	// Store is an EntityStore scoped to this entity, tenant, and transaction.
	// Available only within WithTx — nil outside.
	Store persistence.EntityStore

	// Ctx is the request context with cancellation propagated from the HTTP connection.
	Ctx context.Context //nolint:containedctx
}

// ActionResult is returned by ActionDef.Fn on success.
type ActionResult struct {
	// Message is a short human-readable confirmation shown in the SDUI toast.
	Message string `json:"message,omitempty"`

	// Data is optional structured payload returned in the response data field.
	Data any `json:"data,omitempty"`

	// WorkflowID is set when the action triggered a Temporal workflow.
	// Results in HTTP 202 Accepted instead of 200 OK.
	WorkflowID string `json:"workflow_id,omitempty"`
}

// RegisterActions mounts action routes for def.Actions under the entity's prefix.
// Called by bootstrap after RegisterEntity mounts the base CRUD routes.
//
// Each action registers: POST /{prefix}/{action.Name}
// The full path from the router root becomes: /api/{tableName}/:id/{action.Name}
func RegisterActions(
	router fiber.Router,
	entDef *def.EntityDefinition,
	store persistence.TenantStore,
	viewerFn ViewerFromCtx,
	opts ...HandlerOption,
) {
	if len(entDef.Actions) == 0 {
		return
	}

	h := &Handler{
		def:    entDef,
		store:  store,
		viewer: viewerFn,
	}
	for _, o := range opts {
		o(h)
	}

	for _, action := range entDef.Actions {
		action := action // capture loop var
		router.Post("/:id/"+action.Name, makeActionHandler(h, action))
	}
}

// makeActionHandler returns a Fiber handler for one ActionDef.
func makeActionHandler(h *Handler, action def.ActionDef) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if action.Fn == nil {
			return Err(c, fiber.StatusNotImplemented, "not_implemented",
				fmt.Sprintf("action %q has no handler", action.Name))
		}

		recordID, err := parseUUID(c, "id")
		if err != nil {
			return err
		}

		viewer, tenantID, err := h.extractContext(c)
		if err != nil {
			return err
		}

		// Parse optional JSON body.
		var input map[string]any
		if len(c.Body()) > 0 {
			if err := c.BodyParser(&input); err != nil {
				return Err(c, fiber.StatusBadRequest, "bad_request", "invalid JSON body")
			}
		}

		var result any
		txErr := h.store.WithTx(c.Context(), tenantID, func(tx persistence.TenantTx) error {
			es := tx.ForEntity(h.def.Name)

			rec, err := es.FindByID(c.Context(), recordID)
			if err != nil {
				return err
			}
			if err := h.enforcer.Allow(c.Context(), h.def, viewer, def.OpUpdate, rec); err != nil {
				return errForbidden
			}

			mut := &def.Mutation{
				Op:       def.OpUpdate,
				Before:   rec,
				TenantID: tenantID.String(),
				ActorID:  viewer.ActorID(),
			}

			result, err = action.Fn(c.Context(), mut)
			return err
		})
		if txErr != nil {
			return fiberErr(txErr)
		}

		// If result carries a WorkflowID, respond 202 Accepted.
		if ar, ok := result.(*ActionResult); ok && ar != nil && ar.WorkflowID != "" {
			return Accepted(c, ar, fiber.Map{"workflow_id": ar.WorkflowID})
		}
		return OK(c, result)
	}
}
