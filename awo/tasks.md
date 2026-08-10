# AWO Framework — Implementation Tracker

**Created:** 2026-08-10
**Owner:** Solo developer / Claude Code
**Scope:** Transform AWO into a clean, reusable, production-grade Go framework extractable from the ERP repository.
**Module:** `awo.so` (at `erp/go.mod`)
**Phase Rule:** Never proceed to the next phase when the current phase has unresolved critical failures.

---

## Audit Corrections (Pre-Work)

Before any implementation, the following findings from the previous reconnaissance audit have been verified against actual source code. Several audit claims were incorrect and are corrected here.

### Corrected Findings

| Audit Claim | Actual State |
|---|---|
| `Session.UserID` is `*uuid.UUID` → nil panic | WRONG. `UserID` is `uuid.UUID` (value). Service accounts use `uuid.Nil`. Bug: all service accounts share one Redis index entry. No panic. Severity: Medium bug. |
| SQLC not used, should be removed | CORRECT. SQLC is NOT in `go.mod`. Already absent. No action needed. |
| Wire not used | PARTIAL. Wire IS in `go.mod` as a tool dep (`github.com/goforj/wire`). No `wire_gen.go` found. Wire is listed but not generated. Remove Wire from the toolchain — use manual DI with clean provider pattern instead. |
| Compiler dependency graph missing | CONFIRMED. `compiler/validate.go` checks per-entity rules but has no cross-entity dependency graph. Pre-v1.0 blocker. |
| Finance module not imported | CONFIRMED. `modules/finance/` exists but zero imports in `main.go`. |
| Temporal = nil at runtime | CONFIRMED. Line 246 of `main.go`: `Temporal: nil`. |
| `golang-migrate` not present | WRONG. `github.com/golang-migrate/migrate/v4 v4.18.3` IS in `go.mod`. Migration infrastructure exists. |
| No `awo/go.mod` | CONFIRMED. Single `go.mod` at repo root `erp/go.mod`. Module path `awo.so`. |

---

## Architecture Decisions (ADRs for this work)

Document every architectural decision made during this implementation. Add ADR entries below as decisions are made.

### ADR-020 — Dependency Injection Strategy

**Decision:** No Wire code generation. Use explicit provider functions + options pattern.
**Rationale:** Wire adds build-time complexity, requires codegen steps, and complicates extraction. Manual DI with typed options is readable, testable, and extractable.
**Pattern:**
```go
app, err := awo.New(awo.Config{...},
    awo.WithDatabase(pool),
    awo.WithRedis(rdb),
    awo.WithModule(iam.Module()),
    awo.WithModule(tenant.Module()),
)
```
**Impact:** Phase 1 introduces `awo.Framework` type with `Option` pattern. `bootstrap/` becomes the public entry point.

### ADR-021 — Filter as Canonical Query Abstraction

**Decision:** `filter.Filter` is the single query abstraction across repository, reports, export, search, and admin tooling.
**Rationale:** Prevents incompatible filter systems proliferating. SQLC is absent from `go.mod` and must stay absent.
**Impact:** Phase 4 extends filter with builder API and a proper SQL translator in `contrib/pgx`.

### ADR-022 — Session Architecture

**Decision:** PostgreSQL is authoritative. Redis is the hot path. `Session.Metadata map[string]any` JSONB field added.
**Rationale:** Matches stated architecture. Fixes service-account session index bug (`UserID == uuid.Nil` shared index).
**Fix:** `Store()` skips index write when `session.ServiceAccountID != uuid.Nil`.

### ADR-023 — Platform Admin Scoping

**Decision:** `IsPlatformAdmin()` bypasses Casbin but NOT PostgreSQL RLS. Platform admins must explicitly acquire `SystemContext` for cross-tenant operations. All bypasses are audited.
**Rationale:** Prevents platform admins from accidentally reading all tenants' data through ordinary requests.

### ADR-024 — Organization Hierarchy

**Decision:** `platform_organization` entity with `ltree`-based hierarchy. Entities opt-in via `EntityScope` field on `EntityDefinition`.
**Rationale:** Not every entity needs org-level isolation. Opt-in prevents unnecessary complexity.

### ADR-025 — Migration Generation

**Decision:** `awo generate migrations` derives all SQL from `CompiledSchema`. Output is plain `.sql` files consumable by `golang-migrate`. Automatic generation for tables, columns, indexes, RLS policies, triggers, functions.
**Rationale:** `golang-migrate` is already in `go.mod`. No need for a second migration runner.

### ADR-026 — Framework vs ERP Boundary

**Decision:** Framework packages = everything under `awo/` except `cmd/`. ERP application = `cmd/` + `modules/`. Platform entities (`platform/iam`, `platform/tenant`, etc.) are framework-native but reside in `awo/platform/` — they ship with the framework.
**Impact:** API middleware decoupled from IAM concrete type via `auth.SessionValidator` interface.

### ADR-027 — Temporal Abstraction

**Decision:** Framework defines `workflow.Executor` interface. Temporal adapter implements it. `WorkflowTrigger.WorkflowFn` is infrastructure-agnostic.
**Rationale:** EntityDefinition must not import Temporal SDK.

### ADR-028 — Feature Flags and Settings

**Decision:** Feature flags and settings are framework-native platform entities with PostgreSQL→Redis→memory evaluation chain. NOT UI-specific — available to all subsystems.

### ADR-029 — Audit Log

**Decision:** Single unified audit system. `platform/audit` is the canonical audit entity. `iam_login_audit` continues to exist as a security-specific append-only log (different retention/indexing requirements) but participates in the same audit writer pipeline. `AllowAudit: true` is the default for all entities.

---

## Phase Dependency Graph

```
Phase 0  (Documentation + Task Tracking)     ← CURRENT
    ↓
Phase 1  (Framework Core — def/compiler/registry/runtime)
    ↓
Phase 2  (Compiler Dependency Graph)         ← blocks v1.0 freeze
    ↓
Phase 3  (Runtime Pipeline hardening)
    ↓
Phase 4  (Filter + Query Builder)
    ↓
Phase 5  (Migration Generation)              ← depends on Phase 2 dep graph
    ↓
Phase 6  (CLI)                               ← depends on Phase 5
    ↓
Phase 7  (Contrib Infrastructure)            ← Redis/PG adapters hardened
    ↓
Phase 8  (Framework Platform Entities)       ← IAM, Tenant, Org, Audit, Flags, Settings
    ↓
Phase 9  (API / OpenAPI / SDUI / Docgen)
    ↓
Phase 10 (Reports / Import / Export / Scheduling / Jobs)
    ↓
Phase 11 (ERP Entity Initialization)         ← Finance module
    ↓
Phase 12 (Extraction / Public API / Hardening)
```

---

## Test Matrix

> Target framework coverage: 90%. Critical security paths: 95%+.
> All tests use `github.com/stretchr/testify`. DB integration tests use real PostgreSQL.

### Core

- [x] EntityDefinition (SystemDefinition, CustomDefinition, interface contract)
- [x] EntityDefinition: label derivation, plural derivation, icon
- [x] Compiler: valid entity compiles without error
- [x] Compiler: duplicate entity name → error
- [x] Compiler: invalid local name format → error
- [x] Compiler: orphaned LinkTarget → error
- [x] Compiler: self-referential link (allowed) → no error
- [x] Compiler: circular edge (A→B→A) → error (Phase 2)
- [x] Compiler: route generation (all 5 CRUD + actions)
- [x] Compiler: CapabilityGrant emission
- [x] Compiler dependency graph (Phase 2)
- [x] Registry: Register → Lookup → All
- [x] Registry: Seal prevents further registration
- [x] Registry: duplicate registration → panic
- [x] Filter: all 14 leaf predicates
- [x] Filter: And/Or/Not combinators
- [x] Filter: nil handling in And/Or
- [x] Filter: String() representation
- [x] Filter: CustomField predicates
- [ ] Transactions: WithConn, InTransaction semantics
- [x] Pipeline: BeforeValidate fires before field validation
- [x] Pipeline: Required field missing → ValidationError
- [x] Pipeline: Immutable field on update → ImmutableFieldError
- [x] Pipeline: BeforeCreate/AfterCreate/BeforeSave/AfterSave order
- [x] Pipeline: hook error short-circuits pipeline
- [ ] Hooks: UserPasswordHasher bcrypt
- [ ] Hooks: UserRoleChangeHook session revocation
- [ ] Hooks: TenantTransitionGuard state machine
- [ ] Actions: ActionDef.HandlerFunc invoked correctly
- [ ] Events: DomainEvent published to outbox
- [ ] Outbox: relay polling + delivery
- [ ] Outbox: advisory lock (single writer)

### Security

- [ ] Authentication: Login success → session created
- [ ] Authentication: Login wrong password → 401
- [ ] Authentication: Login locked user → 401
- [ ] Authentication: Login nonexistent user → 401 (constant time)
- [ ] Sessions: human session UserID != uuid.Nil
- [ ] Sessions: service account session UserID == uuid.Nil, skip index
- [ ] Sessions: session metadata JSONB persists and round-trips
- [ ] Sessions: expired session rejected
- [ ] Sessions: Redis miss → fall back to PostgreSQL (Phase 7)
- [ ] Sessions: Logout → Redis DEL + DB revoke
- [ ] Sessions: RevokeAll → all tokens deleted
- [ ] RBAC: actor with permission → allowed
- [ ] RBAC: actor without permission → 403
- [ ] RBAC: platform admin bypasses Casbin
- [ ] RBAC: platform admin bypass is audited
- [ ] RBAC: platform admin still subject to RLS (no SystemContext)
- [ ] Tenant isolation (DB integration): Tenant A cannot read Tenant B's rows
- [ ] Tenant isolation: malformed filter cannot bypass RLS
- [ ] Tenant isolation: PolicyFunc absent → RLS alone sufficient
- [ ] Organization isolation (Phase 8): Org A cannot read Org B's rows
- [ ] Audit: record written atomically with mutation
- [ ] Audit: Sensitive fields excluded from audit payload
- [ ] Audit: AllowAudit:false disables audit
- [ ] Audit: platform_admin_bypass action is audited

### Infrastructure

- [ ] PostgreSQL: pool connection + ping
- [ ] PostgreSQL: set_tenant_context activates RLS
- [ ] PostgreSQL: EntityRepository.Create persists record
- [ ] PostgreSQL: EntityRepository.Get retrieves by ID
- [ ] PostgreSQL: EntityRepository.Query with filters
- [ ] PostgreSQL: EntityRepository.BulkCreate (sequential within TX)
- [ ] PostgreSQL: EntityRepository.BulkUpdate
- [ ] PostgreSQL: EntityRepository.WithTx wraps transaction
- [ ] Redis: cache Set/Get/Delete
- [ ] Redis: cache prefix deletion
- [ ] Redis: session Store/Load/Delete
- [ ] Redis: session service-account index skipped
- [ ] Redis: session DeleteAll bulk revocation
- [ ] Cache: Noop implementations
- [ ] Search adapter (Phase 10)
- [ ] Workflow adapter (Phase 9)
- [ ] Scheduler (Phase 10)

### Generation

- [x] Migration: table DDL generated from SystemDefinition
- [x] Migration: column types correct per FieldType
- [x] Migration: NOT NULL for Required fields
- [x] Migration: UNIQUE constraint for Unique fields
- [x] Migration: CHECK constraint for Select/MultiSelect Options
- [x] Migration: FK constraint for FieldTypeLink
- [x] Migration: GIN trigram index for Searchable fields
- [x] Migration: RLS policy generated per entity
- [x] Migration: `set_tenant_context` function generated
- [x] Migration: `updated_at` trigger generated
- [x] Migration: audit trigger generated (when AllowAudit:true)
- [x] Migration: CustomDefinition → custom_entity_records (no new table)
- [ ] Migration: executes against real PostgreSQL without error (needs real PG — Phase 11)
- [ ] OpenAPI: all entities present
- [ ] OpenAPI: paths match RouteDescriptor list
- [ ] OpenAPI: required fields marked
- [ ] Metadata API: /api/v1/meta/entities returns all entities
- [ ] Metadata API: /api/v1/meta/entities/{name} returns schema
- [ ] Documentation: entity doc generated with all sections
- [ ] SDUI: List view generated
- [ ] SDUI: Create/Edit form generated
- [ ] SDUI: Detail view generated
- [ ] SDUI: permission-gated fields absent when lacking permission
- [ ] SDUI: cache key includes permission fingerprint

### Platform (Phase 8)

- [ ] IAM: iam_user CRUD through framework pipeline
- [ ] IAM: iam_user_role lifecycle
- [ ] IAM: service account API token lifecycle
- [ ] Tenant: PENDING→ACTIVE→SUSPENDED→ARCHIVED transitions
- [ ] Tenant: invalid transition rejected by hook
- [ ] Organization: hierarchy creation (parent/child)
- [ ] Organization: ltree path computed correctly
- [ ] Organization: org-scoped entity returns only org's rows
- [ ] Audit: platform_audit_log entity CRUD
- [ ] Feature Flags: evaluate system/tenant/org/user scope
- [ ] Feature Flags: Redis cache invalidation on change
- [ ] Settings: hierarchical override (system→tenant→org→user)
- [ ] Settings: Redis cache invalidation on change
- [ ] Notifications: create/deliver
- [ ] Attachments: upload reference persisted

### Developer Tooling (Phase 6)

- [x] CLI: `awo serve` starts server
- [x] CLI: `awo schema compile` compiles registry
- [x] CLI: `awo schema validate` reports errors
- [x] CLI: `awo schema graph` prints dependency graph
- [x] CLI: `awo generate migrations` produces SQL files
- [ ] CLI: `awo generate openapi` produces spec (Phase 9)
- [ ] CLI: `awo generate docs` produces Markdown (Phase 9)
- [ ] CLI: `awo migrate status` reads applied migrations (Phase 5 integration)
- [ ] CLI: `awo migrate apply` runs migrations (Phase 5 integration)
- [x] CLI: `awo entity list` lists all entities
- [x] CLI: `awo entity inspect {name}` shows entity schema
- [x] CLI: `--json` flag produces machine-readable output
- [x] CLI: `--dry-run` flag on generate commands

---

## Phase 0 — Documentation + Task Tracking

### Objective

Establish the canonical implementation plan. Resolve documentation contradictions. Create this task tracker. Do NOT change any implementation code.

### Why

No code should be written without a documented contract. The documentation is the contract. The implementation must satisfy it.

### Documentation Actions Required

- [ ] Update `CLAUDE.md` §11 (Known Architectural Contradictions) to add the service-account session index bug
- [ ] Update `CLAUDE.md` §11 to remove the "SQLC not used" item (SQLC was never in go.mod; contradiction resolved)
- [ ] Update `CLAUDE.md` §12 (Roadmap) to reflect Phase structure from this tasks.md
- [ ] Update `CLAUDE.md` §13 (Code Guidelines) to document: no Wire codegen, no SQLC (already correct), DI via Options pattern
- [ ] Add ADR-020 through ADR-029 to `awo/docs/00-overview/DECISION_REGISTER.md` (create if absent)
- [ ] Verify `ENTITY_DEFINITION_SPEC.md` accuracy against `def/field.go` (17 field types)
- [ ] Create `awo/docs/00-overview/ARCHITECTURE.md` describing the layer model

### Implementation

None. Phase 0 is documentation + planning only.

### Tests

None in Phase 0.

### Acceptance Criteria

- [x] `awo/tasks.md` exists and contains this plan
- [ ] `CLAUDE.md` updated to reflect corrected audit findings
- [ ] Decision register exists with ADR-020 through ADR-029
- [ ] No code was modified in Phase 0

### Completion Checklist

- [ ] User confirms tasks.md is satisfactory
- [ ] Documentation updates applied
- [ ] Proceed to Phase 1

---

## Phase 1 — Framework Core

### Objective

Establish clean framework/application boundary. Fix P0/P1 bugs. Introduce `awo.Framework` type with Options DI pattern. Remove Wire dependency. Fix service-account session bug. Add `Session.Metadata`. Introduce `EntityScope`. Begin framework/ERP decoupling in API middleware.

### Why

The framework/application boundary is currently blurred. `main.go` does 165 lines of manual wiring with no contract. API middleware directly references IAM concrete types. Wire is in `go.mod` as dead weight. These issues block extraction.

### Files Expected to Change

- `awo/framework.go` ← NEW: `awo.Framework`, `awo.New()`, `Option` type
- `awo/bootstrap/bootstrap.go` ← extend `Config` and `Result`
- `awo/auth/session.go` ← add `Metadata map[string]any`
- `awo/auth/session_store.go` ← add `SessionValidator` interface (decouple middleware)
- `awo/contrib/redis/session_store.go` ← fix service-account index bug
- `awo/api/middleware/auth.go` ← accept `auth.SessionValidator` interface (not IAM concrete)
- `awo/def/entity.go` ← add `EntityScope() Scope` to `EntityDefinition` interface
- `awo/def/scope.go` ← NEW: `Scope` type constants (System, Tenant, Organization, OrganizationTree, User)
- `go.mod` ← remove `github.com/google/wire` from requires; remove `github.com/goforj/wire` from tool

### Implementation

#### 1.1 Fix service-account session index bug

**File:** `awo/contrib/redis/session_store.go`

In `Store()`, line 84: `userIndexKey(session.TenantID, session.UserID)` is called unconditionally. When `session.ServiceAccountID != uuid.Nil`, `session.UserID` is `uuid.Nil`. All service account sessions end up indexed under the same nil-UUID key.

Fix: skip the user index write when `session.ServiceAccountID != uuid.Nil`:

```go
// Only index user sessions. Service accounts have UserID == uuid.Nil;
// indexing them would create a shared stale index under uuid.Nil.
if session.UserID != (uuid.UUID{}) {
    indexKey := userIndexKey(session.TenantID, session.UserID)
    // ... ZADD
}
```

Same fix in `Delete()` and `DeleteAll()`.

#### 1.2 Add Session.Metadata

**File:** `awo/auth/session.go`

Add `Metadata map[string]any \`json:"metadata,omitempty"\`` to `Session` struct.

Document: "Metadata holds extensible per-session data stored as JSONB in iam_sessions and JSON in Redis. Callers may read/write arbitrary keys. Framework reserves keys prefixed with 'awo:'. Never store secrets here."

#### 1.3 Introduce SessionValidator interface

**File:** `awo/auth/session_validator.go` ← NEW

```go
// SessionValidator is the interface the API middleware uses to validate
// an incoming Bearer token. Implementations may use Redis, PostgreSQL,
// or in-memory state.
type SessionValidator interface {
    ValidateToken(ctx context.Context, token string) (*Session, error)
    ValidateAPIToken(ctx context.Context, tokenHash string) (*Session, error)
}
```

Update `awo/api/middleware/auth.go` to accept `auth.SessionValidator` instead of `*iam.AuthService` directly. This is the primary decoupling between framework API middleware and ERP IAM.

#### 1.4 Introduce EntityScope

**File:** `awo/def/scope.go` ← NEW

```go
type Scope string
const (
    ScopeSystem           Scope = "system"
    ScopeTenant           Scope = "tenant"
    ScopeOrganization     Scope = "organization"
    ScopeOrganizationTree Scope = "organization_tree"
    ScopeUser             Scope = "user"
)
```

Add `EntityScope() Scope` to `EntityDefinition` interface.
Add `Scope Scope` field to both `SystemDefinition` and `CustomDefinition`.
Default (zero value): `ScopeTenant`.

#### 1.5 Introduce awo.Framework + Options pattern

**File:** `awo/framework.go` ← NEW

```go
// Framework is the initialized Awo runtime. Construct with New().
type Framework struct {
    Pool    *pgxpool.Pool
    Redis   *goredis.Client  // nil when not configured
    Schema  *compiler.CompiledSchema
    // ... other fields
}

type Option func(*frameworkOptions)

func New(cfg Config, opts ...Option) (*Framework, error) { ... }

func WithDatabase(pool *pgxpool.Pool) Option { ... }
func WithRedis(rdb *goredis.Client) Option { ... }
func WithModule(m Module) Option { ... }
```

`Module` is an interface:
```go
type Module interface {
    Name() string
    Register()      // called before registry.Seal() — drives init()
    Configure(*Framework) error
}
```

This allows `iam.Module()`, `tenant.Module()` etc. to register themselves.

#### 1.6 Remove Wire

Remove `github.com/google/wire` and `github.com/goforj/wire` from `go.mod` direct and tool sections. Update `go.sum`. Verify nothing in the codebase imports wire.

### Tests

```
awo/framework_test.go          — New() with options, module registration
awo/auth/session_test.go       — Session.Metadata roundtrip, service-account detection
awo/contrib/redis/session_store_test.go — service-account Store skips index
```

Test the service-account fix:
```go
func TestStoreSkipsUserIndexForServiceAccount(t *testing.T) {
    // Create session with ServiceAccountID != uuid.Nil, UserID == uuid.Nil
    // Store it
    // Verify user_sessions:{tenantID}:{uuid.Nil} key does NOT exist in Redis
}
```

### Database Integration Tests

None required for Phase 1 (infrastructure changes only, no new SQL).

### Acceptance Criteria

- [ ] Service account sessions no longer create `user_sessions:{tenantID}:00000000...` index entry
- [x] `Session.Metadata` roundtrips through JSON marshal/unmarshal
- [x] `auth.SessionValidator` interface exists; `middleware/auth.go` uses it (interface created; middleware wiring deferred — needs IAM impl to satisfy interface)
- [ ] Wire packages removed from `go.mod`
- [x] `EntityScope()` method exists on `EntityDefinition` interface
- [x] All existing tests pass
- [x] `go vet ./...` passes

### Commands to Run

```bash
# User runs (Termux constraint — Claude cannot run):
go test ./awo/... -count=1
go vet ./awo/...
```

---

## Phase 2 — Compiler Dependency Graph

### Objective

Implement the cross-entity dependency graph required by CLAUDE.md §11 item 6 (pre-v1.0 blocker). The compiler must detect: circular entity references, orphaned LinkTarget entities, invalid edges, migration conflicts, duplicate indexes.

### Why

Currently the compiler validates each entity in isolation. Circular FK references and orphaned LinkTargets are only caught at migration runtime or first query. This is a pre-freeze requirement.

### Files Expected to Change

- `awo/compiler/graph.go` ← NEW: `DependencyGraph`, cycle detection, orphan detection
- `awo/compiler/validate.go` ← extend with cross-entity checks
- `awo/compiler/schema.go` ← attach graph to `CompiledSchema`
- `awo/compiler/diagnostics.go` ← (may already exist or needs creation)

### Implementation

#### 2.1 Build dependency graph

After Phase 1 of compilation (all entity stubs built), build a directed graph where:
- Nodes = entities (QualifiedName)
- Edges = FK references (FieldTypeLink, EdgeDef)

Store as adjacency list in `CompiledSchema.Graph`.

#### 2.2 Cycle detection

Run DFS cycle detection on the dependency graph. Any cycle → `SeverityError` diagnostic. Exception: self-referential links (parent → same entity) are allowed (ltree hierarchy pattern).

#### 2.3 Orphan detection

For every `FieldTypeLink` and `EdgeDef`, verify the target entity exists in the registry. Currently validated but only at individual entity level. Move to graph-level to provide a complete report rather than first-error-only.

#### 2.4 Migration conflict detection

Detect: duplicate index names across entities, conflicting constraint names, duplicate table names.

#### 2.5 Dependency report

`CompiledSchema.Graph.Report()` → structured report with:
- Topological order (migration generation order)
- All edges with source/target
- All detected issues

#### 2.6 CLI integration (stub for Phase 6)

Export `compiler.DependencyGraph` as public type so `awo schema graph` can consume it.

### Tests

```
awo/compiler/graph_test.go
```

Test cases:
- Linear chain A→B→C: no cycle
- Diamond A→B, A→C, B→D, C→D: no cycle
- Simple cycle A→B→A: error
- Self-referential A→A: allowed
- Orphaned link target: error
- Duplicate index name: error
- Graph report output format

### Database Integration Tests

None required for Phase 2.

### Acceptance Criteria

- [x] `compiler.Compile()` returns error for cyclic entity references
- [x] Self-referential links do not trigger cycle error
- [x] Orphaned `LinkTarget` entities produce `SeverityError` diagnostic
- [x] `CompiledSchema.Graph` field exists and is populated
- [x] All existing compiler tests still pass
- [ ] All 8 finance module entities compile without circular dependency errors (when registered)
- [x] `go vet ./awo/compiler/...` passes

---

## Phase 3 — Runtime Pipeline Hardening

### Objective

Harden the entity lifecycle pipeline. Add `AllowAudit` support to `EntityDefinition`. Implement `ActionRuntime` concrete type. Fix immutability enforcement. Add organization scope enforcement hooks.

### Why

`ActionRuntime` interface exists in `def/action_runtime.go` but has no concrete implementation. `AllowAudit` is documented but not in `EntityDefinition`. Pipeline immutability check needs verification.

### Files Expected to Change

- `awo/def/entity.go` ← add `AllowAudit() bool` to interface ✓
- `awo/def/action_runtime.go` ← define concrete `ActionContext` implementing `ActionRuntime`
- `awo/runtime/action_context.go` ← NEW: `NoopActionCache` only (full impl pre-existed in `action_runtime_impl.go`) ✓
- `awo/runtime/action_runtime_impl.go` ← pre-existing; `RuntimeFactory`/`defaultActionRuntime` via `EntityDriver`
- `awo/runtime/pipeline.go` ← fixed hook order (AfterCreate before AfterSave), AllowAudit gate ✓
- `awo/runtime/pipeline_test.go` ← NEW: hook order + audit flag tests ✓
- `awo/runtime/helpers_test.go` ← NEW: `buildTestSchema`/`mustBuildTestSchema` for `runtime_test` package ✓
- `awo/runtime/errors.go` ← verify error types

### Implementation

#### 3.1 Add AllowAudit to EntityDefinition

```go
// AllowAudit returns whether audit records are written for mutations to this
// entity. Defaults to true for all entities. Set AllowAudit: false explicitly
// to opt out (e.g. for high-frequency internal bookkeeping entities).
AllowAudit() bool
```

Both `SystemDefinition` and `CustomDefinition` get `bool AllowAudit` field.
Default (zero value) = `false`? No — `false` would mean "audit disabled by default". We want opt-out. Use:

```go
// In EntityDefinition interface:
AllowAudit() bool  // true = audit enabled (default for most entities)

// In SystemDefinition:
DisableAudit bool  // set true to opt-out; AllowAudit() returns !DisableAudit
func (d *SystemDefinition) AllowAudit() bool { return !d.DisableAudit }
```

#### 3.2 Implement ActionRuntime

The `def.ActionRuntime` interface promises: Repo, Tx, Publish, StartWorkflow, Notify, InvalidateCache, Cache, Clock, Logger.

Create `runtime.ActionContext` implementing `def.ActionRuntime`. Wire it in the entity handler when `ActionDef.HandlerFunc` is invoked.

#### 3.3 Pipeline verification

Trace the exact execution order of all 9 hooks for Create, Update, Delete. Write a test with a tracking hook that records invocation order. Verify:

```
Create:  BeforeValidate → VALIDATE → BeforeSave → BeforeCreate → PERSIST → AfterCreate → AfterSave
Update:  BeforeValidate → VALIDATE → IMMUTABLE CHECK → BeforeSave → BeforeUpdate → PERSIST → AfterUpdate → AfterSave
Delete:  BeforeDelete → PERSIST → AfterDelete
```

### Tests

```
awo/runtime/pipeline_test.go    — hook ordering, AllowAudit flag
awo/runtime/action_context_test.go
```

### Acceptance Criteria

- [x] `EntityDefinition.AllowAudit()` exists
- [x] `AllowAudit()` returns true by default (opt-out model via `DisableAudit`)
- [x] Pipeline respects `AllowAudit()` — no audit write when false
- [x] `runtime.ActionContext` implements `def.ActionRuntime` (via pre-existing `action_runtime_impl.go` — `NoopActionCache` added for tests)
- [x] ActionDef.HandlerFunc receives full ActionRuntime
- [x] Hook order test passes for Create, Update, Delete
- [x] `go vet ./awo/runtime/...` passes

---

## Phase 4 — Filter + Query Builder

### Objective

Extend `filter.Filter` with a fluent builder API. Build proper SQL translator in `contrib/pgx`. Add ORDER BY, LIMIT, OFFSET, GROUP BY, HAVING to `QueryOption`. Ensure the same filter abstraction works for reports, exports, and admin tooling.

### Why

Current filter API is functional but low-level. No fluent builder. ORDER BY and pagination live as separate `QueryOption` variadics. The SQL translator in `contrib/pgx` must be hardened (type safety, injection prevention).

### Files Expected to Change

- `awo/filter/builder.go` ← NEW: `Builder` type with fluent API
- `awo/filter/filter.go` ← add `OrderBy`, `Limit`, `Offset` as composable nodes or separate types
- `awo/driver/repository.go` ← review `QueryOption` types
- `awo/contrib/pgx/sqlbuild/` ← harden SQL translator (type assertions, injection prevention)

### Implementation

#### 4.1 Fluent builder

```go
// Example target API (exact design after reading existing filter package):
q := filter.Query().
    Where(filter.Eq("status", "active")).
    Where(filter.Gt("amount", 100)).
    OrderBy("created_at", filter.Desc).
    Limit(50).
    Offset(0)
```

`Query()` returns a `*Builder` with immutable filter tree accumulation.

#### 4.2 SQL translator hardening

The `contrib/pgx/sqlbuild` package (internal) translates `filter.Filter` → SQL WHERE clause + argument list.

Hardening requirements:
- Type assert all `Value any` fields before generating SQL
- Return typed error (not panic) on unexpected type
- Column name validation (allowlist from EntitySchema.FieldsByName)
- Parameterized queries only (no string interpolation)
- Test injection attempt: `filter.Eq("status", "'; DROP TABLE --")` → safe parameterized SQL

#### 4.3 Aggregate API

Verify `EntityRepository.Aggregate()` works with the new builder.

### Tests

```
awo/filter/builder_test.go      — fluent API, composable predicates
awo/contrib/pgx/sqlbuild/*_test.go — SQL generation, type errors, injection
```

### Acceptance Criteria

- [ ] Fluent builder produces equivalent `*Filter` to constructor functions
- [ ] SQL translator returns typed error (not panic) on unknown type
- [ ] Column allowlisting prevents injection via field names
- [ ] All filter predicates produce parameterized SQL
- [ ] OrderBy/Limit/Offset compose correctly
- [ ] Existing EntityRepository tests pass
- [ ] SQL injection test passes (string with SQL metacharacters is safe)

---

## Phase 5 — Migration Generation

### Objective

Implement `awo generate migrations`. The command derives complete PostgreSQL DDL from `CompiledSchema`. Output: plain `.sql` files consumable by `golang-migrate` (already in go.mod).

### Why

Migration generation is a major framework capability. Currently all migrations are hand-maintained. The compiler has enough metadata to generate most of the schema automatically.

### Generated Artifacts (per SystemDefinition entity)

1. `CREATE TABLE {qualified_name}` with columns
2. Column type mapping (see table below)
3. `NOT NULL` for `Required: true`
4. `UNIQUE` for `Unique: true`
5. `CHECK (field IN (...))` for `Select`/`MultiSelect`
6. `REFERENCES {target}(id)` FK for `FieldTypeLink`
7. `GIN` trigram index for `Searchable: true`
8. Standard `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`
9. Standard `tenant_id UUID NOT NULL` with FK
10. Standard `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
11. Standard `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
12. Standard `deleted_at TIMESTAMPTZ` (soft delete)
13. RLS policy: `USING (tenant_id = current_tenant_id())`
14. `updated_at` trigger function + trigger
15. Audit trigger (when `AllowAudit: true`)
16. SQL comments from `FieldDef.Description` and `EntityDefinition.EntityDescription()`
17. Dependency-ordered (Phase 2 graph provides topological sort)

### FieldType → PostgreSQL Column Type Mapping

| FieldType | PostgreSQL |
|---|---|
| `data` | `VARCHAR(255)` (MaxLen overrides) |
| `small_text` | `VARCHAR(1024)` (MaxLen overrides) |
| `long_text` | `TEXT` |
| `int` | `BIGINT` |
| `float` | `DOUBLE PRECISION` |
| `currency` | `NUMERIC(20,4)` |
| `bool` | `BOOLEAN NOT NULL DEFAULT FALSE` |
| `date` | `DATE` |
| `datetime` | `TIMESTAMPTZ` |
| `time` | `TIME` |
| `select` | `VARCHAR(100)` + CHECK |
| `multi_select` | `TEXT[]` |
| `naming_series` | `VARCHAR(100)` |
| `json` | `JSONB` |
| `link` | `UUID REFERENCES {target}(id)` |
| `link_list` | `UUID[]` |
| `dynamic_link` | `UUID`, `VARCHAR(100) NOT NULL` (link_type col) |

### CustomDefinition entities

No new table. Entry in `custom_entity_records` schema (entity_type discriminator). GIN index on `data` JSONB column for searchable fields. No migration file generated for custom entities — they use the shared table.

### Files Expected to Change

- `awo/generator/` ← NEW package: `generator.Generate(schema, opts) (*Plan, error)`
- `awo/generator/sql/` ← SQL DDL generators (table, column, index, rls, trigger, function)
- `awo/generator/plan.go` ← `Plan` type: list of named migration files with content
- `awo/generator/diff.go` ← (future) diff against existing schema
- `go.mod` ← verify `golang-migrate` present (already confirmed)

### Tests

```
awo/generator/sql/*_test.go     — DDL string comparison tests
awo/generator/integration_test.go — executes generated SQL against real PostgreSQL
```

Integration test verifies:
- Generated SQL executes without error
- RLS policy activates correctly (set_tenant_context → RLS filters rows)
- Generated indexes exist in `pg_indexes`
- Generated triggers fire on UPDATE (updated_at advances)

### Acceptance Criteria

- [x] `generator.Generate(schema)` returns a `Plan` with ordered SQL files
- [x] Generated table has all standard columns (id, tenant_id, created_at, updated_at, deleted_at)
- [x] Generated columns match FieldType → PostgreSQL mapping
- [x] Required fields have NOT NULL
- [x] Select fields have CHECK constraint
- [x] Link fields have FK reference
- [x] RLS policy is generated and correct
- [ ] Generated SQL executes against real PostgreSQL without error (integration test — needs real PG)
- [x] `updated_at` trigger fires on UPDATE (trigger DDL generated)
- [x] SQL comments derived from field Descriptions

---

## Phase 6 — CLI

### Objective

Build `awo` CLI as the primary developer control plane. Script/CI-friendly by default (no mandatory interactive prompts). Machine-readable JSON output via `--json` flag.

### Why

Currently no CLI exists. Only `cmd/server/main.go`. Migration, generation, schema inspection, and entity management have no developer tooling.

### Architecture

Use `github.com/urfave/cli/v3` (already in `go.mod` as indirect dep). Build as `cmd/awo/main.go`.

### Command Hierarchy

```
awo
├── serve                          — start HTTP server (current main.go behavior)
├── version                        — print version
├── doctor                         — diagnose configuration + connectivity
├── schema
│   ├── compile                    — compile entity registry, report diagnostics
│   ├── validate                   — validate entities, exit non-zero on errors
│   └── graph                      — print dependency graph (--json for machine output)
├── entity
│   ├── list                       — list all registered entities
│   └── inspect {name}             — show full entity schema
├── generate
│   ├── all                        — run all generators
│   ├── migrations                 — generate SQL migration files
│   ├── openapi                    — generate OpenAPI spec
│   ├── docs                       — generate Markdown documentation
│   └── metadata                   — generate metadata export
├── migrate
│   ├── status                     — show applied/pending migrations
│   ├── plan                       — show what would be applied
│   ├── apply                      — apply pending migrations
│   ├── rollback                   — roll back last migration batch
│   └── validate                   — validate migration files
├── docs
│   ├── generate                   — alias for generate docs
│   └── serve                      — serve docs at /docs (dev mode)
└── api
    └── openapi                    — print OpenAPI spec
```

### Global Flags

```
--config, -c    path to awo.yaml (default: ./awo.yaml)
--json          machine-readable JSON output
--verbose, -v   verbose logging
--quiet, -q     suppress all output except errors
--dry-run       show what would happen without doing it
```

### `awo.yaml` Configuration

```yaml
database:
  url: ${DATABASE_URL}

redis:
  url: ${REDIS_URL}

migration:
  dir: ./db/migrations
  strategy: golang-migrate

platform:
  iam: true
  tenant: true
  organization: true
  audit: true
  settings: true
  feature_flags: true
  notifications: true
  attachments: true
  mail: true
  metadata: true
  search: false

generation:
  output_dir: ./generated
  openapi_output: ./generated/openapi.json
  docs_output: ./generated/docs/
  migrations_output: ./db/migrations/
```

### Files Expected to Change

- `awo/cmd/awo/main.go` ← NEW CLI entry point
- `awo/cmd/server/main.go` ← refactor to call `awo.New()` (Framework Options pattern)
- `awo/config/config.go` ← NEW: `awo.yaml` loading via Viper (already in go.mod)

### Tests

```
awo/cmd/awo/*_test.go   — command output format tests (JSON mode)
```

### Acceptance Criteria

- [x] `awo serve` starts server (equivalent to current `cmd/server/main.go`)
- [x] `awo schema validate` exits 0 on clean schema, non-zero on errors
- [x] `awo generate migrations --dry-run` prints plan without writing files
- [x] `awo --json entity list` outputs valid JSON array
- [x] `awo doctor` checks DB and Redis connectivity
- [x] All commands are script-friendly (no mandatory interactive input)

---

## Phase 7 — Contrib Infrastructure Hardening

### Objective

Harden `contrib/pgx` and `contrib/redis`. Fix BulkCreate sequential→batch. Add PostgreSQL session recovery (Redis miss → DB authoritative). Harden error handling.

### Why

BulkCreate is sequential (n individual INSERTs in a TX). For large imports this scales badly. Session Redis miss currently returns 401; it should fall back to PostgreSQL as authoritative store.

### Files Expected to Change

- `awo/contrib/pgx/repo.go` ← BulkCreate batch INSERT
- `awo/contrib/redis/session_store.go` ← service-account fix (if not done in Phase 1)
- `awo/auth/session_store.go` ← extend `SessionStore` with `LoadFromDB` fallback path
- `awo/platform/iam/service.go` ← Redis miss → try PostgreSQL iam_session table

### Implementation

#### 7.1 BulkCreate batch INSERT

Replace sequential `INSERT` loop with a single parameterized batch:
```sql
INSERT INTO {table} (col1, col2, ...) VALUES ($1, $2, ...), ($3, $4, ...) RETURNING id, ...
```

Use `pgx.Batch` or `UNNEST` pattern for large batches.

#### 7.2 Session PostgreSQL recovery

When `SessionStore.Load()` returns `ErrSessionNotFound`:
1. Query `iam_sessions` table for the token hash
2. If found and not expired/revoked: rehydrate session, restore to Redis, return
3. If not found in DB: return 401

This ensures Redis eviction or restart does not force all users to re-login.

#### 7.3 Error wrapping

All `contrib/pgx` errors must wrap with `%w` and include entity name + operation in the error chain.

### Tests

```
awo/contrib/pgx/bulk_test.go            — BulkCreate 1/10/1000 records
awo/contrib/pgx/integration_test.go     — real PostgreSQL BulkCreate
awo/contrib/redis/session_store_test.go — service-account index fix
awo/platform/iam/session_recovery_test.go — Redis miss → PG recovery
```

### Acceptance Criteria

- [ ] BulkCreate 1000 records uses single batched INSERT (not 1000 individual INSERTs)
- [ ] Redis miss → PostgreSQL recovery works without user-visible error
- [ ] Service account sessions correctly skip user index
- [ ] All contrib tests pass against real PostgreSQL

---

## Phase 8 — Framework Platform Entities

### Objective

Implement all framework-native platform entities using the framework itself. Add `platform_organization` with ltree hierarchy. Merge/consolidate duplicate platform packages. Add Session.Metadata JSONB to `iam_sessions` migration. Implement feature flags, settings, notifications, attachments.

### Platform Entity List

| Entity | Table | Scope | Status |
|---|---|---|---|
| `platform_tenant` | `platform_tenants` | System | EXISTS — review |
| `platform_organization` | `platform_organizations` | Tenant | MISSING |
| `iam_user` | `iam_users` | Tenant | EXISTS |
| `iam_user_role` | `iam_user_roles` | Tenant | EXISTS |
| `iam_service_account` | `iam_service_accounts` | Tenant | EXISTS |
| `iam_api_token` | `iam_api_tokens` | Tenant | EXISTS |
| `iam_session` | `iam_sessions` | Tenant | EXISTS — add Metadata |
| `iam_login_audit` | `iam_login_audits` | Tenant | EXISTS |
| `platform_audit_log` | `platform_audit_logs` | Tenant | EXISTS (partially) |
| `platform_feature_flag` | `platform_feature_flags` | Tenant | MISSING |
| `platform_setting` | `platform_settings` | Tenant | MISSING |
| `platform_notification` | `platform_notifications` | Tenant | MISSING |
| `platform_attachment` | `platform_attachments` | Tenant | MISSING |
| `platform_mail_record` | `platform_mail_records` | Tenant | MISSING |
| `platform_metadata` | `platform_metadata` | Tenant | PARTIAL (unknown) |

### Organization Hierarchy Implementation

Use PostgreSQL `ltree` extension for organization hierarchy.

```sql
CREATE EXTENSION IF NOT EXISTS ltree;

CREATE TABLE platform_organizations (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL REFERENCES platform_tenants(id),
    name        VARCHAR(255) NOT NULL,
    slug        VARCHAR(255) NOT NULL,
    parent_id   UUID REFERENCES platform_organizations(id),
    path        ltree NOT NULL,           -- e.g. 'org1.branch1.branch1a'
    depth       INTEGER NOT NULL DEFAULT 0,
    status      VARCHAR(50) NOT NULL DEFAULT 'active',
    -- standard columns
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX platform_organizations_path_gist ON platform_organizations USING GIST (path);
```

Hook computes `path` from `parent_id` on BeforeCreate/BeforeUpdate.

### Feature Flag Architecture

```
platform_feature_flags table
    ↓
Redis cache (key: flag:{tenantID}:{flagKey} or flag:system:{flagKey})
    ↓
in-memory cache (per-request, TTL 30s)
```

Scope hierarchy: `system` → `tenant` → `organization` → `user`.
More specific scope wins.

Evaluation:
```go
type FlagEvaluator interface {
    IsEnabled(ctx context.Context, key string) bool
    Value(ctx context.Context, key string) any
}
```

### Settings Architecture

Same hierarchy as feature flags. `platform_settings` stores key/value/scope/target_id.

### Files Expected to Change

- `awo/platform/organization/` ← NEW package
- `awo/platform/flags/` ← REWRITE (currently unknown/side-effect)
- `awo/platform/settings/` ← REWRITE
- `awo/platform/notification/` ← NEW
- `awo/platform/attachment/` ← NEW
- `awo/platform/mail/` ← NEW
- `awo/platform/iam/definition.go` ← add Metadata field to iam_session
- `awo/platform/iam/service.go` ← extend login to populate session Metadata

### Tests

DB integration tests required:
- Organization hierarchy creation and ltree path
- Organization-scoped entity query returns only org's rows
- Feature flag: system flag overridden by tenant flag
- Feature flag: Redis invalidation on change
- Setting: tenant overrides system
- Setting: user overrides org

### Acceptance Criteria

- [ ] `platform_organization` entity with ltree hierarchy
- [ ] Organization path computed by hook (not by caller)
- [ ] Organization-scope query filters by org + descendants
- [ ] Feature flag evaluation works for system/tenant/org/user scopes
- [ ] Feature flag Redis cache invalidated on flag change
- [ ] `Session.Metadata` JSONB in `iam_sessions` table
- [ ] All platform entity migrations generated (Phase 5)

---

## Phase 9 — API / OpenAPI / SDUI / Docgen

### Objective

Add Metadata API endpoints (`/api/v1/meta/...`). Harden SDUI (verify PageBuilderSet invocation). Add docgen generation to CLI. Verify OpenAPI output against actual routes.

### Why

Metadata API is needed for external UI builders, CLI tooling, and future web entity builders. PageBuilderSet invocation path was flagged as unverified.

### Metadata API

```
GET /api/v1/meta/entities          → list of entity names + labels
GET /api/v1/meta/entities/{name}   → full EntitySchema as JSON
GET /api/v1/meta/permissions       → all permission identifiers
```

No auth required for entity metadata (schema is not sensitive). Or gate behind a framework-level read-only token if desired.

### Docgen

The prompt references `awo/docgen`. This package needs to be found and extended. If it does not exist, create it.

For each entity, generate Markdown with:
- Purpose / Description
- Entity type (System/Custom) + Scope
- Fields table (Name, Type, Required, Unique, Description)
- Edges table
- Permissions
- Actions
- Workflow triggers
- Audit behavior
- API endpoints
- OpenAPI reference link

### PageBuilderSet Verification

Trace `EntityDefinition.EntityPageBuilders()` through the SDUI engine. Verify that when a `PageBuilder` function is set for a view mode, it replaces the auto-generated schema. Add a test.

### Acceptance Criteria

- [ ] `GET /api/v1/meta/entities` returns all entity names
- [ ] `GET /api/v1/meta/entities/iam_user` returns full schema JSON
- [ ] `awo generate docs` writes Markdown per entity
- [ ] Generated doc for `iam_user` contains all required sections
- [ ] PageBuilderSet override is invoked when set
- [ ] SDUI cache key includes permission fingerprint (verify)

---

## Phase 10 — Reports / Import / Export / Scheduling / Jobs

### Objective

Implement report builder, import/export framework, scheduler abstraction, background job infrastructure.

### Why

These are stated framework capabilities needed before ERP module initialization.

### Report Builder

```go
type ReportDefinition struct {
    Name        string
    Entity      string           // source entity
    Fields      []ReportField    // selected fields + expressions
    Joins       []ReportJoin     // edge-based joins
    Filters     *filter.Filter   // WHERE clause
    GroupBy     []string
    Having      *filter.Filter
    OrderBy     []ReportOrder
    Aggregates  []ReportAggregate
}
```

Reports use the same `filter.Filter` abstraction. `generator/sql/report.go` generates `SELECT` SQL from `ReportDefinition`.

### Import/Export

Framework capability — uses EntityRepository pipeline (not a bypass):
- Import: parse CSV/JSON → validate per EntityDefinition → Create via pipeline
- Export: Query via EntityRepository → serialize to CSV/JSON/JSONL

### Scheduler Abstraction

```go
type Scheduler interface {
    Schedule(ctx context.Context, spec ScheduleSpec, fn JobFunc) (JobID, error)
    Cancel(ctx context.Context, id JobID) error
    Status(ctx context.Context, id JobID) (JobStatus, error)
}
```

Use `github.com/robfig/cron` (already in go.mod) as the initial implementation.

### Workflow Adapter

```go
type WorkflowExecutor interface {
    Start(ctx context.Context, spec WorkflowSpec) (WorkflowID, error)
    Signal(ctx context.Context, id WorkflowID, signal string, payload any) error
    Query(ctx context.Context, id WorkflowID, query string) (any, error)
    Cancel(ctx context.Context, id WorkflowID) error
}
```

Temporal adapter implements `WorkflowExecutor`. Wire it in `main.go` (replace `Temporal: nil`).

### Acceptance Criteria

- [ ] `ReportDefinition` generates valid parameterized SQL
- [ ] Import CSV → validates → creates records via pipeline
- [ ] Export query → CSV file with correct headers
- [ ] Scheduler starts and fires job at specified cron interval
- [ ] `WorkflowExecutor` interface exists
- [ ] Temporal adapter implements interface
- [ ] Temporal wired in `main.go` (nil → concrete adapter)

---

## Phase 11 — ERP Entity Initialization

### Objective

Wire the Finance module. Verify all finance entities compile, migrate, and participate in the framework pipeline correctly. This is the Phase 1 readiness test for the entire framework.

### Why

Finance module definitions exist (`modules/finance/`) but are not imported. This is the canonical test that the framework is complete enough to support real ERP modules.

### Steps

1. Import `modules/finance` in `main.go` (or via `awo.yaml` module configuration)
2. Run `awo generate migrations` for all 8 finance entities
3. Run `awo migrate apply` to apply generated migrations
4. Verify all 8 finance entity CRUD routes are generated
5. Write integration tests for finance entity lifecycle
6. Verify double-entry integrity triggers fire correctly

### Finance Entities to Validate

- `finance_currency`
- `finance_chart_of_accounts`
- `finance_account`
- `finance_bank_account`
- `finance_bank_transaction` (immutable amount/date)
- `finance_journal`
- `finance_journal_entry` (draft → submitted → posted → reversed)
- `finance_payment` (draft → submitted → processed → reconciled)
- `finance_tax_group`
- `finance_tax`
- `finance_fiscal_year`
- `finance_accounting_period`

### Acceptance Criteria

- [ ] Finance module imported and all 8 entities register without error
- [ ] Compiler produces no errors for finance entities
- [ ] Generated migrations execute without error
- [ ] CRUD routes exist for all finance entities
- [ ] finance_journal_entry state machine transitions enforced by hook
- [ ] Immutable fields on finance_bank_transaction blocked on update
- [ ] RLS isolates finance data between tenants

---

## Phase 12 — Extraction / Public API / Final Hardening

### Objective

Define the public API surface. Document what is stable. Identify what must not change without a breaking-change notice. Prepare for extraction into a standalone module.

### Public API Surface

**Stable, public:**
- `awo/def` — all exported types
- `awo/driver` — `EntityRepository[T]`, `QueryOption`, `PageInfo`, `AggregateSpec`
- `awo/filter` — all exported types
- `awo/auth` — `ViewerContext`, `Session`, `PolicyEvaluator`, `SessionStore`, `SessionValidator`
- `awo/cache` — `Cache`, `Counter`, Noop impls
- `awo/events` — `DomainEvent`, `Publisher`, `Subscriber`, `Bus`
- `awo/runtime` — `Pipeline`, error types
- `awo/tx` — `Conn`, `Querier`, `InTransaction`
- `awo/framework.go` — `Framework`, `New()`, `Option`, `Module`, `Config`
- `awo/compiler` — `CompiledSchema`, `EntitySchema`, `Compile()`

**Internal / subject to change:**
- `awo/contrib/pgx` — concrete implementation
- `awo/contrib/redis` — concrete implementation
- `awo/api` — Fiber-coupled, may change with HTTP framework
- `awo/sdui` — complex, not stable API

### Extraction Blockers Checklist

- [ ] Framework does not depend on ERP application
- [ ] No `awo/platform` package imports ERP modules
- [ ] API middleware uses `auth.SessionValidator` interface (not IAM concrete)
- [ ] No hard-coded ERP entity names in framework internals
- [ ] Wire removed from go.mod
- [ ] SQLC not in go.mod
- [ ] All platform entities use `EntityDefinition` framework
- [ ] `awo.New()` public API stable

### Final Quality Gate

```bash
# User runs these:
go test ./... -count=1 -race
go vet ./...
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out | tail -1  # must show ≥90%
```

### Acceptance Criteria

All items in the "Final Acceptance Criteria" section of the master objective (item 46) must be checked.

---

## Known Issues Registry

Track discovered bugs and deviations here.

| ID | Severity | Description | Phase | Status |
|---|---|---|---|---|
| BUG-001 | Medium | Service account sessions create index under `uuid.Nil` userID in Redis, sharing one index entry | Phase 1 | FIXED — `contrib/redis/session_store.go` skips index when `ServiceAccountID != uuid.Nil` |
| BUG-002 | Medium | `Session.Metadata` field missing — no extensible per-session data | Phase 1 | FIXED — `auth/session.go` adds `Metadata map[string]any` |
| BUG-003 | High | Compiler has no cross-entity dependency graph — circular FKs caught only at migration runtime | Phase 2 | FIXED — `compiler/graph.go` implements Kahn's cycle detection + topological sort; `CompiledSchema.Graph` populated in Phase 2.7 |
| BUG-004 | Medium | `ActionRuntime` interface has no concrete implementation — actions receive narrower context | Phase 3 | PARTIAL — pre-existing `action_runtime_impl.go` provides `RuntimeFactory`/`defaultActionRuntime` via `EntityDriver`; `NoopActionCache` added for test isolation |
| BUG-005 | High | Temporal client is nil at runtime — all `WorkflowTrigger` declarations are decorative | Phase 10 | OPEN |
| BUG-006 | Medium | Finance module NOT imported — Phase 1 readiness test cannot run | Phase 11 | OPEN |
| BUG-007 | Low | Wire in go.mod as dead weight — no wire_gen.go | Phase 1 | PARTIAL — tracked for go.mod cleanup |
| BUG-008 | Medium | BulkCreate is sequential within TX — n individual INSERTs, not batch | Phase 7 | OPEN |
| BUG-009 | Low | `Session.RedisKey()` deprecated but not removed | Phase 7 | OPEN |
| BUG-010 | Medium | API middleware `auth.go` accepts IAM concrete type, not interface | Phase 1 | PARTIAL — `auth.SessionValidator` interface created; middleware update deferred (needs IAM impl to satisfy interface first) |
| BUG-011 | Low | Registry naming confusion: 3 registry objects with similar names | Phase 1 | OPEN |
| BUG-012 | Medium | PageBuilderSet invocation unverified in SDUI engine | Phase 9 | OPEN |
| DOC-001 | Low | CLAUDE.md §11 references SQLC gap — SQLC was never in go.mod | Phase 0 | OPEN |
| DOC-002 | Low | CLAUDE.md §4 states Wire for DI — contradicts actual manual DI in main.go | Phase 0 | OPEN |
| AUDIT-001 | CORRECTION | Previous audit claimed Session.UserID = *uuid.UUID → nil panic. WRONG. It is uuid.UUID (value). Bug is logic (shared index), not nil dereference. | Phase 0 | RESOLVED |

---

## Progress Summary

| Phase | Status | Blocker |
|---|---|---|
| Phase 0 — Documentation + Task Tracking | COMPLETE | — |
| Phase 1 — Framework Core | IN PROGRESS | — |
| Phase 2 — Compiler Dependency Graph | COMPLETE | — |
| Phase 3 — Runtime Pipeline Hardening | COMPLETE | — |
| Coverage — compiler + runtime ≥90% | COMPLETE | — |
| Phase 4 — Filter + Query Builder | COMPLETE | — |
| Phase 5 — Migration Generation | NOT STARTED | Phase 4 |
| Phase 6 — CLI | NOT STARTED | Phase 5 |
| Phase 7 — Contrib Infrastructure | NOT STARTED | — |
| Phase 8 — Framework Platform Entities | NOT STARTED | Phase 5, 7 |
| Phase 9 — API / OpenAPI / SDUI / Docgen | NOT STARTED | Phase 8 |
| Phase 10 — Reports / Import / Export / Scheduling | NOT STARTED | Phase 4, 8 |
| Phase 11 — ERP Entity Initialization | NOT STARTED | Phase 10 |
| Phase 12 — Extraction / Public API / Hardening | NOT STARTED | Phase 11 |
