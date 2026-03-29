package handlers

import (
	"fmt"
	"sync"

	"github.com/gofiber/fiber/v2"

	authHandler "awo.so/internal/api/handlers/auth"
	auditHandler "awo.so/internal/api/handlers/audit"
	financeHandler "awo.so/internal/api/handlers/finance"
	"awo.so/internal/api/handlers/health"
	schemaHandler "awo.so/internal/api/handlers/schema"
	tenantHandler "awo.so/internal/api/handlers/tenant"
	uiHandler "awo.so/internal/api/handlers/ui"
	userHandler "awo.so/internal/api/handlers/user"
	middlewarePkg "awo.so/internal/api/middleware"
	db "awo.so/db/sqlc"
	"awo.so/internal/core/audit"
	"awo.so/internal/core/entity"
	financeService "awo.so/internal/core/finance/service"

	"awo.so/internal/core/iam"
	coreTenant "awo.so/internal/core/tenant"
	"awo.so/internal/shared/errors"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

// Module names as constants for consistency
const (
	ModuleHealth  = "health"
	ModuleTenant  = "tenant"
	ModuleUser    = "user"
	ModuleFinance = "finance"
	ModuleAuth    = "auth"
)

// ============================================================================
// ROUTE REGISTRY
// ============================================================================

// ModuleInfo contains information about a registered module
type ModuleInfo struct {
	Name       string         `json:"name"`
	BasePath   string         `json:"base_path"`
	Middleware []string       `json:"middleware"`
	RouteCount int            `json:"route_count"`
	Registered bool           `json:"registered"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

// RouteRegistry manages route registration and provides module information
type RouteRegistry struct {
	logger            logger.Logger
	metrics           metrics.MetricsProvider
	tracer            tracing.Service
	registeredModules map[string]*ModuleInfo
	registeredPaths   map[string]bool
	mu                sync.RWMutex
}

// NewRouteRegistry creates a new route registry
func NewRouteRegistry(logger logger.Logger, metrics metrics.MetricsProvider, tracer tracing.Service) *RouteRegistry {
	return &RouteRegistry{
		logger:            logger,
		metrics:           metrics,
		tracer:            tracer,
		registeredModules: make(map[string]*ModuleInfo),
		registeredPaths:   make(map[string]bool),
	}
}

// RegisterModuleWithMiddleware registers a module with middleware
func (r *RouteRegistry) RegisterModuleWithMiddleware(
	app *fiber.App,
	moduleName string,
	basePath string,
	middleware []string,
	setupRoutes func(fiber.Router),
	deps *Dependencies,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if module already registered
	if info, exists := r.registeredModules[moduleName]; exists && info.Registered {
		return errors.NewBusinessError("MODULE_ALREADY_REGISTERED",
			fmt.Sprintf("module %s is already registered", moduleName)).
			WithCategory(errors.CategorySystem).
			WithDetail("module_name", moduleName)
	}

	// Check for path conflicts
	if r.registeredPaths[basePath] {
		return errors.NewBusinessError("PATH_CONFLICT",
			fmt.Sprintf("path %s is already registered", basePath)).
			WithCategory(errors.CategorySystem).
			WithDetail("base_path", basePath)
	}

	// Create router group for the module
	router := app.Group(basePath)

	// NOTE: Legacy middleware switching is now replaced by RouteSecurityManager
	// The security manager handles all middleware configuration based on route groups:
	// - Public routes: minimal security (health, metrics)
	// - API routes: full security (auth, tenant, rate limiting, observability)
	// - UI routes: session-based security (CSRF, security headers)

	// Apply middleware instances based on middleware names (legacy support)
	for _, middlewareName := range middleware {
		switch middlewareName {
		case "observability":
			// Apply comprehensive observability middleware
			observabilityConfig := middlewarePkg.DefaultObservabilityConfig()
			observabilityConfig.ServiceName = "erp-api"
			observabilityConfig.DetailedLogging = true
			observabilityMiddleware := middlewarePkg.CreateObservabilityMiddleware(
				r.logger, r.metrics, r.tracer, &observabilityConfig)
			router.Use(observabilityMiddleware)
		case "cors":
			// Use development CORS config for tests
			corsConfig := middlewarePkg.DevelopmentCORSConfig([]int{3000, 8080})
			router.Use(middlewarePkg.NewCORSMiddleware(corsConfig))
		case "auth":
			// TODO: Apply auth middleware when available
			// router.Use(middleware.AuthMiddleware())
		case "ratelimit":
			// Apply rate limiting middleware using existing system
			// TODO: Configure rate limiting with proper config
			// router.Use(middlewarePkg.RateLimitMiddleware(config, logger))
		case "tenant":
			// Apply tenant middleware for RLS context
			if deps != nil && deps.TenantMiddleware != nil {
				router.Use(deps.TenantMiddleware)
			}
		default:
			// Log unknown middleware but don't fail
			r.logger.Info(fmt.Sprintf("unknown middleware: %s", middlewareName))
		}
	}

	// Setup routes for the module
	setupRoutes(router)

	// Register module info
	info := &ModuleInfo{
		Name:       moduleName,
		BasePath:   basePath,
		Middleware: middleware,
		RouteCount: 1, // Simplified - would count actual routes
		Registered: true,
		Metadata:   make(map[string]any),
	}

	r.registeredModules[moduleName] = info
	r.registeredPaths[basePath] = true

	// Record metrics
	r.metrics.IncrementCounter("modules_registered_total", metrics.Fields{
		"module_name": moduleName,
	})

	r.logger.Info(fmt.Sprintf("registered module: %s at %s", moduleName, basePath))

	return nil
}

// ListRoutes returns all registered modules
func (r *RouteRegistry) ListRoutes() map[string]*ModuleInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]*ModuleInfo)
	for name, info := range r.registeredModules {
		// Deep copy the info
		infoCopy := &ModuleInfo{
			Name:       info.Name,
			BasePath:   info.BasePath,
			Middleware: make([]string, len(info.Middleware)),
			RouteCount: info.RouteCount,
			Registered: info.Registered,
			Metadata:   make(map[string]any),
		}
		copy(infoCopy.Middleware, info.Middleware)
		for k, v := range info.Metadata {
			infoCopy.Metadata[k] = v
		}
		result[name] = infoCopy
	}

	return result
}

// Dependencies contains all required dependencies for handlers.
// All fields are required and must be non-nil.
type Dependencies struct {
	Logger           logger.Logger
	Metrics          metrics.MetricsProvider
	Tracer           tracing.Service
	conf             *health.Config
	TenantService    coreTenant.Service
	UserService      iam.UserService
	FinanceServices  *financeService.Services
	TenantMiddleware fiber.Handler
	SecurityManager  *middlewarePkg.RouteSecurityManager

	// Session-based auth (S6/S8 — replaces JWT+IAM approach)
	// Set these to enable Authenticate/Authorize middleware on protected routes.
	SessionService iam.SessionService
	AuthConfig     *middlewarePkg.AuthConfig // nil = auth middleware disabled

	// SSOService enables OAuth/OIDC login routes.
	// Optional — SSO routes are skipped when nil.
	SSOService iam.SSOService

	// APIKeyService enables API key management routes and Bearer token validation.
	// Optional — API key routes are skipped when nil.
	APIKeyService iam.APIKeyService

	// Store is the raw DB store used by schema endpoints (BootHandler).
	// Optional — schema routes are skipped when nil.
	Store db.Store

	// AuditService enables the GET /api/v1/audit-logs endpoint.
	// Optional — audit routes are skipped when nil.
	AuditService audit.Service
}

// Validate ensures all required dependencies are present.
func (d *Dependencies) Validate() error {
	if d == nil {
		return errors.NewBusinessError("INVALID_DEPENDENCIES", "dependencies cannot be nil").
			WithCategory(errors.CategorySystem).
			WithSeverity(errors.SeverityCritical).
			WithSuggestion("Ensure dependencies struct is properly initialized before creating router")
	}

	if d.Logger == nil {
		return errors.NewBusinessError("MISSING_LOGGER", "logger is required").
			WithCategory(errors.CategorySystem).
			WithSeverity(errors.SeverityCritical).
			WithSuggestion("Ensure logger is initialized before creating handler dependencies")
	}

	if d.Metrics == nil {
		return errors.NewBusinessError("MISSING_METRICS", "metrics provider is required").
			WithCategory(errors.CategorySystem).
			WithSeverity(errors.SeverityCritical).
			WithSuggestion("Ensure metrics provider is initialized before creating handler dependencies")
	}

	if d.Tracer == nil {
		return errors.NewBusinessError("MISSING_TRACER", "tracing service is required").
			WithCategory(errors.CategorySystem).
			WithSeverity(errors.SeverityCritical).
			WithSuggestion("Ensure tracing service is initialized before creating handler dependencies")
	}

	return nil
}

// Router wraps the route registry and provides module registration.
type Router struct {
	registry *RouteRegistry
	deps     *Dependencies
}

// NewRouter creates a new router with the given dependencies.
func NewRouter(deps *Dependencies) (*Router, error) {
	if err := deps.Validate(); err != nil {
		return nil, err
	}

	return &Router{
		registry: NewRouteRegistry(deps.Logger, deps.Metrics, deps.Tracer),
		deps:     deps,
	}, nil
}

// RegisterAll registers all application routes with proper security configuration.
func (r *Router) RegisterAll(app *fiber.App) error {
	// Register route groups with different security configurations
	if err := r.registerPublicRoutes(app); err != nil {
		return fmt.Errorf("failed to register public routes: %w", err)
	}

	if err := r.registerAPIRoutes(app); err != nil {
		return fmt.Errorf("failed to register API routes: %w", err)
	}

	if err := r.registerUIRoutes(app); err != nil {
		return fmt.Errorf("failed to register UI routes: %w", err)
	}

	return nil
}

// registerPublicRoutes registers public routes with minimal security
func (r *Router) registerPublicRoutes(app *fiber.App) error {
	// Create public route group
	publicGroup := app.Group("")

	// Apply public security configuration
	if r.deps.SecurityManager != nil {
		r.deps.SecurityManager.ConfigurePublicRoutes(publicGroup)
	}

	// Register health endpoints
	if err := r.registerHealth(publicGroup); err != nil {
		return fmt.Errorf("failed to register health module: %w", err)
	}

	r.deps.Logger.Info("registered public routes")
	return nil
}

// registerAPIRoutes registers API routes with full security
func (r *Router) registerAPIRoutes(app *fiber.App) error {
	// Create API route group
	apiGroup := app.Group("/api")

	// Apply API security configuration
	if r.deps.SecurityManager != nil {
		r.deps.SecurityManager.ConfigureAPIRoutes(apiGroup)
	}

	// Define API modules to register
	modules := []struct {
		name string
		fn   func(fiber.Router) error
	}{
		{ModuleAuth, r.registerAuthAPI},
		{ModuleTenant, r.registerTenantAPI},
		{ModuleUser, r.registerUserAPI},
		{"entity", r.registerEntityAPI},
		{ModuleFinance, r.registerFinanceAPI},
		{"schema", r.registerSchemaAPI},
		{"audit", r.registerAuditAPI},
	}

	// Register each API module
	for _, module := range modules {
		if err := module.fn(apiGroup); err != nil {
			return errors.NewBusinessError("API_MODULE_REGISTRATION_FAILED",
				fmt.Sprintf("failed to register %s API module", module.name)).
				WithCategory(errors.CategorySystem).
				WithSeverity(errors.SeverityCritical).
				WithDetail("module_name", module.name).
				WithDetail("error", err.Error()).
				WithSuggestion("Check API module configuration and dependencies")
		}

		r.deps.Logger.Info(fmt.Sprintf("registered %s API module", module.name))
	}

	return nil
}

// registerUIRoutes registers UI routes with session-based security
func (r *Router) registerUIRoutes(app *fiber.App) error {
	// Configure static file serving first
	app.Static("/static", "./web/static")

	// Serve web assets at root level so relative paths from /ui/demo resolve correctly
	// (e.g., ../sdk/sdk.css from /ui/demo resolves to /sdk/sdk.css)
	app.Static("/sdk", "./web/sdk")
	app.Static("/schemas", "./web/schemas")
	app.Static("/utils", "./web/utils")

	// Create UI route group
	uiGroup := app.Group("/ui")

	// Apply UI security configuration
	if r.deps.SecurityManager != nil {
		r.deps.SecurityManager.ConfigureUIRoutes(uiGroup)
	}

	// Create UI handler
	handler := uiHandler.NewUIHandler(r.deps.Logger, r.deps.Metrics, r.deps.Tracer)

	// Register UI demo routes
	uiGroup.Get("/demo", handler.ServeDemo)
	uiGroup.Get("/demo/components", handler.ServeComponents)
	uiGroup.Get("/demo/forms", handler.ServeForms)
	uiGroup.Get("/login", handler.ServeLogin)

	// Register schema, SDK, and utils routes
	handler.RegisterRoutes(uiGroup)

	// Redirect root UI to demo for now
	uiGroup.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/ui/demo")
	})

	r.deps.Logger.Info("registered UI routes with static assets")
	return nil
}

// registerHealth registers health check routes.
func (r *Router) registerHealth(router fiber.Router) error {
	// FIXME:pass real system Config here
	handler := health.NewHealthHandler(r.deps.Logger, r.deps.Metrics, r.deps.Tracer, nil)

	// Register health endpoints directly on the provided router
	healthGroup := router.Group("/health")
	healthGroup.Get("/", handler.Get)

	// Uncomment when implementing Kubernetes-style health checks:
	// healthGroup.Get("/ready", handler.Ready)   // Readiness probe
	// healthGroup.Get("/live", handler.Live)     // Liveness probe
	// healthGroup.Get("/startup", handler.Startup) // Startup probe

	r.deps.Logger.Info("registered health endpoints")
	return nil
}

// registerTenantAPI registers tenant management API routes.
func (r *Router) registerTenantAPI(apiRouter fiber.Router) error {
	// Use existing tenant service
	if r.deps.TenantService == nil {
		return errors.NewBusinessError("MISSING_TENANT_SERVICE", "Tenant service is required").
			WithCategory(errors.CategorySystem).
			WithSeverity(errors.SeverityCritical).
			WithSuggestion("Ensure tenant service is initialized before creating router")
	}

	handler := tenantHandler.NewTenantHandler(r.deps.TenantService, r.deps.Logger, r.deps.Metrics, r.deps.Tracer)

	// Tenant management routes — no tenant middleware (these manage tenants themselves)
	tenantsGroup := apiRouter.Group("/v1/tenants")

	tenantsGroup.Get("/", handler.List)         // GET /api/v1/tenants - List tenants with pagination
	tenantsGroup.Post("/", handler.Create)      // POST /api/v1/tenants - Create new tenant
	tenantsGroup.Get("/:id", handler.Get)       // GET /api/v1/tenants/:id - Get tenant by ID
	tenantsGroup.Put("/:id", handler.Update)    // PUT /api/v1/tenants/:id - Update tenant
	tenantsGroup.Patch("/:id", handler.Update)  // PATCH /api/v1/tenants/:id - Partial update
	tenantsGroup.Delete("/:id", handler.Delete) // DELETE /api/v1/tenants/:id - Delete tenant

	// Lifecycle actions
	tenantsGroup.Post("/:id/activate", handler.Activate) // POST /api/v1/tenants/:id/activate
	tenantsGroup.Post("/:id/suspend", handler.Suspend)   // POST /api/v1/tenants/:id/suspend
	tenantsGroup.Post("/:id/archive", handler.Archive)   // POST /api/v1/tenants/:id/archive

	// Complex operations
	apiRouter.Post("/v1/tenants/onboard", handler.Onboard) // POST /api/v1/tenants/onboard

	r.deps.Logger.Info("registered tenant API endpoints as /tenants")
	return nil
}

// registerUserAPI registers user management API routes.
func (r *Router) registerUserAPI(apiRouter fiber.Router) error {
	// Use existing user service
	if r.deps.UserService == nil {
		return errors.NewBusinessError("MISSING_USER_SERVICE", "User service is required").
			WithCategory(errors.CategorySystem).
			WithSeverity(errors.SeverityCritical).
			WithSuggestion("Ensure user service is initialized before creating router")
	}

	handler := userHandler.NewUserHandler(r.deps.UserService, r.deps.Logger, r.deps.Metrics, r.deps.Tracer)

	// Register user endpoints directly on the API router
	usersGroup := apiRouter.Group("/v1/users")

	// Apply tenant middleware for RLS if available
	if r.deps.TenantMiddleware != nil {
		usersGroup.Use(r.deps.TenantMiddleware)
	}

	usersGroup.Get("/", handler.List)                               // GET /api/v1/users - List users with pagination
	usersGroup.Post("/", handler.Create)                            // POST /api/v1/users - Create new user
	usersGroup.Get("/:id", handler.Get)                             // GET /api/v1/users/:id - Get user by ID
	usersGroup.Put("/:id", handler.Update)                          // PUT /api/v1/users/:id - Update user
	usersGroup.Delete("/:id", handler.Delete)                       // DELETE /api/v1/users/:id - Delete user
	usersGroup.Post("/authenticate", handler.Authenticate)          // POST /api/v1/users/authenticate - User authentication
	usersGroup.Post("/:id/change-password", handler.ChangePassword) // POST /api/v1/users/:id/change-password - Change password

	r.deps.Logger.Info("registered user API endpoints")
	return nil
}

// registerFinanceAPI registers finance management API routes.
func (r *Router) registerFinanceAPI(apiRouter fiber.Router) error {
	// Use existing finance services
	if r.deps.FinanceServices == nil {
		r.deps.Logger.Warn("Finance services not available, skipping finance API registration")
		return nil // Skip registration gracefully instead of failing
	}

	handler := financeHandler.NewFinanceHandler(r.deps.FinanceServices, r.deps.Logger, r.deps.Metrics, r.deps.Tracer)

	// Register finance endpoints directly on the API router
	financeGroup := apiRouter.Group("/v1/finance")

	// Apply Authenticate + tenant middleware for all finance routes (S10).
	// If SessionService is not configured the middleware is a no-op (dev mode).
	financeGroup.Use(r.authenticateMiddleware())
	if r.deps.TenantMiddleware != nil {
		financeGroup.Use(r.deps.TenantMiddleware)
	}

	// Account management endpoints — require finance.accounts.read / create / update etc.
	accountsGroup := financeGroup.Group("/accounts")
	accountsGroup.Use(middlewarePkg.Authorize("finance.accounts.read"))
	accountsGroup.Post("/", handler.CreateAccount)               // POST /api/v1/finance/accounts - Create account
	accountsGroup.Get("/", handler.ListAccounts)                 // GET /api/v1/finance/accounts - List accounts with filters
	accountsGroup.Get("/:id", handler.GetAccount)                // GET /api/v1/finance/accounts/:id - Get account by ID
	accountsGroup.Put("/:id", handler.UpdateAccount)             // PUT /api/v1/finance/accounts/:id - Update account
	accountsGroup.Delete("/:id", handler.DeleteAccount)          // DELETE /api/v1/finance/accounts/:id - Delete account
	accountsGroup.Get("/:id/balance", handler.GetAccountBalance) // GET /api/v1/finance/accounts/:id/balance - Get account balance

	// Transaction management endpoints — require finance.transactions.read
	transactionsGroup := financeGroup.Group("/transactions")
	transactionsGroup.Use(middlewarePkg.Authorize("finance.transactions.read"))
	transactionsGroup.Post("/", handler.CreateTransaction) // POST /api/v1/finance/transactions - Create transaction
	transactionsGroup.Get("/", handler.ListTransactions)   // GET /api/v1/finance/transactions - List transactions with filters
	transactionsGroup.Get("/:id", handler.GetTransaction)  // GET /api/v1/finance/transactions/:id - Get transaction by ID

	// Reporting endpoints
	reportsGroup := financeGroup.Group("/reports")
	reportsGroup.Get("/trial-balance", handler.GetTrialBalance) // GET /api/v1/finance/reports/trial-balance - Trial balance report

	r.deps.Logger.Info("registered finance API endpoints")
	return nil
}

// registerSchemaAPI registers the AMIS schema endpoints.
func (r *Router) registerSchemaAPI(apiRouter fiber.Router) error {
	if r.deps.Store == nil {
		r.deps.Logger.Warn("Store not configured, skipping schema route registration")
		return nil
	}

	schemaGroup := apiRouter.Group("/v1/schema")
	schemaGroup.Use(r.authenticateMiddleware())

	// GET /api/v1/schema/boot — AMIS app shell (nav filtered by flags + permissions)
	schemaGroup.Get("/boot", schemaHandler.BootHandler(r.deps.Store))

	r.deps.Logger.Info("registered schema API endpoints")
	return nil
}

// registerAuthAPI registers the login/logout endpoints.
// These are public (no Authenticate middleware) — the handlers themselves
// set the session cookie on success.
func (r *Router) registerAuthAPI(apiRouter fiber.Router) error {
	if r.deps.SessionService == nil {
		r.deps.Logger.Warn("SessionService not configured, skipping auth route registration")
		return nil
	}

	cookieName := "session"
	if r.deps.AuthConfig != nil {
		cookieName = r.deps.AuthConfig.CookieName
	}

	loginCfg := authHandler.DefaultLoginConfig()

	authGroup := apiRouter.Group("/v1/auth")
	authGroup.Post("/login", authHandler.LoginHandler(r.deps.SessionService, loginCfg))
	authGroup.Post("/logout", authHandler.LogoutHandler(r.deps.SessionService, cookieName))

	// Password reset — public endpoints (no session required)
	if r.deps.UserService != nil {
		authGroup.Post("/forgot-password", authHandler.ForgotPasswordHandler(r.deps.UserService))
		authGroup.Post("/reset-password", authHandler.ResetPasswordHandler(r.deps.UserService))
	}

	// MFA endpoints — /auth/mfa/complete is public; the others require an active session.
	mfaGroup := authGroup.Group("/mfa")
	mfaGroup.Post("/complete", authHandler.MFACompleteHandler(r.deps.SessionService, loginCfg))

	if r.deps.UserService != nil {
		mfaGroup.Post("/initiate", r.authenticateMiddleware(), authHandler.MFAInitiateHandler(r.deps.UserService))
		mfaGroup.Post("/confirm", r.authenticateMiddleware(), authHandler.MFAConfirmHandler(r.deps.UserService))
		mfaGroup.Delete("/", r.authenticateMiddleware(), authHandler.MFADisableHandler(r.deps.UserService))
	}

	// OAuth/OIDC SSO endpoints — public (no session required for begin or callback).
	if r.deps.SSOService != nil {
		oauthGroup := authGroup.Group("/oauth")
		oauthGroup.Get("/:provider", authHandler.OAuthBeginHandler(r.deps.SSOService))
		oauthGroup.Get("/:provider/callback", authHandler.OAuthCallbackHandler(r.deps.SSOService, r.deps.SessionService, loginCfg))
	}

	// API key management — requires active session (not an API key itself).
	if r.deps.APIKeyService != nil {
		apiKeysGroup := authGroup.Group("/api-keys")
		apiKeysGroup.Use(r.authenticateMiddleware())
		apiKeysGroup.Post("/", authHandler.CreateAPIKeyHandler(r.deps.APIKeyService))
		apiKeysGroup.Get("/", authHandler.ListAPIKeysHandler(r.deps.APIKeyService))
		apiKeysGroup.Delete("/:id", authHandler.RevokeAPIKeyHandler(r.deps.APIKeyService))
	}

	r.deps.Logger.Info("registered auth API endpoints")
	return nil
}

// registerEntityAPI registers entity management endpoints.
// Requires tenant middleware (RLS) but no session auth — same pattern as user creation.
func (r *Router) registerEntityAPI(apiRouter fiber.Router) error {
	if r.deps.Store == nil {
		r.deps.Logger.Warn("Store not configured, skipping entity route registration")
		return nil
	}

	repo := entity.NewRepository(r.deps.Store, r.deps.Tracer, r.deps.Metrics)
	svc := entity.NewService(repo, r.deps.Tracer, r.deps.Metrics)

	entitiesGroup := apiRouter.Group("/v1/entities")
	if r.deps.TenantMiddleware != nil {
		entitiesGroup.Use(r.deps.TenantMiddleware)
	}

	entitiesGroup.Post("/", func(c *fiber.Ctx) error {
		var req entity.CreateEntityRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}
		if req.Name == "" || req.Code == "" || req.Type == "" {
			return fiber.NewError(fiber.StatusUnprocessableEntity, "name, code, and type are required")
		}
		ent, err := svc.CreateEntity(c.UserContext(), req)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": ent})
	})

	entitiesGroup.Get("/", func(c *fiber.Ctx) error {
		entities, err := svc.ListEntities(c.UserContext(), entity.ListEntitiesRequest{Limit: 100})
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(fiber.Map{"data": entities})
	})

	r.deps.Logger.Info("registered entity API endpoints")
	return nil
}

// registerAuditAPI registers the audit log read endpoint.
func (r *Router) registerAuditAPI(apiRouter fiber.Router) error {
	if r.deps.AuditService == nil {
		r.deps.Logger.Warn("AuditService not configured, skipping audit route registration")
		return nil
	}

	auditGroup := apiRouter.Group("/v1/audit-logs")
	auditGroup.Use(r.authenticateMiddleware())
	auditGroup.Get("/", auditHandler.ListAuditEventsHandler(r.deps.AuditService))

	r.deps.Logger.Info("registered audit API endpoints")
	return nil
}

// authenticateMiddleware returns the Authenticate handler if session auth is configured,
// otherwise returns a no-op pass-through.
func (r *Router) authenticateMiddleware() fiber.Handler {
	if r.deps.SessionService == nil || r.deps.AuthConfig == nil {
		return func(c *fiber.Ctx) error { return c.Next() }
	}
	cfg := *r.deps.AuthConfig
	cfg.APIKeyService = r.deps.APIKeyService // inject if wired
	return middlewarePkg.Authenticate(cfg)
}

// Future module registration examples:
//
// func (r *Router) registerFinance(app *fiber.App) error {
// 	handler := finance.NewHandler(r.deps.Logger, r.deps.Metrics, r.deps.Tracer)
//
// 	return r.registry.RegisterModuleWithMiddleware(
// 		app,
// 		ModuleFinance,
// 		"/api/v1/finance",
// 		[]string{"cors", "auth", "ratelimit"},
// 		func(router fiber.Router) {
// 			router.Get("/accounts", handler.ListAccounts)
// 			router.Post("/transactions", handler.CreateTransaction)
// 			router.Get("/reports", handler.GetReports)
// 		},
// 	)
// }

// ListRoutes returns information about all registered routes.
func (r *Router) ListRoutes() map[string]*ModuleInfo {
	return r.registry.ListRoutes()
}

// PrintRoutes logs all registered routes (useful for debugging and startup).
func (r *Router) PrintRoutes() {
	routes := r.registry.ListRoutes()

	r.deps.Logger.Info("registered routes:")
	for name, info := range routes {
		r.deps.Logger.Info(fmt.Sprintf("  - %s: %v", name, info))
	}
}

// GetModuleInfo returns information about a specific module.
func (r *Router) GetModuleInfo(moduleName string) (*ModuleInfo, error) {
	routes := r.registry.ListRoutes()

	info, exists := routes[moduleName]
	if !exists {
		return nil, errors.NewBusinessError("MODULE_NOT_FOUND",
			fmt.Sprintf("module %s not found", moduleName)).
			WithCategory(errors.CategoryBusiness).
			WithDetail("module_name", moduleName).
			WithSuggestion("Check if the module has been registered")
	}

	return info, nil
}

// HealthCheck performs a health check on all registered modules.
// This can be extended to check database connections, external services, etc.
func (r *Router) HealthCheck() error {
	// Verify dependencies are still valid
	if err := r.deps.Validate(); err != nil {
		return errors.NewBusinessError("HEALTH_CHECK_FAILED", "dependency validation failed").
			WithCategory(errors.CategorySystem).
			WithSeverity(errors.SeverityError).
			WithDetail("error", err.Error())
	}

	// Check if routes are registered
	routes := r.registry.ListRoutes()
	if len(routes) == 0 {
		return errors.NewBusinessError("HEALTH_CHECK_FAILED", "no routes registered").
			WithCategory(errors.CategorySystem).
			WithSeverity(errors.SeverityWarning).
			WithSuggestion("Ensure RegisterAll was called successfully")
	}

	return nil
}
