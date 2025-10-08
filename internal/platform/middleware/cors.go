package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
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
		// TODO: Implement tenant-specific CORS configuration lookup
		// This could be extended to load different CORS settings per tenant
		
		// For now, use the base configuration
		corsHandler := cors.New(cors.Config{
			AllowOrigins:     joinStrings(baseConfig.AllowedOrigins),
			AllowMethods:     joinStrings(baseConfig.AllowedMethods),
			AllowHeaders:     joinStrings(baseConfig.AllowedHeaders),
			ExposeHeaders:    joinStrings(baseConfig.ExposedHeaders),
			AllowCredentials: baseConfig.AllowCredentials,
			MaxAge:           baseConfig.MaxAge,
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