[<-- Back to Index](README.md)

## Workflow Integration

### Why Workflows Need Authorization

Temporal workflows run asynchronously — sometimes hours or days after the initial request. The user who triggered the workflow may have had their role revoked by the time an activity executes. Workflows must re-check authorization at each sensitive activity, not just at trigger time.

```markdown
WORKFLOW AUTHORIZATION POINTS:

Trigger time (HTTP):
  → authn middleware validates JWT
  → authz.Middleware checks "workflow/invoice-approval/start" → "execute" allow
  → Workflow started with actor.Principal embedded

Activity time (async — could be minutes/hours later):
  → Activity receives actor.Principal from workflow input
  → Activity calls authz.Enforce before mutating data
  → If actor's role expired → DENY → activity fails → workflow compensates

ACTOR PRINCIPAL FLOWS THROUGH WORKFLOW:

type InvoiceApprovalInput struct {
    InvoiceID string
    Actor     authz.Principal  // embedded at trigger time, re-checked at activity
}
```

### Temporal Workflow Pattern

```go
// internal/workflows/invoice_approval.go

// Workflow definition
func InvoiceApprovalWorkflow(ctx workflow.Context, input InvoiceApprovalInput) error {
    // Step 1: Validate invoice (read-only, no authz needed if data is public within tenant)
    var invoice Invoice
    err := workflow.ExecuteActivity(ctx, activities.GetInvoice, input.InvoiceID).Get(ctx, &invoice)
    if err != nil {
        return fmt.Errorf("get invoice: %w", err)
    }

    // Step 2: Check approval authority (authz-protected activity)
    var approved bool
    err = workflow.ExecuteActivity(ctx, activities.CheckApprovalAuthority, CheckApprovalInput{
        Actor:     input.Actor,
        InvoiceID: input.InvoiceID,
        Amount:    invoice.Amount,
    }).Get(ctx, &approved)
    if err != nil {
        return fmt.Errorf("approval check: %w", err)
    }
    if !approved {
        return fmt.Errorf("insufficient approval authority")
    }

    // Step 3: Post to GL (authz-protected activity)
    return workflow.ExecuteActivity(ctx, activities.PostInvoiceToGL, PostGLInput{
        Actor:     input.Actor,
        InvoiceID: input.InvoiceID,
    }).Get(ctx, nil)
}
```

```go
// internal/workflows/activities/invoice_activities.go

type InvoiceActivities struct {
    authz authz.Service
    repo  InvoiceRepository
    gl    GLService
}

// CheckApprovalAuthority — re-checks actor's authorization at activity time
func (a *InvoiceActivities) CheckApprovalAuthority(ctx context.Context, input CheckApprovalInput) (bool, error) {
    // Determine object path based on amount (same logic as HTTP handler)
    obj := classifyInvoiceObject(input.InvoiceID, input.Amount)

    ok, err := a.authz.Enforce(ctx, authz.Request{
        Subject: input.Actor.Subject,
        Domain:  input.Actor.Domain,
        Object:  obj,
        Action:  "approve",
    })
    if err != nil {
        return false, fmt.Errorf("authz enforce: %w", err)
    }
    // If the actor's role expired since workflow started → ok=false → workflow aborts
    return ok, nil
}

// PostInvoiceToGL — authz check before GL mutation
func (a *InvoiceActivities) PostInvoiceToGL(ctx context.Context, input PostGLInput) error {
    ok, err := a.authz.Enforce(ctx, authz.Request{
        Subject: input.Actor.Subject,
        Domain:  input.Actor.Domain,
        Object:  "journal/invoice/" + input.InvoiceID,
        Action:  "post",
    })
    if err != nil || !ok {
        return authz.ErrForbidden
    }
    return a.gl.PostInvoice(ctx, input.InvoiceID)
}
```

### Service Account Workflows

Scheduled jobs and system-initiated workflows run as service accounts (API actors), not as human users. Their authorization is stable — service accounts don't have temporal roles.

```go
// Scheduled nightly reconciliation — runs as service account
func NightlyReconciliationWorkflow(ctx workflow.Context, tenantID string) error {
    // Service account principal — set at workflow start from job config
    actor := authz.Principal{
        Subject: authz.APISubject("cli_reconciliation_svc"),
        Domain:  authz.TenantDomain(tenantID),
    }

    // Enforce each activity
    err := workflow.ExecuteActivity(ctx, activities.ReconcilePayments, ReconcileInput{
        Actor:    actor,
        TenantID: tenantID,
    }).Get(ctx, nil)
    return err
}
```

### Approval Workflows with Human-in-the-Loop

Temporal supports human approval signals. When a manager must approve a workflow step:

```go
// Dual-approval workflow for large purchases
func LargePurchaseApprovalWorkflow(ctx workflow.Context, input PurchaseInput) error {
    // Step 1: Auto-check first approver (finance manager)
    ok, _ := activities.EnforceApproval(ctx, input.Actor, "purchase/large", "approve")
    if !ok {
        // Request human approval signal
        var signal ApprovalSignal
        workflow.GetSignalChannel(ctx, "approval").Receive(ctx, &signal)

        // Re-check the approver's authorization when signal arrives
        ok2, _ := activities.EnforceApproval(ctx, signal.ApproverPrincipal, "purchase/large", "approve")
        if !ok2 {
            return fmt.Errorf("approver %s lacks authority", signal.ApproverPrincipal.Subject)
        }
    }

    return activities.ExecutePurchase(ctx, input)
}
```

### Workflow Authorization Summary

```markdown
RULE: Every Temporal activity that mutates data MUST call authz.Enforce.

REASON: The actor's role may have been revoked between workflow start
and activity execution. Temporal retries mean an activity may run
multiple times — each run should re-check.

SERVICE ACCOUNTS:
  → Use api:{clientID} subject
  → Domain: {tenantID}:api or {tenantID} (if elevated grant)
  → Roles never expire (permanent assignment)
  → Principal embedded in workflow input at start time

HUMAN ACTORS:
  → Use tenant:{userID} subject
  → Principal embedded in workflow input at start time
  → Role may expire → activity fails gracefully → workflow compensates

COMPENSATION (Temporal saga pattern):
  If a later activity fails due to authz deny:
  → Workflow calls compensation activities to undo earlier steps
  → e.g., un-post a GL entry if the final approval fails
```

---

Next: [Performance & Caching](./16-performance-and-caching.md)
