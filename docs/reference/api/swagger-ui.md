# Interactive API Documentation

## AWO ERP API Explorer

The AWO ERP system provides a comprehensive REST API with 50+ endpoints across multiple domains. All APIs are auto-documented using OpenAPI 3.0 specifications generated directly from the Go codebase.

### Access the Interactive API Documentation

<div style="text-align: center; margin: 2rem 0;">
  <a href="http://localhost:8080/swagger/" target="_blank" class="md-button md-button--primary">
    🚀 Open Interactive API Explorer
  </a>
</div>

### API Categories

The API is organized into the following major categories:

#### 🔐 **Authentication & Authorization (ABAC)**
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

#### 👥 **User Management**
- **Base Path**: `/api/v1/users/*`
- **Endpoints**: 8 endpoints
- **Features**:
  - User lifecycle management
  - Profile and preferences
  - Role and permission assignment
  - Multi-tenant user isolation

#### 🏢 **Organization & Tenant Management** 
- **Base Path**: `/api/v1/organizations/*`, `/api/v1/tenants/*`
- **Endpoints**: 12 endpoints
- **Features**:
  - Hierarchical organization structure
  - Tenant provisioning and configuration
  - Multi-tenant data isolation
  - Subscription and billing management

#### 📋 **Access Request Workflows**
- **Base Path**: `/api/v1/access-requests/*`
- **Endpoints**: 6 endpoints
- **Features**:
  - Automated approval workflows
  - Request lifecycle management
  - Analytics and reporting
  - Integration with ABAC policies

#### 📊 **User Analytics & Insights**
- **Base Path**: `/api/v1/analytics/*`
- **Endpoints**: 8 endpoints
- **Features**:
  - Behavioral analysis
  - Anomaly detection
  - Risk assessment
  - Usage insights and patterns

#### 🚩 **Feature Flag Management**
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

- **OpenAPI 3.0**: [http://localhost:8080/swagger/openapi3.json](http://localhost:8080/swagger/openapi3.json)
- **OpenAPI 2.0**: [http://localhost:8080/swagger/openapi.json](http://localhost:8080/swagger/openapi.json)

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
- **Postman Collection**: [Download Collection](../../api-source/utilities/postman/)
- **Test Scripts**: [API Testing Scripts](../../api-source/utilities/scripts/)

## Getting Started

1. **Obtain JWT Token**: Use the `/api/v1/auth/login` endpoint
2. **Explore APIs**: Visit the [Interactive API Explorer](http://localhost:8080/swagger/)
3. **Test Endpoints**: Use the built-in "Try it out" functionality
4. **Review Examples**: Check the practical examples in each section

For detailed implementation guides, see the [API Integration Tutorial](../tutorials/api-integration.md).