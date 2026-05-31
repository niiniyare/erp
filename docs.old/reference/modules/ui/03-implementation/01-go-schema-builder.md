# Go Schema Builder

> Last verified: 2026-05-18 | Code pointer: `internal/web/ui/types.go`, `internal/web/amis/`, `internal/web/dsl/`

---

## 📖 Why Go Builds the UI

AMIS JSON is the compiled artifact. Go functions are the source code. This inversion means all UI logic — permission checks, feature flags, conditional rendering, locale-aware labels — lives in Go where it can be tested, typed, and cached. The browser receives static JSON and renders it. No logic runs client-side.

---

## Two Approaches

Both are supported. New pages should use the **modern approach**.

| | Legacy (`amis.Ctx` + `PageFn`) | Modern (`UISessionContext` + `ASTPageFn`) |
|--|--|--|
| Context type | `amis.Ctx` | `ui.UISessionContext` |
| Function type | `ui.PageFn` / `amis.SchemaFn` | `ui.ASTPageFn` |
| Output type | `ui.Schema` (raw `map[string]any`) | `ast.Node` (typed, compiled to Schema) |
| IAM integration | Manual `Can()` closure | Pre-resolved by `AuthzStage` |
| Cache support | Works (shared cache) | Works (shared cache) |
| Compile-time checks | Partial | Full — type-safe node tree |
| Registration | `registry.Register(path, fn)` *(old)* | `registry.RegisterPage(PageRegistration{...})` |
| Pipeline dispatch | `DataKeyPageFn` fallback | `DataKeyASTPageFn` preferred |

**`CompileStage` checks for `ASTPageFn` first. Falls back to `PageFn` only if `ASTPageFn` is absent.**

---

## Core Types

Source: `internal/web/ui/types.go`

```go
// M is map[string]any. Every AMIS node is an M.
type M = map[string]any

// A is []any. AMIS arrays.
type A = []any

// Schema is the root compiled AMIS document (= M).
type Schema = M

// PageFn — legacy, still supported
type PageFn func(sess UISessionContext) Schema

// ASTPageFn — preferred for new pages
// Returns ast.Node (typed tree), compiled to Schema by CompileStage.
// Uses `any` return to avoid circular import between ui and ast packages.
type ASTPageFn func(sess UISessionContext) any // returns ast.Node

// UIBlock — reusable schema fragment, used for dashboard widgets etc.
type UIBlock func(sess UISessionContext) M
```

---

## Modern Approach: `ASTPageFn` + `PageRegistration`

### Registration

```go
// internal/web/pages/finance/invoices/schema.go
package invoices

import (
    "awo.so/internal/web/dsl/blocks"
    "awo.so/internal/web/dsl/screens"
    "awo.so/internal/web/registry"
    "awo.so/internal/web/ui"
)

func init() {
    registry.RegisterPage(registry.PageRegistration{
        Route:       "/finance/invoices",
        Module:      "finance",
        Title:       "Invoices",
        Description: "Invoice list and management",
        ASTFn:       Schema, // ASTPageFn takes priority over Fn
    })
}

// Schema returns a typed AST node tree.
// CompileStage calls ast.CompileTree(node) to produce the final Schema.
func Schema(sess ui.UISessionContext) any {
    return screens.CRUDScreen(screens.CRUDScreenConfig{
        Title:     "Invoices",
        API:       "get:/api/v1/finance/invoices",
        CanCreate: sess.Can("create", "invoice"),
        CanEdit:   sess.Can("update", "invoice"),
        CanDelete: sess.Can("delete", "invoice"),
        Columns: []screens.Column{
            {Name: "number",   Label: "Invoice #", Sortable: true},
            {Name: "supplier", Label: "Supplier"},
            {Name: "amount",   Label: "Amount",    Type: "currency", Align: "right"},
            {Name: "status",   Label: "Status",    Map: statusBadgeMap()},
            {Name: "due_date", Label: "Due",       Type: "date", Sortable: true},
        },
    })
}
```

### When to use `ASTPageFn`

- New pages (always prefer)
- Pages with complex conditional column sets
- Pages where you want compile-time structural guarantees
- When migrating from `PageFn`

---

## Legacy Approach: `PageFn` + raw `amis.*` builders

Still fully supported. Use for:
- Existing pages that have not been migrated
- Simple one-off pages where typed AST overhead is not worth it
- When you need AMIS node types not yet in the DSL

### Registration

```go
// internal/web/pages/dashboard/schema.go
package dashboard

import (
    "awo.so/internal/web/amis"
    "awo.so/internal/web/registry"
    "awo.so/internal/web/ui"
)

func init() {
    registry.RegisterPage(registry.PageRegistration{
        Route:  "/dashboard",
        Module: "core",
        Title:  "Dashboard",
        Fn:     Schema, // PageFn — legacy path
    })
}

func Schema(sess ui.UISessionContext) ui.Schema {
    return amis.M{
        "type":  "page",
        "title": "Dashboard",
        "data": amis.M{
            "can_view_finance": sess.Can("read", "invoice"),
            "is_platform":      sess.IsPlatform,
            "currency":         sess.Currency,
        },
        "body": amis.M{
            "type": "grid",
            "columns": amis.A{
                amis.M{"md": 6, "body": invoiceSummaryPanel(sess)},
                amis.M{"md": 6, "body": quickActionsPanel(sess)},
            },
        },
    }
}
```

---

## `amis` Builder Package Reference

Source: `internal/web/amis/`

### Page builders (`amis/page.go`)

| Function | AMIS `type` | Key methods |
|----------|------------|-------------|
| `amis.Page(title)` | `page` | implicit via `map[string]any` |
| `amis.Grid(cols...)` | `grid` | accepts column `M` objects |
| `amis.Panel(title)` | `panel` | |
| `amis.Tabs()` | `tabs` | |
| `amis.Chart(api)` | `chart` | always transparent background |
| `amis.Alert(level, body)` | `alert` | |
| `amis.Tpl(str)` | `tpl` | |

### CRUD builders (`amis/crud.go`)

| Function | Purpose |
|----------|---------|
| `amis.CRUD(api)` | List/table — always sets `syncLocation: true` |
| `amis.Column(name, label)` | `.Type()`, `.Sortable()`, `.Map()`, `.Buttons()` |
| `amis.CreateBtn(label, api, fields...)` | Toolbar create → dialog |
| `amis.EditBtn(api, fields...)` | Row edit → dialog |
| `amis.ViewBtn(body)` | Row view → drawer |
| `amis.DeleteBtn(api)` | Row delete + confirmation |

### Form builders (`amis/form.go`)

| Function | AMIS type |
|----------|-----------|
| `amis.TextField(name, label)` | `input-text` |
| `amis.NumberField(name, label)` | `input-number` |
| `amis.DateField(name, label)` | `input-date` |
| `amis.SelectField(name, label, opts...)` | `select` |
| `amis.SelectAPIField(name, label, api)` | `select` with API source |
| `amis.SwitchField(name, label)` | `switch` |
| `amis.FileField(name, label, api)` | `input-file` |
| `amis.HiddenField(name)` | `hidden` |

Field modifiers (take a field `M`, return `M`):

```go
amis.Required(field)                // marks required
amis.Optional(field)                // adds "(optional)" remark
amis.Placeholder(field, text)       // sets placeholder
amis.Default(field, val)            // sets default value
amis.VisibleOn(field, expr)         // conditional + clearValueOnHidden
amis.DisabledOn(field, expr)        // conditional disabled
amis.Validate(field, "isEmail")     // validation rule
```

---

## Permission-Gating Patterns

### Pattern 1: Expose as AMIS template variable (recommended)

```go
func Schema(sess ui.UISessionContext) ui.Schema {
    return amis.M{
        "type": "page",
        "data": amis.M{
            "can_create": sess.Can("create", "invoice"),
            "can_delete": sess.Can("delete", "invoice"),
        },
        "body": amis.M{
            "type": "crud",
            // ...
            "headerToolbar": amis.A{
                amis.M{
                    "type":      "button",
                    "label":     "New Invoice",
                    "visibleOn": "${can_create}", // AMIS hides if false
                },
            },
            "columns": amis.A{
                // ...
                amis.M{
                    "type":      "operation",
                    "label":     "Actions",
                    "visibleOn": "${can_delete || can_create}",
                    "buttons": amis.A{
                        amis.M{
                            "type":        "button",
                            "label":       "Delete",
                            "level":       "danger",
                            "visibleOn":   "${can_delete}",
                        },
                    },
                },
            },
        },
    }
}
```

### Pattern 2: Structural branching in Go (for significant layout differences)

```go
func Schema(sess ui.UISessionContext) ui.Schema {
    var body []any
    body = append(body, invoiceTable(sess))

    if sess.Can("approve", "invoice") {
        body = append(body, approvalInboxBlock(sess))
    }

    if sess.IsPlatform {
        body = append(body, crossTenantSummary(sess))
    }

    return amis.M{
        "type": "page",
        "body": body,
    }
}
```

**Trade-off:** Pattern 1 produces schemas cacheable across all users with the same role set. Pattern 2 produces schemas conditional on the boolean result — but since permissions are part of the cache fingerprint, both patterns are correctly cached.

---

## AMIS Expression Gotchas

### Return value from page function

**Page functions return raw schema. Never return `AmisResponse`.**

```go
// WRONG — produces nested envelope {status:0, data:{status:0, data:{...}}}
func Schema(sess ui.UISessionContext) ui.Schema {
    return amis.M{
        "status": 0,
        "data": amis.M{
            "type": "page",
            // ...
        },
    }
}

// CORRECT — SchemaHandler wraps it automatically
func Schema(sess ui.UISessionContext) ui.Schema {
    return amis.M{
        "type": "page",
        // ...
    }
}
```

### CRUD syncLocation

All `crud` schemas must have `syncLocation: true`. `NormalizeStage` enforces this. The `amis.CRUD()` builder sets it automatically. If constructing `crud` manually:

```go
amis.M{
    "type":         "crud",
    "api":          "get:/api/v1/finance/invoices",
    "syncLocation": true, // required — NormalizeStage enforces this
}
```

### Chart background

Chart schemas must have `"style": {"background": "transparent"}`. The `amis.Chart()` builder sets it. If constructing manually, add it.

### API prefix

All API strings in schemas must start with `get:`, `post:`, `put:`, `delete:`, `patch:`. `NormalizeStage` or `ValidateStage` will reject bare URLs.

---

## Adding a New Page: Full Checklist

1. Create `internal/web/pages/<module>/<page>/schema.go`
2. Write `func Schema(sess ui.UISessionContext) any` (for ASTPageFn) or `ui.Schema` (for PageFn)
3. Register in `init()` with `registry.RegisterPage(registry.PageRegistration{...})`
4. Add any new permissions to `AllUIPermissions` in `internal/web/authz/service.go`
5. Run `ValidateRegistry()` check: startup will panic on duplicate routes or missing required fields
6. Schema is served at `/schema/<module>/<page>` — no changes to routing code needed

See [Page Registration Pattern](02-page-registration-pattern.md) for the complete registration guide.
