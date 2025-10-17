# Component Lifecycle

**FILE PURPOSE**: Complete guide to component creation, validation, and rendering  
**SCOPE**: From schema definition to HTML output  
**TARGET AUDIENCE**: Component developers, UI engineers

## 🔄 Component Lifecycle Overview

Understanding how components flow through our schema-driven system from definition to rendered HTML.

### Lifecycle Stages
```
1. Schema Definition → 2. Props Validation → 3. Component Creation → 4. Rendering → 5. HTML Output
```

## 📋 Stage 1: Schema Definition

### Creating Component Schema
```json
// components/atoms/CustomButtonSchema.json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "./CustomButtonSchema.json",
  "type": "object",
  "properties": {
    "type": {
      "type": "string",
      "const": "custom-button"
    },
    "text": {
      "type": "string",
      "description": "Button text content",
      "minLength": 1,
      "maxLength": 50
    },
    "variant": {
      "type": "string",
      "enum": ["primary", "secondary", "success", "danger", "warning"],
      "default": "primary"
    },
    "size": {
      "type": "string",
      "enum": ["xs", "sm", "md", "lg", "xl"],
      "default": "md"
    },
    "disabled": {
      "type": "boolean",
      "default": false
    },
    "loading": {
      "type": "boolean", 
      "default": false
    },
    "icon": {
      "type": "string",
      "description": "Icon name from Flowbite icons"
    },
    "onClick": {
      "type": "string",
      "description": "JavaScript or HTMX action"
    }
  },
  "required": ["type", "text"],
  "additionalProperties": false
}
```

### Schema Registration
```go
// Register schema with factory
factory.componentFactory["CustomButtonSchema"] = &CustomButtonRenderer{}
```

## ✅ Stage 2: Props Validation

### Input Validation Process
```go
func (factory *SchemaFactory) ValidateProps(schemaType string, props map[string]interface{}) error {
    // 1. Get schema definition
    schema, exists := factory.schemaRegistry[schemaType]
    if !exists {
        return fmt.Errorf("schema not found: %s", schemaType)
    }
    
    // 2. Validate required properties
    for _, required := range schema.Required {
        if _, exists := props[required]; !exists {
            return fmt.Errorf("required property missing: %s", required)
        }
    }
    
    // 3. Validate property types and values
    for key, value := range props {
        if property, exists := schema.Properties[key]; exists {
            if err := factory.validateProperty(key, value, property); err != nil {
                return fmt.Errorf("property %s: %w", key, err)
            }
        }
    }
    
    return nil
}
```

### Validation Rules
- **Type Checking**: String, boolean, number validation
- **Enum Validation**: Value must be in allowed list
- **Length Validation**: String min/max length
- **Pattern Validation**: Regex pattern matching
- **Custom Validation**: Business-specific rules

## 🏗️ Stage 3: Component Creation

### Component Renderer Implementation
```go
type CustomButtonRenderer struct{}

func (r *CustomButtonRenderer) GetSchemaType() string {
    return "CustomButtonSchema"
}

func (r *CustomButtonRenderer) Render(ctx context.Context, props map[string]interface{}, schema *JsonSchema) (TemplComponent, error) {
    // Extract and type-convert properties
    buttonProps := atoms.ButtonProps{
        Text:     getStringProp(props, "text", ""),
        Type:     getStringProp(props, "type", "button"),
        Variant:  atoms.ButtonVariant(getStringProp(props, "variant", "primary")),
        Size:     atoms.ButtonSize(getStringProp(props, "size", "md")),
        Disabled: getBoolProp(props, "disabled", false),
        Loading:  getBoolProp(props, "loading", false),
        Icon:     getStringProp(props, "icon", ""),
        ID:       getStringProp(props, "id", ""),
        Class:    getStringProp(props, "className", ""),
        OnClick:  getStringProp(props, "onClick", ""),
    }
    
    // Create component structure
    return TemplComponent{
        Type:  "CustomButton",
        Props: buttonProps,
        Attributes: map[string]string{
            "type": buttonProps.Type,
            "id":   buttonProps.ID,
            "data-variant": string(buttonProps.Variant),
        },
        Events: map[string]string{
            "click": buttonProps.OnClick,
        },
        CSS: generateButtonCSS(buttonProps),
    }, nil
}
```

### Property Extraction Helpers
```go
func getStringProp(props map[string]interface{}, key string, defaultValue string) string {
    if value, exists := props[key]; exists {
        if str, ok := value.(string); ok {
            return str
        }
    }
    return defaultValue
}

func getBoolProp(props map[string]interface{}, key string, defaultValue bool) bool {
    if value, exists := props[key]; exists {
        if b, ok := value.(bool); ok {
            return b
        }
    }
    return defaultValue
}
```

## 🎨 Stage 4: CSS and Styling

### CSS Generation
```go
func generateButtonCSS(props atoms.ButtonProps) string {
    classes := []string{"btn"} // Base class
    
    // Variant classes
    switch props.Variant {
    case atoms.ButtonPrimary:
        classes = append(classes, "btn-primary", "bg-blue-600", "text-white")
    case atoms.ButtonSecondary:
        classes = append(classes, "btn-secondary", "bg-gray-600", "text-white")
    case atoms.ButtonSuccess:
        classes = append(classes, "btn-success", "bg-green-600", "text-white")
    case atoms.ButtonDanger:
        classes = append(classes, "btn-danger", "bg-red-600", "text-white")
    }
    
    // Size classes
    switch props.Size {
    case atoms.ButtonSizeXS:
        classes = append(classes, "px-2", "py-1", "text-xs")
    case atoms.ButtonSizeSM:
        classes = append(classes, "px-3", "py-1.5", "text-sm")
    case atoms.ButtonSizeMD:
        classes = append(classes, "px-4", "py-2", "text-sm")
    case atoms.ButtonSizeLG:
        classes = append(classes, "px-5", "py-2.5", "text-base")
    case atoms.ButtonSizeXL:
        classes = append(classes, "px-6", "py-3", "text-base")
    }
    
    // State classes
    if props.Disabled {
        classes = append(classes, "opacity-50", "cursor-not-allowed")
    }
    
    if props.Loading {
        classes = append(classes, "cursor-wait")
    }
    
    // Custom classes
    if props.Class != "" {
        classes = append(classes, props.Class)
    }
    
    return strings.Join(classes, " ")
}
```

## 🖥️ Stage 5: Templ Rendering

### Templ Component Generation
```go
func (str *SchemaTemplRenderer) renderCustomButton(component TemplComponent) templ.Component {
    return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
        props := component.Props.(atoms.ButtonProps)
        
        // Generate button HTML
        html := fmt.Sprintf(`
            <button type="%s" 
                    class="%s" 
                    id="%s"
                    %s
                    %s>
                %s%s
            </button>`,
            props.Type,
            component.CSS,
            props.ID,
            renderDisabled(props.Disabled),
            renderEvents(component.Events),
            renderIcon(props.Icon),
            html.EscapeString(props.Text),
        )
        
        _, err := w.Write([]byte(html))
        return err
    })
}
```

### Helper Functions
```go
func renderDisabled(disabled bool) string {
    if disabled {
        return `disabled="disabled"`
    }
    return ""
}

func renderEvents(events map[string]string) string {
    var attrs []string
    for event, handler := range events {
        if handler != "" {
            // Support both JavaScript and HTMX
            if strings.HasPrefix(handler, "hx-") {
                attrs = append(attrs, fmt.Sprintf(`%s="%s"`, event, handler))
            } else {
                attrs = append(attrs, fmt.Sprintf(`on%s="%s"`, event, handler))
            }
        }
    }
    return strings.Join(attrs, " ")
}

func renderIcon(icon string) string {
    if icon != "" {
        return fmt.Sprintf(`<i class="icon-%s mr-2"></i>`, icon)
    }
    return ""
}
```

## 🔄 Complete Usage Example

### End-to-End Component Usage
```go
func main() {
    // Initialize factory
    factory, _ := engine.NewSchemaFactory("docs/ui/Schema")
    
    // Define component properties
    props := map[string]interface{}{
        "text":     "Save Changes",
        "variant":  "success",
        "size":     "lg",
        "disabled": false,
        "loading":  false,
        "icon":     "check",
        "onClick":  "hx-post='/api/save'",
    }
    
    // Generate component through full lifecycle
    component, err := factory.RenderToTempl(
        context.Background(), 
        "CustomButtonSchema", 
        props,
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Use in Templ template
    // @component will render the button
}
```

### Generated HTML Output
```html
<button type="button" 
        class="btn btn-success bg-green-600 text-white px-5 py-2.5 text-base" 
        id=""
        hx-post="/api/save">
    <i class="icon-check mr-2"></i>Save Changes
</button>
```

## ⚡ Performance Optimizations

### Lifecycle Optimizations
1. **Schema Caching**: Pre-load schemas at startup
2. **Validation Caching**: Cache validation results for identical props
3. **Component Memoization**: Reuse rendered components
4. **CSS Generation**: Cache CSS strings for common combinations

### Performance Monitoring
```go
// Add timing metrics to each stage
func (factory *SchemaFactory) RenderToTemplWithMetrics(ctx context.Context, schemaType string, props map[string]interface{}) (templ.Component, time.Duration, error) {
    start := time.Now()
    
    component, err := factory.RenderToTempl(ctx, schemaType, props)
    
    duration := time.Since(start)
    
    // Log performance metrics
    if duration > 10*time.Millisecond {
        log.Printf("Slow component render: %s took %v", schemaType, duration)
    }
    
    return component, duration, err
}
```

## 🐛 Error Handling

### Common Lifecycle Errors
1. **Schema Not Found**: Invalid schema name
2. **Validation Errors**: Props don't match schema
3. **Rendering Errors**: Component generation fails
4. **CSS Errors**: Invalid styling properties

### Error Recovery
```go
func (factory *SchemaFactory) RenderWithFallback(ctx context.Context, schemaType string, props map[string]interface{}) templ.Component {
    component, err := factory.RenderToTempl(ctx, schemaType, props)
    if err != nil {
        // Log error and return fallback component
        log.Printf("Component render failed: %v", err)
        return factory.renderErrorComponent(schemaType, err)
    }
    return component
}

func (factory *SchemaFactory) renderErrorComponent(schemaType string, err error) templ.Component {
    return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
        html := fmt.Sprintf(`
            <div class="error-component p-4 bg-red-100 border border-red-400 text-red-700 rounded">
                <strong>Component Error:</strong> %s<br>
                <small>Schema: %s</small>
            </div>`,
            html.EscapeString(err.Error()),
            html.EscapeString(schemaType),
        )
        _, writeErr := w.Write([]byte(html))
        return writeErr
    })
}
```

## 📚 Related Documentation

- **[Schema System](schema-system.md)**: Complete schema documentation
- **[Creating Components](../development/creating-components.md)**: Development guide
- **[Validation Patterns](../development/validation-patterns.md)**: Testing strategies
- **[Architecture](architecture.md)**: System design principles

The component lifecycle provides a robust, validated path from schema definition to rendered HTML, ensuring type safety and consistency across our entire UI system.