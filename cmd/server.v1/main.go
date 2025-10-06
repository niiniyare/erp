package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/niiniyare/erp/cmd/server/bootstrap"
	"github.com/niiniyare/erp/cmd/server/server"
	"github.com/niiniyare/erp/cmd/server/services"
	"github.com/niiniyare/erp/internal/shared/logger"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	// Initialize application
	app, err := bootstrap.NewApplication()
	if err != nil {
		logger.Fatal("Failed to initialize application", logger.Fields{"error": err.Error()})
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start application core
	if err := app.Start(ctx); err != nil {
		logger.Fatal("Failed to start application", logger.Fields{"error": err.Error()})
	}

	// Initialize dependencies
	deps, err := bootstrap.InitializeDependencies(app)
	if err != nil {
		logger.Fatal("Failed to initialize dependencies", logger.Fields{"error": err.Error()})
	}

	// Initialize core business services
	coreServices, err := services.InitializeCoreServices(
		deps.Store,
		deps.RedisClient,
		deps.Logger,
		deps.Metrics,
		deps.Tracing,
	)
	if err != nil {
		logger.Fatal("Failed to initialize core services", logger.Fields{"error": err.Error()})
	}

	// Initialize finance services (optional)
	financeServices, err := services.InitializeFinanceServices(
		deps.Store,
		deps.RedisClient,
		deps.Logger,
		deps.Metrics,
		deps.Tracing,
		coreServices,
	)
	if err != nil {
		logger.Warn("Finance services initialization failed, continuing without them", logger.Fields{"error": err.Error()})
		financeServices = nil
	}

	// Initialize GOA server
	goaServer, err := server.InitializeGOAServer(
		coreServices,
		financeServices,
		deps.Store,
		deps.RedisClient,
		deps.Metrics,
		deps.Tracing,
	)
	if err != nil {
		logger.Fatal("Failed to initialize GOA server", logger.Fields{"error": err.Error()})
	}

	// Create HTTP server
	srv := app.CreateHTTPServer(goaServer.Handler)

	// Start server in a goroutine
	go func() {
		logger.Info("Starting HTTP server", logger.Fields{
			"port":    app.Config.Server.Port,
			"address": ":" + app.Config.Server.Port,
		})

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start", logger.Fields{
				"error": err.Error(),
				"port":  app.Config.Server.Port,
			})
		}
	}()

	// Setup graceful shutdown
	gracefulShutdown(ctx, cancel, srv, app)
}

// gracefulShutdown handles graceful shutdown of the application
func gracefulShutdown(_ context.Context, cancel context.CancelFunc, srv *http.Server, app *bootstrap.Application) {
	// Create a channel to receive OS signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive a signal
	<-quit
	logger.Info("Received shutdown signal, starting graceful shutdown...")

	// Cancel the main context
	cancel()

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Failed to shutdown HTTP server gracefully", logger.Fields{
			"error": err.Error(),
		})
	} else {
		logger.Info("HTTP server stopped")
	}

	// Shutdown application core
	if err := app.Stop(shutdownCtx); err != nil {
		logger.Error("Failed to shutdown application gracefully", logger.Fields{
			"error": err.Error(),
		})
	} else {
		logger.Info("Application stopped successfully")
	}

	logger.Info("Graceful shutdown completed")
}
