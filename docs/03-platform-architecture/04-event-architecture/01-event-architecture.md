---
title: Event Architecture
portal: 3 — Platform Architecture
section: 04-event-architecture
audience: [architect, backend-engineer, tech-lead]
related:
  - "[Redis Architecture](../08-redis-architecture/01-redis-overview.md)"
  - "[Temporal Architecture](../07-temporal-architecture/01-temporal-overview.md)"
  - "[Event Patterns Reference](../../04-backend-engineering/00-module-development-guide/13-event-driven/06-event-patterns-reference.md)"
---

# Event Architecture

AwoERP uses a two-mode event system: **in-process** for development/testing and **Redis Streams** for production.

## Event Flow

```
Service Layer
    │
    ▼ publishAsync (goroutine)
eventbus.Publisher interface
    │
    ├── InMemoryBus (dev/test)
    │   └── Direct function call to subscribers
    │
    └── RedisStreamBus (production)
        └── XADD to stream
            └── Consumer group reads
                └── Handler function
```

## Domain Event Interface

Every publishable event implements:

```go
// internal/platform/events/event.go
type Event interface {
    Topic() string           // e.g. "contracts.submitted"
    GetTenantID() uuid.UUID  // for tenant-scoped routing
}
```

## Topic Naming

`{module}.{noun}.{past_tense_verb}`

| Topic | Emitted when |
|-------|-------------|
| `contracts.contract.created` | Contract created |
| `contracts.contract.submitted` | Contract submitted for review |
| `contracts.contract.approved` | Contract approved |
| `contracts.contract.activated` | Contract activated |
| `contracts.contract.terminated` | Contract terminated |
| `finance.transaction.posted` | Financial transaction posted |
| `iam.user.created` | User registered |
| `iam.session.created` | User logged in |
| `tenant.tenant.activated` | Tenant activated |

## Publisher Interface

```go
// internal/platform/eventbus/bus.go
type Publisher interface {
    Publish(ctx context.Context, event PublishRequest) error
}

type PublishRequest struct {
    ID          string          // uuid — deduplication key
    Topic       string
    TenantID    string
    PublishedAt time.Time
    Payload     json.RawMessage
}
```

## Subscriber Interface

```go
type Subscriber interface {
    Subscribe(topic string, handler HandlerFunc) error
}

type HandlerFunc func(ctx context.Context, msg Message) error

type Message struct {
    ID          string
    Topic       string
    TenantID    string
    PublishedAt time.Time
    Payload     json.RawMessage
    Attempt     int
}
```

## Redis Streams Implementation

### Publishing

```go
// internal/platform/eventbus/redis_bus.go
func (b *redisStreamBus) Publish(ctx context.Context, req PublishRequest) error {
    payload, err := json.Marshal(req)
    if err != nil {
        return fmt.Errorf("eventbus: marshal failed: %w", err)
    }
    stream := streamName(req.Topic)   // "events:{topic}"
    return b.client.XAdd(ctx, &redis.XAddArgs{
        Stream: stream,
        MaxLen: 10000,  // keep last 10k events per stream
        Approx: true,
        Values: map[string]any{
            "payload": string(payload),
        },
    }).Err()
}

func streamName(topic string) string {
    return "events:" + topic
}
```

### Consuming

Consumer groups ensure each event is processed by exactly one instance in a replica set:

```go
func (b *redisStreamBus) consume(ctx context.Context, topic string, handler HandlerFunc) {
    stream := streamName(topic)
    group  := "awoerp"   // one group per application

    // Create group (idempotent)
    b.client.XGroupCreateMkStream(ctx, stream, group, "$")

    for {
        results, err := b.client.XReadGroup(ctx, &redis.XReadGroupArgs{
            Group:    group,
            Consumer: b.consumerID,   // pod name or uuid
            Streams:  []string{stream, ">"},
            Count:    10,
            Block:    5 * time.Second,
        }).Result()

        for _, result := range results {
            for _, msg := range result.Messages {
                b.dispatch(ctx, msg, handler)
                b.client.XAck(ctx, stream, group, msg.ID)
            }
        }
    }
}
```

## Outbox Pattern for Critical Events

For events that must be delivered even if Redis is temporarily unavailable, use the transactional outbox:

```sql
CREATE TABLE event_outbox (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   uuid NOT NULL,
    topic       text NOT NULL,
    payload     jsonb NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    published_at timestamptz,
    attempts    integer NOT NULL DEFAULT 0,
    last_error  text
);
```

Write the event to `event_outbox` within the same transaction as the business operation. A relay goroutine reads pending events and publishes them to Redis Streams.

```go
// In repository — same transaction
payload, _ := json.Marshal(event)
q.InsertOutboxEvent(ctx, sqlc.InsertOutboxEventParams{
    TenantID: tenantID,
    Topic:    event.Topic(),
    Payload:  payload,
})
```

## When to Use Outbox vs Async Goroutine

| Pattern | Use when |
|---------|---------|
| Async goroutine | Event loss is tolerable (notifications, metrics side effects) |
| Outbox (transactional) | Event must be delivered — downstream systems depend on it |

## Dead Letter Queue

Failed events after 3 attempts go to the DLQ stream:

```
events:{topic}:dlq
```

Ops runbook: [Event Outbox Stuck](../../09-operations/07-event-outbox.md)

## In-Memory Bus (Dev/Test)

```go
type inMemoryBus struct {
    mu       sync.RWMutex
    handlers map[string][]HandlerFunc
}

func (b *inMemoryBus) Publish(ctx context.Context, req PublishRequest) error {
    b.mu.RLock()
    handlers := b.handlers[req.Topic]
    b.mu.RUnlock()

    var msg Message
    json.Unmarshal(req.Payload, &msg)
    msg.Topic = req.Topic
    msg.TenantID = req.TenantID

    for _, h := range handlers {
        if err := h(ctx, msg); err != nil {
            return err
        }
    }
    return nil
}
```

Wire binds the correct implementation based on environment:

```go
// In platform wire.go
func ProvideEventBus(cfg Config, redis *redis.Client) eventbus.Publisher {
    if cfg.IsDev() {
        return eventbus.NewInMemoryBus()
    }
    return eventbus.NewRedisStreamBus(redis, cfg.PodName)
}
```
