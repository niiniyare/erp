# Schema System Overview

**FILE PURPOSE**: Complete guide to our JSON schema-driven component system  
**SCOPE**: Schema organization, validation, component generation  
**TARGET AUDIENCE**: Component developers, schema authors

## 🎯 Schema System Architecture

Our schema system transforms **937 JSON schemas** into **type-safe Go components** with automated validation and styling.

### System Overview
```
JSON Schemas (937 files) → Schema Factory → Go Types → Templ Components → HTML
```

**Key Statistics**:
- **Core Schemas**: 673 CSS property definitions
- **Component Schemas**: 141 UI component definitions  
- **Interaction Schemas**: 60 behavioral patterns
- **Utility Schemas**: 61 helper definitions
- **Success Rate**: 99.7% schema processing

## 📁 Schema Organization

### Directory Structure
```
docs/ui/Schema/definitions/
├── core/                    # 673 schemas (71.8%)
│   ├── layout/             # Position, display, flexbox
│   ├── typography/         # Fonts, text properties
│   ├── color/              # Colors, backgrounds
│   ├── spacing/            # Margins, padding, gaps
│   ├── datatypes/          # Type definitions
│   └── compatibility/      # Browser-specific CSS
├── components/             # 141 schemas (15.0%)
│   ├── atoms/              # Basic elements (27 schemas)
│   ├── molecules/          # Composite components (48 schemas)
│   ├── organisms/          # Complex components (53 schemas)
│   └── templates/          # Page layouts (12 schemas)
├── interactions/           # 60 schemas (6.4%)
│   ├── forms/              # Form behaviors
│   ├── navigation/         # Navigation patterns
│   └── data/               # Data interactions
└── utility/               # 61 schemas (6.5%)
    └── helpers/            # Configuration schemas
```

## 🏗️ Component Schema Structure

### Standard Schema Format
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "./ComponentSchema.json",
  "type": "object",
  "properties": {
    "type": {
      "type": "string",
      "const": "component-type"
    },
    "text": {
      "type": "string",
      "description": "Component text content"
    },
    "variant": {
      "type": "string",
      "enum": ["primary", "secondary", "success", "danger"],
      "description": "Visual variant"
    },
    "size": {
      "type": "string", 
      "enum": ["sm", "md", "lg", "xl"],
      "description": "Component size"
    },
    "disabled": {
      "type": "boolean",
      "description": "Whether component is disabled"
    }
  },
  "required": ["type"],
  "additionalProperties": false
}
```

### Schema Metadata
Each schema includes comprehensive metadata:

```json
{
  "meta": {
    "title": "Component Name",
    "description": "Component purpose and usage",
    "category": "atoms|molecules|organisms|templates",
    "subcategory": "forms|layout|navigation|data",
    "dependencies": ["RequiredSchema1", "RequiredSchema2"],
    "examples": ["path/to/example1.json"],
    "accessibility": {
      "wcagLevel": "AA",
      "keyboardSupport": true,
      "screenReaderSupport": true
    },
    "performance": {
      "renderCost": "low|medium|high",
      "optimizations": ["memoization", "lazy-loading"]
    },
    "version": "1.0.0",
    "tags": ["button", "form", "interactive"]
  }
}
```

## ⚙️ Schema Factory Implementation

### Core Factory Pattern
```go
type SchemaFactory struct {
    schemaRegistry   map[string]*JsonSchema      // Loaded schemas
    componentFactory map[string]ComponentRenderer // Renderers
    cssFactory       *css.Factory                // CSS generation
    validationEngine *ValidationEngine           // Validation rules
    templRenderer    *SchemaTemplRenderer        // Templ generation
}
```

### Component Generation Process
```go
// 1. Load schema definition
schema, exists := factory.schemaRegistry["ButtonSchema"]

// 2. Validate props against schema
err := renderer.ValidateProps(props, schema)

// 3. Generate component
component, err := renderer.Render(ctx, props, schema)

// 4. Apply CSS styling
err := factory.applyCSSToComponent(&component, props)

// 5. Convert to Templ component
templComponent := factory.templRenderer.RenderComponent(component)
```

## 🔍 Schema Discovery

### Finding Schemas
```bash
# Search by category
find docs/ui/Schema/definitions/components -name "*.json"

# Search by name pattern
find docs/ui/Schema/definitions -name "*button*"

# Get schema statistics
go run docs/ui/Schema/definitions/schema_discovery.go
```

### Schema Registry
```json
{
  "schemas": [
    {
      "name": "ButtonGroupSchema",
      "path": "components/atoms/ButtonGroupSchema.json",
      "category": "atoms",
      "dependencies": ["BadgeObject", "SchemaClassName"],
      "lastModified": "2025-10-17T18:06:31Z"
    }
  ],
  "categories": {
    "atoms": 27,
    "molecules": 48,
    "organisms": 53
  },
  "totalSchemas": 937
}
```

## ✅ Schema Validation

### Validation Layers
1. **JSON Schema Validation**: Structure and type checking
2. **Business Rule Validation**: Component-specific rules
3. **CSS Property Validation**: Valid CSS values
4. **Accessibility Validation**: WCAG compliance

### Validation Example
```go
// Define validation rules
rules := ValidationRules{
    Required: []string{"type", "text"},
    Patterns: map[string]string{
        "variant": "^(primary|secondary|success|danger)$",
    },
    Custom: map[string]func(interface{}) error{
        "size": validateSizeValue,
    },
}

// Validate props
err := factory.ValidateAgainstSchema("ButtonSchema", props)
```

## 🎨 CSS Integration

### CSS Property Schemas
Our system includes **800+ CSS property schemas** for type-safe styling:

```go
// CSS properties are validated against schemas
cssProps := &css.Styles{
    Color:           "#3b82f6",     // Validated against color schema
    BackgroundColor: "blue-500",   // TailwindCSS class validation
    FontSize:        "16px",       // Size validation
    BorderRadius:    "8px",        // Border schema validation
}

cssString, err := factory.cssFactory.GenerateCSS(*cssProps)
```

### Theme System
```go
// Theme configuration through schemas
themeProps := map[string]interface{}{
    "primary": "#3b82f6",
    "secondary": "#6b7280", 
    "success": "#10b981",
    "danger": "#ef4444",
}
```

## 🔧 Creating Custom Schemas

### Schema Development Process
1. **Define Structure**: Create JSON schema file
2. **Add Metadata**: Include comprehensive metadata
3. **Create Renderer**: Implement ComponentRenderer interface
4. **Add Validation**: Define validation rules
5. **Test Integration**: Validate with schema factory

### Custom Component Renderer
```go
type CustomRenderer struct{}

func (r *CustomRenderer) GetSchemaType() string {
    return "CustomComponentSchema"
}

func (r *CustomRenderer) ValidateProps(props map[string]interface{}, schema *JsonSchema) error {
    // Custom validation logic
    return nil
}

func (r *CustomRenderer) Render(ctx context.Context, props map[string]interface{}, schema *JsonSchema) (TemplComponent, error) {
    // Custom rendering logic
    return TemplComponent{
        Type:  "CustomComponent",
        Props: props,
    }, nil
}
```

## 📊 Schema Performance

### Optimization Strategies
- **Schema Caching**: Pre-load frequently used schemas
- **Lazy Loading**: Load schemas on demand
- **Validation Caching**: Cache validation results
- **Component Memoization**: Reduce re-rendering

### Performance Metrics
- **Schema Loading**: <5ms per schema
- **Validation Time**: <2ms per component
- **Component Generation**: <10ms per component
- **Memory Usage**: <100MB for full schema registry

## 🧪 Testing Schemas

### Validation Testing
```bash
# Test all schemas
docs/ui/Schema/check_doc_accuracy.sh

# Test specific schema
go run test-schema.go ButtonGroupSchema
```

### Integration Testing
```go
func TestSchemaIntegration(t *testing.T) {
    factory, err := engine.NewSchemaFactory("docs/ui/Schema")
    require.NoError(t, err)
    
    props := map[string]interface{}{
        "text": "Test Button",
        "variant": "primary",
    }
    
    component, err := factory.RenderToTempl(context.Background(), "ButtonGroupSchema", props)
    require.NoError(t, err)
    require.NotNil(t, component)
}
```

## 🚀 Best Practices

### Schema Design
- **Keep schemas focused**: One responsibility per schema
- **Use clear naming**: Descriptive, consistent names
- **Include metadata**: Complete documentation
- **Validate thoroughly**: Test all property combinations

### Component Development
- **Follow atomic design**: Use proper component hierarchy
- **Implement validation**: Always validate props
- **Add accessibility**: Include ARIA attributes
- **Optimize performance**: Use efficient rendering

### Maintenance
- **Version schemas**: Track breaking changes
- **Update metadata**: Keep documentation current
- **Monitor performance**: Track rendering metrics
- **Validate accuracy**: Regular schema audits

## 📚 Related Documentation

- **[Architecture](architecture.md)**: System design principles
- **[Component Development](../development/creating-components.md)**: Building components
- **[Validation Patterns](../development/validation-patterns.md)**: Testing strategies
- **[Integration Guides](../integration/)**: Technology integration

Our schema system provides the foundation for consistent, type-safe, and maintainable UI development at enterprise scale.