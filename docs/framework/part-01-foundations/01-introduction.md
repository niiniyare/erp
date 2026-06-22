# Introduction

## 1.1 What Awo Framework Is

### 1.1.1 A Go-native platform for building multi-tenant ERP systems

Awo is a Go framework for building multi-tenant enterprise resource planning systems. It provides structural scaffolding, integration points, and runtime dispatch that ERP systems require — not business logic of any particular ERP domain. The framework defines how data is described, stored, validated, presented, authorized, and processed; your modules and applications supply what that data means and how it flows through your business.

The "multi-tenant" qualifier is load-bearing in a precise sense: every abstraction in Awo — the `EntityDefinition`, the `EntityRepository`, the permission system, the workflow engine integration, the UI schema server — was designed from the ground up assuming a single running process serves many tenants simultaneously, each with their own data isolation, schema extensions, configuration, and potentially their own regulatory requirements. Tenancy is not a conditional feature or a filter bolted on after the fact. It is a structural property of the framework that touches every layer, from the database row to the HTTP response.

"ERP" here refers to software characterized by: a unified data model spanning multiple business functions (finance, inventory, HR, procurement, sales); strong transactional integrity requirements; complex approval and workflow processes; regulatory compliance obligations; and the need to support tenant-specific customization without forking core code. Awo addresses each of these requirements explicitly.

Go was chosen for architectural reasons. ERP systems are latency-sensitive at the request level but throughput-sensitive at the batch level. Go's goroutine model handles both without thread pool management. Go's static type system catches ERP-specific bugs at compile time — mismatched currency types, incorrect field references, missing permission checks — that dynamic languages only catch at runtime in production. Go's compilation model produces a single static binary that deploys cleanly in containerized environments, which matters when operating in regulated cloud environments.

### 1.1.2 What "framework" means here vs "library" vs "application"

This distinction is not semantic pedantry — it determines what you own and what the framework owns, and therefore where bugs live and who fixes them.

A library is code your application calls. You control the call sites, the data flow, and the lifecycle. A library provides utility; it does not impose structure. A framework inverts this: the framework calls your code at defined extension points, controls the main execution loop, and imposes structure on how your code is organized. An application is a complete deployable system with specific business logic baked in — not designed to be extended by users, only configured.

Awo is a framework in the strict sense. The main function, HTTP server initialization, worker registration, tenant boot sequence, request lifecycle — these are owned by Awo. You provide `EntityDefinition` declarations, hook implementations, permission policies, and workflow activity implementations. Awo calls your code at the right moment in the right context. This means you cannot arbitrarily restructure the request pipeline or bypass the permission system without modifying the framework itself. That constraint is the point: the framework's value comes precisely from the guarantees it makes about how every entity, every request, and every workflow interaction behaves.

When you start an Awo project, you are not writing a Go HTTP server that uses some ERP-related libraries. You are implementing extension points that Awo exposes. Your codebase contains `EntityDefinition` declarations, hook functions, and Temporal activity implementations. The Awo framework binary wraps them. This is a different mental model from how most Go developers approach web service development.

> **Note:** Awo exposes certain components as standalone packages — `entitydef` can be imported without the full framework runtime for tooling, code generation, and migration scripts. This does not change the framework relationship for running applications.

### 1.1.3 What Awo is not — anti-patterns explicitly out of scope

Knowing what a system will not do is as important as knowing what it will do, because the boundaries shape how you design extensions and where you look when something does not behave as expected.

Awo is not a general-purpose web framework. It does not compete with Fiber, Echo, or Gin as a basis for arbitrary HTTP services. It uses Fiber internally for its HTTP layer, but Awo's routing is generated from `EntityDefinition` declarations and module registrations, not hand-written route tables. If you need a bespoke API endpoint that does not correspond to an entity operation or a framework-defined action, you register it as a module extension — not by patching the framework router.

Awo is not a database ORM in the general sense. The `EntityRepository` interface (see [§8 The Persistence Interface]) is a domain-specific persistence abstraction for entities — structured records with identity, lifecycle, and relationships. It is not a query builder for arbitrary SQL. Complex reporting queries, analytics aggregations, and data warehouse exports go through the `Aggregate` method or through dedicated read-only reporting services that operate outside the entity lifecycle entirely.

Awo is not an event-sourcing framework. It uses an outbox pattern for reliable event delivery and Temporal for durable workflow execution, but the primary source of truth is the relational database state, not an event log. Rebuilding state by replaying events is not a design goal and is not supported.

Awo is not a microservices framework. A single Awo process handles HTTP requests, runs Temporal activities, and manages tenant sessions. Horizontal scaling is achieved by running multiple instances of the same binary behind a load balancer, not by decomposing into independent services per domain.

> **Warning:** Bypassing the `EntityDefinition` system by writing direct database queries in request handlers produces a system no longer protected by Awo's permission evaluation, audit logging, or multi-tenancy guarantees. The framework cannot protect data accessed outside its abstractions. This applies with particular force to the RLS policies described in [§1.2.6]: bypassing the `EntityRepository` with raw SQL does not mean RLS is not active — it means you lose all the framework's safety around *how* RLS is applied, with no protection if a session variable is misconfigured.

### 1.1.4 Relationship to ERPNext and Frappe — conceptual debts and deliberate departures

Awo was influenced by Frappe Framework and ERPNext at the conceptual level. The `EntityDefinition` is a direct analogue of Frappe's DocType: a descriptor that drives the entire system's understanding of what a record is, how it is stored, how it is presented, and what operations are permitted on it. The "everything is a DocType" philosophy of Frappe is a genuine contribution to ERP framework design, and Awo inherits it under a different name with a structurally different implementation.

The departures are deliberate and consequential. Frappe is Python-based and uses a global process-level DocType registry loaded from MariaDB at startup. DocType definitions are data, not code, which means type errors in field references are runtime errors, not compile errors. Frappe's permission system uses a role-based matrix that is powerful but difficult to audit and extend programmatically. Frappe's background jobs are Python threads with limited durability guarantees. Frappe's forms are Jinja-rendered HTML, tying the UI to server-side rendering.

Awo replaces each of these with a Go-native equivalent that trades some dynamic flexibility for compile-time correctness, operational clarity, and better fit with the Go deployment model. System entities are Go structs; field references are Go expressions; type mismatches fail at compile time. The permission system is a Go interface implementable with full access to the type system and test infrastructure. Background processes are Temporal workflows with durable execution and automatic retry. The UI is amis JSON schemas served from the API — data, but data generated from `EntityDefinition` declarations.

For teams migrating from Frappe, see [§1.4 Awo vs Frappe — Direct Comparison] for systematic concept mapping, and [§1.4.6 Migration strategy] for a practical migration path.

---

## 1.2 Design Philosophy

### 1.2.1 `EntityDefinition` as the central primitive — not just a schema, a dispatch key

Most frameworks treat entity schema as a database concern — a description of columns and types that drives SQL generation. Awo treats `EntityDefinition` as a dispatch key: a runtime token the framework uses to route every operation through the correct implementation path.

When a request arrives at `POST /api/v1/entities/invoice`, the framework does not look up a handler registered for that route path. It extracts the entity name `invoice` from the path, looks up the corresponding `EntityDefinition` in the registry, and routes the operation through the `EntityDefinition`'s declared execution path. That path includes: persistence backend (system typed columns or JSONB custom fields), field validators, permission policy, hook chain, audit log configuration, and amis page builder. The `EntityDefinition` is the switch.

This dispatch model has a critical implication: adding a new entity to the system does not require writing route handlers, registering middleware, or wiring up dependency injection manually. The `EntityDefinition` registration is the complete declaration of what the entity is and how it behaves. Everything else is derived from it.

The dispatch model also makes framework behavior auditable. To understand what happens when a record is created, you read the `EntityDefinition`. The hooks, validators, permission policy, and persistence path are all declared in one place. There is no hidden middleware, no implicit behavior injected by a library version update.

### 1.2.2 Compile-time correctness for system entities, runtime flexibility for custom entities

The tension between compile-time correctness and runtime flexibility is a genuine architectural tradeoff. A system requiring all entities to be defined at compile time cannot support tenant-specific extensions without redeployment. A system representing all entities as runtime metadata cannot use the compiler to catch field reference errors or missing permission checks.

Awo resolves this by maintaining two entity tiers simultaneously. System entities — LedgerEntry, StockMove, Tenant, User, Payment, and all core ERP entities — are Go structs registered at compile time. Their fields are typed Go fields, their relationships are typed Go references, and field-reference errors are compile errors. Custom entities — tenant-specific extensions, industry-specific records, rapidly evolving configurations — are defined at runtime as metadata stored in the shared `entity_definitions` table, scoped by `tenant_id`, and loaded during tenant boot.

The `EntityResolver` makes this distinction transparent to framework consumers. A caller operating on the `EntityRepository` interface does not need to know whether the entity underneath is a compiled Go struct or a runtime JSONB-backed definition. The same interface methods — `Get`, `Query`, `Create`, `Update`, `Delete` — work on both.

The tradeoff: custom entities cannot use Go's type system for compile-time safety. Validation of custom entity records happens at runtime against `EntityFieldDefinition` metadata. Bugs in custom entity definitions surface at runtime rather than compile time. The validation system surfaces these errors as early as possible — at schema load time during tenant boot rather than at request time.

### 1.2.3 The hybrid data model rationale — why not everything in JSONB, why not everything in typed columns

Two extremes exist in the design space for multi-tenant data storage. At one extreme: every record is stored as JSONB, schema changes require no migrations, and the database behaves as a document store. At the other extreme: every record is stored in strongly typed relational tables, schema changes require migrations, and the database enforces referential integrity at the SQL level.

The all-JSONB approach eliminates migration friction and supports arbitrary tenant customization without redeployment. Its failure mode: the database can no longer enforce integrity constraints, index individual fields efficiently, or participate meaningfully in cross-record queries. For ERP systems this is serious: double-entry accounting requires debits equal credits — a property that must be enforced at the database level, not trusted to application-level logic under concurrent writes. Inventory accuracy requires that stock quantities cannot go below zero without an explicit constraint-level override.

The all-typed-columns approach preserves integrity and query performance but creates a migration problem for multi-tenant deployments: adding a column to a shared table affects every row across every tenant and must be executed with care for zero-downtime. It also makes tenant-specific customization expensive — every customization requires a schema migration and a deployment.

Awo uses a hybrid. System entities requiring integrity guarantees use typed columns with SQL constraints and generated queries. Custom entities requiring flexibility use a `custom_fields` JSONB column on shared tables, with application-level validation against `EntityFieldDefinition` metadata. The boundary is not arbitrary: any entity that participates in financial transactions, inventory accounting, or identity management must be a system entity. Everything else may be a custom entity.

### 1.2.4 Interface-first persistence — the `EntityRepository` contract and why the implementation is replaceable

Database coupling is one of the most common sources of long-term maintainability problems in large Go codebases. When business logic calls a specific query library directly, that library becomes an implicit dependency of the business logic. For a framework intended to be used across many projects and potentially many database access patterns, this coupling is unacceptable.

Awo defines all persistence through the `EntityRepository` interface. Framework consumers — hook implementations, workflow activities, service layer code — interact exclusively with this interface. No concrete database type appears in any position visible to framework consumers.

```go
// `EntityRepository` is the persistence abstraction all framework consumers use.
// Never import any concrete implementation type in code that uses this interface.
package entity

import (
    "context"
    "time"
)

// `EntityRecord` is the unified return type for all entity operations.
// Every record carries tenant_id, org_id, version, and audit metadata
// regardless of whether it is a system entity or a custom entity.
type EntityRecord struct {
    ID       string
    Type     string
    TenantID string
    OrgID    string
    Fields   map[string]any
    Meta     RecordMeta
}

// RecordMeta holds the version and audit columns present on every
// tenant-scoped table. The framework populates these automatically;
// callers never set them directly.
type RecordMeta struct {
    Version   int64     // optimistic-lock counter, incremented on every UPDATE
    CreatedAt time.Time
    CreatedBy string    // actor ID of the user who created the record
    UpdatedAt time.Time
    UpdatedBy string    // actor ID of the user who last modified the record
    DeletedAt *time.Time // non-nil for soft-deleted records
    DeletedBy *string
}

type EntityRepository interface {
    Get(ctx context.Context, entityType, id string) (*EntityRecord, error)
    Query(ctx context.Context, q EntityQuery) ([]*EntityRecord, error)
    Create(ctx context.Context, entityType string, fields map[string]any) (*EntityRecord, error)
    Update(ctx context.Context, entityType, id string, fields map[string]any, expectedVersion int64) (*EntityRecord, error)
    Delete(ctx context.Context, entityType, id string) error
    WithTx(ctx context.Context, fn func(tx `EntityRepository`) error) error
    Aggregate(ctx context.Context, q AggregateQuery) (*AggregateResult, error)
}
```

The `expectedVersion` parameter on `Update` enforces optimistic concurrency control: if the record's current version does not match the caller's expectation, the update is rejected with `ErrVersionConflict`. This prevents lost updates under concurrent modification without requiring the caller to hold a database lock. The framework increments the version column atomically inside the update path.

The interface contract specifies semantics, not implementation. `WithTx` guarantees atomicity; whether that comes from a SQL transaction, an optimistic lock chain, or a saga compensator is an implementation detail. `Aggregate` returns computed summaries; whether those come from a SQL `GROUP BY` or a materialized view is equally irrelevant to the caller.

### 1.2.5 Workflow-first ERP — why business processes belong in Temporal, not in request handlers

The most common architectural mistake in ERP development is encoding business processes as request handlers. A "submit invoice" HTTP handler that atomically validates, checks inventory, reserves stock, posts ledger entries, submits to KRA eTIMS, enqueues notifications, and responds — all within a single synchronous call — is seductive in its simplicity and fragile in ways that compound at production scale.

The deeper problem: ERP business processes are multi-step, long-running, involve external systems (KRA eTIMS, payment gateways, bank APIs, SMS providers), and frequently require human approval steps that may take hours or days. Encoding these in request handlers requires either: making the handler block until all steps complete (fails the moment any step requires human interaction), or making the handler asynchronous with manual coordination logic (re-implementing a workflow engine, badly, without durability or automatic retry).

Temporal is a durable workflow engine. A Temporal workflow is a Go function that survives process restarts, retries failed activities automatically, and can wait indefinitely for signals — human approval decisions, payment callbacks, eTIMS acknowledgements. A "submit invoice" operation in Awo starts a Temporal workflow, records the workflow ID in the entity record, and returns `202 Accepted` immediately. The workflow then executes each step as a separate Temporal activity that can be retried independently.

```go
// InvoiceSubmissionWorkflow runs in a Temporal worker, not an HTTP handler.
// Each Activity call is automatically retried on transient failure.
package workflows

import (
    "fmt"
    "time"

    "go.temporal.io/sdk/temporal"
    "go.temporal.io/sdk/workflow"
    "awo.so/internal/core/finance"
)

type InvoiceSubmissionInput struct {
    TenantID  string
    OrgID     string
    InvoiceID string
    ActorID   string
}

func InvoiceSubmissionWorkflow(ctx workflow.Context, input InvoiceSubmissionInput) error {
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 30 * time.Second,
        RetryPolicy: &temporal.RetryPolicy{
            MaximumAttempts:        3,
            InitialInterval:        time.Second,
            BackoffCoefficient:     2.0,
            NonRetryableErrorTypes: []string{"ErrInvoiceAlreadySubmitted"},
        },
    }
    ctx = workflow.WithActivityOptions(ctx, ao)

    // Step 1: Reserve inventory.
    var reservation finance.ReservationResult
    if err := workflow.ExecuteActivity(ctx, finance.ReserveInventoryActivity, input).Get(ctx, &reservation); err != nil {
        return fmt.Errorf("inventory reservation: %w", err)
    }

    // Step 2: Post double-entry ledger entries.
    if err := workflow.ExecuteActivity(ctx, finance.PostLedgerEntriesActivity, input, reservation).Get(ctx, nil); err != nil {
        // Compensate: release reservation before surfacing error.
        if compErr := workflow.ExecuteActivity(ctx, finance.ReleaseReservationActivity, reservation).Get(ctx, nil); compErr != nil {
            workflow.GetLogger(ctx).Error("reservation release failed after ledger error",
                "reservationID", reservation.ID, "error", compErr)
        }
        return fmt.Errorf("ledger posting: %w", err)
    }

    // Step 3: Submit to KRA eTIMS.
    var etimsRef finance.ETIMSReference
    if err := workflow.ExecuteActivity(ctx, finance.SubmitETIMSActivity, input).Get(ctx, &etimsRef); err != nil {
        // eTIMS failure does not roll back the ledger — invoice is posted.
        // Flag for scheduled resubmission.
        return workflow.ExecuteActivity(ctx, finance.FlagETIMSRetryActivity, input, err.Error()).Get(ctx, nil)
    }

    // Step 4: Wait for manager approval if invoice exceeds KES 500,000.
    if reservation.TotalKES > finance.ManagerApprovalThresholdKES {
        approvalCh := workflow.GetSignalChannel(ctx, "invoice-approval")
        var approval finance.ApprovalDecision
        approvalCh.Receive(ctx, &approval)
        if !approval.Approved {
            return workflow.ExecuteActivity(ctx, finance.RejectInvoiceActivity, input, approval.Reason).Get(ctx, nil)
        }
    }

    return workflow.ExecuteActivity(ctx, finance.NotifyInvoiceSubmittedActivity, input, etimsRef).Get(ctx, nil)
}
```

### 1.2.6 Multi-tenancy as a first-class concern — shared tables with Row-Level Security

Awo uses a shared-table, shared-schema model with PostgreSQL Row-Level Security (RLS) policies to enforce tenant and organizational data isolation. Every table that stores tenant-specific data carries a `tenant_id` column and an `org_id` column. RLS policies on each table ensure that a database session can only read and write rows belonging to the tenant and organizational scope set on that session. Isolation is enforced by the database engine, not by a `WHERE` clause the application must remember to include in every query.

**Why shared tables with RLS rather than schema-per-tenant**

Schema-per-tenant isolation is intuitive and structurally clean: a query running against the wrong schema simply fails at the database level. Its operational costs, however, compound with tenant count. Schema migrations must be applied to every tenant schema individually; connection pools must be partitioned or routed by schema; cross-tenant analytics require dynamic schema stitching; and the PostgreSQL system catalog grows with the number of schemas and their contained objects, creating real catalog contention at scale.

Shared-table RLS imposes isolation through policy rather than namespace. The operational benefits are material: a migration is a single DDL statement run once; a connection pool is shared across all tenants without per-tenant routing; cross-tenant analytics run as straightforward queries with elevated privileges; and the catalog size is bounded by the number of tables, not the number of tenants. The isolation guarantee is equivalent — a session whose `tenant_id` session variable is set to `t_01H9XYZ` cannot read rows where `tenant_id = 't_02H8WXY'` even if it tries — because the database evaluates the policy before the query touches any rows.

**The `org_id` column and organizational hierarchy**

`org_id` represents a node in the tenant's organizational hierarchy — a company, division, branch, department, or cost centre, depending on how the tenant has structured their organization. For platform users (Awo administrators and framework-level service accounts), `org_id` is not evaluated; platform users have cross-organization visibility within their authorized scope. For tenant users, the framework enforces that a user can only access records belonging to organizations at or below their assigned node in the hierarchy.

This means data access clearance is modelled through the organizational tree, not only through role assignments. A finance manager assigned to the East Africa Division node can read invoices belonging to Kenya Branch and Uganda Branch (children of East Africa Division) but cannot read invoices belonging to West Africa Division (a sibling node). Role permissions determine *what actions* a user may take; organizational hierarchy determines *which records* those actions may be applied to.

**Standard columns on every tenant-scoped table**

Every table that stores tenant-specific data must include the following columns. The framework's migration tooling enforces their presence and will refuse to register a table missing any of them.

```sql
-- Columns required on every tenant-scoped table.
-- The framework enforces these at migration time; do not omit them.

tenant_id    TEXT        NOT NULL,  -- Identifies the tenant. Matches the RLS session variable.
org_id       TEXT        NOT NULL,  -- Identifies the organizational node that owns this record.
                                    -- For platform-level records, set to the sentinel value 'platform'.

-- Optimistic concurrency and soft-delete
version      BIGINT      NOT NULL DEFAULT 1,       -- Incremented on every UPDATE. Used for optimistic lock checks.
deleted_at   TIMESTAMPTZ,                          -- NULL = live record. Non-NULL = soft-deleted.
deleted_by   TEXT,                                 -- Actor ID of the user who performed the soft delete.

-- Audit trail
created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
created_by   TEXT        NOT NULL,                 -- Actor ID (user or service account) that created the record.
updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
updated_by   TEXT        NOT NULL,                 -- Actor ID that last modified the record.

-- RLS policy anchor (applied to every tenant-scoped table)
-- The policy reads the session variable set by the connection acquisition logic.
-- No query should ever need to filter on tenant_id or org_id explicitly.
```

A representative table definition for the `invoices` system entity illustrates how these columns compose with business columns:

```sql
CREATE TABLE invoices (
    -- Identity
    id           TEXT PRIMARY KEY,

    -- Tenant and org isolation (required)
    tenant_id    TEXT        NOT NULL,
    org_id       TEXT        NOT NULL,

    -- Business columns
    invoice_no   TEXT        NOT NULL,
    customer_id  TEXT        NOT NULL REFERENCES customers(id),
    total_kes    NUMERIC(18, 2) NOT NULL CHECK (total_kes >= 0),
    status       TEXT        NOT NULL DEFAULT 'DRAFT',
    custom_fields JSONB      NOT NULL DEFAULT '{}',

    -- Versioning and soft-delete
    version      BIGINT      NOT NULL DEFAULT 1,
    deleted_at   TIMESTAMPTZ,
    deleted_by   TEXT,

    -- Audit
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by   TEXT        NOT NULL,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by   TEXT        NOT NULL
);

-- RLS: enable and force for all roles including table owner
ALTER TABLE invoices ENABLE ROW LEVEL SECURITY;
ALTER TABLE invoices FORCE ROW LEVEL SECURITY;

-- SELECT policy: tenant isolation via session variable
CREATE POLICY invoices_tenant_isolation ON invoices
    FOR ALL
    USING (
        tenant_id = current_setting('awo.tenant_id', true)
        AND (
            -- Platform users bypass org filtering
            current_setting('awo.is_platform_user', true) = 'true'
            OR org_id = ANY(
                -- org_ids returns the current user's accessible org subtree
                string_to_array(current_setting('awo.accessible_org_ids', true), ',')
            )
        )
        AND deleted_at IS NULL  -- soft-deleted rows are invisible by default
    );
```

**Tenant context flows through the session, not the query**

Tenant context flows through the system via the request context. The middleware layer resolves the tenant from the HTTP host header (subdomain routing) or `X-Tenant-ID` header, validates the session, and injects tenant and organizational identity into the context before the request reaches any handler. The `EntityRepository` implementation reads these values from the context to set the correct PostgreSQL session variables before executing any query. Tenant and org isolation are not parameters you pass to every database call — they are structural properties of the session.

```go
// TenantContext carries all identity required to scope a database session.
// The `EntityRepository` sets the corresponding PostgreSQL session variables
// (awo.tenant_id, awo.org_id, awo.accessible_org_ids, awo.is_platform_user)
// on every connection acquired from the pool.
package tenancy

import (
    "context"
    "fmt"
)

type contextKey struct{}

// TenantContext is resolved once per request by the middleware layer
// and injected into the request context. All downstream code reads
// tenant identity from here; nothing constructs a TenantContext manually.
type TenantContext struct {
    TenantID         string   // opaque tenant identifier, e.g. "t_01H9XYZ"
    OrgID            string   // the organizational node this session is scoped to
    AccessibleOrgIDs []string // this node and all descendants in the org tree
    IsPlatformUser   bool     // true for Awo admins; bypasses org-level filtering
    Locale           string   // BCP-47 language tag, e.g. "sw" for Swahili
    Timezone         string   // IANA tz database name; "Africa/Nairobi" for Kenya
}

func WithTenant(ctx context.Context, t TenantContext) context.Context {
    return context.WithValue(ctx, contextKey{}, t)
}

func TenantFromContext(ctx context.Context) (TenantContext, error) {
    t, ok := ctx.Value(contextKey{}).(TenantContext)
    if !ok {
        return TenantContext{}, fmt.Errorf("tenancy: no tenant in context; " +
            "ensure this code is called within a tenant-scoped request or activity")
    }
    return t, nil
}
```

Custom entity definitions are stored in the shared `entity_definitions` table, filtered by `tenant_id` during tenant boot. When a request arrives for a tenant whose EntityRegistry is not yet populated, the boot sequence loads all custom entity definitions belonging to that tenant and registers them with the `EntityResolver`. Tenant-specific customizations are available within the first request from that tenant, without a global process restart.

### 1.2.7 Server-driven UI as a force multiplier — eliminating the frontend bottleneck

Traditional ERP development requires a frontend team to translate every backend change into UI updates. Add a field to a system entity and a frontend developer must add that field to every form and list view. Multiply this by the number of custom entities a tenant may define and the number of views referencing each entity, and frontend development becomes the bottleneck that determines how fast the system can evolve.

Awo uses amis as its UI layer. In amis, the UI is described by a JSON schema that the server delivers to the browser at page load time. The browser interprets the schema and renders components — forms, tables, charts, dashboards — without requiring custom JavaScript. The schema is data, not code, and it is generated by the framework from `EntityDefinition` declarations.

When you register an `EntityDefinition`, the framework's UI builder derives a default set of amis schemas: a list view, a create form, an edit form, and a detail view. These defaults are complete and functional without additional work. The force multiplier effect is most apparent in the tenant dimension: a tenant administrator can define a custom entity through the admin interface, add fields to it, and immediately see a working list view and create form in the browser — without any frontend developer involvement, without a deployment, and without any code change.

> **Note:** amis does not replace all custom frontend development. Complex dashboards, highly interactive visualizations, and tenant-branded landing pages often require custom amis schema composition or custom amis component registration. The framework provides the generation defaults; custom composition is a supported extension point.

---

## 1.3 Who This Documentation Is For

### 1.3.1 Framework contributors — building Awo core

Framework contributors are Go developers working on the Awo codebase itself: the `EntityDefinition` system, the `EntityResolver`, the permission evaluation engine, the amis schema builder, the Temporal integration layer, the multi-tenancy infrastructure, and the `EntityRepository` interface with its reference implementation. This is the smallest audience and the one requiring the deepest understanding of every layer.

For framework contributors, the most important sections are Part II (The `EntityDefinition` System), Part III (The Persistence Layer), and Part IV (The Workflow Layer). Understanding the `EntityResolver` dispatch model (see [§2.2.3]) and the `EntityRepository` interface contract (see [§1.2.4]) is a prerequisite for making changes to framework core without breaking the guarantees that module developers and application developers depend on.

Framework contributors must understand the testing model. Awo's test suite uses real PostgreSQL instances, not mocks, because the RLS policy model and the custom entity JSONB storage are PostgreSQL-specific and cannot be meaningfully tested against an in-memory mock. Any change to the persistence layer must be accompanied by integration tests that run against a real PostgreSQL database with RLS enabled and with correctly configured session variables — testing without RLS enabled is not sufficient, because the policy evaluation path differs.

Read all of Part I before touching any implementation. Design decisions that appear arbitrary often reflect constraints from the ERP domain, the Go type system, or the multi-tenancy model that are not immediately obvious from the code alone.

### 1.3.2 Module developers — building built-in ERP modules on top of Awo

Module developers build the ERP domain modules: Finance (chart of accounts, ledger entries, journals, reconciliation), Inventory (items, warehouses, stock movements, batch tracking), Procurement (purchase orders, supplier invoices, goods receipts), HR (employees, payroll, leave management), and equivalent domain modules for other ERP functions.

The primary extension points for module developers are: `EntityDefinition` registration, hook implementations, permission policy bindings, and Temporal workflow and activity definitions. These are covered in Part II through Part V.

Module developers do not modify the framework core. They implement the interfaces the framework defines and register their implementations at startup. A module is a Go package that exports a `Register(registry *entity.Registry) error` function. The framework calls this function during the boot sequence.

Every entity table a module introduces must include the standard `tenant_id`, `org_id`, version, and audit columns described in [§1.2.6]. The module's migration files are validated against the framework's column checklist at registration time; a module whose tables are missing required columns will fail to register, not merely log a warning.

For module developers building for the Kenyan market: currency amounts are KES formatted as `KES 1,234.56`, tax calculations reference KRA tax rates and codes, and eTIMS submission is a required step in the invoice and credit note lifecycle. These are structural properties of the Finance module's entity definitions and workflow definitions, not configuration options.

### 1.3.3 Application developers — building tenant-specific customisations

Application developers work within a deployed Awo instance for a specific tenant and use Awo's extension mechanisms to add tenant-specific fields, entities, workflows, and UI layouts. Application developers may or may not be Go developers — many tenant customizations are achievable entirely through the admin UI.

Application developers who do write code typically produce: custom entity definitions, custom hook implementations for tenant-specific validation logic, and custom amis schema overrides. They do not modify core framework code or module code. Their customizations are scoped to their tenant through the `tenant_id` column in the `entity_definitions` table and are isolated from every other tenant by the RLS policies described in [§1.2.6].

The relevant documentation for application developers is primarily Part VI (Customisation), Part VII (Permissions), and Part VIII (API Reference). The conceptual chapters in Part I are useful for understanding why certain constraints exist — particularly the custom entity vs system entity distinction in [§2.3 When to Use Each Entity Type] and the organizational hierarchy access model in [§7.3 Organizational Scope and Data Clearance].

> **Note:** Application developers who need capabilities that cannot be achieved through the declared extension points should engage a module developer to build a proper framework module rather than attempting to work around the framework's boundaries. Working around the boundaries typically means losing the RLS-enforced isolation, permission evaluation, and audit trail that make the system safe to operate.

### 1.3.4 System integrators — deploying and operating Awo for a client

System integrators deploy Awo for specific clients, configure it for the client's environment, onboard initial tenant data, and operate it in production. They are primarily concerned with deployment configuration, database provisioning, tenant lifecycle management, observability, and disaster recovery.

The relevant documentation is Part IX (Deployment), Part X (Operations), and Part XI (Security). The API reference in Part VIII is relevant for integrators who need to automate tenant provisioning or synchronize data with external systems.

For Kenyan deployments, system integrators are responsible for ensuring KRA eTIMS connectivity is configured and verified before any tenant goes live in production. eTIMS integration requires a taxpayer PIN and TCC (Tax Compliance Certificate) from KRA for each tenant, correct configuration of the eTIMS API credentials, and successful execution of the eTIMS test environment integration checklist before the production endpoint is activated. A tenant with an incorrectly configured eTIMS integration will fail to post any taxable transaction.

System integrators are also responsible for verifying that RLS is correctly active on the production database before any tenant data is loaded. The framework provides a pre-flight check (`awo db rls-verify`) that connects to the database as an unprivileged role and asserts that cross-tenant row reads are rejected by the policies on all registered tables. This check must pass before the first tenant is provisioned.

### 1.3.5 How to navigate this documentation by audience

The documentation is structured to be read front-to-back by framework contributors, who need every part. For other audiences:

| Audience | Start here | Core reading | Can defer |
|---|---|---|---|
| Framework contributor | §1 (this chapter) | §2, §3, §4, §5, §6, §7, §8 | Nothing |
| Module developer | §1.1, §1.2 | §2, §4, §5, §6 | §8 implementation internals (but read the `EntityRepository` interface contract in §1.2.4) |
| Application developer | §1.1, §2.3 | §6 (Customisation), §7, §8 (API) | §3, §4, §5 internals |
| System integrator | §1.1, §1.2.6 | §9, §10, §11 | §2 through §8 |

Cross-references throughout use the format `[§{number} {Title}]`. When a concept is defined in an earlier section and referenced later, the cross-reference carries you back to the full explanation. Follow cross-references rather than searching.

---

## 1.4 Awo vs Frappe — Direct Comparison

### 1.4.1 DocType vs `EntityDefinition`

Frappe's DocType is the central abstraction of the Frappe framework: a record in the `tabDocType` database table that describes a class of documents — its fields, permissions, form layout, and linked documents. DocTypes are data stored in MariaDB and loaded at runtime. A DocType definition can be created, edited, and deleted through Frappe's admin interface without any code change.

Awo's `EntityDefinition` is conceptually equivalent but structurally different. For system entities, an `EntityDefinition` is a Go struct registered at compile time. The definition is code, not data: it participates in the Go type system. Field references in hook code are Go expressions that the compiler verifies. Relationship traversals are typed method calls, not string lookups against a runtime registry. For custom entities, `EntityDefinition`s are records stored in the shared `entity_definitions` table, scoped to their tenant by `tenant_id` — closer to the Frappe model — and they drive the same dispatch mechanism through the `EntityResolver`.

Adding a field to a Frappe DocType is an action a non-developer can take through the admin interface immediately, without a deployment. Adding a field to an Awo system entity requires a Go code change, a database migration, and a deployment. Adding a field to an Awo custom entity requires only an admin interface action, like Frappe. Awo's system entity development cycle is deliberately slower: you are trading iteration speed for compile-time correctness.

The other key difference: Frappe DocType definitions are per-installation (shared across all sites), while Awo's custom `EntityDefinition`s are per-tenant, isolated by RLS. System `EntityDefinition`s in Awo are per-installation, consistent with Frappe's behavior. For multi-tenant deployments where different tenants operate in different industries — one in manufacturing, another in professional services — the per-tenant custom entity model allows each tenant's data model to diverge without affecting any other tenant.

### 1.4.2 Python hooks vs Go interfaces

Frappe's hook system uses Python's dynamic dispatch: a hook is a Python function reference expressed as a dotted string path stored in a `hooks.py` file. Frappe discovers hooks by scanning installed apps and loading their `hooks.py` at startup. The system is flexible — any Python function can be a hook — but provides no static verification that the hook signature is correct. A hook receiving the wrong number of arguments fails at runtime, not at load time.

Awo's extension points are Go interfaces. A before-insert hook is a value that implements the `entity.BeforeCreateHook` interface, and the compiler verifies this at build time:

```go
// BeforeCreateHook is called before a new entity record is persisted.
package entity

import "context"

type BeforeCreateHook interface {
    BeforeCreate(ctx context.Context, entityType string, fields map[string]any) error
}

// Compile-time verification that MinimumAmountValidator satisfies BeforeCreateHook.
var _ BeforeCreateHook = (*MinimumAmountValidator)(nil)

type MinimumAmountValidator struct {
    MinimumKES float64
}

func (v *MinimumAmountValidator) BeforeCreate(ctx context.Context, entityType string, fields map[string]any) error {
    total, ok := fields["total_amount"].(float64)
    if !ok {
        return fmt.Errorf("invoice: total_amount must be numeric, got %T — "+
            "Kiasi cha ankara lazima kiwe namba", fields["total_amount"])
    }
    if total < v.MinimumKES {
        return fmt.Errorf("invoice: total KES %.2f is below minimum of KES %.2f — "+
            "Jumla ya KES %.2f ni chini ya kiwango cha chini cha KES %.2f",
            total, v.MinimumKES, total, v.MinimumKES)
    }
    return nil
}
```

A hook with the wrong method signature does not compile. The tradeoff is expressiveness: Frappe's hook system allows any Python expression at any hook point. Awo's interface model requires hooks conform to a declared signature. Teams needing hook behavior outside an existing interface must propose a new extension point to the framework maintainers.

### 1.4.3 MariaDB-first vs PostgreSQL-first

Frappe is built on MariaDB. Its data model exploits MariaDB's TEXT column type, its full-text search uses MariaDB FULLTEXT indexes, and its naming series generation relies on MariaDB autoincrement behavior. Frappe can run on MySQL but not on PostgreSQL.

Awo is PostgreSQL-first and does not support any other database. This is a structural dependency, not a product positioning choice. Awo's data isolation uses PostgreSQL Row-Level Security policies — a feature with no drop-in equivalent in MariaDB. Awo's custom entity storage uses the PostgreSQL `JSONB` type with GIN indexes for field-level querying. Awo's migration system uses Atlas CLI's PostgreSQL driver. Awo's audit trail uses PostgreSQL triggers to capture row-level changes. None of these have meaningful equivalents in MariaDB.

PostgreSQL's implementation of serializable snapshot isolation (SSI) provides stronger consistency guarantees under concurrent writes than MariaDB's locking-based serializable isolation mode. For double-entry accounting — where concurrent journal entries must maintain the invariant that debits equal credits — SSI eliminates a class of anomalies that require explicit application-level locking to prevent in MariaDB. PostgreSQL's RLS evaluation is integrated into the query planner and executes before any rows are touched, making it structurally impossible for a correctly configured session to observe another tenant's data, not merely unlikely.

> **Note:** Teams migrating from Frappe to Awo must migrate their data from MariaDB to PostgreSQL. See [§1.4.6 Migration strategy] for tooling recommendations.

### 1.4.4 Frappe Workflow vs Temporal

Frappe's Workflow module allows non-developers to define state machines — a document can be in state A, B, or C, transitions require specific roles, and transitions may trigger email notifications. This is sufficient for simple single-approver document approval. It breaks down for: parallel approvals where multiple parties must approve independently before a transition; compensating transactions where a failed step must roll back prior steps; long-running external waits where the workflow must pause for a payment callback or KRA eTIMS acknowledgement; and retry logic where a failed step should retry automatically with backoff.

Temporal addresses all of these. A Temporal workflow is a durable function execution: the function runs, persists its execution state after every step, and resumes from the last persisted state after a process restart. Activities are individual steps in the workflow retried automatically on failure according to a configurable retry policy. The workflow can block on named signal channels and resume when the signal arrives. Parallel execution is expressed as concurrent goroutines in the workflow function.

The tradeoff is operational complexity. Frappe Workflow is a database table and a UI configuration screen with no additional infrastructure requirements. Temporal is a separately deployed cluster with its own storage backend, observability requirements, upgrade lifecycle, and failure modes. Running Awo in production requires operating a Temporal cluster.

For the workflows that Awo's Finance and Procurement modules implement — invoice approval with eTIMS submission, multi-level purchase order approval with KES authorization limits, payroll run processing, supplier payment batching — the operational overhead is justified. These workflows have enough error surface area, enough external system interaction, and enough regulatory consequence that ad-hoc retry logic in a request handler is not a responsible choice.

### 1.4.5 Jinja forms vs amis SDUI

Frappe renders its main interface using a rich JavaScript desk application that communicates with the Frappe API. Customizing the desk requires writing client-side JavaScript and maintaining it across Frappe version upgrades. Print formats use Jinja HTML templates rendered server-side.

Awo's UI layer is amis, a JSON-schema-driven UI framework where the server delivers a JSON document describing the complete UI at page load time. The amis SDK in the browser interprets the schema and renders the appropriate components. Application developers customize the UI by modifying the JSON schema — they do not write JavaScript for standard entity views.

The amis approach has one firm limitation: the amis component library is finite. UI requirements going beyond the available amis components — real-time collaborative interfaces, advanced geospatial visualizations, highly interactive workflow canvases — require either building a custom amis component (a React component conforming to the amis extension interface) or delivering that UI through a separate frontend application that uses the Awo API directly.

Frappe's desk is more flexible in that it is a full JavaScript application that developers can extend with arbitrary JavaScript. It is less maintainable in that Frappe desk customizations written against one Frappe major version frequently break on upgrade, creating a recurring maintenance cost that many Frappe deployments absorb silently until it becomes a migration blocker.

### 1.4.6 Migration strategy for teams coming from Frappe

Teams migrating from Frappe to Awo face three categories of work: data migration (moving records from MariaDB to PostgreSQL), concept mapping (translating DocTypes, hooks, workflows, and permissions to their Awo equivalents), and UI translation (converting Frappe desk customizations to amis schemas).

Data migration is the most mechanically routine but operationally riskiest category. The recommended process:

1. Export each Frappe DocType's data as JSON using Frappe's built-in Data Export tool. Export in batches by DocType, not as a full-site dump, to enable per-DocType validation.
2. Map each DocType to its Awo equivalent. DocTypes that correspond to Awo system entities (Invoice, StockEntry, Supplier, Customer) need their data transformed to match the typed column layout, including the addition of `tenant_id`, `org_id`, and audit columns that the source system did not carry. DocTypes that correspond to Awo custom entities need their field data stored as JSONB in the `custom_fields` column.
3. Execute the migration by calling the Awo entity creation API for each record rather than inserting directly into the database. Using the API ensures the hook chain runs, audit columns are populated by the framework (not manually), and RLS policies are exercised rather than bypassed. The API accepts the `X-Migration-Source: frappe` header which disables hooks that would otherwise reject records missing fields that did not exist in the source system.
4. Validate migrated data by running integrity checks through the Awo API: ledger balance assertions, inventory quantity reconciliation, and record count comparisons between the Frappe export and Awo query results. Run the `awo db rls-verify` pre-flight check after data load to confirm that cross-tenant isolation is intact.

Concept mapping is straightforward for hooks and permissions. Frappe `before_insert` maps to `BeforeCreateHook`. Frappe `on_submit` maps to an `AfterCreateHook` filtered to the SUBMITTED status transition. Frappe role-based permissions map to Awo permission policies, extended by org-hierarchy scope described in [§7.3]. Frappe Workflows require re-expression as Temporal workflow functions — the state model is preserved but the implementation changes from a database configuration to Go code. This step requires a Go developer.

UI translation requires the most amis-specific knowledge. Frappe desk list views and form views have functional equivalents in framework-generated amis schemas; for simple forms, generated defaults require only minor customization. Frappe pages with significant JavaScript customization must be re-expressed as amis schema properties (`visibleOn`, `source`, `validations`). There is no automated translation tool; it requires manual schema authorship for each customized view.

> **Warning:** Do not run Frappe and Awo in parallel against the same data during migration. The data models are incompatible at the field level, and concurrent writes from both systems will produce inconsistent state with no safe merge path. The migration must be a hard cutover: freeze Frappe writes at a defined timestamp, migrate the snapshot of data at that timestamp, validate in Awo — including RLS verification — then switch traffic.

Teams with extensive Frappe custom app code should have a module developer audit the custom code before beginning migration. The audit identifies: code that maps cleanly to Awo extension points (hooks, permission policies, `EntityDefinition`s), code that requires rethinking (heavy use of the global `frappe` object, direct MariaDB queries, client-side form scripts), and code that has no equivalent in Awo and represents product scope that may be deferred.
