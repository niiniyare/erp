package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/niiniyare/erp/cmd/server/services"
	db "github.com/niiniyare/erp/db/sqlc"
	financeService "github.com/niiniyare/erp/internal/core/finance/service"
	"github.com/niiniyare/erp/internal/core/iam"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/platform/config"
	"github.com/niiniyare/erp/internal/platform/middleware"
	sharedLogger "github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

type FiberServer struct {
	App    *fiber.App
	Config *config.Config
}

func NewFiberServer(
	cfg *config.Config,
	coreServices *services.CoreServices,
	financeServices *financeService.Services,
	store db.Store,
	cacheService cache.Service,
	metricsService *metrics.MetricsService,
	tracingService tracing.TracingService,
	iamService iam.Service,
) (*FiberServer, error) {
	// Create Fiber app with configuration from config
	app := fiber.New(fiber.Config{
		ErrorHandler:  createFiberErrorHandler(),
		ReadTimeout:   cfg.Server.ReadTimeout,
		WriteTimeout:  cfg.Server.WriteTimeout,
		IdleTimeout:   120 * time.Second,
		Prefork:       false, // Set to true in production for better performance
		ServerHeader:  "ERP-API",
		AppName:       cfg.App.Name + " v" + cfg.App.Version,
		CaseSensitive: false,
		StrictRouting: false,
	})

	// Initialize Fiber middleware
	fiberMiddleware := middleware.NewFiberMiddleware(
		cfg,
		coreServices.TenantService,
		iamService,
		store,
		sharedLogger.WithFields(sharedLogger.Fields{"component": "middleware"}),
		metricsService,
		tracingService,
	)

	// Add middleware stack
	setupMiddleware(app, cfg, fiberMiddleware)

	// Content negotiation middleware (replaces simple header-based routing)
	app.Use(middleware.ContentNegotiationMiddleware())

	// Initialize all routes
	setupRoutes(app, cfg, coreServices, financeServices, store, cacheService, metricsService, tracingService, fiberMiddleware)

	return &FiberServer{
		App:    app,
		Config: cfg,
	}, nil
}

func setupMiddleware(app *fiber.App, cfg *config.Config, fiberMiddleware *middleware.FiberMiddleware) {
	// Recovery middleware (should be first)
	app.Use(recover.New())

	// Request ID middleware
	app.Use(requestid.New())

	// Security middleware
	app.Use(helmet.New())

	// CORS middleware with multi-tenant support
	app.Use(fiberMiddleware.CORSMiddleware())

	// Rate limiting middleware
	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
	}))

	// Request logging middleware
	if cfg.Logger.Level == "debug" || cfg.App.Debug {
		app.Use(logger.New(logger.Config{
			Format:     "[${time}] ${status} - ${method} ${path} ${latency} | ${ip} | ${reqHeaders}\n",
			TimeFormat: "15:04:05",
		}))
	} else {
		app.Use(logger.New(logger.Config{
			Format:     "[${time}] ${status} - ${method} ${path} ${latency}\n",
			TimeFormat: "15:04:05",
		}))
	}

	// Validation middleware
	app.Use(fiberMiddleware.ValidationMiddleware())

	// Tenant middleware (applied to all routes except public ones)
	app.Use(fiberMiddleware.TenantMiddleware())

	// JWT authentication middleware (applied to protected routes)
	app.Use(fiberMiddleware.JWTAuthMiddleware())
}

// Note: headerBasedRoutingMiddleware replaced by middleware.ContentNegotiationMiddleware()

func setupRoutes(
	app *fiber.App,
	cfg *config.Config,
	coreServices *services.CoreServices,
	financeServices *financeService.Services,
	store db.Store,
	cacheService cache.Service,
	metricsService *metrics.MetricsService,
	tracingService tracing.TracingService,
	fiberMiddleware *middleware.FiberMiddleware,
) {
	// Health endpoints (public, no auth required)
	healthChecker := NewHealthChecker(store, nil, sharedLogger.WithFields(sharedLogger.Fields{}), metricsService, tracingService)

	app.Get("/health", func(c *fiber.Ctx) error {
		dependencies := healthChecker.CheckDependencies(c.Context())

		// Determine overall status
		overallStatus := "healthy"
		for _, result := range dependencies {
			if result.Status == "critical" {
				overallStatus = "unhealthy"
				break
			} else if result.Status == "warning" && overallStatus == "healthy" {
				overallStatus = "degraded"
			}
		}

		return Success(c, fiber.Map{
			"status":       overallStatus,
			"timestamp":    time.Now(),
			"dependencies": dependencies,
		})
	})

	app.Get("/api/v1/health", func(c *fiber.Ctx) error {
		dependencies := healthChecker.CheckDependencies(c.Context())

		// Determine overall status
		overallStatus := "healthy"
		for _, result := range dependencies {
			if result.Status == "critical" {
				overallStatus = "unhealthy"
				break
			} else if result.Status == "warning" && overallStatus == "healthy" {
				overallStatus = "degraded"
			}
		}

		return Success(c, fiber.Map{
			"status":       overallStatus,
			"timestamp":    time.Now(),
			"dependencies": dependencies,
		})
	})

	// API v1 routes
	v1 := app.Group("/api/v1")

	// Create unified handlers that serve both HTML and JSON
	authHandler := NewAuthUnifiedHandler(coreServices, tracingService, metricsService)
	tenantHandler := NewTenantUnifiedHandler(coreServices, tracingService, metricsService)

	// Setup unified route groups (handle both HTML and JSON responses)
	SetupAuthUnifiedRoutes(app, authHandler)     // Root level for UI routes
	SetupTenantUnifiedRoutes(app, tenantHandler) // Root level for UI routes

	// API v1 routes (explicit JSON endpoints)
	SetupAuthUnifiedRoutes(v1, authHandler)     // API endpoints
	SetupTenantUnifiedRoutes(v1, tenantHandler) // API endpoints

	// TODO: Implement other route groups with proper handlers
	// setupUserRoutes(v1, coreServices, tracingService, metricsService)
	// setupOrganizationRoutes(v1, coreServices, tracingService, metricsService)
	// setupABACRoutes(v1, coreServices, tracingService, metricsService)
	// setupFeatureFlagRoutes(v1, coreServices, tracingService, metricsService)

	// Finance routes (if available)
	if financeServices != nil {
		setupFinanceRoutes(v1, financeServices, tracingService, metricsService)
	}

	// Static UI routes (for HTML responses)
	setupStaticRoutes(app, cfg)

	sharedLogger.Info("Fiber server routes initialized", sharedLogger.Fields{
		"api_version": "v1",
		"endpoints":   "auth, tenants, health",
		"port":        cfg.Server.Port,
	})
}

func setupFinanceRoutes(v1 fiber.Router, financeServices *financeService.Services, tracingService tracing.TracingService, metricsService *metrics.MetricsService) {
	// TODO: Implement proper Fiber finance handlers
	finance := v1.Group("/finance")

	finance.Get("/accounts", func(c *fiber.Ctx) error {
		return Success(c, fiber.Map{"message": "Finance accounts endpoint - TODO: implement"})
	})

	finance.Get("/transactions", func(c *fiber.Ctx) error {
		return Success(c, fiber.Map{"message": "Finance transactions endpoint - TODO: implement"})
	})
}

func setupStaticRoutes(app *fiber.App, cfg *config.Config) {
	// Serve static files from web directory if it exists
	staticPath := "./web/static"
	app.Static("/static", staticPath)

	// Dashboard route (home page after login)
	app.Get("/dashboard", func(c *fiber.Ctx) error {
		responseMode := middleware.GetResponseMode(c)

		switch responseMode {
		case middleware.ResponseModeJSON:
			return Success(c, fiber.Map{
				"message": "Dashboard data",
				"user":    "authenticated_user_data",
			})
		default:
			// Serve dashboard HTML page
			html := `<!DOCTYPE html>
<html>
<head>
    <title>Dashboard - ERP System</title>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <script src="https://unpkg.com/htmx.org@1.9.8"></script>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-50">
    <div class="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <div class="px-4 py-6 sm:px-0">
            <div class="border-4 border-dashed border-gray-200 rounded-lg p-6">
                <h1 class="text-2xl font-bold text-gray-900 mb-6">Dashboard</h1>
                <p class="text-gray-600">Welcome to the ERP System!</p>
                <div class="mt-4">
                    <a href="/tenants" class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700">
                        View Tenants
                    </a>
                    <button hx-post="/logout" class="ml-2 inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50">
                        Logout
                    </button>
                </div>
            </div>
        </div>
    </div>
</body>
</html>`
			c.Set("Content-Type", "text/html; charset=utf-8")
			return c.SendString(html)
		}
	})

	// Serve UI for HTML requests (catch-all route)
	app.Get("/*", func(c *fiber.Ctx) error {
		if middleware.WantsJSON(c) {
			return c.Status(404).JSON(fiber.Map{
				"error":  "Not Found",
				"code":   404,
				"path":   c.Path(),
				"method": c.Method(),
			})
		}

		// For HTML responses, redirect to login page
		return c.Redirect("/login", 302)
	})
}

func createFiberErrorHandler() fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		message := "Internal Server Error"

		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
			message = e.Message
		}

		// Use content negotiation middleware to determine response format
		if middleware.WantsJSON(c) {
			return c.Status(code).JSON(fiber.Map{
				"error":  message,
				"code":   code,
				"path":   c.Path(),
				"method": c.Method(),
			})
		}

		// For HTML responses, serve styled error page
		html := `<!DOCTYPE html>
<html>
<head>
    <title>Error ` + strconv.Itoa(code) + ` - ERP System</title>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-50">
    <div class="min-h-screen flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
        <div class="max-w-md w-full space-y-8">
            <div class="text-center">
                <h1 class="text-6xl font-bold text-gray-900">` + strconv.Itoa(code) + `</h1>
                <h2 class="mt-2 text-3xl font-bold text-gray-900">Error</h2>
                <p class="mt-2 text-sm text-gray-600">` + message + `</p>
                <div class="mt-5">
                    <a href="/" class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700">
                        Go Home
                    </a>
                </div>
            </div>
        </div>
    </div>
</body>
</html>`
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.Status(code).SendString(html)
	}
}
