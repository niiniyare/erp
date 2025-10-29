# Modern Form System - Implementation Guide

## Overview

This is a production-ready form rendering system built with Go, templ, Alpine.js, HTMX, and the condition package. It provides enterprise-grade features including:

- **Alpine.js** for reactive client-side state management
- **HTMX** for seamless server interactions without full page reloads  
- **Condition Package** for runtime conditional field visibility
- **Templ** for type-safe HTML templates
- **Context-aware** - all functions accept `context.Context`
- **Security** - CSP nonces, CSRF protection, input validation
- **Accessibility** - WCAG 2.1 AA compliant with ARIA attributes
- **Performance** - Caching, debouncing, and optimized rendering

## Key Improvements Over Original

### 1. **Replaced JavaScript with Alpine.js**
   - ✅ Declarative reactive state management
   - ✅ No manual DOM manipulation
   - ✅ Built-in transitions and animations
   - ✅ Event handling with `@click`, `@input`, etc.
   - ✅ Conditional rendering with `x-show`, `x-if`

### 2. **HTMX Integration**
   - ✅ Partial page updates without full reloads
   - ✅ Built-in loading states
   - ✅ Automatic request headers
   - ✅ Progressive enhancement
   - ✅ Form validation before submission

### 3. **Condition Package Integration**
   - ✅ Server-side field visibility evaluation
   - ✅ Complex boolean logic (AND/OR/NOT)
   - ✅ 17+ comparison operators
   - ✅ Safe execution with timeouts
   - ✅ Custom function support

### 4. **Context-Aware Functions**
   - ✅ All functions accept `context.Context` as first parameter
   - ✅ Proper context propagation
   - ✅ Cancellation support
   - ✅ Request scoping for middleware data

### 5. **Better Code Organization**
   - ✅ Separated concerns (types, templates, examples)
   - ✅ Reusable components
   - ✅ Custom renderer system
   - ✅ Plugin architecture

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         Browser                              │
├─────────────────────────────────────────────────────────────┤
│  Alpine.js Store (State)  │  HTMX (Ajax)  │  Native HTML5  │
│  - Form state             │  - Requests   │  - Validation  │
│  - Validation errors      │  - Updates    │  - Events      │
│  - Auto-save              │  - Loading    │                │
└──────────────┬────────────────────┬─────────────────────────┘
               │                    │
               ▼                    ▼
┌──────────────────────────────────────────────────────────────┐
│                      Go Backend                               │
├──────────────────────────────────────────────────────────────┤
│  Schema Definition  →  Condition Eval  →  Templ Render      │
│  - Fields          │  - Visibility     │  - HTML Output     │
│  - Validation      │  - Logic rules    │  - Alpine attrs    │
│  - Layout          │  - Custom funcs   │  - HTMX attrs      │
└──────────────────────────────────────────────────────────────┘
```

## Quick Start

### 1. Install Dependencies

```bash
# Go dependencies
go get github.com/niiniyare/erp/condition
go get github.com/a-h/templ

# Generate templ files
templ generate
```

### 2. Basic Form Example

```go
package main

import (
    "context"
    "net/http"
    "yourapp/schema"
)

func UserFormHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Add CSP nonce
    ctx = context.WithValue(ctx, "csp_nonce", generateNonce())
    
    // Create schema
    formSchema := &schema.Schema{
        ID:    "user-form",
        Title: "User Registration",
        Fields: []schema.Field{
            {
                Name:  "email",
                Label: "Email",
                Type:  "email",
                Validation: &schema.Validation{
                    Required: true,
                },
            },
            {
                Name:  "password",
                Label: "Password",
                Type:  "password",
                Validation: &schema.Validation{
                    Required:  true,
                    MinLength: intPtr(8),
                },
            },
        },
        Actions: []schema.Action{
            {
                ID:      "submit",
                Type:    schema.ActionSubmit,
                Text:    "Register",
                Variant: "primary",
            },
        },
        HTMX: &schema.HTMXConfig{
            Enabled: true,
            Post:    "/api/register",
        },
    }
    
    // Render form
    data := map[string]any{}
    errors := map[string][]string{}
    
    component := schema.FormTemplate(
        ctx,
        formSchema,
        data,
        errors,
        nil, // renderer registry
        nil, // token resolver
    )
    
    component.Render(ctx, w)
}
```

### 3. Conditional Fields

```go
// Field only shows when userType is "business"
{
    Name:  "companyName",
    Label: "Company Name",
    Type:  "text",
    Conditions: &schema.FieldConditions{
        Show: []schema.Condition{
            {
                Field:    "userType",
                Operator: "eq",
                Value:    "business",
            },
        },
    },
}
```

### 4. Multi-Step Form

```go
schema := &schema.Schema{
    Layout: &schema.Layout{
        Steps: []schema.Step{
            {
                ID:     "step-1",
                Title:  "Personal Info",
                Fields: []string{"firstName", "lastName"},
            },
            {
                ID:     "step-2",
                Title:  "Contact Info",
                Fields: []string{"email", "phone"},
            },
        },
    },
}
```

### 5. Tabbed Layout

```go
schema := &schema.Schema{
    Layout: &schema.Layout{
        Tabs: []schema.Tab{
            {
                ID:     "profile",
                Title:  "Profile",
                Icon:   "👤",
                Fields: []string{"name", "bio"},
            },
            {
                ID:     "security",
                Title:  "Security",
                Icon:   "🔒",
                Fields: []string{"password", "twoFactor"},
            },
        },
    },
}
```

## Advanced Features

### Custom Field Renderers

```go
type DatePickerRenderer struct{}

func (d *DatePickerRenderer) Render(field *schema.Field, value any, errors []string) (string, error) {
    return fmt.Sprintf(`
        <div x-data="{ date: '%v' }">
            <input 
                type="date" 
                name="%s"
                x-model="date"
                @change="$dispatch('date-changed', { date: $el.value })"
            />
        </div>
    `, value, field.Name), nil
}

func (d *DatePickerRenderer) RequiredAssets() []string {
    return []string{"datepicker.css", "datepicker.js"}
}

// Register
registry := schema.NewRendererRegistry()
registry.Register("date", &DatePickerRenderer{})
```

### Complex Conditional Logic with Condition Package

```go
import "github.com/niiniyare/erp/condition"

func EvaluateComplexRule(ctx context.Context, data map[string]any) (bool, error) {
    // Build complex rule: (age >= 18 AND country IN ["US", "CA"]) OR premium == true
    builder := condition.NewBuilder(condition.ConjunctionOr)
    
    // Group 1: age and country check
    group1 := condition.NewBuilder(condition.ConjunctionAnd)
    group1.AddRule("age", condition.OpGreaterOrEqual, 18)
    group1.AddRule("country", condition.OpIn, "US", "CA")
    
    builder.AddGroup(group1.Build())
    
    // Condition 2: premium check
    builder.AddRule("premium", condition.OpEqual, true)
    
    rule := builder.Build()
    
    // Evaluate
    evalCtx := condition.NewEvalContext(data, condition.DefaultEvalOptions())
    evaluator := condition.NewEvaluator(nil, condition.DefaultEvalOptions())
    
    return evaluator.Evaluate(ctx, rule, evalCtx)
}
```

### Auto-save with Alpine.js

```html
<input 
    x-data="{ value: '' }"
    x-model="value"
    @input.debounce.1000ms="$autosave('fieldName')(value)"
    name="fieldName"
/>
```

### Custom Validation with HTMX

```go
{
    Name:  "username",
    Label: "Username",
    Type:  "text",
    Attrs: map[string]string{
        "hx-post":       "/api/validate/username",
        "hx-trigger":    "blur",
        "hx-target":     "#username-validation",
        "hx-swap":       "innerHTML",
    },
}
```

## HTML Layout Structure

Your base layout should include Alpine.js, HTMX, and the form scripts:

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Forms</title>
    
    <!-- Alpine.js -->
    <script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
    
    <!-- HTMX -->
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
</head>
<body>
    <!-- Your form renders here -->
    <main id="app">
        {{ template "form" . }}
    </main>
    
    <!-- Form scripts bundle -->
    {{ template "FormScriptsBundle" }}
</body>
</html>
```

## Security Best Practices

### 1. CSP Nonces

```go
func CSPMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        nonce := generateNonce()
        ctx := context.WithValue(r.Context(), "csp_nonce", nonce)
        
        w.Header().Set("Content-Security-Policy", 
            fmt.Sprintf("script-src 'nonce-%s' 'strict-dynamic'", nonce))
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### 2. CSRF Protection

```go
schema := &schema.Schema{
    Security: &schema.Security{
        CSRF: &schema.CSRFConfig{
            Enabled:   true,
            FieldName: "_csrf",
        },
    },
}
```

### 3. Input Validation

```go
{
    Validation: &schema.Validation{
        Required:  true,
        Pattern:   `^[a-zA-Z0-9]+$`,
        MinLength: intPtr(3),
        MaxLength: intPtr(50),
        Message:   "Username must be 3-50 alphanumeric characters",
    },
}
```

## Performance Optimization

### 1. Condition Caching

```go
evaluator := condition.NewEvaluator(nil, condition.EvalOptions{
    CacheResults: true,  // Enable formula compilation cache
    Timeout:      100 * time.Millisecond,
})
```

### 2. Asset Bundling

Group related CSS/JS files together to reduce HTTP requests:

```go
registry.Register("rich-text", &RichTextRenderer{})
// Returns: ["rich-text.bundle.css", "rich-text.bundle.js"]
```

### 3. Lazy Loading Sections

```html
<section 
    x-data="{ loaded: false }"
    x-init="$nextTick(() => loaded = true)"
    x-show="loaded"
    x-transition
>
    <!-- Heavy content loads after initial render -->
</section>
```

## Testing

### Unit Tests

```go
func TestFieldVisibility(t *testing.T) {
    ctx := context.Background()
    
    field := &schema.Field{
        Name: "companyName",
        Conditions: &schema.FieldConditions{
            Show: []schema.Condition{
                {Field: "userType", Operator: "eq", Value: "business"},
            },
        },
    }
    
    data := map[string]any{"userType": "business"}
    
    visible := schema.ShouldRenderField(ctx, field, data)
    assert.True(t, visible)
}
```

### Integration Tests

```go
func TestFormSubmission(t *testing.T) {
    handler := http.HandlerFunc(UserFormHandler)
    
    req := httptest.NewRequest("POST", "/form", strings.NewReader("email=test@example.com"))
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    
    rr := httptest.NewRecorder()
    handler.ServeHTTP(rr, req)
    
    assert.Equal(t, http.StatusOK, rr.Code)
}
```

## Migration from Original Code

### Before (Inline JavaScript)
```html
<button onclick="FormUI.toggleSection('section-1')">Toggle</button>
```

### After (Alpine.js)
```html
<div x-data="{ expanded: true }">
    <button @click="expanded = !expanded">Toggle</button>
    <div x-show="expanded" x-transition>Content</div>
</div>
```

### Before (Manual DOM)
```javascript
document.getElementById('field').classList.add('error');
```

### After (Reactive)
```html
<input 
    :class="{ 'input--error': $store.form.errors['field'] }"
    @blur="$validate('field')"
/>
```

## Debugging

### Alpine.js DevTools

```javascript
// In browser console
Alpine.store('form')  // Inspect form state
```

### HTMX Events

```javascript
// Log all HTMX events
htmx.on('htmx:afterRequest', (e) => {
    console.log('Request completed:', e.detail);
});
```

### Condition Evaluation

```go
metrics := evalCtx.GetMetrics()
log.Printf("Evaluated %d rules in %v", 
    metrics.RulesEvaluated, 
    metrics.Duration)
```

## Common Pitfalls

### ❌ Not propagating context
```go
// Bad
func helper() {
    // No context
}

// Good  
func helper(ctx context.Context) {
    // Context available
}
```

### ❌ Missing Alpine.js initialization
```html
<!-- Bad: Alpine runs before script loads -->
<script src="alpine.js"></script>

<!-- Good: Defer loading -->
<script defer src="alpine.js"></script>
```

### ❌ Blocking condition evaluation
```go
// Bad: No timeout
evaluator := condition.NewEvaluator(nil, condition.EvalOptions{})

// Good: Set reasonable timeout
evaluator := condition.NewEvaluator(nil, condition.EvalOptions{
    Timeout: 100 * time.Millisecond,
})
```

## Resources

- [Alpine.js Documentation](https://alpinejs.dev)
- [HTMX Documentation](https://htmx.org)
- [Templ Documentation](https://templ.guide)
- [Condition Package](https://github.com/niiniyare/erp/tree/dev/condition)


### example usage 

```go 
package schema

import (
	"context"
	"net/http"
	"time"

	"github.com/niiniyare/erp/condition"
)

// ExampleFormHandler demonstrates how to use the form template
func ExampleFormHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Add CSP nonce to context
	ctx = context.WithValue(ctx, "csp_nonce", generateCSPNonce())
	
	// Create schema with conditional fields
	schema := CreateUserRegistrationForm(ctx)
	
	// Prepare data
	data := map[string]any{
		"email":      "",
		"userType":   "individual",
		"country":    "US",
		"_csrf":      getCSRFToken(r),
	}
	
	// Errors from previous submission
	errors := map[string][]string{}
	
	// Create renderer registry
	registry := NewRendererRegistry()
	registry.Register("select", &SelectRenderer{})
	registry.Register("textarea", &TextAreaRenderer{})
	
	// Render form
	component := FormTemplate(ctx, schema, data, errors, registry, nil)
	component.Render(ctx, w)
}

// CreateUserRegistrationForm creates a complex form with conditions
func CreateUserRegistrationForm(ctx context.Context) *Schema {
	return &Schema{
		ID:          "user-registration",
		Title:       "User Registration",
		Description: "Create your account to get started",
		Fields: []Field{
			{
				Name:        "email",
				Label:       "Email Address",
				Type:        "email",
				Placeholder: "you@example.com",
				HelpText:    "We'll never share your email",
				Validation: &Validation{
					Required:  true,
					Pattern:   `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
					Message:   "Please enter a valid email address",
				},
			},
			{
				Name:        "password",
				Label:       "Password",
				Type:        "password",
				Placeholder: "••••••••",
				Validation: &Validation{
					Required:  true,
					MinLength: intPtr(8),
					MaxLength: intPtr(128),
					Pattern:   `^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$`,
					Message:   "Password must be at least 8 characters with uppercase, lowercase, number, and special character",
				},
			},
			{
				Name:  "userType",
				Label: "Account Type",
				Type:  "select",
				Options: []FieldOption{
					{Label: "Individual", Value: "individual", Selected: true},
					{Label: "Business", Value: "business"},
					{Label: "Enterprise", Value: "enterprise"},
				},
				Validation: &Validation{
					Required: true,
				},
			},
			// Conditional field: Only show for business/enterprise
			{
				Name:        "companyName",
				Label:       "Company Name",
				Type:        "text",
				Placeholder: "Acme Corp",
				Conditions: &FieldConditions{
					Show: []Condition{
						{
							Field:    "userType",
							Operator: "in",
							Value:    []string{"business", "enterprise"},
						},
					},
				},
				Validation: &Validation{
					Required: true,
				},
			},
			// Conditional field: Only for enterprise
			{
				Name:     "employeeCount",
				Label:    "Number of Employees",
				Type:     "number",
				HelpText: "Approximate number of employees",
				Conditions: &FieldConditions{
					Show: []Condition{
						{
							Field:    "userType",
							Operator: "eq",
							Value:    "enterprise",
						},
					},
				},
				Validation: &Validation{
					Required: true,
					Min:      float64Ptr(1),
					Max:      float64Ptr(1000000),
				},
			},
			{
				Name:  "country",
				Label: "Country",
				Type:  "select",
				Options: []FieldOption{
					{Label: "United States", Value: "US"},
					{Label: "Canada", Value: "CA"},
					{Label: "United Kingdom", Value: "GB"},
					{Label: "Germany", Value: "DE"},
				},
				Validation: &Validation{
					Required: true,
				},
			},
			{
				Name:  "agreeToTerms",
				Label: "I agree to the Terms of Service and Privacy Policy",
				Type:  "checkbox",
				Validation: &Validation{
					Required: true,
					Message:  "You must agree to the terms to continue",
				},
			},
		},
		Actions: []Action{
			{
				ID:      "submit",
				Type:    ActionSubmit,
				Text:    "Create Account",
				Variant: "primary",
				HTMX: &ActionHTMX{
					Post:   "/api/register",
					Target: "#registration-form",
					Swap:   "outerHTML",
				},
			},
			{
				ID:      "cancel",
				Type:    ActionButton,
				Text:    "Cancel",
				Variant: "secondary",
				Config: &ActionConfig{
					URL: "/",
				},
			},
		},
		Config: &FormConfig{
			Action:   "/api/register",
			Method:   "POST",
			Encoding: "application/json",
		},
		HTMX: &HTMXConfig{
			Enabled:  true,
			Post:     "/api/register",
			Target:   "#registration-form",
			Swap:     "outerHTML",
			Validate: true,
		},
		Alpine: &AlpineConfig{
			Enabled: true,
		},
		Security: &Security{
			CSRF: &CSRFConfig{
				Enabled:   true,
				FieldName: "_csrf",
			},
		},
		Theme: &Theme{
			PrimaryColor: "#3b82f6",
			Spacing:      "1rem",
			BorderRadius: "0.5rem",
		},
	}
}

// CreateMultiStepForm creates a multi-step form example
func CreateMultiStepForm(ctx context.Context) *Schema {
	return &Schema{
		ID:          "onboarding",
		Title:       "Account Setup",
		Description: "Complete your profile in 3 easy steps",
		Fields: []Field{
			// Step 1: Personal Info
			{Name: "firstName", Label: "First Name", Type: "text", Validation: &Validation{Required: true}},
			{Name: "lastName", Label: "Last Name", Type: "text", Validation: &Validation{Required: true}},
			{Name: "email", Label: "Email", Type: "email", Validation: &Validation{Required: true}},
			
			// Step 2: Company Info
			{Name: "company", Label: "Company", Type: "text", Validation: &Validation{Required: true}},
			{Name: "role", Label: "Role", Type: "text", Validation: &Validation{Required: true}},
			{Name: "industry", Label: "Industry", Type: "select", Options: []FieldOption{
				{Label: "Technology", Value: "tech"},
				{Label: "Finance", Value: "finance"},
				{Label: "Healthcare", Value: "healthcare"},
			}},
			
			// Step 3: Preferences
			{Name: "notifications", Label: "Email Notifications", Type: "checkbox"},
			{Name: "newsletter", Label: "Subscribe to Newsletter", Type: "checkbox"},
			{Name: "language", Label: "Preferred Language", Type: "select", Options: []FieldOption{
				{Label: "English", Value: "en", Selected: true},
				{Label: "Spanish", Value: "es"},
				{Label: "French", Value: "fr"},
			}},
		},
		Layout: &Layout{
			Steps: []Step{
				{
					ID:          "step-personal",
					Title:       "Personal Information",
					Description: "Tell us about yourself",
					Fields:      []string{"firstName", "lastName", "email"},
					Validation:  true,
				},
				{
					ID:          "step-company",
					Title:       "Company Details",
					Description: "Where do you work?",
					Fields:      []string{"company", "role", "industry"},
					Validation:  true,
				},
				{
					ID:          "step-preferences",
					Title:       "Preferences",
					Description: "Customize your experience",
					Fields:      []string{"notifications", "newsletter", "language"},
				},
			},
		},
		Actions: []Action{
			{
				ID:      "submit",
				Type:    ActionSubmit,
				Text:    "Complete Setup",
				Variant: "primary",
			},
		},
		HTMX: &HTMXConfig{
			Enabled: true,
			Post:    "/api/onboarding",
		},
		Alpine: &AlpineConfig{
			Enabled: true,
		},
	}
}

// CreateTabbedForm creates a tabbed form example
func CreateTabbedForm(ctx context.Context) *Schema {
	return &Schema{
		ID:          "settings",
		Title:       "Account Settings",
		Description: "Manage your account preferences",
		Fields: []Field{
			// Profile tab
			{Name: "avatar", Label: "Profile Picture", Type: "file"},
			{Name: "bio", Label: "Bio", Type: "textarea", Validation: &Validation{MaxLength: intPtr(500)}},
			{Name: "website", Label: "Website", Type: "url"},
			
			// Security tab
			{Name: "currentPassword", Label: "Current Password", Type: "password"},
			{Name: "newPassword", Label: "New Password", Type: "password"},
			{Name: "confirmPassword", Label: "Confirm Password", Type: "password"},
			{Name: "twoFactor", Label: "Enable Two-Factor Authentication", Type: "checkbox"},
			
			// Notifications tab
			{Name: "emailNotifications", Label: "Email Notifications", Type: "checkbox"},
			{Name: "smsNotifications", Label: "SMS Notifications", Type: "checkbox"},
			{Name: "pushNotifications", Label: "Push Notifications", Type: "checkbox"},
		},
		Layout: &Layout{
			Tabs: []Tab{
				{
					ID:     "tab-profile",
					Title:  "Profile",
					Icon:   "👤",
					Fields: []string{"avatar", "bio", "website"},
				},
				{
					ID:     "tab-security",
					Title:  "Security",
					Icon:   "🔒",
					Fields: []string{"currentPassword", "newPassword", "confirmPassword", "twoFactor"},
				},
				{
					ID:     "tab-notifications",
					Title:  "Notifications",
					Icon:   "🔔",
					Fields: []string{"emailNotifications", "smsNotifications", "pushNotifications"},
				},
			},
		},
		Actions: []Action{
			{
				ID:      "save",
				Type:    ActionSubmit,
				Text:    "Save Changes",
				Variant: "primary",
				HTMX: &ActionHTMX{
					Post:   "/api/settings",
					Target: "#settings",
				},
			},
		},
		HTMX: &HTMXConfig{
			Enabled: true,
		},
		Alpine: &AlpineConfig{
			Enabled: true,
		},
	}
}

// Example of using condition package for complex field visibility
func EvaluateFieldVisibility(ctx context.Context, field *Field, data map[string]any) (bool, error) {
	if field.Conditions == nil || len(field.Conditions.Show) == 0 {
		return true, nil
	}
	
	// Build condition group
	builder := condition.NewBuilder(condition.ConjunctionAnd)
	
	for _, cond := range field.Conditions.Show {
		var op condition.OperatorType
		switch cond.Operator {
		case "eq":
			op = condition.OpEqual
		case "ne":
			op = condition.OpNotEqual
		case "gt":
			op = condition.OpGreater
		case "gte":
			op = condition.OpGreaterOrEqual
		case "lt":
			op = condition.OpLess
		case "lte":
			op = condition.OpLessOrEqual
		case "in":
			op = condition.OpIn
		case "contains":
			op = condition.OpContains
		default:
			op = condition.OpEqual
		}
		
		builder.AddRule(cond.Field, op, cond.Value)
	}
	
	group := builder.Build()
	
	// Create evaluator with timeout
	evaluator := condition.NewEvaluator(nil, condition.EvalOptions{
		MaxDepth:      10,
		MaxConditions: 100,
		Timeout:       100 * time.Millisecond,
		CacheResults:  true,
	})
	
	// Create evaluation context
	evalCtx := condition.NewEvalContext(data, condition.DefaultEvalOptions())
	
	// Evaluate
	return evaluator.Evaluate(ctx, group, evalCtx)
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}

func generateCSPNonce() string {
	// Implementation would generate a secure random nonce
	return "random-nonce-123"
}

func getCSRFToken(r *http.Request) string {
	// Implementation would get CSRF token from session
	return "csrf-token-123"
}

// Example custom renderers
type SelectRenderer struct{}

func (s *SelectRenderer) Render(field *Field, value any, errors []string) (string, error) {
	hasError := len(errors) > 0
	errorClass := ""
	if hasError {
		errorClass = " select--error"
	}
	
	html := fmt.Sprintf(`
		<label for="%s" class="field-label">
			%s
			%s
		</label>
		<select 
			id="%s" 
			name="%s" 
			class="select%s"
			@change="$el.dispatchEvent(new CustomEvent('field-change', { detail: { field: '%s', value: $el.value }, bubbles: true }))"
		>`,
		field.Name,
		field.Label,
		func() string {
			if field.Validation != nil && field.Validation.Required {
				return `<span class="field-required" aria-label="required">*</span>`
			}
			return ""
		}(),
		field.Name,
		field.Name,
		errorClass,
		field.Name,
	)
	
	for _, opt := range field.Options {
		selected := ""
		if opt.Value == fmt.Sprintf("%v", value) || opt.Selected {
			selected = " selected"
		}
		html += fmt.Sprintf(`<option value="%s"%s>%s</option>`, opt.Value, selected, opt.Label)
	}
	
	html += `</select>`
	return html, nil
}

func (s *SelectRenderer) RequiredAssets() []string {
	return []string{}
}

type TextAreaRenderer struct{}

func (t *TextAreaRenderer) Render(field *Field, value any, errors []string) (string, error) {
	hasError := len(errors) > 0
	errorClass := ""
	if hasError {
		errorClass = " textarea--error"
	}
	
	valueStr := ""
	if value != nil {
		valueStr = fmt.Sprintf("%v", value)
	}
	
	rows := 4
	if field.Attrs != nil {
		if r, ok := field.Attrs["rows"]; ok {
			fmt.Sscanf(r, "%d", &rows)
		}
	}
	
	return fmt.Sprintf(`
		<label for="%s" class="field-label">
			%s
			%s
		</label>
		<textarea 
			id="%s" 
			name="%s" 
			class="textarea%s"
			rows="%d"
			placeholder="%s"
			@input="$el.dispatchEvent(new CustomEvent('field-change', { detail: { field: '%s', value: $el.value }, bubbles: true }))"
		>%s</textarea>`,
		field.Name,
		field.Label,
		func() string {
			if field.Validation != nil && field.Validation.Required {
				return `<span class="field-required" aria-label="required">*</span>`
			}
			return ""
		}(),
		field.Name,
		field.Name,
		errorClass,
		rows,
		field.Placeholder,
		field.Name,
		valueStr,
	), nil
}

func (t *TextAreaRenderer) RequiredAssets() []string {
	return []string{}
}
```
