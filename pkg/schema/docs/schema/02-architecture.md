# Architecture Overview

## Three-Layer Architecture

```
┌─────────────────────────────────────┐
│  Layer 1: Schema Definition         │
│  JSON Schema → Go Struct            │
└────────────┬────────────────────────┘
             ↓
┌─────────────────────────────────────┐
│  Layer 2: Processing                │
│  Parse → Validate → Enrich → Cache  │
└────────────┬────────────────────────┘
             ↓
┌─────────────────────────────────────┐
│  Layer 3: Rendering                 │
│  Schema → templ → HTML              │
└─────────────────────────────────────┘
```

## System Components

```
┌────────────────────────────────────────────────┐
│              Browser (Frontend)                 │
│  ┌──────────┐  ┌──────────┐  ┌──────────────┐ │
│  │  HTML5   │  │  HTMX    │  │  Alpine.js   │ │
│  └──────────┘  └──────────┘  └──────────────┘ │
└────────────────────┬───────────────────────────┘
                     │ HTTPS (HTML over wire)
┌────────────────────▼───────────────────────────┐
│            Go Server (Fiber v2)                 │
│                                                 │
│  ┌──────────────────────────────────────────┐  │
│  │  Fiber Handlers + Middleware             │  │
│  │  ├─ Auth                                 │  │
│  │  ├─ Tenant Isolation                     │  │
│  │  └─ Rate Limiting                        │  │
│  └────────────┬─────────────────────────────┘  │
│               │                                 │
│  ┌────────────▼─────────────────────────────┐  │
│  │  Schema Processing                       │  │
│  │  ├─ Registry (storage + cache)          │  │
│  │  ├─ Parser (JSON → Go)                   │  │
│  │  ├─ Validator (rules)                    │  │
│  │  └─ Enricher (permissions)               │  │
│  └────────────┬─────────────────────────────┘  │
│               │                                 │
│  ┌────────────▼─────────────────────────────┐  │
│  │  Rendering (templ)                       │  │
│  │  ├─ Form components                      │  │
│  │  ├─ Field components                     │  │
│  │  └─ Layout components                    │  │
│  └────────────┬─────────────────────────────┘  │
│               │                                 │
│               ▼ HTML Output                     │
└─────────────────────────────────────────────────┘
                     │
┌────────────────────▼───────────────────────────┐
│              Storage Layer                      │
│  ┌──────────┐  ┌──────────┐  ┌──────────────┐ │
│  │PostgreSQL│  │  Redis   │  │ S3/Filesystem│ │
│  └──────────┘  └──────────┘  └──────────────┘ │
└─────────────────────────────────────────────────┘
```

## Component Responsibilities

### Frontend

**HTML5**
- Native form validation (instant feedback)
- Semantic markup
- Accessibility (ARIA)

**HTMX**
- Form submission without page reload
- Partial page updates
- HTTP requests triggered by events

**Alpine.js**
- Client-side state
- Show/hide logic
- Reactive data binding

### Backend

**Fiber Handler**
```go
func HandleForm(c *fiber.Ctx) error {
    // 1. Get user from context
    user := GetUserFromContext(c)
    
    // 2. Load schema
    schema, err := registry.Get(c.Context(), "user-form")
    if err != nil {
        return c.Status(404).SendString("Schema not found")
    }
    
    // 3. Enrich with permissions
    enriched, err := enricher.Enrich(c.Context(), schema, user)
    
    // 4. Render
    return views.FormPage(enriched, nil).Render(c.Context(), c.Response().BodyWriter())
}
```

**Registry**
- Loads schemas from storage (PostgreSQL, Redis, files, S3)
- Caches parsed schemas
- Validates schema structure
- Version management

**Parser**
- Converts JSON to Go structs
- Validates JSON structure
- Handles errors
- Sets defaults

**Validator**
- HTML5 validation rules
- Server-side business rules
- Cross-field validation
- Integration with condition package

**Enricher**
- Adds runtime permissions
- Populates default values
- Applies tenant customization
- Injects user context

**templ Renderer**
- Type-safe template rendering
- Maps field types to components
- Generates HTML
- Injects HTMX/Alpine.js attributes

## Data Flow

### Display Form (GET)

```
1. Browser → GET /forms/user-registration
             ↓
2. Fiber Handler
   ├─ Auth middleware (verify user)
   ├─ Tenant middleware (isolate tenant)
   └─ Handler function
             ↓
3. Registry.Get("user-registration")
   ├─ Check Redis cache
   ├─ If miss: load from PostgreSQL
   ├─ Parse JSON to Schema struct
   └─ Return *Schema
             ↓
4. Enricher.Enrich(schema, user)
   ├─ Get user permissions
   ├─ Apply tenant customization
   ├─ Set field.Runtime.Visible
   └─ Return enriched Schema
             ↓
5. Renderer (templ)
   ├─ Loop through fields
   ├─ Render components based on type
   ├─ Add HTMX attributes
   └─ Generate HTML
             ↓
6. Response → HTML (50KB, ~200ms)
             ↓
7. Browser displays form immediately
```

### Submit Form (POST)

```
1. Browser → Fill form → Click submit
             ↓
2. HTML5 validation (instant client-side)
             ↓
3. HTMX → POST /api/users
             ↓
4. Fiber Handler
   ├─ Auth middleware
   ├─ Parse request body
   └─ Handler function
             ↓
5. Registry.Get("user-registration")
   (get schema for validation)
             ↓
6. Validator.ValidateData(schema, data)
   ├─ Check required fields
   ├─ Check field types
   ├─ Check constraints (min, max, pattern)
   ├─ Check uniqueness (database query)
   └─ Run business rules (condition package)
             ↓
7. If valid:
   ├─ Save to PostgreSQL
   ├─ Return success HTML fragment
   └─ HTMX swaps into #result div
   
   If invalid:
   ├─ Return error HTML fragment
   └─ HTMX displays errors
```

## Package Structure

```
github.com/niiniyare/erp/
│
├── pkg/
│   ├── condition/          # Business rules engine
│   └── schema/             # Schema system
│       ├── schema.go       # Core Schema struct
│       ├── field.go        # Field struct
│       ├── types.go        # Enums (SchemaType, FieldType)
│       ├── validation.go   # Validation structures
│       ├── runtime.go      # Runtime structs
│       │
│       ├── parse/          # Parser
│       │   └── parser.go
│       │
│       ├── validate/       # Validator
│       │   └── validator.go
│       │
│       ├── enrich/         # Enricher
│       │   └── enricher.go
│       │
│       └── registry/       # Registry
│           ├── registry.go     # Interface
│           ├── postgres.go     # PostgreSQL impl
│           ├── redis.go        # Redis impl
│           ├── filesystem.go   # File impl
│           └── s3.go          # S3/Minio impl
│
├── views/                  # templ components
│   ├── form.templ
│   └── fields/
│       ├── text.templ
│       ├── select.templ
│       └── ...
│
└── internal/
    └── handlers/           # Fiber handlers
        ├── form.go
        └── submit.go
```

## Storage Interface

The registry uses an interface to support multiple storage backends:

```go
type Storage interface {
    // Get retrieves a schema by ID
    Get(ctx context.Context, id string) ([]byte, error)
    
    // Set stores a schema
    Set(ctx context.Context, id string, data []byte) error
    
    // Delete removes a schema
    Delete(ctx context.Context, id string) error
    
    // List returns all schema IDs
    List(ctx context.Context) ([]string, error)
    
    // Exists checks if schema exists
    Exists(ctx context.Context, id string) (bool, error)
}
```

**Implementations:**
- `PostgresStorage` - ACID compliant, primary storage
- `RedisStorage` - Fast cache, optional TTL
- `FilesystemStorage` - Simple, good for development
- `S3Storage` - Cloud object storage (S3, Minio)

See [09-storage-interface.md](09-storage-interface.md) for details.

## Middleware Stack

```go
app := fiber.New()

// Global middleware
app.Use(logger.New())
app.Use(recover.New())

// Schema routes
schemas := app.Group("/schemas")
schemas.Use(authMiddleware)      // Verify JWT
schemas.Use(tenantMiddleware)    // Extract tenant_id
schemas.Use(rateLimitMiddleware) // Rate limiting

// Handlers
schemas.Get("/:id/new", handleFormNew)
schemas.Post("/:id", handleFormSubmit)
schemas.Get("/:id/:recordId", handleFormEdit)
schemas.Put("/:id/:recordId", handleFormUpdate)
```

## Request Context

Middleware adds data to Fiber context:

```go
type RequestContext struct {
    UserID      string
    TenantID    string
    Permissions []string
    Roles       []string
}

// In middleware
func tenantMiddleware(c *fiber.Ctx) error {
    tenantID := extractTenantID(c)
    c.Locals("tenant_id", tenantID)
    return c.Next()
}

// In handler
func handleForm(c *fiber.Ctx) error {
    tenantID := c.Locals("tenant_id").(string)
    // Use tenantID for isolation
}
```

## Caching Strategy

```
Request → Registry.Get(id)
          ├─ Check memory cache (Go map)
          │  └─ If found: return (1ms)
          │
          ├─ Check Redis cache
          │  └─ If found: parse + return (5ms)
          │
          └─ Load from PostgreSQL
             ├─ Parse JSON
             ├─ Cache in Redis (TTL: 1 hour)
             ├─ Cache in memory (TTL: 5 min)
             └─ Return (50ms)
```

## Security Layers

1. **Transport**: HTTPS only
2. **Authentication**: JWT tokens (middleware)
3. **Authorization**: Permission checks (enricher)
4. **Tenant Isolation**: Row-level security (PostgreSQL)
5. **Validation**: Client (HTML5) + Server (Go)
6. **Rate Limiting**: Per user/IP (middleware)
7. **CSRF**: Token validation (optional)

## Key Design Decisions

### 1. Why Fiber?
- **Fast**: One of the fastest Go frameworks
- **Express-like**: Familiar API for many developers
- **Middleware**: Rich ecosystem
- **Low memory**: Efficient for high-traffic

### 2. Why templ?
- **Type-safe**: Compile-time checks
- **Fast**: Generates Go code
- **Simple**: Just Go + HTML
- **No runtime**: Zero overhead

### 3. Why Storage Interface?
- **Flexibility**: Choose backend per environment
- **Testing**: Easy to mock
- **Migration**: Switch storage without code changes
- **Scalability**: Start with files, move to PostgreSQL + Redis

### 4. Why Server-Side Rendering?
- **Performance**: Instant first paint
- **SEO**: Content in HTML
- **Simplicity**: No complex build tools
- **Resilience**: Works without JavaScript

## Performance Characteristics

**With caching:**
- Schema load: 1-5ms (memory/Redis)
- Enrichment: 2-10ms (permission checks)
- Rendering: 5-20ms (templ)
- **Total: 10-35ms**

**Without caching:**
- Schema load: 50-200ms (PostgreSQL + parse)
- Enrichment: 2-10ms
- Rendering: 5-20ms
- **Total: 60-230ms**

**Recommendation:** Use Redis for production.

## Next Steps

- [03-quick-start.md](03-quick-start.md) - Build a form
- [04-schema-structure.md](04-schema-structure.md) - Schema reference
- [09-storage-interface.md](09-storage-interface.md) - Storage backends
- [10-registry.md](10-registry.md) - Registry implementation

---

[← Back to Introduction](01-introduction.md) | [Next: Quick Start →](03-quick-start.md)