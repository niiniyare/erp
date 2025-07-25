package main

import (
	"context"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

func runDBMigration(migrationURL string, dbSource string) {
	ctx := context.Background()
	log := logger.WithFields(logger.Fields{
		"migration_url": migrationURL,
		"db_source":     dbSource,
		"operation":     "database_migration",
	})

	// Initialize tracing service
	tracingConfig := tracing.DefaultConfig()
	tracingConfig.ServiceName = "erp-migration"
	t, err := tracing.NewTracingService(tracingConfig)
	if err != nil {
		log.Error("failed to initialize tracing", logger.Fields{"error": err.Error()})
	}
	defer func() {
		if t != nil {
			if shutdownErr := t.Shutdown(context.Background()); shutdownErr != nil {
				log.Error("failed to shutdown tracing", logger.Fields{"error": shutdownErr.Error()})
			}
		}
	}()

	// Start tracing span
	spanCtx, span := t.StartSpan(ctx, "database_migration",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("migration.url", migrationURL),
			attribute.String("db.source", dbSource),
			attribute.String("operation.type", "migration"),
		),
	)
	defer span.End()

	startTime := time.Now()
	log.Info("starting database migration")
	t.AddEvent(spanCtx, "migration_started", attribute.String("timestamp", startTime.Format(time.RFC3339)))

	// Create migration instance
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		span.SetStatus(codes.Error, "failed to create migration instance")
		span.RecordError(err)
		t.RecordError(spanCtx, err, tracing.WithErrorStatus(), tracing.WithErrorAttributes(
			attribute.String("error.type", "migration_instance_creation"),
			attribute.String("migration.url", migrationURL),
		))
		log.Fatal("cannot create new migrate instance", logger.Fields{
			"error":         err.Error(),
			"migration_url": migrationURL,
			"db_source":     dbSource,
		})
		return
	}

	t.AddEvent(spanCtx, "migration_instance_created")
	log.Debug("migration instance created successfully")

	// Get current version before migration
	version, dirty, versionErr := migration.Version()
	if versionErr != nil && versionErr != migrate.ErrNilVersion {
		log.Warn("could not get current migration version", logger.Fields{"error": versionErr.Error()})
	} else {
		log.Info("current migration status", logger.Fields{
			"version": version,
			"dirty":   dirty,
		})
		t.SetAttributes(spanCtx,
			attribute.Int("migration.current_version", int(version)),
			attribute.Bool("migration.dirty", dirty),
		)
	}

	// Run migration
	t.AddEvent(spanCtx, "migration_execution_started")
	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		span.SetStatus(codes.Error, "migration failed")
		span.RecordError(err)
		t.RecordError(spanCtx, err, tracing.WithErrorStatus(), tracing.WithErrorAttributes(
			attribute.String("error.type", "migration_execution"),
			attribute.String("migration.operation", "up"),
		))
		log.Fatal("failed to run migrate up", logger.Fields{
			"error":         err.Error(),
			"migration_url": migrationURL,
			"db_source":     dbSource,
			"duration":      time.Since(startTime).String(),
		})
		return
	}

	// Handle no change case
	if err == migrate.ErrNoChange {
		log.Info("no new migrations to apply", logger.Fields{
			"current_version": version,
			"duration":        time.Since(startTime).String(),
		})
		t.AddEvent(spanCtx, "migration_no_change",
			attribute.String("reason", "no_new_migrations"),
			attribute.String("duration", time.Since(startTime).String()),
		)
		span.SetStatus(codes.Ok, "no new migrations to apply")
		return
	}

	// Get final version after migration
	finalVersion, finalDirty, finalVersionErr := migration.Version()
	if finalVersionErr != nil {
		log.Warn("could not get final migration version", logger.Fields{"error": finalVersionErr.Error()})
	} else {
		t.SetAttributes(spanCtx,
			attribute.Int("migration.final_version", int(finalVersion)),
			attribute.Bool("migration.final_dirty", finalDirty),
		)
	}

	duration := time.Since(startTime)
	span.SetStatus(codes.Ok, "migration completed successfully")
	t.AddEvent(spanCtx, "migration_completed",
		attribute.String("duration", duration.String()),
		attribute.Int("final_version", int(finalVersion)),
		attribute.Bool("final_dirty", finalDirty),
	)

	log.Info("database migrated successfully", logger.Fields{
		"duration":           duration.String(),
		"final_version":      finalVersion,
		"dirty":              finalDirty,
		"migrations_applied": finalVersion - version,
	})

	// Log metrics
	log.Info("migration metrics", logger.Fields{
		"metric.migration.duration_ms":        duration.Milliseconds(),
		"metric.migration.success":            1,
		"metric.migration.version_from":       version,
		"metric.migration.version_to":         finalVersion,
		"metric.migration.migrations_applied": finalVersion - version,
	})
}
