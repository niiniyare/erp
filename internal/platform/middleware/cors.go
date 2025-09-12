package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/niiniyare/erp/internal/shared/logger"
)

// CORSConfig defines the configuration for CORS middleware
type CORSConfig struct {
	AllowedOrigins      []string `json:"allowed_origins"`
	AllowedMethods      []string `json:"allowed_methods"`
	AllowedHeaders      []string `json:"allowed_headers"`
	ExposedHeaders      []string `json:"exposed_headers"`
	AllowCredentials    bool     `json:"allow_credentials"`
	MaxAge              int      `json:"max_age"`
	EnableInDevelopment bool     `json:"enable_in_development"`
}

// DefaultCORSConfig returns a secure default CORS configuration for ERP multi-tenant architecture
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins: []string{
			// FIXME: put real domain names
			// Production domains
			"https://app.erp.company.com",
			"https://api.erp.company.com",
			// Multi-tenant subdomains (configured at runtime)
		},
		AllowedMethods: []string{
			"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS",
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
			"X-Requested-With",
			"X-Tenant-ID",  // Multi-tenant support
			"X-Request-ID", // Request tracing
			"X-API-Key",    // API authentication
			"Cache-Control",
		},
		ExposedHeaders: []string{
			"X-Total-Count",          // Pagination
			"X-Rate-Limit-Remaining", // Rate limiting
			"X-Request-ID",           // Request tracing
		},
		AllowCredentials:    true,
		MaxAge:              86400, // 24 hours
		EnableInDevelopment: true,
	}
}

// CORSMiddleware creates a CORS middleware for multi-tenant ERP architecture
func CORSMiddleware(config *CORSConfig, logger logger.Logger) func(http.Handler) http.Handler {
	if config == nil {
		config = DefaultCORSConfig()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Handle multi-tenant subdomain origins
			if isAllowedOrigin(origin, config, r) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			// Set other CORS headers
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))

			if len(config.ExposedHeaders) > 0 {
				w.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ", "))
			}

			if config.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if config.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", string(rune(config.MaxAge)))
			}

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isAllowedOrigin checks if the origin is allowed, including multi-tenant subdomain support
func isAllowedOrigin(origin string, config *CORSConfig, r *http.Request) bool {
	if origin == "" {
		return false
	}

	// Development mode - be more permissive
	if config.EnableInDevelopment && (os.Getenv("ENVIRONMENT") == "development" || os.Getenv("ENVIRONMENT") == "dev") {
		if strings.HasPrefix(origin, "http://localhost") ||
			strings.HasPrefix(origin, "http://127.0.0.1") ||
			strings.HasPrefix(origin, "https://localhost") {
			return true
		}
	}

	// Check exact matches first
	for _, allowedOrigin := range config.AllowedOrigins {
		if origin == allowedOrigin {
			return true
		}
	}

	// Multi-tenant subdomain support
	// Allow tenant subdomains: https://tenant1.erp.company.com
	return isAllowedTenantSubdomain(origin)
}

// isAllowedTenantSubdomain validates tenant subdomain origins
func isAllowedTenantSubdomain(origin string) bool {
	baseDomain := os.Getenv("BASE_DOMAIN") // e.g., "erp.company.com"
	if baseDomain == "" {
		return false
	}

	// Extract hostname from origin
	if !strings.HasPrefix(origin, "https://") {
		return false // Only allow HTTPS for tenant subdomains
	}

	hostname := strings.TrimPrefix(origin, "https://")

	// Check if it matches the pattern: *.baseDomain
	if strings.HasSuffix(hostname, "."+baseDomain) {
		// Extract tenant subdomain
		tenantPart := strings.TrimSuffix(hostname, "."+baseDomain)

		// Validate tenant subdomain format (alphanumeric + hyphens, 3-63 chars)
		return isValidTenantSubdomain(tenantPart)
	}

	return false
}

// isValidTenantSubdomain validates tenant subdomain format
func isValidTenantSubdomain(subdomain string) bool {
	if len(subdomain) < 3 || len(subdomain) > 63 {
		return false
	}

	// Allow only alphanumeric characters and hyphens
	for _, char := range subdomain {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-') {
			return false
		}
	}

	// Cannot start or end with hyphen
	if subdomain[0] == '-' || subdomain[len(subdomain)-1] == '-' {
		return false
	}

	return true
}
