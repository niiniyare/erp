package admin

import (
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/web"
)

// Handler serves the embedded admin UI
type Handler struct {
	fileSystem http.FileSystem
	indexHTML  []byte
}

// NewHandler creates a new admin UI handler
func NewHandler() (*Handler, error) {
	// Get the admin UI files from the embedded filesystem
	adminFS, err := fs.Sub(web.AdminUI, "admin")
	if err != nil {
		return nil, err
	}

	// Read the index.html file for SPA routing
	indexData, err := fs.ReadFile(adminFS, "index.html")
	if err != nil {
		return nil, err
	}

	return &Handler{
		fileSystem: http.FS(adminFS),
		indexHTML:  indexData,
	}, nil
}

// ServeHTTP handles HTTP requests for the admin UI
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Clean the path
	path := strings.TrimPrefix(r.URL.Path, "/admin")
	if path == "" || path == "/" {
		path = "/index.html"
	}

	if path == "/test.html" {
		file, err := h.fileSystem.Open("/test.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer file.Close()
		stat, _ := file.Stat()
		http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
		return
	}

	// For SPA routing - serve index.html for admin routes
	if strings.HasPrefix(path, "/tenants") || 
	   strings.HasPrefix(path, "/users") || 
	   strings.HasPrefix(path, "/finance") || 
	   strings.HasPrefix(path, "/system") ||
	   strings.HasPrefix(path, "/dashboard") {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(h.indexHTML)
		return
	}

	// Try to serve the static file
	file, err := h.fileSystem.Open(path)
	if err != nil {
		// If file not found, serve index.html for SPA routing
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(h.indexHTML)
		return
	}
	defer file.Close()

	// Get file info for content type
	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	// Set content type based on file extension
	ext := filepath.Ext(path)
	switch ext {
	case ".html":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".json":
		w.Header().Set("Content-Type", "application/json")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	case ".ico":
		w.Header().Set("Content-Type", "image/x-icon")
	}

	// Add cache headers for static assets
	if ext == ".js" || ext == ".css" || ext == ".png" || ext == ".jpg" || ext == ".svg" {
		w.Header().Set("Cache-Control", "public, max-age=31536000") // 1 year
	} else {
		w.Header().Set("Cache-Control", "no-cache")
	}

	// Serve the file
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
}

// Mount adds the admin UI routes to the muxer
func Mount(mux http.Handler) http.Handler {
	adminHandler, err := NewHandler()
	if err != nil {
		logger.Error("Failed to create admin UI handler", logger.Fields{
			"error": err.Error(),
		})
		// Return the original mux if admin UI fails to load
		return mux
	}

	logger.Info("🎨 Admin UI mounted successfully", logger.Fields{
		"path": "/admin/*",
		"type": "embedded",
		"framework": "amis",
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
			
			adminHandler.ServeHTTP(w, r)
			return
		}

		// For all other requests, use the original mux (API routes)
		mux.ServeHTTP(w, r)
	})
}