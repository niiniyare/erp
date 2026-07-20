> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

# Chapter 18 — Web Rendering Architecture

| | |
|---|---|
| **Volume** | 04 — Rendering |
| **Chapter** | 18 |
| **Status** | Implemented |
| **Source** | `web/pages/index.html`, `web/sdk/`, `web/schemas/pages/` |

---

## Table of Contents

1. [18.1 AMIS SDK Overview](#181-amis-sdk-overview)
2. [18.2 Browser Shell](#182-browser-shell)
3. [18.3 Schema Fetching](#183-schema-fetching)
4. [18.4 AMIS Data Chain](#184-amis-data-chain)
5. [18.5 401 Handling](#185-401-handling)
6. [18.6 Dark Mode](#186-dark-mode)
7. [18.7 Static Schemas](#187-static-schemas)
8. [18.8 Current Limitations](#188-current-limitations)

---

## 18.1 AMIS SDK Overview

The web UI is rendered by the **AMIS SDK** — an open-source low-code rendering
engine from Baidu. AwoERP ships a vendored copy in `web/sdk/`:

```
web/sdk/
├── sdk.js       # AMIS runtime (React-based, self-contained)
├── sdk.css      # AMIS design-system styles
├── charts.js    # ECharts integration (loaded alongside sdk.js)
├── rest.js      # REST preset helpers
└── iconfont.css # Icon font for AMIS components
```

The SDK exposes a single global function `amisRequire('amis/embed')` that
accepts a JSON schema and mounts a fully interactive UI into a DOM element.
No build step is required — AMIS schemas are JSON, not JSX.

### Why AMIS?

| Requirement | How AMIS satisfies it |
|---|---|
| Go backend emits UI | JSON schema, no frontend build pipeline |
| Data fetching | `initApi` on any node triggers a fetch |
| Form submission | `api` field on FormNode |
| Pagination, sort, filter | CRUDNode handles automatically |
| Charts | ECharts bridge via `"type": "chart"` |
| Dialogs, drawers | `ActionNode` with `actionType: "dialog"/"drawer"` |

---

## 18.2 Browser Shell

`web/pages/index.html` is the single-page application shell. It provides:

- **Sidebar** — collapsible nav with groups, icons, and a search filter.
- **Topbar** — breadcrumb trail, theme toggle button.
- **Content area** (`#content`) — AMIS mounts here.
- **Footer** — version / settings link.

### Theme modes

The shell supports three modes toggled by the `cycleTheme()` function:

| Mode | Behaviour |
|---|---|
| `system` | Reads `prefers-color-scheme` from the OS |
| `light` | Forces light theme |
| `dark` | Forces dark theme |

The chosen mode is persisted to `localStorage` under key `awo-theme`. On each
page load the shell reads this value and calls `applyTheme()` before AMIS mounts
— preventing a flash of unstyled content.

### Hash-based routing

Navigation is driven by `window.location.hash`. The shell listens on
`hashchange` and calls `navigate()` on every change:

```javascript
window.addEventListener('hashchange', navigate);
```

The `navigate()` function:

1. Reads `window.location.hash` (e.g. `#finance/dashboard`).
2. Strips the leading `#` to get the route string (`finance/dashboard`).
3. Unmounts the current AMIS instance if one exists.
4. Shows a loading spinner in `#content`.
5. Fetches `GET /schema/finance/dashboard`.
6. Mounts the returned schema into `#content` via `amis.embed()`.

### menuConfig

The sidebar menu is defined as a plain JavaScript array `menuConfig` at the top
of the script block in `index.html`:

```javascript
var menuConfig = [
  { type: 'item', id: 'dashboard',   label: 'Dashboard', icon: 'fa-chart-line', hash: '#dashboard' },
  {
    type: 'group', id: 'finance', label: 'Finance', icon: 'fa-calculator', expanded: false,
    items: [
      { id: 'finance/dashboard',     label: 'Overview',         hash: '#finance/dashboard' },
      { id: 'finance/accounts',      label: 'Chart of Accounts', hash: '#finance/accounts' },
      { id: 'finance/transactions',  label: 'Transactions',     hash: '#finance/transactions' },
      // ...
    ]
  },
  { type: 'group', id: 'admin', label: 'Administration', icon: 'fa-cogs', expanded: false,
    items: [
      { id: 'users',         label: 'Users',         hash: '#users' },
      { id: 'organizations', label: 'Organizations', hash: '#organizations' }
    ]
  },
  { type: 'item', id: 'settings', label: 'Settings', icon: 'fa-sliders-h', hash: '#settings' }
];
```

This array is **static JavaScript** — it is not driven by the Go backend.
See section 18.8 for limitations and Ch 19 for the planned Go-driven navigation.

### AMIS environment (`amisEnv`)

The shell configures an `amisEnv` object passed to every `amis.embed()` call:

| Hook | Purpose |
|---|---|
| `fetcher` | Custom HTTP layer: converts AMIS pagination params, attaches CSRF token, normalises backend envelope to AMIS format |
| `jumpTo` | Intercepts AMIS link navigation and rewrites it to hash routing |
| `notify` | Logs alerts to the console (can be extended to toast notifications) |
| `copy` | Uses `navigator.clipboard` for copy-to-clipboard actions |
| `locale` | `'en-US'` — passed to AMIS number/date formatting |

#### Pagination translation

The backend uses `offset`/`limit` pagination. AMIS CRUDNode uses `page`/`perPage`.
The `fetcher` hook translates on the fly:

```
AMIS: ?page=2&perPage=20
→ backend: ?offset=20&limit=20
```

#### Backend envelope normalisation

The backend uses `{ success, data, meta }`. AMIS expects `{ status, data }`.
The `fetcher` hook normalises:

```javascript
// List response
{ success: true, data: [...], meta: { pagination: { total_records: 84 } } }
→ { status: 0, data: { items: [...], count: 84 } }

// Single item / action
{ success: true, data: { id: "..." } }
→ { status: 0, data: { id: "..." } }

// Error
{ success: false, message: "Not found" }
→ { status: 1, msg: "Not found" }
```

---

## 18.3 Schema Fetching

When `navigate()` runs it makes one HTTP request:

```
GET /schema/<route>
Accept: application/json
```

The server responds with the AMIS envelope:

```json
{
  "status": 0,
  "data": { /* AMIS schema */ }
}
```

The shell unwraps `data` and passes it to `amis.embed()`:

```javascript
var envelope = await res.json();
var schema = (envelope.status === 0 && envelope.data) ? envelope.data : envelope;
currentInstance = amis.embed(contentEl, schema, { locale: 'en-US' }, amisEnv);
```

If the route is not found, `fetch()` returns a non-OK status and the shell
renders an error page in `#content` with a "Back to Dashboard" link.

Each navigation unmounts the previous AMIS instance before mounting the new one:

```javascript
if (currentInstance && currentInstance.unmount) {
  currentInstance.unmount();
  currentInstance = null;
}
```

This prevents memory leaks and event listener accumulation.

---

## 18.4 AMIS Data Chain

AMIS uses a hierarchical **data chain** for state management. There is no
Redux, Zustand, or any global store. Each component inherits its parent's
data scope.

### How it works

1. **Page scope**: the root `"type": "page"` node fetches data from `initApi`
   and puts the response in the page scope.
2. **Child access**: any child node can reference page-scope variables using
   `${fieldName}` expressions in any string field.
3. **Service scope**: a `"type": "service"` child node fetches its own data and
   merges it into a sub-scope visible to its own children.
4. **Form scope**: a form node tracks its own field values; children reference
   them via `${fieldName}` relative to the form.
5. **CRUDNode row scope**: row actions and row cells operate in the scope of the
   current row record.

### Example: permission variables

The pipeline injects `can_*` booleans and tenant context into the page data
before the schema is served (see Ch 24, section 24.7). Once AMIS mounts the
page:

```json
{
  "type": "page",
  "initApi": "/api/v1/finance/invoices/abc-123",
  "data": {
    "can_approve": true,
    "tenant_currency": "USD"
  },
  "body": [
    {
      "type": "button",
      "label": "Approve",
      "actionType": "ajax",
      "api": "POST:/api/v1/finance/invoices/${id}/approve",
      "disabledOn": "${!can_approve}"
    }
  ]
}
```

The Approve button is disabled when `can_approve` is false — evaluated by AMIS
in the browser using the data chain, with no additional HTTP call.

### Data chain hierarchy

```
page scope   (initApi result + data{} injections)
  └─ service scope  (service.api result)
       └─ form scope (field values)
            └─ crud row scope  (current row record)
```

Higher-level scopes are always visible to lower-level scopes. A form field can
reference a page-scope variable. The reverse is not true.

---

## 18.5 401 Handling

The backend returns authentication errors in one of two ways:

### HTTP 401 before the pipeline

If the auth middleware rejects the request before the schema pipeline runs, the
shell `navigate()` function detects the HTTP status code directly:

```javascript
if (res.status === 401 || res.status === 403) {
  window.location.href = '/ui/login?next=' + encodeURIComponent(
    window.location.pathname + window.location.hash
  );
  return;
}
```

### AMIS envelope 401

If the pipeline itself returns an auth error embedded in a valid AMIS envelope,
the shell checks the envelope body:

```javascript
if (envelope.error === true || envelope.status === 401) {
  window.location.href = '/ui/login?next=' + encodeURIComponent(
    window.location.pathname + window.location.hash
  );
  return;
}
```

### API call 401 (within AMIS)

For API calls made by AMIS components (not schema fetching), the `fetcher` hook
intercepts the 401:

```javascript
if (res.status === 401) {
  window.location.href = '/ui/login?next=' + encodeURIComponent(
    window.location.pathname + window.location.hash
  );
  return Promise.reject(new Error('Session expired'));
}
```

In all three cases the user is redirected to `/ui/login` with the current URL
encoded as `?next=` so they are returned to the same page after login.

---

## 18.6 Dark Mode

AMIS SDK ships zero dark-mode CSS. The `theme("dark")` variant uses a
`classPrefix: "dark-"` that resolves to no existing CSS rules — it would break
all styling.

The solution is to **override AMIS design tokens** at the CSS custom property
level rather than touching individual `.cxd-*` component classes.

### Token strategy

AMIS uses three neutral scales:

| Scale | Purpose | Direction |
|---|---|---|
| `--colors-neutral-fill-N` | Backgrounds | 1 = darkest, 11 = lightest |
| `--colors-neutral-text-N` | Text colours | same |
| `--colors-neutral-line-N` | Borders | same |

In dark mode, the scales are **inverted**: fill-11 (white in light mode) becomes
the darkest surface colour.

### Dark mode activation

The `applyTheme()` function adds/removes the `dark` class on `<html>` and
`<body>`:

```javascript
if (isDark) {
  document.documentElement.classList.add('dark');
  document.body.classList.add('dark', 'amis-theme-dark');
} else {
  document.documentElement.classList.remove('dark');
  document.body.classList.remove('dark', 'amis-theme-dark');
}
```

### CSS override block

`web/pages/index.html` contains an `html.dark { ... }` block that overrides
all AMIS design tokens. Key overrides:

```css
html.dark {
  /* Surface backgrounds */
  --colors-neutral-fill-11: #1e1e2d;   /* was #ffffff */
  --colors-neutral-fill-10: #262637;   /* was #f7f8fa */

  /* Text */
  --colors-neutral-text-2:  #e4e4e7;   /* primary text */

  /* Borders */
  --colors-neutral-line-8:  #363648;   /* default border */

  /* AMIS component aliases */
  --Page-main-bg:          #151521;
  --Panel-bg-color:        #1e1e2d;
  --Table-bg:              #1e1e2d;
  --Table-thead-bg:        #262637;
  --Card-bg:               #1e1e2d;
  --Modal-bg:              #1e1e2d;
  --Drawer-bg:             #1e1e2d;
}
```

### Portal reinforcement

AMIS portals (dropdowns, modals, drawers, tooltips) are appended directly to
`<body>` — outside the AMIS root element. CSS variables cascade from `html.dark`
but explicit fallback rules cover edge cases:

```css
html.dark .cxd-Modal-content  { background: var(--Modal-bg);  color: var(--text-color); }
html.dark .cxd-Drawer-content { background: var(--Drawer-bg); color: var(--text-color); }
html.dark .cxd-Select-menu    { background: var(--Select-menu-bg); }
html.dark .cxd-PopOver        { background: var(--PopOver-bg); }
```

**Rule**: never override individual `.cxd-*` background colours with
`!important`. Always use the token layer. Targeted `!important` overrides
create a whack-a-mole problem as new AMIS components are added.

---

## 18.7 Static Schemas

During the migration from legacy screens to the DSL pipeline, some pages are
served from static JSON files on disk rather than from the Go schema pipeline.
These live in:

```
web/schemas/pages/
├── dashboard.json
├── users.json
└── ...
```

The `SchemaHandler` serves these files when a route is not registered in the
Go `registry` (or when the file is explicitly configured as a static fallback).
The response format is the same AMIS envelope: `{ "status": 0, "data": { ... } }`.

Static schemas are a migration tool. The goal is for every screen to be
generated by Go AST nodes through the pipeline (see Vol 03). Once a screen is
migrated, its static JSON file is removed.

---

## 18.8 Current Limitations

| Limitation | Impact | Tracking |
|---|---|---|
| `menuConfig` is hardcoded JS | Adding a new screen requires editing `index.html` directly | OPEN — Ch 19 describes planned Go-driven nav |
| No param routing | Edit screens (`/finance/invoices/:id`) cannot be deep-linked | OPEN 2 from review.md |
| Theme applies globally | Cannot scope dark mode to a sub-panel | Accepted; CSS tokens work at `:root` |
| AMIS data chain is page-scoped | Cross-page state (e.g. a shopping cart) has no persistence | Accepted for ERP use case |
| No offline support | All data fetched live from server | Planned — see Ch 22 |
| No real-time push | Stale data until user navigates away and back | Planned — see Ch 21 |
| Static schemas can diverge | If the backend changes its API contract, static JSON can go stale | Mitigated by migration; accepted during transition |
