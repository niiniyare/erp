# Awo Framework — Final Architecture Sign-Off Review
## Pre-v1.0 Freeze | Internal Consistency Audit | 2025-07-06

**Classification:** Final Architecture Sign-Off (ADR-SIGNOFF-001)
**Precondition:** All prior review recommendations implemented.
**Scope:** Internal consistency only. No new features. No redesign.
**Verdict format:** APPROVED FOR v1.0 | APPROVED WITH REQUIRED CHANGES | NOT READY FOR v1.0

---

## 1. Cross-Layer Consistency

### 1.1 Package Responsibility Audit

Every package must have exactly one reason to change.

| Package | Declared Responsibility | Violation? |
|---|---|---|
| `def/` | Type definitions for entity metadata | **VIOLATION:** `def/` contains both type definitions AND validation logic. Validation belongs in the compiler. When validation rules change, `def/` changes — but `def/` should only change when the type system itself changes. Separate `compiler/validate/` from `def/`. |
| `filter/` | Filter DSL types and wire format | Clean. |
| `driver/` | Infrastructure interface contracts | **VIOLATION:** `driver/` contains both interfaces AND default no-op implementations. No-op implementations are test helpers, not driver contracts. Move them to `driver/testing/` or `internal/testutil/`. |
| `registry/` | Module registration and compilation | **VIOLATION:** Registry is doing three things: (1) accepting registrations, (2) resolving dependencies, (3) compiling artifacts. These are different reasons to change. Separate into `registry/` (registration), `resolver/` (dependency resolution), `compiler/` (artifact compilation). |
| `framework/bootstrap/` | Application startup wiring | **VIOLATION:** Bootstrap knows about specific driver implementations. Bootstrap should only accept `driver.EntityStore`, `driver.SessionStore`, etc. — not concrete types. |
| `internal/platform/` | Platform module implementations | Clean. Same patterns as business modules, correct. |
| `internal/shared/` | Cross-cutting utilities | **VIOLATION:** "Utilities" is not a responsibility. This package will accumulate every miscellaneous thing. Audit every export in `internal/shared/` and move each to the package it actually belongs to. Shared packages that contain more than one concept are technical debt. |

**Hidden cycles (dependency analysis):**

The most dangerous cycle pattern in Go framework code:

```
def/ → compiler/ (for validation helpers) → def/ (for types)
```

If `def/` imports anything from `compiler/` for validation, a cycle forms. The rule must be absolute: `def/` imports nothing from the framework. It is the root package. Nothing in `def/` may import `compiler/`, `registry/`, `runtime/`, `driver/`, or `filter/`. Validation logic that requires knowledge of the compiled graph lives in `compiler/`, not `def/`.

Secondary cycle risk:

```
runtime/ → driver/ → runtime/ (for context types)
```

If driver interfaces reference runtime context types (e.g., a custom context type defined in `runtime/`), drivers become coupled to the runtime. All context types passed through driver interfaces must be from `def/` or standard library only. `context.Context` is the only acceptable runtime context type crossing a driver boundary.

Tertiary cycle risk:

```
framework/bootstrap/ → internal/platform/iam/ → framework/bootstrap/
```

Bootstrap that imports platform packages to wire IAM will create a cycle if platform packages need anything from bootstrap. Platform packages must declare their dependencies through `def/` interfaces, not through bootstrap types.

### 1.2 Abstraction Leakage Audit

**Leak 1 — SQL grammar in Filter DSL.**
If `filter.Eq("tenant_id", x)` compiles differently for PostgreSQL vs. SQLite vs. a document store, the filter semantics are implementation-neutral. But if any filter type in `filter/` uses SQL-specific concepts (e.g., `filter.RawSQL()`), SQL grammar has leaked into the public API. `filter.RawSQL()` must not exist. If it does, it is a permanent public API that ties the framework to SQL databases.

**Leak 2 — PostgreSQL error codes in domain errors.**
If the error returned by `EntityRepository.Create()` mentions PostgreSQL error code `23505`, the storage driver has leaked through the repository interface. Domain errors must be expressed in domain vocabulary only: `ErrDuplicateRecord`, not `pgError 23505 unique_violation`. The driver translates. The interface never exposes transport.

**Leak 3 — Temporal workflow ID format in entity records.**
If entity records store Temporal workflow IDs as a field visible to the domain layer, the workflow engine implementation detail has leaked. The domain layer should see `WorkflowRef` — an opaque token the framework uses internally. The concrete Temporal workflow ID is `WorkflowRef.ExternalID()`, accessible only through the workflow driver.

**Leak 4 — Redis key format in session tokens.**
Session tokens returned to clients must be opaque. If any part of the framework constructs a Redis key by concatenating a session token with a prefix (e.g., `"session:" + token`), the session storage strategy is part of the token contract. If Redis is replaced, the key format changes, and existing tokens become invalid. Session tokens must be opaque to all layers above the session driver.

**Leak 5 — amis JSON structure in PageBuilder return types.**
If `PageBuilderFunc` returns `map[string]any` that happens to be amis JSON, the amis schema format has leaked into the public API. `PageBuilderFunc` must return an abstract SDUI type (`def.UIPage`) that the renderer serializes. This was identified as a critical gap in v2.md review. Confirming: if not fixed, it is the most expensive abstraction leak in the framework.

---

## 2. Architectural Invariants

Every invariant listed here must be provably enforceable — not by convention, not by documentation, but by the type system, the compiler, or a startup-time panic that is deterministic and immediate.

### 2.1 Invariant Catalog

**INV-001: Registry is frozen before compilation begins.**
- Enforcement: `Registry` type has two states. Before `Compile()`, mutations are allowed. After `Compile()`, all mutation methods panic immediately. This is not advisory — it is a hard contract.
- Risk: If `Compile()` can be called multiple times, the invariant is undefined. `Compile()` must be idempotent on the second call (no-op, returns same schema) OR must return an error on the second call. It must never silently recompile with a different result.
- Current enforcement: Unknown. Must be verified.

**INV-002: CompiledSchema is immutable after production.**
- Enforcement: `CompiledSchema` is a value type (struct, not pointer) OR a pointer type whose fields are all unexported and accessible only through methods. No method on `CompiledSchema` may modify state. All methods are pure reads.
- Risk: If `CompiledSchema` is a `*CompiledSchema`, callers can retain a pointer and observe mutations if any are made. `CompiledSchema` must be designed so there is no mechanism to mutate it after `Compile()` returns, even if the caller tries.

**INV-003: Tenant context set before any entity store operation.**
- Enforcement: The entity store driver must check that `app.current_tenant_id` is set before executing any query. If not set, it returns `ErrMissingTenantContext`, not a silent query with no RLS applied.
- Risk: A background goroutine that acquires a database connection and runs a query without setting tenant context will silently execute outside RLS. Background jobs must set tenant context explicitly, and the driver must verify it is set.
- This invariant cannot be enforced by the Go type system alone. It requires a runtime check inside the driver. The check must be unconditional — not skippable by callers.

**INV-004: Hook chains execute in declaration order.**
- Enforcement: The compiled hook chain is an ordered slice. Appending to it after compilation is not possible (see INV-001). The slice index is the execution order.
- Risk: If hooks from multiple modules are merged (module A hooks + module B hooks on the same entity), the merge order must be deterministic. Module registration order must be deterministic (topological sort from dependency graph). If module load order is non-deterministic, hook execution order is non-deterministic.
- Required: The compiler must produce the same hook chain given the same module graph, regardless of which module's `init()` fires first. This requires post-registration sorting, not rely-on-init-order.

**INV-005: Policy chains are evaluated in full, not short-circuit on first match.**
- Clarification needed: Is the policy model "first matching policy wins" (like Casbin default) or "all policies must pass" (strict AND semantics) or "any policy may grant" (OR semantics)?
- This is a semantic invariant, not an implementation detail. The semantics must be locked before v1.0 and documented as permanent. Changing from OR to AND semantics after third-party modules write policies is a security regression.
- Recommendation: AND semantics for operation-level permissions (all declared roles must be satisfied). OR semantics for row-level filters (filters from all policies are composed with AND — the intersection of all allowed rows).

**INV-006: Migrations apply in strict version order.**
- Enforcement: The migration runner reads all applied migration versions from the schema_migrations table, sorts remaining unapplied migrations by version number, and applies in order. No gaps allowed. No out-of-order allowed.
- Risk: If two modules contribute migrations with the same timestamp prefix (14-digit Unix timestamp), their ordering is undefined. Migration filenames must include a module namespace: `{timestamp}_{module}_{description}.up.sql`. The runner sorts by timestamp, then by module name for same-timestamp files.

**INV-007: WorkflowStart is always outside the database transaction.**
- Enforcement: The entity lifecycle engine must guarantee this structurally — not by convention. If workflow triggers are dispatched inside `after_save` (which runs inside the TX), and `after_save` is documented as "inside TX," then workflow dispatch inside `after_save` violates INV-007.
- The correct position is after the TX commits. The lifecycle must make this structurally impossible to violate: workflow dispatch is a separate phase that receives the committed record, not a hook.

**INV-008: No request succeeds without tenant context (except platform-admin operations).**
- Enforcement: Middleware pipeline must enforce this. The tenant resolution middleware must return 400 if no tenant can be identified, before any request reaches the handler layer. Platform-admin operations are identified by a separate route namespace (`/platform/`) that bypasses tenant resolution.
- Risk: If any auto-generated route can be reached without tenant context, the invariant is broken. Route registration must verify that every non-platform route has tenant resolution middleware applied.

**INV-009: Custom field writes do not corrupt system field columns.**
- Enforcement: Custom fields are stored in `custom_fields jsonb`. System fields are stored in typed columns. A write operation that attempts to write a custom field into a system column (by manipulating the field name to match a system column) must be rejected. The store driver must maintain a whitelist of writable column names per entity and reject any write targeting an unlisted column.
- Risk: SQL injection through field names. A custom field named `; DROP TABLE finance_invoice; --` must be rejected at field name validation, not at SQL execution. Field names must be validated against `^[a-z][a-z0-9_]{0,63}$` at registration time.

**INV-010: Module dependency graph is acyclic.**
- Enforcement: The resolver must run cycle detection before accepting any module registration as valid. A cycle in the module dependency graph means compilation order is undefined. The resolver must detect cycles and report them as fatal compile errors with the full cycle path in the error message.

**INV-011: EntityDefinition.Name is globally unique at compile time.**
- Enforcement: The registry checks for duplicate names during registration. Duplicate registration is fatal. The error message must include both the new module and the existing module that already registered that name.
- Risk: If modules are registered in parallel (concurrent `init()` goroutines), the uniqueness check has a race condition. Module registration must be serialized. A `sync.Mutex` in the registry is sufficient. Alternatively, use `sync.Once` around the registration phase.

**INV-012: The compiled route table has no ambiguous patterns.**
- Enforcement: The compiler verifies that no two routes produce the same pattern after parameter substitution for any input. `POST /api/v1/entities/{type}/{id}/{action}` must not overlap with `POST /api/v1/entities/{type}/{id}` for any value of `{action}`. The router must be validated deterministic: same URL always routes to same handler.

### 2.2 Unenforced Invariants (Current Gaps)

These invariants are required but cannot currently be enforced by the type system:

- **INV-003** requires a runtime check inside every driver. Currently documented as convention.
- **INV-004** requires deterministic module load order. Currently depends on `init()` order, which is undefined across packages in Go.
- **INV-005** requires a semantic decision that appears not yet formally recorded.
- **INV-007** requires structural separation of workflow dispatch from `after_save` hook phase.

All four must be enforced before v1.0.

---

## 3. Hidden Contradictions

### CON-001: Immutable Compiled Schema vs. Per-Tenant Custom Fields

The compiled schema is immutable (INV-002). But tenants add custom fields at runtime without a redeploy. These are not the same thing — but they look the same to query code.

The contradiction: field resolution at query time must handle both compiled fields (in the schema) and runtime custom fields (in the metadata module's `custom_field_defs` table). If the query builder reads field definitions from the compiled schema, it will not see runtime custom fields. If it reads from both, the compiled schema is no longer the single source of truth.

**Resolution required before v1.0:** Define a `FieldResolver` interface that the query builder uses to resolve fields at query time. The compiled schema provides the base implementation. The metadata module provides an implementation that overlays custom fields on top. The query builder calls `FieldResolver.Resolve(entityName, fieldName)` — it does not directly access the compiled schema. This maintains INV-002 (the compiled schema does not change) while allowing runtime extensions.

### CON-002: Driver Abstraction vs. PostgreSQL-Specific RLS

The driver abstraction promises any storage backend can implement `driver.EntityStore`. But tenant isolation is enforced by PostgreSQL RLS via `set_tenant_context()` — a PostgreSQL-specific stored procedure. A non-PostgreSQL driver cannot implement RLS the same way.

The contradiction: if RLS is the isolation mechanism and RLS is PostgreSQL-specific, then "any storage backend" is not true for tenant isolation. A SQLite driver cannot have the same isolation guarantee. A DynamoDB driver cannot.

**Resolution required before v1.0:** The driver contract must formally document the isolation guarantee required of any `EntityStore` implementation, independent of how it is achieved:

> "Every `EntityStore` implementation must guarantee that a query executing in the context of tenant A never returns or modifies records belonging to tenant B, regardless of the query parameters. The mechanism for achieving this guarantee is implementation-specific. PostgreSQL RLS is the reference implementation."

This makes the guarantee the invariant, not the mechanism. A conformance test in `driver/conformance/` must verify the guarantee: given two tenants with records, queries for tenant A must never return tenant B's records, even with adversarial filter inputs.

Without this clarification, the driver abstraction and the RLS enforcement are in direct contradiction — one promises mechanism independence, the other hardcodes a mechanism.

### CON-003: Workflow Abstraction vs. Outbox Pattern

The workflow driver abstracts Temporal. The outbox pattern is the reliability bridge between entity writes and workflow starts. The outbox emits events. The workflow driver subscribes to those events (or the outbox processor calls the workflow driver directly).

The contradiction: the outbox processor must know which workflow driver to call to start a workflow. If the workflow driver is injected into the Runtime, and the outbox processor is a separate background goroutine, the outbox processor must either:
(a) Import the runtime and call the workflow driver through it, or
(b) Have the workflow driver injected directly into the outbox processor.

Option (a) creates a cycle: outbox processor → runtime → outbox processor (for lifecycle management).
Option (b) works but means the outbox processor is coupled to the workflow driver interface.

**Resolution required before v1.0:** The outbox processor must be a driver itself. `driver.OutboxProcessor` accepts a `driver.WorkflowStarter` (a minimal interface with one method: `StartWorkflow(ctx, ref WorkflowRef) error`). The runtime injects both. No cycle, no direct import of the workflow driver implementation.

### CON-004: Startup Order vs. Tenant-Specific Compilation

The startup order is: config → PostgreSQL → Redis → EntityRegistry → compile → Fiber → Temporal.

But: some features require per-tenant schema customization (custom fields, tenant-overridable naming series prefixes, per-tenant page overrides). These customizations exist in PostgreSQL (in tenant-scoped tables). They cannot be compiled into the static CompiledSchema because the schema is compiled before requests begin.

The contradiction: compilation happens before tenant data is available, but tenant customizations affect entity behavior at request time.

**Resolution (must be locked before v1.0):** Draw a hard line between two categories:

- **Schema-time customization**: What fields exist, what types they have, what edges connect them. These are compiled into `CompiledSchema`. Tenants cannot change these without a migration. This is correct and intentional.
- **Request-time customization**: Custom field values, naming series prefix overrides, page layout adjustments. These are NOT in `CompiledSchema`. They are read from the database at request time (with caching). They are a runtime overlay, not a schema extension.

The `FieldResolver` from CON-001 handles request-time field overlays. Naming series prefixes are resolved at sequence-generation time, not compile time. Page layout adjustments are injected into the SDUI tree at schema-serve time.

This line must be formally documented. Every customization mechanism must declare which side of the line it is on. If this is ambiguous, the invariant "compiled schema is the source of truth" becomes meaningless.

### CON-005: Hook Registration Ordering vs. Module Load Ordering

Hooks must execute in declaration order (INV-004). Module load order is determined by dependency graph topological sort. But within a module, `init()` functions can register hooks in arbitrary order depending on how the Go runtime processes them.

The contradiction: the declared hook order within a module may differ from the registration order if multiple `init()` functions in the same module register hooks. Go's spec says: within a package, `init()` functions are called in the order they appear in source files (sorted by filename). This is deterministic but non-obvious.

**Resolution required before v1.0:** Each module must have exactly one hook registration function, called from exactly one `init()`. Multiple `init()` calls in one module that each register hooks are forbidden. The linter must enforce this. The registration order must be explicit in source, not inferred from file sort order.

### CON-006: Transactional Integrity vs. Distributed Workflow Durability

Entity writes are transactional (atomic). Workflow starts are durable (Temporal guarantees eventual execution). The outbox bridges them: record is written + outbox entry written in one transaction; outbox processor starts workflow after commit.

The contradiction: if the outbox processor fails after dequeuing an entry but before calling Temporal, the entry may be redelivered (at-least-once). The workflow must be idempotent. But workflow IDs in Temporal are `{tenant-uuid}.{entity-type}.{record-id}.{event}` — if a workflow with that ID is already running, starting it again returns the existing workflow (Temporal's deduplication). This is correct behavior.

But: if the same event triggers two different workflow types (both `EventOnSubmit`), they share the same workflow ID pattern and Temporal cannot distinguish them. Two workflows for the same event on the same record would collide.

**Resolution required before v1.0:** Workflow ID format must include the workflow function name: `{tenant-uuid}.{entity-type}.{record-id}.{event}.{workflow-function}`. This makes IDs globally unique per workflow type and prevents collision when multiple workflows trigger on the same event.

### CON-007: SDUI Neutrality vs. amis Page Cache

The SDUI layer (as corrected by prior reviews) produces abstract page schemas. The renderer converts them to amis JSON. But the Redis cache key is `page:{entity}:{version}:{tenant}`.

What is cached — the abstract schema or the amis JSON? If the abstract schema is cached, the renderer runs on every cache hit (slow). If the amis JSON is cached, the cache is renderer-specific and must be invalidated when the renderer changes.

The contradiction: a renderer-neutral SDUI system has a renderer-specific cache.

**Resolution required before v1.0:** Cache the abstract schema. The renderer is fast (pure JSON serialization, no logic). The abstract schema cache is renderer-agnostic and survives renderer upgrades. Cache key remains `page:{entity}:{version}:{tenant}` — the "version" incorporates the entity schema version. When the abstract schema changes, the version changes, cache is automatically invalidated.

### CON-008: Single Binary Compilation vs. Module Version Diversity

The compilation-unit model compiles all modules into one binary. But the module ecosystem allows third-party modules. If module A is built against `def/` v1.2 and module B is built against `def/` v1.1, they cannot coexist in the same binary — they require different versions of the same package.

This is the classic diamond dependency problem. In Go's module system, you can have at most one version of a given module in a build graph (unless using major version suffixes).

The contradiction: the promise that the ecosystem supports thousands of modules is incompatible with the single-binary compilation model unless all modules are always built against the same `def/` version.

**Resolution required before v1.0 (policy, not code):** The `def/`, `filter/`, and `driver/` packages use major version suffixes (`def/v2`, `def/v3`) for breaking changes. Within a major version, all modules must be compatible (this is the standard Go module contract). The framework's LTS policy commits to supporting two major versions simultaneously. Module authors declare `requires_def: ">=1.0.0, <2.0.0"` — if they require v2, they must be recompiled against it. There is no mechanism to have v1 and v2 modules coexist. This is intentional and must be explicitly documented as a constraint.

---

## 4. Long-Term Maintenance Burden

### 4.1 Five-Year Horizon (2030)

**B1 — The two-stack problem.** `internal/` (legacy Wire DI, production today) and `framework/` (clean abstraction). At five years, engineers who joined after v1.0 will not know why two stacks exist. They will pick one arbitrarily. The stacks will diverge. Features added to one will not appear in the other. **Resolution:** The legacy stack must be eliminated before v1.0 or formally deprecated with a hard removal date. Parallel stacks that are both "production" are an exponential maintenance burden.

**B2 — FieldDef struct growth.** Without the extension map pattern (from prior reviews), FieldDef will have ~40 fields by 2030. Every new subsystem adds fields. At 40 fields, the struct is unmaintainable — no engineer can hold all field interactions in their head. This compounds every review, every feature, every bug fix involving fields.

**B3 — Test surface against the compiled schema.** Every test that asserts against compiled schema content (field counts, route patterns, hook chain lengths) will break when any entity definition changes. At 100 modules, the schema changes weekly. The test suite becomes a maintenance treadmill.

### 4.2 Ten-Year Horizon (2035)

**B4 — Go version evolution.** Go's type system will evolve significantly by 2035. Generics, added in Go 1.18, are still maturing. New language features (sum types, error handling improvements, improved type inference) may make the current `EntityRepository[T Entity]` generic interface significantly more expressible. APIs locked at v1.0 in pre-evolution Go idiom will look archaic by 2035. This is unavoidable — but the impact is minimized if the public API surface is small and well-bounded.

**B5 — Temporal API evolution.** Temporal's Go SDK has changed significantly between versions. If the workflow driver interface is too thin (exposes too little) or too thick (exposes Temporal-specific concepts), it will need to be updated with every major Temporal SDK release. The driver interface must abstract exactly what workflows need from a durable execution engine — no more, no less.

**B6 — PostgreSQL RLS limitations at scale.** At 1 million tenants and 100 million rows per table, RLS policy evaluation overhead becomes measurable. PostgreSQL's RLS is evaluated per-row during query execution. The stored procedure `set_tenant_context()` is called per-transaction. At 100,000 requests/second, this is 100,000 `set_config()` calls/second — entirely within PostgreSQL's capability. But per-row policy evaluation at 100M rows/table requires careful index design. The framework's migration templates must enforce that every RLS-protected table has an index on `(tenant_id, ...)` as the leading columns. Without this, sequential scans with RLS will become the primary performance bottleneck.

**B7 — The `internal/shared/` junk drawer.** Without aggressive governance, `internal/shared/` accumulates every utility that "doesn't fit elsewhere." By 2035, it will be a 50-file package with zero coherent responsibility. Engineers will import it for one function and get everything else as a side effect. Subdivide or eliminate before v1.0.

### 4.3 Twenty-Year Horizon (2045)

**B8 — The EntityDefinition field count.** EntityDefinition today has: name, module, label, label_plural, fields, edges, hooks, policies, actions, workflow_triggers, pages, permissions. At 2045, if not disciplined, it will have 50+ fields. Every new platform capability (AI hints, compliance tags, data lineage, archival policies, sync policies, replication factors) will want a home in EntityDefinition. The extension map pattern must exist and must be the only acceptable way to add new metadata after v1.0. Core fields must be frozen.

**B9 — The abstract SDUI schema evolution.** The abstract SDUI schema will evolve with UI paradigms. By 2045, conversational UIs, spatial UIs, and AI-generated UIs will exist. If the SDUI schema cannot express these without a breaking change, the SDUI layer will be abandoned and replaced with direct renderer coupling — the exact problem the abstraction was designed to prevent. The SDUI schema must have a stable extension mechanism from day one.

**B10 — Workflow ID namespace collision at scale.** At 1 million tenants and 100 workflow types per entity and 1 billion records, the Temporal workflow ID space becomes very large. Temporal has limits on workflow ID length (up to 1000 characters but practically much less for readability). At `{tenant-uuid}.{entity-type}.{record-id}.{event}.{workflow-function}`, each ID is approximately 200 characters. This is fine. But if workflow IDs ever need to embed additional context, the format becomes unwieldy. Lock the format now and resist additions.

---

## 5. Extensibility Audit

For each capability, the question is: does the current architecture support adding it without modifying any existing public API?

**GraphQL — BLOCKED without changes.**
Current state: API descriptor is not abstract. Routes are generated as REST directly from EntityDefinition. Adding GraphQL requires either: (a) generating GraphQL schema from EntityDefinition at compile time, or (b) building a GraphQL-to-REST translation layer. Option (a) requires a new compiler phase (additive — OK if compiler plugin hooks exist). Option (b) is a workaround, not architecture. If abstract API descriptors exist (MUST-2 from v2.md), GraphQL is a new renderer plugin — no changes to existing code. **Verdict: Blocked until abstract API descriptor exists.**

**gRPC — BLOCKED without changes.**
Same root cause as GraphQL. gRPC service definitions (`.proto` files) must be generated from the abstract API descriptor. **Verdict: Blocked until abstract API descriptor exists.**

**WebSockets — PARTIAL.**
WebSockets require long-lived connections, push notifications, and subscription semantics. EntityDefinition's event model (outbox → workflow → delivery) is pull-based and push-unfriendly. WebSocket support requires a `SubscriptionEvent` type in the event system — a concept that does not currently exist. Adding it requires a new event category, not a change to existing event semantics. If the event driver interface has an extension point for subscription events, WebSocket support is additive. **Verdict: Requires event driver extension point. Not currently present.**

**CQRS — SUPPORTED.**
CQRS means separate read and write paths. Awo already separates reads (`Query()`, `Get()`) from writes (`Create()`, `Update()`, `Delete()`) in the EntityRepository interface. A CQRS driver could implement the read methods against a read replica or a projection store, and write methods against the primary. No interface change required. **Verdict: Supported, no changes needed.**

**Event Sourcing — BLOCKED without design.**
Event sourcing requires that the source of truth is an event log, not a row. Awo's EntityRepository is row-based. Adding event sourcing would require a new entity storage type (`EventSourced: true`) that stores events instead of row mutations and rebuilds state by replaying. This is a significant compiler and driver change. The outbox is not event sourcing — it is a side effect of a row mutation, not the primary store. **Verdict: Not supported. Requires new EntityDef storage type and driver variant. Design before v1.0 even if implementation is v2.**

**Offline Sync — PARTIAL.**
Record identity is UUID v7 (client-generatable). The EntityRepository interface could be implemented by a local store. But conflict resolution semantics are not declared anywhere. Two offline clients creating the same record produce two records with different IDs — correct. Two offline clients modifying the same record produce a conflict — no resolution strategy is defined. `ConflictStrategy` must be a declarable field on EntityDefinition before offline sync drivers can implement deterministic behavior. **Verdict: Hook points exist, conflict semantics not declared. Needs ConflictStrategy field.**

**AI Agents — PARTIAL.**
AI agents need to: query entities in natural language, execute actions by name, understand entity relationships. Filter DSL is structured data — translatable by LLMs. Actions are declared with names — callable by name. But action descriptions are human labels, not machine-readable semantic descriptions. An AI agent needs: what does this action do, what are its preconditions, what are its effects, what fields does it read/write. `ActionDef` must include a `SemanticDescription` (machine-readable: preconditions, effects, IO schema) before AI agents can use actions reliably. **Verdict: Queryable structure exists, action semantics not machine-readable. Needs SemanticDescription.**

**Streaming — BLOCKED without changes.**
`EntityRepository.Query()` returns `[]T` — a slice, not a stream. At 100M records, streaming results is essential. Adding `QueryStream(ctx, f Filter) (<-chan T, error)` to the driver interface is a new method — additive in Go if done through an optional interface:

```go
type StreamableEntityStore interface {
    driver.EntityStore
    QueryStream(ctx context.Context, f filter.Filter, opts ...QueryOption) (<-chan Record, error)
}
```

This is additive. Existing drivers ignore it. New drivers implement it. **Verdict: Supported through optional interface extension. Document pattern before v1.0.**

**Analytics — PARTIAL.**
`EntityRepository.Aggregate()` exists. But analytics often require cross-entity aggregations, time-series rollups, and window functions — none of which are expressible in the current Filter + AggregateSpec model. An `AnalyticsDriver` optional interface with `RunAnalyticsQuery(ctx, q AnalyticsQuery)` is the right extension point. **Verdict: Basic aggregation supported. Advanced analytics requires optional AnalyticsDriver interface.**

**Search Engines — SUPPORTED.**
A search driver that implements `EntityStore.Query()` by routing to Elasticsearch or similar is architecturally sound. The Filter DSL must have a well-defined semantics specification so search drivers know exactly which filters to support and which to fall back to the primary store for. **Verdict: Supported if Filter DSL semantics are formally specified (MUST-4 from v2.md).**

**Alternative Workflow Engines — SUPPORTED.**
The workflow driver interface abstracts Temporal. Any durable execution engine that can: start a workflow by function reference, signal a running workflow, and query workflow status can implement the interface. **Verdict: Supported.**

**Alternative Databases — SUPPORTED (with caveat).**
The EntityStore driver interface supports alternative databases. The caveat is tenant isolation (CON-002): non-SQL databases must prove equivalent isolation guarantees through the conformance test suite. **Verdict: Supported if conformance suite exists.**

**Alternative UI Renderers — SUPPORTED (after MUST-1).**
After the abstract SDUI schema exists, any renderer that can serialize it is a valid renderer driver. Before MUST-1 is implemented, alternative renderers are not supported. **Verdict: Blocked until abstract SDUI schema exists.**

**Alternative Auth Providers — SUPPORTED.**
The session driver and identity driver are abstracted. An OAuth2 driver, SAML driver, or passkey driver can implement the same interfaces. **Verdict: Supported.**

---

## 6. Operational Architecture

### 6.1 Rolling Deployments

Rolling deployments require that old and new binary versions can process the same requests simultaneously during the rollover window.

**Risk:** If v1.1 adds a new field to `CompiledSchema` that is read by request handlers, and v1.0 handlers do not know about it, the behavior during rollover is undefined. Old handlers processing requests that touch new fields will encounter missing data.

**Requirement:** During a rolling deployment, both versions must be able to process any request. This requires:
1. New fields added to records must be nullable (old handlers ignore null, new handlers read it)
2. New routes added in new binary must not be accessible until all instances are on new binary (traffic routing concern, not framework concern, but must be documented)
3. New compiler output that changes existing behavior must be gated behind a feature flag that is disabled until all instances are upgraded

The framework must document its rolling deployment protocol before v1.0.

### 6.2 Blue/Green Deployments

Blue/green is simpler: full cutover from old to new. The concern is database migration. If v1.1 requires a migration that the v1.0 blue deployment cannot handle, the cutover cannot be reversed.

**Requirement:** Migrations must be backward-compatible. The zero-downtime migration patterns documented in CLAUDE.md (add column nullable first, then backfill, then add constraint) must be enforced by the migration runner. A migration that adds a NOT NULL column without a default is rejected by the runner as unsafe for rollback.

### 6.3 Multi-Region

Multi-region requires that tenant data can be geographically placed (regulatory) and that operations originating in one region can be fulfilled by another in case of failover.

**Current gap:** No geo-routing, no tenant placement metadata, no cross-region consistency model. The `TenantPlacement` hook point (MUST-15 from v2.md) allows this to be added without breaking changes. Without it, multi-region is a deployment model that requires forking the application code — unacceptable for enterprise customers.

### 6.4 Disaster Recovery

Disaster recovery requires: point-in-time recovery from PostgreSQL WAL, Redis snapshot restoration, and Temporal workflow state recovery.

**Framework responsibility:** The framework must document which stores are sources of truth (PostgreSQL), which are caches (Redis — all data reconstructible from PostgreSQL), and which are external (Temporal workflow state — owned by Temporal cluster).

After a disaster recovery, Redis is empty. The framework must start correctly with empty Redis: session validation fails (users must re-authenticate — correct behavior), page schema cache rebuilds on first request, feature flag cache rebuilds on first evaluation. This must be tested in the operational runbook.

### 6.5 Schema Evolution During Operation

When a migration adds a column to a system entity table, in-flight requests that read the entity before migration and write it after migration may produce inconsistent data (old read, new write, missing column in the write).

**Requirement:** The migration runner must enforce a "safe to apply" check: a migration adding a column to a table that has active connections is safe only if the column has a default or is nullable. The runner must query `pg_stat_activity` and warn if active connections exist to tables being migrated.

### 6.6 Tenant Migration

Moving a tenant's data between PostgreSQL instances (e.g., geo-migration or shard rebalancing) requires:
1. Export all records for tenant T from source database
2. Import all records for tenant T into target database
3. Verify checksums
4. Redirect requests for tenant T to target database
5. Delete records for tenant T from source database

Steps 1–5 must be supported by the framework's CLI tooling. Step 4 requires the geo-routing mechanism (MUST-15). Without step 4, tenant migration requires downtime.

**This is an operational requirement that affects v1.0 architecture.** The tenant placement concept must be in the data model at v1.0 — specifically, the `tenants` table must have a `primary_region` field and the routing layer must read it. Even if multi-region routing is not implemented until v1.2, the data model must support it at v1.0 because adding a column to the `tenants` table after v1.0 requires a migration that all deployments must apply.

---

## 7. Go Language Architecture

### 7.1 Interfaces

**Go philosophy:** Define interfaces where they are used, not where types are implemented. Small, composable interfaces.

**Violation 1:** `EntityRepository[T Entity]` has 12 methods. This is a large interface. In Go, large interfaces are hard to mock, hard to implement for alternative backends, and hard to satisfy partially. Consider splitting:

```go
// Immutable — read-only operations
type EntityReader[T Entity] interface {
    Get(ctx context.Context, id uuid.UUID) (T, error)
    Query(ctx context.Context, f filter.Filter, opts ...QueryOption) ([]T, PageInfo, error)
    Exists(ctx context.Context, f filter.Filter) (bool, error)
    Count(ctx context.Context, f filter.Filter) (int, error)
}

// Mutable — write operations
type EntityWriter[T Entity] interface {
    Create(ctx context.Context, input CreateInput) (T, error)
    Update(ctx context.Context, id uuid.UUID, input UpdateInput) (T, error)
    Delete(ctx context.Context, id uuid.UUID) error
}

// EntityRepository = EntityReader + EntityWriter + Aggregate + Tx
type EntityRepository[T Entity] interface {
    EntityReader[T]
    EntityWriter[T]
    Aggregate(ctx context.Context, f filter.Filter, spec AggregateSpec) (AggregateResult, error)
    WithTx(ctx context.Context, fn func(ctx context.Context, repo EntityRepository[T]) error) error
}
```

Read-only hooks (like `OnRead`) receive `EntityReader[T]`, not the full repository. This prevents a read hook from triggering writes, which would be a silent, untestable side effect.

**Violation 2:** Interface definitions in `def/` (or equivalent) that reference concrete types from `driver/` or `runtime/`. Any interface that references a concrete type has leaked its dependency. Every interface must only reference types from the same package or packages lower in the dependency hierarchy.

### 7.2 Package Boundaries

Go packages are the unit of encapsulation. The package boundary determines what is visible. Using `internal/` packages correctly is critical.

**Rule:** Every package exported to third-party module authors must be in the module root (not `internal/`). Every package that is implementation detail must be in `internal/`. The framework's public API is exactly the set of non-internal packages.

**Current risk:** If `internal/platform/` packages are not truly internal to the framework binary (i.e., third-party modules might need to reference platform types directly), they should not be in `internal/`. But if platform packages are framework internals, they should never be imported by third-party modules. Clarify this before v1.0.

### 7.3 Constructors

Go convention: constructor functions are named `New{Type}()` and return the type or `(*Type, error)`. Never return a partial object that requires additional configuration calls before use.

**Violation pattern:** A registry that is valid to call `Register()` on before `Compile()` but not after is a two-state object whose validity depends on call sequence, not type. The Registry must be designed so the invalid state is represented by a different type:

```go
// Pre-compile: accepts registrations
type RegistryBuilder struct { ... }
func (rb *RegistryBuilder) Register(def *def.EntityDefinition) *RegistryBuilder { ... }
func (rb *RegistryBuilder) Compile() (*CompiledSchema, error) { ... }

// Post-compile: immutable, no registration methods
type CompiledSchema struct { ... }
```

`RegistryBuilder` and `CompiledSchema` are different types. The method set of `CompiledSchema` does not include `Register()`. The invalid state (registering after compile) is a compile-time error, not a runtime panic.

### 7.4 Context Propagation

Every function that interacts with I/O (database, cache, network, workflow engine) must take `context.Context` as its first parameter. This is Go's mechanism for cancellation, timeout, and value propagation.

**Violation to watch for:** Values stored in context that are required for correct behavior (not just optimization). If tenant ID is stored in context and is required for RLS (and it is, based on the architecture), then a function that runs correctly without tenant context in tests but fails in production is hiding a required dependency in a dynamic lookup. Tenant ID must be explicit in the driver call, not hidden in context, OR the driver must verify context contains it and fail fast if not.

The architectural ruling: `context.Context` carries observability values (trace ID, request ID) and cancellation signals. It does NOT carry required business data (tenant ID, user ID). Required business data is passed as explicit parameters. If this rule is violated — if tenant context is pulled from `context.Context` rather than passed explicitly — the function's required inputs are invisible at the call site.

**Contradiction with current architecture:** The entire RLS system relies on `set_tenant_context()` being called as a side effect of query execution, which reads the tenant ID from context. This is a violation of the rule above. Tenant ID is required for correctness and it is hidden in context.

This is a known trade-off, not an oversight. The trade-off is: explicit tenant ID in every driver call (verbose, correct) vs. tenant ID in context (implicit, convenient, enforced by middleware). The framework must pick one and document it. If context is chosen, the middleware enforcement must be airtight (INV-008).

### 7.5 Generics

`EntityRepository[T Entity]` is a generic interface. This is correct use of Go generics. However:

**Risk:** The `Entity` type constraint must be stable. If `Entity` constraint adds methods after v1.0, all existing entity implementations break. The constraint must be frozen: whatever methods `Entity` requires, they are requirements on all entity types forever.

**Anti-pattern to avoid:** Using generics where interface dispatch is sufficient. If `T` in `EntityRepository[T Entity]` is never used to call a method on `T` (only `entity.Record` methods are called), then the generic parameter adds complexity without benefit. Audit actual usage of `T` in the repository implementation.

### 7.6 Global State

Go's `init()` functions and package-level variables are global state. The framework uses `init()` for entity registration. This is the accepted trade-off for ergonomics. But:

**Rule:** Package-level variables in framework packages must be either: (a) constants (truly immutable) or (b) initialized exactly once via `sync.Once`. Any package-level variable that is written after `init()` completes is a data race in concurrent programs.

The registry, if implemented as a package-level variable that `Register()` modifies, requires a mutex around all mutations. This is correct for the registration phase. After `Compile()`, the registry must become read-only — and this must be enforced by the mutex state machine (after `Compile()`, the mutex is no longer acquired for reads but any write attempt returns an error or panics).

---

## 8. Database Architecture

### 8.1 Migration Model Completeness

golang-migrate handles ordered migrations correctly. The critical gap: no migration validation pass in the framework compiler.

Before any migration is accepted as part of the build, the compiler must:
1. Parse the SQL and verify it is syntactically valid PostgreSQL
2. Verify the migration does not contain `DROP TABLE`, `DROP COLUMN`, or `TRUNCATE` without explicit human override flag
3. Verify that every new tenant-scoped table has RLS enabled and a tenant isolation policy (auto-check for `FORCE ROW LEVEL SECURITY` and `CREATE POLICY tenant_isolation`)
4. Verify that every `CREATE INDEX` uses `CONCURRENTLY` (no table locks in production)
5. Verify the down migration reverses exactly the operations of the up migration (structural check, not semantic)

These checks run at build time, not at deploy time. Finding a missing RLS policy at build time prevents a security incident in production.

### 8.2 Transaction Boundaries

The entity lifecycle defines when transactions begin and end. The architectural invariant is:

- `before_save` hook: outside transaction
- `persist`: inside transaction (TX begins)
- `after_save` hook: inside transaction
- TX commits
- workflow dispatch: outside transaction

**Gap:** BulkCreate and BulkUpdate operations — are they one transaction per record or one transaction for all records? The interface must be unambiguous. `BulkCreate` must be atomic (all succeed or all fail) or document explicitly that it is non-atomic (partial success is possible). Partial success in a financial ERP is a catastrophic data state. Default must be atomic. Non-atomic bulk operations require an explicit `BulkCreatePartial` method name to communicate the semantics.

### 8.3 Optimistic Locking

No mention of optimistic locking in the architecture. For concurrent edits to the same record (two users editing the same invoice simultaneously), last-write-wins produces data loss.

**Required before v1.0:** Every system entity must have a `version int` column (or `updated_at timestamptz` used as ETag). `Update()` must accept an `IfVersion(n)` query option. If the record's version has changed since the caller read it, `Update()` returns `ErrOptimisticLockConflict` (HTTP 409). The UI must handle 409 by refreshing and asking the user to re-apply their change.

Without optimistic locking, concurrent edits silently corrupt data. This is not observable in development (single user) and catastrophic in production (concurrent users).

### 8.4 JSONB Strategy

Custom entities store data in JSONB. The GIN index on the JSONB column enables efficient key-value queries. However:

**Risk 1 — Unbounded JSONB growth.** A custom entity with 100 custom fields and 10 million records stores 100 × 10M = 1B JSON values in one column. At typical JSON overhead, this is 50–100 GB per table before data content. JSONB has an 8 KB per-value decompression overhead. Large JSONB documents degrade query performance.

**Requirement:** Custom entities must declare `MaxFields int` in their definition. The framework must enforce this limit at write time. Default: 200 fields per custom entity. Maximum: 500. Above 500 custom fields, the entity must be escalated to a system entity.

**Risk 2 — JSONB migration.** Renaming a custom field that is stored as a JSONB key requires updating every record that contains that key. At 10M records, this is a long-running UPDATE that locks the table. The framework's field renaming API must generate a migration that uses `jsonb_set` with batched updates, not a single UPDATE statement.

### 8.5 Connection Lifecycle

PgBouncer in transaction mode is the required connection pooling strategy. This means:
- Session-level Postgres features (prepared statements, advisory locks, `SET` commands that persist across transactions) do not work
- `set_tenant_context()` sets `app.current_tenant_id` with `TRUE` (transaction-local) — this is correct for transaction mode
- Prepared statements cannot be used across connections — the framework must not assume prepared statement caching at the application level

**Architectural constraint to document:** The framework assumes PgBouncer in transaction mode. Any deployment using session mode will have RLS failures (tenant context leaks across transactions). The connection pool configuration must be validated at startup: if the framework detects session mode (by checking `SHOW transaction_isolation` and comparing with expected `BEGIN` behavior), it must refuse to start.

---

## 9. Distributed Systems Architecture

### 9.1 Eventual Consistency Boundaries

The outbox guarantees at-least-once delivery of events to the workflow engine. Within a single tenant's operations, this produces eventual consistency: the entity record is committed, the workflow will eventually start. The gap between commit and workflow start is the inconsistency window.

**Consumers must know their consistency level.** Awo must document:
- EntityRepository.Get() reads from primary — strong consistency
- EntityRepository.Query() with read-replica option — eventual consistency (replica lag)
- Workflow input data — snapshot consistency (data at workflow start time, may be stale by the time activity runs)

These are not implementation details. They are consistency contracts that affect how applications are written. Changing them after v1.0 breaks application correctness assumptions.

### 9.2 Idempotency Requirements

Every operation in the framework that crosses a network boundary must be idempotent. Concretely:

- `EntityRepository.Create()` — not idempotent by default. If the request is retried after a network timeout, a duplicate record is created. The framework must support idempotency keys on Create: `Create(ctx, input, idempotency.Key("client-generated-uuid"))`. If a record with that idempotency key was already created, return the existing record.
- `WorkflowStarter.StartWorkflow()` — idempotent by design (Temporal deduplication on workflow ID). This is correct.
- Outbox processor delivery — at-least-once, so consumers must be idempotent. The framework must provide an idempotency helper that any activity implementation can use.

**The idempotency key for Create must be:** the client-provided UUID v7 record ID. If the client generates the ID before the Create call (which UUID v7 enables), the ID itself is the idempotency key. This is the cleanest model. `CreateInput` must include an optional `ID uuid.UUID` field. If provided, the framework uses it. If not, it generates one.

### 9.3 Clock Assumptions

Temporal workflows must not use wall clock time (`time.Now()`). This is already a documented rule. But the framework itself uses wall clock time for: `created_at`, `updated_at`, `deleted_at` timestamps on records; session expiry; rate limiter windows.

**Assumption that must hold:** The wall clocks of all application servers and the database server are synchronized to within 1 second (via NTP). If this assumption is violated, `updated_at` timestamps will be wrong relative to each other, session expiry may be premature, and rate limiter windows will be skewed.

**Requirement:** The framework must verify at startup that the application server's clock is within 5 seconds of the database server's clock (by comparing `NOW()` result from PostgreSQL with `time.Now()` in Go). If the drift exceeds 5 seconds, the framework logs a critical warning and optionally refuses to start (configurable).

### 9.4 Cache Invalidation

The current cache invalidation strategy:
- Page schemas: 5-minute TTL, invalidated on permission change or feature flag change
- Feature flag results: 5-minute TTL, invalidated on flag change
- Session tokens: exact-match key, invalidated on logout or expiry

**Gap 1 — Permission change propagation.** If a role is revoked from a user, the page schema cache for that user still contains the old schema for up to 5 minutes. During this window, the user sees UI elements they should no longer see. The elements are absent from the schema if permissions are rechecked at schema-serve time. Confirmed: the architecture says "permission-gated elements are absent from schema" and "permission check runs at schema-serve time." This means cache TTL determines how long a revoked permission remains visible in the UI. Five minutes is probably acceptable. It must be documented as an explicit security trade-off.

**Gap 2 — Feature flag cache stampede.** If a feature flag changes and 10,000 tenants' caches expire simultaneously, 10,000 concurrent requests to evaluate the flag hit the database. PER (Probabilistic Early Revalidation) was identified in prior reviews. Confirming: it must be implemented before v1.0 for any cache key class that is written uniformly across many tenants (which feature flag changes are).

### 9.5 Network Partition Behavior

When PostgreSQL is unavailable:
- All entity operations fail with 503
- Session validation fails (requires Redis) — wait, session tokens are in Redis not PostgreSQL. Session validation remains possible if Redis is up.
- Actually: session tokens are in Redis (lookup by token). User data is in PostgreSQL (for authorization). A session can be validated (token exists in Redis) but the user's roles cannot be fetched (PostgreSQL down). What does the framework return? 503? Cached role data?

**This is undefined behavior and must be defined before v1.0.**

Recommendation: If PostgreSQL is down, all write operations fail with 503. All read operations that require PostgreSQL fail with 503. Session validation succeeds only if roles are cached in the session token (embedding roles in the Redis session value at login time). This makes PostgreSQL failure non-fatal for read-only operations that do not require fresh data — an acceptable degraded mode.

The framework must document its partition behavior table:

| Component down | Effect |
|---|---|
| PostgreSQL primary | All writes fail (503). All reads fail (503). |
| PostgreSQL replica | Read-replica queries fall back to primary. |
| Redis | All auth fails (403). No caching. All reads/writes succeed if PostgreSQL is up. |
| Temporal | Workflow triggers fail. Entity CRUD succeeds. Outbox queues events for retry. |
| Outbox processor | Outbox accumulates. Delivery delayed. No data loss. |

---

## 10. Framework Minimalism

The following abstractions or mechanisms should be removed or collapsed before v1.0:

**Remove: `SmallText` field type.** `Data` (varchar indexed) and `SmallText` (varchar 1024, no index) differ only in storage detail. This is not a domain distinction. Module authors should not make storage decisions at the field type level. Replace both with `Text` qualified by `Searchable()` (adds GIN index) or `Indexed()` (adds B-tree index). The storage length is inferred from the index type.

**Remove: `DynamicLink` field type.** Polymorphic links without FK constraints create data integrity gaps. The standard Go pattern for polymorphism is interfaces. In entity terms, a polymorphic link target should be expressed as "any entity that provides capability X." This is a capability token, not a dynamic string. `DynamicLink` allows incorrect field names to be stored without validation. It cannot be referentially constrained. Remove it. Alternatives: typed union (entity declares multiple optional link fields, only one is set) or capability-based edge (edge to any entity providing a capability token).

**Remove: `Float` field type for ERP use cases.** The field type exists and the documentation says "scientific/percentages only, never money." But having `Float` as an available type in an ERP framework invites accidental misuse. Every developer who reaches for `Float` for a price or quantity is making a mistake the type system could prevent. Remove `Float` from the public API. If floating-point representation is genuinely needed (it almost never is in ERP), it can be added as a `RawFloat` type that requires an explicit acknowledgment comment in the EntityDefinition.

**Collapse: `before_validate` and `before_save` into one pre-save chain.** The current distinction is that `before_validate` runs before validation and `before_save` runs after validation but before persistence. Both are outside the transaction. In practice, hooks that run before validation are rare — most validation is expressed in `FieldDef` constraints. The distinction adds cognitive overhead for minimal benefit. Collapse into `before_save`, and move all validation into `FieldDef` declarative constraints. If a hook needs to run before other validation, it declares itself as a `ValidationHook` (a separate interface) that runs in the validation phase.

**Remove: Configuration vs. convention ambiguity in the middleware pipeline.** The middleware pipeline order is "not configurable at runtime — changing requires code change." This is correct. But if the middleware pipeline is expressed as a list of middleware functions that the bootstrap assembles, the order appears configurable even if it is documented as fixed. Express the pipeline as a fixed struct, not a slice:

```go
type MiddlewarePipeline struct {
    RequestID       Middleware  // always first
    StructuredLog   Middleware  // always second
    PanicRecover    Middleware  // always third
    CORS            Middleware  // always fourth
    TenantResolution Middleware // always fifth
    SessionValidation Middleware // always sixth
    RateLimiting    Middleware  // always seventh
    // No other fields. Cannot be reordered.
}
```

A struct with named fields communicates immutable ordering. A slice communicates configurable ordering.

---

## 11. v1.0 Freeze Audit

### 11.1 Freeze Forever (Kernel APIs)

These must never change after v1.0:

```
def.Record.ID() uuid.UUID
def.Record.TenantID() uuid.UUID
def.Record.EntityName() string
def.Record.SchemaVersion() int
def.Record.CreatedAt() time.Time
def.Record.UpdatedAt() time.Time

def.ViewerContext.ActorID() string
def.ViewerContext.TenantID() uuid.UUID
def.ViewerContext.Roles() []string
def.ViewerContext.IsSystem() bool

filter.Filter (wire format, once formally specified)
filter.Eq, filter.Neq, filter.Lt, filter.Gt, filter.In, filter.And, filter.Or, filter.Not

driver.EntityStore interface method signatures (all 10 methods)
driver.SessionStore interface method signatures
driver.WorkflowStarter interface

def.EntityDefinition.Name (field name, not field value)
def.EntityDefinition.Module (field name, not field value)
def.FieldDef.Name (field name)
def.FieldDef.Type (field name)

Workflow ID format: {tenant-uuid}.{entity-type}.{record-id}.{event}.{workflow-function}
Entity name convention: {module}_{noun}
Migration file format: {14-digit-timestamp}_{module}_{description}.up.sql
Event wire format: CloudEvents v1.0 envelope

HTTP API base path: /api/v1/entities/{entity-type}
HTTP API action path: /api/v1/entities/{entity-type}/{id}/{action}
HTTP error envelope: {"error": {"code": "...", "message": "..."}}
HTTP success envelope: {"data": {...}, "meta": {...}}

RLS mechanism: set_tenant_context() stored procedure
Tenant context variable: app.current_tenant_id
```

### 11.2 Make Internal Before v1.0 (Should Not Be Public)

These are currently reachable from outside the framework but should not be:

```
CompiledSchema internal fields (all fields) — expose only via methods
Registry.entities map — expose via Registry.Get(name) only
Hook chain slices — not addressable by external code
Casbin enforcer instance — never expose outside policy evaluation layer
Raw pgx pool — never expose outside EntityStore driver
SQL template strings — never expose outside driver
IR types (IREntity, IRField, etc.) — internal to compiler
Outbox table schema — internal to outbox driver
Session token Redis key format — internal to session driver
Compiler pass result structs — internal if not part of CompilerPass interface
```

### 11.3 Mark Experimental (Stable in v1.x, may change in v2)

```
CompilerPass interface
SDUI abstract schema types
AnalyticsDriver optional interface
StreamableEntityStore optional interface
ConflictStrategy field on EntityDefinition
SemanticDescription field on ActionDef
TenantPlacement on context
Module manifest Signature field (before signing infrastructure is fully operational)
```

---

## 12. Technical Debt Forecast

### 2030 (5 years)

**Most likely regret:** The two-stack architecture (`internal/` and `framework/`). Engineers who joined after v1.0 will maintain both stacks simultaneously, not understanding why both exist. Features will be added to one but not the other. The stacks will diverge. By 2030 this is already painful.

**Second most likely regret:** FieldDef struct size. If the extension map pattern was not enforced, FieldDef will have 30–40 fields by 2030. Every code review that touches FieldDef requires understanding all 40 fields.

**Third most likely regret:** The `DynamicLink` field type. Every use of DynamicLink in third-party modules will be a data integrity incident waiting to happen. By 2030, the support burden from DynamicLink-related bugs will be significant.

### 2035 (10 years)

**Most likely regret:** The SDUI schema design. If the abstract SDUI layer was designed hastily, it will be impossible to express new UI paradigms (conversational, spatial, AI-generated) without a breaking schema change. The 2035 equivalent of "mobile-first" will require SDUI schema evolution. If the schema has no extension mechanism, it is a hard block.

**Second most likely regret:** No formal GraphQL/gRPC transport. By 2035, REST alone will be limiting for enterprise integrations. If the abstract API descriptor was not built at v1.0, adding gRPC requires touching every entity definition and every driver.

**Third most likely regret:** Temporal as the only first-class workflow engine. By 2035, Temporal will have competitors. If the workflow driver interface was designed too thin (only what Temporal supports today), new workflow engines with better primitives cannot be fully supported.

### 2040 (15 years)

**Most likely regret:** Go module compatibility. Go will have had 3–4 major evolutions by 2040. The `def/` package's use of generics (as of Go 1.18-era), interface patterns, and type system features will look archaic. Code generation tooling will have changed. The public API will be frozen at v1.0 idioms.

**Second most likely regret:** PostgreSQL as the only storage model. By 2040, distributed SQL (Spanner, CockroachDB, YugabyteDB) will be standard for global deployments. If the EntityStore driver interface assumes single-primary PostgreSQL semantics (advisory locks, RLS via stored procedure, transaction isolation levels), distributed SQL drivers will need compatibility shims forever.

### 2045 (20 years)

**Most likely regret:** The compilation-unit model for plugins. By 2045, every major framework will have runtime-safe plugin isolation (WASM or equivalent). The decision to require all modules to be compiled into the binary will look like a 20-year mistake — even though it was the correct choice in 2025. The cost is: every module update requires a redeploy. At 500 modules, controlled by 500 different teams, redeploy coordination is an operational nightmare.

**Mitigation designed in v1.0:** The hook interfaces use serializable types, so WASM migration (from Go code to WASM functions) requires no interface change — only the execution mechanism changes. If this is designed correctly at v1.0, the 2045 migration to WASM plugins is a driver upgrade, not an ecosystem migration.

---

## 13. Final Architecture Verdict

### Scale Capacity Assessment

**10 developers: YES.**
The architecture is coherent, the five-layer structure is teachable, and EntityDefinition as the central primitive reduces the concepts a developer needs to understand before being productive.

**100 developers: YES, with governance.**
At 100 developers, the architectural constitution (Section 1.1) is the critical control. Without it, consistency degrades. With it, 100 developers can work independently on different modules without breaking each other. The module manifest and capability negotiation are also critical at this scale.

**1,000 developers: YES, with strong tooling.**
At 1,000 developers, linting, compliance tooling, and the compiler pass system are the critical controls. Every rule documented in the architectural constitution must have an automated check. Human review cannot scale to 1,000 developers. The compiler must catch violations.

**10,000 entities: YES.**
Compilation time with incremental compilation: acceptable. Route table with parametric dispatch: acceptable. Redis page schema cache: acceptable. No issues at this scale.

**100,000 entities: CONDITIONAL.**
Without incremental compilation: startup takes minutes. Unacceptable for rolling deployments and crash recovery. With incremental compilation: acceptable. This scale requires the incremental compiler from MUST-5/MUST-6.

**1 million tenants: CONDITIONAL.**
At 1 million tenants, `set_tenant_context()` is called 1M times/second at peak. This is within PostgreSQL's capability. But Redis session storage at 1M tenants × average 5 concurrent sessions = 5M session tokens. At 1 KB per session, this is 5 GB of Redis memory — manageable. The critical constraint is PgBouncer: connection pool must be large enough to handle peak concurrency without queuing. At 1M tenants, this requires careful capacity planning. The framework cannot solve this — it can only document the constraint.

**Fortune 500 deployments: CONDITIONAL.**
Fortune 500 requires: SOC 2 Type 2 (audit log integrity — supported), GDPR (field encryption, data residency — partially supported, needs TenantPlacement), SLA guarantees (SLO framework — supported), enterprise SSO (auth driver abstraction — supported), support contracts (documentation quality — TBD). With MUST items closed: YES.

**Government deployments: CONDITIONAL.**
Government requires: air-gapped operation (compilation-unit model supports this — no internet required at runtime), FIPS-compliant cryptography (key management driver — supported in principle, needs FIPS-compliant implementation), FedRAMP or equivalent (audit integrity — supported, external notarization needed). With regulated-industry MUST items closed: YES for most government use cases.

**Banking: CONDITIONAL.**
Banking requires: double-entry ledger integrity (SQL constraints on system entities — supported), regulatory audit trail (hash-chained audit log — supported), field-level encryption with key escrow (supported in principle — needs escrow in key management driver), real-time transaction processing (EntityRepository.Create() latency < 50ms — achievable with correct indexing). **Critical gap: optimistic locking.** Without optimistic locking, concurrent balance updates can produce incorrect results. This is a MUST before banking deployment. **Verdict: Requires MUST items including optimistic locking before banking.**

**Healthcare: CONDITIONAL.**
Healthcare requires: HIPAA (audit log + field-level encryption + access logging — supported), data residency (TenantPlacement — partially supported), HL7/FHIR integration (gRPC or REST driver — supported via abstract API descriptor after MUST-2). **Verdict: Requires MUST items before healthcare.**

**Military: NO (without significant additions).**
Military deployments require: multi-level security (MLS) with strict access control beyond Casbin's model, certified cryptographic modules (NSA Suite B / CNSA), deterministic behavior verification (formal proofs of security properties), supply chain verification beyond module signing, and security classification labels on data. None of these are in the current architecture. Military-grade security is a different architecture class. Awo is not designed for it. This is not a failure — it is a correct scope boundary. Document it as out-of-scope for v1.0.

---

## 14. Final Sign-Off

### APPROVED WITH REQUIRED CHANGES

The architecture is sound. The foundational decisions (metadata-first, structural tenancy, driver abstraction, compilation-unit modules, transactional outbox, five-layer dependency direction) are correct and will survive twenty years with proper governance.

However, four structural gaps exist that are not cosmetic — they are architectural contradictions that will require breaking changes if not resolved before v1.0:

---

**REQUIRED CHANGE 1 — Eliminate the two-stack architecture.**

`internal/` (legacy Wire DI) and `framework/` cannot both be "production" at v1.0. One must be the framework. The other must be deprecated with a removal date. If legacy `internal/` code serves current production traffic, the migration plan must be: (a) define the deadline, (b) run both stacks in parallel with feature parity, (c) cut over, (d) remove `internal/` before v1.0. The documentation cannot cover both stacks. Developers cannot learn both stacks. One stack. One framework.

**REQUIRED CHANGE 2 — Resolve CON-001 before v1.0 is frozen.**

Immutable compiled schema and per-tenant runtime custom fields are both correct requirements. They are architecturally contradictory without a `FieldResolver` interface that separates schema-time field definitions from request-time field overlays. Without this, custom fields either break the compiled schema invariant or cannot be queried uniformly with system fields. The resolution is straightforward (see CON-001) but must be in the type system before any documentation describes how custom fields work.

**REQUIRED CHANGE 3 — Optimistic locking must be in the EntityRepository interface before v1.0.**

Without `IfVersion(n)` query option on `Update()`, concurrent edits produce silent data loss. This is not observable in development and catastrophic in production. It is not a feature — it is a correctness requirement. Adding it after v1.0 changes the `EntityRepository` interface — a Kernel API. This change cannot happen post-freeze.

**REQUIRED CHANGE 4 — INV-005 (policy evaluation semantics) must be formally specified.**

The policy model must choose: first-match wins, all-must-pass (AND), or any-may-grant (OR). This choice affects every security decision in every deployment. It cannot be changed after third-party modules write policies against it. The specification must be a formal written ruling, not derivable from the implementation. Once specified, it is frozen.

---

**These four changes are the minimum before documentation begins.**

All other items from prior reviews (abstract SDUI schema, abstract API descriptor, incremental compiler, module signing, etc.) remain required for a complete v1.0. But these four are the only ones that affect the type system or the public API contract in ways that cannot be retrofitted after documentation begins.

Close these four. Then documentation may begin.

*Signed: Chief Software Architect*
*Date: 2025-07-06*
*Status: APPROVED WITH REQUIRED CHANGES — 4 architectural changes before documentation*
