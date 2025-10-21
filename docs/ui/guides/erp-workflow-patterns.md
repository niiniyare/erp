# ERP Workflow Patterns

**FILE PURPOSE**: Common ERP business process patterns and implementation guides  
**SCOPE**: Business workflows, process automation, and multi-step operations  
**TARGET AUDIENCE**: Developers implementing ERP business processes and workflow automation

## 📋 Guide Overview

This guide demonstrates common ERP workflow patterns using the component library. It covers complete business processes from lead generation to order fulfillment, showing how to implement complex multi-step workflows with proper state management and user experience.

## 🔄 Core Workflow Patterns

### 1. Lead-to-Customer Conversion
### 2. Quote-to-Order Process
### 3. Order Fulfillment Workflow
### 4. Invoice-to-Payment Cycle
### 5. Inventory Management Process
### 6. Employee Onboarding Workflow

## 📊 Lead-to-Customer Conversion Workflow

### Lead Capture Form
```json
{
    "type": "page",
    "title": "Lead Capture",
    "body": [
        {
            "type": "form",
            "api": "post:/api/leads",
            "title": "Contact Information",
            "body": [
                {"type": "input-text", "name": "company_name", "label": "Company Name", "required": true},
                {"type": "input-text", "name": "contact_person", "label": "Contact Person", "required": true},
                {"type": "input-email", "name": "email", "label": "Email Address", "required": true},
                {"type": "input-text", "name": "phone", "label": "Phone Number"},
                {"type": "select", "name": "industry", "label": "Industry", "source": "/api/industries"},
                {"type": "select", "name": "company_size", "label": "Company Size", "options": ["1-10", "11-50", "51-200", "201-500", "500+"]},
                {"type": "textarea", "name": "requirements", "label": "Requirements", "placeholder": "Tell us about your needs..."},
                {"type": "select", "name": "budget_range", "label": "Budget Range", "options": ["<$10K", "$10K-$50K", "$50K-$100K", "$100K+"]},
                {"type": "select", "name": "timeline", "label": "Implementation Timeline", "options": ["Immediate", "1-3 months", "3-6 months", "6+ months"]}
            ],
            "actions": [
                {"type": "submit", "label": "Submit Inquiry", "level": "primary"}
            ],
            "redirect": "/leads/thank-you"
        }
    ]
}
```

### Lead Management Dashboard
```json
{
    "type": "page",
    "title": "Lead Management",
    "subTitle": "Manage and qualify potential customers",
    "initApi": "/api/leads/dashboard",
    
    "aside": [
        {
            "type": "panel",
            "title": "Lead Pipeline",
            "body": [
                {
                    "type": "service",
                    "api": "/api/leads/pipeline-stats",
                    "interval": 60000,
                    "body": [
                        {"type": "stats", "title": "New Leads", "value": "${new_leads}", "className": "text-blue-600"},
                        {"type": "stats", "title": "Qualified", "value": "${qualified_leads}", "className": "text-green-600"},
                        {"type": "stats", "title": "In Progress", "value": "${in_progress_leads}", "className": "text-orange-600"},
                        {"type": "stats", "title": "Converted", "value": "${converted_leads}", "className": "text-purple-600"},
                        {"type": "divider"},
                        {"type": "stats", "title": "Conversion Rate", "value": "${conversion_rate}%", "className": "text-indigo-600"}
                    ]
                }
            ]
        }
    ],
    
    "body": [
        {
            "type": "crud2",
            "api": "/api/leads",
            "title": "Lead Pipeline",
            
            "headerToolbar": [
                {
                    "type": "button",
                    "label": "Add Lead",
                    "actionType": "dialog",
                    "level": "primary",
                    "dialog": {
                        "title": "Add New Lead",
                        "body": {
                            "type": "form",
                            "api": "post:/api/leads",
                            "body": [
                                {"type": "input-text", "name": "company_name", "label": "Company Name", "required": true},
                                {"type": "input-text", "name": "contact_person", "label": "Contact Person", "required": true},
                                {"type": "input-email", "name": "email", "label": "Email", "required": true},
                                {"type": "input-text", "name": "phone", "label": "Phone"},
                                {"type": "select", "name": "source", "label": "Lead Source", "options": ["website", "referral", "trade_show", "cold_call", "social_media"]}
                            ]
                        }
                    }
                }
            ],
            
            "columns": [
                {"name": "company_name", "label": "Company", "sortable": true},
                {"name": "contact_person", "label": "Contact"},
                {"name": "email", "label": "Email", "type": "email"},
                {"name": "phone", "label": "Phone"},
                {"name": "source", "label": "Source"},
                {"name": "score", "label": "Score", "type": "progress", "showLabel": true},
                {"name": "status", "label": "Status", "type": "status"},
                {"name": "assigned_to", "label": "Assigned To"},
                {"name": "created_date", "label": "Created", "type": "date"},
                {
                    "type": "operation",
                    "label": "Actions",
                    "buttons": [
                        {
                            "type": "button",
                            "label": "Qualify",
                            "actionType": "dialog",
                            "level": "primary",
                            "visibleOn": "${status === 'new'}",
                            "dialog": {
                                "title": "Qualify Lead",
                                "body": {
                                    "type": "form",
                                    "api": "put:/api/leads/${id}/qualify",
                                    "body": [
                                        {"type": "textarea", "name": "qualification_notes", "label": "Qualification Notes", "required": true},
                                        {"type": "select", "name": "qualified_status", "label": "Status", "options": ["qualified", "disqualified"], "required": true},
                                        {"type": "users-select", "name": "assigned_to", "label": "Assign To Sales Rep", "source": "/api/users?role=sales"}
                                    ]
                                }
                            }
                        },
                        {
                            "type": "button",
                            "label": "Convert to Customer",
                            "actionType": "dialog",
                            "level": "success",
                            "visibleOn": "${status === 'qualified'}",
                            "dialog": {
                                "title": "Convert to Customer",
                                "body": {
                                    "type": "form",
                                    "api": "post:/api/leads/${id}/convert",
                                    "body": [
                                        {"type": "alert", "level": "info", "body": "This will create a new customer record and move the lead to converted status."},
                                        {"type": "textarea", "name": "conversion_notes", "label": "Conversion Notes"}
                                    ]
                                }
                            }
                        },
                        {
                            "type": "dropdown-button",
                            "label": "More",
                            "buttons": [
                                {"type": "button", "label": "Add Note", "actionType": "dialog"},
                                {"type": "button", "label": "Schedule Call", "actionType": "dialog"},
                                {"type": "button", "label": "Send Email", "actionType": "dialog"},
                                {"type": "divider"},
                                {"type": "button", "label": "Delete", "actionType": "ajax", "api": "delete:/api/leads/${id}", "confirmText": "Delete this lead?", "level": "danger"}
                            ]
                        }
                    ]
                }
            ]
        }
    ]
}
```

## 💰 Quote-to-Order Process

### Quote Builder Wizard
```json
{
    "type": "page",
    "title": "Create Quote",
    "subTitle": "Build a quote for ${customer.company_name}",
    "initApi": "/api/customers/${customer_id}",
    
    "body": [
        {
            "type": "wizard",
            "steps": [
                {
                    "title": "Customer Information",
                    "body": [
                        {
                            "type": "service",
                            "api": "/api/customers/${customer_id}",
                            "body": [
                                {"type": "static", "name": "company_name", "label": "Company"},
                                {"type": "static", "name": "contact_person", "label": "Contact"},
                                {"type": "static", "name": "email", "label": "Email"},
                                {"type": "static", "name": "billing_address", "label": "Billing Address", "tpl": "${street}, ${city}, ${state} ${zip}"}
                            ]
                        },
                        {
                            "type": "form",
                            "body": [
                                {"type": "input-date", "name": "quote_date", "label": "Quote Date", "value": "${TODAY()}", "required": true},
                                {"type": "input-date", "name": "valid_until", "label": "Valid Until", "required": true},
                                {"type": "select", "name": "payment_terms", "label": "Payment Terms", "options": ["Net 30", "Net 15", "Due on Receipt"], "value": "Net 30"},
                                {"type": "textarea", "name": "notes", "label": "Quote Notes"}
                            ]
                        }
                    ]
                },
                {
                    "title": "Products & Services",
                    "body": [
                        {
                            "type": "form",
                            "body": [
                                {
                                    "type": "combo",
                                    "name": "line_items",
                                    "label": "Quote Items",
                                    "multiple": true,
                                    "addable": true,
                                    "removable": true,
                                    "items": [
                                        {
                                            "type": "select",
                                            "name": "product_id",
                                            "label": "Product/Service",
                                            "source": "/api/products",
                                            "required": true,
                                            "autoFill": {
                                                "api": "/api/products/${value}",
                                                "fillMapping": {
                                                    "description": "name",
                                                    "unit_price": "price"
                                                }
                                            }
                                        },
                                        {"type": "textarea", "name": "description", "label": "Description"},
                                        {"type": "input-number", "name": "quantity", "label": "Quantity", "min": 1, "required": true, "value": 1},
                                        {"type": "input-number", "name": "unit_price", "label": "Unit Price", "prefix": "$", "min": 0, "required": true},
                                        {"type": "input-number", "name": "discount", "label": "Discount %", "suffix": "%", "min": 0, "max": 100, "value": 0}
                                    ]
                                }
                            ]
                        }
                    ]
                },
                {
                    "title": "Pricing & Totals",
                    "body": [
                        {
                            "type": "service",
                            "api": "/api/quotes/calculate",
                            "data": {"line_items": "${line_items}"},
                            "body": [
                                {
                                    "type": "table",
                                    "source": "${line_items}",
                                    "title": "Quote Summary",
                                    "columns": [
                                        {"name": "description", "label": "Item"},
                                        {"name": "quantity", "label": "Qty", "type": "number"},
                                        {"name": "unit_price", "label": "Unit Price", "type": "number", "prefix": "$"},
                                        {"name": "discount", "label": "Discount", "type": "number", "suffix": "%"},
                                        {"name": "line_total", "label": "Total", "type": "number", "prefix": "$"}
                                    ]
                                },
                                {
                                    "type": "grid",
                                    "columns": [
                                        {"md": 8, "body": []},
                                        {
                                            "md": 4,
                                            "body": [
                                                {"type": "static", "label": "Subtotal", "value": "$${subtotal | number}"},
                                                {"type": "static", "label": "Tax", "value": "$${tax_amount | number}"},
                                                {"type": "static", "label": "Total", "value": "$${total_amount | number}", "className": "font-bold text-lg"}
                                            ]
                                        }
                                    ]
                                }
                            ]
                        }
                    ]
                },
                {
                    "title": "Review & Send",
                    "body": [
                        {
                            "type": "form",
                            "api": "post:/api/quotes",
                            "body": [
                                {"type": "switch", "name": "send_immediately", "label": "Send Quote", "option": "Send quote to customer immediately", "value": true},
                                {"type": "textarea", "name": "cover_message", "label": "Cover Message", "placeholder": "Message to include with the quote...", "visibleOn": "${send_immediately}"},
                                {"type": "checkboxes", "name": "cc_recipients", "label": "CC Recipients", "source": "/api/team-members", "visibleOn": "${send_immediately}"}
                            ],
                            "actions": [
                                {"type": "button", "label": "Save as Draft", "actionType": "ajax", "api": "post:/api/quotes?status=draft"},
                                {"type": "submit", "label": "Create Quote", "level": "primary"}
                            ]
                        }
                    ]
                }
            ]
        }
    ]
}
```

### Quote Management Interface
```json
{
    "type": "page",
    "title": "Quote Management",
    "subTitle": "Manage customer quotes and proposals",
    
    "body": [
        {
            "type": "crud2",
            "api": "/api/quotes",
            "title": "Customer Quotes",
            
            "headerToolbar": [
                {
                    "type": "button",
                    "label": "New Quote",
                    "actionType": "link",
                    "link": "/quotes/create",
                    "level": "primary"
                },
                {
                    "type": "dropdown-button",
                    "label": "Templates",
                    "buttons": [
                        {"type": "button", "label": "Standard Service Quote", "actionType": "link", "link": "/quotes/create?template=service"},
                        {"type": "button", "label": "Product Quote", "actionType": "link", "link": "/quotes/create?template=product"},
                        {"type": "button", "label": "Custom Solution", "actionType": "link", "link": "/quotes/create?template=custom"}
                    ]
                }
            ],
            
            "filter": {
                "body": [
                    {"type": "select", "name": "status", "label": "Status", "options": ["draft", "sent", "viewed", "accepted", "declined", "expired"]},
                    {"type": "input-date-range", "name": "quote_date_range", "label": "Quote Date"},
                    {"type": "users-select", "name": "created_by", "label": "Created By", "source": "/api/users"}
                ]
            },
            
            "columns": [
                {"name": "quote_number", "label": "Quote #", "type": "link", "href": "/quotes/${id}"},
                {"name": "customer_name", "label": "Customer", "sortable": true},
                {"name": "quote_date", "label": "Quote Date", "type": "date"},
                {"name": "valid_until", "label": "Valid Until", "type": "date", "classNameExpr": "${DATETODAY(valid_until) < 0 ? 'text-red-600' : ''}"},
                {"name": "total_amount", "label": "Amount", "type": "number", "prefix": "$", "sortable": true},
                {"name": "status", "label": "Status", "type": "status"},
                {"name": "created_by", "label": "Sales Rep"},
                {
                    "type": "operation",
                    "label": "Actions",
                    "buttons": [
                        {
                            "type": "button",
                            "label": "Send",
                            "actionType": "ajax",
                            "api": "post:/api/quotes/${id}/send",
                            "level": "primary",
                            "visibleOn": "${status === 'draft'}"
                        },
                        {
                            "type": "button",
                            "label": "Convert to Order",
                            "actionType": "dialog",
                            "level": "success",
                            "visibleOn": "${status === 'accepted'}",
                            "dialog": {
                                "title": "Convert Quote to Order",
                                "body": {
                                    "type": "form",
                                    "api": "post:/api/quotes/${id}/convert-to-order",
                                    "body": [
                                        {"type": "alert", "level": "info", "body": "This will create a new order based on this quote."},
                                        {"type": "input-date", "name": "order_date", "label": "Order Date", "value": "${TODAY()}", "required": true},
                                        {"type": "textarea", "name": "order_notes", "label": "Order Notes"}
                                    ]
                                }
                            }
                        },
                        {
                            "type": "dropdown-button",
                            "label": "More",
                            "buttons": [
                                {"type": "button", "label": "Duplicate", "actionType": "ajax", "api": "post:/api/quotes/${id}/duplicate"},
                                {"type": "button", "label": "Download PDF", "actionType": "download", "api": "/api/quotes/${id}/pdf"},
                                {"type": "button", "label": "Email Quote", "actionType": "dialog"},
                                {"type": "divider"},
                                {"type": "button", "label": "Delete", "actionType": "ajax", "api": "delete:/api/quotes/${id}", "confirmText": "Delete this quote?", "level": "danger"}
                            ]
                        }
                    ]
                }
            ]
        }
    ]
}
```

## 📦 Order Fulfillment Workflow

### Order Processing Dashboard
```json
{
    "type": "page",
    "title": "Order Fulfillment",
    "subTitle": "Process and track order fulfillment",
    "initApi": "/api/fulfillment/dashboard",
    
    "body": [
        {
            "type": "tabs",
            "tabs": [
                {
                    "title": "Pending Orders",
                    "badge": {"mode": "text", "text": "${pending_count}", "className": "bg-orange-500"},
                    "body": [
                        {
                            "type": "service",
                            "api": "/api/orders?status=pending",
                            "body": [
                                {
                                    "type": "cards",
                                    "source": "${orders}",
                                    "card": {
                                        "header": {
                                            "title": "Order #${order_number}",
                                            "subTitle": "${customer_name}",
                                            "avatar": {"type": "icon", "icon": "shopping-cart", "className": "bg-orange-500 text-white"}
                                        },
                                        "body": [
                                            {"type": "text", "text": "Total: $${total_amount}"},
                                            {"type": "text", "text": "Items: ${item_count}"},
                                            {"type": "text", "text": "Priority: ${priority}", "className": "${priority === 'high' ? 'text-red-600 font-bold' : ''}"},
                                            {"type": "text", "text": "Order Date: ${order_date | date}"}
                                        ],
                                        "actions": [
                                            {
                                                "type": "button",
                                                "label": "Process Order",
                                                "actionType": "dialog",
                                                "level": "primary",
                                                "dialog": {
                                                    "title": "Process Order #${order_number}",
                                                    "size": "lg",
                                                    "body": {
                                                        "$ref": "#/definitions/OrderProcessingForm"
                                                    }
                                                }
                                            }
                                        ]
                                    }
                                }
                            ]
                        }
                    ]
                },
                {
                    "title": "In Progress",
                    "badge": {"mode": "text", "text": "${processing_count}", "className": "bg-blue-500"},
                    "body": [
                        {
                            "type": "service",
                            "api": "/api/orders?status=processing",
                            "body": [
                                {
                                    "type": "list",
                                    "source": "${orders}",
                                    "listItem": {
                                        "title": "Order #${order_number}",
                                        "subTitle": "${customer_name} • ${order_date | fromNow}",
                                        "desc": "Total: $${total_amount} • Processing at: ${fulfillment_center}",
                                        "remark": {
                                            "type": "progress",
                                            "value": "${fulfillment_progress}",
                                            "showLabel": true
                                        },
                                        "actions": [
                                            {"type": "button", "label": "Update Status", "actionType": "dialog", "level": "primary"},
                                            {"type": "button", "label": "View Details", "actionType": "link", "link": "/orders/${id}"}
                                        ]
                                    }
                                }
                            ]
                        }
                    ]
                },
                {
                    "title": "Ready to Ship",
                    "badge": {"mode": "text", "text": "${ready_to_ship_count}", "className": "bg-green-500"},
                    "body": [
                        {
                            "type": "service",
                            "api": "/api/orders?status=ready_to_ship",
                            "body": [
                                {
                                    "type": "table",
                                    "source": "${orders}",
                                    "columns": [
                                        {"name": "order_number", "label": "Order #"},
                                        {"name": "customer_name", "label": "Customer"},
                                        {"name": "shipping_address", "label": "Ship To"},
                                        {"name": "shipping_method", "label": "Method"},
                                        {"name": "total_weight", "label": "Weight"},
                                        {
                                            "type": "operation",
                                            "buttons": [
                                                {
                                                    "type": "button",
                                                    "label": "Generate Shipping Label",
                                                    "actionType": "ajax",
                                                    "api": "post:/api/orders/${id}/shipping-label",
                                                    "level": "primary"
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
    ],
    
    "definitions": {
        "OrderProcessingForm": {
            "type": "form",
            "api": "put:/api/orders/${id}/process",
            "initApi": "/api/orders/${id}/processing-details",
            "body": [
                {
                    "type": "tabs",
                    "tabs": [
                        {
                            "title": "Inventory Check",
                            "body": [
                                {
                                    "type": "table",
                                    "source": "${order_items}",
                                    "title": "Item Availability",
                                    "columns": [
                                        {"name": "product_name", "label": "Product"},
                                        {"name": "quantity_ordered", "label": "Ordered", "type": "number"},
                                        {"name": "quantity_available", "label": "Available", "type": "number"},
                                        {"name": "availability_status", "label": "Status", "type": "status"}
                                    ]
                                }
                            ]
                        },
                        {
                            "title": "Fulfillment",
                            "body": [
                                {"type": "select", "name": "fulfillment_center", "label": "Fulfillment Center", "source": "/api/fulfillment-centers", "required": true},
                                {"type": "select", "name": "priority", "label": "Processing Priority", "options": ["standard", "expedited", "rush"], "value": "standard"},
                                {"type": "textarea", "name": "fulfillment_notes", "label": "Special Instructions"}
                            ]
                        },
                        {
                            "title": "Shipping",
                            "body": [
                                {"type": "select", "name": "shipping_method", "label": "Shipping Method", "source": "/api/shipping-methods", "required": true},
                                {"type": "select", "name": "carrier", "label": "Carrier", "source": "/api/carriers", "required": true},
                                {"type": "switch", "name": "require_signature", "label": "Require Signature"},
                                {"type": "switch", "name": "insurance", "label": "Add Insurance"}
                            ]
                        }
                    ]
                }
            ]
        }
    }
}
```

## 💳 Invoice-to-Payment Cycle

### Invoice Generation Workflow
```json
{
    "type": "page",
    "title": "Invoice Management",
    "subTitle": "Generate and track customer invoices",
    
    "body": [
        {
            "type": "crud2",
            "api": "/api/invoices",
            "title": "Customer Invoices",
            
            "headerToolbar": [
                {
                    "type": "dropdown-button",
                    "label": "Create Invoice",
                    "level": "primary",
                    "buttons": [
                        {"type": "button", "label": "From Order", "actionType": "dialog", "dialog": {"$ref": "#/definitions/CreateFromOrderDialog"}},
                        {"type": "button", "label": "Manual Invoice", "actionType": "link", "link": "/invoices/create"},
                        {"type": "button", "label": "Recurring Invoice", "actionType": "dialog"}
                    ]
                }
            ],
            
            "columns": [
                {"name": "invoice_number", "label": "Invoice #", "type": "link", "href": "/invoices/${id}"},
                {"name": "customer_name", "label": "Customer"},
                {"name": "invoice_date", "label": "Date", "type": "date"},
                {"name": "due_date", "label": "Due Date", "type": "date", "classNameExpr": "${DATETODAY(due_date) < 0 ? 'text-red-600 font-bold' : ''}"},
                {"name": "amount", "label": "Amount", "type": "number", "prefix": "$"},
                {"name": "amount_paid", "label": "Paid", "type": "number", "prefix": "$"},
                {"name": "balance_due", "label": "Balance", "type": "number", "prefix": "$", "classNameExpr": "${balance_due > 0 ? 'text-red-600' : 'text-green-600'}"},
                {"name": "status", "label": "Status", "type": "status"},
                {
                    "type": "operation",
                    "buttons": [
                        {
                            "type": "button",
                            "label": "Record Payment",
                            "actionType": "dialog",
                            "level": "success",
                            "visibleOn": "${balance_due > 0}",
                            "dialog": {
                                "title": "Record Payment",
                                "body": {
                                    "type": "form",
                                    "api": "post:/api/invoices/${id}/payments",
                                    "body": [
                                        {"type": "input-number", "name": "amount", "label": "Payment Amount", "prefix": "$", "required": true, "max": "${balance_due}"},
                                        {"type": "input-date", "name": "payment_date", "label": "Payment Date", "value": "${TODAY()}", "required": true},
                                        {"type": "select", "name": "payment_method", "label": "Payment Method", "options": ["check", "wire", "credit_card", "cash"], "required": true},
                                        {"type": "input-text", "name": "reference", "label": "Reference/Check Number"},
                                        {"type": "textarea", "name": "notes", "label": "Notes"}
                                    ]
                                }
                            }
                        },
                        {
                            "type": "dropdown-button",
                            "label": "Actions",
                            "buttons": [
                                {"type": "button", "label": "Send Invoice", "actionType": "ajax", "api": "post:/api/invoices/${id}/send"},
                                {"type": "button", "label": "Download PDF", "actionType": "download", "api": "/api/invoices/${id}/pdf"},
                                {"type": "button", "label": "View Payments", "actionType": "dialog"},
                                {"type": "divider"},
                                {"type": "button", "label": "Void Invoice", "actionType": "ajax", "api": "put:/api/invoices/${id}/void", "confirmText": "Void this invoice?", "level": "danger", "visibleOn": "${status !== 'void'}"}
                            ]
                        }
                    ]
                }
            ]
        }
    ],
    
    "definitions": {
        "CreateFromOrderDialog": {
            "title": "Create Invoice from Order",
            "body": {
                "type": "form",
                "api": "post:/api/invoices/from-order",
                "body": [
                    {"type": "select", "name": "order_id", "label": "Select Order", "source": "/api/orders?status=completed&invoiced=false", "required": true},
                    {"type": "input-date", "name": "invoice_date", "label": "Invoice Date", "value": "${TODAY()}", "required": true},
                    {"type": "input-date", "name": "due_date", "label": "Due Date", "required": true},
                    {"type": "textarea", "name": "invoice_notes", "label": "Invoice Notes"}
                ]
            }
        }
    }
}
```

This guide provides comprehensive patterns for implementing common ERP workflows using the component library, ensuring consistent user experience and proper business process automation.