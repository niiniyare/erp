# Awo Framework — Documentation Master Blueprint
## Chief Technical Writer Design Document | Pre-v1.0 | 2025-07-06

**Classification:** Documentation Architecture Blueprint
**Status:** Master plan — no content written yet
**Scope:** Complete `awo/docs/` directory, 2025–2045 lifecycle

---

## Part I — Directory Tree

```
docs/
│
├── README.md
├── GLOSSARY.md
├── CHANGELOG.md
│
├── 01-introduction/
│   ├── README.md
│   ├── philosophy.md
│   ├── design-goals.md
│   ├── non-goals.md
│   ├── architecture-overview.md
│   ├── why-go.md
│   ├── why-postgresql.md
│   ├── why-compile-metadata.md
│   ├── why-drivers.md
│   ├── comparison.md
│   └── quick-orientation.md
│
├── 02-theory/
│   ├── README.md
│   ├── entity-model.md
│   ├── metadata-model.md
│   ├── compilation-theory.md
│   ├── execution-model.md
│   ├── tenancy-model.md
│   ├── security-model.md
│   ├── consistency-model.md
│   ├── storage-model.md
│   ├── policy-model.md
│   ├── workflow-model.md
│   ├── event-model.md
│   ├── sdui-model.md
│   ├── extension-philosophy.md
│   ├── failure-philosophy.md
│   └── performance-philosophy.md
│
├── 03-kernel/                          ← NORMATIVE. Frozen at v1.0.
│   ├── README.md
│   ├── stability-policy.md
│   ├── versioning.md
│   ├── breaking-change-policy.md
│   ├── compatibility-matrix.md
│   ├── invariants.md                   ← LAW-001 through LAW-020
│   │
│   └── types/
│       ├── README.md
│       ├── entity-definition.md
│       ├── field-def.md
│       ├── field-type.md
│       ├── edge-def.md
│       ├── hook-registration.md
│       ├── lifecycle-stage.md
│       ├── record.md
│       ├── mutable-record.md
│       ├── viewer-context.md
│       ├── policy-func.md
│       ├── policy-result.md
│       ├── action-def.md
│       ├── workflow-trigger.md
│       ├── trigger-event.md
│       ├── filter.md
│       ├── query-option.md
│       ├── create-input.md
│       ├── update-input.md
│       ├── page-info.md
│       ├── aggregate-spec.md
│       ├── aggregate-result.md
│       ├── compiled-schema.md
│       ├── module-manifest.md
│       ├── capability-token.md
│       └── ui-page.md
│
├── 04-compiler/
│   ├── README.md
│   ├── overview.md
│   ├── phases/
│   │   ├── README.md
│   │   ├── 01-registration.md
│   │   ├── 02-name-resolution.md
│   │   ├── 03-semantic-analysis.md
│   │   ├── 04-policy-compilation.md
│   │   ├── 05-route-table.md
│   │   ├── 06-hook-chain.md
│   │   ├── 07-sql-templates.md
│   │   ├── 08-sdui-tree.md
│   │   ├── 09-workflow-graph.md
│   │   └── 10-schema-seal.md
│   ├── ir.md
│   ├── dependency-graph.md
│   ├── validation-passes.md
│   ├── generated-artifacts.md
│   ├── fingerprinting.md
│   ├── incremental-compilation.md
│   ├── plugin-passes.md
│   ├── diagnostics.md
│   └── cache.md
│
├── 05-runtime/
│   ├── README.md
│   ├── boot-sequence.md
│   ├── middleware-pipeline.md
│   ├── request-lifecycle.md
│   ├── entity-lifecycle.md
│   ├── hook-execution.md
│   ├── policy-evaluation.md
│   ├── transaction-boundaries.md
│   ├── workflow-dispatch.md
│   ├── outbox-processing.md
│   ├── cache-lifecycle.md
│   ├── background-workers.md
│   ├── scheduler.md
│   ├── leader-election.md
│   ├── graceful-shutdown.md
│   └── crash-recovery.md
│
├── 06-drivers/
│   ├── README.md
│   ├── driver-contract.md
│   ├── optional-interfaces.md
│   ├── conformance/
│   │   ├── README.md
│   │   ├── entity-store.md
│   │   ├── session-store.md
│   │   ├── workflow-driver.md
│   │   ├── search-driver.md
│   │   ├── cache-driver.md
│   │   ├── ui-renderer.md
│   │   ├── outbox-driver.md
│   │   └── key-management.md
│   ├── developing/
│   │   ├── entity-store.md
│   │   ├── session-store.md
│   │   ├── workflow-driver.md
│   │   ├── search-driver.md
│   │   ├── cache-driver.md
│   │   ├── ui-renderer.md
│   │   ├── auth-provider.md
│   │   ├── outbox-driver.md
│   │   └── key-management.md
│   ├── testing-drivers.md
│   └── certification.md
│
├── 07-framework-dev/
│   ├── README.md
│   ├── repository-structure.md
│   ├── architecture-laws.md
│   ├── import-rules.md
│   ├── layer-rules.md
│   ├── coding-standards.md
│   ├── concurrency-rules.md
│   ├── memory-rules.md
│   ├── compatibility-rules.md
│   ├── performance-rules.md
│   └── review-checklist.md
│
├── 08-app-dev/
│   ├── README.md
│   ├── getting-started.md
│   │
│   ├── entities/
│   │   ├── README.md
│   │   ├── system-entities.md
│   │   ├── custom-entities.md
│   │   ├── choosing-entity-type.md
│   │   ├── naming-conventions.md
│   │   └── entity-lifecycle.md
│   │
│   ├── fields/
│   │   ├── README.md
│   │   ├── field-types.md
│   │   ├── constraints.md
│   │   ├── custom-fields.md
│   │   ├── sensitive-fields.md
│   │   ├── encrypted-fields.md
│   │   └── field-lifecycle.md
│   │
│   ├── relationships/
│   │   ├── README.md
│   │   ├── edges.md
│   │   ├── link-fields.md
│   │   └── querying-edges.md
│   │
│   ├── policies/
│   │   ├── README.md
│   │   ├── operation-permissions.md
│   │   ├── row-filters.md
│   │   ├── policy-composition.md
│   │   └── built-in-policies.md
│   │
│   ├── hooks/
│   │   ├── README.md
│   │   ├── hook-lifecycle.md
│   │   ├── writing-hooks.md
│   │   ├── hook-ordering.md
│   │   ├── cross-module-hooks.md
│   │   └── testing-hooks.md
│   │
│   ├── actions/
│   │   ├── README.md
│   │   ├── custom-actions.md
│   │   ├── action-permissions.md
│   │   └── tenant-actions.md
│   │
│   ├── workflows/
│   │   ├── README.md
│   │   ├── triggers.md
│   │   ├── writing-workflows.md
│   │   ├── writing-activities.md
│   │   ├── saga-pattern.md
│   │   ├── long-running.md
│   │   └── testing-workflows.md
│   │
│   ├── ui/
│   │   ├── README.md
│   │   ├── page-builders.md
│   │   ├── sdui-schema.md
│   │   ├── custom-pages.md
│   │   └── permissions-and-ui.md
│   │
│   ├── modules/
│   │   ├── README.md
│   │   ├── module-structure.md
│   │   ├── manifests.md
│   │   ├── capabilities.md
│   │   ├── dependencies.md
│   │   ├── versioning.md
│   │   └── publishing.md
│   │
│   ├── testing/
│   │   ├── README.md
│   │   ├── unit-testing.md
│   │   ├── integration-testing.md
│   │   └── workflow-testing.md
│   │
│   └── deployment/
│       ├── README.md
│       ├── binary-build.md
│       ├── configuration.md
│       └── environment.md
│
├── 09-operations/
│   ├── README.md
│   │
│   ├── deployment/
│   │   ├── README.md
│   │   ├── rolling-upgrades.md
│   │   ├── blue-green.md
│   │   ├── canary.md
│   │   └── air-gapped.md
│   │
│   ├── scaling/
│   │   ├── README.md
│   │   ├── horizontal-scaling.md
│   │   ├── connection-pooling.md
│   │   └── leader-election.md
│   │
│   ├── migrations/
│   │   ├── README.md
│   │   ├── running-migrations.md
│   │   ├── rolling-back.md
│   │   ├── zero-downtime-patterns.md
│   │   └── cross-module-dependencies.md
│   │
│   ├── observability/
│   │   ├── README.md
│   │   ├── metrics.md
│   │   ├── structured-logging.md
│   │   ├── distributed-tracing.md
│   │   ├── health-checks.md
│   │   └── slo-reference.md
│   │
│   ├── disaster-recovery/
│   │   ├── README.md
│   │   ├── backup-strategy.md
│   │   ├── restore-procedures.md
│   │   ├── failover.md
│   │   └── partition-behavior.md
│   │
│   └── multi-region/
│       ├── README.md
│       ├── tenant-placement.md
│       ├── geo-routing.md
│       └── regional-failover.md
│
├── 10-diagnostics/
│   ├── README.md
│   ├── schema-inspection.md
│   ├── hook-tracing.md
│   ├── policy-tracing.md
│   ├── compilation-diagnostics.md
│   ├── profiling.md
│   │
│   └── troubleshooting/
│       ├── README.md
│       ├── startup-failures.md
│       ├── rls-failures.md
│       ├── migration-failures.md
│       ├── workflow-failures.md
│       ├── performance-degradation.md
│       └── cache-issues.md
│
├── 11-security/
│   ├── README.md
│   ├── threat-model.md
│   ├── trust-boundaries.md
│   ├── supply-chain.md
│   ├── authentication.md
│   ├── authorization.md
│   ├── row-level-security.md
│   ├── secrets-management.md
│   ├── audit-log.md
│   ├── field-encryption.md
│   ├── attack-surface.md
│   │
│   └── compliance/
│       ├── README.md
│       ├── regulated-industries.md
│       ├── gdpr.md
│       ├── hipaa.md
│       └── kenya-dpa.md
│
├── 12-reference/
│   ├── README.md
│   ├── glossary.md                     ← canonical, referenced everywhere
│   ├── configuration.md
│   ├── environment-variables.md
│   ├── field-types.md
│   ├── filter-dsl.md
│   ├── lifecycle-stages.md
│   ├── event-types.md
│   ├── error-catalog.md
│   ├── system-roles.md
│   ├── http-api.md
│   ├── reserved-names.md
│   ├── migration-reference.md
│   └── driver-catalog.md
│
├── 13-adrs/
│   ├── README.md
│   ├── template.md
│   ├── ADR-001-metadata-first-architecture.md
│   ├── ADR-002-compilation-unit-modules.md
│   ├── ADR-003-postgresql-rls-isolation.md
│   ├── ADR-004-temporal-workflow-engine.md
│   ├── ADR-005-transactional-outbox.md
│   ├── ADR-006-abstract-sdui-renderer.md
│   ├── ADR-007-casbin-policy-engine.md
│   ├── ADR-008-uuid-v7-record-identity.md
│   ├── ADR-009-driver-abstraction-model.md
│   ├── ADR-010-open-field-type-system.md
│   ├── ADR-011-hook-registration-model.md
│   ├── ADR-012-filter-dsl-wire-format.md
│   ├── ADR-013-cloudevents-wire-format.md
│   ├── ADR-014-module-capability-tokens.md
│   ├── ADR-015-system-viewer-semantics.md
│   │
│   ├── rejected/
│   │   ├── REJ-001-go-plugin-package.md
│   │   ├── REJ-002-orm-direct-access.md
│   │   ├── REJ-003-graphql-first-api.md
│   │   ├── REJ-004-dynamic-link-field.md
│   │   ├── REJ-005-session-mode-pgbouncer.md
│   │   └── REJ-006-hookset-struct.md
│   │
│   └── deprecated/
│       └── README.md
│
├── 14-contributing/
│   ├── README.md
│   ├── development-setup.md
│   ├── contribution-workflow.md
│   ├── testing-guide.md
│   ├── benchmarks.md
│   ├── ci-cd.md
│   ├── release-process.md
│   ├── documentation-style-guide.md
│   └── code-review-process.md
│
└── 15-appendices/
    ├── README.md
    │
    ├── examples/
    │   ├── README.md
    │   ├── invoice-module/
    │   │   ├── README.md
    │   │   ├── definition.go
    │   │   ├── hooks.go
    │   │   ├── policy.go
    │   │   ├── workflow.go
    │   │   └── migration.sql
    │   ├── crm-contact/
    │   │   ├── README.md
    │   │   └── ...
    │   └── approval-workflow/
    │       ├── README.md
    │       └── ...
    │
    ├── patterns/
    │   ├── README.md
    │   ├── saga-pattern.md
    │   ├── event-driven-audit.md
    │   ├── multi-step-approval.md
    │   ├── bulk-import.md
    │   ├── scheduled-report.md
    │   └── cross-module-integration.md
    │
    ├── anti-patterns/
    │   ├── README.md
    │   ├── raw-sql-in-hooks.md
    │   ├── float-for-money.md
    │   ├── bypassing-rls.md
    │   ├── lazy-loading-edges.md
    │   ├── conditional-registration.md
    │   ├── dynamic-link-field.md
    │   └── mutable-compiled-schema.md
    │
    ├── diagrams/
    │   ├── README.md
    │   ├── package-dependency-graph.svg
    │   ├── compilation-flow.svg
    │   ├── boot-sequence.svg
    │   ├── request-lifecycle.svg
    │   ├── hook-sequence.svg
    │   ├── transaction-boundary.svg
    │   ├── outbox-flow.svg
    │   ├── module-dependency-resolution.svg
    │   └── tenant-isolation.svg
    │
    └── formal/
        ├── README.md
        ├── architecture-laws.md        ← LAW-001–LAW-020, formal statement
        ├── filter-grammar.md           ← BNF grammar for Filter DSL
        └── state-machines.md           ← entity, tenant, field, hook lifecycles
```

**Total:** 178 files across 15 sections.

---

## Part II — Directory Rationale

### Why 15 top-level sections, not fewer?

**Audience isolation.** A driver author never needs to read the app dev guide. An operator never needs the kernel specification. A security auditor needs only sections 03, 11, 13. Flat documentation forces every reader through content irrelevant to them. Numbered directories communicate reading order without enforcing it.

**Stability isolation.** `03-kernel/` is frozen at v1.0 and must never change. `08-app-dev/` evolves every minor version. `13-adrs/` is append-only. Mixed sections produce mixed stability signals. Separate directories let the version control history clearly show which section changed and why.

**Generation isolation.** `12-reference/` is partially auto-generated. Mixing auto-generated and hand-written files in the same directory produces maintenance confusion. Auto-generated files belong in `12-reference/` exclusively.

**Normative vs. informative isolation.** Sections 03, 04, 05, and `15-appendices/formal/` are normative (specification). Everything else is informative (guide). Readers of normative docs need different context than readers of guides.

### Why `02-theory/` before kernel specification?

The kernel defines what the types are. Theory explains why they exist and why they are shaped the way they are. A reader who reads kernel types without theory memorizes an API. A reader who reads theory first understands the design. Understanding beats memorization over 20 years.

### Why separate `06-drivers/conformance/` from `06-drivers/developing/`?

Conformance documents are normative tests — they define what a correct driver is. Developing documents are informative guides — they describe how to build one. A driver author certifying their implementation needs the conformance spec. A driver author starting from scratch needs the developing guide. These have different authority and different update frequency.

### Why `13-adrs/rejected/`?

Rejected design records prevent re-litigation. Without `REJ-001-go-plugin-package.md`, every new team member will re-propose Go plugins. The rejected section does not document why the idea was bad — it documents why the rejected design was considered, what trade-offs were evaluated, and why it was not chosen. Future contributors read it and either accept the ruling or bring new evidence that was not available when the ruling was made.

### Why `15-appendices/formal/`?

The formal invariants, filter grammar, and state machine diagrams are the closest thing Awo has to a mathematical specification. They are reference material for implementors, auditors, and future spec committee members. They do not belong in the tutorial (`08-app-dev/`), the reference (`12-reference/`), or the ADRs (`13-adrs/`). They are in their own directory so their authority is unambiguous.

---

## Part III — Per-File Specification

### docs/README.md

| Property | Value |
|---|---|
| **Purpose** | Navigation index. One paragraph per section. Who should read what. |
| **Audience** | Everyone (entry point) |
| **Prerequisites** | None |
| **Length** | 3–4 pages |
| **Dependencies** | All section READMEs |
| **Type** | Informative |
| **Priority** | Write last (after all sections exist) |
| **Stability** | Evolves (add new sections, update section descriptions) |

---

### docs/GLOSSARY.md

| Property | Value |
|---|---|
| **Purpose** | Canonical definitions of all terms used across all documentation. Every ambiguous term has exactly one definition here. |
| **Audience** | Everyone. Every other document links here on first use of a term. |
| **Prerequisites** | None |
| **Length** | 15–20 pages (100+ terms) |
| **Dependencies** | None |
| **Type** | Normative |
| **Priority** | Write first (before anything else) |
| **Stability** | Extends continuously (add terms), never redefines (existing definitions frozen) |

**Critical terms that must be defined:** Entity, EntityDefinition, System Entity, Custom Entity, Module, Capability Token, Compiled Schema, Compiler Phase, Driver, Conformance, Hook, Lifecycle Stage, Policy, Row Filter, Operation Permission, ViewerContext, SystemViewer, Record, SchemaVersion, Tenant, TenantContext, RLS, Outbox, Workflow, Activity, Idempotency Key, Filter, Filter Wire Format, SDUI, Abstract UI Schema, Renderer, Module Manifest, Architecture Law, Kernel, def/, Invariant, Normative, Informative.

---

### 01-introduction/philosophy.md

| Property | Value |
|---|---|
| **Purpose** | The five architectural axioms. Why Awo chooses constraints over flexibility. Philosophy before technology. |
| **Audience** | All developers, decision makers evaluating adoption |
| **Prerequisites** | None |
| **Length** | 8–10 pages |
| **Dependencies** | GLOSSARY.md |
| **Type** | Informative |
| **Priority** | Write second (after GLOSSARY) |
| **Stability** | Frozen after v1.0 (philosophy does not change) |

**Must cover:** Five axioms formal statement. Why each axiom exists. What each axiom prohibits. Historical precedents (Linux ABI stability, Git object model, PostgreSQL wire protocol). Why constraints produce long-lived software.

---

### 01-introduction/architecture-overview.md

| Property | Value |
|---|---|
| **Purpose** | Complete 10,000-foot view with diagrams. Five layers. All subsystems. How they connect. No implementation detail. |
| **Audience** | All developers (first architectural document to read) |
| **Prerequisites** | philosophy.md |
| **Length** | 12–15 pages |
| **Dependencies** | GLOSSARY.md, philosophy.md |
| **Type** | Informative |
| **Priority** | Third |
| **Stability** | Frozen after v1.0 (architecture is frozen) |

**Must include:** Five-layer diagram (UI → API → Domain → Workflow → Store). Startup sequence overview. EntityDefinition as the central primitive driving five subsystems. Package dependency diagram (simplified). The two-binary model (`cmd/server`, `cmd/migrate`).

---

### 01-introduction/comparison.md

| Property | Value |
|---|---|
| **Purpose** | Side-by-side comparison with Frappe, Odoo, Django, Spring, custom builds. Honest assessment of trade-offs. |
| **Audience** | Evaluators, architects considering adoption |
| **Prerequisites** | architecture-overview.md |
| **Length** | 10–12 pages |
| **Dependencies** | design-goals.md, non-goals.md |
| **Type** | Informative |
| **Priority** | Low (after all normative docs exist) |
| **Stability** | Evolves (competitors change, assessment updates) |

---

### 02-theory/compilation-theory.md

| Property | Value |
|---|---|
| **Purpose** | Why compiling metadata (rather than interpreting it at runtime) is the foundational architectural choice. Compiler theory applied to metadata. Trade-offs. |
| **Audience** | Framework contributors, advanced module authors, architects |
| **Prerequisites** | entity-model.md, metadata-model.md |
| **Length** | 15–20 pages |
| **Dependencies** | GLOSSARY.md, entity-model.md |
| **Type** | Informative |
| **Priority** | Before compiler specification (04) |
| **Stability** | Frozen after v1.0 |

**Must cover:** Why runtime interpretation fails at scale. What compilation guarantees. The ten compiler phases as a concept (not implementation). Determinism requirement. Incremental compilation theory. Cache correctness. Why the registry must freeze before compilation.

---

### 02-theory/tenancy-model.md

| Property | Value |
|---|---|
| **Purpose** | Multi-tenancy as a structural invariant, not a feature. How isolation is achieved at every layer. Why RLS is the enforcement point. |
| **Audience** | All developers |
| **Prerequisites** | entity-model.md |
| **Length** | 10–12 pages |
| **Dependencies** | GLOSSARY.md, security-model.md |
| **Type** | Informative |
| **Priority** | Before runtime docs (05) |
| **Stability** | Frozen after v1.0 |

---

### 02-theory/consistency-model.md

| Property | Value |
|---|---|
| **Purpose** | Formal consistency guarantees. What is strongly consistent. What is eventually consistent. What is explicitly undefined. Partition behavior. |
| **Audience** | System architects, operations engineers, database administrators |
| **Prerequisites** | storage-model.md, event-model.md |
| **Length** | 12–15 pages |
| **Dependencies** | workflow-model.md, storage-model.md |
| **Type** | Normative |
| **Priority** | Before operations docs (09) |
| **Stability** | Frozen after v1.0 (consistency guarantees are permanent) |

**Must document:** EntityRepository.Get() = strong consistency (primary). EntityRepository.Query() with replica option = eventual consistency. Workflow input data = snapshot consistency at trigger time. Session validation = strong consistency (Redis required). Outbox delivery = at-least-once. Cache = best-effort. Partition behavior table.

---

### 03-kernel/stability-policy.md

| Property | Value |
|---|---|
| **Purpose** | What "Kernel" stability means. What changes require a major version. What is permitted in minor versions. How deprecation works. The breaking change checklist. |
| **Audience** | All contributors, framework team, module authors |
| **Prerequisites** | GLOSSARY.md, philosophy.md |
| **Length** | 8–10 pages |
| **Dependencies** | versioning.md, breaking-change-policy.md |
| **Type** | Normative |
| **Priority** | Write before any kernel type documentation |
| **Stability** | Frozen at v1.0 (the policy is the promise) |

---

### 03-kernel/invariants.md

| Property | Value |
|---|---|
| **Purpose** | Formal statement of all 20 architecture laws. Rationale for each. Consequences of violation. How each is enforced (type system / compiler / runtime panic / CI check). |
| **Audience** | Framework contributors, auditors, core team |
| **Prerequisites** | All theory documents (02) |
| **Length** | 20–25 pages |
| **Dependencies** | All of 02-theory/, all kernel types (03/types/) |
| **Type** | Normative |
| **Priority** | Must exist before any contribution to core packages |
| **Stability** | Frozen at v1.0 (laws cannot change without major version + RFC) |

**This document is the constitution of the framework. It must be reviewed by every person who merges code to the core packages.**

---

### 03-kernel/types/entity-definition.md

| Property | Value |
|---|---|
| **Purpose** | Complete specification of EntityDefinition. Every field. Every constraint. Interaction with compiler. Which fields are schema-time vs. behavioral. Stability class of each field. |
| **Audience** | Module authors, compiler contributors, framework contributors |
| **Prerequisites** | all other types in 03/types/ |
| **Length** | 20–25 pages |
| **Dependencies** | field-def.md, edge-def.md, hook-registration.md, policy-func.md, action-def.md, workflow-trigger.md, ui-page.md |
| **Type** | Normative |
| **Priority** | Central — after all referenced types are documented |
| **Stability** | Frozen at v1.0 |

---

### 03-kernel/types/filter.md

| Property | Value |
|---|---|
| **Purpose** | Complete Filter DSL specification. All operators. Semantic contract for each operator. Wire format specification with version field. Composition rules. Evaluation model. |
| **Audience** | Driver authors, module authors, API client authors |
| **Prerequisites** | GLOSSARY.md, record.md |
| **Length** | 20–30 pages |
| **Dependencies** | record.md, field-def.md |
| **Type** | Normative |
| **Priority** | High — clients depend on wire format. Write before any client library exists. |
| **Stability** | Frozen at v1.0 (wire format permanent, semantics permanent) |

**Note:** This document must include the formal BNF grammar (also in `15-appendices/formal/filter-grammar.md`). It is the single most permanent public API document in the entire documentation set.

---

### 03-kernel/types/hook-registration.md

| Property | Value |
|---|---|
| **Purpose** | Complete specification of HookRegistration, LifecycleStage constants, Hook interface, per-stage hook interfaces, hook priority model, hook ownership semantics, HookPolicy. |
| **Audience** | Module authors, hook implementors |
| **Prerequisites** | record.md, viewer-context.md, lifecycle-stage.md |
| **Length** | 15–20 pages |
| **Dependencies** | lifecycle-stage.md, record.md |
| **Type** | Normative |
| **Priority** | High — every module author reads this |
| **Stability** | Frozen at v1.0 (stage names permanent) |

---

### 03-kernel/types/compiled-schema.md

| Property | Value |
|---|---|
| **Purpose** | Complete specification of CompiledSchema. Immutability guarantees. Content hash semantics. What is included in hash computation. Serialization interface. Version stamp. Cluster coordination role. |
| **Audience** | Framework contributors, compiler contributors, operators |
| **Prerequisites** | All entity types, compiler theory (02) |
| **Length** | 10–12 pages |
| **Dependencies** | entity-definition.md, filter.md |
| **Type** | Normative |
| **Priority** | After all input types are specified |
| **Stability** | Frozen at v1.0 |

---

### 04-compiler/phases/README.md

| Property | Value |
|---|---|
| **Purpose** | Overview of all ten compilation phases. What each phase produces. Input-output contract. Error handling. Phase ordering rationale. |
| **Audience** | Framework contributors, compiler contributors |
| **Prerequisites** | compilation-theory.md (02) |
| **Length** | 8–10 pages |
| **Dependencies** | All phase-specific documents |
| **Type** | Normative |
| **Priority** | Before individual phase docs |
| **Stability** | Evolves with compiler additions (phase additions are non-breaking) |

---

### 04-compiler/ir.md

| Property | Value |
|---|---|
| **Purpose** | Complete specification of the Intermediate Representation. All IR node types. Relationship between IR and EntityDefinition. Relationship between IR and compiled artifacts. When IR is constructed and when it is discarded. |
| **Audience** | Framework contributors, plugin pass authors |
| **Prerequisites** | All compiler phases, all entity types |
| **Length** | 20–25 pages |
| **Dependencies** | All of 04-compiler/phases/ |
| **Type** | Normative (internal — not public API, but specified for plugin authors) |
| **Priority** | After phase docs |
| **Stability** | Stability: Experimental at v1.0. Stabilizes at v1.2. |

---

### 04-compiler/plugin-passes.md

| Property | Value |
|---|---|
| **Purpose** | Complete specification of CompilerPass interface. How to write a custom compiler pass. What IR nodes are accessible. Diagnostic API. Error vs. warning semantics. Pass registration. Phase declaration. |
| **Audience** | Module authors writing compliance or validation passes |
| **Prerequisites** | ir.md, phases/README.md |
| **Length** | 12–15 pages |
| **Dependencies** | ir.md |
| **Type** | Normative |
| **Priority** | After ir.md |
| **Stability** | Stability: Experimental at v1.0 (IR is experimental). Frozen at v1.2. |

---

### 05-runtime/boot-sequence.md

| Property | Value |
|---|---|
| **Purpose** | Precise 12-phase startup sequence. What happens in each phase. What constitutes a fatal failure vs. degraded mode. Startup health check semantics. |
| **Audience** | Operations engineers, framework contributors |
| **Prerequisites** | architecture-overview.md |
| **Length** | 12–15 pages |
| **Dependencies** | 05-runtime/leader-election.md, all driver specs |
| **Type** | Normative |
| **Priority** | Early in runtime section |
| **Stability** | Frozen at v1.0 (startup contract is a deployment guarantee) |

---

### 05-runtime/transaction-boundaries.md

| Property | Value |
|---|---|
| **Purpose** | Precise specification of where transactions begin, where they end, what runs inside, what runs outside. Entity lifecycle transaction diagram. Outbox write within TX. Workflow dispatch outside TX. BulkCreate atomicity. |
| **Audience** | Module authors, hook authors, framework contributors |
| **Prerequisites** | entity-lifecycle.md, hook-execution.md |
| **Length** | 10–12 pages |
| **Dependencies** | workflow-dispatch.md, outbox-processing.md |
| **Type** | Normative |
| **Priority** | Critical — every hook author needs this |
| **Stability** | Frozen at v1.0 |

---

### 06-drivers/driver-contract.md

| Property | Value |
|---|---|
| **Purpose** | What a driver must guarantee beyond its interface methods. Isolation guarantee. Error semantics. Context propagation requirements. Tenant context verification. Connection lifecycle. |
| **Audience** | Driver authors |
| **Prerequisites** | consistency-model.md, tenancy-model.md |
| **Length** | 12–15 pages |
| **Dependencies** | 03-kernel/types/*, consistency-model.md |
| **Type** | Normative |
| **Priority** | First file in 06-drivers/ |
| **Stability** | Frozen at v1.0 |

---

### 06-drivers/optional-interfaces.md

| Property | Value |
|---|---|
| **Purpose** | The optional interface extension pattern. How to add capabilities to driver interfaces without breaking existing implementations. Type assertion at runtime. How the runtime detects optional capabilities. |
| **Audience** | Driver authors, framework contributors |
| **Prerequisites** | driver-contract.md |
| **Length** | 8–10 pages |
| **Dependencies** | driver-contract.md |
| **Type** | Normative |
| **Priority** | Second file in 06-drivers/ |
| **Stability** | Frozen at v1.0 (this is the permanent extension model) |

---

### 06-drivers/conformance/entity-store.md

| Property | Value |
|---|---|
| **Purpose** | Complete conformance test specification for EntityStore implementations. Every behavioral guarantee that must be verified. Test scenarios with expected inputs and outputs. |
| **Audience** | Driver authors, third-party storage vendors |
| **Prerequisites** | driver-contract.md, 03-kernel/types/filter.md |
| **Length** | 30–40 pages |
| **Dependencies** | All kernel type specs |
| **Type** | Normative |
| **Priority** | Before any third-party driver ships |
| **Stability** | Evolves (new test cases added, never removed) |

---

### 06-drivers/certification.md

| Property | Value |
|---|---|
| **Purpose** | The official certification process. What "Awo Certified" means. How to submit for certification. What the certification test suite covers. Certification maintenance requirements. |
| **Audience** | Driver authors, third-party vendors |
| **Prerequisites** | All conformance docs |
| **Length** | 8–10 pages |
| **Dependencies** | All of 06-drivers/conformance/ |
| **Type** | Informative |
| **Priority** | Before ecosystem launches |
| **Stability** | Evolves with certification process |

---

### 08-app-dev/getting-started.md

| Property | Value |
|---|---|
| **Purpose** | First entity in 15 minutes. Complete working example: define entity, fields, basic policy, register, run, observe auto-generated CRUD routes and UI. |
| **Audience** | New application developers |
| **Prerequisites** | architecture-overview.md (recommended), GLOSSARY.md |
| **Length** | 12–15 pages |
| **Dependencies** | entities/README.md, fields/field-types.md, policies/operation-permissions.md |
| **Type** | Informative |
| **Priority** | First app-dev document written |
| **Stability** | Evolves with API changes |

---

### 08-app-dev/hooks/hook-ordering.md

| Property | Value |
|---|---|
| **Purpose** | How hook execution order is determined across modules. Topological sort from module dependency graph. Priority within a single module. Determinism guarantee. How to declare ordering dependencies. |
| **Audience** | Module authors |
| **Prerequisites** | hook-lifecycle.md, modules/dependencies.md |
| **Length** | 8–10 pages |
| **Dependencies** | 03-kernel/types/hook-registration.md, modules/manifests.md |
| **Type** | Normative |
| **Priority** | After basic hook docs |
| **Stability** | Frozen at v1.0 |

---

### 08-app-dev/entities/entity-lifecycle.md

| Property | Value |
|---|---|
| **Purpose** | Active → Deprecated → Retired → Archived lifecycle. How to declare each stage. What the compiler generates for each stage. What happens to routes, writes, reads at each stage. Migration path. |
| **Audience** | Module authors, long-term maintainers |
| **Prerequisites** | system-entities.md, custom-entities.md |
| **Length** | 10–12 pages |
| **Dependencies** | 03-kernel/types/entity-definition.md |
| **Type** | Normative |
| **Priority** | After basic entity docs |
| **Stability** | Frozen at v1.0 |

---

### 08-app-dev/fields/field-lifecycle.md

| Property | Value |
|---|---|
| **Purpose** | Field Active → Deprecated → Retired → Archived lifecycle. RetiredAt, ArchivedAt semantics. How the compiler generates behavior changes per stage. Safe field removal procedure. |
| **Audience** | Module authors |
| **Prerequisites** | field-types.md, constraints.md |
| **Length** | 8–10 pages |
| **Dependencies** | 03-kernel/types/field-def.md, 09-operations/migrations/ |
| **Type** | Normative |
| **Priority** | After basic field docs |
| **Stability** | Frozen at v1.0 |

---

### 08-app-dev/modules/manifests.md

| Property | Value |
|---|---|
| **Purpose** | Complete ModuleManifest specification from the application developer perspective. Every field. Capability token format. Version constraint syntax. Module signing. AuditRequired flag. |
| **Audience** | Module authors, ecosystem developers |
| **Prerequisites** | modules/module-structure.md, capabilities.md |
| **Length** | 15–20 pages |
| **Dependencies** | 03-kernel/types/module-manifest.md, capabilities.md |
| **Type** | Normative |
| **Priority** | Before any module ships to ecosystem |
| **Stability** | Frozen at v1.0 (manifest format is permanent) |

---

### 09-operations/migrations/zero-downtime-patterns.md

| Property | Value |
|---|---|
| **Purpose** | Concrete, tested patterns for zero-downtime schema changes. Add column, backfill, constrain. Rename with dual-write. Index creation. RLS policy addition. Each with exact SQL, timing guidance, rollback procedure. |
| **Audience** | Operations engineers, DBAs |
| **Prerequisites** | running-migrations.md |
| **Length** | 20–25 pages |
| **Dependencies** | rolling-upgrades.md |
| **Type** | Informative |
| **Priority** | Before first production deployment |
| **Stability** | Evolves (new patterns added) |

---

### 09-operations/disaster-recovery/partition-behavior.md

| Property | Value |
|---|---|
| **Purpose** | Formal table: each component, what behavior changes when it is unavailable. PostgreSQL down, Redis down, Temporal down, outbox processor down. Expected HTTP responses. Data safety guarantees. |
| **Audience** | Operations engineers, SRE teams |
| **Prerequisites** | consistency-model.md (02) |
| **Length** | 8–10 pages |
| **Dependencies** | consistency-model.md |
| **Type** | Normative |
| **Priority** | Before production deployment |
| **Stability** | Frozen at v1.0 (partition behavior is a contract) |

---

### 11-security/threat-model.md

| Property | Value |
|---|---|
| **Purpose** | Formal threat model. Assets at risk. Threat actors. Attack vectors. Mitigations at the architecture level. What the architecture does not protect against (scope). |
| **Audience** | Security teams, auditors, CTOs |
| **Prerequisites** | trust-boundaries.md |
| **Length** | 20–25 pages |
| **Dependencies** | trust-boundaries.md, supply-chain.md, row-level-security.md |
| **Type** | Normative |
| **Priority** | Before regulated industry deployment |
| **Stability** | Evolves (threat landscape changes) |

---

### 11-security/supply-chain.md

| Property | Value |
|---|---|
| **Purpose** | Supply chain attack surface. What module signing provides. What it does not provide. Trust model for third-party modules. SBOM. Reproducible builds. Future WASM isolation. |
| **Audience** | Security teams, operators |
| **Prerequisites** | trust-boundaries.md |
| **Length** | 10–12 pages |
| **Dependencies** | 08-app-dev/modules/manifests.md |
| **Type** | Normative |
| **Priority** | Before ecosystem launches |
| **Stability** | Evolves (WASM isolation adds new content at v2) |

---

### 12-reference/filter-dsl.md

| Property | Value |
|---|---|
| **Purpose** | Complete Filter DSL reference. Every operator with name, semantics, Go constructor, wire format, SQL translation for PostgreSQL driver, error cases. Version field specification. BNF grammar link. |
| **Audience** | API client authors, driver authors, module authors |
| **Prerequisites** | 03-kernel/types/filter.md |
| **Length** | 25–30 pages |
| **Dependencies** | 03-kernel/types/filter.md, 15-appendices/formal/filter-grammar.md |
| **Type** | Normative |
| **Priority** | High — client library authors need this |
| **Stability** | Frozen at v1.0 (operator semantics permanent). New operators additive (new pages, never modify existing) |

---

### 12-reference/error-catalog.md

| Property | Value |
|---|---|
| **Purpose** | Every error code the framework produces. HTTP status. Machine-readable code. User-facing message template. Causes. Resolution steps. |
| **Audience** | API client authors, operations engineers, support teams |
| **Prerequisites** | None (reference document) |
| **Length** | 30–40 pages |
| **Dependencies** | Partially auto-generated from error code constants |
| **Type** | Normative |
| **Priority** | Before API clients are built |
| **Stability** | Extends (new codes added). Existing codes never change HTTP status or machine code. |

---

### 12-reference/reserved-names.md

| Property | Value |
|---|---|
| **Purpose** | All reserved entity names, field names, route prefixes, capability token namespaces, lifecycle stage names, event type names. Why each is reserved. |
| **Audience** | Module authors |
| **Prerequisites** | GLOSSARY.md |
| **Length** | 8–10 pages |
| **Dependencies** | 03-kernel/types/entity-definition.md, 03-kernel/types/lifecycle-stage.md |
| **Type** | Normative |
| **Priority** | Before module ecosystem launches |
| **Stability** | Frozen (reserved names cannot become unreserved) |

---

### 13-adrs/ADR-001-metadata-first-architecture.md

| Property | Value |
|---|---|
| **Purpose** | Why metadata-first was chosen. Alternatives considered (code generation, convention-based, annotation-based). Trade-offs. Why this decision is permanent. |
| **Audience** | Future core team, framework architects |
| **Prerequisites** | None |
| **Length** | 8–10 pages |
| **Dependencies** | None |
| **Type** | Informative (historical record) |
| **Priority** | Write before any other ADR |
| **Stability** | Frozen forever (ADRs are immutable historical records) |

---

### 13-adrs/rejected/REJ-006-hookset-struct.md

| Property | Value |
|---|---|
| **Purpose** | Why HookSet-as-struct was rejected. The extensibility problem. What the open HookRegistration model provides. Why this was a v1.0 blocker. |
| **Audience** | Future contributors who might re-propose struct-based hooks |
| **Prerequisites** | ADR-011 |
| **Length** | 5–6 pages |
| **Dependencies** | ADR-011-hook-registration-model.md |
| **Type** | Informative (historical record) |
| **Priority** | Low |
| **Stability** | Frozen forever |

---

### 15-appendices/formal/architecture-laws.md

| Property | Value |
|---|---|
| **Purpose** | Formal statement of all 20 architecture laws. For each law: formal statement, rationale, enforcement mechanism, consequences of violation, who is responsible for enforcement. |
| **Audience** | Core team, auditors, specification committee |
| **Prerequisites** | All of 03-kernel/, 05-runtime/ |
| **Length** | 25–30 pages |
| **Dependencies** | 03-kernel/invariants.md (this is the authoritative version) |
| **Type** | Normative |
| **Priority** | After all kernel specs |
| **Stability** | Frozen at v1.0. Law additions require RFC + major version. |

---

### 15-appendices/formal/filter-grammar.md

| Property | Value |
|---|---|
| **Purpose** | Formal BNF grammar of the Filter DSL wire format. Complete, machine-readable, version-stamped. |
| **Audience** | Client library authors, compiler authors, parser implementors |
| **Prerequisites** | 03-kernel/types/filter.md |
| **Length** | 5–8 pages |
| **Dependencies** | 03-kernel/types/filter.md |
| **Type** | Normative |
| **Priority** | Before any client library is built |
| **Stability** | Frozen at v1.0 |

---

### 15-appendices/formal/state-machines.md

| Property | Value |
|---|---|
| **Purpose** | Formal state machine diagrams for: tenant lifecycle (PENDING→ACTIVE→SUSPENDED→ARCHIVED), entity lifecycle (ACTIVE→DEPRECATED→RETIRED→ARCHIVED), field lifecycle (same), hook lifecycle (registered→compiled→executing→completed/failed). |
| **Audience** | Framework contributors, auditors |
| **Prerequisites** | All theory docs (02) |
| **Length** | 12–15 pages |
| **Dependencies** | 02-theory/*, 08-app-dev/entities/entity-lifecycle.md |
| **Type** | Normative |
| **Priority** | After theory section complete |
| **Stability** | Frozen at v1.0 |

---

### 15-appendices/anti-patterns/README.md

| Property | Value |
|---|---|
| **Purpose** | Why anti-patterns are documented. How to use this section. How an anti-pattern is nominated and accepted. |
| **Audience** | All developers |
| **Prerequisites** | None |
| **Length** | 2–3 pages |
| **Dependencies** | None |
| **Type** | Informative |
| **Priority** | Low |
| **Stability** | Evolves |

---

## Part IV — Auto-Generated Documents

These files must be generated by tooling from source code, not hand-written. Generated content should be clearly marked.

| File | Generator | Source | Regenerate trigger |
|---|---|---|---|
| `12-reference/field-types.md` | `awo-docgen field-types` | FieldType constants + registered handlers | On any FieldType change |
| `12-reference/error-catalog.md` (partial) | `awo-docgen errors` | Error code constants + message templates | On any error constant change |
| `12-reference/environment-variables.md` | `awo-docgen config` | Config struct env tags | On Config struct change |
| `12-reference/http-api.md` (partial) | `awo-docgen routes` | Compiled route table | On any EntityDefinition change |
| `12-reference/event-types.md` | `awo-docgen events` | TriggerEvent constants | On TriggerEvent constant change |
| `12-reference/lifecycle-stages.md` | `awo-docgen stages` | LifecycleStage constants | On LifecycleStage constant change |
| `15-appendices/formal/filter-grammar.md` | `awo-docgen filter-grammar` | Filter type definitions | On Filter type change |
| `15-appendices/diagrams/*.svg` | `awo-docgen diagrams` | Package dependency analysis | On package structure change |

**Generator naming:** All generators live in `cmd/awo-docgen/`. Running `awo-docgen all` regenerates all auto-generated files. CI rejects PRs where generated files are out of sync with source.

---

## Part V — Stability Classification

### Frozen at v1.0 — Never change

These files document permanent decisions. Changing them requires a major version + RFC + ecosystem migration:

```
docs/GLOSSARY.md (existing definitions only)
docs/01-introduction/philosophy.md
docs/01-introduction/architecture-overview.md
docs/02-theory/compilation-theory.md
docs/02-theory/tenancy-model.md
docs/02-theory/consistency-model.md
docs/02-theory/security-model.md
docs/03-kernel/**                          (entire section)
docs/04-compiler/phases/*                  (phase ordering and I/O contracts)
docs/05-runtime/transaction-boundaries.md
docs/05-runtime/boot-sequence.md
docs/06-drivers/driver-contract.md
docs/06-drivers/optional-interfaces.md
docs/09-operations/disaster-recovery/partition-behavior.md
docs/12-reference/filter-dsl.md            (existing operators)
docs/12-reference/reserved-names.md
docs/13-adrs/**                            (all ADRs — immutable historical)
docs/15-appendices/formal/**
```

### Evolves — New content only (existing content frozen)

Adding without modifying:

```
docs/GLOSSARY.md (new terms can be added)
docs/12-reference/error-catalog.md
docs/12-reference/field-types.md
docs/12-reference/filter-dsl.md (new operators only)
docs/06-drivers/conformance/** (new test cases, never removed)
docs/13-adrs/** (new ADRs added, existing frozen)
```

### Evolves freely

Updated with new minor/patch versions:

```
docs/08-app-dev/**
docs/09-operations/**
docs/10-diagnostics/**
docs/11-security/threat-model.md
docs/14-contributing/**
docs/15-appendices/examples/**
docs/15-appendices/patterns/**
docs/15-appendices/anti-patterns/**
docs/CHANGELOG.md
```

---

## Part VI — Writing Order

Strict dependency order. Each phase cannot begin until the prior phase is complete.

**Phase 0 — Foundation (write before everything else)**

1. `docs/GLOSSARY.md` — all other documents reference it
2. `01-introduction/philosophy.md` — all other documents assume it
3. `01-introduction/architecture-overview.md` — orients all readers

**Phase 1 — Theory (must exist before normative specs)**

4. `02-theory/entity-model.md`
5. `02-theory/metadata-model.md`
6. `02-theory/compilation-theory.md`
7. `02-theory/tenancy-model.md`
8. `02-theory/security-model.md`
9. `02-theory/consistency-model.md`
10. `02-theory/storage-model.md`
11. `02-theory/policy-model.md`
12. `02-theory/workflow-model.md`
13. `02-theory/event-model.md`
14. `02-theory/sdui-model.md`
15. `02-theory/failure-philosophy.md`

**Phase 2 — Kernel Specification (normative core)**

16. `03-kernel/stability-policy.md`
17. `03-kernel/versioning.md`
18. `03-kernel/breaking-change-policy.md`
19. `03-kernel/types/record.md`
20. `03-kernel/types/viewer-context.md`
21. `03-kernel/types/filter.md`  ← highest priority type document
22. `03-kernel/types/field-def.md`
23. `03-kernel/types/field-type.md`
24. `03-kernel/types/edge-def.md`
25. `03-kernel/types/lifecycle-stage.md`
26. `03-kernel/types/hook-registration.md`
27. `03-kernel/types/policy-func.md`
28. `03-kernel/types/policy-result.md`
29. `03-kernel/types/action-def.md`
30. `03-kernel/types/trigger-event.md`
31. `03-kernel/types/workflow-trigger.md`
32. `03-kernel/types/create-input.md`
33. `03-kernel/types/update-input.md`
34. `03-kernel/types/page-info.md`
35. `03-kernel/types/aggregate-spec.md`
36. `03-kernel/types/aggregate-result.md`
37. `03-kernel/types/ui-page.md`
38. `03-kernel/types/module-manifest.md`
39. `03-kernel/types/capability-token.md`
40. `03-kernel/types/entity-definition.md`  ← depends on all above
41. `03-kernel/types/compiled-schema.md`
42. `03-kernel/invariants.md`  ← depends on all types

**Phase 3 — Compiler and Runtime Specification**

43. `04-compiler/phases/README.md`
44. `04-compiler/phases/01-registration.md` through `10-schema-seal.md`
45. `04-compiler/ir.md`
46. `04-compiler/dependency-graph.md`
47. `04-compiler/validation-passes.md`
48. `04-compiler/plugin-passes.md`
49. `04-compiler/fingerprinting.md`
50. `04-compiler/generated-artifacts.md`
51. `04-compiler/incremental-compilation.md`
52. `04-compiler/diagnostics.md`
53. `05-runtime/boot-sequence.md`
54. `05-runtime/middleware-pipeline.md`
55. `05-runtime/entity-lifecycle.md`
56. `05-runtime/transaction-boundaries.md`
57. `05-runtime/hook-execution.md`
58. `05-runtime/policy-evaluation.md`
59. `05-runtime/workflow-dispatch.md`
60. `05-runtime/outbox-processing.md`
61. `05-runtime/request-lifecycle.md`
62. `05-runtime/graceful-shutdown.md`

**Phase 4 — Driver Specification**

63. `06-drivers/driver-contract.md`
64. `06-drivers/optional-interfaces.md`
65. `06-drivers/conformance/entity-store.md`  ← most important
66. `06-drivers/conformance/session-store.md`
67. `06-drivers/conformance/workflow-driver.md`
68. (remaining conformance docs)
69. `06-drivers/developing/entity-store.md`
70. (remaining developing docs)
71. `06-drivers/certification.md`

**Phase 5 — Formal Appendices**

72. `15-appendices/formal/architecture-laws.md`
73. `15-appendices/formal/filter-grammar.md`
74. `15-appendices/formal/state-machines.md`

**Phase 6 — Application Developer Guides**

75. `08-app-dev/getting-started.md`
76. `08-app-dev/entities/` (all files)
77. `08-app-dev/fields/` (all files)
78. `08-app-dev/relationships/` (all files)
79. `08-app-dev/policies/` (all files)
80. `08-app-dev/hooks/` (all files)
81. `08-app-dev/actions/` (all files)
82. `08-app-dev/workflows/` (all files)
83. `08-app-dev/ui/` (all files)
84. `08-app-dev/modules/` (all files)
85. `08-app-dev/testing/` (all files)

**Phase 7 — Operations, Security, Diagnostics**

86. `11-security/trust-boundaries.md`
87. `11-security/threat-model.md`
88. `11-security/row-level-security.md`
89. (remaining security docs)
90. `09-operations/` (all files)
91. `10-diagnostics/` (all files)

**Phase 8 — Reference (partially auto-generated)**

92. `12-reference/glossary.md` (extracted from GLOSSARY.md)
93. `12-reference/filter-dsl.md`
94. `12-reference/error-catalog.md`
95. (remaining reference docs — run generators)

**Phase 9 — ADRs**

96. Write all ADRs, starting with ADR-001
97. Write all rejected design docs
98. Write framework-dev guide

**Phase 10 — Final**

99. Examples, patterns, anti-patterns
100. `docs/README.md` (last — now all sections exist to summarize)
101. Diagrams (can be generated alongside)

---

## Part VII — Size Estimates

| Section | Files | Est. Pages | Est. Words |
|---|---|---|---|
| 01-introduction | 10 | 80 | 32,000 |
| 02-theory | 15 | 150 | 60,000 |
| 03-kernel | 30 | 300 | 120,000 |
| 04-compiler | 18 | 150 | 60,000 |
| 05-runtime | 15 | 120 | 48,000 |
| 06-drivers | 22 | 200 | 80,000 |
| 07-framework-dev | 11 | 80 | 32,000 |
| 08-app-dev | 35 | 280 | 112,000 |
| 09-operations | 18 | 140 | 56,000 |
| 10-diagnostics | 12 | 80 | 32,000 |
| 11-security | 13 | 100 | 40,000 |
| 12-reference | 13 | 150 | 60,000 |
| 13-adrs | 25 | 180 | 72,000 |
| 14-contributing | 9 | 60 | 24,000 |
| 15-appendices | 30+ | 200 | 80,000 |
| **Total** | **~276** | **~2,270** | **~908,000** |

Practical comparison:
- PostgreSQL documentation: ~3,000 pages
- Kubernetes documentation: ~2,500 pages
- Django documentation: ~1,500 pages
- **Awo target: ~2,300 pages** — appropriate for the scope

At one experienced technical writer producing 2,000 words/day of publication-quality documentation, Phase 0–5 (normative core, ~360,000 words) takes approximately 6 months. Full documentation corpus takes approximately 18 months.

---

## Part VIII — Self-Critique of This Blueprint

### Weaknesses Identified

**Weakness 1 — API client documentation is absent.**

There is no `16-api-clients/` section covering REST API consumers, Python/TypeScript/Java client libraries, OpenAPI schema, and API authentication. Client developers — who may never read Go code — need a standalone section. This is a significant gap if Awo supports external API consumers.

**Weakness 2 — Upgrade guides between major versions are not planned.**

The blueprint assumes v1.0 documentation. There is no `UPGRADING.md` or upgrade guide section. When v2.0 is released, every breaking change from v1.0 must be documented as a concrete migration step. A `migration/` directory should be added to `13-adrs/` or as a standalone section.

**Weakness 3 — Localization is not addressed.**

Awo targets Kenyan deployments (EAT timezone, KES currency). If the framework is used internationally, documentation may need localization. The blueprint has no localization strategy. At minimum, the documentation should flag which sections contain locale-specific assumptions.

**Weakness 4 — Security documentation assumes reader familiarity with threat modeling.**

`11-security/threat-model.md` as written is useful to experienced security engineers but opaque to developers who are new to threat modeling. A companion `11-security/security-primer.md` should explain threat modeling concepts before presenting Awo's specific model.

**Weakness 5 — No documentation for the platform modules.**

The seven built-in platform modules (Tenant, IAM, Feature Flags, Settings, Audit Log, Metadata, Module Registry) are not documented as a group. They are documented implicitly through the theory and app-dev sections. There should be a `platform/` section (or `08-app-dev/platform-modules/`) that documents each platform module's behavior, API, and configuration. These are the first things a new application developer encounters.

**Weakness 6 — Examples directory is underspecified.**

`15-appendices/examples/` has three examples (invoice, CRM contact, approval workflow). These are too few. A production-quality example set needs: a complete module (Finance with ledger, invoice, payment), a healthcare module example, a government module example, a module that integrates with an external API via activity. The examples section should be treated as a first-class product, not an afterthought.

**Weakness 7 — Documentation for emergency procedures is missing.**

`09-operations/` covers steady-state operations. But what does an SRE do at 2 AM when the outbox processor is stuck? Each background process should have an emergency runbook: indicators, immediate mitigations, escalation path. These belong in `09-operations/` but were not listed.

**Weakness 8 — No changelog format is specified.**

`docs/CHANGELOG.md` is listed but its format is not specified. Changelogs must distinguish: breaking changes, new features, bug fixes, deprecations, removed items. The format must be machine-readable (semantic release tooling) and human-readable. The style guide (`14-contributing/documentation-style-guide.md`) must specify this.

**Weakness 9 — The 07-framework-dev/ section is too internal.**

Framework developer documentation is critical for contributor velocity but is underspecified in this blueprint. The layer rules and import rules documents need concrete examples of violations and their consequences — not just statements of the rules. The reviewer checklist needs enough specificity to catch the architectural violations described in the v4 review.

**Weakness 10 — No offline documentation packaging strategy.**

The user requirement states "learn the framework entirely offline." But 276 markdown files with inter-document links do not read well offline without a static site generator. The documentation build must produce a single-file offline archive (e.g., a compiled PDF or offline HTML site). This build process belongs in `14-contributing/ci-cd.md` but is not explicitly planned.

### Missing Sections (Recommend Adding)

```
docs/
├── 16-api-clients/              (REST API consumer documentation)
│   ├── README.md
│   ├── authentication.md
│   ├── filter-syntax.md
│   ├── pagination.md
│   ├── error-handling.md
│   └── openapi-schema.md
│
├── 17-platform-modules/         (built-in platform module reference)
│   ├── README.md
│   ├── tenant-module.md
│   ├── iam-module.md
│   ├── feature-flags.md
│   ├── settings.md
│   ├── audit-log.md
│   ├── metadata.md
│   └── module-registry.md
│
└── UPGRADING.md                 (version migration guide, appended per release)
```

**Revised total with additions: ~310 files, ~2,600 pages, ~1,040,000 words.**

---

## Part IX — Final Recommendations

**Three documents that determine the quality of everything else:**

1. `docs/GLOSSARY.md` — If terms are imprecise here, every document that uses them inherits the imprecision. Write this with the same rigor as an RFC.

2. `03-kernel/invariants.md` — If invariants are incomplete here, framework contributors will violate them. Every invariant must have an enforcement mechanism, not just a statement.

3. `15-appendices/formal/filter-grammar.md` — The Filter DSL is the single most permanent public API. If the grammar is wrong, it is wrong forever.

**One decision before writing begins:**

Choose the documentation tooling before writing any content. Markdown files that will be rendered by `mdBook`, `Docusaurus`, `Hugo`, or `MkDocs` have slightly different cross-linking and formatting conventions. Writing 276 files and then changing the tool means reformatting all of them. The tool choice is an infrastructure decision, not a content decision, but it must be made first.

**One organizational principle for the writing team:**

No section should be written by the same person who designed the thing it documents. The writer must read the source, ask questions, and document what actually exists — not what was intended. Architectural design documents and specification documents written by the same author produce idealized descriptions of what should be, not accurate descriptions of what is.

*Signed: Chief Technical Writer*
*Date: 2025-07-06*
*Status: Blueprint complete. Awaiting tooling decision and team assignment before writing begins.*
