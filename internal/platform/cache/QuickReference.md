# Cache Service - Quick Reference

## Setup (One Time Only!)

```go
// 1. Initialize cache
cfg := cache.DefaultRedisConfig(&config.RedisConfig{
    Host: "localhost",
    Port: 6379,
})
cacheService, _ := cache.NewRedisClient(cfg)

// 2. Add middleware to router
r.Use(middleware.TenantContextMiddleware())

// 3. Done! All cache operations are now tenant-aware
```

## Basic Operations

```go
// All operations automatically use tenant context from middleware

// Get
var user User
err := cache.Get(ctx, "user:123", &user)

// Set
err := cache.Set(ctx, "user:123", user, 1*time.Hour)

// Delete
err := cache.Delete(ctx, "user:123")

// Exists
exists, err := cache.Exists(ctx, "user:123")
```

## Bulk Operations

```go
// Get multiple
results, _ := cache.MGet(ctx, []string{"key1", "key2"})

// Set multiple
pairs := map[string]any{"key1": val1, "key2": val2}
cache.MSet(ctx, pairs, 1*time.Hour)

// Delete multiple
cache.MDelete(ctx, []string{"key1", "key2"})
```

## Memory Cache

```go
// Tenant-specific memory cache
cache.SetMemory(ctx, "active-users", users, 5*time.Minute)
cache.GetMemory(ctx, "active-users", &users)

// Global memory cache (formulas, configs)
cache.SetGlobalMemory("formula:vat", formula, 1*time.Hour)
cache.GetGlobalMemory("formula:vat", &formula)
```

## Pattern Operations

```go
// Delete pattern
cache.DeletePattern(ctx, "user:*")

// Get keys
keys, _ := cache.Keys(ctx, "session:*")
```

## Middleware Options

```go
// Basic tenant extraction
r.Use(middleware.TenantContextMiddleware())

// JWT tenant extraction
r.Use(middleware.JWTTenantMiddleware())

// Require tenant
r.Use(middleware.RequireTenantMiddleware())

// Add namespace
r.Use(middleware.NamespaceMiddleware("v1"))
```

## Manual Context (When Needed)

```go
// Add tenant ID
ctx = cache.WithTenantID(ctx, tenantID)

// Add tenant slug
ctx = cache.WithTenantSlug(ctx, "acme-corp")

// Add tenant subdomain
ctx = cache.WithTenantSubdomain(ctx, "acme")

// Add namespace
ctx = cache.WithNamespace(ctx, "products")
```

## Error Handling

```go
err := cache.Get(ctx, key, &value)

if errors.Is(err, cache.ErrCacheMiss) {
    // Key not found - fetch from DB
}
if errors.Is(err, cache.ErrNoTenantContext) {
    // No tenant context - check middleware
}
if errors.Is(err, cache.ErrCircuitOpen) {
    // Circuit breaker open - use fallback
}
```

## Configuration

```go
cfg := cache.DefaultRedisConfig(baseCfg)

// Tenant settings
cfg.RequireTenantContext = true  // Enforce tenant on all ops
cfg.AllowGlobalOperations = false // Disallow non-tenant ops

// Memory cache
cfg.EnableMemoryCache = true
cfg.MemoryCacheMaxSize = 1000
cfg.MemoryCacheDefaultTTL = 5 * time.Minute

// Performance
cfg.EnableCompression = true
cfg.PoolSize = 10
cfg.BatchDeleteSize = 1000

// Circuit breaker
cfg.EnableCircuitBreaker = true
cfg.CircuitBreakerThreshold = 5
cfg.CircuitBreakerTimeout = 30 * time.Second
```

## Common Patterns

### Cache-Aside
```go
err := cache.Get(ctx, key, &data)
if errors.Is(err, cache.ErrCacheMiss) {
    data = fetchFromDB()
    cache.Set(ctx, key, data, 1*time.Hour)
}
```

### Write-Through
```go
updateDB(data)
cache.Set(ctx, key, data, 1*time.Hour)
```

### Invalidation
```go
updateDB(data)
cache.MDelete(ctx, []string{"key1", "key2"})
cache.DeletePattern(ctx, "list:*")
```

## Monitoring

```go
// Get stats
stats := cache.Stats()
fmt.Printf("Hit Ratio: %.2f%%\n", stats.HitRatio*100)
fmt.Printf("Memory Cache: %d items\n", stats.MemoryCacheSize)

// Health check
err := cache.Ping(ctx)

// Reset stats
cache.Reset()
```

## TTL Constants

```go
const (
    TTLRealtime = 1 * time.Minute
    TTLFrequent = 10 * time.Minute
    TTLStandard = 1 * time.Hour
    TTLStable   = 24 * time.Hour
)
```

## Key Patterns

```go
// Use consistent patterns
"user:{id}"
"user:{id}:profile"
"user:{id}:permissions"
"session:user:{id}:{token}"
"list:users:page:{page}"
```

## What's Automatic?

✅ Tenant isolation via middleware
✅ Compression for large values
✅ Key hashing for long keys
✅ Circuit breaker protection
✅ Connection pooling
✅ Statistics tracking
✅ Expiration cleanup (memory cache)

## What You Control?

🎛️ Tenant context (via middleware or manual)
🎛️ Cache keys and structure
🎛️ TTL values
🎛️ Memory vs Redis cache choice
🎛️ Error handling and fallbacks
🎛️ Invalidation strategy

## Remember

- **Set up middleware once** - never call WithTenant() again
- **Use memory cache for hot data** - formulas, configs, permissions
- **Handle cache failures gracefully** - always fallback to database
- **Use appropriate TTLs** - short for changing data, long for stable
- **Invalidate smartly** - clear related entries, not entire cache
