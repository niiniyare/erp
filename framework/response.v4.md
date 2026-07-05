# Awo Framework — Language Specification Committee Review
## Final Pre-Freeze Sign-Off | v1.0 Permanent Architecture | 2025-07-06

**Classification:** Language Specification Committee Review (LSC-FINAL-001)
**Precondition:** All v1–v3 review recommendations implemented.
**Scope:** Permanent design mistakes only. Irreversible decisions only.
**Standard:** PostgreSQL wire protocol stability (1996–present). Git object model stability (2005–present). Linux ABI stability (1994–present).

No improvement suggestions. No optional features. No repeated findings.
Every finding here, if ignored, forces awo/v2.

---

## 1. Kernel Contract Completeness

### 1.1 `HookSet` is a Struct. This Is a Permanent Extensibility Trap.

`HookSet` is declared in `def/` as a struct with fixed fields:

```go
type HookSet struct {
    BeforeCreate  []BeforeCreateHook
    AfterCreate   []AfterCreateHook
    BeforeUpdate  []BeforeUpdateHook
    AfterUpdate   []AfterUpdateHook
    BeforeDelete  []BeforeDeleteHook
    AfterDelete   []AfterDeleteHook
    AfterSave     []AfterSaveHook
    // ...
}
```

`def/` is frozen at v1.0. Adding `OnRead []OnReadHook` to HookSet requires modifying a frozen package. This forces one of:

(a) Modify `def/` → major version bump → ecosystem fragmentation
(b) Secondary hook registration API outside `def/` → two competing hook systems
(c) A `HookSet.Extensions map[string][]any` escape hatch → loses all type safety

None of these are acceptable after v1.0.

**Root cause:** Lifecycle stages are modeled as struct fields instead of as a keyed registry. A struct encodes "these are all the lifecycle stages that will ever exist." This is false.

**Required fix before freeze:** Replace `HookSet` struct with a lifecycle-keyed registration model:

```go
type LifecycleStage string

const (
    StageBeforeCreate  LifecycleStage = "before_create"
    StageAfterCreate   LifecycleStage = "after_create"
    StageBeforeUpdate  LifecycleStage = "before_update"
    StageAfterUpdate   LifecycleStage = "after_update"
    StageBeforeDelete  LifecycleStage = "before_delete"
    StageAfterDelete   LifecycleStage = "after_delete"
    StageAfterSave     LifecycleStage = "after_save"
    StageOnRead        LifecycleStage = "on_read"
    StageAfterCommit   LifecycleStage = "after_commit"
)

type HookRegistration struct {
    Stage    LifecycleStage
    Hook     Hook           // single interface, type-asserted per stage
    Priority int
    Module   string         // owning module, for conflict detection
}
```

`EntityDefinition.Hooks` becomes `[]HookRegistration`. New lifecycle stages are new `LifecycleStage` constants — additive, no struct change. The kernel `Hook` interface is a marker. Per-stage interfaces are type-asserted at compile time by the compiler, not by the kernel.

This is the only model that allows `def/` to remain frozen while the lifecycle model evolves.

### 1.2 `FieldType` is a Closed Enum. This Prevents Ecosystem Field Extensions.

`FieldType` is defined in `def/` as:

```go
type FieldType string  // or int constant

const (
    FieldTypeData     FieldType = "data"
    FieldTypeCurrency FieldType = "currency"
    // ...
)
```

A third-party healthcare module needs `FieldTypeHL7Code`. A GIS module needs `FieldTypeGeometry`. A banking module needs `FieldTypeIBAN`.

These modules cannot add field types without modifying `def/`. If `def/` is frozen, custom field types are impossible. Every third-party module that needs a domain-specific field type is blocked.

**Required fix before freeze:** `FieldType` must be a string type (not an integer constant) with a validation-free kernel definition. The kernel declares built-in field types as constants. Third-party modules declare their own constants in their own packages. The compiler validates that every `FieldType` used in the binary has a registered driver-level handler. Unknown field types are compile errors, not runtime panics.

```go
// def/ — kernel only declares the type, not all values
type FieldType string

// Kernel constants (frozen)
const (
    FieldTypeData     FieldType = "data"
    FieldTypeCurrency FieldType = "currency"
    FieldTypeInt      FieldType = "int"
    // ...
)

// Third-party can add:
// const FieldTypeGeometry FieldType = "com.example.gis.geometry"
// Registered via FieldTypeDriver in their init()
```

The compiler phase for semantic analysis checks that every `FieldType` used in the entity graph has a corresponding registered type handler. This is a compile-time error, not a runtime panic.

### 1.3 `WorkflowTrigger.On` is a Closed Event Type. Same Problem.

`entity.EventOnCreate`, `entity.EventOnSubmit`, `entity.EventOnDelete` are constants in `def/`. A third-party approval module needs `EventOnApprove`. A custom action module needs `EventOnCustomAction`.

Adding new trigger events requires modifying `def/`. Frozen `def/` prevents this.

**Required fix before freeze:** Same pattern as FieldType. `TriggerEvent` is a string type. Kernel declares standard events. Third-party modules declare domain events. The compiler verifies that every declared trigger event maps to a lifecycle stage that the entity lifecycle engine can emit.

### 1.4 `nil` ViewerContext is Undefined Behavior

System operations (background jobs, cron workers, migration runners, outbox processors) execute without a human viewer. What is their ViewerContext?

If `nil` is passed as ViewerContext, every policy that calls `viewer.ActorID()` panics. If a "system viewer" sentinel is passed, policies must explicitly handle it. If the compiler guarantees that system operations bypass policies entirely, the architectural contract must say so.

Currently undefined. Will produce inconsistent implementations across every module that needs to run background operations.

**Required ruling before freeze:** Declare exactly one of these semantics (not both, not configurable):

**Option A — System operations have a typed SystemViewer:**
```go
type SystemViewer struct{}
func (SystemViewer) ActorID() string    { return "system" }
func (SystemViewer) TenantID() uuid.UUID { return uuid.Nil } // special
func (SystemViewer) Roles() []string    { return []string{"role:system"} }
func (SystemViewer) IsSystem() bool     { return true }
```
Policies check `viewer.IsSystem()` and short-circuit. Casbin enforcer has a `role:system` that bypasses tenant-scoped rules.

**Option B — System operations receive a tenant-scoped system viewer:**
Background jobs always run in the context of a specific tenant. The system viewer carries the tenant ID. There is no tenant-nil system operation.

One of these must be the kernel contract. Both cannot be valid simultaneously. Modules that choose the wrong assumption will silently fail under the other's deployment model.

### 1.5 `CreateInput` / `UpdateInput` are Structurally Undefined

The EntityRepository interface declares:

```go
Create(ctx context.Context, input CreateInput) (T, error)
Update(ctx context.Context, id uuid.UUID, input UpdateInput) (T, error)
```

What is `CreateInput`? What is `UpdateInput`? If they are `map[string]any`, they are untyped bags. If they are specific structs, they cannot accommodate custom fields. If they are interfaces, what methods do they expose?

This is not a naming question. This is a kernel contract question. The `CreateInput` type determines:
- How field validation is triggered
- How custom fields are distinguished from system fields
- How optional vs required fields are expressed
- Whether partial updates (PATCH semantics) are possible
- Whether input validation happens before or inside the repository call

If `CreateInput` is `map[string]any`, it survives 20 years but loses compile-time safety. If it is a typed struct, it cannot express custom fields. If it is an interface, the interface must be frozen.

**Required decision before freeze:** Specify `CreateInput` and `UpdateInput` precisely. Recommend:

```go
type FieldValue struct {
    Field  string
    Value  any
    Source FieldSource  // SystemField | CustomField — drives storage routing
}

type CreateInput struct {
    Fields  []FieldValue
    IDHint  *uuid.UUID   // client-provided ID for idempotency (optional)
    Options []WriteOption
}

type UpdateInput struct {
    Fields    []FieldValue
    IfVersion *int         // optimistic locking (nil = unconditional)
    Options   []WriteOption
}
```

The `FieldSource` tag is how the driver knows to write to a SQL column vs. the JSONB custom_fields column. Without it, the driver must infer this from the compiled schema — which couples the driver to schema knowledge it should not need.

### 1.6 `AggregateSpec` and `AggregateResult` are Undefined Kernel Types

`EntityRepository.Aggregate()` exists. Its parameters are `AggregateSpec` and `AggregateResult`. If these types are in `def/` and are frozen, the aggregate operations available to the framework are permanently limited to what was designed in 2025.

In 2035, time-series aggregation, percentile calculations, and window functions will be standard ERP requirements. If `AggregateSpec` cannot express them, every module needing advanced analytics works around the interface.

**Required before freeze:** `AggregateSpec` must be extensible. Recommend: a base set of operations (sum, count, avg, min, max, group_by) in kernel, with a `driver.AggregateExtension` optional interface for advanced operations. Drivers declare which extensions they support.

### 1.7 `PageInfo` in Query Returns Leaks Pagination Implementation

```go
Query(ctx context.Context, f Filter, opts ...QueryOption) ([]T, PageInfo, error)
```

`PageInfo` carries the pagination result. If `PageInfo` contains `TotalCount int` — that count requires a `COUNT(*)` query that is expensive at scale. Some callers do not need total count (infinite scroll). Some do (paginated tables). If `TotalCount` is always populated, all queries pay the count cost. If `TotalCount` is sometimes zero, callers cannot distinguish "zero records" from "count not computed."

**Required before freeze:**

```go
type PageInfo struct {
    HasNextPage  bool
    NextCursor   string    // opaque, for keyset pagination
    TotalCount   *int64    // nil = not computed (caller did not request it)
}
```

`TotalCount` is only populated when the caller passes `QueryOption.WithTotalCount()`. Default is nil. This is a kernel contract — if changed after v1.0, every caller that depends on `TotalCount` being populated breaks.

---

## 2. Extension Model

### 2.1 Cross-Module Hook Registration Has No Ownership Semantics

A CRM module can call `definition.Register()` with a hook on `finance_invoice.BeforeCreate`. Nothing in the architecture prevents this. The Finance module, which owns `finance_invoice`, has no mechanism to:

(a) Declare that its entity does not accept external hooks
(b) Declare that external hooks require explicit permission
(c) Know which modules registered hooks on its entities

At 500 modules, this produces untraceable behavioral interference. A bug in invoice creation behavior leads an engineer to read the Finance module's hooks and find them all correct — because the bug is in the CRM module's hook registered on a Finance entity.

**Required before freeze:** `HookRegistration.Module` (the registering module's name) must be recorded. Entity definitions must be able to declare hook policies:

```go
type HookPolicy int
const (
    HookPolicyOpen     HookPolicy = iota // any module may register (default)
    HookPolicyAllowList                   // only listed modules may register
    HookPolicySealed                      // only the owning module may register
)
```

The compiler enforces hook policies at compile time. An `AllowList` entity that receives a hook from an unlisted module is a compile error. A `Sealed` entity that receives any external hook is a compile error.

This is not a restriction on the ecosystem — it is an explicit contract. Open entities explicitly allow external hooks. Sealed financial entities explicitly prevent them. The contract is visible in source.

### 2.2 Duplicate Route Ownership is Unresolved

Two modules registering entities with the same name produce a compile error (INV-011). But what about two modules registering *different* entities that produce the *same route pattern*?

Example: Module A defines entity `crm_contact`. Module B defines entity `contact` (forgetting the module prefix). Both produce route `/api/v1/entities/contact/*` if the route generator strips the module prefix for URL generation.

The route table compiler must verify uniqueness of all generated routes. If a collision exists, it must report: which two entities, which route pattern, which modules. This is a compiler enforcement question — but the rule must be in the kernel contract before modules exist to violate it.

**Required before freeze:** Formally define the URL generation rule for entity routes. The entity name is the URL segment. Entity names include module prefix (`crm_contact` → `/api/v1/entities/crm_contact/`). No stripping. No aliasing without explicit declaration. Route uniqueness is a compile-time invariant.

### 2.3 UI Renderer Extension Has No Abstract Schema Contract

If the abstract SDUI schema is the extension point for UI renderers, the schema itself must be a versioned, formally specified type in a public package. A third-party renderer (React Native, terminal UI, PDF generator) implements `driver.UIRenderer`:

```go
type UIRenderer interface {
    RenderPage(ctx context.Context, page def.UIPage, viewer def.ViewerContext) ([]byte, string, error)
    // Returns: serialized page, content-type, error
}
```

`def.UIPage` is the abstract SDUI type. If this type is in `def/` and is frozen, new UI concepts (streaming updates, real-time collaboration indicators, AI-generated layouts) cannot be added.

The SDUI abstract schema has the same extensibility problem as `HookSet`. If `def.UIPage` is a struct with fixed fields, adding a new UI concept requires modifying `def/`.

**Required before freeze:** `def.UIPage` must have a typed extension map:

```go
type UIPage struct {
    Title    string
    Layout   UILayout
    Sections []UISection
    // Renderer-specific extensions — never iterate in kernel
    Extensions map[string]any
}
```

The kernel defines the minimum set. Renderers define their own extension keys (namespaced: `"com.example.renderer.sidebar"`). The kernel never reads extension keys — only the specific renderer that wrote them reads them.

---

## 3. Versioning Strategy

### 3.1 The Filter DSL Has Two Versioning Problems

**Problem A — No version in the wire format.**

A filter serialized by a v1.0 client and sent to a v1.5 server must be parseable. A filter serialized by a v1.5 client with a new operator (`VectorSimilarity`) and sent to a v1.0 server must fail explicitly, not silently produce wrong results.

The filter wire format must embed a version:

```json
{
  "v": 1,
  "op": "and",
  "args": [...]
}
```

A v1.0 server receiving `"v": 2` returns `HTTP 422` with `error.code = "filter_version_not_supported"`. A v1.5 server receiving `"v": 1` parses it with backward-compatible rules. The version field is mandatory. A filter without it is rejected.

**Problem B — No versioning of operator semantics.**

`filter.Contains("name", "acme")` — does this mean SQL `LIKE '%acme%'` or full-text search or GIN trigram match? The behavior is driver-dependent. If the PostgreSQL driver changes from `LIKE` to trigram in v1.2 (for performance), the semantic behavior changes without a version bump.

**Required before freeze:** Every filter operator must declare its semantic contract in the specification document (not the code). The specification is permanent. Drivers that deviate from the specification are non-conformant. This is the only way to guarantee filter portability across drivers and versions.

### 3.2 Entity SchemaVersion Semantics Against Multi-Version Records

Records carry `SchemaVersion`. A record written at schema version 2 is read after the entity evolves to schema version 4. The record is missing fields added in versions 3 and 4.

`EntityRepository.Get()` returns this record. What does the caller receive?

- Fields added in v3 and v4 are absent from the record? Or zero-valued? Or explicitly marked as missing?
- Queries filtering on a v3+ field against v2 records — do v2 records match or not match?
- Hooks declared in v3 that depend on fields added in v3 — do they run correctly against v2 records?

This is not an implementation question. It is a kernel semantic question. Every module author that reads multi-version records needs to know the answer. If the answer is undefined, every module author guesses differently. The guesses compound over 20 years.

**Required ruling before freeze (choose one permanently):**

**Semantic A — Missing fields are nil/zero:** Records at old schema versions return nil for fields that didn't exist yet. Callers must handle nil. Hooks must guard against nil field access.

**Semantic B — Schema migration is mandatory before access:** Records at old schema versions cannot be read until migrated to current schema. The repository returns `ErrSchemaMismatch` for records at wrong version. Lazy migration on read: the repository migrates the record in-place before returning it.

**Semantic C — Explicit version gating:** Queries must declare which schema version they expect. Reading a v2 record under a v4 query context returns the record with v2 semantics only.

Semantic A is the most practical for ERP. Semantic B is the safest for regulated industries. Semantic C is the most explicit but most complex. The kernel must declare which is the default and whether the others are configurable.

### 3.3 Migration Ordering Across Independent Modules

Module A's migration `20250601000000_finance_create_invoice.up.sql` and Module B's migration `20250601000001_crm_create_contact.up.sql` are ordered by timestamp. But these modules are developed independently. Module B's author does not know about Module A's timestamp.

If both modules are developed simultaneously, their timestamps may collide or interleave arbitrarily. The migration runner applies them in timestamp order. The order between independent modules is non-deterministic (both timestamps are valid, neither is definitively "first" in a way that matters).

The real problem arises when Module B's migration depends on Module A's migration:

```sql
-- Module B migration: add FK to finance_customer (owned by Module A)
ALTER TABLE crm_contact ADD COLUMN customer_id uuid REFERENCES finance_customer(id);
```

If Module A's migration has a higher timestamp, it runs after Module B's migration, and the FK fails.

**Required before freeze:** Cross-module migration dependencies must be explicit. Module B's manifest must declare:

```go
MigrationRequires: []string{"finance >= 20250601000000"}
```

The migration runner enforces this: Module B's migrations do not run until all declared migration prerequisites are applied. This is a migration dependency graph, not just a timestamp sort.

### 3.4 Driver Interface Evolution

Driver interfaces are in `driver/`, which is a public package. Once `driver.EntityStore` has 10 methods at v1.0, adding method 11 in v1.1 breaks every driver implementation.

The Go interface compatibility problem for evolving driver interfaces:

**Option A — New interfaces for new capabilities:**
```go
// v1.0
type EntityStore interface { 10 methods }

// v1.1 — additive, does not break v1.0 implementations
type StreamableEntityStore interface {
    EntityStore
    QueryStream(...) (<-chan Record, error)
}
```

Runtime: `if ss, ok := store.(StreamableEntityStore); ok { ... }`

This is the only backward-compatible approach in Go. It is already recommended in prior reviews. **Confirm it as the permanent pattern.** Document it as immutable: no new methods ever added to existing driver interfaces. New capabilities always go into new optional interfaces.

**Option B — Wrong:** Adding methods to existing interfaces. This breaks every driver. Forbidden after v1.0.

The pattern must be documented as a kernel law, not a recommendation.

---

## 4. Module Ecosystem

### 4.1 The Entity Ownership Problem

`finance_invoice` is owned by the Finance module. The CRM module cannot modify it — but the CRM module *can*:

- Register hooks on it (addressed in Section 2.1)
- Define actions on it (currently unrestricted)
- Define policies on it (currently unrestricted)
- Reference it in its own entity edges (acceptable — read-only reference)

Registering actions on another module's entity is a sovereignty problem. If the CRM module adds a `POST /api/v1/entities/finance_invoice/{id}/crm_tag` action, it has modified the public API surface of the Finance module without the Finance module's knowledge.

In year 5, the Finance module changes its invoice model and the CRM action silently breaks. The Finance module cannot know this — it did not declare the action. The CRM module cannot know this — it does not know about Finance's internal changes.

**Required before freeze:** Action registration must carry ownership. An action on entity E is owned by the module that owns E. Cross-module action registration must be explicit: entity E's manifest declares `AllowExternalActions: true`. Otherwise, only the owning module may register actions.

This is the same pattern as HookPolicy but for actions.

### 4.2 Capability Token Namespace Collision

Two vendors independently define capability token `"reporting.export"`. Their modules both declare `Provides: ["reporting.export"]`. The resolver detects a conflict. Neither module is wrong — both provide an export capability. But the resolver cannot resolve.

The fix is reverse-DNS namespacing: `"io.vendor-a.reporting.export"` and `"io.vendor-b.reporting.export"` never collide. But consumers that want "any export capability" now must know specific vendor names.

**Required before freeze:** Define two classes of capability tokens:

1. **Framework-defined capability tokens** (short names, no namespace): `"platform.tenant"`, `"platform.iam"`, `"finance.ledger"`. These are defined by the framework specification. Vendors may not use unnamespaced tokens.

2. **Vendor-defined capability tokens** (reverse-DNS namespace): `"io.vendor.crm.contact_management"`. Consumers declare dependencies on specific vendor capabilities.

Framework-defined tokens are the stable ecosystem contracts. Vendor tokens are bilateral agreements. The specification must define the complete set of framework-defined tokens at v1.0. Adding a new framework-defined token is a minor version change.

### 4.3 No Mechanism for Capability Compatibility Promises

Module A provides `"finance.ledger"` at capability version `1.2`. Module B requires `"finance.ledger" >= 1.0`. But what does capability version `1.2` actually mean?

Without a formal definition of what constitutes a breaking change in a capability, the version number is meaningless. A capability version increment could mean "added an optional field to a hook input" (non-breaking) or "changed a hook signature" (breaking). Consumers cannot know which.

**Required before freeze:** Framework-defined capability tokens must have published compatibility specifications, not just version numbers. The specification declares:
- What the capability provides (which entity types, which actions, which events)
- What the capability requires from the consumer (which hook interfaces, which viewer methods)
- What constitutes a breaking change

Without this, capability versioning is theater.

---

## 5. Metadata Evolution Across Decades

### 5.1 Field Retirement Has No Safe Path

At v1.0, `finance_invoice` has field `fax_number`. By 2030, fax machines are gone. The field must be retired.

The field is in the `EntityDefinition`. It has a SQL column in the schema. Records have data in it. Reports reference it. The outbox may have events that carry it.

**No current retirement mechanism exists.** Options:

(a) Delete the field from `EntityDefinition` → compilation error for existing records that reference it → must add a migration to drop the column → data loss
(b) Mark field `Deprecated: true` → field still exists, still queryable, but hidden from UI and documentation → data preserved
(c) Mark field `Retired: true` after `Deprecated` → field is still in schema but write operations are rejected → preserves historical data, prevents new data
(d) Archive field to JSONB after `Retired` → field migrated from typed column to custom_fields JSONB → preserves query semantics but changes storage

**Required before freeze:** A formal field lifecycle must be declared in the kernel:

```
Active → Deprecated (reads OK, writes OK, hidden from UI)
       → Retired (reads OK, writes rejected, hidden from all)
       → Archived (reads via custom_fields only, typed column dropped)
```

This lifecycle must be expressible in `FieldDef`:

```go
type FieldDef struct {
    // ...
    Lifecycle FieldLifecycle  // Active | Deprecated | Retired | Archived
    RetiredAt *time.Time      // when the field was retired
    ArchivedAt *time.Time     // when the column was dropped
}
```

The compiler generates different behavior per lifecycle stage. The migration runner refuses to drop a non-Archived field's column. The query builder rejects writes to Retired fields with `ErrFieldRetired`.

Without this, field retirement is a manual, error-prone process that each team handles differently.

### 5.2 Entity Retirement Has No Safe Path

Same problem at the entity level. An entity `legacy_sync_log` served a purpose in 2025. By 2035 its workflow is replaced. The entity has 100M records. It cannot be dropped. But it must stop accepting new writes. Existing code references it.

**Required before freeze:** Entity lifecycle parallel to field lifecycle:

```
Active → Deprecated → Retired → Archived
```

`Retired` entity: routes return 410 Gone. Hooks are not invoked. Writes rejected. Reads allowed. Migration still present. The compiled schema includes the entity with `Retired` status — the compiler generates 410 responses for its routes rather than removing it.

### 5.3 Compiled Metadata Cache Invalidation Is Not Formally Specified

Redis caches compiled page schemas. If the compiled schema changes (entity redeploy, schema version bump), cached schemas must be invalidated.

The cache key is `page:{entity}:{version}:{tenant}`. The `{version}` is the entity schema version. If the schema version is an integer that increments on every EntityDefinition change, cache invalidation is automatic on deploy.

But: who increments the schema version? If it's the developer (manually), they will forget. If it's automatic (based on a hash of the EntityDefinition), what exactly is hashed?

If the hash includes field labels (human-readable strings), changing a label from "Invoice Number" to "Invoice No." invalidates the schema cache and forces all tenants to re-fetch page schemas. This is disproportionate.

If the hash excludes labels, it only covers structural changes. This is correct.

**Required before freeze:** Formally define what constitutes a schema version increment:

- **Increments version:** field added, field removed, field type changed, edge added, edge removed, action added, action removed
- **Does NOT increment version:** label changed, help text changed, UI hints changed, documentation changed

This taxonomy must be in the specification and enforced by the compiler (which computes the hash using only structural elements).

---

## 6. Security Architecture

### 6.1 The Compiled Schema Endpoint Is a Security Intelligence Surface

`GET /_framework/schema/entity/{name}` returns the compiled definition: fields, edges, policies, hooks, workflow triggers. This reveals:

- Which fields are encrypted (`Encrypted: true`) — tells an attacker which fields are high-value
- Which roles have permission to which operations — full Casbin policy dump
- Which hooks exist — reveals business logic structure
- Which workflow triggers exist — reveals what server-side automation can be triggered

For a regulated deployment, this endpoint is a compliance liability. An internal attacker who discovers this endpoint has a complete map of the system's security model.

**Required before freeze:** The schema diagnostic endpoint must:
1. Require platform-admin authentication (already stated in prior reviews)
2. Filter out security-sensitive fields from the response by default: no encryption flags, no full policy expansion, no hook implementation details
3. Have an explicit `?full=true` parameter that returns the complete schema, additionally gated by a separate `diagnostic:full_schema` permission

The public schema (used by SDUI renderers, API documentation generators) must never include security annotations.

### 6.2 Hook Input Carries Sensitive Data to Third-Party Code

A `BeforeCreate` hook for `finance_invoice` receives the full `MutableRecord`. The record may contain sensitive fields (`Sensitive: true`): bank account numbers, tax IDs, personally identifiable information.

If a third-party hook is registered on this entity (via `HookPolicy.Open`), that third-party code receives sensitive data without any data-masking mechanism. The framework cannot mask sensitive fields before passing them to hooks, because the hook may legitimately need them (e.g., a validation hook for an IBAN field).

But a third-party analytics hook that aggregates invoice totals has no legitimate need for the bank account number in the same record.

**Required before freeze:** Hook registration must declare which fields the hook reads:

```go
type HookRegistration struct {
    Stage    LifecycleStage
    Hook     Hook
    Priority int
    Module   string
    Reads    []string  // field names this hook reads; "*" = all
    Writes   []string  // field names this hook may write; "*" = all (owning module only)
}
```

The runtime masks fields declared `Sensitive: true` from hooks that do not list them in `Reads`. A hook that declares `Reads: []string{"total_kes"}` receives a `MutableRecord` where `bank_account` is nil, even if the underlying record has data.

This is capability-based data access for hooks — a hook only sees what it declares it needs.

### 6.3 Registry Poisoning Has No Architectural Defense

`init()` functions register entity definitions. In Go, `init()` functions run automatically when a package is imported. A malicious dependency that is transitively imported can call `definition.Register()` with a crafted entity definition that:

- Installs a hook on `platform_user.BeforeCreate` that exfiltrates the new user's credentials
- Registers a policy that grants broad access to a backdoor role
- Registers a workflow trigger that starts an attacker-controlled workflow

The framework has no mechanism to verify that `definition.Register()` is called from trusted code.

**Architectural reality:** In Go's compilation model, this attack requires code execution at build time or supply-chain compromise of a dependency. This is an inherent limitation of the `init()` registration pattern.

**Required before freeze (architecture, not implementation):** The framework must document this trust model explicitly:

> "Awo's security model trusts all code compiled into the binary equally. All modules are trusted at the same level as the framework itself. Supply chain security is the deploying organization's responsibility. Module signing verifies module identity but does not sandbox module code from framework access. WASM-based hook isolation (v2.0) is the long-term mitigation."

Without this documentation, deployers assume that module signing provides runtime isolation. It does not. Module signing proves origin. It does not constrain behavior. The difference must be explicit in the security architecture document.

### 6.4 The Outbox Table Is a Workflow Trigger Attack Surface

The outbox table contains entries of the form:
```
(tenant_id, entity_name, record_id, event_type, payload, status)
```

The outbox processor reads `status = 'pending'` entries and starts workflows. If an attacker gains write access to the outbox table (SQL injection, compromised application role, compromised background service), they can insert arbitrary outbox entries and trigger arbitrary workflows.

**Required before freeze:** The outbox table must be append-only for the application role. The application role has `INSERT` on the outbox table but no `UPDATE` or `DELETE`. The outbox processor runs under a separate role that has `SELECT` and `UPDATE` (to mark entries as processed). No application code and no hook can mark outbox entries as processed — only the processor can.

This is a database architecture constraint, not an application constraint. It must be in the migration template for the outbox table.

---

## 7. Distributed Architecture

### 7.1 Compiled Schema Divergence Across Instances Is Possible

All instances of the same binary version compile the same schema. But `init()` functions can behave conditionally:

```go
func init() {
    if os.Getenv("ENABLE_LEGACY_BILLING") == "true" {
        definition.Register(&LegacyBillingEntity)
    }
}
```

Two instances started with different environment variables compile different schemas. They serve different routes. They accept different entity types. They may even have different RLS policies.

A request for `/api/v1/entities/legacy_billing/123` routed to an instance where `ENABLE_LEGACY_BILLING=false` returns 404. Routed to an instance where `ENABLE_LEGACY_BILLING=true` returns the record. The same request produces different results depending on which instance handles it. This is a correctness violation.

**Required before freeze:** The framework must detect schema divergence across instances and refuse to serve traffic if detected. Mechanism: the compiled schema content hash (from prior reviews) is registered in a cluster coordination key (Redis: `cluster:schema_hash`). On startup, each instance writes its hash. If two different hashes exist simultaneously, the instance with the newer startup time logs a critical error and stops accepting traffic (returns 503 on all endpoints).

Additionally: entity definitions must not be conditionally registered based on environment variables. The compiler must enforce this by analyzing the `init()` call graph — if any registration is inside a conditional, it is a compile error. Conditional module activation must use the Feature Flag system, not conditional registration.

### 7.2 Background Worker and Request Worker Share the Same Compiled Schema

In a single binary, background workers (outbox processor, cron scheduler) and request handlers run in the same process, sharing the same compiled schema and the same driver instances.

If the outbox processor acquires a PostgreSQL connection and starts a long transaction while the request handler also needs a connection, connection pool exhaustion can occur. The outbox processor can starve request handlers.

This is not a tuning problem. It is an architectural isolation problem. The outbox processor and the request handlers compete for the same resource pool.

**Required before freeze:** The framework must support separate connection pool quotas for background workers vs. request handlers. Not separate pools — separate quotas within the same pool. The EntityStore driver must accept a `ConnectionClass` annotation in the context:

```go
ctx = driver.WithConnectionClass(ctx, driver.ConnClassBackground)
```

The pool reserves `N%` of connections for `ConnClassBackground` callers and refuses to exceed that percentage, ensuring request handlers always have connections available.

### 7.3 Cron Job Semantics Under Horizontal Scaling

Leader election (advisory lock) ensures one instance runs the scheduler. But when the leader crashes, the election must re-occur before the next scheduled job. If the election takes 30 seconds and a job was scheduled to run every 30 seconds, the job is skipped.

This is acceptable for non-critical jobs. It is not acceptable for compliance-critical jobs (end-of-day batch, regulatory reporting).

**Required before freeze:** Cron jobs must declare `CriticalityClass`:

```go
type CronDef struct {
    Schedule  string         // cron expression
    Handler   CronHandler
    Critical  bool           // if true: skipped execution produces an alert, not silent drop
}
```

A `Critical: true` job that is skipped because of leader election delay must produce an observable event: an error log at CRITICAL level, a Prometheus counter increment, and an outbox entry that can trigger an alert workflow.

The framework cannot guarantee exactly-once execution of cron jobs in all failure scenarios. It must guarantee that missed executions of critical jobs are observable.

---

## 8. Long-Term Maintainability

### 8.1 The `init()` Registration Pattern Hides Causality

An engineer in 2038 reads:

```go
func init() {
    definition.Register(&InvoiceDefinition)
}
```

Questions they cannot answer from reading this code:

- When does this `init()` run relative to other modules' `init()`s?
- What happens if `definition.Register()` fails? (It panics — invisible from the call site)
- How can I see all registered entities without running the server?
- What if I want to register conditionally based on a runtime configuration?

The `init()` pattern is idiomatic Go for side-effecting registration. It is also deeply magical to anyone unfamiliar with Go's package initialization model.

**Architectural mitigation (not a change to the pattern, but a required contract):** The framework must provide an offline inspection tool that, given a Go binary, outputs the compiled schema without starting the server. This tool uses the `/_framework/schema` endpoint against a `--dry-run` binary mode that compiles the schema and exits. Engineers in 2038 can inspect the schema without understanding `init()` semantics.

Additionally: `definition.Register()` must never panic silently. If it panics (duplicate entity name, invalid definition), the panic message must include: the entity name, the calling module, and the prior registrant. Stack traces alone are insufficient.

### 8.2 The Five-Layer Architecture Requires a Violation Detector

The five-layer rule (UI → API → Domain → Workflow → Store, strict top-down) is currently enforced by convention and code review. In 2038, with 1000 developers and 500 modules, convention does not scale.

**Required before freeze:** A Go analysis pass (implemented using `golang.org/x/tools/go/analysis`) that detects imports from lower layers to higher layers. This pass runs in CI. A domain package importing an API package is a build failure, not a code review comment.

The package hierarchy that the analyzer enforces must be documented as a permanent architectural constraint, not a style guide.

### 8.3 The Workflow Task Queue Name Is a Hidden Configuration Contract

```go
WorkflowTriggers: []entity.WorkflowTrigger{
    {
        TaskQueue: "finance.invoice.submit",
    },
}
```

`"finance.invoice.submit"` is a string constant that must match the task queue name configured in the Temporal worker. If a deployment changes the task queue name (e.g., for environment isolation: `"staging.finance.invoice.submit"`), the EntityDefinition must be modified. EntityDefinition is code, not configuration.

This couples deployment topology to entity definitions. In 2038, when a new environment is added, every EntityDefinition that hardcodes a task queue name must be modified.

**Required before freeze:** Task queue names must be resolvable at runtime, not hardcoded at compile time:

```go
type WorkflowTrigger struct {
    TaskQueue TaskQueueResolver  // interface, not string
}

// Simplest implementation:
type StaticTaskQueue string
func (q StaticTaskQueue) Resolve(ctx context.Context, tenantID uuid.UUID) (string, error) {
    return string(q), nil
}

// Environment-aware implementation:
type EnvPrefixedTaskQueue struct {
    Base string
}
func (q EnvPrefixedTaskQueue) Resolve(ctx context.Context, tenantID uuid.UUID) (string, error) {
    return os.Getenv("TASK_QUEUE_PREFIX") + "." + q.Base, nil
}
```

Most modules use `StaticTaskQueue("finance.invoice.submit")`. Deployments that need environment isolation inject an `EnvPrefixedTaskQueue`. No EntityDefinition modification required.

---

## 9. Formal Architecture Invariants (Permanent Laws)

These are the permanent architecture laws of Awo Framework v1.0. Any violation is an architecture defect, regardless of whether tests pass.

```
LAW-001  The kernel (def/) imports nothing from the framework.
         def/ depends only on the Go standard library.

LAW-002  The runtime consumes only CompiledSchema. The runtime never reads
         EntityDefinition directly after compilation is complete.

LAW-003  The registry accepts registrations only during the initialization
         phase. After Compile() returns, all registration methods return
         an error or panic deterministically. No exception.

LAW-004  CompiledSchema is immutable after Compile() returns.
         No method on CompiledSchema mutates state.
         CompiledSchema is safe for concurrent reads without locks.

LAW-005  No entity store operation executes without tenant context.
         The entity store driver returns ErrMissingTenantContext
         if called without tenant context. No exception for background
         operations — background operations set tenant context explicitly.

LAW-006  The outbox entry is written in the same transaction as the
         entity record. Workflow dispatch occurs after the transaction
         commits. Workflow dispatch is never inside a database transaction.

LAW-007  Hook execution order within a single module is declaration order.
         Hook execution order across modules is topological sort order
         of the module dependency graph. Both are deterministic given
         identical module graph.

LAW-008  No layer imports from a higher layer.
         Store may not import Workflow, API, or UI packages.
         Domain may not import API or UI packages.
         API may not import UI packages.
         UI may not import any non-kernel framework package.

LAW-009  Driver interfaces (driver.EntityStore, driver.SessionStore,
         driver.WorkflowStarter, driver.UIRenderer, driver.SearchDriver)
         never gain new methods after v1.0. New capabilities are new
         optional interfaces, type-asserted at runtime.

LAW-010  The compiled schema content hash is the single source of truth
         for schema identity. Cache keys, cluster coordination, and
         rollout validation all use this hash. No other schema identifier
         exists.

LAW-011  Entity names are globally unique across all modules in a binary.
         Entity name format is {module}_{noun}. The compiler detects
         and rejects duplicate entity names at compilation, not at startup.

LAW-012  Migrations are append-only. No migration modifies or removes
         a prior migration. Each migration has a deterministic version
         number. Migrations apply in version order. No gaps. No skips.

LAW-013  Sensitive fields (Sensitive: true) are excluded from all logs,
         error messages, hook inputs (unless the hook declares explicit
         read permission), and API responses (unless the viewer holds
         the field's access permission).

LAW-014  The Filter DSL wire format is versioned. A filter without a
         version field is rejected. A filter with an unrecognized version
         returns HTTP 422 with error.code = "filter_version_not_supported".

LAW-015  Tenant isolation is enforced by the database. Application-layer
         tenant filtering (WHERE tenant_id = ?) is supplementary and
         must never be the sole isolation mechanism. RLS (or equivalent
         driver-level guarantee) is always the primary enforcement.

LAW-016  The workflow ID format is permanently:
         {tenant-uuid}.{entity-type}.{record-id}.{event}.{workflow-function}
         No component may be added, removed, or reordered.
         Workflow IDs are stored in external systems (Temporal history)
         and cannot be migrated.

LAW-017  Module hook registration on entities owned by other modules
         requires the owning entity to declare HookPolicy.Open.
         Default is HookPolicy.Sealed for financial and IAM entities.
         Default is HookPolicy.Open for all other entities.

LAW-018  No application code reads from or writes to the outbox table
         directly. The outbox table is the exclusive domain of the
         outbox driver. Application code produces events through the
         entity lifecycle engine only.

LAW-019  The compiled schema is identical across all instances running
         the same binary with the same environment variable set.
         Conditional entity registration based on environment variables
         is forbidden. Feature flags are the mechanism for runtime
         behavioral variation.

LAW-020  Every public API in def/, filter/, and driver/ that is frozen
         at v1.0 carries a stability annotation in its godoc comment:
         // Stability: Kernel — this signature is frozen for the lifetime
         // of this major version.
         Absence of this annotation means the API is experimental.
```

---

## 10. Final v1.0 Blocking Issues

These are the only issues that, if unresolved, guarantee awo/v2 before 2035.

---

**BLOCK-001 — HookSet extensibility**

`HookSet` as a struct in `def/` makes new lifecycle stages impossible without modifying the frozen kernel. Replace with keyed `[]HookRegistration` before any third-party module writes a hook. Once modules write hooks against `HookSet` field names, the field names become a permanent public API.

Condition: **Resolved when `HookSet` struct no longer exists in `def/`.**

---

**BLOCK-002 — FieldType extensibility**

`FieldType` as a closed constant set prevents domain-specific field types. Change to open string type with compiler-verified registration. Once modules use built-in field type constants, the constant values become permanent.

Condition: **Resolved when third-party modules can define FieldType constants without modifying def/.**

---

**BLOCK-003 — WorkflowTrigger.On extensibility**

Same root cause as BLOCK-001 and BLOCK-002. Trigger events as closed constants prevent domain-specific workflow triggers. Change to open string type.

Condition: **Resolved when third-party modules can define trigger events without modifying def/.**

---

**BLOCK-004 — nil ViewerContext semantics**

System operations without a human viewer are undefined in the current contract. Every module that needs background job support will implement a different convention. Once conventions are established and third-party code depends on them, formalizing the contract is a breaking change.

Condition: **Resolved when the kernel formally specifies SystemViewer or TenantScopedSystemViewer as the canonical viewer for background operations, with specified policy evaluation behavior.**

---

**BLOCK-005 — CreateInput / UpdateInput are unspecified kernel types**

The EntityRepository interface references `CreateInput` and `UpdateInput` without specifying their structure. Every driver implementation assumes a different structure. Once driver implementations exist, the structure becomes the de facto contract. Specifying it after drivers exist is a breaking change.

Condition: **Resolved when CreateInput and UpdateInput are formally specified types in def/ with declared semantics for field source routing and optimistic locking.**

---

**BLOCK-006 — Policy evaluation semantics are unspecified**

AND vs. OR vs. first-match is undefined. Every module author that writes a policy guesses. Once policies are in production, changing semantics is a silent security regression. This cannot be changed after third-party policies exist.

Condition: **Resolved when policy evaluation semantics are formally declared in a permanent specification document and enforced by the compiler's policy analysis pass.**

---

**BLOCK-007 — Filter wire format has no version field**

Once any client serializes a filter and sends it to Awo, the format is a public API. Without a version field, the format cannot evolve. Adding a version field after clients exist is a breaking change.

Condition: **Resolved when the filter wire format specification is published with a mandatory `"v"` version field and the parser rejects filters without it.**

---

**BLOCK-008 — Entity and field lifecycle (retirement) is unspecified**

No mechanism exists for retiring entities or fields safely. Every deployment that needs to retire a field will implement an ad-hoc approach. Once ad-hoc approaches are in production, they become the de facto standard. Specifying the lifecycle after fields are retired in the wild is a breaking change.

Condition: **Resolved when FieldDef and EntityDefinition both carry a formal Lifecycle field with documented state transitions and compiler-enforced semantics.**

---

**BLOCK-009 — Cross-module hook and action sovereignty**

No mechanism prevents Module B from registering hooks and actions on Module A's entities. Once this pattern exists in the ecosystem, prohibiting it is a breaking change for all modules that use it.

Condition: **Resolved when HookRegistration carries an owning module identifier and EntityDefinition carries a HookPolicy declaration, enforced by the compiler.**

---

**BLOCK-010 — Schema divergence detection across instances**

Conditional entity registration based on environment variables can produce instances in the same cluster with different compiled schemas, silently serving different behaviors. Once operational procedures assume this is possible (e.g., gradual module rollout via env var), enforcing uniformity is a breaking operational change.

Condition: **Resolved when the compiler analysis pass rejects conditional entity registration, and the runtime cluster coordination mechanism detects and alerts on schema hash divergence.**

---

## Summary

Ten blocks. All architectural. None cosmetic.

BLOCK-001 through BLOCK-003 share the same root cause: closed enum/struct types in `def/` for concepts that must be extensible. Fix once with the open string type pattern.

BLOCK-004 through BLOCK-006 are unspecified semantic contracts that will produce ecosystem fragmentation the moment third-party code makes assumptions.

BLOCK-007 and BLOCK-008 are permanent API surfaces without version or lifecycle semantics — time bombs that will explode the first time evolution is needed.

BLOCK-009 and BLOCK-010 are distributed system correctness properties that cannot be added after the ecosystem forms without breaking existing modules or operational procedures.

Every other issue identified across all four reviews is either already resolved (per the precondition of this review) or is future work (v1.1 through v2.0).

Close these ten. Then declare v1.0.

---

*Signed: Language Specification Committee*
*Date: 2025-07-06*
*Status: NOT READY FOR v1.0 — 10 blocking issues*
*Resubmit after BLOCK-001 through BLOCK-010 are resolved.*
