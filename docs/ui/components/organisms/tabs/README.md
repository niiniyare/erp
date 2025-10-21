# Tabs Component

**FILE PURPOSE**: Tab-based navigation organism for organizing content into switchable panels  
**SCOPE**: Content organization, multi-view interfaces, settings panels, and information hierarchy  
**TARGET AUDIENCE**: Developers implementing tabbed interfaces, content organization, and multi-panel layouts

## 📋 Component Overview

Tabs provides comprehensive tab functionality with support for dynamic tabs, lazy loading, drag and drop, closable tabs, and multiple display modes. Essential for organizing complex content and providing intuitive navigation in ERP systems.

### Schema Reference
- **Primary Schema**: `TabsSchema.json`
- **Related Schemas**: `TabSchema.json`, `TabsMode.json`, `ActionSchema.json`
- **Base Interface**: Tab container organism for content organization

## Basic Usage

```json
{
    "type": "tabs",
    "tabs": [
        {"title": "General", "body": [{"type": "text", "text": "General content"}]},
        {"title": "Settings", "body": [{"type": "text", "text": "Settings content"}]},
        {"title": "Advanced", "body": [{"type": "text", "text": "Advanced content"}]}
    ]
}
```

## Go Type Definition

```go
type TabsProps struct {
    Type                    string              `json:"type"`
    Tabs                    []interface{}       `json:"tabs"`                // Tab definitions
    
    // Data Source
    Source                  string              `json:"source"`              // Dynamic tabs source
    
    // Display Mode
    TabsMode                string              `json:"tabsMode"`            // Display style
    ContentClassName        string              `json:"contentClassName"`    // Content area CSS
    LinksClassName          string              `json:"linksClassName"`      // Tab links CSS
    
    // Behavior
    MountOnEnter            bool                `json:"mountOnEnter"`        // Lazy loading
    UnmountOnExit           bool                `json:"unmountOnExit"`       // Destroy on hide
    DefaultKey              interface{}         `json:"defaultKey"`          // Default active tab
    ActiveKey               interface{}         `json:"activeKey"`           // Current active tab
    
    // Interactive Features
    Addable                 bool                `json:"addable"`             // Allow adding tabs
    Closable                bool                `json:"closable"`            // Allow closing tabs
    Draggable               bool                `json:"draggable"`           // Drag and drop
    Editable                bool                `json:"editable"`            // Edit tab names
    Swipeable               bool                `json:"swipeable"`           // Touch swipe navigation
    
    // Toolbar and Actions
    Toolbar                 interface{}         `json:"toolbar"`             // Additional toolbar
    
    // Sub-form Configuration
    SubFormMode             string              `json:"subFormMode"`         // Sub-form display mode
    SubFormHorizontal       interface{}         `json:"subFormHorizontal"`   // Horizontal layout
    
    // Overflow Handling
    CollapseOnExceed        int                 `json:"collapseOnExceed"`    // Collapse threshold
    CollapseBtnLabel        string              `json:"collapseBtnLabel"`    // Collapse button text
    Scrollable              bool                `json:"scrollable"`          // Scroll overflow
    
    // Customization
    ShowTip                 bool                `json:"showTip"`             // Show tooltips
    ShowTipClassName        string              `json:"showTipClassName"`    // Tooltip CSS
    AddBtnText              string              `json:"addBtnText"`          // Add button text
    SidePosition            string              `json:"sidePosition"`        // Side position for editor
}
```

## Tab Display Modes

### Default Horizontal Tabs
```json
{
    "type": "tabs",
    "tabsMode": "line",
    "tabs": [
        {
            "title": "Customer Information",
            "body": [
                {"type": "input-text", "name": "company_name", "label": "Company Name"},
                {"type": "input-email", "name": "email", "label": "Email"}
            ]
        },
        {
            "title": "Address Details",
            "body": [
                {"type": "input-text", "name": "street", "label": "Street Address"},
                {"type": "input-text", "name": "city", "label": "City"}
            ]
        }
    ]
}
```

### Card-Style Tabs
```json
{
    "type": "tabs",
    "tabsMode": "card",
    "contentClassName": "border border-gray-200 rounded-b-lg p-4",
    "tabs": [
        {"title": "Overview", "body": [{"type": "text", "text": "Overview content"}]},
        {"title": "Details", "body": [{"type": "text", "text": "Details content"}]},
        {"title": "History", "body": [{"type": "text", "text": "History content"}]}
    ]
}
```

### Vertical Sidebar Tabs
```json
{
    "type": "tabs",
    "tabsMode": "sidebar",
    "sidePosition": "left",
    "tabs": [
        {"title": "Profile", "icon": "user", "body": [{"type": "text", "text": "Profile content"}]},
        {"title": "Security", "icon": "shield", "body": [{"type": "text", "text": "Security content"}]},
        {"title": "Notifications", "icon": "bell", "body": [{"type": "text", "text": "Notifications content"}]}
    ]
}
```

## Real-World Use Cases

### Customer Management Tabs
```json
{
    "type": "tabs",
    "tabsMode": "line",
    "mountOnEnter": true,
    "unmountOnExit": false,
    "defaultKey": 0,
    "tabs": [
        {
            "title": "General Information",
            "icon": "info",
            "body": [
                {
                    "type": "form",
                    "mode": "horizontal",
                    "api": "put:/api/customers/${id}",
                    "initApi": "/api/customers/${id}",
                    "body": [
                        {"type": "input-text", "name": "company_name", "label": "Company Name", "required": true},
                        {"type": "input-text", "name": "contact_person", "label": "Contact Person"},
                        {"type": "input-email", "name": "email", "label": "Email Address", "required": true},
                        {"type": "input-text", "name": "phone", "label": "Phone Number"},
                        {"type": "input-text", "name": "website", "label": "Website"},
                        {"type": "select", "name": "industry", "label": "Industry", "source": "/api/industries"},
                        {"type": "select", "name": "status", "label": "Status", "options": ["active", "inactive", "prospect"]}
                    ]
                }
            ]
        },
        {
            "title": "Address & Location",
            "icon": "map-pin",
            "body": [
                {
                    "type": "form",
                    "mode": "horizontal",
                    "api": "put:/api/customers/${id}/address",
                    "initApi": "/api/customers/${id}/address",
                    "body": [
                        {"type": "input-text", "name": "street_address", "label": "Street Address", "required": true},
                        {"type": "input-text", "name": "suite", "label": "Suite/Unit"},
                        {"type": "group", "direction": "horizontal", "body": [
                            {"type": "input-city", "name": "city", "label": "City", "required": true},
                            {"type": "select", "name": "state", "label": "State", "source": "/api/states", "required": true}
                        ]},
                        {"type": "group", "direction": "horizontal", "body": [
                            {"type": "input-text", "name": "zip_code", "label": "ZIP Code", "required": true},
                            {"type": "select", "name": "country", "label": "Country", "value": "US", "source": "/api/countries"}
                        ]},
                        {"type": "textarea", "name": "address_notes", "label": "Address Notes"}
                    ]
                }
            ]
        },
        {
            "title": "Financial Information",
            "icon": "dollar-sign",
            "body": [
                {
                    "type": "form",
                    "mode": "horizontal",
                    "api": "put:/api/customers/${id}/financial",
                    "initApi": "/api/customers/${id}/financial",
                    "body": [
                        {"type": "input-text", "name": "tax_id", "label": "Tax ID Number"},
                        {"type": "select", "name": "payment_terms", "label": "Payment Terms", "options": ["Net 30", "Net 15", "Due on Receipt", "Net 60"]},
                        {"type": "input-number", "name": "credit_limit", "label": "Credit Limit", "prefix": "$", "min": 0},
                        {"type": "select", "name": "currency", "label": "Preferred Currency", "source": "/api/currencies", "value": "USD"},
                        {"type": "select", "name": "payment_method", "label": "Preferred Payment Method", "options": ["Check", "Wire Transfer", "Credit Card", "ACH"]},
                        {"type": "textarea", "name": "billing_notes", "label": "Billing Notes"}
                    ]
                }
            ]
        },
        {
            "title": "Orders & History",
            "icon": "shopping-cart",
            "body": [
                {
                    "type": "crud2",
                    "mode": "table2",
                    "api": "/api/customers/${id}/orders",
                    "title": "Order History",
                    "columns": [
                        {"name": "order_number", "label": "Order #", "type": "link", "href": "/orders/${id}"},
                        {"name": "order_date", "label": "Date", "type": "date", "format": "MMM DD, YYYY"},
                        {"name": "total_amount", "label": "Total", "type": "number", "prefix": "$"},
                        {"name": "status", "label": "Status", "type": "status"}
                    ],
                    "perPage": 10,
                    "autoFillHeight": true
                }
            ]
        },
        {
            "title": "Notes & Documents",
            "icon": "file-text",
            "body": [
                {
                    "type": "form",
                    "api": "put:/api/customers/${id}/notes",
                    "initApi": "/api/customers/${id}/notes",
                    "body": [
                        {"type": "textarea", "name": "internal_notes", "label": "Internal Notes", "rows": 6, "placeholder": "Internal notes about this customer..."},
                        {"type": "textarea", "name": "customer_notes", "label": "Customer Visible Notes", "rows": 4, "placeholder": "Notes visible to the customer..."},
                        {"type": "input-file", "name": "documents", "label": "Customer Documents", "multiple": true, "accept": ".pdf,.doc,.docx,.jpg,.png"}
                    ]
                }
            ]
        }
    ]
}
```

### Product Management with Dynamic Tabs
```json
{
    "type": "tabs",
    "tabsMode": "card",
    "addable": true,
    "closable": true,
    "draggable": true,
    "editable": true,
    "mountOnEnter": true,
    "addBtnText": "Add Category",
    "collapseOnExceed": 6,
    "collapseBtnLabel": "More Categories",
    "source": "${product_categories}",
    "tabs": [
        {
            "title": "${category_name}",
            "hash": "${category_id}",
            "body": [
                {
                    "type": "crud2",
                    "mode": "cards",
                    "api": "/api/products?category_id=${category_id}",
                    "title": "${category_name} Products",
                    "headerToolbar": [
                        {
                            "type": "button",
                            "label": "Add Product",
                            "actionType": "dialog",
                            "level": "primary",
                            "dialog": {
                                "title": "Add Product to ${category_name}",
                                "body": {
                                    "type": "form",
                                    "api": "post:/api/products",
                                    "body": [
                                        {"type": "hidden", "name": "category_id", "value": "${category_id}"},
                                        {"type": "input-text", "name": "name", "label": "Product Name", "required": true},
                                        {"type": "input-text", "name": "sku", "label": "SKU", "required": true},
                                        {"type": "input-number", "name": "price", "label": "Price", "prefix": "$"},
                                        {"type": "textarea", "name": "description", "label": "Description"}
                                    ]
                                }
                            }
                        }
                    ],
                    "card": {
                        "header": {
                            "title": "${name}",
                            "subTitle": "SKU: ${sku}",
                            "avatar": {"type": "image", "src": "${image_url}"}
                        },
                        "body": [
                            {"type": "text", "text": "${description}"},
                            {"type": "text", "text": "Price: $${price}", "className": "font-bold"}
                        ],
                        "actions": [
                            {"type": "button", "label": "Edit", "actionType": "dialog", "level": "primary"},
                            {"type": "button", "label": "Delete", "actionType": "ajax", "level": "danger"}
                        ]
                    },
                    "perPage": 12
                }
            ]
        }
    ]
}
```

### System Settings with Sidebar Tabs
```json
{
    "type": "tabs",
    "tabsMode": "sidebar",
    "sidePosition": "left",
    "contentClassName": "min-h-96",
    "linksClassName": "w-64",
    "showTip": true,
    "tabs": [
        {
            "title": "General Settings",
            "icon": "settings",
            "tip": "Basic system configuration",
            "body": [
                {
                    "type": "form",
                    "title": "Company Information",
                    "api": "put:/api/settings/general",
                    "initApi": "/api/settings/general",
                    "mode": "horizontal",
                    "body": [
                        {"type": "input-text", "name": "company_name", "label": "Company Name", "required": true},
                        {"type": "input-text", "name": "company_address", "label": "Company Address"},
                        {"type": "input-text", "name": "company_phone", "label": "Phone Number"},
                        {"type": "input-email", "name": "company_email", "label": "Email Address"},
                        {"type": "input-text", "name": "website", "label": "Website"},
                        {"type": "select", "name": "timezone", "label": "Timezone", "source": "/api/timezones"},
                        {"type": "select", "name": "date_format", "label": "Date Format", "options": ["MM/DD/YYYY", "DD/MM/YYYY", "YYYY-MM-DD"]},
                        {"type": "select", "name": "currency", "label": "Default Currency", "source": "/api/currencies"}
                    ]
                }
            ]
        },
        {
            "title": "User Management",
            "icon": "users",
            "tip": "Manage users and permissions",
            "body": [
                {
                    "type": "crud2",
                    "mode": "table2",
                    "api": "/api/users",
                    "title": "System Users",
                    "headerToolbar": [
                        {
                            "type": "button",
                            "label": "Add User",
                            "actionType": "dialog",
                            "level": "primary",
                            "dialog": {
                                "title": "Add New User",
                                "body": {
                                    "type": "form",
                                    "api": "post:/api/users",
                                    "body": [
                                        {"type": "input-text", "name": "username", "label": "Username", "required": true},
                                        {"type": "input-email", "name": "email", "label": "Email", "required": true},
                                        {"type": "input-text", "name": "first_name", "label": "First Name", "required": true},
                                        {"type": "input-text", "name": "last_name", "label": "Last Name", "required": true},
                                        {"type": "select", "name": "role", "label": "Role", "source": "/api/roles", "required": true},
                                        {"type": "switch", "name": "is_active", "label": "Active User", "value": true}
                                    ]
                                }
                            }
                        }
                    ],
                    "columns": [
                        {"name": "username", "label": "Username"},
                        {"name": "email", "label": "Email"},
                        {"name": "full_name", "label": "Name"},
                        {"name": "role", "label": "Role"},
                        {"name": "is_active", "label": "Status", "type": "status"},
                        {"type": "operation", "label": "Actions", "buttons": [
                            {"type": "button", "label": "Edit", "actionType": "dialog", "level": "primary"},
                            {"type": "button", "label": "Delete", "actionType": "ajax", "level": "danger"}
                        ]}
                    ]
                }
            ]
        },
        {
            "title": "Security Settings",
            "icon": "shield",
            "tip": "Configure security options",
            "body": [
                {
                    "type": "form",
                    "title": "Security Configuration",
                    "api": "put:/api/settings/security",
                    "initApi": "/api/settings/security",
                    "mode": "horizontal",
                    "body": [
                        {"type": "switch", "name": "require_2fa", "label": "Require Two-Factor Authentication", "option": "Require all users to enable 2FA"},
                        {"type": "input-number", "name": "session_timeout", "label": "Session Timeout (minutes)", "min": 15, "max": 1440, "value": 60},
                        {"type": "input-number", "name": "password_min_length", "label": "Minimum Password Length", "min": 6, "max": 32, "value": 8},
                        {"type": "switch", "name": "password_require_uppercase", "label": "Require Uppercase Letters"},
                        {"type": "switch", "name": "password_require_numbers", "label": "Require Numbers"},
                        {"type": "switch", "name": "password_require_symbols", "label": "Require Special Characters"},
                        {"type": "input-number", "name": "login_attempts", "label": "Max Login Attempts", "min": 3, "max": 10, "value": 5},
                        {"type": "input-number", "name": "lockout_duration", "label": "Account Lockout Duration (minutes)", "min": 5, "max": 60, "value": 15}
                    ]
                }
            ]
        },
        {
            "title": "Email Settings",
            "icon": "mail",
            "tip": "Configure email notifications",
            "body": [
                {
                    "type": "form",
                    "title": "Email Configuration",
                    "api": "put:/api/settings/email",
                    "initApi": "/api/settings/email",
                    "mode": "horizontal",
                    "body": [
                        {"type": "input-text", "name": "smtp_host", "label": "SMTP Host", "required": true},
                        {"type": "input-number", "name": "smtp_port", "label": "SMTP Port", "value": 587},
                        {"type": "input-email", "name": "from_email", "label": "From Email Address", "required": true},
                        {"type": "input-text", "name": "from_name", "label": "From Name"},
                        {"type": "input-text", "name": "smtp_username", "label": "SMTP Username"},
                        {"type": "input-password", "name": "smtp_password", "label": "SMTP Password"},
                        {"type": "switch", "name": "smtp_tls", "label": "Enable TLS/SSL", "value": true},
                        {"type": "button", "label": "Test Email Configuration", "actionType": "ajax", "api": "post:/api/settings/email/test", "level": "info"}
                    ]
                }
            ]
        },
        {
            "title": "Backup & Restore",
            "icon": "database",
            "tip": "Manage system backups",
            "body": [
                {
                    "type": "page",
                    "body": [
                        {
                            "type": "panel",
                            "title": "System Backup",
                            "body": [
                                {
                                    "type": "form",
                                    "api": "post:/api/backup/create",
                                    "body": [
                                        {"type": "checkboxes", "name": "backup_types", "label": "Backup Types", "options": [
                                            {"label": "Database", "value": "database"},
                                            {"label": "Files", "value": "files"},
                                            {"label": "Configuration", "value": "config"}
                                        ], "value": ["database", "config"]},
                                        {"type": "textarea", "name": "backup_notes", "label": "Backup Notes", "placeholder": "Description for this backup..."}
                                    ],
                                    "actions": [
                                        {"type": "submit", "label": "Create Backup", "level": "primary"}
                                    ]
                                }
                            ]
                        },
                        {
                            "type": "crud2",
                            "mode": "table2",
                            "api": "/api/backups",
                            "title": "Backup History",
                            "columns": [
                                {"name": "created_date", "label": "Date", "type": "datetime", "format": "MMM DD, YYYY HH:mm"},
                                {"name": "backup_types", "label": "Types", "type": "tags"},
                                {"name": "file_size", "label": "Size", "type": "bytes"},
                                {"name": "notes", "label": "Notes"},
                                {"type": "operation", "label": "Actions", "buttons": [
                                    {"type": "button", "label": "Download", "actionType": "download", "api": "/api/backups/${id}/download"},
                                    {"type": "button", "label": "Restore", "actionType": "ajax", "api": "post:/api/backups/${id}/restore", "confirmText": "Are you sure you want to restore this backup?", "level": "warning"},
                                    {"type": "button", "label": "Delete", "actionType": "ajax", "api": "delete:/api/backups/${id}", "confirmText": "Delete this backup?", "level": "danger"}
                                ]}
                            ]
                        }
                    ]
                }
            ]
        }
    ]
}
```

### Project Dashboard with Swipeable Tabs
```json
{
    "type": "tabs",
    "tabsMode": "line",
    "swipeable": true,
    "mountOnEnter": true,
    "toolbar": {
        "type": "button",
        "label": "Export Report",
        "actionType": "download",
        "api": "/api/projects/${project_id}/export",
        "level": "primary"
    },
    "tabs": [
        {
            "title": "Overview",
            "icon": "eye",
            "body": [
                {
                    "type": "grid",
                    "columns": [
                        {
                            "body": [
                                {"type": "stats", "title": "Progress", "value": "${progress_percentage}%"},
                                {"type": "stats", "title": "Tasks Complete", "value": "${completed_tasks}/${total_tasks}"},
                                {"type": "stats", "title": "Team Members", "value": "${team_count}"}
                            ]
                        },
                        {
                            "body": [
                                {"type": "chart", "api": "/api/projects/${project_id}/progress-chart", "config": {"type": "line"}}
                            ]
                        }
                    ]
                }
            ]
        },
        {
            "title": "Tasks",
            "icon": "check-square",
            "badge": {
                "mode": "text",
                "text": "${pending_tasks_count}",
                "className": "bg-orange-500",
                "visibleOn": "${pending_tasks_count > 0}"
            },
            "body": [
                {
                    "type": "crud2",
                    "mode": "table2",
                    "api": "/api/projects/${project_id}/tasks",
                    "title": "Project Tasks",
                    "columns": [
                        {"name": "title", "label": "Task", "type": "link", "href": "/tasks/${id}"},
                        {"name": "assignee", "label": "Assigned To"},
                        {"name": "priority", "label": "Priority", "type": "status"},
                        {"name": "status", "label": "Status", "type": "status"},
                        {"name": "due_date", "label": "Due Date", "type": "date"}
                    ]
                }
            ]
        },
        {
            "title": "Team",
            "icon": "users",
            "body": [
                {
                    "type": "cards",
                    "source": "/api/projects/${project_id}/team",
                    "card": {
                        "header": {
                            "title": "${name}",
                            "subTitle": "${role}",
                            "avatar": {"type": "image", "src": "${avatar_url}"}
                        },
                        "body": [
                            {"type": "text", "text": "Tasks: ${assigned_tasks_count}"},
                            {"type": "text", "text": "Completed: ${completed_tasks_count}"}
                        ]
                    }
                }
            ]
        },
        {
            "title": "Files",
            "icon": "folder",
            "body": [
                {
                    "type": "crud2",
                    "mode": "table2",
                    "api": "/api/projects/${project_id}/files",
                    "title": "Project Files",
                    "headerToolbar": [
                        {
                            "type": "button",
                            "label": "Upload File",
                            "actionType": "dialog",
                            "level": "primary",
                            "dialog": {
                                "title": "Upload Project File",
                                "body": {
                                    "type": "form",
                                    "api": "post:/api/projects/${project_id}/files",
                                    "body": [
                                        {"type": "input-file", "name": "file", "label": "Select File", "required": true, "autoUpload": false},
                                        {"type": "textarea", "name": "description", "label": "Description"}
                                    ]
                                }
                            }
                        }
                    ],
                    "columns": [
                        {"name": "filename", "label": "File Name", "type": "link"},
                        {"name": "file_size", "label": "Size", "type": "bytes"},
                        {"name": "uploaded_by", "label": "Uploaded By"},
                        {"name": "upload_date", "label": "Date", "type": "datetime"}
                    ]
                }
            ]
        }
    ]
}
```

This component provides essential tab functionality for ERP systems requiring organized content presentation, multi-panel interfaces, and complex information hierarchy management with responsive design and interactive features.