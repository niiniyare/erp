# Table Component

**FILE PURPOSE**: Data table display organism for structured data presentation  
**SCOPE**: Data tables, column management, sorting, row operations, and data visualization  
**TARGET AUDIENCE**: Developers implementing data display tables, reporting interfaces, and structured data presentation

## 📋 Component Overview

Table provides comprehensive data table functionality with column configuration, sorting, selection, row operations, and responsive design. Essential for displaying structured data in ERP systems without the full CRUD functionality.

### Schema Reference
- **Primary Schema**: `TableSchema.json`
- **Related Schemas**: `TableColumn.json`, `TableColumnObject.json`, `TableColumnWithType.json`
- **Base Interface**: Data display organism for tabular information

## Basic Usage

```json
{
    "type": "table",
    "source": "${items}",
    "columns": [
        {"name": "name", "label": "Name"},
        {"name": "email", "label": "Email"},
        {"name": "status", "label": "Status"}
    ]
}
```

## Go Type Definition

```go
type TableProps struct {
    Type                        string              `json:"type"`
    Source                      interface{}         `json:"source"`              // Data source
    Title                       interface{}         `json:"title"`               // Table title
    
    // Column Configuration
    Columns                     []interface{}       `json:"columns"`             // Column definitions
    ColumnsTogglable           interface{}         `json:"columnsTogglable"`    // Column visibility
    
    // Selection
    Selectable                 bool                `json:"selectable"`          // Enable selection
    Multiple                   bool                `json:"multiple"`            // Multi-selection
    CheckOnItemClick           bool                `json:"checkOnItemClick"`    // Click to select
    
    // Display Options
    ShowHeader                 bool                `json:"showHeader"`          // Header visibility
    ShowFooter                 bool                `json:"showFooter"`          // Footer visibility
    Bordered                   bool                `json:"bordered"`            // Table borders
    Striped                    bool                `json:"striped"`             // Striped rows
    Size                       string              `json:"size"`                // Table size
    
    // Row Configuration
    RowClassName               string              `json:"rowClassName"`        // Row CSS classes
    RowClassNameExpr           string              `json:"rowClassNameExpr"`    // Dynamic row CSS
    LineHeight                 string              `json:"lineHeight"`          // Row height
    
    // Sorting
    OrderBy                    string              `json:"orderBy"`             // Default sort field
    OrderDir                   string              `json:"orderDir"`            // Sort direction
    
    // Data Management
    KeyField                   string              `json:"keyField"`            // Primary key field
    ValueField                 string              `json:"valueField"`          // Value field
    LabelField                 string              `json:"labelField"`          // Label field
    
    // Placeholder and Loading
    Placeholder                string              `json:"placeholder"`         // Empty data message
    Loading                    interface{}         `json:"loading"`             // Loading state
    
    // Layout
    AutoFillHeight             bool                `json:"autoFillHeight"`      // Fill height
    Responsive                 bool                `json:"responsive"`          // Responsive design
    Sticky                     bool                `json:"sticky"`              // Sticky header
    TableLayout                string              `json:"tableLayout"`         // Layout mode
    
    // Footer
    Footer                     interface{}         `json:"footer"`              // Table footer
    FooterClassName            string              `json:"footerClassName"`     // Footer CSS
    
    // Performance
    LazyRenderAfter            int                 `json:"lazyRenderAfter"`     // Lazy render threshold
    
    // Nested Data
    ChildrenColumnName         string              `json:"childrenColumnName"`  // Child data field
    Expandable                 bool                `json:"expandable"`          // Row expansion
    
    // Styling
    TableClassName             string              `json:"tableClassName"`      // Table CSS
    HeaderClassName            string              `json:"headerClassName"`     // Header CSS
    BodyClassName              string              `json:"bodyClassName"`       // Body CSS
}
```

## Column Types

### Basic Text Column
```json
{
    "name": "name",
    "label": "Name",
    "sortable": true,
    "searchable": true,
    "width": 200
}
```

### Status Column
```json
{
    "name": "status",
    "label": "Status",
    "type": "status",
    "map": {
        "active": "<span class='label label-success'>Active</span>",
        "inactive": "<span class='label label-default'>Inactive</span>",
        "pending": "<span class='label label-warning'>Pending</span>"
    }
}
```

### Date Column
```json
{
    "name": "created_date",
    "label": "Created",
    "type": "date",
    "format": "YYYY-MM-DD",
    "sortable": true
}
```

### Link Column
```json
{
    "name": "company_name",
    "label": "Company",
    "type": "link",
    "href": "/customers/${id}",
    "blank": false
}
```

### Operation Column
```json
{
    "type": "operation",
    "label": "Actions",
    "width": 150,
    "buttons": [
        {
            "type": "button",
            "label": "Edit",
            "actionType": "dialog",
            "level": "primary"
        },
        {
            "type": "button",
            "label": "Delete",
            "actionType": "ajax",
            "level": "danger",
            "confirmText": "Are you sure?"
        }
    ]
}
```

## Real-World Use Cases

### Employee Directory Table
```json
{
    "type": "table",
    "title": "Employee Directory",
    "source": "/api/employees",
    "columns": [
        {
            "name": "employee_id",
            "label": "ID",
            "width": 80,
            "sortable": true
        },
        {
            "name": "avatar",
            "label": "",
            "type": "image",
            "width": 60,
            "className": "w-10 h-10 rounded-full"
        },
        {
            "name": "full_name",
            "label": "Name",
            "sortable": true,
            "searchable": true,
            "type": "link",
            "href": "/employees/${id}"
        },
        {
            "name": "email",
            "label": "Email",
            "type": "email",
            "copyable": true
        },
        {
            "name": "department",
            "label": "Department",
            "sortable": true
        },
        {
            "name": "position",
            "label": "Position"
        },
        {
            "name": "hire_date",
            "label": "Hire Date",
            "type": "date",
            "format": "MMM DD, YYYY",
            "sortable": true
        },
        {
            "name": "status",
            "label": "Status",
            "type": "status",
            "map": {
                "active": "<span class='label label-success'>Active</span>",
                "inactive": "<span class='label label-default'>Inactive</span>",
                "on_leave": "<span class='label label-warning'>On Leave</span>"
            }
        },
        {
            "type": "operation",
            "label": "Actions",
            "width": 120,
            "buttons": [
                {
                    "type": "button",
                    "label": "View",
                    "actionType": "link",
                    "link": "/employees/${id}",
                    "level": "link",
                    "size": "sm"
                },
                {
                    "type": "button",
                    "label": "Edit",
                    "actionType": "dialog",
                    "level": "primary",
                    "size": "sm",
                    "dialog": {
                        "title": "Edit Employee",
                        "body": {
                            "type": "form",
                            "api": "put:/api/employees/${id}",
                            "initApi": "/api/employees/${id}"
                        }
                    }
                }
            ]
        }
    ],
    "selectable": true,
    "multiple": true,
    "bordered": true,
    "striped": true,
    "sticky": true,
    "autoFillHeight": true,
    "columnsTogglable": "auto",
    "placeholder": "No employees found",
    "orderBy": "full_name",
    "orderDir": "asc"
}
```

### Product Inventory Table
```json
{
    "type": "table",
    "title": "Product Inventory",
    "source": "/api/inventory",
    "columns": [
        {
            "name": "sku",
            "label": "SKU",
            "width": 120,
            "sortable": true,
            "fixed": "left"
        },
        {
            "name": "product_image",
            "label": "Image",
            "type": "image",
            "width": 80,
            "enlargeAble": true,
            "thumbMode": "cover",
            "thumbRatio": "1:1"
        },
        {
            "name": "product_name",
            "label": "Product",
            "sortable": true,
            "searchable": true,
            "type": "link",
            "href": "/products/${product_id}"
        },
        {
            "name": "category",
            "label": "Category",
            "sortable": true
        },
        {
            "name": "current_stock",
            "label": "Current Stock",
            "type": "number",
            "sortable": true,
            "classNameExpr": "${current_stock <= reorder_level ? 'text-red-600 font-bold' : 'text-green-600'}"
        },
        {
            "name": "reorder_level",
            "label": "Reorder Level",
            "type": "number"
        },
        {
            "name": "unit_cost",
            "label": "Unit Cost",
            "type": "number",
            "prefix": "$",
            "precision": 2
        },
        {
            "name": "total_value",
            "label": "Total Value",
            "type": "tpl",
            "tpl": "$${current_stock * unit_cost | number:2}",
            "sortable": true
        },
        {
            "name": "location",
            "label": "Location",
            "sortable": true
        },
        {
            "name": "stock_status",
            "label": "Status",
            "type": "mapping",
            "map": {
                "in_stock": {
                    "type": "status",
                    "value": "In Stock",
                    "className": "label-success"
                },
                "low_stock": {
                    "type": "status", 
                    "value": "Low Stock",
                    "className": "label-warning"
                },
                "out_of_stock": {
                    "type": "status",
                    "value": "Out of Stock", 
                    "className": "label-danger"
                }
            }
        },
        {
            "name": "last_updated",
            "label": "Last Updated",
            "type": "datetime",
            "format": "MMM DD, HH:mm",
            "sortable": true
        }
    ],
    "bordered": true,
    "sticky": true,
    "autoFillHeight": true,
    "columnsTogglable": true,
    "size": "sm",
    "rowClassNameExpr": "${current_stock <= reorder_level ? 'bg-red-50' : ''}",
    "footer": {
        "type": "flex",
        "items": [
            {
                "type": "tpl",
                "tpl": "Total Products: ${items | length}",
                "className": "text-sm text-gray-600"
            },
            {
                "type": "tpl", 
                "tpl": "Low Stock Items: ${items | filter:current_stock:lte:reorder_level | length}",
                "className": "text-sm text-orange-600"
            },
            {
                "type": "tpl",
                "tpl": "Total Value: $${items | sum:total_value | number:2}",
                "className": "text-sm font-bold text-green-600"
            }
        ]
    }
}
```

### Order History Table
```json
{
    "type": "table",
    "title": "Order History",
    "source": "/api/orders",
    "columns": [
        {
            "name": "order_number",
            "label": "Order #",
            "width": 120,
            "sortable": true,
            "type": "link",
            "href": "/orders/${id}"
        },
        {
            "name": "customer_name",
            "label": "Customer",
            "sortable": true,
            "searchable": true
        },
        {
            "name": "order_date",
            "label": "Order Date",
            "type": "date",
            "format": "MMM DD, YYYY",
            "sortable": true
        },
        {
            "name": "items",
            "label": "Items",
            "type": "list",
            "listItem": {
                "title": "${product_name}",
                "subTitle": "Qty: ${quantity} × $${unit_price}"
            },
            "maxDisplayedItems": 3,
            "placeholder": "No items"
        },
        {
            "name": "total_amount",
            "label": "Total",
            "type": "number",
            "prefix": "$",
            "precision": 2,
            "sortable": true
        },
        {
            "name": "payment_status",
            "label": "Payment",
            "type": "status",
            "map": {
                "paid": "<span class='label label-success'>Paid</span>",
                "pending": "<span class='label label-warning'>Pending</span>",
                "failed": "<span class='label label-danger'>Failed</span>",
                "refunded": "<span class='label label-info'>Refunded</span>"
            }
        },
        {
            "name": "order_status",
            "label": "Status",
            "type": "status",
            "map": {
                "pending": "<span class='label label-warning'>Pending</span>",
                "processing": "<span class='label label-info'>Processing</span>",
                "shipped": "<span class='label label-primary'>Shipped</span>",
                "delivered": "<span class='label label-success'>Delivered</span>",
                "cancelled": "<span class='label label-danger'>Cancelled</span>"
            }
        },
        {
            "type": "operation",
            "label": "Actions",
            "width": 150,
            "buttons": [
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
                    "level": "primary",
                    "buttons": [
                        {
                            "type": "button",
                            "label": "Print Invoice",
                            "actionType": "download",
                            "api": "/api/orders/${id}/invoice.pdf"
                        },
                        {
                            "type": "button",
                            "label": "Send Email",
                            "actionType": "ajax",
                            "api": "post:/api/orders/${id}/send-email"
                        },
                        {
                            "type": "button",
                            "label": "Cancel Order",
                            "actionType": "ajax",
                            "api": "put:/api/orders/${id}/cancel",
                            "confirmText": "Are you sure you want to cancel this order?",
                            "visibleOn": "${order_status === 'pending' || order_status === 'processing'}"
                        }
                    ]
                }
            ]
        }
    ],
    "selectable": true,
    "multiple": true,
    "bordered": true,
    "striped": true,
    "sticky": true,
    "autoFillHeight": true,
    "orderBy": "order_date",
    "orderDir": "desc",
    "placeholder": "No orders found"
}
```

### Financial Transactions Table
```json
{
    "type": "table",
    "title": "Financial Transactions",
    "source": "/api/transactions",
    "columns": [
        {
            "name": "transaction_id",
            "label": "Transaction ID",
            "width": 140,
            "copyable": true,
            "type": "text"
        },
        {
            "name": "date",
            "label": "Date",
            "type": "datetime",
            "format": "MMM DD, YYYY HH:mm",
            "sortable": true,
            "width": 160
        },
        {
            "name": "type",
            "label": "Type",
            "type": "status",
            "map": {
                "payment": "<span class='label label-success'>Payment</span>",
                "refund": "<span class='label label-warning'>Refund</span>",
                "chargeback": "<span class='label label-danger'>Chargeback</span>",
                "adjustment": "<span class='label label-info'>Adjustment</span>"
            }
        },
        {
            "name": "description",
            "label": "Description",
            "type": "text",
            "breakWord": true
        },
        {
            "name": "account",
            "label": "Account",
            "type": "text"
        },
        {
            "name": "amount",
            "label": "Amount",
            "type": "number",
            "prefix": "$",
            "precision": 2,
            "sortable": true,
            "classNameExpr": "${amount >= 0 ? 'text-green-600' : 'text-red-600'}"
        },
        {
            "name": "balance",
            "label": "Balance",
            "type": "number",
            "prefix": "$",
            "precision": 2,
            "classNameExpr": "${balance >= 0 ? 'text-green-600' : 'text-red-600 font-bold'}"
        },
        {
            "name": "status",
            "label": "Status",
            "type": "status",
            "map": {
                "completed": "<span class='label label-success'>Completed</span>",
                "pending": "<span class='label label-warning'>Pending</span>",
                "failed": "<span class='label label-danger'>Failed</span>",
                "cancelled": "<span class='label label-default'>Cancelled</span>"
            }
        },
        {
            "type": "operation",
            "label": "Actions",
            "width": 100,
            "buttons": [
                {
                    "type": "button",
                    "label": "Details",
                    "actionType": "dialog",
                    "level": "link",
                    "dialog": {
                        "title": "Transaction Details",
                        "body": {
                            "type": "service",
                            "api": "/api/transactions/${transaction_id}",
                            "body": [
                                {"type": "static", "name": "transaction_id", "label": "Transaction ID"},
                                {"type": "static", "name": "amount", "label": "Amount"},
                                {"type": "static", "name": "fee", "label": "Processing Fee"},
                                {"type": "static", "name": "net_amount", "label": "Net Amount"},
                                {"type": "static", "name": "payment_method", "label": "Payment Method"},
                                {"type": "static", "name": "reference", "label": "Reference"}
                            ]
                        }
                    }
                }
            ]
        }
    ],
    "bordered": true,
    "size": "sm",
    "sticky": true,
    "autoFillHeight": true,
    "orderBy": "date",
    "orderDir": "desc",
    "footer": {
        "type": "flex",
        "justify": "between",
        "items": [
            {
                "type": "tpl",
                "tpl": "Total Transactions: ${items | length}",
                "className": "text-sm text-gray-600"
            },
            {
                "type": "tpl",
                "tpl": "Net Amount: $${items | sum:amount | number:2}",
                "className": "text-sm font-bold text-blue-600"
            }
        ]
    }
}
```

### Project Tasks Table with Nested Data
```json
{
    "type": "table",
    "title": "Project Tasks",
    "source": "/api/projects/${project_id}/tasks",
    "columns": [
        {
            "name": "task_name",
            "label": "Task",
            "sortable": true,
            "type": "link",
            "href": "/tasks/${id}"
        },
        {
            "name": "assigned_to",
            "label": "Assigned To",
            "type": "avatar",
            "showName": true
        },
        {
            "name": "priority",
            "label": "Priority",
            "type": "status",
            "map": {
                "high": "<span class='label label-danger'>High</span>",
                "medium": "<span class='label label-warning'>Medium</span>",
                "low": "<span class='label label-success'>Low</span>"
            }
        },
        {
            "name": "progress",
            "label": "Progress",
            "type": "progress",
            "showLabel": true,
            "width": 150
        },
        {
            "name": "due_date",
            "label": "Due Date",
            "type": "date",
            "format": "MMM DD",
            "sortable": true,
            "classNameExpr": "${DATETODAY(due_date) < 0 ? 'text-red-600' : ''}"
        },
        {
            "name": "status",
            "label": "Status",
            "type": "status",
            "map": {
                "todo": "<span class='label label-default'>To Do</span>",
                "in_progress": "<span class='label label-info'>In Progress</span>",
                "review": "<span class='label label-warning'>Review</span>",
                "done": "<span class='label label-success'>Done</span>"
            }
        }
    ],
    "expandable": true,
    "expandableOn": "${subtasks && subtasks.length > 0}",
    "childrenColumnName": "subtasks",
    "bordered": true,
    "size": "sm",
    "autoFillHeight": true,
    "rowClassNameExpr": "${status === 'done' ? 'opacity-60' : ''}",
    "placeholder": "No tasks assigned to this project"
}
```

This component provides essential table functionality for ERP systems requiring structured data display, column management, and row operations without the full complexity of CRUD interfaces.