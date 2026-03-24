package auth

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"awo.so/internal/core/iam"
)

// ─── Request / Response types ─────────────────────────────────────────────────

type mfaCompleteRequest struct {
	PendingToken string `json:"pending_token"`
	Code         string `json:"code"`
}

type mfaConfirmRequest struct {
	Code string `json:"code"`
}

type mfaDisableRequest struct {
	Password string `json:"password"`
}

// ─── Handlers ─────────────────────────────────────────────────────────────────

// MFAInitiateHandler handles POST /auth/mfa/initiate.
// Requires an active session (Authenticate middleware must run before this).
// Generates a fresh TOTP secret, caches it pending confirmation, and returns
// the provisioning URI that the user scans with their authenticator app.
func MFAInitiateHandler(userSvc iam.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, ok := c.Locals(iam.LocalsKeySession).(*iam.ResolvedSession)
		if !ok || sess == nil {
			return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
		}

		setup, err := userSvc.InitiateMFA(c.Context(), sess.UserID)
		if err != nil {
			return mapMFAError(c, err)
		}

		// Return the TOTP URI for QR code generation by the client.
		// The raw secret is included so the user can also enter it manually.
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"qr_uri": setup.QRURI,
			"secret": setup.Secret,
		})
	}
}

// MFAConfirmHandler handles POST /auth/mfa/confirm.
// Requires an active session (Authenticate middleware must run before this).
// Validates the first TOTP code and, on success, persists the secret to the DB
// and marks MFA as enabled for the user.
func MFAConfirmHandler(userSvc iam.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, ok := c.Locals(iam.LocalsKeySession).(*iam.ResolvedSession)
		if !ok || sess == nil {
			return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
		}

		var req mfaConfirmRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid request body",
			})
		}
		req.Code = strings.TrimSpace(req.Code)
		if req.Code == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "code is required",
			})
		}

		if err := userSvc.ConfirmMFA(c.Context(), sess.UserID, req.Code); err != nil {
			return mapMFAError(c, err)
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "MFA enabled successfully",
		})
	}
}

// MFACompleteHandler handles POST /auth/mfa/complete.
// Public endpoint — no existing session required.
// Exchanges a pending MFA token + TOTP code for a full session.
// Called after Login returns {mfa_required: true, pending_token: "..."}.
func MFACompleteHandler(sessionSvc iam.SessionService, cfg LoginConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req mfaCompleteRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid request body",
			})
		}

		req.PendingToken = strings.TrimSpace(req.PendingToken)
		req.Code = strings.TrimSpace(req.Code)

		if req.PendingToken == "" || req.Code == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "pending_token and code are required",
			})
		}

		resolved, rawToken, err := sessionSvc.CompleteMFALogin(c.Context(), req.PendingToken, req.Code)
		if err != nil {
			return mapAuthError(c, err)
		}

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

// MFADisableHandler handles DELETE /auth/mfa.
// Requires an active session (Authenticate middleware must run before this).
// Re-verifies the user's password before clearing the MFA secret.
func MFADisableHandler(userSvc iam.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, ok := c.Locals(iam.LocalsKeySession).(*iam.ResolvedSession)
		if !ok || sess == nil {
			return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
		}

		var req mfaDisableRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid request body",
			})
		}
		if req.Password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "password is required",
			})
		}

		if err := userSvc.DisableMFA(c.Context(), sess.UserID, req.Password); err != nil {
			return mapMFAError(c, err)
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "MFA disabled successfully",
		})
	}
}

// ─── Helper ───────────────────────────────────────────────────────────────────

// mapMFAError maps MFA-specific service errors to HTTP responses.
func mapMFAError(c *fiber.Ctx, err error) error {
	// Delegate to the shared auth error mapper which handles BusinessError unwrapping.
	return mapAuthError(c, err)
}

// userIDFromLocals extracts the UserID from the session stored in Fiber locals.
// Returns uuid.Nil if no session is present (should not happen after Authenticate middleware).
func userIDFromLocals(c *fiber.Ctx) uuid.UUID {
	if sess, ok := c.Locals(iam.LocalsKeySession).(*iam.ResolvedSession); ok && sess != nil {
		return sess.UserID
	}
	return uuid.Nil
}

// keep compiler from complaining about unused userIDFromLocals
var _ = userIDFromLocals
