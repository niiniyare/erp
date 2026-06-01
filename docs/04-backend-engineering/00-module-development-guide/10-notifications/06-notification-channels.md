---
title: Notification Channels
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer]
related:
  - "[Notifications Overview](01-notifications-overview.md)"
  - "[Notification Types](02-notification-types.md)"
  - "[Sending Notifications](03-sending-notifications.md)"
  - "[Notification Preferences](04-notification-preferences.md)"
---

# Notification Channels

## Available Channels

| Channel | Use case | Delivery |
|---------|---------|---------|
| In-app | Real-time alerts visible in the UI | WebSocket push or polling |
| Email | Async summaries, approval requests | SMTP / SendGrid |
| Webhook | System integrations | HTTP POST to tenant-configured URL |

## Channel Selection

The `NotificationService` routes to the appropriate channel based on:
1. Notification category
2. User's preference for that category
3. Tenant-level channel configuration

```go
type NotificationService interface {
    Send(ctx context.Context, n Notification) error
}

type Notification struct {
    TenantID    uuid.UUID
    UserID      uuid.UUID           // recipient
    Category    NotificationCategory
    Title       string
    Body        string
    ResourceID  uuid.UUID
    ResourceURL string
    Metadata    map[string]string
}
```

## In-App Notifications

Stored in `notifications` table, surfaced via `GET /api/v1/notifications`:

```sql
CREATE TABLE notifications (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    uuid NOT NULL REFERENCES tenants(id),
    user_id      uuid NOT NULL,
    category     text NOT NULL,
    title        text NOT NULL,
    body         text NOT NULL,
    resource_id  uuid,
    resource_url text,
    read_at      timestamptz,
    created_at   timestamptz NOT NULL DEFAULT now()
);
```

### Mark as Read

```
PATCH /api/v1/notifications/:id/read
POST  /api/v1/notifications/read-all
```

### Unread Count

```
GET /api/v1/notifications/unread-count
```

Response:
```json
{ "count": 5 }
```

Polled by the frontend every 30 seconds to update badge count.

## Email Notifications

Email is sent via the `EmailSender` interface:

```go
type EmailSender interface {
    Send(ctx context.Context, msg EmailMessage) error
}

type EmailMessage struct {
    To      []string
    Subject string
    HTML    string
    Text    string
}
```

Email content is rendered from templates in `internal/shared/notifications/templates/`:

```
templates/
  contracts/
    submitted.html
    submitted.txt
    approved.html
    approved.txt
  iam/
    user_invited.html
    user_invited.txt
```

Templates receive the `Notification` struct as data.

### Email Preferences

Users can opt out of email per category. The service checks before sending:

```go
func (s *notificationService) Send(ctx context.Context, n Notification) error {
    prefs, err := s.prefRepo.GetPreferences(ctx, n.TenantID, n.UserID)
    if err != nil {
        return err
    }

    // Always send in-app
    if err := s.inApp.Store(ctx, n); err != nil {
        s.logger.ErrorContext(ctx, "in-app notify failed", "error", err)
    }

    // Email only if opted in
    if prefs.EmailEnabled(n.Category) {
        if err := s.email.Send(ctx, buildEmail(n)); err != nil {
            s.logger.ErrorContext(ctx, "email send failed", "error", err)
        }
    }

    return nil
}
```

Email failures are logged but do not fail the notification call — delivery is best-effort.

## Webhook Notifications

Tenant-configured webhooks receive POST requests for specified categories:

```json
{
  "event": "contract.submitted",
  "tenant_id": "...",
  "resource_id": "...",
  "data": {
    "contract_number": "CONT-2025-0001",
    "title": "...",
    "submitted_by": "..."
  },
  "timestamp": "2025-01-15T10:30:00Z"
}
```

Webhook configuration is per-tenant, managed via the Tenant API (`POST /api/v1/tenants/:id/webhooks`).

Delivery includes:
- `X-AwoERP-Signature` header — HMAC-SHA256 of the payload body using the tenant's webhook secret
- 3 retry attempts with exponential backoff
- 30-second timeout per delivery attempt

## Adding a New Channel

1. Implement the channel interface:

```go
type SMSChannel interface {
    Send(ctx context.Context, msg SMSMessage) error
}
```

2. Register in the `NotificationService` via Wire

3. Add channel preference to `notification_preferences` table

4. Update the `Send` method to route to the new channel when opted in

5. Document the channel in the Tenant API reference
