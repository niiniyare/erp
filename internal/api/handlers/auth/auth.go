package auth

import (
	stderrors "errors"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/niiniyare/erp/internal/core/identity/session"
	sharedErrors "github.com/niiniyare/erp/internal/shared/errors"
)

// LoginConfig holds tunable settings for the Login handler.
type LoginConfig struct {
	// CookieName is the HttpOnly cookie set on successful login.
	CookieName string
	// CookieDomain scopes the cookie (empty = same origin).
	CookieDomain string
	// SecureCookie enforces the Secure flag (should be true in production).
	// NOTE(settings): can be read from platform/config.AuthConfig.RequireHTTPS
	SecureCookie bool
}

// DefaultLoginConfig returns safe defaults.
func DefaultLoginConfig() LoginConfig {
	return LoginConfig{
		CookieName:   "session",
		SecureCookie: true,
	}
}

// loginRequest is the expected request body for POST /auth/login.
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginHandler handles POST /auth/login.
// On success it returns the ResolvedSession JSON and sets an HttpOnly cookie
// carrying the raw token. The raw token is NEVER included in the JSON body.
//
// NOTE(tenant-context): The tenant must already be resolved by the
// ResolveTenant middleware (which sets cache.TenantIDKey in ctx) before
// this handler runs.
func LoginHandler(svc session.Service, cfg LoginConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req loginRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid request body",
			})
		}

		req.Email = strings.TrimSpace(req.Email)
		if req.Email == "" || req.Password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "email and password are required",
			})
		}

		resolved, rawToken, err := svc.Login(c.Context(), req.Email, req.Password)
		if err != nil {
			return mapAuthError(c, err)
		}

		// Set HttpOnly session cookie — raw token never exposed in JSON.
		c.Cookie(&fiber.Cookie{
			Name:     cfg.CookieName,
			Value:    rawToken,
			HTTPOnly: true,
			Secure:   cfg.SecureCookie,
			SameSite: "Lax",
			Domain:   cfg.CookieDomain,
			Expires:  time.Now().Add(8 * time.Hour), // TODO(settings): use session TTL from config
		})

		return c.Status(fiber.StatusOK).JSON(resolved)
	}
}

// LogoutHandler handles POST /auth/logout.
// It invalidates the session in the DB + cache and clears the cookie.
func LogoutHandler(svc session.Service, cookieName string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Cookies(cookieName)
		if token == "" {
			// Also accept Authorization: Bearer header for API clients.
			if h := c.Get(fiber.HeaderAuthorization); strings.HasPrefix(h, "Bearer ") {
				token = strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
			}
		}

		if token != "" {
			// Best-effort: invalidate even if there is a transient error.
			_ = svc.Logout(c.Context(), token)
		}

		// Clear the cookie regardless.
		c.ClearCookie(cookieName)

		return c.Status(fiber.StatusNoContent).Send(nil)
	}
}

// mapAuthError converts identity/session errors to appropriate HTTP responses.
// We deliberately use generic messages for auth failures to prevent oracle attacks.
func mapAuthError(c *fiber.Ctx, err error) error {
	// 423 Locked for brute-force lockouts (checked before generic ToHTTPError
	// because ErrAccountLocked is registered as 403 in the shared errors package).
	if stderrors.Is(err, sharedErrors.ErrAccountLocked) {
		return c.Status(fiber.StatusLocked).JSON(fiber.Map{
			"error": "account is temporarily locked, please try again later",
		})
	}

	httpErr := sharedErrors.ToHTTPError(err)
	if httpErr == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "internal server error",
		})
	}

	switch httpErr.Status {
	case http.StatusUnauthorized:
		// ErrInvalidCredentials / ErrAuthenticationFailed — no oracle
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid email or password",
		})
	default:
		return c.Status(httpErr.Status).JSON(fiber.Map{
			"error": httpErr.Message,
			"code":  httpErr.Code,
		})
	}
}
