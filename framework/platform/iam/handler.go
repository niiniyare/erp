package iam

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Handler holds the IAM HTTP handler methods.
type Handler struct {
	svc *AuthService
}

// NewHandler creates a Handler backed by svc.
func NewHandler(svc *AuthService) *Handler {
	return &Handler{svc: svc}
}

// Mount registers auth routes on r.
//
//	POST /auth/login    — issue access + refresh token pair
//	POST /auth/logout   — revoke access token
//	POST /auth/refresh  — rotate tokens using a refresh token
func (h *Handler) Mount(r fiber.Router) {
	r.Post("/auth/login", h.login)
	r.Post("/auth/logout", h.logout)
	r.Post("/auth/refresh", h.refresh)
}

// ── Request / response types ───────────────────────────────────────────────

type loginRequest struct {
	TenantID string `json:"tenant_id"` // alternative to X-Awo-Tenant header
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken          string    `json:"access_token"`
	RefreshToken         string    `json:"refresh_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
	UserID               string    `json:"user_id"`
	TenantID             string    `json:"tenant_id"`
	Roles                []string  `json:"roles"`
	Permissions          []string  `json:"permissions"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// ── Handlers ──────────────────────────────────────────────────────────────

func (h *Handler) login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if req.Email == "" || req.Password == "" {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "email and password required")
	}

	// Resolve tenant: X-Awo-Tenant header takes precedence over body field.
	tenantStr := c.Get("X-Awo-Tenant")
	if tenantStr == "" {
		tenantStr = req.TenantID
	}
	tenantID, err := uuid.Parse(tenantStr)
	if err != nil || tenantID == uuid.Nil {
		return fiber.NewError(fiber.StatusBadRequest, "X-Awo-Tenant header or tenant_id body field required")
	}

	result, err := h.svc.Login(c.Context(), tenantID, req.Email, req.Password, c.IP())
	if err != nil {
		return mapAuthError(err)
	}

	return c.Status(fiber.StatusOK).JSON(loginResponse{
		AccessToken:          result.AccessToken,
		RefreshToken:         result.RefreshToken,
		AccessTokenExpiresAt: result.Session.AccessExpiresAt,
		UserID:               result.Session.UserID.String(),
		TenantID:             result.Session.TenantID.String(),
		Roles:                result.Session.Roles,
		Permissions:          result.Session.Permissions,
	})
}

func (h *Handler) logout(c *fiber.Ctx) error {
	rawToken := extractBearer(c)
	if rawToken == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "no token provided")
	}
	if err := h.svc.Logout(c.Context(), rawToken); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "logout failed")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) refresh(c *fiber.Ctx) error {
	var req refreshRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if req.RefreshToken == "" {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "refresh_token required")
	}

	result, err := h.svc.Refresh(c.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired refresh token")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "refresh failed")
	}

	return c.Status(fiber.StatusOK).JSON(loginResponse{
		AccessToken:          result.AccessToken,
		RefreshToken:         result.RefreshToken,
		AccessTokenExpiresAt: result.Session.AccessExpiresAt,
		UserID:               result.Session.UserID.String(),
		TenantID:             result.Session.TenantID.String(),
		Roles:                result.Session.Roles,
		Permissions:          result.Session.Permissions,
	})
}

// mapAuthError converts service-layer auth errors to Fiber HTTP errors.
func mapAuthError(err error) error {
	switch {
	case errors.Is(err, ErrUserNotFound), errors.Is(err, ErrInvalidPassword):
		// Never distinguish which field was wrong — prevents user enumeration.
		return fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")
	case errors.Is(err, ErrUserLocked):
		return fiber.NewError(fiber.StatusTooManyRequests, "account temporarily locked — try again in 15 minutes")
	case errors.Is(err, ErrUserInactive):
		return fiber.NewError(fiber.StatusForbidden, "account is not active")
	}
	return fiber.NewError(fiber.StatusInternalServerError, "login failed")
}
