package api

// context.go — request-context extraction and org-unit scope enforcement.
//
// These helpers are shared by all five HTTP handlers and are kept in a
// separate file to make the scope-enforcement logic easy to audit in isolation.

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"awo.so/framework/def"
)

// extractContext resolves the [def.ViewerContext] and the canonical
// tenant UUID from the incoming Fiber request.
//
// The viewer is obtained via the host-supplied [ViewerFromCtx] function.  If
// the viewer's TenantID is not a valid UUID (e.g. a dev slug such as
// "acme-demo"), the optional [TenantResolver] is invoked.  If no resolver is
// configured, a non-UUID TenantID causes a 401.
//
// A nil tenantID (uuid.Nil after successful resolution) is rejected as a
// misconfigured auth middleware — the host must ensure every authenticated
// request carries a resolvable tenant.
func (h *Handler) extractContext(c *fiber.Ctx) (def.ViewerContext, uuid.UUID, error) {
	viewer, err := h.viewer(c)
	if err != nil {
		// Do not expose the underlying error to the caller; it may contain
		// internal details about the auth middleware.
		return nil, uuid.Nil, fiber.ErrUnauthorized
	}

	tenantID, err := uuid.Parse(viewer.TenantID())
	if err != nil {
		// TenantID is not a UUID — delegate to the resolver when one is set.
		if h.tenantResolver == nil {
			return nil, uuid.Nil, fiber.NewError(
				fiber.StatusUnauthorized,
				"tenant ID is not a valid UUID and no resolver is configured",
			)
		}

		tenantID, err = h.tenantResolver(c.Context(), viewer.TenantID())
		if err != nil {
			return nil, uuid.Nil, fiber.NewError(
				fiber.StatusUnauthorized,
				"could not resolve tenant: "+err.Error(),
			)
		}
	}

	if tenantID == uuid.Nil {
		// uuid.Nil is used as a sentinel throughout the codebase to indicate
		// "no tenant".  A request that resolves to Nil is a misconfiguration.
		return nil, uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "missing tenant")
	}

	return viewer, tenantID, nil
}

// enforceOrgScope verifies that rec is accessible to viewer when the entity is
// unit-scoped and an [org.Tree] has been configured.
//
// The check is intentionally skipped in three cases:
//  1. The entity def is not unit-scoped (IsUnitScoped() == false).
//  2. No org.Tree was configured on the handler (h.orgTree == nil).
//  3. The viewer is a tenant-wide principal (OrgUnitID() == uuid.Nil) — such
//     viewers are permitted to access records in any unit.
//
// When the check runs, the record must implement the OrgUnitID() accessor
// (typically provided by the persistence layer's mapRecord type).  If it does
// not, access is denied as a safe default.
//
// This method is called from findByID, update, and delete to close the
// ID-enumeration gap: without per-record checks, a scoped viewer could access
// records outside their subtree by guessing UUIDs, even though the list
// endpoint would not surface those records.
func (h *Handler) enforceOrgScope(
	ctx context.Context,
	tenantID uuid.UUID,
	viewer def.ViewerContext,
	rec def.Record,
) error {
	// Fast path: no unit-scope enforcement needed.
	if !h.def.IsUnitScoped() || h.orgTree == nil || viewer.OrgUnitID() == uuid.Nil {
		return nil
	}

	// Extract the record's owning org unit.
	scoped, ok := rec.(interface{ OrgUnitID() uuid.UUID })
	if !ok {
		// The record does not expose OrgUnitID — deny access as a safe default
		// rather than silently granting it.
		return errForbidden
	}
	recordUnit := scoped.OrgUnitID()

	// A record with no assigned unit is accessible to any scoped viewer.
	// This handles cases where unit assignment is optional on the entity.
	if recordUnit == uuid.Nil {
		return nil
	}

	// Resolve the viewer's permitted subtree and check membership.
	descendants, err := h.orgTree.Descendants(ctx, tenantID, viewer.OrgUnitID())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "org tree lookup failed")
	}

	for _, allowed := range descendants {
		if allowed == recordUnit {
			return nil
		}
	}

	// Record's unit is outside the viewer's permitted subtree.
	return errForbidden
}
