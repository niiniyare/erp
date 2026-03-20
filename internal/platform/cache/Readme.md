#  Multi-Tenant Redis Cache Usage Guide

A  guide for using the  multi-tenant Redis cache that maintains backward compatibility while adding powerful new features.

## Table of Contents

- [Overview](#overview)
- [Backward Compatibility](#backward-compatibility)
- [Multi-Tenant Features](#multi-tenant-features)
- [ Configuration](#-configuration)
- [Usage Patterns](#usage-patterns)
- [Advanced Features](#advanced-features)
- [Monitoring and Health Checks](#monitoring-and-health-checks)
- [Best Practices](#best-practices)
- [Migration Guide](#migration-guide)

## Overview

The  Redis cache provides:

✅ **100% Backward Compatibility** - Existing code works unchanged  
✅ **Automatic Multi-Tenant Support** - Tenant isolation via context  
✅ ** Performance** - Compression, connection pooling, circuit breaker  
✅ **Advanced Features** - Bulk operations, distributed locks, health monitoring  
✅ **Production Ready** - Metrics, monitoring, error handling  

### Key Improvements

- **Multi-tenant isolation** through automatic key prefixing
- **Data compression** for large values
- **Circuit breaker** pattern for resilience
- **Connection pooling** optimization
- ** metrics** and monitoring
- **Bulk operations** for better performance
- **Distributed locking** support

## Backward Compatibility

Your existing code continues to work exactly as before:

```go
// Your existing code works unchanged!
cache := cache.NewRedisClient(redisConfig)

// These calls work exactly as they did before
err := cache.Get(ctx, "user:123", &user)
err = cache.Set(ctx, "user:123", user, time.Hour)
err = cache.Delete(ctx, "user:123")
err = cache.Flush(ctx)
```

The difference is that now these operations are **automatically tenant-aware** when tenant context is provided.

## Multi-Tenant Features

### 1. Automatic Tenant Isolation

```go
import (
    "github.com/google/uuid"
    "your-app/internal/cache"
)

func ExampleTenantIsolation() {
    cache := cache.NewRedisClient(redisConfig)
    
    tenantID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
    tenantSlug := "acme-corp"
    
    // Method 1: Set tenant context once
    ctx := cache.WithTenant(context.Background(), tenantID, tenantSlug)
    
    // All operations automatically use tenant prefix
    cache.Set(ctx, "user:123", user, time.Hour)
    // Key becomes: "erp:tenant:acme-corp:user:123"
    
    cache.Get(ctx, "user:123", &user)
    // Automatically looks for: "erp:tenant:acme-corp:user:123"
    
    // Method 2: Use tenant-aware helper
    cacheManager := cache.NewCacheManager(cache.(cache.Service))
    err := cacheManager.SetTenantData(ctx, tenantID.String(), tenantSlug, "user:123", user, time.Hour)
}
```

### 2. Namespace Support

```go
func ExampleNamespaces() {
    cache := cache.NewRedisClient(redisConfig)
    
    // Create different namespaces for different data types
    userCtx := cache.WithNamespace(context.Background(), "users")
    sessionCtx := cache.WithNamespace(context.Background(), "sessions")
    cacheCtx := cache.WithNamespace(context.Background(), "computed")
    
    // Each namespace is isolated
    cache.Set(userCtx, "123", user, time.Hour)     // Key: "erp:global:users:123"
    cache.Set(sessionCtx, "123", session, time.Hour) // Key: "erp:global:sessions:123"
    cache.Set(cacheCtx, "123", result, time.Hour)  // Key: "erp:global:computed:123"
}
```

### 3. Combined Tenant and Namespace

```go
func ExampleTenantWithNamespace() {
    cache := cache.NewRedisClient(redisConfig)
    
    tenantID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
    
    // Combine tenant and namespace context
    ctx := cache.WithTenantAndNamespace(
        context.Background(), 
        tenantID, 
        "acme-corp", 
        "analytics",
    )
    
    cache.Set(ctx, "monthly-report", report, 24*time.Hour)
    // Key becomes: "erp:tenant:acme-corp:analytics:monthly-report"
}
```

##  Configuration

### 1. Basic Configuration (Backward Compatible)

```go
// Your existing configuration works unchanged
redisConfig := &config.RedisConfig{
    Host:     "localhost",
    Port:     6379,
    Password: "",
    DB:       0,
}

cache := cache.NewRedisClient(redisConfig)
```

### 2.  Configuration

```go
//  configuration with all new features
Config := &cache.RedisConfig{
    RedisConfig: &config.RedisConfig{
        Host:     "localhost",
        Port:     6379,
        Password: "",
        DB:       0,
    },
    
    // Connection pool settings
    PoolSize:        20,
    MinIdleConns:    5,
    MaxConnAge:      time.Hour,
    PoolTimeout:     5 * time.Second,
    IdleTimeout:     10 * time.Minute,
    
    // Performance settings
    EnableCompression:   true,
    CompressionLevel:    6,
    KeyPrefix:          "myapp",
    MaxRetries:         3,
    RetryDelay:         100 * time.Millisecond,
    
    // Circuit breaker
    EnableCircuitBreaker:    true,
    CircuitBreakerThreshold: 5,
    CircuitBreakerTimeout:   30 * time.Second,
}

cache := cache.NewRedisClient(Config)
```

### 3. Environment-Specific Configurations

```go
func NewDevCache(baseConfig *config.RedisConfig) cache.Service {
    config := cache.DefaultRedisConfig(baseConfig)
    config.PoolSize = 5
    config.EnableCompression = false // Disable compression in dev
    config.EnableCircuitBreaker = false
    return cache.NewRedisClient(config)
}

func NewProdCache(baseConfig *config.RedisConfig) cache.Service {
    config := cache.DefaultRedisConfig(baseConfig)
    config.PoolSize = 50
    config.MinIdleConns = 10
    config.EnableCompression = true
    config.CompressionLevel = 9 // Maximum compression
    config.EnableCircuitBreaker = true
    return cache.NewRedisClient(config)
}
```

## Usage Patterns

### 1. Service Layer Integration

```go
type UserService struct {
    cache cache.Service
    db    database.Store
}

func (s *UserService) GetUser(ctx context.Context, userID string) (*User, error) {
    // Try cache first (automatically tenant-aware if context has tenant info)
    var user User
    err := s.cache.Get(ctx, fmt.Sprintf("user:%s", userID), &user)
    if err == nil {
        return &user, nil
    }
    
    // Cache miss - get from database
    if err == cache.ErrCacheMiss {
        dbUser, err := s.db.GetUser(ctx, userID)
        if err != nil {
            return nil, err
        }
        
        // Cache for next time (fire and forget)
        go func() {
            ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
            defer cancel()
            s.cache.Set(ctx, fmt.Sprintf("user:%s", userID), dbUser, time.Hour)
        }()
        
        return dbUser, nil
    }
    
    return nil, err
}

func (s *UserService) UpdateUser(ctx context.Context, user *User) error {
    // Update in database
    err := s.db.UpdateUser(ctx, user)
    if err != nil {
        return err
    }
    
    // Invalidate cache
    s.cache.Delete(ctx, fmt.Sprintf("user:%s", user.ID))
    return nil
}
```

### 2. Repository Pattern with Caching

```go
type UserRepository struct {
    cacheRepo *cache.CacheRepository[User]
    db        database.Store
}

func NewUserRepository(cacheService cache.Service, db database.Store) *UserRepository {
    cacheRepo := cache.NewCacheRepository[User](
        cacheService,
        func(id string) string { return fmt.Sprintf("user:%s", id) },
        func(ctx context.Context, id string) (*User, error) {
            return db.GetUser(ctx, id)
        },
        time.Hour, // TTL
    )
    
    return &UserRepository{
        cacheRepo: cacheRepo,
        db:        db,
    }
}

func (r *UserRepository) GetUser(ctx context.Context, id string) (*User, error) {
    return r.cacheRepo.Get(ctx, id)
}

func (r *UserRepository) UpdateUser(ctx context.Context, user *User) error {
    err := r.db.UpdateUser(ctx, user)
    if err != nil {
        return err
    }
    
    // Update cache
    return r.cacheRepo.Set(ctx, user.ID, user)
}
```

### 3. HTTP Middleware Integration

```go
func TenantCacheMiddleware(cache cache.Service) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract tenant from request (URL, header, etc.)
            tenantSlug := mux.Vars(r)["tenant"]
            
            if tenantSlug != "" {
                // Add tenant context for cache operations
                tenantID := getTenantIDFromSlug(tenantSlug) // Your implementation
                ctx := cache.WithTenant(r.Context(), tenantID, tenantSlug)
                r = r.WithContext(ctx)
            }
            
            next.ServeHTTP(w, r)
        })
    }
}

// Response caching middleware
func ResponseCacheMiddleware(cache cache.Service) func(http.Handler) http.Handler {
    cacheMiddleware := cache.NewCacheMiddleware(cache)
    
    return cacheMiddleware.WithCache(
        func(r *http.Request) string {
            // Generate cache key from request
            return fmt.Sprintf("response:%s:%s", r.Method, r.URL.Path)
        },
        5*time.Minute, // Cache responses for 5 minutes
    )
}
```

## Advanced Features

### 1. Bulk Operations

```go
func ExampleBulkOperations() {
    cache := cache.NewRedisClient(Config)
    
    // Bulk set
    data := map[string]interface{}{
        "user:1": user1,
        "user:2": user2,
        "user:3": user3,
    }
    err := cache.MSet(ctx, data, time.Hour)
    
    // Bulk get
    keys := []string{"user:1", "user:2", "user:3"}
    var users []User
    err = cache.MGet(ctx, keys, &users)
    
    // Bulk delete
    err = cache.MDelete(ctx, keys)
}
```

### 2. Distributed Locking

```go
func ExampleDistributedLock() {
    cache := cache.NewRedisClient(redisConfig)
    
    // Method 1: Manual lock management
    lock := cache.NewDistributedLock(cache, "process-orders", 30*time.Second)
    
    acquired, err := lock.TryLock(ctx)
    if err != nil {
        return err
    }
    if !acquired {
        return errors.New("could not acquire lock")
    }
    defer lock.Unlock(ctx)
    
    // Critical section - only one instance can execute this
    processOrders()
    
    // Method 2: Automatic lock management
    err = cache.WithLock(ctx, cache, "process-orders", 30*time.Second, func() error {
        return processOrders()
    })
}
```

### 3. Cache-Aside Pattern

```go
func ExampleCacheAside() {
    cache := cache.NewRedisClient(redisConfig)
    
    // Create cache with fallback to database
    cacheWithFallback := cache.NewCacheWithFallback(
        cache,
        func(ctx context.Context, key string) (interface{}, error) {
            // Extract user ID from key
            userID := strings.TrimPrefix(key, "user:")
            return getUserFromDatabase(ctx, userID)
        },
    )
    
    // This will automatically fallback to database if cache miss
    var user User
    err := cacheWithFallback.Get(ctx, "user:123", &user)
    // If cache miss, it automatically loads from database and caches the result
}
```

### 4. Compression Control

```go
func ExampleCompression() {
    cache := cache.NewRedisClient(Config)
    
    // Large data will be automatically compressed
    largeData := generateLargeReport() // > 1KB
    
    // Will be automatically compressed
    cache.Set(ctx, "large-report", largeData, time.Hour)
    
    // Force compression off for specific operation
    ctx = cache.WithCompression(ctx, false)
    cache.Set(ctx, "small-data", data, time.Hour)
    
    // Force compression on for specific operation
    ctx = cache.WithCompression(ctx, true)
    cache.Set(ctx, "force-compressed", data, time.Hour)
}
```

## Monitoring and Health Checks

### 1. Basic Health Check

```go
func HealthCheckHandler(cache cache.Service) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
        defer cancel()
        
        err := cache.Ping(ctx)
        if err != nil {
            http.Error(w, "Cache unhealthy", http.StatusServiceUnavailable)
            return
        }
        
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Cache healthy"))
    }
}
```

### 2.  Monitoring

```go
func SetupCacheMonitoring(cache cache.Service) {
    monitor := cache.NewCacheHealthMonitor(cache)
    
    // Start continuous monitoring
    ctx := context.Background()
    healthResults := monitor.StartHealthMonitoring(ctx, 30*time.Second)
    
    go func() {
        for metrics := range healthResults {
            log.Printf("Cache Health: %+v", metrics)
            
            if !metrics.IsHealthy {
                // Send alert
                sendAlert("cache_unhealthy", metrics)
            }
            
            if metrics.ErrorRate > 0.1 { // 10% error rate
                log.Printf("High cache error rate: %.2f%%", metrics.ErrorRate*100)
            }
            
            // Log stats
            if metrics.Stats != nil {
                log.Printf("Cache Stats - Hits: %d, Misses: %d, Hit Ratio: %.2f%%",
                    metrics.Stats.Hits,
                    metrics.Stats.Misses,
                    metrics.Stats.HitRatio*100,
                )
            }
        }
    }()
}
```

### 3. Metrics Export

```go
func ExportCacheMetrics(cache cache.Service) map[string]interface{} {
    stats := cache.Stats()
    
    return map[string]interface{}{
        "cache_hits_total":           stats.Hits,
        "cache_misses_total":         stats.Misses,
        "cache_sets_total":           stats.Sets,
        "cache_deletes_total":        stats.Deletes,
        "cache_errors_total":         stats.Errors,
        "cache_hit_ratio":            stats.HitRatio,
        "cache_average_latency_ms":   stats.AverageLatency.Milliseconds(),
        "cache_connections_active":   stats.ConnectionsActive,
        "cache_connections_idle":     stats.ConnectionsIdle,
    }
}
```

## Best Practices

### 1. Context Management

```go
// Good: Set tenant context early and pass it down
func HandleRequest(w http.ResponseWriter, r *http.Request) {
    tenantSlug := extractTenantFromRequest(r)
    tenantID := getTenantID(tenantSlug)
    
    // Set tenant context once
    ctx := cache.WithTenant(r.Context(), tenantID, tenantSlug)
    
    // Pass context to all service calls
    processRequest(ctx, requestData)
}

func processRequest(ctx context.Context, data RequestData) {
    // All cache operations automatically use tenant context
    cache.Set(ctx, "key", data, time.Hour)
    cache.Get(ctx, "key", &result)
}
```

### 2. Error Handling

```go
func HandleCacheErrors(err error) {
    switch {
    case errors.Is(err, cache.ErrCacheMiss):
        // Normal cache miss - not an error
        log.Debug("Cache miss occurred")
        
    case errors.Is(err, cache.ErrCacheUnavailable):
        // Cache is down - degrade gracefully
        log.Error("Cache unavailable, falling back to database")
        
    case errors.Is(err, cache.ErrCircuitBreakerOpen):
        // Circuit breaker is open - cache is unhealthy
        log.Error("Cache circuit breaker is open")
        
    default:
        // Other cache errors
        log.Error("Cache error: %v", err)
    }
}
```

### 3. Key Management

```go
// Good: Use consistent key patterns
const (
    UserCacheKey     = "user:%s"
    SessionCacheKey  = "session:%s"
    ReportCacheKey   = "report:%s:%s" // type:date
)

func GetUserCacheKey(userID string) string {
    return fmt.Sprintf(UserCacheKey, userID)
}

func GetReportCacheKey(reportType, date string) string {
    return fmt.Sprintf(ReportCacheKey, reportType, date)
}
```

### 4. TTL Management

```go
var (
    // Define TTLs as constants
    UserCacheTTL     = time.Hour
    SessionCacheTTL  = 30 * time.Minute
    ReportCacheTTL   = 24 * time.Hour
    TempCacheTTL     = 5 * time.Minute
)

func CacheUserData(ctx context.Context, cache cache.Service, user *User) {
    key := GetUserCacheKey(user.ID)
    cache.Set(ctx, key, user, UserCacheTTL)
}
```

## Migration Guide

### 1. Zero-Downtime Migration

Your existing code requires **no changes** for basic migration:

```go
// Before (still works exactly the same)
cache := cache.NewRedisClient(redisConfig)
cache.Get(ctx, "key", &value)
cache.Set(ctx, "key", value, time.Hour)

// After (same code, but now with multi-tenant support when context provides tenant info)
// No code changes needed!
```

### 2. Gradual Enhancement

Add tenant context gradually:

```go
// Step 1: Add tenant context in new code
func NewHandler(cache cache.Service) *Handler {
    return &Handler{cache: cache}
}

func (h *Handler) HandleNewEndpoint(w http.ResponseWriter, r *http.Request) {
    // Add tenant context for new endpoints
    ctx := addTenantContext(r.Context(), r)
    
    // Now cache operations are tenant-aware
    h.processWithCache(ctx)
}

func (h *Handler) HandleOldEndpoint(w http.ResponseWriter, r *http.Request) {
    // Old endpoints continue to work without tenant context
    h.processWithCache(r.Context())
}
```

### 3. Testing Migration

```go
func TestCacheCompatibility(t *testing.T) {
    cache := cache.NewRedisClient(testConfig)
    
    // Test 1: Existing functionality still works
    ctx := context.Background()
    err := cache.Set(ctx, "test", "value", time.Minute)
    assert.NoError(t, err)
    
    var result string
    err = cache.Get(ctx, "test", &result)
    assert.NoError(t, err)
    assert.Equal(t, "value", result)
    
    // Test 2: Tenant isolation works
    tenant1 := cache.WithTenant(ctx, uuid.New(), "tenant1")
    tenant2 := cache.WithTenant(ctx, uuid.New(), "tenant2")
    
    cache.Set(tenant1, "shared-key", "tenant1-value", time.Minute)
    cache.Set(tenant2, "shared-key", "tenant2-value", time.Minute)
    
    var value1, value2 string
    cache.Get(tenant1, "shared-key", &value1)
    cache.Get(tenant2, "shared-key", &value2)
    
    assert.Equal(t, "tenant1-value", value1)
    assert.Equal(t, "tenant2-value", value2)
}
```

## Overview

This cache service is **multi-tenant native** - every operation automatically uses tenant context without manual `WithTenant()` calls everywhere. Simply set up middleware once, and all cache operations become tenant-aware.

## Quick Start

### 1. Setup (One Time)

```go
package main

import (
    "awo/internal/platform/cache"
    "awo/internal/platform/config"
    "awo/middleware"
)

func main() {
    // Initialize cache
    cfg := cache.DefaultRedisConfig(&config.RedisConfig{
        Host: "localhost",
        Port: 6379,
    })
    
    cacheService, err := cache.NewRedisClient(cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer cacheService.Close()
    
    // Setup router with tenant middleware
    r := chi.NewRouter()
    
    // THIS IS THE KEY - Set up tenant middleware once
    r.Use(middleware.TenantContextMiddleware())
    
    // Now all your handlers automatically have tenant context!
    r.Get("/users/{id}", getUserHandler(cacheService))
    
    http.ListenAndServe(":8080", r)
}
```

### 2. Use Cache Everywhere (No WithTenant Needed!)

```go
func getUserHandler(cache cache.Service) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context() // Already has tenant context from middleware!
        
        userID := chi.URLParam(r, "id")
        var user User
        
        // Just use cache normally - tenant isolation is automatic!
        err := cache.Get(ctx, "user:"+userID, &user)
        if err != nil {
            if errors.Is(err, cache.ErrCacheMiss) {
                // Fetch from DB and cache it
                user = fetchUserFromDB(userID)
                cache.Set(ctx, "user:"+userID, user, 1*time.Hour)
            } else {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
        }
        
        json.NewEncoder(w).Encode(user)
    }
}
```

## Configuration

### Multi-Tenant Settings

```go
cfg := cache.DefaultRedisConfig(baseCfg)

// Tenant enforcement (default: true)
cfg.RequireTenantContext = true   // Require tenant context on all operations
cfg.AllowGlobalOperations = false // Disallow operations without tenant context

// When true: All cache operations MUST have tenant_id, tenant_slug, or tenant_subdomain
// When false: Operations without tenant context are allowed (uses global keyspace)
```

### Memory Cache (for Formulas & Shared Data)

```go
cfg.EnableMemoryCache = true
cfg.MemoryCacheMaxSize = 1000              // Max items in memory
cfg.MemoryCacheDefaultTTL = 5 * time.Minute
cfg.MemoryCacheCleanupInterval = 1 * time.Minute
```

## Tenant Context Methods

### Automatic (Recommended)

Use middleware - set tenant context once, use everywhere:

```go
// Setup middleware
r.Use(middleware.TenantContextMiddleware())

// That's it! All cache operations are now tenant-aware
```

### Manual (When Needed)

```go
// Add tenant ID to context
ctx = cache.WithTenantID(ctx, tenantID)

// Add tenant slug to context
ctx = cache.WithTenantSlug(ctx, "acme-corp")

// Add tenant subdomain to context
ctx = cache.WithTenantSubdomain(ctx, "acme")

// Cache operations use whatever is available (priority: ID > slug > subdomain)
```

## Core Operations

All operations are **automatically tenant-aware** when middleware is configured:

```go
// Get
var user User
err := cache.Get(ctx, "user:123", &user)

// Set
err := cache.Set(ctx, "user:123", user, 1*time.Hour)

// Delete
err := cache.Delete(ctx, "user:123")

// Flush (only flushes current tenant's cache)
err := cache.Flush(ctx)

// Exists
exists, err := cache.Exists(ctx, "user:123")

// TTL
ttl, err := cache.TTL(ctx, "user:123")

// Expire
err := cache.Expire(ctx, "user:123", 2*time.Hour)
```

## Bulk Operations

```go
// MGet - Get multiple keys
results, err := cache.MGet(ctx, []string{"user:1", "user:2", "user:3"})
for _, result := range results {
    if result.Err == nil {
        // Use result.
