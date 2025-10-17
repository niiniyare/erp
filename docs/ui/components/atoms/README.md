# Atomic Components

**FILE PURPOSE**: Foundational UI elements that cannot be broken down further  
**SCOPE**: 27 atomic components with complete specifications  
**TARGET AUDIENCE**: Developers building composite components

## 🧱 Atoms Overview

Atomic components are the fundamental building blocks of our ERP interface. They represent the smallest functional units that maintain their meaning when isolated. Each atom is designed to be composable, accessible, and consistent with our design system.

### Design Principles for Atoms
- **Single responsibility**: Each atom has one clear purpose
- **Composable**: Can be combined to create molecules
- **Accessible**: WCAG 2.1 AA compliant by default
- **Consistent**: Follows design system tokens
- **Type-safe**: Schema-driven validation

## 📋 Complete Atoms Library

### Form Controls (8 components)
Essential input elements for data collection and user interaction.

| Component | Purpose | Schema | Props | Status |
|-----------|---------|--------|-------|--------|
| **[Input](input/)** | Text input fields | `TextControlSchema.json` | text, type, validation | ✅ |
| **[Checkbox](checkbox/)** | Boolean selections | `CheckboxControlSchema.json` | checked, label, value | ✅ |
| **[Radio](radio/)** | Single choice selections | `RadioControlSchema.json` | name, value, selected | ✅ |
| **[Switch](switch/)** | Toggle controls | `SwitchControlSchema.json` | checked, size, disabled | ✅ |
| **[Hidden](hidden/)** | Hidden form values | `HiddenControlSchema.json` | name, value | 🔄 |
| **[UUID](uuid/)** | Unique identifier inputs | `UUIDControlSchema.json` | format, readonly | 🔄 |
| **[Color Input](color-input/)** | Color selection | `InputColorControlSchema.json` | value, format, palette | 🔄 |
| **[Static](static/)** | Read-only displays | `StaticExactControlSchema.json` | value, format | 🔄 |

### Action Elements (2 components)  
Interactive elements that trigger actions and behaviors.

| Component | Purpose | Schema | Props | Status |
|-----------|---------|--------|-------|--------|
| **[Button](button/)** | Action triggers | `ButtonGroupSchema.json` | text, variant, onClick | ✅ |
| **[Action](action/)** | Generic actions | `ActionSchema.json` | type, actionType, payload | 🔄 |

### Display Elements (9 components)
Visual elements for presenting information and content.

| Component | Purpose | Schema | Props | Status |
|-----------|---------|--------|-------|--------|
| **[Badge](badge/)** | Status indicators | `BadgeObject.json` | text, variant, size | ✅ |
| **[Tag](tag/)** | Categorization labels | `TagSchema.json` | label, color, removable | ✅ |
| **[Icon](icon/)** | Visual symbols | `IconSchema.json` | name, size, color | ✅ |
| **[Image](image/)** | Media display | `ImageSchema.json` | src, alt, responsive | 🔄 |
| **[Link](link/)** | Navigation elements | `LinkSchema.json` | href, target, text | ✅ |
| **[Status](status/)** | State indicators | `StatusSchema.json` | level, text, icon | ✅ |
| **[Progress](progress/)** | Progress tracking | `ProgressSchema.json` | value, max, variant | ✅ |
| **[Spinner](spinner/)** | Loading indicators | `SpinnerSchema.json` | size, color, speed | ✅ |
| **[Divider](divider/)** | Content separation | `DividerSchema.json` | orientation, variant | ✅ |

### Utility Elements (8 components)
Supporting elements for layout, configuration, and metadata.

| Component | Purpose | Schema | Props | Status |
|-----------|---------|--------|-------|--------|
| **[State](state/)** | Component state | `StateSchema.json` | conditions, values | 🔄 |
| **[Label Align](label-align/)** | Label positioning | `LabelAlign.json` | alignment, position | 🔄 |
| **[Option](option/)** | Selection options | `Option.json` | value, label, disabled | 🔄 |
| **[Options](options/)** | Option collections | `Options.json` | source, labelField | 🔄 |
| **[Status Source](status-source/)** | Status data source | `StatusSource.json` | api, mapping | 🔄 |
| **[Icon Checked](icon-checked/)** | Checked state icons | `IconCheckedSchema.json` | icon, checkedIcon | 🔄 |
| **[Icon Item](icon-item/)** | Icon list items | `IconItemSchema.json` | icon, label, value | 🔄 |
| **[Field Types](field-types/)** | Field type definitions | Multiple schemas | type, validation | 🔄 |

## 🎨 Design System Integration

### Common Properties
All atomic components share these base properties:

```go
type BaseAtomProps struct {
    // Identity
    ID        string `json:"id"`
    ClassName string `json:"className"`
    TestID    string `json:"testid"`
    
    // Visibility
    Hidden    bool   `json:"hidden"`
    HiddenOn  string `json:"hiddenOn"`
    Visible   bool   `json:"visible"`
    VisibleOn string `json:"visibleOn"`
    
    // State
    Disabled   bool   `json:"disabled"`
    DisabledOn string `json:"disabledOn"`
    
    // Styling
    Style       map[string]interface{} `json:"style"`
    StaticClass string                 `json:"staticClassName"`
    
    // Events
    OnEvent map[string]EventConfig `json:"onEvent"`
}
```

### Size System
Consistent sizing across all atoms:

```go
type AtomSize string

const (
    SizeXS AtomSize = "xs"    // 16px base
    SizeSM AtomSize = "sm"    // 20px base  
    SizeMD AtomSize = "md"    // 24px base (default)
    SizeLG AtomSize = "lg"    // 32px base
    SizeXL AtomSize = "xl"    // 40px base
)
```

### Color Variants
Semantic color system for all atoms:

```go
type AtomVariant string

const (
    VariantDefault   AtomVariant = "default"   // Neutral gray
    VariantPrimary   AtomVariant = "primary"   // Brand color
    VariantSecondary AtomVariant = "secondary" // Subtle gray
    VariantSuccess   AtomVariant = "success"   // Green
    VariantWarning   AtomVariant = "warning"   // Yellow/Orange
    VariantDanger    AtomVariant = "danger"    // Red
    VariantInfo      AtomVariant = "info"      // Blue
)
```

## 🏗️ Implementation Patterns

### Schema Validation Pattern
```go
// Standard atom validation
func ValidateAtomProps(props map[string]interface{}, schema *JsonSchema) error {
    // 1. Validate required properties
    for _, required := range schema.Required {
        if _, exists := props[required]; !exists {
            return fmt.Errorf("required property missing: %s", required)
        }
    }
    
    // 2. Validate property types
    for key, value := range props {
        if err := validatePropertyType(key, value, schema); err != nil {
            return err
        }
    }
    
    // 3. Validate enums and patterns
    return validateConstraints(props, schema)
}
```

### Rendering Pattern
```go
// Standard atom renderer
type AtomRenderer interface {
    GetSchemaType() string
    ValidateProps(props map[string]interface{}, schema *JsonSchema) error
    Render(ctx context.Context, props map[string]interface{}) (TemplComponent, error)
}

// Example implementation
type InputRenderer struct{}

func (r *InputRenderer) Render(ctx context.Context, props map[string]interface{}) (TemplComponent, error) {
    inputProps := extractInputProps(props)
    
    return TemplComponent{
        Type: "Input",
        Props: inputProps,
        CSS: generateInputCSS(inputProps),
        Attributes: generateInputAttributes(inputProps),
        Events: generateInputEvents(inputProps),
    }, nil
}
```

### Component Factory Registration
```go
// Register all atoms with factory
func RegisterAtoms(factory *SchemaFactory) {
    // Form controls
    factory.RegisterRenderer("TextControlSchema", &InputRenderer{})
    factory.RegisterRenderer("CheckboxControlSchema", &CheckboxRenderer{})
    factory.RegisterRenderer("RadioControlSchema", &RadioRenderer{})
    factory.RegisterRenderer("SwitchControlSchema", &SwitchRenderer{})
    
    // Action elements
    factory.RegisterRenderer("ButtonGroupSchema", &ButtonRenderer{})
    factory.RegisterRenderer("ActionSchema", &ActionRenderer{})
    
    // Display elements
    factory.RegisterRenderer("BadgeObject", &BadgeRenderer{})
    factory.RegisterRenderer("TagSchema", &TagRenderer{})
    factory.RegisterRenderer("IconSchema", &IconRenderer{})
    // ... continue for all atoms
}
```

## 📱 Responsive Design for Atoms

### Mobile Adaptations
```css
/* Touch-friendly sizing on mobile */
@media (max-width: 479px) {
    .atom-input,
    .atom-button,
    .atom-checkbox {
        min-height: 44px;  /* Touch target */
        font-size: 16px;   /* Prevent zoom */
    }
    
    .atom-xs { min-height: 36px; }  /* Smaller minimum on mobile */
    .atom-sm { min-height: 40px; }
    .atom-md { min-height: 44px; }
    .atom-lg { min-height: 48px; }
    .atom-xl { min-height: 52px; }
}
```

### Responsive Properties
```go
type ResponsiveAtomProps struct {
    Mobile  interface{} `json:"mobile"`   // Mobile-specific props
    Tablet  interface{} `json:"tablet"`   // Tablet-specific props
    Desktop interface{} `json:"desktop"`  // Desktop-specific props
}
```

## ♿ Accessibility Standards

### Mandatory Accessibility Features
All atoms must implement:

```go
type AccessibilityProps struct {
    // Labels
    AriaLabel       string `json:"ariaLabel"`
    AriaLabelledBy  string `json:"ariaLabelledBy"`
    AriaDescribedBy string `json:"ariaDescribedBy"`
    
    // State
    AriaDisabled    bool   `json:"ariaDisabled"`
    AriaHidden      bool   `json:"ariaHidden"`
    AriaExpanded    *bool  `json:"ariaExpanded"`
    AriaPressed     *bool  `json:"ariaPressed"`
    AriaSelected    *bool  `json:"ariaSelected"`
    
    // Navigation
    TabIndex        int    `json:"tabIndex"`
    Role            string `json:"role"`
    
    // Live regions
    AriaLive        string `json:"ariaLive"`
    AriaAtomic      bool   `json:"ariaAtomic"`
}
```

### Testing Requirements
- **Keyboard navigation**: All focusable elements
- **Screen reader**: Proper announcements
- **Color contrast**: 4.5:1 minimum ratio
- **Focus indicators**: Clearly visible
- **Touch targets**: 44px minimum

## 🧪 Testing Strategy

### Atom Testing Template
```go
func TestAtomComponent(t *testing.T) {
    tests := []struct {
        name     string
        props    map[string]interface{}
        schema   string
        expected []string  // Expected CSS classes or attributes
        errors   []string  // Expected validation errors
    }{
        {
            name: "valid props",
            props: map[string]interface{}{
                "type": "text",
                "value": "test",
                "required": true,
            },
            schema: "TextControlSchema",
            expected: []string{"input", "required"},
        },
        {
            name: "invalid props",
            props: map[string]interface{}{
                "type": "invalid",
            },
            schema: "TextControlSchema",
            errors: []string{"invalid type"},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            component, err := renderAtom(tt.props, tt.schema)
            
            if tt.errors != nil {
                assert.Error(t, err)
                for _, expectedError := range tt.errors {
                    assert.Contains(t, err.Error(), expectedError)
                }
            } else {
                assert.NoError(t, err)
                for _, expectedClass := range tt.expected {
                    assert.Contains(t, component.CSS, expectedClass)
                }
            }
        })
    }
}
```

### Visual Regression Testing
```javascript
// Test all atom variants
test.describe('Atoms Visual Tests', () => {
    test('all atom variants', async ({ page }) => {
        await page.goto('/components/atoms');
        
        // Test each size variant
        for (const size of ['xs', 'sm', 'md', 'lg', 'xl']) {
            await expect(page.locator(`[data-size="${size}"]`)).toHaveScreenshot(`atom-${size}.png`);
        }
        
        // Test each color variant
        for (const variant of ['default', 'primary', 'success', 'danger']) {
            await expect(page.locator(`[data-variant="${variant}"]`)).toHaveScreenshot(`atom-${variant}.png`);
        }
    });
});
```

## 🚀 Performance Optimization

### Atom-Specific Optimizations
```go
// Memoization for frequently used atoms
var atomCache = make(map[string]templ.Component)

func CachedAtom(key string, atomType string, props map[string]interface{}) templ.Component {
    cacheKey := fmt.Sprintf("%s:%s:%s", atomType, key, hashProps(props))
    
    if cached, exists := atomCache[cacheKey]; exists {
        return cached
    }
    
    component, _ := factory.RenderToTempl(context.Background(), atomType, props)
    atomCache[cacheKey] = component
    return component
}

// Bundle splitting for atoms
func LoadAtomsAsync() {
    // Load critical atoms immediately
    loadAtoms([]string{"Input", "Button", "Checkbox"})
    
    // Load remaining atoms on demand
    go loadAtoms(getAllAtoms())
}
```

## 📚 Usage Examples

### Basic Atom Usage
```go
templ ExampleForm() {
    <form class="space-y-4">
        @Input(InputProps{
            Type:        "text",
            Name:        "username",
            Label:       "Username",
            Required:    true,
            Placeholder: "Enter username",
        })
        
        @Checkbox(CheckboxProps{
            Name:    "terms",
            Label:   "I agree to the terms",
            Required: true,
        })
        
        @Button(ButtonProps{
            Text:    "Submit",
            Type:    "submit",
            Variant: "primary",
        })
    </form>
}
```

### Advanced Atom Composition
```go
templ AdvancedExample() {
    <div class="form-group">
        @Input(InputProps{
            Type:     "email",
            Name:     "email",
            Label:    "Email Address",
            Required: true,
            Icon:     "envelope",
            Validation: ValidationConfig{
                Pattern: `^[^\s@]+@[^\s@]+\.[^\s@]+$`,
                Message: "Please enter a valid email address",
            },
        })
        
        @Badge(BadgeProps{
            Text:    "Required",
            Variant: "warning",
            Size:    "sm",
        })
    </div>
}
```

## 🔗 Related Documentation

### Deep Dive
- **[Button Component](button/)**: Complete button documentation
- **[Input Component](input/)**: Text input specifications
- **[Design System](../../design_system.md)**: Visual guidelines
- **[Schema System](../../fundamentals/schema-system.md)**: Type validation

### Integration
- **[Molecules](../molecules/)**: Atom combinations
- **[Form Patterns](../organisms/forms/)**: Form composition
- **[Component Lifecycle](../../fundamentals/component-lifecycle.md)**: Creation process

---

**Documentation Status**: 52% complete (14 of 27 atoms documented)  
**Active Development**: Comprehensive documentation with examples and accessibility  
**Last Updated**: October 2025  
**Maintainer**: Atoms Team