package application

import (
	"context"
	"fmt"
	"sync"

	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/config"
	"github.com/niiniyare/erp/internal/infrastructure"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/platform/temporal"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// Core represents the application core that orchestrates all components
type Core struct {
	config         *config.AppConfig
	infrastructure *infrastructure.Container
	services       *Services
	financeServices *FinanceServices
	started        bool
	mu             sync.RWMutex
}

// Services represents the core business services
type Services struct {
	// These will be populated from cmd/server/services.go patterns
	Store            db.Store
	RedisClient      cache.Service
	Logger           logger.Logger
	Metrics          *metrics.MetricsService
	Tracing          tracing.TracingService
	TemporalPlatform *temporal.Platform
}

// FinanceServices represents finance-specific services
type FinanceServices struct {
	// This will be populated from cmd/server/temporal_integration.go patterns
	// We'll populate this when we integrate the finance module
}

// NewCore creates a new application core
func NewCore(cfg *config.AppConfig) *Core {
	return &Core{
		config:         cfg,
		infrastructure: infrastructure.NewContainer(cfg),
	}
}

// Start initializes and starts the application core
func (c *Core) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.started {
		return fmt.Errorf("application core already started")
	}

	logger.Info("Starting application core", logger.Fields{
		"app_name": c.config.App.Name,
		"version":  c.config.App.Version,
		"stage":    c.config.App.Stage,
	})

	// Start infrastructure first
	if err := c.infrastructure.Start(ctx); err != nil {
		return fmt.Errorf("failed to start infrastructure: %w", err)
	}

	// Initialize services from infrastructure components
	if err := c.initializeServices(); err != nil {
		return fmt.Errorf("failed to initialize services: %w", err)
	}

	// Initialize finance services (when ready)
	if err := c.initializeFinanceServices(); err != nil {
		return fmt.Errorf("failed to initialize finance services: %w", err)
	}

	c.started = true

	logger.Info("Application core started successfully", logger.Fields{
		"status": "ready",
		"services_initialized": true,
	})

	return nil
}

// Stop gracefully shuts down the application core
func (c *Core) Stop(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.started {
		return nil
	}

	logger.Info("Stopping application core")

	// Stop infrastructure (handles all cleanup)
	if err := c.infrastructure.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop infrastructure: %w", err)
	}

	c.started = false

	logger.Info("Application core stopped successfully")
	return nil
}

// Health checks the health of the entire application
func (c *Core) Health(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.started {
		return fmt.Errorf("application core not started")
	}

	// Check infrastructure health
	if err := c.infrastructure.Health(ctx); err != nil {
		return fmt.Errorf("infrastructure health check failed: %w", err)
	}

	// Additional service-level health checks could go here

	return nil
}

// GetConfig returns the application configuration
func (c *Core) GetConfig() *config.AppConfig {
	return c.config
}

// GetInfrastructure returns the infrastructure container
func (c *Core) GetInfrastructure() *infrastructure.Container {
	return c.infrastructure
}

// GetServices returns the core services
func (c *Core) GetServices() *Services {
	return c.services
}

// GetFinanceServices returns the finance services
func (c *Core) GetFinanceServices() *FinanceServices {
	return c.financeServices
}

// IsStarted returns whether the core is started
func (c *Core) IsStarted() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.started
}

// initializeServices creates the core services from infrastructure components
func (c *Core) initializeServices() error {
	observability := c.infrastructure.GetObservability()
	database := c.infrastructure.GetDatabase()
	cache := c.infrastructure.GetCache()
	messaging := c.infrastructure.GetMessaging()

	c.services = &Services{
		Store:            database.GetStore(),
		RedisClient:      cache.GetClient(),
		Logger:           observability.GetLogger(),
		Metrics:          observability.GetMetrics(),
		Tracing:          observability.GetTracing(),
		TemporalPlatform: messaging.GetPlatform(),
	}

	logger.Info("Core services initialized", logger.Fields{
		"database_ready": c.services.Store != nil,
		"cache_ready":    c.services.RedisClient != nil,
		"temporal_ready": c.services.TemporalPlatform != nil,
	})

	return nil
}

// initializeFinanceServices creates finance-specific services
func (c *Core) initializeFinanceServices() error {
	// Placeholder for finance services initialization
	// This will be populated when we integrate the finance module
	c.financeServices = &FinanceServices{}

	logger.Info("Finance services initialized (placeholder)", logger.Fields{
		"status": "placeholder_ready",
	})

	return nil
}