---
title: AMIS Schema Overview
portal: 5 — Frontend Engineering
section: 05-frontend-engineering
audience: [frontend-engineer, full-stack-engineer]
related:
  - "[Web UI Architecture](../01-web-ui-architecture/01-web-ui-overview.md)"
  - "[CRUD Patterns](02-crud-patterns.md)"
  - "[Form Patterns](03-form-patterns.md)"
  - "[Dark Mode](../03-theming/01-dark-mode.md)"
---

# AMIS Schema Overview

Every AwoERP screen is a JSON schema rendered by the AMIS SDK. This document covers the schema structure, component vocabulary, and conventions to follow when building new screens.

## Schema Structure

Every schema has a `type` field that tells AMIS which component to render. Components nest via `body`, `items`, `columns`, or component-specific props:

```json
{
  "type": "page",
  "title": "Page Title",
  "toolbar": [],
  "body": {
    "type": "crud",
    "api": "/api/v1/resource",
    "columns": []
  }
}
```

## Core Component Types

| Type | Purpose |
|------|---------|
| `page` | Top-level page container |
| `crud` | List + search + pagination + actions |
| `form` | Create / update form |
| `dialog` | Modal dialog |
| `drawer` | Slide-in panel |
| `tabs` | Tab container |
| `panel` | Card-like container with header |
| `grid` | Responsive column grid |
| `tpl` | Template string renderer |
| `button` | Action button |
| `service` | Data fetcher — injects data into children |

## API Format

AMIS `api` accepts either a string (GET shorthand) or an object:

```json
// Shorthand
"api": "/api/v1/contracts"

// Full object
"api": {
  "method": "post",
  "url": "/api/v1/contracts",
  "data": {
    "title": "${title}",
    "contract_type": "${contract_type}"
  },
  "responseData": {
    "id": "${id}"
  }
}
```

AMIS automatically:
- Injects the auth `fetcher` from the app shell
- Shows loading state
- Displays error toasts on non-2xx responses
- Maps response data into the component's data context

## Data Context and `${}` Expressions

AMIS uses a cascading data context. `${fieldName}` reads from the nearest context scope:

```json
{
  "type": "tpl",
  "tpl": "Contract ${contract_number} — Status: ${status}"
}
```

In forms, field `name` maps to context keys. Nested objects use dot notation: `${vendor.name}`.

## Screen File Conventions

```
web/schemas/pages/
  {module}/
    list.json       # CRUD list screen
    detail.json     # Read-only detail view
    create.json     # Create form
    edit.json       # Edit form (reuses form body)
```

Use `service` component at root to pre-fetch data for detail/edit screens:

```json
{
  "type": "service",
  "api": "/api/v1/contracts/${id}",
  "body": {
    "type": "form",
    "title": "Edit Contract",
    "api": { "method": "put", "url": "/api/v1/contracts/${id}" },
    "body": []
  }
}
```

## Permissions in Schemas

Gate buttons and actions on session context using AMIS `visibleOn`:

```json
{
  "type": "button",
  "label": "Approve",
  "visibleOn": "${PERMISSIONS && PERMISSIONS['contracts.contract.approve']}"
}
```

The app shell injects `PERMISSIONS` into the root data context after login.

## Required Schema Checklist

Before shipping a new screen:

- [ ] `title` set on `page` component
- [ ] `api` uses relative path (no hardcoded domain)
- [ ] Forms have `required: true` on mandatory fields
- [ ] List screens have at least one search filter
- [ ] Delete actions have `confirmText`
- [ ] Monetary fields use `type: "number"` with `precision: 6`
- [ ] Status fields use `type: "mapping"` with color-coded map

## Prohibited Patterns

- No `tenant_id` in schema data — comes from session, not user input
- No `api: "http://..."` absolute URLs — always relative paths
- No `disabledOn: "false"` hacks — use `visibleOn` for conditional display
- No inline styles — use AMIS CSS class props or `web/styles/theme.css`
