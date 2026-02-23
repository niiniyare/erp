# Chapter 19: Event-Driven Integration

[← Inventory Integration](./18-inventory-integration.md) | [Next: Search & Analysis →](./20-search-and-analysis.md)

---

## Overview

Configuration changes propagate through the system via Redis Streams. When a configuration value is updated, an event is published to the stream. Consuming services subscribe to these events to invalidate caches, recalculate derived values, and trigger workflows.

---

## Configuration Event Bus

```go
type ConfigurationEventBus struct {
    redisClient *redis.Client
    subscribers map[string][]ConfigurationEventHandler
}

// Event published on every configuration write
type ConfigurationChangedEvent struct {
    ID        string                 `json:"id"`
    Type      string                 `json:"type"`       // "configuration.changed"
    Time      time.Time              `json:"timestamp"`
    TenantID  string                 `json:"tenant_id"`
    EntityID  *string                `json:"entity_id,omitempty"`
    Module    string                 `json:"module"`
    ConfigKey string                 `json:"config_key"`
    OldValue  interface{}            `json:"old_value"`
    NewValue  interface{}            `json:"new_value"`
    Source    string                 `json:"source"`     // "tenant" | "entity" | "template"
    ChangedBy string                 `json:"changed_by"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Event published on template application
type TemplateAppliedEvent struct {
    ID             string    `json:"id"`
    TenantID       string    `json:"tenant_id"`
    TemplateID     string    `json:"template_id"`
    TemplateName   string    `json:"template_name"`
    TargetType     string    `json:"target_type"`
    TargetID       string    `json:"target_id"`
    AppliedConfigs int       `json:"applied_configs"`
    ConflictCount  int       `json:"conflict_count"`
    AppliedBy      string    `json:"applied_by"`
}
```

---

## Publishing Events

```go
func (c *ConfigurationEventBus) PublishConfigurationChange(
    ctx context.Context, change ConfigurationChangedEvent,
) error {
    data, err := json.Marshal(change)
    if err != nil {
        return err
    }

    return c.redisClient.XAdd(ctx, &redis.XAddArgs{
        Stream: "settings:configuration:changed",
        Values: map[string]interface{}{
            "tenant_id":  change.TenantID,
            "event_type": "configuration.changed",
            "module":     change.Module,
            "config_key": change.ConfigKey,
            "data":       string(data),
            "timestamp":  change.Timestamp().Unix(),
        },
    }).Err()
}
```

---

## Subscribing to Events

```go
func (c *ConfigurationEventBus) Subscribe(
    ctx context.Context, pattern string, handler ConfigurationEventHandler,
) error {
    consumerGroup := fmt.Sprintf("settings-consumer-%s", handler.GetSubscriptionPattern())
    streamKey := "settings:configuration:changed"

    // Create consumer group (idempotent)
    c.redisClient.XGroupCreateMkStream(ctx, streamKey, consumerGroup, "0")

    go func() {
        for {
            select {
            case <-ctx.Done():
                return
            default:
                streams, err := c.redisClient.XReadGroup(ctx, &redis.XReadGroupArgs{
                    Group:    consumerGroup,
                    Consumer: "settings-handler",
                    Streams:  []string{streamKey, ">"},
                    Count:    10,
                    Block:    time.Second,
                }).Result()

                if err != nil {
                    continue
                }

                for _, stream := range streams {
                    for _, message := range stream.Messages {
                        if err := c.processMessage(ctx, handler, message); err != nil {
                            continue // move to dead letter queue in production
                        }
                        c.redisClient.XAck(ctx, streamKey, consumerGroup, message.ID)
                    }
                }
            }
        }
    }()

    return nil
}
```

---

## Cache Invalidation on Change

The high-performance configuration client invalidates cache entries when a change event arrives:

```go
func (c *ConfigurationClient) HandleConfigurationChange(
    ctx context.Context, event ConfigurationChangedEvent,
) error {
    // Invalidate the specific key
    cacheKey := c.buildCacheKey(event.TenantID, event.EntityID, event.Module, event.ConfigKey)
    c.cache.Delete(cacheKey)

    // Invalidate related batch cache entries
    c.invalidateRelatedCache(event.TenantID, event.EntityID, event.Module)

    // Notify local subscribers
    c.mu.RLock()
    defer c.mu.RUnlock()

    change := ConfigurationChange{
        TenantID:  event.TenantID,
        Module:    event.Module,
        ConfigKey: event.ConfigKey,
        OldValue:  event.OldValue,
        NewValue:  event.NewValue,
    }

    for pattern, subscribers := range c.subscriptions {
        if c.matchesPattern(pattern, event.Module, event.ConfigKey) {
            for _, sub := range subscribers {
                go sub.OnConfigurationChanged(ctx, change)
            }
        }
    }

    return nil
}
```

---

## Cross-Service Synchronization

For distributed deployments, the `ConfigurationSynchronizer` coordinates cache invalidation and service notifications:

```go
func (c *ConfigurationSynchronizer) SynchronizeConfigurationChange(
    ctx context.Context, change ConfigurationChangedEvent,
) error {
    // 1. Invalidate distributed caches
    patterns := []string{
        fmt.Sprintf("config:*:%s:%s:%s", change.TenantID, change.Module, change.ConfigKey),
        fmt.Sprintf("batch:config:%s:*", change.TenantID),
    }
    for _, pattern := range patterns {
        c.cacheManager.InvalidatePattern(ctx, pattern)
    }

    // 2. Notify affected services via webhook
    affected := c.identifyAffectedServices(change.Module, change.ConfigKey)
    for _, svc := range affected {
        go c.notificationSvc.NotifyService(ctx, ServiceNotification{
            ServiceName:      svc,
            NotificationType: "configuration_changed",
            Payload:          change,
        })
    }

    // 3. Trigger dependent configuration recalculations
    for _, dep := range c.identifyDependentConfigurations(change) {
        go c.recalculateDependentConfiguration(ctx, dep, change)
    }

    return nil
}
```

---

## Service Dependency Map

```go
var serviceDependencies = map[string]map[string][]string{
    "finance": {
        "default_currency":       {"finance-service", "reporting-service", "analytics-service"},
        "invoice_prefix":         {"finance-service", "document-service"},
        "auto_approval_limit":    {"finance-service", "workflow-service"},
        "multi_currency_enabled": {"finance-service", "exchange-service"},
    },
    "hr": {
        "default_pay_frequency":   {"payroll-service", "hr-service"},
        "overtime_threshold":      {"payroll-service", "time-tracking-service"},
        "benefits_waiting_period": {"hr-service", "benefits-service"},
    },
    "inventory": {
        "default_valuation_method": {"inventory-service", "accounting-service"},
        "auto_reorder_enabled":     {"inventory-service", "purchasing-service"},
        "lot_tracking_required":    {"inventory-service", "warehouse-service"},
    },
}
```

---

[← Inventory Integration](./18-inventory-integration.md) | [Next: Search & Analysis →](./20-search-and-analysis.md)
