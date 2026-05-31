# Glossary

> Last verified: 2026-05-18 | Code pointer: `internal/web/ui/types.go`, `internal/web/registry/registry.go`

Terms that cause confusion or are easy to misuse. When in doubt, check the code pointer.

---

## UIContext vs UISessionContext (CRITICAL)

**`UIContext` does not exist.** There is no type named `UIContext` in this codebase.

The correct type is `UISessionContext`. It is defined in `internal/web/ui/types.go`:

```go
type UISessionContext struct {
    UserID      string
    TenantID    string
    DisplayName string
    IsPlatform  bool
    IsPortal    bool
    Locale      string
    Timezone    string
    Currency    string
    // private fields
}
```

If you see `UIContext` in old documentation, code review comments, or example code, replace it with `UISessionContext`. A compile error with a message like `undefined: ui.UIContext` means someone used the wrong name.

**Why this matters:** `UISessionContext` is the *only* identity object that reaches a `PageFn` or `ASTPageFn`. It is not a pointer — it is passed by value and is immutable after construction.

---

## PageFn vs ASTPageFn

Both are function types that build a page schema. They differ in return type and when to use them.

**`PageFn`** (legacy):

```go
type PageFn func(sess UISessionContext) Schema
// Schema = map[string]any
```

Returns a raw `map[string]any`. No compile-time guarantees about schema correctness. Structural validation happens at request time in `ValidateStage`. This is the "write JSON in Go" approach.

**`ASTPageFn`** (preferred for new pages):

```go
type ASTPageFn func(sess UISessionContext) any // returns ast.Node
```

Returns a typed `ast.Node`. The return type is `any` only to avoid a circular import — the actual value must implement `ast.Node`. `CompileStage` asserts it at runtime. The typed AST enforces structural rules (syncLocation, transparent background, API prefixes) at build time, so `ValidateStage` skips structural checks for AST-compiled schemas.

**When to use which:**
- New pages: always use `ASTPageFn` (`ASTFn` field in `PageRegistration`).
- Existing pages: `PageFn` continues to work unchanged during the migration window.
- Both set simultaneously: `CompileStage` always uses `ASTPageFn` and ignores `PageFn`.

---

## contract.SessionContext vs UISessionContext

These are two different types that serve different layers.

**`contract.SessionContext`** (`internal/core/iam/contract`):

The IAM layer's session object. Contains the authenticated user's identity, tenant membership, roles, and raw feature flags. It is a Go interface. It is available in HTTP handlers via `contract.FromContext(ctx)`. It has methods like `UserID()`, `TenantID()`, `IsPlatform()`, `IsPortal()`, `FeatureEnabled()`, `Preference()`, `Setting()`.

**`UISessionContext`** (`internal/web/ui`):

A snapshot of the IAM session, narrowed and extended for the UI compilation pipeline. It is a concrete struct, not an interface. It contains pre-resolved permissions (the result of `BulkEnforce()`) instead of raw IAM access. It is constructed once per request by `AuthzStage` and then passed immutably through the pipeline to `PageFn`/`ASTPageFn`.

`UISessionContext` is **not** a `contract.SessionContext`. You cannot use one where the other is expected.

**The relationship:**

```
HTTP Request
    ↓
contract.InjectSessionContext() middleware
    → stores contract.SessionContext in Go context
        ↓
AuthzStage
    → reads contract.SessionContext from context
    → calls BulkEnforce() → permissions map
    → calls ui.NewUISessionContext(sc, permissions)
        → creates UISessionContext (snapshot)
            ↓
CompileStage → PageFn(UISessionContext) → Schema
```

After `AuthzStage`, the `contract.SessionContext` is no longer needed by the UI pipeline. Only `UISessionContext` flows forward.

---

## Permission Key Format

Permissions are checked using `sess.Can(action, resource)`:

```go
sess.Can("view", "invoice")
sess.Can("create", "purchase_order")
sess.Can("approve", "leave_request")
```

Internally, `Can("view", "invoice")` checks the resolved map for the key `"invoice.view"` (resource first, then action, separated by dot):

```go
func (u UISessionContext) Can(action, resource string) bool {
    return u.permissions[resource+"."+action]
}
```

**The format is `resource.action`, not `action:resource` or `action.resource`.**

Common mistake:

```go
// WRONG — checks "view.invoice", which is never true
sess.Can("invoice", "view")

// CORRECT
sess.Can("view", "invoice")
```

The Casbin policy file uses the same `resource.action` format. If you are adding a new permission to `authz.AllUIPermissions`, use `"resource.action"` (lowercase, dot-separated).

---

## Envelope / AmisResponse

**AMIS envelope** is the JSON structure AMIS expects from every API call:

```json
{
  "status": 0,
  "data": { ... }
}
```

`status: 0` means success. Non-zero means error. AMIS reads this format for all its `api:` calls.

**Go backend envelope** is what the Go API handlers return:

```json
{
  "success": true,
  "data": [...],
  "meta": { "pagination": { "total_records": 142 } }
}
```

The frontend `fetcher` function bridges between these two formats.

**Schema endpoint envelope:** When a `PageFn` returns a `Schema`, the `SchemaHandler` wraps it:

```json
{
  "status": 0,
  "data": { ...the schema... }
}
```

**Critical rule:** A `PageFn` must return the raw schema (`map[string]any`), not a pre-wrapped envelope. If a `PageFn` returns `map[string]any{"status": 0, "data": theSchema}`, the handler wraps it again, producing a double-nested response that AMIS cannot parse.

---

## Stage vs Handler vs Middleware

**Stage** (`pipeline.Stage`): A single unit of work in the UI compilation pipeline. Implements `Execute(opCtx) (StageResult, error)`. Stages communicate via `opCtx.Data`. Examples: `SessionStage`, `AuthzStage`, `CompileStage`. Defined in `internal/web/stages/`.

**Handler** (Fiber handler, `fiber.Handler`): An HTTP endpoint function. In the UI layer, `SchemaHandler` is the Fiber handler that builds `opCtx`, runs the pipeline, and writes the HTTP response. Handlers are thin wrappers — no business logic lives in a handler.

**Middleware** (Fiber middleware, `fiber.Handler`): A function in the Fiber route chain that runs before the handler. Examples: `contract.InjectSessionContext()` (injects the IAM session into the Go context), `AuthMiddleware` (validates the JWT). Middleware runs at the HTTP layer. Stages run inside the pipeline abstraction, which has no knowledge of HTTP.

Summary:

```
HTTP Request
  → Middleware 1 (auth token validation)
  → Middleware 2 (contract.InjectSessionContext)
  → Handler (SchemaHandler)
       → pipeline.Run()
            → Stage 10 (SessionStage)
            → Stage 20 (AuthzStage)
            → ...
       ← DataKeyResponse
  → c.JSON(200, response)
```

---

## Soft vs Hard Cache Invalidation

**Soft invalidation**: The cache entry is not deleted. It expires naturally after its TTL (5 minutes for UI schemas). Until expiry, stale data is served. This happens automatically and requires no action.

**Hard invalidation**: The cache entry is actively deleted before its TTL expires. In the UI layer, this is done by calling `DeletePattern` on the cache service with a pattern that matches the relevant cache keys. Hard invalidation is triggered by significant state changes: a user's role assignment changes, a feature flag is toggled for a tenant, a schema is deployed.

**Which to use:**
- Role assignment change → hard invalidation of all cache entries for that user (match on `tenantID:userID` or `permFP`)
- Feature flag change for tenant → hard invalidation for all users in that tenant
- New deployment with schema changes → increment `CacheVersions` generation, which effectively hard-invalidates all entries by making the old keys unreachable (they still expire naturally)
- Minor configuration change with no permission impact → soft invalidation (wait for TTL)

---

## Registry vs Router

**Registry** (`internal/web/registry`): The in-memory map from route path (`"/finance/invoices"`) to `PageRegistration`. It is populated by `init()` functions calling `RegisterPage()`. It has no knowledge of HTTP. It is queried by `RegistryStage` during pipeline execution.

**Router** (Fiber router): The HTTP routing layer that maps HTTP method + path to a Fiber handler function. The Fiber router maps `GET /schema/finance/invoices` to `SchemaHandler`. `SchemaHandler` then queries the Registry to find the `PageFn` for that route.

The distinction matters when debugging a 404:
- If the Fiber router has no route for the path → HTTP 404 from Fiber, never reaches SchemaHandler.
- If the Fiber router routes to SchemaHandler but the Registry has no entry for the route → `PageNotFoundError` from `RegistryStage`, HTTP 404 with a body.

Check the Fiber route definitions for the first case. Check `registry.Paths()` for the second.
