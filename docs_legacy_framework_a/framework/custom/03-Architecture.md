> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

### Chapter 3 — Architecture Overview

#### 3.1. The Five-Layer Model

Awo is organised into five layers with a strict dependency rule: each layer depends only on the layer below it and on the framework's shared type package. No layer reaches upward. A violation of this rule — a persistence layer function that imports a Fiber handler, for example — is treated as a build-time error enforced by Go's package visibility model. Understanding this layering is the fastest way to understand where to place new code.

##### 3.1.1. UI layer — amis SDUI JSON served by the API

The UI layer is not a running process; it is a contract. The Awo server produces amis-compatible JSON page definitions at the `/api/v1/pages/{page-name}` endpoints. The amis client — a JavaScript runtime loaded in the browser — renders those definitions into forms, tables, dashboards, and wizards. The server owns the complete page specification; the client owns only the rendering.

This layer has no Go code in the traditional sense. Page builder functions (§21.3) produce Go structs that serialise to amis JSON; the framework handles serialisation and caching. The UI layer's "code" is the set of page builder functions registered against each `EntityDefinition`. Custom renderers (§25) are the one exception: they are JavaScript bundles served as static assets and referenced by type string in the amis JSON.

##### 3.1.2. API layer — Fiber HTTP server, middleware pipeline, route handlers

The API layer is the Fiber HTTP server. It receives HTTP requests, runs them through the middleware pipeline, dispatches to route handlers, and returns HTTP responses. Route handlers in this layer are thin: they extract validated input, call the domain layer, and serialise the result. They do not contain business logic. A route handler that performs a database query directly, without going through an `EntityRepository`, is a framework violation.

The middleware pipeline — documented fully in §13 — is responsible for cross-cutting concerns: request ID assignment, structured logging, tenant resolution, session validation, rate limiting, and CORS. Middleware runs before any route handler. The order is fixed and not configurable by module developers; the canonical order is defined in §13.1.

##### 3.1.3. Domain layer — EntityDefinition system, hooks, validators, permission policies

The domain layer is where business logic lives. It contains the `EntityDefinition` descriptors, the hook implementations (`before_save`, `after_save`, `on_submit`, `on_cancel`, `before_delete`), the field validators, and the privacy policies. This layer is the primary workbench for module developers.

The domain layer does not import Fiber. It does not know about HTTP status codes, request bodies, or response envelopes. It receives typed inputs via the hook context and returns typed errors that the API layer translates into HTTP responses. This separation makes domain logic fully testable without spinning up an HTTP server. A `before_save` hook can be unit-tested by constructing a `hook.Context` with mock dependencies and calling the function directly.

##### 3.1.4. Workflow layer — Temporal workflows, activities, sagas

The workflow layer contains Temporal workflow functions, activity functions, and saga definitions. It handles multi-step business processes that must survive process crashes, span long time periods, or require compensating transactions. This layer is the only place where cross-entity, multi-step mutations should be orchestrated.

The workflow layer communicates with the domain layer exclusively through the `EntityRepository` interface — activities call `repo.Get`, `repo.Update`, and similar methods, and they receive typed `EntityRecord` values in return. The workflow layer does not import Fiber and does not construct HTTP requests to the API layer to make changes; it calls the repository directly. Full workflow documentation is in Part V.

##### 3.1.5. Store layer — EntityRepository interface, current ent implementation, PostgreSQL

The store layer is the `EntityRepository` interface and its implementations. It translates `Filter` structs, `EntityRecord` mutations, and transaction scopes into SQL operations executed against PostgreSQL. The ent reference implementation lives in `internal/store/ent` and is the only package in the codebase that imports `entgo.io/ent`. The store layer knows nothing about HTTP, Temporal, or amis.

The store layer always operates within a tenant schema context. Every database connection used by the store layer has its PostgreSQL search path set to `t_{tenant_slug}` before any query executes. This is enforced by the connection pool management described in §14.4, not by individual queries. A query that forgets a `WHERE tenant_id = ?` clause is still safe because it can only see rows in the current tenant's schema.

---

#### 3.2. Request Lifecycle — Read Path

Understanding the read path in full detail is important because every layer adds context that later layers depend on. A missing step in this chain — a middleware that fails to propagate tenant context, a handler that skips the permission check — is a security vulnerability, not just a bug.

##### 3.2.1. HTTP request arrives at Fiber

Fiber receives the raw HTTP request on its listener socket and constructs a `*fiber.Ctx`. At this point, nothing about the request has been validated. The Fiber context carries the raw headers, query parameters, path parameters, and body. The middleware pipeline has not run. No tenant, session, or user information is attached yet.

##### 3.2.2. Middleware pipeline — request ID, tenant resolution, session validation, rate limiting

The request passes through the middleware pipeline in the fixed order defined in §13.1. Request ID middleware assigns a UUID to the request and attaches it to the Fiber context and to the `context.Context` used for downstream calls. Structured logging middleware starts a per-request log record with the request ID.

Tenant resolution middleware parses the subdomain from the `Host` header (`{tenant}.awo.app`), falls back to the `X-Awo-Tenant` header if the subdomain pattern does not match, looks up the tenant record, and attaches the `TenantContext` to the request context. If tenant resolution fails — the subdomain is unknown, or the tenant is suspended — the middleware returns the appropriate HTTP error response and the pipeline halts. Session validation middleware then reads the session cookie, looks up the session in Redis, validates the session has not expired, and attaches the user's identity and resolved permissions to the request context. Rate limiting middleware enforces per-tenant and per-user limits using a Redis sliding window counter.

##### 3.2.3. Route handler resolves EntityDefinition from URL

The route handler extracts the entity name from the URL path parameter. For `/api/v1/sales-invoices/inv-001`, the entity name is `sales-invoice`, which the framework denormalises to `SalesInvoice` using the kebab-to-Pascal mapping registered at startup. The handler then calls `entity.Resolve(ctx, tenantCtx, "SalesInvoice")` to obtain the `EntityRepository` for this entity in this tenant's context.

```go
// Example: Route handler resolving entity and fetching a record
func GetEntityRecordHandler(c *fiber.Ctx) error {
    tc, err := tenant.FromContext(c.UserContext())
    if err != nil {
        return err
    }
    entityName := entity.KebabToPascal(c.Params("entity"))
    repo, err := entity.Resolve(c.UserContext(), tc, entityName)
    if err != nil {
        return entity.NewNotFoundError(entityName)
    }
    record, err := repo.Get(c.UserContext(), c.Params("id"))
    if err != nil {
        return err
    }
    return c.JSON(entity.SuccessResponse(record))
}
```

##### 3.2.4. EntityResolver selects execution path — system or custom

The `EntityResolver` performs the two-stage registry lookup described in §2.2.3. For a system entity, it returns the ent-backed `EntityRepository` instance configured with the tenant's schema connection. For a custom entity, it returns the JSONB engine wrapped behind the same interface, loaded with the tenant's `CustomFieldDef` metadata. The route handler does not know or care which path was taken.

##### 3.2.5. Permission policy evaluated against resolved tenant + user context

Before the handler calls `repo.Get`, the framework's permission middleware has already verified that the user's role includes the `read` permission for this entity (§16.3.1). If the role check passes, the `EntityRepository` implementation evaluates the entity's registered privacy policies against the current user and tenant context (§9). For a `TenantIsolation` policy, this adds an implicit filter that restricts results to the current tenant's schema — redundant with the schema-level isolation, but explicit. For an `OwnerOnly` policy, this adds an additional predicate that restricts results to records where `assigned_to` matches the current user's ID.

##### 3.2.6. EntityRepository.Query() called — implementation dispatches to ent or JSONB engine

For a `Get` call, the repository retrieves the single record by primary key. For a `Query` call, the repository translates the `Filter` struct into SQL (for system entities) or a JSONB path expression (for custom entities), applies any additional privacy policy predicates, executes the query against the tenant's PostgreSQL schema, and returns a slice of `EntityRecord` values along with pagination metadata.

##### 3.2.7. Privacy policy applied to result set

Privacy policies that operate at the row level — `OwnerOnly`, `DepartmentScope`, `RoleFilter` — are applied as additional WHERE clause predicates that the `EntityRepository` implementation injects before executing the query. Privacy policies that operate at the field level — masking a `Sensitive` field for roles that should not see it — are applied after the query returns, by iterating over the `EntityRecord` values and nulling out or redacting the restricted fields.

##### 3.2.8. Response serialised and returned

The route handler receives the `EntityRecord` or slice of `EntityRecord` values, wraps them in the standard response envelope (§17.2), and serialises to JSON. The response envelope includes pagination metadata for list responses. The HTTP status is 200. The request ID is included in the `X-Request-ID` response header.

---

#### 3.3. Request Lifecycle — Write Path with Workflow

The write path carries the same middleware pipeline as the read path and adds a transaction boundary, hook invocations, and an optional Temporal workflow trigger. The transaction boundary is the most important detail: it determines what is atomic and what is not.

##### 3.3.1. HTTP request arrives, middleware pipeline runs

Identical to the read path. The middleware pipeline runs in the same fixed order. By the time the route handler executes, the request has a tenant context, a validated session, a rate-limit check, and a structured log record.

##### 3.3.2. Route handler extracts payload, validates against EntityDefinition field schema

The route handler reads the request body as JSON, unmarshals it into a raw map, and passes it to the framework's field validator. The validator iterates over the `EntityDefinition`'s declared fields, checks that required fields are present, checks that each value matches its declared type and constraints (`MaxLen`, `Min`, `Max`, regex patterns), and runs any registered field validators. Validation errors are collected and returned as a 422 response before any hook or database call is made.

```go
// Example: Write handler with payload extraction and field validation
func CreateEntityHandler(c *fiber.Ctx) error {
    tc, err := tenant.FromContext(c.UserContext())
    if err != nil {
        return err
    }
    entityName := entity.KebabToPascal(c.Params("entity"))
    repo, err := entity.Resolve(c.UserContext(), tc, entityName)
    if err != nil {
        return entity.NewNotFoundError(entityName)
    }
    var payload map[string]any
    if err := c.BodyParser(&payload); err != nil {
        return entity.NewValidationError("body", "invalid JSON")
    }
    record, err := repo.Create(c.UserContext(), payload)
    if err != nil {
        return err
    }
    return c.Status(fiber.StatusCreated).JSON(entity.SuccessResponse(record))
}
```

##### 3.3.3. EntityRecord assembled from validated input

After field validation passes, the framework assembles an `EntityRecord` from the validated input map. This includes setting default values for fields not provided by the caller (static defaults or Go function defaults declared in the `EntityDefinition`), assigning a naming series value if the entity has one configured (§5.4), and setting system-managed fields like `created_at`, `created_by`, and `tenant_id`.

##### 3.3.4. before_save hooks invoked

The framework invokes all registered `before_save` hooks in declaration order (§7.3). Each hook receives a `hook.Context` carrying the assembled `EntityRecord`, the tenant context, the user context, and the `EntityRepository`. A hook may modify the `EntityRecord` (normalise values, compute derived fields), perform additional validation queries against the repository, or return an error to abort the operation. If any hook returns a non-nil error, the framework halts the pipeline, does not open a database transaction, and returns the error to the caller.

##### 3.3.5. EntityRepository.Mutate() called — ent transaction opened

With `before_save` hooks satisfied, the repository opens a PostgreSQL transaction scoped to the tenant's schema and executes the INSERT (for creates) or UPDATE (for updates) SQL. The transaction is held open; it does not commit yet.

##### 3.3.6. after_save hooks invoked inside transaction

The framework invokes all registered `after_save` hooks while the transaction is still open (§7.4). This means that work done inside an `after_save` hook — inserting related records, updating a running total on a parent entity, writing an audit log entry — is part of the same transaction as the primary record's INSERT or UPDATE. If an `after_save` hook returns an error, the transaction is rolled back and the primary record is not persisted.

> **Warning:** Do not perform slow operations inside `after_save` hooks. External API calls, email sending, and anything with unpredictable latency must not happen here. Use `after_save` only for synchronous, fast, transactional side effects. Everything else belongs in a Temporal activity triggered after the transaction commits.

##### 3.3.7. Temporal workflow triggered via signal or start — carries EntityRecord ID

If the `EntityDefinition` has a workflow trigger binding for `after_save` or `on_submit`, the framework calls `temporal.Client.ExecuteWorkflow` after the `after_save` hooks complete but before the transaction commits. The workflow start call is made synchronously; the call returns a workflow run handle immediately without waiting for the workflow to complete.

The workflow receives only the entity's primary key and the tenant context as its input. It does not receive the full `EntityRecord` because workflow inputs must be serialisable and stable; instead, the first activity in the workflow calls `repo.Get` to fetch the current state of the record.

> **Note:** The Temporal workflow start call is inside the transaction in the sense that it happens before commit, but the Temporal server does not participate in the PostgreSQL transaction. If the PostgreSQL transaction commits and the Temporal workflow start fails, the workflow will not run. This edge case is mitigated by the idempotency of workflow IDs: a retry of the HTTP request will attempt to start the workflow again with the same ID, which is safe due to the `RejectDuplicate` policy. For truly critical workflows, consider the transactional outbox pattern described in the notification module documentation.

##### 3.3.8. Transaction committed — workflow runs asynchronously

The PostgreSQL transaction commits. The `EntityRecord` is now durably persisted. The Temporal workflow begins executing asynchronously in the background. The HTTP response is returned without waiting for the workflow to complete.

##### 3.3.9. HTTP response returned — workflow outcome delivered via notification or polling

The response to the caller is the created or updated `EntityRecord` with HTTP 201 (create) or 200 (update). The caller does not receive the workflow's result in this response. Workflow outcomes are delivered asynchronously through the notification system (a server-sent events endpoint backed by Redis pub/sub) or through polling the entity record for status field changes written by the workflow's activities.

---

#### 3.4. Multi-Tenant Architecture

Every aspect of the system's runtime behaviour is conditioned on the resolved tenant. This section describes the structural mechanisms that enforce tenant isolation across the stack.

##### 3.4.1. Tenant identification — subdomain parsing, X-Awo-Tenant header fallback

The canonical tenant identification mechanism is subdomain parsing: a request to `acme.awo.app` identifies tenant `acme`. The `Host` header is parsed by the tenant resolution middleware, and the subdomain component is looked up in the `tenants` system table (which lives in the `public` schema, outside all tenant schemas). If the `Host` header does not contain a recognised subdomain pattern — for example, a raw IP address or a localhost request — the middleware falls back to the `X-Awo-Tenant` header. This fallback is available in all environments and is the standard mechanism for mobile API clients and automated integration tests.

A third fallback using a `?_tenant=` query parameter is available in development mode only and is disabled by the feature flag system in staging and production. Allowing tenant identification via query parameter in production would make it trivial to forge cross-tenant requests.

##### 3.4.2. Per-tenant EntityRegistry — custom entities loaded at boot

Each tenant has its own slice of the `EntityRegistry` (§2.5) populated at tenant boot with that tenant's `CustomFieldDef` records. This per-tenant slice is a superset of the global system entity registry; a lookup for `SalesOrder` in a tenant's context finds the system entity, and a lookup for a tenant-specific custom entity finds the custom entity definition.

The per-tenant registry slice is loaded lazily on the first request for a given tenant after a server restart. Subsequent requests for the same tenant use the cached registry, protected by a per-tenant `sync.RWMutex`. Registry invalidation is triggered when a tenant creates or modifies a `CustomFieldDef` record.

##### 3.4.3. Per-tenant database schema — schema-per-tenant in PostgreSQL

Each tenant's data lives in a dedicated PostgreSQL schema named `t_{tenant_slug}` (e.g., `t_acme`, `t_shellmaanzoni`). All tenant business data — every row in every entity table — lives inside this schema. The only shared schema is `public`, which holds the `tenants` system table and any global reference data. There are no cross-tenant foreign keys; schema isolation is absolute at the database level.

This model provides defence in depth: even if the application-level tenant context is somehow compromised, a connection operating with its search path set to `t_acme` physically cannot see rows in `t_other_tenant` without an explicit schema-qualified query. The migration system (§11) applies migrations to each tenant schema independently, which adds complexity to the deployment pipeline but preserves strict isolation.

##### 3.4.4. Per-tenant connection pool — pgx pool per schema

Each tenant has a dedicated pgx connection pool configured with its schema's search path. Connections in the pool for tenant `acme` have `SET search_path TO t_acme` applied on acquisition. This means no query in the pool can accidentally operate against the wrong tenant's schema through a missing WHERE clause or a forgotten JOIN condition.

Pool sizing is configurable per tenant via `TenantConfig` (§39.1). A high-volume tenant can be allocated a larger pool ceiling than a trial tenant. There is a global ceiling on total connections to prevent a single tenant's traffic spike from exhausting the PostgreSQL `max_connections` limit for all other tenants. The global ceiling is enforced by a PgBouncer instance in front of PostgreSQL (§45.2).

##### 3.4.5. Per-tenant feature flags — Redis-backed, per-tenant overrides

Feature flags are stored in Redis with a key schema that includes the tenant slug (§41.2). The system default for each flag is defined in the flag definition entity; tenants can override individual flags via the `TenantConfig` admin UI. Flags are loaded into the tenant context during tenant resolution and cached for the duration of the request; they do not change mid-request.

Feature flags control module enablement, beta feature access, percentage rollouts, and UI-layer conditional blocks in page builder JSON. A module that is disabled for a tenant will have its API routes return 404 and its menu items hidden by the page builder. The flag system is documented in full in §20.

##### 3.4.6. Per-tenant configuration — TenantConfig entity, inheritance from system defaults

The `TenantConfig` entity holds all tenant-level configuration values: timezone, date format, base currency, KRA eTIMS credentials, NEMA compliance thresholds, module enablement flags, and custom permission overrides. Configuration values are loaded into the `TenantContext` at tenant boot and are available to all framework layers via `tc.Config().Get("key")` typed getter functions — no string key lookups at runtime.

System defaults for every configuration key are defined in the framework's own configuration package. A tenant that has not explicitly set a configuration value inherits the system default. The Kenya-specific defaults — EAT timezone, DD/MM/YYYY date format, KES base currency — are the system defaults and require no per-tenant configuration for Kenyan deployments.

---

#### 3.5. Component Dependency Map

Understanding the dependency graph prevents incorrect startup ordering, clarifies which failure modes cascade and which are isolated, and guides where to focus reliability investment.

##### 3.5.1. Startup order — config → database → Redis → EntityRegistry → Fiber → Temporal worker

The startup sequence is strictly ordered. Configuration is loaded and validated first; a missing required environment variable causes an immediate fatal exit before any external connection is attempted. The PostgreSQL connection pool is established next; if the database is unreachable, startup fails immediately rather than starting a server that cannot serve any request correctly.

Redis is established after the database; Redis failure during startup is also fatal because session validation and feature flag loading both require Redis. The `EntityRegistry` is populated next, which involves loading `CustomFieldDef` records for any pre-warmed tenants and calling each module's `Register` function to register system entities. Only after the registry is fully populated does the Fiber HTTP server begin accepting connections. The Temporal worker is registered last; Temporal connectivity failure during startup is non-fatal and causes the worker to enter a retry loop (§3.5.3).

```
config
  └── database
        └── Redis
              └── EntityRegistry
                    ├── Fiber (accepts traffic)
                    └── Temporal worker (starts retry loop if unreachable)
```

##### 3.5.2. What depends on what — dependency graph for contributors

The Fiber HTTP server depends on the `EntityRegistry` being fully populated. The `EntityRegistry` depends on the database (to load `CustomFieldDef` records). Route handlers depend on the `EntityResolver`, which depends on the registry. The Temporal worker depends on the Temporal server being reachable but not on the HTTP server; the two can start and stop independently.

The `EntityRepository` implementations depend on the database connection pool but not on Redis, Temporal, or Fiber. Hooks depend on the `EntityRepository` interface only. Page builders depend on the `EntityRegistry` (to introspect field lists) and on the permission system (to adjust the emitted JSON based on the user's role). No component in the domain layer or store layer imports Fiber or Temporal.

##### 3.5.3. What can fail independently without cascading — Redis, Temporal

Redis failure after startup does not bring down the process. The session store, feature flag cache, and page definition cache all have defined fallback behaviours (§41.5): session store failure causes authentication to fail gracefully with a 401, feature flag cache failure causes the flag evaluation engine to use system defaults, and page definition cache failure causes the page builder to run on every request (a performance degradation, not a correctness failure).

Temporal worker failure — loss of connectivity to the Temporal server — means that no new workflows are started and no pending workflow tasks are executed. HTTP read and write operations continue to function normally. Mutations that trigger workflow starts will succeed at the persistence layer; the workflow starts are either queued (if the Temporal client uses a durable start queue) or logged as failed events for manual retry. The API does not return an error to the caller in this case by default, though specific high-criticality entities can be configured to fail the HTTP request if the workflow start fails.

##### 3.5.4. What failing brings the process down — database, EntityRegistry

Database failure after startup is the only failure that makes the process unserviceable. Without database connectivity, no `EntityRepository` operation can succeed, no `CustomFieldDef` records can be loaded, and no session validation can be performed (session data is validated against the database for user existence even when the session itself is in Redis). The process does not exit on database failure but returns 503 Service Unavailable for all requests that require database access.

`EntityRegistry` population failure during startup is fatal. If a module's `Register` call returns an error — because a system entity name conflicts with another, because a field declaration is invalid, or because a dependency entity is not yet registered — the process exits with a descriptive error. This is intentional: starting a server with a partially populated registry would lead to unpredictable runtime behaviour that is far harder to diagnose than a clean startup failure.

> **Danger:** Never recover from `EntityRegistry` population failures using `recover()` or by ignoring errors returned from `app.RegisterEntity`. A partially registered entity is worse than no entity: some API routes will work and others will not, privacy policies may not be applied to the registered paths, and the problem will manifest as a data access bug rather than a startup error.

---

#### Chapter summary

Chapter 3 defines the five-layer architecture (§3.1), the read and write request lifecycles (§3.2 and §3.3), the multi-tenant runtime model (§3.4), and the component dependency graph (§3.5). The most important concepts for day-to-day development are the write path transaction boundary (§3.3.5–3.3.8, which determines what is atomic and what is eventually consistent) and the per-tenant schema isolation model (§3.4.3, which is the structural foundation of all tenant safety guarantees). The startup order (§3.5.1) is critical reading before writing any module bootstrap code.

**Next chapters to read:**

- §13 — Middleware Pipeline (expands §3.2.2 into the full middleware reference, including the canonical execution order and each middleware's behaviour)
- §14 — Multi-Tenancy Middleware (expands §3.4 into the implementation details of tenant resolution, context propagation, and per-tenant connection management)
- §8 — The Persistence Interface (expands §3.1.5 into the full `EntityRepository` contract, the Filter DSL, and the ent reference implementation details)
