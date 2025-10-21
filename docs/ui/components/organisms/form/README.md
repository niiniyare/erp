# Form Component

**FILE PURPOSE**: Complete form interface for data input, validation, and submission  
**SCOPE**: Data collection, validation, submission workflows, and business process forms  
**TARGET AUDIENCE**: Developers implementing data entry forms, business processes, and user input interfaces

## 📋 Component Overview

Form provides comprehensive form functionality with validation, submission handling, layout options, and advanced features like async submission, persistence, and multi-step workflows. Essential for all data collection needs in ERP systems.

### Schema Reference
- **Primary Schema**: `FormSchema.json`
- **Related Schemas**: `FormHorizontal.json`, `ActionSchema.json`, `FormControlSchema.json`
- **Base Interface**: Complete form organism for data collection

## Basic Usage

```json
{
    "type": "form",
    "api": "post:/api/customers",
    "body": [
        {"type": "input-text", "name": "name", "label": "Name", "required": true},
        {"type": "input-email", "name": "email", "label": "Email", "required": true}
    ]
}
```

## Go Type Definition

```go
type FormProps struct {
    Type                        string              `json:"type"`
    Title                       string              `json:"title"`               // Form title
    Body                        []interface{}       `json:"body"`                // Form fields
    Actions                     []interface{}       `json:"actions"`             // Form buttons
    
    // API Configuration
    API                         interface{}         `json:"api"`                 // Submit API
    InitAPI                     interface{}         `json:"initApi"`             // Init data API
    AsyncAPI                    interface{}         `json:"asyncApi"`            // Async submit API
    InitAsyncAPI               interface{}         `json:"initAsyncApi"`        // Async init API
    
    // Submission Behavior
    SubmitText                  string              `json:"submitText"`          // Submit button text
    SubmitOnChange              bool                `json:"submitOnChange"`      // Auto submit on change
    SubmitOnInit                bool                `json:"submitOnInit"`        // Submit on init
    ResetAfterSubmit            bool                `json:"resetAfterSubmit"`    // Reset after submit
    ClearAfterSubmit            bool                `json:"clearAfterSubmit"`    // Clear after submit
    
    // Layout and Display
    Mode                        string              `json:"mode"`                // "normal", "inline", "horizontal", "flex"
    ColumnCount                 int                 `json:"columnCount"`         // Column layout
    Horizontal                  interface{}         `json:"horizontal"`          // Horizontal layout config
    WrapWithPanel               bool                `json:"wrapWithPanel"`       // Panel wrapper
    AffixFooter                 bool                `json:"affixFooter"`         // Fixed footer
    
    // Validation
    Rules                       []interface{}       `json:"rules"`               // Validation rules
    Messages                    interface{}         `json:"messages"`            // Error messages
    PreventEnterSubmit          bool                `json:"preventEnterSubmit"`  // Disable enter submit
    
    // Initialization
    InitFetch                   bool                `json:"initFetch"`           // Fetch on init
    InitFetchOn                 string              `json:"initFetchOn"`         // Init condition
    Data                        interface{}         `json:"data"`                // Initial data
    
    // Polling and Async
    Interval                    int                 `json:"interval"`            // Polling interval
    SilentPolling               bool                `json:"silentPolling"`       // Silent polling
    StopAutoRefreshWhen         string              `json:"stopAutoRefreshWhen"` // Stop condition
    CheckInterval               int                 `json:"checkInterval"`       // Async check interval
    FinishedField               string              `json:"finishedField"`       // Completion field
    InitFinishedField           string              `json:"initFinishedField"`   // Init completion field
    InitCheckInterval           int                 `json:"initCheckInterval"`   // Init check interval
    
    // Persistence
    PersistData                 string              `json:"persistData"`         // Local storage
    PersistDataKeys             []string            `json:"persistDataKeys"`     // Storage keys
    ClearPersistDataAfterSubmit bool                `json:"clearPersistDataAfterSubmit"` // Clear on submit
    
    // Navigation
    Target                      string              `json:"target"`              // Target component
    Redirect                    string              `json:"redirect"`            // Redirect URL
    Reload                      string              `json:"reload"`              // Reload target
    
    // UX Features
    AutoFocus                   bool                `json:"autoFocus"`           // Auto focus first field
    PromptPageLeave             bool                `json:"promptPageLeave"`     // Page leave warning
    PromptPageLeaveMessage      string              `json:"promptPageLeaveMessage"` // Custom warning
    
    // Development
    Debug                       bool                `json:"debug"`               // Debug mode
    DebugConfig                 interface{}         `json:"debugConfig"`         // Debug settings
    
    // Styling
    PanelClassName              string              `json:"panelClassName"`      // Panel CSS
    LabelAlign                  string              `json:"labelAlign"`          // Label alignment
    LabelWidth                  interface{}         `json:"labelWidth"`          // Label width
    
    // Advanced
    PrimaryField                string              `json:"primaryField"`        // Primary key field
    Feedback                    interface{}         `json:"feedback"`            // Success feedback
    Name                        string              `json:"name"`                // Form name
}
```

## Layout Modes

### Normal Layout (Default)
```json
{
    "type": "form",
    "api": "post:/api/employees",
    "title": "Employee Registration",
    "body": [
        {"type": "input-text", "name": "first_name", "label": "First Name", "required": true},
        {"type": "input-text", "name": "last_name", "label": "Last Name", "required": true},
        {"type": "input-email", "name": "email", "label": "Email", "required": true},
        {"type": "select", "name": "department", "label": "Department", "source": "/api/departments"}
    ]
}
```

### Horizontal Layout
```json
{
    "type": "form",
    "mode": "horizontal",
    "horizontal": {
        "left": 3,
        "right": 9
    },
    "api": "post:/api/customers",
    "body": [
        {"type": "input-text", "name": "company", "label": "Company Name", "required": true},
        {"type": "input-email", "name": "email", "label": "Email", "required": true}
    ]
}
```

### Multi-Column Layout
```json
{
    "type": "form",
    "columnCount": 2,
    "api": "post:/api/contacts",
    "body": [
        {"type": "input-text", "name": "first_name", "label": "First Name"},
        {"type": "input-text", "name": "last_name", "label": "Last Name"},
        {"type": "input-email", "name": "email", "label": "Email"},
        {"type": "input-text", "name": "phone", "label": "Phone"}
    ]
}
```

### Inline Layout
```json
{
    "type": "form",
    "mode": "inline",
    "api": "get:/api/search",
    "submitText": "Search",
    "body": [
        {"type": "input-text", "name": "keyword", "placeholder": "Search..."},
        {"type": "select", "name": "category", "placeholder": "Category", "source": "/api/categories"}
    ]
}
```

## Real-World Use Cases

### Customer Registration Form
```json
{
    "type": "form",
    "title": "Customer Registration",
    "api": "post:/api/customers",
    "initApi": "/api/form-defaults/customer",
    "mode": "horizontal",
    "horizontal": {
        "left": 3,
        "right": 9
    },
    "wrapWithPanel": true,
    "affixFooter": true,
    "autoFocus": true,
    "promptPageLeave": true,
    "promptPageLeaveMessage": "You have unsaved changes. Are you sure you want to leave?",
    "persistData": "customer-registration",
    "clearPersistDataAfterSubmit": true,
    "body": [
        {
            "type": "fieldset",
            "title": "Company Information",
            "collapsible": false,
            "body": [
                {
                    "type": "input-text",
                    "name": "company_name",
                    "label": "Company Name",
                    "required": true,
                    "validations": {"minLength": 2},
                    "validationErrors": {"minLength": "Company name must be at least 2 characters"}
                },
                {
                    "type": "input-text",
                    "name": "tax_id",
                    "label": "Tax ID",
                    "placeholder": "Federal Tax ID Number"
                },
                {
                    "type": "select",
                    "name": "industry",
                    "label": "Industry",
                    "source": "/api/industries",
                    "searchable": true,
                    "clearable": true
                },
                {
                    "type": "select",
                    "name": "company_size",
                    "label": "Company Size",
                    "options": [
                        "1-10 employees",
                        "11-50 employees", 
                        "51-200 employees",
                        "201-500 employees",
                        "500+ employees"
                    ]
                }
            ]
        },
        {
            "type": "fieldset",
            "title": "Primary Contact",
            "body": [
                {
                    "type": "group",
                    "direction": "horizontal",
                    "body": [
                        {"type": "input-text", "name": "contact_first_name", "label": "First Name", "required": true},
                        {"type": "input-text", "name": "contact_last_name", "label": "Last Name", "required": true}
                    ]
                },
                {
                    "type": "input-text",
                    "name": "contact_title",
                    "label": "Job Title"
                },
                {
                    "type": "input-email",
                    "name": "contact_email",
                    "label": "Email Address",
                    "required": true,
                    "validations": {"isEmail": true},
                    "validationErrors": {"isEmail": "Please enter a valid email address"}
                },
                {
                    "type": "input-text",
                    "name": "contact_phone",
                    "label": "Phone Number",
                    "inputMask": "(999) 999-9999"
                }
            ]
        },
        {
            "type": "fieldset",
            "title": "Business Address",
            "body": [
                {
                    "type": "input-text",
                    "name": "address_street",
                    "label": "Street Address",
                    "required": true
                },
                {
                    "type": "input-text",
                    "name": "address_suite",
                    "label": "Suite/Unit",
                    "placeholder": "Apartment, suite, unit, building, floor, etc."
                },
                {
                    "type": "group",
                    "direction": "horizontal",
                    "body": [
                        {"type": "input-city", "name": "address_city", "label": "City", "required": true},
                        {"type": "select", "name": "address_state", "label": "State", "source": "/api/states", "required": true}
                    ]
                },
                {
                    "type": "group",
                    "direction": "horizontal",
                    "body": [
                        {"type": "input-text", "name": "address_zip", "label": "ZIP Code", "required": true},
                        {"type": "select", "name": "address_country", "label": "Country", "value": "US", "source": "/api/countries"}
                    ]
                }
            ]
        },
        {
            "type": "fieldset",
            "title": "Account Preferences",
            "body": [
                {
                    "type": "select",
                    "name": "preferred_contact_method",
                    "label": "Preferred Contact Method",
                    "options": ["Email", "Phone", "Mail"],
                    "value": "Email"
                },
                {
                    "type": "select",
                    "name": "payment_terms",
                    "label": "Payment Terms",
                    "options": ["Net 30", "Net 15", "Due on Receipt", "Custom"],
                    "value": "Net 30"
                },
                {
                    "type": "switch",
                    "name": "email_marketing",
                    "label": "Email Marketing",
                    "option": "Receive marketing emails and product updates"
                },
                {
                    "type": "textarea",
                    "name": "notes",
                    "label": "Additional Notes",
                    "placeholder": "Any additional information about this customer...",
                    "rows": 3
                }
            ]
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
            "label": "Save Draft",
            "actionType": "ajax",
            "api": "post:/api/customers/draft",
            "level": "secondary"
        },
        {
            "type": "submit",
            "label": "Register Customer",
            "level": "primary"
        }
    ],
    "feedback": {
        "title": "Customer Registered Successfully!",
        "body": "Customer ${company_name} has been registered. You can now create orders and manage their account."
    },
    "redirect": "/customers/${id}",
    "rules": [
        {
            "rule": "data.contact_email && data.company_name",
            "message": "Both company name and contact email are required"
        }
    ]
}
```

### Product Creation Form
```json
{
    "type": "form",
    "title": "Add New Product",
    "api": "post:/api/products",
    "columnCount": 2,
    "wrapWithPanel": true,
    "submitText": "Create Product",
    "resetAfterSubmit": true,
    "autoFocus": true,
    "body": [
        {
            "type": "input-text",
            "name": "name",
            "label": "Product Name",
            "required": true,
            "columnCount": 2,
            "validations": {"minLength": 3},
            "validationErrors": {"minLength": "Product name must be at least 3 characters"}
        },
        {
            "type": "input-text",
            "name": "sku",
            "label": "SKU",
            "required": true,
            "placeholder": "Product SKU code"
        },
        {
            "type": "select",
            "name": "category_id",
            "label": "Category",
            "source": "/api/categories",
            "required": true
        },
        {
            "type": "input-number",
            "name": "price",
            "label": "Price",
            "prefix": "$",
            "min": 0,
            "step": 0.01,
            "required": true
        },
        {
            "type": "input-number",
            "name": "cost",
            "label": "Cost",
            "prefix": "$",
            "min": 0,
            "step": 0.01
        },
        {
            "type": "input-number",
            "name": "weight",
            "label": "Weight",
            "suffix": "lbs",
            "min": 0,
            "step": 0.1
        },
        {
            "type": "input-number",
            "name": "stock_quantity",
            "label": "Initial Stock",
            "min": 0,
            "step": 1
        },
        {
            "type": "textarea",
            "name": "description",
            "label": "Description",
            "columnCount": 2,
            "rows": 4,
            "maxLength": 500
        },
        {
            "type": "input-image",
            "name": "images",
            "label": "Product Images",
            "columnCount": 2,
            "multiple": true,
            "maxLength": 5,
            "autoUpload": true,
            "fileApi": "/api/products/images/upload"
        },
        {
            "type": "tags",
            "name": "tags",
            "label": "Product Tags",
            "columnCount": 2,
            "placeholder": "Add tags for better searchability"
        },
        {
            "type": "switch",
            "name": "is_active",
            "label": "Active Product",
            "value": true,
            "option": "Make this product available for sale"
        },
        {
            "type": "switch",
            "name": "track_inventory",
            "label": "Track Inventory",
            "value": true,
            "option": "Monitor stock levels for this product"
        }
    ],
    "feedback": {
        "title": "Product Created!",
        "body": "Product ${name} has been created successfully with SKU ${sku}."
    },
    "redirect": "/products/${id}"
}
```

### Employee Onboarding Form
```json
{
    "type": "form",
    "title": "Employee Onboarding",
    "api": "post:/api/employees",
    "initApi": "/api/form-templates/employee-onboarding",
    "mode": "horizontal",
    "horizontal": {"left": 3, "right": 9},
    "wrapWithPanel": true,
    "affixFooter": true,
    "persistData": "employee-onboarding",
    "promptPageLeave": true,
    "debug": false,
    "body": [
        {
            "type": "tabs",
            "tabs": [
                {
                    "title": "Personal Information",
                    "body": [
                        {
                            "type": "group",
                            "direction": "horizontal",
                            "body": [
                                {"type": "input-text", "name": "first_name", "label": "First Name", "required": true},
                                {"type": "input-text", "name": "last_name", "label": "Last Name", "required": true}
                            ]
                        },
                        {
                            "type": "input-email",
                            "name": "personal_email",
                            "label": "Personal Email",
                            "required": true
                        },
                        {
                            "type": "input-text",
                            "name": "phone",
                            "label": "Phone Number",
                            "inputMask": "(999) 999-9999"
                        },
                        {
                            "type": "input-date",
                            "name": "birth_date",
                            "label": "Date of Birth"
                        },
                        {
                            "type": "input-text",
                            "name": "ssn",
                            "label": "Social Security Number",
                            "inputMask": "999-99-9999",
                            "required": true
                        }
                    ]
                },
                {
                    "title": "Employment Details",
                    "body": [
                        {
                            "type": "input-text",
                            "name": "employee_id",
                            "label": "Employee ID",
                            "disabled": true,
                            "value": "${auto_generated_id}"
                        },
                        {
                            "type": "input-date",
                            "name": "start_date",
                            "label": "Start Date",
                            "required": true
                        },
                        {
                            "type": "select",
                            "name": "department_id",
                            "label": "Department",
                            "source": "/api/departments",
                            "required": true
                        },
                        {
                            "type": "select",
                            "name": "position_id",
                            "label": "Position",
                            "source": "/api/positions?department_id=${department_id}",
                            "required": true
                        },
                        {
                            "type": "users-select",
                            "name": "manager_id",
                            "label": "Direct Manager",
                            "source": "/api/managers?department_id=${department_id}"
                        },
                        {
                            "type": "select",
                            "name": "employment_type",
                            "label": "Employment Type",
                            "options": ["Full-time", "Part-time", "Contract", "Intern"],
                            "required": true
                        },
                        {
                            "type": "input-number",
                            "name": "salary",
                            "label": "Annual Salary",
                            "prefix": "$",
                            "min": 0
                        }
                    ]
                },
                {
                    "title": "System Access",
                    "body": [
                        {
                            "type": "input-email",
                            "name": "work_email",
                            "label": "Work Email",
                            "required": true,
                            "validations": {"isEmail": true}
                        },
                        {
                            "type": "input-text",
                            "name": "username",
                            "label": "Username",
                            "required": true,
                            "validations": {"minLength": 3}
                        },
                        {
                            "type": "input-password",
                            "name": "temporary_password",
                            "label": "Temporary Password",
                            "required": true,
                            "showStrength": true
                        },
                        {
                            "type": "checkboxes",
                            "name": "system_access",
                            "label": "System Access",
                            "options": [
                                {"label": "Email System", "value": "email"},
                                {"label": "HR Portal", "value": "hr"},
                                {"label": "Time Tracking", "value": "timesheet"},
                                {"label": "Project Management", "value": "projects"}
                            ]
                        }
                    ]
                },
                {
                    "title": "Documents & Agreements",
                    "body": [
                        {
                            "type": "checkboxes",
                            "name": "required_documents",
                            "label": "Required Documents",
                            "required": true,
                            "options": [
                                {"label": "Employee Handbook Acknowledgment", "value": "handbook"},
                                {"label": "Code of Conduct Agreement", "value": "conduct"},
                                {"label": "Confidentiality Agreement", "value": "confidentiality"},
                                {"label": "Safety Training Completion", "value": "safety"}
                            ],
                            "validations": {"minLength": 4},
                            "validationErrors": {"minLength": "All required documents must be acknowledged"}
                        },
                        {
                            "type": "input-signature",
                            "name": "employee_signature",
                            "label": "Employee Signature",
                            "required": true,
                            "placeholder": "Please sign to confirm all information is accurate"
                        },
                        {
                            "type": "textarea",
                            "name": "additional_notes",
                            "label": "Additional Notes",
                            "placeholder": "Any additional information or special accommodations needed..."
                        }
                    ]
                }
            ]
        }
    ],
    "actions": [
        {
            "type": "button",
            "label": "Save Draft",
            "actionType": "ajax",
            "api": "post:/api/employees/draft",
            "level": "secondary"
        },
        {
            "type": "submit",
            "label": "Complete Onboarding",
            "level": "primary"
        }
    ],
    "feedback": {
        "title": "Employee Onboarding Complete!",
        "body": "Welcome ${first_name} ${last_name}! Your employee profile has been created. Please check your work email for login credentials and next steps."
    },
    "redirect": "/employees/${id}",
    "rules": [
        {
            "rule": "data.required_documents && data.required_documents.length === 4",
            "message": "All required documents must be acknowledged before completing onboarding"
        },
        {
            "rule": "data.employee_signature",
            "message": "Employee signature is required to complete onboarding"
        }
    ]
}
```

### Invoice Creation Form with Async Processing
```json
{
    "type": "form",
    "title": "Create Invoice",
    "api": "post:/api/invoices",
    "asyncApi": "get:/api/invoices/${id}/status",
    "checkInterval": 2000,
    "finishedField": "processing_complete",
    "wrapWithPanel": true,
    "submitText": "Generate Invoice",
    "body": [
        {
            "type": "select",
            "name": "customer_id",
            "label": "Customer",
            "source": "/api/customers",
            "required": true,
            "autoFill": {
                "api": "/api/customers/${value}",
                "fillMapping": {
                    "billing_address": "address",
                    "payment_terms": "default_payment_terms"
                }
            }
        },
        {
            "type": "input-date",
            "name": "invoice_date",
            "label": "Invoice Date",
            "value": "${current_date}",
            "required": true
        },
        {
            "type": "input-date",
            "name": "due_date",
            "label": "Due Date",
            "required": true
        },
        {
            "type": "combo",
            "name": "line_items",
            "label": "Line Items",
            "multiple": true,
            "addable": true,
            "removable": true,
            "items": [
                {
                    "type": "select",
                    "name": "product_id",
                    "label": "Product",
                    "source": "/api/products",
                    "required": true,
                    "autoFill": {
                        "api": "/api/products/${value}",
                        "fillMapping": {
                            "unit_price": "price",
                            "description": "name"
                        }
                    }
                },
                {
                    "type": "input-number",
                    "name": "quantity",
                    "label": "Quantity",
                    "min": 1,
                    "required": true
                },
                {
                    "type": "input-number",
                    "name": "unit_price",
                    "label": "Unit Price",
                    "prefix": "$",
                    "min": 0,
                    "step": 0.01,
                    "required": true
                }
            ]
        },
        {
            "type": "textarea",
            "name": "notes",
            "label": "Notes",
            "placeholder": "Additional notes for this invoice..."
        }
    ],
    "feedback": {
        "title": "Invoice Generated Successfully!",
        "body": "Invoice #${invoice_number} has been created and will be processed. You can view it in the invoices section."
    },
    "redirect": "/invoices/${id}"
}
```

This component provides essential form functionality for ERP systems requiring comprehensive data collection, validation, submission workflows, and business process automation.