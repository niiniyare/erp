package auth

import (
	stderrors "errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"awo.so/internal/core/iam"
	sharedErrors "awo.so/internal/shared/errors"
)

// Request types

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

type resetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// Handlers

// ForgotPasswordHandler handles POST /auth/forgot-password.
// Always returns 200 regardless of whether the email exists — this prevents
// user enumeration attacks. The service generates a token and the caller is
// responsible for emailing it; here we only log/emit the token for now
// (notification service integration is a TODO).
//
// NOTE(notification): In production, wire a notification service here to send
// the reset link by email. Until then the token is silently discarded.
func ForgotPasswordHandler(userSvc iam.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req forgotPasswordRequest
		if err := c.BodyParser(&req); err != nil {
			// Return 200 even on bad input — don't reveal anything
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"message": "If that email is registered you will receive a reset link",
			})
		}

		req.Email = strings.TrimSpace(strings.ToLower(req.Email))
		if req.Email == "" {
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"message": "If that email is registered you will receive a reset link",
			})
		}

		// ForgotPassword returns ("", uuid.Nil, nil) when email not found.
		// We intentionally ignore rawToken — emit via notification service when available.
		_, _, _ = userSvc.ForgotPassword(c.Context(), req.Email)

		// Always 200 — never reveal whether the email exists.
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "If that email is registered you will receive a reset link",
		})
	}
}

// ResetPasswordHandler handles POST /auth/reset-password.
// Validates the token and sets a new password. Specific error codes are
// returned so the client can show an appropriate message (expired, already
// used, too weak, recently used).
func ResetPasswordHandler(userSvc iam.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req resetPasswordRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid request body",
			})
		}

		req.Token = strings.TrimSpace(req.Token)
		req.NewPassword = strings.TrimSpace(req.NewPassword)

		if req.Token == "" || req.NewPassword == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "token and new_password are required",
			})
		}

		if err := userSvc.ResetPassword(c.Context(), req.Token, req.NewPassword); err != nil {
			return mapPasswordResetError(c, err)
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Password reset successfully. Please log in with your new password.",
		})
	}
}

// Error mapping

func mapPasswordResetError(c *fiber.Ctx, err error) error {
	switch {
	case stderrors.Is(err, sharedErrors.ErrPasswordResetTokenNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Reset link not found. Please request a new one.",
			"code":  sharedErrors.CodePasswordResetTokenNotFound,
		})
	case stderrors.Is(err, sharedErrors.ErrPasswordResetTokenExpired):
		return c.Status(fiber.StatusGone).JSON(fiber.Map{
			"error": "Reset link has expired. Please request a new one.",
			"code":  sharedErrors.CodePasswordResetTokenExpired,
		})
	case stderrors.Is(err, sharedErrors.ErrPasswordResetTokenUsed):
		return c.Status(fiber.StatusGone).JSON(fiber.Map{
			"error": "Reset link has already been used. Please request a new one.",
			"code":  sharedErrors.CodePasswordResetTokenUsed,
		})
	case stderrors.Is(err, sharedErrors.ErrPasswordTooWeak):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Password must be at least 12 characters with uppercase, lowercase, digit, and special character.",
			"code":  sharedErrors.CodePasswordTooWeak,
		})
	case stderrors.Is(err, sharedErrors.ErrPasswordReused):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "This password was used recently. Please choose a different one.",
			"code":  sharedErrors.CodePasswordReused,
		})
	}

	httpErr := sharedErrors.ToHTTPError(err)
	if httpErr == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "internal server error",
		})
	}
	return c.Status(httpErr.Status).JSON(fiber.Map{
		"error": httpErr.Message,
		"code":  httpErr.Code,
	})
}
