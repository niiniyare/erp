> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Notification Types
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[Notifications Overview](01-notifications-overview.md)"
  - "[Sending Notifications](03-sending-notifications.md)"
  - "[Notification Channels](06-notification-channels.md)"
---

# Notification Types

Each contract lifecycle transition maps to a category string and a target audience.

## Category Naming

Format: `{module}.{event}` — lowercase, dot-separated.

```
contracts.submitted
contracts.under_review
contracts.approved
contracts.rejected
contracts.activated
contracts.suspended
contracts.terminated
contracts.expiring_soon
contracts.expired
```

## Category Registry

Define category constants in the domain package so handlers and services share them:

```go
// internal/core/contracts/domain/notification_categories.go
package domain

const (
    NotifContractSubmitted    = "contracts.submitted"
    NotifContractUnderReview  = "contracts.under_review"
    NotifContractApproved     = "contracts.approved"
    NotifContractRejected     = "contracts.rejected"
    NotifContractActivated    = "contracts.activated"
    NotifContractSuspended    = "contracts.suspended"
    NotifContractTerminated   = "contracts.terminated"
    NotifContractExpiringSoon = "contracts.expiring_soon"
)
```

## Audience Matrix

| Category | Audience | Role / User |
|----------|----------|-------------|
| `contracts.submitted` | Role | `contracts.reviewer` |
| `contracts.under_review` | User | Submitter (createdBy) |
| `contracts.approved` | User | Submitter (createdBy) |
| `contracts.rejected` | User | Submitter (createdBy) |
| `contracts.activated` | User | Contract owner (assignedTo) |
| `contracts.suspended` | Role | `contracts.admin` |
| `contracts.terminated` | Role | `contracts.admin` |
| `contracts.expiring_soon` | User | Contract owner (assignedTo) |

## Priority Mapping

| Category | Priority |
|----------|----------|
| `contracts.submitted` | `normal` |
| `contracts.approved` | `normal` |
| `contracts.rejected` | `high` |
| `contracts.activated` | `normal` |
| `contracts.suspended` | `high` |
| `contracts.terminated` | `urgent` |
| `contracts.expiring_soon` | `high` |

## Building a Notification

```go
// internal/core/contracts/service/notifications.go
package service

import (
    "fmt"

    "github.com/google/uuid"

    "awo.so/internal/core/contracts/domain"
    "awo.so/internal/core/notifications/port"
)

func buildSubmittedNotification(tenantID, contractID uuid.UUID, contractNumber string) port.Notification {
    return port.Notification{
        TenantID:     tenantID,
        Audience:     port.AudienceRole,
        Role:         "contracts.reviewer",
        Category:     domain.NotifContractSubmitted,
        Title:        "Contract Submitted for Review",
        Body:         fmt.Sprintf("Contract %s has been submitted and awaits review.", contractNumber),
        ResourceID:   contractID,
        ResourceType: "contract",
        Priority:     port.PriorityNormal,
    }
}

func buildApprovedNotification(tenantID, contractID uuid.UUID, contractNumber string, submitterID uuid.UUID) port.Notification {
    return port.Notification{
        TenantID:    tenantID,
        Audience:    port.AudienceUser,
        RecipientID: submitterID,
        Category:    domain.NotifContractApproved,
        Title:       "Contract Approved",
        Body:        fmt.Sprintf("Contract %s has been approved.", contractNumber),
        ResourceID:  contractID,
        ResourceType: "contract",
        Priority:    port.PriorityNormal,
    }
}

func buildRejectedNotification(tenantID, contractID uuid.UUID, contractNumber string, submitterID uuid.UUID, reason string) port.Notification {
    body := fmt.Sprintf("Contract %s has been returned to draft.", contractNumber)
    if reason != "" {
        body += " Reason: " + reason
    }
    return port.Notification{
        TenantID:    tenantID,
        Audience:    port.AudienceUser,
        RecipientID: submitterID,
        Category:    domain.NotifContractRejected,
        Title:       "Contract Returned to Draft",
        Body:        body,
        ResourceID:  contractID,
        ResourceType: "contract",
        Priority:    port.PriorityHigh,
        Metadata: map[string]string{
            "reason": reason,
        },
    }
}
```

## Expiry Notifications

Expiry notifications are triggered by a scheduled job, not by a user action:

```go
// internal/core/contracts/scheduler/expiry_check.go
// Called by a cron/Temporal workflow — not directly from service layer
func (j *expiryCheckJob) run(ctx context.Context) error {
    // Find contracts expiring within 30 days
    contracts, err := j.repo.ListContractsExpiringSoon(ctx, repository.ExpiryCheckParams{
        DaysAhead: 30,
    })
    if err != nil {
        return err
    }
    for _, c := range contracts {
        j.notifSvc.Notify(ctx, buildExpiringSoonNotification(c))
    }
    return nil
}
```
