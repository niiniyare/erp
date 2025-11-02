## Section 18: Handler Patterns

### 🎯 Learning Objectives

By the end of this section, you will understand:
- HTTP handler patterns for schema-driven forms
- CRUD operations with schemas
- Form submission handling
- Partial updates with HTMX
- Error handling strategies

### 18.1 Handler Architecture

**Request Flow:**
```
HTTP Request
    ↓
Middleware (Auth, Tenant, etc.)
    ↓
Handler
    ├─ Load Schema
    ├─ Enrich with Context
    ├─ Validate Data (if POST)
    └─ Render Response
    ↓
HTTP Response (templ → HTML)
```

### 18.2 Basic CRUD Handlers

```go
// handlers/schema_form.go

package handlers

import (
    "net/http"
    "github.com/go-chi/chi/v5"
    "github.com/yourusername/awoerp/internal/schema"
    "github.com/yourusername/awoerp/views/schema"
)

type SchemaFormHandler struct {
    registry  schema.Registry
    enricher  schema.Enricher
    validator schema.Validator
    repo      Repository
}

// NewSchemaFormHandler creates a new handler
func NewSchemaFormHandler(
    registry schema.Registry,
    enricher schema.Enricher,
    validator schema.Validator,
    repo Repository,
) *SchemaFormHandler {
    return &SchemaFormHandler{
        registry:  registry,
        enricher:  enricher,
        validator: validator,
        repo:      repo,
    }
}

// HandleNew renders empty form for creating new record
func (h *SchemaFormHandler) HandleNew(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    schemaID := chi.URLParam(r, "schemaID")
    
    // Load schema
    s, err := h.registry.Get(ctx, schemaID)
    if err != nil {
        http.Error(w, "Schema not found", http.StatusNotFound)
        return
    }
    
    // Enrich with user context
    user := GetUserFromContext(ctx)
    enriched, err := h.enricher.Enrich(ctx, s, user)
    if err != nil {
        http.Error(w, "Enrichment failed", http.StatusInternalServerError)
        return
    }
    
    // Render form with empty data
    views.FormRenderer(enriched, nil).Render(ctx, w)
}

// HandleEdit renders form with existing data
func (h *SchemaFormHandler) HandleEdit(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    schemaID := chi.URLParam(r, "schemaID")
    id := chi.URLParam(r, "id")
    
    // Load schema
    s, err := h.registry.Get(ctx, schemaID)
    if err != nil {
        http.Error(w, "Schema not found", http.StatusNotFound)
        return
    }
    
    // Load existing data
    data, err := h.repo.Get(ctx, schemaID, id)
    if err != nil {
        http.Error(w, "Record not found", http.StatusNotFound)
        return
    }
    
    // Enrich with user context
    user := GetUserFromContext(ctx)
    enriched, err := h.enricher.Enrich(ctx, s, user)
    if err != nil {
        http.Error(w, "Enrichment failed", http.StatusInternalServerError)
        return
    }
    
    // Render form with data
    views.FormRenderer(enriched, data).Render(ctx, w)
}

// HandleCreate processes form submission for new record
func (h *SchemaFormHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    schemaID := chi.URLParam(r, "schemaID")
    
    // Parse form
    if err := r.ParseForm(); err != nil {
        http.Error(w, "Invalid form data", http.StatusBadRequest)
        return
    }
    data := formToMap(r.Form)
    
    // Load schema
    s, err := h.registry.Get(ctx, schemaID)
    if err != nil {
        http.Error(w, "Schema not found", http.StatusNotFound)
        return
    }
    
    // Validate
    errors := h.validator.ValidateData(ctx, s, data)
    if len(errors) > 0 {
        views.ValidationErrors(errors).Render(ctx, w)
        return
    }
    
    // Save
    id, err := h.repo.Create(ctx, schemaID, data)
    if err != nil {
        views.ErrorMessage("Failed to create record").Render(ctx, w)
        return
    }
    
    // Success response
    views.SuccessMessage(fmt.Sprintf("Record created successfully (ID: %s)", id)).Render(ctx, w)
}

// HandleUpdate processes form submission for existing record
func (h *SchemaFormHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    schemaID := chi.URLParam(r, "schemaID")
    id := chi.URLParam(r, "id")
    
    // Parse form
    if err := r.ParseForm(); err != nil {
        http.Error(w, "Invalid form data", http.StatusBadRequest)
        return
    }
    data := formToMap(r.Form)
    
    // Load schema
    s, err := h.registry.Get(ctx, schemaID)
    if err != nil {
        http.Error(w, "Schema not found", http.StatusNotFound)
        return
    }
    
    // Validate
    errors := h.validator.ValidateData(ctx, s, data)
    if len(errors) > 0 {
        views.ValidationErrors(errors).Render(ctx, w)
        return
    }
    
    // Update
    err = h.repo.Update(ctx, schemaID, id, data)
    if err != nil {
        views.ErrorMessage("Failed to update record").Render(ctx, w)
        return
    }
    
    // Success response
    views.SuccessMessage("Record updated successfully").Render(ctx, w)
}

// HandleDelete deletes a record
func (h *SchemaFormHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    schemaID := chi.URLParam(r, "schemaID")
    id := chi.URLParam(r, "id")
    
    // Delete
    err := h.repo.Delete(ctx, schemaID, id)
    if err != nil {
        views.ErrorMessage("Failed to delete record").Render(ctx, w)
        return
    }
    
    // Success response
    views.SuccessMessage("Record deleted successfully").Render(ctx, w)
}
```

### 18.3 Field Validation Handler (AJAX)

```go
// HandleFieldValidation validates single field
func (h *SchemaFormHandler) HandleFieldValidation(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    schemaID := chi.URLParam(r, "schemaID")
    fieldName := r.URL.Query().Get("field")
    
    // Load schema
    s, err := h.registry.Get(ctx, schemaID)
    if err != nil {
        http.Error(w, "Schema not found", http.StatusNotFound)
        return
    }
    
    // Get field
    field, ok := s.GetField(fieldName)
    if !ok {
        http.Error(w, "Field not found", http.StatusNotFound)
        return
    }
    
    // Get value
    value := r.FormValue(fieldName)
    
    // Get all form data (for cross-field validation)
    r.ParseForm()
    allData := formToMap(r.Form)
    
    // Validate field
    errors := h.validator.ValidateField(ctx, field, value, allData)
    
    if len(errors) > 0 {
        // Return error message
        w.WriteHeader(http.StatusUnprocessableEntity)
        w.Write([]byte(errors[0].Message))
        return
    }
    
    // Return success indicator
    w.Write([]byte("✓"))
}
```

### 18.4 Dependent Field Handler

```go
// HandleDependentField loads options for dependent field
func (h *SchemaFormHandler) HandleDependentField(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    schemaID := chi.URLParam(r, "schemaID")
    fieldName := chi.URLParam(r, "fieldName")
    
    // Load schema
    s, err := h.registry.Get(ctx, schemaID)
    if err != nil {
        http.Error(w, "Schema not found", http.StatusNotFound)
        return
    }
    
    // Get field
    field, ok := s.GetField(fieldName)
    if !ok {
        http.Error(w, "Field not found", http.StatusNotFound)
        return
    }
    
    // Get dependency values from query params
    dependencyValues := make(map[string]string)
    for _, dep := range field.Dependencies {
        dependencyValues[dep] = r.URL.Query().Get(dep)
    }
    
    // Load options (this would call your data source)
    options, err := h.loadDependentOptions(ctx, field, dependencyValues)
    if err != nil {
        http.Error(w, "Failed to load options", http.StatusInternalServerError)
        return
    }
    
    // Update field options
    field.Options = options
    
    // Render just the field (for HTMX to replace)
    views.SelectInput(field, "").Render(ctx, w)
}

func (h *SchemaFormHandler) loadDependentOptions(ctx context.Context, field *schema.Field, dependencies map[string]string) ([]schema.Option, error) {
    // This would typically query a database or API
    // Example: Load states based on country
    if field.Name == "state" && dependencies["country"] != "" {
        return h.repo.GetStates(ctx, dependencies["country"])
    }
    
    return nil, nil
}
```

### 18.5 Router Setup

```go
// routes/schema_routes.go

package routes

import (
    "github.com/go-chi/chi/v5"
    "github.com/yourusername/awoerp/handlers"
)

func SetupSchemaRoutes(r chi.Router, handler *handlers.SchemaFormHandler) {
    r.Route("/schema", func(r chi.Router) {
        // Form display
        r.Get("/{schemaID}/new", handler.HandleNew)
        r.Get("/{schemaID}/{id}/edit", handler.HandleEdit)
        
        // Form submission
        r.Post("/{schemaID}", handler.HandleCreate)
        r.Put("/{schemaID}/{id}", handler.HandleUpdate)
        r.Delete("/{schemaID}/{id}", handler.HandleDelete)
        
        // Field validation (AJAX)
        r.Post("/{schemaID}/validate", handler.HandleFieldValidation)
        
        // Dependent fields (HTMX)
        r.Get("/{schemaID}/field/{fieldName}", handler.HandleDependentField)
    })
}
```

### 18.6 Middleware Integration

```go
// middleware/schema_middleware.go

package middleware

import (
    "context"
    "net/http"
)

type contextKey string

const (
    userKey   contextKey = "user"
    tenantKey contextKey = "tenant"
)

// AuthMiddleware extracts user from session/JWT
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Get user from session/JWT
        user := getUserFromSession(r)
        if user == nil {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        
        // Add to context
        ctx := context.WithValue(r.Context(), userKey, user)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// TenantMiddleware extracts tenant
func TenantMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        user := GetUserFromContext(r.Context())
        
        // Add tenant to context
        ctx := context.WithValue(r.Context(), tenantKey, user.TenantID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// GetUserFromContext retrieves user from context
func GetUserFromContext(ctx context.Context) *User {
    user, _ := ctx.Value(userKey).(*User)
    return user
}

// GetTenantFromContext retrieves tenant from context
func GetTenantFromContext(ctx context.Context) string {
    tenant, _ := ctx.Value(tenantKey).(string)
    return tenant
}
```

---

