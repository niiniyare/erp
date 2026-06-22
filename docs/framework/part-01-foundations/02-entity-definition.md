# The EntityDefinition — The Central Abstraction

## 2.1 Why EntityDefinition Is the Central Primitive

### 2.1.1 What the framework does when it sees an EntityDefinition

When a module developer calls `registry.Register(invoiceDefinition)`, the framework does not simply store a schema object in a map. It executes a registration pipeline that derives an entire set of runtime artifacts from the EntityDefinition: it generates API route handlers for CRUD and action endpoints, compiles permission policy bindings into the enforcer, registers hook chains in the lifecycle manager, builds a default amis page schema set for each view type, and registers the entity with the EntityResolver so it can be dispatched at request time.

This registration pipeline runs once at startup for system entities, and once per-tenant-boot for custom entities. After registration, the EntityDefinition does not participate in normal request processing — the derived artifacts do. The EntityDefinition is a declaration; the derived artifacts are the executable system. If you change an EntityDefinition and do not restart the process (for system entities), or do not trigger a tenant reload (for custom entities), the system continues operating on the previously derived artifacts from the prior registration.

The registration pipeline is intentionally opaque to module developers. You declare what your entity is; the framework decides how to implement every aspect of how it behaves. This is the key to the framework's productivity claim: a single declaration drives six different subsystems simultaneously without requiring you to wire any of them.

Understanding what the registration pipeline derives is useful when debugging unexpected behavior. If an API route is returning an unexpected permission error, the relevant artifact is the permission policy that was compiled from the EntityDefinition at registration time. If an amis form is showing unexpected fields, the relevant artifact is the page schema that was generated at registration time. Reading the EntityDefinition tells you what was intended; reading the derived artifacts tells you what was compiled.

### 2.1.2 The five things an EntityDefinition drives

An EntityDefinition simultaneously drives five distinct subsystems. Each subsystem consumes a different part of the EntityDefinition, and each has a different timing for when the derived artifact is computed.

**Persistence routing** — The EntityDefinition's `Type` field (system or custom) and its field schema are consumed by the EntityResolver to determine which persistence path to dispatch operations to. For system entities, the field schema maps to typed SQL columns in generated queries. For custom entities, the field schema maps to JSONB storage with GIN-indexed paths. The persistence routing is derived at registration time and stored in the EntityResolver's dispatch table.

**API generation** — The EntityDefinition's name, fields, and declared actions are consumed by the route generator to create Fiber route handlers for all standard CRUD operations plus any declared custom actions. The URL structure follows the convention `/api/v1/{entity-name}` (see [§17 REST API Conventions]). The route handlers are registered on the Fiber app at startup, not at request time.

**UI generation** — The EntityDefinition's fields, edges, and optional page builder override functions are consumed by the amis page builder to generate JSON schemas for the list view, create form, edit form, and detail view. Default page schemas are generated at tenant boot and cached in Redis with a TTL keyed to the entity definition version.

**Permission evaluation** — The EntityDefinition's permission policy bindings are compiled into the Casbin enforcer at registration time. Every request to an entity endpoint is evaluated against these compiled policies before any hook or persistence operation executes. The permission check cannot be bypassed by hook code because the check runs before hooks are invoked.

**Workflow triggering** — The EntityDefinition's workflow trigger bindings declare which Temporal workflows to start in response to specific lifecycle events (e.g., on_submit, on_status_change). These bindings are stored in the lifecycle manager's trigger table and evaluated by the `after_save` hook dispatcher.

### 2.1.3 Why this is not just a schema definition tool

Schema definition tools like JSON Schema, Protobuf, and OpenAPI describe data shapes. They tell you what fields a record has and what types those fields carry. They are inputs to code generation or validation logic that you wire up yourself. The EntityDefinition goes further: it is a complete behavioral declaration for a class of records within the framework runtime.

The distinction matters in practice. If you use a schema definition tool to describe an invoice, you still have to separately: write route handlers, implement permission checks, set up a validation pipeline, connect to a persistence layer, wire UI rendering, and integrate with workflow triggers. The schema tool does not do any of this for you.

If you declare an Awo EntityDefinition for an invoice, all of this is derived automatically. The declaration is the wiring. This is why the EntityDefinition is described as a dispatch key and not just a schema: it is the token that the framework's runtime uses to route every operation on invoice records through the entire infrastructure the framework provides.

---

## 2.2 The Two Entity Types

### 2.2.1 System entities — strongly typed, SQL-backed

System entities are Go struct types registered with the framework at compile time. The framework knows about system entities before the process starts accepting requests. Their fields correspond to typed SQL columns in PostgreSQL, and all database operations against system entities use generated queries (via SQLC or equivalent) that are verified against the database schema at code generation time.

A system entity definition in Go looks like a struct with framework field tags:

```go
// Invoice is a system entity. Its fields map to typed SQL columns.
// Registered with the EntityRegistry at module init time.
package finance

import (
    "time"
    "awo.so/internal/entity"
)

// InvoiceDefinition declares the Invoice system entity.
var InvoiceDefinition = entity.SystemDefinition{
    Name:        "invoice",
    Module:      "finance",
    Label:       "Invoice",
    LabelPlural: "Invoices",
    Fields: []entity.FieldDef{
        {Name: "number",     Type: entity.FieldNamingSeries, Series: "INV-{YYYY}-{SEQ:5}"},
        {Name: "customer",   Type: entity.FieldLink, LinkTarget: "customer", Required: true},
        {Name: "status",     Type: entity.FieldSelect, Options: []string{"Draft", "Submitted", "Paid", "Cancelled"}, Default: "Draft"},
        {Name: "issue_date", Type: entity.FieldDate, Required: true},
        {Name: "due_date",   Type: entity.FieldDate, Required: true},
        {Name: "total_kes",  Type: entity.FieldCurrency, Required: true},
        {Name: "tax_amount", Type: entity.FieldCurrency},
        {Name: "etims_ref",  Type: entity.FieldData, Immutable: true},
        {Name: "notes",      Type: entity.FieldLongText},
    },
    Edges: []entity.EdgeDef{
        {Name: "lines", Target: "invoice_line", Type: entity.EdgeOneToMany, CascadeDelete: true},
    },
    Hooks: entity.HookSet{
        BeforeCreate: []entity.BeforeCreateHook{&InvoiceDraftValidator{}},
        AfterCreate:  []entity.AfterCreateHook{&InvoiceNumberAssigner{}},
        OnSubmit:     []entity.LifecycleHook{&InvoiceSubmitTrigger{}},
    },
}
```

System entities that participate in double-entry accounting, inventory accuracy, or identity management must be system entities. Their fields are enforced with SQL constraints, their queries are generated and type-checked, and their data cannot be corrupted by a malformed JSON field value.

### 2.2.2 Custom entities — tenant-defined, JSONB-backed, metadata-driven

Custom entities are defined by tenant administrators at runtime through the admin interface or the configuration API. Their definitions are stored in the tenant's database schema as `EntityDefinition` and `EntityFieldDefinition` records, not in Go code. The framework loads them from the database during tenant boot and registers them with the EntityResolver alongside system entities.

Because custom entity definitions are data rather than code, they cannot use the Go type system for compile-time safety. A custom entity's fields are described as `EntityFieldDefinition` records (name, type, required, validations), and all records are stored as JSONB in the tenant's `custom_entity_records` table. Field-level validation happens at the application layer, not at the database constraint layer.

Custom entities are appropriate for: tenant contact extensions, custom approval metadata, industry-specific fields, and rapidly evolving data shapes that would require too many redeployments if encoded as system entities. They are not appropriate for any entity that participates in financial transactions, inventory accounting, or identity management.

### 2.2.3 The EntityResolver — how the framework picks the execution path at runtime

The EntityResolver is a registry component that sits between the route handler and the EntityRepository. When a request arrives for entity type `vehicle_inspection` (a custom entity defined by a petroleum company tenant), the EntityResolver:

1. Looks up the entity name in the tenant's EntityRegistry.
2. Determines whether the entity is a system entity or a custom entity.
3. Returns an `EntityRepository` implementation scoped to that entity's execution path — either the system entity SQL path or the custom entity JSONB path.

The route handler receives an `EntityRepository` in either case. It calls `repo.Create(ctx, fields)` or `repo.Query(ctx, filter)` without knowing or caring which path the resolver selected. The resolver's decision is an implementation detail hidden behind the interface contract.

```go
// EntityResolver dispatches to the correct repository implementation.
// This function is called by the route handler layer, not by module code.
package entity

import (
    "context"
    "fmt"
    "awo.so/internal/tenancy"
)

type Resolver struct {
    systemRegistry map[string]*SystemDefinition
    // tenantRegistries is keyed by tenant ID.
    // Populated during tenant boot; protected by per-tenant RWMutex.
    tenantRegistries map[string]*TenantRegistry
}

// Resolve returns an EntityRepository scoped to the entity type and tenant in ctx.
// Returns ErrEntityNotFound if the entity is not registered for the tenant.
func (r *Resolver) Resolve(ctx context.Context, entityType string) (EntityRepository, error) {
    tc, err := tenancy.TenantFromContext(ctx)
    if err != nil {
        return nil, fmt.Errorf("resolver: %w", err)
    }

    // System entities take priority over custom entities with the same name.
    if def, ok := r.systemRegistry[entityType]; ok {
        return newSystemEntityRepo(ctx, def, tc), nil
    }

    tr, ok := r.tenantRegistries[tc.ID]
    if !ok {
        return nil, fmt.Errorf("resolver: tenant %s not booted", tc.ID)
    }

    if def, ok := tr.CustomEntities[entityType]; ok {
        return newCustomEntityRepo(ctx, def, tc), nil
    }

    return nil, fmt.Errorf("resolver: entity type %q not found for tenant %s: %w",
        entityType, tc.ID, ErrEntityNotFound)
}
```

### 2.2.4 What callers see — the EntityRecord as a unified surface

Regardless of whether an entity is a system entity backed by typed SQL columns or a custom entity backed by JSONB, all read and write operations on the EntityRepository interface return and accept `*EntityRecord` values. The EntityRecord is a unified envelope that carries entity data in a `map[string]any` field map alongside framework-managed metadata.

For system entities, the field map is populated by deserializing typed SQL column values. An integer column is populated as `int64`, a numeric(20,4) column is populated as `decimal.Decimal`, a timestamptz column is populated as `time.Time` in the EAT timezone. For custom entities, the field map is populated by deserializing the JSONB blob with field types coerced according to the EntityFieldDefinition.

The unified surface means that framework consumers — hooks, page builders, workflow activities — can be written once and work with both entity types without any conditional logic. A hook that reads `fields["total_kes"]` works the same way whether `invoice` is a system entity or a custom entity.

### 2.2.5 Why this distinction is invisible to framework consumers by design

The system entity / custom entity distinction is an internal routing concern, not a caller concern. Exposing it to framework consumers would mean that every hook, every page builder function, and every workflow activity would need to handle two different code paths depending on entity type. That would make the framework's extension model significantly more complex and would create a barrier to writing reusable module code.

The invisibility is achieved by the EntityResolver (see [§2.2.3]) and the unified EntityRecord type (see [§2.2.4]). The resolver selects the execution path; the EntityRecord normalizes the output. Every layer above the resolver sees only the interface.

---

## 2.3 When to Use Each Entity Type

### 2.3.1 Use a system entity when

System entities exist because some records require guarantees that only the database layer can provide. Use a system entity when any of the following conditions apply:

**Financial integrity is required.** Any entity that participates in double-entry accounting — journal entries, ledger entries, payment records, bank transactions — must be a system entity with typed numeric columns and SQL constraints. `numeric(20,4)` storage prevents floating-point rounding errors that are unacceptable in financial records. SQL CHECK constraints prevent debit/credit imbalance at the database level, not just at the application level.

**Inventory accuracy is required.** Stock movements, warehouse receipts, and issue records must be system entities. Negative stock quantities must be prevented by a SQL constraint (`CHECK (quantity_on_hand >= 0)`) enforced at the row level, not by application-level logic that can be bypassed under concurrent writes.

**Identity and access management.** User records, session records, and role assignment records must be system entities. IAM data must never be stored in a JSONB blob where a malformed write could corrupt permission data without a database-level constraint catching it.

**High-frequency writes.** Custom entity JSONB storage is efficient for moderate write rates, but system entity typed columns benefit from PostgreSQL's more efficient indexing strategies. If an entity receives more than a few hundred writes per second, evaluate system entity storage.

### 2.3.2 Use a custom entity when

Custom entities exist because the alternative — requiring a framework code change and redeployment for every new tenant-specific data shape — would make the framework unusable for its primary audience: businesses with evolving, industry-specific data requirements.

Use a custom entity when: the entity represents tenant-specific business data with no cross-tenant integrity requirements; the entity's schema is expected to evolve frequently during the tenant's early operational phase; the entity does not participate in financial transactions or inventory accounting; and the entity's write rate is low enough that JSONB storage is efficient.

Good candidates for custom entities: site visit records for field service companies, vehicle inspection checklists for logistics companies, customer satisfaction survey responses, custom approval workflow metadata, and industry-specific classification fields that extend but do not replace system entity fields.

### 2.3.3 Entities that must never be custom

The following entities must always be system entities regardless of business requirements, regulatory context, or tenant preferences:

| Entity | Reason |
|---|---|
| LedgerEntry | Participates in double-entry accounting; requires SQL numeric constraints |
| StockMove | Participates in inventory accounting; requires SQL quantity constraints |
| Payment | Financial transaction record; requires SQL constraints and audit triggers |
| User | IAM data; corruption risk in JSONB storage |
| Tenant | Platform identity; must be accessible before per-tenant schemas are loaded |
| JournalEntry | Double-entry accounting; debit/credit balance enforced at DB level |
| TaxEntry | KRA eTIMS submission record; regulatory compliance requires integrity |

This list is not configurable. It is enforced by the EntityRegistry's registration validator, which rejects registration attempts for these entity names as custom entities.

### 2.3.4 Entities that should always be custom

These entity categories are poor fits for system entities because they are inherently variable, tenant-specific, and do not require database-level integrity constraints:

Tenant contact extension data (additional fields on contacts that differ by industry), custom approval metadata (tenant-specific approval reason codes, escalation notes), locale-specific classification fields (Kenyan county codes on addresses, VAT category codes for specific sectors), integration mapping records (external system ID mappings that differ by tenant), and user preference records.

### 2.3.5 The grey zone — when to escalate a custom entity to a system entity

A custom entity should be escalated to a system entity when it begins to exhibit characteristics that require database-level enforcement:

When a custom entity accumulates more than 10 million records in production, JSONB storage and GIN indexing may become a performance bottleneck for complex queries. When a custom entity's fields are referenced in financial calculations that require numeric precision greater than what JSON number representation can provide. When a custom entity needs a foreign key constraint pointing to a system entity's primary key — a constraint that cannot be enforced across the JSONB / typed-table boundary.

The escalation process: the custom entity definition is frozen, a system entity with equivalent fields is defined by a module developer, a data migration copies records from the custom entity's JSONB storage to the new typed columns, and the EntityResolver mapping is updated to point the entity name to the new system entity path. During the migration window, both paths may need to be active simultaneously.

---

## 2.4 EntityDefinition Anatomy

### 2.4.1 Name and identifier conventions

The entity name is the stable identifier used throughout the system — in URL paths, in Redis keys, in Temporal workflow IDs, in amis page schema keys, and in permission policies. Names must be lowercase, underscore-separated, and unique within the tenant scope.

System entity names are globally unique within the framework installation. Custom entity names are unique within a tenant but may collide with other tenants' custom entities or with system entity names (though system entity names take priority in the resolver). The recommended naming convention for module developers is `{module}_{noun}` — for example, `finance_journal`, `inventory_stock_move`, `forecourt_shift`. This convention prevents name collisions between modules without requiring a namespace prefix.

Avoid names that change: the entity name is embedded in migration file names, in Temporal workflow IDs that are stored in Temporal's history database for years, and in Redis keys that may be in cache across a deployment boundary. Renaming an entity name is a breaking change that requires a coordinated migration of all dependent systems.

### 2.4.2 Field list — types, constraints, metadata

Fields are declared as a slice of `entity.FieldDef` values. Each field declaration specifies: the field name (stable, underscore-separated), the field type (see [§5 Field System] for the complete type reference), and any constraints and metadata applicable to that type.

```go
// Field declaration examples covering the most common patterns.
Fields: []entity.FieldDef{
    // Required linked field.
    {Name: "customer", Type: entity.FieldLink, LinkTarget: "customer", Required: true},

    // Currency field — stored as numeric(20,4), never floating point.
    {Name: "total_kes", Type: entity.FieldCurrency, Required: true},

    // Select field with declared option set.
    {Name: "status", Type: entity.FieldSelect,
        Options: []string{"Draft", "Submitted", "Paid", "Cancelled"},
        Default: "Draft"},

    // Sensitive field — excluded from logs and standard API responses.
    {Name: "bank_account_number", Type: entity.FieldData, Sensitive: true},

    // Immutable field — set on create, rejected on update.
    {Name: "created_by", Type: entity.FieldLink, LinkTarget: "user", Immutable: true},

    // Field with custom validator.
    {Name: "email", Type: entity.FieldData,
        Validators: []entity.FieldValidator{entity.ValidateEmail}},

    // DateTime field — stored as UTC, rendered in EAT for Kenyan tenants.
    {Name: "submitted_at", Type: entity.FieldDateTime},
},
```

### 2.4.3 Edge declarations — relationships to other entities

Edges declare relationships between entities. An edge declaration generates the foreign key column (on the many side of a one-to-many edge), the appropriate index, and the join method on the EntityRepository's Query method.

```go
Edges: []entity.EdgeDef{
    // One-to-many: invoice has many invoice_lines.
    // CascadeDelete: deleting an invoice also deletes its lines.
    {Name: "lines", Target: "invoice_line", Type: entity.EdgeOneToMany, CascadeDelete: true},

    // Many-to-one: invoice belongs to a customer.
    // Inverse of the customer → invoices edge declared on the customer entity.
    {Name: "customer", Target: "customer", Type: entity.EdgeManyToOne},

    // Many-to-many: invoice can be tagged with multiple tax categories.
    {Name: "tax_categories", Target: "tax_category", Type: entity.EdgeManyToMany},
},
```

### 2.4.4 Hook registrations

Hooks are functions called at specific points in the entity record lifecycle. They are declared as slices of interface implementations on the `entity.HookSet` struct. Hook execution order within each lifecycle stage is the order of declaration — first declared, first executed. See [§7 The EntityRecord Lifecycle] for the complete lifecycle and hook model.

```go
Hooks: entity.HookSet{
    BeforeValidate: []entity.BeforeValidateHook{
        &InvoiceDueDateNormalizer{},   // Normalize due_date to end-of-day EAT.
    },
    BeforeCreate: []entity.BeforeCreateHook{
        &InvoiceCustomerValidator{},   // Verify customer is active.
        &InvoiceTaxCalculator{},       // Compute tax_amount from total_kes.
    },
    AfterCreate: []entity.AfterCreateHook{
        &InvoiceNumberAssigner{},      // Assign naming series number.
        &InvoiceAuditLogger{},         // Write audit log entry.
    },
    OnSubmit: []entity.LifecycleHook{
        &InvoiceSubmitWorkflowTrigger{}, // Start InvoiceSubmissionWorkflow.
    },
    BeforeDelete: []entity.BeforeDeleteHook{
        &InvoiceDeleteGuard{},         // Block deletion of submitted invoices.
    },
},
```

### 2.4.5 Permission policy bindings

Permission policy bindings declare which roles can perform which operations on this entity. The binding is a declarative map from operation to the minimum role required. These bindings are compiled into the Casbin enforcer at registration time.

```go
Permissions: entity.PermissionSet{
    Create: []string{"role:finance.accounts_payable", "role:tenant.admin"},
    Read:   []string{"role:finance.viewer", "role:finance.accounts_payable", "role:tenant.admin"},
    Write:  []string{"role:finance.accounts_payable", "role:tenant.admin"},
    Delete: []string{"role:tenant.admin"},
    Submit: []string{"role:finance.accounts_payable", "role:tenant.admin"},
    Cancel: []string{"role:finance.manager", "role:tenant.admin"},
},
```

### 2.4.6 Workflow trigger bindings

Workflow trigger bindings declare which Temporal workflows to start in response to entity lifecycle events. The binding specifies the workflow function name, the task queue, and a function to build the workflow input from the entity record and trigger context.

```go
WorkflowTriggers: []entity.WorkflowTrigger{
    {
        On:          entity.EventOnSubmit,
        WorkflowFn:  "InvoiceSubmissionWorkflow",
        TaskQueue:   "finance.invoice.submit",
        InputBuilder: func(rec *entity.EntityRecord, ctx entity.TriggerContext) (any, error) {
            return finance.InvoiceSubmissionInput{
                TenantID:  rec.TenantID,
                InvoiceID: rec.ID,
                ActorID:   ctx.ActorID,
            }, nil
        },
    },
},
```

### 2.4.7 UI page builder bindings

The framework generates a default amis page schema set for every EntityDefinition. If the default is sufficient, no page builder binding is needed. If you need a custom layout for any view, you register a page builder function that returns the amis JSON for that view.

```go
PageBuilders: entity.PageBuilderSet{
    // Override only the detail view; use defaults for list, create, edit.
    Detail: finance.BuildInvoiceDetailPage,
},
```

### 2.4.8 Naming series configuration

Naming series generate human-readable document numbers with structured formats and per-series atomic counters. A naming series declaration specifies the format string using tokens that are evaluated at creation time.

```go
// NamingSeries declares the format for auto-generated document numbers.
// {SEQ:5} = a zero-padded 5-digit sequence number, reset annually.
NamingSeries: &entity.NamingSeriesDef{
    Field:  "number",
    Format: "INV-{YYYY}-{SEQ:5}",
    ResetOn: entity.ResetAnnually,
    // Allow tenant administrators to override the prefix.
    TenantOverridable: true,
},
```

---

## 2.5 The EntityRegistry

### 2.5.1 Global registry — what it holds and when it is populated

The EntityRegistry is a process-global data structure that holds all registered EntityDefinitions and the derived artifacts compiled from them. It is populated in two phases: at process startup for system entity definitions, and at tenant-boot time for each tenant's custom entity definitions. The registry is read-only during normal request processing; writes happen only during startup and tenant boot.

The registry holds: the raw EntityDefinition structs, the compiled Casbin policy rules for each entity, the generated Fiber route handlers registered on the app, the amis default page schema cache keyed by entity name and view type, the hook chain slices for each lifecycle stage, and the EntityResolver dispatch table.

### 2.5.2 System entity registration at compile time

System entities are registered by module `init()` functions or by explicit registration calls in the application's startup sequence. The registration function validates the EntityDefinition, compiles the permission policies, generates the route handlers, and registers the dispatch entry in the EntityResolver. All of this happens before the Fiber HTTP server starts accepting requests.

```go
// Register is called from the finance module's init function.
// All system entity registrations happen before main() starts the HTTP server.
func Register(registry *entity.Registry) error {
    if err := registry.RegisterSystem(InvoiceDefinition); err != nil {
        return fmt.Errorf("finance: register invoice: %w", err)
    }
    if err := registry.RegisterSystem(InvoiceLineDefinition); err != nil {
        return fmt.Errorf("finance: register invoice_line: %w", err)
    }
    if err := registry.RegisterSystem(JournalEntryDefinition); err != nil {
        return fmt.Errorf("finance: register journal_entry: %w", err)
    }
    return nil
}
```

### 2.5.3 Custom entity loading at tenant boot

When the first request arrives for a tenant whose custom entities have not yet been loaded into the registry, the tenant boot sequence fires. This sequence connects to the tenant's PostgreSQL schema, queries the `entity_definitions` and `entity_field_definitions` tables, constructs `CustomDefinition` structs from the results, and calls `registry.RegisterCustomForTenant(tenantID, def)` for each definition.

The boot sequence runs once per process lifetime per tenant. After the initial boot, custom entity definitions for that tenant are available in the process-local in-memory registry and served from there without any additional database queries. Custom entity schema changes made through the admin API invalidate the in-memory registration and trigger a re-boot on the next request.

### 2.5.4 Registry lookup — by name, by tenant, by type

The EntityRegistry provides several lookup methods used by the EntityResolver and other framework components:

```go
// LookupSystem returns the system entity definition for the given name.
// Returns (nil, false) if no system entity with that name is registered.
def, ok := registry.LookupSystem("invoice")

// LookupCustom returns the custom entity definition for the given tenant and name.
// Returns (nil, false) if the tenant has no custom entity with that name.
def, ok := registry.LookupCustom(tenantID, "vehicle_inspection")

// ListSystemEntities returns all registered system entity definitions.
// Used by migration tooling and the admin interface.
defs := registry.ListSystemEntities()

// ListCustomEntities returns all custom entity definitions registered for a tenant.
defs := registry.ListCustomEntities(tenantID)
```

### 2.5.5 Registry concurrency model — reads vs writes during tenant boot

The EntityRegistry is designed for a read-heavy access pattern during normal operation: every incoming request performs one or more registry lookups to resolve entity types. Writes to the registry happen only during process startup (system entity registration) and tenant boot (custom entity loading).

The registry uses a sync.RWMutex for safe concurrent access: read operations acquire a read lock (allowing concurrent reads), and write operations acquire the full write lock (blocking all reads during the write). Tenant boot writes are expected to be rare relative to request read operations, so the write-lock contention is acceptable.

The per-tenant sub-registries use a separate sync.RWMutex per tenant. This means a tenant boot operation for tenant A does not block registry reads for tenant B. The global registry mutex is held only for the brief period during which a new tenant's sub-registry is inserted into the global map, not for the entire duration of the tenant boot sequence.

> **Important:** Do not call `registry.RegisterCustomForTenant` from within a request handler. Tenant boot is an infrastructure-level operation managed by the tenant resolution middleware. Calling registration from handler code races with concurrent requests reading from the registry.
