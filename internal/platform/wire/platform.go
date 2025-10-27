// Package wire - Platform layer providers
package wire

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/go-redis/redis/v8"
	temporalclient "go.temporal.io/sdk/client"

	"github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/platform/config"
	"github.com/niiniyare/erp/internal/platform/temporal"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// ============================================================================
// DATABASE PROVIDERS
// ============================================================================

// NewDatabaseConnection creates a new PostgreSQL connection pool
func NewDatabaseConnection(cfg *config.Config) (*pgxpool.Pool, error) {
	ctx := context.Background()
	
	// Create connection config
	pgxConfig, err := pgxpool.ParseConfig(cfg.Database.GetDatabaseURL())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}
	
	// Configure connection pool
	pgxConfig.MaxConns = int32(cfg.Database.MaxOpenConns)
	pgxConfig.MinConns = int32(cfg.Database.MaxIdleConns)
	pgxConfig.MaxConnLifetime = cfg.Database.ConnMaxLifetime
	pgxConfig.MaxConnIdleTime = 30 * time.Minute
	
	// Create connection pool with retries
	pool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}
	
	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	
	return pool, nil
}

// NewDBStore creates a new SQLC database store
func NewDBStore(pool *pgxpool.Pool) *sqlc.Store {
	return sqlc.NewStore(pool)
}

// ============================================================================
// CACHE PROVIDERS
// ============================================================================

// NewRedisClient creates a new Redis client
func NewRedisClient(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		
		// Connection pool settings
		PoolSize:     10,
		MinIdleConns: 5,
		
		// Timeouts
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		
		// Retry settings
		MaxRetries:      3,
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 512 * time.Millisecond,
	})
	
	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	
	return client, nil
}

// NewCacheService creates a new cache service
func NewCacheService(cfg *config.Config) (cache.Service, error) {
	redisClient, err := NewRedisClient(cfg)
	if err != nil {
		return nil, err
	}
	
	return cache.NewService(redisClient), nil
}

// ============================================================================
// OBSERVABILITY PROVIDERS
// ============================================================================

// NewLogger creates a new logger based on configuration
func NewLogger(cfg *config.Config) (logger.Logger, error) {
	switch cfg.Logger.Type {
	case "zerolog":
		return logger.NewZerologLogger(logger.ZerologConfig{
			Level:       cfg.Logger.Level,
			Format:      cfg.Logger.Format,
			Development: cfg.Logger.Development,
			ServiceName: cfg.Logger.ServiceName,
			Version:     cfg.Logger.Version,
			Output:      cfg.Logger.Output,
		})
	case "zap":
		return logger.NewZapLogger(logger.ZapConfig{
			Level:       cfg.Logger.Level,
			Format:      cfg.Logger.Format,
			Development: cfg.Logger.Development,
			ServiceName: cfg.Logger.ServiceName,
			Version:     cfg.Logger.Version,
			Output:      cfg.Logger.Output,
		})
	case "slog":
		return logger.NewSlogLogger(logger.SlogConfig{
			Level:       cfg.Logger.Level,
			Format:      cfg.Logger.Format,
			Development: cfg.Logger.Development,
			ServiceName: cfg.Logger.ServiceName,
			Version:     cfg.Logger.Version,
			Output:      cfg.Logger.Output,
		})
	default:
		return logger.NewZerologLogger(logger.ZerologConfig{
			Level:       cfg.Logger.Level,
			Format:      cfg.Logger.Format,
			Development: cfg.Logger.Development,
			ServiceName: cfg.Logger.ServiceName,
			Version:     cfg.Logger.Version,
			Output:      cfg.Logger.Output,
		})
	}
}

// NewMetricsProvider creates a new metrics provider
func NewMetricsProvider(cfg *config.Config, log logger.Logger) (metrics.MetricsProvider, error) {
	return metrics.NewPrometheusMetrics(metrics.PrometheusConfig{
		ServiceName: cfg.App.Name,
		Version:     cfg.App.Version,
		Environment: cfg.App.Environment,
		Namespace:   cfg.App.Namespace,
	})
}

// NewTracingService creates a new tracing service
func NewTracingService(cfg *config.Config, log logger.Logger) (tracing.TracingService, error) {
	return tracing.NewService(tracing.Config{
		ServiceName:    cfg.App.Name,
		ServiceVersion: cfg.App.Version,
		Environment:    cfg.App.Environment,
		Endpoint:       "", // Configure based on your tracing backend
		SampleRate:     1.0, // Adjust based on environment
	})
}

// ============================================================================
// TEMPORAL PROVIDERS
// ============================================================================

// NewTemporalClient creates a new Temporal client
func NewTemporalClient(cfg *config.Config, log logger.Logger) (temporalclient.Client, error) {
	return temporal.NewClient(temporal.ClientConfig{
		HostPort:  cfg.Temporal.HostPort,
		Namespace: cfg.Temporal.Namespace,
		Identity:  cfg.Temporal.Client.Identity,
		Logger:    log,
	})
}

// ============================================================================
// UTILITY PROVIDERS
// ============================================================================

// NewDatabaseMigrator creates a database migrator (for development/testing)
func NewDatabaseMigrator(cfg *config.Config) (*DatabaseMigrator, error) {
	return &DatabaseMigrator{
		DatabaseURL:    cfg.Database.GetDatabaseURL(),
		MigrationsPath: cfg.Migration.URL,
		Timeout:        cfg.Migration.Timeout,
		Verbose:        cfg.Migration.Verbose,
	}, nil
}

// DatabaseMigrator handles database migrations
type DatabaseMigrator struct {
	DatabaseURL    string
	MigrationsPath string
	Timeout        time.Duration
	Verbose        bool
}

// Migrate runs database migrations
func (m *DatabaseMigrator) Migrate() error {
	// Implement migration logic using golang-migrate
	// This would integrate with your existing migration setup
	return nil
}