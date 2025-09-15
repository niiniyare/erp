package cache

import (
	"context"
	"fmt"

	"github.com/niiniyare/erp/internal/config"
	"github.com/niiniyare/erp/internal/platform/cache"
	platformConfig "github.com/niiniyare/erp/internal/platform/config"
	"github.com/niiniyare/erp/internal/shared/logger"
)

// RedisComponent manages Redis cache connections and lifecycle
type RedisComponent struct {
	client cache.Service
	config *config.RedisSettings
}

// NewRedis creates a new Redis component
func NewRedis(cfg *config.RedisSettings) *RedisComponent {
	return &RedisComponent{
		config: cfg,
	}
}

// Start initializes the Redis connection
func (r *RedisComponent) Start(ctx context.Context) error {
	// Test Redis connection
	logger.Info("Testing Redis connection", logger.Fields{
		"redis_host": r.config.Host,
		"redis_port": r.config.Port,
	})

	// Convert new config format to platform config format first
	platformRedisConfig := &platformConfig.RedisConfig{
		Host:     r.config.Host,
		Port:     r.config.Port,
		Password: r.config.Password,
		DB:       r.config.DB,
	}

	redisConfig := cache.DefaultRedisConfig(platformRedisConfig)
	testRedisClient := cache.NewRedisClient(redisConfig)

	if err := testRedisClient.Ping(ctx); err != nil {
		logger.Error("Redis connection failed", logger.Fields{
			"error":      err.Error(),
			"redis_host": r.config.Host,
			"redis_port": r.config.Port,
		})
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Redis connection test successful")

	// Store the client for use
	r.client = testRedisClient

	logger.Info("Redis cache client initialized", logger.Fields{
		"redis_host": r.config.Host,
		"redis_port": r.config.Port,
	})

	return nil
}

// Stop gracefully shuts down the Redis connection
func (r *RedisComponent) Stop(ctx context.Context) error {
	if r.client != nil {
		if err := r.client.Close(); err != nil {
			logger.Error("Failed to close Redis connection", logger.Fields{
				"error": err.Error(),
			})
			return err
		}
		logger.Info("Redis connection closed")
	}
	return nil
}

// Health checks the Redis connection health
func (r *RedisComponent) Health(ctx context.Context) error {
	if r.client == nil {
		return fmt.Errorf("Redis client not initialized")
	}

	if err := r.client.Ping(ctx); err != nil {
		return fmt.Errorf("Redis health check failed: %w", err)
	}

	return nil
}

// GetClient returns the Redis cache service
func (r *RedisComponent) GetClient() cache.Service {
	return r.client
}