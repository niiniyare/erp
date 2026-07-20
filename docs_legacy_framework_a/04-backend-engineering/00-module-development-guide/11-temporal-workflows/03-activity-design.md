> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Activity Design
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[Workflow Design](02-workflow-design.md)"
  - "[Temporal Overview](01-temporal-overview.md)"
---

# Activity Design

Activities contain the actual side-effects: database writes, HTTP calls, email sends. They must be idempotent — Temporal may retry them multiple times.

## Activity Struct

Activities are methods on a struct that holds the service dependencies:

```go
// internal/core/contracts/activities/contract_activities.go
package activities

import (
    "context"

    "go.temporal.io/sdk/activity"

    "awo.so/internal/core/contracts/repository"
    "awo.so/internal/core/contracts/domain"
    "awo.so/internal/core/notifications/port"
    "github.com/google/uuid"
)

// ContractActivities holds dependencies for all contract activities.
type ContractActivities struct {
    repo     repository.ContractRepository
    notifSvc port.NotificationService
}

func NewContractActivities(
    repo repository.ContractRepository,
    notifSvc port.NotificationService,
) *ContractActivities {
    return &ContractActivities{repo: repo, notifSvc: notifSvc}
}
```

## Activity: NotifyReviewers

```go
type NotifyReviewersInput struct {
    TenantID       string
    ContractID     string
    ContractNumber string
}

func (a *ContractActivities) NotifyReviewers(ctx context.Context, input NotifyReviewersInput) error {
    logger := activity.GetLogger(ctx)
    logger.Info("NotifyReviewers activity", "contractID", input.ContractID)

    tenantID := uuid.MustParse(input.TenantID)
    contractID := uuid.MustParse(input.ContractID)

    return a.notifSvc.NotifyTenant(ctx, tenantID, "contracts.reviewer",
        port.Notification{
            TenantID:     tenantID,
            Audience:     port.AudienceRole,
            Role:         "contracts.reviewer",
            Category:     domain.NotifContractSubmitted,
            Title:        "Contract Submitted for Review",
            Body:         "Contract " + input.ContractNumber + " requires review.",
            ResourceID:   contractID,
            ResourceType: "contract",
            Priority:     port.PriorityNormal,
        },
    )
}
```

## Activity: ApproveContract

```go
type ApproveContractInput struct {
    TenantID   string
    ContractID string
    ReviewerID string
    Comment    string
}

func (a *ContractActivities) ApproveContract(ctx context.Context, input ApproveContractInput) error {
    logger := activity.GetLogger(ctx)
    logger.Info("ApproveContract activity", "contractID", input.ContractID)

    tenantID   := uuid.MustParse(input.TenantID)
    contractID := uuid.MustParse(input.ContractID)
    reviewerID := uuid.MustParse(input.ReviewerID)

    // Idempotent: if already approved, repo returns the current contract without error
    _, err := a.repo.UpdateStatus(ctx, repository.UpdateContractStatusParams{
        ID:        contractID,
        TenantID:  tenantID,
        Status:    domain.ContractStatusApproved,
        UpdatedBy: reviewerID,
        // Note: no Version check here — workflow owns the version
    })
    return err
}
```

## Activity: FindExpiringContracts

```go
type FindExpiringContractsInput struct {
    TenantID  string
    DaysAhead int
}

func (a *ContractActivities) FindExpiringContracts(ctx context.Context, input FindExpiringContractsInput) ([]string, error) {
    tenantID := uuid.MustParse(input.TenantID)

    contracts, err := a.repo.ListContractsExpiringSoon(ctx, repository.ExpiryCheckParams{
        TenantID:  tenantID,
        DaysAhead: input.DaysAhead,
    })
    if err != nil {
        return nil, err
    }

    ids := make([]string, len(contracts))
    for i, c := range contracts {
        ids[i] = c.ID.String()
    }
    return ids, nil
}
```

## Idempotency Patterns

Activities may be called multiple times. Design for safe re-execution:

| Scenario | Pattern |
|----------|---------|
| INSERT into DB | Use `ON CONFLICT DO NOTHING` or check-then-insert |
| Status update | Guard with `WHERE status = 'expected_status'` |
| Email send | Deduplicate via idempotency key in outbox table |
| Notification | Notification service deduplicates by category+resource+window |

```go
// Safe idempotent status update in activity
_, err = a.repo.UpdateStatusIfExpected(ctx, repository.UpdateStatusIfExpectedParams{
    ID:             contractID,
    TenantID:       tenantID,
    ExpectedStatus: domain.ContractStatusSubmitted, // only update if still submitted
    NewStatus:      domain.ContractStatusApproved,
    UpdatedBy:      reviewerID,
})
if errors.Is(err, domain.ErrContractNotFound) {
    // Already transitioned — idempotent success
    return nil
}
return err
```

## Long-Running Activities: Heartbeating

Activities expected to run > 10 seconds must heartbeat:

```go
func (a *ContractActivities) BulkImportContracts(ctx context.Context, input BulkImportInput) error {
    for i, row := range input.Rows {
        // Heartbeat every 100 rows
        if i%100 == 0 {
            activity.RecordHeartbeat(ctx, i)
        }

        if err := a.importRow(ctx, row); err != nil {
            return fmt.Errorf("row %d: %w", i, err)
        }
    }
    return nil
}
```

The `StartToCloseTimeout` on bulk activities should be set generously (e.g., 30 minutes).
