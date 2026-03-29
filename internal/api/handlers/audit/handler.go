// Package audit provides HTTP handlers for the audit log API.
package audit

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"awo.so/internal/core/audit"
	"awo.so/internal/core/iam"
	sharedErrors "awo.so/internal/shared/errors"
)

// ListAuditEventsHandler returns paginated audit events for the authenticated tenant.
//
// Route: GET /api/v1/audit-logs  (requires Authenticate middleware)
// Permission gate: iam.sessions.read
//
// Query params:
//
//	limit  int  — default 50, max 1000
//	offset int  — default 0
func ListAuditEventsHandler(svc audit.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, ok := c.Locals(iam.LocalsKeySession).(*iam.ResolvedSession)
		if !ok || sess == nil {
			return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
		}

		if !sess.Can("iam.sessions.read") {
			return fiber.NewError(fiber.StatusForbidden, "insufficient permissions")
		}

		limit := 50
		if l := c.Query("limit"); l != "" {
			if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 1000 {
				limit = v
			}
		}

		offset := 0
		if o := c.Query("offset"); o != "" {
			if v, err := strconv.Atoi(o); err == nil && v >= 0 {
				offset = v
			}
		}

		filters := audit.AuditEventFilters{
			Limit:  limit,
			Offset: offset,
		}

		events, err := svc.GetAuditEvents(c.Context(), sess.TenantID, filters)
		if err != nil {
			httpErr := sharedErrors.ToHTTPError(err)
			if httpErr == nil {
				return fiber.ErrInternalServerError
			}
			return c.Status(httpErr.Status).JSON(httpErr)
		}

		return c.JSON(fiber.Map{
			"data":   events,
			"limit":  limit,
			"offset": offset,
		})
	}
}
