package main

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/adapters"
	"github.com/niiniyare/erp/internal/core/abac"
	"github.com/niiniyare/erp/internal/core/abac/repository"
	"github.com/niiniyare/erp/internal/core/access/approval"
	"github.com/niiniyare/erp/internal/core/access/conditional"
	"github.com/niiniyare/erp/internal/core/access/execution"
	"github.com/niiniyare/erp/internal/core/access/request"
	"github.com/niiniyare/erp/internal/core/analytics"
	"github.com/niiniyare/erp/internal/core/audit"
	"github.com/niiniyare/erp/internal/core/entity"
	"github.com/niiniyare/erp/internal/core/featureflag"
	"github.com/niiniyare/erp/internal/core/finance"
	financeRepo "github.com/niiniyare/erp/internal/core/finance/repository"
	financeService "github.com/niiniyare/erp/internal/core/finance/service"
	"github.com/niiniyare/erp/internal/core/identity"
	"github.com/niiniyare/erp/internal/core/notification"
	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/platform/config"
	"github.com/niiniyare/erp/internal/platform/temporal"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// bootstrap initializes the application by setting up the configuration, logger,
// database, services, and other dependencies. It returns a fully configured
// application struct or an error if any step fails.
func bootstrap() (*application, error) {
	fmt.Println("bootstrap: start")
	// Load configuration using the structured loader.
	cfg := config.Load()
	fmt.Println("bootstrap: config loaded")

	// Initialize the logger from the configuration.
	if err := logger.InitializeFromEnv(); err != nil { // Note: Logger initialization might be special
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}
	fmt.Println("bootstrap: logger initialized")
	log := logger.WithFields(logger.Fields{
		"service": cfg.App.Name,
		"version": cfg.App.Version,
	})
	log.Info("Logger initialized")
	log.Info("Configuration loaded", logger.Fields{"stage": cfg.App.Stage, "environment": cfg.App.Environment})
	fmt.Println("bootstrap: logger configured")

	// Initialize tracing service.
	tracer, err := tracing.NewTracingService(tracing.TracingConfig{
		ServiceName:    cfg.App.Name,
		ServiceVersion: cfg.App.Version,
		Environment:    cfg.App.Environment,
		ExporterType:   tracing.StdoutExporter,
		SamplingRatio:  1.0,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tracing service: %w", err)
	}
	log.Info("Tracing service initialized")
	fmt.Println("bootstrap: tracing service initialized")

	// Initialize metrics service.
	metric, err := metrics.NewMetricsService(metrics.MetricsConfig{
		Namespace: cfg.App.Environment,
		Subsystem: "server",
		Provider:  "prometheus",
		Enabled:   true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize metrics service: %w", err)
	}
	log.Info("Metrics service initialized")
	fmt.Println("bootstrap: metrics service initialized")

	// Initialize and test database connection.
	dbURL := cfg.Database.GetDatabaseURL()
	fmt.Println("bootstrap: db url: ", dbURL)
	store, err := db.NewDB(dbURL)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}
	log.Info("Database connection established")
	fmt.Println("bootstrap: db connection established")

	// Initialize and test Redis connection.
	redisConfig := cache.DefaultRedisConfig(&cfg.Redis)
	fmt.Println("bootstrap: redis config loaded")
	redisClient := cache.NewRedisClient(redisConfig)
	fmt.Println("bootstrap: redis client created")
	if err := redisClient.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}
	log.Info("Redis connection established")
	fmt.Println("bootstrap: redis connection established")

	// Run database migrations using settings from config.
	fmt.Println("bootstrap: running db migration")
	if err := runDBMigration(cfg.Migration, dbURL, log); err != nil {
		return nil, fmt.Errorf("database migration failed: %w", err)
	}
	fmt.Println("bootstrap: db migration complete")

	// Initialize Temporal platform.
	fmt.Println("bootstrap: initializing temporal platform")
	temporalPlatform, err := temporal.NewPlatform(&cfg.Temporal, log)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize temporal platform: %w", err)
	}
	log.Info("Temporal platform initialized")
	fmt.Println("bootstrap: temporal platform initialized")

	// Initialize all application services.
	fmt.Println("bootstrap: initializing services")
	appServices, err := initializeServices(store, redisClient, log, metric, tracer)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}
	log.Info("Application services initialized")
	fmt.Println("bootstrap: services initialized")

	// Initialize all finance services.
	fmt.Println("bootstrap: initializing finance services")
	finServices, err := initializeFinanceServices(store, redisClient, log, metric, tracer, appServices)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize finance services: %w", err)
	}
	log.Info("Finance services initialized")
	fmt.Println("bootstrap: finance services initialized")

	// Register finance module with Temporal.
	fmt.Println("bootstrap: registering finance module")
	err = registerFinanceModule(temporalPlatform, appServices, finServices, redisClient, log, metric, tracer)
	if err != nil {
		return nil, fmt.Errorf("failed to register finance module with temporal: %w", err)
	}
	fmt.Println("bootstrap: finance module registered")

	// Build the application dependency container.
	fmt.Println("bootstrap: building application container")
	app := &application{
		config:          cfg,
		logger:          log,
		store:           store,
		redis:           redisClient,
		temporal:        temporalPlatform,
		tracer:          tracer,
		metrics:         metric,
		services:        appServices,
		financeServices: finServices,
	}

	log.Info("Application bootstrap complete")
	fmt.Println("bootstrap: complete")
	return app, nil
}


// shutdown handles the graceful shutdown of application components.
func (app *application) shutdown() {
	app.logger.Info("Shutting down application")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if app.temporal != nil {
		if err := app.temporal.Stop(ctx); err != nil {
			app.logger.Error("Error stopping Temporal platform", logger.Fields{"error": err})
		}
	}
	if app.tracer != nil {
		if err := app.tracer.Shutdown(ctx); err != nil {
			app.logger.Error("Error shutting down tracer", logger.Fields{"error": err})
		}
	}
	if app.redis != nil {
		if err := app.redis.Close(); err != nil {
			app.logger.Error("Error closing Redis connection", logger.Fields{"error": err})
		}
	}
	if app.store != nil {
		app.store.Close()
	}
	app.logger.Info("Shutdown complete")
}

// runDBMigration executes database migrations.
func runDBMigration(cfg config.MigrationConfig, dbSource string, log logger.Logger) error {
	startTime := time.Now()
	log = log.WithFields(logger.Fields{"operation": "database_migration"})
	log.Info("Starting database migration", logger.Fields{"url": cfg.URL})

	migration, err := migrate.New(cfg.URL, dbSource)
	if err != nil {
		return fmt.Errorf("cannot create new migrate instance: %w", err)
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrate up: %w", err)
	}

	if err == migrate.ErrNoChange {
		log.Info("No new migrations to apply")
	} else {
		log.Info("Database migrated successfully", logger.Fields{"duration": time.Since(startTime).String()})
	}

	return nil
}

// initializeServices sets up the core application services.
func initializeServices(store db.Store, redisClient cache.Service, logger logger.Logger, metricsService *metrics.MetricsService, tracingService tracing.TracingService) (*Services, error) {
	// Repositories
	tenantRepo := tenant.NewRepository(store, tracingService)
	entityRepo := entity.NewRepository(store, tracingService, metricsService)
	identityRepo := identity.NewRepository(store, tracingService, metricsService)
	auditRepo := audit.NewRepository(store, logger, tracingService, metricsService)
	notificationRepo := notification.NewRepository(store)
	policyRepo := repository.NewPolicyRepository(store, redisClient, logger, metricsService, tracingService)
	attributeRepo := repository.NewAttributeRepository(store, redisClient, logger, metricsService, tracingService)
	policyEvaluationRepo := repository.NewPolicyEvaluationRepository(store, redisClient, logger, metricsService, tracingService)
	featureFlagRepo := featureflag.NewRepository(store)

	// Core Services
	tenantService := tenant.NewService(tenantRepo, redisClient, tracingService)
	entityService := entity.NewService(entityRepo, tracingService, metricsService)
	identityService := identity.NewService(identityRepo, redisClient, tracingService, metricsService)
	auditService := audit.NewService(auditRepo, redisClient, logger, tracingService, metricsService)
	abacService := abac.NewService(policyRepo, attributeRepo, policyEvaluationRepo, identityService, tenantService, logger, metricsService, tracingService)

	// Feature Flag Service (with caching)
	webSocketService := featureflag.NewWebSocketService(store, tracingService, metricsService)
	baseFeatureFlagService := featureflag.NewService(featureFlagRepo, tenantService, store, auditService, webSocketService)
	featureFlagService := featureflag.NewCachedFeatureFlagService(baseFeatureFlagService, redisClient, logger, metricsService, tracingService)
	adminFeatureFlagService := featureflag.NewAdminService(featureFlagService, auditService, logger, metricsService, tracingService, nil)

	// Access Services
	approverService := approval.NewApproverService(identityService, tracingService, metricsService)
	notificationService := notification.NewNotificationService(tracingService, metricsService, nil, nil, notificationRepo)
	executionService := execution.NewAccessExecutionService(identityRepo, identityService, tracingService, metricsService)
	userServiceAdapter := adapters.NewUserServiceAdapter(identityService)
	auditServiceAdapter := adapters.NewAuditServiceAdapter(auditService)
	accessRequestService := request.NewAccessRequestService(nil, redisClient, tracingService, metricsService, userServiceAdapter, notificationService, approverService, executionService, auditService)
	conditionalAccessService := conditional.NewConditionalAccessService(tracingService, metricsService, auditServiceAdapter)
	analyticsService := analytics.NewUserAnalyticsService(tracingService, metricsService, auditService)

	return &Services{
		TenantService:            tenantService,
		EntityService:            entityService,
		IdentityService:          identityService,
		ABACService:              abacService,
		AuditService:             auditService,
		FeatureFlagService:       featureFlagService,
		AdminFeatureFlagService:  adminFeatureFlagService,
		AccessRequestService:     accessRequestService,
		ConditionalAccessService: conditionalAccessService,
		AnalyticsService:         analyticsService,
		NotificationService:      notificationService,
		ApproverService:          approverService,
		ExecutionService:         executionService,
	}, nil
}

// initializeFinanceServices sets up the finance-specific services.
func initializeFinanceServices(store db.Store, cacheService cache.Service, logger logger.Logger, metricsService *metrics.MetricsService, tracingService tracing.TracingService, services *Services) (*financeService.Services, error) {
	accountRepo := financeRepo.NewAccountsRepository(store, cacheService, tracingService)
	transactionRepo := financeRepo.NewTransactionRepository(store, tracingService)

	deps := financeService.Dependencies{
		AccountRepo:        accountRepo,
		TransactionRepo:    transactionRepo,
		Tracing:            tracingService,
		Metrics:            metricsService,
		FeatureFlagService: services.FeatureFlagService,
	}
	if err := deps.Validate(); err != nil {
		return nil, fmt.Errorf("finance service dependencies validation failed: %w", err)
	}

	return financeService.NewServices(deps), nil
}

// registerFinanceModule registers the finance module with the Temporal platform.
func registerFinanceModule(temporalPlatform *temporal.Platform, services *Services, financeServices *financeService.Services, cacheService cache.Service, logger logger.Logger, metricsService *metrics.MetricsService, tracingService tracing.TracingService) error {
	cfg := finance.TemporalIntegrationConfig{
		Services:            financeServices,
		AuditService:        services.AuditService,
		FeatureFlagService:  services.FeatureFlagService,
		NotificationService: services.NotificationService,
		CacheService:        cacheService,
		TemporalClient:      temporalPlatform.GetClient(),
		Logger:              logger,
		Metrics:             metricsService,
		Tracer:              tracingService,
	}

	integration, err := finance.NewTemporalIntegration(cfg)
	if err != nil {
		return fmt.Errorf("failed to create finance Temporal integration: %w", err)
	}

	if err := integration.RegisterWithPlatform(temporalPlatform); err != nil {
		return fmt.Errorf("failed to register finance module with Temporal platform: %w", err)
	}

	logger.Info("Finance module successfully registered with Temporal platform")
	return nil
}
