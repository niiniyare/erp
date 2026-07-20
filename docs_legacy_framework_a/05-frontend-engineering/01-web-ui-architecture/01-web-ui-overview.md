> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Web UI Architecture Overview
portal: 5 — Frontend Engineering
section: 05-frontend-engineering
audience: [frontend-engineer, full-stack-engineer]
related:
  - "[AMIS Schema Guide](../02-amis-schemas/01-amis-schema-overview.md)"
  - "[Dark Mode Implementation](../03-theming/01-dark-mode.md)"
  - "[API Design](../../04-backend-engineering/00-module-development-guide/14-api-design/01-api-design-overview.md)"
---

# Web UI Architecture Overview

AwoERP's web UI is a **schema-driven single-page application** built on the [AMIS SDK](https://aisuda.bce.baidu.com/amis/en-US/docs). The UI renders from JSON schemas — no React component tree, no build pipeline. Every screen is a JSON file.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Rendering engine | AMIS SDK 6.x (`sdk.js`, `sdk.css`) |
| Charts | ECharts (bundled with AMIS) |
| Shell | Vanilla JS + HTML |
| No build tool | No webpack, vite, or npm compile step |

## File Structure

```
web/
├── pages/
│   └── index.html          # App shell: sidebar + AMIS mount point
├── schemas/
│   └── pages/
│       ├── dashboard.json
│       ├── contracts/
│       │   ├── list.json
│       │   ├── detail.json
│       │   └── create.json
│       ├── tenants/
│       └── entities/
├── sdk/
│   ├── sdk.js              # AMIS core bundle
│   ├── sdk.css             # AMIS default theme
│   └── charts.js           # ECharts bundle
└── styles/
    └── theme.css           # CSS variable overrides (dark mode, brand)
```

## How It Works

1. `index.html` loads `sdk.js` and mounts AMIS on `<div id="root">`
2. Sidebar navigation fetches the target schema JSON on route change
3. AMIS deserializes the JSON schema and renders the component tree
4. Components make API calls directly to `/api/v1/...` using AMIS `api` properties

```html
<!-- web/pages/index.html (simplified) -->
<div id="root"></div>
<script>
amis.embed('#root', {
  type: 'app',
  brandName: 'AwoERP',
  logo: '/assets/logo.svg',
  pages: [
    {
      label: 'Contracts',
      icon: 'fa fa-file-contract',
      url: '/contracts',
      schema: { type: 'service', schemaApi: '/schemas/pages/contracts/list.json' }
    }
  ]
});
</script>
```

## Schema-Driven Pattern

Every screen is a JSON schema loaded at runtime — no redeployment needed for UI changes:

```json
// web/schemas/pages/contracts/list.json
{
  "type": "page",
  "title": "Contracts",
  "body": {
    "type": "crud",
    "api": "/api/v1/contracts",
    "columns": [
      { "name": "contract_number", "label": "Number" },
      { "name": "title", "label": "Title" },
      { "name": "status", "label": "Status", "type": "mapping",
        "map": { "draft": "secondary", "active": "success" } }
    ],
    "headerToolbar": ["create", "export-csv"]
  }
}
```

## API Integration

AMIS components call the backend via the `api` property:

```json
{
  "type": "form",
  "api": {
    "method": "post",
    "url": "/api/v1/contracts",
    "responseData": { "id": "${id}" }
  },
  "body": [
    { "type": "input-text", "name": "title", "label": "Title", "required": true },
    { "type": "select", "name": "contract_type", "label": "Type",
      "options": [
        { "label": "Service", "value": "service" },
        { "label": "Supply", "value": "supply" }
      ]
    }
  ]
}
```

AMIS handles CSRF-safe requests, loading states, and error display automatically.

## Authentication

The app shell stores the session token in `localStorage` after login and injects it into every AMIS API request:

```javascript
amis.embed('#root', schema, {}, {
  // Request interceptor — inject auth header
  fetcher: ({ url, method, data, headers }) => {
    const token = localStorage.getItem('session_token');
    return window.fetch(url, {
      method,
      headers: { ...headers, 'Authorization': `Bearer ${token}` },
      body: data ? JSON.stringify(data) : undefined,
    }).then(r => r.json());
  }
});
```

## Adding a New Screen

1. Create `web/schemas/pages/<module>/<screen>.json`
2. Add a page entry to the `pages` array in `index.html`
3. No build step — refresh browser

See [AMIS Schema Guide](../02-amis-schemas/01-amis-schema-overview.md) for component reference.

## Dark Mode

AMIS SDK has no built-in dark CSS. Dark mode works via CSS custom property overrides on `html.dark`:

```css
/* web/styles/theme.css */
html.dark {
  --colors-neutral-fill-11: #1a1a1a;
  --colors-neutral-text-2: #e8e8e8;
  --background: #121212;
  --body-bg: #1e1e1e;
  --Page-main-bg: #1e1e1e;
  --Panel-bg-color: #2a2a2a;
  --Table-bg: #2a2a2a;
}
```

Toggle: `document.documentElement.classList.toggle('dark')`.

Do not override individual `.cxd-*` class backgrounds — CSS variable cascade handles all components.

See [Dark Mode Implementation](../03-theming/01-dark-mode.md) for full variable list.

## Constraints

- All `tenant_id` values come from the session — never hardcode in schemas
- Never import `@internal/*` Go packages from JS — all data via REST API
- Schema files are served as static files — no server-side rendering
- AMIS SDK version is pinned in `web/sdk/` — do not auto-upgrade without testing
