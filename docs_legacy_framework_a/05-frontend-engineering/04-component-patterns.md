> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Component Patterns
portal: 5 — Frontend Engineering
section: 05-frontend-engineering
audience: [frontend-engineer]
related:
  - "[AMIS Schema Guide](02-amis-schema-guide.md)"
  - "[Dark Theme Guide](03-dark-theme-guide.md)"
  - "[API Reference](../08-api-reference/01-api-reference-overview.md)"
---

# Component Patterns

## Status Badge

Display colored status labels using `mapping` component:

```json
{
  "type": "mapping",
  "value": "${status}",
  "map": {
    "draft":        "<span class='badge badge-secondary'>Draft</span>",
    "under_review": "<span class='badge badge-warning'>Under Review</span>",
    "approved":     "<span class='badge badge-info'>Approved</span>",
    "active":       "<span class='badge badge-success'>Active</span>",
    "suspended":    "<span class='badge badge-danger'>Suspended</span>",
    "terminated":   "<span class='badge badge-dark'>Terminated</span>"
  }
}
```

For table columns use `"type": "tpl"` with inline styles when CSS classes aren't available:

```json
{
  "type": "tpl",
  "tpl": "<span style='color: ${status == 'active' ? '#52c41a' : '#999'}'>${status}</span>"
}
```

## Confirm-Before-Action Button

All destructive or irreversible actions must show a confirmation dialog:

```json
{
  "type": "button",
  "label": "Terminate Contract",
  "level": "danger",
  "confirmText": "Terminate this contract? This cannot be undone.",
  "actionType": "ajax",
  "api": {
    "method": "post",
    "url": "/api/v1/contracts/${id}/terminate",
    "data": {
      "version": "${version}",
      "reason": "Manual termination"
    }
  },
  "onEvent": {
    "submitSucc": {
      "actions": [
        { "actionType": "toast", "args": { "msgType": "success", "msg": "Contract terminated." } },
        { "actionType": "reload", "componentId": "contracts-list" }
      ]
    }
  }
}
```

## Optimistic Lock Version

Always include `version` in update/delete/action requests. Set it from the current record data:

```json
{
  "type": "form",
  "api": {
    "method": "put",
    "url": "/api/v1/contracts/${id}",
    "data": {
      "title":       "${title}",
      "total_value": "${total_value}",
      "version":     "${version}"
    }
  },
  "body": [
    { "type": "hidden", "name": "version" },
    { "type": "input-text", "name": "title", "label": "Title", "required": true }
  ]
}
```

On 409 response, show a message telling the user to refresh and retry:

```json
{
  "onEvent": {
    "submitFail": {
      "actions": [{
        "actionType": "toast",
        "args": {
          "msgType": "warning",
          "msg": "Record was modified by another user. Please refresh and try again."
        }
      }]
    }
  }
}
```

## Permission-Gated Buttons

Use `visibleOn` to hide buttons the user cannot access:

```json
{
  "type": "button",
  "label": "Approve",
  "visibleOn": "${permissions | includes: 'contracts.contract.approve'}",
  "actionType": "ajax",
  "api": "post:/api/v1/contracts/${id}/approve"
}
```

Inject `permissions` into AMIS initial data from the session token response:

```javascript
const me = await fetch('/api/v1/auth/me').then(r => r.json());
amis.embed('#app', schema, {
    permissions: me.permissions ?? [],
    tenant_id:   me.tenant_id,
    user_id:     me.user_id,
}, { theme: 'cxd' });
```

## Inline Editable Table

Use `table` with `editable` column configs for in-place editing:

```json
{
  "type": "crud",
  "api": "/api/v1/contracts",
  "saveApi": {
    "method": "put",
    "url": "/api/v1/contracts/${id}",
    "data": { "title": "${title}", "version": "${version}" }
  },
  "columns": [
    {
      "name": "title",
      "label": "Title",
      "quickEdit": {
        "type": "input-text",
        "saveImmediately": true
      }
    },
    { "name": "status", "label": "Status" }
  ]
}
```

## Monetary Value Display

Always display monetary values with currency symbol and 2 decimal places. Never display raw `numeric(20,6)` precision:

```json
{
  "type": "tpl",
  "tpl": "${currency} ${total_value | number: 2}"
}
```

For input fields, use `input-number` with `precision: 2`:

```json
{
  "type": "input-number",
  "name": "total_value",
  "label": "Total Value",
  "precision": 2,
  "min": 0,
  "required": true
}
```

## Date Range Filter

Standard filter pattern for list pages:

```json
{
  "type": "form",
  "mode": "inline",
  "target": "contracts-list",
  "body": [
    {
      "type": "input-date-range",
      "name": "date_range",
      "label": "Date Range",
      "format": "YYYY-MM-DD",
      "joinValues": false,
      "startPlaceholder": "Start date",
      "endPlaceholder": "End date"
    },
    {
      "type": "select",
      "name": "status",
      "label": "Status",
      "clearable": true,
      "options": [
        { "label": "Draft", "value": "draft" },
        { "label": "Active", "value": "active" },
        { "label": "Terminated", "value": "terminated" }
      ]
    },
    { "type": "submit", "label": "Search" },
    { "type": "reset", "label": "Clear" }
  ]
}
```

## Empty State

Always define `placeholder` for empty CRUD tables:

```json
{
  "type": "crud",
  "placeholder": "No contracts found. Create your first contract to get started.",
  ...
}
```

## Loading / Error States

AMIS handles loading state automatically for `api` calls. For error state, define `messages`:

```json
{
  "type": "service",
  "api": "/api/v1/contracts/${id}",
  "messages": {
    "fetchFailed": "Failed to load contract. Please refresh.",
    "fetchSuccess": ""
  }
}
```

## Patterns to Check Before Building New Components

Before creating a custom component, check:

1. `web/types/components/` — TypeScript type definitions showing available AMIS component props
2. `web/components/utils/ui/` — utility helpers (formatCurrency, formatDate, permissionCheck)
3. `web/styles/css/` — existing CSS classes (badge, status colors, layout helpers)

Most patterns already exist. Re-use, don't reinvent.
