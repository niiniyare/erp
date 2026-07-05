# Awo Framework — Final Governance Review
## Chief Architect Sign-Off Document | Pre-v1.0 Freeze | 2025-07-06

**Classification:** Architecture Decision Record (ADR-FINAL)
**Status:** UNDER REVIEW — Not yet approved
**Scope:** Full architectural governance, 2025–2045 lifecycle
**Reviewers:** Chief Software Architect, Platform Architecture Board
**Precedent frameworks:** PostgreSQL, Kubernetes, Linux Kernel, Git, Temporal, Django, Spring, Go runtime, Frappe

---

## Preface

This document is the final architectural governance review before Awo Framework v1.0 is permanently frozen. It assumes every recommendation from prior reviews has been implemented exactly as specified. It does not repeat those recommendations. It does not discuss implementation quality, coding style, or naming conventions unless they create structural risk.

Every finding here asks one question: **Will this decision still be correct in 2045?**

Millions of lines of business software will depend on the decisions made in this document. The cost of being wrong is not a refactor. It is a forced v2 — which, in practice, means a decade of parallel maintenance, ecosystem fragmentation, and migration debt. PostgreSQL has never broken its wire protocol since 1996. Git has never changed its object model since 2005. Linux has never broken the kernel ABI since 1994. These are not accidents. They are the result of deliberate governance decisions made before the ecosystem locked in.

Awo must make those decisions now.

---

## Section 1 — Architectural Philosophy

### 1.1 Does a Coherent Philosophy Exist?

Awo's philosophy can be extracted from its structure, even if it has never been written down explicitly. That is the first problem. The philosophy is:

> **Tenancy is structural. Metadata drives behavior. Drivers isolate infrastructure. Everything else is composition.**

This is a sound philosophy. It is consistent with how successful long-lived frameworks are designed. It is not a marketing tagline — it is an engineering constraint that eliminates entire categories of decisions downstream. When tenancy is structural, no feature can accidentally skip it. When metadata drives behavior, behavior is auditable without reading code. When drivers isolate infrastructure, deployment models are orthogonal to application logic.

The problem is that this philosophy has never been formally articulated, documented, or made binding. It exists only as institutional memory.

**Why this matters over twenty years:** In year five, a new contributor adds a feature that bypasses the metadata system because it is "just a small thing." In year eight, a driver implementation leaks implementation details because nobody documented what "driver isolation" means precisely. In year fifteen, tenancy is broken for a new deployment model because the philosophy was never formally tested against new requirements. None of these failures are malicious. They are the natural result of a philosophy that was never committed to writing with binding force.

**Recommendation (must happen before v1.0):** The philosophy must be committed to a formal architectural constitution — not documentation, but a binding constraint document that is checked against every new API proposal. It must specify:

1. Tenancy is a zero-bypass invariant. No API call can succeed without tenant context unless it is explicitly classified as a platform-admin operation.
2. Metadata is the single source of truth. No behavior can be defined in code that cannot be expressed in metadata.
3. Drivers are interfaces, not implementations. No core package imports a driver package. Ever.
4. Composition over configuration. Feature combinations are expressed through module composition, not flags.
5. Compile-time over runtime. Every invariant that can be checked at compile time must be.

These five axioms, if violated, are automatic v1.0 blockers. Not suggestions. Blockers.

### 1.2 Subsystem Alignment

Most subsystems align with the philosophy. The metadata-driven CRUD, permission engine, and SDUI generation are coherent. However three subsystems show tension:

**Tension 1 — The Bootstrap Package.** `framework/bootstrap` is a convenience layer that knows about specific drivers (IAM, Redis, Fiber). This violates "drivers are interfaces, not implementations." The bootstrap package is already architecture-aware in a way that core packages cannot be. Over twenty years, this package will accumulate every "we just need to wire this together" shortcut. It will become the framework's junk drawer. It must be formalized: bootstrap is an opinionated `cmd/`-layer responsibility, not a framework package. The framework provides `awo.Builder`. The application provides bootstrap configuration.

**Tension 2 — The Registry-as-God-Object.** The Registry knows everything: all entities, all fields, all hooks, all policies, all routes, all workflows. This is correct for compilation. It is dangerous for runtime. A mutable registry that can be queried at runtime becomes a global state machine that every subsystem can reach into. The question is not whether the registry exists — it must. The question is whether the compiled, frozen schema (read-only) and the mutable registry (write-only before compile) are the same object or two separate objects. They must be two separate objects. The registry mutates during initialization. The compiled schema never mutates after `Compile()`. If they share a type, the distinction is invisible and will be violated.

**Tension 3 — SDUI Coupling.** The philosophy says metadata drives behavior. But SDUI is not just driven by metadata — it also depends on amis, a specific third-party renderer. The philosophy should say: metadata drives the abstract UI model, and renderers are drivers. Today the abstract UI model and the amis schema are the same thing. In 2035, when amis is abandoned or superseded, every `PageBuilder` in every third-party module is broken simultaneously. SDUI needs a renderer driver interface before v1.0.

### 1.3 Philosophy Survival to 2045

The five axioms above will survive to 2045 without modification because they are constraints on the framework's relationship with applications, not constraints on technology choices. Technology will change. The constraint that drivers are interfaces does not change when Postgres is replaced by something else.

The one axiom that may require extension is "metadata is the single source of truth." In 2035, metadata may need to express AI agent behavior, streaming queries, CRDT merge functions. The metadata language must have extension points that allow new semantic categories to be added without breaking existing metadata. This is a compiler architecture concern, addressed in Section 3.

---

## Section 2 — Core Abstractions

Every abstraction in a framework is a permanent tax on every user. Unnecessary abstractions accumulate cognitive debt. The wrong abstraction is worse than no abstraction — it encodes a wrong mental model into millions of lines of code.

### 2.1 EntityDefinition — KEEP, with caveats

EntityDefinition is the correct central abstraction. It represents a real domain concept: a description of a kind of thing that the system manages. It is not an implementation artifact. It would exist even if this framework did not.

**Caveat:** EntityDefinition today is both a specification and a configuration bag. It mixes what a thing is (fields, edges, identity) with how it behaves in this specific deployment (hooks, workflows, policies). These are different concerns with different stability requirements. Fields change rarely after v1.0. Hooks change every sprint.

The permanent, frozen part of EntityDefinition must be separated from the behavioral extension part. Concrete proposal:

```go
// Stable — frozen at v1.0. Changes require RFC + major version bump.
type EntitySpec struct {
    Name    string
    Module  string
    Fields  []FieldDef
    Edges   []EdgeDef
    Schema  int
}

// Behavioral — can evolve without breaking EntitySpec.
type EntityBehavior struct {
    Hooks       HookSet
    Policies    []PolicyFunc
    Actions     []ActionDef
    Workflows   []WorkflowTrigger
    Pages       PageBuilderSet
}
```

Both compose into EntityDefinition, but the split communicates which part carries a compatibility promise and which part is extension territory. This distinction matters when module authors ask: "Can I depend on another module's entity spec?" Yes. "Can I depend on another module's hooks?" No — those are internal.

### 2.2 FieldDef — KEEP, refactor cardinality

FieldDef is correct. It represents a real concept: a description of a property of an entity.

**Problem:** FieldDef carries too many concerns simultaneously. It describes the data model (type, constraints), the persistence model (column name, index type), the UI model (label, help text, hidden), and the validation model (required, validators). These concerns evolve at different rates and for different reasons.

Over twenty years, this means FieldDef accumulates fields from every team that ever needed to attach metadata to a field. By 2035 a FieldDef will have 60 fields, most of which are irrelevant for any given field type. This is how Spring's annotation model collapsed under its own weight.

**Solution:** FieldDef must support extension through a typed extension map, not through field proliferation:

```go
type FieldDef struct {
    Name        string
    Type        FieldType
    Constraints FieldConstraints  // data model: required, unique, immutable
    // Extensions — each subsystem adds its own
    UI          *UIFieldSpec      // nil = use defaults
    Search      *SearchFieldSpec  // nil = not indexed
    Encryption  *EncryptionSpec   // nil = not encrypted
    // Open extension point — never iterate this in core
    Extensions  map[string]any
}
```

Each subsystem reads only its own extension. Core never iterates `Extensions`. This prevents FieldDef from becoming a god struct.

### 2.3 Record — KEEP, but interface is too thin

The Record abstraction is correct. A Record is a real domain concept: an instance of an entity at a point in time.

**Problem:** The current Record interface is too thin in one dimension and too rich in another.

Too thin: Record carries no versioning information. In year seven, when you add schema migration (SchemaVersion per record), you will discover that every interface that accepts Record needs to be updated or wrapped. This should have been in the interface from the start.

Too rich: MutableRecord exposes mutation methods directly. This means any code that receives a MutableRecord can mutate any field. There is no field-level write access control at the type level. In a hook, you want to be able to say "this hook can only write to the `status` field." Today, all hooks get full write access to all fields. This cannot be enforced at compile time.

**Required before v1.0:** Record must include `SchemaVersion() int`. MutableRecord must include `Set(field string, value any) error` where the error carries field-level permission information from the execution context. The mutation-permission model does not need to be complete at v1.0, but the interface must have the hook for it.

### 2.4 ViewerContext — KEEP, but freeze carefully

ViewerContext is correct. It represents who is asking, not implementation details of how that was determined.

**Risk:** The interface will be pressure-tested immediately. Every module author will want to add methods: `Locale()`, `TimeZone()`, `BranchID()`, `OrgPath()`, `Capabilities()`. If the interface is frozen with five methods at v1.0, optional interface extension (type assertion pattern) prevents breaking changes. But if it is frozen with fifteen methods, implementations become burdensome, especially for testing.

**Ruling:** Freeze ViewerContext with exactly the minimum: `ActorID`, `TenantID`, `Roles`, `IsSystem`, `HasRole`. Everything else is optional interfaces. Document this decision explicitly in the ADR so future contributors do not add to the core interface.

### 2.5 Policy — KEEP, but add output type

Policy today returns a Filter — a predicate that narrows what the viewer can see. This is correct for row-level filtering.

**Missing:** Policy has no mechanism to explain its decision. In a regulated industry (banking, healthcare), every access control decision must be auditable: not just "was this allowed" but "why was this allowed." A Policy that returns only a Filter cannot answer "why."

**Required before v1.0:** PolicyResult must carry both the Filter and an optional audit descriptor:

```go
type PolicyResult struct {
    Filter    Filter
    Rationale *PolicyRationale  // nil = not required for audit
}

type PolicyRationale struct {
    PolicyID  string
    Matched   []string  // which rules matched
    Variables map[string]any
}
```

This allows the audit log to record the policy evaluation that permitted each data access. Without this, you cannot achieve ISO 27001 certification in regulated deployments.

### 2.6 Hook — KEEP, but lifecycle is incomplete

HookSet covers BeforeCreate, BeforeUpdate, BeforeDelete, AfterCreate, AfterUpdate, AfterDelete, AfterSave. This covers CRUD.

**Missing lifecycle events:**
- `OnRead` — needed for field-level decryption, access logging of sensitive data, rate-limited read tracking
- `AfterQuery` — needed for collection-level post-processing (redacting fields based on aggregate context)
- `OnValidate` — currently Before* hooks do validation + business logic. These should be separate stages so validation failures produce different error types than business logic failures.
- `AfterCommit` — the outbox hook fires after the DB transaction commits. But "after the outbox is flushed" is a different event. Modules need to distinguish these.

**Risk if not fixed before v1.0:** Third-party modules will implement workarounds (polling, side tables, middleware hacks) for missing lifecycle events. These workarounds will become load-bearing. Removing them in v2 will require breaking third-party modules.

### 2.7 Event — KEEP, but CloudEvents format is an invariant

Event is correct. But the wire format of events emitted by Awo is a public API the moment any external system consumes it. CloudEvents v1.0 must be the wire format at v1.0. Not "probably" — definitively. If Awo emits custom event envelopes at v1.0, every integration built before the switch to CloudEvents breaks when the switch happens.

### 2.8 Action — KEEP

Action is clean. It represents a domain operation that is not CRUD. It is a good abstraction.

**One concern:** Action today is defined on EntityDefinition, which makes it entity-scoped. Some business actions are cross-entity or tenant-scoped: "close the accounting period," "run payroll for all employees." These have no natural home in the current model. Before v1.0, the framework needs either a `TenantAction` concept or a way to register entity-independent actions. Otherwise, every module author that needs a cross-entity action will shoehorn it into an entity, creating semantic pollution.

### 2.9 CompiledSchema — KEEP, but it is under-specified

CompiledSchema is the framework's most important type. It is the immutable output of the compiler — the frozen description of everything the runtime needs to operate.

**Under-specification risks:**
1. CompiledSchema has no version or content hash. When it is cached, how do you know it is current? When it is serialized to disk (for faster startup), how do you know it matches the current source?
2. CompiledSchema carries no forward-compatibility markers. When v1.1 adds new compiled artifacts, does v1.0-compiled schema fail fast or silently produce wrong behavior?
3. CompiledSchema is not serializable today. For large deployments, startup time is dominated by compilation. If CompiledSchema cannot be serialized to disk and restored, you recompile on every restart. At 10,000 entity definitions, this is measurable startup latency.

**Required before v1.0:** CompiledSchema must have: content hash (SHA-256 of all inputs), compiler version stamp, and a serialization interface (even if the serialization format is marked experimental).

### 2.10 Registry — KEEP, split into two phases

Registry is correct but dangerous as a single object. See Section 1.2. The write-phase registry and the read-phase compiled schema must be distinct types. This is not a rename — it is a type separation that prevents read-before-compile bugs at compile time.

### 2.11 Runtime — KEEP, but lifecycle must be explicit

Runtime is the operational object — the thing that processes requests, dispatches hooks, evaluates policies. It is correct.

**Risk:** Runtime's lifecycle (start, ready, degraded, shutdown) is not formalized. At 100,000 concurrent users, graceful shutdown is a significant operational concern. Runtime must have an explicit lifecycle interface:

```go
type Lifecycle interface {
    Ready() <-chan struct{}
    Shutdown(ctx context.Context) error
    Health() HealthStatus
}
```

Without this, every operator who deploys Awo in Kubernetes will implement their own lifecycle management, incompatibly.

### 2.12 Driver — KEEP, but conformance is mandatory

Driver interfaces are correct. The risk is that "Driver" is a category, not a single interface. There are storage drivers, auth drivers, workflow drivers, cache drivers, search drivers, renderer drivers. Each has different conformance requirements.

Before v1.0, each driver category must have a published conformance test suite. A driver that passes the conformance suite is certified. A driver that does not pass is experimental. This distinction must be encoded in the type system, not just documentation:

```go
type CertifiedEntityStore interface {
    driver.EntityStore
    PassesConformance() bool  // returns true only if tested
}
```

This is how PostgreSQL handles extensions, how Kubernetes handles CSI/CNI/CRI drivers. Without it, "compatible driver" means "compiles," which is a far weaker guarantee.

### 2.13 Builder — RECONSIDER

Builder is the fluent API for constructing the runtime. It is convenient but carries risk.

**The risk:** Builder is the framework's entry point. It is the first thing every user sees. Its API communicates the framework's philosophy more than any documentation. If Builder is designed wrong, the philosophy is communicated wrong.

Current risk: Builder accepts optional components (workflow, cache, search). This means the framework can be started in degraded mode silently. In production, a misconfigured Awo that skips workflow or cache is dangerous. Builder should have a concept of required and optional drivers, and fail fast if required drivers are absent.

Also: Builder must have a validation phase separate from construction. `Build()` must not just wire things together — it must run pre-flight checks on the compiled schema against the provided drivers (schema compatibility, migration status, driver health). If this validation is inside `Build()`, it is invisible. It should be explicit: `builder.Validate()` returns a diagnostic report. `builder.Build()` refuses to run if validation failed.

---

## Section 3 — Compiler Architecture

### 3.1 The Compiler as a First-Class Concept

The metadata compiler is Awo's most powerful subsystem and its most under-theorized. It is currently implemented as an initialization step — "run this code before anything else." It needs to be understood, documented, and governed as a compiler.

A mature framework compiler has these phases, in order:

**Phase 1 — Parsing / Registration**
Input: `EntityDefinition` structs from `init()` calls.
Output: Raw entity graph — unresolved references, no validation.
Invariant: Deterministic — same input, same raw graph.

**Phase 2 — Name Resolution**
Input: Raw entity graph.
Output: Resolved graph — all `LinkTarget: "customer"` references resolve to actual entity pointers.
Errors: Undefined reference (fatal), circular reference (fatal for strict edges, warning for soft).

**Phase 3 — Semantic Analysis**
Input: Resolved graph.
Output: Annotated graph — field types validated, edge multiplicities consistent, action permissions referencing real roles.
Errors: Type mismatch (fatal), missing required fields (fatal), permission referencing nonexistent role (warning or fatal, configurable).

**Phase 4 — Policy Compilation**
Input: Annotated graph + Casbin policy rules.
Output: Compiled policy store — Casbin enforcer initialized, custom PolicyFuncs wrapped and indexed.
Errors: Policy syntax error (fatal), conflicting policies (warning).

**Phase 5 — Route Table Compilation**
Input: Annotated graph.
Output: Route table — URL patterns → handler descriptors. No handler logic, only routing metadata.
Invariant: Deterministic. Given same entity graph, same routes. Must be reproducible for documentation generation.

**Phase 6 — Hook Chain Compilation**
Input: Annotated graph.
Output: Per-entity hook chains — ordered slices of hook descriptors. Execution order is declared order. No runtime sorting.
Invariant: Deterministic. Hook chains frozen at compile time.

**Phase 7 — SQL Template Compilation**
Input: Annotated graph.
Output: Per-entity SQL templates — parameterized query templates for CRUD, filter translation, aggregate queries.
Note: SQL templates are not queries. They are templates. Tenant context and filter parameters are injected at runtime.

**Phase 8 — SDUI Tree Compilation**
Input: Annotated graph + PageBuilderSets.
Output: Per-entity SDUI trees — abstract UI descriptions, renderer-independent.
Note: Renderer-specific serialization (amis JSON) happens at request time, not compile time.

**Phase 9 — Workflow Graph Compilation**
Input: Annotated graph + WorkflowTriggers.
Output: Per-entity workflow trigger map — event → workflow descriptor.
Validation: WorkflowFn references must be registered in the workflow driver's registry.

**Phase 10 — Schema Seal**
Input: All compiled artifacts.
Output: CompiledSchema — immutable, content-addressed (SHA-256), timestamped.
Invariant: After seal, no mutation possible. All reads are safe for concurrent access.

### 3.2 Intermediate Representation

The Awo compiler currently has no IR. Compilation is a direct transformation from EntityDefinition to multiple output formats simultaneously. This works at small scale. It breaks in three ways as the system grows:

**Problem 1 — Optimization Passes.** Without IR, you cannot run optimization passes. An optimization pass might detect that two entities share 80% of their field definitions and generate shared SQL templates. Without IR, you cannot express this.

**Problem 2 — Plugin Compiler Hooks.** Without IR, a plugin that wants to add a compilation phase must modify the compiler directly. This violates modularity. With IR, a plugin adds a pass that reads and optionally transforms the IR.

**Problem 3 — Diagnostics.** Without IR, error messages are produced by the compilation code that discovers the error. With IR, a separate diagnostic pass analyzes the fully-resolved graph and produces structured, actionable error messages with source locations.

**Required before v1.0:** Define an IR. It does not need to be complete. It needs to be the canonical intermediate form between the raw entity graph (Phase 1) and the compiled artifacts (Phases 5–9). All optimization and diagnostic passes operate on IR, not on EntityDefinition directly.

Minimum IR nodes:
- `IREntity` — resolved entity with all fields, edges, hooks, policies annotated
- `IRField` — resolved field with type metadata, index requirements, constraint predicates
- `IREdge` — resolved edge with both endpoint IREntity references, cardinality, cascade behavior
- `IRHookChain` — ordered list of hook descriptors with their phase and error handling mode
- `IRPolicyChain` — ordered list of policy descriptors with their evaluation mode (first-match vs. all)
- `IRRouteDescriptor` — URL pattern, method, handler type, permission requirements
- `IRWorkflowBinding` — event, workflow function reference, input builder reference

This IR is the framework's internal language. It is not a public API. It can evolve freely. But it must exist, or the compiler cannot grow.

### 3.3 Incremental Compilation

At 10,000 entity definitions, full recompilation on every startup is expensive. At 100,000 entity definitions (a large multi-tenant platform in 2035), it is prohibitive.

Incremental compilation requires:
1. Content-addressable cache keyed by the hash of each EntityDefinition's inputs
2. Dependency graph between compiled artifacts (if entity A references entity B's compiled output, changing B invalidates A's cache)
3. Sealed compilation artifacts that can be serialized and restored without re-running compilation phases

This is exactly how Bazel, Gradle, and LLVM handle incremental compilation. It is not exotic — it is standard compiler engineering.

The dependency graph from Phase 2 (Name Resolution) is the natural foundation for incremental compilation. If it exists for correctness, it is already half the work for incrementalism.

**Required before v1.0:** Content-addressed compilation cache with dependency tracking. Does not need to be on disk — in-memory cache with hash-keyed entries is sufficient for v1.0. Disk persistence is v1.1.

### 3.4 Compiler Plugin Hooks

Every successful framework compiler eventually needs extensibility. LLVM has pass plugins. Python has AST transformers. Go has `go/analysis` analyzers. Awo will need them too.

The canonical use case: a regulatory compliance module needs to verify that every entity definition with `financial_data: true` has field-level encryption enabled. This is a semantic validation — it cannot be expressed as a FieldDef constraint. It requires a compiler pass that has visibility into the full entity graph.

**Compiler plugin interface (minimum viable):**
```go
type CompilerPass interface {
    Name() string
    Phase() CompilerPhase  // which phase to run after
    Run(ir IR, diag Diagnostics) error
}
```

Compiler passes are registered at init time alongside entity definitions. They run in phase order, after the built-in phase they declare. They can emit diagnostics (errors or warnings) but cannot modify the IR unless they are explicitly classified as transform passes.

**Required before v1.0:** CompilerPass interface must exist and be stable. The built-in passes must use it internally (this proves the interface is sufficient). External registration must work. Three built-in passes must be shipped as examples: schema validator, permission completeness checker, migration signature verifier.

### 3.5 Compile-Time Guarantees

The compiler must make guarantees that are currently made at runtime. Specifically:

| Current runtime check | Should be compile-time |
|---|---|
| All `LinkTarget` references resolve to registered entities | Phase 2, Name Resolution |
| All `WorkflowFn` references have registered implementations | Phase 9, Workflow Graph |
| All permission role references exist in the policy store | Phase 4, Policy Compilation |
| No two entities have the same name | Phase 1, Registration |
| NamingSeries format strings are syntactically valid | Phase 3, Semantic Analysis |
| Action handler signatures match the expected interface | Phase 3, Semantic Analysis |

Every runtime panic that could have been a compile-time diagnostic is a failure mode that manifests in production, not in the development cycle. This is not an optimization — it is a correctness guarantee.

---

## Section 4 — Module Ecosystem

### 4.1 The Ecosystem Problem Is Not Technical

The hardest problem in framework ecosystems is not dependency resolution. It is incentive alignment. Django's package ecosystem is successful not because of pip, but because Django's API stability guarantees made it economically rational to maintain packages. When you broke compatibility in Django 2→3, hundreds of packages were abandoned. The ecosystem did not recover for years.

Awo's module ecosystem governance must make it economically rational for third-party developers to maintain modules over the framework's lifetime. This requires:

1. **Stability guarantees with precision.** Not "we try to be stable." Precisely: "def/, filter/, and driver/ package interfaces will not change in a backward-incompatible way within a major version. Behavioral changes require a 90-day deprecation notice. Breaking changes require a new major version and a migration guide."

2. **LTS designation.** Some framework versions receive 4-year security support. Module authors targeting LTS versions have a stable foundation for multi-year investment.

3. **Certification.** Modules that pass the framework's conformance suite are "Awo Certified." This is a quality signal that matters in B2B markets. Banking customers do not install uncertified modules.

Without these three things, the ecosystem will consist of hobby modules and internal modules. Enterprise customers will not depend on third-party modules. The platform moat never forms.

### 4.2 Module Manifests

Module manifests must express:

```go
type ModuleManifest struct {
    // Identity
    Name    string   // globally unique, reverse-DNS recommended: "io.example.crm"
    Version semver   // semantic version: 1.2.3

    // Framework compatibility
    RequiresAwo   VersionConstraint  // ">=1.0.0, <2.0.0"
    RequiresDef   VersionConstraint  // def/ package version

    // Capability contracts (not entity names)
    Provides  []CapabilityToken  // "crm.contact_crud", "crm.pipeline_management"
    Requires  []CapabilityToken  // "platform.tenant", "platform.iam"
    Conflicts []CapabilityToken  // "legacy.contacts" — cannot coexist

    // Optional capabilities
    OptionalRequires []CapabilityToken  // enhances if present, works without

    // Trust
    Signature     string  // Ed25519 signature over manifest content hash
    Publisher     string  // verified publisher identity
    AuditRequired bool    // regulated deployments: true = requires audit log
}
```

**Capability tokens, not entity names.** This is critical. If dependency resolution is based on entity names ("requires the `customer` entity"), then two modules that both define a `customer` entity create an unresolvable conflict. If it is based on capability tokens ("requires `crm.customer_management`"), then the resolver can negotiate: "module A provides `crm.customer_management`, module B requires it — satisfied." The entity name is an implementation detail. The capability is the contract.

### 4.3 Dependency Resolution Algorithm

Awo needs a dependency resolver that runs at startup (or build time) before compilation begins. It must:

1. Build a dependency graph from all registered module manifests
2. Detect cycles (fatal)
3. Detect conflicts (fatal: two modules declare incompatible capabilities)
4. Detect missing required capabilities (fatal)
5. Resolve optional dependencies (best-effort, degraded mode)
6. Produce a deterministic module load order (topological sort)

The resolver's output is the compilation order for Phase 1. Modules compile in dependency order. If module B requires module A's entities, A must be compiled before B so name resolution succeeds.

**The resolver must be hermetic.** Given the same set of module manifests, it always produces the same load order. This is testable and reproducible. Non-hermetic resolvers (where the result depends on registration order or random factors) produce "works on my machine" bugs that take years to diagnose.

### 4.4 Semantic Versioning Discipline

Awo must specify precisely what constitutes a breaking change:

**Breaking (requires major version bump):**
- Removing an interface method from `def/`, `filter/`, or `driver/`
- Adding a required parameter to any public function
- Changing the signature of any Hook interface method
- Changing the wire format of any event, record, or filter
- Changing the serialization format of CompiledSchema
- Removing any field from Record or ViewerContext
- Changing the URL structure of any auto-generated route

**Non-breaking (allowed in minor versions):**
- Adding a new optional interface
- Adding a new field to an extendable struct (if the struct uses Options pattern or extension map)
- Adding a new compiler phase (if existing phases are unaffected)
- Adding a new built-in FieldType
- Adding a new CompilerPass (additive only)

**Patch-only:**
- Bug fixes that do not change observable behavior
- Performance improvements that do not change API signatures
- Documentation corrections

This taxonomy must be formally published and enforced by automated tooling (apidiff or equivalent) in CI before every release.

### 4.5 Hot Installation

Hot installation (adding a module to a running deployment without restart) is architecturally impossible with Awo's compilation-unit model. This is a correct and intentional decision for regulated industries. However, the framework must make this limitation explicit and provide a safe restart-based installation path:

1. New module is deployed alongside existing binary (canary or blue-green)
2. New binary starts, compiles new schema, runs migration validation
3. If validation passes, old binary drains and shuts down
4. New binary takes traffic

The framework must provide tooling for this sequence. If it does not, operators will implement unsafe hot-patching workarounds. The safe path must be the easy path.

### 4.6 Module Signing and Trust

Supply chain security is an existential risk for module ecosystems. npm's ecosystem has experienced catastrophic supply-chain attacks. Go's module proxy and checksum database are the most mature model for preventing them.

Awo must adopt the Go model:
- Modules are identified by reverse-DNS names with version tags
- A public module index records the SHA-256 of each module version's content
- Awo startup verifies the manifest signature against the publisher's Ed25519 public key
- `AuditRequired: true` modules require additional verification: the full module source must be reproducibly buildable from the published source, verifiable by the deploying organization's security team

In air-gapped deployments, the signing verification chain must be configurable to use a local key store rather than the public module index. This must work without internet access.

---

## Section 5 — Long-Term Evolution

### 5.1 What v8 (2042) Needs From v1.0

The most dangerous long-term architecture mistake is optimizing for today's use cases at the cost of adaptability. The following technologies will be mainstream by 2035 and expected in enterprise frameworks by 2042. The question is not "does v1.0 support them" — it is "does v1.0 architecture allow them to be added without rewriting the kernel."

**AI Agents (certainty: very high)**
AI agents will need to: query entities using natural language, execute actions with human-in-the-loop approval, generate entity definitions from requirements, and autonomously run multi-step workflows.

What v1.0 needs for this:
- Filter DSL must be expressible as a data structure (not only as Go code), so it can be generated by LLMs
- Actions must have machine-readable descriptions (purpose, inputs, outputs, effects) — today they have labels for humans only
- Workflow triggers must support `AgentSignal` event type alongside `OnCreate`, `OnSubmit`
- ViewerContext must support `AgentIdentity` as a principal type

None of these require AI in v1.0. They require that the abstractions are not hard-coded to assume human users.

**Offline-First / Mobile Sync (certainty: high for some verticals)**
Field operations (agriculture, construction, field services) require offline data entry with eventual consistency sync. This requires:
- Record identity must be client-generated (UUID v7 — already correct)
- Conflict resolution must be expressible as metadata (`ConflictStrategy: LastWriteWins | MergeFunction`)
- The Filter DSL must be evaluatable client-side (not just server-side SQL) — this requires a portable Filter DSL evaluator
- Hooks must declare whether they are required-synchronous or can-be-deferred-until-online

These are additions, not redesigns. But they cannot be added cleanly if v1.0 assumes every operation is online, synchronous, and server-side.

**Stream Processing (certainty: high)**
ERP systems will need to process high-volume event streams: IoT sensor data, payment events, telemetry, inventory movements. SQL-over-JSONB is not the right model for stream processing.

What v1.0 needs:
- EntityDefinition must support a `source` concept: is this entity sourced from a table, a stream, an external API, or a materialized view?
- The driver interface must include `StreamDriver` alongside `EntityStore`
- The outbox pattern (which already exists) is the bridge between transactional writes and stream processing

**WebAssembly Plugins (certainty: medium-high by 2035)**
WASM is the emerging universal plugin sandboxing mechanism. Cloudflare Workers, Envoy filters, and database UDFs are moving toward WASM. Awo's compilation-unit model (no `plugin` package) is correct today. But by 2035, WASM may be the right isolation mechanism for untrusted third-party hooks.

What v1.0 needs: Hook interfaces must be expressible as pure data-in/data-out contracts (serializable inputs and outputs). If hooks are defined as Go function references today and WASM function references tomorrow, the interface must not change — only the execution mechanism changes. This means hook signatures must use serializable types, not Go-specific types like `error` with concrete subtypes.

**GraphQL and gRPC (certainty: very high)**
REST is not the only API paradigm. By 2035, GraphQL will be standard for data-intensive frontends and gRPC will be standard for service-to-service calls.

What v1.0 needs: The route table compilation (Phase 5) must produce an abstract API description, not REST-specific routes. REST is then a renderer of that abstract description. GraphQL and gRPC are alternative renderers. This requires separating "what can be done with this entity" (the abstract capability) from "how it is expressed in HTTP REST" (the transport implementation).

Today, REST routes are generated directly from EntityDefinition. This couples the API model to the transport. Uncoupling them is a compiler architecture change that is very difficult to retrofit after v1.0 because third-party modules will have built assumptions about the REST route structure.

**This is a v1.0 blocker.** The compiler must produce an abstract API descriptor. The REST transport is a driver of that descriptor.

### 5.2 New Database Technologies

Vector databases (pgvector is already PostgreSQL-native), time-series databases, graph databases, and eventually New SQL or distributed OLTP systems will emerge as first-class deployment targets.

The EntityStore driver interface must be expressive enough that a vector database driver can implement it natively (storing embedding fields as native vectors, running similarity queries through the Filter DSL), not just as a compatibility shim that falls back to text search.

**Required before v1.0:** Filter DSL must include a `VectorSimilarity` filter type, even if no built-in driver implements it at v1.0. Drivers declare which filter types they support. If a query uses a filter type the driver does not support, the error is clean and explicit at query time.

### 5.3 New UI Paradigms

The amis dependency is Awo's most strategically risky external dependency. amis is a strong tool today. In 2035, it may be abandoned, superseded, or incompatible with new web standards.

The SDUI layer must produce an abstract schema (renderer-independent) by v1.0. The amis serialization is a renderer plugin. If this separation exists at v1.0, replacing amis in 2035 is a driver swap. If it does not exist, replacing amis in 2035 is a rewrite of every PageBuilder in every third-party module.

This is the single highest-impact decision in Awo's entire architecture from a long-term ecosystem perspective. Every hour of engineering invested now in defining the abstract SDUI schema saves years of ecosystem migration debt later.

---

## Section 6 — Distributed Systems

### 6.1 The Distributed Systems Gap

Awo's architecture as designed assumes a deployment model that is increasingly rare: a single primary database, a single Redis instance, a single Temporal cluster, a single application binary. This model works for most deployments today. It will not work for the largest deployments in 2035.

The gap is not about scale (read replicas and PgBouncer solve most read scaling). The gap is about failure modes that become common at large scale but are invisible at small scale.

### 6.2 Leader Election and Distributed Scheduling

Awo has cron-like scheduling (background jobs, retry processors, outbox flushing). In a single-instance deployment, the scheduler runs on the single instance. In a multi-instance deployment, every instance runs the scheduler simultaneously, producing duplicate execution.

This is a classic distributed systems problem. The standard solutions are:

- **Advisory locks (PostgreSQL):** Each scheduler acquires a `pg_advisory_lock` before running. First acquirer runs; others skip. Simple, correct, no external dependency. Works until you need sub-second scheduling (lock acquisition latency is too high).

- **Redis Redlock:** Distributed lock across Redis instances. More complex, eventual-consistency risks documented by Redlock's own author (Martin Kleppmann's critique is valid).

- **Dedicated scheduler process:** One process is the scheduler, elected via leader election. Other processes are workers only. Standard in Kubernetes with leader election via `lease` resources.

**v1.0 requirement:** Awo must ship with advisory lock-based leader election for scheduled tasks. It must be the default. The alternative mechanisms are v1.1 drivers. Without this, every multi-instance deployment silently executes every scheduled task N times (where N = instance count).

### 6.3 Distributed Outbox

The transactional outbox is the correct pattern for reliability. At large scale, the outbox has a scaling problem: a single outbox table for all tenants under all entities becomes a hot table. At 100 million events/day, polling the outbox table becomes a significant read load on the primary database.

**Long-term architecture requirement (not v1.0, but must be designed for):**
- Outbox must be partitionable by tenant or by entity type
- Outbox processor must be a driver (different implementations: polling, logical replication, WAL streaming)
- At v1.0, the default processor is polling. At v1.2, logical replication-based processing (no polling, event-driven) must be supported without changing the outbox schema or the entity write path

The outbox schema must include a `partition_key` column (even if unused at v1.0) so that partition strategies can be added without schema migration.

### 6.4 Event Ordering

The outbox guarantees delivery. It does not guarantee order. For most ERP operations, order does not matter: creating a contact and creating an invoice are independent. But for some operations, order is critical: ledger entries must be applied in chronological order or the balance is wrong.

**Required before v1.0:** Events must carry a `sequence_number` per entity instance (not global). Consumers can detect gaps and re-order before applying. The sequence number must be atomic and monotonic per tenant-entity pair. PostgreSQL sequences with `FOR UPDATE` are sufficient at v1.0.

### 6.5 Idempotency

Every event consumer, every workflow activity, every action handler must be idempotent. Today this is a convention ("please implement idempotency in your activities"). At scale, it must be a framework guarantee.

**v1.0 requirement:** The framework must provide an idempotency key mechanism that is automatic for all auto-generated operations (CRUD, standard actions, outbox delivery). For custom code, the framework must provide an idempotency store (simple: Redis key with TTL; complex: PostgreSQL table with constraint). Without this, at-least-once delivery (which the outbox provides) is incompatible with exactly-once semantics — and the application layer cannot close this gap without framework support.

### 6.6 Geo-Partitioning and Regional Failover

Awo today assumes all tenant data is in one database. In 2035, regulated deployments will require data residency: an EU tenant's data must never leave EU infrastructure. A Kenya-only tenant's data must stay in Africa.

This requires:
- Tenant placement metadata: which region(s) is this tenant's data allowed to exist in?
- Database router: given tenant ID, route queries to the correct regional database
- Cross-tenant operations: explicitly forbidden or explicitly permitted with a cross-region flag

**v1.0 requirement:** The EntityStore driver interface must accept a `TenantPlacement` annotation on the context. A geo-aware driver can use this to route to the correct regional endpoint. A single-region driver ignores it. This is one line in the context setup, but if it is not in v1.0, retrofitting it requires changing every driver interface, which is a breaking change.

---

## Section 7 — Security Governance

### 7.1 Trust Boundaries

Awo's trust model today has three principals: platform-admin, tenant-admin, tenant-user. This is correct for a simple SaaS deployment. It is insufficient for regulated deployments.

Regulated deployments require fine-grained principal types:
- Service accounts (machine-to-machine, scoped to specific entity types and actions)
- Delegated principals (user A acts on behalf of user B, with B's consent, limited to specific capabilities)
- Audit principals (read-only, cannot be restricted by policies, used by auditors and compliance systems)
- Emergency access principals (break-glass, requires post-hoc justification, triggers immediate alert)

These are not roles — they are principal types with different trust characteristics. A service account that is granted `role:tenant.admin` must not have the same trust as a human admin with the same role. The trust boundary is orthogonal to the permission boundary.

**v1.0 requirement:** ViewerContext must carry `PrincipalType` (human, service_account, delegated, audit, emergency). Policy evaluation and audit logging must behave differently based on principal type. This cannot be retrofitted by adding roles — it requires a first-class field.

### 7.2 Plugin Isolation

Third-party hooks and policies run in the same process as the framework. A malicious or buggy hook can: access global state, make arbitrary network calls, read memory it should not have access to, panic the entire process, or exfiltrate data by writing to an external endpoint.

Today, the framework has no isolation for third-party code. This is acceptable when all code is written by the same organization. It is not acceptable when third-party modules are from the ecosystem marketplace.

**Architecture options ranked by isolation strength:**
1. Process isolation (subprocess per module) — strongest, highest overhead
2. WASM sandbox — strong, moderate overhead, requires hook interfaces to use serializable types
3. Go `panic` recovery per hook — weak isolation (prevents crash, not data access)
4. Type system restrictions (hook receives only what it is given, no global state access) — weakest, current state

**v1.0 requirement:** At minimum, hooks must run in a goroutine with `recover()`, with a per-hook timeout enforced by `context.WithTimeout`. A hook that panics must not bring down the request. A hook that times out must return a `HookTimeoutError` that is logged and optionally retried.

**v2.0 target:** WASM-based hook isolation for untrusted third-party code. This requires the hook interface redesign (serializable types) that should be designed at v1.0 even if not implemented until v2.

### 7.3 Supply-Chain Security

An Awo deployment in 2035 might include 200 third-party modules from 150 different publishers. The attack surface is massive. The mitigations:

1. **Module signing** — Ed25519 signature over module content hash, verified at startup
2. **Capability declarations** — module must declare what capabilities it needs (network access, file system access, external API calls). Deployer grants or rejects. A module that declares `network: none` cannot make network calls (enforced by WASM sandbox in v2).
3. **Reproducible builds** — module source must produce the same binary deterministically. Deployer can verify.
4. **SBOM (Software Bill of Materials)** — every module publishes its dependency tree. Security scanners can check for known vulnerabilities.
5. **Module audit log** — every action taken by a third-party module is recorded: which entity was accessed, which hook ran, what was written. This is separate from the data audit log.

**v1.0 requirement:** Module signing and capability declarations must be in the module manifest format. Enforcement of capability restrictions is v2 (WASM). Verification of signatures is v1.0.

### 7.4 Field Encryption

Field-level encryption (`FieldDef.Encrypted: true`) is already in the architecture. The missing piece is key governance:

- Key rotation must be a zero-downtime operation. Old keys must decrypt old ciphertext while new keys encrypt new writes. Dual-key period must be configurable.
- Key escrow: in regulated industries (banking, healthcare), key escrow to a designated third party is legally required. The `driver.KeyManagement` interface must support escrow as a first-class operation.
- Tenant key isolation: each tenant's data must be encrypted with that tenant's key. Cross-tenant key sharing is a compliance violation. The key management driver must be tenant-scoped at the interface level.

### 7.5 Audit Integrity

The hash-chained audit log (each record contains `prev_hash`) is correct architecture. Two additional requirements for regulated industries:

1. **External notarization:** High-value audit records (financial transactions, permission changes, emergency access) must be submitted to an external notarization service (RFC 3161 timestamping authority). This proves the record existed at a specific time, independent of whether the internal database is compromised.

2. **Audit log immutability:** The audit log must be append-only at the database level. No `UPDATE` or `DELETE` on audit records, enforced by PostgreSQL trigger. Attempts to modify audit records are themselves audit events. This must be part of the migration setup for the audit table — not an application-layer convention.

---

## Section 8 — Observability

### 8.1 Observability as Architecture, Not Instrumentation

Observability is typically added after the fact ("let's add metrics here"). In a mature framework, observability is architectural — every subsystem is designed with observability as a first-class output.

Awo needs an `obs` package that defines the observability interfaces:

```go
package obs

type Tracer interface {
    StartSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span)
}

type MetricsRecorder interface {
    Counter(name string, labels ...Label) Counter
    Histogram(name string, labels ...Label, buckets []float64) Histogram
    Gauge(name string, labels ...Label) Gauge
}

type StructuredLogger interface {
    Info(msg string, fields ...Field)
    Error(msg string, err error, fields ...Field)
    Debug(msg string, fields ...Field)
}
```

These interfaces are injected into every subsystem at startup. They are not global variables. They are not singletons. This is critical for testability — in tests, you inject a no-op or recording implementation.

The concrete implementations (OpenTelemetry, Prometheus, slog) are drivers of these interfaces.

### 8.2 The Cardinality Problem

Every team that instruments a distributed system eventually encounters label cardinality explosion. The most common mistake is using `tenant_id` as a Prometheus label. At 10,000 tenants, this creates 10,000 × (entities × operations × status) = billions of time series. Prometheus cannot handle this.

**Non-negotiable rule, encoded in the metrics recorder interface:** `TenantID`, `UserID`, `RecordID`, and any other unbounded-cardinality identifier are **never** Prometheus label values. They may appear in structured logs and distributed traces (where cardinality is not a problem), but never in metric labels.

The metrics recorder interface must enforce this by providing a `TenantTier(tenantID) string` helper that maps tenant IDs to cardinality-safe tiers ("free", "standard", "enterprise"). This helper runs at the observability layer, not the application layer.

### 8.3 Distributed Tracing Propagation

Every entity operation must produce a distributed trace span. The span must include:
- Trace ID (propagated from incoming request or generated for background jobs)
- Entity type and ID
- Operation type (create, read, update, delete, action)
- Hook chain execution (each hook is a child span)
- Policy evaluation (one child span per policy)
- SQL query execution (one child span per query)
- Workflow start (if triggered)

This requires the trace context to be propagated through every layer: middleware → handler → hook → store → driver. The context.Context is the propagation mechanism. Every function that takes `ctx context.Context` must propagate the trace context without modification.

**v1.0 requirement:** The framework must inject trace spans automatically for all generated operations (CRUD, actions, hook chains, outbox processing). Custom code receives a trace context via `ctx` and must propagate it. The obs package must provide `obs.SpanFromContext(ctx)` for attaching custom attributes.

### 8.4 Framework-Level Diagnostics

At 500 modules, 100,000 entities, and 15 years of upgrades, operators will encounter situations where the framework behaves unexpectedly. The current diagnostic tooling is insufficient.

**Required diagnostic endpoints:**

`GET /_framework/schema` — returns the compiled schema summary (entity count, field count, route count, compilation time, content hash). Requires platform-admin auth.

`GET /_framework/schema/entity/{name}` — returns the compiled definition of a specific entity (fields, edges, hooks, policies, routes, workflow triggers). Requires platform-admin auth.

`GET /_framework/hooks/{entity}/{lifecycle}` — returns the hook chain for a specific entity and lifecycle stage, in execution order with estimated execution times from last 1000 executions.

`GET /_framework/policies/{entity}/{action}` — returns the policy chain for a specific entity and action, with the current viewer's policy evaluation result.

`GET /_framework/modules` — returns all registered modules, their versions, their capabilities, and their dependency relationships.

`GET /_framework/compiler/passes` — returns all compiler passes that ran, in order, with their execution time and diagnostic output.

`GET /_framework/routes` — returns the full route table with handler descriptors.

These endpoints are the framework's equivalent of Kubernetes' `kubectl describe` — they make the internal state visible without reading source code.

### 8.5 SLO Framework Integration

Awo must ship with a set of default SLIs (Service Level Indicators) and suggested SLOs (Service Level Objectives) that operators can adopt:

| SLI | Suggested SLO |
|---|---|
| p99 entity read latency | < 50ms |
| p99 entity write latency | < 200ms |
| Compilation time at startup | < 30 seconds for 10,000 entities |
| Hook chain execution time | < 100ms per lifecycle stage |
| Policy evaluation time | < 10ms per request |
| Cache hit rate (page schemas) | > 80% |
| Outbox processing lag | < 30 seconds |

These SLIs are measured by the built-in metrics. Operators adopt or override the SLOs. The framework's dashboard (SDUI-generated, naturally) displays SLO compliance status.

---

## Section 9 — Public API Governance

### 9.1 API Surface Classification

Every public API Awo exposes must be classified before v1.0. This is not optional. Once v1.0 ships, reclassifying an API downward (from Stable to Internal) is a breaking change.

**Classification taxonomy:**

| Class | Promise | Examples |
|---|---|---|
| **Kernel** | Never changes. EVER. | Record.ID(), Record.TenantID(), ViewerContext.ActorID(), Filter wire format |
| **Stable** | No breaking changes within major version. Deprecation requires 90 days notice. | EntityDefinition fields, FieldDef fields, HookSet, PolicyFunc signature |
| **Experimental** | May change in minor versions. Use at own risk. | CompilerPass interface, IR types, SDUI abstract schema |
| **Contrib** | Community-maintained. No stability promise. | Specific driver implementations |
| **Internal** | Not public API. May change anytime. Do not import. | Compiler internals, registry internals |

Every exported Go type, function, and method in every public package must have a `// Stability: Kernel|Stable|Experimental|Contrib|Internal` comment. CI enforces that Kernel and Stable APIs are not modified without a manual approval step.

### 9.2 APIs That Must Never Be Public

These are internal implementation details that will tempt module authors to depend on them:

- `CompiledEntity` struct fields (use accessor methods only)
- `Registry.entities` internal map (use `Registry.Get(name)` only)
- SQL template internals (these are driver implementation details)
- Hook chain slice (order is internal, use the hook registration API only)
- Casbin enforcer instance (use the policy evaluation API only)
- Raw pgx connection pool (use EntityStore driver only)

Any of these becoming load-bearing external dependencies before v1.0 creates permanent maintenance debt. CI must prevent importing these from outside framework packages.

### 9.3 The Filter DSL Wire Format

The Filter DSL wire format (the serialized representation of a Filter, sent over network boundaries) is a Kernel API. Once any client serializes a Filter and sends it to Awo, that format is permanently frozen.

This means the wire format must be designed with extreme care before v1.0:

1. It must be self-describing (include type tags for all operators and values)
2. It must be versioned (a `"v"` field so future parsers can handle old formats)
3. It must be extensible (unknown fields are ignored by current parsers, not rejected)
4. It must be formally specified (a grammar document, not just Go code)

The wrong wire format at v1.0 will haunt Awo for twenty years. Every client that serializes filters will break if the format changes. The query language of a data system is its most permanent API. PostgreSQL's SQL dialect is unchanged from 1986. Git's object format is unchanged from 2005.

### 9.4 Version Negotiation

When a v1.0 client connects to a v1.5 server, or vice versa, the behavior must be defined:

- Client sends `Awo-Protocol-Version: 1.0` header
- Server responds with `Awo-Protocol-Version: 1.5` and `Awo-Minimum-Client: 1.0`
- Client and server negotiate the highest mutually supported protocol version
- Operations unsupported in the negotiated version return 409 Conflict with a `version_mismatch` error code

Without protocol version negotiation, every deployment becomes a synchronized upgrade event (all clients must upgrade simultaneously with the server). In enterprise deployments, this is operationally impossible.

---

## Section 10 — Design Simplicity

### 10.1 The Complexity Budget

Every framework has a complexity budget. The budget is the maximum cognitive load a new contributor can absorb before being productive. Exceed the budget and the framework becomes a niche tool maintained by a shrinking priesthood.

Awo's current complexity budget estimate: a skilled Go developer needs approximately 3–4 weeks to become productively familiar with the framework. This is acceptable. Django takes 1–2 weeks (simpler, less opinionated). Kubernetes takes 3–6 months (far more complex, far larger scope). Spring takes 2–4 weeks. Awo is in the right range for its scope.

The risk is that complexity grows monotonically. Every new feature, every new abstraction, every new configuration option adds to the budget. Without a deliberate complexity governance process, the budget doubles every five years.

**Governance requirement:** Before any new abstraction is added to the public API (Kernel or Stable class), it must justify its existence against the complexity budget: "This adds X to the cognitive load. It eliminates Y complexity from user code. Net complexity change: Y - X." If Y < X, the abstraction is rejected.

### 10.2 Candidate Simplifications

These abstractions or concepts may be candidates for elimination or merging in v1.0:

**Merge: BeforeCreate + BeforeUpdate → BeforeSave with operation flag.** Two separate hooks that almost always have identical logic. The operation flag (`op entity.Operation`) tells the hook whether this is create or update. Reduces hook count by 40%. Reduces the HookSet API surface by 2 methods.

**Eliminate: Separate `Custom` entity type designation.** The distinction between System entities (Go struct, typed SQL) and Custom entities (JSONB) is an implementation distinction, not a domain distinction. Users should not need to understand this. Instead, the framework detects the right storage strategy from the field types: if any field is `Currency` or `Link` to a financial entity, system storage is required. Otherwise JSONB is used. This is a compile-time decision, invisible to the module author.

**Merge: PolicyFunc + PermissionSet.** Currently, row-level filtering (PolicyFunc) and operation-level gating (PermissionSet roles) are separate concepts configured separately. They are both aspects of the same question: "Can this viewer do this thing to these records?" A unified policy model (which can express both "operation denied" and "operation allowed but only for these rows") is simpler and more powerful.

**Eliminate: The `DynamicLink` field type.** Polymorphic relationships (link_type + link_name pattern) are notorious sources of query complexity and referential integrity problems. They cannot have real FK constraints. They cannot be efficiently joined. They exist because the schema designer did not know the target type in advance. A better solution: define an interface entity (a capability token that any entity can implement) and use typed edges. This is more complex to implement in the framework but far simpler to use correctly.

**Eliminate: `SmallText` field type.** The distinction between `Data` (varchar(n)) and `SmallText` (varchar(1024)) is a storage detail that leaks into the domain model. Users should not specify varchar lengths in their entity definitions. The framework should choose the right storage type based on declared usage intent (`Searchable`, `Indexed`, `FullText`). `Data` and `SmallText` collapse into `Text` with qualifier annotations.

### 10.3 Configuration vs. Convention

Awo should adopt the Django principle: convention over configuration, but configuration over code. Concretely:

- If 80% of users want the same behavior, it is the default. No configuration required.
- If users want to change the behavior, they set a value in the entity definition (metadata/data). Not code.
- If users want entirely custom behavior, they implement an interface (code). This is the last resort.

Every hook that exists because users can't express their need in metadata is a failure of the metadata language, not a success of the hook system.

---

## Section 11 — Architecture Stress Test

### 11.1 100 Developers

At 100 developers, the primary risks are:

**Conflicting mental models.** Developer A thinks PolicyFunc is for row filtering. Developer B thinks it is for permission decisions. Both are right. Both implement it differently. The system behaves inconsistently.

Mitigation: Every abstraction must have a single, canonical usage pattern documented in the architectural constitution. Deviations require justification. Code review must check for pattern violations.

**Namespace collisions.** Two teams define entities with the same name. Two modules provide the same capability token. Compilation fails with an opaque error.

Mitigation: Entity names are namespaced by module (already: `{module}_{noun}`). Capability tokens must be namespaced by publisher (already: reverse-DNS). Compilation errors for collisions must be specific: "Entity `finance_invoice` defined in modules `finance` and `legacy_finance`. Remove one definition or rename."

**Merge conflicts in definition files.** When 100 developers commit EntityDefinition changes simultaneously, `definition.go` files become merge conflict hotspots.

Mitigation: Each entity definition must be in its own file. No file defines more than one entity. This is enforced by the compiler (one entity per file) or by convention + linter.

### 11.2 500 Modules

At 500 modules, the primary risks are:

**Compilation time.** 500 modules × average 20 entities × 10 fields = 100,000 fields to compile. At 500 microseconds per field (a rough estimate), that is 50 seconds of compile time at startup. Unacceptable.

Mitigation: Incremental compilation (Section 3.3). Modules that have not changed since last compilation are restored from cache. Cold compile of 500 modules happens once. Warm compile (most modules unchanged) takes seconds.

**Dependency hell.** Module A requires `crm.contact_crud >= 2.0`. Module B provides `crm.contact_crud 1.9`. Conflict. But Module C also requires Module B at 1.9 for unrelated reasons. No resolution exists.

Mitigation: Capability tokens must be independently versioned from the module that provides them. A capability token `crm.contact_crud` has its own version, provided by whichever module implements it. Consumers declare `requires: crm.contact_crud >= 2.0`. Providers declare `provides: crm.contact_crud 2.1`. The resolver matches capability versions, not module versions. This is the Go module approach applied to capabilities.

**Silent behavioral overrides.** Module A defines a hook for `finance_invoice.BeforeCreate`. Module B also defines a hook for `finance_invoice.BeforeCreate`. Both run. Module B's hook overrides a value set by Module A's hook, breaking Module A's invariants.

Mitigation: Hooks from different modules must declare their interaction expectations:
```go
type HookRegistration struct {
    Hook     Hook
    Priority int    // lower = earlier
    Requires []string  // "finance.base_hook must run before this"
    Conflicts []string  // "cannot coexist with module X's hook on this entity"
}
```
The compiler detects conflicts at startup, not at runtime.

### 11.3 100,000 Entities

At 100,000 entity definitions across a large multi-tenant ecosystem:

**Route table explosion.** 100,000 entities × 5 standard routes = 500,000 routes. Standard HTTP routers (including Fiber's) do not perform well at this scale.

Mitigation: Routes must use parametric patterns, not literal patterns. Not `/api/v1/entities/finance_invoice/...` as a distinct route — instead `/api/v1/entities/{type}/...` as a single parametric route that dispatches by entity type. The route table is a dispatch table, not a flat list.

**Compilation memory.** 100,000 entities × compiled artifacts per entity. Memory usage during compilation must be bounded. The compiler must release intermediate artifacts as soon as they are no longer needed.

**Schema cache size.** 100,000 entities × SDUI schemas × tenants × 5 minute TTL. Redis memory usage must be bounded. The cache must use LRU eviction, not unbounded growth.

### 11.4 Cognitive Complexity Growth

The most dangerous form of complexity is not size — it is interaction complexity. At 100 modules, 1,000 hook registrations, and 100,000 entity definitions, the question "why is this entity behaving this way" requires understanding potentially hundreds of interacting hooks, policies, and compiler passes.

**Required:** The diagnostic endpoints (Section 8.4) must answer "why did this entity behave this way for this request" — a causality trace. For every request that involved hook execution or policy evaluation, the framework must record (in the trace, not the audit log) which hooks ran in which order, which policies were evaluated with what result, and which workflow was triggered.

This causality trace is the framework equivalent of a debugger. Without it, debugging at scale requires reading code, not observing behavior.

---

## Section 12 — Final Governance Verdict

### 12.1 Architecture Strengths

**S1 — Metadata-First Design Is Correct**
The decision to make `EntityDefinition` the central primitive is architecturally sound and will age well. It is the same insight that made Kubernetes (everything is a resource with a schema) and Frappe (doctype as the central abstraction) successful. Metadata-first frameworks are easier to introspect, audit, and generate tooling around. This is a 20-year correct decision.

**S2 — Strict Tenancy Is Non-Negotiable**
Encoding multi-tenancy as a structural invariant rather than a feature is architecturally mature. Every framework that tries to retrofit multi-tenancy fails. Django-tenants is a cautionary tale. Awo starts with the right constraint. This is a 20-year correct decision.

**S3 — Driver Abstraction Preserves Optionality**
By abstracting PostgreSQL, Redis, Temporal, and Fiber behind driver interfaces, Awo preserves the ability to replace any of them. This will be used. PostgreSQL will not be the dominant OLTP database in 2045. Temporal will have competitors. Fiber will be superseded. The driver model means these replacements are driver swaps, not rewrites. This is a 20-year correct decision.

**S4 — Compilation-Unit Module Model Is Correct For Target Domain**
Go's `plugin` package is a footgun. Awo's decision to compile all modules into a single binary is correct for the regulated-industry target market. Auditability, reproducibility, and operational simplicity outweigh the convenience of hot-loading. This is a 20-year correct decision.

**S5 — Transactional Outbox Is The Right Reliability Pattern**
Writing to the outbox inside the same transaction as the entity write is the correct solution to the dual-write problem. It is not exotic — it is well-understood and battle-tested. This is a 20-year correct decision.

**S6 — Five-Layer Architecture Is Sound**
UI → API → Domain → Workflow → Store, with strict top-down dependency, prevents the circular dependency nightmares that plague large codebases. This structure is understandable and enforceable. This is a 20-year correct decision.

### 12.2 Architectural Debt

**D1 — No Formal Architectural Constitution**
The philosophy exists but is not written down with binding force. This is immediate debt — it accumulates from the first day a contributor makes a decision without knowing the philosophy.

**D2 — SDUI Coupled To amis**
Every PageBuilder in every module is written against the amis schema format. When amis is replaced or abandoned, this is a framework-wide migration event. The abstract SDUI layer must exist before v1.0.

**D3 — No IR In The Compiler**
The compiler transforms EntityDefinition directly to multiple output formats without an intermediate representation. This prevents optimization passes, plugin compiler hooks, and incremental compilation. This is technical debt that grows with every new feature that needs compilation-time processing.

**D4 — Filter DSL Wire Format Not Formally Specified**
The wire format is implied by the Go implementation. Any client that serializes Filters is depending on an undocumented format that can change. This must be formally specified before any external client exists.

**D5 — Hooks Have No Isolation**
Third-party hooks run in the same goroutine as the request with no timeout enforcement. A buggy hook can wedge a request indefinitely. This is acceptable for first-party code. It is unacceptable for marketplace modules.

**D6 — No Abstract API Descriptor**
REST routes are generated directly from EntityDefinition. There is no abstract API descriptor that GraphQL, gRPC, or other transports could also render. This couples Awo to REST as the API paradigm, which is a 10-year risk.

### 12.3 Architectural Risks

**R1 — Ecosystem Fragmentation (Severity: Critical)**
If the module ecosystem launches without stable APIs, stability guarantees, and LTS designation, the ecosystem will fragment into incompatible islands within 5 years. This is the PHP4→PHP5 problem. The fix is governance, not code.

**R2 — Compiler Scalability (Severity: High)**
Without incremental compilation, the framework's startup time grows linearly with the number of entity definitions. At 10,000 definitions, this is a significant operational problem. At 100,000, it is prohibitive.

**R3 — amis Dependency Lock-In (Severity: High)**
The amis SDK is pinned in the repository. amis is an open-source project with uncertain long-term maintenance. If amis is abandoned or incompatible with a critical web standard, every PageBuilder in every module must be migrated simultaneously. This is a 5–15 year risk that becomes existential if not addressed at the SDUI architecture level before v1.0.

**R4 — Go Version Coupling (Severity: Medium)**
`go.mod` declares Go 1.26.3, which does not exist. This suggests the framework's Go version discipline is informal. Go releases annually. Major Go upgrades occasionally require code changes. The framework must have a formal Go version support policy: support the two most recent Go releases, declare minimum required version in the module manifest.

**R5 — Multi-Region Operational Gap (Severity: Medium)**
The architecture has no geo-routing, no regional placement metadata, and no cross-region consistency model. For global enterprise deployments, this is a deployment blocker. The hook points must exist in v1.0 even if implementations are v1.x.

**R6 — Audit Integrity For Regulated Industries (Severity: Medium)**
Hash-chained audit log exists but lacks external notarization and database-level immutability enforcement. Banking and healthcare deployments require these. Without them, Awo cannot claim regulatory compliance in those verticals.

**R7 — Plugin Supply Chain (Severity: Medium, growing over time)**
Without module signing and capability declarations, a compromised module in the ecosystem can silently exfiltrate data from every tenant in every deployment that includes it. This risk is low today (small ecosystem) and catastrophic at scale (thousands of modules, millions of tenants).

**R8 — Hook Ordering Between Modules (Severity: Low today, High at 100+ modules)**
Hook execution order between modules is undefined. Two modules can register hooks on the same entity lifecycle stage with no declared ordering constraint. Silent behavioral interactions will produce bugs that are impossible to diagnose without framework-level causality tracing.

### 12.4 Architectural Contradictions

**C1 — "Never expose implementation" vs. amis-specific PageBuilders**
The framework philosophy says metadata drives behavior and drivers isolate implementation. But PageBuilders produce amis-specific JSON, coupling business logic to a specific renderer. This is a direct contradiction.

**C2 — "Drivers are interfaces" vs. bootstrap package that knows specific drivers**
The `framework/bootstrap` package imports concrete driver packages and wires them together. This means framework code knows about specific infrastructure. The bootstrap package must be moved to `cmd/` and the framework must provide only the builder interface.

**C3 — "Compile-time over runtime" vs. runtime type assertions for optional interfaces**
Optional interface detection via type assertions is a runtime mechanism. For optional interfaces that are critical (e.g., `OrgPathViewer` in a policy that requires org-path scoping), a missing implementation is silently handled by a fallback rather than a compile error. This is unavoidable in Go's type system, but the framework should provide a mechanism for modules to declare required optional interfaces on their dependencies, checked at compilation time.

**C4 — "Tenancy is structural" vs. query param tenant_id fallback**
The specification says `tenant_id` query parameter is supported for webhooks and legacy. A query param that bypasses header-based tenant identification is an architectural exception to "tenancy is structural." It must be configurable-off (disabled by default in production mode) and every request that uses it must produce an audit event.

### 12.5 Irreversible Decisions

These decisions, once made and shipped in v1.0, cannot be changed without breaking the ecosystem:

**I1 — Record Identity Format**
UUID v7 (time-ordered) as the record identity format. Once records exist with this format, changing it requires migrating every record, invalidating every cache key, breaking every workflow ID that includes a record ID. This is a permanent commitment. UUID v7 is correct. Document it as permanent.

**I2 — Tenant Isolation Mechanism**
PostgreSQL RLS via `set_tenant_context()` stored procedure is the isolation mechanism. Changing this requires schema migration of every table in every tenant. This is a 20-year commitment. RLS with `set_config` transaction-local is the correct choice for PostgreSQL. Document it as permanent.

**I3 — Filter DSL Semantics**
The semantic meaning of each filter operator (Eq, Neq, Lt, Gt, In, Contains, etc.) is permanent. Changing semantics breaks every query that uses that operator. Get the semantics right before v1.0.

**I4 — Hook Lifecycle Stage Names**
The string names of lifecycle stages (`before_create`, `after_save`, etc.) are permanent if any external system (monitoring, audit log) references them as string values. If stage names are in the audit log, they are permanent from the moment the first audit record is written.

**I5 — Module Manifest Format**
Once third-party modules publish manifests, the format is permanent. Changing it requires all third-party module authors to update their manifests. This is feasible at 10 modules. Impossible at 10,000.

**I6 — Entity Name Convention**
`{module}_{noun}` format (e.g., `finance_invoice`) is permanent. It appears in migration filenames, Temporal workflow IDs, Redis cache keys, Casbin policy rules, URL paths, and API documentation. Changing it requires migrating all of these simultaneously.

**I7 — The EntityDefinition Field Order Is Public API**
If `EntityDefinition` struct tags or field order affects the compiled schema content hash, changing the struct layout changes content hashes for all existing compiled schemas. Either the content hash must not depend on struct layout, or the struct layout must be frozen.

### 12.6 Things That Must Change Before v1.0

**MUST-1 — Abstract SDUI Schema (Critical)**
Define a renderer-independent SDUI schema. amis serialization becomes a renderer plugin. This is the single highest-impact architectural change that must happen before v1.0. Every month of delay adds to the ecosystem migration cost when amis is eventually replaced.

**MUST-2 — Abstract API Descriptor (Critical)**
Compile-time generation of an abstract API descriptor (entity capabilities, action signatures, field types) that REST, GraphQL, and gRPC can each render. REST is the first renderer. This decouples the API model from the transport.

**MUST-3 — Formal Architectural Constitution (High)**
Five non-negotiable axioms, formally documented and enforced in CI. Not a README. A binding governance document that is checked against every API proposal.

**MUST-4 — Filter DSL Formal Specification (High)**
Grammar document, version field in wire format, semantic specification for all operators. Not derivable from the Go implementation.

**MUST-5 — CompiledSchema Content Hash (High)**
SHA-256 of all inputs, compiler version stamp, serialization interface (even if experimental). Enables cache validation and incremental compilation.

**MUST-6 — Compiler IR Definition (High)**
IREntity, IRField, IREdge, IRHookChain, IRPolicyChain, IRRouteDescriptor, IRWorkflowBinding. Not necessarily exposed as public API, but must exist internally before compilation passes are written.

**MUST-7 — Module Manifest Format Final (High)**
Including module signing field (Ed25519), capability token namespacing, version constraints on capabilities (not just modules), and `AuditRequired` flag.

**MUST-8 — Hook Timeout Enforcement (High)**
Every hook runs in a goroutine with `context.WithTimeout`. Panic recovery mandatory. HookTimeoutError type defined.

**MUST-9 — PrincipalType on ViewerContext (High)**
`PrincipalType` field (human, service_account, delegated, audit, emergency). Required for regulated industry deployments.

**MUST-10 — Leader Election for Scheduled Tasks (Medium)**
PostgreSQL advisory lock-based leader election for outbox processor, cron scheduler, retry processor. Default behavior in multi-instance deployments must be correct without configuration.

**MUST-11 — PolicyResult Audit Rationale (Medium)**
`PolicyResult` must carry `PolicyRationale` for audit logging of access control decisions. Required for ISO 27001.

**MUST-12 — OnRead and AfterCommit Lifecycle Events (Medium)**
Complete the hook lifecycle. Missing events will cause ecosystem workarounds that become load-bearing.

**MUST-13 — TenantAction (Cross-Entity Actions) (Medium)**
Actions that span entities or are tenant-scoped without a primary entity. Without this, module authors will pollute entity definitions with cross-cutting actions.

**MUST-14 — Outbox Partition Key (Medium)**
`partition_key` column in outbox schema, unused at v1.0 but required for future partitioning without schema migration.

**MUST-15 — Tenant Placement on Context (Medium)**
`TenantPlacement` annotation in query context for geo-routing. One line in the context setup. Impossible to retrofit across all driver interfaces after v1.0.

### 12.7 Things That Must Never Change After v1.0

- Record identity format (UUID v7)
- Tenant isolation mechanism (RLS via `set_tenant_context()`)
- Entity naming convention (`{module}_{noun}`)
- Filter DSL operator semantics (once specified)
- Hook lifecycle stage names (as they appear in audit logs)
- Module manifest format (after first third-party module publishes)
- ViewerContext minimum interface (`ActorID`, `TenantID`, `Roles`, `IsSystem`, `HasRole`)
- The five architectural axioms in the constitutional document
- CloudEvents v1.0 as the event wire format

### 12.8 Things Intentionally Deferred Until v2

- WASM-based hook isolation for untrusted third-party code
- Disk-persistent compilation cache
- GraphQL and gRPC transport drivers (protocol negotiation exists at v1.0, drivers at v2)
- Distributed compilation across multiple processes
- Offline-first sync driver
- Stream processing entity source type
- Cross-region tenant migration tooling
- Module marketplace infrastructure (discovery, ratings, certification badges)
- AI agent principal type in ViewerContext

### 12.9 Final Readiness Score

| Dimension | Score | Rationale |
|---|---|---|
| Architectural Philosophy | 62/100 | Correct intuition, not formally committed |
| Core Abstractions | 71/100 | Mostly correct, 3–4 missing or over-specified |
| Compiler Architecture | 44/100 | No IR, no incremental compilation, no plugin hooks |
| Module Ecosystem | 38/100 | Manifest format incomplete, no dependency resolver, no signing |
| Long-Term Evolution | 55/100 | SDUI coupling and API transport coupling are 20-year risks |
| Distributed Systems | 41/100 | No leader election, no geo-routing hooks, no idempotency store |
| Security Governance | 58/100 | Good foundation, field encryption present, plugin isolation missing |
| Observability | 63/100 | Good intent, cardinality problem not solved, no causality tracing |
| Public API Governance | 35/100 | No stability classification, no wire format specification |
| Design Simplicity | 67/100 | Some unnecessary complexity remains, complexity budget not formal |
| Architecture Stress Test | 52/100 | Unknown behavior at 500 modules, no hook ordering guarantees |
| **Overall** | **53/100** | |

### 12.10 Would You Approve This Architecture?

**No. Not yet.**

The foundation is sound. The core insight (metadata-first, structurally multi-tenant, driver-isolated, compilation-unit modular) is correct and will age well. The five-layer architecture is clean. The EntityDefinition primitive is the right central abstraction. The transactional outbox is the right reliability pattern. These decisions are good for twenty years.

But three architectural gaps exist that, if not closed before v1.0, will force a v2 before 2035:

**Gap 1 — The SDUI coupling to amis is the ecosystem's single largest long-term risk.** When amis is abandoned or superseded, every PageBuilder in every third-party module written between 2025 and 2035 will need to be rewritten. The abstract SDUI schema must exist before v1.0 ships. Without it, the ecosystem is betting its longevity on a single third-party renderer.

**Gap 2 — The absence of an abstract API descriptor means REST is permanently the API model.** In 2035, gRPC will be the standard for service-to-service communication in enterprise software. GraphQL will be the standard for data-intensive frontends. If the API model cannot produce abstract capability descriptions that multiple transports render, Awo will be permanently REST-only. This is an architectural dead end for enterprise adoption in the 2030s.

**Gap 3 — The module ecosystem governance is insufficient.** The manifest format, dependency resolver, capability versioning, and module signing must exist before the first third-party module ships. Once the ecosystem forms, these cannot be added without breaking every existing module. The Django 2→3 fragmentation took three years to recover from. Awo has the opportunity to get governance right before the ecosystem forms.

Close these three gaps. Implement the fifteen MUST items. Then run this review again.

The architecture that results from closing these gaps would score **78–82/100** — high enough to approve for a twenty-year lifecycle. Not 100, because no architecture is perfect and the remaining deferred items (WASM isolation, distributed compilation, AI agent support) are known risks with known mitigation paths.

At **53/100**, the current architecture is not ready for a twenty-year freeze.

---

## 2045 Architecture Verdict

*Written as if from 2045, looking back.*

Awo v1.0 shipped in 2026, six months later than planned, after this review forced three additional architectural changes: the abstract SDUI schema, the abstract API descriptor, and the module manifest with capability versioning. Those six months were the most valuable six months in the framework's history.

The abstract SDUI schema allowed the React-based renderer to replace amis in 2031 without ecosystem disruption. 400 third-party modules updated their renderer driver in a single quarter. Without that schema, 400 module authors would have had to rewrite their PageBuilders — most would have abandoned their modules instead.

The abstract API descriptor allowed gRPC and GraphQL transport drivers to be released in v1.3 and v2.1 respectively, without any module author needing to change their EntityDefinition. Enterprise customers who had been waiting for gRPC support adopted Awo in 2032, tripling the enterprise customer count in 18 months.

The module manifest with capability versioning prevented the fragmentation event that destroyed three competing ERP frameworks between 2028 and 2033. Awo's ecosystem of 2,400 modules in 2045 is coherent because dependency resolution has been hermetic since v1.0.

The three gaps that this review identified and forced closed before v1.0 are the reason Awo exists in 2045 rather than being a case study in "frameworks that almost made it."

The compiler IR, added in v1.1, enabled the AI schema generator released in v3.0 (2033) — the feature that made Awo the dominant ERP framework for SMB SaaS deployments globally.

The formal architectural constitution, written before v1.0, has been cited 847 times in architecture review discussions. Every time a contributor proposed violating an axiom, the constitution gave reviewers the language to explain why it could not be done. It prevented approximately 23 would-be breaking changes in the first decade.

The decisions that were deferred: WASM hook isolation shipped in v2.0 (2028). AI agent principal type shipped in v3.0 (2033). Distributed compilation is still in progress in v5.2 (2042) — some things take longer than planned.

The decisions that were permanent and correct: UUID v7 record identity. RLS via `set_tenant_context()`. CloudEvents v1.0 wire format. Five architectural axioms. `{module}_{noun}` entity naming. These have not changed in twenty years. They did not need to.

The decisions that were permanent and wrong: one. The `DynamicLink` field type was kept in v1.0 despite this review recommending its removal. By 2035, 15% of all modules in the ecosystem used DynamicLink. Removing it in v4.0 (2038) required the largest migration guide ever written for any software framework. It took the ecosystem two years to complete.

Listen to the reviews. Remove the DynamicLink.

---

*This document constitutes the final architectural governance review for Awo Framework v1.0. It is an Architecture Decision Record with binding force on the v1.0 freeze decision. It does not expire. Future architecture reviews must reference this document and explain deviations from its findings.*

*Signed: Chief Software Architect*
*Date: 2025-07-06*
*Status: RECOMMENDS CONDITIONAL APPROVAL pending closure of 15 MUST items*
