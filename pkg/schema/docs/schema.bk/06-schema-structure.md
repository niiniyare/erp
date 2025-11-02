## Section 6: Schema Structure

### 🎯 Learning Objectives

By the end of this section, you will understand:
- The complete Schema struct definition
- Every field's purpose and usage
- Required vs optional fields
- How to structure valid schemas
- Schema validation rules

### 6.1 Complete Schema Struct Definition

Here is the actual Schema struct from your codebase with detailed documentation:

```go
// Schema is the main entry point - defines a complete form, page, or UI component.
// All UI elements are configured through this structure via JSON.
type Schema struct {
    // ═══════════════════════════════════════════════════════
    // IDENTITY - Required fields that identify the schema
    // ═══════════════════════════════════════════════════════
    
    // ID: Unique identifier (kebab-case recommended)
    // Examples: "user-registration", "invoice-form", "customer-list"
    // Must be: 1-100 characters, unique across system
    ID string `json:"id" validate:"required,min=1,max=100"`
    
    // Type: What kind of UI element this schema represents
    // Values: "form", "component", "layout", "workflow", "theme", "page"
    // Most common: "form"
    Type Type `json:"type" validate:"required"`
    
    // Version: Semantic versioning (MAJOR.MINOR.PATCH)
    // Used for schema migration and compatibility
    // Example: "1.2.0"
    Version string `json:"version,omitempty" validate:"semver"`
    
    // Title: Human-readable title shown to users
    // Examples: "Create User", "Invoice Entry", "Contact Form"
    // Must be: 1-200 characters
    Title string `json:"title" validate:"required,min=1,max=200"`
    
    // Description: Optional longer description
    // Shows below title, provides context to users
    // Max: 1000 characters
    Description string `json:"description,omitempty" validate:"max=1000"`
    
    // ═══════════════════════════════════════════════════════
    // CORE UI STRUCTURE - The actual form/page content
    // ═══════════════════════════════════════════════════════
    
    // Config: HTTP/API configuration for form submission
    // Defines: action URL, method, encoding, timeout, headers
    Config *Config `json:"config,omitempty"`
    
    // Layout: Visual layout structure
    // Defines: grid columns, tabs, steps, sections, responsive behavior
    Layout *Layout `json:"layout,omitempty"`
    
    // Fields: The actual form fields (inputs, selects, etc.)
    // This is the main content - array of Field structs
    // Each field is validated with "dive" (validates each element)
    Fields []Field `json:"fields,omitempty" validate:"dive"`
    
    // Actions: Buttons and actions (submit, reset, custom)
    // Usually contains at least a submit button
    Actions []Action `json:"actions,omitempty" validate:"dive"`
    
    // ═══════════════════════════════════════════════════════
    // ENTERPRISE FEATURES - Advanced functionality
    // ═══════════════════════════════════════════════════════
    
    // Security: CSRF protection, rate limiting, encryption
    Security *Security `json:"security,omitempty"`
    
    // Tenant: Multi-tenancy configuration
    // Defines: tenant field, isolation strategy
    Tenant *Tenant `json:"tenant,omitempty"`
    
    // Workflow: Approval workflows and state machines
    // Defines: steps, transitions, approvers
    Workflow *Workflow `json:"workflow,omitempty"`
    
    // Validation: Cross-field validation rules
    // Defines: rules that depend on multiple fields
    Validation *Validation `json:"validation,omitempty"`
    
    // Events: Lifecycle event handlers
    // Defines: onCreate, onUpdate, onDelete, onSubmit, etc.
    Events *Events `json:"events,omitempty"`
    
    // I18n: Internationalization support
    // Defines: locales, translations
    I18n *I18n `json:"i18n,omitempty"`
    
    // ═══════════════════════════════════════════════════════
    // FRAMEWORK INTEGRATION - HTMX and Alpine.js
    // ═══════════════════════════════════════════════════════
    
    // HTMX: HTMX configuration for this schema
    // Defines: post URL, target, swap strategy
    HTMX *HTMX `json:"htmx,omitempty"`
    
    // Alpine: Alpine.js configuration
    // Defines: x-data, initial state
    Alpine *Alpine `json:"alpine,omitempty"`
    
    // ═══════════════════════════════════════════════════════
    // ORGANIZATION - Categorization and discovery
    // ═══════════════════════════════════════════════════════
    
    // Meta: Creation/update metadata
    // Contains: createdAt, updatedAt, createdBy, updatedBy
    Meta *Meta `json:"meta,omitempty"`
    
    // Tags: Search tags for organization
    // Examples: ["user", "admin", "public"]
    // Each tag must be alphanumeric
    Tags []string `json:"tags,omitempty" validate:"dive,alphanum"`
    
    // Category: Schema category (single value)
    // Examples: "users", "invoices", "reports"
    // Must be alphanumeric
    Category string `json:"category,omitempty" validate:"alphanum"`
    
    // Module: ERP module name
    // Examples: "hr", "accounting", "inventory"
    // Must be alphanumeric
    Module string `json:"module,omitempty" validate:"alphanum"`
    
    // ═══════════════════════════════════════════════════════
    // RUNTIME STATE - Managed by system, not in JSON
    // ═══════════════════════════════════════════════════════
    
    // State: Current form state (values, errors, touched)
    // Managed by frontend (Alpine.js), not stored
    State *State `json:"state,omitempty"`
    
    // Context: Request context (user, tenant, permissions)
    // Injected by middleware, used by enricher
    Context *Context `json:"context,omitempty"`
}
```

### 6.2 Field-by-Field Reference

#### Identity Fields

**ID Field:**
```go
ID string `json:"id" validate:"required,min=1,max=100"`
```

**Purpose:** Unique identifier for the schema across the entire system.

**Rules:**
- ✅ Must be unique (enforced by registry)
- ✅ Recommended format: kebab-case
- ✅ Length: 1-100 characters
- ✅ No special characters except hyphen
- ❌ Not a UUID (human-readable)

**Examples:**
```json
// ✅ GOOD
"id": "user-registration-form"
"id": "invoice-entry"
"id": "customer-list-view"
"id": "employee-profile"

// ❌ BAD
"id": "UserRegistrationForm"  // Not kebab-case
"id": "user_registration"     // Underscores (use hyphen)
"id": "a1b2c3d4-e5f6-..."     // UUID (not human-readable)
"id": "form"                  // Too generic
```

**Type Field:**
```go
Type Type `json:"type" validate:"required"`
```

**Purpose:** Defines what kind of UI element this schema represents.

**Values:**
```go
const (
    TypeForm      Type = "form"      // Complete form with fields and submission
    TypeComponent Type = "component" // Reusable UI component
    TypeLayout    Type = "layout"    // Layout container only
    TypeWorkflow  Type = "workflow"  // Multi-step workflow
    TypeTheme     Type = "theme"     // Theme/styling configuration
    TypePage      Type = "page"      // Full page layout
)
```

**Usage:**
```json
// Most common: form
{"type": "form"}

// For reusable components
{"type": "component"}

// For multi-step wizards
{"type": "workflow"}
```

**Version Field:**
```go
Version string `json:"version,omitempty" validate:"semver"`
```

**Purpose:** Tracks schema versions for compatibility and migration.

**Format:** Semantic versioning (MAJOR.MINOR.PATCH)

**Rules:**
- MAJOR: Breaking changes (incompatible)
- MINOR: New features (backward compatible)
- PATCH: Bug fixes (backward compatible)

**Examples:**
```json
"version": "1.0.0"  // Initial release
"version": "1.1.0"  // Added new field (compatible)
"version": "2.0.0"  // Removed field (breaking)
"version": "1.0.1"  // Fixed validation bug
```

**Title Field:**
```go
Title string `json:"title" validate:"required,min=1,max=200"`
```

**Purpose:** Human-readable title displayed to users.

**Rules:**
- ✅ Required
- ✅ Length: 1-200 characters
- ✅ Shown at top of form/page

**Examples:**
```json
"title": "Create User Account"
"title": "Invoice Entry Form"
"title": "Customer Contact Information"
```

**Description Field:**
```go
Description string `json:"description,omitempty" validate:"max=1000"`
```

**Purpose:** Optional longer description providing context.

**Rules:**
- ✅ Optional
- ✅ Max: 1000 characters
- ✅ Shown below title

**Examples:**
```json
"description": "Complete this form to create a new user account. Required fields are marked with an asterisk (*)"
"description": "Enter invoice details. Tax will be calculated automatically based on the customer's location."
```

#### Core Structure Fields

**Config Field:**
```go
Config *Config `json:"config,omitempty"`
```

**Purpose:** HTTP/API configuration for form submission.

**Structure:**
```go
type Config struct {
    Action   string            `json:"action,omitempty" validate:"url"`
    Method   string            `json:"method,omitempty" validate:"oneof=GET POST PUT PATCH DELETE"`
    Target   string            `json:"target,omitempty" validate:"html_id"`
    Encoding string            `json:"encoding,omitempty" validate:"oneof=application/json multipart/form-data application/x-www-form-urlencoded"`
    Timeout  int               `json:"timeout,omitempty" validate:"min=0,max=300000"` // milliseconds
    Headers  map[string]string `json:"headers,omitempty"`
    Params   map[string]string `json:"params,omitempty"`
    Cache    bool              `json:"cache,omitempty"`
    CacheTTL int               `json:"cacheTTL,omitempty" validate:"min=0"` // seconds
}
```

**Example:**
```json
"config": {
  "action": "/api/v1/users",
  "method": "POST",
  "encoding": "application/json",
  "timeout": 30000,
  "headers": {
    "X-API-Version": "1.0"
  }
}
```

**Layout Field:**
```go
Layout *Layout `json:"layout,omitempty"`
```

**Purpose:** Defines visual arrangement of fields.

**See Section 9 for complete Layout documentation.**

**Fields Field:**
```go
Fields []Field `json:"fields,omitempty" validate:"dive"`
```

**Purpose:** The actual form fields - the main content.

**Rules:**
- ✅ Array of Field structs
- ✅ Each field validated with "dive"
- ✅ Can be empty for layout-only schemas
- ✅ Order matters (display order)

**See Section 7 for complete Field documentation.**

**Actions Field:**
```go
Actions []Action `json:"actions,omitempty" validate:"dive"`
```

**Purpose:** Buttons and actions (submit, reset, custom).

**Structure:**
```go
type Action struct {
    ID      string     `json:"id" validate:"required"`
    Type    ActionType `json:"type" validate:"required"`
    Text    string     `json:"text" validate:"required"`
    Variant string     `json:"variant,omitempty"` // primary, secondary, danger
    Size    string     `json:"size,omitempty"`    // sm, md, lg
    Icon    string     `json:"icon,omitempty"`
    // ... more fields
}
```

**Example:**
```json
"actions": [
  {
    "id": "submit",
    "type": "submit",
    "text": "Save",
    "variant": "primary",
    "size": "md"
  },
  {
    "id": "reset",
    "type": "reset",
    "text": "Clear",
    "variant": "secondary"
  }
]
```

#### Enterprise Features Fields

**Security Field:**
```go
Security *Security `json:"security,omitempty"`
```

**Purpose:** Security configuration (CSRF, rate limiting, encryption).

**Structure:**
```go
type Security struct {
    CSRF       *CSRF       `json:"csrf,omitempty"`
    RateLimit  *RateLimit  `json:"rateLimit,omitempty"`
    Encryption bool        `json:"encryption,omitempty"`
}

type CSRF struct {
    Enabled    bool   `json:"enabled"`
    FieldName  string `json:"fieldName"`
    HeaderName string `json:"headerName"`
}

type RateLimit struct {
    Enabled     bool   `json:"enabled"`
    MaxRequests int    `json:"maxRequests"`
    WindowSec   int64  `json:"windowSec"`
    ByUser      bool   `json:"byUser"`
    ByIP        bool   `json:"byIP"`
}
```

**Example:**
```json
"security": {
  "csrf": {
    "enabled": true,
    "fieldName": "_csrf",
    "headerName": "X-CSRF-Token"
  },
  "rateLimit": {
    "enabled": true,
    "maxRequests": 100,
    "windowSec": 3600,
    "byUser": true
  }
}
```

**Tenant Field:**
```go
Tenant *Tenant `json:"tenant,omitempty"`
```

**Purpose:** Multi-tenancy configuration.

**Structure:**
```go
type Tenant struct {
    Enabled   bool   `json:"enabled"`
    Field     string `json:"field"`     // Field name containing tenant ID
    Isolation string `json:"isolation"` // "strict", "shared"
}
```

**Example:**
```json
"tenant": {
  "enabled": true,
  "field": "tenant_id",
  "isolation": "strict"
}
```

**Workflow Field:**
```go
Workflow *Workflow `json:"workflow,omitempty"`
```

**Purpose:** Approval workflows and state machines.

**Note:** Schema represents workflows but doesn't execute them. Integration with workflow engines (like Temporal) is external.

**Structure:**
```go
type Workflow struct {
    Enabled     bool            `json:"enabled"`
    Type        string          `json:"type"` // "linear", "branching"
    Steps       []WorkflowStep  `json:"steps"`
    Transitions []Transition    `json:"transitions"`
}
```

**Validation Field:**
```go
Validation *Validation `json:"validation,omitempty"`
```

**Purpose:** Cross-field validation rules (rules involving multiple fields).

**Structure:**
```go
type Validation struct {
    Rules []CrossFieldRule `json:"rules"`
}

type CrossFieldRule struct {
    ID          string                   `json:"id"`
    Description string                   `json:"description"`
    Condition   condition.ConditionGroup `json:"condition"`
    Message     string                   `json:"message"`
}
```

**Example:**
```json
"validation": {
  "rules": [
    {
      "id": "end-after-start",
      "description": "End date must be after start date",
      "condition": {
        "conjunction": "and",
        "rules": [
          {
            "left": {"type": "field", "field": "end_date"},
            "op": "greater",
            "right": {"type": "field", "field": "start_date"}
          }
        ]
      },
      "message": "End date must be after start date"
    }
  ]
}
```

**Events Field:**
```go
Events *Events `json:"events,omitempty"`
```

**Purpose:** Lifecycle event handlers.

**Structure:**
```go
type Events struct {
    OnCreate  []EventHandler `json:"onCreate,omitempty"`
    OnUpdate  []EventHandler `json:"onUpdate,omitempty"`
    OnDelete  []EventHandler `json:"onDelete,omitempty"`
    OnSubmit  []EventHandler `json:"onSubmit,omitempty"`
    OnSuccess []EventHandler `json:"onSuccess,omitempty"`
    OnError   []EventHandler `json:"onError,omitempty"`
}

type EventHandler struct {
    Type   string         `json:"type"`   // "webhook", "function", "log"
    Config map[string]any `json:"config"`
}
```

**Example:**
```json
"events": {
  "onSubmit": [
    {
      "type": "webhook",
      "config": {
        "url": "https://api.example.com/notify",
        "method": "POST"
      }
    }
  ]
}
```

**I18n Field:**
```go
I18n *I18n `json:"i18n,omitempty"`
```

**Purpose:** Internationalization support.

**Structure:**
```go
type I18n struct {
    Enabled          bool              `json:"enabled"`
    DefaultLocale    string            `json:"defaultLocale"`
    SupportedLocales []string          `json:"supportedLocales"`
    Translations     map[string]Locale `json:"translations"`
}

type Locale struct {
    Title       string            `json:"title,omitempty"`
    Description string            `json:"description,omitempty"`
    Fields      map[string]string `json:"fields,omitempty"`
}
```

**Example:**
```json
"i18n": {
  "enabled": true,
  "defaultLocale": "en",
  "supportedLocales": ["en", "es", "fr"],
  "translations": {
    "es": {
      "title": "Crear Usuario",
      "fields": {
        "email": "Correo Electrónico",
        "password": "Contraseña"
      }
    }
  }
}
```

#### Framework Integration Fields

**HTMX Field:**
```go
HTMX *HTMX `json:"htmx,omitempty"`
```

**Purpose:** HTMX configuration for this schema.

**Structure:**
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
"htmx": {
  "enabled": true,
  "post": "/api/users",
  "target": "#result",
  "swap": "innerHTML",
  "indicator": "#spinner"
}
```

**Alpine Field:**
```go
Alpine *Alpine `json:"alpine,omitempty"`
```

**Purpose:** Alpine.js configuration.

**Structure:**
```go
type Alpine struct {
    Enabled bool   `json:"enabled"`
    XData   string `json:"xData,omitempty"`   // Initial data
    XInit   string `json:"xInit,omitempty"`   // Initialization code
}
```

**Example:**
```json
"alpine": {
  "enabled": true,
  "xData": "{ loading: false, errors: {} }",
  "xInit": "console.log('Form initialized')"
}
```

#### Organization Fields

**Meta Field:**
```go
Meta *Meta `json:"meta,omitempty"`
```

**Purpose:** Creation and update metadata.

**Structure:**
```go
type Meta struct {
    CreatedAt  time.Time      `json:"createdAt"`
    UpdatedAt  time.Time      `json:"updatedAt"`
    CreatedBy  string         `json:"createdBy,omitempty"`
    UpdatedBy  string         `json:"updatedBy,omitempty"`
    CustomData map[string]any `json:"customData,omitempty"`
}
```

**Example:**
```json
"meta": {
  "createdAt": "2024-01-15T10:30:00Z",
  "updatedAt": "2024-01-20T14:45:00Z",
  "createdBy": "user-123",
  "updatedBy": "user-456"
}
```

**Tags Field:**
```go
Tags []string `json:"tags,omitempty" validate:"dive,alphanum"`
```

**Purpose:** Search tags for organization.

**Rules:**
- ✅ Each tag must be alphanumeric
- ✅ No spaces or special characters
- ✅ Used for search and filtering

**Example:**
```json
"tags": ["user", "admin", "public", "v2"]
```

**Category Field:**
```go
Category string `json:"category,omitempty" validate:"alphanum"`
```

**Purpose:** Single category classification.

**Example:**
```json
"category": "users"
```

**Module Field:**
```go
Module string `json:"module,omitempty" validate:"alphanum"`
```

**Purpose:** ERP module name.

**Example:**
```json
"module": "hr"
```

#### Runtime Fields

**State Field:**
```go
State *State `json:"state,omitempty"`
```

**Purpose:** Current form state (managed by frontend).

**Structure:**
```go
type State struct {
    Values      map[string]any    `json:"values,omitempty"`
    Errors      map[string]string `json:"errors,omitempty"`
    Touched     map[string]bool   `json:"touched,omitempty"`
    Dirty       map[string]bool   `json:"dirty,omitempty"`
    Valid       bool              `json:"valid,omitempty"`
    Submitting  bool              `json:"submitting,omitempty"`
    SubmitCount int               `json:"submitCount,omitempty"`
    LastUpdated time.Time         `json:"lastUpdated,omitempty"`
    CurrentStep string            `json:"currentStep,omitempty"`
    CurrentTab  string            `json:"currentTab,omitempty"`
}
```

**Note:** Not stored in database, managed by Alpine.js.

**Context Field:**
```go
Context *Context `json:"context,omitempty"`
```

**Purpose:** Request context (injected by middleware).

**Structure:**
```go
type Context struct {
    UserID      string   `json:"userId,omitempty"`
    TenantID    string   `json:"tenantId,omitempty"`
    SessionID   string   `json:"sessionId,omitempty"`
    RequestID   string   `json:"requestId,omitempty"`
    IP          string   `json:"ip,omitempty"`
    UserAgent   string   `json:"userAgent,omitempty"`
    Locale      string   `json:"locale,omitempty"`
    Timezone    string   `json:"timezone,omitempty"`
    Permissions []string `json:"permissions,omitempty"`
    Roles       []string `json:"roles,omitempty"`
    Environment string   `json:"environment,omitempty"`
    Debug       bool     `json:"debug,omitempty"`
    Data        map[string]any `json:"data,omitempty"`
}
```

**Note:** Injected at runtime, not in stored schema.

### 6.3 Minimal Valid Schema

The absolute minimum required fields:

```json
{
  "id": "minimal-form",
  "type": "form",
  "title": "Minimal Form"
}
```

This is valid but not useful. A practical minimum:

```json
{
  "id": "contact-form",
  "type": "form",
  "title": "Contact Us",
  "fields": [
    {
      "name": "email",
      "type": "email",
      "label": "Email",
      "required": true
    }
  ],
  "actions": [
    {
      "id": "submit",
      "type": "submit",
      "text": "Submit"
    }
  ],
  "config": {
    "action": "/api/contact",
    "method": "POST"
  }
}
```

### 6.4 Complete Example Schema

Here's a comprehensive example using most features:

```json
{
  "id": "user-registration-complete",
  "type": "form",
  "version": "1.0.0",
  "title": "User Registration",
  "description": "Create a new user account with complete profile information",
  
  "config": {
    "action": "/api/users",
    "method": "POST",
    "encoding": "application/json",
    "timeout": 30000
  },
  
  "layout": {
    "type": "grid",
    "columns": 2,
    "gap": "1rem",
    "responsive": true
  },
  
  "fields": [
    {
      "name": "first_name",
      "type": "text",
      "label": "First Name",
      "required": true,
      "validation": {
        "minLength": 2,
        "maxLength": 50
      }
    },
    {
      "name": "last_name",
      "type": "text",
      "label": "Last Name",
      "required": true,
      "validation": {
        "minLength": 2,
        "maxLength": 50
      }
    },
    {
      "name": "email",
      "type": "email",
      "label": "Email Address",
      "required": true,
      "validation": {
        "maxLength": 255,
        "server": {
          "unique": true
        }
      }
    },
    {
      "name": "password",
      "type": "password",
      "label": "Password",
      "required": true,
      "validation": {
        "minLength": 8,
        "pattern": "^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d).+$",
        "messages": {
          "pattern": "Password must contain uppercase, lowercase, and number"
        }
      }
    },
    {
      "name": "role",
      "type": "select",
      "label": "Role",
      "required": true,
      "options": [
        {"value": "user", "label": "User"},
        {"value": "admin", "label": "Administrator"}
      ],
      "requirePermission": "users.assign_role"
    }
  ],
  
  "actions": [
    {
      "id": "submit",
      "type": "submit",
      "text": "Create Account",
      "variant": "primary",
      "size": "md"
    },
    {
      "id": "cancel",
      "type": "button",
      "text": "Cancel",
      "variant": "secondary"
    }
  ],
  
  "security": {
    "csrf": {
      "enabled": true,
      "fieldName": "_csrf"
    },
    "rateLimit": {
      "enabled": true,
      "maxRequests": 10,
      "windowSec": 3600,
      "byIP": true
    }
  },
  
  "tenant": {
    "enabled": true,
    "field": "tenant_id",
    "isolation": "strict"
  },
  
  "htmx": {
    "enabled": true,
    "post": "/api/users",
    "target": "#result",
    "swap": "innerHTML"
  },
  
  "alpine": {
    "enabled": true,
    "xData": "{ submitting: false, success: false }"
  },
  
  "tags": ["user", "registration", "public"],
  "category": "users",
  "module": "auth"
}
```

### 6.5 Schema Validation Rules

The Schema struct has validation rules enforced by go-playground/validator:

```go
// Required fields
ID    string `validate:"required,min=1,max=100"`
Type  Type   `validate:"required"`
Title string `validate:"required,min=1,max=200"`

// Optional with constraints
Description string `validate:"max=1000"`
Version     string `validate:"semver"`
Tags        []string `validate:"dive,alphanum"`
Category    string `validate:"alphanum"`
Module      string `validate:"alphanum"`

// Nested validation
Fields  []Field  `validate:"dive"`  // Validates each field
Actions []Action `validate:"dive"`  // Validates each action
```

**Validation happens:**
1. When parsing JSON → Go struct
2. When calling `schema.Validate()`
3. Before saving to registry

**Common validation errors:**
- Missing required fields (id, type, title)
- ID too long (>100 chars)
- Invalid semantic version
- Duplicate field names
- Invalid field types
- Circular dependencies

### 6.6 Schema Methods

Your Schema struct has these methods:

#### Constructor

```go
func NewSchema(id string, schemaType Type, title string) *Schema
```

Creates new schema with defaults:
- Sets ID, Type, Title
- Initializes Meta (created/updated timestamps)
- Initializes State (empty maps)
- Initializes empty Fields and Actions slices

**Example:**
```go
schema := schema.NewSchema("my-form", schema.TypeForm, "My Form")
```

#### Field Management

```go
func (s *Schema) AddField(field Field)
```

Appends field to Fields slice and updates timestamp.

```go
func (s *Schema) GetField(name string) (*Field, bool)
```

Retrieves field by name, returns field and existence boolean.

```go
func (s *Schema) HasField(name string) bool
```

Checks if field exists.

```go
func (s *Schema) GetVisibleFields(data map[string]any) []Field
```

Returns fields visible based on conditional logic and data.

```go
func (s *Schema) GetRequiredFields(data map[string]any) []Field
```

Returns required fields based on conditional logic and data.

#### Action Management

```go
func (s *Schema) AddAction(action Action)
```

Appends action to Actions slice and updates timestamp.

#### Validation

```go
func (s *Schema) Validate() error
```

Performs comprehensive validation:
- Checks required fields
- Validates all fields
- Validates all actions
- Checks for duplicate field names
- Validates field dependencies exist
- Detects circular dependencies

Returns nil if valid, or error with details.

```go
func (s *Schema) DetectCircularDependencies() error
```

Checks for circular field dependencies using DFS algorithm.

#### Utility

```go
func (s *Schema) Clone() *Schema
```

Creates deep copy of schema.

```go
func (s *Schema) MarshalJSON() ([]byte, error)
```

Custom JSON marshaling that updates timestamp before serializing.

```go
func (s *Schema) UnmarshalJSON(data []byte) error
```

Custom JSON unmarshaling that validates after parsing.

#### Interface Methods

```go
func (s *Schema) GetID() string
func (s *Schema) GetType() string
func (s *Schema) GetMetadata() *Meta
func (s *Schema) GetVersion() string
```

Getter methods for interface compliance.

### 6.7 Common Pitfalls

#### Pitfall 1: Duplicate Field Names

```json
// ❌ ERROR: Duplicate field names
{
  "fields": [
    {"name": "email", "type": "email", "label": "Email"},
    {"name": "email", "type": "text", "label": "Backup Email"}
  ]
}
```

**Solution:** Use unique names
```json
// ✅ CORRECT
{
  "fields": [
    {"name": "email", "type": "email", "label": "Primary Email"},
    {"name": "backup_email", "type": "email", "label": "Backup Email"}
  ]
}
```

#### Pitfall 2: Missing Required Fields

```json
// ❌ ERROR: Missing title
{
  "id": "my-form",
  "type": "form"
  // Missing: "title"
}
```

#### Pitfall 3: Invalid Dependencies

```json
// ❌ ERROR: Depends on non-existent field
{
  "fields": [
    {
      "name": "country",
      "type": "select",
      "dependencies": ["continent"]  // "continent" doesn't exist
    }
  ]
}
```

#### Pitfall 4: Circular Dependencies

```json
// ❌ ERROR: Circular dependency
{
  "fields": [
    {
      "name": "field_a",
      "dependencies": ["field_b"]
    },
    {
      "name": "field_b",
      "dependencies": ["field_a"]  // Circular!
    }
  ]
}
```

#### Pitfall 5: Invalid Type Values

```json
// ❌ ERROR: Invalid type (uppercase)
{
  "id": "my-form",
  "type": "FORM",  // Should be "form" (lowercase)
  "title": "My Form"
}
```

---

