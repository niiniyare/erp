package routes

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"go.opentelemetry.io/otel/attribute"

	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// RouteSetupFunc defines the function signature for setting up routes in a module
type RouteSetupFunc func(router fiber.Router)

// ModuleInfo contains metadata about a registered module
type ModuleInfo struct {
	Name        string
	BasePath    string
	Middlewares []string
	RouteCount  int
}

// RouteRegistry manages route registration and middleware application
type RouteRegistry struct {
	logger           logger.Logger
	metrics          metrics.MetricsProvider
	tracer           tracing.TracingService
	registeredModules map[string]*ModuleInfo
	registeredPaths   map[string]bool
	mu               sync.RWMutex
}

// NewRouteRegistry creates a new route registry with dependencies
func NewRouteRegistry(logger logger.Logger, metrics metrics.MetricsProvider, tracer tracing.TracingService) *RouteRegistry {
	return &RouteRegistry{
		logger:            logger,
		metrics:           metrics,
		tracer:            tracer,
		registeredModules: make(map[string]*ModuleInfo),
		registeredPaths:   make(map[string]bool),
	}
}

// RegisterModule registers a module with routes at the specified base path
func (r *RouteRegistry) RegisterModule(app *fiber.App, moduleName, basePath string, setupFunc RouteSetupFunc) error {
	return r.RegisterModuleWithMiddleware(app, moduleName, basePath, nil, setupFunc)
}

// RegisterModuleWithMiddleware registers a module with specified middleware
func (r *RouteRegistry) RegisterModuleWithMiddleware(app *fiber.App, moduleName, basePath string, middlewares []string, setupFunc RouteSetupFunc) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	ctx := context.Background()
	ctx, span := r.tracer.StartSpan(ctx, "route.register_module")
	defer span.End()

	// Validate inputs
	if moduleName == "" {
		return errors.New("module name cannot be empty")
	}

	if err := r.validatePath(basePath); err != nil {
		return fmt.Errorf("invalid base path for module %s: %w", moduleName, err)
	}

	// Check for module conflicts
	if _, exists := r.registeredModules[moduleName]; exists {
		return fmt.Errorf("module %s already registered", moduleName)
	}

	// Check for path conflicts
	if r.registeredPaths[basePath] {
		return fmt.Errorf("path %s already registered", basePath)
	}

	// Log module registration
	r.logger.InfoContext(ctx, "Registering module", logger.Fields{
		"module":      moduleName,
		"base_path":   basePath,
		"middlewares": middlewares,
	})

	// Create route group
	var group fiber.Router
	if basePath == "" {
		group = app
	} else {
		group = app.Group(basePath)
	}

	// Apply middlewares
	for _, middleware := range middlewares {
		if err := r.applyMiddleware(group, middleware); err != nil {
			r.logger.ErrorContext(ctx, "Failed to apply middleware", logger.Fields{
				"module":     moduleName,
				"middleware": middleware,
				"error":      err.Error(),
			})
			return fmt.Errorf("failed to apply middleware %s: %w", middleware, err)
		}
	}

	// Setup routes
	setupFunc(group)

	// Track registration
	r.registeredModules[moduleName] = &ModuleInfo{
		Name:        moduleName,
		BasePath:    basePath,
		Middlewares: middlewares,
		RouteCount:  1, // Simplified for now
	}
	r.registeredPaths[basePath] = true

	// Record metrics
	r.metrics.IncrementCounter("routes_registered_total", metrics.Fields{
		"module": moduleName,
	})

	// Add span attributes
	span.SetAttributes(
		attribute.String("module.name", moduleName),
		attribute.String("module.base_path", basePath),
		attribute.Int("module.middleware_count", len(middlewares)),
	)

	r.logger.InfoContext(ctx, "Module registered successfully", logger.Fields{
		"module":    moduleName,
		"base_path": basePath,
	})

	return nil
}

// ValidatePath validates a route path for security and format compliance
func (r *RouteRegistry) ValidatePath(path string) error {
	return r.validatePath(path)
}

func (r *RouteRegistry) validatePath(path string) error {
	// Allow empty path for root-level registration
	if path == "" {
		return nil
	}

	// Must start with /
	if !strings.HasPrefix(path, "/") {
		return errors.New("path must start with /")
	}

	// Check for invalid characters
	invalidChars := []string{"<", ">", "`", "'", "&", "<script>", "javascript:"}
	for _, char := range invalidChars {
		if strings.Contains(path, char) {
			return fmt.Errorf("path contains invalid character: %s", char)
		}
	}

	// Validate path format with regex
	validPath := regexp.MustCompile(`^/[a-zA-Z0-9/_:.-]*$`)
	if !validPath.MatchString(path) {
		return errors.New("path contains invalid characters")
	}

	return nil
}

// applyMiddleware applies the specified middleware to the router
func (r *RouteRegistry) applyMiddleware(router fiber.Router, middlewareType string) error {
	switch middlewareType {
	case "cors":
		router.Use(cors.New(cors.Config{
			AllowOrigins: "*",
			AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
			AllowHeaders: "Origin,Content-Type,Accept,Authorization,X-Tenant-ID",
		}))
	case "security":
		router.Use(helmet.New())
	case "tenant":
		// Custom tenant middleware - placeholder for now
		router.Use(func(c *fiber.Ctx) error {
			c.Set("X-Tenant-Context", "enabled")
			return c.Next()
		})
	default:
		return fmt.Errorf("unknown middleware type: %s", middlewareType)
	}
	return nil
}

// GetRegisteredModules returns a list of all registered module names
func (r *RouteRegistry) GetRegisteredModules() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	modules := make([]string, 0, len(r.registeredModules))
	for name := range r.registeredModules {
		modules = append(modules, name)
	}
	return modules
}

// GetModuleInfo returns information about a registered module
func (r *RouteRegistry) GetModuleInfo(moduleName string) (*ModuleInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	module, exists := r.registeredModules[moduleName]
	if !exists {
		return nil, false
	}

	// Return a copy to prevent external modification
	return &ModuleInfo{
		Name:        module.Name,
		BasePath:    module.BasePath,
		Middlewares: append([]string(nil), module.Middlewares...),
		RouteCount:  module.RouteCount,
	}, true
}

// ListRoutes returns a summary of all registered routes (simplified implementation)
func (r *RouteRegistry) ListRoutes() map[string]*ModuleInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]*ModuleInfo)
	for name, module := range r.registeredModules {
		result[name] = &ModuleInfo{
			Name:        module.Name,
			BasePath:    module.BasePath,
			Middlewares: append([]string(nil), module.Middlewares...),
			RouteCount:  module.RouteCount,
		}
	}
	return result
}