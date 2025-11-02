# Registry

## Overview

Registry manages schema storage, caching, and retrieval.

## Interface

```go
type Registry interface {
    Get(ctx context.Context, id string) (*Schema, error)
    Set(ctx context.Context, schema *Schema) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context) ([]*Schema, error)
    Exists(ctx context.Context, id string) (bool, error)
}
```

## Create Registry

```go
// File-based
registry := registry.NewFilesystemRegistry("./schemas")

// PostgreSQL
registry := registry.NewPostgresRegistry(db)

// With Redis cache
cached := registry.NewCachedRegistry(
    registry.NewPostgresRegistry(db),
    redisClient,
    1*time.Hour,
)
```

## Usage in Fiber Handler

```go
func HandleForm(c *fiber.Ctx) error {
    schemaID := c.Params("id")
    
    // Load schema
    schema, err := registry.Get(c.Context(), schemaID)
    if err != nil {
        return c.Status(404).SendString("Schema not found")
    }
    
    // Render
    return views.FormPage(schema, nil).Render(
        c.Context(), 
        c.Response().BodyWriter(),
    )
}
```

## Caching Strategy

```
Request
  ↓
Memory Cache (fastest, 1-5 min TTL)
  ↓ if miss
Redis Cache (fast, 1 hour TTL)
  ↓ if miss
PostgreSQL (source of truth)
  ↓
Cache in Redis + Memory
  ↓
Return schema
```

[← Back](09-storage-interface.md) | [Next: Enricher →](11-enricher.md)