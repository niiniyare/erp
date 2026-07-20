> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

# Organization Management API

The Organization Management API provides  functionality for managing organizational structures, hierarchies, and entities within the AWO ERP System. This API supports complex organizational relationships, department management, and hierarchical data structures.

##  Overview

**Base URL**: `/api/v1/organizations`

**Purpose**: Manage organizational structures, hierarchies, departments, and business entities with support for complex organizational relationships and multi-level hierarchies.

**Key Features**:
- Organization CRUD operations
- Hierarchical organization structures
- Department and division management
- Organization archiving and status management
- Organization type classification
- Nested organization relationships

##  Available Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/organizations` | Create a new organization |
| `GET` | `/api/v1/organizations` | List all organizations with pagination |
| `GET` | `/api/v1/organizations/{id}` | Get organization by ID |
| `PUT` | `/api/v1/organizations/{id}` | Update organization information |
| `GET` | `/api/v1/organizations/{id}/hierarchy` | Get organization hierarchy |
| `PATCH` | `/api/v1/organizations/{id}/archive` | Archive organization |

##  Data Models

### Organization Object
```json
{
  "id": "string (UUID)",
  "name": "string",
  "organization_type": "string (CORPORATION|LLC|PARTNERSHIP|NONPROFIT|GOVERNMENT|DEPARTMENT|DIVISION|TEAM)",
  "status": "string (ACTIVE|INACTIVE|ARCHIVED)",
  "created_at": "string (ISO 8601)",
  "updated_at": "string (ISO 8601)"
}
```

### Organization Hierarchy Object
```json
{
  "id": "string (UUID)",
  "name": "string",
  "organization_type": "string",
  "status": "string",
  "parent_id": "string (UUID, optional)",
  "children": [
    {
      "id": "string (UUID)",
      "name": "string",
      "organization_type": "string",
      "children": []
    }
  ],
  "level": "integer",
  "path": "string"
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

### Create Organization
```bash
curl -X POST http://localhost:8080/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Corporation",
    "organization_type": "CORPORATION"
  }' | jq .
```

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Acme Corporation",
  "organization_type": "CORPORATION",
  "status": "ACTIVE",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### List Organizations
```bash
curl -X GET "http://localhost:8080/api/v1/organizations?limit=10&offset=0" | jq .
```

**Response:**
```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Acme Corporation",
      "organization_type": "CORPORATION",
      "status": "ACTIVE",
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

### Get Organization
```bash
curl -X GET http://localhost:8080/api/v1/organizations/550e8400-e29b-41d4-a716-446655440000 | jq .
```

### Update Organization
```bash
curl -X PUT http://localhost:8080/api/v1/organizations/550e8400-e29b-41d4-a716-446655440000 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Corporation Updated",
    "organization_type": "LLC"
  }' | jq .
```

### Get Organization Hierarchy
```bash
curl -X GET http://localhost:8080/api/v1/organizations/550e8400-e29b-41d4-a716-446655440000/hierarchy | jq .
```

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Acme Corporation",
  "organization_type": "CORPORATION",
  "status": "ACTIVE",
  "children": [
    {
      "id": "650e8400-e29b-41d4-a716-446655440001",
      "name": "Engineering Department",
      "organization_type": "DEPARTMENT",
      "children": [
        {
          "id": "750e8400-e29b-41d4-a716-446655440002",
          "name": "Backend Team",
          "organization_type": "TEAM",
          "children": []
        }
      ]
    }
  ]
}
```

### Archive Organization
```bash
curl -X PATCH http://localhost:8080/api/v1/organizations/550e8400-e29b-41d4-a716-446655440000/archive \
  -H "Content-Type: application/json" \
  -d '{
    "reason": "Organizational restructuring"
  }'
```

##  Advanced Usage

### Organization Types

| Type | Description | Use Case |
|------|-------------|----------|
| `CORPORATION` | Corporate entity | Main company entity |
| `LLC` | Limited Liability Company | Business subsidiary |
| `PARTNERSHIP` | Business partnership | Joint ventures |
| `NONPROFIT` | Non-profit organization | Charitable organizations |
| `GOVERNMENT` | Government entity | Public sector organizations |
| `DEPARTMENT` | Department within organization | HR, Finance, Engineering |
| `DIVISION` | Business division | Product lines, geographic regions |
| `TEAM` | Work team | Project teams, cross-functional teams |

### Pagination Parameters

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `limit` | integer | Number of items per page | 20 |
| `offset` | integer | Number of items to skip | 0 |

### Filter Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `organization_type` | string | Filter by organization type |
| `status` | string | Filter by organization status |
| `name` | string | Filter by organization name (partial match) |

### Example with Filters
```bash
curl -X GET "http://localhost:8080/api/v1/organizations?organization_type=DEPARTMENT&status=ACTIVE&limit=5" | jq .
```

## ️ Hierarchical Operations

### Building Organization Hierarchy
```bash
# 1. Create parent organization
CORP_ID=$(curl -s -X POST http://localhost:8080/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Tech Solutions Inc",
    "organization_type": "CORPORATION"
  }' | jq -r '.id')

# 2. Create department under corporation
DEPT_ID=$(curl -s -X POST http://localhost:8080/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Engineering Department",
    "organization_type": "DEPARTMENT",
    "parent_id": "'$CORP_ID'"
  }' | jq -r '.id')

# 3. Create team under department
TEAM_ID=$(curl -s -X POST http://localhost:8080/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Backend Development Team",
    "organization_type": "TEAM", 
    "parent_id": "'$DEPT_ID'"
  }' | jq -r '.id')

# 4. Get complete hierarchy
curl -X GET http://localhost:8080/api/v1/organizations/$CORP_ID/hierarchy | jq .
```

### Traversing Hierarchy
```bash
# Get organization with all descendants
curl -X GET http://localhost:8080/api/v1/organizations/$CORP_ID/hierarchy?depth=all | jq .

# Get organization with limited depth
curl -X GET http://localhost:8080/api/v1/organizations/$CORP_ID/hierarchy?depth=2 | jq .

# Get only direct children
curl -X GET http://localhost:8080/api/v1/organizations/$CORP_ID/hierarchy?depth=1 | jq .
```

##  Error Handling

### Common Error Responses

| Status Code | Error Type | Description |
|-------------|------------|-------------|
| `400` | `bad_request` | Invalid request payload or parameters |
| `404` | `not_found` | Organization not found |
| `409` | `conflict` | Organization name conflict or circular reference |
| `422` | `unprocessable_entity` | Invalid organization type or status |
| `500` | `internal_error` | Internal server error |

### Error Response Format
```json
{
  "error": "string",
  "message": "string",
  "details": "string (optional)"
}
```

### Specific Error Examples

#### Circular Reference Error
```json
{
  "error": "conflict",
  "message": "Cannot set parent: would create circular reference",
  "details": "Organization cannot be its own ancestor"
}
```

#### Invalid Organization Type
```json
{
  "error": "unprocessable_entity",
  "message": "Invalid organization type",
  "details": "Must be one of: CORPORATION, LLC, PARTNERSHIP, NONPROFIT, GOVERNMENT, DEPARTMENT, DIVISION, TEAM"
}
```

## ✅ Validation Rules

### Organization Creation
- **name**: Required, 1-255 characters, unique within parent scope
- **organization_type**: Required, valid enum value
- **parent_id**: Optional, must be valid UUID of existing organization
- **status**: Optional, defaults to "ACTIVE"

### Organization Update
- **name**: Optional, 1-255 characters, unique within parent scope if provided
- **organization_type**: Optional, valid enum value
- **parent_id**: Optional, must not create circular references
- **status**: Optional, valid status enum

### Hierarchy Constraints
- Organizations cannot be their own parent (direct circular reference)
- Organizations cannot be ancestors of themselves (indirect circular reference)
- Maximum hierarchy depth: 10 levels
- Archived organizations cannot have new children

##  Testing

### Basic CRUD Operations Test
```bash
echo "=== Organization CRUD Test ==="

# 1. Create organization
ORG_ID=$(curl -s -X POST http://localhost:8080/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Corporation",
    "organization_type": "CORPORATION"
  }' | jq -r '.id')

echo "Created organization: $ORG_ID"

# 2. Get organization
curl -X GET http://localhost:8080/api/v1/organizations/$ORG_ID | jq .

# 3. Update organization
curl -X PUT http://localhost:8080/api/v1/organizations/$ORG_ID \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Test Corporation",
    "organization_type": "LLC"
  }' | jq .

# 4. List organizations
curl -X GET http://localhost:8080/api/v1/organizations | jq .

# 5. Archive organization
curl -X PATCH http://localhost:8080/api/v1/organizations/$ORG_ID/archive \
  -H "Content-Type: application/json" \
  -d '{"reason": "Test completed"}'

echo "=== CRUD test completed ==="
```

### Hierarchy Operations Test
```bash
echo "=== Hierarchy Test ==="

# Create hierarchical structure
CORP_ID=$(curl -s -X POST http://localhost:8080/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{"name": "Hierarchy Test Corp", "organization_type": "CORPORATION"}' | jq -r '.id')

DEPT_ID=$(curl -s -X POST http://localhost:8080/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Department", "organization_type": "DEPARTMENT", "parent_id": "'$CORP_ID'"}' | jq -r '.id')

TEAM_ID=$(curl -s -X POST http://localhost:8080/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Team", "organization_type": "TEAM", "parent_id": "'$DEPT_ID'"}' | jq -r '.id')

# Get hierarchy
curl -X GET http://localhost:8080/api/v1/organizations/$CORP_ID/hierarchy | jq .

echo "=== Hierarchy test completed ==="
```

##  Performance Considerations

### Caching Strategy
- Organization data cached for 30 minutes
- Hierarchy data cached for 15 minutes
- Organization lists use database pagination

### Rate Limiting
- Standard rate limits apply (100 requests/minute per IP)
- Bulk operations may have lower limits

### Optimization Tips
- Use pagination for large organization lists
- Cache hierarchy lookups when possible
- Minimize deep hierarchy queries
- Use appropriate filters to reduce result sets

##  Related APIs

- **[User Management API](../users/README.md)** - Users belong to organizations
- **[Entity Management API](../entities/README.md)** - Entities are organization-scoped
- **[Tenant Management API](../tenants/README.md)** - Organizations belong to tenants

##  Additional Resources

- **[curl Examples](curl-examples.md)** - Complete curl command examples
- **[Testing Scripts](../../utilities/scripts/test-organizations.sh)** - Automated testing script
- **[Postman Collection](../../utilities/postman/organizations-collection.json)** - Postman requests

---

**Last Updated**: 2025-07-19  
**API Version**: 1.0.0  
**Maintained by**: AWO ERP System Team