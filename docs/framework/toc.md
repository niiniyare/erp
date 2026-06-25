# Awo Framework — Complete Documentation
## Table of Contents

> **Reading guide**
> This TOC is the authoritative structure for all Awo Framework documentation. Every
> section number is stable and used as a cross-reference anchor throughout the docs.
> Part I establishes the mental model every other part depends on. Read it before
> jumping to any implementation chapter. The DB persistence layer (Parts II–III) is
> documented against the `EntityRepository` interface; the current reference
> implementation is `pgstore` (raw pgx v5 against PostgreSQL), but the interface is
> the contract — implementation can be swapped without touching framework consumers.

---

## Part I — Framework Foundations & Mental Model

### 1. Introduction

#### 1.1. What Awo Framework Is
##### 1.1.1. A Go-native platform for building multi-tenant ERP systems
##### 1.1.2. What "framework" means here vs "library" vs "application"
##### 1.1.3. What Awo is not — anti-patterns explicitly out of scope
##### 1.1.4. Relationship to ERPNext and Frappe — conceptual debts and deliberate departures

#### 1.2. Design Philosophy
##### 1.2.1. EntityDefinition as the central primitive — not just a schema, a dispatch key
##### 1.2.2. Compile-time correctness for system entities, runtime flexibility for custom entities
##### 1.2.3. The hybrid data model rationale — why not everything in JSONB, why not everything in typed tables
##### 1.2.4. Interface-first persistence — the `EntityRepository` contract and why the implementation is replaceable
##### 1.2.5. Workflow-first ERP — why business processes belong in Temporal, not in request handlers
##### 1.2.6. Multi-tenancy as a first-class concern, not a plugin
##### 1.2.7. Server-driven UI as a force multiplier — eliminating the frontend bottleneck

#### 1.3. Who This Documentation Is For
##### 1.3.1. Framework contributors — building Awo core
##### 1.3.2. Module developers — building ERP modules on top of Awo (separate repositories)
##### 1.3.3. Application developers — building tenant-specific customisations
##### 1.3.4. System integrators — deploying and operating Awo for a client
##### 1.3.5. How to navigate this documentation by audience

#### 1.4. Awo vs Frappe — Direct Comparison
##### 1.4.1. DocType vs EntityDefinition
##### 1.4.2. Python hooks vs Go interfaces
##### 1.4.3. MariaDB-first vs PostgreSQL-first
##### 1.4.4. Frappe Workflow vs Temporal
##### 1.4.5. Jinja forms vs amis SDUI
##### 1.4.6. Migration strategy for teams coming from Frappe

---

### 2. The EntityDefinition — The Central Abstraction

#### 2.1. Why EntityDefinition Is the Central Primitive
##### 2.1.1. What the framework does when it sees an EntityDefinition
##### 2.1.2. The five things an EntityDefinition drives — persistence routing, API generation, UI generation, permission evaluation, workflow triggering
##### 2.1.3. Why this is not just a schema definition tool

#### 2.2. The Two Entity Types
##### 2.2.1. System entities — strongly typed, SQL-backed, SQLC-generated or ent-generated queries
##### 2.2.2. Custom entities — tenant-defined, JSONB-backed, metadata-driven
##### 2.2.3. The EntityResolver — how the framework picks the execution path at runtime
##### 2.2.4. What callers see — the EntityRecord as a unified surface regardless of entity type
##### 2.2.5. Why this distinction is invisible to framework consumers by design

#### 2.3. When to Use Each Entity Type
##### 2.3.1. Use a system entity when — financial integrity, inventory accuracy, IAM, high-frequency writes
##### 2.3.2. Use a custom entity when — tenant-specific extensions, industry-specific fields, rapid configuration without deploys
##### 2.3.3. Entities that must never be custom — LedgerEntry, StockMove, Payment, User, Tenant
##### 2.3.4. Entities that should always be custom — tenant contact extensions, custom approval metadata, locale-specific fields
##### 2.3.5. The grey zone — when to escalate a custom entity to a system entity

#### 2.4. EntityDefinition Anatomy
##### 2.4.1. Name and identifier conventions
##### 2.4.2. Field list — types, constraints, metadata
##### 2.4.3. Edge declarations — relationships to other entities
##### 2.4.4. Hook registrations
##### 2.4.5. Permission policy bindings
##### 2.4.6. Workflow trigger bindings
##### 2.4.7. UI page builder bindings
##### 2.4.8. Naming series configuration
##### 2.4.9. Print template binding — optional `PrintTemplate` field for PDF generation

#### 2.5. The EntityRegistry
##### 2.5.1. Global registry — what it holds and when it is populated
##### 2.5.2. System entity registration at compile time
##### 2.5.3. Custom entity loading at tenant boot
##### 2.5.4. Registry lookup — by name, by tenant, by type
##### 2.5.5. Registry concurrency model — reads vs writes during tenant boot

#### 2.6. Where EntityDefinitions Live
##### 2.6.1. Framework package (`awo.so/framework`) — platform entities owned by core
###### 2.6.1.1. Platform entity list: Tenant, OrgNode (Company / Division), User, Role, Permission, UserRole, RolePermission, Session, AuditLog, FeatureFlag, FeatureFlagOverride, Notification, TenantConfig, CustomFieldDef, ReportDefinition
###### 2.6.1.2. Why these live in the framework — every Awo deployment needs them; they underpin IAM, tenancy, and configurability
###### 2.6.1.3. Versioning — platform entity schema changes follow the framework semver contract
##### 2.6.2. Module packages — domain entities owned by separate Go modules
###### 2.6.2.1. Finance module (separate repository): JournalEntry, Account, CostCentre, FiscalYear, FiscalPeriod, ExchangeRate
###### 2.6.2.2. CRM module: Customer, Contact, Address, Lead, Opportunity
###### 2.6.2.3. Inventory module: Item, ItemGroup, Warehouse, Bin, StockEntry, StockEntryLine
###### 2.6.2.4. HR module: Employee, Department, LeaveType, LeaveAllocation, LeaveRequest, Attendance, PayrollRun, Payslip
###### 2.6.2.5. Forecourt module: Site, Tank, Pump, Nozzle, Shift, MeterReading, DipReading, FleetCard
###### 2.6.2.6. Module packages import `awo.so/framework` but are not imported by it — no circular dependency
##### 2.6.3. Application code — tenant-specific custom entities defined at runtime
###### 2.6.3.1. Custom entities use the JSONB engine and require no code deploy — defined via API or admin UI
###### 2.6.3.2. They may extend system entities (extra fields via CustomFieldDef) or stand alone
###### 2.6.3.3. Custom entities are per-tenant — loaded into the EntityRegistry at tenant boot

---

### 3. Architecture Overview

#### 3.1. The Five-Layer Model
##### 3.1.1. UI layer — amis SDUI JSON served by the API
##### 3.1.2. API layer — Fiber HTTP server, middleware pipeline, route handlers
##### 3.1.3. Domain layer — EntityDefinition system, hooks, validators, permission policies
##### 3.1.4. Workflow layer — Temporal workflows, activities, sagas
##### 3.1.5. Store layer — EntityRepository interface, current pgstore (pgx v5) implementation, PostgreSQL with RLS

#### 3.2. Request Lifecycle — Read Path
##### 3.2.1. HTTP request arrives at Fiber
##### 3.2.2. Middleware pipeline — request ID, tenant resolution, session validation, rate limiting
##### 3.2.3. Route handler resolves EntityDefinition from URL
##### 3.2.4. EntityResolver selects execution path — system or custom
##### 3.2.5. Permission policy evaluated against resolved tenant + user context
##### 3.2.6. EntityRepository.Query() called — implementation dispatches to ent or JSONB engine
##### 3.2.7. Privacy policy applied to result set
##### 3.2.8. Response serialised and returned

#### 3.3. Request Lifecycle — Write Path with Workflow
##### 3.3.1. HTTP request arrives, middleware pipeline runs
##### 3.3.2. Route handler extracts payload, validates against EntityDefinition field schema
##### 3.3.3. EntityRecord assembled from validated input
##### 3.3.4. before_save hooks invoked
##### 3.3.5. EntityRepository.Mutate() called — ent transaction opened
##### 3.3.6. after_save hooks invoked inside transaction
##### 3.3.7. Temporal workflow triggered via signal or start — carries EntityRecord ID
##### 3.3.8. Transaction committed — workflow runs asynchronously
##### 3.3.9. HTTP response returned — workflow outcome delivered via notification or polling

#### 3.4. Multi-Tenant Architecture
##### 3.4.1. Tenant identification — subdomain parsing, X-Awo-Tenant header fallback
##### 3.4.2. Per-tenant EntityRegistry — custom entities loaded at boot
##### 3.4.3. RLS isolation — all tenants share a single PostgreSQL database and schema; isolation is enforced by Row-Level Security (RLS) keyed on `tenant_id`
##### 3.4.4. `set_tenant_context(uuid)` stored procedure — called at connection acquisition, sets the RLS context variable
##### 3.4.5. Per-tenant feature flags — Redis-backed, per-tenant overrides
##### 3.4.6. Per-tenant configuration — TenantConfig entity, inheritance from system defaults

#### 3.5. Component Dependency Map
##### 3.5.1. Startup order — config → database → Redis → EntityRegistry → Fiber → Temporal worker
##### 3.5.2. What depends on what — dependency graph for contributors
##### 3.5.3. What can fail independently without cascading — Redis, Temporal
##### 3.5.4. What failing brings the process down — database, EntityRegistry

---

### 4. Quick Start — Your First EntityDefinition in 30 Minutes

#### 4.1. Prerequisites and Tooling
##### 4.1.1. Go 1.22 or later
##### 4.1.2. Docker Compose for local dependencies
##### 4.1.3. Awo CLI — `awo` binary distributed with the framework, no separate download
##### 4.1.4. Temporal CLI installation

#### 4.2. Scaffold and Run the Starter Project
##### 4.2.1. `awo new <project-name>` — scaffold a new Awo project
##### 4.2.2. Repository layout of the starter
##### 4.2.3. `docker compose up` — PostgreSQL, Redis, Temporal
##### 4.2.4. `awo serve` — first run

#### 4.3. Define a System Entity
##### 4.3.1. Run `awo entity create --type=system`
##### 4.3.2. Scaffold structure — schema file, repository interface, hook stubs
##### 4.3.3. Add fields to the schema
##### 4.3.4. Declare an edge to an existing entity

#### 4.4. Generate and Apply the Migration
##### 4.4.1. `awo migrate create <description>` — create a new migration file pair
##### 4.4.2. Edit the generated .up.sql and .down.sql files
##### 4.4.3. `awo migrate up` — apply pending migrations

#### 4.5. Wire an API Route
##### 4.5.1. Auto-generated CRUD routes from EntityDefinition
##### 4.5.2. Add a custom action route
##### 4.5.3. Test the endpoint with curl

#### 4.6. Emit an amis Page Definition
##### 4.6.1. Register a page builder function
##### 4.6.2. Visit the page in the browser

#### 4.7. Trigger a Simple Workflow
##### 4.7.1. Define a one-activity workflow
##### 4.7.2. Bind it to the entity's on_submit hook
##### 4.7.3. Watch it run in the Temporal Web UI

---

## Part II — The EntityDefinition System

### 5. Field System

#### 5.1. Field Types Reference
##### 5.1.1. Scalar types
###### 5.1.1.1. `Data` — UTF-8 string, configurable max length
###### 5.1.1.2. `SmallText` — unindexed, up to 1024 characters
###### 5.1.1.3. `LongText` — unbounded, stored as `text` column
###### 5.1.1.4. `Int` — 64-bit signed integer
###### 5.1.1.5. `Float` — 64-bit IEEE 754
###### 5.1.1.6. `Currency` — `numeric(20,4)`, never floating point
###### 5.1.1.7. `Bool` — boolean, never nullable
###### 5.1.1.8. `Date` — calendar date, no timezone
###### 5.1.1.9. `DateTime` — timestamp with timezone, stored as UTC
###### 5.1.1.10. `Time` — time of day
###### 5.1.1.11. `UUID` — `uuid` column, auto-generated default
##### 5.1.2. Structured types
###### 5.1.2.1. `Select` — single value from a declared option set
###### 5.1.2.2. `MultiSelect` — set of values from a declared option set
###### 5.1.2.3. `JSON` — arbitrary JSONB, schema-validated at the application layer
##### 5.1.3. Relational types
###### 5.1.3.1. `Link` — foreign key to another EntityDefinition
###### 5.1.3.2. `DynamicLink` — polymorphic foreign key, carries entity name + id
###### 5.1.3.3. `Table` — child entity inline (one-to-many in same form)
##### 5.1.4. File types
###### 5.1.4.1. `Attach` — file reference, stored path or object storage key
###### 5.1.4.2. `AttachImage` — image reference with thumbnail metadata

#### 5.2. Field Options and Constraints
##### 5.2.1. `Required` — non-nullable, validated before persist
##### 5.2.2. `Unique` — unique index, validated before persist
##### 5.2.3. `Immutable` — set on create, rejected on update
##### 5.2.4. `Sensitive` — excluded from logs, excluded from API responses unless explicitly requested
##### 5.2.5. `Default` — static value or Go function
##### 5.2.6. `MaxLen` — enforced at validator, not only at DB
##### 5.2.7. `Min` / `Max` — for numeric fields
##### 5.2.8. `Options` — declared option set for Select and MultiSelect
##### 5.2.9. `Translatable` — value stored with locale key, resolved at response time

#### 5.3. Field Validators
##### 5.3.1. Built-in validators — email, phone (E.164), URL, regex, KES amount
##### 5.3.2. Writing a custom field validator
##### 5.3.3. Cross-field validators — validators that read sibling field values
##### 5.3.4. Async validators — validators that query the DB (uniqueness checks)
##### 5.3.5. Validator execution order and short-circuit behaviour
##### 5.3.6. Returning field-level validation errors for amis rendering

#### 5.4. Naming Series
##### 5.4.1. What naming series are and why ERP documents need them
##### 5.4.2. Declaring a naming series on an EntityDefinition
##### 5.4.3. Format tokens — `{PREFIX}`, `{YYYY}`, `{MM}`, `{DD}`, `{SEQ}`, `{TENANT}`
##### 5.4.4. Sequence management — per-series atomic counter in PostgreSQL
##### 5.4.5. Reset rules — annual reset, monthly reset, never reset
##### 5.4.6. Tenant-specific series prefix overrides
##### 5.4.7. Retroactive renumbering — when it is safe and when it is never safe

---

### 6. Edges — Relationships Between EntityDefinitions

#### 6.1. Edge Fundamentals
##### 6.1.1. What an edge declaration generates — FK column, index, join method
##### 6.1.2. Edge direction — owner side vs inverse side
##### 6.1.3. Edge naming conventions

#### 6.2. One-to-Many Edges
##### 6.2.1. Declaring the edge on both sides
##### 6.2.2. FK column placement — always on the many side
##### 6.2.3. Eager loading vs lazy loading — performance implications
##### 6.2.4. Cascade delete — when to use, when to guard with a before_delete hook
##### 6.2.5. Orphan handling — restrict, set null, cascade

#### 6.3. Many-to-Many Edges
##### 6.3.1. Junction table generation
##### 6.3.2. Junction table annotations — adding payload fields to the relationship
##### 6.3.3. Querying through many-to-many edges
##### 6.3.4. Performance characteristics of deep many-to-many joins

#### 6.4. Self-Referencing Edges
##### 6.4.1. Tree structures — parent/children pattern
##### 6.4.2. Materialised path for deep hierarchies (account trees, org charts)
##### 6.4.3. Adjacency list for shallow hierarchies
##### 6.4.4. Querying ancestors and descendants efficiently

#### 6.5. Polymorphic Relationships
##### 6.5.1. When to use DynamicLink vs a union of concrete Links
##### 6.5.2. DynamicLink storage — `{field}_type` + `{field}_id` column pair
##### 6.5.3. Querying polymorphic edges
##### 6.5.4. Limitations — no FK constraint, application-layer integrity only

---

### 7. The EntityRecord Lifecycle

#### 7.1. Lifecycle Stages
##### 7.1.1. Stage overview — CREATE → VALIDATE → AUTHORIZE → PERSIST → POST-PROCESS
##### 7.1.2. Where hooks fire relative to stages
##### 7.1.3. What can be aborted and at which stage
##### 7.1.4. Transaction boundaries — what is inside the DB transaction

#### 7.2. The `before_validate` Hook
##### 7.2.1. Purpose — compute derived fields, normalise input before validation
##### 7.2.2. What is available in context at this point
##### 7.2.3. Aborting from `before_validate` — validation errors vs system errors

#### 7.3. The `before_save` Hook
##### 7.3.1. Purpose — enforce business rules that require the full validated record
##### 7.3.2. Accessing the previous version of the record on update
##### 7.3.3. Triggering synchronous side effects that must be transactional

#### 7.4. The `after_save` Hook
##### 7.4.1. Purpose — post-persist side effects, cache invalidation, event emission
##### 7.4.2. Still inside the transaction — implications
##### 7.4.3. Triggering Temporal workflows from `after_save`
##### 7.4.4. Avoiding slow operations inside `after_save`

#### 7.5. The `before_delete` Hook
##### 7.5.1. Purpose — guard deletion, check referential integrity that the DB cannot
##### 7.5.2. Soft delete pattern — marking `deleted_at` instead of hard delete
##### 7.5.3. Returning user-facing errors from `before_delete`

#### 7.6. The `on_submit` and `on_cancel` Hooks
##### 7.6.1. What submission means in ERP context — document finalisation
##### 7.6.2. Immutability after submission — which fields lock and which do not
##### 7.6.3. `on_cancel` as a compensating action — reversing GL postings, stock moves
##### 7.6.4. Amendment workflow — creating a new version of a submitted document

#### 7.7. Writing Testable Hooks
##### 7.7.1. The interceptor pattern — hooks as dependencies, not global registrations
##### 7.7.2. Unit testing hooks in isolation without a live database
##### 7.7.3. Integration testing hooks with an in-memory ent client

#### 7.8. Chaining Multiple Hooks
##### 7.8.1. Hook execution order when multiple hooks are registered on one entity
##### 7.8.2. Early exit — stopping the chain without an error
##### 7.8.3. Sharing context between chained hooks

---

### 8. The Persistence Interface

#### 8.1. The `EntityRepository` Interface
##### 8.1.1. Why the interface exists — the persistence layer must be swappable
##### 8.1.2. Interface contract — the methods every implementation must provide
##### 8.1.3. What the interface intentionally does NOT expose — no ORM leakage
##### 8.1.4. How framework code depends on the interface, never the implementation

#### 8.2. Interface Methods — Read Operations
##### 8.2.1. `Get(ctx, id) → (EntityRecord, error)` — single record by primary key
##### 8.2.2. `Query(ctx, filter) → ([]EntityRecord, PageInfo, error)` — filtered list
##### 8.2.3. `Exists(ctx, predicate) → (bool, error)` — existence check without fetch
##### 8.2.4. `Count(ctx, filter) → (int, error)` — aggregate count
##### 8.2.5. `Aggregate(ctx, spec) → (AggregateResult, error)` — sum, avg, min, max, group by

#### 8.3. Interface Methods — Write Operations
##### 8.3.1. `Create(ctx, input) → (EntityRecord, error)`
##### 8.3.2. `Update(ctx, id, input) → (EntityRecord, error)`
##### 8.3.3. `Delete(ctx, id) → error`
##### 8.3.4. `BulkCreate(ctx, inputs) → ([]EntityRecord, error)`
##### 8.3.5. `BulkUpdate(ctx, filter, patch) → (int, error)`

#### 8.4. Interface Methods — Transaction Support
##### 8.4.1. `WithTx(ctx, fn) → error` — scoped transaction
##### 8.4.2. Nested transaction semantics — savepoints vs flat transactions
##### 8.4.3. Passing the transactional repository through context
##### 8.4.4. Rollback on hook error — automatic vs manual

#### 8.5. The Filter and Query DSL
##### 8.5.1. `Filter` struct — field predicates, logical operators
##### 8.5.2. Comparison operators — eq, neq, gt, gte, lt, lte, in, not_in
##### 8.5.3. String operators — contains, starts_with, ends_with, ilike
##### 8.5.4. Null operators — is_null, is_not_null
##### 8.5.5. Logical composition — And, Or, Not
##### 8.5.6. JSONB predicates — path operators for custom entity fields
##### 8.5.7. Pagination — cursor-based (default), offset-based (reports only)
##### 8.5.8. Sorting — multi-field, nulls-last default

#### 8.6. The `ent` Reference Implementation
##### 8.6.1. How the ent implementation satisfies the `EntityRepository` interface
##### 8.6.2. Schema file layout and conventions
##### 8.6.3. How ent predicates are generated from `Filter` structs
##### 8.6.4. Connection pool management with pgx
##### 8.6.5. RLS context injection — `set_tenant_context()` called on connection acquisition
##### 8.6.6. Known limitations of the ent implementation

#### 8.7. Swapping the Implementation
##### 8.7.1. When you would swap — performance requirements, licensing, SQLC preference
##### 8.7.2. What a new implementation must satisfy — full interface contract + test suite
##### 8.7.3. The implementation test suite — running it against a new implementation
##### 8.7.4. Registering a custom implementation via the framework bootstrap

---

### 9. Privacy Policies — Row-Level Security

#### 9.1. Why Privacy Policies Are Separate From RBAC
##### 9.1.1. RBAC controls what operations a role can perform
##### 9.1.2. Privacy policies control what rows a query can return and modify
##### 9.1.3. Why application-level WHERE clauses are insufficient
##### 9.1.4. How privacy policies are enforced at the `EntityRepository` interface layer

#### 9.2. Policy Types
##### 9.2.1. Query rules — applied to all SELECT operations
##### 9.2.2. Mutation rules — applied to CREATE, UPDATE, DELETE
##### 9.2.3. Field visibility rules — masking or excluding fields from results

#### 9.3. Built-in Policy Primitives
##### 9.3.1. `TenantIsolation` — every query scoped to the resolved tenant
##### 9.3.2. `OwnerOnly` — user can only see their own records
##### 9.3.3. `RoleFilter` — additional filter predicate applied for a given role
##### 9.3.4. `DepartmentScope` — records visible within the user's department subtree

#### 9.4. Writing Custom Privacy Policies
##### 9.4.1. The `Policy` interface
##### 9.4.2. Accessing tenant and user context inside a policy
##### 9.4.3. Returning additional filter predicates
##### 9.4.4. Returning field masks

#### 9.5. Composing Policies
##### 9.5.1. `privacy.And` — all policies must pass
##### 9.5.2. `privacy.Or` — at least one policy must pass
##### 9.5.3. `privacy.Not` — inversion
##### 9.5.4. Execution order and short-circuit behaviour

#### 9.6. Testing Privacy Policies
##### 9.6.1. Unit testing a policy with a mock context
##### 9.6.2. Integration testing — verifying rows are filtered correctly end-to-end
##### 9.6.3. Common mistakes — policies that silently pass everything

---

### 10. Custom Fields — Runtime Schema Extension

#### 10.1. The Custom Field Model
##### 10.1.1. How custom fields extend both system entities and custom entities
##### 10.1.2. Storage — the `custom_fields JSONB` column pattern on system entities
##### 10.1.3. The `CustomFieldDef` platform entity — metadata table
##### 10.1.4. Scope — custom fields are per-tenant, per-entity

#### 10.2. Defining Custom Fields
##### 10.2.1. Via the admin UI — field type, label, options, validation rules
##### 10.2.2. Via the API — `POST /api/v1/custom-field-defs`
##### 10.2.3. Via fixture files — for module developers shipping pre-configured fields
##### 10.2.4. Supported field types for custom fields — subset of the full field type list

#### 10.3. Validation Rules on Custom Fields
##### 10.3.1. Required, max length, regex — declared in `CustomFieldDef`
##### 10.3.2. How custom field validators are loaded and executed
##### 10.3.3. Validation error reporting — same field-level error format as system fields

#### 10.4. Querying Custom Fields
##### 10.4.1. JSONB path predicates in the Filter DSL
##### 10.4.2. GIN index strategy for custom fields — which paths to index
##### 10.4.3. Performance characteristics — when JSONB queries are acceptable, when to promote to a column

#### 10.5. Surfacing Custom Fields in the SDUI Layer
##### 10.5.1. Page builder reads `CustomFieldDef` records at render time
##### 10.5.2. Automatic form field injection — ordering and section placement
##### 10.5.3. Tenant control over field placement via UI metadata on `CustomFieldDef`

#### 10.6. Custom Field Lifecycle
##### 10.6.1. Adding a custom field — no migration required for system entities
##### 10.6.2. Renaming a custom field — key stability vs label changes
##### 10.6.3. Changing a field type — data migration implications
##### 10.6.4. Deprecating and removing a custom field — soft removal first
##### 10.6.5. Promoting a custom field to a system field — when and how

---

### 11. Database Migrations

#### 11.1. Migration Strategy Overview
##### 11.1.1. Why manual reviewed migrations, not auto-migrate, in production
##### 11.1.2. `golang-migrate` as the migration tool — PostgreSQL driver, versioned SQL files
##### 11.1.3. The migration lifecycle — write → review → test → apply → verify
##### 11.1.4. RLS-aware migrations — DDL is applied once to the shared schema; no per-tenant DDL is ever needed

#### 11.2. Migration File Format
##### 11.2.1. File naming — `{timestamp}_{description}.up.sql` and `{timestamp}_{description}.down.sql`
##### 11.2.2. Timestamp format — Unix epoch in seconds (14-digit): `20240615120000_add_customer_tier.up.sql`
##### 11.2.3. Up migration — forward SQL: CREATE TABLE, ALTER TABLE, CREATE INDEX, etc.
##### 11.2.4. Down migration — rollback SQL: DROP TABLE, ALTER TABLE DROP COLUMN, DROP INDEX (required, not optional)
##### 11.2.5. Editing a migration after it has been applied — unsafe; write a new migration instead

#### 11.3. CLI Commands
##### 11.3.1. `awo migrate up` — apply all pending migrations
##### 11.3.2. `awo migrate down [N]` — roll back N migrations (default 1)
##### 11.3.3. `awo migrate version` — print the current applied migration version
##### 11.3.4. `awo migrate create <name>` — create a new timestamped migration file pair
##### 11.3.5. `awo migrate force <V>` — mark a specific version as applied without running SQL (emergency use)

#### 11.4. RLS Migration Pattern
##### 11.4.1. Single shared schema — all tenants' data lives in the same tables
##### 11.4.2. Every table must include a `tenant_id uuid NOT NULL` column
##### 11.4.3. RLS policy migration — enabling RLS and creating policies alongside table DDL
##### 11.4.4. No per-tenant migration — applying a migration once applies it for all tenants
##### 11.4.5. Per-tenant seed data — seeded via the provisioning workflow, not via migrations

#### 11.5. Rolling Migrations — Zero-Downtime Patterns
##### 11.5.1. Expand-contract pattern — adding nullable columns before making them required
##### 11.5.2. Multi-phase column renames — add new, dual-write, remove old
##### 11.5.3. Index creation with `CREATE INDEX CONCURRENTLY`
##### 11.5.4. Operations that always require downtime — and how to schedule them

#### 11.6. Drift Detection
##### 11.6.1. What drift is — manual DB changes that are not reflected in migration files
##### 11.6.2. Running drift detection in CI — comparing applied schema against migration-derived schema
##### 11.6.3. Resolving drift — generating a catch-up migration vs reverting the manual change

#### 11.7. Complex DDL — Views, Functions, and Triggers
##### 11.7.1. SQL views — when to use them (reporting, denormalised read models, cross-entity joins)
##### 11.7.2. Embedding views in migration files — `CREATE OR REPLACE VIEW` in a `.up.sql` file
##### 11.7.3. Dropping views in down migrations — always `DROP VIEW IF EXISTS`
##### 11.7.4. Stored functions — `set_tenant_context(uuid)` for RLS, naming series counter atomics
##### 11.7.5. `updated_at` auto-update trigger — single trigger function shared across all tables
##### 11.7.6. Audit log trigger — fires on UPDATE/DELETE, writes to audit_log with old and new values
##### 11.7.7. RLS policies — `CREATE POLICY`, `ALTER TABLE ... ENABLE ROW LEVEL SECURITY`
##### 11.7.8. Pattern: keep all DDL in numbered migration files — never apply DDL manually in production

#### 11.8. Testing Migrations
##### 11.8.1. Migration test helper — applies all migrations to a temporary database
##### 11.8.2. Round-trip test — up then down, verify schema returns to prior state
##### 11.8.3. Dangerous migration detection in CI — flag `DROP COLUMN`, `DROP TABLE`, non-concurrent index on large tables

#### 11.9. Migration Rollback Procedures
##### 11.9.1. Running `awo migrate down N`
##### 11.9.2. When rollback is not possible — destructive changes and recovery options
##### 11.9.3. Emergency `migrate force` — marking a migration applied after manual fix

---

## Part III — The API Layer

### 12. Server Setup

#### 12.1. Entry Point and Process Structure
##### 12.1.1. `main.go` — bootstrap sequence
##### 12.1.2. Dependency injection — what is wired at startup
##### 12.1.3. Graceful shutdown — signal handling, drain timeout

#### 12.2. Configuration Loading
##### 12.2.1. Environment variables — the `AWO_*` namespace
##### 12.2.2. Config struct — typed, validated at startup, no runtime access by key string
##### 12.2.3. Environment-specific overrides — development vs staging vs production
##### 12.2.4. Secrets — never in config files, loaded from environment or secret manager

#### 12.3. TLS and Reverse Proxy
##### 12.3.1. Caddy as the recommended reverse proxy — automatic TLS
##### 12.3.2. X-Forwarded-For and real IP extraction
##### 12.3.3. Trusted proxy configuration — why getting this wrong is a security issue

#### 12.4. Health and Readiness Endpoints
##### 12.4.1. `GET /health/live` — process is alive
##### 12.4.2. `GET /health/ready` — process is ready to serve traffic
##### 12.4.3. Deep health check — database, Redis, Temporal reachability
##### 12.4.4. Health check authentication — should it be open or protected

---

### 13. Middleware Pipeline

#### 13.1. Middleware Execution Order
##### 13.1.1. Why order matters — request ID must come before logging
##### 13.1.2. The canonical order — annotated list

#### 13.2. Request ID Middleware
##### 13.2.1. Generating a UUID request ID
##### 13.2.2. Honouring an incoming `X-Request-ID` header — trust rules
##### 13.2.3. Propagating the request ID to all downstream logs and traces

#### 13.3. Structured Logging Middleware
##### 13.3.1. `slog` integration — fields attached to every request log
##### 13.3.2. What is always logged — method, path, status, duration, request ID, tenant ID
##### 13.3.3. What is never logged — request bodies, `Sensitive` fields, credentials
##### 13.3.4. Slow request threshold — logging full context for requests over budget

#### 13.4. Panic Recovery Middleware
##### 13.4.1. Converting panics to 500 responses
##### 13.4.2. Logging the stack trace with request context
##### 13.4.3. Alerting on panic rate — threshold-based

#### 13.5. CORS Middleware
##### 13.5.1. Allowed origins — amis frontend origin, tenant subdomains
##### 13.5.2. Preflight caching
##### 13.5.3. Credentials — why `AllowCredentials` requires explicit origin (not `*`)

#### 13.6. Rate Limiting Middleware
##### 13.6.1. Per-tenant rate limiting — Redis sliding window
##### 13.6.2. Per-user rate limiting — layered on top of tenant limit
##### 13.6.3. Rate limit headers — `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `Retry-After`
##### 13.6.4. Rate limit bypass for internal service calls

---

### 14. Multi-Tenancy Middleware

#### 14.1. Tenant Identification
##### 14.1.1. Subdomain parsing — `{tenant}.awo.app` pattern
##### 14.1.2. `X-Awo-Tenant` header fallback — for API clients and mobile
##### 14.1.3. Query parameter fallback — development only, disabled in production
##### 14.1.4. Tenant not found — 404 vs 400, and why it matters for information leakage

#### 14.2. Tenant Context Propagation
##### 14.2.1. What goes into the tenant context — ID, RLS context variable, config, feature flags
##### 14.2.2. Propagating tenant context through `context.Context`
##### 14.2.3. Extracting tenant context in route handlers and hooks
##### 14.2.4. Propagating tenant context into Temporal workflow starts

#### 14.3. Tenant State Handling
##### 14.3.1. Active tenant — normal flow; `set_tenant_context()` sets RLS variable, all queries proceed
##### 14.3.2. Suspended tenant — 402 with suspension reason; `set_tenant_context()` raises exception if status != ACTIVE
##### 14.3.3. Trial tenant — feature flag restrictions applied
##### 14.3.4. Archived tenant — 410 Gone; middleware rejects before reaching DB

#### 14.4. RLS Context and Connection Management
##### 14.4.1. `set_tenant_context(uuid)` stored procedure — sets `app.current_tenant_id` session variable
##### 14.4.2. Called on every connection acquisition before executing any query
##### 14.4.3. Connection pool behaviour — RLS context reset between requests
##### 14.4.4. Why RLS is a defence-in-depth measure — application layer also scopes queries

---

### 15. Authentication and Session Management

#### 15.1. Session-Based Authentication Rationale
##### 15.1.1. Why session cookies over JWT for browser clients
##### 15.1.2. Where JWT is appropriate — mobile, inter-service, webhook delivery
##### 15.1.3. The session/JWT hybrid — browser gets a session, mobile gets a token

#### 15.2. Login Flow
##### 15.2.1. Credential submission and verification — password hash comparison
##### 15.2.2. MFA challenge — TOTP and backup codes
##### 15.2.3. Session creation — generating a session ID, storing in Redis
##### 15.2.4. Session cookie attributes — `HttpOnly`, `Secure`, `SameSite=Lax`, `Domain`, `Path`, `Max-Age`
##### 15.2.5. Logging the login event — IP, user agent, timestamp

#### 15.3. Session Validation Middleware
##### 15.3.1. Cookie extraction and session ID validation
##### 15.3.2. Redis lookup — session data structure
##### 15.3.3. Session expiry — absolute vs sliding expiry
##### 15.3.4. Session not found — 401 vs redirect to login
##### 15.3.5. The 401 interception flow in amis — re-auth overlay without page reload

#### 15.4. Session Invalidation
##### 15.4.1. Logout — session deletion from Redis
##### 15.4.2. Forced invalidation — all sessions for a user (password change, role change)
##### 15.4.3. Concurrent session limits — configurable per tenant
##### 15.4.4. Idle timeout vs absolute timeout

#### 15.5. Password Management
##### 15.5.1. Password hashing — bcrypt with cost factor configuration
##### 15.5.2. Password reset flow — time-limited token, email delivery
##### 15.5.3. Password policy enforcement — length, complexity, reuse prevention
##### 15.5.4. Credential stuffing protection — rate limiting + lockout

---

### 16. Role-Based Access Control

#### 16.1. Role Model
##### 16.1.1. System roles — defined in framework code, apply across all tenants
##### 16.1.2. Tenant roles — defined per tenant, extend or restrict system roles
##### 16.1.3. Role hierarchy — how permissions accumulate
##### 16.1.4. The superuser role — what it bypasses and what it cannot bypass

#### 16.2. Permission Model
##### 16.2.1. Permission levels — `create`, `read`, `write`, `delete`, `submit`, `cancel`, `amend`
##### 16.2.2. Per-EntityDefinition permissions — the permission matrix
##### 16.2.3. Field-level permissions — read-only, hidden, write-restricted per role
##### 16.2.4. Action permissions — custom actions beyond CRUD

#### 16.3. Permission Evaluation
##### 16.3.1. `RequireRole` middleware — guard at the route level
##### 16.3.2. Per-route permission annotation on EntityDefinition
##### 16.3.3. Permission resolution order — system defaults → tenant overrides → user overrides
##### 16.3.4. Merging permissions from multiple roles — union vs intersection
##### 16.3.5. Permission caching — per-request cache, invalidation on role change

#### 16.4. Surfacing Permissions in the SDUI Layer
##### 16.4.1. The `Permissions` struct passed to every page builder
##### 16.4.2. `CanView`, `CanEdit`, `CanDelete`, `CanSubmit`, `CanCancel` methods
##### 16.4.3. Conditional visibility expressions in amis JSON
##### 16.4.4. Hiding sections vs disabling fields — UX guidelines

#### 16.5. Auditing Permission Checks
##### 16.5.1. What is logged — entity, action, user, role, outcome
##### 16.5.2. Denied access log — security monitoring
##### 16.5.3. Permission audit report for compliance

---

### 17. REST API Conventions

#### 17.1. URL Structure
##### 17.1.1. `/api/v1/{entity-name}` — auto-generated CRUD
##### 17.1.2. `/api/v1/{entity-name}/{id}` — single resource
##### 17.1.3. `/api/v1/{entity-name}/{id}/{action}` — entity actions (submit, cancel)
##### 17.1.4. `/api/v1/{entity-name}/{id}/pdf` — PDF generation endpoint
##### 17.1.5. Tenant context comes from middleware, not the URL
##### 17.1.6. Custom routes — convention for module-specific endpoints

#### 17.2. Standard Response Envelope
##### 17.2.1. Success shape — `{ data, meta }`
##### 17.2.2. Error shape — `{ errors: [{ field, code, message }] }`
##### 17.2.3. Paginated list shape — `{ data: [], meta: { total, cursor, has_next } }`
##### 17.2.4. Why a consistent envelope matters for amis auto-wiring

#### 17.3. Filtering
##### 17.3.1. Query string filter convention — `filter[field][op]=value`
##### 17.3.2. Logical operators in query strings — `filter[_or]`, `filter[_and]`
##### 17.3.3. JSONB path filters for custom fields
##### 17.3.4. Full-text search — `q=` parameter, `pg_trgm` backed

#### 17.4. Sorting and Pagination
##### 17.4.1. `sort=field,-other_field` — ascending/descending convention
##### 17.4.2. Cursor pagination — `cursor=` and `limit=` parameters
##### 17.4.3. Offset pagination — `offset=` and `limit=` — reports only, why
##### 17.4.4. Default page size, maximum page size

#### 17.5. Bulk Operations
##### 17.5.1. `POST /api/v1/{entity}/bulk-create`
##### 17.5.2. `POST /api/v1/{entity}/bulk-update`
##### 17.5.3. `POST /api/v1/{entity}/bulk-delete`
##### 17.5.4. Partial success handling — which records failed and why
##### 17.5.5. Bulk operation size limits

#### 17.6. Idempotency
##### 17.6.1. `Idempotency-Key` header — how to use it
##### 17.6.2. How the server deduplicates — Redis key with response cache
##### 17.6.3. Idempotency key TTL
##### 17.6.4. What operations require idempotency keys — non-GET mutations

---

### 18. Error Handling

#### 18.1. Error Type Hierarchy
##### 18.1.1. `ValidationError` — field-level, user-correctable
##### 18.1.2. `NotFoundError` — entity or record does not exist
##### 18.1.3. `PermissionError` — role lacks required permission
##### 18.1.4. `ConflictError` — uniqueness violation, optimistic lock failure
##### 18.1.5. `BusinessRuleError` — hook rejected the operation
##### 18.1.6. `WorkflowError` — Temporal activity or workflow failed
##### 18.1.7. `InternalError` — unhandled, logged, not exposed to client

#### 18.2. HTTP Status Code Mapping
##### 18.2.1. `ValidationError` → 422 Unprocessable Entity
##### 18.2.2. `NotFoundError` → 404 Not Found
##### 18.2.3. `PermissionError` → 403 Forbidden (authenticated) / 401 Unauthorized (not authenticated)
##### 18.2.4. `ConflictError` → 409 Conflict
##### 18.2.5. `BusinessRuleError` → 422 Unprocessable Entity with error code
##### 18.2.6. `WorkflowError` → 202 Accepted (workflow is async) vs 500 (start failed)
##### 18.2.7. `InternalError` → 500 Internal Server Error

#### 18.3. Field-Level Validation Errors for amis
##### 18.3.1. Error shape expected by amis — `{ name: "field_name", errors: ["message"] }`
##### 18.3.2. Mapping EntityRepository constraint errors to field errors
##### 18.3.3. Mapping hook-returned errors to field errors
##### 18.3.4. Top-level (non-field) errors — how amis renders them

#### 18.4. Localisation of Error Messages
##### 18.4.1. Error codes as the stable key — not message strings
##### 18.4.2. Message catalogue — Go `embed` for locale files
##### 18.4.3. Swahili error messages for the Kenyan market
##### 18.4.4. Fallback chain — tenant locale → system locale → English

---

### 19. API Versioning

#### 19.1. Versioning Strategy
##### 19.1.1. URI prefix versioning — `/api/v1/`, `/api/v2/`
##### 19.1.2. Version lifetime policy
##### 19.1.3. What constitutes a breaking change

#### 19.2. Deprecation Lifecycle
##### 19.2.1. `Deprecation` and `Sunset` response headers
##### 19.2.2. Deprecation notice period
##### 19.2.3. Communication — changelog, email, in-app notice
##### 19.2.4. Removal process

#### 19.3. Backward Compatibility Rules
##### 19.3.1. Safe changes — adding optional fields, adding new endpoints
##### 19.3.2. Unsafe changes — removing fields, changing field types, changing error codes
##### 19.3.3. The contract test suite — automatically detecting backward-incompatible changes

---

### 20. Feature Flags

#### 20.1. Feature Flag Service
##### 20.1.1. Flag definition — name, type (boolean, string, percentage), default value
##### 20.1.2. Redis storage layout for flags
##### 20.1.3. Flag loading at tenant boot — cached per request

#### 20.2. Flag Evaluation
##### 20.2.1. Evaluation order — system default → tenant override → user override
##### 20.2.2. Percentage rollout — stable hashing on tenant ID + flag name
##### 20.2.3. Injecting flags into Fiber context for use in handlers and page builders

#### 20.3. Flag Lifecycle
##### 20.3.1. Draft → active → deprecated → removed
##### 20.3.2. Dead flag cleanup — automated detection of flags with 0% or 100% rollout
##### 20.3.3. Flag dependency graph — enabling flag A requires flag B

#### 20.4. Feature Flags in the SDUI Layer
##### 20.4.1. Page builders receive the resolved flag set
##### 20.4.2. Conditionally including or excluding UI blocks based on flags
##### 20.4.3. Flag-gated menu items and navigation

---

## Part IV — The SDUI Layer

### 21. Server-Driven UI Philosophy

#### 21.1. Why SDUI for an ERP Framework
##### 21.1.1. The frontend team bottleneck problem
##### 21.1.2. How amis JSON becomes a rendered UI without a frontend build step
##### 21.1.3. The SDUI contract — what the server guarantees, what the client can rely on
##### 21.1.4. When SDUI is appropriate and when it is not
##### 21.1.5. amis version pinning — why and how

#### 21.2. The Page Builder Pipeline
##### 21.2.1. Step 1 — HTTP request arrives at `/api/v1/pages/{page-name}`
##### 21.2.2. Step 2 — tenant, user, permissions, and feature flags resolved
##### 21.2.3. Step 3 — page builder function invoked with resolved context
##### 21.2.4. Step 4 — amis JSON assembled from blocks
##### 21.2.5. Step 5 — page definition cached in Redis with TTL
##### 21.2.6. Step 6 — JSON returned to amis client for rendering
##### 21.2.7. Cache invalidation — when and how

#### 21.3. Page Builder Functions
##### 21.3.1. Function signature — inputs and outputs
##### 21.3.2. Naming convention — `Build{Entity}{View}Page`
##### 21.3.3. What goes in a page builder — only presentation decisions, no business logic
##### 21.3.4. Registering a page builder with the EntityDefinition
##### 21.3.5. Testing page builders — snapshot tests on the JSON output

---

### 22. Foundation Components Reference

#### 22.1. Text and Data Fields
##### 22.1.1. `TextField` — options, placeholder, character counter, masks, read-only mode
##### 22.1.2. `NumberField` — precision, thousand separator, currency prefix (KES)
##### 22.1.3. `TextAreaField` — rows, resize behaviour, character limit display
##### 22.1.4. `RichTextField` — when to use vs TextAreaField

#### 22.2. Choice Fields
##### 22.2.1. `SelectField` — static options, API-sourced options, clearable
##### 22.2.2. `MultiSelectField` — tag input, max selection limit
##### 22.2.3. `RadioGroupField` — when to prefer over Select
##### 22.2.4. `SwitchField` and `CheckboxField` — boolean input patterns

#### 22.3. Date and Time Fields
##### 22.3.1. `DateField` — format, min/max, disable weekends
##### 22.3.2. `DateRangeField` — linked start/end pickers
##### 22.3.3. `DateTimeField` — timezone handling for East Africa (EAT)
##### 22.3.4. `TimeField`

#### 22.4. Relational Fields
##### 22.4.1. `LinkField` — linked EntityDefinition picker, search-as-you-type
##### 22.4.2. `LinkField` API source — how the search endpoint is wired
##### 22.4.3. `DynamicLinkField` — entity type selector + ID picker

#### 22.5. File and Media Fields
##### 22.5.1. `FileUploadField` — accepted types, size limit, multiple files
##### 22.5.2. `ImageUploadField` — preview, crop, aspect ratio constraints

#### 22.6. ERP-Specific Components
##### 22.6.1. `LineItemEditor` — child table editor for order lines, invoice lines
##### 22.6.2. `BarcodeField` — camera scan + manual entry, EAN-13 and QR
##### 22.6.3. `SerialBatchPicker` — serial number and batch selection for inventory
##### 22.6.4. `CurrencyField` — KES formatting, multi-currency display
##### 22.6.5. `NamingSeriesField` — displays the generated series, allows prefix override

---

### 23. Composite Blocks Reference

#### 23.1. `AmisCRUD` — The Full List + Form Block
##### 23.1.1. When to use AmisCRUD vs separate list and form pages
##### 23.1.2. Configuration — columns, filters, form layout, row actions
##### 23.1.3. Inline editing vs modal form vs separate page
##### 23.1.4. Toolbar configuration — create button, bulk actions, export

#### 23.2. `AmisForm` — Standalone Form
##### 23.2.1. Sections and tabs — when each is appropriate
##### 23.2.2. Conditional field visibility — amis expression syntax
##### 23.2.3. Field dependencies — reacting to another field's value
##### 23.2.4. Form-level validation vs field-level validation
##### 23.2.5. Submit and reset behaviour

#### 23.3. `AmisList` — Read-Only Table
##### 23.3.1. Column configuration — field, label, width, sortable, formatter
##### 23.3.2. Row-level actions — edit, delete, custom
##### 23.3.3. Inline filters — quick filter bar above the table
##### 23.3.4. Column visibility toggler

#### 23.4. `AmisDetailPage` — Master + Related Tables
##### 23.4.1. Header section — primary entity fields
##### 23.4.2. Tab panels — related entities as child tables
##### 23.4.3. Action toolbar — submit, cancel, amend, print, email, download PDF

#### 23.5. `AmisDashboard` — Stat Cards and Charts
##### 23.5.1. KPI stat cards — value, label, trend indicator
##### 23.5.2. Chart blocks — bar, line, pie, area — wired to API endpoints
##### 23.5.3. Date range filter wired to all blocks simultaneously
##### 23.5.4. Refresh interval for live dashboards

#### 23.6. `AmisKanban` — Status Column Board
##### 23.6.1. Column definition — status values as columns
##### 23.6.2. Card fields — which entity fields appear on each card
##### 23.6.3. Drag-to-update — triggering a status change action
##### 23.6.4. Column count limits and card overflow

#### 23.7. `AmisWizard` — Multi-Step Form
##### 23.7.1. Step definition — title, fields, validation guard
##### 23.7.2. Conditional step skipping
##### 23.7.3. Data passing between steps
##### 23.7.4. Final step submission — wiring to the create or update endpoint

#### 23.8. `AmisReportPage` — Parameterised Report
##### 23.8.1. Parameter form — date range, entity filters, grouping options
##### 23.8.2. Results table — sortable, exportable
##### 23.8.3. Summary row — totals, averages
##### 23.8.4. Chart section — visualising the report data

---

### 24. Theming and Branding Per Tenant

#### 24.1. CSS Variable Injection
##### 24.1.1. The `/api/v1/theme.css` endpoint — per-tenant CSS
##### 24.1.2. Overridable variables — primary colour, font, border radius
##### 24.1.3. How amis loads the theme endpoint

#### 24.2. Logo and Colour Configuration
##### 24.2.1. Logo upload and storage
##### 24.2.2. Colour picker and validation — contrast ratio enforcement
##### 24.2.3. Dark mode support — automatic or tenant-controlled

#### 24.3. Print Stylesheets
##### 24.3.1. Print layout for invoices
##### 24.3.2. Print layout for delivery notes and purchase orders
##### 24.3.3. Relationship to the server-side PDF generation system (see §25)

---

### 25. Print and PDF Generation

#### 25.1. Overview and Approach
##### 25.1.1. Server-side PDF via `chromedp` — headless Chrome controlled from Go
##### 25.1.2. Why not wkhtmltopdf — abandoned project, security vulnerabilities, no modern CSS support
##### 25.1.3. Why not a PDF library — poor font rendering, no CSS layout engine, complex table handling

#### 25.2. HTML Print Templates
##### 25.2.1. Go `html/template` — one template per document type
##### 25.2.2. Template types: invoice, delivery note, payslip, purchase order, statement of account
##### 25.2.3. Template registration on EntityDefinition — the optional `PrintTemplate` field
##### 25.2.4. Template data binding — entity record + related edges passed as template data
##### 25.2.5. Tenant branding injection — logo URL, colours, address from TenantConfig
##### 25.2.6. `@media print` CSS embedded in template — page size, margins, header, footer
##### 25.2.7. Fonts — system fonts or embedded base64 for consistent cross-environment output

#### 25.3. The PDF Endpoint
##### 25.3.1. `GET /api/v1/{entity}/{id}/pdf` — render template, convert to PDF, stream response
##### 25.3.2. Content-Type: `application/pdf`; Content-Disposition: inline or attachment
##### 25.3.3. Permission check — same permission as `read` on the entity
##### 25.3.4. Cache header — `ETag` based on record `updated_at`

#### 25.4. PDF Caching
##### 25.4.1. Generate once, store in object storage (S3-compatible)
##### 25.4.2. Cache key — `{tenant_id}/{entity}/{id}/{updated_at_hash}.pdf`
##### 25.4.3. Cache invalidation — on record update, delete cached PDF
##### 25.4.4. Serving from cache — pre-signed URL redirect vs proxy through API

#### 25.5. Background PDF Generation
##### 25.5.1. Temporal activity for large or batch documents — payslip run, bulk statements
##### 25.5.2. `GeneratePDFActivity` — renders template, uploads to object storage, returns S3 URI
##### 25.5.3. Notification to user when batch PDF is ready — in-app notification with download link
##### 25.5.4. Concurrency — chromedp instances are pooled, not spawned per request

#### 25.6. CSS Page Layout
##### 25.6.1. `@page` rule — size (A4), margins
##### 25.6.2. `page-break-before` / `page-break-after` — controlling breaks in multi-page documents
##### 25.6.3. Header and footer via CSS `@page` margin boxes
##### 25.6.4. Table continuation across pages — `thead` repeat behaviour

---

### 26. Custom Renderers

#### 26.1. When a Custom Renderer Is Justified
##### 26.1.1. Capability gap — amis built-in components cannot express the UI needed
##### 26.1.2. Performance gap — a custom renderer can be more efficient
##### 26.1.3. When to push back — over-customisation cost

#### 26.2. Registering a Custom Renderer
##### 26.2.1. The amis custom component API
##### 26.2.2. Serving the custom renderer JS bundle
##### 26.2.3. Registering the renderer type string

#### 26.3. Worked Examples
##### 26.3.1. Forecourt pump status renderer — real-time pump state display
##### 26.3.2. Custom chart renderer with recharts
##### 26.3.3. Map renderer — site locations with leaflet

---

## Part V — The Workflow Engine

### 27. Temporal Fundamentals for Awo

#### 27.1. Why Temporal Over Queues or Cron
##### 27.1.1. Durability — workflow state survives process crashes
##### 27.1.2. Audit trail — complete execution history queryable
##### 27.1.3. Long-running transactions — approval chains that wait days
##### 27.1.4. Saga pattern — compensating transactions without distributed locks

#### 27.2. Core Temporal Concepts
##### 27.2.1. Workflow — a durable function that orchestrates activities
##### 27.2.2. Activity — a function that interacts with the outside world
##### 27.2.3. Worker — a process that executes workflows and activities
##### 27.2.4. Task queue — the named channel that routes work to workers
##### 27.2.5. Workflow ID — the unique, stable identity of a workflow instance
##### 27.2.6. Temporal namespace — the isolation boundary for workflow history

#### 27.3. Awo Worker Process Setup
##### 27.3.1. Running the worker alongside Fiber vs as a separate process
##### 27.3.2. Task queue naming conventions — `{module}.{entity}.{action}`
##### 27.3.3. Worker concurrency configuration
##### 27.3.4. Registering workflows and activities on the worker

---

### 28. Defining Workflows

#### 28.1. Workflow Function Anatomy
##### 28.1.1. Function signature — `(ctx workflow.Context, input WorkflowInput) (WorkflowResult, error)`
##### 28.1.2. What is allowed inside a workflow — determinism rules
##### 28.1.3. What is forbidden inside a workflow — system calls, random, time.Now()
##### 28.1.4. Using `workflow.Now()`, `workflow.Sleep()`, `workflow.SideEffect()`

#### 28.2. Workflow Options
##### 28.2.1. `WorkflowExecutionTimeout` — the hard outer limit
##### 28.2.2. `WorkflowRunTimeout` — per-run limit (relevant after continue-as-new)
##### 28.2.3. `WorkflowTaskTimeout` — how long a workflow task can take to process
##### 28.2.4. Retry policy — is retrying a workflow appropriate or should only activities retry
##### 28.2.5. Search attributes — indexable metadata for the Temporal Web UI and queries

#### 28.3. Workflow Versioning
##### 28.3.1. Why workflow code changes break running workflows
##### 28.3.2. `workflow.GetVersion()` — the safe way to add branching logic to existing workflows
##### 28.3.3. Version deprecation — removing old branches after all running instances complete
##### 28.3.4. The `continue-as-new` pattern — for workflows that run indefinitely

#### 28.4. Child Workflows
##### 28.4.1. When to decompose into child workflows
##### 28.4.2. `workflow.ExecuteChildWorkflow()` — fire and wait vs fire and forget
##### 28.4.3. Parent-child cancellation propagation
##### 28.4.4. Workflow ID namespacing for child workflows

#### 28.5. Workflow ID Naming Conventions
##### 28.5.1. The `{tenant}.{entity}.{id}.{action}` pattern
##### 28.5.2. Idempotency via workflow ID — re-triggering is safe
##### 28.5.3. Workflow ID uniqueness enforcement — `AllowDuplicate` vs `RejectDuplicate` policies

---

### 29. Defining Activities

#### 29.1. Activity Function Anatomy
##### 29.1.1. Function signature — `(ctx context.Context, input ActivityInput) (ActivityResult, error)`
##### 29.1.2. The `Activities` struct pattern — dependency injection for activities
##### 29.1.3. Accessing the EntityRepository inside an activity
##### 29.1.4. Accessing the tenant context inside an activity

#### 29.2. Activity Options
##### 29.2.1. `StartToCloseTimeout` — required, always set explicitly
##### 29.2.2. `ScheduleToStartTimeout` — queue wait timeout
##### 29.2.3. `HeartbeatTimeout` — for activities that must heartbeat
##### 29.2.4. Retry policy — `MaxAttempts`, `InitialInterval`, `BackoffCoefficient`, `MaxInterval`, `NonRetryableErrorTypes`

#### 29.3. Heartbeating
##### 29.3.1. When heartbeating is required — long-running, blocking, or batch activities
##### 29.3.2. `activity.RecordHeartbeat(ctx, details)` — what to pass as details
##### 29.3.3. Detecting cancellation via heartbeat context
##### 29.3.4. Heartbeat-based progress reporting

#### 29.4. Idempotent Activities
##### 29.4.1. Why activities must be safe to retry
##### 29.4.2. Database upsert pattern for idempotent writes
##### 29.4.3. External API idempotency keys
##### 29.4.4. Testing retry behaviour

#### 29.5. Local Activities
##### 29.5.1. What local activities are — in-process, no history event
##### 29.5.2. When to use local activities vs regular activities
##### 29.5.3. Local activity limitations — no heartbeat, no separate retry from workflow

---

### 30. Saga Pattern — Compensating Transactions

#### 30.1. When to Use a Saga
##### 30.1.1. Multi-service or multi-table operations that must be atomic in business terms
##### 30.1.2. Operations where database-level transactions are insufficient
##### 30.1.3. Long-running operations where holding a DB transaction is not feasible

#### 30.2. Saga Implementation in Temporal
##### 30.2.1. Forward steps — activities that make changes
##### 30.2.2. Compensation stack — appending compensations as forward steps succeed
##### 30.2.3. Backward recovery — running compensations in reverse on failure
##### 30.2.4. The `saga` helper — Awo's built-in compensation manager

#### 30.3. Designing Compensations
##### 30.3.1. Compensations are not rollbacks — they are forward-moving corrections
##### 30.3.2. Idempotency requirement — compensations must be safe to re-run
##### 30.3.3. Compensation failure — when a compensation itself fails
##### 30.3.4. Storing compensation state — what to record for observability

#### 30.4. Worked Examples
##### 30.4.1. Sales order submit saga
##### 30.4.2. Fuel delivery reconciliation saga

---

### 31. Signals, Queries, and Human-in-the-Loop

#### 31.1. Signals
##### 31.1.1. What signals are — external input to a running workflow
##### 31.1.2. Sending a signal from a Fiber handler
##### 31.1.3. Receiving a signal inside a workflow — `workflow.GetSignalChannel()`
##### 31.1.4. Signal delivery guarantees — at-least-once, workflow must be idempotent on signal
##### 31.1.5. Signal naming conventions

#### 31.2. Queries
##### 31.2.1. What queries are — synchronous read of workflow state, no side effects
##### 31.2.2. Defining a query handler inside a workflow
##### 31.2.3. Calling a query from a Fiber handler
##### 31.2.4. What to expose via queries — current stage, pending approvals, error details

#### 31.3. Approval Gate Pattern
##### 31.3.1. The `WaitForApproval` activity — workflow pauses at a signal channel
##### 31.3.2. The approval API endpoint — validates permission, sends signal
##### 31.3.3. Approval timeout — what happens if no signal arrives within the deadline
##### 31.3.4. Rejection path — separate signal value, compensation triggered

#### 31.4. Worked Example — Shift Close Approval Flow
##### 31.4.1. Cashier submits shift close request
##### 31.4.2. Workflow starts, notifies supervisor
##### 31.4.3. Supervisor reviews via AmisDetailPage
##### 31.4.4. Supervisor approves — signal sent, GL posting activity runs
##### 31.4.5. Supervisor rejects — signal sent, shift re-opened

---

### 32. Schedules and Cron

#### 32.1. Temporal Schedules vs Traditional Cron
##### 32.1.1. Why Temporal schedules — durability, catch-up, observability
##### 32.1.2. What Temporal schedules cannot do that cron can
##### 32.1.3. Schedule vs cron expression — calendar-based vs interval-based

#### 32.2. Defining a Schedule
##### 32.2.1. Schedule ID — naming convention
##### 32.2.2. Schedule spec — interval, calendar, jitter
##### 32.2.3. Schedule action — which workflow to start and with what input
##### 32.2.4. Overlap policy — skip, buffer, allow, terminate previous

#### 32.3. Built-in Scheduled Workflows
##### 32.3.1. Nightly GL period check
##### 32.3.2. Daily wetstock reconciliation
##### 32.3.3. Weekly KPI snapshot
##### 32.3.4. Monthly tenant usage report

#### 32.4. Managing Schedules
##### 32.4.1. Creating schedules at service start — idempotent upsert
##### 32.4.2. Pausing and resuming schedules
##### 32.4.3. Triggering a schedule manually — backfill and forced trigger
##### 32.4.4. Deleting a schedule

---

### 33. Workflow Observability

#### 33.1. Temporal Web UI
##### 33.1.1. Reading workflow event history — what each event type means
##### 33.1.2. Searching workflows by search attribute
##### 33.1.3. Workflow status states — Running, Completed, Failed, Timed Out, Cancelled, Terminated

#### 33.2. OpenTelemetry Integration
##### 33.2.1. Trace propagation from Fiber into Temporal activities
##### 33.2.2. Custom spans inside activities
##### 33.2.3. Correlating Temporal workflow ID with distributed traces

#### 33.3. Workflow Alerting
##### 33.3.1. Alert on workflow failure rate
##### 33.3.2. Alert on workflow pending age — stuck workflows
##### 33.3.3. Alert on schedule not running

---

## Part VI — Platform Entities

### 34. Tenant and Organisation

#### 34.1. The Tenant Entity
##### 34.1.1. Tenant as the top-level isolation boundary
##### 34.1.2. Tenant entity fields — ID, slug, name, domain, plan, status, locale, fiscal settings
##### 34.1.3. No `db_schema` field — there is no per-tenant schema; RLS provides isolation
##### 34.1.4. Status machine — PENDING → ACTIVE → SUSPENDED → ARCHIVED

#### 34.2. OrgNode — Company and Division Hierarchy
##### 34.2.1. OrgNode entity — hierarchical organisational units within a tenant
##### 34.2.2. Node types: Company (top level), Division, Department, Branch, CostCentre
##### 34.2.3. Materialised path storage for efficient ancestor/descendant queries
##### 34.2.4. Linking users and entities to OrgNodes
##### 34.2.5. OrgNode-scoped permissions — restricting a user's view to their branch

#### 34.3. Tenant Provisioning and Lifecycle
##### 34.3.1. Provisioning is pure data — no DDL, no CREATE SCHEMA
##### 34.3.2. Provisioning workflow overview — seed data, admin user, feature flags, activation
##### 34.3.3. Suspension, reactivation, and archival (see Chapter 38 for full detail)

---

### 35. IAM — Users, Roles, and Permissions

#### 35.1. The User Entity
##### 35.1.1. User fields — ID, tenant_id, email, name, password_hash, status, mfa_secret
##### 35.1.2. User status — INVITED, ACTIVE, SUSPENDED, DELETED
##### 35.1.3. User uniqueness — email is unique per tenant, not globally
##### 35.1.4. System users — service accounts that do not have login credentials

#### 35.2. The Role Entity
##### 35.2.1. Role fields — name, description, is_system, tenant_id
##### 35.2.2. Built-in system roles — admin, accountant, cashier, salesperson, hr_manager, store_keeper, viewer
##### 35.2.3. Tenant-defined roles — extending the system roles
##### 35.2.4. Role assignment — UserRole junction entity

#### 35.3. The Permission Entity
##### 35.3.1. Permission granularity — entity × operation
##### 35.3.2. RolePermission junction entity — role ID, entity name, operations bitmask
##### 35.3.3. Field-level permissions — stored as JSON on the RolePermission row
##### 35.3.4. Permission inheritance — tenant role inherits system role permissions, then adds or restricts

#### 35.4. RBAC Evaluation
##### 35.4.1. Evaluation inputs — user ID, tenant ID, entity name, operation
##### 35.4.2. Resolution pipeline — see §16.3 for full detail
##### 35.4.3. Permission caching — evaluated once per request, cached in Fiber context
##### 35.4.4. Superuser bypass — tenant admin bypasses permission checks, not RLS

#### 35.5. The Session Entity
##### 35.5.1. Session fields — ID, user_id, tenant_id, ip, user_agent, created_at, expires_at, last_seen_at
##### 35.5.2. Sessions stored in Redis (fast path) and PostgreSQL (audit trail)
##### 35.5.3. Session invalidation events — logout, forced expiry, tenant suspension

---

### 36. Audit Log

#### 36.1. What the Audit Log Tracks
##### 36.1.1. Every CREATE, UPDATE, DELETE on auditable entities
##### 36.1.2. AuditLog entity fields — id, tenant_id, entity_name, record_id, operation, user_id, timestamp, previous_value, new_value, ip, request_id
##### 36.1.3. Auditable entities — all platform entities plus any module entity that declares `Auditable: true`

#### 36.2. Write Path
##### 36.2.1. DB trigger approach — `audit_log_trigger` fires on INSERT/UPDATE/DELETE, writes to audit_log table
##### 36.2.2. Application hook approach — `after_save` hook writes via the AuditLog repository
##### 36.2.3. Trade-offs — trigger approach captures all writes including direct SQL; hook approach captures application context (user, request ID)
##### 36.2.4. Awo default — both: trigger for completeness, hook for enriched metadata

#### 36.3. Retention and Search
##### 36.3.1. Retention policy — configurable per tenant, default 7 years (Kenya Companies Act)
##### 36.3.2. Partitioning strategy — audit_log partitioned by month for efficient expiry
##### 36.3.3. Search endpoint — `GET /api/v1/audit-log?entity=&record_id=&user_id=&from=&to=`
##### 36.3.4. Compliance report — full change history for a record

---

### 37. Feature Flags

#### 37.1. The FeatureFlag Entity
##### 37.1.1. FeatureFlag fields — key, name, description, type (boolean/string/percentage), default_value, status
##### 37.1.2. FeatureFlagOverride fields — flag_key, tenant_id, user_id (nullable), value, enabled_at
##### 37.1.3. System flags vs tenant-specific flags

#### 37.2. Evaluation Engine
##### 37.2.1. Evaluation order — system default → tenant override → user override
##### 37.2.2. Percentage rollout — `hash(tenant_id + ":" + flag_key) % 100 < threshold`
##### 37.2.3. Dependency graph — flags that require other flags to be enabled first
##### 37.2.4. Circular dependency detection at boot

#### 37.3. Redis Caching
##### 37.3.1. Cache key — `flags:{tenant_id}` — full flag set as a Redis hash
##### 37.3.2. TTL — 60 seconds; changes propagate within one minute
##### 37.3.3. Cache miss — load from DB, populate cache, return value
##### 37.3.4. Cache invalidation on override update

#### 37.4. Telemetry
##### 37.4.1. Flag evaluation events emitted to OpenTelemetry
##### 37.4.2. Dead flag detection — weekly job reports flags with no variation in evaluations
##### 37.4.3. Flag impact analysis — listing entities and page builders that read a given flag

---

### 38. Notifications

#### 38.1. The Notification Entity
##### 38.1.1. Notification fields — id, tenant_id, user_id, title, body, type, channel, status, read_at, metadata (JSONB)
##### 38.1.2. Notification types — info, warning, error, success, action_required
##### 38.1.3. Delivery channels — in-app (SSE), email, SMS
##### 38.1.4. Delivery status — PENDING, SENT, DELIVERED, FAILED, READ

#### 38.2. In-App Notifications via SSE
##### 38.2.1. `GET /api/v1/notifications/stream` — persistent SSE connection per user
##### 38.2.2. Redis pub/sub as the fan-out mechanism — worker publishes, SSE handler subscribes
##### 38.2.3. amis SSE component — subscribes to the stream, renders notification badges
##### 38.2.4. Mark as read — `POST /api/v1/notifications/{id}/read`

#### 38.3. Email Notifications
##### 38.3.1. Email template system — Go `html/template`, one template per notification type
##### 38.3.2. Email delivery via SMTP or transactional email service (configurable per tenant)
##### 38.3.3. Delivery tracking — webhook callback updates notification status

#### 38.4. SMS Notifications
##### 38.4.1. SMS gateway integration — Africa's Talking for Kenya market
##### 38.4.2. SMS templates — short, action-focused
##### 38.4.3. Opt-out handling

#### 38.5. Notification Delivery Workflow
##### 38.5.1. `SendNotificationWorkflow` — Temporal workflow for reliable delivery with retry
##### 38.5.2. Channel selection — determined by notification type and user preferences
##### 38.5.3. Failure handling — retry with exponential backoff, escalation on repeated failure

---

### 39. Settings — TenantConfig

#### 39.1. The TenantConfig Entity
##### 39.1.1. TenantConfig fields — tenant_id, category, key, value (text), updated_at, updated_by
##### 39.1.2. Configuration categories — locale, modules, integrations, limits, branding, notifications
##### 39.1.3. System defaults — global rows with tenant_id = NULL; tenant rows override system defaults

#### 39.2. Type-Safe Access
##### 39.2.1. Typed getter functions — `config.GetString(ctx, "locale.timezone")`, `config.GetBool(ctx, "modules.forecourt.enabled")`
##### 39.2.2. No key-string lookups at call sites — category constants defined in framework
##### 39.2.3. Getter caching — config loaded at tenant boot, cached for request lifetime

#### 39.3. Module Enablement
##### 39.3.1. Enabling and disabling modules per tenant via config keys
##### 39.3.2. Module feature subset flags — enabling a module but restricting features
##### 39.3.3. Module dependency — disabling Finance blocks Inventory GL posting

#### 39.4. Locale and Regional Settings
##### 39.4.1. Timezone — East Africa Time (EAT, UTC+3) as default
##### 39.4.2. Date format — DD/MM/YYYY for Kenya market
##### 39.4.3. Currency — KES as base, other currencies as secondary
##### 39.4.4. Language — English and Swahili supported

#### 39.5. Kenya-Specific Configuration
##### 39.5.1. KRA eTIMS integration settings — taxpayer PIN, device serial, environment
##### 39.5.2. NEMA compliance thresholds — tank variance limits
##### 39.5.3. NHIF and NSSF rates — updated from configuration, not hardcoded

---

### 40. Metadata and Reporting

#### 40.1. CustomFieldDef
##### 40.1.1. CustomFieldDef entity fields — key, label, entity_name, field_type, options, validation_rules, ui_metadata, tenant_id
##### 40.1.2. Loading at tenant boot — all CustomFieldDef rows for a tenant are loaded into the EntityRegistry
##### 40.1.3. Runtime extension of system entities — adding fields without code deployment
##### 40.1.4. Lifecycle — add, rename (label only), type change (migration needed), deprecate, remove

#### 40.2. ReportDefinition
##### 40.2.1. ReportDefinition entity fields — name, label, query (SQL or filter DSL), parameters, columns, chart_config, tenant_id
##### 40.2.2. SQL reports — parameterised queries with `{{.param}}` template syntax, RLS applied automatically via `set_tenant_context()`
##### 40.2.3. Filter DSL reports — composed from EntityRepository filters, no raw SQL
##### 40.2.4. Report parameters — date range, entity selectors, grouping
##### 40.2.5. `GET /api/v1/reports/{name}?{params}` — execute report, return results in standard envelope
##### 40.2.6. Export — CSV and PDF export from report results
##### 40.2.7. Report scheduling — run on a Temporal schedule, email results

---

### 41b. Implementing Platform Modules — Worked Reference

> Full source: `docs/framework/part-06-modules/implementing-platform-modules.md`

#### 41b.1. Module Layout Convention
##### 41b.1.1. Directory structure shared by platform and business modules
##### 41b.1.2. The `init()` + `definition.Register()` wiring pattern
##### 41b.1.3. Why the framework cannot distinguish platform from business EntityDefinitions

#### 41b.2. Tenant Module
##### 41b.2.1. TenantDef — `OrgScope: ScopeLevelGlobal`, no RLS, system-only write policies
##### 41b.2.2. OrgNodeDef — materialised path hook, tree-scoped read policy
##### 41b.2.3. Status transition hook — validating PENDING→ACTIVE, blocking PENDING→SUSPENDED
##### 41b.2.4. Migration — no RLS on `tenants` table; RLS on `org_nodes`

#### 41b.3. IAM Module
##### 41b.3.1. UserDef — Sensitive fields (password_hash, mfa_secret), `allowSelf` read policy
##### 41b.3.2. RoleDef — protecting system roles via BeforeHook
##### 41b.3.3. UserRoleDef — `set_assigned_by` hook, Casbin reload AfterHook
##### 41b.3.4. SessionDef — no user CRUD; IAM service layer drives login/logout
##### 41b.3.5. Migration — partial unique index for active session token lookup

#### 41b.4. Feature Flags Module
##### 41b.4.1. FeatureFlagDef — global catalogue, platform_admin write policy
##### 41b.4.2. FeatureFlagOverrideDef — tenant-scoped, cache invalidation hook
##### 41b.4.3. EvalService — Redis fast path, three-tier cascade evaluation

#### 41b.5. Settings Module
##### 41b.5.1. ConfigKeyDef — global key catalogue, read-only for tenant callers
##### 41b.5.2. ConfigValueDef — tenant-scoped, scope=system|tenant|entity, value type validation hook
##### 41b.5.3. Resolution service — entity beats tenant beats system default

#### 41b.6. Audit Log Module
##### 41b.6.1. AuditLogDef — append-only; `OpCreate` for system only; `OpWrite/Delete` denied
##### 41b.6.2. DB trigger `audit_log_trigger_fn` — captures all SQL writes
##### 41b.6.3. Monthly partitioning — 7-year Kenya compliance retention

#### 41b.7. Metadata / Custom Fields Module
##### 41b.7.1. CustomFieldDefDef — soft-delete, validateFieldName hook, SDUI cache invalidation
##### 41b.7.2. How business modules receive custom fields — `custom_fields jsonb` column
##### 41b.7.3. SDUI form field merging — framework fields + tenant custom fields at render time

#### 41b.8. Plugin / Module Registry
##### 41b.8.1. ModuleDef — global catalogue, populated by module `init()` calls
##### 41b.8.2. TenantModuleDef — which modules a tenant has activated; navigation reload hook
##### 41b.8.3. Business module self-registration pattern — `registry.RegisterModule()`

#### 41b.9. Startup Wire-Up
##### 41b.9.1. Import order — platform modules before business modules
##### 41b.9.2. `bootstrap.Mount` reads `definition.All()` — single pass for all modules

#### 41b.10. Platform vs Business Module Comparison Table

---

## Part VII — Multi-Tenancy and Configuration

### 41. Tenant Lifecycle

#### 41.1. Tenant Provisioning Workflow
##### 41.1.1. Registration trigger — new tenant signup or admin creation
##### 41.1.2. No DDL during provisioning — all tables already exist; provisioning inserts seed data rows tagged with `tenant_id`
##### 41.1.3. Seed data — system roles, default chart of accounts, leave types, PAYE bands, feature flag overrides, tenant config
##### 41.1.4. Tenant activation — setting status to ACTIVE after provisioning workflow completes
##### 41.1.5. Idempotency — provisioning workflow ID is `{slug}.provision`, duplicate runs are rejected

#### 41.2. Tenant Suspension and Reactivation
##### 41.2.1. Suspension trigger — billing lapse, terms violation, manual admin action
##### 41.2.2. What suspension does — API returns 402; `set_tenant_context()` rejects DB access; background jobs paused
##### 41.2.3. Reactivation flow — set status ACTIVE, resume jobs, notify admin users

#### 41.3. Tenant Offboarding
##### 41.3.1. Data export — full tenant data extract (SQL/JSON filtered by `WHERE tenant_id = ?`)
##### 41.3.2. Archive upload — encrypted to S3-compatible object storage
##### 41.3.3. Mark ARCHIVED — terminal; API returns 410 Gone
##### 41.3.4. Retention period before physical row deletion — configurable, Kenya minimum 7 years for financial records

#### 41.4. Tenant Cloning
##### 41.4.1. Use cases — staging environments, franchise template, disaster recovery test
##### 41.4.2. Clone operation — `INSERT ... SELECT` rewriting `tenant_id`; no schema manipulation
##### 41.4.3. Clone limitations — does not copy transactional data by default

---

### 42. Tenant Configuration System

#### 42.1. `TenantConfig` Entity
##### 42.1.1. Configuration categories — modules, locale, integrations, limits
##### 42.1.2. Configuration inheritance — system defaults overridden by tenant values
##### 42.1.3. Type-safe config access — typed getter functions, no key-string lookups

#### 42.2. Module Enablement
##### 42.2.1. Enabling and disabling built-in modules per tenant
##### 42.2.2. Module feature subset flags — enabling a module but restricting features
##### 42.2.3. Module dependency — enabling Finance implicitly enables nothing, but disabling it blocks Inventory GL posting

#### 42.3. Locale and Regional Settings
##### 42.3.1. Timezone — East Africa Time (EAT, UTC+3) as default
##### 42.3.2. Date format — DD/MM/YYYY for Kenya market
##### 42.3.3. Currency — KES as base, other currencies as secondary
##### 42.3.4. Language — English and Swahili supported

#### 42.4. Kenya-Specific Configuration
##### 42.4.1. KRA eTIMS integration settings — taxpayer PIN, device serial, environment
##### 42.4.2. NEMA compliance thresholds — tank variance limits
##### 42.4.3. NHIF and NSSF rates — updated from configuration, not hardcoded

---

### 43. Feature Flag System — Deep Dive

#### 43.1. Flag Storage and Loading
##### 43.1.1. Flag definition entity — system-level, not per-tenant
##### 43.1.2. Redis key schema for flag values
##### 43.1.3. Flag loading at tenant context resolution
##### 43.1.4. TTL and cache refresh

#### 43.2. Flag Evaluation Engine
##### 43.2.1. Evaluation order — system default → tenant override → user override
##### 43.2.2. Percentage rollout — deterministic hash of `{tenant_id}:{flag_name}`
##### 43.2.3. Flag dependency graph — circular dependency detection at load time

#### 43.3. Flag Telemetry
##### 43.3.1. Flag evaluation events — which flag, which tenant, which outcome
##### 43.3.2. Dead flag detection — automated report of stale flags
##### 43.3.3. Flag impact analysis — entities and pages affected by a flag

---

### 44. Redis Usage in Awo

#### 44.1. Redis as Session Store
##### 44.1.1. Session key schema — `session:{tenant}:{id}`
##### 44.1.2. Session data structure
##### 44.1.3. TTL management — sliding expiry implementation

#### 44.2. Redis as Feature Flag Cache
##### 44.2.1. Flag key schema
##### 44.2.2. TTL strategy — short TTL for fast rollout, long TTL for stable flags

#### 44.3. Redis as Page Definition Cache
##### 44.3.1. Cache key — `page:{tenant}:{page_name}:{permissions_hash}:{flags_hash}`
##### 44.3.2. Invalidation — on role change, on feature flag change, on EntityDefinition change

#### 44.4. Redis Pub/Sub for Real-Time Notifications
##### 44.4.1. Channel naming conventions
##### 44.4.2. Server-sent events (SSE) endpoint consuming Redis pub/sub
##### 44.4.3. amis receiving SSE notifications

#### 44.5. Redis Failure Handling
##### 44.5.1. Session store failure — graceful degradation vs hard failure
##### 44.5.2. Feature flag cache failure — fallback to system defaults
##### 44.5.3. Page cache failure — pass-through to page builder

#### 44.6. Redis Connection Management
##### 44.6.1. rueidis client configuration
##### 44.6.2. Connection pool sizing
##### 44.6.3. Redis key naming conventions — full reference

---

## Part VIII — Deployment and Operations

### 45. Environment Architecture

#### 45.1. Local Development
##### 45.1.1. Docker Compose stack — PostgreSQL, Redis, Temporal
##### 45.1.2. `awo serve` with hot reload
##### 45.1.3. Seeding a dev tenant — `awo tenant create` + `awo seed`

#### 45.2. Staging
##### 45.2.1. Production-parity requirements — same infra, smaller size
##### 45.2.2. Data anonymisation for staging — no real tenant data
##### 45.2.3. Staging-specific feature flags

#### 45.3. Production
##### 45.3.1. Single VPS deployment — Nairobi region
##### 45.3.2. Stateless Fiber processes — horizontal scaling model
##### 45.3.3. Shared PostgreSQL and Redis — connection pooling requirements
##### 45.3.4. Load balancer — Caddy configuration

---

### 46. Docker and Containerisation

#### 46.1. Multi-Stage Dockerfile
##### 46.1.1. Build stage — Go binary compilation
##### 46.1.2. Runtime stage — minimal base image
##### 46.1.3. Non-root user — security requirement
##### 46.1.4. Binary size optimisation — `-ldflags "-s -w"`

#### 46.2. Temporal Worker Container
##### 46.2.1. Shared binary with the API process vs separate binary
##### 46.2.2. Sidecar deployment vs separate service
##### 46.2.3. Worker-specific environment variables

#### 46.3. Docker Compose for Local Stack
##### 46.3.1. Services — `api`, `worker`, `postgres`, `redis`, `temporal`
##### 46.3.2. Volume mounts — code, migration files
##### 46.3.3. Environment variable injection from `.env`

#### 46.4. Image Tagging and Secrets
##### 46.4.1. Tagging convention — `{git_sha}` for production, `latest` never in production
##### 46.4.2. Secrets in containers — environment variables injected at runtime
##### 46.4.3. Never baking secrets into images

---

### 47. CI/CD Pipeline

#### 47.1. GitHub Actions Workflow Structure
##### 47.1.1. Trigger events — push to main, pull request, release tag
##### 47.1.2. Job dependency graph

#### 47.2. Test Stage
##### 47.2.1. Unit tests — `go test ./...`
##### 47.2.2. Integration tests — PostgreSQL and Redis service containers
##### 47.2.3. Temporal workflow tests — Temporal test environment
##### 47.2.4. Contract tests — API contract verification with `hurl`
##### 47.2.5. Migration tests — apply all migrations to a temporary database, verify round-trip

#### 47.3. Build and Push Stage
##### 47.3.1. Docker buildx for multi-platform images
##### 47.3.2. Push to container registry

#### 47.4. Deploy Stage
##### 47.4.1. Rolling update strategy — one instance at a time
##### 47.4.2. Migration run before traffic switches — `awo migrate up`
##### 47.4.3. Smoke tests post-deploy
##### 47.4.4. Automatic rollback on failed smoke tests

---

### 48. PostgreSQL Operations

#### 48.1. RLS Architecture
##### 48.1.1. All tenants share a single database and schema — no `CREATE SCHEMA` per tenant
##### 48.1.2. Every tenant-scoped table has `tenant_id uuid NOT NULL REFERENCES tenants(id)`
##### 48.1.3. RLS enabled on every tenant-scoped table — `ALTER TABLE ... ENABLE ROW LEVEL SECURITY`
##### 48.1.4. Row security policy — `USING (tenant_id = current_setting('app.current_tenant_id')::uuid)`
##### 48.1.5. `set_tenant_context(uuid)` — sets `app.current_tenant_id`, validates tenant status
##### 48.1.6. Application role is `ROW SECURITY` enforced — no bypass privilege

#### 48.2. Connection Pooling
##### 48.2.1. PgBouncer configuration — transaction mode
##### 48.2.2. Per-tenant pool sizing
##### 48.2.3. Max connections ceiling — PostgreSQL `max_connections`
##### 48.2.4. RLS context reset between pooled connections — `set_tenant_context()` called on acquire

#### 48.3. Backup Strategy
##### 48.3.1. `pg_dump` — daily full backup of the shared database
##### 48.3.2. WAL archiving — continuous, enables PITR
##### 48.3.3. Backup retention policy
##### 48.3.4. Backup encryption and offsite storage

#### 48.4. Point-in-Time Recovery
##### 48.4.1. PITR procedure
##### 48.4.2. Per-tenant data recovery — restore to a staging database, export filtered by `tenant_id`, import

#### 48.5. Performance Tuning
##### 48.5.1. Autovacuum tuning for write-heavy ERP tables
##### 48.5.2. Bloat monitoring — pg_stat_user_tables
##### 48.5.3. Read replica setup for reporting queries

---

### 49. Observability Stack

#### 49.1. Structured Logging
##### 49.1.1. `slog` setup — JSON output in production, text in development
##### 49.1.2. Mandatory fields on every log line — tenant_id, request_id, user_id, duration
##### 49.1.3. Log levels and when to use each
##### 49.1.4. Log shipping to Loki

#### 49.2. Distributed Tracing
##### 49.2.1. OpenTelemetry SDK setup
##### 49.2.2. Trace propagation — Fiber → ent/JSONB engine → Temporal activities
##### 49.2.3. Custom span attributes — tenant_id, entity_name, workflow_id
##### 49.2.4. Trace sampling strategy

#### 49.3. Metrics
##### 49.3.1. Prometheus exposition from Fiber — `prometheus.New()` middleware
##### 49.3.2. Custom business metrics — active tenants, workflows per minute, GL postings per hour
##### 49.3.3. Grafana dashboard — request rate, error rate, p99 latency, workflow lag

#### 49.4. Alerting Rules
##### 49.4.1. SLO-based alerts — error budget burn rate
##### 49.4.2. Workflow failure rate alert
##### 49.4.3. DB connection pool saturation alert
##### 49.4.4. Redis memory pressure alert

#### 49.5. Audit Log
##### 49.5.1. Audit log entity — who, what, when, from where, previous value, new value
##### 49.5.2. Write path — DB trigger + `after_save` hook enrichment
##### 49.5.3. Retention policy
##### 49.5.4. Audit log search endpoint for compliance

---

### 50. Security Hardening

#### 50.1. TLS Configuration
##### 50.1.1. Certificate management with Caddy — automatic ACME
##### 50.1.2. TLS version — 1.2 minimum, 1.3 preferred
##### 50.1.3. HSTS configuration

#### 50.2. PostgreSQL Access Controls
##### 50.2.1. Least-privilege DB roles — application user, migration user, backup user
##### 50.2.2. Row-Level Security as the primary tenant isolation mechanism — all tenant data access goes through RLS
##### 50.2.3. Application DB role has no `BYPASSRLS` privilege
##### 50.2.4. pg_hba.conf — restrict connections to app server IPs

#### 50.3. Temporal mTLS
##### 50.3.1. Certificate generation
##### 50.3.2. Client certificate configuration in the Temporal Go SDK

#### 50.4. Dependency Vulnerability Scanning
##### 50.4.1. `govulncheck` — runs in CI on every push
##### 50.4.2. `trivy` — container image scanning

#### 50.5. Penetration Testing Checklist
##### 50.5.1. Tenant isolation verification — can tenant A read tenant B's data (RLS bypass attempt)
##### 50.5.2. Privilege escalation — can a non-admin obtain admin access
##### 50.5.3. Injection — SQL, JSONB path injection
##### 50.5.4. Mass assignment — can a client set fields it should not

---

### 51. Incident Response Playbooks

#### 51.1. Database Connection Exhaustion
##### 51.1.1. Symptoms and detection
##### 51.1.2. Immediate mitigation — PgBouncer pool reset
##### 51.1.3. Root cause — long-running queries, leaked connections
##### 51.1.4. Remediation and post-incident

#### 51.2. Redis Eviction Under Memory Pressure
##### 51.2.1. Symptoms — session lookup failures, flag cache misses
##### 51.2.2. Immediate mitigation — increase memory or evict less critical keys
##### 51.2.3. Key sizing audit

#### 51.3. Temporal Worker Stopped Processing
##### 51.3.1. Detection — workflow pending age alert fires
##### 51.3.2. Diagnosis — worker logs, Temporal Web UI
##### 51.3.3. Restart procedure
##### 51.3.4. Workflow backlog clearance

#### 51.4. Stuck Workflow
##### 51.4.1. Definition — workflow running longer than expected, not completing
##### 51.4.2. Diagnosis via Temporal Web UI — event history inspection
##### 51.4.3. Safe termination — `temporal workflow terminate` with reason
##### 51.4.4. Compensation — manually triggering compensating actions if needed

#### 51.5. Failed Migration
##### 51.5.1. Detection — `awo migrate up` exits non-zero
##### 51.5.2. Single shared schema — a failed migration affects all tenants simultaneously
##### 51.5.3. Rollback procedure — `awo migrate down 1`, fix the migration, re-run
##### 51.5.4. Emergency `migrate force` — marking a partially applied migration and fixing manually

#### 51.6. Tenant Data Isolation Breach
##### 51.6.1. Containment — suspend affected tenants immediately
##### 51.6.2. Scope assessment — which tenants, which entities, which time window
##### 51.6.3. Notification obligations — Kenya Data Protection Act 2019, tenant contracts
##### 51.6.4. Remediation — patch RLS policy, re-test isolation, verify with pen test

---

### 52. CLI Reference — the `awo` binary

#### 52.1. Overview
##### 52.1.1. `awo` is a standalone binary distributed with the framework — no separate tool download
##### 52.1.2. Inspired by frappe-bench but simpler — one binary, project-scoped commands
##### 52.1.3. Configuration — reads `AWO_*` environment variables; `--config` flag overrides

#### 52.2. Project Scaffolding
##### 52.2.1. `awo new <project-name>` — scaffold a new Awo project with directory layout, go.mod, docker-compose, and starter config
##### 52.2.2. Scaffolded structure — `cmd/`, `internal/`, `migrations/`, `fixtures/`, `templates/`, `docker-compose.yml`

#### 52.3. `awo serve` — Run the Application
##### 52.3.1. `awo serve` — start API server and Temporal worker in one process (development default)
##### 52.3.2. `awo serve --port=8080` — set the HTTP port
##### 52.3.3. `awo serve --worker` — start API + worker (explicit)
##### 52.3.4. `awo worker` — start Temporal worker only (production sidecar pattern)

#### 52.4. `awo migrate` — Database Migrations
##### 52.4.1. `awo migrate up` — apply all pending migrations
##### 52.4.2. `awo migrate down [N]` — roll back N migrations (default 1)
##### 52.4.3. `awo migrate version` — print current applied migration version
##### 52.4.4. `awo migrate create <name>` — create a new timestamped `.up.sql` / `.down.sql` file pair
##### 52.4.5. `awo migrate force <V>` — mark version V as applied without running SQL (emergency use only)

#### 52.5. `awo tenant` — Tenant Management
##### 52.5.1. `awo tenant create --name="Name" --slug=slug --email=admin@example.com` — provision a new tenant via the provisioning workflow
##### 52.5.2. `awo tenant list` — list all tenants with status and plan
##### 52.5.3. `awo tenant suspend --slug=slug` — suspend a tenant
##### 52.5.4. `awo tenant activate --slug=slug` — reactivate a suspended tenant
##### 52.5.5. `awo tenant export --slug=slug --out=dir` — export all tenant rows to JSON files

#### 52.6. `awo entity` — Entity Inspection
##### 52.6.1. `awo entity list` — list all registered EntityDefinitions (name, type, owner package)
##### 52.6.2. `awo entity inspect --name=Name` — print full EntityDefinition with fields, edges, hooks, permissions

#### 52.7. `awo seed` — Data Seeding
##### 52.7.1. `awo seed --module=<module>` — seed a module's default reference data
##### 52.7.2. `awo seed --module=<module> --tenant=<slug>` — seed into a specific tenant

#### 52.8. `awo console` — Interactive REPL
##### 52.8.1. What `awo console` provides — Go REPL with framework context loaded (DB, Redis, tenant context)
##### 52.8.2. Use cases — debugging, ad-hoc queries, scripted data fixes
##### 52.8.3. Analogous to `bench console` in Frappe

#### 52.9. `awo jobs` — Workflow Management
##### 52.9.1. `awo jobs list [--tenant=slug]` — list running Temporal workflows for a tenant
##### 52.9.2. `awo jobs cancel <workflow-id>` — cancel a running workflow
##### 52.9.3. `awo jobs retry <workflow-id>` — retry a failed workflow

#### 52.10. `awo config` — Tenant Configuration
##### 52.10.1. `awo config get <key> --tenant=<slug>` — print current value of a config key
##### 52.10.2. `awo config set <key> <value> --tenant=<slug>` — set a config value

#### 52.11. `awo version`
##### 52.11.1. Prints framework version, application version (from build-time injection), and Go runtime version

---

## Part IX — Internals and Extending the Framework

### 53. Framework Internals

#### 53.1. Server Startup Sequence
##### 53.1.1. Step 1 — config loaded and validated
##### 53.1.2. Step 2 — database connection pool established
##### 53.1.3. Step 3 — Redis connection established
##### 53.1.4. Step 4 — EntityRegistry populated — system entities registered
##### 53.1.5. Step 5 — Temporal client connected, worker registered
##### 53.1.6. Step 6 — Fiber server starts accepting connections
##### 53.1.7. What happens when any step fails — startup aborts vs degraded mode

#### 53.2. The EntityResolver — Dispatch Logic
##### 53.2.1. Input — entity name, tenant context
##### 53.2.2. Registry lookup — is this a system entity or a custom entity for this tenant
##### 53.2.3. System entity path — resolves to the `EntityRepository` implementation for that entity
##### 53.2.4. Custom entity path — resolves to the JSONB engine with the `CustomFieldDef` metadata loaded
##### 53.2.5. What the caller receives — an `EntityRepository` interface, regardless of path
##### 53.2.6. EntityResolver caching — resolver results cached per tenant boot

#### 53.3. The Permission Resolution Pipeline
##### 53.3.1. Input — user, tenant, entity name, operation type
##### 53.3.2. Step 1 — load role assignments for user
##### 53.3.3. Step 2 — load permission matrix for each role × entity × operation
##### 53.3.4. Step 3 — merge permissions across roles
##### 53.3.5. Step 4 — apply tenant-level overrides
##### 53.3.6. Step 5 — apply user-level overrides (if enabled for tenant)
##### 53.3.7. Output — `Permissions` struct passed to hooks, page builders, privacy policies

#### 53.4. The Page Builder Pipeline — Internals
##### 53.4.1. Cache key computation
##### 53.4.2. Cache hit path — deserialise and return
##### 53.4.3. Cache miss path — invoke builder function
##### 53.4.4. Builder function receives — entity metadata, permissions, feature flags, tenant config
##### 53.4.5. Output — amis JSON struct, serialised and cached

---

### 54. Plugin System

#### 54.1. Plugin Interface
##### 54.1.1. `Plugin` interface definition
##### 54.1.2. `Name()` and `Version()` — identification
##### 54.1.3. `Register(framework.App)` — the single registration entry point

#### 54.2. What a Plugin Can Register
##### 54.2.1. New EntityDefinitions — system or custom
##### 54.2.2. New workflow types
##### 54.2.3. New amis page builders
##### 54.2.4. New amis custom renderers
##### 54.2.5. New CLI commands

#### 54.3. Plugin Versioning and Compatibility
##### 54.3.1. Plugin manifest — minimum Awo version, declared dependencies
##### 54.3.2. Compatibility checking at startup
##### 54.3.3. Plugin isolation — panics in a plugin must not crash the framework

---

### 55. Testing Guide

#### 55.1. Unit Testing EntityDefinition Hooks
##### 55.1.1. The interceptor pattern — injecting mock dependencies
##### 55.1.2. Testing `before_save` with a mock EntityRepository
##### 55.1.3. Asserting hook return values and side effects

#### 55.2. Integration Testing Fiber Handlers
##### 55.2.1. Using `net/http/httptest` with the Fiber app
##### 55.2.2. Test database setup — separate test database with full migrations applied
##### 55.2.3. Seeding test fixtures
##### 55.2.4. Asserting response shape against expected amis envelope

#### 55.3. Testing Temporal Workflows
##### 55.3.1. The Temporal test environment — `testsuite.WorkflowTestSuite`
##### 55.3.2. Time skipping — `env.Sleep()` advances Temporal's clock
##### 55.3.3. Mocking activities — isolating workflow logic from external calls
##### 55.3.4. Testing signals — injecting signals in the test environment
##### 55.3.5. Testing compensation — triggering activity failures and asserting compensations ran

#### 55.4. Testing Privacy Policies
##### 55.4.1. Unit testing with mock context
##### 55.4.2. Integration testing — asserting filtered query results end-to-end

#### 55.5. Contract Testing
##### 55.5.1. What contract testing tests — API response shape, not behaviour
##### 55.5.2. `hurl` test files — structure and conventions
##### 55.5.3. Running contract tests in CI
##### 55.5.4. Contract tests as backward-compatibility guards

#### 55.6. Multi-Tenant Test Fixtures
##### 55.6.1. Tenant fixture helper — creates an isolated test tenant
##### 55.6.2. Cross-tenant isolation tests — asserting tenant A cannot read tenant B's data (RLS verification)

#### 55.7. Load Testing
##### 55.7.1. `k6` scripts for ERP workloads — concurrent form submits, list pagination
##### 55.7.2. Baseline benchmarks — what acceptable p99 latency looks like
##### 55.7.3. Identifying bottlenecks — DB queries, page builder, serialisation

#### 55.8. Test Coverage
##### 55.8.1. Coverage targets — per package
##### 55.8.2. Enforcement in CI — failing build below threshold
##### 55.8.3. Coverage exemptions — generated code, main packages

---

## Part X — Contributing to Awo

### 56. Repository Structure and Conventions

#### 56.1. Repository Layout
##### 56.1.1. `cmd/` — entry points
##### 56.1.2. `internal/` — framework core — not importable by external code
##### 56.1.3. `pkg/` — importable packages — the public framework API
##### 56.1.4. `testkit/` — test helpers for framework consumers
##### 56.1.5. `migrations/` — versioned SQL migration files (golang-migrate format)
##### 56.1.6. `templates/` — print templates for platform entities

#### 56.2. Branch and PR Conventions
##### 56.2.1. Branch naming — `feat/`, `fix/`, `docs/`, `chore/`
##### 56.2.2. PR size guidelines — under 400 lines diff preferred
##### 56.2.3. Required reviewers — framework core changes require two approvals

#### 56.3. Commit Message Format
##### 56.3.1. Conventional commits — `feat:`, `fix:`, `docs:`, `refactor:`, `test:`
##### 56.3.2. Breaking changes — `!` suffix and `BREAKING CHANGE:` footer
##### 56.3.3. Commit scope — `(entity)`, `(api)`, `(workflow)`, `(ui)`, `(ops)`

#### 56.4. Adding a New Platform EntityDefinition
##### 56.4.1. Schema file
##### 56.4.2. Repository interface implementation file (system entity)
##### 56.4.3. Hook files
##### 56.4.4. Privacy policy file
##### 56.4.5. Page builder file
##### 56.4.6. Migration file pair (.up.sql / .down.sql)
##### 56.4.7. Test files — unit, integration, contract
##### 56.4.8. Documentation requirement — the entity must have a corresponding section in Part VI

#### 56.5. Release Process
##### 56.5.1. Semantic versioning — `v{major}.{minor}.{patch}`
##### 56.5.2. Changelog — `CHANGELOG.md` maintained per release
##### 56.5.3. Release branch cut and tag

---

### 57. Roadmap and Known Limitations

#### 57.1. Current Limitations
##### 57.1.1. Runtime EntityDefinition creation — currently requires a code deploy and migration
##### 57.1.2. GraphQL API — not yet available, under evaluation
##### 57.1.3. Offline-first mobile — research phase

#### 57.2. Planned
##### 57.2.1. Connect-Go / mobile SDK — protobuf contracts for Flutter
##### 57.2.2. KRA eTIMS e-invoicing integration
##### 57.2.3. Plugin marketplace

---

## Appendices

### Appendix A — EntityDefinition Field Types Quick Reference
#### A.1. Scalar field types — type name, Go type, PostgreSQL column type, nullable behaviour
#### A.2. Structured field types
#### A.3. Relational field types
#### A.4. File field types
#### A.5. Custom field type subset — which types are available for custom fields

### Appendix B — `EntityRepository` Interface — Full Signature Reference
#### B.1. Read methods
#### B.2. Write methods
#### B.3. Transaction methods
#### B.4. Filter DSL types

### Appendix C — golang-migrate Quick Reference
#### C.1. CLI commands
#### C.2. Migration file naming convention
#### C.3. RLS policy boilerplate
#### C.4. Common DDL patterns (views, triggers, functions)
