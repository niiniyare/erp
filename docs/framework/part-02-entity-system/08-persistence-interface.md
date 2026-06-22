---
title: "Chapter 8: The Persistence Interface"
part: "Part II — The EntityDefinition System"
chapter: 8
section: "08-persistence-interface"
related:
  - "[Chapter 6: Edges](06-edges.md)"
  - "[Chapter 7: The EntityRecord Lifecycle](07-entity-record-lifecycle.md)"
  - "[Chapter 9: Privacy Policies](09-privacy-policies.md)"
  - "[Chapter 11: Database Migrations](11-database-migrations.md)"
---

# Chapter 8: The Persistence Interface

The EntityRepository interface is the persistence contract in Awo. All business logic depends on this interface — never on a specific ORM, database driver, or SQL dialect. This chapter explains the interface's design rationale, every method in detail, the filter and query DSL, the reference `ent` implementation, and the path for replacing or extending that implementation.

---

## 8.1. The `EntityRepository` Interface

### 8.1.1. Why the Interface Exists — The Persistence Layer Must Be Swappable

ERP systems have a longer operational lifespan than most software. An ERP deployed in 2025 may still be running in 2035, by which time `entgo.io` may be deprecated, PostgreSQL may have been replaced by a NewSQL database, or a customer may require a different storage backend entirely (Oracle, SQL Server, CockroachDB).

If business logic calls `ent.Client.Invoice.Query()` directly, none of that is possible without rewriting the business layer. If business logic calls `EntityRepository.Query()`, the implementation can change without touching a single line of domain code.

The interface also enables testing without a live database: mock implementations of `EntityRepository` satisfy the interface and allow unit tests to run in milliseconds.

### 8.1.2. Interface Contract — The Methods Every Implementation Must Provide

```go
// EntityRepository is the universal persistence contract for all entity types.
// All framework code depends on this interface. Never depend on a concrete implementation.
type EntityRepository[T Entity] interface {
    // Read operations
    Get(ctx context.Context, id uuid.UUID) (T, error)
    Query(ctx context.Context, f Filter, opts ...QueryOption) ([]T, PageInfo, error)
    Exists(ctx context.Context, f Filter) (bool, error)
    Count(ctx context.Context, f Filter) (int, error)
    Aggregate(ctx context.Context, f Filter, spec AggregateSpec) (AggregateResult, error)

    // Write operations
    Create(ctx context.Context, input CreateInput) (T, error)
    Update(ctx context.Context, id uuid.UUID, input UpdateInput) (T, error)
    Delete(ctx context.Context, id uuid.UUID) error
    BulkCreate(ctx context.Context, inputs []CreateInput) ([]T, error)
    BulkUpdate(ctx context.Context, f Filter, patch Patch) (int, error)

    // Transaction support
    WithTx(ctx context.Context, fn func(ctx context.Context, repo EntityRepository[T]) error) error
}
```

The interface is generic: `EntityRepository[Invoice]` is distinct from `EntityRepository[Customer]`. This gives compile-time type safety without sacrificing the abstraction.

### 8.1.3. What the Interface Intentionally Does NOT Expose

The interface deliberately excludes:

**No raw SQL**: Callers cannot inject SQL strings. All filtering is through the typed `Filter` DSL.

**No ORM-specific types**: No `*ent.Invoice`, no `*sqlx.Rows`, no `pgx.Batch`. Callers receive typed entity structs (from generated code) or `EntityRecord` for custom entities.

**No connection management**: The repository handles pool management, search_path setting, and retry logic internally. Callers receive a logical repository, not a physical connection.

**No lazy-loading magic**: There are no ORM-style proxy objects that trigger queries when accessed. Related data must be explicitly requested via `QueryOption`s.

**No schema-specific predicates**: The `Filter` DSL is database-agnostic. The implementation translates predicates to ent predicates (which become SQL). This means the same filter expression works against any compliant implementation.

### 8.1.4. How Framework Code Depends on the Interface

Every framework service declares its dependency as an interface:

```go
type InvoiceService struct {
    repo     InvoiceRepository      // EntityRepository[Invoice]
    customer CustomerRepository     // EntityRepository[Customer]
    gl       GLRepository           // EntityRepository[GLEntry]
    temporal temporal.Client
}

func NewInvoiceService(
    repo InvoiceRepository,
    customer CustomerRepository,
    gl GLRepository,
    tc temporal.Client,
) *InvoiceService {
    return &InvoiceService{repo: repo, customer: customer, gl: gl, temporal: tc}
}
```

Wire (Google's dependency injection) binds the concrete `ent` implementation to the interface at startup:

```go
// In wire providers
func ProvideInvoiceRepository(client *ent.Client) InvoiceRepository {
    return &entInvoiceRepository{client: client}
}
```

During tests, a mock or test-double is injected instead:

```go
func TestInvoiceService(t *testing.T) {
    svc := NewInvoiceService(
        &MockInvoiceRepository{...},
        &MockCustomerRepository{...},
        &MockGLRepository{...},
        temporaltest.NewClient(t),
    )
    // test against svc
}
```

---

## 8.2. Interface Methods — Read Operations

### 8.2.1. `Get(ctx, id) → (T, error)`

Fetches a single record by primary key UUID. Returns `errs.ErrNotFound` (which maps to HTTP 404) when the record does not exist.

```go
invoice, err := repo.Get(ctx, invoiceID)
if err != nil {
    if errors.Is(err, errs.ErrNotFound) {
        return nil, errs.NotFound("Invoice", invoiceID)
    }
    return nil, err
}
```

Privacy policies are applied to `Get`: if the caller's policy excludes the record, `Get` returns `ErrNotFound` (not `ErrForbidden`). Leaking the existence of a record to an unauthorised caller via a 403 is itself a security vulnerability.

`Get` does not populate related entities by default. To load edges, use `GetWith`:

```go
invoice, err := repo.GetWith(ctx, invoiceID,
    query.With(edge.Customer),
    query.With(edge.Items),
)
// invoice.Customer and invoice.Items are pre-populated
```

### 8.2.2. `Query(ctx, filter) → ([]T, PageInfo, error)`

Returns a paginated list of records matching the filter. `PageInfo` carries the cursor for the next page and the total count (if requested).

```go
invoices, page, err := repo.Query(ctx,
    filter.Eq("status", "submitted").
    And(filter.Gte("total_amount", decimal.NewFromInt(10000))),
    query.Limit(20),
    query.After("cursor_from_previous_response"),
)
```

`Query` applies privacy policies to every result. Records that the caller's policy excludes are silently omitted — the total count reflects only visible records.

### 8.2.3. `Exists(ctx, predicate) → (bool, error)`

Checks existence without fetching the record. More efficient than `Count > 0` for large tables.

```go
exists, err := repo.Exists(ctx,
    filter.Eq("email", input.Email).
    And(filter.Neq("id", currentUserID)),
)
if err != nil {
    return err
}
if exists {
    return validate.FieldErrorf("email", "email address is already registered")
}
```

Under the hood: `SELECT EXISTS(SELECT 1 FROM invoices WHERE ... LIMIT 1)` — the DB stops scanning after the first match.

### 8.2.4. `Count(ctx, filter) → (int, error)`

Returns the count of records matching the filter. Use for pagination total counts, dashboard metrics, and threshold checks.

```go
overdueCount, err := repo.Count(ctx,
    filter.Eq("status", "overdue").
    And(filter.Lt("due_date", time.Now())),
)
```

Note: `COUNT(*)` on large unindexed tables is expensive. For dashboard widgets that display approximate counts, consider maintaining a denormalised counter column updated by hooks, or use PostgreSQL's `pg_class.reltuples` for approximate counts.

### 8.2.5. `Aggregate(ctx, spec) → (AggregateResult, error)`

Runs aggregation functions (SUM, AVG, MIN, MAX, COUNT, GROUP BY) over a filtered record set.

```go
// Total outstanding invoices by customer
result, err := repo.Aggregate(ctx,
    filter.Eq("status", "submitted"),
    aggregate.GroupBy("customer_id").Sum("total_amount").Count("id"),
)

for _, row := range result.Rows {
    customerID := row.GetUUID("customer_id")
    total := row.GetDecimal("sum_total_amount")
    count := row.GetInt("count_id")
    fmt.Printf("Customer %s: %d invoices totalling KES %s\n",
        customerID, count, total.StringFixed(2))
}
```

The `AggregateSpec` DSL generates SQL like:
```sql
SELECT customer_id,
       SUM(total_amount) AS sum_total_amount,
       COUNT(id) AS count_id
FROM invoices
WHERE status = 'submitted'
GROUP BY customer_id;
```

---

## 8.3. Interface Methods — Write Operations

### 8.3.1. `Create(ctx, input) → (T, error)`

Creates a new record. The `CreateInput` carries field values as a typed map. The framework applies defaults, runs `before_validate`, validates, runs `before_save`, opens a transaction, inserts the row, runs `after_save`, and commits.

```go
invoice, err := repo.Create(ctx, invoice.Create{
    CustomerID:  customerID,
    TotalAmount: decimal.NewFromFloat(45000.00),
    DueDate:     time.Now().AddDate(0, 0, 30),
    Status:      "draft",
    Items: []invoice.ItemCreate{
        {ProductID: prodID, Quantity: 10, UnitPrice: decimal.NewFromFloat(4500.00)},
    },
})
```

On success, the returned entity has all computed fields: `id`, `created_at`, `invoice_number` (from naming series), and any `before_validate`-computed fields.

### 8.3.2. `Update(ctx, id, input) → (T, error)`

Applies a partial update (patch semantics). Only the fields present in `UpdateInput` are changed. Fields not included in the input are left at their current values.

```go
updated, err := repo.Update(ctx, invoiceID, invoice.Update{
    Tags: []string{"priority", "key-account"},
})
```

Immutable fields included in the input trigger `ErrImmutableField`. Fields that fail validation trigger a 422 with field-level errors. Fields that fail business rules in `before_save` trigger the appropriate `*errs.BusinessError`.

### 8.3.3. `Delete(ctx, id) → error`

Deletes a record by primary key. For soft-deletable entities, this sets `deleted_at` instead of removing the row. For hard-deletable entities, it executes `DELETE FROM table WHERE id = $1`.

`before_delete` runs before the SQL executes. If `before_delete` returns an error, the delete is aborted.

```go
if err := repo.Delete(ctx, invoiceID); err != nil {
    var bizErr *errs.BusinessError
    if errors.As(err, &bizErr) {
        // Surface to user as 422
        return c.Status(fiber.StatusUnprocessableEntity).JSON(bizErr)
    }
    return err
}
```

### 8.3.4. `BulkCreate(ctx, inputs) → ([]T, error)`

Creates multiple records in a single database statement. All records in the batch succeed or all fail — the operation is atomic.

```go
items, err := itemRepo.BulkCreate(ctx, lineItems)
```

Bulk creates run `before_validate` and validators on each record individually, collect all errors, and return them as a `BulkValidationError` before opening a transaction. A single invalid record aborts the entire bulk operation before touching the DB.

For large bulk imports (>10,000 rows), use the `BulkImport` workflow instead, which processes records in batches via Temporal activities and provides progress reporting and partial failure recovery.

### 8.3.5. `BulkUpdate(ctx, filter, patch) → (int, error)`

Applies a patch to all records matching the filter. Returns the count of affected rows.

```go
// Mark all overdue invoices as requiring attention
affected, err := repo.BulkUpdate(ctx,
    filter.Eq("status", "submitted").And(filter.Lt("due_date", time.Now())),
    patch.Set("status", "overdue").Set("overdue_notified_at", nil),
)
```

**Warning**: `BulkUpdate` bypasses per-record hooks. `before_save` and `after_save` do not run for each affected record. Use bulk update only for administrative operations where hook execution is not required, and always document the bypass explicitly in code comments.

For updates that require hooks, use the `BulkUpdateWithHooks` method, which fetches each record, runs it through the full lifecycle, and commits in batches (slower but safe):

```go
// Hooks run for each record — safe but slower
affected, err := repo.BulkUpdateWithHooks(ctx,
    filter.Eq("status", "draft").And(filter.Lt("due_date", time.Now())),
    func(ctx context.Context, record T) error {
        record.Set("status", "overdue")
        return nil
    },
)
```

---

## 8.4. Interface Methods — Transaction Support

### 8.4.1. `WithTx(ctx, fn) → error`

Opens a database transaction, passes the transactional repository to `fn`, and commits on success or rolls back on any error.

```go
err := repo.WithTx(ctx, func(ctx context.Context, txRepo InvoiceRepository) error {
    // All operations through txRepo are in the same transaction

    invoice, err := txRepo.Create(ctx, invoiceInput)
    if err != nil {
        return err  // triggers rollback
    }

    for _, item := range lineItems {
        item.InvoiceID = invoice.ID
        if _, err := itemRepo.WithTxRepo(ctx).Create(ctx, item); err != nil {
            return err  // triggers rollback
        }
    }

    return nil  // triggers commit
})
```

The transactional context is propagated through `ctx`. Any repository that reads the transaction from context participates in the same transaction automatically.

### 8.4.2. Nested Transaction Semantics — Savepoints vs Flat Transactions

Awo uses **flat transactions with savepoints** for nested `WithTx` calls:

```go
// Outer transaction
err := repo.WithTx(ctx, func(ctx context.Context, outer InvoiceRepository) error {
    // ... outer work ...

    // Inner "transaction" — actually a savepoint
    err := itemRepo.WithTx(ctx, func(ctx context.Context, inner ItemRepository) error {
        // ... inner work ...
        return nil  // releases savepoint
    })
    if err != nil {
        // Inner savepoint was rolled back; outer transaction continues
        log.Warn("item creation failed, continuing without items")
    }

    return nil  // outer transaction commits
})
```

If the inner `WithTx` fails, only the inner savepoint is rolled back. The outer transaction continues. This allows partial-success patterns — but use them cautiously in ERP, where partial success is often incorrect behaviour.

### 8.4.3. Passing the Transactional Repository Through Context

The framework stores the active transaction in the context under a typed key. Helper functions extract it:

```go
// In your service method — works regardless of whether a TX is active
func (s *InvoiceService) createGLEntries(ctx context.Context, invoice *Invoice) error {
    // If a TX is active in ctx, this uses it. Otherwise, auto-transaction.
    return s.gl.WithTx(ctx, func(ctx context.Context, gl GLRepository) error {
        for _, posting := range computeGLPostings(invoice) {
            if _, err := gl.Create(ctx, posting); err != nil {
                return err
            }
        }
        return nil
    })
}
```

### 8.4.4. Rollback on Hook Error — Automatic vs Manual

All rollbacks in Awo are automatic. You never call `tx.Rollback()` explicitly. The pattern is:

- Return `nil` from `WithTx` callback → commit
- Return any non-nil `error` → rollback

The `after_save` hook receives the transactional context. If `after_save` returns an error, the transaction is rolled back automatically — including the main entity persist. This is the correct, safe default.

---

## 8.5. The Filter and Query DSL

The Filter DSL is a type-safe query language that translates to SQL predicates. It is database-agnostic by design.

### 8.5.1. `Filter` Struct — Field Predicates, Logical Operators

```go
import "awo.so/core/filter"

f := filter.Eq("status", "submitted").
    And(filter.Gte("total_amount", decimal.NewFromInt(1000))).
    And(filter.Not(filter.In("customer_id", excludedIDs)))
```

Filters compose via `And`, `Or`, `Not` operators. The resulting `filter.Filter` value is a tree of predicates that the repository implementation translates to SQL.

### 8.5.2. Comparison Operators

| Operator | Method | SQL equivalent |
|---|---|---|
| Equal | `filter.Eq("field", val)` | `field = val` |
| Not equal | `filter.Neq("field", val)` | `field != val` |
| Greater than | `filter.Gt("field", val)` | `field > val` |
| Greater or equal | `filter.Gte("field", val)` | `field >= val` |
| Less than | `filter.Lt("field", val)` | `field < val` |
| Less or equal | `filter.Lte("field", val)` | `field <= val` |
| In set | `filter.In("field", vals)` | `field IN (...)` |
| Not in set | `filter.NotIn("field", vals)` | `field NOT IN (...)` |

### 8.5.3. String Operators

| Operator | Method | SQL equivalent |
|---|---|---|
| Contains | `filter.Contains("name", "acme")` | `name ILIKE '%acme%'` |
| Starts with | `filter.StartsWith("code", "INV")` | `name LIKE 'INV%'` |
| Ends with | `filter.EndsWith("email", ".ke")` | `email LIKE '%.ke'` |
| Case-insensitive like | `filter.ILike("name", "%ltd%")` | `name ILIKE '%ltd%'` |

String operators use `ILIKE` (case-insensitive) by default. For exact case-sensitive matching, use `filter.Eq`.

### 8.5.4. Null Operators

```go
filter.IsNull("deleted_at")      // WHERE deleted_at IS NULL
filter.IsNotNull("approved_at")  // WHERE approved_at IS NOT NULL
```

### 8.5.5. Logical Composition

```go
// All conditions must be true (AND)
filter.And(
    filter.Eq("status", "active"),
    filter.Gte("amount", decimal.NewFromInt(100)),
)

// Any condition must be true (OR)
filter.Or(
    filter.Eq("status", "overdue"),
    filter.Lt("due_date", time.Now()),
)

// Negate a condition
filter.Not(filter.Eq("status", "cancelled"))
```

Composing with method chaining (`.And()`, `.Or()`) is equivalent to using the package-level functions:
```go
// These are equivalent:
filter.Eq("a", 1).And(filter.Eq("b", 2))
filter.And(filter.Eq("a", 1), filter.Eq("b", 2))
```

### 8.5.6. JSONB Predicates — Path Operators for Custom Entity Fields

For `JSON` and `custom_fields JSONB` columns:

```go
// Custom field filter: custom_fields->>'industry' = 'Technology'
filter.JSONPath("custom_fields", "industry", filter.Eq("", "Technology"))

// Nested path: custom_fields->'address'->>'city' = 'Nairobi'
filter.JSONPath("custom_fields", "address.city", filter.Eq("", "Nairobi"))

// Containment: custom_fields @> '{"tags": ["vip"]}'::jsonb
filter.JSONContains("custom_fields", map[string]interface{}{"tags": []string{"vip"}})
```

JSONB predicates are only efficient when supported by a GIN index on the JSONB column. Without an index, every row must be scanned. See Chapter 10 for GIN index strategy.

### 8.5.7. Pagination — Cursor-Based (Default) vs Offset-Based

**Cursor-based pagination** (default) is stable under concurrent inserts and deletes. The cursor encodes the sort key of the last seen record:

```go
// First page
invoices, page, _ := repo.Query(ctx, f, query.Limit(20))
cursor := page.EndCursor  // opaque string

// Next page
invoices, page, _ := repo.Query(ctx, f, query.Limit(20), query.After(cursor))
```

**Offset-based pagination** is available for reports and exports where cursor instability is acceptable:

```go
invoices, page, _ := repo.Query(ctx, f,
    query.Limit(100),
    query.Offset(500),
    query.OrderBy("created_at", query.Desc),
)
```

Never use offset pagination for user-facing lists. The "page 5 shows a different set of records than it did two seconds ago" problem is a common source of user confusion and missed records in ERP data exports.

### 8.5.8. Sorting — Multi-Field, Nulls-Last Default

```go
// Sort by amount descending, then by created_at ascending for ties
invoices, _, _ := repo.Query(ctx, f,
    query.OrderBy("total_amount", query.Desc),
    query.OrderBy("created_at", query.Asc),
)
```

Default sort when none is specified: `created_at DESC` (most recent first). Default null handling: `NULLS LAST` on all sorts.

---

## 8.6. The `ent` Reference Implementation

### 8.6.1. How the ent Implementation Satisfies the `EntityRepository` Interface

The `ent` implementation wraps the code-generated `ent.Client` and adapts its API to the `EntityRepository` interface:

```go
type entInvoiceRepository struct {
    client *ent.Client
}

func (r *entInvoiceRepository) Get(ctx context.Context, id uuid.UUID) (*Invoice, error) {
    row, err := r.client.Invoice.Get(ctx, id)
    if err != nil {
        if ent.IsNotFound(err) {
            return nil, errs.ErrNotFound
        }
        return nil, err
    }
    return mapEntToInvoice(row), nil
}

func (r *entInvoiceRepository) Query(ctx context.Context, f Filter, opts ...QueryOption) ([]*Invoice, PageInfo, error) {
    q := r.client.Invoice.Query()
    q = applyFilter(q, f)
    q = applyOptions(q, opts)
    rows, err := q.All(ctx)
    if err != nil {
        return nil, PageInfo{}, err
    }
    return mapEntSliceToInvoices(rows), buildPageInfo(rows, opts), nil
}
// ... remaining methods
```

The mapping functions (`mapEntToInvoice`, `mapEntSliceToInvoices`) convert ent-specific types to domain structs. This indirection is the price of the abstraction — but it also means adding or removing fields from the domain struct does not require changes throughout the codebase.

### 8.6.2. Schema File Layout and Conventions

```
internal/
  ent/
    schema/
      invoice.go          # Field, Edge, Hook, Index, Policy declarations
      invoice_item.go
      customer.go
      ...
    invoice.go            # Generated — do not edit
    invoice_query.go      # Generated — do not edit
    invoice_create.go     # Generated — do not edit
    ...
    client.go             # Generated entry point
    migrate/
      schema.go           # Generated migration schema
```

Schema files are the only files you edit. Generated files are overwritten by `go generate`. Convention:
- One schema file per entity
- Schema file name matches entity name in snake_case
- No business logic in schema files — only field/edge/hook/index/policy declarations

### 8.6.3. How ent Predicates Are Generated from `Filter` Structs

The `applyFilter` function in the ent implementation translates `filter.Filter` trees to ent predicate functions:

```go
func applyFilter(q *ent.InvoiceQuery, f filter.Filter) *ent.InvoiceQuery {
    if f == nil {
        return q
    }
    pred := translateFilter(f)
    return q.Where(pred)
}

func translateFilter(f filter.Filter) predicate.Invoice {
    switch v := f.(type) {
    case *filter.EqFilter:
        return translateEq(v)
    case *filter.AndFilter:
        preds := make([]predicate.Invoice, len(v.Filters))
        for i, sub := range v.Filters {
            preds[i] = translateFilter(sub)
        }
        return invoice.And(preds...)
    case *filter.OrFilter:
        // similar
    case *filter.NotFilter:
        return invoice.Not(translateFilter(v.Filter))
    // ...
    }
}
```

This translation is generated per-entity via code generation. Adding a new field to the schema automatically adds support for filtering by that field.

### 8.6.4. Connection Pool Management with pgx

The `ent` implementation uses `pgx/v5` as its PostgreSQL driver, configured with a connection pool:

```go
pool, err := pgxpool.New(ctx, dsn)
if err != nil {
    return nil, fmt.Errorf("pgxpool.New: %w", err)
}
pool.Config().MaxConns = 20
pool.Config().MinConns = 2
pool.Config().MaxConnLifetime = 30 * time.Minute
pool.Config().MaxConnIdleTime = 5 * time.Minute

drv := entsql.OpenDB("pgx", stdlib.OpenDBFromPool(pool))
client := ent.NewClient(ent.Driver(drv))
```

Pool sizing: `MaxConns = (number of cores * 2) + number of disks` is a common heuristic. For most ERP workloads, 20 connections per application instance is sufficient. Avoid very large pools — PostgreSQL's connection overhead is significant above ~100 connections per database.

### 8.6.5. RLS Tenant Routing in the ent Client

Awo uses a shared schema with PostgreSQL Row-Level Security (RLS) rather than schema-per-tenant. All tenant data lives in the same schema; RLS policies restrict each query to the current tenant's rows.

Tenant routing is set by `store.SetTenantContextFromCtx(ctx)` in the tenant middleware. This calls the `set_tenant_context($1)` stored procedure, which sets the transaction-local `app.current_tenant_id` variable via `set_config('app.current_tenant_id', $1, TRUE)`. All RLS policies on tenant-scoped tables read this variable via `current_tenant_id()`:

```sql
CREATE POLICY tenant_isolation ON invoices
    USING (tenant_id = current_tenant_id());
```

The ent implementation uses a shared pgx connection pool (via PgBouncer in transaction mode). No per-tenant connection pool exists; the `is_local = TRUE` flag ensures the tenant context resets at transaction end so connections are clean for the next request.

### 8.6.6. Known Limitations of the ent Implementation

1. **No polymorphic FK support**: ent does not support polymorphic foreign keys natively. DynamicLink fields are implemented as two plain fields without FK constraint.

2. **Code generation required after schema change**: Every schema change requires running `go generate ./ent/...` before the code compiles.

3. **Migration via Atlas only**: ent's built-in schema migration is not supported in production. All migrations go through Atlas CLI (see Chapter 11).

4. **No partial index support in schema declarations**: Partial indexes (e.g. `CREATE UNIQUE INDEX ON invoices(email) WHERE deleted_at IS NULL`) must be declared as raw SQL in migration files, not in ent schema.

---

## 8.7. Swapping the Implementation

### 8.7.1. When You Would Swap

- **Performance**: ent generates ORM-style SQL that is not always optimal for complex reporting queries. A SQLC implementation with hand-written SQL may outperform ent for analytics-heavy workloads.
- **Licensing**: ent is Apache 2.0; this is almost always acceptable.
- **Database**: If moving to CockroachDB, Spanner, or a non-PostgreSQL store, the ent implementation may need replacement.
- **Team preference**: Teams experienced with SQLC or sqlboiler may prefer a different implementation.

### 8.7.2. What a New Implementation Must Satisfy

Any replacement implementation must:

1. Satisfy the full `EntityRepository[T]` generic interface
2. Apply privacy policies from the entity's `Policy()` method to every read operation
3. Run the full hook lifecycle (before_validate, before_save, after_save, before_delete) on every mutation
4. Respect the transactional context passed via `ctx`
5. Set PostgreSQL `search_path` from the tenant context on every connection
6. Return `errs.ErrNotFound` (not database-specific errors) for missing records
7. Pass the implementation test suite (see §8.7.3)

### 8.7.3. The Implementation Test Suite

```go
// Run the standard suite against any implementation
func TestEntImplementation(t *testing.T) {
    repo := NewEntInvoiceRepository(enttest.Open(t, "sqlite3", ":memory:?_fk=1"))
    persistence.RunEntityRepositoryTestSuite(t, repo)
}

// The suite covers:
// - Create, Get, Update, Delete round-trip
// - Filter operators (all 20+)
// - Pagination (cursor and offset)
// - Aggregate functions
// - Transaction commit and rollback
// - Privacy policy application
// - Concurrent write safety
// - Soft delete visibility
```

Run this suite against your new implementation before deploying. The suite is in `awo.so/testing/persistence`.

### 8.7.4. Registering a Custom Implementation via the Framework Bootstrap

```go
func main() {
    app := awo.New()
    app.Persistence(func(reg persistence.Registry) {
        reg.Register("Invoice", NewSQLCInvoiceRepository(db))
        reg.Register("Customer", NewSQLCCustomerRepository(db))
        // Unregistered entities fall back to the default ent implementation
    })
    app.Start()
}
```

Entity types registered with a custom implementation use that implementation for all framework operations. Unregistered types fall back to the default ent implementation, allowing incremental migration.

---

## Chapter Summary

Chapter 8 defines the complete `EntityRepository` interface contract (§8.1), all read and write methods with usage patterns and error semantics (§8.2–8.3), the transaction support methods and savepoint semantics (§8.4), the full Filter and Query DSL including cursor pagination and JSONB predicates (§8.5), and the ent reference implementation with its RLS-based tenant routing (§8.6).

The three most critical concepts:

- **Interface-only dependency** (§8.1.4): all module code depends on `entity.EntityRepository`, never on `*ent.Client` or pgx directly. This is structurally enforced by Go module visibility.
- **`entity.ErrNotFound` is the canonical not-found sentinel** — check with `errors.Is(err, entity.ErrNotFound)`, not with type assertion. Privacy policies returning `ErrDeny` also surface as `ErrNotFound` to avoid information leakage.
- **`BulkUpdate` bypasses all hooks** (§8.3.5) — it is an administrative tool, not a general-purpose update method. Restrict its use to privileged code paths.

**Next chapters to read:**

- [§9 — Privacy Policies](09-privacy-policies.md) — the row and field visibility policies injected by the `EntityRepository` at query time; both chapters must be understood together to reason about data access
- [§11 — Database Migrations](11-database-migrations.md) — the Atlas migration workflow that keeps the schema in sync with `EntityDefinition` declarations
- [§17 — REST API Conventions](../part-03-api/17-rest-api-conventions.md) — how URL query parameters are translated into `Filter` structs and `QueryOption` values
