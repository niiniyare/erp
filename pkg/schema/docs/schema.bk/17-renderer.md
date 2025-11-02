## Section 17: Renderer System (templ Integration)

### 🎯 Learning Objectives

By the end of this section, you will understand:
- templ component architecture
- Form rendering strategy
- Field rendering patterns
- HTMX integration in templates
- Alpine.js integration in templates
- Best practices for maintainable templates

### 17.1 Renderer Architecture

**Philosophy:**
- Schema → Go structs → templ components → HTML
- templ provides type safety (no string templates)
- Components are composable and reusable
- Each field type has its own renderer

**Structure:**
```
views/
├── schema/
│   ├── form.templ          # Main form wrapper
│   ├── layout.templ        # Layout renderers (grid, tabs, steps)
│   ├── field.templ         # Field dispatcher
│   ├── fields/
│   │   ├── text.templ      # Text input
│   │   ├── email.templ     # Email input
│   │   ├── select.templ    # Select dropdown
│   │   ├── textarea.templ  # Textarea
│   │   └── ...             # Other field types
│   ├── actions.templ       # Action buttons
│   └── validation.templ    # Validation messages
```

### 17.2 Main Form Renderer

```go
// views/schema/form.templ

package schema

import (
    "github.com/yourusername/awoerp/internal/schema"
)

// FormRenderer renders a complete form from schema
templ FormRenderer(s *schema.Schema, data map[string]any) {
    <div 
        class="schema-form" 
        id={ "form-" + s.ID }
        if s.Alpine != nil && s.Alpine.Enabled {
            x-data={ s.Alpine.XData }
            if s.Alpine.XInit != "" {
                x-init={ s.Alpine.XInit }
            }
        }
    >
        // Form title and description
        <div class="form-header">
            <h2 class="form-title">{ s.Title }</h2>
            if s.Description != "" {
                <p class="form-description">{ s.Description }</p>
            }
        </div>
        
        // Actual form element
        <form
            if s.Config != nil {
                action={ s.Config.Action }
                method={ s.Config.Method }
            }
            if s.HTMX != nil && s.HTMX.Enabled {
                if s.HTMX.Post != "" {
                    hx-post={ s.HTMX.Post }
                }
                if s.HTMX.Get != "" {
                    hx-get={ s.HTMX.Get }
                }
                if s.HTMX.Target != "" {
                    hx-target={ s.HTMX.Target }
                }
                if s.HTMX.Swap != "" {
                    hx-swap={ s.HTMX.Swap }
                }
                if s.HTMX.Indicator != "" {
                    hx-indicator={ s.HTMX.Indicator }
                }
            }
        >
            // CSRF token
            if s.Security != nil && s.Security.CSRF != nil && s.Security.CSRF.Enabled {
                <input 
                    type="hidden" 
                    name={ s.Security.CSRF.FieldName }
                    value={ getCSRFToken(ctx) }
                />
            }
            
            // Render layout
            if s.Layout != nil {
                @LayoutRenderer(s.Layout, s.Fields, data)
            } else {
                // Default: simple vertical layout
                @SimpleLayout(s.Fields, data)
            }
            
            // Actions (buttons)
            if len(s.Actions) > 0 {
                <div class="form-actions">
                    for _, action := range s.Actions {
                        @ActionRenderer(&action)
                    }
                </div>
            }
        </form>
        
        // Result area (for HTMX responses)
        if s.HTMX != nil && s.HTMX.Target != "" {
            <div id={ s.HTMX.Target.TrimPrefix("#") }></div>
        }
        
        // Loading indicator
        if s.HTMX != nil && s.HTMX.Indicator != "" {
            <div id={ s.HTMX.Indicator.TrimPrefix("#") } class="htmx-indicator">
                <svg class="animate-spin h-5 w-5" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                <span>Loading...</span>
            </div>
        }
    </div>
}

// SimpleLayout renders fields in simple vertical layout
templ SimpleLayout(fields []schema.Field, data map[string]any) {
    <div class="space-y-4">
        for _, field := range fields {
            if field.Runtime != nil && field.Runtime.Visible {
                @FieldRenderer(&field, data[field.Name])
            }
        }
    </div>
}
```

### 17.3 Layout Renderers

```go
// views/schema/layout.templ

package schema

import (
    "github.com/yourusername/awoerp/internal/schema"
)

// LayoutRenderer dispatches to appropriate layout renderer
templ LayoutRenderer(layout *schema.Layout, fields []schema.Field, data map[string]any) {
    switch layout.Type {
    case schema.LayoutGrid:
        @GridLayout(layout, fields, data)
    case schema.LayoutFlex:
        @FlexLayout(layout, fields, data)
    case schema.LayoutTabs:
        @TabsLayout(layout, fields, data)
    case schema.LayoutSteps:
        @StepsLayout(layout, fields, data)
    case schema.LayoutSections:
        @SectionsLayout(layout, fields, data)
    case schema.LayoutGroups:
        @GroupsLayout(layout, fields, data)
    default:
        @SimpleLayout(fields, data)
    }
}

// GridLayout renders fields in CSS grid
templ GridLayout(layout *schema.Layout, fields []schema.Field, data map[string]any) {
    <div 
        class="grid"
        style={
            templ.SafeCSSProperty("grid-template-columns", fmt.Sprintf("repeat(%d, 1fr)", layout.Columns)),
            templ.SafeCSSProperty("gap", layout.Gap),
        }
    >
        for _, field := range fields {
            if field.Runtime != nil && field.Runtime.Visible {
                <div
                    if field.Layout != nil {
                        if field.Layout.ColSpan > 0 {
                            style={ templ.SafeCSSProperty("grid-column", fmt.Sprintf("span %d", field.Layout.ColSpan)) }
                        }
                        if field.Layout.RowSpan > 0 {
                            style={ templ.SafeCSSProperty("grid-row", fmt.Sprintf("span %d", field.Layout.RowSpan)) }
                        }
                    }
                    if field.ShowIf != "" {
                        x-show={ field.ShowIf }
                    }
                >
                    @FieldRenderer(&field, data[field.Name])
                </div>
            }
        }
    </div>
}

// TabsLayout renders fields in tabs
templ TabsLayout(layout *schema.Layout, fields []schema.Field, data map[string]any) {
    <div class="tabs" x-data="{ activeTab: 0 }">
        // Tab headers
        <div class="tab-headers">
            for i, tab := range layout.Tabs {
                <button
                    type="button"
                    class="tab-header"
                    :class="{ 'active': activeTab === { i } }"
                    @click={ fmt.Sprintf("activeTab = %d", i) }
                >
                    if tab.Icon != "" {
                        <i class={ "icon-" + tab.Icon }></i>
                    }
                    { tab.Label }
                    if tab.Badge != "" {
                        <span class="badge">{ tab.Badge }</span>
                    }
                </button>
            }
        </div>
        
        // Tab content
        <div class="tab-content">
            for i, tab := range layout.Tabs {
                <div
                    x-show={ fmt.Sprintf("activeTab === %d", i) }
                    class="tab-panel"
                >
                    <div class="space-y-4">
                        for _, fieldName := range tab.Fields {
                            for _, field := range fields {
                                if field.Name == fieldName && field.Runtime != nil && field.Runtime.Visible {
                                    @FieldRenderer(&field, data[field.Name])
                                }
                            }
                        }
                    </div>
                </div>
            }
        </div>
    </div>
}

// StepsLayout renders multi-step wizard
templ StepsLayout(layout *schema.Layout, fields []schema.Field, data map[string]any) {
    <div class="steps" x-data="{ currentStep: 0 }">
        // Step indicators
        <div class="step-indicators">
            for i, step := range layout.Steps {
                <div 
                    class="step-indicator"
                    :class="{
                        'active': currentStep === { i },
                        'completed': currentStep > { i }
                    }"
                >
                    <div class="step-number">{ fmt.Sprintf("%d", i+1) }</div>
                    <div class="step-title">{ step.Title }</div>
                </div>
            }
        </div>
        
        // Step content
        for i, step := range layout.Steps {
            <div
                x-show={ fmt.Sprintf("currentStep === %d", i) }
                class="step-content"
            >
                <h3 class="step-title">{ step.Title }</h3>
                if step.Description != "" {
                    <p class="step-description">{ step.Description }</p>
                }
                
                <div class="space-y-4">
                    for _, fieldName := range step.Fields {
                        for _, field := range fields {
                            if field.Name == fieldName && field.Runtime != nil && field.Runtime.Visible {
                                @FieldRenderer(&field, data[field.Name])
                            }
                        }
                    }
                </div>
                
                // Navigation buttons
                <div class="step-actions">
                    if i > 0 {
                        <button
                            type="button"
                            class="btn-secondary"
                            @click={ fmt.Sprintf("currentStep = %d", i-1) }
                        >
                            ← Back
                        </button>
                    }
                    
                    if i < len(layout.Steps)-1 {
                        <button
                            type="button"
                            class="btn-primary"
                            @click={ fmt.Sprintf("currentStep = %d", i+1) }
                        >
                            Next →
                        </button>
                    } else {
                        <button type="submit" class="btn-primary">
                            Submit
                        </button>
                    }
                </div>
            </div>
        }
    </div>
}

// SectionsLayout renders fields in collapsible sections
templ SectionsLayout(layout *schema.Layout, fields []schema.Field, data map[string]any) {
    <div class="sections">
        for _, section := range layout.Sections {
            <div class="section" x-data="{ collapsed: { section.Collapsed } }">
                <div class="section-header" @click="if ({ section.Collapsible }) { collapsed = !collapsed }">
                    if section.Icon != "" {
                        <i class={ "icon-" + section.Icon }></i>
                    }
                    <h3>{ section.Title }</h3>
                    if section.Collapsible {
                        <button type="button" class="collapse-toggle">
                            <span x-show="!collapsed">▼</span>
                            <span x-show="collapsed">▶</span>
                        </button>
                    }
                </div>
                
                if section.Description != "" {
                    <p class="section-description">{ section.Description }</p>
                }
                
                <div class="section-content" x-show="!collapsed">
                    <div 
                        class="grid"
                        if section.Columns > 0 {
                            style={ templ.SafeCSSProperty("grid-template-columns", fmt.Sprintf("repeat(%d, 1fr)", section.Columns)) }
                        }
                    >
                        for _, fieldName := range section.Fields {
                            for _, field := range fields {
                                if field.Name == fieldName && field.Runtime != nil && field.Runtime.Visible {
                                    @FieldRenderer(&field, data[field.Name])
                                }
                            }
                        }
                    </div>
                </div>
            </div>
        }
    </div>
}
```

### 17.4 Field Dispatcher

```go
// views/schema/field.templ

package schema

import (
    "github.com/yourusername/awoerp/internal/schema"
    "github.com/yourusername/awoerp/views/schema/fields"
)

// FieldRenderer dispatches to appropriate field renderer
templ FieldRenderer(field *schema.Field, value any) {
    // Don't render if not visible
    if field.Runtime != nil && !field.Runtime.Visible {
        return
    }
    
    // Wrapper with conditional visibility
    <div 
        class="field-wrapper"
        data-field={ field.Name }
        if field.ShowIf != "" {
            x-show={ field.ShowIf }
        }
    >
        // Dispatch to field type renderer
        switch field.Type {
        case schema.FieldText:
            @fields.TextInput(field, value)
        case schema.FieldEmail:
            @fields.EmailInput(field, value)
        case schema.FieldPassword:
            @fields.PasswordInput(field, value)
        case schema.FieldNumber:
            @fields.NumberInput(field, value)
        case schema.FieldTextarea:
            @fields.TextareaInput(field, value)
        case schema.FieldSelect:
            @fields.SelectInput(field, value)
        case schema.FieldMultiSelect:
            @fields.MultiSelectInput(field, value)
        case schema.FieldRadio:
            @fields.RadioInput(field, value)
        case schema.FieldCheckbox:
            @fields.CheckboxInput(field, value)
        case schema.FieldDate:
            @fields.DateInput(field, value)
        case schema.FieldTime:
            @fields.TimeInput(field, value)
        case schema.FieldDateTime:
            @fields.DateTimeInput(field, value)
        case schema.FieldFile:
            @fields.FileInput(field, value)
        case schema.FieldImage:
            @fields.ImageInput(field, value)
        case schema.FieldSwitch:
            @fields.SwitchInput(field, value)
        case schema.FieldSlider:
            @fields.SliderInput(field, value)
        case schema.FieldRichText:
            @fields.RichTextInput(field, value)
        case schema.FieldHidden:
            @fields.HiddenInput(field, value)
        // ... other field types
        default:
            @fields.TextInput(field, value) // Fallback
        }
    </div>
}
```

### 17.5 Individual Field Renderers

```go
// views/schema/fields/text.templ

package fields

import (
    "github.com/yourusername/awoerp/internal/schema"
)

// TextInput renders a text input field
templ TextInput(field *schema.Field, value any) {
    <div class="form-field">
        <label for={ field.Name } class="form-label">
            { field.Label }
            if field.Required {
                <span class="required">*</span>
            }
            if field.Tooltip != "" {
                <span class="tooltip" title={ field.Tooltip }>ⓘ</span>
            }
        </label>
        
        <input
            type="text"
            id={ field.Name }
            name={ field.Name }
            value={ toString(value) }
            class="form-input"
            
            if field.Placeholder != "" {
                placeholder={ field.Placeholder }
            }
            
            if field.Required {
                required
            }
            
            if field.Readonly || (field.Runtime != nil && !field.Runtime.Editable) {
                readonly
            }
            
            if field.Disabled {
                disabled
            }
            
            // Validation attributes
            if field.Validation != nil {
                if field.Validation.MinLength != nil {
                    minlength={ fmt.Sprintf("%d", *field.Validation.MinLength) }
                }
                if field.Validation.MaxLength != nil {
                    maxlength={ fmt.Sprintf("%d", *field.Validation.MaxLength) }
                }
                if field.Validation.Pattern != "" {
                    pattern={ field.Validation.Pattern }
                }
            }
            
            // HTMX attributes
            if field.HTMX != nil {
                if field.HTMX.Trigger != "" {
                    hx-trigger={ field.HTMX.Trigger }
                }
                if field.HTMX.Post != "" {
                    hx-post={ field.HTMX.Post }
                }
                if field.HTMX.Get != "" {
                    hx-get={ field.HTMX.Get }
                }
                if field.HTMX.Target != "" {
                    hx-target={ field.HTMX.Target }
                }
                if field.HTMX.Swap != "" {
                    hx-swap={ field.HTMX.Swap }
                }
            }
            
            // Alpine.js bindings
            if field.Alpine != nil {
                if field.Alpine.XModel != "" {
                    x-model={ field.Alpine.XModel }
                }
            }
        />
        
        if field.Help != "" {
            <p class="form-help">{ field.Help }</p>
        }
        
        // Error message placeholder
        <p class="form-error" id={ field.Name + "-error" }></p>
    </div>
}

// SelectInput renders a select dropdown
templ SelectInput(field *schema.Field, value any) {
    <div class="form-field">
        <label for={ field.Name } class="form-label">
            { field.Label }
            if field.Required {
                <span class="required">*</span>
            }
        </label>
        
        <select
            id={ field.Name }
            name={ field.Name }
            class="form-select"
            
            if field.Required {
                required
            }
            
            if field.Readonly || (field.Runtime != nil && !field.Runtime.Editable) {
                disabled
            }
            
            if field.Disabled {
                disabled
            }
            
            // HTMX attributes
            if field.HTMX != nil {
                if field.HTMX.Trigger != "" {
                    hx-trigger={ field.HTMX.Trigger }
                }
                if field.HTMX.Get != "" {
                    hx-get={ field.HTMX.Get }
                }
                if field.HTMX.Target != "" {
                    hx-target={ field.HTMX.Target }
                }
            }
        >
            <option value="">-- Select --</option>
            for _, option := range field.Options {
                <option 
                    value={ option.Value }
                    if toString(value) == option.Value {
                        selected
                    }
                >
                    { option.Label }
                </option>
            }
        </select>
        
        if field.Help != "" {
            <p class="form-help">{ field.Help }</p>
        }
    </div>
}

// CheckboxInput renders checkbox(es)
templ CheckboxInput(field *schema.Field, value any) {
    <div class="form-field">
        if len(field.Options) == 0 {
            // Single checkbox (boolean)
            <label class="checkbox-label">
                <input
                    type="checkbox"
                    name={ field.Name }
                    value="true"
                    class="form-checkbox"
                    if toBool(value) {
                        checked
                    }
                    if field.Required {
                        required
                    }
                    if field.Disabled {
                        disabled
                    }
                />
                { field.Label }
                if field.Required {
                    <span class="required">*</span>
                }
            </label>
        } else {
            // Multiple checkboxes
            <fieldset class="checkbox-group">
                <legend class="form-label">
                    { field.Label }
                    if field.Required {
                        <span class="required">*</span>
                    }
                </legend>
                
                for _, option := range field.Options {
                    <label class="checkbox-label">
                        <input
                            type="checkbox"
                            name={ field.Name + "[]" }
                            value={ option.Value }
                            class="form-checkbox"
                            if contains(toStringSlice(value), option.Value) {
                                checked
                            }
                            if field.Disabled {
                                disabled
                            }
                        />
                        { option.Label }
                    </label>
                }
            </fieldset>
        }
        
        if field.Help != "" {
            <p class="form-help">{ field.Help }</p>
        }
    </div>
}
```

### 17.6 Action Renderer

```go
// views/schema/actions.templ

package schema

import (
    "github.com/yourusername/awoerp/internal/schema"
)

// ActionRenderer renders an action button
templ ActionRenderer(action *schema.Action) {
    switch action.Type {
    case schema.ActionSubmit:
        <button
            type="submit"
            class={ getButtonClass(action) }
            if action.Disabled {
                disabled
            }
        >
            if action.Icon != "" {
                <i class={ "icon-" + action.Icon }></i>
            }
            { action.Text }
        </button>
        
    case schema.ActionReset:
        <button
            type="reset"
            class={ getButtonClass(action) }
            if action.Disabled {
                disabled
            }
        >
            if action.Icon != "" {
                <i class={ "icon-" + action.Icon }></i>
            }
            { action.Text }
        </button>
        
    case schema.ActionButton:
        if action.URL != "" {
            <a
                href={ action.URL }
                class={ getButtonClass(action) }
            >
                if action.Icon != "" {
                    <i class={ "icon-" + action.Icon }></i>
                }
                { action.Text }
            </a>
        } else {
            <button
                type="button"
                class={ getButtonClass(action) }
                if action.Disabled {
                    disabled
                }
                if action.Confirm != "" {
                    onclick={ fmt.Sprintf("return confirm('%s')", action.Confirm) }
                }
                if action.HTMX != nil {
                    if action.HTMX.Post != "" {
                        hx-post={ action.HTMX.Post }
                    }
                    if action.HTMX.Delete != "" {
                        hx-delete={ action.HTMX.Delete }
                    }
                    if action.HTMX.Confirm {
                        hx-confirm={ action.Confirm }
                    }
                }
            >
                if action.Icon != "" {
                    <i class={ "icon-" + action.Icon }></i>
                }
                { action.Text }
            </button>
        }
        
    case schema.ActionLink:
        <a
            href={ action.URL }
            class={ getButtonClass(action) }
        >
            if action.Icon != "" {
                <i class={ "icon-" + action.Icon }></i>
            }
            { action.Text }
        </a>
    }
}

// Helper function to get button class
func getButtonClass(action *schema.Action) string {
    base := "btn"
    
    if action.Variant != "" {
        base += " btn-" + action.Variant
    } else {
        base += " btn-primary"
    }
    
    if action.Size != "" {
        base += " btn-" + action.Size
    }
    
    return base
}
```

### 17.7 Validation Message Renderer

```go
// views/schema/validation.templ

package schema

import (
    "github.com/yourusername/awoerp/internal/schema"
)

// ValidationErrors renders validation error messages
templ ValidationErrors(errors []schema.ValidationError) {
    <div class="alert alert-danger" role="alert">
        <div class="alert-icon">
            <svg class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd"></path>
            </svg>
        </div>
        <div class="alert-content">
            <h3 class="alert-title">Validation Errors</h3>
            <ul class="alert-list">
                for _, err := range errors {
                    <li>
                        if err.Field != "" {
                            <strong>{ err.Field }:</strong>
                        }
                        { err.Message }
                    </li>
                }
            </ul>
        </div>
    </div>
}

// SuccessMessage renders success message
templ SuccessMessage(message string) {
    <div class="alert alert-success" role="alert">
        <div class="alert-icon">
            <svg class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"></path>
            </svg>
        </div>
        <div class="alert-content">
            { message }
        </div>
    </div>
}

// ErrorMessage renders error message
templ ErrorMessage(message string) {
    <div class="alert alert-danger" role="alert">
        <div class="alert-icon">
            <svg class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd"></path>
            </svg>
        </div>
        <div class="alert-content">
            { message }
        </div>
    </div>
}
```

### 17.8 Best Practices

#### 1. Keep Components Small

```go
// ✅ GOOD: Small, focused component
templ TextInput(field *schema.Field, value any) {
    // Single responsibility: render text input
}

// ❌ BAD: Monolithic component
templ AllFields(fields []schema.Field) {
    // Tries to handle all field types in one component
}
```

#### 2. Use Type Safety

```go
// ✅ GOOD: Type-safe
templ FormRenderer(s *schema.Schema, data map[string]any)

// ❌ BAD: String templates
template := `<form>{{ .Fields }}</form>`
```

#### 3. Separate Concerns

```go
// ✅ GOOD: Separate layout and field rendering
@LayoutRenderer(layout, fields, data)
@FieldRenderer(field, value)

// ❌ BAD: Mixed concerns
// Layout logic mixed with field rendering
```

#### 4. Make Components Reusable

```go
// ✅ GOOD: Reusable
templ TextInput(field *schema.Field, value any)
// Can be used anywhere

// ❌ BAD: Too specific
templ UserNameInput(value string)
// Only works for username field
```

---

