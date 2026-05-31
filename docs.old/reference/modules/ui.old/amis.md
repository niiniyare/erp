# Amis UI — Complete Integration Guide for Awo ERP

> **Scope:** This document is the authoritative reference for how Awo ERP uses
> Baidu amis as its frontend rendering engine. It covers the amis conceptual
> model, every component category relevant to ERP workflows, the exact API
> contract Go must satisfy, the `awo/web/schema` Go package design, Fiber
> route structure, authentication integration, and patterns for every major
> Awo UI surface.

---

## Table of Contents

1. [What amis Is](#1-what-amis-is)
2. [Core Concepts](#2-core-concepts)
3. [How amis Integrates with Awo](#3-how-amis-integrates-with-awo)
4. [Project Setup](#4-project-setup)
5. [API Contract — What Go Must Return](#5-api-contract--what-go-must-return)
6. [Component Reference](#6-component-reference)
7. [Go Schema Builder — `awo/web/schema`](#7-go-schema-builder--awowebschema)
8. [Fiber Route Structure](#8-fiber-route-structure)
9. [Authentication & Permission Gating](#9-authentication--permission-gating)
10. [Flag-Gated Pages](#10-flag-gated-pages)
11. [Awo Page Patterns](#11-awo-page-patterns)
12. [The Rules Builder UI in amis](#12-the-rules-builder-ui-in-amis)
13. [Theming](#13-theming)
14. [Known Limitations & Workarounds](#14-known-limitations--workarounds)
15. [Documentation Sources](#15-documentation-sources)

---

## 1. What amis Is

amis (Adaptive Management Interface System) is Baidu's open-source low-code
frontend framework. Rather than writing React component trees, you describe
pages in JSON. The amis runtime reads that JSON, resolves component types to
React renderers, fetches data from your API, and produces a fully interactive
UI.

```
Your Go backend returns JSON (the schema)
          │
          ▼
   amis runtime (React, runs in browser)
   reads schema → renders page → calls your data API → updates UI
```

**Why it is the right choice for Awo ERP:**

The components amis ships with are specifically designed for backend management
software: CRUD tables with filters/pagination/bulk actions, multi-step forms,
approval inboxes, dashboards with charts, wizard flows, drawer/dialog patterns.
Every ERP page you need to build has a direct amis analog. You do not adapt
amis to ERP use cases — amis was built for exactly this.

**What you do not write:**
- React component files
- CSS (unless theming)
- JavaScript event handlers (amis has a declarative event/action system)
- Pagination logic
- Filter state management
- Form validation wiring

**What you do write:**
- Go structs that produce amis-compatible JSON (your `schema` package)
- Go REST handlers that return data in amis's expected envelope format
- Occasional custom amis renderers for genuinely non-standard components

---

## 2. Core Concepts

### 2.1 JSON Schema

Every amis page is a JSON object. The root object always has a `type` field
that tells the runtime which renderer to use. Every property of that object is
either a static configuration value or a template expression that references the
data domain.

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
      { "name": "total",      "label": "Total",     "type": "number" },
      { "name": "status",     "label": "Status",    "type": "mapping",
        "map": { "draft": "Draft", "submitted": "Submitted", "confirmed": "Confirmed" } }
    ]
  }
}
```

This single JSON object produces a paginated, filterable, sortable table with
row actions — with zero JavaScript.

### 2.2 Data Domain (Data Chain)

The data domain is amis's state management model. It is a tree of lexical
scopes, one per container component (`page`, `crud`, `form`, `dialog`,
`service`). When a component resolves a variable `${name}`, it walks upward
through the scope chain until it finds the variable, exactly like lexical
scoping in a programming language.

```
page  data: { tenant_name: "Acme Ltd" }
  └── crud  data: { rows fetched from API }
        └── form (in dialog)  data: { row being edited }
              │
              ├── ${id}          → resolves from form's data (the row)
              ├── ${tenant_name} → resolves from page's data (walks up)
              └── ${$query.tab}  → resolves from URL query string
```

**Template syntax inside JSON string values:**
- `${field}` — simple interpolation
- `${field | date}` — with a built-in filter
- `${UPPER(field)}` — built-in formula function
- `${{field > 100 ? 'high' : 'low'}}` — inline expression

### 2.3 API Object

The `api` property in amis components is not just a URL string. It can be an
object with full control over the request and response:

```json
{
  "api": {
    "method": "post",
    "url": "/api/v1/purchase-orders",
    "data": {
      "supplier_id": "${supplier_id}",
      "line_items":  "${line_items}",
      "tenant_id":   "${__tenant}"
    },
    "headers": {
      "X-Tenant": "${__tenant}"
    },
    "responseData": {
      "order_id": "${data.id}"
    }
  }
}
```

`data` maps form fields into the request body using template expressions.
`responseData` extracts fields from the response and merges them into the
current data domain.

### 2.4 Actions and Events

amis has a declarative event/action system. Components emit events (`click`,
`change`, `submit`, `success`, `fail`, etc.). You attach action chains to those
events without writing JavaScript.

```json
{
  "type": "button",
  "label": "Approve",
  "onEvent": {
    "click": {
      "actions": [
        {
          "actionType": "ajax",
          "api": "post:/api/v1/approvals/${id}/approve"
        },
        {
          "actionType": "toast",
          "args": { "msg": "Approved successfully", "msgType": "success" }
        },
        {
          "actionType": "reload",
          "componentId": "po-list-crud"
        }
      ]
    }
  }
}
```

Built-in action types: `ajax` · `toast` · `dialog` · `drawer` · `link` ·
`reload` · `setData` · `custom` (JS escape hatch) · `broadcast` ·
`close` · `goBack` · `copy`

### 2.5 `service` Component — Schema Loading

The `service` component is central to Awo's architecture. It loads both data
**and schema** from your API. This means your Go backend can serve dynamic
schemas — per tenant, per feature flag, per user role.

```json
{
  "type": "service",
  "schemaApi": "/schema/accounting/journal",
  "api":       "/api/v1/accounting/journal/summary"
}
```

When the page loads, the `service` component calls `schemaApi`, receives a JSON
schema, renders it, and also calls `api` to populate the data domain. The
rendered component can be any amis schema — a full `crud`, a `form`, a
`dashboard`.

This is the pattern that makes per-tenant UI customisation possible: you return
a different schema based on the tenant's plan and flags.

### 2.6 `app` Component — Full Shell

For a multi-page application like Awo ERP, the root component is `app`. It
defines the navigation sidebar, header, and the routing rules for each nav item.

```json
{
  "type": "app",
  "brandName": "Awo ERP",
  "logo": "/static/logo.svg",
  "header": { "type": "tpl", "tpl": "${tenant_name}" },
  "pages": [
    {
      "label": "Accounting",
      "icon":  "fa fa-calculator",
      "children": [
        { "label": "Journal", "url": "/accounting/journal",
          "schema": { "type": "service", "schemaApi": "/schema/accounting/journal" } },
        { "label": "Accounts", "url": "/accounting/accounts",
          "schema": { "type": "service", "schemaApi": "/schema/accounting/accounts" } }
      ]
    },
    {
      "label": "Purchasing",
      "icon":  "fa fa-shopping-cart",
      "children": [
        { "label": "Purchase Orders", "url": "/purchasing/orders",
          "schema": { "type": "service", "schemaApi": "/schema/purchasing/orders" } }
      ]
    }
  ]
}
```

The `app` schema itself can be loaded dynamically from your Go backend, meaning
the entire navigation structure is driven by the tenant's enabled feature flags.

---

## 3. How amis Integrates with Awo

### Architecture Overview

```
Browser
  │
  │  (1) GET /  → Go returns index.html with amis bundle
  │
  ▼
amis runtime boots
  │
  │  (2) GET /schema/app → Go returns full app schema (nav + initial data)
  │
  ▼
amis renders sidebar, header, initial page
  │
  │  (3) User navigates → GET /schema/{module}/{page}
  │       → Go returns page schema (filtered by flags/permissions)
  │
  │  (4) amis executes data APIs defined inside the schema
  │       GET  /api/v1/{resource}           → list
  │       GET  /api/v1/{resource}/{id}      → detail
  │       POST /api/v1/{resource}           → create
  │       PUT  /api/v1/{resource}/{id}      → update
  │       DELETE /api/v1/{resource}/{id}    → delete
  │
  ▼
amis renders data into components, manages state locally
```

**Two route groups in Fiber:**

| Path prefix | Purpose | Auth | Response |
|---|---|---|---|
| `/schema/*` | Serve amis page schemas | Session cookie / Bearer | JSON schema |
| `/api/v1/*` | Serve business data | Session cookie / Bearer | JSON data envelope |
| `/` | Serve `index.html` | None | HTML |
| `/static/*` | Serve amis bundle + assets | None | Static files |

The existing `/api/v1/` layer documented in the main architecture document is
unchanged. amis is just another consumer of it.

### What Changes vs. the templ/HTMX Architecture

| Aspect | Before (templ/HTMX) | After (amis) |
|---|---|---|
| Rendering | Server-side HTML | Client-side, JSON-driven |
| `web/handlers` | Return HTML via templ | Return JSON data envelopes |
| Templates | `.templ` files | None |
| `HTMX` | Drives partial updates | Removed |
| `Alpine.js` | Local interactivity | Removed |
| New files | Go schema structs | `/schema/*` route handlers |
| Navigation | templ nav component | `app` schema from Go |
| Flag-gated UI | templ `if` blocks | Schema omits nav items / returns 404 |

Everything below `web/` — platform, core, analytics, business rules — is
completely unchanged.

---

## 4. Project Setup

### 4.1 Frontend (amis bundle)

amis is loaded as a pre-built bundle. You do not need a full React build
pipeline unless you are writing custom renderers.

**Option A — CDN (fastest to start):**

```html
<!-- awo/web/static/index.html -->
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8" />
  <title>Awo ERP</title>
  <link rel="stylesheet"
    href="https://unpkg.com/amis/sdk/sdk.css" />
  <link rel="stylesheet"
    href="https://unpkg.com/amis/sdk/helper.css" />
</head>
<body>
  <div id="root"></div>
  <script src="https://unpkg.com/amis/sdk/sdk.js"></script>
  <script>
    const { embed } = window.amis;
    embed('#root',
      { type: 'service', schemaApi: '/schema/app' },
      { data: {} },
      {
        fetcher({ url, method, data, headers }) {
          return fetch(url, {
            method: method.toUpperCase(),
            headers: { 'Content-Type': 'application/json', ...headers },
            body: method !== 'get' ? JSON.stringify(data) : undefined,
          }).then(r => r.json());
        },
        notify(type, msg) {
          console.log(type, msg);
        },
        alert(msg)   { window.alert(msg); },
        confirm(msg) { return window.confirm(msg); },
      }
    );
  </script>
</body>
</html>
```

**Option B — npm + Vite (for custom renderers):**

```bash
npm create vite@latest awo-frontend -- --template react-ts
cd awo-frontend
npm install amis
```

```tsx
// src/main.tsx
import React from 'react';
import ReactDOM from 'react-dom/client';
import { embed } from 'amis';
import 'amis/sdk/sdk.css';

const root = document.getElementById('root')!;
embed(root, { type: 'service', schemaApi: '/schema/app' }, {}, {
  fetcher: ({ url, method, data }) =>
    fetch(url, {
      method: method.toUpperCase(),
      headers: { 'Content-Type': 'application/json' },
      body: method !== 'get' ? JSON.stringify(data) : undefined,
    }).then(r => r.json()),
});
```

### 4.2 Serving Static Files from Go (Fiber)

```go
// awo/web/router/static.go

func registerStatic(app *fiber.App) {
    // amis bundle (built or copied from node_modules/amis/sdk/)
    app.Static("/static", "./web/static", fiber.Static{
        MaxAge: 86400,
        Compress: true,
    })

    // Catch-all: serve index.html for all non-API, non-schema routes
    // amis handles client-side routing
    app.Get("/*", func(c *fiber.Ctx) error {
        if strings.HasPrefix(c.Path(), "/api/") ||
           strings.HasPrefix(c.Path(), "/schema/") {
            return c.Next()
        }
        return c.SendFile("./web/static/index.html")
    })
}
```

---

## 5. API Contract — What Go Must Return

amis has a strict envelope format. Every API response must conform to it or
amis will show an error state. This is the most important section for Go
backend developers.

### 5.1 Standard Response Envelope

```json
{
  "status": 0,
  "msg": "",
  "data": { }
}
```

- `status`: `0` = success. Any non-zero value = error. amis shows `msg` as an
  error toast when status is non-zero.
- `msg`: Human-readable message. Shown to the user on error, optionally on
  success.
- `data`: The payload. Shape depends on the endpoint type.

```go
// awo/web/response/amis.go

type AmisResponse struct {
    Status int    `json:"status"`
    Msg    string `json:"msg,omitempty"`
    Data   any    `json:"data"`
}

func OK(data any) AmisResponse {
    return AmisResponse{Status: 0, Data: data}
}

func OKMsg(msg string, data any) AmisResponse {
    return AmisResponse{Status: 0, Msg: msg, Data: data}
}

func Err(msg string) AmisResponse {
    return AmisResponse{Status: 1, Msg: msg, Data: nil}
}

func ErrCode(code int, msg string) AmisResponse {
    return AmisResponse{Status: code, Msg: msg, Data: nil}
}
```

### 5.2 List Response (for `crud` component)

```json
{
  "status": 0,
  "msg": "",
  "data": {
    "items": [ { "id": "...", "reference": "PO-001", ... } ],
    "total": 142,
    "page":  1,
    "perPage": 20
  }
}
```

The `crud` component sends these query params automatically:
- `page` — current page number (1-based)
- `perPage` — items per page
- `orderBy` — column name
- `orderDir` — `asc` | `desc`
- Any filter field names you define in `filter`

```go
// awo/web/response/amis.go

type AmisListData struct {
    Items   any `json:"items"`
    Total   int `json:"total"`
    Page    int `json:"page"`
    PerPage int `json:"perPage"`
}

func OKList(items any, total, page, perPage int) AmisResponse {
    return AmisResponse{
        Status: 0,
        Data: AmisListData{
            Items:   items,
            Total:   total,
            Page:    page,
            PerPage: perPage,
        },
    }
}
```

### 5.3 Single Record Response (for `form` detail view)

```json
{
  "status": 0,
  "data": {
    "id":        "uuid",
    "reference": "PO-001",
    "supplier":  "Supplier Name",
    "total":     15000.00,
    "status":    "confirmed"
  }
}
```

The `form` component with `initApi` calls this to pre-populate fields.

### 5.4 Mutation Response (create / update)

```json
{
  "status": 0,
  "msg": "Purchase order created",
  "data": {
    "id": "new-uuid"
  }
}
```

amis merges `data` from the mutation response into the current data domain.
This allows the form to navigate to the new record after creation: reference
`${id}` in the `redirect` property.

### 5.5 Validation Error Response

```json
{
  "status": 422,
  "msg": "Validation failed",
  "errors": {
    "supplier_id": "Supplier is required",
    "line_items":  "At least one line item is required"
  }
}
```

amis reads the `errors` object and displays field-level error messages inline
in the form. The field `name` in your JSON schema must match the key in `errors`.

```go
type AmisValidationResponse struct {
    Status int               `json:"status"`
    Msg    string            `json:"msg"`
    Errors map[string]string `json:"errors,omitempty"`
}

func ValidationErr(msg string, fieldErrors map[string]string) AmisValidationResponse {
    return AmisValidationResponse{Status: 422, Msg: msg, Errors: fieldErrors}
}
```

### 5.6 Schema Response

Schema endpoints return a raw amis schema object — no envelope:

```json
{
  "type": "page",
  "title": "Purchase Orders",
  "body": { ... }
}
```

Do **not** wrap schema responses in the `{ status, data }` envelope. amis
renders the returned JSON directly.

### 5.7 Go Handler Pattern

```go
// awo/web/handlers/purchasing/purchase_order_handler.go

func (h *PurchaseOrderHandlers) List(c *fiber.Ctx) error {
    tenantID := middleware.ContextTenantID(c)

    filter := procurement.ListFilter{
        Page:    c.QueryInt("page", 1),
        PerPage: c.QueryInt("perPage", 20),
        OrderBy: c.Query("orderBy", "created_at"),
        OrderDir: c.Query("orderDir", "desc"),
        Status:  c.Query("status"),
        Search:  c.Query("keywords"),
    }

    orders, total, err := h.poSvc.List(c.Context(), tenantID, filter)
    if err != nil {
        return c.JSON(response.Err(err.Error()))
    }

    return c.JSON(response.OKList(orders, total, filter.Page, filter.PerPage))
}

func (h *PurchaseOrderHandlers) Create(c *fiber.Ctx) error {
    tenantID := middleware.ContextTenantID(c)
    userID   := middleware.ContextUserID(c)

    var input procurement.CreatePOInput
    if err := c.BodyParser(&input); err != nil {
        return c.JSON(response.Err("Invalid request body"))
    }

    po, err := h.poSvc.Create(c.Context(), tenantID, userID, input)
    if err != nil {
        var valErr *domain.ValidationError
        if errors.As(err, &valErr) {
            return c.JSON(response.ValidationErr(valErr.Message, valErr.Fields))
        }
        return c.JSON(response.Err(err.Error()))
    }

    return c.JSON(response.OKMsg("Purchase order created", map[string]any{"id": po.ID}))
}
```

---

## 6. Component Reference

This section covers the components you will use in every Awo ERP module.
Each entry shows the minimum viable schema, then the full production-grade
pattern.

### 6.1 `page`

The container for every screen. Provides the page title, toolbar actions, and
a body area.

```json
{
  "type": "page",
  "title":    "Purchase Orders",
  "subTitle": "Manage supplier purchase orders",
  "toolbar": [
    {
      "type":  "button",
      "label": "New Order",
      "level": "primary",
      "actionType": "link",
      "link": "/purchasing/orders/new"
    }
  ],
  "body": { "...": "content goes here" }
}
```

### 6.2 `crud`

The most important component in Awo. Handles list + filter + sort + pagination
+ row actions + bulk actions in a single declaration.

```json
{
  "type": "crud",
  "id":   "po-list",
  "api":  "get:/api/v1/purchase-orders",
  "syncLocation": true,

  "headerToolbar": [
    "search-box",
    "bulkActions",
    { "type": "columns-toggler" },
    { "type": "reload" }
  ],

  "filter": {
    "title": "Filter",
    "body": [
      { "type": "input-text",   "name": "keywords", "label": "Search",   "placeholder": "Reference or supplier" },
      { "type": "select",       "name": "status",   "label": "Status",
        "options": [
          { "label": "Draft",     "value": "draft" },
          { "label": "Submitted", "value": "submitted" },
          { "label": "Confirmed", "value": "confirmed" }
        ],
        "clearable": true
      },
      { "type": "input-date-range", "name": "dateRange", "label": "Date",
        "format": "YYYY-MM-DD" }
    ]
  },

  "columns": [
    { "name": "reference",     "label": "Reference",  "sortable": true },
    { "name": "supplier_name", "label": "Supplier",   "sortable": true },
    { "name": "total",         "label": "Total",
      "type": "number", "prefix": "KES ", "sortable": true },
    { "name": "status",        "label": "Status",
      "type": "tag",
      "colorMap": {
        "draft":     "default",
        "submitted": "processing",
        "confirmed": "success",
        "cancelled": "error"
      }
    },
    { "name": "created_at",    "label": "Date", "type": "date",
      "format": "MMM DD, YYYY", "sortable": true },
    {
      "type": "operation",
      "label": "Actions",
      "buttons": [
        {
          "label": "View",
          "type":  "button",
          "level": "link",
          "actionType": "link",
          "link": "/purchasing/orders/${id}"
        },
        {
          "label": "Submit",
          "type":  "button",
          "level": "link",
          "visibleOn": "${status === 'draft'}",
          "actionType": "ajax",
          "api": "post:/api/v1/purchase-orders/${id}/submit",
          "confirmText": "Submit this order for approval?",
          "onEvent": {
            "success": {
              "actions": [
                { "actionType": "reload", "componentId": "po-list" },
                { "actionType": "toast",  "args": { "msg": "Order submitted" } }
              ]
            }
          }
        }
      ]
    }
  ],

  "bulkActions": [
    {
      "label": "Export Selected",
      "type":  "button",
      "actionType": "ajax",
      "api": {
        "method": "post",
        "url": "/api/v1/purchase-orders/export",
        "data": { "ids": "${ids}" }
      }
    }
  ]
}
```

### 6.3 `form`

Multi-field form with validation, conditional fields, and API submit.

```json
{
  "type":   "form",
  "title":  "New Purchase Order",
  "api":    "post:/api/v1/purchase-orders",
  "initApi": "get:/api/v1/purchase-orders/${id}",
  "redirect": "/purchasing/orders/${id}",
  "body": [
    {
      "type":     "select",
      "name":     "supplier_id",
      "label":    "Supplier",
      "required": true,
      "source":   "get:/api/v1/suppliers?fields=id,name",
      "valueField": "id",
      "labelField": "name",
      "searchable": true
    },
    {
      "type":  "input-date",
      "name":  "expected_date",
      "label": "Expected Delivery",
      "required": true,
      "minDate": "${DATETOSTR(NOW(), 'YYYY-MM-DD')}"
    },
    {
      "type":  "select",
      "name":  "currency",
      "label": "Currency",
      "value": "KES",
      "options": [
        { "label": "KES — Kenyan Shilling", "value": "KES" },
        { "label": "USD — US Dollar",       "value": "USD" },
        { "label": "EUR — Euro",            "value": "EUR" }
      ]
    },
    {
      "type":  "textarea",
      "name":  "notes",
      "label": "Notes"
    },
    {
      "type":  "combo",
      "name":  "line_items",
      "label": "Line Items",
      "required": true,
      "multiple": true,
      "items": [
        {
          "type":  "select",
          "name":  "product_id",
          "label": "Product",
          "source": "get:/api/v1/catalog/products",
          "valueField": "id",
          "labelField": "name",
          "searchable": true,
          "required": true
        },
        {
          "type":  "input-number",
          "name":  "quantity",
          "label": "Qty",
          "required": true,
          "min": 1
        },
        {
          "type":  "input-number",
          "name":  "unit_price",
          "label": "Unit Price",
          "required": true,
          "min": 0
        }
      ]
    }
  ],
  "actions": [
    { "type": "submit", "label": "Save Draft",   "level": "default" },
    { "type": "button", "label": "Cancel",
      "actionType": "link", "link": "/purchasing/orders" }
  ]
}
```

### 6.4 `wizard`

Multi-step form. Natural for payroll runs, period-end close, onboarding flows.

```json
{
  "type":  "wizard",
  "api":   "post:/api/v1/payroll/runs",
  "steps": [
    {
      "title": "Select Period",
      "body": [
        { "type": "select",     "name": "period_id",  "label": "Pay Period",
          "source": "get:/api/v1/payroll/periods?status=open",
          "required": true },
        { "type": "input-date", "name": "pay_date",   "label": "Payment Date",
          "required": true }
      ]
    },
    {
      "title": "Review Employees",
      "initApi": "get:/api/v1/payroll/preview?period_id=${period_id}",
      "body": [
        {
          "type": "table",
          "source": "${employees}",
          "columns": [
            { "name": "name",            "label": "Employee" },
            { "name": "gross",           "label": "Gross Pay",   "type": "number" },
            { "name": "deductions",      "label": "Deductions",  "type": "number" },
            { "name": "net_pay",         "label": "Net Pay",     "type": "number" }
          ]
        }
      ]
    },
    {
      "title": "Confirm & Submit",
      "body": [
        { "type": "tpl",
          "tpl": "You are about to process payroll for <strong>${employee_count}</strong> employees totalling <strong>KES ${ROUND(total_net, 2)}</strong>." }
      ]
    }
  ]
}
```

### 6.5 `dialog` and `drawer`

Inline detail/edit panels without full page navigation.

```json
{
  "type":  "button",
  "label": "Edit",
  "level": "link",
  "actionType": "dialog",
  "dialog": {
    "title": "Edit Purchase Order #${reference}",
    "size":  "lg",
    "body": {
      "type":    "form",
      "api":     "put:/api/v1/purchase-orders/${id}",
      "initApi": "get:/api/v1/purchase-orders/${id}",
      "body": [
        { "type": "input-text", "name": "notes", "label": "Notes" }
      ]
    },
    "actions": [
      { "type": "submit", "label": "Save",   "level": "primary" },
      { "type": "button", "label": "Cancel", "actionType": "close" }
    ]
  }
}
```

`drawer` is the same but slides in from the side. Prefer `drawer` for detail
views; `dialog` for short confirmations and quick edits.

### 6.6 `tabs`

Split complex pages into tabbed sections. Used for record detail pages
(Overview / Line Items / Activity / Documents).

```json
{
  "type": "tabs",
  "tabs": [
    {
      "title": "Overview",
      "body": { "type": "descriptions", "... ": "..." }
    },
    {
      "title": "Line Items",
      "body": {
        "type": "crud",
        "api": "get:/api/v1/purchase-orders/${id}/lines",
        "... ": "..."
      }
    },
    {
      "title": "Approvals",
      "body": {
        "type": "service",
        "api": "get:/api/v1/purchase-orders/${id}/approvals",
        "body": {
          "type": "timeline",
          "source": "${steps}",
          "items": { "... ": "..." }
        }
      }
    }
  ]
}
```

### 6.7 `service`

Fetches data from an API and makes it available to child components. Also
loads dynamic sub-schemas from `schemaApi`.

```json
{
  "type": "service",
  "api":  "get:/api/v1/accounting/dashboard/summary",
  "body": [
    {
      "type": "stat",
      "source": "${total_receivables}",
      "label": "Total Receivables",
      "prefix": "KES "
    },
    {
      "type": "stat",
      "source": "${overdue_count}",
      "label": "Overdue Invoices"
    }
  ]
}
```

### 6.8 `descriptions` (Detail View)

Renders a record's fields in a label/value layout. Used for read-only detail
pages.

```json
{
  "type": "descriptions",
  "title": "Purchase Order Details",
  "source": "get:/api/v1/purchase-orders/${id}",
  "items": [
    { "name": "reference",     "label": "Reference" },
    { "name": "supplier_name", "label": "Supplier" },
    { "name": "total",         "label": "Total",    "type": "number" },
    { "name": "status",        "label": "Status" },
    { "name": "created_at",    "label": "Created",  "type": "date" }
  ]
}
```

### 6.9 `timeline`

Used for approval history, activity logs, audit trails.

```json
{
  "type":   "timeline",
  "source": "${history}",
  "items": {
    "time":  "${occurred_at}",
    "title": "${action}",
    "detail": "${actor_name}: ${comment}"
  }
}
```

### 6.10 `chart`

Wraps ECharts. Used in dashboards and analytics.

```json
{
  "type": "chart",
  "api":  "get:/api/v1/analytics/revenue/monthly",
  "config": {
    "xAxis": { "type": "category", "data": "${months}" },
    "yAxis": { "type": "value" },
    "series": [
      {
        "name": "Revenue",
        "type": "bar",
        "data": "${totals}"
      }
    ]
  }
}
```

### 6.11 `input-kv` and `combo`

`combo` handles dynamic lists of structured objects (line items, addresses).
`input-kv` handles key-value pairs (metadata, custom attributes).

```json
{
  "type":     "combo",
  "name":     "addresses",
  "label":    "Delivery Addresses",
  "multiple": true,
  "multiLine": true,
  "addBtn":    { "label": "Add Address" },
  "items": [
    { "type": "input-text",   "name": "street",   "label": "Street"  },
    { "type": "input-text",   "name": "city",     "label": "City"    },
    { "type": "select",       "name": "country",  "label": "Country",
      "source": "get:/api/v1/countries" }
  ]
}
```

### 6.12 `condition-builder`

A built-in visual condition editor. This is directly relevant to Awo's
business rules UI.

```json
{
  "type":  "condition-builder",
  "name":  "conditions",
  "label": "Conditions",
  "fields": [
    {
      "label":    "PO Total",
      "name":     "po.total",
      "type":     "number",
      "operators": ["equal", "not_equal", "less", "less_or_equal",
                    "greater", "greater_or_equal", "between", "not_between"]
    },
    {
      "label":    "Supplier Tier",
      "name":     "po.supplier_tier",
      "type":     "select",
      "operators": ["equal", "not_equal", "in", "not_in"],
      "options": [
        { "label": "Preferred", "value": "preferred" },
        { "label": "Approved",  "value": "approved"  },
        { "label": "New",       "value": "new"       }
      ]
    }
  ]
}
```

The output JSON from this component maps directly onto `pkg/condition`'s
`ConditionGroup` structure. This eliminates the need to build a bespoke
condition-builder UI.

---

## 7. Go Schema Builder — `awo/web/schema`

Rather than writing raw JSON in Go string literals or maintaining JSON files,
all amis schemas are constructed using a typed Go builder package. This gives
you Go's type checking on schema definitions, makes schemas testable, and
co-locates schema logic with the module it describes.

### 7.1 Package Structure

```
awo/web/schema/
├── types.go         — All amis JSON types as Go structs
├── builder.go       — Fluent builder methods
├── crud.go          — CRUD page builder helpers
├── form.go          — Form builder helpers
├── page.go          — Page / toolbar helpers
├── app.go           — App shell builder
└── accounting/
│   ├── journal.go   — journal page schema
│   └── accounts.go  — accounts page schema
├── purchasing/
│   └── orders.go    — purchase orders page schema
├── settings/
│   └── rules.go     — business rules builder page schema
└── ...
```

### 7.2 Core Types

```go
// awo/web/schema/types.go

package schema

// Schema is the root interface. Any amis component is a Schema.
type Schema map[string]any

// Page is the top-level page container.
type Page struct {
    Type     string `json:"type"`
    Title    string `json:"title,omitempty"`
    SubTitle string `json:"subTitle,omitempty"`
    Toolbar  []any  `json:"toolbar,omitempty"`
    Body     any    `json:"body"`
}

// CRUDSchema is the primary list component.
type CRUDSchema struct {
    Type           string     `json:"type"`
    ID             string     `json:"id,omitempty"`
    API            string     `json:"api"`
    SyncLocation   bool       `json:"syncLocation"`
    Filter         *Form      `json:"filter,omitempty"`
    HeaderToolbar  []any      `json:"headerToolbar,omitempty"`
    Columns        []Column   `json:"columns"`
    BulkActions    []any      `json:"bulkActions,omitempty"`
    FooterToolbar  []any      `json:"footerToolbar,omitempty"`
}

type Column struct {
    Name       string `json:"name"`
    Label      string `json:"label"`
    Type       string `json:"type,omitempty"`
    Sortable   bool   `json:"sortable,omitempty"`
    Toggled    *bool  `json:"toggled,omitempty"`
    Prefix     string `json:"prefix,omitempty"`
    Suffix     string `json:"suffix,omitempty"`
    ColorMap   any    `json:"colorMap,omitempty"`
    Buttons    []any  `json:"buttons,omitempty"`
    VisibleOn  string `json:"visibleOn,omitempty"`
}

type Form struct {
    Type     string `json:"type"`
    Title    string `json:"title,omitempty"`
    API      string `json:"api,omitempty"`
    InitAPI  string `json:"initApi,omitempty"`
    Redirect string `json:"redirect,omitempty"`
    Body     []any  `json:"body"`
    Actions  []any  `json:"actions,omitempty"`
    Mode     string `json:"mode,omitempty"` // "horizontal" | "normal" | "inline"
}

type Button struct {
    Type        string `json:"type"`
    Label       string `json:"label"`
    Level       string `json:"level,omitempty"` // "primary"|"default"|"link"|"danger"
    ActionType  string `json:"actionType,omitempty"`
    Link        string `json:"link,omitempty"`
    API         string `json:"api,omitempty"`
    ConfirmText string `json:"confirmText,omitempty"`
    VisibleOn   string `json:"visibleOn,omitempty"`
    DisabledOn  string `json:"disabledOn,omitempty"`
    OnEvent     any    `json:"onEvent,omitempty"`
    Dialog      any    `json:"dialog,omitempty"`
    Drawer      any    `json:"drawer,omitempty"`
}
```

### 7.3 Page Schemas Per Module

Each module's handler calls its schema function to get the amis schema, then
marshals it to JSON.

```go
// awo/web/schema/purchasing/orders.go

package purchasing

import (
    "github.com/mustafe/awo/web/schema"
    "github.com/mustafe/awo/platform/flags"
)

// OrdersListPage returns the amis schema for the PO list page.
// The flags map is used to conditionally include columns or actions.
func OrdersListPage(fl map[string]bool) schema.Page {
    return schema.Page{
        Type:     "page",
        Title:    "Purchase Orders",
        SubTitle: "Manage supplier purchase orders",
        Toolbar: []any{
            schema.Button{
                Type:  "button", Label: "New Order",
                Level: "primary", ActionType: "link",
                Link:  "/purchasing/orders/new",
            },
        },
        Body: schema.CRUDSchema{
            Type:         "crud",
            ID:           "po-list",
            API:          "get:/api/v1/purchase-orders",
            SyncLocation: true,
            Filter: &schema.Form{
                Type: "form",
                Body: orderFilterFields(),
            },
            HeaderToolbar: []any{"search-box", "bulkActions",
                map[string]string{"type": "columns-toggler"},
                map[string]string{"type": "reload"},
            },
            Columns: orderColumns(fl),
        },
    }
}

func orderColumns(fl map[string]bool) []schema.Column {
    cols := []schema.Column{
        {Name: "reference",     Label: "Reference",  Sortable: true},
        {Name: "supplier_name", Label: "Supplier",   Sortable: true},
        {Name: "total",         Label: "Total",      Type: "number",
         Prefix: "KES ", Sortable: true},
        {Name: "status",        Label: "Status",     Type: "tag",
         ColorMap: orderStatusColorMap()},
        {Name: "created_at",    Label: "Date",       Type: "date", Sortable: true},
        orderActionsColumn(fl),
    }
    return cols
}

func orderActionsColumn(fl map[string]bool) schema.Column {
    buttons := []any{
        schema.Button{Type: "button", Label: "View", Level: "link",
            ActionType: "link", Link: "/purchasing/orders/${id}"},
        schema.Button{Type: "button", Label: "Submit", Level: "link",
            VisibleOn:   "${status === 'draft'}",
            ActionType:  "ajax",
            API:         "post:/api/v1/purchase-orders/${id}/submit",
            ConfirmText: "Submit this order for approval?",
        },
    }

    // Only show MRP-linked indicator if manufacturing flag is on
    if fl["manufacturing.mrp_enabled"] {
        buttons = append(buttons, schema.Button{
            Type: "button", Label: "MRP Link", Level: "link",
            VisibleOn: "${mrp_order_id !== null}",
            ActionType: "link",
            Link: "/production/mrp/${mrp_order_id}",
        })
    }

    return schema.Column{Type: "operation", Label: "Actions",
        Buttons: buttons}
}

func orderStatusColorMap() map[string]string {
    return map[string]string{
        "draft":     "default",
        "submitted": "processing",
        "confirmed": "success",
        "cancelled": "error",
    }
}
```

### 7.4 Schema Handler

```go
// awo/web/handlers/schema/schema_handler.go

package schemahandlers

import (
    "github.com/gofiber/fiber/v2"
    "github.com/mustafe/awo/web/middleware"
    purchasingschema "github.com/mustafe/awo/web/schema/purchasing"
)

type SchemaHandlers struct{}

func (h *SchemaHandlers) PurchasingOrders(c *fiber.Ctx) error {
    fl := middleware.ContextFlags(c)
    page := purchasingschema.OrdersListPage(fl)
    return c.JSON(page)
}
```

---

## 8. Fiber Route Structure

### 8.1 Route Registration

```go
// awo/web/router/routes.go

func registerSchemaRoutes(app *fiber.App, s *Services) {
    schema := app.Group("/schema",
        middleware.ResolveTenant,
        middleware.Authenticate,
        middleware.InjectFlags,
    )

    // App shell — returns the full nav structure based on flags
    schema.Get("/app", s.SchemaHandlers.AppShell)

    // Module schemas
    schema.Get("/accounting/journal",      s.SchemaHandlers.AccountingJournal)
    schema.Get("/accounting/accounts",     s.SchemaHandlers.AccountingAccounts)
    schema.Get("/purchasing/orders",       s.SchemaHandlers.PurchasingOrders)
    schema.Get("/purchasing/orders/new",   s.SchemaHandlers.PurchasingOrdersNew)
    schema.Get("/purchasing/orders/:id",   s.SchemaHandlers.PurchasingOrderDetail)
    schema.Get("/people/employees",        s.SchemaHandlers.PeopleEmployees)
    schema.Get("/settings/rules/:module",  s.SchemaHandlers.SettingsRules)
    // ... all modules
}

func registerAPIRoutes(app *fiber.App, s *Services) {
    api := app.Group("/api/v1",
        middleware.ResolveTenant,
        middleware.Authenticate,
        middleware.InjectFlags,
        middleware.AuditWrap,
    )

    po := api.Group("/purchase-orders",
        middleware.RequirePermission("read", "purchase_orders", s.Access))
    po.Get("/",          s.POHandlers.List)
    po.Post("/",         middleware.RequirePermission("create", "purchase_orders", s.Access),
                         s.POHandlers.Create)
    po.Get("/:id",       s.POHandlers.Get)
    po.Put("/:id",       middleware.RequirePermission("update", "purchase_orders", s.Access),
                         s.POHandlers.Update)
    po.Post("/:id/submit",   s.POHandlers.Submit)
    po.Post("/:id/cancel",   s.POHandlers.Cancel)
    po.Get("/:id/lines",     s.POHandlers.ListLines)
    po.Get("/:id/approvals", s.POHandlers.ListApprovals)
    // ... all resources
}
```

### 8.2 App Shell Handler

The app shell returns the `app` schema with nav items filtered by flags. This
is the first call amis makes after booting.

```go
// awo/web/handlers/schema/app_shell.go

func (h *SchemaHandlers) AppShell(c *fiber.Ctx) error {
    fl   := middleware.ContextFlags(c)
    user := middleware.ContextUser(c)

    pages := []any{
        navSection("Accounting", "fa fa-calculator", []navItem{
            {Label: "Journal",     URL: "/accounting/journal"},
            {Label: "Chart of Accounts", URL: "/accounting/accounts"},
            {Label: "Invoices",    URL: "/accounting/invoices"},
            {Label: "Bills",       URL: "/accounting/bills"},
        }),
    }

    if fl["payroll.module"] {
        pages = append(pages, navSection("People & HR", "fa fa-users", []navItem{
            {Label: "Employees", URL: "/people/employees"},
            {Label: "Payroll",   URL: "/people/payroll"},
            {Label: "Leave",     URL: "/people/leave"},
        }))
    }

    pages = append(pages, navSection("Purchasing", "fa fa-shopping-cart", []navItem{
        {Label: "Purchase Orders", URL: "/purchasing/orders"},
        {Label: "Suppliers",       URL: "/purchasing/suppliers"},
        {Label: "Goods Receipts",  URL: "/purchasing/receipts"},
    }))

    if fl["manufacturing.module"] {
        pages = append(pages, navSection("Production", "fa fa-industry", []navItem{
            {Label: "Manufacturing Orders", URL: "/production/orders"},
            {Label: "Bills of Materials",   URL: "/production/bom"},
        }))
    }

    // Settings always visible
    pages = append(pages, settingsNav(fl))

    app := map[string]any{
        "type":      "app",
        "brandName": "Awo ERP",
        "logo":      "/static/logo.svg",
        "header": map[string]any{
            "type": "tpl",
            "tpl":  user.TenantName,
        },
        "pages": pages,
    }

    return c.JSON(app)
}
```

---

## 9. Authentication & Permission Gating

### 9.1 Session Cookie Flow

The amis `fetcher` function sends cookies automatically (same-origin). Your
Fiber `Authenticate` middleware reads the session cookie on every `/schema/`
and `/api/v1/` request.

```go
// awo/web/middleware/authenticate.go — unchanged from main architecture
func Authenticate(c *fiber.Ctx) error {
    sessionToken := c.Cookies("awo_session")
    if sessionToken == "" {
        // Return 401 as amis response envelope so amis handles it
        return c.Status(401).JSON(response.Err("Session expired. Please log in."))
    }
    session, err := authSvc.ValidateSession(c.Context(), sessionToken)
    if err != nil {
        return c.Status(401).JSON(response.Err("Invalid session"))
    }
    setContextSession(c, session)
    return c.Next()
}
```

When amis receives a non-zero `status` from any API call, it shows the `msg`
as a toast. For 401 specifically, configure the amis `env` to redirect to the
login page:

```javascript
// index.html env configuration
{
  fetcher({ url, method, data, headers }) {
    return fetch(url, { ... }).then(async r => {
      const json = await r.json();
      if (r.status === 401) {
        window.location.href = '/login';
        return;
      }
      return json;
    });
  }
}
```

### 9.2 Permission-Based Column/Action Hiding

Pass permissions into the schema as data so amis can conditionally render
actions:

```go
// awo/web/handlers/schema/schema_handler.go

func (h *SchemaHandlers) PurchasingOrders(c *fiber.Ctx) error {
    fl   := middleware.ContextFlags(c)
    user := middleware.ContextUser(c)

    canCreate := h.access.Can(c.Context(), user.ID, "create", "purchase_orders")
    canApprove := h.access.Can(c.Context(), user.ID, "approve", "purchase_orders")

    page := purchasingschema.OrdersListPage(fl, purchasingschema.Permissions{
        CanCreate: canCreate,
        CanApprove: canApprove,
    })
    return c.JSON(page)
}
```

Inside the schema builder, use `visibleOn` with injected data:

```go
// In the schema, inject permissions as page-level data
page := schema.Page{
    Type:  "page",
    // ...
    Data: map[string]any{
        "can_create":  perms.CanCreate,
        "can_approve": perms.CanApprove,
    },
    Toolbar: []any{
        schema.Button{
            Label: "New Order", Level: "primary",
            ActionType: "link", Link: "/purchasing/orders/new",
            VisibleOn: "${can_create}",
        },
    },
}
```

---

## 10. Flag-Gated Pages

If a user navigates to a schema endpoint for a disabled module, the Go handler
returns an empty page with a message:

```go
// awo/web/handlers/schema/schema_handler.go

func (h *SchemaHandlers) PeoplePayroll(c *fiber.Ctx) error {
    fl := middleware.ContextFlags(c)

    if !fl["payroll.module"] {
        return c.JSON(map[string]any{
            "type":  "page",
            "title": "Payroll",
            "body": map[string]any{
                "type": "alert",
                "level": "info",
                "body": "The Payroll module is not enabled for your organisation. Contact your administrator to activate it.",
            },
        })
    }

    return c.JSON(payrollschema.RunsPage(fl))
}
```

Nav items for disabled modules are simply omitted from the app shell (see
section 8.2), so users do not see them in the sidebar at all. The schema-level
check is a defence-in-depth measure for direct URL access.

---

## 11. Awo Page Patterns

### 11.1 Standard CRUD List Page

Pattern for every module's primary list view (invoices, purchase orders,
employees, etc.):

```
page
  └── crud
        ├── filter (form)
        ├── headerToolbar [search-box, bulkActions, columns-toggler, reload]
        ├── columns [..., operation column with buttons]
        └── footerToolbar [statistics, pagination]
```

### 11.2 Record Detail Page

```
page
  ├── toolbar [action buttons based on status]
  └── tabs
        ├── "Overview"    → descriptions
        ├── "Line Items"  → crud (nested, no filter)
        ├── "Documents"   → crud (attachments)
        ├── "Approvals"   → service + timeline
        └── "Activity"    → service + timeline (audit log)
```

### 11.3 Settings Module Page

```
page
  └── tabs
        ├── "General"     → form (tenant config keys)
        ├── "Rules"       → crud (rule sets) + nested wizard (rule editor)
        └── "Integrations" → crud (webhooks/API keys)
```

### 11.4 Approval Inbox Page

The pending approvals page is a specific pattern used by any user with approval
responsibilities across modules:

```json
{
  "type": "page",
  "title": "Approvals Inbox",
  "body": {
    "type": "crud",
    "id":   "approvals-inbox",
    "api":  "get:/api/v1/approvals/pending",
    "columns": [
      { "name": "record_type",  "label": "Type" },
      { "name": "record_ref",   "label": "Reference" },
      { "name": "requested_by", "label": "Requested By" },
      { "name": "requested_at", "label": "Date",        "type": "date" },
      {
        "type": "operation", "label": "Actions",
        "buttons": [
          {
            "label": "View",  "type": "button", "level": "link",
            "actionType": "drawer",
            "drawer": {
              "title": "Approve ${record_type} #${record_ref}",
              "body": {
                "type": "service",
                "schemaApi": "get:/schema/approvals/${record_type}/${record_id}"
              }
            }
          },
          {
            "label": "Approve", "type": "button", "level": "success",
            "actionType": "ajax",
            "api": {
              "method": "post",
              "url": "/api/v1/approvals/${request_id}/approve",
              "data": { "comment": "${comment}" }
            },
            "dialog": {
              "title": "Approve",
              "body": [
                { "type": "textarea", "name": "comment", "label": "Comment (optional)" }
              ]
            },
            "onEvent": {
              "success": {
                "actions": [
                  { "actionType": "reload",    "componentId": "approvals-inbox" },
                  { "actionType": "toast",     "args": { "msg": "Approved" } }
                ]
              }
            }
          },
          {
            "label": "Reject", "type": "button", "level": "danger",
            "actionType": "ajax",
            "api": "post:/api/v1/approvals/${request_id}/reject",
            "confirmText": "Reject this request?",
            "onEvent": {
              "success": {
                "actions": [
                  { "actionType": "reload", "componentId": "approvals-inbox" }
                ]
              }
            }
          }
        ]
      }
    ]
  }
}
```

### 11.5 Dashboard Page

```json
{
  "type": "page",
  "title": "Accounting Dashboard",
  "body": {
    "type": "grid",
    "columns": [
      {
        "columnClassName": "col-sm-4",
        "body": {
          "type": "service",
          "api": "get:/api/v1/analytics/receivables/summary",
          "body": [
            { "type": "stat", "source": "${total}",    "label": "Total Receivables", "prefix": "KES " },
            { "type": "stat", "source": "${overdue}",  "label": "Overdue",           "prefix": "KES ", "className": "text-danger" }
          ]
        }
      },
      {
        "columnClassName": "col-sm-8",
        "body": {
          "type": "chart",
          "api": "get:/api/v1/analytics/revenue/monthly",
          "config": {
            "xAxis": { "type": "category", "data": "${months}" },
            "yAxis": { "type": "value" },
            "series": [{ "type": "bar", "data": "${values}", "name": "Revenue" }]
          }
        }
      }
    ]
  }
}
```

---

## 12. The Rules Builder UI in amis

The business rules builder (Settings → Rules) is the most complex UI in Awo.
amis's built-in `condition-builder` component handles the hardest part.

### 12.1 Rules List Page

```go
// awo/web/schema/settings/rules.go

func RulesPage(triggerEvent string, fields []FieldDef) map[string]any {
    return map[string]any{
        "type":  "page",
        "title": "Business Rules — " + triggerEventLabel(triggerEvent),
        "toolbar": []any{
            map[string]any{
                "type": "button", "label": "New Rule Set",
                "level": "primary", "actionType": "dialog",
                "dialog": NewRuleSetDialog(triggerEvent),
            },
        },
        "body": map[string]any{
            "type":         "crud",
            "id":           "rules-list",
            "api":          "get:/api/v1/settings/rules?trigger=" + triggerEvent,
            "draggable":    true,
            "saveOrderApi": "post:/api/v1/settings/rules/reorder",
            "columns": []any{
                map[string]string{"name": "name",    "label": "Rule Set Name"},
                map[string]string{"name": "priority", "label": "Priority"},
                map[string]any{
                    "name": "active", "label": "Active",
                    "type": "switch",
                    "onEvent": map[string]any{
                        "change": map[string]any{
                            "actions": []any{
                                map[string]any{
                                    "actionType": "ajax",
                                    "api": map[string]string{
                                        "method": "post",
                                        "url":    "/api/v1/settings/rules/${id}/toggle",
                                    },
                                },
                            },
                        },
                    },
                },
                map[string]any{
                    "type": "operation", "label": "Actions",
                    "buttons": []any{
                        editRuleSetButton(fields),
                        testRuleSetButton(fields),
                        map[string]any{
                            "label": "History", "type": "button", "level": "link",
                            "actionType": "drawer",
                            "drawer": ExecutionHistoryDrawer(),
                        },
                    },
                },
            },
        },
    }
}
```

### 12.2 Rule Editor with `condition-builder`

```go
func editRuleSetButton(fields []FieldDef) map[string]any {
    return map[string]any{
        "label": "Edit", "type": "button", "level": "link",
        "actionType": "drawer",
        "drawer": map[string]any{
            "title": "Edit Rule Set: ${name}",
            "size":  "xl",
            "body": map[string]any{
                "type":    "form",
                "api":     "put:/api/v1/settings/rules/${id}",
                "initApi": "get:/api/v1/settings/rules/${id}",
                "body": []any{
                    map[string]any{
                        "type": "input-text", "name": "name", "label": "Rule Set Name", "required": true,
                    },
                    map[string]any{
                        "type":  "crud",
                        "label": "Rules",
                        "name":  "rules",
                        "mode":  "table",
                        "columns": []any{
                            map[string]string{"name": "name", "label": "Rule Name"},
                            map[string]any{
                                "name": "condition_group", "label": "Conditions",
                                "type": "condition-builder",
                                "fields": amisFieldDefs(fields),
                            },
                            map[string]any{
                                "name":  "actions",
                                "label": "Actions",
                                "type":  "combo",
                                "multiple": true,
                                "items": ruleActionItems(),
                            },
                            map[string]any{
                                "name": "stop_on_match", "label": "Stop on Match",
                                "type": "switch",
                            },
                        },
                    },
                },
            },
        },
    }
}

// amisFieldDefs converts platform/rules FactField slice to amis condition-builder field format
func amisFieldDefs(fields []FieldDef) []map[string]any {
    result := make([]map[string]any, len(fields))
    for i, f := range fields {
        def := map[string]any{
            "name":  f.Key,
            "label": f.Label,
            "type":  amisFieldType(f.Type),
        }
        if len(f.Options) > 0 {
            def["options"] = f.Options
        }
        if len(f.Operators) > 0 {
            def["operators"] = f.Operators
        }
        result[i] = def
    }
    return result
}
```

### 12.3 Test Rule Panel

```go
func testRuleSetButton(fields []FieldDef) map[string]any {
    return map[string]any{
        "label": "Test", "type": "button", "level": "link",
        "actionType": "dialog",
        "dialog": map[string]any{
            "title": "Test Rule Set: ${name}",
            "body": map[string]any{
                "type": "form",
                "api": map[string]any{
                    "method": "post",
                    "url":    "/api/v1/settings/rules/${id}/test",
                },
                "body":    testSampleInputs(fields),
                "actions": []any{
                    map[string]any{"type": "submit", "label": "Run Test", "level": "primary"},
                },
                "onEvent": map[string]any{
                    "submitSucc": map[string]any{
                        "actions": []any{
                            map[string]any{
                                "actionType": "dialog",
                                "dialog": map[string]any{
                                    "title": "Test Result",
                                    "body": []any{
                                        map[string]any{
                                            "type":   "tpl",
                                            "tpl":    "Decision: <strong>${data.decision}</strong>",
                                        },
                                        map[string]any{
                                            "type":   "json",
                                            "source": "${data.matched_rules}",
                                        },
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
    }
}
```

---

## 13. Theming

amis has a built-in theme system. You can customise colours and spacing via CSS
variables without touching any component code.

### 13.1 Built-in Themes

amis ships with `default` and `antd` (Ant Design) themes. Switch via the
`theme` option in `embed()`:

```javascript
embed('#root', schema, {}, { theme: 'antd', ...env });
```

### 13.2 Custom CSS Variables

Override amis's CSS variables for brand colours:

```css
/* awo/web/static/awo-theme.css */
:root {
  --colors-brand-5:  #0066cc;   /* primary action colour */
  --colors-brand-6:  #0052a3;   /* hover state */
  --colors-success-5: #00a854;  /* success / confirmed */
  --colors-warning-5: #faad14;  /* pending / processing */
  --colors-danger-5:  #f5222d;  /* error / cancelled */
}
```

### 13.3 Per-Tenant Theme

Store the tenant's primary colour in `platform/config`. Return it in the app
shell data domain. Reference it as a CSS variable override in the shell:

```go
// In AppShell handler
tenantColor, _ := h.config.GetTenant(ctx, tenantID, "ui.brand_color")

return c.JSON(map[string]any{
    "type": "app",
    // ...
    "data": map[string]any{
        "brand_color": tenantColor.StringOr("#0066cc"),
    },
    // Inject a style block into the page head
    "style": "body { --colors-brand-5: ${brand_color}; }",
})
```

---

## 14. Known Limitations & Workarounds

### 14.1 Documentation is Primarily Chinese

The authoritative documentation lives at `aisuda.bce.baidu.com/amis/zh-CN`.
The English version (`/en-US`) exists but lags the Chinese version and is
incomplete for advanced topics. Strategies:

- Use a browser translation extension when reading Chinese docs
- The GitHub source code is the most reliable reference — component prop types
  are in TypeScript, fully readable without language knowledge
- The online playground at `aisuda.bce.baidu.com/amis/zh-CN/examples/index`
  lets you test JSON schemas interactively

### 14.2 `condition-builder` Output Format

The `condition-builder` output JSON is amis's own format, not identical to
`pkg/condition`'s `ConditionGroup`. Write a converter in Go when saving rules:

```go
// awo/web/converters/conditions.go

// AmisConditionToPkg converts amis condition-builder JSON output
// to pkg/condition ConditionGroup for storage.
func AmisConditionToPkg(amisCond map[string]any) (*condition.ConditionGroup, error) {
    // amis uses "conjunction" as the group key (matches our format)
    // amis uses "left", "op", "right" (matches ConditionRule)
    // The structures align closely; usually only field name normalisation is needed
    data, err := json.Marshal(amisCond)
    if err != nil {
        return nil, err
    }
    var group condition.ConditionGroup
    return &group, json.Unmarshal(data, &group)
}
```

If the field name formats diverge, do the mapping in this converter rather than
changing either side.

### 14.3 No Built-in SSE / WebSocket

amis polls APIs on a timer for "live" updates. For real-time features (payroll
run progress, import status), use the `status` component with a polling
interval or implement a lightweight SSE handler that the amis `service`
component polls:

```json
{
  "type":       "service",
  "api":        "get:/api/v1/payroll/runs/${id}/status",
  "interval":   3000,
  "silentPolling": true,
  "body": {
    "type":   "progress",
    "source": "${progress_pct}",
    "label":  "${status_message}"
  }
}
```

### 14.4 Large Form Performance

For forms with 50+ fields (e.g., full employee profile), split into a `tabs`
layout with one form per tab. Each tab's form loads its own `initApi`. This
avoids rendering all fields at once and keeps interactions snappy.

### 14.5 Custom Renderers

If you need a component that amis does not provide (e.g., a map picker, a
signature pad, a barcode scanner), write a custom renderer:

```typescript
// src/renderers/MapPicker.tsx
import { registerRenderer, RendererProps } from 'amis';

interface MapPickerProps extends RendererProps {
  name: string;
  value?: { lat: number; lng: number };
}

const MapPickerRenderer: React.FC<MapPickerProps> = ({ value, onChange }) => {
  // ... your component
  return <div>...</div>;
};

registerRenderer({
  test: /\bmap-picker\b/,
  component: MapPickerRenderer,
});
```

Then use `{ "type": "map-picker", "name": "location" }` in any schema.

---

## 15. Documentation Sources

| Resource | URL | Notes |
|---|---|---|
| Official docs (Chinese) | `aisuda.bce.baidu.com/amis/zh-CN` | Most complete |
| Official docs (English) | `aisuda.bce.baidu.com/amis/en-US` | Partial |
| GitHub repository | `github.com/baidu/amis` | TypeScript source is the ground truth |
| Online playground | `aisuda.bce.baidu.com/amis/zh-CN/examples/index` | Test schemas interactively |
| npm package | `npmjs.com/package/amis` | `amis` (core + SDK) |
| Visual editor | `aisuda.bce.baidu.com/amis-editor-demo` | Drag-and-drop schema builder |

**Reading Chinese docs without knowing Chinese:**

The JSON schema examples in the docs are universally readable — JSON needs no
translation. For property descriptions, browser auto-translate is sufficient.
The component catalog (all `type` values and their props) is the section you
will reference most; it is fully navigable via the sidebar even untranslated.

---

*End of Awo ERP × amis UI Integration Guide*
