### Chapter 7 — The EntityRecord Lifecycle

Every mutation to an entity — create, update, delete, submit, cancel — passes through a defined sequence of stages. This sequence is the EntityRecord lifecycle. It is not a suggestion or a convention; it is enforced structurally by the framework. Business logic placed in a `before_save` hook executes before the database transaction opens. Side effects placed in an `after_save` hook execute inside the transaction, before commit. Temporal workflow triggers fire after commit. Understanding exactly when each stage executes, what is available in context at each point, and what can be aborted and at what cost is the foundational knowledge for all module development.

This chapter documents every lifecycle stage, every hook type, and the patterns for writing hooks that are correct, testable, and maintainable.

---

#### 7.1. Lifecycle Stages

##### 7.1.1. Stage overview — CREATE → VALIDATE → AUTHORIZE → PERSIST → POST-PROCESS

The five stages execute in strict order for every write operation:

**CREATE** — the framework assembles an `EntityRecord` from the incoming payload, applying type coercion, default values, naming series assignment, and system-managed fields (`created_at`, `created_by`, `id`). No hooks fire in this stage. No database is contacted. The record exists only in memory.

**VALIDATE** — synchronous field validators execute for all fields concurrently, followed by async validators for fields that declared them. Cross-field validators run after per-field validators. If any validator returns an error, the stage fails, all collected field errors are returned to the caller as a 422 response, and the lifecycle halts. No hook has fired, no transaction has opened.

**AUTHORIZE** — the permission system verifies that the caller's role has the required operation permission on this entity (§16.3). Privacy policy mutation rules are evaluated (§9.2.2). If either check fails, the lifecycle halts with a 403 response. The `before_validate` hook, if registered, fires here — after field validation passes and before persistence begins.

**PERSIST** — the database transaction opens. `before_save` hooks execute. The INSERT or UPDATE SQL executes. `after_save` hooks execute inside the open transaction. The transaction commits. If any step in this stage throws an error, the transaction rolls back entirely and the lifecycle halts.

**POST-PROCESS** — after the transaction commits, Temporal workflow starts that were queued during the PERSIST stage are dispatched. Notification events are published to Redis pub/sub. Cache invalidations are triggered. None of the POST-PROCESS work participates in the database transaction; partial failure here does not roll back the persisted record.

##### 7.1.2. Where hooks fire relative to stages

Hook firing positions in the stage sequence:

```
VALIDATE completes
    ↓
before_validate fires (AUTHORIZE stage)
    ↓
AUTHORIZE completes
    ↓
[transaction opens]
before_save fires (PERSIST stage, inside transaction)
    ↓
INSERT / UPDATE executes
    ↓
after_save fires (PERSIST stage, inside transaction)
    ↓
[transaction commits]
    ↓
POST-PROCESS (workflow triggers, cache invalidation)
```

For delete operations, the sequence is: AUTHORIZE → `before_delete` → DELETE SQL → `after_delete` (if registered) → transaction commit → POST-PROCESS. For submit and cancel operations, the sequence follows the write path above, with `on_submit` or `on_cancel` firing in place of the standard `after_save`.

##### 7.1.3. What can be aborted and at which stage

Any hook returning a non-nil error aborts the lifecycle at the point of the hook's execution. The consequences depend on the stage:

A `before_validate` error aborts before the transaction opens. No database resources are consumed. The error is returned to the caller as a 422 (for `ValidationError` or `BusinessRuleError`) or 403 (for `PermissionError`).

A `before_save` error also aborts before the INSERT/UPDATE SQL executes. The transaction has opened but performed no writes. Rolling back costs a single round-trip to PostgreSQL. Return `before_save` errors freely; they are cheap.

An `after_save` error aborts after the INSERT/UPDATE SQL has executed but before the transaction commits. The rollback undoes both the primary record write and any writes performed by the `after_save` hook itself. This is the correct behaviour — the two are intended to be atomic — but it means that if `after_save` errors are common, every occurrence incurs a full write-then-rollback cycle. Design `after_save` hooks to fail rarely; move conditional logic into `before_save` where failures are cheaper.

POST-PROCESS failures (workflow start failures, cache write failures) do not abort or roll back anything. The persisted record remains committed. POST-PROCESS failures must be handled through retry mechanisms, not by assuming the database write will be undone.

##### 7.1.4. Transaction boundaries — what is inside the DB transaction

Exactly the following work is inside the PostgreSQL transaction:

- The primary entity's INSERT or UPDATE.
- Any writes performed by `before_save` hooks via the `hook.Context`'s repository reference.
- Any writes performed by `after_save` hooks via the `hook.Context`'s repository reference.
- The naming series `nextval()` call (for entities with a naming series).
- Any writes to audit log entity records emitted from `after_save`.

Exactly the following work is outside the transaction:

- Field validation (VALIDATE stage).
- Permission and privacy policy evaluation (AUTHORIZE stage).
- `before_validate` hook execution.
- Temporal workflow start calls (dispatched after commit, in POST-PROCESS).
- Redis writes (cache invalidation, pub/sub notification).
- Any external API calls made from any hook.

> **Danger:** Never make an external API call — MPesa, KRA eTIMS, Africa's Talking SMS, or any HTTP request — inside a `before_save` or `after_save` hook. The call executes inside an open PostgreSQL transaction. A slow or hung external endpoint holds the transaction open indefinitely, consuming a connection from the pool and blocking other operations. External calls belong in Temporal activities, triggered after the transaction commits.

---

#### 7.2. The `before_validate` Hook

##### 7.2.1. Purpose — compute derived fields, normalise input before validation

`before_validate` fires after the `EntityRecord` is assembled from the payload but before field validators run. Its purpose is to transform or normalise the input into a state that is ready for validation. Use it to: compute a derived field from raw inputs (calculate `line_total` from `quantity × unit_price` so the required check on `line_total` passes), normalise a string to a canonical form (uppercase a vehicle registration so validation regex applies correctly), or conditionally set a field based on the operation type (set `status` to `pending_approval` if the tenant's config requires approval for new records).

```go
// Example: before_validate hook computing derived fields before validation
func computeLineTotals(ctx hook.Context) error {
    record := ctx.Record()
    qty, _ := record.Get("quantity").(float64)
    price, ok := record.Get("unit_price").(decimal.Decimal)
    if !ok || qty == 0 {
        return nil // let required validators report the missing fields
    }
    lineTotal := price.Mul(decimal.NewFromFloat(qty))
    record.Set("line_total", lineTotal)
    return nil
}
```

##### 7.2.2. What is available in context at this point

At `before_validate` time, the `hook.Context` carries:

- `ctx.Record()` — the assembled `EntityRecord` with raw input values, defaults applied, and naming series assigned. Field values may not yet be valid; `before_validate` runs before validators.
- `ctx.Operation()` — `hook.OpCreate`, `hook.OpUpdate`, or `hook.OpDelete`, indicating which operation triggered the lifecycle.
- `ctx.PreviousRecord()` — on update operations, the record's state before the update was applied. On create operations, this returns nil.
- `ctx.TenantContext()` — the resolved `TenantContext` for the current request.
- `ctx.UserContext()` — the authenticated user's identity and resolved permissions.
- `ctx.Repository()` — the `EntityRepository` for the current entity, usable for existence checks or related record reads. Database calls here are outside the transaction (the transaction has not opened yet).

The `ctx.Record()` value is mutable at `before_validate` time. Calling `record.Set("field", value)` updates the in-memory record; the new value will be seen by validators and by subsequent hooks.

##### 7.2.3. Aborting from `before_validate` — validation errors vs system errors

Returning a `hook.ValidationError` from `before_validate` causes the framework to include the error in the field-level error collection and return a 422 response, consistent with field validator failures. This is the correct return type for business rule checks that belong logically in the validation phase: "if the entity type is `fleet`, then `fleet_account_id` is required".

Returning a bare `error` (not a `hook.ValidationError`) causes the framework to treat the failure as an internal error and return a 500 response. Only use bare errors for genuine system failures — a repository read that fails because of a database connection error, not a business rule that is expected to fire in normal operations. Never use bare errors to communicate business rule violations to the user.

---

#### 7.3. The `before_save` Hook

##### 7.3.1. Purpose — enforce business rules that require the full validated record

`before_save` fires after all field validators have passed and the transaction has opened, but before the INSERT or UPDATE SQL executes. At this point, the `EntityRecord` is fully validated and all fields are their declared types. Use `before_save` to enforce business rules that require: the full validated record (not just individual fields), a database read to check a constraint the DB cannot enforce, or cross-entity consistency (a journal entry's debits must equal its credits — this requires summing line items that are part of the same record assembly).

```go
// Example: before_save hook enforcing double-entry balance
func enforceDoubleEntry(ctx hook.Context) error {
    tc, err := tenant.FromContext(ctx)
    if err != nil {
        return err
    }
    record := ctx.Record()
    journalID := record.ID()
    if ctx.Operation() == hook.OpCreate {
        // Lines are submitted in the same payload; read from the assembled record
        lines := record.GetRelated("lines") // []entity.EntityRecord
        var totalDebit, totalCredit decimal.Decimal
        for _, line := range lines {
            debit, _ := line.Get("debit_amount").(decimal.Decimal)
            credit, _ := line.Get("credit_amount").(decimal.Decimal)
            totalDebit = totalDebit.Add(debit)
            totalCredit = totalCredit.Add(credit)
        }
        if !totalDebit.Equal(totalCredit) {
            return hook.NewBusinessRuleError("unbalanced_entry",
                fmt.Sprintf("journal entry debits (%s) must equal credits (%s)",
                    totalDebit.String(), totalCredit.String()))
        }
        return nil
    }
    // On update: read persisted lines from the repository
    lineRepo, err := entity.Resolve(ctx, tc, "JournalEntryLine")
    if err != nil {
        return err
    }
    filter := entity.NewFilter().Eq("journal_entry_id", journalID)
    lines, _, err := lineRepo.Query(ctx, filter)
    if err != nil {
        return fmt.Errorf("loading journal lines: %w", err)
    }
    var totalDebit, totalCredit decimal.Decimal
    for _, line := range lines {
        debit, _ := line.Get("debit_amount").(decimal.Decimal)
        credit, _ := line.Get("credit_amount").(decimal.Decimal)
        totalDebit = totalDebit.Add(debit)
        totalCredit = totalCredit.Add(credit)
    }
    if !totalDebit.Equal(totalCredit) {
        return hook.NewBusinessRuleError("unbalanced_entry",
            "journal entry is out of balance")
    }
    return nil
}
```

##### 7.3.2. Accessing the previous version of the record on update

On update operations, `ctx.PreviousRecord()` returns the record's state as it existed in the database before the update was applied. This is useful for: detecting which fields changed (compare `ctx.Record()` against `ctx.PreviousRecord()`), enforcing state machine transitions (reject a status change from `posted` back to `draft`), and computing delta values (the difference in `total_amount` between old and new versions to post a correcting GL entry).

```go
// Example: Enforcing a state machine transition in before_save
func enforceStatusTransitions(ctx hook.Context) error {
    if ctx.Operation() != hook.OpUpdate {
        return nil
    }
    prev := ctx.PreviousRecord()
    if prev == nil {
        return nil
    }
    oldStatus, _ := prev.Get("status").(string)
    newStatus, _ := ctx.Record().Get("status").(string)

    allowed := map[string][]string{
        "draft":            {"submitted", "cancelled"},
        "submitted":        {"approved", "rejected"},
        "approved":         {"posted"},
        "rejected":         {"draft"},
        "posted":           {}, // terminal state
        "cancelled":        {}, // terminal state
    }
    transitions, ok := allowed[oldStatus]
    if !ok {
        return hook.NewBusinessRuleError("invalid_status",
            "unknown current status: "+oldStatus)
    }
    for _, t := range transitions {
        if t == newStatus {
            return nil
        }
    }
    return hook.NewBusinessRuleError("invalid_transition",
        fmt.Sprintf("cannot transition from %s to %s", oldStatus, newStatus))
}
```

##### 7.3.3. Triggering synchronous side effects that must be transactional

`before_save` is the correct place for synchronous side effects that must be part of the same transaction as the primary write. Updating a parent entity's running total when a child line is saved, creating a compensating record when a value changes, reserving an inventory slot before a stock move is confirmed — these are all `before_save` patterns.

```go
// Example: Updating parent total in before_save (transactional)
func updateInvoiceTotal(ctx hook.Context) error {
    tc, err := tenant.FromContext(ctx)
    if err != nil {
        return err
    }
    record := ctx.Record()
    invoiceID, _ := record.Get("sales_invoice_id").(string)
    if invoiceID == "" {
        return nil
    }
    lineRepo, err := entity.Resolve(ctx, tc, "SalesInvoiceLine")
    if err != nil {
        return err
    }
    filter := entity.NewFilter().Eq("sales_invoice_id", invoiceID)
    result, err := lineRepo.Aggregate(ctx, entity.AggregateSpec{
        Field: "line_total",
        Fn:    entity.AggSum,
        Filter: filter,
    })
    if err != nil {
        return fmt.Errorf("aggregating line totals: %w", err)
    }
    invoiceRepo, err := entity.Resolve(ctx, tc, "SalesInvoice")
    if err != nil {
        return err
    }
    _, err = invoiceRepo.Update(ctx, invoiceID, map[string]any{
        "total_amount": result.Value,
    })
    return err
}
```

---

#### 7.4. The `after_save` Hook

##### 7.4.1. Purpose — post-persist side effects, cache invalidation, event emission

`after_save` fires after the INSERT or UPDATE SQL has executed successfully but before the transaction commits. The primary record now exists in the database (though not yet visible to other transactions in the default `READ COMMITTED` isolation level until commit). Use `after_save` to: write audit log entries, insert related records that must be co-created with the primary (a `LedgerEntry` header that must exist before its lines are created in the same transaction), update a denormalised summary field on a parent, or emit an event to the transactional outbox.

```go
// Example: after_save hook writing an audit log entry
func writeAuditLog(ctx hook.Context) error {
    tc, err := tenant.FromContext(ctx)
    if err != nil {
        return err
    }
    record := ctx.Record()
    auditRepo, err := entity.Resolve(ctx, tc, "AuditLog")
    if err != nil {
        return err
    }
    prev := ctx.PreviousRecord()
    var prevData map[string]any
    if prev != nil {
        prevData = prev.ToMap()
    }
    _, err = auditRepo.Create(ctx, map[string]any{
        "entity_name":   record.EntityName(),
        "record_id":     record.ID(),
        "operation":     ctx.Operation().String(),
        "previous_data": prevData,
        "new_data":      record.ToMap(),
        "performed_by":  ctx.UserContext().UserID(),
        "performed_at":  entity.Now(),
    })
    return err
}
```

##### 7.4.2. Still inside the transaction — implications

The most important property of `after_save` is that it executes inside the open database transaction. Any write performed here is rolled back if the transaction fails. This makes `after_save` the correct place for writes that must be atomically consistent with the primary record.

It also means that `after_save` must complete quickly. Long-running operations inside an open transaction hold a database connection and lock rows. If `after_save` performs a slow operation, every concurrent request that touches the same rows — including reads in `READ COMMITTED` isolation that may need the same connection pool slot — is delayed. The target execution time for `after_save` hooks is under 50 milliseconds for typical ERP workloads.

##### 7.4.3. Triggering Temporal workflows from `after_save`

`after_save` is the correct hook from which to trigger Temporal workflows. The record is persisted at this point; the workflow's first activity can safely call `repo.Get(ctx, recordID)` and find the record. The workflow start call is queued by the framework and dispatched after the transaction commits (POST-PROCESS stage); it does not block inside the transaction.

The framework's `WorkflowTrigger` binding (§2.4.6) handles this dispatch automatically when declared in the `EntityDefinition`. For programmatic workflow starts from within a hook — when the trigger condition is dynamic rather than declarative — use `ctx.TriggerWorkflow(WorkflowFunc, input, options)`:

```go
// Example: Programmatically triggering a workflow from after_save
func triggerApprovalWorkflow(ctx hook.Context) error {
    record := ctx.Record()
    status, _ := record.Get("status").(string)
    if status != "submitted" {
        return nil
    }
    tc, _ := tenant.FromContext(ctx)
    return ctx.TriggerWorkflow(
        ApprovalWorkflow,
        ApprovalWorkflowInput{
            TenantSlug: tc.Slug(),
            RecordID:   record.ID(),
            EntityName: record.EntityName(),
        },
        entity.WorkflowOptions{
            TaskQueue:  "fieldservice.approval",
            WorkflowID: fmt.Sprintf("%s.%s.%s.approval", tc.Slug(), record.EntityName(), record.ID()),
        },
    )
}
```

##### 7.4.4. Avoiding slow operations inside `after_save`

The three most common mistakes that introduce slow operations into `after_save` are: making HTTP calls to external services, performing aggregate queries over large datasets without appropriate indexes, and loading large sets of related records without limiting their scope. Each of these must be moved to Temporal activities.

The test for whether an operation belongs in `after_save` or in a workflow activity is: can this operation fail without making the primary record persist fail? If yes, it belongs in a workflow activity. Sending an SMS notification when a service request is created can fail without preventing the service request from being saved; it belongs in an activity. Updating a running balance total that other processes read for consistency cannot fail silently; it belongs in `after_save`.

---

#### 7.5. The `before_delete` Hook

##### 7.5.1. Purpose — guard deletion, check referential integrity that the DB cannot

`before_delete` fires before the DELETE SQL executes, inside the transaction. Its primary purpose is to enforce deletion guards that the database cannot express as FK constraints: checking whether a submitted invoice has any associated GL entries that must be reversed before deletion, whether a warehouse has non-zero stock that must be zeroed out first, or whether an employee has open payroll runs that reference them.

```go
// Example: before_delete hook blocking deletion of a posted entity
func blockDeletionIfPosted(ctx hook.Context) error {
    record := ctx.Record()
    status, _ := record.Get("status").(string)
    if status == "posted" || status == "submitted" {
        return hook.NewBusinessRuleError("cannot_delete_posted",
            fmt.Sprintf("cannot delete a %s record with status %q; cancel it first",
                record.EntityName(), status))
    }
    return nil
}
```

```go
// Example: before_delete hook checking for dependent records
func blockDeletionIfHasStockMoves(ctx hook.Context) error {
    tc, err := tenant.FromContext(ctx)
    if err != nil {
        return err
    }
    record := ctx.Record()
    moveRepo, err := entity.Resolve(ctx, tc, "StockMove")
    if err != nil {
        return err
    }
    filter := entity.NewFilter().Eq("warehouse_id", record.ID())
    exists, err := moveRepo.Exists(ctx, filter)
    if err != nil {
        return fmt.Errorf("checking stock moves: %w", err)
    }
    if exists {
        return hook.NewBusinessRuleError("has_stock_moves",
            "cannot delete warehouse: stock movement records exist")
    }
    return nil
}
```

##### 7.5.2. Soft delete pattern — marking `deleted_at` instead of hard delete

For entities where audit trail completeness is important — all financial documents, all inventory records, all HR records — use soft delete rather than hard delete. The entity declares a `deleted_at` field of type `DateTime` (nullable, no default), and the `before_delete` hook converts the deletion into an update that sets `deleted_at` to the current timestamp.

```go
// Example: before_delete hook implementing soft delete
func softDelete(ctx hook.Context) error {
    tc, err := tenant.FromContext(ctx)
    if err != nil {
        return err
    }
    record := ctx.Record()
    repo, err := entity.Resolve(ctx, tc, record.EntityName())
    if err != nil {
        return err
    }
    _, err = repo.Update(ctx, record.ID(), map[string]any{
        "deleted_at": entity.Now(),
    })
    if err != nil {
        return err
    }
    // Abort the actual DELETE by returning a sentinel error that the
    // framework recognises as "converted to update, halt delete pipeline".
    return hook.ErrSoftDeleted
}
```

When `hook.ErrSoftDeleted` is returned, the framework halts the delete pipeline, rolls back the partially opened delete transaction, and returns a 200 response with the updated record (reflecting the `deleted_at` value) rather than a 204 No Content. All subsequent `Query` operations on the entity automatically filter out records where `deleted_at IS NOT NULL` unless the caller explicitly passes `entity.IncludeDeleted()`.

##### 7.5.3. Returning user-facing errors from `before_delete`

All errors returned from `before_delete` should be `hook.NewBusinessRuleError(code, message)` for business rule violations, or `hook.NewValidationError(field, message)` if the deletion is blocked by a field-level condition. Returning a bare `error` from `before_delete` causes a 500 response, which is never the right outcome for a business rule check. The error code and message are serialised into the standard error envelope (§18) and displayed by the amis UI in a toast notification.

---

#### 7.6. The `on_submit` and `on_cancel` Hooks

##### 7.6.1. What submission means in ERP context — document finalisation

Submission is the act of finalising a document: a `SalesInvoice` is submitted when it is issued to the customer, a `JournalEntry` is submitted when it is posted to the general ledger, a `PurchaseOrder` is submitted when it is sent to the supplier. Before submission, a document is a draft — editable, reversible, uncommitted. After submission, the document has legal, financial, or operational significance that makes unilateral editing inappropriate.

The framework implements submission as a two-step operation: the document's `status` field is updated to `submitted`, and then the `on_submit` hook fires. The `on_submit` hook is the correct place to: trigger the GL posting workflow, reserve inventory, send the document to an external system (KRA eTIMS for tax invoices), or lock immutable fields by adding a permission check that prevents updates after submission.

##### 7.6.2. Immutability after submission — which fields lock and which do not

After submission, the framework can enforce field-level immutability based on the document's status. Declare post-submission immutability using `entity.ImmutableAfterSubmit()` on any field that should not change once the document is finalised:

```go
// Example: Fields locked after submission
entity.Field("posting_date").
    Type(entity.Date).
    Required().
    ImmutableAfterSubmit()

entity.Field("total_amount").
    Type(entity.Currency).
    Required().
    ImmutableAfterSubmit()
```

Fields not declared `ImmutableAfterSubmit` remain editable after submission. This is correct for fields like `payment_status` (which is updated by the payment processing workflow), `eTIMS_submission_status` (updated by the KRA integration), and internal notes that do not affect the financial record.

##### 7.6.3. `on_cancel` as a compensating action — reversing GL postings, stock moves

`on_cancel` fires when a submitted document is cancelled. Cancellation is not the same as deletion: a cancelled document remains in the database with `status = "cancelled"` and is visible in audit reports. The `on_cancel` hook's responsibility is to reverse the effects of submission: if `on_submit` posted GL entries, `on_cancel` must post reversal entries; if `on_submit` reserved inventory, `on_cancel` must release it.

```go
// Example: on_cancel hook triggering a reversal workflow
func triggerInvoiceReversalWorkflow(ctx hook.Context) error {
    tc, err := tenant.FromContext(ctx)
    if err != nil {
        return err
    }
    record := ctx.Record()
    return ctx.TriggerWorkflow(
        SalesInvoiceReversalWorkflow,
        SalesInvoiceReversalInput{
            TenantSlug: tc.Slug(),
            InvoiceID:  record.ID(),
        },
        entity.WorkflowOptions{
            TaskQueue:  "finance.sales-invoice.cancel",
            WorkflowID: fmt.Sprintf("%s.sales-invoice.%s.cancel", tc.Slug(), record.ID()),
        },
    )
}
```

The reversal workflow is a saga (§29) that reverses the submission saga's steps in the reverse order. The `on_cancel` hook does not perform the reversal directly; it triggers the workflow that does. This keeps the cancellation hook fast and the reversal logic durable and observable.

##### 7.6.4. Amendment workflow — creating a new version of a submitted document

Some documents cannot be cancelled once issued (a tax invoice that has been submitted to KRA eTIMS, for example) but may need corrections. The amendment pattern creates a new version of the document linked to the original: the original remains submitted and unchanged; the amendment is a new record with an amended naming series value (e.g., `INV-2025-06-0001-A1`) that supersedes the original.

The framework does not provide built-in amendment logic because amendment rules are highly jurisdiction and document-type specific. Module developers implement amendment as a custom action on the entity (§4.5.2) that: validates the original is in a state that allows amendment, creates a copy of the original with amended field values, links the new record to the original via a `Link` field `amends_id`, and potentially triggers a workflow that notifies external systems of the amendment.

---

#### 7.7. Writing Testable Hooks

##### 7.7.1. The interceptor pattern — hooks as dependencies, not global registrations

Hooks become untestable when they depend on global state. A `before_save` hook that calls `globalDB.Query(...)` or accesses a package-level variable cannot be tested in isolation. The interceptor pattern solves this by declaring hooks as methods on a struct that holds its dependencies as interface-typed fields.

```go
// Example: Hooks as methods on a dependency-injected struct
type ServiceRequestHooks struct {
    Repo    entity.EntityRepository
    Notifier Notifier // interface, not a concrete type
}

func (h *ServiceRequestHooks) BeforeSave(ctx hook.Context) error {
    record := ctx.Record()
    status, _ := record.Get("status").(string)
    if status == "completed" && record.Get("completed_at") == nil {
        record.Set("completed_at", entity.Now())
    }
    return nil
}

func (h *ServiceRequestHooks) AfterSave(ctx hook.Context) error {
    record := ctx.Record()
    if ctx.Operation() != hook.OpCreate {
        return nil
    }
    return h.Notifier.NotifyNewRequest(ctx, record.ID())
}
```

Register the hooks from the module's `Register` function where the dependencies are wired:

```go
// Example: Registering hooks from module Register with injected dependencies
func (m *Module) Register(app entity.App) error {
    hooks := &ServiceRequestHooks{
        Repo:     app.Repository("ServiceRequest"),
        Notifier: m.notifier,
    }
    ServiceRequestDefinition = ServiceRequestDefinition.WithHooks(
        hook.BeforeSave(hooks.BeforeSave),
        hook.AfterSave(hooks.AfterSave),
    )
    return app.RegisterEntity(ServiceRequestDefinition)
}
```

##### 7.7.2. Unit testing hooks in isolation without a live database

With the interceptor pattern, hooks are testable without a running database. Use a mock `EntityRepository` that satisfies the interface and returns controlled results:

```go
// Example: Unit test for BeforeSave hook without a database
func TestBeforeSaveSetCompletedAt(t *testing.T) {
    hooks := &ServiceRequestHooks{
        Repo:     entity.NewMockRepository(),
        Notifier: &mockNotifier{},
    }

    record := entity.NewTestRecord("ServiceRequest",
        map[string]any{
            "status":       "completed",
            "completed_at": nil,
        },
    )
    ctx := hook.NewTestContext(
        hook.WithRecord(record),
        hook.WithOperation(hook.OpUpdate),
    )

    err := hooks.BeforeSave(ctx)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    completedAt := record.Get("completed_at")
    if completedAt == nil {
        t.Error("expected completed_at to be set when status is completed")
    }
}
```

`entity.NewMockRepository()`, `entity.NewTestRecord()`, and `hook.NewTestContext()` are provided by the `testkit` package at `github.com/awolabs/awo/testkit`. The mock repository can be pre-loaded with fixture records using `mockRepo.Seed(records...)` and configured to return specific errors using `mockRepo.OnGet(id).Return(nil, someError)`.

##### 7.7.3. Integration testing hooks with an in-memory ent client

For hooks that perform complex multi-entity database operations that are difficult to mock convincingly, use the `testkit.NewIntegrationDB()` helper which creates a real PostgreSQL schema seeded with the test entity definitions and returns a live `EntityRepository` scoped to the test schema. The schema is created in a transaction and rolled back at the end of each test, so tests are isolated.

```go
// Example: Integration test using a real test database
func TestEnforceDoubleEntry(t *testing.T) {
    db := testkit.NewIntegrationDB(t,
        JournalEntryDefinition,
        JournalEntryLineDefinition,
    )
    tc := testkit.NewTestTenantContext(t, db)
    ctx := testkit.NewTestContext(t, tc)

    repo, err := entity.Resolve(ctx, tc, "JournalEntry")
    if err != nil {
        t.Fatal(err)
    }

    // Attempt to create an unbalanced journal entry
    _, err = repo.Create(ctx, map[string]any{
        "entry_date":   "2025-06-01",
        "description":  "Test entry",
        "lines": []map[string]any{
            {"account_id": "acct-001", "debit_amount": "1000.00", "credit_amount": "0"},
            {"account_id": "acct-002", "debit_amount": "0", "credit_amount":  "900.00"},
        },
    })
    if err == nil {
        t.Fatal("expected error for unbalanced entry")
    }
    var bre *hook.BusinessRuleError
    if !errors.As(err, &bre) || bre.Code != "unbalanced_entry" {
        t.Errorf("expected unbalanced_entry BusinessRuleError, got: %v", err)
    }
}
```

---

#### 7.8. Chaining Multiple Hooks

##### 7.8.1. Hook execution order when multiple hooks are registered on one entity

Hooks of the same type execute in the order they are declared in the `EntityDefinition`. If three `before_save` hooks are registered — `validateBalance`, `enforceStatusTransitions`, and `updateParentTotal` — they execute in that declaration order, not concurrently. This ordering is guaranteed and must be relied upon when hooks share assumptions: `updateParentTotal` may assume that `validateBalance` has already passed.

```go
// Example: Multiple hooks registered in explicit execution order
var JournalEntryDefinition = entity.Define("JournalEntry",
    entity.Fields( /* ... */ ),
    entity.Hooks(
        hook.BeforeSave(validateBalance),           // runs first
        hook.BeforeSave(enforceStatusTransitions),  // runs second
        hook.AfterSave(writeAuditLog),              // after save, first
        hook.AfterSave(emitOutboxEvent),            // after save, second
        hook.OnSubmit(triggerGLPostingWorkflow),
    ),
)
```

##### 7.8.2. Early exit — stopping the chain without an error

To stop the hook chain without returning an error — for example, a hook that should only run for create operations and wants to be a no-op for updates — return `hook.ErrSkip`. This sentinel value causes the framework to halt the chain for the current hook type and proceed to the next lifecycle stage without treating the skip as an error.

```go
// Example: Using hook.ErrSkip for conditional no-op
func onlyOnCreate(ctx hook.Context) error {
    if ctx.Operation() != hook.OpCreate {
        return hook.ErrSkip
    }
    // ... create-specific logic
    return nil
}
```

`hook.ErrSkip` skips only the remaining hooks of the same type. It does not abort the entire lifecycle. A `before_save` hook that returns `hook.ErrSkip` stops the remaining `before_save` hooks but does not prevent the INSERT from executing.

##### 7.8.3. Sharing context between chained hooks

Chained hooks sometimes need to share computed state. A `before_validate` hook that fetches a parent record to compute derived fields may want to make that parent record available to the subsequent `before_save` hook without fetching it again. Use `ctx.Set(key, value)` and `ctx.Get(key)` to pass values through the hook context:

```go
// Example: Sharing state between hooks via context values
func fetchParentInBeforeValidate(ctx hook.Context) error {
    tc, err := tenant.FromContext(ctx)
    if err != nil {
        return err
    }
    record := ctx.Record()
    parentID, _ := record.Get("parent_id").(string)
    if parentID == "" {
        return nil
    }
    parentRepo, err := entity.Resolve(ctx, tc, "ParentEntity")
    if err != nil {
        return err
    }
    parent, err := parentRepo.Get(ctx, parentID)
    if err != nil {
        return fmt.Errorf("loading parent: %w", err)
    }
    ctx.Set("parent_record", parent) // share for subsequent hooks
    return nil
}

func useParentInBeforeSave(ctx hook.Context) error {
    parent, ok := ctx.Get("parent_record").(entity.EntityRecord)
    if !ok {
        return nil // parent was not loaded (optional parent)
    }
    parentStatus, _ := parent.Get("status").(string)
    if parentStatus == "closed" {
        return hook.NewBusinessRuleError("parent_closed",
            "cannot add records to a closed parent")
    }
    return nil
}
```

Values set via `ctx.Set` are scoped to the current request's hook chain. They are not persisted, not visible to other requests, and not accessible after the lifecycle completes. Use this mechanism for intra-hook communication only; do not use it to pass values to Temporal workflows (pass via the workflow input struct instead).

---

#### Chapter summary

Chapter 7 defines the five lifecycle stages (§7.1.1) and the transaction boundaries that determine what is and is not atomic (§7.1.4). It documents each hook type — `before_validate` for input normalisation (§7.2), `before_save` for business rule enforcement inside the transaction (§7.3), `after_save` for transactional side effects and workflow triggers (§7.4), `before_delete` for deletion guards and soft delete (§7.5), and `on_submit`/`on_cancel` for document finalisation and reversal (§7.6). The three most critical concepts are the transaction boundary (§7.1.4, which defines what rolls back together and what does not), the prohibition on external I/O inside `before_save` and `after_save` (§7.1.4 Danger callout and §7.4.4), and the interceptor pattern for testable hooks (§7.7.1, which is the required approach for all production hook code).

**Next chapters to read:**

- §8 — The Persistence Interface (the `EntityRepository` methods called throughout this chapter — `Get`, `Query`, `Create`, `Update`, `Exists`, `Aggregate` — are fully specified with their signatures, filter DSL, and error types)
- §27 — Defining Workflows (the Temporal workflows triggered from `on_submit` and `after_save` hooks: understanding workflow inputs, activity options, and retry policies is required before wiring hooks to workflows in production)
- §29 — Saga Pattern — Compensating Transactions (the reversal workflows triggered by `on_cancel` hooks follow the saga pattern: this chapter documents how to implement compensating transactions correctly for multi-step submissions)
