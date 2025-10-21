# Dialog Component

**FILE PURPOSE**: Modal dialog organism for overlays, forms, confirmations, and content display  
**SCOPE**: Modal windows, pop-ups, forms, confirmations, and overlay content  
**TARGET AUDIENCE**: Developers implementing modal dialogs, forms, confirmations, and overlay interfaces

## 📋 Component Overview

Dialog provides comprehensive modal functionality with support for various sizes, custom content, form integration, confirmation dialogs, and advanced features like dragging and custom actions. Essential for creating focused user interactions in ERP systems.

### Schema Reference
- **Primary Schema**: `DialogSchema.json`
- **Related Schemas**: `DialogSchemaBase.json`, `ActionSchema.json`
- **Base Interface**: Modal dialog organism for overlay content

## Basic Usage

```json
{
    "type": "dialog",
    "title": "Edit Customer",
    "body": {
        "type": "form",
        "body": [
            {"type": "input-text", "name": "name", "label": "Name"},
            {"type": "input-email", "name": "email", "label": "Email"}
        ]
    }
}
```

## Go Type Definition

```go
type DialogProps struct {
    Type                    string              `json:"type"`
    Title                   interface{}         `json:"title"`               // Dialog title
    Body                    interface{}         `json:"body"`                // Dialog content
    Actions                 []interface{}       `json:"actions"`             // Custom actions
    
    // Size and Dimensions
    Size                    string              `json:"size"`                // "xs", "sm", "md", "lg", "xl", "full"
    Width                   string              `json:"width"`               // Custom width
    Height                  string              `json:"height"`              // Custom height
    
    // Behavior
    CloseOnEsc              bool                `json:"closeOnEsc"`          // ESC key closes
    CloseOnOutside          bool                `json:"closeOnOutside"`      // Click outside closes
    Draggable               bool                `json:"draggable"`           // Draggable dialog
    Overlay                 bool                `json:"overlay"`             // Show backdrop
    
    // Display Options
    ShowCloseButton         bool                `json:"showCloseButton"`     // Close button visibility
    ShowErrorMsg            bool                `json:"showErrorMsg"`        // Error message display
    ShowLoading             bool                `json:"showLoading"`         // Loading spinner
    Confirm                 bool                `json:"confirm"`             // Auto-generate confirm buttons
    
    // Dialog Type
    DialogType              string              `json:"dialogType"`          // "confirm" for confirmation
    
    // Header and Footer
    Header                  interface{}         `json:"header"`              // Custom header
    HeaderClassName         string              `json:"headerClassName"`     // Header CSS
    Footer                  interface{}         `json:"footer"`              // Custom footer
    
    // Content Styling
    BodyClassName           string              `json:"bodyClassName"`       // Body CSS
    
    // Data and Parameters
    Data                    interface{}         `json:"data"`                // Data mapping
    InputParams             interface{}         `json:"inputParams"`         // Input parameters
    Name                    string              `json:"name"`                // Dialog name
}
```

## Dialog Sizes

### Standard Sizes
```json
{
    "type": "dialog",
    "size": "lg",
    "title": "Large Dialog",
    "body": {"type": "text", "text": "This is a large dialog"}
}
```

### Custom Dimensions
```json
{
    "type": "dialog",
    "width": "800px",
    "height": "600px",
    "title": "Custom Size Dialog",
    "body": {"type": "text", "text": "Custom sized dialog"}
}
```

### Full Screen Dialog
```json
{
    "type": "dialog",
    "size": "full",
    "title": "Full Screen Dialog",
    "body": {"type": "text", "text": "Full screen dialog"}
}
```

## Real-World Use Cases

### Customer Edit Dialog
```json
{
    "type": "dialog",
    "title": "Edit Customer Information",
    "size": "lg",
    "closeOnEsc": true,
    "closeOnOutside": false,
    "showCloseButton": true,
    "body": {
        "type": "form",
        "api": "put:/api/customers/${id}",
        "initApi": "/api/customers/${id}",
        "mode": "horizontal",
        "horizontal": {"left": 3, "right": 9},
        "body": [
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
                            {"type": "select", "name": "industry", "label": "Industry", "source": "/api/industries"},
                            {"type": "select", "name": "status", "label": "Status", "options": ["active", "inactive", "prospect"]}
                        ]
                    },
                    {
                        "title": "Address",
                        "body": [
                            {"type": "input-text", "name": "street", "label": "Street Address"},
                            {"type": "input-text", "name": "suite", "label": "Suite/Unit"},
                            {"type": "group", "direction": "horizontal", "body": [
                                {"type": "input-city", "name": "city", "label": "City"},
                                {"type": "select", "name": "state", "label": "State", "source": "/api/states"}
                            ]},
                            {"type": "group", "direction": "horizontal", "body": [
                                {"type": "input-text", "name": "zip", "label": "ZIP Code"},
                                {"type": "select", "name": "country", "label": "Country", "value": "US", "source": "/api/countries"}
                            ]}
                        ]
                    },
                    {
                        "title": "Notes",
                        "body": [
                            {"type": "textarea", "name": "notes", "label": "Customer Notes", "rows": 6}
                        ]
                    }
                ]
            }
        ]
    },
    "actions": [
        {
            "type": "button",
            "label": "Cancel",
            "actionType": "cancel",
            "level": "default"
        },
        {
            "type": "submit",
            "label": "Save Changes",
            "level": "primary"
        }
    ]
}
```

### Product Details Dialog
```json
{
    "type": "dialog",
    "title": "Product Details",
    "size": "xl",
    "draggable": true,
    "closeOnEsc": true,
    "body": {
        "type": "service",
        "api": "/api/products/${id}",
        "body": [
            {
                "type": "grid",
                "columns": [
                    {
                        "md": 4,
                        "body": [
                            {"type": "image", "src": "${main_image}", "enlargeAble": true, "className": "w-full"},
                            {
                                "type": "images",
                                "name": "gallery",
                                "thumbMode": "cover",
                                "thumbRatio": "1:1"
                            }
                        ]
                    },
                    {
                        "md": 8,
                        "body": [
                            {"type": "static", "name": "name", "label": "Product Name", "className": "text-2xl font-bold"},
                            {"type": "static", "name": "sku", "label": "SKU"},
                            {"type": "static", "name": "price", "label": "Price", "tpl": "$${price}", "className": "text-xl text-green-600 font-bold"},
                            {"type": "static", "name": "category", "label": "Category"},
                            {"type": "static", "name": "stock_quantity", "label": "In Stock"},
                            {"type": "divider"},
                            {"type": "static", "name": "description", "label": "Description"},
                            {"type": "static", "name": "specifications", "label": "Specifications", "tpl": "${specifications | raw}"},
                            {"type": "divider"},
                            {
                                "type": "flex",
                                "items": [
                                    {
                                        "type": "button",
                                        "label": "Edit Product",
                                        "actionType": "dialog",
                                        "level": "primary",
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
                                        "type": "button",
                                        "label": "Delete Product",
                                        "actionType": "ajax",
                                        "api": "delete:/api/products/${id}",
                                        "confirmText": "Are you sure you want to delete this product?",
                                        "level": "danger"
                                    }
                                ]
                            }
                        ]
                    }
                ]
            }
        ]
    },
    "actions": [
        {
            "type": "button",
            "label": "Close",
            "actionType": "close",
            "level": "default"
        }
    ]
}
```

### Confirmation Dialog
```json
{
    "type": "dialog",
    "dialogType": "confirm",
    "title": "Delete Customer",
    "size": "sm",
    "closeOnEsc": true,
    "body": [
        {
            "type": "alert",
            "level": "warning",
            "body": "Are you sure you want to delete customer <strong>${company_name}</strong>?"
        },
        {
            "type": "text",
            "text": "This action cannot be undone. All associated orders, invoices, and historical data will be permanently removed."
        }
    ],
    "actions": [
        {
            "type": "button",
            "label": "Cancel",
            "actionType": "cancel",
            "level": "default"
        },
        {
            "type": "button",
            "label": "Delete Customer",
            "actionType": "ajax",
            "api": "delete:/api/customers/${id}",
            "level": "danger",
            "confirmText": "Please confirm deletion by typing DELETE",
            "confirmInput": true
        }
    ]
}
```

### Order Processing Dialog
```json
{
    "type": "dialog",
    "title": "Process Order #${order_number}",
    "size": "lg",
    "closeOnEsc": false,
    "closeOnOutside": false,
    "showCloseButton": false,
    "body": {
        "type": "form",
        "api": "put:/api/orders/${id}/process",
        "body": [
            {
                "type": "panel",
                "title": "Order Summary",
                "body": [
                    {
                        "type": "grid",
                        "columns": [
                            {
                                "md": 6,
                                "body": [
                                    {"type": "static", "name": "order_number", "label": "Order Number"},
                                    {"type": "static", "name": "customer_name", "label": "Customer"},
                                    {"type": "static", "name": "order_date", "label": "Order Date", "tpl": "${order_date | date:\"MMM DD, YYYY\"}"}
                                ]
                            },
                            {
                                "md": 6,
                                "body": [
                                    {"type": "static", "name": "total_amount", "label": "Total Amount", "tpl": "$${total_amount}"},
                                    {"type": "static", "name": "payment_status", "label": "Payment Status"},
                                    {"type": "static", "name": "shipping_method", "label": "Shipping Method"}
                                ]
                            }
                        ]
                    }
                ]
            },
            {
                "type": "panel",
                "title": "Order Items",
                "body": [
                    {
                        "type": "table",
                        "source": "${items}",
                        "columns": [
                            {"name": "product_name", "label": "Product"},
                            {"name": "quantity", "label": "Quantity"},
                            {"name": "unit_price", "label": "Unit Price", "type": "number", "prefix": "$"},
                            {"name": "total_price", "label": "Total", "type": "number", "prefix": "$"}
                        ]
                    }
                ]
            },
            {
                "type": "panel",
                "title": "Processing Options",
                "body": [
                    {
                        "type": "select",
                        "name": "fulfillment_center",
                        "label": "Fulfillment Center",
                        "source": "/api/fulfillment-centers",
                        "required": true
                    },
                    {
                        "type": "select",
                        "name": "shipping_carrier",
                        "label": "Shipping Carrier",
                        "source": "/api/shipping-carriers",
                        "required": true
                    },
                    {
                        "type": "select",
                        "name": "priority",
                        "label": "Processing Priority",
                        "options": ["standard", "expedited", "rush"],
                        "value": "standard"
                    },
                    {
                        "type": "textarea",
                        "name": "processing_notes",
                        "label": "Processing Notes",
                        "placeholder": "Any special instructions for fulfillment..."
                    },
                    {
                        "type": "switch",
                        "name": "send_tracking_email",
                        "label": "Send Tracking Email",
                        "option": "Automatically send tracking information to customer",
                        "value": true
                    }
                ]
            }
        ]
    },
    "actions": [
        {
            "type": "button",
            "label": "Cancel Processing",
            "actionType": "cancel",
            "level": "default"
        },
        {
            "type": "submit",
            "label": "Process Order",
            "level": "primary"
        }
    ]
}
```

### File Upload Dialog
```json
{
    "type": "dialog",
    "title": "Upload Documents",
    "size": "md",
    "closeOnEsc": true,
    "body": {
        "type": "form",
        "api": "post:/api/documents/upload",
        "body": [
            {
                "type": "input-file",
                "name": "documents",
                "label": "Select Files",
                "multiple": true,
                "accept": ".pdf,.doc,.docx,.jpg,.png,.xls,.xlsx",
                "maxLength": 10,
                "maxSize": 10485760,
                "autoUpload": false,
                "description": "Supported formats: PDF, DOC, DOCX, JPG, PNG, XLS, XLSX. Max file size: 10MB"
            },
            {
                "type": "select",
                "name": "document_type",
                "label": "Document Type",
                "options": [
                    "Invoice",
                    "Receipt", 
                    "Contract",
                    "Report",
                    "Specification",
                    "Other"
                ],
                "required": true
            },
            {
                "type": "tags",
                "name": "tags",
                "label": "Tags",
                "placeholder": "Add tags for better organization"
            },
            {
                "type": "textarea",
                "name": "description",
                "label": "Description",
                "placeholder": "Brief description of the documents..."
            },
            {
                "type": "switch",
                "name": "is_confidential",
                "label": "Confidential",
                "option": "Mark as confidential document"
            }
        ]
    },
    "actions": [
        {
            "type": "button",
            "label": "Cancel",
            "actionType": "cancel",
            "level": "default"
        },
        {
            "type": "submit",
            "label": "Upload Documents",
            "level": "primary"
        }
    ]
}
```

### Report Generation Dialog
```json
{
    "type": "dialog",
    "title": "Generate Sales Report",
    "size": "lg",
    "draggable": true,
    "body": {
        "type": "form",
        "api": "post:/api/reports/generate",
        "mode": "horizontal",
        "horizontal": {"left": 3, "right": 9},
        "body": [
            {
                "type": "fieldset",
                "title": "Report Parameters",
                "body": [
                    {
                        "type": "select",
                        "name": "report_type",
                        "label": "Report Type",
                        "options": [
                            {"label": "Sales Summary", "value": "sales_summary"},
                            {"label": "Product Performance", "value": "product_performance"},
                            {"label": "Customer Analysis", "value": "customer_analysis"},
                            {"label": "Regional Sales", "value": "regional_sales"}
                        ],
                        "required": true,
                        "value": "sales_summary"
                    },
                    {
                        "type": "input-date-range",
                        "name": "date_range",
                        "label": "Date Range",
                        "required": true,
                        "value": "last_month"
                    },
                    {
                        "type": "checkboxes",
                        "name": "data_points",
                        "label": "Include Data Points",
                        "options": [
                            {"label": "Revenue", "value": "revenue"},
                            {"label": "Units Sold", "value": "units"},
                            {"label": "Customer Count", "value": "customers"},
                            {"label": "Average Order Value", "value": "aov"},
                            {"label": "Growth Metrics", "value": "growth"}
                        ],
                        "value": ["revenue", "units", "customers"]
                    },
                    {
                        "type": "select",
                        "name": "grouping",
                        "label": "Group By",
                        "options": ["daily", "weekly", "monthly"],
                        "value": "monthly"
                    }
                ]
            },
            {
                "type": "fieldset",
                "title": "Filters",
                "body": [
                    {
                        "type": "select",
                        "name": "sales_rep",
                        "label": "Sales Representative",
                        "source": "/api/sales-reps",
                        "clearable": true,
                        "placeholder": "All Sales Reps"
                    },
                    {
                        "type": "select",
                        "name": "product_categories",
                        "label": "Product Categories",
                        "source": "/api/categories",
                        "multiple": true,
                        "clearable": true
                    },
                    {
                        "type": "select",
                        "name": "customer_segments",
                        "label": "Customer Segments",
                        "options": ["enterprise", "mid-market", "small-business"],
                        "multiple": true,
                        "clearable": true
                    }
                ]
            },
            {
                "type": "fieldset",
                "title": "Output Options",
                "body": [
                    {
                        "type": "radios",
                        "name": "output_format",
                        "label": "Format",
                        "options": [
                            {"label": "PDF Report", "value": "pdf"},
                            {"label": "Excel Spreadsheet", "value": "excel"},
                            {"label": "CSV Data", "value": "csv"}
                        ],
                        "value": "pdf"
                    },
                    {
                        "type": "switch",
                        "name": "include_charts",
                        "label": "Include Charts",
                        "option": "Add visual charts and graphs",
                        "value": true,
                        "visibleOn": "${output_format === 'pdf'}"
                    },
                    {
                        "type": "switch",
                        "name": "email_delivery",
                        "label": "Email Delivery",
                        "option": "Send report via email when ready"
                    },
                    {
                        "type": "input-email",
                        "name": "email_recipients",
                        "label": "Email Recipients",
                        "multiple": true,
                        "visibleOn": "${email_delivery}",
                        "placeholder": "Enter email addresses..."
                    }
                ]
            }
        ]
    },
    "actions": [
        {
            "type": "button",
            "label": "Cancel",
            "actionType": "cancel",
            "level": "default"
        },
        {
            "type": "button",
            "label": "Preview",
            "actionType": "ajax",
            "api": "post:/api/reports/preview",
            "level": "secondary"
        },
        {
            "type": "submit",
            "label": "Generate Report",
            "level": "primary"
        }
    ]
}
```

### Advanced Search Dialog
```json
{
    "type": "dialog",
    "title": "Advanced Search",
    "size": "xl",
    "closeOnEsc": true,
    "body": {
        "type": "form",
        "target": "search_results",
        "api": "get:/api/search/advanced",
        "body": [
            {
                "type": "tabs",
                "tabs": [
                    {
                        "title": "General",
                        "body": [
                            {"type": "input-text", "name": "keyword", "label": "Keywords", "placeholder": "Search terms..."},
                            {"type": "select", "name": "search_in", "label": "Search In", "options": ["all", "customers", "products", "orders", "invoices"], "value": "all"},
                            {"type": "input-date-range", "name": "date_range", "label": "Date Range"}
                        ]
                    },
                    {
                        "title": "Customers",
                        "body": [
                            {"type": "input-text", "name": "company_name", "label": "Company Name"},
                            {"type": "input-text", "name": "contact_person", "label": "Contact Person"},
                            {"type": "select", "name": "customer_status", "label": "Status", "options": ["active", "inactive", "prospect"]},
                            {"type": "select", "name": "industry", "label": "Industry", "source": "/api/industries"},
                            {"type": "select", "name": "country", "label": "Country", "source": "/api/countries"}
                        ]
                    },
                    {
                        "title": "Products",
                        "body": [
                            {"type": "input-text", "name": "product_name", "label": "Product Name"},
                            {"type": "input-text", "name": "sku", "label": "SKU"},
                            {"type": "select", "name": "category", "label": "Category", "source": "/api/categories"},
                            {"type": "range", "name": "price_range", "label": "Price Range", "min": 0, "max": 10000},
                            {"type": "select", "name": "availability", "label": "Availability", "options": ["in_stock", "low_stock", "out_of_stock"]}
                        ]
                    },
                    {
                        "title": "Orders",
                        "body": [
                            {"type": "input-text", "name": "order_number", "label": "Order Number"},
                            {"type": "select", "name": "order_status", "label": "Status", "options": ["pending", "processing", "shipped", "delivered", "cancelled"]},
                            {"type": "range", "name": "order_total_range", "label": "Order Total", "min": 0, "max": 100000},
                            {"type": "select", "name": "payment_status", "label": "Payment Status", "options": ["pending", "paid", "refunded"]},
                            {"type": "select", "name": "shipping_method", "label": "Shipping Method", "source": "/api/shipping-methods"}
                        ]
                    }
                ]
            }
        ]
    },
    "actions": [
        {
            "type": "button",
            "label": "Clear All",
            "actionType": "clear",
            "level": "default"
        },
        {
            "type": "button",
            "label": "Cancel",
            "actionType": "cancel",
            "level": "default"
        },
        {
            "type": "submit",
            "label": "Search",
            "level": "primary"
        }
    ]
}
```

This component provides essential dialog functionality for ERP systems requiring modal interfaces, forms, confirmations, and overlay content with comprehensive customization options and interactive features.