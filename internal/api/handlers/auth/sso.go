package auth

// SSO (OAuth/OIDC) HTTP handlers.
//
// Flow:
//   GET /auth/oauth/:provider?tenant_id=...
//       → OAuthBeginHandler — redirects the browser to the provider's login page.
//
//   GET /auth/oauth/:provider/callback?code=...&state=...
//       → OAuthCallbackHandler — exchanges the code, builds a session, sets cookie.

import (
	stderrors "errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"awo.so/internal/core/iam"
	sharedErrors "awo.so/internal/shared/errors"
)

// OAuthBeginHandler handles GET /auth/oauth/:provider.
// Reads tenant_id from the query string, looks up the provider config, and
// redirects the user to the identity provider's authorization URL.
//
// Query params:
//
//	tenant_id  (required) — UUID of the tenant whose SSO config to use.
func OAuthBeginHandler(ssoSvc iam.SSOService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		providerStr := c.Params("provider")
		if providerStr == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "provider is required",
			})
		}

		tenantIDStr := c.Query("tenant_id")
		if tenantIDStr == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "tenant_id query parameter is required",
			})
		}
		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid tenant_id",
			})
		}

		authURL, err := ssoSvc.BeginOAuth(c.Context(), tenantID, iam.OAuthProvider(providerStr))
		if err != nil {
			if stderrors.Is(err, sharedErrors.ErrForbidden) || stderrors.Is(err, sharedErrors.ErrUnauthorized) {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "SSO provider not configured for this tenant",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "internal server error",
			})
		}

		return c.Redirect(authURL, fiber.StatusFound)
	}
}

// OAuthCallbackHandler handles GET /auth/oauth/:provider/callback.
// Receives the OAuth code + state, exchanges them for a full session, and
// sets the session cookie. Intended as a browser redirect target.
//
// On success: sets HttpOnly cookie and returns the ResolvedSession JSON.
// On error:   returns appropriate HTTP error; caller should show login page.
func OAuthCallbackHandler(ssoSvc iam.SSOService, sessionSvc iam.SessionService, cfg LoginConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		providerStr := c.Params("provider")
		code := c.Query("code")
		state := c.Query("state")

		if providerStr == "" || code == "" || state == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "provider, code, and state are required",
			})
		}

		// Exchange code → user.
		user, err := ssoSvc.ResolveUser(c.Context(), iam.OAuthProvider(providerStr), code, state)
		if err != nil {
			if stderrors.Is(err, sharedErrors.ErrUnauthorized) {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "invalid or expired OAuth state",
				})
			}
			if stderrors.Is(err, sharedErrors.ErrForbidden) {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "SSO login denied: your account is not provisioned for this tenant",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "internal server error",
			})
		}

		// Build full session (MFA skipped — IdP is the second factor).
		resolved, rawToken, err := sessionSvc.LoginWithSSO(c.Context(), user)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "internal server error",
			})
		}

		// Set HttpOnly session cookie — same policy as password login.
		c.Cookie(&fiber.Cookie{
			Name:     cfg.CookieName,
			Value:    rawToken,
			HTTPOnly: true,
			Secure:   cfg.SecureCookie,
			SameSite: "Lax",
			Domain:   cfg.CookieDomain,
			Expires:  time.Now().Add(8 * time.Hour),
		})

		return c.Status(fiber.StatusOK).JSON(resolved)
	}
}
