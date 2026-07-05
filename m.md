This is a continuation of an ongoing architecture and design effort, not a request to critique the previous response in isolation.

Treat all previous architecture documents as the current working specification. Your role is to continue refining and completing the architecture until we reach a stable v1.0 design and then proceed toward implementation documentation.

The project is Awo Framework, a Go-native ERP application framework whose central principle is that a single "EntityDefinition" declaration becomes the source of truth for persistence, APIs, SDUI, access control, workflows, auditing, events, and future capabilities. We are intentionally designing the framework before implementation so that every package, interface, dependency, and extension point remains coherent for many years.

Assume the previous response is the latest architectural state, not necessarily the final truth. You should continuously review it for correctness, improve it where necessary, remove unnecessary complexity, simplify APIs, and challenge any design that violates the framework philosophy. Do not preserve ideas simply because they already exist.

From this point onward, act as the project's principal software architect. Your responsibility is to evolve the design from architecture to production-ready framework specifications. Every response should move the project closer to implementation.

When making recommendations:

- Keep the architecture minimal, stable, and idiomatic Go.
- Preserve the dependency direction and kernel philosophy unless there is a compelling architectural reason to change it.
- Prefer long-term maintainability over short-term convenience.
- Think like the framework will still be maintained ten years from now.
- Avoid introducing abstractions without clear value.
- Question every package, interface, type, and dependency before accepting it.

After the architecture is finalized, continue producing the remaining project artifacts in order, including but not limited to:

1. Complete package specifications.
2. Public APIs.
3. Interface contracts.
4. Runtime lifecycle.
5. Extension mechanisms.
6. Security model.
7. Multi-tenant architecture.
8. Storage architecture.
9. Workflow architecture.
10. SDUI architecture.
11. Authentication and authorization.
12. Testing strategy.
13. CLI specification.
14. Code generation strategy.
15. Migration strategy.
16. Implementation roadmap.
17. Documentation for framework developers.
18. Documentation for module developers.
19. Coding standards.
20. Step-by-step implementation plan.

Do not skip ahead to writing code unless the architecture and documentation for that area are considered complete.

For every architectural decision:

- Explain the reasoning.
- Discuss trade-offs.
- Identify potential future problems.
- Recommend the most maintainable solution.

Treat every response as another chapter of the Awo Framework design document. Build on previous work instead of restarting or repeating it.

I will now provide the latest architecture request/section that should be considered the continuation point.


Continue directly from there and keep the same response style.

Response format:

...


You are the principal software architect responsible for reviewing and evolving the permanent architecture of Awo Framework. Treat the current architecture as the baseline specification, not as implementation details.

Your task is NOT to redesign the framework from scratch.

Your task is to perform an architectural stress test against the next 10–15 years of ERP framework evolution and identify weaknesses that only become visible at scale.

Assume Awo will eventually power:

- Multi-company and multi-tenant ERP
- Government systems
- Healthcare
- Banking back offices
- Manufacturing
- Logistics
- Retail and POS
- Workflow-heavy enterprises
- AI-assisted applications
- Event-driven architectures
- SaaS with tens of thousands of tenants
- Self-hosted deployments
- Cloud-native deployments
- Plugin ecosystems
- Third-party modules

Evaluate the architecture as if it must remain stable until 2040.

Do not focus on naming, formatting, style, or minor refactoring.

Instead evaluate deep architectural qualities:

1. Kernel Stability
- Are the kernel abstractions truly permanent?
- Which abstractions are accidental rather than essential?
- Which types are likely to require breaking changes in the future?
- Which concepts belong outside the kernel?

2. Dependency Architecture
- Is the dependency graph mathematically acyclic?
- Are there hidden dependency inversions?
- Which packages are likely to become god packages?
- Which packages violate Clean Architecture or Ports & Adapters?

3. Domain Model
Critique every primitive:

- Entity
- Field
- Record
- MutableRecord
- Viewer
- Hook
- Policy
- Action/Command
- Event
- WorkflowTrigger
- Scope
- Storage
- Edge

Should any be split?

Should any disappear?

Are any missing?

4. Extensibility

Determine whether the architecture can support WITHOUT BREAKING CHANGES:

- GraphQL
- gRPC
- Event sourcing
- CQRS
- Read models
- AI agents
- CRDT/offline sync
- Edge computing
- Serverless
- Multi-region
- Plugin marketplace
- Tenant-defined entities
- Tenant-defined workflows
- External storage engines
- Alternative UI renderers
- Mobile-first APIs
- Streaming APIs
- Background jobs
- Domain events

Explain exactly where extension points should exist.

5. Runtime

Stress test startup and runtime lifecycle.

Look for:

- race conditions
- initialization order issues
- registry problems
- dependency cycles
- transaction boundaries
- workflow execution
- hook ordering
- panic isolation
- shutdown correctness

6. Scalability

Assume:

100,000 entities

10 million records per tenant

50,000 tenants

hundreds of modules

thousands of hooks

thousands of concurrent requests

Identify bottlenecks.

Suggest architectural improvements instead of implementation optimizations.

7. Plugin Architecture

Can third-party developers safely extend the framework?

Can plugins:

- register entities
- register drivers
- register validators
- register SDUI widgets
- register workflow engines
- register authentication providers
- register storage engines
- register report engines

without modifying framework code?

If not, design the missing extension model.

8. Evolution Safety

For every exported package determine:

- Stable forever
- Stable with additions
- Internal only
- Should be hidden

Identify every API that should never be exposed publicly.

9. Architectural Risks

Find hidden risks that are easy to miss, including:

- future coupling
- abstraction leakage
- semantic duplication
- capability overlap
- accidental complexity
- framework lock-in
- package explosion
- reflection temptation
- code generation creep
- metadata drift

10. Missing Capabilities

Identify important framework capabilities that should exist but are absent.

Examples (not exhaustive):

- scheduler
- search abstraction
- cache abstraction
- notification abstraction
- secrets
- feature flags
- localization
- versioned APIs
- migrations
- observability
- health system
- diagnostics
- metadata API
- introspection
- testing infrastructure

Only recommend additions if they fit the architecture naturally.

11. Comparison

Compare Awo's architecture against the architectural principles (not feature lists) of:

- Frappe
- Odoo
- Django
- Laravel
- Spring Boot
- .NET Aspire
- Kubernetes control plane
- Temporal
- Ent Framework
- Atlas
- Encore
- Buffalo
- Hasura

Explain where Awo is stronger, weaker, or fundamentally different.

12. Permanent Architecture

Finally, propose the architecture that should survive for the next decade.

Do NOT redesign everything.

Preserve the philosophy:

"One immutable entity declaration drives the entire framework."

Improve only what is necessary.

Every recommendation must satisfy:

- simpler than the current design
- more extensible
- less coupled
- easier to test
- easier to reason about
- future-proof

Output requirements:

- Do not summarize.
- Produce a comprehensive architectural review.
- Use diagrams where useful.
- Include dependency graphs.
- Include package layouts.
- Include rationale for every major recommendation.
- Clearly distinguish between mandatory changes before v1.0 and improvements that can wait until later.
- Favor long-term architectural correctness over short-term implementation convenience.

Treat this as the architecture review that determines whether Awo Framework can become a world-class ERP framework rather than just another CRUD framework.



reponse 

```txt 

● Awo Framework — Architectural Stress Test: 2025–2040

  ---
  Preamble

  This review accepts the architectural baseline as given and attacks it. Findings are categorized: CRITICAL (breaks
  before v1.0 or under load), STRUCTURAL (correct today, breaks at scale or over time), MISSING (absent capability
  that will be needed). Minor issues excluded.

  ---
  Section 1 — Kernel Stability

  The kernel as defined

  awo/def     — Entity, Field, Record, MutableRecord, Viewer, Hook, Policy, Op, Action
  awo/filter  — Filter DSL

  Verdict: Kernel is 70% permanent. Four abstractions will require breaking changes.

  ---
  1.1 FieldType as closed enum — CRITICAL structural flaw

  type FieldType uint8
  const (
      Text FieldType = iota + 1
      Currency
      NamingSeries
      // ...
  )

  Why this breaks: Government systems need NationalID, IBAN, Coordinates. Healthcare needs HL7 FHIR types, SNOMED
  codes. Manufacturing needs UnitOfMeasure with conversion rules. Banking needs SWIFT codes, LEI identifiers. Every
  vertical will need field types that don't exist in the closed set.

  The wrong fix: Switch to string constants (stringly typed). No compile-time guarantee, no exhaustive switch, no
  driver mapping.

  The right fix: Two-tier type system.

  // Closed set of structural kinds — drivers understand these
  type FieldKind uint8
  const (
      KindText   FieldKind = iota + 1
      KindInt
      KindDecimal  // all money and precise numbers
      KindBool
      KindDate
      KindDateTime
      KindSelect
      KindLink
      KindJSON
      KindFile
      KindCustom   // ← extension point
  )

  // Open set of semantic types — for validation, UI, serialization
  type FieldType struct {
      Kind    FieldKind
      Subtype string    // "iban", "national-id", "uom", "hl7-fhir-date", ""
  }

  // Registered semantic handlers
  type SemanticHandler interface {
      Validate(v any) error
      Normalize(v any) (any, error)
      Widget() string   // SDUI widget name
  }

  FieldKind is the closed structural enum. Drivers map FieldKind to column types. FieldType.Subtype is an open
  semantic extension. Semantic handlers are registered in a SemanticRegistry. This pattern survives 20 years without a
   breaking change to the kernel.

  ---
  1.2 Record.Get(field) any — CRITICAL type safety gap

  type Record interface {
      Get(field string) any
      // ...
  }

  Every caller type-asserts. Type assertion failure = runtime panic. In banking/healthcare, one wrong type assertion
  in a hook crashes a financial transaction.

  Fix: generics

  type Record interface {
      GetAny(field string) any
      ID() uuid.UUID
      TenantID() uuid.UUID
      EntityName() string
      Version() int64  // for optimistic locking — see 3.x
  }

  // Package-level helpers — not on interface (interfaces stay small)
  func Get[T any](rec Record, field string) (T, bool) {
      v := rec.GetAny(field)
      t, ok := v.(T)
      return t, ok
  }

  func MustGet[T any](rec Record, field string) T {
      v, ok := Get[T](rec, field)
      if !ok {
          panic(fmt.Sprintf("field %q: type mismatch", field))
      }
      return v
  }

  This is backward compatible — GetAny is the interface method, generic helpers are functions. Drivers implement
  GetAny. Callers use def.Get[decimal.Decimal](rec, "total").

  ---
  1.3 Viewer is too narrow — STRUCTURAL

  type Viewer interface {
      ActorID() string
      TenantID() string
      OrgUnitID() uuid.UUID
      HasRole(string) bool
      IsSystem() bool
  }

  Problems:

  HasRole(string) bool at 50,000 tenants, 1000 hooks per request: 1000 HasRole calls = 1000 function calls. Should be
  Roles() []string. Callers can check membership in a slice.

  OrgUnitID() uuid.UUID — returns one unit. Matrix organizations (healthcare, manufacturing) put people in multiple
  units. Should be OrgUnitIDs() []uuid.UUID.

  Missing Plane() string — "platform" vs "tenant" viewers behave fundamentally differently. Without this, policies
  can't distinguish platform admins from tenant admins without a role convention.

  Missing Claims() map[string]any — extensible principal metadata without interface change. Security clearance,
  regulatory context, data classification access, geographic restriction — these are domain-specific and cannot be
  enumerated in the kernel. Claims() provides the escape hatch.

  type Viewer interface {
      ActorID() string
      TenantID() string
      OrgUnitIDs() []uuid.UUID  // plural
      Roles() []string           // not HasRole
      IsSystem() bool
      Plane() string             // "platform" | "tenant" | "api-client"
      Claims() map[string]any    // extensible
  }

  // Helper — replaces HasRole method
  func HasRole(v Viewer, roles ...string) bool {
      for _, vr := range v.Roles() {
          for _, r := range roles {
              if vr == r { return true }
          }
      }
      return false
  }

  ---
  1.4 StorageKind in the kernel — STRUCTURAL

  type StorageKind uint8
  const (
      StorageSystem StorageKind = iota
      StorageCustom
  )

  System vs Custom is a PostgreSQL implementation detail. The kernel should not know that some entities use JSONB.
  Future storage: time-series (manufacturing telemetry), column store (analytics), graph (logistics network), document
   (clinical notes).

  Fix: Remove StorageKind from def.Entity. Move to driver/pgstore as a driver-side annotation. The driver reads entity
   metadata (e.g., a Table field) and decides how to store. An entity with Table: "" gets JSONB routing. The kernel
  declares intent, not mechanism.

  ---
  1.5 Audited bool and SoftDelete bool — Accidental complexity

  Binary flags that need to become richer declarations:

  Audited bool — healthcare needs field-level sensitivity classification (PHI, PII), not just a flag. GDPR needs
  erasable fields. Banking needs tamper-evident append-only logs. SOX needs access logs. "Audited = true" is
  insufficient.

  SoftDelete bool — GDPR requires hard purge of PII. Banking requires immutable records (neither soft nor hard delete
  allowed). "SoftDelete = bool" covers neither case.

  Fix: Replace with declarative policies:

  type RetentionPolicy uint8
  const (
      RetainForever    RetentionPolicy = iota // immutable — banking, legal
      RetainWithDelete                         // soft delete, never purge
      RetainWithPurge                          // GDPR erasure supported
  )

  type AuditLevel uint8
  const (
      AuditNone    AuditLevel = iota
      AuditWrite              // who changed what, when
      AuditReadWrite          // reads also logged — healthcare, banking
      AuditFull               // field-level diff, all access — SOX/HIPAA
  )

  ---
  1.6 Scope — Insufficient for enterprise hierarchies — STRUCTURAL

  type Scope uint8
  const (
      ScopeGlobal
      ScopeTenant
      ScopeUnit
  )

  ScopeUnit assumes one org tree per tenant. Real enterprises have multiple hierarchy axes:

  - Healthcare: geographic hierarchy + clinical hierarchy + admin hierarchy
  - Manufacturing: plant → line → workstation AND product → category → subcategory
  - Banking: legal entity → cost center → desk AND geography → branch → counter

  ScopeUnit pins "scope" to a single tree. The Tree shape itself is domain-specific.

  Structural fix: Scope declares what dimensions of isolation apply. Actual tree resolution is pluggable.

  type Scope struct {
      Tenant bool
      Unit   bool
      // future axes via Claims-based policy, not kernel types
  }

  // ScopeGlobal, ScopeTenant, ScopeUnit become constructors:
  var (
      ScopeGlobal = Scope{}
      ScopeTenant = Scope{Tenant: true}
      ScopeUnit   = Scope{Tenant: true, Unit: true}
  )

  This is backward compatible and admits future axes without kernel change.

  ---
  Section 2 — Dependency Architecture

  2.1 Dependency graph — correct in theory, leaky in practice

  def ← filter ← registry ← store ← api ← app

  This graph is correct. But three hidden inversions exist:

  Hidden inversion 1: def.Register is in def

  // current
  package def
  func Register(e *Entity) { ... }

  def contains the registration function. Registration requires a mutable global (the registry map). def now has
  mutable global state. Mutable global state in the kernel is a structural violation: the kernel is supposed to be
  pure types.

  Fix: registry.Register only. def stays pure types, no init-time side effects.

  Hidden inversion 2: app depends on platform/*

  platform/iam, platform/tenant, platform/org are framework packages that app wires together. But platform packages
  contain business entities (User, Tenant, OrgUnit). Business entities should not be in the framework — they should be
   in a reference module that users can swap.

  If a user wants to use Keycloak instead of platform/iam, they can't. IAM is baked into bootstrap. This is framework
  lock-in.

  Fix: Platform packages become a contrib module, not part of the framework core.

  framework core: def, filter, registry, store, access, hook, validate, app
  framework contrib: platform/iam, platform/tenant, platform/org, platform/naming

  app.Mount accepts AuthProvider interface instead of hard-wiring platform/iam.

  Hidden inversion 3: api depends on fiber/v2 directly

  api.Handler is a Fiber handler. If Fiber v3 ships breaking changes (happened before), api requires a change. Worse:
  users who want to use net/http or Chi cannot use the framework's API layer.

  Fix: api package exposes http.Handler compatible interface. Thin Fiber adapter in driver/fiber.

  api/    — pure HTTP semantics (Handler interface, routing metadata)
  driver/fiber/ — fiber.Handler adapter
  driver/nethttp/ — net/http adapter

  ---
  2.2 God package risk

  def is growing toward god-package status. Currently contains: Entity, Field, FieldType, Record, MutableRecord,
  Viewer, Hook, Timing, Op, Policy, PolicyVerdict, Action, ActionContext, ActionResult, WorkflowTrigger, Edge,
  PageSet, Scope, ValidationError, DomainError.

  At 100,000 entities and hundreds of plugin authors importing def, the package import graph becomes:

  every plugin → def → [all of the above]

  This is unavoidable for a metadata-driven framework — the kernel must be imported everywhere. But it means def must
  be disciplined: no concrete implementations, no I/O, no non-stdlib imports, ever. One concrete type in def
  contaminates every consumer.

  Current risk: def.FieldError, def.DomainError are concrete error types. They belong in validate package, not def.
  Hooks should return validate.FieldError or plain error. The API layer maps errors to HTTP status codes.

  ---
  2.3 registry.Freeze() breaks plugin ecosystems — CRITICAL

  Freeze() is called at app.Mount. After Freeze, no new entities can register. This is correct for static deployments
  (one binary, all entities known at compile time).

  It is wrong for:
  - Plugin marketplace (load new plugin = load new entities at runtime)
  - Tenant-defined entities (tenant adds custom entity via admin UI = register entity at runtime)
  - Hot reload in development

  Fix: Two-phase registry

  // Phase 1: static registration (init time)
  // All compiled-in entities register here.
  // Freeze() ends phase 1.
  registry.Freeze()

  // Phase 2: dynamic extension (runtime)
  // Plugins and tenant-defined entities use ExtendedRegistry.
  // Extended entities are namespaced: "plugin.acme:invoice-line"
  // Extended entities cannot override static entities.
  // Extended entities can be loaded/unloaded without restart.
  registry.Extend(e *def.Entity) error
  registry.Retract(name string) error

  Phase 2 uses a separate read-write mutex. Phase 1 remains lock-free (immutable after Freeze). Drivers that cache
  entity metadata invalidate on Extend/Retract.

  ---
  Section 3 — Domain Model Critique

  3.1 Policy cannot filter at query time — CRITICAL scalability flaw

  type PolicyFunc func(ctx context.Context, v Viewer, op Op, rec Record) PolicyVerdict

  Policy evaluates per row AFTER fetch. At 10M records per tenant, a List operation that fetches 10M rows then filters
   to 50 is catastrophic.

  This is the most critical architectural flaw. Every ERP system has row-level security at query time. Healthcare:
  doctor sees only their patients. Banking: analyst sees only their desk's trades. Without query-time filtering, the
  framework cannot scale to real enterprise data volumes.

  Fix: Dual policy interface

  // Per-row verdict (existing — for Get, Create, Update, Delete)
  type PolicyFunc func(ctx context.Context, v Viewer, op Op, rec Record) PolicyVerdict

  // Query-time filter contribution (NEW — for List)
  type FilterPolicy func(ctx context.Context, v Viewer) *filter.Filter

  type Policy struct {
      Ops          Op
      Eval         PolicyFunc    // per-row — used for mutations and single-record reads
      ListFilter   FilterPolicy  // query-time — injected into WHERE clause for lists
  }

  The api.Handler List implementation collects all ListFilter contributions, ANDs them with the user's explicit
  filter, passes to the store. Never fetches rows that the viewer can't see.

  Built-in policy constructors produce both:

  func AllowRole(roles ...string) Policy {
      return Policy{
          Eval: func(ctx, v, op, rec) PolicyVerdict {
              if HasRole(v, roles...) { return Allow }
              return Deny
          },
          ListFilter: nil, // role check doesn't restrict which rows, only access
      }
  }

  func AllowOwner(field string) Policy {
      return Policy{
          Eval: func(ctx, v, op, rec) PolicyVerdict {
              if rec == nil { return Skip }
              if fmt.Sprint(rec.GetAny(field)) == v.ActorID() { return Allow }
              return Deny
          },
          ListFilter: func(ctx, v) *filter.Filter {
              return filter.Eq(field, v.ActorID())
          },
      }
  }

  ---
  3.2 Missing: Aggregate Root — STRUCTURAL

  Every entity gets CRUD routes. InvoiceLine gets POST /api/invoice-lines. This is wrong: invoice lines belong to an
  invoice aggregate. Creating a line without an invoice context breaks domain invariants.

  Missing primitive: ChildOf declaration.

  // Invoice line belongs to invoice aggregate.
  // No top-level REST routes generated.
  // Access only via /api/invoices/:id/lines
  var InvoiceLineEntity = def.Entity{
      Name:    "invoice-line",
      ChildOf: "invoice",   // ← aggregate declaration
      // ...
  }

  Framework generates nested routes: GET /api/invoices/:id/lines, POST /api/invoices/:id/lines, PATCH
  /api/invoices/:id/lines/:lid. Never top-level /api/invoice-lines.

  Parent record is passed to hooks, policies, and actions of the child. Child creation validates parent ownership.

  ---
  3.3 Missing: Event as first-class primitive — STRUCTURAL

  Current framework conflates three different concepts:

  ┌──────────────────────────────┬───────────────────────────┬─────────┐
  │           Concept            │          Current          │ Correct │
  ├──────────────────────────────┼───────────────────────────┼─────────┤
  │ Framework lifecycle reaction │ Hook                      │ Same    │
  ├──────────────────────────────┼───────────────────────────┼─────────┤
  │ Business fact that occurred  │ Hook{Timing: AfterCommit} │ Event   │
  ├──────────────────────────────┼───────────────────────────┼─────────┤
  │ User-initiated command       │ Action                    │ Same    │
  └──────────────────────────────┴───────────────────────────┴─────────┘

  AfterCommit hooks are informal events. This is sufficient for simple cases. It breaks for:
  - Event sourcing (need to replay events)
  - CQRS (need events to build read models)
  - Event streaming (Kafka, NATS consumers)
  - Cross-entity event correlation (invoice submitted → check customer credit → notify sales)

  Fix: Event as first-class type in def

  // An event is a named, versioned, immutable fact.
  type EventDef struct {
      Name    string    // "invoice.submitted"
      Version int       // schema version for compatibility
      Entity  string    // which entity produces this event
      On      Op        // which operation(s) trigger it
      Condition func(old, new Record) bool  // nil = always emit
      Build   func(ctx context.Context, old, new Record) (any, error)
  }

  // Entity declares what events it emits:
  var InvoiceEntity = def.Entity{
      Events: []def.EventDef{
          {
              Name:    "invoice.submitted",
              Version: 1,
              On:      def.Update,
              Condition: func(old, new Record) bool {
                  return old.GetAny("status") == "draft" &&
                         new.GetAny("status") == "submitted"
              },
              Build: func(ctx, old, new) (any, error) {
                  return InvoiceSubmittedEvent{
                      InvoiceID:  MustGet[uuid.UUID](new, "id"),
                      CustomerID: MustGet[uuid.UUID](new, "customer"),
                      Total:      MustGet[decimal.Decimal](new, "total"),
                  }, nil
              },
          },
      },
  }

  Events are emitted after commit (via the existing AfterCommit mechanism, now named). EventBus interface receives
  them. Implementations: in-process subscribers, Kafka, NATS, Temporal signals, audit log writer.

  ---
  3.4 Missing: Record.Version() for optimistic locking — CRITICAL for banking/inventory

  type Record interface {
      GetAny(field string) any
      ID() uuid.UUID
      TenantID() uuid.UUID
      EntityName() string
      Version() int64  // ← MISSING
  }

  Without optimistic locking: two users fetch an inventory record, both decrement stock, both update — last write
  wins, double-decrement lost. This is a fundamental correctness failure for inventory, banking, and any concurrent
  write scenario.

  Version() is a system field, not a declared field. Store increments it atomically on every update. MutableRecord
  carries the version it was read at. Store rejects update if current_version != record_version with ErrConflict.

  ---
  3.5 Missing: BeforeList hook — STRUCTURAL

  Hook timings: BeforeValidate, BeforeSave, AfterSave, AfterDelete, AfterCommit.

  Missing: BeforeList. Use cases:

  - Inject default filter (always show non-deleted records, even if SoftDelete is not framework-managed)
  - Add mandatory sort for deterministic pagination
  - Inject viewer-context filter before the query hits the DB (alternative to ListFilter in policy — hook approach is
  more flexible)
  - Audit list queries (HIPAA: who queried what patient records)

  const BeforeList Timing = ...

  // Example: always filter out archived records in list view
  def.OnBeforeList("hide-archived", func(ctx context.Context, opts *store.ListOptions, v def.Viewer) error {
      if opts.Filter == nil {
          opts.Filter = filter.Neq("status", "archived")
      } else {
          opts.Filter = filter.And(opts.Filter, filter.Neq("status", "archived"))
      }
      return nil
  })

  ---
  3.6 WorkflowTrigger is Temporal-specific — STRUCTURAL

  type WorkflowTrigger struct {
      On           def.Submit
      WorkflowType string    // Temporal workflow function name
      TaskQueue    string    // Temporal task queue
      IDFunc       func(Record) string
  }

  WorkflowType and TaskQueue are Temporal concepts. Amazon SWF, Conductor, Cadence, and future engines won't match.
  The kernel has been coupled to one workflow engine.

  Fix:

  type WorkflowTrigger struct {
      On        Op
      Engine    string   // "temporal" | "conductor" | "inline" | plugin-registered name
      Type      string   // engine-specific workflow type name
      Queue     string   // engine-specific routing key
      Condition func(old, new Record) bool  // nil = always trigger
      IDFunc    func(Record) string         // required — idempotency key
      Input     func(ctx context.Context, rec Record) (any, error)
  }

  workflow.DriverRegistry maps Engine string to workflow.Driver interface. Temporal driver registered in
  driver/temporal. Framework core knows nothing about Temporal.

  ---
  3.7 Edge is underdeveloped — STRUCTURAL

  Edge currently handles eager-loading. It needs:

  type Edge struct {
      Name        string
      Target      string     // entity name
      Field       string     // FK field on this entity (for outbound) or target entity (for inbound)
      Cardinality Cardinality // OneToOne, OneToMany, ManyToMany
      Direction   Direction  // Outbound (this.field = target.id), Inbound (target.field = this.id)
      Eager       bool       // load with parent query
      Cascade     Cascade    // OnDelete behavior: Restrict, SetNull, Delete
  }

  Without cardinality and cascade declarations, the store can't enforce FK integrity and migrations can't generate
  correct SQL constraints.

  ---
  3.8 Hook lacks dependency injection — STRUCTURAL

  type Hook struct {
      Run func(ctx context.Context, rec MutableRecord) error
  }

  Hook only receives (ctx, rec). What if a BeforeSave hook needs to query related records?

  Bad pattern (will be written):
  // Hook author pulls store from context — hidden coupling
  store := StoreFromContext(ctx)
  customer, _ := store.Get(ctx, customerID)

  This creates invisible coupling between hooks and context structure. Untestable without context setup. Context
  becomes a dependency injection bag.

  Better pattern:

  // Hook is a struct, not a function — dependencies explicit
  type InvoiceCreditCheckHook struct {
      CustomerStore store.EntityStore  // injected at registration
  }

  func (h *InvoiceCreditCheckHook) Run(ctx context.Context, rec def.MutableRecord) error {
      customerID := MustGet[uuid.UUID](rec, "customer")
      customer, err := h.CustomerStore.Get(ctx, customerID)
      // ...
  }

  Framework change required: Hook.Run stays as func(ctx, rec) error. But the framework should provide a HookFactory
  interface for hooks that need dependencies, wired by app.Mount:

  type HookFactory interface {
      CreateHook(deps HookDeps) HookFn
  }

  type HookDeps struct {
      Store   store.TenantStore
      Events  event.Emitter
      // other framework services
  }

  Alternatively: accept that hooks pulling from context is the pattern, but document exactly what is in context and
  provide typed accessors. Context pollution is real but pragmatic.

  ---
  Section 4 — Extensibility Analysis

  4.1 Extension point map

  Capability          Current state          Required extension point
  ──────────────────  ─────────────────────  ────────────────────────
  Storage engines     pgstore only           store.Driver interface
  Auth providers      platform/iam only      AuthProvider interface
  UI renderers        amis only              UIRenderer interface
  Workflow engines    temporal only          workflow.Driver interface
  HTTP adapters       fiber only             http.Handler compatible api
  Search engines      none                   SearchProvider interface
  Cache providers     redis hard-wired       CacheProvider interface
  Event buses         none                   EventBus interface
  Schedulers          none                   Scheduler interface
  Report engines      none                   ReportEngine interface
  Validators          closed set             SemanticRegistry
  SDUI widgets        amis fields only       WidgetRegistry
  Feature flags       none                   FlagProvider interface
  Notification        none                   Notifier interface

  Each extension point follows the same pattern:

  // Interface defined in framework package (not driver)
  type SearchProvider interface {
      Index(ctx context.Context, e *def.Entity, rec def.Record) error
      Search(ctx context.Context, e *def.Entity, query string, opts SearchOpts) ([]def.Record, error)
      Delete(ctx context.Context, e *def.Entity, id uuid.UUID) error
  }

  // Registered in driver package
  driver/meilisearch/provider.go:
  func New(client *meilisearch.Client) search.Provider { ... }

  // Wired in app.Mount via Options
  type Options struct {
      Search search.Provider  // nil = no search indexing
      // ...
  }

  ---
  4.2 GraphQL — addable without framework changes (conditional)

  GraphQL needs:
  1. Type definitions from entity declarations — possible today via registry.All()
  2. Resolvers that use store.EntityStore — possible today
  3. Policy enforcement — possible via access.Enforcer — but ListFilter policies are critical (section 3.1)
  4. Subscription (real-time) — not possible today — needs event bus

  A driver/graphql package can implement GraphQL without framework changes if ListFilter policies exist. Without them,
   GraphQL would have to post-filter — unacceptable.

  ---
  4.3 Event sourcing — needs Event primitive

  Without EventDef as a first-class primitive (section 3.3), event sourcing requires:
  - Event log as a storage engine (implement store.Driver)
  - Event replay mechanism (needs framework hook for "rebuild state from events")
  - Snapshot support

  The framework can support event sourcing as a storage driver (Engine: "eventstore") if EventDef exists and
  AfterCommit is reliable (outbox pattern — section 5.4).

  ---
  4.4 CRDT/offline sync — requires kernel extension

  CRDT semantics (Last-Write-Wins, counter, set, OR-set) need to be declared per field. The kernel doesn't have this.
  Adding CRDT means:

  type Field struct {
      // ...
      CRDT CRDTKind  // None, LWW, Counter, Set (add-only), ORSet (add/remove)
  }

  This is a kernel change. But it's additive (zero value CRDTKind = none = current behavior). Safe to add in v1.x
  without breaking changes.

  ---
  4.5 Tenant-defined entities — conflicts with Freeze() — CRITICAL

  SaaS tenants want custom entities (not just custom fields on existing entities, but entirely new entity types).
  Current architecture: Freeze() prevents runtime registration.

  Two-phase registry (section 2.3) solves this architecturally. But tenant-defined entities also need:
  - Tenant-scoped schema (tenant's custom entity is invisible to other tenants)
  - Migration-less schema change (no SQL migration for JSONB-backed tenant entities)
  - Runtime route registration (new entity = new API routes)
  - Runtime SDUI page generation (new entity = new navigation entry)

  All of this is architecturally achievable:
  - JSONB backend for tenant entities (already exists via StorageCustom / custom_entity_records)
  - Route registration must happen on Extend, not just at startup
  - SDUI navigation regenerated on Extend
  - registry.Extend takes entity scoped to tenant(s)

  The hard part: tenant entity schema must not leak to other tenants. Extended entities must carry TenantID in their
  declaration so For() routing scopes correctly.

  ---
  4.6 Multi-region / data residency — not addressable by framework alone

  GDPR Article 46, HIPAA, PDPA require that data for a given tenant/user stays in a specific geographic region. The
  current architecture has no concept of data residency.

  Architectural fix: Tenant record carries DataRegion string. TenantStore.For() routes to the correct regional pool.
  Regional pools are configured at startup, not per-entity.

  This is an infrastructure configuration concern. The framework can define the routing interface; the deployment
  wires the pools.

  type Options struct {
      RegionalPools map[string]*pgxpool.Pool // region code → pool
      // When set, TenantStore routes to tenant's regional pool
  }

  ---
  4.7 Serverless — AfterCommit hooks break

  In a serverless environment (AWS Lambda, Google Cloud Run), process exits after request. AfterCommit hooks fire in
  the same process, after the HTTP response is sent. In serverless, the process may terminate before RunAfterCommit
  completes.

  Fix: Outbox pattern (also needed for correctness generally — section 5.4). AfterCommit hooks write to an outbox
  table (inside the transaction). A separate worker (persistent process) delivers from the outbox. The hook runner
  becomes a deliverer, not an inline caller.

  With this fix, serverless is supported: the Lambda handles the request, the outbox is populated, a persistent worker
   delivers events/workflows.

  ---
  Section 5 — Runtime Stress Test

  5.1 Race condition: bindWorkflowTriggers mutates registered entities

  func bindWorkflowTriggers(exec *workflow.Executor) {
      for _, d := range def.All() {
          for _, t := range d.WorkflowTriggers {
              d.Hooks = append(d.Hooks, def.AfterCommitHook(...))
          }
      }
  }

  d.Hooks = append(d.Hooks, ...) mutates a registered entity after Freeze(). If any concurrent reader (incoming
  request during slow startup) reads d.Hooks while this append is in progress: data race.

  Fix: Workflow hooks must be attached before Freeze(). Or hooks must be stored in a separate structure that's built
  once at startup and never mutated.

  Correct startup order:
  1. All init() complete → entities registered
  2. bindWorkflowTriggers (mutates d.Hooks)
  3. registry.Freeze() — now immutable
  4. Begin serving requests

  ---
  5.2 Race condition: policy panics are unrecovered

  access.Enforcer.Allow() calls user-provided PolicyFunc. If a policy panics, the panic propagates to the request
  goroutine and produces a 500 with a stack trace (depending on Fiber's panic recovery position in middleware chain).

  Fix: Enforcer wraps policy calls with recover:

  func (e Enforcer) Allow(ctx, entity, viewer, op, rec) (allow bool, err error) {
      defer func() {
          if r := recover(); r != nil {
              err = fmt.Errorf("policy panic: %v", r)
          }
      }()
      // ... evaluate policies
  }

  ---
  5.3 Hook ordering across plugins is undefined

  Plugin A registers BeforeSave("validate-credit-limit", ...) on Invoice. Plugin B registers
  BeforeSave("validate-inventory", ...) on Invoice. Execution order depends on which plugin's init() ran first, which
  depends on import order.

  For correctness: credit limit check must run before inventory reservation (no point reserving inventory if customer
  can't pay). But there's no mechanism to declare this.

  Fix: Hook priority

  type Hook struct {
      Name     string
      When     Timing
      Ops      Op
      Priority int    // lower runs first; default 0; negative = before default; positive = after
      Run      func(ctx context.Context, rec MutableRecord) error
  }

  Built-in framework hooks (naming series assignment, RLS context setting) use Priority: -100. Plugin hooks default to
   0. Domain hooks can use positive priority to run last.

  ---
  5.4 AfterCommit reliability — CRITICAL for financial systems

  DB transaction commits
  ↓
  RunAfterCommit(ctx, entity, op, rec) called
  ↓ ← process crash here = workflow never starts
  AfterCommit hooks execute (start Temporal workflow, emit event)

  If process crashes between commit and RunAfterCommit, the event/workflow is permanently lost. For a financial
  system, this means "payment received" event never emitted, accounting never updated.

  Fix: Transactional Outbox

  // Inside the same DB transaction as the entity mutation:
  // 1. Save entity record
  // 2. Write to outbox table:
  INSERT INTO awo_outbox (entity, record_id, op, payload, created_at)
  VALUES ($1, $2, $3, $4, NOW())

  // Separate Outbox Worker (persistent, not serverless):
  // - Polls awo_outbox for undelivered rows
  // - Calls registered AfterCommit handlers
  // - Marks delivered
  // - Supports exactly-once delivery via idempotency keys

  This is the correct pattern for distributed systems. The outbox table is the durability guarantee. AfterCommit hooks
   become outbox consumers.

  // awo/outbox package
  type OutboxWorker struct {
      Store   store.TenantStore  // to read outbox
      Emitter event.Emitter
      Executor workflow.Executor
  }
  func (w *OutboxWorker) Run(ctx context.Context) error // long-running worker

  ---
  5.5 Startup validation — schema drift undetected

  Framework registers entities. DB schema exists. No comparison made at startup. Drift accumulates silently. One day a
   field renamed in Go but not in SQL = queries return wrong values with no error.

  Fix: Schema verifier

  // awo/verify package
  // Called during app.Mount, before Freeze, with a DB connection
  func VerifySchema(ctx context.Context, pool *pgxpool.Pool) error {
      for _, e := range registry.All() {
          if e.Storage == StorageSystem {
              if err := verifyEntity(ctx, pool, e); err != nil {
                  return fmt.Errorf("schema mismatch: %s: %w", e.Name, err)
              }
          }
      }
      return nil
  }

  // verifyEntity checks: table exists, columns exist, column types match field types
  // Fails fast on mismatch — better to crash at startup than corrupt data silently

  This is non-optional for production systems. Atlas framework does this. Awo should too.

  ---
  5.6 Shutdown correctness

  Current graceful shutdown:
  1. Stop accepting new requests
  2. Drain in-flight requests
  3. Stop Temporal worker
  4. Close DB pool

  Missing:
  - AfterCommit hooks / outbox worker drain (stop accepting, let worker finish pending outbox rows)
  - Entity-level shutdown hooks (release pessimistic locks held by hooks)
  - Health probes should return not-ready before shutdown begins (prevent load balancer from sending new traffic)

  Fix: Add Shutdown interface that capabilities implement:

  type Shutdowner interface {
      Shutdown(ctx context.Context) error  // context carries deadline
  }

  app.Mount collects Shutdowner implementations. On SIGTERM, calls them in reverse registration order with a deadline.

  ---
  Section 6 — Scalability Architecture

  6.1 Policy post-fetch filtering at 10M records

  Already covered in section 3.1. This is the most critical scalability flaw. Without ListFilter in Policy, the
  framework cannot be used for any entity with more than ~10,000 records per tenant when row-level access control is
  required.

  No mitigation except architectural fix.

  6.2 SDUI navigation at 100,000 entities

  GET /sdui/nav → all registered entities → navigation tree

  100,000 entities = 100,000 nav items. Unusable.

  Fix: Nav tree is module-structured, not entity-flat. Navigation declares modules; entities within modules are loaded
   on demand (expand module → fetch entity list). Nav endpoint returns module list only. Entity list within module is
  a separate call.

  Also: Entity.NavVisible bool — entities that are sub-resources (invoice lines) or internal (audit entries) should
  not appear in navigation.

  6.3 Redis schema cache at 50,000 tenants × 100,000 entities

  page:{entityHash}:{viewType}:{tenantID}

  = 100,000 × 4 view types × 50,000 tenants = 20 billion potential cache keys (in theory).

  In practice: only entities actually accessed per tenant are cached. But Redis memory still becomes a concern.

  Fix: TTL reduction + eviction policy awareness. Schema cache should use Redis OBJECT FREQ eviction (LFU policy).
  Keys for entities not accessed in 24h evict first. Current 5min TTL with no eviction policy is wrong for this scale.

  Also: schema cache key should include permissions hash, not just tenantID. A tenant with 10 different role
  configurations needs 10 different cached schemas. Currently the cache key doesn't account for viewer roles.

  6.4 Hook runner at thousands of hooks per entity

  1000 hooks on a high-frequency entity (stock move, 10,000/second) = 1000 hook calls per mutation. Serial execution.

  Fix: Hook parallelism where semantics allow.

  BeforeValidate hooks → serial (order matters for error accumulation)
  BeforeSave hooks     → serial (order matters, can abort)
  AfterSave hooks      → parallel (all inside TX, independently safe)
  AfterCommit hooks    → parallel (no TX, independently safe)

  AfterSave and AfterCommit hooks can run concurrently. errgroup collects errors. AfterSave: first error triggers
  rollback of all. AfterCommit: all run regardless (errors logged/retried).

  6.5 set_tenant_context($1) round-trip per request

  At 1000 concurrent requests × 1 round-trip = 1000 extra DB calls per second just for RLS setup. This is real
  overhead.

  Fix: PgBouncer statement-level caching is impossible for parameterized SET. But: connection affinity. If a tenant's
  requests go to the same connection repeatedly, set_tenant_context becomes a no-op when the same tenant is already
  set. Tenant-affinity routing at PgBouncer layer reduces actual round-trips significantly. This is an infrastructure
  configuration, not a framework change.

  ---
  Section 7 — Plugin Architecture

  7.1 Current state: no plugin architecture

  All extension happens at compile time via Go packages. Third-party code is included as a Go module dependency.
  init() handles registration. This is the simplest model and correct for trusted modules.

  For a plugin marketplace (untrusted third-party code), this is insufficient.

  7.2 Minimal plugin contract for v1.0

  // awo/plugin package
  type Manifest struct {
      Name        string
      Version     string
      APIVersion  string   // framework version this plugin targets: "v1"
      Provides    []string // "entities", "validators", "widgets", "drivers"
      Requires    []string // other plugin names this plugin needs
  }

  type Plugin interface {
      Manifest() Manifest
      Init(reg PluginRegistry) error
      Shutdown(ctx context.Context) error
  }

  type PluginRegistry interface {
      RegisterEntity(e *def.Entity) error
      RegisterValidator(name string, v def.SemanticHandler) error
      RegisterWidget(name string, w sdui.Widget) error
      RegisterDriver(kind string, d store.Driver) error
      RegisterAuthProvider(p auth.Provider) error
  }

  Plugins are loaded as Go .so files or compiled in. The PluginRegistry is the only surface plugins touch — they
  cannot call registry.Register directly (which would bypass plugin manifest tracking).

  7.3 Sandboxing strategy

  For v1.0: no sandboxing. Plugins are trusted (compiled in or .so loaded). Document: "plugins have full process
  access."

  For v2.0: WASM plugins. WASI provides: isolation, capability-based access, cross-platform. Go compiles to WASM.
  Plugin implements a WASM interface; framework hosts via wazero or similar. Plugin can only call functions the host
  explicitly exports (no arbitrary stdlib access).

  WASM plugin interface:

  Plugin → WASM boundary → Host
           ← RegisterEntity(json)
           ← CallHook(ctx, entity, timing, record_json) → error_json
           ← EvaluatePolicy(ctx, viewer_json, op, record_json) → verdict

  The serialization overhead (JSON in/out) is real but acceptable for plugin hooks (not hot path).

  ---
  Section 8 — Evolution Safety

  Package stability classification

  STABLE FOREVER (no changes, ever):
    def/    — kernel types. Any change here = major version bump.
    filter/ — DSL. Additive only. Removing node types = breaking.

  STABLE WITH ADDITIONS:
    registry/  — Add Get variants, ByModule, Search. Never remove.
    store/     — EntityStore interface. Add methods = breaking. Prefer extension interfaces.
    access/    — Enforcer is a struct. Behavior changes = breaking if relied upon.
    hook/      — Runner. Adding hooks is safe. Changing timing semantics = breaking.
    validate/  — Add error types. Never remove.
    event/     — EventDef, Emitter. Add event types. Interface changes = breaking.

  INTERNAL ONLY (not importable by user code):
    driver/*   — Move to internal/driver/*. Users configure via Options, not import.
    sdui/amis/ — Internal renderer detail.
    platform/  — Recommend contrib module, not framework core.

  COMPOSITION ROOT (changes with framework, not stable):
    app/       — Options struct grows. Old option fields deprecated not removed.
                Users import this once; breaking changes here are acceptable with deprecation notice.

  SHOULD NOT EXIST:
    framework/auth/session.go — orphan (already identified, delete)
    framework/naming/series.go — shim (delete)
    def/policy.go, def/viewer.go — empty stubs (delete)

  APIs that must never be public

  registry.freeze()          // internal — startup mechanism, never called by users
  registry.globalRegistry    // internal — never exported
  store.tenantTx             // internal — transaction implementation
  pgstore.*                  // move to internal/driver/pgstore
  iam.tokenHash()            // internal — security implementation
  bootstrap.anonymousViewer  // internal — dev fallback

  ---
  Section 9 — Architectural Risks

  9.1 context.Context as implicit dependency injection bag

  Every hook, policy, and action receives context.Context. Framework sets: TenantIDFromContext, ViewerFromContext,
  TxFromContext. This is convenient but creates invisible coupling.

  A hook that calls TenantIDFromContext(ctx) is implicitly coupled to the middleware that set it. Testing requires
  setting up this context correctly. Documentation can't easily show what's in context.

  Risk: As the framework grows, more things go into context. A hook author writing a credit check needs store access —
   pulls store from context. Now hook is coupled to context structure, not to the store interface. Brittle.

  Mitigation (not a fix): Document exactly what the framework puts in context. Provide typed accessors (awo/ctx
  package). Don't put anything in context that isn't needed by hooks/policies universally. Store access goes through
  HookDeps (section 3.6), not context.

  9.2 Semantic duplication: Op and def.Timing

  Op controls which operations trigger a hook/policy. Timing controls when in the lifecycle. These are orthogonal but
  related. Currently, Op: Create|Update means "for create and update mutations." But Timing: AfterCommit means "after
  transaction." A hook with Timing: AfterCommit, Op: Create is valid.

  The semantics are fine but the relationship is not explicit. A BeforeValidate hook with Op: Delete — what happens?
  Delete doesn't go through validation in the current flow.

  Risk: Framework grows special-case handling for timing/op combinations. Each special case is undocumented
  complexity.

  Fix: Document the valid timing/op matrix explicitly. Enforce at registration time:

  | Timing          | Valid Ops            |
  | BeforeValidate  | Create, Update       |
  | BeforeSave      | Create, Update       |
  | AfterSave       | Create, Update       |
  | AfterDelete     | Delete               |
  | AfterCommit     | Create, Update, Delete |
  | BeforeList      | (no Op — always)     |

  Invalid combinations: Freeze() panics with clear message. This prevents silent no-ops.

  9.3 Framework lock-in via fiber.Ctx in action handlers

  type ActionContext struct {
      Record   Record
      Store    EntityStore
      Viewer   Viewer
      TenantID uuid.UUID
      // no fiber.Ctx — good
  }

  Currently ActionContext is clean. But as framework grows, there will be temptation to add HTTP-level access (request
   body, headers, query params) to ActionContext. Once fiber.Ctx or *http.Request appears in ActionContext, all action
   handlers are forever coupled to HTTP.

  Design rule: ActionContext must never carry HTTP request types. Input parsing is the API layer's responsibility.
  Actions receive typed, validated, deserialized data. The API layer extracts from HTTP, the Action executes business
  logic.

  9.4 Code generation creep

  awo entity generates entity skeleton. awo migrate generates migrations. Over time:

  - awo gen sdui-page (generate custom SDUI page template)
  - awo gen hook (generate hook boilerplate)
  - awo gen workflow (generate Temporal workflow stub)
  - awo gen module (generate module structure)

  Each generator adds surface that must be maintained, tested, and kept in sync with framework evolution. Generators
  that fall behind produce code that doesn't compile against the current framework.

  Rule: Only generate what cannot be expressed in the framework directly. Currently: entity skeleton (saves lookup
  time) + migration (saves SQL knowledge). Everything else is // TODO: comment in the skeleton.

  9.5 Metadata drift

  EntityDefinition in Go code. DB schema in SQL migrations. SDUI schema in cache. Audit log assumes current field
  names. These four representations can drift:

  1. Field renamed in def.Entity → migration not written → column still old name → queries break silently
  2. Field removed from def.Entity → migration deletes column → SDUI cache still has old field → schema serves wrong
  UI
  3. Naming series format changed → existing records have old format → lookup by series fails

  Fix: Schema verifier at startup (section 5.5). Cache invalidation on registry change. Naming series format is
  immutable after first use (document this rule).

  ---
  Section 10 — Missing Capabilities

  Capability        Priority    Architecture fit
  ────────────────  ──────────  ─────────────────────────────────────────────────────────────
  Outbox/events     CRITICAL    awo/outbox — transactional outbox; natural fit with AfterCommit
  Schema verifier   CRITICAL    awo/verify — startup check; calls DB, part of app.Mount
  Optimistic lock   CRITICAL    Record.Version() + store conflict detection
  SearchProvider    HIGH        awo/search — interface; implementations in driver/meilisearch etc
  CacheProvider     HIGH        awo/cache — interface; redis impl in driver/redis
  EventBus          HIGH        awo/event — interface + in-process impl as default
  Scheduler/Jobs    HIGH        awo/jobs — interface; Temporal impl, cron impl
  AuthProvider      HIGH        awo/auth — interface; platform/iam is one implementation
  AuditLevel        HIGH        Richer than bool; on Entity struct
  HealthChecker     MEDIUM      awo/health — each capability registers a check
  Notifications     MEDIUM      awo/notify — Notifier interface; email, SMS, push implementations
  Feature Flags     MEDIUM      awo/flags — FlagProvider interface; in-memory + Redis impls
  Localization      MEDIUM      awo/i18n — MessageCatalog interface; labels/errors/UI strings
  Secrets           MEDIUM      awo/secrets — Provider interface; env, Vault, AWS SM impls
  Diagnostics       MEDIUM      HTTP endpoint + CLI; /debug/entities /debug/routes /debug/hooks
  API versioning    MEDIUM      Route prefix strategy; EntityDefinition.APIVersion int
  Read models       LOW         CQRS projection; needs EventBus first
  Rate limiting     LOW         RateLimiter interface; current redis impl as default
  Pessimistic lock  LOW         store.EntityStore.Lock(ctx, id, timeout) → unlock func
  Data lineage      LOW         Field-level provenance; regulated industries only
  CRDT fields       LOW         Field.CRDT semantic; needs offline-sync use case first

  ---
  Section 11 — Framework Comparison

  Framework       Architecture center    Multitenancy    Type safety    Extensibility   Metadata-driven
  ──────────────  ─────────────────────  ──────────────  ─────────────  ──────────────  ───────────────
  Awo             EntityDefinition       Native          Go (partial)   Plugin gaps     Central
  Frappe          DocType (JSON)         Bolted-on       None (Python)  Server Scripts  Central
  Odoo            ORM class inheritance  Bolted-on       None (Python)  Class inherit   Partial
  Django          Model + View + URL     None native     ORM only       Signal system   Weak
  Laravel         MVC + Eloquent ORM     None native     None (PHP)     Service prov.   Weak
  Spring Boot     Bean + Annotation      None native     Strong (Java)  Strong          Annotation-driven
  .NET Aspire     Service composition    None native     Strong (C#)    Strong          Infrastructure
  Kubernetes CP   CRD + Controller loop  Namespace       Strong (Go)    Operator pattern  Strong
  Temporal        Workflow + Activity    Namespace       Strong (Go)    Worker plugins  None
  Ent             Graph schema + codegen None native     Strong (Go)    Extension APIs  Schema-driven
  Atlas           Schema migration only  None            Strong (Go)    Provider model  Schema-driven
  Encore          Service + API decl.    None native     Strong (Go)    Limited         Annotation-driven
  Buffalo         MVC + codegen          None            Partial (Go)   Middleware       Weak
  Hasura          DB schema → GraphQL    None native     GraphQL types  Action hooks    Schema-driven

  Where Awo is stronger

  vs Frappe/Odoo: Type safety, performance, multi-tenancy from ground up, deterministic behavior. Frappe's DocType
  system is Awo's EntityDefinition but in JSON with no compile-time checks. Odoo's inheritance model creates deep
  class hierarchies that are hard to reason about. Awo's hook system is cleaner than either.

  vs Django/Laravel: Go performance, native multi-tenancy, metadata-driven API generation. Django and Laravel require
  hand-writing controllers, serializers, URL configurations for every entity. Awo derives these from the declaration.

  vs Spring Boot: Same "convention over configuration" philosophy but without annotation magic (Go doesn't have
  annotations). Awo's startup validation is cleaner. Spring's DI complexity is absent in Awo.

  vs Ent: Awo has multi-tenancy, SDUI, workflow integration, and full API generation. Ent has better type safety
  (generated typed queries), better edge traversal, and is production-proven at scale (Facebook, Entgo). Awo's filter
  DSL is weaker than Ent's generated type-safe predicates.

  vs Hasura: Business logic (hooks, policies, workflows) is a first-class citizen in Awo. Hasura's business logic is
  bolted on via "Actions" (external HTTP calls). Awo's multi-tenancy is architecturally integrated; Hasura's is a
  permission rule.

  Where Awo is weaker

  vs Frappe (ecosystem): Frappe has 15 years of production ERP entities. 500+ document types, accounting engine, stock
   accounting, HR, payroll, manufacturing — all battle-tested. Awo has none.

  vs Ent (type safety): Ent's code generation produces type-safe query builders.
  entity.Query().Where(entity.StatusEQ("submitted")) vs Awo's filter.Eq("status", "submitted"). Ent catches typos at
  compile time. Awo doesn't.

  vs Kubernetes (extensibility): K8s Operator pattern is the most mature extensibility model in Go. Admission
  webhooks, Custom Resources, controllers — this is what a production plugin ecosystem looks like. Awo has no
  comparable plugin model yet.

  vs Temporal (workflow): Temporal is specialized for durable execution; Awo treats workflow as a trigger. Complex
  workflow patterns (signals, queries, child workflows, cron) require dropping to the Temporal SDK directly. Awo's
  abstraction is thin.

  vs Atlas (migrations): Atlas generates migrations from schema declarations and verifies schema drift. Awo requires
  hand-written SQL migrations and has no drift detection.

  Fundamental differences

  Awo vs every framework above: Awo is the only Go framework where a single struct declaration drives persistence,
  REST, SDUI, permissions, audit, and workflow simultaneously. Every other framework requires separate files for each
  concern (model, view, controller, policy, serializer). This is Awo's unique value proposition and its most fragile
  assumption (when EntityDefinition becomes too complex, it becomes the bottleneck for everything).

  ---
  Section 12 — The Permanent Architecture

  Philosophy preserved

  One immutable entity declaration drives the entire framework.

  Changes that are mandatory before v1.0

  1. Dual policy interface (ListFilter + Eval)

  Without this, the framework cannot be used at scale. This is a breaking change to the Policy type. Do it now.

  2. Record.Version() int64 for optimistic locking

  Banking, inventory, any concurrent-write entity needs this. Adding to the Record interface is breaking. Do it now.

  3. Viewer.Roles() []string replacing HasRole(string) bool

  Performance and multi-org correctness. Breaking interface change. Do it now.

  4. WorkflowTrigger.Engine string and workflow.DriverRegistry

  Decouples kernel from Temporal. Breaking change to WorkflowTrigger. Do it now.

  5. FieldType two-tier (FieldKind + Subtype)

  Enables field type extensibility. Breaking to existing Field declarations. Do it now.

  6. StorageKind removed from kernel def.Entity

  Driver concern, not kernel concern. Breaking. Do it now.

  7. Event / EventDef as first-class primitive

  AfterCommit hooks are not enough. Do it now to avoid retrofitting later.

  8. ChildOf declaration on Entity

  Aggregate roots need this. Without it, frameworks generate wrong routes for sub-entities. Do it now.

  9. registry.Register moves out of def

  def must be pure types with no side effects. Breaking import path change. Do it now.

  10. Outbox pattern for AfterCommit reliability

  Financial systems require this. Architectural, not implementation detail.

  ---
  The permanent package layout

  awo/
  │
  ├── def/              STABLE FOREVER
  │   entity.go         Entity struct
  │   field.go          Field, FieldKind, FieldType
  │   record.go         Record, MutableRecord interfaces — GENERIC
  │   viewer.go         Viewer interface
  │   hook.go           Hook, Timing, constructor helpers
  │   policy.go         Policy, PolicyFunc, FilterPolicy, PolicyVerdict
  │   action.go         Action, ActionContext, ActionResult
  │   event.go          EventDef (NEW)
  │   edge.go           Edge, Cardinality, Cascade
  │   scope.go          Scope struct
  │   op.go             Op bitmask + constants
  │   errors.go         ValidationError, DomainError (types only)
  │
  ├── filter/           STABLE FOREVER
  │   filter.go         Filter node types, constructors
  │   eval.go           In-memory filter evaluation (for testing)
  │
  ├── registry/         STABLE WITH ADDITIONS
  │   registry.go       Register, All, Get, ByModule, Freeze, Extend, Retract
  │
  ├── store/            STABLE WITH ADDITIONS
  │   store.go          EntityStore, TenantStore interfaces, ListOptions, ListPage
  │   errors.go         ErrNotFound, ErrConflict, ErrVersionConflict (NEW)
  │   mem/              In-memory EntityStore — for testing ONLY
  │
  ├── access/           STABLE WITH ADDITIONS
  │   enforcer.go       Enforcer (handles both Eval + ListFilter collection)
  │
  ├── hook/             STABLE WITH ADDITIONS
  │   runner.go         Runner (serial + parallel scheduling by timing)
  │
  ├── event/            STABLE WITH ADDITIONS
  │   emitter.go        Emitter interface
  │   bus.go            In-process EventBus (default implementation)
  │
  ├── outbox/           NEW — CRITICAL
  │   outbox.go         OutboxWriter (writes to outbox in TX)
  │   worker.go         OutboxWorker (delivers from outbox)
  │
  ├── validate/         STABLE WITH ADDITIONS
  │   errors.go         ValidationError constructors (FieldError, FieldErrors)
  │   validators.go     Built-in field validators
  │
  ├── auth/             NEW — INTERFACE ONLY
  │   provider.go       AuthProvider interface, ViewerFromRequest type
  │
  ├── search/           NEW
  │   provider.go       SearchProvider interface
  │
  ├── cache/            NEW
  │   provider.go       CacheProvider interface
  │   mem/              In-memory cache (for testing)
  │
  ├── jobs/             NEW
  │   scheduler.go      Scheduler interface
  │
  ├── notify/           NEW
  │   notifier.go       Notifier interface
  │
  ├── flags/            NEW
  │   provider.go       FlagProvider interface
  │
  ├── health/           NEW
  │   health.go         Checker interface, registry
  │
  ├── verify/           NEW
  │   verify.go         VerifySchema() — startup schema drift detection
  │
  ├── api/              INTERNAL INTERFACE
  │   handler.go        Handler struct (HTTP framework-agnostic)
  │   routing.go        RouteSpec (metadata about generated routes)
  │   options.go        HandlerOption
  │
  ├── sdui/             INTERNAL
  │   builder.go        Builder
  │   handler.go        HTTP handler
  │   cache.go          Cache interface
  │   widget.go         WidgetRegistry (NEW — for custom field type widgets)
  │
  ├── workflow/         INTERNAL INTERFACE
  │   driver.go         Driver interface, DriverRegistry
  │   executor.go       Executor (dispatches to registered driver)
  │   trigger.go        TriggerConfig
  │   saga.go           Saga, ActivitySaga
  │
  ├── plugin/           NEW
  │   manifest.go       Manifest, Plugin interface
  │   registry.go       PluginRegistry (what plugins can touch)
  │
  ├── driver/           INTERNAL — NOT IMPORTABLE BY USER CODE
  │   pgstore/          pgx implementation of store.TenantStore
  │   redis/            CacheProvider + session store + rate limiter
  │   temporal/         workflow.Driver implementation
  │   amis/             SDUI renderer
  │   fiber/            HTTP adapter (fiber.Handler → api.Handler)
  │   nethttp/          HTTP adapter (net/http)
  │   meilisearch/      SearchProvider implementation (contrib)
  │
  ├── platform/         CONTRIB — separate module, not framework core
  │   iam/              User, Session, AuthService, AuthProvider impl
  │   tenant/           Tenant entity, StatusMiddleware
  │   org/              OrgUnit entity, Tree interface, pg impl
  │   naming/           NamingSeries counter
  │
  └── app/              COMPOSITION ROOT
      app.go            App, Options, Mount
      shutdown.go       Shutdowner collection + SIGTERM handling

  ---
  The permanent dependency graph

  stdlib only
      ↑
  ┌─────────────────────────────────────────────────────────────┐
  │  def        filter       (zero external dependencies)       │
  └─────────────────────────────────────────────────────────────┘
                ↑
  ┌─────────────────────────────────────────────────────────────┐
  │  registry   store        validate   event    auth           │
  │  (def)      (def,filter) (def)      (def)    (interface)    │
  └─────────────────────────────────────────────────────────────┘
                ↑
  ┌─────────────────────────────────────────────────────────────┐
  │  access     hook         outbox     search   cache   jobs   │
  │  (store,    (def,        (store,    (def)    (iface) (iface)│
  │   def)       validate)    event)                            │
  └─────────────────────────────────────────────────────────────┘
                ↑
  ┌─────────────────────────────────────────────────────────────┐
  │  driver/*   (implement interfaces; import external libs)    │
  │  pgstore → pgx                                              │
  │  redis   → go-redis                                         │
  │  temporal→ temporal-sdk                                     │
  │  fiber   → fiber/v2                                         │
  └─────────────────────────────────────────────────────────────┘
                ↑
  ┌─────────────────────────────────────────────────────────────┐
  │  api        sdui         workflow                           │
  │  (store,    (registry,   (driver/*)                         │
  │   access,    def,                                           │
  │   hook,      amis driver)                                   │
  │   filter)                                                   │
  └─────────────────────────────────────────────────────────────┘
                ↑
  ┌─────────────────────────────────────────────────────────────┐
  │  platform/* (contrib)                                       │
  │  iam, tenant, org, naming                                   │
  │  (same pattern as business modules — register entities)     │
  └─────────────────────────────────────────────────────────────┘
                ↑
  ┌─────────────────────────────────────────────────────────────┐
  │  app   (composition root — depends on everything)           │
  └─────────────────────────────────────────────────────────────┘
                ↑
  ┌─────────────────────────────────────────────────────────────┐
  │  [business modules]                                         │
  │  import: def, filter, registry, validate, event             │
  │  never import: api, sdui, driver/*, app                     │
  └─────────────────────────────────────────────────────────────┘

  Invariants that make this graph stable for decades:

  1. def and filter import nothing from the framework. This is verifiable with go list -deps awo/def. If this ever
  includes an external dependency, the architecture has been violated.
  2. driver/* packages are the only packages that import external libraries. This is the isolation boundary. All
  third-party churn is contained here.
  3. Business modules import only from the bottom two rows (def, filter, registry, validate, event). They never import
   api, sdui, app, or driver/*. This is enforceable with tooling (depguard or custom go vet check).
  4. The graph is a DAG. Any cycle = architecture violation. Verify with go mod why and CI enforcement.

  ---
  The ten-year architecture in one sentence

  def.Entity is immutable metadata; every framework capability is a pure function of that metadata; every external
  dependency is isolated in a driver; the kernel is a closed set of stable Go types with no external imports.

  ---
  Mandatory vs Deferred

  Mandatory before v1.0

  ┌──────────────────────────────────────┬────────────────────────────────────────────────────────┬───────────────┐
  │                Change                │                          Why                           │   Breaking?   │
  ├──────────────────────────────────────┼────────────────────────────────────────────────────────┼───────────────┤
  │ ListFilter in Policy                 │ Scalability — post-fetch filter breaks at 10K+ records │ Yes           │
  ├──────────────────────────────────────┼────────────────────────────────────────────────────────┼───────────────┤
  │ Record.Version()                     │ Banking/inventory correctness — optimistic locking     │ Yes           │
  ├──────────────────────────────────────┼────────────────────────────────────────────────────────┼───────────────┤
  │ Viewer.Roles() replacing HasRole     │ Performance + multi-org                                │ Yes           │
  ├──────────────────────────────────────┼────────────────────────────────────────────────────────┼───────────────┤
  │ WorkflowTrigger.Engine string        │ Kernel decoupled from Temporal                         │ Yes           │
  ├──────────────────────────────────────┼────────────────────────────────────────────────────────┼───────────────┤
  │ FieldKind + FieldType.Subtype        │ Extensible field types                                 │ Yes           │
  ├──────────────────────────────────────┼────────────────────────────────────────────────────────┼───────────────┤
  │ Remove StorageKind from kernel       │ Driver concern in kernel                               │ Yes           │
  ├──────────────────────────────────────┼────────────────────────────────────────────────────────┼───────────────┤
  │ EventDef on Entity                   │ Event-driven architecture foundation                   │ Yes           │
  ├──────────────────────────────────────┼────────────────────────────────────────────────────────┼───────────────┤
  │ Entity.ChildOf aggregate declaration │ Correct route generation                               │ Yes           │
  ├──────────────────────────────────────┼────────────────────────────────────────────────────────┼───────────────┤
  │ registry.Register not in def         │ Pure kernel, no side effects                           │ Yes           │
  ├──────────────────────────────────────┼────────────────────────────────────────────────────────┼───────────────┤
  │ Outbox pattern                       │ Reliability for financial systems                      │ No (additive) │
  ├──────────────────────────────────────┼────────────────────────────────────────────────────────┼───────────────┤
  │ Schema drift detection at startup    │ Silent corruption prevention                           │ No (additive) │
  ├──────────────────────────────────────┼────────────────────────────────────────────────────────┼───────────────┤
  │ AuthProvider interface               │ Platform/IAM not hard-wired                            │ No (additive) │
  ├──────────────────────────────────────┼────────────────────────────────────────────────────────┼───────────────┤
  │ Hook priority field                  │ Plugin ordering                                        │ No (additive) │
  ├──────────────────────────────────────┼────────────────────────────────────────────────────────┼───────────────┤
  │ BeforeList hook timing               │ Query filtering, audit                                 │ No (additive) │
  ├──────────────────────────────────────┼────────────────────────────────────────────────────────┼───────────────┤
  │ Panic recovery in Enforcer           │ Runtime stability                                      │ No (additive) │
  ├──────────────────────────────────────┼────────────────────────────────────────────────────────┼───────────────┤
  │ bindWorkflowTriggers before Freeze   │ Race condition fix                                     │ No (fix)      │
  └──────────────────────────────────────┴────────────────────────────────────────────────────────┴───────────────┘

  Deferred to v1.x

  ┌─────────────────────────────────────┬───────────────────────────────────────────┐
  │               Change                │             Trigger condition             │
  ├─────────────────────────────────────┼───────────────────────────────────────────┤
  │ Two-phase registry (Extend/Retract) │ When tenant-defined entities needed       │
  ├─────────────────────────────────────┼───────────────────────────────────────────┤
  │ WASM plugin sandboxing              │ When untrusted third-party plugins needed │
  ├─────────────────────────────────────┼───────────────────────────────────────────┤
  │ CRDT field semantics                │ When offline-sync use case appears        │
  ├─────────────────────────────────────┼───────────────────────────────────────────┤
  │ Multi-region pool routing           │ When GDPR/HIPAA data residency required   │
  ├─────────────────────────────────────┼───────────────────────────────────────────┤
  │ Read model / CQRS support           │ After EventBus is stable                  │
  ├─────────────────────────────────────┼───────────────────────────────────────────┤
  │ GraphQL driver                      │ After ListFilter policies exist           │
  ├─────────────────────────────────────┼───────────────────────────────────────────┤
  │ Full localization                   │ When non-English deployment needed        │
  ├─────────────────────────────────────┼───────────────────────────────────────────┤
  │ Pessimistic locking                 │ When specific workflow requires it        │
  └─────────────────────────────────────┴───────────────────────────────────────────┘
```
