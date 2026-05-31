[<-- Back to Index](README.md)

> **🚫 SUPERSEDED.** This document describes a design specification that has since been implemented
> differently. The directory structure, type names (`UIContext`), and compilation model here do not
> match the current codebase.
>
> **Read instead:**
> - [Pipeline Deep Dive](../02-architecture/02-pipeline-deep-dive.md) — actual 9-stage pipeline
> - [Page Registration Pattern](../03-implementation/02-page-registration-pattern.md) — how to register pages
> - [IAM Integration](../02-architecture/03-iam-integration.md) — `UISessionContext` and permission resolution
>
> This file is kept as an **Architecture Decision Record** showing the original design intent.

# Go-First UI Compiler — Production Architecture

Go functions are the source code of the UI.
AMIS JSON is the compiled artifact.
The frontend is a dumb renderer.

---

## Mental Model

```
Go UI Functions  →  UI Compiler  →  AMIS JSON  →  Browser Renderer
      ↑                  ↑
  UIContext          Registry
  (user, flags,     (path → PageFn)
   permissions)
```

The compiler is not a template engine. It is a function evaluator: call the registered Go function with the request context, get back a JSON object. The frontend has no logic. It fetches JSON and renders it.

---

## Directory Structure

```
internal/web/
├── ui/
│   ├── context.go       — UIContext, ContextFromFiber()
│   ├── types.go         — Schema, PageFn, NavFn
│   └── permissions.go   — Permission key helpers
├── amis/
│   ├── types.go         — M, A aliases
│   ├── page.go          — Page, Service, Grid, Panel, Tabs, Alert, Tpl
│   ├── crud.go          — CRUD, Column, action buttons
│   ├── form.go          — Form, Wizard, all field builders + modifiers
│   ├── action.go        — Button, ButtonGroup, onEvent builder
│   └── nav.go           — NavGroup, NavLink (used by app schema fn)
├── registry/
│   ├── pages.go         — Register(path, PageFn) + Lookup(path)
│   └── nav.go           — RegisterNav(NavFn) + AllNav()
├── compiler/
│   └── compiler.go      — Compile(path, ctx) Schema
├── handler/
│   └── schema.go        — One Fiber handler for all /schema/* routes
└── pages/
    ├── app/
    │   └── schema.go    — /schema/app — full nav shell
    ├── dashboard/
    │   └── schema.go
    └── finance/
        ├── module.go    — nav contribution for Finance group
        ├── accounts/
        │   └── schema.go
        └── invoices/
            └── schema.go
```

---

## Layer 1 — UIContext

`UIContext` is the only input to every page function.
Same context → same JSON. Page functions are pure.

```go
// internal/web/ui/context.go
package ui

import (
    "github.com/gofiber/fiber/v2"
    "awo.so/internal/middleware"
)

// UIContext carries all request-scoped data into schema functions.
// It is immutable after construction.
type UIContext struct {
    UserID   string
    TenantID string

    // RBAC
    Roles []string

    // Pre-resolved ABAC: "action:resource" → allowed
    // Built by middleware from Casbin evaluation.
    // Schema functions read this — they never call Casbin directly.
    Permissions map[string]bool

    // Feature flags resolved for this tenant
    FeatureFlags map[string]bool

    // User preferences from profile
    Preferences map[string]any

    // Locale for label/format decisions
    Locale string

    // Tenant config (branding, currency, timezone)
    TenantConfig TenantConfig
}

type TenantConfig struct {
    Name     string
    Currency string
    Timezone string
    LogoURL  string
}

// Can returns true if the current user has permission for action on resource.
// Use this in schema functions to gate UI elements.
// The API must enforce the same permission independently.
func (ctx UIContext) Can(action, resource string) bool {
    return ctx.Permissions[action+":"+resource]
}

// HasRole returns true if the user holds the given role.
func (ctx UIContext) HasRole(role string) bool {
    for _, r := range ctx.Roles {
        if r == role {
            return true
        }
    }
    return false
}

// Flag returns true if the feature flag is enabled for this tenant.
func (ctx UIContext) Flag(name string) bool {
    return ctx.FeatureFlags[name]
}

// Pref returns a user preference value, with a fallback default.
func (ctx UIContext) Pref(key string, defaultVal any) any {
    if v, ok := ctx.Preferences[key]; ok {
        return v
    }
    return defaultVal
}

// ContextFromFiber builds UIContext from a Fiber request context.
// Call this in the schema handler after all middleware has run.
func ContextFromFiber(c *fiber.Ctx) UIContext {
    user    := middleware.ContextUser(c)
    tenant  := middleware.ContextTenant(c)
    flags   := middleware.ContextFlags(c)
    prefs   := middleware.ContextPreferences(c)
    perms   := middleware.ContextPermissions(c)

    return UIContext{
        UserID:   user.ID,
        TenantID: tenant.ID,
        Roles:    user.Roles,
        Permissions: perms,
        FeatureFlags: flags,
        Preferences:  prefs,
        Locale:  user.Locale,
        TenantConfig: TenantConfig{
            Name:     tenant.Name,
            Currency: tenant.Currency,
            Timezone: tenant.Timezone,
            LogoURL:  tenant.LogoURL,
        },
    }
}
```

---

## Layer 2 — Schema Types

```go
// internal/web/ui/types.go
package ui

// Schema is the root type all page functions return.
// It marshals directly to AMIS-compatible JSON.
type Schema = map[string]any

// PageFn is the contract every page schema function must satisfy.
// Pure function: same UIContext → same Schema.
type PageFn func(ctx UIContext) Schema

// NavFn contributes nav items to the app shell.
// Returns nil if the module is disabled for this context.
type NavFn func(ctx UIContext) []NavItem

// NavItem represents one entry in the sidebar.
type NavItem struct {
    Label    string
    Icon     string
    URL      string       // hash route: "/finance/invoices"
    Children []NavItem    // if non-empty, renders as NavGroup
}
```

---

## Layer 3 — AMIS Builder Package

### `amis/types.go`

```go
// internal/web/amis/types.go
package amis

// M is a schema map — alias for map[string]any.
// Used everywhere in schema construction.
type M = map[string]any

// A is a schema array — alias for []any.
type A = []any
```

### `amis/page.go`

```go
// internal/web/amis/page.go
package amis

// pageBuilder builds a "page" component.
type pageBuilder struct {
    m M
}

func Page(title string) *pageBuilder {
    return &pageBuilder{m: M{"type": "page", "title": title}}
}

func (b *pageBuilder) Body(body ...any) *pageBuilder {
    if len(body) == 1 {
        b.m["body"] = body[0]
    } else {
        b.m["body"] = body
    }
    return b
}

// Data injects values into the page data domain.
// Child components access these via ${key} expressions.
// Use this to inject pre-resolved permissions and tenant config —
// never inline business logic in AMIS expressions.
func (b *pageBuilder) Data(data M) *pageBuilder {
    b.m["data"] = data
    return b
}

func (b *pageBuilder) Toolbar(items ...any) *pageBuilder {
    b.m["toolbar"] = items
    return b
}

func (b *pageBuilder) Aside(body any) *pageBuilder {
    b.m["aside"] = body
    return b
}

func (b *pageBuilder) Build() M { return b.m }

// ── Service ──────────────────────────────────────────────────────────

type serviceBuilder struct{ m M }

// Service loads data from an API and injects it into child components.
// Use instead of crud when you need summary data, not a paginated list.
func Service(api string) *serviceBuilder {
    return &serviceBuilder{m: M{"type": "service", "api": api}}
}

// SchemaAPI makes the service load its body schema from Go dynamically.
// The Go endpoint returns raw AMIS JSON (not wrapped in envelope).
func (b *serviceBuilder) SchemaAPI(api string) *serviceBuilder {
    b.m["schemaApi"] = api
    return b
}

func (b *serviceBuilder) Body(body ...any) *serviceBuilder {
    if len(body) == 1 {
        b.m["body"] = body[0]
    } else {
        b.m["body"] = body
    }
    return b
}

// Polling configures interval-based refresh (ms). Stops when expr is true.
func (b *serviceBuilder) Polling(intervalMs int, stopWhen string) *serviceBuilder {
    b.m["interval"]              = intervalMs
    b.m["silentPolling"]         = true
    b.m["stopAutoRefreshWhen"]   = stopWhen
    return b
}

func (b *serviceBuilder) Build() M { return b.m }

// ── Grid ──────────────────────────────────────────────────────────────

// Grid renders children in a column grid.
func Grid(columns ...M) M {
    return M{"type": "grid", "columns": columns}
}

// Col wraps a body component in a grid column with Bootstrap width class.
func Col(widthClass string, body any) M {
    return M{"columnClassName": widthClass, "body": body}
}

// ── Panel ─────────────────────────────────────────────────────────────

type panelBuilder struct{ m M }

func Panel(title string) *panelBuilder {
    return &panelBuilder{m: M{"type": "panel", "title": title}}
}

func (b *panelBuilder) Body(body any) *panelBuilder  { b.m["body"] = body; return b }
func (b *panelBuilder) Footer(f any) *panelBuilder   { b.m["footer"] = f; return b }
func (b *panelBuilder) ClassName(c string) *panelBuilder { b.m["className"] = c; return b }
func (b *panelBuilder) Build() M { return b.m }

// ── Tabs ──────────────────────────────────────────────────────────────

type tabsBuilder struct {
    m    M
    tabs A
}

func Tabs() *tabsBuilder {
    return &tabsBuilder{m: M{"type": "tabs"}}
}

// Tab adds a tab. Set lazy=true to mount only on first click.
func (b *tabsBuilder) Tab(title string, body any, lazy bool) *tabsBuilder {
    tab := M{"title": title, "body": body}
    if lazy {
        tab["mountOnEnter"]  = true
        tab["unmountOnExit"] = false
    }
    b.tabs = append(b.tabs, tab)
    return b
}

// Mode sets visual style: "line" (default), "card", "radio", "vertical".
func (b *tabsBuilder) Mode(mode string) *tabsBuilder { b.m["mode"] = mode; return b }

func (b *tabsBuilder) Build() M {
    b.m["tabs"] = b.tabs
    return b.m
}

// ── Chart ─────────────────────────────────────────────────────────────

type chartBuilder struct{ m M }

// Chart creates an ECharts component.
// backgroundColor is always transparent — required by Decision 8.
func Chart(api string) *chartBuilder {
    return &chartBuilder{m: M{
        "type": "chart",
        "api":  api,
        "config": M{"backgroundColor": "transparent"},
    }}
}

func (b *chartBuilder) Config(cfg M) *chartBuilder {
    cfg["backgroundColor"] = "transparent" // enforce Decision 8
    b.m["config"] = cfg
    return b
}

func (b *chartBuilder) Height(h int) *chartBuilder { b.m["height"] = h; return b }
func (b *chartBuilder) Build() M { return b.m }

// ── Alert ─────────────────────────────────────────────────────────────

type alertBuilder struct{ m M }

// Alert renders an info/warning/danger/success banner.
func Alert(level, body string) *alertBuilder {
    return &alertBuilder{m: M{"type": "alert", "level": level, "body": body}}
}

func (b *alertBuilder) VisibleOn(expr string) *alertBuilder { b.m["visibleOn"] = expr; return b }
func (b *alertBuilder) Build() M { return b.m }

// ── Tpl ───────────────────────────────────────────────────────────────

// Tpl renders a template string with ${expr} interpolation.
func Tpl(template string) M {
    return M{"type": "tpl", "tpl": template}
}

// Stat renders a KPI stat card.
func Stat(source, label string) M {
    return M{"type": "stat", "source": source, "label": label}
}

// Descriptions renders a key-value description list (detail view).
type descriptionsBuilder struct{ m M; items A }

func Descriptions(title string) *descriptionsBuilder {
    return &descriptionsBuilder{m: M{"type": "descriptions", "title": title}}
}

func (b *descriptionsBuilder) Item(label, name string) *descriptionsBuilder {
    b.items = append(b.items, M{"label": label, "name": name})
    return b
}

func (b *descriptionsBuilder) Columns(n int) *descriptionsBuilder { b.m["columns"] = n; return b }
func (b *descriptionsBuilder) Build() M {
    b.m["items"] = b.items
    return b.m
}
```

### `amis/crud.go`

```go
// internal/web/amis/crud.go
package amis

// crudBuilder builds a "crud" component (paginated table with filters).
type crudBuilder struct {
    m       M
    columns A
    toolbar A
    filter  A
    bulk    A
}

// CRUD creates a server-side paginated, filterable, sortable table.
// syncLocation: true is always set — filter/page state persists in URL.
func CRUD(api string) *crudBuilder {
    return &crudBuilder{m: M{
        "type":         "crud",
        "api":          api,
        "syncLocation": true,
        "headerToolbar": A{
            "search-box",
            "bulkActions",
            M{"type": "columns-toggler"},
            M{"type": "reload"},
            M{"type": "export-csv"},
        },
        "footerToolbar": A{"statistics", "pagination"},
    }}
}

func (b *crudBuilder) ID(id string) *crudBuilder        { b.m["id"] = id; return b }
func (b *crudBuilder) PerPage(n int) *crudBuilder       { b.m["perPage"] = n; return b }
func (b *crudBuilder) DefaultSort(col, dir string) *crudBuilder {
    b.m["orderBy"]  = col
    b.m["orderDir"] = dir
    return b
}

func (b *crudBuilder) Columns(cols ...M) *crudBuilder {
    b.columns = append(b.columns, toAny(cols)...)
    return b
}

func (b *crudBuilder) Toolbar(items ...any) *crudBuilder {
    b.toolbar = append(b.toolbar, items...)
    return b
}

func (b *crudBuilder) Filter(fields ...M) *crudBuilder {
    b.filter = append(b.filter, toAny(fields)...)
    return b
}

func (b *crudBuilder) BulkActions(actions ...M) *crudBuilder {
    b.bulk = append(b.bulk, toAny(actions)...)
    return b
}

// EmptyText sets the message shown when there are no results.
// Two variants: noData (no records exist) and noFilter (filters exclude all).
func (b *crudBuilder) EmptyText(noData, noFilter string) *crudBuilder {
    b.m["placeholder"] = M{
        "empty": noData,
    }
    // Shown when filters are active and return nothing
    b.m["filterEmptyText"] = noFilter
    return b
}

func (b *crudBuilder) Build() M {
    b.m["columns"] = b.columns
    if len(b.toolbar) > 0 {
        // Prepend to default headerToolbar
        existing := b.m["headerToolbar"].(A)
        b.m["headerToolbar"] = append(b.toolbar, existing...)
    }
    if len(b.filter) > 0 {
        b.m["filter"] = M{
            "body":           b.filter,
            "submitOnChange": false,
        }
    }
    if len(b.bulk) > 0 {
        b.m["bulkActions"] = b.bulk
    }
    return b.m
}

// ── Column ────────────────────────────────────────────────────────────

type columnBuilder struct{ m M }

func Column(name, label string) *columnBuilder {
    return &columnBuilder{m: M{"name": name, "label": label}}
}

func (c *columnBuilder) Type(t string) *columnBuilder      { c.m["type"] = t; return c }
func (c *columnBuilder) Align(a string) *columnBuilder     { c.m["align"] = a; return c }
func (c *columnBuilder) Sortable() *columnBuilder          { c.m["sortable"] = true; return c }
func (c *columnBuilder) Width(w int) *columnBuilder        { c.m["width"] = w; return c }
func (c *columnBuilder) Fixed(side string) *columnBuilder  { c.m["fixed"] = side; return c }

// Tpl sets a template expression for the column cell.
// Use for computed displays: "${amount | number:2} ${currency}"
func (c *columnBuilder) Tpl(tpl string) *columnBuilder {
    c.m["type"] = "tpl"
    c.m["tpl"]  = tpl
    return c
}

// Map sets a value→label/badge mapping for status columns.
func (c *columnBuilder) Map(mapping M) *columnBuilder {
    c.m["type"]    = "mapping"
    c.m["map"]     = mapping
    return c
}

// ColorMap sets type=tag with color-coded status values.
func (c *columnBuilder) ColorMap(mapping M) *columnBuilder {
    c.m["type"]     = "tag"
    c.m["colorMap"] = mapping
    return c
}

// Buttons sets action buttons on an operation column.
func (c *columnBuilder) Buttons(btns ...M) *columnBuilder {
    c.m["type"]    = "operation"
    c.m["buttons"] = btns
    return c
}

func (c *columnBuilder) Build() M { return c.m }

// ── Action Buttons ────────────────────────────────────────────────────

// ViewBtn opens a drawer with the given body.
func ViewBtn(body any) M {
    return M{
        "type":       "button",
        "label":      "View",
        "level":      "link",
        "actionType": "drawer",
        "drawer": M{
            "size": "lg",
            "body": body,
        },
    }
}

// EditBtn opens a dialog with a pre-filled form for editing.
func EditBtn(api string, fields ...M) M {
    return M{
        "type":       "button",
        "label":      "Edit",
        "level":      "link",
        "actionType": "dialog",
        "dialog": M{
            "title": "Edit",
            "body": Form(api).
                Fields(fields...).
                InitAPI("get:" + extractPath(api) + "/${id}").
                Build(),
        },
    }
}

// DeleteBtn shows a confirmation dialog then calls DELETE.
func DeleteBtn(api string) M {
    return M{
        "type":        "button",
        "label":       "Delete",
        "level":       "link",
        "className":   "text-danger",
        "actionType":  "ajax",
        "api":         api,
        "confirmText": "Delete this record? This cannot be undone.",
        "onEvent": M{
            "success": M{
                "actions": A{
                    M{"actionType": "reload", "componentId": "list"},
                    M{"actionType": "toast", "args": M{"msg": "Deleted", "msgType": "success"}},
                },
            },
        },
    }
}

// CreateBtn adds a toolbar button that opens a create form dialog.
func CreateBtn(label, api string, fields ...M) M {
    return M{
        "type":       "button",
        "label":      label,
        "level":      "primary",
        "actionType": "dialog",
        "dialog": M{
            "title": label,
            "body":  Form(api).Fields(fields...).Build(),
        },
    }
}

// AjaxBtn creates an inline action button that calls an API directly.
func AjaxBtn(label, level, api, confirmText string) M {
    btn := M{
        "type":       "button",
        "label":      label,
        "level":      level,
        "actionType": "ajax",
        "api":        api,
    }
    if confirmText != "" {
        btn["confirmText"] = confirmText
    }
    return btn
}

// VisibleWhen adds a visibleOn condition to any M (button, column, field).
// RULE: expr must only reference boolean values pre-injected into page data.
// Never encode business logic in the expression — Go resolves it, AMIS reads it.
func VisibleWhen(m M, boolKey string) M {
    m["visibleOn"] = "${" + boolKey + "}"
    return m
}

// DisabledWhen adds a disabledOn condition to any M.
func DisabledWhen(m M, boolKey string) M {
    m["disabledOn"] = "${" + boolKey + "}"
    return m
}
```

### `amis/form.go`

```go
// internal/web/amis/form.go
package amis

// formBuilder builds a "form" component.
type formBuilder struct {
    m      M
    fields A
}

// Form creates a form that submits to the given API.
func Form(api string) *formBuilder {
    return &formBuilder{m: M{
        "type":   "form",
        "api":    api,
        "mode":   "horizontal",
        "reload": "list",
    }}
}

// InitAPI sets the API to pre-populate the form (for edit forms).
func (b *formBuilder) InitAPI(api string) *formBuilder { b.m["initApi"] = api; return b }

// Mode sets form layout: "horizontal" (default), "vertical", "inline".
func (b *formBuilder) Mode(mode string) *formBuilder { b.m["mode"] = mode; return b }

// Redirect navigates to a URL after successful submit.
// Use ${id} to reference the created/updated record ID.
func (b *formBuilder) Redirect(url string) *formBuilder { b.m["redirect"] = url; return b }

func (b *formBuilder) Fields(fields ...M) *formBuilder {
    b.fields = append(b.fields, toAny(fields)...)
    return b
}

func (b *formBuilder) Build() M {
    b.m["body"] = b.fields
    return b.m
}

// ── Wizard ────────────────────────────────────────────────────────────

type wizardBuilder struct {
    m     M
    steps A
}

// Wizard creates a multi-step form. Always add a ReviewStep as the last step.
func Wizard(api string) *wizardBuilder {
    return &wizardBuilder{m: M{"type": "wizard", "api": api}}
}

func (b *wizardBuilder) Step(title string, fields ...M) *wizardBuilder {
    b.steps = append(b.steps, M{"title": title, "body": toAny(fields)})
    return b
}

// ReviewStep adds a standard final review step.
func (b *wizardBuilder) ReviewStep() *wizardBuilder {
    return b.Step("Review", M{
        "type":    "tpl",
        "tpl":     "Please review your entries before submitting.",
        "wrapperComponent": "",
    })
}

func (b *wizardBuilder) Build() M {
    b.m["steps"] = b.steps
    return b.m
}

// ── Field Builders ────────────────────────────────────────────────────

func TextField(name, label string) M {
    return M{"type": "input-text", "name": name, "label": label}
}

func TextareaField(name, label string) M {
    return M{"type": "textarea", "name": name, "label": label}
}

func NumberField(name, label string) M {
    return M{"type": "input-number", "name": name, "label": label}
}

func AmountField(name, label, currency string) M {
    return M{
        "type":      "input-number",
        "name":      name,
        "label":     label,
        "prefix":    currency + " ",
        "precision":  2,
        "min":        0,
    }
}

func DateField(name, label string) M {
    return M{"type": "input-date", "name": name, "label": label, "format": "YYYY-MM-DD"}
}

func DateRangeField(name, label string) M {
    return M{"type": "input-date-range", "name": name, "label": label}
}

func SelectField(name, label string, opts ...M) M {
    return M{"type": "select", "name": name, "label": label, "options": opts, "clearable": true}
}

// Opt creates a single select option.
func Opt(label, value string) M {
    return M{"label": label, "value": value}
}

// SelectAPIField loads options from an API endpoint.
func SelectAPIField(name, label, api string) M {
    return M{
        "type":  "select",
        "name":  name,
        "label": label,
        "source": api,
        "clearable": true,
    }
}

func SwitchField(name, label string) M {
    return M{"type": "switch", "name": name, "label": label}
}

func FileField(name, label, uploadAPI string) M {
    return M{
        "type":      "input-file",
        "name":      name,
        "label":     label,
        "uploadType": "fileReceptor",
        "receiver":   uploadAPI,
    }
}

func HiddenField(name string) M {
    return M{"type": "hidden", "name": name}
}

// ── Field Modifiers ───────────────────────────────────────────────────
// Each modifier takes a field M and returns M — chain with compose.

func Required(field M) M {
    field["required"] = true
    return field
}

func Optional(field M) M {
    field["remark"] = "(optional)"
    return field
}

func Placeholder(field M, text string) M {
    field["placeholder"] = text
    return field
}

func DefaultVal(field M, val any) M {
    field["value"] = val
    return field
}

// ShowWhen adds conditional visibility. boolKey must exist in page data.
// Always set clearValueOnHidden to avoid stale values in submit payload.
func ShowWhen(field M, boolKey string) M {
    field["visibleOn"]          = "${" + boolKey + "}"
    field["clearValueOnHidden"] = true
    return field
}

// DisableWhen adds conditional disabled state. boolKey must exist in page data.
func DisableWhen(field M, boolKey string) M {
    field["disabledOn"] = "${" + boolKey + "}"
    return field
}

func Validate(field M, rule string) M {
    field["validations"]        = M{rule: true}
    field["validationErrors"]   = M{rule: "Invalid value"}
    return field
}

func ValidateMsg(field M, rule, msg string) M {
    field["validations"]      = M{rule: true}
    field["validationErrors"] = M{rule: msg}
    return field
}
```

### `amis/nav.go`

```go
// internal/web/amis/nav.go
package amis

import "awo.so/internal/web/ui"

// BuildAppSchema constructs the AMIS "app" schema for the nav shell.
// Go filters nav items by UIContext before returning — the browser
// receives only the items the user is allowed to see.
func BuildAppSchema(ctx ui.UIContext, navItems []ui.NavItem) M {
    return M{
        "type":      "app",
        "brandName": ctx.TenantConfig.Name,
        "logo":      ctx.TenantConfig.LogoURL,
        "header": A{
            Tpl("${tenant_name}"),
            M{"type": "theme-toggle"}, // custom renderer registered in browser
        },
        "pages": buildNavPages(navItems),
        "data": M{
            "tenant_name": ctx.TenantConfig.Name,
            "currency":    ctx.TenantConfig.Currency,
            "locale":      ctx.Locale,
            "user_id":     ctx.UserID,
        },
    }
}

func buildNavPages(items []ui.NavItem) A {
    pages := make(A, 0, len(items))
    for _, item := range items {
        if len(item.Children) == 0 {
            pages = append(pages, M{
                "label":  item.Label,
                "icon":   item.Icon,
                "url":    item.URL,
                "schema": M{"type": "service", "schemaApi": "/schema" + item.URL},
            })
        } else {
            pages = append(pages, M{
                "label":    item.Label,
                "icon":     item.Icon,
                "children": buildNavPages(item.Children),
            })
        }
    }
    return pages
}
```

---

## Layer 4 — Registry

```go
// internal/web/registry/pages.go
package registry

import (
    "fmt"
    "sync"
    "awo.so/internal/web/ui"
)

var pages = &pageRegistry{m: make(map[string]ui.PageFn)}

type pageRegistry struct {
    mu sync.RWMutex
    m  map[string]ui.PageFn
}

// Register maps a URL path to a page schema function.
// Call from init() in each page package.
// Panics on duplicate registration to catch errors at startup.
func Register(path string, fn ui.PageFn) {
    pages.mu.Lock()
    defer pages.mu.Unlock()
    if _, exists := pages.m[path]; exists {
        panic(fmt.Sprintf("ui/registry: duplicate page registration for path %q", path))
    }
    pages.m[path] = fn
}

// Lookup returns the PageFn for a path, or nil if not registered.
func Lookup(path string) ui.PageFn {
    pages.mu.RLock()
    defer pages.mu.RUnlock()
    return pages.m[path]
}

// All returns all registered paths (for debugging/listing).
func All() []string {
    pages.mu.RLock()
    defer pages.mu.RUnlock()
    paths := make([]string, 0, len(pages.m))
    for p := range pages.m {
        paths = append(paths, p)
    }
    return paths
}
```

```go
// internal/web/registry/nav.go
package registry

import (
    "sync"
    "awo.so/internal/web/ui"
)

var navContributors struct {
    mu  sync.RWMutex
    fns []ui.NavFn
}

// RegisterNav adds a nav contributor.
// Each module registers its NavFn here via init().
// The app schema handler calls all contributors, filters by context,
// and builds the full nav tree.
func RegisterNav(fn ui.NavFn) {
    navContributors.mu.Lock()
    defer navContributors.mu.Unlock()
    navContributors.fns = append(navContributors.fns, fn)
}

// AllNav returns all registered NavFns.
func AllNav() []ui.NavFn {
    navContributors.mu.RLock()
    defer navContributors.mu.RUnlock()
    out := make([]ui.NavFn, len(navContributors.fns))
    copy(out, navContributors.fns)
    return out
}
```

---

## Layer 5 — Compiler

```go
// internal/web/compiler/compiler.go
package compiler

import (
    "fmt"
    "awo.so/internal/web/amis"
    "awo.so/internal/web/registry"
    "awo.so/internal/web/ui"
)

// ErrNotFound is returned when no page is registered for a path.
type ErrNotFound struct{ Path string }
func (e ErrNotFound) Error() string {
    return fmt.Sprintf("no page registered for path %q", e.Path)
}

// Compile resolves the page function for path, calls it with ctx,
// and returns the AMIS Schema. Returns ErrNotFound if path is unknown.
func Compile(path string, ctx ui.UIContext) (ui.Schema, error) {
    fn := registry.Lookup(path)
    if fn == nil {
        return nil, ErrNotFound{Path: path}
    }
    return fn(ctx), nil
}

// CompileApp builds the full app shell schema with nav filtered by ctx.
func CompileApp(ctx ui.UIContext) ui.Schema {
    contributors := registry.AllNav()
    var items []ui.NavItem

    for _, fn := range contributors {
        contributed := fn(ctx)
        items = append(items, contributed...)
    }

    return amis.BuildAppSchema(ctx, items)
}
```

---

## Layer 6 — Schema Handler

```go
// internal/web/handler/schema.go
package handler

import (
    "errors"
    "strings"

    "github.com/gofiber/fiber/v2"
    "awo.so/internal/web/compiler"
    "awo.so/internal/web/ui"
)

type SchemaHandler struct{}

// Handle is the single Fiber handler for ALL /schema/* routes.
// Wire it once; new pages register themselves via init() imports.
func (h *SchemaHandler) Handle(c *fiber.Ctx) error {
    // Build UIContext from middleware-populated Fiber locals
    ctx := ui.ContextFromFiber(c)

    // Extract path: /schema/finance/invoices → /finance/invoices
    path := strings.TrimPrefix(c.Path(), "/schema")
    if path == "" || path == "/app" {
        // Special case: app shell with full nav
        schema := compiler.CompileApp(ctx)
        return c.JSON(schema)
    }

    schema, err := compiler.Compile(path, ctx)
    if err != nil {
        var notFound compiler.ErrNotFound
        if errors.As(err, &notFound) {
            return c.Status(404).JSON(fiber.Map{
                "type":  "page",
                "title": "Not Found",
                "body": fiber.Map{
                    "type":  "alert",
                    "level": "warning",
                    "body":  "Page not found: " + notFound.Path,
                },
            })
        }
        return c.Status(500).JSON(fiber.Map{"status": 1, "msg": "Schema compilation failed"})
    }

    // Schema endpoints return raw AMIS JSON — NOT wrapped in {status, data}.
    // AMIS renders the returned object directly.
    return c.JSON(schema)
}
```

---

## Layer 7 — Route Registration

```go
// internal/api/routes.go (schema section)
package api

import (
    "github.com/gofiber/fiber/v2"
    "awo.so/internal/middleware"
    "awo.so/internal/web/handler"

    // Each blank import triggers init() → registry.Register()
    // This is the ONLY place module pages are wired in.
    _ "awo.so/internal/web/pages/app"
    _ "awo.so/internal/web/pages/dashboard"
    _ "awo.so/internal/web/pages/finance"           // imports module.go (nav)
    _ "awo.so/internal/web/pages/finance/accounts"
    _ "awo.so/internal/web/pages/finance/invoices"
    _ "awo.so/internal/web/pages/finance/journals"
    _ "awo.so/internal/web/pages/procurement"
    _ "awo.so/internal/web/pages/procurement/orders"
)

func RegisterSchemaRoutes(app *fiber.App, schemaH *handler.SchemaHandler) {
    schema := app.Group("/schema",
        middleware.ResolveTenant,
        middleware.Authenticate,
        middleware.InjectFlags,
        middleware.InjectPermissions, // pre-resolves all permissions into Fiber locals
    )
    schema.Get("/*", schemaH.Handle)
}
```

---

## Layer 8 — Module Pattern

A module is a Go package with exactly two responsibilities:
1. `module.go` — registers the nav contribution
2. `<page>/schema.go` — registers page schemas

### Module Nav Registration

```go
// internal/web/pages/finance/module.go
package finance

import (
    "awo.so/internal/web/registry"
    "awo.so/internal/web/ui"
)

func init() {
    registry.RegisterNav(Nav)
}

// Nav contributes the Finance group to the app nav.
// Returns nil items for any sub-page the user cannot access.
// The entire group is omitted if no children survive filtering.
func Nav(ctx ui.UIContext) []ui.NavItem {
    var children []ui.NavItem

    // Each child is added only if the user has read permission
    if ctx.Can("read", "account") {
        children = append(children, ui.NavItem{
            Label: "Chart of Accounts",
            Icon:  "fa fa-sitemap",
            URL:   "/finance/accounts",
        })
    }
    if ctx.Can("read", "invoice") {
        children = append(children, ui.NavItem{
            Label: "Invoices",
            Icon:  "fa fa-file-invoice",
            URL:   "/finance/invoices",
        })
    }
    if ctx.Can("read", "journal") && ctx.Flag("gl.module") {
        children = append(children, ui.NavItem{
            Label: "Journal Entries",
            Icon:  "fa fa-book",
            URL:   "/finance/journals",
        })
    }

    if len(children) == 0 {
        return nil // entire Finance group hidden
    }

    return []ui.NavItem{{
        Label:    "Finance",
        Icon:     "fa fa-calculator",
        Children: children,
    }}
}
```

### Page Schema

```go
// internal/web/pages/finance/invoices/schema.go
package invoices

import (
    "awo.so/internal/web/amis"
    "awo.so/internal/web/registry"
    "awo.so/internal/web/ui"
)

func init() {
    registry.Register("/finance/invoices", Schema)
}

// Schema builds the Invoices list page.
// All conditional logic is resolved here in Go.
// The frontend receives a JSON object with no branching.
func Schema(ctx ui.UIContext) ui.Schema {
    // Resolve all permissions upfront.
    // These become boolean values in page data.
    // AMIS reads ${can_create} etc — never evaluates business logic.
    canCreate  := ctx.Can("create", "invoice")
    canExport  := ctx.Can("export", "invoice")
    canDelete  := ctx.Can("delete", "invoice")
    currency   := ctx.TenantConfig.Currency

    // Build columns — omit columns the user cannot see
    cols := []amis.M{
        amis.Column("number", "Invoice #").Sortable().Fixed("left").Build(),
        amis.Column("supplier_name", "Supplier").Sortable().Build(),
        amis.Column("amount", "Amount").
            Tpl(currency + " ${amount | number:2}").
            Align("right").Sortable().Build(),
        amis.Column("status", "Status").
            ColorMap(amis.M{
                "DRAFT":    "default",
                "OPEN":     "processing",
                "PAID":     "success",
                "OVERDUE":  "error",
                "VOIDED":   "warning",
            }).Build(),
        amis.Column("due_date", "Due Date").Type("date").Sortable().Build(),
        amis.Column("created_at", "Created").Type("date").Sortable().Build(),
        buildActions(canDelete),
    }

    // Toolbar: export button only if permitted
    toolbar := []amis.M{}
    if canCreate {
        toolbar = append(toolbar, amis.CreateBtn("New Invoice", "post:/api/v1/finance/invoices",
            buildFormFields(currency)...,
        ))
    }
    if canExport {
        toolbar = append(toolbar, amis.M{
            "type": "button", "label": "Export", "level": "default",
            "actionType": "ajax", "api": "post:/api/v1/finance/invoices/export",
        })
    }

    crud := amis.CRUD("get:/api/v1/finance/invoices").
        ID("invoices-list").
        PerPage(20).
        DefaultSort("due_date", "asc").
        Columns(cols...).
        Filter(
            amis.TextField("keywords", "Search"),
            amis.SelectField("status", "Status",
                amis.Opt("Draft", "DRAFT"),
                amis.Opt("Open", "OPEN"),
                amis.Opt("Overdue", "OVERDUE"),
                amis.Opt("Paid", "PAID"),
            ),
            amis.DateRangeField("date_range", "Date Range"),
        ).
        EmptyText(
            "No invoices yet. Create your first invoice.",
            "No invoices match the current filters. Clear filters to see all.",
        ).
        Build()

    if len(toolbar) > 0 {
        crud = amis.CRUD("get:/api/v1/finance/invoices").
            ID("invoices-list").
            PerPage(20).
            DefaultSort("due_date", "asc").
            Columns(cols...).
            Toolbar(toAny(toolbar)...).
            Filter(
                amis.TextField("keywords", "Search"),
                amis.SelectField("status", "Status",
                    amis.Opt("Draft", "DRAFT"),
                    amis.Opt("Open", "OPEN"),
                    amis.Opt("Overdue", "OVERDUE"),
                    amis.Opt("Paid", "PAID"),
                ),
                amis.DateRangeField("date_range", "Date Range"),
            ).
            EmptyText(
                "No invoices yet. Create your first invoice.",
                "No invoices match the current filters.",
            ).
            Build()
    }

    return amis.Page("Invoices").
        Body(crud).
        Build()
}

func buildActions(canDelete bool) amis.M {
    btns := []amis.M{
        amis.ViewBtn(detailDrawer()),
        amis.EditBtn("put:/api/v1/finance/invoices/${id}", buildFormFields("")...),
        amis.AjaxBtn("Void", "link", "post:/api/v1/finance/invoices/${id}/void",
            "Void this invoice? This cannot be undone."),
    }
    if canDelete {
        btns = append(btns, amis.DeleteBtn("delete:/api/v1/finance/invoices/${id}"))
    }
    return amis.Column("", "Actions").Buttons(btns...).Fixed("right").Build()
}

func buildFormFields(currency string) []amis.M {
    if currency == "" {
        currency = "KES"
    }
    return []amis.M{
        amis.Required(amis.SelectAPIField("supplier_id", "Supplier", "get:/api/v1/suppliers?perPage=100")),
        amis.Required(amis.DateField("invoice_date", "Invoice Date")),
        amis.Required(amis.DateField("due_date", "Due Date")),
        amis.Required(amis.AmountField("amount", "Amount", currency)),
        amis.Optional(amis.TextareaField("notes", "Notes")),
    }
}

func detailDrawer() amis.M {
    return amis.M{
        "type":    "service",
        "api":     "get:/api/v1/finance/invoices/${id}",
        "body": amis.Descriptions("Invoice Details").
            Item("Invoice #", "number").
            Item("Supplier", "supplier_name").
            Item("Amount", "amount").
            Item("Status", "status").
            Item("Due Date", "due_date").
            Columns(2).
            Build(),
    }
}

func toAny(ms []amis.M) []any {
    out := make([]any, len(ms))
    for i, m := range ms { out[i] = m }
    return out
}
```

---

## Layer 9 — Permission Middleware

The schema handler needs all permissions pre-resolved before calling `ContextFromFiber`. This middleware runs Casbin for every resource/action pair the UI might need, and stores results in Fiber locals.

```go
// internal/middleware/permissions.go
package middleware

import (
    "github.com/gofiber/fiber/v2"
    "awo.so/internal/core/access"
)

// InjectPermissions pre-resolves all UI-relevant permissions into Fiber locals.
// Schema functions read these via UIContext.Can() — they never call Casbin.
//
// Add new entries here as new resources are introduced.
// Format: "action:resource"
var uiPermissions = []struct{ action, resource string }{
    {"read",   "account"},
    {"create", "account"},
    {"update", "account"},
    {"read",   "invoice"},
    {"create", "invoice"},
    {"update", "invoice"},
    {"delete", "invoice"},
    {"export", "invoice"},
    {"read",   "journal"},
    {"create", "journal"},
    {"post",   "journal"},
    {"read",   "purchase_order"},
    {"create", "purchase_order"},
    {"approve","purchase_order"},
    // ... extend as modules are added
}

func InjectPermissions(svc access.Service) fiber.Handler {
    return func(c *fiber.Ctx) error {
        user   := ContextUser(c)
        tenant := ContextTenant(c)
        perms  := make(map[string]bool, len(uiPermissions))

        for _, p := range uiPermissions {
            allowed, _ := svc.Can(c.Context(), tenant.ID, user.ID, p.action, p.resource)
            perms[p.action+":"+p.resource] = allowed
        }

        c.Locals("permissions", perms)
        return c.Next()
    }
}

func ContextPermissions(c *fiber.Ctx) map[string]bool {
    if p, ok := c.Locals("permissions").(map[string]bool); ok {
        return p
    }
    return map[string]bool{}
}
```

---

## Adding a New Module — Exact Steps

This is the only workflow developers need to follow.

```
1. Create internal/web/pages/<module>/module.go
   - func init() { registry.RegisterNav(Nav) }
   - func Nav(ctx ui.UIContext) []ui.NavItem { ... }

2. Create internal/web/pages/<module>/<page>/schema.go
   - func init() { registry.Register("/<module>/<page>", Schema) }
   - func Schema(ctx ui.UIContext) ui.Schema { ... }

3. Add two blank imports to internal/api/routes.go:
   _ "awo.so/internal/web/pages/<module>"
   _ "awo.so/internal/web/pages/<module>/<page>"

4. Add new permissions to InjectPermissions middleware if needed.

5. Done. /schema/<module>/<page> is live.
   Nav appears automatically for users with read permission.
   No frontend changes required.
```

---

## Invariants — Non-Negotiable Rules

These are enforced by code review and linting. Violations break the architecture.

### I1: PageFn is pure

```go
// CORRECT: same ctx → same schema
func Schema(ctx ui.UIContext) ui.Schema {
    canCreate := ctx.Can("create", "invoice")
    return amis.Page("Invoices").
        Body(buildCRUD(canCreate)).Build()
}

// WRONG: fetching data inside schema function
func Schema(ctx ui.UIContext) ui.Schema {
    count, _ := db.CountInvoices(ctx.TenantID) // NEVER
    ...
}
```

Schema functions must not call databases, external services, or I/O. They receive context, produce JSON. Data is fetched by AMIS at render time via `api` properties in the schema.

### I2: No business logic in AMIS expressions

```go
// CORRECT: Go resolves, AMIS reads a boolean
canApprove := ctx.Can("approve", "purchase_order")
// In schema: visibleOn: "${can_approve}"

// WRONG: AMIS evaluates business logic
// In schema: visibleOn: "${user.role === 'FINANCE_MANAGER' && amount > 10000}"
```

If you see a `visibleOn` expression that does anything beyond reading a `${boolean_key}`, it is wrong. Business logic belongs in Go.

### I3: Permissions enforced at two layers

Go schema function gates UI elements (UX layer).
Go API handler enforces the same permission (security layer).
These must never diverge.

```go
// Schema function
canCreate := ctx.Can("create", "invoice")
// → toolbar create button visible only if canCreate

// API handler
func (h *InvoiceHandler) Create(c *fiber.Ctx) error {
    if !h.access.Can(c.Context(), tenantID, userID, "create", "invoice") {
        return c.Status(403).JSON(response.Err("Forbidden"))
    }
    // ...
}
```

### I4: Feature flags resolved in Go only

```go
// CORRECT: Go gates entire module nav and schema
func Nav(ctx ui.UIContext) []ui.NavItem {
    if !ctx.Flag("payroll.module") {
        return nil // entire module hidden
    }
    // ...
}

// WRONG: feature flag check in schema JSON
// "visibleOn": "${featureFlags.payroll_enabled}"
```

### I5: No inline styles or hardcoded colours in schemas

```go
// CORRECT: semantic class or CSS variable reference
amis.M{"className": "text-danger"}

// WRONG: literal colour
amis.M{"style": "color: #ff0000"}
```

---

## Performance Considerations

### Schema Caching

Page schemas are deterministic: same UIContext → same JSON. For high-traffic deployments, schemas can be cached by a composite key:

```go
// compiler/compiler.go — with optional caching layer
func cacheKey(path string, ctx ui.UIContext) string {
    // Cache key: path + permission fingerprint + flag fingerprint
    // NOT user ID — permissions are the actual differentiator
    return path + "|" + permFingerprint(ctx.Permissions) + "|" + flagFingerprint(ctx.FeatureFlags)
}
```

Invalidate cache entries when: permissions change for a role, feature flags are toggled for a tenant.

Most schemas will have fewer than 10 distinct cache keys per path (combinations of permission sets per role × flag variants). This is extremely cacheable.

### Permission Pre-Resolution Cost

`InjectPermissions` runs Casbin for N permission checks per schema request. At 30 permissions, this is 30 Casbin calls. Casbin with a loaded policy evaluates in microseconds. Total overhead: < 1ms for the permission batch.

If this becomes a bottleneck: batch the Casbin calls into a single `FilterEnforce` call if the Casbin adapter supports it, or cache the permission map per (userID, tenantID) with a short TTL (30s).

---

## What This Architecture Achieves

| Goal | How |
|---|---|
| Go owns all UI logic | PageFn is the only place UI is defined |
| No frontend conditionals | All branching in Go before JSON is emitted |
| Feature flag gating | NavFn returns nil → module vanishes from nav |
| RBAC/ABAC gating | Can() in PageFn → boolean keys in page data |
| Adding a module | 3 files + 2 import lines. No other changes. |
| 100+ modules | Registry pattern; no central handler list |
| Schema caching | Deterministic functions → cache by permission+flag fingerprint |
| Compile-time safety | Builder API; raw `map[string]any` only at leaf level |
| AMIS philosophy | Declarative JSON out; data chain for state; no custom renderers for standard use cases |
