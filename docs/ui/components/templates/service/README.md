# Service Template

**FILE PURPOSE**: Data service template for API integration and dynamic content loading  
**SCOPE**: API data fetching, real-time updates, dynamic schema loading, and data orchestration  
**TARGET AUDIENCE**: Developers implementing data-driven components, API integration, and dynamic content

## 📋 Component Overview

Service provides comprehensive data service functionality with support for API integration, WebSocket connections, polling, schema loading, and data providers. Essential for creating data-driven ERP interfaces with real-time updates and dynamic content.

### Schema Reference
- **Primary Schema**: `ServiceSchema.json`
- **Related Schemas**: `SchemaApi`, `ComposedDataProvider`
- **Base Interface**: Data service template for API integration

## Basic Usage

```json
{
    "type": "service",
    "api": "/api/dashboard/stats",
    "body": [
        {"type": "stats", "title": "Revenue", "value": "${revenue}"},
        {"type": "stats", "title": "Orders", "value": "${orders}"}
    ]
}
```

## Go Type Definition

```go
type ServiceProps struct {
    Type                    string              `json:"type"`
    Body                    interface{}         `json:"body"`                // Content to render
    
    // Data Sources
    API                     interface{}         `json:"api"`                 // Primary data API
    WS                      string              `json:"ws"`                  // WebSocket address
    DataProvider            interface{}         `json:"dataProvider"`        // External data function
    
    // Schema Loading
    SchemaAPI               interface{}         `json:"schemaApi"`           // Remote schema API
    InitFetchSchema         bool                `json:"initFetchSchema"`     // Auto-load schema
    InitFetchSchemaOn       string              `json:"initFetchSchemaOn"`   // Schema load condition
    
    // Fetch Control
    InitFetch               bool                `json:"initFetch"`           // Auto-fetch on init
    InitFetchOn             string              `json:"initFetchOn"`         // Init fetch condition
    FetchOn                 string              `json:"fetchOn"`             // Fetch trigger condition
    
    // Polling and Updates
    Interval                int                 `json:"interval"`            // Polling interval
    SilentPolling           bool                `json:"silentPolling"`       // Silent polling mode
    StopAutoRefreshWhen     string              `json:"stopAutoRefreshWhen"` // Stop polling condition
    
    // Configuration
    LoadingConfig           interface{}         `json:"loadingConfig"`       // Loading display config
    ShowErrorMsg            bool                `json:"showErrorMsg"`        // Show API errors
    Messages                interface{}         `json:"messages"`            // Message configuration
    Name                    string              `json:"name"`                // Service identifier
}
```

## Data Integration Patterns

### Simple API Data Loading
```json
{
    "type": "service",
    "api": "/api/customer/${id}",
    "body": [
        {"type": "static", "name": "company_name", "label": "Company"},
        {"type": "static", "name": "email", "label": "Email"},
        {"type": "static", "name": "phone", "label": "Phone"}
    ]
}
```

### Real-time Data with WebSocket
```json
{
    "type": "service",
    "ws": "wss://api.example.com/realtime/dashboard",
    "body": [
        {"type": "stats", "title": "Active Users", "value": "${active_users}"},
        {"type": "stats", "title": "Live Orders", "value": "${live_orders}"}
    ]
}
```

### Polling for Live Updates
```json
{
    "type": "service",
    "api": "/api/monitoring/status",
    "interval": 5000,
    "silentPolling": true,
    "stopAutoRefreshWhen": "${status === 'offline'}",
    "body": [
        {"type": "status", "name": "system_status"},
        {"type": "progress", "name": "cpu_usage", "label": "CPU Usage"},
        {"type": "progress", "name": "memory_usage", "label": "Memory Usage"}
    ]
}
```

## Real-World Use Cases

### Customer Profile Service
```json
{
    "type": "service",
    "name": "customer_profile",
    "api": "/api/customers/${id}/profile",
    "initFetch": true,
    "body": [
        {
            "type": "grid",
            "columns": [
                {
                    "md": 8,
                    "body": [
                        {
                            "type": "panel",
                            "title": "Customer Information",
                            "body": [
                                {
                                    "type": "form",
                                    "api": "put:/api/customers/${id}",
                                    "mode": "horizontal",
                                    "body": [
                                        {"type": "static", "name": "company_name", "label": "Company Name"},
                                        {"type": "static", "name": "contact_person", "label": "Contact Person"},
                                        {"type": "static", "name": "email", "label": "Email Address"},
                                        {"type": "static", "name": "phone", "label": "Phone Number"},
                                        {"type": "static", "name": "industry", "label": "Industry"},
                                        {"type": "static", "name": "status", "label": "Status", "type": "status"}
                                    ]
                                }
                            ]
                        },
                        {
                            "type": "panel",
                            "title": "Recent Orders",
                            "body": [
                                {
                                    "type": "table",
                                    "source": "${recent_orders}",
                                    "columns": [
                                        {"name": "order_number", "label": "Order #", "type": "link", "href": "/orders/${id}"},
                                        {"name": "order_date", "label": "Date", "type": "date"},
                                        {"name": "total_amount", "label": "Amount", "type": "number", "prefix": "$"},
                                        {"name": "status", "label": "Status", "type": "status"}
                                    ]
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
                            "title": "Customer Stats",
                            "body": [
                                {"type": "stats", "title": "Total Orders", "value": "${total_orders}"},
                                {"type": "stats", "title": "Total Spent", "value": "$${total_spent | number}"},
                                {"type": "stats", "title": "Average Order", "value": "$${average_order | number}"},
                                {"type": "stats", "title": "Customer Since", "value": "${customer_since | date:\"MMM YYYY\"}"}
                            ]
                        },
                        {
                            "type": "panel",
                            "title": "Contact History",
                            "body": [
                                {
                                    "type": "list",
                                    "source": "${contact_history}",
                                    "listItem": {
                                        "title": "${contact_type}",
                                        "subTitle": "${contact_date | fromNow}",
                                        "desc": "${notes}"
                                    }
                                }
                            ]
                        }
                    ]
                }
            ]
        }
    ]
}
```

### Real-time Dashboard Service
```json
{
    "type": "service",
    "api": "/api/dashboard/realtime",
    "ws": "wss://api.example.com/dashboard/live",
    "interval": 10000,
    "silentPolling": true,
    "loadingConfig": {
        "show": true,
        "root": ".dashboard-container"
    },
    "body": [
        {
            "type": "grid",
            "className": "dashboard-container",
            "columns": [
                {
                    "md": 3,
                    "body": [
                        {
                            "type": "card",
                            "className": "bg-blue-50 border-blue-200",
                            "body": [
                                {"type": "stats", "title": "Live Revenue", "value": "$${live_revenue | number}", "className": "text-blue-600"},
                                {"type": "tpl", "tpl": "${revenue_trend > 0 ? '↗' : '↘'} ${revenue_trend}%", "className": "${revenue_trend > 0 ? 'text-green-600' : 'text-red-600'}"}
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
                                {"type": "stats", "title": "Active Orders", "value": "${active_orders}", "className": "text-green-600"},
                                {"type": "tpl", "tpl": "Processing: ${processing_orders}", "className": "text-sm text-gray-600"}
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
                                {"type": "stats", "title": "Online Users", "value": "${online_users}", "className": "text-orange-600"},
                                {"type": "tpl", "tpl": "Peak: ${peak_users}", "className": "text-sm text-gray-600"}
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
                                {"type": "stats", "title": "System Load", "value": "${system_load}%", "className": "text-purple-600"},
                                {"type": "progress", "value": "${system_load}", "showLabel": false, "className": "mt-2"}
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
                            "title": "Live Activity Feed",
                            "body": [
                                {
                                    "type": "list",
                                    "source": "${activity_feed}",
                                    "listItem": {
                                        "title": "${activity_title}",
                                        "subTitle": "${activity_time | fromNow}",
                                        "desc": "${activity_description}",
                                        "avatar": {"type": "icon", "icon": "${activity_icon}", "className": "${activity_color}"}
                                    },
                                    "placeholder": "No recent activity"
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
                            "title": "System Alerts",
                            "body": [
                                {
                                    "type": "each",
                                    "items": "${alerts}",
                                    "name": "alert",
                                    "body": [
                                        {
                                            "type": "alert",
                                            "level": "${alert.level}",
                                            "body": "${alert.message}",
                                            "closable": true
                                        }
                                    ]
                                }
                            ]
                        }
                    ]
                }
            ]
        }
    ]
}
```

### Dynamic Form Schema Service
```json
{
    "type": "service",
    "schemaApi": "/api/forms/${form_type}/schema",
    "api": "/api/forms/${form_type}/data",
    "initFetchSchema": true,
    "initFetch": true,
    "initFetchOn": "${form_type}",
    "body": [
        {
            "type": "form",
            "api": "post:/api/forms/${form_type}/submit",
            "body": "${schema.fields}",
            "actions": "${schema.actions}"
        }
    ]
}
```

### Order Processing Workflow Service
```json
{
    "type": "service",
    "name": "order_workflow",
    "api": "/api/orders/${order_id}/workflow",
    "interval": 5000,
    "silentPolling": true,
    "stopAutoRefreshWhen": "${workflow_status === 'completed' || workflow_status === 'cancelled'}",
    "body": [
        {
            "type": "steps",
            "value": "${current_step}",
            "status": "${step_statuses}",
            "steps": [
                {
                    "title": "Order Received",
                    "description": "Order has been received and validated",
                    "body": [
                        {"type": "static", "name": "order_number", "label": "Order Number"},
                        {"type": "static", "name": "order_date", "label": "Order Date", "tpl": "${order_date | date}"},
                        {"type": "static", "name": "customer_name", "label": "Customer"}
                    ]
                },
                {
                    "title": "Payment Processing",
                    "description": "Payment is being processed",
                    "body": [
                        {"type": "static", "name": "payment_method", "label": "Payment Method"},
                        {"type": "static", "name": "payment_amount", "label": "Amount", "tpl": "$${payment_amount}"},
                        {"type": "progress", "name": "payment_progress", "label": "Processing Progress"}
                    ]
                },
                {
                    "title": "Inventory Check",
                    "description": "Checking product availability",
                    "body": [
                        {
                            "type": "table",
                            "source": "${order_items}",
                            "columns": [
                                {"name": "product_name", "label": "Product"},
                                {"name": "quantity", "label": "Quantity"},
                                {"name": "availability_status", "label": "Availability", "type": "status"}
                            ]
                        }
                    ]
                },
                {
                    "title": "Fulfillment",
                    "description": "Order is being prepared for shipping",
                    "body": [
                        {"type": "static", "name": "fulfillment_center", "label": "Fulfillment Center"},
                        {"type": "static", "name": "estimated_ship_date", "label": "Estimated Ship Date", "tpl": "${estimated_ship_date | date}"},
                        {"type": "progress", "name": "fulfillment_progress", "label": "Preparation Progress"}
                    ]
                },
                {
                    "title": "Shipped",
                    "description": "Order has been shipped",
                    "body": [
                        {"type": "static", "name": "tracking_number", "label": "Tracking Number"},
                        {"type": "static", "name": "carrier", "label": "Shipping Carrier"},
                        {"type": "static", "name": "ship_date", "label": "Ship Date", "tpl": "${ship_date | date}"},
                        {"type": "static", "name": "estimated_delivery", "label": "Estimated Delivery", "tpl": "${estimated_delivery | date}"}
                    ]
                }
            ]
        }
    ]
}
```

### Inventory Monitoring Service
```json
{
    "type": "service",
    "api": "/api/inventory/monitoring",
    "interval": 15000,
    "silentPolling": true,
    "body": [
        {
            "type": "grid",
            "columns": [
                {
                    "md": 6,
                    "body": [
                        {
                            "type": "panel",
                            "title": "Low Stock Alerts",
                            "className": "border-red-200",
                            "body": [
                                {
                                    "type": "table",
                                    "source": "${low_stock_items}",
                                    "columns": [
                                        {"name": "product_name", "label": "Product"},
                                        {"name": "current_stock", "label": "Current", "type": "number"},
                                        {"name": "reorder_level", "label": "Reorder Level", "type": "number"},
                                        {
                                            "type": "operation",
                                            "buttons": [
                                                {"type": "button", "label": "Reorder", "actionType": "dialog", "level": "primary", "size": "sm"}
                                            ]
                                        }
                                    ],
                                    "placeholder": "No low stock items"
                                }
                            ]
                        }
                    ]
                },
                {
                    "md": 6,
                    "body": [
                        {
                            "type": "panel",
                            "title": "Out of Stock",
                            "className": "border-red-500",
                            "body": [
                                {
                                    "type": "list",
                                    "source": "${out_of_stock_items}",
                                    "listItem": {
                                        "title": "${product_name}",
                                        "subTitle": "SKU: ${sku}",
                                        "desc": "Last updated: ${last_updated | fromNow}",
                                        "actions": [
                                            {"type": "button", "label": "Urgent Reorder", "actionType": "ajax", "api": "post:/api/inventory/urgent-reorder", "level": "danger", "size": "sm"}
                                        ]
                                    },
                                    "placeholder": "No out of stock items"
                                }
                            ]
                        }
                    ]
                }
            ]
        },
        {
            "type": "panel",
            "title": "Inventory Overview",
            "body": [
                {
                    "type": "grid",
                    "columns": [
                        {
                            "md": 2,
                            "body": [{"type": "stats", "title": "Total Products", "value": "${total_products}"}]
                        },
                        {
                            "md": 2,
                            "body": [{"type": "stats", "title": "In Stock", "value": "${in_stock_count}", "className": "text-green-600"}]
                        },
                        {
                            "md": 2,
                            "body": [{"type": "stats", "title": "Low Stock", "value": "${low_stock_count}", "className": "text-orange-600"}]
                        },
                        {
                            "md": 2,
                            "body": [{"type": "stats", "title": "Out of Stock", "value": "${out_of_stock_count}", "className": "text-red-600"}]
                        },
                        {
                            "md": 2,
                            "body": [{"type": "stats", "title": "Total Value", "value": "$${total_inventory_value | number}"}]
                        },
                        {
                            "md": 2,
                            "body": [{"type": "stats", "title": "Reorder Needed", "value": "${reorder_needed_count}"}]
                        }
                    ]
                }
            ]
        }
    ]
}
```

### Financial Analytics Service
```json
{
    "type": "service",
    "api": "/api/financial/analytics",
    "dataProvider": "financialDataProvider",
    "fetchOn": "${date_range_changed || filter_changed}",
    "body": [
        {
            "type": "form",
            "target": "financial_analytics",
            "mode": "inline",
            "body": [
                {"type": "input-date-range", "name": "date_range", "label": "Date Range", "value": "last_30_days"},
                {"type": "select", "name": "metric", "label": "Metric", "options": ["revenue", "profit", "expenses", "cash_flow"], "value": "revenue"},
                {"type": "select", "name": "grouping", "label": "Group By", "options": ["daily", "weekly", "monthly"], "value": "daily"}
            ]
        },
        {
            "type": "service",
            "name": "financial_analytics",
            "api": "/api/financial/analytics/data",
            "body": [
                {
                    "type": "grid",
                    "columns": [
                        {
                            "md": 8,
                            "body": [
                                {
                                    "type": "panel",
                                    "title": "${metric | upper} Trend",
                                    "body": [
                                        {
                                            "type": "chart",
                                            "source": "${chart_data}",
                                            "config": {
                                                "type": "line",
                                                "xField": "date",
                                                "yField": "value",
                                                "smooth": true
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
                                    "title": "Key Metrics",
                                    "body": [
                                        {"type": "stats", "title": "Current Period", "value": "$${current_period | number}"},
                                        {"type": "stats", "title": "Previous Period", "value": "$${previous_period | number}"},
                                        {"type": "stats", "title": "Change", "value": "${period_change}%", "className": "${period_change > 0 ? 'text-green-600' : 'text-red-600'}"},
                                        {"type": "stats", "title": "YTD Total", "value": "$${ytd_total | number}"}
                                    ]
                                }
                            ]
                        }
                    ]
                },
                {
                    "type": "panel",
                    "title": "Detailed Breakdown",
                    "body": [
                        {
                            "type": "table",
                            "source": "${detailed_data}",
                            "columns": [
                                {"name": "date", "label": "Date", "type": "date"},
                                {"name": "revenue", "label": "Revenue", "type": "number", "prefix": "$"},
                                {"name": "expenses", "label": "Expenses", "type": "number", "prefix": "$"},
                                {"name": "profit", "label": "Profit", "type": "number", "prefix": "$"},
                                {"name": "margin", "label": "Margin", "type": "number", "suffix": "%"}
                            ]
                        }
                    ]
                }
            ]
        }
    ]
}
```

This component provides essential data service functionality for ERP systems requiring API integration, real-time updates, dynamic content loading, and complex data orchestration with polling, WebSocket support, and external data providers.