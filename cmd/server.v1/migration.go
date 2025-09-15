package main

import (
	"time"

	"github.com/golang-migrate/migrate/v4"

	"github.com/niiniyare/erp/internal/shared/logger"
)

func runDBMigration(migrationURL string, dbSource string) {
	startTime := time.Now()
	log := logger.WithFields(logger.Fields{
		"operation": "database_migration",
	})

	log.Info("starting database migration")

	// Create migration instance
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal("cannot create new migrate instance", logger.Fields{
			"error": err.Error(),
		})
		return
	}

	// Get current version before migration
	version, dirty, versionErr := migration.Version()
	if versionErr != nil && versionErr != migrate.ErrNilVersion {
		log.Warn("could not get current migration version", logger.Fields{"error": versionErr.Error()})
	} else if dirty {
		log.Warn("database is in dirty state", logger.Fields{
			"version": version,
		})
	}

	// Run migration
	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("failed to run migrate up", logger.Fields{
			"error":    err.Error(),
			"duration": time.Since(startTime).String(),
		})
		return
	}

	// Handle no change case
	if err == migrate.ErrNoChange {
		log.Info("no new migrations to apply")
		return
	}

	// Get final version after migration
	finalVersion, _, finalVersionErr := migration.Version()
	if finalVersionErr != nil {
		log.Warn("could not get final migration version", logger.Fields{"error": finalVersionErr.Error()})
	}

	duration := time.Since(startTime)
	migrationsApplied := finalVersion - version

	log.Info("database migrated successfully", logger.Fields{
		"duration":           duration.String(),
		"final_version":      finalVersion,
		"migrations_applied": migrationsApplied,
	})
}
