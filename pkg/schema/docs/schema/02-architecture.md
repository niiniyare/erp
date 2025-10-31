## Section 2: Architecture Overview

### 🎯 Learning Objectives

By the end of this section, you will understand:
- The complete system architecture
- How data flows through the system
- The role of each component
- Package structure and dependencies
- Integration points

### 2.1 High-Level Architecture

#### The Big Picture

```
┌────────────────────────────────────────────────────────────┐
│                      Frontend (Browser)                     │
│                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌─────────────────┐  │
│  │    HTML5     │  │    HTMX      │  │   Alpine.js     │  │
│  │  Validation  │  │  (14KB CDN)  │  │   (15KB CDN)    │  │
│  └──────────────┘  └──────────────┘  └─────────────────┘  │
│         │                  │                    │           │
└─────────┼──────────────────┼────────────────────┼───────────┘
          │                  │                    │
          │        HTTP/HTTPS (HTML over wire)    │
          │                  │                    │
┌─────────▼──────────────────▼────────────────────▼───────────┐
│                     Backend (Go Server)                      │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              HTTP Handlers Layer                      │  │
│  │  ├─ Authentication Middleware                         │  │
│  │  ├─ Authorization Middleware                          │  │
│  │  ├─ Tenant Isolation Middleware                       │  │
│  │  └─ Rate Limiting Middleware                          │  │
│  └────────────────────┬─────────────────────────────────┘  │
│                       │                                     │
│  ┌────────────────────▼─────────────────────────────────┐  │
│  │             Schema Processing Layer                   │  │
│  │                                                        │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────────────┐   │  │
│  │  │ Registry │  │  Parser  │  │    Validator     │   │  │
│  │  │ (Cache)  │  │  (JSON)  │  │   (Rules)        │   │  │
│  │  └──────────┘  └──────────┘  └──────────────────┘   │  │
│  │                                                        │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────────────┐   │  │
│  │  │ Enricher │  │Condition │  │     Builder      │   │  │
│  │  │(Runtime) │  │ Package  │  │    (Fluent)      │   │  │
│  │  └──────────┘  └──────────┘  └──────────────────┘   │  │
│  └────────────────────┬─────────────────────────────────┘  │
│                       │                                     │
│  ┌────────────────────▼─────────────────────────────────┐  │
│  │              Rendering Layer (templ)                  │  │
│  │                                                        │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────────────┐   │  │
│  │  │FormPage  │  │  Field   │  │     Layout       │   │  │
│  │  │Component │  │Components│  │   Components     │   │  │
│  │  └──────────┘  └──────────┘  └──────────────────┘   │  │
│  └────────────────────┬─────────────────────────────────┘  │
│                       │                                     │
│                       ▼                                     │
│                   HTML Output                               │
└─────────────────────────────────────────────────────────────┘
          │                  │                    │
          ▼                  ▼                    ▼
┌────────────────────────────────────────────────────────────┐
│                    Data Layer                               │
│                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌─────────────────┐  │
│  │  PostgreSQL  │  │    Redis     │  │   File System   │  │
│  │   Schemas    │  │    Cache     │  │  (JSON files)   │  │
│  │   Users      │  │              │  │                 │  │
│  │   Data       │  │              │  │                 │  │
│  └──────────────┘  └──────────────┘  └─────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 Component Responsibilities

#### Frontend Components

**HTML5**
- Native form validation (instant feedback)
- Semantic markup
- Accessibility (ARIA attributes)
- SEO-friendly structure

**HTMX**
- Form submission without page reload
- Partial page updates (swap HTML fragments)
- HTTP requests triggered by events
- Response handling (success, error)

**Alpine.js**
- Client-side state management
- Show/hide logic
- Custom validation messages
- Reactive data binding

#### Backend Components

**HTTP Handlers**
```go
func HandleForm(w http.ResponseWriter, r *http.Request) {
    // 1. Get user context
    user := GetUserFromContext(r.Context())
    
    // 2. Load schema
    schema, err := registry.Get(r.Context(), "contact-form")
    
    // 3. Enrich with permissions
    enriched, err := enricher.Enrich(r.Context(), schema, user)
    
    // 4. Render
    views.FormPage(enriched, nil).Render(r.Context(), w)
}
```

**Registry**
- Loads schemas from storage (DB, files)
- Caches parsed schemas (Redis, memory)
- Validates schema structure
- Version management

**Parser**
- JSON → Go struct conversion
- Validation of JSON structure
- Error handling
- Default value population

**Validator**
- HTML5 rule validation
- Server-side business rules
- Cross-field validation
- Integration with condition package

**Enricher**
- Adds runtime permissions
- Populates default values
- Applies tenant customization
- Adds user context

**templ Components**
- Type-safe template rendering
- Field type → component mapping
- HTML generation
- HTMX/Alpine.js attribute injection

### 2.3 Data Flow Diagrams

#### Read Flow (Display Form)

```
User                Browser              Server              Database
 │                     │                    │                    │
 ├─ Click link ───────→│                    │                    │
 │                     │                    │                    │
 │                     ├─ GET /form ───────→│                    │
 │                     │                    │                    │
 │                     │                    ├─ Auth check       │
 │                     │                    │   (middleware)     │
 │                     │                    │                    │
 │                     │                    ├─ Get schema ──────→│
 │                     │                    │                    │
 │                     │                    │←─ Return JSON ─────┤
 │                     │                    │                    │
 │                     │                    ├─ Parse JSON        │
 │                     │                    │   (parser)         │
 │                     │                    │                    │
 │                     │                    ├─ Enrich schema     │
 │                     │                    │   (permissions)    │
 │                     │                    │                    │
 │                     │                    ├─ Render HTML       │
 │                     │                    │   (templ)          │
 │                     │                    │                    │
 │                     │←─ HTML (50KB) ─────┤                    │
 │                     │                    │                    │
 │←─ Display form ─────┤                    │                    │
 │                     │                    │                    │
```

#### Write Flow (Submit Form)

```
User                Browser              Server              Database
 │                     │                    │                    │
 ├─ Fill form ────────→│                    │                    │
 │                     │                    │                    │
 ├─ Click submit ─────→│                    │                    │
 │                     │                    │                    │
 │                     ├─ Validate (HTML5) │                    │
 │                     │   (instant)        │                    │
 │                     │                    │                    │
 │                     ├─ POST /api ───────→│                    │
 │                     │   (HTMX)           │                    │
 │                     │                    │                    │
 │                     │                    ├─ Auth check       │
 │                     │                    │                    │
 │                     │                    ├─ Parse body        │
 │                     │                    │                    │
 │                     │                    ├─ Get schema ──────→│
 │                     │                    │                    │
 │                     │                    │←─ Return schema ───┤
 │                     │                    │                    │
 │                     │                    ├─ Validate data     │
 │                     │                    │   - HTML5 rules    │
 │                     │                    │   - Unique check ─→│
 │                     │                    │                    │
 │                     │                    │←─ Exists? ─────────┤
 │                     │                    │                    │
 │                     │                    ├─ Business rules    │
 │                     │                    │   (condition pkg)  │
 │                     │                    │                    │
 │                     │                    ├─ If valid: save ─→│
 │                     │                    │                    │
 │                     │                    │←─ ID returned ─────┤
 │                     │                    │                    │
 │                     │                    ├─ Render success    │
 │                     │                    │   (templ)          │
 │                     │                    │                    │
 │                     │←─ HTML fragment ───┤                    │
 │                     │   (success msg)    │                    │
 │                     │                    │                    │
 │←─ Update UI ────────┤                    │                    │
 │   (HTMX swap)       │                    │                    │
```

### 2.4 Package Structure

#### Go Project Layout

```
github.com/niiniyare/erp/
│
├── pkg/
│   │
│   ├── condition/              # Existing business rules engine
│   │   ├── evaluator.go
│   │   ├── condition.go
│   │   └── ...
│   │
│   └── schema/                 # Schema system (THIS PROJECT)
│       │
│       ├── Core Types (no dependencies)
│       ├── schema.go           # Main Schema struct
│       ├── field.go            # Field struct + types
│       ├── layout.go           # Layout configuration
│       ├── validation.go       # Validation structures
│       ├── action.go           # Button/action definitions
│       ├── types.go            # Enums (SchemaType, FieldType, etc.)
│       ├── runtime.go          # Runtime structs (FieldRuntime, etc.)
│       ├── errors.go           # Error types
│       ├── meta.go             # Metadata structs
│       ├── security.go         # Security configuration
│       ├── events.go           # Event handling
│       ├── i18n.go             # Internationalization
│       ├── htmx.go             # HTMX configuration
│       ├── alpine.go           # Alpine.js configuration
│       ├── mixin.go            # Mixin system
│       ├── builder.go          # Fluent builder
│       │
│       ├── parse/              # Parser (Layer 2)
│       │   ├── parser.go       # Parser interface + impl
│       │   ├── json.go         # JSON parsing
│       │   └── yaml.go         # YAML parsing (optional)
│       │
│       ├── validate/           # Validator (Layer 2)
│       │   ├── validator.go    # Validator interface + impl
│       │   ├── html5.go        # HTML5 rules validation
│       │   ├── server.go       # Server-side validation
│       │   ├── business.go     # Business rules (uses condition pkg)
│       │   └── custom.go       # Custom validator registry
│       │
│       ├── enrich/             # Enricher (Layer 2)
│       │   ├── enricher.go     # Enricher interface + impl
│       │   ├── permissions.go  # Permission enrichment
│       │   ├── defaults.go     # Default values
│       │   └── tenant.go       # Tenant customization
│       │
│       └── registry/           # Registry (Layer 3)
│           ├── registry.go     # Registry interface + impl
│           ├── memory.go       # In-memory storage
│           ├── postgres.go     # PostgreSQL storage
│           ├── file.go         # File-based storage
│           └── cache.go        # Caching layer (Redis)
│
├── views/                      # templ components (Layer 3)
│   ├── form.templ              # Form page renderer
│   ├── fields/                 # Field components
│   │   ├── text.templ
│   │   ├── email.templ
│   │   ├── select.templ
│   │   └── ... (40+ types)
│   ├── layout/                 # Layout components
│   │   ├── grid.templ
│   │   ├── tabs.templ
│   │   └── steps.templ
│   └── partials/               # Reusable partials
│       ├── error.templ
│       ├── success.templ
│       └── loading.templ
│
├── cmd/
│   └── server/
│       └── main.go             # HTTP server entry point
│
├── internal/
│   ├── handlers/               # HTTP handlers
│   │   ├── form.go
│   │   ├── submit.go
│   │   └── ...
│   └── middleware/             # Middleware
│       ├── auth.go
│       ├── tenant.go
│       └── ratelimit.go
│
└── schemas/                    # Example JSON schemas
    ├── contact-form.json
    ├── user-registration.json
    └── ...
```

#### Dependency Flow (No Cycles)

```
┌──────────────────────────────────────────────┐
│       Core Types (pkg/schema/*.go)           │
│       • schema.go, field.go, types.go        │
│       • NO external dependencies             │
│       • Only imports: standard library       │
└────────────────┬─────────────────────────────┘
                 │
                 │ imports
                 ▼
┌──────────────────────────────────────────────┐
│         Sub-packages (Layer 2)               │
│                                              │
│  ┌────────────┐  ┌────────────┐            │
│  │  parse/    │  │ validate/  │            │
│  │            │  │            │            │
│  └────────────┘  └────────────┘            │
│         │              │                    │
│         │              ├─imports─→ condition│
│         │              │                    │
│  ┌────────────┐  ┌────────────┐            │
│  │  enrich/   │  │ registry/  │            │
│  │            │  │            │            │
│  └────────────┘  └────────────┘            │
│         │              │                    │
│         └──────┬───────┘                    │
│                │                            │
│                │ can import parse           │
└────────────────┴─────────────────────────────┘
                 │
                 │ imports
                 ▼
┌──────────────────────────────────────────────┐
│       Views (templ components)               │
│       • Imports schema package               │
│       • Renders schemas to HTML              │
└────────────────┬─────────────────────────────┘
                 │
                 │ used by
                 ▼
┌──────────────────────────────────────────────┐
│       HTTP Handlers                          │
│       • Imports schema, views                │
│       • Orchestrates request flow            │
└──────────────────────────────────────────────┘
```

**Key Rules:**
1. ✅ Core types import NOTHING (except stdlib)
2. ✅ Sub-packages import core types only
3. ✅ Registry can import parse (same layer)
4. ✅ Views import schema package
5. ✅ Handlers import everything
6. ❌ NO circular imports possible

### 2.5 Integration Points

#### Database Integration

**Schema Storage (PostgreSQL):**
```sql
CREATE TABLE schemas (
    id          VARCHAR(100) PRIMARY KEY,
    type        VARCHAR(50) NOT NULL,
    version     VARCHAR(20) NOT NULL,
    title       VARCHAR(200) NOT NULL,
    description TEXT,
    json_schema JSONB NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by  UUID REFERENCES users(id),
    updated_by  UUID REFERENCES users(id),
    tenant_id   UUID REFERENCES tenants(id)
);

CREATE INDEX idx_schemas_type ON schemas(type);
CREATE INDEX idx_schemas_tenant ON schemas(tenant_id);
CREATE INDEX idx_schemas_updated ON schemas(updated_at DESC);
```

**Registry Implementation:**
```go
type PostgresRegistry struct {
    db *sql.DB
}

func (r *PostgresRegistry) Get(ctx context.Context, id string) (*Schema, error) {
    var jsonData []byte
    err := r.db.QueryRowContext(ctx, 
        "SELECT json_schema FROM schemas WHERE id = $1", 
        id,
    ).Scan(&jsonData)
    
    if err != nil {
        return nil, err
    }
    
    parser := parse.NewParser()
    return parser.Parse(jsonData)
}
```

#### Cache Integration (Redis)

**Caching Layer:**
```go
type CachedRegistry struct {
    source Registry  // Underlying registry (DB)
    cache  *redis.Client
    ttl    time.Duration
}

func (r *CachedRegistry) Get(ctx context.Context, id string) (*Schema, error) {
    // Try cache first
    cached, err := r.cache.Get(ctx, "schema:"+id).Bytes()
    if err == nil {
        parser := parse.NewParser()
        return parser.Parse(cached)
    }
    
    // Cache miss - load from source
    schema, err := r.source.Get(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Cache for next time
    jsonData, _ := json.Marshal(schema)
    r.cache.Set(ctx, "schema:"+id, jsonData, r.ttl)
    
    return schema, nil
}
```

#### Authentication Integration

**Middleware:**
```go
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extract token
        token := r.Header.Get("Authorization")
        
        // Validate token
        user, err := validateToken(token)
        if err != nil {
            http.Error(w, "Unauthorized", 401)
            return
        }
        
        // Add user to context
        ctx := context.WithValue(r.Context(), userKey, user)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

#### Authorization Integration

**Permission Check:**
```go
type User struct {
    ID          string
    TenantID    string
    Roles       []string
    Permissions []string
}

func (u *User) HasPermission(permission string) bool {
    for _, p := range u.Permissions {
        if p == permission || p == "*" {
            return true
        }
    }
    return false
}
```

**Enrichment:**
```go
func (e *Enricher) Enrich(ctx context.Context, schema *Schema, user *User) (*Schema, error) {
    for i := range schema.Fields {
        field := &schema.Fields[i]
        
        // Check permission
        if field.RequirePermission != "" {
            hasPermission := user.HasPermission(field.RequirePermission)
            
            field.Runtime = &FieldRuntime{
                Visible:  hasPermission,
                Editable: hasPermission,
                Reason:   "permission_required",
            }
        } else {
            field.Runtime = &FieldRuntime{
                Visible:  true,
                Editable: !field.Readonly,
            }
        }
    }
    
    return schema, nil
}
```

---

