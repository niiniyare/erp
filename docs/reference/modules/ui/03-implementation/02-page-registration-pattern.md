# Page Registration Pattern

> Last verified: 2026-05-18 | Code pointer: `internal/web/registry/registry.go`

---

Every page in the UI must be registered before the HTTP server starts. Registration is how the pipeline's RegistryStage maps a URL route to a Go function that builds the page schema. Without registration, the route returns 404.

---

## PageRegistration Struct Fields

```go
// internal/web/registry/registry.go
type PageRegistration struct {
    Route       string       // required — URL path, e.g. "/finance/invoices"
    Module      string       // required — top-level domain, e.g. "finance"
    Title       string       // required — human-readable, used in logs and nav tree
    Description string       // optional — surfaced in the registry debug endpoint
    Fn          ui.PageFn    // legacy builder — nil if ASTFn is set
    ASTFn       ui.ASTPageFn // preferred builder — returns ast.Node
}
```

Field rules enforced by `ValidateRegistry()`:

| Field | Rule | Error Code |
|-------|------|------------|
| `Route` | Must start with `/`. Must not have a trailing slash (except `"/"` itself). | `REGISTRY_INVALID_ROUTE` |
| `Module` | Must not be empty. Use lowercase: `"finance"`, `"iam"`, `"inventory"`, `"dashboard"`. | `REGISTRY_MISSING_MODULE` |
| `Title` | Must not be empty. | `REGISTRY_MISSING_TITLE` |
| `Fn` / `ASTFn` | At least one must be non-nil. | `REGISTRY_MISSING_FN` |

`Description` is optional but recommended. It is surfaced in the registry debug endpoint and in trace metadata.

---

## The `init()` Pattern

Go's `init()` functions run once at program startup, before `main()`, after all package-level variables are initialized. This is the correct place to register pages: it is guaranteed to run before `ValidateRegistry()` is called, and it requires no dependency injection.

### Complete Working Example

```go
// internal/web/pages/finance/invoices/schema.go
package invoices

import (
    "awo.so/internal/web/registry"
    "awo.so/internal/web/ui"
)

func init() {
    registry.RegisterPage(registry.PageRegistration{
        Route:       "/finance/invoices",
        Module:      "finance",
        Title:       "Invoices",
        Description: "Paginated list of all invoices with filter by status and date range.",
        ASTFn:       PageSchema,  // preferred — typed AST
        // Fn:       LegacyPageSchema,  // legacy fallback if not yet migrated
    })
}

// PageSchema builds the invoice list page schema.
// Same UISessionContext produces identical output — no I/O, no side effects.
func PageSchema(sess ui.UISessionContext) any { // returns ast.Node
    // ... build and return an ast.Node
    return nil // replace with real implementation
}
```

### Ensuring `init()` Runs

Go only runs `init()` for packages that are imported. A page package that is never imported will never register itself. The standard pattern is a blank import in the server's main package or a dedicated `pages` aggregate package:

```go
// internal/web/pages/pages.go — blank imports all page packages
package pages

import (
    _ "awo.so/internal/web/pages/dashboard"
    _ "awo.so/internal/web/pages/finance/invoices"
    _ "awo.so/internal/web/pages/finance/purchase-orders"
    _ "awo.so/internal/web/pages/iam/users"
    // add new pages here
)
```

Then in main or the server setup:

```go
import _ "awo.so/internal/web/pages"
```

One import triggers all page registrations.

---

## ValidateRegistry() Rules

`ValidateRegistry()` is called at application startup, after all `init()` functions have run and before the HTTP server accepts traffic.

```go
// Call this in your startup sequence, after all init() have run:
if err := registry.ValidateRegistry(); err != nil {
    panic(err)
}
```

What it checks for every registered `PageRegistration`:

1. `Route` is non-empty and starts with `"/"`.
2. `Route` does not end with `"/"` (unless the route is exactly `"/"`).
3. `Module` is non-empty.
4. `Title` is non-empty.
5. At least one of `Fn` or `ASTFn` is non-nil.

If any registration fails, `ValidateRegistry()` returns a `*sharedErrors.BusinessError` with:
- Code: `REGISTRY_VALIDATION_FAILED`
- HTTP status: 500
- Detail fields: each violation listed as `violation_1`, `violation_2`, etc.

**The intent is that this panics.** A misconfigured page registration is a programmer error, not a runtime condition. The server should not start with broken registrations.

### Duplicate Route Behavior

`RegisterPage()` panics immediately on a duplicate route — it does not wait for `ValidateRegistry()`. The error message is:

```
panic: ui/registry: duplicate registration for route "/finance/invoices"
```

This catches `init()` ordering bugs (two packages both registering the same route) at startup rather than at request time.

---

## ASTPageFn vs PageFn: Which to Use

**Use `ASTFn` (preferred) for all new pages.**

`ASTPageFn` returns an `ast.Node` (typed as `any` to avoid a circular import). The typed AST node tree is compiled by `ast.CompileTree()` inside CompileStage, which validates structural invariants before emitting JSON. This means:

- No `VALIDATE_CRUD_SYNC_LOCATION` errors at runtime — `CRUDNode` always sets `syncLocation:true`.
- No `VALIDATE_CHART_TRANSPARENT_BG` errors at runtime — `ChartNode` always sets the transparent background.
- No `VALIDATE_API_METHOD_PREFIX` errors at runtime — API nodes enforce prefixes.
- `ValidateStage` skips structural rules and runs only security rules.

```go
// Preferred pattern — typed AST
func PageSchema(sess ui.UISessionContext) any { // returns ast.Node
    return ast.PageNode{
        Title: "Invoices",
        Body: ast.CRUDNode{
            API: ast.APIGET("/api/v1/finance/invoices"),
            // syncLocation:true is automatic
        },
    }
}
```

**Use `Fn` (legacy) only when migrating existing pages** that already work with the raw `map[string]any` approach. During the migration window, `Fn` and `ASTFn` can coexist in the same registration — when both are set, `ASTFn` wins:

```go
registry.RegisterPage(registry.PageRegistration{
    Route:  "/finance/invoices",
    Module: "finance",
    Title:  "Invoices",
    Fn:     LegacySchema,   // kept during migration
    ASTFn:  NewASTSchema,   // CompileStage always uses this when present
})
```

Once migration is complete, remove `Fn` from the registration.

---

## Common Errors

### 404 on a Known Page

**Symptom:** `GET /schema/finance/invoices` returns 404.

**Cause checklist:**
1. The page package's `init()` never ran — the package is not imported anywhere. Check the blank imports aggregate.
2. The route has a trailing slash in the registration but not in the request (or vice versa). Routes are matched by exact string.
3. `RegisterPage()` was called after `ValidateRegistry()` ran. Registration after startup is ignored by the running server.
4. `registry.Register()` (deprecated) was used with a path that differs from the HTTP route.

**Diagnosis:** Add a temporary handler that calls `registry.Paths()` and returns the list. Compare the exact string to the route being requested.

### Startup Panic: Duplicate Route

**Symptom:** Server panics at startup with `ui/registry: duplicate registration for route "/finance/invoices"`.

**Cause:** Two `init()` functions call `RegisterPage()` with the same `Route` value.

**Fix:** Find both callers and resolve the conflict. Typically this happens when a page is moved between packages without removing the old registration.

### Startup Panic: REGISTRY_VALIDATION_FAILED

**Symptom:** `ValidateRegistry()` returns an error listing violations, and the caller panics.

**Common violations:**

```
route "/finance/invoices/": trailing slash not allowed
  Fix: remove trailing slash from Route field

route "finance/invoices": must start with "/"
  Fix: add leading slash

route "/finance/invoices": Module is required
  Fix: set Module field (e.g. "finance")

route "/finance/invoices": at least one of Fn or ASTFn must be set
  Fix: set either Fn or ASTFn
```

### CompileStage Error: Wrong Return Type

**Symptom:** HTTP 500 with message `ASTPageFn return value does not implement ast.Node (got *MyStruct)`.

**Cause:** The function registered as `ASTFn` returns a type that does not implement `ast.Node`.

**Fix:** Ensure the returned struct or value satisfies the `ast.Node` interface. Check the ast package for the correct node types to use.

### Stale Cache After Registration Change

**Symptom:** Updated page schema is not reflected in responses.

**Cause:** The old compiled schema is cached in Redis under the old cache key. Since the schema content changed but the fingerprint did not (same user, same permissions, same flags), the cache key is the same.

**Fix:** Increment the relevant `CacheVersions` generation number, or call `DeletePattern` on the route's cache key prefix. In development, restart the server to clear the in-process LRU.
