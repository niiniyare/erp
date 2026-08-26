---
title: Worker Registration
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[Temporal Overview](01-temporal-overview.md)"
  - "[Wire Registration](../09-wire-registration/01-wire-registration-overview.md)"
---

# Worker Registration

## Worker Setup

```go
// internal/core/contracts/worker/worker.go
package worker

import (
    "go.temporal.io/sdk/client"
    "go.temporal.io/sdk/worker"

    "awo.so/internal/core/contracts/activities"
    "awo.so/internal/core/contracts/workflows"
)

// NewContractsWorker creates and returns the contracts Temporal worker.
// The caller (server startup) is responsible for starting and stopping it.
func NewContractsWorker(
    temporalClient client.Client,
    contractActivities *activities.ContractActivities,
) worker.Worker {
    w := worker.New(temporalClient, workflows.ContractsTaskQueue, worker.Options{
        MaxConcurrentActivityExecutionSize:      10,
        MaxConcurrentWorkflowTaskExecutionSize:  10,
    })

    // Register workflows
    w.RegisterWorkflow(workflows.ContractApprovalWorkflow)
    w.RegisterWorkflow(workflows.ContractExpiryWorkflow)
    w.RegisterWorkflow(workflows.ContractBulkImportWorkflow)

    // Register activities (all methods on ContractActivities)
    w.RegisterActivity(contractActivities)

    return w
}
```

## Wire Provider

```go
// internal/core/contracts/contracts.go (additions)

var WorkerSet = wire.NewSet(
    activities.NewContractActivities,
    worker.NewContractsWorker,
)
```

## Server Startup Integration

```go
// cmd/server/main.go
func main() {
    app := wire.InitializeApp(...)

    // Start Temporal workers before accepting HTTP traffic
    for _, w := range app.TemporalWorkers {
        go func(w temporal.Worker) {
            if err := w.Start(); err != nil {
                log.Fatal().Err(err).Msg("temporal worker failed to start")
            }
        }(w)
    }

    // Fiber server
    if err := app.Fiber.Listen(":8080"); err != nil {
        log.Fatal().Err(err).Msg("fiber server failed")
    }
}
```

## Cron Schedules

Temporal schedules replace cron daemons. Register them once at startup:

```go
// internal/temporal/schedules.go
func RegisterSchedules(ctx context.Context, c client.Client, tenantIDs []string) error {
    for _, tenantID := range tenantIDs {
        scheduleID := "contract-expiry-" + tenantID
        _, err := c.ScheduleClient().Create(ctx, client.ScheduleOptions{
            ID: scheduleID,
            Spec: client.ScheduleSpec{
                CronExpressions: []string{"0 2 * * *"}, // 2 AM daily
            },
            Action: &client.ScheduleWorkflowAction{
                Workflow:  workflows.ContractExpiryWorkflow,
                TaskQueue: workflows.ContractsTaskQueue,
                Args: []any{workflows.ExpiryWorkflowInput{
                    TenantID:  tenantID,
                    DaysAhead: 30,
                }},
            },
        })
        if client.IsAlreadyStartedError(err) {
            continue // schedule already registered — skip
        }
        if err != nil {
            return fmt.Errorf("register expiry schedule for tenant %s: %w", tenantID, err)
        }
    }
    return nil
}
```
