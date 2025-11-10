# Schema Package UI Implementation Guide

**Version**: 1.0  
**Last Updated**: November 2025  
**Target Audience**: Frontend developers implementing UI frameworks (React, Templ, Go Templates, etc.)

## Table of Contents

1. [Overview & Architecture](#overview--architecture)
2. [Quick Start](#quick-start)
3. [Core Interfaces](#core-interfaces)
4. [Runtime System Integration](#runtime-system-integration)
5. [Builder API Reference](#builder-api-reference)
6. [Design Token System](#design-token-system)
7. [Validation Integration](#validation-integration)
8. [Internationalization (I18n)](#internationalization-i18n)
9. [State Management](#state-management)
10. [Event Handling](#event-handling)
11. [Conditional Logic](#conditional-logic)
12. [Theme System](#theme-system)
13. [Advanced Features](#advanced-features)
14. [Integration Patterns](#integration-patterns)
15. [Best Practices](#best-practices)
16. [Troubleshooting](#troubleshooting)

---

## Overview & Architecture

The ERP Schema Package is a comprehensive, UI-agnostic system for building dynamic, enterprise-grade forms and interfaces. It provides a JSON-driven approach where backend developers can define complete UIs without writing frontend code.

### Core Philosophy

- **Schema-Driven**: Everything is defined in JSON schemas
- **UI-Agnostic**: Works with any frontend framework
- **Runtime Dynamic**: Conditional logic, validation, and state management
- **Enterprise-Ready**: Multi-tenancy, permissions, audit trails
- **Type-Safe**: Strong typing with Go interfaces
- **Performance-Focused**: Multi-tier caching and optimizations

### System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    UI FRAMEWORK LAYER                       │
│     (React, Templ, Go Templates, etc.)                     │
└─────────────────────┬───────────────────────────────────────┘
                      │ Implements RuntimeRenderer
┌─────────────────────┴───────────────────────────────────────┐
│                    RUNTIME SYSTEM                           │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │    State    │ │  Events     │ │ Conditional │           │
│  │ Management  │ │ & Lifecycle │ │   Engine    │           │
│  └─────────────┘ └─────────────┘ └─────────────┘           │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────┴───────────────────────────────────────┐
│                     CORE SCHEMA SYSTEM                      │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │   Schema    │ │ Validation  │ │  Design     │           │
│  │ Definition  │ │   System    │ │  Tokens     │           │
│  │             │ │             │ │             │           │
│  │  ┌───────┐  │ │ ┌─────────┐ │ │ ┌─────────┐ │           │
│  │  │Fields │  │ │ │Validator│ │ │ │ Themes  │ │           │
│  │  │Actions│  │ │ │Registry │ │ │ │ Tokens  │ │           │
│  │  │Layout │  │ │ │Rules    │ │ │ │Components│ │           │
│  │  └───────┘  │ │ └─────────┘ │ │ └─────────┘ │           │
│  └─────────────┘ └─────────────┘ └─────────────┘           │
└─────────────────────────────────────────────────────────────┘
```

### Package Structure

```
pkg/schema/
├── Core Schema System
│   ├── schema.go              # Main Schema struct (1,333 lines)
│   ├── field.go               # Field system (1,752 lines)
│   ├── action.go              # Actions/buttons
│   ├── layout.go              # Layout system
│   ├── builder.go             # Fluent API
│   └── runtime.go             # Runtime interfaces
│
├── Runtime Subsystems
│   ├── runtime/               # Runtime execution
│   │   ├── runtime.go         # Main orchestrator (950 lines)
│   │   ├── state.go           # State management (694 lines)
│   │   └── events.go          # Event system (685 lines)
│   │
│   ├── validate/              # Validation system
│   │   ├── validator.go       # Core validator (1,054 lines)
│   │   └── registry.go        # Custom validators
│   │
│   ├── registry/              # Schema storage
│   ├── enrich/                # Permission system
│   └── parse/                 # JSON parsing
│
├── Design System
│   ├── theme.go               # Theme management
│   ├── tokens.go              # Design tokens
│   └── i18n.go                # Internationalization
│
└── Enterprise Features
    ├── business_rules.go      # Business logic
    ├── enterprise.go          # Advanced features
    └── mixin.go               # Reusable components
```

---

## Quick Start

### 1. Basic Schema Creation

```go
// Create a simple user form schema
func CreateUserFormSchema() *schema.Schema {
    return schema.NewBuilder("user-form", schema.TypeForm, "Create User").
        WithDescription("Add a new user to the system").
        AddTextField("firstName", "First Name", true).
        AddTextField("lastName", "Last Name", true).
        AddEmailField("email", "Email Address", true).
        AddPasswordField("password", "Password", true).
        AddSelectField("role", "Role", true, []schema.Option{
            {Value: "admin", Label: "Administrator"},
            {Value: "user", Label: "User"},
            {Value: "guest", Label: "Guest"},
        }).
        AddSubmitButton("Create User").
        Build(context.Background())
}
```

### 2. Implement Runtime Renderer

```go
// Example implementation for Templ/HTML
type TemplRenderer struct {
    templates *template.Template
    theme     *schema.Theme
}

func (r *TemplRenderer) RenderField(
    ctx context.Context, 
    field *schema.Field, 
    value any, 
    errors []string, 
    touched, dirty bool,
) (string, error) {
    // Determine field component based on type
    component := r.getFieldComponent(field.Type)
    
    // Apply theme tokens
    styles := r.theme.GetComponentTokens(component)
    
    // Render with appropriate template
    data := FieldRenderData{
        Field:   field,
        Value:   value,
        Errors:  errors,
        Touched: touched,
        Dirty:   dirty,
        Styles:  styles,
    }
    
    return r.templates.ExecuteTemplate(ctx, component, data)
}
```

### 3. Initialize Runtime

```go
// Set up runtime with your implementations
func SetupRuntime() *runtime.Runtime {
    config := schema.DefaultRuntimeConfig()
    
    return runtime.NewRuntime(config).
        WithRenderer(&TemplRenderer{}).
        WithValidator(&CustomValidator{}).
        WithConditionalEngine(&ConditionalEngine{}).
        Build()
}
```

---

## Core Interfaces

### RuntimeRenderer Interface

**Purpose**: The main interface UI frameworks must implement to render schema components.

```go
type RuntimeRenderer interface {
    // Core rendering methods
    RenderField(ctx context.Context, field *Field, value any, errors []string, touched, dirty bool) (string, error)
    RenderForm(ctx context.Context, schema *Schema, state map[string]any, errors map[string][]string) (string, error)
    RenderAction(ctx context.Context, action *Action, enabled bool) (string, error)
    
    // Layout rendering
    RenderLayout(ctx context.Context, layout *Layout, fields []*Field, state map[string]any) (string, error)
    RenderSection(ctx context.Context, section *Section, fields []*Field, state map[string]any) (string, error)
    RenderTab(ctx context.Context, tab *Tab, fields []*Field, state map[string]any) (string, error)
    RenderStep(ctx context.Context, step *Step, fields []*Field, state map[string]any) (string, error)
    RenderGroup(ctx context.Context, group *Group, fields []*Field, state map[string]any) (string, error)
    
    // Error rendering
    RenderErrors(ctx context.Context, errors map[string][]string) (string, error)
}
```

### Implementation Strategy

1. **Field Type Mapping**: Map schema field types to your components
2. **Theme Integration**: Use design tokens for styling
3. **State Binding**: Connect field values to your state system
4. **Error Display**: Implement error rendering patterns
5. **Layout Support**: Handle grid, tabs, steps, sections

```go
// Example field type mapping
func (r *MyRenderer) getFieldComponent(fieldType schema.FieldType) string {
    mapping := map[schema.FieldType]string{
        schema.FieldText:        "text-input",
        schema.FieldEmail:       "email-input", 
        schema.FieldPassword:    "password-input",
        schema.FieldNumber:      "number-input",
        schema.FieldSelect:      "select-input",
        schema.FieldTextarea:    "textarea-input",
        schema.FieldDate:        "date-picker",
        schema.FieldCheckbox:    "checkbox-input",
        schema.FieldRadio:       "radio-group",
        schema.FieldFile:        "file-upload",
        // ... 40+ field types supported
    }
    return mapping[fieldType]
}
```

### RuntimeValidator Interface

```go
type RuntimeValidator interface {
    // Field validation
    ValidateField(ctx context.Context, field *Field, value any, allData map[string]any) []string
    ValidateAllFields(ctx context.Context, schema *Schema, data map[string]any) map[string][]string
    
    // Action validation  
    ValidateAction(ctx context.Context, action *Action, formState map[string]any) error
    
    // Business rules validation
    ValidateBusinessRules(ctx context.Context, rules []ValidationRule, allValues map[string]any) map[string][]string
    
    // Async validation
    ValidateFieldAsync(ctx context.Context, field *Field, value any, callback ValidationCallback) error
}
```

### RuntimeConditionalEngine Interface

```go
type RuntimeConditionalEngine interface {
    // Field conditional evaluation
    EvaluateFieldVisibility(ctx context.Context, field *Field, data map[string]any) (bool, error)
    EvaluateFieldRequired(ctx context.Context, field *Field, data map[string]any) (bool, error) 
    EvaluateFieldEditable(ctx context.Context, field *Field, data map[string]any) (bool, error)
    
    // Action conditional evaluation
    EvaluateActionConditions(ctx context.Context, action *Action, allValues map[string]any) (bool, error)
    
    // Layout conditional evaluation
    EvaluateLayoutConditions(ctx context.Context, layout *Layout, allValues map[string]any) (*LayoutConditionalResult, error)
    
    // Batch evaluations for performance
    EvaluateAllFieldConditions(ctx context.Context, schema *Schema, data map[string]any) (map[string]*FieldConditionalResult, error)
}
```

---

## Runtime System Integration

### Runtime Configuration

```go
type RuntimeConfig struct {
    EnableConditionals  bool             `json:"enableConditionals"`
    EnableValidation    bool             `json:"enableValidation"`
    ValidationTiming    ValidationTiming `json:"validationTiming"`
    EnableDebounce      bool             `json:"enableDebounce"`
    DebounceDelay       time.Duration    `json:"debounceDelay"`
    EnableEventTracking bool             `json:"enableEventTracking"`
    MaxStateSnapshots   int              `json:"maxStateSnapshots"`
}

// Get default configuration
config := schema.DefaultRuntimeConfig()

// Customize for your needs
config.ValidationTiming = schema.ValidateOnChange
config.DebounceDelay = 500 * time.Millisecond
config.EnableEventTracking = true
```

### Runtime Lifecycle

```go
// 1. Initialize runtime
rt := runtime.NewRuntime(config).
    WithRenderer(myRenderer).
    WithValidator(myValidator).
    WithConditionalEngine(myConditionalEngine).
    Build()

// 2. Initialize schema state
err := rt.InitializeState(ctx, schema, initialData)

// 3. Handle field changes
err = rt.SetFieldValue(ctx, "firstName", "John")

// 4. Validate when needed
errors := rt.ValidateAll(ctx)

// 5. Render form
html, err := rt.RenderForm(ctx)

// 6. Handle submission
result, err := rt.Submit(ctx)
```

### Event System Integration

```go
// Register event handlers
rt.OnChange(func(ctx context.Context, event *schema.Event) error {
    log.Printf("Field %s changed from %v to %v", 
        event.Field, event.OldValue, event.Value)
    
    // Trigger conditional evaluation
    return rt.EvaluateConditionals(ctx)
})

rt.OnBlur(func(ctx context.Context, event *schema.Event) error {
    // Validate field on blur
    return rt.ValidateField(ctx, event.Field)
})

rt.OnSubmit(func(ctx context.Context) error {
    // Custom submission logic
    return rt.SubmitToAPI(ctx)
})
```

---

## Builder API Reference

### Schema Builder

The Builder provides a fluent API for programmatically creating schemas.

#### Basic Schema Configuration

```go
builder := schema.NewBuilder("form-id", schema.TypeForm, "Form Title").
    WithDescription("Form description").
    WithVersion("1.0.0").
    WithCategory("user-management").
    WithModule("users").
    WithTags("form", "user", "admin")
```

#### Layout Configuration

```go
// Grid layout
layout := &schema.Layout{
    Type:    schema.LayoutGrid,
    Columns: 2,
    Gap:     "1rem",
    Areas: []schema.GridArea{
        {Name: "header", Row: "1", Column: "1 / -1"},
        {Name: "content", Row: "2", Column: "1 / -1"},
    },
}

builder.WithLayout(layout)
```

#### Field Creation Methods

```go
// Text fields
builder.AddTextField("firstName", "First Name", true)
builder.AddEmailField("email", "Email Address", true)  
builder.AddPasswordField("password", "Password", true)
builder.AddTextareaField("bio", "Biography", false, 4) // 4 rows

// Numbers
min, max := 0.0, 100.0
builder.AddNumberField("age", "Age", true, &min, &max)

// Selections
options := []schema.Option{
    {Value: "admin", Label: "Administrator"},
    {Value: "user", Label: "Standard User"},
}
builder.AddSelectField("role", "Role", true, options)

// Dates
builder.AddDateField("birthDate", "Birth Date", false)

// Boolean
builder.AddCheckboxField("terms", "Accept Terms", false)

// Advanced field with full configuration
field := schema.Field{
    Name:        "customField",
    Type:        schema.FieldText,
    Label:       "Custom Field",
    Help:        "This is a custom field with full configuration",
    Placeholder: "Enter value...",
    Required:    true,
    Validation: &schema.FieldValidation{
        MinLength: &[]int{3}[0],
        MaxLength: &[]int{50}[0],
        Pattern:   "^[a-zA-Z\\s]+$",
    },
    Conditional: &schema.Conditional{
        Show: &schema.ConditionGroup{
            Logic: "AND",
            Conditions: []schema.Condition{
                {Field: "role", Operator: "equal", Value: "admin"},
            },
        },
    },
}
builder.AddFieldWithConfig(field)
```

#### Actions

```go
// Buttons
builder.AddSubmitButton("Save User")
builder.AddResetButton("Clear Form")

// Custom actions
action := schema.Action{
    ID:       "custom-action",
    Type:     schema.ActionButton,
    Text:     "Custom Action",
    Variant:  "secondary",
    OnClick:  "handleCustomAction",
}
builder.AddAction(action)
```

#### Enterprise Features

```go
// Multi-tenancy
builder.WithTenant("companyId", "strict")

// Internationalization  
builder.WithI18n("en", "en", "es", "fr", "ar")

// Security
builder.WithCSRF()
builder.WithRateLimit(10, 60) // 10 requests per 60 seconds

// HTMX Integration
builder.WithHTMX("/api/forms/submit", "#form-results")

// Alpine.js Integration
builder.WithAlpine("{ loading: false, submit() { this.loading = true } }")
```

#### Business Rules

```go
// Conditional visibility rule
rule := &schema.BusinessRule{
    ID:      "admin-only-fields",
    Name:    "Admin Only Fields",
    Type:    schema.RuleTypeFieldVisibility,
    Enabled: true,
    Condition: &condition.ConditionGroup{
        ID:          "admin-check",
        Conjunction: condition.ConjunctionAnd,
        Children: []any{
            &condition.ConditionRule{
                ID:    "role-check",
                Left:  condition.Expression{Type: condition.ValueTypeField, Field: "role"},
                Op:    condition.OpEqual,
                Right: "admin",
            },
        },
    },
    Actions: []schema.BusinessRuleAction{
        {Type: schema.ActionShowField, Target: "adminOnlyField"},
    },
}

builder.WithBusinessRule(rule)
```

#### Mixins

```go
// Apply built-in mixins
builder.WithMixin("audit_fields") // Adds created_at, updated_at, etc.

// Custom mixin
customMixin := &schema.Mixin{
    ID:          "contact_info",
    Name:        "Contact Information",
    Description: "Standard contact fields",
    Fields: []schema.Field{
        {Name: "phone", Type: schema.FieldPhone, Label: "Phone Number"},
        {Name: "address", Type: schema.FieldTextarea, Label: "Address"},
    },
}
builder.WithCustomMixin(customMixin)
```

#### Build Schema

```go
// Build the schema
schema, err := builder.Build(ctx)
if err != nil {
    return nil, fmt.Errorf("failed to build schema: %w", err)
}

// Build with business rules applied
schema, err = builder.BuildWithRules(ctx, map[string]any{
    "userRole": "admin",
})
```

---

## Design Token System

### Three-Tier Architecture

The design token system follows a three-tier architecture for maximum flexibility:

1. **Primitives**: Raw values (colors, spacing, typography)
2. **Semantic**: Functional meanings (primary, success, error)  
3. **Components**: Component-specific styles (button, input, card)

### Token Structure

```go
type DesignTokens struct {
    Primitives *PrimitiveTokens `json:"primitives"`
    Semantic   *SemanticTokens  `json:"semantic"`
    Components *ComponentTokens `json:"components"`
}

type PrimitiveTokens struct {
    Colors     *ColorTokens     `json:"colors"`
    Typography *TypographyTokens `json:"typography"`
    Spacing    *SpacingTokens   `json:"spacing"`
    Borders    *BorderTokens    `json:"borders"`
    Shadows    *ShadowTokens    `json:"shadows"`
    Breakpoints *BreakpointTokens `json:"breakpoints"`
}
```

### Token References

Tokens can reference other tokens using the `{token.path}` syntax:

```json
{
  "primitives": {
    "colors": {
      "blue": {
        "50": "hsl(220, 50%, 95%)",
        "500": "hsl(220, 50%, 50%)",
        "900": "hsl(220, 50%, 10%)"
      }
    }
  },
  "semantic": {
    "colors": {
      "primary": {
        "base": "primitives.colors.blue.500",
        "hover": "primitives.colors.blue.600",
        "active": "primitives.colors.blue.700"
      }
    }
  },
  "components": {
    "button": {
      "primary": {
        "background": "semantic.colors.primary.base",
        "color": "semantic.colors.text.on-primary",
        "border": "none"
      }
    }
  }
}
```

### Token Resolution

```go
// Get resolved tokens for a component
tokens := theme.GetComponentTokens("button.primary")

// Use in your templates/components
styles := map[string]string{
    "background-color": tokens["background"],
    "color":           tokens["color"], 
    "border":          tokens["border"],
    "padding":         tokens["padding"],
}
```

### CSS Custom Properties Generation

```go
// Generate CSS custom properties from tokens
func generateCSS(tokens *schema.DesignTokens) string {
    var css strings.Builder
    css.WriteString(":root {\n")
    
    // Primitives
    for name, value := range tokens.Primitives.Colors.Flatten() {
        css.WriteString(fmt.Sprintf("  --color-%s: %s;\n", name, value))
    }
    
    // Semantic tokens
    for name, value := range tokens.Semantic.Colors.Flatten() {
        css.WriteString(fmt.Sprintf("  --semantic-%s: %s;\n", name, value))
    }
    
    // Component tokens
    for component, tokens := range tokens.Components.Flatten() {
        for property, value := range tokens {
            css.WriteString(fmt.Sprintf("  --%s-%s: %s;\n", component, property, value))
        }
    }
    
    css.WriteString("}\n")
    return css.String()
}
```

### Dark Mode Support

```go
// Theme with dark mode variants
theme := &schema.Theme{
    ID:   "default",
    Name: "Default Theme",
    Tokens: &schema.DesignTokens{
        Semantic: &schema.SemanticTokens{
            Colors: &schema.ColorTokens{
                Background: schema.TokenReference("primitives.colors.white"),
                Text:       schema.TokenReference("primitives.colors.gray.900"),
            },
        },
    },
    DarkMode: &schema.DarkModeTokens{
        Colors: &schema.ColorTokens{
            Background: schema.TokenReference("primitives.colors.gray.900"),
            Text:       schema.TokenReference("primitives.colors.white"),
        },
    },
}

// Apply dark mode
darkTokens := theme.GetDarkModeTokens()
```

---

## Validation Integration

### Built-in Validation

The schema system includes comprehensive built-in validation:

#### Field-Level Validation

```go
field := schema.Field{
    Name:     "email",
    Type:     schema.FieldEmail,
    Required: true,
    Validation: &schema.FieldValidation{
        Format: "email",
        Custom: "checkEmailAvailability", // Custom validator
    },
}
```

#### Validation Rules

```go
type FieldValidation struct {
    // Length constraints
    MinLength *int    `json:"minLength,omitempty"`
    MaxLength *int    `json:"maxLength,omitempty"`
    
    // Numeric constraints  
    Min  *float64 `json:"min,omitempty"`
    Max  *float64 `json:"max,omitempty"`
    Step *float64 `json:"step,omitempty"`
    
    // Pattern matching
    Pattern string `json:"pattern,omitempty"`
    Format  string `json:"format,omitempty"`
    
    // Custom validation
    Custom string `json:"custom,omitempty"`
    
    // Uniqueness (requires database)
    Unique bool `json:"unique,omitempty"`
}
```

### Custom Validators

```go
// Register custom validators
validator := validate.NewValidator(db)
registry := validate.NewValidationRegistry()

// Add custom field validator
registry.Register("checkEmailAvailability", func(ctx context.Context, value any, params map[string]any) error {
    email := value.(string)
    if isEmailTaken(email) {
        return validate.NewValidationError("email_taken", "This email is already in use", "unique")
    }
    return nil
})

// Add cross-field validator
crossFieldRegistry := validate.NewCrossFieldValidationRegistry()
crossFieldRegistry.Register("password_confirmation", func(ctx context.Context, data map[string]any) error {
    password := data["password"]
    confirmation := data["passwordConfirmation"]
    
    if password != confirmation {
        return validate.NewValidationError("password_mismatch", "Passwords do not match", "match")
    }
    return nil
})
```

### Validation Timing

```go
// Configure when validation occurs
config := schema.RuntimeConfig{
    ValidationTiming: schema.ValidateOnBlur,    // Validate on field blur
    EnableDebounce:   true,                     // Debounce validation
    DebounceDelay:    300 * time.Millisecond,  // Delay before validation
}

// Other timing options:
// ValidateOnChange - Immediate validation
// ValidateOnSubmit - Only validate on submission
// ValidateNever    - No automatic validation
```

### Async Validation

```go
// Async validation for expensive operations
func (v *CustomValidator) ValidateFieldAsync(
    ctx context.Context, 
    field *schema.Field, 
    value any, 
    callback validate.ValidationCallback,
) error {
    go func() {
        // Expensive validation (API call, database lookup)
        time.Sleep(500 * time.Millisecond)
        
        var errors []string
        if isInvalid(value) {
            errors = append(errors, "Value is invalid")
        }
        
        callback(errors)
    }()
    
    return nil
}
```

### Business Rules Validation

```go
// Business rules with validation
rule := &schema.BusinessRule{
    ID:   "age_verification",
    Type: schema.RuleTypeValidation,
    Condition: &condition.ConditionGroup{
        Children: []any{
            &condition.ConditionRule{
                Left:  condition.Expression{Type: condition.ValueTypeField, Field: "age"},
                Op:    condition.OpLessThan,
                Right: 18,
            },
        },
    },
    Actions: []schema.BusinessRuleAction{
        {
            Type:    schema.ActionValidationError,
            Target:  "age",
            Message: "Must be 18 or older",
        },
    },
}
```

---

## Internationalization (I18n)

### Embedded Translation System

The schema package uses an **embedded translation system** where translations are compiled into the binary using Go's `embed` directive. This provides zero runtime I/O and automatic locale detection.

#### Key Features:
- **Embedded JSON Files**: Translations stored in `pkg/schema/translations/*.json`
- **Field ID-based Keys**: Field names are automatically used as translation keys
- **Zero Configuration**: Works out of the box with no setup required
- **Automatic Fallbacks**: User locale → Language-only → Default ("en")
- **RTL Support**: Built-in support for right-to-left languages
- **Browser Detection**: Automatic locale detection from Accept-Language headers

### Translation Files Structure

```json
// pkg/schema/translations/en.json
{
  "fields": {
    "firstName": "First Name",
    "lastName": "Last Name", 
    "email": "Email Address",
    "password": "Password"
  },
  "validation": {
    "required": "This field is required",
    "invalidEmail": "Please enter a valid email address",
    "minLength": "Must be at least {min} characters",
    "maxLength": "Must be no more than {max} characters"
  },
  "actions": {
    "submit": "Submit",
    "cancel": "Cancel",
    "save": "Save"
  },
  "status": {
    "loading": "Loading...",
    "saving": "Saving...",
    "error": "An error occurred"
  }
}
```

```json
// pkg/schema/translations/es.json
{
  "fields": {
    "firstName": "Nombre",
    "lastName": "Apellido",
    "email": "Correo Electrónico",
    "password": "Contraseña"
  },
  "validation": {
    "required": "Este campo es requerido",
    "invalidEmail": "Por favor ingresa un email válido",
    "minLength": "Debe tener al menos {min} caracteres",
    "maxLength": "Debe tener máximo {max} caracteres"
  },
  "actions": {
    "submit": "Enviar",
    "cancel": "Cancelar",
    "save": "Guardar"
  },
  "status": {
    "loading": "Cargando...",
    "saving": "Guardando...",
    "error": "Ocurrió un error"
  }
}
```

### Runtime Integration with I18n

The runtime system automatically provides localized validation messages:

```go
// Create runtime with locale support
runtime := runtime.NewRuntimeBuilder(schema).
    WithLocale("es").  // Set default locale
    Build(ctx)

// Or detect user's preferred locale automatically
userPreferences := []string{"es-ES", "es", "en"}
detectedLocale := runtime.DetectUserLocale(userPreferences)
// Returns "es" if available

// Change locale at runtime
err := runtime.SetLocale(ctx, "fr")

// Get current locale
currentLocale := runtime.GetCurrentLocale() // "fr"

// Validation errors are automatically localized
errors := runtime.ValidateField(ctx, "email", "invalid")
// Returns localized error: "Por favor ingresa un email válido" (if locale is "es")
```

### Enricher Integration with I18n

The enricher automatically localizes schemas based on user preferences:

```go
// User interface includes locale preference
type User interface {
    GetID() string
    GetTenantID() string
    GetPermissions() []string
    GetRoles() []string
    HasPermission(permission string) bool
    HasRole(role string) bool
    GetPreferredLocale() string // Returns user's preferred locale
}

// Enrich schema with user's preferred locale
enricher := enrich.NewEnricher()
enrichedSchema, err := enricher.EnrichWithLocale(ctx, schema, user, "")
// Automatically uses user.GetPreferredLocale()

// Override with specific locale
enrichedSchema, err := enricher.EnrichWithLocale(ctx, schema, user, "fr")

// Field labels are now localized:
// enrichedSchema.Fields[0].Label = "Prénom" (firstName in French)
// enrichedSchema.Fields[1].Label = "Nom" (lastName in French)
```

### Schema-Level I18n

Schemas support explicit I18n for titles and descriptions, while fields use embedded translations:

```go
schema := &schema.Schema{
    ID:          "user-form",
    Title:       "User Registration",
    Description: "Please fill out the form",
    I18n: &schema.SchemaI18n{
        Title: map[string]string{
            "en": "User Registration",
            "es": "Registro de Usuario",
            "fr": "Inscription Utilisateur",
            "ar": "تسجيل المستخدم",
        },
        Description: map[string]string{
            "en": "Please fill out the form",
            "es": "Por favor complete el formulario",
            "fr": "Veuillez remplir le formulaire",
            "ar": "يرجى ملء النموذج",
        },
        Direction: map[string]string{
            "en": "ltr",
            "ar": "rtl",
        },
    },
}

// Localization is applied automatically during enrichment
// localizedSchema.Title becomes "Registro de Usuario" for Spanish users
```

### Field-Level I18n

Fields automatically use embedded translations based on their field ID (name):

```go
// Simple field creation - translations handled automatically
field := schema.Field{
    Name:        "firstName",  // Used as translation key
    Type:        schema.FieldText,
    Label:       "First Name", // English fallback
    Placeholder: "Enter your first name",
    Help:        "Your legal first name",
}

// Get localized values (uses embedded translations)
label := field.GetLocalizedLabel("es")        // "Nombre" (from es.json)
placeholder := field.GetLocalizedPlaceholder("es") // Fallback behavior
help := field.GetLocalizedHelp("es")          // Fallback behavior

// For custom fields not in embedded translations, use explicit I18n
customField := schema.Field{
    Name:  "customBusinessField",
    Type:  schema.FieldText,
    Label: "Custom Business Field",
    I18n: &schema.FieldI18n{
        Label: map[string]string{
            "en": "Custom Business Field",
            "es": "Campo de Negocio Personalizado",
            "fr": "Champ Commercial Personnalisé",
        },
    },
}

// GetLocalizedLabel checks I18n first, then embedded translations, then humanizes field ID
customLabel := customField.GetLocalizedLabel("es") // "Campo de Negocio Personalizado"
```

### Validation Message I18n

```go
// Global translation functions (use embedded translations)
validationMsg := schema.T_Validation("es", "required", nil)
// Returns: "Este campo es requerido"

validationMsg := schema.T_Validation("es", "minLength", map[string]any{
    "min": 8,
})
// Returns: "Debe tener al menos 8 caracteres"

// Field labels
fieldLabel := schema.T_Field("es", "firstName") // "Nombre"

// Action text
actionText := schema.T_Action("es", "submit") // "Enviar"

// Status messages
statusText := schema.T_Status("es", "loading") // "Cargando..."

// Runtime automatically localizes validation errors
runtime := runtime.NewRuntimeBuilder(schema).WithLocale("es").Build(ctx)
errors := runtime.ValidateField(ctx, "email", "invalid")
// Returns: ["Por favor ingresa un email válido"]

// Check locale availability
hasSpanish := schema.T_HasLocale("es") // true
availableLocales := schema.Translator.GetAvailableLocales() // ["en", "es", "fr", "ar"]

// Detect best locale from user preferences
userPrefs := []string{"de-DE", "de", "fr", "en"}
bestLocale := schema.T_DetectLocale(userPrefs) // "fr" (since "de" not available)
```

### RTL Support

```go
// Check if locale is RTL
isRTL := schema.T_IsRTL("ar") // true

// Get text direction  
direction := schema.T_Direction("ar") // "rtl"

// Apply RTL styles in your renderer
func (r *TemplRenderer) RenderField(ctx context.Context, field *schema.Field, ...) (string, error) {
    locale := ctx.Value("locale").(string)
    direction := schema.T_Direction(locale)
    
    data := FieldRenderData{
        Field:     field,
        Direction: direction,
        IsRTL:     direction == "rtl",
    }
    
    return r.templates.ExecuteTemplate(ctx, "field", data)
}
```

### Complete I18n Workflow

```go
// 1. Schema writer creates form (only writes in English)
schema := schema.NewBuilder("user-form", schema.TypeForm, "User Registration").
    AddTextField("firstName", "First Name", true).     // Will auto-translate
    AddTextField("lastName", "Last Name", true).       // Will auto-translate  
    AddEmailField("email", "Email Address", true).     // Will auto-translate
    AddPasswordField("password", "Password", true).    // Will auto-translate
    Build(ctx)

// 2. User with Spanish browser/preference
user := &enrich.ExampleUser{
    ID:              "user-123",
    TenantID:        "tenant-abc",
    PreferredLocale: "es",
    Permissions:     []string{"read", "write"},
}

// 3. Browser sends Accept-Language or user sets preference
userBrowserLocales := []string{"es-ES", "es", "en"}

// 4. Enricher automatically localizes schema
enricher := enrich.NewEnricher()
localizedSchema, _ := enricher.EnrichWithLocale(ctx, schema, user, "")
// Schema fields are now localized:
// - localizedSchema.Fields[0].Label = "Nombre" 
// - localizedSchema.Fields[1].Label = "Apellido"
// - localizedSchema.Fields[2].Label = "Correo Electrónico"

// 5. Create runtime with automatic locale detection
runtime := runtime.NewRuntimeBuilder(localizedSchema).
    WithLocale("es").
    Build(ctx)

// Or detect from browser preferences
detectedLocale := runtime.DetectUserLocale(userBrowserLocales) // "es"

// 6. Validation errors are automatically localized
err := runtime.HandleFieldChange(ctx, "email", "invalid-email")
errors := runtime.GetErrors()["email"]
// errors[0] = "Por favor ingresa un email válido"

// 7. UI renderer receives fully localized schema
for _, field := range localizedSchema.Fields {
    fmt.Printf("Field %s: %s\n", field.Name, field.Label)
}
// Output:
// Field firstName: Nombre
// Field lastName: Apellido
// Field email: Correo Electrónico  
// Field password: Contraseña
```

### Available Translation Functions

```go
// Global translation functions available
schema.T_Field(locale, fieldID string) string              // Field labels
schema.T_Validation(locale, key string, params) string     // Validation messages
schema.T_Action(locale, actionKey string) string           // Action text
schema.T_Status(locale, statusKey string) string           // Status messages
schema.T_HasLocale(locale string) bool                     // Check availability
schema.T_DetectLocale(preferences []string) string         // Auto-detect best locale
schema.T_IsRTL(locale string) bool                         // Check RTL support
schema.T_Direction(locale string) string                   // Get text direction

// Translation catalog methods
schema.Translator.GetAvailableLocales() []string           // List available locales
schema.Translator.FieldLabel(locale, fieldID string) string // Direct field lookup
schema.Translator.ValidationMessage(locale, key, params)   // Direct validation lookup
```

### Adding New Locales

To add support for a new locale (e.g., German):

1. **Create translation file**: `pkg/schema/translations/de.json`
2. **Follow the same structure** as existing files
3. **Rebuild the application** - translations are embedded at compile time
4. **Test with new locale**:

```go
// Check if new locale is available
hasGerman := schema.T_HasLocale("de") // true (after rebuild)

// Use new locale
germanLabel := schema.T_Field("de", "firstName") // "Vorname"
```

### Supported Locales

By default, the system includes:
- **English (en)**: Base locale, always available
- **Spanish (es)**: Full translation coverage
- **French (fr)**: Full translation coverage  
- **Arabic (ar)**: Full translation coverage with RTL support

Additional locales can be added by creating corresponding JSON files in the `translations` directory.

---

## State Management

### State Structure

```go
type StateSnapshot struct {
    Values      map[string]any      `json:"values"`      // Current field values
    Initial     map[string]any      `json:"initial"`     // Initial values
    Touched     map[string]bool     `json:"touched"`     // Fields that were focused
    Dirty       map[string]bool     `json:"dirty"`       // Fields with changed values
    Errors      map[string][]string `json:"errors"`      // Validation errors
    Timestamp   time.Time           `json:"timestamp"`   // When snapshot was taken
    Initialized time.Time           `json:"initialized"` // When state was initialized
}
```

### State Operations

```go
// Initialize state with schema and initial data
initialData := map[string]any{
    "firstName": "John",
    "lastName":  "Doe",
    "email":     "john.doe@example.com",
}

err := runtime.InitializeState(ctx, schema, initialData)

// Get and set values
value, exists := runtime.GetValue("firstName")
err = runtime.SetValue("firstName", "Jane")
allValues := runtime.GetAll()

// Track field state
isTouched := runtime.IsTouched("firstName")
runtime.Touch("firstName")

isDirty := runtime.IsDirty("firstName")
hasChanges := runtime.IsAnyDirty()

// Manage errors
errors := runtime.GetErrors("firstName")
runtime.SetErrors("firstName", []string{"Field is required"})
allErrors := runtime.GetAllErrors()
isValid := runtime.IsValid()
```

### State Snapshots (Undo/Redo)

```go
// Create snapshot
snapshot := runtime.CreateSnapshot()

// Make changes
runtime.SetValue("firstName", "Jane")
runtime.SetValue("email", "jane@example.com")

// Restore previous state
runtime.RestoreSnapshot(snapshot)

// Get state statistics
stats := runtime.GetStats()
fmt.Printf("Fields: %d, Touched: %d, Dirty: %d, Errors: %d",
    stats.FieldCount, stats.TouchedCount, stats.DirtyCount, stats.ErrorCount)
```

### State Persistence

```go
// Export state for persistence
stateData := runtime.Export()
json, err := json.Marshal(stateData)

// Save to localStorage, session, database, etc.
saveToStorage(json)

// Restore state
var stateData map[string]any
err = json.Unmarshal(json, &stateData)
err = runtime.Import(stateData)
```

### Reactive State Updates

```go
// Listen for state changes
runtime.OnStateChange(func(ctx context.Context, changes *StateChanges) error {
    for field, change := range changes.Values {
        log.Printf("Field %s changed from %v to %v", field, change.Old, change.New)
        
        // Trigger UI updates
        notifyUI(field, change.New)
    }
    return nil
})

// Batch state updates
runtime.BatchUpdate(func(state *State) error {
    state.SetValue("firstName", "Jane")
    state.SetValue("lastName", "Smith") 
    state.SetValue("email", "jane.smith@example.com")
    return nil
})
// OnStateChange fires once for all changes
```

---

## Event Handling

### Event Types

```go
const (
    EventChange EventType = "change" // Field value changed
    EventBlur   EventType = "blur"   // Field lost focus  
    EventFocus  EventType = "focus"  // Field gained focus
    EventSubmit EventType = "submit" // Form submitted
    EventReset  EventType = "reset"  // Form reset
    EventInit   EventType = "init"   // Form initialized
)
```

### Event Structure

```go
type Event struct {
    Type      EventType `json:"type"`               // Event type
    Field     string    `json:"field,omitempty"`    // Field name (for field events)
    Value     any       `json:"value,omitempty"`    // Current value  
    OldValue  any       `json:"old_value,omitempty"` // Previous value
    Timestamp time.Time `json:"timestamp"`          // When event occurred
    Context   any       `json:"context,omitempty"`  // Additional context
}
```

### Event Handlers

```go
// Register event handlers
runtime.OnChange(func(ctx context.Context, event *Event) error {
    log.Printf("Field %s changed: %v → %v", event.Field, event.OldValue, event.Value)
    
    // Validate field if needed
    if runtime.GetConfig().ValidationTiming == ValidateOnChange {
        return runtime.ValidateField(ctx, event.Field)
    }
    
    return nil
})

runtime.OnBlur(func(ctx context.Context, event *Event) error {
    // Mark field as touched
    runtime.Touch(event.Field)
    
    // Validate on blur
    if runtime.GetConfig().ValidationTiming == ValidateOnBlur {
        return runtime.ValidateField(ctx, event.Field)
    }
    
    return nil
})

runtime.OnFocus(func(ctx context.Context, event *Event) error {
    // Clear field errors when focused
    runtime.ClearErrors(event.Field)
    return nil
})

runtime.OnSubmit(func(ctx context.Context) error {
    // Validate all fields
    if !runtime.IsValid() {
        return errors.New("form contains validation errors")
    }
    
    // Submit to API
    return submitToAPI(ctx, runtime.GetAll())
})
```

### Custom Events

```go
// Register custom event handler
runtime.RegisterEventHandler("custom-validation", func(ctx context.Context, event *Event) error {
    // Custom validation logic
    return validateBusinessRules(ctx, event)
})

// Trigger custom event
event := &Event{
    Type:      "custom-validation",
    Field:     "specialField",
    Value:     "some-value",
    Timestamp: time.Now(),
}

err := runtime.TriggerEvent(ctx, event)
```

### Event Configuration

```go
// Configure debouncing
config := &DebouncedConfig{
    ChangeDelay: 500 * time.Millisecond, // Debounce change events
    BlurDelay:   100 * time.Millisecond, // Debounce blur events  
    FocusDelay:  0,                      // No debounce for focus
}

runtime.SetDebouncingConfig(config)

// Configure event tracking
eventConfig := &EventHandlerConfig{
    ValidationTiming: ValidateOnBlur,
    EnableTracking:   true, // Track events for analytics
    Enabled:          true,
}

runtime.SetEventConfig(eventConfig)
```

### Event Statistics

```go
// Get event statistics
stats := runtime.GetEventStats()

fmt.Printf("Total events: %d", stats.TotalEvents)
fmt.Printf("Events by type: %+v", stats.EventsByType)
fmt.Printf("Events by field: %+v", stats.EventsByField)
fmt.Printf("Last event: %s on %s at %s", 
    stats.LastEventType, stats.LastEventField, stats.LastEventTime)
```

---

## Conditional Logic

### Field Conditionals

```go
field := schema.Field{
    Name:  "companyName",
    Type:  schema.FieldText,
    Label: "Company Name",
    Conditional: &schema.Conditional{
        Show: &schema.ConditionGroup{
            Logic: "AND",
            Conditions: []schema.Condition{
                {Field: "accountType", Operator: "equal", Value: "business"},
            },
        },
    },
}
```

### Complex Conditions

```go
// Multiple conditions with logic
conditional := &schema.Conditional{
    Show: &schema.ConditionGroup{
        Logic: "OR",
        Conditions: []schema.Condition{
            {Field: "role", Operator: "equal", Value: "admin"},
            {
                Logic: "AND", 
                Conditions: []schema.Condition{
                    {Field: "role", Operator: "equal", Value: "manager"},
                    {Field: "department", Operator: "equal", Value: "IT"},
                },
            },
        },
    },
    Hide: &schema.ConditionGroup{
        Logic: "AND",
        Conditions: []schema.Condition{
            {Field: "status", Operator: "equal", Value: "inactive"},
        },
    },
}
```

### Conditional Operators

```go
// Available operators
const (
    OperatorEqual              = "equal"
    OperatorNotEqual          = "not_equal"
    OperatorGreaterThan       = "greater_than"
    OperatorLessThan          = "less_than"
    OperatorGreaterThanEqual  = "greater_than_equal"
    OperatorLessThanEqual     = "less_than_equal"
    OperatorContains          = "contains"
    OperatorNotContains       = "not_contains"
    OperatorIn                = "in"
    OperatorNotIn             = "not_in"
    OperatorIsEmpty           = "is_empty"
    OperatorIsNotEmpty        = "is_not_empty"
    OperatorStartsWith        = "starts_with"
    OperatorEndsWith          = "ends_with"
    OperatorMatches           = "matches"         // Regex
    OperatorBetween           = "between"
    OperatorNotBetween        = "not_between"
)

// Examples
conditions := []schema.Condition{
    {Field: "age", Operator: "greater_than", Value: 18},
    {Field: "name", Operator: "contains", Value: "john"},
    {Field: "role", Operator: "in", Value: []string{"admin", "manager"}},
    {Field: "email", Operator: "matches", Value: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`},
    {Field: "salary", Operator: "between", Value: []float64{50000, 100000}},
}
```

### Action Conditionals

```go
action := schema.Action{
    ID:   "delete-button",
    Type: schema.ActionButton,
    Text: "Delete",
    Conditional: &schema.Conditional{
        Show: &schema.ConditionGroup{
            Logic: "AND",
            Conditions: []schema.Condition{
                {Field: "role", Operator: "in", Value: []string{"admin", "owner"}},
                {Field: "status", Operator: "not_equal", Value: "readonly"},
            },
        },
    },
}
```

### Layout Conditionals  

```go
section := &schema.Section{
    Title: "Advanced Settings",
    Conditional: &schema.Conditional{
        Show: &schema.ConditionGroup{
            Logic: "OR",
            Conditions: []schema.Condition{
                {Field: "userLevel", Operator: "equal", Value: "advanced"},
                {Field: "role", Operator: "equal", Value: "admin"},
            },
        },
    },
    Fields: []string{"advancedField1", "advancedField2"},
}
```

### Conditional Engine Implementation

```go
type ConditionalEngine struct {
    evaluator *condition.Evaluator
}

func (c *ConditionalEngine) EvaluateFieldVisibility(
    ctx context.Context, 
    field *schema.Field, 
    data map[string]any,
) (bool, error) {
    if field.Conditional == nil {
        return true, nil // Visible by default
    }
    
    // Evaluate show condition
    if field.Conditional.Show != nil {
        visible, err := c.evaluateConditionGroup(field.Conditional.Show, data)
        if err != nil || !visible {
            return false, err
        }
    }
    
    // Evaluate hide condition  
    if field.Conditional.Hide != nil {
        hidden, err := c.evaluateConditionGroup(field.Conditional.Hide, data)
        if err != nil {
            return false, err
        }
        if hidden {
            return false, nil
        }
    }
    
    return true, nil
}

func (c *ConditionalEngine) EvaluateFieldRequired(
    ctx context.Context,
    field *schema.Field,
    data map[string]any,
) (bool, error) {
    // Base required state
    required := field.Required
    
    // Apply conditional required logic
    if field.ConditionalRequired != nil {
        condRequired, err := c.evaluateConditionGroup(field.ConditionalRequired, data)
        if err != nil {
            return required, err
        }
        required = required || condRequired
    }
    
    return required, nil
}
```

---

## Theme System

### Theme Structure

```go
type Theme struct {
    ID          string        `json:"id"`
    Name        string        `json:"name"`
    Description string        `json:"description"`
    Tokens      *DesignTokens `json:"tokens"`
    DarkMode    *DarkModeTokens `json:"darkMode,omitempty"`
    Variants    map[string]*DesignTokens `json:"variants,omitempty"`
    Tenant      *TenantTheme  `json:"tenant,omitempty"`
}
```

### Default Themes

```go
// Get default theme
theme := schema.GetDefaultTheme()

// Get all available themes
themes := schema.GetAvailableThemes()

// Load specific theme
theme, err := schema.LoadTheme("dark")
```

### Creating Custom Themes

```go
customTheme := &schema.Theme{
    ID:          "company-theme",
    Name:        "Company Theme", 
    Description: "Custom theme for company branding",
    Tokens: &schema.DesignTokens{
        Primitives: &schema.PrimitiveTokens{
            Colors: &schema.ColorTokens{
                Primary: map[string]string{
                    "50":  "hsl(210, 100%, 95%)",
                    "100": "hsl(210, 100%, 90%)",
                    "500": "hsl(210, 100%, 50%)",
                    "900": "hsl(210, 100%, 10%)",
                },
                Gray: map[string]string{
                    "50":  "hsl(0, 0%, 97%)",
                    "100": "hsl(0, 0%, 90%)",
                    "500": "hsl(0, 0%, 50%)",
                    "900": "hsl(0, 0%, 10%)",
                },
            },
            Typography: &schema.TypographyTokens{
                FontFamily: map[string]string{
                    "sans":      "Inter, system-ui, sans-serif",
                    "serif":     "Georgia, serif",
                    "mono":      "JetBrains Mono, monospace",
                    "heading":   "Poppins, sans-serif",
                },
                FontSize: map[string]string{
                    "xs":   "0.75rem",
                    "sm":   "0.875rem",
                    "base": "1rem",
                    "lg":   "1.125rem",
                    "xl":   "1.25rem",
                    "2xl":  "1.5rem",
                },
                LineHeight: map[string]string{
                    "tight":  "1.25",
                    "normal": "1.5",
                    "loose":  "1.75",
                },
            },
            Spacing: &schema.SpacingTokens{
                Scale: map[string]string{
                    "0":   "0",
                    "1":   "0.25rem",
                    "2":   "0.5rem", 
                    "4":   "1rem",
                    "8":   "2rem",
                    "16":  "4rem",
                },
            },
        },
        Semantic: &schema.SemanticTokens{
            Colors: &schema.ColorTokens{
                Primary:    schema.TokenReference("primitives.colors.primary.500"),
                Success:    schema.TokenReference("primitives.colors.green.500"),
                Warning:    schema.TokenReference("primitives.colors.yellow.500"),
                Error:      schema.TokenReference("primitives.colors.red.500"),
                Background: schema.TokenReference("primitives.colors.white"),
                Text:       schema.TokenReference("primitives.colors.gray.900"),
            },
        },
        Components: &schema.ComponentTokens{
            Button: map[string]map[string]string{
                "primary": {
                    "background":    "semantic.colors.primary",
                    "color":        "semantic.colors.white",
                    "border":       "none",
                    "borderRadius": "primitives.borders.radius.md",
                    "padding":      "primitives.spacing.2} {primitives.spacing.4",
                },
                "secondary": {
                    "background":    "transparent", 
                    "color":        "semantic.colors.primary",
                    "border":       "1px solid {semantic.colors.primary",
                    "borderRadius": "primitives.borders.radius.md",
                    "padding":      "primitives.spacing.2} {primitives.spacing.4",
                },
            },
            Input: map[string]map[string]string{
                "default": {
                    "background":    "semantic.colors.background",
                    "color":        "semantic.colors.text",
                    "border":       "1px solid {primitives.colors.gray.300",
                    "borderRadius": "primitives.borders.radius.md",
                    "padding":      "primitives.spacing.2} {primitives.spacing.3",
                },
                "error": {
                    "borderColor": "semantic.colors.error",
                },
            },
        },
    },
}

// Register theme
err := schema.RegisterTheme(customTheme)
```

### Dark Mode Support

```go
// Add dark mode tokens
customTheme.DarkMode = &schema.DarkModeTokens{
    Colors: &schema.ColorTokens{
        Background: schema.TokenReference("primitives.colors.gray.900"),
        Text:       schema.TokenReference("primitives.colors.white"),
        Primary:    schema.TokenReference("primitives.colors.primary.400"),
    },
}

// Get dark mode tokens
darkTokens := theme.GetDarkModeTokens()

// Apply dark mode in renderer
func (r *Renderer) applyTheme(darkMode bool) map[string]string {
    if darkMode {
        return r.theme.GetDarkModeTokens()
    }
    return r.theme.GetTokens()
}
```

### Tenant-Specific Themes

```go
// Create tenant-specific theme override
tenantTheme := &schema.TenantTheme{
    TenantID: "company-abc",
    Overrides: &schema.DesignTokens{
        Primitives: &schema.PrimitiveTokens{
            Colors: &schema.ColorTokens{
                Primary: map[string]string{
                    "500": "hsl(260, 60%, 50%)", // Company brand color
                },
            },
        },
    },
}

// Apply tenant theme
theme.ApplyTenantOverrides(tenantTheme)
```

### Using Themes in Renderers

```go
type ThemeAwareRenderer struct {
    baseRenderer Renderer
    theme        *schema.Theme
    darkMode     bool
}

func (r *ThemeAwareRenderer) RenderField(
    ctx context.Context,
    field *schema.Field,
    value any,
    errors []string,
    touched, dirty bool,
) (string, error) {
    // Get component tokens
    componentName := r.getComponentName(field.Type)
    tokens := r.theme.GetComponentTokens(componentName)
    
    // Apply state-based tokens
    if len(errors) > 0 {
        errorTokens := r.theme.GetComponentTokens(componentName + ".error")
        for key, value := range errorTokens {
            tokens[key] = value
        }
    }
    
    // Resolve token references
    resolvedTokens := r.theme.ResolveTokens(tokens)
    
    // Render with tokens
    return r.baseRenderer.RenderField(ctx, field, value, errors, touched, dirty, resolvedTokens)
}

func (r *ThemeAwareRenderer) getComponentName(fieldType schema.FieldType) string {
    mapping := map[schema.FieldType]string{
        schema.FieldText:     "input",
        schema.FieldEmail:    "input", 
        schema.FieldPassword: "input",
        schema.FieldNumber:   "input",
        schema.FieldTextarea: "textarea",
        schema.FieldSelect:   "select",
        schema.FieldCheckbox: "checkbox",
        schema.FieldRadio:    "radio",
    }
    return mapping[fieldType]
}
```

---

## Advanced Features

### Multi-Tenancy

```go
// Enable tenant isolation
builder.WithTenant("companyId", "strict")

// Tenant-aware field access
field := schema.Field{
    Name: "sensitiveData",
    Type: schema.FieldText,
    Tenant: &schema.FieldTenant{
        Required: true,
        Field:    "companyId",
        Access: map[string]schema.TenantAccess{
            "company-a": {Read: true, Write: true},
            "company-b": {Read: true, Write: false},
        },
    },
}
```

### Permissions & Enrichment

```go
// User context for permission evaluation
userContext := &schema.UserContext{
    UserID:      "user-123",
    Role:        "manager", 
    Permissions: []string{"read:users", "write:users"},
    TenantID:    "company-a",
    Metadata: map[string]any{
        "department": "IT",
        "level":     "senior",
    },
}

// Enrich schema with permissions
enricher := enrich.NewEnricher(userContext)
enrichedSchema, err := enricher.EnrichSchema(ctx, schema)

// Fields may be hidden, readonly, or have different defaults based on permissions
```

### Workflow Integration

```go
// Multi-step workflow schema
workflow := &schema.Workflow{
    ID:   "user-onboarding",
    Name: "User Onboarding",
    Steps: []schema.WorkflowStep{
        {
            ID:          "personal-info",
            Title:       "Personal Information", 
            Description: "Enter your personal details",
            Fields:      []string{"firstName", "lastName", "email"},
            Required:    true,
        },
        {
            ID:          "company-info",
            Title:       "Company Information",
            Fields:      []string{"company", "role", "department"},
            Required:    true,
            Conditional: &schema.Conditional{
                Show: &schema.ConditionGroup{
                    Conditions: []schema.Condition{
                        {Field: "accountType", Operator: "equal", Value: "business"},
                    },
                },
            },
        },
        {
            ID:    "review",
            Title: "Review & Submit",
            Type:  "review",
        },
    },
}

builder.WithWorkflow(workflow)
```

### Business Rules Engine

```go
// Complex business rule
rule := &schema.BusinessRule{
    ID:          "discount-eligibility",
    Name:        "Discount Eligibility",
    Description: "Apply discount based on customer criteria",
    Type:        schema.RuleTypeDataCalculation,
    Priority:    100,
    Condition: &condition.ConditionGroup{
        Conjunction: condition.ConjunctionAnd,
        Children: []any{
            &condition.ConditionRule{
                ID:    "membership-check",
                Left:  condition.Expression{Type: condition.ValueTypeField, Field: "membershipType"},
                Op:    condition.OpEqual,
                Right: "premium",
            },
            &condition.ConditionRule{
                ID:    "amount-check", 
                Left:  condition.Expression{Type: condition.ValueTypeField, Field: "orderAmount"},
                Op:    condition.OpGreaterThan,
                Right: 1000,
            },
        },
    },
    Actions: []schema.BusinessRuleAction{
        {
            Type:   schema.ActionCalculate,
            Target: "discountAmount",
            Formula: "orderAmount * 0.15", // 15% discount
        },
        {
            Type:    schema.ActionSetValue,
            Target:  "discountApplied", 
            Value:   true,
        },
    },
}

// Register rule
ruleEngine := schema.NewBusinessRuleEngine()
err = ruleEngine.RegisterRule(rule)

// Apply rules to data
modifiedData, err := ruleEngine.ApplyRules(ctx, schema, formData)
```

### Repeatable Fields (Dynamic Arrays)

```go
// Line items for invoice
repeatableField := &schema.RepeatableField{
    Field: schema.Field{
        Name:  "lineItems",
        Type:  schema.FieldRepeatable,
        Label: "Line Items",
    },
    Template: []schema.Field{
        {Name: "product", Type: schema.FieldText, Label: "Product", Required: true},
        {Name: "quantity", Type: schema.FieldNumber, Label: "Quantity", Required: true},
        {Name: "price", Type: schema.FieldNumber, Label: "Unit Price", Required: true},
        {Name: "total", Type: schema.FieldNumber, Label: "Total", Readonly: true},
    },
    MinItems: 1,
    MaxItems: 50,
    DefaultItems: 3,
    BusinessRules: []schema.BusinessRule{
        {
            ID:   "calculate-line-total",
            Type: schema.RuleTypeDataCalculation,
            Actions: []schema.BusinessRuleAction{
                {
                    Type:    schema.ActionCalculate,
                    Target:  "total",
                    Formula: "quantity * price",
                },
            },
        },
    },
}

builder.WithRepeatable(repeatableField)
```

### Advanced Validation

```go
// Cross-field validation
crossFieldRule := &schema.BusinessRule{
    ID:   "password-confirmation",
    Type: schema.RuleTypeValidation,
    Condition: &condition.ConditionGroup{
        Children: []any{
            &condition.ConditionRule{
                Left:  condition.Expression{Type: condition.ValueTypeField, Field: "password"},
                Op:    condition.OpNotEqual,
                Right: condition.Expression{Type: condition.ValueTypeField, Field: "passwordConfirm"},
            },
        },
    },
    Actions: []schema.BusinessRuleAction{
        {
            Type:    schema.ActionValidationError,
            Target:  "passwordConfirm",
            Message: "Passwords must match",
        },
    },
}

// Async validation with debouncing
field := schema.Field{
    Name: "username",
    Type: schema.FieldText,
    Validation: &schema.FieldValidation{
        Custom:      "checkUsernameAvailability",
        AsyncDelay:  500 * time.Millisecond, // Debounce API calls
        AsyncTimeout: 5 * time.Second,
    },
}
```

---

## Integration Patterns

### React Integration

```typescript
// React hook for schema runtime
import { useSchemaRuntime } from './schema-runtime-hook';

function SchemaForm({ schemaId, initialData }: SchemaFormProps) {
    const { 
        schema, 
        runtime, 
        state, 
        errors, 
        isLoading,
        setValue,
        handleSubmit 
    } = useSchemaRuntime(schemaId, initialData);

    if (isLoading) {
        return <div>Loading schema...</div>;
    }

    return (
        <form onSubmit={handleSubmit}>
            {schema.fields.map(field => (
                <SchemaField 
                    key={field.name}
                    field={field}
                    value={state.values[field.name]}
                    errors={errors[field.name]}
                    touched={state.touched[field.name]}
                    dirty={state.dirty[field.name]}
                    onChange={(value) => setValue(field.name, value)}
                />
            ))}
        </form>
    );
}

// Field component
function SchemaField({ field, value, errors, touched, dirty, onChange }: FieldProps) {
    const theme = useTheme();
    const tokens = theme.getComponentTokens(getFieldComponentType(field.type));
    
    switch (field.type) {
        case 'text':
            return (
                <TextInput
                    name={field.name}
                    label={field.label}
                    value={value}
                    errors={errors}
                    placeholder={field.placeholder}
                    required={field.required}
                    onChange={onChange}
                    style={tokens}
                />
            );
        case 'email':
            return (
                <EmailInput
                    name={field.name}
                    label={field.label}
                    value={value}
                    errors={errors}
                    onChange={onChange}
                    style={tokens}
                />
            );
        // ... other field types
    }
}
```

### Templ Integration

```go
// Templ component for form rendering
package components

import (
    "context"
    "github.com/niiniyare/erp/pkg/schema"
)

templ SchemaForm(s *schema.Schema, state map[string]any, errors map[string][]string) {
    <form hx-post={ s.Config.Action } hx-target="#form-results">
        <div class="form-container">
            for _, field := range s.Fields {
                @SchemaField(field, state[field.Name], errors[field.Name])
            }
        </div>
        
        for _, action := range s.Actions {
            @SchemaAction(action)
        }
    </form>
}

templ SchemaField(field schema.Field, value any, fieldErrors []string) {
    <div class="field-container">
        <label for={ field.Name }>
            { field.Label }
            if field.Required {
                <span class="required">*</span>
            }
        </label>
        
        switch field.Type {
            case schema.FieldText:
                @TextField(field, value, fieldErrors)
            case schema.FieldEmail:
                @EmailField(field, value, fieldErrors)
            case schema.FieldSelect:
                @SelectField(field, value, fieldErrors)
            default:
                @GenericField(field, value, fieldErrors)
        }
        
        if len(fieldErrors) > 0 {
            <div class="field-errors">
                for _, error := range fieldErrors {
                    <span class="error">{ error }</span>
                }
            </div>
        }
        
        if field.Help != "" {
            <small class="field-help">{ field.Help }</small>
        }
    </div>
}

templ TextField(field schema.Field, value any, errors []string) {
    <input
        type="text"
        id={ field.Name }
        name={ field.Name }
        value={ toString(value) }
        placeholder={ field.Placeholder }
        required?={ field.Required }
        class={ getFieldClasses("text", len(errors) > 0) }
        hx-post="/api/validate/field"
        hx-trigger="blur"
        hx-include="form"
    />
}
```

### Go Templates Integration

```go
// Template functions for schema rendering
func GetTemplateFuncs() template.FuncMap {
    return template.FuncMap{
        "renderField": func(field schema.Field, value any, errors []string) string {
            return renderSchemaField(field, value, errors)
        },
        "getFieldType": func(field schema.Field) string {
            return string(field.Type)
        },
        "hasErrors": func(errors []string) bool {
            return len(errors) > 0
        },
        "isRequired": func(field schema.Field) bool {
            return field.Required
        },
        "getFieldClasses": func(field schema.Field, hasErrors bool) string {
            classes := []string{"field", string(field.Type)}
            if hasErrors {
                classes = append(classes, "error")
            }
            if field.Required {
                classes = append(classes, "required")
            }
            return strings.Join(classes, " ")
        },
    }
}

// Template
// schema-form.html
{{define "schema-form"}}
<form action="{{.Schema.Config.Action}}" method="{{.Schema.Config.Method}}">
    <div class="form-container">
        {{range .Schema.Fields}}
            <div class="field-container {{getFieldClasses . (hasErrors (index $.Errors .Name))}}">
                <label for="{{.Name}}">
                    {{.Label}}
                    {{if isRequired .}}<span class="required">*</span>{{end}}
                </label>
                
                {{renderField . (index $.Values .Name) (index $.Errors .Name)}}
                
                {{if hasErrors (index $.Errors .Name)}}
                    <div class="field-errors">
                        {{range index $.Errors .Name}}
                            <span class="error">{{.}}</span>
                        {{end}}
                    </div>
                {{end}}
                
                {{if .Help}}
                    <small class="field-help">{{.Help}}</small>
                {{end}}
            </div>
        {{end}}
    </div>
    
    <div class="form-actions">
        {{range .Schema.Actions}}
            <button type="{{.Type}}" class="btn btn-{{.Variant}}">
                {{.Text}}
            </button>
        {{end}}
    </div>
</form>
{{end}}
```

### API Integration

```go
// REST API for schema operations
package api

func (h *SchemaHandler) GetSchema(c *fiber.Ctx) error {
    schemaID := c.Params("id")
    
    // Load schema from registry
    schema, err := h.registry.Get(c.Context(), schemaID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Schema not found"})
    }
    
    // Enrich schema with user context
    userCtx := getUserContext(c)
    enrichedSchema, err := h.enricher.EnrichSchema(c.Context(), schema, userCtx)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to enrich schema"})
    }
    
    return c.JSON(enrichedSchema)
}

func (h *SchemaHandler) ValidateField(c *fiber.Ctx) error {
    var req FieldValidationRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
    }
    
    // Get schema
    schema, err := h.registry.Get(c.Context(), req.SchemaID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Schema not found"})
    }
    
    // Find field
    field, exists := schema.GetField(req.FieldName)
    if !exists {
        return c.Status(400).JSON(fiber.Map{"error": "Field not found"})
    }
    
    // Validate field
    errors := h.validator.ValidateField(c.Context(), field, req.Value, req.AllData)
    
    return c.JSON(fiber.Map{
        "valid":  len(errors) == 0,
        "errors": errors,
    })
}

func (h *SchemaHandler) SubmitForm(c *fiber.Ctx) error {
    var req FormSubmissionRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
    }
    
    // Get and enrich schema
    schema, err := h.getEnrichedSchema(c.Context(), req.SchemaID)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to load schema"})
    }
    
    // Validate all data
    result, err := h.validator.ValidateDataDetailed(c.Context(), schema, req.Data)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Validation failed"})
    }
    
    if !result.Valid {
        return c.Status(400).JSON(fiber.Map{
            "valid":  false,
            "errors": result.Errors,
        })
    }
    
    // Process submission
    processedData, err := h.processor.ProcessSubmission(c.Context(), schema, result.Data)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Processing failed"})
    }
    
    return c.JSON(fiber.Map{
        "valid": true,
        "data":  processedData,
    })
}
```

---

## Best Practices

### Schema Design

1. **Use Builder Pattern**: Always use the builder for programmatic schema creation
2. **Leverage Mixins**: Create reusable field groups for common patterns
3. **Apply Business Rules**: Use business rules for complex conditional logic
4. **Version Schemas**: Always version schemas for backward compatibility
5. **Validate Early**: Use builder validation during development

```go
// Good: Using builder with validation
schema, err := schema.NewBuilder("user-form", schema.TypeForm, "User Form").
    WithVersion("1.0.0").
    AddTextField("email", "Email", true).
    WithMixin("audit_fields").
    Build(ctx)

// Bad: Manual schema creation without validation
schema := &schema.Schema{
    ID:    "user-form",
    Fields: []schema.Field{
        {Name: "email", Type: "text"}, // Missing validation
    },
}
```

### Performance Optimization

1. **Use Registry Caching**: Configure multi-tier caching for schema storage
2. **Debounce Events**: Enable debouncing for validation and change events
3. **Batch Operations**: Use batch updates for multiple field changes
4. **Optimize Conditionals**: Minimize complex conditional evaluations

```go
// Configure caching
registryConfig := &registry.Config{
    MemoryCacheTTL: 5 * time.Minute,
    RedisCacheTTL:  1 * time.Hour,
    EnableMemory:   true,
    EnableRedis:    true,
}

// Enable debouncing
runtimeConfig := &schema.RuntimeConfig{
    EnableDebounce: true,
    DebounceDelay:  300 * time.Millisecond,
}

// Batch updates
runtime.BatchUpdate(func(state *State) error {
    state.SetValue("field1", "value1")
    state.SetValue("field2", "value2")
    return nil
})
```

### Error Handling

1. **Use Typed Errors**: Always use the provided error types
2. **Handle Async Errors**: Properly handle async validation errors
3. **Provide User Feedback**: Show meaningful error messages
4. **Log for Debugging**: Log schema operations for troubleshooting

```go
// Good error handling
result, err := runtime.ValidateField(ctx, fieldName)
if err != nil {
    if validationErr, ok := err.(*schema.ValidationError); ok {
        log.Errorf("Validation failed for %s: %s", validationErr.Field, validationErr.Message)
        return handleValidationError(validationErr)
    }
    log.Errorf("Unexpected error: %v", err)
    return err
}
```

### Security Best Practices

1. **Validate on Server**: Always validate on the server side
2. **Sanitize Input**: Sanitize all user input before processing
3. **Check Permissions**: Use enricher for permission-based field access
4. **Enable CSRF Protection**: Use CSRF tokens for form submissions
5. **Rate Limit**: Configure rate limiting for form submissions

```go
// Security configuration
builder.WithCSRF().
    WithRateLimit(10, 60).
    WithTenant("companyId", "strict")

// Permission-based field access
field := schema.Field{
    Name: "sensitiveField",
    Type: schema.FieldText,
    Permissions: &schema.FieldPermissions{
        Read:  []string{"admin", "manager"},
        Write: []string{"admin"},
    },
}
```

### Testing Strategies

1. **Test Schema Validation**: Test schema structure and validation rules
2. **Test Business Rules**: Verify conditional logic and business rules
3. **Test Rendering**: Test UI rendering with different states
4. **Test Integration**: Test runtime integration with your UI framework

```go
// Schema validation test
func TestUserSchema(t *testing.T) {
    schema := createUserSchema()
    
    ctx := context.Background()
    err := schema.Validate(ctx)
    require.NoError(t, err)
    
    // Test required fields
    field, exists := schema.GetField("email")
    require.True(t, exists)
    require.True(t, field.Required)
    require.Equal(t, schema.FieldEmail, field.Type)
}

// Business rules test
func TestBusinessRules(t *testing.T) {
    rule := createDiscountRule()
    
    data := map[string]any{
        "membershipType": "premium",
        "orderAmount":    1500.0,
    }
    
    engine := schema.NewBusinessRuleEngine()
    engine.RegisterRule(rule)
    
    result, err := engine.ApplyRules(ctx, schema, data)
    require.NoError(t, err)
    require.Equal(t, 225.0, result["discountAmount"]) // 15% of 1500
}
```

---

## Troubleshooting

### Common Issues

#### 1. Schema Validation Errors

**Problem**: Schema fails validation during build
**Solution**: Check field configurations and required properties

```go
// Debug schema validation
schema, err := builder.Build(ctx)
if err != nil {
    if validationErr, ok := err.(*schema.ValidationError); ok {
        log.Printf("Field: %s, Message: %s, Code: %s", 
            validationErr.Field, validationErr.Message, validationErr.Code)
    }
}
```

#### 2. Token Resolution Failures

**Problem**: Design tokens not resolving properly
**Solution**: Check token references and circular dependencies

```go
// Debug token resolution
tokens := theme.GetComponentTokens("button")
for key, value := range tokens {
    resolved, err := theme.ResolveToken(value)
    if err != nil {
        log.Printf("Failed to resolve token %s: %v", key, err)
    }
}
```

#### 3. Runtime State Issues

**Problem**: State not updating correctly
**Solution**: Check event handlers and state initialization

```go
// Debug state operations
runtime.OnStateChange(func(ctx context.Context, changes *StateChanges) error {
    log.Printf("State changed: %+v", changes)
    return nil
})
```

#### 4. Conditional Logic Not Working

**Problem**: Fields not showing/hiding as expected
**Solution**: Debug condition evaluation

```go
// Debug conditional evaluation
visible, err := conditionalEngine.EvaluateFieldVisibility(ctx, field, data)
log.Printf("Field %s visible: %t (data: %+v)", field.Name, visible, data)
```

#### 5. Validation Timing Issues

**Problem**: Validation not triggering at the right time
**Solution**: Check validation timing configuration

```go
// Check validation configuration
config := runtime.GetConfig()
log.Printf("Validation timing: %s, Debounce: %t (%v)", 
    config.ValidationTiming, config.EnableDebounce, config.DebounceDelay)
```

### Debugging Tools

#### 1. Enable Debug Logging

```go
// Enable debug logging for schema operations
logger := log.WithField("component", "schema")
runtime.SetLogger(logger)
```

#### 2. Schema Inspector

```go
// Inspect schema structure
func InspectSchema(s *schema.Schema) {
    fmt.Printf("Schema: %s (%s)\n", s.ID, s.Type)
    fmt.Printf("Fields: %d\n", len(s.Fields))
    fmt.Printf("Actions: %d\n", len(s.Actions))
    
    for _, field := range s.Fields {
        fmt.Printf("  Field: %s (%s) Required: %t\n", 
            field.Name, field.Type, field.Required)
        
        if field.Conditional != nil {
            fmt.Printf("    Has conditionals\n")
        }
        
        if field.Validation != nil {
            fmt.Printf("    Has validation rules\n")
        }
    }
}
```

#### 3. State Debugger

```go
// Debug runtime state
func DebugRuntimeState(runtime *Runtime) {
    stats := runtime.GetStats()
    fmt.Printf("Runtime Stats:\n")
    fmt.Printf("  Fields: %d (Visible: %d, Touched: %d, Dirty: %d)\n",
        stats.FieldCount, stats.VisibleFields, stats.TouchedFields, stats.DirtyFields)
    fmt.Printf("  Errors: %d, Valid: %t\n", stats.ErrorCount, stats.IsValid)
    fmt.Printf("  Uptime: %v\n", stats.Uptime)
    
    if stats.EventStats != nil {
        fmt.Printf("  Events: %d total\n", stats.EventStats.TotalEvents)
        for eventType, count := range stats.EventStats.EventsByType {
            fmt.Printf("    %s: %d\n", eventType, count)
        }
    }
}
```

### Performance Monitoring

```go
// Monitor schema operations
type SchemaMetrics struct {
    RenderCount    int
    ValidationTime time.Duration
    StateUpdates   int
}

// Add middleware for performance tracking
func (r *Renderer) RenderFieldWithMetrics(ctx context.Context, field *schema.Field, ...) (string, error) {
    start := time.Now()
    defer func() {
        metrics.RenderCount++
        log.Printf("Field %s rendered in %v", field.Name, time.Since(start))
    }()
    
    return r.RenderField(ctx, field, value, errors, touched, dirty)
}
```

---

## Conclusion

This guide provides a comprehensive overview of the ERP Schema Package for UI implementations. The system is designed to be:

- **Flexible**: Works with any UI framework
- **Powerful**: Supports complex business requirements  
- **Performant**: Optimized for enterprise-scale applications
- **Secure**: Built with security and multi-tenancy in mind
- **Maintainable**: Clean interfaces and separation of concerns

### Next Steps

1. **Choose Your Integration**: Select React, Templ, Go Templates, or another framework
2. **Implement Core Interfaces**: Start with RuntimeRenderer for your chosen framework  
3. **Add Validation**: Implement RuntimeValidator for your validation needs
4. **Enable Conditionals**: Add RuntimeConditionalEngine for dynamic behavior
5. **Apply Theming**: Use the design token system for consistent styling
6. **Test Thoroughly**: Implement comprehensive tests for your integration

### Resources

- **Example Implementations**: See `pkg/schema/examples/` for reference implementations
- **API Documentation**: Full API docs at `/docs/api/schema/`  
- **Community Support**: Join discussions at `https://github.com/niiniyare/erp/discussions`

---

**Last Updated**: November 2025  
**Version**: 1.0.0  
**Package Version**: github.com/niiniyare/erp/pkg/schema@v1.0.0
