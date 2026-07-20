# CLAUDE.md — AWO ERP

Technical reference for Claude (and Claude Code) working on this repository. Assumes Go + PostgreSQL familiarity.

**Repository root:** `/data/data/com.termux/files/home/project/erp/`
**Module path:** `awo.so`
**Framework package root:** `awo.so/awo/...`
**Stage:** Pre-v1.0 kernel freeze — validating architecture before module implementations.

---

## Table of Contents

1. [Architecture Overview](#1-architecture-overview)
2. [Repository Layout](#2-repository-layout)
3. [Kernel Packages](#3-kernel-packages)
4. [Key Patterns](#4-key-patterns)
5. [Authorization Model](#5-authorization-model)
6. [Multi-Tenancy](#6-multi-tenancy)
7. [Entity Definitions](#7-entity-definitions)
8. [Migrations](#8-migrations)
9. [Workflow (Temporal)](#9-workflow-temporal)
10. [SDUI (AMIS)](#10-sdui-amis)
11. [Known Architectural Contradictions](#11-known-architectural-contradictions)
12. [Roadmap](#12-roadmap)
13. [Code Guidelines](#13-code-guidelines)
14. [Module Development Checklist](#14-module-development-checklist)

---

## 1. Architecture Overview

**Core thesis:** Multi-tenant, permission-centric ERP kernel in Go. One `EntityDefinition` registration drives five subsystems simultaneously — persistence, API generation, SDUI, authorization, and workflow.

**Stack:**

| Layer | Technology |
|---|---|
| HTTP | Fiber v2 |
| Database | PostgreSQL + Row-Level Security (RLS) |
| Query gen | SQLC |
| DI | Google Wire |
| Authorization | Casbin v2 / CEL expressions |
| Async workflows | Temporal SDK |
| Cache | Redis (go-redis/v8) |
| Web UI | AMIS-based Server-Driven UI (SDUI) |
| Mobile | Flutter (planned) |
| Observability | OpenTelemetry + Prometheus |

**Primary application:** AwoERP (in `awo/` subdirectory)

---

## 2. Repository Layout

```
erp/
├── CLAUDE.md                  ← this file
├── ENTITY_DEFINITION_SPEC.md  ← entity def canonical reference
├── ACTOR_MODEL.md             ← stub, see awo/docs/03-auth/ACTOR_SPEC.md
├── AUTHORIZATION_SPEC.md      ← stub, see awo/docs/03-auth/AUTHORIZATION_SPEC.md
├── SESSION_MODEL.md           ← stub, see awo/docs/03-auth/SESSION_SPEC.md
├── modules/
│   └── finance/               ← Finance module entity definitions
│       ├── definition_bank.go
│       ├── definition_coa.go
│       ├── definition_currency.go
│       ├── definition_invoice.go
│       ├── definition_journal.go
│       ├── definition_payment.go
│       ├── definition_period.go
│       └── definition_tax.go
└── awo/                       ← Framework + ERP app (module: awo.so)
    ├── go.mod
    ├── cmd/server/main.go     ← entry point
    ├── bootstrap/             ← startup sequence
    ├── api/
    │   ├── authz/             ← RequirePermission middleware
    │   ├── handler/           ← entity CRUD handlers
    │   ├── middleware/        ← TenantResolver, RequireAuth, RateLimit
    │   └── router/            ← auto-generated routes from CompiledSchema
    ├── auth/                  ← Session, ViewerContext, Actor
    ├── compiler/              ← EntityDefinition → CompiledSchema
    ├── contrib/
    │   ├── pgx/               ← PostgreSQL EntityRepository impl
    │   └── redis/             ← Redis cache impl
    ├── def/                   ← DSL layer (EntityDefinition, FieldDef, PermissionSet)
    ├── driver/                ← EntityRepository[T] interface
    ├── events/outbox/         ← transactional outbox relay
    ├── filter/                ← query filter predicates
    ├── naming/                ← NamingSeries auto-ID generation
    ├── observability/
    │   ├── health/
    │   └── metrics/
    ├── platform/
    │   ├── audit/
    │   ├── flags/
    │   ├── iam/               ← login, session management, RBAC
    │   ├── metadata/
    │   ├── settings/
    │   └── tenant/            ← tenant lifecycle, context propagation
    ├── registry/              ← runtime entity/action/policy registry
    ├── runtime/               ← entity lifecycle pipeline
    └── docs/
        └── 03-auth/           ← Frozen Tier-0 auth specs
```

---

## 3. Kernel Packages

### `awo/def` — Metadata / DSL Layer

Single source of truth for all entity metadata. Module authors touch only this package when declaring entities.

**Key types:**

```go
// EntityDefinition — implemented by SystemDefinition and CustomDefinition
type EntityDefinition interface {
    EntityName() string
    EntityFields() []FieldDef
    EntityPermissions() PermissionSet
    EntityHooks() HookSet
    EntityActions() []ActionDef
    EntityWorkflows() []WorkflowDef
}

// SystemDefinition — typed PostgreSQL columns; use for financial, IAM, inventory
type SystemDefinition struct { ... }

// CustomDefinition — JSONB in custom_entity_records; use for tenant-specific schemas
type CustomDefinition struct { ... }

// PermissionSet — permission identifiers only; NO role names
type PermissionSet struct {
    Create  []string
    Read    []string
    Update  []string
    Delete  []string
    Actions map[string][]string
}
```

**Register in `init()` only:**

```go
func init() {
    def.Register(&MyEntityDefinition)
}
```

Never call `Register` from a handler or after bootstrap.

### `awo/compiler` — Schema Compiler

Validates EntityDefinitions and compiles them into `CompiledSchema` with `CapabilityGrant` values used by the runtime and authz layer.

**Key output:** `CapabilityGrant` — binds a permission identifier to a compiled enforcement rule.

### `awo/registry` — Runtime Registry

Maintains all registered `SystemDefinition` and `CustomDefinition` entities. Populated during bootstrap via `init()` side effects. Read-only after bootstrap completes.

### `awo/runtime` — Entity Lifecycle

Executes the entity lifecycle pipeline for every mutating operation:

```
BeforeCreate → Validate → Authorize → Persist → AfterCreate
```

Same pattern for Update, Delete. Hooks fire at each stage.

### `awo/driver` — Storage Abstraction

```go
// EntityRepository[T] — generic interface; impl in contrib/pgx
type EntityRepository[T any] interface {
    Create(ctx context.Context, record T) (T, error)
    Get(ctx context.Context, id uuid.UUID) (T, error)
    Update(ctx context.Context, id uuid.UUID, patch T) (T, error)
    Delete(ctx context.Context, id uuid.UUID) error
    Query(ctx context.Context, opts QueryOptions) ([]T, error)
}
```

All DB access goes through this interface. No raw SQL outside drivers.

### `awo/filter` — Query Filtering

Composable filter predicates for queries:

```go
filter.And(
    filter.Eq("status", "active"),
    filter.Gt("amount", 0),
)
```

CEL-based expression evaluation for complex predicates. Type-safety gaps exist (see §11).

### `awo/auth` — Session & ViewerContext

**Session** — stored in Redis (`session:{token}`) and PostgreSQL (durable).

```go
type Session struct {
    Token            string
    UserID           *uuid.UUID
    ServiceAccountID *uuid.UUID
    TenantID         uuid.UUID
    Roles            []string
    ExpiresAt        time.Time
    IssuedAt         time.Time
    DeviceID         string
    IPAddress        string
    RequestID        string
}
```

**ViewerContext** — authorization subject, injected into every request context:

```go
type ViewerContext interface {
    TenantID() uuid.UUID
    UserID() *uuid.UUID
    ServiceAccountID() *uuid.UUID
    Roles() []string
    HasRole(role string) bool
    IsPlatformAdmin() bool  // derived from Roles; never a bool field (ADR-003)
}
```

Retrieve with `auth.ViewerFromContext(ctx)` — panics if middleware was bypassed (fail-fast by design).

### `awo/platform/iam` — Identity & Access Management

Login, session management, role loading. Dual-plane IAM:

- **Platform Plane:** manages tenants, platform admins, billing
- **Tenant Plane:** manages users, OU hierarchy, roles within a tenant

ltree-based OU scoping for hierarchical access delegation.

### `awo/events/outbox` — Transactional Outbox

Domain events written to the outbox table within the same DB transaction as the mutation. Relay polls and publishes asynchronously.

### `awo/naming` — NamingSeries

Auto-generates formatted sequential IDs:

```
INV-{YYYY}-{SEQ:6}   →   INV-2026-000001
PAY-{YYYY}-{SEQ:6}   →   PAY-2026-000001
JE-{YYYY}-{SEQ:5}    →   JE-2026-00001
```

Tenants can override prefixes via `TenantOverridable: true`.

---

## 4. Key Patterns

### HTTP Routes

Auto-generated from `CompiledSchema`:

```
GET    /api/v1/{module}/{resource}               List
GET    /api/v1/{module}/{resource}/:id           Get
POST   /api/v1/{module}/{resource}               Create
PATCH  /api/v1/{module}/{resource}/:id           Update
DELETE /api/v1/{module}/{resource}/:id           Delete
POST   /api/v1/{module}/{resource}/:id/:action   Custom action
```

### Middleware Stack

Applied at the `/api/v1` group in order:

1. `TenantResolver` — extracts tenant from host/header, validates, injects into context
2. `RequireAuth` — validates session token, builds `ViewerContext`, injects via `auth.WithViewer(ctx, viewer)`
3. `RateLimit` — per-tenant/user, Redis-backed

### Dependency Injection (Wire)

All services use Google Wire. No global state.

```go
// Provider declaration
func NewMyService(repo driver.EntityRepository[MyEntity], ...) *MyService { ... }

// Wire provider set
var MyServiceSet = wire.NewSet(NewMyService)
```

Wire generates `wire_gen.go` at build time. Never edit generated files.

### Bootstrap Sequence

```
1. Config validation (DatabaseURL, RedisURL required)
2. PostgreSQL pool init + ping
3. Redis client init + ping
4. EntityRegistry init (all modules via init())
5. Schema compilation → CompiledSchema
6. Fiber app init + route registration
7. HTTP server start
8. Temporal worker start (concurrent; degraded-ok if Temporal unavailable)
```

---

## 5. Authorization Model

**Two-layer architecture (ADR-001, ADR-011). Frozen at v1.0.**

### Layer 1 — Declaration (`awo/def`)

`PermissionSet` on `EntityDefinition`. Permission identifiers only — stable names for capabilities.

```go
Permissions: def.PermissionSet{
    Create: []string{"finance.invoice.create"},
    Read:   []string{"finance.invoice.read"},
    Update: []string{"finance.invoice.update"},
    Delete: []string{"finance.invoice.delete"},
    Actions: map[string][]string{
        "submit": {"finance.invoice.submit"},
        "cancel": {"finance.invoice.cancel"},
    },
},
```

**Critical invariant:** `EntityDefinition` MUST NOT reference role names, JWT claims, RBAC constructs, or any authorization backend detail. These belong exclusively to `PolicyEvaluator`.

### Layer 2 — Enforcement (`awo/auth`)

`PolicyEvaluator` interface. Default: Casbin. Replaceable with OPA, ReBAC, or custom engine without touching entity definitions.

```go
type PolicyEvaluator interface {
    CanPerform(ctx context.Context, viewer ViewerContext, permissionID string) (bool, error)
}
```

### Authorization Flow (per request)

```
Request
  → TenantResolver (inject tenant)
  → RequireAuth (inject ViewerContext)
  → RequirePermission middleware:
      if viewer.IsPlatformAdmin() → bypass (no Casbin call)
      else → PolicyEvaluator.CanPerform(permissionID)
              true  → continue
              false → 403 Forbidden
              error → 500 Internal Server Error
  → Handler
  → RLS (PostgreSQL policies enforce row visibility)
```

### Platform Admin Bypass

`IsPlatformAdmin()` is derived from `Roles` at runtime. Not a stored boolean field (ADR-003). Platform admins skip Casbin entirely — all rows across all tenants are visible.

### PolicyFunc (Row-Level Authorization)

Beyond `PermissionSet`, entities can declare a `PolicyFunc` that injects additional filter predicates into every query, restricting which rows a viewer can see at the application layer (before RLS).

---

## 6. Multi-Tenancy

### Tenant Context Propagation

Every request carries a `TenantID` extracted by `TenantResolver`. Never pass tenant ID as a function argument through business logic — always carry it in `context.Context`.

```go
// Inject (in middleware)
ctx = tenant.WithTenant(ctx, tenantID)

// Retrieve (in service/repo)
tenantID := tenant.FromContext(ctx)  // panics if missing
```

### PostgreSQL RLS

Every `SystemDefinition` table has a `tenant_id` column and a PostgreSQL RLS policy:

```sql
-- Policy on every entity table
CREATE POLICY tenant_isolation ON finance_invoice
    USING (tenant_id = current_tenant_id());

-- current_tenant_id() reads from session variable set per connection
```

The PostgreSQL function `current_tenant_id()` is set via a connection-level `SET LOCAL` before every query execution in `contrib/pgx`.

**RLS is the last line of defense.** Application-layer filters (PolicyFunc) run first. RLS catches anything that leaks through.

### Tenant Isolation Invariant

No query must ever return rows from another tenant, even if the application layer has a bug. RLS enforces this at the database level unconditionally.

---

## 7. Entity Definitions

See `ENTITY_DEFINITION_SPEC.md` for the full field type table and constraint reference.

### Naming Convention

```
{module}_{noun}   →   finance_invoice, iam_user, platform_tenant
```

- Module prefix = Go package name of owning module
- Noun = singular snake_case
- **Never rename** — embedded in migration filenames, Temporal workflow IDs, Redis keys, Casbin policies

### Field Types Reference

| Type | PostgreSQL | Go |
|---|---|---|
| `FieldTypeData` | `varchar(n)` | `string` |
| `FieldTypeCurrency` | `numeric(20,4)` | `decimal.Decimal` |
| `FieldTypeLink` | FK + index | `uuid.UUID` |
| `FieldTypeNamingSeries` | `varchar(100)` | `string` (auto-generated) |
| `FieldTypeSelect` | `varchar` + CHECK | `string` |
| `FieldTypeJSON` | `jsonb` | `map[string]any` |

### Finance Module Entities

| Entity | Purpose |
|---|---|
| `finance_currency` | ISO 4217 currency master |
| `finance_exchange_rate` | Point-in-time forex snapshots (immutable) |
| `finance_chart_of_accounts` | COA master header |
| `finance_account` | GL account in COA hierarchy |
| `finance_bank_account` | Tenant bank account mapped to GL |
| `finance_bank_transaction` | Imported bank statement lines (immutable amount/date) |
| `finance_journal` | Journal master (GL, Sales, Purchase, Bank, Cash, Opening) |
| `finance_journal_entry` | Double-entry header — draft → submitted → posted → reversed |
| `finance_payment` | Payment record — draft → submitted → processed → reconciled |
| `finance_payment_method` | Payment method master mapped to GL account |
| `finance_tax_group` | Tax grouping for VAT, withholding |
| `finance_tax` | Individual tax rate (percentage, fixed, compound) |
| `finance_fiscal_year` | Fiscal year with lifecycle locking |
| `finance_accounting_period` | Monthly/quarterly period — must be open for journal entries |

Double-entry integrity enforced by PostgreSQL triggers. SQLC safe queries only.

---

## 8. Migrations

**Two-tier migration system:**

| Tier | Table | Owns |
|---|---|---|
| Framework | `awo_schema_migrations` | Core tables (tenants, sessions, IAM, RLS policies) |
| Application | `schema_migrations` | Module entity tables, indexes, triggers |

Framework migrations run first at bootstrap. Application migrations run after entity registration.

**Migration file naming:** `{seq}_{description}.sql`
- seq = 6-digit zero-padded integer (e.g., `001002`)
- Keep framework and app sequences in separate namespaces

---

## 9. Workflow (Temporal)

Temporal SDK handles async, multi-step entity lifecycle operations.

**When to use Temporal:**
- Any operation requiring multiple DB writes that must be atomic across time
- Long-running approvals, reconciliation, scheduled jobs
- Operations that need retry/compensation logic

**When NOT to use Temporal:**
- User-facing queries (synchronous reads)
- Simple single-table writes

**Workflow skeleton:**

```go
// Definition (in workflow package)
func (w *InvoiceWorkflow) SubmitInvoice(ctx workflow.Context, invoiceID uuid.UUID) error {
    ao := workflow.ActivityOptions{StartToCloseTimeout: 30 * time.Second}
    ctx = workflow.WithActivityOptions(ctx, ao)

    var a *InvoiceActivities
    if err := workflow.ExecuteActivity(ctx, a.ValidateInvoice, invoiceID).Get(ctx, nil); err != nil {
        return err
    }
    return workflow.ExecuteActivity(ctx, a.PostJournalEntry, invoiceID).Get(ctx, nil)
}

// Call from handler
workflowRun, err := temporal.Client().ExecuteWorkflow(ctx,
    client.StartWorkflowOptions{ID: "invoice-submit-" + invoiceID.String()},
    workflow.InvoiceWorkflow.SubmitInvoice,
    invoiceID,
)
```

Workflow IDs use entity name + record ID to ensure idempotency.

---

## 10. SDUI (AMIS)

AMIS-based Server-Driven UI. Blocks defined in Go, rendered on web/mobile.

- Web UI: `awo/web/pages/index.html` — sidebar + AMIS embed
- Schemas: `awo/web/schemas/pages/*.json`
- SDK: `awo/web/sdk/` (sdk.js, sdk.css, charts)

**Dark mode:** No built-in AMIS dark CSS. Override CSS custom property tokens at root — scale `--colors-neutral-*` vars on `html.dark`. Never override individual `.cxd-*` backgrounds with `!important`.

SDUI blocks are defined in `awo/sdui` package and composed at the framework level from EntityDefinition metadata.

---

## 11. Known Architectural Contradictions

These are unresolved at the time of kernel freeze. Each has a documented decision for v1.0; full resolution is post-freeze work.

### 1. Hook Dependency Injection — Early Lifecycle, Multi-Module Coordination

**Problem:** Hooks fire before Wire-injected dependencies are fully available in multi-module scenarios. Module A's `BeforeCreate` hook may need Module B's service, but Wire builds a single graph — cross-module hook dependencies create circular DI risk.

**v1.0 decision:** Hooks receive only the minimal `HookContext` (actor, tenant, record). Cross-module coordination must go through Temporal workflows (async) or events (outbox). Direct service injection into hooks is prohibited at v1.0.

### 2. Plugin Entity Registration — Avoiding Central Coupling

**Problem:** `def.Register()` uses a global registry populated by `init()`. Pure plugin model (load `.so` at runtime) conflicts with Go's `init()` compile-time guarantee. Dynamic loading without a compile-time registry risks registration races.

**v1.0 decision:** All modules are compiled into the binary. Runtime plugin loading (`.so`) deferred to post-v1.0. `init()` registration is the only supported pattern.

### 3. Row-Visibility Enforcement Layer Arbitration

**Problem:** Two layers enforce row visibility — `PolicyFunc` (application layer) and PostgreSQL RLS. These can produce inconsistent counts (e.g., `Count()` at app layer vs. DB layer), unexpected empty results when only one layer is correctly configured, and double-filtering overhead.

**v1.0 decision:** RLS is authoritative for correctness. `PolicyFunc` is an optional performance optimization (pre-filter before DB round-trip). If `PolicyFunc` is absent, RLS alone is sufficient. Never rely on `PolicyFunc` as a security boundary.

### 4. Type-Safety Gaps in Filter / Expression Evaluation

**Problem:** `filter` package accepts `any` values for predicate operands. CEL evaluation is dynamic. No compile-time guarantee that a filter like `filter.Eq("amount", "string")` against a `numeric(20,4)` column is type-safe. Runtime panics possible.

**v1.0 decision:** Document the gap. Add runtime type assertion in `contrib/pgx` filter translator with descriptive error messages. Compile-time typed filter DSL deferred to post-v1.0.

### 5. Dual-Plane IAM — JWT Key Separation at Runtime

**Problem:** Platform Plane and Tenant Plane use separate JWT signing keys. Key rotation must be coordinated. During rotation, in-flight sessions signed with old keys must still validate. Redis session cache does not store the signing key version, making zero-downtime rotation complex.

**v1.0 decision:** Single key pair at v1.0. Dual-key rotation support deferred. Document as known limitation. Session table stores `issued_at` for future key-version correlation.

### 6. Metadata Compiler — Missing Dependency Graph and Conflict Detection

**Problem:** `awo/compiler` validates individual EntityDefinitions but does not build a cross-entity dependency graph. Circular FK references, orphaned `FieldTypeLink` targets, and conflicting migration sequences are not caught at compile time — they fail at migration run time or at first query.

**v1.0 decision:** Add a compiler report pass before freeze: compilation summary, dependency graph, conflict checks, migration fingerprint. This is a pre-freeze addition requirement. Implement before declaring v1.0 kernel frozen.

---

## 12. Roadmap

```
Current:   Resolve kernel contradictions → finalize architecture
           Add metadata compiler report (§11 item 6)
           ↓
Freeze:    v1.0 kernel — all packages in §3 locked
           ↓
Phase 1:   Finance module as readiness test
           Build using ONLY existing framework primitives
           Any missing primitive → add to framework first, then use
           ↓
Phase 2:   Inventory module
Phase 3:   CRM module
Phase 4:   Procurement + Sales modules
Phase 5:   HR + Assets + Manufacturing modules
           ↓
Future:    Extract awo/ into separate repo (awo.so/awo/...)
           AwoERP depends on awo as external module
           Open-source kernel; ERP is the reference implementation
```

---

## 13. Code Guidelines

### General

- All kernel packages require godoc-quality comments on exported types and functions
- Use Wire for DI; no global state (exception: `def.Register` global registry — by design)
- No raw SQL outside `contrib/pgx` and `contrib/redis`; use SQLC for all generated queries
- Temporal for async workflows; synchronous paths for user-facing queries only
- `errors.As` for error unwrapping — never type switch on errors

### Security

- RLS-first: every entity query passes through PostgreSQL policies
- Never set `SET LOCAL tenant_id` from user-supplied input without validation
- `Sensitive: true` fields must be excluded from logs and standard API responses
- Platform admin bypass is in `authz` middleware only — never replicate this check in business logic

### Testing

- Cover multi-tenant isolation: verify tenant A cannot read tenant B's rows
- Cover RLS enforcement: drop `PolicyFunc`, verify DB-level RLS still blocks cross-tenant reads
- Cover permission evaluation: test 403 response for actors without required permission identifiers
- Use `cache.NoopCounter` and noop implementations for unit test isolation

### Do Not

- Call `def.Register` outside of `init()`
- Rename an entity after its migration has been applied to any environment
- Add role names or RBAC constructs to `EntityDefinition`
- Write raw SQL in handler or service layers
- Mock the database in integration tests (RLS enforcement requires a real PostgreSQL connection)
- Run `go build`, `go run`, `go vet`, or `go test` — tell the user to run these (Termux constraint)

---

## 14. Module Development Checklist

Use this when building a new module (Finance is the Phase 1 reference):

### Define

- [ ] Declare entities using `def.SystemDefinition` (or `CustomDefinition` for tenant-specific)
- [ ] Follow naming convention: `{module}_{noun}` in singular snake_case
- [ ] Set `PermissionSet` with permission identifiers only (`"{module}.{entity}.{operation}"`)
- [ ] Declare `HookSet` for lifecycle extension points
- [ ] Declare `ActionDef` for custom actions with their permission identifiers
- [ ] Declare `WorkflowDef` for any async multi-step operations
- [ ] Register in `init()` via `def.Register(&MyEntityDefinition)`

### Compile & Validate

- [ ] Run `awo/compiler` validation — check for field type errors, duplicate names, circular links
- [ ] Verify compiler report: dependency graph, migration fingerprint, conflict checks

### Persistence

- [ ] Write SQLC schema (`.sql` files for entity tables)
- [ ] Add RLS policy: `USING (tenant_id = current_tenant_id())`
- [ ] Run SQLC codegen; review generated Go types
- [ ] Write migration file in correct sequence under `db/migration/`

### API

- [ ] Implement handlers in Fiber using `awo/api` patterns
- [ ] Register custom action handlers for `ActionDef` actions
- [ ] Verify routes auto-generated correctly from CompiledSchema

### Authorization

- [ ] Confirm `PermissionSet` identifiers are assigned to roles in Casbin policy
- [ ] Test: actor with permission → 200; actor without → 403
- [ ] Test: platform admin → bypass (200 on all operations)

### Multi-Tenant Isolation

- [ ] Test: tenant A creates record; tenant B query returns empty (RLS working)
- [ ] Test: `current_tenant_id()` correctly set per connection in `contrib/pgx`
- [ ] Test: `PolicyFunc` (if declared) does not produce different count than RLS count

### Workflows

- [ ] Implement Temporal workflows for any async multi-step operations
- [ ] Use `{entity}-{action}-{recordID}` as workflow ID for idempotency
- [ ] Test workflow replay correctness

### SDUI

- [ ] Define AMIS block schemas for list, form, detail views
- [ ] Test dark mode rendering (override CSS tokens, not `.cxd-*` classes)

### Documentation

- [ ] Write module overview in `awo/docs/`
- [ ] Document any new framework primitives added during module development
- [ ] Update this checklist if new patterns emerge
