> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

# Tenant Management API

The Tenant Management API provides  functionality for managing multi-tenant configurations in the AWO ERP System. This API enables tenant isolation, configuration, and management across the enterprise.

##  Overview

**Base URL**: `/api/v1/tenants`

**Purpose**: Multi-tenant architecture support with tenant isolation, configuration, and lifecycle management.

**Key Features**:
- Tenant CRUD operations
- Multi-tenant data isolation  
- Tenant configuration management
- Health monitoring
- Tenant status management

##  Available Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/tenants` | Create a new tenant |
| `GET` | `/api/v1/tenants` | List all tenants with pagination |
| `GET` | `/api/v1/tenants/{id}` | Get tenant by ID |
| `PUT` | `/api/v1/tenants/{id}` | Update tenant information |
| `DELETE` | `/api/v1/tenants/{id}` | Delete/deactivate tenant |
| `GET` | `/api/v1/tenants/health` | Get tenant service health status |

##  Data Models

### Tenant Object
```json
{
  "id": "string (UUID)",
  "name": "string",
  "status": "string (ACTIVE|INACTIVE|SUSPENDED)",
  "created_at": "string (ISO 8601)",
  "updated_at": "string (ISO 8601)"
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

##  Quick Examples

### Create Tenant
```bash
curl -X POST http://localhost:8080/api/v1/tenants \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Corporation"
  }' | jq .
```

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Acme Corporation", 
  "status": "active",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### List Tenants
```bash
curl -X GET "http://localhost:8080/api/v1/tenants?limit=10&offset=0" | jq .
```

**Response:**
```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Acme Corporation",
      "status": "active", 
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "current_page": 1,
    "page_size": 20,
    "total_items": 1,
    "total_pages": 1,
    "has_next": false,
    "has_prev": false
  }
}
```

### Get Tenant
```bash
curl -X GET http://localhost:8080/api/v1/tenants/550e8400-e29b-41d4-a716-446655440000 | jq .
```

### Update Tenant
```bash
curl -X PUT http://localhost:8080/api/v1/tenants/550e8400-e29b-41d4-a716-446655440000 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Corporation Updated"
  }' | jq .
```

### Delete Tenant
```bash
curl -X DELETE http://localhost:8080/api/v1/tenants/550e8400-e29b-41d4-a716-446655440000
```

### Health Check
```bash
curl -X GET http://localhost:8080/api/v1/tenants/health | jq .
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T00:00:00Z", 
  "version": "1.0.0"
}
```

##  Advanced Usage

### Pagination Parameters

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `limit` | integer | Number of items per page | 20 |
| `offset` | integer | Number of items to skip | 0 |

### Filter Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `status` | string | Filter by tenant status |
| `name` | string | Filter by tenant name (partial match) |

### Example with Filters
```bash
curl -X GET "http://localhost:8080/api/v1/tenants?status=ACTIVE&limit=5" | jq .
```

##  Error Handling

### Common Error Responses

| Status Code | Error Type | Description |
|-------------|------------|-------------|
| `400` | `bad_request` | Invalid request payload or parameters |
| `404` | `not_found` | Tenant not found |
| `409` | `conflict` | Tenant name or subdomain already exists |
| `500` | `internal_error` | Internal server error |

### Error Response Format
```json
{
  "error": "string",
  "message": "string",
  "details": "string (optional)"
}
```

## ✅ Validation Rules

### Tenant Creation
- **name**: Required, 1-255 characters, unique
- **status**: Optional, defaults to "ACTIVE"

### Tenant Update  
- **name**: Optional, 1-255 characters, unique if provided
- **status**: Optional, valid status enum

##  Testing

### Health Check Test
```bash
# Test tenant service health
curl -X GET http://localhost:8080/api/v1/tenants/health

# Expected: HTTP 200 with health status
```

### CRUD Operations Test
```bash
# 1. Create tenant
TENANT_ID=$(curl -s -X POST http://localhost:8080/api/v1/tenants \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Tenant"}' | jq -r '.id')

# 2. Get tenant
curl -X GET http://localhost:8080/api/v1/tenants/$TENANT_ID

# 3. Update tenant
curl -X PUT http://localhost:8080/api/v1/tenants/$TENANT_ID \
  -H "Content-Type: application/json" \
  -d '{"name": "Updated Test Tenant"}'

# 4. List tenants
curl -X GET http://localhost:8080/api/v1/tenants

# 5. Delete tenant
curl -X DELETE http://localhost:8080/api/v1/tenants/$TENANT_ID
```

##  Performance Considerations

### Caching Strategy
- Tenant data is cached for 30 minutes
- Health status is cached for 5 minutes
- List operations use database pagination

### Rate Limiting
- Standard rate limits apply (100 requests/minute per IP)
- Bulk operations may have lower limits

### Optimization Tips
- Use pagination for large tenant lists
- Cache tenant lookups when possible
- Monitor health endpoint for service status

##  Related APIs

- **[User Management API](../user/README.md)** - Users belong to tenants
- **Entity Management API** - Entities are tenant-scoped
- **[Authentication API](../../api-source/workflows/auth/README.md)** - Tenant-aware authentication

##  Additional Resources

- **[curl Examples](curl-examples.md)** - Complete curl command examples
- **Testing Scripts** - Automated testing script
- **Postman Collection** - Postman requests

---

**Last Updated**: 2025-07-19  
**API Version**: 1.0.0  
**Maintained by**: AWO ERP System Team