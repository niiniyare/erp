# AWO ERP — UI Block System Reference

> **Version:** 2.0 | **Last Updated:** 2026-05-24
> **Scope:** `internal/web/dsl/blocks/` · `internal/web/ast/` · `web/schemas/pages/`
> **Renderer:** [Baidu amis v3.x](https://aisuda.bce.baidu.com/amis/zh-CN/docs/index) — `cxd` theme (default)
> **Design Influence:** Factumsoft/LynxERP aesthetic adapted to amis-ui standards
> **Mode:** Light mode only — dark mode deferred to v2.x

---

## Table of Contents

- [Preface](#preface)
- [Part 1 — Design System](#part-1--design-system)
  - [1.1 Design Philosophy](#11-design-philosophy)
  - [1.2 Colour System](#12-colour-system)
  - [1.3 Typography & Density](#13-typography--density)
  - [1.4 Spacing & Layout Grid](#14-spacing--layout-grid)
  - [1.5 Icons](#15-icons)
  - [1.6 Status Communication](#16-status-communication)
  - [1.7 CSS Override Strategy](#17-css-override-strategy)
  - [1.8 Performance Principles](#18-performance-principles)
- [Part 2 — Amis Fundamentals](#part-2--amis-fundamentals)
  - [2.1 The Schema Model](#21-the-schema-model)
  - [2.2 The Data Chain (数据链)](#22-the-data-chain-数据链)
  - [2.3 API Object Convention](#23-api-object-convention)
  - [2.4 Expression Language](#24-expression-language)
  - [2.5 Go Compilation Model](#25-go-compilation-model)
- [Part 3 — Foundation Components](#part-3--foundation-components)
  - [3A — Layout & Container Types](#3a--layout--container-types)
  - [3B — Display & Feedback Types](#3b--display--feedback-types)
  - [3C — Navigation Types](#3c--navigation-types)
  - [3D — Action Types](#3d--action-types)
  - [3E — Overlay Types](#3e--overlay-types)
  - [3F — Data Source Types](#3f--data-source-types)
  - [3G — Form Field Types](#3g--form-field-types)
- [Part 4 — Composite Blocks](#part-4--composite-blocks)
  - [4A — Page Shell Blocks](#4a--page-shell-blocks)
  - [4B — Filter & Search Blocks](#4b--filter--search-blocks)
  - [4C — Data List Blocks](#4c--data-list-blocks)
  - [4D — Dashboard & Analytics Blocks](#4d--dashboard--analytics-blocks)
  - [4E — Document Form Blocks](#4e--document-form-blocks)
  - [4F — Report Blocks](#4f--report-blocks)
  - [4G — Navigation & Action Blocks](#4g--navigation--action-blocks)
- [Part 5 — Page Composition Templates](#part-5--page-composition-templates)
- [Part 6 — Runtime Contract](#part-6--runtime-contract)
- [Part 7 — Performance & Caching Reference](#part-7--performance--caching-reference)
- [Part 8 — Developer Guide](#part-8--developer-guide)
- [Appendix A — Block Inventory](#appendix-a--block-inventory)
- [Appendix B — Accessibility Checklist](#appendix-b--accessibility-checklist)
- [Appendix C — Dark Mode (Deferred)](#appendix-c--dark-mode-deferred)

---

## Preface

### What This Document Is

This is the **single authoritative reference** for all UI work in AWO ERP. It covers the design system, every amis component type used in the codebase, every reusable block function, page composition patterns, and the full runtime contract between Go handlers and the amis renderer.

If you are building a new page or block, read this document first. If something you need is not documented here, add it here before shipping.

### What AWO ERP UI Is Not

AWO ERP is not a design showcase. It is a tool that accountants, station managers, and finance directors use for six to eight hours per day. The design system optimises for:

- **Predictability** — users form habits; the same action always works the same way
- **Data density** — more information visible per screen, fewer clicks per workflow
- **Speed** — pages load fast; forms validate immediately; reports compute in under a second
- **Legibility** — high contrast, clear labels, no ambiguity in status indicators

Visual distinctiveness is a secondary concern. Where the Factumsoft/LynxERP aesthetic and amis defaults conflict, **amis defaults win** — they come with battle-tested accessibility, keyboard navigation, and upgrade compatibility. Factumsoft aesthetic elements are applied only where they improve data density or legibility without overriding component behaviour.

### The Compilation Pipeline

AWO ERP never hand-writes amis JSON. Go functions return `ast.Node` values that compile to JSON:

```
Screen function (Go)
  └─ returns ast.Node tree
        └─ ast.CompileTree(root)
              └─ JSON schema (served at /schema/<page>)
                    └─ amis SDK in browser renders the page
```

Critical rules that follow from this:

- **Never write raw JSON** in screen files — use `ast.*` struct types and block functions.
- **Permission gates happen in Go** before JSON is produced. Unauthorised actions are structurally absent from the schema — never sent to the browser, never hidden with CSS.
- **Blocks are pure functions.** No HTTP calls, no database calls, no IAM calls inside a block. The `UISessionContext` carries all permission and session data the block needs.

---

## Part 1 — Design System

### 1.1 Design Philosophy

AWO ERP adapts the **Factumsoft/LynxERP** aesthetic within amis-ui's default `cxd` theme. The Factumsoft design (Behance gallery 59752763) is characterised by:

- A **dark navigation sidebar** (`#001529`) contrasting with a **light content area** (white `#FFFFFF`)
- **Compact, high-density data tables** with tight row padding and clear column separators
- **Chip-based quick filters** — horizontal strips of selectable tag chips above data tables for instant status filtering without opening a filter form
- **Clean flat panels** — card surfaces with a subtle border (`1px solid #e8e8e8`) rather than heavy drop shadows
- **Semantic colour restricted to status** — the base UI is monochromatic (greys); colour appears only for status badges, alerts, and directional trend indicators
- **Font Awesome icons throughout** — consistent icon vocabulary, no custom icon fonts

Within amis standards, these translate to:

| Factumsoft Design Element | amis Implementation | Notes |
|---|---|---|
| Dark sidebar | `nav` inside `page.aside`, `className: "erp-nav"`, CSS variable override | 2-variable override only |
| Compact table rows | `crud` with density `className` toggle via `localStorage` | No amis internals overridden |
| Chip quick filters | `form` (inline) + `input-tag` or `checkboxes` (inline) with `submitOnChange: true` | Native amis components |
| Flat panels | `panel` — amis default panel border is already subtle; no override needed | — |
| Semantic status colour | `mapping` component with amis `level` values | Use `level`, never custom colours |
| Font Awesome icons | `icon` component with `vendor: ""` | Prevents amis's default prefix collision |

The design is **not** a pixel-perfect reproduction of Factumsoft. It inherits the information architecture philosophy — dense, structured, navigation-first — while remaining entirely within amis component boundaries.

---

### 1.2 Colour System

AWO ERP uses the amis `cxd` default colour palette without modification for all interactive elements. The palette is extended with two overrides for structural surfaces.

#### amis `cxd` default palette (do not override these)

| Role | Hex | Used for |
|---|---|---|
| Primary | `#1677ff` | Buttons (primary level), links, focus rings, active nav items |
| Success | `#52c41a` | Status: Paid, Active, Approved, Completed |
| Warning | `#fa8c16` | Status: Pending, Draft, Partial, Suspended |
| Danger | `#f5222d` | Status: Overdue, Rejected, Voided, Error |
| Info | `#1677ff` (blue) | Status: Sent, In-Progress, Informational |
| Default | `#d9d9d9` / `#666` text | Status: Neutral, Archived, N/A |
| Text primary | `#262626` | Body text, labels |
| Text secondary | `#8c8c8c` | Hints, metadata, secondary labels |
| Border | `#d9d9d9` | Input borders, table dividers, panel borders |
| Background | `#f5f5f5` | Page background behind panels |
| Surface | `#ffffff` | Panel/card surfaces |

#### AWO structural overrides (only these two)

| Surface | Default | AWO Override | Reason |
|---|---|---|---|
| Sidebar background | `#ffffff` | `#001529` | Dark sidebar = Factumsoft pattern; reduces eye fatigue during long sessions; clear visual separation from content |
| Sidebar text (inactive) | — | `rgba(255,255,255,0.65)` | Passes WCAG AA contrast on `#001529` (ratio 4.7:1 ✓) |

These overrides are applied via two CSS custom properties in `web/static/theme/awo.css`:

```css
/* web/static/theme/awo.css — only file allowed to contain overrides */
:root {
  --erp-nav-bg:          #001529;
  --erp-nav-text:        rgba(255, 255, 255, 0.65);
  --erp-nav-text-active: #ffffff;
  --erp-nav-item-active: rgba(255, 255, 255, 0.08);
}

.erp-nav .cxd-Nav {
  background: var(--erp-nav-bg);
  color: var(--erp-nav-text);
}
.erp-nav .cxd-Nav-item.is-active {
  background: var(--erp-nav-item-active);
  color: var(--erp-nav-text-active);
}
```

**Rule: No other CSS overrides are permitted.** If a visual requirement cannot be met with amis properties and schema-level configuration, raise it for design review — do not add CSS.

#### Status colour usage rules

1. **Status colours are always paired with text or an icon.** Colour is never the sole carrier of meaning (accessibility and print legibility).
2. **Status colours appear only in `mapping` components and `alert` components.** Never set a custom colour on a `tpl` or `container` to communicate status.
3. **Variance directionality is computed server-side.** The API returns a `variance_class` field (`text-success` or `text-danger`) based on account type — revenue over budget is favourable (green), expenses over budget are adverse (red). The renderer displays without knowing the rule.

---

### 1.3 Typography & Density

#### Font stack

AWO ERP uses the system font stack — no web fonts are loaded:

```css
font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto,
             "Helvetica Neue", Arial, sans-serif, "Apple Color Emoji";
```

This is amis's default. It loads in 0ms (fonts are already installed on the user's device), eliminates FOUT (flash of unstyled text), and produces a native-looking interface that feels familiar on every OS. No web font will be added.

#### Density modes

AWO ERP supports two table density modes:

| Mode | Row padding | When to use |
|---|---|---|
| **Standard** | amis default (`12px` vertical) | New users, management, occasional users |
| **Compact** | `8px` vertical (Bootstrap `table-sm`) | Power users, accountants reviewing long trial balances |

Density is toggled per-user, stored in `localStorage` as `erp:tableDensity`, and applied as a single CSS class on the document root at application mount:

```javascript
// app-shell.js — runs once at mount
const density = localStorage.getItem('erp:tableDensity') || 'standard';
if (density === 'compact') {
  document.documentElement.classList.add('erp-compact');
}
```

```css
/* awo.css — the only CSS rule required */
.erp-compact .cxd-Table-table td {
  padding-top: 8px;
  padding-bottom: 8px;
}
```

No re-render, no API call. A toggle button in the table toolbar (`DataTableBlock` emits this automatically) calls `localStorage.setItem` and toggles the class.

---

### 1.4 Spacing & Layout Grid

AWO ERP uses the amis `grid` system, which maps to Bootstrap's 12-column MD breakpoints.

#### Standard column splits

| Page Type | Split | `md` values |
|---|---|---|
| Document form | Content / Sidebar | 8 / 4 |
| Two equal panels | Left / Right | 6 / 6 |
| Three dashboard cards | Equal thirds | 4 / 4 / 4 |
| Four KPI cards | Equal quarters | 3 / 3 / 3 / 3 |
| Settings / Master data | Nav / Content | 3 / 9 |
| Full-width report | Single column | 12 |

#### Three-region page layout

Every AWO ERP page uses this structure, rendered by `page` with `aside`:

```
┌──────────────────────────────────────────────────────────┐
│  SIDEBAR (fixed, dark, 220px)  │  CONTENT AREA (fluid)   │
│                                │                          │
│  [AWO Logo]                    │  ┌─ Page Header ───────┐ │
│                                │  │ Breadcrumb  Actions │ │
│  ▸ General Ledger              │  └─────────────────────┘ │
│  ▸ Receivables                 │                          │
│  ▸ Payables                    │  ┌─ Page Body ─────────┐ │
│  ▸ Inventory                   │  │                     │ │
│  ▸ Payroll                     │  │   [content]         │ │
│  ▸ Reports                     │  │                     │ │
│  ▸ Settings                    │  └─────────────────────┘ │
└──────────────────────────────────────────────────────────┘
```

The sidebar is always visible on desktop (≥ 768px). It does not collapse by default — persistent visibility builds navigation muscle memory faster than a toggle pattern. On mobile (< 768px) a bottom tab bar replaces the sidebar for the three mobile-optimised workflows.

#### Navigation depth rule

**Maximum two levels.** One level of groups (General Ledger, Receivables) and one level of leaf links within each group. No third level. This is a hard constraint — if a feature requires a third level, the information architecture must be restructured, not the rule relaxed.

Human short-term memory holds 7 ± 2 items. A two-level nav with 8 groups and 5–6 items per group = 40–48 total destinations, all reachable in two clicks. Three levels multiply that by 5× and exceed comfortable recall.

#### Panel spacing

`panel` components use amis default padding (`16px`). Do not add `className` for padding changes — if a panel feels cramped, the content needs restructuring, not a CSS patch.

---

### 1.5 Icons

AWO ERP uses **Font Awesome 5/6** loaded by the amis bundle. Always use the `icon` component with `vendor: ""` to prevent amis's automatic `fa-` prefix from doubling.

```go
// Correct
ast.IconNode{Icon: "fa fa-plus",    Vendor: ""}
ast.IconNode{Icon: "fa fa-coins",   Vendor: ""}
ast.IconNode{Icon: "fa fa-check",   Vendor: ""}

// Wrong — amis adds another fa- prefix
ast.IconNode{Icon: "plus"}
ast.IconNode{Icon: "fa-coins"}
```

#### Standard icon vocabulary

| Context | Icon class | Used in |
|---|---|---|
| Create / Add | `fa fa-plus` | New invoice, new journal, add line |
| Edit | `fa fa-pencil` | Edit record |
| Delete / Void | `fa fa-trash` | Void invoice, delete draft |
| Approve | `fa fa-check` | Approve action |
| Reject | `fa fa-times` | Reject action |
| Export | `fa fa-download` | Export CSV / PDF |
| Filter | `fa fa-filter` | Filter toggle |
| Search | `fa fa-search` | Search input prefix |
| Reports | `fa fa-chart-bar` | Report nav links |
| Revenue / Money | `fa fa-coins` | KPI cards, finance nav |
| Calendar | `fa fa-calendar` | Date fields, period nav |
| Attachment | `fa fa-paperclip` | Attachments block |
| Notes | `fa fa-sticky-note` | Internal notes block |
| Approval | `fa fa-tasks` | Approval workflow block |
| Warning | `fa fa-exclamation-triangle` | Alert components |
| History | `fa fa-history` | Activity feed block |

Do not invent new icons for concepts already covered above. Consistency across the application matters more than perfect semantic precision.

---

### 1.6 Status Communication

Every status indicator in AWO ERP must communicate through at least two channels: colour **and** text. Colour alone fails on print, on colour-deficient screens, and on low-quality mobile displays.

#### Status mapping standard

```go
// Standard status mappings — reuse these across all documents
var InvoiceStatusMappings = []blocks.StatusMapping{
    {Value: "draft",    Label: "Draft",    Color: "default"},
    {Value: "sent",     Label: "Sent",     Color: "info"},
    {Value: "partial",  Label: "Partial",  Color: "warning"},
    {Value: "paid",     Label: "Paid",     Color: "success"},
    {Value: "overdue",  Label: "Overdue",  Color: "danger"},
    {Value: "void",     Label: "Void",     Color: "warning"},
}
```

The amis `mapping` component renders these as coloured pill badges with text — both channels simultaneously. The `level` property maps to amis's semantic colours:

| `level` | amis colour | Use for |
|---|---|---|
| `success` | Green | Completed, Paid, Approved, Active |
| `warning` | Amber | Pending, Draft, Partial, Suspended, Void |
| `danger` | Red | Overdue, Rejected, Failed, Error |
| `info` | Blue | Sent, In-Progress, Informational |
| `default` | Grey | Neutral, Archived, Unknown |

**Rule: use `mapping`, never a disabled `select` for read-only status display.** `mapping` is the amis component specifically designed for value→coloured-label rendering. A disabled `select` renders as a greyed-out dropdown that looks broken.

---

### 1.7 CSS Override Strategy

AWO ERP's CSS policy has three tiers:

| Tier | What | Where | Permitted? |
|---|---|---|---|
| **Tier 1** | amis `cxd` theme defaults | amis bundle | Never override |
| **Tier 2** | Structural surface colours | `web/static/theme/awo.css` | 2 variable overrides only (sidebar bg, sidebar text) |
| **Tier 3** | Utility classes (density, print) | `web/static/theme/awo.css` | `.erp-compact` density class only |

The total custom CSS surface is four CSS rules. Every proposed addition to `awo.css` requires a design review justification. If a visual requirement needs more than four CSS rules, the approach is wrong — solve it at the schema and component-configuration level.

**Why this matters:** Every amis upgrade ships component improvements — better virtual scrolling, improved mobile tables, accessibility fixes. When custom CSS is minimal, amis upgrades apply cleanly. Heavy CSS customisation creates merge conflicts on every upgrade and causes the application to diverge from tested component behaviour.

---

### 1.8 Performance Principles

These six principles constrain how every block and page is built. They are not aspirational — they are requirements.

**Principle 1 — Never load what is not visible.**
Tabs use `mountOnEnter: true, unmountOnExit: false`. Accordions are `collapsed: true` by default. Modal content loads on open. Expandable table row sub-data loads on row expand. Nothing renders until the user asks for it.

**Principle 2 — Pre-compute everything that can be pre-computed.**
Financial reports read from the `period_account_balances` summary table (refreshed nightly at 02:00 EAT), not from raw `journal_entry_lines`. A P&L on a closed period runs in 100–400ms. On raw lines it would take 8–15 seconds.

**Principle 3 — Server-side paginate all transactional lists.**
`loadDataOnce: true` is banned on `crud` components that display transactional data (invoices, journals, transactions, GRNs). A station generating 45,000 transactions/month transferred client-side is 9MB per page load. Server-side pagination sends exactly 20 records = ~12KB.

**Principle 4 — Cache aggressively, invalidate specifically.**
See Part 7 for the full cache TTL table. Short version: options endpoints cache 60s, KPIs cache 1h, nav schema caches 24h. Nothing is cache-busted on page navigation.

**Principle 5 — List endpoints return only what the table displays.**
A 20-column table showing 50 rows of 5-field objects is 10× smaller than returning full 50-field records. Backend engineers own this — the frontend specifies which fields it needs per endpoint.

**Principle 6 — Background operations never block the UI.**
Payroll runs, large exports, and bulk operations use Temporal workflows. The user sees a progress state and can navigate away. Completion triggers a toast notification. The page never hangs.

#### Performance budget

| Metric | Target | How measured |
|---|---|---|
| Time to First Byte | < 200ms | Go handler P95 |
| Largest Contentful Paint | < 2.5s | Lighthouse, 3G throttle |
| Total Blocking Time | < 200ms | Lighthouse audit |
| amis bundle (gzipped) | < 720KB | Already fixed — amis bundle |
| AWO custom JS (gzipped) | < 30KB | Webpack bundle analyser |
| API calls per page | ≤ 5 | Chrome DevTools Network |
| Dashboard TTI | < 1.5s | Broadband, warm cache |
| P&L report (closed period) | < 500ms | API P50, pre-aggregated |
| P&L report (current period) | < 1.5s | API P50, partial real-time |
| Transaction list (server-paged) | < 800ms | API P50, 20 records |

---

## Part 2 — Amis Fundamentals

### 2.1 The Schema Model

Every amis page is a JSON tree. Each node has a mandatory `type` field naming the component. The amis renderer resolves `type` at runtime and mounts the corresponding React component.

```json
{
  "type": "page",
  "title": "Invoices",
  "body": {
    "type": "crud",
    "api": "/api/v1/finance/invoices",
    "columns": [
      { "name": "number", "label": "Invoice #" },
      { "name": "total",  "label": "Total", "type": "number" }
    ]
  }
}
```

`body` accepts a single node or an array of nodes (`SchemaCollection`). Container types create **data scope boundaries** — a child component can read its parent's data but cannot write back to it.

In Go, this maps directly to struct types:

```go
ast.PageNode{
    Title: "Invoices",
    Body: ast.CRUDNode{
        API:     ast.APISpec{Method: "get", URL: "/api/v1/finance/invoices"},
        Columns: []ast.TableColumn{
            {Name: "number", Label: "Invoice #"},
            {Name: "total",  Label: "Total", Type: "number"},
        },
    },
}
```

---

### 2.2 The Data Chain (数据链)

Amis implements a **hierarchical data scope** model. When a component evaluates `${foo}`, amis walks up the scope chain until it finds `foo` or reaches the root.

#### Scope-creating components

| Component | Creates scope? | Scope content |
|---|---|---|
| `page` | Yes | Root scope — initial data from `initApi` response + static `data` map |
| `service` | Yes | Merges its `api` response into its own scope |
| `form` | Yes | Form field values; this is what gets POSTed on submit |
| `crud` | Yes | Each data row creates a row-level scope |
| `dialog` | Yes | Isolated — only receives explicitly passed data |
| `combo` | Yes (per row) | Each repeat row has its own scope |

#### Fixed scope variables (backend must always populate these)

| Variable | Type | Populated by | Consumed by |
|---|---|---|---|
| `${id}` | `string (UUID)` | URL param / CRUD row scope | Row actions, activity feed |
| `${record}` | `object` | Detail page `initApi` | `DetailCardBlock`, form default values |
| `${totals}` | `[]{ label, amount }` | Document `initApi` | `TotalsSummaryBlock` |
| `${tax_lines}` | `[]{ tax_name, rate_pct, taxable_amount, tax_amount }` | Document `initApi` | `TaxSummaryBlock` |
| `${breadcrumbs}` | `[]{ label, href }` | All pages `initApi` | `EntityBreadcrumbBlock` |
| `${can_approve}` | `bool` | Document page `Data` map | `ApprovalWorkflowBlock` |
| `${api_url}` | `string` | Report page `Data` map | Export links in `ReportTableBlock` |
| `${current_user_id}` | `string (UUID)` | All pages `Data` map | Ownership expressions |
| `${tenant_currency}` | `string` | All pages `Data` map | Currency formatting in KPI blocks |

#### Data chain lookup direction

```
combo row → form → service → crud → page (root)
```

To prevent a child from reading a parent variable, set `canAccessSuperData: false` on the child component. To shadow a parent variable in a dialog, set `"key": "__undefined"` in the `dialog.data` map.

---

### 2.3 API Object Convention

#### Standard AWO response envelopes

All API endpoints must return one of these three shapes. The `status: 0` field tells amis the request succeeded.

```json
// Single record (detail, create, update):
{
  "status": 0,
  "msg":    "ok",
  "data":   { ...record fields, plus scope variables like totals, breadcrumbs... }
}

// List (crud, service, select with remote source):
{
  "status": 0,
  "msg":    "ok",
  "data": {
    "items":           [...records...],
    "total":           1247,
    "hasFiltersApplied": false,
    "totalUnfiltered": 1247
  }
}

// Options (select source, picker source):
{
  "status": 0,
  "msg":    "ok",
  "data":   [{ "label": "Display Name", "value": "uuid-or-code" }]
}

// Error:
{
  "status": 1,
  "msg":    "Invoice not found"
}
```

The `hasFiltersApplied` and `totalUnfiltered` fields in list responses enable the three-type empty state system (§ 6.8). **All list endpoints must return these fields.**

#### `initApi` vs `api` vs `source`

| Property | Component | Timing | Purpose |
|---|---|---|---|
| `initApi` | `page`, `form`, `service` | On component mount | Load initial data into scope |
| `api` | `form`, `crud` | On submit or explicit refresh | Form POST/PUT; CRUD list fetch |
| `source` | `select`, `picker`, `input-tree` | On mount + tracked var change | Remote options loading |

#### API object properties used in AWO

| Property | Usage | Notes |
|---|---|---|
| `method` | `get` / `post` / `put` / `delete` | Match HTTP verb to semantics |
| `url` | `/api/v1/...` always absolute | May contain `${variable}` expressions |
| `sendOn` | Suppress send when a field is empty | Used on filter form fields |
| `cache` | `60000` on options endpoints | Prevents repeated options fetches |
| `data` | Override payload when scope defaults are insufficient | Always document overrides |
| `adaptor` | Not used in block layer | Adaptors belong in Go handlers |

---

### 2.4 Expression Language

Amis supports a lightweight expression language in string-valued properties. Expressions are enclosed in `${...}`.

#### Template expressions (string interpolation)

```
"${customer_name}"                     Simple variable
"${record.billing_city}"               Dot-notation traversal
"${amount | number: 2}"                Pipe filter — 2 decimal places
"${document_date | date: DD MMM YYYY}" Date formatting
"${qty * unit_price}"                  Arithmetic
"${status === 'active'}"               Comparison (boolean result)
"${!!id}"                              Double-negate: truthy check
"${!id}"                               Falsy check (create mode)
```

#### State control properties

| Property | Type | Effect |
|---|---|---|
| `visibleOn` | expression | Component renders when `true` |
| `hiddenOn` | expression | Component is hidden (DOM removed) when `true` |
| `disabledOn` | expression | Input is non-interactive when `true` |
| `required` | expression or `bool` | Field is required when `true` |

#### Standard expression patterns

| Purpose | Expression |
|---|---|
| Show only in edit mode (record exists) | `"${!!id}"` |
| Show only in create mode (new record) | `"${!id}"` |
| Lock fields after approval | `"${status === 'approved' \|\| status === 'paid'}"` |
| Gate approval button on permission | `"${!can_approve}"` |
| Conditional required | `"${show_shipping}"` |
| Disable when terms are auto-computed | `"${!!payment_terms_id}"` |

---

### 2.5 Go Compilation Model

#### ast.Node types map 1:1 to amis `type` values

```go
// ast/nodes.go — every amis type has a corresponding Go struct

type PageNode struct {
    Title   string
    InitAPI APISpec
    Data    map[string]any   // static values merged into page scope at mount
    Body    []Node
    Toolbar []Node           // top-right action area
    Aside   Node             // sidebar slot
}

type CRUDNode struct {
    API         APISpec
    PageSize    int
    PrimaryKey  string
    Columns     []TableColumn
    BulkActions []Node
    Expandable  *ExpandableConfig
    // ...
}
```

#### UISessionContext — what blocks receive

```go
type UISessionContext interface {
    Can(action, resource string) bool   // permission check
    UserID() string
    TenantID() string
    Currency() string                    // tenant default currency code
    FeatureEnabled(flag string) bool     // feature flag check
    Locale() string
}
```

Blocks call `sess.Can(...)` before deciding whether to include a node. If the user cannot perform an action, the node is never constructed — not hidden with CSS, not grayed out, not sent to the browser.

---

## Part 3 — Foundation Components

> The canonical catalogue of every amis `type` used in AWO ERP. Each entry shows the amis type, the Go `ast.*` counterpart, the key configuration options, and a usage example.

---

### 3A — Layout & Container Types

#### `page` — root shell

Every amis page starts with `page`. It renders the top bar (title, toolbar), optional aside (sidebar slot), and the main body area.

```go
ast.PageNode{
    Title:   "Invoice #${ref_number}",
    InitAPI: ast.APISpec{Method: "get", URL: "/api/v1/finance/invoices/${id}"},
    Data: ui.M{
        "can_approve":       sess.Can("approve", "invoice"),
        "current_user_id":   sess.UserID(),
        "tenant_currency":   sess.Currency(),
    },
    Body:    []ast.Node{ ... },
    Toolbar: []ast.Node{ ... },  // primary actions, top-right
}
```

Key properties:

| Property | Type | Purpose |
|---|---|---|
| `Title` | `string` | Page heading; may contain `${expr}` |
| `InitAPI` | `APISpec` | Fetches page data on mount; response merged into root scope |
| `Data` | `map[string]any` | Static values available immediately (before `initApi` resolves) |
| `Body` | `[]Node` | Main content area |
| `Toolbar` | `[]Node` | Top-right action bar |
| `Aside` | `Node` | Left panel slot; used for the sidebar `nav` |

**Filter context subtitle pattern (drill-down pages):** On financial report drill pages, add a `tpl` immediately after the breadcrumb to show which period and cost centre the user is looking at. This subtitle persists even when the filter panel is collapsed:

```go
ast.TplNode{
    Tpl:       "${periodName}${costCentreName ? ' · ' + costCentreName : ''}",
    ClassName: "text-muted erp-filter-context",
    VisibleOn: "${periodName}",
}
```

---

#### `grid` — N-column layout

```go
ast.GridNode{
    Columns: []ast.GridColumn{
        {MD: 8, Body: []ast.Node{ ...main content... }},
        {MD: 4, Body: []ast.Node{ ...sidebar content... }},
    },
}
```

`MD` values must sum to 12. Use the standard splits from §1.4. Do not use `grid` for single-column content — use the page `body` directly.

---

#### `panel` — card surface

`panel` is the universal card surface. Every visual "card" in AWO ERP emits a `panel`. It renders a bordered container with an optional header, body, and footer.

```go
ast.PanelNode{
    Header: "Line Items",
    Body:   []ast.Node{ ... },
    // Footer: rarely used — only for the primary submit button on simple single-panel forms
}
```

Amis default panel styling (subtle `1px solid #e8e8e8` border, white background, 16px padding) matches the Factumsoft flat-card aesthetic without any CSS override.

---

#### `tabs` — tabbed sections

**Critical configuration:** Always set `mountOnEnter: true, unmountOnExit: false` on every `tabs` node in AWO ERP. This is not the amis default — the default destroys and re-mounts tab content on every switch, causing redundant API calls.

```go
ast.TabsNode{
    TabsMode:      "line",        // line | card | radio — "line" matches Factumsoft style
    MountOnEnter:  true,          // render tab content only when first clicked
    UnmountOnExit: false,         // keep mounted after first visit (no re-fetch on switch-back)
    Tabs: []ast.Tab{
        {Title: "Details",   Body: []ast.Node{ detailForm }},
        {Title: "Lines",     Body: []ast.Node{ lineItems }},
        {Title: "Approvals", Body: []ast.Node{ approvals }},
        {Title: "History",   Body: []ast.Node{ activityFeed }},
    },
}
```

Why this matters: with `unmountOnExit: false`, switching from "Details" to "Lines" and back to "Details" does not re-fire the Details API. Each tab's data is fetched once per session and cached in memory. This saves approximately 22 redundant API calls per user per day.

Memory note: keeping 5 tabs mounted adds approximately 200–400KB heap for a typical detail record. Acceptable on desktop workstations. On low-RAM shared tablets, apply `unmountOnExit: true` to low-priority tabs only (Attachments, Audit Trail).

---

#### `collapse` — collapsible section

```go
ast.CollapseNode{
    Title:     "Internal Notes",
    Collapsed: true,    // always true in Awo blocks — collapsed by default
    Body:      []ast.Node{ notesField },
}
```

All `collapse` nodes emitted by AWO blocks default to `Collapsed: true`. Low-priority sections (notes, attachments, advanced settings) are hidden by default to reduce form height for the common case.

---

#### `flex` — row alignment

```go
ast.FlexNode{
    Justify: "space-between",
    Align:   "center",
    Items: []ast.Node{
        breadcrumb,
        buttonToolbar,
    },
}
```

Use `flex` for toolbar rows (breadcrumb left, actions right) and for inline label+value pairs in KPI cards. Do not use `flex` for multi-column page layouts — use `grid` for that.

---

#### `wizard` — multi-step workflow

```go
ast.WizardNode{
    API:  ast.APISpec{Method: "post", URL: "/api/v1/finance/journals"},
    Steps: []ast.WizardStep{
        {Title: "Journal Details", Body: []ast.Node{ headerFields }},
        {Title: "Line Items",      Body: []ast.Node{ lineItems }},
        {Title: "Review & Post",   Body: []ast.Node{ summary }},
    },
}
```

Use `wizard` for document creation workflows with three or more distinct stages where proceeding to the next stage requires completing the current one. Do not use `wizard` for simple forms — a single `form` with sections is simpler.

---

### 3B — Display & Feedback Types

#### `mapping` — status badge renderer

`mapping` is the correct component for rendering value → coloured label. Always use this for status display — never a disabled `select`, never a custom `tpl` with inline styles.

```go
ast.MappingNode{
    Value: "${status}",
    Map: map[string]ast.MappingItem{
        "draft":   {Label: "Draft",   Level: "default"},
        "sent":    {Label: "Sent",    Level: "info"},
        "paid":    {Label: "Paid",    Level: "success"},
        "overdue": {Label: "Overdue", Level: "danger"},
        "void":    {Label: "Void",    Level: "warning"},
    },
}
```

For accessibility, the HTML output must include an `aria-label`:

```json
{
  "type": "mapping",
  "map": {
    "overdue": "<span class='label label-danger' aria-label='Status: Overdue'><i class='fa fa-clock'></i> Overdue</span>"
  }
}
```

---

#### `property` — label:value detail card

`property` renders a two-column label:value grid for read-only record display. This is the correct component for `DetailCardBlock`. Do not construct this pattern with a custom `table`.

```go
ast.PropertyNode{
    Title: "Invoice Details",
    Column: 2,
    Items: []ast.PropertyItem{
        {Label: "Customer",  Content: "${customer_name}"},
        {Label: "Issued",    Content: "${document_date | date: DD MMM YYYY}"},
        {Label: "Total",     Content: "${total_amount | currency}"},
        {Label: "Status",    Content: "${status}", Type: "mapping"},
    },
}
```

---

#### `tpl` — template text

```go
ast.TplNode{
    Tpl: "Total: <strong>${total | number: 2}</strong> ${currency}",
}
```

Use `tpl` for formatted text display, currency amounts, and inline computed values. Avoid putting business logic in `tpl` expressions — if a computation is complex, compute it server-side and send a pre-formatted value.

---

#### `chart` — ECharts wrapper

```go
ast.ChartNode{
    API:    ast.APISpec{Method: "get", URL: "/api/v1/finance/reports/revenue-trend"},
    Height: 300,
    Config: map[string]any{
        "xAxis":  map[string]any{"type": "category"},
        "yAxis":  map[string]any{"type": "value"},
        "series": []map[string]any{{"type": "line"}},
    },
}
```

The backend chart API returns `{ "xAxis": { "data": [...] }, "series": [{ "name": "...", "data": [...] }] }` merged into the ECharts config.

For sparklines (mini trend charts in KPI cards), use bare `chart` with no axes, no legend, no tooltip — just the line shape for maximum information density at small size.

---

#### `alert` — banner notifications

```go
ast.AlertNode{
    Level:          "info",   // info | warning | danger | success
    Body:           "This report covers the selected period only.",
    ShowCloseButton: false,
}
```

Use `alert` for report context banners (`ReportHeaderBlock`) and feature-gated notices. Do not use for field validation — amis handles that automatically on `form` fields with `required` set.

---

#### `progress`, `badge`, `tag`, `sparkline`, `log`, `image`, `icon`, `json`

| Type | `ast.*` | Use |
|---|---|---|
| `progress` | `ast.ProgressNode` | Completion bars on period close checklists |
| `badge` | `ast.BadgeNode` | Overlay counter (wraps nav items with pending counts) |
| `tag` | `ast.TagNode` | Inline coloured chip tags in property displays |
| `sparkline` | `ast.SparklineNode` | 6-period mini trend in KPI cards |
| `log` | `ast.LogNode` | Streaming audit log feed in `ActivityFeedBlock` |
| `image` | `ast.ImageNode` | Product images, avatars |
| `icon` | `ast.IconNode` | Standalone Font Awesome icons; always set `vendor: ""` |
| `json` | `ast.JSONNode` | Collapsible JSON tree — dev/debug only, never in production screens |

---

### 3C — Navigation Types

#### `breadcrumb`

```go
ast.BreadcrumbNode{
    Source: "${breadcrumbs}",  // []{ label, href } populated by backend initApi
}
```

Rules: breadcrumb appears on every page. The last item has no `href` (current page). Backend populates `breadcrumbs` in `initApi` response.

---

#### `nav`

```go
ast.NavNode{
    Stacked:      true,
    SaveExpanded: true,            // persists group state in sessionStorage
    ClassName:    "erp-nav",       // applies the dark sidebar CSS
    Links: []ast.NavLink{
        {
            Label: "General Ledger",
            Icon:  "fa fa-book",
            Children: []ast.NavLink{
                {Label: "Chart of Accounts", To: "#/gl/coa",      Icon: "fa fa-sitemap"},
                {Label: "Journal Entries",   To: "#/gl/journals", Icon: "fa fa-pencil-square"},
                {Label: "Bank Accounts",     To: "#/gl/banks",    Icon: "fa fa-university"},
            },
        },
        {
            Label: "Receivables",
            Icon:  "fa fa-arrow-circle-right",
            Children: []ast.NavLink{
                {Label: "Invoices",   To: "#/ar/invoices"},
                {Label: "Customers",  To: "#/ar/customers"},
                {Label: "Receipts",   To: "#/ar/receipts"},
            },
        },
        // ...
    },
}
```

Nav group labels use **business vocabulary, not module names.** "Receivables" not "AR Module". "Payroll" not "HR Module". Labels match the language accountants use in their daily work.

The nav schema is served from a static JSON file cached on CDN for 24 hours. Role-based visibility is applied client-side using `visibleOn` expressions against `${currentUser.roles}` in the page root scope. This avoids per-role server-generated schemas and keeps the CDN cache effective.

---

#### `steps`, `anchor-nav`, `pagination`

| Type | `ast.*` | Use |
|---|---|---|
| `steps` | `ast.StepsNode` | Step indicator inside `wizard` pages |
| `anchor-nav` | `ast.AnchorNavNode` | Section-jump navigation on long settings forms |
| `pagination` | `ast.PaginationNode` | Page controls — auto-rendered inside `crud`; rarely used standalone |

---

### 3D — Action Types

#### `action` — the preferred action node

`action` is more powerful than `button` and should be used whenever an action navigates, opens an overlay, or calls an API. Use `button` only for form submit/cancel within a `form`.

```go
// Link action (navigate to record)
ast.ActionNode{
    Label:      "View",
    ActionType: "link",
    Level:      "link",
    Target:     "#/finance/invoices/${id}",
}

// Dialog action (open confirmation or short form)
ast.ActionNode{
    Label:      "Approve",
    ActionType: "dialog",
    Level:      "primary",
    Icon:       "fa fa-check",
    Dialog: ast.DialogNode{
        Title: "Approve Invoice",
        Size:  "sm",
        Body:  []ast.Node{ commentTextarea },
        Actions: []ast.Node{
            ast.ButtonNode{Label: "Cancel",  ActionType: "cancel",  Level: "default"},
            ast.ButtonNode{Label: "Approve", ActionType: "submit",  Level: "primary"},
        },
    },
}

// Ajax action (bulk operation, status change)
ast.ActionNode{
    Label:       "Void Invoice",
    ActionType:  "ajax",
    Level:       "danger",
    ConfirmText: "Void this invoice? This cannot be undone.",
    API:         ast.APISpec{Method: "post", URL: "/api/v1/finance/invoices/${id}/void"},
}
```

#### Button levels

| `level` | Colour | Use for |
|---|---|---|
| `primary` | Blue | Primary CTA — Save, Create, Approve, Submit |
| `warning` | Amber | Caution actions — Archive, Suspend, Send for Review |
| `danger` | Red | Destructive actions — Void, Delete, Reject |
| `default` | Grey | Secondary actions — Cancel, Back, Close |
| `link` | Underlined text | In-table navigation, breadcrumb-style row CTAs |

#### `button-group`, `button-toolbar`, `dropdown-button`, `link`

| Type | `ast.*` | Use |
|---|---|---|
| `button-group` | `ast.ButtonGroupNode` | Visually connected inline buttons (shared border) — Approve/Reject pairs |
| `button-toolbar` | `ast.ButtonToolbarNode` | Spaced button row for page toolbars |
| `dropdown-button` | `ast.DropdownButtonNode` | Button with expandable action menu |
| `link` | `ast.LinkNode` | Plain hyperlink text (non-button inline links) |

---

### 3E — Overlay Types

#### `dialog` — modal overlay

Use modals only when the interaction can be completed in under 60 seconds without needing to reference another part of the system. If the user might need to check another screen while in the dialog, use a full page route instead.

```go
ast.DialogNode{
    Title:  "Reject Invoice",
    Size:   "sm",   // sm | md | lg | xl
    Body:   []ast.Node{ reasonTextarea },
    Actions: []ast.Node{
        ast.ButtonNode{Label: "Cancel", ActionType: "cancel", Level: "default"},
        ast.ButtonNode{Label: "Reject", ActionType: "submit", Level: "danger"},
    },
}
```

Modal size guidelines:

| Content | `size` | Notes |
|---|---|---|
| Confirmation (text only) | `sm` | No input fields |
| Single action + comment | `sm` | ≤ 1 textarea |
| Quick form | `md` | ≤ 5 fields |
| Record preview (read-only) | `lg` | No field limit since read-only |
| **Never** | `xl` | Editable forms with sub-grids always become full pages |

For modals containing a CRUD with inner scroll, set `scrollable: true` on the inner CRUD and constrain its height to prevent the modal exceeding 80vh.

---

#### `drawer` — side panel overlay

```go
ast.DrawerNode{
    Title:    "Filter Options",
    Position: "right",   // left | right
    Size:     "md",
    Body:     []ast.Node{ advancedFilterForm },
}
```

Use `drawer` for supplementary context (advanced filters, help panels, detail previews) where the user needs to see the background page while the panel is open. The drawer does not block the underlying page content.

---

### 3F — Data Source Types

#### `service` — data fetcher

`service` fetches data on mount and merges the response into its own scope. Use it to wrap blocks that need independent data feeds separate from the page `initApi`.

```go
ast.ServiceNode{
    API:  ast.APISpec{Method: "get", URL: "/api/v1/audit/invoices/activity?limit=20"},
    Body: []ast.Node{ activityList },
}
```

For auto-refreshing panels (dashboard exceptions panel), add `Interval: 300000` (5 minutes in ms).

---

#### `crud` — the primary list component

`crud` manages list data: API calls, pagination, sorting, filtering, bulk actions, and row actions. **Never build a list page with raw `table` + manual pagination — always use `crud`.**

```go
ast.CRUDNode{
    API:        ast.APISpec{Method: "get", URL: "/api/v1/finance/invoices"},
    PageSize:   20,
    PrimaryKey: "id",
    HeaderToolbar: []any{
        "bulkActions",
        map[string]any{"type": "tpl", "tpl": "Total: ${total} records"},
        "export-excel",
    },
    FooterToolbar: []any{"statistics", "pagination"},
    AffixHeader:   true,     // sticky column headers on scroll
    Columns: []ast.TableColumn{
        {Name: "number",   Label: "Invoice #",  Sortable: true,  Fixed: "left"},
        {Name: "customer", Label: "Customer"},
        {Name: "amount",   Label: "Amount",     Type: "number"},
        {Name: "actions",  Label: "Actions",    Type: "operation", Fixed: "right"},
    },
    BulkActions: []ast.ActionNode{ ... },
}
```

Key properties:

| Property | Value | Notes |
|---|---|---|
| `PageSize` | `20` (default) | Never use `loadDataOnce: true` for transactional data |
| `AffixHeader` | `true` | Always — sticky headers on wide financial tables |
| `Fixed: "left"` | On identifier columns | Prevents identifier scrolling off-screen on wide tables |
| `Fixed: "right"` | On action column | Actions always visible regardless of horizontal scroll |

**Server-side pagination API contract** (backend must follow exactly):

```
GET /api/v1/finance/invoices?page=2&perPage=20&orderBy=dueDate&orderDir=asc&status=overdue

Response: {
  "status": 0,
  "data": {
    "total": 45721,          // total matching records (for pagination display)
    "items": [...],          // exactly perPage items
    "hasFiltersApplied": true,
    "totalUnfiltered": 48329
  }
}
```

**Expandable rows** — for sub-record preview without page navigation:

```go
ast.CRUDNode{
    // ...
    Expandable: &ast.ExpandableConfig{
        ExpandedRowRender: ast.CRUDNode{
            API:          ast.APISpec{Method: "get", URL: "/api/gl/journals/${id}/lines"},
            LoadDataOnce: true,   // cache sub-records in row scope — no re-fetch on re-expand
        },
    },
}
```

---

#### `form` — form container

```go
ast.FormNode{
    Mode:         "horizontal",    // horizontal | normal | inline
    LabelWidth:   160,             // px — fixed across all AWO forms
    ValidateOn:   "blur",          // validate immediately on field leave, not on submit
    API:          ast.APISpec{Method: "put", URL: "/api/v1/finance/invoices/${id}"},
    Body:         []ast.Node{ ...fields... },
    Actions: []ast.Node{
        ast.ButtonNode{Label: "Cancel", ActionType: "cancel", Level: "default"},
        ast.ButtonNode{Label: "Save",   ActionType: "submit", Level: "primary"},
    },
}
```

Form modes:

| `mode` | Layout | Use for |
|---|---|---|
| `horizontal` | Label left (160px), input right | All document edit forms — efficient use of desktop width |
| `inline` | Fields side by side in a row | Filter bars, report filter panels |
| `normal` | Label above input | Simple single-column forms (settings, preferences) |

`ValidateOn: "blur"` is set at the form level, not per-field. This ensures consistent validation timing across all fields in the form.

---

### 3G — Form Field Types

All form fields live inside a `form` body. The `name` property writes the value into the form data scope on every change.

#### Text & Number

| amis `type` | Purpose | Key config |
|---|---|---|
| `input-text` | Single-line text | `placeholder`, `clearable`, `prefix` (icon), `suffix` |
| `input-number` | Numeric input | `min`, `max`, `step`, `precision: 2` for currency amounts |
| `textarea` | Multi-line text | `maxLength`, `showCount: true`, `minRows: 3`, `maxRows: 8` |
| `input-password` | Password entry | `revealPassword: true` |
| `input-group` | Input with addon | Wraps `input-number` + `select` for amount-with-currency |

```go
// Currency input with unit selector (Factumsoft pattern)
ast.InputGroupNode{
    Body: []ast.Node{
        ast.InputNumberNode{Name: "amount",        Precision: 2, Min: 0},
        ast.SelectNode{Name:   "currency_code",    Source: "/api/v1/platform/currencies/options",
                       Cache: 60000, Width: 80},
    },
}
```

---

#### Date & Time

| amis `type` | Purpose | Format |
|---|---|---|
| `input-date` | Date picker | `"YYYY-MM-DD"` |
| `input-datetime` | Date + time | `"YYYY-MM-DD HH:mm"` |
| `input-time` | Time only | `"HH:mm"` |
| `input-date-range` | From/to pair | `"YYYY-MM-DD"` with shortcut presets |
| `input-month` | Month picker | Report period filters |
| `input-quarter` | Quarter picker | Report period filters |
| `input-year` | Year picker | Report period filters |

Always set `format` and `valueFormat` explicitly — never rely on the amis locale default:

```go
ast.InputDateNode{
    Name:        "document_date",
    Label:       "Invoice Date",
    Format:      "DD MMM YYYY",   // display format to user
    ValueFormat: "YYYY-MM-DD",   // value written to form data scope
    Required:    true,
    Clearable:   false,
}
```

---

#### Select & Choice

| amis `type` | Purpose | Notes |
|---|---|---|
| `select` | Dropdown; static or remote | `source` (remote URL), `multiple`, `searchable`, `clearable`, `cache` |
| `checkboxes` | Checkbox group | `inline: true` for horizontal chips layout |
| `checkbox` | Single boolean | `option` label text |
| `radios` | Radio group | `inline: true` |
| `switch` | Toggle on/off | `onText`, `offText` |
| `input-tag` | Chip multi-select | `submitOnChange: true` for filter chips |
| `nested-select` | Cascader / hierarchy | COA account picker |
| `input-tree` | Tree-based picker | COA and org hierarchy |
| `picker` | Popup list picker | Entity selection (customer, supplier) |
| `transfer` | Shuttle list | Role → permission assignment |
| `condition-builder` | Visual filter builder | Advanced search block |

##### `select` with remote options

```go
ast.SelectNode{
    Name:       "customer_id",
    Label:      "Customer",
    Source:     "/api/v1/crm/customers/options",
    Searchable: true,
    Clearable:  true,
    Required:   true,
    Cache:      60000,   // 1-minute options cache
}
```

Options endpoint must return `[{ "label": "...", "value": "..." }]`.

##### `picker` — correct entity selector for long lists

Use `picker` (not `select`) when the entity list is long and the user needs multiple visible columns to identify the right record:

```go
ast.PickerNode{
    Name:      "supplier_id",
    Label:     "Supplier",
    ModalSize: "lg",
    PickerSchema: ast.CRUDNode{
        API:      ast.APISpec{Method: "get", URL: "/api/v1/procurement/suppliers"},
        PageSize: 10,
        Columns: []ast.TableColumn{
            {Name: "code",  Label: "Code"},
            {Name: "name",  Label: "Name"},
            {Name: "phone", Label: "Phone"},
        },
    },
    LabelField: "supplier_name",
    ValueField: "supplier_id",
}
```

##### `input-tag` — the Factumsoft chip filter component

This is the component that renders the horizontal chip strip above data tables. `submitOnChange: true` fires the parent CRUD refresh immediately on chip selection without a submit button:

```go
ast.InputTagNode{
    Name:           "status_filter",
    Options: []ast.SelectOption{
        {Label: "All",     Value: ""},
        {Label: "Draft",   Value: "draft"},
        {Label: "Sent",    Value: "sent"},
        {Label: "Overdue", Value: "overdue"},
        {Label: "Paid",    Value: "paid"},
    },
    SubmitOnChange: true,
    Clearable:      true,
}
```

---

#### Composite & Repeating

| amis `type` | Purpose | Notes |
|---|---|---|
| `combo` | Repeating inline field group | `multiple: true`, preferred for flat mobile layouts |
| `input-table` | Inline editable table | Column headers visible; preferred for commercial line items on desktop |
| `input-kv` | Key-value pair list | Settings / configuration forms |
| `group` | Horizontal field row | Aligns multiple fields side-by-side in one row |

##### `input-table` vs `combo` for line items

- Use `input-table` when column headers improve readability: invoices, purchase orders, journal entries.
- Use `combo` for mobile-optimised or flat repeating groups.
- Both write to the same `line_items` field name.

```go
// input-table (desktop commercial documents)
ast.InputTableNode{
    Name:      "line_items",
    Addable:   true,
    Removable: true,
    Columns: []ast.InputTableColumn{
        {Name: "product_code", Label: "Code",      Type: "input-text"},
        {Name: "description",  Label: "Description",Type: "input-text"},
        {Name: "qty",          Label: "Qty",        Type: "input-number", Min: 0},
        {Name: "unit_price",   Label: "Unit Price", Type: "input-number", Precision: 2},
        {
            Name:     "subtotal",
            Label:    "Subtotal",
            Type:     "formula",
            Expr:     "${qty * unit_price * (1 - discount_pct / 100)}",
            ReadOnly: true,
        },
    },
}
```

---

#### Utility Fields

| amis `type` | Purpose |
|---|---|
| `hidden` | Carries `id`, `tenant_id`, etc. — no visible render |
| `formula` | Client-side computed field; expression evaluated from scope on every change |
| `static` | Read-only display field inside a `form` |
| `input-rich-text` | Rich text editor (Froala / Quill) — used in `InternalNotesBlock` |
| `input-file` | File attachment upload — `multiple: true`, `drag: true` |

##### `formula` — live computed fields

```go
// Journal balance indicator
ast.FormulaNode{
    Name:    "_balance",
    Formula: "ARRAYREDUCE(line_items, (acc, l) => acc + (l.debit || 0) - (l.credit || 0), 0)",
    InitSet: true,
}

// Balance display (reads the _balance formula result)
ast.TplNode{
    Tpl: `<div class="${_balance === 0 ? 'text-success' : 'text-danger'}">
            ${_balance === 0 ? '✓ Balanced' : '⚠ Out of balance: ' + ABS(_balance)}
          </div>`,
}
```

Formula evaluation is client-side JavaScript — no API call. On a journal with 20 lines, the formula iterates 20 objects in approximately 0.1ms — imperceptible. Shows the balance state in real-time as the accountant enters each line.

---

## Part 4 — Composite Blocks

> Every block is a Go function in `internal/web/dsl/blocks/` returning `ast.Node`. Blocks are pure functions — no I/O, no IAM calls. This section documents each block's config struct, the amis types it emits, and complete usage examples.

---

### 4A — Page Shell Blocks

#### `PageHeaderBlock`

**Source:** `internal/web/dsl/blocks/page_header.go`
**Emits:** `flex` containing `breadcrumb` (left) + `button-toolbar` (right)

The standard top-of-page header row. Used on all list and document pages. Breadcrumb reads `${breadcrumbs}` from page scope. Actions are permission-filtered in Go before the node is built.

```go
blocks.PageHeaderBlock(sess, blocks.PageHeaderConfig{
    Actions: []blocks.QuickAction{
        {
            Label:      "New Invoice",
            URL:        "#/finance/invoices/new",
            Permission: "invoice.create",
            Icon:       "fa fa-plus",
            Level:      "primary",
        },
    },
})
```

---

### 4B — Filter & Search Blocks

#### `FilterBarBlock`

**Source:** `internal/web/dsl/blocks/filter_bar.go`
**Emits:** `form` (`mode: "inline"`) containing selected field types

The universal filter header for listing pages. Never write a custom filter form in a screen file — always use this block. All filter values are appended as query parameters to the `crud` API call.

Active filters are always shown as dismissible chips below the filter form, regardless of whether the form is expanded or collapsed. This prevents the "forgotten filter" problem where users misread filtered results as complete data.

```go
blocks.FilterBarBlock(sess, blocks.FilterBarConfig{
    ShowSearch:        true,
    SearchPlaceholder: "Search invoices…",
    ShowDateRange:     true,
    ShowStatus:        true,
    StatusOptions: []ast.SelectOption{
        {Label: "Draft",   Value: "draft"},
        {Label: "Sent",    Value: "sent"},
        {Label: "Paid",    Value: "paid"},
        {Label: "Overdue", Value: "overdue"},
    },
    ShowEntityPicker: true,
    EntityURL:        "/api/v1/crm/customers/options",
    EntityFieldName:  "customer_id",
    EntityLabel:      "Customer",
    ShowAmountRange:  true,
    Collapsible:      true,
})
```

| Config Field | Type | Default | Purpose |
|---|---|---|---|
| `ShowSearch` | `bool` | `false` | Keyword search — field name: `keywords` |
| `SearchPlaceholder` | `string` | `"Search…"` | Input placeholder |
| `ShowDateRange` | `bool` | `false` | `input-date-range` — field name: `date_range` |
| `ShowStatus` | `bool` | `false` | Multi-select status — requires `StatusOptions` |
| `StatusOptions` | `[]ast.SelectOption` | — | Required when `ShowStatus: true` |
| `ShowEntityPicker` | `bool` | `false` | Searchable remote select for a related entity |
| `EntityURL` | `string` | — | Options API endpoint for entity picker |
| `EntityFieldName` | `string` | `"entity_id"` | Form field name |
| `EntityLabel` | `string` | `"Entity"` | Picker label |
| `ShowCurrency` | `bool` | `false` | Currency select (platform currencies API) |
| `ShowAmountRange` | `bool` | `false` | `amount_min` + `amount_max` number inputs |
| `ShowTypeFilter` | `bool` | `false` | Type select — configure `TypeOptions`, `TypeFieldName` |
| `Collapsible` | `bool` | `false` | Wraps form in a `collapse` (`collapsed: false` by default) |

Filter changes debounce 300ms before triggering the CRUD API call — prevents a network request on every keystroke in the keyword search field.

**Saved filter presets API** — server-side storage so presets are accessible across devices:

```
GET    /api/users/me/filter-presets?context=ar-invoices
POST   /api/users/me/filter-presets  { context, name, filters: {...} }
DELETE /api/users/me/filter-presets/:id
```

---

#### `QuickFilterChipsBlock`

**Source:** `internal/web/dsl/blocks/quick_filter_chips.go`
**Emits:** `form` (`mode: "inline"`) + `input-tag`

The Factumsoft-style horizontal chip strip. Renders above the data table. Each chip is a filter shortcut. `submitOnChange: true` fires the CRUD refresh immediately on chip selection — no submit button needed.

```go
blocks.QuickFilterChipsBlock(sess, blocks.QuickFilterChipsConfig{
    FieldName: "status_filter",
    Options: []ast.SelectOption{
        {Label: "All",     Value: ""},
        {Label: "Draft",   Value: "draft"},
        {Label: "Sent",    Value: "sent"},
        {Label: "Overdue", Value: "overdue"},
        {Label: "Paid",    Value: "paid"},
    },
    Multiple: false,   // false = single active chip; true = multi-chip filter
})
```

This is one of the most visible Factumsoft design patterns adapted into amis. The `input-tag` component with `submitOnChange: true` reproduces it exactly using a native amis component — no custom code.

---

#### `AdvancedSearchBlock`

**Source:** `internal/web/dsl/blocks/advanced_search.go`
**Emits:** `collapse` + `condition-builder`

Full expand/collapse advanced filter panel using amis's native `condition-builder` component. Users build field-operator-value filter rows visually. The result posts as structured JSON to the API.

```go
blocks.AdvancedSearchBlock(sess, blocks.AdvancedSearchConfig{
    Fields: []ast.ConditionField{
        {Name: "amount",        Label: "Amount",   Type: "number"},
        {Name: "status",        Label: "Status",   Type: "select", Options: statusOptions},
        {Name: "document_date", Label: "Date",     Type: "date"},
        {Name: "customer_name", Label: "Customer", Type: "text"},
    },
})
```

Collapsed by default. Toggle label: "Advanced Search" ↔ "Hide".

---

### 4C — Data List Blocks

#### `DataTableBlock`

**Source:** `internal/web/dsl/blocks/data_table.go`
**Emits:** `crud` with all standard table features pre-configured

Use this block for every listing page. It handles the toolbar, bulk actions, density toggle, empty state, and permission gating — never construct `ast.CRUDNode` directly in a screen file.

```go
blocks.DataTableBlock(sess, blocks.DataTableConfig{
    APIURL:     "/api/v1/finance/invoices",
    Title:      "Invoices",
    PageSize:   20,
    PrimaryKey: "id",
    Columns: []blocks.ColumnDef{
        {Name: "number",   Label: "Invoice #", Sortable: true, Fixed: "left"},
        {Name: "customer", Label: "Customer"},
        {Name: "amount",   Label: "Amount",    Type: "currency"},
        {Name: "due_date", Label: "Due",        Type: "date",    Sortable: true},
        blocks.StatusBadgeColumn(blocks.StatusBadgeConfig{
            FieldName: "status",
            Mappings:  invoiceStatusMappings,
        }),
    },
    AllowCreate: sess.Can("create", "invoice"),
    CreateURL:   "#/finance/invoices/new",
    AllowExport: sess.Can("export", "invoice"),
    ShowDensityToggle: true,    // emits Standard/Compact toggle button in toolbar
    BulkActions: []blocks.BulkActionDef{
        {
            Label:      "Bulk Archive",
            Permission: "invoice.delete",
            APIURL:     "/api/v1/finance/invoices/bulk-archive",
            APIMethod:  "post",
            Level:      "warning",
            Confirm:    "Archive selected invoices?",
        },
    },
    RowActions: []ast.ActionNode{
        {Label: "View",  ActionType: "link", Target: "#/finance/invoices/${id}"},
        {
            Label:     "Edit",
            ActionType: "link",
            Target:    "#/finance/invoices/${id}/edit",
            VisibleOn: `${status === 'draft'}`,
        },
    },
    EmptyState: blocks.EmptyStateConfig{
        Title:       "No invoices",
        Description: "Your outstanding customer invoices will appear here.",
        ActionLabel: "Create Invoice",
        ActionURL:   "#/finance/invoices/new",
    },
})
```

Column type reference:

| `Type` | Renders as | Notes |
|---|---|---|
| `"text"` | Plain string | Default |
| `"date"` | Formatted date | Uses `DD MMM YYYY` display format |
| `"number"` | Numeric with locale formatting | Comma thousands separator |
| `"currency"` | Symbol + 2 decimal places | Reads `${tenant_currency}` for symbol |
| `"status"` | Status badge | Use `StatusBadgeColumn` helper — emits `mapping` type |
| `"link"` | Hyperlink | `href` field in data |
| `"image"` | Thumbnail | `src` field in data |

`BulkActionDef.Permission` is checked **structurally in Go** — users without the permission never receive the button in the schema.

---

#### `StatusBadgeColumn`

**Source:** `internal/web/dsl/blocks/status_badge.go`
**Emits:** `ast.TableColumn` with `type: "mapping"`

```go
blocks.StatusBadgeColumn(blocks.StatusBadgeConfig{
    FieldName: "status",
    Label:     "Status",
    Mappings: []blocks.StatusMapping{
        {Value: "active",    Label: "Active",    Color: "success"},
        {Value: "suspended", Label: "Suspended", Color: "warning"},
        {Value: "archived",  Label: "Archived",  Color: "default"},
    },
})
```

---

#### `StatusBadgeBlock`

**Source:** `internal/web/dsl/blocks/status_badge.go`
**Emits:** `mapping` node (standalone, outside a table)

```go
blocks.StatusBadgeBlock(sess, blocks.StatusBadgeConfig{
    FieldName: "status",
    Mappings:  invoiceStatusMappings,
})
```

---

#### `EmptyStateBlock`

**Source:** `internal/web/dsl/blocks/empty_state.go`
**Emits:** `container` + `tpl` + optional `button`

AWO ERP uses three distinct empty state types. The block selects the appropriate one based on the `hasFiltersApplied` and `totalUnfiltered` fields in the CRUD API response:

**Type 1 — No data exists** (`total: 0, hasFiltersApplied: false`):
```
[icon: inbox]  No invoices yet
               Your outstanding customer invoices will appear here.
               [Create Invoice]
```

**Type 2 — Filters exclude all results** (`total: 0, hasFiltersApplied: true`):
```
[icon: funnel]  No results match your current filters
                247 invoices exist but your current filters show none of them.
                [Clear All Filters]
```

**Type 3 — Permission restricts view**:
```
[icon: lock]  No records visible
              You may not have permission to view records in this area.
```

```go
blocks.EmptyStateBlock(sess, blocks.EmptyStateConfig{
    Title:       "No invoices yet",
    Description: "Your outstanding customer invoices will appear here.",
    ActionLabel: "Create Invoice",
    ActionURL:   "#/finance/invoices/new",
})
```

---

#### `DetailCardBlock`

**Source:** `internal/web/dsl/blocks/detail_card.go`
**Emits:** `panel` + `property`

Read-only label:value detail card. Reads from `${record}` in page scope. Uses amis `property` component — not a custom table.

```go
blocks.DetailCardBlock(sess, blocks.DetailCardConfig{
    Title: "Invoice Details",
    Fields: []blocks.FieldDef{
        {Label: "Customer",   Key: "customer_name"},
        {Label: "Issued",     Key: "document_date",  Format: "date"},
        {Label: "Total",      Key: "total_amount",   Format: "currency"},
        {Label: "Tax",        Key: "tax_amount",     Format: "currency"},
        {Label: "Status",     Key: "status",         Format: "status"},
    },
})
```

`Format` values: `""` / `"text"` (default), `"date"`, `"currency"`, `"percent"`, `"number"`, `"status"` (renders a `mapping` inline).

---

### 4D — Dashboard & Analytics Blocks

#### `StatCardBlock`

**Source:** `internal/web/dsl/blocks/stat_card.go`
**Emits:** `panel` + `flex` (icon + value + label) + optional `sparkline`

```go
blocks.StatCardBlock(sess, blocks.StatCardConfig{
    Label:     "Total Revenue",
    ValueKey:  "total_revenue",
    Format:    "currency",
    Currency:  "",                 // defaults to sess.Currency()
    TrendKey:  "revenue_trend_pct",
    Trend:     blocks.TrendUp,     // TrendUp | TrendDown | TrendNeutral
    IconClass: "fa fa-coins",
    ShowSparkline: true,
})
```

`Trend` direction controls colour logic:
- `TrendUp` — green when `${trendKey} > 0`, red when negative (revenue: more is better)
- `TrendDown` — red when `${trendKey} > 0`, green when negative (expenses: less is better)
- `TrendNeutral` — no colour coding (purely informational KPIs)

---

#### `KPIRowBlock`

**Source:** `internal/web/dsl/blocks/kpi_row.go`
**Emits:** `grid` + N × `StatCardBlock`

```go
blocks.KPIRowBlock(sess, []blocks.StatCardConfig{
    {Label: "Revenue",    ValueKey: "revenue",    Format: "currency", Trend: blocks.TrendUp},
    {Label: "Invoices",   ValueKey: "inv_count",  Format: "number"},
    {Label: "Overdue",    ValueKey: "overdue_amt", Format: "currency", Trend: blocks.TrendDown},
    {Label: "Cash",       ValueKey: "cash_balance",Format: "currency"},
})
```

Column width = `12 / len(cards)`. Max 6 cards before wrapping to `md: 2` each. Four KPI cards is the standard dashboard layout (3/3/3/3).

---

#### `ChartPanelBlock`

**Source:** `internal/web/dsl/blocks/chart_panel.go`
**Emits:** `panel` + optional `select` (period picker) + `chart`

```go
blocks.ChartPanelBlock(sess, blocks.ChartPanelConfig{
    Title:        "Revenue by Month",
    ChartType:    blocks.ChartTypeLine,   // ChartTypeBar | ChartTypeLine | ChartTypePie
    APIURL:       "/api/v1/finance/reports/revenue-trend",
    PeriodPicker: true,   // adds Month/Quarter/Year selector above chart
    Height:       300,
})
```

Backend chart API must return:
```json
{
  "status": 0,
  "data": {
    "xAxis": { "data": ["Jan", "Feb", "Mar", ...] },
    "series": [{ "name": "Revenue", "data": [12000, 15000, ...] }]
  }
}
```

---

#### `ActivityPanelBlock`

**Source:** `internal/web/dsl/blocks/activity_panel.go`
**Emits:** `panel` + `service` + `list`

```go
blocks.ActivityPanelBlock(sess, blocks.ActivityPanelConfig{
    Title:    "Recent Transactions",
    Resource: "transactions",
    Limit:    10,
})
```

Calls `/api/v1/transactions/recent?limit=10`. Response: `[{ title, sub_title, avatar, meta }]`.

---

### 4E — Document Form Blocks

#### `DocumentHeaderBlock`

**Source:** `internal/web/dsl/blocks/document_header.go`
**Emits:** `group` containing `input-text`, `input-date`, optional `input-date`, optional `select` × 2

```go
blocks.DocumentHeaderBlock(sess, blocks.DocumentHeaderConfig{
    ResourceURL:  "/api/v1/finance/invoices",
    ShowDueDate:  true,
    ShowCurrency: true,
    ShowStatus:   true,
    StatusOptions: []ast.SelectOption{
        {Label: "Draft", Value: "draft"},
        {Label: "Sent",  Value: "sent"},
        {Label: "Paid",  Value: "paid"},
    },
    ReadOnly: !sess.Can("update", "invoice"),
})
```

| Field name | Type | Notes |
|---|---|---|
| `ref_number` | `input-text` | Always read-only — auto-generated by backend |
| `document_date` | `input-date` | Required |
| `due_date` | `input-date` | Shown when `ShowDueDate: true` |
| `currency_id` | `select` (remote) | Shown when `ShowCurrency: true` |
| `status` | `select` (static) | Shown when `ShowStatus: true` and `StatusOptions` non-empty |

---

#### `PartyBlock`

**Source:** `internal/web/dsl/blocks/party.go`
**Emits:** `panel` + `picker`

```go
// Use presets — never pass raw config unless building a custom entity type
blocks.PartyBlock(sess, blocks.DefaultCustomerConfig())
blocks.PartyBlock(sess, blocks.DefaultSupplierConfig())
blocks.PartyBlock(sess, blocks.DefaultEmployeeConfig())
```

| Preset | `FieldName` | `OptionsURL` |
|---|---|---|
| `DefaultCustomerConfig()` | `customer_id` | `/api/v1/crm/customers/options` |
| `DefaultSupplierConfig()` | `supplier_id` | `/api/v1/procurement/suppliers/options` |
| `DefaultEmployeeConfig()` | `employee_id` | `/api/v1/hr/employees/options` |

The `picker` popup renders a searchable `crud` with multiple visible columns for entity identification.

---

#### `AddressBlock`

**Source:** `internal/web/dsl/blocks/address.go`
**Emits:** `panel` + `collapse-group` + `input-text` × 4 + `select` (country)

```go
blocks.AddressBlock(sess, blocks.AddressConfig{
    ShowBilling:  true,
    ShowShipping: true,
    ReadOnly:     !sess.Can("update", "invoice"),
})
```

Field names: `billing_street`, `billing_city`, `billing_state`, `billing_postal_code`, `billing_country_id` (and `shipping_*` variants). Country options from `/api/v1/platform/countries/options` with `cache: 3600000` (1-hour cache).

---

#### `ProductServiceLineBlock`

**Source:** `internal/web/dsl/blocks/line_items.go`
**Emits:** `panel` + `input-table` (default) or `combo` (`FlatMode: true`)

**Every document with line items must use this block.** Never define a custom line item table in a screen file — domain changes must be made here once and propagate everywhere.

```go
blocks.ProductServiceLineBlock(sess, blocks.DefaultLineItemConfig())
blocks.ProductServiceLineBlock(sess, blocks.GRNLineItemConfig())
blocks.ProductServiceLineBlock(sess, blocks.JournalLineItemConfig())

// Custom
blocks.ProductServiceLineBlock(sess, blocks.LineItemConfig{
    ShowProductCode: true,
    ShowDescription: true,
    ShowQty:         true,
    ShowUOM:         true,
    ShowUnitPrice:   true,
    ShowDiscount:    true,
    ShowTaxRate:     true,
    ShowSubtotal:    true,
    MaxLines:        50,
    ReadOnly:        !sess.Can("update", "invoice"),
})
```

| Preset | Visible columns |
|---|---|
| `DefaultLineItemConfig()` | Code, Description, Qty, UOM, Unit Price, Discount %, Subtotal |
| `GRNLineItemConfig()` | Code, Description, Qty, UOM (no pricing) |
| `JournalLineItemConfig()` | Description, Account (nested-select), Debit, Credit |

The `subtotal` column is a `formula` field: `${qty * unit_price * (1 - discount_pct / 100)}`. Field name `line_items` is fixed — backend always expects this key.

---

#### `TotalsSummaryBlock`

**Source:** `internal/web/dsl/blocks/totals.go`
**Emits:** `panel` + `table` (static, `source: "${totals}"`)

```go
blocks.TotalsSummaryBlock(sess)
```

No config. Reads `${totals}` from page scope. Backend must populate:

```json
"totals": [
  { "label": "Subtotal",  "amount": 10000.00 },
  { "label": "Tax (16%)", "amount":  1600.00 },
  { "label": "Total",     "amount": 11600.00 }
]
```

---

#### `TaxSummaryBlock`

**Source:** `internal/web/dsl/blocks/tax_summary.go`
**Emits:** `panel` + `table` (static, `source: "${tax_lines}"`)

```go
blocks.TaxSummaryBlock(sess)
```

Reads `${tax_lines}`. Backend must populate:

```json
"tax_lines": [
  { "tax_name": "VAT", "rate_pct": 16, "taxable_amount": 10000.00, "tax_amount": 1600.00 }
]
```

---

#### `PaymentTermsBlock`

**Source:** `internal/web/dsl/blocks/payment_terms.go`
**Emits:** `group` + `select` (remote) + `input-date`

```go
blocks.PaymentTermsBlock(sess, blocks.PaymentTermsConfig{
    ReadOnly: !sess.Can("update", "invoice"),
})
```

Field names: `payment_terms_id`, `payment_due_date`. The `payment_due_date` field is disabled when `payment_terms_id` is set (auto-computed). Terms options from `/api/v1/finance/payment-terms/options`.

---

#### `InternalNotesBlock`

**Source:** `internal/web/dsl/blocks/notes.go`
**Emits:** `collapse` (`collapsed: true`) + `input-rich-text`

```go
blocks.InternalNotesBlock(sess)
```

No config. Field name: `internal_notes`. Max 2000 chars. Always collapsed by default — the common case is no notes, so the form stays compact.

---

#### `AttachmentsBlock`

**Source:** `internal/web/dsl/blocks/attachments.go`
**Emits:** `collapse` (`collapsed: true`) + `input-file`

```go
blocks.AttachmentsBlock(sess)
```

No config. Field name: `attachments`. `multiple: true`, `drag: true`. Always collapsed by default.

---

#### `ApprovalWorkflowBlock`

**Source:** `internal/web/dsl/blocks/approval.go`
**Emits:** `panel` (feature-flag gated) + `static` (`mapping`) + `textarea` + `button-group`

```go
blocks.ApprovalWorkflowBlock(sess)
// Feature flag "approval_workflow" checked internally.
// If flag disabled → renders collapsed placeholder; never nil.
// Approve/Reject disabledOn: "${!can_approve}"
// Set in page Data: "can_approve": sess.Can("approve", "invoice")
```

The approve button is disabled when `can_approve` is false. The `can_approve` variable must be set in the page's static `Data` map — it cannot be computed inside the block (blocks are pure functions of their inputs).

Approved decisions are recorded via signed HMAC-SHA256 tokens in approval emails:
```
token = HMAC-SHA256(key, { approvalStepId, action, approverId, expiresAt })
URL: /approvals/action?token=<token>   (expires 48h, single-use)
```

---

#### `ActivityFeedBlock`

**Source:** `internal/web/dsl/blocks/activity_feed.go`
**Emits:** `panel` + `service` + `log`

```go
blocks.ActivityFeedBlock(sess, blocks.ActivityFeedConfig{
    Title:        "Invoice History",
    Resource:     "invoices",
    ResourceID:   "${id}",
    ShowComments: true,
    Limit:        20,
    VisibleOn:    "${!!id}",   // hide on create mode — no document yet
})
```

Calls `/api/v1/audit/invoices/activity?resource_id=${id}&limit=20`.

---

### 4F — Report Blocks

> Use these on financial report pages (P&L, balance sheet, ledger). **Never** use `DataTableBlock` on a report page — reports have different data characteristics (grouped rows, subtotals, grand total, no CRUD operations).

#### URL drill navigation structure

Report drill-down uses URL-based page routes so the browser Back button works correctly and drill URLs are shareable:

```
/reports/pl?periodId=P2024-03&costCentreId=CC-FUEL
/reports/pl/drill/section?periodId=P2024-03&sectionId=REVENUE
/reports/pl/drill/account?periodId=P2024-03&accountId=ACC-4000
/reports/pl/drill/journal?periodId=P2024-03&journalId=JNL-8821
```

Level 1 (section) drill can open as a modal (quick reference). Level 2+ (account, journal) always opens as a full page route — the user is doing investigative work that may require copying journal numbers or comparing with other screens.

---

#### `ReportHeaderBlock`

**Source:** `internal/web/dsl/blocks/report_header.go`
**Emits:** `alert` (`level: "info"`) + `tpl` (title + description)

```go
blocks.ReportHeaderBlock(sess, blocks.ReportHeaderConfig{
    Title:       "Profit & Loss Statement",
    Description: "Net income for the selected period",
})
```

---

#### `ReportFilterBlock`

**Source:** `internal/web/dsl/blocks/report_filter.go`
**Emits:** `form` (inline) + period picker + optional entity, currency, comparison selects

```go
blocks.ReportFilterBlock(sess, blocks.ReportFilterConfig{
    ShowPeriod:       true,    // always true — all financial reports require a period
    ShowEntityPicker: false,
    ShowCurrency:     true,
    ShowComparison:   true,    // adds "Compare With" prior / prior year select
})
```

`ShowComparison: true` adds `compare_period` select: `"none"` (default), `"prior"`, `"prior_year"`.

---

#### `ReportTableBlock`

**Source:** `internal/web/dsl/blocks/report_table.go`
**Emits:** `table` (static, `source: "${rows}"`) + `button-toolbar` (export)

```go
blocks.ReportTableBlock(sess, blocks.ReportTableConfig{
    Columns: []blocks.ReportColumnDef{
        {Name: "account_name", Label: "Account", Type: "text",     Sortable: true},
        {Name: "debit",        Label: "Debit",   Type: "currency"},
        {Name: "credit",       Label: "Credit",  Type: "currency"},
        {Name: "balance",      Label: "Balance", Type: "currency", Sortable: true},
    },
    GroupBy:        []string{"account_type"},
    ShowSubtotals:  true,
    ShowGrandTotal: true,
    Exportable:     sess.Can("export", "report"),
})
```

Export buttons use `${api_url}/export?format=csv` and `${api_url}/export?format=pdf`. The page `Data` map must include `api_url`.

Report data source strategy:
- **Closed periods:** Read from `period_account_balances` pre-aggregated table (P50: 120ms, P95: 300ms)
- **Current open period:** UNION pre-aggregated closed + real-time aggregate of current period lines (P50: 400ms, P95: 800ms)
- **Drill level 2+ (raw journals):** Read from `journal_entry_lines` with index (P50: 300ms, P95: 700ms)

---

#### `ReportChartBlock`

**Source:** `internal/web/dsl/blocks/report_chart.go`
**Emits:** bare `chart` (no panel wrapper, no period picker)

```go
blocks.ReportChartBlock(sess, blocks.ReportChartConfig{
    Type:   blocks.ChartTypeLine,
    Height: 250,
})
```

Lighter than `ChartPanelBlock` — no card, no period selector. Placed directly between the filter and table on report pages. The `ReportFilterBlock` above it drives the period parameter.

---

### 4G — Navigation & Action Blocks

#### `EntityBreadcrumbBlock`

**Source:** `internal/web/dsl/blocks/entity_breadcrumb.go`
**Emits:** `breadcrumb` with `source: "${breadcrumbs}"`

```go
blocks.EntityBreadcrumbBlock(sess)
```

No config. Always present. Backend populates `breadcrumbs` in every `initApi` response.

---

#### `QuickActionsBlock`

**Source:** `internal/web/dsl/blocks/quick_actions.go`
**Emits:** `button-toolbar` + permission-filtered `action` nodes

```go
blocks.QuickActionsBlock(sess, []blocks.QuickAction{
    {
        Label:      "New Invoice",
        URL:        "#/finance/invoices/new",
        Permission: "invoice.create",
        Icon:       "fa fa-plus",
        Level:      "primary",
    },
    {
        Label: "Reports",
        URL:   "#/finance/reports",
        Icon:  "fa fa-chart-bar",
        // Permission: "" means always visible
    },
})
```

Actions with `Permission` not held by the session are structurally excluded — never sent to the browser. Empty `Permission` string = always visible.

---

#### `BulkActionsBlock`

**Source:** `internal/web/dsl/blocks/bulk_actions.go`
**Emits:** `[]ast.Node` — permission-filtered `button` nodes

```go
actions := blocks.BulkActionsBlock(sess, []blocks.BulkActionDef{
    {
        Label:      "Approve Selected",
        Permission: "invoice.approve",
        APIURL:     "/api/v1/finance/invoices/bulk-approve",
        APIMethod:  "post",
        Level:      "primary",
        Confirm:    "Approve all selected invoices?",
    },
    {
        Label:      "Void Selected",
        Permission: "invoice.delete",
        APIURL:     "/api/v1/finance/invoices/bulk-void",
        APIMethod:  "post",
        Level:      "danger",
        Confirm:    "Void selected invoices? This cannot be undone.",
    },
})
```

Returns empty slice (never nil) if no actions pass the permission check. Pass directly to `DataTableConfig.BulkActions`.

---

## Part 5 — Page Composition Templates

Templates are documented patterns showing how blocks and amis container types assemble into a complete `page` node. Use these as starting points — do not deviate from the structure without a documented reason.

### 5.1 List Page Template

```
page
  └─ PageHeaderBlock (breadcrumb + QuickActionsBlock)
  └─ QuickFilterChipsBlock   ← Factumsoft chip strip
  └─ FilterBarBlock
  └─ DataTableBlock
        └─ crud
              └─ placeholder → EmptyStateBlock
```

```go
func InvoiceListSchema(sess ui.UISessionContext) any {
    return ast.PageNode{
        Title:   "Invoices",
        InitAPI: ast.APISpec{Method: "get", URL: "/api/v1/finance/invoices/meta"},
        Data: ui.M{
            "current_user_id": sess.UserID(),
            "tenant_currency": sess.Currency(),
        },
        Body: []ast.Node{
            blocks.PageHeaderBlock(sess, blocks.PageHeaderConfig{
                Actions: []blocks.QuickAction{
                    {Label: "New Invoice", URL: "#/finance/invoices/new",
                     Permission: "invoice.create", Icon: "fa fa-plus", Level: "primary"},
                },
            }),
            blocks.QuickFilterChipsBlock(sess, blocks.QuickFilterChipsConfig{
                FieldName: "status_filter",
                Options:   invoiceStatusChips(),
            }),
            blocks.FilterBarBlock(sess, blocks.FilterBarConfig{
                ShowSearch:    true, SearchPlaceholder: "Search by number, customer…",
                ShowDateRange: true,
                ShowStatus:    true, StatusOptions: invoiceStatusOptions(),
                ShowAmountRange: true,
                Collapsible:   true,
            }),
            blocks.DataTableBlock(sess, blocks.DataTableConfig{
                APIURL:     "/api/v1/finance/invoices",
                PrimaryKey: "id",
                Columns:    invoiceColumns(sess),
                AllowCreate: sess.Can("create", "invoice"),
                CreateURL:   "#/finance/invoices/new",
                AllowExport: sess.Can("export", "invoice"),
                ShowDensityToggle: true,
                BulkActions: invoiceBulkActions(sess),
                RowActions:  invoiceRowActions(),
                EmptyState:  blocks.EmptyStateConfig{
                    Title:       "No invoices yet",
                    Description: "Your outstanding customer invoices will appear here.",
                    ActionLabel: "Create Invoice",
                    ActionURL:   "#/finance/invoices/new",
                },
            }),
        },
    }
}
```

---

### 5.2 Document Form Template

```
page (initApi: detail endpoint)
  └─ PageHeaderBlock
  └─ form (api: PUT endpoint, mode: horizontal, validateOn: blur)
        └─ grid 8/4
              ├─ col-8
              │     DocumentHeaderBlock
              │     PartyBlock
              │     ProductServiceLineBlock
              │     InternalNotesBlock
              └─ col-4
                    TotalsSummaryBlock
                    TaxSummaryBlock
                    AddressBlock
                    PaymentTermsBlock
                    ApprovalWorkflowBlock
                    AttachmentsBlock
  └─ ActivityFeedBlock (visibleOn: ${!!id})
```

```go
func InvoiceFormSchema(sess ui.UISessionContext) any {
    readOnly := !sess.Can("update", "invoice")
    return ast.PageNode{
        Title:   "Invoice #${ref_number}",
        InitAPI: ast.APISpec{Method: "get", URL: "/api/v1/finance/invoices/${id}"},
        Data: ui.M{
            "can_approve":     sess.Can("approve", "invoice"),
            "current_user_id": sess.UserID(),
            "tenant_currency": sess.Currency(),
        },
        Body: []ast.Node{
            blocks.PageHeaderBlock(sess, blocks.PageHeaderConfig{}),
            ast.FormNode{
                Mode:       "horizontal",
                LabelWidth: 160,
                ValidateOn: "blur",
                API:        ast.APISpec{Method: "put", URL: "/api/v1/finance/invoices/${id}"},
                Body: []ast.Node{
                    ast.GridNode{
                        Columns: []ast.GridColumn{
                            {MD: 8, Body: []ast.Node{
                                blocks.DocumentHeaderBlock(sess, blocks.DocumentHeaderConfig{
                                    ShowDueDate: true, ShowCurrency: true, ShowStatus: true,
                                    StatusOptions: invoiceStatusOptions(), ReadOnly: readOnly,
                                }),
                                blocks.PartyBlock(sess, blocks.DefaultCustomerConfig()),
                                blocks.ProductServiceLineBlock(sess, blocks.DefaultLineItemConfig()),
                                blocks.InternalNotesBlock(sess),
                            }},
                            {MD: 4, Body: []ast.Node{
                                blocks.TotalsSummaryBlock(sess),
                                blocks.TaxSummaryBlock(sess),
                                blocks.AddressBlock(sess, blocks.AddressConfig{
                                    ShowBilling: true, ShowShipping: true, ReadOnly: readOnly,
                                }),
                                blocks.PaymentTermsBlock(sess, blocks.PaymentTermsConfig{
                                    ReadOnly: readOnly,
                                }),
                                blocks.ApprovalWorkflowBlock(sess),
                                blocks.AttachmentsBlock(sess),
                            }},
                        },
                    },
                },
                Actions: []ast.Node{
                    ast.ButtonNode{Label: "Cancel", ActionType: "cancel", Level: "default"},
                    ast.ButtonNode{
                        Label: "Save", ActionType: "submit", Level: "primary",
                        DisabledOn: boolExpr(readOnly),
                    },
                },
            },
            blocks.ActivityFeedBlock(sess, blocks.ActivityFeedConfig{
                Title: "Invoice History", Resource: "invoices",
                ResourceID: "${id}", Limit: 20, VisibleOn: "${!!id}",
            }),
        },
    }
}
```

---

### 5.3 Dashboard Template

```
page (initApi: dashboard stats endpoint)
  └─ KPIRowBlock (4 cards: Revenue, Invoices, Overdue, Cash)
  └─ grid 8/4
        ├─ col-8: ChartPanelBlock (revenue trend)
        └─ col-4: ActivityPanelBlock (recent transactions)
  └─ QuickActionsBlock
```

Dashboard performance budget: 5 API calls total, ~14KB payload, TTI < 1.5s broadband / < 4s 3G.

---

### 5.4 Financial Report Template

```
page
  └─ ReportHeaderBlock
  └─ ReportFilterBlock
  └─ ReportChartBlock (optional)
  └─ ReportTableBlock
```

The `api_url` static data field must be set to the report API endpoint — `ReportTableBlock` appends `?format=csv` / `?format=pdf` for export links.

---

### 5.5 Master Data / Settings Template

```
page
  └─ grid 3/9
        ├─ col-3: nav (section links, stacked)
        └─ col-9: DataTableBlock or tabs + form
```

Used for Chart of Accounts, Tax Rates, Payment Terms, Currencies, Warehouses. The left `nav` provides section-level navigation; the right panel renders the selected section as a data table (for list-based master data) or a settings form (for configuration).

---

### 5.6 Journal Entry Template (wizard variant)

```
page
  └─ wizard (3 steps)
        ├─ Step 1 — Journal Details
        │     DocumentHeaderBlock (no party)
        │     select (journal_type)
        │     textarea (description)
        ├─ Step 2 — Line Items
        │     ProductServiceLineBlock (JournalLineItemConfig)
        │     formula (_balance) + tpl (balance indicator — green/red)
        └─ Step 3 — Review & Post
              DetailCardBlock (header summary)
              TotalsSummaryBlock
              checkbox (confirm_balanced, required, disabledOn: "${_balance !== 0}")
```

The balance indicator on Step 2 is the critical UX element: it shows debit total, credit total, and whether the entry is balanced in real-time as each line is entered. The "Post" button on Step 3 is disabled until the journal is balanced.

---

## Part 6 — Runtime Contract

### 6.1 Fixed Scope Variables — Complete Table

| Variable | Type | Where set | Consumer blocks |
|---|---|---|---|
| `${id}` | UUID string | CRUD row scope / URL param | Row actions, `ActivityFeedBlock` |
| `${record}` | object | `initApi` response | `DetailCardBlock`, form defaults |
| `${totals}` | `[]{ label, amount }` | `initApi` response | `TotalsSummaryBlock` |
| `${tax_lines}` | `[]{ tax_name, rate_pct, taxable_amount, tax_amount }` | `initApi` response | `TaxSummaryBlock` |
| `${breadcrumbs}` | `[]{ label, href }` | `initApi` response | `EntityBreadcrumbBlock` |
| `${can_approve}` | bool | Page `Data` map | `ApprovalWorkflowBlock` |
| `${api_url}` | string | Page `Data` map | `ReportTableBlock` exports |
| `${current_user_id}` | UUID string | Page `Data` map | Ownership expressions |
| `${tenant_currency}` | string | Page `Data` map | Currency display in KPI blocks |
| `${items}` | []object | `crud` API response | Auto-managed by amis `crud` |
| `${total}` | int | `crud` API response | Auto-managed by amis `crud` |
| `${periodName}` | string | `initApi` response | Drill context subtitle `tpl` |
| `${costCentreName}` | string | `initApi` response | Drill context subtitle `tpl` |

### 6.2 List Response — Full Shape

```go
// handler — all list endpoints follow this exactly
c.JSON(fiber.Map{
    "status": 0,
    "msg":    "ok",
    "data": fiber.Map{
        "total":              totalCount,
        "items":              records,       // exactly perPage items
        "hasFiltersApplied":  filtersActive, // true if any filter param was non-empty
        "totalUnfiltered":    totalWithoutFilters,
    },
})
```

### 6.3 Detail Response — Full Shape

```go
// handler — detail pages include scope variables alongside record fields
c.JSON(fiber.Map{
    "status": 0,
    "msg":    "ok",
    "data": fiber.Map{
        // record fields
        "id":           invoice.ID,
        "ref_number":   invoice.RefNumber,
        // ...

        // required scope variables
        "totals":       buildTotals(invoice),
        "tax_lines":    buildTaxLines(invoice),
        "breadcrumbs":  buildBreadcrumbs(invoice),
        "periodName":   "",   // set on report pages; empty on document pages
    },
})
```

### 6.4 Standard Expression Patterns

```go
// Go helper functions used inside block code

func disabledWhen(readOnly bool) string {
    if readOnly { return "true" }
    return ""
}

func hiddenInCreateMode() string { return "${!id}" }
func hiddenInEditMode() string   { return "${!!id}" }

func disabledAfterApproval() string {
    return `${status === 'approved' || status === 'paid'}`
}

func disabledWhenTermsSet() string {
    return `${!!payment_terms_id}`
}
```

### 6.5 `submitOnChange` Filter Pattern

`FilterBarBlock` uses `submitOnChange: false` — the CRUD refreshes on explicit submit (Enter / Search button).

`QuickFilterChipsBlock` uses `submitOnChange: true` — each chip toggle fires the CRUD immediately.

Both drive the same `crud` — the `crud` is configured with the filter form's `name` so amis knows which form drives it. This is handled internally by `DataTableBlock`.

### 6.6 Live Computation Pattern (`formula`)

Line-item subtotals use `formula` type — client-side, no API call, runs on every change:

```json
{
  "type": "formula",
  "name": "subtotal",
  "expr": "${qty * unit_price * (1 - discount_pct / 100)}",
  "readOnly": true
}
```

Document totals (`TotalsSummaryBlock`) are **not** computed client-side. They read `${totals}` from the server response. The server computes totals on every save and includes them in the detail API response.

### 6.7 Destructive Action Friction Scale

| Action | Friction required | Implementation |
|---|---|---|
| Delete a draft record | Single confirm dialog | `confirmText` on `action` |
| Delete an approved record | Must type the record number | Custom confirm dialog with input validation |
| Reverse a posted journal | Typed confirmation + mandatory reason | Dialog with `input-text` + `textarea` + notifies original poster |
| Hard-close an accounting period | Finance Director PIN + 24h delay | Temporal workflow with cancellation window |
| Delete a user | Blocked — deactivate only | Backend enforces; delete endpoint returns 405 |

### 6.8 Empty State Response Contract

All list endpoints must include `hasFiltersApplied` and `totalUnfiltered` in their responses. This enables three distinct empty state messages:

```
hasFiltersApplied: false, total: 0
→ "No [records] yet. Click + to create your first."

hasFiltersApplied: true, total: 0, totalUnfiltered: 247
→ "No results match your filters. 247 [records] exist but your current filters show none."

HTTP 403 from list endpoint
→ "No records visible. You may not have permission to view records in this area."
```

---

## Part 7 — Performance & Caching Reference

### 7.1 Cache TTL Table

| Resource | Cache location | TTL | Invalidation trigger |
|---|---|---|---|
| Nav schema | CDN | 24h | Manual deployment |
| Chart of accounts | Redis | 1h | Account created/modified |
| Cost centres | Redis | 1h | Cost centre created/modified |
| Countries/currencies | Redis | 24h | Rarely change; manual clear |
| Payment terms options | Redis | 1h | Terms created/modified |
| Period list | Redis | 15m | Period status changed |
| P&L (closed periods) | Redis | 24h | Period re-opened (rare) |
| P&L (current period) | No cache | — | Always real-time |
| Dashboard KPIs | Redis | 1h | Posted journal / payment |
| Dashboard exceptions | Redis | 5m | After shift close, payment |
| Nav badge counts | Redis | 5m | Event-driven update |
| Approval email tokens | DB (`approval_steps.used_at`) | 48h | Single-use; recorded on first action |

### 7.2 API Payload Discipline

**List endpoints return only fields displayed in the table.** Each endpoint must have a field-set specification:

```go
// Wrong — returns full record with 50 fields for a 5-column table
SELECT * FROM invoices WHERE tenant_id = $1

// Correct — returns only what the table displays
SELECT id, ref_number, customer_name, total_amount, due_date, status
FROM invoices WHERE tenant_id = $1
```

**Detail endpoints return only the current tab's data.** Sub-record tabs (line items, approvals, history) fire their own `service` API calls on first tab activation — not pre-loaded in the main detail response.

### 7.3 Pre-aggregation Schedule

```
02:00 EAT  — Nightly aggregation job starts
            — Rolls up all posted journal_entry_lines into period_account_balances
            — Rolls up by department into period_dept_balances
            — Refreshes invoice_aging_summary
            — Updates dip_variance_cache from last 24h shift closes

06:00 EAT  — Alert if job has not completed (before business hours start)
```

The current open period is not pre-aggregated (too frequently changing). Reports UNION the pre-aggregated closed periods with a real-time aggregate of current-period lines.

### 7.4 Dashboard API Performance Budget

| Request | Endpoint | Cache TTL | Expected size |
|---|---|---|---|
| Exception panel | `/api/dashboard/exceptions` | 5 min | ~2KB |
| KPI cards | `/api/dashboard/kpis` | 1 hour | ~3KB |
| Sparklines (6-period) | `/api/dashboard/sparklines` | 1 hour | ~8KB |
| Period status | `/api/gl/periods/current-status` | 15 min | ~1KB |
| Nav badge counts | `/api/nav/counts` | 5 min | ~200B |

Total dashboard payload: ~14KB from API. Target TTI: < 1.5s broadband, < 4s on 3G.

---

## Part 8 — Developer Guide

### 8.1 Adding a New Block

1. Create `internal/web/dsl/blocks/<name>.go`
2. Define an exported `*Config` struct. The zero value must be safe — no panics on unset fields.
3. Define `*Block(sess ui.UISessionContext, cfg *Config) ast.Node`
4. Add an entry to Appendix A (block inventory table).
5. Document the block in Part 4 under the appropriate section.
6. Write a unit test (see §8.3).

**Rules — never violate these:**
- No I/O inside a block (no HTTP, no DB, no file reads).
- No IAM service calls inside a block. Use `sess.Can(...)` — it reads from session data, not the IAM service.
- No global state. Blocks are pure functions: same inputs → same output, every time.
- Config struct zero values must be safe. If a field is required, validate it at the top of the function. Return an `ast.AlertNode` with an error message rather than panicking.
- No `className` additions that override amis component internals. `className` is only for AWO-specific layout classes (`erp-nav`, `erp-compact`, `erp-filter-context`).

### 8.2 Block Helper Functions

These are unexported functions in the `blocks` package, available to all blocks:

| Function | Signature | Purpose |
|---|---|---|
| `canPerm` | `(sess, "resource.action") bool` | Checks dot-notation permission |
| `boolExpr` | `(bool) string` | Returns `"true"` or `""` for `disabledOn` |
| `resourceFromURL` | `(url string) string` | Extracts `"module.resource"` from API URL |
| `disabledAfterApproval` | `() string` | Standard approval lock expression |
| `hiddenInCreateMode` | `() string` | Returns `"${!id}"` |
| `hiddenInEditMode` | `() string` | Returns `"${!!id}"` |

Do not export these. Do not call them from screen files.

### 8.3 Unit Test Pattern

Every block must have a unit test that:
1. Constructs the block with default/empty config.
2. Compiles the result via `ast.CompileTree`.
3. Asserts the top-level `type` field matches the expected amis type.
4. Tests at least two config variants (e.g., `ReadOnly: true` vs `ReadOnly: false`).

```go
func TestInternalNotesBlock(t *testing.T) {
    sess := ui.MockSession(t, ui.MockSessionConfig{
        Permissions: []string{"invoice.update"},
    })

    node := blocks.InternalNotesBlock(sess)

    schema, err := ast.CompileTree(node)
    require.NoError(t, err)
    assert.Equal(t, "collapse", schema["type"])
    assert.True(t, schema["collapsed"].(bool))
}

func TestDocumentHeaderBlock_variants(t *testing.T) {
    tests := []struct {
        name          string
        cfg           blocks.DocumentHeaderConfig
        wantReadOnly  bool
        wantCurrency  bool
    }{
        {"defaults",       blocks.DocumentHeaderConfig{},
         false, false},
        {"with currency",  blocks.DocumentHeaderConfig{ShowCurrency: true},
         false, true},
        {"read only",      blocks.DocumentHeaderConfig{ReadOnly: true},
         true, false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            sess := ui.MockSession(t, ui.MockSessionConfig{})
            node := blocks.DocumentHeaderBlock(sess, tt.cfg)
            schema, err := ast.CompileTree(node)
            require.NoError(t, err)
            require.NotEmpty(t, schema)
            // ... field-level assertions ...
        })
    }
}
```

### 8.4 CSS Contribution Rules

Before adding any CSS to `web/static/theme/awo.css`:

1. Can this be solved at the amis schema level (component property, `className` with an existing class)?
2. Can this be solved with an existing CSS variable from the `cxd` theme?
3. Is this a structural surface override (sidebar) or a utility class (density)?

If the answer to all three is no, open a design review. The current file has 4 rules. Every addition must have a documented justification comment above the rule:

```css
/* Justification: density toggle — reduces row height for power users.
   Set by app shell from localStorage:erp:tableDensity.
   No amis property achieves this — table padding is not configurable via schema. */
.erp-compact .cxd-Table-table td {
  padding-top: 8px;
  padding-bottom: 8px;
}
```

---

## Appendix A — Block Inventory

> Master table. Update this when adding a new block. Version = when block was added to the codebase.

| Block | Source file | amis types emitted | Part | Version |
|---|---|---|---|---|
| `PageHeaderBlock` | `page_header.go` | `flex`, `breadcrumb`, `button-toolbar` | 4A | v0.2 |
| `FilterBarBlock` | `filter_bar.go` | `form`, `input-text`, `input-date-range`, `select` | 4B | v0.1 |
| `QuickFilterChipsBlock` | `quick_filter_chips.go` | `form`, `input-tag` | 4B | v0.2 |
| `AdvancedSearchBlock` | `advanced_search.go` | `collapse`, `condition-builder` | 4B | v0.2 |
| `DataTableBlock` | `data_table.go` | `crud` | 4C | v0.1 |
| `StatusBadgeColumn` | `status_badge.go` | `ast.TableColumn` (mapping) | 4C | v0.1 |
| `StatusBadgeBlock` | `status_badge.go` | `mapping` | 4C | v0.1 |
| `EmptyStateBlock` | `empty_state.go` | `container`, `tpl`, `button` | 4C | v0.1 |
| `DetailCardBlock` | `detail_card.go` | `panel`, `property` | 4C | v0.1 |
| `StatCardBlock` | `stat_card.go` | `panel`, `flex`, `sparkline` | 4D | v0.1 |
| `KPIRowBlock` | `kpi_row.go` | `grid`, `panel` × N | 4D | v0.1 |
| `ChartPanelBlock` | `chart_panel.go` | `panel`, `select`, `chart` | 4D | v0.1 |
| `ActivityPanelBlock` | `activity_panel.go` | `panel`, `service`, `list` | 4D | v0.1 |
| `DocumentHeaderBlock` | `document_header.go` | `group`, `input-text`, `input-date`, `select` | 4E | v0.1 |
| `PartyBlock` | `party.go` | `panel`, `picker` | 4E | v0.1 |
| `AddressBlock` | `address.go` | `panel`, `collapse-group`, `input-text`, `select` | 4E | v0.1 |
| `ProductServiceLineBlock` | `line_items.go` | `panel`, `input-table` or `combo`, `formula` | 4E | v0.1 |
| `TotalsSummaryBlock` | `totals.go` | `panel`, `table` | 4E | v0.1 |
| `TaxSummaryBlock` | `tax_summary.go` | `panel`, `table` | 4E | v0.1 |
| `PaymentTermsBlock` | `payment_terms.go` | `group`, `select`, `input-date` | 4E | v0.1 |
| `InternalNotesBlock` | `notes.go` | `collapse`, `input-rich-text` | 4E | v0.1 |
| `AttachmentsBlock` | `attachments.go` | `collapse`, `input-file` | 4E | v0.1 |
| `ApprovalWorkflowBlock` | `approval.go` | `panel`, `static`, `mapping`, `textarea`, `button-group` | 4E | v0.1 |
| `ActivityFeedBlock` | `activity_feed.go` | `panel`, `service`, `log` | 4E | v0.1 |
| `ReportHeaderBlock` | `report_header.go` | `alert`, `tpl` | 4F | v0.1 |
| `ReportFilterBlock` | `report_filter.go` | `form`, `input-month/quarter/year`, `select` | 4F | v0.1 |
| `ReportTableBlock` | `report_table.go` | `table`, `button-toolbar` | 4F | v0.1 |
| `ReportChartBlock` | `report_chart.go` | `chart` | 4F | v0.1 |
| `EntityBreadcrumbBlock` | `entity_breadcrumb.go` | `breadcrumb` | 4G | v0.1 |
| `QuickActionsBlock` | `quick_actions.go` | `button-toolbar`, `action` | 4G | v0.1 |
| `BulkActionsBlock` | `bulk_actions.go` | `[]button` | 4G | v0.1 |

---

## Appendix B — Accessibility Checklist

AWO ERP targets **WCAG 2.1 AA compliance** throughout, inherited primarily from amis's `cxd` theme and React's ARIA implementation. These additional checks are required for AWO-specific patterns.

### Mandatory per-component checks

**`mapping` columns and blocks:**
Every `mapping` HTML value must include an `aria-label` attribute:
```html
<!-- Correct -->
<span class="label label-danger" aria-label="Status: Overdue">
  <i class="fa fa-clock" aria-hidden="true"></i> Overdue
</span>

<!-- Wrong — icon alone has no accessible label -->
<span class="label label-danger">
  <i class="fa fa-clock"></i>
</span>
```

**`tpl` interactive elements:**
Any interactive element rendered inside a `tpl` node (a clickable drill link, a copy button) must have `role="button"` and `tabIndex="0"` if it is not a native `<a>` or `<button>` tag.

**`form` fields:**
Never use `placeholder` as a substitute for `label`. The `label` property on every field is mandatory. Placeholder text disappears on focus and is not read by most screen readers. The `label` in amis is always rendered as a proper `<label>` element with a `for` attribute.

**Modals:**
amis `dialog` traps focus correctly by default. Do not override the default focus management behaviour. When a modal opens, focus must move into it. When it closes, focus must return to the trigger element.

**Status badges on print:**
Status indicators use both colour and text. When printed in black and white, the text label identifies the status — colour is redundant, not the primary carrier.

### Colour contrast checks

| Surface | Foreground | Background | Ratio | WCAG AA (4.5:1 normal text) |
|---|---|---|---|---|
| Body text | `#262626` | `#ffffff` | 16:1 | ✓ |
| Secondary text | `#8c8c8c` | `#ffffff` | 3.5:1 | ✗ — use for captions only, not data |
| Nav active | `#ffffff` | `#001529` | 12.6:1 | ✓ |
| Nav inactive | `rgba(255,255,255,0.65)` | `#001529` | 4.7:1 | ✓ |
| Danger badge | `#ffffff` | `#f5222d` | 4.6:1 | ✓ |
| Warning badge | `#ffffff` | `#fa8c16` | 3.0:1 | ✗ — use dark text on warning badges |
| Input border | `#d9d9d9` | `#ffffff` | 1.6:1 | Non-text — decorative only |

Note: The warning badge (`#fa8c16`) does not meet AA for white text. Warning status badges must use dark (`#262626`) text, not white. This is handled in the amis `cxd` theme defaults — do not override it.

### CI pipeline requirement

Run Lighthouse accessibility audit on every pull request against the staging deployment:

```yaml
# .github/workflows/ci.yml
- name: Lighthouse accessibility check
  run: |
    npx lighthouse $STAGING_URL \
      --only-categories=accessibility \
      --output=json \
      --output-path=./lighthouse-result.json
    node -e "
      const r = require('./lighthouse-result.json');
      const score = r.categories.accessibility.score * 100;
      if (score < 85) { console.error('Accessibility score ' + score + ' below 85'); process.exit(1); }
    "
```

Block deployments with a score below 85/100.

---

## Appendix C — Dark Mode (Deferred)

AWO ERP v1.x ships **light mode only**.

The CSS variable structure in §1.2 and §1.7 is authored for extensibility. When dark mode is added:

1. Add a `web/static/theme/dark.css` file with a `[data-theme="dark"]` selector block overriding the surface and text variables.
2. Toggle via `document.documentElement.setAttribute('data-theme', 'dark')` in the app shell.
3. Store preference in `localStorage` as `erp:theme`.
4. **No block Go code changes are required.** Blocks emit the same amis component types. Only the CSS token layer changes.

The two structural CSS overrides (sidebar background, sidebar text) already use CSS variables — they will automatically pick up dark theme tokens without any additional changes.

---

*AWO ERP — UI Block System Reference v2.0*
*Renderer: Baidu amis v3.x · cxd theme · Factumsoft-inspired design language*
*Maintained in: `internal/web/dsl/blocks/` · `internal/web/ast/` · `web/schemas/pages/`*
