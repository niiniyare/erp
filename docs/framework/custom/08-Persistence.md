### Chapter 8 — The Persistence Interface

The `EntityRepository` interface is the single contract between all framework consumers and the persistence layer. Every database operation in Awo — from a route handler reading a list of invoices to a Temporal activity posting a GL entry — goes through this interface. No code outside `internal/store/ent` imports ent packages, references ent schema types, or calls ent query builders directly. This chapter documents the full interface contract, every method signature, the Filter and Query DSL, the ent reference implementation's internal structure, and the procedure for swapping the implementation entirely.

---

#### 8.1. The `EntityRepository` Interface

##### 8.1.1. Why the interface exists — the persistence layer must be swappable

The interface exists because the persistence layer is an implementation detail, not a framework invariant. The invariants are: tenant isolation, hook execution order, privacy policy enforcement, and the `EntityRecord` as the universal record type. These invariants hold regardless of whether the underlying storage uses ent-generated queries, SQLC-generated queries, or a future implementation. Encoding the persistence contract as a Go interface makes this swappability structural rather than aspirational — the Go compiler enforces it.

The interface also enables testing. Any code that depends on `EntityRepository` can be tested against a mock implementation (provided in `github.com/awolabs/awo/testkit`) without a live database. This is the foundation of the hook testability described in §7.7.

##### 8.1.2. Interface contract — the methods every implementation must provide

The full interface is defined in `github.com/awolabs/awo/pkg/entity`:

```go
// Example: The EntityRepository interface definition
package entity

type EntityRepository interface {
    // Read operations
    Get(ctx context.Context, id string, opts ...QueryOption) (EntityRecord, error)
    Query(ctx context.Context, filter Filter, opts ...QueryOption) ([]EntityRecord, PageInfo, error)
    Exists(ctx context.Context, filter Filter) (bool, error)
    Count(ctx context.Context, filter Filter) (int, error)
    Aggregate(ctx context.Context, spec AggregateSpec) (AggregateResult, error)

    // Write operations
    Create(ctx context.Context, input map[string]any) (EntityRecord, error)
    Update(ctx context.Context, id string, input map[string]any) (EntityRecord, error)
    Delete(ctx context.Context, id string) error
    BulkCreate(ctx context.Context, inputs []map[string]any) ([]EntityRecord, error)
    BulkUpdate(ctx context.Context, filter Filter, patch map[string]any) (int, error)

    // Transaction support
    WithTx(ctx context.Context, fn func(ctx context.Context, repo EntityRepository) error) error
}
```

Every method receives a `context.Context` as its first argument. The tenant context is embedded in this `context.Context` by the tenant resolution middleware; implementations extract it using `tenant.FromContext(ctx)` to determine which PostgreSQL schema to target.

##### 8.1.3. What the interface intentionally does NOT expose — no ORM leakage

The interface exposes no ent types, no ent predicates, no ent mutation builders, and no ent client. It exposes no raw `*sql.DB`, no `pgx.Conn`, and no connection pool reference. It exposes no SQL strings. Callers cannot obtain a lower-level database handle from the interface; if they need it, they are doing something that belongs in a separate infrastructure concern, not in module code.

The interface also does not expose schema introspection methods (listing columns, checking index existence) or DDL methods (creating tables, altering columns). Schema management is the responsibility of the Atlas migration tooling (§11), not of runtime code.

##### 8.1.4. How framework code depends on the interface, never the implementation

Every package that performs database operations declares its dependency as `entity.EntityRepository`, not as `*ent.Client` or any other concrete type. Dependencies are injected at bootstrap time via the module `Register(app entity.App)` call, where `app.Repository("EntityName")` returns an `entity.EntityRepository`. Hook structs receive an `entity.EntityRepository`. Temporal activity structs receive an `entity.EntityRepository`. Route handlers obtain one via `entity.Resolve(ctx, tc, entityName)`.

The only exception is `internal/store/ent`, which is the implementation and necessarily imports ent packages. Everything outside `internal/` is forbidden from importing ent packages by Go's module visibility rules.

---

#### 8.2. Interface Methods — Read Operations

##### 8.2.1. `Get(ctx, id) → (EntityRecord, error)` — single record by primary key

`Get` retrieves a single record by its UUID primary key. It returns `entity.ErrNotFound` (which maps to HTTP 404) if no record with the given ID exists in the tenant's schema. It accepts `QueryOption` variadic arguments for eager loading related records via `Include`.

```go
// Example: Fetching a single record with eager-loaded edges
tc, err := tenant.FromContext(ctx)
if err != nil {
    return err
}
repo, err := entity.Resolve(ctx, tc, "SalesInvoice")
if err != nil {
    return err
}
invoice, err := repo.Get(ctx, invoiceID,
    entity.Include("lineItems"),
    entity.Include("customer"),
)
if err != nil {
    if errors.Is(err, entity.ErrNotFound) {
        return fiber.NewError(fiber.StatusNotFound, "invoice not found")
    }
    return fmt.Errorf("fetching invoice: %w", err)
}
```

`Get` always scopes its query to the tenant schema derived from the context. It is impossible to call `Get` without a tenant context; the method will return `entity.ErrMissingTenantContext` if `tenant.FromContext(ctx)` fails.

##### 8.2.2. `Query(ctx, filter) → ([]EntityRecord, PageInfo, error)` — filtered list

`Query` retrieves a paginated, filtered, sorted list of records. The `Filter` argument describes which records to return (see §8.5 for the full Filter DSL). The `PageInfo` return value carries pagination state: total record count, the cursor for the next page, and a `HasNext` boolean. `QueryOption` variadic arguments control pagination, sorting, and eager loading.

```go
// Example: Querying with filter, sort, and cursor pagination
filter := entity.NewFilter().
    Eq("status", "open").
    Gte("requested_at", startDate)

records, pageInfo, err := repo.Query(ctx, filter,
    entity.Sort("requested_at", entity.Desc),
    entity.Limit(25),
    entity.Cursor(cursorToken),
    entity.Include("assignedTechnician"),
)
if err != nil {
    return fmt.Errorf("querying service requests: %w", err)
}
```

When called without a `Cursor` option, `Query` returns the first page. When `pageInfo.HasNext` is true, pass `pageInfo.NextCursor` as the `Cursor` option in the subsequent call to fetch the next page. The cursor is an opaque base64-encoded value; callers must not attempt to parse or construct it manually.

##### 8.2.3. `Exists(ctx, predicate) → (bool, error)` — existence check without fetch

`Exists` performs a `SELECT 1 WHERE ... LIMIT 1` query and returns `true` if at least one matching record exists, `false` otherwise. It is more efficient than `Count` when the exact count is not needed and significantly more efficient than `Query` when the caller only needs to know whether any records match. Use `Exists` in async validators (§5.3.4), in `before_delete` guards (§7.5.1), and anywhere a boolean existence question is being asked.

```go
// Example: Existence check before creating a duplicate
filter := entity.NewFilter().
    Eq("kra_pin", pin).
    Neq("id", currentRecordID)
alreadyExists, err := repo.Exists(ctx, filter)
if err != nil {
    return fmt.Errorf("checking KRA PIN uniqueness: %w", err)
}
if alreadyExists {
    return entity.NewFieldError("kra_pin", "a customer with this KRA PIN already exists")
}
```

##### 8.2.4. `Count(ctx, filter) → (int, error)` — aggregate count

`Count` returns the number of records matching the filter as an integer. It executes a `SELECT COUNT(*) WHERE ...` query and does not load any record data. Use `Count` for pagination total calculation, for dashboard KPI cards showing record counts, and for any operation that needs a precise count without the overhead of loading full records.

```go
// Example: Counting open service requests for a dashboard KPI
filter := entity.NewFilter().Eq("status", "open")
openCount, err := repo.Count(ctx, filter)
if err != nil {
    return fmt.Errorf("counting open requests: %w", err)
}
```

##### 8.2.5. `Aggregate(ctx, spec) → (AggregateResult, error)` — sum, avg, min, max, group by

`Aggregate` executes a single aggregate query described by an `AggregateSpec`. It supports `sum`, `avg`, `min`, `max`, and `count` aggregate functions, applied to a named field, with an optional `Filter` to scope the aggregation and an optional `GroupBy` to produce per-group results.

```go
// Example: Summing invoice totals grouped by status
spec := entity.AggregateSpec{
    Fn:      entity.AggSum,
    Field:   "total_amount",
    GroupBy: "status",
    Filter:  entity.NewFilter().Gte("invoice_date", periodStart),
}
result, err := repo.Aggregate(ctx, spec)
if err != nil {
    return fmt.Errorf("aggregating invoice totals: %w", err)
}
for _, group := range result.Groups {
    fmt.Printf("status=%s total=%s\n", group.Key, group.Value)
}
```

`AggregateResult` contains either a single `Value` (when no `GroupBy` is specified) or a `Groups` slice where each entry has a `Key` (the group-by field value) and a `Value` (the aggregate result). All aggregate values are returned as `decimal.Decimal` regardless of the field's declared type; the caller is responsible for interpreting the precision appropriately.

---

#### 8.3. Interface Methods — Write Operations

##### 8.3.1. `Create(ctx, input) → (EntityRecord, error)`

`Create` inserts a new record. The `input` map is the validated, assembled payload: field names as keys, Go-typed values as values. The framework's field validation pipeline (§5.3) runs before `Create` is called; the implementation does not re-validate. The framework's hook pipeline (`before_save`, naming series assignment) runs before `Create` is called from standard route handlers. When `Create` is called directly from a hook or a Temporal activity, the caller is responsible for ensuring the input is valid.

`Create` returns the full inserted `EntityRecord` including system-managed fields (`id`, `created_at`, `updated_at`, the assigned naming series value). It returns `entity.ErrConflict` if a unique constraint is violated.

```go
// Example: Creating a record from a Temporal activity
func (a *Activities) CreateGLEntry(ctx context.Context, input GLEntryInput) error {
    tc, err := tenant.FromContext(ctx)
    if err != nil {
        return fmt.Errorf("resolving tenant: %w", err)
    }
    repo, err := entity.Resolve(ctx, tc, "LedgerEntry")
    if err != nil {
        return err
    }
    _, err = repo.Create(ctx, map[string]any{
        "account_id":      input.AccountID,
        "debit_amount":    input.DebitAmount,
        "credit_amount":   input.CreditAmount,
        "posting_date":    input.PostingDate,
        "journal_entry_id": input.JournalEntryID,
        "description":     input.Description,
    })
    return err
}
```

##### 8.3.2. `Update(ctx, id, input) → (EntityRecord, error)`

`Update` applies a partial update to an existing record identified by `id`. The `input` map contains only the fields being changed; fields omitted from the map retain their current values. `Update` validates `Immutable` fields (§5.2.3) and `ImmutableAfterSubmit` fields (§7.6.2) before executing the UPDATE SQL. It returns the updated `EntityRecord` reflecting the new state.

`Update` returns `entity.ErrNotFound` if no record with the given ID exists, `entity.ErrConflict` if a unique constraint is violated, and `entity.ErrImmutableField` if an attempt is made to change an immutable field.

```go
// Example: Partial update of a record's status field
_, err = repo.Update(ctx, recordID, map[string]any{
    "status":       "in_progress",
    "assigned_technician_id": technicianID,
})
if err != nil {
    return fmt.Errorf("updating service request: %w", err)
}
```

##### 8.3.3. `Delete(ctx, id) → error`

`Delete` removes the record with the given ID from the tenant schema. It fires `before_delete` hooks before executing the DELETE SQL and `after_delete` hooks (if registered) after. It returns `entity.ErrNotFound` if no record exists, and `entity.ErrRestricted` if a database-level `ON DELETE RESTRICT` constraint prevents deletion (indicating dependent records exist).

For entities using the soft-delete pattern (§7.5.2), `Delete` is intercepted by the `before_delete` hook which converts it to an `Update` and returns `hook.ErrSoftDeleted`. The caller of `Delete` receives a nil error in this case (the operation succeeded as an update); the soft-delete behaviour is entirely transparent to the caller.

##### 8.3.4. `BulkCreate(ctx, inputs) → ([]EntityRecord, error)`

`BulkCreate` inserts multiple records in a single database round-trip using a batched INSERT. It runs the hook pipeline once per record in sequence (not concurrently), maintaining the guarantee that hooks fire for every record. The return value is a slice of inserted `EntityRecord` values in the same order as the input slice.

`BulkCreate` fails atomically: if any record's validation or hook fails, no records from the batch are inserted. The error return includes which record index failed. Use `BulkCreate` for seeding data, importing records from external systems, and any operation that creates multiple homogeneous records in one logical action.

> **Note:** `BulkCreate` runs hooks sequentially, not concurrently. For batches larger than a few hundred records, the hook overhead becomes significant. If the bulk operation is purely a data load with no business logic hooks, consider using `BulkCreate` with a stripped `EntityDefinition` that has no hooks registered, or use Atlas seed data files for initial data loading.

##### 8.3.5. `BulkUpdate(ctx, filter, patch) → (int, error)`

`BulkUpdate` applies a partial update to all records matching the filter in a single UPDATE SQL statement. It returns the number of records updated. `BulkUpdate` does not run hooks on individual records; it is a direct SQL update. Use it only for administrative operations where hook execution is explicitly not required: marking all open service requests as cancelled during a system maintenance window, resetting a status field across all draft documents, or applying a schema migration that populates a new field.

> **Warning:** `BulkUpdate` bypasses all hooks, validators, and privacy policies beyond tenant isolation. It will update every matching record regardless of status, immutability, or permission checks. Restrict its use to privileged administrative code paths, never to general application logic.

---

#### 8.4. Interface Methods — Transaction Support

##### 8.4.1. `WithTx(ctx, fn) → error` — scoped transaction

`WithTx` opens a PostgreSQL transaction, creates a transactional `EntityRepository` scoped to that transaction, and passes it (and the transactional context) to the provided function. If `fn` returns a nil error, the transaction commits. If `fn` returns a non-nil error, the transaction rolls back. The transactional repository passed to `fn` must be used for all database operations that should participate in the transaction; using the outer non-transactional repository from within `fn` creates two independent database operations.

```go
// Example: Using WithTx for a multi-entity atomic operation
err = invoiceRepo.WithTx(ctx, func(txCtx context.Context, txRepo entity.EntityRepository) error {
    tc, err := tenant.FromContext(txCtx)
    if err != nil {
        return err
    }
    // Update the invoice status
    _, err = txRepo.Update(txCtx, invoiceID, map[string]any{
        "status": "posted",
    })
    if err != nil {
        return fmt.Errorf("updating invoice status: %w", err)
    }
    // Create GL entries in the same transaction
    glRepo, err := entity.Resolve(txCtx, tc, "LedgerEntry")
    if err != nil {
        return err
    }
    for _, entry := range glEntries {
        if _, err := glRepo.Create(txCtx, entry); err != nil {
            return fmt.Errorf("creating GL entry: %w", err)
        }
    }
    return nil
})
```

The standard route handler write path (§3.3) opens its own transaction internally. Use `WithTx` only when you need explicit transaction control outside the standard path: in Temporal activities that perform multi-entity writes that must be atomic, in seed data scripts, or in administrative commands.

##### 8.4.2. Nested transaction semantics — savepoints vs flat transactions

If `WithTx` is called on a repository that is already operating inside a transaction (the transactional repository passed to a `WithTx` callback), the nested call creates a PostgreSQL savepoint rather than a new transaction. The inner function can roll back to the savepoint by returning an error without rolling back the outer transaction.

This semantics allows partial rollback within a larger transaction: a Temporal activity that writes ten GL entries can use an inner `WithTx` for each entry, and if entries 1–9 succeed but entry 10 fails, only entry 10 is rolled back to its savepoint; entries 1–9 remain in the outer transaction. Whether this behaviour is desirable depends on the business logic; often, all-or-nothing is safer.

##### 8.4.3. Passing the transactional repository through context

The transactional context (`txCtx`) returned by `WithTx` carries a context value that makes the transactional repository discoverable by `entity.Resolve`. Any call to `entity.Resolve(txCtx, tc, entityName)` within the transaction callback returns a repository that participates in the same transaction, without requiring the caller to pass the `txRepo` argument explicitly. This allows deeply nested function calls to participate in the transaction transparently.

```go
// Example: entity.Resolve honours the transaction from context
err = repo.WithTx(ctx, func(txCtx context.Context, _ entity.EntityRepository) error {
    tc, _ := tenant.FromContext(txCtx)
    // Both Resolve calls return transaction-aware repositories
    invoiceRepo, _ := entity.Resolve(txCtx, tc, "SalesInvoice")
    paymentRepo, _ := entity.Resolve(txCtx, tc, "Payment")
    // Both operations participate in the same transaction
    _, err := invoiceRepo.Update(txCtx, invoiceID, map[string]any{"status": "paid"})
    if err != nil {
        return err
    }
    _, err = paymentRepo.Create(txCtx, paymentData)
    return err
})
```

##### 8.4.4. Rollback on hook error — automatic vs manual

The standard route handler path performs automatic rollback: if any `before_save` or `after_save` hook returns a non-nil error, the framework rolls back the transaction and returns the error to the route handler. No manual rollback call is needed.

When using `WithTx` directly, rollback is also automatic: returning a non-nil error from the `fn` callback causes `WithTx` to call `tx.Rollback()` before returning the error to the caller. The caller of `WithTx` never needs to call rollback manually. Calling rollback manually inside `fn` will cause a "transaction already closed" error from the subsequent `WithTx` infrastructure rollback attempt; do not do this.

---

#### 8.5. The Filter and Query DSL

The Filter DSL provides a type-safe, implementation-independent way to express SQL WHERE clauses. A `Filter` struct is built using `entity.NewFilter()` and method chaining. The implementation translates the `Filter` into SQL (for system entities) or JSONB path expressions (for custom entities).

##### 8.5.1. `Filter` struct — field predicates, logical operators

`entity.NewFilter()` returns an empty filter that matches all records. Adding predicates narrows the result set. Predicates are ANDed by default when chained:

```go
// Example: Basic filter construction
filter := entity.NewFilter().
    Eq("status", "open").
    Gte("requested_at", time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)).
    IsNotNull("assigned_technician_id")
```

This translates to `WHERE status = 'open' AND requested_at >= '2025-06-01' AND assigned_technician_id IS NOT NULL`.

##### 8.5.2. Comparison operators — eq, neq, gt, gte, lt, lte, in, not_in

All comparison operators take a field name and a value. The value must be Go-typed to match the field's declared type; passing a `string` for a `Currency` field will return an `entity.ErrFilterTypeMismatch` at query time.

```
.Eq("field", value)       — field = value
.Neq("field", value)      — field != value
.Gt("field", value)       — field > value
.Gte("field", value)      — field >= value
.Lt("field", value)       — field < value
.Lte("field", value)      — field <= value
.In("field", []T{...})    — field IN (values)
.NotIn("field", []T{...}) — field NOT IN (values)
```

```go
// Example: In and date range predicates
filter := entity.NewFilter().
    In("status", []string{"open", "in_progress"}).
    Gte("invoice_date", periodStart).
    Lte("invoice_date", periodEnd)
```

##### 8.5.3. String operators — contains, starts_with, ends_with, ilike

String operators apply only to `Data`, `SmallText`, and `LongText` fields. They generate LIKE or ILIKE expressions in SQL. All string operators are case-insensitive by default (using ILIKE); use `.CaseSensitive()` on the filter to switch to LIKE for exact-case matching.

```
.Contains("field", "substr")      — field ILIKE '%substr%'
.StartsWith("field", "prefix")    — field ILIKE 'prefix%'
.EndsWith("field", "suffix")      — field ILIKE '%suffix'
.ILike("field", "pattern%")       — field ILIKE 'pattern%'
```

```go
// Example: Case-insensitive search on customer name
filter := entity.NewFilter().
    Contains("customer_name", searchTerm)
```

> **Warning:** `Contains` with a short search term (one or two characters) generates a `ILIKE '%a%'` pattern that cannot use a B-tree index and will perform a sequential scan on large tables. For full-text search on large datasets, use the `q=` query parameter on the API endpoint which is backed by a `pg_trgm` GIN index. Reserve `Contains` for admin lookups on small tables.

##### 8.5.4. Null operators — is_null, is_not_null

```
.IsNull("field")    — field IS NULL
.IsNotNull("field") — field IS NOT NULL
```

Null checks are valid on any nullable field (any field not declared `Required()` with a non-zero type). They are commonly used to filter for records where an optional relationship has not been set, or where a nullable timestamp has not been populated.

##### 8.5.5. Logical composition — And, Or, Not

Predicates chained directly on a `Filter` are implicitly ANDed. For OR and NOT logic, use `entity.Or(filters...)` and `entity.Not(filter)`:

```go
// Example: OR composition for status filtering
filter := entity.NewFilter().
    Or(
        entity.NewFilter().Eq("status", "open"),
        entity.NewFilter().Eq("status", "in_progress"),
    ).
    Gte("requested_at", cutoffDate)
```

This translates to `WHERE (status = 'open' OR status = 'in_progress') AND requested_at >= cutoffDate`. `Or` wraps its arguments in parentheses before ANDing with the parent filter.

```go
// Example: NOT composition
filter := entity.NewFilter().
    Not(entity.NewFilter().In("status", []string{"cancelled", "completed"}))
// Translates to: WHERE NOT (status IN ('cancelled', 'completed'))
```

Arbitrary nesting of `And`, `Or`, and `Not` is supported. The implementation translates the full predicate tree to SQL recursively.

##### 8.5.6. JSONB predicates — path operators for custom entity fields

Custom entity fields are stored in a JSONB column. Filtering on them uses the same Filter DSL with path notation for the field name: `custom_fields.approval_code` refers to the `approval_code` key inside the `custom_fields` JSONB column.

```go
// Example: Filtering on a custom entity JSONB field
filter := entity.NewFilter().
    Eq("custom_fields.approval_tier", "gold").
    IsNotNull("custom_fields.fleet_account_code")
```

The implementation generates PostgreSQL JSONB path expressions: `custom_fields->>'approval_tier' = 'gold'`. For numeric comparisons on JSONB fields, the implementation casts the extracted value: `(custom_fields->>'credit_limit')::numeric >= 50000`. GIN index hints (§10.4.2) accelerate these queries on high-cardinality custom fields.

##### 8.5.7. Pagination — cursor-based (default), offset-based (reports only)

Cursor pagination is the default and the only option for standard API list endpoints. A cursor encodes the sort position of the last record on the current page; the next page begins after that position. This approach is stable under concurrent inserts — a record inserted between pages will not cause a record to be skipped or duplicated across pages.

```go
// Example: Cursor pagination
records, pageInfo, err := repo.Query(ctx, filter,
    entity.Limit(25),
    entity.Cursor(cursorToken), // empty string for the first page
)
nextCursor := pageInfo.NextCursor
hasMore := pageInfo.HasNext
```

Offset pagination is available for report queries where the caller needs to jump to an arbitrary page and accepts the consistency trade-offs:

```go
// Example: Offset pagination for a report
records, pageInfo, err := repo.Query(ctx, filter,
    entity.Limit(100),
    entity.Offset(200),
    entity.UseOffsetPagination(), // explicit opt-in required
)
```

> **Note:** Offset pagination on tables with frequent concurrent writes can produce duplicate or skipped records across pages. Never use `UseOffsetPagination()` for user-facing list views. It is appropriate only for batch report generation where consistency across the entire dataset is achieved by running the query in a single transaction or by accepting the possibility of minor inconsistencies in large reports.

##### 8.5.8. Sorting — multi-field, nulls-last default

Sorting is declared using `entity.Sort(field, direction)` as a `QueryOption`. Multiple `Sort` options are applied in declaration order. The default direction is `entity.Asc`. Nulls are sorted last by default for both ascending and descending sorts (PostgreSQL's `NULLS LAST` syntax).

```go
// Example: Multi-field sort
records, _, err := repo.Query(ctx, filter,
    entity.Sort("posting_date", entity.Desc),
    entity.Sort("created_at", entity.Desc),
    entity.Limit(50),
)
```

Sort fields must be declared in the `EntityDefinition`; attempting to sort by an undeclared field returns `entity.ErrInvalidSortField`. For custom entity JSONB fields, sort using the path notation: `entity.Sort("custom_fields.priority_score", entity.Desc)`. JSONB sort performance depends on whether a GIN index covers the path.

---

#### 8.6. The `ent` Reference Implementation

##### 8.6.1. How the ent implementation satisfies the `EntityRepository` interface

The ent reference implementation lives in `internal/store/ent`. It wraps an `*ent.Client` (the entgo.io ORM client) in a struct that implements `entity.EntityRepository`. Each interface method is translated into the corresponding ent fluent query builder call. The implementation handles: tenant schema routing (setting the pgx connection's search path), `Filter` to ent predicate translation, `QueryOption` processing (limit, cursor, sort), and `AggregateSpec` to ent aggregate query translation.

The implementation is not a generic adapter; it is a hand-maintained translation layer. For each system entity, there is a concrete `EntityRepository` implementation in `internal/store/ent/{entity_name}_repo.go` that imports the ent-generated types for that entity. This means the translation is type-safe within the ent layer even though it is hidden behind the interface.

##### 8.6.2. Schema file layout and conventions

Ent schema files live in `internal/store/ent/schema/`. Each system entity has a corresponding file `{entity_name}.go` that defines the ent schema: fields, edges, indexes, and mixins. The ent schema is the source of truth for the Atlas migration diffing process — Atlas compares the current database schema against what the ent schema would produce and generates a migration file for the difference.

Ent schema files use the ent fluent API and must match the field types and constraints declared in the `EntityDefinition`. The framework's CI runs a consistency check that verifies the ent schema fields match the `EntityDefinition` field declarations; a mismatch fails the build.

##### 8.6.3. How ent predicates are generated from `Filter` structs

The Filter-to-predicate translation walks the `Filter` tree and calls the corresponding ent predicate builder function for each node. For example, `entity.NewFilter().Eq("status", "open")` translates to `entgen.StatusEQ("open")` for an entity that has a generated `Status` field predicate.

For fields that do not have a generated ent predicate (custom fields stored in JSONB, or fields added by mixins), the translation generates a raw SQL predicate using ent's `sql.P` escape hatch. These raw predicates are the mechanism by which JSONB path filters are applied.

##### 8.6.4. Connection pool management with pgx

The ent implementation uses pgx as its database driver. Each tenant has a dedicated pgx connection pool, configured with the tenant's schema in the search path. Pool lifecycle is managed by the tenant context system: a pool is created on first use for a tenant and evicted after a configurable idle timeout. The global pool ceiling prevents total connections from exceeding PostgreSQL's `max_connections` limit.

The pgx pool configuration per tenant:
```
MinConns:         2
MaxConns:         10 (overridable per tenant in TenantConfig)
MaxConnLifetime:  30m
MaxConnIdleTime:  5m
HealthCheckPeriod: 1m
```

##### 8.6.5. Per-tenant schema routing in the ent client

Schema routing is implemented by creating a separate `*ent.Client` per tenant, each configured with a pgx pool whose `search_path` is set to `t_{tenant_slug}`. The client is obtained inside each repository method by calling `tc.EntClient()`, which returns the pre-configured client for the current tenant. No query appends a schema qualifier to table names; the search path handles routing transparently.

##### 8.6.6. Known limitations of the ent implementation

The current ent implementation has three known limitations that module developers should be aware of:

Full-text search using `pg_trgm` is not supported natively by ent's query builder and is implemented via raw SQL predicates. The `q=` search parameter on API list endpoints generates a `similarity()` expression; this is correct but bypasses ent's type safety.

Aggregate queries with multiple group-by fields (e.g., grouping by both `status` and `posting_date`) are not yet supported by the `Aggregate` interface method. Use a dedicated report query via a custom route handler with a raw SQL view for multi-dimensional aggregations.

The `BulkCreate` implementation executes individual INSERTs in sequence rather than a batched `INSERT INTO ... VALUES (...), (...)` statement because ent does not expose a native batch insert API. For large batch inserts (thousands of records), the performance difference is significant. Teams with high-volume bulk import requirements should implement a custom `BulkCreate` override or use the Atlas seed data mechanism.

---

#### 8.7. Swapping the Implementation

##### 8.7.1. When you would swap — performance requirements, licensing, SQLC preference

The most common reason to swap the implementation is a preference for SQLC. SQLC generates type-safe Go functions directly from SQL queries written by the developer, producing code that is closer to the SQL than ent's schema-first approach and easier to optimise for specific query patterns. Teams with deep PostgreSQL expertise often prefer SQLC's transparency.

Another reason is licensing: ent is Apache 2.0 licensed, which is permissive, but some enterprise legal teams have policies against specific dependencies. A third reason is that the ent implementation's known limitations (§8.6.6) are blocking for a specific use case and implementing the correct behaviour is easier in a new implementation than in a patch to the existing one.

##### 8.7.2. What a new implementation must satisfy — full interface contract + test suite

A new implementation must implement all methods of `entity.EntityRepository` with the following behavioural guarantees:

- All read and write operations must scope to the tenant schema derived from the context.
- `Get` must return `entity.ErrNotFound` (not a wrapped error that is not `errors.Is`-matchable) when no record exists.
- `WithTx` must roll back on any non-nil error returned from `fn`.
- `BulkCreate` must be atomic: all records succeed or none are inserted.
- `Query` with cursor pagination must be stable under concurrent inserts.
- All `Filter` predicates documented in §8.5 must be supported.
- Privacy policy predicates injected by the framework must be applied before any query executes.

The implementation test suite in `testkit/store_test.go` runs against any implementation via a test interface. Run it against a new implementation:

```bash
AWO_STORE_IMPL=your_impl go test ./testkit/... -run TestEntityRepositoryContract
```

All tests in `TestEntityRepositoryContract` must pass before the new implementation can be registered.

##### 8.7.3. The implementation test suite — running it against a new implementation

The test suite covers all interface methods, all Filter predicates, cursor and offset pagination stability, transaction rollback on error, nested transaction savepoint semantics, and error type correctness (`ErrNotFound`, `ErrConflict`, `ErrImmutableField`, `ErrRestricted`). It spins up a real PostgreSQL instance using the `testcontainers-go` package; no mocking is used in the contract tests because the contract includes database behaviour.

The test suite does not test performance. Performance benchmarks for the ent implementation are documented in Appendix I; running equivalent benchmarks against a new implementation is the implementer's responsibility.

##### 8.7.4. Registering a custom implementation via the framework bootstrap

Register a custom implementation by providing it to the app builder in `cmd/server/main.go`:

```go
// Example: Registering a custom EntityRepository implementation
package main

import (
    "github.com/awolabs/awo/pkg/entity"
    "myproject/internal/store/sqlc" // custom SQLC implementation
)

func main() {
    app := entity.NewApp(
        entity.WithStoreFactory(sqlc.NewRepositoryFactory),
    )
    // ... module registrations
    app.Run()
}
```

`entity.WithStoreFactory` accepts a `RepositoryFactory` function with signature `func(tc tenant.TenantContext) entity.EntityRepository`. The factory is called once per tenant on first use and its result is cached for the lifetime of the tenant's context in the current process.

---

#### Chapter summary

Chapter 8 defines the complete `EntityRepository` interface contract (§8.1), all five read methods with usage patterns and error semantics (§8.2), all five write methods including `BulkUpdate`'s hook-bypass warning (§8.3), the transaction support methods and their savepoint semantics (§8.4), the full Filter and Query DSL including cursor vs offset pagination and JSONB path predicates (§8.5), the ent reference implementation's internal structure and known limitations (§8.6), and the full procedure for swapping the implementation including the contract test suite (§8.7). The three most critical concepts are the interface-only dependency rule (§8.1.4, which is the structural enforcement of the framework's swappability guarantee), the `WithTx` callback pattern and its automatic rollback semantics (§8.4.1), and the `BulkUpdate` hook-bypass warning (§8.3.5, which makes it a restricted administrative tool rather than a general-purpose update method).

**Next chapters to read:**

- §9 — Privacy Policies — Row-Level Security (the privacy policies injected into every `Query` and mutation call, which are applied at the `EntityRepository` interface layer before any SQL executes)
- §11 — Database Migrations (the Atlas migration workflow that keeps the ent schema in sync with the PostgreSQL schema; the ent implementation depends on this pipeline being run correctly)
- §17 — REST API Conventions (the HTTP layer that translates query parameters into `Filter` structs and `QueryOption` values, wiring the URL-based filter syntax to the DSL documented in §8.5)
