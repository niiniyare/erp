// Package handlers wires every HTTP route for the AWO ERP API.
//
// # Structure
//
// The file is organised in five clearly separated sections:
//
//  1. Constants & module names  — one place to look up every module key.
//  2. RouteRegistry             — tracks what has been registered, prevents duplicates.
//  3. Dependencies              — all services/middleware a Router needs, with validation.
//  4. Router                   — owns RegisterAll(); calls one private method per module.
//  5. Module registration funcs — each `register<Module>` is self-contained and follows
//     a uniform template (guard → handler → group → middleware → routes).
//
// # Adding a new module
//
// Follow these five steps — nothing else needs changing:
//
//  1. Add a constant in the "Module names" block below.
//  2. Add the service field to Dependencies and a nil-check to Validate() if required.
//  3. Write a `register<Module>API(apiRouter fiber.Router) error` method.
//  4. Append a `{ModuleXxx, r.registerXxxAPI}` entry to the modules slice in registerAPIRoutes.
//  5. Run `make wire` to re-generate the provider graph if you added a Wire provider.
package handlers

import (
	"fmt"
	"sync"

	"github.com/gofiber/fiber/v2"

	auditHandler "awo.so/internal/api/handlers/audit"
	authHandler "awo.so/internal/api/handlers/auth"
	contractHandler "awo.so/internal/api/handlers/contracts"
	entityHandler "awo.so/internal/api/handlers/entity"
	financeHandler "awo.so/internal/api/handlers/finance"
	"awo.so/internal/api/handlers/health"
	iamHandler "awo.so/internal/api/handlers/iam"
	schemaHandler "awo.so/internal/api/handlers/schema"
	webHandler "awo.so/internal/web/handler"

	db "awo.so/db/sqlc"
	tenantHandler "awo.so/internal/api/handlers/tenant"
	uiHandler "awo.so/internal/api/handlers/ui"
	userHandler "awo.so/internal/api/handlers/user"
	middlewarePkg "awo.so/internal/api/middleware"
	"awo.so/internal/core/audit"
	"awo.so/internal/core/contracts"
	"awo.so/internal/core/entity"
	financeService "awo.so/internal/core/finance/service"
	"awo.so/internal/core/iam/contract"
	coreTenant "awo.so/internal/core/tenant"
	"awo.so/internal/shared/errors"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"

	// Blank imports register page schemas into the global registry via init().
	// ── Page schemas (AMIS JSON-driven pages) ────────────────────────────────
	_ "awo.so/internal/web/pages/dashboard"
	_ "awo.so/internal/web/pages/finance/accounts"
	_ "awo.so/internal/web/pages/finance/transactions"
	_ "awo.so/internal/web/pages/organizations"
	_ "awo.so/internal/web/pages/settings"
	_ "awo.so/internal/web/pages/users"

	// ── DSL screens (typed AST path) ─────────────────────────────────────────
	_ "awo.so/internal/web/dsl/screens"

	workflowsTenant "awo.so/internal/workflows/tenant"
	temporalclient "go.temporal.io/sdk/client"
)

// =============================================================================
// SECTION 1 — Module name constants
//
// Use these constants everywhere a module name appears as a string.
// Never hardcode "finance", "auth", etc. directly — it prevents grep-based
// refactoring and makes typos invisible.
// =============================================================================

const (
	ModuleHealth    = "health"
	ModuleAuth      = "auth"
	ModuleTenant    = "tenant"
	ModuleUser      = "user"
	ModuleEntity    = "entity"
	ModuleFinance   = "finance"
	ModuleContracts = "contracts"
	ModuleSchema    = "schema"
	ModuleAudit     = "audit"
	ModuleIAM       = "iam"

	// apiV1Prefix is the common versioned prefix for all REST API routes.
	// Change this once to move all API routes to /api/v2.
	apiV1Prefix = "/v1"
)

// =============================================================================
// SECTION 2 — RouteRegistry
//
// Lightweight registry that tracks which modules and base paths are live.
// It exists so tooling (health checks, debug endpoints, /routes introspection)
// can enumerate registered modules at runtime without scanning the Fiber router.
//
// NOTE: The registry is only written to via trackModule() after a successful
// register<X> call.  It is intentionally not the path through which routes are
// actually created — that is Fiber's job.
// =============================================================================

// ModuleInfo is the read-only view of a registered module that external tooling
// may inspect (e.g. the /health endpoint or an /admin/routes debug view).
type ModuleInfo struct {
	Name       string         `json:"name"`
	BasePath   string         `json:"base_path"`
	Middleware []string       `json:"middleware"`
	RouteCount int            `json:"route_count"`
	Registered bool           `json:"registered"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

// RouteRegistry is the in-memory module catalogue.
// All writes are serialised via mu; reads take an RLock.
type RouteRegistry struct {
	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.Service

	mu                sync.RWMutex
	registeredModules map[string]*ModuleInfo // keyed by module name constant
	registeredPaths   map[string]bool        // keyed by base path string
}

// newRouteRegistry constructs an empty registry.
func newRouteRegistry(log logger.Logger, m metrics.MetricsProvider, t tracing.Service) *RouteRegistry {
	return &RouteRegistry{
		logger:            log,
		metrics:           m,
		tracer:            t,
		registeredModules: make(map[string]*ModuleInfo),
		registeredPaths:   make(map[string]bool),
	}
}

// track records a successfully registered module.
// If the module was already recorded (e.g. RegisterAll called twice in tests),
// it logs a warning and overwrites rather than panicking.
func (r *RouteRegistry) track(name, basePath string, routeCount int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.registeredModules[name]; exists {
		r.logger.Warn(fmt.Sprintf("route registry: module %q re-registered — overwriting previous entry", name))
	}

	r.registeredModules[name] = &ModuleInfo{
		Name:       name,
		BasePath:   basePath,
		RouteCount: routeCount,
		Registered: true,
		Metadata:   make(map[string]any),
	}
	r.registeredPaths[basePath] = true

	r.metrics.IncrementCounter("modules_registered_total", metrics.Fields{
		"module_name": name,
	})
}

// ListModules returns a deep copy of all registered module infos.
// Callers must not mutate the returned values.
func (r *RouteRegistry) ListModules() map[string]*ModuleInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make(map[string]*ModuleInfo, len(r.registeredModules))
	for name, info := range r.registeredModules {
		cp := *info
		cp.Middleware = make([]string, len(info.Middleware))
		copy(cp.Middleware, info.Middleware)
		cp.Metadata = make(map[string]any, len(info.Metadata))
		for k, v := range info.Metadata {
			cp.Metadata[k] = v
		}
		out[name] = &cp
	}
	return out
}

// =============================================================================
// SECTION 3 — Dependencies
//
// Every service, middleware, and client that any registered module needs is
// declared here.  Fields that are *required* have a nil-check in Validate();
// fields that are *optional* are guarded inside their register<X> function with
// a logger.Warn and an early return nil.
//
// Adding a dependency for a new module:
//   - Add the field below with a clear docstring.
//   - If required, add a nil check in Validate().
//   - If optional, add a guard in the register<X> function (see pattern at the
//     top of registerAuditAPI for an example).
// =============================================================================

// Dependencies holds every dependency the Router needs to register all modules.
//
// Required fields (Validate returns an error if nil):
//   - Logger, Metrics, Tracer — observability stack.
//   - TenantService           — used by registerTenantAPI; not gracefully skipped.
//
// Optional fields (module is silently skipped when nil, a Warn is emitted):
//   - All other service fields.
type Dependencies struct {
	// ── Observability (required) ─────────────────────────────────────────────
	Logger  logger.Logger
	Metrics metrics.MetricsProvider
	Tracer  tracing.Service

	// ── Core services ────────────────────────────────────────────────────────

	// TenantService manages tenant lifecycle. Required — no graceful skip.
	TenantService coreTenant.Service

	// UserService manages users within a tenant. Optional.
	UserService contract.UserService

	// FinanceServices is the bundle of finance sub-services. Optional.
	FinanceServices *financeService.Services

	// AuditService enables the GET /api/v1/audit-logs endpoint. Optional.
	AuditService audit.Service

	// ContractService enables contract lifecycle endpoints. Optional.
	ContractService contracts.Service

	// ── Auth / session ───────────────────────────────────────────────────────

	// SessionService enables /auth/login, /auth/logout, and session validation
	// via the Authenticate middleware. Optional — auth routes are skipped and
	// Authenticate becomes a no-op pass-through when nil.
	SessionService contract.SessionService

	// AuthConfig carries the cookie name, AuthzService, and related settings.
	// Optional — auth and authz middleware degrade to no-ops when nil.
	//
	// NOTE: Set AuthConfig.APIKeyService = APIKeyService before passing in so
	// that Authenticate can validate Bearer tokens.  Do not rely on the router
	// to merge these; the explicit assignment makes the dependency graph clear.
	AuthConfig *middlewarePkg.AuthConfig

	// SSOService enables OAuth/OIDC login routes. Optional.
	SSOService contract.SSOService

	// APIKeyService enables API key management routes and Bearer token
	// validation. Optional.
	APIKeyService contract.APIKeyService

	// IAMService enables IAM management routes (policies, roles, assignments).
	// Optional — IAM management routes are skipped when nil.
	IAMService contract.AuthzService

	// ── Infrastructure ───────────────────────────────────────────────────────

	// Store is the raw DB store consumed by schema and entity endpoints.
	// Optional — those modules are skipped when nil.
	Store db.Store

	// TemporalClient enables async workflow routes (e.g. tenant onboarding).
	// Optional — workflow-backed routes return 503 when nil.
	TemporalClient temporalclient.Client

	// ── Middleware ───────────────────────────────────────────────────────────

	// TenantMiddleware extracts X-Tenant-ID from the request, sets the RLS
	// context, and must be applied to every tenant-scoped route group.
	TenantMiddleware fiber.Handler

	// SecurityManager configures CORS, CSRF, and security headers per route
	// class (public / API / UI).  Optional — a permissive dev CORS policy is
	// applied when nil.
	SecurityManager *middlewarePkg.RouteSecurityManager

}

// Validate returns an error if any required dependency is missing.
// Call this before constructing a Router.
func (d *Dependencies) Validate() error {
	if d == nil {
		return errors.NewBusinessError("INVALID_DEPENDENCIES", "dependencies cannot be nil").
			WithCategory(errors.CategorySystem).
			WithSeverity(errors.SeverityCritical)
	}
	if d.Logger == nil {
		return errors.NewBusinessError("MISSING_LOGGER", "Logger is required").
			WithCategory(errors.CategorySystem).WithSeverity(errors.SeverityCritical)
	}
	if d.Metrics == nil {
		return errors.NewBusinessError("MISSING_METRICS", "Metrics is required").
			WithCategory(errors.CategorySystem).WithSeverity(errors.SeverityCritical)
	}
	if d.Tracer == nil {
		return errors.NewBusinessError("MISSING_TRACER", "Tracer is required").
			WithCategory(errors.CategorySystem).WithSeverity(errors.SeverityCritical)
	}
	if d.TenantService == nil {
		return errors.NewBusinessError("MISSING_TENANT_SERVICE", "TenantService is required").
			WithCategory(errors.CategorySystem).WithSeverity(errors.SeverityCritical).
			WithSuggestion("Initialise TenantService via Wire before calling NewRouter")
	}
	return nil
}

// =============================================================================
// SECTION 4 — Router
//
// Router is the single entry-point for route registration.
// Call RegisterAll(app) once from main (or the Wire-generated provider) after
// the Fiber app is created.
// =============================================================================

// Router owns the Fiber app's route topology.
type Router struct {
	registry *RouteRegistry
	deps     *Dependencies
}

// NewRouter validates dependencies and returns a ready-to-use Router.
func NewRouter(deps *Dependencies) (*Router, error) {
	if err := deps.Validate(); err != nil {
		return nil, err
	}
	return &Router{
		registry: newRouteRegistry(deps.Logger, deps.Metrics, deps.Tracer),
		deps:     deps,
	}, nil
}

// RegisterAll registers every route in the correct order:
//
//  1. Public routes   — health, SSO callbacks, auth login/logout.
//  2. API routes      — /api/v1/... with session auth + tenant RLS.
//  3. UI routes       — /ui/... static assets + AMIS schema endpoints.
//
// Returns the first error encountered; a partial registration is a startup
// failure and the caller should treat it as fatal.
func (r *Router) RegisterAll(app *fiber.App) error {
	if err := r.registerPublicRoutes(app); err != nil {
		return fmt.Errorf("public routes: %w", err)
	}
	if err := r.registerAPIRoutes(app); err != nil {
		return fmt.Errorf("api routes: %w", err)
	}
	if err := r.registerUIRoutes(app); err != nil {
		return fmt.Errorf("ui routes: %w", err)
	}
	r.deps.Logger.Info("all routes registered")
	return nil
}

// ListModules returns a snapshot of every registered module.
// Useful for /admin/routes or startup logs.
func (r *Router) ListModules() map[string]*ModuleInfo {
	return r.registry.ListModules()
}

// PrintModules logs every registered module at Info level.
func (r *Router) PrintModules() {
	for name, info := range r.registry.ListModules() {
		r.deps.Logger.Info(fmt.Sprintf("  module %-16s  path=%s  routes=%d", name, info.BasePath, info.RouteCount))
	}
}

// GetModuleInfo returns info for a single module, or an error if not registered.
func (r *Router) GetModuleInfo(name string) (*ModuleInfo, error) {
	modules := r.registry.ListModules()
	info, ok := modules[name]
	if !ok {
		return nil, errors.NewBusinessError("MODULE_NOT_FOUND",
			fmt.Sprintf("module %q is not registered", name)).
			WithCategory(errors.CategoryBusiness).
			WithDetail("module_name", name)
	}
	return info, nil
}

// HealthCheck verifies that dependencies are still valid and routes were registered.
func (r *Router) HealthCheck() error {
	if err := r.deps.Validate(); err != nil {
		return errors.NewBusinessError("HEALTH_CHECK_FAILED", "dependency validation failed").
			WithCategory(errors.CategorySystem).WithSeverity(errors.SeverityError).
			WithDetail("error", err.Error())
	}
	if len(r.registry.ListModules()) == 0 {
		return errors.NewBusinessError("HEALTH_CHECK_FAILED", "no modules registered — was RegisterAll called?").
			WithCategory(errors.CategorySystem).WithSeverity(errors.SeverityWarning)
	}
	return nil
}

// =============================================================================
// SECTION 5 — Route group setup (public / api / ui)
// =============================================================================

// registerPublicRoutes mounts routes that require no authentication:
//   - GET /health/...
//
// CORS and security headers are applied by SecurityManager if present;
// otherwise a permissive dev policy is used so local tooling is not blocked.
func (r *Router) registerPublicRoutes(app *fiber.App) error {
	publicGroup := app.Group("")

	if r.deps.SecurityManager != nil {
		r.deps.SecurityManager.ConfigurePublicRoutes(publicGroup)
	} else {
		// Dev/test fallback — do not use in production.
		corsConfig := middlewarePkg.DevelopmentCORSConfig([]int{3000, 8080})
		publicGroup.Use(middlewarePkg.NewCORSMiddleware(corsConfig))
	}

	if err := r.registerHealth(publicGroup); err != nil {
		return fmt.Errorf("health: %w", err)
	}

	r.deps.Logger.Info("public routes registered")
	return nil
}

// registerAPIRoutes mounts every /api/v1/... module.
//
// To add a new module: append one entry to the modules slice — that's it.
// The loop handles error wrapping, tracking, and logging uniformly.
func (r *Router) registerAPIRoutes(app *fiber.App) error {
	apiGroup := app.Group("/api")

	if r.deps.SecurityManager != nil {
		r.deps.SecurityManager.ConfigureAPIRoutes(apiGroup)
	}

	// ── Module registration table ────────────────────────────────────────────
	// Each entry is { module-name-constant, register-function }.
	// Order matters only when modules share a prefix (none currently do).
	// ────────────────────────────────────────────────────────────────────────
	type apiModule struct {
		name string
		fn   func(fiber.Router) error
	}
	modules := []apiModule{
		{ModuleAuth, r.registerAuthAPI},
		{ModuleTenant, r.registerTenantAPI},
		{ModuleUser, r.registerUserAPI},
		{ModuleEntity, r.registerEntityAPI},
		{ModuleFinance, r.registerFinanceAPI},
		{ModuleContracts, r.registerContractsAPI},
		{ModuleSchema, r.registerSchemaAPI},
		{ModuleAudit, r.registerAuditAPI},
		{ModuleIAM, r.registerIAMAPI},
		// ↑ Add new modules here — one line per module.
	}

	for _, mod := range modules {
		if err := mod.fn(apiGroup); err != nil {
			// Wrap so the caller sees which module failed.
			return fmt.Errorf("module %q registration failed: %w", mod.name, err)
		}
		r.deps.Logger.Info(fmt.Sprintf("api module registered: %s", mod.name))
	}

	r.deps.Logger.Info("api routes registered")
	return nil
}

// registerUIRoutes mounts:
//   - /schema/* — AMIS page schema JSON (requires auth session).
//   - /static, /sdk, /schemas, /utils — static file serving.
//   - /ui/* — demo pages, login page, component gallery.
func (r *Router) registerUIRoutes(app *fiber.App) error {
	// ── AMIS schema endpoint ─────────────────────────────────────────────────
	// /schema/<route> returns the AMIS page definition JSON for that route.
	// Authenticated but intentionally outside /api so the frontend URL stays clean.
	pageSchemaHandler := webHandler.NewDevSchemaHandler(r.registry.logger)
	schemaGroup := app.Group("/schema")
	schemaGroup.Use(r.authenticateMiddleware())
	schemaGroup.Get("/*", pageSchemaHandler.Handle)

	// ── Static assets ────────────────────────────────────────────────────────
	// Served at root-relative paths so relative imports (e.g. ../sdk/sdk.css
	// from /ui/demo) resolve correctly without a base-href rewrite.
	app.Static("/static", "./web/static")
	app.Static("/sdk", "./web/sdk")
	app.Static("/schemas", "./web/schemas")
	app.Static("/utils", "./web/utils")

	// ── UI pages ─────────────────────────────────────────────────────────────
	uiGroup := app.Group("/ui")
	if r.deps.SecurityManager != nil {
		r.deps.SecurityManager.ConfigureUIRoutes(uiGroup)
	}

	handler := uiHandler.NewUIHandler(r.deps.Logger, r.deps.Metrics, r.deps.Tracer)

	// Public UI routes — no auth required.
	uiGroup.Get("/login", handler.ServeLogin)

	// Protected UI routes — Authenticate redirects unauthenticated browsers to
	// /ui/login?redirect=<original-path> (text/html Accept detection in middleware).
	// When auth is not configured (dev/test) authenticateMiddleware is a no-op.
	protected := uiGroup.Group("", r.authenticateMiddleware())
	protected.Get("/demo", handler.ServeDemo)
	protected.Get("/demo/components", handler.ServeComponents)
	protected.Get("/demo/forms", handler.ServeForms)
	protected.Get("/", func(c *fiber.Ctx) error { return c.Redirect("/ui/demo") })

	handler.RegisterRoutes(uiGroup)

	r.deps.Logger.Info("ui routes registered")
	return nil
}

// =============================================================================
// SECTION 6 — Individual module registration functions
//
// Naming convention: register<Module>[API]
//
// Each function follows this template:
//
//  1. Guard   — if required service is nil, return error OR log Warn and return nil.
//  2. Handler — construct the domain handler.
//  3. Group   — create the Fiber route group at the canonical path.
//  4. Middleware — apply authenticate / tenant / authorize in that order.
//  5. Routes  — register every endpoint with a comment showing the full URL.
//  6. Track   — call r.registry.track(...) with the accurate route count.
//
// "Required" modules return an error when their primary service is nil.
// "Optional" modules log a Warn and return nil; the app starts without them.
// =============================================================================

// registerHealth mounts the liveness / readiness probes under /health.
// These are public — no auth, no tenant context.
func (r *Router) registerHealth(router fiber.Router) error {
	handler := health.NewHealthHandler(r.deps.Logger, r.deps.Metrics, r.deps.Tracer, nil)

	g := router.Group("/health")
	g.Get("/", handler.Get)            // GET /health/
	g.Get("/ready", handler.Ready)     // GET /health/ready
	g.Get("/live", handler.Live)       // GET /health/live
	g.Get("/startup", handler.Startup) // GET /health/startup

	r.registry.track(ModuleHealth, "/health", 4)
	return nil
}

// registerAuthAPI mounts login/logout, MFA, OAuth, and API-key management.
//
// Auth routes are intentionally public (no Authenticate middleware on the group
// itself) because the handlers create sessions on success.  Per-route
// middleware is applied to the endpoints that require an active session (MFA
// initiate/confirm/disable, API key management).
//
// Optional — skipped when SessionService is nil.
func (r *Router) registerAuthAPI(apiRouter fiber.Router) error {
	if r.deps.SessionService == nil {
		r.deps.Logger.Warn("SessionService not configured — auth routes skipped") // optional
		return nil
	}

	cookieName := "session"
	if r.deps.AuthConfig != nil {
		cookieName = r.deps.AuthConfig.CookieName
	}
	loginCfg := authHandler.DefaultLoginConfig()

	g := apiRouter.Group(apiV1Prefix + "/auth")

	// Apply tenant middleware so the session can be scoped to the correct tenant
	// (login request carries X-Tenant-ID).
	if r.deps.TenantMiddleware != nil {
		g.Use(r.deps.TenantMiddleware)
	}

	// ── Session ───────────────────────────────────────────────────────────
	g.Post("/login", authHandler.LoginHandler(r.deps.SessionService, loginCfg))
	g.Post("/logout", authHandler.LogoutHandler(r.deps.SessionService, cookieName))

	// ── Password reset (public — no session required) ─────────────────────
	if r.deps.UserService != nil {
		g.Post("/forgot-password", authHandler.ForgotPasswordHandler(r.deps.UserService))
		g.Post("/reset-password", authHandler.ResetPasswordHandler(r.deps.UserService))
	}

	// ── MFA ───────────────────────────────────────────────────────────────
	// /mfa/complete is public (verifies OTP, upgrades pending session).
	// The remaining endpoints require a fully authenticated session.
	mfa := g.Group("/mfa")
	mfa.Post("/complete", authHandler.MFACompleteHandler(r.deps.SessionService, loginCfg))

	if r.deps.UserService != nil {
		auth := r.authenticateMiddleware()
		mfa.Post("/initiate", auth, authHandler.MFAInitiateHandler(r.deps.UserService))
		mfa.Post("/confirm", auth, authHandler.MFAConfirmHandler(r.deps.UserService))
		mfa.Delete("/", auth, authHandler.MFADisableHandler(r.deps.UserService))
	}

	// ── OAuth / OIDC SSO (public — browser redirect flow) ─────────────────
	if r.deps.SSOService != nil {
		oauth := g.Group("/oauth")
		oauth.Get("/:provider", authHandler.OAuthBeginHandler(r.deps.SSOService))
		oauth.Get("/:provider/callback",
			authHandler.OAuthCallbackHandler(r.deps.SSOService, r.deps.SessionService, loginCfg))
	}

	// ── API key management (requires active session) ───────────────────────
	if r.deps.APIKeyService != nil {
		keys := g.Group("/api-keys")
		keys.Use(r.authenticateMiddleware())
		keys.Post("/", authHandler.CreateAPIKeyHandler(r.deps.APIKeyService))
		keys.Get("/", authHandler.ListAPIKeysHandler(r.deps.APIKeyService))
		keys.Delete("/:id", authHandler.RevokeAPIKeyHandler(r.deps.APIKeyService))
	}

	r.registry.track(ModuleAuth, apiV1Prefix+"/auth", 10)
	r.deps.Logger.Info("auth API endpoints registered")
	return nil
}

// registerTenantAPI mounts tenant lifecycle CRUD and async onboarding.
//
// Security model — platform-only:
//   - No TenantMiddleware: these routes manage tenants, they do not run inside
//     a tenant context (no RLS injection, no X-Tenant-ID required).
//   - authenticateMiddleware: caller must hold a valid session or API key.
//   - platformOnlyMiddleware: caller's actor type must be ActorPlatform.
//     Regular user sessions and tenant-scoped API keys are rejected with 403.
//     In dev/test mode (PlatformActorService == nil) the check is a no-op.
//
// Required — returns an error if TenantService is nil.
func (r *Router) registerTenantAPI(apiRouter fiber.Router) error {
	// TenantService is required; Validate() catches nil at startup, but we guard
	// here as well for clarity.
	if r.deps.TenantService == nil {
		return errors.NewBusinessError("MISSING_TENANT_SERVICE", "TenantService is required for tenant routes").
			WithCategory(errors.CategorySystem).WithSeverity(errors.SeverityCritical)
	}

	// Wire the Temporal onboarding starter only when a client is available.
	var onboardStarter *workflowsTenant.Starter
	if r.deps.TemporalClient != nil {
		onboardStarter = workflowsTenant.NewStarter(r.deps.TemporalClient)
	}

	handler := tenantHandler.NewTenantHandler(
		r.deps.TenantService, r.deps.Logger, r.deps.Metrics, r.deps.Tracer, onboardStarter,
	)
	// Inject user + authz + entity services for synchronous onboarding when available.
	if r.deps.UserService != nil && r.deps.IAMService != nil && r.deps.Store != nil {
		entityRepo := entity.NewRepository(r.deps.Store, r.deps.Tracer, r.deps.Metrics)
		entitySvc := entity.NewService(entityRepo, r.deps.Tracer, r.deps.Metrics)
		handler.WithOnboardServices(r.deps.UserService, r.deps.IAMService, entitySvc)
	}

	g := apiRouter.Group(apiV1Prefix + "/tenants")

	// Apply authentication then platform-actor enforcement to the whole group.
	// Order matters: authenticate first (establishes actor identity), then
	// platformOnly (asserts actor type == ActorPlatform).
	g.Use(r.authenticateMiddleware(), r.platformOnlyMiddleware())

	// ── CRUD ──────────────────────────────────────────────────────────────
	g.Post("/", handler.Create)      // POST   /api/v1/tenants
	g.Get("/:id", handler.Get)       // GET    /api/v1/tenants/:id
	g.Put("/:id", handler.Update)    // PUT    /api/v1/tenants/:id
	g.Patch("/:id", handler.Update)  // PATCH  /api/v1/tenants/:id  (partial)
	g.Delete("/:id", handler.Delete) // DELETE /api/v1/tenants/:id

	// ── Lifecycle transitions ─────────────────────────────────────────────
	g.Post("/:id/activate", handler.Activate) // POST /api/v1/tenants/:id/activate
	g.Post("/:id/suspend", handler.Suspend)   // POST /api/v1/tenants/:id/suspend
	g.Post("/:id/archive", handler.Archive)   // POST /api/v1/tenants/:id/archive

	// ── Onboarding (public — no auth/platform guard) ─────────────────────
	// Registered on a separate prefix (/api/v1/onboard/*) so they are NOT
	// caught by the /api/v1/tenants prefix-based authenticate middleware above.
	// Async: enqueues a Temporal workflow (503 when Temporal not wired).
	apiRouter.Post(apiV1Prefix+"/onboard", handler.Onboard) // POST /api/v1/onboard
	// Sync: provisions tenant + admin user in one request; no Temporal needed.
	apiRouter.Post(apiV1Prefix+"/onboard/sync", handler.OnboardSync) // POST /api/v1/onboard/sync

	r.registry.track(ModuleTenant, apiV1Prefix+"/tenants", 12)
	r.deps.Logger.Info("tenant API endpoints registered")
	return nil
}

// registerUserAPI mounts user CRUD, authentication, and password management.
//
// All user routes run inside tenant context (TenantMiddleware) so that RLS
// correctly isolates users by tenant.
//
// Optional — skipped when UserService is nil.
func (r *Router) registerUserAPI(apiRouter fiber.Router) error {
	if r.deps.UserService == nil {
		r.deps.Logger.Warn("UserService not configured — user routes skipped") // optional
		return nil
	}

	handler := userHandler.NewUserHandler(r.deps.UserService, r.deps.Logger, r.deps.Metrics, r.deps.Tracer)

	g := apiRouter.Group(apiV1Prefix + "/users")

	g.Use(r.authenticateMiddleware())
	if r.deps.TenantMiddleware != nil {
		g.Use(r.deps.TenantMiddleware)
	}

	g.Get("/", handler.List)                               // GET    /api/v1/users
	g.Post("/", handler.Create)                            // POST   /api/v1/users
	g.Get("/:id", handler.Get)                             // GET    /api/v1/users/:id
	g.Put("/:id", handler.Update)                          // PUT    /api/v1/users/:id
	g.Delete("/:id", handler.Delete)                       // DELETE /api/v1/users/:id
	g.Post("/authenticate", handler.Authenticate)          // POST   /api/v1/users/authenticate
	g.Post("/:id/change-password", handler.ChangePassword) // POST   /api/v1/users/:id/change-password

	r.registry.track(ModuleUser, apiV1Prefix+"/users", 7)
	r.deps.Logger.Info("user API endpoints registered")
	return nil
}

// registerEntityAPI mounts entity CRUD and hierarchy endpoints.
//
// Security model:
//   - authenticate: caller must hold a valid session or API key.
//   - TenantMiddleware: RLS context injected for every request.
//
// Optional — skipped when Store is nil.
func (r *Router) registerEntityAPI(apiRouter fiber.Router) error {
	if r.deps.Store == nil {
		r.deps.Logger.Warn("Store not configured — entity routes skipped") // optional
		return nil
	}

	repo := entity.NewRepository(r.deps.Store, r.deps.Tracer, r.deps.Metrics)
	svc := entity.NewService(repo, r.deps.Tracer, r.deps.Metrics)
	handler := entityHandler.NewEntityHandler(svc, r.deps.Logger, r.deps.Metrics, r.deps.Tracer)

	g := apiRouter.Group(apiV1Prefix + "/entities")
	g.Use(r.authenticateMiddleware())
	if r.deps.TenantMiddleware != nil {
		g.Use(r.deps.TenantMiddleware)
	}

	// Tree must be registered before /:id to avoid Fiber matching "tree" as an ID.
	g.Get("/tree", handler.GetTree)               // GET    /api/v1/entities/tree
	g.Post("/", handler.Create)                    // POST   /api/v1/entities
	g.Get("/", handler.List)                       // GET    /api/v1/entities
	g.Get("/:id", handler.GetByID)                 // GET    /api/v1/entities/:id
	g.Put("/:id", handler.Update)                  // PUT    /api/v1/entities/:id
	g.Delete("/:id", handler.Delete)               // DELETE /api/v1/entities/:id
	g.Get("/:id/children", handler.GetChildren)    // GET    /api/v1/entities/:id/children

	r.registry.track(ModuleEntity, apiV1Prefix+"/entities", 7)
	r.deps.Logger.Info("entity API endpoints registered")
	return nil
}

// registerFinanceAPI mounts the full finance domain:
// accounts, transactions, fiscal years/periods, currencies, exchange rates,
// budgets, cost centres, tax authorities/codes, and bank reconciliation.
//
// Authorization model:
//   - authenticate middleware guards the entire /finance group.
//   - Each sub-resource group has a per-verb authorizeMiddleware call so that
//     read-only users cannot mutate data.  See the table below.
//
// Permission table (sub-resource → minimum permission per HTTP verb):
//
//	accounts       read=finance.accounts.read     write=finance.accounts.write
//	transactions   read=finance.transactions.read write=finance.transactions.write
//	fiscal-years   read=finance.periods.read      write=finance.periods.write
//	periods        read=finance.periods.read      write=finance.periods.write
//	currencies     read=finance.currencies.read   write=finance.currencies.write
//	exchange-rates read=finance.currencies.read   write=finance.currencies.write
//	budgets        read=finance.budgets.read       write=finance.budgets.write
//	cost-centers   read=finance.cost_centers.read write=finance.cost_centers.write
//	tax            read=finance.tax.read           write=finance.tax.write
//	reconciliation read=finance.reconciliation.read write=finance.reconciliation.write
//
// Optional — skipped when FinanceServices is nil.
func (r *Router) registerFinanceAPI(apiRouter fiber.Router) error {
	if r.deps.FinanceServices == nil {
		r.deps.Logger.Warn("FinanceServices not configured — finance routes skipped") // optional
		return nil
	}

	handler := financeHandler.NewFinanceHandler(
		r.deps.FinanceServices, r.deps.Logger, r.deps.Metrics, r.deps.Tracer,
	)

	g := apiRouter.Group(apiV1Prefix + "/finance")

	// Every finance endpoint requires an authenticated session + tenant context.
	g.Use(r.authenticateMiddleware())
	if r.deps.TenantMiddleware != nil {
		g.Use(r.deps.TenantMiddleware)
	}

	// ── Accounts ──────────────────────────────────────────────────────────
	accounts := g.Group("/accounts")
	accounts.Get("/", r.authorizeMiddleware("finance.accounts.read"), handler.ListAccounts)
	accounts.Get("/:id", r.authorizeMiddleware("finance.accounts.read"), handler.GetAccount)
	accounts.Get("/:id/balance", r.authorizeMiddleware("finance.accounts.read"), handler.GetAccountBalance)
	accounts.Post("/", r.authorizeMiddleware("finance.accounts.write"), handler.CreateAccount)
	accounts.Put("/:id", r.authorizeMiddleware("finance.accounts.write"), handler.UpdateAccount)
	accounts.Delete("/:id", r.authorizeMiddleware("finance.accounts.write"), handler.DeleteAccount)

	// ── Transactions ──────────────────────────────────────────────────────
	txns := g.Group("/transactions")
	txns.Get("/", r.authorizeMiddleware("finance.transactions.read"), handler.ListTransactions)
	txns.Get("/:id", r.authorizeMiddleware("finance.transactions.read"), handler.GetTransaction)
	txns.Post("/", r.authorizeMiddleware("finance.transactions.write"), handler.CreateTransaction)
	txns.Post("/:id/submit", r.authorizeMiddleware("finance.transactions.write"), handler.SubmitTransaction)
	txns.Post("/:id/approve", r.authorizeMiddleware("finance.transactions.write"), handler.ApproveTransaction)
	txns.Post("/:id/reject", r.authorizeMiddleware("finance.transactions.write"), handler.RejectTransaction)

	// ── Fiscal years & periods ────────────────────────────────────────────
	fy := g.Group("/fiscal-years")
	fy.Get("/", r.authorizeMiddleware("finance.periods.read"), handler.ListFiscalYears)
	fy.Get("/:id", r.authorizeMiddleware("finance.periods.read"), handler.GetFiscalYear)
	fy.Post("/", r.authorizeMiddleware("finance.periods.write"), handler.CreateFiscalYear)
	fy.Get("/:id/periods", r.authorizeMiddleware("finance.periods.read"), handler.ListPeriods)
	fy.Post("/:id/periods", r.authorizeMiddleware("finance.periods.write"), handler.CreatePeriod)

	periods := g.Group("/periods")
	periods.Get("/current", r.authorizeMiddleware("finance.periods.read"), handler.GetCurrentPeriod)
	periods.Get("/:id", r.authorizeMiddleware("finance.periods.read"), handler.GetPeriod)
	periods.Patch("/:id/status", r.authorizeMiddleware("finance.periods.write"), handler.ChangePeriodStatus)

	// ── Currencies ────────────────────────────────────────────────────────
	currencies := g.Group("/currencies")
	currencies.Get("/", r.authorizeMiddleware("finance.currencies.read"), handler.ListCurrencies)
	currencies.Post("/", r.authorizeMiddleware("finance.currencies.write"), handler.CreateCurrency)
	currencies.Put("/:id", r.authorizeMiddleware("finance.currencies.write"), handler.UpdateCurrency)

	// ── Exchange rates ────────────────────────────────────────────────────
	rates := g.Group("/exchange-rates")
	rates.Get("/", r.authorizeMiddleware("finance.currencies.read"), handler.GetExchangeRate)
	rates.Get("/history", r.authorizeMiddleware("finance.currencies.read"), handler.ListExchangeRates)
	rates.Post("/", r.authorizeMiddleware("finance.currencies.write"), handler.UpsertExchangeRate)

	// ── Budgets ───────────────────────────────────────────────────────────
	budgets := g.Group("/budgets")
	budgets.Get("/", r.authorizeMiddleware("finance.budgets.read"), handler.ListBudgets)
	budgets.Get("/:id", r.authorizeMiddleware("finance.budgets.read"), handler.GetBudget)
	budgets.Get("/:id/lines", r.authorizeMiddleware("finance.budgets.read"), handler.GetBudgetLines)
	budgets.Post("/", r.authorizeMiddleware("finance.budgets.write"), handler.CreateBudget)
	budgets.Post("/:id/submit", r.authorizeMiddleware("finance.budgets.write"), handler.SubmitBudget)
	budgets.Post("/:id/approve", r.authorizeMiddleware("finance.budgets.write"), handler.ApproveBudget)
	budgets.Post("/:id/reject", r.authorizeMiddleware("finance.budgets.write"), handler.RejectBudget)
	budgets.Post("/:id/close", r.authorizeMiddleware("finance.budgets.write"), handler.CloseBudget)

	// ── Cost centres ──────────────────────────────────────────────────────
	costCenters := g.Group("/cost-centers")
	costCenters.Get("/", r.authorizeMiddleware("finance.cost_centers.read"), handler.ListCostCenters)
	costCenters.Get("/:id", r.authorizeMiddleware("finance.cost_centers.read"), handler.GetCostCenter)
	costCenters.Post("/", r.authorizeMiddleware("finance.cost_centers.write"), handler.CreateCostCenter)
	costCenters.Put("/:id", r.authorizeMiddleware("finance.cost_centers.write"), handler.UpdateCostCenter)
	costCenters.Delete("/:id", r.authorizeMiddleware("finance.cost_centers.write"), handler.DeleteCostCenter)

	// ── Tax authorities & codes ───────────────────────────────────────────
	tax := g.Group("/tax")
	taxAuth := tax.Group("/authorities")
	taxAuth.Get("/", r.authorizeMiddleware("finance.tax.read"), handler.ListTaxAuthorities)
	taxAuth.Get("/:id", r.authorizeMiddleware("finance.tax.read"), handler.GetTaxAuthority)
	taxAuth.Post("/", r.authorizeMiddleware("finance.tax.write"), handler.CreateTaxAuthority)
	taxAuth.Put("/:id", r.authorizeMiddleware("finance.tax.write"), handler.UpdateTaxAuthority)

	taxCodes := tax.Group("/codes")
	taxCodes.Get("/", r.authorizeMiddleware("finance.tax.read"), handler.ListTaxCodes)
	taxCodes.Get("/:id", r.authorizeMiddleware("finance.tax.read"), handler.GetTaxCode)
	taxCodes.Post("/", r.authorizeMiddleware("finance.tax.write"), handler.CreateTaxCode)
	taxCodes.Put("/:id", r.authorizeMiddleware("finance.tax.write"), handler.UpdateTaxCode)

	// ── Bank reconciliation ───────────────────────────────────────────────
	recon := g.Group("/reconciliation/statements")
	recon.Get("/", r.authorizeMiddleware("finance.reconciliation.read"), handler.ListBankStatements)
	recon.Get("/:id", r.authorizeMiddleware("finance.reconciliation.read"), handler.GetBankStatement)
	recon.Get("/:id/lines", r.authorizeMiddleware("finance.reconciliation.read"), handler.ListStatementLines)
	recon.Post("/", r.authorizeMiddleware("finance.reconciliation.write"), handler.ImportBankStatement)
	recon.Post("/:id/lines/:line_id/match", r.authorizeMiddleware("finance.reconciliation.write"), handler.MatchStatementLine)
	recon.Delete("/:id/lines/:line_id/match", r.authorizeMiddleware("finance.reconciliation.write"), handler.UnmatchStatementLine)
	recon.Post("/:id/complete", r.authorizeMiddleware("finance.reconciliation.write"), handler.CompleteReconciliation)

	// ── Reporting ─────────────────────────────────────────────────────────
	// TODO(finance-reporting): wire up reporting handler once deps are confirmed.
	// reports := g.Group("/reports")
	// reports.Get("/trial-balance", r.authorizeMiddleware("finance.reports.read"), handler.GetTrialBalance)

	r.registry.track(ModuleFinance, apiV1Prefix+"/finance", 54)
	r.deps.Logger.Info("finance API endpoints registered")
	return nil
}

// registerContractsAPI mounts contract lifecycle CRUD and state transitions.
//
// Route registration is delegated to handler.Routes() to keep the contract
// handler self-contained; this function applies the common auth + tenant
// middleware stack before handing off.
//
// Optional — skipped when ContractService is nil.
func (r *Router) registerContractsAPI(apiRouter fiber.Router) error {
	if r.deps.ContractService == nil {
		r.deps.Logger.Warn("ContractService not configured — contract routes skipped") // optional
		return nil
	}

	handler := contractHandler.New(
		r.deps.ContractService, r.deps.Logger, r.deps.Metrics, r.deps.Tracer,
	)

	// Create the group at the canonical contracts path.
	// handler.Routes() registers sub-paths relative to this group.
	g := apiRouter.Group(apiV1Prefix + "/contracts")
	g.Use(r.authenticateMiddleware())
	if r.deps.TenantMiddleware != nil {
		g.Use(r.deps.TenantMiddleware)
	}

	handler.Routes(g)

	r.registry.track(ModuleContracts, apiV1Prefix+"/contracts", 0 /* counted by handler */)
	r.deps.Logger.Info("contracts API endpoints registered")
	return nil
}

// registerSchemaAPI mounts the AMIS boot endpoint.
//
// GET /api/v1/schema/boot returns the full application shell (navigation,
// feature flags, initial permissions) used by the AMIS frontend on first load.
//
// Optional — skipped when Store is nil.
func (r *Router) registerSchemaAPI(apiRouter fiber.Router) error {
	if r.deps.Store == nil {
		r.deps.Logger.Warn("Store not configured — schema routes skipped") // optional
		return nil
	}

	g := apiRouter.Group(apiV1Prefix + "/schema")
	g.Use(r.authenticateMiddleware())

	// Wrap AuthzService in a PolicyChecker so the boot handler can filter nav
	// items by permission without calling .Enforce() directly (boundary guard).
	var checker contract.PolicyChecker
	if r.deps.AuthConfig != nil && r.deps.AuthConfig.AuthzService != nil {
		checker = contract.NewPolicyChecker(r.deps.AuthConfig.AuthzService)
	}
	g.Get("/boot", schemaHandler.BootHandler(r.deps.Store, checker))

	r.registry.track(ModuleSchema, apiV1Prefix+"/schema", 1)
	r.deps.Logger.Info("schema API endpoints registered")
	return nil
}

// registerIAMAPI mounts IAM management endpoints: policies, role assignments.
//
// Security model:
//   - authenticate: session or API key required.
//   - TenantMiddleware: domain derived from the tenant context.
//   - Endpoints that mutate policies are restricted to tenant admins via
//     authorizeMiddleware("iam.policies.write") / ("iam.roles.write").
//
// Optional — skipped when IAMService is nil.
func (r *Router) registerIAMAPI(apiRouter fiber.Router) error {
	if r.deps.IAMService == nil {
		r.deps.Logger.Warn("IAMService not configured — IAM management routes skipped") // optional
		return nil
	}

	handler := iamHandler.NewIAMHandler(r.deps.IAMService, r.deps.Logger, r.deps.Metrics, r.deps.Tracer)

	g := apiRouter.Group(apiV1Prefix + "/iam")
	g.Use(r.authenticateMiddleware())
	if r.deps.TenantMiddleware != nil {
		g.Use(r.deps.TenantMiddleware)
	}

	// ── Policies ──────────────────────────────────────────────────────────
	g.Get("/policies", r.authorizeMiddleware("iam.policies.read"), handler.ListPolicies)    // GET    /api/v1/iam/policies?domain=<tenantID>
	g.Post("/policies", r.authorizeMiddleware("iam.policies.write"), handler.AddPolicy)     // POST   /api/v1/iam/policies
	g.Delete("/policies", r.authorizeMiddleware("iam.policies.write"), handler.RemovePolicy) // DELETE /api/v1/iam/policies (body)

	// ── Role assignments ──────────────────────────────────────────────────
	g.Get("/assignments", r.authorizeMiddleware("iam.roles.read"), handler.ListAssignments)  // GET    /api/v1/iam/assignments?subject=<sub>&domain=<dom>
	g.Post("/roles/assign", r.authorizeMiddleware("iam.roles.write"), handler.AssignRole)    // POST   /api/v1/iam/roles/assign
	g.Post("/roles/revoke", r.authorizeMiddleware("iam.roles.write"), handler.RevokeRole)    // POST   /api/v1/iam/roles/revoke

	// ── Roles query ───────────────────────────────────────────────────────
	g.Get("/roles", r.authorizeMiddleware("iam.roles.read"), handler.GetRoles) // GET /api/v1/iam/roles?subject=<sub>&domain=<dom>

	r.registry.track(ModuleIAM, apiV1Prefix+"/iam", 7)
	r.deps.Logger.Info("IAM API endpoints registered")
	return nil
}

// registerAuditAPI mounts the audit log read endpoint.
//
// Permission enforced by authorizeMiddleware at the route level.
//
// Optional — skipped when AuditService is nil.
func (r *Router) registerAuditAPI(apiRouter fiber.Router) error {
	if r.deps.AuditService == nil {
		r.deps.Logger.Warn("AuditService not configured — audit routes skipped") // optional
		return nil
	}

	g := apiRouter.Group(apiV1Prefix + "/audit-logs")
	g.Use(r.authenticateMiddleware())

	g.Get("/", r.authorizeMiddleware("iam.sessions.read"), auditHandler.ListAuditEventsHandler(r.deps.AuditService))

	r.registry.track(ModuleAudit, apiV1Prefix+"/audit-logs", 1)
	r.deps.Logger.Info("audit API endpoints registered")
	return nil
}

// =============================================================================
// SECTION 7 — Shared middleware helpers
//
// These helpers are the single call-site for creating auth middleware so that
// the AuthConfig + APIKeyService coupling is resolved once and consistently.
// =============================================================================

// authenticateMiddleware returns the session/API-key authentication middleware.
//
// Returns a no-op pass-through when SessionService or AuthConfig is nil so that
// the application starts cleanly in dev/test mode without a session store.
func (r *Router) authenticateMiddleware() fiber.Handler {
	if r.deps.SessionService == nil || r.deps.AuthConfig == nil {
		// No-op: auth is not configured (dev/test mode).
		return func(c *fiber.Ctx) error { return c.Next() }
	}
	// Build a local copy so we can inject APIKeyService without mutating the
	// shared AuthConfig pointer (which may be used elsewhere).
	cfg := *r.deps.AuthConfig
	if cfg.APIKeyService == nil {
		cfg.APIKeyService = r.deps.APIKeyService
	}
	return middlewarePkg.Authenticate(cfg)
}

// authorizeMiddleware returns a permission-enforcement middleware for the given
// permission string (format: <module>.<resource>.<action>).
//
// Returns a no-op pass-through when AuthConfig or AuthzService is nil so that
// dev/test environments without a policy engine are not blocked.
func (r *Router) authorizeMiddleware(permission string) fiber.Handler {
	if r.deps.AuthConfig == nil || r.deps.AuthConfig.AuthzService == nil {
		// No-op: authz is not configured (dev/test mode).
		return func(c *fiber.Ctx) error { return c.Next() }
	}
	return middlewarePkg.Authorize(*r.deps.AuthConfig, permission)
}

// platformOnlyMiddleware returns a middleware that rejects callers whose actor
// type is not ActorPlatform (HTTP 403 Forbidden).
//
// Must run AFTER authenticateMiddleware — reads LocalsKeySession to determine
// the actor type. When auth is disabled (no-op authenticateMiddleware), the
// session local is nil and this middleware is also a no-op pass-through so
// that dev/test mode is not blocked.
func (r *Router) platformOnlyMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		resolved, ok := c.Locals(contract.LocalsKeySession).(*contract.ResolvedSession)
		if !ok || resolved == nil {
			// Auth disabled (dev/test) — pass through.
			return c.Next()
		}
		if !resolved.IsPlatform() {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   true,
				"status":  fiber.StatusForbidden,
				"message": "platform access required",
			})
		}
		return c.Next()
	}
}
