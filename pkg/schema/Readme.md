# Schema System

A comprehensive JSON-driven UI schema system for building enterprise forms and components. Backend developers can define complete UIs by writing JSON files - no frontend code required.

## Overview

The schema system enables:
- **JSON-First Development**: Define forms entirely in JSON
- **Multi-Tenancy**: Built-in tenant isolation with strict/shared/hybrid modes
- **Enterprise Security**: CSRF protection, rate limiting, field encryption
- **Workflow Support**: Approval workflows with configurable stages
- **Framework Integration**: HTMX and Alpine.js support out of the box
- **Type Safety**: Comprehensive Go structs with validation

## File Structure

```
web/schema/models/
      ├── schema.go           # Main schema definition and core types
      ├── field.go            # Field types and configuration
      ├── action.go           # Action/button types
      ├── layout.go           # Layout types (grid, tabs, steps, sections)
      ├── enterprise.go       # Enterprise features (security, tenant, workflow)
      ├── errors.go           # Error types and handling
      ├── builder.go          # Fluent builder pattern for programmatic creation
      └── examples.go         # Usage examples and patterns
```

## Quick Start

### Basic Form

```go
schema := models.NewBuilder("user-form", models.TypeForm, "Create User").
    WithConfig(models.NewSimpleConfig("/api/v1/users", "POST")).
    WithCSRF().
    AddTextField("name", "Full Name", true).
    AddEmailField("email", "Email", true).
    AddPasswordField("password", "Password", true).
    AddSubmitButton("Create").
    MustBuild()
```

### With Layout

```go
schema := models.NewBuilder("product-form", models.TypeForm, "Add Product").
    WithLayout(models.NewGridLayout(2)).
    AddTextField("name", "Product Name", true).
    AddTextField("sku", "SKU", true).
    AddNumberField("price", "Price", true, floatPtr(0), nil).
    AddSelectField("category", "Category", true, []models.Option{
        {Value: "electronics", Label: "Electronics"},
        {Value: "clothing", Label: "Clothing"},
    }).
    AddSubmitButton("Save").
    MustBuild()
```

### Multi-Step Wizard

```go
schema := models.NewBuilder("onboarding", models.TypeWorkflow, "Employee Onboarding").
    WithLayout(&models.Layout{
        Type: models.LayoutSteps,
        Steps: []models.Step{
            {
                ID:     "personal",
                Title:  "Personal Info",
                Fields: []string{"name", "email"},
                Order:  0,
            },
            {
                ID:     "employment",
                Title:  "Employment",
                Fields: []string{"job_title", "department"},
                Order:  1,
            },
        },
    }).
    AddTextField("name", "Full Name", true).
    AddEmailField("email", "Email", true).
    AddTextField("job_title", "Job Title", true).
    AddTextField("department", "Department", true).
    AddSubmitButton("Complete").
    MustBuild()
```

## Field Types

### Basic Inputs
- `FieldText` - Text input
- `FieldEmail` - Email input
- `FieldPassword` - Password input
- `FieldNumber` - Number input
- `FieldPhone` - Phone number input
- `FieldURL` - URL input

### Date & Time
- `FieldDate` - Date picker
- `FieldTime` - Time picker
- `FieldDateTime` - Date and time picker
- `FieldDateRange` - Date range picker

### Selection
- `FieldSelect` - Dropdown select
- `FieldMultiSelect` - Multiple selection dropdown
- `FieldRadio` - Radio buttons
- `FieldCheckbox` - Checkboxes
- `FieldTreeSelect` - Hierarchical tree select
- `FieldCascader` - Cascading dropdown

### Rich Content
- `FieldTextarea` - Multi-line text
- `FieldRichText` - WYSIWYG editor
- `FieldCode` - Code editor with syntax highlighting
- `FieldJSON` - JSON editor

### Files
- `FieldFile` - File upload
- `FieldImage` - Image upload with preview
- `FieldSignature` - Signature pad

### Specialized
- `FieldCurrency` - Money input with formatting
- `FieldTags` - Tag input
- `FieldLocation` - Location/address picker
- `FieldSlider` - Range slider
- `FieldRating` - Star rating
- `FieldColor` - Color picker

## Validation

### Field-Level Validation

```go
field := models.NewField("username", models.FieldText).
    WithLabel("Username").
    Required().
    WithValidation(&models.FieldValidation{
        MinLength: intPtr(3),
        MaxLength: intPtr(50),
        Pattern:   "^[a-zA-Z0-9_]+$",
        Messages: models.Messages{
            Required:  "Username is required",
            MinLength: "Must be at least 3 characters",
            Pattern:   "Only letters, numbers, and underscores allowed",
        },
    }).
    Build()
```

### Cross-Field Validation

```go
models.Validation = &models.Validation{
    Rules: []models.ValidationRule{
        {
            ID:      "password-match",
            Type:    "compare",
            Fields:  []string{"password", "confirm_password"},
            Message: "Passwords must match",
        },
    },
    Mode: "onBlur",
}
```

## Enterprise Features

### Multi-Tenancy

```go
schema := models.NewBuilder("invoice", models.TypeForm, "Create Invoice").
    WithTenant("tenant_id", "strict"). // strict | shared | hybrid
    AddTextField("invoice_number", "Invoice #", true).
    MustBuild()
```

### Security

```go
schema := models.NewBuilder("payment", models.TypeForm, "Process Payment").
    WithCSRF().
    WithRateLimit(10, 60). // 10 requests per 60 seconds
    WithSecurity(&models.Security{
        Encryption: &models.Encryption{
            Enabled:   true,
            Fields:    []string{"card_number", "cvv"},
            Algorithm: "AES-256-GCM",
        },
    }).
    MustBuild()
```

### Workflows

```go
models.Workflow = &models.Workflow{
    Enabled: true,
    Actions: []models.WorkflowAction{
        {
            ID:       "approve",
            Label:    "Approve",
            Type:     "approve",
            ToStage:  "approved",
            Permissions: []string{"expense.approve"},
        },
    },
    Approvals: &models.ApprovalConfig{
        Required:     true,
        MinApprovals: 2,
        Approvers: []models.ApproverConfig{
            {Type: "role", Value: "manager"},
        },
    },
}
```

## Conditional Logic

### Conditional Display

```go
field := models.Field{
    Name:  "international_phone",
    Type:  models.FieldPhone,
    Label: "International Phone",
    Conditional: &models.Conditional{
        Show: &models.ConditionGroup{
            Logic: "AND",
            Conditions: []models.Condition{
                {Field: "country", Operator: "!=", Value: "US"},
            },
        },
    },
}
```

### Conditional Requirements

```go
field := models.Field{
    Name:  "passport_number",
    Type:  models.FieldText,
    Label: "Passport Number",
    Conditional: &models.Conditional{
        Required: &models.ConditionGroup{
            Logic: "AND",
            Conditions: []models.Condition{
                {Field: "international_travel", Operator: "==", Value: true},
            },
        },
    },
}
```

## Framework Integration

### HTMX

```go
schema := models.NewBuilder("user-form", models.TypeForm, "Create User").
    WithHTMX("/api/v1/users", "#content").
    MustBuild()

// Or per-field
field := models.Field{
    Name: "email",
    Type: models.FieldEmail,
    HTMX: &models.FieldHTMX{
        Post:    "/api/v1/validate-email",
        Trigger: "blur",
        Target:  "#email-validation",
    },
}
```

### Alpine.js

```go
schema := models.NewBuilder("calculator", models.TypeForm, "Calculator").
    WithAlpine(`{
        quantity: 0,
        price: 0,
        get total() { return this.quantity * this.price }
    }`).
    MustBuild()
```

## Data Sources

### Dynamic Options

```go
field := models.Field{
    Name:  "customer",
    Type:  models.FieldSelect,
    Label: "Customer",
    DataSource: &models.DataSource{
        Type:     "api",
        URL:      "/api/v1/customers",
        Method:   "GET",
        CacheTTL: 300, // Cache for 5 minutes
    },
}
```

## Layouts

### Grid Layout

```go
layout := &models.Layout{
    Type:    models.LayoutGrid,
    Columns: 3,
    Gap:     "1rem",
    Responsive: true,
}
```

### Sections

```go
layout := &models.Layout{
    Type: models.LayoutSections,
    Sections: []models.Section{
        {
            ID:     "basic",
            Title:  "Basic Information",
            Fields: []string{"name", "email"},
        },
        {
            ID:     "advanced",
            Title:  "Advanced Settings",
            Fields: []string{"role", "permissions"},
            Collapsible: true,
        },
    },
}
```

### Tabs

```go
layout := &models.Layout{
    Type: models.LayoutTabs,
    Tabs: []models.Tab{
        {
            ID:     "profile",
            Label:  "Profile",
            Icon:   "user",
            Fields: []string{"name", "email"},
        },
        {
            ID:     "settings",
            Label:  "Settings",
            Icon:   "settings",
            Fields: []string{"theme", "language"},
        },
    },
}
```

## Best Practices

1. **Use the Builder**: The fluent builder pattern makes schema creation readable
2. **Validate Early**: Call `Validate()` or use `MustBuild()` to catch errors
3. **Leverage Types**: Use the provided field types for automatic validation
4. **Group Related Fields**: Use sections or tabs for better UX
5. **Add Help Text**: Use `Help` and `Description` to guide users
6. **Enable Security**: Always use CSRF for forms that modify data
7. **Cache Data Sources**: Set appropriate `CacheTTL` for dynamic options
8. **Test Conditions**: Thoroughly test conditional logic before deployment

## Error Handling

The system uses a comprehensive error handling framework with typed errors:

```go
schema, err := builder.Build()
if err != nil {
    // Check error type
    if models.IsValidationError(err) {
        // Handle validation errors
        if vecErr, ok := err.(*models.ValidationErrorCollection); ok {
            // Get errors by field
            fieldErrors := vecErr.ErrorsByField()
            for field, errors := range fieldErrors {
                log.Printf("Field %s has %d errors", field, len(errors))
            }
        }
    }
    
    // Check specific error types
    if models.IsNotFoundError(err) {
        // Handle not found
    }
    if models.IsPermissionError(err) {
        // Handle permission denied
    }
    
    // Get error details
    code := models.GetErrorCode(err)
    details := models.GetErrorDetails(err)
    
    // Convert to HTTP response
    statusCode := models.GetHTTPStatusCode(err)
    response := models.ToErrorResponse(err)
}
```

### Error Types

- `ErrorTypeValidation` - Validation failures (400)
- `ErrorTypeNotFound` - Resource not found (404)
- `ErrorTypeConflict` - Duplicate resources (409)
- `ErrorTypePermission` - Access denied (403)
- `ErrorTypeDataSource` - Data fetch failures (502)
- `ErrorTypeRender` - Rendering failures (500)
- `ErrorTypeWorkflow` - Workflow errors (422)
- `ErrorTypeTenant` - Tenant isolation violations (403)
- `ErrorTypeInternal` - Internal errors (500)

## Performance Tips

1. **Pre-compile Conditions**: Compile conditions once and cache them
2. **Use Caching**: Enable caching for data sources that don't change frequently
3. **Lazy Load**: Use tabs/steps to load fields only when needed
4. **Validate Incrementally**: Use `onBlur` validation mode for better UX
5. **Optimize Layouts**: Use appropriate column counts for your data

## Contributing

When adding new field types or features:
1. Update the appropriate file (`field.go`, `action.go`, etc.)
2. Add validation logic
3. Create usage examples in `examples.go`
4. Update this README

