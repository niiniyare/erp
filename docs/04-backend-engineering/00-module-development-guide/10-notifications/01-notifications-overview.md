---
title: Notifications Overview
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[Notification Channels](06-notification-channels.md)"
  - "[Service Layer](../06-service-layer/01-service-overview.md)"
  - "[Event-Driven Patterns](../13-event-driven/06-event-patterns-reference.md)"
---

# Notifications Overview

The notification platform delivers real-time and deferred messages to users when contract lifecycle events occur. Modules publish to the notification service; they do not send notifications directly.

## Architecture

```
Service Layer
    │
    ▼
NotificationService interface  ←── injected via Wire
    │
    ▼
internal/core/notifications/service
    │
    ├── real-time: WebSocket push (Fiber WebSocket)
    ├── in-app: stored in DB, polled by UI
    └── email: queued to outbox, delivered async
```

## Notification vs. Event

| Concern | Domain Event | Notification |
|---------|-------------|-------------|
| Audience | Internal bus (services) | End users (UI, email) |
| Format | Go struct | Human-readable message |
| Storage | Event store / outbox | `notifications` table |
| Trigger | Always fired | Conditional on preferences |

## NotificationService Interface

```go
// internal/core/notifications/port/service.go
package port

import (
    "context"
    "github.com/google/uuid"
)

type NotificationService interface {
    // Notify delivers a notification to one or more recipients.
    Notify(ctx context.Context, n Notification) error

    // NotifyTenant delivers to all users with a given role in a tenant.
    NotifyTenant(ctx context.Context, tenantID uuid.UUID, roleFilter string, n Notification) error
}
```

## Notification Struct

```go
// internal/core/notifications/port/notification.go
package port

import "github.com/google/uuid"

// Audience describes who receives the notification.
type Audience string

const (
    AudienceUser   Audience = "user"   // single user
    AudienceRole   Audience = "role"   // all users with role in tenant
    AudienceTenant Audience = "tenant" // all users in tenant
)

type Notification struct {
    TenantID   uuid.UUID
    Audience   Audience
    RecipientID uuid.UUID  // UserID when Audience == AudienceUser
    Role       string      // role name when Audience == AudienceRole
    Category   string      // "contract.submitted", "contract.approved", ...
    Title      string
    Body       string
    ResourceID uuid.UUID   // contract ID — enables deep-link in UI
    ResourceType string    // "contract"
    Priority   Priority
    Metadata   map[string]string
}

type Priority string

const (
    PriorityLow    Priority = "low"
    PriorityNormal Priority = "normal"
    PriorityHigh   Priority = "high"
    PriorityUrgent Priority = "urgent"
)
```

## When to Notify

Notify on transitions that require **human action** or **awareness**:

| Event | Notify? | Audience |
|-------|---------|---------|
| Contract created (draft) | No | — internal, not actionable yet |
| Contract submitted | Yes | Reviewers role |
| Contract under review | Yes | Submitter (acknowledgement) |
| Contract approved | Yes | Submitter |
| Contract rejected → draft | Yes | Submitter |
| Contract activated | Yes | Contract owner |
| Contract suspended | Yes | Contract owner + admin |
| Contract terminated | Yes | Contract owner + admin |

## Async Pattern

Always fire notifications asynchronously — network latency must not block the HTTP response:

```go
func (s *contractService) publishNotificationAsync(n port.Notification) {
    go func() {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        defer func() {
            if r := recover(); r != nil {
                s.log.Error().Interface("panic", r).Msg("notification panic recovered")
            }
        }()

        if err := s.notifSvc.Notify(ctx, n); err != nil {
            s.log.Warn().Err(err).Str("category", n.Category).Msg("notification delivery failed")
        }
    }()
}
```

Notification failures are **logged**, never propagated to the caller.
