package handlers

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/niiniyare/erp/internal/api/swagger"
	"github.com/niiniyare/erp/internal/shared/logger"
)

// SwaggerHandler handles Swagger UI requests
type SwaggerHandler struct{}

// NewSwaggerHandler creates a new Swagger handler
func NewSwaggerHandler() *SwaggerHandler {
	return &SwaggerHandler{}
}

// ServeSwaggerUI serves the Swagger UI files
func (h *SwaggerHandler) ServeSwaggerUI(w http.ResponseWriter, r *http.Request) {
	// Get the requested file path, default to index.html
	filePath := strings.TrimPrefix(r.URL.Path, "/swagger-ui/")
	if filePath == "" || filePath == "/" {
		filePath = "index.html"
	}

	// Read the file from the embedded filesystem
	content, err := swagger.Files.ReadFile(filePath)
	if err != nil {
		logger.Error("Failed to serve Swagger UI file", logger.Fields{
			"path":  filePath,
			"error": err.Error(),
		})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "File not found",
			"file":  filePath,
		})
		return
	}

	// Set appropriate content type
	contentType := getContentType(filePath)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year

	logger.Info("Swagger UI file served", logger.Fields{
		"path":           filePath,
		"content_length": len(content),
		"content_type":   contentType,
	})

	w.WriteHeader(http.StatusOK)
	w.Write(content)
}

func getContentType(filename string) string {
	ext := filepath.Ext(filename)
	switch ext {
	case ".html":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".png":
		return "image/png"
	case ".ico":
		return "image/x-icon"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	case ".svg":
		return "image/svg+xml"
	default:
		return "application/octet-stream"
	}
}
