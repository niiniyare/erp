package swagger

import (
	"embed"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/niiniyare/erp/internal/shared/logger"
)

//go:embed *
var Files embed.FS

// ServeHTTP creates an HTTP handler for serving the embedded Swagger UI files
func ServeHTTP() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Remove the /swagger-ui/ prefix from the path
		path := strings.TrimPrefix(r.URL.Path, "/swagger-ui/")
		if path == "" || path == "/" {
			path = "index.html"
		}

		// Try to read the file from embedded filesystem
		content, err := fs.ReadFile(Files, path)
		if err != nil {
			logger.Error("Failed to serve Swagger UI file", logger.Fields{
				"path":  path,
				"error": err.Error(),
			})
			http.NotFound(w, r)
			return
		}

		// Set appropriate content type
		contentType := getContentType(path)
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		w.Write(content)

		logger.Info("Swagger UI file served", logger.Fields{
			"path":           path,
			"content_length": len(content),
			"content_type":   contentType,
		})
	})
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
	default:
		return "application/octet-stream"
	}
}
