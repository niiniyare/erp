# Chapter 05 — Server-Driven UI Fundamentals

> **Volume:** II — DSL and AST
> **Audience:** Backend Engineers, Platform Engineers
> **Prerequisites:** Chapters 01–04

---

## Table of Contents

- [5.1 What Makes This SDUI](#51-what-makes-this-sdui)
- [5.2 The AMIS Schema as the Wire Format](#52-the-amis-schema-as-the-wire-format)
- [5.3 UISessionContext: The Input to Every Page](#53-uisessioncontext-the-input-to-every-page)
- [5.4 PageFn vs ASTPageFn](#54-pagefn-vs-astpagefn)
- [5.5 How the Browser Uses the Schema](#55-how-the-browser-uses-the-schema)
- [5.6 Data Loading: InitAPI and API](#56-data-loading-initapi-and-api)
- [5.7 AMIS Expression Syntax](#57-amis-expression-syntax)
- [5.8 The AMIS Response Envelope](#58-the-amis-response-envelope)

---

## 5.1 What Makes This SDUI

In a conventional frontend architecture, the browser downloads an application bundle that contains the UI logic. When a user opens a page, the browser runs JavaScript that fetches data and assembles a view.

In AwoERP's SDUI model:
1. The browser requests a schema: `GET /schema/finance/dashboard`
2. The backend compiles a JSON description of the page — including which components appear, which data sources they use, which actions are available — specific to this user's authorization context
3. The browser's AMIS SDK receives the schema and renders it

The browser never makes a UI decision based on the user's role or permissions. Those decisions were made during compilation. The browser executes exactly what it receives.

This produces a structural security guarantee: if a component is not in the schema, it cannot be accessed from the browser, regardless of what the user attempts.

---

## 5.2 The AMIS Schema as the Wire Format

An AMIS schema is a `map[string]any` — in Go terms, `ui.Schema` (which is `ui.M` which is `map[string]any`). Every AMIS node has a `"type"` field. The AMIS SDK reads the type and renders the appropriate component.

A minimal page schema:

```json
{
  "type": "page",
  "title": "Finance Dashboard",
  "body": [
    {
      "type": "tpl",
      "tpl": "Welcome, ${user.name}"
    }
  ]
}
```

A schema with data loading:

```json
{
  "type": "page",
  "title": "Invoices",
  "initApi": {
    "method": "get",
    "url": "/api/v1/finance/dashboard/summary"
  },
  "body": [
    {
      "type": "stat",
      "label": "Total Revenue",
      "value": "${revenue}",
      "format": "currency"
    }
  ]
}
```

A schema with a data table:

```json
{
  "type": "crud",
  "api": {
    "method": "get",
    "url": "/api/v1/finance/invoices"
  },
  "syncLocation": false,
  "columns": [
    {"name": "invoice_number", "label": "Invoice #"},
    {"name": "status", "label": "Status", "type": "mapping",
     "map": {"DRAFT": "<span class='badge'>Draft</span>"}}
  ]
}
```

The Go types that produce these structures:

```go
// ui.M = map[string]any
// ui.A = []any
// ui.Schema = map[string]any

type M = map[string]any
type A = []any
type Schema = M
```

---

## 5.3 UISessionContext: The Input to Every Page

`UISessionContext` is the only input a page function receives. It carries the resolved identity, permissions, feature flags, locale, and URL params for the current request.

```go
// internal/web/ui/types.go

type UISessionContext struct {
    // Identity
    UserID      string
    TenantID    string
    DisplayName string
    IsPlatform  bool
    IsPortal    bool

    // Locale
    Locale   string // e.g. "en-GB"
    Timezone string // e.g. "Africa/Nairobi"
    Currency string // e.g. "KES"

    // URL route params (e.g. "/finance/invoices/:id" → Params["id"])
    Params map[string]string

    // private: resolved by AuthzStage
    permissions  map[string]bool
    featureFlags map[string]bool
    prefs        map[string]string
}

// Can returns true if the pre-resolved permission map contains "resource.action".
// This is NOT a Casbin call.
func (u UISessionContext) Can(action, resource string) bool {
    return u.permissions[resource+"."+action]
}

// Flag returns true if the named feature flag is active for this session.
func (u UISessionContext) Flag(name string) bool {
    return u.featureFlags[name]
}

// Param returns a URL route parameter, or fallback if absent.
func (u UISessionContext) Param(key, fallback string) string {
    if v, ok := u.Params[key]; ok {
        return v
    }
    return fallback
}
```

`UISessionContext` is constructed exclusively by AuthzStage. Page functions must not construct it themselves. They receive it as a function argument.

**Using UISessionContext in a page function:**

```go
func InvoicePage(sess ui.UISessionContext) ast.Node {
    actions := []ast.Node{
        ast.ActionNode{Label: "Save Draft", ActionType: "ajax", Level: "default"},
    }

    // Conditionally add actions based on pre-resolved permissions
    if sess.Can("approve", "finance.invoices") {
        actions = append(actions, ast.ActionNode{
            Label:      "Approve",
            ActionType: "ajax",
            Level:      "success",
            API:        &ast.APISpec{Method: "post", URL: "/api/v1/finance/invoices/${id}/approve"},
        })
    }

    // Conditionally include components based on feature flags
    var body []ast.Node
    if sess.Flag("finance.auto_reconcile") {
        body = append(body, ast.SectionNode{Title: "Auto-Reconcile"})
    }

    return ast.PageNode{
        Title:  "Invoice",
        Body:   append(body, ast.FormNode{Actions: actions}),
    }
}
```

---

## 5.4 PageFn vs ASTPageFn

Two function signatures can be registered for a route:

```go
// Legacy — returns raw map[string]any
type PageFn func(sess UISessionContext) Schema

// Preferred — returns ast.Node; CompileTree validates before serializing
type ASTPageFn func(sess UISessionContext) any // actual return type: ast.Node
```

`CompileStage` checks for `ASTPageFn` first (`DataKeyASTPageFn`). If present, it calls `ast.CompileTree(ASTPageFn(sess))`. Validation errors in the tree are caught before any JSON is emitted, and the pipeline returns an error rather than a broken schema.

If only `PageFn` is registered, `CompileStage` falls back to calling `PageFn(sess)` directly. The result is accepted as-is; only `NormalizeStage` and `ValidateStage` apply corrections and checks.

**Rule:** New pages must use `ASTPageFn`. Use the typed AST (`ast.PageNode`, `ast.CRUDNode`, etc.) and register with `ASTFn`. The legacy `PageFn` path is maintained for backward compatibility during migration.

**When registering:**

```go
registry.RegisterPage(registry.PageRegistration{
    Route:  "/finance/dashboard",
    Module: "finance",
    Title:  "Finance Dashboard",
    ASTFn:  FinanceDashboardScreen,  // preferred
    // Fn: LegacyDashboard,          // only for old pages during migration
})
```

---

## 5.5 How the Browser Uses the Schema

The browser loads `web/pages/index.html`, which bootstraps the AMIS SDK. The AMIS SDK maintains an active schema for the current route.

When the user navigates:
1. The SDK makes a `GET /schema/<route>` request
2. The response envelope `{"status": 0, "data": <schema>}` arrives
3. The SDK reads `data` and renders it

The SDK handles all component lifecycle, data fetching via `initApi`/`api` fields, action execution, and form validation based on the schema it receives. No custom JavaScript is required per page.

On error:
- `{"status": 401, ...}`: The SDK renders the AMIS page schema included in the 401 response (a session-expired alert with a login button)
- `{"status": 404, "msg": "..."}`: The SDK renders the error message
- `{"status": 500, "msg": "..."}`: Same

---

## 5.6 Data Loading: InitAPI and API

Page functions do not fetch data. Data is fetched by the AMIS SDK at runtime, after the schema is rendered.

**`initApi`** — A request made when the page loads. The response is merged into the page's data scope. Children can reference fields with `${fieldName}`.

**`api`** — A request made by a `crud` component to populate its table. Called on load and on filter/sort/page changes.

```go
ast.PageNode{
    Title: "Finance Dashboard",
    // initApi fetches KPI data; ${revenue}, ${ar_balance} are available in Body
    InitAPI: &ast.APISpec{
        Method: "get",
        URL:    "/api/v1/finance/dashboard/summary",
    },
    Body: []ast.Node{
        ast.StatNode{Label: "Revenue", Value: "${revenue}"},
        ast.StatNode{Label: "AR Balance", Value: "${ar_balance}"},
    },
}
```

For a CRUD table:

```go
ast.CRUDNode{
    API: ast.APISpec{
        Method: "get",
        URL:    "/api/v1/finance/invoices",
    },
    Columns: []ast.TableColumn{
        {Name: "invoice_number", Label: "Invoice #"},
        {Name: "amount", Label: "Amount", Type: "number"},
    },
}
```

The business API (`/api/v1/finance/invoices`) returns paginated data. The AMIS SDK handles pagination, sorting, and filtering by appending query parameters to the URL.

---

## 5.7 AMIS Expression Syntax

AMIS uses `${expression}` syntax in string fields for dynamic values. Expressions are evaluated by the AMIS SDK at runtime against the current data scope.

```json
{"type": "tpl", "tpl": "Hello, ${user.name}"}
{"type": "stat", "value": "${total_amount}"}
{"type": "button", "disabledOn": "${!can_approve}"}
{"type": "button", "visibleOn": "${status === 'PENDING'}"}
```

**Important:** The correct syntax is `"${!can_approve}"` (the `!` is inside the expression), not `"!${can_approve}"` (which produces a string `"!true"` or `"!false"`, not a boolean).

When writing legacy `PageFn` schemas, expression strings appear directly in map values:

```go
// Legacy PageFn pattern
func Page(sess ui.UISessionContext) ui.Schema {
    return ui.M{
        "type": "page",
        "body": ui.M{
            "type":       "button",
            "label":      "Approve",
            "disabledOn": "${!can_approve}", // correct
            "visabledOn": "${status === 'PENDING'}",
        },
    }
}
```

In AST nodes, AMIS expression strings appear in fields typed as `string`. The AST does not validate expression syntax — that is the author's responsibility.

---

## 5.8 The AMIS Response Envelope

Every successful schema response uses the AMIS envelope format:

```json
{
  "status": 0,
  "data": {
    "type": "page",
    "title": "...",
    "body": [...]
  }
}
```

The `status: 0` signals success to the AMIS SDK. Any non-zero status is treated as an error.

Error envelopes:

```json
{"status": 404, "msg": "schema not found: /finance/nonexistent"}
{"status": 500, "msg": "An internal error occurred. Please try again."}
```

The 401 response includes an AMIS schema in `data` so the browser renders a usable page:

```json
{
  "status": 401,
  "msg": "Session expired. Please log in.",
  "data": {
    "type": "page",
    "body": {"type": "alert", "body": "Your session has expired.", "level": "warning"},
    "toolbar": [{"type": "button", "label": "Log In", "level": "primary", "actionType": "url", "url": "/login"}]
  }
}
```

This is intentional. If the 401 returned no AMIS schema, the SDK would display a blank page or a raw error string. By returning an AMIS page, the user sees a meaningful message and a login button.

---

*End of Chapter 05*

**Previous:** [Chapter 04 — System Overview](../vol-01-vision/04-system-overview.md)
**Next:** [Chapter 06 — UI DSL Architecture](./06-ui-dsl-architecture.md)
