# Basic Examples

**FILE PURPOSE**: Working examples of common UI patterns  
**PREREQUISITES**: [First Component](first-component.md) completed  
**OUTCOME**: Understanding of schema-driven development patterns

## 🎯 Three Essential Patterns

### Example 1: Form with Validation

Complete form using schema-driven components:

```go
// cmd/examples/form-example/main.go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/niiniyare/erp/web/engine"
)

func main() {
    factory, err := engine.NewSchemaFactory("docs/ui/Schema")
    if err != nil {
        log.Fatal(err)
    }
    
    // Text input component
    inputProps := map[string]interface{}{
        "type":        "text",
        "name":        "username",
        "label":       "Username",
        "placeholder": "Enter your username",
        "required":    true,
        "validation": map[string]interface{}{
            "minLength": 3,
            "maxLength": 20,
        },
    }
    
    // Checkbox component
    checkboxProps := map[string]interface{}{
        "name":    "terms",
        "label":   "I agree to the terms",
        "required": true,
    }
    
    // Submit button
    buttonProps := map[string]interface{}{
        "text":    "Submit",
        "type":    "submit",
        "variant": "primary",
    }
    
    // Generate components
    input, _ := factory.RenderToTempl(context.Background(), "TextControlSchema", inputProps)
    checkbox, _ := factory.RenderToTempl(context.Background(), "CheckboxControlSchema", checkboxProps)
    button, _ := factory.RenderToTempl(context.Background(), "ButtonGroupSchema", buttonProps)
    
    fmt.Println("✅ Form components created:")
    fmt.Printf("- Input: %T\n", input)
    fmt.Printf("- Checkbox: %T\n", checkbox)  
    fmt.Printf("- Button: %T\n", button)
}
```

**Template Usage**:
```go
templ LoginForm() {
    <form hx-post="/api/login" hx-target="#result" class="space-y-4">
        @input   // Schema-generated text input
        @checkbox // Schema-generated checkbox
        @button   // Schema-generated submit button
    </form>
}
```

### Example 2: Data Table with Actions

Schema-driven table with interactive features:

```go
// Table component with action buttons
func createDataTable() {
    factory, _ := engine.NewSchemaFactory("docs/ui/Schema")
    
    tableProps := map[string]interface{}{
        "columns": []map[string]interface{}{
            {"key": "name", "title": "Name", "sortable": true},
            {"key": "email", "title": "Email", "sortable": true},
            {"key": "status", "title": "Status", "type": "badge"},
            {"key": "actions", "title": "Actions", "type": "buttons"},
        },
        "data": []map[string]interface{}{
            {
                "name": "John Doe",
                "email": "john@example.com", 
                "status": "active",
                "actions": []map[string]interface{}{
                    {"text": "Edit", "variant": "secondary", "hx-get": "/users/1/edit"},
                    {"text": "Delete", "variant": "danger", "hx-delete": "/users/1"},
                },
            },
        },
        "pagination": true,
        "search": true,
    }
    
    table, _ := factory.RenderToTempl(context.Background(), "TableSchema", tableProps)
    return table
}
```

### Example 3: Card Layout with Navigation

Responsive card grid with navigation:

```go
// Dashboard cards
func createDashboard() {
    factory, _ := engine.NewSchemaFactory("docs/ui/Schema")
    
    // Individual card
    cardProps := map[string]interface{}{
        "title": "Financial Overview",
        "content": "Monthly revenue and expenses",
        "actions": []map[string]interface{}{
            {"text": "View Details", "variant": "primary", "href": "/finance"},
            {"text": "Export", "variant": "secondary", "hx-get": "/finance/export"},
        },
        "image": "/assets/finance-chart.png",
        "badge": map[string]interface{}{
            "text": "Updated",
            "color": "green",
        },
    }
    
    // Container for responsive grid
    containerProps := map[string]interface{}{
        "layout": "grid",
        "columns": map[string]interface{}{
            "sm": 1,
            "md": 2, 
            "lg": 3,
        },
        "gap": "6",
        "padding": "4",
    }
    
    card, _ := factory.RenderToTempl(context.Background(), "CardSchema", cardProps)
    container, _ := factory.RenderToTempl(context.Background(), "ContainerSchema", containerProps)
    
    return container, card
}
```

## 🔧 Running the Examples

### Quick Test Script
```bash
# Create and run examples
mkdir -p cmd/examples
cd cmd/examples

# Copy the Go code above into files
# Run each example:
go run form-example/main.go
go run table-example/main.go  
go run dashboard-example/main.go
```

### Integration with Web Server
```go
// In your web server handler
func dashboardHandler(c *fiber.Ctx) error {
    container, card := createDashboard()
    
    return render(c, DashboardPage(container, card))
}

templ DashboardPage(container, card templ.Component) {
    @layouts.Base("Dashboard", "ERP Dashboard") {
        @container {
            @card
            @card  // Repeat cards as needed
        }
    }
}
```

## 🎨 Styling and Customization

### Theme Customization
```go
// Custom theme props
themeProps := map[string]interface{}{
    "primaryColor": "#3b82f6",    // Blue
    "secondaryColor": "#6b7280",  // Gray
    "successColor": "#10b981",    // Green
    "dangerColor": "#ef4444",     // Red
    "borderRadius": "8px",
    "fontFamily": "Inter, sans-serif",
}
```

### Responsive Breakpoints
```go
// Responsive component props
responsiveProps := map[string]interface{}{
    "className": "w-full sm:w-1/2 lg:w-1/3",
    "responsive": map[string]interface{}{
        "mobile": {"padding": "2", "fontSize": "sm"},
        "tablet": {"padding": "4", "fontSize": "base"},
        "desktop": {"padding": "6", "fontSize": "lg"},
    },
}
```

## 🚀 Next Steps

### Immediate Actions
1. **Run all three examples** to understand patterns
2. **Modify props** to see how components change  
3. **Create your own combination** of components

### Deep Dive Options
- 📚 [Component Development Guide](../development/creating-components.md)
- 🏗️ [Schema System Architecture](../fundamentals/schema-system.md)
- 🔗 [HTMX Integration Patterns](../integration/htmx.md)
- 🎨 [Styling and Themes](../fundamentals/styling-approach.md)

### Advanced Topics
- **Custom Schema Creation**: [Schema Definitions Guide](../development/schema-definitions.md)
- **Validation Patterns**: [Validation Guide](../development/validation-patterns.md)
- **Testing Strategies**: [Testing Guide](../development/testing-strategies.md)

## ✅ Quick Start Complete!

You now understand:
- ✅ Schema-driven component creation
- ✅ Common UI patterns (forms, tables, cards)
- ✅ Integration with HTMX and Alpine.js
- ✅ Styling and responsive design

**Quick Start Path**:
1. ✅ [Installation](installation.md)
2. ✅ [First Component](first-component.md)  
3. ✅ Basic Examples (you are here)

**Ready for Production**: You now have the foundation to build sophisticated UI components using our schema-driven architecture!