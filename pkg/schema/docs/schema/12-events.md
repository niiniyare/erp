## Section 12: Events & Actions (HTMX Integration)

### 🎯 Learning Objectives

By the end of this section, you will understand:
- Action struct (buttons)
- Event handling
- HTMX integration
- Form submission patterns
- Partial updates
- Success/error handling

### 12.1 Action Struct

```go
type Action struct {
    ID       string     `json:"id" validate:"required"`
    Type     ActionType `json:"type" validate:"required"`
    Text     string     `json:"text" validate:"required"`
    Variant  string     `json:"variant,omitempty"`  // primary, secondary, danger
    Size     string     `json:"size,omitempty"`     // sm, md, lg
    Icon     string     `json:"icon,omitempty"`
    Disabled bool       `json:"disabled,omitempty"`
    Confirm  string     `json:"confirm,omitempty"`  // Confirmation message
    URL      string     `json:"url,omitempty"`      // For links
    HTMX     *ActionHTMX `json:"htmx,omitempty"`
}

type ActionType string

const (
    ActionSubmit ActionType = "submit"
    ActionReset  ActionType = "reset"
    ActionButton ActionType = "button"
    ActionLink   ActionType = "link"
)
```

**Example:**
```json
{
  "actions": [
    {
      "id": "submit",
      "type": "submit",
      "text": "Save",
      "variant": "primary",
      "size": "md",
      "icon": "save"
    },
    {
      "id": "cancel",
      "type": "button",
      "text": "Cancel",
      "variant": "secondary",
      "url": "/users"
    },
    {
      "id": "delete",
      "type": "button",
      "text": "Delete",
      "variant": "danger",
      "icon": "trash",
      "confirm": "Are you sure you want to delete this user?",
      "htmx": {
        "delete": "/api/users/{id}",
        "confirm": true
      }
    }
  ]
}
```

### 12.2 HTMX Configuration

**Schema-Level:**
```go
type HTMX struct {
    Enabled   bool              `json:"enabled"`
    Post      string            `json:"post,omitempty"`
    Get       string            `json:"get,omitempty"`
    Target    string            `json:"target,omitempty"`
    Swap      string            `json:"swap,omitempty"`
    Trigger   string            `json:"trigger,omitempty"`
    Headers   map[string]string `json:"headers,omitempty"`
    Indicator string            `json:"indicator,omitempty"`
}
```

**Example:**
```json
{
  "htmx": {
    "enabled": true,
    "post": "/api/users",
    "target": "#result",
    "swap": "innerHTML",
    "indicator": "#spinner"
  }
}
```

**Field-Level:**
```go
type FieldHTMX struct {
    Trigger   string            `json:"trigger,omitempty"`   // Event to trigger
    Post      string            `json:"post,omitempty"`      // POST endpoint
    Get       string            `json:"get,omitempty"`       // GET endpoint
    Target    string            `json:"target,omitempty"`    // Update target
    Swap      string            `json:"swap,omitempty"`      // Swap method
    Indicator string            `json:"indicator,omitempty"` // Loading indicator
    Headers   map[string]string `json:"headers,omitempty"`   // Extra headers
    Validate  bool              `json:"validate,omitempty"`  // Validate before request
}
```

**Example (dependent field):**
```json
{
  "name": "country",
  "type": "select",
  "label": "Country",
  "htmx": {
    "trigger": "change",
    "get": "/api/states?country={value}",
    "target": "#state-field",
    "swap": "outerHTML"
  }
}
```

### 12.3 Form Submission Patterns

#### Pattern 1: Basic Form Submission

**Schema:**
```json
{
  "config": {
    "action": "/api/users",
    "method": "POST"
  },
  "htmx": {
    "enabled": true,
    "post": "/api/users",
    "target": "#result",
    "swap": "innerHTML"
  }
}
```

**templ:**
```go
templ FormRenderer(schema *Schema, data map[string]any) {
    <form 
        hx-post={schema.HTMX.Post}
        hx-target={schema.HTMX.Target}
        hx-swap={schema.HTMX.Swap}
        hx-indicator={schema.HTMX.Indicator}
    >
        for _, field := range schema.Fields {
            @FieldRenderer(&field, data[field.Name])
        }
        
        for _, action := range schema.Actions {
            if action.Type == schema.ActionSubmit {
                <button type="submit" class="btn-primary">
                    {action.Text}
                </button>
            }
        }
    </form>
    
    <div id="result"></div>
    <div id="spinner" class="htmx-indicator">Loading...</div>
}
```

**HTML:**
```html
<form 
  hx-post="/api/users" 
  hx-target="#result" 
  hx-swap="innerHTML"
  hx-indicator="#spinner"
>
  <!-- Fields -->
  <button type="submit">Save</button>
</form>

<div id="result"></div>
<div id="spinner" class="htmx-indicator">Loading...</div>
```

**Handler:**
```go
func HandleSubmit(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Parse form
    r.ParseForm()
    data := formToMap(r.Form)
    
    // Validate
    schema, _ := registry.Get(ctx, "user-form")
    validator := validate.NewValidator(db)
    errors := validator.ValidateData(ctx, schema, data)
    
    if len(errors) > 0 {
        // Return errors (HTML fragment)
        views.ValidationErrors(errors).Render(ctx, w)
        return
    }
    
    // Save
    id, _ := repo.Save(ctx, data)
    
    // Return success (HTML fragment)
    views.SuccessMessage("User created successfully").Render(ctx, w)
}
```

#### Pattern 2: Field-Dependent Updates

**Schema:**
```json
{
  "fields": [
    {
      "name": "country",
      "type": "select",
      "label": "Country",
      "options": [
        {"value": "US", "label": "United States"},
        {"value": "CA", "label": "Canada"}
      ],
      "htmx": {
        "trigger": "change",
        "get": "/api/states?country={value}",
        "target": "#state-wrapper",
        "swap": "outerHTML"
      }
    },
    {
      "name": "state",
      "type": "select",
      "label": "State",
      "dependencies": ["country"]
    }
  ]
}
```

**templ:**
```go
templ SelectInput(field *schema.Field, value string) {
    <div class="field">
        <label for={field.Name}>{field.Label}</label>
        <select 
            name={field.Name}
            id={field.Name}
            if field.HTMX != nil {
                hx-trigger={field.HTMX.Trigger}
                hx-get={field.HTMX.Get}
                hx-target={field.HTMX.Target}
                hx-swap={field.HTMX.Swap}
            }
        >
            for _, opt := range field.Options {
                <option value={opt.Value}>{opt.Label}</option>
            }
        </select>
    </div>
}
```

**Handler:**
```go
func HandleStates(w http.ResponseWriter, r *http.Request) {
    country := r.URL.Query().Get("country")
    
    // Get states for country
    states := getStates(country)
    
    // Create field with updated options
    field := &schema.Field{
        Name:  "state",
        Type:  schema.FieldSelect,
        Label: "State",
        Options: states,
    }
    
    // Render field (wrapped in div with ID)
    views.SelectInputWrapper(field, "").Render(r.Context(), w)
}
```

#### Pattern 3: Inline Editing

**Schema:**
```json
{
  "name": "status",
  "type": "select",
  "label": "Status",
  "options": [
    {"value": "active", "label": "Active"},
    {"value": "inactive", "label": "Inactive"}
  ],
  "htmx": {
    "trigger": "change",
    "post": "/api/users/{id}/status",
    "target": "#status-feedback",
    "swap": "innerHTML"
  }
}
```

**HTML:**
```html
<div>
  <label>Status</label>
  <select 
    name="status"
    hx-post="/api/users/123/status"
    hx-trigger="change"
    hx-target="#status-feedback"
  >
    <option value="active" selected>Active</option>
    <option value="inactive">Inactive</option>
  </select>
  <span id="status-feedback"></span>
</div>
```

**Handler:**
```go
func HandleStatusUpdate(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    status := r.FormValue("status")
    
    // Update
    err := repo.UpdateStatus(r.Context(), id, status)
    if err != nil {
        w.Write([]byte(`<span class="error">Failed to update</span>`))
        return
    }
    
    w.Write([]byte(`<span class="success">✓ Updated</span>`))
}
```

### 12.4 Success/Error Handling

**templ components:**
```go
templ SuccessMessage(message string) {
    <div class="alert alert-success" role="alert">
        <svg class="icon"><!-- checkmark icon --></svg>
        <span>{message}</span>
    </div>
}

templ ValidationErrors(errors []schema.ValidationError) {
    <div class="alert alert-danger" role="alert">
        <strong>Please fix the following errors:</strong>
        <ul>
            for _, err := range errors {
                <li>{err.Field}: {err.Message}</li>
            }
        </ul>
    </div>
}

templ ErrorMessage(message string) {
    <div class="alert alert-danger" role="alert">
        <svg class="icon"><!-- error icon --></svg>
        <span>{message}</span>
    </div>
}
```

**Usage in handler:**
```go
func HandleSubmit(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // ... validation ...
    
    if len(errors) > 0 {
        views.ValidationErrors(errors).Render(ctx, w)
        return
    }
    
    // ... save ...
    
    if err != nil {
        views.ErrorMessage("Failed to save user").Render(ctx, w)
        return
    }
    
    views.SuccessMessage("User created successfully").Render(ctx, w)
}
```

### 12.5 Loading States

**HTMX provides automatic loading indicators:**

```html
<!-- Spinner shows during request -->
<div id="spinner" class="htmx-indicator">
  <svg class="animate-spin"><!-- spinner icon --></svg>
  Loading...
</div>

<form hx-post="/api/users" hx-indicator="#spinner">
  <!-- When form submits, spinner shows -->
</form>
```

**Or inline:**
```html
<button 
  hx-post="/api/users"
  hx-indicator="find .spinner"
>
  <span class="spinner htmx-indicator">⟳</span>
  Save
</button>
```

**During request:**
- `.htmx-request` class added to element
- `.htmx-indicator` elements become visible

**CSS:**
```css
.htmx-indicator {
  display: none;
}

.htmx-request .htmx-indicator {
  display: inline-block;
}

.htmx-request.htmx-indicator {
  display: inline-block;
}
```

### 12.6 Complete Example

**Schema:**
```json
{
  "id": "user-form",
  "type": "form",
  "title": "Create User",
  
  "config": {
    "action": "/api/users",
    "method": "POST"
  },
  
  "htmx": {
    "enabled": true,
    "post": "/api/users",
    "target": "#result",
    "swap": "innerHTML",
    "indicator": "#spinner"
  },
  
  "fields": [
    {
      "name": "username",
      "type": "text",
      "label": "Username",
      "required": true,
      "validation": {
        "minLength": 3,
        "maxLength": 20
      }
    },
    {
      "name": "email",
      "type": "email",
      "label": "Email",
      "required": true
    },
    {
      "name": "country",
      "type": "select",
      "label": "Country",
      "required": true,
      "htmx": {
        "trigger": "change",
        "get": "/api/states?country={value}",
        "target": "#state-field",
        "swap": "outerHTML"
      }
    },
    {
      "name": "state",
      "type": "select",
      "label": "State",
      "dependencies": ["country"]
    }
  ],
  
  "actions": [
    {
      "id": "submit",
      "type": "submit",
      "text": "Create User",
      "variant": "primary",
      "icon": "user-plus"
    },
    {
      "id": "cancel",
      "type": "button",
      "text": "Cancel",
      "variant": "secondary",
      "url": "/users"
    }
  ]
}
```

**This provides:**
- ✅ AJAX form submission (no page reload)
- ✅ Dependent field updates (country → state)
- ✅ Loading indicator
- ✅ Success/error messages in #result
- ✅ Client-side validation (HTML5)
- ✅ Server-side validation (Go)

---

