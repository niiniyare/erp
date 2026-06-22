---
title: "Chapter 28: Defining Activities"
part: "Part V — The Workflow Engine"
chapter: 28
section: "28-defining-activities"
related:
  - "[Chapter 27: Defining Workflows](27-defining-workflows.md)"
  - "[Chapter 29: Saga Pattern](29-saga-pattern.md)"
---

# Chapter 28: Defining Activities

Activities are the units of actual work in Temporal — they interact with databases, external APIs, file systems, and notification services. Correct activity design (idempotency, retry configuration, heartbeating) is critical for production reliability.

---

## 28.1. Activity Function Anatomy

### 28.1.1. Function Signature

```go
func (a *Activities) PostGLEntries(ctx context.Context, input PostGLInput) (PostGLResult, error)
```

Key differences from workflow functions:
- Uses `context.Context`, not `workflow.Context`
- **Can** make DB calls, HTTP calls, and any I/O
- **Can** use `time.Now()` (not required to be deterministic)
- Must be idempotent (will be retried on failure)

### 28.1.2. The `Activities` Struct Pattern — Dependency Injection

```go
type FinanceActivities struct {
    invoiceRepo InvoiceRepository
    glRepo      GLEntryRepository
    mailer      MailService
    temporal    client.Client
}

func NewFinanceActivities(
    invoiceRepo InvoiceRepository,
    glRepo GLEntryRepository,
    mailer MailService,
    tc client.Client,
) *FinanceActivities {
    return &FinanceActivities{
        invoiceRepo: invoiceRepo,
        glRepo:      glRepo,
        mailer:      mailer,
        temporal:    tc,
    }
}

func (a *FinanceActivities) PostGLEntries(ctx context.Context, input PostGLInput) error {
    // Reconstruct tenant context from input params
    ctx = tenant.WithIDContext(ctx, input.TenantID)

    invoice, err := a.invoiceRepo.Get(ctx, input.InvoiceID)
    if err != nil {
        return err
    }

    postings := computeGLPostings(invoice)
    for _, posting := range postings {
        if _, err := a.glRepo.Create(ctx, posting); err != nil {
            return err
        }
    }
    return nil
}
```

### 28.1.3. Accessing EntityRepository Inside an Activity

The activity accesses repositories directly — they are injected via the `Activities` struct. The tenant context must be re-established from the workflow params (not from an HTTP context, which doesn't exist in an activity):

```go
func (a *FinanceActivities) SendInvoiceEmail(ctx context.Context, input SendEmailInput) error {
    // Re-establish tenant context
    ctx = tenant.WithIDContext(ctx, input.TenantID)

    // Now all repository calls are scoped to the correct tenant schema
    invoice, err := a.invoiceRepo.Get(ctx, input.InvoiceID)
    if err != nil {
        return err
    }

    customer, err := a.customerRepo.Get(ctx, invoice.CustomerID)
    if err != nil {
        return err
    }

    return a.mailer.Send(ctx, MailMessage{
        To:      customer.Email,
        Subject: fmt.Sprintf("Invoice %s from %s", invoice.InvoiceNumber, invoice.TenantName),
        Body:    renderInvoiceEmailTemplate(invoice, customer),
    })
}
```

---

## 28.2. Activity Options

### 28.2.1. `StartToCloseTimeout` — Required, Always Set Explicitly

Every activity must have a `StartToCloseTimeout`. This is the maximum time an activity is allowed to run (from the time it starts executing to the time it returns):

```go
ao := workflow.ActivityOptions{
    StartToCloseTimeout: 30 * time.Second,  // for fast DB operations
}

// For activities that call external APIs:
ao := workflow.ActivityOptions{
    StartToCloseTimeout: 2 * time.Minute,
}

// For bulk processing activities:
ao := workflow.ActivityOptions{
    StartToCloseTimeout: 30 * time.Minute,
    HeartbeatTimeout:    30 * time.Second,  // must heartbeat every 30s
}
```

Never use very large timeouts without heartbeating. A hung activity holding a DB connection for 30 minutes is a production incident.

### 28.2.2. `ScheduleToStartTimeout` — Queue Wait Timeout

Maximum time an activity can wait in the task queue before a worker picks it up. Use when:
- Activity must execute promptly or not at all
- Worker backlog should be monitored

```go
ao := workflow.ActivityOptions{
    ScheduleToStartTimeout: 5 * time.Minute,  // fail if no worker available within 5 min
    StartToCloseTimeout:    30 * time.Second,
}
```

### 28.2.3. Retry Policy

```go
ao := workflow.ActivityOptions{
    StartToCloseTimeout: 30 * time.Second,
    RetryPolicy: &temporal.RetryPolicy{
        InitialInterval:    time.Second,      // first retry after 1 second
        BackoffCoefficient: 2.0,              // doubles each retry: 1s, 2s, 4s, 8s...
        MaxInterval:        time.Minute,      // caps at 1 minute
        MaxAttempts:        5,                // up to 5 total attempts
        NonRetryableErrorTypes: []string{
            "ErrCreditLimitExceeded",  // business errors should not be retried
            "ErrPeriodClosed",
        },
    },
}
```

**Critical**: declare `NonRetryableErrorTypes` for business errors. An activity that fails because the GL period is closed should not be retried — the same failure will occur on every retry until the period is reopened. Business errors that retry waste resources and delay the workflow's failure notification.

---

## 28.3. Heartbeating

### 28.3.1. When Heartbeating Is Required

Heartbeat any activity that:
- Takes more than 60 seconds to complete
- Processes items in a loop (batch processing)
- Blocks on external I/O (file reads, slow API calls)

Without heartbeating, Temporal cannot distinguish between an activity that is working slowly and an activity whose worker process has crashed. With heartbeating, worker crashes are detected within `HeartbeatTimeout` seconds.

### 28.3.2. `activity.RecordHeartbeat()` — Usage

```go
func (a *Activities) BulkImportProducts(ctx context.Context, input BulkImportInput) error {
    ctx = tenant.WithIDContext(ctx, input.TenantID)

    for i, row := range input.Rows {
        // Heartbeat every 100 rows — includes progress for resumption
        if i % 100 == 0 {
            activity.RecordHeartbeat(ctx, BulkImportProgress{
                ProcessedRows: i,
                TotalRows:     len(input.Rows),
                LastRowIndex:  i,
            })
        }

        // Check if activity was cancelled (via heartbeat)
        if ctx.Err() != nil {
            return ctx.Err()
        }

        if err := a.productRepo.Create(ctx, row.ToCreateInput()); err != nil {
            return fmt.Errorf("row %d: %w", i, err)
        }
    }
    return nil
}
```

### 28.3.3. Detecting Cancellation via Heartbeat Context

When a workflow is cancelled, activities receive cancellation through the context returned by `activity.GetHeartbeatDetails()`. The activity must check `ctx.Err()` regularly:

```go
if err := ctx.Err(); err != nil {
    // Graceful cleanup before returning
    cleanupPartialImport(ctx, importID)
    return temporal.NewCanceledError("import cancelled after %d rows", processed)
}
```

### 28.3.4. Heartbeat-Based Progress Reporting

The heartbeat details (the struct passed to `RecordHeartbeat`) are stored in Temporal history and retrievable via the query API. This enables live progress display in the SDUI:

```go
// Page builder polling endpoint:
// GET /api/v1/bulk-imports/{import_id}/progress
func (h *BulkImportHandler) Progress(c *fiber.Ctx) error {
    workflowID := fmt.Sprintf("bulk-import.%s", importID)
    response, err := temporalClient.QueryWorkflow(c.Context(), workflowID, "", "progress")
    // returns the last heartbeat details
}
```

---

## 28.4. Idempotent Activities

### 28.4.1. Why Activities Must Be Safe to Retry

Temporal retries failed activities from the beginning (not from mid-execution). If an activity has already partially completed (e.g. created 3 of 5 GL entries), retrying from the beginning will try to create all 5 again.

Activities must be designed so that running them twice produces the same result as running them once.

### 28.4.2. Database Upsert Pattern

```go
func (a *Activities) CreateGLEntry(ctx context.Context, input GLEntryInput) error {
    // Use upsert to make idempotent
    // The external_ref is derived from the workflow + step, stable across retries
    _, err := a.glRepo.Upsert(ctx, GLEntryCreate{
        ExternalRef:   input.IdempotencyKey,  // "workflow-id.step-name"
        DebitAccount:  input.DebitAccountID,
        CreditAccount: input.CreditAccountID,
        Amount:        input.Amount,
        Description:   input.Description,
    }, UpsertOnConflict("external_ref"))
    return err
}
```

### 28.4.3. External API Idempotency Keys

For external APIs (KRA eTIMS, M-PESA), pass an idempotency key derived from stable workflow information:

```go
func (a *Activities) SubmitToKRAeTIMS(ctx context.Context, input KRASubmitInput) error {
    // Idempotency key: stable across retries
    idempotencyKey := fmt.Sprintf("awo.%s.%s", input.TenantID, input.InvoiceID)

    resp, err := a.kraClient.SubmitInvoice(ctx, KRAInvoiceRequest{
        InvoiceNumber:  input.InvoiceNumber,
        IdempotencyKey: idempotencyKey,
        // ...
    })
    if err != nil {
        // Check if it's a "duplicate submission" error — which means we already succeeded
        if kraErr, ok := err.(KRAError); ok && kraErr.Code == "DUPLICATE_SUBMISSION" {
            return nil  // success — already submitted
        }
        return err
    }
    return nil
}
```

---

## 28.5. Local Activities

### 28.5.1. What Local Activities Are

Local activities execute in the same worker process as the workflow, without creating a separate Temporal history event for the activity execution. They are faster (no round-trip to Temporal server) but have no independent retry tracking or heartbeat support.

### 28.5.2. When to Use Local Activities

Use local activities for:
- Fast lookups (cache reads, simple DB reads) that don't need retry tracking
- Pure computation that might fail transiently (retried inline)
- Operations that must complete within one workflow task timeout (10 seconds)

```go
// Local activity: fast lookup, no separate retry needed
lao := workflow.LocalActivityOptions{
    StartToCloseTimeout: 5 * time.Second,
}
ctx = workflow.WithLocalActivityOptions(ctx, lao)

var customerName string
workflow.ExecuteLocalActivity(ctx, a.GetCustomerName, customerID).Get(ctx, &customerName)
```

### 28.5.3. Local Activity Limitations

- No heartbeat support
- No separate timeout/retry from the workflow task
- If the worker crashes during a local activity, the workflow task (not just the activity) is retried
- Not visible as a separate event in Temporal Web UI (debugging is harder)

Use regular activities for anything that takes more than a few seconds or involves external API calls.
