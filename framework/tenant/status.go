// Package tenant provides tenant lifecycle enforcement for the API layer.
// It defines the TenantStatus type, the StatusService interface, and the
// StatusMiddleware Fiber handler that gates requests by tenant status.
package tenant

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// TenantStatus represents the lifecycle state of a tenant account.
type TenantStatus string

const (
	StatusPending   TenantStatus = "PENDING"
	StatusActive    TenantStatus = "ACTIVE"
	StatusSuspended TenantStatus = "SUSPENDED"
	StatusArchived  TenantStatus = "ARCHIVED"
)

// ErrTenantNotFound is returned by StatusService when the tenant UUID is unknown.
var ErrTenantNotFound = errors.New("tenant not found")

// StatusService resolves and updates the lifecycle status of a tenant.
// Implementations are expected to cache reads (e.g. Redis with 5-minute TTL)
// and write-through on status change.
//
// Implementations must be safe for concurrent use.
type StatusService interface {
	// GetStatus returns the current lifecycle status for the tenant.
	// Returns ErrTenantNotFound when the tenantID is unknown.
	GetStatus(ctx context.Context, tenantID uuid.UUID) (TenantStatus, error)

	// SetStatus persists the new status and invalidates any cached value.
	// The caller (provisioning workflow / admin handler) is responsible for
	// validating state-machine transitions before calling SetStatus.
	SetStatus(ctx context.Context, tenantID uuid.UUID, status TenantStatus) error
}

// statusErr is the JSON error body used by StatusMiddleware responses.
type statusErr struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func newStatusErr(code, message string) statusErr {
	e := statusErr{}
	e.Error.Code = code
	e.Error.Message = message
	return e
}

// StatusMiddleware enforces tenant lifecycle status on every tenant-scoped API
// request. It must be placed AFTER TenantResolutionMiddleware so that
// c.Locals("tenant_id") is populated.
//
// HTTP status mapping (per docs):
//
//	PENDING   → 503 Service Unavailable (Retry-After: 60)
//	SUSPENDED → 402 Payment Required
//	ARCHIVED  → 410 Gone
//	unknown   → 404 Not Found
func StatusMiddleware(svc StatusService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		raw := c.Locals("tenant_id")
		if raw == nil {
			// No tenant context — global route, skip enforcement.
			return c.Next()
		}
		tenantID, ok := raw.(uuid.UUID)
		if !ok || tenantID == uuid.Nil {
			return c.Next()
		}

		status, err := svc.GetStatus(c.Context(), tenantID)
		if err != nil {
			if errors.Is(err, ErrTenantNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(
					newStatusErr("tenant_not_found", "Tenant not found"))
			}
			log.Error().
				Err(err).
				Str("tenant_id", tenantID.String()).
				Str("request_id", c.Locals("request_id").(string)).
				Msg("tenant status lookup failed")
			return c.Status(fiber.StatusInternalServerError).JSON(
				newStatusErr("internal_error", "Could not verify tenant status"))
		}

		switch status {
		case StatusActive:
			return c.Next()

		case StatusPending:
			c.Set("Retry-After", "60")
			return c.Status(fiber.StatusServiceUnavailable).JSON(
				newStatusErr("tenant_pending", "Tenant account is still provisioning"))

		case StatusSuspended:
			return c.Status(fiber.StatusPaymentRequired).JSON(
				newStatusErr("tenant_suspended", "Tenant account is suspended"))

		case StatusArchived:
			return c.Status(fiber.StatusGone).JSON(
				newStatusErr("tenant_archived", "Tenant account has been archived"))

		default:
			return c.Status(fiber.StatusNotFound).JSON(
				newStatusErr("tenant_not_found", "Tenant not found"))
		}
	}
}
