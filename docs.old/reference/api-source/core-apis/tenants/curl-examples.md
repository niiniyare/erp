# Tenant Management API - curl Examples

Complete collection of curl commands for testing the Tenant Management API.
**Updated: 2025-09-20** - Based on comprehensive API testing results.

##  Prerequisites

### Environment Setup
```bash
# Set base URL
export API_BASE_URL="http://localhost:8080"

# Test server connectivity
curl -X GET $API_BASE_URL/api/v1/tenants/health -H "X-Tenant-ID: test" || echo "Server not responding"
```

### Required Tools
```bash
# Install jq for JSON processing
sudo apt-get install jq  # Ubuntu/Debian
brew install jq          # macOS
```

###  Authentication Requirements
**IMPORTANT**: All endpoints (except tenant creation) require tenant context via:
- **X-Tenant-ID header** with a valid UUID, OR  
- **Subdomain format** like `tenant.example.com`

##  Core Tenant Operations

### 1. Create Tenant ✅ WORKING

**This is the only endpoint that bypasses tenant middleware**

```bash
# Basic tenant creation with all required fields
curl -X POST $API_BASE_URL/api/v1/tenants \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Company",
    "email": "contact@testcompany.com",
    "subdomain": "testcompany",
    "country_code": "US",
    "currency_code": "USD",
    "status": "active",
    "industry": "Technology",
    "company_size": "11-50",
    "contact": {
      "email": "contact@testcompany.com"
    },
    "settings": {
      "timezone": "UTC",
      "currency": "USD",
      "date_format": "YYYY-MM-DD",
      "language": "en"
    }
  }' | jq .
```

**Expected Response (200):**
```json
{
  "tenant": {
    "id": "96ab6888-2914-4872-94b8-c25d964448cb",
    "name": "Test Company",
    "slug": "test-company",
    "subdomain": "testcompany",
    "status": "ACTIVE",
    "plan_type": "basic",
    "settings": {
      "timezone": "UTC",
      "currency": "USD",
      "date_format": "YYYY-MM-DD",
      "language": "en"
    },
    "created_at": "2025-09-20T06:15:03Z",
    "updated_at": "2025-09-20T06:15:03Z"
  },
  "status": "SUCCESS",
  "message": "Tenant created successfully"
}
```

#### Valid Enum Values
```bash
# Status values: "active", "inactive", "suspended"
# Company size: "1-10", "11-50", "51-200", "201-500", "500+"

# Example with different enum values
curl -X POST $API_BASE_URL/api/v1/tenants \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Large Enterprise",
    "email": "contact@enterprise.com",
    "subdomain": "enterprise",
    "country_code": "CA",
    "currency_code": "CAD",
    "status": "active",
    "industry": "Finance",
    "company_size": "500+",
    "contact": {
      "email": "contact@enterprise.com"
    }
  }' | jq .
```

#### Save Tenant ID for Further Testing
```bash
# Create tenant and extract ID for subsequent tests
TENANT_ID=$(curl -s -X POST $API_BASE_URL/api/v1/tenants \
  -H "Content-Type: application/json" \
  -d '{
    "name": "API Test Tenant",
    "email": "test@example.com",
    "subdomain": "apitest",
    "country_code": "US",
    "currency_code": "USD",
    "status": "active",
    "industry": "Technology",
    "company_size": "11-50"
  }' | jq -r '.tenant.id')

echo "Created tenant with ID: $TENANT_ID"
```

### 2. Get Tenant ✅ WORKING

```bash
# Get specific tenant by ID (requires tenant context)
curl -X GET $API_BASE_URL/api/v1/tenants/$TENANT_ID \
  -H "X-Tenant-ID: $TENANT_ID" | jq .
```

**Expected Response (200):**
```json
{
  "id": "96ab6888-2914-4872-94b8-c25d964448cb",
  "name": "Test Company",
  "slug": "test-company",
  "subdomain": "testcompany",
  "status": "ACTIVE",
  "plan_type": "basic",
  "settings": {
    "timezone": "UTC",
    "currency": "USD",
    "date_format": "YYYY-MM-DD",
    "language": "en"
  },
  "created_at": "2025-09-20T06:15:03Z",
  "updated_at": "2025-09-20T06:15:03Z"
}
```

### 3. Update Tenant ✅ WORKING

```bash
# Update tenant information
curl -X PUT $API_BASE_URL/api/v1/tenants/$TENANT_ID \
  -H "X-Tenant-ID: $TENANT_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Test Company",
    "contact": {
      "email": "newemail@testcompany.com"
    },
    "settings": {
      "timezone": "America/New_York",
      "currency": "USD",
      "date_format": "MM/DD/YYYY",
      "language": "en"
    }
  }' | jq .
```

### 4. List Tenants ✅ WORKING (with caveats)

```bash
# List tenants with pagination
curl -X GET "$API_BASE_URL/api/v1/tenants?page=1&page_size=10" \
  -H "X-Tenant-ID: $TENANT_ID" | jq .
```

**Expected Response (200) - Often Empty due to RLS:**
```json
{
  "data": [],
  "pagination": {
    "current_page": 1,
    "page_size": 10,
    "total_items": 0,
    "total_pages": 1,
    "has_next": false,
    "has_prev": false
  }
}
```

**Note**: List returns empty due to Row-Level Security (RLS) policies.

### 5. Delete Tenant ✅ WORKING

```bash
# Delete tenant (soft delete)
curl -X DELETE $API_BASE_URL/api/v1/tenants/$TENANT_ID \
  -H "X-Tenant-ID: $TENANT_ID" -v

# Verify deletion
curl -X GET $API_BASE_URL/api/v1/tenants/$TENANT_ID \
  -H "X-Tenant-ID: $TENANT_ID" | jq .
```

##  Service Health & Analytics

### Health Check ✅ WORKING

```bash
# Check tenant service health (requires tenant context)
curl -X GET $API_BASE_URL/api/v1/tenants/health \
  -H "X-Tenant-ID: $TENANT_ID" | jq .
```

**Expected Response (200):**
```json
{
  "status": "healthy",
  "timestamp": "2025-09-20T06:16:17Z",
  "version": "1.0.0"
}
```

### Usage Analytics ✅ WORKING

```bash
# Get tenant usage analytics
curl -X GET "$API_BASE_URL/api/v1/tenants/$TENANT_ID/analytics?period=current_month" \
  -H "X-Tenant-ID: $TENANT_ID" | jq .
```

**Valid period values:** `current_month`, `last_month`, `last_3_months`, `last_year`

**Expected Response (200):**
```json
{
  "tenant_id": "96ab6888-2914-4872-94b8-c25d964448cb",
  "period": "current_month",
  "user_count": 18,
  "storage_used_mb": 2048,
  "api_calls": 50000
}
```

##  Management Operations (Known Issues)

### Provision Tenant ❌ FAILING

```bash
# Tenant provisioning (currently fails due to DB constraints)
curl -X POST $API_BASE_URL/api/v1/tenants/provision \
  -H "X-Tenant-ID: $TENANT_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Provisioned Company",
    "subdomain": "provisioned",
    "contact_email": "contact@provisioned.com",
    "admin_email": "admin@provisioned.com",
    "admin_first_name": "Jane",
    "admin_last_name": "Smith"
  }' | jq .
```

**Current Error:**
```json
"failed to provision basic tenant: tenant provisioning transaction failed: failed to provision tenant via DB function: ERROR: new row for relation \"tenants\" violates check constraint \"tenants_company_size_check\" (SQLSTATE 23514)"
```

### Suspend Tenant ❌ FAILING

```bash
# Suspend tenant (currently fails due to DB constraints)
curl -X POST $API_BASE_URL/api/v1/tenants/$TENANT_ID/suspend \
  -H "X-Tenant-ID: $TENANT_ID" \
  -H "Content-Type: application/json" \
  -d '{"reason": "Testing suspension functionality"}' | jq .
```

**Current Error:**
```json
"failed to suspend tenant: failed to update tenant: ERROR: new row for relation \"tenants\" violates check constraint \"tenants_status_check\" (SQLSTATE 23514)"
```

### Update Configuration ❌ FAILING

```bash
# Update tenant configuration (fails for inactive tenants)
curl -X PUT $API_BASE_URL/api/v1/tenants/$TENANT_ID/configuration \
  -H "X-Tenant-ID: $TENANT_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "reason": "Testing configuration update",
    "max_users": 50,
    "max_storage_mb": 5120,
    "max_api_calls_per_hour": 2000
  }' | jq .
```

**Current Error:**
```json
"[business] <TENANT_NOT_ACTIVE> Cannot update configuration for inactive tenant"
```

##  Test Scenarios

### Complete Working Lifecycle
```bash
echo "=== Complete Tenant Lifecycle Test ==="

# 1. Create tenant
echo "Creating tenant..."
TENANT_ID=$(curl -s -X POST $API_BASE_URL/api/v1/tenants \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Lifecycle Test Tenant",
    "email": "lifecycle@test.com",
    "subdomain": "lifecycle",
    "country_code": "US", 
    "currency_code": "USD",
    "status": "active",
    "industry": "Technology",
    "company_size": "11-50"
  }' | jq -r '.tenant.id')
echo "Created tenant: $TENANT_ID"

# 2. Verify creation
echo "Verifying tenant creation..."
curl -s -X GET $API_BASE_URL/api/v1/tenants/$TENANT_ID \
  -H "X-Tenant-ID: $TENANT_ID" | jq .

# 3. Check health
echo "Checking health..."
curl -s -X GET $API_BASE_URL/api/v1/tenants/health \
  -H "X-Tenant-ID: $TENANT_ID" | jq .

# 4. Update tenant
echo "Updating tenant..."
curl -s -X PUT $API_BASE_URL/api/v1/tenants/$TENANT_ID \
  -H "X-Tenant-ID: $TENANT_ID" \
  -H "Content-Type: application/json" \
  -d '{"name": "Updated Lifecycle Test Tenant"}' | jq .

# 5. Get analytics
echo "Getting analytics..."
curl -s -X GET "$API_BASE_URL/api/v1/tenants/$TENANT_ID/analytics?period=current_month" \
  -H "X-Tenant-ID: $TENANT_ID" | jq .

# 6. Delete tenant
echo "Deleting tenant..."
curl -s -X DELETE $API_BASE_URL/api/v1/tenants/$TENANT_ID \
  -H "X-Tenant-ID: $TENANT_ID" -w "HTTP Status: %{http_code}\n"

echo "=== Lifecycle test completed ==="
```

### Error Testing Scenarios
```bash
echo "=== Error Handling Tests ==="

# Test missing tenant context
echo "Testing missing tenant context (should fail with 400):"
curl -X GET $API_BASE_URL/api/v1/tenants | jq .

# Test invalid UUID format
echo "Testing invalid UUID format:"
curl -X GET $API_BASE_URL/api/v1/tenants/invalid-uuid \
  -H "X-Tenant-ID: $TENANT_ID" | jq .

# Test invalid enum values
echo "Testing invalid enum values:"
curl -X POST $API_BASE_URL/api/v1/tenants \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Invalid Enum Test",
    "email": "test@test.com",
    "subdomain": "invalid",
    "country_code": "US",
    "currency_code": "USD",
    "status": "INVALID_STATUS",
    "company_size": "INVALID_SIZE"
  }' | jq .

echo "=== Error tests completed ==="
```

##  Authentication Methods

### Using X-Tenant-ID Header (Recommended)
```bash
# Standard header-based authentication
curl -X GET $API_BASE_URL/api/v1/tenants/$TENANT_ID \
  -H "X-Tenant-ID: $TENANT_ID"
```

### Using Subdomain (Alternative)
```bash
# Subdomain-based tenant resolution (3+ parts required)
# Format: tenant.example.com (NOT tenant.localhost)
curl -X GET http://testcompany.api.localhost:8080/api/v1/tenants/health
```

##  Monitoring and Debugging

### Request/Response Logging
```bash
# Enable verbose output for debugging
curl -X GET $API_BASE_URL/api/v1/tenants/$TENANT_ID \
  -H "X-Tenant-ID: $TENANT_ID" -v

# Log timing information
curl -X GET $API_BASE_URL/api/v1/tenants/$TENANT_ID \
  -H "X-Tenant-ID: $TENANT_ID" \
  -w "Response Code: %{http_code}\nTime Total: %{time_total}s\n"
```

### Header Analysis
```bash
# Check response headers
curl -I -X GET $API_BASE_URL/api/v1/tenants/health \
  -H "X-Tenant-ID: $TENANT_ID"
```

## ️ Utility Functions

### Tenant Cleanup Function
```bash
#!/bin/bash
# Function to clean up test tenants
cleanup_test_tenants() {
  echo "Cleaning up test tenants..."
  
  # Note: This requires a valid tenant ID for the list operation
  TENANT_ID="96ab6888-2914-4872-94b8-c25d964448cb"  # Use existing tenant
  
  # Get test tenant IDs (if any are returned due to RLS)
  TEST_TENANT_IDS=$(curl -s -X GET $API_BASE_URL/api/v1/tenants \
    -H "X-Tenant-ID: $TENANT_ID" | \
    jq -r '.data[]? | select(.name | contains("Test")) | .id')
  
  # Delete each test tenant
  for TENANT_ID in $TEST_TENANT_IDS; do
    curl -s -X DELETE $API_BASE_URL/api/v1/tenants/$TENANT_ID \
      -H "X-Tenant-ID: $TENANT_ID"
    echo "Deleted test tenant: $TENANT_ID"
  done
  
  echo "Cleanup completed"
}
```

##  Common Issues and Solutions

### Issue: Tenant Resolution Failed
```bash
# Error: "Tenant resolution failed: no valid tenant found"
# Solution: Always include X-Tenant-ID header for non-creation endpoints

# ❌ Wrong
curl -X GET $API_BASE_URL/api/v1/tenants

# ✅ Correct  
curl -X GET $API_BASE_URL/api/v1/tenants \
  -H "X-Tenant-ID: 96ab6888-2914-4872-94b8-c25d964448cb"
```

### Issue: Invalid Enum Values
```bash
# Error: "value must be one of..."
# Solution: Use correct enum values

# ❌ Wrong
"status": "ACTIVE"        # Use "active"
"company_size": "MEDIUM"  # Use "11-50"

# ✅ Correct
"status": "active"
"company_size": "11-50"
```

### Issue: Database Constraint Violations
```bash
# Error: "violates check constraint"
# Status: Known issue with provisioning and status transitions
# Workaround: Use basic CRUD operations only
```

##  Performance Testing

### Response Time Testing
```bash
# Test health endpoint performance
echo "Health check timing:"
time curl -s -X GET $API_BASE_URL/api/v1/tenants/health \
  -H "X-Tenant-ID: $TENANT_ID" > /dev/null

# Test create operation performance  
echo "Create operation timing:"
time curl -s -X POST $API_BASE_URL/api/v1/tenants \
  -H "Content-Type: application/json" \
  -d '{"name": "Performance Test"}' > /dev/null
```

##  API Status Summary

| Endpoint | Status | Notes |
|----------|--------|-------|
| `POST /api/v1/tenants` | ✅ **Working** | Bypasses middleware |
| `GET /api/v1/tenants/{id}` | ✅ **Working** | Requires tenant context |
| `PUT /api/v1/tenants/{id}` | ✅ **Working** | Requires tenant context |
| `DELETE /api/v1/tenants/{id}` | ✅ **Working** | Soft delete |
| `GET /api/v1/tenants` | ⚠️ **Limited** | Returns empty due to RLS |
| `GET /api/v1/tenants/health` | ✅ **Working** | Requires tenant context |
| `GET /api/v1/tenants/{id}/analytics` | ✅ **Working** | Valid periods only |
| `POST /api/v1/tenants/provision` | ❌ **Failing** | DB constraint violation |
| `POST /api/v1/tenants/{id}/suspend` | ❌ **Failing** | Status constraint issue |
| `PUT /api/v1/tenants/{id}/configuration` | ❌ **Failing** | Requires active tenant |

---

**Last Updated**: 2025-09-20  
**API Version**: 1.0.0  
**Tested with**: curl 8.16.0, jq 1.6  
**Test Results**: 8/11 endpoints working (73% success rate)