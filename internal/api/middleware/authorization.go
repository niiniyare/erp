package middleware

import (
	"awo.so/internal/core/iam"

	"github.com/gofiber/fiber/v2"
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

// AuthorizeCasbin is a Fiber middleware that enforces object+action using the
// full Casbin engine. Reads the Principal from c.Locals(iam.LocalsKeyPrincipal).
//
// For normal request-path authz (O(1) permission map lookup) prefer Authorize()
// from session_middleware.go. Use AuthorizeCasbin only for management operations
// that require the Casbin rule engine (e.g. role assignment, admin panels).
//
// Example:
//
//	app.Post("/admin/roles",
//	    middleware.Authenticate(authCfg),
//	    middleware.AuthorizeCasbin(authzSvc, "role", "assign"),
//	    handler.AssignRole)
func AuthorizeCasbin(svc iam.AuthzService, object, action string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		p, ok := c.Locals(iam.LocalsKeyPrincipal).(iam.Principal)
		if !ok || p.Subject == "" {
			return fiber.NewError(fiber.StatusUnauthorized, iam.ErrUnauthorized.Error())
		}

		obj := object
		if id := c.Params("id"); id != "" {
			obj = object + "/" + id
		}

		allowed, err := svc.Enforce(c.Context(), iam.Request{
			Subject: p.Subject,
			Domain:  p.Domain,
			Object:  obj,
			Action:  action,
		})
		if err != nil {
			return err
		}
		if !allowed {
			return fiber.NewError(fiber.StatusForbidden, iam.ErrForbidden.Error())
		}
		return c.Next()
	}
}
