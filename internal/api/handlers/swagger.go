package handlers

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
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
func (h *SwaggerHandler) ServeSwaggerUI(c *gin.Context) {
	// Get the requested file path, default to index.html
	filePath := strings.TrimPrefix(c.Request.URL.Path, "/swagger-ui/")
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
		c.JSON(http.StatusNotFound, gin.H{
			"error": "File not found",
			"file":  filePath,
		})
		return
	}

	// Set appropriate content type
	contentType := getContentType(filePath)
	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "public, max-age=31536000") // Cache for 1 year

	logger.Info("Swagger UI file served", logger.Fields{
		"path":           filePath,
		"content_length": len(content),
		"content_type":   contentType,
	})

	c.Data(http.StatusOK, contentType, content)
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
