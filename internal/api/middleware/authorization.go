package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"awo.so/internal/core/iam/contract"
)

// AuthorizationConfig configures legacy authorization middleware behaviour.
// Used by MiddlewareStack and SecurityValidator; for new routes use Authorize()
// from session_middleware.go.
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

// =============================================================================
// Context helpers — use in handlers; never pass tenant/user IDs as params
// =============================================================================

// ContextSession returns the ResolvedSession stored by Authenticate middleware.
// Panics on unauthenticated routes — that is a programming error.
func ContextSession(c *fiber.Ctx) *contract.ResolvedSession {
	sess, ok := c.Locals(contract.LocalsKeySession).(*contract.ResolvedSession)
	if !ok || sess == nil {
		panic("middleware: ContextSession called on unauthenticated route")
	}
	return sess
}

// ContextTenantID returns the tenant UUID from the resolved session.
func ContextTenantID(c *fiber.Ctx) uuid.UUID {
	return ContextSession(c).TenantID
}

// ContextUserID returns the user UUID from the resolved session.
func ContextUserID(c *fiber.Ctx) uuid.UUID {
	return ContextSession(c).UserID
}

// ContextPrincipal returns the Casbin Principal from the resolved session.
// Use this when calling AuthorizeCasbin manually inside a handler.
func ContextPrincipal(c *fiber.Ctx) contract.Principal {
	p, _ := c.Locals(contract.LocalsKeyPrincipal).(contract.Principal)
	return p
}

// =============================================================================
// Casbin path — management operations (live rule engine, not session map)
// =============================================================================

// AuthorizeCasbin enforces object+action using the full Casbin engine.
// Reads the Principal from c.Locals(contract.LocalsKeyPrincipal).
//
// For hot-path API routes prefer Authorize() from session_middleware.go.
// Use AuthorizeCasbin only for management operations where a live Casbin
// check is required (e.g. role assignment, admin panels).
//
// Example:
//
//	app.Post("/admin/roles",
//	    middleware.Authenticate(authCfg),
//	    middleware.AuthorizeCasbin(authzSvc, "role", "assign"),
//	    handler.AssignRole)
func AuthorizeCasbin(svc contract.AuthzService, object, action string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		p, ok := c.Locals(contract.LocalsKeyPrincipal).(contract.Principal)
		if !ok || p.Subject == "" {
			return fiber.NewError(fiber.StatusUnauthorized, contract.ErrUnauthorized.Error())
		}

		obj := object
		if id := c.Params("id"); id != "" {
			obj = object + "/" + id
		}

		allowed, err := svc.Enforce(c.Context(), contract.Request{
			Subject: p.Subject,
			Domain:  p.Domain,
			Object:  obj,
			Action:  action,
		})
		if err != nil {
			return err
		}
		if !allowed {
			return fiber.NewError(fiber.StatusForbidden, contract.ErrForbidden.Error())
		}
		return c.Next()
	}
}
