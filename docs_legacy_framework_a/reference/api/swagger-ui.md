> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

# Interactive API Documentation

## AWO ERP API Explorer

The AWO ERP system provides a  REST API with 50+ endpoints across multiple domains. All APIs are auto-documented using OpenAPI 3.0 specifications generated directly from the Go codebase.

### Access the API Documentation

<div style="text-align: center; margin: 2rem 0;">
  <p><strong>Static API Documentation:</strong> Available in the source code at <code>internal/api/swagger/index.html</code></p>
  <small>
    <strong>Live Interactive Explorer:</strong> 
    <a href="http://localhost:8080/swagger-ui/" target="_blank">http://localhost:8080/swagger-ui/</a> 
    (requires server)
  </small>
</div>

### API Categories

The API is organized into the following major categories:

####  **Authentication & Authorization (ABAC)**
- **Base Path**: `/abac/*`
- **Endpoints**: 10 endpoints
- **Features**: 
  - Advanced policy evaluation engine
  - Bulk authorization checks
  - Attribute collection and validation
  - Audit trail and decision history
  - Policy discovery and explanation

```bash
# Example: Simple authorization check
curl -X POST "http://localhost:8080/abac/authorize" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-uuid",
    "resource_id": "resource-uuid", 
    "resource_type": "entity",
    "action": "read"
  }'
```

####  **User Management**
- **Base Path**: `/api/v1/users/*`
- **Endpoints**: 8 endpoints
- **Features**:
  - User lifecycle management
  - Profile and preferences
  - Role and permission assignment
  - Multi-tenant user isolation

####  **Organization & Tenant Management** 
- **Base Path**: `/api/v1/organizations/*`, `/api/v1/tenants/*`
- **Endpoints**: 12 endpoints
- **Features**:
  - Hierarchical organization structure
  - Tenant provisioning and configuration
  - Multi-tenant data isolation
  - Subscription and billing management

####  **Access Request Workflows**
- **Base Path**: `/api/v1/access-requests/*`
- **Endpoints**: 6 endpoints
- **Features**:
  - Automated approval workflows
  - Request lifecycle management
  - Analytics and reporting
  - Integration with ABAC policies

####  **User Analytics & Insights**
- **Base Path**: `/api/v1/analytics/*`
- **Endpoints**: 8 endpoints
- **Features**:
  - Behavioral analysis
  - Anomaly detection
  - Risk assessment
  - Usage insights and patterns

####  **Feature Flag Management**
- **Base Path**: `/api/v1/feature-flags/*`
- **Endpoints**: 6 endpoints
- **Features**:
  - Dynamic feature toggling
  - A/B testing framework
  - Rollout management
  - Usage analytics

### Authentication

All API endpoints require JWT authentication. Include your token in the Authorization header:

```bash
Authorization: Bearer <your-jwt-token>
```

### OpenAPI Specifications

Direct access to machine-readable API specifications:

- **OpenAPI 3.0**: `openapi3.json` (484KB with 50+ endpoints, available in source code at `internal/api/swagger/`)
- **OpenAPI 2.0**: `openapi.json` (legacy format, available in source code at `internal/api/swagger/`)

**Live Endpoints (when server running):**
- **OpenAPI 3.0**: [http://localhost:8080/swagger-ui/openapi3.json](http://localhost:8080/swagger-ui/openapi3.json)
- **OpenAPI 2.0**: [http://localhost:8080/swagger-ui/openapi.json](http://localhost:8080/swagger-ui/openapi.json)

### Response Formats

All APIs follow consistent response patterns:

**Success Response:**
```json
{
  "data": { ... },
  "meta": {
    "timestamp": "2025-01-01T12:00:00Z",
    "request_id": "req-uuid"
  }
}
```

**Error Response:**
```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable message",
    "details": { ... }
  },
  "meta": {
    "timestamp": "2025-01-01T12:00:00Z", 
    "request_id": "req-uuid"
  }
}
```

### Rate Limits

API rate limits are enforced per user/tenant:

- **Standard APIs**: 1000 requests/hour
- **Analytics APIs**: 100 requests/hour  
- **Bulk Operations**: 10 requests/hour
- **Authentication**: 60 requests/hour

### SDK and Integration

- **Go SDK**: Auto-generated from OpenAPI specs
- **Postman Collection**: Available in the source code
- **Test Scripts**: Available in the source code

## Getting Started

1. **Obtain JWT Token**: Use the `/api/v1/auth/login` endpoint
2. **Browse Static Docs**: Open the Static API Documentation
3. **Explore Live APIs**: Visit [Interactive Explorer](http://localhost:8080/swagger-ui/) (when server running)
4. **Review Examples**: Check the practical examples in each API section

For detailed implementation guides, see the [API Integration Tutorial](tutorials/api-integration.md).