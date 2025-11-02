## Section 4: Core Concepts & Terminology

### 🎯 Learning Objectives

By the end of this section, you will understand:
- Key terminology used throughout this specification
- The relationship between different schema components
- Type system fundamentals
- Runtime vs design-time concepts

### 4.1 Schema Hierarchy

#### The Schema Tree

```
Schema (Root)
├── Identity (id, type, version, title)
├── Configuration (config, layout)
├── Content
│   ├── Fields[] (form inputs)
│   └── Actions[] (buttons)
├── Enterprise Features
│   ├── Security (CSRF, rate limiting)
│   ├── Tenant (multi-tenancy)
│   ├── Workflow (approval flows)
│   ├── Validation (cross-field rules)
│   ├── Events (lifecycle hooks)
│   └── I18n (translations)
├── Framework Integration
│   ├── HTMX (configuration)
│   └── Alpine (bindings)
└── Runtime State
    ├── State (form values, errors)
    └── Context (user, permissions)
```

### 4.2 Key Terms

#### Schema-Level Terms

**Schema**
- The root container defining a complete UI component
- Contains fields, actions, layout, and configuration
- Serialized as JSON, parsed into Go struct
- Example: A contact form, user registration, invoice entry

**Type** (`SchemaType`)
- Defines what kind of UI element the schema represents
- Values: `form`, `component`, `layout`, `workflow`, `page`, `theme`
- Determines how the schema is rendered

**Version**
- Semantic versioning (MAJOR.MINOR.PATCH)
- Used for schema migration and compatibility
- Example: `1.2.3`

#### Field-Level Terms

**Field**
- A single input element or UI component
- Has a type, label, validation rules, etc.
- Example: Email input, dropdown select, date picker

**FieldType**
- Defines the kind of input field
- 40+ types available (text, email, select, etc.)
- Maps to HTML input types or custom components

**Validation**
- Rules that determine if a field value is acceptable
- Two types: HTML5-compatible and server-only
- Applied both client-side (instant feedback) and server-side (security)

**Required**
- Boolean flag indicating field must have a value
- Renders `required` attribute in HTML
- Validated on both client and server

**Readonly**
- Boolean flag indicating value cannot be changed
- User can see but not edit
- Different from `disabled` (readonly values are submitted)

**Hidden**
- Boolean flag to hide field from display
- Still exists in schema, just not visible
- Different from conditional visibility

#### Layout Terms

**Layout**
- Defines visual arrangement of fields
- Types: grid, flex, tabs, steps, sections
- Controls responsive behavior

**Section**
- Logical grouping of fields with a title
- Can be collapsible
- Rendered as a visual block (card, panel)

**Group**
- Visual grouping of fields (like HTML fieldset)
- Can have a border and label
- Smaller than section

**Tab**
- Page within a tabbed interface
- Contains fields, navigable via tabs
- Only one tab visible at a time

**Step**
- Stage in a multi-step wizard
- Sequential navigation (back/next)
- Can validate before proceeding

#### Validation Terms

**HTML5 Validation**
- Native browser validation
- Attributes: `required`, `minlength`, `maxlength`, `pattern`, `min`, `max`
- Instant feedback, no server round-trip
- NOT secure (can be bypassed)

**Server Validation**
- Validation performed on the server
- Cannot be bypassed by user
- Includes business rules, uniqueness checks
- Authoritative source of truth

**Business Rule**
- Complex validation logic using condition package
- Example: "Discount only applies if order > $100 AND user is premium"
- Defined declaratively in JSON
- Evaluated at runtime

**Cross-Field Validation**
- Validation that depends on multiple fields
- Example: "End date must be after start date"
- Can be client-side (simple) or server-side (complex)

#### Conditional Logic Terms

**Conditional**
- Rules that determine when fields show/hide or become required
- Three types: `show`, `hide`, `required`
- Uses condition package for complex logic

**ShowIf**
- Simple client-side expression (Alpine.js)
- Example: `showIf: "needs_shipping === true"`
- Instant, no server round-trip
- For simple UI logic only

**RequirePermission**
- Server-side permission check
- Example: `requirePermission: "hr.view_salary"`
- Backend decides, frontend respects
- For security decisions

#### Runtime Terms

**Enrichment**
- Process of adding runtime data to schema
- Happens on server before rendering
- Adds: permissions, default values, tenant customization

**FieldRuntime**
- Runtime flags added during enrichment
- Fields: `Visible`, `Editable`, `Reason`
- Populated by server, respected by client

**Context**
- Runtime execution environment
- Contains: user ID, tenant ID, permissions, roles
- Injected by middleware, used by enricher

**State**
- Current form state (client-side)
- Fields: `Values`, `Errors`, `Touched`, `Dirty`
- Managed by Alpine.js or similar

#### Action Terms

**Action**
- A button or clickable element
- Types: submit, reset, button, link
- Can trigger events, navigate, or execute functions

**Submit Action**
- Submits form data to server
- Triggers validation before submission
- Can use HTMX for AJAX submission

#### Framework Integration Terms

**HTMX**
- Library for HTML over the wire
- Allows AJAX without JavaScript code
- Attributes: `hx-post`, `hx-get`, `hx-target`, `hx-swap`

**Alpine.js**
- Minimal reactive framework
- Handles client-side state and interactions
- Directives: `x-data`, `x-show`, `x-model`, `x-bind`

**templ**
- Type-safe Go template language
- Generates Go code from `.templ` files
- Provides compile-time safety

#### Storage Terms

**Registry**
- Service that manages schema storage and retrieval
- Can use files, database, or memory
- Often includes caching layer

**Parser**
- Converts JSON to Go structs
- Validates JSON structure
- Returns `*Schema` or error

**Validator**
- Checks schema structure and data validity
- Two types: schema validator, data validator
- Returns list of errors

### 4.3 Type System

#### SchemaType Enum

```go
type Type string

const (
    TypeForm      Type = "form"      // Complete form with submission
    TypeComponent Type = "component" // Reusable UI component
    TypeLayout    Type = "layout"    // Container/layout only
    TypeWorkflow  Type = "workflow"  // Multi-step workflow
    TypeTheme     Type = "theme"     // Styling configuration
    TypePage      Type = "page"      // Full page layout
)
```

#### FieldType Enum (40+ Types)

**Basic Text:**
- `text` - Single-line text input
- `email` - Email with validation
- `password` - Masked password input
- `number` - Numeric input
- `phone` - Phone number
- `url` - URL with validation
- `hidden` - Hidden input (submitted but not shown)

**Date/Time:**
- `date` - Date picker
- `time` - Time picker
- `datetime` - Date and time combined
- `daterange` - Start/end date selection

**Text Content:**
- `textarea` - Multi-line text
- `richtext` - WYSIWYG editor
- `code` - Code editor with syntax highlighting
- `json` - JSON editor with validation

**Selection:**
- `select` - Dropdown list
- `multiselect` - Multiple selection dropdown
- `radio` - Radio buttons
- `checkbox` - Checkboxes
- `treeselect` - Hierarchical selection
- `cascader` - Cascading dropdowns
- `transfer` - Two-list transfer

**Interactive:**
- `switch` - Toggle switch
- `slider` - Range slider
- `rating` - Star rating
- `color` - Color picker

**Files:**
- `file` - Generic file upload
- `image` - Image upload with preview
- `signature` - Signature pad

**Specialized:**
- `currency` - Money input with formatting
- `tags` - Tag input (comma-separated)
- `location` - Address/location picker
- `relation` - Foreign key relationship
- `autocomplete` - Autocomplete search

**Display:**
- `display` - Read-only display
- `divider` - Visual separator
- `html` - Raw HTML content

**Collections:**
- `repeatable` - Repeatable field groups
- `table_repeater` - Table-style repeater

### 4.4 Runtime vs Design-Time

#### Design-Time (Static)

**What it is:**
- Schema definition in JSON
- Field configurations
- Validation rules
- Layout structure

**When it exists:**
- Stored in database or files
- Defined by developers
- Version controlled
- Rarely changes

**Example:**
```json
{
  "id": "user-form",
  "fields": [
    {
      "name": "email",
      "type": "email",
      "required": true
    }
  ]
}
```

#### Runtime (Dynamic)

**What it is:**
- Enriched schema with runtime data
- Permission flags
- User-specific customization
- Tenant overrides
- Computed values

**When it exists:**
- Created during request processing
- Unique per user/tenant
- Temporary (not stored)
- Changes with context

**Example:**
```go
field.Runtime = &FieldRuntime{
    Visible:  user.HasPermission("view_email"),
    Editable: user.HasRole("admin"),
    Reason:   "insufficient_permissions",
}
```

### 4.5 Data Flow Concepts

#### Uni-Directional Data Flow

```
Design-Time Schema (JSON)
    ↓
Parse (JSON → Go struct)
    ↓
Validate (Check structure)
    ↓
Enrich (Add runtime data)
    ↓
Render (Go struct → HTML)
    ↓
Browser (Display HTML)
```

**Never:**
```
Browser → Modify Schema ❌
Client → Change Permissions ❌
JavaScript → Alter Validation ❌
```

**Always:**
```
Server → Decides Everything ✅
Client → Respects Server Decisions ✅
```

#### State Management

**Server State (Source of Truth):**
- Schema definition
- User permissions
- Validation rules
- Business logic

**Client State (Temporary):**
- Current form values
- Validation errors (displayed)
- UI interactions (collapse/expand)
- Loading states

**Synchronization:**
- Client submits → Server validates → Server responds
- No client-side schema manipulation
- No optimistic updates that bypass server

### 4.6 Common Patterns

#### Pattern: CRUD Form

```
List View → Detail View → Edit Form → Save → Back to List
   ↓           ↓            ↓          ↓         ↓
Schema A    Schema B    Schema C    Server   Schema A
```

#### Pattern: Multi-Step Wizard

```
Step 1 → Validate → Step 2 → Validate → Step 3 → Submit
  ↓                   ↓                   ↓
Schema              Schema              Schema
(step 1 fields)    (step 2 fields)    (step 3 fields)
```

#### Pattern: Conditional Fields

```
User selects "Yes" → Field appears
   ↓
Alpine.js evaluates showIf expression
   ↓
Field becomes visible (no server call)
```

```
User role = "admin" → Field appears
   ↓
Server enricher checks permission
   ↓
Sets field.Runtime.Visible = true
   ↓
templ respects flag during render
```

#### Pattern: Permission-Based Form

```
1. User requests form
2. Server loads base schema
3. Server gets user permissions
4. Server enriches schema with runtime flags
5. Server renders only accessible fields
6. Client displays enriched HTML
```

### 4.7 Error Types

**Validation Error**
- Field value doesn't meet requirements
- Example: "Email is required", "Min length is 5"
- Displayed near field

**Schema Error**
- Schema structure is invalid
- Example: "Duplicate field name", "Missing required property"
- Prevents schema from loading

**Permission Error**
- User lacks required permission
- Example: "Insufficient permissions to view this field"
- Field is hidden or disabled

**System Error**
- Server error, database error
- Example: "Database connection failed"
- Display generic error message

### 4.8 Glossary

| Term | Definition |
|------|------------|
| **Schema** | Complete UI definition in JSON format |
| **Field** | Single input element with type, label, validation |
| **Action** | Button or clickable element (submit, reset, etc.) |
| **Layout** | Visual arrangement of fields (grid, tabs, etc.) |
| **Validation** | Rules determining if field values are acceptable |
| **Enrichment** | Adding runtime data to schema (permissions, etc.) |
| **Runtime** | Data added during request processing (not stored) |
| **Design-Time** | Static schema definition (stored in DB/files) |
| **HTMX** | Library for HTML over the wire (AJAX without JS) |
| **Alpine.js** | Minimal reactive framework for client state |
| **templ** | Type-safe Go template language |
| **Registry** | Service managing schema storage/retrieval |
| **Parser** | Converts JSON to Go structs |
| **Validator** | Checks schema structure and data validity |
| **Condition** | Logic determining when fields show/hide/required |
| **Permission** | User capability checked by server |
| **Tenant** | Isolated customer/organization in multi-tenant system |
| **Mixin** | Reusable collection of fields (like templates) |
| **Business Rule** | Complex validation logic using condition package |

---

