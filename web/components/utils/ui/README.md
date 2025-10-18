# UI Schema Package

A production-ready, type-safe UI component schema system for the ERP application.

## Overview

This package provides:
- **Type-safe component definitions** with Go structs and validation
- **Factory pattern** for component creation with proper validation
- **Fluent builder interface** for easy component construction
- **Component registry** for managing all UI component types
- **Comprehensive component types** covering forms, layouts, and data display

## Key Benefits vs Generated Schemas

- ✅ **Type Safety**: Proper Go types with compile-time checking
- ✅ **Performance**: No reflection or `any` usage
- ✅ **Maintainability**: Hand-crafted, focused component definitions
- ✅ **Business Logic**: Built-in validation and ERP-specific features
- ✅ **Multi-tenant Ready**: Built-in tenant context support
- ✅ **Extensible**: Easy to add custom components and validation

## Component Types

### Form Components
- **Input**: Text, email, password, number inputs
- **Textarea**: Multi-line text input
- **Select**: Dropdown with static or dynamic options
- **Checkbox**: Boolean input with label
- **Radio**: Radio button groups
- **Button**: Action buttons with variants and icons
- **DatePicker**: Date/time selection with multiple modes
- **FileUpload**: File upload with drag-drop and preview
- **Form**: Container with validation and submission

### Layout Components
- **Container**: Responsive content container
- **Card**: Flexible content card with header/footer
- **Panel**: Collapsible panel component
- **Tabs**: Tabbed interface with multiple layouts
- **Modal**: Overlay dialog component
- **Drawer**: Slide-out side panel

### Data Display Components
- **Table**: Advanced data table with sorting, filtering, pagination
- **List**: Simple list component with pagination
- **Tree**: Hierarchical tree with selection and drag-drop
- **Chart**: Data visualization with multiple chart types
- **Badge**: Status indicator and count display
- **Tag**: Labeling and categorization component

## Quick Start

### Basic Usage

```go
import "github.com/niiniyare/erp/pkg/schema/ui"

// Create registry with all built-in components
registry := ui.NewRegistry()

// Create a simple input field
emailInput := ui.CreateInput("email-input", "email").
    WithLabel("Email Address").
    WithName("email").
    Required().
    Build()

// Create a select dropdown
statusSelect := ui.CreateSelect("status-select", []ui.Option{
    {Value: "active", Label: "Active"},
    {Value: "inactive", Label: "Inactive"},
}).
    WithLabel("Status").
    WithName("status").
    Build()

// Create a submit button
submitBtn := ui.CreateButton("submit-btn", "Save Changes").
    WithVariant(ui.VariantPrimary).
    WithSize(ui.SizeLG).
    Build()
```

### Form Creation

```go
// Create a complete form
userForm := ui.NewComponent(ui.ComponentForm, "user-form").
    WithLabel("User Information").
    WithConfig(ui.FormConfig{
        Method: "POST",
        Action: "/api/users",
        Layout: ui.FormLayoutVertical,
        Validation: &ui.FormValidation{
            ValidateOnSubmit: true,
            ShowErrors:       true,
        },
    }).
    WithChildren(
        ui.CreateInput("first-name", "text").
            WithLabel("First Name").
            WithName("first_name").
            Required().
            Build(),

        ui.CreateInput("last-name", "text").
            WithLabel("Last Name").
            WithName("last_name").
            Required().
            Build(),

        ui.CreateInput("email", "email").
            WithLabel("Email").
            WithName("email").
            Required().
            WithValidator(&ui.Validator{
                Pattern: `^[^\s@]+@[^\s@]+\.[^\s@]+$`,
                Message: "Please enter a valid email address",
            }).
            Build(),

        ui.CreateButton("submit", "Create User").
            WithVariant(ui.VariantPrimary).
            WithConfig(ui.ButtonConfig{
                ButtonType: ui.ButtonSubmit,
            }).
            Build(),
    ).
    Build()
```

### Data Table Creation

```go
// Create a data table with advanced features
usersTable := ui.CreateTable("users-table", []ui.TableColumn{
    {
        Key:        "id",
        Title:      "ID",
        DataType:   ui.DataTypeNumber,
        Width:      "80px",
        Sortable:   true,
    },
    {
        Key:        "name",
        Title:      "Full Name",
        DataType:   ui.DataTypeText,
        Sortable:   true,
        Filterable: true,
        Searchable: true,
    },
    {
        Key:      "email",
        Title:    "Email",
        DataType: ui.DataTypeText,
    },
    {
        Key:      "status",
        Title:    "Status",
        DataType: ui.DataTypeBadge,
        Align:    ui.AlignCenter,
        Filterable: true,
    },
    {
        Key:      "actions",
        Title:    "Actions",
        DataType: ui.DataTypeAction,
        Width:    "120px",
    },
}).
    WithLabel("Users Management").
    WithConfig(ui.TableConfig{
        DataSource: "/api/users",
        Pagination: &ui.Pagination{
            PageSize:        25,
            ShowSizer:       true,
            ShowQuickJumper: true,
        },
        Selection: &ui.Selection{
            Type: ui.SelectionCheckbox,
        },
        Actions: []ui.Action{
            {
                Key:     "create",
                Label:   "Add User",
                Icon:    "plus",
                Variant: ui.VariantPrimary,
                OnClick: "openCreateModal()",
            },
        },
        RowActions: []ui.Action{
            {
                Key:     "edit",
                Label:   "Edit",
                Icon:    "edit",
                OnClick: "editUser(row.id)",
            },
            {
                Key:     "delete",
                Label:   "Delete",
                Icon:    "trash",
                Variant: ui.VariantDanger,
                Confirm: &ui.ConfirmDialog{
                    Title:       "Delete User",
                    Description: "Are you sure you want to delete this user?",
                    OkText:      "Delete",
                    CancelText:  "Cancel",
                },
                OnClick: "deleteUser(row.id)",
            },
        },
    }).
    Build()
```

## Validation

### Component Validation

```go
ctx := context.Background()

// Validate a single component
errors := ui.ValidateComponent(ctx, registry, userForm)
if len(errors) > 0 {
    for _, err := range errors {
        fmt.Printf("Validation error: %s\n", err)
    }
}

// Factory validation
if err := registry.Validate(ctx, emailInput); err != nil {
    fmt.Printf("Factory validation failed: %v\n", err)
}
```

### Custom Validation Rules

```go
validator := &ui.Validator{
    Required:  true,
    MinLength: &[]int{3}[0],
    MaxLength: &[]int{50}[0],
    Pattern:   `^[a-zA-Z\s]+$`,
    Message:   "Name must be 3-50 characters and contain only letters",
}

component := ui.CreateInput("name", "text").
    WithValidator(validator).
    Build()
```

## Component Tree Operations

```go
// Walk through all components
ui.WalkComponents(userForm, func(c ui.Component) error {
    fmt.Printf("Component: %s (type: %s)\n", c.ID, c.Type)
    return nil
})

// Find component by ID
emailField := ui.FindComponentByID(userForm, "email")
if emailField != nil {
    fmt.Printf("Found email field: %s\n", emailField.Label)
}

// Find all components of a specific type
allButtons := ui.FindComponentsByType(userForm, ui.ComponentButton)
fmt.Printf("Found %d buttons\n", len(allButtons))
```

## Multi-Tenant Support

```go
import "github.com/google/uuid"

tenantID := uuid.New()

component := ui.CreateInput("tenant-field", "text").
    WithTenantID(tenantID).
    WithLabel("Tenant-Specific Field").
    Build()
```

## Registry Management

```go
// Get all registered component types
types := registry.GetTypes()
fmt.Printf("Registered types: %v\n", types)

// Get schema for a component type
schema, err := registry.GetSchema(ui.ComponentInput)
if err == nil {
    fmt.Printf("Input schema: %+v\n", schema)
}

// Get all schemas
allSchemas := registry.GetAllSchemas()
for componentType, schema := range allSchemas {
    fmt.Printf("%s: %s\n", componentType, schema.Description)
}
```

## Custom Components

```go
// Define custom component type
const ComponentCustomWidget ComponentType = "custom-widget"

// Create custom config
type CustomWidgetConfig struct {
    Title   string `json:"title"`
    Content string `json:"content"`
}

// Implement factory
type CustomWidgetFactory struct{}

func (f *CustomWidgetFactory) Create(ctx context.Context, config map[string]interface{}) (Component, error) {
    var customConfig CustomWidgetConfig
    if err := mapToStruct(config, &customConfig); err != nil {
        return Component{}, fmt.Errorf("invalid custom widget config: %w", err)
    }
    
    component := NewComponent(ComponentCustomWidget, generateID())
    return component.WithConfig(customConfig).Build(), nil
}

func (f *CustomWidgetFactory) Validate(ctx context.Context, component Component) error {
    return nil
}

func (f *CustomWidgetFactory) GetSchema() ComponentSchema {
    return ComponentSchema{
        Type:        ComponentCustomWidget,
        Title:       "Custom Widget",
        Description: "A custom widget component",
    }
}

// Register with registry
registry.Register(ComponentCustomWidget, &CustomWidgetFactory{})
```

## Error Handling

```go
// Component creation with error handling
component, err := registry.Create(ctx, ui.ComponentInput, map[string]interface{}{
    "input_type": "email",
    "required":   true,
})
if err != nil {
    log.Printf("Failed to create component: %v", err)
    return
}

// Validation error handling
if errors := ui.ValidateComponent(ctx, registry, component); len(errors) > 0 {
    for _, err := range errors {
        log.Printf("Validation error: %s", err)
    }
}
```

## Integration with ERP System

This schema system integrates seamlessly with:

- **Multi-tenant architecture**: Built-in tenant context
- **ABAC security**: Component-level access control ready
- **Audit system**: Component creation/modification tracking
- **Templ templates**: Direct integration for server-side rendering
- **HTMX**: Server-driven UI updates
- **Validation system**: Server-side validation integration

## Performance Considerations

- **Zero allocation** component creation for hot paths
- **Type-safe** operations eliminate runtime reflection
- **Efficient validation** with compiled rules
- **Memory efficient** with pointer-based optional fields
- **Concurrent safe** registry operations

## Testing

```bash
go test ./pkg/schema/ui/... -v
```

The package includes comprehensive tests covering:
- Component creation and validation
- Form building and nested components  
- Table configuration with complex features
- Component tree operations
- Error handling and edge cases

## License

Part of the ERP system - internal use only.
