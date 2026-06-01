---
title: Event Bus Internals
portal: 3 — Platform Architecture
section: 04-event-architecture
audience: [architect, backend-engineer, tech-lead]
related:
  - "[Event Architecture Overview](01-event-architecture.md)"
  - "[Outbox Pattern](03-outbox-pattern.md)"
---

# Event Bus Internals

## Interface

```go
// internal/platform/eventbus/eventbus.go
package eventbus

import "context"

type Publisher interface {
    Publish(ctx context.Context, event Event) error
}

type Subscriber interface {
    Subscribe(topic string, handler EventHandler) error
}

type EventHandler func(ctx context.Context, event Event) error
```

## In-Process Implementation (dev/test)

```go
// internal/platform/eventbus/inprocess/bus.go
package inprocess

type Bus struct {
    mu       sync.RWMutex
    handlers map[string][]EventHandler
}

func (b *Bus) Publish(ctx context.Context, event eventbus.Event) error {
    b.mu.RLock()
    handlers := b.handlers[event.Topic]
    b.mu.RUnlock()

    for _, h := range handlers {
        go func(h eventbus.EventHandler) {
            if err := h(ctx, event); err != nil {
                // Log — in-process bus has no retry
                log.Warn().Err(err).Str("topic", event.Topic).Msg("event handler error")
            }
        }(h)
    }
    return nil
}

func (b *Bus) Subscribe(topic string, handler eventbus.EventHandler) error {
    b.mu.Lock()
    defer b.mu.Unlock()
    b.handlers[topic] = append(b.handlers[topic], handler)
    return nil
}
```

## Redis Streams Implementation (production)

```go
// internal/platform/eventbus/redisstreams/bus.go
package redisstreams

type Bus struct {
    client  *redis.Client
    group   string
    logger  zerolog.Logger
}

func (b *Bus) Publish(ctx context.Context, event eventbus.Event) error {
    payload, err := json.Marshal(event)
    if err != nil {
        return fmt.Errorf("marshal event: %w", err)
    }

    return b.client.XAdd(ctx, &redis.XAddArgs{
        Stream: "events:" + event.Topic,
        Values: map[string]interface{}{
            "id":           event.ID,
            "tenant_id":    event.TenantID,
            "published_at": event.PublishedAt.Format(time.RFC3339),
            "payload":      string(payload),
        },
    }).Err()
}
```

## Wire Binding

```go
// Switch transport via Wire binding:

// Development
var EventBusSet = wire.NewSet(
    inprocess.NewBus,
    wire.Bind(new(eventbus.Publisher), new(*inprocess.Bus)),
    wire.Bind(new(eventbus.Subscriber), new(*inprocess.Bus)),
)

// Production
var EventBusSet = wire.NewSet(
    redisstreams.NewBus,
    wire.Bind(new(eventbus.Publisher), new(*redisstreams.Bus)),
    wire.Bind(new(eventbus.Subscriber), new(*redisstreams.Bus)),
)
```

Module code references only `eventbus.Publisher` and `eventbus.Subscriber` — never the concrete type.

## Consumer Group Pattern (Redis Streams)

```
Stream: events:contracts.submitted
    │
    Consumer Group: "awoerp-consumers"
        ├── Consumer: worker-1  (reads and ACKs messages)
        ├── Consumer: worker-2
        └── Consumer: worker-3

Each message delivered to exactly one consumer in the group.
On handler error: message remains pending until ACKed or claimed by another consumer.
```

## Dead Letter Queue

After `maxRetries` failures, the message is moved to `events:{topic}:dlq` with the error attached. Operations teams monitor DLQ depth via Prometheus alert.
