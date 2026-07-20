> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

### Chapter 1 — Framework Foundations & Mental Model


<!-- toc -->

  - [Chapter 1 — Framework Foundations & Mental Model](#chapter-1--framework-foundations--mental-model)
    - [1.1. What Awo Framework Is](#11-what-awo-framework-is)
      - [1.1.1. A Go-native platform for building multi-tenant ERP systems](#111-a-go-native-platform-for-building-multi-tenant-erp-systems)
      - [1.1.2. What "framework" means here vs "library" vs "application"](#112-what-framework-means-here-vs-library-vs-application)
      - [1.1.3. What Awo is not — anti-patterns explicitly out of scope](#113-what-awo-is-not--anti-patterns-explicitly-out-of-scope)
      - [1.1.4. Relationship to ERPNext and Frappe — conceptual debts and deliberate departures](#114-relationship-to-erpnext-and-frappe--conceptual-debts-and-deliberate-departures)
    - [1.2. Design Philosophy](#12-design-philosophy)
      - [1.2.1. EntityDefinition as the central primitive — not just a schema, a dispatch key](#121-entitydefinition-as-the-central-primitive--not-just-a-schema-a-dispatch-key)
      - [1.2.2. Compile-time correctness for system entities, runtime flexibility for custom entities](#122-compile-time-correctness-for-system-entities-runtime-flexibility-for-custom-entities)
      - [1.2.3. The hybrid data model rationale — why not everything in JSONB, why not everything in typed tables](#123-the-hybrid-data-model-rationale--why-not-everything-in-jsonb-why-not-everything-in-typed-tables)
      - [1.2.4. Interface-first persistence — the `EntityRepository` contract and why the implementation is replaceable](#124-interface-first-persistence--the-entityrepository-contract-and-why-the-implementation-is-replaceable)
      - [1.2.5. Workflow-first ERP — why business processes belong in Temporal, not in request handlers](#125-workflow-first-erp--why-business-processes-belong-in-temporal-not-in-request-handlers)
      - [1.2.6. Multi-tenancy as a first-class concern, not a plugin](#126-multi-tenancy-as-a-first-class-concern-not-a-plugin)
      - [1.2.7. Server-driven UI as a force multiplier — eliminating the frontend bottleneck](#127-server-driven-ui-as-a-force-multiplier--eliminating-the-frontend-bottleneck)
    - [1.3. Who This Documentation Is For](#13-who-this-documentation-is-for)
      - [1.3.1. Framework contributors — building Awo core](#131-framework-contributors--building-awo-core)
      - [1.3.2. Module developers — building built-in ERP modules on top of Awo](#132-module-developers--building-built-in-erp-modules-on-top-of-awo)
      - [1.3.3. Application developers — building tenant-specific customisations](#133-application-developers--building-tenant-specific-customisations)
      - [1.3.4. System integrators — deploying and operating Awo for a client](#134-system-integrators--deploying-and-operating-awo-for-a-client)
      - [1.3.5. How to navigate this documentation by audience](#135-how-to-navigate-this-documentation-by-audience)
    - [1.4. Awo vs Frappe — Direct Comparison](#14-awo-vs-frappe--direct-comparison)
      - [1.4.1. DocType vs EntityDefinition](#141-doctype-vs-entitydefinition)
      - [1.4.2. Python hooks vs Go interfaces](#142-python-hooks-vs-go-interfaces)
      - [1.4.3. MariaDB-first vs PostgreSQL-first](#143-mariadb-first-vs-postgresql-first)
      - [1.4.4. Frappe Workflow vs Temporal](#144-frappe-workflow-vs-temporal)
      - [1.4.5. Jinja forms vs amis SDUI](#145-jinja-forms-vs-amis-sdui)
      - [1.4.6. Migration strategy for teams coming from Frappe](#146-migration-strategy-for-teams-coming-from-frappe)
  - [Chapter 2 — The EntityDefinition — The Central Abstraction](#chapter-2--the-entitydefinition--the-central-abstraction)
    - [2.1. Why EntityDefinition Is the Central Primitive](#21-why-entitydefinition-is-the-central-primitive)
      - [2.1.1. What the framework does when it sees an EntityDefinition](#211-what-the-framework-does-when-it-sees-an-entitydefinition)
      - [2.1.2. The five things an EntityDefinition drives — persistence routing, API generation, UI generation, permission evaluation, workflow triggering](#212-the-five-things-an-entitydefinition-drives--persistence-routing-api-generation-ui-generation-permission-evaluation-workflow-triggering)
      - [2.1.3. Why this is not just a schema definition tool](#213-why-this-is-not-just-a-schema-definition-tool)
    - [2.2. The Two Entity Types](#22-the-two-entity-types)
      - [2.2.1. System entities — strongly typed, SQL-backed, SQLC-generated or ent-generated queries](#221-system-entities--strongly-typed-sql-backed-sqlc-generated-or-ent-generated-queries)
      - [2.2.2. Custom entities — tenant-defined, JSONB-backed, metadata-driven](#222-custom-entities--tenant-defined-jsonb-backed-metadata-driven)
      - [2.2.3. The EntityResolver — how the framework picks the execution path at runtime](#223-the-entityresolver--how-the-framework-picks-the-execution-path-at-runtime)
      - [2.2.4. What callers see — the EntityRecord as a unified surface regardless of entity type](#224-what-callers-see--the-entityrecord-as-a-unified-surface-regardless-of-entity-type)
      - [2.2.5. Why this distinction is invisible to framework consumers by design](#225-why-this-distinction-is-invisible-to-framework-consumers-by-design)
    - [2.3. When to Use Each Entity Type](#23-when-to-use-each-entity-type)
      - [2.3.1. Use a system entity when — financial integrity, inventory accuracy, IAM, high-frequency writes](#231-use-a-system-entity-when--financial-integrity-inventory-accuracy-iam-high-frequency-writes)
      - [2.3.2. Use a custom entity when — tenant-specific extensions, industry-specific fields, rapid configuration without deploys](#232-use-a-custom-entity-when--tenant-specific-extensions-industry-specific-fields-rapid-configuration-without-deploys)
      - [2.3.3. Entities that must never be custom — LedgerEntry, StockMove, Payment, User, Tenant](#233-entities-that-must-never-be-custom--ledgerentry-stockmove-payment-user-tenant)
      - [2.3.4. Entities that should always be custom — tenant contact extensions, custom approval metadata, locale-specific fields](#234-entities-that-should-always-be-custom--tenant-contact-extensions-custom-approval-metadata-locale-specific-fields)
      - [2.3.5. The grey zone — when to escalate a custom entity to a system entity](#235-the-grey-zone--when-to-escalate-a-custom-entity-to-a-system-entity)
    - [2.4. EntityDefinition Anatomy](#24-entitydefinition-anatomy)
      - [2.4.1. Name and identifier conventions](#241-name-and-identifier-conventions)
      - [2.4.2. Field list — types, constraints, metadata](#242-field-list--types-constraints-metadata)
      - [2.4.3. Edge declarations — relationships to other entities](#243-edge-declarations--relationships-to-other-entities)
      - [2.4.4. Hook registrations](#244-hook-registrations)
      - [2.4.5. Permission policy bindings](#245-permission-policy-bindings)
      - [2.4.6. Workflow trigger bindings](#246-workflow-trigger-bindings)
      - [2.4.7. UI page builder bindings](#247-ui-page-builder-bindings)
      - [2.4.8. Naming series configuration](#248-naming-series-configuration)
    - [2.5. The EntityRegistry](#25-the-entityregistry)
      - [2.5.1. Global registry — what it holds and when it is populated](#251-global-registry--what-it-holds-and-when-it-is-populated)
      - [2.5.2. System entity registration at compile time](#252-system-entity-registration-at-compile-time)
      - [2.5.3. Custom entity loading at tenant boot](#253-custom-entity-loading-at-tenant-boot)
      - [2.5.4. Registry lookup — by name, by tenant, by type](#254-registry-lookup--by-name-by-tenant-by-type)
      - [2.5.5. Registry concurrency model — reads vs writes during tenant boot](#255-registry-concurrency-model--reads-vs-writes-during-tenant-boot)
    - [Chapter summary](#chapter-summary)

<!-- tocstop -->

#### 1.1. What Awo Framework Is

##### 1.1.1. A Go-native platform for building multi-tenant ERP systems

Awo is a Go framework for building production-grade, multi-tenant ERP systems. It provides the structural backbone — entity modelling, persistence routing, permission evaluation, workflow orchestration, and server-driven UI generation — so that module developers can focus on business logic rather than infrastructure plumbing. The framework is opinionated about its core architectural decisions and permissive about the business domains built on top of it.

The primary deployment target is the East African market, with Kenya as the first-class locale: KES currency formatting, EAT timezone defaults, KRA eTIMS integration hooks, and Swahili error message catalogues are built-in concerns rather than afterthoughts. This specificity is intentional. Generalising too early produces frameworks that serve no market well; Awo earns the right to generalise by being excellent in one place first.

The Go runtime was chosen for its predictable latency characteristics under concurrent load, its straightforward deployment model (a single static binary), and its type system, which allows the framework to enforce architectural invariants at compile time for system entities while remaining flexible for tenant-defined custom entities at runtime.

##### 1.1.2. What "framework" means here vs "library" vs "application"

A library is a collection of functions a developer calls. A framework is a structure into which a developer plugs their code — the framework calls the developer's code, not the other way around. Awo is a framework in this strict sense. When you define an `EntityDefinition`, the framework consumes it: routing, permission gates, page builders, and workflow triggers are all derived from that definition automatically. Your code fills in the hooks, validators, and business rules; the framework orchestrates when and how they are called.

This inversion-of-control model means the framework owns the request lifecycle, the transaction boundary, and the workflow trigger sequence. Module developers write focused, testable functions — a `before_save` hook, a Temporal activity, a page builder — and register them against an `EntityDefinition`. The framework guarantees their execution order and provides the correct context at each point.

Awo is not the ERP application itself. It is the platform on which ERP applications are built. The built-in modules (Finance, Forecourt, HR, Inventory, CRM) are first-class consumers of the framework, written against the same public API available to any module developer.

##### 1.1.3. What Awo is not — anti-patterns explicitly out of scope

Awo is not a code generator. It does not produce a Go codebase that you then own and modify directly. The framework remains a dependency; your modules import it. This distinction matters because it means framework improvements flow to all consumers via a version upgrade, not a one-time code generation run.

Awo is not a low-code platform with a GUI configuration wizard. The `EntityDefinition` is defined in Go code, compiled into the binary, and versioned in your repository. Tenant-defined custom entities are a runtime extension mechanism — see §2.2 — but the framework's own system entities require a code change and a migration.

Awo does not manage its own frontend build pipeline. The SDUI contract means the server emits amis-compatible JSON; the amis client renders it. There is no Webpack config, no TypeScript compilation step, and no npm dependency in the core framework. Custom renderers, documented in §25, are the only exception, and they are served as pre-built bundles.

##### 1.1.4. Relationship to ERPNext and Frappe — conceptual debts and deliberate departures

Awo owes a conceptual debt to Frappe Framework and ERPNext. The `EntityDefinition` is directly inspired by Frappe's DocType: a single descriptor that drives persistence, UI, and permissions. The document lifecycle hooks (`before_save`, `after_save`, `on_submit`, `on_cancel`) map to Frappe's controller hooks. The child table pattern in `Table` fields mirrors Frappe's child DocType mechanism. These similarities are intentional — the Frappe model has proven itself in production ERP deployments at scale, and Awo takes its best ideas seriously.

The departures are equally deliberate. Frappe's Python runtime and MariaDB-first storage model impose a ceiling on concurrency and type safety that is unacceptable for Awo's target workloads. Frappe Workflow is a state machine configured in the database; Awo uses Temporal, which gives durable execution, audit history, and saga compensation as first-class primitives. Frappe's Jinja-based form rendering requires a frontend build step and tight coupling between server and client; Awo's amis SDUI contract decouples them entirely.

The migration path from Frappe to Awo is addressed in Appendix J. The direct concept mapping is documented in §1.4.

---

#### 1.2. Design Philosophy

##### 1.2.1. EntityDefinition as the central primitive — not just a schema, a dispatch key

The `EntityDefinition` is not merely a schema descriptor. It is the dispatch key that the framework uses to route every operation: which table to query, which permission matrix to consult, which hooks to execute, which page builder to invoke, which workflow to trigger. When a request arrives for entity `SalesOrder`, the framework looks up the `SalesOrder` `EntityDefinition` and derives every subsequent decision from it.

This design concentrates the description of a business object in one place. A developer reading an `EntityDefinition` can determine the full lifecycle of that entity — its fields, its relationships, its permission requirements, its UI presentation, and its workflow integration — without reading scattered configuration files or hunting through middleware registrations. The `EntityDefinition` is the canonical source of truth for everything the framework needs to know about a business object.

The practical consequence is that adding a new business concept to the system requires defining one `EntityDefinition` and filling in its hooks. The framework generates the API routes, the amis page skeleton, and the permission gates automatically. See §2 for a full treatment of `EntityDefinition` anatomy.

##### 1.2.2. Compile-time correctness for system entities, runtime flexibility for custom entities

System entities — `LedgerEntry`, `StockMove`, `User`, `Tenant`, and all built-in module entities — are defined in Go code. Their fields are typed Go structs. Their queries are either SQLC-generated or produced by the ent reference implementation. At compile time, the framework can verify that a hook receives the correct record type and that a query returns the expected shape. Type errors surface at `go build`, not at runtime in a production tenant's session.

Custom entities — defined by tenant administrators at runtime to extend the system with their specific fields and workflows — live in JSONB columns and are described by `CustomFieldDef` metadata records. Their validation is enforced at the application layer rather than the database schema layer. This is an intentional trade-off: custom entities sacrifice some type safety in exchange for the ability to be created and modified without a code deploy or a database migration.

The `EntityResolver` makes this duality transparent to framework consumers. A hook, a permission policy, or a page builder always receives an `EntityRecord` and an `EntityRepository` interface — the same types regardless of whether the entity is system-backed or JSONB-backed. Full details in §2.2.

##### 1.2.3. The hybrid data model rationale — why not everything in JSONB, why not everything in typed tables

Storing everything in JSONB would give maximum flexibility at the cost of query performance, index expressiveness, and the ability to enforce relational integrity at the database level. A `LedgerEntry` that must always have a matching debit and credit total, referencing a valid `Account` and a valid `Period`, cannot afford to be a bag of untyped JSON. The financial integrity constraints are worth the migration cost of a typed table.

Storing everything in typed tables would mean that every tenant customisation — a fleet customer's internal approval codes, a station manager's custom compliance fields — requires a schema migration and a code deploy. For a multi-tenant SaaS platform serving dozens of businesses with distinct operational requirements, this is impractical. Custom fields stored in JSONB with GIN indexes provide acceptable query performance for the field cardinalities typical of tenant extensions.

The rule encoded in §2.3 is: entities where correctness is a financial, legal, or security requirement are system entities with typed tables; entities where flexibility is more valuable than strictness are custom entities with JSONB storage. The framework enforces this rule by making it structurally impossible to define `LedgerEntry` as a custom entity.

##### 1.2.4. Interface-first persistence — the `EntityRepository` contract and why the implementation is replaceable

The `EntityRepository` interface is the only persistence contract that framework code and module code depend on. No import of `entgo.io/ent`, no reference to ent schema files, and no ent-specific query builder appears outside of the `internal/store/ent` package that implements the interface. This boundary is enforced by Go's package visibility rules: the ent implementation lives in `internal/`, which is not importable by consumers.

The rationale is forward compatibility. Ent is the current reference implementation because it provides a productive schema-first workflow and good PostgreSQL support. It is not the only viable implementation. A team with strict SQLC requirements, or one that needs a different query generation approach for performance reasons, can implement the `EntityRepository` interface against their preferred library and register it via the framework bootstrap (see §8.7). The framework's correctness guarantees — tenant isolation, hook execution order, privacy policy application — hold regardless of which implementation is in use, because those guarantees are enforced at the interface layer.

> **Warning:** Any code that imports ent packages directly from a module outside `internal/store/ent` is a framework violation. It creates a hidden dependency on the implementation that will break when the implementation is swapped.

##### 1.2.5. Workflow-first ERP — why business processes belong in Temporal, not in request handlers

An ERP system's defining characteristic is that its operations have business consequences that extend beyond a single HTTP request: a submitted sales order reserves stock, posts GL entries, triggers a delivery, and notifies the customer. These steps must be atomic in business terms — either all succeed or all are compensated — even though they span multiple services and may take seconds or minutes to complete.

Implementing this logic inside a Fiber request handler using nested function calls and manual rollback logic creates code that is fragile, untestable, and invisible to operators. A process crash mid-execution leaves the system in an inconsistent state with no audit trail. Temporal solves this problem structurally: workflow state is durable, execution history is queryable, and compensating transactions are a first-class primitive via the saga pattern (see §29).

The framework's convention is that any operation with more than one side effect belongs in a Temporal workflow triggered from an `after_save` hook. Request handlers are responsible for validation and initial persistence; workflows are responsible for multi-step business processes. This separation keeps handlers fast, keeps business logic observable, and keeps the system recoverable after failures.

##### 1.2.6. Multi-tenancy as a first-class concern, not a plugin

Multi-tenancy is not added to Awo via a middleware plugin or a library import. It is woven into the framework's data model, connection management, permission system, and entity registry at the architectural level. Every database operation executes against a tenant-scoped PostgreSQL schema. Every `EntityRepository` call carries a tenant context that determines which schema is targeted. Every session carries a tenant identifier that is validated before any business logic runs.

This means there is no "single-tenant mode" that skips tenant resolution. A developer running locally against a single test tenant still operates through the same tenant context machinery as a production deployment with fifty tenants. The cost is a small amount of overhead on every request; the benefit is that tenant isolation is structurally guaranteed rather than relying on developer discipline to remember to add `WHERE tenant_id = ?` to every query.

The schema-per-tenant model in PostgreSQL — where each tenant's data lives in a dedicated schema named `t_{tenant_slug}` — provides isolation at the database level as a second line of defence behind the application-layer tenant context. See §14 for the full middleware treatment.

##### 1.2.7. Server-driven UI as a force multiplier — eliminating the frontend bottleneck

In a traditional web application, adding a new form field to an ERP document requires changes in three places: the backend model, the API endpoint, and the frontend component. In a team where backend and frontend engineers are separate, this creates coordination overhead and queues of unshipped UI changes.

The SDUI model eliminates this by making the server the authoritative source of the UI definition. A page builder function in Go emits amis-compatible JSON describing the full page — fields, layout, actions, validation rules, conditional visibility — and the amis client renders it without any frontend code changes. Adding a field to an `EntityDefinition` and updating the page builder is a single deploy that makes the field appear in the UI immediately.

This is not appropriate for every UI. Highly interactive dashboards, map renderers, and real-time displays may require custom renderers (§25). But for the form-heavy, table-driven workflows that constitute the majority of ERP screen real estate, the SDUI model is a significant productivity multiplier. The tradeoff is that amis must be treated as a stable dependency; its version is pinned and upgrade decisions are made deliberately (§21.1.5).

---

#### 1.3. Who This Documentation Is For

##### 1.3.1. Framework contributors — building Awo core

Framework contributors work on the `internal/` and `pkg/` packages: the `EntityResolver`, the `EntityRegistry`, the `EntityRepository` interface and its ent implementation, the middleware pipeline, the permission resolution engine, and the workflow helper library. The relevant chapters are Part I (mental model), Part II (EntityDefinition system internals), Part III (API layer internals), Part V (workflow engine), and Part IX (framework internals and plugin system).

Contributors must understand the interface-first persistence rule (§1.2.4) absolutely. Any change to `EntityRepository` that leaks implementation details into the interface is a breaking change requiring a major version bump. Contributors are also responsible for maintaining the correctness guarantees around tenant isolation and hook execution ordering; these must be verified by the integration test suite on every change.

##### 1.3.2. Module developers — building built-in ERP modules on top of Awo

Module developers build the Finance, Forecourt, HR, Inventory, and CRM modules that ship with the framework, as well as custom modules built by teams deploying Awo for specific clients. They work primarily with `pkg/entity`, `pkg/hook`, `pkg/privacy`, `pkg/workflow`, and `pkg/permission`. The relevant chapters are Part II (EntityDefinition system), Part IV (SDUI layer), Part V (workflow engine), and Part VI (built-in module reference, both as examples and as domain documentation).

Module developers should read §2 before writing a single line of module code. The `EntityDefinition` is the unit of work; everything else follows from it. The chapter on the persistence interface (§8) must be read in full before writing any database interaction code.

##### 1.3.3. Application developers — building tenant-specific customisations

Application developers extend a deployed Awo instance with tenant-specific custom entities, custom fields on system entities, custom page builders, and custom workflows. They work primarily with the API (§17), the custom field system (§10), and the SDUI layer (Part IV). They do not typically write Go code; their customisations are configured through the admin UI or the management API.

Where application developers do write code — custom validators for `CustomFieldDef` records, for example — they are writing against the public API surface in `pkg/`. They should understand the `EntityRecord` abstraction (§2.2.4) and the validation error format expected by amis (§18.3).

##### 1.3.4. System integrators — deploying and operating Awo for a client

System integrators are responsible for infrastructure provisioning, CI/CD pipeline configuration, migration management, backup strategy, and ongoing operations. The relevant chapters are Part VIII (deployment and operations): §42 through §48, which cover environment architecture, Docker, CI/CD, PostgreSQL operations, observability, security hardening, and incident response playbooks.

Integrators should pay particular attention to §45 (PostgreSQL schema-per-tenant layout and connection pooling), §44 (CI/CD with Atlas migration lint), and §48 (incident response playbooks). The Kenya-specific configuration in §39.4 is directly relevant for all initial deployments targeting the Kenyan market.

##### 1.3.5. How to navigate this documentation by audience

The documentation is structured so that Part I (this chapter and §2) is a prerequisite for all audiences. Beyond Part I, each audience has a primary path through the document. Framework contributors read Parts I through III and Part IX. Module developers read Parts I through VI, using Part VII as a reference when configuring multi-tenant behaviour. Application developers read Parts I, IV, and the relevant sections of Part VI for domain context. System integrators read Part I for orientation, then jump directly to Part VIII.

Every chapter is designed to be self-contained, with cross-references rather than assumed reading order. However, §2 introduces the `EntityDefinition` in depth, and many later chapters reference its concepts without re-explaining them. Reading §2 before any implementation chapter will save time.

---

#### 1.4. Awo vs Frappe — Direct Comparison

##### 1.4.1. DocType vs EntityDefinition

Frappe's DocType is a database record — a row in the `tabDocType` table — that is created and modified at runtime through the Frappe UI. It drives form generation, database table creation (via auto-migration), API endpoints, and permission assignment. The DocType model's strength is that non-developers can create new document types through a GUI. Its weakness is that the system state is not fully captured in version control, and schema changes happen automatically without review.

The `EntityDefinition` is a Go struct defined in source code, compiled into the binary, and version-controlled alongside the application. Schema changes require an explicit migration file generated by Atlas and reviewed before application. This makes the system state fully auditable and reproducible, at the cost of requiring a code deploy for new system entities. Custom entities (§2.2.2) provide the runtime flexibility that some use cases require, without compromising the integrity of system entities.

##### 1.4.2. Python hooks vs Go interfaces

Frappe hooks are Python functions registered in a `hooks.py` file, called at document lifecycle events. They are dynamically discovered at runtime and have access to the full Python environment. This flexibility makes them easy to write quickly and difficult to test rigorously — the dynamic dispatch makes it hard to verify at development time that all hooks for a given DocType are correctly registered.

Awo hooks are Go functions that satisfy typed interfaces (`hook.BeforeSave`, `hook.AfterSave`, etc.) and are registered explicitly against an `EntityDefinition` at startup. The Go type system verifies at compile time that the hook function has the correct signature. The interceptor pattern (§7.7) makes hooks injectable as dependencies, enabling unit testing without a live database. The trade-off is more ceremony at registration time, in exchange for correctness guarantees that Python's dynamic model cannot provide.

##### 1.4.3. MariaDB-first vs PostgreSQL-first

Frappe stores all DocType data in MariaDB using a single-database model where every DocType becomes a `tab{DocType}` table in the same database. The schema is flat — no per-tenant isolation at the database level — and multi-tenancy is handled by filtering on `company` fields in most documents. This works for the ERPNext deployment model where each company runs its own Frappe instance, but it is not designed for a shared multi-tenant SaaS deployment.

Awo is PostgreSQL-first. The schema-per-tenant model (`t_{tenant_slug}`) uses PostgreSQL's native schema isolation to provide hard boundaries between tenants at the database level. PostgreSQL's `numeric(20,4)` type is used for all currency columns (never floating point), JSONB with GIN indexes handles custom fields, and Row-Level Security is available as a second-line defence. The features that make PostgreSQL superior for this use case — schema isolation, advanced index types, strong numeric types — are not available in MariaDB.

##### 1.4.4. Frappe Workflow vs Temporal

Frappe Workflow is a state machine defined in the Frappe UI. States are configured as DocType statuses, transitions are triggered by user actions, and optional approval conditions are evaluated as Python expressions. The model is simple and UI-accessible, but it has no concept of durability: if the Frappe process crashes during a workflow transition, the document may be left in an inconsistent state, and there is no automatic recovery.

Temporal provides durable execution at the infrastructure level. Workflow state survives process crashes, server reboots, and network partitions because it is persisted in Temporal's own store and replayed deterministically. Long-running approval workflows that wait days for a human response are first-class Temporal use cases. Saga compensation (§29) allows multi-step operations to be reversed atomically in business terms, which Frappe Workflow has no equivalent for.

##### 1.4.5. Jinja forms vs amis SDUI

Frappe generates forms by rendering Jinja templates server-side, hydrated with DocType field metadata. Form customisation requires either Python-level DocType configuration or JavaScript client scripts that manipulate the DOM. The tight coupling between the server template and the client rendering means that UI changes require coordination between backend and frontend code paths.

amis renders UI from a JSON schema produced entirely server-side by page builder functions. The server controls the complete page definition — field types, layout, visibility conditions, action buttons — and the amis client renders it without any client-side business logic. This means a backend developer can produce a fully functional ERP form without writing any JavaScript. The amis component library (Parts IV) covers the full range of ERP UI patterns.

##### 1.4.6. Migration strategy for teams coming from Frappe

Teams migrating from Frappe to Awo need to map three categories of artefacts: DocTypes to EntityDefinitions (schema and lifecycle hooks), Frappe Workflows to Temporal workflows (state machines to durable orchestration), and Client Scripts to Go hooks or page builder conditional expressions. The concept mapping is detailed in Appendix J.

The data migration path is: export Frappe data to JSON using Frappe's native export tool, transform to Awo's fixture format using the provided migration CLI utilities (`awo migrate frappe`), and import via `awo entity seed`. Financial data — GL entries, ledger balances, stock valuations — requires particular care; the migration tooling includes a validation step that checks double-entry integrity before importing any journal entries.

---

### Chapter 2 — The EntityDefinition — The Central Abstraction

#### 2.1. Why EntityDefinition Is the Central Primitive

##### 2.1.1. What the framework does when it sees an EntityDefinition

When the framework encounters an `EntityDefinition` during startup — either a system entity registered in Go code or a custom entity loaded from `CustomFieldDef` records at tenant boot — it performs a sequence of derived registrations. It registers the entity name in the `EntityRegistry` mapped to its definition. It generates REST API routes for the standard CRUD operations and any declared custom actions. It associates the entity with its declared permission matrix. It binds the entity to its registered page builder functions. It registers the entity's workflow trigger bindings with the Temporal client.

None of these derived registrations require any additional developer action. Defining the `EntityDefinition` completely is the only required step; the framework derives everything else from it. This is the practical meaning of calling it the central primitive: it is not merely declarative metadata, it is the input to a framework-level code generation pipeline that produces live API routes, UI pages, and workflow triggers.

##### 2.1.2. The five things an EntityDefinition drives — persistence routing, API generation, UI generation, permission evaluation, workflow triggering

Persistence routing means that when the `EntityResolver` receives a request to operate on entity `SalesInvoice`, it consults the registry, finds the `SalesInvoice` `EntityDefinition`, determines whether it is a system entity or a custom entity, and selects the appropriate `EntityRepository` execution path. The caller never specifies which storage backend to use; that decision is made by the resolver from the definition.

API generation means that registering an `EntityDefinition` with the framework causes the Fiber router to create `GET /api/v1/sales-invoices`, `GET /api/v1/sales-invoices/:id`, `POST /api/v1/sales-invoices`, `PATCH /api/v1/sales-invoices/:id`, `DELETE /api/v1/sales-invoices/:id`, and any custom action endpoints declared in the definition. The route handlers are generic — they operate on `EntityRecord` and delegate to the correct `EntityRepository` via the resolver.

UI generation means that if the definition includes a registered page builder, the framework exposes a `GET /api/v1/pages/sales-invoice-form` endpoint that invokes the builder and returns the amis JSON page definition. Permission evaluation means that the permission matrix declared in the definition is consulted on every request before any hook or repository call is made. Workflow triggering means that hook registrations on `on_submit` or `after_save` can start Temporal workflows using the entity's ID and tenant context as durable inputs.

##### 2.1.3. Why this is not just a schema definition tool

A pure schema definition tool describes a data structure and generates a table. The `EntityDefinition` does this, but it is not primarily a schema tool. The schema is a consequence of the definition, not its purpose. The purpose is to describe a business object completely enough that the framework can operate on it in every layer of the stack without additional configuration.

The difference is visible in what happens when a developer adds a new field. In a schema-first tool, adding a field means updating the struct and running a migration. In Awo, adding a field to an `EntityDefinition` also updates the amis form (if the page builder uses the field list introspectively), potentially updates permission matrix defaults, affects the audit log schema, and is reflected in any report that enumerates entity fields. The `EntityDefinition` is a business object descriptor; the schema is one of its outputs.

---

#### 2.2. The Two Entity Types

##### 2.2.1. System entities — strongly typed, SQL-backed, SQLC-generated or ent-generated queries

System entities are defined as Go types in source code and backed by dedicated SQL tables in the tenant schema. Their fields map to typed columns: `Currency` fields become `numeric(20,4)` columns, `DateTime` fields become `timestamptz` columns, and `Link` fields become foreign key columns with indexes. The queries that operate on system entities are either generated by ent (the current reference implementation) or by SQLC, depending on the team's preference for the implementation layer. The interface they expose upward — the `EntityRepository` — is identical in both cases.

System entities carry the framework's strongest correctness guarantees. Their field types are verified at compile time, their relational integrity is enforced by database constraints, and their lifecycle hooks are typed interfaces rather than dynamically discovered functions. The cost is that adding a new system entity requires a code change and a reviewed database migration. This cost is appropriate for entities like `LedgerEntry`, `StockMove`, and `PaymentEntry`, where correctness is non-negotiable.

> **Note:** The choice between SQLC and ent as the query generation layer for a new system entity is a team preference that does not affect any other part of the framework. The `EntityRepository` interface is the contract; either implementation satisfies it. See §8.6 for ent-specific details and §8.7 for swapping the implementation.

##### 2.2.2. Custom entities — tenant-defined, JSONB-backed, metadata-driven

Custom entities are defined by tenant administrators at runtime through the admin UI or the management API. Their field definitions are stored as `CustomFieldDef` records in the tenant schema, and their data is stored in a JSONB column in a generic `custom_entity_records` table. There is no SQL migration when a tenant adds a new custom entity or adds a field to an existing one — the schema is stable and the flexibility lives in the JSONB layer.

The framework loads custom entity definitions during tenant boot and registers them in the per-tenant `EntityRegistry`. This means that a tenant that defines a `FleetApprovalRequest` custom entity will see it appear in the API and in the admin UI after a page refresh, with no framework deployment required. The limitation is that custom entities cannot have database-level referential integrity constraints, cannot be indexed on arbitrary JSONB paths without explicit GIN index configuration, and cannot participate in complex SQL joins as efficiently as system entities.

Custom entities are appropriate for tenant-specific extensions: custom approval metadata, locale-specific compliance fields, industry-specific attributes that do not belong in the core data model. They are not appropriate for financial records, inventory movements, or any entity where data integrity must be enforced at the database level.

##### 2.2.3. The EntityResolver — how the framework picks the execution path at runtime

The `EntityResolver` is the internal component responsible for translating an entity name into an `EntityRepository` instance. It receives the entity name and the tenant context, performs a two-stage lookup in the `EntityRegistry` — first in the global system entity registry, then in the per-tenant custom entity registry — and returns an `EntityRepository` that operates correctly for the matched entity type.

For a system entity, the resolver returns the ent-backed (or SQLC-backed) `EntityRepository` for that specific entity type, configured with the correct tenant schema connection. For a custom entity, the resolver returns the JSONB engine wrapped in the `EntityRepository` interface, loaded with the `CustomFieldDef` metadata for that entity in that tenant. The caller receives the same interface in both cases.

```go
// Example: EntityResolver lookup from a route handler
func GetEntityHandler(c *fiber.Ctx) error {
    tc, err := tenant.FromContext(c.UserContext())
    if err != nil {
        return err
    }
    entityName := c.Params("entity")
    repo, err := entity.Resolve(c.UserContext(), tc, entityName)
    if err != nil {
        return err
    }
    id := c.Params("id")
    record, err := repo.Get(c.UserContext(), id)
    if err != nil {
        return err
    }
    return c.JSON(record)
}
```

##### 2.2.4. What callers see — the EntityRecord as a unified surface regardless of entity type

The `EntityRecord` is the framework's universal representation of a single entity instance, regardless of whether the underlying entity is a system entity or a custom entity. It carries the entity name, the record's primary key, a map of field values keyed by field name, and metadata about the entity's current lifecycle stage (draft, submitted, cancelled). Callers — hooks, page builders, workflow activities — interact exclusively with `EntityRecord` and never with the underlying SQL row or JSONB document.

For system entities, the `EntityRecord` is populated from typed SQL columns; field values are returned as their declared Go types (a `Currency` field returns a `decimal.Decimal`, a `DateTime` field returns a `time.Time`). For custom entities, the same `EntityRecord` structure is populated from the JSONB document, with values coerced to their declared types according to the `CustomFieldDef` metadata. A hook that reads a `Currency` field does not need to know whether the entity is system-backed or JSONB-backed; the value it receives is typed correctly in both cases.

> **Note:** The `EntityRecord` type is defined in `github.com/awolabs/awo/pkg/entity`. It is the only record type that module code should import or reference. Never import struct types from the ent-generated `ent/` package in module code.

##### 2.2.5. Why this distinction is invisible to framework consumers by design

The system/custom distinction is an implementation detail that only the `EntityResolver` and the `EntityRepository` implementations need to understand. Module developers, hook authors, page builder writers, and workflow activity authors all operate against `EntityRecord` and `EntityRepository`. This invisibility is intentional: it means that an entity which starts life as a custom entity can be promoted to a system entity (§10.6.5) without any changes to the hooks, page builders, or workflows that operate on it — only the storage backend changes.

The design also means that a module developer writing a `before_save` hook cannot introduce a system/custom entity dependency by accident. There is no API surface through which a hook could detect whether it is operating on a system entity or a custom entity, because the `EntityRepository` interface and the `EntityRecord` type expose nothing implementation-specific.

---

#### 2.3. When to Use Each Entity Type

##### 2.3.1. Use a system entity when — financial integrity, inventory accuracy, IAM, high-frequency writes

A system entity is required whenever the entity participates in operations that demand database-level integrity guarantees. Financial documents — journal entries, payment entries, bank transactions — must have typed `numeric(20,4)` columns for monetary amounts, foreign key constraints linking them to valid accounts and periods, and the ability to be joined efficiently against large transaction tables in financial statement queries. These requirements cannot be satisfied by JSONB storage.

Inventory movements are a second category. A `StockMove` that decrements one warehouse bin and increments another must be atomic at the database level; the foreign keys and the numeric precision of quantity columns must be enforced by the database schema, not by application-layer validation that could be bypassed by a buggy hook.

High-frequency writes are a practical consideration. A forecourt pump that records meter readings every few seconds will produce millions of rows per year. JSONB documents with GIN indexes are adequate for low-to-medium write volumes, but at high frequencies the overhead of JSONB parsing and path-based indexing becomes measurable. System entities with typed columns and standard B-tree indexes handle this workload better.

##### 2.3.2. Use a custom entity when — tenant-specific extensions, industry-specific fields, rapid configuration without deploys

Custom entities are the right choice when a tenant needs a data structure that is specific to their operations and that does not represent a core ERP concept requiring database-level integrity. A fleet fuel distribution company might need a `CardAuthorizationRequest` entity with approval workflow fields specific to their internal process. A manufacturing client might need a `ProductionRunLog` entity with custom compliance fields. These are tenant-specific concerns; building them as system entities would pollute the core data model with domain-specific noise.

Custom entities are also appropriate for rapid iteration during the design phase of a new module. When the field structure of a new concept is still being explored with a client, defining it as a custom entity allows the team to add, rename, and remove fields without database migrations. Once the design stabilises, the entity can be promoted to a system entity (§10.6.5) to gain the integrity and performance benefits of typed storage.

The rule of thumb is: if the entity will be referenced by foreign key from a system entity, make it a system entity. A `SalesOrder` that has a `Link` to a `Customer` entity requires that `Customer` be a system entity with a stable primary key and a foreign key constraint that the database can enforce.

##### 2.3.3. Entities that must never be custom — LedgerEntry, StockMove, Payment, User, Tenant

`LedgerEntry`, `StockMove`, `Payment`, `User`, and `Tenant` are hardcoded as system entities in the framework. No API, no admin UI, and no configuration mechanism can convert them to custom entities. This is enforced at the `EntityRegistry` level: attempts to register a custom entity with one of these reserved names will fail at tenant boot with a startup error.

The rationale for each: `LedgerEntry` and `StockMove` require numeric precision, referential integrity, and immutability after posting that only typed columns can guarantee. `Payment` integrates with external payment processors (MPesa, card networks) and requires exact decimal amounts and idempotency keys backed by unique indexes. `User` and `Tenant` underpin the authentication and multi-tenancy system; their field structures must be known at compile time for the session validation middleware to function correctly.

> **Danger:** Any attempt to work around these restrictions by creating a custom entity with a similar name (e.g., `LedgerEntry_Custom`) and using it for financial postings is a critical architecture violation. Financial data stored in JSONB does not have the precision or integrity guarantees required by Kenyan financial regulations or IFRS standards.

##### 2.3.4. Entities that should always be custom — tenant contact extensions, custom approval metadata, locale-specific fields

The inverse of the above: some entities are structurally required to be custom because they are inherently tenant-specific. Contact extensions — additional fields that a tenant's sales team attaches to `Customer` records beyond the standard system fields — belong in `CustomFieldDef` records on the `Customer` entity, not in a system entity table.

Custom approval metadata — the specific fields a tenant's compliance team requires on an approval record — is another case. Rather than building a generic `ApprovalMetadata` table with dozens of nullable columns for every possible field across all tenants, each tenant defines their own `ApprovalMetadata` custom entity with exactly the fields their process requires. The JSONB storage model is well-suited to this pattern because it does not penalise sparsely populated field sets.

Locale-specific fields — government-mandated fields for a specific regulatory context that are not universal across all Kenyan businesses — are a third case. The NEMA compliance fields for petrol stations are not relevant to a software company using Awo for HR management; making them custom entity fields on a per-tenant opt-in basis keeps the core data model clean.

##### 2.3.5. The grey zone — when to escalate a custom entity to a system entity

A custom entity should be escalated to a system entity when any of the following conditions become true: it is referenced by foreign key from a system entity; it needs to participate in a SQL join in a financial report; it accumulates more than roughly one hundred thousand records across the fleet and JSONB query latency becomes measurable; or it needs a database-level uniqueness constraint that cannot be expressed with a JSONB unique index.

The escalation procedure is documented in §10.6.5. It involves creating a new system entity with the same field structure, writing a data migration that copies records from the JSONB store to the typed table, updating any hooks and page builders that reference the entity (which is straightforward because they all use `EntityRecord` and `EntityRepository`), and deploying the migration before switching the `EntityResolver` to the system entity path.

---

#### 2.4. EntityDefinition Anatomy

##### 2.4.1. Name and identifier conventions

The `EntityDefinition` name is a PascalCase string that uniquely identifies the entity across the system: `SalesOrder`, `LedgerEntry`, `FleetCustomer`. This name is the dispatch key in the `EntityRegistry` and appears in API URL paths as its kebab-case equivalent (`sales-order`, `ledger-entry`, `fleet-customer`). The name must be stable after the entity is first deployed; renaming an entity requires a data migration and API versioning coordination.

The identifier convention for entity names follows these rules: single words are capitalised (`Customer`, `Invoice`), compound names use PascalCase without underscores (`SalesInvoice`, `LedgerEntry`), and module prefix is added when there is ambiguity between domains (`FinanceAccount` vs `HRAccount`). The framework uses the name directly in log lines, audit records, permission identifiers, and workflow IDs, so names should be human-readable and self-explanatory without abbreviation.

> **Note:** Entity names are used as part of workflow ID construction following the `{tenant}.{entity}.{id}.{action}` pattern (§27.5.1). Names containing special characters other than alphanumerics will cause workflow ID validation failures. Stick to PascalCase identifiers.

##### 2.4.2. Field list — types, constraints, metadata

The field list declares every attribute of the entity: its name, type, constraints, and metadata. Field names use snake_case (`account_code`, `posting_date`, `total_amount`) and are stable identifiers used as JSON keys in the API response, as JSONB paths for custom field storage, and as permission matrix row identifiers for field-level permissions. The full field type reference is in §5.

```go
// Example: Declaring fields on an EntityDefinition
var SalesOrderDefinition = entity.Define("SalesOrder",
    entity.Fields(
        entity.Field("order_number").
            Type(entity.Data).
            MaxLen(20).
            Immutable().
            Required(),
        entity.Field("customer_id").
            Type(entity.Link).
            LinkedEntity("Customer").
            Required(),
        entity.Field("order_date").
            Type(entity.Date).
            Required(),
        entity.Field("total_amount").
            Type(entity.Currency).
            Required(),
        entity.Field("status").
            Type(entity.Select).
            Options("draft", "submitted", "cancelled").
            Default("draft").
            Required(),
    ),
)
```

Each field carries metadata beyond its type and constraints: a human-readable label used by the page builder when generating form field labels, an optional description used as help text in the amis form, and a `Sensitive` flag that causes the framework to exclude the field from log lines and from API responses unless explicitly requested with the `fields=` query parameter. The full field options reference is in §5.2.

##### 2.4.3. Edge declarations — relationships to other entities

Edges declare relationships between entities and drive the generation of foreign key columns, join methods, and relational API endpoints. An edge declaration names the related entity, specifies the cardinality (one-to-many, many-to-many, or self-referencing), and configures cascade behaviour. Full edge documentation is in §6.

```go
// Example: Declaring edges on an EntityDefinition
var SalesOrderDefinition = entity.Define("SalesOrder",
    entity.Fields( /* ... */ ),
    entity.Edges(
        entity.Edge("customer").
            To("Customer").
            Required(),
        entity.Edge("line_items").
            To("SalesOrderLine").
            OneToMany().
            CascadeDelete(),
        entity.Edge("delivery_notes").
            From("DeliveryNote").
            OneToMany(),
    ),
)
```

Edge declarations also affect the page builder pipeline: a `Table` field backed by a one-to-many edge will cause the default page builder to render a `LineItemEditor` component in the form view. Edges do not generate automatic eager loading; the `EntityRepository.Query()` method accepts an `Include` option to specify which edges to load, and callers are responsible for specifying this to avoid N+1 query patterns.

##### 2.4.4. Hook registrations

Hooks are registered on an `EntityDefinition` as typed function values that satisfy the hook interface for their lifecycle stage. The framework invokes registered hooks in the order they are declared, passing a `hook.Context` that carries the current `EntityRecord`, the tenant context, the user context, and the `EntityRepository` for the entity being operated on. Full hook documentation is in §7.

```go
// Example: Registering lifecycle hooks on an EntityDefinition
var SalesOrderDefinition = entity.Define("SalesOrder",
    entity.Fields( /* ... */ ),
    entity.Hooks(
        hook.BeforeSave(validateOrderTotals),
        hook.AfterSave(emitOrderCreatedEvent),
        hook.OnSubmit(triggerOrderSubmitWorkflow),
    ),
)

func validateOrderTotals(ctx hook.Context) error {
    tc, err := tenant.FromContext(ctx)
    if err != nil {
        return err
    }
    _ = tc // use tenant context for any tenant-specific validation rules
    record := ctx.Record()
    total := record.Get("total_amount")
    if total == nil {
        return hook.ValidationError("total_amount", "total_amount is required before save")
    }
    return nil
}
```

> **Warning:** Hook functions must not perform long-running I/O synchronously. Operations that might take more than a few hundred milliseconds — external API calls, large aggregate queries — belong in Temporal activities triggered from `after_save`, not in the hook body itself. A slow hook blocks the HTTP request and the open database transaction.

##### 2.4.5. Permission policy bindings

A permission policy binding declares which privacy policies (§9) are applied to queries and mutations on this entity. The bound policies are evaluated after role-based permission checks pass; they further restrict the result set or mutation targets based on runtime conditions such as department membership, ownership, or tenant-specific configuration.

```go
// Example: Binding privacy policies to an EntityDefinition
var SalesOrderDefinition = entity.Define("SalesOrder",
    entity.Fields( /* ... */ ),
    entity.Policy(
        privacy.And(
            privacy.TenantIsolation(),
            privacy.Or(
                privacy.RoleFilter("sales_manager", nil),
                privacy.OwnerOnly("assigned_to"),
            ),
        ),
    ),
)
```

Policy bindings are evaluated at the `EntityRepository` interface layer, meaning they apply to every query and mutation regardless of which code path calls the repository. A report query, a hook that reads related records, and a direct API request all pass through the same policy evaluation. See §9 for the full policy type reference.

##### 2.4.6. Workflow trigger bindings

Workflow trigger bindings connect entity lifecycle events to Temporal workflow starts. A binding specifies the hook event that triggers the workflow, the workflow function to start, and the task queue to use. The workflow receives the entity's record ID and the tenant context as its initial input; it fetches any additional state it needs from the `EntityRepository` inside its activities.

```go
// Example: Binding a Temporal workflow trigger to an EntityDefinition
var SalesOrderDefinition = entity.Define("SalesOrder",
    entity.Fields( /* ... */ ),
    entity.WorkflowTriggers(
        entity.OnSubmit(
            SalesOrderSubmitWorkflow,
            entity.TaskQueue("finance.sales-order.submit"),
            entity.WorkflowIDPattern("{tenant}.sales-order.{id}.submit"),
        ),
    ),
)
```

The `WorkflowIDPattern` uses the naming convention from §27.5.1. The `{tenant}` and `{id}` tokens are interpolated at trigger time from the tenant context and the saved record's primary key. This pattern guarantees idempotency: re-submitting a record that already has a running workflow with the same ID will be handled according to the configured duplicate policy (`RejectDuplicate` by default).

##### 2.4.7. UI page builder bindings

A page builder binding registers a Go function that the framework calls when a client requests the amis page definition for this entity. The binding specifies the page name (used in the URL path), the page type (list, form, detail, report), and the builder function. Full page builder documentation is in §21.

```go
// Example: Binding amis page builders to an EntityDefinition
var SalesOrderDefinition = entity.Define("SalesOrder",
    entity.Fields( /* ... */ ),
    entity.Pages(
        entity.Page("sales-order-list").
            Type(entity.PageTypeList).
            Builder(BuildSalesOrderListPage),
        entity.Page("sales-order-form").
            Type(entity.PageTypeForm).
            Builder(BuildSalesOrderFormPage),
    ),
)
```

Page builder functions receive the resolved permissions, feature flags, and tenant configuration as inputs, not the raw request. This means page builders are pure functions of their inputs and can be tested without a live HTTP server or database. The convention for builder function names is `Build{Entity}{View}Page` as specified in §21.3.2.

##### 2.4.8. Naming series configuration

A naming series provides automatically generated, human-readable document numbers for entities where document identification is a business requirement: sales invoices, purchase orders, journal entries. The series configuration declares a format string composed of tokens, a sequence counter name, and a reset policy.

```go
// Example: Configuring a naming series on an EntityDefinition
var SalesInvoiceDefinition = entity.Define("SalesInvoice",
    entity.Fields( /* ... */ ),
    entity.NamingSeries(
        entity.Series("INV-{YYYY}-{MM}-{SEQ}").
            SequenceName("sales_invoice_seq").
            ResetPolicy(entity.ResetMonthly).
            PaddedLength(5),
    ),
)
```

The available tokens are `{PREFIX}`, `{YYYY}`, `{MM}`, `{DD}`, `{SEQ}`, and `{TENANT}`. The sequence counter is managed as an atomic PostgreSQL sequence in the tenant schema, ensuring that concurrent invoice creation never produces duplicate numbers. Full naming series documentation is in §5.4.

---

#### 2.5. The EntityRegistry

##### 2.5.1. Global registry — what it holds and when it is populated

The `EntityRegistry` is a concurrent-safe in-memory store that maps entity names to their `EntityDefinition` descriptors and, for system entities, to their resolved `EntityRepository` instances. It has two layers: a global layer populated at process startup with system entity definitions, and a per-tenant layer populated at tenant boot with that tenant's custom entity definitions.

The global layer is immutable after startup. System entity definitions are registered by the framework's own `init`-time registration calls and by each built-in module's `Register` call during the bootstrap sequence (§49.1). No runtime code can modify the global layer. This immutability makes the global registry safe for concurrent reads across all request goroutines without locking.

##### 2.5.2. System entity registration at compile time

System entities are registered by calling `entity.Register` with their `EntityDefinition` during the process bootstrap sequence. The framework's core entities are registered in the framework package itself; module entities are registered in each module's `Register(app framework.App)` function, which is called during the startup sequence described in §49.1.

```go
// Example: Registering a system entity during module bootstrap
func (m *FinanceModule) Register(app entity.App) error {
    if err := app.RegisterEntity(SalesInvoiceDefinition); err != nil {
        return fmt.Errorf("finance: registering SalesInvoice: %w", err)
    }
    if err := app.RegisterEntity(JournalEntryDefinition); err != nil {
        return fmt.Errorf("finance: registering JournalEntry: %w", err)
    }
    return nil
}
```

Registration fails with a startup error if an entity with the same name is already registered. This prevents accidental shadowing of system entities by module code, and it makes the "reserved entity names" rule from §2.3.3 enforceable at startup rather than at runtime.

##### 2.5.3. Custom entity loading at tenant boot

When a tenant's HTTP request arrives and passes tenant resolution middleware, the framework checks whether the per-tenant `EntityRegistry` layer has been populated for that tenant. If not — typically on the first request after a server restart — the framework loads all `CustomFieldDef` records for the tenant from the database, constructs `EntityDefinition` descriptors from them, and registers them in the per-tenant registry layer.

```go
// Example: Loading custom entities during tenant boot (internal framework code)
func loadTenantCustomEntities(ctx context.Context, tc tenant.TenantContext, repo entity.EntityRepository) error {
    filter := entity.NewFilter().Eq("tenant_id", tc.ID()).Eq("active", true)
    defs, _, err := repo.Query(ctx, filter)
    if err != nil {
        return fmt.Errorf("loading custom field defs for tenant %s: %w", tc.Slug(), err)
    }
    for _, def := range defs {
        entityDef, err := entity.BuildCustomDefinition(def)
        if err != nil {
            return fmt.Errorf("building custom entity %s: %w", def.Get("entity_name"), err)
        }
        if err := tc.Registry().RegisterCustomEntity(entityDef); err != nil {
            return fmt.Errorf("registering custom entity: %w", err)
        }
    }
    return nil
}
```

This loading is performed once per tenant per process lifetime and cached. A tenant adding a new custom entity via the admin UI triggers a cache invalidation for that tenant's registry layer, causing the next request to reload it.

##### 2.5.4. Registry lookup — by name, by tenant, by type

The registry exposes three lookup patterns. Lookup by name is the most common: `registry.Get("SalesOrder")` returns the `EntityDefinition` for `SalesOrder` if it is a system entity, or an error if the name is unknown. Lookup by tenant combines the global and per-tenant layers: `registry.GetForTenant("FleetApprovalRequest", tenantCtx)` checks the tenant's custom registry first, then the global registry, and returns the first match.

Lookup by type filters the full registry to return all entities of a given category: `registry.AllSystem()` returns all system entity definitions, `registry.AllCustomForTenant(tenantCtx)` returns all custom entities defined for a specific tenant. These bulk lookups are used by the page builder pipeline when generating navigation menus and by the migration tooling when checking for drift.

##### 2.5.5. Registry concurrency model — reads vs writes during tenant boot

The global system entity registry is written once at startup and read concurrently thereafter. It uses no locks after the write phase because writes and reads do not overlap in time: startup is single-goroutine by design, and serving begins only after registration is complete.

The per-tenant custom entity registry uses a `sync.RWMutex` per tenant slot. Reads (the common case — every request that touches a custom entity) take a read lock. Writes (tenant boot loading, admin UI cache invalidation) take a write lock. The per-slot mutex design means that a write lock for tenant A does not block reads for tenant B — tenants do not contend with each other at the registry level.

> **Warning:** Never hold a registry lock across a database call. The tenant boot loading sequence acquires the write lock only after the database query completes. Holding the write lock during the query would serialize all custom entity access across all requests for that tenant, creating a startup latency spike.

---

#### Chapter summary

Chapters 1 and 2 establish the mental model and central abstraction that every other part of the documentation builds on. The framework's inversion-of-control model (§1.1.2), the interface-first persistence contract (§1.2.4), and the dual system/custom entity model (§2.2) are the three concepts most critical to understand before reading any implementation chapter. Chapter 2's treatment of `EntityDefinition` anatomy (§2.4) is the practical reference that module developers will return to most frequently during initial development; every subsection of §2.4 describes a binding that a developer must configure to make a new entity fully operational across all framework layers.

<!-- **Next chapters to read:** -->
<!---->
<!-- - §3 — Architecture Overview (the natural continuation of the mental model into a concrete system diagram, showing how all the components described here interact at runtime) -->
<!-- - §5 — Field System (the immediate practical need after understanding `EntityDefinition`: the complete field type and constraint reference needed to define entity fields correctly) -->
<!-- - §7 — The EntityRecord Lifecycle (the hook system that makes `EntityDefinition` bindings live: how `before_save`, `after_save`, `on_submit`, and the rest are invoked and what is available in context at each stage) -->
<!-- - §8 — The Persistence Interface (the `EntityRepository` contract that all persistence examples in later chapters depend on: read this before writing any code that touches the database) -->
