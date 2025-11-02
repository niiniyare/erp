# Introduction - Schema-Driven UI

## What is Schema-Driven Development?

Schema-driven development is where **backend developers define complete user interfaces using JSON schemas** instead of writing frontend code. The schema is a contract describing:

- What fields exist and their types
- What validation rules apply
- What layout structure to use
- What permissions control access

**Traditional Approach:**
```
Backend → API
Frontend → React/Vue components + validation + state
Result: 2-3 weeks
```

**Schema-Driven Approach:**
```
Backend → JSON schema
Server → Renders HTML
Result: Same day
```

## Example Schema

```json
{
  "id": "contact-form",
  "type": "form",
  "title": "Contact Us",
  "fields": [
    {
      "name": "email",
      "type": "email",
      "label": "Email Address",
      "required": true
    },
    {
      "name": "message",
      "type": "textarea",
      "label": "Message",
      "required": true
    }
  ],
  "config": {
    "action": "/api/contact",
    "method": "POST"
  }
}
```

This JSON automatically provides:
- HTML form with proper attributes
- Client-side validation (HTML5)
- Server-side validation (Go)
- HTMX-powered submission
- Error handling
- Success feedback

**No frontend code required.**

## Why Server-Side Rendering?

This system uses **Server-Side Rendering (SSR)** with progressive enhancement.

### Performance Benefits

| Metric | SPA | SSR | Advantage |
|--------|-----|-----|-----------|
| Initial Load | 1-3MB | 10-50KB | **50-300x smaller** |
| First Paint | 2-5s | ~200ms | **10-25x faster** |
| Time to Interactive | 3-5s | Immediate | **Instant** |
| Works without JS | No | Yes | **Progressive** |
| Type Safety | Runtime | Compile-time | **Safer** |

### Request Flow

**SSR Flow:**
```
Browser → GET /page → Server
                      ├─ Parse schema
                      ├─ Enrich with permissions
                      ├─ Render with templ
                      └─ Return HTML
Browser ← Full HTML (200ms)
        ↓
    (HTMX enhances)
    (Alpine.js for state)
```

## When to Use This System

**✅ Perfect For:**
- Enterprise forms and CRUD interfaces
- Data entry applications
- Admin panels
- Multi-step workflows
- Permission-based forms
- Multi-tenant systems

**❌ Not Ideal For:**
- Highly interactive apps (games, drawing tools)
- Real-time collaboration (Google Docs-style)
- Offline-first applications
- Complex client-side state machines

## Core Principles

### 1. Backend is Source of Truth
All security decisions happen on the server.

```go
// Server decides visibility
field.Runtime = &FieldRuntime{
    Visible: user.HasPermission("view_salary"),
}
```

### 2. Schema is Pure Data
Describes WHAT, not HOW.

```json
// ✅ Good: Describes structure
{
  "name": "email",
  "type": "email",
  "required": true
}

// ❌ Bad: Includes implementation
{
  "name": "email",
  "component": "MyEmailComponent",
  "reactProps": {...}
}
```

### 3. Progressive Enhancement
HTML works without JavaScript.

```html
<!-- Base: Works without JS -->
<form action="/api/submit" method="POST">
  <input type="email" name="email" required />
  <button type="submit">Submit</button>
</form>

<!-- Enhanced: HTMX adds AJAX -->
<form hx-post="/api/submit" hx-target="#result">
  <input type="email" name="email" required />
  <button type="submit">Submit</button>
</form>
```

### 4. Type Safety
Go's compile-time checks prevent errors.

```go
// Compile-time error if structure is wrong
type Schema struct {
    ID     string      `json:"id" validate:"required"`
    Type   SchemaType  `json:"type" validate:"required"`
    Fields []Field     `json:"fields" validate:"dive"`
}
```

### 5. Flexible Storage
Choose the storage that fits your needs:
- **PostgreSQL** - Primary storage with ACID
- **Redis** - Fast caching layer
- **Filesystem** - Development and simple deployments
- **S3/Minio** - Object storage for cloud environments

## Technology Stack

**Backend:**
- Go 1.21+ (type-safe, compiled)
- Fiber v2 (fast web framework)
- templ (type-safe Go templates)
- go-playground/validator (struct validation)

**Frontend:**
- HTMX 1.9+ (HTML over the wire)
- Alpine.js 3.x (minimal reactivity)
- Standard HTML5 (native validation)

**Storage:**
- PostgreSQL 14+ (primary)
- Redis 7+ (caching)
- S3/Minio (optional)
- Filesystem (development)

## System Architecture

```
┌─────────────────────────────────────┐
│  Schema Definition (JSON)           │
│  ├─ Fields, Types, Labels           │
│  ├─ Validation Rules                │
│  └─ Layout Structure                │
└────────────┬────────────────────────┘
             ↓
┌─────────────────────────────────────┐
│  Processing (Go)                    │
│  ├─ Parser: JSON → Schema           │
│  ├─ Validator: Check rules          │
│  ├─ Enricher: Add permissions       │
│  └─ Registry: Cache schemas         │
└────────────┬────────────────────────┘
             ↓
┌─────────────────────────────────────┐
│  Rendering (templ)                  │
│  ├─ Field → Component mapping       │
│  ├─ Validation → HTML attributes    │
│  └─ HTMX → Progressive enhancement  │
└────────────┬────────────────────────┘
             ↓
         HTML Output
```

## Complete Request Lifecycle

```
1. Browser Request
   ↓
   GET /forms/user-registration

2. Fiber Handler
   ↓
   registry.Get("user-registration")
   ├─ Check Redis cache
   ├─ If miss, load from PostgreSQL
   └─ Return *Schema

3. Enricher
   ↓
   enricher.Enrich(schema, user)
   ├─ Get user permissions
   ├─ Apply tenant customization
   ├─ Add default values
   └─ Set field.Runtime flags

4. Renderer
   ↓
   views.FormPage(schema, data)
   ├─ Loop through fields
   ├─ Render components
   ├─ Add HTMX attributes
   └─ Generate HTML

5. Response
   ↓
   HTML sent to browser (200ms)

6. Browser
   ↓
   ├─ Display HTML immediately
   ├─ HTMX enhances interactions
   └─ Alpine.js manages state
```

## Key Components

### Schema
The root JSON document defining a form or component.

### Field
An input element with type, validation, and configuration.

### Registry
Storage layer managing schemas (SQL, Redis, files, S3).

### Enricher
Adds runtime data (permissions, defaults, tenant config).

### Renderer
Converts schemas to HTML using templ components.

### Validator
Validates schema structure and form data.

## Next Steps

- [02-architecture.md](02-architecture.md) - Detailed architecture
- [03-quick-start.md](03-quick-start.md) - Build your first form
- [04-schema-structure.md](04-schema-structure.md) - Complete schema reference

---

[← Back to README](README.md) | [Next: Architecture →](02-architecture.md)