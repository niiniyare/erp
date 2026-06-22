---
title: "Chapter 32: Workflow Observability"
part: "Part V — The Workflow Engine"
chapter: 32
section: "32-workflow-observability"
related:
  - "[Chapter 26: Temporal Fundamentals](26-temporal-fundamentals.md)"
  - "[Chapter 30: Signals, Queries](30-signals-queries.md)"
---

# Chapter 32: Workflow Observability

Temporal provides a rich execution history for every workflow, but understanding that history and correlating it with application traces and alerts requires deliberate integration. This chapter covers the Temporal Web UI, OpenTelemetry integration, and operational alerting.

---

## 32.1. Temporal Web UI

### 32.1.1. Reading Workflow Event History

The Temporal Web UI at `http://temporal:8080` shows the event history for every workflow. Key event types:

| Event | Meaning |
|---|---|
| `WorkflowExecutionStarted` | Workflow started; contains input params |
| `ActivityTaskScheduled` | Activity queued; contains activity type and input |
| `ActivityTaskStarted` | Worker picked up the activity |
| `ActivityTaskCompleted` | Activity succeeded; contains output |
| `ActivityTaskFailed` | Activity failed; contains error message and retry count |
| `TimerStarted` | `workflow.Sleep()` or `workflow.NewTimer()` called |
| `TimerFired` | Timer elapsed |
| `SignalExternalWorkflowExecutionInitiated` | Signal sent to another workflow |
| `WorkflowExecutionSignaled` | Signal received by this workflow |
| `MarkerRecorded` | `workflow.GetVersion()` or `workflow.SideEffect()` called |
| `WorkflowExecutionCompleted` | Workflow finished successfully |
| `WorkflowExecutionFailed` | Workflow failed |

When debugging a stuck workflow: look for `ActivityTaskScheduled` without a following `ActivityTaskStarted` — this means no worker is polling the task queue or all workers are at max concurrency.

### 32.1.2. Searching Workflows by Search Attribute

In the Temporal Web UI search bar:

```
WorkflowType = "InvoiceApprovalWorkflow"
  AND CustomKeyword = "acme"        (TenantSlug search attribute)
  AND ExecutionStatus = "Running"

WorkflowType = "SalesOrderSubmitSaga"
  AND StartTime BETWEEN "2025-07-01" AND "2025-07-04"
  AND ExecutionStatus = "Failed"
```

Set useful search attributes in workflow code:

```go
workflow.UpsertSearchAttributes(ctx, map[string]interface{}{
    "TenantID":   params.TenantID.String(),
    "EntityType": "Invoice",
    "EntityID":   params.InvoiceID.String(),
})
```

### 32.1.3. Workflow Status States

| Status | Meaning | Action |
|---|---|---|
| Running | Currently executing | Monitor |
| Completed | Finished successfully | Archive after retention period |
| Failed | Exhausted retries | Alert, investigate |
| TimedOut | Hit execution timeout | Alert, check for stuck activities |
| Cancelled | Externally cancelled | Normal if intentional |
| Terminated | Force-killed | Investigate — indicates ops intervention |
| ContinuedAsNew | Started a fresh run | Normal for long-running workflows |

---

## 32.2. OpenTelemetry Integration

### 32.2.1. Trace Propagation from Fiber into Temporal

Configure the Temporal SDK with the OpenTelemetry interceptor to propagate trace context:

```go
import "go.temporal.io/sdk/interceptor"
import "go.temporal.io/contrib/interceptors/opentelemetry"

c, err := client.Dial(client.Options{
    HostPort:  cfg.TemporalHost,
    Namespace: cfg.TemporalNamespace,
    Interceptors: []interceptor.ClientInterceptor{
        opentelemetry.NewTracingInterceptor(opentelemetry.TracerOptions{}),
    },
})
```

With this configured:
- The HTTP trace ID from the Fiber request is attached to the workflow start
- Each activity execution creates a child span
- The full trace (HTTP request → workflow start → activity executions) is visible in Jaeger/Tempo

### 32.2.2. Custom Spans Inside Activities

```go
func (a *Activities) PostGLEntries(ctx context.Context, input PostGLInput) error {
    ctx, span := otel.Tracer("awo/finance").Start(ctx, "PostGLEntries",
        trace.WithAttributes(
            attribute.String("tenant_id", input.TenantID.String()),
            attribute.String("invoice_id", input.InvoiceID.String()),
            attribute.Int("posting_count", len(input.Postings)),
        ),
    )
    defer span.End()

    for i, posting := range input.Postings {
        _, postSpan := otel.Tracer("awo/finance").Start(ctx, "CreateGLEntry",
            trace.WithAttributes(attribute.Int("posting_index", i)))
        _, err := a.glRepo.Create(ctx, posting)
        postSpan.End()
        if err != nil {
            span.RecordError(err)
            span.SetStatus(codes.Error, err.Error())
            return err
        }
    }
    return nil
}
```

### 32.2.3. Correlating Temporal Workflow ID with Distributed Traces

Add the Temporal workflow ID as a trace attribute on workflow start:

```go
func (a *Activities) StartWorkflowWithTrace(ctx context.Context, input WorkflowInput) error {
    // Extract trace context from activity context
    span := trace.SpanFromContext(ctx)
    spanCtx := span.SpanContext()

    options := client.StartWorkflowOptions{
        ID:        input.WorkflowID,
        TaskQueue: input.TaskQueue,
        SearchAttributes: map[string]interface{}{
            "TraceID": spanCtx.TraceID().String(),
            "SpanID":  spanCtx.SpanID().String(),
        },
    }
    // Start the workflow...
}
```

With the `TraceID` search attribute set, you can find the Temporal workflow from a Jaeger trace ID, and find the Jaeger trace from a Temporal workflow ID.

---

## 32.3. Workflow Alerting

### 32.3.1. Alert on Workflow Failure Rate

Monitor:
```
count(ExecutionStatus = "Failed") / count(*) > 0.05
→ PagerDuty alert: "Workflow failure rate exceeded 5%"
```

Temporal emits Prometheus metrics at `temporal:9090/metrics`. Key metric:
`temporal_workflow_failed_count` — counter of failed workflows by workflow type.

### 32.3.2. Alert on Workflow Pending Age — Stuck Workflows

A workflow stuck in "Running" for longer than expected indicates a blocked signal wait, hung activity, or worker outage:

```
temporal_activity_schedule_to_start_latency_seconds > 300
→ Alert: "Activity waiting >5 minutes in task queue (possible worker outage)"

count(ExecutionStatus = "Running" AND StartTime < now() - 24h
      AND WorkflowType = "InvoiceApprovalWorkflow") > 0
→ Alert: "Invoice approval workflow running >24 hours (may be stuck)"
```

### 32.3.3. Alert on Schedule Not Running

Monitor scheduled workflow last run time:

```go
// In a periodic health check workflow:
func ScheduleHealthCheckWorkflow(ctx workflow.Context, input ScheduleHealthInput) error {
    var schedules []ScheduleStatus
    workflow.ExecuteActivity(ctx, activities.ListScheduleStatuses).Get(ctx, &schedules)

    for _, s := range schedules {
        if s.LastRunAge > s.ExpectedInterval * 2 {
            // Schedule hasn't run in twice its expected interval
            workflow.ExecuteActivity(ctx, activities.SendAlert, Alert{
                Level:   "warning",
                Message: fmt.Sprintf("Schedule %s hasn't run in %s (expected every %s)",
                    s.ID, s.LastRunAge, s.ExpectedInterval),
            })
        }
    }
    return nil
}
```

Key schedules to alert on if they haven't run:
- `finance.gl-period-check.nightly` — if not run in >26 hours
- `inventory.wetstock-reconciliation.daily` — if not run in >26 hours
- `finance.period-close.monthly` — if not run within 2 days of month end
