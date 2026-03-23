package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"awo.so/internal/core/iam"
)

// AuthConfig holds configuration for session-based authentication middleware.
type AuthConfig struct {
	// SessionService validates tokens and returns ResolvedSessions.
	SessionService iam.SessionService
	// CookieName is the HttpOnly cookie that carries the raw session token.
	CookieName string
}

// DefaultAuthConfig returns an AuthConfig with safe defaults backed by svc.
func DefaultAuthConfig(svc iam.SessionService) AuthConfig {
	return AuthConfig{
		SessionService: svc,
		CookieName:     "session",
	}
}

// Authenticate is a Fiber middleware that validates the session token from
// the cookie (or Authorization: Bearer header) and stores the ResolvedSession
// in c.Locals(iam.LocalsKeySession).
//
// Returns 401 if the token is missing or invalid.
func Authenticate(cfg AuthConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Cookies(cfg.CookieName)
		if token == "" {
			if h := c.Get(fiber.HeaderAuthorization); strings.HasPrefix(h, "Bearer ") {
				token = strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
			}
		}
		if token == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
		}

		resolved, err := cfg.SessionService.ValidateSession(c.Context(), token)
		if err != nil || resolved == nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired session")
		}

		c.Locals(iam.LocalsKeySession, resolved)
		return c.Next()
	}
}

// Authorize returns a Fiber middleware that checks whether the authenticated
// session holds the given permission key (e.g. "finance.accounts.read").
//
// Must run after Authenticate (requires LocalsKeySession to be set).
// Returns 403 if the permission is absent.
func Authorize(permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, ok := c.Locals(iam.LocalsKeySession).(*iam.ResolvedSession)
		if !ok || sess == nil {
			return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
		}
		if !sess.Can(permission) {
			return fiber.NewError(fiber.StatusForbidden, "permission denied")
		}
		return c.Next()
	}
}
