# ERP Views

**Schema-driven UI components built with Templ, HTMX, Alpine.js, and Tailwind CSS**

## 🏗️ Architecture

This UI layer follows the **Shadcn/ui design philosophy** adapted for our Go + HTMX stack:

```
┌─────────────────────────────────────────────────────────────┐
│                    ERP UI ARCHITECTURE                      │
├─────────────────────────────────────────────────────────────┤
│  🧩 Component Hierarchy                                     │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────── │
│  │    Atomic       │  │   Molecular     │  │   Organism    │
│  │  - Button       │  │  - FormField    │  │  - DataTable  │
│  │  - Input        │  │  - SearchBox    │  │  - Form       │
│  │  - Icon         │  │  - MenuItem     │  │  - Modal      │
│  │  - Badge        │  │  - Card         │  │  - Navigation │
│  └─────────────────┘  └─────────────────┘  └─────────────── │
├─────────────────────────────────────────────────────────────┤
│  🎨 Design System                                          │
│  - Design tokens (CSS custom properties)                   │
│  - Component variants (primary, secondary, etc.)           │
│  - Consistent sizing system (xs, sm, md, lg, xl)          │
│  - State management (hover, focus, disabled, error)        │
│  - Dark mode support                                       │
│  - RTL language support                                    │
├─────────────────────────────────────────────────────────────┤
│  ⚡ Tech Stack                                             │
│  - **Templ**: Server-side component templating            │
│  - **HTMX**: Progressive enhancement & hypermedia          │
│  - **Alpine.js**: Lightweight reactive JavaScript         │
│  - **Tailwind CSS**: Utility-first styling               │
│  - **Go Templates**: Fallback for simple cases           │
└─────────────────────────────────────────────────────────────┘
```

## 📁 Project Structure

```
views/
├── components/                 # Templ components
│   ├── ui/                    # Atomic components (Button, Input, etc.)
│   ├── forms/                 # Form-specific components
│   ├── layouts/               # Page layouts and shells
│   └── schemas/               # Schema-driven components
├── static/
│   ├── css/
│   │   ├── tokens.css         # Design tokens (CSS variables)
│   │   ├── base.css           # Base styles and resets
│   │   ├── components.css     # Component-specific styles
│   │   └── utilities.css      # Custom utility classes
│   ├── js/
│   │   ├── alpine/            # Alpine.js components/stores
│   │   ├── htmx/              # HTMX extensions
│   │   ├── datatable/         # Existing TypeScript datatable
│   │   └── utils/             # JavaScript utilities
│   └── icons/                 # SVG icon library
├── handlers/                  # Go HTTP handlers for views
├── types/                     # Go types for UI components
└── examples/                  # Usage examples and docs
```

## 🧩 Component Design Principles

### **Composable**
```html
<!-- Simple button -->
<Button>Click me</Button>

<!-- Composed button -->
<Button variant="primary" size="lg">
  <Icon name="plus" slot="icon-left"/>
  Create User
</Button>
```

### **Accessible**
- WCAG 2.1 Level AA compliance
- Proper ARIA attributes
- Keyboard navigation support
- Screen reader optimization

### **Predictable**
- Consistent naming conventions
- Standardized prop interfaces
- Reliable state management

### **Single Responsibility**
- Each component does one thing well
- Clear separation of concerns
- Easy to test and maintain

## 🎨 Design Token System

All components use CSS custom properties derived from our backend design tokens:

```css
/* Color tokens */
:root {
  --primary: 222.2 84% 4.9%;
  --primary-foreground: 210 40% 98%;
  --secondary: 210 40% 96%;
  /* ... */
}

/* Component usage */
.button-primary {
  background: hsl(var(--primary));
  color: hsl(var(--primary-foreground));
}
```

## 🔄 Schema Integration

Forms are automatically generated from backend schemas:

```go
// Backend schema definition
schema := schema.NewBuilder("user-form", schema.TypeForm, "Create User").
    AddTextField("firstName", "First Name", true).
    AddEmailField("email", "Email Address", true).
    Build(ctx)

// Frontend renders automatically with full validation, i18n, and theming
```

## 🌍 Internationalization

- Automatic locale detection from browser
- RTL layout support for Arabic/Hebrew
- Server-side translation integration
- Dynamic locale switching

## 🚀 Getting Started

1. **Development Setup**
```bash
# Start the development server
make dev

# Watch for changes
make watch

# Build for production
make build
```

2. **Creating Components**
```go
// components/ui/button.templ
templ Button(variant string, children ...templ.Component) {
    <button 
        class={buttonClasses(variant)}
        x-data="button"
        x-on:click="handleClick"
    >
        for _, child := range children {
            @child
        }
    </button>
}
```

3. **Using Components in Pages**
```go
// pages/users/create.templ
templ CreateUserPage() {
    @Layout() {
        @Card() {
            @CardHeader() {
                @CardTitle() { Create New User }
            }
            @CardContent() {
                @SchemaForm(userSchema)
            }
        }
    }
}
```

## 📚 Resources

- [Templ Documentation](https://templ.guide/)
- [HTMX Documentation](https://htmx.org/)
- [Alpine.js Documentation](https://alpinejs.dev/)
- [Tailwind CSS Documentation](https://tailwindcss.com/)
- [Design System Docs](./docs/)