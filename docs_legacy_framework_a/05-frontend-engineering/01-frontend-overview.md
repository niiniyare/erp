> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Frontend Overview
portal: 5 — Frontend Engineering
section: 05-frontend-engineering
audience: [frontend-engineer, fullstack-engineer]
related:
  - "[AMIS Schema Guide](02-amis-schema-guide.md)"
  - "[Dark Theme](03-dark-theme.md)"
  - "[API Design](../04-backend-engineering/00-module-development-guide/14-api-design/01-api-design-overview.md)"
---

# Frontend Overview

The AwoERP web UI is built on [AMIS](https://aisuda.bce.baidu.com/amis/en-US/docs/index) — a JSON-driven UI framework from Baidu. Pages are defined as JSON schemas, eliminating most hand-written frontend code for standard CRUD interfaces.

## Architecture

```
web/
├── pages/
│   └── index.html          # Shell page: sidebar + AMIS renderer
├── schemas/
│   └── pages/
│       ├── contracts.json  # Contracts page schema
│       ├── finance.json
│       └── ...
└── sdk/
    ├── sdk.js              # AMIS SDK bundle
    ├── sdk.css             # AMIS styles
    └── charts/             # ECharts bundle (for chart components)
```

## How it Works

1. `index.html` loads the AMIS SDK and mounts an AMIS `app` component
2. The `app` schema defines the sidebar navigation and page routes
3. Each route loads a page schema JSON from `schemas/pages/`
4. AMIS renders the schema — forms, tables, dialogs — without hand-written components
5. AMIS makes API calls to the backend REST endpoints

## AMIS SDK Dark Theme

AMIS has no built-in dark CSS. The `theme("dark")` option uses `classPrefix: "dark-"` but has zero CSS rules for that prefix — it breaks styling.

**Fix**: override AMIS CSS custom property tokens at the root. Scale `--colors-neutral-fill/text/line` (1=darkest, 11=lightest). Invert on `html.dark` → all components inherit.

Key CSS vars to override:
```css
html.dark {
    --colors-neutral-fill-11: #1a1a1a;
    --colors-neutral-text-2: #e5e5e5;
    --colors-neutral-line-8: #333;
    --background: #121212;
    --body-bg: #121212;
    --Page-main-bg: #1a1a1a;
    --Panel-bg-color: #1e1e1e;
    --Table-bg: #1e1e1e;
}
```

Never override individual `.cxd-*` backgrounds with `!important` — it becomes whack-a-mole.

## When NOT to Use AMIS

Use AMIS schemas for standard CRUD, lists, forms, and dashboards. Build custom React/Vue components when:

- The interaction is highly custom (drag-and-drop, canvas)
- The data visualization is complex beyond ECharts
- Real-time updates via WebSocket are required

Custom components are registered as AMIS custom components and embedded in schemas.

## Schema Conventions

- Use `type: "page"` as root for standalone pages
- Use `type: "crud"` for list + create + edit screens
- API endpoints always include `Authorization` header via AMIS `api.headers`
- Pagination: pass `page` and `pageSize` query params; AMIS handles pagination state
- Error responses: AMIS reads `{"status": 0, "msg": "...", "data": {}}` format

Map backend error responses to AMIS format in the API adapter or use AMIS `responseData` transform.
