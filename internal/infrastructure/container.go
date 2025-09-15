package infrastructure

import (
	"context"
	"fmt"
	"sync"

	"github.com/niiniyare/erp/internal/config"
	"github.com/niiniyare/erp/internal/infrastructure/cache"
	"github.com/niiniyare/erp/internal/infrastructure/database"
	"github.com/niiniyare/erp/internal/infrastructure/messaging"
	"github.com/niiniyare/erp/internal/infrastructure/observability"
	"github.com/niiniyare/erp/internal/shared/logger"
)

// Container manages all infrastructure components
type Container struct {
	config      *config.AppConfig
	database    *database.PostgreSQLComponent
	cache       *cache.RedisComponent
	messaging   *messaging.TemporalComponent
	observability *observability.ObservabilityComponent
	started     bool
	mu          sync.RWMutex
}

// NewContainer creates a new infrastructure container
func NewContainer(cfg *config.AppConfig) *Container {
	return &Container{
		config:        cfg,
		database:      database.NewPostgreSQL(&cfg.Database),
		cache:         cache.NewRedis(&cfg.Redis),
		messaging:     nil, // Will be initialized after observability
		observability: observability.NewObservability(&cfg.Logger, &cfg.App),
	}
}

// Start initializes all infrastructure components in the correct order
func (c *Container) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.started {
		return fmt.Errorf("infrastructure container already started")
	}

	logger.Info("Starting infrastructure container", logger.Fields{
		"components": []string{"observability", "database", "cache", "messaging"},
	})

	// Start observability first (logging, metrics, tracing)
	if err := c.observability.Start(ctx); err != nil {
		return fmt.Errorf("failed to start observability: %w", err)
	}

	// Initialize messaging component now that observability is ready
	c.messaging = messaging.NewTemporal(&c.config.Temporal, c.observability.GetLogger().WithFields(logger.Fields{"component": "temporal"}))

	// Start database connections
	if err := c.database.Start(ctx); err != nil {
		return fmt.Errorf("failed to start database: %w", err)
	}

	// Start cache connections  
	if err := c.cache.Start(ctx); err != nil {
		return fmt.Errorf("failed to start cache: %w", err)
	}

	// Start messaging/workflow engine
	if err := c.messaging.Start(ctx); err != nil {
		return fmt.Errorf("failed to start messaging: %w", err)
	}

	c.started = true

	logger.Info("Infrastructure container started successfully", logger.Fields{
		"status": "ready",
		"components_initialized": 4,
	})

	return nil
}

// Stop gracefully shuts down all infrastructure components in reverse order
func (c *Container) Stop(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.started {
		return nil // Already stopped or never started
	}

	logger.Info("Stopping infrastructure container", logger.Fields{
		"status": "shutting_down",
	})

	var errs []error

	// Stop in reverse order of startup
	if c.messaging != nil {
		if err := c.messaging.Stop(ctx); err != nil {
			logger.Error("Failed to stop messaging", logger.Fields{"error": err.Error()})
			errs = append(errs, fmt.Errorf("messaging stop failed: %w", err))
		}
	}

	if c.cache != nil {
		if err := c.cache.Stop(ctx); err != nil {
			logger.Error("Failed to stop cache", logger.Fields{"error": err.Error()})
			errs = append(errs, fmt.Errorf("cache stop failed: %w", err))
		}
	}

	if c.database != nil {
		if err := c.database.Stop(ctx); err != nil {
			logger.Error("Failed to stop database", logger.Fields{"error": err.Error()})
			errs = append(errs, fmt.Errorf("database stop failed: %w", err))
		}
	}

	if c.observability != nil {
		if err := c.observability.Stop(ctx); err != nil {
			logger.Error("Failed to stop observability", logger.Fields{"error": err.Error()})
			errs = append(errs, fmt.Errorf("observability stop failed: %w", err))
		}
	}

	c.started = false

	if len(errs) > 0 {
		return fmt.Errorf("infrastructure container stop errors: %v", errs)
	}

	logger.Info("Infrastructure container stopped successfully")
	return nil
}

// Health checks the health of all infrastructure components
func (c *Container) Health(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.started {
		return fmt.Errorf("infrastructure container not started")
	}

	// Check all components
	if err := c.database.Health(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	if err := c.cache.Health(ctx); err != nil {
		return fmt.Errorf("cache health check failed: %w", err)
	}

	if c.messaging != nil {
		if err := c.messaging.Health(ctx); err != nil {
			return fmt.Errorf("messaging health check failed: %w", err)
		}
	}

	if err := c.observability.Health(ctx); err != nil {
		return fmt.Errorf("observability health check failed: %w", err)
	}

	return nil
}

// GetDatabase returns the database component
func (c *Container) GetDatabase() *database.PostgreSQLComponent {
	return c.database
}

// GetCache returns the cache component
func (c *Container) GetCache() *cache.RedisComponent {
	return c.cache
}

// GetMessaging returns the messaging component
func (c *Container) GetMessaging() *messaging.TemporalComponent {
	return c.messaging
}

// GetObservability returns the observability component
func (c *Container) GetObservability() *observability.ObservabilityComponent {
	return c.observability
}

// IsStarted returns whether the container is started
func (c *Container) IsStarted() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.started
}