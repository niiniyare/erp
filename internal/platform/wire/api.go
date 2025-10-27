// Package wire - API layer providers
package wire

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"github.com/niiniyare/erp/internal/api/handlers"
	"github.com/niiniyare/erp/internal/api/middleware"
	"github.com/niiniyare/erp/internal/core/finance/service"
	"github.com/niiniyare/erp/internal/core/iam"
	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/platform/config"
	sharedLogger "github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// ============================================================================
// FIBER APP PROVIDER
// ============================================================================

// NewFiberApp creates a new Fiber application with middleware
func NewFiberApp(cfg *config.Config, log sharedLogger.Logger) *fiber.App {
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
			log.Error("Fiber error", sharedLogger.Fields{
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
func setupGlobalMiddleware(app *fiber.App, cfg *config.Config, log sharedLogger.Logger) {
	// Request ID middleware
	app.Use(requestid.New())

	// Recovery middleware
	app.Use(recover.New())

	// Helmet for security headers
	app.Use(helmet.New())

	// CORS middleware
	app.Use(cors.New(cors.Config{
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

	// Rate limiting middleware (for production)
	if cfg.App.Stage != config.DevelopmentStage {
		app.Use(limiter.New(limiter.Config{
			Max:        100,
			Expiration: 1 * time.Minute,
			KeyGenerator: func(c *fiber.Ctx) string {
				return c.IP()
			},
		}))
	}

	// Request logging middleware
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} (${latency})\n",
		Output: log.GetWriter(),
	}))
}

// ============================================================================
// MIDDLEWARE PROVIDERS
// ============================================================================

// MiddlewareConfig contains middleware configuration
type MiddlewareConfig struct {
	Config        *config.Config
	Logger        sharedLogger.Logger
	Metrics       metrics.MetricsProvider
	Tracer        tracing.TracingService
	TenantService tenant.Service
	IAMService    iam.Service
}

// NewMiddlewareConfig creates middleware configuration
func NewMiddlewareConfig(
	cfg *config.Config,
	log sharedLogger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.TracingService,
	tenantService tenant.Service,
	iamService iam.Service,
) *MiddlewareConfig {
	return &MiddlewareConfig{
		Config:        cfg,
		Logger:        log,
		Metrics:       metrics,
		Tracer:        tracer,
		TenantService: tenantService,
		IAMService:    iamService,
	}
}

// ============================================================================
// HANDLER PROVIDERS
// ============================================================================

// NewHandlerDependencies creates handler dependencies
func NewHandlerDependencies(
	log sharedLogger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.TracingService,
	tenantService tenant.Service,
	iamService iam.Service,
	financeServices *service.Services,
) *handlers.Dependencies {
	return &handlers.Dependencies{
		Logger:        log,
		Metrics:       metrics,
		Tracer:        tracer,
		TenantService: tenantService,
		// Add other services as needed
	}
}

// NewRouter creates a new router with all handlers
func NewRouter(deps *handlers.Dependencies) (*handlers.Router, error) {
	return handlers.NewRouter(deps)
}

// ============================================================================
// TENANT-SCOPED PROVIDERS
// ============================================================================

// NewTenantScopedDBStore creates a tenant-scoped database store
// This is used for operations that need tenant context
func NewTenantScopedDBStore(store *sqlc.Store, tenantID string) *TenantScopedStore {
	return &TenantScopedStore{
		Store:    store,
		TenantID: tenantID,
	}
}

// TenantScopedStore wraps the database store with tenant context
type TenantScopedStore struct {
	Store    *sqlc.Store
	TenantID string
}

// NewTenantScopedCache creates a tenant-scoped cache service
func NewTenantScopedCache(cache cache.Service, tenantID string) *TenantScopedCache {
	return &TenantScopedCache{
		Cache:    cache,
		TenantID: tenantID,
	}
}

// TenantScopedCache wraps the cache service with tenant-specific keys
type TenantScopedCache struct {
	Cache    cache.Service
	TenantID string
}

// NewTenantAwareRepositories creates tenant-aware repository wrappers
func NewTenantAwareRepositories(
	tenantStore *TenantScopedStore,
	tenantCache *TenantScopedCache,
) *TenantAwareRepositories {
	return &TenantAwareRepositories{
		Store: tenantStore,
		Cache: tenantCache,
	}
}

// TenantAwareRepositories contains tenant-aware repository implementations
type TenantAwareRepositories struct {
	Store *TenantScopedStore
	Cache *TenantScopedCache
}