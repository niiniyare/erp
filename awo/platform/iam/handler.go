// Package iam — handler.go exposes authentication endpoints over HTTP.
//
// Routes registered by RegisterAuthRoutes (called from api/router):
//
//	POST /api/v1/auth/login   — exchange credentials for session token
//	POST /api/v1/auth/logout  — revoke the current session
//	GET  /api/v1/auth/me      — return claims for the current session
package iam

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"

	"awo.so/awo/api/response"
	"awo.so/awo/runtime"
)

// RegisterAuthRoutes mounts the IAM authentication routes onto app.
// These routes intentionally bypass the RequireAuth middleware (login cannot
// require a token; logout and /me do their own token extraction).
func RegisterAuthRoutes(app *fiber.App, svc *Service) {
	auth := app.Group("/api/v1/auth")
	auth.Post("/login", loginHandler(svc))
	auth.Post("/logout", logoutHandler(svc))
	auth.Get("/me", meHandler(svc))
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func loginHandler(svc *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req loginRequest
		if err := json.Unmarshal(c.Body(), &req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(response.Wrap(&runtime.ValidationError{
				Fields: map[string]string{"_body": "invalid JSON"},
			}))
		}
		if req.Email == "" || req.Password == "" {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(response.Wrap(&runtime.ValidationError{
				Fields: map[string]string{
					"email":    ifEmpty(req.Email, "required"),
					"password": ifEmpty(req.Password, "required"),
				},
			}))
		}

		token, claims, err := svc.Login(c.UserContext(), req.Email, req.Password)
		if err != nil {
			return c.Status(response.HTTPStatus(err)).JSON(response.Wrap(err))
		}

		return c.Status(fiber.StatusOK).JSON(response.Success{
			Data: fiber.Map{
				"token":      token,
				"user_id":    claims.UserID,
				"tenant_id":  claims.TenantID,
				"expires_at": claims.ExpiresAt,
			},
		})
	}
}

func logoutHandler(svc *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := extractBearerTokenFromHeader(c)
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(response.Wrap(&runtime.BusinessError{
				Code:    "iam.missing_token",
				Message: "Authentication token is required",
				Status:  401,
			}))
		}
		if err := svc.Logout(c.UserContext(), token); err != nil {
			return c.Status(response.HTTPStatus(err)).JSON(response.Wrap(err))
		}
		return c.SendStatus(fiber.StatusNoContent)
	}
}

func meHandler(svc *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := extractBearerTokenFromHeader(c)
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(response.Wrap(&runtime.BusinessError{
				Code:    "iam.missing_token",
				Message: "Authentication token is required",
				Status:  401,
			}))
		}
		claims, err := svc.ValidateToken(c.UserContext(), token)
		if err != nil {
			return c.Status(response.HTTPStatus(err)).JSON(response.Wrap(err))
		}
		return c.JSON(response.Success{
			Data: fiber.Map{
				"user_id":    claims.UserID,
				"session_id": claims.SessionID,
				"tenant_id":  claims.TenantID,
				"email":      claims.Email,
				"roles":      claims.Roles,
				"expires_at": claims.ExpiresAt,
			},
		})
	}
}

func extractBearerTokenFromHeader(c *fiber.Ctx) string {
	auth := c.Get("Authorization")
	const prefix = "Bearer "
	if len(auth) > len(prefix) && auth[:len(prefix)] == prefix {
		return auth[len(prefix):]
	}
	return ""
}

// ifEmpty returns msg if s is empty, otherwise "".
func ifEmpty(s, msg string) string {
	if s == "" {
		return msg
	}
	return ""
}
