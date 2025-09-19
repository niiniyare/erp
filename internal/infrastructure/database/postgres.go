package database

import (
	"context"
	"fmt"
	"os"

	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/config"
	"github.com/niiniyare/erp/internal/shared/logger"
)

// PostgreSQLComponent manages PostgreSQL database connections and lifecycle
type PostgreSQLComponent struct {
	store    db.Store
	config   *config.DatabaseSettings
	migrator *Migrator
}

// NewPostgreSQL creates a new PostgreSQL component
func NewPostgreSQL(cfg *config.DatabaseSettings) *PostgreSQLComponent {
	return &PostgreSQLComponent{
		config: cfg,
		migrator: &Migrator{
			MigrationURL: getEnvOrDefault("MIGRATION_URL", "file://db/migration"),
			DatabaseURL:  cfg.GetDatabaseURL(),
		},
	}
}

// Start initializes the PostgreSQL connection and runs migrations
func (p *PostgreSQLComponent) Start(ctx context.Context) error {
	databaseURL := p.config.GetDatabaseURL()

	// Test database connection first
	logger.Info("Testing PostgreSQL connection", logger.Fields{
		"db_host": p.config.Host,
		"db_name": p.config.Database,
	})

	testStore, err := db.NewDB(databaseURL)
	if err != nil {
		logger.Error("PostgreSQL connection failed", logger.Fields{
			"error":   err.Error(),
			"db_host": p.config.Host,
			"db_name": p.config.Database,
		})
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	testStore.Close()

	logger.Info("PostgreSQL connection test successful")

	// Run database migrations
	logger.Info("Running database migrations", logger.Fields{
		"migration_url": p.migrator.MigrationURL,
		"db_host":       p.config.Host,
		"db_name":       p.config.Database,
	})

	if err := p.migrator.Up(); err != nil {
		logger.Error("Database migration failed", logger.Fields{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Info("Database migrations completed successfully")

	// Initialize database store
	store, err := db.NewDB(databaseURL)
	if err != nil {
		return fmt.Errorf("failed to initialize database store: %w", err)
	}

	p.store = store

	logger.Info("PostgreSQL connection established", logger.Fields{
		"database": p.config.Database,
	})

	return nil
}

// Stop gracefully shuts down the PostgreSQL connection
func (p *PostgreSQLComponent) Stop(ctx context.Context) error {
	if p.store != nil {
		p.store.Close()
		logger.Info("PostgreSQL connection closed")
	}
	return nil
}

// Health checks the PostgreSQL connection health
func (p *PostgreSQLComponent) Health(ctx context.Context) error {
	if p.store == nil {
		return fmt.Errorf("PostgreSQL store not initialized")
	}

	// Use a simple ping equivalent (try to get a connection)
	testStore, err := db.NewDB(p.config.GetDatabaseURL())
	if err != nil {
		return fmt.Errorf("PostgreSQL health check failed: %w", err)
	}
	defer testStore.Close()

	return nil
}

// GetStore returns the database store
func (p *PostgreSQLComponent) GetStore() db.Store {
	return p.store
}

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
