---
title: "Chapter 41: Redis Usage Patterns"
part: "Part VII — Multi-Tenancy Operations"
chapter: 41
section: "41-redis-usage"
related:
  - "[Chapter 15: Auth and Sessions](../part-03-api/15-auth-sessions.md)"
  - "[Chapter 20: Feature Flags](../part-03-api/20-feature-flags.md)"
  - "[Chapter 13: Middleware Pipeline](../part-03-api/13-middleware-pipeline.md)"
---

# Chapter 41: Redis Usage Patterns

Redis is used throughout Awo for caching, sessions, rate limiting, pub/sub, and deduplication. This chapter consolidates all Redis key patterns, TTL decisions, failure handling, and operational guidance.

---

## 41.1. Key Namespace Inventory

### 41.1.1. Complete Key Schema

All Redis keys are namespaced to prevent collisions:

| Pattern | TTL | Purpose |
|---|---|---|
| `session:{session_id}` | 8h sliding | User session data |
| `user_sessions:{user_id}` | 30d | Set of session IDs for bulk invalidation |
| `flag:{tenant_id}:{flag_name}` | 15m (30s kill switch) | Feature flag evaluation cache |
| `flag_bulk:{tenant_id}:{hash}` | 2m | Bulk flag evaluation response |
| `tenant_config:{tenant_id}` | 5m | Tenant configuration |
| `tenant:{slug}` | 10m | Tenant record by slug lookup |
| `schema_cache:{page_id}:{tenant_id}:{role_hash}:{flag_hash}` | 30m | AMIS page schema |
| `rate:{ip}:{minute_bucket}` | 2m | Rate limiting counter |
| `idempotency:{key}` | 24h | Idempotent request dedup |
| `casbin_policy:{tenant_id}` | until invalidated | Casbin policy snapshot |
| `lock:{resource}:{id}` | varies | Distributed lock |
| `outbox_cursor:{worker_id}` | no expiry | Event outbox processing cursor |

### 41.1.2. Tenant Scoping

All keys that belong to a tenant are prefixed with `{tenant_id}:` in their namespace segment. This allows Redis `SCAN` with a pattern to find all keys for a tenant — used during offboarding to purge tenant data:

```go
func (r *RedisClient) PurgeTenantKeys(ctx context.Context, tenantID uuid.UUID) error {
    pattern := fmt.Sprintf("*%s*", tenantID.String())
    var cursor uint64
    for {
        keys, nextCursor, err := r.client.Scan(ctx, cursor, pattern, 100).Result()
        if err != nil {
            return err
        }
        if len(keys) > 0 {
            r.client.Del(ctx, keys...)
        }
        if nextCursor == 0 {
            break
        }
        cursor = nextCursor
    }
    return nil
}
```

---

## 41.2. Session Storage

### 41.2.1. Session Data Structure

Sessions are stored as JSON hashes:

```go
type SessionData struct {
    UserID     uuid.UUID `json:"user_id"`
    TenantID   uuid.UUID `json:"tenant_id"`
    Roles      []string  `json:"roles"`
    Permissions []string `json:"permissions"`  // pre-computed for fast auth check
    CreatedAt  int64     `json:"created_at"`
    LastSeenAt int64     `json:"last_seen_at"`
    IP         string    `json:"ip"`
    UserAgent  string    `json:"ua"`
    MFAVerified bool     `json:"mfa_verified"`
}
```

### 41.2.2. Sliding Window Expiry

Each request refreshes the session TTL:

```go
func (s *SessionStore) Touch(ctx context.Context, sessionID string) error {
    key := "session:" + sessionID
    pipe := s.redis.Pipeline()
    pipe.Expire(ctx, key, 8*time.Hour)
    // Update last_seen_at
    data, err := s.redis.Get(ctx, key).Result()
    if err != nil {
        return err
    }
    var session SessionData
    json.Unmarshal([]byte(data), &session)
    session.LastSeenAt = time.Now().Unix()
    updated, _ := json.Marshal(session)
    pipe.Set(ctx, key, updated, 8*time.Hour)
    _, err = pipe.Exec(ctx)
    return err
}
```

The pipeline batches the two operations (expire + set) into one round-trip.

### 41.2.3. Absolute Session Limit

Sliding expiry alone would allow a session to last indefinitely for an active user. The absolute limit (24h) is enforced by comparing `created_at`:

```go
func (s *SessionStore) Validate(ctx context.Context, sessionID string) (*SessionData, error) {
    session, err := s.Get(ctx, sessionID)
    if err != nil {
        return nil, ErrSessionNotFound
    }
    // Absolute expiry check
    if time.Since(time.Unix(session.CreatedAt, 0)) > 24*time.Hour {
        s.Delete(ctx, sessionID)
        return nil, ErrSessionExpired
    }
    s.Touch(ctx, sessionID)
    return session, nil
}
```

---

## 41.3. Rate Limiting

### 41.3.1. Sliding Window Algorithm

Awo uses a Redis sorted-set sliding window rather than a fixed counter. This avoids the burst problem with fixed windows (100 requests at 23:59:59 + 100 at 00:00:01 = 200 in 2 seconds).

```go
func (r *RateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, int, error) {
    now := time.Now()
    windowStart := now.Add(-window)

    pipe := r.redis.Pipeline()
    // Remove old entries
    pipe.ZRemRangeByScore(ctx, key,
        "-inf",
        strconv.FormatFloat(float64(windowStart.UnixNano()), 'f', 0, 64))
    // Count current entries
    countCmd := pipe.ZCard(ctx, key)
    // Add current request
    pipe.ZAdd(ctx, key, redis.Z{
        Score:  float64(now.UnixNano()),
        Member: now.UnixNano(),
    })
    pipe.Expire(ctx, key, window+time.Second)
    _, err := pipe.Exec(ctx)
    if err != nil {
        return true, limit, err  // fail open on Redis error
    }

    count := int(countCmd.Val())
    if count >= limit {
        return false, 0, nil
    }
    return true, limit - count - 1, nil
}
```

**Fail open**: if Redis is unavailable, rate limiting is bypassed (`return true`). The alternative (fail closed) would deny all traffic during a Redis outage. For a multi-tenant ERP, availability is usually more important than perfect rate limiting during an infrastructure incident.

### 41.3.2. Per-Tenant Rate Limit Overrides

Enterprise tenants can have higher rate limits configured:

```go
func rateLimitKey(c *fiber.Ctx, cfg *TenantConfig) (string, int) {
    tenantID := tenant.IDFromCtx(c)
    limit := cfg.RateLimitPerMinute
    if limit == 0 {
        limit = defaultRateLimitPerMinute  // 300
    }
    return fmt.Sprintf("rate:%s:%s:%d",
        tenantID, c.IP(), time.Now().Unix()/60), limit
}
```

---

## 41.4. Pub/Sub — Permission Propagation

### 41.4.1. Why Pub/Sub for Casbin?

Casbin's `SyncedEnforcer` caches policies in memory. When an admin adds a role assignment, the new policy must propagate to all server instances (Awo runs multiple replicas). Without propagation, the change only takes effect on the instance that processed the update, and only after the next cache refresh.

Redis pub/sub provides near-instant propagation:

```go
// Publisher — called after role/permission update
func (s *IAMService) PublishPolicyChange(ctx context.Context, tenantID uuid.UUID) error {
    return s.redis.Publish(ctx, "casbin:policy_updated",
        tenantID.String()).Err()
}

// Subscriber — runs on each server instance
func (s *IAMService) StartPolicySubscriber(ctx context.Context) {
    sub := s.redis.Subscribe(ctx, "casbin:policy_updated")
    defer sub.Close()
    for msg := range sub.Channel() {
        tenantID, _ := uuid.Parse(msg.Payload)
        s.enforcer.LoadPolicy()  // reload from DB
        s.redis.Del(ctx, fmt.Sprintf("casbin_policy:%s", tenantID))
    }
}
```

### 41.4.2. AMIS Schema Cache Invalidation

When a tenant configuration changes (new role added, module enabled), all cached AMIS page schemas for that tenant must be invalidated:

```go
func (s *SchemaCache) InvalidateTenant(ctx context.Context, tenantID uuid.UUID) error {
    pattern := fmt.Sprintf("schema_cache:*:%s:*", tenantID)
    // SCAN + DEL is safer than KEYS * in production
    return scanAndDelete(ctx, s.redis, pattern)
}
```

---

## 41.5. Distributed Locks

### 41.5.1. When to Use Locks

Locks prevent concurrent execution of operations that must not run in parallel:
- Payroll run for the same tenant+month: two concurrent runs would double-post GL entries
- Physical stock count: concurrent stock movements during a count corrupt the count
- Period close: concurrent period closes cause duplicate carry-forward entries

### 41.5.2. Lock Pattern

```go
func (l *Locker) Acquire(ctx context.Context, resource string, ttl time.Duration) (bool, func(), error) {
    key := "lock:" + resource
    token := uuid.New().String()

    // SET NX EX — atomic set if not exists with expiry
    ok, err := l.redis.SetNX(ctx, key, token, ttl).Result()
    if err != nil {
        return false, nil, err
    }
    if !ok {
        return false, nil, nil  // lock held by another process
    }

    release := func() {
        // Only delete if we still own the lock (compare-and-delete via Lua)
        script := `
            if redis.call("GET", KEYS[1]) == ARGV[1] then
                return redis.call("DEL", KEYS[1])
            end
            return 0
        `
        l.redis.Eval(ctx, script, []string{key}, token)
    }

    return true, release, nil
}
```

**Compare-and-delete**: The Lua script ensures only the lock owner can release it. Without this, if the lock expired and another process acquired it, the original process finishing late would inadvertently release the new owner's lock.

**Usage**:
```go
acquired, release, err := locker.Acquire(ctx,
    fmt.Sprintf("payroll:%s:%s", tenantID, month), 30*time.Minute)
if !acquired {
    return errs.NewConflictError("PAYROLL_RUNNING",
        "payroll run for %s is already in progress", month)
}
defer release()
// proceed with payroll
```

---

## 41.6. Redis Failure Handling

### 41.6.1. Circuit Breaker

All Redis calls are wrapped in a circuit breaker. If Redis is consistently failing, the circuit opens and calls return the fallback immediately without waiting:

```go
type RedisClient struct {
    client  *redis.Client
    breaker *gobreaker.CircuitBreaker
}

func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
    result, err := r.breaker.Execute(func() (interface{}, error) {
        return r.client.Get(ctx, key).Result()
    })
    if err == gobreaker.ErrOpenState {
        return "", ErrCircuitOpen
    }
    return result.(string), err
}
```

### 41.6.2. Graceful Degradation Per Use Case

| Feature | Redis unavailable | Behaviour |
|---|---|---|
| Sessions | Returns session-not-found | User must re-login (acceptable) |
| Feature flags | Returns global default | Flags behave as if no overrides |
| Rate limiting | Fail open | No rate limiting (acceptable for brief outage) |
| AMIS schema | Returns empty cache | Schema rebuilt from scratch per request |
| Casbin | Reads from DB directly | Slower but correct |
| Idempotency | Returns no-prior-result | May process duplicate requests |

The degradation policy is explicit: most features work without Redis, just slower or less restricted. The only critical failure is sessions — users are logged out. This is acceptable; the alternative (keeping a local in-memory session cache) creates security risks (session not invalidated after password change).

### 41.6.3. Redis Health Check

The `/health/ready` endpoint checks Redis:

```go
func (h *HealthHandler) Ready(c *fiber.Ctx) error {
    if err := h.redis.Ping(c.Context()).Err(); err != nil {
        return c.Status(503).JSON(fiber.Map{
            "status": "not_ready",
            "reason": "redis_unavailable",
        })
    }
    return c.JSON(fiber.Map{"status": "ready"})
}
```

When Redis is down, the readiness probe fails, and the load balancer stops sending traffic to the instance. Kubernetes restarts the pod after the configured liveness probe timeout. This prevents a cascading failure where all instances are serving degraded traffic indefinitely.
