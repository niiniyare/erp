[<-- Back to Index](README.md)

## Shell Implementation (`web/index.html`)

The entire shell is a single self-contained HTML file. No build step. No npm. No framework.

### HTML Structure

```html
<div id="app">
  <aside id="sidebar">
    <div class="sidebar-brand">        <!-- Logo + collapse toggle -->
    <div class="sidebar-search">       <!-- Search box (cosmetic for now) -->
    <nav class="sidebar-menu">         <!-- Rendered by renderMenu() JS -->
  </aside>

  <div id="sidebar-backdrop">          <!-- Mobile overlay -->

  <div id="main">
    <div id="topbar">                  <!-- Breadcrumb + theme toggle -->
    <div id="content">                 <!-- AMIS renders here -->
    <div id="footer">
  </div>
</div>
```

### CSS Variables (Light/Dark)

All shell styles use CSS custom properties. Dark mode is a second set of values applied under `html.dark`:

```css
:root {
  /* Layout */
  --sidebar-width: 260px;
  --sidebar-collapsed: 60px;
  --topbar-height: 50px;

  /* Colours — light defaults */
  --sidebar-bg: #fff;
  --sidebar-border: #e8eaed;
  --sidebar-active: #e8f4fd;
  --sidebar-active-border: #2196f3;
  --sidebar-hover: #f5f7fa;
  --text-primary: #1a1a2e;
  --text-secondary: #6b7280;
  --bg-main: #f5f6f8;
  --bg-surface: #fff;
}

html.dark {
  --sidebar-bg: #1e1e2d;
  --sidebar-border: #2e2e3e;
  --sidebar-active: #1a2744;
  --sidebar-active-border: #4dabf7;
  --text-primary: #e4e4e7;
  --bg-main: #151521;
  --bg-surface: #1e1e2d;
  /* ... full list in index.html */
}
```

### Theme System

Three states cycle: `system → light → dark → system`. State is stored in `localStorage`.

```javascript
var themeOrder  = ['system', 'light', 'dark'];
var osDarkQuery = window.matchMedia('(prefers-color-scheme: dark)');

function isDarkEffective(theme) {
  if (theme === 'dark')  return true;
  if (theme === 'light') return false;
  return osDarkQuery.matches;  // system: follow OS
}

function applyTheme(theme) {
  var isDark = isDarkEffective(theme);
  document.documentElement.classList.toggle('dark', isDark);
  document.body.classList.toggle('dark', isDark);
  // Update icon + label in topbar
  localStorage.setItem('awo-theme', theme);
}

// Respond to OS theme changes when in "system" mode
osDarkQuery.addEventListener('change', function () {
  if (getStoredTheme() === 'system') applyTheme('system');
});
```

### Menu Config

Navigation is defined as a JS array. Add items here to add them to the sidebar:

```javascript
var menuConfig = [
  {
    type: 'item',
    id: 'dashboard',
    label: 'Dashboard',
    icon: 'fa-chart-line',
    hash: '#dashboard'
  },
  {
    type: 'group',
    id: 'finance',
    label: 'Finance',
    icon: 'fa-calculator',
    expanded: false,
    items: [
      { id: 'invoices',   label: 'Invoices',   hash: '#invoices'   },
      { id: 'payments',   label: 'Payments',   hash: '#payments'   }
    ]
  }
];
```

Rules:
- `type: 'item'` → top-level link (no children)
- `type: 'group'` → collapsible section with children
- Only one group can be expanded at a time (accordion behaviour)
- `id` must match the JSON schema filename: `id: 'invoices'` → loads `schemas/pages/invoices.json`

### Hash Routing

All navigation is hash-based. No server-side routing needed for page switches.

```javascript
async function navigate() {
  var route = window.location.hash.slice(1) || 'dashboard';
  updateNav(route);    // set active item + breadcrumb

  // Unmount previous AMIS instance
  if (currentInstance) currentInstance.unmount();

  // Load schema file
  var schema = await loader.load('pages/' + route + '.json');

  // Mount AMIS in #content
  currentInstance = amis.embed(contentEl, schema, {}, amisEnv);
}

window.addEventListener('hashchange', navigate);
```

### Adding a New Page — Step by Step

```markdown
1. Add entry to menuConfig in index.html:
   { id: 'purchase-orders', label: 'Purchase Orders', hash: '#purchase-orders' }

2. Create web/schemas/pages/purchase-orders.json:
   {
     "type": "page",
     "title": "Purchase Orders",
     "body": {
       "type": "crud",
       "api": "get:/api/v1/purchase-orders",
       "columns": [...]
     }
   }

3. Ensure Go backend handles GET /api/v1/purchase-orders
   and returns { success: true, data: [...], meta: { pagination: {...} } }

Done. No build step needed.
```

### Simplification Notes

**Current issues to address as the project grows:**

```markdown
ISSUE 1: Menu is hardcoded in JS
Current:  menuConfig array in index.html
Fix when: multiple tenants need different nav
Solution: Load menu from GET /api/v1/me/nav (Go returns flag-filtered items)

ISSUE 2: Two envelope formats
Current:  fetcher bridges {success,data,meta} ↔ {status,data}
Fix when: any inconsistency causes bugs
Solution: Standardise Go handlers to return AMIS envelope directly
          AmisResponse{Status: 0, Data: ...} everywhere

ISSUE 3: Schemas are static JSON files
Current:  web/schemas/pages/*.json
Fix when: schemas need per-tenant/per-flag variation
Solution: Switch SchemaLoader base URL from '/schemas/' to '/schema/'
          Go handlers return dynamic schemas

ISSUE 4: Brand colour hardcoded (#2196f3 in CSS)
Current:  .brand-icon { background: #2196f3; }
Fix when: tenant branding is implemented
Solution: --colors-brand-5 CSS variable + tenant config from /api/v1/me/tenant
```

---
