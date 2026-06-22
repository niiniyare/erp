---
title: "Chapter 26: Temporal Fundamentals for Awo"
part: "Part V — The Workflow Engine"
chapter: 26
section: "26-temporal-fundamentals"
related:
  - "[Chapter 27: Defining Workflows](27-defining-workflows.md)"
  - "[Chapter 28: Defining Activities](28-defining-activities.md)"
  - "[Chapter 7: The EntityRecord Lifecycle](../part-02-entity-system/07-entity-record-lifecycle.md)"
---

# Chapter 26: Temporal Fundamentals for Awo

Temporal is Awo's durable workflow engine. It handles all asynchronous, long-running, and compensating operations — from invoice approval chains that wait for human input to nightly reconciliation jobs that process thousands of records. This chapter explains why Temporal was chosen, its core concepts in the Awo context, and how to set up and register workers.

---

## 26.1. Why Temporal Over Queues or Cron

### 26.1.1. Durability — Workflow State Survives Process Crashes

With a message queue (RabbitMQ, SQS, Kafka), if the worker process crashes mid-operation, the message may be re-delivered but the partial work already done is lost. You must implement your own checkpointing, retry tracking, and partial-progress recovery.

With Temporal, every step of a workflow is persisted to the database before execution. If the worker crashes:
- Temporal replays the workflow history on restart
- The workflow continues from the last completed activity
- No application-level checkpointing code is required

This durability is not optional for ERP. An invoice approval workflow that was waiting for a manager's response must resume exactly where it left off after a deployment restart — not start over from the beginning.

### 26.1.2. Audit Trail — Complete Execution History

Every workflow event (activity started, activity completed, signal received, timer fired) is stored in the Temporal history with a timestamp. This history is queryable via the Temporal Web UI and API for months after completion (configurable retention period).

For compliance, this means: "show me every step that happened when processing invoice INV-2025-00042" is answerable from the Temporal history, without building a separate audit log.

### 26.1.3. Long-Running Transactions — Approval Chains

A purchase order approval chain in a Kenyan company might require:
1. Branch manager approval (within 24 hours)
2. Finance director approval (within 48 hours, only for amounts >KES 500,000)
3. MD approval (within 72 hours, only for amounts >KES 2,000,000)

This workflow spans days and involves human actors. A database transaction cannot span days. A message queue cannot express the conditional routing. Temporal models this as a durable workflow with signal-based approval gates.

### 26.1.4. Saga Pattern — Compensating Transactions

When a sales order requires:
1. Reserve inventory (step 1)
2. Post GL entries (step 2)
3. Create delivery note (step 3)

...and step 3 fails, steps 1 and 2 must be reversed. Without Temporal, this requires a complex distributed transaction coordinator or custom idempotency tables. With Temporal, the saga pattern is implemented cleanly with activity compensations (Chapter 29).

---

## 26.2. Core Temporal Concepts

### 26.2.1. Workflow — A Durable Orchestrating Function

A workflow is a Go function marked with `@workflow.defn`. It calls activities, waits for signals, sleeps, and returns a result. Workflow code must be deterministic — given the same history, re-execution must produce the same decisions.

```go
func InvoiceApprovalWorkflow(ctx workflow.Context, params WorkflowParams) (WorkflowResult, error) {
    logger := workflow.GetLogger(ctx)
    logger.Info("InvoiceApproval started", "invoiceID", params.InvoiceID)

    // Step 1: Notify approver
    ao := workflow.ActivityOptions{StartToCloseTimeout: 30 * time.Second}
    ctx = workflow.WithActivityOptions(ctx, ao)
    workflow.ExecuteActivity(ctx, activities.NotifyApprover, params)

    // Step 2: Wait for approval signal (up to 48 hours)
    approvalCh := workflow.GetSignalChannel(ctx, "approval-decision")
    var decision ApprovalDecision

    selector := workflow.NewSelector(ctx)
    selector.AddReceive(approvalCh, func(c workflow.ReceiveChannel, more bool) {
        c.Receive(ctx, &decision)
    })
    // Timeout: escalate if no decision within 48 hours
    timerFuture := workflow.NewTimer(ctx, 48*time.Hour)
    selector.AddFuture(timerFuture, func(f workflow.Future) {
        decision = ApprovalDecision{Action: "escalated"}
    })
    selector.Select(ctx)

    // Step 3: Act on decision
    if decision.Action == "approved" {
        workflow.ExecuteActivity(ctx, activities.PostGLEntries, params)
        workflow.ExecuteActivity(ctx, activities.UpdateInvoiceStatus, params, "approved")
    } else if decision.Action == "rejected" {
        workflow.ExecuteActivity(ctx, activities.UpdateInvoiceStatus, params, "rejected")
        workflow.ExecuteActivity(ctx, activities.NotifyApplicant, params, decision.Reason)
    }

    return WorkflowResult{Decision: decision.Action}, nil
}
```

### 26.2.2. Activity — A Function That Touches the Outside World

Activities are Go functions that perform the actual work: DB queries, API calls, file operations, email sending. They can fail, be retried, and can report progress via heartbeats.

```go
type Activities struct {
    invoiceRepo InvoiceRepository
    glService   GLService
    notifier    NotificationService
}

func (a *Activities) PostGLEntries(ctx context.Context, params WorkflowParams) error {
    // Re-establish tenant context from params
    ctx = tenant.WithIDContext(ctx, params.TenantID)
    return a.glService.PostInvoice(ctx, params.InvoiceID)
}
```

### 26.2.3. Worker — A Process That Executes Workflows and Activities

A worker polls Temporal for tasks, executes workflow and activity functions, and returns results. Awo can run the worker:
- **Co-located with the API server** (development, small deployments): single process runs both Fiber and Temporal worker
- **As a separate process** (production, scaled deployments): worker scales independently of the API server

### 26.2.4. Task Queue — Named Channel That Routes Work to Workers

Every workflow and activity is associated with a task queue. Workers poll specific task queues. Awo uses task queue names to route work to the correct worker type:

```
awo-default        — standard ERP workflows (invoice approval, GL posting)
awo-finance        — finance-specific activities (period close, reconciliation)
awo-notifications  — notification sending (email, SMS, push)
awo-bulk           — large batch operations (bulk import, bulk export)
```

### 26.2.5. Workflow ID — Unique, Stable Identity

The workflow ID is a human-readable string that uniquely identifies a workflow instance within a Temporal namespace:

```
{tenant_slug}.Invoice.{invoice_uuid}.approval
acme.Invoice.550e8400-....approval

{tenant_slug}.{entity}.{record_id}.{action}
```

Using the entity and record ID in the workflow ID provides natural idempotency: if the same approval workflow is triggered twice for the same invoice, the second trigger is a no-op (the workflow is already running).

### 26.2.6. Temporal Namespace — Isolation Boundary

Awo uses a single Temporal namespace (`awo`) for all tenants. Tenant isolation within the namespace is achieved by:
- Workflow IDs prefixed with tenant slug
- Search attributes including `TenantID` for filtering in the Web UI
- Activity code re-establishing tenant context from workflow params before any DB access

For very high security requirements, a dedicated Temporal namespace per tenant is supported but requires additional infrastructure.

---

## 26.3. Awo Worker Process Setup

### 26.3.1. Running the Worker — Deployment Options

**Co-located with API (development/small deployments):**

```go
// In main.go, after starting Fiber:
go func() {
    w := worker.New(temporalClient, "awo-default", worker.Options{
        MaxConcurrentActivityExecutionSize:      50,
        MaxConcurrentWorkflowTaskExecutionSize:  20,
    })
    registerWorkflowsAndActivities(w)
    if err := w.Run(worker.InterruptCh()); err != nil {
        slog.Error("temporal worker stopped", "error", err)
    }
}()
```

**Separate process (production):**

```go
// cmd/worker/main.go
func main() {
    cfg, _ := config.Load()
    temporalClient, _ := client.Dial(client.Options{HostPort: cfg.TemporalHost})
    defer temporalClient.Close()

    w := worker.New(temporalClient, "awo-default", worker.Options{
        MaxConcurrentActivityExecutionSize: 100,
    })
    registerWorkflowsAndActivities(w)
    if err := w.Run(worker.InterruptCh()); err != nil {
        log.Fatal(err)
    }
}
```

### 26.3.2. Task Queue Naming Conventions

```
{module}.{entity}.{action}

finance.invoice.approval
finance.gl.period-close
inventory.stock.transfer
notifications.email.send
```

Use the module-specific task queue for activities that should only run on workers with that module's dependencies loaded. Use `awo-default` for general-purpose activities.

### 26.3.3. Worker Concurrency Configuration

```go
worker.Options{
    // Max activities running simultaneously per worker process
    MaxConcurrentActivityExecutionSize: 50,

    // Max workflow tasks (replay events) running simultaneously
    MaxConcurrentWorkflowTaskExecutionSize: 20,

    // For bulk workers: allow more parallelism
    // MaxConcurrentActivityExecutionSize: 200,
}
```

Rule of thumb: `MaxConcurrentActivityExecutionSize = (available DB connections) / (activities per request)`. With 20 DB connections and each activity using 1 connection: max 20 concurrent activities.

### 26.3.4. Registering Workflows and Activities

```go
func registerWorkflowsAndActivities(w worker.Worker) {
    // Register workflows
    w.RegisterWorkflow(InvoiceApprovalWorkflow)
    w.RegisterWorkflow(SalesOrderSagaWorkflow)
    w.RegisterWorkflow(PeriodCloseWorkflow)
    w.RegisterWorkflow(DailyReconciliationWorkflow)

    // Register activities (injected with dependencies)
    acts := &Activities{
        invoiceRepo: provideInvoiceRepo(),
        glService:   provideGLService(),
        notifier:    provideNotificationService(),
    }
    w.RegisterActivity(acts.PostGLEntries)
    w.RegisterActivity(acts.UpdateInvoiceStatus)
    w.RegisterActivity(acts.NotifyApprover)
    w.RegisterActivity(acts.NotifyApplicant)
    w.RegisterActivity(acts.ReserveInventory)
    w.RegisterActivity(acts.ReleaseInventoryReservation)  // compensation
}
```

Activities are registered as methods on a struct, enabling dependency injection without global variables.
