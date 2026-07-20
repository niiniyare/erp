> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

# Chapter 21 — Event System

| | |
|---|---|
| **Volume** | 04 — Rendering |
| **Chapter** | 21 |
| **Status** | Partially implemented (AMIS onEvent/data chain implemented; server push planned) |
| **Source** | AMIS SDK event model; `internal/web/dsl/ast/` (FilterBarBlock, DataTableBlock) |

---

## Table of Contents

1. [21.1 AMIS Event-Action Model](#211-amis-event-action-model)
2. [21.2 Filter Bar → Table Linkage](#212-filter-bar--table-linkage)
3. [21.3 Form Change Events](#213-form-change-events)
4. [21.4 Data Chain Propagation](#214-data-chain-propagation)
5. [21.5 Planned: Server-Side Event Push](#215-planned-server-side-event-push)

---

## 21.1 AMIS Event-Action Model

AMIS provides a declarative **onEvent** system that allows any component to
react to events fired by itself or by sibling/parent components. An event
carries a data payload that is merged into the target scope.

### Schema structure

```json
{
  "type": "some-component",
  "id": "myComponent",
  "onEvent": {
    "<event-name>": {
      "actions": [
        {
          "actionType": "<action-type>",
          "componentId": "<target-id>",
          "args": { /* action arguments */ }
        }
      ]
    }
  }
}
```

### Common event names

| Event name | Fired by |
|---|---|
| `change` | Any input field when its value changes |
| `search` | Filter bar / search input on submit |
| `submit` | Form on successful submit |
| `success` | Ajax action on HTTP success |
| `failed` | Ajax action on HTTP failure |
| `click` | Button |
| `reload` | CRUDNode when its data is refreshed |
| `fetchInited` | Service or page after `initApi` completes |

### Common action types in onEvent

| ActionType | Effect |
|---|---|
| `reload` | Refreshes the target component's data |
| `setValue` | Sets a value in the target component's scope |
| `broadcast` | Sends a named event to all listening components |
| `toast` | Shows a toast notification |
| `dialog` | Opens a dialog |

---

## 21.2 Filter Bar → Table Linkage

The most common event pattern in AwoERP is a filter bar that drives a data
table. When the user changes filter values and clicks "Search", the CRUDNode
reloads with the new filter parameters.

### How it works

AMIS CRUDNode has a built-in `filter` field. When `filter` is defined:

1. AMIS renders the filter form above the table.
2. When the user submits the filter, AMIS merges the filter form values into the
   CRUDNode's API query parameters and reloads the table.
3. No `onEvent` wiring is required — AMIS handles this automatically.

### Schema example

```json
{
  "type": "crud",
  "id": "invoiceTable",
  "api": "/api/v1/finance/invoices",
  "syncLocation": true,
  "filter": {
    "title": "Filter Invoices",
    "body": [
      {
        "type": "input-date-range",
        "name": "date_range",
        "label": "Date Range"
      },
      {
        "type": "select",
        "name": "status",
        "label": "Status",
        "options": [
          { "label": "Draft",    "value": "DRAFT" },
          { "label": "Approved", "value": "APPROVED" },
          { "label": "Paid",     "value": "PAID" }
        ]
      }
    ]
  },
  "columns": [ /* ... */ ]
}
```

When the user submits the filter form, AMIS appends the values to the CRUD API
URL as query parameters:

```
GET /api/v1/finance/invoices?date_range=2026-01-01,2026-03-31&status=APPROVED
```

### Custom filter → table linkage with onEvent

For more complex cases — e.g. a standalone filter bar outside the CRUDNode —
`onEvent` is used to broadcast a reload:

```json
{
  "type": "form",
  "id": "filterBar",
  "onEvent": {
    "submit": {
      "actions": [
        {
          "actionType": "reload",
          "componentId": "invoiceTable",
          "args": {
            "query": "${event.data}"
          }
        }
      ]
    }
  },
  "body": [ /* filter fields */ ]
}
```

The `FilterBarBlock` in the DSL AST layer generates this pattern. It emits the
filter-change event by wiring the form's `submit` onEvent to a `reload` action
targeting the `DataTableBlock`'s CRUD component by ID.

---

## 21.3 Form Change Events

Individual form fields can react to their own changes using `onChange` hooks or
the `onEvent.change` pattern.

### Use case: dependent dropdowns

When a "Country" field changes, reload the "State" dropdown options:

```json
{
  "type": "select",
  "name": "country",
  "label": "Country",
  "onEvent": {
    "change": {
      "actions": [
        {
          "actionType": "reload",
          "componentId": "stateSelect"
        }
      ]
    }
  }
},
{
  "type": "select",
  "name": "state",
  "id": "stateSelect",
  "label": "State",
  "source": "/api/v1/geo/states?country=${country}"
}
```

When `country` changes, the `stateSelect` component reloads its `source` URL
with the new `country` value from scope.

### Use case: computed fields

```json
{
  "type": "input-number",
  "name": "quantity",
  "onEvent": {
    "change": {
      "actions": [
        {
          "actionType": "setValue",
          "componentId": "totalField",
          "args": { "value": "${quantity * unit_price}" }
        }
      ]
    }
  }
}
```

Note: for simple computed fields, AMIS expressions directly in the `value` field
are simpler than `onEvent`. Use `onEvent` only when the computation requires
triggering a side effect.

---

## 21.4 Data Chain Propagation

AMIS propagates data changes down the component tree automatically. This is the
primary mechanism for reactive UI — no explicit event wiring is needed in most
cases.

### How propagation works

1. A component modifies its own scope (e.g. a form field changes).
2. AMIS re-evaluates all `${expression}` strings in the same and child scopes.
3. Components that depend on the changed variable re-render.

### Example: page-level data driving buttons

```
PageNode.Data: { can_approve: true, invoice_id: "abc-123" }
  ↓ (data chain)
  ButtonNode: label="Approve", disabledOn="${!can_approve}"
  → evaluates to: disabled=false (button enabled)
```

If the pipeline changes `can_approve` to `false` for a different user and serves
a fresh schema, AMIS will render the button as disabled without any client-side
logic.

### Service component as data bridge

A `"type": "service"` component fetches its own data from an API and merges
the result into a child scope:

```json
{
  "type": "service",
  "api": "/api/v1/finance/invoices/${id}/line-items",
  "body": [
    {
      "type": "table",
      "source": "${items}",
      "columns": [ /* ... */ ]
    }
  ]
}
```

The `items` variable from the service API response is visible only within the
service's `body`. The parent page scope is also visible (line items can
reference `${invoice_id}` from the page scope).

---

## 21.5 Planned: Server-Side Event Push

> **Status: Planned — Not Yet Implemented**

The following real-time capabilities are planned but not built:

### Real-time schema updates

When a schema changes (e.g. a feature flag is toggled), the current
implementation requires the user to navigate away and back. A planned
improvement is a WebSocket or SSE channel from which the shell receives
invalidation signals:

```
Server sends: { "type": "schema_invalidated", "route": "finance/dashboard" }
Shell: if current route matches → navigate() (re-fetch and re-mount)
```

### Server-sent data events

For dashboards showing live counters (e.g. pending invoice count), the planned
approach is SSE:

```
GET /api/v1/events/stream
→ data: { "event": "invoice_count", "value": 12 }
```

The shell subscribes and pushes the value into the AMIS page scope via
`currentInstance.updateProps()`.

### Collaborative editing guards

When two users open the same record simultaneously, the server could push a
presence indicator via SSE. The UI would show "User X is also editing this
record." No implementation exists today.

Until real-time push is implemented, all data is fetched on navigation only.
Users must manually reload pages to see changes made by other users or by
background workflows.
