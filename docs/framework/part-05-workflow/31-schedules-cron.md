---
title: "Chapter 31: Schedules and Cron"
part: "Part V — The Workflow Engine"
chapter: 31
section: "31-schedules-cron"
related:
  - "[Chapter 26: Temporal Fundamentals](26-temporal-fundamentals.md)"
  - "[Chapter 27: Defining Workflows](27-defining-workflows.md)"
---

# Chapter 31: Schedules and Cron

Temporal Schedules replace traditional cron jobs for Awo's periodic workflows. They provide the durability, catch-up execution, and observability that cron cannot offer.

---

## 31.1. Temporal Schedules vs Traditional Cron

### 31.1.1. Why Temporal Schedules

| Feature | Traditional cron | Temporal Schedule |
|---|---|---|
| Survives server restart | No — missed runs are lost | Yes — runs execute on recovery |
| Catch-up missed runs | No | Yes — configurable |
| Execution history | Only via log files | Full Temporal history |
| Overlap prevention | Manual with lock files | Built-in overlap policies |
| Pause/resume | Kill crontab entry | API call |
| Backfill | Manual rerun scripts | Built-in backfill command |
| Observability | Log scraping | Temporal Web UI |

### 31.1.2. What Temporal Schedules Cannot Do

- Sub-second scheduling (minimum interval: 1 second, practical minimum: 1 minute)
- Complex timezone-aware schedules with DST handling (use UTC cron expressions)
- Schedules that react to external events (use signals instead)

### 31.1.3. Schedule vs Cron Expression

```go
// Calendar-based (cron-like): nightly at 1 AM EAT = 10 PM UTC
Spec: &schedpb.ScheduleSpec{
    CronExpressions: []string{"0 22 * * *"},  // UTC
}

// Interval-based: every 4 hours
Spec: &schedpb.ScheduleSpec{
    Intervals: []*schedpb.IntervalSpec{
        {Interval: durationpb.New(4 * time.Hour)},
    },
}
```

Always use UTC in cron expressions. Never use local time — DST transitions cause double runs or missed runs.

---

## 31.2. Defining a Schedule

### 31.2.1. Schedule ID — Naming Convention

```
{module}.{workflow-name}.{frequency}

finance.gl-period-check.nightly
inventory.wetstock-reconciliation.daily
reporting.kpi-snapshot.weekly
billing.usage-report.monthly
```

### 31.2.2. Schedule Spec

```go
import (
    schedpb "go.temporal.io/api/schedule/v1"
    "google.golang.org/protobuf/types/known/durationpb"
)

spec := &schedpb.ScheduleSpec{
    CronExpressions: []string{"0 22 * * *"},  // 1 AM EAT daily
    Jitter: durationpb.New(5 * time.Minute),  // random jitter: don't all start at exactly 22:00
}
```

Jitter prevents thundering herd when many scheduled workflows start at the same instant.

### 31.2.3. Schedule Action

```go
action := &schedpb.ScheduleAction{
    Action: &schedpb.ScheduleAction_StartWorkflow{
        StartWorkflow: &workflowpb.NewWorkflowExecutionInfo{
            WorkflowType: &commonpb.WorkflowType{Name: "GLPeriodCheckWorkflow"},
            TaskQueue:    &taskqueuepb.TaskQueue{Name: "awo-finance"},
            Input:        payloads.EncodeProto(&GLPeriodCheckInput{
                CheckDate: time.Now().UTC(),
            }),
            WorkflowExecutionTimeout: durationpb.New(2 * time.Hour),
        },
    },
}
```

### 31.2.4. Overlap Policy

```go
OverlapPolicy: enumspb.SCHEDULE_OVERLAP_POLICY_SKIP
// SKIP: if previous run is still running, skip this run (default for most ERP jobs)

OverlapPolicy: enumspb.SCHEDULE_OVERLAP_POLICY_BUFFER_ONE
// BUFFER_ONE: buffer one run; if previous still running, queue the next one

OverlapPolicy: enumspb.SCHEDULE_OVERLAP_POLICY_TERMINATE_OTHER
// TERMINATE_OTHER: cancel the previous run and start the new one (use rarely)
```

For reconciliation workflows: use `SKIP`. A wetstock reconciliation that takes 3 hours should not be retried while still running — wait for next scheduled run.

---

## 31.3. Built-in Scheduled Workflows

### 31.3.1. Nightly GL Period Check

Runs at 1 AM EAT (22:00 UTC). Checks for:
- Unposted journal entries from the previous day
- GL balance discrepancies
- Entries posted to the wrong period

```go
func GLPeriodCheckWorkflow(ctx workflow.Context, input GLPeriodCheckInput) error {
    ao := workflow.ActivityOptions{StartToCloseTimeout: time.Hour}
    ctx = workflow.WithActivityOptions(ctx, ao)

    // Run check for each active tenant
    var tenants []TenantSummary
    workflow.ExecuteActivity(ctx, activities.ListActiveTenants).Get(ctx, &tenants)

    for _, tenant := range tenants {
        childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
            WorkflowID: fmt.Sprintf("finance.gl-check.%s.%s",
                tenant.ID, input.CheckDate.Format("2006-01-02")),
        })
        workflow.ExecuteChildWorkflow(childCtx, TenantGLPeriodCheckWorkflow,
            TenantGLCheckInput{TenantID: tenant.ID, CheckDate: input.CheckDate})
    }
    return nil
}
```

### 31.3.2. Daily Wetstock Reconciliation

Runs at 6 AM EAT (03:00 UTC) after overnight dip readings are captured.

```go
func DailyWetstockReconciliationWorkflow(ctx workflow.Context, input WetstockInput) error {
    // Per-site reconciliation
    var sites []Site
    workflow.ExecuteActivity(ctx, activities.ListActiveSites, input.TenantID).Get(ctx, &sites)

    for _, site := range sites {
        // Fire and don't wait — run all sites in parallel
        workflow.ExecuteChildWorkflow(ctx,
            workflow.ChildWorkflowOptions{WorkflowID: fmt.Sprintf("wetstock.%s.%s",
                site.ID, input.Date)},
            SiteWetstockReconciliationWorkflow,
            SiteWetstockInput{SiteID: site.ID, Date: input.Date})
    }
    // Could wait for all using workflow.NewSelector or Future collection
    return nil
}
```

### 31.3.3. Weekly KPI Snapshot

Runs every Monday at 08:00 EAT (05:00 UTC). Computes KPIs for the previous week and stores them for the dashboard:

```go
// Runs: 0 5 * * 1  (Monday 05:00 UTC = 08:00 EAT)
func WeeklyKPISnapshotWorkflow(ctx workflow.Context, input KPISnapshotInput) error {
    weekEnd := input.SnapshotDate
    weekStart := weekEnd.AddDate(0, 0, -7)

    ao := workflow.ActivityOptions{StartToCloseTimeout: 30 * time.Minute}
    ctx = workflow.WithActivityOptions(ctx, ao)

    return workflow.ExecuteActivity(ctx, activities.ComputeAndStoreKPISnapshot,
        KPIComputeInput{
            TenantID:  input.TenantID,
            WeekStart: weekStart,
            WeekEnd:   weekEnd,
        }).Get(ctx, nil)
}
```

---

## 31.4. Managing Schedules

### 31.4.1. Creating Schedules at Service Start — Idempotent Upsert

Schedules are created (or confirmed to exist) at worker startup:

```go
func EnsureSchedules(ctx context.Context, tc client.Client) error {
    schedules := []ScheduleConfig{
        {
            ID:   "finance.gl-period-check.nightly",
            Cron: "0 22 * * *",
            Workflow: "GLPeriodCheckWorkflow",
            Queue: "awo-finance",
        },
        // ...
    }

    for _, sched := range schedules {
        _, err := tc.ScheduleClient().Create(ctx, client.ScheduleOptions{
            ID:     sched.ID,
            Spec:   cronSpec(sched.Cron),
            Action: workflowAction(sched),
        })
        if err != nil {
            // Already exists — OK
            if !temporal.IsAlreadyExistsError(err) {
                return fmt.Errorf("create schedule %s: %w", sched.ID, err)
            }
        }
    }
    return nil
}
```

### 31.4.2. Pausing and Resuming Schedules

```go
handle := tc.ScheduleClient().GetHandle(ctx, "finance.gl-period-check.nightly")

// Pause (with reason for audit)
handle.Pause(ctx, client.SchedulePauseOptions{
    Note: "paused for system maintenance 2025-07-04 by ops team",
})

// Resume
handle.Unpause(ctx, client.ScheduleUnpauseOptions{
    Note: "maintenance complete",
})
```

### 31.4.3. Triggering a Schedule Manually

```go
// Trigger one immediate run (ignores overlap policy)
handle.Trigger(ctx, client.ScheduleTriggerOptions{
    Overlap: enumspb.SCHEDULE_OVERLAP_POLICY_ALLOW_ALL,
})

// Backfill: run as if schedule fired at these times
handle.Backfill(ctx, client.ScheduleBackfillOptions{
    Backfill: []client.ScheduleBackfill{
        {
            Start:   time.Date(2025, 7, 1, 22, 0, 0, 0, time.UTC),
            End:     time.Date(2025, 7, 4, 22, 0, 0, 0, time.UTC),
            Overlap: enumspb.SCHEDULE_OVERLAP_POLICY_ALLOW_ALL,
        },
    },
})
```

Use backfill when a schedule was paused or a worker was down and you need to process the missed periods.

### 31.4.4. Deleting a Schedule

```go
handle.Delete(ctx)
```

Deleting a schedule does not affect currently running workflow instances started by that schedule. Running instances continue to completion.
