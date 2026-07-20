> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

# Entity Schema Reference

## Overview
Complete schema definitions for entity-related data structures in the AWO ERP system.

## Core Schemas

### Entity
Base entity structure representing organizational units and business objects.

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "title": "Entity",
  "description": "A business entity or organizational unit",
  "required": ["name", "type"],
  "properties": {
    "id": {
      "type": "string",
      "format": "uuid",
      "description": "Unique entity identifier",
      "readOnly": true
    },
    "name": {
      "type": "string",
      "minLength": 1,
      "maxLength": 255,
      "description": "Entity display name"
    },
    "slug": {
      "type": "string",
      "pattern": "^[a-z0-9-]+$",
      "description": "URL-friendly identifier",
      "readOnly": true
    },
    "description": {
      "type": "string",
      "maxLength": 1000,
      "description": "Optional entity description"
    },
    "type": {
      "type": "string",
      "enum": ["organization", "department", "project", "team", "location"],
      "description": "Entity classification"
    },
    "parent_id": {
      "type": "string",
      "format": "uuid",
      "description": "Parent entity for hierarchical structures"
    },
    "tenant_id": {
      "type": "string",
      "format": "uuid",
      "description": "Tenant isolation identifier",
      "readOnly": true
    },
    "attributes": {
      "type": "object",
      "description": "Custom entity attributes",
      "additionalProperties": {
        "type": "string"
      }
    },
    "status": {
      "type": "string",
      "enum": ["active", "inactive", "archived"],
      "default": "active",
      "description": "Entity status"
    },
    "created_at": {
      "type": "string",
      "format": "date-time",
      "description": "Creation timestamp",
      "readOnly": true
    },
    "updated_at": {
      "type": "string",
      "format": "date-time",
      "description": "Last update timestamp",
      "readOnly": true
    }
  },
  "additionalProperties": false
}
```

### EntityList
Paginated list response for entity collections.

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "title": "EntityList",
  "description": "Paginated list of entities",
  "required": ["entities", "pagination"],
  "properties": {
    "entities": {
      "type": "array",
      "items": {
        "$ref": "#/definitions/EntitySummary"
      },
      "description": "Array of entity summaries"
    },
    "pagination": {
      "$ref": "#/definitions/PaginationMeta"
    }
  },
  "definitions": {
    "EntitySummary": {
      "type": "object",
      "properties": {
        "id": {"type": "string", "format": "uuid"},
        "name": {"type": "string"},
        "slug": {"type": "string"},
        "type": {"type": "string"},
        "parent_id": {"type": "string", "format": "uuid"},
        "status": {"type": "string"},
        "created_at": {"type": "string", "format": "date-time"}
      },
      "required": ["id", "name", "type"]
    },
    "PaginationMeta": {
      "type": "object",
      "properties": {
        "page": {"type": "integer", "minimum": 1},
        "limit": {"type": "integer", "minimum": 1, "maximum": 100},
        "total": {"type": "integer", "minimum": 0},
        "pages": {"type": "integer", "minimum": 0}
      },
      "required": ["page", "limit", "total", "pages"]
    }
  }
}
```

### EntityHierarchy
Hierarchical view of entity relationships.

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "title": "EntityHierarchy",
  "description": "Entity with hierarchical relationships",
  "allOf": [
    {"$ref": "#/definitions/Entity"},
    {
      "properties": {
        "children": {
          "type": "array",
          "items": {
            "$ref": "#/definitions/EntitySummary"
          },
          "description": "Child entities"
        },
        "parent": {
          "$ref": "#/definitions/EntitySummary",
          "description": "Parent entity details"
        },
        "depth": {
          "type": "integer",
          "minimum": 0,
          "description": "Hierarchy depth level"
        }
      }
    }
  ]
}
```

### EntityCreate
Schema for entity creation requests.

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "title": "EntityCreate",
  "description": "Schema for creating new entities",
  "required": ["name", "type"],
  "properties": {
    "name": {
      "type": "string",
      "minLength": 1,
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
      "enum": ["organization", "department", "project", "team", "location"],
      "description": "Entity classification"
    },
    "parent_id": {
      "type": "string",
      "format": "uuid",
      "description": "Parent entity for hierarchical structures"
    },
    "attributes": {
      "type": "object",
      "description": "Custom entity attributes",
      "additionalProperties": {
        "type": "string"
      }
    }
  },
  "additionalProperties": false
}
```

### EntityUpdate
Schema for entity update requests.

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "title": "EntityUpdate",
  "description": "Schema for updating existing entities",
  "properties": {
    "name": {
      "type": "string",
      "minLength": 1,
      "maxLength": 255,
      "description": "Entity display name"
    },
    "description": {
      "type": "string",
      "maxLength": 1000,
      "description": "Entity description"
    },
    "type": {
      "type": "string",
      "enum": ["organization", "department", "project", "team", "location"],
      "description": "Entity classification"
    },
    "parent_id": {
      "type": "string",
      "format": "uuid",
      "description": "Parent entity for hierarchical structures"
    },
    "attributes": {
      "type": "object",
      "description": "Custom entity attributes",
      "additionalProperties": {
        "type": "string"
      }
    },
    "status": {
      "type": "string",
      "enum": ["active", "inactive", "archived"],
      "description": "Entity status"
    }
  },
  "additionalProperties": false,
  "minProperties": 1
}
```

## Validation Rules

### Business Logic Constraints
1. **Unique Names**: Entity names must be unique within the same parent scope
2. **Hierarchy Depth**: Maximum depth of 10 levels to prevent infinite nesting
3. **Circular References**: Entities cannot be their own ancestors
4. **Type Restrictions**: Some entity types cannot be parents of others:
   - `project` entities cannot have child entities
   - `team` entities can only belong to `department` or `project` entities

### Attribute Validation
- Attribute keys must match pattern: `^[a-z][a-z0-9_]*$`
- Attribute values are limited to 500 characters
- Maximum of 50 custom attributes per entity
- Reserved attribute keys: `id`, `created_at`, `updated_at`, `tenant_id`

## Examples

### Valid Entity Creation
```json
{
  "name": "Engineering Department",
  "description": "Software development and infrastructure team",
  "type": "department",
  "parent_id": "550e8400-e29b-41d4-a716-446655440000",
  "attributes": {
    "budget_code": "ENG-001",
    "cost_center": "CC-100",
    "manager_email": "john.doe@company.com"
  }
}
```

### Entity Hierarchy Response
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440001",
  "name": "Engineering Department",
  "type": "department",
  "depth": 1,
  "parent": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "ACME Corporation",
    "type": "organization"
  },
  "children": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440002",
      "name": "Frontend Team",
      "type": "team"
    },
    {
      "id": "550e8400-e29b-41d4-a716-446655440003",
      "name": "Backend Team",
      "type": "team"
    }
  ]
}
```
