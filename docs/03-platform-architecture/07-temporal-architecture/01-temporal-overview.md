---
title: Temporal Architecture Overview
portal: 3 — Platform Architecture
section: 07-temporal-architecture
audience: [architect, backend-engineer]
related:
  - "[Event Architecture](../04-event-architecture/01-event-architecture.md)"
  - "[Temporal Workflows MDG](../../04-backend-engineering/00-module-development-guide/11-temporal-workflows/01-temporal-overview.md)"
  - "[System Overview](../00-overview/01-system-overview.md)"
---

# Temporal Architecture Overview

## Role of Temporal

Temporal handles long-running, multi-step processes that must survive service restarts and require reliable retries.

| Use Temporal for | Use HTTP / goroutines for |
|-----------------|--------------------------|
| Approval workflows (wait for signal) | Fire-and-forget audit log writes |
| Contract expiry checks (cron) | Event publication after HTTP response |
| Multi-step financial reconciliation | Notification delivery |
| Bulk import with per-row retry | Single DB write with immediate result |
| Anything that must wait hours/days | Anything completing in < 30 seconds |

## Infrastructure

| Component | Description |
|-----------|-------------|
| Temporal Server | Workflow orchestration service (Docker Compose locally, self-hosted or Temporal Cloud in prod) |
| PostgreSQL (Temporal) | Temporal's own persistence store — separate DB from AwoERP's application DB |
| Temporal UI | Visibility into workflow history, signals, retries — port 8088 |
| AwoERP Workers | Go processes that poll Temporal for workflow/activity tasks |

## Worker Architecture

AwoERP registers one **task queue** per module:

```
awoerp.contracts    ← contracts workflows + activities
awoerp.finance      ← finance workflows + activities
awoerp.iam          ← IAM / session cleanup workflows
```

Workers are registered at application startup:

```
main() → InitializeApp() → Wire → NewContractsWorker, NewFinanceWorker, ...
                                → worker.Start()
```

Each worker polls its task queue. Temporal routes workflow tasks to workers registered on that queue.

## Determinism Constraint

Workflow functions **must be deterministic**. The same event history must always replay to the same state.

**Prohibited in workflow functions:**

- `time.Now()` → use `workflow.Now(ctx)`
- `rand.Int()` → use `workflow.SideEffect()`
- Goroutines → use `workflow.Go()`
- Network/DB calls → move to activities
- Global mutable state → never

**Side effects (I/O, randomness, time) always happen in activities.**

## Failure and Retry Model

```
Workflow (orchestrator)
  └── Activity (worker executes)
        ├── success → workflow continues
        ├── failure → Temporal retries with backoff (per RetryPolicy)
        └── non-retryable error → workflow receives ApplicationError, decides path
```

Default retry policy for activities:

| Setting | Default |
|---------|---------|
| Initial interval | 1s |
| Backoff multiplier | 2× |
| Max interval | 100s |
| Max attempts | unlimited |

Override per-activity for tight SLAs:

```go
ao := workflow.ActivityOptions{
    StartToCloseTimeout: 30 * time.Second,
    RetryPolicy: &temporal.RetryPolicy{
        MaximumAttempts: 3,
    },
}
ctx = workflow.WithActivityOptions(ctx, ao)
```

## Signal and Query Architecture

**Signals** allow external code (HTTP handlers) to push events into running workflows:

```
HTTP POST /contracts/:id/approve
  → contractService.Approve()
    → temporalClient.SignalWorkflow("ContractApproval-{id}", "approval-decision", payload)
      → workflow selector unblocks
        → next activity executes
```

**Queries** allow reading workflow state without mutation:

```go
// Register in workflow:
workflow.SetQueryHandler(ctx, "status", func() (string, error) {
    return wf.Status, nil
})

// Call from handler:
resp, _ := temporalClient.QueryWorkflow(ctx, workflowID, "", "status")
```

## Cron Workflows

Scheduled workflows use Temporal's built-in cron scheduler:

```go
client.ExecuteWorkflow(ctx, temporal.StartWorkflowOptions{
    ID:           "contract-expiry-checker",
    TaskQueue:    "awoerp.contracts",
    CronSchedule: "0 6 * * *",   // 06:00 UTC daily
}, ContractExpiryCheckWorkflow)
```

Guard against duplicate starts with `IsAlreadyStartedError`.

## Visibility and Observability

Temporal exports workflow history, search attributes, and custom metadata:

```go
// Set searchable attributes
workflow.UpsertSearchAttributes(ctx, map[string]any{
    "TenantID":       tenantID,
    "ContractStatus": "under_review",
})
```

Query via Temporal UI or `tctl wf list --query 'TenantID="..."'`.

Workflow spans link to application traces via `temporal.WithContext(ctx)` — propagates OTel trace context into activity execution.

## Local vs Production

| Aspect | Local (Docker Compose) | Production |
|--------|----------------------|------------|
| Server | `temporalio/auto-setup` image | Temporal Cloud or self-hosted cluster |
| Persistence | SQLite (auto-setup) | PostgreSQL |
| Workers | Same process as API server | Separate Deployment |
| UI | `localhost:8088` | Internal VPN access |
| Namespace | `default` | `awoerp-{env}` |
