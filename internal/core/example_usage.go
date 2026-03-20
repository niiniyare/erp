package core

import (
	"context"
	"log"

	db "awo/db/sqlc"
	"awo/internal/platform/cache"
	"awo/internal/platform/temporal"
	"awo/internal/shared/logger"
	"awo/internal/shared/metrics"
	"awo/internal/shared/tracing"
)

// ExampleUsage demonstrates how to use the new ServiceContainer
// This replaces the old cmd/server/services/ initialization pattern
func ExampleUsage() {
	ctx := context.Background()

	// Mock dependencies (in real usage, these would be properly initialized)
	var (
		store    db.Store                = nil // Initialize from database connection
		cache    cache.Service           = nil // Initialize Redis cache
		logger   logger.Logger           = nil // Initialize logger
		metrics  metrics.MetricsProvider = nil // Initialize metrics service
		tracing  tracing.Service         = nil // Initialize tracing
		temporal *temporal.Platform      = nil // Initialize Temporal platform
	)

	// Create dependencies
	deps := Dependencies{
		Store:    store,
		Cache:    cache,
		Logger:   logger,
		Metrics:  metrics,
		Tracing:  tracing,
		Temporal: temporal,
	}

	// Create service container
	serviceContainer, err := NewServiceContainer(deps)
	if err != nil {
		log.Fatalf("Failed to create service container: %v", err)
	}

	// Initialize all services
	if err := serviceContainer.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize services: %v", err)
	}

	// Use services
	tenantService := serviceContainer.GetTenantService()
	abacService := serviceContainer.GetABACService()
	financeService := serviceContainer.GetFinanceService()

	// Services are now ready for use
	_ = tenantService
	_ = abacService
	_ = financeService

	// Graceful shutdown
	defer func() {
		if err := serviceContainer.Shutdown(ctx); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
	}()
}

// Migration Guide from old cmd/server/services/
//
// OLD WAY (cmd/server/services/core.go):
//   services, err := InitializeCoreServices(store, redisClient, logger, metricsService, tracingService)
//
// NEW WAY (internal/core/service.go):
//   deps := core.Dependencies{
//       Store: store,
//       Cache: redisClient,
//       Logger: logger,
//       Metrics: metricsService,
//       Tracing: tracingService,
//       Temporal: temporalPlatform,
//   }
//   serviceContainer, err := core.NewServiceContainer(deps)
//   if err != nil {
//       return err
//   }
//   err = serviceContainer.Initialize(ctx)
//
// BENEFITS:
// - Clear dependency injection
// - Phase-based initialization prevents dependency cycles
// - Better error handling and validation
// - Extensible design for adding new services
// - Proper lifecycle management (init/shutdown)
// - Thread-safe operations
// - Better observability and logging
