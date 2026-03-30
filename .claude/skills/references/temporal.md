# Temporal Workflows Reference

## Current status

The Temporal SDK (`go.temporal.io/sdk` v1.41.1) is fully wired into the infrastructure layer
(`internal/platform/temporal/`) but **no concrete workflows or activities have been implemented yet**.

Config is loaded from `TEMPORAL_*` env vars via `TemporalConfig` in `config.go`.

## When to use Temporal

Use Temporal for any operation that:
- Has steps that can fail independently (e.g. send email, then update DB, then notify Slack)
- Needs retry-with-backoff across service restarts
- Is long-running or asynchronous (onboarding, bulk jobs, scheduled reports)
- Must be idempotent and auditable

Do NOT use for simple in-process async — use goroutines. Temporal overhead is only worth it for
durable, distributed work.

## File layout convention (follow this when implementing)

```
internal/
  platform/temporal/          # Client + worker bootstrap (existing)
  core/<domain>/
    workflows/
      <noun>_workflow.go      # Workflow function + input/output types
      <noun>_activity.go      # Activity functions
      worker.go               # Worker registration for this domain
```

## Workflow skeleton

```go
// internal/core/<domain>/workflows/<noun>_workflow.go
package workflows

import (
    "go.temporal.io/sdk/workflow"
    "time"
)

type FooWorkflowInput struct {
    TenantID string
    FooID    string
}

type FooWorkflowResult struct {
    Success bool
    Message string
}

// FooWorkflow is the workflow entry point — must be deterministic.
// All non-deterministic work (DB, HTTP, time.Now) goes in activities.
func FooWorkflow(ctx workflow.Context, input FooWorkflowInput) (*FooWorkflowResult, error) {
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 30 * time.Second,
        RetryPolicy: &temporal.RetryPolicy{
            MaximumAttempts: 3,
        },
    }
    ctx = workflow.WithActivityOptions(ctx, ao)

    var result FooActivityResult
    err := workflow.ExecuteActivity(ctx, DoFooActivity, FooActivityInput{
        TenantID: input.TenantID,
        FooID:    input.FooID,
    }).Get(ctx, &result)
    if err != nil {
        return nil, err
    }

    return &FooWorkflowResult{Success: true, Message: result.Message}, nil
}
```

## Activity skeleton

```go
// internal/core/<domain>/workflows/<noun>_activity.go
package workflows

type FooActivityInput struct {
    TenantID string
    FooID    string
}

type FooActivityResult struct {
    Message string
}

// Activities receive injected dependencies via a struct receiver.
type FooActivities struct {
    fooService domain.FooService
    logger     logger.Logger
}

func NewFooActivities(svc domain.FooService, log logger.Logger) *FooActivities {
    return &FooActivities{fooService: svc, logger: log}
}

func (a *FooActivities) DoFooActivity(ctx context.Context, input FooActivityInput) (*FooActivityResult, error) {
    tenantID, err := uuid.Parse(input.TenantID)
    if err != nil {
        return nil, err
    }
    ctx = shared.WithTenantID(ctx, tenantID)  // re-inject tenant into activity ctx

    foo, err := a.fooService.Process(ctx, input.FooID)
    if err != nil {
        return nil, err
    }
    return &FooActivityResult{Message: foo.Status}, nil
}
```

## Worker registration

```go
// internal/core/<domain>/workflows/worker.go
package workflows

import (
    "go.temporal.io/sdk/client"
    "go.temporal.io/sdk/worker"
)

const FooTaskQueue = "foo-task-queue"

func RegisterFooWorker(c client.Client, activities *FooActivities) worker.Worker {
    w := worker.New(c, FooTaskQueue, worker.Options{})
    w.RegisterWorkflow(FooWorkflow)
    w.RegisterActivity(activities)
    return w
}
```

## Triggering a workflow from a service

```go
// in a service method
we, err := temporalClient.ExecuteWorkflow(ctx,
    client.StartWorkflowOptions{
        ID:        fmt.Sprintf("foo-%s", fooID),
        TaskQueue: workflows.FooTaskQueue,
    },
    workflows.FooWorkflow,
    workflows.FooWorkflowInput{
        TenantID: tenantID.String(),
        FooID:    fooID.String(),
    },
)
if err != nil {
    return err
}
// optionally wait: we.Get(ctx, &result)
```

## Important rules

1. **Workflow functions must be deterministic.** No `time.Now()`, `rand`, direct DB calls, or
   HTTP inside a workflow function. Everything non-deterministic goes in activities.
2. **Re-inject tenant context in every activity.** Activities receive a plain `context.Context`
   from Temporal — set `shared.WithTenantID(ctx, ...)` before calling any service/repository.
3. **Use typed input/output structs.** Never pass raw strings as workflow arguments — use
   structs so fields can evolve without breaking the serialization contract.
4. **Workflow IDs should be deterministic** when idempotency matters (e.g. `"password-reset-<userID>"`).
5. **Add workers to Wire.** Register `NewFooActivities` in the domain's provider set and
   include `RegisterFooWorker` in the application startup.
