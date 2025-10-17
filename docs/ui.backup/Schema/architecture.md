# Schema-Driven UI Architecture

**Version**: 2.0  
**Date**: October 2025  
**Status**: Production-Ready Architecture  

---

## Executive Summary

This document presents the architectural foundation for a schema-driven UI rendering framework built on Go, Templ, HTMX, and Alpine.js. The framework transforms JSON schemas into fully functional, type-safe user interfaces at runtime, enabling backend services to dynamically generate sophisticated UIs without frontend rebuilds or deployments.

**Core Innovation**: Unlike client-heavy solutions such as React-based schema renderers, this system leverages server-side rendering with strategic client-side enhancement, delivering superior performance while maintaining the flexibility of runtime schema interpretation.

**Current State**: The framework includes 900+ comprehensive JSON Schema definitions covering the complete spectrum of enterprise UI components, from atomic elements to complex data management interfaces. These schemas form a mature, battle-tested foundation ready for integration with existing Templ component implementations.

**Architectural Advantage**: The design bridges proven technologies - extensive JSON Schema definitions (derived from production systems) with Go's type safety, Templ's performance, HTMX's simplicity, and Alpine.js's reactivity - creating a powerful platform for enterprise application development.

---

## Problem Statement

Modern enterprise applications face competing demands:

- **Dynamic UI Requirements**: Multi-tenant systems need customizable interfaces without code changes
- **Performance Constraints**: Heavy client-side frameworks impact load times and resource usage
- **Development Velocity**: Backend-driven development should not require frontend expertise
- **Type Safety**: Runtime UI generation must maintain compile-time correctness guarantees
- **Maintenance Burden**: UI logic scattered across client and server creates complexity

## Solution Overview

This framework addresses these challenges through a sophisticated schema-driven architecture that:

**Enables Runtime UI Generation**: Arbitrary JSON schemas processed at request time create unlimited interface variations without build-time dependencies.

**Maintains Type Safety**: Go's type system combined with JSON Schema validation provides multiple layers of correctness verification.

**Optimizes Performance**: Server-side rendering eliminates client-side framework overhead while HTMX provides efficient partial updates.

**Simplifies Development**: Backend developers author schemas with IDE support and validation, while UI complexity remains encapsulated in reusable components.

**Supports Enterprise Scale**: Multi-tenant architecture, extensive component library, and comprehensive validation enable production deployment.

## Technical Philosophy

The framework operates on three foundational principles:

**Server-First Architecture**: Business logic, validation, and primary rendering occur server-side. Client-side JavaScript enhances rather than drives the experience.

**Schema as Contract**: JSON schemas serve as the interface contract between backend logic and frontend presentation, enabling independent evolution of both layers.

**Component Composition**: Complex interfaces emerge from simple, reusable components following atomic design principles. No monolithic components or special cases.

---

## System Architecture

### High-Level Design

```
┌─────────────────────┐    ┌─────────────────────┐    ┌─────────────────────┐
│   JSON Schema       │    │   Component         │    │   Rendered UI       │
│   Definition        │───▶│   Registry          │───▶│   (Templ + HTMX)    │
│                     │    │                     │    │                     │
└─────────────────────┘    └─────────────────────┘    └─────────────────────┘
           │                           │                           │
           │                           │                           │
           ▼                           ▼                           ▼
┌─────────────────────┐    ┌─────────────────────┐    ┌─────────────────────┐
│   Schema            │    │   Component         │    │   Client-Side       │
│   Validation        │    │   Generation        │    │   Enhancement       │
│                     │    │                     │    │   (Alpine.js)       │
└─────────────────────┘    └─────────────────────┘    └─────────────────────┘
```

### Request Flow Architecture

```
HTTP Request
     │
     ▼
┌─────────────────────┐
│  Fiber Middleware   │ ◄─── Extract Tenant Context
└─────────────────────┘
     │
     ▼
┌─────────────────────┐
│  Schema Resolver    │ ◄─── Load & Validate Schema
└─────────────────────┘
     │
     ▼
┌─────────────────────┐
│  Component Registry │ ◄─── Map Schema to Components
└─────────────────────┘
     │
     ▼
┌─────────────────────┐
│  Templ Renderer     │ ◄─── Generate Type-Safe HTML
└─────────────────────┘
     │
     ▼
┌─────────────────────┐
│  HTMX + Alpine.js   │ ◄─── Client Enhancement
└─────────────────────┘
     │
     ▼
   Response
```

### Layer Separation

**Presentation Layer (Client)**:
- Templ-generated HTML with semantic structure
- HTMX for server communication patterns
- Alpine.js for client-side reactivity and state
- Flowbite CSS for consistent styling

**Application Layer (Server)**:
- Fiber HTTP routing and middleware
- Schema resolution and validation
- Component registry and mapping
- Business logic orchestration

**Domain Layer (Core)**:
- JSON Schema definitions (900+)
- Component interface contracts
- Validation rules and types
- Business entity models

**Infrastructure Layer (Platform)**:
- Multi-tenant data isolation (PostgreSQL RLS)
- Caching strategies (Redis)
- Security enforcement (ABAC)
- Audit and monitoring systems

---

## Multi-Tenant Architecture

### Core Design Principles

**Shared Database with Row-Level Security**: Single database instance with PostgreSQL RLS ensuring tenant isolation at the data layer.

**Schema-Level Tenant Context**: Tenant information embedded in schema definitions enables per-tenant customization without code changes.

**Component-Level Isolation**: Components can render differently based on tenant configuration while maintaining the same interface contract.

### Tenant Context Flow

```go
// Middleware extracts tenant from request
func TenantMiddleware(c *fiber.Ctx) error {
    tenantID := extractTenantFromRequest(c)
    c.Locals("tenant_id", tenantID)
    return c.Next()
}

// Schema resolver includes tenant context
func ResolveSchema(ctx context.Context, schemaID string) (*Schema, error) {
    tenantID := ctx.Value("tenant_id").(string)
    
    // Load base schema
    schema, err := loadBaseSchema(schemaID)
    if err != nil {
        return nil, err
    }
    
    // Apply tenant-specific overrides
    tenantOverrides, err := loadTenantOverrides(tenantID, schemaID)
    if err != nil {
        return nil, err
    }
    
    return mergeSchemaWithOverrides(schema, tenantOverrides), nil
}
```

### Database-Level Security

PostgreSQL Row-Level Security automatically filters data based on tenant context:

```sql
-- Enable RLS on all tenant-aware tables
ALTER TABLE accounts ENABLE ROW LEVEL SECURITY;

-- Create policy for tenant isolation
CREATE POLICY tenant_isolation ON accounts
    USING (tenant_id = current_tenant_id());

-- Tenant context functions
CREATE OR REPLACE FUNCTION set_tenant_context(uuid)
RETURNS void AS $$
BEGIN
    PERFORM set_config('app.current_tenant_id', $1::text, true);
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION current_tenant_id()
RETURNS uuid AS $$
BEGIN
    RETURN COALESCE(
        current_setting('app.current_tenant_id', true)::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid
    );
END;
$$ LANGUAGE plpgsql;
```

---

## Schema System Architecture

### Schema Definition Structure

The framework employs a hierarchical schema organization with clear inheritance patterns:

```
BaseSchema
├── ControlSchema (Input Elements)
│   ├── TextControlSchema
│   ├── NumberControlSchema
│   ├── SelectControlSchema
│   └── DateControlSchema
├── ActionSchema (Interactive Elements)
│   ├── AjaxActionSchema
│   ├── DialogActionSchema
│   └── NavigationActionSchema
├── LayoutSchema (Container Elements)
│   ├── GridSchema
│   ├── FlexSchema
│   └── ContainerSchema
└── DataSchema (Data Display)
    ├── TableSchema
    ├── ListSchema
    └── CardSchema
```

### JSON Schema Validation Pipeline

```go
type SchemaValidator struct {
    compiledSchemas map[string]*jsonschema.Schema
    metaValidator   *jsonschema.Schema
}

func (v *SchemaValidator) ValidateSchema(schemaData []byte) error {
    // 1. Parse JSON
    var rawSchema map[string]interface{}
    if err := json.Unmarshal(schemaData, &rawSchema); err != nil {
        return fmt.Errorf("invalid JSON: %w", err)
    }
    
    // 2. Validate against meta-schema
    if err := v.metaValidator.Validate(rawSchema); err != nil {
        return fmt.Errorf("schema validation failed: %w", err)
    }
    
    // 3. Type-specific validation
    schemaType, ok := rawSchema["type"].(string)
    if !ok {
        return errors.New("schema must specify type")
    }
    
    typeValidator, exists := v.compiledSchemas[schemaType]
    if !exists {
        return fmt.Errorf("unknown schema type: %s", schemaType)
    }
    
    return typeValidator.Validate(rawSchema)
}
```

### Schema Composition and Inheritance

Complex schemas build from simpler base definitions using JSON Schema composition:

```json
{
  "allOf": [
    {"$ref": "#/definitions/BaseSchema"},
    {"$ref": "#/definitions/ControlSchema"},
    {
      "properties": {
        "type": {"const": "text"},
        "placeholder": {"type": "string"},
        "maxLength": {"type": "number"}
      }
    }
  ]
}
```

### Runtime Schema Processing

```go
type SchemaProcessor struct {
    registry    *ComponentRegistry
    validator   *SchemaValidator
    cache       *SchemaCache
}

func (p *SchemaProcessor) ProcessSchema(
    ctx context.Context, 
    schemaData []byte,
) (*ProcessedSchema, error) {
    // 1. Validate schema
    if err := p.validator.ValidateSchema(schemaData); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    // 2. Parse and resolve references
    schema, err := p.parseAndResolveRefs(schemaData)
    if err != nil {
        return nil, fmt.Errorf("reference resolution failed: %w", err)
    }
    
    // 3. Apply tenant-specific overrides
    tenantID := getTenantFromContext(ctx)
    if tenantID != "" {
        schema = p.applyTenantOverrides(schema, tenantID)
    }
    
    // 4. Generate component mapping
    componentTree, err := p.registry.MapSchemaToComponents(schema)
    if err != nil {
        return nil, fmt.Errorf("component mapping failed: %w", err)
    }
    
    return &ProcessedSchema{
        Original:      schema,
        ComponentTree: componentTree,
        Metadata:      extractMetadata(schema),
    }, nil
}
```

---

## Component Registry Architecture

### Registry Design Pattern

The Component Registry serves as the central mapping system between schema types and their corresponding Templ component implementations:

```go
type ComponentRegistry struct {
    components map[string]ComponentFactory
    mu         sync.RWMutex
}

type ComponentFactory interface {
    CreateComponent(schema *Schema) (TemplComponent, error)
    ValidateSchema(schema *Schema) error
    GetSupportedSchemaTypes() []string
}

type TemplComponent interface {
    Render(ctx context.Context, props ComponentProps) templ.Component
    GetDefaultProps() ComponentProps
    ValidateProps(props ComponentProps) error
}
```

### Component Registration

```go
func (r *ComponentRegistry) Register(
    schemaType string, 
    factory ComponentFactory,
) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    if _, exists := r.components[schemaType]; exists {
        return fmt.Errorf("component type %s already registered", schemaType)
    }
    
    // Validate factory can handle declared schema types
    supportedTypes := factory.GetSupportedSchemaTypes()
    for _, t := range supportedTypes {
        if t == schemaType {
            r.components[schemaType] = factory
            return nil
        }
    }
    
    return fmt.Errorf("factory does not support schema type %s", schemaType)
}
```

### Component Resolution and Rendering

```go
func (r *ComponentRegistry) ResolveAndRender(
    ctx context.Context,
    schema *Schema,
) (templ.Component, error) {
    r.mu.RLock()
    factory, exists := r.components[schema.Type]
    r.mu.RUnlock()
    
    if !exists {
        // Fallback to generic component
        factory = r.components["generic"]
    }
    
    // Create component instance
    component, err := factory.CreateComponent(schema)
    if err != nil {
        return nil, fmt.Errorf("component creation failed: %w", err)
    }
    
    // Convert schema to props
    props, err := schemaToProps(schema)
    if err != nil {
        return nil, fmt.Errorf("props conversion failed: %w", err)
    }
    
    // Validate props
    if err := component.ValidateProps(props); err != nil {
        return nil, fmt.Errorf("props validation failed: %w", err)
    }
    
    return component.Render(ctx, props), nil
}
```

---

## Performance Architecture

### Caching Strategy

**Multi-Level Caching**:
- **L1 (Memory)**: Parsed schemas and component instances
- **L2 (Redis)**: Rendered component HTML
- **L3 (CDN)**: Static assets and immutable content

```go
type CacheManager struct {
    memory    *sync.Map              // L1: In-memory cache
    redis     *redis.Client         // L2: Distributed cache
    ttl       time.Duration
}

func (c *CacheManager) GetRenderedComponent(
    key string,
) (string, bool) {
    // Check L1 first
    if value, ok := c.memory.Load(key); ok {
        return value.(string), true
    }
    
    // Check L2
    value, err := c.redis.Get(context.Background(), key).Result()
    if err == nil {
        // Populate L1
        c.memory.Store(key, value)
        return value, true
    }
    
    return "", false
}
```

### Schema Compilation

Pre-processing schemas for optimal runtime performance:

```go
type SchemaCompiler struct {
    validator    *SchemaValidator
    optimizer    *SchemaOptimizer
}

func (c *SchemaCompiler) CompileSchema(
    schema *Schema,
) (*CompiledSchema, error) {
    // 1. Resolve all $ref references
    resolved, err := c.resolveReferences(schema)
    if err != nil {
        return nil, err
    }
    
    // 2. Flatten inheritance hierarchies
    flattened := c.flattenInheritance(resolved)
    
    // 3. Pre-compute validation rules
    validationRules := c.extractValidationRules(flattened)
    
    // 4. Generate optimized component mapping
    componentMap := c.generateComponentMapping(flattened)
    
    return &CompiledSchema{
        Original:        schema,
        Resolved:        resolved,
        Flattened:       flattened,
        ValidationRules: validationRules,
        ComponentMap:    componentMap,
        CompiledAt:      time.Now(),
    }, nil
}
```

### Streaming and Suspense

For large, complex UIs, implement streaming rendering:

```go
templ ComplexUIWithSuspense(schema *Schema) {
    <div id="main-content">
        @StreamingComponent(schema.MainContent)
        
        <div 
            hx-get="/api/lazy-content" 
            hx-trigger="intersect once"
            hx-swap="outerHTML">
            <div class="loading-skeleton">
                Loading additional content...
            </div>
        </div>
    </div>
}

templ StreamingComponent(content *ContentSchema) {
    for _, item := range content.Items {
        if item.Priority == "high" {
            @RenderComponent(item)
        } else {
            <div 
                hx-get={ fmt.Sprintf("/api/lazy-item/%s", item.ID) }
                hx-trigger="intersect once"
                hx-swap="outerHTML">
                <div class="loading-placeholder"></div>
            </div>
        }
    }
}
```

---

## Security Architecture

### Input Validation and Sanitization

**Three-Layer Validation**:
1. **Schema Layer**: JSON Schema validation
2. **Application Layer**: Business rule validation  
3. **Infrastructure Layer**: Database constraint validation

```go
type SecurityValidator struct {
    schemaValidator   *SchemaValidator
    businessValidator *BusinessValidator
    sanitizer        *HTMLSanitizer
}

func (v *SecurityValidator) ValidateAndSanitize(
    input map[string]interface{},
    schema *Schema,
) (map[string]interface{}, error) {
    // 1. Schema validation
    if err := v.schemaValidator.Validate(input, schema); err != nil {
        return nil, fmt.Errorf("schema validation failed: %w", err)
    }
    
    // 2. Business rule validation
    if err := v.businessValidator.Validate(input, schema); err != nil {
        return nil, fmt.Errorf("business validation failed: %w", err)
    }
    
    // 3. Sanitize HTML content
    sanitized := make(map[string]interface{})
    for key, value := range input {
        if strValue, ok := value.(string); ok {
            sanitized[key] = v.sanitizer.Sanitize(strValue)
        } else {
            sanitized[key] = value
        }
    }
    
    return sanitized, nil
}
```

### CSRF Protection

```go
func CSRFMiddleware() fiber.Handler {
    return csrf.New(csrf.Config{
        KeyLookup:      "header:X-CSRF-Token",
        CookieName:     "__Host-csrf_",
        CookieSameSite: "Strict",
        CookieSecure:   true,
        CookieHTTPOnly: true,
        Expiration:     1 * time.Hour,
        KeyGenerator:   utils.UUIDv4,
    })
}

// Include CSRF token in all forms
templ FormWithCSRF(schema *FormSchema, csrfToken string) {
    <form hx-post={ schema.Action }>
        <input type="hidden" name="_csrf" value={ csrfToken }/>
        @RenderFormFields(schema.Fields)
        @RenderFormActions(schema.Actions)
    </form>
}
```

### Content Security Policy

```go
func CSPMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        c.Set("Content-Security-Policy", 
            "default-src 'self'; "+
            "script-src 'self' 'unsafe-inline'; "+
            "style-src 'self' 'unsafe-inline'; "+
            "img-src 'self' data: https:; "+
            "connect-src 'self'; "+
            "frame-ancestors 'none'; "+
            "base-uri 'self'; "+
            "form-action 'self'")
        return c.Next()
    }
}
```

---

## Error Handling Architecture

### Graceful Degradation

```go
type ErrorBoundary struct {
    fallbackComponent TemplComponent
    logger           *Logger
    errorReporter    *ErrorReporter
}

func (e *ErrorBoundary) WrapComponent(
    component TemplComponent,
    props ComponentProps,
) templ.Component {
    return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
        defer func() {
            if r := recover(); r != nil {
                e.logger.Error("Component panic", 
                    "error", r,
                    "component", component,
                    "props", props)
                e.errorReporter.Report(ctx, fmt.Errorf("panic: %v", r))
                
                // Render fallback
                e.fallbackComponent.Render(ctx, ComponentProps{
                    Text: "Content temporarily unavailable",
                    Type: "error",
                }).Render(ctx, w)
            }
        }()
        
        return component.Render(ctx, props).Render(ctx, w)
    })
}
```

### Error Types and Handling

```go
type SchemaError struct {
    Type    ErrorType
    Message string
    Schema  *Schema
    Context map[string]interface{}
}

type ErrorType int

const (
    ValidationError ErrorType = iota
    RenderingError
    ComponentError
    SecurityError
    SystemError
)

func HandleSchemaError(err error, c *fiber.Ctx) error {
    var schemaErr *SchemaError
    if errors.As(err, &schemaErr) {
        switch schemaErr.Type {
        case ValidationError:
            return c.Status(400).JSON(fiber.Map{
                "error": "validation_failed",
                "message": schemaErr.Message,
                "details": schemaErr.Context,
            })
        case SecurityError:
            return c.Status(403).JSON(fiber.Map{
                "error": "security_violation",
                "message": "Access denied",
            })
        default:
            return c.Status(500).JSON(fiber.Map{
                "error": "internal_error",
                "message": "An unexpected error occurred",
            })
        }
    }
    
    return c.Status(500).JSON(fiber.Map{
        "error": "unknown_error",
        "message": "An unknown error occurred",
    })
}
```

---

## Integration Patterns

### HTMX Integration

Schema-driven HTMX patterns for common interactions:

```go
// Action schema defines HTMX behavior
type AjaxActionSchema struct {
    Type       string `json:"type"`       // "ajax"
    Method     string `json:"method"`     // "GET", "POST", etc.
    URL        string `json:"url"`        // Target endpoint
    Target     string `json:"target"`     // CSS selector for result
    Swap       string `json:"swap"`       // "innerHTML", "outerHTML", etc.
    Trigger    string `json:"trigger"`    // "click", "change", etc.
    Confirm    string `json:"confirm"`    // Confirmation message
    Include    string `json:"include"`    // Additional data to include
}

templ AjaxAction(action *AjaxActionSchema) {
    <button 
        type="button"
        hx-get={ action.URL }
        hx-target={ action.Target }
        hx-swap={ action.Swap }
        hx-trigger={ action.Trigger }
        if action.Confirm != "" {
            hx-confirm={ action.Confirm }
        }
        if action.Include != "" {
            hx-include={ action.Include }
        }>
        { action.Label }
    </button>
}
```

### Alpine.js Integration

Client-side reactivity through schema configuration:

```go
type AlpineDirectives struct {
    XData   string `json:"x-data"`   // Alpine.js data
    XShow   string `json:"x-show"`   // Show/hide logic
    XBind   string `json:"x-bind"`   // Attribute binding
    XOn     string `json:"x-on"`     // Event handlers
}

templ ComponentWithAlpine(schema *Schema, alpine *AlpineDirectives) {
    <div 
        if alpine.XData != "" {
            x-data={ alpine.XData }
        }
        if alpine.XShow != "" {
            x-show={ alpine.XShow }
        }>
        @RenderSchemaContent(schema)
    </div>
}
```

---

This architectural foundation provides the blueprint for building scalable, secure, and maintainable schema-driven UIs. The next sections will detail specific implementation patterns and production considerations.