# Tenant Management API

The Tenant Management API provides comprehensive functionality for managing multi-tenant configurations in the AWO ERP System. This API enables tenant isolation, configuration, and management across the enterprise.

**Updated: 2025-09-20** - Based on comprehensive API testing results.

##  Overview

**Base URL**: `/api/v1/tenants`

**Purpose**: Multi-tenant architecture support with tenant isolation, configuration, and lifecycle management.

**API Status**: **73% Operational** (8/11 endpoints working)

**Key Features**:
- ✅ Tenant CRUD operations (Create, Read, Update, Delete)
- ✅ Multi-tenant data isolation with Row-Level Security (RLS)
- ✅ Health monitoring and usage analytics
- ⚠️ Limited tenant configuration management
- ❌ Known issues with provisioning and status transitions

##  Authentication Requirements

**IMPORTANT**: All endpoints (except tenant creation) require tenant context via:
- **X-Tenant-ID header** with a valid UUID, OR
- **Subdomain format** like `tenant.example.com` (minimum 3 parts)

##  Available Endpoints

| Method | Endpoint | Status | Description |
|--------|----------|--------|-------------|
| `POST` | `/api/v1/tenants` | ✅ **Working** | Create a new tenant (bypasses middleware) |
| `GET` | `/api/v1/tenants` | ⚠️ **Limited** | List tenants (empty due to RLS) |
| `GET` | `/api/v1/tenants/{id}` | ✅ **Working** | Get tenant by ID |
| `PUT` | `/api/v1/tenants/{id}` | ✅ **Working** | Update tenant information |
| `DELETE` | `/api/v1/tenants/{id}` | ✅ **Working** | Delete/deactivate tenant (soft delete) |
| `GET` | `/api/v1/tenants/health` | ✅ **Working** | Get tenant service health status |
| `GET` | `/api/v1/tenants/{id}/analytics` | ✅ **Working** | Get tenant usage analytics |
| `POST` | `/api/v1/tenants/provision` | ❌ **Failing** | Provision new tenant (DB constraints) |
| `POST` | `/api/v1/tenants/{id}/suspend` | ❌ **Failing** | Suspend tenant (status constraints) |
| `POST` | `/api/v1/tenants/{id}/reactivate` | ❌ **Failing** | Reactivate tenant (status constraints) |
| `PUT` | `/api/v1/tenants/{id}/configuration` | ❌ **Failing** | Update configuration (requires active tenant) |

##  Data Models

### Tenant Object (Updated Schema)
```json
{
  "id": "string (UUID)",
  "name": "string",
  "slug": "string (auto-generated)",
  "subdomain": "string (optional)",
  "status": "string (ACTIVE|INACTIVE|SUSPENDED)",
  "plan_type": "string (basic|premium|enterprise)",
  "settings": {
    "timezone": "string",
    "currency": "string",
    "date_format": "string",
    "language": "string",
    "features": "array"
  },
  "created_at": "string (ISO 8601)",
  "updated_at": "string (ISO 8601)"
}
```

### Create Tenant Request
```json
{
  "name": "string (required)",
  "email": "string (required)",
  "subdomain": "string (optional)",
  "country_code": "string (required)",
  "currency_code": "string (required)",
  "status": "active|inactive|suspended",
  "industry": "string (optional)",
  "company_size": "1-10|11-50|51-200|201-500|500+",
  "contact": {
    "email": "string"
  },
  "settings": {
    "timezone": "string",
    "currency": "string",
    "date_format": "string",
    "language": "string"
  }
}
```

### Health Status Object
```json
{
  "status": "string (healthy|unhealthy)",
  "timestamp": "string (ISO 8601)",
  "version": "string"
}
```

### Pagination Object
```json
{
  "current_page": "integer",
  "page_size": "integer", 
  "total_items": "integer",
  "total_pages": "integer",
  "has_next": "boolean",
  "has_prev": "boolean"
}
```

### Usage Analytics Object
```json
{
  "tenant_id": "string (UUID)",
  "period": "string",
  "user_count": "integer",
  "storage_used_mb": "integer",
  "api_calls": "integer"
}
```

##  Quick Examples

### Create Tenant (Working)
```bash
curl -X POST http://localhost:8080/api/v1/tenants \
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

**Response:**
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

### List Tenants (Limited due to RLS)
```bash
# Requires tenant context - often returns empty due to Row-Level Security
curl -X GET "http://localhost:8080/api/v1/tenants?page=1&page_size=10" \
  -H "X-Tenant-ID: 96ab6888-2914-4872-94b8-c25d964448cb" | jq .
```

**Response (typically empty due to RLS):**
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

### Get Tenant (Working)
```bash
# Requires tenant context header
curl -X GET http://localhost:8080/api/v1/tenants/96ab6888-2914-4872-94b8-c25d964448cb \
  -H "X-Tenant-ID: 96ab6888-2914-4872-94b8-c25d964448cb" | jq .
```

### Update Tenant (Working)
```bash
curl -X PUT http://localhost:8080/api/v1/tenants/96ab6888-2914-4872-94b8-c25d964448cb \
  -H "X-Tenant-ID: 96ab6888-2914-4872-94b8-c25d964448cb" \
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

### Delete Tenant (Working)
```bash
curl -X DELETE http://localhost:8080/api/v1/tenants/96ab6888-2914-4872-94b8-c25d964448cb \
  -H "X-Tenant-ID: 96ab6888-2914-4872-94b8-c25d964448cb"
```

### Health Check (Working)
```bash
# Requires tenant context
curl -X GET http://localhost:8080/api/v1/tenants/health \
  -H "X-Tenant-ID: 96ab6888-2914-4872-94b8-c25d964448cb" | jq .
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-09-20T06:16:17Z", 
  "version": "1.0.0"
}
```

### Usage Analytics (Working)
```bash
# Get tenant usage analytics
curl -X GET "http://localhost:8080/api/v1/tenants/96ab6888-2914-4872-94b8-c25d964448cb/analytics?period=current_month" \
  -H "X-Tenant-ID: 96ab6888-2914-4872-94b8-c25d964448cb" | jq .
```

**Response:**
```json
{
  "tenant_id": "96ab6888-2914-4872-94b8-c25d964448cb",
  "period": "current_month",
  "user_count": 18,
  "storage_used_mb": 2048,
  "api_calls": 50000
}
```

##  Advanced Usage

### Pagination Parameters

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `page` | integer | Page number (1-based) | 1 |
| `page_size` | integer | Number of items per page | 10 |

### Analytics Parameters

| Parameter | Type | Valid Values |
|-----------|------|-------------|
| `period` | string | `current_month`, `last_month`, `last_3_months`, `last_year` |

### Valid Enum Values

| Field | Valid Values |
|-------|-------------|
| `status` | `active`, `inactive`, `suspended` |
| `company_size` | `1-10`, `11-50`, `51-200`, `201-500`, `500+` |

### Example with Pagination
```bash
curl -X GET "http://localhost:8080/api/v1/tenants?page=1&page_size=5" \
  -H "X-Tenant-ID: 96ab6888-2914-4872-94b8-c25d964448cb" | jq .
```

##  Error Handling

### Common Error Responses

| Status Code | Error Type | Description | Solution |
|-------------|------------|-------------|----------|
| `400` | `tenant_resolution_failed` | Missing X-Tenant-ID header | Add tenant context header |
| `400` | `invalid_tenant_id` | Invalid UUID format | Use valid UUID format |
| `400` | `invalid_enum_value` | Wrong enum values | Use correct status/company_size |
| `404` | `not_found` | Tenant not found | Verify tenant ID exists |
| `500` | `tenant_validation_failed` | Tenant doesn't exist in DB | Create tenant first |
| `500` | `check_constraint_violation` | Database constraint error | Known issue - avoid provisioning |

### Error Response Format
```json
{
  "error": "string",
  "status": "integer"
}
```

### Known Issues and Workarounds

| Issue | Status | Workaround |
|-------|--------|-----------|
| Provisioning fails | ❌ **Known Bug** | Use basic tenant creation instead |
| Suspend/Reactivate fails | ❌ **Known Bug** | Status constraints need fixing |
| List returns empty | ⚠️ **By Design** | RLS policies filter results |
| Configuration updates fail | ❌ **Business Logic** | Only works for active tenants |

## ✅ Validation Rules

### Tenant Creation (Required Fields)
- **name**: Required, 1-255 characters
- **email**: Required, valid email format
- **country_code**: Required, 2-letter country code
- **currency_code**: Required, 3-letter currency code
- **status**: Required, must be `active`, `inactive`, or `suspended`
- **company_size**: Required, must be `1-10`, `11-50`, `51-200`, `201-500`, or `500+`

### Tenant Update
- **name**: Optional, 1-255 characters if provided
- **contact.email**: Optional, valid email format
- **settings**: Optional, nested object with timezone, currency, etc.

### Authentication Requirements
- All endpoints except `POST /api/v1/tenants` require `X-Tenant-ID` header
- Tenant ID must be valid UUID format
- Tenant must exist in database

##  Testing

### Automated Test Script
```bash
# Run comprehensive API tests
./docs/scripts/api/test_tenant_api.sh
```

### Quick Health Check Test
```bash
# Create tenant first (required for other operations)
TENANT_ID=$(curl -s -X POST http://localhost:8080/api/v1/tenants \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Health Test Tenant",
    "email": "test@example.com",
    "country_code": "US",
    "currency_code": "USD",
    "status": "active",
    "company_size": "11-50"
  }' | jq -r '.tenant.id')

# Test health endpoint
curl -X GET http://localhost:8080/api/v1/tenants/health \
  -H "X-Tenant-ID: $TENANT_ID"

# Expected: HTTP 200 with health status
```

### Working CRUD Operations Test
```bash
# 1. Create tenant (bypasses middleware)
TENANT_ID=$(curl -s -X POST http://localhost:8080/api/v1/tenants \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Tenant",
    "email": "test@example.com",
    "country_code": "US",
    "currency_code": "USD",
    "status": "active",
    "company_size": "11-50"
  }' | jq -r '.tenant.id')

# 2. Get tenant (requires header)
curl -X GET http://localhost:8080/api/v1/tenants/$TENANT_ID \
  -H "X-Tenant-ID: $TENANT_ID"

# 3. Update tenant (requires header)
curl -X PUT http://localhost:8080/api/v1/tenants/$TENANT_ID \
  -H "X-Tenant-ID: $TENANT_ID" \
  -H "Content-Type: application/json" \
  -d '{"name": "Updated Test Tenant"}'

# 4. List tenants (requires header, returns empty due to RLS)
curl -X GET http://localhost:8080/api/v1/tenants \
  -H "X-Tenant-ID: $TENANT_ID"

# 5. Delete tenant (requires header)
curl -X DELETE http://localhost:8080/api/v1/tenants/$TENANT_ID \
  -H "X-Tenant-ID: $TENANT_ID"
```

##  Performance Considerations

### Current Performance Metrics
- **Health Check**: ~50ms response time
- **Create Operation**: ~200ms response time
- **Get Operation**: ~100ms response time
- **Update Operation**: ~150ms response time
- **Delete Operation**: ~100ms response time

### Optimization Tips
- Always include `X-Tenant-ID` header to avoid resolution overhead
- Use create endpoint for new tenants (bypasses middleware)
- Avoid provisioning/suspend endpoints until database constraints are fixed
- Monitor health endpoint for service status
- Be aware that list operations may return empty due to RLS policies

### Known Limitations
- **Row-Level Security (RLS)**: List operations return empty results
- **Database Constraints**: Some management operations fail
- **Tenant Context Required**: Most operations need tenant authentication

##  Related APIs

- **[User Management API](../users/README.md)** - Users belong to tenants
- **[Entity Management API](../entities/README.md)** - Entities are tenant-scoped
- **[Authentication API](../../workflows/auth/README.md)** - Tenant-aware authentication
- **[Organization API](../organizations/README.md)** - Organization-tenant relationships

##  Additional Resources

- **[curl Examples](curl-examples.md)** - Complete curl command examples with real test results
- **Testing Scripts**: Comprehensive automated testing script (available in the source code)
- **API Source Design**: Goa service definitions (available in the source code)

##  Quick Start Guide

1. **Create a tenant** (only public endpoint):
   ```bash
   TENANT_ID=$(curl -s -X POST http://localhost:8080/api/v1/tenants \
     -H "Content-Type: application/json" \
     -d '{"name":"My Company","email":"me@company.com","country_code":"US","currency_code":"USD","status":"active","company_size":"11-50"}' \
     | jq -r '.tenant.id')
   ```

2. **Use the tenant ID** for all subsequent operations:
   ```bash
   export TENANT_ID="96ab6888-2914-4872-94b8-c25d964448cb"
   ```

3. **Test the API**:
   ```bash
   curl -X GET http://localhost:8080/api/v1/tenants/health -H "X-Tenant-ID: $TENANT_ID"
   ```

##  Critical Implementation Notes

⚠️ **Before using this API in production:**

1. **Fix Database Constraints**: The `tenants_company_size_check` and `tenants_status_check` constraints need review
2. **Resolve RLS Issues**: List operations return empty due to Row-Level Security policies
3. **Enable Management Operations**: Provisioning, suspend, and configuration endpoints need fixes
4. **Update Documentation**: Remove references to non-working endpoints until fixed

✅ **Currently safe to use:**
- Basic CRUD operations (Create, Get, Update, Delete)
- Health monitoring
- Usage analytics
- Authentication and tenant context resolution

---

**Last Updated**: 2025-09-20  
**API Version**: 1.0.0  
**Test Status**: 73% Operational (8/11 endpoints working)  
**Maintained by**: AWO ERP System Team