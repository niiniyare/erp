package main

import (
    "net/http"
    
    "github.com/niiniyare/erp/internal/platform/config"
    "github.com/niiniyare/erp/internal/api/handlers"
    "github.com/niiniyare/erp/internal/core/tenant"
    "github.com/niiniyare/erp/internal/platform/cache"
    "github.com/niiniyare/erp/internal/shared/logger"
    db "github.com/niiniyare/erp/db/sqlc"
)

func main() {
    // Initialize logger from environment
    if err := logger.InitializeFromEnv(); err != nil {
        panic("Failed to initialize logger: " + err.Error())
    }
    defer logger.Close()
    
    logger.Info("Starting ERP server", logger.Fields{
        "service": "erp-server",
        "version": "1.0.0",
    })
    
    // Load configuration
    cfg := config.Load()
    
    logger.Info("Configuration loaded", logger.Fields{
        "server_port": cfg.Server.Port,
        "db_host": cfg.Database.Host,
        "redis_host": cfg.Redis.Host,
    })
    
    // Build database URL from config
    databaseURL := cfg.Database.GetDatabaseURL()
    
    // Initialize database store using SQLC
    store, err := db.NewDB(databaseURL)
    if err != nil {
        logger.Fatal("Failed to connect to database", logger.Fields{
            "error": err.Error(),
            "database_url": databaseURL,
        })
    }
    defer store.Close()
    
    logger.Info("Database connection established", logger.Fields{
        "database": cfg.Database.Database,
    })
    
    // Initialize cache
    redisClient := cache.NewRedisClient(&cfg.Redis)
    
    logger.Info("Cache client initialized", logger.Fields{
        "redis_host": cfg.Redis.Host,
        "redis_port": cfg.Redis.Port,
    })
    
    // Initialize repositories
    tenantRepo := tenant.NewRepository(store)
    
    // Initialize services
    tenantService := tenant.NewService(tenantRepo, redisClient)
    
    // Initialize API handlers
    router := handlers.NewRouter(tenantService)
    
    // Start server
    logger.Info("Server starting", logger.Fields{
        "port": cfg.Server.Port,
        "address": ":" + cfg.Server.Port,
    })
    
    if err := http.ListenAndServe(":"+cfg.Server.Port, router); err != nil {
        logger.Fatal("Server failed to start", logger.Fields{
            "error": err.Error(),
            "port": cfg.Server.Port,
        })
    }
}
