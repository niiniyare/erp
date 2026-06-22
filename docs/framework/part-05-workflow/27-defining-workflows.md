---
title: "Chapter 27: Defining Workflows"
part: "Part V — The Workflow Engine"
chapter: 27
section: "27-defining-workflows"
related:
  - "[Chapter 26: Temporal Fundamentals](26-temporal-fundamentals.md)"
  - "[Chapter 28: Defining Activities](28-defining-activities.md)"
  - "[Chapter 30: Signals, Queries, and Human-in-the-Loop](30-signals-queries.md)"
---

# Chapter 27: Defining Workflows

Workflows are the orchestration layer in Awo. They coordinate sequences of activities, implement approval gates, and handle compensations. Writing correct workflows requires understanding Temporal's determinism requirement and its versioning mechanism.

---

## 27.1. Workflow Function Anatomy

### 27.1.1. Function Signature

```go
func MyWorkflow(ctx workflow.Context, input MyWorkflowInput) (MyWorkflowResult, error)
```

- `ctx workflow.Context` — Temporal's workflow context (not `context.Context`)
- `input` — serialisable input struct; must be JSON-serialisable
- `(result, error)` — both serialisable; returned to the caller

```go
type InvoiceApprovalInput struct {
    TenantID  uuid.UUID `json:"tenant_id"`
    InvoiceID uuid.UUID `json:"invoice_id"`
    ActorID   uuid.UUID `json:"actor_id"`
    Amount    string    `json:"amount"`   // decimal as string for serialisation safety
}

type InvoiceApprovalResult struct {
    Decision  string    `json:"decision"`   // "approved" | "rejected" | "escalated"
    DecidedBy uuid.UUID `json:"decided_by"`
    DecidedAt time.Time `json:"decided_at"`
}
```

### 27.1.2. What Is Allowed Inside a Workflow — Determinism Rules

A Temporal workflow is replayed from history on every worker restart. Replay must produce exactly the same sequence of decisions as the original execution. This means:

**ALLOWED:**
- Calling activities via `workflow.ExecuteActivity()`
- Waiting for signals via `workflow.GetSignalChannel()`
- Sleeping via `workflow.Sleep()`
- Getting current time via `workflow.Now()` (not `time.Now()`)
- Using `workflow.SideEffect()` for non-deterministic values
- Calling child workflows
- Pure computation (no I/O, no randomness)

**FORBIDDEN (non-deterministic):**
- `time.Now()` — use `workflow.Now()`
- `rand.Int()` — use `workflow.SideEffect()`
- Direct DB queries — use activities
- HTTP calls — use activities
- Goroutines (use `workflow.Go()` instead)
- Global state mutations

### 27.1.3. What Is Forbidden Inside a Workflow

Breaking the determinism requirement causes a `nondeterministic error` on replay, which puts the workflow into a failed state. Common mistakes:

```go
// WRONG — time.Now() returns different values on replay
if time.Now().After(deadline) {
    return nil, ErrExpired
}

// CORRECT — workflow.Now() is sourced from Temporal history
if workflow.Now(ctx).After(deadline) {
    return nil, ErrExpired
}

// WRONG — direct DB access
invoice, _ := db.GetInvoice(ctx, invoiceID)

// CORRECT — call an activity
var invoice Invoice
workflow.ExecuteActivity(ctx, activities.GetInvoice, invoiceID).Get(ctx, &invoice)
```

### 27.1.4. `workflow.Now()`, `workflow.Sleep()`, `workflow.SideEffect()`

```go
// Current time (deterministic — from Temporal clock)
now := workflow.Now(ctx)

// Sleep (durable — survives process restarts)
workflow.Sleep(ctx, 24*time.Hour)  // workflow pauses for 24 hours

// SideEffect — for non-deterministic values that must be consistent across replays
var randomID string
workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
    return uuid.New().String()
}).Get(&randomID)
// randomID is generated once and stored in history; same value on replay
```

---

## 27.2. Workflow Options

### 27.2.1. `WorkflowExecutionTimeout` — Hard Outer Limit

The maximum total duration of a workflow, including all retries and `continue-as-new` runs. After this timeout, the workflow is automatically cancelled.

```go
client.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
    ID:                       workflowID,
    TaskQueue:                "awo-default",
    WorkflowExecutionTimeout: 30 * 24 * time.Hour,  // 30 days for approval workflows
}, InvoiceApprovalWorkflow, input)
```

### 27.2.2. `WorkflowRunTimeout` — Per-Run Limit

Maximum duration of a single workflow run. For workflows that use `continue-as-new` to avoid history size limits, this is the limit per run (not the total execution):

```go
WorkflowRunTimeout: 24 * time.Hour  // each run limited to 24 hours
```

### 27.2.3. `WorkflowTaskTimeout`

Maximum time for the Temporal SDK to process a single workflow task (one step in the replay). Default is 10 seconds. Only increase if your workflow has very complex decision logic (unusual).

### 27.2.4. Retry Policy for Workflows

Generally: **do not retry workflows**. Retry activities instead. Retrying a workflow from the beginning after it fails mid-execution can cause double-posting and other idempotency issues.

```go
client.StartWorkflowOptions{
    RetryPolicy: &temporal.RetryPolicy{
        MaxAttempts: 1,  // no retry — activities handle their own retries
    },
}
```

Exception: short, fast workflows with no intermediate side effects can be safely retried.

### 27.2.5. Search Attributes — Indexable Metadata

Search attributes allow filtering and searching workflows in the Temporal Web UI:

```go
workflow.UpsertSearchAttributes(ctx, map[string]interface{}{
    "TenantID":      params.TenantID.String(),
    "EntityType":    "Invoice",
    "EntityID":      params.InvoiceID.String(),
    "WorkflowStage": "awaiting_approval",
    "CustomKeyword": "high_priority",
})
```

With search attributes, operators can find "all invoice approval workflows for tenant X that are stuck in awaiting_approval state" from the Temporal Web UI.

---

## 27.3. Workflow Versioning

### 27.3.1. Why Workflow Code Changes Break Running Workflows

When a workflow is replayed from history, it re-executes the workflow code. If the code has changed (new activity added, conditional logic changed), the replay produces a different sequence of decisions than the history recorded. Temporal detects this as a `nondeterministic error`.

This means: **you cannot simply change workflow code while instances are running**.

### 27.3.2. `workflow.GetVersion()` — Safe Branching Logic

```go
func InvoiceApprovalWorkflow(ctx workflow.Context, params WorkflowParams) (WorkflowResult, error) {
    // Before change: just approved or rejected
    // After change: also supports "delegate" action

    v := workflow.GetVersion(ctx, "add-delegate-action", workflow.DefaultVersion, 1)

    if v == workflow.DefaultVersion {
        // Old path — existing running workflows replay this branch
        // ... original approval logic
    } else {
        // New path — new workflow instances execute this branch
        // ... logic that also handles "delegate" action
    }
}
```

`workflow.GetVersion()` records a "marker" event in the history on first execution. On replay, it reads the marker and returns the same version — ensuring old running workflows continue on the old code path.

### 27.3.3. Version Deprecation

After all running instances on the old code path have completed:

```go
// Remove the old branch — all instances are now on version 1
v := workflow.GetVersion(ctx, "add-delegate-action", 1, 1)
// Now v is always 1; the DefaultVersion branch can be deleted
```

Monitor the Temporal Web UI to confirm no instances are still running on the old version before removing the old branch.

### 27.3.4. `continue-as-new` — For Indefinitely Running Workflows

Temporal's history has a size limit (~50,000 events). For workflows that loop indefinitely (e.g. a site reconciliation workflow that runs every day forever), use `continue-as-new` to start a fresh workflow run while preserving the logical continuation:

```go
func SiteReconciliationWorkflow(ctx workflow.Context, params ReconciliationParams) error {
    if err := runDailyReconciliation(ctx, params); err != nil {
        return err
    }

    // After 365 runs, start a new workflow run (fresh history)
    params.RunCount++
    if params.RunCount >= 365 {
        return workflow.NewContinueAsNewError(ctx, SiteReconciliationWorkflow,
            ReconciliationParams{SiteID: params.SiteID, RunCount: 0})
    }

    // Sleep until next day (EAT midnight)
    nextRun := nextMidnightEAT(workflow.Now(ctx))
    workflow.Sleep(ctx, time.Until(nextRun))

    return workflow.NewContinueAsNewError(ctx, SiteReconciliationWorkflow, params)
}
```

---

## 27.4. Child Workflows

### 27.4.1. When to Decompose Into Child Workflows

Use child workflows when:
- A sub-process has its own timeout, retry, and versioning lifecycle
- You want to view the sub-process independently in the Temporal Web UI
- History size would become problematic in a monolithic workflow

```go
// Parent: SalesOrderSaga
// Child workflows: InventoryReservation, GLPosting, DeliveryNoteCreation
```

### 27.4.2. Fire and Wait vs Fire and Forget

```go
// Fire and wait — parent blocks until child completes
var childResult ChildResult
err := workflow.ExecuteChildWorkflow(ctx,
    workflow.ChildWorkflowOptions{
        WorkflowID: childWorkflowID,
    },
    InventoryReservationWorkflow, input,
).Get(ctx, &childResult)

// Fire and forget — parent continues without waiting
childFuture := workflow.ExecuteChildWorkflow(ctx, ..., NotificationWorkflow, input)
// Do not call .Get() — parent continues immediately
```

### 27.4.3. Parent-Child Cancellation Propagation

When a parent workflow is cancelled, all child workflows started with `ParentClosePolicy = ABANDON` continue running. Child workflows started with `ParentClosePolicy = TERMINATE` are cancelled.

Use `TERMINATE` for child workflows that should not run if the parent is cancelled (e.g. stop sending notifications if the main workflow is cancelled). Use `ABANDON` for child workflows that represent independent business operations that must complete regardless (e.g. GL posting that has already started).

---

## 27.5. Workflow ID Naming Conventions

### 27.5.1. The `{tenant}.{entity}.{id}.{action}` Pattern

```go
workflowID := fmt.Sprintf("%s.%s.%s.%s",
    tenantSlug,
    "Invoice",
    invoiceID.String(),
    "approval",
)
// e.g. "acme.Invoice.550e8400-....approval"
```

### 27.5.2. Idempotency via Workflow ID

Starting a workflow with an ID that already exists (and is still running) is a no-op by default (`TerminateIfRunning: false`). This provides natural idempotency:
- If a user submits an invoice twice, only one approval workflow runs
- If the `after_save` hook fires twice (edge case), only one workflow starts

### 27.5.3. Workflow ID Policies

```go
client.StartWorkflowOptions{
    WorkflowIDReusePolicy: enumspb.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
    // REJECT_DUPLICATE: fail if a completed workflow with this ID exists (default for new business operations)
    // ALLOW_DUPLICATE: allow starting a new workflow even if one already completed with this ID
    // TERMINATE_IF_RUNNING: cancel any running workflow with this ID before starting the new one
}
```

For most ERP workflows: `REJECT_DUPLICATE` prevents accidental re-processing of completed documents.
