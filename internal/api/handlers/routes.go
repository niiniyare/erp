package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"

	"github.com/niiniyare/erp/internal/api/handlers/health"
	"github.com/niiniyare/erp/internal/api/routes"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// Module names as constants for consistency
const (
	ModuleHealth = "health"
	// ModuleTenant  = "tenant"
	// ModuleUser    = "user"
	// ModuleFinance = "finance"
)

// Dependencies contains all required dependencies for handlers.
// All fields are required and must be non-nil.
type Dependencies struct {
	Logger  logger.Logger
	Metrics metrics.MetricsProvider
	Tracer  tracing.TracingService
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
	registry *routes.RouteRegistry
	deps     *Dependencies
}

// NewRouter creates a new router with the given dependencies.
func NewRouter(deps *Dependencies) (*Router, error) {
	if err := deps.Validate(); err != nil {
		return nil, err
	}

	return &Router{
		registry: routes.NewRouteRegistry(deps.Logger, deps.Metrics, deps.Tracer),
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
		// Add future modules here:
		// {ModuleTenant, r.registerTenant},
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
	handler := health.NewHealthHandler(r.deps.Logger, r.deps.Metrics, r.deps.Tracer)

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
	)
}

// Future module registration examples:
//
// func (r *Router) registerTenant(app *fiber.App) error {
// 	handler := tenant.NewHandler(r.deps.Logger, r.deps.Metrics, r.deps.Tracer)
//
// 	return r.registry.RegisterModuleWithMiddleware(
// 		app,
// 		ModuleTenant,
// 		"/api/v1/tenants",
// 		[]string{"cors", "auth", "ratelimit"},
// 		func(router fiber.Router) {
// 			router.Get("/", handler.List)
// 			router.Post("/", handler.Create)
// 			router.Get("/:id", handler.GetByID)
// 			router.Put("/:id", handler.Update)
// 			router.Delete("/:id", handler.Delete)
// 		},
// 	)
// }
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
func (r *Router) ListRoutes() map[string]*routes.ModuleInfo {
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
func (r *Router) GetModuleInfo(moduleName string) (*routes.ModuleInfo, error) {
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
