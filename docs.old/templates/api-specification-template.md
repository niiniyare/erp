# {Module Name} - API Reference

**Version**: 1.0  
**Date**: {Current Date}  
**Status**: Draft | Review | Production  
**OpenAPI Version**: 3.0.3

---

## Overview

### API Description
Brief description of the {Module Name} API, its purpose, and core functionality within the AWO ERP system.

### Base Information
- **Base URL**: `https://api.awo-erp.com/api/v1/{module}`
- **Authentication**: Bearer Token (JWT)
- **Content Type**: `application/json`
- **API Version**: `v1`

### Quick Links
- [Interactive API Explorer](../../../reference/api/swagger-ui.md) - Test endpoints directly
- [Authentication Guide](../../../reference/api/auth/index.md) - Get started with API authentication
- [SDK Examples](examples/) - Code examples in multiple languages

---

## Authentication

### Bearer Token Authentication
All API endpoints require authentication using JWT bearer tokens.

```bash
# Include in request headers
Authorization: Bearer <your-jwt-token>
X-Tenant-ID: <tenant-id>  # Required for multi-tenant operations
```

### Required Headers
| Header | Required | Description |
|--------|----------|-------------|
| `Authorization` | Yes | Bearer JWT token |
| `X-Tenant-ID` | Yes | Tenant context identifier |
| `Content-Type` | Yes | `application/json` for POST/PUT requests |
| `Accept` | No | `application/json` (default) |

---

## Core Endpoints

### Entity Management

#### Create Entity
Create a new entity within the tenant context.

**Endpoint**: `POST /{module}/entities`

**Request Body**:
```json
{
  "code": "ENT001",
  "name": "Example Entity",
  "description": "Optional description",
  "type": "standard",
  "properties": {
    "customField1": "value1",
    "customField2": "value2"
  },
  "isActive": true
}
```

**Response** (201 Created):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "tenantId": "123e4567-e89b-12d3-a456-426614174000",
  "code": "ENT001",
  "name": "Example Entity",
  "description": "Optional description",
  "type": "standard",
  "status": "active",
  "isActive": true,
  "properties": {
    "customField1": "value1",
    "customField2": "value2"
  },
  "createdAt": "2024-01-15T10:30:00Z",
  "updatedAt": "2024-01-15T10:30:00Z",
  "version": 1
}
```

**Business Rules**:
- Entity code must be unique within tenant
- Required fields: `code`, `name`, `type`
- Code format: Alphanumeric, max 50 characters
- Name max length: 255 characters

**Error Responses**:
```json
// 400 Bad Request - Validation Error
{
  "error": "validation_failed",
  "message": "Request validation failed",
  "details": [
    {
      "field": "code",
      "code": "required",
      "message": "Entity code is required"
    }
  ],
  "timestamp": "2024-01-15T10:30:00Z",
  "correlationId": "abc123-def456"
}

// 409 Conflict - Duplicate Code
{
  "error": "duplicate_code",
  "message": "Entity code already exists in tenant",
  "timestamp": "2024-01-15T10:30:00Z",
  "correlationId": "abc123-def456"
}
```

#### Get Entity by ID
Retrieve a specific entity by its ID.

**Endpoint**: `GET /{module}/entities/{id}`

**Path Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | UUID | Yes | Entity identifier |

**Response** (200 OK):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "tenantId": "123e4567-e89b-12d3-a456-426614174000",
  "code": "ENT001",
  "name": "Example Entity",
  "description": "Optional description",
  "type": "standard",
  "status": "active",
  "isActive": true,
  "properties": {
    "customField1": "value1",
    "customField2": "value2"
  },
  "createdAt": "2024-01-15T10:30:00Z",
  "updatedAt": "2024-01-15T10:30:00Z",
  "version": 1,
  "_links": {
    "self": {
      "href": "/api/v1/{module}/entities/550e8400-e29b-41d4-a716-446655440000"
    },
    "update": {
      "href": "/api/v1/{module}/entities/550e8400-e29b-41d4-a716-446655440000",
      "method": "PUT"
    },
    "delete": {
      "href": "/api/v1/{module}/entities/550e8400-e29b-41d4-a716-446655440000",
      "method": "DELETE"
    }
  }
}
```

**Error Responses**:
```json
// 404 Not Found
{
  "error": "entity_not_found",
  "message": "Entity not found or access denied",
  "timestamp": "2024-01-15T10:30:00Z",
  "correlationId": "abc123-def456"
}
```

#### List Entities
Retrieve a paginated list of entities with optional filtering.

**Endpoint**: `GET /{module}/entities`

**Query Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `page` | integer | No | Page number (default: 1) |
| `limit` | integer | No | Items per page (default: 20, max: 100) |
| `status` | string | No | Filter by status (`active`, `inactive`, `all`) |
| `type` | string | No | Filter by entity type |
| `search` | string | No | Search in code and name fields |
| `sort` | string | No | Sort field (`code`, `name`, `createdAt`, `updatedAt`) |
| `order` | string | No | Sort order (`asc`, `desc`) |

**Response** (200 OK):
```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "code": "ENT001",
      "name": "Example Entity 1",
      "type": "standard",
      "status": "active",
      "isActive": true,
      "createdAt": "2024-01-15T10:30:00Z",
      "updatedAt": "2024-01-15T10:30:00Z"
    },
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "code": "ENT002", 
      "name": "Example Entity 2",
      "type": "premium",
      "status": "active",
      "isActive": true,
      "createdAt": "2024-01-15T11:00:00Z",
      "updatedAt": "2024-01-15T11:00:00Z"
    }
  ],
  "pagination": {
    "currentPage": 1,
    "totalPages": 5,
    "totalItems": 87,
    "itemsPerPage": 20,
    "hasNextPage": true,
    "hasPreviousPage": false
  },
  "_links": {
    "self": {
      "href": "/api/v1/{module}/entities?page=1&limit=20"
    },
    "next": {
      "href": "/api/v1/{module}/entities?page=2&limit=20"
    },
    "first": {
      "href": "/api/v1/{module}/entities?page=1&limit=20"
    },
    "last": {
      "href": "/api/v1/{module}/entities?page=5&limit=20"
    }
  }
}
```

#### Update Entity
Update an existing entity.

**Endpoint**: `PUT /{module}/entities/{id}`

**Path Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | UUID | Yes | Entity identifier |

**Request Body**:
```json
{
  "name": "Updated Entity Name",
  "description": "Updated description",
  "properties": {
    "customField1": "newValue1",
    "newField": "newValue"
  },
  "version": 1  // For optimistic locking
}
```

**Response** (200 OK):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "tenantId": "123e4567-e89b-12d3-a456-426614174000",
  "code": "ENT001",
  "name": "Updated Entity Name",
  "description": "Updated description",
  "type": "standard",
  "status": "active",
  "isActive": true,
  "properties": {
    "customField1": "newValue1",
    "customField2": "value2",
    "newField": "newValue"
  },
  "createdAt": "2024-01-15T10:30:00Z",
  "updatedAt": "2024-01-15T12:15:00Z",
  "version": 2
}
```

**Error Responses**:
```json
// 409 Conflict - Version Mismatch
{
  "error": "version_conflict",
  "message": "Entity has been modified by another user",
  "currentVersion": 3,
  "providedVersion": 1,
  "timestamp": "2024-01-15T10:30:00Z",
  "correlationId": "abc123-def456"
}
```

#### Delete Entity
Soft delete an entity (sets isActive to false).

**Endpoint**: `DELETE /{module}/entities/{id}`

**Path Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | UUID | Yes | Entity identifier |

**Response** (204 No Content)

**Error Responses**:
```json
// 400 Bad Request - Business Rule Violation
{
  "error": "cannot_delete_entity",
  "message": "Entity cannot be deleted due to active dependencies",
  "details": {
    "dependencies": [
      {
        "type": "transactions",
        "count": 5
      }
    ]
  },
  "timestamp": "2024-01-15T10:30:00Z",
  "correlationId": "abc123-def456"
}
```

### Search Endpoints

#### Search by Code
Find entity by exact code match.

**Endpoint**: `GET /{module}/entities/by-code/{code}`

**Path Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `code` | string | Yes | Entity code |

**Response**: Same as Get Entity by ID

#### Advanced Search
Perform complex search with multiple criteria.

**Endpoint**: `POST /{module}/entities/search`

**Request Body**:
```json
{
  "criteria": {
    "status": ["active", "inactive"],
    "type": "standard",
    "createdAfter": "2024-01-01T00:00:00Z",
    "properties": {
      "customField1": "value1"
    }
  },
  "sort": [
    {
      "field": "createdAt",
      "order": "desc"
    },
    {
      "field": "name",
      "order": "asc"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 50
  }
}
```

---

## Data Models

### Entity Model
```json
{
  "type": "object",
  "properties": {
    "id": {
      "type": "string",
      "format": "uuid",
      "description": "Unique entity identifier",
      "readOnly": true
    },
    "tenantId": {
      "type": "string", 
      "format": "uuid",
      "description": "Tenant identifier",
      "readOnly": true
    },
    "code": {
      "type": "string",
      "maxLength": 50,
      "pattern": "^[A-Za-z0-9_-]+$",
      "description": "Unique entity code within tenant"
    },
    "name": {
      "type": "string",
      "maxLength": 255,
      "description": "Entity display name"
    },
    "description": {
      "type": "string",
      "maxLength": 1000,
      "description": "Optional entity description"
    },
    "type": {
      "type": "string",
      "enum": ["standard", "premium", "enterprise"],
      "description": "Entity classification type"
    },
    "status": {
      "type": "string",
      "enum": ["active", "inactive", "archived"],
      "description": "Current entity status",
      "readOnly": true
    },
    "isActive": {
      "type": "boolean",
      "description": "Active flag for soft delete"
    },
    "properties": {
      "type": "object",
      "description": "Custom properties as key-value pairs",
      "additionalProperties": true
    },
    "createdAt": {
      "type": "string",
      "format": "date-time",
      "description": "Creation timestamp",
      "readOnly": true
    },
    "updatedAt": {
      "type": "string",
      "format": "date-time", 
      "description": "Last update timestamp",
      "readOnly": true
    },
    "version": {
      "type": "integer",
      "description": "Version for optimistic locking",
      "readOnly": true
    }
  },
  "required": ["code", "name", "type"]
}
```

### Error Response Model
```json
{
  "type": "object",
  "properties": {
    "error": {
      "type": "string",
      "description": "Error code identifier"
    },
    "message": {
      "type": "string",
      "description": "Human-readable error message"
    },
    "details": {
      "type": "array",
      "description": "Validation error details",
      "items": {
        "type": "object",
        "properties": {
          "field": {"type": "string"},
          "code": {"type": "string"},
          "message": {"type": "string"}
        }
      }
    },
    "timestamp": {
      "type": "string",
      "format": "date-time",
      "description": "Error occurrence timestamp"
    },
    "correlationId": {
      "type": "string",
      "description": "Request correlation identifier for tracing"
    }
  },
  "required": ["error", "message", "timestamp"]
}
```

---

## Error Handling

### Standard Error Codes

| HTTP Status | Error Code | Description |
|-------------|------------|-------------|
| 400 | `validation_failed` | Request validation failed |
| 400 | `invalid_request` | Malformed request |
| 401 | `unauthorized` | Authentication required |
| 403 | `forbidden` | Insufficient permissions |
| 404 | `entity_not_found` | Entity not found or access denied |
| 409 | `duplicate_code` | Entity code already exists |
| 409 | `version_conflict` | Optimistic locking conflict |
| 422 | `business_rule_violation` | Business rule constraint violated |
| 429 | `rate_limit_exceeded` | API rate limit exceeded |
| 500 | `internal_server_error` | Unexpected server error |

### Error Response Format
All error responses follow a consistent format:

```json
{
  "error": "error_code",
  "message": "Human-readable description",
  "details": [/* Additional context */],
  "timestamp": "2024-01-15T10:30:00Z",
  "correlationId": "abc123-def456"
}
```

---

## Rate Limiting

### Rate Limits
- **Standard Tier**: 1,000 requests per hour
- **Premium Tier**: 5,000 requests per hour  
- **Enterprise Tier**: 10,000 requests per hour

### Rate Limit Headers
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1642248000
```

### Rate Limit Exceeded Response
```json
{
  "error": "rate_limit_exceeded",
  "message": "API rate limit exceeded",
  "details": {
    "limit": 1000,
    "windowSeconds": 3600,
    "retryAfter": 1800
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

---

## Code Examples

### cURL Examples

#### Create Entity
```bash
curl -X POST "https://api.awo-erp.com/api/v1/{module}/entities" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "X-Tenant-ID: YOUR_TENANT_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "ENT001",
    "name": "Example Entity",
    "type": "standard",
    "isActive": true
  }'
```

#### Get Entity List
```bash
curl -X GET "https://api.awo-erp.com/api/v1/{module}/entities?page=1&limit=20&status=active" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "X-Tenant-ID: YOUR_TENANT_ID"
```

### JavaScript/Node.js Examples

#### Using Axios
```javascript
const axios = require('axios');

const client = axios.create({
  baseURL: 'https://api.awo-erp.com/api/v1/{module}',
  headers: {
    'Authorization': `Bearer ${process.env.JWT_TOKEN}`,
    'X-Tenant-ID': process.env.TENANT_ID,
    'Content-Type': 'application/json'
  }
});

// Create entity
async function createEntity(entityData) {
  try {
    const response = await client.post('/entities', entityData);
    return response.data;
  } catch (error) {
    console.error('Error creating entity:', error.response.data);
    throw error;
  }
}

// Get entity by ID
async function getEntity(entityId) {
  try {
    const response = await client.get(`/entities/${entityId}`);
    return response.data;
  } catch (error) {
    if (error.response.status === 404) {
      return null; // Entity not found
    }
    throw error;
  }
}
```

### Go Examples

#### Using HTTP Client
```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
)

type Entity struct {
    ID          string                 `json:"id,omitempty"`
    Code        string                 `json:"code"`
    Name        string                 `json:"name"`
    Type        string                 `json:"type"`
    IsActive    bool                   `json:"isActive"`
    Properties  map[string]interface{} `json:"properties,omitempty"`
}

func createEntity(entity Entity) (*Entity, error) {
    jsonData, err := json.Marshal(entity)
    if err != nil {
        return nil, err
    }

    req, err := http.NewRequest("POST", 
        "https://api.awo-erp.com/api/v1/{module}/entities", 
        bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }

    req.Header.Set("Authorization", "Bearer "+os.Getenv("JWT_TOKEN"))
    req.Header.Set("X-Tenant-ID", os.Getenv("TENANT_ID"))
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusCreated {
        return nil, fmt.Errorf("API error: %s", resp.Status)
    }

    var result Entity
    err = json.NewDecoder(resp.Body).Decode(&result)
    return &result, err
}
```

---

## Testing

### Postman Collection
A  Postman collection is available with:
- Pre-configured environments (dev, staging, production)
- Authentication setup scripts
- Complete endpoint coverage
- Example requests and responses
- Automated tests for response validation

**Download**: [Postman Collection](postman/{module}-api-collection.json)

### API Testing Checklist
- [ ] Authentication works correctly
- [ ] CRUD operations function properly
- [ ] Pagination works as expected
- [ ] Search and filtering work correctly
- [ ] Error responses are properly formatted
- [ ] Rate limiting is enforced
- [ ] Multi-tenancy isolation is verified
- [ ] Performance requirements are met

---

## Changelog

### Version 1.0.0 (2024-01-15)
- Initial API release
- Complete CRUD operations for entities
- Search and filtering capabilities
- Multi-tenant support
- Authentication and authorization
- Rate limiting implementation

### Version 1.1.0 (Planned)
- Bulk operations support
- Webhook notifications
- Advanced analytics endpoints
- GraphQL API support

---

**Document Control**  
- **Version**: 1.0
- **Last Updated**: {Date}
- **API Status**: {Draft | Review | Production}
- **OpenAPI Spec**: [swagger.yaml](openapi/swagger.yaml)