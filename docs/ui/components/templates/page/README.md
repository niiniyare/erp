# Page Template

**FILE PURPOSE**: Complete page layout template for ERP application pages  
**SCOPE**: Full page structure, layout management, sidebar, toolbar, and data initialization  
**TARGET AUDIENCE**: Developers implementing complete ERP pages, dashboards, and application layouts

## 📋 Component Overview

Page provides comprehensive page-level functionality with support for sidebar layouts, toolbars, data initialization, polling, and responsive design. Essential for creating complete ERP application pages with proper structure and data management.

### Schema Reference
- **Primary Schema**: `PageSchema.json`
- **Related Schemas**: `SchemaCollection`, `SchemaApi`
- **Base Interface**: Complete page template for application layouts

## Basic Usage

```json
{
    "type": "page",
    "title": "Customer Management",
    "body": [
        {"type": "text", "text": "Welcome to customer management"}
    ]
}
```

## Go Type Definition

```go
type PageProps struct {
    Type                    string              `json:"type"`
    Title                   string              `json:"title"`               // Page title
    SubTitle                string              `json:"subTitle"`            // Page subtitle
    Remark                  interface{}         `json:"remark"`              // Help information
    
    // Content Areas
    Body                    interface{}         `json:"body"`                // Main content area
    BodyClassName           string              `json:"bodyClassName"`       // Body CSS classes
    Aside                   interface{}         `json:"aside"`               // Sidebar content
    AsideClassName          string              `json:"asideClassName"`      // Sidebar CSS
    Toolbar                 interface{}         `json:"toolbar"`             // Top toolbar
    ToolbarClassName        string              `json:"toolbarClassName"`    // Toolbar CSS
    HeaderClassName         string              `json:"headerClassName"`     // Header CSS
    
    // Sidebar Configuration
    AsidePosition           string              `json:"asidePosition"`       // "left" or "right"
    AsideResizor            bool                `json:"asideResizor"`        // Resizable sidebar
    AsideSticky             bool                `json:"asideSticky"`         // Sticky sidebar
    AsideMinWidth           int                 `json:"asideMinWidth"`       // Min sidebar width
    AsideMaxWidth           int                 `json:"asideMaxWidth"`       // Max sidebar width
    
    // Data Management
    Data                    interface{}         `json:"data"`                // Initial page data
    InitAPI                 interface{}         `json:"initApi"`             // Data initialization API
    InitFetch               bool                `json:"initFetch"`           // Auto-fetch on load
    InitFetchOn             string              `json:"initFetchOn"`         // Conditional fetch
    
    // Polling and Updates
    Interval                int                 `json:"interval"`            // Polling interval
    SilentPolling           bool                `json:"silentPolling"`       // Silent polling
    StopAutoRefreshWhen     string              `json:"stopAutoRefreshWhen"` // Stop polling condition
    
    // Styling
    CSS                     interface{}         `json:"css"`                 // Custom page CSS
    MobileCSS               interface{}         `json:"mobileCSS"`           // Mobile-specific CSS
    CSSVars                 interface{}         `json:"cssVars"`             // CSS variables
    
    // Display Options
    ShowErrorMsg            bool                `json:"showErrorMsg"`        // Show error messages
    LoadingConfig           interface{}         `json:"loadingConfig"`       // Loading configuration
    
    // Mobile Features
    PullRefresh             interface{}         `json:"pullRefresh"`         // Pull-to-refresh config
    
    // Layout Control
    Regions                 []string            `json:"regions"`             // Visible regions
    Definitions             interface{}         `json:"definitions"`         // Reusable definitions
    Messages                interface{}         `json:"messages"`            // Message configuration
    Name                    string              `json:"name"`                // Page identifier
}
```

## Page Layout Types

### Basic Page Layout
```json
{
    "type": "page",
    "title": "Dashboard",
    "subTitle": "Welcome to your ERP dashboard",
    "body": [
        {
            "type": "grid",
            "columns": [
                {
                    "md": 3,
                    "body": [{"type": "stats", "title": "Total Sales", "value": "$124,500"}]
                },
                {
                    "md": 3,
                    "body": [{"type": "stats", "title": "New Customers", "value": "48"}]
                },
                {
                    "md": 3,
                    "body": [{"type": "stats", "title": "Orders", "value": "156"}]
                },
                {
                    "md": 3,
                    "body": [{"type": "stats", "title": "Revenue", "value": "$89,400"}]
                }
            ]
        }
    ]
}
```

### Page with Sidebar
```json
{
    "type": "page",
    "title": "Customer Details",
    "aside": [
        {
            "type": "nav",
            "stacked": true,
            "links": [
                {"label": "Overview", "to": "#overview"},
                {"label": "Orders", "to": "#orders"},
                {"label": "Invoices", "to": "#invoices"},
                {"label": "Support", "to": "#support"}
            ]
        }
    ],
    "asidePosition": "left",
    "asideResizor": true,
    "asideMinWidth": 200,
    "asideMaxWidth": 400,
    "body": [
        {"type": "text", "text": "Customer details content"}
    ]
}
```

### Page with Toolbar
```json
{
    "type": "page",
    "title": "Product Management",
    "toolbar": [
        {
            "type": "button",
            "label": "Add Product",
            "actionType": "dialog",
            "level": "primary",
            "icon": "plus"
        },
        {
            "type": "button",
            "label": "Export",
            "actionType": "download",
            "api": "/api/products/export"
        }
    ],
    "body": [
        {
            "type": "crud2",
            "api": "/api/products",
            "columns": [
                {"name": "name", "label": "Product Name"},
                {"name": "price", "label": "Price"}
            ]
        }
    ]
}
```

## Real-World Use Cases

### Complete ERP Dashboard Page
```json
{
    "type": "page",
    "title": "ERP Dashboard",
    "subTitle": "Overview of your business operations",
    "initApi": "/api/dashboard/data",
    "interval": 30000,
    "silentPolling": true,
    "remark": {
        "type": "remark",
        "content": "This dashboard provides real-time insights into your business performance. Data is updated every 30 seconds."
    },
    "toolbar": [
        {
            "type": "dropdown-button",
            "label": "Quick Actions",
            "buttons": [
                {"type": "button", "label": "New Customer", "actionType": "dialog", "dialog": {"$ref": "#/definitions/CustomerForm"}},
                {"type": "button", "label": "New Order", "actionType": "link", "link": "/orders/create"},
                {"type": "button", "label": "Create Invoice", "actionType": "dialog"}
            ]
        },
        {
            "type": "button",
            "label": "Export Report",
            "actionType": "download",
            "api": "/api/dashboard/export",
            "level": "primary"
        }
    ],
    "body": [
        {
            "type": "grid",
            "columns": [
                {
                    "md": 3,
                    "body": [
                        {
                            "type": "card",
                            "className": "bg-blue-50 border-blue-200",
                            "body": [
                                {"type": "stats", "title": "Total Revenue", "value": "$${total_revenue | number}", "className": "text-blue-600"},
                                {"type": "tpl", "tpl": "${revenue_change > 0 ? '↗' : '↘'} ${revenue_change}% from last month", "className": "${revenue_change > 0 ? 'text-green-600' : 'text-red-600'}"}
                            ]
                        }
                    ]
                },
                {
                    "md": 3,
                    "body": [
                        {
                            "type": "card",
                            "className": "bg-green-50 border-green-200",
                            "body": [
                                {"type": "stats", "title": "New Customers", "value": "${new_customers}", "className": "text-green-600"},
                                {"type": "tpl", "tpl": "${customer_change > 0 ? '↗' : '↘'} ${customer_change}% from last month", "className": "${customer_change > 0 ? 'text-green-600' : 'text-red-600'}"}
                            ]
                        }
                    ]
                },
                {
                    "md": 3,
                    "body": [
                        {
                            "type": "card",
                            "className": "bg-orange-50 border-orange-200",
                            "body": [
                                {"type": "stats", "title": "Open Orders", "value": "${open_orders}", "className": "text-orange-600"},
                                {"type": "tpl", "tpl": "${orders_change > 0 ? '↗' : '↘'} ${orders_change}% from last month", "className": "${orders_change > 0 ? 'text-green-600' : 'text-red-600'}"}
                            ]
                        }
                    ]
                },
                {
                    "md": 3,
                    "body": [
                        {
                            "type": "card",
                            "className": "bg-purple-50 border-purple-200",
                            "body": [
                                {"type": "stats", "title": "Pending Invoices", "value": "${pending_invoices}", "className": "text-purple-600"},
                                {"type": "tpl", "tpl": "Total: $${pending_invoices_amount | number}", "className": "text-sm text-gray-600"}
                            ]
                        }
                    ]
                }
            ]
        },
        {
            "type": "grid",
            "columns": [
                {
                    "md": 8,
                    "body": [
                        {
                            "type": "panel",
                            "title": "Sales Performance",
                            "body": [
                                {
                                    "type": "chart",
                                    "api": "/api/dashboard/sales-chart",
                                    "config": {
                                        "type": "line",
                                        "xField": "month",
                                        "yField": "sales",
                                        "seriesField": "category"
                                    }
                                }
                            ]
                        }
                    ]
                },
                {
                    "md": 4,
                    "body": [
                        {
                            "type": "panel",
                            "title": "Recent Activities",
                            "body": [
                                {
                                    "type": "list",
                                    "source": "/api/dashboard/recent-activities",
                                    "listItem": {
                                        "title": "${activity_title}",
                                        "subTitle": "${activity_time | fromNow}",
                                        "avatar": {"type": "icon", "icon": "${activity_icon}"}
                                    }
                                }
                            ]
                        }
                    ]
                }
            ]
        },
        {
            "type": "panel",
            "title": "Quick Insights",
            "body": [
                {
                    "type": "tabs",
                    "tabs": [
                        {
                            "title": "Top Products",
                            "body": [
                                {
                                    "type": "table",
                                    "source": "/api/dashboard/top-products",
                                    "columns": [
                                        {"name": "product_name", "label": "Product"},
                                        {"name": "sales_count", "label": "Sales", "type": "number"},
                                        {"name": "revenue", "label": "Revenue", "type": "number", "prefix": "$"}
                                    ]
                                }
                            ]
                        },
                        {
                            "title": "Top Customers",
                            "body": [
                                {
                                    "type": "table",
                                    "source": "/api/dashboard/top-customers",
                                    "columns": [
                                        {"name": "customer_name", "label": "Customer"},
                                        {"name": "order_count", "label": "Orders", "type": "number"},
                                        {"name": "total_spent", "label": "Total Spent", "type": "number", "prefix": "$"}
                                    ]
                                }
                            ]
                        },
                        {
                            "title": "Alerts",
                            "badge": {"mode": "text", "text": "${alert_count}", "visibleOn": "${alert_count > 0}"},
                            "body": [
                                {
                                    "type": "list",
                                    "source": "/api/dashboard/alerts",
                                    "listItem": {
                                        "title": "${alert_title}",
                                        "desc": "${alert_description}",
                                        "remark": {"type": "status", "value": "${alert_level}"}
                                    }
                                }
                            ]
                        }
                    ]
                }
            ]
        }
    ],
    "definitions": {
        "CustomerForm": {
            "type": "dialog",
            "title": "Add New Customer",
            "body": {
                "type": "form",
                "api": "post:/api/customers",
                "body": [
                    {"type": "input-text", "name": "company_name", "label": "Company Name", "required": true},
                    {"type": "input-email", "name": "email", "label": "Email", "required": true},
                    {"type": "input-text", "name": "phone", "label": "Phone"}
                ]
            }
        }
    }
}
```

### Customer Management Page with Sidebar
```json
{
    "type": "page",
    "title": "Customer Management",
    "subTitle": "Manage your customer relationships",
    "initApi": "/api/customers/page-data",
    "aside": [
        {
            "type": "panel",
            "title": "Customer Filters",
            "body": [
                {
                    "type": "form",
                    "target": "customer_crud",
                    "body": [
                        {
                            "type": "select",
                            "name": "status",
                            "label": "Status",
                            "options": ["all", "active", "inactive", "prospect"],
                            "value": "all"
                        },
                        {
                            "type": "select",
                            "name": "industry",
                            "label": "Industry",
                            "source": "/api/industries",
                            "clearable": true
                        },
                        {
                            "type": "select",
                            "name": "region",
                            "label": "Region",
                            "source": "/api/regions",
                            "clearable": true
                        },
                        {
                            "type": "input-date-range",
                            "name": "created_date_range",
                            "label": "Created Date"
                        }
                    ],
                    "actions": [
                        {"type": "submit", "label": "Apply Filters", "level": "primary"},
                        {"type": "reset", "label": "Clear Filters"}
                    ]
                }
            ]
        },
        {
            "type": "divider"
        },
        {
            "type": "panel",
            "title": "Quick Stats",
            "body": [
                {"type": "stats", "title": "Total Customers", "value": "${total_customers}"},
                {"type": "stats", "title": "Active", "value": "${active_customers}"},
                {"type": "stats", "title": "This Month", "value": "${new_this_month}"}
            ]
        }
    ],
    "asidePosition": "left",
    "asideResizor": true,
    "asideMinWidth": 280,
    "asideMaxWidth": 400,
    "asideSticky": true,
    "toolbar": [
        {
            "type": "button",
            "label": "Add Customer",
            "actionType": "dialog",
            "icon": "plus",
            "level": "primary",
            "dialog": {
                "title": "Add New Customer",
                "size": "lg",
                "body": {
                    "type": "form",
                    "api": "post:/api/customers",
                    "body": [
                        {"type": "input-text", "name": "company_name", "label": "Company Name", "required": true},
                        {"type": "input-text", "name": "contact_person", "label": "Contact Person"},
                        {"type": "input-email", "name": "email", "label": "Email", "required": true},
                        {"type": "input-text", "name": "phone", "label": "Phone"},
                        {"type": "select", "name": "industry", "label": "Industry", "source": "/api/industries"},
                        {"type": "select", "name": "status", "label": "Status", "options": ["active", "prospect"]}
                    ]
                }
            }
        },
        {
            "type": "dropdown-button",
            "label": "Import/Export",
            "buttons": [
                {"type": "button", "label": "Import CSV", "actionType": "dialog"},
                {"type": "button", "label": "Export CSV", "actionType": "download", "api": "/api/customers/export?format=csv"},
                {"type": "button", "label": "Export Excel", "actionType": "download", "api": "/api/customers/export?format=xlsx"}
            ]
        }
    ],
    "body": [
        {
            "type": "crud2",
            "name": "customer_crud",
            "api": "/api/customers",
            "columns": [
                {"name": "company_name", "label": "Company", "searchable": true, "sortable": true},
                {"name": "contact_person", "label": "Contact"},
                {"name": "email", "label": "Email", "type": "email"},
                {"name": "phone", "label": "Phone"},
                {"name": "industry", "label": "Industry"},
                {"name": "status", "label": "Status", "type": "status"},
                {"name": "created_date", "label": "Created", "type": "date"}
            ],
            "headerToolbar": ["bulkActions", "reload"],
            "footerToolbar": ["statistics", "pagination"],
            "bulkActions": [
                {"type": "button", "label": "Bulk Delete", "actionType": "ajax", "api": "delete:/api/customers/bulk", "confirmText": "Delete selected customers?"},
                {"type": "button", "label": "Bulk Export", "actionType": "download", "api": "/api/customers/bulk-export"}
            ]
        }
    ]
}
```

### Product Catalog Page with Advanced Layout
```json
{
    "type": "page",
    "title": "Product Catalog",
    "subTitle": "Manage your product inventory",
    "initApi": "/api/products/page-data",
    "css": {
        ".product-grid": {
            "display": "grid",
            "grid-template-columns": "repeat(auto-fill, minmax(300px, 1fr))",
            "gap": "1rem"
        }
    },
    "mobileCSS": {
        ".product-grid": {
            "grid-template-columns": "1fr"
        }
    },
    "toolbar": [
        {
            "type": "button",
            "label": "Add Product",
            "actionType": "dialog",
            "icon": "plus",
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
                        {"type": "select", "name": "category_id", "label": "Category", "source": "/api/categories"},
                        {"type": "input-number", "name": "price", "label": "Price", "prefix": "$", "min": 0},
                        {"type": "textarea", "name": "description", "label": "Description"},
                        {"type": "input-image", "name": "images", "label": "Product Images", "multiple": true}
                    ]
                }
            }
        },
        {
            "type": "button-group",
            "buttons": [
                {"type": "button", "label": "Grid View", "actionType": "reload", "level": "default", "active": true},
                {"type": "button", "label": "List View", "actionType": "url", "url": "/products?view=list"}
            ]
        }
    ],
    "body": [
        {
            "type": "form",
            "target": "product_results",
            "mode": "inline",
            "body": [
                {"type": "input-text", "name": "search", "placeholder": "Search products...", "clearable": true},
                {"type": "select", "name": "category", "placeholder": "Category", "source": "/api/categories", "clearable": true},
                {"type": "select", "name": "availability", "placeholder": "Availability", "options": ["in_stock", "low_stock", "out_of_stock"], "clearable": true},
                {"type": "range", "name": "price_range", "label": "Price Range", "min": 0, "max": 1000}
            ],
            "actions": [
                {"type": "submit", "label": "Filter", "level": "primary"},
                {"type": "reset", "label": "Clear"}
            ]
        },
        {
            "type": "service",
            "name": "product_results",
            "api": "/api/products",
            "body": [
                {
                    "type": "cards",
                    "source": "${items}",
                    "className": "product-grid",
                    "card": {
                        "header": {
                            "title": "${name}",
                            "subTitle": "SKU: ${sku}",
                            "avatar": {
                                "type": "image",
                                "src": "${main_image}",
                                "className": "w-16 h-16 object-cover rounded"
                            }
                        },
                        "body": [
                            {"type": "text", "text": "${description}", "className": "text-gray-600 mb-2"},
                            {"type": "text", "text": "Price: $${price}", "className": "font-bold text-xl text-green-600"},
                            {"type": "text", "text": "Stock: ${stock_quantity}", "className": "text-sm"},
                            {"type": "text", "text": "Category: ${category_name}", "className": "text-sm text-gray-500"}
                        ],
                        "actions": [
                            {"type": "button", "label": "Edit", "actionType": "dialog", "level": "primary", "size": "sm"},
                            {"type": "button", "label": "View", "actionType": "link", "link": "/products/${id}", "level": "link", "size": "sm"}
                        ]
                    }
                }
            ]
        }
    ],
    "pullRefresh": {
        "disabled": false,
        "pullingText": "Pull to refresh products",
        "loosingText": "Release to refresh"
    }
}
```

### Settings Page with Complex Layout
```json
{
    "type": "page",
    "title": "System Settings",
    "subTitle": "Configure your ERP system",
    "initApi": "/api/settings/all",
    "aside": [
        {
            "type": "nav",
            "stacked": true,
            "links": [
                {"label": "General", "to": "#general", "icon": "settings"},
                {"label": "Users & Permissions", "to": "#users", "icon": "users"},
                {"label": "Email Settings", "to": "#email", "icon": "mail"},
                {"label": "Security", "to": "#security", "icon": "shield"},
                {"label": "Integrations", "to": "#integrations", "icon": "link"},
                {"label": "Backup & Restore", "to": "#backup", "icon": "database"}
            ]
        }
    ],
    "asidePosition": "left",
    "asideMinWidth": 200,
    "asideSticky": true,
    "body": [
        {
            "type": "tabs",
            "tabsMode": "card",
            "tabs": [
                {
                    "title": "General Settings",
                    "hash": "general",
                    "body": [
                        {
                            "type": "form",
                            "api": "put:/api/settings/general",
                            "title": "Company Information",
                            "body": [
                                {"type": "input-text", "name": "company_name", "label": "Company Name", "required": true},
                                {"type": "input-text", "name": "company_address", "label": "Address"},
                                {"type": "input-email", "name": "company_email", "label": "Email"},
                                {"type": "select", "name": "timezone", "label": "Timezone", "source": "/api/timezones"},
                                {"type": "select", "name": "currency", "label": "Default Currency", "source": "/api/currencies"}
                            ]
                        }
                    ]
                },
                {
                    "title": "User Management",
                    "hash": "users",
                    "body": [
                        {
                            "type": "crud2",
                            "api": "/api/users",
                            "title": "System Users",
                            "headerToolbar": [
                                {"type": "button", "label": "Add User", "actionType": "dialog", "level": "primary"}
                            ],
                            "columns": [
                                {"name": "username", "label": "Username"},
                                {"name": "email", "label": "Email"},
                                {"name": "role", "label": "Role"},
                                {"name": "status", "label": "Status", "type": "status"}
                            ]
                        }
                    ]
                }
            ]
        }
    ]
}
```

This component provides essential page-level functionality for ERP systems requiring complete application layouts, data management, responsive design, and complex multi-area page structures with sidebar navigation and toolbar integration.