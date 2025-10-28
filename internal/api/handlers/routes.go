package handlers

import (
	"fmt"
	"sync"

	"github.com/gofiber/fiber/v2"

	"github.com/niiniyare/erp/internal/api/handlers/health"
	middlewarePkg "github.com/niiniyare/erp/internal/api/middleware"
	tenantHandler "github.com/niiniyare/erp/internal/api/handlers/tenant"
	coreTenant "github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// Module names as constants for consistency
const (
	ModuleHealth = "health"
	ModuleTenant = "tenant"
	// ModuleUser    = "user"
	// ModuleFinance = "finance"
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
	tracer            tracing.TracingService
	registeredModules map[string]*ModuleInfo
	registeredPaths   map[string]bool
	mu                sync.RWMutex
}

// NewRouteRegistry creates a new route registry
func NewRouteRegistry(logger logger.Logger, metrics metrics.MetricsProvider, tracer tracing.TracingService) *RouteRegistry {
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

	// Apply middleware instances based on middleware names
	for _, middlewareName := range middleware {
		switch middlewareName {
		case "cors":
			// Use development CORS config for tests
			corsConfig := middlewarePkg.DevelopmentCORSConfig([]int{3000, 8080})
			router.Use(middlewarePkg.NewCORSMiddleware(corsConfig))
		case "auth":
			// TODO: Apply auth middleware when available
			// router.Use(middleware.AuthMiddleware())
		case "ratelimit":
			// TODO: Apply rate limiting middleware when available
			// router.Use(middleware.RateLimitMiddleware())
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
	Tracer           tracing.TracingService
	conf             *health.Config
	TenantService    coreTenant.Service
	TenantMiddleware fiber.Handler
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

// RegisterAll registers all application routes.
func (r *Router) RegisterAll(app *fiber.App) error {
	// Define modules to register in order
	modules := []struct {
		name string
		fn   func(*fiber.App) error
	}{
		{ModuleHealth, r.registerHealth},
		{ModuleTenant, r.registerTenant},
		// Add future modules here:
		// {ModuleUser, r.registerUser},
		// {ModuleFinance, r.registerFinance},
	}

	// Register each module
	for _, module := range modules {
		if err := module.fn(app); err != nil {
			return errors.NewBusinessError("MODULE_REGISTRATION_FAILED",
				fmt.Sprintf("failed to register %s module", module.name)).
				WithCategory(errors.CategorySystem).
				WithSeverity(errors.SeverityCritical).
				WithDetail("module_name", module.name).
				WithDetail("error", err.Error()).
				WithSuggestion("Check module configuration and dependencies")
		}

		r.deps.Logger.Info(fmt.Sprintf("registered %s module", module.name))
	}

	return nil
}

// registerHealth registers health check routes.
func (r *Router) registerHealth(app *fiber.App) error {
	// FIXME:pass real system Config here
	handler := health.NewHealthHandler(r.deps.Logger, r.deps.Metrics, r.deps.Tracer, nil)

	return r.registry.RegisterModuleWithMiddleware(
		app,
		ModuleHealth,
		"/health",
		[]string{"cors"}, // Minimal middleware for health checks
		func(router fiber.Router) {
			router.Get("/", handler.Get)

			// Uncomment when implementing Kubernetes-style health checks:
			// router.Get("/ready", handler.Ready)   // Readiness probe
			// router.Get("/live", handler.Live)     // Liveness probe
			// router.Get("/startup", handler.Startup) // Startup probe
		},
		r.deps,
	)
}

// registerTenant registers tenant management routes.
func (r *Router) registerTenant(app *fiber.App) error {
	// Use existing tenant service
	if r.deps.TenantService == nil {
		return errors.NewBusinessError("MISSING_TENANT_SERVICE", "Tenant service is required").
			WithCategory(errors.CategorySystem).
			WithSeverity(errors.SeverityCritical).
			WithSuggestion("Ensure tenant service is initialized before creating router")
	}

	handler := tenantHandler.NewTenantHandler(r.deps.TenantService, r.deps.Logger, r.deps.Metrics, r.deps.Tracer)

	return r.registry.RegisterModuleWithMiddleware(
		app,
		ModuleTenant,
		"/api/v1/tenants",
		[]string{"cors", "tenant", "auth", "ratelimit"}, // Include tenant middleware for RLS
		func(router fiber.Router) {
			router.Get("/", handler.List)         // GET /api/v1/tenants - List tenants with pagination
			router.Post("/", handler.Create)      // POST /api/v1/tenants - Create new tenant
			router.Get("/:id", handler.Get)       // GET /api/v1/tenants/:id - Get tenant by ID
			router.Put("/:id", handler.Update)    // PUT /api/v1/tenants/:id - Update tenant
			router.Delete("/:id", handler.Delete) // DELETE /api/v1/tenants/:id - Delete tenant
		},
		r.deps,
	)
}

// Future module registration examples:
//
// func (r *Router) registerUser(app *fiber.App) error {
// 	handler := user.NewHandler(r.deps.Logger, r.deps.Metrics, r.deps.Tracer)
//
// 	return r.registry.RegisterModuleWithMiddleware(
// 		app,
// 		ModuleUser,
// 		"/api/v1/users",
// 		[]string{"cors", "auth", "ratelimit"},
// 		func(router fiber.Router) {
// 			router.Get("/", handler.List)
// 			router.Post("/", handler.Create)
// 			router.Get("/:id", handler.GetByID)
// 			router.Put("/:id", handler.Update)
// 			router.Delete("/:id", handler.Delete)
// 			router.Post("/:id/reset-password", handler.ResetPassword)
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
