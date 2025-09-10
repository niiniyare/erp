# Day 2: Native Middleware Architecture Design

**Date:** 2024-01-15  
**Status:** Design Phase  
**Goal:** Design replacement for Gin tenant middleware

## 🎯 Design Requirements

### **Functional Requirements:**
1. **Tenant Extraction**: Support both `X-Tenant-ID` header and subdomain
2. **Priority Order**: Header takes priority over subdomain
3. **Whitelist Support**: Public endpoints bypass tenant validation
4. **Context Management**: Use `shared.WithTenantID()` for context
5. **Database Session**: Use `store.SetTenantContextFromCtx()`
6. **Error Handling**: Maintain same error responses as current Gin version

### **Non-Functional Requirements:**
1. **Performance**: <5ms overhead per request
2. **Security**: Zero cross-tenant data access
3. **Compatibility**: Drop-in replacement for Gin middleware
4. **Maintainability**: Clean, testable code structure

## 🏗️ Proposed Architecture

### **New Middleware Signature:**
```go
func TenantMiddleware(
    tenantService tenant.Service,
    store db.Store,
    whitelist *EndpointWhitelist,
) func(http.Handler) http.Handler
```

### **Request Flow Design:**
```
HTTP Request
     ↓
[1] Check Whitelist
     ↓ (if not public)
[2] Extract Tenant ID
     ↓ (header priority, subdomain fallback)  
[3] Validate Tenant Exists
     ↓ (using tenantService)
[4] Set Context
     ↓ (using shared.WithTenantID)
[5] Set Database Session  
     ↓ (using store.SetTenantContextFromCtx)
[6] Continue to Handler
```

## 📋 Implementation Plan

### **File Structure:**
```
internal/platform/middleware/
├── tenant_native.go          # New native middleware
├── whitelist.go              # Public endpoint whitelist
├── tenant_native_test.go     # Comprehensive tests
└── whitelist_test.go         # Whitelist tests
```

### **Core Functions:**
1. **`TenantMiddleware()`** - Main middleware function
2. **`extractTenantID()`** - Header + subdomain extraction
3. **`validateTenantFormat()`** - UUID validation
4. **`extractFromSubdomain()`** - Subdomain parsing
5. **`extractFromHeader()`** - Header parsing

## 🔧 Detailed Function Design

### **1. Main Middleware Function:**
```go
func TenantMiddleware(
    tenantService tenant.Service,
    store db.Store,
    whitelist *EndpointWhitelist,
) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // 1. Check if public endpoint
            if whitelist.IsPublicEndpoint(r.Method, r.URL.Path) {
                next.ServeHTTP(w, r)
                return
            }
            
            // 2. Extract tenant ID
            tenantIDStr, err := extractTenantID(r)
            if err != nil {
                http.Error(w, "Tenant resolution failed", http.StatusBadRequest)
                return
            }
            
            // 3. Parse and validate tenant ID
            tenantID, err := uuid.Parse(tenantIDStr)
            if err != nil {
                http.Error(w, "Invalid tenant ID format", http.StatusBadRequest)
                return
            }
            
            // 4. Validate tenant exists
            _, err = tenantService.GetTenantByID(r.Context(), tenantID)
            if err != nil {
                http.Error(w, "Tenant not found", http.StatusNotFound)
                return
            }
            
            // 5. Set context
            ctx := shared.WithTenantID(r.Context(), tenantID)
            
            // 6. Set database session
            if err := store.SetTenantContextFromCtx(ctx); err != nil {
                http.Error(w, "Database context failed", http.StatusInternalServerError)
                return
            }
            
            // 7. Continue with enriched context
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### **2. Tenant Extraction Function:**
```go
func extractTenantID(r *http.Request) (string, error) {
    // Priority 1: X-Tenant-ID header
    if tenantID := r.Header.Get("X-Tenant-ID"); tenantID != "" {
        return strings.TrimSpace(tenantID), nil
    }
    
    // Priority 2: Subdomain extraction
    return extractFromSubdomain(r.Host)
}
```

### **3. Subdomain Extraction (Enhanced):**
```go
func extractFromSubdomain(host string) (string, error) {
    // Remove port if present
    if colonIndex := strings.Index(host, ":"); colonIndex != -1 {
        host = host[:colonIndex]
    }
    
    // Split by dots
    parts := strings.Split(host, ".")
    if len(parts) < 3 {
        return "", fmt.Errorf("invalid subdomain format: need at least 3 parts")
    }
    
    // Handle different patterns:
    // bo.tenant1.domain.com → tenant1
    // portal.tenant2.domain.com → tenant2  
    // tenant3.domain.com → tenant3
    
    if parts[0] == "bo" || parts[0] == "portal" {
        if len(parts) < 4 {
            return "", fmt.Errorf("missing tenant in subdomain")
        }
        return parts[1], nil // Second part is tenant
    }
    
    // Otherwise, first part is the tenant
    return parts[0], nil
}
```

### **4. Whitelist System Design:**
```go
type EndpointWhitelist struct {
    patterns     []string
    exactMatches map[string]bool
    compiled     []*regexp.Regexp // Pre-compiled patterns for performance
}

func (w *EndpointWhitelist) IsPublicEndpoint(method, path string) bool {
    fullEndpoint := method + " " + path
    
    // Check exact matches first (O(1) lookup)
    if w.exactMatches[fullEndpoint] || w.exactMatches[path] {
        return true
    }
    
    // Check compiled patterns
    for _, regex := range w.compiled {
        if regex.MatchString(fullEndpoint) || regex.MatchString(path) {
            return true
        }
    }
    
    return false
}
```

## 🧪 Testing Strategy

### **Unit Tests Required:**
1. **Header Extraction Tests:**
   - Valid `X-Tenant-ID` → returns tenant
   - Invalid format → error
   - Empty header → falls back to subdomain

2. **Subdomain Extraction Tests:**
   - `bo.tenant1.domain.com` → `tenant1`
   - `portal.tenant2.domain.com` → `tenant2`
   - `tenant3.domain.com` → `tenant3`
   - Invalid formats → errors

3. **Whitelist Tests:**
   - Public endpoints bypass tenant validation
   - Protected endpoints require tenant
   - Pattern matching works correctly

4. **Integration Tests:**
   - End-to-end with real tenant data
   - Database session variable verification
   - Context propagation validation

## 🚨 Error Handling Design

### **Error Response Format:**
```go
type TenantError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Status  int    `json:"-"`
}

// Error scenarios:
// 1. Missing tenant → 400 Bad Request
// 2. Invalid format → 400 Bad Request  
// 3. Tenant not found → 404 Not Found
// 4. Database error → 500 Internal Server Error
```

## 📊 Performance Considerations

### **Optimization Strategies:**
1. **Pre-compile regex patterns** in whitelist
2. **Cache tenant validation** (5-minute TTL)
3. **Reuse database connections** for session variables
4. **Minimize allocations** in hot path

### **Performance Targets:**
- **Latency**: <5ms per request
- **Memory**: <1KB per request
- **CPU**: <1% overhead
- **Throughput**: No degradation from Gin version

## 🔄 Migration Strategy

### **Phase 1**: Create new middleware alongside Gin
### **Phase 2**: Update server initialization to use native middleware  
### **Phase 3**: Remove Gin dependencies
### **Phase 4**: Cleanup and optimization

## ✅ Success Criteria

### **Functional Success:**
- [ ] All tenant extraction scenarios work
- [ ] Public endpoints accessible without tenant
- [ ] Protected endpoints require valid tenant  
- [ ] Database session variables set correctly
- [ ] Context propagation works end-to-end

### **Non-Functional Success:**
- [ ] Performance within 5% of Gin baseline
- [ ] Zero cross-tenant data access
- [ ] All tests pass with 95%+ coverage
- [ ] Clean, maintainable code structure

---

**Next Step:** Implement the core middleware function based on this design.