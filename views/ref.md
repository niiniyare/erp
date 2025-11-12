# ERP UI Component System Reference

## Overview

This document provides a comprehensive reference for the ERP UI component system built with **Templ + HTMX + Alpine.js** using **Atomic Design** patterns. The system is designed to work seamlessly with schema-driven forms and provides type-safe, server-rendered components.

## Architecture & Design Choices

### Technology Stack
- **Templ**: Type-safe Go templating with component-based architecture
- **HTMX**: Server-side rendering with dynamic interactions
- **Alpine.js**: Lightweight client-side reactivity and state management
- **Tailwind CSS**: Utility-first styling with design tokens
- **Atomic Design**: Hierarchical component organization

### Key Design Decisions

1. **Schema-First Approach**: Components are designed to work with schema definitions
2. **Type Safety**: Go's type system ensures compile-time safety
3. **Server-Side Rendering**: Better SEO and initial load performance
4. **Progressive Enhancement**: Works without JavaScript, enhanced with it
5. **Component Composability**: Small, reusable components that compose into larger ones

## File Structure

```
views/
├── components/
│   ├── atoms/              # Basic building blocks
│   │   ├── button.templ    # Button variations
│   │   ├── input.templ     # Input field types
│   │   ├── icon.templ      # Icon system
│   │   ├── badge.templ     # Status badges
│   │   └── tags.templ      # Tag components
│   ├── molecules/          # Component combinations
│   │   ├── formfield.templ # Complete form fields
│   │   ├── searchbox.templ # Search functionality
│   │   ├── menuitem.templ  # Navigation items
│   │   ├── validation.templ # Form validation
│   │   └── errorhandling.templ # Error management
│   ├── organisms/          # Complex components
│   │   ├── form.templ      # Complete forms
│   │   ├── table.templ     # Data tables
│   │   ├── navigation.templ # Navigation systems
│   │   └── modal.templ     # Modal dialogs
│   └── templates/          # Layout templates
│       ├── pagelayout.templ # Page structures
│       └── dashboardlayout.templ # Dashboard layouts
├── pages/                  # Complete pages
│   ├── demo.templ         # Component showcase
│   └── simple_demo.templ  # Simplified demo
├── static/                # Static assets
│   ├── css/
│   │   ├── tokens.css     # Design tokens
│   │   └── base.css       # Base styles
│   └── js/               # Additional scripts
└── ref.md                # This reference file
```

## Component Hierarchy (Atomic Design)

### Atoms (Basic Building Blocks)
Located in `views/components/atoms/`

#### Button Component (`button.templ`)
**Purpose**: Reusable button with various styles and states

**Props Structure**:
```go
type ButtonProps struct {
    Type         string           // "button", "submit", "reset"
    Variant      ButtonVariant    // Primary, Secondary, Outline, etc.
    Size         ButtonSize       // SM, MD, LG
    Icon         string           // Icon name
    IconLeft     string           // Left side icon
    IconRight    string           // Right side icon
    Loading      bool             // Show loading state
    Disabled     bool             // Disable interaction
    Class        string           // Additional CSS classes
    ID           string           // HTML ID attribute
    // HTMX attributes
    HXPost       string           // HTMX POST endpoint
    HXGet        string           // HTMX GET endpoint
    HXTarget     string           // HTMX target selector
    HXSwap       string           // HTMX swap strategy
    // Alpine.js attributes
    AlpineClick  string           // Alpine click handler
}
```

**Key Functions**:
- `buttonClasses(props ButtonProps)`: Generates Tailwind CSS classes
- `getButtonIcon(props)`: Determines which icon to display
- `getLoadingIcon()`: Returns loading spinner component

**Usage**:
```go
@atoms.Button(atoms.ButtonProps{
    Variant: atoms.ButtonPrimary,
    Size:    atoms.ButtonSizeMD,
    Icon:    "save",
}) {
    Save Changes
}
```

#### Input Component (`input.templ`)
**Purpose**: Various input field types with consistent styling

**Key Functions**:
- `TextInput(props InputProps)`: Standard text input
- `EmailInput(props InputProps)`: Email validation input
- `PasswordInput(props InputProps)`: Password field with visibility toggle
- `NumberInput(props InputProps)`: Numeric input with validation
- `SearchInput(props InputProps)`: Search field with clear button

**Schema Integration**: Maps to schema field types automatically

#### Icon Component (`icon.templ`)
**Purpose**: Consistent icon system with size variants

**Props**:
```go
type IconProps struct {
    Name  string    // Icon identifier
    Size  IconSize  // XS, SM, MD, LG, XL
    Class string    // Additional classes
}
```

**Design Choice**: Uses string-based icon names that map to external icon libraries or SVGs.

### Molecules (Component Combinations)
Located in `views/components/molecules/`

#### FormField Component (`formfield.templ`)
**Purpose**: Complete form field with label, input, validation, and help text

**Key Features**:
- Automatic validation state styling
- Error message display
- Help text support
- Required field indicators
- Multiple input types support

**Schema Integration**: 
```go
type FormFieldProps struct {
    Type        FormFieldType    // Maps to schema field type
    Size        FormFieldSize    
    ID          string          // Maps to schema field name
    Name        string          
    Label       string          // From schema field label
    Value       string          // Current field value
    Placeholder string          // From schema placeholder
    HelpText    string          // From schema description
    ErrorText   string          // Validation error message
    Required    bool            // From schema required flag
    // Input-specific props from schema
    MinLength    int            
    MaxLength    int            
    Min          string         
    Max          string         
    Pattern      string         // Validation pattern
    Options      []SelectOption // For select fields
    // HTMX validation
    HXPost       string         // Validation endpoint
    HXTrigger    string         // When to validate
}
```

**Key Functions**:
- `FormField(props FormFieldProps)`: Main component
- `textareaInput(props)`: Textarea variant
- `selectInput(props)`: Dropdown select
- `tagsInput(props)`: Tag selection field

#### SearchBox Component (`searchbox.templ`)
**Purpose**: Search functionality with filters and suggestions

**Features**:
- Real-time search with HTMX
- Filter options
- Search history
- Result previews

#### Validation Component (`validation.templ`)
**Purpose**: Form validation logic and error display

**Key Functions**:
- `ValidationMessage(props)`: Individual error messages
- `ValidationSummary(props)`: Form-level error summary
- `FieldValidator(props)`: Real-time field validation

**Schema Integration**: Uses validation rules from schema definitions.

### Organisms (Complex Components)
Located in `views/components/organisms/`

#### Form Component (`form.templ`)
**Purpose**: Complete form with sections, fields, and actions

**Schema Integration**: 
- Automatically generates fields from schema
- Handles form layout based on schema configuration
- Supports nested schemas and sections
- Validates using schema rules

**Props Structure**:
```go
type FormProps struct {
    Layout       FormLayout      // Vertical, Horizontal, Grid, Inline
    Size         FormSize        
    Title        string          
    Description  string          
    Method       string          // HTTP method
    Action       string          // Form action URL
    Sections     []FormSection   // Form sections from schema
    Fields       []molecules.FormFieldProps // Direct fields
    Actions      []FormAction    // Form buttons
    ShowProgress bool            // Multi-step form progress
    // Schema integration
    Schema       interface{}     // JSON Schema definition
    Data         map[string]interface{} // Form data
    Errors       map[string]string      // Validation errors
}
```

**Key Functions**:
- `Form(props FormProps)`: Main form component
- `SimpleForm(props)`: Basic vertical form
- `GridForm(props)`: Multi-column grid layout
- `MultiStepForm(props)`: Wizard-style form
- `formSection(section, stepNumber, formProps)`: Section renderer

#### Table Component (`table.templ`)
**Purpose**: Data table with sorting, filtering, pagination, and actions

**Schema Integration**: 
- Column definitions from schema
- Data type handling
- Search and filter configuration
- Export capabilities

**Key Features**:
- Row selection with bulk actions
- Column sorting and filtering
- Responsive design
- Loading and error states
- Pagination
- Custom cell renderers

**Props Structure**:
```go
type TableProps struct {
    Variant         TableVariant    // Default, Bordered, Striped, Hover
    Size            TableSize       
    Columns         []TableColumn   // Column definitions
    Rows            []TableRow      // Data rows
    Actions         []TableAction   // Table-level actions
    // Features
    Selectable      bool           
    Sortable        bool           
    Searchable      bool           
    Filterable      bool           
    Paginated       bool           
    // State
    CurrentPage     int            
    TotalPages      int            
    SelectedRows    []string       
    SortColumn      string         
    SortDirection   SortDirection  
    SearchQuery     string         
    Filters         map[string]string
}
```

#### Navigation Component (`navigation.templ`)
**Purpose**: Navigation systems (sidebar, topbar, breadcrumbs, tabs, steps)

**Types**:
- **Sidebar**: Collapsible side navigation
- **Topbar**: Horizontal navigation bar
- **Breadcrumb**: Page hierarchy navigation
- **Tabs**: Content switching tabs
- **Steps**: Multi-step process navigation

**Key Functions**:
- `Navigation(props NavigationProps)`: Main router component
- `sidebarNavigation(props)`: Sidebar implementation
- `topbarNavigation(props)`: Topbar implementation
- `breadcrumbNavigation(props)`: Breadcrumb chain
- `tabsNavigation(props)`: Tab system
- `stepsNavigation(props)`: Step indicator

### Templates (Layout Templates)
Located in `views/components/templates/`

#### PageLayout Component (`pagelayout.templ`)
**Purpose**: Complete page layout with navigation, header, and content areas

**Layout Types**:
- **Sidebar**: Traditional admin layout
- **Topbar**: Modern header-based layout
- **Combined**: Both sidebar and topbar
- **Fullscreen**: No navigation (for modals, etc.)

**Props Structure**:
```go
type PageLayoutProps struct {
    Type        PageLayoutType  
    Meta        PageMeta        // SEO and head tags
    Header      PageHeader      // Page header section
    Sidebar     PageSidebar     // Sidebar configuration
    Topbar      PageTopbar      // Topbar configuration
    Footer      PageFooter      // Footer content
    // Features
    HXBoost     bool           // Enable HTMX boost
    Class       string         
}
```

## Schema Integration

### How Components Work with Schemas

1. **Schema Definition**: JSON Schema defines form structure
2. **Code Generation**: Schemas can generate Go structs
3. **Component Mapping**: Components automatically map to schema types
4. **Validation**: Schema validation rules applied to components
5. **Data Binding**: Form data bound to schema-defined structures

### Example Schema Integration:

```go
// Schema definition
type UserSchema struct {
    Name     string `json:"name" validate:"required,min=2,max=100"`
    Email    string `json:"email" validate:"required,email"`
    Role     string `json:"role" validate:"required,oneof=admin user manager"`
    Settings struct {
        Theme        string `json:"theme" validate:"oneof=light dark auto"`
        Notifications bool   `json:"notifications"`
    } `json:"settings"`
}

// Auto-generated form
func UserForm(schema UserSchema, data map[string]interface{}) templ.Component {
    return organisms.Form(organisms.FormProps{
        Title: "User Settings",
        Sections: []organisms.FormSection{
            {
                Title: "Basic Information",
                Fields: []molecules.FormFieldProps{
                    {
                        Type:     molecules.FormFieldText,
                        Name:     "name",
                        Label:    "Full Name",
                        Required: true,
                        Value:    data["name"].(string),
                    },
                    {
                        Type:     molecules.FormFieldEmail,
                        Name:     "email", 
                        Label:    "Email Address",
                        Required: true,
                        Value:    data["email"].(string),
                    },
                },
            },
        },
    })
}
```

## Key Functions Reference

### Utility Functions

#### Class Generation
All components have class generation functions:
```go
func buttonClasses(props ButtonProps) string
func inputClasses(props InputProps) string  
func tableClasses(props TableProps) string
func formClasses(props FormProps) string
```

**Purpose**: Generate Tailwind CSS classes based on component props and state.

#### State Management (Alpine.js)
```go
func getFormAlpineData(props FormProps) string
func getTableAlpineData(props TableProps) string
func getNavigationAlpineData(collapsed bool) string
```

**Purpose**: Generate Alpine.js data objects for client-side reactivity.

#### Validation
```go
func validateField(fieldName string, value interface{}, rules ValidationRules) []string
func validateForm(data map[string]interface{}, schema Schema) map[string][]string
```

**Purpose**: Server-side validation using schema rules.

### Helper Functions

#### Type Conversions
```go
func getStringValue(value interface{}) string
func getBoolValue(value interface{}) bool
func getIntValue(value interface{}) int
```

#### Conditional Rendering
```go
func getButtonVariant(variant ButtonVariant) ButtonVariant
func getInputSize(size InputSize) InputSize
func getErrorIcon(errorType ErrorType) string
```

**Purpose**: Provide default values and handle nil/empty cases.

## Usage Patterns

### 1. Schema-Driven Form
```go
// Define schema
type ProductSchema struct {
    Name        string  `json:"name" validate:"required"`
    Price       float64 `json:"price" validate:"required,min=0"`
    Category    string  `json:"category" validate:"required"`
    Description string  `json:"description"`
}

// Generate form from schema
@organisms.Form(organisms.FormProps{
    Schema: ProductSchema{},
    Layout: organisms.FormLayoutVertical,
    Fields: generateFieldsFromSchema(ProductSchema{}),
    Actions: []organisms.FormAction{
        {Text: "Save", Type: "submit", Variant: atoms.ButtonPrimary},
        {Text: "Cancel", Type: "button", Variant: atoms.ButtonSecondary},
    },
})
```

### 2. Data Table with Schema
```go
// Define table from schema
@organisms.Table(organisms.TableProps{
    Columns: generateColumnsFromSchema(ProductSchema{}),
    Rows: convertDataToRows(products),
    Searchable: true,
    Sortable: true,
    Paginated: true,
})
```

### 3. Dynamic Navigation
```go
// Navigation from configuration
@organisms.Navigation(organisms.NavigationProps{
    Type: organisms.NavigationSidebar,
    Items: generateNavigationFromConfig(userRole),
    Collapsible: true,
})
```

## Best Practices

### 1. Component Composition
- Always compose larger components from smaller ones
- Use props to configure behavior, not hardcode values
- Keep components focused on single responsibilities

### 2. Schema Integration  
- Define schemas first, then generate components
- Use validation rules from schemas
- Keep data binding declarative

### 3. Performance
- Use HTMX for server interactions
- Minimize Alpine.js state
- Lazy load components when possible

### 4. Accessibility
- Include ARIA attributes
- Use semantic HTML elements
- Provide keyboard navigation

### 5. Styling
- Use design tokens for consistency
- Follow Tailwind utility patterns
- Support theme customization

## Error Handling

### Client-Side Errors
```go
@molecules.ErrorBoundary(molecules.ErrorBoundaryProps{
    Title: "Form Error",
    Errors: validationErrors,
    Recoverable: true,
    OnRetry: "submitForm()",
})
```

### Server-Side Validation
```go
// Validation middleware
func validateRequest(schema Schema) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            data := parseRequestData(c)
            errors := validateAgainstSchema(data, schema)
            if len(errors) > 0 {
                return renderFormWithErrors(c, data, errors)
            }
            return next(c)
        }
    }
}
```

## Testing Components

### Unit Testing
```go
func TestButtonComponent(t *testing.T) {
    props := atoms.ButtonProps{
        Variant: atoms.ButtonPrimary,
        Size:    atoms.ButtonSizeMD,
    }
    
    component := atoms.Button(props)
    html := renderComponentToString(component)
    
    assert.Contains(t, html, "btn-primary")
    assert.Contains(t, html, "btn-md")
}
```

### Integration Testing
```go
func TestFormSubmission(t *testing.T) {
    form := organisms.Form(organisms.FormProps{
        Method: "POST",
        Action: "/api/users",
        Fields: userFormFields,
    })
    
    // Test form rendering and submission
}
```

## Migration Guide

### From HTML Forms
1. Replace `<form>` with `@organisms.Form`
2. Convert inputs to `@molecules.FormField`
3. Add validation rules
4. Configure HTMX endpoints

### From React Components
1. Convert JSX to Templ syntax
2. Move state to Alpine.js
3. Replace API calls with HTMX
4. Update CSS classes to Tailwind

## Troubleshooting

### Common Issues

1. **Alpine.js not working**: Check script loading order
2. **CSS not applied**: Verify Tailwind configuration
3. **HTMX requests failing**: Check server endpoints
4. **Validation not working**: Verify schema rules

### Debug Tools
- Browser dev tools for client-side issues
- Server logs for HTMX requests
- Component props debugging
- CSS inspector for styling issues

---

This reference covers the complete UI component system. For specific implementation details, refer to the individual component files in the `views/components/` directory.