[<-- Back to Index](README.md)

> **⚠️ STATUS: IMPLEMENTED — NOT A TODO LIST.**
> This document was written as an implementation task checklist. All core phases are now **complete**.
> Reading it as a roadmap will give a false picture of current state.
>
> The pipeline described here (SessionContext, permission resolver, cache, validator) exists in
> `internal/web/stages/` and `internal/web/ui/pipeline.go`.
>
> **For the actual implemented pipeline, read:**
> - [Pipeline Deep Dive](../02-architecture/02-pipeline-deep-dive.md) — authoritative reference
>
> This file is kept as **historical context** showing how the pipeline was designed.

## UI Services Implementation Pipeline — Zero → Hero

Reference architecture: [§28 Go-First UI Compiler](28-go-ui-compiler.md) · [§29 IAM/Session Compiler](29-iam-session-compiler.md)

Each task is: actionable, testable, commit-ready, verifiable, dependency-aware, sequentially executable.

---

## PHASE 1 — Core Foundation

### TASK 1.1: Define SessionContext and builder

**Description:** Create the `SessionContext` struct and its `Builder` — the single entry point for all UI compilation. This is the foundation every other phase depends on.

**Why:** Without a typed, immutable session carrier, page functions will access permissions ad-hoc (map lookups, type assertions, race conditions). Every downstream task assumes this contract.

**Implementation:**

```
internal/web/session/
├── context.go     — SessionContext struct + exported accessors
└── builder.go     — Builder: flucts, Validate(), Build()
```

`context.go`:
```go
package session

import "time"

type SessionContext struct {
    UserID      string
    TenantID    string
    Email       string
    DisplayName string
    Locale      string
    Timezone    string
    Currency    string
    AuthMethod  string
    ExpiresAt   time.Time

    permissions map[string]bool   // private — access via Can()
    featureFlags map[string]bool  // private — access via Flag()
}

func (s SessionContext) Can(action, resource string) bool {
    return s.permissions[resource+"."+action]
}

func (s SessionContext) CanAny(action string, resources ...string) bool {
    for _, r := range resources {
        if s.Can(action, r) {
            return true
        }
    }
    return false
}

func (s SessionContext) CanAll(action string, resources ...string) bool {
    for _, r := range resources {
        if !s.Can(action, r) {
            return false
        }
    }
    return true
}

func (s SessionContext) Flag(name string) bool {
    return s.featureFlags[name]
}

func (s SessionContext) Authenticated() bool {
    return s.UserID != "" && time.Now().Before(s.ExpiresAt)
}

func (s SessionContext) PermissionFingerprint() string {
    // sorted keys → stable hash
    return fingerprintMap(s.permissions)
}

func (s SessionContext) FlagFingerprint() string {
    return fingerprintMap(s.featureFlags)
}
```

`builder.go`:
```go
package session

import (
    "errors"
    "time"
)

type Builder struct {
    ctx SessionContext
}

func New() *Builder { return &Builder{} }

func (b *Builder) User(id, email, name string) *Builder {
    b.ctx.UserID = id
    b.ctx.Email = email
    b.ctx.DisplayName = name
    return b
}

func (b *Builder) Tenant(id string) *Builder  { b.ctx.TenantID = id; return b }
func (b *Builder) Locale(l string) *Builder   { b.ctx.Locale = l; return b }
func (b *Builder) Timezone(z string) *Builder { b.ctx.Timezone = z; return b }
func (b *Builder) Currency(c string) *Builder { b.ctx.Currency = c; return b }
func (b *Builder) Auth(method string, expires time.Time) *Builder {
    b.ctx.AuthMethod = method
    b.ctx.ExpiresAt = expires
    return b
}

func (b *Builder) Permissions(perms map[string]bool) *Builder {
    b.ctx.permissions = perms
    return b
}

func (b *Builder) Flags(flags map[string]bool) *Builder {
    b.ctx.featureFlags = flags
    return b
}

func (b *Builder) Build() (SessionContext, error) {
    if b.ctx.UserID == "" {
        return SessionContext{}, errors.New("session: UserID required")
    }
    if b.ctx.TenantID == "" {
        return SessionContext{}, errors.New("session: TenantID required")
    }
    if b.ctx.ExpiresAt.IsZero() {
        return SessionContext{}, errors.New("session: ExpiresAt required")
    }
    return b.ctx, nil
}
```

**Expected Behavior:**
- `sess.Can("read", "invoice")` returns true iff `permissions["invoice.read"] == true`
- `sess.Authenticated()` returns false after expiry
- `Build()` returns error on missing required fields
- `PermissionFingerprint()` returns same hash for two sessions with identical permission sets

**Test Cases:**
```go
// Can() — positive
sess := buildTestSession(map[string]bool{"invoice.read": true})
assert(sess.Can("read", "invoice") == true)

// Can() — negative
assert(sess.Can("delete", "invoice") == false)

// CanAny()
assert(sess.CanAny("read", "invoice", "payment") == true)

// Expired session
expired := buildExpiredSession()
assert(expired.Authenticated() == false)

// Builder validation
_, err := New().Build()
assert(err != nil)

// Fingerprint stability
s1 := buildTestSession(map[string]bool{"a": true, "b": false})
s2 := buildTestSession(map[string]bool{"b": false, "a": true})
assert(s1.PermissionFingerprint() == s2.PermissionFingerprint())
```

**Acceptance Criteria:**
- [ ] `SessionContext` compiles with no exported permission/flag map fields
- [ ] All accessor methods tested
- [ ] `Build()` validates all required fields
- [ ] `PermissionFingerprint()` is stable (same input → same hash, deterministic sort)
- [ ] No imports from `internal/iam` or `internal/policy` — session package must be leaf

**Risk:** Fingerprint instability breaks cache sharing. Use sorted keys before hashing, not map iteration order.

**Commit:** `feat(ui/session): add SessionContext and builder`

---

### TASK 1.2: Define PageFn type and schema types

**Description:** Declare the canonical Go types for all UI compilation: `PageFn`, `NavFn`, `UIBlock`, `Schema`, `M`, `A`. These are the contracts all page authors and registry consumers depend on.

**Why:** Prevents type drift. If `PageFn` is redefined per-package, the registry cannot enforce a uniform signature.

**Implementation:**

```
internal/web/types.go
```

```go
package web

import "awo.so/internal/web/session"

// M is a schema map — shorthand for map[string]any.
type M = map[string]any

// A is a schema array — shorthand for []any.
type A = []any

// Schema is the root value every page function returns.
type Schema = M

// PageFn is the signature all page builder functions must satisfy.
// Pure function: same SessionContext → same Schema.
type PageFn func(sess session.SessionContext) Schema

// NavFn builds the navigation tree for the app shell.
type NavFn func(sess session.SessionContext) []M

// UIBlock is a reusable session-aware schema fragment.
type UIBlock func(sess session.SessionContext) M
```

**Expected Behavior:**
- All packages in `internal/web/` import `web.PageFn`, not define their own
- Type aliases (`M`, `A`) mean zero-cost: no wrapping, no conversion

**Test Cases:** Type compatibility — compile-time only.

**Acceptance Criteria:**
- [ ] File compiles
- [ ] All downstream packages use `web.PageFn`, not local redefinitions
- [ ] `M` and `A` are aliases (not new types) — assignable without conversion

**Risk:** None. Pure declaration.

**Commit:** `feat(ui): define canonical PageFn, NavFn, UIBlock, M, A types`

---

### TASK 1.3: Implement fingerprint utility

**Description:** Implement `fingerprintMap(map[string]bool) string` used by `SessionContext.PermissionFingerprint()` and `FlagFingerprint()`.

**Why:** Cache correctness depends entirely on fingerprint stability. Two sessions with identical permission sets must produce identical fingerprints — regardless of insertion order.

**Implementation:**

```
internal/web/session/fingerprint.go
```

```go
package session

import (
    "crypto/sha256"
    "fmt"
    "sort"
    "strings"
)

func fingerprintMap(m map[string]bool) string {
    keys := make([]string, 0, len(m))
    for k, v := range m {
        if v {
            keys = append(keys, k)
        }
    }
    sort.Strings(keys)
    h := sha256.Sum256([]byte(strings.Join(keys, ",")))
    return fmt.Sprintf("%x", h[:8]) // 16-char hex — collision probability negligible for <10k keys
}
```

**Expected Behavior:**
- `fingerprintMap({"a":true,"b":false})` == `fingerprintMap({"b":false,"a":true})`
- Only `true` values included — false permissions don't differentiate sessions
- Returns 16-char hex string

**Test Cases:**
```go
assert(fingerprintMap(map[string]bool{"x":true,"y":true}) ==
       fingerprintMap(map[string]bool{"y":true,"x":true}))

assert(fingerprintMap(map[string]bool{"x":true}) !=
       fingerprintMap(map[string]bool{"x":true,"z":true}))

assert(fingerprintMap(map[string]bool{"x":false}) ==
       fingerprintMap(map[string]bool{})) // false = absent
```

**Acceptance Criteria:**
- [ ] Deterministic across goroutines (no map iteration races)
- [ ] Empty map returns consistent non-empty string
- [ ] Only true-valued entries affect the fingerprint

**Risk:** If false entries were included, a user gaining a new *denied* entry would bust cache unnecessarily.

**Commit:** `feat(ui/session): add stable permission/flag fingerprint`

---

## PHASE 2 — Registry System

### TASK 2.1: Implement page registry

**Description:** Build `registry.Register(path, PageFn)` and `registry.Get(path)` — the lookup table connecting URL paths to schema builder functions.

**Why:** Without a registry, the compiler has no way to dispatch `/schema/finance/invoices` to the right Go function. The registry is what makes `init()` self-registration possible.

**Implementation:**

```
internal/web/registry/registry.go
```

```go
package registry

import (
    "fmt"
    "sync"

    "awo.so/internal/web"
)

var (
    mu    sync.RWMutex
    pages = make(map[string]web.PageFn)
    navFn web.NavFn
)

// Register associates a URL path with a page builder function.
// Called from init() in each page package.
func Register(path string, fn web.PageFn) {
    mu.Lock()
    defer mu.Unlock()
    if _, exists := pages[path]; exists {
        panic(fmt.Sprintf("registry: duplicate page registration for path %q", path))
    }
    pages[path] = fn
}

// Get returns the PageFn for the given path, or nil if not found.
func Get(path string) web.PageFn {
    mu.RLock()
    defer mu.RUnlock()
    return pages[path]
}

// RegisterNav sets the global navigation builder.
// Only one NavFn is allowed; second call panics.
func RegisterNav(fn web.NavFn) {
    mu.Lock()
    defer mu.Unlock()
    if navFn != nil {
        panic("registry: NavFn already registered")
    }
    navFn = fn
}

// GetNav returns the registered NavFn, or nil.
func GetNav() web.NavFn {
    mu.RLock()
    defer mu.RUnlock()
    return navFn
}

// List returns all registered paths — for testing and debugging.
func List() []string {
    mu.RLock()
    defer mu.RUnlock()
    paths := make([]string, 0, len(pages))
    for p := range pages {
        paths = append(paths, p)
    }
    return paths
}
```

**Expected Behavior:**
- `Register` panics on duplicate path — catches copy-paste errors at startup, not runtime
- `Get` returns nil (not error) for unknown path — caller decides 404 vs fallback
- Thread-safe: concurrent `Get` from request goroutines is safe after init phase

**Test Cases:**
```go
// Register and retrieve
registry.Register("/test", testFn)
fn := registry.Get("/test")
assert(fn != nil)

// Duplicate panics
defer func() { assert(recover() != nil) }()
registry.Register("/test", testFn) // should panic

// Unknown path
assert(registry.Get("/nonexistent") == nil)
```

**Acceptance Criteria:**
- [ ] Thread-safe under concurrent reads
- [ ] Duplicate registration panics at startup (not silently overwrites)
- [ ] `List()` returns all registered paths

**Risk:** `init()` order is undefined across packages. Registry must be initialized before any `Register` call. The `sync.RWMutex` + pre-initialized map handles this.

**Commit:** `feat(ui/registry): add page and nav registry with init-time self-registration`

---

### TASK 2.2: Implement module pattern helpers

**Description:** Add `module.go` convention documentation and a `RegisterModule(prefix, navFn)` helper that lets a module register both its nav entries and page-level nav group in one call.

**Why:** Without a pattern, each module author invents their own `init()` structure. Modules need to contribute both a `NavFn` slice and individual `PageFn` registrations. This task standardizes that.

**Implementation:**

Each module follows this structure:
```
internal/web/pages/finance/
├── module.go      — RegisterNav entries for this module
├── invoices/
│   └── schema.go  — Register("/finance/invoices", Schema)
└── accounts/
    └── schema.go  — Register("/finance/accounts", Schema)
```

`internal/web/pages/finance/module.go`:
```go
package finance

import (
    "awo.so/internal/web/amis"
    "awo.so/internal/web/registry"
    "awo.so/internal/web/session"

    _ "awo.so/internal/web/pages/finance/invoices"
    _ "awo.so/internal/web/pages/finance/accounts"
)

func init() {
    registry.RegisterNavGroup(navGroup)
}

func navGroup(sess session.SessionContext) amis.NavGroupSchema {
    links := []amis.NavLinkSchema{}

    if sess.Can("read", "invoice") {
        links = append(links, amis.NavLink("Invoices", "fa fa-file-invoice",
            "/finance/invoices", "/schema/finance/invoices"))
    }
    if sess.Can("read", "account") {
        links = append(links, amis.NavLink("Accounts", "fa fa-landmark",
            "/finance/accounts", "/schema/finance/accounts"))
    }

    if len(links) == 0 {
        return amis.NavGroupSchema{Hidden: true}
    }
    return amis.NavGroup("Finance", "fa fa-calculator", links...)
}
```

Add to `registry.go`:
```go
type NavGroupFn func(sess session.SessionContext) amis.NavGroupSchema

var navGroups []NavGroupFn

func RegisterNavGroup(fn NavGroupFn) {
    mu.Lock()
    defer mu.Unlock()
    navGroups = append(navGroups, fn)
}

func GetNavGroups() []NavGroupFn {
    mu.RLock()
    defer mu.RUnlock()
    result := make([]NavGroupFn, len(navGroups))
    copy(result, navGroups)
    return result
}
```

**Expected Behavior:**
- Module's `init()` registers nav group + pulls in all page schemas via blank imports
- Nav groups with zero visible links return `Hidden: true` — compiler skips them
- Single blank import of the module package pulls the entire module

**Test Cases:**
```go
// Nav group hidden when no permissions
sess := buildSession(map[string]bool{})
group := financeNavGroup(sess)
assert(group.Hidden == true)

// Links filtered by permission
sess = buildSession(map[string]bool{"invoice.read": true})
group = financeNavGroup(sess)
assert(len(group.Links) == 1)
assert(group.Links[0].Label == "Invoices")
```

**Acceptance Criteria:**
- [ ] Single blank import of module package registers all pages and nav group
- [ ] Nav groups absent when user has no permissions for any link in group
- [ ] `registry.GetNavGroups()` returns all registered groups

**Risk:** Blank import chains (`finance` imports `invoices` which imports `registry`) must not create cycles. Keep `registry` as a leaf package with no imports from `internal/web/pages/`.

**Commit:** `feat(ui/registry): add nav group registry and module pattern`

---

## PHASE 3 — AMIS Builder Package

### TASK 3.1: Implement core builder types (M, A, base)

**Description:** Create `internal/web/amis/builder.go` with the `base` embedded struct that all builders share, plus the `Build()` method convention.

**Why:** Without a shared base, every builder reimplements `.VisibleOn()`, `.ClassName()`, `.ID()`, `.TestID()` — ten builders, ten inconsistent implementations.

**Implementation:**

```
internal/web/amis/builder.go
```

```go
package amis

// M and A are re-exported here for convenience in page packages.
type M = map[string]any
type A = []any

// base is embedded in all builders. Provides shared schema fields.
type base struct {
    m M
}

func newBase(typ string) base {
    return base{m: M{"type": typ}}
}

func (b *base) set(k string, v any) {
    b.m[k] = v
}

func (b base) VisibleOn(expr string) M {
    b.m["visibleOn"] = expr
    return b.m
}

func (b base) ClassName(cls string) M {
    b.m["className"] = cls
    return b.m
}

func (b base) Build() M {
    return b.m
}
```

**Expected Behavior:**
- All builders embed `base`
- `.Build()` returns the underlying `M` — identical to passing to `c.JSON()` directly
- Mutating the returned `M` after build does not affect the builder (map is shared — by design, schemas are compiled once per request)

**Acceptance Criteria:**
- [ ] All builders in subsequent tasks embed `base`
- [ ] `Build()` returns the same map on repeated calls (idempotent, not a copy)

**Risk:** None. Foundation only.

**Commit:** `feat(ui/amis): add base builder type`

---

### TASK 3.2: Implement page builders

**Description:** Implement `Page`, `Service`, `Grid`, `Col`, `Panel`, `Tabs`, `Chart`, `Timeline`, `Descriptions`, `Alert`, `Tpl` in `internal/web/amis/page.go`.

**Why:** These are the layout primitives used by every page schema. Without them, page authors write raw `map[string]any` — unreadable, unchecked, inconsistent.

**Implementation:**

```
internal/web/amis/page.go
```

```go
package amis

// --- Page ---

type PageBuilder struct {
    base
}

func Page(title string) *PageBuilder {
    b := &PageBuilder{newBase("page")}
    b.set("title", title)
    return b
}

func (b *PageBuilder) Body(items ...any) *PageBuilder {
    if len(items) == 1 {
        b.set("body", items[0])
    } else {
        b.set("body", items)
    }
    return b
}

func (b *PageBuilder) Data(d M) *PageBuilder     { b.set("data", d); return b }
func (b *PageBuilder) Toolbar(items ...any) *PageBuilder { b.set("toolbar", items); return b }
func (b *PageBuilder) Aside(body any) *PageBuilder { b.set("aside", body); return b }

// --- Service ---

type ServiceBuilder struct{ base }

func Service(api string) *ServiceBuilder {
    b := &ServiceBuilder{newBase("service")}
    b.set("api", api)
    return b
}

func (b *ServiceBuilder) Body(items ...any) *ServiceBuilder {
    b.set("body", items); return b
}

func (b *ServiceBuilder) SchemaAPI(api string) *ServiceBuilder {
    b.set("schemaApi", api); return b
}

// --- Grid ---

type GridBuilder struct{ base }

func Grid(cols ...M) *GridBuilder {
    b := &GridBuilder{newBase("grid")}
    b.set("columns", cols)
    return b
}

func (b *GridBuilder) Gap(g string) *GridBuilder { b.set("gap", g); return b }

// Col wraps a body in a grid column definition.
func Col(width int, body any) M {
    return M{"body": body, "md": width}
}

// --- Panel ---

type PanelBuilder struct{ base }

func Panel(title string) *PanelBuilder {
    b := &PanelBuilder{newBase("panel")}
    b.set("title", title)
    return b
}

func (b *PanelBuilder) Body(items ...any) *PanelBuilder { b.set("body", items); return b }
func (b *PanelBuilder) Footer(f any) *PanelBuilder      { b.set("footer", f); return b }
func (b *PanelBuilder) ClassName(c string) *PanelBuilder { b.set("className", c); return b }

// --- Tabs ---

type TabsBuilder struct{ base }

func Tabs() *TabsBuilder {
    return &TabsBuilder{newBase("tabs")}
}

func (b *TabsBuilder) Tab(title string, body any) *TabsBuilder {
    tabs, _ := b.m["tabs"].(A)
    tabs = append(tabs, M{"title": title, "body": body})
    b.set("tabs", tabs)
    return b
}

func (b *TabsBuilder) Mode(m string) *TabsBuilder { b.set("mode", m); return b }

// --- Chart ---

type ChartBuilder struct{ base }

func Chart(api string) *ChartBuilder {
    b := &ChartBuilder{newBase("chart")}
    b.set("api", api)
    b.set("style", M{"background": "transparent"}) // always transparent
    return b
}

func (b *ChartBuilder) Config(cfg M) *ChartBuilder { b.set("config", cfg); return b }
func (b *ChartBuilder) Height(h int) *ChartBuilder { b.set("height", h); return b }

// --- Alert ---

type AlertBuilder struct{ base }

func Alert(level, body string) *AlertBuilder {
    b := &AlertBuilder{newBase("alert")}
    b.set("level", level)
    b.set("body", body)
    return b
}

func (b *AlertBuilder) VisibleOn(expr string) *AlertBuilder {
    b.set("visibleOn", expr); return b
}

// --- Tpl ---

func Tpl(str string) M {
    return M{"type": "tpl", "tpl": str}
}

// --- Descriptions ---

type DescriptionsBuilder struct{ base }

func Descriptions(title string) *DescriptionsBuilder {
    b := &DescriptionsBuilder{newBase("descriptions")}
    b.set("title", title)
    return b
}

func (b *DescriptionsBuilder) Item(label, name string) *DescriptionsBuilder {
    items, _ := b.m["items"].(A)
    items = append(items, M{"label": label, "name": name})
    b.set("items", items)
    return b
}

func (b *DescriptionsBuilder) Columns(n int) *DescriptionsBuilder {
    b.set("columns", n); return b
}
```

**Expected Behavior:**
- `Page("Orders").Body(crud).Data(M{"x": 1}).Build()` → valid AMIS page schema
- `Chart()` always has `style.background = "transparent"` — enforced in constructor
- `Tabs().Tab("A", bodyA).Tab("B", bodyB).Build()` → `{type: "tabs", tabs: [{title:"A",...},{title:"B",...}]}`

**Test Cases:**
```go
p := amis.Page("Test").Body(amis.Tpl("hello")).Build()
assert(p["type"] == "page")
assert(p["title"] == "Test")

c := amis.Chart("/api/chart").Build()
style := c["style"].(amis.M)
assert(style["background"] == "transparent")
```

**Acceptance Criteria:**
- [ ] All builders return `*Builder` for chaining, `.Build()` returns `M`
- [ ] Chart always has transparent background (constructor, not caller responsibility)
- [ ] No panic on empty body calls

**Commit:** `feat(ui/amis): add page layout builders`

---

### TASK 3.3: Implement CRUD builder

**Description:** Implement `CRUD`, `Column`, `CreateBtn`, `EditBtn`, `ViewBtn`, `DeleteBtn` in `internal/web/amis/crud.go`.

**Why:** CRUD table is the most common ERP component. It has non-trivial defaults (syncLocation, button confirmation dialogs) that must be consistent across all pages.

**Implementation:**

```
internal/web/amis/crud.go
```

```go
package amis

// --- CRUD ---

type CRUDBuilder struct{ base }

func CRUD(api string) *CRUDBuilder {
    b := &CRUDBuilder{newBase("crud")}
    b.set("api", api)
    b.set("syncLocation", true) // always — enables URL pagination state
    return b
}

func (b *CRUDBuilder) Columns(cols ...M) *CRUDBuilder {
    b.set("columns", cols); return b
}

func (b *CRUDBuilder) Toolbar(items ...any) *CRUDBuilder {
    b.set("toolbar", items); return b
}

func (b *CRUDBuilder) Filter(form M) *CRUDBuilder {
    b.set("filter", form); return b
}

func (b *CRUDBuilder) PerPage(n int) *CRUDBuilder { b.set("perPage", n); return b }

// --- Column ---

type ColumnBuilder struct{ m M }

func Column(name, label string) *ColumnBuilder {
    return &ColumnBuilder{m: M{"name": name, "label": label}}
}

func (c *ColumnBuilder) Type(t string) *ColumnBuilder    { c.m["type"] = t; return c }
func (c *ColumnBuilder) Align(a string) *ColumnBuilder   { c.m["align"] = a; return c }
func (c *ColumnBuilder) Sortable() *ColumnBuilder        { c.m["sortable"] = true; return c }
func (c *ColumnBuilder) Width(w int) *ColumnBuilder      { c.m["width"] = w; return c }
func (c *ColumnBuilder) Tpl(tpl string) *ColumnBuilder   { c.m["tpl"] = tpl; c.m["type"] = "tpl"; return c }
func (c *ColumnBuilder) Map(mapping M) *ColumnBuilder    { c.m["map"] = mapping; return c }

func (c *ColumnBuilder) Buttons(btns ...M) *ColumnBuilder {
    c.m["type"] = "operation"
    c.m["buttons"] = btns
    return c
}

func (c *ColumnBuilder) Build() M { return c.m }

// --- Buttons ---

func CreateBtn(label, api string, fields ...M) M {
    return M{
        "type":       "button",
        "label":      label,
        "level":      "primary",
        "actionType": "dialog",
        "dialog": M{
            "title": label,
            "body": Form(api).Fields(fields...).Build(),
        },
    }
}

func EditBtn(api string, fields ...M) M {
    return M{
        "type":       "button",
        "label":      "Edit",
        "actionType": "dialog",
        "dialog": M{
            "title": "Edit",
            "body": Form(api).InitAPI("get:" + api).Fields(fields...).Build(),
        },
    }
}

func ViewBtn(body any) M {
    return M{
        "type":       "button",
        "label":      "View",
        "actionType": "drawer",
        "drawer": M{
            "title": "Details",
            "body":  body,
        },
    }
}

func DeleteBtn(api string) M {
    return M{
        "type":            "button",
        "label":           "Delete",
        "level":           "danger",
        "actionType":      "ajax",
        "api":             "delete:" + api,
        "confirmText":     "Delete this record? This cannot be undone.",
        "reload":          "window",
    }
}
```

**Expected Behavior:**
- `CRUD()` always has `syncLocation: true` — enforced in constructor
- `DeleteBtn()` always has `confirmText` — enforced in constructor
- `EditBtn()` sets `initApi` for pre-population — consistent pattern

**Test Cases:**
```go
crud := amis.CRUD("get:/api/orders").
    Columns(amis.Column("id", "ID").Build()).
    Build()
assert(crud["syncLocation"] == true)

del := amis.DeleteBtn("delete:/api/orders/${id}")
assert(del["confirmText"] != "")
```

**Acceptance Criteria:**
- [ ] `syncLocation` always true on CRUD
- [ ] `DeleteBtn` always has `confirmText`
- [ ] Column `.Build()` not required when passed to `Columns()` — accept both `*ColumnBuilder` and `M`

**Risk:** `Columns()` takes `...M` but column builders return `*ColumnBuilder`. Options: (a) callers call `.Build()`, (b) use interface. Simplest: require `.Build()` — explicit is better.

**Commit:** `feat(ui/amis): add CRUD and column builders`

---

### TASK 3.4: Implement form builders

**Description:** Implement `Form`, `Wizard`, all field constructors, and field modifiers in `internal/web/amis/form.go`.

**Why:** Forms are the second most common ERP component. Field modifier pattern (`Required(field)` wrapping) enables functional composition without method chains on field builders.

**Implementation:**

```
internal/web/amis/form.go
```

```go
package amis

// --- Form ---

type FormBuilder struct{ base }

func Form(api string) *FormBuilder {
    b := &FormBuilder{newBase("form")}
    b.set("api", api)
    b.set("wrapWithPanel", false)
    return b
}

func (b *FormBuilder) InitAPI(api string) *FormBuilder { b.set("initApi", api); return b }
func (b *FormBuilder) Fields(fields ...M) *FormBuilder { b.set("body", fields); return b }
func (b *FormBuilder) Mode(m string) *FormBuilder      { b.set("mode", m); return b }
func (b *FormBuilder) Horizontal() *FormBuilder        { return b.Mode("horizontal") }

// --- Wizard ---

type WizardBuilder struct{ base }

func Wizard(api string) *WizardBuilder {
    b := &WizardBuilder{newBase("wizard")}
    b.set("api", api)
    return b
}

func (b *WizardBuilder) Step(title string, fields ...M) *WizardBuilder {
    steps, _ := b.m["steps"].(A)
    steps = append(steps, M{"title": title, "body": fields})
    b.set("steps", steps)
    return b
}

// --- Field constructors ---

func TextField(name, label string) M {
    return M{"type": "input-text", "name": name, "label": label}
}

func NumberField(name, label string) M {
    return M{"type": "input-number", "name": name, "label": label}
}

func TextareaField(name, label string) M {
    return M{"type": "textarea", "name": name, "label": label}
}

func DateField(name, label string) M {
    return M{"type": "input-date", "name": name, "label": label}
}

func DatetimeField(name, label string) M {
    return M{"type": "input-datetime", "name": name, "label": label}
}

func SelectField(name, label string, opts ...M) M {
    return M{"type": "select", "name": name, "label": label, "options": opts}
}

func SelectAPIField(name, label, api string) M {
    return M{"type": "select", "name": name, "label": label, "source": api}
}

func SwitchField(name, label string) M {
    return M{"type": "switch", "name": name, "label": label}
}

func FileField(name, label, api string) M {
    return M{"type": "input-file", "name": name, "label": label, "receiver": api}
}

func HiddenField(name string) M {
    return M{"type": "hidden", "name": name}
}

// SelectOpt builds a {label, value} option for SelectField.
func SelectOpt(label, value string) M {
    return M{"label": label, "value": value}
}

// --- Field modifiers (functional, return new M) ---

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

func Default(field M, val any) M {
    field["value"] = val
    return field
}

func VisibleOn(field M, expr string) M {
    field["visibleOn"] = expr
    field["clearValueOnHidden"] = true
    return field
}

func DisabledOn(field M, expr string) M {
    field["disabledOn"] = expr
    return field
}

func Validate(field M, rule string) M {
    field["validations"] = rule
    return field
}
```

**Expected Behavior:**
- `Required(TextField("email", "Email"))` → `{type:"input-text", name:"email", label:"Email", required:true}`
- `VisibleOn()` always adds `clearValueOnHidden` — prevents stale data submission
- Modifiers mutate and return the same map — no allocation chain

**Test Cases:**
```go
f := amis.Required(amis.TextField("name", "Name"))
assert(f["required"] == true)

f = amis.VisibleOn(amis.TextField("alt", "Alt"), "${show_alt}")
assert(f["clearValueOnHidden"] == true)
```

**Acceptance Criteria:**
- [ ] All field types present
- [ ] `VisibleOn` always adds `clearValueOnHidden`
- [ ] Modifiers composable: `Required(Placeholder(TextField("x","X"), "hint"))`

**Commit:** `feat(ui/amis): add form and field builders`

---

### TASK 3.5: Implement App shell builder

**Description:** Implement `App`, `NavGroup`, `NavLink`, `NavGroupSchema` in `internal/web/amis/app.go`.

**Why:** When the custom HTML shell is replaced by the Go-compiled AMIS `app` component (Decision 1 trigger already met), this builder is what generates the navigation tree.

**Implementation:**

```
internal/web/amis/app.go
```

```go
package amis

// NavLinkSchema is used by registry nav group functions.
type NavLinkSchema struct {
    Label    string
    Icon     string
    URL      string
    SchemaAPI string
    Hidden   bool
}

// NavGroupSchema is returned by NavGroupFn.
type NavGroupSchema struct {
    Label  string
    Icon   string
    Links  []NavLinkSchema
    Hidden bool
}

func NavLink(label, icon, url, schemaAPI string) NavLinkSchema {
    return NavLinkSchema{Label: label, Icon: icon, URL: url, SchemaAPI: schemaAPI}
}

func NavGroup(label, icon string, links ...NavLinkSchema) NavGroupSchema {
    return NavGroupSchema{Label: label, Icon: icon, Links: links}
}

// NavGroupToAMIS converts a NavGroupSchema to AMIS pages array entry.
func NavGroupToAMIS(g NavGroupSchema) M {
    if g.Hidden {
        return nil
    }
    children := make(A, 0, len(g.Links))
    for _, l := range g.Links {
        if l.Hidden {
            continue
        }
        children = append(children, M{
            "label":     l.Label,
            "icon":      l.Icon,
            "url":       l.URL,
            "schemaApi": l.SchemaAPI,
        })
    }
    return M{
        "label":    g.Label,
        "icon":     g.Icon,
        "children": children,
    }
}

// --- App ---

type AppBuilder struct{ base }

func App(name string) *AppBuilder {
    b := &AppBuilder{newBase("app")}
    b.set("brandName", name)
    return b
}

func (b *AppBuilder) Logo(url string) *AppBuilder       { b.set("logo", url); return b }
func (b *AppBuilder) Pages(pages ...M) *AppBuilder      { b.set("pages", pages); return b }
func (b *AppBuilder) Header(items ...any) *AppBuilder   { b.set("header", items); return b }
```

**Acceptance Criteria:**
- [ ] `NavGroupToAMIS` skips hidden groups and hidden links
- [ ] `App` builder produces valid AMIS `app` schema

**Commit:** `feat(ui/amis): add App shell and nav builders`

---

## PHASE 4 — IAM Integration

### TASK 4.1: Define PolicyEngine interface

**Description:** Define `PolicyEngine` interface and `CasbinEngine` implementation in `internal/web/policy/`.

**Why:** Page functions call `sess.Can()` which reads a pre-resolved map. But that map must be built *somewhere*. `PolicyEngine` is the seam between IAM (Casbin) and UI compilation — swappable for testing.

**Implementation:**

```
internal/web/policy/
├── engine.go          — PolicyEngine interface
└── casbin_engine.go   — CasbinEngine implementation
```

`engine.go`:
```go
package policy

import "awo.so/internal/web/session"

// UIPermission is a canonical UI permission string (format: "resource.action").
type UIPermission string

const (
    InvoiceRead    UIPermission = "invoice.read"
    InvoiceCreate  UIPermission = "invoice.create"
    InvoiceUpdate  UIPermission = "invoice.update"
    InvoiceDelete  UIPermission = "invoice.delete"
    InvoiceApprove UIPermission = "invoice.approve"

    AccountRead   UIPermission = "account.read"
    AccountCreate UIPermission = "account.create"
    AccountUpdate UIPermission = "account.update"

    PaymentRead   UIPermission = "payment.read"
    PaymentCreate UIPermission = "payment.create"

    ReportRead UIPermission = "report.read"

    UserRead   UIPermission = "user.read"
    UserCreate UIPermission = "user.create"
    UserUpdate UIPermission = "user.update"
    UserDelete UIPermission = "user.delete"

    RoleRead   UIPermission = "role.read"
    RoleCreate UIPermission = "role.create"
    RoleUpdate UIPermission = "role.update"
    RoleDelete UIPermission = "role.delete"

    TenantRead   UIPermission = "tenant.read"
    TenantUpdate UIPermission = "tenant.update"

    SettingsRead   UIPermission = "settings.read"
    SettingsUpdate UIPermission = "settings.update"

    DashboardRead UIPermission = "dashboard.read"
    AuditLogRead  UIPermission = "auditlog.read"
)

// AllUIPermissions is the full list resolved for every session.
// Add entries here when adding a new page that needs access control.
var AllUIPermissions = []UIPermission{
    InvoiceRead, InvoiceCreate, InvoiceUpdate, InvoiceDelete, InvoiceApprove,
    AccountRead, AccountCreate, AccountUpdate,
    PaymentRead, PaymentCreate,
    ReportRead,
    UserRead, UserCreate, UserUpdate, UserDelete,
    RoleRead, RoleCreate, RoleUpdate, RoleDelete,
    TenantRead, TenantUpdate,
    SettingsRead, SettingsUpdate,
    DashboardRead, AuditLogRead,
}

// PolicyEngine resolves permissions for a given subject in a tenant context.
type PolicyEngine interface {
    // Resolve returns a map of all UI permissions for the given subject.
    // subject format: "user:<id>" or "role:<id>"
    Resolve(tenantID, subject string) (map[string]bool, error)
}
```

`casbin_engine.go`:
```go
package policy

import (
    "strings"

    "github.com/casbin/casbin/v2"
)

// CasbinEngine resolves permissions using a Casbin enforcer.
type CasbinEngine struct {
    enforcer *casbin.Enforcer
}

func NewCasbinEngine(enforcer *casbin.Enforcer) *CasbinEngine {
    return &CasbinEngine{enforcer: enforcer}
}

func (e *CasbinEngine) Resolve(tenantID, subject string) (map[string]bool, error) {
    result := make(map[string]bool, len(AllUIPermissions))
    for _, perm := range AllUIPermissions {
        parts := strings.SplitN(string(perm), ".", 2)
        if len(parts) != 2 {
            continue
        }
        resource, action := parts[0], parts[1]
        ok, err := e.enforcer.Enforce(subject, tenantID+":"+resource, action)
        if err != nil {
            return nil, err
        }
        result[string(perm)] = ok
    }
    return result, nil
}
```

**Expected Behavior:**
- `Resolve()` makes exactly `len(AllUIPermissions)` Casbin calls per session build
- Returns `map[string]bool` consumed by `session.Builder.Permissions()`
- `MockPolicyEngine` (test helper): returns a fixed map

**Test Cases:**
```go
// Mock engine
mock := &MockPolicyEngine{perms: map[string]bool{"invoice.read": true}}
result, err := mock.Resolve("tenant-1", "user:abc")
assert(result["invoice.read"] == true)
assert(result["invoice.delete"] == false)
```

**Acceptance Criteria:**
- [ ] `AllUIPermissions` contains all permissions referenced by any page function
- [ ] `CasbinEngine.Resolve()` iterates `AllUIPermissions` — no hardcoded strings
- [ ] Interface is satisfiable by a mock for testing

**Risk:** `AllUIPermissions` is a static list — fragile (must be kept in sync with page functions). Mitigation: schema validator (Phase 5) cross-checks that all `sess.Can()` calls reference a declared permission.

**Commit:** `feat(ui/policy): add PolicyEngine interface and Casbin implementation`

---

### TASK 4.2: Define FeatureEngine interface

**Description:** Define `FeatureEngine` interface and `StaticFeatureEngine` (test helper) + `LayeredFeatureEngine` (production, tenant+user override layering).

**Why:** Feature flags must be tenant-then-user layered. A global flag store does not support this. `FeatureEngine` is the seam.

**Implementation:**

```
internal/web/feature/
├── engine.go         — FeatureEngine interface + AllUIFlags
└── layered_engine.go — LayeredFeatureEngine (production impl)
```

`engine.go`:
```go
package feature

// UIFlag is a canonical feature flag name.
type UIFlag string

const (
    FlagAdvancedReporting UIFlag = "advanced_reporting"
    FlagBulkImport        UIFlag = "bulk_import"
    FlagMultiCurrency     UIFlag = "multi_currency"
    FlagApprovalWorkflow  UIFlag = "approval_workflow"
    FlagAIAssist          UIFlag = "ai_assist"
)

var AllUIFlags = []UIFlag{
    FlagAdvancedReporting,
    FlagBulkImport,
    FlagMultiCurrency,
    FlagApprovalWorkflow,
    FlagAIAssist,
}

// FeatureEngine resolves feature flags for a given tenant and user.
type FeatureEngine interface {
    Resolve(tenantID, userID string) (map[string]bool, error)
}
```

`layered_engine.go`:
```go
package feature

// LayeredFeatureEngine resolves flags in order: global → tenant → user.
// Later layers override earlier ones.
type LayeredFeatureEngine struct {
    store FlagStore
}

// FlagStore is the persistence interface for feature flags.
type FlagStore interface {
    GetGlobal(flag UIFlag) (bool, bool, error)         // (value, exists, err)
    GetTenant(tenantID string, flag UIFlag) (bool, bool, error)
    GetUser(tenantID, userID string, flag UIFlag) (bool, bool, error)
}

func NewLayered(store FlagStore) *LayeredFeatureEngine {
    return &LayeredFeatureEngine{store: store}
}

func (e *LayeredFeatureEngine) Resolve(tenantID, userID string) (map[string]bool, error) {
    result := make(map[string]bool, len(AllUIFlags))
    for _, flag := range AllUIFlags {
        val := false

        if v, ok, err := e.store.GetGlobal(flag); err != nil {
            return nil, err
        } else if ok {
            val = v
        }

        if v, ok, err := e.store.GetTenant(tenantID, flag); err != nil {
            return nil, err
        } else if ok {
            val = v
        }

        if v, ok, err := e.store.GetUser(tenantID, userID, flag); err != nil {
            return nil, err
        } else if ok {
            val = v
        }

        result[string(flag)] = val
    }
    return result, nil
}
```

**Acceptance Criteria:**
- [ ] Layering order: global → tenant → user (user wins)
- [ ] `Resolve()` iterates `AllUIFlags` — no hardcoded strings

**Commit:** `feat(ui/feature): add FeatureEngine interface and layered implementation`

---

### TASK 4.3: Implement IAM middleware

**Description:** Implement `AuthMiddleware` (token validation) and `IAMMiddleware` (session+permission resolution) in `internal/web/middleware/`.

**Why:** These are the two Fiber middlewares that convert an HTTP request into a `SessionContext` — the gateway between the auth layer and the UI compiler.

**Implementation:**

```
internal/web/middleware/
├── auth.go   — AuthMiddleware: extract + validate token
└── iam.go    — IAMMiddleware: resolve permissions + flags, build SessionContext
```

`auth.go`:
```go
package middleware

import (
    "strings"

    "github.com/gofiber/fiber/v2"
    "awo.so/internal/web/amis"
)

const SessionKey = "web_session_id"
const UserIDKey  = "web_user_id"
const TenantIDKey = "web_tenant_id"

func AuthMiddleware(sessionStore SessionStore) fiber.Handler {
    return func(c *fiber.Ctx) error {
        token := extractToken(c)
        if token == "" {
            return c.Status(401).JSON(unauthenticatedSchema())
        }

        sess, err := sessionStore.Validate(c.Context(), token)
        if err != nil || sess == nil {
            return c.Status(401).JSON(unauthenticatedSchema())
        }

        c.Locals(SessionKey, sess)
        c.Locals(UserIDKey, sess.UserID)
        c.Locals(TenantIDKey, sess.TenantID)
        return c.Next()
    }
}

func extractToken(c *fiber.Ctx) string {
    if cookie := c.Cookies("session"); cookie != "" {
        return cookie
    }
    auth := c.Get("Authorization")
    if strings.HasPrefix(auth, "Bearer ") {
        return strings.TrimPrefix(auth, "Bearer ")
    }
    return ""
}

func unauthenticatedSchema() amis.M {
    return amis.M{
        "status": 401,
        "msg":    "Session expired. Please log in.",
        "data": amis.M{
            "type":       "page",
            "body":       amis.M{"type": "tpl", "tpl": "Your session has expired."},
            "toolbar": []any{
                amis.M{"type": "button", "label": "Login", "level": "primary",
                    "actionType": "url", "url": "/login"},
            },
        },
    }
}
```

`iam.go`:
```go
package middleware

import (
    "github.com/gofiber/fiber/v2"
    "awo.so/internal/web/feature"
    "awo.so/internal/web/policy"
    "awo.so/internal/web/session"
)

const SessionContextKey = "web_session_ctx"

func IAMMiddleware(pol policy.PolicyEngine, feat feature.FeatureEngine) fiber.Handler {
    return func(c *fiber.Ctx) error {
        raw, ok := c.Locals(SessionKey).(*RawSession)
        if !ok {
            return c.Status(500).JSON(internalErrorSchema())
        }

        perms, err := pol.Resolve(raw.TenantID, "user:"+raw.UserID)
        if err != nil {
            return c.Status(500).JSON(internalErrorSchema())
        }

        flags, err := feat.Resolve(raw.TenantID, raw.UserID)
        if err != nil {
            return c.Status(500).JSON(internalErrorSchema())
        }

        ctx, err := session.New().
            User(raw.UserID, raw.Email, raw.DisplayName).
            Tenant(raw.TenantID).
            Locale(raw.Locale).
            Timezone(raw.Timezone).
            Currency(raw.Currency).
            Auth(raw.AuthMethod, raw.ExpiresAt).
            Permissions(perms).
            Flags(flags).
            Build()
        if err != nil {
            return c.Status(500).JSON(internalErrorSchema())
        }

        c.Locals(SessionContextKey, ctx)
        return c.Next()
    }
}
```

**Expected Behavior:**
- `AuthMiddleware` returns 401 AMIS schema on missing/invalid token — not JSON error, not redirect
- `IAMMiddleware` bulk-resolves all permissions once per request
- `SessionContext` stored in Fiber locals, consumed by `SchemaHandler`

**Acceptance Criteria:**
- [ ] 401 response is valid AMIS envelope (browser renders login button, not blank page)
- [ ] `IAMMiddleware` does not access Casbin directly — uses `PolicyEngine` interface
- [ ] Both middlewares tested with mock `SessionStore` and `PolicyEngine`

**Commit:** `feat(ui/middleware): add AuthMiddleware and IAMMiddleware`

---

## PHASE 5 — Schema Compiler

### TASK 5.1: Implement schema cache

**Description:** Implement `SchemaCache` interface + `InMemorySchemaCache` in `internal/web/compiler/cache.go`.

**Why:** Without caching, every request re-executes the page function + Casbin resolution. With fingerprint-based keys, users sharing the same role set share the same cache entry.

**Implementation:**

```
internal/web/compiler/cache.go
```

```go
package compiler

import (
    "fmt"
    "sync"
    "time"

    "awo.so/internal/web"
)

// SchemaCache stores compiled schemas keyed by session fingerprint.
type SchemaCache interface {
    Get(key string) (web.Schema, bool)
    Set(key string, schema web.Schema, ttl time.Duration)
    Invalidate(tenantID string) // flush all entries for a tenant
}

type cacheEntry struct {
    schema    web.Schema
    expiresAt time.Time
}

// InMemorySchemaCache is an LRU-style in-memory cache.
// Replace with Redis for multi-instance deployments.
type InMemorySchemaCache struct {
    mu      sync.RWMutex
    entries map[string]cacheEntry
    maxSize int
}

func NewInMemoryCache(maxSize int) *InMemorySchemaCache {
    return &InMemorySchemaCache{
        entries: make(map[string]cacheEntry, maxSize),
        maxSize: maxSize,
    }
}

func (c *InMemorySchemaCache) Get(key string) (web.Schema, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    e, ok := c.entries[key]
    if !ok || time.Now().After(e.expiresAt) {
        return nil, false
    }
    return e.schema, true
}

func (c *InMemorySchemaCache) Set(key string, schema web.Schema, ttl time.Duration) {
    c.mu.Lock()
    defer c.mu.Unlock()
    if len(c.entries) >= c.maxSize {
        c.evictOldest()
    }
    c.entries[key] = cacheEntry{schema: schema, expiresAt: time.Now().Add(ttl)}
}

func (c *InMemorySchemaCache) Invalidate(tenantID string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    prefix := tenantID + "|"
    for k := range c.entries {
        if len(k) > len(prefix) && k[:len(prefix)] == prefix {
            delete(c.entries, k)
        }
    }
}

func (c *InMemorySchemaCache) evictOldest() {
    var oldestKey string
    var oldestTime time.Time
    for k, e := range c.entries {
        if oldestKey == "" || e.expiresAt.Before(oldestTime) {
            oldestKey = k
            oldestTime = e.expiresAt
        }
    }
    if oldestKey != "" {
        delete(c.entries, oldestKey)
    }
}

// CacheKey builds the canonical cache key for a schema request.
func CacheKey(tenantID, path, permFingerprint, flagFingerprint string) string {
    return fmt.Sprintf("%s|%s|%s|%s", tenantID, path, permFingerprint, flagFingerprint)
}
```

**Expected Behavior:**
- Cache miss on expired entry (TTL-based)
- `Invalidate(tenantID)` evicts all entries for a tenant — called on role change
- `evictOldest` prevents unbounded growth

**Test Cases:**
```go
cache := NewInMemoryCache(2)
cache.Set("k1", web.Schema{"type": "page"}, time.Minute)
s, ok := cache.Get("k1")
assert(ok && s["type"] == "page")

// Expiry
cache.Set("k2", web.Schema{}, time.Millisecond)
time.Sleep(2 * time.Millisecond)
_, ok = cache.Get("k2")
assert(!ok)

// Eviction at maxSize
cache.Set("k3", web.Schema{}, time.Minute)
cache.Set("k4", web.Schema{}, time.Minute) // evicts oldest
```

**Acceptance Criteria:**
- [ ] Expired entries return cache miss
- [ ] `Invalidate` removes all entries prefixed with tenantID
- [ ] Thread-safe under concurrent reads and writes

**Commit:** `feat(ui/compiler): add in-memory schema cache with TTL and tenant invalidation`

---

### TASK 5.2: Implement schema compiler

**Description:** Implement `Compiler.Compile(path, sess)` — the central dispatch: cache lookup → registry lookup → PageFn execution → validation → cache store.

**Why:** This is the core of the entire UI Services layer. Everything else feeds into this function.

**Implementation:**

```
internal/web/compiler/compiler.go
```

```go
package compiler

import (
    "errors"
    "fmt"
    "time"

    "awo.so/internal/web"
    "awo.so/internal/web/registry"
    "awo.so/internal/web/session"
)

const defaultTTL = 5 * time.Minute

var (
    ErrNotFound      = errors.New("compiler: page not found")
    ErrUnauthenticated = errors.New("compiler: session unauthenticated")
    ErrExpired       = errors.New("compiler: session expired")
)

// Compiler compiles page schemas from session context.
type Compiler struct {
    cache     SchemaCache
    validator *SchemaValidator
}

func New(cache SchemaCache, validator *SchemaValidator) *Compiler {
    return &Compiler{cache: cache, validator: validator}
}

// Compile returns the AMIS schema for the given path and session.
func (c *Compiler) Compile(path string, sess session.SessionContext) (web.Schema, error) {
    if !sess.Authenticated() {
        return nil, ErrUnauthenticated
    }

    key := CacheKey(
        sess.TenantID,
        path,
        sess.PermissionFingerprint(),
        sess.FlagFingerprint(),
    )

    if schema, ok := c.cache.Get(key); ok {
        return schema, nil
    }

    fn := registry.Get(path)
    if fn == nil {
        return nil, fmt.Errorf("%w: %s", ErrNotFound, path)
    }

    schema := fn(sess)

    if c.validator != nil {
        if err := c.validator.Validate(schema); err != nil {
            return nil, fmt.Errorf("compiler: schema validation failed for %s: %w", path, err)
        }
    }

    c.cache.Set(key, schema, defaultTTL)
    return schema, nil
}

// CompileApp returns the navigation schema for the app shell.
func (c *Compiler) CompileApp(sess session.SessionContext) (web.Schema, error) {
    if !sess.Authenticated() {
        return nil, ErrUnauthenticated
    }

    groups := registry.GetNavGroups()
    pages := make([]web.M, 0, len(groups))
    for _, groupFn := range groups {
        group := groupFn(sess)
        if m := amisNavGroupToM(group); m != nil {
            pages = append(pages, m)
        }
    }

    return web.Schema{
        "type":  "app",
        "pages": pages,
    }, nil
}
```

**Expected Behavior:**
- Cache hit returns immediately without calling `PageFn`
- Cache miss: calls `PageFn`, validates, stores with 5m TTL
- `ErrNotFound` for unregistered paths (caller returns 404)
- `ErrUnauthenticated` never results in a cached schema

**Test Cases:**
```go
// Cache hit
mock := &mockPageFn{callCount: 0}
registry.Register("/test", mock.fn)
compiler.Compile("/test", sess)
compiler.Compile("/test", sess) // same fingerprint
assert(mock.callCount == 1) // fn called once, not twice

// Cache miss on different permission set
sess2 := buildSession(differentPerms)
compiler.Compile("/test", sess2)
assert(mock.callCount == 2)

// 404
_, err := compiler.Compile("/nonexistent", sess)
assert(errors.Is(err, ErrNotFound))
```

**Acceptance Criteria:**
- [ ] PageFn called exactly once for identical fingerprint
- [ ] PageFn called again after TTL expiry
- [ ] Validator called on every cache miss (not on cache hit)
- [ ] `ErrUnauthenticated` returned before any registry lookup

**Commit:** `feat(ui/compiler): add schema compiler with cache and validation`

---

### TASK 5.3: Implement schema validator

**Description:** Implement `SchemaValidator` that walks the AMIS schema tree and enforces invariants: no IAM expressions, API method prefix, CRUD syncLocation, chart transparent bg.

**Why:** Without validation at compile time, IAM leakage (role checks in `visibleOn`) and structural errors ship to the browser silently.

**Implementation:**

```
internal/web/compiler/validator.go
```

```go
package compiler

import (
    "fmt"
    "strings"

    "awo.so/internal/web"
)

// ValidationError reports a schema invariant violation.
type ValidationError struct {
    Path    string
    Rule    string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("schema validation [%s] at %s: %s", e.Rule, e.Path, e.Message)
}

// SchemaValidator checks compiled schemas for invariant violations.
type SchemaValidator struct {
    rules []validationRule
}

type validationRule struct {
    name string
    fn   func(path string, node web.M) *ValidationError
}

func NewSchemaValidator() *SchemaValidator {
    v := &SchemaValidator{}
    v.rules = []validationRule{
        {"no-iam-in-expressions", checkNoIAMExpressions},
        {"api-method-prefix", checkAPIMethodPrefix},
        {"crud-sync-location", checkCRUDSyncLocation},
        {"chart-transparent-bg", checkChartTransparentBg},
    }
    return v
}

func (v *SchemaValidator) Validate(schema web.Schema) error {
    return walkSchema("$", schema, func(path string, node web.M) error {
        for _, rule := range v.rules {
            if err := rule.fn(path, node); err != nil {
                return err
            }
        }
        return nil
    })
}

func walkSchema(path string, node web.M, fn func(string, web.M) error) error {
    if err := fn(path, node); err != nil {
        return err
    }
    for k, v := range node {
        switch val := v.(type) {
        case web.M:
            if err := walkSchema(path+"."+k, val, fn); err != nil {
                return err
            }
        case []any:
            for i, item := range val {
                if m, ok := item.(web.M); ok {
                    if err := walkSchema(fmt.Sprintf("%s.%s[%d]", path, k, i), m, fn); err != nil {
                        return err
                    }
                }
            }
        }
    }
    return nil
}

var iamKeywords = []string{"role:", "permission:", "sess.Can", "sess.Flag", ".roles", ".permissions"}

func checkNoIAMExpressions(path string, node web.M) *ValidationError {
    for _, key := range []string{"visibleOn", "disabledOn", "hiddenOn"} {
        if expr, ok := node[key].(string); ok {
            for _, kw := range iamKeywords {
                if strings.Contains(expr, kw) {
                    return &ValidationError{
                        Path:    path,
                        Rule:    "no-iam-in-expressions",
                        Message: fmt.Sprintf("%q contains IAM keyword %q — use data variables set by Go, not AMIS expressions", expr, kw),
                    }
                }
            }
        }
    }
    return nil
}

func checkAPIMethodPrefix(path string, node web.M) *ValidationError {
    if api, ok := node["api"].(string); ok {
        if api != "" && !hasMethodPrefix(api) {
            return &ValidationError{
                Path:    path,
                Rule:    "api-method-prefix",
                Message: fmt.Sprintf("api %q missing method prefix (get:/post:/put:/delete:/patch:)", api),
            }
        }
    }
    return nil
}

func hasMethodPrefix(api string) bool {
    for _, p := range []string{"get:", "post:", "put:", "delete:", "patch:"} {
        if strings.HasPrefix(api, p) {
            return true
        }
    }
    return false
}

func checkCRUDSyncLocation(path string, node web.M) *ValidationError {
    if node["type"] == "crud" {
        if sync, ok := node["syncLocation"].(bool); !ok || !sync {
            return &ValidationError{
                Path:    path,
                Rule:    "crud-sync-location",
                Message: "CRUD component missing syncLocation:true",
            }
        }
    }
    return nil
}

func checkChartTransparentBg(path string, node web.M) *ValidationError {
    if node["type"] == "chart" {
        style, _ := node["style"].(web.M)
        if style == nil || style["background"] != "transparent" {
            return &ValidationError{
                Path:    path,
                Rule:    "chart-transparent-bg",
                Message: "chart missing style.background:transparent",
            }
        }
    }
    return nil
}
```

**Expected Behavior:**
- Validator catches `"visibleOn": "${role == 'admin'}"` — fails `no-iam-in-expressions`
- Catches `"api": "/api/orders"` — fails `api-method-prefix`
- Catches CRUD without `syncLocation` — fails `crud-sync-location`
- Passes valid schemas silently

**Test Cases:**
```go
v := NewSchemaValidator()

// IAM expression — fail
err := v.Validate(web.Schema{"type":"page","body":web.M{"visibleOn":"${role == 'admin'}"}})
assert(err != nil && strings.Contains(err.Error(), "no-iam-in-expressions"))

// API without prefix — fail
err = v.Validate(web.Schema{"type":"page","body":web.M{"type":"crud","api":"/api/orders","syncLocation":true}})
assert(err != nil && strings.Contains(err.Error(), "api-method-prefix"))

// Valid schema — pass
err = v.Validate(web.Schema{"type":"page","body":web.M{"type":"crud","api":"get:/api/orders","syncLocation":true}})
assert(err == nil)
```

**Acceptance Criteria:**
- [ ] All 4 rules implemented
- [ ] Validator walks nested schemas (body, body.body, columns[*], etc.)
- [ ] `ValidationError` includes path for debugging

**Commit:** `feat(ui/compiler): add schema validator with IAM-leakage and structural rules`

---

## PHASE 6 — HTTP Handler

### TASK 6.1: Implement SchemaHandler

**Description:** Implement `SchemaHandler` — the single Fiber handler for all `/schema/*` routes — in `internal/web/handler/schema.go`.

**Why:** Without a handler, the compiler has no HTTP surface. The handler bridges Fiber context → `SessionContext` → `Compiler.Compile()` → JSON response.

**Implementation:**

```
internal/web/handler/schema.go
```

```go
package handler

import (
    "errors"

    "github.com/gofiber/fiber/v2"
    "awo.so/internal/web/compiler"
    "awo.so/internal/web/middleware"
    "awo.so/internal/web/session"
)

// SchemaHandler serves AMIS schemas at /schema/:path+
type SchemaHandler struct {
    compiler *compiler.Compiler
}

func NewSchemaHandler(c *compiler.Compiler) *SchemaHandler {
    return &SchemaHandler{compiler: c}
}

func (h *SchemaHandler) Handle(c *fiber.Ctx) error {
    sess, ok := c.Locals(middleware.SessionContextKey).(session.SessionContext)
    if !ok || !sess.Authenticated() {
        return c.Status(401).JSON(fiber.Map{
            "status": 401,
            "msg":    "Unauthorized",
        })
    }

    path := "/" + c.Params("*")
    schema, err := h.compiler.Compile(path, sess)
    if err != nil {
        if errors.Is(err, compiler.ErrNotFound) {
            return c.Status(404).JSON(fiber.Map{
                "status": 404,
                "msg":    "Page not found: " + path,
            })
        }
        // validation error or internal — log but return 500
        c.Context().Logger().Printf("schema compile error for %s: %v", path, err)
        return c.Status(500).JSON(fiber.Map{
            "status": 500,
            "msg":    "Internal error compiling schema",
        })
    }

    return c.JSON(fiber.Map{
        "status": 0,
        "data":   schema,
    })
}

// AppHandler serves the AMIS app shell schema at /schema/app
type AppHandler struct {
    compiler *compiler.Compiler
}

func NewAppHandler(c *compiler.Compiler) *AppHandler {
    return &AppHandler{compiler: c}
}

func (h *AppHandler) Handle(c *fiber.Ctx) error {
    sess, ok := c.Locals(middleware.SessionContextKey).(session.SessionContext)
    if !ok || !sess.Authenticated() {
        return c.Status(401).JSON(fiber.Map{"status": 401, "msg": "Unauthorized"})
    }

    schema, err := h.compiler.CompileApp(sess)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"status": 500, "msg": "Internal error"})
    }

    return c.JSON(fiber.Map{"status": 0, "data": schema})
}
```

Route registration:
```go
// In routes setup (e.g., internal/api/handlers/routes.go)
schemaHandler := handler.NewSchemaHandler(appCompiler)
appHandler    := handler.NewAppHandler(appCompiler)

api := app.Group("/schema", middleware.AuthMiddleware(sessionStore), middleware.IAMMiddleware(policyEngine, featureEngine))
api.Get("/app", appHandler.Handle)
api.Get("/*", schemaHandler.Handle)
```

**Expected Behavior:**
- `GET /schema/finance/invoices` → `{status:0, data:{type:"page",...}}`
- `GET /schema/nonexistent` → `{status:404, msg:"Page not found: /nonexistent"}`
- Unauthenticated → `{status:401, msg:"Unauthorized"}`
- Path wildcard: `/schema/*` captures everything after `/schema/`

**Test Cases:**
```go
// Happy path
req := httptest.NewRequest("GET", "/schema/finance/invoices", nil)
// with valid session cookie
resp := app.Test(req)
assert(resp.StatusCode == 200)
body := parseJSON(resp)
assert(body["status"] == 0)
assert(body["data"].(M)["type"] == "page")

// 404
req = httptest.NewRequest("GET", "/schema/does/not/exist", nil)
resp = app.Test(req)
assert(resp.StatusCode == 404)
```

**Acceptance Criteria:**
- [ ] All responses are valid AMIS envelope `{status, data}`
- [ ] 404 for unknown path, not 500
- [ ] Compiler errors log but return 500 (not leak internal details)
- [ ] Both `/schema/app` and `/schema/*` routes tested

**Commit:** `feat(ui/handler): add SchemaHandler and AppHandler`

---

## PHASE 7 — First Page Schemas

### TASK 7.1: Implement dashboard schema

**Description:** Implement `internal/web/pages/dashboard/schema.go` — the first real page using the new compiler pipeline end-to-end.

**Why:** The dashboard is the simplest page (no create/edit/delete) and validates that the entire pipeline works: `init()` registration → middleware → compiler → cache → JSON.

**Implementation:**

```
internal/web/pages/dashboard/schema.go
```

```go
package dashboard

import (
    "awo.so/internal/web/amis"
    "awo.so/internal/web/registry"
    "awo.so/internal/web/session"
)

func init() {
    registry.Register("/dashboard", Schema)
}

func Schema(sess session.SessionContext) amis.Schema {
    body := []any{
        summaryRow(sess),
    }

    if sess.Flag("advanced_reporting") {
        body = append(body, advancedMetricsPanel(sess))
    }

    return amis.Page("Dashboard").
        Data(amis.M{
            "user_name":    sess.DisplayName,
            "tenant_id":   sess.TenantID,
        }).
        Body(body...).
        Build()
}

func summaryRow(sess session.SessionContext) amis.M {
    cols := []amis.M{}

    if sess.Can("read", "invoice") {
        cols = append(cols, amis.Col(3, amis.Panel("Invoices").
            Body(amis.Chart("get:/api/v1/finance/invoices/summary").Height(200).Build()).
            Build()))
    }

    if sess.Can("read", "payment") {
        cols = append(cols, amis.Col(3, amis.Panel("Payments").
            Body(amis.Chart("get:/api/v1/finance/payments/summary").Height(200).Build()).
            Build()))
    }

    if len(cols) == 0 {
        return amis.Alert("info", "No data available for your role.").Build()
    }

    return amis.Grid(cols...).Build()
}

func advancedMetricsPanel(sess session.SessionContext) amis.M {
    return amis.Panel("Advanced Metrics").
        Body(amis.Chart("get:/api/v1/reports/metrics").Height(300).Build()).
        Build()
}
```

**Expected Behavior:**
- User with `invoice.read` → sees invoice chart panel
- User without any permissions → sees `Alert` ("No data available")
- User with `advanced_reporting` flag → sees additional metrics panel
- `GET /schema/dashboard` returns valid AMIS page JSON

**Test Cases:**
```go
// Full permissions
sess := buildSession(map[string]bool{"invoice.read": true, "payment.read": true})
schema := dashboard.Schema(sess)
assert(schema["type"] == "page")
// body contains grid with 2 cols

// No permissions
sess = buildSession(map[string]bool{})
schema = dashboard.Schema(sess)
// body is alert

// Feature flag
sess = buildSessionWithFlag("advanced_reporting", true)
schema = dashboard.Schema(sess)
// body includes advanced metrics panel
```

**Acceptance Criteria:**
- [ ] Schema passes `SchemaValidator` with zero errors
- [ ] Dashboard rendered in browser for a test session
- [ ] No permissions → graceful degradation (not empty page)

**Commit:** `feat(ui/pages): add dashboard schema`

---

### TASK 7.2: Implement finance/invoices schema

**Description:** Implement the full invoices page with CRUD, create dialog, edit dialog, view drawer, delete button — all gated by permissions.

**Implementation:**

```
internal/web/pages/finance/invoices/schema.go
```

```go
package invoices

import (
    "awo.so/internal/web/amis"
    "awo.so/internal/web/registry"
    "awo.so/internal/web/session"
)

func init() {
    registry.Register("/finance/invoices", Schema)
}

func Schema(sess session.SessionContext) amis.Schema {
    cols := []amis.M{
        amis.Column("number", "Invoice #").Sortable().Build(),
        amis.Column("supplier_name", "Supplier").Build(),
        amis.Column("amount", "Amount").Type("currency").Align("right").Build(),
        amis.Column("status", "Status").Map(statusMap()).Build(),
        amis.Column("due_date", "Due Date").Type("date").Sortable().Build(),
    }

    actionBtns := []amis.M{}
    actionBtns = append(actionBtns, amis.ViewBtn(detailDrawer()))
    if sess.Can("update", "invoice") {
        actionBtns = append(actionBtns, amis.EditBtn("put:/api/v1/finance/invoices/${id}", editFields()...))
    }
    if sess.Can("delete", "invoice") {
        actionBtns = append(actionBtns, amis.DeleteBtn("delete:/api/v1/finance/invoices/${id}"))
    }
    cols = append(cols, amis.Column("", "Actions").Buttons(actionBtns...).Build())

    toolbar := []any{}
    if sess.Can("create", "invoice") {
        toolbar = append(toolbar, amis.CreateBtn("New Invoice", "post:/api/v1/finance/invoices", editFields()...))
    }

    crud := amis.CRUD("get:/api/v1/finance/invoices").
        Columns(cols...).
        Build()
    if len(toolbar) > 0 {
        crud["toolbar"] = toolbar
    }

    return amis.Page("Invoices").Body(crud).Build()
}

func statusMap() amis.M {
    return amis.M{
        "draft":    "<span class='label label-default'>Draft</span>",
        "pending":  "<span class='label label-warning'>Pending</span>",
        "approved": "<span class='label label-success'>Approved</span>",
        "paid":     "<span class='label label-info'>Paid</span>",
        "overdue":  "<span class='label label-danger'>Overdue</span>",
    }
}

func editFields() []amis.M {
    return []amis.M{
        amis.Required(amis.SelectAPIField("supplier_id", "Supplier", "get:/api/v1/suppliers?select=id,name")),
        amis.Required(amis.NumberField("amount", "Amount")),
        amis.Required(amis.SelectField("currency", "Currency",
            amis.SelectOpt("KES", "KES"),
            amis.SelectOpt("USD", "USD"),
            amis.SelectOpt("EUR", "EUR"),
        )),
        amis.Required(amis.DateField("due_date", "Due Date")),
        amis.Optional(amis.TextareaField("notes", "Notes")),
    }
}

func detailDrawer() amis.M {
    return amis.Service("get:/api/v1/finance/invoices/${id}").
        Body(
            amis.Descriptions("Invoice Details").
                Item("Invoice #", "number").
                Item("Supplier", "supplier_name").
                Item("Amount", "amount").
                Item("Status", "status").
                Item("Due Date", "due_date").
                Item("Notes", "notes").
                Columns(2).
                Build(),
        ).Build()
}
```

**Acceptance Criteria:**
- [ ] User with only `invoice.read` → no create/edit/delete buttons
- [ ] User with `invoice.approve` → approve action visible
- [ ] Schema passes `SchemaValidator`
- [ ] CRUD has `syncLocation: true` (enforced by builder)

**Commit:** `feat(ui/pages): add finance/invoices schema`

---

## PHASE 8 — Wire-Up and Integration

### TASK 8.1: Wire compiler into application

**Description:** Instantiate and connect all components: `PolicyEngine`, `FeatureEngine`, `SchemaCache`, `SchemaValidator`, `Compiler`, middlewares, and handlers in the application bootstrap.

**Why:** All individual components are implemented but disconnected. This task wires them into the Fiber app.

**Implementation:**

Add to application bootstrap (e.g., `cmd/server/main.go` or `internal/api/app.go`):

```go
// --- UI Services wiring ---
import (
    "awo.so/internal/web/compiler"
    "awo.so/internal/web/feature"
    "awo.so/internal/web/handler"
    "awo.so/internal/web/middleware"
    "awo.so/internal/web/policy"

    // Page registrations — blank imports trigger init()
    _ "awo.so/internal/web/pages/dashboard"
    _ "awo.so/internal/web/pages/finance"   // pulls invoices + accounts via module.go
    _ "awo.so/internal/web/pages/admin"
    _ "awo.so/internal/web/pages/settings"
)

func setupUIServices(app *fiber.App, casbinEnforcer *casbin.Enforcer, flagStore feature.FlagStore, sessionStore middleware.SessionStore) {
    pol  := policy.NewCasbinEngine(casbinEnforcer)
    feat := feature.NewLayered(flagStore)
    cache := compiler.NewInMemoryCache(5000)
    val  := compiler.NewSchemaValidator()
    comp := compiler.New(cache, val)

    schemaHandler := handler.NewSchemaHandler(comp)
    appHandler    := handler.NewAppHandler(comp)

    auth := middleware.AuthMiddleware(sessionStore)
    iam  := middleware.IAMMiddleware(pol, feat)

    schema := app.Group("/schema", auth, iam)
    schema.Get("/app", appHandler.Handle)
    schema.Get("/*", schemaHandler.Handle)
}
```

**Expected Behavior:**
- `GET /schema/dashboard` with valid session → 200 with page JSON
- `GET /schema/dashboard` without session → 401
- `GET /schema/nonexistent` → 404

**Acceptance Criteria:**
- [ ] All blank imports present — all pages reachable
- [ ] Middleware chain: Auth → IAM → Handler
- [ ] No page package imported directly (only blank imports)

**Commit:** `feat(ui): wire schema compiler, middleware, and handlers into app`

---

### TASK 8.2: Add cache invalidation hooks

**Description:** Add `cache.Invalidate(tenantID)` calls at IAM mutation points: role assignment, permission change, tenant update.

**Why:** Without invalidation, a user who gains a new role still sees the old (cached) schema until TTL expires.

**Implementation:**

In IAM service (wherever roles are assigned/revoked):
```go
func (s *IAMService) AssignRole(ctx context.Context, tenantID, userID, roleID string) error {
    err := s.repo.AssignRole(ctx, tenantID, userID, roleID)
    if err != nil {
        return err
    }
    // Invalidate all cached schemas for this tenant — users will recompile on next request
    s.schemaCache.Invalidate(tenantID)
    return nil
}
```

Expose `SchemaCache.Invalidate` via a lightweight interface so IAM service doesn't import the compiler package:

```go
// internal/shared/ports/schema_cache.go
package ports

type SchemaCacheInvalidator interface {
    Invalidate(tenantID string)
}
```

**Acceptance Criteria:**
- [ ] Role assignment triggers cache invalidation for tenant
- [ ] IAM service depends on interface, not concrete compiler package
- [ ] Invalidation tested: role change → next request recompiles schema

**Commit:** `feat(ui): invalidate schema cache on IAM mutations`

---

## PHASE 9 — Observability

### TASK 9.1: Add schema compilation tracing

**Description:** Add structured logging to the compiler: cache hit/miss, compilation duration, validation errors, page path, tenant ID.

**Why:** Without observability, slow schemas and cache problems are invisible in production.

**Implementation:**

In `compiler.go`, add to `Compile()`:
```go
start := time.Now()

if schema, ok := c.cache.Get(key); ok {
    c.logger.Info("schema cache hit",
        "path", path,
        "tenant", sess.TenantID,
        "latency_us", time.Since(start).Microseconds(),
    )
    return schema, nil
}

// ... compile ...

c.logger.Info("schema compiled",
    "path", path,
    "tenant", sess.TenantID,
    "latency_ms", time.Since(start).Milliseconds(),
    "cache_miss", true,
)
```

Add compilation metrics (counters):
```
ui_schema_cache_hits_total{path, tenant}
ui_schema_cache_misses_total{path, tenant}
ui_schema_compile_duration_ms{path, tenant}
ui_schema_validation_errors_total{path, rule}
```

**Acceptance Criteria:**
- [ ] Cache hit/miss logged at INFO with path and tenant
- [ ] Compilation duration logged on cache miss
- [ ] Validation errors logged at WARN with rule name and path in schema

**Commit:** `feat(ui/compiler): add structured logging and metrics for schema compilation`

---

## PHASE 9 — Cutover

### TASK 9.2: Remove static JSON schemas

**Description:** Delete `web/schemas/pages/*.json` and verify that all pages are served by the Go compiler. Update `web/pages/index.html` to remove any static schema fallback paths.

**Why:** Static JSON files are the old system. Leaving them creates confusion about the source of truth and a maintenance split.

**Implementation:**
1. Verify all paths in `web/schemas/pages/` have corresponding Go page functions registered
2. Delete `web/schemas/pages/` directory
3. Update `index.html` to remove static schema fallback (if any)
4. Verify `navigate()` in `index.html` calls `/schema/<path>` exclusively

**Acceptance Criteria:**
- [ ] `web/schemas/pages/` directory deleted
- [ ] All previously static pages now served by Go compiler
- [ ] No 404s in browser after cutover
- [ ] `index.html` has no references to `.json` schema files

**Commit:** `chore(ui): remove static JSON schemas — all pages served by Go compiler`

---

## Appendix — Dependency Graph

```
Phase 1 (Session) → Phase 2 (Registry) → Phase 3 (Builders)
                                       ↘
Phase 4 (IAM/Middleware) ───────────────→ Phase 5 (Compiler) → Phase 6 (Handler)
                                                              ↓
                                                       Phase 7 (Pages)
                                                              ↓
                                                       Phase 8 (Wire-Up)
                                                              ↓
                                                       Phase 9 (Observability + Cutover)
```

No phase can begin until all phases it depends on are complete. Within each phase, tasks are sequential (each task builds on the previous).

---

## Appendix — Acceptance Gate (End-to-End)

Before Phase 9 cutover is done, the following must all pass:

- [ ] `GET /schema/dashboard` with finance-manager session → 200, type=page, chart panels present
- [ ] `GET /schema/finance/invoices` with read-only session → 200, no create/edit/delete buttons in schema
- [ ] `GET /schema/finance/invoices` with full session → 200, all action buttons present
- [ ] Two requests with identical permission fingerprint → PageFn called exactly once (cache hit verified via logs)
- [ ] Role change via API → next schema request recompiles (cache invalidated)
- [ ] `GET /schema/nonexistent` → 404
- [ ] `GET /schema/dashboard` without session cookie → 401 with AMIS login button
- [ ] Schema validator catches IAM expression in test page → compilation returns error
- [ ] All 25 `AllUIPermissions` entries have at least one corresponding `sess.Can()` call in page schemas
