# Implementation Guide - Gin to Goa Migration

**Purpose:** Detailed technical guidance for implementing the Gin to Goa migration  
**Audience:** Solo developer working on the migration  
**Prerequisites:** Understanding of current Gin middleware and Goa framework

---

## 🏗️ ARCHITECTURE OVERVIEW

### **Current State (Gin):**
```
HTTP Request → Gin Engine → Gin Middleware → Goa Handlers → Response
                    ↓
               Tenant Context
                    ↓
            Database Session (RLS)
```

### **Target State (Native Goa):**
```
HTTP Request → Native Middleware Chain → Goa Handlers → Response
                        ↓
                 Tenant Context
                        ↓  
              Database Session (RLS)
```

---

## 📋 IMPLEMENTATION DETAILS

### **1. TENANT MIDDLEWARE IMPLEMENTATION**

#### **File:** `internal/platform/middleware/tenant_native.go`

```go
package middleware

import (
    "context"
    "net/http"
    "strings"
    "fmt"
    
    "github.com/niiniyare/erp/internal/shared"
    "github.com/niiniyare/erp/internal/core/tenant"
)

// TenantMiddleware creates HTTP middleware for tenant context injection
func TenantMiddleware(
    tenantService tenant.Service,
    whitelist *EndpointWhitelist,
) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Check if this is a public endpoint
            if whitelist.IsPublicEndpoint(r.Method, r.URL.Path) {
                next.ServeHTTP(w, r)
                return
            }
            
            // Extract tenant ID from request
            tenantID, err := extractTenantID(r)
            if err != nil {
                http.Error(w, fmt.Sprintf("Tenant resolution failed: %v", err), 
                    http.StatusBadRequest)
                return
            }
            
            // Validate tenant exists
            if !isValidTenant(tenantService, tenantID) {
                http.Error(w, "Tenant not found", http.StatusNotFound)
                return
            }
            
            // Set tenant context
            ctx := shared.SetTenantID(r.Context(), tenantID)
            
            // Set database session variable
            if err := setTenantDatabaseSession(ctx, tenantID); err != nil {
                http.Error(w, "Database context failed", 
                    http.StatusInternalServerError)
                return
            }
            
            // Continue with enriched context
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// extractTenantID extracts tenant ID from header or subdomain
func extractTenantID(r *http.Request) (string, error) {
    // Priority 1: X-Tenant-ID header
    if tenantID := r.Header.Get("X-Tenant-ID"); tenantID != "" {
        return validateTenantFormat(tenantID)
    }
    
    // Priority 2: Subdomain extraction
    return extractFromSubdomain(r.Host)
}

// validateTenantFormat validates tenant ID format (UUID)
func validateTenantFormat(tenantID string) (string, error) {
    // Add UUID validation logic here
    if len(tenantID) == 0 {
        return "", fmt.Errorf("empty tenant ID")
    }
    // Add more validation as needed
    return tenantID, nil
}

// extractFromSubdomain extracts tenant from subdomain
func extractFromSubdomain(host string) (string, error) {
    // Remove port if present
    if colonIndex := strings.Index(host, ":"); colonIndex != -1 {
        host = host[:colonIndex]
    }
    
    // Split by dots
    parts := strings.Split(host, ".")
    if len(parts) < 3 {
        return "", fmt.Errorf("invalid subdomain format")
    }
    
    // Extract tenant from first part if it's a known prefix
    subdomain := parts[0]
    if subdomain == "bo" || subdomain == "portal" {
        if len(parts) < 4 {
            return "", fmt.Errorf("missing tenant in subdomain")
        }
        return parts[1], nil // tenant is second part
    }
    
    // Otherwise, first part is the tenant
    return subdomain, nil
}

// isValidTenant checks if tenant exists
func isValidTenant(tenantService tenant.Service, tenantID string) bool {
    // Implement tenant validation logic
    // This could be a cache lookup or database query
    return true // Placeholder
}

// setTenantDatabaseSession sets PostgreSQL session variable
func setTenantDatabaseSession(ctx context.Context, tenantID string) error {
    // Implement database session variable setting
    // Execute: SET app.current_tenant_id = $1
    return nil // Placeholder
}
```

### **2. WHITELIST SYSTEM IMPLEMENTATION**

#### **File:** `internal/platform/middleware/whitelist.go`

```go
package middleware

import (
    "path/filepath"
    "strings"
)

// EndpointWhitelist manages public endpoints that don't require tenant context
type EndpointWhitelist struct {
    patterns     []string
    exactMatches map[string]bool
}

// NewEndpointWhitelist creates a new whitelist from configuration
func NewEndpointWhitelist(patterns []string, exactMatches []string) *EndpointWhitelist {
    exact := make(map[string]bool)
    for _, endpoint := range exactMatches {
        exact[endpoint] = true
    }
    
    return &EndpointWhitelist{
        patterns:     patterns,
        exactMatches: exact,
    }
}

// IsPublicEndpoint checks if an endpoint should bypass tenant validation
func (w *EndpointWhitelist) IsPublicEndpoint(method, path string) bool {
    // Combine method and path for checking
    fullPath := method + " " + path
    
    // Check exact matches first (faster)
    if w.exactMatches[fullPath] || w.exactMatches[path] {
        return true
    }
    
    // Check pattern matches
    for _, pattern := range w.patterns {
        if matched, _ := filepath.Match(pattern, fullPath); matched {
            return true
        }
        if matched, _ := filepath.Match(pattern, path); matched {
            return true
        }
    }
    
    return false
}
```

#### **File:** `config/public_endpoints.yaml`

```yaml
public_endpoints:
  patterns:
    - "GET /api/v1/health*"
    - "GET /api/v1/openapi/*"
    - "GET /swagger-ui/*"
    - "GET /api/v1/version"
  
  exact_matches:
    - "POST /api/v1/auth/login"
    - "POST /api/v1/auth/refresh"
    - "POST /api/v1/tenant/onboard"
    - "GET /api/v1/health"

validation:
  require_approval: true
  change_tracking: true
```

### **3. SERVER INTEGRATION**

#### **File:** `cmd/server/main.go` (Updated)

```go
package main

import (
    "context"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    // Remove: "github.com/gin-gonic/gin"
    
    "github.com/niiniyare/erp/internal/platform/middleware"
    "github.com/niiniyare/erp/internal/api/gen/http/..."
    // ... other imports
)

func main() {
    // Initialize configuration
    config := loadConfiguration()
    
    // Initialize services
    tenantService := initializeTenantService(config)
    // ... other services
    
    // Create whitelist
    whitelist := middleware.NewEndpointWhitelist(
        config.PublicEndpoints.Patterns,
        config.PublicEndpoints.ExactMatches,
    )
    
    // Create Goa endpoints
    endpoints := createGoaEndpoints()
    
    // Create HTTP mux
    var handler http.Handler = goahttp.NewMux(endpoints...)
    
    // CRITICAL: Middleware order matters
    // 1. Tenant context (MUST be first for security)
    handler = middleware.TenantMiddleware(tenantService, whitelist)(handler)
    
    // 2. Authentication (needs tenant context)
    handler = middleware.JWTAuthMiddleware(authService)(handler)
    
    // 3. Logging (needs user context)
    handler = middleware.LoggingMiddleware(logger)(handler)
    
    // 4. Tracing (needs full context)
    handler = middleware.TracingMiddleware(tracer)(handler)
    
    // 5. Metrics (needs all context for labels)
    handler = middleware.MetricsMiddleware(metrics)(handler)
    
    // 6. CORS (can be anywhere)
    handler = middleware.CORSMiddleware(corsConfig)(handler)
    
    // Create HTTP server
    server := &http.Server{
        Addr:    config.Server.Address,
        Handler: handler,
        // Add timeouts and other configuration
    }
    
    // Start server with graceful shutdown
    go func() {
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server failed to start: %v", err)
        }
    }()
    
    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    // Graceful shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := server.Shutdown(ctx); err != nil {
        log.Fatalf("Server forced to shutdown: %v", err)
    }
}
```

---

## 🧪 TESTING IMPLEMENTATION

### **Unit Testing Example**

#### **File:** `internal/platform/middleware/tenant_native_test.go`

```go
package middleware

import (
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestTenantMiddleware_HeaderExtraction(t *testing.T) {
    tests := []struct {
        name           string
        header         string
        expectedTenant string
        expectError    bool
    }{
        {
            name:           "Valid UUID header",
            header:         "550e8400-e29b-41d4-a716-446655440000",
            expectedTenant: "550e8400-e29b-41d4-a716-446655440000",
            expectError:    false,
        },
        {
            name:        "Empty header",
            header:      "",
            expectError: true,
        },
        {
            name:        "Invalid format",
            header:      "invalid-uuid",
            expectError: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest("GET", "/test", nil)
            if tt.header != "" {
                req.Header.Set("X-Tenant-ID", tt.header)
            }
            
            tenantID, err := extractTenantID(req)
            
            if tt.expectError {
                if err == nil {
                    t.Errorf("Expected error but got none")
                }
                return
            }
            
            if err != nil {
                t.Errorf("Unexpected error: %v", err)
            }
            
            if tenantID != tt.expectedTenant {
                t.Errorf("Expected tenant %s, got %s", tt.expectedTenant, tenantID)
            }
        })
    }
}

func TestTenantMiddleware_SubdomainExtraction(t *testing.T) {
    tests := []struct {
        name           string
        host           string
        expectedTenant string
        expectError    bool
    }{
        {
            name:           "BO subdomain",
            host:           "bo.tenant1.awoerp.com",
            expectedTenant: "tenant1",
            expectError:    false,
        },
        {
            name:           "Portal subdomain",
            host:           "portal.tenant2.awoerp.com",
            expectedTenant: "tenant2",
            expectError:    false,
        },
        {
            name:           "Direct tenant subdomain",
            host:           "tenant3.awoerp.com",
            expectedTenant: "tenant3",
            expectError:    false,
        },
        {
            name:        "Invalid format",
            host:        "awoerp.com",
            expectError: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tenantID, err := extractFromSubdomain(tt.host)
            
            if tt.expectError {
                if err == nil {
                    t.Errorf("Expected error but got none")
                }
                return
            }
            
            if err != nil {
                t.Errorf("Unexpected error: %v", err)
            }
            
            if tenantID != tt.expectedTenant {
                t.Errorf("Expected tenant %s, got %s", tt.expectedTenant, tenantID)
            }
        })
    }
}

func TestWhitelist_IsPublicEndpoint(t *testing.T) {
    whitelist := NewEndpointWhitelist(
        []string{"GET /api/v1/health*", "GET /swagger-ui/*"},
        []string{"POST /api/v1/auth/login", "GET /api/v1/version"},
    )
    
    tests := []struct {
        method   string
        path     string
        expected bool
    }{
        {"GET", "/api/v1/health", true},
        {"GET", "/api/v1/health/check", true},
        {"POST", "/api/v1/auth/login", true},
        {"GET", "/api/v1/version", true},
        {"GET", "/swagger-ui/index.html", true},
        {"GET", "/api/v1/finance/accounts", false},
        {"POST", "/api/v1/finance/transactions", false},
    }
    
    for _, tt := range tests {
        t.Run(tt.method+" "+tt.path, func(t *testing.T) {
            result := whitelist.IsPublicEndpoint(tt.method, tt.path)
            if result != tt.expected {
                t.Errorf("Expected %v, got %v", tt.expected, result)
            }
        })
    }
}
```

---

## 🔍 DEBUGGING & TROUBLESHOOTING

### **Common Issues and Solutions:**

#### **1. Context Not Propagating**
```go
// Problem: Tenant ID not available in handlers
// Solution: Ensure middleware runs before handlers
ctx := shared.SetTenantID(r.Context(), tenantID)
r = r.WithContext(ctx)

// In handler, verify:
tenantID, ok := shared.GetTenantID(ctx)
if !ok {
    // Context not set properly
}
```

#### **2. Database Session Not Working**
```sql
-- Check if session variable is set:
SELECT current_setting('app.current_tenant_id', true);

-- Should return the tenant UUID
-- If empty, session variable not being set
```

#### **3. Performance Issues**
```go
// Add timing to middleware:
start := time.Now()
defer func() {
    duration := time.Since(start)
    if duration > 50*time.Millisecond {
        log.Printf("Slow middleware: %v", duration)
    }
}()
```

#### **4. Whitelist Not Working**
```go
// Debug whitelist matching:
func (w *EndpointWhitelist) IsPublicEndpoint(method, path string) bool {
    fullPath := method + " " + path
    log.Printf("Checking: %s", fullPath)
    
    for _, pattern := range w.patterns {
        if matched, _ := filepath.Match(pattern, fullPath); matched {
            log.Printf("Matched pattern: %s", pattern)
            return true
        }
    }
    
    return false
}
```

---

## 📊 PERFORMANCE OPTIMIZATION

### **Connection Pool Optimization:**
```go
// Optimize database connections for tenant switching
type TenantAwarePool struct {
    pools map[string]*sql.DB
    mu    sync.RWMutex
}

func (p *TenantAwarePool) GetConnection(tenantID string) *sql.DB {
    // Implement tenant-specific connection pooling
}
```

### **Caching Strategies:**
```go
// Cache tenant validation results
type TenantCache struct {
    cache map[string]bool
    ttl   time.Duration
    mu    sync.RWMutex
}

func (c *TenantCache) IsValid(tenantID string) bool {
    // Implement caching logic
}
```

---

## ✅ VALIDATION CHECKLIST

### **Before Deployment:**
- [ ] All unit tests pass
- [ ] Integration tests with real data pass
- [ ] Performance benchmarks meet targets
- [ ] Error handling covers all scenarios
- [ ] Logging provides adequate debugging info
- [ ] Configuration is externalized
- [ ] Security review completed

### **Post-Deployment:**
- [ ] Monitor error rates
- [ ] Check performance metrics
- [ ] Verify tenant isolation
- [ ] Test public endpoints
- [ ] Validate database session variables

---

**This guide provides the technical foundation for your implementation. Refer back to it as you work through each phase of the migration.**