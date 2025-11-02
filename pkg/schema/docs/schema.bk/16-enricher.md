## Section 16: Enricher System

### 🎯 Learning Objectives

By the end of this section, you will understand:
- Schema enrichment concept
- Permission-based enrichment
- Tenant-based enrichment
- Field visibility and editability
- Context injection

### 16.1 Enricher Interface

```go
// Enricher enriches schemas with runtime context
type Enricher interface {
    // Enrich adds runtime context to schema
    Enrich(ctx context.Context, schema *Schema, user *User) (*Schema, error)
}

// User represents the authenticated user
type User struct {
    ID          string
    TenantID    string
    Roles       []string
    Permissions []string
}
```

### 16.2 Enricher Implementation

```go
// enricher implements the Enricher interface
type enricher struct {
    // Could inject services here
}

// NewEnricher creates a new enricher
func NewEnricher() Enricher {
    return &enricher{}
}

// Enrich adds runtime context to schema
func (e *enricher) Enrich(ctx context.Context, schema *Schema, user *User) (*Schema, error) {
    // Clone schema to avoid modifying original
    enriched := schema.Clone()
    
    // Set context
    enriched.Context = &Context{
        UserID:      user.ID,
        TenantID:    user.TenantID,
        Roles:       user.Roles,
        Permissions: user.Permissions,
        RequestID:   getRequestID(ctx),
        Timestamp:   time.Now(),
    }
    
    // Enrich fields
    for i := range enriched.Fields {
        field := &enriched.Fields[i]
        e.enrichField(field, user)
    }
    
    // Enrich layout (tabs, steps, sections)
    if enriched.Layout != nil {
        e.enrichLayout(enriched.Layout, user)
    }
    
    // Set tenant field value
    if enriched.Tenant != nil && enriched.Tenant.Enabled {
        for i := range enriched.Fields {
            field := &enriched.Fields[i]
            if field.Name == enriched.Tenant.Field {
                field.Value = user.TenantID
                field.Readonly = true
                field.Hidden = true
            }
        }
    }
    
    return enriched, nil
}

// enrichField adds runtime flags to field
func (e *enricher) enrichField(field *Field, user *User) {
    visible := true
    editable := true
    reason := ""
    
    // Check RequirePermission
    if field.RequirePermission != "" {
        if !user.HasPermission(field.RequirePermission) {
            visible = false
            editable = false
            reason = "insufficient_permissions"
        }
    }
    
    // Check Permissions (more granular)
    if field.Permissions != nil {
        if len(field.Permissions.View) > 0 {
            if !user.HasAnyRole(field.Permissions.View) {
                visible = false
                reason = "role_required"
            }
        }
        
        if len(field.Permissions.Edit) > 0 {
            if !user.HasAnyRole(field.Permissions.Edit) {
                editable = false
                reason = "role_required_for_edit"
            }
        }
    }
    
    // Respect readonly flag
    if field.Readonly {
        editable = false
    }
    
    // Respect hidden flag
    if field.Hidden {
        visible = false
    }
    
    // Set runtime
    field.Runtime = &FieldRuntime{
        Visible:  visible,
        Editable: editable,
        Reason:   reason,
    }
}

// enrichLayout enriches layout elements
func (e *enricher) enrichLayout(layout *Layout, user *User) {
    // Enrich tabs
    for i := range layout.Tabs {
        tab := &layout.Tabs[i]
        if tab.Conditional != nil {
            // Could evaluate tab visibility based on permissions
        }
    }
    
    // Enrich steps
    for i := range layout.Steps {
        step := &layout.Steps[i]
        if step.Conditional != nil {
            // Could evaluate step visibility based on permissions
        }
    }
    
    // Enrich sections
    for i := range layout.Sections {
        section := &layout.Sections[i]
        if section.Conditional != nil {
            // Could evaluate section visibility based on permissions
        }
    }
}
```

### 16.3 User Methods

```go
// HasPermission checks if user has permission
func (u *User) HasPermission(permission string) bool {
    for _, p := range u.Permissions {
        // Exact match
        if p == permission {
            return true
        }
        
        // Wildcard match
        if p == "*" {
            return true
        }
        
        // Prefix wildcard: "hr.*" matches "hr.view_salary"
        if strings.HasSuffix(p, ".*") {
            prefix := strings.TrimSuffix(p, ".*")
            if strings.HasPrefix(permission, prefix+".") {
                return true
            }
        }
    }
    return false
}

// HasRole checks if user has role
func (u *User) HasRole(role string) bool {
    for _, r := range u.Roles {
        if r == role || r == "*" {
            return true
        }
    }
    return false
}

// HasAnyRole checks if user has any of the roles
func (u *User) HasAnyRole(roles []string) bool {
    for _, role := range roles {
        if u.HasRole(role) {
            return true
        }
    }
    return false
}

// HasAllPermissions checks if user has all permissions
func (u *User) HasAllPermissions(permissions []string) bool {
    for _, permission := range permissions {
        if !u.HasPermission(permission) {
            return false
        }
    }
    return true
}
```

### 16.4 Enricher Usage Example

```go
func HandleForm(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Get user from context (set by auth middleware)
    user := GetUserFromContext(ctx)
    
    // Load schema
    schema, err := registry.Get(ctx, "employee-form")
    if err != nil {
        http.Error(w, "Schema not found", 404)
        return
    }
    
    // Enrich schema with user context
    enricher := NewEnricher()
    enriched, err := enricher.Enrich(ctx, schema, user)
    if err != nil {
        http.Error(w, "Enrichment failed", 500)
        return
    }
    
    // Now enriched schema has:
    // - Fields hidden if user lacks permissions
    // - Fields readonly if user can view but not edit
    // - Tenant ID pre-filled
    // - Runtime flags set
    
    // Render form
    views.FormRenderer(enriched, nil).Render(ctx, w)
}
```

---

