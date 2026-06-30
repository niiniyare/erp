package iam

import (
	"errors"

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

// Mount registers auth routes on r (at root, no prefix).
//
//	POST /auth/login             — issue access + refresh token pair (tenant user)
//	POST /auth/logout            — revoke access token
//	POST /auth/refresh           — rotate tokens using a refresh token
//	POST /auth/register          — create workspace (tenant) + admin user
//	GET  /auth/me                — return current session info
//	POST /auth/platform/login    — authenticate platform admin (Plane 1, no tenant required)
func (h *Handler) Mount(r fiber.Router) {
	r.Post("/auth/login", h.login)
	r.Post("/auth/logout", h.logout)
	r.Post("/auth/refresh", h.refresh)
	r.Post("/auth/register", h.register)
	r.Get("/auth/me", h.me)
	r.Post("/auth/platform/login", h.platformLogin)
}

// ── Request types ──────────────────────────────────────────────────────────

type loginRequest struct {
	TenantID string `json:"tenant_id" form:"tenant_id"` // alternative to X-Awo-Tenant header
	Email    string `json:"email"     form:"email"`
	Password string `json:"password"  form:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" form:"refresh_token"`
}

// ── Helpers ────────────────────────────────────────────────────────────────

// authOK wraps a LoginResult in the AMIS-compatible {status:0, data:{...}} envelope.
func authOK(c *fiber.Ctx, r *LoginResult) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": 0,
		"msg":    "",
		"data": fiber.Map{
			"access_token":            r.AccessToken,
			"refresh_token":           r.RefreshToken,
			"access_token_expires_at": r.Session.AccessExpiresAt,
			"user_id":                 r.Session.UserID.String(),
			"tenant_id":               r.Session.TenantID.String(),
			"display_name":            r.Session.UserID.String(), // TODO: load full_name from tenant_users
			"roles":                   r.Session.Roles,
			"permissions":             r.Session.Permissions,
			"plane":                   r.Session.Plane,
		},
	})
}

// ── Handlers ───────────────────────────────────────────────────────────────

func (h *Handler) login(c *fiber.Ctx) error {
	var req loginRequest
	_ = c.BodyParser(&req)
	if req.Email == "" {
		req.Email = c.FormValue("email")
	}
	if req.Password == "" {
		req.Password = c.FormValue("password")
	}
	if req.TenantID == "" {
		req.TenantID = c.FormValue("tenant_id")
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
	return authOK(c, result)
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
	_ = c.BodyParser(&req)
	if req.RefreshToken == "" {
		req.RefreshToken = c.FormValue("refresh_token")
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
	return authOK(c, result)
}

// register handles "Create workspace" — creates a tenant + admin user in one call.
// Body: {workspace_name, admin_email, admin_name, password}
// Response: {status:0, data:{tenant_slug, display_name}}
func (h *Handler) register(c *fiber.Ctx) error {
	var req struct {
		WorkspaceName string `json:"workspace_name" form:"workspace_name"`
		AdminEmail    string `json:"admin_email"    form:"admin_email"`
		AdminName     string `json:"admin_name"     form:"admin_name"`
		Password      string `json:"password"       form:"password"`
	}
	_ = c.BodyParser(&req)
	if req.AdminEmail == "" {
		req.AdminEmail = c.FormValue("admin_email")
	}
	if req.Password == "" {
		req.Password = c.FormValue("password")
	}
	if req.WorkspaceName == "" {
		req.WorkspaceName = c.FormValue("workspace_name")
	}
	if req.AdminName == "" {
		req.AdminName = c.FormValue("admin_name")
	}
	if req.AdminEmail == "" || req.Password == "" || req.WorkspaceName == "" {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "workspace_name, admin_email and password required")
	}

	slug, tenantID, err := h.svc.CreateWorkspace(c.Context(), req.WorkspaceName, req.AdminEmail, req.AdminName, req.Password)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "workspace creation failed: "+err.Error())
	}

	// Auto-login the new admin so the UI can proceed immediately.
	result, err := h.svc.Login(c.Context(), tenantID, req.AdminEmail, req.Password, c.IP())
	if err != nil {
		// Created but login failed — return slug so client can log in manually.
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"status": 0,
			"msg":    "workspace created",
			"data": fiber.Map{
				"tenant_slug":  slug,
				"display_name": req.AdminName,
			},
		})
	}

	data := fiber.Map{
		"tenant_slug":             slug,
		"display_name":            req.AdminName,
		"access_token":            result.AccessToken,
		"refresh_token":           result.RefreshToken,
		"access_token_expires_at": result.Session.AccessExpiresAt,
		"user_id":                 result.Session.UserID.String(),
		"tenant_id":               result.Session.TenantID.String(),
		"roles":                   result.Session.Roles,
		"permissions":             result.Session.Permissions,
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": 0, "msg": "", "data": data})
}

func (h *Handler) me(c *fiber.Ctx) error {
	v, ok := c.Locals(viewerLocalKey).(*SessionViewer)
	if !ok || v == nil {
		return fiber.NewError(fiber.StatusUnauthorized, "not authenticated")
	}
	sess := v.sess
	return c.JSON(fiber.Map{
		"status": 0,
		"msg":    "",
		"data": fiber.Map{
			"user_id":      sess.UserID.String(),
			"tenant_id":    sess.TenantID.String(),
			"display_name": sess.UserID.String(), // TODO: load full_name from tenant_users
			"roles":        sess.Roles,
			"permissions":  sess.Permissions,
			"plane":        sess.Plane,
		},
	})
}

func (h *Handler) platformLogin(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"    form:"email"`
		Password string `json:"password" form:"password"`
	}
	_ = c.BodyParser(&req)
	if req.Email == "" {
		req.Email = c.FormValue("email")
	}
	if req.Password == "" {
		req.Password = c.FormValue("password")
	}
	if req.Email == "" || req.Password == "" {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "email and password required")
	}

	result, err := h.svc.PlatformLogin(c.Context(), req.Email, req.Password, c.IP())
	if err != nil {
		return mapAuthError(err)
	}
	return authOK(c, result)
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
