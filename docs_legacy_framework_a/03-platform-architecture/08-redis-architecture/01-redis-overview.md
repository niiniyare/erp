> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Redis Architecture
portal: 3 — Platform Architecture
section: 08-redis-architecture
audience: [architect, backend-engineer, devops]
related:
  - "[Session Architecture](../02-iam/02-session-architecture.md)"
  - "[Event Bus Internals](../04-event-architecture/02-event-bus-internals.md)"
  - "[System Overview](../00-overview/01-system-overview.md)"
---

# Redis Architecture

## Role of Redis

Redis serves two purposes in AwoERP:

| Use | Redis feature | TTL / retention |
|-----|--------------|-----------------|
| Session storage | `SET key value EX ttl` (string) | 24h sliding |
| Event pub/sub (prod) | Redis Streams (`XADD`/`XREAD`) | 7 days (trimmed) |

Redis is **not** used for caching application data. PostgreSQL query performance is sufficient for all current use cases. Cache invalidation complexity is avoided.

## Session Storage

Session tokens are UUIDs stored as keys in Redis:

```
Key:   session:{uuid}
Value: JSON-encoded ResolvedSession
TTL:   24h, sliding (reset on every authenticated request)
```

`ResolvedSession` contains pre-computed data:
- `UserID`, `TenantID`, `Email`
- `Roles []string`
- `EntityScope` (type + entity_id)
- `FeatureFlags map[string]bool`
- `Settings map[string]string`
- `ExpiresAt time.Time`

Pre-computation means Redis gets hit once per request (token lookup), not Postgres.

### Session Lifecycle

```
Login:
  1. Validate credentials (Postgres)
  2. Compute ResolvedSession (roles, entity scope, feature flags)
  3. SETEX session:{uuid} 86400 {json}
  4. Return token to client

Authenticated request:
  1. GET session:{token}   → missing = 401
  2. Check ExpiresAt       → expired = 401
  3. EXPIRE session:{token} 86400   (slide TTL)
  4. Inject session into Fiber locals

Logout:
  1. DEL session:{token}

Force revoke:
  1. SCAN session:* WHERE tenant_id = X → DEL each
```

### Invalidation

Session data changes (role assignment, entity scope change, feature flag toggle) require explicit invalidation. The IAM service calls:

```go
sessionSvc.InvalidateUserSessions(ctx, userID)     // on role change
sessionSvc.InvalidateTenantSessions(ctx, tenantID) // on tenant suspend
```

Both use `SCAN + DEL` which is O(n) but infrequent. User must re-login after invalidation.

## Event Streams (Redis Streams)

Redis Streams replace the in-process event bus in production:

```
Producer (AwoERP service):
  XADD awoerp.contracts.submitted * topic "contracts.submitted" payload "{...}"

Consumer group (AwoERP worker):
  XREADGROUP GROUP awoerp-workers worker-1 COUNT 10 STREAMS awoerp.contracts.submitted >
  → process messages
  XACK awoerp.contracts.submitted awoerp-workers {msg-id}
```

### Stream Naming

```
awoerp.{module}.{topic}

Examples:
  awoerp.contracts.submitted
  awoerp.contracts.approved
  awoerp.finance.transaction_posted
  awoerp.iam.user_created
```

### Consumer Groups

One consumer group per logical subscriber type:

```
awoerp.contracts.submitted → consumers:
  awoerp-audit-workers    (writes audit log)
  awoerp-notif-workers    (sends notifications)
  awoerp-finance-workers  (triggers invoice creation)
```

Each group processes every message independently. Messages are not deleted until all groups acknowledge.

### Stream Trimming

Streams are trimmed to prevent unbounded growth:

```go
// On XADD:
r.client.XAdd(ctx, &redis.XAddArgs{
    Stream: streamName,
    MaxLen: 100_000,   // keep last 100k messages
    Approx: true,      // ~ prefix, more efficient
    Values: payload,
})
```

For guaranteed replay beyond the trim window, use the outbox pattern (messages persisted in Postgres first).

## Connection Configuration

```go
rdb := redis.NewClient(&redis.Options{
    Addr:         cfg.RedisURL,
    Password:     cfg.RedisPassword,
    DB:           0,
    PoolSize:     20,
    MinIdleConns: 5,
    DialTimeout:  3 * time.Second,
    ReadTimeout:  1 * time.Second,
    WriteTimeout: 1 * time.Second,
})
```

## Local vs Production

| Aspect | Local | Production |
|--------|-------|------------|
| Redis | Docker Compose `redis:7` | Redis Cloud or self-hosted Sentinel/Cluster |
| Persistence | None (ephemeral) | AOF + RDB snapshots |
| Auth | No password | `requirepass` in redis.conf |
| TLS | No | Yes (`rediss://` URL) |
| HA | Single node | Sentinel (failover) or Cluster (sharding) |

## Failure Modes

| Failure | Impact | Mitigation |
|---------|--------|-----------|
| Redis down | All authenticated requests fail (session lookup fails) | Redis HA (Sentinel); circuit breaker returning 503 |
| Session evicted (maxmemory) | User gets 401 unexpectedly | Set `maxmemory-policy noeviction` — never evict (prefer OOM kill to silent 401) |
| Stream lag | Events processed late | Consumer group pending count alert; DLQ for stuck messages |
| Redis restart (no persistence) | All sessions lost | Users re-login; acceptable for dev, not prod |
