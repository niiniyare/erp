package middleware

import (
	"github.com/gofiber/fiber/v2"
	"awo/internal/core/authz"
	"awo/internal/core/identity/session"
)

// AuthorizationConfig configures legacy authorization middleware behaviour.
// Used by MiddlewareStack and SecurityValidator; for new routes use Authorize().
type AuthorizationConfig struct {
	DefaultDeny           bool     `json:"default_deny"`
	RequireAuthentication bool     `json:"require_authentication"`
	EndpointRules         []string `json:"endpoint_rules,omitempty"`
}

// DefaultAuthorizationConfig returns secure defaults.
func DefaultAuthorizationConfig() AuthorizationConfig {
	return AuthorizationConfig{
		DefaultDeny:           true,
		RequireAuthentication: true,
	}
}

// Authorize returns a Fiber handler that enforces a single permission string
// using the pre-computed permission map stored in ResolvedSession.
//
// This is an O(1) map lookup — it never hits the database or the Casbin engine.
// Use it on every protected route.
//
// Example:
//
//	app.Get("/finance/invoices",
//	    middleware.Authenticate(authCfg),
//	    middleware.Authorize("finance.receivables.invoices.read"),
//	    handler.ListInvoices)
func Authorize(permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, ok := c.Locals(session.LocalsKeySession).(*session.ResolvedSession)
		if !ok || sess == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "authentication required",
			})
		}
		if !sess.Can(permission) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":      "access denied",
				"permission": permission,
			})
		}
		return c.Next()
	}
}

// AuthorizeCasbin is a thin wrapper around authz.Service.Middleware for
// management operations that need the full Casbin engine (e.g. role assignment,
// tenant admin panels). For normal request-path authz prefer Authorize().
//
// Example:
//
//	app.Post("/admin/roles",
//	    middleware.Authenticate(authCfg),
//	    middleware.AuthorizeCasbin(authzSvc, "role", "assign"),
//	    handler.AssignRole)
func AuthorizeCasbin(svc authz.Service, object, action string) fiber.Handler {
	return svc.Middleware(object, action)
}
