# Styling Approach

**FILE PURPOSE**: Complete guide to CSS architecture and styling methodology  
**SCOPE**: TailwindCSS, Flowbite integration, custom themes, responsive design  
**TARGET AUDIENCE**: Frontend developers, designers, component builders

## 🎨 Styling Architecture Overview

Our styling system combines **TailwindCSS utilities** with **Flowbite components** in a **schema-driven architecture** for consistent, maintainable, and scalable design.

### Technology Stack
- **[TailwindCSS](https://tailwindcss.com)**: Utility-first CSS framework
- **[Flowbite](https://flowbite.com)**: Component library built on Tailwind
- **Schema-driven CSS**: Type-safe styling through Go validation
- **Responsive Design**: Mobile-first, progressive enhancement

## 🏗️ CSS Architecture Layers

### 1. Foundation Layer (TailwindCSS)
```css
/* Base utilities for spacing, typography, colors */
.p-4         /* padding: 1rem */
.text-lg     /* font-size: 1.125rem */
.bg-blue-500 /* background-color: #3b82f6 */
```

### 2. Component Layer (Flowbite)
```css
/* Pre-built component classes */
.btn         /* Base button styles */
.card        /* Card component */
.form-input  /* Input field styles */
```

### 3. Schema Layer (Type-safe CSS)
```go
// CSS validation through schemas
cssProps := &css.ExpandedStyles{
    Color:           "#3b82f6",     // Validated
    BackgroundColor: "bg-blue-500", // TailwindCSS class
    Padding:         "16px",        // Size validation
}
```

### 4. Theme Layer (Customization)
```go
// Theme configuration
theme := map[string]interface{}{
    "primaryColor":   "#3b82f6",    // Blue
    "secondaryColor": "#6b7280",    // Gray
    "borderRadius":   "8px",
    "spacing":        "1rem",
}
```

## 📐 Design System Principles

### Atomic Design Implementation
```
atoms/      → Basic elements (buttons, inputs, icons)
molecules/  → Component combinations (search box, card header)
organisms/  → Complex components (navigation, data tables)
templates/  → Page layouts and structure
```

### Design Tokens
```go
// Standardized design values
type DesignTokens struct {
    Colors struct {
        Primary   string // #3b82f6
        Secondary string // #6b7280
        Success   string // #10b981
        Danger    string // #ef4444
        Warning   string // #f59e0b
    }
    Spacing struct {
        XS string // 0.25rem
        SM string // 0.5rem
        MD string // 1rem
        LG string // 1.5rem
        XL string // 3rem
    }
    Typography struct {
        FontFamily string // Inter, sans-serif
        FontSizes  map[string]string
        LineHeight map[string]string
    }
}
```

## 🎨 Flowbite Integration

### Component Styling Pattern
```go
// Schema-driven Flowbite component
type ButtonProps struct {
    Text     string         `json:"text"`
    Variant  ButtonVariant  `json:"variant"`  // primary, secondary, success
    Size     ButtonSize     `json:"size"`     // sm, md, lg, xl
    Class    string         `json:"className"` // Additional classes
}

func (p ButtonProps) GetClasses() string {
    classes := []string{"btn"}
    
    // Flowbite variant classes
    switch p.Variant {
    case ButtonPrimary:
        classes = append(classes, "text-white", "bg-blue-700", "hover:bg-blue-800")
    case ButtonSecondary:
        classes = append(classes, "text-gray-900", "bg-white", "border", "hover:bg-gray-100")
    }
    
    // Size classes
    switch p.Size {
    case ButtonSM:
        classes = append(classes, "px-3", "py-2", "text-xs")
    case ButtonMD:
        classes = append(classes, "px-5", "py-2.5", "text-sm")
    }
    
    // Custom classes
    if p.Class != "" {
        classes = append(classes, p.Class)
    }
    
    return strings.Join(classes, " ")
}
```

### Flowbite Component Library
```html
<!-- Button Components -->
<button class="text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:ring-blue-300 font-medium rounded-lg text-sm px-5 py-2.5">
    Primary Button
</button>

<!-- Form Components -->
<input class="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5">

<!-- Card Components -->
<div class="max-w-sm p-6 bg-white border border-gray-200 rounded-lg shadow">
    <h5 class="mb-2 text-2xl font-bold tracking-tight text-gray-900">Card Title</h5>
    <p class="mb-3 font-normal text-gray-700">Card content</p>
</div>
```

## 📱 Responsive Design Strategy

### Mobile-First Approach
```go
// Responsive component props
type ResponsiveProps struct {
    Mobile  ComponentSize  `json:"mobile"`   // Default
    Tablet  ComponentSize  `json:"tablet"`   // sm: prefix
    Desktop ComponentSize  `json:"desktop"`  // md: prefix
    Large   ComponentSize  `json:"large"`    // lg: prefix
}

// Usage in components
func generateResponsiveClasses(props ResponsiveProps) string {
    classes := []string{
        getClasses(props.Mobile),                    // Base mobile styles
        "sm:" + getClasses(props.Tablet),          // Tablet and up
        "md:" + getClasses(props.Desktop),         // Desktop and up
        "lg:" + getClasses(props.Large),           // Large screens
    }
    return strings.Join(classes, " ")
}
```

### Breakpoint System
```css
/* TailwindCSS default breakpoints */
/* sm: 640px   - Small tablets */
/* md: 768px   - Tablets */
/* lg: 1024px  - Small laptops */
/* xl: 1280px  - Laptops */
/* 2xl: 1536px - Large screens */
```

### Responsive Component Example
```go
// Responsive grid layout
gridProps := map[string]interface{}{
    "responsive": map[string]interface{}{
        "mobile":  map[string]interface{}{"columns": 1, "gap": "4"},
        "tablet":  map[string]interface{}{"columns": 2, "gap": "6"},
        "desktop": map[string]interface{}{"columns": 3, "gap": "8"},
        "large":   map[string]interface{}{"columns": 4, "gap": "10"},
    },
}

// Generated classes: "grid-cols-1 gap-4 sm:grid-cols-2 sm:gap-6 md:grid-cols-3 md:gap-8 lg:grid-cols-4 lg:gap-10"
```

## 🎨 Custom Theme Implementation

### Theme Configuration
```go
type Theme struct {
    Name        string `json:"name"`
    Description string `json:"description"`
    Colors      struct {
        Primary     ColorPalette `json:"primary"`
        Secondary   ColorPalette `json:"secondary"`
        Success     ColorPalette `json:"success"`
        Warning     ColorPalette `json:"warning"`
        Danger      ColorPalette `json:"danger"`
        Gray        ColorPalette `json:"gray"`
    } `json:"colors"`
    Typography struct {
        FontFamily   string            `json:"fontFamily"`
        FontSizes    map[string]string `json:"fontSizes"`
        FontWeights  map[string]string `json:"fontWeights"`
        LineHeights  map[string]string `json:"lineHeights"`
    } `json:"typography"`
    Spacing     map[string]string `json:"spacing"`
    BorderRadius map[string]string `json:"borderRadius"`
    Shadows     map[string]string `json:"shadows"`
}

type ColorPalette struct {
    50  string `json:"50"`   // Lightest
    100 string `json:"100"`
    200 string `json:"200"`
    300 string `json:"300"`
    400 string `json:"400"`
    500 string `json:"500"`  // Base color
    600 string `json:"600"`
    700 string `json:"700"`
    800 string `json:"800"`
    900 string `json:"900"`  // Darkest
}
```

### Theme Application
```go
// Apply theme to component
func applyTheme(component *TemplComponent, theme Theme) {
    // Replace color classes with theme colors
    classes := component.CSS
    
    // Primary color mapping
    classes = strings.ReplaceAll(classes, "bg-blue-500", fmt.Sprintf("bg-[%s]", theme.Colors.Primary[500]))
    classes = strings.ReplaceAll(classes, "text-blue-700", fmt.Sprintf("text-[%s]", theme.Colors.Primary[700]))
    
    // Update component
    component.CSS = classes
}
```

### Custom CSS Variables
```css
/* Generate CSS custom properties from theme */
:root {
  --color-primary-50: #eff6ff;
  --color-primary-500: #3b82f6;
  --color-primary-900: #1e3a8a;
  
  --font-family-sans: Inter, ui-sans-serif, system-ui;
  --spacing-xs: 0.25rem;
  --spacing-sm: 0.5rem;
  --border-radius-lg: 0.5rem;
}
```

## 🔧 CSS Validation System

### Schema-Based Validation
```go
// CSS property validation through schemas
type CSSValidator struct {
    colorSchemas    map[string]*ColorSchema
    spacingSchemas  map[string]*SpacingSchema
    layoutSchemas   map[string]*LayoutSchema
}

func (v *CSSValidator) ValidateStyles(styles css.ExpandedStyles) error {
    // Validate color values
    if err := v.validateColor(styles.Color); err != nil {
        return fmt.Errorf("invalid color: %w", err)
    }
    
    // Validate spacing values
    if err := v.validateSpacing(styles.Padding); err != nil {
        return fmt.Errorf("invalid padding: %w", err)
    }
    
    // Validate layout properties
    if err := v.validateLayout(styles.Display); err != nil {
        return fmt.Errorf("invalid display: %w", err)
    }
    
    return nil
}
```

### CSS Property Schemas
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "./ColorSchema.json",
  "type": "object",
  "properties": {
    "color": {
      "oneOf": [
        {
          "type": "string",
          "pattern": "^#[0-9A-Fa-f]{6}$",
          "description": "Hex color"
        },
        {
          "type": "string",
          "enum": [
            "text-gray-900", "text-blue-600", "text-green-500",
            "text-red-500", "text-yellow-500", "text-purple-600"
          ],
          "description": "TailwindCSS text color"
        }
      ]
    }
  }
}
```

## 🏗️ Component Styling Patterns

### Button Styling Pattern
```go
type ButtonStyleConfig struct {
    Base      []string          // Always applied
    Variants  map[string][]string // Variant-specific classes
    Sizes     map[string][]string // Size-specific classes
    States    map[string][]string // State classes (hover, focus, disabled)
}

var ButtonStyles = ButtonStyleConfig{
    Base: []string{
        "inline-flex", "items-center", "justify-center",
        "font-medium", "rounded-lg", "transition-colors",
        "focus:outline-none", "focus:ring-4",
    },
    Variants: map[string][]string{
        "primary": {
            "text-white", "bg-blue-700", "hover:bg-blue-800",
            "focus:ring-blue-300", "dark:bg-blue-600",
        },
        "secondary": {
            "text-gray-900", "bg-white", "border", "border-gray-200",
            "hover:bg-gray-100", "focus:ring-gray-200",
        },
    },
    Sizes: map[string][]string{
        "xs": {"px-3", "py-2", "text-xs"},
        "sm": {"px-3", "py-2", "text-sm"},
        "md": {"px-5", "py-2.5", "text-sm"},
        "lg": {"px-5", "py-3", "text-base"},
        "xl": {"px-6", "py-3.5", "text-base"},
    },
    States: map[string][]string{
        "disabled": {"opacity-50", "cursor-not-allowed"},
        "loading":  {"cursor-wait"},
    },
}
```

### Form Input Styling
```go
var InputStyles = map[string][]string{
    "base": {
        "block", "w-full", "text-gray-900", "border", "rounded-lg",
        "focus:ring-blue-500", "focus:border-blue-500",
    },
    "sizes": {
        "sm": {"p-2", "text-sm"},
        "md": {"p-2.5", "text-sm"},
        "lg": {"p-4", "text-base"},
    },
    "variants": {
        "default": {"bg-gray-50", "border-gray-300"},
        "error":   {"bg-red-50", "border-red-500", "text-red-900"},
        "success": {"bg-green-50", "border-green-500", "text-green-900"},
    },
}
```

## 🌙 Dark Mode Support

### Dark Mode Implementation
```go
type DarkModeStyles struct {
    Light []string `json:"light"` // Light mode classes
    Dark  []string `json:"dark"`  // Dark mode classes (with dark: prefix)
}

func generateDarkModeClasses(styles DarkModeStyles) string {
    classes := []string{}
    
    // Light mode classes (default)
    classes = append(classes, styles.Light...)
    
    // Dark mode classes with dark: prefix
    for _, darkClass := range styles.Dark {
        classes = append(classes, "dark:"+darkClass)
    }
    
    return strings.Join(classes, " ")
}

// Usage
buttonClasses := generateDarkModeClasses(DarkModeStyles{
    Light: []string{"bg-white", "text-gray-900", "border-gray-200"},
    Dark:  []string{"bg-gray-800", "text-white", "border-gray-600"},
})
// Result: "bg-white text-gray-900 border-gray-200 dark:bg-gray-800 dark:text-white dark:border-gray-600"
```

### Theme Toggle Component
```go
templ ThemeToggle() {
    <button 
        x-data="{ darkMode: false }"
        @click="darkMode = !darkMode; document.documentElement.classList.toggle('dark')"
        class="p-2 rounded-lg border border-gray-200 dark:border-gray-600">
        <span x-show="!darkMode">🌙</span>
        <span x-show="darkMode">☀️</span>
    </button>
}
```

## 🚀 Performance Optimization

### CSS Optimization Strategies
1. **Purge Unused Styles**: TailwindCSS purging in production
2. **Component-Specific CSS**: Load only required styles
3. **Critical CSS**: Inline critical path styles
4. **CSS Modules**: Scope styles to components

### Build Configuration
```javascript
// tailwind.config.js
module.exports = {
  content: [
    "./web/**/*.{templ,go}",
    "./docs/ui/**/*.md",
  ],
  theme: {
    extend: {
      colors: {
        primary: {
          50: '#eff6ff',
          500: '#3b82f6',
          900: '#1e3a8a',
        }
      }
    }
  },
  plugins: [
    require('flowbite/plugin')
  ]
}
```

## 📏 Design Consistency

### Style Guide Enforcement
```go
// Validate component adherence to design system
type DesignSystemValidator struct {
    allowedColors   []string
    allowedSpacing  []string
    allowedFonts    []string
}

func (v *DesignSystemValidator) ValidateComponent(component TemplComponent) []string {
    violations := []string{}
    
    // Check for non-standard colors
    if !v.isValidColor(component.CSS) {
        violations = append(violations, "uses non-standard colors")
    }
    
    // Check spacing consistency
    if !v.isValidSpacing(component.CSS) {
        violations = append(violations, "uses inconsistent spacing")
    }
    
    return violations
}
```

### Automated Style Linting
```bash
# CSS validation script
#!/bin/bash
echo "Validating component styles..."

# Check for non-Flowbite classes
grep -r "class.*custom-" web/ && echo "❌ Custom classes found"

# Validate color usage
grep -r "bg-\[#" web/ && echo "❌ Arbitrary color values found"

# Check for consistent spacing
echo "✅ Style validation complete"
```

## 📚 Related Documentation

- **[Architecture](architecture.md)**: System design principles
- **[Schema System](schema-system.md)**: Type-safe component generation
- **[Component Lifecycle](component-lifecycle.md)**: Component creation process
- **[Templ Integration](templ-integration.md)**: Template development patterns

Our styling approach ensures consistent, maintainable, and performant designs across the entire ERP system while providing flexibility for customization and theming.