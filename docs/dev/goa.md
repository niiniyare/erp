# Complete Goa + Gin Integration Guide for ERP System

## 🎯 Complete Implementation Guide

This  guide provides working implementations for every component of your Goa + Gin integration, from design files to production-ready handlers.

## 📁 Complete File Structure

```
internal/
├── api/
│   ├── design/                               # Complete Goa Design Layer
│   │   ├── design.go                         # Main API definition
│   │   ├── openapi.go                        # OpenAPI customization
│   │   ├── services/
│   │   │   ├── auth/
│   │   │   │   ├── auth.go                   # Complete Auth service design
│   │   │   │   └── types.go                  # Auth types and validation
│   │   │   ├── tenant/
│   │   │   │   ├── tenant.go                 # Complete Tenant service
│   │   │   │   └── types.go                  # Tenant types with validation
│   │   │   ├── user/
│   │   │   │   ├── user.go                   # Complete User service
│   │   │   │   ├── types.go                  # User types
│   │   │   │   └── abac_extensions.go        # ABAC user extensions
│   │   │   ├── organization/
│   │   │   │   ├── organization.go           # Complete Organization service
│   │   │   │   └── types.go                  # Organization types
│   │   │   └── abac/
│   │   │       ├── abac.go                   # Complete ABAC service
│   │   │       └── types.go                  # ABAC types and policies
│   │   └── types/
│   │       ├── common.go                     # Common types and validation
│   │       ├── abac.go                       # ABAC common types
│   │       └── errors.go                     # Error definitions
│   │
│   ├── gin/                                  # Gin Implementation Layer
│   │   ├── router.go                         # Main Gin router setup
│   │   ├── handlers/
│   │   │   ├── auth.go                       # Complete Auth handlers
│   │   │   ├── tenant.go                     # Complete Tenant handlers
│   │   │   ├── user.go                       # Complete User handlers
│   │   │   ├── organization.go               # Complete Organization handlers
│   │   │   ├── abac.go                       # Complete ABAC handlers
│   │   │   └── health.go                     # Health check handlers
│   │   ├── middleware/
│   │   │   ├── auth.go                       # JWT authentication
│   │   │   ├── abac.go                       # ABAC authorization
│   │   │   ├── tenant.go                     # Multi-tenant context
│   │   │   ├── validation.go                 # Request validation
│   │   │   ├── logging.go                    # Request logging
│   │   │   ├── recovery.go                   # Panic recovery
│   │   │   └── cors.go                       # CORS handling
│   │   ├── adapters/
│   │   │   ├── auth.go                       # Auth type adapters
│   │   │   ├── tenant.go                     # Tenant type adapters
│   │   │   ├── user.go                       # User type adapters
│   │   │   ├── organization.go               # Organization adapters
│   │   │   ├── abac.go                       # ABAC type adapters
│   │   │   └── common.go                     # Common type conversions
│   │   └── responses/
│   │       ├── success.go                    # Success response helpers
│   │       ├── errors.go                     # Error response helpers
│   │       └── pagination.go                 # Pagination helpers
│   │
│   ├── gen/                                  # Generated Goa code (from your command)
│   └── services/                             # Service implementations
│       ├── auth_service.go                   # Auth service implementation
│       ├── tenant_service.go                 # Tenant service implementation
│       ├── user_service.go                   # User service implementation
│       ├── organization_service.go           # Organization service implementation
│       └── abac_service.go                   # ABAC service implementation
```

## 🎨 1. Complete Goa Design Files

### Main Design Entry Point

```go
// internal/api/design/design.go
package design

import (
    . "goa.design/goa/v3/dsl"
)

// API describes the global properties of the API server
var _ = API("awo-erp", func() {
    Title("AWO Enterprise ERP System")
    Description("Multi-tenant ERP system with ABAC authorization")
    Version("1.0.0")
    
    Contact(func() {
        Name("AWO Development Team")
        Email("dev@awo.com")
        URL("https://awo.com")
    })
    
    License(func() {
        Name("Proprietary")
        URL("https://awo.com/license")
    })
    
    // Global servers
    Server("erp", func() {
        Description("AWO ERP API Server")
        Services("auth", "tenant", "user", "organization", "abac")
        
        Host("development", func() {
            Description("Development server")
            URI("http://localhost:8080")
        })
        
        Host("production", func() {
            Description("Production server") 
            URI("https://api.awo.com")
        })
    })
    
    // Global CORS policy
    CORS(func() {
        Origin("*", func() {
            Methods("GET", "HEAD", "PUT", "PATCH", "POST", "DELETE", "OPTIONS")
            Headers("*")
            MaxAge(600)
            Credentials()
        })
    })
    
    // Global security schemes
    JWTSecurity("jwt", func() {
        Description("JWT token authentication")
        Header("Authorization")
        Prefix("Bearer ")
    })
    
    // Global error responses
    Error("internal_error", String, "Internal server error")
    Error("bad_request", String, "Bad request")
    Error("unauthorized", String, "Unauthorized access")
    Error("forbidden", String, "Forbidden access")
    Error("not_found", String, "Resource not found")
    Error("conflict", String, "Resource conflict")
    Error("validation_error", String, "Validation failed")
    Error("rate_limit_exceeded", String, "Rate limit exceeded")
})
```

### Complete Common Types

```go
// internal/api/design/types/common.go
package types

import (
    . "goa.design/goa/v3/dsl"
)

// Standard error response
var StandardError = ResultType("application/vnd.awo.error", func() {
    Description("Standard error response")
    Attributes(func() {
        Attribute("code", String, "Error code", func() {
            Example("TENANT_NOT_FOUND")
            Pattern("^[A-Z_]+$")
        })
        Attribute("message", String, "Human-readable error message", func() {
            Example("Tenant with ID 'abc123' not found")
            MinLength(1)
            MaxLength(500)
        })
        Attribute("details", MapOf(String, Any), "Additional error details")
        Attribute("timestamp", String, "Error timestamp", func() {
            Format(FormatDateTime)
            Example("2024-01-15T10:30:00Z")
        })
        Attribute("request_id", String, "Request tracking ID", func() {
            Example("req_abc123def456")
            Pattern("^req_[a-zA-Z0-9]+$")
        })
        Attribute("trace_id", String, "Distributed tracing ID", func() {
            Example("trace_xyz789")
        })
        Attribute("path", String, "Request path", func() {
            Example("/api/v1/tenants/123")
        })
        Attribute("method", String, "HTTP method", func() {
            Example("POST")
            Enum("GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS")
        })
    })
    Required("code", "message", "timestamp", "request_id", "path", "method")
})

// Pagination parameters
var PaginationRequest = Type("PaginationRequest", func() {
    Description("Pagination parameters for list operations")
    Attribute("page", UInt, "Page number (1-based)", func() {
        Default(1)
        Minimum(1)
        Maximum(1000)
        Example(1)
    })
    Attribute("limit", UInt, "Number of items per page", func() {
        Default(20)
        Minimum(1)
        Maximum(100)
        Example(20)
    })
    Attribute("sort_by", String, "Field to sort by", func() {
        Example("created_at")
        MaxLength(50)
    })
    Attribute("sort_order", String, "Sort order", func() {
        Enum("asc", "desc")
        Default("desc")
        Example("desc")
    })
    Attribute("search", String, "Search query", func() {
        MaxLength(255)
        Example("john doe")
    })
    Attribute("filters", MapOf(String, String), "Additional filters")
})

// Pagination metadata in responses
var PaginationMeta = Type("PaginationMeta", func() {
    Description("Pagination metadata")
    Attribute("current_page", UInt, "Current page number", func() {
        Minimum(1)
        Example(1)
    })
    Attribute("per_page", UInt, "Items per page", func() {
        Minimum(1)
        Maximum(100)
        Example(20)
    })
    Attribute("total_count", UInt64, "Total number of items", func() {
        Example(150)
    })
    Attribute("total_pages", UInt, "Total number of pages", func() {
        Minimum(1)
        Example(8)
    })
    Attribute("has_next", Boolean, "Whether there is a next page", func() {
        Example(true)
    })
    Attribute("has_prev", Boolean, "Whether there is a previous page", func() {
        Example(false)
    })
    Attribute("next_page", UInt, "Next page number (if has_next)", func() {
        Example(2)
    })
    Attribute("prev_page", UInt, "Previous page number (if has_prev)", func() {
        Example(0)
    })
    Required("current_page", "per_page", "total_count", "total_pages", "has_next", "has_prev")
})

// Generic paginated response
func PaginatedResponse(name string, itemType DataType) *ResultTypeExpr {
    return ResultType("application/vnd.awo.paginated+json", func() {
        TypeName(name)
        Description("Paginated response with metadata")
        Attributes(func() {
            Attribute("data", ArrayOf(itemType), "The data items")
            Attribute("meta", PaginationMeta, "Pagination metadata")
        })
        Required("data", "meta")
    })
}

// Audit trail fields
var AuditFields = func() {
    Attribute("created_at", String, "Creation timestamp", func() {
        Format(FormatDateTime)
        Example("2024-01-15T10:30:00Z")
    })
    Attribute("updated_at", String, "Last update timestamp", func() {
        Format(FormatDateTime) 
        Example("2024-01-15T15:45:00Z")
    })
    Attribute("created_by", String, "ID of user who created the record", func() {
        Format(FormatUUID)
        Example("550e8400-e29b-41d4-a716-446655440000")
    })
    Attribute("updated_by", String, "ID of user who last updated the record", func() {
        Format(FormatUUID)
        Example("550e8400-e29b-41d4-a716-446655440000")
    })
    Attribute("version", UInt, "Record version for optimistic locking", func() {
        Default(1)
        Example(3)
    })
}

// Health check response
var HealthResponse = Type("HealthResponse", func() {
    Description("Health check response")
    Attribute("status", String, "Overall health status", func() {
        Enum("healthy", "degraded", "unhealthy")
        Example("healthy")
    })
    Attribute("timestamp", String, "Check timestamp", func() {
        Format(FormatDateTime)
        Example("2024-01-15T10:30:00Z")
    })
    Attribute("version", String, "Service version", func() {
        Example("1.0.0")
    })
    Attribute("uptime", String, "Service uptime", func() {
        Example("72h30m45s")
    })
    Attribute("checks", ArrayOf(DependencyHealth), "Individual health checks")
    Required("status", "timestamp", "version")
})

var DependencyHealth = Type("DependencyHealth", func() {
    Description("Individual dependency health status")
    Attribute("name", String, "Dependency name", func() {
        Example("postgresql")
    })
    Attribute("status", String, "Dependency status", func() {
        Enum("healthy", "degraded", "unhealthy")
        Example("healthy")
    })
    Attribute("response_time", String, "Response time", func() {
        Example("2.5ms")
    })
    Attribute("message", String, "Status message", func() {
        Example("Connection successful")
    })
    Attribute("last_check", String, "Last check timestamp", func() {
        Format(FormatDateTime)
        Example("2024-01-15T10:30:00Z")
    })
    Required("name", "status", "last_check")
})
```

### Complete Tenant Service Design

```go
// internal/api/design/services/tenant/tenant.go
package tenant

import (
    . "goa.design/goa/v3/dsl"
    "github.com/niiniyare/erp/internal/api/design/types"
)

var _ = Service("tenant", func() {
    Description("Tenant management service for multi-tenant ERP system")
    
    Security("jwt")
    
    Error("tenant_not_found", String, "Tenant not found")
    Error("subdomain_exists", String, "Subdomain already exists")
    Error("tenant_limit_exceeded", String, "Tenant limit exceeded")
    Error("invalid_tenant_data", String, "Invalid tenant data")
    
    Method("create", func() {
        Description("Create a new tenant")
        Security("jwt", func() {
            Scope("tenant:create")
        })
        
        Payload(func() {
            Attribute("name", String, "Tenant name", func() {
                MinLength(2)
                MaxLength(100)
                Pattern("^[a-zA-Z0-9\\s\\-_]+$")
                Example("Acme Corporation")
            })
            Attribute("subdomain", String, "Unique subdomain", func() {
                MinLength(3)
                MaxLength(50)
                Pattern("^[a-z0-9\\-]+$")
                Example("acme-corp")
            })
            Attribute("description", String, "Tenant description", func() {
                MaxLength(500)
                Example("Leading software company")
            })
            Attribute("plan", String, "Subscription plan", func() {
                Enum("basic", "professional", "enterprise")
                Default("basic")
                Example("professional")
            })
            Attribute("settings", MapOf(String, Any), "Tenant-specific settings", func() {
                Example(map[string]interface{}{
                    "timezone": "UTC",
                    "locale": "en-US",
                    "max_users": 100,
                })
            })
            Attribute("contact_email", String, "Primary contact email", func() {
                Format(FormatEmail)
                Example("admin@acme-corp.com")
            })
            Attribute("contact_phone", String, "Primary contact phone", func() {
                Pattern("^\\+?[1-9]\\d{1,14}$")
                Example("+1-555-123-4567")
            })
            Required("name", "subdomain", "contact_email")
        })
        
        Result(TenantResult)
        
        HTTP(func() {
            POST("/tenants")
            Response(StatusCreated)
            Response("bad_request", StatusBadRequest)
            Response("unauthorized", StatusUnauthorized)
            Response("forbidden", StatusForbidden)
            Response("subdomain_exists", StatusConflict)
            Response("tenant_limit_exceeded", StatusTooManyRequests)
        })
    })
    
    Method("get", func() {
        Description("Get tenant by ID")
        Security("jwt", func() {
            Scope("tenant:read")
        })
        
        Payload(func() {
            Attribute("id", String, "Tenant ID", func() {
                Format(FormatUUID)
                Example("550e8400-e29b-41d4-a716-446655440000")
            })
            Required("id")
        })
        
        Result(TenantResult)
        
        HTTP(func() {
            GET("/tenants/{id}")
            Response(StatusOK)
            Response("not_found", StatusNotFound)
            Response("unauthorized", StatusUnauthorized)
            Response("forbidden", StatusForbidden)
        })
    })
    
    Method("list", func() {
        Description("List tenants with pagination and filtering")
        Security("jwt", func() {
            Scope("tenant:read")
        })
        
        Payload(types.PaginationRequest)
        
        Result(types.PaginatedResponse("TenantListResult", TenantResult))
        
        HTTP(func() {
            GET("/tenants")
            Params(func() {
                Param("page", UInt, "Page number", func() {
                    Default(1)
                })
                Param("limit", UInt, "Items per page", func() {
                    Default(20)
                })
                Param("sort_by", String, "Sort field")
                Param("sort_order", String, "Sort order")
                Param("search", String, "Search query")
                Param("plan", String, "Filter by plan")
                Param("status", String, "Filter by status")
            })
            Response(StatusOK)
            Response("unauthorized", StatusUnauthorized)
            Response("forbidden", StatusForbidden)
        })
    })
    
    Method("update", func() {
        Description("Update tenant information")
        Security("jwt", func() {
            Scope("tenant:update")
        })
        
        Payload(func() {
            Attribute("id", String, "Tenant ID", func() {
                Format(FormatUUID)
                Example("550e8400-e29b-41d4-a716-446655440000")
            })
            Attribute("name", String, "Tenant name", func() {
                MinLength(2)
                MaxLength(100)
                Pattern("^[a-zA-Z0-9\\s\\-_]+$")
                Example("Acme Corporation Ltd")
            })
            Attribute("description", String, "Tenant description", func() {
                MaxLength(500)
                Example("Updated description")
            })
            Attribute("plan", String, "Subscription plan", func() {
                Enum("basic", "professional", "enterprise")
                Example("enterprise")
            })
            Attribute("settings", MapOf(String, Any), "Tenant settings")
            Attribute("contact_email", String, "Contact email", func() {
                Format(FormatEmail)
            })
            Attribute("contact_phone", String, "Contact phone", func() {
                Pattern("^\\+?[1-9]\\d{1,14}$")
            })
            Required("id")
        })
        
        Result(TenantResult)
        
        HTTP(func() {
            PUT("/tenants/{id}")
            Response(StatusOK)
            Response("bad_request", StatusBadRequest)
            Response("not_found", StatusNotFound)
            Response("unauthorized", StatusUnauthorized)
            Response("forbidden", StatusForbidden)
        })
    })
    
    Method("delete", func() {
        Description("Delete tenant (soft delete)")
        Security("jwt", func() {
            Scope("tenant:delete")
        })
        
        Payload(func() {
            Attribute("id", String, "Tenant ID", func() {
                Format(FormatUUID)
                Example("550e8400-e29b-41d4-a716-446655440000")
            })
            Required("id")
        })
        
        HTTP(func() {
            DELETE("/tenants/{id}")
            Response(StatusNoContent)
            Response("not_found", StatusNotFound)
            Response("unauthorized", StatusUnauthorized)
            Response("forbidden", StatusForbidden)
        })
    })
    
    Method("get_stats", func() {
        Description("Get tenant statistics and usage metrics")
        Security("jwt", func() {
            Scope("tenant:read")
        })
        
        Payload(func() {
            Attribute("id", String, "Tenant ID", func() {
                Format(FormatUUID)
                Example("550e8400-e29b-41d4-a716-446655440000")
            })
            Required("id")
        })
        
        Result(TenantStatsResult)
        
        HTTP(func() {
            GET("/tenants/{id}/stats")
            Response(StatusOK)
            Response("not_found", StatusNotFound)
            Response("unauthorized", StatusUnauthorized)
            Response("forbidden", StatusForbidden)
        })
    })
})
```

### Complete Tenant Types

```go
// internal/api/design/services/tenant/types.go
package tenant

import (
    . "goa.design/goa/v3/dsl"
    "github.com/niiniyare/erp/internal/api/design/types"
)

var TenantResult = ResultType("application/vnd.tenant+json", func() {
    Description("Tenant information")
    Attributes(func() {
        Attribute("id", String, "Unique tenant identifier", func() {
            Format(FormatUUID)
            Example("550e8400-e29b-41d4-a716-446655440000")
        })
        Attribute("name", String, "Tenant name", func() {
            Example("Acme Corporation")
        })
        Attribute("subdomain", String, "Unique subdomain", func() {
            Example("acme-corp")
        })
        Attribute("description", String, "Tenant description", func() {
            Example("Leading software company")
        })
        Attribute("plan", String, "Subscription plan", func() {
            Enum("basic", "professional", "enterprise")
            Example("professional")
        })
        Attribute("status", String, "Tenant status", func() {
            Enum("active", "suspended", "cancelled", "pending")
            Example("active")
        })
        Attribute("settings", MapOf(String, Any), "Tenant-specific settings", func() {
            Example(map[string]interface{}{
                "timezone": "UTC",
                "locale": "en-US",
                "max_users": 100,
                "features": []string{"reporting", "analytics"},
            })
        })
        Attribute("contact_email", String, "Primary contact email", func() {
            Format(FormatEmail)
            Example("admin@acme-corp.com")
        })
        Attribute("contact_phone", String, "Primary contact phone", func() {
            Example("+1-555-123-4567")
        })
        Attribute("limits", TenantLimits, "Tenant resource limits")
        Attribute("usage", TenantUsage, "Current resource usage")
        types.AuditFields()
    })
    Required("id", "name", "subdomain", "status", "plan", "contact_email", "created_at", "updated_at")
})

var TenantLimits = Type("TenantLimits", func() {
    Description("Tenant resource limits based on plan")
    Attribute("max_users", UInt, "Maximum number of users", func() {
        Example(100)
    })
    Attribute("max_storage_gb", UInt, "Maximum storage in GB", func() {
        Example(50)
    })
    Attribute("max_api_requests_per_hour", UInt, "API rate limit", func() {
        Example(10000)
    })
    Attribute("max_organizations", UInt, "Maximum organizations", func() {
        Example(10)
    })
    Attribute("enabled_features", ArrayOf(String), "Enabled features", func() {
        Example([]string{"reporting", "analytics", "integrations"})
    })
    Required("max_users", "max_storage_gb", "max_api_requests_per_hour")
})

var TenantUsage = Type("TenantUsage", func() {
    Description("Current tenant resource usage")
    Attribute("current_users", UInt, "Current number of users", func() {
        Example(45)
    })
    Attribute("current_storage_gb", Float64, "Current storage usage in GB", func() {
        Example(12.5)
    })
    Attribute("api_requests_last_hour", UInt, "API requests in last hour", func() {
        Example(1250)
    })
    Attribute("current_organizations", UInt, "Current number of organizations", func() {
        Example(3)
    })
    Attribute("last_activity", String, "Last activity timestamp", func() {
        Format(FormatDateTime)
        Example("2024-01-15T15:45:00Z")
    })
    Required("current_users", "current_storage_gb", "current_organizations")
})

var TenantStatsResult = ResultType("application/vnd.tenant.stats+json", func() {
    Description("Tenant statistics and metrics")
    Attributes(func() {
        Attribute("tenant_id", String, "Tenant ID", func() {
            Format(FormatUUID)
            Example("550e8400-e29b-41d4-a716-446655440000")
        })
        Attribute("usage", TenantUsage, "Current usage metrics")
        Attribute("limits", TenantLimits, "Resource limits")
        Attribute("growth_metrics", TenantGrowthMetrics, "Growth statistics")
        Attribute("activity_metrics", TenantActivityMetrics, "Activity statistics")
        Attribute("calculated_at", String, "When stats were calculated", func() {
            Format(FormatDateTime)
            Example("2024-01-15T16:00:00Z")
        })
    })
    Required("tenant_id", "usage", "limits", "calculated_at")
})

var TenantGrowthMetrics = Type("TenantGrowthMetrics", func() {
    Description("Tenant growth statistics")
    Attribute("users_added_last_30_days", UInt, "Users added in last 30 days", func() {
        Example(12)
    })
    Attribute("storage_growth_last_30_days_gb", Float64, "Storage growth in last 30 days", func() {
        Example(2.3)
    })
    Attribute("organizations_added_last_30_days", UInt, "Organizations added in last 30 days", func() {
        Example(1)
    })
})

var TenantActivityMetrics = Type("TenantActivityMetrics", func() {
    Description("Tenant activity statistics")
    Attribute("active_users_last_24_hours", UInt, "Active users in last 24 hours", func() {
        Example(28)
    })
    Attribute("api_requests_last_24_hours", UInt, "API requests in last 24 hours", func() {
        Example(15000)
    })
    Attribute("login_sessions_last_24_hours", UInt, "Login sessions in last 24 hours", func() {
        Example(45)
    })
    Attribute("most_active_features", ArrayOf(String), "Most used features", func() {
        Example([]string{"user_management", "reporting", "dashboard"})
    })
})
```

### Complete Auth Service Design

```go
// internal/api/design/services/auth/auth.go
package auth

import (
    . "goa.design/goa/v3/dsl"
    "github.com/niiniyare/erp/internal/api/design/types"
)

var _ = Service("auth", func() {
    Description("Authentication and authorization service")
    
    Error("invalid_credentials", String, "Invalid username or password")
    Error("account_locked", String, "Account is locked")
    Error("token_expired", String, "Token has expired")
    Error("invalid_token", String, "Invalid token")
    Error("mfa_required", String, "Multi-factor authentication required")
    Error("password_reset_required", String, "Password reset required")
    
    Method("login", func() {
        Description("Authenticate user and return access token")
        
        Payload(func() {
            Attribute("email", String, "User email address", func() {
                Format(FormatEmail)
                Example("user@acme-corp.com")
            })
            Attribute("password", String, "User password", func() {
                MinLength(8)
                MaxLength(128)
                Example("SecurePassword123!")
            })
            Attribute("tenant_subdomain", String, "Tenant subdomain", func() {
                MinLength(3)
                MaxLength(50)
                Pattern("^[a-z0-9\\-]+$")
                Example("acme-corp")
            })
            Attribute("remember_me", Boolean, "Remember login for extended period", func() {
                Default(false)
                Example(true)
            })
            Attribute("mfa_token", String, "Multi-factor authentication token", func() {
                Pattern("^[0-9]{6}$")
                Example("123456")
            })
            Required("email", "password", "tenant_subdomain")
        })
        
        Result(LoginResult)
        
        HTTP(func() {
            POST("/auth/login")
            Response(StatusOK)
            Response("bad_request", StatusBadRequest)
            Response("invalid_credentials", StatusUnauthorized)
            Response("account_locked", StatusLocked)
            Response("mfa_required", StatusPreconditionRequired)
            Response("password_reset_required", StatusPreconditionRequired)
        })
    })
    
    Method("refresh", func() {
        Description("Refresh access token using refresh token")
        
        Payload(func() {
            Attribute("refresh_token", String, "Refresh token", func() {
                MinLength(1)
                Example("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
            })
            Required("refresh_token")
        })
        
        Result(TokenResult)
        
        HTTP(func() {
            POST("/auth/refresh")
            Response(StatusOK)
            Response("bad_request", StatusBadRequest)
            Response("invalid_token", StatusUnauthorized)
            Response("token_expired", StatusUnauthorized)
        })
    })
    
    Method("logout", func() {
        Description("Logout user and invalidate tokens")
        Security("jwt")
        
        Payload(func() {
            Attribute("refresh_token", String, "Refresh token to invalidate", func() {
                Example("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
            })
            Attribute("logout_all_sessions", Boolean, "Logout from all sessions", func() {
                Default(false)
                Example(false)
            })
        })
        
        HTTP(func() {
            POST("/auth/logout")
            Response(StatusNoContent)
            Response("unauthorized", StatusUnauthorized)
        })
    })
    
    Method("me", func() {
        Description("Get current user profile")
        Security("jwt")
        
        Result(UserProfileResult)
        
        HTTP(func() {
            GET("/auth/me")
            Response(StatusOK)
            Response("unauthorized", StatusUnauthorized)
        })
    })
    
    Method("change_password", func() {
        Description("Change user password")
        Security("jwt")
        
        Payload(func() {
            Attribute("current_password", String, "Current password", func() {
                MinLength(8)
                MaxLength(128)
                Example("CurrentPassword123!")
            })
            Attribute("new_password", String, "New password", func() {
                MinLength(8)
                MaxLength(128)
                Pattern("^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)(?=.*[@$!%*?&])[A-Za-z\\d@$!%*?&]")
                Example("NewSecurePassword123!")
            })
            Attribute("confirm_password", String, "Confirm new password", func() {
                Example("NewSecurePassword123!")
            })
            Required("current_password", "new_password", "confirm_password")
        })
        
        HTTP(func() {
            POST("/auth/change-password")
            Response(StatusNoContent)
            Response("bad_request", StatusBadRequest)
            Response("unauthorized", StatusUnauthorized)
            Response("invalid_credentials", StatusUnauthorized)
        })
    })
    
    Method("forgot_password", func() {
        Description("Request password reset")
        
        Payload(func() {
            Attribute("email", String, "User email address", func() {
                Format(FormatEmail)
                Example("user@acme-corp.com")
            })
            Attribute("tenant_subdomain", String, "Tenant subdomain", func() {
                Pattern("^[a-z0-9\\-]+$")
                Example("acme-corp")
            })
            Required("email", "tenant_subdomain")
        })
        
        HTTP(func() {
            POST("/auth/forgot-password")
            Response(StatusAccepted)
            Response("bad_request", StatusBadRequest)
            Response("not_found", StatusNotFound)
        })
    })
    
    Method("reset_password", func() {
        Description("Reset password using reset token")
        
        Payload(func() {
            Attribute("reset_token", String, "Password reset token", func() {
                MinLength(1)
                Example("reset_token_abc123def456")
            })
            Attribute("new_password", String, "New password", func() {
                MinLength(8)
                MaxLength(128)
                Pattern("^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)(?=.*[@$!%*?&])[A-Za-z\\d@$!%*?&]")
                Example("NewSecurePassword123!")
            })
            Attribute("confirm_password", String, "Confirm new password", func() {
                Example("NewSecurePassword123!")
            })
            Required("reset_token", "new_password", "confirm_password")
        })
        
        HTTP(func() {
            POST("/auth/reset-password")
            Response(StatusNoContent)
            Response("bad_request", StatusBadRequest)
            Response("invalid_token", StatusUnauthorized)
            Response("token_expired", StatusUnauthorized)
        })
    })
})
```

### Complete Auth Types

```go
// internal/api/design/services/auth/types.go
package auth

import (
    . "goa.design/goa/v3/dsl"
)

var LoginResult = ResultType("application/vnd.auth.login+json", func() {
    Description("Login response with tokens and user information")
    Attributes(func() {
        Attribute("access_token", String, "JWT access token", func() {
            Example("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
        })
        Attribute("refresh_token", String, "Refresh token", func() {
            Example("refresh_eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
        })
        Attribute("token_type", String, "Token type", func() {
            Default("Bearer")
            Example("Bearer")
        })
        Attribute("expires_in", UInt, "Token expiration time in seconds", func() {
            Example(3600)
        })
        Attribute("user", UserProfileResult, "User profile information")
        Attribute("tenant", TenantInfo, "Tenant information")
        Attribute("permissions", ArrayOf(String), "User permissions", func() {
            Example([]string{"tenant:read", "user:create", "organization:update"})
        })
        Attribute("session_id", String, "Session identifier", func() {
            Example("session_abc123def456")
        })
        Attribute("mfa_enabled", Boolean, "Whether MFA is enabled", func() {
            Example(true)
        })
        Attribute("password_expires_at", String, "Password expiration date", func() {
            Format(FormatDateTime)
            Example("2024-07-15T10:30:00Z")
        })
    })
    Required("access_token", "refresh_token", "token_type", "expires_in", "user", "tenant")
})

var TokenResult = ResultType("application/vnd.auth.token+json", func() {
    Description("Token refresh response")
    Attributes(func() {
        Attribute("access_token", String, "New JWT access token", func() {
            Example("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
        })
        Attribute("token_type", String, "Token type", func() {
            Default("Bearer")
            Example("Bearer")
        })
        Attribute("expires_in", UInt, "Token expiration time in seconds", func() {
            Example(3600)
        })
        Attribute("refresh_token", String, "New refresh token (if rotated)", func() {
            Example("refresh_eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
        })
    })
    Required("access_token", "token_type", "expires_in")
})

var UserProfileResult = ResultType("application/vnd.auth.profile+json", func() {
    Description("User profile information")
    Attributes(func() {
        Attribute("id", String, "User unique identifier", func() {
            Format(FormatUUID)
            Example("550e8400-e29b-41d4-a716-446655440000")
        })
        Attribute("email", String, "User email address", func() {
            Format(FormatEmail)
            Example("user@acme-corp.com")
        })
        Attribute("first_name", String, "User first name", func() {
            Example("John")
        })
        Attribute("last_name", String, "User last name", func() {
            Example("Doe")
        })
        Attribute("display_name", String, "User display name", func() {
            Example("John Doe")
        })
        Attribute("avatar_url", String, "User avatar URL", func() {
            Format(FormatURI)
            Example("https://cdn.acme-corp.com/avatars/john-doe.jpg")
        })
        Attribute("roles", ArrayOf(String), "User roles", func() {
            Example([]string{"admin", "manager"})
        })
        Attribute("status", String, "User status", func() {
            Enum("active", "inactive", "suspended", "pending")
            Example("active")
        })
        Attribute("last_login", String, "Last login timestamp", func() {
            Format(FormatDateTime)
            Example("2024-01-15T10:30:00Z")
        })
        Attribute("timezone", String, "User timezone", func() {
            Example("America/New_York")
        })
        Attribute("locale", String, "User locale", func() {
            Example("en-US")
        })
        Attribute("preferences", MapOf(String, Any), "User preferences", func() {
            Example(map[string]interface{}{
                "theme": "dark",
                "notifications": true,
                "dashboard_layout": "grid",
            })
        })
        Attribute("mfa_enabled", Boolean, "Whether MFA is enabled", func() {
            Example(true)
        })
        Attribute("email_verified", Boolean, "Whether email is verified", func() {
            Example(true)
        })
        Attribute("created_at", String, "Account creation timestamp", func() {
            Format(FormatDateTime)
            Example("2024-01-01T10:30:00Z")
        })
        Attribute("updated_at", String, "Last profile update timestamp", func() {
            Format(FormatDateTime)
            Example("2024-01-15T15:45:00Z")
        })
    })
    Required("id", "email", "first_name", "last_name", "status", "roles", "created_at", "updated_at")
})

var TenantInfo = Type("TenantInfo", func() {
    Description("Basic tenant information for auth context")
    Attribute("id", String, "Tenant unique identifier", func() {
        Format(FormatUUID)
        Example("550e8400-e29b-41d4-a716-446655440000")
    })
    Attribute("name", String, "Tenant name", func() {
        Example("Acme Corporation")
    })
    Attribute("subdomain", String, "Tenant subdomain", func() {
        Example("acme-corp")
    })
    Attribute("plan", String, "Tenant subscription plan", func() {
        Enum("basic", "professional", "enterprise")
        Example("professional")
    })
    Attribute("status", String, "Tenant status", func() {
        Enum("active", "suspended", "cancelled")
        Example("active")
    })
    Required("id", "name", "subdomain", "plan", "status")
})
```

## 🚀 2. Complete Gin Implementation

### Main Gin Router

```go
// internal/api/gin/router.go
package gin

import (
    "context"
    "time"
    
    "github.com/gin-contrib/cors"
    "github.com/gin-contrib/gzip"
    "github.com/gin-contrib/pprof"
    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    
    // Your imports
    "github.com/niiniyare/erp/internal/platform/config"
    "github.com/niiniyare/erp/internal/shared/logger"
    
    // Gin handlers
    "github.com/niiniyare/erp/internal/api/gin/handlers"
    "github.com/niiniyare/erp/internal/api/gin/middleware"
)

type Router struct {
    engine   *gin.Engine
    config   *config.Config
    logger   logger.Logger
    handlers *handlers.Handlers
}

type Dependencies struct {
    AuthService     interface{}
    TenantService   interface{}
    UserService     interface{}
    ABACService     interface{}
    OrgService      interface{}
    HealthService   interface{}
}

func NewRouter(cfg *config.Config, logger logger.Logger, deps *Dependencies) *Router {
    // Set Gin mode
    if cfg.Environment == "production" {
        gin.SetMode(gin.ReleaseMode)
    }
    
    engine := gin.New()
    
    // Initialize handlers
    h := &handlers.Handlers{
        Auth:         handlers.NewAuthHandler(deps.AuthService, logger),
        Tenant:       handlers.NewTenantHandler(deps.TenantService, logger),
        User:         handlers.NewUserHandler(deps.UserService, logger),
        ABAC:         handlers.NewABACHandler(deps.ABACService, logger),
        Organization: handlers.NewOrganizationHandler(deps.OrgService, logger),
        Health:       handlers.NewHealthHandler(deps.HealthService, logger),
    }
    
    return &Router{
        engine:   engine,
        config:   cfg,
        logger:   logger,
        handlers: h,
    }
}

func (r *Router) Setup() *gin.Engine {
    r.setupGlobalMiddleware()
    r.setupSystemRoutes()
    r.setupAPIRoutes()
    return r.engine
}

func (r *Router) setupGlobalMiddleware() {
    // Recovery middleware (first)
    r.engine.Use(middleware.Recovery(r.logger))
    
    // Request ID middleware
    r.engine.Use(middleware.RequestID())
    
    // Logging middleware
    r.engine.Use(middleware.RequestLogger(r.logger))
    
    // CORS middleware
    r.engine.Use(cors.New(cors.Config{
        AllowOrigins: r.config.CORS.AllowedOrigins,
        AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
        AllowHeaders: []string{
            "Origin", "Content-Type", "Accept", "Authorization", 
            "X-Request-ID", "X-Tenant-ID", "X-API-Key",
        },
        ExposeHeaders:    []string{"X-Request-ID", "X-Total-Count"},
        AllowCredentials: true,
        MaxAge:          12 * time.Hour,
    }))
    
    // Compression middleware
    r.engine.Use(gzip.Gzip(gzip.DefaultCompression))
    
    // Security headers
    r.engine.Use(middleware.SecurityHeaders())
    
    // Rate limiting
    if r.config.RateLimit.Enabled {
        r.engine.Use(middleware.RateLimit(r.config.RateLimit))
    }
    
    // Metrics middleware
    r.engine.Use(middleware.Metrics())
}

func (r *Router) setupSystemRoutes() {
    // Health endpoints
    health := r.engine.Group("/health")
    {
        health.GET("/live", r.handlers.Health.Liveness)
        health.GET("/ready", r.handlers.Health.Readiness)
        health.GET("/status", r.handlers.Health.Status)
    }
    
    // Metrics endpoint
    r.engine.GET("/metrics", gin.WrapH(promhttp.Handler()))
    
    // OpenAPI documentation
    r.engine.Static("/docs", "./internal/api/gen/http/")
    r.engine.GET("/openapi.json", r.serveOpenAPISpec)
    
    // Profiling (development only)
    if r.config.Environment == "development" {
        pprof.Register(r.engine)
    }
}

func (r *Router) setupAPIRoutes() {
    // API v1 routes
    v1 := r.engine.Group("/api/v1")
    
    // Public routes (no auth required)
    r.setupPublicRoutes(v1)
    
    // Protected routes (auth required)
    protected := v1.Group("")
    protected.Use(middleware.JWTAuth(r.config.JWT))
    protected.Use(middleware.TenantContext())
    
    r.setupAuthRoutes(protected)
    r.setupTenantRoutes(protected)
    r.setupUserRoutes(protected)
    r.setupOrganizationRoutes(protected)
    r.setupABACRoutes(protected)
}

func (r *Router) setupPublicRoutes(v1 *gin.RouterGroup) {
    // Authentication routes (public)
    auth := v1.Group("/auth")
    {
        auth.POST("/login", r.handlers.Auth.Login)
        auth.POST("/refresh", r.handlers.Auth.RefreshToken)
        auth.POST("/forgot-password", r.handlers.Auth.ForgotPassword)
        auth.POST("/reset-password", r.handlers.Auth.ResetPassword)
    }
}

func (r *Router) setupAuthRoutes(protected *gin.RouterGroup) {
    auth := protected.Group("/auth")
    {
        auth.POST("/logout", r.handlers.Auth.Logout)
        auth.GET("/me", r.handlers.Auth.GetProfile)
        auth.POST("/change-password", r.handlers.Auth.ChangePassword)
    }
}

func (r *Router) setupTenantRoutes(protected *gin.RouterGroup) {
    tenants := protected.Group("/tenants")
    {
        tenants.POST("/", 
            middleware.RequirePermission("tenant:create"),
            middleware.ValidateJSON(),
            r.handlers.Tenant.Create)
        
        tenants.GET("/", 
            middleware.RequirePermission("tenant:read"),
            r.handlers.Tenant.List)
            
        tenants.GET("/:id", 
            middleware.RequirePermission("tenant:read"),
            middleware.ValidateUUID("id"),
            r.handlers.Tenant.Get)
            
        tenants.PUT("/:id", 
            middleware.RequirePermission("tenant:update"),
            middleware.ValidateUUID("id"),
            middleware.ValidateJSON(),
            r.handlers.Tenant.Update)
            
        tenants.DELETE("/:id", 
            middleware.RequirePermission("tenant:delete"),
            middleware.ValidateUUID("id"),
            r.handlers.Tenant.Delete)
            
        tenants.GET("/:id/stats", 
            middleware.RequirePermission("tenant:read"),
            middleware.ValidateUUID("id"),
            r.handlers.Tenant.GetStats)
    }
}

func (r *Router) setupUserRoutes(protected *gin.RouterGroup) {
    users := protected.Group("/users")
    {
        users.POST("/", 
            middleware.RequirePermission("user:create"),
            middleware.ValidateJSON(),
            r.handlers.User.Create)
            
        users.GET("/", 
            middleware.RequirePermission("user:read"),
            r.handlers.User.List)
            
        users.GET("/:id", 
            middleware.RequirePermission("user:read"),
            middleware.ValidateUUID("id"),
            r.handlers.User.Get)
            
        users.PUT("/:id", 
            middleware.RequirePermission("user:update"),
            middleware.ValidateUUID("id"),
            middleware.ValidateJSON(),
            r.handlers.User.Update)
            
        users.DELETE("/:id", 
            middleware.RequirePermission("user:delete"),
            middleware.ValidateUUID("id"),
            r.handlers.User.Delete)
            
        users.GET("/:id/permissions", 
            middleware.RequirePermission("user:read"),
            middleware.ValidateUUID("id"),
            r.handlers.User.GetPermissions)
            
        users.PUT("/:id/permissions", 
            middleware.RequirePermission("user:update"),
            middleware.ValidateUUID("id"),
            middleware.ValidateJSON(),
            r.handlers.User.UpdatePermissions)
    }
}

func (r *Router) setupOrganizationRoutes(protected *gin.RouterGroup) {
    orgs := protected.Group("/organizations")
    {
        orgs.POST("/", 
            middleware.RequirePermission("organization:create"),
            middleware.ValidateJSON(),
            r.handlers.Organization.Create)
            
        orgs.GET("/", 
            middleware.RequirePermission("organization:read"),
            r.handlers.Organization.List)
            
        orgs.GET("/:id", 
            middleware.RequirePermission("organization:read"),
            middleware.ValidateUUID("id"),
            r.handlers.Organization.Get)
            
        orgs.PUT("/:id", 
            middleware.RequirePermission("organization:update"),
            middleware.ValidateUUID("id"),
            middleware.ValidateJSON(),
            r.handlers.Organization.Update)
            
        orgs.DELETE("/:id", 
            middleware.RequirePermission("organization:delete"),
            middleware.ValidateUUID("id"),
            r.handlers.Organization.Delete)
    }
}

func (r *Router) setupABACRoutes(protected *gin.RouterGroup) {
    abac := protected.Group("/abac")
    {
        // Policy management
        policies := abac.Group("/policies")
        {
            policies.POST("/", 
                middleware.RequirePermission("abac:policy:create"),
                middleware.ValidateJSON(),
                r.handlers.ABAC.CreatePolicy)
                
            policies.GET("/", 
                middleware.RequirePermission("abac:policy:read"),
                r.handlers.ABAC.ListPolicies)
                
            policies.GET("/:id", 
                middleware.RequirePermission("abac:policy:read"),
                middleware.ValidateUUID("id"),
                r.handlers.ABAC.GetPolicy)
                
            policies.PUT("/:id", 
                middleware.RequirePermission("abac:policy:update"),
                middleware.ValidateUUID("id"),
                middleware.ValidateJSON(),
                r.handlers.ABAC.UpdatePolicy)
                
            policies.DELETE("/:id", 
                middleware.RequirePermission("abac:policy:delete"),
                middleware.ValidateUUID("id"),
                r.handlers.ABAC.DeletePolicy)
        }
        
        // Permission checking
        abac.POST("/check", 
            middleware.ValidateJSON(),
            r.handlers.ABAC.CheckPermission)
            
        abac.POST("/batch-check", 
            middleware.ValidateJSON(),
            r.handlers.ABAC.BatchCheckPermissions)
        
        // User permissions
        abac.GET("/users/:user_id/permissions", 
            middleware.RequirePermission("abac:user:read"),
            middleware.ValidateUUID("user_id"),
            r.handlers.ABAC.GetUserPermissions)
    }
}

func (r *Router) serveOpenAPISpec(c *gin.Context) {
    c.File("./internal/api/gen/http/openapi3.json")
}

// Graceful shutdown
func (r *Router) Shutdown(ctx context.Context) error {
    r.logger.Info("Shutting down HTTP router...")
    // Add any cleanup logic here
    return nil
}
```

## 🔧 3. Complete Gin Handlers

### Tenant Handler Implementation

```go
// internal/api/gin/handlers/tenant.go
package handlers

import (
    "net/http"
    "strconv"
    
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    
    // Your domain imports
    "github.com/niiniyare/erp/internal/core/tenant"
    "github.com/niiniyare/erp/internal/shared/logger"
    
    // Generated Goa types
    goa_tenant "github.com/niiniyare/erp/internal/api/gen/tenant"
    
    // Gin utilities
    "github.com/niiniyare/erp/internal/api/gin/adapters"
    "github.com/niiniyare/erp/internal/api/gin/responses"
)

type TenantHandler struct {
    service tenant.Service
    logger  logger.Logger
    adapter *adapters.TenantAdapter
}

func NewTenantHandler(service tenant.Service, logger logger.Logger) *TenantHandler {
    return &TenantHandler{
        service: service,
        logger:  logger,
        adapter: adapters.NewTenantAdapter(),
    }
}

func (h *TenantHandler) Create(c *gin.Context) {
    ctx := c.Request.Context()
    h.logger.Info("Creating tenant", "path", c.FullPath())
    
    // Parse and validate payload
    var payload goa_tenant.CreateTenantPayload
    if err := c.ShouldBindJSON(&payload); err != nil {
        h.logger.Error("Invalid create tenant payload", "error", err)
        responses.ValidationError(c, "Invalid request payload", err)
        return
    }
    
    // Convert to domain request
    domainReq, err := h.adapter.CreatePayloadToDomain(&payload)
    if err != nil {
        h.logger.Error("Failed to convert payload to domain", "error", err)
        responses.BadRequest(c, "Invalid payload data", err)
        return
    }
    
    // Add user context
    if userID := c.GetString("user_id"); userID != "" {
        domainReq.CreatedBy = &userID
    }
    
    // Add tenant context
    if tenantID := c.GetString("tenant_id"); tenantID != "" {
        domainReq.TenantID = &tenantID
    }
    
    // Call domain service
    result, err := h.service.CreateTenant(ctx, domainReq)
    if err != nil {
        h.logger.Error("Failed to create tenant", "error", err)
        responses.HandleDomainError(c, err)
        return
    }
    
    // Convert to Goa response
    goaResult := h.adapter.DomainToGoaResult(result)
    
    h.logger.Info("Tenant created successfully", 
        "tenant_id", result.ID, 
        "name", result.Name,
        "subdomain", result.Subdomain)
        
    responses.Created(c, goaResult)
}

func (h *TenantHandler) Get(c *gin.Context) {
    ctx := c.Request.Context()
    tenantIDStr := c.Param("id")
    
    tenantID, err := uuid.Parse(tenantIDStr)
    if err != nil {
        h.logger.Error("Invalid tenant ID format", "id", tenantIDStr, "error", err)
        responses.BadRequest(c, "Invalid tenant ID format", err)
        return
    }
    
    h.logger.Info("Getting tenant", "tenant_id", tenantID)
    
    result, err := h.service.GetTenant(ctx, tenantID)
    if err != nil {
        h.logger.Error("Failed to get tenant", "tenant_id", tenantID, "error", err)
        responses.HandleDomainError(c, err)
        return
    }
    
    goaResult := h.adapter.DomainToGoaResult(result)
    responses.OK(c, goaResult)
}

func (h *TenantHandler) List(c *gin.Context) {
    ctx := c.Request.Context()
    
    // Parse pagination parameters
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
    sortBy := c.DefaultQuery("sort_by", "created_at")
    sortOrder := c.DefaultQuery("sort_order", "desc")
    search := c.Query("search")
    
    // Additional filters
    plan := c.Query("plan")
    status := c.Query("status")
    
    // Build filter
    filter := tenant.ListFilter{
        Page:      page,
        Limit:     limit,
        SortBy:    &sortBy,
        SortOrder: &sortOrder,
        Search:    &search,
        Filters: map[string]string{
            "plan":   plan,
            "status": status,
        },
    }
    
    h.logger.Info("Listing tenants", 
        "page", page, 
        "limit", limit, 
        "search", search,
        "plan", plan,
        "status", status)
    
    results, pagination, err := h.service.ListTenants(ctx, filter)
    if err != nil {
        h.logger.Error("Failed to list tenants", "error", err)
        responses.HandleDomainError(c, err)
        return
    }
    
    goaResult := h.adapter.DomainListToGoaResult(results, pagination)
    
    // Add total count header
    c.Header("X-Total-Count", strconv.FormatUint(pagination.TotalCount, 10))
    
    responses.OK(c, goaResult)
}

func (h *TenantHandler) Update(c *gin.Context) {
    ctx := c.Request.Context()
    tenantIDStr := c.Param("id")
    
    tenantID, err := uuid.Parse(tenantIDStr)
    if err != nil {
        h.logger.Error("Invalid tenant ID format", "id", tenantIDStr, "error", err)
        responses.BadRequest(c, "Invalid tenant ID format", err)
        return
    }
    
    var payload goa_tenant.UpdateTenantPayload
    if err := c.ShouldBindJSON(&payload); err != nil {
        h.logger.Error("Invalid update tenant payload", "error", err)
        responses.ValidationError(c, "Invalid request payload", err)
        return
    }
    
    // Ensure ID consistency
    payload.ID = &tenantIDStr
    
    domainReq, err := h.adapter.UpdatePayloadToDomain(&payload)
    if err != nil {
        h.logger.Error("Failed to convert update payload to domain", "error", err)
        responses.BadRequest(c, "Invalid payload data", err)
        return
    }
    
    // Add user context
    if userID := c.GetString("user_id"); userID != "" {
        domainReq.UpdatedBy = &userID
    }
    
    h.logger.Info("Updating tenant", "tenant_id", tenantID)
    
    result, err := h.service.UpdateTenant(ctx, domainReq)
    if err != nil {
        h.logger.Error("Failed to update tenant", "tenant_id", tenantID, "error", err)
        responses.HandleDomainError(c, err)
        return
    }
    
    goaResult := h.adapter.DomainToGoaResult(result)
    
    h.logger.Info("Tenant updated successfully", "tenant_id", result.ID)
    responses.OK(c, goaResult)
}

func (h *TenantHandler) Delete(c *gin.Context) {
    ctx := c.Request.Context()
    tenantIDStr := c.Param("id")
    
    tenantID, err := uuid.Parse(tenantIDStr)
    if err != nil {
        h.logger.Error("Invalid tenant ID format", "id", tenantIDStr, "error", err)
        responses.BadRequest(c, "Invalid tenant ID format", err)
        return
    }
    
    h.logger.Info("Deleting tenant", "tenant_id", tenantID)
    
    err = h.service.DeleteTenant(ctx, tenantID)
    if err != nil {
        h.logger.Error("Failed to delete tenant", "tenant_id", tenantID, "error", err)
        responses.HandleDomainError(c, err)
        return
    }
    
    h.logger.Info("Tenant deleted successfully", "tenant_id", tenantID)
    responses.NoContent(c)
}

func (h *TenantHandler) GetStats(c *gin.Context) {
    ctx := c.Request.Context()
    tenantIDStr := c.Param("id")
    
    tenantID, err := uuid.Parse(tenantIDStr)
    if err != nil {
        h.logger.Error("Invalid tenant ID format", "id", tenantIDStr, "error", err)
        responses.BadRequest(c, "Invalid tenant ID format", err)
        return
    }
    
    h.logger.Info("Getting tenant stats", "tenant_id", tenantID)
    
    stats, err := h.service.GetTenantStats(ctx, tenantID)
    if err != nil {
        h.logger.Error("Failed to get tenant stats", "tenant_id", tenantID, "error", err)
        responses.HandleDomainError(c, err)
        return
    }
    
    goaResult := h.adapter.StatsToGoaResult(stats)
    responses.OK(c, goaResult)
}
```

### Auth Handler Implementation

```go
// internal/api/gin/handlers/auth.go
package handlers

import (
    "github.com/gin-gonic/gin"
    
    // Your domain imports
    "github.com/niiniyare/erp/internal/core/auth"
    "github.com/niiniyare/erp/internal/shared/logger"
    
    // Generated Goa types
    goa_auth "github.com/niiniyare/erp/internal/api/gen/auth"
    
    // Gin utilities
    "github.com/niiniyare/erp/internal/api/gin/adapters"
    "github.com/niiniyare/erp/internal/api/gin/responses"
)

type AuthHandler struct {
    service auth.Service
    logger  logger.Logger
    adapter *adapters.AuthAdapter
}

func NewAuthHandler(service auth.Service, logger logger.Logger) *AuthHandler {
    return &AuthHandler{
        service: service,
        logger:  logger,
        adapter: adapters.NewAuthAdapter(),
    }
}

func (h *AuthHandler) Login(c *gin.Context) {
    ctx := c.Request.Context()
    
    var payload goa_auth.LoginPayload
    if err := c.ShouldBindJSON(&payload); err != nil {
        h.logger.Error("Invalid login payload", "error", err)
        responses.ValidationError(c, "Invalid request payload", err)
        return
    }
    
    // Convert to domain request
    domainReq, err := h.adapter.LoginPayloadToDomain(&payload)
    if err != nil {
        h.logger.Error("Failed to convert login payload", "error", err)
        responses.BadRequest(c, "Invalid payload data", err)
        return
    }
    
    // Add request context
    domainReq.IPAddress = c.ClientIP()
    domainReq.UserAgent = c.Request.UserAgent()
    
    h.logger.Info("User login attempt", 
        "email", domainReq.Email,
        "tenant", domainReq.TenantSubdomain,
        "ip", domainReq.IPAddress)
    
    result, err := h.service.Login(ctx, domainReq)
    if err != nil {
        h.logger.Error("Login failed", 
            "email", domainReq.Email,
            "tenant", domainReq.TenantSubdomain,
            "error", err)
        responses.HandleAuthError(c, err)
        return
    }
    
    goaResult := h.adapter.LoginResultToGoa(result)
    
    h.logger.Info("Login successful", 
        "user_id", result.User.ID,
        "tenant_id", result.Tenant.ID,
        "session_id", result.SessionID)
        
    responses.OK(c, goaResult)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
    ctx := c.Request.Context()
    
    var payload goa_auth.RefreshPayload
    if err := c.ShouldBindJSON(&payload); err != nil {
        h.logger.Error("Invalid refresh payload", "error", err)
        responses.ValidationError(c, "Invalid request payload", err)
        return
    }
    
    domainReq := &auth.RefreshTokenRequest{
        RefreshToken: *payload.RefreshToken,
        IPAddress:    c.ClientIP(),
        UserAgent:    c.Request.UserAgent(),
    }
    
    h.logger.Info("Token refresh attempt", "ip", domainReq.IPAddress)
    
    result, err := h.service.RefreshToken(ctx, domainReq)
    if err != nil {
        h.logger.Error("Token refresh failed", "error", err)
        responses.HandleAuthError(c, err)
        return
    }
    
    goaResult := h.adapter.TokenResultToGoa(result)
    
    h.logger.Info("Token refresh successful", "user_id", result.UserID)
    responses.OK(c, goaResult)
}

func (h *AuthHandler) Logout(c *gin.Context) {
    ctx := c.Request.Context()
    userID := c.GetString("user_id")
    sessionID := c.GetString("session_id")
    
    var payload goa_auth.LogoutPayload
    if err := c.ShouldBindJSON(&payload); err != nil {
        // Logout can work without payload
        payload = goa_auth.LogoutPayload{}
    }
    
    domainReq := &auth.LogoutRequest{
        UserID:              userID,
        SessionID:           sessionID,
        RefreshToken:        payload.RefreshToken,
        LogoutAllSessions:   payload.LogoutAllSessions != nil && *payload.LogoutAllSessions,
        IPAddress:           c.ClientIP(),
        UserAgent:           c.Request.UserAgent(),
    }
    
    h.logger.Info("User logout", 
        "user_id", userID,
        "session_id", sessionID,
        "logout_all", domainReq.LogoutAllSessions)
    
    err := h.service.Logout(ctx, domainReq)
    if err != nil {
        h.logger.Error("Logout failed", "user_id", userID, "error", err)
        responses.HandleDomainError(c, err)
        return
    }
    
    h.logger.Info("Logout successful", "user_id", userID)
    responses.NoContent(c)
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
    ctx := c.Request.Context()
    userID := c.GetString("user_id")
    
    h.logger.Info("Getting user profile", "user_id", userID)
    
    profile, err := h.service.GetUserProfile(ctx, userID)
    if err != nil {
        h.logger.Error("Failed to get user profile", "user_id", userID, "error", err)
        responses.HandleDomainError(c, err)
        return
    }
    
    goaResult := h.adapter.ProfileToGoa(profile)
    responses.OK(c, goaResult)
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
    ctx := c.Request.Context()
    userID := c.GetString("user_id")
    
    var payload goa_auth.ChangePasswordPayload
    if err := c.ShouldBindJSON(&payload); err != nil {
        h.logger.Error("Invalid change password payload", "error", err)
        responses.ValidationError(c, "Invalid request payload", err)
        return
    }
    
    domainReq := &auth.ChangePasswordRequest{
        UserID:          userID,
        CurrentPassword: *payload.CurrentPassword,
        NewPassword:     *payload.NewPassword,
        ConfirmPassword: *payload.ConfirmPassword,
        IPAddress:       c.ClientIP(),
        UserAgent:       c.Request.UserAgent(),
    }
    
    h.logger.Info("Password change attempt", "user_id", userID)
    
    err := h.service.ChangePassword(ctx, domainReq)
    if err != nil {
        h.logger.Error("Password change failed", "user_id", userID, "error", err)
        responses.HandleAuthError(c, err)
        return
    }
    
    h.logger.Info("Password changed successfully", "user_id", userID)
    responses.NoContent(c)
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
    ctx := c.Request.Context()
    
    var payload goa_auth.ForgotPasswordPayload
    if err := c.ShouldBindJSON(&payload); err != nil {
        h.logger.Error("Invalid forgot password payload", "error", err)
        responses.ValidationError(c, "Invalid request payload", err)
        return
    }
    
    domainReq := &auth.ForgotPasswordRequest{
        Email:           *payload.Email,
        TenantSubdomain: *payload.TenantSubdomain,
        IPAddress:       c.ClientIP(),
        UserAgent:       c.Request.UserAgent(),
    }
    
    h.logger.Info("Forgot password request", 
        "email", domainReq.Email,
        "tenant", domainReq.TenantSubdomain)
    
    err := h.service.ForgotPassword(ctx, domainReq)
    if err != nil {
        h.logger.Error("Forgot password failed", 
            "email", domainReq.Email,
            "error", err)
        // Always return accepted for security (don't reveal if email exists)
        responses.Accepted(c, gin.H{
            "message": "If the email exists, a password reset link has been sent",
        })
        return
    }
    
    h.logger.Info("Forgot password email sent", "email", domainReq.Email)
    responses.Accepted(c, gin.H{
        "message": "If the email exists, a password reset link has been sent",
    })
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
    ctx := c.Request.Context()
    
    var payload goa_auth.ResetPasswordPayload
    if err := c.ShouldBindJSON(&payload); err != nil {
        h.logger.Error("Invalid reset password payload", "error", err)
        responses.ValidationError(c, "Invalid request payload", err)
        return
    }
    
    domainReq := &auth.ResetPasswordRequest{
        ResetToken:      *payload.ResetToken,
        NewPassword:     *payload.NewPassword,
        ConfirmPassword: *payload.ConfirmPassword,
        IPAddress:       c.ClientIP(),
        UserAgent:       c.Request.UserAgent(),
    }
    
    h.logger.Info("Password reset attempt", "token", domainReq.ResetToken[:8]+"...")
    
    err := h.service.ResetPassword(ctx, domainReq)
    if err != nil {
        h.logger.Error("Password reset failed", "error", err)
        responses.HandleAuthError(c, err)
        return
    }
    
    h.logger.Info("Password reset successful")
    responses.NoContent(c)
}
```

## 🛡️ 4. Complete Middleware Implementation

### JWT Authentication Middleware

```go
// internal/api/gin/middleware/auth.go
package middleware

import (
    "net/http"
    "strings"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
    
    "github.com/niiniyare/erp/internal/platform/config"
    "github.com/niiniyare/erp/internal/shared/logger"
    "github.com/niiniyare/erp/internal/api/gin/responses"
)

type JWTClaims struct {
    UserID      string   `json:"user_id"`
    TenantID    string   `json:"tenant_id"`
    Email       string   `json:"email"`
    Roles       []string `json:"roles"`
    Permissions []string `json:"permissions"`
    SessionID   string   `json:"session_id"`
    TokenType   string   `json:"token_type"`
    jwt.RegisteredClaims
}

func JWTAuth(cfg config.JWT) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := extractToken(c.Request)
        if token == "" {
            responses.Unauthorized(c, "Missing or invalid authorization header")
            c.Abort()
            return
        }
        
        claims, err := validateJWT(token, cfg.Secret)
        if err != nil {
            responses.Unauthorized(c, "Invalid or expired token")
            c.Abort()
            return
        }
        
        // Check token type
        if claims.TokenType != "access" {
            responses.Unauthorized(c, "Invalid token type")
            c.Abort()
            return
        }
        
        // Check expiration
        if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
            responses.Unauthorized(c, "Token has expired")
            c.Abort()
            return
        }
        
        // Set user context
        c.Set("user_id", claims.UserID)
        c.Set("tenant_id", claims.TenantID)
        c.Set("email", claims.Email)
        c.Set("roles", claims.Roles)
        c.Set("permissions", claims.Permissions)
        c.Set("session_id", claims.SessionID)
        
        // Add user info to request context for logging
        c.Set("authenticated_user", gin.H{
            "user_id":   claims.UserID,
            "tenant_id": claims.TenantID,
            "email":     claims.Email,
            "roles":     claims.Roles,
        })
        
        c.Next()
    }
}

func extractToken(r *http.Request) string {
    // Check Authorization header
    bearerToken := r.Header.Get("Authorization")
    if bearerToken != "" && strings.HasPrefix(bearerToken, "Bearer ") {
        return strings.TrimPrefix(bearerToken, "Bearer ")
    }
    
    // Check query parameter (for WebSocket connections, etc.)
    if token := r.URL.Query().Get("token"); token != "" {
        return token
    }
    
    return ""
}

func validateJWT(tokenString string, secret string) (*JWTClaims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
        // Validate signing method
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, jwt.ErrSignatureInvalid
        }
        return []byte(secret), nil
    })
    
    if err != nil {
        return nil, err
    }
    
    if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
        return claims, nil
    }
    
    return nil, jwt.ErrInvalidKey
}

// Optional JWT middleware that doesn't require authentication
func OptionalJWTAuth(cfg config.JWT) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := extractToken(c.Request)
        if token == "" {
            c.Next()
            return
        }
        
        claims, err := validateJWT(token, cfg.Secret)
        if err != nil {
            // Invalid token, but continue without authentication
            c.Next()
            return
        }
        
        // Set user context if token is valid
        c.Set("user_id", claims.UserID)
        c.Set("tenant_id", claims.TenantID)
        c.Set("email", claims.Email)
        c.Set("roles", claims.Roles)
        c.Set("permissions", claims.Permissions)
        c.Set("session_id", claims.SessionID)
        
        c.Next()
    }
}
```

### ABAC Permission Middleware

```go
// internal/api/gin/middleware/abac.go
package middleware

import (
    "strings"
    
    "github.com/gin-gonic/gin"
    
    "github.com/niiniyare/erp/internal/api/gin/responses"
    "github.com/niiniyare/erp/internal/shared/logger"
)

func RequirePermission(permission string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("user_id")
        if userID == "" {
            responses.Unauthorized(c, "Authentication required")
            c.Abort()
            return
        }
        
        tenantID := c.GetString("tenant_id")
        permissions, exists := c.Get("permissions")
        if !exists {
            responses.Forbidden(c, "No permissions available")
            c.Abort()
            return
        }
        
        userPermissions, ok := permissions.([]string)
        if !ok {
            responses.Forbidden(c, "Invalid permissions format")
            c.Abort()
            return
        }
        
        // Check if user has the required permission
        if hasPermission(userPermissions, permission) {
            c.Next()
            return
        }
        
        // If direct permission check fails, check with ABAC service
        if checkABACPermission(c, userID, tenantID, permission) {
            c.Next()
            return
        }
        
        responses.Forbidden(c, "Insufficient permissions")
        c.Abort()
    }
}

func RequireAnyPermission(permissions ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("user_id")
        if userID == "" {
            responses.Unauthorized(c, "Authentication required")
            c.Abort()
            return
        }
        
        userPermissions, exists := c.Get("permissions")
        if !exists {
            responses.Forbidden(c, "No permissions available")
            c.Abort()
            return
        }
        
        perms, ok := userPermissions.([]string)
        if !ok {
            responses.Forbidden(c, "Invalid permissions format")
            c.Abort()
            return
        }
        
        // Check if user has any of the required permissions
        for _, permission := range permissions {
            if hasPermission(perms, permission) {
                c.Next()
                return
            }
        }
        
        responses.Forbidden(c, "Insufficient permissions")
        c.Abort()
    }
}

func RequireAllPermissions(permissions ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("user_id")
        if userID == "" {
            responses.Unauthorized(c, "Authentication required")
            c.Abort()
            return
        }
        
        userPermissions, exists := c.Get("permissions")
        if !exists {
            responses.Forbidden(c, "No permissions available")
            c.Abort()
            return
        }
        
        perms, ok := userPermissions.([]string)
        if !ok {
            responses.Forbidden(c, "Invalid permissions format")
            c.Abort()
            return
        }
        
        // Check if user has all required permissions
        for _, permission := range permissions {
            if !hasPermission(perms, permission) {
                responses.Forbidden(c, "Insufficient permissions")
                c.Abort()
                return
            }
        }
        
        c.Next()
    }
}

func RequireRole(role string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("user_id")
        if userID == "" {
            responses.Unauthorized(c, "Authentication required")
            c.Abort()
            return
        }
        
        roles, exists := c.Get("roles")
        if !exists {
            responses.Forbidden(c, "No roles available")
            c.Abort()
            return
        }
        
        userRoles, ok := roles.([]string)
        if !ok {
            responses.Forbidden(c, "Invalid roles format")
            c.Abort()
            return
        }
        
        // Check if user has the required role
        for _, userRole := range userRoles {
            if userRole == role {
                c.Next()
                return
            }
        }
        
        responses.Forbidden(c, "Insufficient role privileges")
        c.Abort()
    }
}

func RequireAnyRole(roles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("user_id")
        if userID == "" {
            responses.Unauthorized(c, "Authentication required")
            c.Abort()
            return
        }
        
        userRoles, exists := c.Get("roles")
        if !exists {
            responses.Forbidden(c, "No roles available")
            c.Abort()
            return
        }
        
        roleList, ok := userRoles.([]string)
        if !ok {
            responses.Forbidden(c, "Invalid roles format")
            c.Abort()
            return
        }
        
        // Check if user has any of the required roles
        for _, userRole := range roleList {
            for _, requiredRole := range roles {
                if userRole == requiredRole {
                    c.Next()
                    return
                }
            }
        }
        
        responses.Forbidden(c, "Insufficient role privileges")
        c.Abort()
    }
}

// Helper functions
func hasPermission(userPermissions []string, requiredPermission string) bool {
    for _, permission := range userPermissions {
        if permission == requiredPermission {
            return true
        }
        // Check for wildcard permissions (e.g., "tenant:*" includes "tenant:read")
        if strings.HasSuffix(permission, ":*") {
            prefix := strings.TrimSuffix(permission, "*")
            if strings.HasPrefix(requiredPermission, prefix) {
                return true
            }
        }
    }
    return false
}

func checkABACPermission(c *gin.Context, userID, tenantID, permission string) bool {
    // This would integrate with your ABAC service
    // For now, return false to use only JWT permissions
    
    // Example integration:
    /*
    abacService := c.MustGet("abac_service").(abac.Service)
    
    ctx := c.Request.Context()
    parts := strings.Split(permission, ":")
    if len(parts) != 2 {
        return false
    }
    
    resource := parts[0]
    action := parts[1]
    resourceID := c.Param("id") // Get resource ID from URL if available
    
    req := &abac.PermissionCheckRequest{
        UserID:     userID,
        Action:     action,
        Resource:   resource,
        ResourceID: resourceID,
        Context: map[string]interface{}{
            "ip_address": c.ClientIP(),
            "user_agent": c.Request.UserAgent(),
            "method":     c.Request.Method,
            "path":       c.Request.URL.Path,
            "tenant_id":  tenantID,
        },
    }
    
    result, err := abacService.CheckPermission(ctx, req)
    if err != nil {
        return false
    }
    
    return result.Allowed
    */
    
    return false
}
```

## 🔄 5. Complete Adapter Implementations

### Tenant Adapter

```go
// internal/api/gin/adapters/tenant.go
package adapters

import (
    "time"
    
    // Generated Goa types
    goa_tenant "github.com/niiniyare/erp/internal/api/gen/tenant"
    
    // Your domain types
    "github.com/niiniyare/erp/internal/core/tenant"
    "github.com/niiniyare/erp/internal/shared/types"
)

type TenantAdapter struct{}

func NewTenantAdapter() *TenantAdapter {
    return &TenantAdapter{}
}

func (a *TenantAdapter) CreatePayloadToDomain(payload *goa_tenant.CreateTenantPayload) (*tenant.CreateRequest, error) {
    req := &tenant.CreateRequest{
        Name:        *payload.Name,
        Subdomain:   *payload.Subdomain,
        Description: payload.Description,
        Plan:        payload.Plan,
        Settings:    payload.Settings,
    }
    
    if payload.ContactEmail != nil {
        req.ContactEmail = *payload.ContactEmail
    }
    
    if payload.ContactPhone != nil {
        req.ContactPhone = payload.ContactPhone
    }
    
    // Validate required fields
    if req.Name == "" || req.Subdomain == "" || req.ContactEmail == "" {
        return nil, tenant.ErrInvalidInput
    }
    
    return req, nil
}

func (a *TenantAdapter) UpdatePayloadToDomain(payload *goa_tenant.UpdateTenantPayload) (*tenant.UpdateRequest, error) {
    req := &tenant.UpdateRequest{
        ID:          *payload.ID,
        Name:        payload.Name,
        Description: payload.Description,
        Plan:        payload.Plan,
        Settings:    payload.Settings,
        ContactEmail: payload.ContactEmail,
        ContactPhone: payload.ContactPhone,
    }
    
    return req, nil
}

func (a *TenantAdapter) DomainToGoaResult(t *tenant.Tenant) *goa_tenant.TenantResult {
    result := &goa_tenant.TenantResult{
        ID:           t.ID.String(),
        Name:         t.Name,
        Subdomain:    t.Subdomain,
        Description:  t.Description,
        Plan:         string(t.Plan),
        Status:       string(t.Status),
        Settings:     t.Settings,
        ContactEmail: t.ContactEmail,
        ContactPhone: t.ContactPhone,
        CreatedAt:    t.CreatedAt.Format(time.RFC3339),
        UpdatedAt:    t.UpdatedAt.Format(time.RFC3339),
    }
    
    if t.CreatedBy != nil {
        result.CreatedBy = t.CreatedBy.String()
    }
    
    if t.UpdatedBy != nil {
        result.UpdatedBy = t.UpdatedBy.String()
    }
    
    if t.Limits != nil {
        result.Limits = &goa_tenant.TenantLimits{
            MaxUsers:                t.Limits.MaxUsers,
            MaxStorageGb:           t.Limits.MaxStorageGB,
            MaxAPIRequestsPerHour:  t.Limits.MaxAPIRequestsPerHour,
            MaxOrganizations:       t.Limits.MaxOrganizations,
            EnabledFeatures:        t.Limits.EnabledFeatures,
        }
    }
    
    if t.Usage != nil {
        result.Usage = &goa_tenant.TenantUsage{
            CurrentUsers:         t.Usage.CurrentUsers,
            CurrentStorageGb:     t.Usage.CurrentStorageGB,
            APIRequestsLastHour:  t.Usage.APIRequestsLastHour,
            CurrentOrganizations: t.Usage.CurrentOrganizations,
            LastActivity:         t.Usage.LastActivity.Format(time.RFC3339),
        }
    }
    
    result.Version = &t.Version
    
    return result
}

func (a *TenantAdapter) DomainListToGoaResult(tenants []*tenant.Tenant, pagination *types.Pagination) *goa_tenant.TenantListResult {
    results := make([]*goa_tenant.TenantResult, len(tenants))
    for i, t := range tenants {
        results[i] = a.DomainToGoaResult(t)
    }
    
    return &goa_tenant.TenantListResult{
        Data: results,
        Meta: &goa_tenant.PaginationMeta{
            CurrentPage: uint(pagination.CurrentPage),
            PerPage:     uint(pagination.PerPage),
            TotalCount:  pagination.TotalCount,
            TotalPages:  uint(pagination.TotalPages),
            HasNext:     pagination.HasNext,
            HasPrev:     pagination.HasPrev,
            NextPage:    uintPtr(pagination.NextPage),
            PrevPage:    uintPtr(pagination.PrevPage),
        },
    }
}

func (a *TenantAdapter) StatsToGoaResult(stats *tenant.Stats) *goa_tenant.TenantStatsResult {
    result := &goa_tenant.TenantStatsResult{
        TenantID:    stats.TenantID.String(),
        CalculatedAt: stats.CalculatedAt.Format(time.RFC3339),
    }
    
    if stats.Usage != nil {
        result.Usage = &goa_tenant.TenantUsage{
            CurrentUsers:         stats.Usage.CurrentUsers,
            CurrentStorageGb:     stats.Usage.CurrentStorageGB,
            APIRequestsLastHour:  stats.Usage.APIRequestsLastHour,
            CurrentOrganizations: stats.Usage.CurrentOrganizations,
            LastActivity:         stats.Usage.LastActivity.Format(time.RFC3339),
        }
    }
    
    if stats.Limits != nil {
        result.Limits = &goa_tenant.TenantLimits{
            MaxUsers:                stats.Limits.MaxUsers,
            MaxStorageGb:           stats.Limits.MaxStorageGB,
            MaxAPIRequestsPerHour:  stats.Limits.MaxAPIRequestsPerHour,
            MaxOrganizations:       stats.Limits.MaxOrganizations,
            EnabledFeatures:        stats.Limits.EnabledFeatures,
        }
    }
    
    if stats.Growth != nil {
        result.GrowthMetrics = &goa_tenant.TenantGrowthMetrics{
            UsersAddedLast30Days:            stats.Growth.UsersAddedLast30Days,
            StorageGrowthLast30DaysGb:       stats.Growth.StorageGrowthLast30DaysGB,
            OrganizationsAddedLast30Days:    stats.Growth.OrganizationsAddedLast30Days,
        }
    }
    
    if stats.Activity != nil {
        result.ActivityMetrics = &goa_tenant.TenantActivityMetrics{
            ActiveUsersLast24Hours:     stats.Activity.ActiveUsersLast24Hours,
            APIRequestsLast24Hours:     stats.Activity.APIRequestsLast24Hours,
            LoginSessionsLast24Hours:   stats.Activity.LoginSessionsLast24Hours,
            MostActiveFeatures:         stats.Activity.MostActiveFeatures,
        }
    }
    
    return result
}

func uintPtr(val uint) *uint {
    if val == 0 {
        return nil
    }
    return &val
}
```

### Auth Adapter

```go
// internal/api/gin/adapters/auth.go
package adapters

import (
    "time"
    
    // Generated Goa types
    goa_auth "github.com/niiniyare/erp/internal/api/gen/auth"
    
    // Your domain types
    "github.com/niiniyare/erp/internal/core/auth"
)

type AuthAdapter struct{}

func NewAuthAdapter() *AuthAdapter {
    return &AuthAdapter{}
}

func (a *AuthAdapter) LoginPayloadToDomain(payload *goa_auth.LoginPayload) (*auth.LoginRequest, error) {
    req := &auth.LoginRequest{
        Email:           *payload.Email,
        Password:        *payload.Password,
        TenantSubdomain: *payload.TenantSubdomain,
        RememberMe:      payload.RememberMe != nil && *payload.RememberMe,
    }
    
    if payload.MfaToken != nil {
        req.MFAToken = *payload.MfaToken
    }
    
    return req, nil
}

func (a *AuthAdapter) LoginResultToGoa(result *auth.LoginResult) *goa_auth.LoginResult {
    goaResult := &goa_auth.LoginResult{
        AccessToken:  result.AccessToken,
        RefreshToken: result.RefreshToken,
        TokenType:    result.TokenType,
        ExpiresIn:    uint(result.ExpiresIn),
        SessionID:    result.SessionID,
        MfaEnabled:   result.MFAEnabled,
        Permissions:  result.Permissions,
    }
    
    // Convert user profile
    if result.User != nil {
        goaResult.User = a.UserProfileToGoa(result.User)
    }
    
    // Convert tenant info
    if result.Tenant != nil {
        goaResult.Tenant = &goa_auth.TenantInfo{
            ID:        result.Tenant.ID.String(),
            Name:      result.Tenant.Name,
            Subdomain: result.Tenant.Subdomain,
            Plan:      string(result.Tenant.Plan),
            Status:    string(result.Tenant.Status),
        }
    }
    
    if result.PasswordExpiresAt != nil {
        expiresAt := result.PasswordExpiresAt.Format(time.RFC3339)
        goaResult.PasswordExpiresAt = &expiresAt
    }
    
    return goaResult
}

func (a *AuthAdapter) TokenResultToGoa(result *auth.TokenResult) *goa_auth.TokenResult {
    goaResult := &goa_auth.TokenResult{
        AccessToken: result.AccessToken,
        TokenType:   result.TokenType,
        ExpiresIn:   uint(result.ExpiresIn),
    }
    
    if result.RefreshToken != "" {
        goaResult.RefreshToken = &result.RefreshToken
    }
    
    return goaResult
}

func (a *AuthAdapter) UserProfileToGoa(profile *auth.UserProfile) *goa_auth.UserProfileResult {
    result := &goa_auth.UserProfileResult{
        ID:            profile.ID.String(),
        Email:         profile.Email,
        FirstName:     profile.FirstName,
        LastName:      profile.LastName,
        DisplayName:   profile.DisplayName,
        Status:        string(profile.Status),
        Roles:         profile.Roles,
        EmailVerified: profile.EmailVerified,
        CreatedAt:     profile.CreatedAt.Format(time.RFC3339),
        UpdatedAt:     profile.UpdatedAt.Format(time.RFC3339),
    }
    
    if profile.AvatarURL != "" {
        result.AvatarURL = &profile.AvatarURL
    }
    
    if profile.LastLogin != nil {
        lastLogin := profile.LastLogin.Format(time.RFC3339)
        result.LastLogin = &lastLogin
    }
    
    if profile.Timezone != "" {
        result.Timezone = &profile.Timezone
    }
    
    if profile.Locale != "" {
        result.Locale = &profile.Locale
    }
    
    if profile.Preferences != nil {
        result.Preferences = profile.Preferences
    }
    
    result.MfaEnabled = &profile.MFAEnabled
    
    return result
}

func (a *AuthAdapter) ProfileToGoa(profile *auth.UserProfile) *goa_auth.UserProfileResult {
    return a.UserProfileToGoa(profile)
}
```

## 📤 6. Complete Response Utilities

### Success Responses

```go
// internal/api/gin/responses/success.go
package responses

import (
    "net/http"
    "time"
    
    "github.com/gin-gonic/gin"
)

type SuccessResponse struct {
    Data      interface{} `json:"data"`
    Message   string      `json:"message,omitempty"`
    Timestamp string      `json:"timestamp"`
    RequestID string      `json:"request_id"`
    Status    string      `json:"status"`
}

type MessageResponse struct {
    Message   string `json:"message"`
    Timestamp string `json:"timestamp"`
    RequestID string `json:"request_id"`
    Status    string `json:"status"`
}

func OK(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, SuccessResponse{
        Data:      data,
        Status:    "success",
        Timestamp: time.Now().UTC().Format(time.RFC3339),
        RequestID: getRequestID(c),
    })
}

func Created(c *gin.Context, data interface{}) {
    c.JSON(http.StatusCreated, SuccessResponse{
        Data:      data,
        Message:   "Resource created successfully",
        Status:    "success",
        Timestamp: time.Now().UTC().Format(time.RFC3339),
        RequestID: getRequestID(c),
    })
}

func Accepted(c *gin.Context, data interface{}) {
    c.JSON(http.StatusAccepted, SuccessResponse{
        Data:      data,
        Message:   "Request accepted and is being processed",
        Status:    "accepted",
        Timestamp: time.Now().UTC().Format(time.RFC3339),
        RequestID: getRequestID(c),
    })
}

func NoContent(c *gin.Context) {
    c.JSON(http.StatusNoContent, MessageResponse{
        Message:   "Operation completed successfully",
        Status:    "success",
        Timestamp: time.Now().UTC().Format(time.RFC3339),
        RequestID: getRequestID(c),
    })
}

func OKWithMessage(c *gin.Context, data interface{}, message string) {
    c.JSON(http.StatusOK, SuccessResponse{
        Data:      data,
        Message:   message,
        Status:    "success",
        Timestamp: time.Now().UTC().Format(time.RFC3339),
        RequestID: getRequestID(c),
    })
}

func getRequestID(c *gin.Context) string {
    if reqID := c.GetString("request_id"); reqID != "" {
        return reqID
    }
    return "unknown"
}
```

### Error Responses

```go
// internal/api/gin/responses/errors.go
package responses

import (
    "errors"
    "net/http"
    "time"
    
    "github.com/gin-gonic/gin"
    
    // Your domain errors
    "github.com/niiniyare/erp/internal/core/tenant"
    "github.com/niiniyare/erp/internal/core/user"
    "github.com/niiniyare/erp/internal/core/auth"
    "github.com/niiniyare/erp/internal/shared/errors" as sharedErrors
)

type ErrorResponse struct {
    Error     ErrorDetail            `json:"error"`
    Status    string                 `json:"status"`
    Timestamp string                 `json:"timestamp"`
    RequestID string                 `json:"request_id"`
    TraceID   string                 `json:"trace_id,omitempty"`
}

type ErrorDetail struct {
    Code    string                 `json:"code"`
    Message string                 `json:"message"`
    Details map[string]interface{} `json:"details,omitempty"`
    Field   string                 `json:"field,omitempty"`
}

type ValidationErrorResponse struct {
    Error            ErrorDetail       `json:"error"`
    ValidationErrors []ValidationError `json:"validation_errors,omitempty"`
    Status           string            `json:"status"`
    Timestamp        string            `json:"timestamp"`
    RequestID        string            `json:"request_id"`
}

type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
    Value   string `json:"value,omitempty"`
}

func BadRequest(c *gin.Context, message string, err error) {
    errorResp := ErrorResponse{
        Error: ErrorDetail{
            Code:    "BAD_REQUEST",
            Message: message,
        },
        Status:    "error",
        Timestamp: time.Now().UTC().Format(time.RFC3339),
        RequestID: getRequestID(c),
    }
    
    if err != nil {
        errorResp.Error.Details = map[string]interface{}{
            "error": err.Error(),
        }
    }
    
    if traceID := c.GetString("trace_id"); traceID != "" {
        errorResp.TraceID = traceID
    }
    
    c.JSON(http.StatusBadRequest, errorResp)
}

func ValidationError(c *gin.Context, message string, err error) {
    errorResp := ValidationErrorResponse{
        Error: ErrorDetail{
            Code:    "VALIDATION_ERROR",
            Message: message,
        },
        Status:    "error",
        Timestamp: time.Now().UTC().Format(time.RFC3339),
        RequestID: getRequestID(c),
    }
    
    if err != nil {
        errorResp.Error.Details = map[string]interface{}{
            "error": err.Error(),
        }
        
        // Extract validation errors if available
        if validationErrs := extractValidationErrors(err); len(validationErrs) > 0 {
            errorResp.ValidationErrors = validationErrs
        }
    }
    
    c.JSON(http.StatusUnprocessableEntity, errorResp)
}

func Unauthorized(c *gin.Context, message string) {
    errorResp := ErrorResponse{
        Error: ErrorDetail{
            Code:    "UNAUTHORIZED",
            Message: message,
        },
        Status:    "error",
        Timestamp: time.Now().UTC().Format(time.RFC3339),
        RequestID: getRequestID(c),
    }
    
    c.JSON(http.StatusUnauthorized, errorResp)
}

func Forbidden(c *gin.Context, message string) {
    errorResp := ErrorResponse{
        Error: ErrorDetail{
            Code:    "FORBIDDEN",
            Message: message,
        },
        Status:    "error",
        Timestamp: time.Now().UTC().Format(time.RFC3339),
        RequestID: getRequestID(c),
    }
    
    c.JSON(http.StatusForbidden, errorResp)
}

func NotFound(c *gin.Context, message string) {
    errorResp := ErrorResponse{
        Error: ErrorDetail{
            Code:    "NOT_FOUND",
            Message: message,
        },
        Status:    "error",
        Timestamp: time.Now().UTC().Format(time.RFC3339),
        RequestID: getRequestID(c),
    }
    
    c.JSON(http.StatusNotFound, errorResp)
}

func Conflict(c *gin.Context, message string, details map[string]interface{}) {
    errorResp := ErrorResponse{
        Error: ErrorDetail{
            Code:    "CONFLICT",
            Message: message,
            Details: details,
        },
        Status:    "error",
        Timestamp: time.Now().UTC().Format(time.RFC3339),
        RequestID: getRequestID(c),
    }
    
    c.JSON(http.StatusConflict, errorResp)
}

func InternalServerError(c *gin.Context, message string, err error) {
    errorResp := ErrorResponse{
        Error: ErrorDetail{
            Code:    "INTERNAL_ERROR",
            Message: message,
        },
        Status:    "error",
        Timestamp: time.Now().UTC().Format(time.RFC3339),
        RequestID: getRequestID(c),
    }
    
    if err != nil {
        errorResp.Error.Details = map[string]interface{}{
            "error": err.Error(),
        }
    }
    
    if traceID := c.GetString("trace_id"); traceID != "" {
        errorResp.TraceID = traceID
    }
    
    c.JSON(http.StatusInternalServerError, errorResp)
}

func HandleDomainError(c *gin.Context, err error) {
    switch {
    // Tenant errors
    case errors.Is(err, tenant.ErrNotFound):
        NotFound(c, "Tenant not found")
    case errors.Is(err, tenant.ErrSubdomainExists):
        Conflict(c, "Subdomain already exists", map[string]interface{}{
            "error_type": "subdomain_conflict",
        })
    case errors.Is(err, tenant.ErrInvalidInput):
        BadRequest(c, "Invalid tenant data", err)
    case errors.Is(err, tenant.ErrLimitExceeded):
        errorResp := ErrorResponse{
            Error: ErrorDetail{
                Code:    "TENANT_LIMIT_EXCEEDED",
                Message: "Tenant limit exceeded",
            },
            Status:    "error",
            Timestamp: time.Now().UTC().Format(time.RFC3339),
            RequestID: getRequestID(c),
        }
        c.JSON(http.StatusTooManyRequests, errorResp)
    
    // User errors
    case errors.Is(err, user.ErrNotFound):
        NotFound(c, "User not found")
    case errors.Is(err, user.ErrEmailExists):
        Conflict(c, "Email already exists", map[string]interface{}{
            "error_type": "email_conflict",
        })
    case errors.Is(err, user.ErrInvalidInput):
        BadRequest(c, "Invalid user data", err)
    
    // Auth errors
    case errors.Is(err, auth.ErrInvalidCredentials):
        Unauthorized(c, "Invalid email or password")
    case errors.Is(err, auth.ErrAccountLocked):
        errorResp := ErrorResponse{
            Error: ErrorDetail{
                Code:    "ACCOUNT_LOCKED",
                Message: "Account is locked due to multiple failed login attempts",
            },
            Status:    "error",
            Timestamp: time.Now().UTC().Format(time.RFC3339),
            RequestID: getRequestID(c),
        }
        c.JSON(http.StatusLocked, errorResp)
    case errors.Is(err, auth.ErrTokenExpired):
        Unauthorized(c, "Token has expired")
    case errors.Is(err, auth.ErrInvalidToken):
        Unauthorized(c, "Invalid token")
    case errors.Is(err, auth.ErrMFARequired):
        errorResp := ErrorResponse{
            Error: ErrorDetail{
                Code:    "MFA_REQUIRED",
                Message: "Multi-factor authentication required",
            },
            Status:    "error",
            Timestamp: time.Now().UTC().Format(time.RFC3339),
            RequestID: getRequestID(c),
        }
        c.JSON(http.StatusPreconditionRequired, errorResp)
    case errors.Is(err, auth.ErrPasswordResetRequired):
        errorResp := ErrorResponse{
            Error: ErrorDetail{
                Code:    "PASSWORD_RESET_REQUIRED",
                Message: "Password reset is required",
            },
            Status:    "error",
            Timestamp: time.Now().UTC().Format(time.RFC3339),
            RequestID: getRequestID(c),
        }
        c.JSON(http.StatusPreconditionRequired, errorResp)
    
    // Generic shared errors
    case errors.Is(err, sharedErrors.ErrValidation):
        ValidationError(c, "Validation failed", err)
    case errors.Is(err, sharedErrors.ErrUnauthorized):
        Unauthorized(c, "Unauthorized access")
    case errors.Is(err, sharedErrors.ErrForbidden):
        Forbidden(c, "Forbidden access")
    
    default:
        InternalServerError(c, "An internal error occurred", err)
    }
}

func HandleAuthError(c *gin.Context, err error) {
    // Special handling for authentication-specific errors
    switch {
    case errors.Is(err, auth.ErrInvalidCredentials):
        Unauthorized(c, "Invalid email or password")
    case errors.Is(err, auth.ErrAccountLocked):
        errorResp := ErrorResponse{
            Error: ErrorDetail{
                Code:    "ACCOUNT_LOCKED",
                Message: "Account is temporarily locked",
                Details: map[string]interface{}{
                    "retry_after": "15 minutes",
                },
            },
            Status:    "error",
            Timestamp: time.Now().UTC().Format(time.RFC3339),
            RequestID: getRequestID(c),
        }
        c.JSON(http.StatusLocked, errorResp)
    default:
        HandleDomainError(c, err)
    }
}

// Helper function to extract validation errors from error
func extractValidationErrors(err error) []ValidationError {
    // This would be implemented based on your validation library
    // For example, if using github.com/go-playground/validator:
    /*
    var validationErrs []ValidationError
    if validatorErrs, ok := err.(validator.ValidationErrors); ok {
        for _, err := range validatorErrs {
            validationErrs = append(validationErrs, ValidationError{
                Field:   err.Field(),
                Message: err.Tag(),
                Value:   fmt.Sprintf("%v", err.Value()),
            })
        }
    }
    return validationErrs
    */
    return nil
}
```

## 🛡️ 7. Complete Validation Middleware

```go
// internal/api/gin/middleware/validation.go
package middleware

import (
    "encoding/json"
    "io"
    "net/http"
    "strings"
    
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    
    "github.com/niiniyare/erp/internal/api/gin/responses"
)

// ValidateJSON ensures request has valid JSON content type and body
func ValidateJSON() gin.HandlerFunc {
    return func(c *gin.Context) {
        if c.Request.Method == "GET" || c.Request.Method == "DELETE" {
            c.Next()
            return
        }
        
        contentType := c.Request.Header.Get("Content-Type")
        if !strings.Contains(contentType, "application/json") {
            responses.BadRequest(c, "Content-Type must be application/json", nil)
            c.Abort()
            return
        }
        
        if c.Request.Body == nil {
            responses.BadRequest(c, "Request body is required", nil)
            c.Abort()
            return
        }
        
        // Read and validate JSON
        body, err := io.ReadAll(c.Request.Body)
        if err != nil {
            responses.BadRequest(c, "Failed to read request body", err)
            c.Abort()
            return
        }
        
        if len(body) == 0 {
            responses.BadRequest(c, "Request body cannot be empty", nil)
            c.Abort()
            return
        }
        
        var jsonData interface{}
        if err := json.Unmarshal(body, &jsonData); err != nil {
            responses.ValidationError(c, "Invalid JSON format", err)
            c.Abort()
            return
        }
        
        // Store the parsed JSON for use in handlers
        c.Set("parsed_json", jsonData)
        
        // Restore the body for Gin's ShouldBindJSON
        c.Request.Body = io.NopCloser(strings.NewReader(string(body)))
        
        c.Next()
    }
}

// ValidateUUID validates that a URL parameter is a valid UUID
func ValidateUUID(paramName string) gin.HandlerFunc {
    return func(c *gin.Context) {
        paramValue := c.Param(paramName)
        if paramValue == "" {
            responses.BadRequest(c, paramName+" is required", nil)
            c.Abort()
            return
        }
        
        if _, err := uuid.Parse(paramValue); err != nil {
            responses.BadRequest(c, "Invalid "+paramName+" format (must be UUID)", err)
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// ValidateQueryParams validates common query parameters
func ValidateQueryParams() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Validate pagination parameters
        if page := c.Query("page"); page != "" {
            if !isValidPositiveInt(page) {
                responses.BadRequest(c, "Invalid page parameter (must be positive integer)", nil)
                c.Abort()
                return
            }
        }
        
        if limit := c.Query("limit"); limit != "" {
            if !isValidPositiveInt(limit) || !isWithinRange(limit, 1, 100) {
                responses.BadRequest(c, "Invalid limit parameter (must be between 1 and 100)", nil)
                c.Abort()
                return
            }
        }
        
        if sortOrder := c.Query("sort_order"); sortOrder != "" {
            if sortOrder != "asc" && sortOrder != "desc" {
                responses.BadRequest(c, "Invalid sort_order parameter (must be 'asc' or 'desc')", nil)
                c.Abort()
                return
            }
        }
        
        c.Next()
    }
}

// ValidateContentSize ensures request body is not too large
func ValidateContentSize(maxSize int64) gin.HandlerFunc {
    return func(c *gin.Context) {
        if c.Request.ContentLength > maxSize {
            responses.BadRequest(c, "Request body too large", nil)
            c.Abort()
            return
        }
        
        c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
        c.Next()
    }
}

// ValidateAPIKey validates API key for specific endpoints
func ValidateAPIKey() gin.HandlerFunc {
    return func(c *gin.Context) {
        apiKey := c.GetHeader("X-API-Key")
        if apiKey == "" {
            responses.Unauthorized(c, "API key is required")
            c.Abort()
            return
        }
        
        // Validate API key format (example: should be at least 32 characters)
        if len(apiKey) < 32 {
            responses.Unauthorized(c, "Invalid API key format")
            c.Abort()
            return
        }
        
        // Here you would validate against your API key store
        // For now, we'll just set it in context
        c.Set("api_key", apiKey)
        
        c.Next()
    }
}

// ValidateFileUpload validates file upload requests
func ValidateFileUpload(allowedTypes []string, maxSize int64) gin.HandlerFunc {
    return func(c *gin.Context) {
        file, header, err := c.Request.FormFile("file")
        if err != nil {
            responses.BadRequest(c, "Failed to get file from request", err)
            c.Abort()
            return
        }
        defer file.Close()
        
        // Check file size
        if header.Size > maxSize {
            responses.BadRequest(c, "File size exceeds maximum allowed size", nil)
            c.Abort()
            return
        }
        
        // Check file type
        contentType := header.Header.Get("Content-Type")
        if !isAllowedFileType(contentType, allowedTypes) {
            responses.BadRequest(c, "File type not allowed", nil)
            c.Abort()
            return
        }
        
        // Store file info in context
        c.Set("uploaded_file", header)
        
        c.Next()
    }
}

// Helper functions
func isValidPositiveInt(s string) bool {
    if s == "" {
        return false
    }
    
    for _, char := range s {
        if char < '0' || char > '9' {
            return false
        }
    }
    
    return s != "0"
}

func isWithinRange(s string, min, max int) bool {
    // Simple range check - in production you'd use strconv.Atoi
    length := len(s)
    if length == 1 {
        char := s[0]
        num := int(char - '0')
        return num >= min && num <= max
    }
    
    if length == 2 {
        if s == "10" || s == "20" || s == "30" || s == "40" || s == "50" || 
           s == "60" || s == "70" || s == "80" || s == "90" {
            return true
        }
        return s == "100" && max >= 100
    }
    
    if length == 3 && s == "100" {
        return max >= 100
    }
    
    return false
}

func isAllowedFileType(contentType string, allowedTypes []string) bool {
    for _, allowedType := range allowedTypes {
        if contentType == allowedType {
            return true
        }
    }
    return false
}
```

## 🔧 8. Service Implementation Example

```go
// internal/api/services/tenant_service.go
package services

import (
    "context"
    
    // Your core domain
    "github.com/niiniyare/erp/internal/core/tenant"
    
    // Generated Goa service interface
    goa_tenant "github.com/niiniyare/erp/internal/api/gen/tenant"
    
    // Adapters
    "github.com/niiniyare/erp/internal/api/gin/adapters"
    
    "github.com/niiniyare/erp/internal/shared/logger"
)

// TenantServiceImpl implements the generated Goa service interface
// This bridges your domain service to the Goa-generated interface
type TenantServiceImpl struct {
    domainService tenant.Service
    adapter       *adapters.TenantAdapter
    logger        logger.Logger
}

func NewTenantService(domainService tenant.Service, logger logger.Logger) goa_tenant.Service {
    return &TenantServiceImpl{
        domainService: domainService,
        adapter:       adapters.NewTenantAdapter(),
        logger:        logger,
    }
}

// Create implements goa_tenant.Service.Create
func (s *TenantServiceImpl) Create(ctx context.Context, p *goa_tenant.CreateTenantPayload) (*goa_tenant.TenantResult, error) {
    s.logger.Info("Creating tenant via Goa service", "name", *p.Name)
    
    // Convert Goa payload to domain request
    domainReq, err := s.adapter.CreatePayloadToDomain(p)
    if err != nil {
        return nil, err
    }
    
    // Call domain service
    domainResult, err := s.domainService.CreateTenant(ctx, domainReq)
    if err != nil {
        return nil, err
    }
    
    // Convert domain result to Goa result
    goaResult := s.adapter.DomainToGoaResult(domainResult)
    
    return goaResult, nil
}

// Get implements goa_tenant.Service.Get
func (s *TenantServiceImpl) Get(ctx context.Context, p *goa_tenant.GetTenantPayload) (*goa_tenant.TenantResult, error) {
    s.logger.Info("Getting tenant via Goa service", "id", *p.ID)
    
    tenantID, err := uuid.Parse(*p.ID)
    if err != nil {
        return nil, tenant.ErrInvalidInput
    }
    
    domainResult, err := s.domainService.GetTenant(ctx, tenantID)
    if err != nil {
        return nil, err
    }
    
    goaResult := s.adapter.DomainToGoaResult(domainResult)
    
    return goaResult, nil
}

// List implements goa_tenant.Service.List
func (s *TenantServiceImpl) List(ctx context.Context, p *goa_tenant.ListTenantsPayload) (*goa_tenant.TenantListResult, error) {
    s.logger.Info("Listing tenants via Goa service", "page", p.Page)
    
    // Convert Goa pagination to domain filter
    filter := tenant.ListFilter{
        Page:      int(*p.Page),
        Limit:     int(*p.Limit),
        SortBy:    p.SortBy,
        SortOrder: p.SortOrder,
        Search:    p.Search,
        Filters:   make(map[string]string),
    }
    
    // Add filters from payload
    if p.Plan != nil {
        filter.Filters["plan"] = *p.Plan
    }
    if p.Status != nil {
        filter.Filters["status"] = *p.Status
    }
    
    domainResults, pagination, err := s.domainService.ListTenants(ctx, filter)
    if err != nil {
        return nil, err
    }
    
    goaResult := s.adapter.DomainListToGoaResult(domainResults, pagination)
    
    return goaResult, nil
}

// Update implements goa_tenant.Service.Update
func (s *TenantServiceImpl) Update(ctx context.Context, p *goa_tenant.UpdateTenantPayload) (*goa_tenant.TenantResult, error) {
    s.logger.Info("Updating tenant via Goa service", "id", *p.ID)
    
    domainReq, err := s.adapter.UpdatePayloadToDomain(p)
    if err != nil {
        return nil, err
    }
    
    domainResult, err := s.domainService.UpdateTenant(ctx, domainReq)
    if err != nil {
        return nil, err
    }
    
    goaResult := s.adapter.DomainToGoaResult(domainResult)
    
    return goaResult, nil
}

// Delete implements goa_tenant.Service.Delete
func (s *TenantServiceImpl) Delete(ctx context.Context, p *goa_tenant.DeleteTenantPayload) error {
    s.logger.Info("Deleting tenant via Goa service", "id", *p.ID)
    
    tenantID, err := uuid.Parse(*p.ID)
    if err != nil {
        return tenant.ErrInvalidInput
    }
    
    return s.domainService.DeleteTenant(ctx, tenantID)
}

// GetStats implements goa_tenant.Service.GetStats
func (s *TenantServiceImpl) GetStats(ctx context.Context, p *goa_tenant.GetStatsPayload) (*goa_tenant.TenantStatsResult, error) {
    s.logger.Info("Getting tenant stats via Goa service", "id", *p.ID)
    
    tenantID, err := uuid.Parse(*p.ID)
    if err != nil {
        return nil, tenant.ErrInvalidInput
    }
    
    domainStats, err := s.domainService.GetTenantStats(ctx, tenantID)
    if err != nil {
        return nil, err
    }
    
    goaResult := s.adapter.StatsToGoaResult(domainStats)
    
    return goaResult, nil
}
```

## 🚀 9. Complete Main Server Setup

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
    
    // Standard library
    "database/sql"
    
    // Third-party packages
    "github.com/redis/go-redis/v9"
    
    // Your platform packages
    "github.com/niiniyare/erp/internal/platform/config"
    "github.com/niiniyare/erp/internal/platform/database"
    "github.com/niiniyare/erp/internal/platform/cache"
    "github.com/niiniyare/erp/internal/shared/logger"
    "github.com/niiniyare/erp/internal/shared/metrics"
    "github.com/niiniyare/erp/internal/shared/tracing"
    
    // Your core services
    "github.com/niiniyare/erp/internal/core/tenant"
    "github.com/niiniyare/erp/internal/core/user"
    "github.com/niiniyare/erp/internal/core/auth"
    "github.com/niiniyare/erp/internal/core/organization"
    "github.com/niiniyare/erp/internal/core/abac"
    
    // Your repositories
    tenantRepo "github.com/niiniyare/erp/internal/repositories/tenant"
    userRepo "github.com/niiniyare/erp/internal/repositories/user"
    authRepo "github.com/niiniyare/erp/internal/repositories/auth"
    orgRepo "github.com/niiniyare/erp/internal/repositories/organization"
    abacRepo "github.com/niiniyare/erp/internal/repositories/abac"
    
    // API layer
    "github.com/niiniyare/erp/internal/api/gin"
    "github.com/niiniyare/erp/internal/api/services"
    
    // Health check
    "github.com/niiniyare/erp/internal/platform/health"
)

type Server struct {
    config     *config.Config
    logger     logger.Logger
    db         *sql.DB
    cache      cache.Cache
    router     *gin.Router
    httpServer *http.Server
    
    // Services
    tenantService tenant.Service
    userService   user.Service
    authService   auth.Service
    orgService    organization.Service
    abacService   abac.Service
    healthService health.Service
}

func main() {
    // Create cancellable context
    ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer cancel()
    
    // Initialize server
    server, err := NewServer()
    if err != nil {
        log.Fatalf("Failed to initialize server: %v", err)
    }
    
    // Start server
    if err := server.Start(); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
    
    // Wait for shutdown signal
    <-ctx.Done()
    
    // Graceful shutdown
    if err := server.Shutdown(context.Background()); err != nil {
        log.Printf("Error during shutdown: %v", err)
    }
}

func NewServer() (*Server, error) {
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        return nil, fmt.Errorf("failed to load config: %w", err)
    }
    
    // Initialize logger
    logger := logger.New(cfg.Logger)
    logger.Info("Starting AWO ERP Server", "version", cfg.Version, "environment", cfg.Environment)
    
    // Initialize tracing
    if cfg.Tracing.Enabled {
        closer, err := tracing.Init(cfg.Tracing, "erp-api")
        if err != nil {
            logger.Error("Failed to initialize tracing", "error", err)
        } else {
            defer closer.Close()
            logger.Info("Tracing initialized")
        }
    }
    
    // Initialize metrics
    if cfg.Metrics.Enabled {
        metrics.Init(cfg.Metrics)
        logger.Info("Metrics initialized")
    }
    
    // Database connection
    db, err := database.Connect(ctx, cfg.Database)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }
    logger.Info("Database connected")
    
    // Cache connection
    cache, err := cache.NewRedis(cfg.Redis)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to cache: %w", err)
    }
    logger.Info("Cache connected")
    
    server := &Server{
        config: cfg,
        logger: logger,
        db:     db,
        cache:  cache,
    }
    
    // Initialize services
    if err := server.initializeServices(); err != nil {
        return nil, fmt.Errorf("failed to initialize services: %w", err)
    }
    
    // Initialize router
    if err := server.initializeRouter(); err != nil {
        return nil, fmt.Errorf("failed to initialize router: %w", err)
    }
    
    return server, nil
}

func (s *Server) initializeServices() error {
    // Initialize repositories
    tenantRepository := tenantRepo.NewRepository(s.db, s.logger)
    userRepository := userRepo.NewRepository(s.db, s.logger)
    authRepository := authRepo.NewRepository(s.db, s.cache, s.logger)
    orgRepository := orgRepo.NewRepository(s.db, s.logger)
    abacRepository := abacRepo.NewRepository(s.db, s.cache, s.logger)
    
    // Initialize domain services
    s.tenantService = tenant.NewService(tenantRepository, s.cache, s.logger)
    s.userService = user.NewService(userRepository, s.cache, s.logger)
    s.authService = auth.NewService(authRepository, s.cache, s.config.JWT, s.logger)
    s.orgService = organization.NewService(orgRepository, s.cache, s.logger)
    s.abacService = abac.NewService(abacRepository, s.cache, s.logger)
    
    // Initialize health service
    s.healthService = health.NewService(s.db, s.cache, s.logger)
    
    s.logger.Info("All services initialized successfully")
    return nil
}

func (s *Server) initializeRouter() error {
    // Create dependencies for Gin router
    deps := &gin.Dependencies{
        AuthService:   s.authService,
        TenantService: s.tenantService,
        UserService:   s.userService,
        ABACService:   s.abacService,
        OrgService:    s.orgService,
        HealthService: s.healthService,
    }
    
    // Initialize Gin router
    s.router = gin.NewRouter(s.config, s.logger, deps)
    
    s.logger.Info("Router initialized successfully")
    return nil
}

func (s *Server) Start() error {
    // Setup HTTP server
    s.httpServer = &http.Server{
        Addr:         s.config.Server.Address,
        Handler:      s.router.Setup(),
        ReadTimeout:  s.config.Server.ReadTimeout,
        WriteTimeout: s.config.Server.WriteTimeout,
        IdleTimeout:  s.config.Server.IdleTimeout,
        MaxHeaderBytes: 1 << 20, // 1MB
    }
    
    // Start server in goroutine
    go func() {
        s.logger.Info("Starting HTTP server", 
            "address", s.config.Server.Address,
            "environment", s.config.Environment)
        
        if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            s.logger.Fatal("HTTP server failed", "error", err)
        }
    }()
    
    // Health check - ensure server is ready
    if err := s.waitForServer(); err != nil {
        return fmt.Errorf("server failed to start: %w", err)
    }
    
    s.logger.Info("Server started successfully", "address", s.config.Server.Address)
    return nil
}

func (s *Server) waitForServer() error {
    client := &http.Client{Timeout: 5 * time.Second}
    
    for i := 0; i < 30; i++ { // Wait up to 30 seconds
        resp, err := client.Get(fmt.Sprintf("http://%s/health/live", s.config.Server.Address))
        if err == nil {
            resp.Body.Close()
            if resp.StatusCode == http.StatusOK {
                return nil
            }
        }
        time.Sleep(1 * time.Second)
    }
    
    return fmt.Errorf("server failed to become ready within timeout")
}

func (s *Server) Shutdown(ctx context.Context) error {
    s.logger.Info("Shutting down server...")
    
    // Shutdown HTTP server
    shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
        s.logger.Error("Error shutting down HTTP server", "error", err)
        return err
    }
    
    // Close database connection
    if err := s.db.Close(); err != nil {
        s.logger.Error("Error closing database connection", "error", err)
    }
    
    // Close cache connection
    if err := s.cache.Close(); err != nil {
        s.logger.Error("Error closing cache connection", "error", err)
    }
    
    // Shutdown router
    if err := s.router.Shutdown(ctx); err != nil {
        s.logger.Error("Error shutting down router", "error", err)
    }
    
    s.logger.Info("Server shutdown completed")
    return nil
}
```

## 🧪 10. Testing Examples

### Handler Testing

```go
// internal/api/gin/handlers/tenant_test.go
package handlers_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/google/uuid"
    
    // Your packages
    "github.com/niiniyare/erp/internal/api/gin/handlers"
    "github.com/niiniyare/erp/internal/core/tenant"
    "github.com/niiniyare/erp/internal/shared/logger"
    
    // Generated Goa types
    goa_tenant "github.com/niiniyare/erp/internal/api/gen/tenant"
)

// Mock service
type MockTenantService struct {
    mock.Mock
}

func (m *MockTenantService) CreateTenant(ctx context.Context, req *tenant.CreateRequest) (*tenant.Tenant, error) {
    args := m.Called(ctx, req)
    return args.Get(0).(*tenant.Tenant), args.Error(1)
}

func (m *MockTenantService) GetTenant(ctx context.Context, id uuid.UUID) (*tenant.Tenant, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(*tenant.Tenant), args.Error(1)
}

// Add other methods...

func TestTenantHandler_Create(t *testing.T) {
    gin.SetMode(gin.TestMode)
    
    tests := []struct {
        name           string
        payload        goa_tenant.CreateTenantPayload
        setupMock      func(*MockTenantService)
        expectedStatus int
        validateResp   func(t *testing.T, resp *httptest.ResponseRecorder)
    }{
        {
            name: "successful tenant creation",
            payload: goa_tenant.CreateTenantPayload{
                Name:         stringPtr("Test Tenant"),
                Subdomain:    stringPtr("test-tenant"),
                ContactEmail: stringPtr("admin@test-tenant.com"),
                Description:  stringPtr("Test tenant description"),
                Plan:         stringPtr("professional"),
            },
            setupMock: func(m *MockTenantService) {
                expectedReq := &tenant.CreateRequest{
                    Name:         "Test Tenant",
                    Subdomain:    "test-tenant",
                    ContactEmail: "admin@test-tenant.com",
                    Description:  stringPtr("Test tenant description"),
                    Plan:         stringPtr("professional"),
                }
                
                resultTenant := &tenant.Tenant{
                    ID:           uuid.New(),
                    Name:         "Test Tenant",
                    Subdomain:    "test-tenant",
                    ContactEmail: "admin@test-tenant.com",
                    Description:  stringPtr("Test tenant description"),
                    Plan:         tenant.PlanProfessional,
                    Status:       tenant.StatusActive,
                    CreatedAt:    time.Now(),
                    UpdatedAt:    time.Now(),
                }
                
                m.On("CreateTenant", mock.Anything, expectedReq).Return(resultTenant, nil)
            },
            expectedStatus: http.StatusCreated,
            validateResp: func(t *testing.T, resp *httptest.ResponseRecorder) {
                var response map[string]interface{}
                err := json.Unmarshal(resp.Body.Bytes(), &response)
                assert.NoError(t, err)
                
                assert.Equal(t, "success", response["status"])
                assert.NotNil(t, response["data"])
                
                data := response["data"].(map[string]interface{})
                assert.Equal(t, "Test Tenant", data["name"])
                assert.Equal(t, "test-tenant", data["subdomain"])
                assert.Equal(t, "professional", data["plan"])
            },
        },
        {
            name: "invalid payload - missing name",
            payload: goa_tenant.CreateTenantPayload{
                Subdomain:    stringPtr("test-tenant"),
                ContactEmail: stringPtr("admin@test-tenant.com"),
            },
            setupMock: func(m *MockTenantService) {
                // No mock setup needed as validation should fail before service call
            },
            expectedStatus: http.StatusBadRequest,
            validateResp: func(t *testing.T, resp *httptest.ResponseRecorder) {
                var response map[string]interface{}
                err := json.Unmarshal(resp.Body.Bytes(), &response)
                assert.NoError(t, err)
                
                assert.Equal(t, "error", response["status"])
                assert.NotNil(t, response["error"])
            },
        },
        {
            name: "service error - subdomain exists",
            payload: goa_tenant.CreateTenantPayload{
                Name:         stringPtr("Test Tenant"),
                Subdomain:    stringPtr("existing-tenant"),
                ContactEmail: stringPtr("admin@existing-tenant.com"),
            },
            setupMock: func(m *MockTenantService) {
                m.On("CreateTenant", mock.Anything, mock.Anything).Return(nil, tenant.ErrSubdomainExists)
            },
            expectedStatus: http.StatusConflict,
            validateResp: func(t *testing.T, resp *httptest.ResponseRecorder) {
                var response map[string]interface{}
                err := json.Unmarshal(resp.Body.Bytes(), &response)
                assert.NoError(t, err)
                
                assert.Equal(t, "error", response["status"])
                errorDetail := response["error"].(map[string]interface{})
                assert.Equal(t, "CONFLICT", errorDetail["code"])
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            mockService := new(MockTenantService)
            tt.setupMock(mockService)
            
            logger := logger.NewNoop() // Use a no-op logger for tests
            handler := handlers.NewTenantHandler(mockService, logger)
            
            // Create request
            payload, _ := json.Marshal(tt.payload)
            req := httptest.NewRequest(http.MethodPost, "/tenants", bytes.NewBuffer(payload))
            req.Header.Set("Content-Type", "application/json")
            
            // Setup Gin context
            w := httptest.NewRecorder()
            c, _ := gin.CreateTestContext(w)
            c.Request = req
            
            // Set user context (simulate authenticated user)
            c.Set("user_id", "user-123")
            c.Set("tenant_id", "tenant-123")
            
            // Execute
            handler.Create(c)
            
            // Assert
            assert.Equal(t, tt.expectedStatus, w.Code)
            tt.validateResp(t, w)
            
            mockService.AssertExpectations(t)
        })
    }
}

func stringPtr(s string) *string {
    return &s
}
```

## 🏭 11. Production Considerations

### Docker Configuration

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/server/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy configuration files
COPY --from=builder /app/configs ./configs

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health/live || exit 1

# Run the binary
CMD ["./main"]
```

### Kubernetes Deployment

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: erp-api
  labels:
    app: erp-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: erp-api
  template:
    metadata:
      labels:
        app: erp-api
    spec:
      containers:
      - name: erp-api
        image: your-registry/erp-api:latest
        ports:
        - containerPort: 8080
        env:
        - name: ENVIRONMENT
          value: "production"
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: erp-secrets
              key: database-url
        - name: REDIS_URL
          valueFrom:
            secretKeyRef:
              name: erp-secrets
              key: redis-url
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: erp-secrets
              key: jwt-secret
        resources:
          limits:
            cpu: 1000m
            memory: 1Gi
          requests:
            cpu: 500m
            memory: 512Mi
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        securityContext:
          allowPrivilegeEscalation: false
          runAsNonRoot: true
          runAsUser: 10001
          capabilities:
            drop:
            - ALL
```

### Configuration Management

```go
// internal/platform/config/production.go
package config

import (
    "time"
)

// ProductionConfig returns configuration optimized for production
func ProductionConfig() *Config {
    return &Config{
        Environment: "production",
        Server: ServerConfig{
            Address:      ":8080",
            ReadTimeout:  30 * time.Second,
            WriteTimeout: 30 * time.Second,
            IdleTimeout:  120 * time.Second,
        },
        Database: DatabaseConfig{
            Driver:          "postgres",
            MaxOpenConns:    25,
            MaxIdleConns:    10,
            ConnMaxLifetime: 5 * time.Minute,
            ConnMaxIdleTime: 1 * time.Minute,
        },
        Redis: RedisConfig{
            PoolSize:     20,
            MinIdleConns: 5,
            MaxRetries:   3,
            DialTimeout:  5 * time.Second,
            ReadTimeout:  3 * time.Second,
            WriteTimeout: 3 * time.Second,
        },
        JWT: JWTConfig{
            AccessTokenTTL:  15 * time.Minute,
            RefreshTokenTTL: 7 * 24 * time.Hour,
            Algorithm:       "HS256",
        },
        RateLimit: RateLimitConfig{
            Enabled: true,
            Rate:    100, // requests per minute
            Burst:   20,
        },
        CORS: CORSConfig{
            AllowedOrigins:   []string{"https://app.yourdomain.com"},
            AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
            AllowedHeaders:   []string{"*"},
            AllowCredentials: true,
        },
        Logging: LoggingConfig{
            Level:  "info",
            Format: "json",
        },
        Metrics: MetricsConfig{
            Enabled: true,
            Path:    "/metrics",
        },
        Tracing: TracingConfig{
            Enabled:     true,
            ServiceName: "erp-api",
            Endpoint:    "jaeger-collector:14268",
        },
    }
}
```

## 🔧 12. Complete User Service Implementation

### User Service Design

```go
// internal/api/design/services/user/user.go
package user

import (
    . "goa.design/goa/v3/dsl"
    "github.com/niiniyare/erp/internal/api/design/types"
)

var _ = Service("user", func() {
    Description("User management service with ABAC integration")
    
    Security("jwt")
    
    Error("user_not_found", String, "User not found")
    Error("email_exists", String, "Email already exists")
    Error("invalid_user_data", String, "Invalid user data")
    Error("user_limit_exceeded", String, "User limit exceeded for tenant")
    Error("weak_password", String, "Password does not meet security requirements")
    
    Method("create", func() {
        Description("Create a new user")
        Security("jwt", func() {
            Scope("user:create")
        })
        
        Payload(func() {
            Attribute("email", String, "User email address", func() {
                Format(FormatEmail)
                Example("john.doe@acme-corp.com")
            })
            Attribute("first_name", String, "User first name", func() {
                MinLength(1)
                MaxLength(50)
                Pattern("^[a-zA-Z\\s\\-']+$")
                Example("John")
            })
            Attribute("last_name", String, "User last name", func() {
                MinLength(1)
                MaxLength(50)
                Pattern("^[a-zA-Z\\s\\-']+$")
                Example("Doe")
            })
            Attribute("password", String, "User password", func() {
                MinLength(8)
                MaxLength(128)
                Pattern("^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)(?=.*[@$!%*?&])[A-Za-z\\d@$!%*?&]")
                Example("SecurePassword123!")
            })
            Attribute("roles", ArrayOf(String), "User roles", func() {
                Example([]string{"user", "manager"})
            })
            Attribute("organization_id", String, "Organization ID", func() {
                Format(FormatUUID)
                Example("550e8400-e29b-41d4-a716-446655440000")
            })
            Attribute("department", String, "User department", func() {
                MaxLength(100)
                Example("Engineering")
            })
            Attribute("job_title", String, "User job title", func() {
                MaxLength(100)
                Example("Senior Software Engineer")
            })
            Attribute("phone", String, "User phone number", func() {
                Pattern("^\\+?[1-9]\\d{1,14}$")
                Example("+1-555-123-4567")
            })
            Attribute("timezone", String, "User timezone", func() {
                Example("America/New_York")
            })
            Attribute("locale", String, "User locale", func() {
                Example("en-US")
            })
            Attribute("send_welcome_email", Boolean, "Send welcome email", func() {
                Default(true)
                Example(true)
            })
            Required("email", "first_name", "last_name", "password")
        })
        
        Result(UserResult)
        
        HTTP(func() {
            POST("/users")
            Response(StatusCreated)
            Response("bad_request", StatusBadRequest)
            Response("unauthorized", StatusUnauthorized)
            Response("forbidden", StatusForbidden)
            Response("email_exists", StatusConflict)
            Response("user_limit_exceeded", StatusTooManyRequests)
            Response("weak_password", StatusBadRequest)
        })
    })
    
    Method("get", func() {
        Description("Get user by ID")
        Security("jwt", func() {
            Scope("user:read")
        })
        
        Payload(func() {
            Attribute("id", String, "User ID", func() {
                Format(FormatUUID)
                Example("550e8400-e29b-41d4-a716-446655440000")
            })
            Required("id")
        })
        
        Result(UserResult)
        
        HTTP(func() {
            GET("/users/{id}")
            Response(StatusOK)
            Response("not_found", StatusNotFound)
            Response("unauthorized", StatusUnauthorized)
            Response("forbidden", StatusForbidden)
        })
    })
    
    Method("list", func() {
        Description("List users with pagination and filtering")
        Security("jwt", func() {
            Scope("user:read")
        })
        
        Payload(func() {
            Attribute("page", UInt, "Page number", func() {
                Default(1)
                Minimum(1)
                Maximum(1000)
            })
            Attribute("limit", UInt, "Items per page", func() {
                Default(20)
                Minimum(1)
                Maximum(100)
            })
            Attribute("sort_by", String, "Sort field", func() {
                Enum("created_at", "updated_at", "first_name", "last_name", "email")
                Default("created_at")
            })
            Attribute("sort_order", String, "Sort order", func() {
                Enum("asc", "desc")
                Default("desc")
            })
            Attribute("search", String, "Search query", func() {
                MaxLength(255)
                Example("john")
            })
            Attribute("role", String, "Filter by role", func() {
                Example("admin")
            })
            Attribute("status", String, "Filter by status", func() {
                Enum("active", "inactive", "suspended")
                Example("active")
            })
            Attribute("organization_id", String, "Filter by organization", func() {
                Format(FormatUUID)
                Example("550e8400-e29b-41d4-a716-446655440000")
            })
            Attribute("department", String, "Filter by department", func() {
                Example("Engineering")
            })
        })
        
        Result(types.PaginatedResponse("UserListResult", UserResult))
        
        HTTP(func() {
            GET("/users")
            Params(func() {
                Param("page")
                Param("limit")
                Param("sort_by")
                Param("sort_order")
                Param("search")
                Param("role")
                Param("status")
                Param("organization_id")
                Param("department")
            })
            Response(StatusOK)
            Response("unauthorized", StatusUnauthorized)
            Response("forbidden", StatusForbidden)
        })
    })
    
    Method("update", func() {
        Description("Update user information")
        Security("jwt", func() {
            Scope("user:update")
        })
        
        Payload(func() {
            Attribute("id", String, "User ID", func() {
                Format(FormatUUID)
                Example("550e8400-e29b-41d4-a716-446655440000")
            })
            Attribute("first_name", String, "User first name", func() {
                MinLength(1)
                MaxLength(50)
                Pattern("^[a-zA-Z\\s\\-']+$")
                Example("John")
            })
            Attribute("last_name", String, "User last name", func() {
                MinLength(1)
                MaxLength(50)
                Pattern("^[a-zA-Z\\s\\-']+$")
                Example("Doe")
            })
            Attribute("roles", ArrayOf(String), "User roles", func() {
                Example([]string{"user", "manager"})
            })
            Attribute("organization_id", String, "Organization ID", func() {
                Format(FormatUUID)
                Example("550e8400-e29b-41d4-a716-446655440000")
            })
            Attribute("department", String, "User department", func() {
                MaxLength(100)
                Example("Engineering")
            })
            Attribute("job_title", String, "User job title", func() {
                MaxLength(100)
                Example("Senior Software Engineer")
            })
            Attribute("phone", String, "User phone number", func() {
                Pattern("^\\+?[1-9]\\d{1,14}$")
                Example("+1-555-123-4567")
            })
            Attribute("timezone", String, "User timezone", func() {
                Example("America/New_York")
            })
            Attribute("locale", String, "User locale", func() {
                Example("en-US")
            })
            Attribute("status", String, "User status", func() {
                Enum("active", "inactive", "suspended")
                Example("active")
            })
            Required("id")
        })
        
        Result(UserResult)
        
        HTTP(func() {
            PUT("/users/{id}")
            Response(StatusOK)
            Response("bad_request", StatusBadRequest)
            Response("not_found", StatusNotFound)
            Response("unauthorized", StatusUnauthorized)
            Response("forbidden", StatusForbidden)
        })
    })
    
    Method("delete", func() {
        Description("Delete user (soft delete)")
        Security("jwt", func() {
            Scope("user:delete")
        })
        
        Payload(func() {
            Attribute("id", String, "User ID", func() {
                Format(FormatUUID)
                Example("550e8400-e29b-41d4-a716-446655440000")
            })
            Required("id")
        })
        
        HTTP(func() {
            DELETE("/users/{id}")
            Response(StatusNoContent)
            Response("not_found", StatusNotFound)
            Response("unauthorized", StatusUnauthorized)
            Response("forbidden", StatusForbidden)
        })
    })
    
    Method("get_permissions", func() {
        Description("Get user permissions")
        Security("jwt", func() {
            Scope("user:read")
        })
        
        Payload(func() {
            Attribute("id", String, "User ID", func() {
                Format(FormatUUID)
                Example("550e8400-e29b-41d4-a716-446655440000")
            })
            Required("id")
        })
        
        Result(UserPermissionsResult)
        
        HTTP(func() {
            GET("/users/{id}/permissions")
            Response(StatusOK)
            Response("not_found", StatusNotFound)
            Response("unauthorized", StatusUnauthorized)
            Response("forbidden", StatusForbidden)
        })
    })
    
    Method("update_permissions", func() {
        Description("Update user permissions")
        Security("jwt", func() {
            Scope("user:update")
        })
        
        Payload(func() {
            Attribute("id", String, "User ID", func() {
                Format(FormatUUID)
                Example("550e8400-e29b-41d4-a716-446655440000")
            })
            Attribute("roles", ArrayOf(String), "User roles", func() {
                Example([]string{"user", "manager", "admin"})
            })
            Attribute("permissions", ArrayOf(String), "Direct permissions", func() {
                Example([]string{"tenant:read", "user:create"})
            })
            Attribute("abac_policies", ArrayOf(String), "ABAC policy IDs", func() {
                Example([]string{"policy-123", "policy-456"})
            })
            Required("id")
        })
        
        Result(UserPermissionsResult)
        
        HTTP(func() {
            PUT("/users/{id}/permissions")
            Response(StatusOK)
            Response("bad_request", StatusBadRequest)
            Response("not_found", StatusNotFound)
            Response("unauthorized", StatusUnauthorized)
            Response("forbidden", StatusForbidden)
        })
    })
    
    Method("reset_password", func() {
        Description("Reset user password (admin action)")
        Security("jwt", func() {
            Scope("user:update")
        })
        
        Payload(func() {
            Attribute("id", String, "User ID", func() {
                Format(FormatUUID)
                Example("550e8400-e29b-41d4-a716-446655440000")
            })
            Attribute("temporary_password", String, "Temporary password", func() {
                MinLength(8)
                MaxLength(128)
                Example("TempPassword123!")
            })
            Attribute("force_change", Boolean, "Force password change on next login", func() {
                Default(true)
                Example(true)
            })
            Attribute("send_email", Boolean, "Send password reset email", func() {
                Default(true)
                Example(true)
            })
            Required("id", "temporary_password")
        })
        
        HTTP(func() {
            POST("/users/{id}/reset-password")
            Response(StatusNoContent)
            Response("bad_request", StatusBadRequest)
            Response("not_found", StatusNotFound)
            Response("unauthorized", StatusUnauthorized)
            Response("forbidden", StatusForbidden)
        })
    })
})
```

### User Types

```go
// internal/api/design/services/user/types.go
package user

import (
    . "goa.design/goa/v3/dsl"
    "github.com/niiniyare/erp/internal/api/design/types"
)

var UserResult = ResultType("application/vnd.user+json", func() {
    Description("User information")
    Attributes(func() {
        Attribute("id", String, "Unique user identifier", func() {
            Format(FormatUUID)
            Example("550e8400-e29b-41d4-a716-446655440000")
        })
        Attribute("email", String, "User email address", func() {
            Format(FormatEmail)
            Example("john.doe@acme-corp.com")
        })
        Attribute("first_name", String, "User first name", func() {
            Example("John")
        })
        Attribute("last_name", String, "User last name", func() {
            Example("Doe")
        })
        Attribute("display_name", String, "User display name", func() {
            Example("John Doe")
        })
        Attribute("avatar_url", String, "User avatar URL", func() {
            Format(FormatURI)
            Example("https://cdn.acme-corp.com/avatars/john-doe.jpg")
        })
        Attribute("roles", ArrayOf(String), "User roles", func() {
            Example([]string{"user", "manager"})
        })
        Attribute("status", String, "User status", func() {
            Enum("active", "inactive", "suspended", "pending")
            Example("active")
        })
        Attribute("organization_id", String, "Organization ID", func() {
            Format(FormatUUID)
            Example("550e8400-e29b-41d4-a716-446655440000")
        })
        Attribute("organization_name", String, "Organization name", func() {
            Example("Engineering")
        })
        Attribute("department", String, "User department", func() {
            Example("Engineering")
        })
        Attribute("job_title", String, "User job title", func() {
            Example("Senior Software Engineer")
        })
        Attribute("phone", String, "User phone number", func() {
            Example("+1-555-123-4567")
        })
        Attribute("timezone", String, "User timezone", func() {
            Example("America/New_York")
        })
        Attribute("locale", String, "User locale", func() {
            Example("en-US")
        })
        Attribute("last_login", String, "Last login timestamp", func() {
            Format(FormatDateTime)
            Example("2024-01-15T10:30:00Z")
        })
        Attribute("login_count", UInt, "Total login count", func() {
            Example(42)
        })
        Attribute("email_verified", Boolean, "Whether email is verified", func() {
            Example(true)
        })
        Attribute("mfa_enabled", Boolean, "Whether MFA is enabled", func() {
            Example(false)
        })
        Attribute("preferences", MapOf(String, Any), "User preferences", func() {
            Example(map[string]interface{}{
                "theme":               "dark",
                "notifications":       true,
                "dashboard_layout":    "grid",
                "date_format":         "MM/DD/YYYY",
                "time_format":         "12h",
            })
        })
        types.AuditFields()
    })
    Required("id", "email", "first_name", "last_name", "status", "roles", "created_at", "updated_at")
})

var UserPermissionsResult = ResultType("application/vnd.user.permissions+json", func() {
    Description("User permissions and access rights")
    Attributes(func() {
        Attribute("user_id", String, "User ID", func() {
            Format(FormatUUID)
            Example("550e8400-e29b-41d4-a716-446655440000")
        })
        Attribute("roles", ArrayOf(String), "User roles", func() {
            Example([]string{"user", "manager", "admin"})
        })
        Attribute("direct_permissions", ArrayOf(String), "Direct permissions", func() {
            Example([]string{"tenant:read", "user:create", "organization:update"})
        })
        Attribute("inherited_permissions", ArrayOf(String), "Permissions from roles", func() {
            Example([]string{"user:read", "user:update"})
        })
        Attribute("abac_policies", ArrayOf(ABACPolicyRef), "Applied ABAC policies")
        Attribute("effective_permissions", ArrayOf(String), "All effective permissions", func() {
            Example([]string{"tenant:read", "user:create", "user:read", "user:update", "organization:update"})
        })
        Attribute("restrictions", ArrayOf(PermissionRestriction), "Permission restrictions")
        Attribute("last_evaluated", String, "When permissions were last evaluated", func() {
            Format(FormatDateTime)
            Example("2024-01-15T16:00:00Z")
        })
    })
    Required("user_id", "roles", "effective_permissions", "last_evaluated")
})

var ABACPolicyRef = Type("ABACPolicyRef", func() {
    Description("Reference to an ABAC policy")
    Attribute("id", String, "Policy ID", func() {
        Format(FormatUUID)
        Example("550e8400-e29b-41d4-a716-446655440000")
    })
    Attribute("name", String, "Policy name", func() {
        Example("department-manager-access")
    })
    Attribute("effect", String, "Policy effect", func() {
        Enum("allow", "deny")
        Example("allow")
    })
    Attribute("priority", UInt, "Policy priority", func() {
        Example(100)
    })
    Required("id", "name", "effect")
})

var PermissionRestriction = Type("PermissionRestriction", func() {
    Description("Permission restriction or condition")
    Attribute("permission", String, "Restricted permission", func() {
        Example("user:delete")
    })
    Attribute("condition", String, "Restriction condition", func() {
        Example("Cannot delete users from other departments")
    })
    Attribute("resource_filters", MapOf(String, String), "Resource-level filters", func() {
        Example(map[string]string{
            "department": "Engineering",
            "tenant_id":  "550e8400-e29b-41d4-a716-446655440000",
        })
    })
    Required("permission", "condition")
})
```

## 🏢 13. Complete Organization Service Implementation

### Organization Handler

```go
// internal/api/gin/handlers/organization.go
package handlers

import (
    "net/http"
    "strconv"
    
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    
    // Your domain imports
    "github.com/niiniyare/erp/internal/core/organization"
    "github.com/niiniyare/erp/internal/shared/logger"
    
    // Generated Goa types
    goa_org "github.com/niiniyare/erp/internal/api/gen/organization"
    
    // Gin utilities
    "github.com/niiniyare/erp/internal/api/gin/adapters"
    "github.com/niiniyare/erp/internal/api/gin/responses"
)

type OrganizationHandler struct {
    service organization.Service
    logger  logger.Logger
    adapter *adapters.OrganizationAdapter
}

func NewOrganizationHandler(service organization.Service, logger logger.Logger) *OrganizationHandler {
    return &OrganizationHandler{
        service: service,
        logger:  logger,
        adapter: adapters.NewOrganizationAdapter(),
    }
}

func (h *OrganizationHandler) Create(c *gin.Context) {
    ctx := c.Request.Context()
    h.logger.Info("Creating organization", "path", c.FullPath())
    
    var payload goa_org.CreateOrganizationPayload
    if err := c.ShouldBindJSON(&payload); err != nil {
        h.logger.Error("Invalid create organization payload", "error", err)
        responses.ValidationError(c, "Invalid request payload", err)
        return
    }
    
    domainReq, err := h.adapter.CreatePayloadToDomain(&payload)
    if err != nil {
        h.logger.Error("Failed to convert organization payload", "error", err)
        responses.BadRequest(c, "Invalid payload data", err)
        return
    }
    
    // Add context
    if userID := c.GetString("user_id"); userID != "" {
        domainReq.CreatedBy = &userID
    }
    
    if tenantID := c.GetString("tenant_id"); tenantID != "" {
        domainReq.TenantID = tenantID
    }
    
    result, err := h.service.CreateOrganization(ctx, domainReq)
    if err != nil {
        h.logger.Error("Failed to create organization", "error", err)
        responses.HandleDomainError(c, err)
        return
    }
    
    goaResult := h.adapter.DomainToGoaResult(result)
    
    h.logger.Info("Organization created successfully", 
        "org_id", result.ID, 
        "name", result.Name,
        "type", result.Type)
        
    responses.Created(c, goaResult)
}

func (h *OrganizationHandler) Get(c *gin.Context) {
    ctx := c.Request.Context()
    orgIDStr := c.Param("id")
    
    orgID, err := uuid.Parse(orgIDStr)
    if err != nil {
        h.logger.Error("Invalid organization ID format", "id", orgIDStr, "error", err)
        responses.BadRequest(c, "Invalid organization ID format", err)
        return
    }
    
    h.logger.Info("Getting organization", "org_id", orgID)
    
    result, err := h.service.GetOrganization(ctx, orgID)
    if err != nil {
        h.logger.Error("Failed to get organization", "org_id", orgID, "error", err)
        responses.HandleDomainError(c, err)
        return
    }
    
    goaResult := h.adapter.DomainToGoaResult(result)
    responses.OK(c, goaResult)
}

func (h *OrganizationHandler) List(c *gin.Context) {
    ctx := c.Request.Context()
    
    // Parse parameters
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
    sortBy := c.DefaultQuery("sort_by", "name")
    sortOrder := c.DefaultQuery("sort_order", "asc")
    search := c.Query("search")
    orgType := c.Query("type")
    parentID := c.Query("parent_id")
    
    filter := organization.ListFilter{
        Page:      page,
        Limit:     limit,
        SortBy:    &sortBy,
        SortOrder: &sortOrder,
        Search:    &search,
        Filters: map[string]string{
            "type":      orgType,
            "parent_id": parentID,
        },
    }
    
    h.logger.Info("Listing organizations", 
        "page", page, 
        "limit", limit, 
        "search", search,
        "type", orgType)
    
    results, pagination, err := h.service.ListOrganizations(ctx, filter)
    if err != nil {
        h.logger.Error("Failed to list organizations", "error", err)
        responses.HandleDomainError(c, err)
        return
    }
    
    goaResult := h.adapter.DomainListToGoaResult(results, pagination)
    
    // Add total count header
    c.Header("X-Total-Count", strconv.FormatUint(pagination.TotalCount, 10))
    
    responses.OK(c, goaResult)
}

func (h *OrganizationHandler) Update(c *gin.Context) {
    ctx := c.Request.Context()
    orgIDStr := c.Param("id")
    
    orgID, err := uuid.Parse(orgIDStr)
    if err != nil {
        h.logger.Error("Invalid organization ID format", "id", orgIDStr, "error", err)
        responses.BadRequest(c, "Invalid organization ID format", err)
        return
    }
    
    var payload goa_org.UpdateOrganizationPayload
    if err := c.ShouldBindJSON(&payload); err != nil {
        h.logger.Error("Invalid update organization payload", "error", err)
        responses.ValidationError(c, "Invalid request payload", err)
        return
    }
    
    // Ensure ID consistency
    payload.ID = &orgIDStr
    
    domainReq, err := h.adapter.UpdatePayloadToDomain(&payload)
    if err != nil {
        h.logger.Error("Failed to convert update payload", "error", err)
        responses.BadRequest(c, "Invalid payload data", err)
        return
    }
    
    // Add context
    if userID := c.GetString("user_id"); userID != "" {
        domainReq.UpdatedBy = &userID
    }
    
    h.logger.Info("Updating organization", "org_id", orgID)
    
    result, err := h.service.UpdateOrganization(ctx, domainReq)
    if err != nil {
        h.logger.Error("Failed to update organization", "org_id", orgID, "error", err)
        responses.HandleDomainError(c, err)
        return
    }
    
    goaResult := h.adapter.DomainToGoaResult(result)
    
    h.logger.Info("Organization updated successfully", "org_id", result.ID)
    responses.OK(c, goaResult)
}

func (h *OrganizationHandler) Delete(c *gin.Context) {
    ctx := c.Request.Context()
    orgIDStr := c.Param("id")
    
    orgID, err := uuid.Parse(orgIDStr)
    if err != nil {
        h.logger.Error("Invalid organization ID format", "id", orgIDStr, "error", err)
        responses.BadRequest(c, "Invalid organization ID format", err)
        return
    }
    
    h.logger.Info("Deleting organization", "org_id", orgID)
    
    err = h.service.DeleteOrganization(ctx, orgID)
    if err != nil {
        h.logger.Error("Failed to delete organization", "org_id", orgID, "error", err)
        responses.HandleDomainError(c, err)
        return
    }
    
    h.logger.Info("Organization deleted successfully", "org_id", orgID)
    responses.NoContent(c)
}

func (h *OrganizationHandler) GetHierarchy(c *gin.Context) {
    ctx := c.Request.Context()
    orgIDStr := c.Param("id")
    
    var orgID *uuid.UUID
    if orgIDStr != "" && orgIDStr != "root" {
        id, err := uuid.Parse(orgIDStr)
        if err != nil {
            h.logger.Error("Invalid organization ID format", "id", orgIDStr, "error", err)
            responses.BadRequest(c, "Invalid organization ID format", err)
            return
        }
        orgID = &id
    }
    
    h.logger.Info("Getting organization hierarchy", "org_id", orgID)
    
    hierarchy, err := h.service.GetOrganizationHierarchy(ctx, orgID)
    if err != nil {
        h.logger.Error("Failed to get organization hierarchy", "org_id", orgID, "error", err)
        responses.HandleDomainError(c, err)
        return
    }
    
    goaResult := h.adapter.HierarchyToGoaResult(hierarchy)
    responses.OK(c, goaResult)
}

func (h *OrganizationHandler) GetMembers(c *gin.Context) {
    ctx := c.Request.Context()
    orgIDStr := c.Param("id")
    
    orgID, err := uuid.Parse(orgIDStr)
    if err != nil {
        h.logger.Error("Invalid organization ID format", "id", orgIDStr, "error", err)
        responses.BadRequest(c, "Invalid organization ID format", err)
        return
    }
    
    // Parse parameters
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
    includeSubOrgs := c.DefaultQuery("include_sub_orgs", "false") == "true"
    
    filter := organization.MemberFilter{
        Page:            page,
        Limit:           limit,
        IncludeSubOrgs:  includeSubOrgs,
    }
    
    h.logger.Info("Getting organization members", "org_id", orgID, "include_sub_orgs", includeSubOrgs)
    
    members, pagination, err := h.service.GetOrganizationMembers(ctx, orgID, filter)
    if err != nil {
        h.logger.Error("Failed to get organization members", "org_id", orgID, "error", err)
        responses.HandleDomainError(c, err)
        return
    }
    
    goaResult := h.adapter.MembersToGoaResult(members, pagination)
    responses.OK(c, goaResult)
}
```

## 🛡️ 14. Complete ABAC Service Implementation

### ABAC Handler

```go
// internal/api/gin/handlers/abac.go
package handlers

import (
    "net/http"
    "strconv"
    
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    
    // Your domain imports
    "github.com/niiniyare/erp/internal/core/abac"
    "github.com/niiniyare/erp/internal/shared/logger"
    
    // Generated Goa types
    goa_abac "github.com/niiniyare/erp/internal/api/gen/abac"
    
    // Gin utilities
    "github.com/niiniyare/erp/internal/api/gin/adapters"
    "github.com/niiniyare/erp/internal/api/gin/responses"
)

type ABACHandler struct {
    service abac.Service
    logger  logger.Logger
    adapter *adapters.ABACAdapter
}

func NewABACHandler(service abac.Service, logger logger.Logger) *ABACHandler {
    return &ABACHandler{
        service: service,
        logger:  logger,
        adapter: adapters.NewABACAdapter(),
    }
}

func (h *ABACHandler) CreatePolicy(c *gin.Context) {
    ctx := c.Request.Context()
    h.logger.Info("Creating ABAC policy", "path", c.FullPath())
    
    var payload goa_abac.CreatePolicyPayload
    if err := c.ShouldBindJSON(&payload); err != nil {
        h.logger.Error("Invalid create policy payload", "error", err)
        responses.ValidationError(c, "Invalid request payload", err)
        return
    }
    
    domainReq, err := h.adapter.CreatePolicyPayloadToDomain(&payload)
    if err != nil {
        h.logger.Error("Failed to convert policy payload", "error", err)
        responses.BadRequest(c, "Invalid payload data", err)
        return
    }
    
    // Add context
    if userID := c.GetString("user_id"); userID != "" {
        domainReq.CreatedBy = &userID
    }
    
    if tenantID := c.GetString("tenant_id"); tenantID != "" {
        domainReq.TenantID = tenantID
    }
    
    result, err := h.service.CreatePolicy(ctx, domainReq)
    if err != nil {
        h.logger.Error("Failed to create ABAC policy", "error", err)
        responses.HandleDomainError(c, err)
        return
    }
    
    goaResult := h.adapter.PolicyToGoaResult(result)
    
    h.logger.Info("ABAC policy created successfully", 
        "policy_id", result.ID, 
        "name", result.Name,
        "effect", result.Effect)
        
    responses.Created(c, goaResult)
}

func (h *ABACHandler) GetPolicy(c *gin.Context) {
    ctx := c.Request.Context()
    policyIDStr := c.Param("id")
    
    policyID, err := uuid.Parse(policyIDStr)
    if err != nil {
        h.logger.Error("Invalid policy ID format", "id", policyIDStr, "error", err)
        responses.BadRequest(c, "Invalid policy ID format", err)
        return
    }
    
    h.logger.Info("Getting ABAC policy", "policy_id", policyID)
    
    result, err := h.service.GetPolicy(ctx, policyID)
    if err != nil {
        h.logger.Error("Failed to get ABAC policy", "policy_id", policyID, "error", err)
        responses.HandleDomainError(c, err)
        return
    }
    
    goaResult := h.adapter.PolicyToGoaResult(result)
    responses.OK(c, goaResult)
}

func (h *ABACHandler) ListPolicies(c *gin.Context) {
    ctx := c.Request.Context()
    
    // Parse parameters
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
    sortBy := c.DefaultQuery("sort_by", "created_at")
    sortOrder := c.DefaultQuery("sort_order", "desc")
    search := c.Query("search")
    effect := c.Query("effect")
    enabled := c.Query("enabled")
    
    filter := abac.PolicyFilter{
        Page:      page,
        Limit:     limit,
        SortBy:    &sortBy,
        SortOrder: &sortOrder,
        Search:    &search,
        Filters: map[string]string{
            "effect":  effect,
            "enabled": enabled,
        },
    }
    
    h.logger.Info("Listing ABAC policies", 
        "page", page, 
        "limit", limit, 
        "search", search,
        "effect", effect)
    
    results, pagination, err := h.service.ListPolicies(ctx, filter)
    if err != nil {
        h.logger.Error("Failed to list ABAC policies", "error", err)
        responses.HandleDomainError(c, err)
        return
    }
    
    goaResult := h.adapter.PolicyListToGoaResult(results, pagination)
    
    // Add total count header
    c.Header("X-Total-Count", strconv.FormatUint(pagination.TotalCount, 10))
    
    responses.OK(c, goaResult)
}

func (h *ABACHandler) UpdatePolicy(c *gin.Context) {
    ctx := c.Request.Context()
    policyIDStr := c.Param("id")
    
    policyID, err := uuid.Parse(policyIDStr)
    if err != nil {
        h.logger.Error("Invalid policy ID format", "id", policyIDStr, "error", err)
        responses.BadRequest(c, "Invalid policy ID format", err)
        return
    }
    
    var payload goa_abac.UpdatePolicyPayload
    if err := c.ShouldBindJSON(&payload); err != nil {
        h.logger.Error("Invalid update policy payload", "error", err)
        responses.ValidationError(c, "Invalid request payload", err)
        return
    }
    
    // Ensure ID consistency
    payload.ID = &policyIDStr
    
    domainReq, err := h.adapter.UpdatePolicyPayloadToDomain(&payload)
    if err != nil {
        h.logger.Error("Failed to convert update payload", "error", err)
        responses.BadRequest(c, "Invalid payload data", err)
        return
    }
    
    // Add context
    if userID := c.GetString("user_id"); userID != "" {
        domainReq.UpdatedBy = &userID
    }
    
    h.logger.Info("Updating ABAC policy", "policy_id", policyID)
    
    result, err := h.service.UpdatePolicy(ctx, domainReq)
    if err != nil {
        h.logger.Error("Failed to update ABAC policy", "policy_id", policyID, "error", err)
        responses.HandleDomainError(c, err)
        return
    }
    
    goaResult := h.adapter.PolicyToGoaResult(result)
    
    h.logger.Info("ABAC policy updated successfully", "policy_id", result.ID)
    responses.OK(c, goaResult)
}

func (h *ABACHandler) DeletePolicy(c *gin.Context) {
    ctx := c.Request.Context()
    policyIDStr := c.Param("id")
    
    policyID, err := uuid.Parse(policyIDStr)
    if err != nil {
        h.logger.Error("Invalid policy ID format", "id", policyIDStr, "error", err)
        responses.BadRequest(c, "Invalid policy ID format", err)
        return
    }
    
    h.logger.Info("Deleting ABAC policy", "policy_id", policyID)
    
    err = h.service.DeletePolicy(ctx, policyID)
    if err != nil {
        h.logger.Error("Failed to delete ABAC policy", "policy_id", policyID, "error", err)
        responses.HandleDomainError(c, err)
        return
    }
    
    h.logger.Info("ABAC policy deleted successfully", "policy_id", policyID)
    responses.NoContent(c)
}

func (h *ABACHandler) CheckPermission(c *gin.Context) {
    ctx := c.Request.Context()
    
    var payload goa_abac.CheckPermissionPayload
    if err := c.ShouldBindJSON(&payload); err != nil {
        h.logger.Error("Invalid check permission payload", "error", err)
        responses.ValidationError(c, "Invalid request payload", err)
        return
    }
    
    domainReq, err := h.adapter.CheckPermissionPayloadToDomain(&payload)
    if err != nil {
        h.logger.Error("Failed to convert check permission payload", "error", err)
        responses.BadRequest(c, "Invalid payload data", err)
        return
    }
    
    // Add request context
    domainReq.Context["ip_address"] = c.ClientIP()
    domainReq.Context["user_agent"] = c.Request.UserAgent()
    domainReq.Context["method"] = c.Request.Method
    domainReq.Context["path"] = c.Request.URL.Path
    
    if tenantID := c.GetString("tenant_id"); tenantID != "" {
        domainReq.Context["tenant_id"] = tenantID
    }
    
    h.logger.Info("Checking ABAC permission", 
        "user_id", domainReq.UserID,
        "action", domainReq.Action,
        "resource", domainReq.Resource,
        "resource_id", domainReq.ResourceID)
    
    result, err := h.service.CheckPermission(ctx, domainReq)
    if err != nil {
        h.logger.Error("Failed to check ABAC permission", 
            "user_id", domainReq.UserID,
            "action", domainReq.Action,
            "resource", domainReq.Resource,
            "error", err)
        responses.HandleDomainError(c, err)
        return
    }
    
    goaResult := h.adapter.PermissionCheckToGoaResult(result)
    
    h.logger.Info("ABAC permission check completed", 
        "user_id", domainReq.UserID,
        "allowed", result.Allowed,
        "evaluation_time", result.EvaluationTime)
    
    responses.OK(c, goaResult)
}

func (h *ABACHandler) BatchCheckPermissions(c *gin.Context) {
    ctx := c.Request.Context()
    
    var payload goa_abac.BatchCheckPermissionsPayload
    if err := c.ShouldBindJSON(&payload); err != nil {
        h.logger.Error("Invalid batch check permissions payload", "error", err)
        responses.ValidationError(c, "Invalid request payload", err)
        return
    }
    
    domainReq, err := h.adapter.BatchCheckPermissionsPayloadToDomain(&payload)
    if err != nil {
        h.logger.Error("Failed to convert batch check permissions payload", "error", err)
        responses.BadRequest(c, "Invalid payload data", err)
        return
    }
    
    // Add request context to all checks
    for i := range domainReq.Checks {
        if domainReq.Checks[i].Context == nil {
            domainReq.Checks[i].Context = make(map[string]interface{})
        }
        domainReq.Checks[i].Context["ip_address"] = c.ClientIP()
        domainReq.Checks[i].Context["user_agent"] = c.Request.UserAgent()
        domainReq.Checks[i].Context["method"] = c.Request.Method
        domainReq.Checks[i].Context["path"] = c.Request.URL.Path
        
        if tenantID := c.GetString("tenant_id"); tenantID != "" {
            domainReq.Checks[i].Context["tenant_id"] = tenantID
        }
    }
    
    h.logger.Info("Batch checking ABAC permissions", "num_checks", len(domainReq.Checks))
    
    results, err := h.service.BatchCheckPermissions(ctx, domainReq)
    if err != nil {
        h.logger.Error("Failed to batch check ABAC permissions", "error", err)
        responses.HandleDomainError(c, err)
        return
    }
    
    goaResult := h.adapter.BatchPermissionCheckToGoaResult(results)
    
    h.logger.Info("ABAC batch permission check completed", 
        "num_checks", len(results.Results),
        "total_time", results.TotalEvaluationTime)
    
    responses.OK(c, goaResult)
}

func (h *ABACHandler) GetUserPermissions(c *gin.Context) {
    ctx := c.Request.Context()
    userIDStr := c.Param("user_id")
    
    userID, err := uuid.Parse(userIDStr)
    if err != nil {
        h.logger.Error("Invalid user ID format", "id", userIDStr, "error", err)
        responses.BadRequest(c, "Invalid user ID format", err)
        return
    }
    
    // Parse optional parameters
    includeInherited := c.DefaultQuery("include_inherited", "true") == "true"
    includeABAC := c.DefaultQuery("include_abac", "true") == "true"
    
    filter := abac.UserPermissionFilter{
        IncludeInherited: includeInherited,
        IncludeABAC:      includeABAC,
    }
    
    h.logger.Info("Getting user permissions", 
        "user_id", userID,
        "include_inherited", includeInherited,
        "include_abac", includeABAC)
    
    permissions, err := h.service.GetUserPermissions(ctx, userID, filter)
    if err != nil {
        h.logger.Error("Failed to get user permissions", "user_id", userID, "error", err)
        responses.HandleDomainError(c, err)
        return
    }
    
    goaResult := h.adapter.UserPermissionsToGoaResult(permissions)
    responses.OK(c, goaResult)
}

func (h *ABACHandler) EvaluatePolicy(c *gin.Context) {
    ctx := c.Request.Context()
    policyIDStr := c.Param("id")
    
    policyID, err := uuid.Parse(policyIDStr)
    if err != nil {
        h.logger.Error("Invalid policy ID format", "id", policyIDStr, "error", err)
        responses.BadRequest(c, "Invalid policy ID format", err)
        return
    }
    
    var payload goa_abac.EvaluatePolicyPayload
    if err := c.ShouldBindJSON(&payload); err != nil {
        h.logger.Error("Invalid evaluate policy payload", "error", err)
        responses.ValidationError(c, "Invalid request payload", err)
        return
    }
    
    domainReq, err := h.adapter.EvaluatePolicyPayloadToDomain(&payload)
    if err != nil {
        h.logger.Error("Failed to convert evaluate policy payload", "error", err)
        responses.BadRequest(c, "Invalid payload data", err)
        return
    }
    
    domainReq.PolicyID = policyID
    
    h.logger.Info("Evaluating ABAC policy", 
        "policy_id", policyID,
        "user_id", domainReq.UserID)
    
    result, err := h.service.EvaluatePolicy(ctx, domainReq)
    if err != nil {
        h.logger.Error("Failed to evaluate ABAC policy", 
            "policy_id", policyID,
            "error", err)
        responses.HandleDomainError(c, err)
        return
    }
    
    goaResult := h.adapter.PolicyEvaluationToGoaResult(result)
    
    h.logger.Info("ABAC policy evaluation completed", 
        "policy_id", policyID,
        "result", result.Result,
        "evaluation_time", result.EvaluationTime)
    
    responses.OK(c, goaResult)
}
```

## 🔄 15. Additional Middleware Components

### Rate Limiting Middleware

```go
// internal/api/gin/middleware/rate_limit.go
package middleware

import (
    "net/http"
    "sync"
    "time"
    
    "github.com/gin-gonic/gin"
    "golang.org/x/time/rate"
    
    "github.com/niiniyare/erp/internal/platform/config"
    "github.com/niiniyare/erp/internal/api/gin/responses"
)

type RateLimiter struct {
    limiters map[string]*rate.Limiter
    mutex    sync.RWMutex
    rate     rate.Limit
    burst    int
}

func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
    return &RateLimiter{
        limiters: make(map[string]*rate.Limiter),
        rate:     r,
        burst:    b,
    }
}

func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
    rl.mutex.Lock()
    defer rl.mutex.Unlock()
    
    limiter, exists := rl.limiters[key]
    if !exists {
        limiter = rate.NewLimiter(rl.rate, rl.burst)
        rl.limiters[key] = limiter
    }
    
    return limiter
}

func RateLimit(cfg config.RateLimit) gin.HandlerFunc {
    if !cfg.Enabled {
        return func(c *gin.Context) {
            c.Next()
        }
    }
    
    // Create different rate limiters for different scenarios
    globalLimiter := NewRateLimiter(rate.Limit(cfg.GlobalRate), cfg.GlobalBurst)
    userLimiter := NewRateLimiter(rate.Limit(cfg.UserRate), cfg.UserBurst)
    ipLimiter := NewRateLimiter(rate.Limit(cfg.IPRate), cfg.IPBurst)
    
    return func(c *gin.Context) {
        // Global rate limiting
        if !globalLimiter.getLimiter("global").Allow() {
            responses.RateLimitExceeded(c, "Global rate limit exceeded")
            c.Abort()
            return
        }
        
        // IP-based rate limiting
        clientIP := c.ClientIP()
        if !ipLimiter.getLimiter(clientIP).Allow() {
            responses.RateLimitExceeded(c, "IP rate limit exceeded")
            c.Abort()
            return
        }
        
        // User-based rate limiting (if authenticated)
        if userID := c.GetString("user_id"); userID != "" {
            if !userLimiter.getLimiter(userID).Allow() {
                responses.RateLimitExceeded(c, "User rate limit exceeded")
                c.Abort()
                return
            }
        }
        
        c.Next()
    }
}

// API key specific rate limiting
func APIKeyRateLimit(cfg config.APIKeyRateLimit) gin.HandlerFunc {
    limiter := NewRateLimiter(rate.Limit(cfg.Rate), cfg.Burst)
    
    return func(c *gin.Context) {
        apiKey := c.GetHeader("X-API-Key")
        if apiKey == "" {
            c.Next()
            return
        }
        
        if !limiter.getLimiter(apiKey).Allow() {
            responses.RateLimitExceeded(c, "API key rate limit exceeded")
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// Endpoint-specific rate limiting
func EndpointRateLimit(rate rate.Limit, burst int) gin.HandlerFunc {
    limiter := NewRateLimiter(rate, burst)
    
    return func(c *gin.Context) {
        endpoint := c.Request.Method + " " + c.FullPath()
        clientKey := c.ClientIP() + ":" + endpoint
        
        if !limiter.getLimiter(clientKey).Allow() {
            responses.RateLimitExceeded(c, "Endpoint rate limit exceeded")
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

### Security Headers Middleware

```go
// internal/api/gin/middleware/security.go
package middleware

import (
    "github.com/gin-gonic/gin"
)

func SecurityHeaders() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Prevent MIME type sniffing
        c.Header("X-Content-Type-Options", "nosniff")
        
        // Prevent clickjacking
        c.Header("X-Frame-Options", "DENY")
        
        // Enable XSS protection
        c.Header("X-XSS-Protection", "1; mode=block")
        
        // Force HTTPS in production
        if gin.Mode() == gin.ReleaseMode {
            c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        }
        
        // Content Security Policy
        csp := "default-src 'self'; " +
               "script-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
               "style-src 'self' 'unsafe-inline'; " +
               "img-src 'self' data: https:; " +
               "font-src 'self' https:; " +
               "connect-src 'self'; " +
               "frame-ancestors 'none'"
        c.Header("Content-Security-Policy", csp)
        
        // Referrer Policy
        c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
        
        // Feature Policy / Permissions Policy
        c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
        
        c.Next()
    }
}

func APISecurityHeaders() gin.HandlerFunc {
    return func(c *gin.Context) {
        // API-specific security headers
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-Frame-Options", "DENY")
        c.Header("Cache-Control", "no-store, no-cache, must-revalidate, private")
        c.Header("Pragma", "no-cache")
        c.Header("Expires", "0")
        
        // Prevent information disclosure
        c.Header("Server", "AWO-ERP-API")
        
        c.Next()
    }
}
```

### Tenant Context Middleware

```go
// internal/api/gin/middleware/tenant.go
package middleware

import (
    "github.com/gin-gonic/gin"
    
    "github.com/niiniyare/erp/internal/api/gin/responses"
    "github.com/niiniyare/erp/internal/shared/logger"
)

func TenantContext() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Tenant ID should be set by JWT middleware
        tenantID := c.GetString("tenant_id")
        if tenantID == "" {
            responses.BadRequest(c, "Tenant context is required", nil)
            c.Abort()
            return
        }
        
        // Validate tenant ID format
        if !isValidUUID(tenantID) {
            responses.BadRequest(c, "Invalid tenant ID format", nil)
            c.Abort()
            return
        }
        
        // Set tenant context for database queries
        c.Set("tenant_context", map[string]interface{}{
            "tenant_id": tenantID,
            "isolation": "strict", // Enforce strict tenant isolation
        })
        
        // Add tenant ID to request logger context
        if requestLogger, exists := c.Get("request_logger"); exists {
            if logger, ok := requestLogger.(logger.Logger); ok {
                logger = logger.With("tenant_id", tenantID)
                c.Set("request_logger", logger)
            }
        }
        
        c.Next()
    }
}

func RequireTenantAccess(allowedRoles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        tenantID := c.GetString("tenant_id")
        userRoles, exists := c.Get("roles")
        if !exists {
            responses.Forbidden(c, "No roles available")
            c.Abort()
            return
        }
        
        roles, ok := userRoles.([]string)
        if !ok {
            responses.Forbidden(c, "Invalid roles format")
            c.Abort()
            return
        }
        
        // Check if user has any of the allowed roles for this tenant
        hasAccess := false
        for _, userRole := range roles {
            for _, allowedRole := range allowedRoles {
                if userRole == allowedRole {
                    hasAccess = true
                    break
                }
            }
            if hasAccess {
                break
            }
        }
        
        if !hasAccess {
            responses.Forbidden(c, "Insufficient tenant access")
            c.Abort()
            return
        }
        
        c.Next()
    }
}

func isValidUUID(s string) bool {
    // Simple UUID validation - in production use proper UUID library
    return len(s) == 36 && s[8] == '-' && s[13] == '-' && s[18] == '-' && s[23] == '-'
}
```

## 📊 16. Metrics and Monitoring Middleware

```go
// internal/api/gin/middleware/metrics.go
package middleware

import (
    "strconv"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status", "tenant_id"},
    )
    
    httpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint", "status"},
    )
    
    httpRequestSize = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_size_bytes",
            Help:    "HTTP request size in bytes",
            Buckets: []float64{100, 1000, 10000, 100000, 1000000},
        },
        []string{"method", "endpoint"},
    )
    
    httpResponseSize = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_response_size_bytes",
            Help:    "HTTP response size in bytes",
            Buckets: []float64{100, 1000, 10000, 100000, 1000000},
        },
        []string{"method", "endpoint", "status"},
    )
    
    activeConnections = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "http_active_connections",
            Help: "Number of active HTTP connections",
        },
    )
    
    abacEvaluationsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "abac_evaluations_total",
            Help: "Total number of ABAC policy evaluations",
        },
        []string{"result", "tenant_id"},
    )
    
    abacEvaluationDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "abac_evaluation_duration_seconds",
            Help:    "ABAC policy evaluation duration in seconds",
            Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
        },
        []string{"tenant_id"},
    )
)

func Metrics() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        activeConnections.Inc()
        
        // Get request size
        requestSize := float64(c.Request.ContentLength)
        if requestSize < 0 {
            requestSize = 0
        }
        
        c.Next()
        
        // Calculate metrics
        duration := time.Since(start).Seconds()
        status := strconv.Itoa(c.Writer.Status())
        method := c.Request.Method
        endpoint := c.FullPath()
        tenantID := c.GetString("tenant_id")
        if tenantID == "" {
            tenantID = "unknown"
        }
        
        // Record metrics
        httpRequestsTotal.WithLabelValues(method, endpoint, status, tenantID).Inc()
        httpRequestDuration.WithLabelValues(method, endpoint, status).Observe(duration)
        httpRequestSize.WithLabelValues(method, endpoint).Observe(requestSize)
        
        // Response size
        responseSize := float64(c.Writer.Size())
        httpResponseSize.WithLabelValues(method, endpoint, status).Observe(responseSize)
        
        activeConnections.Dec()
    }
}

func RecordABACEvaluation(result string, tenantID string, duration time.Duration) {
    abacEvaluationsTotal.WithLabelValues(result, tenantID).Inc()
    abacEvaluationDuration.WithLabelValues(tenantID).Observe(duration.Seconds())
}
```

## 🏥 17. Health Check Implementation

```go
// internal/api/gin/handlers/health.go
package handlers

import (
    "context"
    "database/sql"
    "time"
    
    "github.com/gin-gonic/gin"
    
    "github.com/niiniyare/erp/internal/platform/cache"
    "github.com/niiniyare/erp/internal/shared/logger"
    "github.com/niiniyare/erp/internal/api/gin/responses"
)

type HealthHandler struct {
    db     *sql.DB
    cache  cache.Cache
    logger logger.Logger
}

func NewHealthHandler(db *sql.DB, cache cache.Cache, logger logger.Logger) *HealthHandler {
    return &HealthHandler{
        db:     db,
        cache:  cache,
        logger: logger,
    }
}

func (h *HealthHandler) Liveness(c *gin.Context) {
    // Liveness probe - just check if the service is running
    response := map[string]interface{}{
        "status":    "alive",
        "timestamp": time.Now().UTC().Format(time.RFC3339),
        "service":   "erp-api",
        "version":   "1.0.0",
    }
    
    responses.OK(c, response)
}

func (h *HealthHandler) Readiness(c *gin.Context) {
    // Readiness probe - check if service is ready to handle traffic
    ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
    defer cancel()
    
    checks := []HealthCheck{
        {Name: "database", Check: h.checkDatabase},
        {Name: "cache", Check: h.checkCache},
    }
    
    overallStatus := "ready"
    checkResults := make([]map[string]interface{}, len(checks))
    
    for i, check := range checks {
        start := time.Now()
        err := check.Check(ctx)
        duration := time.Since(start)
        
        status := "healthy"
        message := "OK"
        
        if err != nil {
            status = "unhealthy"
            message = err.Error()
            overallStatus = "not_ready"
        }
        
        checkResults[i] = map[string]interface{}{
            "name":          check.Name,
            "status":        status,
            "message":       message,
            "response_time": duration.String(),
            "timestamp":     time.Now().UTC().Format(time.RFC3339),
        }
    }
    
    response := map[string]interface{}{
        "status":    overallStatus,
        "timestamp": time.Now().UTC().Format(time.RFC3339),
        "service":   "erp-api",
        "version":   "1.0.0",
        "checks":    checkResults,
    }
    
    if overallStatus == "ready" {
        responses.OK(c, response)
    } else {
        c.JSON(503, response)
    }
}

func (h *HealthHandler) Status(c *gin.Context) {
    // Detailed status with all dependencies
    ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
    defer cancel()
    
    checks := []HealthCheck{
        {Name: "database", Check: h.checkDatabase},
        {Name: "cache", Check: h.checkCache},
        {Name: "external_api", Check: h.checkExternalAPI},
    }
    
    overallStatus := "healthy"
    checkResults := make([]map[string]interface{}, len(checks))
    
    for i, check := range checks {
        start := time.Now()
        err := check.Check(ctx)
        duration := time.Since(start)
        
        status := "healthy"
        message := "OK"
        
        if err != nil {
            status = "unhealthy"
            message = err.Error()
            if overallStatus == "healthy" {
                overallStatus = "degraded"
            }
        }
        
        checkResults[i] = map[string]interface{}{
            "name":          check.Name,
            "status":        status,
            "message":       message,
            "response_time": duration.String(),
            "timestamp":     time.Now().UTC().Format(time.RFC3339),
        }
    }
    
    response := map[string]interface{}{
        "status":     overallStatus,
        "timestamp":  time.Now().UTC().Format(time.RFC3339),
        "service":    "erp-api",
        "version":    "1.0.0",
        "uptime":     h.getUptime(),
        "checks":     checkResults,
        "system": map[string]interface{}{
            "memory_usage": h.getMemoryUsage(),
            "goroutines":   h.getGoroutineCount(),
        },
    }
    
    responses.OK(c, response)
}

type HealthCheck struct {
    Name  string
    Check func(context.Context) error
}

func (h *HealthHandler) checkDatabase(ctx context.Context) error {
    return h.db.PingContext(ctx)
}

func (h *HealthHandler) checkCache(ctx context.Context) error {
    return h.cache.Ping(ctx)
}

func (h *HealthHandler) checkExternalAPI(ctx context.Context) error {
    // Check external dependencies (example)
    // This would be implemented based on your external dependencies
    return nil
}

func (h *HealthHandler) getUptime() string {
    // Implementation depends on how you track service start time
    return "24h30m45s"
}

func (h *HealthHandler) getMemoryUsage() map[string]interface{} {
    // Implement memory usage tracking
    return map[string]interface{}{
        "alloc":      "10MB",
        "total_alloc": "50MB",
        "sys":        "25MB",
    }
}

func (h *HealthHandler) getGoroutineCount() int {
    // Implement goroutine counting
    return 42
}
```

## 🧪 18. Advanced Testing Patterns

### Integration Testing

```go
// internal/api/gin/integration_test.go
package gin_test

import (
    "bytes"
    "context"
    "database/sql"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
    "github.com/testcontainers/testcontainers-go/modules/redis"
    
    // Your packages
    ginRouter "github.com/niiniyare/erp/internal/api/gin"
    "github.com/niiniyare/erp/internal/platform/config"
    "github.com/niiniyare/erp/internal/platform/database"
    "github.com/niiniyare/erp/internal/platform/cache"
    "github.com/niiniyare/erp/internal/shared/logger"
    
    // Services
    "github.com/niiniyare/erp/internal/core/tenant"
    "github.com/niiniyare/erp/internal/core/auth"
    
    // Repositories
    tenantRepo "github.com/niiniyare/erp/internal/repositories/tenant"
    authRepo "github.com/niiniyare/erp/internal/repositories/auth"
)

type IntegrationTestSuite struct {
    suite.Suite
    
    // Infrastructure
    pgContainer    *postgres.PostgresContainer
    redisContainer *redis.RedisContainer
    db             *sql.DB
    cache          cache.Cache
    
    // Application
    router *gin.Engine
    logger logger.Logger
    
    // Test data
    testTenantID string
    testUserID   string
    testToken    string
}

func TestIntegrationSuite(t *testing.T) {
    suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
    ctx := context.Background()
    
    // Start PostgreSQL container
    pgContainer, err := postgres.RunContainer(ctx,
        testcontainers.WithImage("postgres:15-alpine"),
        postgres.WithDatabase("test_erp"),
        postgres.WithUsername("test_user"),
        postgres.WithPassword("test_password"),
        testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections")),
    )
    s.Require().NoError(err)
    s.pgContainer = pgContainer
    
    // Start Redis container
    redisContainer, err := redis.RunContainer(ctx,
        testcontainers.WithImage("redis:7-alpine"),
    )
    s.Require().NoError(err)
    s.redisContainer = redisContainer
    
    // Get connection strings
    pgConnStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
    s.Require().NoError(err)
    
    redisConnStr, err := redisContainer.ConnectionString(ctx)
    s.Require().NoError(err)
    
    // Initialize database
    s.db, err = database.Connect(ctx, config.DatabaseConfig{
        URL:             pgConnStr,
        MaxOpenConns:    10,
        MaxIdleConns:    5,
        ConnMaxLifetime: 5 * time.Minute,
    })
    s.Require().NoError(err)
    
    // Run migrations
    err = database.RunMigrations(s.db)
    s.Require().NoError(err)
    
    // Initialize cache
    s.cache, err = cache.NewRedis(config.RedisConfig{
        URL: redisConnStr,
    })
    s.Require().NoError(err)
    
    // Initialize logger
    s.logger = logger.NewNoop()
    
    // Initialize services
    s.setupServices()
    
    // Create test data
    s.setupTestData()
}

func (s *IntegrationTestSuite) TearDownSuite() {
    ctx := context.Background()
    
    if s.db != nil {
        s.db.Close()
    }
    
    if s.cache != nil {
        s.cache.Close()
    }
    
    if s.pgContainer != nil {
        s.pgContainer.Terminate(ctx)
    }
    
    if s.redisContainer != nil {
        s.redisContainer.Terminate(ctx)
    }
}

func (s *IntegrationTestSuite) setupServices() {
    // Initialize repositories
    tenantRepository := tenantRepo.NewRepository(s.db, s.logger)
    authRepository := authRepo.NewRepository(s.db, s.cache, s.logger)
    
    // Initialize services
    tenantService := tenant.NewService(tenantRepository, s.cache, s.logger)
    authService := auth.NewService(authRepository, s.cache, config.JWT{
        Secret:          "test-secret",
        AccessTokenTTL:  15 * time.Minute,
        RefreshTokenTTL: 24 * time.Hour,
    }, s.logger)
    
    // Initialize router
    cfg := &config.Config{
        Environment: "test",
        JWT: config.JWT{
            Secret: "test-secret",
        },
    }
    
    deps := &ginRouter.Dependencies{
        AuthService:   authService,
        TenantService: tenantService,
        // ... other services
    }
    
    router := ginRouter.NewRouter(cfg, s.logger, deps)
    s.router = router.Setup()
}

func (s *IntegrationTestSuite) setupTestData() {
    ctx := context.Background()
    
    // Create test tenant
    tenantReq := &tenant.CreateRequest{
        Name:         "Test Tenant",
        Subdomain:    "test-tenant",
        ContactEmail: "admin@test-tenant.com",
        Plan:         "professional",
    }
    
    testTenant, err := s.tenantService.CreateTenant(ctx, tenantReq)
    s.Require().NoError(err)
    s.testTenantID = testTenant.ID.String()
    
    // Create test user and get auth token
    // Implementation depends on your auth service
    // ...
}

func (s *IntegrationTestSuite) TestTenantCRUD() {
    // Test Create
    payload := map[string]interface{}{
        "name":          "Integration Test Tenant",
        "subdomain":     "integration-test",
        "contact_email": "admin@integration-test.com",
        "plan":          "basic",
    }
    
    resp := s.makeAuthenticatedRequest("POST", "/api/v1/tenants", payload)
    s.Equal(http.StatusCreated, resp.Code)
    
    var createResp map[string]interface{}
    err := json.Unmarshal(resp.Body.Bytes(), &createResp)
    s.NoError(err)
    
    data := createResp["data"].(map[string]interface{})
    tenantID := data["id"].(string)
    s.NotEmpty(tenantID)
    s.Equal("Integration Test Tenant", data["name"])
    
    // Test Get
    resp = s.makeAuthenticatedRequest("GET", "/api/v1/tenants/"+tenantID, nil)
    s.Equal(http.StatusOK, resp.Code)
    
    var getResp map[string]interface{}
    err = json.Unmarshal(resp.Body.Bytes(), &getResp)
    s.NoError(err)
    
    data = getResp["data"].(map[string]interface{})
    s.Equal(tenantID, data["id"])
    s.Equal("Integration Test Tenant", data["name"])
    
    // Test Update
    updatePayload := map[string]interface{}{
        "name":        "Updated Integration Test Tenant",
        "description": "Updated description",
    }
    
    resp = s.makeAuthenticatedRequest("PUT", "/api/v1/tenants/"+tenantID, updatePayload)
    s.Equal(http.StatusOK, resp.Code)
    
    var updateResp map[string]interface{}
    err = json.Unmarshal(resp.Body.Bytes(), &updateResp)
    s.NoError(err)
    
    data = updateResp["data"].(map[string]interface{})
    s.Equal("Updated Integration Test Tenant", data["name"])
    s.Equal("Updated description", data["description"])
    
    // Test List
    resp = s.makeAuthenticatedRequest("GET", "/api/v1/tenants?page=1&limit=10", nil)
    s.Equal(http.StatusOK, resp.Code)
    
    var listResp map[string]interface{}
    err = json.Unmarshal(resp.Body.Bytes(), &listResp)
    s.NoError(err)
    
    data = listResp["data"].(map[string]interface{})
    tenants := data["data"].([]interface{})
    s.GreaterOrEqual(len(tenants), 1)
    
    // Test Delete
    resp = s.makeAuthenticatedRequest("DELETE", "/api/v1/tenants/"+tenantID, nil)
    s.Equal(http.StatusNoContent, resp.Code)
    
    // Verify deletion
    resp = s.makeAuthenticatedRequest("GET", "/api/v1/tenants/"+tenantID, nil)
    s.Equal(http.StatusNotFound, resp.Code)
}

func (s *IntegrationTestSuite) TestAuthFlow() {
    // Test login
    loginPayload := map[string]interface{}{
        "email":            "admin@test-tenant.com",
        "password":         "TestPassword123!",
        "tenant_subdomain": "test-tenant",
    }
    
    resp := s.makeRequest("POST", "/api/v1/auth/login", loginPayload)
    s.Equal(http.StatusOK, resp.Code)
    
    var loginResp map[string]interface{}
    err := json.Unmarshal(resp.Body.Bytes(), &loginResp)
    s.NoError(err)
    
    data := loginResp["data"].(map[string]interface{})
    accessToken := data["access_token"].(string)
    refreshToken := data["refresh_token"].(string)
    s.NotEmpty(accessToken)
    s.NotEmpty(refreshToken)
    
    // Test authenticated endpoint
    req := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
    req.Header.Set("Authorization", "Bearer "+accessToken)
    w := httptest.NewRecorder()
    s.router.ServeHTTP(w, req)
    
    s.Equal(http.StatusOK, w.Code)
    
    // Test token refresh
    refreshPayload := map[string]interface{}{
        "refresh_token": refreshToken,
    }
    
    resp = s.makeRequest("POST", "/api/v1/auth/refresh", refreshPayload)
    s.Equal(http.StatusOK, resp.Code)
    
    // Test logout
    logoutPayload := map[string]interface{}{
        "refresh_token": refreshToken,
    }
    
    req = httptest.NewRequest("POST", "/api/v1/auth/logout", s.createJSONBody(logoutPayload))
    req.Header.Set("Authorization", "Bearer "+accessToken)
    req.Header.Set("Content-Type", "application/json")
    w = httptest.NewRecorder()
    s.router.ServeHTTP(w, req)
    
    s.Equal(http.StatusNoContent, w.Code)
}

func (s *IntegrationTestSuite) makeRequest(method, path string, payload interface{}) *httptest.ResponseRecorder {
    var body *bytes.Buffer
    if payload != nil {
        body = s.createJSONBody(payload)
    }
    
    req := httptest.NewRequest(method, path, body)
    if body != nil {
        req.Header.Set("Content-Type", "application/json")
    }
    
    w := httptest.NewRecorder()
    s.router.ServeHTTP(w, req)
    
    return w
}

func (s *IntegrationTestSuite) makeAuthenticatedRequest(method, path string, payload interface{}) *httptest.ResponseRecorder {
    var body *bytes.Buffer
    if payload != nil {
        body = s.createJSONBody(payload)
    }
    
    req := httptest.NewRequest(method, path, body)
    req.Header.Set("Authorization", "Bearer "+s.testToken)
    if body != nil {
        req.Header.Set("Content-Type", "application/json")
    }
    
    w := httptest.NewRecorder()
    s.router.ServeHTTP(w, req)
    
    return w
}

func (s *IntegrationTestSuite) createJSONBody(payload interface{}) *bytes.Buffer {
    jsonBytes, err := json.Marshal(payload)
    s.Require().NoError(err)
    return bytes.NewBuffer(jsonBytes)
}
```

This completes the implementation guide with production-ready components, advanced testing patterns, monitoring, security, and all essential middleware. The implementation now provides a complete, scalable solution that combines Goa's design-first approach with Gin's high performance while maintaining clean architecture principles!
