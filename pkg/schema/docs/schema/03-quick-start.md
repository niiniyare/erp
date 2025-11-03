# Quick Start (5 Minutes)

Build your first schema-driven form in 5 minutes.

## Prerequisites

- Go 1.21+
- templ CLI: `go install github.com/a-h/templ/cmd/templ@latest`

## Step 1: Install Dependencies

```bash
# Install packages
go get github.com/gofiber/fiber/v2
go get github.com/go-playground/validator/v10

# Install templ
go install github.com/a-h/templ/cmd/templ@latest
```

## Step 2: Create Schema

**File:** `schemas/contact-form.json`

```json
{
  "id": "contact-form",
  "type": "form",
  "title": "Contact Us",
  "fields": [
    {
      "name": "name",
      "type": "text",
      "label": "Your Name",
      "required": true,
      "validation": {
        "minLength": 2,
        "maxLength": 100
      }
    },
    {
      "name": "email",
      "type": "email",
      "label": "Email",
      "required": true
    },
    {
      "name": "message",
      "type": "textarea",
      "label": "Message",
      "required": true,
      "config": {
        "rows": 5
      }
    }
  ],
  "actions": [
    {
      "id": "submit",
      "type": "submit",
      "text": "Send Message"
    }
  ],
  "htmx": {
    "enabled": true,
    "post": "/api/contact",
    "target": "#result"
  }
}
```

## Step 3: Create Handler

**File:** `internal/handlers/contact.go`

```go
package handlers

import (
    "os"
    "github.com/gofiber/fiber/v2"
    "github.com/niiniyare/erp/pkg/schema/parse"
    "github.com/niiniyare/erp/pkg/schema/validate"
    "github.com/niiniyare/erp/views"
)

// ShowContactForm displays the form
func ShowContactForm(c *fiber.Ctx) error {
    // Load schema from file
    jsonData, err := os.ReadFile("schemas/contact-form.json")
    if err != nil {
        return c.Status(404).SendString("Schema not found")
    }
    
    // Parse schema
    parser := parse.NewParser()
    schema, err := parser.Parse(jsonData)
    if err != nil {
        return c.Status(500).SendString("Invalid schema")
    }
    
    // Render form
    return views.FormPage(schema, nil).Render(c.Context(), c.Response().BodyWriter())
}

// SubmitContact processes submission
func SubmitContact(c *fiber.Ctx) error {
    // Get form data
    data := map[string]any{
        "name":    c.FormValue("name"),
        "email":   c.FormValue("email"),
        "message": c.FormValue("message"),
    }
    
    // Load schema
    jsonData, _ := os.ReadFile("schemas/contact-form.json")
    parser := parse.NewParser()
    schema, _ := parser.Parse(jsonData)
    
    // Validate
    validator := validate.NewValidator(nil)
    errors := validator.ValidateData(c.Context(), schema, data)
    
    if len(errors) > 0 {
        // Return error HTML
        return views.ValidationErrors(errors).Render(c.Context(), c.Response().BodyWriter())
    }
    
    // TODO: Save to database, send email
    
    // Return success
    return views.SuccessMessage("Thank you!").Render(c.Context(), c.Response().BodyWriter())
}
```

## Step 4: Create templ View

**File:** `views/form.templ`

```go
package views

import "github.com/niiniyare/erp/pkg/schema"

// FormPage renders complete form page
templ FormPage(s *schema.Schema, data map[string]any) {
    <!DOCTYPE html>
    <html>
    <head>
        <title>{ s.Title }</title>
        <script src="https://unpkg.com/htmx.org@1.9.10"></script>
        <script src="https://unpkg.com/alpinejs@3.13.3" defer></script>
        <style>
            body { font-family: system-ui; max-width: 600px; margin: 2rem auto; }
            .field { margin-bottom: 1rem; }
            label { display: block; margin-bottom: 0.5rem; font-weight: 500; }
            input, textarea { width: 100%; padding: 0.5rem; border: 1px solid #ddd; }
            button { background: #0066cc; color: white; padding: 0.75rem 1.5rem; border: none; cursor: pointer; }
            .required { color: red; }
        </style>
    </head>
    <body>
        <h1>{ s.Title }</h1>
        @FormRenderer(s, data)
        <div id="result"></div>
    </body>
    </html>
}

// FormRenderer renders the form
templ FormRenderer(s *schema.Schema, data map[string]any) {
    <form
        if s.HTMX != nil && s.HTMX.Enabled {
            hx-post={ s.HTMX.Post }
            hx-target={ s.HTMX.Target }
        }
    >
        for _, field := range s.Fields {
            <div class="field">
                @FieldRenderer(&field, data[field.Name])
            </div>
        }
        
        for _, action := range s.Actions {
            if action.Type == "submit" {
                <button type="submit">{ action.Text }</button>
            }
        }
    </form>
}

// FieldRenderer renders individual fields
templ FieldRenderer(field *schema.Field, value any) {
    <label>
        { field.Label }
        if field.Required {
            <span class="required">*</span>
        }
    </label>
    
    switch field.Type {
        case "text", "email":
            <input 
                type={ string(field.Type) }
                name={ field.Name }
                if field.Required {
                    required
                }
            />
        case "textarea":
            <textarea
                name={ field.Name }
                if field.Required {
                    required
                }
                rows="5"
            ></textarea>
    }
}

// SuccessMessage renders success
templ SuccessMessage(msg string) {
    <div style="background: #d4edda; padding: 1rem; border-radius: 4px;">
        { msg }
    </div>
}

// ValidationErrors renders errors
templ ValidationErrors(errors []error) {
    <div style="background: #f8d7da; padding: 1rem; border-radius: 4px;">
        <strong>Errors:</strong>
        <ul>
            for _, err := range errors {
                <li>{ err.Error() }</li>
            }
        </ul>
    </div>
}
```

## Step 5: Create Server

**File:** `cmd/server/main.go`

```go
package main

import (
    "log"
    "github.com/gofiber/fiber/v2"
    "github.com/yourusername/yourapp/internal/handlers"
)

func main() {
    app := fiber.New()
    
    // Routes
    app.Get("/contact", handlers.ShowContactForm)
    app.Post("/api/contact", handlers.SubmitContact)
    
    // Start server
    log.Println("Server starting on :3000")
    log.Fatal(app.Listen(":3000"))
}
```

## Step 6: Generate templ and Run

```bash
# Generate templ code
cd views
templ generate

# Run server
cd ..
go run cmd/server/main.go
```

## Step 7: Test

Open browser: `http://localhost:3000/contact`

You should see:
- Contact form with 3 fields
- HTML5 validation (instant)
- HTMX submission (no page reload)
- Success/error messages

## What Just Happened?

1. **JSON Schema** defined the form structure
2. **Parser** converted JSON to Go structs
3. **templ** rendered Go structs to HTML
4. **Fiber** served the HTML
5. **HTMX** enhanced with AJAX
6. **Validator** checked data server-side

**No frontend code written!**

## Next Steps

### Add More Fields

```json
{
  "name": "phone",
  "type": "text",
  "label": "Phone"
}
```

### Add Validation

```json
{
  "validation": {
    "pattern": "^[0-9]{10}$",
    "messages": {
      "pattern": "Must be 10 digits"
    }
  }
}
```

### Add Storage

Replace `os.ReadFile` with registry:

```go
// Create registry
registry := registry.NewFilesystemRegistry("./schemas")

// In handler
schema, err := registry.Get(c.Context(), "contact-form")
```

### Add Caching

Use Redis registry:

```go
// Create with Redis cache
redisClient := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})
registry := registry.NewCachedRegistry(
    registry.NewPostgresRegistry(db),
    redisClient,
    1*time.Hour, // TTL
)
```

## Common Patterns

### Load from Database

```go
type PostgresRegistry struct {
    db *sql.DB
}

func (r *PostgresRegistry) Get(ctx context.Context, id string) (*schema.Schema, error) {
    var jsonData []byte
    err := r.db.QueryRowContext(ctx, 
        "SELECT json_schema FROM schemas WHERE id = $1", 
        id,
    ).Scan(&jsonData)
    
    parser := parse.NewParser()
    return parser.Parse(jsonData)
}
```

### Add Permissions

```go
// After loading schema
enricher := enrich.NewEnricher()
user := GetUserFromContext(c)
enriched, err := enricher.Enrich(c.Context(), schema, user)
```

### Multi-Tenant

```go
// Middleware adds tenant_id
func TenantMiddleware(c *fiber.Ctx) error {
    tenantID := extractTenantID(c)
    c.Locals("tenant_id", tenantID)
    return c.Next()
}

// Query filters by tenant
WHERE tenant_id = $1 AND id = $2
```

## Troubleshooting

**Schema not found:**
- Check file path: `schemas/contact-form.json`
- Verify file exists
- Check working directory

**Parse error:**
- Validate JSON syntax
- Check required fields (id, type, title)
- Verify field types are valid

**templ error:**
- Run `templ generate` in views directory
- Check generated `*_templ.go` files exist

**Validation not working:**
- Check HTML5 attributes render correctly
- Verify server-side validation runs
- Test with browser DevTools network tab

## Complete File Structure

```
yourapp/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   └── handlers/
│       └── contact.go
├── pkg/
│   └── schema/
│       ├── schema.go
│       ├── parse/
│       └── validate/
├── views/
│   ├── form.templ
│   └── *_templ.go (generated)
├── schemas/
│   └── contact-form.json
└── go.mod
```

## Next Documentation

- [04-schema-structure.md](04-schema-structure.md) - Complete schema reference
- [05-field-types.md](05-field-types.md) - All field types
- [06-validation.md](06-validation.md) - Validation rules
- [09-storage-interface.md](09-storage-interface.md) - Storage backends
- [10-registry.md](10-registry.md) - Registry implementation

---

[← Back to Architecture](02-architecture.md) | [Next: Schema Structure →](04-schema-structure.md)
