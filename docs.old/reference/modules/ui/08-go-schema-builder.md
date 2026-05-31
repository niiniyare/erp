[<-- Back to Index](README.md)

> **📌 TWO APPROACHES EXIST.** This document describes the `amis.Ctx` / `SchemaFn` approach — the
> original, still-supported pattern. A newer approach uses `UISessionContext` + `ASTPageFn` +
> `PageRegistration`, which provides full IAM integration, typed AST nodes, and pipeline caching.
>
> | | Legacy (`amis.Ctx`) | Modern (`UISessionContext` + `ASTPageFn`) |
> |--|--|--|
> | Type safety | Partial | Full (typed AST nodes) |
> | IAM integration | Manual `Can()` closure | Pre-resolved by `AuthzStage` |
> | Cache support | None | Redis + fingerprint cache |
> | Permission key format | `Can("create", "invoice")` | `sess.Can("create", "invoice")` → checks `"invoice.create"` |
> | Registration | `registry.Register(path, fn)` | `registry.RegisterPage(PageRegistration{...})` |
>
> **New pages should use the modern approach.**
> See [Page Registration Pattern](../03-implementation/02-page-registration-pattern.md) and
> [Pipeline Deep Dive](../02-architecture/02-pipeline-deep-dive.md).

## Go Schema Builder

### Why a Builder Package

Raw `map[string]any` is unreadable and has no compile-time checks. The `amis` package provides a fluent Go API that marshals directly to valid AMIS JSON — no `.Build()` call needed when passing to Fiber's `c.JSON()`.

```
WITHOUT builder (unreadable):
  map[string]any{"type":"crud","api":"get:/api/v1/orders","syncLocation":true,"columns":[]any{...}}

WITH builder (clear intent):
  amis.CRUD("get:/api/v1/orders").
      Columns(amis.Column("id","ID"), amis.Column("status","Status")).
      Toolbar(amis.CreateBtn("New Order", "post:/api/v1/orders", ...))
```

---

### Package Structure

```
internal/web/
├── amis/
│   ├── builder.go     — Core types: M, A, Schema, Ctx, CtxUser, SchemaFn, base
│   ├── page.go        — Page, Service, Grid, Panel, Tabs, Chart, Timeline, Descriptions, Alert, Tpl
│   ├── crud.go        — CRUD, Column, EditBtn, ViewBtn, DeleteBtn, CreateBtn
│   ├── form.go        — Form, Wizard, all field helpers, field modifiers
│   └── app.go         — App shell builder + NavPage, NavLink, NavGroup
├── registry/
│   └── registry.go    — Register(path, fn) + Get(path)
├── handler/
│   └── schema.go      — SchemaHandler — one Fiber handler for all schemas
└── pages/
    └── dashboard/
        └── schema.go  — Example: dashboard schema function
```

---

### Core Types

```go
// M is a schema map — alias for map[string]any
type M = map[string]any

// A is a schema array — alias for []any
type A = []any

// Schema is the root type every page function returns
type Schema = M

// Ctx carries request-scoped data into schema functions.
// Schema functions are pure: same Ctx → same JSON.
type Ctx struct {
    Flags map[string]bool
    User  CtxUser
    Can   func(action, resource string) bool
}

// SchemaFn is the signature every page builder function must satisfy.
type SchemaFn func(ctx Ctx) Schema
```

---

### Writing a Page Schema

Every page is a Go function with signature `func(ctx amis.Ctx) amis.Schema`.

```go
// internal/web/pages/finance/invoices/schema.go
package invoices

import (
    "awo.so/internal/web/amis"
    "awo.so/internal/web/registry"
)

func init() {
    registry.Register("/finance/invoices", Schema)
}

func Schema(ctx amis.Ctx) amis.Schema {
    return amis.Page("Invoices").
        Data(amis.M{
            "can_create": ctx.Can("create", "invoice"),
        }).
        Body(
            amis.CRUD("get:/api/v1/finance/invoices").
                Columns(
                    amis.Column("number", "Invoice #").Sortable(),
                    amis.Column("supplier", "Supplier"),
                    amis.Column("amount", "Amount").Type("currency").Align("right"),
                    amis.Column("status", "Status").Map(statusMap()),
                    amis.Column("due_date", "Due").Type("date").Sortable(),
                    amis.Column("", "").Buttons(
                        amis.ViewBtn(detailDrawer()),
                        amis.EditBtn("put:/api/v1/finance/invoices/${id}", editFields()...),
                        amis.DeleteBtn("delete:/api/v1/finance/invoices/${id}"),
                    ),
                ).
                Toolbar(
                    amis.M{"visibleOn": "${can_create}",
                        "type": "button", "label": "New Invoice",
                        "level": "primary", "actionType": "dialog",
                        "dialog": amis.M{"title": "New Invoice",
                            "body": amis.Form("post:/api/v1/finance/invoices").
                                Fields(editFields()...).Build(),
                        },
                    },
                ).
                Build(),
        ).
        Build()
}
```

---

### Registering Pages

Pages self-register via `init()`. Import them blank in the route setup:

```go
// internal/api/handlers/routes.go (imports section)
import (
    _ "awo.so/internal/web/pages/dashboard"
    _ "awo.so/internal/web/pages/finance/invoices"
    _ "awo.so/internal/web/pages/finance/accounts"
    // adding a new page = add one import line here
)
```

No other changes needed in `web/`. The schema is automatically served at `/schema/finance/invoices`.

---

### Builder Reference

#### Page builders (`amis/page.go`)

| Function | AMIS type | Key methods |
|---|---|---|
| `Page(title)` | `page` | `.Body()`, `.Data()`, `.Toolbar()`, `.Aside()` |
| `Service(api)` | `service` | `.Body()`, `.SchemaAPI()` |
| `Grid(cols...)` | `grid` | `.Gap()` |
| `Col(width, body)` | grid column | — (returns M) |
| `Panel(title)` | `panel` | `.Body()`, `.Footer()`, `.ClassName()` |
| `Tabs()` | `tabs` | `.Tab(title, body)`, `.Mode()` |
| `Chart(api)` | `chart` | `.Config(M)`, `.Height()` — always transparent bg |
| `Timeline(api)` | `timeline` | `.Items(...)` |
| `Descriptions(title)` | `descriptions` | `.Item(label, name)`, `.Columns(n)` |
| `Alert(level, body)` | `alert` | `.VisibleOn()` |
| `Tpl(str)` | `tpl` | returns M directly |

#### CRUD builders (`amis/crud.go`)

| Function | Purpose |
|---|---|
| `CRUD(api)` | List/table — always has `syncLocation: true` |
| `Column(name, label)` | `.Type()`, `.Align()`, `.Sortable()`, `.Tpl()`, `.Map()`, `.Buttons()` |
| `CreateBtn(label, api, fields...)` | Toolbar create button → dialog |
| `EditBtn(api, fields...)` | Row edit button → dialog |
| `ViewBtn(body)` | Row view button → drawer |
| `DeleteBtn(api)` | Row delete with confirmation |

#### Form builders (`amis/form.go`)

| Function | AMIS type |
|---|---|
| `Form(api)` | `form` |
| `Wizard(api)` | `wizard` with `.Step()` and `.ReviewStep()` |
| `TextField(name, label)` | `input-text` |
| `NumberField(name, label)` | `input-number` |
| `DateField(name, label)` | `input-date` |
| `SelectField(name, label, opts...)` | `select` |
| `SelectAPIField(name, label, api)` | `select` with API source |
| `SwitchField(name, label)` | `switch` |
| `FileField(name, label, api)` | `input-file` |
| `HiddenField(name)` | `hidden` |

Field modifiers (take a field M, return M):

```go
Required(field)               // marks required
Optional(field)               // adds "(optional)" remark
Placeholder(field, text)      // sets placeholder
Default(field, val)           // sets default value
VisibleOn(field, expr)        // conditional + clearValueOnHidden
DisabledOn(field, expr)       // conditional disabled
Validate(field, "isEmail")    // validation rule
```

#### App shell (`amis/app.go`)

```go
App("Awo ERP").
    Logo("/public/logo.svg").
    Header(Tpl("${tenant_name}"), M{"type": "theme-toggle"}).
    Pages(
        NavGroup("Finance", "fa fa-calculator",
            NavLink("Invoices", "fa fa-file", "/finance/invoices", "/schema/finance/invoices"),
            NavLink("Accounts", "fa fa-bank", "/finance/accounts", "/schema/finance/accounts"),
        ),
        NavLink("Dashboard", "fa fa-chart-line", "/dashboard", "/schema/dashboard"),
    ).
    Build()
```

---

### How Schemas Reach the Browser

```
Browser hash changes to #finance/invoices
  → index.html: fetch('/schema/finance/invoices')
    → Fiber: GET /schema/finance/invoices → SchemaHandler.Handle()
      → registry.Get("/finance/invoices") → invoices.Schema(ctx)
        → returns amis.Schema (map[string]any)
          → JSON: { status: 0, data: { type: "page", ... } }
            → index.html: amis.embed(el, schema.data, {}, amisEnv)
```

---

### Adding a New Page (Checklist)

1. Create `internal/web/pages/<module>/<page>/schema.go`
2. Write `func Schema(ctx amis.Ctx) amis.Schema { ... }`
3. Add `registry.Register("/<module>/<page>", Schema)` in `init()`
4. Add `_ "awo.so/internal/web/pages/<module>/<page>"` to routes.go imports
5. Done — `/schema/<module>/<page>` is live. No changes to `web/`.
