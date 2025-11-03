# Awo ERP UI Component Library

A comprehensive, shadcn/ui-inspired component library built with **Go**, **Templ**, **HTMX**, and **Alpine.js** for the Awo ERP system. This library provides production-ready, accessible, and beautiful UI components for building enterprise applications.

## 🎨 Features

- **✨ Beautiful Design** - Inspired by shadcn/ui with modern, clean aesthetics
- **🔒 Type-Safe** - Built with Go and Templ for compile-time safety
- **♿ Accessible** - WCAG 2.1 AA compliant with proper ARIA attributes
- **🚀 Fast** - Server-side rendered with progressive enhancement
- **🎯 Schema-Driven** - Automatically render forms from JSON schemas
- **🌗 Dark Mode** - Built-in dark mode support
- **📱 Responsive** - Mobile-first design that works on all devices
- **🔌 Extensible** - Easy to customize and extend
- **⚡ HTMX Integration** - Seamless AJAX without complex JavaScript
- **🎭 Alpine.js** - Lightweight reactivity where needed

## 📦 Installation

### 1. Install Dependencies

```bash
# Install templ
go install github.com/a-h/templ/cmd/templ@latest

# Install Go dependencies
go get github.com/gofiber/fiber/v2
go get github.com/niiniyare/erp/pkg/schema
```

### 2. Install Frontend Dependencies

Add these to your HTML `<head>`:

```html
<!-- Tailwind CSS -->
<script src="https://cdn.tailwindcss.com"></script>

<!-- OR use your own Tailwind build -->
<link rel="stylesheet" href="/css/components.css">

<!-- HTMX -->
<script src="https://unpkg.com/htmx.org@1.9.10"></script>

<!-- Alpine.js -->
<script src="https://unpkg.com/alpinejs@3.13.3" defer></script>

<!-- Font Awesome (optional, for icons) -->
<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.5.0/css/all.min.css">
```

### 3. Copy Component Files

Copy the `ui/` directory to your project:

```
your-project/
├── views/
│   └── ui/
│       ├── button.templ
│       ├── input.templ
│       ├── form-controls.templ
│       ├── components.templ
│       ├── field-renderer.templ
│       ├── advanced-fields.templ
│       ├── schema-renderer.templ
│       └── overlays.templ
└── styles/
    └── components.css
```

### 4. Generate Templ Code

```bash
cd views/ui
templ generate
```

## 🚀 Quick Start

### Example 1: Simple Form

```go
package views

import "github.com/niiniyare/erp/views/ui"

templ ContactForm() {
    @ui.Form(ui.FormProps{
        Action: "/api/contact",
        Method: "POST",
        Class:  "space-y-4",
    }) {
        @ui.FormField(ui.FormFieldProps{
            Label:    "Email",
            Name:     "email",
            ID:       "email",
            Required: true,
        }) {
            @ui.Input(ui.InputProps{
                Type:        ui.InputTypeEmail,
                Name:        "email",
                ID:          "email",
                Placeholder: "john@example.com",
                Required:    true,
            })
        }
        
        @ui.Button(ui.ButtonProps{
            Type:    "submit",
            Variant: ui.ButtonVariantDefault,
        }) {
            <span>Submit</span>
        }
    }
}
```

### Example 2: Schema-Driven Form

```go
// In your handler
func HandleForm(c *fiber.Ctx) error {
    // Load schema
    schema, _ := registry.Get(c.Context(), "user-registration")
    
    // Enrich with permissions
    enriched, _ := enricher.Enrich(c.Context(), schema, user)
    
    // Render
    return views.RegistrationForm(enriched, nil, nil).Render(
        c.Context(),
        c.Response().BodyWriter(),
    )
}
```

```go
// In your view
templ RegistrationForm(schema *schema.Schema, data map[string]any, errors map[string][]string) {
    <!DOCTYPE html>
    <html>
    <head>
        <title>{ schema.Title }</title>
        <link rel="stylesheet" href="/css/components.css">
        <script src="https://unpkg.com/htmx.org@1.9.10"></script>
        <script src="https://unpkg.com/alpinejs@3.13.3" defer></script>
    </head>
    <body>
        <div class="container mx-auto max-w-2xl py-12">
            @ui.SchemaForm(ui.SchemaFormProps{
                Schema: schema,
                Data:   data,
                Errors: errors,
            })
        </div>
    </body>
    </html>
}
```

## 📚 Component Documentation

### Base Components

#### Button

```go
@ui.Button(ui.ButtonProps{
    Type:    "button",
    Variant: ui.ButtonVariantDefault, // default, destructive, outline, secondary, ghost, link
    Size:    ui.ButtonSizeDefault,    // default, sm, lg, icon
    Disabled: false,
}) {
    <span>Click Me</span>
}
```

#### Input

```go
@ui.Input(ui.InputProps{
    Type:        ui.InputTypeText,
    Name:        "username",
    ID:          "username",
    Placeholder: "Enter username",
    Required:    true,
    MinLength:   3,
    MaxLength:   50,
})
```

#### Select

```go
@ui.Select(ui.SelectProps{
    Name: "country",
    ID:   "country",
    Options: []ui.SelectOption{
        {Value: "us", Label: "United States"},
        {Value: "uk", Label: "United Kingdom"},
        {Value: "ca", Label: "Canada"},
    },
    Placeholder: "Select a country",
})
```

### Form Components

#### FormField (with Label, Description, Error)

```go
@ui.FormField(ui.FormFieldProps{
    Label:       "Email Address",
    Name:        "email",
    ID:          "email",
    Required:    true,
    Description: "We'll never share your email",
    Error:       "Invalid email address",
}) {
    @ui.Input(ui.InputProps{
        Type: ui.InputTypeEmail,
        Name: "email",
        ID:   "email",
    })
}
```

#### Textarea

```go
@ui.Textarea(ui.TextareaProps{
    Name:        "description",
    ID:          "description",
    Placeholder: "Enter description",
    Rows:        4,
    MaxLength:   500,
})
```

#### Checkbox

```go
@ui.CheckboxWithLabel(ui.CheckboxProps{
    Name:    "terms",
    ID:      "terms",
    Checked: false,
}, "I agree to the terms and conditions")
```

#### Radio Group

```go
@ui.RadioGroup("payment_method", []ui.SelectOption{
    {Value: "card", Label: "Credit Card"},
    {Value: "paypal", Label: "PayPal"},
    {Value: "bank", Label: "Bank Transfer"},
}, "card", false) // false = vertical, true = horizontal
```

#### Switch

```go
@ui.SwitchWithLabel(ui.SwitchProps{
    Name:    "notifications",
    ID:      "notifications",
    Checked: true,
}, "Enable notifications")
```

### Advanced Components

#### File Upload

```go
@ui.FileFieldRenderer(&schema.Field{
    Name:        "document",
    Type:        schema.FieldTypeFile,
    Label:       "Upload Document",
    Config: map[string]any{
        "accept":   ".pdf,.docx",
        "maxSize":  5242880, // 5MB
        "multiple": false,
    },
}, nil, nil)
```

#### Image Upload with Preview

```go
@ui.ImageFieldRenderer(&schema.Field{
    Name:  "avatar",
    Type:  schema.FieldTypeImage,
    Label: "Profile Picture",
}, nil, nil)
```

#### Tags Input

```go
@ui.TagsFieldRenderer(&schema.Field{
    Name:        "skills",
    Type:        schema.FieldTypeTags,
    Label:       "Skills",
    Placeholder: "Add a skill",
    Config: map[string]any{
        "maxTags": 10,
    },
}, nil, nil)
```

#### Rating

```go
@ui.RatingFieldRenderer(&schema.Field{
    Name:  "rating",
    Type:  schema.FieldTypeRating,
    Label: "Rate this product",
    Config: map[string]any{
        "max": 5,
    },
}, nil, nil)
```

#### Slider

```go
@ui.SliderFieldRenderer(&schema.Field{
    Name:  "volume",
    Type:  schema.FieldTypeSlider,
    Label: "Volume",
    Config: map[string]any{
        "min":  0,
        "max":  100,
        "step": 1,
    },
}, 50, nil)
```

### Layout Components

#### Card

```go
@ui.Card(ui.CardProps{}) {
    @ui.CardHeader() {
        @ui.CardTitle("Card Title", "")
        @ui.CardDescription("This is a description")
    }
    @ui.CardContent() {
        <p>Card content goes here</p>
    }
    @ui.CardFooter() {
        @ui.Button(ui.ButtonProps{Type: "button"}) {
            <span>Action</span>
        }
    }
}
```

#### Grid

```go
@ui.Grid(ui.GridProps{
    Cols: 3, // 1, 2, 3, 4, 6
    Gap:  "gap-4",
}) {
    <div>Column 1</div>
    <div>Column 2</div>
    <div>Column 3</div>
}
```

#### Stack (Vertical/Horizontal)

```go
// Vertical
@ui.Stack("space-y-4") {
    <div>Item 1</div>
    <div>Item 2</div>
}

// Horizontal
@ui.HStack("space-x-4") {
    <div>Item 1</div>
    <div>Item 2</div>
}
```

### Feedback Components

#### Alert

```go
@ui.Alert(ui.AlertProps{
    Variant:     ui.AlertVariantSuccess,
    Title:       "Success!",
    Description: "Your changes have been saved.",
    Icon:        "fa fa-check-circle",
    Dismissible: true,
})
```

#### Badge

```go
@ui.Badge(ui.BadgeProps{
    Text:    "New",
    Variant: ui.BadgeVariantDefault,
    Icon:    "fa fa-star",
})
```

#### Progress Bar

```go
@ui.Progress(75, 100, true) // value, max, showLabel
```

#### Loading Spinner

```go
@ui.LoadingSpinner("Loading data...")
```

#### Skeleton

```go
@ui.SkeletonCard()
@ui.SkeletonText(3) // 3 lines
```

### Overlay Components

#### Modal

```go
// Modal definition
@ui.Modal(ui.ModalProps{
    ID:          "my-modal",
    Title:       "Modal Title",
    Description: "Modal description",
    Size:        "md",
    Dismissible: true,
}) {
    <p>Modal content</p>
    @ui.ModalFooter() {
        @ui.Button(ui.ButtonProps{Type: "button"}) {
            <span x-on:click="$dispatch('close-modal', { id: 'my-modal' })">Close</span>
        }
    }
}

// Trigger button
@ui.Button(ui.ButtonProps{Type: "button"}) {
    <span x-on:click="$dispatch('open-modal', { id: 'my-modal' })">Open Modal</span>
}
```

#### Dropdown Menu

```go
@ui.Dropdown(ui.DropdownProps{
    ID:        "user-menu",
    Trigger:   "Options",
    Placement: "bottom",
}) {
    @ui.DropdownItem("fa fa-user", "Profile", "console.log('Profile')")
    @ui.DropdownItem("fa fa-cog", "Settings", "console.log('Settings')")
    @ui.DropdownSeparator()
    @ui.DropdownItem("fa fa-sign-out", "Logout", "console.log('Logout')")
}
```

#### Tooltip

```go
@ui.Tooltip("This is a helpful tooltip", "top") {
    @ui.Button(ui.ButtonProps{Type: "button"}) {
        <span>Hover me</span>
    }
}
```

### Data Display

#### Table

```go
@ui.Table(ui.TableProps{
    Columns: []ui.TableColumn{
        {Name: "id", Label: "ID", Type: "text", Sortable: true},
        {Name: "name", Label: "Name", Type: "text", Sortable: true},
        {Name: "email", Label: "Email", Type: "text"},
        {Name: "status", Label: "Status", Type: "badge"},
    },
    Data: []map[string]any{
        {"id": "1", "name": "John Doe", "email": "john@example.com", "status": "Active"},
        {"id": "2", "name": "Jane Smith", "email": "jane@example.com", "status": "Inactive"},
    },
    Selectable: true,
    Sortable:   true,
})
```

#### CRUD Table

```go
@ui.CRUDTable(schema, data, totalCount)
```

## 🎨 Customization

### Tailwind Configuration

Add this to your `tailwind.config.js`:

```javascript
module.exports = {
  darkMode: 'class',
  content: [
    './views/**/*.{templ,go}',
  ],
  theme: {
    extend: {
      colors: {
        border: 'hsl(var(--border))',
        input: 'hsl(var(--input))',
        ring: 'hsl(var(--ring))',
        background: 'hsl(var(--background))',
        foreground: 'hsl(var(--foreground))',
        primary: {
          DEFAULT: 'hsl(var(--primary))',
          foreground: 'hsl(var(--primary-foreground))',
        },
        // ... add more colors from design-tokens.md
      },
      borderRadius: {
        lg: 'var(--radius)',
        md: 'calc(var(--radius) - 2px)',
        sm: 'calc(var(--radius) - 4px)',
      },
    },
  },
}
```

### Custom Styling

Override component styles in your CSS:

```css
.btn-default {
  @apply bg-blue-600 text-white hover:bg-blue-700;
}

.input {
  @apply border-2 border-blue-300 focus:border-blue-500;
}
```

## 📖 Schema Field Types

All field types from your schema documentation are supported:

- **Basic**: text, email, password, number, phone, url, hidden, color
- **Date/Time**: date, time, datetime, daterange, month, year
- **Selection**: select, radio, checkbox, checkboxes, switch, tree-select
- **File**: file, image
- **Text**: textarea, richtext, code, json
- **Specialized**: currency, tags, rating, slider, icon-picker
- **Relational**: relation, autocomplete
- **Layout**: repeatable, group, tabs, fieldset
- **Display**: static, divider, formula

## 🔒 Security Best Practices

1. **Always validate on server-side** - Client-side validation is for UX only
2. **Sanitize user input** - Use Go's html/template or sanitize libraries
3. **Use CSRF tokens** - Protect against cross-site request forgery
4. **Implement rate limiting** - Prevent abuse
5. **Escape output** - Templ handles this automatically
6. **Use HTTPS** - Always in production

## ♿ Accessibility

All components follow WCAG 2.1 AA guidelines:

- Proper ARIA attributes
- Keyboard navigation support
- Screen reader friendly
- Focus management
- Color contrast compliance
- Semantic HTML

## 📝 Examples

See the `examples/` directory for complete working examples:

- `ContactFormPage` - Simple contact form
- `RegistrationFormPage` - Schema-driven registration
- `UsersTablePage` - CRUD table with pagination
- `ProductFormPage` - Complex form with all field types
- `DashboardPage` - Dashboard with cards and stats
- `ModalsExamplePage` - Modal and dialog usage

## 🤝 Contributing

Contributions are welcome! Please follow these guidelines:

1. Follow Go and Templ best practices
2. Maintain accessibility standards
3. Add tests for new components
4. Update documentation
5. Follow the existing code style

## 📄 License

MIT License - see LICENSE file for details

## 🙏 Credits

- Inspired by [shadcn/ui](https://ui.shadcn.com/)
- Built with [Templ](https://templ.guide/)
- Powered by [HTMX](https://htmx.org/)
- Enhanced with [Alpine.js](https://alpinejs.dev/)
- Styled with [Tailwind CSS](https://tailwindcss.com/)

## 📞 Support

For issues and questions:
- GitHub Issues: [your-repo/issues](https://github.com/your-repo/issues)
- Documentation: [your-docs-site.com](https://your-docs-site.com)
- Email: support@yoursite.com

---

Made with ❤️ for the Awo ERP system
