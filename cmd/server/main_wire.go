// Package main provides the main entry point for the Awo ERP server
// This version uses Google Wire for dependency injection
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/niiniyare/erp/internal/platform/config"
)

func main() {
	// Initialize application with all dependencies using Wire
	app, err := InitializeApplication()
	if err != nil {
		log.Fatal("Failed to initialize application:", err)
	}

	// Start the server
	if err := startServer(app); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// startServer starts the HTTP server with graceful shutdown
func startServer(app *Application) error {
	// Register all routes using the Wire-injected router
	if err := app.Router.RegisterAll(app.App); err != nil {
		return fmt.Errorf("failed to register routes: %w", err)
	}

	// Print registered routes for debugging
	app.Router.PrintRoutes()

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Channel to listen for interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	serverAddr := fmt.Sprintf("%s:%s", "0.0.0.0", app.Config.Server.Port)

	go func() {
		fmt.Printf("🚀 Server starting on %s\n", serverAddr)
		fmt.Printf("📱 Environment: %s\n", app.Config.App.Environment)
		fmt.Printf("🏢 Application: %s v%s\n", app.Config.App.Name, app.Config.App.Version)
		fmt.Printf("🔧 Debug mode: %v\n", app.Config.App.Debug)
		fmt.Printf("📊 Health check: http://%s/health\n", serverAddr)
		fmt.Printf("📚 Documentation: http://%s/docs\n", serverAddr)

		if err := app.App.Listen(serverAddr); err != nil {
			log.Printf("Server error: %v", err)
			cancel()
		}
	}()

	// Wait for interrupt signal or context cancellation
	select {
	case sig := <-sigChan:
		fmt.Printf("\n🛑 Received signal: %v\n", sig)
	case <-ctx.Done():
		fmt.Println("\n🛑 Server context cancelled")
	}

	// Graceful shutdown
	fmt.Println("🔄 Initiating graceful shutdown...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := app.App.ShutdownWithContext(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	fmt.Println("✅ Server shutdown completed")
	return nil
}

// Development and testing helpers

// InitializeForTesting creates a minimal application for testing
func InitializeForTesting() (*Application, error) {
	// Override configuration for testing
	os.Setenv("APP_STAGE", "testing")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_NAME", "erp_test")
	os.Setenv("REDIS_HOST", "localhost")

	return InitializeApplication()
}

// InitializeForDevelopment creates an application with development settings
func InitializeForDevelopment() (*Application, error) {
	// Set development-specific environment variables
	os.Setenv("APP_STAGE", "development")
	os.Setenv("APP_DEBUG", "true")
	os.Setenv("LOG_LEVEL", "debug")

	return InitializeApplication()
}

// validateEnvironment checks that required environment variables are set
func validateEnvironment() error {
	required := []string{
		"DB_HOST",
		"DB_USER",
		"DB_PASSWORD",
		"DB_NAME",
	}

	var missing []string
	for _, env := range required {
		if os.Getenv(env) == "" {
			missing = append(missing, env)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %v", missing)
	}

	return nil
}

// healthCheck performs basic application health checks
func healthCheck(app *Application) error {
	// Check all application components
	if app == nil || app.Config == nil || app.App == nil || app.Router == nil {
		return fmt.Errorf("application is not properly initialized")
	}

	// Check router health
	if err := app.Router.HealthCheck(); err != nil {
		return fmt.Errorf("router health check failed: %w", err)
	}

	// Add more health checks as needed
	// - Database connectivity
	// - Redis connectivity
	// - External service dependencies

	return nil
}

// printBanner displays the application startup banner
func printBanner(cfg *config.Config) {
	fmt.Println(`
    ██████╗ ██╗    ██╗ ██████╗     ███████╗██████╗ ██████╗ 
   ██╔══██╗██║    ██║██╔═══██╗    ██╔════╝██╔══██╗██╔══██╗
   ███████║██║ █╗ ██║██║   ██║    █████╗  ██████╔╝██████╔╝
   ██╔══██║██║███╗██║██║   ██║    ██╔══╝  ██╔══██╗██╔═══╝ 
   ██║  ██║╚███╔███╔╝╚██████╔╝    ███████╗██║  ██║██║     
   ╚═╝  ╚═╝ ╚══╝╚══╝  ╚═════╝     ╚══════╝╚═╝  ╚═╝╚═╝     
                                                           
   Enterprise Resource Planning System
   Multi-Tenant • Secure • Scalable
   `)

	fmt.Printf("   Version: %s | Environment: %s | Stage: %s\n\n",
		cfg.App.Version, cfg.App.Environment, cfg.App.Stage)
}
