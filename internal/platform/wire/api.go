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

	db "awo/db/sqlc"
	"awo/internal/api/handlers"
	"awo/internal/api/middleware"
	"awo/internal/core/authz"
	financeService "awo/internal/core/finance/service"
	"awo/internal/core/iam"
	"awo/internal/core/identity"
	"awo/internal/core/identity/session"
	"awo/internal/core/tenant"
	"awo/internal/platform/cache"
	"awo/internal/platform/config"
	"awo/internal/shared/logger"
	"awo/internal/shared/metrics"
	"awo/internal/shared/tracing"
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
// IDENTITY / AUTHZ / SESSION PROVIDERS
// ============================================================================

// NewIdentityRepository constructs the identity repository.
func NewIdentityRepository(store db.Store, cacheSvc cache.Service, tracer tracing.Service, m metrics.MetricsProvider) identity.Repository {
	return identity.NewRepository(store, cacheSvc, tracer, m)
}

// NewIdentityService constructs the identity service.
func NewIdentityService(repo identity.Repository, cacheSvc cache.Service, tracer tracing.Service, m metrics.MetricsProvider) identity.Service {
	return identity.NewService(repo, cacheSvc, tracer, m)
}

// NewAuthzService constructs the Casbin-backed authorization service.
func NewAuthzService(store db.Store, cacheSvc cache.Service, log logger.Logger, m metrics.MetricsProvider, tracer tracing.Service) (authz.Service, error) {
	return authz.New(authz.Config{
		Store:   store,
		Cache:   cacheSvc,
		Logger:  log,
		Metrics: m,
		Tracer:  tracer,
	})
}

// NewSessionRepository constructs the session repository.
func NewSessionRepository(store db.Store, cacheSvc cache.Service, tracer tracing.Service, m metrics.MetricsProvider) session.Repository {
	return session.NewRepository(store, cacheSvc, tracer, m)
}

// NewSessionService constructs the session service.
func NewSessionService(
	identitySvc identity.Service,
	authzSvc authz.Service,
	repo session.Repository,
	cacheSvc cache.Service,
	tracer tracing.Service,
	m metrics.MetricsProvider,
	log logger.Logger,
) session.Service {
	return session.New(identitySvc, authzSvc, repo, cacheSvc, tracer, m, log)
}

// ============================================================================
// HANDLER PROVIDERS
// ============================================================================

// NewHandlerDependencies creates handler dependencies
func NewHandlerDependencies(
	log logger.Logger,
	m metrics.MetricsProvider,
	tracer tracing.Service,
	tenantService tenant.Service,
	iamService iam.Service,
	sessionSvc session.Service,
	financeServices *financeService.Services,
	tenantMiddleware fiber.Handler,
) *handlers.Dependencies {
	authCfg := middleware.DefaultAuthConfig(sessionSvc)
	return &handlers.Dependencies{
		Logger:           log,
		Metrics:          m,
		Tracer:           tracer,
		TenantService:    tenantService,
		UserService:      iamService.Authentication(),
		FinanceServices:  financeServices,
		TenantMiddleware: tenantMiddleware,
		SessionService:   sessionSvc,
		AuthConfig:       &authCfg,
	}
}

// NewRouter creates a new router with all handlers
func NewRouter(deps *handlers.Dependencies) (*handlers.Router, error) {
	return handlers.NewRouter(deps)
}
