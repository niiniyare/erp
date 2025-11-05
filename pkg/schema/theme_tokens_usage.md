# Theme & Token Usage Guide

## Overview

The schema package provides a modern three-tier theme and token system that replaces legacy field-based theming. This system follows a **Primitives → Semantic → Components** architecture where raw design values flow through functional assignments to component-specific styles. The system supports runtime customization, multi-tenant theming, and maintains type safety throughout the token resolution pipeline.

---

## 1. Creating and Registering Themes

Use `ThemeBuilder` for fluent theme construction and `ThemeManager` for registration and retrieval.

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/niiniyare/erp/pkg/schema"
)

func main() {
    // Get the global theme manager
    manager := schema.GetGlobalThemeManager()
    
    // Create a corporate theme with custom tokens
    theme, err := schema.NewTheme("Corporate Theme").
        WithID("corporate-v2").
        WithDescription("Modern corporate theme with blue accent").
        WithVersion("2.0.0").
        WithAuthor("Design Team").
        WithTokens(schema.GetDefaultTokens()).
        WithDarkMode(&schema.DarkModeConfig{
            Enabled:  true,
            Strategy: "class",
        }).
        WithAccessibility(&schema.AccessibilityConfig{
            HighContrast:     true,
            MinContrastRatio: 4.5,
            KeyboardNav:      true,
        }).
        Build()
    
    if err != nil {
        log.Fatal("Failed to build theme:", err)
    }
    
    // Register the theme
    err = manager.RegisterTheme(theme)
    if err != nil {
        log.Fatal("Failed to register theme:", err)
    }
    
    // List all available themes
    themes, _ := manager.ListThemes(context.Background())
    fmt.Printf("Available themes: %d\n", len(themes))
    for _, t := range themes {
        fmt.Printf("- %s (%s)\n", t.Name, t.ID)
    }
}
```

---

## 2. Token Resolution

Resolve token references using the `{token.path}` syntax. The resolver handles circular references and caches results for performance.

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/niiniyare/erp/pkg/schema"
)

func main() {
    // Get token registry
    registry := schema.GetDefaultRegistry()
    ctx := context.Background()
    
    // Resolve individual tokens
    primaryColor, err := registry.ResolveToken(ctx, schema.TokenReference("{colors.blue.500}"))
    if err != nil {
        log.Fatal("Token resolution failed:", err)
    }
    fmt.Printf("Primary color: %s\n", primaryColor)
    
    // Resolve semantic tokens
    bgColor, err := registry.ResolveToken(ctx, schema.TokenReference("{semantic.colors.background.default}"))
    if err != nil {
        log.Fatal("Semantic token resolution failed:", err)
    }
    fmt.Printf("Background color: %s\n", bgColor)
    
    // Resolve component tokens
    buttonBg, err := registry.ResolveToken(ctx, schema.TokenReference("{components.button.primary.background}"))
    if err != nil {
        log.Fatal("Component token resolution failed:", err)
    }
    fmt.Printf("Button background: %s\n", buttonBg)
    
    // Invalid token reference (will error)
    _, err = registry.ResolveToken(ctx, schema.TokenReference("{invalid.token.path}"))
    if err != nil {
        fmt.Printf("Expected error for invalid token: %v\n", err)
    }
}
```

**Note**: Token structures are immutable by design. The resolver automatically detects circular references and returns validation errors for malformed token paths.

---

## 3. Applying Themes to Schemas

Integrate themes with schema objects for consistent styling across forms and components.

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/niiniyare/erp/pkg/schema"
)

func main() {
    ctx := context.Background()
    
    // Create a new schema
    schemaBuilder := schema.NewBuilder("invoice-form").
        WithTitle("Invoice Creation").
        WithDescription("Create new customer invoices")
    
    // Add some fields
    schemaBuilder.AddField(&schema.Field{
        Name:  "customer_name",
        Type:  schema.FieldTypeText,
        Label: "Customer Name",
        Required: true,
    })
    
    schemaBuilder.AddField(&schema.Field{
        Name:  "amount",
        Type:  schema.FieldTypeNumber,
        Label: "Amount",
        Required: true,
    })
    
    invoiceSchema, err := schemaBuilder.Build()
    if err != nil {
        log.Fatal("Failed to build schema:", err)
    }
    
    // Apply theme to schema
    err = invoiceSchema.ApplyTheme(ctx, "corporate-v2")
    if err != nil {
        log.Fatal("Failed to apply theme:", err)
    }
    
    fmt.Printf("Schema '%s' themed with: %s\n", 
        invoiceSchema.Title, 
        invoiceSchema.Meta.Theme.ID)
    
    // Apply theme with custom overrides
    overrides := &schema.ThemeOverrides{
        TokenOverrides: map[string]string{
            "colors.primary.500": "#dc2626", // Custom red primary
        },
        CustomCSS: `.invoice-form { border: 2px solid var(--primary-500); }`,
    }
    
    err = invoiceSchema.ApplyThemeWithOverrides(ctx, "corporate-v2", overrides)
    if err != nil {
        log.Fatal("Failed to apply theme with overrides:", err)
    }
    
    fmt.Println("Applied theme with custom red primary color")
}
```

**Multi-tenant Context**: In multi-tenant setups, themes are automatically resolved based on the tenant context key in the request context.

---

## 4. Runtime Theme Customization

Customize themes dynamically for tenant-specific branding and user preferences.

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/niiniyare/erp/pkg/schema"
)

func main() {
    manager := schema.GetGlobalThemeManager()
    tenantManager := schema.GetGlobalTenantManager()
    
    // Set up tenant-specific theme
    ctx := schema.WithTenantContext(context.Background(), "tenant-acme")
    
    // Configure tenant theme with overrides
    overrides := &schema.ThemeOverrides{
        TokenOverrides: map[string]string{
            "colors.primary.500":   "#f59e0b", // ACME brand amber
            "colors.primary.600":   "#d97706",
            "typography.brand.family": "'ACME Sans', sans-serif",
        },
        ComponentOverrides: map[string]interface{}{
            "button.borderRadius": "8px",
            "card.shadow":        "0 4px 12px rgba(0,0,0,0.1)",
        },
        CustomCSS: `
            .header { background: linear-gradient(135deg, #f59e0b, #d97706); }
            .logo { font-family: 'ACME Sans', sans-serif; }
        `,
    }
    
    err := tenantManager.SetTenantTheme(ctx, "tenant-acme", "corporate-v2", overrides)
    if err != nil {
        log.Fatal("Failed to set tenant theme:", err)
    }
    
    // Retrieve tenant-specific theme
    tenantTheme, err := tenantManager.GetTenantTheme(ctx, "tenant-acme")
    if err != nil {
        log.Fatal("Failed to get tenant theme:", err)
    }
    
    fmt.Printf("Tenant theme: %s\n", tenantTheme.Name)
    
    // Theme resolution with context
    contextTheme, err := schema.GetThemeFromContext(ctx)
    if err != nil {
        log.Fatal("Failed to resolve theme from context:", err)
    }
    
    fmt.Printf("Context-resolved theme: %s\n", contextTheme.ID)
    
    // Get theme with runtime overrides
    runtimeTheme, err := manager.GetThemeWithOverrides(ctx, "corporate-v2", &schema.ThemeOverrides{
        TokenOverrides: map[string]string{
            "spacing.base": "1rem",
        },
    })
    if err != nil {
        log.Fatal("Failed to get theme with overrides:", err)
    }
    
    fmt.Printf("Runtime customized theme: %s\n", runtimeTheme.Name)
}
```

---

## 5. Importing & Exporting Themes

Use JSON serialization for theme storage, version control, and cross-environment transport.

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/niiniyare/erp/pkg/schema"
)

func main() {
    // Create a theme for export
    theme, err := schema.NewTheme("Export Example").
        WithID("export-demo").
        WithVersion("1.0.0").
        WithTokens(schema.GetDefaultTokens()).
        Build()
    
    if err != nil {
        log.Fatal("Failed to create theme:", err)
    }
    
    // Export theme to JSON
    jsonData, err := schema.ExportTheme(theme)
    if err != nil {
        log.Fatal("Failed to export theme:", err)
    }
    
    fmt.Printf("Exported theme JSON (%d bytes):\n", len(jsonData))
    fmt.Printf("%.200s...\n", string(jsonData)) // First 200 chars
    
    // Import theme from JSON
    imported, err := schema.ImportTheme(jsonData)
    if err != nil {
        log.Fatal("Failed to import theme:", err)
    }
    
    fmt.Printf("Imported theme: %s (v%s)\n", imported.Name, imported.Version)
    fmt.Printf("Token count: %d primitives\n", len(imported.Tokens.Primitives.Colors.Blue.Scale500))
    
    // Validate imported theme
    err = schema.ValidateTheme(imported)
    if err != nil {
        log.Printf("Theme validation warning: %v\n", err)
    } else {
        fmt.Println("Theme validation: PASSED")
    }
    
    // Register imported theme
    manager := schema.GetGlobalThemeManager()
    err = manager.RegisterTheme(imported)
    if err != nil {
        log.Fatal("Failed to register imported theme:", err)
    }
    
    fmt.Println("Successfully imported and registered theme")
}
```

**Note**: JSON export includes versioning metadata and supports cross-environment theme transport. Always validate imported themes before registration.

---

## 6. Testing & Validation Examples

Write comprehensive tests for token resolution, concurrency safety, and theme overrides.

```go
package schema_test

import (
    "context"
    "fmt"
    "sync"
    "testing"
    
    "github.com/stretchr/testify/suite"
    "github.com/niiniyare/erp/pkg/schema"
)

type ThemeTestSuite struct {
    suite.Suite
    manager *schema.ThemeManager
    ctx     context.Context
}

func (suite *ThemeTestSuite) SetupTest() {
    suite.manager = schema.NewThemeManager(nil, nil)
    suite.ctx = context.Background()
}

// Test token resolution with circular reference detection
func (suite *ThemeTestSuite) TestTokenResolution() {
    registry := schema.GetDefaultRegistry()
    
    // Valid token resolution
    resolved, err := registry.ResolveToken(suite.ctx, schema.TokenReference("{colors.blue.500}"))
    suite.NoError(err)
    suite.NotEmpty(resolved)
    
    // Test circular reference detection
    tokens := schema.GetDefaultTokens()
    
    // Create circular reference (this would be caught by validation)
    tokens.Semantic.Colors.Background.Default = schema.TokenReference("{semantic.colors.background.default}")
    
    err = registry.SetTokens(tokens)
    suite.Error(err, "Should detect circular reference")
    suite.Contains(err.Error(), "circular reference")
}

// Test concurrent theme access
func (suite *ThemeTestSuite) TestConcurrentThemeAccess() {
    var wg sync.WaitGroup
    errorChan := make(chan error, 100)
    
    // Concurrent theme registration
    for i := 0; i < 50; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            theme, err := schema.NewTheme(fmt.Sprintf("Concurrent Theme %d", id)).
                WithID(fmt.Sprintf("concurrent-%d", id)).
                WithTokens(schema.GetDefaultTokens()).
                Build()
            
            if err != nil {
                errorChan <- err
                return
            }
            
            err = suite.manager.RegisterTheme(theme)
            if err != nil {
                errorChan <- err
            }
        }(i)
    }
    
    wg.Wait()
    close(errorChan)
    
    // Check for errors
    for err := range errorChan {
        suite.Fail("Concurrent access error: %v", err)
    }
    
    // Verify all themes were registered
    themes, err := suite.manager.ListThemes(suite.ctx)
    suite.NoError(err)
    suite.GreaterOrEqual(len(themes), 50)
}

// Test theme overrides
func (suite *ThemeTestSuite) TestThemeOverrides() {
    // Create base theme
    baseTheme, err := schema.NewTheme("Base Theme").
        WithID("base").
        WithTokens(schema.GetDefaultTokens()).
        Build()
    suite.NoError(err)
    
    err = suite.manager.RegisterTheme(baseTheme)
    suite.NoError(err)
    
    // Apply overrides
    overrides := &schema.ThemeOverrides{
        TokenOverrides: map[string]string{
            "colors.primary.500": "#custom-color",
        },
        CustomCSS: ".custom { color: red; }",
    }
    
    customTheme, err := suite.manager.GetThemeWithOverrides(suite.ctx, "base", overrides)
    suite.NoError(err)
    suite.NotNil(customTheme)
    suite.Contains(customTheme.CustomCSS, ".custom { color: red; }")
}

// Benchmark token resolution performance
func BenchmarkTokenResolution(b *testing.B) {
    registry := schema.GetDefaultRegistry()
    ctx := context.Background()
    
    tokenRef := schema.TokenReference("{semantic.colors.background.default}")
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := registry.ResolveToken(ctx, tokenRef)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func TestThemeSystemSuite(t *testing.T) {
    suite.Run(t, new(ThemeTestSuite))
}
```

**Coverage Goal**: Maintain >85% test coverage. See `pkg/schema/execution_plan.md` for the complete test suite structure and critical test scenarios.

---

## 7. Recommended Conventions

Follow these naming and design conventions for consistency across the ERP system:

### Token Path Conventions
```go
// ✅ Good - Semantic token usage
background := resolver.ResolveToken(ctx, "{semantic.colors.background.default}")
primaryBtn := resolver.ResolveToken(ctx, "{components.button.primary.background}")

// ❌ Avoid - Direct primitive access
blueColor := resolver.ResolveToken(ctx, "{colors.blue.500}")
```

### Token Naming Structure
- **Primitives**: `colors.blue.500`, `spacing.md`, `typography.sans.family`
- **Semantic**: `semantic.interactive.primary`, `semantic.feedback.success`
- **Components**: `components.button.primary.background`, `components.card.border`

### Theme Versioning
```go
// ✅ Semantic versioning
theme.WithVersion("2.1.0")

// ✅ Environment-specific themes
theme.WithID("corporate-prod-v2")
theme.WithID("corporate-staging-v2")
```

### Multi-tenant Patterns
```go
// ✅ Tenant context propagation
ctx = schema.WithTenantContext(ctx, tenantID)
ctx = schema.WithThemeContext(ctx, themeID)

// ✅ Tenant-specific theme naming
tenantThemeID := fmt.Sprintf("%s-tenant-%s", baseThemeID, tenantID)
```

### Error Handling
```go
// ✅ Use schema error types
if schema.IsValidationError(err) {
    // Handle validation errors
}

if schema.IsNotFoundError(err) {
    // Handle missing theme/token
}
```

---

## 8. Migration Notes

Migrate from legacy field-based theming to the token system:

### Before (Legacy)
```go
// Old field-based approach
field.Theme = &FieldTheme{
    BackgroundColor: "#ffffff",
    BorderColor:     "#e5e7eb",
    TextColor:       "#111827",
}
```

### After (Token-based)
```go
// New token-based approach
field.Tokens = map[string]schema.TokenReference{
    "background": "{semantic.colors.background.default}",
    "border":     "{semantic.colors.border.default}",
    "text":       "{semantic.colors.text.primary}",
}

// Or resolve directly
bg, _ := registry.ResolveToken(ctx, "{semantic.colors.background.default}")
```

### Migration Helper
```go
// Temporary helper for transitional compatibility
func GetLegacyToken(field *schema.Field, property string) string {
    if field.Tokens != nil {
        if tokenRef, exists := field.Tokens[property]; exists {
            resolved, _ := registry.ResolveToken(ctx, tokenRef)
            return resolved
        }
    }
    // Fallback to legacy field values
    return field.GetLegacyValue(property)
}
```

### Breaking Changes
- `FormTheme`, `FieldTheme`, `ButtonTheme` structs removed
- Direct color/spacing properties replaced with token references  
- Theme application now requires `context.Context`
- Caching behavior changed from field-level to token-level

### Compatibility Layer
The system provides backward compatibility during the migration period. Use `schema.EnableLegacyCompat()` to maintain existing functionality while migrating incrementally.

---

**Design Alignment**: This system follows the immutability, extensibility, and clarity principles outlined in `pkg/schema/design.md`. Token resolution is deterministic, theme structures are validated at build time, and the API surface remains minimal while supporting advanced use cases.