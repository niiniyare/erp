package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/niiniyare/erp/internal/shared/logger"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	// Initialize infrastructure (logger, tracing, metrics, config)
	infra, err := InitializeInfrastructure()
	if err != nil {
		panic("Failed to initialize infrastructure: " + err.Error())
	}
	defer infra.Shutdown(context.Background())

	// Initialize database and cache
	database, err := InitializeDatabase(infra.Config)
	if err != nil {
		logger.Fatal("Failed to initialize database", logger.Fields{"error": err})
	}
	defer database.Close()

	// Initialize all services
	services, err := InitializeServices(database.Store, database.RedisClient, infra.Logger, infra.Metrics, infra.Tracing)
	if err != nil {
		logger.Fatal("Failed to initialize services", logger.Fields{"error": err})
	}

	// Initialize finance services
	financeServices, err := InitializeFinanceServices(database.Store, database.RedisClient, infra.Logger, infra.Metrics, infra.Tracing, services)
	if err != nil {
		logger.Fatal("Failed to initialize finance services", logger.Fields{"error": err})
	}

	// Register finance module with Temporal platform
	if err := RegisterFinanceModule(infra.Temporal, services, financeServices, database.RedisClient, infra.Logger, infra.Metrics, infra.Tracing); err != nil {
		logger.Fatal("Failed to register finance module with Temporal", logger.Fields{"error": err})
	}

	// Initialize GOA server
	goaServer, err := InitializeGOAServer(services, database.Store, database.RedisClient, infra.Metrics, infra.Tracing)
	if err != nil {
		logger.Fatal("Failed to initialize GOA server", logger.Fields{"error": err})
	}

	// Determine server mode from environment
	serverMode := getServerMode()
	var handler http.Handler

	switch serverMode {
	case "goa-only":
		// Production mode: Pure GOA server
		handler = goaServer.Handler
		logger.Info("Server starting in GOA-only mode (production)", logger.Fields{
			"port":    infra.Config.Server.Port,
			"address": ":" + infra.Config.Server.Port,
			"mode":    "goa-only",
		})

	case "migration":
		// Migration mode: Combined handler (fallback option)
		ginRouter, err := InitializeGinRouter(services, infra.Metrics, infra.Tracing)
		if err != nil {
			logger.Fatal("Failed to initialize Gin router", logger.Fields{"error": err})
		}

		handler = &CombinedHandler{
			goaHandler: goaServer.Handler,
			ginHandler: ginRouter,
		}
		logger.Info("Server starting in migration mode (GOA+Gin)", logger.Fields{
			"port":    infra.Config.Server.Port,
			"address": ":" + infra.Config.Server.Port,
			"mode":    "migration",
		})

	default:
		logger.Fatal("Invalid server mode", logger.Fields{
			"mode":         serverMode,
			"supported":    "goa-only, migration",
			"default_mode": "goa-only",
			"env_variable": "SERVER_MODE",
		})
	}

	// Start HTTP server
	srv := &http.Server{
		Addr:              ":" + infra.Config.Server.Port,
		Handler:           handler,
		ReadHeaderTimeout: time.Second * 60,
		ReadTimeout:       infra.Config.Server.ReadTimeout,
		WriteTimeout:      infra.Config.Server.WriteTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		logger.Fatal("Server failed to start", logger.Fields{
			"error": err.Error(),
			"port":  infra.Config.Server.Port,
			"mode":  serverMode,
		})
	}
}

// getServerMode determines the server mode from environment variables
func getServerMode() string {
	mode := os.Getenv("SERVER_MODE")
	if mode == "" {
		return "goa-only" // Default to production GOA-only mode
	}
	return mode
}
