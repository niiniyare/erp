# Goa Framework Integration Guide for Clean Architecture ERP

## 🌟 Goa Framework Overview

**Goa** is a design-first API framework for Go that generates code from a DSL (Domain Specific Language) written in Go. It follows the principle of "design first, implement later" and provides type-safe, well-documented APIs with minimal boilerplate.

### Key Goa Concepts

1. **Design DSL**: Define APIs using Go code in a declarative manner
2. **Code Generation**: Automatically generate transport, validation, and documentation
3. **Transport Agnostic**: Support HTTP, gRPC, and custom transports
4. **Type Safety**: Compile-time verification of API contracts
5. **Auto Documentation**: Generate OpenAPI specs, client SDKs, and docs

## 🏗️ Goa + Clean Architecture Integration

Your Clean Architecture maps perfectly to Goa's layered approach:

```
┌─────────────────────────────────────────────────────────────┐
│                  Goa Design Layer                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │   API DSL   │  │   Types     │  │  Security   │        │
│  │  design/    │  │ design/     │  │ design/     │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────────────────────────────────────────────┘
                              │ goa gen
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                Generated Transport Layer                    │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │  HTTP       │  │    gRPC     │  │ Validation  │        │
│  │  gen/http/  │  │  gen/grpc/  │  │ gen/tenant/ │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────────────────────────────────────────────┘
                              │ implements
                              ▼
┌─────────────────────────────────────────────────────────────┐
│              Your Clean Architecture API Layer              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │   Service   │  │  Adapters   │  │ Middleware  │        │
│  │ Impls       │  │ Goa↔Domain  │  │  Custom     │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│           Your Existing Core Business Layer                 │
│         (Unchanged - Domain Services & Models)              │
└─────────────────────────────────────────────────────────────┘
```

## 📁 Complete Goa-Enhanced Directory Structure

```
erp-system/
├── design/                           # 🎨 Goa Design Layer
│   ├── design.go                     # Main entry point, API metadata
│   ├── api.go                        # Global API config (CORS, security)
│   ├── types/                        # 📊 Shared type definitions
│   │   ├── common.go                 # Common types (pagination, errors)
│   │   ├── tenant.go                 # Tenant domain types
│   │   ├── user.go                   # User domain types
│   │   ├── financial.go              # Financial domain types
│   │   └── inventory.go              # Inventory domain types
│   │
│   └── services/                     # 🔧 Service definitions per domain
│       ├── auth/
│       │   ├── auth.go               # Authentication service design
│       │   ├── types.go              # Auth-specific types
│       │   └── security.go           # Security schemes (JWT, OAuth)
│       ├── tenant/
│       │   ├── tenant.go             # Tenant management service
│       │   ├── types.go              # Request/response types
│       │   └── errors.go             # Service-specific error types
│       ├── user/
│       │   ├── user.go               # User management service
│       │   └── types.go
│       ├── inventory/
│       │   ├── product.go            # Product management
│       │   ├── warehouse.go          # Warehouse management
│       │   └── types.go
│       ├── financial/
│       │   ├── invoice.go            # Invoice management
│       │   ├── payment.go            # Payment processing
│       │   └── types.go
│       └── reporting/
│           ├── analytics.go          # Analytics and reporting
│           └── types.go
│
├── gen/                              # 🔄 Generated Code (DO NOT MODIFY)
│   ├── http/                         # HTTP transport implementations
│   │   ├── tenant/
│   │   │   ├── server/               # HTTP server code
│   │   │   ├── client/               # HTTP client code
│   │   │   └── cli/                  # CLI commands
│   │   ├── user/
│   │   ├── auth/
│   │   └── openapi3.json             # Generated OpenAPI spec
│   │
│   ├── grpc/                         # gRPC transport (optional)
│   │   ├── tenant/
│   │   ├── user/
│   │   └── protos/
│   │
│   ├── tenant/                       # Service interfaces & types
│   │   ├── service.go                # Service interface
│   │   ├── endpoints.go              # Endpoint definitions
│   │   └── views.go                  # View types
│   │
│   ├── user/
│   ├── auth/
│   └── client/                       # Generated client SDKs
│       ├── tenant/
│       └── user/
│
├── cmd/
│   ├── server/
│   │   └── main.go                   # Enhanced Goa server setup
│   ├── cli/                          # Generated CLI tools
│   │   └── main.go
│   └── migrate/
│       └── main.go
│
├── internal/                         # 🏛️ Your Clean Architecture
│   ├── api/                          # API Adapter Layer
│   │   ├── services/                 # 🔌 Goa service implementations
│   │   │   ├── tenant.go             # Implements gen/tenant.Service
│   │   │   ├── user.go               # Implements gen/user.Service
│   │   │   ├── auth.go               # Implements gen/auth.Service
│   │   │   └── base.go               # Common service patterns
│   │   │
│   │   ├── adapters/                 # 🔄 Type conversion layer
│   │   │   ├── tenant.go             # Goa ↔ Domain type conversions
│   │   │   ├── user.go
│   │   │   ├── common.go             # Common conversion utilities
│   │   │   └── errors.go             # Error mapping
│   │   │
│   │   ├── middleware/               # 🛡️ Custom middleware
│   │   │   ├── auth.go               # Authentication middleware
│   │   │   ├── tenant.go             # Multi-tenant context
│   │   │   ├── logging.go            # Request logging
│   │   │   ├── metrics.go            # Metrics collection
│   │   │   └── recovery.go           # Panic recovery
│   │   │
│   │   └── endpoints/                # 🎯 Endpoint customization
│   │       ├── tenant.go             # Custom endpoint middleware
│   │       └── common.go             # Shared endpoint logic
│   │
│   ├── core/                         # 🧠 Your Core Business Layer (Unchanged)
│   │   ├── tenant/
│   │   │   ├── model.go              # Domain models
│   │   │   ├── service.go            # Business logic
│   │   │   ├── repository.go         # Repository interface
│   │   │   └── errors.go             # Domain-specific errors
│   │   ├── user/
│   │   │   ├── model.go
│   │   │   ├── service.go
│   │   │   └── repository.go
│   │   ├── inventory/
│   │   ├── financial/
│   │   └── shared/
│   │       ├── types.go              # Shared domain types
│   │       ├── errors.go             # Common domain errors
│   │       └── interfaces.go         # Cross-domain interfaces
│   │
│   ├── repositories/                 # 💾 Repository Layer
│   │   ├── tenant/
│   │   │   ├── repository.go         # Repository implementation
│   │   │   ├── converters.go         # Domain ↔ DB conversions
│   │   │   └── queries.go            # SQLC integration
│   │   ├── user/
│   │   ├── inventory/
│   │   └── interfaces/
│   │       └── repositories.go       # All repository interfaces
│   │
│   └── platform/                     # 🏗️ Infrastructure Layer
│       ├── database/
│       │   ├── postgres.go           # PostgreSQL connection
│       │   ├── migrations.go         # Migration runner
│       │   └── health.go             # Health checks
│       ├── cache/
│       │   ├── redis.go              # Redis client
│       │   └── memory.go             # In-memory cache
│       ├── config/
│       │   ├── config.go             # Configuration loader
│       │   └── validation.go         # Config validation
│       ├── messaging/
│       │   ├── rabbitmq.go           # Message queue
│       │   └── events.go             # Event publishing
│       └── monitoring/
│           ├── metrics.go            # Metrics collection
│           ├── tracing.go            # Distributed tracing
│           └── logging.go            # Structured logging
│
├── sql/                              # 🗄️ Database Layer
│   ├── migrations/
│   │   ├── 001_initial_schema.sql
│   │   ├── 002_tenant_tables.sql
│   │   └── 003_user_tables.sql
│   ├── queries/                      # SQLC queries
│   │   ├── tenant.sql
│   │   ├── user.sql
│   │   └── inventory.sql
│   └── schema/
│       └── schema.sql
│
├── configs/                          # ⚙️ Configuration
│   ├── development.yaml
│   ├── staging.yaml
│   ├── production.yaml
│   └── docker-compose.yml
│
├── tests/                            # 🧪 Testing
│   ├── integration/
│   │   ├── tenant_test.go
│   │   ├── user_test.go
│   │   └── fixtures/
│   ├── e2e/                          # End-to-end tests
│   └── mocks/                        # Generated mocks
│       ├── tenant/
│       └── user/
│
├── docs/                             # 📚 Documentation
│   ├── api/                          # Generated API docs
│   │   ├── openapi.yaml
│   │   └── postman_collection.json
│   ├── architecture.md
│   └── deployment.md
│
├── deployments/                      # 🚀 Deployment
│   ├── docker/
│   │   ├── Dockerfile
│   │   └── docker-compose.yml
│   ├── kubernetes/
│   └── terraform/
│
├── tools/                           # 🔧 Development tools
│   ├── goa.go                       # Goa generation script
│   └── mockgen.go                   # Mock generation
│
├── goa.design                       # Goa design file
├── go.mod
├── go.sum
├── Makefile                         # Build automation
└── README.md
```

## 🎨 Goa Design Language Deep Dive

### 1. Main Design Entry Point

```go
// design/design.go
package design

import (
    . "goa.design/goa/v3/dsl"
)

// API describes the global properties of the API server.
var _ = API("erp-system", func() {
    Title("Enterprise ERP System API")
    Description("Comprehensive ERP system with multi-tenant support")
    Version("1.0.0")
    
    // Global configuration
    Server("erp", func() {
        Host("localhost", func() {
            URI("http://localhost:8080")
            URI("https://api.erp-system.com")
        })
    })
    
    // Global CORS policy
    CORS(func() {
        Origin("*", func() {
            Methods("GET", "POST", "PUT", "DELETE", "OPTIONS")
            Headers("*")
            MaxAge(600)
            Credentials()
        })
    })
    
    // Global error responses
    Error("internal_error", ErrorResult)
    Error("bad_request", ErrorResult)
    Error("unauthorized", ErrorResult)
    Error("forbidden", ErrorResult)
    Error("not_found", ErrorResult)
    Error("conflict", ErrorResult)
    Error("unprocessable_entity", ErrorResult)
})

// ErrorResult defines the error response structure
var ErrorResult = ResultType("application/vnd.erp.error", func() {
    Description("Error response")
    Attributes(func() {
        Attribute("code", String, "Error code", func() {
            Example("TENANT_NOT_FOUND")
        })
        Attribute("message", String, "Error message", func() {
            Example("Tenant with ID 'abc123' not found")
        })
        Attribute("details", MapOf(String, Any), "Additional error details")
        Attribute("timestamp", String, "Error timestamp", func() {
            Format(FormatDateTime)
            Example("2023-12-07T10:30:00Z")
        })
        Attribute("request_id", String, "Request ID for tracking", func() {
            Example("req_abc123def456")
        })
    })
    Required("code", "message", "timestamp", "request_id")
})
```

### 2. Global API Configuration

```go
// design/api.go
package design

import (
    . "goa.design/goa/v3/dsl"
)

// Security schemes used across the API
var _ = SecurityScheme("jwt", func() {
    JWTSecurity("JWT", func() {
        Description("JWT token authentication")
        Scope("api:read", "Read access to API")
        Scope("api:write", "Write access to API")
        Scope("admin", "Administrative access")
    })
})

var _ = SecurityScheme("api_key", func() {
    APIKeySecurity("api_key", func() {
        Description("API key authentication for service-to-service communication")
    })
})

// Global middleware
var LoggingMiddleware = func() {
    Middleware(func() {
        Description("Request logging middleware")
    })
}

var AuthMiddleware = func() {
    Middleware(func() {
        Description("Authentication middleware")
    })
}

var TenantMiddleware = func() {
    Middleware(func() {
        Description("Multi-tenant context middleware")
    })
}
```

### 3. Shared Types Definition

```go
// design/types/common.go
package types

import (
    . "goa.design/goa/v3/dsl"
)

// Pagination provides common pagination parameters
var Pagination = Type("Pagination", func() {
    Description("Pagination parameters")
    Attribute("page", UInt, "Page number (1-based)", func() {
        Default(1)
        Minimum(1)
        Example(1)
    })
    Attribute("page_size", UInt, "Number of items per page", func() {
        Default(20)
        Minimum(1)
        Maximum(100)
        Example(20)
    })
    Attribute("sort_by", String, "Field to sort by", func() {
        Example("created_at")
    })
    Attribute("sort_order", String, "Sort order", func() {
        Enum("asc", "desc")
        Default("desc")
        Example("desc")
    })
})

// PaginatedResponse provides common paginated response structure
var PaginatedResponse = func(itemType DataType) *ResultTypeExpr {
    return ResultType("application/vnd.erp.paginated", func() {
        Description("Paginated response")
        Attributes(func() {
            Attribute("data", ArrayOf(itemType), "The data items")
            Attribute("pagination", PaginationMeta, "Pagination metadata")
        })
        Required("data", "pagination")
    })
}

// PaginationMeta provides pagination metadata
var PaginationMeta = Type("PaginationMeta", func() {
    Description("Pagination metadata")
    Attribute("current_page", UInt, "Current page number")
    Attribute("page_size", UInt, "Items per page")
    Attribute("total_items", UInt, "Total number of items")
    Attribute("total_pages", UInt, "Total number of pages")
    Attribute("has_next", Boolean, "Whether there is a next page")
    Attribute("has_prev", Boolean, "Whether there is a previous page")
    Required("current_page", "page_size", "total_items", "total_pages", "has_next", "has_prev")
})

// AuditFields provides common audit trail fields
var AuditFields = func() {
    Attribute("created_at", String, "Creation timestamp", func() {
        Format(FormatDateTime)
        Example("2023-12-07T10:30:00Z")
    })
    Attribute("updated_at", String, "Last update timestamp", func() {
        Format(FormatDateTime)
        Example("2023-12-07T15:45:00Z")
    })
    Attribute("created_by", String, "ID of user who created the record", func() {
        Format(FormatUUID)
        Example("550e8400-e29b-41d4-a716-446655440000")
    })
    Attribute("updated_by", String, "ID of user who last updated the record", func() {
        Format(FormatUUID)
        Example("550e8400-e29b-41d4-a716-446655440000")
    })
}
```

### 4. Tenant Service Design

```go
// design/services/tenant/tenant.go
package tenant

import (
    . "goa.design/goa/v3/dsl"
    "design/types"
)

// Service describes the tenant management service
var _ = Service("tenant", func() {
    Description("Tenant management service for multi-tenant ERP system")
    
    // Apply global middleware
    HTTP(func() {
        Path("/api/v1/tenants")
    })
    
    // Security requirements
    Security("jwt", func() {
        Scope("api:read", "api:write")
    })
    
    // Create tenant endpoint
    Method("create", func() {
        Description("Create a new tenant")
        
        Payload(CreateTenantPayload)
        Result(TenantResult)
        
        Error("bad_request")
        Error("conflict") // For subdomain conflicts
        Error("unauthorized")
        Error("unprocessable_entity")
        
        HTTP(func() {
            POST("/")
            Response(StatusCreated)
            Response("bad_request", StatusBadRequest)
            Response("conflict", StatusConflict)
            Response("unauthorized", StatusUnauthorized)
            Response("unprocessable_entity", StatusUnprocessableEntity)
        })
    })
    
    // Get tenant by ID
    Method("get", func() {
        Description("Get tenant by ID")
        
        Payload(func() {
            Attribute("id", String, "Tenant ID", func() {
                Format(FormatUUID)
                Example("550e8400-e29b-41d4-a716-446655440000")
            })
            Required("id")
        })
        
        Result(TenantResult)
        
        Error("not_found")
        Error("unauthorized")
        
        HTTP(func() {
            GET("/{id}")
            Response(StatusOK)
            Response("not_found", StatusNotFound)
            Response("unauthorized", StatusUnauthorized)
        })
    })
    
    // List tenants with pagination
    Method("list", func() {
        Description("List tenants with pagination and filtering")
        
        Payload(func() {
            Extend(types.Pagination)
            Attribute("name_filter", String, "Filter by tenant name", func() {
                Example("acme")
            })
            Attribute("status_filter", String, "Filter by status", func() {
                Enum("active", "inactive", "suspended")
                Example("active")
            })
        })
        
        Result(types.PaginatedResponse(TenantResult))
        
        Error("bad_request")
        Error("unauthorized")
        
        HTTP(func() {
            GET("/")
            Param("page")
            Param("page_size")
            Param("sort_by")
            Param("sort_order")
            Param("name_filter")
            Param("status_filter")
            Response(StatusOK)
            Response("bad_request", StatusBadRequest)
            Response("unauthorized", StatusUnauthorized)
        })
    })
    
    // Update tenant
    Method("update", func() {
        Description("Update an existing tenant")
        
        Payload(UpdateTenantPayload)
        Result(TenantResult)
        
        Error("bad_request")
        Error("not_found")
        Error("conflict")
        Error("unauthorized")
        Error("unprocessable_entity")
        
        HTTP(func() {
            PUT("/{id}")
            Response(StatusOK)
            Response("bad_request", StatusBadRequest)
            Response("not_found", StatusNotFound)
            Response("conflict", StatusConflict)
            Response("unauthorized", StatusUnauthorized)
            Response("unprocessable_entity", StatusUnprocessableEntity)
        })
    })
    
    // Delete tenant
    Method("delete", func() {
        Description("Delete a tenant (soft delete)")
        
        Payload(func() {
            Attribute("id", String, "Tenant ID", func() {
                Format(FormatUUID)
                Example("550e8400-e29b-41d4-a716-446655440000")
            })
            Required("id")
        })
        
        Result(Empty) // No content response
        
        Error("not_found")
        Error("unauthorized")
        Error("conflict") // If tenant has dependencies
        
        HTTP(func() {
            DELETE("/{id}")
            Response(StatusNoContent)
            Response("not_found", StatusNotFound)
            Response("unauthorized", StatusUnauthorized)
            Response("conflict", StatusConflict)
        })
    })
    
    // Health check endpoint
    Method("health", func() {
        Description("Health check for tenant service")
        
        Result(func() {
            Attribute("status", String, "Service status", func() {
                Enum("healthy", "degraded", "unhealthy")
                Example("healthy")
            })
            Attribute("timestamp", String, "Check timestamp", func() {
                Format(FormatDateTime)
                Example("2023-12-07T10:30:00Z")
            })
            Attribute("version", String, "Service version", func() {
                Example("1.2.3")
            })
            Required("status", "timestamp", "version")
        })
        
        HTTP(func() {
            GET("/health")
            Response(StatusOK)
        })
        
        // No authentication required for health checks
        NoSecurity()
    })
})
```

### 5. Tenant Types Definition

```go
// design/services/tenant/types.go
package tenant

import (
    . "goa.design/goa/v3/dsl"
    "design/types"
)

// TenantResult describes the tenant response
var TenantResult = ResultType("application/vnd.tenant", func() {
    Description("Tenant information")
    Attributes(func() {
        Attribute("id", String, "Unique tenant identifier", func() {
            Format(FormatUUID)
            Example("550e8400-e29b-41d4-a716-446655440000")
        })
        Attribute("name", String, "Tenant display name", func() {
            MinLength(1)
            MaxLength(100)
            Example("Acme Corporation")
        })
        Attribute("slug", String, "Tenant URL slug", func() {
            Pattern("^[a-z0-9-]+$")
            MinLength(3)
            MaxLength(50)
            Example("acme-corp")
        })
        Attribute("subdomain", String, "Subdomain for tenant", func() {
            Pattern("^[a-z0-9-]+$")
            MinLength(3)
            MaxLength(50)
            Example("acme")
        })
        Attribute("status", String, "Tenant status", func() {
            Enum("active", "inactive", "suspended", "pending")
            Example("active")
        })
        Attribute("description", String, "Tenant description", func() {
            MaxLength(500)
            Example("Leading provider of roadrunner traps and anvils")
        })
        Attribute("settings", TenantSettings, "Tenant-specific settings")
        Attribute("subscription", SubscriptionInfo, "Subscription information")
        Attribute("contact", ContactInfo, "Primary contact information")
        types.AuditFields()
    })
    Required("id", "name", "slug", "status", "created_at", "updated_at")
})

// CreateTenantPayload describes the payload for creating a tenant
var CreateTenantPayload = Type("CreateTenantPayload", func() {
    Description("Payload for creating a new tenant")
    Attribute("name", String, "Tenant display name", func() {
        MinLength(1)
        MaxLength(100)
        Example("Acme Corporation")
    })
    Attribute("subdomain", String, "Desired subdomain (optional)", func() {
        Pattern("^[a-z0-9-]+$")
        MinLength(3)
        MaxLength(50)
        Example("acme")
    })
    Attribute("description", String, "Tenant description", func() {
        MaxLength(500)
        Example("Leading provider of roadrunner traps and anvils")
    })
    Attribute("contact", ContactInfo, "Primary contact information")
    Attribute("settings", TenantSettings, "Initial tenant settings")
    Required("name")
})

// UpdateTenantPayload describes the payload for updating a tenant
var UpdateTenantPayload = Type("UpdateTenantPayload", func() {
    Description("Payload for updating an existing tenant")
    Attribute("id", String, "Tenant ID", func() {
        Format(FormatUUID)
        Example("550e8400-e29b-41d4-a716-446655440000")
    })
    Attribute("name", String, "Tenant display name", func() {
        MinLength(1)
        MaxLength(100)
        Example("Acme Corporation Updated")
    })
    Attribute("description", String, "Tenant description", func() {
        MaxLength(500)
        Example("Updated description")
    })
    Attribute("status", String, "Tenant status", func() {
        Enum("active", "inactive", "suspended")
        Example("active")
    })
    Attribute("contact", ContactInfo, "Primary contact information")
    Attribute("settings", TenantSettings, "Tenant settings")
    Required("id")
})

// TenantSettings describes tenant-specific configuration
var TenantSettings = Type("TenantSettings", func() {
    Description("Tenant-specific settings and configuration")
    Attribute("timezone", String, "Default timezone", func() {
        Example("America/New_York")
        Default("UTC")
    })
    Attribute("currency", String, "Default currency code", func() {
        Pattern("^[A-Z]{3}$")
        Example("USD")
        Default("USD")
    })
    Attribute("date_format", String, "Preferred date format", func() {
        Enum("MM/DD/YYYY", "DD/MM/YYYY", "YYYY-MM-DD")
        Example("MM/DD/YYYY")
        Default("MM/DD/YYYY")
    })
    Attribute("language", String, "Default language", func() {
        Pattern("^[a-z]{2}$")
        Example("en")
        Default("en")
    })
    Attribute("features", ArrayOf(String), "Enabled features", func() {
        Example([]string{"inventory", "financial", "hr"})
    })
    Attribute("limits", TenantLimits, "Usage limits")
})

// TenantLimits describes usage limits for the tenant
var TenantLimits = Type("TenantLimits", func() {
    Description("Usage limits for tenant")
    Attribute("max_users", UInt, "Maximum number of users", func() {
        Minimum(1)
        Example(100)
    })
    Attribute("max_storage_mb", UInt, "Maximum storage in MB", func() {
        Minimum(100)
        Example(10240) // 10GB
    })
    Attribute("max_api_calls_per_hour", UInt, "API rate limit per hour", func() {
        Minimum(100)
        Example(10000)
    })
})

// SubscriptionInfo describes subscription details
var SubscriptionInfo = Type("SubscriptionInfo", func() {
    Description("Tenant subscription information")
    Attribute("plan", String, "Subscription plan", func() {
        Enum("starter", "professional", "enterprise")
        Example("professional")
    })
    Attribute("status", String, "Subscription status", func() {
        Enum("active", "past_due", "canceled", "trialing")
        Example("active")
    })
    Attribute("billing_cycle", String, "Billing cycle", func() {
        Enum("monthly", "yearly")
        Example("monthly")
    })
    Attribute("next_billing_date", String, "Next billing date", func() {
        Format(FormatDate)
        Example("2023-12-07")
    })
    Attribute("trial_ends_at", String, "Trial end date", func() {
        Format(FormatDate)
        Example("2023-12-07")
    })
})

// ContactInfo describes contact information
var ContactInfo = Type("ContactInfo", func() {
    Description("Contact information")
    Attribute("name", String, "Contact person name", func() {
        MinLength(1)
        MaxLength(100)
        Example("John Doe")
    })
    Attribute("email", String, "Contact email", func() {
        Format(FormatEmail)
        Example("john.doe@acme.com")
    })
    Attribute("phone", String, "Contact phone number", func() {
        Pattern("^\\+?[1-9]\\d{1,14}$")
        Example("+1-555-123-4567")
    })
    Attribute("title", String, "Job title", func() {
        MaxLength(100)
        Example("Chief Technology Officer")
    })
})
```

## 🔧 Goa Service Implementation Patterns

### 1. Service Implementation

```go
// internal/api/services/tenant.go
package services

import (
    "context"
    "fmt"
    
    "internal/api/adapters"
    "internal/core/tenant"
    "internal/platform/logging"
    
    // Generated Goa interfaces
    goa_tenant "gen/tenant"
)

// TenantService implements the generated tenant.Service interface
type TenantService struct {
    domainService tenant.Service  // Your domain service
    logger        logging.Logger
    adapter       *adapters.TenantAdapter
}

// NewTenantService creates a new Goa tenant service implementation
func NewTenantService(
    domainService tenant.Service,
    logger logging.Logger,
) goa_tenant.Service {
    return &TenantService{
        domainService: domainService,
        logger:        logger,
        adapter:       adapters.NewTenantAdapter(),
    }
}

// Create implements the create endpoint
func (s *TenantService) Create(
    ctx context.Context,
    p *goa_tenant.CreateTenantPayload,
) (*goa_tenant.TenantResult, error) {
    // 1. Log the request
    s.logger.Info("Creating tenant", "name", p.Name)
    
    // 2. Convert Goa payload to domain request
    domainReq, err := s.adapter.CreatePayloadToDomain(p)
    if err != nil {
        return nil, goa_tenant.MakeBadRequest(fmt.Errorf("invalid payload: %w", err))
    }
    
    // 3. Call domain service (your existing business logic!)
    domainTenant, err := s.domainService.CreateTenant(ctx, domainReq)
    if err != nil {
        // 4. Convert domain errors to Goa errors
        return nil, s.adapter.DomainErrorToGoa(err)
    }
    
    // 5. Convert domain model to Goa result
    result := s.adapter.DomainToResult(domainTenant)
    
    s.logger.Info("Tenant created successfully", "id", result.ID)
    return result, nil
}

// Get implements the get endpoint
func (s *TenantService) Get(
    ctx context.Context,
    p *goa_tenant.GetPayload,
) (*goa_tenant.TenantResult, error) {
    s.logger.Info("Getting tenant", "id", p.ID)
    
    // Parse UUID
    tenantID, err := s.adapter.ParseUUID(p.ID)
    if err != nil {
        return nil, goa_tenant.MakeBadRequest(fmt.Errorf("invalid tenant ID: %w", err))
    }
    
    // Call domain service
    domainTenant, err := s.domainService.GetTenant(ctx, tenantID)
    if err != nil {
        return nil, s.adapter.DomainErrorToGoa(err)
    }
    
    result := s.adapter.DomainToResult(domainTenant)
    return result, nil
}

// List implements the list endpoint with pagination
func (s *TenantService) List(
    ctx context.Context,
    p *goa_tenant.ListPayload,
) (*goa_tenant.ListResult, error) {
    s.logger.Info("Listing tenants", "page", p.Page, "page_size", p.PageSize)
    
    // Convert Goa pagination to domain filter
    filter := s.adapter.ListPayloadToDomainFilter(p)
    
    // Call domain service
    tenants, pagination, err := s.domainService.ListTenants(ctx, filter)
    if err != nil {
        return nil, s.adapter.DomainErrorToGoa(err)
    }
    
    // Convert to Goa result
    result := s.adapter.DomainListToGoaResult(tenants, pagination)
    return result, nil
}

// Update implements the update endpoint
func (s *TenantService) Update(
    ctx context.Context,
    p *goa_tenant.UpdateTenantPayload,
) (*goa_tenant.TenantResult, error) {
    s.logger.Info("Updating tenant", "id", p.ID)
    
    // Convert payload
    domainReq, err := s.adapter.UpdatePayloadToDomain(p)
    if err != nil {
        return nil, goa_tenant.MakeBadRequest(fmt.Errorf("invalid payload: %w", err))
    }
    
    // Call domain service
    domainTenant, err := s.domainService.UpdateTenant(ctx, domainReq)
    if err != nil {
        return nil, s.adapter.DomainErrorToGoa(err)
    }
    
    result := s.adapter.DomainToResult(domainTenant)
    s.logger.Info("Tenant updated successfully", "id", result.ID)
    return result, nil
}

// Delete implements the delete endpoint
func (s *TenantService) Delete(
    ctx context.Context,
    p *goa_tenant.DeletePayload,
) error {
    s.logger.Info("Deleting tenant", "id", p.ID)
    
    tenantID, err := s.adapter.ParseUUID(p.ID)
    if err != nil {
        return goa_tenant.MakeBadRequest(fmt.Errorf("invalid tenant ID: %w", err))
    }
    
    // Call domain service
    err = s.domainService.DeleteTenant(ctx, tenantID)
    if err != nil {
        return s.adapter.DomainErrorToGoa(err)
    }
    
    s.logger.Info("Tenant deleted successfully", "id", p.ID)
    return nil
}

// Health implements the health check endpoint
func (s *TenantService) Health(
    ctx context.Context,
) (*goa_tenant.HealthResult, error) {
    // Check domain service health
    health := s.domainService.HealthCheck(ctx)
    
    return &goa_tenant.HealthResult{
        Status:    health.Status,
        Timestamp: health.Timestamp.Format(time.RFC3339),
        Version:   "1.0.0", // Could come from build info
    }, nil
}
```

### 2. Type Adapter Implementation

```go
// internal/api/adapters/tenant.go
package adapters

import (
    "errors"
    "fmt"
    "time"
    
    "github.com/google/uuid"
    
    "internal/core/tenant"
    goa_tenant "gen/tenant"
)

type TenantAdapter struct{}

func NewTenantAdapter() *TenantAdapter {
    return &TenantAdapter{}
}

// CreatePayloadToDomain converts Goa create payload to domain request
func (a *TenantAdapter) CreatePayloadToDomain(p *goa_tenant.CreateTenantPayload) (tenant.CreateRequest, error) {
    req := tenant.CreateRequest{
        Name:        p.Name,
        Description: p.Description,
    }
    
    if p.Subdomain != nil {
        req.Subdomain = p.Subdomain
    }
    
    if p.Contact != nil {
        req.Contact = &tenant.ContactInfo{
            Name:  p.Contact.Name,
            Email: p.Contact.Email,
            Phone: p.Contact.Phone,
            Title: p.Contact.Title,
        }
    }
    
    if p.Settings != nil {
        settings, err := a.goaSettingsToDomain(p.Settings)
        if err != nil {
            return req, fmt.Errorf("invalid settings: %w", err)
        }
        req.Settings = settings
    }
    
    return req, nil
}

// UpdatePayloadToDomain converts Goa update payload to domain request
func (a *TenantAdapter) UpdatePayloadToDomain(p *goa_tenant.UpdateTenantPayload) (tenant.UpdateRequest, error) {
    tenantID, err := uuid.Parse(p.ID)
    if err != nil {
        return tenant.UpdateRequest{}, fmt.Errorf("invalid tenant ID: %w", err)
    }
    
    req := tenant.UpdateRequest{
        ID:          tenantID,
        Name:        p.Name,
        Description: p.Description,
    }
    
    if p.Status != nil {
        status := tenant.Status(*p.Status)
        req.Status = &status
    }
    
    if p.Contact != nil {
        req.Contact = &tenant.ContactInfo{
            Name:  p.Contact.Name,
            Email: p.Contact.Email,
            Phone: p.Contact.Phone,
            Title: p.Contact.Title,
        }
    }
    
    if p.Settings != nil {
        settings, err := a.goaSettingsToDomain(p.Settings)
        if err != nil {
            return req, fmt.Errorf("invalid settings: %w", err)
        }
        req.Settings = settings
    }
    
    return req, nil
}

// ListPayloadToDomainFilter converts Goa list payload to domain filter
func (a *TenantAdapter) ListPayloadToDomainFilter(p *goa_tenant.ListPayload) tenant.ListFilter {
    filter := tenant.ListFilter{
        Page:      int(*p.Page),
        PageSize:  int(*p.PageSize),
        SortBy:    p.SortBy,
        SortOrder: p.SortOrder,
    }
    
    if p.NameFilter != nil {
        filter.NameFilter = p.NameFilter
    }
    
    if p.StatusFilter != nil {
        status := tenant.Status(*p.StatusFilter)
        filter.StatusFilter = &status
    }
    
    return filter
}

// DomainToResult converts domain tenant to Goa result
func (a *TenantAdapter) DomainToResult(t *tenant.Tenant) *goa_tenant.TenantResult {
    result := &goa_tenant.TenantResult{
        ID:        t.ID.String(),
        Name:      t.Name,
        Slug:      t.Slug,
        Status:    string(t.Status),
        CreatedAt: t.CreatedAt.Format(time.RFC3339),
        UpdatedAt: t.UpdatedAt.Format(time.RFC3339),
    }
    
    if t.Subdomain != nil {
        result.Subdomain = *t.Subdomain
    }
    
    if t.Description != nil {
        result.Description = *t.Description
    }
    
    if t.Contact != nil {
        result.Contact = &goa_tenant.ContactInfo{
            Name:  t.Contact.Name,
            Email: t.Contact.Email,
            Phone: t.Contact.Phone,
            Title: t.Contact.Title,
        }
    }
    
    if t.Settings != nil {
        result.Settings = a.domainSettingsToGoa(t.Settings)
    }
    
    if t.Subscription != nil {
        result.Subscription = a.domainSubscriptionToGoa(t.Subscription)
    }
    
    if t.CreatedBy != nil {
        result.CreatedBy = (*t.CreatedBy).String()
    }
    
    if t.UpdatedBy != nil {
        result.UpdatedBy = (*t.UpdatedBy).String()
    }
    
    return result
}

// DomainListToGoaResult converts domain list to Goa paginated result
func (a *TenantAdapter) DomainListToGoaResult(
    tenants []*tenant.Tenant,
    pagination *tenant.PaginationMeta,
) *goa_tenant.ListResult {
    // Convert tenants
    data := make([]*goa_tenant.TenantResult, len(tenants))
    for i, t := range tenants {
        data[i] = a.DomainToResult(t)
    }
    
    // Convert pagination
    paginationMeta := &goa_tenant.PaginationMeta{
        CurrentPage: uint(pagination.CurrentPage),
        PageSize:    uint(pagination.PageSize),
        TotalItems:  uint(pagination.TotalItems),
        TotalPages:  uint(pagination.TotalPages),
        HasNext:     pagination.HasNext,
        HasPrev:     pagination.HasPrev,
    }
    
    return &goa_tenant.ListResult{
        Data:       data,
        Pagination: paginationMeta,
    }
}

// DomainErrorToGoa converts domain errors to Goa errors
func (a *TenantAdapter) DomainErrorToGoa(err error) error {
    switch {
    case errors.Is(err, tenant.ErrNotFound):
        return goa_tenant.MakeNotFound(err)
    case errors.Is(err, tenant.ErrSubdomainExists):
        return goa_tenant.MakeConflict(err)
    case errors.Is(err, tenant.ErrInvalidInput):
        return goa_tenant.MakeBadRequest(err)
    case errors.Is(err, tenant.ErrUnauthorized):
        return goa_tenant.MakeUnauthorized(err)
    case errors.Is(err, tenant.ErrForbidden):
        return goa_tenant.MakeForbidden(err)
    case errors.Is(err, tenant.ErrValidation):
        return goa_tenant.MakeUnprocessableEntity(err)
    default:
        return goa_tenant.MakeInternalError(err)
    }
}

// ParseUUID safely parses a UUID string
func (a *TenantAdapter) ParseUUID(s string) (uuid.UUID, error) {
    return uuid.Parse(s)
}

// Helper methods for nested type conversions
func (a *TenantAdapter) goaSettingsToDomain(s *goa_tenant.TenantSettings) (*tenant.Settings, error) {
    settings := &tenant.Settings{
        Timezone:   s.Timezone,
        Currency:   s.Currency,
        DateFormat: s.DateFormat,
        Language:   s.Language,
        Features:   s.Features,
    }
    
    if s.Limits != nil {
        settings.Limits = &tenant.Limits{
            MaxUsers:            int(*s.Limits.MaxUsers),
            MaxStorageMB:        int(*s.Limits.MaxStorageMb),
            MaxAPICallsPerHour:  int(*s.Limits.MaxAPICallsPerHour),
        }
    }
    
    return settings, nil
}

func (a *TenantAdapter) domainSettingsToGoa(s *tenant.Settings) *goa_tenant.TenantSettings {
    settings := &goa_tenant.TenantSettings{
        Timezone:   &s.Timezone,
        Currency:   &s.Currency,
        DateFormat: &s.DateFormat,
        Language:   &s.Language,
        Features:   s.Features,
    }
    
    if s.Limits != nil {
        maxUsers := uint(s.Limits.MaxUsers)
        maxStorage := uint(s.Limits.MaxStorageMB)
        maxAPICalls := uint(s.Limits.MaxAPICallsPerHour)
        
        settings.Limits = &goa_tenant.TenantLimits{
            MaxUsers:            &maxUsers,
            MaxStorageMb:        &maxStorage,
            MaxAPICallsPerHour:  &maxAPICalls,
        }
    }
    
    return settings
}

func (a *TenantAdapter) domainSubscriptionToGoa(s *tenant.Subscription) *goa_tenant.SubscriptionInfo {
    subscription := &goa_tenant.SubscriptionInfo{
        Plan:         s.Plan,
        Status:       s.Status,
        BillingCycle: s.BillingCycle,
    }
    
    if !s.NextBillingDate.IsZero() {
        nextBilling := s.NextBillingDate.Format("2006-01-02")
        subscription.NextBillingDate = &nextBilling
    }
    
    if s.TrialEndsAt != nil && !s.TrialEndsAt.IsZero() {
        trialEnd := s.TrialEndsAt.Format("2006-01-02")
        subscription.TrialEndsAt = &trialEnd
    }
    
    return subscription
}
```

## 🚀 Enhanced Server Setup

```go
// cmd/server/main.go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    // Generated Goa packages
    "gen/http/tenant/server"
    "gen/tenant"
    
    // Your packages
    "internal/api/middleware"
    "internal/api/services"
    "internal/core/tenant"
    "internal/platform/config"
    "internal/platform/database"
    "internal/platform/logging"
    "internal/repositories/tenant"
    
    // Goa packages
    goahttp "goa.design/goa/v3/http"
    "goa.design/goa/v3/middleware"
)

func main() {
    // 1. Load configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }
    
    // 2. Setup logging
    logger := logging.New(cfg.Logging)
    
    // 3. Setup infrastructure
    ctx := context.Background()
    
    // Database
    db, err := database.Connect(ctx, cfg.Database)
    if err != nil {
        logger.Fatal("Failed to connect to database", "error", err)
    }
    defer db.Close()
    
    // Cache
    cache, err := redis.NewClient(cfg.Redis)
    if err != nil {
        logger.Fatal("Failed to connect to cache", "error", err)
    }
    defer cache.Close()
    
    // SQLC store
    store := sqlc.NewStore(db)
    
    // 4. Setup repositories
    tenantRepo := tenantRepo.NewRepository(store, logger)
    
    // 5. Setup domain services
    tenantService := tenant.NewService(tenantRepo, cache, logger)
    
    // 6. Setup Goa services (your API adapters)
    tenantGoaService := services.NewTenantService(tenantService, logger)
    
    // 7. Setup Goa endpoints
    tenantEndpoints := tenant.NewEndpoints(tenantGoaService)
    
    // Add endpoint middleware
    tenantEndpoints.Use(middleware.RequestID())
    tenantEndpoints.Use(middleware.Log(logger))
    tenantEndpoints.Use(middleware.Trace("erp-system"))
    
    // 8. Setup HTTP transport
    var (
        tenantServer *server.Server
    )
    
    // Create HTTP muxer
    mux := goahttp.NewMuxer()
    
    // Setup tenant HTTP server
    tenantServer = server.New(
        tenantEndpoints,
        mux,
        goahttp.RequestDecoder,
        goahttp.ResponseEncoder,
        errorHandler(logger),
        nil, // No file server
    )
    
    // Mount generated HTTP handlers
    server.Mount(mux, tenantServer)
    
    // 9. Add custom middleware
    var handler http.Handler = mux
    
    // Goa middleware
    handler = middleware.RequestID()(handler)
    handler = middleware.Log(logger)(handler)
    handler = middleware.Trace("erp-system")(handler)
    
    // Custom middleware
    handler = middleware.NewAuthMiddleware(cfg.Auth)(handler)
    handler = middleware.NewTenantMiddleware(tenantService)(handler)
    handler = middleware.NewCORSMiddleware(cfg.CORS)(handler)
    handler = middleware.NewRateLimitMiddleware(cfg.RateLimit)(handler)
    
    // Recovery middleware (should be outermost)
    handler = middleware.NewRecoveryMiddleware(logger)(handler)
    
    // 10. Setup HTTP server
    httpServer := &http.Server{
        Addr:         cfg.Server.Address,
        Handler:      handler,
        ReadTimeout:  cfg.Server.ReadTimeout,
        WriteTimeout: cfg.Server.WriteTimeout,
        IdleTimeout:  cfg.Server.IdleTimeout,
    }
    
    // 11. Start server with graceful shutdown
    go func() {
        logger.Info("Starting HTTP server", "address", cfg.Server.Address)
        if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Fatal("HTTP server failed", "error", err)
        }
    }()
    
    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    logger.Info("Shutting down server...")
    
    // Graceful shutdown with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := httpServer.Shutdown(ctx); err != nil {
        logger.Error("Server forced to shutdown", "error", err)
    }
    
    logger.Info("Server exited")
}

// errorHandler creates a function to handle errors
func errorHandler(logger logging.Logger) func(context.Context, http.ResponseWriter, error) {
    return goahttp.ErrorHandler(func(ctx context.Context, w http.ResponseWriter, err error) {
        // Log the error
        logger.Error("HTTP error", "error", err)
        
        // Extract request ID from context
        if reqID := middleware.RequestID(ctx); reqID != "" {
            w.Header().Set("X-Request-ID", reqID)
        }
        
        // Let Goa handle the error response formatting
        goahttp.ErrorHandler()(ctx, w, err)
    })
}
```

## 🔄 Code Generation Workflow

### 1. Makefile for Automation

```makefile
# Makefile
.PHONY: generate clean build test lint

# Goa code generation
generate:
	@echo "Generating Goa code..."
	goa gen design -o .
	goa example design -o .

# Clean generated code
clean:
	@echo "Cleaning generated code..."
	rm -rf gen/

# Generate mocks for testing
mocks:
	@echo "Generating mocks..."
	go generate ./...

# Build the application
build: generate
	@echo "Building application..."
	go build -o bin/server cmd/server/main.go

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	go test -v -tags=integration ./tests/integration/...

# Lint code
lint:
	@echo "Running linter..."
	golangci-lint run

# Run the server
run: generate
	@echo "Starting server..."
	go run cmd/server/main.go

# Generate OpenAPI documentation
docs: generate
	@echo "Generating API documentation..."
	goa gen design -o docs/ --cmd=openapi

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t erp-system:latest .

# Docker run
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 erp-system:latest
```

### 2. Goa Generation Script

```go
// tools/goa.go
//go:build tools
// +build tools

package tools

import (
    _ "goa.design/goa/v3/cmd/goa"
)

//go:generate goa gen design -o ..
//go:generate goa example design -o ..
```

## 🎯 Benefits Summary

### 1. **Design-First Development**
- ✅ API contracts defined before implementation
- ✅ Automatic OpenAPI specification generation
- ✅ Client SDK generation
- ✅ Consistent API structure across services

### 2. **Type Safety & Validation**
- ✅ Compile-time API contract verification
- ✅ Automatic request/response validation
- ✅ Type-safe error handling
- ✅ Generated test mocks

### 3. **Developer Experience**
- ✅ Reduced boilerplate code
- ✅ Automatic documentation generation
- ✅ Built-in middleware support
- ✅ Multiple transport protocols (HTTP, gRPC)

### 4. **Clean Architecture Preservation**
- ✅ Domain logic remains pure and unchanged
- ✅ Repository pattern maintained
- ✅ Dependency inversion preserved
- ✅ Easy testing with generated mocks

### 5. **Enterprise Features**
- ✅ Built-in security schemes (JWT, OAuth, API keys)
- ✅ CORS support
- ✅ Rate limiting
- ✅ Request tracing and logging
- ✅ Health checks and metrics

Your Clean Architecture provides the perfect foundation for Goa integration, resulting in a powerful, maintainable, and scalable ERP system with all the benefits of design-first API development!
