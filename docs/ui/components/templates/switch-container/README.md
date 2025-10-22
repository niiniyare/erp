# Switch Container Template

**FILE PURPOSE**: Conditional rendering template for displaying different content based on state or conditions  
**SCOPE**: State-based UI rendering, conditional components, and dynamic content switching  
**TARGET AUDIENCE**: Developers implementing conditional interfaces, multi-state components, and dynamic routing

## 📋 Component Overview

SwitchContainer provides powerful conditional rendering capabilities, allowing you to display different components or layouts based on application state, user permissions, or data conditions. Essential for building adaptive interfaces that respond to changing context and state.

### Schema Reference
- **Primary Schema**: `SwitchContainerSchema.json`
- **Related Schemas**: `StateSchema`
- **Base Interface**: Conditional rendering template for state-based components

## Basic Usage

```json
{
    "type": "switch-container",
    "items": [
        {
            "test": "${user.role === 'admin'}",
            "schema": {
                "type": "div",
                "body": [
                    {"type": "text", "text": "Admin Dashboard"}
                ]
            }
        },
        {
            "test": "${user.role === 'user'}",
            "schema": {
                "type": "div",
                "body": [
                    {"type": "text", "text": "User Dashboard"}
                ]
            }
        }
    ]
}
```

## Go Type Definition

```go
type SwitchContainerProps struct {
    Type                    string              `json:"type"`
    Items                   []StateSchema       `json:"items"`             // Conditional items
}

type StateSchema struct {
    Test                    string              `json:"test"`              // Condition expression
    Schema                  interface{}         `json:"schema"`            // Component to render
}
```

## Conditional Rendering Patterns

### User Role-Based Rendering
```json
{
    "type": "switch-container",
    "items": [
        {
            "test": "${user.permissions.includes('admin.full_access')}",
            "schema": {
                "type": "page",
                "title": "Administrator Dashboard",
                "body": [
                    {
                        "type": "grid",
                        "columns": [
                            {
                                "md": 3,
                                "body": [
                                    {"type": "stats", "title": "Total Users", "value": "${stats.total_users}"},
                                    {"type": "stats", "title": "Active Sessions", "value": "${stats.active_sessions}"},
                                    {"type": "stats", "title": "System Load", "value": "${stats.system_load}%"}
                                ]
                            },
                            {
                                "md": 9,
                                "body": [
                                    {
                                        "type": "crud2",
                                        "api": "/api/admin/users",
                                        "title": "User Management",
                                        "columns": [
                                            {"name": "name", "label": "Name"},
                                            {"name": "email", "label": "Email"},
                                            {"name": "role", "label": "Role"},
                                            {"name": "last_login", "label": "Last Login", "type": "date"}
                                        ]
                                    }
                                ]
                            }
                        ]
                    }
                ]
            }
        },
        {
            "test": "${user.permissions.includes('manager.view')}",
            "schema": {
                "type": "page",
                "title": "Manager Dashboard",
                "body": [
                    {
                        "type": "tabs",
                        "tabs": [
                            {
                                "title": "Team Overview",
                                "body": [
                                    {
                                        "type": "service",
                                        "api": "/api/manager/team-stats",
                                        "body": [
                                            {"type": "chart", "config": "${team_performance_chart}"}
                                        ]
                                    }
                                ]
                            },
                            {
                                "title": "Reports",
                                "body": [
                                    {
                                        "type": "crud2",
                                        "api": "/api/manager/reports",
                                        "columns": [
                                            {"name": "report_name", "label": "Report"},
                                            {"name": "generated_date", "label": "Date", "type": "date"}
                                        ]
                                    }
                                ]
                            }
                        ]
                    }
                ]
            }
        },
        {
            "test": "true",
            "schema": {
                "type": "page",
                "title": "Employee Dashboard",
                "body": [
                    {
                        "type": "alert",
                        "level": "info",
                        "body": "Welcome to your employee dashboard. View your assignments and submit time reports."
                    },
                    {
                        "type": "service",
                        "api": "/api/employee/assignments",
                        "body": [
                            {
                                "type": "list",
                                "source": "${assignments}",
                                "listItem": {
                                    "title": "${task_name}",
                                    "subTitle": "Due: ${due_date | date}",
                                    "desc": "${description}"
                                }
                            }
                        ]
                    }
                ]
            }
        }
    ]
}
```

### Application State-Based Rendering
```json
{
    "type": "switch-container",
    "items": [
        {
            "test": "${app.loading === true}",
            "schema": {
                "type": "div",
                "className": "loading-container flex items-center justify-center min-h-screen",
                "body": [
                    {
                        "type": "div",
                        "className": "text-center",
                        "body": [
                            {"type": "spinner", "size": "lg"},
                            {"type": "text", "text": "Loading application...", "className": "mt-4 text-gray-600"}
                        ]
                    }
                ]
            }
        },
        {
            "test": "${app.error}",
            "schema": {
                "type": "div",
                "className": "error-container flex items-center justify-center min-h-screen",
                "body": [
                    {
                        "type": "alert",
                        "level": "danger",
                        "title": "Application Error",
                        "body": "${app.error.message}",
                        "actions": [
                            {
                                "type": "button",
                                "label": "Retry",
                                "actionType": "ajax",
                                "api": "/api/app/retry",
                                "level": "primary"
                            },
                            {
                                "type": "button",
                                "label": "Contact Support",
                                "actionType": "link",
                                "link": "/support"
                            }
                        ]
                    }
                ]
            }
        },
        {
            "test": "${app.maintenance_mode === true}",
            "schema": {
                "type": "div",
                "className": "maintenance-container flex items-center justify-center min-h-screen bg-gray-100",
                "body": [
                    {
                        "type": "card",
                        "className": "max-w-md text-center",
                        "header": {
                            "title": "Scheduled Maintenance",
                            "avatar": {"type": "icon", "icon": "tool", "className": "bg-orange-500 text-white"}
                        },
                        "body": [
                            {"type": "text", "text": "We're currently performing scheduled maintenance to improve your experience."},
                            {"type": "text", "text": "Expected completion: ${app.maintenance_end | date}", "className": "mt-2 font-semibold"},
                            {"type": "text", "text": "Thank you for your patience.", "className": "mt-2 text-gray-600"}
                        ]
                    }
                ]
            }
        },
        {
            "test": "true",
            "schema": {
                "$ref": "#/definitions/MainApplication"
            }
        }
    ]
}
```

### Data State-Based Rendering
```json
{
    "type": "service",
    "api": "/api/orders/${id}",
    "body": [
        {
            "type": "switch-container",
            "items": [
                {
                    "test": "${order.status === 'pending'}",
                    "schema": {
                        "type": "card",
                        "className": "order-pending border-l-4 border-orange-500",
                        "header": {
                            "title": "Order #${order.order_number}",
                            "subTitle": "Pending Processing",
                            "avatar": {"type": "icon", "icon": "clock", "className": "bg-orange-500 text-white"}
                        },
                        "body": [
                            {"type": "text", "text": "This order is waiting to be processed."},
                            {"type": "text", "text": "Order Date: ${order.order_date | date}"},
                            {"type": "text", "text": "Customer: ${order.customer_name}"},
                            {"type": "text", "text": "Total: $${order.total_amount | number}"}
                        ],
                        "actions": [
                            {
                                "type": "button",
                                "label": "Start Processing",
                                "actionType": "dialog",
                                "level": "primary",
                                "dialog": {
                                    "title": "Start Order Processing",
                                    "body": {
                                        "type": "form",
                                        "api": "put:/api/orders/${order.id}/start-processing",
                                        "body": [
                                            {"type": "select", "name": "fulfillment_center", "label": "Fulfillment Center", "source": "/api/fulfillment-centers"},
                                            {"type": "textarea", "name": "processing_notes", "label": "Processing Notes"}
                                        ]
                                    }
                                }
                            }
                        ]
                    }
                },
                {
                    "test": "${order.status === 'processing'}",
                    "schema": {
                        "type": "card",
                        "className": "order-processing border-l-4 border-blue-500",
                        "header": {
                            "title": "Order #${order.order_number}",
                            "subTitle": "In Processing",
                            "avatar": {"type": "icon", "icon": "activity", "className": "bg-blue-500 text-white"}
                        },
                        "body": [
                            {"type": "progress", "value": "${order.processing_progress}", "showLabel": true},
                            {"type": "text", "text": "Processing at: ${order.fulfillment_center}"},
                            {"type": "text", "text": "Started: ${order.processing_started | fromNow}"},
                            {"type": "text", "text": "Estimated completion: ${order.estimated_completion | fromNow}"}
                        ],
                        "actions": [
                            {
                                "type": "button",
                                "label": "Update Progress",
                                "actionType": "dialog",
                                "level": "primary"
                            },
                            {
                                "type": "button",
                                "label": "Mark Complete",
                                "actionType": "ajax",
                                "api": "put:/api/orders/${order.id}/complete",
                                "level": "success",
                                "confirmText": "Mark this order as complete?"
                            }
                        ]
                    }
                },
                {
                    "test": "${order.status === 'shipped'}",
                    "schema": {
                        "type": "card",
                        "className": "order-shipped border-l-4 border-green-500",
                        "header": {
                            "title": "Order #${order.order_number}",
                            "subTitle": "Shipped",
                            "avatar": {"type": "icon", "icon": "truck", "className": "bg-green-500 text-white"}
                        },
                        "body": [
                            {"type": "text", "text": "Tracking Number: ${order.tracking_number}", "className": "font-mono"},
                            {"type": "text", "text": "Carrier: ${order.carrier}"},
                            {"type": "text", "text": "Shipped Date: ${order.ship_date | date}"},
                            {"type": "text", "text": "Estimated Delivery: ${order.estimated_delivery | date}"}
                        ],
                        "actions": [
                            {
                                "type": "button",
                                "label": "Track Package",
                                "actionType": "link",
                                "link": "${order.tracking_url}",
                                "target": "_blank"
                            },
                            {
                                "type": "button",
                                "label": "Generate Invoice",
                                "actionType": "ajax",
                                "api": "post:/api/orders/${order.id}/generate-invoice"
                            }
                        ]
                    }
                },
                {
                    "test": "${order.status === 'delivered'}",
                    "schema": {
                        "type": "card",
                        "className": "order-delivered border-l-4 border-purple-500",
                        "header": {
                            "title": "Order #${order.order_number}",
                            "subTitle": "Delivered",
                            "avatar": {"type": "icon", "icon": "check-circle", "className": "bg-purple-500 text-white"}
                        },
                        "body": [
                            {"type": "text", "text": "Delivered on: ${order.delivery_date | date}"},
                            {"type": "text", "text": "Signed by: ${order.signed_by}"},
                            {"type": "alert", "level": "success", "body": "Order successfully completed!"}
                        ],
                        "actions": [
                            {
                                "type": "button",
                                "label": "Request Feedback",
                                "actionType": "dialog",
                                "dialog": {
                                    "title": "Request Customer Feedback",
                                    "body": {
                                        "type": "form",
                                        "api": "post:/api/orders/${order.id}/request-feedback",
                                        "body": [
                                            {"type": "textarea", "name": "feedback_message", "label": "Message", "value": "We'd love to hear about your experience with this order."}
                                        ]
                                    }
                                }
                            }
                        ]
                    }
                },
                {
                    "test": "${order.status === 'cancelled'}",
                    "schema": {
                        "type": "card",
                        "className": "order-cancelled border-l-4 border-red-500",
                        "header": {
                            "title": "Order #${order.order_number}",
                            "subTitle": "Cancelled",
                            "avatar": {"type": "icon", "icon": "x-circle", "className": "bg-red-500 text-white"}
                        },
                        "body": [
                            {"type": "text", "text": "Cancellation Reason: ${order.cancellation_reason}"},
                            {"type": "text", "text": "Cancelled Date: ${order.cancellation_date | date}"},
                            {"type": "text", "text": "Refund Status: ${order.refund_status}"}
                        ],
                        "actions": [
                            {
                                "type": "button",
                                "label": "Process Refund",
                                "actionType": "dialog",
                                "visibleOn": "${order.refund_status === 'pending'}"
                            }
                        ]
                    }
                }
            ]
        }
    ]
}
```

## Real-World Use Cases

### Multi-Tenant Application Routing
```json
{
    "type": "service",
    "api": "/api/tenant/${tenant_id}/features",
    "body": [
        {
            "type": "switch-container",
            "items": [
                {
                    "test": "${features.includes('advanced_analytics')}",
                    "schema": {
                        "type": "page",
                        "title": "Advanced Analytics Dashboard",
                        "body": [
                            {
                                "type": "tabs",
                                "tabs": [
                                    {
                                        "title": "Revenue Analytics",
                                        "body": [
                                            {"type": "chart", "config": "${advanced_revenue_chart}"}
                                        ]
                                    },
                                    {
                                        "title": "Customer Insights",
                                        "body": [
                                            {"type": "chart", "config": "${customer_segmentation_chart}"}
                                        ]
                                    },
                                    {
                                        "title": "Predictive Analytics",
                                        "body": [
                                            {"type": "chart", "config": "${predictive_models_chart}"}
                                        ]
                                    }
                                ]
                            }
                        ]
                    }
                },
                {
                    "test": "${features.includes('basic_analytics')}",
                    "schema": {
                        "type": "page",
                        "title": "Basic Analytics Dashboard",
                        "body": [
                            {
                                "type": "grid",
                                "columns": [
                                    {
                                        "md": 6,
                                        "body": [
                                            {"type": "chart", "config": "${basic_revenue_chart}"}
                                        ]
                                    },
                                    {
                                        "md": 6,
                                        "body": [
                                            {"type": "stats", "title": "Total Revenue", "value": "$${total_revenue | number}"},
                                            {"type": "stats", "title": "Total Orders", "value": "${total_orders}"}
                                        ]
                                    }
                                ]
                            }
                        ]
                    }
                },
                {
                    "test": "true",
                    "schema": {
                        "type": "page",
                        "title": "Basic Dashboard",
                        "body": [
                            {
                                "type": "alert",
                                "level": "info",
                                "title": "Upgrade for Analytics",
                                "body": "Upgrade your plan to access advanced analytics and reporting features.",
                                "actions": [
                                    {
                                        "type": "button",
                                        "label": "View Plans",
                                        "actionType": "link",
                                        "link": "/billing/plans"
                                    }
                                ]
                            },
                            {
                                "type": "service",
                                "api": "/api/dashboard/basic-stats",
                                "body": [
                                    {
                                        "type": "grid",
                                        "columns": [
                                            {
                                                "md": 4,
                                                "body": [
                                                    {"type": "stats", "title": "Orders Today", "value": "${orders_today}"}
                                                ]
                                            },
                                            {
                                                "md": 4,
                                                "body": [
                                                    {"type": "stats", "title": "Revenue Today", "value": "$${revenue_today | number}"}
                                                ]
                                            },
                                            {
                                                "md": 4,
                                                "body": [
                                                    {"type": "stats", "title": "Active Customers", "value": "${active_customers}"}
                                                ]
                                            }
                                        ]
                                    }
                                ]
                            }
                        ]
                    }
                }
            ]
        }
    ]
}
```

### Form Step Validation
```json
{
    "type": "switch-container",
    "items": [
        {
            "test": "${step === 1}",
            "schema": {
                "type": "form",
                "title": "Step 1: Basic Information",
                "body": [
                    {"type": "input-text", "name": "company_name", "label": "Company Name", "required": true},
                    {"type": "input-email", "name": "email", "label": "Email", "required": true},
                    {"type": "input-text", "name": "phone", "label": "Phone"}
                ],
                "actions": [
                    {
                        "type": "button",
                        "label": "Next",
                        "actionType": "ajax",
                        "api": "post:/api/wizard/step1",
                        "level": "primary"
                    }
                ]
            }
        },
        {
            "test": "${step === 2}",
            "schema": {
                "type": "form",
                "title": "Step 2: Business Details",
                "body": [
                    {"type": "select", "name": "industry", "label": "Industry", "source": "/api/industries", "required": true},
                    {"type": "input-number", "name": "annual_revenue", "label": "Annual Revenue", "prefix": "$"},
                    {"type": "textarea", "name": "business_description", "label": "Business Description"}
                ],
                "actions": [
                    {
                        "type": "button",
                        "label": "Previous",
                        "actionType": "ajax",
                        "api": "post:/api/wizard/previous"
                    },
                    {
                        "type": "button",
                        "label": "Next",
                        "actionType": "ajax",
                        "api": "post:/api/wizard/step2",
                        "level": "primary"
                    }
                ]
            }
        },
        {
            "test": "${step === 3}",
            "schema": {
                "type": "form",
                "title": "Step 3: Review & Submit",
                "body": [
                    {
                        "type": "alert",
                        "level": "info",
                        "body": "Please review your information before submitting."
                    },
                    {"type": "static", "label": "Company", "value": "${company_name}"},
                    {"type": "static", "label": "Email", "value": "${email}"},
                    {"type": "static", "label": "Industry", "value": "${industry}"}
                ],
                "actions": [
                    {
                        "type": "button",
                        "label": "Previous",
                        "actionType": "ajax",
                        "api": "post:/api/wizard/previous"
                    },
                    {
                        "type": "button",
                        "label": "Submit",
                        "actionType": "ajax",
                        "api": "post:/api/wizard/complete",
                        "level": "primary"
                    }
                ]
            }
        }
    ]
}
```

### Device-Responsive Layouts
```json
{
    "type": "switch-container",
    "items": [
        {
            "test": "${device.type === 'mobile'}",
            "schema": {
                "type": "div",
                "className": "mobile-layout",
                "body": [
                    {
                        "type": "nav",
                        "className": "mobile-nav fixed bottom-0 left-0 right-0",
                        "mode": "horizontal",
                        "links": [
                            {"label": "Home", "icon": "home", "to": "/"},
                            {"label": "Orders", "icon": "shopping-cart", "to": "/orders"},
                            {"label": "Menu", "icon": "menu", "to": "/menu"}
                        ]
                    },
                    {
                        "type": "div",
                        "className": "mobile-content pb-16",
                        "body": [
                            {"$ref": "#/definitions/MobileContent"}
                        ]
                    }
                ]
            }
        },
        {
            "test": "${device.type === 'tablet'}",
            "schema": {
                "type": "grid",
                "columns": [
                    {
                        "md": 3,
                        "body": [
                            {
                                "type": "nav",
                                "className": "tablet-sidebar",
                                "links": [
                                    {"label": "Dashboard", "to": "/dashboard"},
                                    {"label": "Orders", "to": "/orders"},
                                    {"label": "Customers", "to": "/customers"}
                                ]
                            }
                        ]
                    },
                    {
                        "md": 9,
                        "body": [
                            {"$ref": "#/definitions/TabletContent"}
                        ]
                    }
                ]
            }
        },
        {
            "test": "true",
            "schema": {
                "type": "div",
                "className": "desktop-layout",
                "body": [
                    {
                        "type": "nav",
                        "className": "desktop-nav",
                        "links": [
                            {"label": "Dashboard", "to": "/dashboard"},
                            {"label": "Orders", "to": "/orders"},
                            {"label": "Customers", "to": "/customers"},
                            {"label": "Reports", "to": "/reports"}
                        ]
                    },
                    {
                        "type": "div",
                        "className": "desktop-content",
                        "body": [
                            {"$ref": "#/definitions/DesktopContent"}
                        ]
                    }
                ]
            }
        }
    ]
}
```

## Advanced Patterns

### Nested Switch Containers
```json
{
    "type": "switch-container",
    "items": [
        {
            "test": "${user.authenticated === true}",
            "schema": {
                "type": "switch-container",
                "items": [
                    {
                        "test": "${page.route === 'dashboard'}",
                        "schema": {"$ref": "#/definitions/DashboardPage"}
                    },
                    {
                        "test": "${page.route === 'orders'}",
                        "schema": {"$ref": "#/definitions/OrdersPage"}
                    }
                ]
            }
        },
        {
            "test": "true",
            "schema": {"$ref": "#/definitions/LoginPage"}
        }
    ]
}
```

### Feature Flag-Based Rendering
```json
{
    "type": "switch-container",
    "items": [
        {
            "test": "${feature_flags.new_dashboard === true}",
            "schema": {"$ref": "#/definitions/NewDashboard"}
        },
        {
            "test": "true",
            "schema": {"$ref": "#/definitions/LegacyDashboard"}
        }
    ]
}
```

This SwitchContainer template provides essential conditional rendering capabilities for building adaptive, state-aware ERP interfaces that respond dynamically to user context, application state, and business conditions.