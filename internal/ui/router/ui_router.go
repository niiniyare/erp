package router

import (
	"context"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/niiniyare/erp/internal/core/abac"
	"github.com/niiniyare/erp/internal/core/audit"
	"github.com/niiniyare/erp/internal/core/iam"
	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/ui/handlers/console"
	"github.com/niiniyare/erp/internal/ui/middleware"
)

// UIRouter manages all UI services on a single port
type UIRouter struct {
	mux    *http.ServeMux
	logger logger.Logger

	// Services
	iamService    iam.Service
	abacService   abac.Service
	tenantService tenant.Service
	auditService  audit.Service
	cacheService  cache.Service

	// Middleware
	authMiddleware *middleware.UIAuthMiddleware

	// Handlers
	consoleAuthHandler      *console.AuthHandler
	consoleDashboardHandler *console.DashboardHandler
	consoleTenantHandler    *console.TenantHandler

	// Configuration
	config UIRouterConfig
}

// UIRouterConfig configures the UI router
type UIRouterConfig struct {
	// Base paths for each UI service
	ConsolePath   string
	WorkspacePath string
	PortalPath    string

	// Asset serving
	StaticPath    string
	AssetsEnabled bool

	// Security
	CSRFEnabled  bool
	SessionTTL   string
	CookieSecure bool
	CookieDomain string
}

// DefaultUIRouterConfig returns default configuration
func DefaultUIRouterConfig() UIRouterConfig {
	return UIRouterConfig{
		ConsolePath:   "/console",
		WorkspacePath: "/workspace",
		PortalPath:    "/portal",
		StaticPath:    "/static",
		AssetsEnabled: true,
		CSRFEnabled:   true,
		SessionTTL:    "24h",
		CookieSecure:  true,
		CookieDomain:  "",
	}
}

// NewUIRouter creates a new UI router with all services
func NewUIRouter(
	iamService iam.Service,
	abacService abac.Service,
	tenantService tenant.Service,
	auditService audit.Service,
	cacheService cache.Service,
	logger logger.Logger,
	config UIRouterConfig,
) *UIRouter {
	// Create authentication middleware
	authMiddleware := middleware.NewUIAuthMiddleware(
		iamService,
		abacService,
		cacheService,
		logger,
	)

	// Create console handlers
	consoleAuthHandler := console.NewAuthHandler(
		iamService,
		abacService,
		cacheService,
		authMiddleware,
		logger,
	)

	consoleDashboardHandler := console.NewDashboardHandler(
		iamService,
		abacService,
		tenantService,
		auditService,
		logger,
	)

	consoleTenantHandler := console.NewTenantHandler(
		tenantService,
		logger,
	)

	router := &UIRouter{
		mux:                     http.NewServeMux(),
		logger:                  logger,
		iamService:              iamService,
		abacService:             abacService,
		tenantService:           tenantService,
		auditService:            auditService,
		cacheService:            cacheService,
		authMiddleware:          authMiddleware,
		consoleAuthHandler:      consoleAuthHandler,
		consoleDashboardHandler: consoleDashboardHandler,
		consoleTenantHandler:    consoleTenantHandler,
		config:                  config,
	}

	// Setup routes
	router.setupRoutes()

	return router
}

// setupRoutes configures all routes for UI services
func (r *UIRouter) setupRoutes() {
	// Root redirect
	r.mux.HandleFunc("/", r.handleRootRedirect)

	// Setup console routes
	r.setupConsoleRoutes()

	// Setup workspace routes (placeholder)
	r.setupWorkspaceRoutes()

	// Setup portal routes (placeholder)
	r.setupPortalRoutes()

	// Setup static asset serving
	if r.config.AssetsEnabled {
		r.setupStaticRoutes()
	}

	// Setup health check
	r.mux.HandleFunc("/health", r.handleHealthCheck)
}

// setupConsoleRoutes configures admin console routes
func (r *UIRouter) setupConsoleRoutes() {
	// Console authentication middleware
	consoleAuth := r.authMiddleware.RequireAuthentication(middleware.UIRoleConsoleAdmin)

	// Public console routes (no auth required)
	r.mux.HandleFunc("/console/login", r.consoleAuthHandler.ShowLoginPage)
	r.mux.HandleFunc("/console/auth/login", r.consoleAuthHandler.HandleLogin)

	// Protected console routes
	r.mux.Handle("/console/auth/logout", consoleAuth(http.HandlerFunc(r.consoleAuthHandler.HandleLogout)))
	r.mux.Handle("/console/auth/refresh", consoleAuth(http.HandlerFunc(r.consoleAuthHandler.HandleRefreshSession)))
	r.mux.Handle("/console/auth/status", consoleAuth(http.HandlerFunc(r.consoleAuthHandler.CheckAuthStatus)))

	// Console dashboard routes
	r.mux.Handle("/console/dashboard", consoleAuth(http.HandlerFunc(r.consoleDashboardHandler.ShowDashboard)))
	r.mux.Handle("/console/api/stats", consoleAuth(http.HandlerFunc(r.consoleDashboardHandler.GetSystemStats)))
	r.mux.Handle("/console/api/activity", consoleAuth(http.HandlerFunc(r.consoleDashboardHandler.GetRecentActivity)))

	// Console tenant management routes
	r.mux.Handle("/console/tenants", consoleAuth(http.HandlerFunc(r.consoleTenantHandler.ListTenants)))
	r.mux.Handle("/console/tenants/", consoleAuth(http.HandlerFunc(r.consoleTenantHandler.RouteTenantRequest)))

	// Default console route
	r.mux.Handle("/console/", consoleAuth(http.HandlerFunc(r.consoleDashboardHandler.ShowDashboard)))

	r.logger.Info("Console routes configured", logger.Fields{
		"base_path": r.config.ConsolePath,
	})
}

// setupWorkspaceRoutes configures tenant workspace routes (placeholder)
func (r *UIRouter) setupWorkspaceRoutes() {
	// Workspace authentication middleware
	workspaceAuth := r.authMiddleware.RequireAuthentication(middleware.UIRoleTenantUser)

	// Placeholder workspace routes
	r.mux.HandleFunc("/workspace/login", r.handleWorkspaceLogin)
	r.mux.Handle("/workspace/dashboard", workspaceAuth(http.HandlerFunc(r.handleWorkspaceDashboard)))
	r.mux.Handle("/workspace/", workspaceAuth(http.HandlerFunc(r.handleWorkspaceDashboard)))

	r.logger.Info("Workspace routes configured", logger.Fields{
		"base_path": r.config.WorkspacePath,
	})
}

// setupPortalRoutes configures client portal routes (placeholder)
func (r *UIRouter) setupPortalRoutes() {
	// Portal authentication middleware
	portalAuth := r.authMiddleware.RequireAuthentication(middleware.UIRoleClientUser)

	// Placeholder portal routes
	r.mux.HandleFunc("/portal/login", r.handlePortalLogin)
	r.mux.Handle("/portal/dashboard", portalAuth(http.HandlerFunc(r.handlePortalDashboard)))
	r.mux.Handle("/portal/", portalAuth(http.HandlerFunc(r.handlePortalDashboard)))

	r.logger.Info("Portal routes configured", logger.Fields{
		"base_path": r.config.PortalPath,
	})
}

// setupStaticRoutes configures static asset serving
func (r *UIRouter) setupStaticRoutes() {
	// In a production environment, you would serve embedded assets
	// For now, this is a placeholder that would serve from filesystem or embedded FS
	r.mux.HandleFunc("/static/", r.handleStaticAssets)

	r.logger.Info("Static asset routes configured", logger.Fields{
		"base_path": r.config.StaticPath,
	})
}

// ServeHTTP implements http.Handler interface
func (r *UIRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Add request logging middleware
	ctx := req.Context()
	ctx = r.addRequestContext(ctx, req)
	req = req.WithContext(ctx)

	// Log request
	r.logger.InfoContext(ctx, "UI request", logger.Fields{
		"method": req.Method,
		"path":   req.URL.Path,
		"remote": req.RemoteAddr,
	})

	// Add security headers
	r.addSecurityHeaders(w)

	// Route the request
	r.mux.ServeHTTP(w, req)
}

// Route handlers

// handleRootRedirect redirects root requests to appropriate UI based on user role
func (r *UIRouter) handleRootRedirect(w http.ResponseWriter, req *http.Request) {
	// For now, redirect to console login
	// In a full implementation, this would detect user role and redirect appropriately
	http.Redirect(w, req, "/console/login", http.StatusSeeOther)
}

// handleHealthCheck provides health check endpoint
func (r *UIRouter) handleHealthCheck(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"ui-router"}`))
}

// Placeholder handlers for workspace and portal

func (r *UIRouter) handleWorkspaceLogin(w http.ResponseWriter, req *http.Request) {
	// Placeholder for workspace login
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
		<!DOCTYPE html>
		<html>
		<head><title>Workspace Login</title></head>
		<body>
			<h1>Tenant Workspace</h1>
			<p>Login page coming soon...</p>
			<a href="/console/login">Admin Console</a>
		</body>
		</html>
	`))
}

func (r *UIRouter) handleWorkspaceDashboard(w http.ResponseWriter, req *http.Request) {
	// Placeholder for workspace dashboard
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
		<!DOCTYPE html>
		<html>
		<head><title>Workspace Dashboard</title></head>
		<body>
			<h1>Tenant Workspace Dashboard</h1>
			<p>Dashboard coming soon...</p>
		</body>
		</html>
	`))
}

func (r *UIRouter) handlePortalLogin(w http.ResponseWriter, req *http.Request) {
	// Placeholder for portal login
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
		<!DOCTYPE html>
		<html>
		<head><title>Client Portal Login</title></head>
		<body>
			<h1>Client Portal</h1>
			<p>Login page coming soon...</p>
			<a href="/console/login">Admin Console</a>
		</body>
		</html>
	`))
}

func (r *UIRouter) handlePortalDashboard(w http.ResponseWriter, req *http.Request) {
	// Placeholder for portal dashboard
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
		<!DOCTYPE html>
		<html>
		<head><title>Client Portal Dashboard</title></head>
		<body>
			<h1>Client Portal Dashboard</h1>
			<p>Dashboard coming soon...</p>
		</body>
		</html>
	`))
}

func (r *UIRouter) handleStaticAssets(w http.ResponseWriter, req *http.Request) {
	// Placeholder for static asset serving
	// In production, this would serve from embedded filesystem
	path := strings.TrimPrefix(req.URL.Path, "/static/")

	// Set appropriate content type based on file extension
	ext := filepath.Ext(path)
	switch ext {
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	// Placeholder response
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Static asset not found"))
}

// Helper methods

// addRequestContext adds common context values to the request
func (r *UIRouter) addRequestContext(ctx context.Context, req *http.Request) context.Context {
	// Add request ID
	requestID := req.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = "req_" + generateRequestID()
	}
	ctx = context.WithValue(ctx, "request_id", requestID)

	// Add client IP
	clientIP := req.Header.Get("X-Forwarded-For")
	if clientIP == "" {
		clientIP = req.RemoteAddr
	}
	ctx = context.WithValue(ctx, "client_ip", clientIP)

	return ctx
}

// addSecurityHeaders adds security headers to the response
func (r *UIRouter) addSecurityHeaders(w http.ResponseWriter) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

	// Only add HSTS in production with HTTPS
	if r.config.CookieSecure {
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	}
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	// Simple timestamp-based ID for now
	// In production, you'd want a more sophisticated ID
	return "12345" // Placeholder
}

// GetHandler returns the HTTP handler for the UI router
func (r *UIRouter) GetHandler() http.Handler {
	return r
}
