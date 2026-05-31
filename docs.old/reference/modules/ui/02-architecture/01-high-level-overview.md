# High-Level Architecture Overview

> Last verified: 2026-05-18 | Code pointer: `web/pages/index.html`, `internal/web/handler/schema.go`, `internal/web/ui/pipeline.go`

---

## 📖 The Core Idea

The browser is a dumb renderer. Go is the UI compiler. Every page is a Go function that returns JSON. The browser fetches JSON and hands it to AMIS, which renders it. No React, no Vue, no build step for pages. Permission logic, feature flags, and tenant customization all happen in Go — before a single byte reaches the browser.

---

## Full Request Flow (Current State)

```
User opens browser → GET / → Go serves web/pages/index.html
        │
        ▼
Browser parses index.html:
├── Loads /sdk/sdk.js     (AMIS runtime, ~2MB, cached)
├── Loads /sdk/sdk.css    (AMIS styles)
├── Loads /sdk/charts.js  (ECharts for chart component)
└── Shell renders: sidebar + topbar from hardcoded menuConfig JS array

        │
        ▼
Shell JS runs:
├── Reads localStorage for theme preference → applies html.dark if needed
├── Renders sidebar from menuConfig
├── Detects URL hash (e.g. #finance/invoices)
└── Calls navigate('#finance/invoices')

        │
        ▼
navigate() → GET /schema/finance/invoices
  (NOT static JSON — Go-driven schema endpoint)
        │
        ▼
Middleware chain:
  iam/middleware.Authenticate → validates session token, sets Fiber Locals
  contract.InjectSessionContext → bridges Fiber Locals → Go context
        │
        ▼
SchemaHandler.Handle():
  contract.FromContext(ctx) → validates session present
  Builds OperationContext
  pipeline.Run(opCtx)
        │
        ▼
UI Pipeline (9 stages, P=10→90):
  SessionStage  → validates contract.SessionContext
  AuthzStage    → BulkEnforce → 35 permissions resolved → UISessionContext
  CacheLookup   → key=tenant:route:permFP:flagFP → Redis hit? → jump to Response
  Registry      → registry.GetRegistration(route) → ASTPageFn or PageFn
  Compile       → fn(UISessionContext) → Schema
  Normalize     → lowercase types, trim API strings
  Validate      → reject forbidden expressions
  CacheStore    → Redis write TTL=5min
  Response      → UISchemaOutput assembled
        │
        ▼
SchemaHandler → c.JSON({"status": 0, "data": schema})
        │
        ▼
Browser: amis.embed(#content, schema.data, {}, amisEnv)
AMIS renders components. Each component with api: field calls /api/v1/...
        │
        ▼
AMIS fetcher → /api/v1/finance/invoices?offset=0&limit=20
  Translates AMIS query params (page/perPage) → Go query params (offset/limit)
  Translates Go response {success, data, meta} → AMIS {status, data: {items, count}}
```

---

## File Map

```
web/
├── pages/
│   └── index.html        ← Shell: CSS + HTML + JS. No build step.
├── sdk/
│   ├── sdk.js            ← AMIS runtime (~2MB, pre-built, never modify)
│   ├── sdk.css           ← AMIS base styles (never modify directly)
│   ├── charts.js         ← ECharts integration
│   └── helper.css        ← AMIS helper utilities
├── utils/
│   └── locale-en.js      ← AMIS English locale strings
└── public/               ← Static assets (favicon, logo)

internal/web/
├── handler/
│   └── schema.go         ← SchemaHandler: one Fiber handler for all /schema/* routes
├── ui/
│   └── pipeline.go       ← Priority constants, DataKey constants, IO types
├── stages/
│   ├── session.go        ← SessionStage (P=10)
│   ├── authz.go          ← AuthzStage (P=20)
│   ├── cache.go          ← CacheLookupStage (P=30) + CacheStoreStage (P=80)
│   ├── registry.go       ← RegistryStage (P=40)
│   ├── compile.go        ← CompileStage (P=50)
│   ├── normalize.go      ← NormalizeStage (P=60)
│   ├── validate.go       ← ValidateStage (P=70)
│   └── response.go       ← ResponseStage (P=90)
├── registry/
│   └── registry.go       ← PageRegistration + RegisterPage() + GetRegistration()
├── authz/
│   └── service.go        ← UIAuthzService, AllUIPermissions, fingerprints
├── amis/
│   ├── builder.go        ← M, A, Schema, Ctx, SchemaFn (legacy)
│   ├── page.go           ← Page, Grid, Panel, Tabs, Chart, Alert, Tpl builders
│   ├── crud.go           ← CRUD, Column, CreateBtn, EditBtn, ViewBtn, DeleteBtn
│   ├── form.go           ← Form, Wizard, all field helpers + modifiers
│   └── app.go            ← App shell builder (NavGroup, NavLink)
├── ast/                  ← Typed AST node system (ASTPageFn compilation target)
├── dsl/
│   ├── screens/          ← High-level screen builders (CRUDScreen, FormScreen, etc.)
│   └── blocks/           ← ~26 reusable UIBlock fragments
└── pages/                ← Page schema functions (one package per page)
```

---

## Shell Architecture

The shell (`web/pages/index.html`) is a single self-contained HTML file. No npm. No build step. No framework.

### HTML Structure

```html
<div id="app">
  <aside id="sidebar">
    .sidebar-brand      ← Logo + collapse toggle
    .sidebar-search     ← Search (cosmetic — no implementation yet)
    nav.sidebar-menu    ← Rendered by renderMenu() from menuConfig
  </aside>

  <div id="sidebar-backdrop">  ← Mobile overlay

  <div id="main">
    <div id="topbar">    ← Breadcrumb + theme toggle
    <div id="content">   ← AMIS embeds here
    <div id="footer">
  </div>
</div>
```

### Navigation (menuConfig)

Navigation is a JS array in `index.html`. Items must be added here manually:

```javascript
var menuConfig = [
  {
    type: 'item',
    id:    'dashboard',
    label: 'Dashboard',
    icon:  'fa-chart-line',
    hash:  '#dashboard'
  },
  {
    type:     'group',
    id:       'finance',
    label:    'Finance',
    icon:     'fa-calculator',
    expanded: false,
    items: [
      { id: 'invoices', label: 'Invoices', hash: '#invoices' }
    ]
  }
];
```

**`id` maps directly to the schema route:** `id: 'invoices'` → `GET /schema/invoices`

### Theme System

Three modes: `system | light | dark`. Stored in `localStorage('awo-theme')`. `system` follows OS preference. Applied by toggling `html.dark` class. See [Dark Mode](../05-dark-mode.md) for full CSS details.

---

## Envelope Formats

Two formats coexist. The AMIS fetcher bridges between them.

**Schema endpoints** (`/schema/*`) — return AMIS envelope directly:
```json
{ "status": 0, "data": { "type": "page", ... } }
```

**Data API endpoints** (`/api/v1/*`) — Go handlers return:
```json
{ "success": true, "data": [...], "meta": { "pagination": { "total_records": 142 } } }
```

The fetcher translates data API responses to AMIS format. See [API Contracts](../04-reference/02-api-contracts.md).

---

## Known Architecture Issues (from formal audit 2026-05-17)

These are tracked issues in the implementation, not the docs.

### Critical (production-blocking)

| # | Issue | Location | Fix |
|---|-------|----------|-----|
| C1 | `API_BASE = 'http://localhost:8080/'` hardcoded | `index.html:1040` | Remove — use relative URLs |
| C2 | 401 → login redirect missing from fetcher | `index.html:1314` | Add `if res.status === 401 → /login` |
| C4 | XSS in breadcrumb: `innerHTML` with interpolated string | `index.html:1272` | Use `textContent` |
| H7 | Font Awesome on external CDN, no SRI | `index.html:11` | Bundle in `web/sdk/` |

### High (architectural)

| # | Issue | Impact |
|---|-------|--------|
| H1 | Docs say schemas are "future Go-driven" — they already are | Documentation lag, confuses new devs |
| H2 | `max-height: 500px` CSS ceiling on menu groups | Clips at ~12 items; fails at ERP scale |
| H6 | No CSRF protection | State-mutating requests forgeable |
| H8 | `alert(msg)` native browser dialog in `amisEnv` | Freezes JS thread on AMIS errors |

### Medium (scalability)

| # | Issue | Impact |
|---|-------|--------|
| M3 | `collapseSidebar()` fires on every `#main` click | Sidebar collapses on every table row click |
| M4 | Sidebar search is cosmetic (no implementation) | Users click, nothing happens |
| M5 | Single-open accordion hostile to ERP multi-module workflows | 3+ clicks to cross-module navigate |

---

## Decision Status

Decisions from `27-decisions-and-opinions.md`:

| Decision | Original State | Current State |
|----------|---------------|---------------|
| 1: Custom shell vs AMIS `app` | Custom HTML — migrate when nav needs Go-driving | **Trigger met** (navigate() calls /schema/). Shell migration should happen before module #10. |
| 2: Static JSON vs Go-driven schemas | Static files — switch when flags/perms needed | **Already switched** — navigate() fetches /schema/ |
| 3: One envelope format | Two formats, fetcher bridges | Data APIs standardised on AMIS envelope |
| Drill-down navigation | Dialog vs URL — unresolved | Still unresolved — mandating choice before first detail page |

---

## What Is NOT in the Architecture

These features are documented aspirations, not current implementation:

- Navigation driven by Go (permissions filtering nav items) → `[ROADMAP]`
- AMIS `app` component replacing custom shell → `[ROADMAP]`
- Sidebar search → `[ROADMAP]`
- Notification system → `[ROADMAP]`
- Saved views / filter persistence → `[ROADMAP]`
- Tenant-configurable branding → `[ROADMAP]`

See [Planned Features](../05-roadmap/02-planned.md) for implementation status of each.
