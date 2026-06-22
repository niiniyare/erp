---
title: "Chapter 30: Signals, Queries, and Human-in-the-Loop"
part: "Part V — The Workflow Engine"
chapter: 30
section: "30-signals-queries"
related:
  - "[Chapter 27: Defining Workflows](27-defining-workflows.md)"
  - "[Chapter 29: Saga Pattern](29-saga-pattern.md)"
---

# Chapter 30: Signals, Queries, and Human-in-the-Loop

Signals enable external systems to communicate with running workflows. Queries enable external systems to read workflow state without affecting it. Together they power approval gates, human-in-the-loop processes, and live progress dashboards.

---

## 30.1. Signals

### 30.1.1. What Signals Are

A signal is an asynchronous message delivered to a running workflow. The workflow can be waiting for a signal — paused at a signal channel receive — and will resume when the signal arrives. Signals are at-least-once delivered by Temporal.

### 30.1.2. Sending a Signal from a Fiber Handler

```go
// Approval API endpoint
func (h *ApprovalHandler) Approve(c *fiber.Ctx) error {
    entityID := c.Params("id")
    workflowID := fmt.Sprintf("%s.Invoice.%s.approval",
        middleware.GetTenant(c).Slug, entityID)

    var body ApprovalRequest
    if err := c.BodyParser(&body); err != nil {
        return fiber.ErrBadRequest
    }

    decision := ApprovalDecision{
        Action:    "approved",
        ApprovedBy: middleware.GetActorID(c),
        Comment:   body.Comment,
        Timestamp: time.Now(),
    }

    err := h.temporal.SignalWorkflow(c.Context(),
        workflowID, "",
        "approval-decision",
        decision,
    )
    if err != nil {
        if temporal.IsNotFoundError(err) {
            return c.Status(404).JSON(fiber.Map{
                "message": "no pending approval workflow found for this invoice",
            })
        }
        return err
    }

    return c.JSON(fiber.Map{"status": "approved", "workflow_id": workflowID})
}
```

### 30.1.3. Receiving a Signal Inside a Workflow

```go
func InvoiceApprovalWorkflow(ctx workflow.Context, params WorkflowParams) (WorkflowResult, error) {
    approvalCh := workflow.GetSignalChannel(ctx, "approval-decision")

    var decision ApprovalDecision
    timerCtx, cancelTimer := workflow.WithCancel(ctx)
    defer cancelTimer()

    selector := workflow.NewSelector(ctx)

    // Wait for signal or timeout
    selector.AddReceive(approvalCh, func(c workflow.ReceiveChannel, more bool) {
        c.Receive(ctx, &decision)
    })
    selector.AddFuture(workflow.NewTimer(timerCtx, 48*time.Hour), func(f workflow.Future) {
        decision = ApprovalDecision{Action: "escalated", Reason: "approval timeout"}
    })

    selector.Select(ctx)  // blocks until signal or timer fires

    // decision is now populated
    return handleApprovalDecision(ctx, params, decision)
}
```

### 30.1.4. Signal Delivery Guarantees

Signals are delivered at-least-once. Your workflow must handle duplicate signals gracefully:

```go
var alreadyDecided bool
selector.AddReceive(approvalCh, func(c workflow.ReceiveChannel, more bool) {
    var decision ApprovalDecision
    c.Receive(ctx, &decision)
    if !alreadyDecided {
        alreadyDecided = true
        finalDecision = decision
    }
    // Ignore duplicate signals
})
```

### 30.1.5. Signal Naming Conventions

```
{action}-{noun}

approval-decision
payment-received
shipment-confirmed
cancellation-requested
```

Signal names must be unique within a workflow. Use descriptive, noun-verb names that describe the event from the external system's perspective.

---

## 30.2. Queries

### 30.2.1. What Queries Are

A query is a synchronous read of workflow state, with no side effects. Queries can be called at any time against a running or completed workflow. They do not affect the workflow's execution.

### 30.2.2. Defining a Query Handler Inside a Workflow

```go
func InvoiceApprovalWorkflow(ctx workflow.Context, params WorkflowParams) (WorkflowResult, error) {
    // State tracked throughout the workflow
    state := &ApprovalWorkflowState{
        Stage:           "pending",
        PendingApprovers: params.RequiredApprovers,
        StartedAt:       workflow.Now(ctx),
    }

    // Register query handler — callable at any time
    err := workflow.SetQueryHandler(ctx, "workflow-state", func() (*ApprovalWorkflowState, error) {
        return state, nil
    })
    if err != nil {
        return WorkflowResult{}, err
    }

    // ... rest of workflow, updating state as it progresses
    state.Stage = "awaiting_branch_manager"
    // wait for signal...
    state.Stage = "awaiting_finance_director"
    // wait for signal...
    state.Stage = "completed"
    state.CompletedAt = ptr.Time(workflow.Now(ctx))

    return WorkflowResult{}, nil
}
```

### 30.2.3. Calling a Query from a Fiber Handler

```go
func (h *InvoiceHandler) ApprovalStatus(c *fiber.Ctx) error {
    workflowID := fmt.Sprintf("%s.Invoice.%s.approval",
        middleware.GetTenant(c).Slug, c.Params("id"))

    response, err := h.temporal.QueryWorkflow(c.Context(),
        workflowID, "",
        "workflow-state",
    )
    if err != nil {
        if temporal.IsNotFoundError(err) {
            return c.Status(404).JSON(fiber.Map{
                "message": "no approval workflow found",
            })
        }
        return err
    }

    var state ApprovalWorkflowState
    if err := response.Get(&state); err != nil {
        return err
    }

    return c.JSON(fiber.Map{"data": state})
}
```

### 30.2.4. What to Expose via Queries

Good things to expose:
- Current workflow stage / step name
- List of pending approvers and their deadlines
- History of completed steps (without internal implementation details)
- Estimated completion time
- Error details (if in a failed state)

Do NOT expose:
- Internal implementation details (activity names, retry counts)
- Sensitive data not visible to the querying actor
- Temporal-internal state

---

## 30.3. Approval Gate Pattern

### 30.3.1. The `WaitForApproval` Activity

Reusable approval gate:

```go
type ApprovalGate struct {
    WorkflowID    string
    Stage         string
    RequiredRoles []string
    TimeoutHours  int
    OnTimeout     string  // "escalate" | "reject" | "auto_approve"
}

func waitForApprovalGate(ctx workflow.Context, gate ApprovalGate) (ApprovalDecision, error) {
    // Notify approvers via activity
    workflow.ExecuteActivity(ctx, activities.NotifyApprovers, gate)

    signalCh := workflow.GetSignalChannel(ctx, fmt.Sprintf("approval-%s", gate.Stage))
    var decision ApprovalDecision

    timeout := time.Duration(gate.TimeoutHours) * time.Hour
    selector := workflow.NewSelector(ctx)
    selector.AddReceive(signalCh, func(c workflow.ReceiveChannel, more bool) {
        c.Receive(ctx, &decision)
    })
    selector.AddFuture(workflow.NewTimer(ctx, timeout), func(f workflow.Future) {
        switch gate.OnTimeout {
        case "escalate":
            decision = ApprovalDecision{Action: "escalated", Reason: "timeout"}
        case "reject":
            decision = ApprovalDecision{Action: "rejected", Reason: "approval timeout"}
        case "auto_approve":
            decision = ApprovalDecision{Action: "approved", Reason: "auto-approved after timeout"}
        }
    })
    selector.Select(ctx)

    // Update state for query visibility
    workflow.UpsertSearchAttributes(ctx, map[string]interface{}{
        "ApprovalStage":  gate.Stage,
        "ApprovalStatus": decision.Action,
    })

    return decision, nil
}
```

### 30.3.2. The Approval API Endpoint

```go
// POST /api/v1/invoices/:id/approve
// Requires permission: Invoice:approve
func (h *ApprovalHandler) ApproveInvoice(c *fiber.Ctx) error {
    // Permission check
    if err := middleware.RequirePermission("Invoice:approve", "execute")(c); err != nil {
        return err
    }

    invoiceID := c.Params("id")
    actor := middleware.GetActor(c)

    var body struct {
        Stage   string `json:"stage"`    // which approval gate
        Comment string `json:"comment"`
    }
    c.BodyParser(&body)

    decision := ApprovalDecision{
        Action:    "approved",
        ApprovedBy: actor.ID,
        Comment:   body.Comment,
        Stage:     body.Stage,
    }

    workflowID := fmt.Sprintf("%s.Invoice.%s.approval",
        middleware.GetTenant(c).Slug, invoiceID)

    return h.temporal.SignalWorkflow(c.Context(), workflowID, "",
        fmt.Sprintf("approval-%s", body.Stage), decision)
}
```

### 30.3.3. Approval Timeout Handling

Never let a critical ERP approval workflow hang indefinitely. Every approval gate must have a timeout with a defined fallback:

| Scenario | Timeout | On timeout |
|---|---|---|
| Branch manager approval (<KES 50K) | 24 hours | Auto-approve |
| Finance director approval (KES 50K-500K) | 48 hours | Escalate to MD |
| MD approval (>KES 500K) | 72 hours | Notify CFO, leave pending |
| Emergency change request | 4 hours | Reject, require resubmission |

---

## 30.4. Worked Example — Shift Close Approval Flow

### 30.4.1. Cashier Submits Shift Close Request

```
POST /api/v1/shift-closes
{ "shift_id": "uuid", "cash_amount": 45000, "discrepancy": -500 }

→ ShiftClose record created (status: "pending_approval")
→ after_save hook triggers ShiftCloseApprovalWorkflow
```

### 30.4.2. Workflow Starts, Notifies Supervisor

```go
func ShiftCloseApprovalWorkflow(ctx workflow.Context, params ShiftCloseParams) error {
    // Register query handler for live status
    state := &ShiftCloseState{Stage: "awaiting_supervisor", StartedAt: workflow.Now(ctx)}
    workflow.SetQueryHandler(ctx, "state", func() (*ShiftCloseState, error) {
        return state, nil
    })

    // Notify supervisor
    workflow.ExecuteActivity(ctx, activities.NotifySupervisor, params)

    // Wait for supervisor decision
    decision, err := waitForApprovalGate(ctx, ApprovalGate{
        Stage:         "supervisor",
        RequiredRoles: []string{"shift_supervisor"},
        TimeoutHours:  2,
        OnTimeout:     "reject",
    })
    if err != nil {
        return err
    }

    if decision.Action == "approved" {
        state.Stage = "posting_gl"
        workflow.ExecuteActivity(ctx, activities.PostShiftCloseGL, params)
        workflow.ExecuteActivity(ctx, activities.UpdateShiftStatus, params.ShiftID, "closed")
        state.Stage = "completed"
    } else {
        workflow.ExecuteActivity(ctx, activities.UpdateShiftStatus, params.ShiftID, "rejected")
        workflow.ExecuteActivity(ctx, activities.NotifyCashierRejection, params, decision.Reason)
        state.Stage = "rejected"
    }

    return nil
}
```

### 30.4.3. Supervisor Reviews via AmisDetailPage

The supervisor opens the shift close detail page in amis. The page builder queries the workflow state:

```go
func BuildShiftCloseDetailPage(ctx PageBuilderContext) *amis.Schema {
    // Workflow state is fetched via API and displayed
    statusCard := amis.Service("GET /api/v1/shift-closes/${id}/approval-status").
        AutoRefresh(10000)  // poll every 10s

    if ctx.Permissions.CanExecute("ShiftClose", "approve") {
        approveBtn := amis.ActionButton("Approve").
            API("POST /api/v1/shift-closes/${id}/approve").
            VisibleOn("${approval_status.stage === 'awaiting_supervisor'}")
        rejectBtn := amis.ActionButton("Reject").
            API("POST /api/v1/shift-closes/${id}/reject").
            Level("danger").
            VisibleOn("${approval_status.stage === 'awaiting_supervisor'}")
    }
    // ...
}
```

### 30.4.4. Supervisor Approves

```
POST /api/v1/shift-closes/{id}/approve
{ "stage": "supervisor", "comment": "Cash verified, approved" }

→ Signal sent to workflow: approval-supervisor = { action: "approved" }
→ Workflow resumes, runs PostShiftCloseGL activity
→ GL entries posted, shift status updated to "closed"
→ Cashier sees status update in real-time (10s poll)
```
