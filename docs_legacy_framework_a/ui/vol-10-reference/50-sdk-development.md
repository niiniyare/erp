> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: "50 – SDK Development"
volume: "vol-10-reference"
chapter: 50
section: "Reference"
status: "partial — web SDK in place; typed client SDKs planned"
---

# Chapter 50 – SDK Development

## Table of Contents
- [50.1 What an SDK Would Cover](#501-what-an-sdk-would-cover)
- [50.2 Current: AMIS Web SDK](#502-current-amis-web-sdk)
- [50.3 Planned: Typed Client SDKs for Flutter and React Native](#503-planned-typed-client-sdks-for-flutter-and-react-native)

---

## 50.1 What an SDK Would Cover

An "AwoERP UI SDK" in the full sense would provide:

| Capability | Description |
|---|---|
| Schema contract types | Machine-generated TypeScript / Dart types matching every `ast.*` struct, so client code is type-safe against schema changes. |
| Schema fetch client | A typed HTTP client that fetches a schema for a given route and deserialises it into the schema contract types. |
| Renderer integration | A renderer component (web widget, Flutter widget) that accepts a schema object and renders the AMIS-compatible UI. |
| Auth token management | Helpers for attaching session tokens to schema requests, handling 401 refresh, and secure token storage. |
| Cache invalidation hooks | Client-side cache invalidation when the server emits a schema version change signal. |
| Error boundary | A standardised error display component for schema fetch failures or pipeline errors. |

Today, only a subset of these exist — the web shell covers them for the browser
client; nothing exists yet for mobile clients.

---

## 50.2 Current: AMIS Web SDK

The web shell in `web/` bundles the AMIS SDK and handles schema rendering for
browser clients.

```
web/
  pages/
    index.html         — Shell HTML: sidebar nav + AMIS embed target
  sdk/
    sdk.js             — AMIS SDK (self-contained bundle)
    sdk.css            — AMIS SDK base styles
    charts/            — ECharts bundle (for chart components)
  styles/
    css/               — AwoERP theme: CSS custom property overrides
                         for both light and dark mode
  schemas/
    pages/             — Static JSON schemas (development only)
                         Production schemas fetched from SchemaHandler
```

### Schema fetch in the web shell

`index.html` fetches the schema for the current route on navigation and passes
it to `amis.embed`:

```javascript
async function loadSchema(route) {
  const resp = await fetch(`/ui/schema?route=${encodeURIComponent(route)}`, {
    headers: { Authorization: `Bearer ${getSessionToken()}` }
  });
  if (!resp.ok) {
    renderError(resp.status);
    return;
  }
  const schema = await resp.json();
  amis.embed('#main-content', schema, {}, {
    locale: window.__ERP_SESSION__.locale,
  });
}
```

This pattern requires no additional client SDK — the browser fetches JSON,
passes it to AMIS, and the AMIS SDK handles all rendering.

### Theme customisation

Dark and light mode are controlled via CSS custom property overrides in
`web/styles/css/`.  AMIS SDK components inherit these tokens without any SDK
modification:

```css
/* web/styles/css/theme-dark.css */
html.dark {
  --colors-neutral-fill-11:   #1a1a1a;
  --colors-neutral-text-2:    #e0e0e0;
  --colors-neutral-line-8:    #333333;
  --body-bg:                  #121212;
  --Panel-bg-color:           #1e1e1e;
  --Table-bg:                 #1e1e1e;
}
```

---

## 50.3 Planned: Typed Client SDKs for Flutter and React Native

> **PLANNED** — not yet started.

Two mobile client SDKs are on the roadmap:

### Flutter SDK

| Component | Description |
|---|---|
| `erp_ui_schema` Dart package | Dart classes generated from `ast.*` Go structs via `protoc` or JSON schema code generation. |
| `SchemaFetchClient` | Dart HTTP client for `/ui/schema` endpoint with retry and token refresh. |
| `AMISRenderer` Flutter widget | Renders a schema tree as Flutter widgets; maps `ast.CRUDNode` → `DataTable`, `ast.FormNode` → `Form`, etc. |
| `SchemaCache` | In-memory Dart cache keyed by route + schema version header. |

```dart
// Planned usage
final client = SchemaFetchClient(baseUrl: 'https://api.awoerp.com', token: token);
final schema = await client.fetch('/finance/dashboard');

// In widget tree
AMISRenderer(schema: schema)
```

### React Native SDK

| Component | Description |
|---|---|
| `@awoerp/ui-schema` npm package | TypeScript types generated from `ast.*` structs. |
| `useSchema` hook | React hook that fetches and caches a schema for a given route. |
| `SchemaRenderer` component | React Native component that renders the schema tree; maps node types to RN primitives. |

```tsx
// Planned usage
function FinanceDashboardScreen() {
  const { schema, loading, error } = useSchema('/finance/dashboard');
  if (loading) return <LoadingSpinner />;
  return <SchemaRenderer schema={schema} />;
}
```

### Schema version negotiation (prerequisite)

Both mobile SDKs require the schema versioning system described in §43.5 to be
implemented first.  Mobile clients must be able to declare the schema version
they support so the server can serve compatible output across app store release
cycles.

Until the Flutter/RN SDKs are available, mobile access to AwoERP UI is via the
web shell in a WebView.
