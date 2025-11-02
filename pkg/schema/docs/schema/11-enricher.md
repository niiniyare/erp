# Enricher

## Purpose

Adds runtime data to schemas before rendering:
- User permissions
- Tenant customization
- Default values
- Dynamic configuration

## Usage

```go
enricher := enrich.NewEnricher()

// Get user from context
user := GetUserFromContext(c)

// Enrich schema
enriched, err := enricher.Enrich(c.Context(), schema, user)
```

## Permission-Based Fields

```go
for i := range schema.Fields {
    field := &schema.Fields[i]
    
    if field.RequirePermission != "" {
        hasPermission := user.HasPermission(field.RequirePermission)
        
        field.Runtime = &FieldRuntime{
            Visible:  hasPermission,
            Editable: hasPermission,
            Reason:   "permission_required",
        }
    }
}
```

## Tenant Customization

```go
// Get tenant overrides
overrides := getTenantOverrides(ctx, schema.ID, user.TenantID)

// Apply overrides
for fieldName, override := range overrides {
    field := schema.FindField(fieldName)
    if override.Label != "" {
        field.Label = override.Label
    }
    if override.Required != nil {
        field.Required = *override.Required
    }
}
```

## Default Values

```go
if field.Value == nil && field.DefaultValue != nil {
    field.Value = field.DefaultValue
}

// Dynamic defaults
if field.Name == "created_by" {
    field.Value = user.ID
}
if field.Name == "tenant_id" {
    field.Value = user.TenantID
}
```

## Complete Example

```go
func HandleForm(c *fiber.Ctx) error {
    user := GetUserFromContext(c)
    
    // 1. Load schema
    schema, _ := registry.Get(c.Context(), "invoice-form")
    
    // 2. Enrich
    enricher := enrich.NewEnricher()
    enriched, _ := enricher.Enrich(c.Context(), schema, user)
    
    // 3. Render (permissions applied)
    return views.FormPage(enriched, nil).Render(
        c.Context(),
        c.Response().BodyWriter(),
    )
}
```

[← Back](10-registry.md) | [Next: Renderer →](12-renderer.md)