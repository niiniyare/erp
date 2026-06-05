# Chapter 20 — Action System

| | |
|---|---|
| **Volume** | 04 — Rendering |
| **Chapter** | 20 |
| **Status** | Implemented |
| **Source** | `internal/web/dsl/ast/` (ActionNode, DialogNode, DrawerNode) |

---

## Table of Contents

1. [20.1 Action Types Overview](#201-action-types-overview)
2. [20.2 ActionNode Struct Fields](#202-actionnode-struct-fields)
3. [20.3 Ajax Actions](#203-ajax-actions)
4. [20.4 Dialog Actions](#204-dialog-actions)
5. [20.5 Drawer Actions](#205-drawer-actions)
6. [20.6 Conditional Enable / Disable](#206-conditional-enable--disable)
7. [20.7 Row Actions in CRUDNode](#207-row-actions-in-crudnode)
8. [20.8 Permission-Aware Actions](#208-permission-aware-actions)

---

## 20.1 Action Types Overview

Every interactive button, toolbar item, or row action in the UI is represented
by an `ActionNode`. The `ActionType` field determines what happens when the user
clicks.

| ActionType | What it does |
|---|---|
| `"ajax"` | Makes an HTTP call; shows success/error feedback |
| `"url"` | Navigates to a new route (hash or absolute path) |
| `"dialog"` | Opens a modal dialog containing a nested AMIS schema |
| `"drawer"` | Opens a side drawer containing a nested AMIS schema |
| `"reload"` | Triggers a data reload on the target component |
| `"link"` | Alias for `"url"` — navigates within the shell |

---

## 20.2 ActionNode Struct Fields

```go
// ActionNode represents a single interactive button or action.
// It compiles to an AMIS "button" schema fragment.
type ActionNode struct {
    Label      string      // Button label text. Required.
    ActionType string      // One of: "ajax", "url", "dialog", "drawer", "reload", "link"
    Level      string      // Visual style: "primary", "danger", "default", "link", "success"
    Icon       string      // FontAwesome class, e.g. "fa fa-plus"
    DisabledOn string      // AMIS expression: button is disabled when this evaluates to true
    VisibleOn  string      // AMIS expression: button is hidden when this evaluates to false
    API        *APISpec    // Required when ActionType == "ajax"
    Dialog     *DialogNode // Required when ActionType == "dialog"
    Drawer     *DrawerNode // Required when ActionType == "drawer"
    Target     string      // Required when ActionType == "url" or "link"
    Confirm    string      // Optional: shows a confirmation dialog with this text before acting
}
```

### Compiled JSON output

An `ActionNode` with `ActionType: "ajax"` compiles to:

```json
{
  "type": "button",
  "label": "Approve",
  "actionType": "ajax",
  "level": "primary",
  "icon": "fa fa-check",
  "disabledOn": "${!can_approve}",
  "confirmText": "Approve this invoice?",
  "api": {
    "method": "POST",
    "url": "/api/v1/finance/invoices/${id}/approve"
  }
}
```

---

## 20.3 Ajax Actions

Ajax actions make an HTTP request when clicked. The `API` field specifies the
endpoint.

### Go struct literal

```go
ActionNode{
    Label:      "Approve",
    ActionType: "ajax",
    Level:      "primary",
    Icon:       "fa fa-check",
    Confirm:    "Approve this invoice?",
    DisabledOn: "${!can_approve}",
    API: &APISpec{
        Method: "POST",
        URL:    "/api/v1/finance/invoices/${id}/approve",
    },
}
```

### AMIS JSON output

```json
{
  "type": "button",
  "label": "Approve",
  "actionType": "ajax",
  "level": "primary",
  "icon": "fa fa-check",
  "confirmText": "Approve this invoice?",
  "disabledOn": "${!can_approve}",
  "api": {
    "method": "POST",
    "url": "/api/v1/finance/invoices/${id}/approve"
  }
}
```

### Success and failure behaviour

AMIS handles ajax success/failure automatically:

- **Success** (`status: 0`): shows a green toast notification with `msg` if
  provided; triggers a reload of the closest CRUDNode or form.
- **Failure** (`status != 0`): shows a red alert with the `msg` from the
  response body.

No additional configuration is required in the `ActionNode` for this default
behaviour. To customise the reload target or redirect on success, use AMIS
`redirect` and `reload` fields — these can be added to the compiled schema as
extra keys if needed.

### Confirm dialogs

Setting `Confirm` causes AMIS to show a native confirmation dialog before the
HTTP call is made. The user must click "OK" to proceed. This is the recommended
pattern for destructive operations (delete, archive, cancel).

---

## 20.4 Dialog Actions

Dialog actions open a modal overlay containing a nested AMIS schema. The
`Dialog` field must be non-nil when `ActionType` is `"dialog"`.

### Go struct literal

```go
ActionNode{
    Label:      "Edit",
    ActionType: "dialog",
    Level:      "default",
    Icon:       "fa fa-pencil",
    Dialog: &DialogNode{
        Title: "Edit Invoice",
        Size:  "lg",
        Body: []Node{
            FormNode{
                API:    &APISpec{Method: "PUT", URL: "/api/v1/finance/invoices/${id}"},
                Fields: []FieldNode{ /* ... */ },
            },
        },
    },
}
```

### AMIS JSON output

```json
{
  "type": "button",
  "label": "Edit",
  "actionType": "dialog",
  "level": "default",
  "icon": "fa fa-pencil",
  "dialog": {
    "title": "Edit Invoice",
    "size": "lg",
    "body": [
      {
        "type": "form",
        "api": { "method": "PUT", "url": "/api/v1/finance/invoices/${id}" },
        "body": [ /* form fields */ ]
      }
    ]
  }
}
```

### Dialog sizes

| `Size` | Width |
|---|---|
| `"sm"` | ~400 px |
| `""` (default) | ~600 px |
| `"lg"` | ~800 px |
| `"xl"` | ~1000 px |
| `"full"` | Fills the viewport |

### Key constraint

If `ActionType` is `"dialog"` and `Dialog` is nil, the compiled schema will
produce a button with no dialog — clicking it does nothing. The `ValidateStage`
does not currently check for this; it is a programmer error to be caught in
testing.

---

## 20.5 Drawer Actions

Drawer actions open a side panel (right or bottom) containing a nested AMIS
schema. The `Drawer` field must be non-nil when `ActionType` is `"drawer"`.

Drawers are preferred over dialogs for detail views and multi-step forms because
they do not obscure the underlying list.

### Go struct literal

```go
ActionNode{
    Label:      "View Details",
    ActionType: "drawer",
    Level:      "link",
    Drawer: &DrawerNode{
        Title:    "Invoice Details",
        Position: "right",
        Size:     "md",
        Body: []Node{
            ServiceNode{
                API: &APISpec{Method: "GET", URL: "/api/v1/finance/invoices/${id}"},
                Body: []Node{ /* read-only panel */ },
            },
        },
    },
}
```

### AMIS JSON output

```json
{
  "type": "button",
  "label": "View Details",
  "actionType": "drawer",
  "level": "link",
  "drawer": {
    "title": "Invoice Details",
    "position": "right",
    "size": "md",
    "body": [
      {
        "type": "service",
        "api": { "method": "GET", "url": "/api/v1/finance/invoices/${id}" },
        "body": [ /* read-only panel */ ]
      }
    ]
  }
}
```

### Key constraint

If `ActionType` is `"drawer"` and `Drawer` is nil, the compiled schema produces
a non-functional button. Same programmer-error caveat as dialogs.

---

## 20.6 Conditional Enable / Disable

`ActionNode` has two conditional fields that control visibility and interactivity:

| Field | Effect when expression is truthy |
|---|---|
| `DisabledOn` | Button is rendered but greyed out and non-clickable |
| `VisibleOn` | Button is hidden entirely from the DOM |

Both accept AMIS expression strings evaluated in the current data scope. The
expressions can reference any variable in scope — most commonly the pre-resolved
`can_*` permission booleans injected by the pipeline (see Ch 24, section 24.7).

### Examples

```go
// Disable Approve button if user cannot approve
DisabledOn: "${!can_approve}"

// Hide Delete button if record is already archived
VisibleOn: "${status != 'ARCHIVED'}"

// Disable Submit if form has validation errors or record is locked
DisabledOn: "${locked || !form_valid}"
```

### Security constraint

`DisabledOn` and `VisibleOn` must **never** contain permission strings like
`"finance:invoices:approve"`. The `ValidateStage` (Ch 23, section 23.3) will
reject schemas that contain IAM permission strings in AMIS expressions.

Use the pre-resolved `can_*` booleans instead. The server resolves these before
serving the schema (Ch 24, section 24.7), so the browser receives a boolean and
cannot infer the underlying permission name.

---

## 20.7 Row Actions in CRUDNode

CRUD tables have a dedicated column for per-row actions. These are defined as
`RowActions []ActionNode` on `CRUDNode`.

### Placement requirement

Row actions must be included in `CRUDNode.Children()` output, not added as a
top-level button. AMIS requires them inside the column definition:

```json
{
  "type": "crud",
  "api": "/api/v1/finance/invoices",
  "columns": [
    { "name": "number",  "label": "Invoice #" },
    { "name": "amount",  "label": "Amount" },
    { "name": "status",  "label": "Status" },
    {
      "type": "operation",
      "label": "Actions",
      "buttons": [
        {
          "type": "button",
          "label": "Edit",
          "actionType": "dialog",
          "dialog": { /* ... */ }
        },
        {
          "type": "button",
          "label": "Delete",
          "actionType": "ajax",
          "level": "danger",
          "confirmText": "Delete this invoice?",
          "api": { "method": "DELETE", "url": "/api/v1/finance/invoices/${id}" }
        }
      ]
    }
  ]
}
```

### Row scope

Row actions operate in the **row scope** — all fields of the current row record
are available via `${fieldName}`. In the example above, `${id}` resolves to
the `id` field of the row the user clicked.

### Toolbar actions

Actions that operate on the whole list (e.g. "Create New", "Export") are placed
in `CRUDNode.HeaderToolbar` and operate in the page scope, not the row scope.

---

## 20.8 Permission-Aware Actions

The recommended pattern for permission-aware actions is:

1. **Server side** (pipeline): inject `can_approve: true/false` into `PageNode.Data`
   based on the user's resolved permissions (Ch 24, section 24.7).
2. **Schema side** (Go AST): set `DisabledOn: "${!can_approve}"` on the button.
3. **Browser side** (AMIS): evaluates the expression and disables the button.

This keeps permission logic on the server. The browser only receives a boolean.

### Complete example

```go
// PageNode.Data injection (pipeline)
data := map[string]any{
    "can_approve":      sess.HasPermission("finance:invoices:approve"),
    "can_delete":       sess.HasPermission("finance:invoices:delete"),
    "tenant_currency":  sess.TenantCurrency,
}

// ActionNode definitions (Go AST)
approveAction := ActionNode{
    Label:      "Approve",
    ActionType: "ajax",
    Level:      "primary",
    Icon:       "fa fa-check",
    DisabledOn: "${!can_approve}",
    Confirm:    "Approve this invoice?",
    API:        &APISpec{Method: "POST", URL: "/api/v1/finance/invoices/${id}/approve"},
}

deleteAction := ActionNode{
    Label:      "Delete",
    ActionType: "ajax",
    Level:      "danger",
    DisabledOn: "${!can_delete}",
    Confirm:    "Permanently delete this invoice?",
    API:        &APISpec{Method: "DELETE", URL: "/api/v1/finance/invoices/${id}"},
}
```

The compiled AMIS schema delivered to the browser:

```json
{
  "type": "button",
  "label": "Approve",
  "actionType": "ajax",
  "level": "primary",
  "icon": "fa fa-check",
  "disabledOn": "${!can_approve}",
  "confirmText": "Approve this invoice?",
  "api": { "method": "POST", "url": "/api/v1/finance/invoices/${id}/approve" }
}
```

The user's `can_approve` value (`true` or `false`) is already in the AMIS data
scope from `PageNode.Data`. AMIS evaluates `${!can_approve}` on the client.
No IAM permission string (`"finance:invoices:approve"`) ever reaches the browser.
