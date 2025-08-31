# AWO ERP API Reference

## Overview

The AWO ERP API provides comprehensive programmatic access to all system functionality through a RESTful interface. Built with the Goa framework, all APIs are auto-documented with OpenAPI 3.0 specifications and include interactive testing capabilities.

## Quick Start

### 1. Authentication
```bash
# Login to get JWT token
curl -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "your-username",
    "password": "your-password", 
    "tenant_id": "your-tenant-id"
  }'
```

### 2. Make API Calls
```bash
# Use JWT token in subsequent requests
curl -X GET "http://localhost:8080/api/v1/users/profile" \
  -H "Authorization: Bearer $JWT_TOKEN"
```

### 3. Explore Interactive Documentation
Visit the [Swagger UI](swagger-ui.md) for hands-on API exploration and testing.

## API Categories

| Category | Base Path | Endpoints | Description |
|----------|-----------|-----------|-------------|
| **[ABAC Security](abac/)** | `/abac/*` | 10 | Advanced authorization and policy evaluation |
| **[Authentication](auth/)** | `/api/v1/auth/*` | 4 | User login, token management, sessions |
| **[Users](users/)** | `/api/v1/users/*` | 8 | User management and profiles |
| **[Organizations](organizations/)** | `/api/v1/organizations/*` | 6 | Organization hierarchy management |  
| **Tenants** | `/api/v1/tenants/*` | 6 | Multi-tenant administration |
| **Access Requests** | `/api/v1/access-requests/*` | 6 | Workflow and approval management |
| **Analytics** | `/api/v1/analytics/*` | 8 | User behavior and insights |
| **Feature Flags** | `/api/v1/feature-flags/*` | 6 | Dynamic feature management |

## Standards & Conventions

### HTTP Methods
- **GET**: Retrieve resources (read-only)
- **POST**: Create new resources
- **PUT**: Update entire resources
- **PATCH**: Partial resource updates
- **DELETE**: Remove resources

### Status Codes
- **200 OK**: Successful GET, PUT, PATCH
- **201 Created**: Successful POST
- **204 No Content**: Successful DELETE
- **400 Bad Request**: Invalid request data
- **401 Unauthorized**: Missing or invalid authentication
- **403 Forbidden**: Insufficient permissions
- **404 Not Found**: Resource doesn't exist
- **429 Too Many Requests**: Rate limit exceeded
- **500 Internal Server Error**: Server-side error

### Pagination
List endpoints support cursor-based pagination:

```bash
# Request with pagination
curl "http://localhost:8080/api/v1/users?page=1&limit=20"

# Response includes pagination metadata
{
  "data": [...],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "has_next": true,
    "next_cursor": "eyJ..."
  }
}
```

### Filtering & Sorting
Many endpoints support query parameters for filtering and sorting:

```bash
# Filter users by status and department
curl "http://localhost:8080/api/v1/users?status=active&department=engineering"

# Sort by created date
curl "http://localhost:8080/api/v1/users?sort=created_at&order=desc"
```

## Error Handling

All error responses follow a consistent format:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Request validation failed",
    "details": {
      "field": "email",
      "reason": "Invalid email format"
    }
  },
  "meta": {
    "timestamp": "2025-01-01T12:00:00Z",
    "request_id": "req-12345",
    "trace_id": "trace-67890"
  }
}
```

### Common Error Codes
- `VALIDATION_ERROR`: Request data validation failed
- `AUTHENTICATION_REQUIRED`: Missing authentication token
- `AUTHORIZATION_FAILED`: Insufficient permissions
- `RESOURCE_NOT_FOUND`: Requested resource doesn't exist
- `RATE_LIMIT_EXCEEDED`: Too many requests
- `INTERNAL_ERROR`: Unexpected server error

## Security

### Authentication
All endpoints require JWT authentication except:
- `/api/v1/auth/login`
- `/api/v1/health`
- `/swagger/*` (documentation)

### Authorization
Fine-grained permissions are enforced using ABAC (Attribute-Based Access Control). Each request is evaluated against policies that consider:
- User attributes (role, department, clearance)
- Resource attributes (type, sensitivity, owner)
- Environmental attributes (time, location, context)
- Action being performed

### Rate Limiting
API usage is tracked and limited per user/tenant:
- Standard operations: 1000/hour
- Authentication attempts: 60/hour
- Analytics queries: 100/hour
- Bulk operations: 10/hour

## Testing & Development

### Interactive Testing
Use the [Swagger UI](swagger-ui.md) for interactive API testing with live data.

### Automated Testing
- **Test Scripts**: [Shell scripts](../api-source/utilities/scripts/) for automated testing
- **Postman Collections**: [JSON collections](../api-source/utilities/postman/) for API exploration
- **SDK Examples**: Code samples for common integration patterns

### Development Environment
For local development, the API server runs on `http://localhost:8080` with:
- Hot reloading enabled
- Detailed error responses
- CORS configured for frontend development
- Debug logging available

## Support & Documentation

- **Interactive API Explorer**: [Swagger UI](http://localhost:8080/swagger/)
- **OpenAPI Specifications**: Available in JSON format
- **Integration Guides**: Step-by-step implementation tutorials
- **Code Examples**: Practical usage patterns and best practices

Need help? Check the [troubleshooting guide](../../contributing/01-best-practices.md) or review the [API integration tutorial](tutorials/api-integration.md).