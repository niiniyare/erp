package middleware

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"

	"awo/internal/shared/logger"
)

// CompressionConfig defines compression middleware configuration
type CompressionConfig struct {
	Level            int      `json:"level"`             // Compression level (1-9, 6 is default)
	MinLength        int      `json:"min_length"`        // Minimum response size to compress
	ContentTypes     []string `json:"content_types"`     // Content types to compress
	DisableStreaming bool     `json:"disable_streaming"` // Disable streaming compression
}

// DefaultCompressionConfig returns optimized compression settings for ERP API
func DefaultCompressionConfig() *CompressionConfig {
	return &CompressionConfig{
		Level:     6,    // Balanced compression/speed
		MinLength: 1024, // 1KB minimum
		ContentTypes: []string{
			"application/json",
			"application/javascript",
			"text/html",
			"text/css",
			"text/plain",
			"text/xml",
			"application/xml",
			"application/rss+xml",
			"application/atom+xml",
			"image/svg+xml",
		},
		DisableStreaming: false,
	}
}

// compressionWriter wraps http.ResponseWriter with gzip compression
type compressionWriter struct {
	http.ResponseWriter
	writer         io.Writer
	gzipWriter     *gzip.Writer
	wroteHeader    bool
	shouldCompress bool
	config         *CompressionConfig
	logger         logger.Logger
}

var gzipWriterPool = sync.Pool{
	New: func() any {
		gw, _ := gzip.NewWriterLevel(io.Discard, 6)
		return gw
	},
}

// CompressionMiddleware creates a gzip compression middleware for Goa HTTP responses
func CompressionMiddleware(config *CompressionConfig, logger logger.Logger) func(http.Handler) http.Handler {
	if config == nil {
		config = DefaultCompressionConfig()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip compression if client doesn't support gzip
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			// Skip compression for certain paths (like health checks)
			if shouldSkipCompression(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			// Create compression writer
			cw := &compressionWriter{
				ResponseWriter: w,
				config:         config,
				logger:         logger,
			}

			// Set response header
			w.Header().Set("Vary", "Accept-Encoding")

			defer func() {
				if cw.gzipWriter != nil {
					cw.gzipWriter.Close()
					gzipWriterPool.Put(cw.gzipWriter)
				}
			}()

			next.ServeHTTP(cw, r)
		})
	}
}

// Write implements http.ResponseWriter interface with compression logic
func (cw *compressionWriter) Write(data []byte) (int, error) {
	if !cw.wroteHeader {
		cw.WriteHeader(http.StatusOK)
	}

	if !cw.shouldCompress {
		return cw.ResponseWriter.Write(data)
	}

	if cw.gzipWriter == nil {
		cw.initGzipWriter()
	}

	return cw.gzipWriter.Write(data)
}

// WriteHeader implements http.ResponseWriter interface
func (cw *compressionWriter) WriteHeader(code int) {
	if cw.wroteHeader {
		return
	}

	cw.wroteHeader = true

	// Determine if we should compress based on content type and other factors
	contentType := cw.Header().Get("Content-Type")
	contentLength := cw.Header().Get("Content-Length")

	cw.shouldCompress = cw.shouldCompressResponse(contentType, contentLength, code)

	if cw.shouldCompress {
		cw.Header().Set("Content-Encoding", "gzip")
		cw.Header().Del("Content-Length") // Let gzip determine the final length

		cw.logger.Debug("Enabling gzip compression", logger.Fields{
			"content_type": contentType,
			"status_code":  code,
			"compression":  "gzip",
		})
	}

	cw.ResponseWriter.WriteHeader(code)
}

// initGzipWriter initializes the gzip writer from the pool
func (cw *compressionWriter) initGzipWriter() {
	cw.gzipWriter = gzipWriterPool.Get().(*gzip.Writer)
	cw.gzipWriter.Reset(cw.ResponseWriter)

	// Set compression level if different from pool default
	if cw.config.Level != 6 {
		cw.gzipWriter.Close()
		gzipWriterPool.Put(cw.gzipWriter)

		var err error
		cw.gzipWriter, err = gzip.NewWriterLevel(cw.ResponseWriter, cw.config.Level)
		if err != nil {
			cw.logger.Warn("Failed to create gzip writer with custom level", logger.Fields{
				"level": cw.config.Level,
				"error": err.Error(),
			})
			// Fallback to default level
			cw.gzipWriter, _ = gzip.NewWriterLevel(cw.ResponseWriter, 6)
		}
	}

	cw.writer = cw.gzipWriter
}

// shouldCompressResponse determines if the response should be compressed
func (cw *compressionWriter) shouldCompressResponse(contentType, contentLength string, statusCode int) bool {
	// Don't compress error responses or redirects
	if statusCode < 200 || statusCode >= 300 {
		return false
	}

	// Check content type
	if !cw.isCompressibleContentType(contentType) {
		return false
	}

	// Check minimum length (if specified)
	if contentLength != "" {
		// Implementation could parse content length and check against MinLength
		// For simplicity, we'll compress if content-type matches
	}

	return true
}

// isCompressibleContentType checks if the content type should be compressed
func (cw *compressionWriter) isCompressibleContentType(contentType string) bool {
	if contentType == "" {
		return false
	}

	// Remove charset and other parameters
	mainType := strings.Split(contentType, ";")[0]
	mainType = strings.TrimSpace(strings.ToLower(mainType))

	for _, ct := range cw.config.ContentTypes {
		if mainType == strings.ToLower(ct) {
			return true
		}
	}

	return false
}

// shouldSkipCompression checks if compression should be skipped for certain paths
func shouldSkipCompression(path string) bool {
	skipPaths := []string{
		"/health",
		"/ping",
		"/metrics",
		"/debug/",
	}

	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}

	return false
}

// Flush implements http.Flusher interface if the underlying ResponseWriter supports it
func (cw *compressionWriter) Flush() {
	if cw.gzipWriter != nil {
		cw.gzipWriter.Flush()
	}

	if flusher, ok := cw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack implements http.Hijacker interface if the underlying ResponseWriter supports it
func (cw *compressionWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := cw.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("responsewriter does not support hijacking")
}
