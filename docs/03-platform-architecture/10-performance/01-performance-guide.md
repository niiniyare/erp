---
title: Performance Guide
portal: 3 — Platform Architecture
section: 03-platform-architecture
audience: [architect, backend-engineer, sre]
related:
  - "[Query Patterns](../03-data-architecture/04-query-patterns.md)"
  - "[Observability Overview](../05-observability/01-observability-overview.md)"
  - "[Redis Architecture](../08-redis-architecture/01-redis-overview.md)"
---

# Performance Guide

## Performance Targets

| Metric | Target | Alert threshold |
|--------|--------|----------------|
| p50 response time | < 50ms | — |
| p99 response time | < 500ms | > 2s |
| p999 response time | < 2s | — |
| DB query time (simple) | < 5ms | — |
| DB query time (complex) | < 50ms | — |
| Throughput | 500 req/s per pod | — |

## Database Performance

### Index Strategy

Every tenant-scoped query must be covered by an index. Standard indexes per table:

```sql
-- Required: tenant_id + status filter
CREATE INDEX idx_{table}_tenant_status
    ON {table} (tenant_id, status)
    WHERE deleted_at IS NULL;

-- Required: tenant_id + created_at for time-range queries
CREATE INDEX idx_{table}_tenant_created
    ON {table} (tenant_id, created_at DESC)
    WHERE deleted_at IS NULL;

-- Optional: tenant_id + frequently filtered FK
CREATE INDEX idx_{table}_tenant_{fk}
    ON {table} (tenant_id, {fk_column})
    WHERE deleted_at IS NULL;
```

Partial indexes (`WHERE deleted_at IS NULL`) exclude soft-deleted rows — smaller, faster.

### Avoid N+1 Queries

Never query in a loop. Use `IN` with arrays or `JSON_AGG` for nested data:

```sql
-- ❌ N+1 pattern (do not do this)
-- for each contract:
--   SELECT * FROM contract_lines WHERE contract_id = $1

-- ✅ Single query with JSON aggregation
SELECT c.*, json_agg(cl.*) AS lines
FROM contracts c
LEFT JOIN contract_lines cl ON cl.contract_id = c.id AND cl.deleted_at IS NULL
WHERE c.id = ANY(@ids::uuid[])
GROUP BY c.id;
```

### Connection Pool Sizing

```go
// internal/platform/db/pool.go
pool, _ := pgxpool.NewWithConfig(ctx, &pgxpool.Config{
    ConnConfig:            connConfig,
    MaxConns:              int32(cfg.MaxConns),     // default: 20
    MinConns:              int32(cfg.MinConns),     // default: 5
    MaxConnLifetime:       30 * time.Minute,
    MaxConnIdleTime:       5 * time.Minute,
    HealthCheckPeriod:     1 * time.Minute,
})
```

Pool sizing formula: `max_conns = num_pods × max_conns_per_pod < postgresql_max_connections`

For PostgreSQL with `max_connections = 200` and 3 pods: `max_conns_per_pod = 200 / 3 - 10 (reserve) ≈ 55`

### EXPLAIN ANALYZE

For slow queries, capture the plan:

```sql
EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON)
SELECT * FROM contracts
WHERE tenant_id = '...'
  AND status = 'active'
  AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT 20;
```

Key metrics:
- `actual time` > `planning time` by 10x → query complexity OK
- `Seq Scan` on large table → missing index
- `Rows Removed by Filter` >> `Rows` → index selectivity poor

## Application Performance

### Avoid Allocations in Hot Paths

```go
// ❌ Allocates new slice every call
func tenantFilter(id uuid.UUID) []byte {
    return []byte(id.String())
}

// ✅ Pre-allocate once; use sync.Pool for very hot paths
var uuidBuf [36]byte
```

### Pagination

Always paginate list endpoints. Hard cap at 100 items per page:

```go
func sanitizePageSize(requested int) int {
    if requested <= 0 {
        return 20
    }
    if requested > 100 {
        return 100
    }
    return requested
}
```

`COUNT(*) OVER()` in the SQLC query avoids a separate count query:

```sql
SELECT *, COUNT(*) OVER() AS total_count
FROM contracts
WHERE tenant_id = current_setting('app.tenant_id')::uuid
  AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT @limit_ OFFSET @offset_;
```

### Response Payload Size

- Never return full DB rows — map through response DTOs
- Omit `null` fields in JSON with `omitempty`
- For large lists, return only summary fields (not full content) — detail endpoint for single items
- Compress responses with gzip for responses > 1KB:

```go
// In Fiber app initialization
app.Use(compress.New(compress.Config{
    Level: compress.LevelBestSpeed,
}))
```

## Caching

### Session Cache

Sessions are pre-computed at login and stored in Redis. No DB hit on session lookup.

See: [Session Architecture](../02-iam/02-session-architecture.md)

### Feature Flags

The materialized view `mv_tenant_feature_flags_cache` is queried at session creation time. No per-request DB lookup.

See: [Feature Flags](../01-multi-tenancy/04-feature-flags.md)

### Read-Through Cache Pattern

For expensive read operations called frequently:

```go
func (s *contractService) GetContractStats(ctx context.Context, tenantID uuid.UUID) (*Stats, error) {
    key := fmt.Sprintf("stats:contracts:%s", tenantID)

    // Try cache first
    if cached, err := s.cache.Get(ctx, key); err == nil {
        var stats Stats
        if err := json.Unmarshal([]byte(cached), &stats); err == nil {
            return &stats, nil
        }
    }

    // Cache miss — compute from DB
    stats, err := s.repo.ComputeStats(ctx, tenantID)
    if err != nil {
        return nil, err
    }

    // Write to cache with TTL
    b, _ := json.Marshal(stats)
    s.cache.Set(ctx, key, string(b), 5*time.Minute)
    return stats, nil
}
```

Invalidate on write:
```go
s.cache.Del(ctx, fmt.Sprintf("stats:contracts:%s", tenantID))
```

## Profiling

```bash
# Enable pprof endpoint (only in staging/dev)
import _ "net/http/pprof"
go func() { http.ListenAndServe(":6060", nil) }()

# CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Goroutine dump (check for leaks)
go tool pprof http://localhost:6060/debug/pprof/goroutine

# Memory heap
go tool pprof http://localhost:6060/debug/pprof/heap
```

Goroutine leak check — look for large numbers of goroutines in `contract.service.*` functions.

## Load Testing

Baseline test before any major release:

```bash
# Using k6
k6 run --vus 50 --duration 60s scripts/load/contracts-list.js
```

Target: 50 concurrent users, p99 < 500ms, zero 5xx errors.
