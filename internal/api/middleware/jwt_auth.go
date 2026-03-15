package middleware

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/niiniyare/erp/internal/core/authz"
	"github.com/niiniyare/erp/internal/core/identity/session"
	"github.com/niiniyare/erp/internal/platform/cache"
)

// AuthConfig configures the Authenticate middleware.
type AuthConfig struct {
	// SessionSvc validates tokens and resolves sessions.
	SessionSvc session.Service
	// CookieName is the HttpOnly cookie name used to carry the session token.
	// Falls back to reading the Authorization: Bearer header if the cookie is absent.
	CookieName string
	// FIXME(feature-flags): add condition.Evaluator here once the MFA feature-flag
	// integration location is decided (Service vs middleware).
}

// DefaultAuthConfig returns sensible defaults.
func DefaultAuthConfig(svc session.Service) AuthConfig {
	return AuthConfig{
		SessionSvc: svc,
		CookieName: "session",
	}
}

// Authenticate validates the session token and populates Fiber Locals:
//   - session.LocalsKeySession  → *session.ResolvedSession
//   - authz.LocalsKeyPrincipal  → authz.Principal
//
// It also injects the tenant_id into the request context via cache.TenantIDKey
// so that downstream DB queries can set the Postgres RLS tenant context.
//
// Token resolution order: HttpOnly cookie → Authorization: Bearer header.
func Authenticate(cfg AuthConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := extractToken(c, cfg.CookieName)
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "authentication required",
			})
		}

		resolved, err := cfg.SessionSvc.ValidateSession(c.Context(), token)
		if err != nil || resolved == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired session",
			})
		}

		// Store the resolved session in Fiber Locals for handlers and authz middleware.
		c.Locals(session.LocalsKeySession, resolved)

		// Store the Casbin Principal for management-operation authz.
		c.Locals(authz.LocalsKeyPrincipal, resolved.ToPrincipal())

		// Inject tenant_id into the request context for RLS / SQLC queries.
		ctx := context.WithValue(c.Context(), cache.TenantIDKey, resolved.TenantID.String())
		c.SetUserContext(ctx)

		return c.Next()
	}
}

// extractToken reads the session token from the cookie or Authorization header.
func extractToken(c *fiber.Ctx, cookieName string) string {
	if cookieName != "" {
		if t := c.Cookies(cookieName); t != "" {
			return t
		}
	}
	header := c.Get(fiber.HeaderAuthorization)
	if after, ok := strings.CutPrefix(header, "Bearer "); ok {
		return strings.TrimSpace(after)
	}
	return ""
}
