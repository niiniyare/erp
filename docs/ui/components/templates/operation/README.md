# Operation Template

**FILE PURPOSE**: Action bar template for grouping related operations and actions  
**SCOPE**: Action grouping, operation bars, toolbar creation, and button collections  
**TARGET AUDIENCE**: Developers implementing action bars, toolbars, and grouped operations

## 📋 Component Overview

Operation provides a template for grouping related actions and operations into cohesive action bars. Essential for creating consistent operation interfaces, toolbars, and action groupings in ERP systems with proper spacing and layout.

### Schema Reference
- **Primary Schema**: `OperationSchema.json`
- **Related Schemas**: `ActionSchema`
- **Base Interface**: Action bar template for operation grouping

## Basic Usage

```json
{
    "type": "operation",
    "buttons": [
        {"type": "button", "label": "Edit", "level": "primary"},
        {"type": "button", "label": "Delete", "level": "danger"},
        {"type": "button", "label": "View", "level": "link"}
    ]
}
```

## Go Type Definition

```go
type OperationProps struct {
    Type                    string              `json:"type"`
    Buttons                 []interface{}       `json:"buttons"`             // Action buttons
    Placeholder             string              `json:"placeholder"`         // Empty state text
}
```

## Operation Patterns

### Basic Action Bar
```json
{
    "type": "operation",
    "buttons": [
        {"type": "button", "label": "Save", "level": "primary"},
        {"type": "button", "label": "Cancel", "level": "default"}
    ]
}
```

### Grouped Operations with Dropdown
```json
{
    "type": "operation",
    "buttons": [
        {"type": "button", "label": "Edit", "level": "primary"},
        {
            "type": "dropdown-button",
            "label": "More Actions",
            "buttons": [
                {"type": "button", "label": "Duplicate"},
                {"type": "button", "label": "Archive"},
                {"type": "divider"},
                {"type": "button", "label": "Delete", "level": "danger"}
            ]
        }
    ]
}
```

### Conditional Operations
```json
{
    "type": "operation",
    "buttons": [
        {
            "type": "button",
            "label": "Approve",
            "level": "success",
            "visibleOn": "${status === 'pending'}"
        },
        {
            "type": "button",
            "label": "Reject",
            "level": "danger",
            "visibleOn": "${status === 'pending'}"
        },
        {
            "type": "button",
            "label": "View Details",
            "level": "link"
        }
    ]
}
```

## Real-World Use Cases

### Customer Management Operations
```json
{
    "type": "operation",
    "buttons": [
        {
            "type": "button",
            "label": "Edit Customer",
            "icon": "edit",
            "level": "primary",
            "actionType": "dialog",
            "dialog": {
                "title": "Edit Customer",
                "size": "lg",
                "body": {
                    "type": "form",
                    "api": "put:/api/customers/${id}",
                    "initApi": "/api/customers/${id}",
                    "body": [
                        {"type": "input-text", "name": "company_name", "label": "Company Name", "required": true},
                        {"type": "input-email", "name": "email", "label": "Email", "required": true},
                        {"type": "input-text", "name": "phone", "label": "Phone"}
                    ]
                }
            }
        },
        {
            "type": "dropdown-button",
            "label": "Customer Actions",
            "icon": "more-horizontal",
            "buttons": [
                {
                    "type": "button",
                    "label": "Create Order",
                    "icon": "shopping-cart",
                    "actionType": "link",
                    "link": "/orders/create?customer_id=${id}"
                },
                {
                    "type": "button",
                    "label": "Generate Invoice",
                    "icon": "file-text",
                    "actionType": "dialog",
                    "dialog": {
                        "title": "Generate Invoice",
                        "body": {
                            "type": "form",
                            "api": "post:/api/invoices",
                            "body": [
                                {"type": "hidden", "name": "customer_id", "value": "${id}"},
                                {"type": "input-date", "name": "invoice_date", "label": "Invoice Date", "value": "${TODAY()}", "required": true},
                                {"type": "input-date", "name": "due_date", "label": "Due Date", "required": true},
                                {"type": "textarea", "name": "description", "label": "Description"}
                            ]
                        }
                    }
                },
                {
                    "type": "button",
                    "label": "Send Email",
                    "icon": "mail",
                    "actionType": "dialog",
                    "dialog": {
                        "title": "Send Email to ${company_name}",
                        "body": {
                            "type": "form",
                            "api": "post:/api/customers/${id}/send-email",
                            "body": [
                                {"type": "input-text", "name": "subject", "label": "Subject", "required": true},
                                {"type": "textarea", "name": "message", "label": "Message", "required": true, "rows": 6},
                                {"type": "switch", "name": "send_copy", "label": "Send Copy", "option": "Send a copy to myself"}
                            ]
                        }
                    }
                },
                {"type": "divider"},
                {
                    "type": "button",
                    "label": "View Order History",
                    "icon": "list",
                    "actionType": "link",
                    "link": "/orders?customer_id=${id}"
                },
                {
                    "type": "button",
                    "label": "Customer Report",
                    "icon": "bar-chart",
                    "actionType": "download",
                    "api": "/api/customers/${id}/report.pdf"
                },
                {"type": "divider"},
                {
                    "type": "button",
                    "label": "Archive Customer",
                    "icon": "archive",
                    "actionType": "ajax",
                    "api": "put:/api/customers/${id}/archive",
                    "confirmText": "Archive this customer?",
                    "level": "warning"
                },
                {
                    "type": "button",
                    "label": "Delete Customer",
                    "icon": "trash",
                    "actionType": "ajax",
                    "api": "delete:/api/customers/${id}",
                    "confirmText": "Are you sure you want to permanently delete this customer and all associated data?",
                    "level": "danger"
                }
            ]
        }
    ]
}
```

### Order Processing Operations
```json
{
    "type": "operation",
    "buttons": [
        {
            "type": "button",
            "label": "Process Order",
            "icon": "play",
            "level": "primary",
            "actionType": "dialog",
            "visibleOn": "${status === 'pending'}",
            "dialog": {
                "title": "Process Order #${order_number}",
                "size": "lg",
                "body": {
                    "type": "form",
                    "api": "put:/api/orders/${id}/process",
                    "body": [
                        {"type": "select", "name": "fulfillment_center", "label": "Fulfillment Center", "source": "/api/fulfillment-centers", "required": true},
                        {"type": "select", "name": "priority", "label": "Priority", "options": ["standard", "expedited", "rush"], "value": "standard"},
                        {"type": "textarea", "name": "processing_notes", "label": "Processing Notes"}
                    ]
                }
            }
        },
        {
            "type": "button",
            "label": "Mark as Shipped",
            "icon": "truck",
            "level": "success",
            "actionType": "dialog",
            "visibleOn": "${status === 'processing'}",
            "dialog": {
                "title": "Ship Order #${order_number}",
                "body": {
                    "type": "form",
                    "api": "put:/api/orders/${id}/ship",
                    "body": [
                        {"type": "input-text", "name": "tracking_number", "label": "Tracking Number", "required": true},
                        {"type": "select", "name": "carrier", "label": "Shipping Carrier", "source": "/api/carriers", "required": true},
                        {"type": "input-date", "name": "ship_date", "label": "Ship Date", "value": "${TODAY()}", "required": true},
                        {"type": "input-date", "name": "estimated_delivery", "label": "Estimated Delivery"},
                        {"type": "switch", "name": "send_notification", "label": "Send Notification", "option": "Send tracking info to customer", "value": true}
                    ]
                }
            }
        },
        {
            "type": "dropdown-button",
            "label": "Order Actions",
            "icon": "settings",
            "buttons": [
                {
                    "type": "button",
                    "label": "Print Pick List",
                    "icon": "printer",
                    "actionType": "download",
                    "api": "/api/orders/${id}/pick-list.pdf"
                },
                {
                    "type": "button",
                    "label": "Print Packing Slip",
                    "icon": "package",
                    "actionType": "download",
                    "api": "/api/orders/${id}/packing-slip.pdf"
                },
                {
                    "type": "button",
                    "label": "Generate Invoice",
                    "icon": "file-text",
                    "actionType": "ajax",
                    "api": "post:/api/orders/${id}/generate-invoice",
                    "visibleOn": "${status === 'shipped' && !invoiced}"
                },
                {"type": "divider"},
                {
                    "type": "button",
                    "label": "Add Note",
                    "icon": "message-square",
                    "actionType": "dialog",
                    "dialog": {
                        "title": "Add Order Note",
                        "body": {
                            "type": "form",
                            "api": "post:/api/orders/${id}/notes",
                            "body": [
                                {"type": "textarea", "name": "note", "label": "Note", "required": true, "rows": 4},
                                {"type": "switch", "name": "customer_visible", "label": "Customer Visible", "option": "Show this note to the customer"}
                            ]
                        }
                    }
                },
                {
                    "type": "button",
                    "label": "Update Address",
                    "icon": "map-pin",
                    "actionType": "dialog",
                    "visibleOn": "${status === 'pending' || status === 'processing'}",
                    "dialog": {
                        "title": "Update Shipping Address",
                        "body": {
                            "type": "form",
                            "api": "put:/api/orders/${id}/shipping-address",
                            "initApi": "/api/orders/${id}/shipping-address",
                            "body": [
                                {"type": "input-text", "name": "street", "label": "Street Address", "required": true},
                                {"type": "input-text", "name": "city", "label": "City", "required": true},
                                {"type": "select", "name": "state", "label": "State", "source": "/api/states", "required": true},
                                {"type": "input-text", "name": "zip", "label": "ZIP Code", "required": true}
                            ]
                        }
                    }
                },
                {"type": "divider"},
                {
                    "type": "button",
                    "label": "Cancel Order",
                    "icon": "x-circle",
                    "actionType": "dialog",
                    "level": "danger",
                    "visibleOn": "${status === 'pending' || status === 'processing'}",
                    "dialog": {
                        "title": "Cancel Order #${order_number}",
                        "body": {
                            "type": "form",
                            "api": "put:/api/orders/${id}/cancel",
                            "body": [
                                {"type": "alert", "level": "warning", "body": "Canceling this order will stop all processing and may trigger refund procedures."},
                                {"type": "select", "name": "cancellation_reason", "label": "Cancellation Reason", "options": ["customer_request", "inventory_unavailable", "payment_failed", "other"], "required": true},
                                {"type": "textarea", "name": "cancellation_notes", "label": "Notes", "placeholder": "Additional details about the cancellation..."},
                                {"type": "switch", "name": "notify_customer", "label": "Notify Customer", "option": "Send cancellation email to customer", "value": true}
                            ]
                        }
                    }
                }
            ]
        }
    ]
}
```

### Invoice Management Operations
```json
{
    "type": "operation",
    "buttons": [
        {
            "type": "button",
            "label": "Send Invoice",
            "icon": "send",
            "level": "primary",
            "actionType": "dialog",
            "visibleOn": "${status === 'draft' || status === 'sent'}",
            "dialog": {
                "title": "Send Invoice #${invoice_number}",
                "body": {
                    "type": "form",
                    "api": "post:/api/invoices/${id}/send",
                    "body": [
                        {"type": "input-email", "name": "recipient_email", "label": "Send To", "value": "${customer_email}", "required": true},
                        {"type": "input-email", "name": "cc_emails", "label": "CC", "multiple": true},
                        {"type": "textarea", "name": "message", "label": "Message", "value": "Please find attached your invoice.", "rows": 4},
                        {"type": "switch", "name": "include_payment_link", "label": "Include Payment Link", "option": "Add online payment link to email", "value": true}
                    ]
                }
            }
        },
        {
            "type": "button",
            "label": "Record Payment",
            "icon": "dollar-sign",
            "level": "success",
            "actionType": "dialog",
            "visibleOn": "${balance_due > 0}",
            "dialog": {
                "title": "Record Payment for Invoice #${invoice_number}",
                "body": {
                    "type": "form",
                    "api": "post:/api/invoices/${id}/payments",
                    "body": [
                        {"type": "static", "label": "Invoice Amount", "value": "$${amount | number}"},
                        {"type": "static", "label": "Amount Paid", "value": "$${amount_paid | number}"},
                        {"type": "static", "label": "Balance Due", "value": "$${balance_due | number}", "className": "font-bold"},
                        {"type": "divider"},
                        {"type": "input-number", "name": "payment_amount", "label": "Payment Amount", "prefix": "$", "required": true, "max": "${balance_due}"},
                        {"type": "input-date", "name": "payment_date", "label": "Payment Date", "value": "${TODAY()}", "required": true},
                        {"type": "select", "name": "payment_method", "label": "Payment Method", "options": ["check", "wire_transfer", "credit_card", "cash", "ach"], "required": true},
                        {"type": "input-text", "name": "reference_number", "label": "Reference/Check Number"},
                        {"type": "textarea", "name": "payment_notes", "label": "Notes"}
                    ]
                }
            }
        },
        {
            "type": "dropdown-button",
            "label": "Invoice Actions",
            "icon": "more-horizontal",
            "buttons": [
                {
                    "type": "button",
                    "label": "Download PDF",
                    "icon": "download",
                    "actionType": "download",
                    "api": "/api/invoices/${id}/pdf"
                },
                {
                    "type": "button",
                    "label": "Print Invoice",
                    "icon": "printer",
                    "actionType": "download",
                    "api": "/api/invoices/${id}/print"
                },
                {
                    "type": "button",
                    "label": "Duplicate Invoice",
                    "icon": "copy",
                    "actionType": "ajax",
                    "api": "post:/api/invoices/${id}/duplicate"
                },
                {"type": "divider"},
                {
                    "type": "button",
                    "label": "View Payment History",
                    "icon": "credit-card",
                    "actionType": "dialog",
                    "dialog": {
                        "title": "Payment History - Invoice #${invoice_number}",
                        "size": "lg",
                        "body": {
                            "type": "service",
                            "api": "/api/invoices/${id}/payments",
                            "body": [
                                {
                                    "type": "table",
                                    "source": "${payments}",
                                    "columns": [
                                        {"name": "payment_date", "label": "Date", "type": "date"},
                                        {"name": "amount", "label": "Amount", "type": "number", "prefix": "$"},
                                        {"name": "payment_method", "label": "Method"},
                                        {"name": "reference_number", "label": "Reference"},
                                        {"name": "notes", "label": "Notes"}
                                    ],
                                    "placeholder": "No payments recorded"
                                }
                            ]
                        }
                    }
                },
                {
                    "type": "button",
                    "label": "Add Credit/Adjustment",
                    "icon": "minus-circle",
                    "actionType": "dialog",
                    "dialog": {
                        "title": "Add Credit/Adjustment",
                        "body": {
                            "type": "form",
                            "api": "post:/api/invoices/${id}/adjustments",
                            "body": [
                                {"type": "select", "name": "type", "label": "Adjustment Type", "options": ["credit", "discount", "correction"], "required": true},
                                {"type": "input-number", "name": "amount", "label": "Amount", "prefix": "$", "required": true},
                                {"type": "textarea", "name": "reason", "label": "Reason", "required": true}
                            ]
                        }
                    }
                },
                {"type": "divider"},
                {
                    "type": "button",
                    "label": "Void Invoice",
                    "icon": "x-circle",
                    "actionType": "ajax",
                    "api": "put:/api/invoices/${id}/void",
                    "confirmText": "Are you sure you want to void this invoice? This action cannot be undone.",
                    "level": "danger",
                    "visibleOn": "${status !== 'void' && status !== 'paid'}"
                }
            ]
        }
    ]
}
```

### Product Management Operations
```json
{
    "type": "operation",
    "buttons": [
        {
            "type": "button",
            "label": "Edit Product",
            "icon": "edit",
            "level": "primary",
            "actionType": "dialog",
            "dialog": {
                "title": "Edit Product",
                "size": "lg",
                "body": {
                    "type": "form",
                    "api": "put:/api/products/${id}",
                    "initApi": "/api/products/${id}",
                    "body": [
                        {"type": "input-text", "name": "name", "label": "Product Name", "required": true},
                        {"type": "input-text", "name": "sku", "label": "SKU", "required": true},
                        {"type": "input-number", "name": "price", "label": "Price", "prefix": "$", "min": 0},
                        {"type": "select", "name": "category_id", "label": "Category", "source": "/api/categories"},
                        {"type": "textarea", "name": "description", "label": "Description"}
                    ]
                }
            }
        },
        {
            "type": "dropdown-button",
            "label": "Inventory Actions",
            "icon": "package",
            "buttons": [
                {
                    "type": "button",
                    "label": "Update Stock",
                    "icon": "trending-up",
                    "actionType": "dialog",
                    "dialog": {
                        "title": "Update Stock - ${name}",
                        "body": {
                            "type": "form",
                            "api": "put:/api/products/${id}/stock",
                            "body": [
                                {"type": "static", "label": "Current Stock", "value": "${current_stock}"},
                                {"type": "input-number", "name": "new_stock", "label": "New Stock Level", "required": true},
                                {"type": "select", "name": "adjustment_type", "label": "Adjustment Type", "options": ["received", "sold", "damaged", "correction"], "required": true},
                                {"type": "textarea", "name": "reason", "label": "Reason for Adjustment"}
                            ]
                        }
                    }
                },
                {
                    "type": "button",
                    "label": "Set Reorder Level",
                    "icon": "alert-triangle",
                    "actionType": "dialog",
                    "dialog": {
                        "title": "Set Reorder Level",
                        "body": {
                            "type": "form",
                            "api": "put:/api/products/${id}/reorder-level",
                            "body": [
                                {"type": "input-number", "name": "reorder_level", "label": "Reorder Level", "required": true, "value": "${reorder_level}"},
                                {"type": "input-number", "name": "reorder_quantity", "label": "Reorder Quantity", "required": true, "value": "${reorder_quantity}"}
                            ]
                        }
                    }
                },
                {
                    "type": "button",
                    "label": "Create Purchase Order",
                    "icon": "shopping-cart",
                    "actionType": "link",
                    "link": "/purchase-orders/create?product_id=${id}",
                    "visibleOn": "${current_stock <= reorder_level}"
                }
            ]
        },
        {
            "type": "dropdown-button",
            "label": "More Actions",
            "icon": "more-horizontal",
            "buttons": [
                {
                    "type": "button",
                    "label": "Duplicate Product",
                    "icon": "copy",
                    "actionType": "ajax",
                    "api": "post:/api/products/${id}/duplicate"
                },
                {
                    "type": "button",
                    "label": "View Sales History",
                    "icon": "bar-chart",
                    "actionType": "link",
                    "link": "/reports/product-sales?product_id=${id}"
                },
                {
                    "type": "button",
                    "label": "Update Images",
                    "icon": "image",
                    "actionType": "dialog",
                    "dialog": {
                        "title": "Update Product Images",
                        "body": {
                            "type": "form",
                            "api": "put:/api/products/${id}/images",
                            "body": [
                                {"type": "input-image", "name": "images", "label": "Product Images", "multiple": true, "maxLength": 10}
                            ]
                        }
                    }
                },
                {"type": "divider"},
                {
                    "type": "button",
                    "label": "Archive Product",
                    "icon": "archive",
                    "actionType": "ajax",
                    "api": "put:/api/products/${id}/archive",
                    "confirmText": "Archive this product?",
                    "level": "warning"
                },
                {
                    "type": "button",
                    "label": "Delete Product",
                    "icon": "trash",
                    "actionType": "ajax",
                    "api": "delete:/api/products/${id}",
                    "confirmText": "Are you sure you want to delete this product?",
                    "level": "danger"
                }
            ]
        }
    ]
}
```

This component provides essential operation grouping functionality for ERP systems requiring consistent action bars, grouped operations, and contextual actions with proper organization and user experience patterns.