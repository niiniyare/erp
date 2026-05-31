# Awo ERP API Handler Development Guide

**Version:** 5.0 - Complete Edition (Architecture + Implementation)  
**Last Updated:** October 26, 2025  
**Audience:** Developers (All Levels), Architects, Technical Leaders, Product Managers  
**System:** Awo ERP - Multi-tenant Enterprise Resource Planning Platform

---

##  How to Use This Guide

**New Developers**: Start with Quick Start (§2), then read Part II linearly  
**Experienced Developers**: Use as reference, jump to specific patterns  
**Architects**: Focus on Part I and "Why This Choice" sections  
**Non-Technical Readers**: Read "Analogy" sections, skip code blocks

---

## Table of Contents

### PART I: FOUNDATIONS & ARCHITECTURE
1. [Introduction & Philosophy](#1-introduction--philosophy)
2. [Quick Start Guide](#2-quick-start-guide)
3. [Architectural Overview](#3-architectural-overview)
4. [Technology Choices & Rationale](#4-technology-choices--rationale)
5. [Project Structure](#5-project-structure)

### PART II: CORE HANDLER DEVELOPMENT
6. [Type System & Contract Design](#6-type-system--contract-design)
7. [Handler Implementation Patterns](#7-handler-implementation-patterns)
8. [Request Flow & Processing Pipeline](#8-request-flow--processing-pipeline)
9. [Error Handling Strategy](#9-error-handling-strategy)

### PART III: ADVANCED INTEGRATION TOPICS
10. [Multi-Tenancy & Data Isolation](#10-multi-tenancy--data-isolation)
11. [UI Integration with TemplUI](#11-ui-integration-with-templui)
12. [DataTable Patterns & Metadata](#12-datatable-patterns--metadata)
13. [Content-Type Negotiation](#13-content-type-negotiation)
14. [Pagination & Filtering](#14-pagination--filtering)

### PART IV: QUALITY ASSURANCE & OPERATIONS
15. [Testing Strategy & Patterns](#15-testing-strategy--patterns)
16. [Common Patterns & Best Practices](#16-common-patterns--best-practices)
17. [Troubleshooting Guide](#17-troubleshooting-guide)

---

# PART I: FOUNDATIONS & ARCHITECTURE

---

## 1. Introduction & Philosophy

### 1.1 Purpose & Scope

This guide is the **definitive resource** for developing API handlers in the Awo ERP system. It uniquely combines:

- **Architectural Theory**: Why we made specific technology and design choices
- **Practical Implementation**: Step-by-step code examples and patterns
- **Industry Context**: How our approach compares to systems like Stripe, Salesforce, and Shopify
- **Development Guidelines**: Best practices and common pitfalls to avoid

**What This Guide Covers:**
- ✅ Handler structure and patterns (with code)
- ✅ Goa type definitions (theory + examples)
- ✅ Error handling with user-friendly messages
- ✅ Database Store integration (RLS, transactions)
- ✅ UI integration via TemplUI components
- ✅ DataTable patterns with actions and metadata
- ✅ Content-type negotiation (JSON/HTML)
- ✅ Testing strategies with examples
- ✅ Architectural decisions and rationale

**What This Guide Does NOT Cover:**
- ❌ Service/Workflow implementation details → See `@internal/core/` and `@internal/workflows/`
- ❌ Domain logic → Handled in service layer
- ❌ Database schema design → See migrations
- ❌ Infrastructure setup → Docker, Kubernetes, etc.

### 1.2 Key Technologies

| Technology | Purpose | Location |
|------------|---------|----------|
| **Goa** | Type generation from DSL | `@internal/api/design/` |
| **Fiber** | HTTP framework | Handlers use `*fiber.Ctx` |
| **DB Store** | Database + RLS + Transactions | `@internal/db/sqlc/` |
| **TemplUI** | UI component generation | `@web/components/` |
| **HTMX** | Dynamic UI updates | Frontend framework |

### 1.3 What Are API Handlers?

Think of API handlers as the **reception desk** of your application. When someone (a user, another system, or a mobile app) wants to interact with your ERP system, they send a request—like asking for account information or creating a new transaction. The API handler is responsible for:

1. **Receiving** the request and understanding what's being asked
2. **Validating** that the request makes sense and is properly formatted
3. **Coordinating** with the business logic layer to fulfill the request
4. **Packaging** the response in the appropriate format
5. **Sending** the response back to the requester

**Real-World Analogy**: Imagine a restaurant. The API handler is like the waiter who:
- Takes your order (receives request)
- Checks if what you ordered exists on the menu (validates)
- Passes your order to the kitchen (delegates to business logic)
- Brings you the food when ready (returns response)
- Handles any issues, like telling you an item is sold out (error handling)

### 1.2 What This Guide Covers

This guide focuses exclusively on **the handler layer**—the interface between the outside world and your business logic. It does NOT cover:

- **Business Logic**: The "what should happen" rules (e.g., calculating interest, applying discounts)
- **Database Design**: How data is structured and stored
- **Infrastructure**: Servers, deployment, scaling

**Why This Separation?**

This follows the **Separation of Concerns** principle, used by major systems like:
- **Stripe API**: Handlers process payments, but business logic calculates fees
- **Shopify API**: Handlers manage product requests, but inventory logic lives elsewhere
- **GitHub API**: Handlers serve repository data, but Git operations are separate

This separation allows us to:
- Change how we store data without rewriting handlers
- Modify business rules without touching API contracts
- Replace technology stacks independently
- Test each layer in isolation

### 1.3 Document Scope

**For Developers**: This guide provides architectural understanding and conceptual frameworks for building handlers.

**For Non-Technical Readers**: Each section includes analogies and explanations of why certain decisions were made, without assuming programming knowledge.

**For Technical Leaders**: Architectural decisions reference industry standards and explain trade-offs between different approaches.

---



---


## 2. Quick Start

### 2.1 Adding a New Endpoint (5-Step Process)

Let's add an endpoint to archive an account:

#### Step 1: Define Types in Goa Design

```go
// @internal/api/design/finance.go
package design

import . "goa.design/goa/v3/dsl"

// Types only - no methods/services
var AccountArchivePayload = Type("AccountArchivePayload", func() {
    Field(1, "id", String, "Account ID", UUID)
    Required("id")
})

var AccountArchiveResult = Type("AccountArchiveResult", func() {
    Field(1, "id", String, "Account ID", UUID)
    Field(2, "name", String, "Account name")
    Field(3, "code", String, "Account code")
    Field(4, "archived", Boolean, "Archive status")
    Field(5, "archived_at", String, "Archive timestamp", Timestamp)
    Required("id", "name", "code", "archived")
})
```

#### Step 2: Generate Types

```bash
$ goa gen awo/internal/api/design
# ✓ Generated: internal/api/gen/finance/types.go
```

#### Step 3: Write Handler

```go
// @internal/api/handlers/finance/account_handler.go
package finance

import (
    "github.com/gofiber/fiber/v2"
    "awo.so/internal/api/gen/finance"
    "awo.so/internal/core/account"
)

type AccountHandler struct {
    accountService account.Service
    // ... other deps
}

func (h *AccountHandler) Archive(c *fiber.Ctx) error {
    // 1. Parse path parameter
    accountID := c.Params("id")
    
    // 2. Extract context
    ctx := c.Context()
    userID := c.Locals("user_id").(string)
    
    // 3. Call service (business logic)
    result, err := h.accountService.Archive(ctx, accountID, userID)
    if err != nil {
        return h.handleError(c, err)
    }
    
    // 4. Map to response type
    response := &finance.AccountArchiveResult{
        ID:         result.ID,
        Name:       result.Name,
        Code:       result.Code,
        Archived:   result.Archived,
        ArchivedAt: result.ArchivedAt,
    }
    
    // 5. Return response (content negotiation handled automatically)
    return h.respond(c, fiber.StatusOK, response)
}
```

#### Step 4: Register Route

```go
// @internal/api/routes/routes.go
accounts := api.Group("/accounts",
    mw.Auth,   // Authentication
    mw.Tenant, // RLS setup
)

accounts.Post("/:id/archive", accountHandler.Archive)
```

#### Step 5: Test

```bash
curl -X POST http://localhost:8080/api/v1/accounts/123/archive \
  -H "Authorization: Bearer <token>"
```



---



## 4. Request Lifecycle

### 4.1 HTTP Request Flow

```
┌────────────────────────────────────────────────────────┐
│ 1. HTTP Request Arrives                                │
│    POST /api/v1/accounts                               │
│    Content-Type: application/json                      │
│    Authorization: Bearer <token>                       │
└─────────────────┬──────────────────────────────────────┘
                  │
┌─────────────────▼──────────────────────────────────────┐
│ 2. Fiber Router                                        │
│    - Matches route                                     │
│    - Applies middleware chain                          │
└─────────────────┬──────────────────────────────────────┘
                  │
┌─────────────────▼──────────────────────────────────────┐
│ 3. Middleware Chain                                    │
│    A. Recovery (panic handling)                        │
│    B. Request ID (generate/extract)                    │
│    C. Auth (validate JWT, extract user/tenant)         │
│    D. Tenant (RLS setup via Store.SetTenantContext)    │
│    E. Observability (logging, tracing, metrics)        │
└─────────────────┬──────────────────────────────────────┘
                  │
┌─────────────────▼──────────────────────────────────────┐
│ 4. Handler Execution                                   │
│    - Parse request body                                │
│    - Validate (Goa types)                              │
│    - Extract context                                   │
│    - Call service/workflow                             │
│    - Map to response type                              │
│    - Return response (JSON or HTML)                    │
└─────────────────┬──────────────────────────────────────┘
                  │
┌─────────────────▼──────────────────────────────────────┐
│ 5. Service/Workflow (@internal/core or @workflows)    │
│    - Business logic execution                          │
│    - Database operations via Store                     │
└─────────────────┬──────────────────────────────────────┘
                  │
┌─────────────────▼──────────────────────────────────────┐
│ 6. Database (via Store interface)                     │
│    - RLS enforced (tenant_id from middleware)          │
│    - Transaction support (Store.WithTx)                │
│    - SQLC-generated queries                            │
└─────────────────┬──────────────────────────────────────┘
                  │
┌─────────────────▼──────────────────────────────────────┐
│ 7. Response Construction                               │
│    - Content-Type negotiation                          │
│    - JSON: c.JSON(response)                            │
│    - HTML: h.renderComponent(c, component)             │
└─────────────────┬──────────────────────────────────────┘
                  │
┌─────────────────▼──────────────────────────────────────┐
│ 8. HTTP Response                                       │
│    Status: 201 Created                                 │
│    Body: JSON or HTML                                  │
└────────────────────────────────────────────────────────┘
```

### 4.2 Key Points

1. **Middleware handles cross-cutting concerns** - handlers focus on business logic
2. **RLS is set once in middleware** - handlers never touch database directly
3. **Services encapsulate business logic** - handlers delegate to services
4. **Content negotiation is automatic** - handlers return data, framework handles format



---


## 3. Technology Choices & Rationale

### 3.1 Why Goa for Type Definitions?

**What is Goa?** A tool that generates code from type definitions. You describe what your API looks like, and Goa creates the necessary code.

**The Problem It Solves**:

Traditional approach:
1. Write API handler code
2. Write type definitions
3. Write validation code
4. Write API documentation
5. Keep all four in sync (❌ error-prone)

**Goa Approach**:
1. Write type definition once
2. Goa generates everything else (✓ always in sync)

**Real-World Examples**:
- **Google**: Uses Protocol Buffers (similar concept) for 12,000+ services
- **Twilio**: Uses OpenAPI specifications (similar declarative approach)
- **Stripe**: API definitions generate client libraries automatically

**Why Awo ERP Chose Goa Over Alternatives**:

| Approach | Pros | Cons | Awo ERP Decision |
|----------|------|------|------------------|
| **Manual coding** | Full control | High maintenance, inconsistency | ❌ Too error-prone |
| **OpenAPI/Swagger** | Industry standard | Verbose, lacks Go integration | ❌ Not Go-native |
| **gRPC/Protobuf** | Fast, type-safe | Complex setup, not REST | ❌ Overkill for HTTP |
| **Goa** | Go-native, type-safe, generates validation | Smaller ecosystem | ✅ Best fit |

**Benefits Realized**:
- **Zero** type mismatches between request/response (compiler catches errors)
- **Automatic** validation (no manual if/else chains)
- **Instant** documentation updates (types = documentation)

### 3.2 Why Fiber Web Framework?

**What is Fiber?** A web framework that handles HTTP requests/responses. Think of it as the foundation for building web APIs.

**Key Characteristics**:
- **Performance**: Handles 50,000+ requests/second (vs Express.js: 15,000/s)
- **Simplicity**: Similar to Express.js (popular Node.js framework)
- **Low memory**: 10MB per 10,000 connections (vs Gin: 25MB)

**Comparison with Alternatives**:

| Framework | Speed | Memory | Ecosystem | Awo ERP Decision |
|-----------|-------|--------|-----------|------------------|
| **net/http** (standard) | Medium | Low | Huge | ❌ Too basic |
| **Gin** | Fast | Medium | Large | ❌ Lacks features |
| **Echo** | Fast | Low | Medium | ❌ Less intuitive |
| **Fiber** | Fastest | Lowest | Growing | ✅ Best performance |

**Real-World Usage**:
- **PayPal**: Uses similar high-performance frameworks for payment APIs
- **Netflix**: Switched to performance-first frameworks for streaming APIs
- **Discord**: Uses fast frameworks for real-time messaging (millions of concurrent users)

**Why It Matters for Awo ERP**:
- Multi-tenant systems need efficient resource usage (one server, hundreds of tenants)
- Financial data requires low latency (users expect instant responses)
- Cost efficiency (fewer servers needed = lower infrastructure costs)

### 3.3 Why SQLC for Database Access?

**What is SQLC?** Generates Go code from SQL queries. You write SQL, it creates type-safe Go functions.

**The Database Access Problem**:

Traditional approaches:
1. **Raw SQL strings**: `db.Query("SELECT * FROM accounts WHERE id = " + id)` → SQL injection risk
2. **ORMs** (Hibernate, Sequelize): Hide SQL, slow performance, complex queries difficult
3. **Query builders**: Verbose, learning curve, abstracts SQL too much

**SQLC Approach**:
- Write standard SQL (database optimized)
- SQLC generates type-safe Go code
- Get compile-time safety + SQL performance

**Industry Precedent**:
- **Google**: Uses code generation for database access (internal tools)
- **Facebook**: Generates data access code from schemas
- **LinkedIn**: Auto-generates data access layers

**Why Awo ERP Chose SQLC**:

| Aspect | ORM (GORM) | Query Builder | Raw SQL | SQLC | Decision |
|--------|------------|---------------|---------|------|----------|
| Type safety | ❌ Runtime | ⚠️ Partial | ❌ None | ✅ Compile-time | ✅ SQLC |
| Performance | ⚠️ Slower | ✅ Good | ✅ Best | ✅ Best | ✅ SQLC |
| SQL visibility | ❌ Hidden | ⚠️ Abstracted | ✅ Clear | ✅ Clear | ✅ SQLC |
| Complexity | ⚠️ High | ⚠️ Medium | ✅ Low | ✅ Low | ✅ SQLC |

**Benefits**:
- Database experts can optimize SQL directly
- Compiler catches mistakes (wrong types, missing fields)
- No runtime overhead (no ORM translation layer)
- Clear audit trail (see exactly what SQL runs)

### 3.4 Why TemplUI for UI Components?

**What is TemplUI?** Generates Go templates from component definitions. Instead of writing HTML, you call components.

**The UI Problem**:

Traditional approach:
1. Backend returns JSON
2. Frontend parses JSON
3. Frontend builds HTML
4. Result: Two codebases, separate deployment, complex coordination

**TemplUI Approach**:
- Backend generates HTML directly
- Single codebase, single deployment
- Frontend uses HTMX (minimal JavaScript)

**Similar to**:
- **Ruby on Rails**: Server-side rendering with partials
- **Phoenix LiveView**: Server-generated UI updates
- **Laravel Livewire**: Backend-driven interfaces

**Why Awo ERP Chose This**:

Traditional SPA (React/Vue):
- ✅ Rich interactivity
- ❌ Complex build process
- ❌ Separate API layer required
- ❌ SEO challenges
- ❌ Larger bundle sizes

Server-Side Rendering (TemplUI + HTMX):
- ✅ Simple deployment
- ✅ Shared types between API and UI
- ✅ Built-in SEO
- ✅ Smaller bundle
- ⚠️ Less rich client interactions (acceptable for ERP use case)

**Industry Trend**: Return to server-side rendering
- **Basecamp (Hey.com)**: Built with Hotwire (similar approach)
- **GitHub**: Uses server-rendered HTML with minimal JS
- **Stack Overflow**: Server-side rendering for performance

**For ERP Systems Specifically**:
- Forms, tables, and data entry don't need heavy JavaScript
- Users value speed and reliability over rich animations
- Simpler stack = easier maintenance (critical for long-term business software)

---



---


## 3. Project Structure

### 3.1 API-Related Directories

```
awo-erp/
├── internal/
│   ├── api/
│   │   ├── design/               # Goa type definitions
│   │   │   ├── common.go         # Shared types (UUID, Pagination, etc.)
│   │   │   ├── tenant.go         # Tenant types
│   │   │   ├── finance.go        # Finance types
│   │   │   └── iam.go            # IAM types
│   │   │
│   │   ├── gen/                  # Goa-generated types
│   │   │   └── finance/          # Generated finance types
│   │   │       └── types.go
│   │   │
│   │   ├── handlers/             # HTTP handlers
│   │   │   ├── tenant/
│   │   │   │   └── handler.go
│   │   │   │
│   │   │   ├── finance/
│   │   │   │   ├── account_handler.go
│   │   │   │   ├── transaction_handler.go
│   │   │   │   └── budget_handler.go
│   │   │   │
│   │   │   └── iam/
│   │   │       └── user_handler.go
│   │   │
│   │   └── routes/
│   │       └── routes.go         # Route registration
│   │
│   ├── core/                     # Business logic → Call from handlers
│   ├── workflows/                # Multi-service orchestration → Call from handlers
│   ├── db/sqlc/                  # Database Store interface
│   ├── platform/middleware/      # Custom middleware
│   └── shared/                   # Errors, logger, metrics, tracing
│
└── web/components/               # TemplUI-generated components
    ├── datatable.templ           # DataTable component
    ├── account_row.templ         # Account row component
    └── error_message.templ       # Error display component
```

### 3.2 Handler Organization

**Single-Service Modules**: One handler file
```
internal/api/handlers/tenant/handler.go  # All tenant CRUD
```

**Multi-Service Modules**: One file per sub-service
```
internal/api/handlers/finance/
├── account_handler.go      # Chart of accounts
├── transaction_handler.go  # Transactions
└── budget_handler.go       # Budgets
```

**Complex Handlers**: Split by related operations (if >500 lines)
```
internal/api/handlers/finance/account/
├── create.go    # Create + validation
├── update.go    # Update operations
├── list.go      # List + pagination
└── archive.go   # Archive operations
```




---

# PART II: CORE HANDLER DEVELOPMENT

---

## 5. Type System & Contract Design

### 5.1 Why Types Matter

**What Are Types?**

Types define the "shape" of data. Like a form with specific fields:
- Name: (must be text, max 100 characters)
- Age: (must be number, between 0-120)
- Email: (must be valid email format)

**Why Types Are Critical**:

Without types:
```
User sends: { "account_code": 1001 } (number)
Handler expects: { "account_code": "1001" } (string)
→ System breaks (type mismatch)
```

With types:
```
Type definition says: account_code must be string
Compiler checks: Is it string? No → ERROR AT BUILD TIME
→ Bug caught before deployment
```

**Real-World Impact**:

**Knight Capital (2012)**:
- Type mismatch in trading system
- Lost $440 million in 45 minutes
- Company bankrupted

**Equifax Breach (2017)**:
- Data validation failure (wrong types accepted)
- 147 million records exposed
- $700 million settlement

**Awo ERP Approach**:
- All types defined upfront
- Compiler validates at build time
- Runtime validation as second layer

### 5.2 Goa Type Definitions

**Centralized Type Definitions**:

All API types live in one location: `@internal/api/design/`

```
design/
├── common.go      → Shared types (UUID, Pagination, Timestamps)
├── tenant.go      → Tenant-related types
├── finance.go     → Finance module types
├── iam.go         → Identity & Access types
└── inventory.go   → Inventory module types
```

**Why Centralized?**

**Single Source of Truth**:
- One place to look for type definitions
- Changes ripple through entire system
- No duplicate definitions (leading to inconsistencies)

**Similar to**:
- **GraphQL Schema**: Centralized type definitions
- **Protobuf**: Single .proto file per service
- **OpenAPI Specification**: Centralized API contract

### 5.3 Type Categories

**1. Request Types (Input)**:

Define what clients send to API:
- Create operations: Required fields
- Update operations: Optional fields (change only what's needed)
- Query parameters: Filters, pagination, sorting

**Example - Creating an Account**:
```
Required fields:
- Name (3-100 characters)
- Code (4-10 digits)
- Type (must be: ASSET, LIABILITY, EQUITY, REVENUE, EXPENSE)

Optional fields:
- Parent account ID
- Description
```

**2. Response Types (Output)**:

Define what API returns to clients:
- Success responses: Created/updated resource
- List responses: Array of items + pagination
- Error responses: Error code + user message

**Example - Account Response**:
```
Always included:
- ID (UUID)
- Name
- Code
- Type
- Status (active/archived)
- Timestamps (created, updated)

Sometimes included:
- Balance (only if requested)
- Child accounts (only for detailed view)
```

**3. Metadata Types**:

Define configuration for UI components:
- Table columns (which fields to show?)
- Available actions (create, edit, delete)
- Filters (what can users filter by?)
- Sort options (what columns are sortable?)

This allows **server to control UI behavior**—change backend, UI updates automatically.

### 5.4 Validation Rules

Types include **validation rules** that catch errors early:

**Format Validation**:
- UUIDs must match UUID pattern
- Emails must be valid email format
- Dates must be ISO8601 format
- Phone numbers must match regional format

**Range Validation**:
- Pagination: page must be ≥ 1, per_page must be 1-100
- Amounts: must be positive for revenue, negative allowed for expenses
- Quantities: must be ≥ 0

**Business Validation**:
- Enum fields: must be one of allowed values
- Required relationships: parent_id must exist if provided
- String lengths: min/max characters

**Why Multi-Layer Validation?**

**Layer 1 - Type System** (Compile Time):
- Catches type mismatches
- Ensures required fields exist
- No runtime cost

**Layer 2 - Goa Validation** (Request Time):
- Validates format (regex patterns)
- Checks ranges (min/max values)
- Fast execution (milliseconds)

**Layer 3 - Business Logic** (Service Layer):
- Checks business rules (account code uniqueness)
- Verifies relationships (parent account exists)
- Database-dependent validation

**Analogy**: Airport security (multiple checkpoints)
1. Ticket check (type validation) - do you have a ticket?
2. TSA screening (format validation) - is luggage acceptable?
3. Gate agent (business validation) - is flight still boarding?

**Industry Standard**:
- **Stripe API**: Validates credit cards at multiple layers
- **AWS API**: Progressive validation (format → quota → business)
- **Salesforce**: Field validation + automation rules + triggers

---


## 5. Goa Type Design

### 5.1 Design Philosophy

Goa designs define **types only** - no methods or services. All type definitions are in `@internal/api/design/`. Goa will generate types under **one package** (`@internal/api/gen/`), but design files are **separated** by domain.

### 5.2 Common Type Patterns

#### Pattern 1: Shared Types

```go
// @internal/api/design/common.go
package design

import . "goa.design/goa/v3/dsl"

// UUID pattern (used everywhere)
var UUID = func() {
    Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
    Example("550e8400-e29b-41d4-a716-446655440000")
}

// Timestamp format
var Timestamp = func() {
    Format(FormatDateTime)
    Example("2025-01-15T10:30:00Z")
}

// Pagination params (used in all list operations)
var PaginationParams = Type("PaginationParams", func() {
    Field(1, "page", Int, "Page number", func() {
        Default(1)
        Minimum(1)
    })
    Field(2, "per_page", Int, "Items per page", func() {
        Default(20)
        Minimum(1)
        Maximum(100)
    })
    Required("page", "per_page")
})

// Pagination response metadata
var PaginationResponse = Type("PaginationResponse", func() {
    Field(1, "total", Int, "Total number of items")
    Field(2, "page", Int, "Current page number")
    Field(3, "per_page", Int, "Items per page")
    Field(4, "total_pages", Int, "Total number of pages")
    Required("total", "page", "per_page", "total_pages")
})
```

#### Pattern 2: Resource Types (with Views)

```go
// @internal/api/design/tenant.go
package design

import . "goa.design/goa/v3/dsl"

var Tenant = ResultType("application/vnd.erp.tenant", func() {
    Description("Business tenant")
    
    Attributes(func() {
        Field(1, "id", String, "Tenant UUID", UUID)
        Field(2, "slug", String, "URL-friendly identifier")
        Field(3, "name", String, "Tenant name")
        Field(4, "email", String, "Contact email", func() {
            Format(FormatEmail)
        })
        Field(5, "status", String, "Tenant status", func() {
            Enum("ACTIVE", "SUSPENDED", "TRIAL", "ARCHIVED")
        })
        Field(6, "created_at", String, "Creation timestamp", Timestamp)
        Field(7, "updated_at", String, "Last update timestamp", Timestamp)
        Field(8, "metadata", MapOf(String, Any), "Additional metadata")
        
        Required("id", "slug", "name", "email", "status")
    })
    
    // Default view (for lists)
    View("default", func() {
        Attribute("id")
        Attribute("slug")
        Attribute("name")
        Attribute("status")
        Attribute("created_at")
    })
    
    // Detailed view (for single resource)
    View("detailed", func() {
        Attribute("id")
        Attribute("slug")
        Attribute("name")
        Attribute("email")
        Attribute("status")
        Attribute("metadata")
        Attribute("created_at")
        Attribute("updated_at")
    })
})
```

#### Pattern 3: Request/Response Types

```go
// @internal/api/design/finance.go
package design

import . "goa.design/goa/v3/dsl"

// Create account payload
var CreateAccountPayload = Type("CreateAccountPayload", func() {
    Field(1, "name", String, "Account name", func() {
        MinLength(3)
        MaxLength(100)
    })
    Field(2, "code", String, "Account code", func() {
        Pattern("^[0-9]{4,10}$")
    })
    Field(3, "type", String, "Account type", func() {
        Enum("ASSET", "LIABILITY", "EQUITY", "REVENUE", "EXPENSE")
    })
    Field(4, "parent_id", String, "Parent account ID", UUID)
    Field(5, "description", String, "Account description")
    
    Required("name", "code", "type")
})

// Account response
var AccountResult = Type("AccountResult", func() {
    Field(1, "id", String, "Account ID", UUID)
    Field(2, "name", String, "Account name")
    Field(3, "code", String, "Account code")
    Field(4, "type", String, "Account type")
    Field(5, "parent_id", String, "Parent account ID", UUID)
    Field(6, "is_active", Boolean, "Active status")
    Field(7, "created_at", String, "Creation timestamp", Timestamp)
    Field(8, "updated_at", String, "Last update timestamp", Timestamp)
    
    Required("id", "name", "code", "type", "is_active")
})

// List accounts response (with pagination)
var AccountListResult = Type("AccountListResult", func() {
    Field(1, "data", CollectionOf(AccountResult), "Accounts")
    Field(2, "pagination", PaginationResponse, "Pagination metadata")
    Field(3, "actions", ArrayOf(String), "Available actions")
    
    Required("data", "pagination", "actions")
})
```

#### Pattern 4: Extended Metadata Types (for DataTables)

```go
// @internal/api/design/metadata.go
package design

import . "goa.design/goa/v3/dsl"

// DataTable metadata (for UI integration)
var DataTableMetadata = Type("DataTableMetadata", func() {
    Field(1, "columns", ArrayOf(DataTableColumn), "Column definitions")
    Field(2, "actions", ArrayOf(DataTableAction), "Available actions")
    Field(3, "filters", ArrayOf(DataTableFilter), "Available filters")
    Field(4, "sort_by", String, "Current sort column")
    Field(5, "sort_order", String, "Current sort order", func() {
        Enum("asc", "desc")
    })
    
    Required("columns", "actions")
})

var DataTableColumn = Type("DataTableColumn", func() {
    Field(1, "key", String, "Column key")
    Field(2, "label", String, "Column label")
    Field(3, "type", String, "Data type", func() {
        Enum("text", "number", "date", "boolean", "currency", "enum")
    })
    Field(4, "sortable", Boolean, "Can sort by this column")
    Field(5, "searchable", Boolean, "Can search in this column")
    Field(6, "width", String, "Column width (CSS)")
    
    Required("key", "label", "type")
})

var DataTableAction = Type("DataTableAction", func() {
    Field(1, "key", String, "Action key (e.g., 'edit', 'delete')")
    Field(2, "label", String, "Action label")
    Field(3, "icon", String, "Icon name")
    Field(4, "variant", String, "Button variant", func() {
        Enum("primary", "secondary", "danger", "success")
    })
    Field(5, "requires_confirmation", Boolean, "Requires user confirmation")
    Field(6, "bulk_action", Boolean, "Can apply to multiple items")
    
    Required("key", "label")
})

var DataTableFilter = Type("DataTableFilter", func() {
    Field(1, "key", String, "Filter key")
    Field(2, "label", String, "Filter label")
    Field(3, "type", String, "Filter type", func() {
        Enum("text", "select", "date_range", "boolean")
    })
    Field(4, "options", ArrayOf(Any), "Filter options (for select)")
    
    Required("key", "label", "type")
})
```

### 5.3 Generation Commands

```bash
# Generate types from all design files
$ goa gen awo/internal/api/design -o internal/api

# Output structure (all in one package):
# internal/api/gen/
# └── types.go  # All types generated here
```



---


## 4. Handler Architecture & Responsibilities

### 4.1 What Handlers Do (and Don't Do)

**Handler Responsibilities** (The DO list):
1. ✅ Accept HTTP requests
2. ✅ Parse request body/parameters
3. ✅ Validate input format (is it JSON? are required fields present?)
4. ✅ Extract user context (who is making this request?)
5. ✅ Call appropriate service
6. ✅ Format response (JSON or HTML)
7. ✅ Return HTTP status code (200 OK, 404 Not Found, etc.)
8. ✅ Handle errors gracefully

**Handler Responsibilities** (The DON'T list):
1. ❌ Make business decisions (is this account code valid per accounting rules?)
2. ❌ Perform calculations (calculate invoice totals)
3. ❌ Write SQL queries
4. ❌ Call external APIs directly
5. ❌ Contain complex logic (if/else chains)

**Analogy**: Airport security screening
- **Handler** = TSA agent at checkpoint
  - Checks tickets and IDs (validation)
  - Directs to appropriate line (routing)
  - Handles issues (error responses)
- **Service** = Flight operations
  - Determines if flight can depart (business logic)
  - Coordinates crew and aircraft (orchestration)

TSA doesn't decide if weather is good enough to fly—that's operations' job.

### 4.2 Handler Structure & Dependencies

Every handler follows a consistent structure, similar to how buildings have consistent floor plans:

**Core Components**:

1. **Service Dependencies**: The business logic providers
   - Example: `AccountService`, `TransactionService`
   - Injected via constructor (Dependency Injection pattern)

2. **Infrastructure Dependencies**: The supporting utilities
   - Logger: Records what happened (audit trail)
   - Metrics: Tracks performance (how long did request take?)
   - Tracer: Follows request through system (debugging)

3. **Handler Methods**: The operations you can perform
   - Create: Add new records
   - Read: Get existing records
   - Update: Modify records
   - Delete: Remove records
   - Custom actions: Archive, approve, export, etc.

**Why Dependency Injection?**

Think of it like a toolkit. Instead of the handler building its own tools, you hand it the tools it needs:

Without DI (❌ Tightly Coupled):
```
Handler creates its own database connection
→ Can't test handler without real database
→ Can't swap database implementation
→ Can't reuse across environments
```

With DI (✅ Loosely Coupled):
```
Handler receives database connection as parameter
→ Can pass mock database for testing
→ Can swap implementations easily
→ Can configure per environment
```

**Industry Standard**:
- **Spring Framework (Java)**: Pioneered DI in enterprise systems
- **ASP.NET Core**: Built-in DI container
- **NestJS**: DI for Node.js applications
- **Google's Internal Systems**: Extensive use of DI

### 4.3 Standard Handler Pattern

Every handler operation follows a **five-step pattern**:

```
1. RECEIVE & VALIDATE
   └─ Parse request body
   └─ Check required fields
   └─ Validate format (UUID, email, etc.)

2. EXTRACT CONTEXT
   └─ Get current user ID
   └─ Get tenant ID (which company)
   └─ Get request metadata

3. DELEGATE TO SERVICE
   └─ Call business logic
   └─ Pass validated parameters
   └─ Let service make decisions

4. HANDLE RESULT
   └─ If success: map to response type
   └─ If error: convert to user message

5. RETURN RESPONSE
   └─ Choose format (JSON vs HTML)
   └─ Set status code
   └─ Send to client
```

**Why This Pattern?**

**Consistency**: Every operation looks the same
- New developers know what to expect
- Code reviews easier (spot deviations quickly)
- Bugs less likely (proven pattern)

**Error Handling**: Failures caught at each step
- Step 1 fails: Return 400 Bad Request
- Step 2 fails: Return 401 Unauthorized
- Step 3 fails: Return business error (e.g., "Account code exists")

**Observability**: Each step logged/traced
- See exactly where request fails
- Measure time spent in each step
- Debug production issues easily

**Similar Patterns in Industry**:
- **AWS Lambda**: Input validation → Processing → Response
- **Kubernetes Controllers**: Observe → Analyze → Act
- **Redux (React)**: Action → Reducer → State Update

### 4.4 Response Formatting

Handlers return responses in two formats, determined automatically:

**JSON Format** (for API clients):
- Mobile apps
- Third-party integrations
- Automated systems

**HTML Format** (for web UI):
- Browser-based users
- HTMX interactions
- Server-side rendered pages

**How Selection Works**:

```
Client sends "Accept: application/json" header
→ Handler returns JSON

Client sends "Accept: text/html" header
→ Handler returns HTML

No header or HTMX request
→ Handler returns HTML (default)
```

**Why Both Formats?**

Modern systems need multiple interfaces:

**Example - Shopify**:
- Merchants use web UI (HTML)
- Mobile apps use REST API (JSON)
- Partners use GraphQL API (JSON)
- All backed by same handlers

**Benefits for Awo ERP**:
1. **Single endpoint** serves both web and mobile
2. **No duplication** of logic
3. **Consistent** behavior across clients
4. **Easier maintenance** (change once, affects all clients)

---


## 6. Handler Development

### 6.1 Standard Handler Structure

Every handler follows this consistent pattern:

```go
package finance

import (
    "github.com/gofiber/fiber/v2"
    "awo.so/internal/api/gen"
    accountsvc "awo.so/internal/core/finance/account"
    "awo.so/internal/shared/logger"
    "awo.so/internal/shared/metrics"
    "awo.so/internal/shared/tracing"
)

// Handler struct with dependencies
type AccountHandler struct {
    accountService accountsvc.Service
    logger         logger.Logger
    metrics        metrics.MetricsProvider
    tracer         tracing.TracingService
}

// Constructor (dependency injection)
func NewAccountHandler(
    svc accountsvc.Service,
    log logger.Logger,
    met metrics.MetricsProvider,
    trc tracing.TracingService,
) *AccountHandler {
    return &AccountHandler{
        accountService: svc,
        logger:         log,
        metrics:        met,
        tracer:         trc,
    }
}

// Handler method (5-step pattern)
func (h *AccountHandler) Create(c *fiber.Ctx) error {
    // STEP 1: Start observability
    ctx := c.Context()
    ctx, span := h.tracer.StartSpan(ctx, "AccountHandler.Create")
    defer span.End()
    
    // STEP 2: Parse request
    var req gen.CreateAccountPayload
    if err := c.BodyParser(&req); err != nil {
        h.logger.WarnContext(ctx, "Failed to parse request body", logger.Fields{"error": err})
        return h.badRequest(c, "INVALID_REQUEST_BODY", "Failed to parse request body")
    }
    
    // STEP 3: Extract context
    userID := c.Locals("user_id").(string)
    tenantID := c.Locals("tenant_id").(string)
    
    // STEP 4: Call service
    result, err := h.accountService.Create(ctx, accountsvc.CreateParams{
        TenantID:  tenantID,
        Name:      req.Name,
        Type:      req.Type,
        Code:      req.Code,
        CreatedBy: userID,
    })
    if err != nil {
        return h.handleError(c, err)
    }
    
    // STEP 5: Return response (content negotiation automatic)
    response := &gen.AccountResult{
        ID:        result.ID,
        Name:      result.Name,
        Type:      result.Type,
        Code:      result.Code,
        IsActive:  result.IsActive,
        CreatedAt: result.CreatedAt,
        UpdatedAt: result.UpdatedAt,
    }
    
    h.metrics.IncrementCounter("accounts_created", metrics.Fields{"type": result.Type})
    
    return h.respond(c, fiber.StatusCreated, response)
}
```

### 6.2 Standard Helper Methods

Every handler should include these helpers:

```go
// respond handles content-type negotiation automatically
func (h *AccountHandler) respond(c *fiber.Ctx, status int, data interface{}) error {
    accept := c.Get("Accept")
    contentType := c.Get("Content-Type")
    isHTMX := c.Get("HX-Request") == "true"
    
    // Check if client wants JSON
    if strings.Contains(accept, "application/json") || 
       strings.Contains(contentType, "application/json") {
        return c.Status(status).JSON(data)
    }
    
    // Default to HTML for UI/HTMX requests
    return h.renderComponent(c, status, data)
}

// renderComponent renders TemplUI components
func (h *AccountHandler) renderComponent(c *fiber.Ctx, status int, data interface{}) error {
    switch v := data.(type) {
    case *gen.AccountResult:
        component := components.AccountRow(v)
        c.Status(status)
        return component.Render(c.Context(), c.Response().BodyWriter())
        
    case *gen.AccountListResult:
        component := components.AccountTable(v)
        c.Status(status)
        return component.Render(c.Context(), c.Response().BodyWriter())
        
    default:
        return c.Status(status).SendString("")
    }
}

// handleError converts errors to HTTP responses with user-friendly messages
// See Section 7 for full implementation
func (h *AccountHandler) handleError(c *fiber.Ctx, err error) error {
    return handleHTTPError(c, err, h.logger)
}

// badRequest returns 400 with error details
func (h *AccountHandler) badRequest(c *fiber.Ctx, code, message string) error {
    return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
        "error": fiber.Map{
            "code":    code,
            "message": message,
        },
    })
}
```

### 6.3 Handler Patterns

#### Pattern 1: Create (POST)

```go
func (h *AccountHandler) Create(c *fiber.Ctx) error {
    ctx := c.Context()
    
    // Parse payload
    var req gen.CreateAccountPayload
    if err := c.BodyParser(&req); err != nil {
        return h.badRequest(c, "INVALID_REQUEST", "Failed to parse request")
    }
    
    // Extract user context
    userID := c.Locals("user_id").(string)
    
    // Call service
    result, err := h.accountService.Create(ctx, toServiceParams(req, userID))
    if err != nil {
        return h.handleError(c, err)
    }
    
    // Return 201 Created
    return h.respond(c, fiber.StatusCreated, toResultType(result))
}
```

#### Pattern 2: Get Single Resource (GET)

```go
func (h *AccountHandler) Get(c *fiber.Ctx) error {
    ctx := c.Context()
    
    // Extract ID from path
    accountID := c.Params("id")
    if accountID == "" {
        return h.badRequest(c, "MISSING_ID", "Account ID is required")
    }
    
    // Call service
    result, err := h.accountService.Get(ctx, accountID)
    if err != nil {
        return h.handleError(c, err)
    }
    
    // Return 200 OK
    return h.respond(c, fiber.StatusOK, toResultType(result))
}
```

#### Pattern 3: List with Pagination (GET)

```go
func (h *AccountHandler) List(c *fiber.Ctx) error {
    ctx := c.Context()
    
    // Parse pagination params
    page := c.QueryInt("page", 1)
    perPage := c.QueryInt("per_page", 20)
    
    // Parse filters
    filters := accountsvc.ListFilters{
        Type:   c.Query("type"),
        Status: c.Query("status"),
        Search: c.Query("search"),
    }
    
    // Call service
    accounts, total, err := h.accountService.List(ctx, accountsvc.ListParams{
        Page:    page,
        PerPage: perPage,
        Filters: filters,
    })
    if err != nil {
        return h.handleError(c, err)
    }
    
    // Build response with metadata
    response := &gen.AccountListResult{
        Data: toResultTypes(accounts),
        Pagination: &gen.PaginationResponse{
            Total:      total,
            Page:       page,
            PerPage:    perPage,
            TotalPages: (total + perPage - 1) / perPage,
        },
        Actions: []string{"create", "edit", "delete", "archive"},
    }
    
    // Return 200 OK
    return h.respond(c, fiber.StatusOK, response)
}
```

#### Pattern 4: Update (PUT/PATCH)

```go
func (h *AccountHandler) Update(c *fiber.Ctx) error {
    ctx := c.Context()
    
    // Extract ID
    accountID := c.Params("id")
    
    // Parse payload
    var req gen.UpdateAccountPayload
    if err := c.BodyParser(&req); err != nil {
        return h.badRequest(c, "INVALID_REQUEST", "Failed to parse request")
    }
    
    // Call service
    result, err := h.accountService.Update(ctx, accountID, toServiceParams(req))
    if err != nil {
        return h.handleError(c, err)
    }
    
    // Return 200 OK
    return h.respond(c, fiber.StatusOK, toResultType(result))
}
```

#### Pattern 5: Delete (DELETE)

```go
func (h *AccountHandler) Delete(c *fiber.Ctx) error {
    ctx := c.Context()
    
    // Extract ID
    accountID := c.Params("id")
    
    // Call service
    if err := h.accountService.Delete(ctx, accountID); err != nil {
        return h.handleError(c, err)
    }
    
    // Return 204 No Content
    return c.SendStatus(fiber.StatusNoContent)
}
```

#### Pattern 6: Custom Actions (POST)

```go
func (h *AccountHandler) Archive(c *fiber.Ctx) error {
    ctx := c.Context()
    
    // Extract ID
    accountID := c.Params("id")
    userID := c.Locals("user_id").(string)
    
    // Call service
    result, err := h.accountService.Archive(ctx, accountID, userID)
    if err != nil {
        return h.handleError(c, err)
    }
    
    // Return 200 OK
    return h.respond(c, fiber.StatusOK, toResultType(result))
}
```



---


## 6. Request Flow & Processing Pipeline

### 6.1 Complete Request Journey

Understanding the full path a request takes helps troubleshoot issues and optimize performance.

**Step 1: Client Sends Request**

```
User clicks "Create Account" button
↓
Frontend sends:
POST /api/v1/accounts
Headers:
  Authorization: Bearer jwt_token_here
  Content-Type: application/json
Body:
  { "name": "Cash", "code": "1001", "type": "ASSET" }
```

**Step 2: Request Hits Server**

Web framework (Fiber) receives request and routes to handler:
- Matches URL pattern: `/api/v1/accounts`
- Identifies HTTP method: `POST`
- Selects handler: `AccountHandler.Create`

**Step 3: Middleware Chain**

Before handler runs, request passes through **middleware**—like a series of checkpoints:

```
Request
  ↓
[1] Recovery Middleware
    └─ Catches panics (crashes)
    └─ Prevents server from dying
  ↓
[2] Request ID Middleware
    └─ Assigns unique ID to request
    └─ Used for tracing in logs
  ↓
[3] Authentication Middleware
    └─ Validates JWT token
    └─ Extracts user_id and tenant_id
    └─ Returns 401 if invalid
  ↓
[4] Tenant Middleware
    └─ Sets database tenant context (RLS)
    └─ Ensures data isolation
  ↓
[5] Logging Middleware
    └─ Records request details
    └─ Starts performance timer
  ↓
Handler (AccountHandler.Create)
```

**Why Middleware Chain?**

**Separation of Concerns**: Each middleware does one job
- Recovery = error handling
- Auth = security
- Logging = observability

**Reusability**: Same middleware for all handlers
- Don't duplicate auth logic 200 times
- Change authentication? Update one place

**Order Matters**: Must run in specific sequence
- Recovery first (catches all errors below)
- Auth before tenant (need user to determine tenant)
- Logging after auth (log authenticated user)

**Industry Examples**:
- **Express.js**: Middleware pattern originated here
- **Django**: Middleware for auth, sessions, caching
- **ASP.NET Core**: Pipeline with middleware components
- **Kong API Gateway**: Plugin-based middleware

**Step 4: Handler Execution**

Handler receives validated, enriched request:

```
1. Parse request body
   └─ Extract JSON fields
   └─ Check required fields present
   └─ Validate formats

2. Extract context
   └─ Get user_id (from auth middleware)
   └─ Get tenant_id (from auth middleware)
   └─ Get request_id (from request ID middleware)

3. Call service
   └─ Pass: tenant_id, user_id, account data
   └─ Service applies business rules
   └─ Service performs database operations

4. Handle response
   └─ Success: Map to response type
   └─ Error: Convert to user-friendly message

5. Return HTTP response
   └─ Status code (201 Created)
   └─ Body (JSON or HTML)
```

**Step 5: Response Journey Back**

```
Handler returns response
  ↓
Logging Middleware
  └─ Records response status
  └─ Records response time
  └─ Sends metrics
  ↓
Fiber Framework
  └─ Serializes response body
  └─ Sets HTTP headers
  └─ Compresses if needed
  ↓
Network Stack
  └─ TCP/IP transmission
  ↓
Client receives response
  └─ Frontend processes result
  └─ UI updates
```

### 6.2 Error Handling Flow

Errors can occur at any step. Each is handled differently:

**Middleware Errors**:
```
Auth Middleware fails (invalid token)
→ Returns 401 Unauthorized immediately
→ Handler never runs
→ User sees "Please log in again"
```

**Validation Errors**:
```
Handler detects invalid format
→ Returns 400 Bad Request
→ Service never called
→ User sees "Account code must be 4-10 digits"
```

**Business Logic Errors**:
```
Service finds account code exists
→ Returns BusinessError
→ Handler converts to 409 Conflict
→ User sees "Account code 1001 already exists"
```

**System Errors**:
```
Database connection lost
→ Returns 503 Service Unavailable
→ User sees "System temporarily unavailable"
→ Alert sent to operations team
```

### 6.3 Observability (Logging, Metrics, Tracing)

Every request is fully observable—we can see what happened and how long it took:

**Logging**: What happened?
```
[INFO] Request started: POST /api/v1/accounts
[INFO] User authenticated: user_123 (tenant_456)
[INFO] Account service called: CreateAccount
[INFO] Database query executed: INSERT INTO accounts
[INFO] Response sent: 201 Created (125ms)
```

**Metrics**: How many? How fast?
```
accounts_created_total: 1247
account_creation_duration_seconds: 0.125 (average)
account_creation_errors_total: 3 (last hour)
```

**Tracing**: Where did time go?
```
Total request time: 125ms
├─ Authentication: 5ms
├─ Validation: 2ms
├─ Service layer: 115ms
│  ├─ Business logic: 5ms
│  └─ Database query: 110ms (BOTTLENECK!)
└─ Response formatting: 3ms
```

**Why This Matters**:

**Production Issues**: When something breaks, we know:
- Which user affected
- What they tried to do
- Where it failed
- How long before failure

**Performance Optimization**: We can see:
- Slow database queries
- Heavy API endpoints
- Peak usage times

**Business Intelligence**: We track:
- Most-used features
- Error rates by tenant
- User behavior patterns

**Industry Standard**:
- **Google**: Dapper tracing system (distributed tracing)
- **Netflix**: Atlas metrics + Zipkin tracing
- **Uber**: Jaeger tracing across microservices
- **DataDog/New Relic**: Commercial observability platforms

---



---


## 7. Error Handling Philosophy

### 7.1 Two Audiences for Errors

**The Core Problem**: Errors serve two very different audiences with conflicting needs.

**Audience 1: Developers & Operators**
- Need technical details (stack traces, database errors)
- Need context (what was system state?)
- Need to debug and fix issues
- Language: Technical jargon acceptable

**Audience 2: End Users**
- Need to understand what happened
- Need to know what to do next
- Need reassurance system is working
- Language: Plain, friendly language required

**Bad Example** (Technical error shown to user):
```
Error: pq: duplicate key value violates unique constraint "accounts_code_key"
DETAIL: Key (tenant_id, code)=(456, 1001) already exists.
```

**Good Example** (User-friendly message):
```
The account code "1001" is already in use. Please choose a different code.

Suggestions:
• Try using code 1002-1099
• View existing codes in your chart of accounts
```

### 7.2 Error Categories

Errors fall into distinct categories, each handled differently:

**1. Validation Errors** (400 Bad Request)
- **Cause**: Client sent invalid data
- **User Message**: Specific, actionable
- **Example**: "Account name must be 3-100 characters"
- **Action**: User can fix immediately

**2. Business Rule Errors** (409 Conflict, 422 Unprocessable)
- **Cause**: Request violates business rules
- **User Message**: Explain the rule
- **Example**: "Cannot delete account with child accounts"
- **Action**: User must change approach

**3. Authorization Errors** (403 Forbidden)
- **Cause**: User lacks permission
- **User Message**: Vague (security)
- **Example**: "You don't have permission to delete accounts"
- **Action**: User needs different role

**4. Not Found Errors** (404 Not Found)
- **Cause**: Resource doesn't exist
- **User Message**: Clear, may suggest alternatives
- **Example**: "Account not found. It may have been deleted."
- **Action**: User verifies ID or searches

**5. System Errors** (500 Internal Server Error, 503 Service Unavailable)
- **Cause**: Server-side failure (database down, bug)
- **User Message**: Apologetic, no technical details
- **Example**: "Something went wrong on our end. Please try again shortly."
- **Action**: User waits, support team alerted

### 7.3 Error Structure

Every error contains multiple pieces of information:

**Technical Information** (for logs):
- Error code: `ACCOUNT_CODE_EXISTS`
- Technical message: `duplicate key value violates unique constraint`
- Stack trace: Where in code did it fail?
- Context: What was the account code? Who was the user?

**User Information** (for display):
- User message: `The account code "1001" is already in use`
- Suggestions: `Try using a different code`, `Check chart of accounts`
- Severity: Warning (not critical system failure)

**Metadata**:
- HTTP status code: 409 Conflict
- Category: Business Rule violation
- Timestamp: When did it occur?
- Request ID: Which request failed?

### 7.4 Error Response Formats

**JSON Format** (API clients):
```json
{
  "error": {
    "code": "ACCOUNT_CODE_EXISTS",
    "message": "The account code '1001' is already in use. Please choose a different code.",
    "category": "business_rule",
    "suggestions": [
      "Try using a different account code",
      "Check your chart of accounts for existing codes"
    ],
    "details": {
      "account_code": "1001",
      "tenant_id": "tenant_456"
    }
  },
  "request_id": "req_789xyz"
}
```

**HTML Format** (Web UI):
```html
<div class="alert alert-warning">
  <icon>⚠️</icon>
  <h4>Cannot Use This Account Code</h4>
  <p>The account code '1001' is already in use. Please choose a different code.</p>
  <ul>
    <li>Try using a different account code</li>
    <li>Check your chart of accounts for existing codes</li>
  </ul>
</div>
```

### 7.5 Error Handling Best Practices

**1. Never Expose Technical Details to Users**

Bad:
```
SQL Error: INSERT INTO accounts VALUES (...) 
ERROR: duplicate key value on index accounts_pkey
```

Good:
```
This account code already exists. Please choose a different one.
```

**2. Always Provide Suggestions**

Bad:
```
Invalid account type.
```

Good:
```
Invalid account type. Please choose one of: Asset, Liability, Equity, Revenue, or Expense.
```

**3. Be Specific When Possible**

Bad:
```
Bad request.
```

Good:
```
Account code must be 4-10 digits. You entered 3 digits.
```

**4. Log Full Context**

Even though users see friendly messages, logs contain:
- Full error stack trace
- Database state at time of error
- User actions leading to error
- System metrics at failure point

**5. Test Both Formats**

Verify errors display correctly in:
- JSON API responses (mobile apps, integrations)
- HTML UI responses (web browsers)
- Error tracking systems (DataDog, Sentry)

### 7.6 Industry Examples

**Stripe API**:
- Detailed error codes (`card_declined`, `insufficient_funds`)
- Clear user messages ("Your card was declined")
- Actionable suggestions ("Try another payment method")
- Full technical details in developer dashboard

**Shopify API**:
- Categorized errors (validation, rate_limit, server)
- HTTP status codes match error type
- Documentation links in error responses

**GitHub API**:
- Structured error responses with detailed context
- Rate limit information included
- Link to relevant documentation

**Slack API**:
- User-friendly error messages
- Warning vs error distinction
- Retry-after headers for rate limits

---


## 7. Error Handling & User Messages

### 7.1 Error Philosophy

**Errors must serve two audiences:**
1. **Developers** - Need technical details, stack traces, error codes
2. **End Users** - Need clear, actionable, human-readable messages

**This is critical for UI integration** - users see user messages, not technical jargon.

### 7.2 Error Structure

```go
// @internal/shared/errors/types.go
package errors

type BusinessError struct {
    Code        string                 `json:"code"`         // Technical code (e.g., "ACCOUNT_NOT_FOUND")
    Message     string                 `json:"message"`      // Technical message (for logs)
    UserMessage string                 `json:"user_message"` // Human-readable message for UI
    HTTPStatus  int                    `json:"-"`
    Category    ErrorCategory          `json:"category"`
    Details     map[string]interface{} `json:"details,omitempty"`
    Suggestions []string               `json:"suggestions,omitempty"`
}

type ErrorCategory string

const (
    CategoryValidation ErrorCategory = "validation"
    CategoryBusiness   ErrorCategory = "business"
    CategorySystem     ErrorCategory = "system"
    CategorySecurity   ErrorCategory = "security"
)
```

### 7.3 Creating Domain Errors with User Messages

```go
// @internal/core/finance/account/errors.go
package account

import (
    "fmt"
    "net/http"
    "awo.so/internal/shared/errors"
)

// Error codes
const (
    CodeAccountNotFound      = "ACCOUNT_NOT_FOUND"
    CodeAccountCodeExists    = "ACCOUNT_CODE_EXISTS"
    CodeAccountHasChildren   = "ACCOUNT_HAS_CHILDREN"
    CodeInvalidAccountCode   = "INVALID_ACCOUNT_CODE"
)

// NewAccountNotFoundError - user-friendly not found error
func NewAccountNotFoundError(accountID string) *errors.BusinessError {
    return errors.NewBusinessError(
        CodeAccountNotFound,
        "Account not found", // Technical message (for logs)
    ).WithHTTPStatus(http.StatusNotFound).
      WithCategory(errors.CategoryBusiness).
      WithUserMessage("We couldn't find this account. It may have been deleted or you don't have permission to view it."). // For UI
      WithDetail("account_id", accountID).
      WithSuggestion("Check that you're using the correct account ID").
      WithSuggestion("Ask your administrator if you need access to this account")
}

// NewAccountCodeExistsError - user-friendly conflict error
func NewAccountCodeExistsError(code string) *errors.BusinessError {
    return errors.NewBusinessError(
        CodeAccountCodeExists,
        "Account code already exists", // Technical
    ).WithHTTPStatus(http.StatusConflict).
      WithCategory(errors.CategoryValidation).
      WithUserMessage(fmt.Sprintf("The account code '%s' is already in use. Please choose a different code.", code)). // User-friendly
      WithDetail("code", code).
      WithSuggestion("Try using a different account code").
      WithSuggestion("Check your chart of accounts for existing codes")
}

// NewAccountHasChildrenError - deletion prevention with clear message
func NewAccountHasChildrenError(accountID string, childCount int) *errors.BusinessError {
    return errors.NewBusinessError(
        CodeAccountHasChildren,
        "Cannot delete account with child accounts", // Technical
    ).WithHTTPStatus(http.StatusConflict).
      WithCategory(errors.CategoryBusiness).
      WithUserMessage(fmt.Sprintf("This account cannot be deleted because it has %d child account(s). Please delete or move the child accounts first.", childCount)). // User-friendly with count
      WithDetail("account_id", accountID).
      WithDetail("child_count", childCount).
      WithSuggestion("Delete or move all child accounts first").
      WithSuggestion("Archive the account instead of deleting it")
}

// NewInvalidAccountCodeError - validation error with clear reason
func NewInvalidAccountCodeError(code string, reason string) *errors.BusinessError {
    return errors.NewBusinessError(
        CodeInvalidAccountCode,
        "Invalid account code", // Technical
    ).WithHTTPStatus(http.StatusBadRequest).
      WithCategory(errors.CategoryValidation).
      WithUserMessage(fmt.Sprintf("The account code '%s' is not valid: %s", code, reason)). // User-friendly with reason
      WithDetail("code", code).
      WithDetail("reason", reason).
      WithSuggestion("Account codes must be 4-10 digits").
      WithSuggestion("Use only numbers in account codes")
}
```

### 7.4 Handler Error Handling

```go
// @internal/api/handlers/finance/error_handler.go
package finance

import (
    "strings"
    "github.com/gofiber/fiber/v2"
    accounterrors "awo.so/internal/core/finance/account"
    "awo.so/internal/shared/errors"
    "awo.so/internal/shared/logger"
    "awo.so/web/components"
)

// handleError converts domain errors to HTTP responses
func (h *AccountHandler) handleError(c *fiber.Ctx, err error) error {
    ctx := c.Context()
    
    // Log the error (technical details)
    h.logger.ErrorContext(ctx, "Handler error",
        logger.Fields{
            "error":  err,
            "path":   c.Path(),
            "method": c.Method(),
        })
    
    // Check if it's a BusinessError (already has HTTP status and user message)
    if be, ok := err.(*errors.BusinessError); ok {
        return h.respondWithError(c, be)
    }
    
    // Check specific domain error types
    if accounterrors.IsAccountNotFound(err) {
        be := accounterrors.NewAccountNotFoundError("")
        return h.respondWithError(c, be)
    }
    
    if accounterrors.IsAccountCodeExists(err) {
        be := accounterrors.NewAccountCodeExistsError("")
        return h.respondWithError(c, be)
    }
    
    // Generic internal error (don't expose details to user)
    return h.respondWithError(c, errors.NewBusinessError(
        "INTERNAL_ERROR",
        "An unexpected error occurred", // Technical
    ).WithHTTPStatus(fiber.StatusInternalServerError).
      WithCategory(errors.CategorySystem).
      WithUserMessage("Something went wrong on our end. Please try again in a moment."). // User-friendly
      WithSuggestion("If the problem persists, contact support"))
}

// respondWithError sends error response (JSON or HTML based on Accept header)
func (h *AccountHandler) respondWithError(c *fiber.Ctx, err *errors.BusinessError) error {
    accept := c.Get("Accept")
    
    // JSON response (for API clients)
    if strings.Contains(accept, "application/json") {
        return c.Status(err.HTTPStatus).JSON(fiber.Map{
            "error": fiber.Map{
                "code":         err.Code,
                "message":      err.UserMessage, // Use user-friendly message
                "details":      err.Details,
                "suggestions":  err.Suggestions,
                "category":     err.Category,
            },
        })
    }
    
    // HTML response (for UI via TemplUI)
    component := components.ErrorMessage(&components.ErrorProps{
        Title:       getErrorTitle(err.Category),
        Message:     err.UserMessage, // User-friendly message displayed in UI
        Suggestions: err.Suggestions,
        Severity:    getErrorSeverity(err.HTTPStatus),
    })
    
    c.Status(err.HTTPStatus)
    return component.Render(c.Context(), c.Response().BodyWriter())
}

// Helper: Get friendly error title based on category
func getErrorTitle(category errors.ErrorCategory) string {
    switch category {
    case errors.CategoryValidation:
        return "Invalid Input"
    case errors.CategoryBusiness:
        return "Cannot Complete Action"
    case errors.CategorySecurity:
        return "Access Denied"
    case errors.CategorySystem:
        return "System Error"
    default:
        return "Error"
    }
}

// Helper: Get error severity for UI styling
func getErrorSeverity(httpStatus int) string {
    switch {
    case httpStatus >= 500:
        return "critical"
    case httpStatus >= 400:
        return "warning"
    default:
        return "info"
    }
}
```

### 7.5 Example Error Responses

#### JSON Response (API Client)

```json
{
  "error": {
    "code": "ACCOUNT_CODE_EXISTS",
    "message": "The account code '1001' is already in use. Please choose a different code.",
    "category": "validation",
    "details": {
      "code": "1001"
    },
    "suggestions": [
      "Try using a different account code",
      "Check your chart of accounts for existing codes"
    ]
  }
}
```

#### HTML Response (UI via TemplUI)

```html
<!-- Rendered by components.ErrorMessage -->
<div class="alert alert-warning" role="alert">
  <div class="alert-icon">
    <svg>...</svg>
  </div>
  <div class="alert-content">
    <h4 class="alert-title">Invalid Input</h4>
    <p class="alert-message">
      The account code '1001' is already in use. Please choose a different code.
    </p>
    <ul class="alert-suggestions">
      <li>Try using a different account code</li>
      <li>Check your chart of accounts for existing codes</li>
    </ul>
  </div>
</div>
```

### 7.6 Best Practices

1. **Always provide user messages** - Technical messages are for logs, user messages for UI
2. **Be specific** - "Account code 1001 already exists" > "Invalid input"
3. **Suggest solutions** - Tell users what to do next
4. **Use appropriate severity** - Validation errors are warnings, system errors are critical
5. **Log technical details** - Full error details go to logs, not to users
6. **Test both formats** - Verify JSON and HTML error responses




---

# PART III: ADVANCED INTEGRATION TOPICS

---

## 8. Multi-Tenancy & Data Isolation

### 8.1 What is Multi-Tenancy?

**Definition**: One application instance serves multiple customers (tenants), with complete data separation.

**Analogy**: Apartment building
- **Multi-Tenant**: One building, many apartments
  - Shared infrastructure (plumbing, electricity)
  - Private spaces (each tenant can't access others' apartments)
  - Efficient resource use (one building = many customers)

- **Single-Tenant**: Individual houses
  - Each customer gets own house (expensive)
  - Complete isolation (secure but wasteful)
  - Harder to maintain (must visit each house separately)

**Real-World Examples**:

**Salesforce**: Largest multi-tenant CRM
- One codebase serves 150,000+ companies
- Each company's data completely isolated
- Shared infrastructure (cost effective)

**Slack**: Team communication platform
- Millions of workspaces on shared infrastructure
- Your Slack messages never leak to other workspaces
- But all run on same servers

**Shopify**: E-commerce platform
- 4+ million stores on same platform
- Each store's products/orders completely separate
- Shared payment processing, hosting, updates

### 8.2 Why Multi-Tenancy for Awo ERP?

**Economic Benefits**:
- **Cost Efficiency**: One server handles 100 tenants (vs 100 servers for single-tenant)
- **Maintenance**: Deploy update once, all tenants get it
- **Scaling**: Add new tenant = add database row (vs provision new infrastructure)

**Operational Benefits**:
- **Consistency**: All tenants on same version (no "your version doesn't have that feature")
- **Monitoring**: Single dashboard for all tenants
- **Backup**: One backup process, all data protected

**Customer Benefits**:
- **Faster Onboarding**: New tenant ready in minutes (vs days for infrastructure)
- **Lower Cost**: Shared infrastructure = lower subscription price
- **Automatic Updates**: Always on latest version

### 8.3 Row-Level Security (RLS)

**The Challenge**: How do we prevent Tenant A from seeing Tenant B's data?

**Bad Solutions**:

**Option 1: Application-Level Filtering** (❌ Error-Prone)
```
Every query must include: WHERE tenant_id = current_tenant
Developer forgets once → Data leak
```

**Option 2: Separate Databases** (❌ Expensive)
```
Each tenant gets own database
100 tenants = 100 databases = 100× maintenance
```

**Good Solution: Row-Level Security (RLS)** (✅ Database-Enforced)

PostgreSQL automatically filters rows:
```
SET app.tenant_id = '456';
SELECT * FROM accounts;
→ Database automatically adds: WHERE tenant_id = '456'
→ Impossible to see other tenants' data
```

**How RLS Works**:

1. **Middleware Sets Context** (once per request):
   ```
   Request arrives → Extract tenant_id → Set in database session
   ```

2. **Database Enforces Policy** (automatically):
   ```
   ALL queries to accounts table
   → Automatically filter WHERE tenant_id = session_tenant_id
   → Handler/service code doesn't need to remember
   ```

3. **Guaranteed Isolation**:
   ```
   Even if developer writes: SELECT * FROM accounts
   → Database only returns current tenant's accounts
   → Cannot bypass (database-level enforcement)
   ```

**Why This is Secure**:

**Defense in Depth**:
- **Layer 1**: Application checks tenant_id
- **Layer 2**: Database enforces with RLS
- **Layer 3**: Audit logs track all access

**Fail-Safe**: If application has bug, database still prevents leak

**Compliance**: Auditors can verify isolation at database level

### 8.4 Implementation Strategy

**One-Time Setup** (Middleware):
```
Request Arrives
  ↓
Auth Middleware extracts tenant_id from JWT
  ↓
Tenant Middleware calls: database.SetTenantContext(tenant_id)
  ↓
Database session now locked to tenant_id
  ↓
All subsequent queries auto-filtered
```

**Handler Code** (Simplified):
```
Handler doesn't think about tenancy at all!

Handler just calls: accountService.GetAll()
Service just calls: database.QueryAccounts()
Database automatically filters by tenant_id

Developer cannot accidentally query wrong tenant's data.
```

### 8.5 Industry Adoption

**PostgreSQL RLS**: Used by major companies
- **Supabase**: RLS for all multi-tenant apps
- **Salesforce**: Database-level isolation (proprietary)
- **AWS RDS**: Supports RLS for PostgreSQL instances

**Alternative Approaches**:

| Approach | Used By | Pros | Cons |
|----------|---------|------|------|
| Separate Databases | **Heroku**, **MongoDB Atlas** | Complete isolation | Expensive, hard to maintain |
| Schema per Tenant | **Microsoft Dynamics** | Good isolation | Still expensive, complex migrations |
| Application Filtering | **Early SaaS apps** | Simple initially | Error-prone, security risk |
| Row-Level Security | **Modern apps**, **Awo ERP** | Secure, efficient | Requires PostgreSQL 9.5+ |

**Why Awo ERP Chose RLS**:
- **Security**: Database-enforced (not relying on developer memory)
- **Performance**: Single database = simpler queries
- **Cost**: Shared infrastructure = lower operational costs
- **Scalability**: Can handle thousands of tenants

---


## 8. Database & RLS Integration

### 8.1 Store Interface

All database operations use the **Store interface** defined in `@internal/db/sqlc/store.go`:

```go
// Store interface (extended from SQLC)
type Store interface {
    Querier  // SQLC-generated interface
    
    // Modern tenant context methods (✅ USE THESE)
    SetTenantContextFromCtx(ctx context.Context) error
    WithTenantFromCtx(ctx context.Context, fn func(context.Context, Store) error) error
    BeginTxWithTenantFromCtx(ctx context.Context) (pgx.Tx, Store, error)
    
    // General transaction methods
    WithTx(ctx context.Context, fn func(context.Context, Store) error) error
    WithTxOptions(ctx context.Context, opts pgx.TxOptions, fn func(context.Context, Store) error) error
    
    // Connection management
    Close()
    GetPool() *pgxpool.Pool
    HealthCheck(ctx context.Context) error
}
```

### 8.2 RLS Setup in Middleware

**Tenant middleware sets RLS once** - handlers never touch database directly:

```go
// @internal/platform/middleware/tenant.go
package middleware

import (
    "github.com/gofiber/fiber/v2"
    "awo.so/internal/db/sqlc"
)

func TenantMiddleware(store db.Store) fiber.Handler {
    return func(c *fiber.Ctx) error {
        ctx := c.Context()
        
        // Extract tenant ID from auth middleware
        tenantID, ok := c.Locals("tenant_id").(string)
        if !ok || tenantID == "" {
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
                "error": "Missing tenant context",
            })
        }
        
        // Set tenant context for RLS (via Store interface)
        // This sets app.tenant_id session variable in PostgreSQL
        if err := store.SetTenantContextFromCtx(ctx); err != nil {
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                "error": "Failed to set tenant context",
            })
        }
        
        // Continue to handler
        return c.Next()
    }
}
```

### 8.3 Using Store in Services

Services receive Store via dependency injection:

```go
// @internal/core/finance/account/service.go
package account

import (
    "context"
    "awo.so/internal/db/sqlc"
)

type accountService struct {
    store db.Store
    // ... other deps
}

func NewService(store db.Store) Service {
    return &accountService{
        store: store,
    }
}

// Simple query (RLS already set by middleware)
func (s *accountService) Get(ctx context.Context, id string) (*Account, error) {
    // RLS is already active, no need to set tenant context
    account, err := s.store.GetAccount(ctx, id)
    if err != nil {
        return nil, err
    }
    return toDomain(account), nil
}

// Query with transaction
func (s *accountService) CreateWithBalance(ctx context.Context, params CreateParams) (*Account, error) {
    var result *Account
    
    // Use Store.WithTenantFromCtx for tenant-scoped transaction
    err := s.store.WithTenantFromCtx(ctx, func(ctx context.Context, store db.Store) error {
        // Create account
        account, err := store.CreateAccount(ctx, db.CreateAccountParams{
            Name: params.Name,
            Code: params.Code,
            Type: params.Type,
        })
        if err != nil {
            return err
        }
        
        // Create initial balance
        _, err = store.CreateAccountBalance(ctx, db.CreateAccountBalanceParams{
            AccountID: account.ID,
            Balance:   params.InitialBalance,
        })
        if err != nil {
            return err
        }
        
        result = toDomain(account)
        return nil
    })
    
    return result, err
}
```

### 8.4 Transaction Patterns

#### Pattern 1: Simple Transaction

```go
func (s *accountService) Update(ctx context.Context, id string, params UpdateParams) (*Account, error) {
    var result *Account
    
    err := s.store.WithTx(ctx, func(ctx context.Context, store db.Store) error {
        // Check if account exists
        existing, err := store.GetAccount(ctx, id)
        if err != nil {
            return err
        }
        
        // Update account
        updated, err := store.UpdateAccount(ctx, db.UpdateAccountParams{
            ID:   id,
            Name: params.Name,
        })
        if err != nil {
            return err
        }
        
        result = toDomain(updated)
        return nil
    })
    
    return result, err
}
```

#### Pattern 2: Transaction with Manual Control

```go
func (s *accountService) ComplexOperation(ctx context.Context, params Params) error {
    // Begin transaction with tenant context
    tx, store, err := s.store.BeginTxWithTenantFromCtx(ctx)
    if err != nil {
        return err
    }
    
    // Always rollback on error
    defer tx.Rollback(ctx)
    
    // Step 1: Create account
    account, err := store.CreateAccount(ctx, createParams)
    if err != nil {
        return err
    }
    
    // Step 2: Create child accounts
    for _, child := range params.Children {
        _, err := store.CreateAccount(ctx, childParams)
        if err != nil {
            return err // Transaction automatically rolls back
        }
    }
    
    // Step 3: Commit
    if err := tx.Commit(ctx); err != nil {
        return err
    }
    
    return nil
}
```

### 8.5 Key Points

1. **RLS is set in middleware** - handlers/services never call `SetTenantContext` directly
2. **Store interface abstracts SQLC** - services use Store, not direct SQLC queries
3. **Tenant context from ctx** - Use `WithTenantFromCtx` to read tenant ID from context
4. **Automatic rollback** - Transactions rollback on error unless explicitly committed
5. **Services receive Store** - Not direct DB connection or SQLC queries



---


## 9. UI Integration Strategy

### 9.1 Server-Side Rendering Philosophy

**Modern Debate**: Where should UI be rendered?

**Option 1: Client-Side (React, Vue, Angular)**:
```
Server sends JSON → Client builds HTML → Client displays UI
```

**Option 2: Server-Side (Traditional)**:
```
Server builds HTML → Client displays HTML directly
```

**Awo ERP Choice**: Server-side rendering with modern tooling (TemplUI + HTMX)

### 9.2 Why Server-Side for ERP?

**ERP-Specific Considerations**:

**1. Forms Over Features**:
- ERP is mostly data entry forms, tables, reports
- Rich animations less important than data accuracy
- Users value speed and reliability over flashy UI

**2. Security**:
- Server controls business rules
- Client can't bypass validations
- Sensitive calculations stay server-side

**3. Consistency**:
- Server and client share same types
- No API contract mismatches
- One codebase = easier maintenance

**4. Performance**:
- Large datasets (10,000 row tables)
- Server can filter/paginate efficiently
- Client doesn't need to download entire dataset

**Comparison**:

| Aspect | Client-Side SPA | Server-Side (Awo ERP) |
|--------|-----------------|------------------------|
| Initial Load | Slow (large JS bundle) | Fast (just HTML) |
| Subsequent Interactions | Fast | Fast (HTMX partial updates) |
| SEO | Complex (needs SSR) | Built-in |
| Development Complexity | High (2 codebases) | Lower (1 codebase) |
| Security | Client can manipulate | Server-controlled |
| Offline Support | Possible | Limited |

### 9.3 TemplUI Component System

**What is TemplUI?**

Think of it like LEGO blocks for UI:
- Pre-built components (buttons, forms, tables)
- Assemble them to build pages
- Consistent look and feel
- Maintained centrally

**Component Philosophy**:

Instead of writing HTML:
```html
<!-- Traditional: Manually write HTML -->
<table class="table table-striped">
  <thead>
    <tr>
      <th>Code</th>
      <th>Name</th>
      <th>Type</th>
    </tr>
  </thead>
  <tbody>
    <!-- Repeat for each row... -->
  </tbody>
</table>
```

Call components:
```go
// Modern: Call component function
DataTable({
  columns: ["Code", "Name", "Type"],
  data: accounts,
  actions: ["create", "edit", "delete"]
})
```

**Benefits**:

**1. Consistency**:
- All tables look identical
- Same pagination controls everywhere
- Uniform button styles

**2. Maintainability**:
- Change table styling once, updates everywhere
- Add new feature (e.g., bulk actions) to component, all tables get it

**3. Type Safety**:
- Compiler verifies component props
- Can't pass wrong data type
- Catches errors at build time

**4. Backend-Frontend Alignment**:
- Same type definitions for API and UI
- Update API response, UI automatically adjusts
- No synchronization issues

### 9.4 HTMX Integration

**What is HTMX?**

Allows HTML to make HTTP requests without JavaScript:

```html
<!-- Button that updates account when clicked -->
<button hx-post="/api/v1/accounts/123/archive" 
        hx-target="#account-123"
        hx-swap="outerHTML">
  Archive
</button>
```

**How It Works**:
1. User clicks "Archive" button
2. HTMX sends: `POST /api/v1/accounts/123/archive`
3. Server processes request
4. Server returns HTML fragment (just the updated row)
5. HTMX replaces old row with new row
6. User sees updated UI (no page refresh)

**Why HTMX for Awo ERP?**:

**1. Progressive Enhancement**:
- Works without JavaScript (accessibility)
- Enhanced with JavaScript (better UX)
- Graceful degradation

**2. Minimal JavaScript**:
- HTMX library: 14KB (vs React: 140KB)
- No build step required
- Faster page loads

**3. Server-Controlled**:
- Server decides what HTML to return
- Client can't manipulate business logic
- Security maintained

**4. Natural Fit for ERP**:
- Most actions are: click → server processes → update display
- Complex client state not needed
- Reduced frontend complexity

**Industry Trend**:

**Companies Using Server-Side Rendering + HTMX**:
- **Basecamp** (Hey.com): Hotwire (similar to HTMX)
- **GitHub**: Server-rendered HTML with minimal JS
- **Stack Overflow**: Mostly server-side
- **GOV.UK**: Progressive enhancement approach

**Why the Shift?**

After years of complex SPAs, industry recognizing:
- Most apps don't need rich client state
- Simpler stacks = faster development
- Server-side easier to secure
- Better performance for typical use cases

---


## 9. UI Integration with TemplUI

### 9.1 TemplUI Overview

TemplUI (templui.io) generates Go templ components. **You don't write HTML** - you call generated components.

### 9.2 Component Structure

```
web/components/
├── datatable.templ       # Generated DataTable component
├── account_row.templ     # Account row for tables
├── account_form.templ    # Account creation/edit form
├── error_message.templ   # Error display
└── button.templ          # Reusable button component
```

### 9.3 Rendering Components in Handlers

```go
// @internal/api/handlers/finance/account_handler.go

import (
    "awo.so/web/components"
)

// renderComponent renders TemplUI components based on data type
func (h *AccountHandler) renderComponent(c *fiber.Ctx, status int, data interface{}) error {
    switch v := data.(type) {
    case *gen.AccountResult:
        // Render single account row
        component := components.AccountRow(&components.AccountRowProps{
            ID:        v.ID,
            Name:      v.Name,
            Code:      v.Code,
            Type:      v.Type,
            IsActive:  v.IsActive,
            CreatedAt: v.CreatedAt,
        })
        c.Status(status)
        return component.Render(c.Context(), c.Response().BodyWriter())
        
    case *gen.AccountListResult:
        // Render full table
        component := components.DataTable(&components.DataTableProps{
            Columns: []components.Column{
                {Key: "code", Label: "Code", Type: "text", Sortable: true},
                {Key: "name", Label: "Name", Type: "text", Sortable: true, Searchable: true},
                {Key: "type", Label: "Type", Type: "enum", Sortable: true},
                {Key: "created_at", Label: "Created", Type: "date", Sortable: true},
            },
            Rows:       toTableRows(v.Data),
            Pagination: v.Pagination,
            Actions:    toTableActions(v.Actions),
        })
        c.Status(status)
        return component.Render(c.Context(), c.Response().BodyWriter())
        
    default:
        return c.Status(status).SendString("")
    }
}

// Helper: Convert result data to table rows
func toTableRows(accounts []*gen.AccountResult) []map[string]interface{} {
    rows := make([]map[string]interface{}, len(accounts))
    for i, acc := range accounts {
        rows[i] = map[string]interface{}{
            "id":         acc.ID,
            "code":       acc.Code,
            "name":       acc.Name,
            "type":       acc.Type,
            "created_at": acc.CreatedAt,
        }
    }
    return rows
}

// Helper: Convert action strings to table actions
func toTableActions(actions []string) []components.Action {
    actionMap := map[string]components.Action{
        "create": {
            Key:     "create",
            Label:   "New Account",
            Icon:    "plus",
            Variant: "primary",
        },
        "edit": {
            Key:     "edit",
            Label:   "Edit",
            Icon:    "pencil",
            Variant: "secondary",
        },
        "delete": {
            Key:                "delete",
            Label:              "Delete",
            Icon:               "trash",
            Variant:            "danger",
            RequiresConfirmation: true,
        },
        "archive": {
            Key:     "archive",
            Label:   "Archive",
            Icon:    "archive",
            Variant: "secondary",
        },
    }
    
    result := make([]components.Action, 0, len(actions))
    for _, key := range actions {
        if action, ok := actionMap[key]; ok {
            result = append(result, action)
        }
    }
    return result
}
```

### 9.4 HTMX Integration

HTMX requests include `hx-request` header. Detect and respond with HTML fragments:

```go
func (h *AccountHandler) respond(c *fiber.Ctx, status int, data interface{}) error {
    // Check if HTMX request
    isHTMX := c.Get("HX-Request") == "true"
    accept := c.Get("Accept")
    
    // HTMX or HTML request → return component
    if isHTMX || strings.Contains(accept, "text/html") {
        return h.renderComponent(c, status, data)
    }
    
    // Otherwise return JSON
    return c.Status(status).JSON(data)
}
```



---


## 10. Data Presentation & Metadata

### 10.1 Metadata-Driven UI

**Traditional Approach** (❌ Tightly Coupled):
```
Frontend hardcodes table structure:
- columns: ["Code", "Name", "Type"]
- actions: ["create", "edit", "delete"]

Problem: Change backend, must also change frontend separately
Risk: Frontend and backend get out of sync
```

**Metadata Approach** (✅ Loosely Coupled):
```
Backend sends table structure:
{
  "columns": [
    {"key": "code", "label": "Code", "sortable": true},
    {"key": "name", "label": "Name", "searchable": true},
    {"key": "type", "label": "Type", "filterable": true}
  ],
  "actions": ["create", "edit", "delete"],
  "data": [...]
}

Frontend renders based on metadata
```

**Benefits**:

**1. Single Source of Truth**:
- Backend defines table structure
- Frontend just follows instructions
- Cannot get out of sync

**2. Dynamic Configuration**:
- Different users see different columns (permissions)
- Different tenants have different actions (feature flags)
- A/B testing different layouts (metadata changes)

**3. Easier Changes**:
- Add column: Update backend only
- Remove action: Change metadata only
- Reorder columns: Update metadata, no code changes

### 10.2 DataTable Metadata Structure

Every list endpoint returns comprehensive metadata:

**Column Metadata**:
- **Key**: Internal field name (`account_code`)
- **Label**: Display name ("Account Code")
- **Type**: Data type (`text`, `number`, `date`, `currency`, `boolean`, `enum`)
- **Sortable**: Can user sort by this column?
- **Searchable**: Can user search in this column?
- **Width**: Column width for UI

**Action Metadata**:
- **Key**: Action identifier (`create`, `edit`, `delete`, `archive`)
- **Label**: Display text ("New Account", "Edit", "Delete")
- **Icon**: Visual indicator
- **Variant**: Style (`primary`, `secondary`, `danger`)
- **Requires Confirmation**: Should show "Are you sure?" dialog?
- **Bulk Action**: Can apply to multiple rows at once?

**Filter Metadata**:
- **Key**: Filter field name
- **Label**: Filter display name
- **Type**: Filter control type (`text`, `select`, `date_range`, `boolean`)
- **Options**: Available values (for select filters)

**Pagination Metadata**:
- **Total**: Total number of items
- **Page**: Current page number
- **Per Page**: Items per page
- **Total Pages**: Calculated total pages

### 10.3 Why Metadata Matters

**Example Scenario**: Adding "Balance" column to accounts table

**Without Metadata** (Traditional):
1. Update database query (backend)
2. Update API response type (backend)
3. Update frontend table component (frontend)
4. Update mobile app table (mobile team)
5. Test all platforms
6. Coordinate deployment (all must deploy together)

**With Metadata** (Awo ERP):
1. Update metadata to include balance column (backend only)
2. Deploy backend
3. Frontend automatically renders new column (metadata-driven)
4. Mobile app automatically shows new column (metadata-driven)
5. No frontend code changes needed

**Time Saved**: 2 hours → 15 minutes
**Risk Reduced**: 3 teams coordinating → 1 team change

### 10.4 Industry Examples

**Salesforce Lightning Data Service**:
- Metadata defines object fields, relationships, permissions
- UI components built dynamically from metadata
- Admin changes metadata, UI updates automatically

**Microsoft Power Apps**:
- Canvas apps built from data source metadata
- Connectors provide schema, UI adjusts
- Low-code approach powered by metadata

**Shopify Admin API**:
- Returns metafields with resource data
- Frontend renders based on metafield definitions
- Extensible without API changes

**Why This Approach Succeeds**:
- **Reduces coupling** between frontend and backend
- **Enables customization** per tenant/user
- **Accelerates development** (change once vs everywhere)
- **Improves consistency** (one definition used by all clients)

---


## 10. DataTable Patterns & Metadata

### 10.1 Standard DataTable Response

**Every GET list endpoint should return this structure:**

```go
type DataTableResponse struct {
    Data       []interface{}       `json:"data"`        // Actual data rows
    Pagination PaginationMetadata  `json:"pagination"`  // Pagination info
    Metadata   DataTableMetadata   `json:"metadata"`    // Table configuration
    Actions    []ActionMetadata    `json:"actions"`     // Available actions
}
```

### 10.2 Complete List Handler Example

```go
func (h *AccountHandler) List(c *fiber.Ctx) error {
    ctx := c.Context()
    
    // Parse pagination
    page := c.QueryInt("page", 1)
    perPage := c.QueryInt("per_page", 20)
    
    // Parse sorting
    sortBy := c.Query("sort_by", "code")
    sortOrder := c.Query("sort_order", "asc")
    
    // Parse filters
    filters := accountsvc.ListFilters{
        Type:   c.Query("type"),
        Status: c.Query("status"),
        Search: c.Query("search"),
    }
    
    // Call service
    accounts, total, err := h.accountService.List(ctx, accountsvc.ListParams{
        Page:      page,
        PerPage:   perPage,
        SortBy:    sortBy,
        SortOrder: sortOrder,
        Filters:   filters,
    })
    if err != nil {
        return h.handleError(c, err)
    }
    
    // Build comprehensive response with full metadata
    response := &DataTableResponse{
        Data: toDataTableRows(accounts),
        Pagination: PaginationMetadata{
            Total:      total,
            Page:       page,
            PerPage:    perPage,
            TotalPages: (total + perPage - 1) / perPage,
        },
        Metadata: DataTableMetadata{
            Columns: []ColumnMetadata{
                {
                    Key:        "code",
                    Label:      "Account Code",
                    Type:       "text",
                    Sortable:   true,
                    Searchable: false,
                    Width:      "120px",
                },
                {
                    Key:        "name",
                    Label:      "Account Name",
                    Type:       "text",
                    Sortable:   true,
                    Searchable: true,
                    Width:      "300px",
                },
                {
                    Key:        "type",
                    Label:      "Type",
                    Type:       "enum",
                    Sortable:   true,
                    Searchable: false,
                    Width:      "150px",
                },
                {
                    Key:        "balance",
                    Label:      "Balance",
                    Type:       "currency",
                    Sortable:   true,
                    Searchable: false,
                    Width:      "150px",
                },
                {
                    Key:        "is_active",
                    Label:      "Active",
                    Type:       "boolean",
                    Sortable:   true,
                    Searchable: false,
                    Width:      "80px",
                },
                {
                    Key:        "created_at",
                    Label:      "Created",
                    Type:       "date",
                    Sortable:   true,
                    Searchable: false,
                    Width:      "150px",
                },
            },
            Filters: []FilterMetadata{
                {
                    Key:   "type",
                    Label: "Account Type",
                    Type:  "select",
                    Options: []interface{}{
                        map[string]string{"value": "ASSET", "label": "Asset"},
                        map[string]string{"value": "LIABILITY", "label": "Liability"},
                        map[string]string{"value": "EQUITY", "label": "Equity"},
                        map[string]string{"value": "REVENUE", "label": "Revenue"},
                        map[string]string{"value": "EXPENSE", "label": "Expense"},
                    },
                },
                {
                    Key:   "status",
                    Label: "Status",
                    Type:  "select",
                    Options: []interface{}{
                        map[string]string{"value": "active", "label": "Active"},
                        map[string]string{"value": "inactive", "label": "Inactive"},
                        map[string]string{"value": "archived", "label": "Archived"},
                    },
                },
                {
                    Key:   "search",
                    Label: "Search",
                    Type:  "text",
                },
            },
            SortBy:    sortBy,
            SortOrder: sortOrder,
        },
        Actions: []ActionMetadata{
            {
                Key:        "create",
                Label:      "New Account",
                Icon:       "plus",
                Variant:    "primary",
                BulkAction: false,
            },
            {
                Key:        "edit",
                Label:      "Edit",
                Icon:       "pencil",
                Variant:    "secondary",
                BulkAction: false,
            },
            {
                Key:                  "delete",
                Label:                "Delete",
                Icon:                 "trash",
                Variant:              "danger",
                RequiresConfirmation: true,
                BulkAction:           true,
            },
            {
                Key:        "archive",
                Label:      "Archive",
                Icon:       "archive",
                Variant:    "secondary",
                BulkAction: true,
            },
            {
                Key:        "export",
                Label:      "Export",
                Icon:       "download",
                Variant:    "secondary",
                BulkAction: false,
            },
        },
    }
    
    return h.respond(c, fiber.StatusOK, response)
}
```

### 10.3 Frontend Usage

The TemplUI DataTable component automatically consumes this metadata:

```html
<!-- GET /api/v1/accounts returns full DataTable with metadata -->
<div hx-get="/api/v1/accounts" 
     hx-trigger="load"
     hx-target="#account-table"
     hx-swap="innerHTML">
  Loading...
</div>

<div id="account-table">
  <!-- DataTable component renders here with:
       - Columns from metadata.columns
       - Sortable headers
       - Search/filter controls
       - Pagination controls
       - Action buttons (Create, Edit, Delete, Archive, Export)
       - Bulk action checkboxes (for delete/archive)
  -->
</div>
```



---


## 11. Content-Type Negotiation

### 11.1 Strategy

Use **content negotiation** to serve both JSON (API clients) and HTML (UI):

```go
func (h *Handler) respond(c *fiber.Ctx, status int, data interface{}) error {
    // Check request headers
    accept := c.Get("Accept")
    contentType := c.Get("Content-Type")
    isHTMX := c.Get("HX-Request") == "true"
    
    // Decide format
    if strings.Contains(accept, "application/json") || 
       strings.Contains(contentType, "application/json") {
        return c.Status(status).JSON(data)
    }
    
    // Default to HTML for UI/HTMX requests
    return h.renderComponent(c, status, data)
}
```

### 11.2 Example Responses

**Request 1: JSON Client**
```bash
curl -H "Accept: application/json" http://localhost:8080/api/v1/accounts/123
```

**Response:**
```json
{
  "id": "123",
  "name": "Cash",
  "code": "1001",
  "type": "ASSET",
  "is_active": true
}
```

**Request 2: HTMX/UI**
```html
<div hx-get="/api/v1/accounts/123">
  Click to load
</div>
```

**Response:**
```html
<tr id="account-123">
  <td>1001</td>
  <td>Cash</td>
  <td><span class="badge badge-primary">Asset</span></td>
  <td>
    <button hx-delete="/api/v1/accounts/123">Delete</button>
  </td>
</tr>
```



---


## 12. Pagination & Filtering

### 12.1 Query Parameters

Standard query parameters for list endpoints:

```
GET /api/v1/accounts?page=2&per_page=20&sort_by=name&sort_order=asc&type=ASSET&search=cash
```

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `page` | int | 1 | Page number (1-indexed) |
| `per_page` | int | 20 | Items per page (1-100) |
| `sort_by` | string | "created_at" | Column to sort by |
| `sort_order` | string | "desc" | Sort order ("asc" or "desc") |
| `search` | string | "" | Search query |
| `<field>` | string | "" | Filter by field (e.g., `type=ASSET`) |

### 12.2 Parsing in Handler

```go
func (h *AccountHandler) List(c *fiber.Ctx) error {
    // Pagination
    page := c.QueryInt("page", 1)
    perPage := c.QueryInt("per_page", 20)
    
    // Validate limits
    if page < 1 {
        page = 1
    }
    if perPage < 1 || perPage > 100 {
        perPage = 20
    }
    
    // Sorting
    sortBy := c.Query("sort_by", "created_at")
    sortOrder := c.Query("sort_order", "desc")
    
    // Validate sort order
    if sortOrder != "asc" && sortOrder != "desc" {
        sortOrder = "desc"
    }
    
    // Filters
    filters := accountsvc.ListFilters{
        Type:   c.Query("type"),
        Status: c.Query("status"),
        Search: c.Query("search"),
    }
    
    // Call service
    accounts, total, err := h.accountService.List(ctx, accountsvc.ListParams{
        Page:      page,
        PerPage:   perPage,
        SortBy:    sortBy,
        SortOrder: sortOrder,
        Filters:   filters,
    })
    // ...
}
```




---

# PART IV: QUALITY ASSURANCE & OPERATIONS

---

## 11. Testing Strategy

### 11.1 Why Test Handlers Separately?

**Testing Pyramid**:

```
        /\
       /  \      Unit Tests (70%)
      /____\     → Fast, isolated, many
     /      \    Integration Tests (20%)
    /________\   → Medium speed, some dependencies
   /          \  End-to-End Tests (10%)
  /____________\ → Slow, full system, few
```

**Handler Testing Focus**: Unit Tests + Integration Tests

**Unit Tests** (Handlers in Isolation):
- Mock all dependencies (service, database)
- Test handler logic only (parsing, validation, response formatting)
- Fast execution (milliseconds)
- Run on every code change

**Integration Tests** (Handlers + Real Services):
- Real database (test database)
- Real services (business logic)
- Slower execution (seconds)
- Run before deployment

**Why Not Just End-to-End?**

E2E tests are:
- **Slow**: Start browser, navigate UI, wait for responses (minutes)
- **Brittle**: Break when UI changes
- **Hard to debug**: Where did it fail? Handler? Service? Database? UI?

Unit tests are:
- **Fast**: Run hundreds in seconds
- **Focused**: Know exactly what failed
- **Easy to write**: Less setup required

### 11.2 Testing Patterns

**Pattern 1: Mock Service Dependencies**

Test handler without touching database:

```
Test Setup:
1. Create mock service (fake AccountService)
2. Define mock behavior (when Create called, return success)
3. Create handler with mock service
4. Send fake HTTP request to handler
5. Verify response

Benefits:
- No database needed (fast)
- Test error scenarios easily (mock service returns errors)
- Isolate handler logic
```

**Pattern 2: Test Table Approach**

Test multiple scenarios efficiently:

```
Define test cases:
1. Success case (valid input → 201 Created)
2. Validation error (missing field → 400 Bad Request)
3. Business error (duplicate code → 409 Conflict)
4. Auth error (invalid token → 401 Unauthorized)

Run all tests:
- Same setup code
- Different inputs
- Different expected outputs
- Quickly catch regressions
```

**Pattern 3: Verify Response Format**

Test both JSON and HTML responses:

```
Test 1: JSON Response
- Send "Accept: application/json"
- Verify response is valid JSON
- Check status code
- Check response structure

Test 2: HTML Response
- Send "Accept: text/html"
- Verify response is valid HTML
- Check status code
- Verify component rendered correctly
```

### 11.3 What to Test

**DO Test in Handler Tests**:
✅ Request parsing (are fields extracted correctly?)
✅ Validation (are invalid requests rejected?)
✅ Error handling (are errors converted to proper responses?)
✅ Response formatting (JSON vs HTML)
✅ Status codes (201 Created, 400 Bad Request, etc.)
✅ Content-type negotiation

**DON'T Test in Handler Tests**:
❌ Business logic (test in service layer)
❌ Database queries (test in data access layer)
❌ Complex workflows (test in integration tests)

**Why This Separation?**

**Faster Tests**:
- Handler tests run in milliseconds
- No database startup overhead
- No complex setup required

**Clearer Failures**:
- Handler test fails → problem in handler
- Service test fails → problem in business logic
- Clear attribution speeds debugging

**Better Coverage**:
- Can test error scenarios hard to trigger with real database
- Mock service can return any error
- Test edge cases exhaustively

### 11.4 Industry Standards

**Google Testing Approach**:
- 80% unit tests (fast, isolated)
- 15% integration tests (medium)
- 5% end-to-end tests (slow)

**Martin Fowler's Test Pyramid**:
- Broad base of unit tests
- Narrow top of E2E tests
- Optimize for speed and feedback

**Kent Beck (Creator of TDD)**:
- Write test before code
- Test one thing at a time
- Fast tests enable confidence

**Stripe's Testing Philosophy**:
- Mock external dependencies
- Test error paths extensively
- Fast feedback loop (tests in < 30s)

---


## 13. Testing API Handlers

### 13.1 Unit Test Structure

```go
// @internal/api/handlers/finance/account_handler_test.go
package finance_test

import (
    "bytes"
    "encoding/json"
    "net/http/httptest"
    "testing"
    
    "github.com/gofiber/fiber/v2"
    "github.com/golang/mock/gomock"
    "github.com/stretchr/testify/assert"
    
    "awo.so/internal/api/handlers/finance"
    "awo.so/internal/core/finance/account"
    "awo.so/internal/core/finance/account/mocks"
)

func TestAccountHandler_Create(t *testing.T) {
    tests := []struct {
        name           string
        requestBody    map[string]interface{}
        setupMock      func(*mocks.MockService)
        expectedStatus int
        expectedBody   map[string]interface{}
    }{
        {
            name: "success",
            requestBody: map[string]interface{}{
                "name": "Cash",
                "type": "ASSET",
                "code": "1001",
            },
            setupMock: func(m *mocks.MockService) {
                m.EXPECT().
                    Create(gomock.Any(), gomock.Any()).
                    Return(&account.Account{
                        ID:   "acc_123",
                        Name: "Cash",
                        Type: "ASSET",
                        Code: "1001",
                    }, nil)
            },
            expectedStatus: fiber.StatusCreated,
            expectedBody: map[string]interface{}{
                "id":   "acc_123",
                "name": "Cash",
                "code": "1001",
            },
        },
        {
            name: "validation error",
            requestBody: map[string]interface{}{
                "type": "ASSET",
                "code": "1001",
                // Missing required "name"
            },
            setupMock:      func(m *mocks.MockService) {},
            expectedStatus: fiber.StatusBadRequest,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()
            
            mockService := mocks.NewMockService(ctrl)
            tt.setupMock(mockService)
            
            handler := finance.NewAccountHandler(
                mockService,
                &mockLogger{},
                &mockMetrics{},
                &mockTracer{},
            )
            
            app := fiber.New()
            app.Post("/accounts", func(c *fiber.Ctx) error {
                // Simulate middleware
                c.Locals("user_id", "user_789")
                c.Locals("tenant_id", "tenant_456")
                return handler.Create(c)
            })
            
            // Execute
            body, _ := json.Marshal(tt.requestBody)
            req := httptest.NewRequest("POST", "/accounts", bytes.NewReader(body))
            req.Header.Set("Content-Type", "application/json")
            
            resp, _ := app.Test(req)
            
            // Assert
            assert.Equal(t, tt.expectedStatus, resp.StatusCode)
            
            if tt.expectedBody != nil {
                var result map[string]interface{}
                json.NewDecoder(resp.Body).Decode(&result)
                
                for key, expectedValue := range tt.expectedBody {
                    assert.Equal(t, expectedValue, result[key])
                }
            }
        })
    }
}
```



---


## 12. Best Practices & Patterns

### 12.1 Handler Development Principles

**Principle 1: Single Responsibility**

Each handler method does ONE thing:
- ✅ `Create()` - creates resource
- ✅ `Update()` - updates resource
- ✅ `Archive()` - archives resource
- ❌ `CreateAndNotify()` - does two things (split into separate methods)

**Why?**
- Easier to test (test one behavior)
- Easier to understand (clear purpose)
- Easier to modify (change one thing)

**Principle 2: Fail Fast**

Validate early, fail immediately:

```
1. Parse request body
   → Invalid JSON? Return 400 immediately (don't call service)

2. Check required fields
   → Missing field? Return 400 immediately (don't call service)

3. Validate formats
   → Invalid UUID? Return 400 immediately (don't call service)

4. Call service only when everything validated
   → Service sees only valid, complete requests
```

**Benefits**:
- Saves resources (no wasted database calls)
- Clearer errors (user knows exactly what's wrong)
- Protects service layer (doesn't receive garbage input)

**Principle 3: Delegate, Don't Implement**

Handlers coordinate, they don't calculate:

```
✅ Good:
result = accountService.CalculateBalance(id)
return result

❌ Bad:
transactions = database.GetTransactions(id)
balance = 0
for each transaction:
  balance += transaction.amount
return balance
```

**Why?**
- Business logic changes frequently (easier to change in one place)
- Handlers are HTTP-specific (business logic should work in batch jobs too)
- Testing simpler (mock service vs complex calculation logic)

**Principle 4: Be Consistent**

All handlers follow same pattern:

```
1. Parse & validate
2. Extract context
3. Call service
4. Handle result
5. Return response

Every handler, every time.
```

**Benefits**:
- Code reviews faster (reviewers know what to expect)
- Onboarding easier (new devs see pattern once, understand all handlers)
- Bugs less likely (proven pattern, less improvisation)

### 12.2 Common Patterns

**Pattern 1: Create Operations**

```
POST /api/v1/resource
→ Parse body
→ Validate required fields
→ Call service.Create()
→ Return 201 Created with new resource
→ Location header points to new resource
```

**Pattern 2: Update Operations**

```
PUT /PATCH /api/v1/resource/:id
→ Extract ID from path
→ Parse body (may be partial for PATCH)
→ Validate fields
→ Call service.Update(id, fields)
→ Return 200 OK with updated resource
```

**Pattern 3: List Operations**

```
GET /api/v1/resources?page=1&per_page=20&filter=value
→ Parse query parameters
→ Validate pagination (page ≥ 1, per_page 1-100)
→ Call service.List(params)
→ Return 200 OK with:
  - data array
  - pagination metadata
  - table metadata
  - available actions
```

**Pattern 4: Delete Operations**

```
DELETE /api/v1/resource/:id
→ Extract ID
→ Call service.Delete(id)
→ Return 204 No Content (success, no body)
→ Or 200 OK with confirmation message
```

**Pattern 5: Custom Actions**

```
POST /api/v1/resource/:id/action
→ Extract ID
→ Parse body (if needed)
→ Call service.PerformAction(id, params)
→ Return 200 OK with result
```

Examples:
- `POST /accounts/123/archive` - archive account
- `POST /invoices/456/approve` - approve invoice
- `POST /transactions/789/reconcile` - reconcile transaction

### 12.3 Error Handling Patterns

**Pattern 1: Validation Errors**

```
Missing required field
→ Return 400 Bad Request
→ User message: "Field X is required"
→ Suggestion: "Please provide X"
```

**Pattern 2: Business Rule Violations**

```
Duplicate account code
→ Return 409 Conflict
→ User message: "Code already exists"
→ Suggestion: "Try a different code"
→ Include list of available codes if short
```

**Pattern 3: Authorization Failures**

```
User lacks permission
→ Return 403 Forbidden
→ User message: "You don't have permission to X"
→ Don't suggest how to get permission (security)
```

**Pattern 4: Resource Not Found**

```
Account doesn't exist
→ Return 404 Not Found
→ User message: "Account not found"
→ Suggestion: "It may have been deleted" or "Verify the account ID"
```

**Pattern 5: System Errors**

```
Database connection failed
→ Return 503 Service Unavailable
→ User message: "System temporarily unavailable"
→ Log full technical error
→ Alert operations team
→ Include Retry-After header if known
```

### 12.4 Performance Considerations

**Optimize Database Queries**:
- Use pagination (don't load 10,000 records at once)
- Index frequently filtered columns
- Avoid N+1 queries (load related data in one query)

**Cache Frequently Accessed Data**:
- User permissions (check once per session)
- Lookup tables (account types, status values)
- Tenant configuration

**Monitor Slow Endpoints**:
- Set performance budgets (e.g., 95th percentile < 200ms)
- Alert when endpoints slow down
- Investigate and optimize bottlenecks

**Avoid Doing Work in Handlers**:
- Heavy calculations → Background jobs
- Email sending → Message queue
- Report generation → Async processing

### 12.5 Security Best Practices

**Never Trust Client Input**:
- Validate everything
- Sanitize user input
- Use parameterized queries (prevent SQL injection)

**Verify Authorization**:
- Check user has permission for operation
- Verify resource belongs to user's tenant
- Don't rely on client-side checks

**Protect Sensitive Data**:
- Don't log passwords, tokens, PII
- Redact sensitive fields in error messages
- Use HTTPS for all communication

**Rate Limiting**:
- Prevent abuse (too many requests)
- Per-user limits (prevent one user hogging resources)
- Per-tenant limits (fair resource sharing)

### 12.6 Documentation Standards

**Every Handler Method Should Document**:
- Purpose (what does it do?)
- HTTP method and path
- Required parameters
- Optional parameters
- Response format
- Possible errors
- Required permissions

**Example**:
```
// Archive archives an account, preventing new transactions
// but preserving historical data.
//
// HTTP: POST /api/v1/accounts/:id/archive
// Required: id (path parameter, UUID)
// Permissions: accounts:write
// Returns: 200 OK with archived account
// Errors:
//   - 404 if account not found
//   - 409 if account has pending transactions
//   - 403 if user lacks permission
```

---


## 14. Common Patterns

### 14.1 Pattern: Batch Operations

```go
func (h *AccountHandler) BatchArchive(c *fiber.Ctx) error {
    ctx := c.Context()
    
    // Parse batch request
    var req struct {
        IDs []string `json:"ids"`
    }
    if err := c.BodyParser(&req); err != nil {
        return h.badRequest(c, "INVALID_REQUEST", "Failed to parse request")
    }
    
    // Validate
    if len(req.IDs) == 0 {
        return h.badRequest(c, "MISSING_IDS", "At least one ID is required")
    }
    
    // Process batch
    results := make([]interface{}, 0, len(req.IDs))
    errors := make([]interface{}, 0)
    
    for _, id := range req.IDs {
        account, err := h.accountService.Archive(ctx, id)
        if err != nil {
            errors = append(errors, map[string]interface{}{
                "id":    id,
                "error": err.Error(),
            })
            continue
        }
        results = append(results, toResultType(account))
    }
    
    // Return batch response
    return c.JSON(fiber.Map{
        "successful": len(results),
        "failed":     len(errors),
        "results":    results,
        "errors":     errors,
    })
}
```

### 14.2 Pattern: File Export

```go
func (h *AccountHandler) Export(c *fiber.Ctx) error {
    ctx := c.Context()
    
    // Get accounts
    accounts, _, err := h.accountService.List(ctx, accountsvc.ListParams{
        Page:    1,
        PerPage: 10000, // Export all
    })
    if err != nil {
        return h.handleError(c, err)
    }
    
    // Export format
    format := c.Query("format", "csv")
    
    switch format {
    case "csv":
        return h.exportCSV(c, accounts)
    case "xlsx":
        return h.exportExcel(c, accounts)
    case "pdf":
        return h.exportPDF(c, accounts)
    default:
        return h.badRequest(c, "INVALID_FORMAT", "Unsupported export format")
    }
}
```



---


## 15. Troubleshooting

### Issue: "Tenant context missing"
**Solution**: Ensure tenant middleware is applied before handler
```go
accounts := api.Group("/accounts",
    mw.Auth,    // Sets tenant_id in context
    mw.Tenant,  // REQUIRED: Sets RLS via Store
)
```

### Issue: RLS policy violation
**Solution**: Verify RLS is enabled and Store.SetTenantContextFromCtx is called
```sql
-- Check RLS is enabled
SELECT tablename, rowsecurity FROM pg_tables WHERE tablename = 'accounts';

-- Check policy exists
SELECT * FROM pg_policies WHERE tablename = 'accounts';
```

### Issue: TemplUI component not rendering
**Solution**: Check Accept header and ensure component is registered
```go
// Verify Accept header
isHTMX := c.Get("HX-Request") == "true"
accept := c.Get("Accept")

// Ensure component import
import "awo.so/web/components"
```

### Issue: Error messages showing technical details to users
**Solution**: Always use `UserMessage` field in BusinessError
```go
// ✅ Good: User-friendly message
err.WithUserMessage("The account code is already in use. Please choose a different code.")

// ❌ Bad: Technical message exposed to UI
err.WithMessage("duplicate key value violates unique constraint")
```




---

## Conclusion

This comprehensive guide has combined architectural theory with practical implementation details for developing API handlers in the Awo ERP system.

### Key Takeaways

**Architectural Principles:**
1. **Separation of Concerns**: Handlers focus on HTTP, services handle business logic
2. **Type Safety**: Goa-generated types prevent errors at compile time  
3. **Multi-Tenancy**: Row-level security ensures data isolation
4. **Metadata-Driven UI**: Server controls client behavior
5. **Error Transparency**: Technical details for devs, friendly messages for users

**Development Workflow:**
1. Define types in Goa design files
2. Generate code with `goa gen`
3. Implement handler following 5-step pattern
4. Register routes with appropriate middleware
5. Test thoroughly (unit + integration)
6. Deploy with confidence

**Technology Choices Validated By:**
- **Goa**: Google (Protocol Buffers), Stripe (OpenAPI generation)
- **Fiber**: Discord, PayPal (high-performance frameworks)
- **SQLC**: Google, Facebook (code-generated data access)
- **Server-Side Rendering**: GitHub, Basecamp, Stack Overflow

### Next Steps

**For Developers Implementing Handlers:**
1. Follow the Quick Start (Section 2) to create your first endpoint
2. Use Section 7 patterns for standard CRUD operations
3. Reference Section 9 for proper error handling
4. Write tests following Section 15 patterns

**For Architects Planning Features:**
1. Review Part I for architectural alignment
2. Understand technology trade-offs (Section 4)
3. Plan for multi-tenancy from the start (Section 10)
4. Consider metadata-driven UI (Section 12)

**For Technical Leaders:**
- This architecture scales from 10 to 10,000+ tenants
- Proven patterns reduce time-to-market
- Type safety and RLS provide strong security guarantees
- Simplified stack reduces maintenance costs

### References & Further Reading

**Industry Standards:**
- REST API Design: [Microsoft API Guidelines](https://github.com/microsoft/api-guidelines)
- Multi-Tenancy: [AWS SaaS Architecture](https://aws.amazon.com/partners/programs/saas-factory/)
- Error Handling: [Google API Design Guide](https://cloud.google.com/apis/design/errors)

**Technology Documentation:**
- Goa: https://goa.design/
- Fiber: https://docs.gofiber.io/
- SQLC: https://sqlc.dev/
- PostgreSQL RLS: https://www.postgresql.org/docs/current/ddl-rowsecurity.html
- TemplUI: https://templ.guide/

**Internal Resources:**
- Service Layer: `@internal/core/` package documentation
- Workflows: `@internal/workflows/` package documentation
- Database: `@internal/db/` migrations and schema
- Shared Utilities: `@internal/shared/` package documentation

---

**Document Version:** 5.0 - Complete Edition  
**Last Updated:** October 26, 2025  
**Maintained By:** Awo ERP Architecture Team  
**Questions?** Open an issue or consult the architecture team

---

*This guide combines architectural theory from Version 4.0 with practical implementation from Version 3.0, providing both the "why" and the "how" for API handler development.*
