# Component Index & Quick Reference

Complete reference for all UI components in the Awo ERP Component Library.

## 📦 Component Files

### Core Files
- `button.templ` - Button components with variants
- `input.templ` - Input fields for all types
- `form-controls.templ` - Label, Textarea, Select, Checkbox, Radio, Switch
- `components.templ` - Card, Alert, Badge, Form, Grid, Loading, Skeleton
- `field-renderer.templ` - Schema-driven field renderer
- `advanced-fields.templ` - File, Tags, Rating, Slider, Repeatable fields
- `schema-renderer.templ` - Complete form and table renderer
- `overlays.templ` - Modal, Dropdown, Toast, Popover, Tooltip, Progress

### Style Files
- `components.css` - All component styles and utilities

## 🎯 Quick Reference by Use Case

### Creating a Simple Form
```go
@ui.Form() {
    @ui.FormField() { @ui.Input() }
    @ui.Button() { <span>Submit</span> }
}
```

### Creating a Schema-Driven Form
```go
@ui.SchemaForm(ui.SchemaFormProps{
    Schema: schema,
    Data:   data,
    Errors: errors,
})
```

### Creating a Data Table
```go
@ui.Table(ui.TableProps{
    Columns: columns,
    Data:    data,
})
```

### Creating a CRUD Interface
```go
@ui.CRUDTable(schema, data, total)
```

### Creating a Card
```go
@ui.Card() {
    @ui.CardHeader() {
        @ui.CardTitle("Title", "")
    }
    @ui.CardContent() {
        // Content
    }
}
```

### Creating a Modal
```go
@ui.Modal(ui.ModalProps{ID: "my-modal"}) {
    // Content
    @ui.ModalFooter() {
        // Buttons
    }
}
```

### Showing Alerts
```go
@ui.Alert(ui.AlertProps{
    Variant: ui.AlertVariantSuccess,
    Title:   "Success!",
})
```

## 📊 Component Matrix

| Component | File | Props | Children | HTMX | Alpine.js |
|-----------|------|-------|----------|------|-----------|
| Button | button.templ | ButtonProps | ✅ | ✅ | ✅ |
| Input | input.templ | InputProps | ❌ | ✅ | ✅ |
| Label | form-controls.templ | LabelProps | ❌ | ❌ | ❌ |
| Textarea | form-controls.templ | TextareaProps | ❌ | ✅ | ✅ |
| Select | form-controls.templ | SelectProps | ❌ | ✅ | ✅ |
| Checkbox | form-controls.templ | CheckboxProps | ❌ | ❌ | ✅ |
| Radio | form-controls.templ | RadioProps | ❌ | ❌ | ✅ |
| Switch | form-controls.templ | SwitchProps | ❌ | ❌ | ✅ |
| Card | components.templ | CardProps | ✅ | ❌ | ❌ |
| Alert | components.templ | AlertProps | ❌ | ❌ | ✅ |
| Badge | components.templ | BadgeProps | ❌ | ❌ | ❌ |
| FormField | components.templ | FormFieldProps | ✅ | ❌ | ❌ |
| Form | components.templ | FormProps | ✅ | ✅ | ❌ |
| Grid | components.templ | GridProps | ✅ | ❌ | ❌ |
| Modal | overlays.templ | ModalProps | ✅ | ❌ | ✅ |
| Dropdown | overlays.templ | DropdownProps | ✅ | ❌ | ✅ |
| Toast | overlays.templ | AlertVariant | ❌ | ❌ | ✅ |
| Tooltip | overlays.templ | string | ✅ | ❌ | ✅ |
| Progress | overlays.templ | int | ❌ | ❌ | ✅ |
| Table | schema-renderer.templ | TableProps | ❌ | ❌ | ❌ |
| CRUDTable | schema-renderer.templ | Schema | ❌ | ✅ | ✅ |
| SchemaForm | schema-renderer.templ | SchemaFormProps | ❌ | ✅ | ✅ |
| FieldRenderer | field-renderer.templ | Field | ❌ | ✅ | ✅ |

## 🎨 Variant Options

### Button Variants
- `ButtonVariantDefault` - Primary action
- `ButtonVariantDestructive` - Dangerous action
- `ButtonVariantOutline` - Secondary action
- `ButtonVariantSecondary` - Alternative action
- `ButtonVariantGhost` - Minimal style
- `ButtonVariantLink` - Link style

### Button Sizes
- `ButtonSizeDefault` - Standard size
- `ButtonSizeSm` - Small
- `ButtonSizeLg` - Large
- `ButtonSizeIcon` - Icon only

### Alert Variants
- `AlertVariantDefault` - Standard
- `AlertVariantDestructive` - Error
- `AlertVariantSuccess` - Success
- `AlertVariantWarning` - Warning
- `AlertVariantInfo` - Information

### Badge Variants
- `BadgeVariantDefault` - Primary
- `BadgeVariantSecondary` - Secondary
- `BadgeVariantDestructive` - Error
- `BadgeVariantOutline` - Outlined
- `BadgeVariantSuccess` - Success
- `BadgeVariantWarning` - Warning

### Input Types
- `InputTypeText` - Text input
- `InputTypeEmail` - Email input
- `InputTypePassword` - Password input
- `InputTypeNumber` - Number input
- `InputTypeTel` - Phone input
- `InputTypeURL` - URL input
- `InputTypeDate` - Date picker
- `InputTypeTime` - Time picker
- `InputTypeDatetime` - Datetime picker
- `InputTypeSearch` - Search input
- `InputTypeHidden` - Hidden field
- `InputTypeColor` - Color picker
- `InputTypeFile` - File upload

## 🔧 Common Patterns

### Form with Validation
```go
@ui.FormField(ui.FormFieldProps{
    Label:       "Email",
    Name:        "email",
    ID:          "email",
    Required:    true,
    Description: "Your email address",
    Error:       errors["email"][0], // Show first error
}) {
    @ui.Input(ui.InputProps{
        Type:        ui.InputTypeEmail,
        Name:        "email",
        ID:          "email",
        Required:    true,
        AriaInvalid: len(errors["email"]) > 0,
    })
}
```

### HTMX Form Submission
```go
@ui.Form(ui.FormProps{
    HxPost:   "/api/submit",
    HxTarget: "#result",
    HxSwap:   "innerHTML",
}) {
    // Fields
    @ui.Button(ui.ButtonProps{Type: "submit"}) {
        <span>Submit</span>
    }
}
<div id="result"></div>
```

### Conditional Field Visibility (Alpine.js)
```go
<div x-data="{ showAdvanced: false }">
    @ui.CheckboxWithLabel(ui.CheckboxProps{
        AlpineModel: "showAdvanced",
    }, "Show advanced options")
    
    <div x-show="showAdvanced" x-transition>
        // Advanced fields
    </div>
</div>
```

### Loading State
```go
<div x-data="{ loading: false }">
    @ui.LoadingButton(ui.ButtonProps{
        Type:    "submit",
        AlpineClick: "loading = true",
    }, "Save Changes", false)
    
    <div x-show="loading">
        @ui.LoadingSpinner("Saving...")
    </div>
</div>
```

### Responsive Grid
```go
@ui.Grid(ui.GridProps{
    Cols: 3, // Automatically responsive: 1 col mobile, 2 col tablet, 3 col desktop
}) {
    <div>Column 1</div>
    <div>Column 2</div>
    <div>Column 3</div>
}
```

### File Upload with Preview
```go
@ui.ImageFieldRenderer(&schema.Field{
    Name:  "avatar",
    Type:  schema.FieldTypeImage,
    Label: "Profile Picture",
    Config: map[string]any{
        "maxSize": 5242880, // 5MB
        "accept":  "image/*",
    },
}, currentValue, nil)
```

## 🎯 Schema Field Type Mapping

| Schema Field Type | Templ Component | Notes |
|-------------------|-----------------|-------|
| `text` | `Input(InputTypeText)` | Basic text input |
| `email` | `Input(InputTypeEmail)` | Email with validation |
| `password` | `PasswordInput` | With show/hide toggle |
| `number` | `NumberInput` | With optional stepper |
| `phone` | `Input(InputTypeTel)` | Phone number |
| `url` | `Input(InputTypeURL)` | URL with validation |
| `date` | `Input(InputTypeDate)` | Date picker |
| `time` | `Input(InputTypeTime)` | Time picker |
| `datetime` | `Input(InputTypeDatetime)` | Datetime picker |
| `textarea` | `Textarea` | Multi-line text |
| `select` | `Select` | Dropdown |
| `radio` | `RadioGroup` | Radio buttons |
| `checkbox` | `Checkbox` | Single checkbox |
| `checkboxes` | CheckboxGroup | Multiple checkboxes |
| `switch` | `Switch` | Toggle switch |
| `currency` | `CurrencyInput` | Money input |
| `file` | `FileFieldRenderer` | File upload |
| `image` | `ImageFieldRenderer` | Image upload with preview |
| `tags` | `TagsFieldRenderer` | Tag input |
| `rating` | `RatingFieldRenderer` | Star rating |
| `slider` | `SliderFieldRenderer` | Range slider |
| `repeatable` | `RepeatableFieldRenderer` | Dynamic field group |

## 🚀 Performance Tips

1. **Use HTMX** for partial updates instead of full page reloads
2. **Lazy load** large lists with pagination
3. **Debounce** search inputs with Alpine.js
4. **Cache** schema renderings on server
5. **Minimize** JavaScript with Alpine.js directives
6. **Optimize** images before upload
7. **Use** skeleton loaders for better perceived performance

## ✨ Customization Tips

1. **Override CSS** classes in your own stylesheet
2. **Extend props** by adding new properties to structs
3. **Create variants** by adding new enum values
4. **Compose components** to build new patterns
5. **Use Tailwind** utilities for quick styling
6. **Add Alpine.js** directives for interactivity

## 📚 Further Reading

- [Templ Documentation](https://templ.guide/)
- [HTMX Documentation](https://htmx.org/docs/)
- [Alpine.js Documentation](https://alpinejs.dev/start-here)
- [Tailwind CSS Documentation](https://tailwindcss.com/docs)
- [WCAG 2.1 Guidelines](https://www.w3.org/WAI/WCAG21/quickref/)

---

**Last Updated:** November 2024  
**Version:** 1.0.0
