package main

import (
	"context"
	"net/http"

	"github.com/niiniyare/erp/internal/api/handlers"
	"github.com/niiniyare/erp/internal/core/entity"
	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/core/user"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/platform/config"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"

	db "github.com/niiniyare/erp/db/sqlc"
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
		Namespace: "erp",
		Subsystem: "server",
		Provider:  "prometheus",
		Enabled:   true,
	})
	if err != nil {
		logger.Fatal("Failed to initialize metrics", logger.Fields{"error": err})
	}

	// Build database URL from config
	databaseURL := cfg.Database.GetDatabaseURL()

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
	redisClient := cache.NewRedisClient(&cfg.Redis)

	logger.Info("Cache client initialized", logger.Fields{
		"redis_host": cfg.Redis.Host,
		"redis_port": cfg.Redis.Port,
	})

	// Initialize repositories
	tenantRepo := tenant.NewRepository(store)
	entityRepo := entity.NewRepository(store, tracingService, metricsService)
	userRepo := user.NewRepository(store, tracingService, metricsService)

	// Initialize services
	tenantService := tenant.NewService(tenantRepo, redisClient)
	entityService := entity.NewService(entityRepo, tracingService, metricsService)
	userService := user.NewService(userRepo, redisClient, tracingService, metricsService)

	// Initialize API handlers
	router := handlers.NewRouter(tenantService, entityService, userService, tracingService, metricsService)

	// Start server
	logger.Info("Server starting", logger.Fields{
		"port":    cfg.Server.Port,
		"address": ":" + cfg.Server.Port,
	})

	if err := http.ListenAndServe(":"+cfg.Server.Port, router); err != nil {
		logger.Fatal("Server failed to start", logger.Fields{
			"error": err.Error(),
			"port":  cfg.Server.Port,
		})
	}
}
