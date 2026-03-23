// Package wire - Platform layer providers
package wire

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	temporalclient "go.temporal.io/sdk/client"

	db "awo.so/db/sqlc"
	"awo.so/internal/platform/cache"
	"awo.so/internal/platform/config"
	"awo.so/internal/platform/temporal"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
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
func NewDBStore(pool *pgxpool.Pool) db.Store {
	return db.NewStore(pool)
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
	redisConfig := cache.DefaultRedisConfig(&cfg.Redis)
	return cache.NewRedisClient(redisConfig)
}

// ============================================================================
// OBSERVABILITY PROVIDERS
// ============================================================================

// NewLogger creates a new logger based on configuration
func NewLogger(cfg *config.Config) (logger.Logger, error) {
	factory := &logger.LoggerFactory{}

	// Convert the config properly using the ToLoggerConfig method
	loggerConfig := cfg.Logger.ToLoggerConfig(&cfg.App)

	return factory.NewLogger(logger.Config{
		Type:        logger.LoggerType(loggerConfig.Type),
		Level:       logger.LogLevel(loggerConfig.Level),
		Output:      loggerConfig.Output, // This ensures we have a valid io.Writer
		Format:      loggerConfig.Format,
		Development: loggerConfig.Development,
		ServiceName: loggerConfig.ServiceName,
		Version:     loggerConfig.Version,
	})
}

// NewMetricsProvider creates a new metrics provider
func NewMetricsProvider(cfg *config.Config, log logger.Logger) (metrics.MetricsProvider, error) {
	return metrics.NewMetricsService(metrics.MetricsConfig{
		Provider:  "prometheus",
		Namespace: cfg.App.Name,
		Subsystem: "api",
		Enabled:   true,
	})
}

// NewTracingService creates a new tracing service
func NewTracingService(cfg *config.Config, log logger.Logger) (tracing.Service, error) {
	return tracing.NewService(tracing.Config{
		ServiceName:        cfg.App.Name,
		ServiceVersion:     cfg.App.Version,
		Environment:        cfg.App.Environment,
		ExporterType:       tracing.ExporterType(cfg.Tracing.Exporter),
		Protocol:           tracing.Protocol(cfg.Tracing.Protocol),
		Endpoint:           cfg.Tracing.Endpoint,
		Insecure:           cfg.Tracing.Insecure,
		SamplingRatio:      cfg.Tracing.SamplingRatio,
		BatchTimeout:       time.Second * 5,
		MaxExportBatchSize: 100,
		MaxQueueSize:       1000,
		Enabled:            cfg.Tracing.Enabled,
	})
}

// ============================================================================
// TEMPORAL PROVIDERS
// ============================================================================

// NewTemporalClient creates a new Temporal client
func NewTemporalClient(cfg *config.Config, log logger.Logger) (temporalclient.Client, error) {
	clientManager, err := temporal.NewClientManager(&cfg.Temporal, log)
	if err != nil {
		return nil, err
	}
	return clientManager.GetClient(), nil
}
