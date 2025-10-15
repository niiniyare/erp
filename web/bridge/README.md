# ERP UI Bridge System

A seamless bridge connecting the advanced CSS Runtime System (`@pkg/schema/ui/`) with the mature Templ component architecture (`@web/components/`), enabling unified schema-driven UI development.

## 🌉 Overview

The Bridge System provides:
- **Bidirectional Conversion** between schema components and Templ components
- **Unified Registry** supporting both development approaches
- **CSS Runtime Integration** with automatic styling preservation
- **Template-Based Creation** for common UI patterns
- **CLI Tooling** for development workflow automation

## 🚀 Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "github.com/niiniyare/erp/web/bridge"
)

func main() {
    ctx := context.Background()
    
    // Create bridge instance
    bridge := bridge.NewBridge()
    
    // Create unified registry
    registry := bridge.NewUnifiedRegistry()
    
    // Run demonstration
    demo := bridge.NewDemo()
    demo.RunFullDemo(ctx)
}
```

### Schema to Templ Conversion

```go
// Create schema component
schemaComponent := schemaui.NewComponent(schemaui.ComponentButton, "save-btn").
    WithLabel("Save Changes").
    WithVariant(schemaui.VariantPrimary).
    WithConfig(schemaui.ButtonConfig{
        Text: "Save Changes",
        Icon: "save",
        ButtonType: schemaui.ButtonSubmit,
    }).
    Build()

// Convert to Templ component
templComponent, err := bridge.ConvertSchemaToTempl(ctx, schemaComponent)
if err != nil {
    log.Fatal(err)
}

// Use in Templ template
// @atoms.Button(templComponent.Props.(atoms.ButtonProps))
```

### Templ to Schema Conversion

```go
// Create Templ component props
inputProps := atoms.InputProps{
    Type:        "email",
    Name:        "user_email",
    Placeholder: "Enter your email",
    Required:    true,
    Size:        atoms.InputSizeLG,
}

templComponent := bridge.TemplComponent{
    Type:  "Input",
    Props: inputProps,
}

// Convert to schema component
schemaComponent, err := bridge.ConvertTemplToSchema(ctx, templComponent)
if err != nil {
    log.Fatal(err)
}
```

## 📋 Supported Components

### Form Components
- **Button** - Action buttons with variants and icons
- **Input** - Text inputs with validation
- **Textarea** - Multi-line text inputs
- **Select** - Dropdown selections with options
- **Checkbox** - Boolean checkboxes
- **Radio** - Radio button groups

### Component Features
- ✅ **Size variants** (xs, sm, md, lg, xl)
- ✅ **Color variants** (primary, secondary, success, danger, warning, info)
- ✅ **Validation rules** (required, pattern, length limits)
- ✅ **Accessibility** (ARIA labels, keyboard navigation)
- ✅ **HTMX integration** (server-driven interactions)
- ✅ **Alpine.js integration** (client-side reactivity)

## 🛠 CLI Tools

The bridge includes a comprehensive CLI for development workflow automation:

### Installation

```bash
# Build the CLI tool
go build -o bridge ./web/bridge/cmd/
```

### Commands

```bash
# Convert schema JSON to Templ component
bridge schema-to-templ button.json button.templ

# Convert Templ props to schema JSON  
bridge templ-to-schema Button '{"text":"Click me"}' button-schema.json

# Create template in both formats
bridge create-template login-form ./output

# Validate components
bridge validate schema component.json
bridge validate templ props.json

# List supported types
bridge list-types

# List available templates
bridge list-templates
```

### Available Templates

- **`login-form`** - Complete authentication form (email, password, submit)
- **`user-form`** - User registration form (name, email, role selection)
- **`data-table`** - Data table interface (search, filters, actions)

## 🏗 Architecture

### Bridge Components

```
┌─────────────────┐    ┌─────────────────┐
│  Schema System  │◄──►│  Bridge System  │◄──►│  Templ System   │
│  (@pkg/schema)  │    │  (@web/bridge)  │    │ (@web/components)│
└─────────────────┘    └─────────────────┘    └─────────────────┘
│                                                                 │
│ • 913 JSON Schemas      • Bidirectional Conversion            │
│ • CSS Runtime           • Unified Registry                     │
│ • Military Validation   • Template Creation                    │
│ • Auto-Generation       • CLI Tooling                         │
│                                                                 │
│ • 35+ Templ Components                                         │
│ • HTMX Integration                                             │
│ • Alpine.js Reactivity                                        │
│ • Flowbite Design                                             │
└─────────────────────────────────────────────────────────────────┘
```

### Conversion Flow

```
Schema Component → Bridge Converter → Templ Component → Template Rendering
     ↓                    ↓                 ↓                ↓
 JSON Config     →    Props Mapping   →   Go Struct    →   HTML Output
 CSS Styles      →    Class Mapping   →   Templ Props  →   Styled HTML
 Validation      →    Rule Conversion →   Client Logic →   Form Behavior
```

## 🔧 Development Workflows

### Schema-First Workflow

```go
// 1. Define schema
schema := schemaui.NewComponent(schemaui.ComponentInput, "email-field").
    WithLabel("Email Address").
    WithConfig(schemaui.InputConfig{
        InputType: schemaui.InputEmail,
        Pattern: `^[^\s@]+@[^\s@]+\.[^\s@]+$`,
    }).
    Build()

// 2. Convert to Templ
templComp, _ := bridge.ConvertSchemaToTempl(ctx, schema)

// 3. Use in template
// @atoms.Input(templComp.Props.(atoms.InputProps))
```

### Component-First Workflow

```go
// 1. Create Templ component
props := atoms.ButtonProps{
    Text: "Save User",
    Variant: atoms.ButtonPrimary,
    Type: "submit",
}

// 2. Convert to schema
templComp := bridge.TemplComponent{Type: "Button", Props: props}
schema, _ := bridge.ConvertTemplToSchema(ctx, templComp)

// 3. Use in schema system
registry.ValidateSchemaComponent(ctx, schema)
```

### Unified Workflow

```go
// Create components using unified registry
registry := bridge.NewUnifiedRegistry()

// Method 1: Schema creation
schemaComp, _ := registry.CreateSchemaComponent(ctx, schemaui.ComponentButton, config)

// Method 2: Templ creation  
templComp, _ := registry.CreateTemplComponent(ctx, schemaui.ComponentButton, config)

// Method 3: Template creation
templComps, schemaComps, _ := registry.CreateFromTemplate(ctx, "login-form", data)
```

## 🎨 CSS Runtime Integration

The bridge automatically integrates the CSS Runtime System with Templ components:

```go
// Schema component with styles
styledComponent := schemaui.NewComponent(schemaui.ComponentButton, "custom-btn").
    WithStyles(&css.Styles{
        BackgroundColor: css.ColorPrimary,
        Padding: css.SpacingLG,
        BorderRadius: css.BorderRadiusLG,
    }).
    Build()

// Convert to Templ (CSS classes automatically generated)
templComponent, _ := bridge.ConvertSchemaToTempl(ctx, styledComponent)

// Templ component now includes generated CSS classes
buttonProps := templComponent.Props.(atoms.ButtonProps)
// buttonProps.Class contains generated CSS classes
```

## 🧪 Testing

### Run Demonstrations

```go
// Full system demonstration
demo := bridge.NewDemo()
demo.RunFullDemo(context.Background())
```

### Test Conversion

```go
func TestButtonConversion(t *testing.T) {
    bridge := bridge.NewBridge()
    ctx := context.Background()
    
    // Create schema button
    schema := schemaui.NewComponent(schemaui.ComponentButton, "test-btn").
        WithLabel("Test Button").
        Build()
    
    // Convert to Templ
    templ, err := bridge.ConvertSchemaToTempl(ctx, schema)
    assert.NoError(t, err)
    assert.Equal(t, "Button", templ.Type)
    
    // Convert back to schema
    converted, err := bridge.ConvertTemplToSchema(ctx, templ)
    assert.NoError(t, err)
    assert.Equal(t, schemaui.ComponentButton, converted.Type)
}
```

## 📚 Examples

### Complete Login Form

```go
func createLoginForm(ctx context.Context) error {
    registry := bridge.NewUnifiedRegistry()
    
    // Create login form template
    templComponents, schemaComponents, err := registry.CreateFromTemplate(
        ctx, 
        "login-form", 
        map[string]any{},
    )
    if err != nil {
        return err
    }
    
    // Generate Templ template code
    cli := bridge.NewCLI()
    templCode, err := cli.generateTemplCodeFromComponents(templComponents, "LoginForm")
    if err != nil {
        return err
    }
    
    // Write to file
    return os.WriteFile("login_form.templ", []byte(templCode), 0644)
}
```

### Custom Component Bridge

```go
func bridgeCustomComponent(ctx context.Context) error {
    // Create custom Templ component
    customProps := atoms.ButtonProps{
        Text: "Custom Action",
        Icon: "gear",
        Variant: atoms.ButtonSecondary,
        Size: atoms.ButtonSizeLG,
        OnClick: "handleCustomAction()",
    }
    
    templComponent := bridge.TemplComponent{
        Type: "Button", 
        Props: customProps,
    }
    
    // Convert to schema for validation
    bridge := bridge.NewBridge()
    schemaComponent, err := bridge.ConvertTemplToSchema(ctx, templComponent)
    if err != nil {
        return err
    }
    
    // Validate using schema system
    registry := bridge.NewUnifiedRegistry()
    return registry.ValidateSchemaComponent(ctx, schemaComponent)
}
```

## 🔗 Integration Points

### With Existing Systems

```go
// Integration with @pkg/schema/ui CSS Runtime
cssFactory := css.NewFactory("docs/ui/Schema")
bridge := bridge.NewBridge() // Automatically integrates CSS factory

// Integration with @web/components Templ templates
templComponent, _ := bridge.ConvertSchemaToTempl(ctx, schemaComponent)
// Use in any existing .templ file:
// @atoms.Button(templComponent.Props.(atoms.ButtonProps))

// Integration with PatternRenderer system
pattern := &schemas.PatternDefinition{
    Type: "form",
    Components: schemaComponents,
}
renderer := engine.NewPatternRenderer()
html, _ := renderer.RenderPattern(ctx, pattern, data)
```

### With Visual Builder

```go
// Visual builder can now work with both systems
builder := visualbuilder.New()

// Add schema components to palette
for _, compType := range registry.GetSupportedTypes() {
    builder.AddToPalette(compType, registry.GetSchemaForType(compType))
}

// Add Templ components to palette  
templRegistry := bridge.NewTemplRegistry()
builder.AddTemplComponentsToPalette(templRegistry.GetFactories())

// Live preview supports both
preview := builder.CreatePreview()
preview.SetBridgeConverter(bridge) // Enables real-time conversion
```

## 🚀 Future Enhancements

### Phase 2 Roadmap
- [ ] **Advanced Component Mapping** - Support for organisms and molecules
- [ ] **Live Development Server** - Hot reload with bridge conversion
- [ ] **Visual Builder Integration** - Drag-and-drop with both systems
- [ ] **Advanced CSS Features** - Theme variables, responsive design
- [ ] **Performance Optimization** - Caching, lazy loading

### Phase 3 Roadmap  
- [ ] **Multi-tenant Bridge** - Tenant-specific component variations
- [ ] **Security Integration** - ABAC-aware component rendering
- [ ] **Audit Trail** - Track component creation and modification
- [ ] **Migration Tools** - Automated migration between systems

## 📖 API Reference

### Core Interfaces

```go
type Bridge interface {
    ConvertSchemaToTempl(ctx context.Context, schema schemaui.Component) (TemplComponent, error)
    ConvertTemplToSchema(ctx context.Context, templ TemplComponent) (schemaui.Component, error)
}

type UnifiedRegistry interface {
    CreateSchemaComponent(ctx context.Context, componentType schemaui.ComponentType, config map[string]any) (schemaui.Component, error)
    CreateTemplComponent(ctx context.Context, componentType schemaui.ComponentType, config map[string]any) (TemplComponent, error)
    CreateFromTemplate(ctx context.Context, templateName string, data map[string]any) ([]TemplComponent, []schemaui.Component, error)
}

type TemplComponent struct {
    Type  string      `json:"type"`
    Props interface{} `json:"props"`
}
```

---

**The Bridge System enables seamless integration between two sophisticated UI architectures, providing developers with flexible workflows while maintaining the strengths of both systems.**