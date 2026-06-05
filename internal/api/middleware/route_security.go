package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

// RouteSecurityConfig defines security configuration for different route groups
type RouteSecurityConfig struct {
	// API route security
	API APISecurityConfig `json:"api"`

	// UI route security
	UI UISecurityConfig `json:"ui"`

	// Public route security
	Public PublicSecurityConfig `json:"public"`

	// Global security settings
	Global GlobalSecurityConfig `json:"global"`
}

// APISecurityConfig defines security for API routes (/api/*)
type APISecurityConfig struct {
	// Enable authentication middleware
	RequireAuth bool `json:"require_auth"`

	// Enable tenant isolation middleware
	RequireTenant bool `json:"require_tenant"`

	// Rate limiting for API endpoints
	RateLimit RateLimitConfig `json:"rate_limit"`

	// Enable CORS for API
	EnableCORS bool `json:"enable_cors"`

	// API key authentication
	RequireAPIKey bool `json:"require_api_key"`

	// JWT token validation
	ValidateJWT bool `json:"validate_jwt"`

	// Request size limits
	MaxRequestSize int64 `json:"max_request_size"`
}

// UISecurityConfig defines security for UI routes (/ui/*, /, etc.)
type UISecurityConfig struct {
	// Enable session-based authentication
	RequireSession bool `json:"require_session"`

	// Enable CSRF protection
	EnableCSRF bool `json:"enable_csrf"`

	// Security headers (Helmet)
	EnableSecurityHeaders bool `json:"enable_security_headers"`

	// Content Security Policy
	CSPConfig CSPConfig `json:"csp_config"`

	// Session timeout
	SessionTimeout time.Duration `json:"session_timeout"`

	// Cookie security settings
	SecureCookies bool `json:"secure_cookies"`
}

// PublicSecurityConfig defines security for public routes (/health, /metrics, etc.)
type PublicSecurityConfig struct {
	// Rate limiting for public endpoints
	RateLimit BasicRateLimitConfig `json:"rate_limit"`

	// Basic security headers
	EnableBasicHeaders bool `json:"enable_basic_headers"`

	// Allowed origins for public endpoints
	AllowedOrigins []string `json:"allowed_origins"`
}

// GlobalSecurityConfig defines global security settings
type GlobalSecurityConfig struct {
	// Request ID generation
	EnableRequestID bool `json:"enable_request_id"`

	// Panic recovery
	EnableRecovery bool `json:"enable_recovery"`

	// Global request timeout
	RequestTimeout time.Duration `json:"request_timeout"`

	// Enable observability
	EnableObservability bool `json:"enable_observability"`
}

// CSPConfig defines Content Security Policy settings
type CSPConfig struct {
	DefaultSrc []string `json:"default_src"`
	ScriptSrc  []string `json:"script_src"`
	StyleSrc   []string `json:"style_src"`
	ImageSrc   []string `json:"image_src"`
	ConnectSrc []string `json:"connect_src"`
	FontSrc    []string `json:"font_src"`
	ObjectSrc  []string `json:"object_src"`
	MediaSrc   []string `json:"media_src"`
	FrameSrc   []string `json:"frame_src"`
}

// BasicRateLimitConfig for simple rate limiting
type BasicRateLimitConfig struct {
	Max        int           `json:"max"`
	Expiration time.Duration `json:"expiration"`
	Enabled    bool          `json:"enabled"`
}

// RouteSecurityManager manages security middleware for different route groups
type RouteSecurityManager struct {
	config  RouteSecurityConfig
	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.Service
}

// NewRouteSecurityManager creates a new route security manager
func NewRouteSecurityManager(
	config RouteSecurityConfig,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.Service,
) *RouteSecurityManager {
	return &RouteSecurityManager{
		config:  config,
		logger:  logger,
		metrics: metrics,
		tracer:  tracer,
	}
}

// ============================================================================
// ROUTE GROUP SECURITY BUILDERS
// ============================================================================

// ConfigureAPIRoutes applies security middleware for API routes
func (m *RouteSecurityManager) ConfigureAPIRoutes(router fiber.Router) {
	// Global middleware first
	if m.config.Global.EnableRequestID {
		router.Use(requestid.New())
	}

	if m.config.Global.EnableRecovery {
		router.Use(recover.New(recover.Config{
			EnableStackTrace: true,
		}))
	}

	// Observability middleware
	if m.config.Global.EnableObservability && m.logger != nil && m.metrics != nil && m.tracer != nil {
		observabilityConfig := DefaultObservabilityConfig()
		observabilityConfig.ServiceName = "erp-api"
		observabilityConfig.DetailedLogging = true
		router.Use(CreateObservabilityMiddleware(m.logger, m.metrics, m.tracer, &observabilityConfig))
	}

	// Security headers for API
	router.Use(helmet.New(helmet.Config{
		XSSProtection:             "1; mode=block",
		ContentTypeNosniff:        "nosniff",
		XFrameOptions:             "DENY",
		ReferrerPolicy:            "no-referrer",
		CrossOriginEmbedderPolicy: "require-corp",
	}))

	// CORS for API
	if m.config.API.EnableCORS {
		corsConfig := DevelopmentCORSConfig([]int{3000, 8080, 5173}) // Include common dev ports
		router.Use(NewCORSMiddleware(corsConfig))
	}

	// Request size limiting
	if m.config.API.MaxRequestSize > 0 {
		router.Use(func(c *fiber.Ctx) error {
			if int64(len(c.Body())) > m.config.API.MaxRequestSize {
				return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
					"error":    "Request entity too large",
					"max_size": m.config.API.MaxRequestSize,
				})
			}
			return c.Next()
		})
	}

	// Rate limiting for API
	if m.config.API.RateLimit.Enabled {
		// Use the comprehensive rate limiting middleware
		rateLimitConfig := DefaultRateLimitConfig()
		rateLimitConfig.GlobalRPS = m.config.API.RateLimit.GlobalRPS
		rateLimitConfig.IPRPS = m.config.API.RateLimit.IPRPS
		rateLimitConfig.UserRPS = m.config.API.RateLimit.UserRPS
		router.Use(CreateRateLimitMiddleware(&rateLimitConfig, nil, m.logger)) // cache would be injected
	}

}

// ConfigureUIRoutes applies security middleware for UI routes
func (m *RouteSecurityManager) ConfigureUIRoutes(router fiber.Router) {
	// Global middleware
	if m.config.Global.EnableRequestID {
		router.Use(requestid.New())
	}

	if m.config.Global.EnableRecovery {
		router.Use(recover.New())
	}

	// Comprehensive security headers for UI
	if m.config.UI.EnableSecurityHeaders {
		cspDirectives := m.buildCSPDirectives()
		router.Use(helmet.New(helmet.Config{
			XSSProtection:             "1; mode=block",
			ContentTypeNosniff:        "nosniff",
			XFrameOptions:             "SAMEORIGIN", // Allow framing from same origin for UI
			ReferrerPolicy:            "strict-origin-when-cross-origin",
			CrossOriginEmbedderPolicy: "unsafe-none", // More lenient for UI
			ContentSecurityPolicy:     cspDirectives,
			HSTSMaxAge:                31536000, // 1 year
		}))
	}

	// CSRF protection for UI.
	// CookieHTTPOnly MUST be false: JS reads the cookie value and sends it as
	// X-Csrf-Token header (double-submit cookie pattern). HttpOnly blocks that.
	// Cookie name avoids __Secure- prefix when SecureCookies is false — browsers
	// silently reject __Secure- cookies without Secure+HTTPS.
	if m.config.UI.EnableCSRF {
		cookieName := "csrf_token"
		if m.config.UI.SecureCookies {
			cookieName = "__Secure-csrf_"
		}
		router.Use(csrf.New(csrf.Config{
			KeyLookup:      "header:X-Csrf-Token",
			CookieName:     cookieName,
			CookieSameSite: "Strict",
			CookieSecure:   m.config.UI.SecureCookies,
			CookieHTTPOnly: false, // JS must read this to send X-Csrf-Token header
			Expiration:     1 * time.Hour,
			KeyGenerator:   csrf.ConfigDefault.KeyGenerator,
		}))
	}

	// Request timeout for UI
	if m.config.Global.RequestTimeout > 0 {
		router.Use(func(c *fiber.Ctx) error {
			// Implement request timeout
			return c.Next()
		})
	}
}

// ConfigurePublicRoutes applies minimal security for public routes
func (m *RouteSecurityManager) ConfigurePublicRoutes(router fiber.Router) {
	// Minimal security for public endpoints
	if m.config.Public.EnableBasicHeaders {
		router.Use(helmet.New(helmet.Config{
			XSSProtection:      "1; mode=block",
			ContentTypeNosniff: "nosniff",
			XFrameOptions:      "DENY",
		}))
	}

	// Basic rate limiting for public endpoints
	if m.config.Public.RateLimit.Enabled {
		router.Use(limiter.New(limiter.Config{
			Max:        m.config.Public.RateLimit.Max,
			Expiration: m.config.Public.RateLimit.Expiration,
			LimitReached: func(c *fiber.Ctx) error {
				return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
					"error": "Rate limit exceeded for public endpoint",
				})
			},
		}))
	}

	// CORS for public endpoints
	if len(m.config.Public.AllowedOrigins) > 0 {
		corsConfig := CORSConfig{
			AllowedOrigins:   m.config.Public.AllowedOrigins,
			AllowedMethods:   []string{"GET", "HEAD", "OPTIONS"},
			AllowedHeaders:   []string{"Content-Type"},
			ExposedHeaders:   []string{},
			AllowCredentials: false,
			MaxAge:           86400, // 24 hours
		}
		router.Use(NewCORSMiddleware(&corsConfig))
	}
}

// ============================================================================
// HELPER METHODS
// ============================================================================

// buildCSPDirectives creates Content Security Policy directives
func (m *RouteSecurityManager) buildCSPDirectives() string {
	csp := m.config.UI.CSPConfig

	directives := []string{}

	if len(csp.DefaultSrc) > 0 {
		directives = append(directives, "default-src "+joinCSPSources(csp.DefaultSrc))
	}
	if len(csp.ScriptSrc) > 0 {
		directives = append(directives, "script-src "+joinCSPSources(csp.ScriptSrc))
	}
	if len(csp.StyleSrc) > 0 {
		directives = append(directives, "style-src "+joinCSPSources(csp.StyleSrc))
	}
	if len(csp.ImageSrc) > 0 {
		directives = append(directives, "img-src "+joinCSPSources(csp.ImageSrc))
	}
	if len(csp.ConnectSrc) > 0 {
		directives = append(directives, "connect-src "+joinCSPSources(csp.ConnectSrc))
	}
	if len(csp.FontSrc) > 0 {
		directives = append(directives, "font-src "+joinCSPSources(csp.FontSrc))
	}
	if len(csp.ObjectSrc) > 0 {
		directives = append(directives, "object-src "+joinCSPSources(csp.ObjectSrc))
	}
	if len(csp.MediaSrc) > 0 {
		directives = append(directives, "media-src "+joinCSPSources(csp.MediaSrc))
	}
	if len(csp.FrameSrc) > 0 {
		directives = append(directives, "frame-src "+joinCSPSources(csp.FrameSrc))
	}

	return joinStringSlice(directives, "; ")
}

// joinCSPSources joins CSP sources with spaces
func joinCSPSources(sources []string) string {
	return joinStringSlice(sources, " ")
}

// joinStringSlice joins string slice with separator
func joinStringSlice(slice []string, sep string) string {
	if len(slice) == 0 {
		return ""
	}

	result := slice[0]
	for i := 1; i < len(slice); i++ {
		result += sep + slice[i]
	}
	return result
}

// ============================================================================
// DEFAULT CONFIGURATIONS
// ============================================================================

// DefaultAPISecurityConfig returns secure defaults for API routes
func DefaultAPISecurityConfig() APISecurityConfig {
	return APISecurityConfig{
		RequireAuth:    true,
		RequireTenant:  true,
		EnableCORS:     true,
		RequireAPIKey:  false, // Enable based on needs
		ValidateJWT:    true,
		MaxRequestSize: 10 * 1024 * 1024, // 10MB
		RateLimit:      DefaultRateLimitConfig(),
	}
}

// DefaultUISecurityConfig returns secure defaults for UI routes
func DefaultUISecurityConfig() UISecurityConfig {
	return UISecurityConfig{
		RequireSession:        true,
		EnableCSRF:            true,
		EnableSecurityHeaders: true,
		SessionTimeout:        8 * time.Hour,
		SecureCookies:         true, // Set to false for development
		CSPConfig: CSPConfig{
			DefaultSrc: []string{"'self'"},
			ScriptSrc:  []string{"'self'", "'unsafe-inline'", "'unsafe-eval'"}, // Adjust for your needs
			StyleSrc:   []string{"'self'", "'unsafe-inline'"},
			ImageSrc:   []string{"'self'", "data:", "https:"},
			ConnectSrc: []string{"'self'"},
			FontSrc:    []string{"'self'", "https:"},
			ObjectSrc:  []string{"'none'"},
			MediaSrc:   []string{"'self'"},
			FrameSrc:   []string{"'none'"},
		},
	}
}

// DefaultPublicSecurityConfig returns secure defaults for public routes
func DefaultPublicSecurityConfig() PublicSecurityConfig {
	return PublicSecurityConfig{
		EnableBasicHeaders: true,
		AllowedOrigins:     []string{"*"}, // Adjust for your needs
		RateLimit: BasicRateLimitConfig{
			Enabled:    true,
			Max:        100,
			Expiration: 1 * time.Minute,
		},
	}
}

// DefaultGlobalSecurityConfig returns secure defaults for global settings
func DefaultGlobalSecurityConfig() GlobalSecurityConfig {
	return GlobalSecurityConfig{
		EnableRequestID:     true,
		EnableRecovery:      true,
		EnableObservability: true,
		RequestTimeout:      30 * time.Second,
	}
}

// DefaultRouteSecurityConfig returns a complete secure configuration
func DefaultRouteSecurityConfig() RouteSecurityConfig {
	return RouteSecurityConfig{
		API:    DefaultAPISecurityConfig(),
		UI:     DefaultUISecurityConfig(),
		Public: DefaultPublicSecurityConfig(),
		Global: DefaultGlobalSecurityConfig(),
	}
}

// DevelopmentRouteSecurityConfig returns a development-friendly configuration
func DevelopmentRouteSecurityConfig() RouteSecurityConfig {
	config := DefaultRouteSecurityConfig()

	// Relax some restrictions for development
	config.UI.SecureCookies = false
	config.UI.CSPConfig.ScriptSrc = append(config.UI.CSPConfig.ScriptSrc, "'unsafe-eval'")
	config.API.RateLimit.GlobalRPS = 10000 // Higher limits for dev
	config.Public.AllowedOrigins = []string{"*"}

	return config
}
