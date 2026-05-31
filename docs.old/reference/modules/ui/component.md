# AWO ERP — amis UI Component Documentation

> **Framework:** [Baidu amis](https://aisuda.bce.baidu.com/amis/en-US/docs/index) v3.x  
> **Rendering Mode:** JSON-driven schema (no JSX required for most components)  
> **Scope:** Navigation, Tabs, DataTables with Filters, Breadcrumbs, Reporting Drill-Down, Workflow Configuration  
> **Design System:** amis default theme (`cxd`) — light mode with KES-locale number formatting

---

## Table of Contents

1. [amis Architecture Overview](#1-amis-architecture-overview)
2. [Theming & Global Defaults](#2-theming--global-defaults)
3. [Navigation (Nav)](#3-navigation-nav)
   - [3.1 Left Sidebar Navigation](#31-left-sidebar-navigation)
   - [3.2 Collapsible Groups](#32-collapsible-groups)
   - [3.3 Role-Based Menu Filtering](#33-role-based-menu-filtering)
   - [3.4 Nav with Badge Counts](#34-nav-with-badge-counts)
4. [Breadcrumbs](#4-breadcrumbs)
   - [4.1 Static Breadcrumb](#41-static-breadcrumb)
   - [4.2 Dynamic Breadcrumb (Context-Driven)](#42-dynamic-breadcrumb-context-driven)
   - [4.3 Breadcrumb + Page Title Row](#43-breadcrumb--page-title-row)
5. [Tabs](#5-tabs)
   - [5.1 Standard Horizontal Tabs](#51-standard-horizontal-tabs)
   - [5.2 Tabs with Lazy Loading](#52-tabs-with-lazy-loading)
   - [5.3 Vertical Tabs (Detail Panels)](#53-vertical-tabs-detail-panels)
   - [5.4 Tab-Level Toolbar Actions](#54-tab-level-toolbar-actions)
6. [DataTables with Filters](#6-datatables-with-filters)
   - [6.1 Basic CRUD Table](#61-basic-crud-table)
   - [6.2 Inline Filter Toolbar](#62-inline-filter-toolbar)
   - [6.3 Collapsible Filter Panel (Query Form)](#63-collapsible-filter-panel-query-form)
   - [6.4 Column-Level Features](#64-column-level-features)
   - [6.5 Server-Side Pagination & Sorting](#65-server-side-pagination--sorting)
   - [6.6 Bulk Actions & Row Selection](#66-bulk-actions--row-selection)
   - [6.7 Expandable Row Detail](#67-expandable-row-detail)
   - [6.8 Fixed Columns & Frozen Headers](#68-fixed-columns--frozen-headers)
7. [Reporting Drill-Down](#7-reporting-drill-down)
   - [7.1 Architecture Overview](#71-architecture-overview)
   - [7.2 P&L Report Page Schema](#72-pl-report-page-schema)
   - [7.3 Drill-Down Modal — Transaction Detail](#73-drill-down-modal--transaction-detail)
   - [7.4 Drill-Down Breadcrumb Trail](#74-drill-down-breadcrumb-trail)
   - [7.5 Multi-Level Drill (Section → Account → Journal → Line)](#75-multi-level-drill-section--account--journal--line)
   - [7.6 Export from Drill-Down Context](#76-export-from-drill-down-context)
   - [7.7 Comparative Columns (Period vs Period)](#77-comparative-columns-period-vs-period)
8. [Workflow Configuration UI](#8-workflow-configuration-ui)
   - [8.1 Approval Policy List](#81-approval-policy-list)
   - [8.2 Policy Builder — Rule Editor](#82-policy-builder--rule-editor)
   - [8.3 Rule Condition Form](#83-rule-condition-form)
   - [8.4 Approval Chain Visualizer](#84-approval-chain-visualizer)
   - [8.5 Live Approval Instance Tracker](#85-live-approval-instance-tracker)
   - [8.6 Escalation Configuration](#86-escalation-configuration)
9. [Full Page Composition Patterns](#9-full-page-composition-patterns)
10. [API Contract Reference](#10-api-contract-reference)
11. [Common amis Patterns & Pitfalls](#11-common-amis-patterns--pitfalls)

---

## 1. amis Architecture Overview

amis renders a complete UI from a **JSON schema**. Every page, form, table, dialog, and toolbar is described as a JSON object tree. There is no JSX or template HTML. The renderer maps `type` values to React components internally.

```
JSON Schema
    └── Page (type: page)
         ├── Toolbar (type: toolbar)
         ├── Body (any component)
         │    ├── CRUD (type: crud) → Table + Filters + Pagination + Toolbar
         │    ├── Form (type: form) → Fields, Validation, Submit
         │    ├── Tabs (type: tabs) → Lazy panels
         │    └── Chart (type: chart) → ECharts wrapper
         └── Aside (type: nav)
```

### Mounting the Renderer (React host)

```tsx
import { render as renderAmis } from 'amis';
import { ToastComponent, AlertComponent } from 'amis-ui';
import 'amis/lib/themes/cxd.css';

function ERPPage({ schema }: { schema: object }) {
  return (
    <>
      <ToastComponent />
      <AlertComponent />
      {renderAmis(schema, {
        // env: inject router, fetcher, copy, notify
        fetcher: ({ url, method, data, config }) =>
          axios({ url, method, data, ...config }).then(r => r.data),
        notify: (type, msg) => toast[type](msg),
        jumpTo: (to) => navigate(to),
        isCurrentUrl: (url) => location.pathname === url,
      })}
    </>
  );
}
```

### Data Flow Fundamentals

| Concept | amis Term | Description |
|---|---|---|
| URL variables | `$query.xxx` | Query string values |
| Page-level data | `$page.data.xxx` | Set via `data` prop on page |
| Action payload | `$event.data.xxx` | Data emitted by events |
| Store variables | `$$xxx` | Values in CRUD row context |
| Expression | `${varName}` | Template string interpolation |
| Condition | `visibleOn`, `disabledOn` | JavaScript expression string |

---

## 2. Theming & Global Defaults

amis ships with three themes: `cxd` (default blue), `antd`, and `dark`. AWO ERP uses `cxd`.

### CSS Variable Overrides (global.css)

Apply these after the amis theme CSS to match AWO brand:

```css
:root {
  /* Brand primary — keep close to amis cxd blue */
  --colors-brand-5: #1677ff;
  --colors-brand-6: #0958d9;

  /* Table row hover */
  --Table-onRow-bg: #f0f5ff;

  /* Sidebar background */
  --Nav-aside-bg: #001529;
  --Nav-aside-color: rgba(255,255,255,0.65);
  --Nav-aside-activeColor: #ffffff;
  --Nav-aside-activeBg: #1677ff;

  /* Form label width (wider for financial field names) */
  --Form-label-width: 160px;

  /* Currency formatting hint (used in column definitions) */
  --erp-currency-prefix: "KES ";
}
```

### Number Format Helpers

amis `tpl` columns and `quickEdit` support a `format` helper. For financial figures, use these column `type` and `tpl` patterns throughout:

```json
// Currency column (reuse this pattern everywhere)
{
  "name": "amount",
  "label": "Amount (KES)",
  "type": "tpl",
  "tpl": "KES ${amount | number: 2}",
  "align": "right",
  "width": 140
}

// Percentage column
{
  "name": "variance_pct",
  "label": "Variance %",
  "type": "tpl",
  "tpl": "${variance_pct | number: 2}%",
  "align": "right"
}

// Badge/status column
{
  "name": "status",
  "label": "Status",
  "type": "mapping",
  "map": {
    "GREEN":  "<span class='label label-success'>On Track</span>",
    "AMBER":  "<span class='label label-warning'>At Risk</span>",
    "RED":    "<span class='label label-danger'>Off Track</span>",
    "FAV":    "<span class='label label-success'>Favourable</span>",
    "ADV":    "<span class='label label-danger'>Adverse</span>"
  }
}
```

---

## 3. Navigation (Nav)

The nav component renders a vertical or horizontal navigation tree. For ERP sidebars, always use the vertical `aside` pattern inside a `page` with `asideResizable: true`.

### 3.1 Left Sidebar Navigation

The top-level page schema that frames every ERP page. The `aside` receives the nav; the `body` receives the routed content.

```json
{
  "type": "page",
  "aside": {
    "type": "nav",
    "stacked": true,
    "className": "erp-sidebar",
    "links": [
      {
        "label": "Dashboard",
        "to": "/dashboard",
        "icon": "fa fa-tachometer-alt"
      },
      {
        "label": "General Ledger",
        "icon": "fa fa-book",
        "children": [
          { "label": "Journal Entries",  "to": "/gl/journals" },
          { "label": "Chart of Accounts","to": "/gl/accounts" },
          { "label": "Trial Balance",    "to": "/gl/trial-balance" },
          { "label": "Periods",          "to": "/gl/periods" }
        ]
      },
      {
        "label": "Financial Reports",
        "icon": "fa fa-chart-bar",
        "children": [
          { "label": "Profit & Loss",   "to": "/reports/pl" },
          { "label": "Balance Sheet",   "to": "/reports/bs" },
          { "label": "Cash Flow",       "to": "/reports/cf" },
          { "label": "Budget vs Actual","to": "/reports/bva" }
        ]
      },
      {
        "label": "Payroll",
        "icon": "fa fa-users",
        "children": [
          { "label": "Payroll Runs",    "to": "/payroll/runs" },
          { "label": "Employees",       "to": "/payroll/employees" },
          { "label": "Rate Tables",     "to": "/payroll/rates" }
        ]
      },
      {
        "label": "Inventory",
        "icon": "fa fa-boxes",
        "children": [
          { "label": "Items",           "to": "/inventory/items" },
          { "label": "Movements",       "to": "/inventory/movements" },
          { "label": "Dip Readings",    "to": "/inventory/dip" },
          { "label": "ABC Analysis",    "to": "/inventory/abc" }
        ]
      },
      {
        "label": "Receivables",
        "icon": "fa fa-file-invoice-dollar",
        "children": [
          { "label": "Invoices",        "to": "/ar/invoices" },
          { "label": "Aging Report",    "to": "/ar/aging" },
          { "label": "Customers",       "to": "/ar/customers" }
        ]
      },
      {
        "label": "Payables",
        "icon": "fa fa-receipt",
        "children": [
          { "label": "Bills",           "to": "/ap/bills" },
          { "label": "Suppliers",       "to": "/ap/suppliers" },
          { "label": "Payment Runs",    "to": "/ap/payment-runs" }
        ]
      },
      {
        "label": "Approvals",
        "icon": "fa fa-check-circle",
        "to": "/approvals/inbox",
        "badge": {
          "text": "${pendingApprovals}",
          "visibleOn": "${pendingApprovals > 0}",
          "level": "danger"
        }
      },
      {
        "label": "Settings",
        "icon": "fa fa-cog",
        "children": [
          { "label": "Approval Policies","to": "/settings/approval-policies" },
          { "label": "Reorder Rules",    "to": "/settings/reorder-rules" },
          { "label": "Credit Profiles",  "to": "/settings/credit-profiles" },
          { "label": "KPI Definitions",  "to": "/settings/kpis" },
          { "label": "Users & Roles",    "to": "/settings/users" }
        ]
      }
    ]
  },
  "body": {
    "type": "container",
    "className": "erp-main-content",
    "body": "${pageBody}"
  }
}
```

### 3.2 Collapsible Groups

By default, clicking a parent with `children` toggles the group open/closed. Control the initial expanded state with `defaultExpanded`:

```json
{
  "label": "Financial Reports",
  "icon": "fa fa-chart-bar",
  "defaultExpanded": true,
  "children": [...]
}
```

To remember the expanded state across page navigations, use `saveExpanded`:

```json
{
  "type": "nav",
  "stacked": true,
  "saveExpanded": true,
  "expandedKeys": ["financial-reports"],
  "links": [...]
}
```

### 3.3 Role-Based Menu Filtering

Use `visibleOn` on any link. The expression receives the page-level data context, which should include `currentUser.roles[]`.

```json
{
  "label": "Settings",
  "icon": "fa fa-cog",
  "visibleOn": "${ARRAYINCLUDES(currentUser.roles, 'ADMIN') || ARRAYINCLUDES(currentUser.roles, 'FINANCE_MANAGER')}",
  "children": [
    {
      "label": "Rate Tables",
      "to": "/payroll/rates",
      "visibleOn": "${ARRAYINCLUDES(currentUser.roles, 'PAYROLL_ADMIN')}"
    }
  ]
}
```

> **Best practice:** Filter roles server-side too. `visibleOn` is a UI convenience only — it does not enforce security.

### 3.4 Nav with Badge Counts

Dynamic badge counts come from a page-level API that populates context variables:

```json
{
  "type": "page",
  "initApi": "/api/nav/counts",
  "aside": {
    "type": "nav",
    "stacked": true,
    "links": [
      {
        "label": "Approvals Inbox",
        "to": "/approvals/inbox",
        "icon": "fa fa-inbox",
        "badge": {
          "text": "${navCounts.pendingApprovals}",
          "visibleOn": "${navCounts.pendingApprovals > 0}",
          "level": "danger",
          "position": "top-right"
        }
      },
      {
        "label": "Reorder Alerts",
        "to": "/inventory/alerts",
        "icon": "fa fa-bell",
        "badge": {
          "text": "${navCounts.reorderAlerts}",
          "visibleOn": "${navCounts.reorderAlerts > 0}",
          "level": "warning"
        }
      }
    ]
  }
}
```

> The `initApi` response should return: `{ "navCounts": { "pendingApprovals": 4, "reorderAlerts": 2 } }`

---

## 4. Breadcrumbs

amis uses the `breadcrumb` component. Place it at the top of every `page` body to orient users within the ERP hierarchy.

### 4.1 Static Breadcrumb

```json
{
  "type": "breadcrumb",
  "items": [
    { "label": "Home",     "href": "/dashboard" },
    { "label": "Receivables" },
    { "label": "Invoices", "href": "/ar/invoices" },
    { "label": "INV-20241" }
  ]
}
```

### 4.2 Dynamic Breadcrumb (Context-Driven)

When navigating to a detail page, the breadcrumb label should reflect the record name. Use template expressions against the page's data context (populated by `initApi`):

```json
{
  "type": "page",
  "initApi": "/api/ar/invoices/${id}",
  "body": [
    {
      "type": "breadcrumb",
      "items": [
        { "label": "Home",       "href": "/dashboard" },
        { "label": "Receivables" },
        { "label": "Invoices",   "href": "/ar/invoices" },
        { "label": "${invoiceNumber}" }
      ]
    }
  ]
}
```

### 4.3 Breadcrumb + Page Title Row

Combine breadcrumb with a page title and action buttons in a flex toolbar row. This is the standard AWO ERP page header:

```json
{
  "type": "wrapper",
  "className": "erp-page-header",
  "body": [
    {
      "type": "breadcrumb",
      "items": [
        { "label": "Home",     "href": "/dashboard" },
        { "label": "GL",       "href": "/gl" },
        { "label": "Journals", "href": "/gl/journals" },
        { "label": "${journalNumber}" }
      ]
    },
    {
      "type": "flex",
      "alignItems": "center",
      "justifyContent": "space-between",
      "className": "mt-1",
      "items": [
        {
          "type": "tpl",
          "tpl": "<h2 class='erp-page-title'>${journalNumber} — ${description}</h2>"
        },
        {
          "type": "button-group",
          "buttons": [
            {
              "type": "button",
              "label": "Post",
              "level": "primary",
              "icon": "fa fa-check",
              "visibleOn": "${status === 'DRAFT'}",
              "actionType": "ajax",
              "api": "POST:/api/gl/journals/${id}/post",
              "confirmText": "Post this journal entry? This cannot be undone."
            },
            {
              "type": "button",
              "label": "Export PDF",
              "icon": "fa fa-file-pdf",
              "actionType": "ajax",
              "api": "GET:/api/gl/journals/${id}/export?format=pdf",
              "responseType": "blob"
            }
          ]
        }
      ]
    }
  ]
}
```

> **Spacing convention:** Always add `"className": "mt-2"` to the breadcrumb wrapper to separate it from the page top edge. Use `"className": "mb-3"` on the header wrapper before the content body.

---

## 5. Tabs

amis `tabs` render as a tabbed panel. Each tab can have its own `api` for lazy data loading — only fetched when the tab becomes active.

### 5.1 Standard Horizontal Tabs

```json
{
  "type": "tabs",
  "tabsMode": "line",
  "tabs": [
    {
      "title": "Overview",
      "icon": "fa fa-home",
      "body": {
        "type": "form",
        "mode": "horizontal",
        "wrapWithPanel": false,
        "body": [...]
      }
    },
    {
      "title": "Journal Lines",
      "icon": "fa fa-list",
      "body": {
        "type": "crud",
        "api": "/api/gl/journals/${id}/lines",
        "columns": [...]
      }
    },
    {
      "title": "Audit Trail",
      "icon": "fa fa-history",
      "body": {
        "type": "crud",
        "api": "/api/audit/${id}?entity=journal",
        "columns": [...]
      }
    }
  ]
}
```

**`tabsMode` options:**

| Value | Appearance | Use When |
|---|---|---|
| `line` | Underlined tabs (default) | Most content pages |
| `card` | Card-style tabs with background | Dashboards |
| `radio` | Button-group style | Small option sets (≤4) |
| `vertical` | Left-side tabs | Detail panels with many sections |
| `chrome` | Browser-tab style | Full-page multi-document views |

### 5.2 Tabs with Lazy Loading

Lazy loading prevents unnecessary API calls. The tab only fires its API when the user clicks it:

```json
{
  "type": "tabs",
  "tabs": [
    {
      "title": "Summary",
      "body": { "type": "form", "body": [...] }
      // No api here — summary data comes from parent page's initApi
    },
    {
      "title": "Invoice Lines",
      "mountOnEnter": true,     // render only when first activated
      "unmountOnExit": false,   // keep mounted after first visit (don't re-fetch)
      "body": {
        "type": "crud",
        "api": "/api/ar/invoices/${id}/lines",
        "columns": [...]
      }
    },
    {
      "title": "Payment History",
      "mountOnEnter": true,
      "unmountOnExit": false,
      "body": {
        "type": "crud",
        "api": "/api/ar/invoices/${id}/payments",
        "columns": [...]
      }
    },
    {
      "title": "Attachments",
      "mountOnEnter": true,
      "body": {
        "type": "crud",
        "api": "/api/attachments?entity=invoice&entityId=${id}",
        "columns": [
          { "name": "fileName", "label": "File" },
          { "name": "uploadedBy", "label": "Uploaded By" },
          { "name": "uploadedAt", "label": "Date", "type": "date" },
          {
            "type": "operation",
            "label": "Actions",
            "buttons": [
              {
                "label": "Download",
                "actionType": "url",
                "url": "${downloadUrl}"
              }
            ]
          }
        ]
      }
    }
  ]
}
```

### 5.3 Vertical Tabs (Detail Panels)

Use `tabsMode: vertical` on record detail pages where there are many sections. The left rail scrolls independently:

```json
{
  "type": "tabs",
  "tabsMode": "vertical",
  "className": "erp-detail-tabs",
  "tabs": [
    {
      "title": "General",
      "icon": "fa fa-info-circle",
      "body": { "type": "form", "mode": "horizontal", "body": [...] }
    },
    {
      "title": "Deductions",
      "icon": "fa fa-minus-circle",
      "mountOnEnter": true,
      "body": { "type": "crud", "api": "/api/payroll/employees/${id}/deductions", "columns": [...] }
    },
    {
      "title": "Pay History",
      "icon": "fa fa-history",
      "mountOnEnter": true,
      "body": { "type": "crud", "api": "/api/payroll/employees/${id}/history", "columns": [...] }
    },
    {
      "title": "Leave",
      "icon": "fa fa-calendar",
      "mountOnEnter": true,
      "body": { "type": "crud", "api": "/api/hr/employees/${id}/leave", "columns": [...] }
    },
    {
      "title": "Documents",
      "icon": "fa fa-folder",
      "mountOnEnter": true,
      "body": { "type": "crud", "api": "/api/attachments?entity=employee&entityId=${id}", "columns": [...] }
    }
  ]
}
```

### 5.4 Tab-Level Toolbar Actions

Add action buttons scoped to a specific tab using the `toolbar` property:

```json
{
  "title": "Journal Lines",
  "toolbar": [
    {
      "type": "button",
      "label": "Add Line",
      "level": "primary",
      "icon": "fa fa-plus",
      "actionType": "dialog",
      "dialog": {
        "title": "Add Journal Line",
        "body": { "type": "form", "api": "POST:/api/gl/journals/${id}/lines", "body": [...] }
      }
    },
    {
      "type": "button",
      "label": "Validate Balance",
      "icon": "fa fa-balance-scale",
      "actionType": "ajax",
      "api": "POST:/api/gl/journals/${id}/validate"
    }
  ],
  "body": {
    "type": "crud",
    "api": "/api/gl/journals/${id}/lines",
    "columns": [...]
  }
}
```

---

## 6. DataTables with Filters

The `crud` component is the most powerful amis component. It combines a data table, server-side filter form, pagination, bulk actions, inline editing, and row-level actions in a single JSON block.

### 6.1 Basic CRUD Table

Minimum viable CRUD schema:

```json
{
  "type": "crud",
  "api": "GET:/api/ar/invoices",
  "primaryField": "id",
  "perPage": 20,
  "columns": [
    { "name": "invoiceNumber", "label": "Invoice #", "sortable": true },
    { "name": "customerName",  "label": "Customer",  "sortable": true },
    { "name": "dueDate",       "label": "Due Date",  "type": "date", "sortable": true },
    {
      "name": "amountOutstanding",
      "label": "Outstanding (KES)",
      "type": "tpl",
      "tpl": "KES ${amountOutstanding | number: 2}",
      "align": "right",
      "sortable": true
    },
    {
      "name": "status",
      "label": "Status",
      "type": "mapping",
      "map": {
        "OPEN":    "<span class='label label-primary'>Open</span>",
        "PARTIAL": "<span class='label label-warning'>Partial</span>",
        "PAID":    "<span class='label label-success'>Paid</span>",
        "OVERDUE": "<span class='label label-danger'>Overdue</span>"
      }
    },
    {
      "type": "operation",
      "label": "Actions",
      "width": 120,
      "buttons": [
        {
          "label": "View",
          "type": "button",
          "actionType": "link",
          "link": "/ar/invoices/${id}"
        },
        {
          "label": "Send Reminder",
          "type": "button",
          "visibleOn": "${status !== 'PAID'}",
          "actionType": "ajax",
          "api": "POST:/api/ar/invoices/${id}/remind",
          "confirmText": "Send payment reminder to ${customerName}?"
        }
      ]
    }
  ]
}
```

### 6.2 Inline Filter Toolbar

The `filter` property in `crud` renders a form above the table. Every field in the filter form maps to a query parameter sent to the API.

```json
{
  "type": "crud",
  "api": "GET:/api/ar/invoices",
  "filter": {
    "title": "",
    "wrapWithPanel": false,
    "submitOnChange": false,
    "body": [
      {
        "type": "input-text",
        "name": "keyword",
        "label": "Search",
        "placeholder": "Invoice number or customer name",
        "clearable": true,
        "size": "md"
      },
      {
        "type": "select",
        "name": "status",
        "label": "Status",
        "multiple": true,
        "clearable": true,
        "options": [
          { "label": "Open",    "value": "OPEN" },
          { "label": "Partial", "value": "PARTIAL" },
          { "label": "Overdue", "value": "OVERDUE" },
          { "label": "Paid",    "value": "PAID" }
        ]
      },
      {
        "type": "input-date-range",
        "name": "dueDateRange",
        "label": "Due Date",
        "format": "YYYY-MM-DD",
        "clearable": true
      },
      {
        "type": "select",
        "name": "customerId",
        "label": "Customer",
        "source": "/api/ar/customers?select=id,name",
        "labelField": "name",
        "valueField": "id",
        "searchable": true,
        "clearable": true
      }
    ],
    "actions": [
      { "type": "submit", "label": "Search", "level": "primary" },
      { "type": "reset",  "label": "Reset" }
    ]
  },
  "columns": [...]
}
```

### 6.3 Collapsible Filter Panel (Query Form)

For complex reports with many filters, use a collapsible panel that hides advanced options behind an "Advanced Filters" toggle:

```json
{
  "type": "crud",
  "api": "GET:/api/gl/journals",
  "filterTogglable": true,
  "filterDefaultVisible": true,
  "filter": {
    "wrapWithPanel": false,
    "body": [
      {
        "type": "group",
        "body": [
          {
            "type": "input-text",
            "name": "keyword",
            "label": "Search",
            "placeholder": "Journal number or description",
            "clearable": true
          },
          {
            "type": "select",
            "name": "status",
            "label": "Status",
            "options": [
              { "label": "All",     "value": "" },
              { "label": "Draft",   "value": "DRAFT" },
              { "label": "Posted",  "value": "POSTED" },
              { "label": "Reversed","value": "REVERSED" }
            ]
          },
          {
            "type": "select",
            "name": "periodId",
            "label": "Period",
            "source": "/api/gl/periods?status=OPEN,SOFT_CLOSED",
            "labelField": "name",
            "valueField": "id",
            "clearable": true
          }
        ]
      },
      {
        "type": "collapse",
        "title": "Advanced Filters",
        "collapsed": true,
        "body": [
          {
            "type": "group",
            "body": [
              {
                "type": "select",
                "name": "costCentreId",
                "label": "Cost Centre",
                "source": "/api/cost-centres",
                "labelField": "name",
                "valueField": "id",
                "clearable": true
              },
              {
                "type": "input-number",
                "name": "amountFrom",
                "label": "Amount ≥",
                "min": 0,
                "precision": 2
              },
              {
                "type": "input-number",
                "name": "amountTo",
                "label": "Amount ≤",
                "min": 0,
                "precision": 2
              },
              {
                "type": "input-date-range",
                "name": "postedAtRange",
                "label": "Posted Date",
                "format": "YYYY-MM-DD"
              }
            ]
          }
        ]
      }
    ],
    "actions": [
      { "type": "submit", "label": "Apply", "level": "primary", "icon": "fa fa-search" },
      { "type": "reset",  "label": "Clear All" }
    ]
  },
  "columns": [...]
}
```

### 6.4 Column-Level Features

#### Sortable Column

```json
{ "name": "amount", "label": "Amount", "sortable": true, "defaultSort": "desc" }
```

#### Searchable Column (Quick Filter)

```json
{ "name": "customerName", "label": "Customer", "searchable": true }
```

#### Quick Edit (Inline)

```json
{
  "name": "creditLimit",
  "label": "Credit Limit (KES)",
  "quickEdit": {
    "type": "input-number",
    "saveImmediately": {
      "api": "PATCH:/api/ar/customers/${id}",
      "data": { "creditLimit": "${creditLimit}" }
    },
    "validations": { "minimum": 0 }
  }
}
```

#### Conditional Row Styling

Apply row-level background colours based on data values using `rowClassNameExpr`:

```json
{
  "type": "crud",
  "rowClassNameExpr": "${status === 'OVERDUE' ? 'table-danger' : status === 'PARTIAL' ? 'table-warning' : ''}",
  "columns": [...]
}
```

#### Column with Tooltip

```json
{
  "name": "dipVariancePct",
  "label": "Variance %",
  "type": "tpl",
  "tpl": "${dipVariancePct | number: 3}%",
  "align": "right",
  "remark": "Positive = shrinkage (theoretical > actual). Normal range: ±0.3% diesel, ±0.5% petrol."
}
```

### 6.5 Server-Side Pagination & Sorting

The API must return a specific response envelope for amis pagination to work:

```json
// API Response: GET /api/ar/invoices?page=2&perPage=20&orderBy=dueDate&orderDir=asc
{
  "status": 0,
  "data": {
    "total": 347,
    "items": [...],
    "count": 20
  }
}
```

CRUD schema for server-side pagination:

```json
{
  "type": "crud",
  "api": "GET:/api/ar/invoices",
  "perPage": 20,
  "defaultParams": {
    "orderBy": "dueDate",
    "orderDir": "asc"
  },
  "syncLocation": true,
  "keepItemSelectionOnPageChange": true,
  "columns": [...]
}
```

> `syncLocation: true` writes pagination state to the URL query string so the browser Back button works correctly.

### 6.6 Bulk Actions & Row Selection

```json
{
  "type": "crud",
  "api": "GET:/api/ap/bills",
  "bulkActions": [
    {
      "label": "Approve Selected",
      "level": "primary",
      "icon": "fa fa-check",
      "actionType": "ajax",
      "api": {
        "method": "post",
        "url": "/api/ap/bills/bulk-approve",
        "data": { "ids": "${ids}" }
      },
      "confirmText": "Approve ${ids.length} selected bills?"
    },
    {
      "label": "Export Selected",
      "icon": "fa fa-download",
      "actionType": "ajax",
      "api": {
        "method": "post",
        "url": "/api/ap/bills/export",
        "data": { "ids": "${ids}", "format": "excel" },
        "responseType": "blob"
      }
    }
  ],
  "headerToolbar": [
    "bulkActions",
    "columns-toggler",
    {
      "type": "button",
      "label": "New Bill",
      "level": "primary",
      "icon": "fa fa-plus",
      "actionType": "link",
      "link": "/ap/bills/new"
    },
    "export-excel",
    "reload"
  ],
  "columns": [...]
}
```

The `ids` variable in bulk action payloads is automatically populated by amis with the `primaryField` values of all checked rows.

### 6.7 Expandable Row Detail

Show a mini-preview of line items or sub-records by expanding a row in place — avoiding a page navigation for quick inspection:

```json
{
  "type": "crud",
  "api": "GET:/api/gl/journals",
  "expandable": {
    "expandedRowRender": {
      "type": "crud",
      "api": "/api/gl/journals/${id}/lines",
      "title": "",
      "className": "ml-4 mt-1 mb-2",
      "columns": [
        { "name": "accountCode",  "label": "Account" },
        { "name": "accountName",  "label": "Name" },
        {
          "name": "debit",
          "label": "Debit",
          "type": "tpl",
          "tpl": "${debit > 0 ? 'KES ' + FORMAT_NUMBER(debit, 2) : '—'}",
          "align": "right"
        },
        {
          "name": "credit",
          "label": "Credit",
          "type": "tpl",
          "tpl": "${credit > 0 ? 'KES ' + FORMAT_NUMBER(credit, 2) : '—'}",
          "align": "right"
        },
        { "name": "costCentreName", "label": "Cost Centre" },
        { "name": "description",    "label": "Narration" }
      ]
    }
  },
  "columns": [...]
}
```

### 6.8 Fixed Columns & Frozen Headers

For wide financial tables, freeze the reference column and action column:

```json
{
  "type": "crud",
  "api": "GET:/api/reports/trial-balance",
  "tableClassName": "table-bordered",
  "affixHeader": true,
  "columns": [
    {
      "name": "accountCode",
      "label": "Code",
      "fixed": "left",
      "width": 80,
      "sortable": true
    },
    {
      "name": "accountName",
      "label": "Account Name",
      "fixed": "left",
      "width": 240
    },
    { "name": "openingDebit",   "label": "Opening Dr",  "type": "tpl", "tpl": "KES ${openingDebit | number:2}", "align": "right", "width": 140 },
    { "name": "openingCredit",  "label": "Opening Cr",  "type": "tpl", "tpl": "KES ${openingCredit | number:2}", "align": "right", "width": 140 },
    { "name": "periodDebit",    "label": "Period Dr",   "type": "tpl", "tpl": "KES ${periodDebit | number:2}", "align": "right", "width": 140 },
    { "name": "periodCredit",   "label": "Period Cr",   "type": "tpl", "tpl": "KES ${periodCredit | number:2}", "align": "right", "width": 140 },
    { "name": "closingDebit",   "label": "Closing Dr",  "type": "tpl", "tpl": "KES ${closingDebit | number:2}", "align": "right", "width": 140 },
    { "name": "closingCredit",  "label": "Closing Cr",  "type": "tpl", "tpl": "KES ${closingCredit | number:2}", "align": "right", "width": 140 },
    {
      "type": "operation",
      "label": "Actions",
      "fixed": "right",
      "width": 90,
      "buttons": [
        { "label": "Drill", "icon": "fa fa-search-plus", "actionType": "dialog", "dialog": { "$ref": "#drillDownDialog" } }
      ]
    }
  ]
}
```

---

## 7. Reporting Drill-Down

This is the most complex UI pattern in the AWO ERP. A drill-down originates from a summary figure in a financial report (e.g., a P&L section total) and progressively reveals the underlying data through multiple levels.

### 7.1 Architecture Overview

```
Level 0: Report Summary
  P&L page — sections with totals (Revenue, COGS, Expenses)
      │
      ▼ click section total
Level 1: Account List
  All accounts in section — each with period balance
      │
      ▼ click account balance
Level 2: Journal Entry List
  All posted journals touching this account in this period
      │
      ▼ click journal number
Level 3: Journal Detail
  Full journal with all lines, narrations, cost centres
      │
      ▼ click source document link
Level 4: Source Document
  Original invoice, bill, receipt, or payroll run
```

Each level passes a `DrillDownContext` to the next:
- `tenantId` (from session — never user-supplied)
- `periodId`
- `accountId` (added at Level 1→2)
- `journalId` (added at Level 2→3)
- `costCentreId` (optional filter — passed through all levels)

### 7.2 P&L Report Page Schema

The top-level P&L page. Section totals are clickable links that open the Level 1 drill-down:

```json
{
  "type": "page",
  "title": "Profit & Loss Statement",
  "toolbar": [
    {
      "type": "form",
      "wrapWithPanel": false,
      "mode": "inline",
      "body": [
        {
          "type": "select",
          "name": "periodId",
          "label": "Period",
          "source": "/api/gl/periods?status=OPEN,SOFT_CLOSED,HARD_CLOSED",
          "labelField": "name",
          "valueField": "id",
          "value": "${defaultPeriodId}"
        },
        {
          "type": "select",
          "name": "costCentreId",
          "label": "Department",
          "source": "/api/cost-centres",
          "labelField": "name",
          "valueField": "id",
          "clearable": true,
          "placeholder": "All Departments"
        },
        {
          "type": "select",
          "name": "compareWith",
          "label": "Compare",
          "options": [
            { "label": "None",             "value": "" },
            { "label": "Prior Period",     "value": "prior_period" },
            { "label": "Prior Year",       "value": "prior_year" },
            { "label": "Budget",           "value": "budget" }
          ]
        }
      ],
      "actions": [
        { "type": "submit", "label": "Refresh", "level": "primary" },
        {
          "type": "dropdown-button",
          "label": "Export",
          "icon": "fa fa-download",
          "buttons": [
            {
              "label": "Export PDF",
              "icon": "fa fa-file-pdf",
              "actionType": "ajax",
              "api": "POST:/api/reports/pl/export?format=pdf&periodId=${periodId}&costCentreId=${costCentreId}",
              "responseType": "blob"
            },
            {
              "label": "Export Excel",
              "icon": "fa fa-file-excel",
              "actionType": "ajax",
              "api": "POST:/api/reports/pl/export?format=excel&periodId=${periodId}&costCentreId=${costCentreId}",
              "responseType": "blob"
            }
          ]
        }
      ]
    }
  ],
  "initApi": {
    "url": "/api/reports/pl",
    "method": "get",
    "data": {
      "periodId": "${periodId || defaultPeriodId}",
      "costCentreId": "${costCentreId}",
      "compareWith": "${compareWith}"
    },
    "sendOn": "${periodId}"
  },
  "body": {
    "type": "crud",
    "api": "/api/reports/pl?periodId=${periodId}&costCentreId=${costCentreId}&compareWith=${compareWith}",
    "loadDataOnce": true,
    "showFooter": true,
    "tableClassName": "table-bordered erp-pl-table",
    "columns": [
      {
        "name": "sectionName",
        "label": "Description",
        "width": "40%",
        "type": "tpl",
        "tpl": "<span class='pl-indent-${level}'>${sectionName}</span>"
      },
      {
        "name": "amount",
        "label": "Current Period (KES)",
        "align": "right",
        "type": "tpl",
        "tpl": "<a class='drill-link ${isTotal ? 'font-bold' : ''}' data-action='drill' data-period='${periodId}' data-account='${accountId}' data-section='${sectionId}'>${isZero ? '—' : 'KES ' + FORMAT_NUMBER(amount, 2)}</a>"
      },
      {
        "name": "compareAmount",
        "label": "${compareLabel}",
        "align": "right",
        "visibleOn": "${compareWith !== ''}",
        "type": "tpl",
        "tpl": "KES ${compareAmount | number: 2}"
      },
      {
        "name": "variance",
        "label": "Variance",
        "align": "right",
        "visibleOn": "${compareWith !== ''}",
        "type": "tpl",
        "tpl": "<span class='${varianceClass}'>${variance | number: 2}%</span>"
      }
    ],
    "onEvent": {
      "click": {
        "actions": [
          {
            "actionType": "dialog",
            "args": {
              "dialog": {
                "$ref": "drillDownL1Dialog"
              }
            },
            "data": {
              "drillPeriodId": "${event.data.periodId}",
              "drillAccountId": "${event.data.accountId}",
              "drillSectionId": "${event.data.sectionId}"
            }
          }
        ]
      }
    }
  }
}
```

### 7.3 Drill-Down Modal — Transaction Detail

Level 1 modal: shows all accounts in the clicked section with their individual balances.

```json
{
  "type": "dialog",
  "title": "Accounts — ${drillSectionName}",
  "size": "xl",
  "body": [
    {
      "type": "breadcrumb",
      "items": [
        { "label": "P&L",              "onClick": "closeAllModals()" },
        { "label": "${drillSectionName}" }
      ]
    },
    {
      "type": "crud",
      "api": {
        "url": "/api/reports/drill/accounts",
        "method": "get",
        "data": {
          "periodId":   "${drillPeriodId}",
          "sectionId":  "${drillSectionId}",
          "costCentreId": "${drillCostCentreId}"
        }
      },
      "loadDataOnce": true,
      "columns": [
        { "name": "accountCode", "label": "Code",    "width": 80 },
        { "name": "accountName", "label": "Account", "width": "40%" },
        {
          "name": "amount",
          "label": "Balance (KES)",
          "align": "right",
          "type": "tpl",
          "tpl": "<a class='drill-link' href='javascript:void(0)' onclick='drillToJournals(\"${accountId}\",\"${drillPeriodId}\")'>KES ${amount | number: 2}</a>"
        },
        {
          "type": "operation",
          "label": "",
          "width": 60,
          "buttons": [
            {
              "icon": "fa fa-search-plus",
              "tooltip": "Drill to journal entries",
              "actionType": "dialog",
              "data": {
                "drillAccountId": "${accountId}",
                "drillAccountName": "${accountName}"
              },
              "dialog": {
                "$ref": "drillDownL2Dialog"
              }
            }
          ]
        }
      ],
      "footerToolbar": [
        {
          "type": "tpl",
          "tpl": "<strong>Total: KES ${ARRAYREDUCE(items, (sum, r) => sum + r.amount, 0) | number: 2}</strong>"
        }
      ]
    }
  ],
  "actions": [
    { "type": "button", "label": "Close", "actionType": "close" },
    {
      "type": "button",
      "label": "Export This View",
      "icon": "fa fa-download",
      "actionType": "ajax",
      "api": "POST:/api/reports/drill/export?periodId=${drillPeriodId}&sectionId=${drillSectionId}&format=excel",
      "responseType": "blob"
    }
  ]
}
```

Level 2 modal: journal entries for a specific account in the period.

```json
{
  "type": "dialog",
  "title": "Journal Entries — ${drillAccountName}",
  "size": "xl",
  "body": [
    {
      "type": "breadcrumb",
      "items": [
        { "label": "P&L",                "onClick": "closeAllModals()" },
        { "label": "${drillSectionName}", "onClick": "goBack(1)" },
        { "label": "${drillAccountName}" }
      ]
    },
    {
      "type": "crud",
      "api": {
        "url": "/api/reports/drill/journals",
        "method": "get",
        "data": {
          "periodId":     "${drillPeriodId}",
          "accountId":    "${drillAccountId}",
          "costCentreId": "${drillCostCentreId}"
        }
      },
      "columns": [
        { "name": "journalNumber",  "label": "Journal #",   "sortable": true },
        { "name": "postedAt",       "label": "Date",        "type": "date", "sortable": true },
        { "name": "description",    "label": "Description" },
        { "name": "reference",      "label": "Reference" },
        { "name": "postedByName",   "label": "Posted By" },
        {
          "name": "netAmount",
          "label": "Amount (KES)",
          "align": "right",
          "type": "tpl",
          "tpl": "KES ${netAmount | number: 2}"
        },
        {
          "type": "operation",
          "label": "",
          "buttons": [
            {
              "icon": "fa fa-eye",
              "tooltip": "View journal detail",
              "actionType": "dialog",
              "data": { "drillJournalId": "${id}", "drillJournalNumber": "${journalNumber}" },
              "dialog": { "$ref": "drillDownL3Dialog" }
            }
          ]
        }
      ]
    }
  ]
}
```

### 7.4 Drill-Down Breadcrumb Trail

The drill-down breadcrumb must be maintained manually across modal levels since amis modals are not routed. Use a shared state array in the page data context:

```javascript
// JavaScript helpers injected via amis env.utils
window.drillStack = [];

function pushDrill(level) {
  window.drillStack.push(level);
  renderDrillBreadcrumb();
}

function popDrill() {
  window.drillStack.pop();
  renderDrillBreadcrumb();
}

function closeAllModals() {
  window.drillStack = [];
  // trigger amis action: close all dialogs
}

function renderDrillBreadcrumb() {
  // updates the breadcrumb component via amis store
}
```

For a simpler approach, implement each drill level as a full page navigation (not modal), using URL query parameters to carry the drill context. This leverages the browser's native Back button and makes each drill level bookmarkable:

```
/reports/pl?periodId=P2024-03
/reports/pl/drill/section?periodId=P2024-03&sectionId=REVENUE
/reports/pl/drill/account?periodId=P2024-03&accountId=ACC-4000
/reports/pl/drill/journal?periodId=P2024-03&accountId=ACC-4000&journalId=JNL-8821
```

Breadcrumb on drill pages then reads from URL params:

```json
{
  "type": "breadcrumb",
  "items": [
    { "label": "P&L",           "href": "/reports/pl?periodId=${query.periodId}" },
    { "label": "${sectionName}","href": "/reports/pl/drill/section?periodId=${query.periodId}&sectionId=${query.sectionId}", "visibleOn": "${query.sectionId}" },
    { "label": "${accountName}","href": "/reports/pl/drill/account?periodId=${query.periodId}&accountId=${query.accountId}", "visibleOn": "${query.accountId}" },
    { "label": "${journalNumber}", "visibleOn": "${query.journalId}" }
  ]
}
```

### 7.5 Multi-Level Drill (Section → Account → Journal → Line)

Level 3 — Journal Lines detail modal:

```json
{
  "type": "dialog",
  "title": "Journal ${drillJournalNumber}",
  "size": "lg",
  "body": [
    {
      "type": "form",
      "mode": "horizontal",
      "wrapWithPanel": false,
      "api": "/api/gl/journals/${drillJournalId}",
      "body": [
        { "type": "static", "name": "journalNumber", "label": "Journal #" },
        { "type": "static", "name": "postedAt",      "label": "Posted",      "format": "date" },
        { "type": "static", "name": "postedByName",  "label": "Posted By" },
        { "type": "static", "name": "description",   "label": "Description" },
        { "type": "static", "name": "reference",     "label": "Reference" }
      ]
    },
    {
      "type": "divider"
    },
    {
      "type": "crud",
      "api": "/api/gl/journals/${drillJournalId}/lines",
      "loadDataOnce": true,
      "title": "Journal Lines",
      "tableClassName": "table-bordered",
      "columns": [
        { "name": "accountCode",    "label": "Account",     "width": 80 },
        { "name": "accountName",    "label": "Name" },
        { "name": "costCentreName", "label": "Cost Centre" },
        { "name": "description",    "label": "Narration" },
        {
          "name": "debit",
          "label": "Debit (KES)",
          "align": "right",
          "type": "tpl",
          "tpl": "${debit > 0 ? 'KES ' + FORMAT_NUMBER(debit, 2) : ''}"
        },
        {
          "name": "credit",
          "label": "Credit (KES)",
          "align": "right",
          "type": "tpl",
          "tpl": "${credit > 0 ? 'KES ' + FORMAT_NUMBER(credit, 2) : ''}"
        }
      ],
      "footerToolbar": [
        {
          "type": "tpl",
          "tpl": "<div class='flex justify-end gap-8'><span><strong>Total Debits:</strong> KES ${ARRAYREDUCE(items, (s,r) => s+r.debit, 0) | number:2}</span><span><strong>Total Credits:</strong> KES ${ARRAYREDUCE(items, (s,r) => s+r.credit, 0) | number:2}</span></div>"
        }
      ]
    },
    {
      "type": "action",
      "label": "View Source Document",
      "visibleOn": "${sourceDocumentType && sourceDocumentId}",
      "actionType": "link",
      "link": "/documents/${sourceDocumentType}/${sourceDocumentId}"
    }
  ],
  "actions": [
    {
      "type": "button",
      "label": "View Full Journal",
      "level": "link",
      "actionType": "link",
      "link": "/gl/journals/${drillJournalId}"
    },
    { "type": "button", "label": "Close", "actionType": "close" }
  ]
}
```

### 7.6 Export from Drill-Down Context

Every drill-down level should offer an export. The export API receives the same context parameters as the data API:

```json
{
  "type": "dropdown-button",
  "label": "Export",
  "icon": "fa fa-download",
  "size": "sm",
  "buttons": [
    {
      "label": "Export to Excel",
      "actionType": "ajax",
      "api": {
        "url": "/api/reports/drill/export",
        "method": "post",
        "data": {
          "level":        "journals",
          "periodId":     "${drillPeriodId}",
          "accountId":    "${drillAccountId}",
          "costCentreId": "${drillCostCentreId}",
          "format":       "excel"
        },
        "responseType": "blob",
        "headers": { "Accept": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" }
      }
    },
    {
      "label": "Export to CSV",
      "actionType": "ajax",
      "api": {
        "url": "/api/reports/drill/export",
        "method": "post",
        "data": {
          "level":    "journals",
          "periodId": "${drillPeriodId}",
          "accountId":"${drillAccountId}",
          "format":   "csv"
        },
        "responseType": "blob",
        "headers": { "Accept": "text/csv" }
      }
    }
  ]
}
```

### 7.7 Comparative Columns (Period vs Period)

When `compareWith` is set, the P&L table renders additional columns. Manage column visibility with `visibleOn`:

```json
{
  "type": "crud",
  "api": "/api/reports/pl?periodId=${periodId}&compareWith=${compareWith}",
  "columns": [
    { "name": "description", "label": "Description", "width": "35%" },
    {
      "name": "currentAmount",
      "label": "Current Period",
      "align": "right",
      "type": "tpl",
      "tpl": "KES ${currentAmount | number: 2}"
    },
    {
      "name": "compareAmount",
      "label": "${compareLabel}",
      "align": "right",
      "visibleOn": "${compareWith !== ''}",
      "type": "tpl",
      "tpl": "KES ${compareAmount | number: 2}"
    },
    {
      "name": "varianceAmount",
      "label": "Variance (KES)",
      "align": "right",
      "visibleOn": "${compareWith !== ''}",
      "type": "tpl",
      "tpl": "<span class='${varianceAmount >= 0 && rowType === 'REVENUE' ? 'text-success' : varianceAmount < 0 && rowType === 'EXPENSE' ? 'text-success' : 'text-danger'}'>${varianceAmount | number: 2}</span>"
    },
    {
      "name": "variancePct",
      "label": "Variance %",
      "align": "right",
      "visibleOn": "${compareWith !== ''}",
      "type": "tpl",
      "tpl": "<span class='${variancePct >= 0 && rowType === 'REVENUE' ? 'text-success' : 'text-danger'}'>${variancePct | number: 1}%</span>"
    }
  ]
}
```

---

## 8. Workflow Configuration UI

The approval workflow configuration spans three interconnected pages: the Policy List, the Policy Builder (with rule-level editors), and the live Approval Instance Tracker.

### 8.1 Approval Policy List

```json
{
  "type": "page",
  "title": "Approval Policies",
  "body": [
    {
      "type": "breadcrumb",
      "items": [
        { "label": "Home",     "href": "/dashboard" },
        { "label": "Settings" },
        { "label": "Approval Policies" }
      ]
    },
    {
      "type": "crud",
      "api": "GET:/api/settings/approval-policies",
      "primaryField": "id",
      "headerToolbar": [
        {
          "type": "button",
          "label": "New Policy",
          "level": "primary",
          "icon": "fa fa-plus",
          "actionType": "link",
          "link": "/settings/approval-policies/new"
        },
        "reload",
        "columns-toggler"
      ],
      "columns": [
        { "name": "name",        "label": "Policy Name", "sortable": true },
        {
          "name": "entityType",
          "label": "Applies To",
          "type": "mapping",
          "map": {
            "PURCHASE_ORDER":    "Purchase Order",
            "PAYMENT":           "Payment",
            "JOURNAL_ENTRY":     "Journal Entry",
            "EXPENSE_CLAIM":     "Expense Claim",
            "CREDIT_NOTE":       "Credit Note"
          }
        },
        {
          "name": "ruleCount",
          "label": "Rules",
          "type": "tpl",
          "tpl": "${ruleCount} rule${ruleCount !== 1 ? 's' : ''}"
        },
        {
          "name": "isActive",
          "label": "Active",
          "type": "switch",
          "saveImmediately": {
            "api": "PATCH:/api/settings/approval-policies/${id}",
            "data": { "isActive": "${isActive}" }
          }
        },
        { "name": "updatedAt", "label": "Last Modified", "type": "datetime", "format": "DD MMM YYYY HH:mm" },
        {
          "type": "operation",
          "label": "Actions",
          "buttons": [
            {
              "label": "Configure",
              "icon": "fa fa-cog",
              "actionType": "link",
              "link": "/settings/approval-policies/${id}"
            },
            {
              "label": "Duplicate",
              "icon": "fa fa-copy",
              "actionType": "ajax",
              "api": "POST:/api/settings/approval-policies/${id}/duplicate",
              "confirmText": "Duplicate '${name}'?"
            },
            {
              "label": "Delete",
              "icon": "fa fa-trash",
              "level": "danger",
              "visibleOn": "${!isActive}",
              "actionType": "ajax",
              "api": "DELETE:/api/settings/approval-policies/${id}",
              "confirmText": "Permanently delete '${name}'? This cannot be undone."
            }
          ]
        }
      ]
    }
  ]
}
```

### 8.2 Policy Builder — Rule Editor

The policy detail page. It combines a policy header form with an ordered list of approval rules:

```json
{
  "type": "page",
  "initApi": "/api/settings/approval-policies/${id}",
  "body": [
    {
      "type": "breadcrumb",
      "items": [
        { "label": "Home",             "href": "/dashboard" },
        { "label": "Settings" },
        { "label": "Approval Policies","href": "/settings/approval-policies" },
        { "label": "${name}" }
      ]
    },
    {
      "type": "tabs",
      "tabs": [
        {
          "title": "Policy Settings",
          "icon": "fa fa-sliders-h",
          "body": {
            "type": "form",
            "api": "PUT:/api/settings/approval-policies/${id}",
            "mode": "horizontal",
            "labelWidth": 200,
            "body": [
              {
                "type": "input-text",
                "name": "name",
                "label": "Policy Name",
                "required": true,
                "placeholder": "e.g. Purchase Order Approval"
              },
              {
                "type": "select",
                "name": "entityType",
                "label": "Applies To",
                "required": true,
                "options": [
                  { "label": "Purchase Order",  "value": "PURCHASE_ORDER" },
                  { "label": "Payment",         "value": "PAYMENT" },
                  { "label": "Journal Entry",   "value": "JOURNAL_ENTRY" },
                  { "label": "Expense Claim",   "value": "EXPENSE_CLAIM" }
                ]
              },
              {
                "type": "switch",
                "name": "isActive",
                "label": "Active",
                "onText": "Yes",
                "offText": "No"
              },
              {
                "type": "textarea",
                "name": "description",
                "label": "Description",
                "rows": 3
              }
            ],
            "actions": [
              { "type": "submit", "label": "Save Settings", "level": "primary" }
            ]
          }
        },
        {
          "title": "Approval Rules",
          "icon": "fa fa-list-ol",
          "toolbar": [
            {
              "type": "button",
              "label": "Add Rule",
              "level": "primary",
              "icon": "fa fa-plus",
              "actionType": "dialog",
              "dialog": { "$ref": "addRuleDialog" }
            }
          ],
          "body": {
            "type": "crud",
            "api": "/api/settings/approval-policies/${id}/rules",
            "primaryField": "id",
            "draggable": true,
            "itemDraggableOn": true,
            "saveOrderApi": "POST:/api/settings/approval-policies/${id}/rules/reorder",
            "columns": [
              {
                "name": "",
                "label": "",
                "width": 30,
                "type": "tpl",
                "tpl": "<i class='fa fa-grip-vertical text-muted drag-handle'></i>"
              },
              {
                "name": "sequenceNumber",
                "label": "Step",
                "width": 60,
                "type": "tpl",
                "tpl": "<span class='step-badge'>${sequenceNumber}</span>"
              },
              {
                "name": "conditionSummary",
                "label": "Condition",
                "type": "tpl",
                "tpl": "<code>${conditionField} ${conditionOp} ${conditionValue}</code>"
              },
              {
                "name": "approverType",
                "label": "Approver",
                "type": "mapping",
                "map": {
                  "ROLE":            "<i class='fa fa-users'></i> Role: ${approverRef}",
                  "USER":            "<i class='fa fa-user'></i> User: ${approverRef}",
                  "MANAGER":         "<i class='fa fa-user-tie'></i> Line Manager",
                  "COST_CENTRE_HEAD": "<i class='fa fa-building'></i> Cost Centre Head",
                  "FINANCE_DIRECTOR": "<i class='fa fa-user-shield'></i> Finance Director"
                }
              },
              {
                "name": "escalationDays",
                "label": "Escalate After",
                "type": "tpl",
                "tpl": "${escalationDays ? escalationDays + ' day(s)' : 'No escalation'}"
              },
              {
                "type": "operation",
                "label": "Actions",
                "buttons": [
                  {
                    "label": "Edit",
                    "icon": "fa fa-edit",
                    "actionType": "dialog",
                    "data": { "$$": "$$" },
                    "dialog": { "$ref": "editRuleDialog" }
                  },
                  {
                    "label": "Delete",
                    "icon": "fa fa-trash",
                    "level": "danger",
                    "actionType": "ajax",
                    "api": "DELETE:/api/settings/approval-policies/${policyId}/rules/${id}",
                    "confirmText": "Delete this approval rule?"
                  }
                ]
              }
            ]
          }
        },
        {
          "title": "Test Policy",
          "icon": "fa fa-vial",
          "mountOnEnter": true,
          "body": { "$ref": "policyTestPanel" }
        }
      ]
    }
  ]
}
```

### 8.3 Rule Condition Form

The dialog form for creating or editing a single approval rule. The `conditionField` dropdown drives the available `conditionOp` options and the `conditionValue` field type via `visibleOn` and `requiredOn`:

```json
{
  "type": "dialog",
  "title": "Approval Rule",
  "size": "md",
  "body": {
    "type": "form",
    "api": "${id ? 'PUT:/api/settings/approval-policies/${policyId}/rules/${id}' : 'POST:/api/settings/approval-policies/${policyId}/rules'}",
    "mode": "horizontal",
    "labelWidth": 180,
    "body": [
      {
        "type": "divider",
        "title": "Condition (when this rule applies)"
      },
      {
        "type": "select",
        "name": "conditionField",
        "label": "Field",
        "required": true,
        "options": [
          { "label": "Transaction Amount",    "value": "amount" },
          { "label": "Transaction Type",      "value": "entity_type" },
          { "label": "Cost Centre",           "value": "cost_centre_id" },
          { "label": "Vendor Category",       "value": "vendor_category" },
          { "label": "Requester Role",        "value": "requester_role" },
          { "label": "Department",            "value": "department_id" }
        ]
      },
      {
        "type": "select",
        "name": "conditionOp",
        "label": "Operator",
        "required": true,
        "visibleOn": "${conditionField === 'amount'}",
        "options": [
          { "label": ">  Greater than",      "value": "GT" },
          { "label": "≥  At least",          "value": "GTE" },
          { "label": "<  Less than",         "value": "LT" },
          { "label": "≤  At most",           "value": "LTE" },
          { "label": "=  Exactly",           "value": "EQ" },
          { "label": "Between (range)",      "value": "BETWEEN" }
        ]
      },
      {
        "type": "select",
        "name": "conditionOp",
        "label": "Operator",
        "required": true,
        "visibleOn": "${conditionField !== 'amount'}",
        "options": [
          { "label": "= Is",                 "value": "EQ" },
          { "label": "≠ Is not",             "value": "NEQ" },
          { "label": "In list",              "value": "IN" }
        ]
      },
      {
        "type": "input-number",
        "name": "conditionValue",
        "label": "Amount (KES)",
        "required": true,
        "min": 0,
        "precision": 2,
        "visibleOn": "${conditionField === 'amount' && conditionOp !== 'BETWEEN'}"
      },
      {
        "type": "group",
        "visibleOn": "${conditionField === 'amount' && conditionOp === 'BETWEEN'}",
        "body": [
          { "type": "input-number", "name": "conditionValueFrom", "label": "From (KES)", "required": true, "min": 0 },
          { "type": "input-number", "name": "conditionValueTo",   "label": "To (KES)",   "required": true, "min": 0 }
        ]
      },
      {
        "type": "select",
        "name": "conditionValue",
        "label": "Cost Centre",
        "source": "/api/cost-centres",
        "labelField": "name",
        "valueField": "id",
        "multiple": "${conditionOp === 'IN'}",
        "visibleOn": "${conditionField === 'cost_centre_id'}",
        "required": true
      },
      {
        "type": "divider",
        "title": "Approver"
      },
      {
        "type": "select",
        "name": "approverType",
        "label": "Approver Type",
        "required": true,
        "options": [
          { "label": "Specific User",         "value": "USER" },
          { "label": "Role",                  "value": "ROLE" },
          { "label": "Line Manager",          "value": "MANAGER" },
          { "label": "Cost Centre Head",      "value": "COST_CENTRE_HEAD" },
          { "label": "Finance Director",      "value": "FINANCE_DIRECTOR" }
        ]
      },
      {
        "type": "select",
        "name": "approverRef",
        "label": "User",
        "source": "/api/users?role=APPROVER",
        "labelField": "name",
        "valueField": "id",
        "searchable": true,
        "required": true,
        "visibleOn": "${approverType === 'USER'}"
      },
      {
        "type": "select",
        "name": "approverRef",
        "label": "Role",
        "options": [
          { "label": "Finance Manager",   "value": "FINANCE_MANAGER" },
          { "label": "Finance Director",  "value": "FINANCE_DIRECTOR" },
          { "label": "Operations Manager","value": "OPERATIONS_MANAGER" },
          { "label": "Board Member",      "value": "BOARD_MEMBER" }
        ],
        "required": true,
        "visibleOn": "${approverType === 'ROLE'}"
      },
      {
        "type": "divider",
        "title": "Escalation"
      },
      {
        "type": "input-number",
        "name": "escalationDays",
        "label": "Escalate After (days)",
        "min": 1,
        "max": 30,
        "placeholder": "Leave blank for no escalation",
        "description": "Auto-escalate to the next level if no decision is made within this many business days."
      },
      {
        "type": "select",
        "name": "escalateTo",
        "label": "Escalate To",
        "visibleOn": "${escalationDays > 0}",
        "options": [
          { "label": "Finance Director",  "value": "FINANCE_DIRECTOR" },
          { "label": "Managing Director", "value": "MANAGING_DIRECTOR" }
        ]
      }
    ]
  }
}
```

### 8.4 Approval Chain Visualizer

A read-only panel showing how a hypothetical transaction would route through the policy. This helps administrators verify their configuration. Uses amis `steps` component:

```json
{
  "type": "panel",
  "title": "Simulated Approval Chain",
  "body": [
    {
      "type": "form",
      "wrapWithPanel": false,
      "mode": "inline",
      "body": [
        {
          "type": "input-number",
          "name": "testAmount",
          "label": "Amount (KES)",
          "min": 0,
          "value": 50000
        },
        {
          "type": "select",
          "name": "testCostCentreId",
          "label": "Cost Centre",
          "source": "/api/cost-centres",
          "labelField": "name",
          "valueField": "id"
        }
      ],
      "actions": [
        {
          "type": "submit",
          "label": "Simulate",
          "level": "primary",
          "icon": "fa fa-play"
        }
      ],
      "api": {
        "url": "/api/settings/approval-policies/${policyId}/simulate",
        "method": "post"
      }
    },
    {
      "type": "divider"
    },
    {
      "type": "steps",
      "value": "${simulatedChain.length}",
      "steps": "${simulatedChain}",
      "labelPlacement": "vertical",
      "visibleOn": "${simulatedChain && simulatedChain.length > 0}"
    },
    {
      "type": "alert",
      "level": "warning",
      "body": "No approval rules matched this transaction. It will be auto-approved.",
      "visibleOn": "${simulatedChain && simulatedChain.length === 0}"
    }
  ]
}
```

The `/simulate` API should return:

```json
{
  "status": 0,
  "data": {
    "simulatedChain": [
      { "title": "Step 1",  "subTitle": "Finance Manager",  "description": "Amount > KES 10,000 → ROLE:FINANCE_MANAGER", "status": "finish" },
      { "title": "Step 2",  "subTitle": "Finance Director", "description": "Amount > KES 100,000 → ROLE:FINANCE_DIRECTOR", "status": "wait" }
    ]
  }
}
```

### 8.5 Live Approval Instance Tracker

The approvals inbox page. Shows pending approvals for the logged-in user with full context and one-click approve/reject actions:

```json
{
  "type": "page",
  "title": "Approvals Inbox",
  "body": [
    {
      "type": "breadcrumb",
      "items": [
        { "label": "Home",    "href": "/dashboard" },
        { "label": "Approvals Inbox" }
      ]
    },
    {
      "type": "tabs",
      "tabs": [
        {
          "title": "Pending",
          "badge": { "text": "${pendingCount}", "level": "danger", "visibleOn": "${pendingCount > 0}" },
          "body": {
            "type": "crud",
            "api": "/api/approvals/inbox?status=PENDING&approverId=${currentUser.id}",
            "autoFillHeight": true,
            "perPage": 10,
            "headerToolbar": ["reload"],
            "rowClassNameExpr": "${isOverdue ? 'table-warning' : ''}",
            "columns": [
              {
                "name": "entityType",
                "label": "Type",
                "type": "mapping",
                "map": {
                  "PURCHASE_ORDER": "<span class='label label-info'>PO</span>",
                  "PAYMENT":        "<span class='label label-primary'>Payment</span>",
                  "EXPENSE_CLAIM":  "<span class='label label-default'>Expense</span>",
                  "JOURNAL_ENTRY":  "<span class='label label-warning'>Journal</span>"
                }
              },
              { "name": "entityReference", "label": "Reference",   "sortable": true },
              { "name": "description",     "label": "Description" },
              { "name": "requesterName",   "label": "Requested By" },
              {
                "name": "amount",
                "label": "Amount (KES)",
                "align": "right",
                "type": "tpl",
                "tpl": "KES ${amount | number: 2}"
              },
              {
                "name": "submittedAt",
                "label": "Submitted",
                "type": "datetime",
                "fromNow": true,
                "sortable": true
              },
              {
                "name": "isOverdue",
                "label": "",
                "type": "tpl",
                "tpl": "${isOverdue ? '<span class=\"label label-danger\"><i class=\"fa fa-clock\"></i> Overdue</span>' : ''}"
              },
              {
                "type": "operation",
                "label": "Actions",
                "width": 200,
                "buttons": [
                  {
                    "label": "View",
                    "icon": "fa fa-eye",
                    "actionType": "dialog",
                    "dialog": {
                      "type": "dialog",
                      "title": "${entityReference} — Review",
                      "size": "xl",
                      "body": {
                        "type": "service",
                        "api": "/api/${entityType.toLowerCase().replace('_','-')}/${entityId}",
                        "body": { "$ref": "entityDetailPanel" }
                      },
                      "actions": [
                        {
                          "type": "button",
                          "label": "Approve",
                          "level": "success",
                          "icon": "fa fa-check",
                          "actionType": "ajax",
                          "api": {
                            "url": "/api/approvals/${approvalStepId}/approve",
                            "method": "post",
                            "data": { "comment": "${comment}" }
                          },
                          "reload": "window",
                          "close": true
                        },
                        {
                          "type": "button",
                          "label": "Reject",
                          "level": "danger",
                          "icon": "fa fa-times",
                          "actionType": "dialog",
                          "dialog": {
                            "title": "Confirm Rejection",
                            "body": {
                              "type": "form",
                              "body": [
                                {
                                  "type": "textarea",
                                  "name": "rejectionReason",
                                  "label": "Reason",
                                  "required": true,
                                  "placeholder": "State the reason for rejection..."
                                }
                              ],
                              "api": {
                                "url": "/api/approvals/${approvalStepId}/reject",
                                "method": "post"
                              }
                            }
                          }
                        },
                        { "type": "button", "label": "Close", "actionType": "close" }
                      ]
                    }
                  },
                  {
                    "label": "Approve",
                    "level": "success",
                    "icon": "fa fa-check",
                    "actionType": "ajax",
                    "api": "POST:/api/approvals/${approvalStepId}/approve",
                    "confirmText": "Approve ${entityReference}?"
                  },
                  {
                    "label": "Reject",
                    "level": "danger",
                    "icon": "fa fa-times",
                    "actionType": "dialog",
                    "dialog": {
                      "title": "Reject ${entityReference}",
                      "body": {
                        "type": "form",
                        "api": "POST:/api/approvals/${approvalStepId}/reject",
                        "body": [
                          {
                            "type": "textarea",
                            "name": "rejectionReason",
                            "label": "Reason",
                            "required": true,
                            "minLength": 10
                          }
                        ]
                      }
                    }
                  }
                ]
              }
            ]
          }
        },
        {
          "title": "Approved",
          "mountOnEnter": true,
          "body": {
            "type": "crud",
            "api": "/api/approvals/inbox?status=APPROVED&approverId=${currentUser.id}",
            "columns": [
              { "name": "entityReference", "label": "Reference" },
              { "name": "description",     "label": "Description" },
              { "name": "amount",          "label": "Amount (KES)", "type": "tpl", "tpl": "KES ${amount|number:2}", "align": "right" },
              { "name": "decidedAt",       "label": "Approved",     "type": "datetime" }
            ]
          }
        },
        {
          "title": "Rejected",
          "mountOnEnter": true,
          "body": {
            "type": "crud",
            "api": "/api/approvals/inbox?status=REJECTED&approverId=${currentUser.id}",
            "columns": [
              { "name": "entityReference",  "label": "Reference" },
              { "name": "rejectionReason",  "label": "Reason" },
              { "name": "amount",           "label": "Amount (KES)", "type": "tpl", "tpl": "KES ${amount|number:2}", "align": "right" },
              { "name": "decidedAt",        "label": "Rejected",     "type": "datetime" }
            ]
          }
        }
      ]
    }
  ]
}
```

### 8.6 Escalation Configuration

A dedicated section within the policy for configuring escalation targets and SLA timers per entity type:

```json
{
  "type": "form",
  "title": "Escalation Settings",
  "api": "PUT:/api/settings/approval-policies/${id}/escalation",
  "mode": "horizontal",
  "labelWidth": 220,
  "body": [
    {
      "type": "switch",
      "name": "escalationEnabled",
      "label": "Enable Auto-Escalation",
      "description": "Automatically escalate approvals that are not actioned within the SLA."
    },
    {
      "type": "input-number",
      "name": "defaultSlaHours",
      "label": "Default SLA (hours)",
      "min": 1,
      "max": 720,
      "value": 48,
      "visibleOn": "${escalationEnabled}",
      "description": "Business hours. 48 = 2 business days."
    },
    {
      "type": "select",
      "name": "escalationTargetRole",
      "label": "Escalate To",
      "visibleOn": "${escalationEnabled}",
      "options": [
        { "label": "Finance Director",  "value": "FINANCE_DIRECTOR" },
        { "label": "Managing Director", "value": "MANAGING_DIRECTOR" },
        { "label": "Board Member",      "value": "BOARD_MEMBER" }
      ]
    },
    {
      "type": "switch",
      "name": "notifyRequesterOnEscalation",
      "label": "Notify Requester",
      "visibleOn": "${escalationEnabled}",
      "description": "Send the requester an email when their request is escalated."
    },
    {
      "type": "switch",
      "name": "notifyOriginalApproverOnEscalation",
      "label": "Notify Original Approver",
      "visibleOn": "${escalationEnabled}",
      "description": "Notify the overdue approver that escalation has occurred."
    }
  ],
  "actions": [
    { "type": "submit", "label": "Save Escalation Settings", "level": "primary" }
  ]
}
```

---

## 9. Full Page Composition Patterns

### Standard List Page Pattern

Every list/index page in the ERP follows this structure:

```
Page
├── Breadcrumb
├── Page Header Row (title + primary action button)
├── CRUD
│    ├── Header Toolbar (bulk actions, column toggler, export, reload)
│    ├── Filter Form (collapsed by default on wide screens)
│    ├── Table (with sortable columns, status badges, row actions)
│    └── Footer Toolbar (pagination + total count)
└── [optional] Summary Stats Bar (total, overdue count, etc.)
```

### Standard Detail Page Pattern

```
Page (initApi: /api/resource/${id})
├── Breadcrumb (dynamic — last crumb = record identifier)
├── Page Header Row (record title + status badge + action buttons)
├── Alert Bar (if status = ATTENTION_REQUIRED, show contextual warning)
└── Tabs (vertical on wide screens)
     ├── Tab: Overview (read-only form, horizontal mode)
     ├── Tab: Related Records (lazy CRUD, mountOnEnter)
     ├── Tab: Activity / Timeline (lazy CRUD of audit events)
     └── Tab: Attachments (lazy CRUD of documents)
```

### Dashboard Widget Pattern

```
Page (initApi: /api/dashboard?tenantId=${tenantId})
├── Grid (3 columns, responsive)
│    ├── Stat Card: Revenue MTD
│    ├── Stat Card: Outstanding AR
│    ├── Stat Card: Pending Approvals
│    ├── Stat Card: Dip Variance Alerts
│    ├── Chart: P&L Trend (6 months)    [span 2 columns]
│    └── Chart: AR Aging Distribution   [span 1 column]
└── Two-Column Layout
     ├── CRUD: Recent Transactions (last 10, no pagination)
     └── CRUD: Pending Approvals (inbox preview, top 5)
```

---

## 10. API Contract Reference

All API responses must conform to the amis response envelope:

### Success Response

```json
{
  "status": 0,
  "msg": "",
  "data": {
    // resource-specific payload
  }
}
```

### Error Response

```json
{
  "status": 400,
  "msg": "Journal imbalance: net 150.00. Debits must equal credits.",
  "errors": {
    "lines": "Debit total 10150.00 ≠ Credit total 10000.00"
  },
  "data": null
}
```

### List Response (for CRUD pagination)

```json
{
  "status": 0,
  "data": {
    "total": 347,
    "items": [...],
    "count": 20,
    "page": 2
  }
}
```

### Key API Endpoints

| Screen | Method | URL | Notes |
|---|---|---|---|
| Nav badge counts | GET | `/api/nav/counts` | Returns `{ navCounts: { pendingApprovals, reorderAlerts } }` |
| P&L Report | GET | `/api/reports/pl` | Params: `periodId`, `costCentreId`, `compareWith` |
| P&L Drill — Accounts | GET | `/api/reports/drill/accounts` | Params: `periodId`, `sectionId`, `costCentreId` |
| P&L Drill — Journals | GET | `/api/reports/drill/journals` | Params: `periodId`, `accountId`, `costCentreId` |
| P&L Drill — Export | POST | `/api/reports/drill/export` | Body: `{ level, periodId, accountId, format }` |
| Simulate Approval | POST | `/api/settings/approval-policies/{id}/simulate` | Body: `{ testAmount, testCostCentreId }` |
| Approve Step | POST | `/api/approvals/{stepId}/approve` | Body: `{ comment }` |
| Reject Step | POST | `/api/approvals/{stepId}/reject` | Body: `{ rejectionReason }` |
| Reorder Rules | POST | `/api/settings/approval-policies/{id}/rules/reorder` | Body: `{ ids: [...] }` |
| Export Report | POST | `/api/reports/{type}/export` | Body: `{ periodId, format }`, `responseType: blob` |

---

## 11. Common amis Patterns & Pitfalls

### Pitfall 1 — Float Formatting

Never rely on JavaScript's default `toLocaleString` inside `tpl`. Use amis's built-in `| number: N` filter:

```json
// ✅ Correct
"tpl": "KES ${amount | number: 2}"

// ❌ Wrong — may produce locale-incorrect output
"tpl": "KES ${amount.toFixed(2)}"
```

### Pitfall 2 — Concurrent Filter + Sort Reset

When a user changes a filter, `crud` resets to page 1. This is correct behaviour. But if you also store sort state in the URL (`syncLocation: true`), ensure your API reads `orderBy` and `orderDir` from query params, not just body.

### Pitfall 3 — CRUD `loadDataOnce` vs Live Queries

Use `loadDataOnce: true` only for static or slow-changing reference data (chart of accounts, cost centres). For transactional data (journal entries, invoices), always fetch live so you get the latest state.

### Pitfall 4 — Tenant ID in Filters

The `tenant_id` is **never** a user-visible filter. Inject it server-side from the authenticated session. If a CRUD API is accidentally exposed without tenant isolation, it becomes a data breach. Check every API: `$1` in all SQL queries is `tenant_id` from session, never from request body.

### Pitfall 5 — Dialog Data Scope

In amis, data inside a `dialog` is isolated from the parent page. Explicitly pass required context using the `data` property on the action that opens the dialog:

```json
{
  "actionType": "dialog",
  "data": {
    "drillPeriodId": "${periodId}",
    "drillAccountId": "${accountId}"
  },
  "dialog": { ... }
}
```

### Pattern: Refresh Parent After Dialog Submit

When a dialog form submits and closes, reload the parent CRUD:

```json
{
  "type": "dialog",
  "body": {
    "type": "form",
    "api": "POST:/api/resource",
    "reload": "parentCrudName"   // matches the crud's `name` property
  }
}
```

### Pattern: Conditional Required Fields

Use `requiredOn` for fields that are required only under certain conditions:

```json
{
  "type": "textarea",
  "name": "rejectionReason",
  "label": "Rejection Reason",
  "requiredOn": "${action === 'REJECT'}",
  "visibleOn": "${action === 'REJECT'}"
}
```

### Pattern: Role-Based Field Visibility

Protect sensitive fields from non-privileged users:

```json
{
  "type": "input-number",
  "name": "creditLimit",
  "label": "Credit Limit (KES)",
  "visibleOn": "${ARRAYINCLUDES(currentUser.roles, 'CREDIT_MANAGER') || ARRAYINCLUDES(currentUser.roles, 'ADMIN')}",
  "disabled": "${!ARRAYINCLUDES(currentUser.roles, 'CREDIT_MANAGER')}"
}
```

### Pattern: Auto-Submit Filter on Select Change

For quick filters (e.g., period selector), auto-submit so users don't need to click "Search":

```json
{
  "type": "crud",
  "filter": {
    "submitOnChange": true,
    "body": [
      {
        "type": "select",
        "name": "periodId",
        "label": "Period",
        "source": "/api/gl/periods",
        "labelField": "name",
        "valueField": "id"
      }
    ]
  }
}
```

### Pattern: Preserve Filter State in URL

Financial report pages should persist filter state in the URL so sharing a link (or refreshing the page) restores the same view:

```json
{
  "type": "crud",
  "syncLocation": true,
  "defaultParams": {
    "periodId": "${query.periodId || defaultPeriodId}",
    "compareWith": "${query.compareWith || ''}"
  }
}
```

---

*AWO ERP — amis UI Component Documentation*  
*Version 1.0 — Internal Engineering Reference*  
*Covers: Navigation, Breadcrumbs, Tabs, DataTables, Reporting Drill-Down, Workflow Configuration*
