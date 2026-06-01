---
title: Notification Preferences
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[Sending Notifications](03-sending-notifications.md)"
  - "[Notification Channels](06-notification-channels.md)"
---

# Notification Preferences

Users can configure per-category delivery preferences (in-app, email, both, or none). The notification service respects preferences automatically — modules do not need to check them.

## Preference Resolution

The notification service resolves delivery channels before sending:

```
Notify(n) called
    │
    ▼
Load user preferences for n.Category
    │
    ├── channel: in_app  → write to notifications table
    ├── channel: email   → write to email outbox
    ├── channel: both    → write to both
    └── channel: none    → drop silently
```

Modules fire `Notify` unconditionally. Channel routing is the notification platform's concern, not the module's.

## Preference Schema

```sql
-- notification_preferences table (managed by notifications module)
CREATE TABLE notification_preferences (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   uuid NOT NULL REFERENCES tenants(id),
    user_id     uuid NOT NULL REFERENCES users(id),
    category    text NOT NULL,   -- e.g. "contracts.submitted"
    channel     text NOT NULL,   -- "in_app" | "email" | "both" | "none"
    updated_at  timestamptz NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, user_id, category)
);
ALTER TABLE notification_preferences ENABLE ROW LEVEL SECURITY;
CREATE POLICY rls_notification_preferences ON notification_preferences
    USING (tenant_id = current_setting('app.tenant_id')::uuid);
```

## Default Preferences

If no preference row exists, the notification service uses tenant-level defaults from settings:

```
settings key: notifications.default_channel
values: "in_app" | "email" | "both" | "none"
default: "in_app"
```

## Module Responsibilities

The contracts module only:

1. Defines category constants in `domain/notification_categories.go`
2. Builds `port.Notification` structs with correct category/audience/priority
3. Calls `notifSvc.Notify` (or `NotifyTenant`) asynchronously after successful transitions

The module does **not**:

- Check user preferences
- Route to email vs in-app
- Throttle notification frequency
- Track delivery status
