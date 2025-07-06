package main

import (
    "log"
    "net/http"
    
    "github.com/niiniyare/erp/internal/platform/config"
    "github.com/niiniyare/erp/internal/api/handlers"
    "github.com/niiniyare/erp/internal/core/tenant"
    "github.com/niiniyare/erp/internal/platform/cache"
    db "github.com/niiniyare/erp/db/sqlc"
)

func main() {
    // Load configuration
    cfg := config.Load()
    
    // Build database URL from config
    databaseURL := cfg.Database.GetDatabaseURL()
    
    // Initialize database store using SQLC
    store, err := db.NewDB(databaseURL)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer store.Close()
    
    // Initialize cache
    redisClient := cache.NewRedisClient(&cfg.Redis)
    
    // Initialize repositories
    tenantRepo := tenant.NewRepository(store)
    
    // Initialize services
    tenantService := tenant.NewService(tenantRepo, redisClient)
    
    // Initialize API handlers
    router := handlers.NewRouter(tenantService)
    
    // Start server
    log.Printf("Server starting on port %s", cfg.Server.Port)
    log.Fatal(http.ListenAndServe(":"+cfg.Server.Port, router))
}
