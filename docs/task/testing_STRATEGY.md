# Testing Strategy - Gin to Goa Migration

**Purpose:**  testing approach to ensure migration success  
**Focus:** Security, Performance, Functionality validation  
**Risk Level:** High (tenant isolation and API functionality are critical)

---

## 🎯 TESTING OBJECTIVES

### **Primary Goals:**
1. **🔒 Security:** Verify tenant isolation is bulletproof
2. **⚡ Performance:** Ensure no regression from Gin baseline  
3. **🔧 Functionality:** Confirm all APIs work exactly as before
4. **🛡️ Reliability:** Validate error handling and edge cases

### **Success Criteria:**
- Zero cross-tenant data access incidents
- <5% performance degradation from baseline
- 100% API functionality preserved
- All error scenarios handled gracefully

---

## 📋 TESTING PHASES

## **PHASE 1: UNIT TESTING (Day 8)**

### **Test Coverage Requirements:**
- **Minimum:** 90% code coverage
- **Focus:** All middleware logic, edge cases, error scenarios
- **Framework:** Go's built-in testing + testify for assertions

### **1.1 Tenant Extraction Testing**

#### **File:** `tenant_extraction_test.go`
```go
func TestExtractTenantID(t *testing.T) {
    tests := []struct {
        name        string
        headers     map[string]string
        host        string
        expected    string
        expectError bool
    }{
        // Header-based extraction
        {
            name:     "Valid UUID in header",
            headers:  map[string]string{"X-Tenant-ID": "550e8400-e29b-41d4-a716-446655440000"},
            expected: "550e8400-e29b-41d4-a716-446655440000",
        },
        {
            name:        "Invalid UUID format",
            headers:     map[string]string{"X-Tenant-ID": "invalid-uuid"},
            expectError: true,
        },
        {
            name:        "Empty header value",
            headers:     map[string]string{"X-Tenant-ID": ""},
            expectError: true,
        },
        
        // Subdomain-based extraction
        {
            name:     "BO subdomain with tenant",
            host:     "bo.tenant1.awoerp.com",
            expected: "tenant1",
        },
        {
            name:     "Portal subdomain with tenant",
            host:     "portal.tenant2.awoerp.com",
            expected: "tenant2",
        },
        {
            name:     "Direct tenant subdomain",
            host:     "tenant3.awoerp.com",
            expected: "tenant3",
        },
        {
            name:        "Missing tenant in subdomain",
            host:        "bo.awoerp.com",
            expectError: true,
        },
        {
            name:        "No subdomain",
            host:        "awoerp.com",
            expectError: true,
        },
        {
            name:        "Localhost development",
            host:        "localhost:8080",
            expectError: true,
        },
        
        // Priority testing (header beats subdomain)
        {
            name:     "Header takes priority over subdomain",
            headers:  map[string]string{"X-Tenant-ID": "header-tenant"},
            host:     "bo.subdomain-tenant.awoerp.com",
            expected: "header-tenant",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest("GET", "/test", nil)
            req.Host = tt.host
            
            for key, value := range tt.headers {
                req.Header.Set(key, value)
            }
            
            result, err := extractTenantID(req)
            
            if tt.expectError {
                assert.Error(t, err)
                return
            }
            
            assert.NoError(t, err)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### **1.2 Whitelist Testing**

#### **Test Cases:**
```go
func TestWhitelistFunctionality(t *testing.T) {
    whitelist := NewEndpointWhitelist(
        []string{"GET /api/v1/health*", "GET /swagger-ui/*"},
        []string{"POST /api/v1/auth/login", "GET /api/v1/version"},
    )
    
    publicEndpoints := []struct {
        method string
        path   string
    }{
        {"GET", "/api/v1/health"},
        {"GET", "/api/v1/health/check"},
        {"GET", "/api/v1/health/detailed"},
        {"POST", "/api/v1/auth/login"},
        {"GET", "/api/v1/version"},
        {"GET", "/swagger-ui/"},
        {"GET", "/swagger-ui/index.html"},
    }
    
    protectedEndpoints := []struct {
        method string
        path   string
    }{
        {"GET", "/api/v1/finance/accounts"},
        {"POST", "/api/v1/finance/transactions"},
        {"GET", "/api/v1/users"},
        {"POST", "/api/v1/auth/logout"}, // Not in whitelist
    }
    
    // Test public endpoints
    for _, endpoint := range publicEndpoints {
        t.Run(fmt.Sprintf("Public: %s %s", endpoint.method, endpoint.path), func(t *testing.T) {
            assert.True(t, whitelist.IsPublicEndpoint(endpoint.method, endpoint.path))
        })
    }
    
    // Test protected endpoints
    for _, endpoint := range protectedEndpoints {
        t.Run(fmt.Sprintf("Protected: %s %s", endpoint.method, endpoint.path), func(t *testing.T) {
            assert.False(t, whitelist.IsPublicEndpoint(endpoint.method, endpoint.path))
        })
    }
}
```

### **1.3 Middleware Integration Testing**

#### **Test Cases:**
```go
func TestTenantMiddleware_EndToEnd(t *testing.T) {
    // Create test middleware
    middleware := TenantMiddleware(mockTenantService, testWhitelist)
    
    // Test handler that checks context
    testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tenantID, ok := shared.GetTenantID(r.Context())
        if !ok {
            http.Error(w, "No tenant context", http.StatusInternalServerError)
            return
        }
        w.Header().Set("X-Test-Tenant", tenantID)
        w.WriteHeader(http.StatusOK)
    })
    
    handler := middleware(testHandler)
    
    tests := []struct {
        name           string
        method         string
        path           string
        headers        map[string]string
        expectedStatus int
        expectedTenant string
    }{
        {
            name:           "Valid tenant header",
            method:         "GET",
            path:           "/api/v1/finance/accounts",
            headers:        map[string]string{"X-Tenant-ID": "test-tenant"},
            expectedStatus: http.StatusOK,
            expectedTenant: "test-tenant",
        },
        {
            name:           "Public endpoint bypasses tenant",
            method:         "GET",
            path:           "/api/v1/health",
            expectedStatus: http.StatusOK,
        },
        {
            name:           "Missing tenant on protected endpoint",
            method:         "GET",
            path:           "/api/v1/finance/accounts",
            expectedStatus: http.StatusBadRequest,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest(tt.method, tt.path, nil)
            for key, value := range tt.headers {
                req.Header.Set(key, value)
            }
            
            w := httptest.NewRecorder()
            handler.ServeHTTP(w, req)
            
            assert.Equal(t, tt.expectedStatus, w.Code)
            
            if tt.expectedTenant != "" {
                assert.Equal(t, tt.expectedTenant, w.Header().Get("X-Test-Tenant"))
            }
        })
    }
}
```

---

## **PHASE 2: INTEGRATION TESTING (Day 9)**

### **2.1 Tenant Isolation Testing**

#### **Setup Requirements:**
```bash
# Create test database with multiple tenants
createdb erp_test
psql erp_test < db/migrations/all.sql

# Insert test data for multiple tenants
INSERT INTO tenants (id, name, subdomain) VALUES 
('tenant-1-uuid', 'Tenant 1', 'tenant1'),
('tenant-2-uuid', 'Tenant 2', 'tenant2'),
('tenant-3-uuid', 'Tenant 3', 'tenant3');

# Insert tenant-specific data
INSERT INTO finance_accounts (tenant_id, account_code, account_name) VALUES
('tenant-1-uuid', '1100', 'Tenant 1 Cash Account'),
('tenant-2-uuid', '1100', 'Tenant 2 Cash Account'),
('tenant-3-uuid', '1100', 'Tenant 3 Cash Account');
```

#### **Test Cases:**
```go
func TestTenantIsolation_Integration(t *testing.T) {
    // Setup test server with real database
    server := setupTestServer(t)
    defer server.Close()
    
    tests := []struct {
        name           string
        tenantID       string
        endpoint       string
        expectedData   []string // Account names we expect to see
        forbiddenData  []string // Account names we should NOT see
    }{
        {
            name:          "Tenant 1 sees only their data",
            tenantID:      "tenant-1-uuid",
            endpoint:      "/api/v1/finance/accounts",
            expectedData:  []string{"Tenant 1 Cash Account"},
            forbiddenData: []string{"Tenant 2 Cash Account", "Tenant 3 Cash Account"},
        },
        {
            name:          "Tenant 2 sees only their data",
            tenantID:      "tenant-2-uuid", 
            endpoint:      "/api/v1/finance/accounts",
            expectedData:  []string{"Tenant 2 Cash Account"},
            forbiddenData: []string{"Tenant 1 Cash Account", "Tenant 3 Cash Account"},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Make request with tenant header
            req, _ := http.NewRequest("GET", server.URL+tt.endpoint, nil)
            req.Header.Set("X-Tenant-ID", tt.tenantID)
            
            resp, err := http.DefaultClient.Do(req)
            require.NoError(t, err)
            defer resp.Body.Close()
            
            assert.Equal(t, http.StatusOK, resp.StatusCode)
            
            // Parse response
            var result AccountListResult
            err = json.NewDecoder(resp.Body).Decode(&result)
            require.NoError(t, err)
            
            // Check expected data is present
            responseText := string(respBytes)
            for _, expected := range tt.expectedData {
                assert.Contains(t, responseText, expected)
            }
            
            // Check forbidden data is NOT present
            for _, forbidden := range tt.forbiddenData {
                assert.NotContains(t, responseText, forbidden)
            }
        })
    }
}
```

### **2.2 Database Session Validation**

#### **Test Cases:**
```go
func TestDatabaseSessionVariable(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    tests := []struct {
        name     string
        tenantID string
    }{
        {"Valid UUID tenant", "550e8400-e29b-41d4-a716-446655440000"},
        {"Different tenant", "660e8400-e29b-41d4-a716-446655440001"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Set tenant context
            ctx := shared.SetTenantID(context.Background(), tt.tenantID)
            
            // Set database session variable
            err := setTenantDatabaseSession(ctx, tt.tenantID)
            require.NoError(t, err)
            
            // Verify session variable is set
            var currentTenant string
            err = db.QueryRowContext(ctx, 
                "SELECT current_setting('app.current_tenant_id', true)").Scan(&currentTenant)
            require.NoError(t, err)
            
            assert.Equal(t, tt.tenantID, currentTenant)
        })
    }
}
```

### **2.3 Public Endpoint Testing**

#### **Test Cases:**
```go
func TestPublicEndpoints_Integration(t *testing.T) {
    server := setupTestServer(t)
    defer server.Close()
    
    publicEndpoints := []struct {
        method   string
        path     string
        body     string
        expected int
    }{
        {"GET", "/api/v1/health", "", http.StatusOK},
        {"GET", "/api/v1/version", "", http.StatusOK},
        {"POST", "/api/v1/auth/login", `{"email":"test@example.com","password":"password"}`, http.StatusOK},
        {"GET", "/swagger-ui/", "", http.StatusOK},
    }
    
    for _, endpoint := range publicEndpoints {
        t.Run(fmt.Sprintf("%s %s", endpoint.method, endpoint.path), func(t *testing.T) {
            var body io.Reader
            if endpoint.body != "" {
                body = strings.NewReader(endpoint.body)
            }
            
            req, _ := http.NewRequest(endpoint.method, server.URL+endpoint.path, body)
            if endpoint.body != "" {
                req.Header.Set("Content-Type", "application/json")
            }
            
            // Deliberately NO tenant header
            resp, err := http.DefaultClient.Do(req)
            require.NoError(t, err)
            defer resp.Body.Close()
            
            assert.Equal(t, endpoint.expected, resp.StatusCode)
        })
    }
}
```

---

## **PHASE 3: PERFORMANCE TESTING (Day 9)**

### **3.1 Baseline Comparison**

#### **Setup:**
```bash
# Before migration (Gin baseline)
wrk -t4 -c100 -d30s --header "X-Tenant-ID: test-tenant" \
    http://localhost:8080/api/v1/finance/accounts > gin_baseline.txt

# After migration (Native Goa)
wrk -t4 -c100 -d30s --header "X-Tenant-ID: test-tenant" \
    http://localhost:8080/api/v1/finance/accounts > goa_results.txt
```

#### **Metrics to Compare:**
```go
type PerformanceMetrics struct {
    RequestsPerSecond float64
    LatencyP50        time.Duration
    LatencyP95        time.Duration
    LatencyP99        time.Duration
    ErrorRate         float64
    MemoryUsage       int64
    CPUUsage          float64
}

func BenchmarkTenantMiddleware(b *testing.B) {
    middleware := TenantMiddleware(mockTenantService, testWhitelist)
    handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))
    
    req := httptest.NewRequest("GET", "/api/v1/finance/accounts", nil)
    req.Header.Set("X-Tenant-ID", "test-tenant")
    
    b.ResetTimer()
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        w := httptest.NewRecorder()
        handler.ServeHTTP(w, req)
    }
}
```

### **3.2 Load Testing**

#### **Multi-Tenant Load Test:**
```bash
# Create load test script
cat > multi_tenant_load.lua << 'EOF'
-- Load testing with multiple tenants
tenants = {"tenant-1", "tenant-2", "tenant-3", "tenant-4", "tenant-5"}

request = function()
    tenant = tenants[math.random(#tenants)]
    wrk.headers["X-Tenant-ID"] = tenant
    return wrk.format(nil, "/api/v1/finance/accounts")
end
EOF

# Run multi-tenant load test
wrk -t12 -c400 -d60s --script multi_tenant_load.lua http://localhost:8080
```

### **3.3 Memory and CPU Profiling**

#### **Profiling Test:**
```go
func TestMiddleware_MemoryProfile(t *testing.T) {
    // Start CPU profile
    cpuFile, _ := os.Create("cpu.prof")
    defer cpuFile.Close()
    pprof.StartCPUProfile(cpuFile)
    defer pprof.StopCPUProfile()
    
    // Start memory profile
    defer func() {
        memFile, _ := os.Create("mem.prof")
        defer memFile.Close()
        pprof.WriteHeapProfile(memFile)
    }()
    
    // Run middleware many times
    middleware := TenantMiddleware(mockTenantService, testWhitelist)
    handler := middleware(testHandler)
    
    for i := 0; i < 10000; i++ {
        req := httptest.NewRequest("GET", "/test", nil)
        req.Header.Set("X-Tenant-ID", "test-tenant")
        w := httptest.NewRecorder()
        handler.ServeHTTP(w, req)
    }
}
```

---

## **PHASE 4: SECURITY TESTING**

### **4.1 Tenant Isolation Security Tests**

#### **Attack Scenarios:**
```go
func TestSecurity_TenantIsolationAttacks(t *testing.T) {
    server := setupTestServer(t)
    defer server.Close()
    
    attackTests := []struct {
        name           string
        attack         func() *http.Request
        expectedStatus int
        shouldFail     bool
    }{
        {
            name: "SQL Injection in tenant header",
            attack: func() *http.Request {
                req, _ := http.NewRequest("GET", "/api/v1/finance/accounts", nil)
                req.Header.Set("X-Tenant-ID", "'; DROP TABLE accounts; --")
                return req
            },
            expectedStatus: http.StatusBadRequest,
            shouldFail:     true,
        },
        {
            name: "Cross-tenant access attempt",
            attack: func() *http.Request {
                req, _ := http.NewRequest("GET", "/api/v1/finance/accounts", nil)
                req.Header.Set("X-Tenant-ID", "non-existent-tenant")
                return req
            },
            expectedStatus: http.StatusNotFound,
            shouldFail:     true,
        },
        {
            name: "Header manipulation",
            attack: func() *http.Request {
                req, _ := http.NewRequest("GET", "/api/v1/finance/accounts", nil)
                req.Header.Set("X-Tenant-ID", "valid-tenant")
                // Try to override with subdomain
                req.Host = "malicious.tenant.awoerp.com"
                return req
            },
            expectedStatus: http.StatusOK, // Header should take priority
            shouldFail:     false,
        },
    }
    
    for _, tt := range attackTests {
        t.Run(tt.name, func(t *testing.T) {
            req := tt.attack()
            resp, err := http.DefaultClient.Do(req)
            require.NoError(t, err)
            defer resp.Body.Close()
            
            assert.Equal(t, tt.expectedStatus, resp.StatusCode)
            
            if tt.shouldFail {
                // Ensure no sensitive data is returned
                body, _ := io.ReadAll(resp.Body)
                assert.NotContains(t, string(body), "account_id")
                assert.NotContains(t, string(body), "balance")
            }
        })
    }
}
```

---

## **DAILY TESTING ROUTINE**

### **Quick Validation Script:**
```bash
#!/bin/bash
# daily_test.sh - Run this every day

echo "🧪 Running daily validation tests..."

# 1. Unit tests
echo "Running unit tests..."
go test ./internal/platform/middleware/... -v

# 2. Build test
echo "Testing build..."
go build ./cmd/server || exit 1

# 3. Start server for integration tests
echo "Starting test server..."
./server &
SERVER_PID=$!
sleep 3

# 4. Basic functionality tests
echo "Testing basic functionality..."

# Public endpoint (should work without tenant)
curl -s http://localhost:8080/api/v1/health || echo "❌ Health check failed"

# Tenant endpoint (should work with tenant)
curl -s -H "X-Tenant-ID: test-tenant" http://localhost:8080/api/v1/finance/accounts || echo "❌ Tenant endpoint failed"

# Protected endpoint without tenant (should fail)
if curl -s http://localhost:8080/api/v1/finance/accounts | grep -q "error"; then
    echo "✅ Protection working"
else
    echo "❌ Tenant protection not working"
fi

# 5. Performance spot check
echo "Running performance check..."
wrk -t1 -c10 -d10s --header "X-Tenant-ID: test-tenant" \
    http://localhost:8080/api/v1/finance/accounts > perf_check.txt

# 6. Cleanup
kill $SERVER_PID
echo "🎉 Daily tests complete!"
```

---

## ✅ **TESTING COMPLETION CRITERIA**

### **Unit Testing (Day 8):**
- [ ] 90%+ code coverage achieved
- [ ] All edge cases tested
- [ ] Error scenarios validated
- [ ] Performance benchmarks recorded

### **Integration Testing (Day 9):**
- [ ] Tenant isolation verified with real data
- [ ] Database session variables working
- [ ] Public endpoints accessible
- [ ] Cross-tenant access prevented

### **Performance Testing (Day 9):**
- [ ] <5% latency increase from baseline
- [ ] Memory usage stable or improved
- [ ] Load testing passes with multiple tenants
- [ ] No performance regressions identified

### **Security Testing:**
- [ ] Attack scenarios fail appropriately
- [ ] No sensitive data leakage
- [ ] Input validation working
- [ ] Audit trail functioning

---

**This testing strategy ensures your migration maintains security, performance, and functionality. Follow it systematically for confidence in your implementation.**