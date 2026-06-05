# Chapter 24 — Data Sources

| | |
|---|---|
| **Volume** | 05 — Runtime |
| **Chapter** | 24 |
| **Status** | Implemented |
| **Source** | `internal/web/dsl/ast/` (APISpec, PageNode, CRUDNode, ServiceNode) |

---

## Table of Contents

1. [24.1 APISpec Struct](#241-apispec-struct)
2. [24.2 PageNode.InitAPI](#242-pagenodesinitiapi)
3. [24.3 CRUDNode API](#243-crudnode-api)
4. [24.4 ServiceNode](#244-servicenode)
5. [24.5 AMIS Expression Access](#245-amis-expression-access)
6. [24.6 ResponseData Filtering Bug Fix](#246-responsedata-filtering-bug-fix)
7. [24.7 Permission Scope Variables](#247-permission-scope-variables)

---

## 24.1 APISpec Struct

`APISpec` is the canonical representation of an HTTP data source in the DSL
layer. It is used wherever a component needs to fetch or submit data:

```go
// APISpec describes an HTTP endpoint used for data fetching or submission.
type APISpec struct {
    Method       string         // HTTP verb: "GET", "POST", "PUT", "PATCH", "DELETE"
    URL          string         // Endpoint path, may contain ${expr} variables
    ResponseData map[string]any // DEPRECATED — do not use (see section 24.6)
}
```

### Usage across node types

| Node type | Field | Purpose |
|---|---|---|
| `PageNode` | `InitAPI` | Fetch record data at page load |
| `CRUDNode` | `API` | Fetch list data for the table |
| `CRUDNode` | `QuickSaveAPI` | Inline cell editing |
| `CRUDNode` | `SaveOrderAPI` | Drag-to-reorder rows |
| `FormNode` | `API` | Submit form data (create or update) |
| `ServiceNode` | `API` | Fetch sub-section data within a page |
| `ActionNode` | `API` | Ajax action endpoint |

### Compiled AMIS output

An `APISpec{Method: "GET", URL: "/api/v1/finance/invoices/${id}"}` compiles to:

```json
{
  "method": "GET",
  "url": "/api/v1/finance/invoices/${id}"
}
```

The `${id}` expression is resolved by AMIS at runtime using the current data
scope. Before the request is sent, AMIS interpolates all `${...}` segments in
the URL.

---

## 24.2 PageNode.InitAPI

`PageNode.InitAPI` is an optional `*APISpec` that, when present, causes AMIS to
make an HTTP GET request immediately after the page mounts. The response data is
merged into the top-level page scope.

### Go definition

```go
type PageNode struct {
    Title   string
    InitAPI *APISpec       // Optional. Fetches record data at page load.
    Data    map[string]any // Static data merged into page scope (see section 24.7)
    Body    []Node
}
```

### Example: invoice detail page

```go
PageNode{
    Title: "Invoice Detail",
    InitAPI: &APISpec{
        Method: "GET",
        URL:    "/api/v1/finance/invoices/${id}",
    },
    Data: map[string]any{
        "can_approve": sess.HasPermission("finance:invoices:approve"),
        "tenant_currency": sess.TenantCurrency,
    },
    Body: []Node{
        // Body nodes can reference ${number}, ${amount}, ${status}
        // from the initApi response, and ${can_approve}, ${tenant_currency}
        // from Data.
    },
}
```

### Compiled AMIS output

```json
{
  "type": "page",
  "title": "Invoice Detail",
  "initApi": {
    "method": "GET",
    "url": "/api/v1/finance/invoices/${id}"
  },
  "data": {
    "can_approve": true,
    "tenant_currency": "USD"
  },
  "body": [ /* ... */ ]
}
```

### Timing

AMIS mounts the page, then immediately fires the `initApi` request. Until the
response arrives, the body components render with whatever data is in `Data`
(the static injections). Once the response arrives, AMIS merges it into the
page scope and re-renders any components that reference the new variables.

For list pages (no specific record), `InitAPI` is typically nil. The data is
fetched by `CRUDNode.API` instead.

---

## 24.3 CRUDNode API

`CRUDNode.API` is the primary data source for list/table screens. AMIS sends a
GET request to this URL whenever the table renders, the user changes filters,
sorts a column, or pages through results.

### Go definition

```go
type CRUDNode struct {
    API          *APISpec       // Required. List data source.
    QuickSaveAPI *APISpec       // Optional. Inline cell edit endpoint.
    SaveOrderAPI *APISpec       // Optional. Row reorder endpoint.
    SyncLocation bool           // Must be true (ValidateStage enforces this).
    Filter       *FilterSpec    // Optional. Filter form above the table.
    Columns      []ColumnNode   // Table columns.
    RowActions   []ActionNode   // Per-row action buttons.
    Toolbar      []Node         // Header toolbar (create button, export, etc.)
}
```

### Compiled AMIS output (partial)

```json
{
  "type": "crud",
  "api": {
    "method": "GET",
    "url": "/api/v1/finance/invoices"
  },
  "syncLocation": true,
  "columns": [ /* ... */ ],
  "headerToolbar": [ /* ... */ ]
}
```

### Pagination

AMIS CRUDNode sends pagination as `page` and `perPage` query parameters. The
`amisEnv.fetcher` hook in `index.html` translates these to `offset` and `limit`
before the request reaches the backend:

```
AMIS sends:    GET /api/v1/finance/invoices?page=2&perPage=20
Fetcher sends: GET /api/v1/finance/invoices?offset=20&limit=20
```

The backend responds with:

```json
{
  "success": true,
  "data": [ /* items */ ],
  "meta": { "pagination": { "total_records": 84 } }
}
```

The `fetcher` normalises this to the AMIS format:

```json
{
  "status": 0,
  "data": {
    "items": [ /* items */ ],
    "count": 84
  }
}
```

AMIS uses `count` to calculate the total number of pages and render the
pagination controls.

### QuickSaveAPI

When a column is marked `quickEdit: true`, AMIS renders an inline editor when
the user clicks the cell. On change, it calls `QuickSaveAPI`:

```go
QuickSaveAPI: &APISpec{
    Method: "PUT",
    URL:    "/api/v1/finance/invoices/${id}/quick-save",
},
```

The request body contains only the changed field: `{ "field": "notes", "value": "Updated" }`.

### SaveOrderAPI

When rows are drag-reorderable, AMIS calls `SaveOrderAPI` with the new order:

```json
{ "ids": ["uuid-3", "uuid-1", "uuid-2"] }
```

---

## 24.4 ServiceNode

`ServiceNode` is a headless data-fetching component. It fetches data from its
`API` and merges the result into a child scope visible only to its `Body` nodes.

It is used to fetch sub-section data without making a full page `initApi` call.
Common use cases:

- A panel showing related records (e.g. line items on an invoice page).
- A sidebar showing summary statistics alongside a form.
- A tab that fetches data only when the user switches to it.

### Go definition

```go
type ServiceNode struct {
    API  *APISpec // Required.
    Body []Node   // Rendered with the fetched data in scope.
}
```

### Compiled AMIS output

```json
{
  "type": "service",
  "api": {
    "method": "GET",
    "url": "/api/v1/finance/invoices/${id}/line-items"
  },
  "body": [
    {
      "type": "table",
      "source": "${items}",
      "columns": [ /* ... */ ]
    }
  ]
}
```

The `items` variable from the service response is available to the `table` node
via `"source": "${items}"`. The parent page scope is also available — the table
can reference `${tenant_currency}` from the page `Data`.

### ServiceNode vs. CRUDNode

| Feature | ServiceNode | CRUDNode |
|---|---|---|
| Pagination | No | Yes |
| Filter bar | No | Yes |
| Sort columns | No | Yes |
| Row actions | No | Yes |
| Simple data display | Yes | Overkill |
| Fetches on mount | Yes | Yes |

Use `ServiceNode` for read-only sub-sections with fixed data. Use `CRUDNode`
for interactive lists with sort, filter, and pagination.

---

## 24.5 AMIS Expression Access

Once data is in the AMIS scope (from `initApi`, `Data`, or a `ServiceNode`),
any child component can reference it using `${fieldName}` expressions in any
string field.

### Examples

```json
// Page scope: { invoice: { id, number, amount }, tenant_currency: "USD" }

// Reference in a template column
{ "type": "tpl", "tpl": "Invoice #${invoice.number}" }

// Reference in an API URL
{ "url": "/api/v1/finance/invoices/${invoice.id}/approve" }

// Reference in a button label
{ "label": "Pay ${tenant_currency} ${invoice.amount}" }

// Reference in a conditional expression
{ "disabledOn": "${invoice.status === 'PAID'}" }

// Nested access
{ "tpl": "${address.city}, ${address.country}" }

// Array access
{ "tpl": "${line_items[0].description}" }
```

### Expression evaluation

AMIS evaluates expressions using its own lightweight formula engine. Supported
operations:

- Property access: `${obj.field}`, `${arr[0]}`
- Arithmetic: `${quantity * unit_price}`
- Comparison: `${status === 'APPROVED'}`, `${amount > 1000}`
- Boolean: `${!can_approve}`, `${can_edit && !locked}`
- Ternary: `${status === 'PAID' ? 'Settled' : 'Outstanding'}`
- String methods: `${name.toUpperCase()}`

Expressions are only evaluated — they cannot make HTTP calls or access the DOM.

---

## 24.6 ResponseData Filtering Bug Fix

`APISpec.ResponseData` was originally added to filter which fields from a
backend response were put into AMIS scope. For example:

```go
// Intended: only expose "invoice" key, hide internal fields
ResponseData: map[string]any{
    "invoice": "${invoice}",
}
```

**This caused a bug**: it also blocked pipeline-injected scope variables like
`${totals}`, `${tax_lines}`, and `${can_approve}` from reaching the AMIS scope.
Child components that referenced these variables rendered blank.

The fix was to **remove `ResponseData` filtering entirely**. The backend now
controls which fields appear in the response at the API handler level (the
handler only serialises what it should expose). AMIS receives the full response
and puts all fields into scope.

**Current status**: `APISpec.ResponseData` field is retained in the struct for
backward compatibility but is marked deprecated and must not be set. All
existing usages have been removed. The compiled AMIS schema omits
`responseData` entirely.

If field-level filtering is needed in the future, it should be implemented in
the API handler response DTO, not in the AMIS schema.

---

## 24.7 Permission Scope Variables

The pipeline injects pre-resolved permission booleans and tenant context into
`PageNode.Data` before the schema is served. This is the mechanism by which
the server controls UI behaviour without exposing raw IAM permission strings to
the browser.

### What is injected

```go
// Injected by the pipeline for every authenticated page request:
PageNode.Data = map[string]any{
    // Permission booleans (resolved from user's IAM session)
    "can_create":   sess.HasPermission(resource + ":create"),
    "can_read":     sess.HasPermission(resource + ":read"),
    "can_update":   sess.HasPermission(resource + ":update"),
    "can_delete":   sess.HasPermission(resource + ":delete"),
    "can_approve":  sess.HasPermission(resource + ":approve"),
    "can_export":   sess.HasPermission(resource + ":export"),

    // Tenant context
    "tenant_id":       sess.TenantID,
    "tenant_name":     sess.TenantName,
    "tenant_currency": sess.TenantCurrency,
    "tenant_timezone": sess.TenantTimezone,

    // User context
    "user_id":   sess.UserID,
    "user_name": sess.UserFullName,
    "user_email": sess.UserEmail,
}
```

Module pages may inject additional domain-specific booleans (e.g.
`can_void`, `can_reconcile`, `can_close_period`) by extending `Data` in their
`ASTFn` implementation.

### Why booleans and not permission strings

The browser receives `can_approve: true`, not `"finance:invoices:approve"`. This
means:

1. The browser cannot enumerate the tenant's IAM permission names.
2. An attacker who intercepts the schema cannot discover undocumented permissions.
3. The UI code (Go AST, static schemas) is decoupled from the IAM permission
   naming scheme — renaming a permission does not require schema changes.

### How AMIS uses them

```json
{
  "type": "button",
  "label": "Approve",
  "disabledOn": "${!can_approve}",
  "visibleOn":  "${can_approve || can_admin}"
}
```

AMIS reads `can_approve` from the page scope (populated from `PageNode.Data`)
and evaluates the expression. No HTTP call is made. The decision was made
server-side; the browser only acts on the result.
