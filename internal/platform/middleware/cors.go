package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/google/uuid"
)

// CORSConfig defines CORS configuration for the middleware
type CORSConfig struct {
	AllowedOrigins   []string `json:"allowed_origins"`
	AllowedMethods   []string `json:"allowed_methods"`
	AllowedHeaders   []string `json:"allowed_headers"`
	ExposedHeaders   []string `json:"exposed_headers"`
	AllowCredentials bool     `json:"allow_credentials"`
	MaxAge           int      `json:"max_age"`
}

// DefaultCORSConfig returns a default CORS configuration
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Tenant-ID",
			"X-Request-ID",
			"X-API-Key",
		},
		ExposedHeaders: []string{
			"X-Request-ID",
			"X-Total-Count",
		},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}
}

// CORSMiddleware returns a Fiber CORS middleware with the given configuration
func CORSMiddleware(config *CORSConfig) fiber.Handler {
	if config == nil {
		config = DefaultCORSConfig()
	}

	return cors.New(cors.Config{
		AllowOrigins:     joinStrings(config.AllowedOrigins),
		AllowMethods:     joinStrings(config.AllowedMethods),
		AllowHeaders:     joinStrings(config.AllowedHeaders),
		ExposeHeaders:    joinStrings(config.ExposedHeaders),
		AllowCredentials: config.AllowCredentials,
		MaxAge:           config.MaxAge,
	})
}

// MultiTenantCORSMiddleware provides tenant-aware CORS configuration
func MultiTenantCORSMiddleware(baseConfig *CORSConfig) fiber.Handler {
	if baseConfig == nil {
		baseConfig = DefaultCORSConfig()
	}

	return func(c *fiber.Ctx) error {
		// Get tenant-specific configuration if available
		effectiveConfig := baseConfig

		// Try to extract tenant ID from context
		if tenantID, err := GetTenantID(c); err == nil {
			// NOTE: This implementation provides a foundation for tenant-specific CORS configuration.
			// Production implementation should:
			// 1. Add tenant.Service.GetCORSConfig(ctx, tenantID) method to tenant service
			// 2. Implement configuration caching with reasonable TTL (5-15 minutes)
			// 3. Add fallback to base config if tenant config lookup fails
			// 4. Support tenant-specific allowed origins for custom domains
			// 5. Allow per-tenant restriction of methods and headers

			// For now, create tenant-aware configuration with subdomain support
			tenantConfig := createTenantAwareCORSConfig(baseConfig, tenantID, c.Hostname())
			effectiveConfig = tenantConfig
		}

		corsHandler := cors.New(cors.Config{
			AllowOrigins:     joinStrings(effectiveConfig.AllowedOrigins),
			AllowMethods:     joinStrings(effectiveConfig.AllowedMethods),
			AllowHeaders:     joinStrings(effectiveConfig.AllowedHeaders),
			ExposeHeaders:    joinStrings(effectiveConfig.ExposedHeaders),
			AllowCredentials: effectiveConfig.AllowCredentials,
			MaxAge:           effectiveConfig.MaxAge,
		})

		return corsHandler(c)
	}
}

// AdaptCORSConfig adapts CORSConfig to Fiber's cors.Config format
func AdaptCORSConfig(config *CORSConfig) cors.Config {
	if config == nil {
		config = DefaultCORSConfig()
	}

	return cors.Config{
		AllowOrigins:     joinStrings(config.AllowedOrigins),
		AllowMethods:     joinStrings(config.AllowedMethods),
		AllowHeaders:     joinStrings(config.AllowedHeaders),
		ExposeHeaders:    joinStrings(config.ExposedHeaders),
		AllowCredentials: config.AllowCredentials,
		MaxAge:           config.MaxAge,
	}
}

// createTenantAwareCORSConfig creates a tenant-specific CORS configuration
// This is a basic implementation that can be extended with full tenant service integration
func createTenantAwareCORSConfig(baseConfig *CORSConfig, tenantID uuid.UUID, hostname string) *CORSConfig {
	// Create a copy of base configuration
	tenantConfig := &CORSConfig{
		AllowedOrigins:   make([]string, len(baseConfig.AllowedOrigins)),
		AllowedMethods:   make([]string, len(baseConfig.AllowedMethods)),
		AllowedHeaders:   make([]string, len(baseConfig.AllowedHeaders)),
		ExposedHeaders:   make([]string, len(baseConfig.ExposedHeaders)),
		AllowCredentials: baseConfig.AllowCredentials,
		MaxAge:           baseConfig.MaxAge,
	}

	// Copy arrays
	copy(tenantConfig.AllowedOrigins, baseConfig.AllowedOrigins)
	copy(tenantConfig.AllowedMethods, baseConfig.AllowedMethods)
	copy(tenantConfig.AllowedHeaders, baseConfig.AllowedHeaders)
	copy(tenantConfig.ExposedHeaders, baseConfig.ExposedHeaders)

	// Add tenant-specific allowed origins based on subdomain
	if hostname != "" && hostname != "localhost" {
		// Add the current hostname as an allowed origin
		tenantOrigin := "https://" + hostname
		tenantConfig.AllowedOrigins = append(tenantConfig.AllowedOrigins, tenantOrigin)

		// Also allow HTTP for development environments
		if !strings.Contains(hostname, "prod") && !strings.Contains(hostname, "production") {
			devOrigin := "http://" + hostname
			tenantConfig.AllowedOrigins = append(tenantConfig.AllowedOrigins, devOrigin)
		}
	}

	return tenantConfig
}

// joinStrings joins a slice of strings with commas
func joinStrings(strs []string) string {
	if len(strs) == 0 {
		return ""
	}

	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += "," + strs[i]
	}
	return result
}
