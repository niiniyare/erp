package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/niiniyare/erp/internal/adapters"
	"github.com/niiniyare/erp/internal/api/handlers"
	"github.com/niiniyare/erp/internal/core/access/approval"
	"github.com/niiniyare/erp/internal/core/access/conditional"
	"github.com/niiniyare/erp/internal/core/access/execution"
	"github.com/niiniyare/erp/internal/core/access/request"
	"github.com/niiniyare/erp/internal/core/analytics"
	"github.com/niiniyare/erp/internal/core/audit"
	"github.com/niiniyare/erp/internal/core/entity"
	"github.com/niiniyare/erp/internal/core/identity"
	"github.com/niiniyare/erp/internal/core/notification"
	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/platform/config"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	db "github.com/niiniyare/erp/db/sqlc"

	// GOA generated packages
	auth "github.com/niiniyare/erp/gen/auth"
	authsvr "github.com/niiniyare/erp/gen/http/auth/server"
	openapisvr "github.com/niiniyare/erp/gen/http/openapi/server"
	organizationsvr "github.com/niiniyare/erp/gen/http/organization/server"
	tenantsvr "github.com/niiniyare/erp/gen/http/tenant/server"
	usersvr "github.com/niiniyare/erp/gen/http/user/server"
	openapi "github.com/niiniyare/erp/gen/openapi"
	organization "github.com/niiniyare/erp/gen/organization"
	goaTenant "github.com/niiniyare/erp/gen/tenant"
	goaUser "github.com/niiniyare/erp/gen/user"
	"goa.design/clue/debug"
	clueLog "goa.design/clue/log"
	goahttp "goa.design/goa/v3/http"
)

func main() {
	// Initialize logger from environment
	err := logger.InitializeFromEnv()
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	logger.Info("Starting Awo ERP server", logger.Fields{
		"service": "awo-server",
		"version": "1.0.0",
	})

	// Load configuration
	cfg := config.Load()

	logger.Info("Configuration loaded", logger.Fields{
		"server_port": cfg.Server.Port,
		"db_host":     cfg.Database.Host,
		"redis_host":  cfg.Redis.Host,
	})

	// Initialize tracing
	tracingService, err := tracing.NewTracingService(tracing.TracingConfig{
		ServiceName:    "awo-server",
		ServiceVersion: "1.0.0",
		Environment:    "development",
		ExporterType:   tracing.StdoutExporter,
		SamplingRatio:  1.0,
	})
	if err != nil {
		logger.Fatal("Failed to initialize tracing", logger.Fields{"error": err})
	}
	defer tracingService.Shutdown(context.Background())

	// Initialize metrics
	metricsService, err := metrics.NewMetricsService(metrics.MetricsConfig{
		Namespace: "awo-erp",
		Subsystem: "server",
		Provider:  "prometheus",
		Enabled:   true,
	})
	if err != nil {
		logger.Fatal("Failed to initialize metrics", logger.Fields{"error": err})
	}

	// Build database URL from config
	databaseURL := cfg.Database.GetDatabaseURL()

	// Run database migrations before initializing the store
	migrationURL := os.Getenv("MIGRATION_URL")
	if migrationURL == "" {
		migrationURL = "file://db/migration" // Default migration path
	}

	logger.Info("Running database migrations", logger.Fields{
		"migration_url": migrationURL,
		"db_host":       cfg.Database.Host,
		"db_name":       cfg.Database.Database,
	})

	runDBMigration(migrationURL, databaseURL)

	logger.Info("Database migrations completed successfully")

	// Initialize database store using SQLC
	store, err := db.NewDB(databaseURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", logger.Fields{
			"error":        err.Error(),
			"database_url": databaseURL,
		})
	}
	defer store.Close()

	logger.Info("Database connection established", logger.Fields{
		"database": cfg.Database.Database,
	})

	// Initialize cache
	redisConfig := cache.DefaultRedisConfig(&cfg.Redis)
	redisClient := cache.NewRedisClient(redisConfig)

	logger.Info("Cache client initialized", logger.Fields{
		"redis_host": cfg.Redis.Host,
		"redis_port": cfg.Redis.Port,
	})

	// Initialize repositories
	tenantRepo := tenant.NewRepository(store, tracingService)
	entityRepo := entity.NewRepository(store, tracingService, metricsService)
	identityRepo := identity.NewRepository(store, tracingService, metricsService)
	auditRepo := audit.NewRepository(store)
	notificationRepo := notification.NewRepository(store)

	// Initialize core business services
	tenantService := tenant.NewService(tenantRepo, redisClient, tracingService)
	entityService := entity.NewService(entityRepo, tracingService, metricsService)
	identityService := identity.NewService(identityRepo, redisClient, tracingService, metricsService)

	// Initialize domain services
	auditService := audit.NewService(auditRepo)
	approverService := approval.NewApproverService(identityService, tracingService, metricsService)
	notificationService := notification.NewNotificationService(tracingService, metricsService, nil, nil, notificationRepo)
	executionService := execution.NewAccessExecutionService(identityRepo, identityService, tracingService, metricsService)

	// Create adapters for interface compatibility
	userServiceAdapter := adapters.NewUserServiceAdapter(identityService)
	auditServiceAdapter := adapters.NewAuditServiceAdapter(auditService)

	accessRequestService := request.NewAccessRequestService(
		nil, // TODO: Implement AccessRequestRepository
		redisClient,
		tracingService,
		metricsService,
		userServiceAdapter,
		notificationService,
		approverService,
		executionService,
		auditService,
	)
	conditionalAccessService := conditional.NewConditionalAccessService(tracingService, metricsService, auditServiceAdapter)
	analyticsService := analytics.NewUserAnalyticsService(tracingService, metricsService, auditService)

	logger.Info("Core business services initialized", logger.Fields{
		"tenant_service":         "ready",
		"entity_service":         "ready",
		"user_service":           "ready",
		"access_request_service": "ready",
		"conditional_access":     "ready",
		"analytics_service":      "ready",
	})

	// Initialize GOA services using handlers package following Clean Architecture
	var (
		authSvc         auth.Service
		organizationSvc organization.Service
		tenantSvc       goaTenant.Service
		userSvc         goaUser.Service
		openapiSvc      openapi.Service
	)
	{
		// Use handlers package following the data flow pattern
		authSvc = handlers.NewAuthHandler(identityService, tracingService, metricsService)
		organizationSvc = handlers.NewOrganizationGoaHandler(entityService, tracingService, metricsService)
		tenantSvc = handlers.NewTenantGoaHandler(tenantService, tracingService, metricsService)
		userSvc = handlers.NewUserGoaHandler(identityService, accessRequestService, conditionalAccessService, analyticsService, tracingService, metricsService)
		openapiSvc = handlers.NewOpenapiHandler()
	}

	// Create GOA endpoints
	var (
		authEndpoints         *auth.Endpoints
		organizationEndpoints *organization.Endpoints
		tenantEndpoints       *goaTenant.Endpoints
		userEndpoints         *goaUser.Endpoints
		openapiEndpoints      *openapi.Endpoints
	)
	{
		authEndpoints = auth.NewEndpoints(authSvc)
		authEndpoints.Use(debug.LogPayloads())
		authEndpoints.Use(clueLog.Endpoint)
		organizationEndpoints = organization.NewEndpoints(organizationSvc)
		organizationEndpoints.Use(debug.LogPayloads())
		organizationEndpoints.Use(clueLog.Endpoint)
		tenantEndpoints = goaTenant.NewEndpoints(tenantSvc)
		tenantEndpoints.Use(debug.LogPayloads())
		tenantEndpoints.Use(clueLog.Endpoint)
		userEndpoints = goaUser.NewEndpoints(userSvc)
		userEndpoints.Use(debug.LogPayloads())
		userEndpoints.Use(clueLog.Endpoint)
		openapiEndpoints = openapi.NewEndpoints(openapiSvc)
		openapiEndpoints.Use(debug.LogPayloads())
		openapiEndpoints.Use(clueLog.Endpoint)
	}

	// Create GOA HTTP mux
	var (
		dec = goahttp.RequestDecoder
		enc = goahttp.ResponseEncoder
	)

	var mux goahttp.Muxer
	{
		mux = goahttp.NewMuxer()
		// Mount debug handlers
		debug.MountPprofHandlers(debug.Adapt(mux))
		debug.MountDebugLogEnabler(debug.Adapt(mux))
	}

	// Create GOA HTTP servers
	var (
		authServer         *authsvr.Server
		organizationServer *organizationsvr.Server
		tenantServer       *tenantsvr.Server
		userServer         *usersvr.Server
		openapiServer      *openapisvr.Server
	)
	{
		eh := func(ctx context.Context, w http.ResponseWriter, err error) {
			logger.Error("HTTP Error", logger.Fields{"error": err.Error()})
		}
		authServer = authsvr.New(authEndpoints, mux, dec, enc, eh, nil)
		organizationServer = organizationsvr.New(organizationEndpoints, mux, dec, enc, eh, nil)
		tenantServer = tenantsvr.New(tenantEndpoints, mux, dec, enc, eh, nil)
		userServer = usersvr.New(userEndpoints, mux, dec, enc, eh, nil)
		openapiServer = openapisvr.New(openapiEndpoints, mux, dec, enc, eh, nil)
	}

	// Mount GOA HTTP servers
	authsvr.Mount(mux, authServer)
	organizationsvr.Mount(mux, organizationServer)
	tenantsvr.Mount(mux, tenantServer)
	usersvr.Mount(mux, userServer)
	openapisvr.Mount(mux, openapiServer)

	// Create GOA-compatible context with logger
	format := clueLog.FormatJSON
	if clueLog.IsTerminal() {
		format = clueLog.FormatTerminal
	}
	ctx := clueLog.Context(context.Background(), clueLog.WithFormat(format))

	// Add logging and debugging to the GOA handler
	var handler http.Handler = mux
	handler = clueLog.HTTP(ctx)(handler)
	handler = debug.HTTP()(handler)

	// Log mounted endpoints
	for _, m := range authServer.Mounts {
		logger.Info("GOA HTTP endpoint mounted", logger.Fields{
			"method":  m.Method,
			"verb":    m.Verb,
			"pattern": m.Pattern,
		})
	}
	for _, m := range organizationServer.Mounts {
		logger.Info("GOA HTTP endpoint mounted", logger.Fields{
			"method":  m.Method,
			"verb":    m.Verb,
			"pattern": m.Pattern,
		})
	}
	for _, m := range tenantServer.Mounts {
		logger.Info("GOA HTTP endpoint mounted", logger.Fields{
			"method":  m.Method,
			"verb":    m.Verb,
			"pattern": m.Pattern,
		})
	}
	for _, m := range userServer.Mounts {
		logger.Info("GOA HTTP endpoint mounted", logger.Fields{
			"method":  m.Method,
			"verb":    m.Verb,
			"pattern": m.Pattern,
		})
	}
	for _, m := range openapiServer.Mounts {
		logger.Info("GOA HTTP endpoint mounted", logger.Fields{
			"method":  m.Method,
			"verb":    m.Verb,
			"pattern": m.Pattern,
		})
	}

	// Start server
	logger.Info("Server starting with GOA integration", logger.Fields{
		"port":    cfg.Server.Port,
		"address": ":" + cfg.Server.Port,
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Server.Port,
		Handler:           handler,
		ReadHeaderTimeout: time.Second * 60,
	}

	if err := srv.ListenAndServe(); err != nil {
		logger.Fatal("Server failed to start", logger.Fields{
			"error": err.Error(),
			"port":  cfg.Server.Port,
		})
	}
}
