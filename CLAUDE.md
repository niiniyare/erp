# Awo Framework — CLAUDE.md

Awo is a Go-native multi-tenant ERP framework. Tenancy is structural, not optional — every abstraction assumes one process serves many isolated tenants simultaneously.

---

## CRITICAL RULES

- **Never run `go build`, `go run`, `go vet`, `go test`** — tell user to run them instead
- **Never auto-migrate** — all schema changes go through reviewed `.up.sql` / `.down.sql` files
- **Never raw SQL in business logic** — use `EntityRepository` interface and Filter DSL
- **Never `WHERE tenant_id = ?` in app code** — RLS enforces this at DB level automatically
- **Never store money as float** — always `Currency` field type → `numeric(20,4)` → `decimal.Decimal`
- **Never call `registry.RegisterCustomForTenant` from a request handler** — races with concurrent reads
- **Never lazy-load edges** — explicitly request related data via `QueryOption`s
- **Never update amis SDK version** without a full compatibility audit of all page builder output
- **Never bypass `set_tenant_context()`** — it is the single RLS enforcement point
- **Never expose internal stack traces to clients** — structured error envelope only
- **Never hard-code secrets** — environment variables or secret manager (Vault in prod)
- **Never write custom React/JS for standard ERP views** — amis JSON schemas only
- **Never write I/O inside Temporal workflow functions** — I/O goes in activities only
- **Never write custom CRUD route handlers** — auto-generated from EntityDefinition
- **Never use `time.Now()` or `time.Sleep()` in workflow code** — use `workflow.Now()` / `workflow.Sleep()`
- **Never use ORM types directly** — all persistence through `EntityRepository` interface
- **If a suggestion violates any rule above, state the violation explicitly and explain the trade-off before proceeding**

---

## Project Layout

```
cmd/
  server/           ← API + Temporal worker entrypoint (main.go)
  migrate/          ← Migration runner (separate process, CI-safe)
internal/
  platform/         ← Built-in platform modules (tenant, iam, flags, settings, audit, metadata, registry)
  core/             ← Business domain modules
  shared/           ← Shared utilities (errors, pagination, etc.)
  config/           ← Typed config struct + env loader
db/
  migration/        ← golang-migrate .up.sql / .down.sql pairs
web/
  pages/            ← index.html sidebar + amis embed
  schemas/pages/    ← amis JSON schema files
  sdk/              ← Pinned amis SDK (sdk.js, sdk.css, charts)
framework/
  definition/       ← EntityDefinition, FieldDef, EdgeDef, HookDef, PolicyFunc types
```

---

## The EntityDefinition — Central Primitive

One `definition.Register(&MyEntityDef)` call drives **5 subsystems simultaneously**:

1. **Persistence routing** — SQL (system entity) vs JSONB (custom entity)
2. **API generation** — auto-generates CRUD + action Fiber routes
3. **UI generation** — builds amis JSON page schemas
4. **Permission evaluation** — compiles Casbin policies
5. **Workflow triggering** — binds Temporal workflow starts to lifecycle events

Actual package: `awo.so/framework/definition`. Type: `definition.EntityDefinition`. Builder: `definition.Field("name").OfType(definition.FieldTypeData)`. Registration: `definition.Register(&def)` (package-level function, not method on registry).

```go
var InvoiceDefinition = entity.SystemDefinition{
    Name:        "invoice",        // stable: used in URLs, Redis keys, Temporal IDs, Casbin policies
    Module:      "finance",
    Label:       "Invoice",
    LabelPlural: "Invoices",
    Fields: []entity.FieldDef{
        {Name: "number",     Type: entity.FieldNamingSeries, Series: "INV-{YYYY}-{SEQ:5}"},
        {Name: "customer",   Type: entity.FieldLink, LinkTarget: "customer", Required: true},
        {Name: "status",     Type: entity.FieldSelect, Options: []string{"Draft","Submitted","Paid","Cancelled"}, Default: "Draft"},
        {Name: "total_kes",  Type: entity.FieldCurrency, Required: true},
        {Name: "notes",      Type: entity.FieldLongText},
    },
    Edges: []entity.EdgeDef{
        {Name: "lines", Target: "invoice_line", Type: entity.EdgeOneToMany, CascadeDelete: true},
    },
    Hooks: entity.HookSet{
        BeforeCreate: []entity.BeforeCreateHook{&InvoiceValidator{}},
        AfterCreate:  []entity.AfterCreateHook{&InvoiceNumberAssigner{}},
    },
    Permissions: entity.PermissionSet{
        Create: []string{"role:finance.accounts_payable", "role:tenant.admin"},
        Read:   []string{"role:finance.viewer", "role:tenant.admin"},
        Write:  []string{"role:finance.accounts_payable", "role:tenant.admin"},
        Delete: []string{"role:tenant.admin"},
    },
    WorkflowTriggers: []entity.WorkflowTrigger{
        {
            On:         entity.EventOnSubmit,
            WorkflowFn: "InvoiceSubmissionWorkflow",
            TaskQueue:  "finance.invoice.submit",
            InputBuilder: func(rec *entity.EntityRecord, ctx entity.TriggerContext) (any, error) {
                return finance.InvoiceSubmissionInput{TenantID: rec.TenantID, InvoiceID: rec.ID}, nil
            },
        },
    },
}

func init() {
    definition.Register(&InvoiceDefinition)
}
```

---

## Two Entity Types

### System Entity — Go struct, typed SQL columns

Use when: financial integrity required, inventory accuracy required, IAM data, high-frequency writes (>hundreds/sec).

These **must always** be system entities (non-negotiable, enforced by EntityRegistry validator):

| Entity | Reason |
|---|---|
| LedgerEntry | Double-entry accounting; SQL numeric constraints |
| StockMove | Inventory accounting; SQL quantity constraints |
| Payment | Financial transaction; SQL constraints + audit triggers |
| User | IAM data; JSONB corruption risk |
| Tenant | Platform identity; accessible before per-tenant schemas load |
| JournalEntry | Double-entry; debit/credit balance enforced at DB level |
| TaxEntry | KRA eTIMS record; regulatory compliance |

### Custom Entity — DB records, JSONB storage

Use when: tenant-specific data, frequently evolving schema, no financial/inventory/IAM participation, low write rate.

Good candidates: site visit records, inspection checklists, survey responses, approval metadata, industry-specific classification fields.

**Escalate custom→system when**: >10M records, fields used in financial calculations requiring precision, FK constraints to system entity PKs needed.

### Entity Naming Convention

Format: `{module}_{noun}` — e.g. `finance_journal`, `inventory_stock_move`, `forecourt_shift`.

**Never rename entity names** — embedded in migration filenames, Temporal workflow IDs (stored for years), Redis cache keys.

---

## Five-Layer Architecture (strict top-down dependencies)

```
UI Layer      → amis JSON schemas served from API
API Layer     → Fiber v2, middleware pipeline, thin route handlers (~50 lines max)
Domain Layer  → EntityDefinition, hooks, validators, permission policies (stateless)
Workflow Layer→ Temporal workflows + activities (async, durable)
Store Layer   → EntityRepository interface → PostgreSQL via pgx
```

No layer may import from a higher layer. Domain layer has zero external dependencies at the interface level.

---

## Startup Order (hard dependency chain)

```
Config load & validate
    ↓
PostgreSQL pool init (ping)
    ↓
Redis client init (ping)
    ↓
EntityRegistry init + system entity registration (all modules via init())
    ↓
Fiber app init + route registration (derived from EntityRegistry)
    ↓
Fiber server start
    ↓ (concurrent)
Temporal worker start
```

- EntityRegistry failure = **fatal** (process exits — cannot serve requests without registered routes)
- Temporal failure = **degraded** (CRUD works; workflow starts fail gracefully)
- Redis failure = **503 on all auth** (session validation requires Redis; security-correct behavior)
- PostgreSQL failure = **hard dependency** (process exits at startup if pool fails; 503 during operation)

---

## Multi-Tenancy

### Tenant Identification (priority order)

1. `X-Tenant-ID` header (UUID) — preferred for API clients
2. `tenant_id` query param — webhooks/legacy only; disable in production
3. Subdomain parsing — browser access (`bo.`, `portal.`, `app.`, `api.` prefixes handled)

### RLS Enforcement

Every tenant-scoped table has `FORCE ROW LEVEL SECURITY`:

```sql
CREATE POLICY tenant_isolation ON invoice
    USING (tenant_id = current_tenant_id());
```

`current_tenant_id()` reads `app.current_tenant_id` PostgreSQL transaction-local var, set by:

```go
store.SetTenantContextFromCtx(ctx)  // calls set_tenant_context($1) stored procedure
```

`set_tenant_context()` validates tenant exists + `status = 'ACTIVE'`, then `set_config('app.current_tenant_id', $1, TRUE)`. `TRUE` = transaction-local, resets on COMMIT/ROLLBACK automatically.

**Requires PgBouncer in transaction mode** — session mode breaks the transaction-local reset.

### Global Tables (no RLS, read-only for app role)

`tenants`, `audit_log`, `timezones`, `currencies`, `countries`, `paye_bands`, `platform_admins`

### Tenant Status Machine

```
PENDING → ACTIVE → SUSPENDED → ACTIVE    (payment resolved)
PENDING → ARCHIVED                        (abandoned)
ACTIVE  → ARCHIVED                        (account deletion)
SUSPENDED → ARCHIVED                      (grace period expired)
```

ARCHIVED = terminal. HTTP responses: PENDING→503 + Retry-After:60, SUSPENDED→402, ARCHIVED→410.

---

## EntityRecord Lifecycle (every mutation)

```
ASSEMBLE → before_validate → VALIDATE → AUTHORIZE → before_save
         → [TX begins] → PERSIST → after_save → [TX commits]
         → Temporal workflow start (OUTSIDE TX)
```

| Hook | Inside TX? | Abort via |
|---|---|---|
| `before_validate` | No | `ValidationError` |
| `before_save` | No (TX starts after) | `BusinessError` |
| `after_save` | **Yes** | error → rollback |

Temporal `StartWorkflow` is outside the transaction. If TX commits but Temporal call fails → failure recorded in retry queue for re-attempt. Entity record is saved with workflow ID pre-set.

Hook execution order within each stage = declaration order in `HookSet`.

---

## EntityRepository Interface

```go
type EntityRepository[T Entity] interface {
    // Reads
    Get(ctx context.Context, id uuid.UUID) (T, error)
    Query(ctx context.Context, f Filter, opts ...QueryOption) ([]T, PageInfo, error)
    Exists(ctx context.Context, f Filter) (bool, error)
    Count(ctx context.Context, f Filter) (int, error)
    Aggregate(ctx context.Context, f Filter, spec AggregateSpec) (AggregateResult, error)
    // Writes
    Create(ctx context.Context, input CreateInput) (T, error)
    Update(ctx context.Context, id uuid.UUID, input UpdateInput) (T, error)
    Delete(ctx context.Context, id uuid.UUID) error
    BulkCreate(ctx context.Context, inputs []CreateInput) ([]T, error)
    BulkUpdate(ctx context.Context, f Filter, patch Patch) (int, error)
    // Transactions
    WithTx(ctx context.Context, fn func(ctx context.Context, repo EntityRepository[T]) error) error
}
```

Interface deliberately excludes: raw SQL, ORM types, connection management, lazy loading, schema-specific predicates.

---

## Field Types Reference

### Scalar
| Type | PostgreSQL | Notes |
|---|---|---|
| `Data` | `varchar(n)` | Short strings, indexed. `Searchable()` → GIN trgm index |
| `SmallText` | `varchar(1024)` | Medium prose, no B-tree index |
| `LongText` | `text` | Free-form, no index. Use tsvector for FTS |
| `Int` | `bigint` | Counts, quantities, sequences. Never money |
| `Float` | `double precision` | Scientific/percentages only. **Never money** |
| `Currency` | `numeric(20,4)` | **Only correct type for money**. Go: `decimal.Decimal` |
| `Bool` | `boolean` | — |
| `Date` | `date` | — |
| `DateTime` | `timestamptz` | Stored UTC, serialized as EAT offset ISO 8601 for Kenyan tenants |
| `Time` | `time` | — |

### Structured
| Type | Notes |
|---|---|
| `Select` | Generates SQL `CHECK (col IN (...))`. Options declared in FieldDef |
| `MultiSelect` | Array column |
| `NamingSeries` | Format: `INV-{YYYY}-{SEQ:5}`. Atomic counter, configurable reset. `TenantOverridable: true` for prefix overrides |
| `JSON` | Freeform JSONB. GIN indexed |

### Relational
| Type | Notes |
|---|---|
| `Link` | FK to another entity. Generates FK column + index |
| `LinkList` | Many references |
| `DynamicLink` | Polymorphic — `link_type` + `link_name` pattern |

### Field Constraints
`Required`, `Unique`, `Immutable` (set-once, rejected on update), `Sensitive` (excluded from logs/standard responses), `Searchable` (GIN trgm), `MaxLen`, `Min`, `Max`, custom `Validators []FieldValidator`

---

## Privacy Policies (Row Filtering)

RBAC = operation-level gate. Privacy policies = row-level filter. Both required. Policies inject WHERE predicates into every query — declared once, enforced everywhere, cannot be forgotten.

```go
// Declared on EntityDefinition
Policy: entity.PolicyFunc(func(ctx context.Context) entity.Filter {
    actor := session.ActorFromContext(ctx)
    return filter.Eq("assigned_to", actor.UserID)  // OwnerOnly example
})
```

Built-in policies: `TenantIsolation` (automatic), `OwnerOnly`, `BranchScoped`, `SensitiveFieldMask`

---

## Custom Fields (Runtime Schema Extension)

Tenants add fields to any entity via admin UI — no migration, no redeploy.

- **On system entities**: stored in `custom_fields jsonb` column (every system entity has one; GIN indexed)
- **On custom entities**: additional keys in same JSONB doc

API, SDUI, and Filter DSL treat custom fields identically to system fields.

---

## RBAC (Casbin)

Policy model: `(subject, domain, object, action)`
- Subject: `user:{uuid}` or `role:{name}`
- Domain: `_platform_` or tenant UUID
- Object: entity type name
- Action: `read`, `write`, `create`, `delete`, `submit`, `cancel`, etc.

**System roles** (seeded at tenant bootstrap, cannot delete):

| Role | Scope |
|---|---|
| `role:platform-admin` | Bypasses Casbin entirely. All tenants, platform config |
| `role:tenant.admin` | Full access within one tenant |
| `role:tenant.user` | Standard, customized per tenant |
| `role:api-client` | Machine-to-machine, limited scopes |

Role hierarchy via Casbin `g` assertions — inheritance accumulates permissions upward.

---

## Middleware Pipeline (fixed order)

1. Request ID (from `X-Request-ID` or generated UUID)
2. Structured logging (`slog`, JSON format)
3. Panic recovery
4. CORS (per-tenant subdomain origin validation)
5. Tenant resolution → `set_tenant_context()`
6. Session validation (Redis lookup, expiry check)
7. Rate limiting (Redis sliding window, per-tenant + per-user)

**Order is not configurable at runtime** — changing it requires code change + redeploy. This is intentional (security-sensitive).

---

## SDUI Layer (amis)

amis (Baidu open-source React renderer) interprets JSON schemas. Server generates complete UI descriptions in Go; browser renders without custom JS or build step.

**Ideal for**: CRUD forms/lists (90% of ERP UI), dashboards, approval flows, report pages.
**Not ideal for**: real-time/streaming UI, pixel-perfect branded portals, drag-and-drop, mobile-native.

### Default Page Generation

Every `EntityDefinition` auto-generates: list, create form, edit form, detail view. Cached in Redis (5min TTL, keyed by entity name + version). Custom `PageBuilderSet` overrides any view.

```go
PageBuilders: entity.PageBuilderSet{
    Detail: BuildInvoiceDetailPage,  // only override what you need
},
```

**Permission-gated elements are absent from schema** (not just disabled). Permission check runs at schema-serve time, not just data-fetch time.

### amis SDK

- Pinned in `web/sdk/` — **never auto-update**
- Update only after full compatibility audit of all page builder output
- No `theme("dark")` built-in dark CSS (none exists for `classPrefix: "dark-"`) — override CSS custom property tokens at `html.dark` root instead

---

## Workflow Engine (Temporal)

### Why Temporal

- **Durability**: workflow state persists across crashes; auto-replay from last checkpoint
- **Audit trail**: complete event history (Temporal Web UI), months retention
- **Long-running**: approval chains spanning days (signal-based gates)
- **Saga pattern**: compensating transactions for distributed failures

### Workflow Rules (Determinism)

Workflow code **must be deterministic** — same history = same decisions on replay:
- No `time.Now()` → use `workflow.Now(ctx)`
- No `time.Sleep()` → use `workflow.Sleep(ctx, duration)`
- No `rand` → use `workflow.SideEffect`
- No direct I/O → call activities
- No goroutines → use `workflow.Go`

### Activity Pattern

```go
// Activities struct allows dependency injection for testability
type Activities struct {
    EmailClient notifications.EmailClient
    Repo        entity.EntityRepository
}

func (a *Activities) SendWelcomeEmailActivity(ctx context.Context, input Input) error {
    return a.EmailClient.Send(ctx, ...)
}
```

### Saga Pattern

Each activity registers a compensation. Failure at step N → compensations run N-1 through 1 in reverse order. Framework provides `SagaCompensator` helper.

### Workflow ID Convention

`{tenant-uuid}.{entity-type}.{record-id}.{event}` — e.g. `abc123.invoice.inv456.on_submit`

### Triggering from EntityDefinition

```go
WorkflowTriggers: []entity.WorkflowTrigger{
    {
        On:         entity.EventOnSubmit,
        WorkflowFn: "InvoiceSubmissionWorkflow",
        TaskQueue:  "finance.invoice.submit",
        InputBuilder: func(rec *entity.EntityRecord, tc entity.TriggerContext) (any, error) {
            return InvoiceSubmissionInput{TenantID: rec.TenantID, InvoiceID: rec.ID}, nil
        },
    },
},
```

Temporal start is **outside PostgreSQL transaction**. Failure → retry queue, not rollback.

---

## Platform Modules (Built-in)

7 modules in `internal/platform/` — unconditional, every deployment:

| Module | What |
|---|---|
| **Tenant** | Top-level isolation boundary, lifecycle state machine |
| **IAM** | Users, roles, permissions, sessions, auth |
| **Feature Flags** | Per-tenant on/off switches; Redis-cached; no deploys needed |
| **Settings** | Hierarchical config: system→tenant→branch |
| **Audit Log** | Tamper-evident record of every data change; legal compliance |
| **Metadata** | Custom fields runtime extension |
| **Module Registry** | Tracks installed/activated business modules per tenant |

**No special framework path** — platform modules use identical patterns to business modules (Finance, HR, CRM). Same `EntityDefinition`, `Register()`, `PolicyFunc`, `HookDef`, migrations.

### Module Directory Layout (all modules, platform or business)

```
internal/platform/<module>/
    <module>.go       ← init() calls definition.Register()
    definition.go     ← EntityDefinition variable declarations
    policy.go         ← PolicyFunc implementations
    hooks.go          ← HookDef function implementations
    service.go        ← thin service layer on EntityStore
    handler.go        ← custom HTTP handlers (if needed beyond CRUD)
    migrations/
        YYYYMMDDHHMMSS_create_<table>.up.sql
        YYYYMMDDHHMMSS_create_<table>.down.sql
```

---

## Database Migrations

Tool: `golang-migrate` with PostgreSQL driver. Files: `.up.sql` / `.down.sql` pairs.

**File naming**: Unix timestamp (14 digits) + description slug:
```
20241215143022_create_contact.up.sql
20241215143022_create_contact.down.sql
```

**Never auto-migrate** — silent column drops, no down-migration, no audit trail.

**Single shared schema** = one migration applies to ALL tenants simultaneously. Per-tenant differences go in seed data (provisioning workflow), not schema differences.

### Zero-Downtime Migration Patterns

Adding column: `ADD COLUMN ... DEFAULT NULL` first, then backfill, then add constraint.
Adding index: `CREATE INDEX CONCURRENTLY` — no table lock.
Renaming: add new column, dual-write, backfill, switch reads, drop old. Never single-step rename in production.

### RLS-Aware: Every New Tenant-Scoped Table Needs

```sql
ALTER TABLE my_entity ENABLE ROW LEVEL SECURITY;
ALTER TABLE my_entity FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON my_entity
    USING (tenant_id = current_tenant_id());
```

---

## Redis Usage

| Use | Key Pattern | TTL |
|---|---|---|
| Session tokens | `session:{token}` | Session expiry |
| Feature flag results | `eval:{sha256(flag+tenant+user)}` | 5 min |
| amis page schemas | `page:{entity}:{version}:{tenant}` | 5 min |
| Rate limiting | `rl:{tenant}:{user}:{window}` | Window duration |

Redis is **not** source of truth for any data — PostgreSQL is. Redis failure degrades gracefully for flags/cache; **hard failure for session validation** (correct behavior: cannot authenticate without session store).

---

## Error Handling Pattern

```go
// Repository layer
func parseTenantDBError(err error, op string) error {
    // pgx error codes → domain sentinel or *BusinessError
    // 23505 → unique constraint violation
    // 23514 → check constraint violation
}

// Handler layer
func mapTenantError(err error) *BusinessError {
    // domain sentinel → *BusinessError with HTTP status
}

// shared/errors/http.go
// Use errors.As (not type switch) to unwrap — critical for error chain
```

Always chain errors with `fmt.Errorf("op: %w", err)`. Never type-switch on errors directly — use `errors.As` to unwrap chains.

---

## API Conventions

- Base URL: `/api/v1/entities/{entity-type}`
- Standard CRUD auto-generated: `GET /`, `GET /:id`, `POST /`, `PATCH /:id`, `DELETE /:id`
- Custom actions: `POST /api/v1/entities/{entity-type}/{id}/{action-name}`
- Response envelope: `{"data": {...}, "meta": {...}}` (success), `{"error": {"code": "...", "message": "...", "fields": {...}}}` (error)
- Validation errors: HTTP 422 with field-level messages in amis error format
- Workflow-triggered responses: HTTP 202 Accepted with `{data: {id}, workflow_id: "..."}`
- DateTime serialization: stored UTC, returned as EAT-offset ISO 8601 for Kenyan tenants
- Currency serialization: formatted as `KES 1,234.5600` in Kenya locale contexts

---

## Feature Flags

Evaluation order: system default → tenant override → user override.
Cache: Redis `eval:{sha256(flag+tenant+user)}`, 5min TTL.
Flag changes invalidate cache immediately.
Used to gate: UI elements (absent from schema if flagged off), API behavior, module activation.

---

## Docs Conventions

- `related:` frontmatter uses markdown link format: `"[Title](relative-path.md)"` — never `path:/title:` YAML
- `section:` for Portal 4 files: `04-backend-engineering`
- MDG directory numbers: service=`06-service-layer`, handler=`07-handler-layer`, repo=`05-repository-layer`, wire=`09-wire-registration`

---

## Custom Actions

Standard CRUD = zero boilerplate. Custom actions declared in entity `Actions` field:

```go
Actions: []entity.ActionDef{
    {
        Name:        "submit",
        Method:      entity.ActionMethodPost,
        Label:       "Submit for Approval",
        Permission:  "role:finance.accounts_payable",
        HandlerFunc: SubmitInvoiceAction,
    },
},
```

Auto-generates route: `POST /api/v1/entities/{entity-type}/{id}/{action-name}`

```go
func SubmitInvoiceAction(ctx context.Context, action entity.ActionContext) (*entity.ActionResult, error) {
    // action.Repo — scoped EntityRepository (tenant + permissions already applied)
    // action.RecordID — target record UUID
    // action.Actor — authenticated user from session context
    // TODO: implement business logic
    return &entity.ActionResult{Message: "Submitted"}, nil
}
```

Handler receives pre-resolved, permission-checked context. No manual tenant or auth wiring.

---

## Error Handling — Full Pattern

### Error Types

```go
// ValidationError — field-level, HTTP 422
type ValidationError struct {
    Fields map[string]string  // field name → user-facing message
}

// BusinessError — domain rule violation, HTTP 400/409/etc.
type BusinessError struct {
    Code    string  // machine-readable: "invoice.already_submitted"
    Message string  // user-facing
    Status  int     // HTTP status
}

// NotFoundError — HTTP 404
// PermissionError — HTTP 403
```

### Repository Layer

```go
func parseDBError(err error, op string) error {
    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) {
        switch pgErr.Code {
        case "23505": // unique_violation
            return &BusinessError{Code: "duplicate", Message: "Record already exists", Status: 409}
        case "23514": // check_violation
            return &BusinessError{Code: "constraint", Message: pgErr.Message, Status: 400}
        }
    }
    return fmt.Errorf("%s: %w", op, err)
}
```

### Handler Layer

```go
func mapEntityError(err error) (int, ErrorResponse) {
    var be *BusinessError
    if errors.As(err, &be) {   // errors.As — not type switch — unwraps chains
        return be.Status, ErrorResponse{Code: be.Code, Message: be.Message}
    }
    var ve *ValidationError
    if errors.As(err, &ve) {
        return 422, ErrorResponse{Code: "validation_error", Fields: ve.Fields}
    }
    // Never expose internals
    slog.Error("unhandled error", "err", err)
    return 500, ErrorResponse{Code: "internal_error", Message: "An unexpected error occurred"}
}
```

Always log with context:
```go
slog.Error("operation failed",
    "request_id", requestID,
    "tenant_id",  tenantID,
    "user_id",    userID,
    "err",        err,
)
```

---

## Testing Guidelines

- Unit test hooks, policies, service logic using `EntityRepository` mocks (mock satisfies interface)
- Test workflows with Temporal's `testsuite.WorkflowTestSuite`
- Test activities independently — inject mock dependencies via struct receivers
- Never mock the database in integration tests — use real PostgreSQL (past incident: mock/prod divergence masked broken migration)
- **Never auto-run `go test` or `go vet`** — provide test code, tell user to run
- Place unit tests alongside source: `entity_contact_test.go` next to `entity_contact.go`
- Table-driven tests preferred for validators and hook logic

```go
// Hook unit test pattern
func TestInvoiceValidator_BeforeCreate(t *testing.T) {
    repo := &mockEntityRepository{}  // mock satisfies EntityRepository[Invoice]
    hook := &InvoiceValidator{repo: repo}

    tests := []struct{
        name    string
        record  *entity.EntityRecord
        wantErr bool
    }{
        {"valid invoice", validInvoiceRecord(), false},
        {"missing customer", missingCustomerRecord(), true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := hook.BeforeCreate(context.Background(), tt.record)
            if (err != nil) != tt.wantErr {
                t.Errorf("got err=%v, wantErr=%v", err, tt.wantErr)
            }
        })
    }
}
```

---

## Configuration and Secrets

- All config via environment variables; load into typed struct at startup
- Validate all config fields at startup — fail fast, never silently default to insecure values
- **Never hard-code**: DB passwords, Redis credentials, Temporal namespace tokens, API keys, JWT signing keys
- Production secret management: HashiCorp Vault (or equivalent); inject as env vars at process start
- Config struct carries typed fields (not raw `map[string]string`):

```go
type Config struct {
    Port        int           `env:"PORT,required"`
    DatabaseURL string        `env:"DATABASE_URL,required"`
    RedisURL    string        `env:"REDIS_URL,required"`
    TemporalHost string       `env:"TEMPORAL_HOST,required"`
    LogLevel    slog.Level    `env:"LOG_LEVEL" envDefault:"info"`
    // Sensitive — never log these fields
    JWTSecret   string        `env:"JWT_SECRET,required"`
}
```

Sensitive config fields must never appear in logs or error responses.

---

## Performance

- **Cache stampede prevention**: use Redis `SET NX` + short jitter on TTL for heavily-contested keys
- **No per-tenant connection pools** — single PgBouncer pool for all tenants; isolation via RLS
- **GIN indexes** on all JSONB columns and `Searchable()` fields — generated automatically by framework
- **Avoid N+1 queries** — use `QueryOption`s to explicitly load edges in one query, not per-record loops
- **Batch operations** for bulk writes — `BulkCreate`, `BulkUpdate` instead of looping `Create`/`Update`
- **`CREATE INDEX CONCURRENTLY`** for all production index additions — never lock production tables
- **Page schema cache** (Redis, 5min TTL) — invalidate on permission change or feature flag change, not on every request
- **Feature flag cache** (Redis, 5min TTL, SHA-256 keyed) — evaluate once per request, not per check

---

## Observability

### Structured Logging

All log entries carry: `request_id`, `tenant_id`, `user_id`, `method`, `path`, `duration_ms`.

```go
slog.Info("request completed",
    "request_id",  c.Locals("request_id"),
    "tenant_id",   tenantID,
    "user_id",     userID,
    "method",      c.Method(),
    "path",        c.Path(),
    "status",      c.Response().StatusCode(),
    "duration_ms", time.Since(start).Milliseconds(),
)
```

### Prometheus Metrics (expose at `/metrics`)

Key metrics to instrument:
- `http_request_duration_seconds` (histogram, labels: method, path, status)
- `db_query_duration_seconds` (histogram, labels: entity, operation)
- `workflow_started_total` (counter, labels: workflow_type, tenant)
- `workflow_failed_total` (counter, labels: workflow_type, tenant)
- `cache_hit_total` / `cache_miss_total` (counter, labels: cache_key_type)

### Health Checks

```
GET /health/live   → 200 if process is running (no dependencies checked)
GET /health/ready  → 200 only if PostgreSQL + Redis reachable and EntityRegistry populated
```

Readiness check fails → remove from load balancer rotation. Liveness check fails → restart process.

---

## Code Generation Standards

All generated code must follow:

| Construct | Convention |
|---|---|
| Go structs / types | `PascalCase` |
| Go fields / vars | `camelCase` |
| DB column names | `snake_case` |
| Entity names | `snake_case`, `{module}_{noun}` |
| Hook struct names | `{Entity}{Purpose}Hook` e.g. `InvoiceSubmitGuard` |
| Activity func names | `{Verb}{Noun}Activity` e.g. `SendWelcomeEmailActivity` |
| Workflow func names | `{Entity}{Event}Workflow` e.g. `InvoiceApprovalWorkflow` |

Every function interacting with DB or external services takes `context.Context` as first arg.

Use `// TODO:` comments for all placeholder logic — never leave silent no-ops.

```go
// TODO: validate customer credit limit before allowing invoice creation
```

Include error context in all wraps:
```go
return fmt.Errorf("invoice.BeforeCreate: validate customer: %w", err)
```

---

## Quick Reference: What Goes Where

| Concern | Layer | Location |
|---|---|---|
| Entity shape + fields | Domain | `definition.go` |
| Permission policy | Domain | `policy.go` or `Permissions` in EntityDef |
| Business rules (pre-save) | Domain | `hooks.go` → `before_save` hook |
| Post-save side effects | Domain | `hooks.go` → `after_save` hook (inside TX) |
| HTTP route handler | API | `handler.go` (thin, ~50 lines max) |
| Multi-step async process | Workflow | `workflows/` → Temporal workflow |
| External API calls | Workflow | `workflows/` → Temporal activity |
| Schema change | Store | `migrations/` → `.up.sql` + `.down.sql` |
| UI description | SDUI | `PageBuilderSet` on EntityDefinition |
| Tenant-specific config | Settings module | `tenant_configs` table |
| Runtime field extensions | Metadata module | `CustomFieldDef` entity |
