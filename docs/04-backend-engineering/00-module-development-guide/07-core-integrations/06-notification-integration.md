---
title: Notification Integration
portal: 4 — Backend Engineering
section: 00-module-development-guide/07-core-integrations
audience: [backend-engineer, tech-lead]
related:
  - path: ./05-audit-integration.md
    title: Audit Integration
---

# Notification Integration

Notifications alert users to events that require their attention. They are always fire-and-forget — notification failure never fails the originating operation.

## When to Send Notifications

Send notifications for events that require user action or awareness:

| Event | Who to notify |
|-------|--------------|
| Contract submitted | Users with `contract_reviewer` role in the tenant |
| Contract approved | Contract creator |
| Contract activated | Contract creator, vendor contact |
| Contract expiring in 30 days | Contract owner, team managers |
| Contract suspended | Contract creator, vendor contact |

Do **not** send notifications for:
- Administrative creates/updates in draft status
- System-initiated operations (e.g., auto-activation)

## Notification Message Structure

```go
type Message struct {
	TenantID    uuid.UUID
	Type        string              // "contract.submitted", "contract.approved", etc.
	Title       string
	Body        string
	Audience    Audience            // who receives this notification
	ResourceID  string              // contract UUID — for deep linking
	Priority    Priority            // "low", "normal", "high"
	Metadata    map[string]string   // extra context for the notification renderer
}
```

## Audience Types

```go
// Notify a specific user
notifications.AudienceByUser(userID)

// Notify all users with a specific role in the tenant
notifications.AudienceByRole("contract_reviewer")

// Notify multiple specific users
notifications.AudienceByUsers([]uuid.UUID{...})
```

## Sending Asynchronously

```go
// Never block the response on notification delivery
go func() {
	notifCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.notifSvc.Send(notifCtx, notifications.Message{
		TenantID:   contract.TenantID,
		Type:       "contract.submitted",
		Title:      "Contract submitted for review",
		Body:       fmt.Sprintf("Contract %s (%s) has been submitted and requires review.", contract.ContractNumber, contract.Title),
		Audience:   notifications.AudienceByRole("contract_reviewer"),
		ResourceID: contract.ID.String(),
		Priority:   notifications.PriorityNormal,
	}); err != nil {
		s.logger.Warn().Err(err).Str("type", "contract.submitted").Msg("notification send failed")
	}
}()
```

Use `context.Background()` — not the request context. Log failures at WARN level, not ERROR — notification delivery failure is not a critical system error.

## Notification Preferences

The notification service checks user preferences before delivering. The service will not send a notification to a user who has opted out of that type. The business module does not need to check preferences manually.

## Helper in Service

The service wraps the async notification in a helper to reduce boilerplate:

```go
func (s *contractService) notifyAsync(ctx context.Context, msg notifications.Message) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error().Interface("panic", r).Msg("notification goroutine panic")
			}
		}()
		notifCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.notifSvc.Send(notifCtx, msg); err != nil {
			s.logger.Warn().Err(err).Str("type", msg.Type).Msg("notification send failed")
		}
	}()
}
```

The `recover()` prevents a panicking notification service from crashing the server process.
