# Storage Interface

The registry uses a storage interface to support multiple backends: PostgreSQL, Redis, filesystem, and S3/Minio.

## Storage Interface Definition

```go
package registry

import "context"

// Storage defines the interface for schema storage backends
type Storage interface {
    // Get retrieves a schema by ID
    Get(ctx context.Context, id string) ([]byte, error)
    
    // Set stores a schema
    Set(ctx context.Context, id string, data []byte) error
    
    // Delete removes a schema
    Delete(ctx context.Context, id string) error
    
    // List returns all schema IDs
    List(ctx context.Context) ([]string, error)
    
    // Exists checks if a schema exists
    Exists(ctx context.Context, id string) (bool, error)
}
```

## PostgreSQL Storage

Primary storage with ACID guarantees.

### Schema

```sql
CREATE TABLE schemas (
    id VARCHAR(100) PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    version VARCHAR(20) NOT NULL,
    title VARCHAR(200) NOT NULL,
    json_schema JSONB NOT NULL,
    tenant_id UUID,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by UUID,
    updated_by UUID
);

CREATE INDEX idx_schemas_tenant ON schemas(tenant_id);
CREATE INDEX idx_schemas_type ON schemas(type);
CREATE INDEX idx_schemas_updated ON schemas(updated_at DESC);

-- For multi-tenant isolation
ALTER TABLE schemas ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON schemas
    FOR ALL
    USING (tenant_id = current_setting('app.current_tenant')::UUID);
```

### Implementation

```go
package registry

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
)

type PostgresStorage struct {
    db *sql.DB
}

func NewPostgresStorage(db *sql.DB) *PostgresStorage {
    return &PostgresStorage{db: db}
}

func (s *PostgresStorage) Get(ctx context.Context, id string) ([]byte, error) {
    var data []byte
    query := `SELECT json_schema FROM schemas WHERE id = $1`
    
    err := s.db.QueryRowContext(ctx, query, id).Scan(&data)
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("schema not found: %s", id)
    }
    if err != nil {
        return nil, fmt.Errorf("database error: %w", err)
    }
    
    return data, nil
}

func (s *PostgresStorage) Set(ctx context.Context, id string, data []byte) error {
    query := `
        INSERT INTO schemas (id, type, version, title, json_schema)
        SELECT $1, $2, $3, $4, $5
        ON CONFLICT (id) DO UPDATE
        SET json_schema = EXCLUDED.json_schema,
            updated_at = NOW()
    `
    
    // Extract metadata from JSON
    var schema map[string]any
    if err := json.Unmarshal(data, &schema); err != nil {
        return err
    }
    
    _, err := s.db.ExecContext(ctx, query,
        id,
        schema["type"],
        schema["version"],
        schema["title"],
        data,
    )
    
    return err
}

func (s *PostgresStorage) Delete(ctx context.Context, id string) error {
    query := `DELETE FROM schemas WHERE id = $1`
    result, err := s.db.ExecContext(ctx, query, id)
    if err != nil {
        return err
    }
    
    rows, _ := result.RowsAffected()
    if rows == 0 {
        return fmt.Errorf("schema not found: %s", id)
    }
    
    return nil
}

func (s *PostgresStorage) List(ctx context.Context) ([]string, error) {
    query := `SELECT id FROM schemas ORDER BY updated_at DESC`
    rows, err := s.db.QueryContext(ctx, query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var ids []string
    for rows.Next() {
        var id string
        if err := rows.Scan(&id); err != nil {
            return nil, err
        }
        ids = append(ids, id)
    }
    
    return ids, nil
}

func (s *PostgresStorage) Exists(ctx context.Context, id string) (bool, error) {
    var exists bool
    query := `SELECT EXISTS(SELECT 1 FROM schemas WHERE id = $1)`
    err := s.db.QueryRowContext(ctx, query, id).Scan(&exists)
    return exists, err
}
```

### Usage

```go
import "database/sql"
import _ "github.com/lib/pq"

// Connect to PostgreSQL
db, err := sql.Open("postgres", 
    "host=localhost port=5432 user=user password=pass dbname=erp sslmode=disable")

// Create storage
storage := registry.NewPostgresStorage(db)

// Use with registry
reg := registry.NewRegistry(storage)
```

## Redis Storage

Fast caching layer with optional TTL.

### Implementation

```go
package registry

import (
    "context"
    "fmt"
    "time"
    "github.com/redis/go-redis/v9"
)

type RedisStorage struct {
    client *redis.Client
    prefix string
    ttl    time.Duration
}

func NewRedisStorage(client *redis.Client, ttl time.Duration) *RedisStorage {
    return &RedisStorage{
        client: client,
        prefix: "schema:",
        ttl:    ttl,
    }
}

func (s *RedisStorage) key(id string) string {
    return s.prefix + id
}

func (s *RedisStorage) Get(ctx context.Context, id string) ([]byte, error) {
    data, err := s.client.Get(ctx, s.key(id)).Bytes()
    if err == redis.Nil {
        return nil, fmt.Errorf("schema not found: %s", id)
    }
    if err != nil {
        return nil, fmt.Errorf("redis error: %w", err)
    }
    
    return data, nil
}

func (s *RedisStorage) Set(ctx context.Context, id string, data []byte) error {
    return s.client.Set(ctx, s.key(id), data, s.ttl).Err()
}

func (s *RedisStorage) Delete(ctx context.Context, id string) error {
    result := s.client.Del(ctx, s.key(id))
    if result.Val() == 0 {
        return fmt.Errorf("schema not found: %s", id)
    }
    return result.Err()
}

func (s *RedisStorage) List(ctx context.Context) ([]string, error) {
    keys, err := s.client.Keys(ctx, s.prefix+"*").Result()
    if err != nil {
        return nil, err
    }
    
    // Remove prefix
    ids := make([]string, len(keys))
    for i, key := range keys {
        ids[i] = key[len(s.prefix):]
    }
    
    return ids, nil
}

func (s *RedisStorage) Exists(ctx context.Context, id string) (bool, error) {
    result := s.client.Exists(ctx, s.key(id))
    return result.Val() > 0, result.Err()
}
```

### Usage

```go
import "github.com/redis/go-redis/v9"

// Connect to Redis
rdb := redis.NewClient(&redis.Options{
    Addr:     "localhost:6379",
    Password: "",
    DB:       0,
})

// Create storage with 1 hour TTL
storage := registry.NewRedisStorage(rdb, 1*time.Hour)

// Use as cache
reg := registry.NewRegistry(storage)
```

## Filesystem Storage

Simple file-based storage for development.

### Implementation

```go
package registry

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    "strings"
)

type FilesystemStorage struct {
    basePath string
}

func NewFilesystemStorage(basePath string) *FilesystemStorage {
    // Ensure directory exists
    os.MkdirAll(basePath, 0755)
    
    return &FilesystemStorage{
        basePath: basePath,
    }
}

func (s *FilesystemStorage) filePath(id string) string {
    return filepath.Join(s.basePath, id+".json")
}

func (s *FilesystemStorage) Get(ctx context.Context, id string) ([]byte, error) {
    data, err := os.ReadFile(s.filePath(id))
    if os.IsNotExist(err) {
        return nil, fmt.Errorf("schema not found: %s", id)
    }
    if err != nil {
        return nil, fmt.Errorf("file read error: %w", err)
    }
    
    return data, nil
}

func (s *FilesystemStorage) Set(ctx context.Context, id string, data []byte) error {
    return os.WriteFile(s.filePath(id), data, 0644)
}

func (s *FilesystemStorage) Delete(ctx context.Context, id string) error {
    err := os.Remove(s.filePath(id))
    if os.IsNotExist(err) {
        return fmt.Errorf("schema not found: %s", id)
    }
    return err
}

func (s *FilesystemStorage) List(ctx context.Context) ([]string, error) {
    entries, err := os.ReadDir(s.basePath)
    if err != nil {
        return nil, err
    }
    
    var ids []string
    for _, entry := range entries {
        if entry.IsDir() {
            continue
        }
        if strings.HasSuffix(entry.Name(), ".json") {
            id := strings.TrimSuffix(entry.Name(), ".json")
            ids = append(ids, id)
        }
    }
    
    return ids, nil
}

func (s *FilesystemStorage) Exists(ctx context.Context, id string) (bool, error) {
    _, err := os.Stat(s.filePath(id))
    if os.IsNotExist(err) {
        return false, nil
    }
    if err != nil {
        return false, err
    }
    return true, nil
}
```

### Usage

```go
// Create storage pointing to schemas directory
storage := registry.NewFilesystemStorage("./schemas")

// Use with registry
reg := registry.NewRegistry(storage)
```

## S3/Minio Storage

Cloud object storage for scalability.

### Implementation

```go
package registry

import (
    "bytes"
    "context"
    "fmt"
    "io"
    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Storage struct {
    client     *minio.Client
    bucketName string
    prefix     string
}

func NewS3Storage(endpoint, accessKey, secretKey, bucket string) (*S3Storage, error) {
    client, err := minio.New(endpoint, &minio.Options{
        Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
        Secure: true,
    })
    if err != nil {
        return nil, err
    }
    
    return &S3Storage{
        client:     client,
        bucketName: bucket,
        prefix:     "schemas/",
    }, nil
}

func (s *S3Storage) objectName(id string) string {
    return s.prefix + id + ".json"
}

func (s *S3Storage) Get(ctx context.Context, id string) ([]byte, error) {
    obj, err := s.client.GetObject(ctx, s.bucketName, s.objectName(id), minio.GetObjectOptions{})
    if err != nil {
        return nil, fmt.Errorf("s3 error: %w", err)
    }
    defer obj.Close()
    
    data, err := io.ReadAll(obj)
    if err != nil {
        return nil, fmt.Errorf("read error: %w", err)
    }
    
    if len(data) == 0 {
        return nil, fmt.Errorf("schema not found: %s", id)
    }
    
    return data, nil
}

func (s *S3Storage) Set(ctx context.Context, id string, data []byte) error {
    _, err := s.client.PutObject(
        ctx,
        s.bucketName,
        s.objectName(id),
        bytes.NewReader(data),
        int64(len(data)),
        minio.PutObjectOptions{
            ContentType: "application/json",
        },
    )
    
    return err
}

func (s *S3Storage) Delete(ctx context.Context, id string) error {
    return s.client.RemoveObject(
        ctx,
        s.bucketName,
        s.objectName(id),
        minio.RemoveObjectOptions{},
    )
}

func (s *S3Storage) List(ctx context.Context) ([]string, error) {
    objects := s.client.ListObjects(ctx, s.bucketName, minio.ListObjectsOptions{
        Prefix:    s.prefix,
        Recursive: true,
    })
    
    var ids []string
    for obj := range objects {
        if obj.Err != nil {
            return nil, obj.Err
        }
        
        // Remove prefix and .json suffix
        name := obj.Key
        name = name[len(s.prefix):]
        name = name[:len(name)-5] // remove .json
        ids = append(ids, name)
    }
    
    return ids, nil
}

func (s *S3Storage) Exists(ctx context.Context, id string) (bool, error) {
    _, err := s.client.StatObject(ctx, s.bucketName, s.objectName(id), minio.StatObjectOptions{})
    if err != nil {
        errResp := minio.ToErrorResponse(err)
        if errResp.Code == "NoSuchKey" {
            return false, nil
        }
        return false, err
    }
    return true, nil
}
```

### Usage with AWS S3

```go
// Connect to S3
storage, err := registry.NewS3Storage(
    "s3.amazonaws.com",
    "YOUR_ACCESS_KEY",
    "YOUR_SECRET_KEY",
    "my-schemas-bucket",
)

// Use with registry
reg := registry.NewRegistry(storage)
```

### Usage with Minio

```go
// Connect to Minio
storage, err := registry.NewS3Storage(
    "localhost:9000",
    "minioadmin",
    "minioadmin",
    "schemas",
)

// Use with registry
reg := registry.NewRegistry(storage)
```

## Layered Storage (Recommended)

Combine multiple storage backends for best performance.

```go
type LayeredStorage struct {
    cache   Storage // Redis
    primary Storage // PostgreSQL
}

func NewLayeredStorage(cache, primary Storage) *LayeredStorage {
    return &LayeredStorage{
        cache:   cache,
        primary: primary,
    }
}

func (s *LayeredStorage) Get(ctx context.Context, id string) ([]byte, error) {
    // Try cache first
    data, err := s.cache.Get(ctx, id)
    if err == nil {
        return data, nil
    }
    
    // Cache miss - load from primary
    data, err = s.primary.Get(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Store in cache for next time (async)
    go s.cache.Set(context.Background(), id, data)
    
    return data, nil
}

func (s *LayeredStorage) Set(ctx context.Context, id string, data []byte) error {
    // Write to primary
    if err := s.primary.Set(ctx, id, data); err != nil {
        return err
    }
    
    // Update cache (async)
    go s.cache.Set(context.Background(), id, data)
    
    return nil
}
```

### Usage

```go
// Create storages
postgres := registry.NewPostgresStorage(db)
redis := registry.NewRedisStorage(rdb, 1*time.Hour)

// Combine
storage := NewLayeredStorage(redis, postgres)

// Use with registry
reg := registry.NewRegistry(storage)
```

## Performance Comparison

| Storage | Read (cached) | Read (uncached) | Write | Best For |
|---------|--------------|-----------------|-------|----------|
| **Memory** | <1ms | <1ms | <1ms | Testing |
| **Redis** | 1-5ms | 1-5ms | 1-5ms | Caching |
| **PostgreSQL** | - | 10-50ms | 10-50ms | Primary storage |
| **Filesystem** | 1-10ms | 1-10ms | 1-10ms | Development |
| **S3/Minio** | - | 50-200ms | 50-200ms | Cloud/backup |
| **Layered** | 1-5ms | 10-50ms | 10-50ms | Production |

## Recommendations by Environment

### Development
```go
storage := registry.NewFilesystemStorage("./schemas")
```
- Simple
- Easy to edit schemas
- Version control friendly

### Production (Small)
```go
storage := registry.NewPostgresStorage(db)
```
- ACID guarantees
- Multi-tenant support
- Good enough for <10K requests/min

### Production (Large)
```go
postgres := registry.NewPostgresStorage(db)
redis := registry.NewRedisStorage(rdb, 1*time.Hour)
storage := NewLayeredStorage(redis, postgres)
```
- Fast reads (Redis)
- Reliable writes (PostgreSQL)
- Handles >100K requests/min

### Cloud-Native
```go
s3 := registry.NewS3Storage(endpoint, key, secret, bucket)
redis := registry.NewRedisStorage(rdb, 1*time.Hour)
storage := NewLayeredStorage(redis, s3)
```
- Scalable storage (S3)
- Fast reads (Redis)
- No database needed

## Next Steps

- [10-registry.md](10-registry.md) - Registry implementation
- [04-schema-structure.md](04-schema-structure.md) - Schema format
- [13-testing.md](13-testing.md) - Testing storage implementations

---

[← Back to Layout](07-layout.md) | [Next: Registry →](10-registry.md)