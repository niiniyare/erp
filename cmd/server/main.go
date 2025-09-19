package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/niiniyare/erp/internal/application"
	"github.com/niiniyare/erp/internal/config"
	"github.com/niiniyare/erp/internal/shared/logger"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
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

// initializeHTTPServer creates the HTTP server using existing GOA setup
// This is a temporary bridge function that will be refactored
func initializeHTTPServer(app *application.Core) (http.Handler, error) {
	// Extract services from application core
	services := app.GetServices()
	cfg := app.GetConfig()

	// For now, we'll use a simple handler
	// TODO: Integrate with existing GOA server setup
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Health check endpoint
		if r.URL.Path == "/health" {
			ctx := r.Context()
			if err := app.Health(ctx); err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte(`{"status":"unhealthy","error":"` + err.Error() + `"}`))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"healthy"}`))
			return
		}

		// Placeholder response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Simple JSON response (avoiding external deps for now)
		w.Write([]byte(`{
			"app": "` + cfg.App.Name + `",
			"version": "` + cfg.App.Version + `",
			"status": "running",
			"message": "New architecture working! (placeholder)"
		}`))

		// Log the request
		logger.Info("HTTP request handled", logger.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
			"status": 200,
		})

		// Use services to show they're working
		_ = services.Store       // Database connection
		_ = services.RedisClient // Cache connection
		// More integration will be added in next phase
	}), nil
}

// gracefulShutdown handles graceful shutdown of the application
func gracefulShutdown(ctx context.Context, cancel context.CancelFunc, srv *http.Server, app *application.Core) {
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
