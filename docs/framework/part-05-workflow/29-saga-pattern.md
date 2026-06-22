---
title: "Chapter 29: Saga Pattern — Compensating Transactions"
part: "Part V — The Workflow Engine"
chapter: 29
section: "29-saga-pattern"
related:
  - "[Chapter 28: Defining Activities](28-defining-activities.md)"
  - "[Chapter 27: Defining Workflows](27-defining-workflows.md)"
---

# Chapter 29: Saga Pattern — Compensating Transactions

A saga is a sequence of activities where each step has a corresponding compensation activity. If any step fails, all previously completed steps are compensated in reverse order. Awo provides a `saga` helper that manages the compensation stack automatically.

---

## 29.1. When to Use a Saga

### 29.1.1. Multi-Service Operations That Must Be Atomic in Business Terms

A sales order submission requires:
1. Reserve inventory for each line item
2. Post GL debit entries (cost of goods)
3. Post GL credit entries (sales revenue)
4. Create a delivery note

These four operations span multiple tables and conceptually represent one atomic business event. If step 3 fails after step 1 and 2 have completed, the inventory reservation must be released and the GL debits must be reversed.

### 29.1.2. Why Database Transactions Are Insufficient

A single DB transaction could cover all four operations — but only if they all complete within a few seconds. If step 4 involves calling an external logistics API (which may take 30 seconds or time out), holding a DB transaction for 30 seconds is unacceptable under any load.

Sagas replace long-held transactions with explicit compensation logic.

### 29.1.3. Long-Running Operations

Any business process that spans multiple seconds, minutes, or requires human approval cannot use a DB transaction. Sagas are the correct pattern.

---

## 29.2. Saga Implementation in Temporal

### 29.2.1. The `saga` Helper

```go
import "awo.so/workflow/saga"

func SalesOrderSubmitWorkflow(ctx workflow.Context, params SalesOrderParams) error {
    s := saga.New()

    // Step 1: Reserve inventory
    ao := workflow.ActivityOptions{StartToCloseTimeout: 30 * time.Second}
    ctx = workflow.WithActivityOptions(ctx, ao)

    var reservation InventoryReservation
    err := workflow.ExecuteActivity(ctx, activities.ReserveInventory, params).
        Get(ctx, &reservation)
    if err != nil {
        return err  // nothing to compensate yet
    }
    // Register compensation for step 1
    s.AddCompensation(func(ctx workflow.Context) {
        workflow.ExecuteActivity(ctx, activities.ReleaseInventoryReservation,
            reservation.ReservationID)
    })

    // Step 2: Post GL entries
    var glPostingID uuid.UUID
    err = workflow.ExecuteActivity(ctx, activities.PostSalesOrderGL, params).
        Get(ctx, &glPostingID)
    if err != nil {
        // Compensate step 1
        s.Compensate(ctx)
        return err
    }
    s.AddCompensation(func(ctx workflow.Context) {
        workflow.ExecuteActivity(ctx, activities.ReverseGLPosting, glPostingID)
    })

    // Step 3: Create delivery note
    err = workflow.ExecuteActivity(ctx, activities.CreateDeliveryNote, params).Get(ctx, nil)
    if err != nil {
        // Compensate steps 2 and 1 (in reverse order)
        s.Compensate(ctx)
        return err
    }

    return nil
}
```

### 29.2.2. Forward Steps — Activities That Make Changes

Each forward step:
1. Executes an activity
2. On success: registers a compensation for that step
3. On failure: triggers `s.Compensate(ctx)` to run all registered compensations in reverse

### 29.2.3. Backward Recovery — Running Compensations in Reverse

`s.Compensate(ctx)` runs compensations in LIFO (last-in, first-out) order. If steps 1, 2, 3 completed and step 4 fails: compensations run in order 3, 2, 1. This mirrors the classic "undo" semantic of rollbacks.

### 29.2.4. The `saga` Helper — Awo's Built-In Manager

```go
type Saga struct {
    compensations []CompensationFunc
}

type CompensationFunc func(ctx workflow.Context)

func (s *Saga) AddCompensation(fn CompensationFunc) {
    s.compensations = append(s.compensations, fn)
}

func (s *Saga) Compensate(ctx workflow.Context) {
    // Run in reverse order
    for i := len(s.compensations) - 1; i >= 0; i-- {
        s.compensations[i](ctx)
    }
}
```

---

## 29.3. Designing Compensations

### 29.3.1. Compensations Are Forward-Moving Corrections

A compensation is NOT a database rollback. It is a new forward-moving action that corrects the effect of the failed step. The compensation creates audit trail entries and may trigger notifications.

```go
// Compensation for inventory reservation
func (a *Activities) ReleaseInventoryReservation(ctx context.Context, reservationID uuid.UUID) error {
    // This is NOT a DELETE — it's a status change with an audit entry
    _, err := a.reservationRepo.Update(ctx, reservationID, InventoryReservationUpdate{
        Status:      "released",
        ReleasedAt:  time.Now(),
        ReleasedBy:  "saga-compensation",
        ReleaseNote: "released due to sales order submission failure",
    })
    return err
}
```

### 29.3.2. Idempotency Requirement

Compensations are activities and therefore must be idempotent — they may be retried by Temporal if they fail.

```go
func (a *Activities) ReverseGLPosting(ctx context.Context, postingID uuid.UUID) error {
    existing, err := a.glRepo.Get(ctx, postingID)
    if err != nil {
        return err
    }

    // Idempotency guard — don't reverse twice
    if existing.ReversedAt != nil {
        return nil  // already reversed
    }

    _, err = a.glRepo.Create(ctx, GLEntryCreate{
        Type:         "reversal",
        SourcePostingID: postingID,
        // reversed debits and credits
    })
    if err != nil {
        return err
    }

    return a.glRepo.Update(ctx, postingID, GLEntryUpdate{
        ReversedAt: ptr.Time(time.Now()),
    })
}
```

### 29.3.3. Compensation Failure — When a Compensation Itself Fails

If a compensation activity fails after all retries, the saga is in an inconsistent state: some forward steps were compensated, others were not. Temporal logs this as a workflow failure.

Handling compensation failure:
1. The workflow fails with an error describing which step's compensation failed
2. The failure is logged as a high-priority alert
3. A human operator reviews the workflow history and manually applies the missing compensation
4. The `awo manual-compensation` CLI command assists with guided manual recovery

This is a rare scenario — compensations are designed to be simple, idempotent, and unlikely to fail. But designing for it is important.

### 29.3.4. Storing Compensation State

Always store enough information in the compensation input to execute the compensation independently:

```go
type GLReversalInput struct {
    TenantID    uuid.UUID   // needed to set search_path
    PostingID   uuid.UUID   // which posting to reverse
    WorkflowID  string      // for audit trail
    CompensationReason string
}
```

---

## 29.4. Worked Examples

### 29.4.1. Sales Order Submit Saga

```go
func SalesOrderSubmitSaga(ctx workflow.Context, params SalesOrderParams) error {
    s := saga.New()
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: time.Minute,
        RetryPolicy: &temporal.RetryPolicy{MaxAttempts: 3},
    }
    ctx = workflow.WithActivityOptions(ctx, ao)

    // Step 1: Reserve inventory for each line
    var reservations []InventoryReservation
    err := workflow.ExecuteActivity(ctx, activities.ReserveInventoryBulk, params).
        Get(ctx, &reservations)
    if err != nil {
        return fmt.Errorf("inventory reservation failed: %w", err)
    }
    s.AddCompensation(func(ctx workflow.Context) {
        for _, r := range reservations {
            workflow.ExecuteActivity(ctx, activities.ReleaseInventoryReservation, r.ID)
        }
    })

    // Step 2: Post GL entries (debit COGS, credit inventory)
    var glPosting GLPosting
    err = workflow.ExecuteActivity(ctx, activities.PostSalesOrderCOGS, params).
        Get(ctx, &glPosting)
    if err != nil {
        s.Compensate(ctx)
        return fmt.Errorf("COGS GL posting failed: %w", err)
    }
    s.AddCompensation(func(ctx workflow.Context) {
        workflow.ExecuteActivity(ctx, activities.ReverseGLPosting, glPosting.ID)
    })

    // Step 3: Create delivery note
    var deliveryNoteID uuid.UUID
    err = workflow.ExecuteActivity(ctx, activities.CreateDeliveryNote, params).
        Get(ctx, &deliveryNoteID)
    if err != nil {
        s.Compensate(ctx)
        return fmt.Errorf("delivery note creation failed: %w", err)
    }

    // Step 4: Update sales order status to "confirmed"
    err = workflow.ExecuteActivity(ctx, activities.UpdateSalesOrderStatus,
        params.SalesOrderID, "confirmed").Get(ctx, nil)
    if err != nil {
        // Note: delivery note was created — compensation deletes it
        s.AddCompensation(func(ctx workflow.Context) {
            workflow.ExecuteActivity(ctx, activities.DeleteDeliveryNote, deliveryNoteID)
        })
        s.Compensate(ctx)
        return fmt.Errorf("status update failed: %w", err)
    }

    return nil
}
```

### 29.4.2. Fuel Delivery Reconciliation Saga

```go
func FuelDeliveryReconciliationSaga(ctx workflow.Context, params FuelDeliveryParams) error {
    s := saga.New()

    // Step 1: Validate meter readings
    var validation MeterValidationResult
    err := workflow.ExecuteActivity(ctx, activities.ValidateMeterReadings, params).
        Get(ctx, &validation)
    if err != nil || !validation.Valid {
        return fmt.Errorf("meter validation failed: %s", validation.Reason)
    }

    // Step 2: Compute and post variance entry
    var varianceEntryID uuid.UUID
    err = workflow.ExecuteActivity(ctx, activities.PostFuelVarianceGL, params, validation).
        Get(ctx, &varianceEntryID)
    if err != nil {
        return err
    }
    s.AddCompensation(func(ctx workflow.Context) {
        workflow.ExecuteActivity(ctx, activities.ReverseGLPosting, varianceEntryID)
    })

    // Step 3: Update wetstock dip record
    err = workflow.ExecuteActivity(ctx, activities.UpdateWetStockDip, params, validation).
        Get(ctx, nil)
    if err != nil {
        s.Compensate(ctx)
        return fmt.Errorf("wetstock update failed: %w", err)
    }

    // Step 4: Mark delivery as reconciled
    err = workflow.ExecuteActivity(ctx, activities.MarkDeliveryReconciled, params.DeliveryID).
        Get(ctx, nil)
    if err != nil {
        s.Compensate(ctx)
        return err
    }

    return nil
}
```
