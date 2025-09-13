package main

import (
	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/core/finance"
	"github.com/niiniyare/erp/internal/core/finance/repository"
	"github.com/niiniyare/erp/internal/core/finance/service"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/platform/temporal"
	loggerPkg "github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// InitializeFinanceServices creates and initializes all finance services
func InitializeFinanceServices(
	store db.Store,
	cacheService cache.Service,
	logger loggerPkg.Logger,
	metricsService *metrics.MetricsService,
	tracingService tracing.TracingService,
	services *Services,
) (*service.Services, error) {
	logger.Info("Initializing Finance services", loggerPkg.Fields{
		"module": "finance",
		"status": "initializing_services",
	})

	// Create finance repositories
	accountRepo := repository.NewAccountsRepository(store, cacheService, tracingService)
	transactionRepo := repository.NewTransactionRepository(store, tracingService)

	// Create finance service dependencies
	financeServiceDeps := service.Dependencies{
		AccountRepo:        accountRepo,
		TransactionRepo:    transactionRepo,
		Tracing:            tracingService,
		Metrics:            metricsService,
		IAMService:         nil, // Will be initialized later when IAM module is complete
		FeatureFlagService: services.FeatureFlagService,
	}

	// Validate dependencies
	if err := financeServiceDeps.Validate(); err != nil {
		logger.Error("Finance service dependencies validation failed", loggerPkg.Fields{
			"error": err.Error(),
		})
		return nil, err
	}

	// Create finance services
	financeServices := service.NewServices(financeServiceDeps)

	logger.Info("✅ Finance services initialized successfully", loggerPkg.Fields{
		"module":   "finance",
		"services": []string{"account", "transaction", "transaction_entry"},
		"status":   "ready",
	})

	return financeServices, nil
}

// RegisterFinanceModule registers the finance module with the Temporal platform
func RegisterFinanceModule(
	temporalPlatform *temporal.Platform,
	services *Services,
	financeServices *service.Services,
	cacheService cache.Service,
	logger loggerPkg.Logger,
	metricsService *metrics.MetricsService,
	tracingService tracing.TracingService,
) error {
	logger.Info("Registering Finance module with Temporal platform", loggerPkg.Fields{
		"module": "finance",
		"status": "initializing_temporal_integration",
	})

	// Create temporal integration configuration with proper services
	temporalConfig := finance.TemporalIntegrationConfig{
		Services:            financeServices,
		IAMService:          nil, // Will be available when IAM module is complete
		AuditService:        services.AuditService,
		FeatureFlagService:  services.FeatureFlagService,
		SettingsService:     nil, // Will be available when settings module is complete
		NotificationService: services.NotificationService,
		CacheService:        cacheService,
		TemporalClient:      temporalPlatform.GetClient(),
		Logger:              logger,
		Metrics:             metricsService,
		Tracer:              tracingService,
	}

	// Create temporal integration
	temporalIntegration, err := finance.NewTemporalIntegration(temporalConfig)
	if err != nil {
		logger.Error("Failed to create finance Temporal integration", loggerPkg.Fields{
			"error": err.Error(),
		})
		return err
	}

	// Register finance module with Temporal platform
	if err := temporalIntegration.RegisterWithPlatform(temporalPlatform); err != nil {
		logger.Error("Failed to register finance module with Temporal platform", loggerPkg.Fields{
			"error": err.Error(),
		})
		return err
	}

	logger.Info("✅ Finance module successfully registered with Temporal platform", loggerPkg.Fields{
		"module":             "finance",
		"status":             "registered",
		"activities_count":   "40+",
		"workflows_count":    "11",
		"integration_status": "complete",
	})

	return nil
}
