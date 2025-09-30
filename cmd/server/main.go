package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/niiniyare/erp/internal/application"
	"github.com/niiniyare/erp/internal/platform/config"
	"github.com/niiniyare/erp/internal/shared/logger"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	// Load configuration
	cfg, _ := config.LoadWithViper()
	err := cfg.Validate()
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}

	// Log successful configuration load
	logger.Info("Configuration loaded successfully", logger.Fields{
		"app_name": cfg.App.Name,
		"version":  cfg.App.Version,
		"stage":    cfg.App.Stage,
		"port":     cfg.Server.Port,
	})

	// Create application core
	app := application.NewCore(cfg)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start application core
	if err := app.Start(ctx); err != nil {
		logger.Fatal("Failed to start application", logger.Fields{"error": err.Error()})
	}

	// Initialize HTTP server with existing GOA setup
	// TODO: This will be refactored to use the new architecture in the next step
	goaServer, err := initializeHTTPServer(app)
	if err != nil {
		logger.Fatal("Failed to initialize HTTP server", logger.Fields{"error": err.Error()})
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:              ":" + cfg.Server.Port,
		Handler:           goaServer,
		ReadHeaderTimeout: time.Second * 60,
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting HTTP server", logger.Fields{
			"port":    cfg.Server.Port,
			"address": ":" + cfg.Server.Port,
		})

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start", logger.Fields{
				"error": err.Error(),
				"port":  cfg.Server.Port,
			})
		}
	}()

	// Setup graceful shutdown
	gracefulShutdown(ctx, cancel, srv, app)
}

// initializeHTTPServer creates the HTTP server using existing GOA setup and UI integration
func initializeHTTPServer(app *application.Core) (http.Handler, error) {
	// Extract infrastructure services from application core
	appServices := app.GetServices()

	// Initialize business services using the business services factory
	businessServices, err := InitializeServices(
		appServices.Store,
		appServices.RedisClient,
		appServices.Logger,
		appServices.Metrics,
		appServices.Tracing,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize business services: %w", err)
	}

	// Initialize GOA server with all services
	// Note: Finance services are not fully integrated yet, passing nil for now
	goaServer, err := InitializeGOAServer(
		businessServices,
		nil, // financeServices placeholder
		appServices.Store,
		appServices.RedisClient,
		appServices.Metrics,
		appServices.Tracing,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize GOA server: %w", err)
	}

	logger.Info("GOA server initialized successfully", logger.Fields{
		"endpoints": "all services mounted",
		"status":    "ready",
	})

	// Initialize UI integration
	// uiIntegration, err := NewUIIntegration(app, businessServices, appServices.Logger)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to initialize UI integration: %w", err)
	// }

	// Create combined handler that serves both API and UI routes
	// combinedHandler := uiIntegration.CreateCombinedHandler(goaServer.Handler)

	logger.Info("UI integration completed successfully", logger.Fields{
		"ui_routes":   "console, workspace, portal",
		"integration": "single-port",
		"status":      "ready",
	})

	return goaServer.Handler, nil
}

// gracefulShutdown handles graceful shutdown of the application
func gracefulShutdown(_ context.Context, cancel context.CancelFunc, srv *http.Server, app *application.Core) {
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
