package database

import (
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/niiniyare/erp/internal/shared/logger"
)

// Migrator handles database schema migrations
type Migrator struct {
	MigrationURL string
	DatabaseURL  string
}

// Up runs all pending migrations
func (m *Migrator) Up() error {
	startTime := time.Now()
	log := logger.WithFields(logger.Fields{
		"operation": "database_migration",
	})

	log.Info("Starting database migration")

	// Create migration instance
	migration, err := migrate.New(m.MigrationURL, m.DatabaseURL)
	if err != nil {
		log.Error("Cannot create new migrate instance", logger.Fields{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to create migration instance: %w", err)
	}
	defer migration.Close()

	// Get current version before migration
	version, dirty, versionErr := migration.Version()
	if versionErr != nil && versionErr != migrate.ErrNilVersion {
		log.Warn("Could not get current migration version", logger.Fields{"error": versionErr.Error()})
	} else if dirty {
		log.Warn("Database is in dirty state", logger.Fields{
			"version": version,
		})
	}

	// Run migration
	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Error("Failed to run migrate up", logger.Fields{
			"error":    err.Error(),
			"duration": time.Since(startTime).String(),
		})
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Handle no change case
	if err == migrate.ErrNoChange {
		log.Info("No new migrations to apply")
		return nil
	}

	// Get final version after migration
	finalVersion, _, finalVersionErr := migration.Version()
	if finalVersionErr != nil {
		log.Warn("Could not get final migration version", logger.Fields{"error": finalVersionErr.Error()})
	}

	duration := time.Since(startTime)
	migrationsApplied := finalVersion - version

	log.Info("Database migrated successfully", logger.Fields{
		"duration":           duration.String(),
		"final_version":      finalVersion,
		"migrations_applied": migrationsApplied,
	})

	return nil
}

// Down rolls back one migration
func (m *Migrator) Down() error {
	log := logger.WithFields(logger.Fields{
		"operation": "database_migration_rollback",
	})

	log.Info("Starting database migration rollback")

	migration, err := migrate.New(m.MigrationURL, m.DatabaseURL)
	if err != nil {
		log.Error("Cannot create new migrate instance", logger.Fields{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to create migration instance: %w", err)
	}
	defer migration.Close()

	if err = migration.Steps(-1); err != nil {
		log.Error("Failed to run migrate down", logger.Fields{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	log.Info("Database migration rolled back successfully")
	return nil
}

// Force forces the migration version without running migrations
func (m *Migrator) Force(version int) error {
	log := logger.WithFields(logger.Fields{
		"operation": "database_migration_force",
		"version":   version,
	})

	log.Info("Forcing database migration version")

	migration, err := migrate.New(m.MigrationURL, m.DatabaseURL)
	if err != nil {
		log.Error("Cannot create new migrate instance", logger.Fields{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to create migration instance: %w", err)
	}
	defer migration.Close()

	if err = migration.Force(version); err != nil {
		log.Error("Failed to force migration version", logger.Fields{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to force migration version: %w", err)
	}

	log.Info("Database migration version forced successfully")
	return nil
}

// Version returns the current migration version
func (m *Migrator) Version() (uint, bool, error) {
	migration, err := migrate.New(m.MigrationURL, m.DatabaseURL)
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migration instance: %w", err)
	}
	defer migration.Close()

	return migration.Version()
}