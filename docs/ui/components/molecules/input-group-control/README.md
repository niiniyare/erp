# Input Group Control Component

**FILE PURPOSE**: Grouped input control for related data entry fields  
**SCOPE**: Input grouping, combined data entry, structured forms, and related field management  
**TARGET AUDIENCE**: Developers implementing grouped data entry, structured forms, and related field collections

## 📋 Component Overview

Input Group Control provides functionality for grouping related input fields with shared validation, formatting, and behavior. Essential for structured data entry and related field management in ERP systems.

### Schema Reference
- **Primary Schema**: `InputGroupControlSchema.json`
- **Related Schemas**: `FormHorizontal.json`, `BaseApiObject.json`
- **Base Interface**: Form input control for grouped field input

## Basic Usage

```json
{
    "type": "input-group",
    "name": "address_group",
    "label": "Address Information",
    "body": [
        {"type": "input-text", "name": "street", "label": "Street"},
        {"type": "input-text", "name": "city", "label": "City"}
    ]
}
```

## Go Type Definition

```go
type InputGroupControlProps struct {
    Type               string              `json:"type"`
    Name               string              `json:"name"`
    Label              interface{}         `json:"label"`
    Body               []interface{}       `json:"body"`             // Group fields
    Gap                string              `json:"gap"`              // Field spacing
    Direction          string              `json:"direction"`        // Layout direction
    SubFormMode        string              `json:"subFormMode"`      // Subform behavior
    ShowLabel          bool                `json:"showLabel"`        // Show group label
    CollapsibleBody    interface{}         `json:"collapsibleBody"`  // Collapsible content
    Collapsible        bool                `json:"collapsible"`      // Enable collapse
    Collapsed          bool                `json:"collapsed"`        // Initial state
    ExpandIcon         string              `json:"expandIcon"`       // Expand icon
    CollapseIcon       string              `json:"collapseIcon"`     // Collapse icon
    BodyClassName      string              `json:"bodyClassName"`    // Body CSS classes
    LabelClassName     string              `json:"labelClassName"`   // Label CSS classes
    Tabs               interface{}         `json:"tabs"`             // Tab configuration
    TabsMode           string              `json:"tabsMode"`         // Tab display mode
    ValidateApi        interface{}         `json:"validateApi"`      // Group validation
    InitApi            interface{}         `json:"initApi"`          // Initialization API
    Interval           int                 `json:"interval"`         // Polling interval
    SilentPolling      bool                `json:"silentPolling"`    // Silent polling
    StopAutoRefreshWhen string             `json:"stopAutoRefreshWhen"` // Stop condition
}
```

## Essential Variants

### Basic Field Group
```json
{
    "type": "input-group",
    "name": "contact_info",
    "label": "Contact Information",
    "body": [
        {"type": "input-text", "name": "email", "label": "Email", "required": true},
        {"type": "input-text", "name": "phone", "label": "Phone"},
        {"type": "input-text", "name": "fax", "label": "Fax"}
    ],
    "gap": "md"
}
```

### Horizontal Layout Group
```json
{
    "type": "input-group",
    "name": "name_group",
    "label": "Full Name",
    "direction": "horizontal",
    "body": [
        {"type": "input-text", "name": "first_name", "label": "First Name", "required": true},
        {"type": "input-text", "name": "last_name", "label": "Last Name", "required": true}
    ],
    "gap": "sm"
}
```

### Collapsible Group
```json
{
    "type": "input-group",
    "name": "advanced_settings",
    "label": "Advanced Settings",
    "collapsible": true,
    "collapsed": true,
    "body": [
        {"type": "switch", "name": "enable_notifications", "label": "Enable Notifications"},
        {"type": "select", "name": "timezone", "label": "Timezone", "source": "/api/timezones"}
    ]
}
```

### Tabbed Group
```json
{
    "type": "input-group",
    "name": "user_details",
    "label": "User Details",
    "tabs": [
        {"title": "Basic Info", "body": [
            {"type": "input-text", "name": "username", "label": "Username"},
            {"type": "input-email", "name": "email", "label": "Email"}
        ]},
        {"title": "Profile", "body": [
            {"type": "textarea", "name": "bio", "label": "Bio"},
            {"type": "input-image", "name": "avatar", "label": "Avatar"}
        ]}
    ]
}
```

## Real-World Use Cases

### Customer Address Group
```json
{
    "type": "input-group",
    "name": "customer_address",
    "label": "Customer Address",
    "showLabel": true,
    "body": [
        {
            "type": "input-text",
            "name": "street_address",
            "label": "Street Address",
            "placeholder": "Enter street address",
            "required": true,
            "validations": {"minLength": 5},
            "validationErrors": {"minLength": "Street address must be at least 5 characters"}
        },
        {
            "type": "input-text",
            "name": "apartment",
            "label": "Apartment/Suite",
            "placeholder": "Apt, Suite, Unit (optional)"
        },
        {
            "type": "group",
            "direction": "horizontal",
            "gap": "md",
            "body": [
                {
                    "type": "input-city",
                    "name": "city",
                    "label": "City",
                    "placeholder": "Enter city",
                    "required": true
                },
                {
                    "type": "select",
                    "name": "state",
                    "label": "State/Province",
                    "placeholder": "Select state",
                    "source": "/api/states",
                    "required": true
                }
            ]
        },
        {
            "type": "group",
            "direction": "horizontal",
            "gap": "md",
            "body": [
                {
                    "type": "input-text",
                    "name": "postal_code",
                    "label": "Postal Code",
                    "placeholder": "Postal/ZIP code",
                    "required": true
                },
                {
                    "type": "select",
                    "name": "country",
                    "label": "Country",
                    "placeholder": "Select country",
                    "source": "/api/countries",
                    "value": "US",
                    "required": true
                }
            ]
        }
    ],
    "gap": "lg",
    "bodyClassName": "address-form-group",
    "validateApi": "/api/addresses/validate",
    "autoFill": {
        "api": "/api/addresses/geocode/${postal_code}",
        "fillMapping": {
            "latitude": "lat",
            "longitude": "lng",
            "timezone": "timezone"
        }
    }
}
```

### Employee Information Group
```json
{
    "type": "input-group",
    "name": "employee_info",
    "label": "Employee Information",
    "tabs": [
        {
            "title": "Personal Information",
            "body": [
                {
                    "type": "group",
                    "direction": "horizontal",
                    "gap": "md",
                    "body": [
                        {"type": "input-text", "name": "first_name", "label": "First Name", "required": true},
                        {"type": "input-text", "name": "last_name", "label": "Last Name", "required": true}
                    ]
                },
                {
                    "type": "group",
                    "direction": "horizontal",
                    "gap": "md",
                    "body": [
                        {"type": "input-date", "name": "birth_date", "label": "Date of Birth"},
                        {"type": "select", "name": "gender", "label": "Gender", "options": ["Male", "Female", "Other", "Prefer not to say"]}
                    ]
                },
                {"type": "input-text", "name": "ssn", "label": "Social Security Number", "inputMask": "999-99-9999"}
            ]
        },
        {
            "title": "Contact Information",
            "body": [
                {"type": "input-email", "name": "personal_email", "label": "Personal Email", "required": true},
                {"type": "input-text", "name": "phone_primary", "label": "Primary Phone", "inputMask": "(999) 999-9999"},
                {"type": "input-text", "name": "phone_secondary", "label": "Secondary Phone", "inputMask": "(999) 999-9999"},
                {"type": "textarea", "name": "emergency_contact", "label": "Emergency Contact", "rows": 3}
            ]
        },
        {
            "title": "Employment Details",
            "body": [
                {"type": "input-text", "name": "employee_id", "label": "Employee ID", "disabled": true},
                {"type": "input-date", "name": "hire_date", "label": "Hire Date", "required": true},
                {"type": "select", "name": "department", "label": "Department", "source": "/api/departments", "required": true},
                {"type": "select", "name": "position", "label": "Position", "source": "/api/positions", "required": true},
                {"type": "select", "name": "employment_type", "label": "Employment Type", "options": ["Full-time", "Part-time", "Contract", "Intern"]}
            ]
        }
    ],
    "tabsMode": "line",
    "required": true
}
```

### Product Specification Group
```json
{
    "type": "input-group",
    "name": "product_specifications",
    "label": "Product Specifications",
    "collapsible": true,
    "collapsed": false,
    "body": [
        {
            "type": "group",
            "direction": "horizontal",
            "gap": "md",
            "body": [
                {"type": "input-text", "name": "sku", "label": "SKU", "required": true},
                {"type": "input-text", "name": "barcode", "label": "Barcode"}
            ]
        },
        {
            "type": "group",
            "direction": "horizontal",
            "gap": "md",
            "body": [
                {"type": "input-number", "name": "weight", "label": "Weight (kg)", "min": 0, "step": 0.01},
                {"type": "input-number", "name": "length", "label": "Length (cm)", "min": 0},
                {"type": "input-number", "name": "width", "label": "Width (cm)", "min": 0},
                {"type": "input-number", "name": "height", "label": "Height (cm)", "min": 0}
            ]
        },
        {"type": "select", "name": "category", "label": "Category", "source": "/api/product-categories", "required": true},
        {"type": "tags", "name": "tags", "label": "Product Tags", "placeholder": "Add tags"},
        {"type": "textarea", "name": "description", "label": "Description", "rows": 4}
    ],
    "gap": "lg",
    "validateApi": "/api/products/validate-specifications"
}
```

### Financial Account Group
```json
{
    "type": "input-group",
    "name": "financial_account",
    "label": "Financial Account Information",
    "body": [
        {
            "type": "group",
            "direction": "horizontal",
            "gap": "md",
            "body": [
                {"type": "input-text", "name": "account_number", "label": "Account Number", "required": true},
                {"type": "select", "name": "account_type", "label": "Account Type", "options": ["Checking", "Savings", "Business"], "required": true}
            ]
        },
        {
            "type": "group",
            "direction": "horizontal",
            "gap": "md",
            "body": [
                {"type": "input-text", "name": "routing_number", "label": "Routing Number", "required": true},
                {"type": "input-text", "name": "bank_name", "label": "Bank Name", "required": true}
            ]
        },
        {"type": "input-text", "name": "account_holder_name", "label": "Account Holder Name", "required": true},
        {
            "type": "group",
            "label": "Billing Address",
            "collapsible": true,
            "body": [
                {"type": "input-text", "name": "billing_street", "label": "Street Address"},
                {
                    "type": "group",
                    "direction": "horizontal",
                    "gap": "md",
                    "body": [
                        {"type": "input-text", "name": "billing_city", "label": "City"},
                        {"type": "input-text", "name": "billing_zip", "label": "ZIP Code"}
                    ]
                }
            ]
        }
    ],
    "gap": "lg",
    "validateApi": "/api/financial/validate-account"
}
```

### Project Settings Group
```json
{
    "type": "input-group",
    "name": "project_settings",
    "label": "Project Settings",
    "tabs": [
        {
            "title": "General",
            "body": [
                {"type": "input-text", "name": "project_name", "label": "Project Name", "required": true},
                {"type": "textarea", "name": "project_description", "label": "Description", "rows": 3},
                {"type": "select", "name": "project_status", "label": "Status", "options": ["Planning", "Active", "On Hold", "Completed"]},
                {
                    "type": "group",
                    "direction": "horizontal",
                    "gap": "md",
                    "body": [
                        {"type": "input-date", "name": "start_date", "label": "Start Date"},
                        {"type": "input-date", "name": "end_date", "label": "End Date"}
                    ]
                }
            ]
        },
        {
            "title": "Team & Resources",
            "body": [
                {"type": "users-select", "name": "project_manager", "label": "Project Manager", "required": true},
                {"type": "users-select", "name": "team_members", "label": "Team Members", "multiple": true},
                {"type": "input-number", "name": "budget", "label": "Budget", "prefix": "$", "min": 0},
                {"type": "select", "name": "priority", "label": "Priority", "options": ["Low", "Medium", "High", "Critical"]}
            ]
        },
        {
            "title": "Advanced",
            "body": [
                {"type": "tags", "name": "technologies", "label": "Technologies", "placeholder": "Add technologies"},
                {"type": "switch", "name": "public_project", "label": "Public Project"},
                {"type": "switch", "name": "notifications_enabled", "label": "Enable Notifications"},
                {"type": "select", "name": "methodology", "label": "Methodology", "options": ["Agile", "Waterfall", "Kanban", "Scrum"]}
            ]
        }
    ],
    "required": true
}
```

This component provides essential grouped input functionality for ERP systems requiring structured data entry, related field management, and organized form layouts.