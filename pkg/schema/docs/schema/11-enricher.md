# Schema Enricher

## Overview

The **Enricher** transforms base schemas into runtime-ready schemas by injecting contextual data before rendering. It acts as the bridge between static schema definitions and dynamic, user-specific UI requirements.

### Purpose

- Apply **user permissions** to control field visibility and editability
- Inject **tenant-specific customizations** (labels, defaults, validations)
- Set **dynamic default values** based on context (user ID, timestamp, etc.)
- Apply **runtime state management** for UI behavior
- Handle **tenant isolation** for multi-tenant applications

### When to Use

```
Registry (base schema) → Enricher (add context) → Runtime (execution) → Renderer (UI)
```

The enricher should be called on **every request** before rendering to ensure up-to-date permissions and context.

---

## Interface Definition

```go
// Enricher adds runtime context to schemas
type Enricher interface {
    // Enrich a schema with user context
    Enrich(ctx context.Context, schema *schema.Schema, user User) (*schema.Schema, error)
    
    // EnrichField enriches a single field with runtime data
    EnrichField(ctx context.Context, field *schema.Field, user User, data map[string]any) error
    
    // SetTenantProvider sets the tenant customization provider
    SetTenantProvider(provider TenantProvider)
}

// User represents a user in the system for enrichment purposes
type User interface {
    GetID() string
    GetTenantID() string
    GetPermissions() []string
    GetRoles() []string
    HasPermission(permission string) bool
    HasRole(role string) bool
}

// TenantProvider interface for retrieving tenant customizations
type TenantProvider interface {
    GetCustomization(ctx context.Context, schemaID, tenantID string) (*TenantCustomization, error)
}

// TenantCustomization holds all tenant-specific overrides for a schema
type TenantCustomization struct {
    SchemaID   string                    `json:"schema_id"`
    TenantID   string                    `json:"tenant_id"`
    Overrides  map[string]TenantOverride `json:"overrides"`
    CreatedAt  time.Time                 `json:"created_at"`
    UpdatedAt  time.Time                 `json:"updated_at"`
}

// TenantOverride represents tenant-specific customizations
type TenantOverride struct {
    FieldName    string `json:"field_name"`
    Label        string `json:"label,omitempty"`
    Required     *bool  `json:"required,omitempty"`
    Hidden       *bool  `json:"hidden,omitempty"`
    DefaultValue any    `json:"default_value,omitempty"`
    Placeholder  string `json:"placeholder,omitempty"`
    Help         string `json:"help,omitempty"`
}

// FieldRuntime holds runtime state populated by enricher
type FieldRuntime struct {
    Visible  bool   `json:"visible"`  // Whether field should be visible to user
    Editable bool   `json:"editable"` // Whether field can be edited by user  
    Reason   string `json:"reason"`   // Reason for visibility/editability state
}
```

---

## Basic Usage

### Simple Enrichment

```go
func HandleForm(c *fiber.Ctx) error {
    // 1. Get user from authentication middleware
    user := GetUserFromContext(c)
    if user == nil {
        return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
    }
    
    // 2. Load base schema from registry
    schema, err := registry.Get(c.Context(), "invoice-form")
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Schema not found"})
    }
    
    // 3. Clone schema to avoid mutating cached version
    schemaCopy := schema.Clone()
    
    // 4. Enrich with user context
    enricher := enrich.NewEnricher()
    enriched, err := enricher.Enrich(c.Context(), schemaCopy, user)
    if err != nil {
        log.Error().Err(err).Msg("Failed to enrich schema")
        return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
    }
    
    // 5. Render the enriched schema
    return views.FormPage(enriched, nil).Render(
        c.Context(),
        c.Response().BodyWriter(),
    )
}
```

### Advanced Enrichment with Options

```go
func HandleAdminForm(c *fiber.Ctx) error {
    user := GetUserFromContext(c)
    schema, err := registry.Get(c.Context(), "user-management-form")
    if err != nil {
        return err
    }
    
    // Enrich with custom options
    opts := &enrich.EnrichOptions{
        User:             user,
        TenantID:         c.Query("tenant_id"),
        SkipPermissions:  user.IsAdmin(), // Admins bypass permission checks
        DefaultOverrides: map[string]interface{}{
            "status": "active",
            "role":   "user",
        },
        AuditEnrichment: true, // Track enrichment for logging
    }
    
    enricher := enrich.NewEnricher()
    result, err := enricher.EnrichWithOptions(c.Context(), schema.Clone(), opts)
    if err != nil {
        return err
    }
    
    // Log enrichment details for audit
    log.Info().
        Str("schema_id", schema.ID).
        Str("user_id", user.ID).
        Strs("fields_hidden", result.FieldsHidden).
        Dur("enrichment_time", result.EnrichmentTime).
        Msg("Schema enriched")
    
    return c.JSON(result.Schema)
}
```

---

## Permission-Based Field Control

### Basic Permission Checks

```go
// applyPermissions applies permission-based field restrictions
func (e *DefaultEnricher) applyPermissions(field *schema.Field, user User) {
    // Initialize runtime if not exists
    if field.Runtime == nil {
        field.Runtime = &schema.FieldRuntime{}
    }
    
    // Check if field requires specific permission
    if field.RequirePermission != "" {
        hasPermission := user.HasPermission(field.RequirePermission)
        
        field.Runtime.Visible = hasPermission
        field.Runtime.Editable = hasPermission
        
        if !hasPermission {
            field.Runtime.Reason = "permission_required"
        }
    } else {
        // Default to visible and editable if no permission required
        field.Runtime.Visible = true
        field.Runtime.Editable = !field.Readonly
    }
    
    // Check role-based restrictions
    if len(field.RequireRoles) > 0 {
        hasRequiredRole := false
        for _, requiredRole := range field.RequireRoles {
            if user.HasRole(requiredRole) {
                hasRequiredRole = true
                break
            }
        }
        
        if !hasRequiredRole {
            field.Runtime.Visible = false
            field.Runtime.Editable = false
            field.Runtime.Reason = "role_required"
        }
    }
}
```

### Field Visibility and Editability

```go
// Field state is controlled through Runtime struct populated by enricher
type FieldRuntime struct {
    Visible  bool   `json:"visible"`  // Whether field should be visible to user
    Editable bool   `json:"editable"` // Whether field can be edited by user
    Reason   string `json:"reason"`   // Reason for visibility/editability state
}

// Example: checking field state after enrichment
func CheckFieldAccess(enrichedField *schema.Field) {
    if !enrichedField.Runtime.Visible {
        // Field should not be rendered
        return
    }
    
    if !enrichedField.Runtime.Editable {
        // Render as readonly
        enrichedField.Readonly = true
    }
    
    // Log reason if field is restricted
    if enrichedField.Runtime.Reason != "" {
        log.Info().Str("field", enrichedField.Name).
            Str("reason", enrichedField.Runtime.Reason).
            Msg("Field access restricted")
    }
}
```

---

## Tenant Customization

### Loading Tenant Overrides

```go
// TenantOverride represents tenant-specific customizations
type TenantOverride struct {
    FieldName    string
    Label        *string
    Placeholder  *string
    Required     *bool
    DefaultValue interface{}
    Validation   *ValidationRule
    HelpText     *string
}

// Get tenant overrides with caching
var tenantCache = ttlcache.New[string, []TenantOverride](
    ttlcache.WithTTL[string, []TenantOverride](15 * time.Minute),
)

func (e *Enricher) getTenantOverrides(
    ctx context.Context,
    schemaID string,
    tenantID string,
) ([]TenantOverride, error) {
    cacheKey := fmt.Sprintf("%s:%s", schemaID, tenantID)
    
    // Check cache
    if cached := tenantCache.Get(cacheKey); cached != nil {
        return cached.Value(), nil
    }
    
    // Fetch from database
    var overrides []TenantOverride
    err := e.db.WithContext(ctx).
        Where("schema_id = ? AND tenant_id = ?", schemaID, tenantID).
        Find(&overrides).Error
    if err != nil {
        return nil, fmt.Errorf("failed to fetch tenant overrides: %w", err)
    }
    
    // Cache for future requests
    tenantCache.Set(cacheKey, overrides, ttlcache.DefaultTTL)
    
    return overrides, nil
}
```

### Applying Tenant Overrides

```go
func (e *Enricher) applyTenantOverrides(
    ctx context.Context,
    schema *Schema,
    tenantID string,
) error {
    overrides, err := e.getTenantOverrides(ctx, schema.ID, tenantID)
    if err != nil {
        return err
    }
    
    for _, override := range overrides {
        field := schema.FindField(override.FieldName)
        if field == nil {
            log.Warn().
                Str("field", override.FieldName).
                Msg("Tenant override for non-existent field")
            continue
        }
        
        // Apply overrides
        if override.Label != nil {
            field.Label = *override.Label
        }
        if override.Placeholder != nil {
            field.Placeholder = *override.Placeholder
        }
        if override.Required != nil {
            field.Required = *override.Required
        }
        if override.DefaultValue != nil {
            field.Default = override.DefaultValue
        }
        if override.HelpText != nil {
            field.HelpText = *override.HelpText
        }
        if override.Validation != nil {
            field.Validations = append(field.Validations, override.Validation)
        }
    }
    
    return nil
}
```

---

## Dynamic Default Values

### Static and Dynamic Defaults

```go
// applyDynamicDefaults applies dynamic default values based on context
func (e *DefaultEnricher) applyDynamicDefaults(field *schema.Field, data map[string]any) {
    // Only apply defaults if field has no current value
    if field.Value != nil {
        return
    }
    
    // Apply static default value if exists
    if field.Default != nil {
        field.Value = field.Default
        return
    }
    
    // Apply dynamic defaults based on field name
    switch field.Name {
    case "created_by", "user_id", "author_id":
        if userID, ok := data["user_id"].(string); ok {
            field.Value = userID
        }
    case "tenant_id", "organization_id":
        if tenantID, ok := data["tenant_id"].(string); ok {
            field.Value = tenantID
        }
    case "created_at", "timestamp":
        if timestamp, ok := data["timestamp"].(time.Time); ok {
            field.Value = timestamp.Format(time.RFC3339)
        }
    case "updated_at":
        field.Value = time.Now().Format(time.RFC3339)
    }
}

// applyTenantIsolation applies tenant isolation for multi-tenant fields
func (e *DefaultEnricher) applyTenantIsolation(field *schema.Field, user User) {
    // Automatically set tenant_id field for tenant isolation
    if field.Name == "tenant_id" {
        field.Value = user.GetTenantID()
        field.Hidden = true // Hide tenant_id from user
        field.Readonly = true
    }
}
```

### Computed Defaults

```go
// ComputedDefault calculates a default value
type ComputedDefault func(ctx context.Context, schema *Schema, user *User) (interface{}, error)

// Register computed defaults
var computedDefaults = map[string]ComputedDefault{
    "invoice_number": func(ctx context.Context, schema *Schema, user *User) (interface{}, error) {
        // Generate next invoice number
        return generateInvoiceNumber(ctx, user.TenantID)
    },
    
    "order_total": func(ctx context.Context, schema *Schema, user *User) (interface{}, error) {
        // Calculate from line items
        items := schema.FindField("line_items")
        if items == nil || items.Value == nil {
            return 0.0, nil
        }
        return calculateTotal(items.Value), nil
    },
    
    "full_name": func(ctx context.Context, schema *Schema, user *User) (interface{}, error) {
        // Combine first and last name
        firstName := schema.FindField("first_name")
        lastName := schema.FindField("last_name")
        if firstName != nil && lastName != nil {
            return fmt.Sprintf("%s %s", firstName.Value, lastName.Value), nil
        }
        return "", nil
    },
}

func (e *Enricher) applyComputedDefaults(
    ctx context.Context,
    schema *Schema,
    user *User,
) error {
    for fieldName, computeFn := range computedDefaults {
        field := schema.FindField(fieldName)
        if field == nil || field.Value != nil {
            continue
        }
        
        value, err := computeFn(ctx, schema, user)
        if err != nil {
            log.Error().Err(err).Str("field", fieldName).Msg("Computed default failed")
            continue // Don't fail enrichment for computed defaults
        }
        
        field.Value = value
    }
    
    return nil
}
```

---

## Conditional Logic

### Conditional Field Visibility

```go
// ConditionalRule defines when a field should be visible/editable
type ConditionalRule struct {
    Field     string // Target field name
    Condition string // Condition expression
    Action    string // "show", "hide", "enable", "disable"
}

func (e *Enricher) applyConditionalRules(schema *Schema) error {
    rules := []ConditionalRule{
        {
            Field:     "shipping_address",
            Condition: "delivery_method == 'ship'",
            Action:    "show",
        },
        {
            Field:     "pickup_location",
            Condition: "delivery_method == 'pickup'",
            Action:    "show",
        },
        {
            Field:     "approval_notes",
            Condition: "status == 'pending_approval'",
            Action:    "show",
        },
    }
    
    for _, rule := range rules {
        field := schema.FindField(rule.Field)
        if field == nil {
            continue
        }
        
        // Evaluate condition
        result, err := e.evaluateCondition(rule.Condition, schema)
        if err != nil {
            log.Error().Err(err).Str("condition", rule.Condition).Msg("Failed to evaluate condition")
            continue
        }
        
        // Apply action
        if field.Runtime == nil {
            field.Runtime = &FieldRuntime{}
        }
        
        switch rule.Action {
        case "show":
            field.Runtime.Visible = result
        case "hide":
            field.Runtime.Visible = !result
        case "enable":
            field.Runtime.Editable = result
        case "disable":
            field.Runtime.Editable = !result
        }
        
        field.Runtime.Condition = rule.Condition
    }
    
    return nil
}

// Simple expression evaluator (use a proper library in production)
func (e *Enricher) evaluateCondition(condition string, schema *Schema) (bool, error) {
    // Parse "field == value" expressions
    parts := strings.Split(condition, "==")
    if len(parts) != 2 {
        return false, fmt.Errorf("invalid condition format")
    }
    
    fieldName := strings.TrimSpace(parts[0])
    expectedValue := strings.Trim(strings.TrimSpace(parts[1]), "'\"")
    
    field := schema.FindField(fieldName)
    if field == nil {
        return false, nil
    }
    
    return fmt.Sprintf("%v", field.Value) == expectedValue, nil
}
```

---

## Complete Enricher Implementation

```go
package enrich

import (
    "context"
    "fmt"
    "time"
    
    "github.com/jellydator/ttlcache/v3"
    "gorm.io/gorm"
)

type Enricher struct {
    db                *gorm.DB
    tenantCache       *ttlcache.Cache[string, []TenantOverride]
    permissionChecker PermissionChecker
}

func NewEnricher(db *gorm.DB) *Enricher {
    return &Enricher{
        db: db,
        tenantCache: ttlcache.New[string, []TenantOverride](
            ttlcache.WithTTL[string, []TenantOverride](15 * time.Minute),
        ),
        permissionChecker: NewPermissionChecker(),
    }
}

func (e *Enricher) Enrich(
    ctx context.Context,
    schema *Schema,
    user *User,
) (*Schema, error) {
    return e.EnrichWithOptions(ctx, schema, &EnrichOptions{
        User: user,
    })
}

func (e *Enricher) EnrichWithOptions(
    ctx context.Context,
    schema *Schema,
    opts *EnrichOptions,
) (*EnrichmentResult, error) {
    startTime := time.Now()
    result := &EnrichmentResult{
        Schema:    schema,
        Timestamp: startTime,
    }
    
    // Validate inputs
    if schema == nil {
        return nil, fmt.Errorf("schema cannot be nil")
    }
    if opts.User == nil && !opts.SkipPermissions {
        return nil, fmt.Errorf("user required for permission checks")
    }
    
    // Apply enrichments in order
    steps := []struct {
        name     string
        skip     bool
        function func() error
    }{
        {
            name: "permissions",
            skip: opts.SkipPermissions,
            function: func() error {
                return e.applyPermissions(schema, opts.User)
            },
        },
        {
            name: "tenant_overrides",
            skip: opts.SkipTenant,
            function: func() error {
                tenantID := opts.TenantID
                if tenantID == "" && opts.User != nil {
                    tenantID = opts.User.TenantID
                }
                return e.applyTenantOverrides(ctx, schema, tenantID)
            },
        },
        {
            name: "defaults",
            skip: opts.SkipDefaults,
            function: func() error {
                if err := e.applyDefaultValues(schema, opts.User); err != nil {
                    return err
                }
                return e.applyComputedDefaults(ctx, schema, opts.User)
            },
        },
        {
            name: "conditional_rules",
            skip: false,
            function: func() error {
                return e.applyConditionalRules(schema)
            },
        },
    }
    
    // Execute enrichment steps
    for _, step := range steps {
        if step.skip {
            continue
        }
        
        if err := step.function(); err != nil {
            return nil, fmt.Errorf("enrichment step '%s' failed: %w", step.name, err)
        }
    }
    
    // Apply default overrides
    if opts.DefaultOverrides != nil {
        for fieldName, value := range opts.DefaultOverrides {
            if field := schema.FindField(fieldName); field != nil {
                field.Value = value
                result.DefaultsApplied[fieldName] = value
            }
        }
    }
    
    // Track enrichment details if requested
    if opts.AuditEnrichment {
        e.trackEnrichment(schema, result)
    }
    
    result.EnrichmentTime = time.Since(startTime)
    return result, nil
}

func (e *Enricher) trackEnrichment(schema *Schema, result *EnrichmentResult) {
    for _, field := range schema.Fields {
        if field.Runtime != nil {
            if !field.Runtime.Visible {
                result.FieldsHidden = append(result.FieldsHidden, field.Name)
            }
            if field.Runtime.Modified {
                result.FieldsModified = append(result.FieldsModified, field.Name)
            }
        }
    }
}
```

---

## Middleware Integration

### Fiber Middleware

```go
// EnrichmentMiddleware enriches schemas automatically
func EnrichmentMiddleware(enricher *enrich.Enricher, registry SchemaRegistry) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Get schema ID from route or query
        schemaID := c.Params("schemaID")
        if schemaID == "" {
            schemaID = c.Query("schema")
        }
        
        if schemaID == "" {
            return c.Next() // No schema specified, skip enrichment
        }
        
        // Get user from context
        user := GetUserFromContext(c)
        if user == nil {
            return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
        }
        
        // Load base schema
        schema, err := registry.Get(c.Context(), schemaID)
        if err != nil {
            return c.Status(404).JSON(fiber.Map{
                "error": "Schema not found",
                "schema_id": schemaID,
            })
        }
        
        // Clone to avoid mutating cached schema
        schemaCopy := schema.Clone()
        
        // Enrich
        result, err := enricher.Enrich(c.Context(), schemaCopy, user)
        if err != nil {
            log.Error().Err(err).Str("schema_id", schemaID).Msg("Enrichment failed")
            return c.Status(500).JSON(fiber.Map{"error": "Failed to enrich schema"})
        }
        
        // Store in context for downstream handlers
        c.Locals("enriched_schema", result.Schema)
        c.Locals("enrichment_result", result)
        
        return c.Next()
    }
}

// Usage
app.Use("/api/forms/:schemaID", EnrichmentMiddleware(enricher, registry))

app.Get("/api/forms/:schemaID", func(c *fiber.Ctx) error {
    schema := c.Locals("enriched_schema").(*Schema)
    return c.JSON(schema)
})
```

---

## Security Best Practices

### ⚠️ Critical Security Rules

```go
// ❌ NEVER accept enriched schemas from clients
func HandleSubmitBAD(c *fiber.Ctx) error {
    var schema Schema
    if err := c.BodyParser(&schema); err != nil {
        return err
    }
    
    // DANGER: Attacker could set all fields to visible/editable
    // and bypass permission checks!
    return processForm(schema)
}

// ✅ ALWAYS enrich server-side on every request
func HandleSubmitGOOD(c *fiber.Ctx) error {
    user := GetUserFromContext(c)
    
    // Get submitted data
    var formData map[string]interface{}
    if err := c.BodyParser(&formData); err != nil {
        return err
    }
    
    // Load and enrich schema server-side
    schema, _ := registry.Get(c.Context(), "invoice-form")
    enriched, _ := enricher.Enrich(c.Context(), schema.Clone(), user)
    
    // Validate submitted data against enriched schema
    if err := validator.Validate(enriched, formData); err != nil {
        return c.Status(400).JSON(err)
    }
    
    // Process only permitted fields
    return processForm(enriched, formData)
}
```

### Sanitize Before Sending to Client

```go
// Remove sensitive enrichment metadata before sending to client
func (e *Enricher) SanitizeForClient(schema *Schema) *Schema {
    sanitized := schema.Clone()
    
    for i := range sanitized.Fields {
        field := &sanitized.Fields[i]
        
        // Remove permission requirements (security through obscurity isn't security,
        // but no need to advertise what permissions exist)
        field.RequirePermission = ""
        field.Permissions = nil
        
        // Remove internal metadata
        if field.Metadata != nil {
            delete(field.Metadata, "internal")
            delete(field.Metadata, "system")
        }
        
        // Keep only public runtime info
        if field.Runtime != nil {
            field.Runtime.Reason = "" // Don't expose permission denial reasons
        }
    }
    
    return sanitized
}
```

### Validate Permissions on Every Operation

```go
// Never trust client state - always re-validate
func HandleFieldUpdate(c *fiber.Ctx) error {
    fieldName := c.Params("field")
    var newValue interface{}
    c.BodyParser(&newValue)
    
    user := GetUserFromContext(c)
    schema, _ := registry.Get(c.Context(), "invoice-form")
    
    // Re-enrich to get current permissions
    enriched, _ := enricher.Enrich(c.Context(), schema.Clone(), user)
    
    // Find the field
    field := enriched.FindField(fieldName)
    if field == nil {
        return c.Status(404).JSON(fiber.Map{"error": "Field not found"})
    }
    
    // Check if user can edit this field
    if field.Runtime == nil || !field.Runtime.Editable {
        return c.Status(403).JSON(fiber.Map{
            "error": "You don't have permission to edit this field",
        })
    }
    
    // Proceed with update
    return updateField(c.Context(), fieldName, newValue)
}
```

---

## Performance Optimization

### Cache Strategies

```go
// Multi-level caching
type CachedEnricher struct {
    enricher        *Enricher
    schemaCache     *ttlcache.Cache[string, *Schema]          // Base schemas
    tenantCache     *ttlcache.Cache[string, []TenantOverride] // Tenant overrides
    permissionCache *ttlcache.Cache[string, bool]             // Permission results
}

func NewCachedEnricher(enricher *Enricher) *CachedEnricher {
    return &CachedEnricher{
        enricher: enricher,
        schemaCache: ttlcache.New[string, *Schema](
            ttlcache.WithTTL[string, *Schema](10 * time.Minute),
        ),
        tenantCache: ttlcache.New[string, []TenantOverride](
            ttlcache.WithTTL[string, []TenantOverride](15 * time.Minute),
        ),
        permissionCache: ttlcache.New[string, bool](
            ttlcache.WithTTL[string, bool](5 * time.Minute),
        ),
    }
}
```

### Lazy Enrichment

```go
// Only enrich visible sections
func (e *Enricher) EnrichSection(
    ctx context.Context,
    schema *Schema,
    section string,
    user *User,
) (*Schema, error) {
    // Create a new schema with only the requested section
    sectionSchema := &Schema{
        ID:      schema.ID,
        Version: schema.Version,
        Fields:  make([]Field, 0),
    }
    
    // Copy only fields in this section
    for _, field := range schema.Fields {
        if field.Section == section {
            sectionSchema.Fields = append(sectionSchema.Fields, field)
        }
    }
    
    // Enrich only this section
    return e.Enrich(ctx, sectionSchema, user)
}
```

### Bulk Enrichment

```go
// Enrich multiple schemas in one pass
func (e *Enricher) EnrichBulk(
    ctx context.Context,
    schemas []*Schema,
    user *User,
) ([]*Schema, error) {
    // Pre-load all tenant overrides in one query
    schemaIDs := make([]string, len(schemas))
    for i, schema := range schemas {
        schemaIDs[i] = schema.ID
    }
    
    overrides, err := e.loadTenantOverridesBulk(ctx, schemaIDs, user.TenantID)
    if err != nil {
        return nil, err
    }
    
    // Enrich each schema
    enriched := make([]*Schema, len(schemas))
    for i, schema := range schemas {
        // Apply cached overrides
        e.applyOverridesFromCache(schema, overrides[schema.ID])
        
        // Apply other enrichments
        result, err := e.Enrich(ctx, schema, user)
        if err != nil {
            return nil, err
        }
        enriched[i] = result.Schema
    }
    
    return enriched, nil
}
```

---

## Testing

### Unit Tests

```go
package enrich_test

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestEnricher_Permissions(t *testing.T) {
    tests := []struct {
        name           string
        user           *User
        fieldName      string
        permission     string
        expectVisible  bool
        expectEditable bool
    }{
        {
            name: "admin can view and edit salary",
            user: &User{
                Role:        "admin",
                Permissions: []string{"view:salary", "edit:salary"},
            },
            fieldName:      "salary",
            permission:     "view:salary",
            expectVisible:  true,
            expectEditable: true,
        },
        {
            name: "manager can view but not edit salary",
            user: &User{
                Role:        "manager",
                Permissions: []string{"view:salary"},
            },
            fieldName:      "salary",
            permission:     "view:salary",
            expectVisible:  true,
            expectEditable: false,
        },
        {
            name: "employee cannot view salary",
            user: &User{
                Role:        "employee",
                Permissions: []string{},
            },
            fieldName:      "salary",
            permission:     "view:salary",
            expectVisible:  false,
            expectEditable: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            schema := &Schema{
                Fields: []Field{
                    {
                        Name:              tt.fieldName,
                        RequirePermission: tt.permission,
                    },
                },
            }
            
            enricher := NewEnricher(nil)
            
            // Execute
            result, err := enricher.Enrich(context.Background(), schema, tt.user)
            
            // Assert
            require.NoError(t, err)
            field := result.Schema.FindField(tt.fieldName)
            require.NotNil(t, field)
            require.NotNil(t, field.Runtime)
            
            assert.Equal(t, tt.expectVisible, field.Runtime.Visible)
            assert.Equal(t, tt.expectEditable, field.Runtime.Editable)
        })
    }
}

func TestEnricher_DefaultValues(t *testing.T) {
    user := &User{
        ID:       "user-123",
        TenantID: "tenant-456",
    }
    
    schema := &Schema{
        Fields: []Field{
            {Name: "created_by"},
            {Name: "tenant_id"},
            {Name: "status", DefaultValue: "draft"},
        },
    }
    
    enricher := NewEnricher(nil)
    result, err := enricher.Enrich(context.Background(), schema, user)
    
    require.NoError(t, err)
    
    createdBy := result.Schema.FindField("created_by")
    assert.Equal(t, user.ID, createdBy.Value)
    
    tenantID := result.Schema.FindField("tenant_id")
    assert.Equal(t, user.TenantID, tenantID.Value)
    
    status := result.Schema.FindField("status")
    assert.Equal(t, "draft", status.Value)
}

func TestEnricher_TenantOverrides(t *testing.T) {
    // Mock tenant overrides
    mockDB := setupMockDB(t)
    mockDB.Create(&TenantOverride{
        SchemaID:  "invoice-form",
        TenantID:  "tenant-123",
        FieldName: "tax_rate",
        Label:     stringPtr("VAT Rate"),
        Required:  boolPtr(true),
    })
    
    schema := &Schema{
        ID: "invoice-form",
        Fields: []Field{
            {
                Name:     "tax_rate",
                Label:    "Tax Rate",
                Required: false,
            },
        },
    }
    
    enricher := NewEnricher(mockDB)
    result, err := enricher.EnrichWithOptions(context.Background(), schema, &EnrichOptions{
        User: &User{TenantID: "tenant-123"},
    })
    
    require.NoError(t, err)
    
    field := result.Schema.FindField("tax_rate")
    assert.Equal(t, "VAT Rate", field.Label)
    assert.True(t, field.Required)
}
```

### Integration Tests

```go
func TestEnricher_Integration(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    registry := NewRegistry(db)
    enricher := NewEnricher(db)
    
    // Create test schema
    schema := &Schema{
        ID:    "test-form",
        Title: "Test Form",
        Fields: []Field{
            {Name: "public_field"},
            {Name: "private_field", RequirePermission: "admin"},
        },
    }
    registry.Register(context.Background(), schema)
    
    // Test as regular user
    t.Run("regular_user", func(t *testing.T) {
        user := &User{Role: "user", Permissions: []string{}}
        
        loaded, _ := registry.Get(context.Background(), "test-form")
        enriched, _ := enricher.Enrich(context.Background(), loaded.Clone(), user)
        
        assert.True(t, enriched.FindField("public_field").Runtime.Visible)
        assert.False(t, enriched.FindField("private_field").Runtime.Visible)
    })
    
    // Test as admin
    t.Run("admin_user", func(t *testing.T) {
        admin := &User{Role: "admin", Permissions: []string{"admin"}}
        
        loaded, _ := registry.Get(context.Background(), "test-form")
        enriched, _ := enricher.Enrich(context.Background(), loaded.Clone(), admin)
        
        assert.True(t, enriched.FindField("public_field").Runtime.Visible)
        assert.True(t, enriched.FindField("private_field").Runtime.Visible)
    })
}
```

---

## Troubleshooting

### Common Issues

#### 1. **Enriched Schema Not Reflecting Permission Changes**

**Problem:** User permissions changed but schema still shows old permissions.

**Solution:**
```go
// Clear permission cache after role/permission changes
func (s *UserService) UpdateUserRole(userID, newRole string) error {
    if err := s.db.UpdateRole(userID, newRole); err != nil {
        return err
    }
    
    // Clear cached permissions
    s.permissionCache.Delete(userID)
    
    return nil
}
```

#### 2. **Tenant Overrides Not Appearing**

**Problem:** Tenant customizations aren't showing up in enriched schema.

**Solution:**
```go
// Ensure tenant ID is correctly passed
opts := &EnrichOptions{
    User:     user,
    TenantID: user.TenantID, // Explicitly set tenant ID
}

// Invalidate tenant cache if overrides were just updated
enricher.InvalidateTenantCache(schemaID, tenantID)
```

#### 3. **Performance Degradation with Large Schemas**

**Problem:** Enrichment is slow for schemas with many fields.

**Solution:**
```go
// Enable lazy enrichment for large forms
enricher.SetLazyMode(true)

// Or enrich only visible sections
visibleSection := c.Query("section", "basic")
enriched, _ := enricher.EnrichSection(ctx, schema, visibleSection, user)
```

#### 4. **Memory Leaks from Cloning**

**Problem:** Memory usage grows over time.

**Solution:**
```go
// Ensure schemas are properly garbage collected
func HandleRequest(c *fiber.Ctx) error {
    schema, _ := registry.Get(c.Context(), "form")
    
    // Clone in limited scope
    {
        schemaCopy := schema.Clone()
        enriched, _ := enricher.Enrich(c.Context(), schemaCopy, user)
        return c.JSON(enriched)
    } // schemaCopy eligible for GC here
}
```

---

## Monitoring & Observability

### Performance Metrics

```go
// Add instrumentation
func (e *Enricher) Enrich(ctx context.Context, schema *Schema, user *User) (*Schema, error) {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        metrics.RecordEnrichmentDuration(schema.ID, duration)
        
        if duration > 100*time.Millisecond {
            log.Warn().
                Str("schema_id", schema.ID).
                Dur("duration", duration).
                Msg("Slow enrichment detected")
        }
    }()
    
    // ... enrichment logic
}
```

### Audit Logging

```go
func (e *Enricher) auditEnrichment(ctx context.Context, result *EnrichmentResult, user *User) {
    log.Info().
        Str("schema_id", result.Schema.ID).
        Str("user_id", user.ID).
        Strs("fields_hidden", result.FieldsHidden).
        Strs("fields_modified", result.FieldsModified).
        Dur("duration", result.EnrichmentTime).
        Msg("Schema enriched")
}
```

---

## Best Practices Summary

✅ **DO:**
- Always enrich on the server side
- Clone schemas before enriching to avoid cache mutations
- Cache tenant overrides and permission checks
- Validate enrichment results before rendering
- Log slow enrichments for performance monitoring
- Re-enrich on every request to ensure fresh permissions
- Use middleware for consistent enrichment

❌ **DON'T:**
- Accept enriched schemas from clients
- Skip permission checks for "trusted" users
- Cache enriched schemas per user (cache base schemas only)
- Expose permission rules or internal metadata to clients
- Assume permissions are static (always re-check)
- Enrich schemas synchronously during high-traffic operations

---

[← Back: Registry](10-registry.md) | [Next: Renderer →](12-renderer.md)
