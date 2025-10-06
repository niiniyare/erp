package middleware

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/core/iam"
	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/platform/config"
	"github.com/niiniyare/erp/internal/shared"
	sharedErrors "github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// FiberMiddleware provides all middleware functions for Fiber
type FiberMiddleware struct {
	config        *config.Config
	tenantService tenant.Service
	iamService    iam.Service
	store         db.Store
	logger        logger.Logger
	metrics       *metrics.MetricsService
	tracing       tracing.TracingService
	whitelist     *EndpointWhitelist
}

// NewFiberMiddleware creates a new Fiber middleware instance
func NewFiberMiddleware(
	cfg *config.Config,
	tenantService tenant.Service,
	iamService iam.Service,
	store db.Store,
	logger logger.Logger,
	metrics *metrics.MetricsService,
	tracing tracing.TracingService,
) *FiberMiddleware {
	whitelist, err := NewEndpointWhitelist(
		[]string{"GET /health*", "GET /api/v1/health*", "GET /static/*", "POST /api/v1/tenants", "GET /admin/*"},
		[]string{"POST /api/v1/auth/login", "POST /api/v1/auth/refresh", "GET /api/v1/version"},
	)
	if err != nil {
		logger.Warn("Failed to create endpoint whitelist, using default", map[string]interface{}{"error": err.Error()})
		whitelist = DefaultWhitelist()
	}

	return &FiberMiddleware{
		config:        cfg,
		tenantService: tenantService,
		iamService:    iamService,
		store:         store,
		logger:        logger,
		metrics:       metrics,
		tracing:       tracing,
		whitelist:     whitelist,
	}
}

// TenantMiddleware handles tenant context for Fiber
func (m *FiberMiddleware) TenantMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Bypass tenant validation for public endpoints
		if strings.HasPrefix(c.Path(), "/admin/") ||
			strings.HasPrefix(c.Path(), "/static/") ||
			strings.HasPrefix(c.Path(), "/favicon.ico") {
			return c.Next()
		}

		// Check if this is a public endpoint
		if m.whitelist != nil && m.whitelist.IsPublicEndpoint(c.Method(), c.Path()) {
			m.logger.Debug("Public endpoint, bypassing tenant middleware", logger.Fields{
				"method": c.Method(),
				"path":   c.Path(),
			})
			return c.Next()
		}

		// Special case: Tenant creation endpoints
		if (c.Method() == "POST" && c.Path() == "/api/v1/tenants") ||
			(c.Method() == "POST" && c.Path() == "/api/v1/tenant/onboard") {
			return c.Next()
		}

		// Extract tenant ID from request
		tenantIDStr, err := m.extractTenantIDFromFiber(c)
		if err != nil {
			m.logger.Warn("Failed to extract tenant ID", logger.Fields{
				"error":  err.Error(),
				"method": c.Method(),
				"path":   c.Path(),
				"host":   c.Hostname(),
			})
			return c.Status(400).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "tenant_resolution_failed",
					"message": "Tenant resolution failed: " + err.Error(),
				},
				"metadata": fiber.Map{
					"request_id": c.Locals("requestid"),
					"timestamp":  time.Now(),
					"version":    "1.0",
				},
			})
		}

		// Parse tenant ID as UUID
		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			m.logger.Warn("Invalid tenant ID format", logger.Fields{
				"tenant_id": tenantIDStr,
				"error":     err.Error(),
			})
			return c.Status(400).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "invalid_input",
					"message": "Invalid tenant ID format",
				},
				"metadata": fiber.Map{
					"request_id": c.Locals("requestid"),
					"timestamp":  time.Now(),
					"version":    "1.0",
				},
			})
		}

		// Validate tenant exists
		tenantEntity, err := m.tenantService.GetTenantByID(c.Context(), tenantID)
		if err != nil {
			if errors.Is(err, sharedErrors.ErrNotFound) {
				return c.Status(404).JSON(fiber.Map{
					"error": fiber.Map{
						"code":    "not_found",
						"message": "Tenant not found",
					},
					"metadata": fiber.Map{
						"request_id": c.Locals("requestid"),
						"timestamp":  time.Now(),
						"version":    "1.0",
					},
				})
			}

			m.logger.Error("Failed to validate tenant", logger.Fields{
				"tenant_id": tenantID.String(),
				"error":     err.Error(),
			})
			return c.Status(500).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "internal_error",
					"message": "Tenant validation failed",
				},
				"metadata": fiber.Map{
					"request_id": c.Locals("requestid"),
					"timestamp":  time.Now(),
					"version":    "1.0",
				},
			})
		}

		// Set tenant context
		ctx := shared.WithTenantID(c.Context(), tenantID)
		c.SetUserContext(ctx)

		// Set database session variable for Row-Level Security (RLS)
		if err := m.store.SetTenantContextFromCtx(ctx); err != nil {
			m.logger.Error("Failed to set database tenant context", logger.Fields{
				"tenant_id": tenantID.String(),
				"error":     err.Error(),
			})
			return c.Status(500).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "internal_error",
					"message": "Database context setup failed",
				},
				"metadata": fiber.Map{
					"request_id": c.Locals("requestid"),
					"timestamp":  time.Now(),
					"version":    "1.0",
				},
			})
		}

		// Store tenant information in Fiber locals
		c.Locals("tenant_id", tenantID)
		c.Locals("tenant", tenantEntity)

		m.logger.Debug("Tenant context established", logger.Fields{
			"tenant_id":   tenantID.String(),
			"tenant_name": tenantEntity.Name,
			"method":      c.Method(),
			"path":        c.Path(),
		})

		return c.Next()
	}
}

// JWTAuthMiddleware handles JWT authentication for Fiber
func (m *FiberMiddleware) JWTAuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if this is a public endpoint
		if m.whitelist != nil && m.whitelist.IsPublicEndpoint(c.Method(), c.Path()) {
			return c.Next()
		}

		// Extract token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "unauthorized",
					"message": "Authorization token required",
				},
				"metadata": fiber.Map{
					"request_id": c.Locals("requestid"),
					"timestamp":  time.Now(),
					"version":    "1.0",
				},
			})
		}

		// Remove Bearer prefix
		token := strings.TrimPrefix(authHeader, "Bearer ")
		token = strings.TrimSpace(token)

		if token == "" {
			return c.Status(401).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "unauthorized",
					"message": "Invalid authorization token format",
				},
				"metadata": fiber.Map{
					"request_id": c.Locals("requestid"),
					"timestamp":  time.Now(),
					"version":    "1.0",
				},
			})
		}

		// Validate token using IAM service
		result, err := m.iamService.Authentication().ValidateToken(c.Context(), token)
		if err != nil {
			m.logger.Warn("JWT token validation failed", logger.Fields{
				"error": err.Error(),
			})
			return c.Status(401).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "unauthorized",
					"message": "Invalid or expired token",
				},
				"metadata": fiber.Map{
					"request_id": c.Locals("requestid"),
					"timestamp":  time.Now(),
					"version":    "1.0",
				},
			})
		}

		if !result.Valid {
			return c.Status(401).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "unauthorized",
					"message": "Invalid token",
				},
				"metadata": fiber.Map{
					"request_id": c.Locals("requestid"),
					"timestamp":  time.Now(),
					"version":    "1.0",
				},
			})
		}

		// Get user details
		user, err := m.iamService.Authentication().GetUser(c.Context(), result.UserID)
		if err != nil {
			m.logger.Warn("Failed to get user details", logger.Fields{
				"user_id": result.UserID.String(),
				"error":   err.Error(),
			})
			return c.Status(401).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "unauthorized",
					"message": "User lookup failed",
				},
				"metadata": fiber.Map{
					"request_id": c.Locals("requestid"),
					"timestamp":  time.Now(),
					"version":    "1.0",
				},
			})
		}

		// Check if account is locked
		locked, err := m.iamService.Authentication().IsAccountLocked(c.Context(), user.ID)
		if err != nil {
			m.logger.Error("Failed to check account lock status", logger.Fields{
				"user_id": user.ID.String(),
				"error":   err.Error(),
			})
		}
		if locked {
			return c.Status(403).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "account_locked",
					"message": "Account is locked",
				},
				"metadata": fiber.Map{
					"request_id": c.Locals("requestid"),
					"timestamp":  time.Now(),
					"version":    "1.0",
				},
			})
		}

		// Store authentication information in Fiber locals
		c.Locals("authenticated", true)
		c.Locals("user_id", user.ID)
		c.Locals("user", user)
		c.Locals("user_email", user.Email)
		c.Locals("auth_claims", result.Claims)
		c.Locals("auth_token", token)
		c.Locals("auth_time", time.Now())

		m.logger.Debug("JWT authentication successful", logger.Fields{
			"user_id":       user.ID.String(),
			"email":         user.Email,
			"authenticated": true,
		})

		return c.Next()
	}
}

// CORSMiddleware handles CORS for Fiber with multi-tenant support
func (m *FiberMiddleware) CORSMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		origin := c.Get("Origin")

		if origin != "" && m.isAllowedOrigin(origin, c) {
			c.Set("Access-Control-Allow-Origin", origin)
		}

		c.Set("Access-Control-Allow-Methods", "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Origin,Content-Type,Accept,Authorization,X-Tenant-ID,Accept-Encoding,X-Request-ID")
		c.Set("Access-Control-Expose-Headers", "X-Total-Count,X-Rate-Limit-Remaining,X-Request-ID")
		c.Set("Access-Control-Allow-Credentials", "true")
		c.Set("Access-Control-Max-Age", "86400")

		// Handle preflight requests
		if c.Method() == "OPTIONS" {
			return c.SendStatus(204)
		}

		return c.Next()
	}
}

// ValidationMiddleware handles request validation
func (m *FiberMiddleware) ValidationMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Set Content-Type to application/json for API requests if not set
		if strings.HasPrefix(c.Path(), "/api/") && c.Get("Content-Type") == "" {
			if c.Method() == "POST" || c.Method() == "PUT" || c.Method() == "PATCH" {
				c.Set("Content-Type", "application/json")
			}
		}

		// Validate Content-Type for POST/PUT/PATCH requests
		if c.Method() == "POST" || c.Method() == "PUT" || c.Method() == "PATCH" {
			contentType := c.Get("Content-Type")
			if strings.HasPrefix(c.Path(), "/api/") && !strings.Contains(contentType, "application/json") {
				return c.Status(400).JSON(fiber.Map{
					"error": fiber.Map{
						"code":    "invalid_input",
						"message": "Content-Type must be application/json for API requests",
					},
					"metadata": fiber.Map{
						"request_id": c.Locals("requestid"),
						"timestamp":  time.Now(),
						"version":    "1.0",
					},
				})
			}
		}

		return c.Next()
	}
}

// Helper methods

func (m *FiberMiddleware) extractTenantIDFromFiber(c *fiber.Ctx) (string, error) {
	// Priority 1: Check X-Tenant-ID header
	if tenantID := c.Get("X-Tenant-ID"); tenantID != "" {
		tenantID = strings.TrimSpace(tenantID)
		if tenantID != "" {
			return tenantID, nil
		}
	}

	// Priority 2: Extract from subdomain
	return extractFromSubdomain(c.Hostname())
}

func (m *FiberMiddleware) isAllowedOrigin(origin string, c *fiber.Ctx) bool {
	if origin == "" {
		return false
	}

	// Development mode - be more permissive
	if m.config.App.Stage == config.DevelopmentStage {
		if strings.HasPrefix(origin, "http://localhost") ||
			strings.HasPrefix(origin, "http://127.0.0.1") ||
			strings.HasPrefix(origin, "https://localhost") {
			return true
		}
	}

	// Production allowed origins
	allowedOrigins := []string{
		"https://app.domain.so",
		"https://console.domain.so",
	}

	for _, allowedOrigin := range allowedOrigins {
		if origin == allowedOrigin {
			return true
		}
	}

	// Multi-tenant subdomain support
	return m.isAllowedTenantSubdomain(origin)
}

func (m *FiberMiddleware) isAllowedTenantSubdomain(origin string) bool {
	baseDomain := "app.domain.so" // Configure this based on your domain
	if !strings.HasPrefix(origin, "https://") {
		return false
	}

	hostname := strings.TrimPrefix(origin, "https://")
	if strings.HasSuffix(hostname, "."+baseDomain) {
		tenantPart := strings.TrimSuffix(hostname, "."+baseDomain)
		return m.isValidTenantSubdomain(tenantPart)
	}

	return false
}

func (m *FiberMiddleware) isValidTenantSubdomain(subdomain string) bool {
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

// Helper functions to extract values from Fiber context

func GetTenantIDFromFiber(c *fiber.Ctx) (uuid.UUID, bool) {
	if tenantID, ok := c.Locals("tenant_id").(uuid.UUID); ok {
		return tenantID, true
	}
	return uuid.Nil, false
}

func GetUserFromFiber(c *fiber.Ctx) (interface{}, bool) {
	if user := c.Locals("user"); user != nil {
		return user, true
	}
	return nil, false
}

func GetUserIDFromFiber(c *fiber.Ctx) (uuid.UUID, bool) {
	if userID, ok := c.Locals("user_id").(uuid.UUID); ok {
		return userID, true
	}
	return uuid.Nil, false
}

func IsAuthenticatedFiber(c *fiber.Ctx) bool {
	if authenticated, ok := c.Locals("authenticated").(bool); ok {
		return authenticated
	}
	return false
}

func GetAuthTokenFromFiber(c *fiber.Ctx) (string, bool) {
	if token, ok := c.Locals("auth_token").(string); ok {
		return token, true
	}
	return "", false
}