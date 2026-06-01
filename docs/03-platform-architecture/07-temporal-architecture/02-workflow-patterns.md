---
title: Workflow Patterns
portal: 3 — Platform Architecture
section: 03-platform-architecture
audience: [architect, backend-engineer]
related:
  - "[Temporal Overview](01-temporal-overview.md)"
  - "[Workflow Design MDG](../../04-backend-engineering/00-module-development-guide/11-temporal-workflows/02-workflow-design.md)"
  - "[Activity Design MDG](../../04-backend-engineering/00-module-development-guide/11-temporal-workflows/03-activity-design.md)"
---

# Workflow Patterns

## Pattern 1: Approval Gate (Signal/Select)

Used when a workflow must pause and wait for a human decision.

```
Start workflow → Send notifications → Wait for signal (approve/reject)
  ├── "approved" signal → activate contract → notify
  └── "rejected" signal → archive contract → notify
  └── 72h timer fires → escalate → wait again
```

Key primitives:
- `workflow.NewChannel` for signal channel
- `workflow.NewTimer` for deadline
- `workflow.NewSelector` to wait on whichever fires first

Pattern code lives in MDG §11-02. All platform approval workflows follow this same structure.

## Pattern 2: Scheduled Check (Cron)

Used for periodic batch operations — expiry checks, reconciliation, cleanup.

```
Cron trigger (daily 06:00 UTC) → query DB for records → for each: activity
```

Guidelines:
- Cron workflows are stateless between runs — don't accumulate state across iterations
- Use `workflow.GetInfo(ctx).CronSchedule` to log the schedule from within the workflow
- Idempotency key = `{date}_{tenantID}_{operation}` to prevent duplicate processing if restarted mid-run

## Pattern 3: Saga (Compensating Transactions)

Used when a multi-step operation must be rolled back on failure.

```
Step 1: reserve inventory     → compensate: release reservation
Step 2: create invoice        → compensate: void invoice
Step 3: send confirmation     → compensate: send cancellation
```

Implementation with deferred compensation:

```go
func OrderWorkflow(ctx workflow.Context, req OrderRequest) error {
    var compensations []func(workflow.Context) error

    defer func() {
        if err != nil {
            for i := len(compensations) - 1; i >= 0; i-- {
                compensations[i](ctx)
            }
        }
    }()

    if err = workflow.ExecuteActivity(ctx, activities.ReserveInventory, req); err != nil {
        return err
    }
    compensations = append(compensations, func(ctx workflow.Context) error {
        return workflow.ExecuteActivity(ctx, activities.ReleaseReservation, req).Get(ctx, nil)
    })
    // ... next steps
}
```

## Pattern 4: Child Workflows

Used to break large workflows into manageable, independently retryable units:

```go
cwo := workflow.ChildWorkflowOptions{
    WorkflowID:         fmt.Sprintf("contract-line-import-%s", req.ImportID),
    ParentClosePolicy:  enums.PARENT_CLOSE_POLICY_ABANDON,
}
childCtx := workflow.WithChildOptions(ctx, cwo)

var result LineImportResult
err := workflow.ExecuteChildWorkflow(childCtx, ContractLineImportWorkflow, req).Get(ctx, &result)
```

`PARENT_CLOSE_POLICY_ABANDON` — child continues even if parent is terminated. Use when child represents independent work that must complete regardless.

## Pattern 5: Long-Running Activity with Heartbeat

Used for activities processing large datasets that could exceed Temporal's activity timeout:

```go
func (a *ContractActivities) ProcessBulkExport(ctx context.Context, req ExportRequest) error {
    for i, id := range req.ContractIDs {
        if err := activity.RecordHeartbeat(ctx, i); err != nil {
            return err  // context cancelled — stop cleanly
        }
        // process record...
    }
    return nil
}
```

Set `HeartbeatTimeout` in `ActivityOptions`. If the worker dies, Temporal reschedules on another worker starting from the last heartbeat checkpoint.

## Idempotency Guidelines

All activities must be safe to re-execute:

| Operation | Idempotency strategy |
|-----------|---------------------|
| DB insert | `ON CONFLICT DO NOTHING` or check-then-insert |
| DB update | `WHERE version = $v` optimistic lock |
| HTTP call to external system | Idempotency key in request header |
| Send email/notification | Check `sent_at IS NOT NULL` before sending |
| Publish event | Outbox dedup by event ID |

## Workflow ID Conventions

Workflow IDs must be deterministic and unique per logical operation:

```
contract-approval-{contract_id}         # one active per contract
contract-expiry-check                   # singleton cron
bulk-import-{import_id}                 # one per import job
finance-reconciliation-{tenant}-{date}  # one per tenant per day
```

Use `StartWorkflowOptions.WorkflowIDReusePolicy = REJECT_DUPLICATE` to prevent concurrent runs of the same workflow ID.

## Error Classification

| Error type | Retry? | Use when |
|------------|--------|----------|
| `temporal.ApplicationError` (retryable) | Yes | Transient failure (DB down, network) |
| `temporal.ApplicationError` (non-retryable) | No | Business rule violation (contract already terminated) |
| `context.DeadlineExceeded` | Yes | Timeout — Temporal retries |
| `temporal.CanceledError` | No | Workflow was explicitly cancelled |

```go
// Non-retryable business error
return temporal.NewNonRetryableApplicationError(
    "contract already terminated",
    "ErrContractAlreadyTerminated",
    nil,
)
```
