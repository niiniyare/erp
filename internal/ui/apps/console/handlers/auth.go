package console

//FIXME:
// import (
// 	"fmt"
// 	"net/http"
// 	"time"
//
// 	"github.com/niiniyare/erp/internal/core/abac"
// 	"github.com/niiniyare/erp/internal/core/iam"
// 	"github.com/niiniyare/erp/internal/platform/cache"
// 	"github.com/niiniyare/erp/internal/shared/errors"
// 	"github.com/niiniyare/erp/internal/shared/logger"
//
// 	"github.com/niiniyare/erp/internal/ui/middleware"
// 	console "github.com/niiniyare/erp/internal/ui/services/console/handlers"
// 	"github.com/niiniyare/erp/internal/ui/types"
// )
//
// // AuthHandler handles authentication for the admin console
// type AuthHandler struct {
// 	iamService     iam.Service
// 	abacService    abac.Service
// 	cacheService   cache.Service
// 	authMiddleware *middleware.UIAuthMiddleware
// 	logger         logger.Logger
// }
//
// // NewAuthHandler creates a new authentication handler
// func NewAuthHandler(
// 	iamService iam.Service,
// 	abacService abac.Service,
// 	cacheService cache.Service,
// 	authMiddleware *middleware.UIAuthMiddleware,
// 	logger logger.Logger,
// ) *AuthHandler {
// 	return &AuthHandler{
// 		iamService:     iamService,
// 		abacService:    abacService,
// 		cacheService:   cacheService,
// 		authMiddleware: authMiddleware,
// 		logger:         logger,
// 	}
// }
//
// // ShowLoginPage displays the admin console login page
// func (h *AuthHandler) ShowLoginPage(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
//
// 	// Check if user is already authenticated
// 	if sessionCookie, err := r.Cookie("ui_session"); err == nil && sessionCookie.Value != "" {
// 		// Validate existing session
// 		if _, err := h.authMiddleware.ValidateSession(ctx, sessionCookie.Value); err == nil {
// 			// User is already authenticated, redirect to dashboard
// 			http.Redirect(w, r, "/console/dashboard", http.StatusSeeOther)
// 			return
// 		}
// 	}
//
// 	// Generate CSRF token for the form
// 	csrfToken, err := h.generateCSRFToken()
// 	if err != nil {
// 		h.logger.ErrorContext(ctx, "Failed to generate CSRF token", logger.Fields{
// 			"error": err.Error(),
// 		})
// 		http.Error(w, "Internal server error", http.StatusInternalServerError)
// 		return
// 	}
//
// 	// Get redirect URL from query parameter
// 	redirectURL := r.URL.Query().Get("redirect")
// 	if redirectURL == "" {
// 		redirectURL = "/console/dashboard"
// 	}
//
// 	// Get error message from query parameter (for failed login attempts)
// 	errorMsg := r.URL.Query().Get("error")
//
// 	data := types.LoginPageData{
// 		Title:       "Admin Console - Login",
// 		CSRFToken:   csrfToken,
// 		Error:       errorMsg,
// 		RedirectURL: redirectURL,
// 	}
// 	// Render login template
// 	if err := console.LoginPage(data).Render(ctx, w); err != nil {
// 		h.logger.ErrorContext(ctx, "Failed to render login page", logger.Fields{
// 			"error": err.Error(),
// 		})
// 		http.Error(w, "Internal server error", http.StatusInternalServerError)
// 		return
// 	}
// }
//
// // HandleLogin processes login form submission
// func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
//
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}
//
// 	// Parse form data
// 	if err := r.ParseForm(); err != nil {
// 		h.redirectWithError(w, r, "Invalid form data", "")
// 		return
// 	}
//
// 	username := r.FormValue("username")
// 	password := r.FormValue("password")
// 	csrfToken := r.FormValue("csrf_token")
// 	redirectURL := r.FormValue("redirect_url")
//
// 	// Validate required fields
// 	if username == "" || password == "" {
// 		h.redirectWithError(w, r, "Username and password are required", redirectURL)
// 		return
// 	}
//
// 	// Validate CSRF token (basic validation - in production you'd want more sophisticated CSRF protection)
// 	if csrfToken == "" {
// 		h.redirectWithError(w, r, "Security token missing", redirectURL)
// 		return
// 	}
//
// 	// Authenticate user with console admin role
// 	uiCtx, err := h.authMiddleware.AuthenticateUser(ctx, username, password, middleware.UIRoleConsoleAdmin)
// 	if err != nil {
// 		h.logger.WarnContext(ctx, "Console authentication failed", logger.Fields{
// 			"username": username,
// 			"error":    err.Error(),
// 		})
//
// 		// Check if it's a business error with specific message
// 		if businessErr, ok := err.(*errors.BusinessError); ok {
// 			h.redirectWithError(w, r, businessErr.Message, redirectURL)
// 			return
// 		}
//
// 		h.redirectWithError(w, r, "Invalid credentials", redirectURL)
// 		return
// 	}
//
// 	// Create session cookie
// 	sessionCookie := h.authMiddleware.CreateSessionCookie(uiCtx.SessionID)
// 	http.SetCookie(w, sessionCookie)
//
// 	// Log successful login
// 	h.logger.InfoContext(ctx, "Console admin logged in successfully", logger.Fields{
// 		"user_id":    uiCtx.UserID,
// 		"tenant_id":  uiCtx.TenantID,
// 		"session_id": uiCtx.SessionID,
// 	})
//
// 	// Redirect to dashboard or specified URL
// 	if redirectURL == "" {
// 		redirectURL = "/console/dashboard"
// 	}
//
// 	// If it's an HTMX request, use HX-Redirect
// 	if r.Header.Get("HX-Request") == "true" {
// 		w.Header().Set("HX-Redirect", redirectURL)
// 		w.WriteHeader(http.StatusOK)
// 		return
// 	}
//
// 	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
// }
//
// // HandleLogout processes logout requests
// func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
//
// 	// Get session cookie
// 	sessionCookie, err := r.Cookie("ui_session")
// 	if err == nil && sessionCookie.Value != "" {
// 		// Logout user (this will invalidate the session)
// 		if err := h.authMiddleware.LogoutUser(ctx, sessionCookie.Value); err != nil {
// 			h.logger.ErrorContext(ctx, "Failed to logout user", logger.Fields{
// 				"session_id": sessionCookie.Value,
// 				"error":      err.Error(),
// 			})
// 		}
// 	}
//
// 	// Delete session cookie
// 	deleteCookie := h.authMiddleware.DeleteSessionCookie()
// 	http.SetCookie(w, deleteCookie)
//
// 	h.logger.InfoContext(ctx, "Console admin logged out")
//
// 	// Redirect to login page
// 	if r.Header.Get("HX-Request") == "true" {
// 		w.Header().Set("HX-Redirect", "/console/login")
// 		w.WriteHeader(http.StatusOK)
// 		return
// 	}
//
// 	http.Redirect(w, r, "/console/login", http.StatusSeeOther)
// }
//
// // HandleRefreshSession refreshes the user's session
// func (h *AuthHandler) HandleRefreshSession(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
//
// 	// Get UI context from middleware
// 	uiCtx, ok := middleware.GetUIContext(ctx)
// 	if !ok {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}
//
// 	// Validate that this is a console admin
// 	if uiCtx.Role != middleware.UIRoleConsoleAdmin {
// 		http.Error(w, "Forbidden", http.StatusForbidden)
// 		return
// 	}
//
// 	// Update session activity (this happens automatically in the middleware)
// 	// Just return success
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// 	w.Write([]byte(`{"status":"success","last_activity":"` + uiCtx.LastActivity.Format(time.RFC3339) + `"}`))
// }
//
// // CheckAuthStatus checks the current authentication status
// func (h *AuthHandler) CheckAuthStatus(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
//
// 	// Try to get UI context
// 	uiCtx, ok := middleware.GetUIContext(ctx)
// 	if !ok {
// 		w.Header().Set("Content-Type", "application/json")
// 		w.WriteHeader(http.StatusUnauthorized)
// 		w.Write([]byte(`{"authenticated":false}`))
// 		return
// 	}
//
// 	// Return authentication status
// 	response := fmt.Sprintf(`{
// 		"authenticated": true,
// 		"user_id": "%s",
// 		"role": "%s",
// 		"session_id": "%s",
// 		"last_activity": "%s",
// 		"csrf_token": "%s"
// 	}`,
// 		uiCtx.UserID.String(),
// 		string(uiCtx.Role),
// 		uiCtx.SessionID,
// 		uiCtx.LastActivity.Format(time.RFC3339),
// 		uiCtx.CSRFToken,
// 	)
//
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// 	w.Write([]byte(response))
// }
//
// // Private helper methods
//
// // redirectWithError redirects to login page with error message
// func (h *AuthHandler) redirectWithError(w http.ResponseWriter, r *http.Request, errorMsg, redirectURL string) {
// 	loginURL := "/console/login?error=" + errorMsg
// 	if redirectURL != "" {
// 		loginURL += "&redirect=" + redirectURL
// 	}
//
// 	if r.Header.Get("HX-Request") == "true" {
// 		w.Header().Set("HX-Redirect", loginURL)
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}
//
// 	http.Redirect(w, r, loginURL, http.StatusSeeOther)
// }
//
// // generateCSRFToken generates a CSRF token (simple implementation)
// func (h *AuthHandler) generateCSRFToken() (string, error) {
// 	// In a production environment, you would use the token service
// 	// For now, we'll use a simple timestamp-based token
// 	return fmt.Sprintf("csrf_%d", time.Now().Unix()), nil
// }
//
// // SetupRoutes sets up authentication routes for the console
// func (h *AuthHandler) SetupRoutes(mux *http.ServeMux) {
// 	// Public routes (no authentication required)
// 	mux.HandleFunc("/console/login", h.ShowLoginPage)
// 	mux.HandleFunc("/console/auth/login", h.HandleLogin)
//
// 	// Protected routes (require authentication)
// 	authMux := http.NewServeMux()
// 	authMux.HandleFunc("/console/auth/logout", h.HandleLogout)
// 	authMux.HandleFunc("/console/auth/refresh", h.HandleRefreshSession)
// 	authMux.HandleFunc("/console/auth/status", h.CheckAuthStatus)
//
// 	// Apply authentication middleware to protected routes
// 	authMiddleware := h.authMiddleware.RequireAuthentication(middleware.UIRoleConsoleAdmin)
// 	mux.Handle("/console/auth/", authMiddleware(authMux))
// }
