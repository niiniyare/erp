[<-- Back to Index](README.md)

## Component Reference

Core AMIS components used in every Awo ERP module.

### `page` — Screen Container

```json
{
  "type":     "page",
  "title":    "Purchase Orders",
  "subTitle": "Manage supplier purchase orders",
  "toolbar": [
    { "type": "button", "label": "New Order", "level": "primary",
      "actionType": "link", "link": "/purchasing/orders/new" }
  ],
  "body": { "...": "content" }
}
```

### `crud` — List + Filter + Sort + Pagination + Actions

The most important component. One declaration gives you everything.

```json
{
  "type": "crud",
  "id":   "po-list",
  "api":  "get:/api/v1/purchase-orders",
  "syncLocation": true,

  "headerToolbar": ["search-box", "bulkActions",
    { "type": "columns-toggler" }, { "type": "reload" }],

  "filter": {
    "title": "Filter",
    "body": [
      { "type": "input-text",       "name": "keywords", "label": "Search" },
      { "type": "select",           "name": "status",   "label": "Status",
        "options": [{"label":"Draft","value":"draft"},{"label":"Confirmed","value":"confirmed"}],
        "clearable": true },
      { "type": "input-date-range", "name": "date_range", "label": "Date",
        "format": "YYYY-MM-DD" }
    ]
  },

  "columns": [
    { "name": "reference",     "label": "Reference",  "sortable": true },
    { "name": "supplier_name", "label": "Supplier",   "sortable": true },
    { "name": "total",         "label": "Total",      "type": "number",
      "prefix": "KES ", "sortable": true },
    { "name": "status",        "label": "Status",     "type": "tag",
      "colorMap": { "draft": "default", "confirmed": "success", "cancelled": "error" } },
    { "name": "created_at",    "label": "Date",       "type": "date", "sortable": true },
    {
      "type": "operation", "label": "Actions",
      "buttons": [
        { "type": "button", "label": "View",   "level": "link",
          "actionType": "link", "link": "/purchasing/orders/${id}" },
        { "type": "button", "label": "Submit", "level": "link",
          "visibleOn":   "${status === 'draft'}",
          "actionType":  "ajax",
          "api":         "post:/api/v1/purchase-orders/${id}/submit",
          "confirmText": "Submit this order for approval?" }
      ]
    }
  ],

  "bulkActions": [
    { "type": "button", "label": "Export Selected",
      "actionType": "ajax",
      "api": { "method": "post", "url": "/api/v1/purchase-orders/export",
               "data": { "ids": "${ids}" } } }
  ]
}
```

### `form` — Create / Edit

```json
{
  "type":     "form",
  "api":      "post:/api/v1/purchase-orders",
  "initApi":  "get:/api/v1/purchase-orders/${id}",
  "redirect": "/purchasing/orders/${id}",
  "body": [
    { "type": "select",       "name": "supplier_id", "label": "Supplier",
      "required": true, "source": "get:/api/v1/suppliers", "searchable": true,
      "valueField": "id", "labelField": "name" },
    { "type": "input-date",   "name": "expected_date", "label": "Expected Delivery",
      "required": true },
    { "type": "combo",        "name": "line_items",    "label": "Line Items",
      "required": true, "multiple": true,
      "items": [
        { "type": "select",       "name": "product_id", "label": "Product",
          "source": "get:/api/v1/catalog/products", "required": true },
        { "type": "input-number", "name": "quantity",   "label": "Qty",   "min": 1 },
        { "type": "input-number", "name": "unit_price", "label": "Price", "min": 0 }
      ] }
  ],
  "actions": [
    { "type": "submit", "label": "Save Draft", "level": "default" },
    { "type": "button", "label": "Cancel", "actionType": "link",
      "link": "/purchasing/orders" }
  ]
}
```

### `wizard` — Multi-Step Process

```json
{
  "type":  "wizard",
  "api":   "post:/api/v1/payroll/runs",
  "steps": [
    { "title": "Select Period",
      "body": [
        { "type": "select",     "name": "period_id", "label": "Pay Period",
          "source": "get:/api/v1/payroll/periods?status=open", "required": true },
        { "type": "input-date", "name": "pay_date",  "label": "Payment Date", "required": true }
      ]
    },
    { "title": "Review Employees",
      "initApi": "get:/api/v1/payroll/preview?period_id=${period_id}",
      "body": [
        { "type": "table", "source": "${employees}",
          "columns": [
            { "name": "name",    "label": "Employee" },
            { "name": "gross",   "label": "Gross",   "type": "number" },
            { "name": "net_pay", "label": "Net Pay", "type": "number" }
          ]}
      ]
    },
    { "title": "Confirm & Submit",
      "body": [
        { "type": "tpl",
          "tpl": "Processing payroll for <strong>${employee_count}</strong> employees — KES ${ROUND(total_net,2)}." }
      ]
    }
  ]
}
```

### `dialog` and `drawer`

Use `drawer` for detail views and editing. Use `dialog` for short confirmations.

```json
{
  "type": "button", "label": "Edit", "level": "link",
  "actionType": "drawer",
  "drawer": {
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

### `tabs` — Sectioned Detail Pages

```json
{
  "type": "tabs",
  "tabs": [
    { "title": "Overview",    "body": { "type": "descriptions", "...": "..." } },
    { "title": "Line Items",  "body": { "type": "crud", "api": "get:/api/v1/purchase-orders/${id}/lines" } },
    { "title": "Approvals",   "body": { "type": "service", "api": "get:/api/v1/purchase-orders/${id}/approvals",
        "body": { "type": "timeline", "source": "${steps}" } } },
    { "title": "Activity",    "body": { "type": "service", "api": "get:/api/v1/purchase-orders/${id}/audit",
        "body": { "type": "timeline", "source": "${events}" } } }
  ]
}
```

### `descriptions` — Read-Only Detail

```json
{
  "type":   "descriptions",
  "source": "get:/api/v1/purchase-orders/${id}",
  "items": [
    { "name": "reference",     "label": "Reference" },
    { "name": "supplier_name", "label": "Supplier"  },
    { "name": "total",         "label": "Total",   "type": "number" },
    { "name": "status",        "label": "Status"   },
    { "name": "created_at",    "label": "Created", "type": "date"   }
  ]
}
```

### `chart` — ECharts Wrapper

```json
{
  "type": "chart",
  "api":  "get:/api/v1/analytics/revenue/monthly",
  "config": {
    "backgroundColor": "transparent",
    "xAxis":  { "type": "category", "data": "${months}" },
    "yAxis":  { "type": "value" },
    "series": [{ "type": "bar", "data": "${values}", "name": "Revenue KES" }]
  }
}
```

Always set `"backgroundColor": "transparent"` for dark mode compatibility.

### `timeline` — History / Audit Log

```json
{
  "type":   "timeline",
  "source": "${history}",
  "items": {
    "time":   "${occurred_at}",
    "title":  "${action}",
    "detail": "${actor_name}: ${comment}"
  }
}
```

### `condition-builder` — Visual Rule Editor

Used in the Business Rules module. See [§11 Page Patterns](./11-page-patterns.md) for the full rules builder pattern.

```json
{
  "type":  "condition-builder",
  "name":  "conditions",
  "label": "Conditions",
  "fields": [
    { "label": "PO Total", "name": "po.total", "type": "number",
      "operators": ["equal","less","greater","between"] },
    { "label": "Supplier Tier", "name": "po.supplier_tier", "type": "select",
      "operators": ["equal","in","not_in"],
      "options": [{"label":"Preferred","value":"preferred"},{"label":"New","value":"new"}] }
  ]
}
```

---
