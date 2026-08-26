---
title: Workflow Patterns Quick Reference
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer]
related:
  - "[Temporal Overview](01-temporal-overview.md)"
  - "[Workflow Design](02-workflow-design.md)"
  - "[Activity Design](03-activity-design.md)"
  - "[Worker Registration](04-worker-registration.md)"
  - "[Workflow Checklist](07-workflow-checklist.md)"
---

# Workflow Patterns Quick Reference

## Starting a Workflow

```go
// From a service method
options := client.StartWorkflowOptions{
    ID:        fmt.Sprintf("contract-approval-%s", contractID),
    TaskQueue: "awoerp.contracts",
    WorkflowExecutionTimeout: 72 * time.Hour,
    WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
}

we, err := s.temporalClient.ExecuteWorkflow(ctx, options, ContractApprovalWorkflow, WorkflowInput{
    ContractID: contractID,
    TenantID:   tenantID,
})
if temporal.IsAlreadyStartedError(err) {
    // workflow already running — OK
    return nil
}
if err != nil {
    return fmt.Errorf("start workflow: %w", err)
}
s.logger.InfoContext(ctx, "workflow started", "run_id", we.GetRunID())
```

## Signaling a Workflow

```go
err := s.temporalClient.SignalWorkflow(
    ctx,
    fmt.Sprintf("contract-approval-%s", contractID),
    "",   // latest run
    "approval-decision",
    ApprovalSignal{
        Approved:   true,
        Comment:    "Reviewed and approved",
        ApproverID: sess.UserID,
    },
)
```

## Workflow with Signal and Timer

```go
func ContractApprovalWorkflow(ctx workflow.Context, input WorkflowInput) error {
    approvalCh := workflow.GetSignalChannel(ctx, "approval-decision")
    deadlineTimer := workflow.NewTimer(ctx, 72*time.Hour)

    var signal ApprovalSignal
    s := workflow.NewSelector(ctx)
    s.AddReceive(approvalCh, func(ch workflow.ReceiveChannel, more bool) {
        ch.Receive(ctx, &signal)
    })
    s.AddFuture(deadlineTimer, func(f workflow.Future) {
        // timer fired — no action needed, signal will be nil
    })
    s.Select(ctx)

    if signal.ApproverID == uuid.Nil {
        // Timer fired without approval
        return executeActivity(ctx, activities.EscalateContract, input.ContractID)
    }
    if signal.Approved {
        return executeActivity(ctx, activities.ActivateContract, input.ContractID)
    }
    return executeActivity(ctx, activities.ReturnToDraft, input.ContractID)
}
```

## Activity with Retry Policy

```go
ao := workflow.ActivityOptions{
    StartToCloseTimeout: 30 * time.Second,
    RetryPolicy: &temporal.RetryPolicy{
        InitialInterval:    time.Second,
        BackoffCoefficient: 2.0,
        MaximumInterval:    30 * time.Second,
        MaximumAttempts:    5,
        NonRetryableErrorTypes: []string{
            "ErrContractNotFound",
            "ErrContractAlreadyTerminated",
        },
    },
}
ctx = workflow.WithActivityOptions(ctx, ao)
err := workflow.ExecuteActivity(ctx, activities.ActivateContract, contractID).Get(ctx, nil)
```

## Non-Retryable Error

```go
// In activity — business rule violation
return temporal.NewNonRetryableApplicationError(
    "contract is not in approvable state",
    "ErrContractNotEditable",
    nil,
)
```

## Heartbeat (Long Activity)

```go
func (a *ContractActivities) ProcessBulkExport(ctx context.Context, req ExportRequest) error {
    for i, id := range req.ContractIDs {
        if i%10 == 0 {
            if err := activity.RecordHeartbeat(ctx, i); err != nil {
                return err   // context cancelled — stop cleanly
            }
        }
        // process record
    }
    return nil
}
```

Set `HeartbeatTimeout: 30 * time.Second` in activity options.

## Cron Workflow

```go
we, err := s.temporalClient.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
    ID:           "contract-expiry-check",
    TaskQueue:    "awoerp.contracts",
    CronSchedule: "0 6 * * *",   // 06:00 UTC daily
}, ContractExpiryCheckWorkflow)
if temporal.IsAlreadyStartedError(err) {
    return nil   // already scheduled
}
```

## Query Workflow State

```go
// Register in workflow
workflow.SetQueryHandler(ctx, "status", func() (string, error) {
    return wf.CurrentStatus, nil
})

// Query from service
resp, err := s.temporalClient.QueryWorkflow(ctx,
    fmt.Sprintf("contract-approval-%s", contractID),
    "",
    "status",
)
var status string
resp.Get(&status)
```

## Determinism Rules

| In workflow functions | Use instead |
|----------------------|-------------|
| `time.Now()` | `workflow.Now(ctx)` |
| `time.Sleep()` | `workflow.Sleep(ctx, duration)` |
| `rand.*` | `workflow.SideEffect()` |
| goroutines | `workflow.Go()` |
| DB/HTTP calls | Activities |
| Global mutable state | Struct fields on workflow struct |

## Helper: executeActivity

```go
// Reduces boilerplate for simple activity calls
func executeActivity(ctx workflow.Context, fn any, args ...any) error {
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 30 * time.Second,
    }
    return workflow.ExecuteActivity(
        workflow.WithActivityOptions(ctx, ao),
        fn,
        args...,
    ).Get(ctx, nil)
}
```
