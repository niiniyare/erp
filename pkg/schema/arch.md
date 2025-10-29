# Schema System Architecture

## Overview

The schema system has been modularized into focused files, each handling specific concerns. This makes the codebase maintainable, testable, and easy to extend.

## File Organization

### Core Files

**schema.go** (Main Entry Point)
- Core `Schema` struct definition
- Type definitions (`Type` enum)
- Runtime state (`State`, `Context`)
- Core configuration (`Config`)
- Main interfaces (SchemaValidator, SchemaRenderer, SchemaRegistry, etc.)
- Core helper methods (GetField, HasField, GetVisibleFields, etc.)
- Circular dependency detection
- Schema validation logic
- JSON marshaling/unmarshaling

**field.go** (Field System)
- `Field` struct and all field-related types
- 30+ field types (text, email, select, date, file, etc.)
- Field validation configuration (`FieldValidation`, `Messages`)
- Field transformations and masking
- Field layout and styling
- Field events and conditions
- Field permissions and security
- HTMX and Alpine.js integration for fields
- Field-level helper methods

**action.go** (Actions/Buttons)
- `Action` struct for buttons and interactive elements
- Action types (submit, reset, button, link, custom)
- Action configuration and behavior
- Confirmation dialogs
- Action permissions
- HTMX and Alpine.js integration for actions
- Action styling helpers (Flowbite CSS classes)

**layout.go** (Visual Layout)
- `Layout` struct and layout types
- Grid, flex, tabs, steps, sections layouts
- `Section`, `Group`, `Tab`, `Step` definitions
- Responsive breakpoints
- Layout validation
- Helper methods for retrieving layout elements

**enterprise.go** (Enterprise Features)
- Security configuration (CSRF, rate limiting, encryption)
- Multi-tenancy configuration
- Workflow and approval system
- Cross-field validation
- Lifecycle events
- Internationalization (I18n)
- HTMX/Alpine.js framework integration
- Metadata and changelog

### Support Files

**errors.go** (Error Handling)
- Standard error variables
- `ValidationError` with context
- `ValidationErrors` collection
- `PermissionError` for authorization
- `TenantError` for isolation violations
- `RenderError` for rendering issues
- `DataSourceError` for data fetching
- Error helper functions

**builder.go** (Fluent API)
- `Builder` for programmatic schema creation
- `FieldBuilder` for field creation
- Fluent chaining methods
- Helper constructors (NewGridLayout, NewSimpleConfig, etc.)
- Option creation helpers

**examples.go** (Usage Patterns)
- Real-world example schemas
- User registration form
- Product form with sections
- Multi-tenant invoice form
- Multi-step employee onboarding wizard
- Expense report with approval workflow
- Field builder examples

**doc.go** (Package Documentation)
- Go package documentation
- Usage examples
- Core concepts explanation
- Best practices

## Design Principles

### 1. Single Responsibility
Each file handles one aspect of the system:
- `field.go` → Everything about fields
- `action.go` → Everything about actions
- `layout.go` → Everything about layouts
- etc.

### 2. Progressive Disclosure
- Simple use cases are simple
- Complex features available when needed
- Builder pattern for programmatic creation
- JSON for declarative definition

### 3. Type Safety
- Strong typing throughout
- Validation at multiple levels
- Compile-time type checking
- Runtime validation

### 4. Extensibility
- Interface-based design
- Plugin points for custom behavior
- Framework agnostic core
- HTMX/Alpine.js integration optional

### 5. Enterprise Ready
- Multi-tenancy built-in
- Security by default
- Workflow support
- Audit trails and metadata

## Key Interfaces

### Core Interfaces (in schema.go)

```go
// Validates schema structure and data
type SchemaValidator interface {
    ValidateSchema(ctx context.Context, schema *Schema) error
    ValidateData(ctx context.Context, schema *Schema, data map[string]any) error
}

// Renders schema to HTML/templates
type SchemaRenderer interface {
    Render(ctx context.Context, schema *Schema, data map[string]any) (string, error)
    RenderField(ctx context.Context, field *Field, value any) (string, error)
}

// Manages schema storage
type SchemaRegistry interface {
    Register(ctx context.Context, schema *Schema) error
    Get(ctx context.Context, id string) (*Schema, error)
    List(ctx context.Context, filter map[string]any) ([]*Schema, error)
    Update(ctx context.Context, schema *Schema) error
    Delete(ctx context.Context, id string) error
    GetVersion(ctx context.Context, id, version string) (*Schema, error)
}

// Evaluates conditions
type ConditionEvaluator interface {
    Evaluate(ctx context.Context, condition *condition.ConditionGroup, data map[string]any) (bool, error)
    Compile(condition *condition.ConditionGroup) error
}

// Resolves dynamic data sources
type DataSourceResolver interface {
    Resolve(ctx context.Context, source *DataSource, params map[string]string) ([]Option, error)
    InvalidateCache(ctx context.Context, source *DataSource) error
}

// Transforms field values
type TransformProcessor interface {
    Transform(ctx context.Context, transform *Transform, value any) (any, error)
}
```

## Usage Patterns

### 1. Simple Form (Builder Pattern)

```go
schema := models.NewBuilder("user-form", models.TypeForm, "Create User").
    WithConfig(models.NewSimpleConfig("/api/v1/users", "POST")).
    WithCSRF().
    AddTextField("name", "Name", true).
    AddEmailField("email", "Email", true).
    AddSubmitButton("Save").
    MustBuild()
```

### 2. Complex Form (Declarative)

Load from JSON file:
```json
{
  "id": "invoice-form",
  "type": "form",
  "title": "Create Invoice",
  "fields": [...],
  "actions": [...]
}
```

Parse in Go:
```go
var schema models.Schema
json.Unmarshal(data, &schema)
models.Validate() // Automatic validation
```

### 3. Implementation Flow

```
JSON File → Parse → Schema Struct → Validate → Render → HTML
                                               ↓
                                         Submit → Validate Data → Process
```

## Extension Points

### Custom Validators

```go
type MyValidator struct{}

func (v *MyValidator) ValidateSchema(ctx context.Context, s *models.Schema) error {
    // Custom validation logic
}

func (v *MyValidator) ValidateData(ctx context.Context, s *models.Schema, data map[string]any) error {
    // Custom data validation
}
```

### Custom Renderers

```go
type TemplRenderer struct {
    templates *template.Template
}

func (r *TemplRenderer) Render(ctx context.Context, s *models.Schema, data map[string]any) (string, error) {
    return r.templates.ExecuteTemplate(writer, "form", s)
}
```

### Custom Data Sources

```go
type DatabaseDataSource struct {
    db *sql.DB
}

func (ds *DatabaseDataSource) Resolve(ctx context.Context, source *models.DataSource, params map[string]string) ([]models.Option, error) {
    // Query database and return options
}
```

## Performance Considerations

### 1. Condition Compilation
Pre-compile conditions for repeated evaluation:
```go
// Compile once
compiledCondition := evaluator.Compile(field.Conditional.Show)
field.compiledCondition = compiledCondition

// Evaluate many times (fast)
result := compiledCondition.Evaluate(data)
```

### 2. Schema Caching
Cache compiled schemas in registry:
```go
type CachedRegistry struct {
    cache map[string]*CompiledSchema
}
```

### 3. Data Source Caching
Use CacheTTL for dynamic options:
```go
DataSource: &DataSource{
    Type: "api",
    URL: "/api/v1/customers",
    CacheTTL: 300, // 5 minutes
}
```

## Integration with Awo ERP

### 1. Multi-Tenancy
Integrates with your RLS-based tenant isolation:
```go
models.Tenant = &Tenant{
    Enabled: true,
    Field: "tenant_id",
    Isolation: "strict", // Uses RLS policies
}
```

### 2. Permissions
Integrates with your RBAC/ABAC system:
```go
field.Permissions = &FieldPermissions{
    View: []string{"role:admin", "role:manager"},
    Edit: []string{"role:admin"},
}
```

### 3. Workflow
Integrates with your approval system:
```go
models.Workflow = &Workflow{
    Enabled: true,
    Approvals: &ApprovalConfig{...},
}
```

## Future Enhancements

### Potential Additions

1. **Schema Versioning**: Migration system for schema updates
2. **Schema Inheritance**: Extend base schemas
3. **Formula Engine**: Excel-like formulas for computed fields
4. **Advanced Layouts**: Kanban, calendar, timeline views
5. **Real-time Collaboration**: Multi-user editing
6. **Schema IDE**: Visual schema builder
7. **Theme System**: Complete UI theming
8. **Plugin System**: Third-party extensions

## Best Practices

### For Developers

1. **Use the Builder**: More readable than raw struct creation
2. **Validate Early**: Use `MustBuild()` during development
3. **Type Constants**: Always use type constants (FieldText, not "text")
4. **Add Examples**: Document complex schemas in examples.go
5. **Test Thoroughly**: Write tests for custom validators/renderers

### For Schema Authors

1. **Keep It Simple**: Start with basic fields, add complexity as needed
2. **Use Help Text**: Guide users with descriptions and tooltips
3. **Group Logically**: Use sections/tabs to organize fields
4. **Validate Early**: Use `onBlur` validation for better UX
5. **Test Conditions**: Verify conditional logic works correctly

## Testing Strategy

### Unit Tests
- Field validation logic
- Circular dependency detection
- Error handling
- Builder pattern

### Integration Tests
- Schema parsing from JSON
- Rendering to HTML
- Data submission and validation
- Multi-tenancy isolation

### E2E Tests
- Complete form workflows
- Multi-step wizards
- Approval processes
- File uploads

## Conclusion

The schema system provides a solid foundation for building enterprise forms and UIs. The modular architecture makes it easy to understand, extend, and maintain. Each file has a clear purpose, and the builder pattern makes programmatic creation intuitive.

The system is production-ready and integrates seamlessly with your existing Awo ERP architecture (multi-tenancy, security, workflows).
