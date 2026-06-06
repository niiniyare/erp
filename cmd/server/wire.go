//go:build wireinject
// +build wireinject

// Package main provides Wire dependency injection for the Awo ERP server
package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/wire"

	"awo.so/internal/api/handlers"
	"awo.so/internal/platform/config"
	"awo.so/internal/platform/temporal"
	wirepkg "awo.so/internal/platform/wire"
)

// ============================================================================
// APPLICATION INJECTOR
// ============================================================================

// Application represents the fully wired application
type Application struct {
	Config           *config.Config
	App              *fiber.App
	Router           *handlers.Router
	TemporalPlatform *temporal.Platform
}

// InitializeApplication creates a fully configured application with all dependencies
func InitializeApplication() (*Application, error) {
	wire.Build(
		// All provider sets
		wirepkg.ApplicationProviderSet,

		// Application struct constructor
		NewApplication,
	)
	return &Application{}, nil
}

// NewApplication creates a new application instance
func NewApplication(
	cfg *config.Config,
	app *fiber.App,
	router *handlers.Router,
	temporalPlatform *temporal.Platform,
) *Application {
	return &Application{
		Config:           cfg,
		App:              app,
		Router:           router,
		TemporalPlatform: temporalPlatform,
	}
}
