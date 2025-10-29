package middleware

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestRouteSecurityManager_APIRoutes(t *testing.T) {
	// Create security manager with API configuration
	config := DefaultRouteSecurityConfig()
	securityManager := NewRouteSecurityManager(config, nil, nil, nil)

	// Create Fiber app
	app := fiber.New()

	// Create API route group
	apiGroup := app.Group("/api")
	securityManager.ConfigureAPIRoutes(apiGroup)

	// Add test endpoint
	apiGroup.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "API endpoint"})
	})

	// Test API request
	req := httptest.NewRequest("GET", "/api/test", nil)
	resp, err := app.Test(req)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Verify security headers are present
	assert.NotEmpty(t, resp.Header.Get("X-Xss-Protection"))
	assert.NotEmpty(t, resp.Header.Get("X-Content-Type-Options"))
	assert.NotEmpty(t, resp.Header.Get("X-Frame-Options"))
}

func TestRouteSecurityManager_UIRoutes(t *testing.T) {
	// Create security manager with UI configuration
	config := DefaultRouteSecurityConfig()
	securityManager := NewRouteSecurityManager(config, nil, nil, nil)

	// Create Fiber app
	app := fiber.New()

	// Create UI route group
	uiGroup := app.Group("/ui")
	securityManager.ConfigureUIRoutes(uiGroup)

	// Add test endpoint
	uiGroup.Get("/dashboard", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "UI endpoint"})
	})

	// Test UI request
	req := httptest.NewRequest("GET", "/ui/dashboard", nil)
	resp, err := app.Test(req)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Verify UI-specific security headers
	assert.NotEmpty(t, resp.Header.Get("Content-Security-Policy"))
	assert.Equal(t, "SAMEORIGIN", resp.Header.Get("X-Frame-Options"))
}

func TestRouteSecurityManager_PublicRoutes(t *testing.T) {
	// Create security manager with public configuration
	config := DefaultRouteSecurityConfig()
	securityManager := NewRouteSecurityManager(config, nil, nil, nil)

	// Create Fiber app
	app := fiber.New()

	// Create public route group
	publicGroup := app.Group("")
	securityManager.ConfigurePublicRoutes(publicGroup)

	// Add test endpoint
	publicGroup.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "healthy"})
	})

	// Test public request
	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Verify minimal security headers for public routes
	assert.NotEmpty(t, resp.Header.Get("X-Xss-Protection"))
	assert.Equal(t, "DENY", resp.Header.Get("X-Frame-Options"))
}

func TestRouteSecurityManager_CSPConfiguration(t *testing.T) {
	// Setup mocks
	// Setup

	// Create custom CSP configuration
	config := DefaultRouteSecurityConfig()
	config.UI.CSPConfig = CSPConfig{
		DefaultSrc: []string{"'self'"},
		ScriptSrc:  []string{"'self'", "'unsafe-inline'"},
		StyleSrc:   []string{"'self'", "'unsafe-inline'"},
		ImageSrc:   []string{"'self'", "data:", "https:"},
	}

	securityManager := NewRouteSecurityManager(config, nil, nil, nil)

	// Test CSP directive building
	cspDirectives := securityManager.buildCSPDirectives()

	// Assertions
	assert.Contains(t, cspDirectives, "default-src 'self'")
	assert.Contains(t, cspDirectives, "script-src 'self' 'unsafe-inline'")
	assert.Contains(t, cspDirectives, "style-src 'self' 'unsafe-inline'")
	assert.Contains(t, cspDirectives, "img-src 'self' data: https:")
}

func TestRouteSecurityManager_RateLimitConfiguration(t *testing.T) {
	// Setup mocks
	// Setup

	// Create configuration with rate limiting enabled
	config := DefaultRouteSecurityConfig()
	config.Public.RateLimit.Enabled = true
	config.Public.RateLimit.Max = 5
	config.Public.RateLimit.Expiration = 1 * time.Minute

	securityManager := NewRouteSecurityManager(config, nil, nil, nil)

	// Create Fiber app
	app := fiber.New()

	// Configure public routes with rate limiting
	publicGroup := app.Group("")
	securityManager.ConfigurePublicRoutes(publicGroup)

	// Add test endpoint
	publicGroup.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Test rate limiting by making multiple requests
	for i := 0; i < 6; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)

		if i < 5 {
			// First 5 requests should succeed
			assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		} else {
			// 6th request should be rate limited
			assert.Equal(t, fiber.StatusTooManyRequests, resp.StatusCode)
		}
	}
}

func TestRouteSecurityManager_SecurityConfigDefaults(t *testing.T) {
	tests := []struct {
		name           string
		configFunc     func() interface{}
		expectedFields map[string]interface{}
	}{
		{
			name:       "API Security Config",
			configFunc: func() interface{} { return DefaultAPISecurityConfig() },
			expectedFields: map[string]interface{}{
				"RequireAuth":   true,
				"RequireTenant": true,
				"EnableCORS":    true,
				"ValidateJWT":   true,
			},
		},
		{
			name:       "UI Security Config",
			configFunc: func() interface{} { return DefaultUISecurityConfig() },
			expectedFields: map[string]interface{}{
				"RequireSession":        true,
				"EnableCSRF":            true,
				"EnableSecurityHeaders": true,
				"SecureCookies":         true,
			},
		},
		{
			name:       "Public Security Config",
			configFunc: func() interface{} { return DefaultPublicSecurityConfig() },
			expectedFields: map[string]interface{}{
				"EnableBasicHeaders": true,
			},
		},
		{
			name:       "Global Security Config",
			configFunc: func() interface{} { return DefaultGlobalSecurityConfig() },
			expectedFields: map[string]interface{}{
				"EnableRequestID":     true,
				"EnableRecovery":      true,
				"EnableObservability": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.configFunc()

			// Use reflection or type assertions to verify expected fields
			// This is a simplified check - in real tests you'd use reflection
			switch c := config.(type) {
			case APISecurityConfig:
				if val, exists := tt.expectedFields["RequireAuth"]; exists {
					assert.Equal(t, val, c.RequireAuth)
				}
				if val, exists := tt.expectedFields["RequireTenant"]; exists {
					assert.Equal(t, val, c.RequireTenant)
				}
				if val, exists := tt.expectedFields["EnableCORS"]; exists {
					assert.Equal(t, val, c.EnableCORS)
				}
				if val, exists := tt.expectedFields["ValidateJWT"]; exists {
					assert.Equal(t, val, c.ValidateJWT)
				}
			case UISecurityConfig:
				if val, exists := tt.expectedFields["RequireSession"]; exists {
					assert.Equal(t, val, c.RequireSession)
				}
				if val, exists := tt.expectedFields["EnableCSRF"]; exists {
					assert.Equal(t, val, c.EnableCSRF)
				}
				if val, exists := tt.expectedFields["EnableSecurityHeaders"]; exists {
					assert.Equal(t, val, c.EnableSecurityHeaders)
				}
				if val, exists := tt.expectedFields["SecureCookies"]; exists {
					assert.Equal(t, val, c.SecureCookies)
				}
			case PublicSecurityConfig:
				if val, exists := tt.expectedFields["EnableBasicHeaders"]; exists {
					assert.Equal(t, val, c.EnableBasicHeaders)
				}
			case GlobalSecurityConfig:
				if val, exists := tt.expectedFields["EnableRequestID"]; exists {
					assert.Equal(t, val, c.EnableRequestID)
				}
				if val, exists := tt.expectedFields["EnableRecovery"]; exists {
					assert.Equal(t, val, c.EnableRecovery)
				}
				if val, exists := tt.expectedFields["EnableObservability"]; exists {
					assert.Equal(t, val, c.EnableObservability)
				}
			}
		})
	}
}

func TestRouteSecurityManager_DevelopmentConfiguration(t *testing.T) {
	// Get development configuration
	devConfig := DevelopmentRouteSecurityConfig()

	// Verify development-specific settings
	assert.False(t, devConfig.UI.SecureCookies, "Secure cookies should be disabled in development")
	assert.Contains(t, devConfig.UI.CSPConfig.ScriptSrc, "'unsafe-eval'", "Should allow unsafe-eval in development")
	assert.Equal(t, 10000, devConfig.API.RateLimit.GlobalRPS, "Should have higher rate limits in development")
	assert.Contains(t, devConfig.Public.AllowedOrigins, "*", "Should allow all origins in development")
}

func TestRouteSecurityManager_SecurityHeaderValidation(t *testing.T) {
	// Setup mocks
	// Setup

	// Create security manager
	config := DefaultRouteSecurityConfig()
	securityManager := NewRouteSecurityManager(config, nil, nil, nil)

	// Create Fiber app with different route groups
	app := fiber.New()

	// Configure API routes
	apiGroup := app.Group("/api")
	securityManager.ConfigureAPIRoutes(apiGroup)
	apiGroup.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"type": "api"})
	})

	// Configure UI routes
	uiGroup := app.Group("/ui")
	securityManager.ConfigureUIRoutes(uiGroup)
	uiGroup.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"type": "ui"})
	})

	// Configure public routes
	publicGroup := app.Group("")
	securityManager.ConfigurePublicRoutes(publicGroup)
	publicGroup.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"type": "public"})
	})

	tests := []struct {
		name            string
		path            string
		expectedHeaders map[string]string
	}{
		{
			name: "API Route Headers",
			path: "/api/test",
			expectedHeaders: map[string]string{
				"X-Frame-Options":        "DENY",
				"X-Content-Type-Options": "nosniff",
				"X-Xss-Protection":       "1; mode=block",
				"Referrer-Policy":        "no-referrer",
			},
		},
		{
			name: "UI Route Headers",
			path: "/ui/test",
			expectedHeaders: map[string]string{
				"X-Frame-Options":        "SAMEORIGIN",
				"X-Content-Type-Options": "nosniff",
				"X-Xss-Protection":       "1; mode=block",
			},
		},
		{
			name: "Public Route Headers",
			path: "/health",
			expectedHeaders: map[string]string{
				"X-Frame-Options":        "DENY",
				"X-Content-Type-Options": "nosniff",
				"X-Xss-Protection":       "1; mode=block",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusOK, resp.StatusCode)

			// Verify expected headers
			for headerName, expectedValue := range tt.expectedHeaders {
				actualValue := resp.Header.Get(headerName)
				assert.Equal(t, expectedValue, actualValue,
					"Header %s should be %s, got %s", headerName, expectedValue, actualValue)
			}
		})
	}
}

func TestRouteSecurityManager_CSRFProtection(t *testing.T) {
	// Setup mocks
	// Setup

	// Create security manager with CSRF enabled
	config := DefaultRouteSecurityConfig()
	config.UI.EnableCSRF = true
	securityManager := NewRouteSecurityManager(config, nil, nil, nil)

	// Create Fiber app
	app := fiber.New()

	// Configure UI routes with CSRF
	uiGroup := app.Group("/ui")
	securityManager.ConfigureUIRoutes(uiGroup)
	uiGroup.Post("/form", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Form submitted"})
	})

	// Test POST request without CSRF token (should fail)
	reqWithoutCSRF := httptest.NewRequest("POST", "/ui/form", strings.NewReader("data=test"))
	reqWithoutCSRF.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	respWithoutCSRF, err := app.Test(reqWithoutCSRF)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, respWithoutCSRF.StatusCode)
}

// Benchmark test for route security middleware
func BenchmarkRouteSecurityMiddleware(b *testing.B) {
	// Setup mocks
	// Setup

	// Create security manager
	config := DefaultRouteSecurityConfig()
	securityManager := NewRouteSecurityManager(config, nil, nil, nil)

	// Create Fiber app
	app := fiber.New()

	// Configure API routes
	apiGroup := app.Group("/api")
	securityManager.ConfigureAPIRoutes(apiGroup)
	apiGroup.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/test", nil)
		resp, err := app.Test(req)
		if err != nil {
			b.Fatal(err)
		}
		if resp.StatusCode != fiber.StatusOK {
			b.Fatalf("Expected status 200, got %d", resp.StatusCode)
		}
	}
}
