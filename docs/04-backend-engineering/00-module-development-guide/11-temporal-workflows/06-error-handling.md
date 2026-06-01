---
title: Workflow Error Handling
portal: 4 — Backend Engineering
section: 00-module-development-guide/11-temporal-workflows
audience: [backend-engineer, tech-lead]
related:
  - path: ./02-workflow-design.md
    title: Workflow Design
  - path: ./03-activity-design.md
    title: Activity Design
---

# Workflow Error Handling

## Error Categories

| Error Type | Temporal Behavior | How to Handle |
|-----------|-------------------|---------------|
| Transient (network timeout, DB unavailable) | Retry via RetryPolicy | Let policy retry |
| Business logic error (contract not found) | Non-retryable | Wrap in `temporal.NewNonRetryableApplicationError` |
| Activity panic | Marks workflow as failed | Fix code; use `recover()` defensively |
| Workflow logic error | Marks workflow as failed | Fix code |

## Non-Retryable Business Errors

```go
// In activity
func (a *ContractActivities) ApproveContract(ctx context.Context, input ApproveContractInput) error {
    _, err := a.repo.UpdateStatus(ctx, ...)
    if err != nil {
        if errors.Is(err, domain.ErrContractNotFound) {
            // Contract deleted mid-workflow — no point retrying
            return temporal.NewNonRetryableApplicationError(
                "contract not found",
                "ContractNotFound",
                err,
            )
        }
        // Other errors — let Temporal retry
        return err
    }
    return nil
}
```

## Compensating Transactions (Saga)

When a multi-step workflow fails mid-way, run compensation activities to undo completed steps:

```go
func ContractBulkImportWorkflow(ctx workflow.Context, input BulkImportInput) (err error) {
    var importedIDs []string

    // Deferred compensation
    defer func() {
        if err != nil && len(importedIDs) > 0 {
            compensateCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
                StartToCloseTimeout: 5 * time.Minute,
            })
            cErr := workflow.ExecuteActivity(compensateCtx,
                activities.RollbackImportedContracts,
                activities.RollbackInput{IDs: importedIDs},
            ).Get(ctx, nil)
            if cErr != nil {
                workflow.GetLogger(ctx).Error("rollback failed", "error", cErr)
            }
        }
    }()

    // ... import loop; append to importedIDs on each success ...
    return nil
}
```

## Retry Policy per Activity

Different activities warrant different retry configurations:

```go
// Fast, idempotent DB write — retry quickly
dbActivityOptions := workflow.ActivityOptions{
    StartToCloseTimeout: 30 * time.Second,
    RetryPolicy: &temporal.RetryPolicy{
        InitialInterval:    500 * time.Millisecond,
        MaximumAttempts:    5,
        BackoffCoefficient: 2.0,
    },
}

// External email service — retry slowly
emailActivityOptions := workflow.ActivityOptions{
    StartToCloseTimeout: 60 * time.Second,
    RetryPolicy: &temporal.RetryPolicy{
        InitialInterval:    5 * time.Second,
        MaximumAttempts:    3,
        BackoffCoefficient: 2.0,
        MaximumInterval:    30 * time.Second,
    },
}
```

## Workflow Cancellation

Handle cancellation gracefully to avoid leaving state inconsistent:

```go
func ContractApprovalWorkflow(ctx workflow.Context, input ApprovalWorkflowInput) error {
    // ...
    selector := workflow.NewSelector(ctx)

    cancelCh := workflow.GetSignalChannel(ctx, "cancel")
    selector.AddReceive(cancelCh, func(ch workflow.ReceiveChannel, more bool) {
        ch.Receive(ctx, nil)
        // Clean up — e.g., revert contract to draft
        _ = workflow.ExecuteActivity(ctx, activities.RevertToDraft, input.ContractID)
    })

    // ... other selector cases ...
    selector.Select(ctx)
    return nil
}
```
