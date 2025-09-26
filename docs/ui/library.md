# Templ UI Component Library Documentation

## Table of Contents

1. [Introduction](#introduction)
2. [Quick Start Guide](#quick-start-guide)
3. [Core Concepts](#core-concepts)
4. [Element Reference](#element-reference)
5. [Composition Patterns](#composition-patterns)
6. [Styling & Theming](#styling--theming)
7. [Accessibility](#accessibility)
8. [Best Practices](#best-practices)
9. [Publishing Guide](#publishing-guide)
10. [References](#references)

---

## 1. Introduction

This library provides foundational UI components built with **Templ** (Go templating), styled with **Flowbite/Tailwind CSS**, and enhanced with optional **Alpine.js** interactivity. Components follow atomic design principles to enable flexible composition.

### Technology Stack

| Technology | Purpose | Required |
|------------|---------|----------|
| **[Templ](https://templ.guide)** | Go templating language that compiles to type-safe Go code | ✅ Yes |
| **[Flowbite](https://flowbite.com)** | UI component library built on Tailwind CSS | ✅ Yes |
| **[Alpine.js](https://alpinejs.dev)** | Lightweight JavaScript framework for reactivity | ❌ Optional |

### Design Principles

- **Atomic**: Each component has a single, clear purpose
- **Composable**: Components combine to create complex UIs
- **Accessible**: WCAG 2.1 AA compliance built-in
- **Consistent**: Unified design system across all components
- **Progressive**: Works without JavaScript, enhanced with it

---

## 2. Quick Start Guide

### Installation

**1. Install Templ CLI**
```bash
go install github.com/a-h/templ/cmd/templ@latest
```

**2. Initialize Project**
```bash
mkdir my-ui-project && cd my-ui-project
go mod init my-ui-project
go get github.com/a-h/templ
```

**3. Add Dependencies**

Add to your HTML `<head>`:
```html
<!-- Flowbite CSS (required) -->
<link href="https://cdn.jsdelivr.net/npm/flowbite@2.5.1/dist/flowbite.min.css" rel="stylesheet">

<!-- Alpine.js (optional) -->
<script defer src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"></script>

<!-- Flowbite JS (required for interactive components) -->
<script src="https://cdn.jsdelivr.net/npm/flowbite@2.5.1/dist/flowbite.min.js"></script>
```

### Basic Example

**hello.templ**
```go
package components

type ButtonProps struct {
    Text    string
    Variant string
    OnClick string
}

templ Button(props ButtonProps) {
    <button 
        class={ "px-5 py-2.5 text-sm font-medium rounded-lg focus:ring-4 focus:outline-none " + getButtonClasses(props.Variant) }
        onclick={ props.OnClick }
        type="button"
    >
        { props.Text }
    </button>
}

func getButtonClasses(variant string) string {
    switch variant {
    case "primary":
        return "text-white bg-blue-700 hover:bg-blue-800 focus:ring-blue-300"
    case "secondary":
        return "text-gray-900 bg-white border border-gray-300 hover:bg-gray-100 focus:ring-gray-200"
    default:
        return "text-white bg-blue-700 hover:bg-blue-800 focus:ring-blue-300"
    }
}
```

**Generate and use:**
```bash
templ generate
```

**main.go**
```go
package main

import (
    "net/http"
    "github.com/a-h/templ"
    "my-ui-project/components"
)

func main() {
    button := components.Button(components.ButtonProps{
        Text: "Click me",
        Variant: "primary",
        OnClick: "alert('Hello!')",
    })
    
    http.Handle("/", templ.Handler(button))
    http.ListenAndServe(":8080", nil)
}
```

---

## 3. Core Concepts

### Component Architecture

Components are organized in three layers:

```
Application Components (your code)
    ↓
Complex Components (composed)
    ↓  
Atomic Elements (this library)
```

### Props Pattern

All components follow a consistent props structure:

```go
type ComponentProps struct {
    // Content properties
    Text        string
    Value       string
    Placeholder string
    
    // Styling properties  
    Variant     string    // "primary", "secondary", etc.
    Size        string    // "sm", "base", "lg"
    Color       string    // "blue", "red", "green"
    
    // Behavior properties
    Disabled    bool
    Loading     bool
    OnClick     string    // JavaScript handler
    
    // HTML properties
    ID          string
    Class       string    // Additional CSS classes
    AriaLabel   string
}
```

### State Management Options

| Approach | Use Case | Example |
|----------|----------|---------|
| **No JavaScript** | Static content, server-side forms | Contact forms, documentation |
| **Alpine.js** | Client-side interactivity, form validation | Multi-step forms, filters |
| **HTMX** | Server-driven interactions | Live updates, partial page loads |
| **Custom JS** | Complex behaviors | Charts, rich text editors |

---

## 4. Element Reference

### 4.1 Button

**Purpose**: Interactive element for user actions

**Props**
```go
type ButtonProps struct {
    Text        string  // Button text
    Variant     string  // "primary", "secondary", "success", "danger", "warning", "info"
    Size        string  // "xs", "sm", "base", "lg", "xl"
    Type        string  // "button", "submit", "reset"
    Disabled    bool    // Disable interaction
    Loading     bool    // Show spinner, disable interaction
    FullWidth   bool    // Expand to container width
    OnClick     string  // JavaScript click handler
    ID          string
    Class       string
    AriaLabel   string
}
```

**Usage Examples**

```go
// Primary action button
Button(ButtonProps{
    Text: "Save Changes",
    Variant: "primary",
    Type: "submit",
})

// Loading state
Button(ButtonProps{
    Text: "Processing...",
    Variant: "primary", 
    Loading: true,
    Disabled: true,
})

// Destructive action
Button(ButtonProps{
    Text: "Delete Account",
    Variant: "danger",
    OnClick: "confirmDelete()",
})
```

### 4.2 Input

**Purpose**: Single-line text input with validation

**Props**
```go
type InputProps struct {
    Type         string  // "text", "email", "password", "number", "tel", "url", "search"
    Value        string  // Input value
    Placeholder  string  // Placeholder text
    Name         string  // Form field name
    Label        string  // Associated label
    HelperText   string  // Help or error message
    Size         string  // "sm", "base", "lg"
    Validation   string  // "none", "success", "error"
    Required     bool    // HTML5 required attribute
    Disabled     bool    // Disable input
    ReadOnly     bool    // Make read-only
    ID           string
    Class        string
    OnChange     string  // JavaScript change handler
    OnFocus      string
    OnBlur       string
}
```

**Usage Examples**

```go
// Basic text input
Input(InputProps{
    Label: "Full Name",
    Type: "text",
    Name: "fullName",
    Required: true,
    Placeholder: "Enter your full name",
})

// Email with validation
Input(InputProps{
    Label: "Email Address",
    Type: "email", 
    Name: "email",
    Validation: "error",
    HelperText: "Please enter a valid email address",
})

// Search input
Input(InputProps{
    Type: "search",
    Placeholder: "Search products...",
    OnChange: "filterProducts(this.value)",
})
```

### 4.3 Card

**Purpose**: Container for grouping related content

**Props**
```go
type CardProps struct {
    Title       string           // Header title
    Subtitle    string           // Header subtitle  
    Content     templ.Component  // Main content area
    Footer      templ.Component  // Footer content
    Image       string           // Header image URL
    ImageAlt    string           // Image alt text
    Bordered    bool             // Show border
    Shadow      string           // "none", "sm", "base", "md", "lg", "xl"
    Padding     string           // "none", "sm", "base", "lg", "xl"
    Clickable   bool             // Make entire card clickable
    OnClick     string           // JavaScript click handler
    MaxWidth    string           // Maximum width
    ID          string
    Class       string
}
```

**Usage Examples**

```go
// Product card
Card(CardProps{
    Title: "Premium Headphones",
    Subtitle: "$299.99",
    Image: "/images/headphones.jpg",
    ImageAlt: "Premium wireless headphones",
    Content: productDescription,
    Footer: productActions,
    Shadow: "lg",
    Clickable: true,
    OnClick: "viewProduct('headphones-123')",
})

// Simple content card
Card(CardProps{
    Title: "Getting Started",
    Content: helpContent,
    Bordered: true,
    Padding: "lg",
})
```

### 4.4 Alert

**Purpose**: Display important messages and notifications

**Props**
```go
type AlertProps struct {
    Message     string           // Alert message
    Title       string           // Optional title
    Variant     string           // "info", "success", "warning", "error"
    Dismissible bool             // Show close button
    Icon        bool             // Show status icon
    Border      bool             // Show colored border
    Actions     templ.Component  // Custom action buttons
    OnDismiss   string           // JavaScript dismiss handler
    ID          string
    Class       string
}
```

**Usage Examples**

```go
// Success notification
Alert(AlertProps{
    Title: "Success!",
    Message: "Your changes have been saved.",
    Variant: "success",
    Dismissible: true,
    Icon: true,
})

// Error with action
Alert(AlertProps{
    Title: "Connection Failed",
    Message: "Unable to save your changes. Please try again.",
    Variant: "error",
    Actions: retryButton,
    OnDismiss: "hideAlert()",
})
```

### 4.5 Modal

**Purpose**: Overlay dialog for focused content

**Props**
```go
type ModalProps struct {
    Title       string           // Header title
    Content     templ.Component  // Modal body
    Footer      templ.Component  // Action buttons
    Size        string           // "xs", "sm", "default", "lg", "xl", "2xl"
    Position    string           // "center", "top"
    Backdrop    string           // "default", "static", "none"
    Keyboard    bool             // ESC key closes modal
    Persistent  bool             // Prevent closing
    ID          string           // Required for targeting
    Class       string
}
```

**Usage Examples**

```go
// Confirmation dialog
Modal(ModalProps{
    ID: "delete-confirmation",
    Title: "Delete Account",
    Content: deleteWarning,
    Footer: confirmationButtons,
    Size: "sm",
    Backdrop: "static",
})

// Large content modal
Modal(ModalProps{
    ID: "terms-modal", 
    Title: "Terms of Service",
    Content: termsContent,
    Size: "xl",
    Position: "center",
})
```

### 4.6 Form Elements

#### Select

```go
type SelectOption struct {
    Value    string
    Label    string
    Selected bool
    Disabled bool
}

type SelectProps struct {
    Options     []SelectOption
    Value       string
    Placeholder string
    Name        string
    Label       string
    Multiple    bool
    Searchable  bool    // Requires Alpine.js
    Size        string  // "sm", "base", "lg"
    Validation  string  // "none", "success", "error"
    Disabled    bool
    Required    bool
    ID          string
    Class       string
    OnChange    string
}
```

#### Checkbox

```go
type CheckboxProps struct {
    Checked       bool
    Label         string
    Name          string
    Value         string
    Disabled      bool
    Indeterminate bool    // Three-state checkbox
    Color         string  // "blue", "red", "green", etc.
    Size          string  // "sm", "base", "lg"
    HelperText    string
    ID            string
    Class         string
    OnChange      string
}
```

#### Textarea

```go
type TextareaProps struct {
    Value       string
    Placeholder string
    Name        string
    Label       string
    Rows        int     // Default: 4
    Resize      string  // "none", "both", "horizontal", "vertical"
    MaxLength   int     // Character limit
    Validation  string  // "none", "success", "error"
    HelperText  string
    Disabled    bool
    ReadOnly    bool
    Required    bool
    ID          string
    Class       string
    OnChange    string
}
```

---

## 5. Composition Patterns

### 5.1 Form Pattern

Create complete forms by combining form elements:

```go
templ ContactForm() {
    <form class="space-y-6">
        @Card(CardProps{
            Title: "Contact Us",
            Content: contactFormContent(),
            Footer: formActions(),
            Padding: "lg",
        })
    </form>
}

templ contactFormContent() {
    <div class="space-y-4">
        @Input(InputProps{
            Label: "Name",
            Type: "text",
            Name: "name",
            Required: true,
        })
        @Input(InputProps{
            Label: "Email", 
            Type: "email",
            Name: "email",
            Required: true,
        })
        @Textarea(TextareaProps{
            Label: "Message",
            Name: "message",
            Rows: 4,
            Required: true,
        })
    </div>
}

templ formActions() {
    <div class="flex justify-end space-x-3">
        @Button(ButtonProps{
            Text: "Cancel",
            Variant: "secondary",
        })
        @Button(ButtonProps{
            Text: "Send Message",
            Variant: "primary",
            Type: "submit",
        })
    </div>
}
```

### 5.2 Navigation Pattern

Build responsive navigation with dropdowns:

```go
templ MainNavigation() {
    <nav class="bg-white border-gray-200 dark:bg-gray-900">
        <div class="max-w-screen-xl flex flex-wrap items-center justify-between mx-auto p-4">
            // Brand/Logo
            <a href="/" class="flex items-center space-x-3">
                <span class="text-2xl font-semibold">MyApp</span>
            </a>
            
            // Navigation Links
            <div class="hidden w-full md:block md:w-auto">
                <ul class="flex flex-col md:flex-row md:space-x-8">
                    @navLink("Home", "/", true)
                    @navLink("About", "/about", false)
                    @navDropdown("Services", serviceItems())
                    @navLink("Contact", "/contact", false)
                </ul>
            </div>
        </div>
    </nav>
}
```

### 5.3 Dashboard Layout

Create admin interfaces with cards and grids:

```go
templ Dashboard() {
    <div class="p-6">
        // Stats Grid
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
            @statsCard("Total Users", "1,234", "+12%", "success")
            @statsCard("Revenue", "$45,678", "+8%", "success") 
            @statsCard("Orders", "892", "-3%", "warning")
        </div>
        
        // Main Content
        <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
            <div class="lg:col-span-2">
                @Card(CardProps{
                    Title: "Recent Activity",
                    Content: activityList(),
                })
            </div>
            <div>
                @Card(CardProps{
                    Title: "Quick Actions", 
                    Content: actionButtons(),
                })
            </div>
        </div>
    </div>
}
```

---

## 6. Styling & Theming

### Color System

Components use Flowbite's semantic color system:

| Color | Use Case | CSS Class |
|-------|----------|-----------|
| `blue` | Primary actions, links | `bg-blue-700`, `text-blue-600` |
| `gray` | Secondary actions, borders | `bg-gray-100`, `text-gray-500` |
| `green` | Success states | `bg-green-700`, `text-green-600` |
| `red` | Error states, destructive actions | `bg-red-700`, `text-red-600` |
| `yellow` | Warning states | `bg-yellow-400`, `text-yellow-800` |

### Customization

**1. CSS Custom Properties**
```css
:root {
  --primary-50: #eff6ff;
  --primary-500: #3b82f6;
  --primary-900: #1e3a8a;
}
```

**2. Tailwind Config Extension**
```javascript
module.exports = {
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
  }
}
```

**3. Component Class Override**
```go
Button(ButtonProps{
    Text: "Custom Button",
    Class: "bg-purple-600 hover:bg-purple-700", // Override default styling
})
```

### Responsive Design

All components include responsive breakpoints:

| Breakpoint | Size | Usage |
|------------|------|-------|
| `sm` | ≥640px | Small tablets |
| `md` | ≥768px | Tablets | 
| `lg` | ≥1024px | Desktop |
| `xl` | ≥1280px | Large desktop |
| `2xl` | ≥1536px | Extra large |

```go
// Responsive grid example
<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
```

---

## 7. Accessibility

### Built-in Features

All components include accessibility features by default:

- **Semantic HTML**: Proper element types and structure
- **ARIA attributes**: Labels, descriptions, states  
- **Keyboard navigation**: Tab order, focus management
- **Color contrast**: WCAG AA compliance
- **Screen reader support**: Descriptive text and states

### Best Practices

**Form Labels**
```go
// ✅ Good: Proper label association
Input(InputProps{
    Label: "Email Address",
    Name: "email",
    Type: "email",
    Required: true,
})

// ❌ Bad: Missing label
Input(InputProps{
    Placeholder: "Enter email", // Not sufficient
    Name: "email",
})
```

**Button Text**
```go
// ✅ Good: Descriptive text
Button(ButtonProps{
    Text: "Save Changes",
    AriaLabel: "Save changes to your profile",
})

// ❌ Bad: Generic text
Button(ButtonProps{
    Text: "Click Here", // Not descriptive
})
```

**Color and Contrast**
```go
// ✅ Good: Don't rely on color alone
Alert(AlertProps{
    Message: "Error: Please fix the issues below",
    Variant: "error",
    Icon: true, // Visual indicator beyond color
})
```

### Testing Checklist

- [ ] Can all interactive elements be reached with Tab key?
- [ ] Do form fields have associated labels?
- [ ] Is color contrast at least 4.5:1 for normal text?
- [ ] Do images have appropriate alt text?
- [ ] Can the interface be used with a screen reader?

---

## 8. Best Practices

### Component Design

**Single Responsibility**
```go
// ✅ Good: Focused component
templ SearchInput(props SearchProps) {
    // Only handles search input functionality
}

// ❌ Bad: Mixed responsibilities  
templ SearchWithResultsAndFilters(props ComplexProps) {
    // Does too many things
}
```

**Composability**
```go
// ✅ Good: Composable design
templ ProductCard(product Product) {
    @Card(CardProps{
        Content: productContent(product),
        Footer: productActions(product),
    })
}

// ❌ Bad: Monolithic component
templ ProductCardWithEverything(product Product, user User, cart Cart) {
    // Too many dependencies
}
```

### Performance

**Minimize Allocations**
```go
// ✅ Good: Reuse slices
var buttonClasses = map[string]string{
    "primary": "bg-blue-700 text-white",
    "secondary": "bg-gray-100 text-gray-900",
}

func getButtonClass(variant string) string {
    if class, ok := buttonClasses[variant]; ok {
        return class
    }
    return buttonClasses["primary"]
}
```

**Conditional Rendering**
```go
templ Button(props ButtonProps) {
    <button class={ getButtonClasses(props) }>
        if props.Loading {
            <svg class="animate-spin h-4 w-4 mr-2" viewBox="0 0 24 24">
                // Spinner SVG
            </svg>
        }
        { props.Text }
    </button>
}
```

### Error Handling

**Validation**
```go
func (p ButtonProps) Validate() error {
    if p.Text == "" {
        return errors.New("button text is required")
    }
    if !isValidVariant(p.Variant) {
        return fmt.Errorf("invalid variant: %s", p.Variant)
    }
    return nil
}
```

**Graceful Degradation**
```go
templ Modal(props ModalProps) {
    if props.ID == "" {
        // Log error, render simple div instead
        <div class="border border-red-500 p-4">
            <p>Error: Modal requires an ID</p>
            @props.Content
        </div>
        return
    }
    // Normal modal rendering
}
```

### Testing

**Component Tests**
```go
func TestButton(t *testing.T) {
    props := ButtonProps{
        Text: "Test Button",
        Variant: "primary",
    }
    
    // Render component
    buf := &bytes.Buffer{}
    err := Button(props).Render(context.Background(), buf)
    require.NoError(t, err)
    
    // Assert HTML output
    html := buf.String()
    assert.Contains(t, html, "Test Button")
    assert.Contains(t, html, "bg-blue-700")
}
```

**Visual Regression Tests**
Use tools like Playwright or Selenium to capture screenshots and detect visual changes.

---

## 9. Publishing Guide

### Package Structure

```
github.com/yourorg/templ-ui/
├── go.mod
├── go.sum
├── README.md
├── LICENSE
├── CHANGELOG.md
├── components/
│   ├── button/
│   │   ├── button.templ
│   │   ├── button.go         // Generated
│   │   └── button_test.go
│   ├── input/
│   │   ├── input.templ
│   │   ├── input.go          // Generated  
│   │   └── input_test.go
│   └── shared/
│       ├── types.go          // Common types
│       └── utils.go          // Helper functions
├── examples/
│   ├── basic/
│   │   └── main.go
│   └── advanced/
│       └── main.go
└── docs/
    ├── components/
    └── guides/
```

### Versioning Strategy

Follow semantic versioning (semver):

- **Major (v2.0.0)**: Breaking API changes
- **Minor (v1.1.0)**: New features, backward compatible
- **Patch (v1.0.1)**: Bug fixes, no API changes

### Release Process

1. **Update version** in go.mod and documentation
2. **Generate components** with `templ generate`  
3. **Run tests** `go test ./...`
4. **Update CHANGELOG.md** with changes
5. **Create Git tag** `git tag v1.0.0`
6. **Push tag** `git push origin v1.0.0`
7. **Publish release** on GitHub with notes

### Documentation Requirements

- **README.md**: Quick start guide and examples
- **Component docs**: Props, examples, accessibility notes
- **Migration guides**: For breaking changes
- **API reference**: Generated from Go docs

---

## 10. References

### Official Documentation

- **[Templ Guide](https://templ.guide)**: Go templating language
- **[Flowbite Components](https://flowbite.com/docs/components/)**: UI component library
- **[Alpine.js Docs](https://alpinejs.dev)**: JavaScript reactivity framework
- **[Tailwind CSS](https://tailwindcss.com/docs)**: Utility-first CSS framework

### Community Resources

- **[Templ Examples](https://github.com/a-h/templ/tree/main/examples)**: Official examples
- **[Go Project Layout](https://github.com/golang-standards/project-layout)**: Standard Go project structure
- **[Web Accessibility](https://www.w3.org/WAI/WCAG21/quickref/)**: WCAG 2.1 guidelines

### Related Tools

- **[HTMX](https://htmx.org)**: HTML-over-the-wire interactions
- **[Air](https://github.com/cosmtrek/air)**: Live reload for Go apps
- **[golangci-lint](https://golangci-lint.run)**: Go linting tool

---

*This documentation provides a comprehensive guide to building and using Templ UI components. For the latest updates and community support, visit our [GitHub repository](https://github.com/yourorg/templ-ui).*
