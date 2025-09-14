package main

import (
	"github.com/niiniyare/erp/internal/shared/logger"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	// 1. Bootstrap the application: load config, initialize logger, DB, services, etc.
	app, err := bootstrap()
	if err != nil {
		// Use a basic logger or panic if the main logger failed to initialize.
		logger.Fatal("Failed to bootstrap application", logger.Fields{"error": err})
		return
	}
	// Schedule a graceful shutdown of all components.
	defer app.shutdown()

	// 2. Create the HTTP router and wire up the handlers with services.
	handler := newRouter(app)

	// 3. Start the HTTP server and block until it's shut down.
	if err := serve(app, handler); err != nil {
		app.logger.Fatal("Server failed to start", logger.Fields{"error": err})
	}
}