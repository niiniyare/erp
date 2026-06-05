package auth

import (
	stderrors "errors"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"awo.so/internal/core/iam"
	"awo.so/internal/shared"
	sharedErrors "awo.so/internal/shared/errors"
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
	// TenantID is required when the X-Tenant-ID header is absent.
	// Pass the tenant UUID to scope authentication to a specific tenant.
	TenantID string `json:"tenant_id"`
}

// LoginHandler handles POST /auth/login.
// On success it returns the ResolvedSession JSON and sets an HttpOnly cookie
// carrying the raw token. The raw token is NEVER included in the JSON body.
//
// NOTE(tenant-context): The tenant must already be resolved by the
// ResolveTenant middleware (which sets cache.TenantIDKey in ctx) before
// this handler runs.
func LoginHandler(svc iam.SessionService, cfg LoginConfig) fiber.Handler {
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

		// Inject tenant context so IAM repository can scope the user lookup.
		// Priority: X-Tenant-ID header, then request body tenant_id.
		tenantIDStr := strings.TrimSpace(c.Get("X-Tenant-ID"))
		if tenantIDStr == "" {
			tenantIDStr = strings.TrimSpace(req.TenantID)
		}
		if tenantIDStr != "" {
			if tid, parseErr := uuid.Parse(tenantIDStr); parseErr == nil {
				c.SetUserContext(shared.WithTenantID(c.UserContext(), tid))
			}
		}

		resolved, rawToken, err := svc.Login(c.UserContext(), req.Email, req.Password)
		if err != nil {
			// MFA step 1: Login succeeded but a second factor is required.
			// rawToken is a short-lived pending token (not a session token).
			if stderrors.Is(err, sharedErrors.ErrMFARequired) {
				return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
					"mfa_required":  true,
					"pending_token": rawToken,
				})
			}
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

		// Include redirect_to so the browser login page can navigate back to
		// the originally requested URL after a successful sign-in.
		// Only echo back relative paths (starts with "/") to prevent open redirect.
		redirectTo := "/ui/demo" // safe fallback
		if r := c.Query("redirect"); strings.HasPrefix(r, "/") && !strings.HasPrefix(r, "//") {
			redirectTo = r
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"data":        resolved,
			"redirect_to": redirectTo,
		})
	}
}

// LogoutHandler handles POST /auth/logout.
// It invalidates the session in the DB + cache and clears the cookie.
func LogoutHandler(svc iam.SessionService, cookieName string) fiber.Handler {
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
		// Return as a Fiber error so the global error handler logs it.
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
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
