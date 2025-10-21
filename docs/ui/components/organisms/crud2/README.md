# CRUD2 Component

**FILE PURPOSE**: Modern CRUD (Create, Read, Update, Delete) interface for data management  
**SCOPE**: Data tables, cards, lists with full CRUD operations, filtering, sorting, and pagination  
**TARGET AUDIENCE**: Developers implementing data management interfaces, business entity management, and administrative panels

## 📋 Component Overview

CRUD2 provides a modern, flexible interface for managing data entities with support for multiple view modes (table, cards, list), advanced filtering, bulk operations, and real-time updates. Essential for all data management needs in ERP systems.

### Schema Reference
- **Primary Schema**: `CRUD2Schema.json`
- **View Schemas**: `CRUD2TableSchema.json`, `CRUD2CardsSchema.json`, `CRUD2ListSchema.json`
- **Related Schemas**: `ColumnSchema.json`, `ActionSchema.json`, `RowSelectionSchema.json`
- **Base Interface**: Complete data management organism

## Basic Usage

```json
{
    "type": "crud2",
    "api": "/api/customers",
    "columns": [
        {"name": "name", "label": "Name"},
        {"name": "email", "label": "Email"},
        {"name": "status", "label": "Status"}
    ]
}
```

## Go Type Definition

```go
type CRUD2Props struct {
    Type                        string              `json:"type"`
    Mode                        string              `json:"mode"`                // "table2", "cards", "list"
    API                         interface{}         `json:"api"`                 // Data source API
    Source                      string              `json:"source"`              // Alternative data source
    Title                       interface{}         `json:"title"`               // CRUD title
    
    // Table Mode Properties
    Columns                     []interface{}       `json:"columns"`             // Table columns
    ColumnsTogglable           interface{}         `json:"columnsTogglable"`    // Column visibility toggle
    RowSelection               interface{}         `json:"rowSelection"`        // Row selection config
    Expandable                 interface{}         `json:"expandable"`          // Row expansion
    Sticky                     bool                `json:"sticky"`              // Sticky header
    Bordered                   bool                `json:"bordered"`            // Table borders
    ShowHeader                 bool                `json:"showHeader"`          // Header visibility
    TableLayout                string              `json:"tableLayout"`         // "fixed" or "auto"
    
    // Data Management
    KeyField                   string              `json:"keyField"`            // Primary key field
    PrimaryField               string              `json:"primaryField"`        // Row identifier
    ChildrenColumnName         string              `json:"childrenColumnName"`  // Nested data field
    LoadType                   string              `json:"loadType"`            // "pagination" or "more"
    PerPage                    int                 `json:"perPage"`             // Items per page
    LoadDataOnce               bool                `json:"loadDataOnce"`        // Frontend pagination
    
    // Actions and Operations
    HeaderToolbar              interface{}         `json:"headerToolbar"`       // Top toolbar
    FooterToolbar              interface{}         `json:"footerToolbar"`       // Bottom toolbar
    Actions                    []interface{}       `json:"actions"`             // Row actions
    QuickSaveAPI               interface{}         `json:"quickSaveApi"`        // Batch save API
    QuickSaveItemAPI           interface{}         `json:"quickSaveItemApi"`    // Single item save
    SaveOrderAPI               interface{}         `json:"saveOrderApi"`        // Sort save API
    
    // Selection and Bulk Operations
    Selectable                 bool                `json:"selectable"`          // Enable selection
    Multiple                   bool                `json:"multiple"`            // Multi-selection
    ShowSelection              bool                `json:"showSelection"`       // Show selection area
    MaxKeepItemSelectionLength int                 `json:"maxKeepItemSelectionLength"` // Max selected items
    KeepItemSelectionOnPageChange bool             `json:"keepItemSelectionOnPageChange"` // Persist selection
    
    // Real-time Updates
    Interval                   int                 `json:"interval"`            // Auto-refresh interval
    SilentPolling              bool                `json:"silentPolling"`       // Silent refresh
    StopAutoRefreshWhen        string              `json:"stopAutoRefreshWhen"` // Stop condition
    
    // URL Synchronization
    SyncLocation               bool                `json:"syncLocation"`        // Sync with URL
    SyncResponse2Query         bool                `json:"syncResponse2Query"`  // Sync response to URL
    PageField                  string              `json:"pageField"`           // Page parameter name
    PerPageField               string              `json:"perPageField"`        // Per-page parameter name
    
    // Performance
    LazyRenderAfter            int                 `json:"lazyRenderAfter"`     // Lazy render threshold
    AutoFillHeight             bool                `json:"autoFillHeight"`      // Fill available height
    CanAccessSuperData         bool                `json:"canAccessSuperData"`  // Access parent data
    
    // UI Customization
    Loading                    interface{}         `json:"loading"`             // Loading state
    ItemBadge                  interface{}         `json:"itemBadge"`          // Row badges
    ShowBadge                  bool                `json:"showBadge"`          // Badge visibility
    RowClassNameExpr           string              `json:"rowClassNameExpr"`   // Custom row CSS
    LineHeight                 string              `json:"lineHeight"`         // Row height
    Footer                     interface{}         `json:"footer"`             // Table footer
}
```

## View Modes

### Table Mode (Default)
```json
{
    "type": "crud2",
    "mode": "table2",
    "api": "/api/employees",
    "title": "Employee Management",
    "columns": [
        {"name": "id", "label": "ID", "width": 80},
        {"name": "name", "label": "Name", "sortable": true},
        {"name": "department", "label": "Department"},
        {"name": "position", "label": "Position"},
        {"name": "status", "label": "Status", "type": "status"}
    ],
    "rowSelection": true,
    "bordered": true,
    "sticky": true
}
```

### Cards Mode
```json
{
    "type": "crud2",
    "mode": "cards",
    "api": "/api/products",
    "title": "Product Catalog",
    "card": {
        "header": {
            "title": "${name}",
            "subTitle": "SKU: ${sku}"
        },
        "body": [
            {"type": "image", "src": "${image_url}"},
            {"type": "text", "text": "${description}"},
            {"type": "text", "text": "Price: $${price}"}
        ]
    }
}
```

### List Mode
```json
{
    "type": "crud2",
    "mode": "list",
    "api": "/api/orders",
    "title": "Order List",
    "listItem": {
        "title": "Order #${order_number}",
        "subTitle": "Customer: ${customer_name}",
        "desc": "Total: $${total_amount}",
        "avatar": {"type": "icon", "icon": "shopping-cart"}
    }
}
```

## Real-World Use Cases

### Customer Management Interface
```json
{
    "type": "crud2",
    "mode": "table2",
    "api": "/api/customers",
    "title": "Customer Management",
    "headerToolbar": [
        {
            "type": "button",
            "label": "Add Customer",
            "actionType": "dialog",
            "icon": "fa fa-plus",
            "level": "primary",
            "dialog": {
                "title": "Add New Customer",
                "body": {
                    "type": "form",
                    "api": "post:/api/customers",
                    "body": [
                        {"type": "input-text", "name": "company_name", "label": "Company Name", "required": true},
                        {"type": "input-email", "name": "email", "label": "Email", "required": true},
                        {"type": "input-text", "name": "phone", "label": "Phone"},
                        {"type": "select", "name": "status", "label": "Status", "options": ["active", "inactive", "prospect"]}
                    ]
                }
            }
        },
        {
            "type": "dropdown-button",
            "label": "Export",
            "buttons": [
                {"type": "button", "label": "Export CSV", "actionType": "download", "api": "/api/customers/export?format=csv"},
                {"type": "button", "label": "Export Excel", "actionType": "download", "api": "/api/customers/export?format=xlsx"}
            ]
        },
        "filter-toggler",
        "reload"
    ],
    "filter": {
        "body": [
            {"type": "input-text", "name": "company_name", "label": "Company Name", "placeholder": "Search by company name"},
            {"type": "select", "name": "status", "label": "Status", "options": ["active", "inactive", "prospect"], "clearable": true},
            {"type": "select", "name": "industry", "label": "Industry", "source": "/api/industries", "clearable": true},
            {"type": "input-date-range", "name": "created_date", "label": "Created Date Range"}
        ]
    },
    "columns": [
        {
            "name": "id",
            "label": "ID",
            "width": 80,
            "sortable": true
        },
        {
            "name": "company_name",
            "label": "Company",
            "sortable": true,
            "searchable": true,
            "type": "link",
            "href": "/customers/${id}"
        },
        {
            "name": "contact_person",
            "label": "Contact Person"
        },
        {
            "name": "email",
            "label": "Email",
            "type": "email"
        },
        {
            "name": "phone",
            "label": "Phone",
            "type": "phone"
        },
        {
            "name": "industry",
            "label": "Industry"
        },
        {
            "name": "status",
            "label": "Status",
            "type": "status",
            "map": {
                "active": "<span class='label label-success'>Active</span>",
                "inactive": "<span class='label label-default'>Inactive</span>",
                "prospect": "<span class='label label-info'>Prospect</span>"
            }
        },
        {
            "name": "created_date",
            "label": "Created",
            "type": "date",
            "format": "MMM DD, YYYY"
        },
        {
            "type": "operation",
            "label": "Actions",
            "width": 150,
            "buttons": [
                {
                    "type": "button",
                    "label": "View",
                    "actionType": "dialog",
                    "level": "link",
                    "dialog": {
                        "title": "Customer Details",
                        "size": "lg",
                        "body": {
                            "type": "service",
                            "api": "/api/customers/${id}",
                            "body": [
                                {"type": "static", "name": "company_name", "label": "Company Name"},
                                {"type": "static", "name": "contact_person", "label": "Contact Person"},
                                {"type": "static", "name": "email", "label": "Email"},
                                {"type": "static", "name": "phone", "label": "Phone"},
                                {"type": "static", "name": "address", "label": "Address"},
                                {"type": "static", "name": "industry", "label": "Industry"},
                                {"type": "static", "name": "status", "label": "Status"}
                            ]
                        }
                    }
                },
                {
                    "type": "button",
                    "label": "Edit",
                    "actionType": "dialog",
                    "level": "primary",
                    "dialog": {
                        "title": "Edit Customer",
                        "body": {
                            "type": "form",
                            "api": "put:/api/customers/${id}",
                            "initApi": "/api/customers/${id}",
                            "body": [
                                {"type": "input-text", "name": "company_name", "label": "Company Name", "required": true},
                                {"type": "input-text", "name": "contact_person", "label": "Contact Person"},
                                {"type": "input-email", "name": "email", "label": "Email", "required": true},
                                {"type": "input-text", "name": "phone", "label": "Phone"},
                                {"type": "textarea", "name": "address", "label": "Address"},
                                {"type": "select", "name": "industry", "label": "Industry", "source": "/api/industries"},
                                {"type": "select", "name": "status", "label": "Status", "options": ["active", "inactive", "prospect"]}
                            ]
                        }
                    }
                },
                {
                    "type": "button",
                    "label": "Delete",
                    "actionType": "ajax",
                    "level": "danger",
                    "confirmText": "Are you sure you want to delete this customer?",
                    "api": "delete:/api/customers/${id}"
                }
            ]
        }
    ],
    "rowSelection": true,
    "multiple": true,
    "footerToolbar": [
        {
            "type": "bulk-actions",
            "align": "left",
            "buttons": [
                {
                    "type": "button",
                    "label": "Bulk Delete",
                    "actionType": "ajax",
                    "api": "delete:/api/customers/bulk",
                    "confirmText": "Are you sure you want to delete selected customers?",
                    "level": "danger"
                },
                {
                    "type": "button",
                    "label": "Bulk Status Update",
                    "actionType": "dialog",
                    "dialog": {
                        "title": "Update Status",
                        "body": {
                            "type": "form",
                            "api": "put:/api/customers/bulk-status",
                            "body": [
                                {"type": "select", "name": "status", "label": "New Status", "options": ["active", "inactive", "prospect"], "required": true}
                            ]
                        }
                    }
                }
            ]
        },
        "statistics",
        "pagination"
    ],
    "perPage": 20,
    "loadType": "pagination",
    "syncLocation": true,
    "autoFillHeight": true,
    "columnsTogglable": "auto"
}
```

### Product Catalog (Cards Mode)
```json
{
    "type": "crud2",
    "mode": "cards",
    "api": "/api/products",
    "title": "Product Catalog",
    "headerToolbar": [
        {
            "type": "button",
            "label": "Add Product",
            "actionType": "dialog",
            "icon": "fa fa-plus",
            "level": "primary",
            "dialog": {
                "title": "Add New Product",
                "size": "lg",
                "body": {
                    "type": "form",
                    "api": "post:/api/products",
                    "body": [
                        {"type": "input-text", "name": "name", "label": "Product Name", "required": true},
                        {"type": "input-text", "name": "sku", "label": "SKU", "required": true},
                        {"type": "textarea", "name": "description", "label": "Description"},
                        {"type": "input-number", "name": "price", "label": "Price", "prefix": "$", "min": 0, "step": 0.01},
                        {"type": "select", "name": "category_id", "label": "Category", "source": "/api/categories"},
                        {"type": "input-image", "name": "image", "label": "Product Image"}
                    ]
                }
            }
        },
        "filter-toggler",
        "reload"
    ],
    "filter": {
        "body": [
            {"type": "input-text", "name": "name", "label": "Product Name"},
            {"type": "select", "name": "category_id", "label": "Category", "source": "/api/categories", "clearable": true},
            {"type": "range", "name": "price_range", "label": "Price Range", "min": 0, "max": 1000, "step": 10}
        ]
    },
    "card": {
        "header": {
            "title": "${name}",
            "subTitle": "SKU: ${sku}",
            "avatar": {
                "type": "image",
                "src": "${image_url}",
                "className": "w-16 h-16 object-cover rounded"
            }
        },
        "body": [
            {"type": "text", "text": "${description}", "className": "text-gray-600 mb-2"},
            {"type": "text", "text": "Price: $${price}", "className": "font-bold text-lg text-green-600"},
            {"type": "text", "text": "Category: ${category_name}", "className": "text-sm text-gray-500"},
            {"type": "text", "text": "Stock: ${stock_quantity}", "className": "text-sm", "visibleOn": "${stock_quantity > 0}"}
        ],
        "actions": [
            {
                "type": "button",
                "label": "Edit",
                "actionType": "dialog",
                "level": "primary",
                "size": "sm"
            },
            {
                "type": "button",
                "label": "Delete",
                "actionType": "ajax",
                "level": "danger",
                "size": "sm",
                "confirmText": "Delete this product?"
            }
        ]
    },
    "loadType": "more",
    "perPage": 12,
    "autoFillHeight": true
}
```

### Order Management (List Mode)
```json
{
    "type": "crud2",
    "mode": "list",
    "api": "/api/orders",
    "title": "Order Management",
    "headerToolbar": [
        {
            "type": "button",
            "label": "New Order",
            "actionType": "link",
            "link": "/orders/create",
            "icon": "fa fa-plus",
            "level": "primary"
        },
        "filter-toggler",
        "reload"
    ],
    "filter": {
        "body": [
            {"type": "input-text", "name": "order_number", "label": "Order Number"},
            {"type": "input-text", "name": "customer_name", "label": "Customer Name"},
            {"type": "select", "name": "status", "label": "Status", "options": ["pending", "processing", "shipped", "delivered", "cancelled"]},
            {"type": "input-date-range", "name": "order_date", "label": "Order Date Range"}
        ]
    },
    "listItem": {
        "title": "Order #${order_number}",
        "subTitle": "Customer: ${customer_name} • ${order_date}",
        "desc": "Total: $${total_amount} • Items: ${item_count}",
        "avatar": {
            "type": "icon",
            "icon": "shopping-cart",
            "className": "bg-blue-500 text-white"
        },
        "actions": [
            {
                "type": "button",
                "label": "View",
                "actionType": "link",
                "link": "/orders/${id}",
                "level": "link"
            },
            {
                "type": "dropdown-button",
                "label": "Actions",
                "buttons": [
                    {"type": "button", "label": "Process", "actionType": "ajax", "api": "put:/api/orders/${id}/process"},
                    {"type": "button", "label": "Ship", "actionType": "ajax", "api": "put:/api/orders/${id}/ship"},
                    {"type": "button", "label": "Cancel", "actionType": "ajax", "api": "put:/api/orders/${id}/cancel", "confirmText": "Cancel this order?"}
                ]
            }
        ],
        "remark": {
            "type": "status",
            "name": "status",
            "map": {
                "pending": "<span class='label label-warning'>Pending</span>",
                "processing": "<span class='label label-info'>Processing</span>",
                "shipped": "<span class='label label-primary'>Shipped</span>",
                "delivered": "<span class='label label-success'>Delivered</span>",
                "cancelled": "<span class='label label-danger'>Cancelled</span>"
            }
        }
    },
    "perPage": 15,
    "loadType": "pagination",
    "autoFillHeight": true
}
```

### Inventory Management with Quick Edit
```json
{
    "type": "crud2",
    "mode": "table2",
    "api": "/api/inventory",
    "title": "Inventory Management",
    "headerToolbar": [
        {
            "type": "button",
            "label": "Add Item",
            "actionType": "dialog",
            "icon": "fa fa-plus",
            "level": "primary"
        },
        {
            "type": "button",
            "label": "Bulk Update Stock",
            "actionType": "dialog",
            "dialog": {
                "title": "Bulk Stock Update",
                "body": {
                    "type": "form",
                    "api": "put:/api/inventory/bulk-update",
                    "body": [
                        {"type": "input-file", "name": "csv_file", "label": "Upload CSV", "accept": ".csv"},
                        {"type": "static", "value": "Download template: <a href='/templates/inventory-update.csv'>inventory-update.csv</a>"}
                    ]
                }
            }
        },
        "filter-toggler",
        "reload"
    ],
    "filter": {
        "body": [
            {"type": "input-text", "name": "item_name", "label": "Item Name"},
            {"type": "input-text", "name": "sku", "label": "SKU"},
            {"type": "select", "name": "category", "label": "Category", "source": "/api/categories"},
            {"type": "select", "name": "location", "label": "Location", "source": "/api/locations"},
            {"type": "select", "name": "stock_status", "label": "Stock Status", "options": ["in_stock", "low_stock", "out_of_stock"]}
        ]
    },
    "columns": [
        {"name": "sku", "label": "SKU", "width": 120, "sortable": true},
        {"name": "item_name", "label": "Item Name", "sortable": true},
        {"name": "category", "label": "Category"},
        {"name": "location", "label": "Location"},
        {
            "name": "current_stock",
            "label": "Current Stock",
            "type": "input-number",
            "quickEdit": true,
            "quickSaveItemApi": "put:/api/inventory/${id}/stock"
        },
        {
            "name": "reorder_level",
            "label": "Reorder Level",
            "type": "input-number",
            "quickEdit": true
        },
        {
            "name": "unit_cost",
            "label": "Unit Cost",
            "type": "input-number",
            "prefix": "$",
            "quickEdit": true
        },
        {
            "name": "stock_status",
            "label": "Status",
            "type": "status",
            "map": {
                "in_stock": "<span class='label label-success'>In Stock</span>",
                "low_stock": "<span class='label label-warning'>Low Stock</span>",
                "out_of_stock": "<span class='label label-danger'>Out of Stock</span>"
            }
        },
        {"name": "last_updated", "label": "Last Updated", "type": "datetime"}
    ],
    "quickSaveApi": "put:/api/inventory/bulk-save",
    "quickSaveItemApi": "put:/api/inventory/${id}",
    "perPage": 25,
    "autoFillHeight": true,
    "interval": 30000,
    "silentPolling": true
}
```

This component provides the essential CRUD functionality for ERP systems requiring comprehensive data management, multiple view modes, and advanced features like bulk operations, real-time updates, and flexible filtering.