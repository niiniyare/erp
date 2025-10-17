# Flowbite Integration Guide

**FILE PURPOSE**: Complete Flowbite integration for ERP UI components  
**SCOPE**: Component library, styling, customization, schema integration  
**TARGET AUDIENCE**: Frontend developers, component builders, designers

## 🎨 Flowbite Overview

[Flowbite](https://flowbite.com) is our primary UI component library, built on TailwindCSS, providing consistent design patterns and pre-built components for the ERP system.

### Why Flowbite for ERP?
- **TailwindCSS Foundation**: Utility-first CSS approach
- **Component Consistency**: Pre-designed, accessible components
- **Customization**: Theme customization and brand adaptation
- **Performance**: Optimized CSS and minimal JavaScript
- **Documentation**: Comprehensive component examples

## 🏗️ Integration Architecture

### Schema-Driven Flowbite Components
```
JSON Schema → Go Types → Flowbite Classes → Templ Components → HTML
```

Our system enhances Flowbite by adding:
- **Type Safety**: Go struct validation for component props
- **Schema Validation**: JSON schema-driven component generation
- **HTMX Integration**: Server-side interactivity patterns
- **Multi-Tenant**: Tenant-aware styling and theming

## 📦 Core Component Categories

### 1. Form Components
Essential interactive elements for data input and user actions.

#### Buttons
```go
type ButtonProps struct {
    Text     string        `json:"text"`
    Variant  ButtonVariant `json:"variant"`  // primary, secondary, success, danger
    Size     ButtonSize    `json:"size"`     // xs, sm, md, lg, xl
    Disabled bool          `json:"disabled"`
    Loading  bool          `json:"loading"`
    Icon     string        `json:"icon"`
    OnClick  string        `json:"onClick"`
}

// Generated Flowbite classes:
// "text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:ring-blue-300 font-medium rounded-lg text-sm px-5 py-2.5"
```

#### Form Inputs
```go
type InputProps struct {
    Name        string     `json:"name"`
    Type        InputType  `json:"type"`        // text, email, password, number
    Label       string     `json:"label"`
    Placeholder string     `json:"placeholder"`
    Value       string     `json:"value"`
    Required    bool       `json:"required"`
    Disabled    bool       `json:"disabled"`
    Error       string     `json:"error"`
    HelpText    string     `json:"helpText"`
}

// Generated classes:
// "bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5"
```

### 2. Layout Components

#### Cards
```go
type CardProps struct {
    Title       string                 `json:"title"`
    Content     string                 `json:"content"`
    Image       string                 `json:"image"`
    Actions     []ButtonProps          `json:"actions"`
    Badge       *BadgeProps            `json:"badge"`
    Variant     CardVariant            `json:"variant"`  // default, bordered, elevated
    MaxWidth    string                 `json:"maxWidth"` // sm, md, lg, xl, full
}
```

#### Tables
```go
type TableProps struct {
    Columns     []TableColumn          `json:"columns"`
    Data        []map[string]interface{} `json:"data"`
    Pagination  *PaginationProps       `json:"pagination"`
    Search      *SearchProps           `json:"search"`
    Sort        *SortProps             `json:"sort"`
    Actions     []TableAction          `json:"actions"`
    Responsive  bool                   `json:"responsive"`
    Striped     bool                   `json:"striped"`
    Hoverable   bool                   `json:"hoverable"`
}
```

### 3. Navigation Components

#### Navbar
```go
type NavbarProps struct {
    Brand       *BrandProps            `json:"brand"`
    Items       []NavItem              `json:"items"`
    Actions     []ButtonProps          `json:"actions"`
    Search      *SearchProps           `json:"search"`
    Theme       NavbarTheme            `json:"theme"`    // light, dark
    Fixed       bool                   `json:"fixed"`
    Border      bool                   `json:"border"`
}
```

#### Sidebar
```go
type SidebarProps struct {
    Items       []SidebarItem          `json:"items"`
    Header      *SidebarHeader         `json:"header"`
    Footer      *SidebarFooter         `json:"footer"`
    Collapsible bool                   `json:"collapsible"`
    Width       SidebarWidth           `json:"width"`    // sm, md, lg
    Position    SidebarPosition        `json:"position"` // left, right
}
```

### 4. Feedback Components

#### Alerts
```go
type AlertProps struct {
    Type        AlertType              `json:"type"`        // info, success, warning, danger
    Title       string                 `json:"title"`
    Message     string                 `json:"message"`
    Dismissible bool                   `json:"dismissible"`
    Icon        string                 `json:"icon"`
    Actions     []ButtonProps          `json:"actions"`
    AutoDismiss *AutoDismissProps      `json:"autoDismiss"`
}
```

#### Modals
```go
type ModalProps struct {
    Title       string                 `json:"title"`
    Content     string                 `json:"content"`
    Size        ModalSize              `json:"size"`        // sm, md, lg, xl, full
    Actions     []ButtonProps          `json:"actions"`
    Closable    bool                   `json:"closable"`
    Backdrop    ModalBackdrop          `json:"backdrop"`    // blur, dark, transparent
    Position    ModalPosition          `json:"position"`    // center, top, bottom
}
```

## 🎨 Theme Customization

### Brand Colors Integration
```go
type FlowbiteTheme struct {
    Primary struct {
        50  string `json:"50"`   // #eff6ff
        100 string `json:"100"`  // #dbeafe
        500 string `json:"500"`  // #3b82f6 (brand primary)
        600 string `json:"600"`  // #2563eb
        700 string `json:"700"`  // #1d4ed8
        900 string `json:"900"`  // #1e3a8a
    } `json:"primary"`
    
    Gray struct {
        50  string `json:"50"`   // #f9fafb
        100 string `json:"100"`  // #f3f4f6
        500 string `json:"500"`  // #6b7280
        900 string `json:"900"`  // #111827
    } `json:"gray"`
}
```

### Custom CSS Variables
```css
:root {
  /* Primary Brand Colors */
  --flowbite-primary-50: #eff6ff;
  --flowbite-primary-500: #3b82f6;
  --flowbite-primary-700: #1d4ed8;
  
  /* ERP Specific Colors */
  --erp-success: #10b981;
  --erp-warning: #f59e0b;
  --erp-danger: #ef4444;
  
  /* Spacing */
  --flowbite-spacing-xs: 0.25rem;
  --flowbite-spacing-sm: 0.5rem;
  --flowbite-spacing-md: 1rem;
}
```

### Dark Mode Support
```go
type DarkModeConfig struct {
    Enabled     bool                   `json:"enabled"`
    Default     DarkModePreference     `json:"default"`     // light, dark, system
    Toggle      *DarkModeToggle        `json:"toggle"`
    Classes     map[string]string      `json:"classes"`     // Component dark mode classes
}

// Dark mode class generation
func generateDarkModeClasses(component ComponentType) string {
    lightClasses := getFlowbiteLightClasses(component)
    darkClasses := getFlowbiteDarkClasses(component)
    
    return fmt.Sprintf("%s %s", lightClasses, darkClasses)
}
```

## 🔧 Schema Integration Patterns

### Component Factory Integration
```go
// Flowbite component renderer
type FlowbiteButtonRenderer struct{}

func (r *FlowbiteButtonRenderer) Render(ctx context.Context, props map[string]interface{}, schema *JsonSchema) (TemplComponent, error) {
    buttonProps := extractButtonProps(props)
    
    return TemplComponent{
        Type: "FlowbiteButton",
        Props: buttonProps,
        CSS: generateFlowbiteButtonCSS(buttonProps),
        Attributes: map[string]string{
            "type": buttonProps.Type,
            "class": buildFlowbiteClasses(buttonProps),
        },
        Events: map[string]string{
            "click": buttonProps.OnClick,
        },
    }, nil
}

func generateFlowbiteButtonCSS(props ButtonProps) string {
    classes := []string{
        "font-medium", "rounded-lg", "text-sm", "px-5", "py-2.5",
        "text-center", "inline-flex", "items-center",
    }
    
    // Variant-specific classes
    switch props.Variant {
    case ButtonPrimary:
        classes = append(classes, 
            "text-white", "bg-blue-700", "hover:bg-blue-800",
            "focus:ring-4", "focus:ring-blue-300")
    case ButtonSecondary:
        classes = append(classes,
            "text-gray-900", "bg-white", "border", "border-gray-300",
            "hover:bg-gray-100", "focus:ring-4", "focus:ring-gray-200")
    }
    
    // Size-specific classes
    switch props.Size {
    case ButtonXS:
        classes = append(classes, "px-3", "py-2", "text-xs")
    case ButtonSM:
        classes = append(classes, "px-3", "py-2", "text-sm")
    }
    
    return strings.Join(classes, " ")
}
```

### Schema Definition Examples
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "./FlowbiteButtonSchema.json",
  "type": "object",
  "properties": {
    "text": {
      "type": "string",
      "description": "Button text content",
      "minLength": 1,
      "maxLength": 50
    },
    "variant": {
      "type": "string",
      "enum": ["primary", "secondary", "success", "danger", "warning"],
      "default": "primary",
      "description": "Flowbite button variant"
    },
    "size": {
      "type": "string", 
      "enum": ["xs", "sm", "md", "lg", "xl"],
      "default": "md",
      "description": "Flowbite button size"
    },
    "disabled": {
      "type": "boolean",
      "default": false,
      "description": "Whether button is disabled"
    },
    "loading": {
      "type": "boolean",
      "default": false,
      "description": "Show loading spinner"
    }
  },
  "required": ["text"],
  "additionalProperties": false
}
```

## 🚀 HTMX Integration Patterns

### Server-Side Form Handling
```go
templ FlowbiteForm() {
    <form 
        class="space-y-6"
        hx-post="/api/users"
        hx-target="#form-result"
        hx-swap="innerHTML">
        
        // Flowbite input with HTMX validation
        <div>
            <label class="block mb-2 text-sm font-medium text-gray-900">Name</label>
            <input 
                type="text"
                name="name"
                class="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5"
                hx-validate="true"
                hx-post="/api/users/validate-name"
                hx-target="#name-error"
                hx-trigger="blur">
            <div id="name-error" class="mt-1 text-sm text-red-600"></div>
        </div>
        
        // Flowbite button with loading state
        <button 
            type="submit"
            class="text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:ring-blue-300 font-medium rounded-lg text-sm px-5 py-2.5 text-center"
            hx-indicator="#loading-spinner">
            <span class="htmx-indicator hidden" id="loading-spinner">
                <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white">...</svg>
            </span>
            Submit
        </button>
    </form>
    
    <div id="form-result"></div>
}
```

### Dynamic Content Updates
```go
templ FlowbiteDataTable() {
    <div class="relative overflow-x-auto shadow-md sm:rounded-lg">
        <table class="w-full text-sm text-left text-gray-500">
            <thead class="text-xs text-gray-700 uppercase bg-gray-50">
                <tr>
                    <th class="px-6 py-3">Name</th>
                    <th class="px-6 py-3">Email</th>
                    <th class="px-6 py-3">Status</th>
                    <th class="px-6 py-3">Actions</th>
                </tr>
            </thead>
            <tbody 
                hx-get="/api/users/table-rows"
                hx-trigger="load, every 30s"
                hx-target="this"
                hx-swap="innerHTML">
                // Dynamic rows loaded via HTMX
            </tbody>
        </table>
    </div>
}
```

## 📱 Responsive Design with Flowbite

### Mobile-First Components
```go
templ ResponsiveCard(props CardProps) {
    <div class={
        "max-w-sm p-6 bg-white border border-gray-200 rounded-lg shadow",
        "sm:max-w-md",
        "md:max-w-lg", 
        "lg:max-w-xl",
        templ.KV("hover:bg-gray-100", props.Hoverable),
    }>
        if props.Image != "" {
            <img class="w-full h-48 object-cover rounded-t-lg" src={ props.Image }>
        }
        
        <div class="p-5">
            <h5 class="mb-2 text-xl font-bold tracking-tight text-gray-900 sm:text-2xl">
                { props.Title }
            </h5>
            <p class="mb-3 font-normal text-gray-700 text-sm sm:text-base">
                { props.Content }
            </p>
            
            if len(props.Actions) > 0 {
                <div class="flex flex-col sm:flex-row gap-2">
                    for _, action := range props.Actions {
                        @FlowbiteButton(action)
                    }
                </div>
            }
        </div>
    </div>
}
```

### Responsive Navigation
```go
templ ResponsiveNavbar(props NavbarProps) {
    <nav class="bg-white border-gray-200 px-4 lg:px-6 py-2.5">
        <div class="flex flex-wrap justify-between items-center mx-auto max-w-screen-xl">
            // Brand - always visible
            <a href="#" class="flex items-center">
                <span class="self-center text-xl font-semibold whitespace-nowrap">
                    { props.Brand.Name }
                </span>
            </a>
            
            // Mobile menu button
            <button 
                class="inline-flex items-center p-2 ml-1 text-sm text-gray-500 rounded-lg lg:hidden hover:bg-gray-100"
                x-data
                @click="$refs.mobileMenu.classList.toggle('hidden')">
                <svg class="w-6 h-6">...</svg>
            </button>
            
            // Desktop menu - hidden on mobile
            <div class="hidden justify-between items-center w-full lg:flex lg:w-auto">
                <ul class="flex flex-col mt-4 font-medium lg:flex-row lg:space-x-8 lg:mt-0">
                    for _, item := range props.Items {
                        <li>
                            <a class="block py-2 pr-4 pl-3 text-gray-700 hover:bg-gray-50 lg:hover:bg-transparent lg:border-0 lg:hover:text-primary-700 lg:p-0">
                                { item.Text }
                            </a>
                        </li>
                    }
                </ul>
            </div>
            
            // Mobile menu - hidden by default
            <div class="hidden w-full lg:hidden" x-ref="mobileMenu">
                <ul class="flex flex-col mt-4 font-medium">
                    for _, item := range props.Items {
                        <li>
                            <a class="block py-2 pr-4 pl-3 text-gray-700 border-b border-gray-100 hover:bg-gray-50">
                                { item.Text }
                            </a>
                        </li>
                    }
                </ul>
            </div>
        </div>
    </nav>
}
```

## 🧪 Testing Flowbite Components

### Component Testing
```go
func TestFlowbiteButton(t *testing.T) {
    props := ButtonProps{
        Text:    "Test Button",
        Variant: ButtonPrimary,
        Size:    ButtonMD,
    }
    
    // Render component
    component, err := renderFlowbiteButton(props)
    require.NoError(t, err)
    
    html, err := templ.ToGoHTML(context.Background(), component)
    require.NoError(t, err)
    
    // Test Flowbite classes
    assert.Contains(t, html, "bg-blue-700")
    assert.Contains(t, html, "hover:bg-blue-800")
    assert.Contains(t, html, "rounded-lg")
    assert.Contains(t, html, "px-5 py-2.5")
}
```

### Visual Regression Testing
```bash
# Component visual testing
npx playwright test --project=chromium --grep="Flowbite Components"

# Test all button variants
npx playwright test tests/flowbite/buttons.spec.ts

# Test responsive behavior
npx playwright test tests/flowbite/responsive.spec.ts
```

## 📚 Component Examples

### Complete Examples Directory
All Flowbite components have been organized into example collections:

- **[Basic Components](../schemas/examples/flowbite-demos/basic/)**: Buttons, inputs, badges, alerts
- **[Form Components](../schemas/examples/flowbite-demos/forms/)**: Complete form examples with validation
- **[Layout Components](../schemas/examples/flowbite-demos/layout/)**: Cards, tables, grids
- **[Navigation Components](../schemas/examples/flowbite-demos/navigation/)**: Navbars, sidebars, breadcrumbs
- **[Advanced Components](../schemas/examples/flowbite-demos/advanced/)**: Modals, dropdowns, datepickers

### Quick Reference Links
- 📋 **[Button Examples](../schemas/examples/flowbite-demos/basic/buttons.md)**
- 📝 **[Form Examples](../schemas/examples/flowbite-demos/forms/complete-forms.md)**
- 📊 **[Table Examples](../schemas/examples/flowbite-demos/layout/tables.md)**
- 🧭 **[Navigation Examples](../schemas/examples/flowbite-demos/navigation/navbar.md)**

## 🚀 Performance Optimization

### CSS Optimization
```javascript
// TailwindCSS configuration for Flowbite optimization
module.exports = {
  content: [
    './web/**/*.{templ,go}',
    './node_modules/flowbite/**/*.js'
  ],
  theme: {
    extend: {
      colors: {
        primary: colors.blue,
      }
    }
  },
  plugins: [
    require('flowbite/plugin')({
      charts: true,
      forms: true,
      tooltips: true,
    })
  ]
}
```

### Component Lazy Loading
```go
templ LazyFlowbiteTable(props TableProps) {
    <div 
        hx-get="/api/tables/render"
        hx-trigger="intersect once"
        hx-vals={ fmt.Sprintf(`{"tableId": "%s"}`, props.ID) }>
        
        // Loading placeholder with Flowbite spinner
        <div class="flex justify-center items-center h-32">
            <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        </div>
    </div>
}
```

## 📚 Migration Guide

### From Raw Flowbite to Schema-Driven
```go
// Before: Raw Flowbite HTML
<button class="text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:ring-blue-300 font-medium rounded-lg text-sm px-5 py-2.5 text-center" type="button">
    Default button
</button>

// After: Schema-driven component
props := ButtonProps{
    Text:    "Default button",
    Variant: ButtonPrimary,
    Size:    ButtonMD,
}
component, _ := factory.RenderToTempl(ctx, "FlowbiteButtonSchema", props)
```

### Component Upgrade Path
1. **Identify existing Flowbite usage**
2. **Create schema definitions** for component props
3. **Implement component renderers**
4. **Test component generation**
5. **Replace raw HTML with schema components**

## 📚 Related Documentation

- **[Styling Approach](../fundamentals/styling-approach.md)**: CSS architecture and methodology
- **[Component Lifecycle](../fundamentals/component-lifecycle.md)**: Component creation process
- **[Schema System](../fundamentals/schema-system.md)**: Type-safe component generation
- **[Templ Integration](../fundamentals/templ-integration.md)**: Template development patterns

Flowbite provides the design foundation for our ERP UI system, enhanced with type safety, schema validation, and seamless server-side integration.