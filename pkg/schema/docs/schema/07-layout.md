# Layout System

Complete guide to page layouts, component organization, data tables, and navigation patterns.

## Page Type (Top Level)

`page` is a **top-level type** - each JSON schema file can have 0 or 1 page definition.

### Correct Structure

```json
{
  "type": "page",
  "title": "Transactions",
  "body": {
    // components go here
  }
}
```

❌ **WRONG - Page as nested type:**
```json
{
  "type": "form",
  "layout": {
    "type": "page"  // WRONG - page is not a layout type
  }
}
```

---

## Page Schema Structure

### Complete Page Definition

```json
{
  "type": "page",
  "id": "transaction-list",
  "title": "Transactions",
  "subTitle": "Manage financial transactions",
  
  "aside": {
    "type": "nav",
    "items": [...]
  },
  
  "toolbar": [
    {
      "type": "button",
      "label": "New Transaction",
      "actionType": "link",
      "link": "/finance/transactions/new",
      "level": "primary"
    }
  ],
  
  "body": [
    {
      "type": "crud",
      "api": "/api/finance/transactions",
      "columns": [...],
      "bulkActions": [...],
      "itemActions": [...]
    }
  ]
}
```

### Page Properties

```go
type Page struct {
    Type     string      `json:"type"`              // "page"
    ID       string      `json:"id,omitempty"`
    Title    string      `json:"title,omitempty"`
    SubTitle string      `json:"subTitle,omitempty"`
    
    // Navigation
    Aside      *Component   `json:"aside,omitempty"`      // Sidebar
    Header     *Component   `json:"header,omitempty"`     // Top bar
    Toolbar    []Component  `json:"toolbar,omitempty"`    // Action buttons
    
    // Content
    Body       []Component  `json:"body"`                 // Main content
    
    // Meta
    ClassName  string       `json:"className,omitempty"`
    CSS        string       `json:"css,omitempty"`
    InitAPI    string       `json:"initApi,omitempty"`
}
```

---

## Page Types and Patterns

### 1. List Page (CRUD Table)

**GET /finance/transactions**

```json
{
  "type": "page",
  "title": "Transactions",
  "subTitle": "Financial transaction management",
  
  "toolbar": [
    {
      "type": "button",
      "label": "New Transaction",
      "icon": "fa fa-plus",
      "actionType": "link",
      "link": "/finance/transactions/new",
      "level": "primary"
    },
    {
      "type": "button",
      "label": "Import",
      "icon": "fa fa-upload",
      "actionType": "dialog",
      "dialog": {
        "title": "Import Transactions",
        "body": {
          "type": "form",
          "api": "/api/finance/transactions/import",
          "body": [
            {
              "type": "input-file",
              "name": "file",
              "label": "Select File",
              "accept": ".csv,.xlsx"
            }
          ]
        }
      }
    }
  ],
  
  "body": [
    {
      "type": "crud",
      "syncLocation": false,
      "api": "/api/finance/transactions",
      "quickSaveApi": "/api/finance/transactions/$id",
      "quickSaveItemApi": "/api/finance/transactions/$id",
      "draggable": false,
      "headerToolbar": [
        "bulkActions",
        {
          "type": "export-excel",
          "label": "Export"
        },
        {
          "type": "columns-toggler",
          "align": "right"
        },
        "pagination"
      ],
      "footerToolbar": ["statistics", "pagination"],
      
      "filter": {
        "title": "Search",
        "body": [
          {
            "type": "input-text",
            "name": "keywords",
            "placeholder": "Search transactions...",
            "clearable": true
          },
          {
            "type": "select",
            "name": "status",
            "label": "Status",
            "options": [
              {"label": "All", "value": ""},
              {"label": "Pending", "value": "pending"},
              {"label": "Completed", "value": "completed"},
              {"label": "Cancelled", "value": "cancelled"}
            ]
          },
          {
            "type": "input-date-range",
            "name": "date_range",
            "label": "Date Range"
          },
          {
            "type": "input-number",
            "name": "amount_min",
            "label": "Min Amount",
            "prefix": "$"
          },
          {
            "type": "input-number",
            "name": "amount_max",
            "label": "Max Amount",
            "prefix": "$"
          }
        ]
      },
      
      "bulkActions": [
        {
          "label": "Approve Selected",
          "actionType": "ajax",
          "api": "post:/api/finance/transactions/bulk-approve",
          "confirmText": "Approve ${ids.length} transactions?"
        },
        {
          "label": "Delete Selected",
          "actionType": "ajax",
          "api": "delete:/api/finance/transactions/bulk-delete",
          "confirmText": "Delete ${ids.length} transactions? This cannot be undone."
        }
      ],
      
      "columns": [
        {
          "name": "id",
          "label": "ID",
          "type": "text",
          "width": 100,
          "sortable": true,
          "searchable": true,
          "toggled": true,
          "tpl": "<a href='/finance/transactions/${id}'>TRX-${id}</a>"
        },
        {
          "name": "date",
          "label": "Date",
          "type": "date",
          "format": "MMM DD, YYYY",
          "sortable": true,
          "searchable": false,
          "toggled": true
        },
        {
          "name": "description",
          "label": "Description",
          "type": "text",
          "sortable": false,
          "toggled": true
        },
        {
          "name": "category",
          "label": "Category",
          "type": "tag",
          "sortable": true,
          "toggled": true
        },
        {
          "name": "amount",
          "label": "Amount",
          "type": "number",
          "prefix": "$",
          "precision": 2,
          "sortable": true,
          "className": "text-right font-bold",
          "toggled": true
        },
        {
          "name": "status",
          "label": "Status",
          "type": "mapping",
          "sortable": true,
          "toggled": true,
          "map": {
            "pending": "<span class='label label-warning'>Pending</span>",
            "completed": "<span class='label label-success'>Completed</span>",
            "cancelled": "<span class='label label-danger'>Cancelled</span>"
          }
        },
        {
          "name": "created_by",
          "label": "Created By",
          "type": "text",
          "toggled": false
        },
        {
          "name": "created_at",
          "label": "Created At",
          "type": "datetime",
          "format": "YYYY-MM-DD HH:mm",
          "sortable": true,
          "toggled": false
        }
      ],
      
      "itemActions": [
        {
          "label": "View",
          "type": "button",
          "actionType": "link",
          "link": "/finance/transactions/${id}",
          "icon": "fa fa-eye",
          "level": "link"
        },
        {
          "label": "Edit",
          "type": "button",
          "actionType": "link",
          "link": "/finance/transactions/${id}/edit",
          "icon": "fa fa-pencil",
          "level": "link",
          "visibleOn": "this.canEdit"
        },
        {
          "label": "Duplicate",
          "type": "button",
          "actionType": "ajax",
          "api": "post:/api/finance/transactions/${id}/duplicate",
          "icon": "fa fa-copy",
          "confirmText": "Duplicate this transaction?"
        },
        {
          "label": "Delete",
          "type": "button",
          "actionType": "ajax",
          "api": "delete:/api/finance/transactions/${id}",
          "icon": "fa fa-trash",
          "level": "danger",
          "confirmText": "Delete transaction TRX-${id}? This cannot be undone.",
          "visibleOn": "this.canDelete"
        }
      ],
      
      "perPage": 25,
      "perPageAvailable": [10, 25, 50, 100],
      "defaultParams": {
        "perPage": 25,
        "orderBy": "date",
        "orderDir": "desc"
      }
    }
  ]
}
```

### 2. Detail Page (View Record)

**GET /finance/transactions/{id}**

```json
{
  "type": "page",
  "title": "Transaction Details",
  "subTitle": "TRX-${id}",
  "initApi": "/api/finance/transactions/${id}",
  
  "toolbar": [
    {
      "type": "button",
      "label": "Back to List",
      "actionType": "link",
      "link": "/finance/transactions",
      "icon": "fa fa-arrow-left"
    },
    {
      "type": "button",
      "label": "Edit",
      "actionType": "link",
      "link": "/finance/transactions/${id}/edit",
      "level": "primary",
      "icon": "fa fa-pencil",
      "visibleOn": "data.canEdit"
    },
    {
      "type": "button",
      "label": "Delete",
      "actionType": "ajax",
      "api": "delete:/api/finance/transactions/${id}",
      "level": "danger",
      "icon": "fa fa-trash",
      "confirmText": "Delete this transaction? This cannot be undone.",
      "visibleOn": "data.canDelete"
    }
  ],
  
  "body": [
    {
      "type": "panel",
      "title": "Transaction Information",
      "body": [
        {
          "type": "grid",
          "columns": [
            {
              "body": [
                {
                  "type": "static",
                  "label": "Transaction ID",
                  "name": "id",
                  "tpl": "TRX-${id}"
                },
                {
                  "type": "static-date",
                  "label": "Date",
                  "name": "date",
                  "format": "YYYY-MM-DD"
                },
                {
                  "type": "static",
                  "label": "Amount",
                  "name": "amount",
                  "tpl": "$${amount|number:2}"
                },
                {
                  "type": "static",
                  "label": "Status",
                  "name": "status",
                  "tpl": "<span class='label label-${status === \"completed\" ? \"success\" : status === \"pending\" ? \"warning\" : \"danger\"}'>${status}</span>"
                }
              ]
            },
            {
              "body": [
                {
                  "type": "static",
                  "label": "Category",
                  "name": "category"
                },
                {
                  "type": "static",
                  "label": "Description",
                  "name": "description"
                },
                {
                  "type": "static",
                  "label": "Created By",
                  "name": "created_by"
                },
                {
                  "type": "static-datetime",
                  "label": "Created At",
                  "name": "created_at",
                  "format": "YYYY-MM-DD HH:mm:ss"
                }
              ]
            }
          ]
        }
      ]
    },
    {
      "type": "panel",
      "title": "Line Items",
      "body": [
        {
          "type": "table",
          "source": "${line_items}",
          "columns": [
            {"name": "product", "label": "Product"},
            {"name": "quantity", "label": "Qty"},
            {"name": "unit_price", "label": "Unit Price", "tpl": "$${unit_price|number:2}"},
            {"name": "total", "label": "Total", "tpl": "$${total|number:2}"}
          ]
        }
      ]
    },
    {
      "type": "panel",
      "title": "Notes",
      "body": [
        {
          "type": "static",
          "name": "notes",
          "tpl": "${notes || 'No notes'}"
        }
      ]
    }
  ]
}
```

### 3. Form Page (Create/Edit)

**GET /finance/transactions/new**

```json
{
  "type": "page",
  "title": "Create Transaction",
  
  "body": [
    {
      "type": "form",
      "api": "post:/api/finance/transactions",
      "redirect": "/finance/transactions/${id}",
      "mode": "horizontal",
      
      "body": [
        {
          "type": "panel",
          "title": "Basic Information",
          "body": [
            {
              "type": "input-date",
              "name": "date",
              "label": "Date",
              "required": true,
              "value": "${NOW()}"
            },
            {
              "type": "input-number",
              "name": "amount",
              "label": "Amount",
              "required": true,
              "min": 0,
              "precision": 2,
              "prefix": "$"
            },
            {
              "type": "select",
              "name": "category",
              "label": "Category",
              "required": true,
              "source": "/api/finance/categories",
              "searchable": true
            },
            {
              "type": "textarea",
              "name": "description",
              "label": "Description",
              "required": true,
              "maxLength": 500
            }
          ]
        },
        {
          "type": "panel",
          "title": "Line Items",
          "body": [
            {
              "type": "input-table",
              "name": "line_items",
              "label": false,
              "addable": true,
              "removable": true,
              "columns": [
                {
                  "name": "product",
                  "label": "Product",
                  "type": "select",
                  "source": "/api/products",
                  "required": true
                },
                {
                  "name": "quantity",
                  "label": "Quantity",
                  "type": "input-number",
                  "required": true,
                  "min": 1
                },
                {
                  "name": "unit_price",
                  "label": "Unit Price",
                  "type": "input-number",
                  "required": true,
                  "prefix": "$",
                  "precision": 2
                },
                {
                  "name": "total",
                  "label": "Total",
                  "type": "static",
                  "tpl": "$${quantity * unit_price | number:2}"
                }
              ]
            }
          ]
        },
        {
          "type": "panel",
          "title": "Additional Details",
          "collapsable": true,
          "collapsed": true,
          "body": [
            {
              "type": "textarea",
              "name": "notes",
              "label": "Notes",
              "maxLength": 1000
            },
            {
              "type": "input-tag",
              "name": "tags",
              "label": "Tags",
              "placeholder": "Add tags..."
            }
          ]
        }
      ],
      
      "actions": [
        {
          "type": "submit",
          "label": "Create Transaction",
          "level": "primary"
        },
        {
          "type": "button",
          "label": "Cancel",
          "actionType": "link",
          "link": "/finance/transactions"
        }
      ]
    }
  ]
}
```

---

## Navigation Components

### Breadcrumbs

Added to page automatically based on URL structure or defined manually:

```json
{
  "type": "page",
  "breadcrumb": [
    {"label": "Home", "href": "/"},
    {"label": "Finance", "href": "/finance"},
    {"label": "Transactions", "href": "/finance/transactions"},
    {"label": "TRX-001"}
  ]
}
```

### Sidebar (Aside)

```json
{
  "type": "page",
  "aside": {
    "type": "nav",
    "stacked": true,
    "links": [
      {
        "label": "Finance",
        "icon": "fa fa-dollar",
        "children": [
          {
            "label": "Dashboard",
            "to": "/finance",
            "icon": "fa fa-dashboard"
          },
          {
            "label": "Transactions",
            "to": "/finance/transactions",
            "icon": "fa fa-list",
            "active": true
          },
          {
            "label": "Invoices",
            "to": "/finance/invoices",
            "icon": "fa fa-file-text",
            "badge": {
              "text": "5",
              "level": "danger"
            }
          }
        ]
      },
      {
        "label": "Inventory",
        "icon": "fa fa-cubes",
        "children": [
          {
            "label": "Products",
            "to": "/inventory/products"
          },
          {
            "label": "Stock",
            "to": "/inventory/stock"
          }
        ]
      }
    ]
  }
}
```

---

## Component Layouts (Body)

### Grid Layout

```json
{
  "type": "grid",
  "columns": [
    {
      "body": [
        {"type": "input-text", "name": "first_name", "label": "First Name"}
      ]
    },
    {
      "body": [
        {"type": "input-text", "name": "last_name", "label": "Last Name"}
      ]
    }
  ]
}
```

### Tabs Layout

```json
{
  "type": "tabs",
  "tabs": [
    {
      "title": "General",
      "body": [
        {"type": "input-text", "name": "name", "label": "Name"},
        {"type": "input-email", "name": "email", "label": "Email"}
      ]
    },
    {
      "title": "Address",
      "body": [
        {"type": "input-text", "name": "street", "label": "Street"},
        {"type": "input-text", "name": "city", "label": "City"}
      ]
    }
  ]
}
```

### Panel (Section)

```json
{
  "type": "panel",
  "title": "User Information",
  "collapsable": true,
  "body": [
    {"type": "input-text", "name": "username", "label": "Username"},
    {"type": "input-password", "name": "password", "label": "Password"}
  ]
}
```

### Service (Wrapper with Data Loading)

```json
{
  "type": "service",
  "api": "/api/user/profile",
  "body": [
    {
      "type": "panel",
      "title": "Profile",
      "body": [
        {"type": "static", "label": "Name", "name": "name"},
        {"type": "static", "label": "Email", "name": "email"}
      ]
    }
  ]
}
```

---

## Column Display Types

### Text Types

```json
// Plain text
{"name": "name", "label": "Name", "type": "text"}

// Number
{"name": "quantity", "label": "Qty", "type": "number"}

// Currency
{"name": "amount", "label": "Amount", "type": "number", "prefix": "$", "precision": 2}

// Percentage
{"name": "discount", "label": "Discount", "type": "number", "suffix": "%", "precision": 1}
```

### Date/Time Types

```json
// Date
{"name": "date", "label": "Date", "type": "date", "format": "YYYY-MM-DD"}

// DateTime
{"name": "created_at", "label": "Created", "type": "datetime", "format": "YYYY-MM-DD HH:mm:ss"}

// Time
{"name": "time", "label": "Time", "type": "time", "format": "HH:mm:ss"}

// Relative time
{"name": "updated_at", "label": "Updated", "type": "datetime", "fromNow": true}
```

### Status/Tags

```json
// Tag
{"name": "category", "label": "Category", "type": "tag"}

// Multiple tags
{"name": "tags", "label": "Tags", "type": "each", "items": {"type": "tag", "label": "${item}"}}

// Status mapping
{
  "name": "status",
  "label": "Status",
  "type": "mapping",
  "map": {
    "active": "<span class='label label-success'>Active</span>",
    "inactive": "<span class='label label-default'>Inactive</span>",
    "pending": "<span class='label label-warning'>Pending</span>"
  }
}
```

### Links and Actions

```json
// Link
{
  "name": "id",
  "label": "ID",
  "type": "text",
  "tpl": "<a href='/items/${id}'>${id}</a>"
}

// Button
{
  "type": "button",
  "label": "View Details",
  "actionType": "link",
  "link": "/items/${id}"
}
```

### Images

```json
// Image
{"name": "avatar", "label": "Avatar", "type": "image", "width": 40, "height": 40}

// Images (multiple)
{"name": "photos", "label": "Photos", "type": "images"}
```

### Boolean

```json
// Switch display
{
  "name": "active",
  "label": "Active",
  "type": "mapping",
  "map": {
    "true": "<i class='fa fa-check text-success'></i>",
    "false": "<i class='fa fa-times text-danger'></i>"
  }
}
```

---

## Standard Page Patterns

### Pattern 1: List → Detail → Edit

**1. List** (`/finance/transactions`)
```json
{
  "type": "page",
  "body": [
    {
      "type": "crud",
      "api": "/api/finance/transactions",
      "itemActions": [
        {"label": "View", "actionType": "link", "link": "/finance/transactions/${id}"},
        {"label": "Edit", "actionType": "link", "link": "/finance/transactions/${id}/edit"}
      ]
    }
  ]
}
```

**2. Detail** (`/finance/transactions/{id}`)
```json
{
  "type": "page",
  "initApi": "/api/finance/transactions/${id}",
  "toolbar": [
    {"label": "Edit", "actionType": "link", "link": "/finance/transactions/${id}/edit"}
  ],
  "body": [
    {"type": "panel", "body": [...]}
  ]
}
```

**3. Edit** (`/finance/transactions/{id}/edit`)
```json
{
  "type": "page",
  "body": [
    {
      "type": "form",
      "initApi": "/api/finance/transactions/${id}",
      "api": "put:/api/finance/transactions/${id}",
      "body": [...]
    }
  ]
}
```

### Pattern 2: Dashboard

```json
{
  "type": "page",
  "title": "Finance Dashboard",
  "body": [
    {
      "type": "grid",
      "columns": [
        {
          "body": [
            {"type": "card", "header": {"title": "Revenue"}, "body": "$125,430"}
          ]
        },
        {
          "body": [
            {"type": "card", "header": {"title": "Transactions"}, "body": "1,234"}
          ]
        }
      ]
    },
    {
      "type": "crud",
      "title": "Recent Transactions",
      "api": "/api/finance/transactions?limit=10"
    }
  ]
}
```

### Pattern 3: Wizard (Multi-Step)

```json
{
  "type": "page",
  "body": [
    {
      "type": "wizard",
      "api": "/api/users",
      "steps": [
        {
          "title": "Account",
          "body": [
            {"type": "input-email", "name": "email", "required": true},
            {"type": "input-password", "name": "password", "required": true}
          ]
        },
        {
          "title": "Profile",
          "body": [
            {"type": "input-text", "name": "first_name", "required": true},
            {"type": "input-text", "name": "last_name", "required": true}
          ]
        },
        {
          "title": "Review",
          "body": [
            {"type": "static", "label": "Email", "name": "email"},
            {"type": "static", "label": "Name", "tpl": "${first_name} ${last_name}"}
          ]
        }
      ]
    }
  ]
}
```

---

## Responsive Behavior

Pages and components automatically adapt to screen size:

```json
{
  "type": "grid",
  "columns": [
    {"md": 6, "body": [...]},  // 50% on desktop
    {"md": 6, "body": [...]}   // 50% on desktop
  ]
}
```

Breakpoints:
- `xs`: < 576px (mobile)
- `sm`: 576px - 768px (tablet)
- `md`: 768px - 992px (small desktop)
- `lg`: 992px - 1200px (desktop)
- `xl`: > 1200px (large desktop)

---

## Summary

**Page structure:**
```
{
  "type": "page",        ← Top-level type
  "title": "...",
  "aside": {...},        ← Sidebar
  "toolbar": [...],      ← Action buttons
  "body": [...]          ← Components (crud, form, etc.)
}
```

**Key points:**
- ✅ `page` is always top-level type
- ✅ Each JSON file = 1 page OR 1 component
- ✅ `body` contains components (crud, form, panel, etc.)
- ✅ CRUD component = data table with actions
- ✅ Forms are components inside page body
- ✅ Layout types (grid, tabs, panel) organize body content

**Common components:**
- `crud` - Data table with CRUD operations
- `form` - Form with fields and submission
- `panel` - Collapsible section
- `grid` - Column layout
- `tabs` - Tabbed interface
- `wizard` - Multi-step form
- `service` - Data loader wrapper

[← Back](06-validation.md) | [Next: Conditional Logic →](08-conditional-logic.md)
