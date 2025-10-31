# Schema System

A comprehensive JSON-driven UI schema system for building enterprise forms and components. Backend developers can define complete UIs by writing JSON files - no frontend code required.

## Overview

The schema system enables:
- **JSON-First Development**: Define forms entirely in JSON
- **Multi-Tenancy**: Built-in tenant isolation with strict/shared/hybrid modes
- **Enterprise Security**: CSRF protection, rate limiting, field encryption
- **Workflow Support**: Approval workflows with configurable stages
- **Framework Integration**: HTMX and Alpine.js support out of the box
- **Type Safety**: Comprehensive Go structs with validation

## Quick Start

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

**This JSON becomes:**
- ✅ HTML form with proper attributes
- ✅ Client-side validation (HTML5)
- ✅ Server-side validation (Go)
- ✅ HTMX-powered submission
- ✅ Error handling and success feedback

**No frontend code required.**

## Installation

```bash
# Install templ
go install github.com/a-h/templ/cmd/templ@latest

# Install schema package
go get github.com/awo-erp/schema

# Install validator
go get github.com/go-playground/validator/v10
```

Include HTMX and Alpine.js via CDN:
```html
<script src="https://unpkg.com/htmx.org@1.9.10"></script>
<script src="https://unpkg.com/alpinejs@3.13.5/dist/cdn.min.js"></script>
```

## Documentation Structure

### Part I: Foundations
- [01. Introduction](./schema/01-introduction.md) - Schema-driven development concepts
- [02. Architecture](./schema/02-architecture.md) - System architecture overview
- [03. Quick Start](./schema/03-quickstart.md) - 5-minute setup guide
- [04. Core Concepts](./schema/04-concepts.md) - Terminology and fundamentals
- [05. Comparison](./schema/05-comparison.md) - Comparison with amis and alternatives

### Part II: Core Schema Specification
- [06. Schema Structure](./schema/06-schema-structure.md) - JSON schema format
- [07. Field System](./schema/07-field-system.md) - All 40+ field types
- [08. Validation](./schema/08-validation.md) - Validation rules and patterns
- [09. Layout System](./schema/09-layout.md) - Responsive layouts and components
- [10. Conditional Logic](./schema/10-conditional.md) - Dynamic field visibility

### Part III: Advanced Features  
- [11. Security](./schema/11-security.md) - Multi-tenancy and permissions
- [12. Events & Actions](./schema/12-events.md) - HTMX integration
- [13. Workflow System](./schema/13-workflow.md) - Approval workflows
- [14. Registry System](./schema/14-registry.md) - Component registration
- [15. Validator](./schema/15-validator.md) - Validation implementation

### Part IV: Implementation
- [16. Enricher System](./schema/16-enricher.md) - Data enrichment patterns
- [17. Renderer System](./schema/17-renderer.md) - templ integration
- [18. Handler Patterns](./schema/18-handlers.md) - HTTP handler patterns
- [19. Data Sources](./schema/19-datasources.md) - Database integration
- [20. Testing](./schema/20-testing.md) - Testing strategies

## Architecture

**Server-Side Rendering** with progressive enhancement:

```
Backend Dev → Writes JSON schema → Server renders HTML → Browser receives ready content
```

**Performance Benefits:**
- Initial Load: ~10KB vs 1-3MB SPA bundles
- First Paint: ~200ms vs 2-5 seconds
- Works without JavaScript
- Native SEO support

## Key Features

| Feature | Status | Description |
|---------|--------|-------------|
| **40+ Field Types** | ✅ Complete | text, email, select, checkbox, file upload, etc. |
| **Layout System** | ✅ Complete | Responsive grids, cards, tabs, accordions |
| **Validation** | ✅ Complete | HTML5 + server-side rules |
| **Multi-Tenancy** | ✅ Complete | Strict/shared/hybrid isolation |
| **Security** | ✅ Complete | CSRF, rate limiting, field encryption |
| **HTMX Integration** | ✅ Complete | Dynamic form submission and updates |
| **Workflow Support** | ✅ Complete | Approval workflows with stages |
| **Type Safety** | ✅ Complete | Go structs with compile-time checking |

## Use Cases

**Perfect for:**
- ✅ Enterprise forms and CRUD interfaces
- ✅ Admin panels and dashboards  
- ✅ Multi-tenant SaaS applications
- ✅ Backend-heavy teams
- ✅ Performance-critical applications

**Consider SPA when:**
- Complex client-side interactions (games, drawing tools)
- Offline-first requirements
- Real-time collaboration features

## Status

**Version:** 1.0.0  
**Production Ready:** ✅  
**Used in:** Awo ERP (80% complete financial module)  
**Target:** Multi-tenant SaaS with white-label capabilities

---

**Get Started:** Begin with [Introduction](./schema/01-introduction.md) or jump to [Quick Start](./schema/03-quickstart.md) for immediate implementation.