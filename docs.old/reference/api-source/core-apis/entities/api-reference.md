# Entity API Reference

## Overview
Complete reference for entity management endpoints in the AWO ERP system. Entities represent the core business objects and organizational units.

## Base URL
```
POST /api/v1/entities
GET  /api/v1/entities
GET  /api/v1/entities/{id}
PUT  /api/v1/entities/{id}
DELETE /api/v1/entities/{id}
```

## Authentication
All entity endpoints require a valid JWT token in the Authorization header:
```
Authorization: Bearer <your-jwt-token>
```

## Endpoints

### Create Entity
**POST /api/v1/entities**

Creates a new entity in the system.

**Request Body:**
```json
{
  "name": "string (required)",
  "description": "string",
  "type": "organization|department|project",
  "parent_id": "uuid (optional)",
  "attributes": {
    "key": "value"
  }
}
```

**Response (201 Created):**
```json
{
  "id": "uuid",
  "name": "string",
  "slug": "string",
  "description": "string",
  "type": "string",
  "parent_id": "uuid",
  "tenant_id": "uuid",
  "attributes": {},
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-01T00:00:00Z"
}
```

### List Entities
**GET /api/v1/entities**

Retrieves a paginated list of entities.

**Query Parameters:**
- `page` (integer): Page number (default: 1)
- `limit` (integer): Items per page (default: 20, max: 100)
- `type` (string): Filter by entity type
- `parent_id` (uuid): Filter by parent entity
- `search` (string): Search in name and description

**Response (200 OK):**
```json
{
  "entities": [
    {
      "id": "uuid",
      "name": "string",
      "slug": "string",
      "type": "string",
      "parent_id": "uuid",
      "created_at": "2025-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "pages": 5
  }
}
```

### Get Entity Details
**GET /api/v1/entities/{id}**

Retrieves detailed information about a specific entity.

**Path Parameters:**
- `id` (uuid): Entity identifier

**Response (200 OK):**
```json
{
  "id": "uuid",
  "name": "string",
  "slug": "string",
  "description": "string",
  "type": "string",
  "parent_id": "uuid",
  "tenant_id": "uuid",
  "attributes": {},
  "children": [
    {
      "id": "uuid",
      "name": "string",
      "type": "string"
    }
  ],
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-01T00:00:00Z"
}
```

### Update Entity
**PUT /api/v1/entities/{id}**

Updates an existing entity.

**Path Parameters:**
- `id` (uuid): Entity identifier

**Request Body:**
```json
{
  "name": "string",
  "description": "string",
  "type": "string",
  "parent_id": "uuid",
  "attributes": {}
}
```

**Response (200 OK):**
```json
{
  "id": "uuid",
  "name": "string",
  "slug": "string",
  "description": "string",
  "type": "string",
  "parent_id": "uuid",
  "tenant_id": "uuid",
  "attributes": {},
  "updated_at": "2025-01-01T00:00:00Z"
}
```

### Delete Entity
**DELETE /api/v1/entities/{id}**

Deletes an entity and all its relationships.

**Path Parameters:**
- `id` (uuid): Entity identifier

**Response (204 No Content)**

## Error Responses

### 400 Bad Request
```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "Validation failed",
    "details": [
      {
        "field": "name",
        "message": "Name is required"
      }
    ]
  }
}
```

### 404 Not Found
```json
{
  "error": {
    "code": "ENTITY_NOT_FOUND",
    "message": "Entity not found"
  }
}
```

### 409 Conflict
```json
{
  "error": {
    "code": "ENTITY_CONFLICT",
    "message": "Entity with this name already exists"
  }
}
```

## Examples

### Create a Department Entity
```bash
curl -X POST "$API_BASE_URL/api/v1/entities" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Engineering Department",
    "description": "Software development and infrastructure",
    "type": "department",
    "attributes": {
      "budget_code": "ENG-001",
      "manager": "john.doe@company.com"
    }
  }'
```

### Search Entities
```bash
curl -X GET "$API_BASE_URL/api/v1/entities?search=engineering&type=department" \
  -H "Authorization: Bearer $AUTH_TOKEN"
```
