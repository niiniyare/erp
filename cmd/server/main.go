package main

import (
	"context"
	"net/http"
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
	services, err := InitializeServices(database.Store, database.RedisClient, infra.Metrics, infra.Tracing)
	if err != nil {
		logger.Fatal("Failed to initialize services", logger.Fields{"error": err})
	}

	// Initialize GOA server
	goaServer, err := InitializeGOAServer(services, infra.Metrics, infra.Tracing)
	if err != nil {
		logger.Fatal("Failed to initialize GOA server", logger.Fields{"error": err})
	}

	// Initialize Gin router for migration testing
	ginRouter, err := InitializeGinRouter(services, infra.Metrics, infra.Tracing)
	if err != nil {
		logger.Fatal("Failed to initialize Gin router", logger.Fields{"error": err})
	}

	// Create combined handler for migration phase
	combinedHandler := &CombinedHandler{
		goaHandler: goaServer.Handler,
		ginHandler: ginRouter,
	}

	// Start HTTP server
	logger.Info("Server starting with GOA+Gin (migration mode)", logger.Fields{
		"port":    infra.Config.Server.Port,
		"address": ":" + infra.Config.Server.Port,
	})

	srv := &http.Server{
		Addr:              ":" + infra.Config.Server.Port,
		Handler:           combinedHandler,
		ReadHeaderTimeout: time.Second * 60,
	}

	if err := srv.ListenAndServe(); err != nil {
		logger.Fatal("Server failed to start", logger.Fields{
			"error": err.Error(),
			"port":  infra.Config.Server.Port,
		})
	}
}
