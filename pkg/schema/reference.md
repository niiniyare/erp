# Schema System Quick Reference

## File Guide

| File | Purpose | Key Types |
|------|---------|-----------|
| `schema.go` | Main entry point, core types | Schema, Type, Config, State, Context |
| `field.go` | All field types and configuration | Field, FieldType, FieldValidation, Option |
| `action.go` | Buttons and actions | Action, ActionType, ActionConfig, Confirm |
| `layout.go` | Visual layouts | Layout, Section, Tab, Step, Group |
| `enterprise.go` | Security, tenancy, workflows | Security, Tenant, Workflow, I18n |
| `errors.go` | Error types | ValidationError, PermissionError |
| `builder.go` | Fluent API | Builder, FieldBuilder |
| `examples.go` | Usage examples | Example functions |

## Common Patterns

### Create a Basic Form
```go
schema := models.NewBuilder("form-id", models.TypeForm, "Title").
    WithConfig(models.NewSimpleConfig("/api/endpoint", "POST")).
    AddTextField("name", "Name", true).
    AddSubmitButton("Save").
    MustBuild()
```

### Add Field with Validation
```go
builder.AddFieldWithConfig(
    models.NewField("email", models.FieldEmail).
        WithLabel("Email").
        Required().
        WithValidation(&models.FieldValidation{
            Format: "email",
            Messages: models.Messages{
                Required: "Email is required",
            },
        }).
        Build(),
)
```

### Enable Security
```go
builder.WithCSRF().
    WithRateLimit(100, 3600).
    WithTenant("tenant_id", "strict")
```

### Create Layout
```go
// Grid
builder.WithLayout(models.NewGridLayout(2))

// Tabs
builder.WithLayout(&models.Layout{
    Type: models.LayoutTabs,
    Tabs: []models.Tab{
        {ID: "tab1", Label: "Tab 1", Fields: []string{"field1"}},
    },
})

// Steps
builder.WithLayout(&models.Layout{
    Type: models.LayoutSteps,
    Steps: []models.Step{
        {ID: "step1", Title: "Step 1", Fields: []string{"field1"}},
    },
})
```

### Dynamic Options
```go
field := models.Field{
    Type: models.FieldSelect,
    DataSource: &models.DataSource{
        Type: "api",
        URL: "/api/v1/options",
        Method: "GET",
        CacheTTL: 300,
    },
}
```

### Conditional Fields
```go
field := models.Field{
    Conditional: &models.Conditional{
        Show: &models.ConditionGroup{
            Logic: "AND",
            Conditions: []models.Condition{
                {Field: "country", Operator: "==", Value: "US"},
            },
        },
    },
}
```

### HTMX Integration
```go
// Form-level
builder.WithHTMX("/api/submit", "#result")

// Field-level
field.HTMX = &models.FieldHTMX{
    Post: "/api/validate",
    Trigger: "blur",
    Target: "#errors",
}
```

### Alpine.js Integration
```go
builder.WithAlpine(`{
    count: 0,
    increment() { this.count++ }
}`)
```

## Field Types Cheat Sheet

### Basic
- `FieldText` - Text input
- `FieldEmail` - Email input
- `FieldPassword` - Password input
- `FieldNumber` - Number input
- `FieldHidden` - Hidden field

### Selection
- `FieldSelect` - Dropdown
- `FieldMultiSelect` - Multi-select dropdown
- `FieldRadio` - Radio buttons
- `FieldCheckbox` - Checkboxes
- `FieldTreeSelect` - Hierarchical select

### Date/Time
- `FieldDate` - Date picker
- `FieldTime` - Time picker
- `FieldDateTime` - Date and time
- `FieldDateRange` - Date range

### Rich Content
- `FieldTextarea` - Multi-line text
- `FieldRichText` - WYSIWYG editor
- `FieldCode` - Code editor
- `FieldJSON` - JSON editor

### Files
- `FieldFile` - File upload
- `FieldImage` - Image upload
- `FieldSignature` - Signature pad

### Specialized
- `FieldCurrency` - Money input
- `FieldPhone` - Phone number
- `FieldURL` - URL input
- `FieldTags` - Tag input
- `FieldLocation` - Location picker
- `FieldSlider` - Range slider
- `FieldRating` - Star rating
- `FieldColor` - Color picker

## Validation Rules

### String
```go
&FieldValidation{
    MinLength: intPtr(3),
    MaxLength: intPtr(50),
    Pattern: "^[a-zA-Z]+$",
}
```

### Number
```go
&FieldValidation{
    Min: float64Ptr(0),
    Max: float64Ptr(100),
    Integer: true,
    Positive: true,
}
```

### Array
```go
&FieldValidation{
    MinItems: intPtr(1),
    MaxItems: intPtr(10),
    UniqueItems: true,
}
```

## Layout Types

| Type | Use Case |
|------|----------|
| `LayoutGrid` | Multi-column form |
| `LayoutFlex` | Flexible arrangement |
| `LayoutTabs` | Tabbed interface |
| `LayoutSteps` | Multi-step wizard |
| `LayoutSections` | Grouped sections |
| `LayoutGroups` | Fieldset groups |

## Action Types

| Type | Description |
|------|-------------|
| `ActionSubmit` | Submit form |
| `ActionReset` | Reset to defaults |
| `ActionButton` | Generic button |
| `ActionLink` | Navigate to URL |
| `ActionCustom` | Custom handler |

## Action Variants

| Variant | Use Case |
|---------|----------|
| `primary` | Main action (blue) |
| `secondary` | Secondary action (gray) |
| `outline` | Outlined button |
| `ghost` | Minimal styling |
| `destructive` | Dangerous action (red) |

## Security Features

### CSRF Protection
```go
builder.WithCSRF()
// or
models.Security = &models.Security{
    CSRF: &models.CSRF{
        Enabled: true,
        FieldName: "_csrf",
    },
}
```

### Rate Limiting
```go
builder.WithRateLimit(100, 3600) // 100 requests per hour
```

### Field Encryption
```go
models.Security = &models.Security{
    Encryption: &models.Encryption{
        Enabled: true,
        Fields: []string{"ssn", "credit_card"},
        Algorithm: "AES-256-GCM",
    },
}
```

## Multi-Tenancy Modes

| Mode | Description |
|------|-------------|
| `strict` | Complete isolation (RLS) |
| `shared` | Shared across tenants |
| `hybrid` | Mix of both |

```go
builder.WithTenant("tenant_id", "strict")
```

## Workflow Actions

```go
models.Workflow = &models.Workflow{
    Enabled: true,
    Actions: []models.WorkflowAction{
        {
            ID: "approve",
            Label: "Approve",
            Type: "approve",
            ToStage: "approved",
            Permissions: []string{"can_approve"},
        },
    },
}
```

## Common Helpers

```go
// Create float64 pointer
func float64Ptr(f float64) *float64 { return &f }

// Create int pointer
func intPtr(i int) *int { return &i }

// Create option
models.CreateOption("value", "label")

// Create option with icon
models.CreateOptionWithIcon("value", "label", "icon-name")

// Create grouped option
models.CreateGroupedOption("value", "label", "group")
```

## Error Handling

```go
// Build with error check
schema, err := builder.Build()
if err != nil {
    // Handle validation collection
    if vecErr, ok := err.(*models.ValidationErrorCollection); ok {
        log.Printf("Validation failed: %d errors", vecErr.Count())
        for field, errors := range vecErr.ErrorsByField() {
            log.Printf("  %s: %v", field, errors)
        }
    }
    
    // Check error type
    if models.IsValidationError(err) {
        // Validation error
    } else if models.IsNotFoundError(err) {
        // Not found error
    } else if models.IsPermissionError(err) {
        // Permission error
    }
    
    // Get HTTP status and response
    statusCode := models.GetHTTPStatusCode(err)
    response := models.ToErrorResponse(err)
}

// Build or panic
schema := builder.MustBuild()

// Create error response for API
errorResponse := models.ToErrorResponse(err)
// Returns: { error, code, type, field, details }

// Create multi-error response
multiResponse := models.ToMultiErrorResponse(validationErrors)
// Returns: { errors[], count, fieldErrors{}, summary }
```

## JSON Schema

### Load from JSON
```go
var schema models.Schema
if err := json.Unmarshal(data, &schema); err != nil {
    // Handle error
}
```

### Save to JSON
```go
data, err := json.Marshal(schema)
```

### Example JSON
```json
{
  "id": "user-form",
  "type": "form",
  "version": "1.0.0",
  "title": "Create User",
  "config": {
    "action": "/api/v1/users",
    "method": "POST"
  },
  "fields": [
    {
      "name": "email",
      "type": "email",
      "label": "Email",
      "required": true
    }
  ],
  "actions": [
    {
      "id": "submit",
      "type": "submit",
      "text": "Create",
      "variant": "primary"
    }
  ]
}
```

## Interfaces to Implement

### Validator
```go
type MyValidator struct{}

func (v *MyValidator) ValidateSchema(ctx context.Context, s *models.Schema) error {
    // Validate schema structure
}

func (v *MyValidator) ValidateData(ctx context.Context, s *models.Schema, data map[string]any) error {
    // Validate submitted data
}
```

### Renderer
```go
type MyRenderer struct{}

func (r *MyRenderer) Render(ctx context.Context, s *models.Schema, data map[string]any) (string, error) {
    // Render schema to HTML
}

func (r *MyRenderer) RenderField(ctx context.Context, f *models.Field, value any) (string, error) {
    // Render individual field
}
```

### Registry
```go
type MyRegistry struct{}

func (r *MyRegistry) Register(ctx context.Context, s *models.Schema) error {
    // Store schema
}

func (r *MyRegistry) Get(ctx context.Context, id string) (*models.Schema, error) {
    // Retrieve schema
}
```

## Tips

1. **Use Builder**: More readable than raw structs
2. **Validate Early**: Call `MustBuild()` during development
3. **Add Help Text**: Guide users with descriptions
4. **Group Fields**: Use sections/tabs for better UX
5. **Cache Data**: Set CacheTTL for dynamic options
6. **Test Conditions**: Verify conditional logic thoroughly
7. **Enable Security**: Always use CSRF for forms
8. **Document Custom**: Add examples for complex schemas
