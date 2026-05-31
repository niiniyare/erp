# TestModule Module - API Reference

**Version**: 1.0  
**Date**: October 12, 2025  
**Status**: In Development  
**OpenAPI Version**: 3.0.3

---

## Overview

### API Description
The TestModule Module API provides comprehensive test_module management capabilities within the AWO ERP system. This business-focused API supports multi-tenant operations, ABAC-based access control, and enterprise-grade compliance features.

### Base Information
- **Base URL**: `https://api.awo-erp.com/api/v1/test-module`
- **Authentication**: Bearer Token with Attribute-Based Access Control
- **Content Type**: `application/json`
- **API Version**: `v1`

### Quick Links
- [Authentication Guide](../../../reference/api/auth/index.md) - Get started with API authentication
- [SDK Examples](examples/) - Code examples in multiple languages

---

## Table of Contents

1. [Overview](#overview)
2. [Authentication](#authentication)
3. [TestModule Endpoints](#test_module-endpoints)
4. [Data Models](#data-models)
5. [Error Handling](#error-handling)
6. [Rate Limiting](#rate-limiting)
7. [Examples](#examples)

---

## Authentication

### Bearer Token Authentication
All API requests require authentication using a Bearer token in the Authorization header:

```http
Authorization: Bearer <your-api-token>
```

### ABAC Permissions Required
The following permissions are required for TestModule operations:


- `test_module.create`: _ operations

- `test_module.read`: _ operations

- `test_module.update`: _ operations

- `test_module.delete`: _ operations

- `test_module.list`: _ operations

- `test_module.search`: _ operations


[Authentication Guide →](../../../reference/api/auth/index.md)

---

## TestModule Endpoints

### List TestModule

Retrieve a list of TestModule with optional filtering and pagination.

**Endpoint**: `GET /test_module`

**Required Permissions**: `test_module.test_module.list`

**Query Parameters**:
| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `limit` | integer | Number of results to return (1-100) | 20 |
| `offset` | integer | Number of results to skip | 0 |
| `status` | string | Filter by status (`active`, `inactive`, `pending`, `archived`) | - |
| `name` | string | Filter by name (partial match) | - |
| `sort_by` | string | Sort field (`name`, `created_at`, `updated_at`) | `created_at` |
| `sort_order` | string | Sort order (`asc`, `desc`) | `desc` |

**Example Request**:
```http
GET /api/v1/test-module/test_module?limit=10&status=active&sort_by=name
Authorization: Bearer <token>
```

**Example Response**:
```json
{
  "data": [
    {
      "id": "01234567-89ab-cdef-0123-456789abcdef",
      "name": "Sample TestModule",
      "description": "A sample test_module for demonstration",
      "status": "active",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ],
  "pagination": {
    "total": 150,
    "count": 10,
    "limit": 10,
    "offset": 0,
    "has_next": true,
    "has_previous": false
  }
}
```

---

### Create TestModule

Create a new test_module with the provided information.

**Endpoint**: `POST /test_module`

**Required Permissions**: `test_module.test_module.create`

**Request Body**:
```json
{
  "name": "string (required, 1-255 characters)",
  "description": "string (optional, max 1000 characters)"
}
```

**Example Request**:
```http
POST /api/v1/test-module/test_module
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "New TestModule",
  "description": "Description for the new test_module"
}
```

**Example Response** (201 Created):
```json
{
  "id": "01234567-89ab-cdef-0123-456789abcdef",
  "name": "New TestModule",
  "description": "Description for the new test_module",
  "status": "active",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

---

### Get TestModule by ID

Retrieve a specific test_module by its unique identifier.

**Endpoint**: `GET /test_module/{id}`

**Required Permissions**: `test_module.test_module.read`

**Path Parameters**:
| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | string (UUID) | Unique identifier of the test_module |

**Example Request**:
```http
GET /api/v1/test-module/test_module/01234567-89ab-cdef-0123-456789abcdef
Authorization: Bearer <token>
```

**Example Response** (200 OK):
```json
{
  "id": "01234567-89ab-cdef-0123-456789abcdef",
  "name": "Sample TestModule",
  "description": "A sample test_module for demonstration",
  "status": "active",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

---

### Update TestModule

Update an existing test_module with new information.

**Endpoint**: `PUT /test_module/{id}`

**Required Permissions**: `test_module.test_module.update`

**Path Parameters**:
| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | string (UUID) | Unique identifier of the test_module |

**Request Body**:
```json
{
  "name": "string (optional, 1-255 characters)",
  "description": "string (optional, max 1000 characters)"
}
```

**Example Request**:
```http
PUT /api/v1/test-module/test_module/01234567-89ab-cdef-0123-456789abcdef
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Updated TestModule",
  "description": "Updated description"
}
```

**Example Response** (200 OK):
```json
{
  "id": "01234567-89ab-cdef-0123-456789abcdef",
  "name": "Updated TestModule",
  "description": "Updated description",
  "status": "active",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T11:45:00Z"
}
```

---

### Delete TestModule

Delete an existing test_module.

**Endpoint**: `DELETE /test_module/{id}`

**Required Permissions**: `test_module.test_module.delete`

**Path Parameters**:
| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | string (UUID) | Unique identifier of the test_module |

**Example Request**:
```http
DELETE /api/v1/test-module/test_module/01234567-89ab-cdef-0123-456789abcdef
Authorization: Bearer <token>
```

**Example Response** (204 No Content):
```
(Empty response body)
```

---

## Data Models

### TestModule

Represents a test_module entity in the system.

```json
{
  "id": "string (UUID)",
  "name": "string",
  "description": "string | null",
  "status": "string (enum)",
  "created_at": "string (ISO 8601)",
  "updated_at": "string (ISO 8601)"
}
```

**Field Descriptions**:
- `id`: Unique identifier (UUID v4)
- `name`: TestModule name (1-255 characters)
- `description`: Optional description (max 1000 characters)
- `status`: Current status (`active`, `inactive`, `pending`, `archived`)
- `created_at`: Creation timestamp
- `updated_at`: Last modification timestamp

### TestModuleStatus Enum

Valid status values for TestModule entities:

- `active`: TestModule is active and operational
- `inactive`: TestModule is inactive but can be reactivated
- `pending`: TestModule is pending approval or setup
- `archived`: TestModule is archived and read-only

### Pagination Response

Standard pagination wrapper for list endpoints:

```json
{
  "data": [/* array of entities */],
  "pagination": {
    "total": "integer",
    "count": "integer", 
    "limit": "integer",
    "offset": "integer",
    "has_next": "boolean",
    "has_previous": "boolean"
  }
}
```

---

## Error Handling

### Error Response Format

All errors follow a consistent JSON structure:

```json
{
  "error": {
    "code": "string",
    "message": "string",
    "details": "string (optional)",
    "correlation_id": "string"
  }
}
```

### HTTP Status Codes

| Status Code | Description | Common Causes |
|-------------|-------------|---------------|
| `400` | Bad Request | Invalid request format, validation errors |
| `401` | Unauthorized | Missing or invalid authentication token |
| `403` | Forbidden | Insufficient permissions for the operation |
| `404` | Not Found | TestModule not found or inaccessible |
| `409` | Conflict | TestModule name already exists |
| `422` | Unprocessable Entity | Business rule validation failures |
| `429` | Too Many Requests | Rate limit exceeded |
| `500` | Internal Server Error | Unexpected server error |

### Example Error Responses

**Validation Error (400)**:
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid test_module data",
    "details": "Name is required and must be between 1 and 255 characters",
    "correlation_id": "req_01234567"
  }
}
```

**Permission Error (403)**:
```json
{
  "error": {
    "code": "INSUFFICIENT_PERMISSIONS",
    "message": "Access denied",
    "details": "Missing required permission: test_module.test_module.create",
    "correlation_id": "req_01234567"
  }
}
```

**Not Found Error (404)**:
```json
{
  "error": {
    "code": "TEST_MODULE_NOT_FOUND",
    "message": "TestModule not found",
    "details": "TestModule with ID 01234567-89ab-cdef-0123-456789abcdef was not found",
    "correlation_id": "req_01234567"
  }
}
```

---

## Rate Limiting

### Rate Limits

The API implements rate limiting to ensure fair usage:

- **Standard Operations**: 1000 requests per hour per token
- **List Operations**: 100 requests per hour per token
- **Bulk Operations**: 50 requests per hour per token

### Rate Limit Headers

Rate limit information is included in response headers:

```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200
```

---

## Examples

### Complete CRUD Workflow

```bash
# 1. Create a new test_module
curl -X POST "https://api.awo-erp.com/api/v1/test-module/test_module" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Example TestModule",
    "description": "This is an example test_module"
  }'

# 2. List TestModule
curl -X GET "https://api.awo-erp.com/api/v1/test-module/test_module?limit=10" \
  -H "Authorization: Bearer <token>"

# 3. Get specific test_module
curl -X GET "https://api.awo-erp.com/api/v1/test-module/test_module/01234567-89ab-cdef-0123-456789abcdef" \
  -H "Authorization: Bearer <token>"

# 4. Update test_module
curl -X PUT "https://api.awo-erp.com/api/v1/test-module/test_module/01234567-89ab-cdef-0123-456789abcdef" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated TestModule",
    "description": "Updated description"
  }'

# 5. Delete test_module
curl -X DELETE "https://api.awo-erp.com/api/v1/test-module/test_module/01234567-89ab-cdef-0123-456789abcdef" \
  -H "Authorization: Bearer <token>"
```

### Filtering and Pagination

```bash
# Get active TestModule with pagination
curl -X GET "https://api.awo-erp.com/api/v1/test-module/test_module?status=active&limit=20&offset=40&sort_by=name&sort_order=asc" \
  -H "Authorization: Bearer <token>"

# Search TestModule by name
curl -X GET "https://api.awo-erp.com/api/v1/test-module/test_module?name=example&limit=10" \
  -H "Authorization: Bearer <token>"
```

---

## SDK Examples

### Go SDK Example

```go
package main

import (
    "context"
    "fmt"
    "awo-sdk-go/test_module"
)

func main() {
    client := test_module.NewClient("https://api.awo-erp.com", "your-api-token")
    
    // Create test_module
    testModule, err := client.CreateTestModule(context.Background(), test_module.CreateTestModuleRequest{
        Name:        "Example TestModule",
        Description: ptr("This is an example"),
    })
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Created test_module: %s\n", testModule.ID)
}

func ptr(s string) *string { return &s }
```

### JavaScript SDK Example

```javascript
import { TestModuleClient } from '@awo-erp/sdk';

const client = new TestModuleClient({
  baseURL: 'https://api.awo-erp.com',
  apiToken: 'your-api-token'
});

// Create test_module
const testModule = await client.createTestModule({
  name: 'Example TestModule',
  description: 'This is an example'
});

console.log('Created test_module:', testModule.id);
```

---

**API Version**: 1.0.0  
**Generated**: 2025-10-12 22:32:13  
**Generator**: awoctl 0.1.0