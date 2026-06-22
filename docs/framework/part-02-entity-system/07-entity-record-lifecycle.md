---
title: "Chapter 7: The EntityRecord Lifecycle"
part: "Part II — The EntityDefinition System"
chapter: 7
section: "07-entity-record-lifecycle"
related:
  - "[Chapter 5: Field System](05-field-system.md)"
  - "[Chapter 8: The Persistence Interface](08-persistence-interface.md)"
  - "[Chapter 26: Temporal Fundamentals](../part-05-workflow/26-temporal-fundamentals.md)"
---

# Chapter 7: The EntityRecord Lifecycle

Every mutation in Awo — whether a new invoice being created, a purchase order being submitted, or a vendor record being updated — passes through a structured lifecycle. The lifecycle is not optional middleware: it is the primary extension point for ERP business logic, the gatekeeper for data integrity, and the integration point between synchronous persistence and asynchronous workflow.

Understanding the lifecycle deeply is what separates an Awo application that works from one that works correctly under all conditions.

---

## 7.1. Lifecycle Stages

### 7.1.1. Stage Overview — CREATE → VALIDATE → AUTHORIZE → PERSIST → POST-PROCESS

Every mutation follows this sequence:

```
HTTP Request
    │
    ▼
┌──────────────────────┐
│   1. ASSEMBLE        │  Parse payload → EntityRecord (unvalidated)
└──────────┬───────────┘
           │
    ▼
┌──────────────────────┐
│   2. before_validate │  Normalise, compute derived fields
└──────────┬───────────┘
           │
    ▼
┌──────────────────────┐
│   3. VALIDATE        │  Field types, constraints, cross-field, async
└──────────┬───────────┘
           │
    ▼
┌──────────────────────┐
│   4. AUTHORIZE       │  RBAC check + privacy policy check
└──────────┬───────────┘
           │
    ▼
┌──────────────────────┐
│   5. before_save     │  Business rules, transactional side effects
└──────────┬───────────┘
           │ ← DB transaction begins here
    ▼
┌──────────────────────┐
│   6. PERSIST         │  INSERT or UPDATE in the database
└──────────┬───────────┘
           │
    ▼
┌──────────────────────┐
│   7. after_save      │  Post-persist side effects, cache invalidation,
│                      │  event emission, workflow signal
└──────────┬───────────┘
           │ ← DB transaction commits here
    ▼
┌──────────────────────┐
│   8. POST-PROCESS    │  Temporal workflow started (if configured)
└──────────────────────┘
           │
    ▼
HTTP Response
```

### 7.1.2. Where Hooks Fire Relative to Stages

| Hook | Stage | Inside TX? | Abort supported? |
|---|---|---|---|
| `before_validate` | 2 | No | Yes — ValidationError |
| `before_save` | 5 | No (TX starts after) | Yes — BusinessError |
| `after_save` | 7 | **Yes** | Yes — causes rollback |
| `on_submit` | 5 (after status change) | Yes | Yes — BusinessError |
| `on_cancel` | 5 (after status change) | Yes | Yes — BusinessError |
| `before_delete` | 5 (for deletes) | No | Yes — BusinessError |

The key point: `after_save` runs **inside the database transaction**. This means an error in `after_save` rolls back the entire insert or update. Use this guarantee to ensure atomicity of side effects that must be consistent with the main record — but be aware that slow operations inside `after_save` hold the transaction open and block other readers/writers on the same rows.

### 7.1.3. What Can Be Aborted and at Which Stage

| Stage | How to abort | HTTP status emitted |
|---|---|---|
| before_validate | Return `validate.Error` | 422 Unprocessable Entity |
| VALIDATE | Automatic on constraint failure | 422 |
| AUTHORIZE | Return `ErrForbidden` | 403 Forbidden |
| before_save | Return `*errs.BusinessError` | 422 or 409 |
| PERSIST | DB constraint violation (caught) | 409 Conflict or 422 |
| after_save | Any error — causes rollback | 500 Internal (or mapped) |

Never panic inside a hook. Panics propagate up through Fiber's recovery middleware, lose all transaction state, and produce a 500 without rollback confirmation. Always return an error.

### 7.1.4. Transaction Boundaries — What Is Inside the DB Transaction

The transaction scope in Awo is deliberately narrow:

```
before_save (pre-TX) → [TX OPEN] → PERSIST → after_save → [TX COMMIT]
```

**Inside the transaction:**
- The main entity insert/update/delete
- Any nested entity mutations triggered by `after_save`
- Any repository operations performed via the transactional repository reference passed to hooks
- Naming series counter increments

**Outside the transaction:**
- `before_validate` (pre-TX; no DB write should happen here)
- `before_save` (pre-TX by default; can explicitly receive a TX reference if configured)
- Temporal workflow starts (post-commit; started after the TX commits)
- Redis cache writes
- Email/SMS notifications

This narrowness is intentional. Holding a DB transaction open for network calls (external APIs, Redis, email) dramatically increases connection pressure and deadlock probability under load. External side effects belong in Temporal workflows, not hooks.

---

## 7.2. The `before_validate` Hook

### 7.2.1. Purpose — Compute Derived Fields, Normalise Input Before Validation

`before_validate` runs after the payload is parsed into an `EntityRecord` but before any field validation occurs. Its purpose is to compute values that validation depends on, or to normalise input so that validation can be deterministic.

**Canonical ERP uses:**

- Normalise phone numbers to E.164 format before phone validator runs
- Look up the currency exchange rate for the invoice date and set `amount_ksh` based on `amount_usd`
- Compute the VAT-inclusive total from line items so the `total_amount` field validates correctly
- Parse a document number into its component parts

```go
func (Invoice) BeforeValidate(ctx context.Context, r *schema.EntityRecord) error {
    // Normalise phone — ensures phone validator receives canonical form
    if phone := r.GetString("contact_phone"); phone != "" {
        normalised, err := phonenumber.Normalise(phone, "KE")
        if err != nil {
            // Do NOT validate here — return a validate.Error so it appears as a field error
            return validate.FieldErrorf("contact_phone",
                "could not parse phone number: %v", err)
        }
        r.Set("contact_phone", normalised)
    }

    // Compute totals so cross-field validators can read them
    if err := recomputeInvoiceTotals(r); err != nil {
        return err
    }

    return nil
}
```

### 7.2.2. What Is Available in Context at This Point

At `before_validate` time, the context carries:
- Tenant context: `tenant.FromContext(ctx)` — fully resolved
- Actor context: `actor.FromContext(ctx)` — authenticated user/token
- The `EntityRecord` in its raw parsed state — fields set from the HTTP payload, defaults applied, no validation yet

The context does **not** carry a database transaction. You may perform read-only DB queries in `before_validate` to look up reference data (exchange rates, tax rates, etc.), but do not open transactions.

### 7.2.3. Aborting from `before_validate` — Validation Errors vs System Errors

Return a `validate.FieldError` for user-facing validation failures. These appear in the `errors[]` array in the 422 response and bind to the correct amis form field.

Return a plain `error` for unexpected system failures (DB errors, network errors). These propagate as 500 Internal Server Error and are logged with full context but do not expose internal details to the caller.

```go
// User error — surfaces as { "field": "vat_number", "message": "..." }
return validate.FieldErrorf("vat_number", "VAT registration not found in KRA registry")

// System error — logged, surfaces as 500
return fmt.Errorf("KRA API unavailable: %w", err)
```

---

## 7.3. The `before_save` Hook

### 7.3.1. Purpose — Enforce Business Rules That Require the Full Validated Record

`before_save` runs after the full record is validated and authorised, but before the DB transaction opens. It is the primary place for business-rule enforcement that requires the full validated state:

```go
func (Invoice) BeforeSave(ctx context.Context, r *schema.EntityRecord, op schema.MutationOp) error {
    if op == schema.OpCreate {
        // Enforce credit limit
        customer, err := customerRepo.Get(ctx, r.GetUUID("customer_id"))
        if err != nil {
            return err
        }
        outstanding, err := invoiceRepo.Aggregate(ctx,
            filter.Eq("customer_id", customer.ID).
            And(filter.In("status", []string{"submitted", "overdue"})),
            aggregate.Sum("total_amount"),
        )
        if err != nil {
            return err
        }
        creditLimit := customer.GetDecimal("credit_limit_ksh")
        newTotal := r.GetDecimal("total_amount")
        if outstanding.Add(newTotal).GreaterThan(creditLimit) {
            return errs.NewBusinessError("CREDIT_LIMIT_EXCEEDED",
                "invoice total would exceed customer credit limit of KES %s",
                creditLimit.StringFixed(2),
            )
        }
    }
    return nil
}
```

**Common before_save patterns in ERP:**
- Credit limit enforcement
- Stock availability check before a sales order
- Approval workflow guard (ensure required approvals exist before document can be submitted)
- Budget check before a purchase requisition
- Period lock check (reject entries to a closed accounting period)

### 7.3.2. Accessing the Previous Version of the Record on Update

For updates, `before_save` receives both the new values and access to the current persisted record:

```go
func (Invoice) BeforeSave(ctx context.Context, r *schema.EntityRecord, op schema.MutationOp) error {
    if op == schema.OpUpdate {
        // Fetch the current state from DB
        current, err := invoiceRepo.Get(ctx, r.GetUUID("id"))
        if err != nil {
            return err
        }

        // Reject status regressions
        currentStatus := current.GetString("status")
        newStatus := r.GetString("status")
        if !isValidStatusTransition(currentStatus, newStatus) {
            return errs.NewBusinessError("INVALID_STATUS_TRANSITION",
                "cannot transition from %q to %q", currentStatus, newStatus)
        }
    }
    return nil
}

func isValidStatusTransition(from, to string) bool {
    valid := map[string][]string{
        "draft":     {"submitted", "cancelled"},
        "submitted": {"approved", "rejected"},
        "approved":  {"paid"},
        "rejected":  {"draft"},
    }
    for _, allowed := range valid[from] {
        if allowed == to {
            return true
        }
    }
    return false
}
```

### 7.3.3. Triggering Synchronous Side Effects That Must Be Transactional

If a side effect must either happen atomically with the main record mutation or not at all, trigger it in `after_save` (which runs inside the transaction), not in `before_save`. Use `before_save` only for read-based guards and validations.

The exception: calling `repo.WithTx` explicitly in `before_save` opens a transaction around both the guard check and the subsequent persist, preventing time-of-check/time-of-use races (TOCTOU) on critical checks like stock reservation.

---

## 7.4. The `after_save` Hook

### 7.4.1. Purpose — Post-Persist Side Effects, Cache Invalidation, Event Emission

`after_save` runs immediately after the row is written to the database, while the transaction is still open. At this point, the record has a guaranteed `id` and all its fields are committed to the row (though not yet visible outside the transaction).

**Canonical `after_save` uses:**
- Publishing to an outbox table for event-driven integration
- Invalidating Redis cache entries for the affected records
- Updating denormalised counters (e.g. `customer.invoice_count`)
- Logging to the audit trail
- Signalling a Temporal workflow

### 7.4.2. Still Inside the Transaction — Implications

Because `after_save` runs inside the transaction:

1. **Writes to the outbox table are atomic with the main record** — if the transaction rolls back, the event is never emitted. This is the correct pattern for event-driven systems.

2. **Reads see the just-written data** — you can read back the record you just saved using the transactional repository reference.

3. **Errors in `after_save` roll back the entire transaction** — the main record insert is also rolled back. Design `after_save` logic to be unlikely to fail.

4. **Long operations block the connection** — never make HTTP calls, send emails, or sleep inside `after_save`. These hold the transaction open and starve the connection pool.

### 7.4.3. Triggering Temporal Workflows from `after_save`

The recommended pattern for workflow triggering is to write a "workflow intent" record to an outbox table in `after_save`, then poll that outbox from a Temporal worker:

```go
func (Invoice) AfterSave(ctx context.Context, r *schema.EntityRecord, op schema.MutationOp) error {
    if op == schema.OpCreate || (op == schema.OpUpdate && r.GetString("status") == "submitted") {
        // Write to outbox — atomic with the invoice row
        return outboxRepo.Create(ctx, &WorkflowOutbox{
            WorkflowType:   "InvoiceApproval",
            EntityType:     "Invoice",
            EntityID:       r.GetUUID("id"),
            TenantID:       tenant.FromContext(ctx).ID,
            Payload:        r.ToJSON(),
            ScheduledAfter: time.Now(),
        })
    }
    return nil
}
```

For low-latency use cases where the invoice must be visible in the workflow immediately, you may call `temporal.Client.SignalWithStartWorkflow` directly from `after_save`, but this couples the DB transaction to a network call — be prepared for the transaction to roll back if the Temporal cluster is unavailable.

### 7.4.4. Avoiding Slow Operations Inside `after_save`

**Never do in `after_save`:**
- HTTP calls to external APIs
- Sending emails or SMS
- Reading from Redis (Redis is fast, but network calls are still network calls under load)
- Complex aggregation queries (defer to Temporal)
- Sleeping or retrying

**Always acceptable in `after_save`:**
- Inserting into the outbox table
- Updating a sibling column on the same table
- Incrementing a denormalised counter
- Writing to the audit log table (a simple INSERT)

---

## 7.5. The `before_delete` Hook

### 7.5.1. Purpose — Guard Deletion, Check Referential Integrity That the DB Cannot

The database enforces referential integrity at the FK level (ON DELETE RESTRICT). But ERP has business-level integrity that the DB cannot enforce:

- Cannot delete a product that has unshipped sales order lines
- Cannot delete a cost centre that has non-zero GL balance
- Cannot delete a user who is the sole approver for a workflow tier
- Cannot delete a bank account with pending payment batches

```go
func (CostCentre) BeforeDelete(ctx context.Context, id uuid.UUID) error {
    // Check for non-zero GL balance
    balance, err := glRepo.Aggregate(ctx,
        filter.Eq("cost_centre_id", id),
        aggregate.Sum("amount"),
    )
    if err != nil {
        return err
    }
    if !balance.IsZero() {
        return errs.NewBusinessError("NONZERO_BALANCE",
            "cost centre has a non-zero GL balance of KES %s and cannot be deleted",
            balance.StringFixed(2),
        )
    }

    // Check for child cost centres
    count, err := costCentreRepo.Count(ctx, filter.Eq("parent_id", id))
    if err != nil {
        return err
    }
    if count > 0 {
        return errs.NewBusinessError("HAS_CHILDREN",
            "cost centre has %d sub-centres; reassign or delete them first", count)
    }

    return nil
}
```

### 7.5.2. Soft Delete Pattern — Marking `deleted_at` Instead of Hard Delete

Most ERP records should never be hard-deleted. Invoice line items, GL postings, and audit records must be retained for compliance. Use the soft-delete pattern:

```go
func (Invoice) Fields() []ent.Field {
    return []ent.Field{
        // ...
        field.Time("deleted_at").
            Optional().
            Nillable().
            Comment("Soft delete timestamp; nil means active"),
    }
}

// Privacy policy automatically excludes soft-deleted records
func (Invoice) Policy() ent.Policy {
    return privacy.Policy{
        privacy.QueryRuleFunc(func(ctx context.Context, q *ent.InvoiceQuery) error {
            q.Where(invoice.DeletedAtIsNil())
            return nil
        }),
    }
}
```

The `DELETE /api/v1/invoices/{id}` endpoint sets `deleted_at = now()` rather than removing the row. Hard delete is available only via an admin-level `DELETE /api/v1/invoices/{id}/purge` endpoint that requires `platform:admin` permission.

### 7.5.3. Returning User-Facing Errors from `before_delete`

Return `*errs.BusinessError` with a meaningful code and user-facing message. The framework maps these to 422 Unprocessable Entity responses:

```json
{
  "status": 422,
  "code": "NONZERO_BALANCE",
  "message": "cost centre has a non-zero GL balance of KES 45,200.00 and cannot be deleted"
}
```

Never return a raw database error from `before_delete`. Raw errors expose schema details and are not actionable by users.

---

## 7.6. The `on_submit` and `on_cancel` Hooks

### 7.6.1. What Submission Means in ERP Context — Document Finalisation

Submission is the act of moving an ERP document from a mutable draft state to a locked, legally significant state. An invoice that is submitted:
- Has been reviewed and approved by the business
- Has generated accounting entries (GL postings)
- Has been (or will be) transmitted to the customer
- Cannot be casually edited

This is distinct from mere record creation. Submission triggers a status transition and a set of irreversible side effects.

```go
func (Invoice) OnSubmit(ctx context.Context, r *schema.EntityRecord) error {
    // 1. Lock the naming series number (no longer retractable)
    // Already set at create time via naming series

    // 2. Generate GL postings
    if err := glService.PostInvoice(ctx, r); err != nil {
        return fmt.Errorf("GL posting failed: %w", err)
    }

    // 3. Set submission metadata
    r.Set("submitted_at", time.Now())
    r.Set("submitted_by_id", actor.FromContext(ctx).ID)

    // 4. Trigger KRA eTIMS transmission via outbox
    return outboxRepo.Create(ctx, &WorkflowOutbox{
        WorkflowType: "KRAeTIMSSubmit",
        EntityID:     r.GetUUID("id"),
        TenantID:     tenant.FromContext(ctx).ID,
    })
}
```

### 7.6.2. Immutability After Submission — Which Fields Lock and Which Do Not

After submission, the document enters a locked state. Field immutability is enforced by the `before_save` hook:

```go
func (Invoice) BeforeSave(ctx context.Context, r *schema.EntityRecord, op schema.MutationOp) error {
    if op == schema.OpUpdate {
        current, _ := invoiceRepo.Get(ctx, r.GetUUID("id"))
        if current.GetString("status") == "submitted" || current.GetString("status") == "paid" {
            // Only certain administrative fields can change post-submission
            allowedMutations := map[string]bool{
                "payment_terms_note": true,  // internal note
                "tags":               true,  // tagging
            }
            for _, changedField := range r.ChangedFields() {
                if !allowedMutations[changedField] {
                    return errs.NewBusinessError("SUBMITTED_INVOICE_IMMUTABLE",
                        "field %q cannot be changed on a submitted invoice", changedField)
                }
            }
        }
    }
    return nil
}
```

Fields that typically lock on submission:
- All line items and amounts
- Customer reference
- Invoice date and period
- Tax calculations
- Naming series number

Fields that may remain mutable:
- Internal notes
- Tags and classifications
- Payment tracking fields

### 7.6.3. `on_cancel` as a Compensating Action — Reversing GL Postings, Stock Moves

Cancellation reverses the effects of submission. Every `on_submit` action should have a corresponding `on_cancel` compensating action:

```go
func (Invoice) OnCancel(ctx context.Context, r *schema.EntityRecord) error {
    // 1. Reverse GL postings
    if err := glService.ReverseInvoicePostings(ctx, r); err != nil {
        return fmt.Errorf("GL reversal failed: %w", err)
    }

    // 2. If goods were delivered, trigger goods-return workflow
    if r.GetString("fulfilment_status") == "delivered" {
        return outboxRepo.Create(ctx, &WorkflowOutbox{
            WorkflowType: "InvoiceCancellationReturn",
            EntityID:     r.GetUUID("id"),
        })
    }

    // 3. Notify customer via outbox
    return outboxRepo.Create(ctx, &WorkflowOutbox{
        WorkflowType: "CancellationNotification",
        EntityID:     r.GetUUID("id"),
    })
}
```

**Critical**: every reversal must be idempotent. If the cancellation workflow is retried (Temporal retries on failure), calling `ReverseInvoicePostings` twice must not double-reverse. Use a `reversal_posted_at` field as an idempotency guard.

### 7.6.4. Amendment Workflow — Creating a New Version of a Submitted Document

When a submitted invoice has an error, the correct ERP pattern is **amendment** — not editing the original:

1. Cancel the original invoice (triggers GL reversal)
2. Create an amendment invoice that references the original
3. Submit the amendment invoice (generates new GL postings)

```go
func (InvoiceAmendment) OnSubmit(ctx context.Context, r *schema.EntityRecord) error {
    originalID := r.GetUUID("original_invoice_id")

    // Cancel the original
    if err := invoiceService.Cancel(ctx, originalID, "amended by "+r.GetString("name")); err != nil {
        return err
    }

    // Post the amendment
    return glService.PostInvoice(ctx, r)
}
```

The amendment workflow preserves the audit trail: both the original and the amendment exist in the database, linked by `original_invoice_id`.

---

## 7.7. Writing Testable Hooks

### 7.7.1. The Interceptor Pattern — Hooks as Dependencies, Not Global Registrations

Avoid registering hooks as global singletons. Instead, inject them as constructor dependencies:

```go
// BAD — global registration, untestable
func init() {
    invoice.RegisterBeforeSave(enforceGlobalCreditLimit)
}

// GOOD — injected dependency
type InvoiceService struct {
    creditChecker CreditChecker
    repo          InvoiceRepository
}

func (s *InvoiceService) Create(ctx context.Context, input CreateInvoiceInput) (*Invoice, error) {
    // Build the record
    record := s.buildRecord(input)

    // Run before_save inline — testable
    if err := s.creditChecker.Check(ctx, record); err != nil {
        return nil, err
    }

    return s.repo.Create(ctx, record)
}
```

The `CreditChecker` interface can be mocked in tests. The global registration pattern cannot.

### 7.7.2. Unit Testing Hooks in Isolation Without a Live Database

```go
func TestCreditLimitHook(t *testing.T) {
    mockCustomerRepo := &MockCustomerRepository{
        GetFn: func(_ context.Context, id uuid.UUID) (*Customer, error) {
            return &Customer{CreditLimitKSH: decimal.NewFromInt(10000)}, nil
        },
    }
    mockInvoiceRepo := &MockInvoiceRepository{
        AggregateFn: func(_ context.Context, _ filter.Filter, _ aggregate.Spec) (decimal.Decimal, error) {
            return decimal.NewFromInt(9000), nil  // KES 9,000 outstanding
        },
    }

    hook := NewCreditLimitHook(mockCustomerRepo, mockInvoiceRepo)

    // Invoice for KES 2,000 — would breach limit
    record := schema.NewEntityRecord("Invoice")
    record.Set("customer_id", uuid.New())
    record.Set("total_amount", decimal.NewFromInt(2000))

    err := hook.BeforeSave(context.Background(), record, schema.OpCreate)
    require.Error(t, err)

    var bizErr *errs.BusinessError
    require.ErrorAs(t, err, &bizErr)
    assert.Equal(t, "CREDIT_LIMIT_EXCEEDED", bizErr.Code)
}
```

### 7.7.3. Integration Testing Hooks With an In-Memory ent Client

For hooks that perform complex queries, use the `enttest` package with an SQLite backend:

```go
func TestInvoiceLifecycle(t *testing.T) {
    client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
    defer client.Close()

    ctx := tenant.WithContext(context.Background(), &Tenant{ID: uuid.New()})
    ctx = actor.WithContext(ctx, &Actor{ID: uuid.New(), Roles: []string{"finance_manager"}})

    // Create a customer
    customer := client.Customer.Create().
        SetName("Acme Ltd").
        SetCreditLimitKsh(decimal.NewFromInt(50000)).
        SaveX(ctx)

    // Create an invoice
    inv, err := invoiceService.Create(ctx, CreateInvoiceInput{
        CustomerID:  customer.ID,
        TotalAmount: decimal.NewFromInt(5000),
    })
    require.NoError(t, err)
    require.Equal(t, "draft", inv.Status)

    // Submit the invoice — should generate GL postings
    submitted, err := invoiceService.Submit(ctx, inv.ID)
    require.NoError(t, err)
    require.Equal(t, "submitted", submitted.Status)
    require.NotNil(t, submitted.SubmittedAt)
}
```

---

## 7.8. Chaining Multiple Hooks

### 7.8.1. Hook Execution Order When Multiple Hooks Are Registered

When multiple hooks are registered for the same lifecycle event on the same entity, they execute in registration order. Registration order is determined by the slice returned from `Hooks()`:

```go
func (Invoice) Hooks() []ent.Hook {
    return []ent.Hook{
        AuditHook{},          // runs first
        CreditCheckHook{},    // runs second
        NotificationHook{},   // runs third
    }
}
```

If the first hook fails, subsequent hooks do not run. This short-circuit behaviour on error is intentional and correct: if the credit check fails, there is no reason to check notifications.

### 7.8.2. Early Exit — Stopping the Chain Without an Error

To stop hook execution without returning an error (e.g. to skip processing for records in a specific state), return the sentinel `schema.ErrSkipHooks`:

```go
func (h AuditHook) BeforeSave(ctx context.Context, r *schema.EntityRecord, op schema.MutationOp) error {
    // Don't audit system-generated records
    if actor.FromContext(ctx).IsSystem() {
        return schema.ErrSkipHooks  // stops chain cleanly, no error
    }
    return h.logChange(ctx, r, op)
}
```

### 7.8.3. Sharing Context Between Chained Hooks

Hooks can share computed data by storing values in the context using typed keys:

```go
type creditCheckResultKey struct{}

type CreditCheckHook struct{}

func (h CreditCheckHook) BeforeSave(ctx context.Context, r *schema.EntityRecord, op schema.MutationOp) error {
    result := &CreditCheckResult{
        AvailableCredit: computeAvailableCredit(ctx, r),
    }
    // Store result in context for downstream hooks
    *ctx = context.WithValue(*ctx, creditCheckResultKey{}, result)
    return nil
}

type NotificationHook struct{}

func (h NotificationHook) BeforeSave(ctx context.Context, r *schema.EntityRecord, op schema.MutationOp) error {
    result, _ := ctx.Value(creditCheckResultKey{}).(*CreditCheckResult)
    if result != nil && result.AvailableCredit.LessThan(decimal.NewFromInt(5000)) {
        // Warn the user that they're approaching their credit limit
        r.SetWarning("credit_limit", "You have KES %s of credit remaining",
            result.AvailableCredit.StringFixed(2))
    }
    return nil
}
```

Use typed context keys (not string keys) to avoid collisions between hooks from different modules.

---

## Chapter Summary

Chapter 7 defines the five lifecycle stages and their transaction boundaries (§7.1), then documents each hook type with real patterns:

- **`before_validate`** (outside TX) — normalise and compute derived fields before validators run
- **`before_save`** (outside TX by default) — enforce business rules with the fully validated record; cheap to abort
- **`after_save`** (inside TX) — transactional side effects; errors cause full rollback; must be fast (< 50ms)
- **`before_delete`** — guard deletion; soft-delete pattern returns `hook.ErrSoftDeleted`
- **`on_submit` / `on_cancel`** — document finalisation and compensating reversal

The two most critical rules:
1. **Never make external API calls inside `before_save` or `after_save`** — they hold the DB transaction open and cause connection exhaustion under load. Move all I/O to Temporal activities.
2. **`after_save` errors roll back the entire transaction** — the primary record write and all side effects undo atomically. Design `after_save` to fail rarely; move conditional guards to `before_save` where aborts are cheaper.

**Next chapters to read:**

- [§8 — The Persistence Interface](08-persistence-interface.md) — the `EntityRepository` methods called throughout this chapter fully specified with their filter DSL and error types
- [§27 — Defining Workflows](../part-05-workflow/27-defining-workflows.md) — the Temporal workflows triggered from `on_submit` and `after_save` hooks: retry policies, versioning, signals
- [§29 — Saga Pattern](../part-05-workflow/29-saga-pattern.md) — the reversal workflows triggered by `on_cancel` hooks follow the saga pattern with compensating transactions
