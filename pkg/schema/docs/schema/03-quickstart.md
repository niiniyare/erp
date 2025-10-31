## Section 3: Quick Start (5 Minutes)

### 🎯 Learning Objectives

By the end of this section, you will:
- Create your first schema in JSON
- Write a simple HTTP handler
- Create a templ component
- See a working form in the browser

### 3.1 Installation

**Prerequisites:**
- Go 1.21 or later
- templ CLI
- (Optional) PostgreSQL for schema storage
- (Optional) Redis for caching

**Install Dependencies:**
```bash
# Install templ
go install github.com/a-h/templ/cmd/templ@latest

# Install schema package
go get github.com/niiniyare/erp/pkg/schema

# Install validator
go get github.com/go-playground/validator/v10

# Install HTMX and Alpine.js (via CDN in HTML)
# No installation needed
```

### 3.2 Create Your First Schema

**File: `schemas/contact-form.json`**
```json
{
  "id": "contact-form",
  "type": "form",
  "version": "1.0.0",
  "title": "Contact Us",
  "description": "Send us a message and we'll get back to you soon",
  "config": {
    "action": "/api/contact",
    "method": "POST"
  },
  "fields": [
    {
      "name": "name",
      "type": "text",
      "label": "Your Name",
      "placeholder": "John Doe",
      "required": true,
      "validation": {
        "minLength": 2,
        "maxLength": 100,
        "messages": {
          "required": "Please enter your name",
          "minLength": "Name must be at least 2 characters"
        }
      }
    },
    {
      "name": "email",
      "type": "email",
      "label": "Email Address",
      "placeholder": "john@example.com",
      "required": true,
      "validation": {
        "maxLength": 255,
        "messages": {
          "required": "Please enter your email",
          "pattern": "Please enter a valid email address"
        }
      }
    },
    {
      "name": "subject",
      "type": "select",
      "label": "Subject",
      "required": true,
      "options": [
        {"value": "general", "label": "General Inquiry"},
        {"value": "support", "label": "Technical Support"},
        {"value": "sales", "label": "Sales Question"},
        {"value": "other", "label": "Other"}
      ]
    },
    {
      "name": "message",
      "type": "textarea",
      "label": "Message",
      "placeholder": "Tell us how we can help...",
      "required": true,
      "config": {
        "rows": 5
      },
      "validation": {
        "minLength": 10,
        "maxLength": 1000,
        "messages": {
          "required": "Please enter a message",
          "minLength": "Message must be at least 10 characters"
        }
      }
    }
  ],
  "actions": [
    {
      "id": "submit",
      "type": "submit",
      "text": "Send Message",
      "variant": "primary",
      "size": "md"
    }
  ],
  "htmx": {
    "enabled": true,
    "post": "/api/contact",
    "target": "#result",
    "swap": "innerHTML"
  }
}
```

### 3.3 Create HTTP Handler

**File: `internal/handlers/contact.go`**
```go
package handlers

import (
    "context"
    "net/http"
    "github.com/niiniyare/erp/pkg/schema"
    "github.com/niiniyare/erp/pkg/schema/parse"
    "github.com/niiniyare/erp/pkg/schema/registry"
    "github.com/niiniyare/erp/views"
    "os"
)

// HandleContactForm displays the contact form
func HandleContactForm(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Load schema from file
    jsonData, err := os.ReadFile("schemas/contact-form.json")
    if err != nil {
        http.Error(w, "Schema not found", 404)
        return
    }
    
    // Parse schema
    parser := parse.NewParser()
    s, err := parser.Parse(jsonData)
    if err != nil {
        http.Error(w, "Invalid schema: "+err.Error(), 500)
        return
    }
    
    // Render form (no data for new form)
    views.FormPage(s, nil).Render(ctx, w)
}

// HandleContactSubmit processes form submission
func HandleContactSubmit(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Parse form data
    if err := r.ParseForm(); err != nil {
        http.Error(w, "Invalid form data", 400)
        return
    }
    
    data := map[string]interface{}{
        "name":    r.FormValue("name"),
        "email":   r.FormValue("email"),
        "subject": r.FormValue("subject"),
        "message": r.FormValue("message"),
    }
    
    // Load schema for validation
    jsonData, _ := os.ReadFile("schemas/contact-form.json")
    parser := parse.NewParser()
    s, _ := parser.Parse(jsonData)
    
    // Validate
    validator := validate.NewValidator(nil)
    errors := validator.ValidateData(ctx, s, data)
    
    if len(errors) > 0 {
        // Return errors
        views.ValidationErrors(errors).Render(ctx, w)
        return
    }
    
    // TODO: Save to database, send email, etc.
    
    // Return success
    views.SuccessMessage("Thank you! We'll get back to you soon.").Render(ctx, w)
}
```

### 3.4 Create templ Components

**File: `views/form.templ`**
```go
package views

import "github.com/niiniyare/erp/pkg/schema"

// FormPage renders a complete form page
templ FormPage(s *schema.Schema, data map[string]interface{}) {
    <!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8"/>
        <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
        <title>{ s.Title }</title>
        
        <!-- HTMX -->
        <script src="https://unpkg.com/htmx.org@1.9.10"></script>
        
        <!-- Alpine.js -->
        <script defer src="https://unpkg.com/alpinejs@3.13.3/dist/cdn.min.js"></script>
        
        <!-- Simple CSS -->
        <style>
            body { font-family: system-ui; max-width: 600px; margin: 2rem auto; padding: 0 1rem; }
            .field { margin-bottom: 1.5rem; }
            label { display: block; margin-bottom: 0.5rem; font-weight: 500; }
            input, select, textarea { width: 100%; padding: 0.5rem; border: 1px solid #ddd; border-radius: 4px; }
            input:focus, select:focus, textarea:focus { outline: none; border-color: #0066cc; }
            button { background: #0066cc; color: white; padding: 0.75rem 1.5rem; border: none; border-radius: 4px; cursor: pointer; }
            button:hover { background: #0052a3; }
            .error { color: #dc3545; font-size: 0.875rem; margin-top: 0.25rem; }
            .required { color: #dc3545; }
            #result { margin-top: 1rem; padding: 1rem; border-radius: 4px; }
            .success { background: #d4edda; color: #155724; }
        </style>
    </head>
    <body>
        <h1>{ s.Title }</h1>
        if s.Description != "" {
            <p>{ s.Description }</p>
        }
        
        @FormRenderer(s, data)
        
        <div id="result"></div>
    </body>
    </html>
}

// FormRenderer renders the form element
templ FormRenderer(s *schema.Schema, data map[string]interface{}) {
    <form
        if s.HTMX != nil && s.HTMX.Enabled {
            hx-post={ s.HTMX.Post }
            hx-target={ s.HTMX.Target }
            hx-swap={ s.HTMX.Swap }
        } else if s.Config != nil {
            action={ s.Config.Action }
            method={ s.Config.Method }
        }
    >
        for _, field := range s.Fields {
            @FieldRenderer(&field, data[field.Name])
        }
        
        for _, action := range s.Actions {
            if action.Type == schema.ActionSubmit {
                <button type="submit">{ action.Text }</button>
            }
        }
    </form>
}

// FieldRenderer delegates to type-specific renderers
templ FieldRenderer(field *schema.Field, value interface{}) {
    <div class="field">
        switch field.Type {
            case schema.FieldText, schema.FieldEmail:
                @TextInput(field, toString(value))
            case schema.FieldSelect:
                @SelectInput(field, toString(value))
            case schema.FieldTextarea:
                @TextareaInput(field, toString(value))
        }
    </div>
}

// TextInput renders text/email inputs
templ TextInput(field *schema.Field, value string) {
    <label for={ field.Name }>
        { field.Label }
        if field.Required {
            <span class="required">*</span>
        }
    </label>
    <input
        type={ string(field.Type) }
        name={ field.Name }
        id={ field.Name }
        value={ value }
        if field.Placeholder != "" {
            placeholder={ field.Placeholder }
        }
        if field.Required {
            required
        }
        if field.Validation != nil {
            if field.Validation.MinLength != nil {
                minlength={ strconv.Itoa(*field.Validation.MinLength) }
            }
            if field.Validation.MaxLength != nil {
                maxlength={ strconv.Itoa(*field.Validation.MaxLength) }
            }
        }
        x-data="{ error: '' }"
        @invalid.prevent="error = $el.validationMessage"
        @input="error = ''"
    />
    <span class="error" x-show="error" x-text="error"></span>
}

// SelectInput renders select dropdowns
templ SelectInput(field *schema.Field, value string) {
    <label for={ field.Name }>
        { field.Label }
        if field.Required {
            <span class="required">*</span>
        }
    </label>
    <select
        name={ field.Name }
        id={ field.Name }
        if field.Required {
            required
        }
    >
        <option value="">-- Select --</option>
        for _, opt := range field.Options {
            <option 
                value={ opt.Value }
                if opt.Value == value {
                    selected
                }
            >
                { opt.Label }
            </option>
        }
    </select>
}

// TextareaInput renders textarea fields
templ TextareaInput(field *schema.Field, value string) {
    <label for={ field.Name }>
        { field.Label }
        if field.Required {
            <span class="required">*</span>
        }
    </label>
    <textarea
        name={ field.Name }
        id={ field.Name }
        if field.Placeholder != "" {
            placeholder={ field.Placeholder }
        }
        if field.Required {
            required
        }
        if field.Config != nil {
            if rows, ok := field.Config["rows"].(float64); ok {
                rows={ strconv.Itoa(int(rows)) }
            }
        }
        if field.Validation != nil {
            if field.Validation.MinLength != nil {
                minlength={ strconv.Itoa(*field.Validation.MinLength) }
            }
            if field.Validation.MaxLength != nil {
                maxlength={ strconv.Itoa(*field.Validation.MaxLength) }
            }
        }
    >{ value }</textarea>
}

// Helper: convert interface{} to string
func toString(v interface{}) string {
    if v == nil {
        return ""
    }
    if s, ok := v.(string); ok {
        return s
    }
    return fmt.Sprintf("%v", v)
}

// SuccessMessage renders success feedback
templ SuccessMessage(message string) {
    <div class="success">{ message }</div>
}

// ValidationErrors renders validation errors
templ ValidationErrors(errors []schema.ValidationError) {
    <div style="color: #dc3545; padding: 1rem; background: #f8d7da; border-radius: 4px;">
        <strong>Please fix the following errors:</strong>
        <ul>
            for _, err := range errors {
                <li>{ err.Message }</li>
            }
        </ul>
    </div>
}
```

### 3.5 Create Main Server

**File: `cmd/server/main.go`**
```go
package main

import (
    "log"
    "net/http"
    "github.com/niiniyare/erp/internal/handlers"
)

func main() {
    // Routes
    http.HandleFunc("/contact", handlers.HandleContactForm)
    http.HandleFunc("/api/contact", handlers.HandleContactSubmit)
    
    // Serve static files (if needed)
    // http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
    
    // Start server
    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### 3.6 Run the Example

**Step 1: Generate templ code**
```bash
cd views
templ generate
```

**Step 2: Run the server**
```bash
go run cmd/server/main.go
```

**Step 3: Open browser**
```
http://localhost:8080/contact
```

**You should see:**
- A contact form with 4 fields (name, email, subject, message)
- Instant HTML5 validation
- HTMX-powered submission (no page reload)
- Success message after submission

### 3.7 What Just Happened?

Let's trace the complete flow:

**1. Browser Request:**
```
GET /contact
```

**2. Handler Execution:**
```go
// Load JSON schema
jsonData, _ := os.ReadFile("schemas/contact-form.json")

// Parse to Go struct
parser := parse.NewParser()
schema, _ := parser.Parse(jsonData)

// Render with templ
views.FormPage(schema, nil).Render(ctx, w)
```

**3. templ Generation:**
```go
// FormPage component loops through fields
for _, field := range schema.Fields {
    // Renders appropriate component based on field.Type
    @FieldRenderer(&field, nil)
}
```

**4. HTML Output:**
```html
<form hx-post="/api/contact" hx-target="#result">
    <div class="field">
        <label for="name">Your Name <span class="required">*</span></label>
        <input 
            type="text" 
            name="name" 
            required 
            minlength="2" 
            maxlength="100"
            x-data="{ error: '' }"
            @invalid.prevent="error = $el.validationMessage"
        />
        <span class="error" x-show="error" x-text="error"></span>
    </div>
    <!-- More fields... -->
    <button type="submit">Send Message</button>
</form>
```

**5. User Interaction:**
- User fills form → HTML5 validates instantly
- User submits → HTMX intercepts
- HTMX sends POST → Handler validates server-side
- Handler returns HTML → HTMX swaps into #result

**No JavaScript code written. No React components. Just JSON schema.**

### 3.8 Next Steps

Now that you have a working example:

1. **Add More Fields**: Add phone, date, checkbox fields
2. **Add Validation**: Try `server` validation with business rules
3. **Add Permissions**: Use `requirePermission` on fields
4. **Add Layout**: Try grid, tabs, or steps layouts
5. **Add Conditional**: Use `showIf` for conditional fields
6. **Add Registry**: Move from files to database storage
7. **Add Enrichment**: Populate runtime flags based on user

Continue to the next sections for deep dives into each topic.

---

