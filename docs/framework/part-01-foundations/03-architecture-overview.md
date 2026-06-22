# Architecture Overview

## 3.1 The Five-Layer Model

### 3.1.1 UI layer — amis SDUI JSON served by the API

The UI layer in Awo is not a separately deployed frontend application. It is a set of JSON documents — amis page schemas — served by the same API process that handles data operations. The browser loads the amis SDK (a JavaScript bundle) once from a CDN or static file server, and after that, every page transition is a JSON fetch to `/api/v1/pages/{page-name}`. The amis SDK interprets the JSON and renders the appropriate UI components without any additional JavaScript from the application developer.

This architecture means that the "frontend" is a server-side concern: the JSON schemas are generated in Go, cached in Redis, and served with the same authentication and permission context that governs data API calls. A user who cannot read invoice records also cannot load the invoice list page schema — the permission check happens at schema-serve time, not just at data-fetch time. The UI layer is the only layer in the stack that the user's browser directly interacts with; all other layers are server-side.

The UI layer's dependency is entirely downward: it reads from the domain layer's EntityDefinition declarations to generate its default page schemas. It does not write to the domain layer, does not have side effects on persistence, and does not participate in workflow execution. This makes it safe to disable the page builder cache and regenerate schemas on every request during development without affecting correctness.

### 3.1.2 API layer — Fiber HTTP server, middleware pipeline, route handlers

The API layer is a Fiber v2 HTTP server that implements the request ingestion, middleware pipeline, route dispatch, and response serialization responsibilities. It is the only layer that touches HTTP — all other layers receive and return typed Go values, not HTTP request/response objects.

The middleware pipeline runs before any route handler and handles: request ID assignment, structured logging setup, tenant resolution, session validation, and rate limiting. The pipeline runs in a fixed order that is not configurable at runtime; altering middleware order requires a code change and redeployment. This constraint is intentional: middleware order is a security-sensitive concern that should not be changed by configuration.

Route handlers in the API layer are thin. A typical entity CRUD handler: extracts the entity type from the URL, calls `resolver.Resolve(ctx, entityType)` to obtain an EntityRepository, calls the appropriate repository method with the validated input, and serializes the response. The handler does not contain business logic, does not call external services, and does not make decisions about what the data means. If a handler is longer than approximately 50 lines, business logic has leaked into the wrong layer.

### 3.1.3 Domain layer — EntityDefinition system, hooks, validators, permission policies

The domain layer contains the framework's core abstractions: the EntityDefinition system, the hook chain executor, the field validator pipeline, and the permission policy bindings. This is where business rules live: what fields an invoice must have, what validation checks run before an invoice is saved, what permission a user must hold to submit an invoice, and what events fire after an invoice is saved.

The domain layer is stateless from the perspective of an individual request: it does not hold mutable state between requests, and its inputs (entity records, user context, hook implementations) fully determine its outputs (validation errors, modified field values, permission decisions). This makes the domain layer the easiest layer to test: unit tests can instantiate domain objects, feed them inputs, and assert on outputs without any HTTP infrastructure or database connection.

Module developers spend most of their time in the domain layer: declaring EntityDefinitions, implementing hook interfaces, writing field validators, and declaring permission policy sets. The framework provides the runtime that executes these declarations; module developers provide the declarations.

### 3.1.4 Workflow layer — Temporal workflows, activities, sagas

The workflow layer is a set of Temporal workflow functions and activity functions that execute multi-step business processes outside the synchronous HTTP request-response cycle. The workflow layer is the right place for any operation that: spans multiple HTTP requests, requires human approval, calls external APIs with retry semantics, or must execute atomically across multiple database operations where a DB-level transaction cannot be held open.

The workflow layer and the API layer interact in one direction at a time: the API layer can start a workflow or send a signal to a running workflow, but a running workflow cannot call the API layer's HTTP endpoints. Instead, workflows interact with the persistence and external systems directly through activity functions that accept injected dependencies (EntityRepository, HTTP clients, Redis client).

The workflow layer has its own failure model: Temporal provides durability through event-sourced workflow history stored in Temporal's own database. A workflow that is mid-execution when the Awo process crashes will resume from its last checkpoint when the process restarts. This durability guarantee is why the workflow layer is the correct home for KRA eTIMS submission: if the Awo process crashes between posting the ledger entry and submitting to eTIMS, Temporal will resume and complete the eTIMS submission after restart.

### 3.1.5 Store layer — EntityRepository interface, PostgreSQL

The store layer is the persistence boundary: the `EntityRepository` interface, its implementation(s), and the PostgreSQL database. The store layer accepts typed Go values from the domain and workflow layers and persists them to PostgreSQL, returns them on reads, and maintains the transaction boundaries that the domain layer requests via `WithTx`.

The store layer is the only layer in the system that knows about PostgreSQL, schema names, SQL query shapes, and connection pool management. All other layers interact with the store layer exclusively through the `EntityRepository` interface. This boundary ensures that the store layer's implementation can be changed — new query generation strategy, different connection pool library, or even a different database for a specific entity type — without affecting any code in the domain, workflow, or API layers.

---

## 3.2 Request Lifecycle — Read Path

### 3.2.1 HTTP request arrives at Fiber

A read request — for example, `GET /api/v1/entities/invoice/INV-2024-00042` — arrives at the Fiber HTTP server. Fiber's router matches the path against registered routes and identifies the route handler responsible for single-entity reads on the `invoice` entity type. The route was registered at startup by the EntityDefinition registration pipeline when the `invoice` system entity was registered.

At this point, the request has been accepted and assigned to a Fiber worker goroutine. No business logic has executed. The Fiber context `c *fiber.Ctx` carries the raw HTTP request data — headers, path parameters, query string — but nothing about the tenant, the authenticated user, or any entity data.

### 3.2.2 Middleware pipeline — request ID, tenant resolution, session validation, rate limiting

The middleware pipeline executes in a fixed sequence before the route handler. The canonical order is:

1. **Request ID** — assigns a UUID request ID from `X-Request-ID` header if present, generates a new one otherwise. The request ID is set in the response header and attached to the Fiber context.
2. **Structured logging** — attaches the request ID, HTTP method, and path to the logger context so all downstream log entries carry these fields automatically.
3. **Panic recovery** — wraps the remaining pipeline in a recover() handler so unexpected panics return a 500 response rather than crashing the goroutine.
4. **CORS** — validates the `Origin` header against the allowed origin list for the tenant subdomain.
5. **Tenant resolution** — extracts the tenant identifier from `X-Tenant-ID` header (first priority), `tenant_id` query parameter (second priority, development only), or subdomain parsing (fallback). Validates the tenant exists and is active. Stores the resolved `*tenant.Tenant` in the request context and calls `store.SetTenantContextFromCtx(ctx)`, which executes `SELECT set_tenant_context($1)` to set the transaction-local `app.current_tenant_id` PostgreSQL variable that RLS policies read.
6. **Session validation** — extracts the session cookie, looks up the session in Redis, validates expiry, and stores the resolved user identity in the Fiber context.
7. **Rate limiting** — applies per-tenant and per-user rate limits using a Redis sliding window counter.

After the middleware pipeline completes, the Fiber context carries: a request ID, a resolved tenant UUID, a validated user identity, and a database connection scoped to the tenant's PostgreSQL schema.

### 3.2.3 Route handler resolves EntityDefinition from URL

The route handler extracts the entity type from the URL path parameter (`:entity_type`) and the record ID from the path parameter (`:id`). It calls `resolver.Resolve(ctx, entityType)` to obtain an `EntityRepository` scoped to the entity type and tenant.

If the entity type is not registered (neither as a system entity nor as a custom entity for this tenant), the resolver returns `ErrEntityNotFound` and the handler returns `404 Not Found`. This is the same response regardless of whether the entity type exists as a system entity in other tenants — the framework never reveals whether an entity type exists in another tenant's scope.

### 3.2.4 EntityResolver selects execution path — system or custom

The EntityResolver checks whether `invoice` is a system entity (it is — registered at startup by the Finance module). It returns a system-entity-scoped `EntityRepository` implementation that uses generated SQL queries against the `invoice` table in the tenant's PostgreSQL schema. If `invoice` were a custom entity, the resolver would return a JSONB-scoped `EntityRepository` that queries the `custom_entity_records` table with a JSONB path filter on `entity_type = 'invoice'`.

The route handler receives an `EntityRepository` in both cases and proceeds identically.

### 3.2.5 Permission policy evaluated against resolved tenant + user context

Before any database query executes, the permission check runs. The route handler calls the framework's built-in `RequirePermission(ctx, entityType, "read")` helper, which evaluates the Casbin policy for the resolved user's roles against the `invoice.read` policy in the tenant's domain.

```go
// Permission check in route handler — runs before any EntityRepository call.
func (h *EntityHandler) GetOne(c *fiber.Ctx) error {
    entityType := c.Params("entity_type")
    id := c.Params("id")

    // Resolve user identity from session context.
    actor, err := session.ActorFromContext(c.Context())
    if err != nil {
        return c.Status(fiber.StatusUnauthorized).JSON(h.unauthorized())
    }

    // Permission check: does this actor have read permission on this entity?
    allowed, err := h.authz.Enforce(c.Context(), iam.Request{
        Subject: actor.Subject,
        Domain:  actor.Domain,
        Object:  entityType,
        Action:  "read",
    })
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(h.internalError(err))
    }
    if !allowed {
        return c.Status(fiber.StatusForbidden).JSON(h.forbidden(entityType, "read"))
    }

    // Permission check passed. Proceed to EntityRepository.
    repo, err := h.resolver.Resolve(c.Context(), entityType)
    if err != nil {
        return c.Status(fiber.StatusNotFound).JSON(h.notFound(entityType))
    }

    rec, err := repo.Get(c.Context(), entityType, id)
    if err != nil {
        return h.handleRepoError(c, err)
    }

    return c.JSON(h.successResponse(rec))
}
```

### 3.2.6 EntityRepository.Query() called — implementation dispatches to SQL or JSONB engine

The route handler calls `repo.Get(ctx, entityType, id)`. The EntityRepository implementation executes a generated SQL query: `SELECT * FROM invoice WHERE id = $1 AND deleted_at IS NULL`. The RLS policy on `invoice` automatically restricts results to the current tenant — no explicit `WHERE tenant_id = ?` is needed in the query. The result is deserialized into an `*EntityRecord` with the field map populated from the typed SQL column values.

Privacy policies registered on the EntityDefinition may add further WHERE predicates on top of the RLS filter — for example, `OwnerOnly` adds `assigned_to = $current_user_id`. The policy predicates run at the database level, not the application level.

### 3.2.7 Privacy policy applied to result set

For read operations that return multiple records (list/query endpoints), privacy policies may filter individual records from the result set after the database query returns. Field-masking privacy policies remove or obscure individual field values from the returned `EntityRecord` based on the current user's role. For example, a `SensitiveFieldMask` policy configured on the `bank_account_number` field returns `"****"` instead of the actual value for any user without the `finance.sensitive` permission.

### 3.2.8 Response serialised and returned

The route handler serializes the `*EntityRecord` into the standard response envelope and returns it. The serialization layer handles EAT timezone conversion for DateTime fields (stored as UTC in PostgreSQL, serialized as EAT-offset ISO 8601 strings in the API response), KES currency formatting for Currency fields, and omission of `Sensitive` fields that were not masked but should not appear in standard responses.

---

## 3.3 Request Lifecycle — Write Path with Workflow

### 3.3.1 HTTP request arrives, middleware pipeline runs

A write request — for example, `POST /api/v1/entities/invoice/INV-2024-00042/submit` — arrives and passes through the same middleware pipeline as the read path: request ID, logging, panic recovery, CORS, tenant resolution, session validation, and rate limiting. After the pipeline, the Fiber context carries a resolved tenant, a validated user identity, and a tenant-scoped database connection.

### 3.3.2 Route handler extracts payload, validates against EntityDefinition field schema

The route handler extracts the JSON request body and validates it against the EntityDefinition's field schema. Field validation runs in two phases: type coercion (string to Currency, string to DateTime, etc.) and then field-level validators (required check, max length, regex, custom validators). If any field fails validation, the handler returns a `422 Unprocessable Entity` response with field-level error messages in the amis error format.

### 3.3.3 EntityRecord assembled from validated input

The route handler assembles an `*EntityRecord` from the validated field map and the framework-managed metadata: the tenant ID from context, the actor ID from the session, the entity type from the URL, and a newly generated record ID (for create operations). For update operations, the framework fetches the existing record from the EntityRepository to populate the previous-state context that `before_save` hooks may need.

### 3.3.4 before_save hooks invoked

The framework's lifecycle manager invokes the `before_save` hook chain (which includes `before_validate`, `before_create`/`before_update`, and `before_save`). Each hook in the chain receives the `*EntityRecord` and may: modify field values, validate business rules, or return an error to abort the operation. If any hook returns a non-nil error, the lifecycle manager stops the chain, rolls back any partial changes, and returns the error to the route handler.

### 3.3.5 EntityRepository.Mutate() called — transaction opened

After all `before_save` hooks complete successfully, the route handler calls the appropriate EntityRepository mutation method (`Create`, `Update`, or a custom action's mutation). The EntityRepository implementation opens a PostgreSQL transaction and executes the SQL mutation within it.

### 3.3.6 after_save hooks invoked inside transaction

With the mutation committed in the transaction (but not yet committed to the database), the framework invokes the `after_save` hook chain. These hooks run inside the open transaction, which means their side effects are part of the same atomic database operation. Hooks that must be part of the same transaction — writing to an audit log table, updating a denormalized counter, inserting a related record — should be registered as `after_save` hooks.

### 3.3.7 Temporal workflow triggered via signal or start

If the EntityDefinition has a workflow trigger binding for the current lifecycle event (in this case, `on_submit`), the `after_save` hook chain includes the framework's built-in `WorkflowTriggerHook`. This hook calls `temporal.StartWorkflow(ctx, options, input)` to start the Temporal workflow for this operation. The workflow start request is sent to Temporal's frontend service as part of the HTTP call from the Awo process — it is not inside the PostgreSQL transaction.

The workflow ID is stored in the entity record before the Temporal start call, so the entity record can be updated to reference the workflow ID within the same transaction if needed. However, the start call itself is outside the transaction. If the transaction commits but the Temporal start call fails (network error, Temporal unavailable), the entity record will have been saved without a running workflow. The `WorkflowTriggerHook` handles this case by recording the failure in a retry queue and scheduling a re-attempt.

### 3.3.8 Transaction committed — workflow runs asynchronously

After all `after_save` hooks complete, the PostgreSQL transaction commits. The entity record is now durably persisted in the tenant's database schema. The Temporal workflow is now running (or queued to run) asynchronously in the Temporal worker process. The HTTP request handler returns immediately with a `202 Accepted` response carrying the entity record ID and the workflow ID.

### 3.3.9 HTTP response returned — workflow outcome delivered via notification or polling

The caller does not wait for the workflow to complete. The `202 Accepted` response tells the caller that the operation was accepted and a workflow is executing. The caller has two options to follow the workflow outcome: poll the entity record's status field (which the workflow updates as it progresses through stages), or subscribe to the workflow's completion notification via the SSE endpoint at `/api/v1/notifications/stream`.

---

## 3.4 Multi-Tenant Architecture

### 3.4.1 Tenant identification — subdomain parsing, X-Tenant-ID header fallback

The tenant identification middleware extracts a tenant identifier from the incoming HTTP request using the following priority order:

1. `X-Tenant-ID` header — the preferred method for API clients and mobile applications. The header value is a tenant UUID string. The middleware validates the UUID format before proceeding.
2. `tenant_id` query parameter — supported for webhook callbacks and legacy integrations. Should be disabled in production environments where tenant impersonation via URL manipulation is a security concern.
3. Subdomain parsing — for browser-based access to the main application. The middleware extracts the tenant identifier from the hostname's subdomain, handling common prefixes (`bo.`, `portal.`, `app.`, `api.`) that precede the tenant identifier component.

After extracting the raw tenant identifier, the middleware validates the tenant exists in the database (with a 5-minute TTL cache to avoid a DB round-trip per request), confirms the tenant status is `Active`, and stores the resolved `*tenant.Tenant` in `c.Locals("entity_id")`.

### 3.4.2 Per-tenant EntityRegistry — custom entities loaded at boot

Each tenant maintains its own sub-registry within the global EntityRegistry. The sub-registry holds the tenant's custom EntityDefinition records loaded from the tenant's database schema during tenant boot. System entity definitions are shared across all tenants from the global registry. The EntityResolver always checks the global system entity registry first, then falls back to the tenant's custom entity sub-registry.

### 3.4.3 Shared schema with Row-Level Security — RLS enforces tenant isolation

All tenants share the same PostgreSQL schema. Every tenant-scoped table has a `tenant_id uuid NOT NULL` column, and every table has `FORCE ROW LEVEL SECURITY` enabled with a policy that restricts rows to the current tenant:

```sql
CREATE POLICY tenant_isolation ON invoice
    USING (tenant_id = current_tenant_id());
```

The `current_tenant_id()` function reads the transaction-local setting `app.current_tenant_id` that is set at the start of every request by `store.SetTenantContextFromCtx(ctx)`. That method calls the `set_tenant_context($1)` stored procedure, which validates that the tenant exists and has `status = 'ACTIVE'` before calling `set_config('app.current_tenant_id', $1, TRUE)`. The `TRUE` flag makes the setting transaction-local: it resets automatically on `COMMIT` or `ROLLBACK`, so no cleanup is needed between requests.

This means a query like `SELECT * FROM invoice` never needs a `WHERE tenant_id = ?` clause — the RLS policy injects it at the database level. Even if application code forgets the filter, the database enforces it.

**Global tables** (platform-wide, no RLS): `tenants`, `audit_log`, `timezones`, `currencies`, `countries`, `paye_bands`, `platform_admins`. These are read-only for the application role and require no tenant filter.

**PgBouncer transaction mode** is required because the `set_config` call uses `is_local = TRUE`, which resets at transaction end. This is safe: each request opens a transaction, sets the tenant context, executes queries, and commits — the context is always fresh.

### 3.4.4 Shared connection pool — PgBouncer in front of PostgreSQL

The store layer uses a single shared pgx connection pool (via PgBouncer in transaction mode). There is no per-tenant pool. Tenant isolation is enforced by the RLS policies, not by connection routing. This model scales to thousands of tenants without creating thousands of connection pools. Pool size is bounded by the PgBouncer `max_client_conn` and `default_pool_size` settings in `config/pgbouncer.ini`.

### 3.4.5 Per-tenant feature flags — Redis-backed, per-tenant overrides

Feature flags are evaluated per-tenant. The flag evaluation order is: system default value → tenant-level override → user-level override. Resolved flag values are cached in Redis with a 5-minute TTL under the key `eval:{hash}` where the hash is a SHA-256 of the flag name, tenant ID, and user ID. See [§40 Feature Flag System — Deep Dive] for the complete flag evaluation model.

### 3.4.6 Per-tenant configuration — TenantConfig entity, inheritance from system defaults

Tenant configuration is stored in the `tenant_configs` table within each tenant's schema as key-value pairs with typed values. Configuration values that are not set at the tenant level fall back to system defaults defined in the framework's configuration struct. The configuration system is accessed through a typed `TenantConfig` service that enforces type safety at read time, not through raw key-string lookups.

---

## 3.5 Component Dependency Map

### 3.5.1 Startup order — config → database → Redis → EntityRegistry → Fiber → Temporal worker

The startup sequence has a strict dependency order. Configuration must be validated before any other component starts, because all components read configuration at initialization time. The database connection pool must be established before the EntityRegistry is populated, because system entity registration validates column existence against the live schema. Redis must be reachable before the session middleware can validate session tokens. The EntityRegistry must be fully populated before Fiber starts accepting requests, because route registration happens during EntityDefinition registration. The Temporal worker can start before Fiber, after, or concurrently — it has no dependency on the HTTP server.

```
Config load & validate
    ↓
PostgreSQL pool init (ping to verify)
    ↓
Redis client init (ping to verify)
    ↓
EntityRegistry init + system entity registration (all modules)
    ↓
Fiber app init + route registration (derived from EntityRegistry)
    ↓
Fiber server start (begin accepting HTTP connections)
    ↓ (concurrent)
Temporal worker start (register workflows + activities, connect to Temporal)
```

### 3.5.2 What depends on what — dependency graph for contributors

Understanding the dependency graph prevents circular imports and clarifies which layer owns which concern:

- API layer depends on: Domain layer (EntityDefinition, hooks, validators), Store layer (EntityRepository), IAM service (AuthzService, SessionService)
- Domain layer depends on: nothing outside the framework core — it is dependency-free at the interface level
- Workflow layer depends on: Store layer (EntityRepository accessed in activities), external services (KRA eTIMS client, email client, SMS client)
- Store layer depends on: PostgreSQL driver (pgx), Redis client (rueidis) for the session and cache services
- No layer may import from a higher layer — the dependency graph is strictly downward

### 3.5.3 What can fail independently without cascading — Redis, Temporal

Redis failure degrades gracefully for most use cases: feature flag evaluation falls back to system defaults, page schema requests regenerate without cache, and rate limiting bypasses to allow-all (configurable). Session validation cannot degrade gracefully because it requires Redis to validate session tokens; a Redis failure causes all authenticated requests to fail with 503 Service Unavailable. This is a security-correct behavior: the alternative (allowing requests without session validation) would be a security regression.

Temporal failure prevents new workflows from starting and prevents running workflows from making progress, but it does not affect synchronous CRUD operations. An invoice can still be created and updated when Temporal is unavailable; it simply cannot be submitted (because submission requires starting an InvoiceSubmissionWorkflow). The system remains operational for read-heavy and create/update workloads during a Temporal outage.

### 3.5.4 What failing brings the process down — database, EntityRegistry

The PostgreSQL database is a hard dependency. If the connection pool cannot establish connections at startup, the process fails to start. If the pool exhausts all connections during operation, requests that require database access fail with 503 until connections become available. The Fiber process itself remains running; it does not exit on database errors during operation.

The EntityRegistry is a hard dependency only during startup. If system entity registration fails (e.g., a module's Register function returns an error), the process fails to start. After startup, the EntityRegistry is read-only and cannot fail — it is an in-memory data structure.

---

## Chapter Summary

Chapter 3 defines the five-layer architecture (§3.1), the read and write request lifecycles (§3.2 and §3.3), the multi-tenant runtime model (§3.4), and the component dependency graph (§3.5).

The most important concepts for day-to-day development are:

- **The write path transaction boundary** (§3.3.5–3.3.8) — determines what is atomic and what is eventually consistent. `after_save` hooks are inside the transaction; Temporal workflow starts are outside it.
- **RLS-based tenant isolation** (§3.4.3) — `set_tenant_context()` at request start is the single enforcement point. Every tenant-scoped table has `FORCE ROW LEVEL SECURITY`. Global tables have no RLS.
- **The startup order** (§3.5.1) — config → PostgreSQL → Redis → EntityRegistry → Fiber → Temporal. Registry failures are fatal; Temporal failures are not.

**Next chapters to read:**

- [§13 — Middleware Pipeline](../part-03-api/13-middleware-pipeline.md) — expands §3.2.2 into the full middleware reference, including canonical execution order and each middleware's failure behaviour
- [§14 — Multi-Tenancy Middleware](../part-03-api/14-multitenancy-middleware.md) — expands §3.4 into the RLS implementation details: `set_tenant_context()`, global vs. tenant-scoped tables, and `validate_tenant_context()` for long-running activities
- [§8 — The Persistence Interface](../part-02-entity-system/08-persistence-interface.md) — expands §3.1.5 into the full `EntityRepository` contract, the Filter DSL, and transaction scoping
