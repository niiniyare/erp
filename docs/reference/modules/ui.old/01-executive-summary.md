[<-- Back to Index](README.md)

## Executive Summary

### What the UI Layer Is

Awo ERP's frontend is not a traditional React or Vue application. It is a **JSON-driven rendering engine** (AMIS) mounted inside a **custom HTML shell** (`web/index.html`).

```markdown
WHAT YOU DO NOT WRITE:
❌ React component files
❌ CSS for page layouts
❌ JavaScript event handlers for forms
❌ Pagination logic
❌ Filter state management
❌ Form validation wiring

WHAT YOU DO WRITE:
✅ Go structs that produce AMIS-compatible JSON (schema package)
✅ Go REST handlers that return data in the correct envelope
✅ JSON schema files in web/schemas/pages/
✅ CSS variable overrides for dark mode and tenant branding
```

### System Architecture Context

```markdown
BROWSER REQUESTS / (index.html)
        │
        ▼
  Custom Shell (web/index.html)
  ├── Sidebar (HTML/CSS/JS — menu, collapse, mobile)
  ├── Topbar (breadcrumb, theme toggle)
  ├── Theme System (light / dark / system — localStorage)
  └── #content div
            │
            ▼
      AMIS Runtime (sdk.js embedded in #content)
      ├── Reads JSON schema from web/schemas/pages/{route}.json
      ├── Renders page components (crud, form, wizard, chart…)
      └── Calls /api/v1/* for data (Go backend)
                │
                ▼
          Go Fiber Backend
          ├── /api/v1/*  — business data (JSON envelope)
          └── /schema/*  — dynamic schemas (future: Go-driven)
```

### Two Routing Systems

| Layer | Routing | Driven By |
|---|---|---|
| Shell navigation | Hash-based (`#dashboard`, `#users`) | `menuConfig` JS array in `index.html` |
| AMIS page content | Loads `web/schemas/pages/{route}.json` | Static JSON files |
| Future (planned) | `/schema/{module}/{page}` from Go | Feature flags + permissions |

### Key Decisions Already Made

**Custom shell, not AMIS `app` component.** AMIS's built-in `app` component can render the sidebar, but the custom HTML shell gives us full control over dark mode, mobile responsiveness, collapse behaviour, and theming without being constrained by AMIS's CSS structure.

**Static JSON schemas for now.** Schemas live as JSON files in `web/schemas/pages/`. When per-tenant or per-flag customisation is needed, the shell's schema loader switches to calling `/schema/*` Go endpoints. No rewrite required — just change the URL.

**Two envelope formats bridged by the fetcher.** The Go backend returns `{success, data, meta}`. AMIS expects `{status, data}`. The custom `fetcher` in `index.html` translates between them transparently.

---
