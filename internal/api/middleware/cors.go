package middleware

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/shared/logger"
)

// CORSConfig defines CORS configuration for the middleware.
type CORSConfig struct {
	AllowedOrigins   []string `json:"allowed_origins" validate:"required,min=1"`
	AllowedMethods   []string `json:"allowed_methods" validate:"required,min=1"`
	AllowedHeaders   []string `json:"allowed_headers" validate:"required"`
	ExposedHeaders   []string `json:"exposed_headers"`
	AllowCredentials bool     `json:"allow_credentials"`
	MaxAge           int      `json:"max_age" validate:"min=0,max=86400"`
}

// TenantCORSProvider defines an interface for retrieving tenant-specific CORS configuration.
type TenantCORSProvider interface {
	GetCORSConfig(ctx context.Context, tenantID uuid.UUID) (*CORSConfig, error)
}

// CORSOptions defines optional configuration for CORS middleware.
type CORSOptions struct {
	// Environment specifies the deployment environment (development, staging, production)
	Environment string

	// TenantProvider supplies tenant-specific CORS configurations
	TenantProvider TenantCORSProvider

	// CacheTTL defines how long to cache tenant CORS configs (default: 5 minutes)
	CacheTTL time.Duration

	// Logger provides structured logging capability
	Logger logger.Logger

	// AllowLocalhostPorts specifies which localhost ports to allow in development
	AllowLocalhostPorts []int
}

// tenantCORSCache provides thread-safe caching of tenant CORS configurations.
type tenantCORSCache struct {
	mu     sync.RWMutex
	cache  map[uuid.UUID]*cacheEntry
	ttl    time.Duration
	logger logger.Logger
}

type cacheEntry struct {
	config    *CORSConfig
	expiresAt time.Time
}

// NewCORSMiddleware creates a standard CORS middleware with the provided configuration.
func NewCORSMiddleware(config *CORSConfig) fiber.Handler {
	if config == nil {
		config = DefaultCORSConfig()
	}

	if err := validateCORSConfig(config); err != nil {
		panic(fmt.Sprintf("invalid CORS configuration: %v", err))
	}

	return cors.New(adaptCORSConfig(config))
}

// NewMultiTenantCORSMiddleware creates a tenant-aware CORS middleware.
func NewMultiTenantCORSMiddleware(baseConfig *CORSConfig, opts *CORSOptions) fiber.Handler {
	if baseConfig == nil {
		baseConfig = DefaultCORSConfig()
	}

	if opts == nil {
		opts = &CORSOptions{}
	}

	if err := validateCORSConfig(baseConfig); err != nil {
		panic(fmt.Sprintf("invalid base CORS configuration: %v", err))
	}

	// Set defaults
	if opts.CacheTTL == 0 {
		opts.CacheTTL = 5 * time.Minute
	}
	if opts.Environment == "" {
		opts.Environment = "production"
	}

	cache := &tenantCORSCache{
		cache:  make(map[uuid.UUID]*cacheEntry),
		ttl:    opts.CacheTTL,
		logger: opts.Logger,
	}

	// Use dynamic origin validation for better performance
	return cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			return cache.validateOrigin(origin, baseConfig, opts)
		},
		AllowMethods:     strings.Join(baseConfig.AllowedMethods, ","),
		AllowHeaders:     strings.Join(baseConfig.AllowedHeaders, ","),
		ExposeHeaders:    strings.Join(baseConfig.ExposedHeaders, ","),
		AllowCredentials: baseConfig.AllowCredentials,
		MaxAge:           baseConfig.MaxAge,
	})
}

// DefaultCORSConfig returns a secure default CORS configuration.
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins: []string{
			"https://app.example.com",
			"https://api.example.com",
		},
		AllowedMethods: []string{
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodPut,
			fiber.MethodPatch,
			fiber.MethodDelete,
			fiber.MethodOptions,
		},
		AllowedHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Tenant-ID",
			"X-Request-ID",
			"X-API-Key",
			"X-CSRF-Token",
		},
		ExposedHeaders: []string{
			"X-Request-ID",
			"X-Total-Count",
			"X-Page-Count",
			"Link",
		},
		AllowCredentials: true,
		MaxAge:           3600, // 1 hour (conservative default)
	}
}

// DevelopmentCORSConfig returns a permissive CORS configuration for development.
// WARNING: Never use this in production environments.
func DevelopmentCORSConfig(localhostPorts []int) *CORSConfig {
	origins := []string{
		"http://localhost:3000",
		"http://localhost:8080",
		"http://127.0.0.1:3000",
		"http://127.0.0.1:8080",
		"null", // Allow file:// origin for local development
	}

	// Add custom localhost ports
	for _, port := range localhostPorts {
		origins = append(origins,
			fmt.Sprintf("http://localhost:%d", port),
			fmt.Sprintf("http://127.0.0.1:%d", port),
		)
	}

	config := DefaultCORSConfig()
	config.AllowedOrigins = origins
	return config
}

// validateOrigin checks if an origin is allowed based on configuration and context.
func (c *tenantCORSCache) validateOrigin(origin string, baseConfig *CORSConfig, opts *CORSOptions) bool {
	if origin == "" {
		return false
	}

	// Check base allowed origins
	if isOriginAllowed(origin, baseConfig.AllowedOrigins) {
		return true
	}

	// In development, allow localhost
	if opts.Environment == "development" {
		if isLocalhostOrigin(origin, opts.AllowLocalhostPorts) {
			return true
		}
	}

	// TODO: Implement tenant-specific origin validation
	// This would require extracting tenant context from the request
	// and checking against tenant-specific allowed origins

	return false
}

// isOriginAllowed checks if an origin matches any of the allowed patterns.
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}

		// Support wildcard subdomains (e.g., "https://*.example.com")
		if strings.Contains(allowed, "*") {
			if matchWildcardOrigin(origin, allowed) {
				return true
			}
		}
	}
	return false
}

// matchWildcardOrigin matches an origin against a wildcard pattern.
func matchWildcardOrigin(origin, pattern string) bool {
	// Simple wildcard matching for subdomains
	// e.g., "https://*.example.com" matches "https://tenant1.example.com"
	pattern = strings.ReplaceAll(pattern, "*", "")
	return strings.HasSuffix(origin, pattern)
}

// isLocalhostOrigin checks if an origin is a localhost origin.
func isLocalhostOrigin(origin string, allowedPorts []int) bool {
	if !strings.HasPrefix(origin, "http://localhost:") &&
		!strings.HasPrefix(origin, "http://127.0.0.1:") {
		return false
	}

	// If no specific ports are configured, allow common development ports
	if len(allowedPorts) == 0 {
		allowedPorts = []int{3000, 3001, 8080, 8081, 5173, 4200}
	}

	for _, port := range allowedPorts {
		if strings.HasSuffix(origin, fmt.Sprintf(":%d", port)) {
			return true
		}
	}

	return false
}

// getCachedConfig retrieves a cached tenant CORS configuration if valid.
func (c *tenantCORSCache) getCachedConfig(tenantID uuid.UUID) (*CORSConfig, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.cache[tenantID]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.expiresAt) {
		return nil, false
	}

	return entry.config, true
}

// setCachedConfig stores a tenant CORS configuration in cache.
func (c *tenantCORSCache) setCachedConfig(tenantID uuid.UUID, config *CORSConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[tenantID] = &cacheEntry{
		config:    config,
		expiresAt: time.Now().Add(c.ttl),
	}

	// Simple cache size management
	if len(c.cache) > 1000 {
		c.evictExpired()
	}
}

// evictExpired removes expired entries from the cache (must be called with lock held).
func (c *tenantCORSCache) evictExpired() {
	now := time.Now()
	for tenantID, entry := range c.cache {
		if now.After(entry.expiresAt) {
			delete(c.cache, tenantID)
		}
	}
}

// adaptCORSConfig converts CORSConfig to Fiber's cors.Config.
func adaptCORSConfig(config *CORSConfig) cors.Config {
	return cors.Config{
		AllowOrigins:     strings.Join(config.AllowedOrigins, ","),
		AllowMethods:     strings.Join(config.AllowedMethods, ","),
		AllowHeaders:     strings.Join(config.AllowedHeaders, ","),
		ExposeHeaders:    strings.Join(config.ExposedHeaders, ","),
		AllowCredentials: config.AllowCredentials,
		MaxAge:           config.MaxAge,
	}
}

// validateCORSConfig ensures the CORS configuration is valid and secure.
func validateCORSConfig(config *CORSConfig) error {
	if len(config.AllowedOrigins) == 0 {
		return fmt.Errorf("at least one allowed origin must be specified")
	}

	if len(config.AllowedMethods) == 0 {
		return fmt.Errorf("at least one allowed method must be specified")
	}

	// Security check: wildcard origins with credentials is not allowed
	if config.AllowCredentials {
		for _, origin := range config.AllowedOrigins {
			if origin == "*" {
				return fmt.Errorf("wildcard origin '*' cannot be used with credentials enabled")
			}
		}
	}

	if config.MaxAge < 0 || config.MaxAge > 86400 {
		return fmt.Errorf("max age must be between 0 and 86400 seconds")
	}

	return nil
}

// Clone creates a deep copy of the CORS configuration.
func (c *CORSConfig) Clone() *CORSConfig {
	if c == nil {
		return nil
	}

	return &CORSConfig{
		AllowedOrigins:   append([]string(nil), c.AllowedOrigins...),
		AllowedMethods:   append([]string(nil), c.AllowedMethods...),
		AllowedHeaders:   append([]string(nil), c.AllowedHeaders...),
		ExposedHeaders:   append([]string(nil), c.ExposedHeaders...),
		AllowCredentials: c.AllowCredentials,
		MaxAge:           c.MaxAge,
	}
}

// WithOrigins returns a new config with additional allowed origins.
func (c *CORSConfig) WithOrigins(origins ...string) *CORSConfig {
	config := c.Clone()
	config.AllowedOrigins = append(config.AllowedOrigins, origins...)
	return config
}

// WithCredentials returns a new config with credentials setting.
func (c *CORSConfig) WithCredentials(allow bool) *CORSConfig {
	config := c.Clone()
	config.AllowCredentials = allow
	return config
}
