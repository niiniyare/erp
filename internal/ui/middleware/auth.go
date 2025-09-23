package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/core/abac"
	"github.com/niiniyare/erp/internal/core/iam"
	"github.com/niiniyare/erp/internal/core/iam/authn"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
)

// UIRole defines the role type for UI services
type UIRole string

const (
	UIRoleConsoleAdmin UIRole = "console_admin"
	UIRoleTenantUser   UIRole = "tenant_user"
	UIRoleClientUser   UIRole = "client_user"
)

// UIContext represents the current UI authentication context
type UIContext struct {
	UserID       uuid.UUID  `json:"user_id"`
	TenantID     uuid.UUID  `json:"tenant_id"`
	EntityID     *uuid.UUID `json:"entity_id,omitempty"`
	Role         UIRole     `json:"role"`
	SessionID    string     `json:"session_id"`
	CSRFToken    string     `json:"csrf_token"`
	LoginTime    time.Time  `json:"login_time"`
	LastActivity time.Time  `json:"last_activity"`
}

// UIAuthMiddleware provides authentication middleware for UI services
type UIAuthMiddleware struct {
	iamService   iam.Service
	abacService  abac.Service
	cacheService cache.Service
	logger       logger.Logger

	// Configuration
	sessionTTL   time.Duration
	cookieName   string
	cookieDomain string
	cookieSecure bool
	csrfEnabled  bool
}

// NewUIAuthMiddleware creates a new UI authentication middleware
func NewUIAuthMiddleware(
	iamService iam.Service,
	abacService abac.Service,
	cacheService cache.Service,
	logger logger.Logger,
) *UIAuthMiddleware {
	return &UIAuthMiddleware{
		iamService:   iamService,
		abacService:  abacService,
		cacheService: cacheService,
		logger:       logger,

		// Default configuration
		sessionTTL:   24 * time.Hour,
		cookieName:   "ui_session",
		cookieDomain: "",
		cookieSecure: true,
		csrfEnabled:  true,
	}
}

// AuthenticateUser authenticates a user and creates a UI session
func (m *UIAuthMiddleware) AuthenticateUser(ctx context.Context, username, password string, role UIRole) (*UIContext, error) {
	// Simple logging instead of tracing for now
	m.logger.InfoContext(ctx, "Starting user authentication", logger.Fields{
		"username": username,
		"role":     string(role),
	})

	// Authenticate using existing IAM service
	authReq := &authn.AuthenticationRequest{
		Email:    username, // Assuming username is email
		Password: password,
		MFACode:  "", // No MFA for now
	}

	authResult, err := m.iamService.Authentication().Authenticate(ctx, authReq)
	if err != nil {
		m.logger.ErrorContext(ctx, "User authentication failed", logger.Fields{
			"username": username,
			"role":     string(role),
			"error":    err.Error(),
		})
		return nil, errors.NewBusinessError("AUTHENTICATION_FAILED", "Invalid credentials")
	}

	// Validate role authorization using ABAC
	if err := m.validateRoleAuthorization(ctx, authResult.User.ID, role); err != nil {
		m.logger.WarnContext(ctx, "Role authorization failed", logger.Fields{
			"user_id": authResult.User.ID,
			"role":    string(role),
			"error":   err.Error(),
		})
		return nil, errors.NewBusinessError("AUTHORIZATION_FAILED", "Insufficient privileges for UI access")
	}

	// Create session
	sessionID, err := m.generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	csrfToken := ""
	if m.csrfEnabled {
		csrfToken, err = m.generateCSRFToken()
		if err != nil {
			return nil, fmt.Errorf("failed to generate CSRF token: %w", err)
		}
	}

	// Handle EntityID which might be nil
	var entityID *uuid.UUID
	if authResult.User.EntityID != (uuid.UUID{}) {
		entityID = &authResult.User.EntityID
	}

	uiCtx := &UIContext{
		UserID:       authResult.User.ID,
		TenantID:     authResult.User.TenantID,
		EntityID:     entityID,
		Role:         role,
		SessionID:    sessionID,
		CSRFToken:    csrfToken,
		LoginTime:    time.Now(),
		LastActivity: time.Now(),
	}

	// Store session in cache with tenant context
	cacheCtx := cache.SetTenantInContext(ctx, authResult.User.TenantID, "")
	cacheCtx = cache.SetNamespaceInContext(cacheCtx, "ui_sessions")

	sessionKey := fmt.Sprintf("session:%s", sessionID)
	if err := m.cacheService.Set(cacheCtx, sessionKey, uiCtx, m.sessionTTL); err != nil {
		m.logger.ErrorContext(ctx, "Failed to store session", logger.Fields{
			"session_id": sessionID,
			"user_id":    authResult.User.ID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	m.logger.InfoContext(ctx, "User authenticated successfully", logger.Fields{
		"user_id":    authResult.User.ID,
		"tenant_id":  authResult.User.TenantID,
		"role":       string(role),
		"session_id": sessionID,
	})

	return uiCtx, nil
}

// RequireAuthentication middleware that requires valid UI authentication
func (m *UIAuthMiddleware) RequireAuthentication(requiredRoles ...UIRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Extract session from cookie
			sessionCookie, err := r.Cookie(m.cookieName)
			if err != nil {
				m.redirectToLogin(w, r, "no_session_cookie")
				return
			}

			// Validate session
			uiCtx, err := m.validateSession(ctx, sessionCookie.Value)
			if err != nil {
				m.redirectToLogin(w, r, "invalid_session")
				return
			}

			// Check role authorization
			if len(requiredRoles) > 0 && !m.hasRequiredRole(uiCtx.Role, requiredRoles) {
				m.handleUnauthorized(w, r, "insufficient_role")
				return
			}

			// Validate CSRF token for non-GET requests
			if m.csrfEnabled && r.Method != "GET" && r.Method != "HEAD" && r.Method != "OPTIONS" {
				if err := m.validateCSRFToken(r, uiCtx.CSRFToken); err != nil {
					m.handleUnauthorized(w, r, "csrf_validation_failed")
					return
				}
			}

			// Update last activity
			if err := m.updateSessionActivity(ctx, uiCtx); err != nil {
				m.logger.WarnContext(ctx, "Failed to update session activity", logger.Fields{
					"session_id": uiCtx.SessionID,
					"error":      err.Error(),
				})
			}

			// Add UI context to request context
			ctx = context.WithValue(ctx, "ui_context", uiCtx)
			ctx = context.WithValue(ctx, "user_id", uiCtx.UserID)
			ctx = context.WithValue(ctx, "tenant_id", uiCtx.TenantID)
			if uiCtx.EntityID != nil {
				ctx = context.WithValue(ctx, "entity_id", *uiCtx.EntityID)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// LogoutUser logs out a user and invalidates their session
func (m *UIAuthMiddleware) LogoutUser(ctx context.Context, sessionID string) error {
	// Get session from cache to extract tenant info
	uiCtx, err := m.validateSession(ctx, sessionID)
	if err != nil {
		// Session already invalid, consider logout successful
		return nil
	}

	// Delete session from cache
	cacheCtx := cache.SetTenantInContext(ctx, uiCtx.TenantID, "")
	cacheCtx = cache.SetNamespaceInContext(cacheCtx, "ui_sessions")

	sessionKey := fmt.Sprintf("session:%s", sessionID)
	if err := m.cacheService.Delete(cacheCtx, sessionKey); err != nil {
		m.logger.ErrorContext(ctx, "Failed to delete session", logger.Fields{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to delete session: %w", err)
	}

	m.logger.InfoContext(ctx, "User logged out successfully", logger.Fields{
		"user_id":    uiCtx.UserID,
		"session_id": sessionID,
	})

	return nil
}

// CreateSessionCookie creates a secure session cookie
func (m *UIAuthMiddleware) CreateSessionCookie(sessionID string) *http.Cookie {
	return &http.Cookie{
		Name:     m.cookieName,
		Value:    sessionID,
		Path:     "/",
		Domain:   m.cookieDomain,
		MaxAge:   int(m.sessionTTL.Seconds()),
		Secure:   m.cookieSecure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
}

// DeleteSessionCookie creates a cookie that deletes the session
func (m *UIAuthMiddleware) DeleteSessionCookie() *http.Cookie {
	return &http.Cookie{
		Name:     m.cookieName,
		Value:    "",
		Path:     "/",
		Domain:   m.cookieDomain,
		MaxAge:   -1,
		Secure:   m.cookieSecure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
}

// Private methods

// validateRoleAuthorization validates if user has permission for the requested UI role
func (m *UIAuthMiddleware) validateRoleAuthorization(ctx context.Context, userID uuid.UUID, role UIRole) error {
	// Define required permissions for each UI role
	var resourceType, action string

	switch role {
	case UIRoleConsoleAdmin:
		resourceType = "system"
		action = "admin_console_access"
	case UIRoleTenantUser:
		resourceType = "tenant"
		action = "workspace_access"
	case UIRoleClientUser:
		resourceType = "client"
		action = "portal_access"
	default:
		return fmt.Errorf("unknown UI role: %s", role)
	}

	// Use ABAC service for permission evaluation
	evalReq := &abac.PermissionEvaluationRequest{
		UserID:       userID,
		ResourceType: resourceType,
		Action:       action,
		Context: map[string]any{
			"ui_access":    true,
			"ui_role":      string(role),
			"request_time": time.Now(),
		},
	}

	result, err := m.abacService.EvaluatePermission(ctx, evalReq)
	if err != nil {
		return fmt.Errorf("permission evaluation failed: %w", err)
	}

	if string(result.Decision) != "allow" {
		return fmt.Errorf("access denied for UI role %s", role)
	}

	return nil
}

// validateSession validates a session ID and returns the UI context
func (m *UIAuthMiddleware) validateSession(ctx context.Context, sessionID string) (*UIContext, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("empty session ID")
	}

	// Try to get session from cache (we'll iterate through tenant contexts if needed)
	sessionKey := fmt.Sprintf("session:%s", sessionID)

	// First try global context
	var uiCtx UIContext
	err := m.cacheService.Get(ctx, sessionKey, &uiCtx)
	if err == nil {
		// Validate session hasn't expired
		if time.Since(uiCtx.LastActivity) > m.sessionTTL {
			return nil, fmt.Errorf("session expired")
		}
		return &uiCtx, nil
	}

	// Check if it's a cache miss or other error
	if err == cache.ErrCacheMiss {
		return nil, fmt.Errorf("session not found")
	}

	// Other cache error
	return nil, fmt.Errorf("cache error: %w", err)
}

// updateSessionActivity updates the last activity time for a session
func (m *UIAuthMiddleware) updateSessionActivity(ctx context.Context, uiCtx *UIContext) error {
	uiCtx.LastActivity = time.Now()

	// Update session in cache with tenant context
	cacheCtx := cache.SetTenantInContext(ctx, uiCtx.TenantID, "")
	cacheCtx = cache.SetNamespaceInContext(cacheCtx, "ui_sessions")

	sessionKey := fmt.Sprintf("session:%s", uiCtx.SessionID)
	return m.cacheService.Set(cacheCtx, sessionKey, uiCtx, m.sessionTTL)
}

// hasRequiredRole checks if the user's role matches any of the required roles
func (m *UIAuthMiddleware) hasRequiredRole(userRole UIRole, requiredRoles []UIRole) bool {
	for _, required := range requiredRoles {
		if userRole == required {
			return true
		}
	}
	return false
}

// validateCSRFToken validates the CSRF token from the request
func (m *UIAuthMiddleware) validateCSRFToken(r *http.Request, expectedToken string) error {
	// Check X-CSRF-Token header first
	token := r.Header.Get("X-CSRF-Token")
	if token == "" {
		// Check form value as fallback
		token = r.FormValue("csrf_token")
	}

	if token == "" {
		return fmt.Errorf("CSRF token missing")
	}

	if token != expectedToken {
		return fmt.Errorf("CSRF token mismatch")
	}

	return nil
}

// generateSessionID generates a cryptographically secure session ID
func (m *UIAuthMiddleware) generateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// generateCSRFToken generates a cryptographically secure CSRF token
func (m *UIAuthMiddleware) generateCSRFToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// redirectToLogin redirects user to login page
func (m *UIAuthMiddleware) redirectToLogin(w http.ResponseWriter, r *http.Request, reason string) {
	m.logger.InfoContext(r.Context(), "Redirecting to login", logger.Fields{
		"reason": reason,
		"path":   r.URL.Path,
	})

	// If it's an HTMX request, return appropriate response
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/login")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Regular redirect
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// handleUnauthorized handles unauthorized access attempts
func (m *UIAuthMiddleware) handleUnauthorized(w http.ResponseWriter, r *http.Request, reason string) {
	m.logger.WarnContext(r.Context(), "Unauthorized access attempt", logger.Fields{
		"reason": reason,
		"path":   r.URL.Path,
	})

	// If it's an HTMX request, return appropriate response
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/unauthorized")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Regular response
	http.Error(w, "Forbidden", http.StatusForbidden)
}

// ValidateSession validates a session ID and returns the UI context (public method)
func (m *UIAuthMiddleware) ValidateSession(ctx context.Context, sessionID string) (*UIContext, error) {
	return m.validateSession(ctx, sessionID)
}

// GetUIContext extracts UI context from request context
func GetUIContext(ctx context.Context) (*UIContext, bool) {
	uiCtx, ok := ctx.Value("ui_context").(*UIContext)
	return uiCtx, ok
}

// RequireRole checks if the UI context has the required role
func RequireRole(ctx context.Context, requiredRole UIRole) error {
	uiCtx, ok := GetUIContext(ctx)
	if !ok {
		return fmt.Errorf("UI context not found")
	}

	if uiCtx.Role != requiredRole {
		return fmt.Errorf("insufficient role: required %s, have %s", requiredRole, uiCtx.Role)
	}

	return nil
}
