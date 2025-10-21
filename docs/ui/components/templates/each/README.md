# Each Template

**FILE PURPOSE**: Loop and iteration template for rendering dynamic lists and repeating components  
**SCOPE**: Data iteration, dynamic list rendering, and collection display patterns  
**TARGET AUDIENCE**: Developers implementing dynamic lists, grids, and repeating UI patterns

## 📋 Component Overview

Each provides powerful iteration capabilities for rendering arrays of data into repeating UI components. Essential for displaying dynamic lists, data grids, and any scenario where you need to render multiple instances of the same component structure with different data.

### Schema Reference
- **Primary Schema**: `EachSchema.json`
- **Related Schemas**: `SchemaCollection`
- **Base Interface**: Data iteration template for repeating components

## Basic Usage

```json
{
    "type": "each",
    "source": "${customers}",
    "items": [
        {
            "type": "card",
            "header": {
                "title": "${company_name}",
                "subTitle": "${contact_person}"
            },
            "body": [
                {"type": "text", "text": "Email: ${email}"},
                {"type": "text", "text": "Phone: ${phone}"}
            ]
        }
    ]
}
```

## Go Type Definition

```go
type EachProps struct {
    Type                    string              `json:"type"`
    Source                  string              `json:"source"`            // Data source expression
    Items                   []interface{}       `json:"items"`             // Template to repeat
    ItemKeyName             string              `json:"itemKeyName"`       // Custom item variable name
    IndexKeyName            string              `json:"indexKeyName"`      // Custom index variable name
    Placeholder             string              `json:"placeholder"`       // Empty state message
}
```

## Iteration Patterns

### Basic Array Iteration
```json
{
    "type": "each",
    "source": "${products}",
    "items": [
        {
            "type": "div",
            "className": "product-item p-4 border rounded mb-2",
            "body": [
                {
                    "type": "text",
                    "text": "${name}",
                    "className": "font-bold text-lg"
                },
                {
                    "type": "text",
                    "text": "Price: $${price}",
                    "className": "text-green-600"
                },
                {
                    "type": "text",
                    "text": "Stock: ${stock_quantity}",
                    "className": "text-gray-500"
                }
            ]
        }
    ],
    "placeholder": "No products available"
}
```

### Custom Variable Names
```json
{
    "type": "each",
    "source": "${order_items}",
    "itemKeyName": "item",
    "indexKeyName": "position",
    "items": [
        {
            "type": "div",
            "className": "order-item flex justify-between p-3 border-b",
            "body": [
                {
                    "type": "text",
                    "text": "${position + 1}. ${item.product_name}",
                    "className": "font-medium"
                },
                {
                    "type": "text",
                    "text": "Qty: ${item.quantity}",
                    "className": "text-gray-600"
                },
                {
                    "type": "text",
                    "text": "$${item.total_price}",
                    "className": "font-bold text-green-600"
                }
            ]
        }
    ]
}
```

### Complex Component Iteration
```json
{
    "type": "each",
    "source": "${team_members}",
    "items": [
        {
            "type": "card",
            "className": "team-member-card mb-4",
            "header": {
                "title": "${first_name} ${last_name}",
                "subTitle": "${job_title}",
                "avatar": {
                    "type": "image",
                    "src": "${profile_picture || '/default-avatar.png'}",
                    "className": "w-12 h-12 rounded-full"
                }
            },
            "body": [
                {
                    "type": "grid",
                    "columns": [
                        {
                            "md": 6,
                            "body": [
                                {"type": "text", "text": "Email: ${email}"},
                                {"type": "text", "text": "Phone: ${phone}"},
                                {"type": "text", "text": "Department: ${department}"}
                            ]
                        },
                        {
                            "md": 6,
                            "body": [
                                {"type": "text", "text": "Start Date: ${start_date | date}"},
                                {"type": "text", "text": "Status: ${status}"},
                                {"type": "progress", "value": "${performance_score}", "showLabel": true}
                            ]
                        }
                    ]
                }
            ],
            "actions": [
                {
                    "type": "button",
                    "label": "View Profile",
                    "actionType": "link",
                    "link": "/employees/${id}"
                },
                {
                    "type": "button",
                    "label": "Send Message",
                    "actionType": "dialog"
                }
            ]
        }
    ]
}
```

## Real-World Use Cases

### Customer List with Actions
```json
{
    "type": "service",
    "api": "/api/customers",
    "body": [
        {
            "type": "div",
            "className": "customers-grid grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4",
            "body": [
                {
                    "type": "each",
                    "source": "${customers}",
                    "items": [
                        {
                            "type": "card",
                            "className": "customer-card hover:shadow-lg transition-shadow",
                            "header": {
                                "title": "${company_name}",
                                "subTitle": "${contact_person}",
                                "avatar": {
                                    "type": "icon",
                                    "icon": "building",
                                    "className": "bg-blue-500 text-white"
                                }
                            },
                            "body": [
                                {
                                    "type": "div",
                                    "className": "space-y-2",
                                    "body": [
                                        {
                                            "type": "div",
                                            "className": "flex items-center text-sm text-gray-600",
                                            "body": [
                                                {"type": "icon", "icon": "mail", "className": "mr-2 w-4 h-4"},
                                                {"type": "text", "text": "${email}"}
                                            ]
                                        },
                                        {
                                            "type": "div",
                                            "className": "flex items-center text-sm text-gray-600",
                                            "body": [
                                                {"type": "icon", "icon": "phone", "className": "mr-2 w-4 h-4"},
                                                {"type": "text", "text": "${phone}"}
                                            ]
                                        },
                                        {
                                            "type": "div",
                                            "className": "flex items-center text-sm text-gray-600",
                                            "body": [
                                                {"type": "icon", "icon": "map-pin", "className": "mr-2 w-4 h-4"},
                                                {"type": "text", "text": "${city}, ${state}"}
                                            ]
                                        }
                                    ]
                                },
                                {
                                    "type": "divider",
                                    "className": "my-3"
                                },
                                {
                                    "type": "grid",
                                    "columns": [
                                        {
                                            "md": 6,
                                            "body": [
                                                {
                                                    "type": "stats",
                                                    "title": "Total Orders",
                                                    "value": "${total_orders}",
                                                    "className": "text-center"
                                                }
                                            ]
                                        },
                                        {
                                            "md": 6,
                                            "body": [
                                                {
                                                    "type": "stats",
                                                    "title": "Total Spent",
                                                    "value": "$${total_spent | number}",
                                                    "className": "text-center"
                                                }
                                            ]
                                        }
                                    ]
                                }
                            ],
                            "actions": [
                                {
                                    "type": "button",
                                    "label": "View Details",
                                    "actionType": "link",
                                    "link": "/customers/${id}",
                                    "level": "primary"
                                },
                                {
                                    "type": "dropdown-button",
                                    "label": "Actions",
                                    "buttons": [
                                        {
                                            "type": "button",
                                            "label": "Create Order",
                                            "actionType": "link",
                                            "link": "/orders/create?customer_id=${id}",
                                            "icon": "shopping-cart"
                                        },
                                        {
                                            "type": "button",
                                            "label": "Send Email",
                                            "actionType": "dialog",
                                            "icon": "mail",
                                            "dialog": {
                                                "title": "Send Email to ${company_name}",
                                                "body": {
                                                    "type": "form",
                                                    "api": "post:/api/customers/${id}/send-email",
                                                    "body": [
                                                        {
                                                            "type": "input-text",
                                                            "name": "subject",
                                                            "label": "Subject",
                                                            "required": true
                                                        },
                                                        {
                                                            "type": "textarea",
                                                            "name": "message",
                                                            "label": "Message",
                                                            "required": true,
                                                            "rows": 5
                                                        }
                                                    ]
                                                }
                                            }
                                        },
                                        {"type": "divider"},
                                        {
                                            "type": "button",
                                            "label": "Delete Customer",
                                            "actionType": "ajax",
                                            "api": "delete:/api/customers/${id}",
                                            "confirmText": "Delete ${company_name}?",
                                            "level": "danger",
                                            "icon": "trash"
                                        }
                                    ]
                                }
                            ]
                        }
                    ],
                    "placeholder": "No customers found. Add your first customer to get started."
                }
            ]
        }
    ]
}
```

### Invoice Line Items
```json
{
    "type": "div",
    "className": "invoice-items",
    "body": [
        {
            "type": "div",
            "className": "items-header grid grid-cols-12 gap-4 p-3 bg-gray-100 font-bold border-b",
            "body": [
                {"type": "text", "text": "#", "className": "col-span-1"},
                {"type": "text", "text": "Description", "className": "col-span-5"},
                {"type": "text", "text": "Qty", "className": "col-span-2 text-center"},
                {"type": "text", "text": "Unit Price", "className": "col-span-2 text-right"},
                {"type": "text", "text": "Total", "className": "col-span-2 text-right"}
            ]
        },
        {
            "type": "each",
            "source": "${invoice.line_items}",
            "itemKeyName": "item",
            "indexKeyName": "index",
            "items": [
                {
                    "type": "div",
                    "className": "item-row grid grid-cols-12 gap-4 p-3 border-b hover:bg-gray-50",
                    "body": [
                        {
                            "type": "text",
                            "text": "${index + 1}",
                            "className": "col-span-1 text-gray-500"
                        },
                        {
                            "type": "div",
                            "className": "col-span-5",
                            "body": [
                                {
                                    "type": "text",
                                    "text": "${item.description}",
                                    "className": "font-medium"
                                },
                                {
                                    "type": "text",
                                    "text": "${item.product_code}",
                                    "className": "text-sm text-gray-500",
                                    "visibleOn": "${item.product_code}"
                                }
                            ]
                        },
                        {
                            "type": "text",
                            "text": "${item.quantity}",
                            "className": "col-span-2 text-center"
                        },
                        {
                            "type": "text",
                            "text": "$${item.unit_price | number}",
                            "className": "col-span-2 text-right"
                        },
                        {
                            "type": "text",
                            "text": "$${item.total_price | number}",
                            "className": "col-span-2 text-right font-bold"
                        }
                    ]
                }
            ]
        },
        {
            "type": "div",
            "className": "invoice-totals mt-4 border-t pt-4",
            "body": [
                {
                    "type": "grid",
                    "columns": [
                        {"md": 8, "body": []},
                        {
                            "md": 4,
                            "body": [
                                {
                                    "type": "div",
                                    "className": "space-y-2",
                                    "body": [
                                        {
                                            "type": "div",
                                            "className": "flex justify-between",
                                            "body": [
                                                {"type": "text", "text": "Subtotal:"},
                                                {"type": "text", "text": "$${invoice.subtotal | number}"}
                                            ]
                                        },
                                        {
                                            "type": "div",
                                            "className": "flex justify-between",
                                            "body": [
                                                {"type": "text", "text": "Tax:"},
                                                {"type": "text", "text": "$${invoice.tax_amount | number}"}
                                            ]
                                        },
                                        {
                                            "type": "div",
                                            "className": "flex justify-between font-bold text-lg border-t pt-2",
                                            "body": [
                                                {"type": "text", "text": "Total:"},
                                                {"type": "text", "text": "$${invoice.total_amount | number}"}
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
    ]
}
```

### Order Processing Pipeline
```json
{
    "type": "service",
    "api": "/api/orders/pipeline",
    "body": [
        {
            "type": "div",
            "className": "order-pipeline",
            "body": [
                {
                    "type": "each",
                    "source": "${pipeline_stages}",
                    "itemKeyName": "stage",
                    "items": [
                        {
                            "type": "div",
                            "className": "pipeline-stage mb-8",
                            "body": [
                                {
                                    "type": "div",
                                    "className": "stage-header flex items-center mb-4",
                                    "body": [
                                        {
                                            "type": "icon",
                                            "icon": "${stage.icon}",
                                            "className": "mr-3 p-2 rounded-full ${stage.color} text-white"
                                        },
                                        {
                                            "type": "div",
                                            "body": [
                                                {
                                                    "type": "text",
                                                    "text": "${stage.name}",
                                                    "className": "text-lg font-bold"
                                                },
                                                {
                                                    "type": "text",
                                                    "text": "${stage.orders.length} orders",
                                                    "className": "text-gray-500"
                                                }
                                            ]
                                        }
                                    ]
                                },
                                {
                                    "type": "div",
                                    "className": "stage-orders grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4",
                                    "body": [
                                        {
                                            "type": "each",
                                            "source": "${stage.orders}",
                                            "itemKeyName": "order",
                                            "items": [
                                                {
                                                    "type": "card",
                                                    "className": "order-card border-l-4 ${stage.border_color}",
                                                    "header": {
                                                        "title": "Order #${order.order_number}",
                                                        "subTitle": "${order.customer_name}"
                                                    },
                                                    "body": [
                                                        {
                                                            "type": "div",
                                                            "className": "order-details space-y-1 text-sm",
                                                            "body": [
                                                                {"type": "text", "text": "Total: $${order.total_amount | number}"},
                                                                {"type": "text", "text": "Items: ${order.item_count}"},
                                                                {"type": "text", "text": "Date: ${order.order_date | date}"},
                                                                {"type": "text", "text": "Priority: ${order.priority}", "className": "${order.priority === 'high' ? 'text-red-600 font-bold' : ''}"}
                                                            ]
                                                        }
                                                    ],
                                                    "actions": [
                                                        {
                                                            "type": "button",
                                                            "label": "View Order",
                                                            "actionType": "link",
                                                            "link": "/orders/${order.id}",
                                                            "size": "sm"
                                                        },
                                                        {
                                                            "type": "button",
                                                            "label": "${stage.action_label}",
                                                            "actionType": "dialog",
                                                            "level": "primary",
                                                            "size": "sm",
                                                            "visibleOn": "${stage.action_label}",
                                                            "dialog": {
                                                                "title": "${stage.action_label} - Order #${order.order_number}",
                                                                "body": {
                                                                    "type": "form",
                                                                    "api": "put:/api/orders/${order.id}/advance-stage",
                                                                    "body": [
                                                                        {
                                                                            "type": "textarea",
                                                                            "name": "notes",
                                                                            "label": "Processing Notes",
                                                                            "placeholder": "Add any notes about this stage transition..."
                                                                        }
                                                                    ]
                                                                }
                                                            }
                                                        }
                                                    ]
                                                }
                                            ],
                                            "placeholder": "No orders in this stage"
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

### Dynamic Dashboard Widgets
```json
{
    "type": "service",
    "api": "/api/dashboard/widgets",
    "body": [
        {
            "type": "div",
            "className": "dashboard-widgets grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6",
            "body": [
                {
                    "type": "each",
                    "source": "${widgets}",
                    "itemKeyName": "widget",
                    "items": [
                        {
                            "type": "card",
                            "className": "widget-card ${widget.size === 'large' ? 'lg:col-span-2' : ''}",
                            "header": {
                                "title": "${widget.title}",
                                "subTitle": "${widget.description}",
                                "actions": [
                                    {
                                        "type": "button",
                                        "icon": "refresh-cw",
                                        "actionType": "ajax",
                                        "api": "/api/widgets/${widget.id}/refresh",
                                        "size": "sm",
                                        "level": "link"
                                    },
                                    {
                                        "type": "dropdown-button",
                                        "icon": "more-horizontal",
                                        "size": "sm",
                                        "level": "link",
                                        "buttons": [
                                            {"type": "button", "label": "Configure", "actionType": "dialog"},
                                            {"type": "button", "label": "Move", "actionType": "dialog"},
                                            {"type": "divider"},
                                            {"type": "button", "label": "Remove", "actionType": "ajax", "api": "delete:/api/widgets/${widget.id}", "level": "danger"}
                                        ]
                                    }
                                ]
                            },
                            "body": [
                                {
                                    "type": "switch-container",
                                    "items": [
                                        {
                                            "test": "${widget.type === 'stats'}",
                                            "schema": {
                                                "type": "stats",
                                                "title": "${widget.stats.title}",
                                                "value": "${widget.stats.value}",
                                                "className": "text-center ${widget.stats.color}"
                                            }
                                        },
                                        {
                                            "test": "${widget.type === 'chart'}",
                                            "schema": {
                                                "type": "chart",
                                                "config": "${widget.chart_config}",
                                                "data": "${widget.chart_data}"
                                            }
                                        },
                                        {
                                            "test": "${widget.type === 'list'}",
                                            "schema": {
                                                "type": "list",
                                                "source": "${widget.list_data}",
                                                "listItem": "${widget.list_template}"
                                            }
                                        }
                                    ]
                                }
                            ]
                        }
                    ],
                    "placeholder": "No widgets configured. Add widgets to customize your dashboard."
                }
            ]
        }
    ]
}
```

## Advanced Patterns

### Nested Iteration
```json
{
    "type": "each",
    "source": "${departments}",
    "itemKeyName": "dept",
    "items": [
        {
            "type": "card",
            "header": {
                "title": "${dept.name}",
                "subTitle": "${dept.employees.length} employees"
            },
            "body": [
                {
                    "type": "each",
                    "source": "${dept.employees}",
                    "itemKeyName": "employee",
                    "items": [
                        {
                            "type": "div",
                            "className": "employee-item flex items-center p-2 border-b",
                            "body": [
                                {"type": "text", "text": "${employee.name}"},
                                {"type": "text", "text": "${employee.title}", "className": "ml-auto text-gray-500"}
                            ]
                        }
                    ]
                }
            ]
        }
    ]
}
```

### Conditional Rendering Within Iteration
```json
{
    "type": "each",
    "source": "${notifications}",
    "items": [
        {
            "type": "alert",
            "level": "${severity}",
            "visibleOn": "${!dismissed}",
            "body": [
                {
                    "type": "text",
                    "text": "${message}"
                }
            ],
            "actions": [
                {
                    "type": "button",
                    "label": "Dismiss",
                    "actionType": "ajax",
                    "api": "put:/api/notifications/${id}/dismiss"
                }
            ]
        }
    ]
}
```

This Each template provides powerful iteration capabilities essential for building dynamic, data-driven ERP interfaces with flexible rendering patterns and comprehensive customization options.