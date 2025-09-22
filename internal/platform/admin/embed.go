package admin

import (
	"net/http"
	"strings"

	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/shared/logger"
)

// StaticHandler serves static assets (if needed in the future)
type StaticHandler struct{}

// NewStaticHandler creates a new static handler
func NewStaticHandler() *StaticHandler {
	return &StaticHandler{}
}

// ServeHTTP handles static file requests (placeholder for future use)
func (h *StaticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// For now, just return 404 since we're using CDN resources
	http.NotFound(w, r)
}

// Mount adds the admin UI routes to the muxer
func Mount(mux http.Handler, tenantService tenant.Service) http.Handler {
	// Create static file handler for assets
	staticHandler := NewStaticHandler()

	// Create admin page handlers
	adminHandlers := NewAdminHandlers(tenantService)

	logger.Info("🎨 Admin UI mounted successfully", logger.Fields{
		"path":      "/admin/*",
		"type":      "templ + flowbite",
		"framework": "go-templ",
	})

	// Create a new mux that handles both API and admin routes
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if this is an admin UI request
		if strings.HasPrefix(r.URL.Path, "/admin") {
			// Add CORS headers for development
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Tenant-ID, Authorization")
			
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Route admin pages
			switch r.URL.Path {
			case "/admin", "/admin/", "/admin/dashboard":
				adminHandlers.DashboardHandler(w, r)
				return
			case "/admin/tenants":
				adminHandlers.TenantsHandler(w, r)
				return
			case "/admin/users":
				adminHandlers.UsersHandler(w, r)
				return
			default:
				// Serve static files (CSS, JS, images) - currently returns 404
				staticHandler.ServeHTTP(w, r)
				return
			}
		}

		// For all other requests, use the original mux (API routes)
		mux.ServeHTTP(w, r)
	})
}