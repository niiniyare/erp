[<-- Back to Index](README.md)

## AMIS Core Concepts

### JSON Schema

Every AMIS page is a JSON object. The `type` field tells the runtime which renderer to use.

```json
{
  "type": "page",
  "title": "Purchase Orders",
  "body": {
    "type": "crud",
    "api": "/api/v1/purchase-orders",
    "columns": [
      { "name": "reference",  "label": "Reference" },
      { "name": "supplier",   "label": "Supplier" },
      { "name": "total",      "label": "Total",  "type": "number" }
    ]
  }
}
```

This single JSON object produces a paginated, filterable, sortable table. Zero JavaScript required.

### Data Domain (Data Chain)

The data domain is AMIS's state management. It is a tree of lexical scopes — one per container component. A component resolves `${name}` by walking upward through scopes, exactly like variable scope in a programming language.

```markdown
page  data: { tenant_name: "Acme Ltd" }
  └── crud  data: { items: [...] }
        └── dialog  data: { row being edited }
              │
              ├── ${id}            → from dialog data (the row)
              ├── ${tenant_name}   → from page data (walks up)
              └── ${$query.tab}    → from URL query string
```

Template expressions inside any string value:

```markdown
${field}                   → simple substitution
${amount | number:2}       → format with filter
${UPPER(name)}             → built-in formula
${{amount > 1000 ? 'large' : 'small'}}   → inline expression
```

### API Object

The `api` property is not just a URL. It controls the full HTTP request and response:

```json
{
  "api": {
    "method": "post",
    "url": "/api/v1/purchase-orders",
    "data": {
      "supplier_id": "${supplier_id}",
      "tenant_id":   "${__tenant}"
    },
    "headers": { "X-Tenant": "${__tenant}" },
    "responseData": {
      "order_id": "${data.id}"
    }
  }
}
```

Shorthand: `"api": "post:/api/v1/purchase-orders"` (method:url string).

### Actions and Events

Declarative event system — no JavaScript needed for most interactions:

```json
{
  "type": "button",
  "label": "Approve",
  "onEvent": {
    "click": {
      "actions": [
        { "actionType": "ajax",   "api": "post:/api/v1/approvals/${id}/approve" },
        { "actionType": "toast",  "args": { "msg": "Approved", "msgType": "success" } },
        { "actionType": "reload", "componentId": "po-list-crud" }
      ]
    }
  }
}
```

Built-in action types:

```markdown
ajax      → call an API
toast     → show a toast notification
dialog    → open a dialog
drawer    → open a side drawer
link      → navigate to a URL
reload    → reload a component by id
setData   → update the data domain
close     → close the current dialog/drawer
goBack    → browser back
copy      → copy to clipboard
broadcast → send event to other components
custom    → JavaScript escape hatch (use sparingly)
```

### `service` Component

The `service` component loads data AND optional schemas from Go. This is the mechanism for per-tenant, per-flag dynamic pages.

```json
{
  "type": "service",
  "api":       "/api/v1/dashboard/summary",
  "schemaApi": "/schema/dashboard",
  "body": [
    { "type": "stat", "source": "${total_receivables}", "label": "Receivables" },
    { "type": "stat", "source": "${overdue_count}",     "label": "Overdue" }
  ]
}
```

When `schemaApi` is set: AMIS calls it, receives a schema object, renders that schema as the body — replacing whatever static `body` was defined. This allows Go to return completely different UI per tenant/flag without any frontend change.

### `app` Component (Not Used in Awo)

AMIS has a built-in `app` component for the full shell. Awo uses a custom HTML shell instead. See [§02 Why AMIS](./02-why-amis.md) for the reasoning. The `app` component is documented here for completeness in case you read AMIS documentation that references it.

---
