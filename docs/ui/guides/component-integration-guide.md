# Component Integration Guide

**FILE PURPOSE**: Practical guide for combining UI components to build complete ERP interfaces  
**SCOPE**: Component composition patterns, data flow, API integration, and best practices  
**TARGET AUDIENCE**: Developers building complete ERP applications using the component library

## 📋 Guide Overview

This guide demonstrates how to effectively combine atoms, molecules, organisms, and templates to create complete ERP business applications. It covers composition patterns, data flow management, and real-world implementation examples.

## 🏗️ Component Architecture Layers

### Atomic Design Hierarchy
```
Templates (Pages, Services, Operations)
    ↓
Organisms (CRUD, Forms, Tables, Navigation)
    ↓
Molecules (Cards, Modals, Complex Inputs)
    ↓
Atoms (Buttons, Inputs, Text, Icons)
```

## 🔄 Data Flow Patterns

### 1. Top-Down Data Flow
```json
{
    "type": "page",
    "initApi": "/api/customer/${id}",
    "body": [
        {
            "type": "service",
            "api": "/api/customer/${id}/orders",
            "body": [
                {
                    "type": "crud2",
                    "source": "${orders}",
                    "columns": [
                        {"name": "order_number", "label": "Order #"},
                        {"name": "total", "label": "Total", "type": "number", "prefix": "$"}
                    ]
                }
            ]
        }
    ]
}
```

### 2. Component Communication
```json
{
    "type": "page",
    "body": [
        {
            "type": "form",
            "name": "filter_form",
            "target": "customer_table",
            "body": [
                {"type": "input-text", "name": "search", "placeholder": "Search customers..."},
                {"type": "select", "name": "status", "options": ["active", "inactive"]}
            ]
        },
        {
            "type": "crud2",
            "name": "customer_table",
            "api": "/api/customers",
            "columns": [
                {"name": "company_name", "label": "Company"},
                {"name": "status", "label": "Status", "type": "status"}
            ]
        }
    ]
}
```

## 🎯 Complete Application Examples

### Customer Management Application

#### 1. Main Customer Page Template
```json
{
    "type": "page",
    "title": "Customer Management",
    "subTitle": "Manage your customer relationships",
    "initApi": "/api/customers/page-data",
    
    "aside": [
        {
            "type": "panel",
            "title": "Quick Filters",
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
                            "type": "range",
                            "name": "annual_revenue_range",
                            "label": "Annual Revenue",
                            "min": 0,
                            "max": 10000000,
                            "step": 100000
                        }
                    ]
                }
            ]
        },
        {
            "type": "divider"
        },
        {
            "type": "panel",
            "title": "Customer Stats",
            "body": [
                {
                    "type": "service",
                    "api": "/api/customers/stats",
                    "interval": 30000,
                    "body": [
                        {"type": "stats", "title": "Total Customers", "value": "${total_customers}"},
                        {"type": "stats", "title": "Active", "value": "${active_customers}"},
                        {"type": "stats", "title": "New This Month", "value": "${new_this_month}"}
                    ]
                }
            ]
        }
    ],
    
    "toolbar": [
        {
            "type": "button",
            "label": "Add Customer",
            "actionType": "dialog",
            "icon": "plus",
            "level": "primary",
            "dialog": {
                "$ref": "#/definitions/CustomerFormDialog"
            }
        },
        {
            "type": "dropdown-button",
            "label": "Import/Export",
            "buttons": [
                {
                    "type": "button",
                    "label": "Import CSV",
                    "actionType": "dialog",
                    "dialog": {
                        "title": "Import Customers",
                        "body": {
                            "type": "form",
                            "api": "post:/api/customers/import",
                            "body": [
                                {"type": "input-file", "name": "csv_file", "label": "CSV File", "accept": ".csv", "required": true},
                                {"type": "switch", "name": "update_existing", "label": "Update Existing", "option": "Update customers if they already exist"}
                            ]
                        }
                    }
                },
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
            "title": "Customer Directory",
            
            "filter": {
                "title": "Advanced Search",
                "body": [
                    {"type": "input-text", "name": "company_name", "label": "Company Name"},
                    {"type": "input-text", "name": "contact_person", "label": "Contact Person"},
                    {"type": "input-email", "name": "email", "label": "Email"},
                    {"type": "select", "name": "country", "label": "Country", "source": "/api/countries"}
                ]
            },
            
            "columns": [
                {
                    "name": "company_name",
                    "label": "Company",
                    "searchable": true,
                    "sortable": true,
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
                    "type": "email",
                    "copyable": true
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
                    "name": "annual_revenue",
                    "label": "Annual Revenue",
                    "type": "number",
                    "prefix": "$",
                    "sortable": true
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
                    "width": 200,
                    "buttons": [
                        {
                            "type": "button",
                            "label": "View",
                            "actionType": "link",
                            "link": "/customers/${id}",
                            "level": "link"
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
                                        {"$ref": "#/definitions/CustomerFormFields"}
                                    ]
                                }
                            }
                        },
                        {
                            "type": "dropdown-button",
                            "label": "More",
                            "buttons": [
                                {"type": "button", "label": "Send Email", "actionType": "dialog"},
                                {"type": "button", "label": "Create Order", "actionType": "link", "link": "/orders/create?customer_id=${id}"},
                                {"type": "button", "label": "View Orders", "actionType": "link", "link": "/orders?customer_id=${id}"},
                                {"type": "divider"},
                                {"type": "button", "label": "Delete", "actionType": "ajax", "api": "delete:/api/customers/${id}", "confirmText": "Delete this customer?", "level": "danger"}
                            ]
                        }
                    ]
                }
            ],
            
            "bulkActions": [
                {"type": "button", "label": "Bulk Email", "actionType": "dialog", "level": "primary"},
                {"type": "button", "label": "Export Selected", "actionType": "download", "api": "/api/customers/bulk-export"},
                {"type": "button", "label": "Bulk Delete", "actionType": "ajax", "api": "delete:/api/customers/bulk", "confirmText": "Delete selected customers?", "level": "danger"}
            ],
            
            "perPage": 25,
            "autoFillHeight": true
        }
    ],
    
    "definitions": {
        "CustomerFormDialog": {
            "title": "Add New Customer",
            "size": "lg",
            "body": {
                "type": "form",
                "api": "post:/api/customers",
                "body": [
                    {"$ref": "#/definitions/CustomerFormFields"}
                ]
            }
        },
        
        "CustomerFormFields": [
            {
                "type": "tabs",
                "tabs": [
                    {
                        "title": "Basic Information",
                        "body": [
                            {"type": "input-text", "name": "company_name", "label": "Company Name", "required": true},
                            {"type": "input-text", "name": "contact_person", "label": "Contact Person"},
                            {"type": "input-email", "name": "email", "label": "Email Address", "required": true},
                            {"type": "input-text", "name": "phone", "label": "Phone Number"},
                            {"type": "input-text", "name": "website", "label": "Website"},
                            {"type": "select", "name": "industry", "label": "Industry", "source": "/api/industries"},
                            {"type": "select", "name": "status", "label": "Status", "options": ["active", "inactive", "prospect"]}
                        ]
                    },
                    {
                        "title": "Address",
                        "body": [
                            {"type": "input-text", "name": "street_address", "label": "Street Address"},
                            {"type": "input-text", "name": "city", "label": "City"},
                            {"type": "select", "name": "state", "label": "State/Province", "source": "/api/states"},
                            {"type": "input-text", "name": "postal_code", "label": "Postal Code"},
                            {"type": "select", "name": "country", "label": "Country", "source": "/api/countries"}
                        ]
                    },
                    {
                        "title": "Business Details",
                        "body": [
                            {"type": "input-number", "name": "annual_revenue", "label": "Annual Revenue", "prefix": "$", "min": 0},
                            {"type": "input-number", "name": "employee_count", "label": "Number of Employees", "min": 1},
                            {"type": "input-text", "name": "tax_id", "label": "Tax ID"},
                            {"type": "select", "name": "payment_terms", "label": "Payment Terms", "options": ["Net 30", "Net 15", "Due on Receipt"]},
                            {"type": "textarea", "name": "notes", "label": "Notes"}
                        ]
                    }
                ]
            }
        ]
    }
}
```

#### 2. Customer Detail Page
```json
{
    "type": "page",
    "title": "${customer.company_name}",
    "subTitle": "Customer Details",
    "initApi": "/api/customers/${id}",
    
    "toolbar": [
        {
            "type": "button",
            "label": "Edit Customer",
            "actionType": "dialog",
            "level": "primary",
            "dialog": {
                "title": "Edit Customer",
                "body": {
                    "type": "form",
                    "api": "put:/api/customers/${id}",
                    "initApi": "/api/customers/${id}",
                    "body": [
                        {"$ref": "#/definitions/CustomerFormFields"}
                    ]
                }
            }
        },
        {
            "type": "dropdown-button",
            "label": "Actions",
            "buttons": [
                {"type": "button", "label": "Create Order", "actionType": "link", "link": "/orders/create?customer_id=${id}"},
                {"type": "button", "label": "Send Email", "actionType": "dialog"},
                {"type": "button", "label": "Generate Report", "actionType": "download", "api": "/api/customers/${id}/report.pdf"},
                {"type": "divider"},
                {"type": "button", "label": "Delete Customer", "actionType": "ajax", "api": "delete:/api/customers/${id}", "confirmText": "Delete this customer?", "level": "danger"}
            ]
        }
    ],
    
    "body": [
        {
            "type": "tabs",
            "tabs": [
                {
                    "title": "Overview",
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
                                                    "type": "grid",
                                                    "columns": [
                                                        {
                                                            "md": 6,
                                                            "body": [
                                                                {"type": "static", "name": "company_name", "label": "Company Name"},
                                                                {"type": "static", "name": "contact_person", "label": "Contact Person"},
                                                                {"type": "static", "name": "email", "label": "Email"},
                                                                {"type": "static", "name": "phone", "label": "Phone"},
                                                                {"type": "static", "name": "website", "label": "Website", "type": "link"}
                                                            ]
                                                        },
                                                        {
                                                            "md": 6,
                                                            "body": [
                                                                {"type": "static", "name": "industry", "label": "Industry"},
                                                                {"type": "static", "name": "annual_revenue", "label": "Annual Revenue", "tpl": "$${annual_revenue | number}"},
                                                                {"type": "static", "name": "employee_count", "label": "Employees"},
                                                                {"type": "static", "name": "status", "label": "Status", "type": "status"},
                                                                {"type": "static", "name": "created_date", "label": "Customer Since", "tpl": "${created_date | date:\"MMM DD, YYYY\"}"}
                                                            ]
                                                        }
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
                                                {
                                                    "type": "service",
                                                    "api": "/api/customers/${id}/stats",
                                                    "body": [
                                                        {"type": "stats", "title": "Total Orders", "value": "${total_orders}"},
                                                        {"type": "stats", "title": "Total Spent", "value": "$${total_spent | number}"},
                                                        {"type": "stats", "title": "Average Order", "value": "$${average_order | number}"},
                                                        {"type": "stats", "title": "Last Order", "value": "${last_order_date | fromNow}"}
                                                    ]
                                                }
                                            ]
                                        }
                                    ]
                                }
                            ]
                        }
                    ]
                },
                {
                    "title": "Orders",
                    "body": [
                        {
                            "type": "service",
                            "api": "/api/customers/${id}/orders",
                            "body": [
                                {
                                    "type": "crud2",
                                    "source": "${orders}",
                                    "title": "Customer Orders",
                                    "headerToolbar": [
                                        {
                                            "type": "button",
                                            "label": "Create Order",
                                            "actionType": "link",
                                            "link": "/orders/create?customer_id=${id}",
                                            "level": "primary"
                                        }
                                    ],
                                    "columns": [
                                        {"name": "order_number", "label": "Order #", "type": "link", "href": "/orders/${id}"},
                                        {"name": "order_date", "label": "Date", "type": "date"},
                                        {"name": "total_amount", "label": "Total", "type": "number", "prefix": "$"},
                                        {"name": "status", "label": "Status", "type": "status"}
                                    ]
                                }
                            ]
                        }
                    ]
                },
                {
                    "title": "Invoices",
                    "body": [
                        {
                            "type": "service",
                            "api": "/api/customers/${id}/invoices",
                            "body": [
                                {
                                    "type": "crud2",
                                    "source": "${invoices}",
                                    "title": "Customer Invoices",
                                    "columns": [
                                        {"name": "invoice_number", "label": "Invoice #", "type": "link", "href": "/invoices/${id}"},
                                        {"name": "invoice_date", "label": "Date", "type": "date"},
                                        {"name": "due_date", "label": "Due Date", "type": "date"},
                                        {"name": "amount", "label": "Amount", "type": "number", "prefix": "$"},
                                        {"name": "status", "label": "Status", "type": "status"}
                                    ]
                                }
                            ]
                        }
                    ]
                },
                {
                    "title": "Communication",
                    "body": [
                        {
                            "type": "service",
                            "api": "/api/customers/${id}/communications",
                            "body": [
                                {
                                    "type": "list",
                                    "source": "${communications}",
                                    "listItem": {
                                        "title": "${subject}",
                                        "subTitle": "${communication_date | fromNow}",
                                        "desc": "${message}",
                                        "avatar": {"type": "icon", "icon": "${type === 'email' ? 'mail' : 'phone'}"}
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

### Order Processing Workflow

#### Order Management Dashboard
```json
{
    "type": "page",
    "title": "Order Management",
    "subTitle": "Process and track customer orders",
    "initApi": "/api/orders/dashboard",
    
    "body": [
        {
            "type": "grid",
            "columns": [
                {
                    "md": 3,
                    "body": [
                        {
                            "type": "service",
                            "api": "/api/orders/stats",
                            "interval": 30000,
                            "body": [
                                {"type": "stats", "title": "Pending Orders", "value": "${pending_orders}", "className": "text-orange-600"},
                                {"type": "stats", "title": "Processing", "value": "${processing_orders}", "className": "text-blue-600"},
                                {"type": "stats", "title": "Shipped Today", "value": "${shipped_today}", "className": "text-green-600"},
                                {"type": "stats", "title": "Total Revenue", "value": "$${daily_revenue | number}", "className": "text-purple-600"}
                            ]
                        }
                    ]
                },
                {
                    "md": 9,
                    "body": [
                        {
                            "type": "tabs",
                            "tabs": [
                                {
                                    "title": "All Orders",
                                    "body": [
                                        {
                                            "type": "crud2",
                                            "api": "/api/orders",
                                            "columns": [
                                                {"name": "order_number", "label": "Order #", "type": "link", "href": "/orders/${id}"},
                                                {"name": "customer_name", "label": "Customer"},
                                                {"name": "order_date", "label": "Date", "type": "date"},
                                                {"name": "total_amount", "label": "Total", "type": "number", "prefix": "$"},
                                                {"name": "status", "label": "Status", "type": "status"},
                                                {
                                                    "type": "operation",
                                                    "buttons": [
                                                        {"type": "button", "label": "Process", "actionType": "dialog", "level": "primary", "visibleOn": "${status === 'pending'}"},
                                                        {"type": "button", "label": "View", "actionType": "link", "link": "/orders/${id}"}
                                                    ]
                                                }
                                            ]
                                        }
                                    ]
                                },
                                {
                                    "title": "Urgent Orders",
                                    "badge": {"mode": "text", "text": "${urgent_orders_count}", "className": "bg-red-500"},
                                    "body": [
                                        {
                                            "type": "service",
                                            "api": "/api/orders?urgent=true",
                                            "body": [
                                                {
                                                    "type": "list",
                                                    "source": "${orders}",
                                                    "listItem": {
                                                        "title": "Order #${order_number}",
                                                        "subTitle": "${customer_name} • ${order_date | fromNow}",
                                                        "desc": "Total: $${total_amount} • ${urgent_reason}",
                                                        "actions": [
                                                            {"type": "button", "label": "Process Now", "actionType": "dialog", "level": "danger"}
                                                        ]
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
            ]
        }
    ]
}
```

## 🛠️ Best Practices

### 1. Component Composition
- **Keep templates simple**: Page and Service should orchestrate, not implement
- **Use organisms for business logic**: CRUD, Forms, Tables handle domain-specific functionality  
- **Molecules for reusable patterns**: Cards, modals, complex inputs used across organisms
- **Atoms for consistency**: Buttons, inputs, text maintain design system

### 2. Data Management
- **Initialize at page level**: Use `initApi` on Page templates
- **Service components for data orchestration**: Transform and distribute data to child components
- **Form targeting**: Use `target` attribute to update specific components
- **Real-time updates**: Use WebSocket and polling for live data

### 3. API Integration
- **RESTful conventions**: Use standard HTTP methods and status codes
- **Consistent response format**: Standardize API responses for components
- **Error handling**: Use `showErrorMsg` and custom error messages
- **Loading states**: Configure loading indicators for better UX

### 4. Performance Optimization
- **Lazy loading**: Use `mountOnEnter` for heavy components
- **Pagination**: Implement server-side pagination for large datasets
- **Caching**: Use browser caching and conditional requests
- **Silent polling**: Use `silentPolling` for background updates

### 5. Security Considerations
- **Input validation**: Validate all form inputs server-side
- **Authorization**: Check permissions before rendering sensitive components
- **XSS prevention**: Sanitize user-generated content
- **CSRF protection**: Include CSRF tokens in form submissions

This guide provides a foundation for building complete ERP applications using the component library with proper architecture, data flow, and integration patterns.