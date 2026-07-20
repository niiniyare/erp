> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

# Chapter 19 — Navigation Framework

| | |
|---|---|
| **Volume** | 04 — Rendering |
| **Chapter** | 19 |
| **Status** | Partially implemented (hash routing implemented; Go-driven nav planned) |
| **Source** | `web/pages/index.html`, `internal/web/dsl/registry/` |

---

## Table of Contents

1. [19.1 Current State](#191-current-state)
2. [19.2 Route Resolution](#192-route-resolution)
3. [19.3 NavFn Type](#193-navfn-type)
4. [19.4 Registry Paths](#194-registry-paths)
5. [19.5 Missing: Param Routing](#195-missing-param-routing)
6. [19.6 Planned: Go-Driven Navigation Menu](#196-planned-go-driven-navigation-menu)

---

## 19.1 Current State

Navigation in the web shell is **hash-based routing driven by a static
JavaScript array**. No server call is made to determine which menu items to
display — the sidebar is built from a `menuConfig` constant embedded in
`web/pages/index.html`.

### What is implemented

| Feature | Implementation |
|---|---|
| Hash-based routing (`#finance/dashboard`) | `hashchange` event listener in `index.html` |
| Sidebar with collapsible groups | Static `menuConfig` array + vanilla JS DOM manipulation |
| Active item highlighting (exact + prefix match) | `updateNav()` function |
| Auto-expand group on navigation | `updateNav()` with prefix matching |
| Breadcrumb trail | `updateBreadcrumb()` reads from `menuConfig` |
| Sidebar search filter | `filterMenu()` filters DOM items by label |
| Collapsed sidebar with tooltips | CSS `.collapsed` class + `data-tooltip` attribute |
| Mobile drawer sidebar | `.mobile-open` class + backdrop |

### What is not yet implemented

- Go-driven navigation menu (menu items defined in Go, served over API)
- Param routing (`/finance/invoices/:id`)
- Per-user menu filtering based on IAM permissions
- Deep-link support for parameterised routes

---

## 19.2 Route Resolution

The routing flow from user click to rendered page:

```
User clicks sidebar item
  → href="#finance/dashboard"
  → window.location.hash = "#finance/dashboard"
  → hashchange event fires
  → navigate() called
  → route = "finance/dashboard"
  → fetch GET /schema/finance/dashboard
  → server returns AMIS schema envelope
  → amis.embed(#content, schema, amisEnv)
  → AMIS mounts and renders
```

### Route string format

Routes are slash-separated path segments corresponding directly to the URL path
after `/schema/`:

| Hash | Route string | Schema endpoint |
|---|---|---|
| `#dashboard` | `dashboard` | `GET /schema/dashboard` |
| `#finance/accounts` | `finance/accounts` | `GET /schema/finance/accounts` |
| `#finance/transactions` | `finance/transactions` | `GET /schema/finance/transactions` |

### Active item highlighting

The shell performs **exact match** first, then **prefix match** as a fallback.
This allows a future param route (`finance/invoices/abc-123`) to highlight the
nearest registered menu item (`finance/invoices/new`) by stripping trailing
`/new` and prefix-comparing:

```javascript
var base = ir.replace(/\/new$/, '');
if (base && route.indexOf(base) === 0 && base.length > bestLen) {
  bestLen = base.length;
  bestEl = items[j];
}
```

---

## 19.3 NavFn Type

The DSL layer defines a `NavFn` function type intended for Go-driven navigation
generation:

```go
// NavFn returns a slice of menu items for a given UI session.
// The session provides the authenticated user's tenant, roles, and
// resolved permission set so menu items can be filtered server-side.
type NavFn func(sess UISessionContext) []M
```

Where `M` is `map[string]any` — the raw JSON map used throughout the DSL layer
for schema fragments.

`NavFn` is designed to be registered alongside page schemas and called when the
browser requests the navigation structure. The return value will be serialised
to JSON and served over a dedicated endpoint (planned: `GET /api/v1/ui/nav`).

**Current status**: `NavFn` is defined but not wired to the web shell. The shell
does not call any navigation endpoint — it reads `menuConfig` directly.

---

## 19.4 Registry Paths

The DSL registry (`internal/web/dsl/registry/`) tracks every registered page
route. The `registry.Paths()` method returns the complete list:

```go
// Paths returns all route strings registered via init() calls.
// Example: ["dashboard", "finance/accounts", "finance/transactions", ...]
func (r *Registry) Paths() []string
```

This is used by:

- The schema handler, to validate that a requested route exists before running
  the pipeline.
- Health checks, to report which modules have registered UI screens.
- (Planned) The navigation endpoint, to enumerate available routes.

Pages are registered using blank-import side effects in
`internal/api/handlers/routes.go`:

```go
import (
    _ "awo.so/internal/web/pages/dashboard"
    _ "awo.so/internal/web/pages/finance/accounts"
    _ "awo.so/internal/web/pages/finance/transactions"
    _ "awo.so/internal/web/pages/organizations"
    _ "awo.so/internal/web/pages/settings"
    _ "awo.so/internal/web/pages/users"
    _ "awo.so/internal/web/dsl/screens"
)
```

Each `init()` function calls `registry.Register(route, pageFn)`.

---

## 19.5 Missing: Param Routing

> **Open item** (OPEN 2 from review.md)

The current shell has no mechanism to navigate to a specific record's detail
page via a URL parameter. For example:

```
#finance/invoices/abc-123-def-456
```

This would require:

1. The shell to extract the param segment (`abc-123-def-456`) from the hash.
2. The schema handler to accept the param and pass it to the page function.
3. The page function (`ASTFn`) to use the param in its `initApi` URL.

**Current workaround**: list pages with row actions that open a dialog or drawer
for editing. This avoids the param routing problem at the cost of UX (no direct
URL to a specific record).

**Planned solution**: extend the hash parser to support a separator between the
route and params:

```
#finance/invoices?id=abc-123
```

or embed params in the path and register wildcard routes:

```
registry.RegisterWildcard("finance/invoices/*", invoiceDetailFn)
```

Neither approach is implemented today.

---

## 19.6 Planned: Go-Driven Navigation Menu

> **Status: Planned — Not Yet Implemented**

Today the sidebar menu is a static JavaScript array. This has two problems:

1. **Maintenance**: adding a new screen requires editing `index.html`.
2. **IAM filtering**: the menu cannot hide items the user lacks permission to
   see, because `index.html` is served before authentication is checked.

The planned architecture:

```
Browser loads index.html (unauthenticated shell)
  → User logs in
  → Shell calls GET /api/v1/ui/nav  (authenticated)
  → Server calls NavFn(sess) for each registered module
  → NavFn filters items based on sess.Permissions
  → Server returns JSON nav tree
  → Shell replaces menuConfig with server response
  → renderMenu() re-renders sidebar
```

The `NavFn` type (section 19.3) is the Go interface point. Each module that
registers screens will also register a `NavFn` that knows which items to include
for a given session.

Until this is implemented, the sidebar always shows all items regardless of the
user's IAM roles.
