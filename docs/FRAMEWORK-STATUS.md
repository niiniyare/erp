# Awo Framework — Status Reference

> Last updated: 2026-06-30. Reflects the state of the `dev` branch after Phases 0–8.

---

## Table of Contents

1. [Feature Checklist](#1-feature-checklist)
2. [API Reference](#2-api-reference)
3. [Entity System Quick Reference](#3-entity-system-quick-reference)
4. [Security Model](#4-security-model)
5. [Remaining Integration Steps](#5-remaining-integration-steps)
6. [Production Readiness Checklist](#6-production-readiness-checklist)

---

## 1. Feature Checklist

### Core Entity System

| Feature | Status | Location |
|---------|--------|----------|
| `EntityDefinition` central meta-model | ✅ | `framework/def/definition.go` |
| System entity (typed SQL table) | ✅ | `framework/persistence/pgstore/store.go` |
| Custom entity (JSONB `custom_entity_records`) | ✅ | `framework/persistence/pgstore/custom_store.go` |
| Entity registry (`def.Register`, `def.All`, `def.Lookup`) | ✅ | `framework/def/registry.go` |
| `EntityRepository[T]` generic typed interface | ✅ | `framework/persistence/store.go` |
| `TypedRepository[T]` adapter | ✅ | `framework/persistence/store.go` |
| Field types (Data/Currency/Select/Link/…) | ✅ | `framework/def/field.go` |
| `SoftDelete` via `deleted_at` column | ✅ | `framework/persistence/sqlbuilder/builder.go` |
| `Audited` flag with audit log writes | ✅ | `framework/persistence/pgstore/store.go` |
| `NamingSeries` (atomic sequence stamping) | ✅ | `framework/platform/naming/` |
| `EntityValidators` (cross-field validation) | ✅ | `framework/validate/validate.go` |
| `EdgeDef` (eager-load edges) | ✅ | `framework/persistence/pgstore/store.go` |

### Hook System

| Feature | Status | Location |
|---------|--------|----------|
| `HookBeforeValidate` (outside TX) | ✅ | `framework/hooks/runner.go` |
| `HookBeforeSave` (outside TX, before TX begins) | ✅ | `framework/hooks/runner.go` |
| `HookAfterSave` (inside TX) | ✅ | `framework/hooks/runner.go` |
| `HookBeforeDelete` (inside TX) | ✅ | `framework/hooks/runner.go` |
| `HookAfterDelete` (inside TX) | ✅ | `framework/hooks/runner.go` |
| `HookAfterCommit` (outside TX — Temporal triggers) | ✅ | `framework/hooks/runner.go` |
| Hook bitmask filtering by `def.Op` | ✅ | `framework/def/hook.go` |

### Multi-Tenancy

| Feature | Status | Location |
|---------|--------|----------|
| `X-Awo-Tenant` header extraction | ✅ | `framework/bootstrap/bootstrap.go` |
| RLS via `set_tenant_context($1)` | ✅ | `framework/persistence/pgstore/store.go` |
| Tenant status enforcement (PENDING→503, SUSPENDED→402, ARCHIVED→410) | ✅ | `framework/platform/tenant/` |
| `ScopeLevelGlobal / Tenant / Unit` on `EntityDefinition` | ✅ | `framework/def/definition.go` |
| Org Unit `Descendants()` subtree filter on List | ✅ | `framework/api/handler.go` |
| Org scope check on FindByID / Update / Delete | ✅ | `framework/api/handler.go` |
| Hierarchical Create (viewer can create in child unit) | ✅ | `framework/api/handler.go` |
| `privacy.AllowWithinOrgScope` policy constructor | ✅ | `framework/privacy/org_policy.go` |
| `hierarchy_paths` closure table (pgorg.PgTree) | ✅ | `framework/platform/org/pgorg/tree.go` |

### API Layer

| Feature | Status | Location |
|---------|--------|----------|
| `GET /{prefix}/{table}` — paginated list | ✅ | `framework/api/handler.go` |
| `POST /{prefix}/{table}` — create | ✅ | `framework/api/handler.go` |
| `GET /{prefix}/{table}/:id` — find by ID | ✅ | `framework/api/handler.go` |
| `PATCH /{prefix}/{table}/:id` — partial update | ✅ | `framework/api/handler.go` |
| `PUT /{prefix}/{table}/:id` — update alias | ✅ | `framework/api/handler.go` |
| `DELETE /{prefix}/{table}/:id` — delete | ✅ | `framework/api/handler.go` |
| `POST /{prefix}/{table}/:id/{action}` — custom action | ✅ | `framework/api/action.go` |
| Standard response envelope `{"data":…,"meta":…}` | ✅ | `framework/api/response.go` |
| Standard error envelope `{"error":{"code":…,"fields":…}}` | ✅ | `framework/api/response.go` |
| HTTP 422 field-level validation errors | ✅ | `framework/api/handler.go` |
| HTTP 202 for workflow-triggered actions | ✅ | `framework/api/action.go` |
| Pagination (`limit`, `offset`, `order_by`, `dir`) | ✅ | `framework/api/handler.go` |
| Full-text search (`q` param → `ILIKE` on searchable fields) | ✅ | `framework/persistence/sqlbuilder/builder.go` |

### Privacy / Permissions

| Feature | Status | Location |
|---------|--------|----------|
| `PolicyFunc` chain with fail-closed semantics | ✅ | `framework/privacy/enforcer.go` |
| `AllowAll`, `DenyAll`, `AllowSystem` built-ins | ✅ | `framework/def/privacy.go` |
| `OwnerOnly`, `OwnerOnlyUnless`, `RequireRole` | ✅ | `framework/def/privacy.go` |
| `AllowWithinOrgScope(tree)` | ✅ | `framework/privacy/org_policy.go` |
| `def.OrgScoped` interface on records | ✅ | `framework/def/privacy.go` + `framework/persistence/pgstore/record.go` |

### Filter DSL

| Feature | Status | Location |
|---------|--------|----------|
| `Eq / Neq / Gt / Gte / Lt / Lte` | ✅ | `framework/filter/filter.go` |
| `Between`, `In`, `NotIn` | ✅ | `framework/filter/filter.go` |
| `Contains`, `StartsWith`, `EndsWith`, `ILike` | ✅ | `framework/filter/filter.go` |
| `IsNull`, `IsNotNull` | ✅ | `framework/filter/filter.go` |
| `And`, `Or`, `Not`, `None` | ✅ | `framework/filter/filter.go` |
| `JSONPath`, `JSONContains` (JSONB predicates) | ✅ | `framework/filter/filter.go` |
| `Walk` traversal + `Fields()` column list | ✅ | `framework/filter/walk.go` |
| `filterToSQL` with index-preserving `startIdx` | ✅ | `framework/persistence/sqlbuilder/filter_sql.go` |
| `pgIdent` safety (injection guard on all column names) | ✅ | `framework/persistence/sqlbuilder/builder.go` |

### Validation

| Feature | Status | Location |
|---------|--------|----------|
| `Required`, `MaxLength`, `MinVal / MaxVal` pipeline | ✅ | `framework/validate/validate.go` |
| `Options` check (Select field values) | ✅ | `framework/validate/validate.go` |
| `Email()`, `Phone()` (E.164), `URL()` | ✅ | `framework/validate/validate.go` |
| `KRAPin()`, `NHIF()`, `Regex()`, `MinLen()` | ✅ | `framework/validate/validate.go` |
| `Currency()` — rejects `float64`, accepts `decimal.Decimal` | ✅ | `framework/validate/validate.go` |
| `CurrencyRange(min, max)` — exact decimal comparison | ✅ | `framework/validate/validate.go` |
| Entity-level cross-field validators (run after per-field pass) | ✅ | `framework/validate/validate.go` |

### SQL Builder

| Feature | Status | Location |
|---------|--------|----------|
| `SelectOne` — single record by PK | ✅ | `framework/persistence/sqlbuilder/builder.go` |
| `SelectList` — paginated with `COUNT(*) OVER()` | ✅ | `framework/persistence/sqlbuilder/builder.go` |
| `Insert` / `BulkInsert` | ✅ | `framework/persistence/sqlbuilder/builder.go` |
| `Update` (mutable fields only) | ✅ | `framework/persistence/sqlbuilder/builder.go` |
| `BulkUpdate` — sorted SET clause, filter WHERE, soft-delete guard | ✅ | `framework/persistence/sqlbuilder/builder.go` |
| `SoftDelete` / `HardDelete` | ✅ | `framework/persistence/sqlbuilder/builder.go` |
| `Count`, `Exists` | ✅ | `framework/persistence/sqlbuilder/builder.go` |
| `OrgUnitIDs → ANY($n)` injection | ✅ | `framework/persistence/sqlbuilder/builder.go` |

### Workflow (Temporal)

| Feature | Status | Location |
|---------|--------|----------|
| `WorkflowTrigger` declared on `EntityDefinition` | ✅ | `framework/def/definition.go` |
| Auto-bind triggers → `HookAfterCommit` at bootstrap | ✅ | `framework/bootstrap/bootstrap.go` |
| `workflow.Executor` — starts workflows via Temporal client | ✅ | `framework/workflow/executor.go` |
| `RetryQueue` (Redis FIFO, 10-attempt cap) | ✅ | `framework/workflow/retry.go` |
| `Saga` — LIFO compensation for Temporal workflows | ✅ | `framework/workflow/saga.go` |
| `ActivitySaga` — LIFO compensation for plain Go | ✅ | `framework/workflow/saga.go` |
| Workflow ID convention `{tenant}.{entity}.{id}.{event}` | ✅ | Documented in CLAUDE.md |

### SDUI (amis)

| Feature | Status | Location |
|---------|--------|----------|
| SDUI schema registration and serving | ✅ | `framework/sdui/` |
| Redis schema cache (5-min TTL) | ✅ | `framework/sdui/` |
| Permission-gated elements (absent, not disabled) | ✅ | `framework/sdui/` |
| `PageBuilderSet` per-view overrides | ✅ | `framework/def/definition.go` |
| Navigation tree endpoint | ✅ | `framework/sdui/` |

### NamingSeries

| Feature | Status | Location |
|---------|--------|----------|
| `{YYYY}`, `{YY}`, `{MM}`, `{DD}` date tokens | ✅ | `framework/platform/naming/` |
| `{SEQ:n}` zero-padded atomic counter | ✅ | `framework/platform/naming/` |
| Period-key based reset (yearly / monthly / never) | ✅ | `framework/platform/naming/` |
| Tenant-overridable prefix | ✅ | `framework/platform/naming/` |

---

## 2. API Reference

### Base URL

```
/api/{table_name}
```

All routes require tenant identification via `X-Awo-Tenant: {tenant-uuid}` header.

### CRUD Endpoints

| Method | Path | Description | Success |
|--------|------|-------------|---------|
| `GET` | `/api/{table}` | List records (paginated) | 200 |
| `POST` | `/api/{table}` | Create record | 201 |
| `GET` | `/api/{table}/:id` | Get record by UUID | 200 |
| `PATCH` | `/api/{table}/:id` | Partial update | 200 |
| `PUT` | `/api/{table}/:id` | Full update (alias for PATCH) | 200 |
| `DELETE` | `/api/{table}/:id` | Delete record | 204 |
| `POST` | `/api/{table}/:id/{action}` | Custom action | 200 or 202 |

### Query Parameters (List)

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `q` | string | — | Full-text search across `IsSearchable` fields |
| `limit` | int | 50 | Max records per page (hard cap 500) |
| `offset` | int | 0 | Zero-based row offset |
| `order_by` | string | `created_at` | Column to sort by |
| `dir` | `asc`/`desc` | `desc` | Sort direction |

### Response Envelopes

**Success — single record:**
```json
{
  "data": { "id": "...", "created_at": "...", ... },
  "meta": {}
}
```

**Success — list:**
```json
{
  "data": [ { "id": "...", ... } ],
  "meta": {
    "total": 142,
    "limit": 50,
    "offset": 0
  }
}
```

**Created (201):**
```json
{
  "data": { "id": "...", ... },
  "meta": {}
}
```

**Workflow-triggered action (202):**
```json
{
  "data": { "message": "Submitted" },
  "meta": { "workflow_id": "tenant.invoice.abc.on_submit" }
}
```

**Validation error (422):**
```json
{
  "error": {
    "code": "validation_error",
    "message": "Validation failed",
    "fields": {
      "email": "must be a valid email address",
      "total_kes": "currency value must not be a floating-point number"
    }
  }
}
```

**Business error (400/409/…):**
```json
{
  "error": {
    "code": "duplicate",
    "message": "Record already exists"
  }
}
```

### SDUI Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/sdui/nav` | Navigation tree (permission-filtered) |
| `GET` | `/api/sdui/{entity}` | CRUD list + detail page schema |
| `GET` | `/api/sdui/{entity}/form` | Create / edit form schema |

---

## 3. Entity System Quick Reference

### Defining an Entity

```go
var InvoiceDefinition = def.EntityDefinition{
    Name:        "invoice",
    Label:       "Invoice",
    LabelPlural: "Invoices",
    Module:      "Finance",
    OrgScope:    org.ScopeLevelUnit,  // rows carry tenant_id + org_unit_id
    SoftDelete:  true,
    Audited:     true,

    Fields: []*def.FieldDef{
        {Name: "number",      Type: def.FieldTypeData,     IsRequired: true, IsImmutable: true},
        {Name: "customer_id", Type: def.FieldTypeLink,     IsRequired: true, LinkTarget: "customer"},
        {Name: "status",      Type: def.FieldTypeSelect,   Options: []string{"Draft","Submitted","Paid","Cancelled"}, Default: "Draft"},
        {Name: "total_kes",   Type: def.FieldTypeCurrency, IsRequired: true},
        {Name: "notes",       Type: def.FieldTypeLongText},
    },

    NamingSeries: &def.NamingSeriesDef{
        Field:  "number",
        Format: "INV-{YYYY}-{SEQ:5}",
    },

    Policies: []def.PolicyDef{
        def.Policy(def.OpAll,    def.AllowSystem),
        def.Policy(def.OpAll,    privacy.AllowWithinOrgScope(myOrgTree)),
        def.Policy(def.OpCreate, def.RequireRole("finance_ap", "tenant_admin")),
        def.Policy(def.OpRead,   def.RequireRole("finance_viewer", "tenant_admin")),
    },

    WorkflowTriggers: []def.WorkflowTrigger{
        {
            Ops:          def.OpSubmit,
            WorkflowType: "InvoiceSubmissionWorkflow",
            TaskQueue:    "finance.invoice",
            IDFunc: func(m *def.Mutation) string {
                return fmt.Sprintf("%s.invoice.%s.on_submit", m.TenantID, m.After.ID())
            },
        },
    },

    Actions: []def.ActionDef{
        {Name: "submit", Label: "Submit", Fn: SubmitInvoiceAction},
    },
}

func init() {
    def.Register(&InvoiceDefinition)
}
```

### Org Scope Levels

| Level | Columns | Access Rule |
|-------|---------|-------------|
| `ScopeLevelGlobal` | none | Shared across all tenants |
| `ScopeLevelTenant` | `tenant_id` | Visible to all org units in the tenant |
| `ScopeLevelUnit` | `tenant_id` + `org_unit_id` | Viewer's unit must be ancestor-or-equal of record's unit |

### Hook Timing

```
Request arrives
    │
    ├─ HookBeforeValidate   (outside TX — normalize input, compute derivations)
    ├─ validate.Run()       (field + entity validators)
    │
    ├─ [DB transaction begins]
    │     ├─ HookBeforeSave  (inside TX — final business rule checks)
    │     ├─ persistence op  (INSERT / UPDATE / DELETE)
    │     ├─ HookAfterSave   (inside TX — cascade writes, audit, sequences)
    │     └─ [TX commits]
    │
    └─ HookAfterCommit      (outside TX — Temporal triggers, outbox, notifications)
                             ⚠ TX already committed; failures do NOT rollback the save.
```

### Policy Chain Semantics

Policies run in declaration order. First definitive result wins:

- `ErrAllow` → grant immediately, stop chain
- `ErrDeny` → deny immediately, stop chain
- `ErrSkip` / `nil` → continue to next policy
- any other error → treated as `ErrDeny`, logged

**Fail-closed**: if chain exhausts without `ErrAllow`, access is denied.

---

## 4. Security Model

### Tenant Isolation (Layers)

```
Layer 1 — PostgreSQL RLS
    Every tenant-scoped table has FORCE ROW LEVEL SECURITY.
    set_tenant_context($tenant_id) sets app.current_tenant_id per transaction.
    Policy: USING (tenant_id = current_tenant_id())
    Cannot be bypassed from application code.

Layer 2 — Tenant Status Middleware
    PENDING  → 503 Service Unavailable + Retry-After: 60
    SUSPENDED → 402 Payment Required
    ARCHIVED  → 410 Gone

Layer 3 — Org Unit Scope (ScopeLevelUnit entities)
    List:          opts.OrgUnitIDs = Descendants(viewer.OrgUnitID)
                   → org_unit_id = ANY($n) in WHERE clause
    FindByID:      assertOrgScope checks viewer is ancestor-or-equal
    Create:        resolveOrgUnitID validates target is in viewer's subtree
    Update/Delete: assertOrgScope on existing record before enforcer
```

### Input Validation (Layers)

```
Layer 1 — pgIdent   Column names pass through [a-zA-Z0-9_] allowlist before
                    interpolation into SQL (prevents SQL injection).

Layer 2 — BulkUpdate field allowlist
                    Column names validated against EntityDefinition.Fields
                    before reaching pgIdent.

Layer 3 — Filter DSL
                    All values are $n parameterised. No string interpolation.

Layer 4 — Currency fields
                    validate.Currency() rejects float32/float64.
                    decimal.Decimal used throughout for exact arithmetic.
```

### Privacy Policy Recommendations

For any unit-scoped entity, chain:

```go
Policies: []def.PolicyDef{
    def.Policy(def.OpAll, def.AllowSystem),           // system callers bypass
    def.Policy(def.OpAll, privacy.AllowWithinOrgScope(orgTree)), // unit fence
    def.Policy(def.OpCreate, def.RequireRole("...")),  // role gate for writes
    def.Policy(def.OpRead,   def.RequireRole("...")),
},
```

`AllowWithinOrgScope` handles:
- `nil` record (list-level) → `ErrSkip` (list already filtered by handler)
- system callers → `ErrAllow`
- tenant-wide viewer (`OrgUnitID == Nil`) → `ErrAllow`
- viewer's unit is ancestor-or-equal → `ErrAllow`
- otherwise → `ErrDeny`

---

## 5. Remaining Integration Steps

These are wiring tasks for the host application — the framework packages are complete.

### 5.1 Supply a Real `ViewerFromCtx`

The default `anonymousViewer` in bootstrap is **for development only**. Replace it:

```go
bootstrap.Mount(app, bootstrap.Options{
    Pool: pool,
    ViewerFn: func(c *fiber.Ctx) (def.ViewerContext, error) {
        // 1. Read session token from cookie / Authorization header
        // 2. Look up session in Redis → get userID, tenantID, orgUnitID, roles
        // 3. Return your ViewerContext implementation
        return mySessionViewer(c)
    },
})
```

### 5.2 Supply a `TenantResolver` (if using slugs)

```go
bootstrap.Mount(app, bootstrap.Options{
    TenantResolver: func(ctx context.Context, slugOrID string) (uuid.UUID, error) {
        // Look up tenants table by slug
        return tenantRepo.FindIDBySlug(ctx, slugOrID)
    },
})
```

### 5.3 Wire Temporal Client

```go
tc, _ := tclient.Dial(tclient.Options{HostPort: cfg.TemporalHost})
bootstrap.Mount(app, bootstrap.Options{
    TemporalClient: tc,
})
// bootstrap.Mount automatically binds WorkflowTriggers → HookAfterCommit
```

### 5.4 Wire Workflow Retry Queue Drainer

The `RetryQueue` enqueues failed Temporal trigger starts. Run a background drainer:

```go
rq := workflow.NewRedisRetryQueue(redisClient, "awo:wf:retry")
// Replace workflow.NewExecutor if you want retry-queue integration:
exec := workflow.NewExecutorWithRetry(temporalClient, rq)

// Drain periodically (e.g. every 30s):
go func() {
    ticker := time.NewTicker(30 * time.Second)
    for range ticker.C {
        _ = rq.Drain(ctx, func(ctx context.Context, item workflow.RetryItem) error {
            return exec.StartFromRetryItem(ctx, item)
        })
    }
}()
```

### 5.5 Add Tenant-Status Middleware

Mount before the ViewerFn runs:

```go
app.Use(tenantStatusMiddleware(tenantStatusSvc))
```

Where `tenantStatusMiddleware` calls `tenant.StatusForID(ctx, tenantID)` and returns the appropriate HTTP status.

### 5.6 Add Rate Limiting (Redis sliding window)

```go
app.Use(ratelimit.New(ratelimit.Config{
    KeyGenerator: func(c *fiber.Ctx) string {
        tenantID := c.Get("X-Awo-Tenant")
        userID   := c.Locals("actor_id").(string)
        return "rl:" + tenantID + ":" + userID
    },
    Max:        100,
    Expiration: time.Minute,
    // Use Redis store for distributed rate limiting
    Storage: redisStorage,
}))
```

### 5.7 Provide Application `OrgTree`

The default `PgTree` is production-ready, but consider a caching wrapper for read-heavy workloads:

```go
bootstrap.Mount(app, bootstrap.Options{
    OrgTree: &cachedOrgTree{
        inner: pgorg.New(pool),
        cache: redisClient,
        ttl:   5 * time.Minute,
    },
})
```

---

## 6. Production Readiness Checklist

### Security ✅ / ⚠ / ❌

| Item | Status | Notes |
|------|--------|-------|
| PostgreSQL RLS active on all tenant tables | ⚠ | Verified for system tables; custom entities use `custom_entity_records` with RLS. Check each new migration includes the three RLS statements. |
| `set_tenant_context()` called before every query | ✅ | `pgstore.TenantStoreAdapter.WithTx` and `ForEntity` both call it. |
| PgBouncer in **transaction mode** | ⚠ | `set_tenant_context` uses `TRUE` (transaction-local). Session mode breaks this silently. Verify PgBouncer config. |
| `X-Stack-Trace` never exposed to clients | ✅ | `fiberErr` in handler maps to safe messages. `slog.Error` logs the real error. |
| Currency fields use `decimal.Decimal` | ✅ | `validate.Currency()` rejects float; pgstore scans to `decimal.Decimal`. |
| SQL injection via column names | ✅ | `pgIdent` guard on all dynamic column names. `BulkUpdate` has field allowlist. |
| Org unit cross-unit data leakage | ✅ | `assertOrgScope` on all read/write paths. |
| Rate limiting | ❌ | Framework provides hook points; host app must wire Redis sliding window. |
| CORS locked to known origins | ❌ | Host app must configure; framework does not set CORS. |

### Operational

| Item | Status | Notes |
|------|--------|-------|
| `GET /health/live` returns 200 if process alive | ❌ | Not in framework; host app registers. |
| `GET /health/ready` checks Pool + Redis | ❌ | Not in framework; host app registers. |
| Structured logging with `request_id`, `tenant_id`, `user_id` | ⚠ | Framework logs errors via `slog`. Request-level middleware logging is host-app responsibility. |
| Prometheus metrics (`http_request_duration_seconds`, etc.) | ❌ | Not yet instrumented. |
| Temporal worker runs as a separate process / goroutine | ⚠ | Not in framework bootstrap. Host app starts `worker.New(...)`. |
| Workflow retry queue drainer is running | ❌ | See §5.4 — host app must start drainer goroutine. |
| PgBouncer transaction mode | ⚠ | Required for RLS correctness. |
| `CREATE INDEX CONCURRENTLY` used for production migrations | ✅ | Documented in CLAUDE.md; verify in every new migration. |

### Testing

| Item | Status | Notes |
|------|--------|-------|
| Integration tests hit real PostgreSQL | ✅ | Mandated by feedback memory (no mocking the DB). |
| Unit tests for validators | ✅ | `framework/validate/*_test.go` |
| Unit tests for filter DSL | ✅ | `framework/filter/*_test.go` |
| Unit tests for sqlbuilder | ✅ | `framework/persistence/sqlbuilder/*_test.go` |
| Unit tests for org scope (`assertOrgScope`) | ✅ | `framework/api/org_scope_test.go` |
| Unit tests for `AllowWithinOrgScope` policy | ✅ | `framework/privacy/org_policy_test.go` |
| Unit tests for Saga / RetryQueue | ✅ | `framework/workflow/*_test.go` |
| End-to-end API tests (Fiber test client) | ❌ | Not yet written; recommended before first deploy. |

### Before First Production Deploy

1. **Replace `anonymousViewer`** with a real session-backed `ViewerFromCtx`.
2. **Verify PgBouncer** is in transaction mode.
3. **Run all migrations** via `cmd/migrate` against the production database.
4. **Set all required env vars**: `DATABASE_URL`, `REDIS_URL`, `TEMPORAL_HOST`, `JWT_SECRET`, `PORT`.
5. **Wire health checks** at `/health/live` and `/health/ready`.
6. **Start the retry queue drainer** goroutine alongside the API server.
7. **Start a Temporal worker** process (separate from the API process) registered with all workflow and activity types.
8. **Lock CORS** to your tenant subdomain origins.
9. **Configure rate limiting** with Redis storage.
10. **Verify `FORCE ROW LEVEL SECURITY`** on every migration's new table via `\d+ table_name` in psql.
