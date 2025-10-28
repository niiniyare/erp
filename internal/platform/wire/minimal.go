package wire

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/wire"

	"github.com/niiniyare/erp/internal/platform/config"
)

// ============================================================================
// MINIMAL WIRE SETUP - Start Simple
// ============================================================================

// MinimalProviderSet provides basic dependencies for testing Wire
var MinimalProviderSet = wire.NewSet(
	// Configuration
	config.Load,

	// Simple Fiber app
	NewFiberApp,
)

// NewSimpleFiberApp creates a basic Fiber app for testing
func NewFiberApp(cfg *config.Config) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: cfg.App.Name + " v" + cfg.App.Version,
	})

	// Add a simple health route
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"app":     cfg.App.Name,
			"version": cfg.App.Version,
		})
	})

	return app
}

