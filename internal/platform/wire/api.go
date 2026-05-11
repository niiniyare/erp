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

	temporalclient "go.temporal.io/sdk/client"

	db "awo.so/db/sqlc"
	"awo.so/internal/api/handlers"
	"awo.so/internal/api/middleware"
	"awo.so/internal/core/audit"
	financeService "awo.so/internal/core/finance/service"
	"awo.so/internal/core/iam"
	"awo.so/internal/core/tenant"
	"awo.so/internal/platform/cache"
	"awo.so/internal/platform/config"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
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
		AllowOrigins:     cfg.Server.AllowedOrigins,
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

// NewRouteSecurityManager creates the route security manager.
// CORS is disabled here because it is already applied globally in setupGlobalMiddleware.
func NewRouteSecurityManager(log logger.Logger, m metrics.MetricsProvider, tracer tracing.Service) *middleware.RouteSecurityManager {
	cfg := middleware.DefaultRouteSecurityConfig()
	cfg.API.EnableCORS = false
	return middleware.NewRouteSecurityManager(cfg, log, m, tracer)
}

// ============================================================================
// IDENTITY / AUTHZ / SESSION PROVIDERS
// ============================================================================

// NewIdentityRepository constructs the IAM user repository.
func NewIdentityRepository(store db.Store, cacheSvc cache.Service, tracer tracing.Service, m metrics.MetricsProvider) iam.UserRepository {
	return iam.NewUserRepository(store, cacheSvc, tracer, m)
}

// NewIdentityService constructs the IAM user service with brute-force config from app config.
// cacheSvc is accepted for wire compatibility but cache is handled by the repository.
// authzSvc is used to bootstrap tenant_admin role on user creation.
func NewIdentityService(repo iam.UserRepository, authzSvc iam.AuthzService, cacheSvc cache.Service, tracer tracing.Service, m metrics.MetricsProvider, cfg *config.Config, log logger.Logger) iam.UserService {
	return iam.NewUserServiceWithConfig(repo, authzSvc, tracer, m, iam.UserConfig{
		MaxFailedAttempts: cfg.Auth.MaxFailedAttempts,
		LockoutDuration:   cfg.Auth.LockoutDuration,
		MFAEncryptionKey:  []byte(cfg.Auth.MFAEncryptionKey),
		MFAIssuer:         cfg.Auth.MFAIssuer,
	}, log)
}

// NewAuthzService constructs the Casbin-backed authorization service.
func NewAuthzService(store db.Store, cacheSvc cache.Service, log logger.Logger, m metrics.MetricsProvider, tracer tracing.Service) (iam.AuthzService, error) {
	return iam.New(iam.Config{
		Store:   store,
		Cache:   cacheSvc,
		Logger:  log,
		Metrics: m,
		Tracer:  tracer,
	})
}

// NewSessionRepository constructs the session repository.
func NewSessionRepository(store db.Store, cacheSvc cache.Service, tracer tracing.Service, m metrics.MetricsProvider) iam.SessionRepository {
	return iam.NewSessionRepository(store, cacheSvc, tracer, m)
}

// NewSessionService constructs the session service with TTL from app config.
func NewSessionService(
	identitySvc iam.UserService,
	repo iam.SessionRepository,
	tracer tracing.Service,
	m metrics.MetricsProvider,
	log logger.Logger,
	cfg *config.Config,
) iam.SessionService {
	return iam.NewSessionServiceWithConfig(identitySvc, repo, tracer, m, log, iam.SessionConfig{
		SessionTTL: cfg.Auth.SessionTTL,
		CookieName: cfg.Auth.CookieName,
	})
}

// ============================================================================
// AUDIT PROVIDERS
// ============================================================================

// NewAuditRepository constructs an audit repository.
func NewAuditRepository(store db.Store, log logger.Logger, tracer tracing.Service, m metrics.MetricsProvider) audit.Repository {
	return audit.NewRepository(store, log, tracer, m)
}

// NewAuditService constructs the audit service.
func NewAuditService(repo audit.Repository, cacheSvc cache.Service, log logger.Logger, tracer tracing.Service, m metrics.MetricsProvider) audit.Service {
	return audit.NewService(repo, cacheSvc, log, tracer, m)
}

// ============================================================================
// API KEY PROVIDERS
// ============================================================================

// NewAPIKeyRepository constructs the API key repository.
func NewAPIKeyRepository(store db.Store) iam.APIKeyRepository {
	return iam.NewAPIKeyRepository(store)
}

// NewAPIKeyService constructs the API key service.
func NewAPIKeyService(repo iam.APIKeyRepository, cacheSvc cache.Service, tracer tracing.Service, m metrics.MetricsProvider) iam.APIKeyService {
	return iam.NewAPIKeyService(repo, cacheSvc, tracer, m)
}

// ============================================================================
// SSO PROVIDERS
// ============================================================================

// NewSSORepository constructs a Postgres-backed SSORepository.
func NewSSORepository(store db.Store) iam.SSORepository {
	return iam.NewSSORepository(store)
}

// NewSSOService constructs the SSOService for OAuth/OIDC login flows.
func NewSSOService(
	repo iam.SSORepository,
	identity iam.UserService,
	cacheSvc cache.Service,
	tracer tracing.Service,
	m metrics.MetricsProvider,
	log logger.Logger,
	cfg *config.Config,
) iam.SSOService {
	return iam.NewSSOService(repo, identity, tracer, m, log, iam.SSOConfig{
		EncryptionKey: []byte(cfg.Auth.SSOEncryptionKey),
		Cache:         cacheSvc,
	})
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
	userSvc iam.UserService,
	sessionSvc iam.SessionService,
	financeServices *financeService.Services,
	tenantMiddleware fiber.Handler,
	securityMgr *middleware.RouteSecurityManager,
	auditSvc audit.Service,
	apiKeySvc iam.APIKeyService,
	ssoSvc iam.SSOService,
	temporalClient temporalclient.Client,
	store db.Store,
) *handlers.Dependencies {
	authCfg := middleware.DefaultAuthConfig(sessionSvc)
	return &handlers.Dependencies{
		Logger:           log,
		Metrics:          m,
		Tracer:           tracer,
		TenantService:    tenantService,
		UserService:      userSvc,
		FinanceServices:  financeServices,
		TenantMiddleware: tenantMiddleware,
		SecurityManager:  securityMgr,
		SessionService:   sessionSvc,
		AuthConfig:       &authCfg,
		AuditService:     auditSvc,
		APIKeyService:    apiKeySvc,
		SSOService:       ssoSvc,
		TemporalClient:   temporalClient,
		Store:            store,
	}
}

// NewRouter creates a new router with all handlers
func NewRouter(deps *handlers.Dependencies) (*handlers.Router, error) {
	return handlers.NewRouter(deps)
}
