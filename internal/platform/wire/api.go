// Package wire - API layer providers
package wire

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"github.com/niiniyare/erp/internal/api/handlers"
	"github.com/niiniyare/erp/internal/api/middleware"
	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/core/iam"
	financeService "github.com/niiniyare/erp/internal/core/finance/service"
	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/platform/config"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// ============================================================================
// FIBER APP PROVIDER
// ============================================================================

// NewFiberApp creates a new Fiber application with middleware
func NewFiberApp(cfg *config.Config, log logger.Logger) *fiber.App {
	// Fiber configuration
	fiberConfig := fiber.Config{
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
		ServerHeader: "Awo-ERP",
		AppName:      cfg.App.Name + " v" + cfg.App.Version,

		// Error handling
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Log the error
			log.Error("Fiber error", logger.Fields{
				"error":  err.Error(),
				"path":   c.Path(),
				"method": c.Method(),
				"ip":     c.IP(),
			})

			// Default error response
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			return c.Status(code).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		},
	}

	app := fiber.New(fiberConfig)

	// Global middleware
	setupGlobalMiddleware(app, cfg, log)

	return app
}

// setupGlobalMiddleware configures global Fiber middleware
func setupGlobalMiddleware(app *fiber.App, cfg *config.Config, log logger.Logger) {
	// Request ID middleware
	app.Use(requestid.New())

	// Recovery middleware
	app.Use(recover.New())

	// Helmet for security headers
	app.Use(helmet.New())

	// CORS middleware
	app.Use(cors.New(cors.Config{
		// TODO: AllowOrigins should be fetched from file
		AllowOrigins:     "http://localhost:3000,https://app.example.com",
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Request-ID,X-TenantID",
		AllowCredentials: true,
		MaxAge:           86400,
	}))

	// Compression middleware
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))
}

// ============================================================================
// MIDDLEWARE PROVIDERS
// ============================================================================

// NewTenantMiddlewareConfig creates tenant middleware configuration
func NewTenantMiddlewareConfig(
	tenantService tenant.Service,
	store db.Store,
) middleware.TenantMiddlewareConfig {
	// Create default whitelist for public endpoints
	whitelist := middleware.DefaultWhitelist()

	return middleware.TenantMiddlewareConfig{
		TenantService: tenantService,
		Store:         store,
		Whitelist:     whitelist,
		CacheTTL:      5 * time.Minute,
		EnableCache:   true,
		SkipPaths: []string{
			"/health",
			"/api/v1/health",
			"/openapi",
			"/swagger-ui",
			"/debug",
		},
	}
}

// NewTenantMiddleware creates the tenant middleware instance
func NewTenantMiddleware(config middleware.TenantMiddlewareConfig) fiber.Handler {
	return middleware.TenantMiddleware(config)
}

// ============================================================================
// HANDLER PROVIDERS
// ============================================================================

// NewHandlerDependencies creates handler dependencies
func NewHandlerDependencies(
	log logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.Service,
	tenantService tenant.Service,
	iamService iam.Service,
	financeServices *financeService.Services,
	tenantMiddleware fiber.Handler,
) *handlers.Dependencies {
	return &handlers.Dependencies{
		Logger:          log,
		Metrics:         metrics,
		Tracer:          tracer,
		TenantService:   tenantService,
		UserService:     iamService.Authentication(), // Get authn service from IAM
		FinanceServices: financeServices,
		TenantMiddleware: tenantMiddleware,
		// TODO: Add SecurityManager when implemented
		// TODO: Add health config when implemented
	}
}

// NewRouter creates a new router with all handlers
func NewRouter(deps *handlers.Dependencies) (*handlers.Router, error) {
	return handlers.NewRouter(deps)
}
