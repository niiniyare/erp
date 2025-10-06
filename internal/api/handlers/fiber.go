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

	// Header-based routing middleware
	app.Use(headerBasedRoutingMiddleware())

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

func headerBasedRoutingMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		acceptHeader := c.Get("Accept")
		contentType := c.Get("Content-Type")

		// Determine response type based on headers
		responseType := "html" // default
		if acceptHeader == "application/json" ||
			contentType == "application/json" ||
			c.Path() == "/api/*" ||
			c.Path()[:4] == "/api" {
			responseType = "json"
		}

		// Set context values for downstream handlers
		c.Locals("responseType", responseType)
		c.Locals("tenantID", c.Get("X-Tenant-ID"))

		return c.Next()
	}
}

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

	// Create handlers
	authHandler := NewAuthFiberHandler(coreServices, tracingService, metricsService)
	tenantHandler := NewTenantFiberHandler(coreServices, tracingService, metricsService)

	// Setup route groups
	SetupAuthRoutes(v1, authHandler)
	SetupTenantRoutes(v1, tenantHandler)

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

	// Serve UI for HTML requests (catch-all route)
	app.Get("/*", func(c *fiber.Ctx) error {
		responseType := c.Locals("responseType")
		if responseType == "json" {
			return c.Status(404).JSON(fiber.Map{
				"error":   "Not Found",
				"code":    404,
				"path":    c.Path(),
				"method":  c.Method(),
			})
		}
		// For HTML responses, serve the UI
		return c.SendFile(staticPath + "/index.html")
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

		responseType := c.Locals("responseType")
		if responseType == "json" {
			return c.Status(code).JSON(fiber.Map{
				"error":   message,
				"code":    code,
				"path":    c.Path(),
				"method":  c.Method(),
			})
		}

		// For HTML responses, serve error page
		return c.Status(code).SendString(`
			<html>
				<head><title>Error ` + strconv.Itoa(code) + `</title></head>
				<body>
					<h1>Error ` + strconv.Itoa(code) + `</h1>
					<p>` + message + `</p>
					<hr>
					<small>ERP System</small>
				</body>
			</html>
		`)
	}
}

