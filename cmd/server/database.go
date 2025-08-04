package main

import (
	"context"
	"os"

	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/platform/config"
	"github.com/niiniyare/erp/internal/shared/logger"
)

type Database struct {
	Store       db.Store
	RedisClient cache.Service
}

func InitializeDatabase(cfg *config.Config) (*Database, error) {
	// Build database URL from config
	databaseURL := cfg.Database.GetDatabaseURL()

	// Test database connection first
	logger.Info("Testing database connection", logger.Fields{
		"db_host": cfg.Database.Host,
		"db_name": cfg.Database.Database,
	})

	testStore, err := db.NewDB(databaseURL)
	if err != nil {
		logger.Fatal("Database connection failed", logger.Fields{
			"error":   err.Error(),
			"db_host": cfg.Database.Host,
			"db_name": cfg.Database.Database,
		})
		return nil, err
	}
	testStore.Close()

	logger.Info("Database connection test successful")

	// Test Redis connection
	logger.Info("Testing Redis connection", logger.Fields{
		"redis_host": cfg.Redis.Host,
		"redis_port": cfg.Redis.Port,
	})

	redisConfig := cache.DefaultRedisConfig(&cfg.Redis)
	testRedisClient := cache.NewRedisClient(redisConfig)

	if err := testRedisClient.Ping(context.Background()); err != nil {
		logger.Fatal("Redis connection failed", logger.Fields{
			"error":      err.Error(),
			"redis_host": cfg.Redis.Host,
			"redis_port": cfg.Redis.Port,
		})
		return nil, err
	}

	logger.Info("Redis connection test successful")

	// Run database migrations
	migrationURL := os.Getenv("MIGRATION_URL")
	if migrationURL == "" {
		migrationURL = "file://db/migration"
	}

	logger.Info("Running database migrations", logger.Fields{
		"migration_url": migrationURL,
		"db_host":       cfg.Database.Host,
		"db_name":       cfg.Database.Database,
	})

	runDBMigration(migrationURL, databaseURL)
	logger.Info("Database migrations completed successfully")

	// Initialize database store
	store, err := db.NewDB(databaseURL)
	if err != nil {
		return nil, err
	}

	logger.Info("Database connection established", logger.Fields{
		"database": cfg.Database.Database,
	})

	logger.Info("Cache client initialized", logger.Fields{
		"redis_host": cfg.Redis.Host,
		"redis_port": cfg.Redis.Port,
	})

	return &Database{
		Store:       store,
		RedisClient: testRedisClient,
	}, nil
}

func (d *Database) Close() error {
	if d.Store != nil {
		d.Store.Close()
	}
	if d.RedisClient != nil {
		return d.RedisClient.Close()
	}
	return nil
}
