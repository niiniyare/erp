> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

# Attribute Definitions API

## Overview

The Attribute Definitions API provides standardized attribute schemas for consistent ABAC evaluation across the organization. This service manages attribute metadata, validation rules, and type definitions.

**Goa Service**: `attribute-definitions`  
**Base Path**: `/api/v1/attributes`

## Business Value

- **Schema Standardization**: Ensure consistent attribute definitions across all systems
- **Data Validation**: Prevent policy evaluation errors with strict type checking
- **Governance**: Centralized control over attribute lifecycle and usage
- **Performance**: Optimized attribute validation and caching

## API Endpoints

### Create Attribute Definition

**Action**: `CreateDefinition`

```yaml
POST /definitions
Content-Type: application/json
Authorization: Bearer <jwt_token>
X-Tenant-ID: <tenant_uuid>

Request Body:
{
  "name": "user.security_clearance",
  "display_name": "Security Clearance Level",
  "description": "User's security clearance level for classified information access",
  "data_type": "ENUM",
  "category": "USER",
  "is_required": true,
  "is_sensitive": true,
  "allowed_values": [
    "UNCLASSIFIED",
    "CONFIDENTIAL", 
    "SECRET",
    "TOP_SECRET"
  ],
  "validation_rules": {
    "enum_strict": true,
    "case_sensitive": false
  },
  "encryption_required": false,
  "default_value": "UNCLASSIFIED"
}

Response: 201 Created
{
  "id": "attr_550e8400-e29b-41d4-a716-446655440000",
  "name": "user.security_clearance",
  "display_name": "Security Clearance Level",
  "category": "USER",
  "data_type": "ENUM",
  "is_active": true,
  "created_at": "2025-01-15T10:30:00Z",
  "updated_at": "2025-01-15T10:30:00Z"
}
```

### List Attribute Definitions

**Action**: `ListDefinitions`

```yaml
GET /definitions?category=USER&is_active=true&page=1&limit=50
Authorization: Bearer <jwt_token>
X-Tenant-ID: <tenant_uuid>

Query Parameters:
- category: USER, RESOURCE, ENVIRONMENT, ACTION, ENTITY, SESSION
- data_type: STRING, NUMBER, BOOLEAN, DATE, TIME, JSON, ARRAY, ENUM
- is_active: boolean
- is_required: boolean
- is_sensitive: boolean
- search: string (searches name and display_name)
- page: integer (default: 1)
- limit: integer (default: 20, max: 100)

Response: 200 OK
{
  "definitions": [
    {
      "id": "attr_550e8400-e29b-41d4-a716-446655440000",
      "name": "user.security_clearance",
      "display_name": "Security Clearance Level",
      "category": "USER",
      "data_type": "ENUM",
      "is_required": true,
      "is_sensitive": true,
      "is_active": true,
      "usage_count": 12,
      "created_at": "2025-01-15T10:30:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 50,
    "total": 1,
    "total_pages": 1,
    "has_next": false,
    "has_previous": false
  },
  "meta": {
    "total_active": 89,
    "total_inactive": 12,
    "by_category": {
      "USER": 34,
      "RESOURCE": 28,
      "ENVIRONMENT": 15,
      "SESSION": 12
    }
  }
}
```

### Get Attribute Definition

**Action**: `GetDefinition`

```yaml
GET /definitions/{definition_id}
Authorization: Bearer <jwt_token>
X-Tenant-ID: <tenant_uuid>

Path Parameters:
- definition_id: UUID (required)

Query Parameters:
- include_usage_stats: boolean (default: false)
- include_related: boolean (default: false)

Response: 200 OK
{
  "id": "attr_550e8400-e29b-41d4-a716-446655440000",
  "name": "user.security_clearance",
  "display_name": "Security Clearance Level",
  "description": "User's security clearance level for classified information access",
  "data_type": "ENUM",
  "category": "USER",
  "is_required": true,
  "is_sensitive": true,
  "allowed_values": ["UNCLASSIFIED", "CONFIDENTIAL", "SECRET", "TOP_SECRET"],
  "validation_rules": {
    "enum_strict": true,
    "case_sensitive": false
  },
  "default_value": "UNCLASSIFIED",
  "usage_stats": {
    "policies_using": 12,
    "evaluations_last_30d": 45678,
    "last_evaluated": "2025-01-15T10:45:00Z",
    "most_common_values": [
      {"value": "SECRET", "count": 234, "percentage": 45.2},
      {"value": "CONFIDENTIAL", "count": 156, "percentage": 30.1}
    ]
  },
  "related_attributes": [
    {
      "id": "attr_550e8400-e29b-41d4-a716-446655440001",
      "name": "user.security_level",
      "relationship": "complementary"
    }
  ],
  "is_active": true,
  "created_at": "2025-01-15T10:30:00Z",
  "updated_at": "2025-01-15T10:30:00Z",
  "created_by": "admin@company.com"
}
```

### Update Attribute Definition

**Action**: `UpdateDefinition`

```yaml
PUT /definitions/{definition_id}
Content-Type: application/json
Authorization: Bearer <jwt_token>
X-Tenant-ID: <tenant_uuid>

Request Body:
{
  "display_name": "Updated Security Clearance Level",
  "description": "Updated description",
  "allowed_values": [
    "UNCLASSIFIED",
    "CONFIDENTIAL",
    "SECRET", 
    "TOP_SECRET",
    "ABOVE_TOP_SECRET"
  ],
  "validation_rules": {
    "enum_strict": true,
    "case_sensitive": false,
    "allow_custom": false
  }
}

Response: 200 OK
{
  "id": "attr_550e8400-e29b-41d4-a716-446655440000",
  "name": "user.security_clearance",
  "display_name": "Updated Security Clearance Level",
  "version": 2,
  "updated_at": "2025-01-15T11:00:00Z",
  "changes": [
    "display_name updated",
    "allowed_values extended",
    "validation_rules modified"
  ],
  "impact_assessment": {
    "policies_affected": 12,
    "cached_evaluations_invalidated": 1547,
    "backward_compatible": true
  }
}
```

### Validate Attribute Value

**Action**: `ValidateValue`

```yaml
POST /definitions/{definition_id}/validate
Content-Type: application/json
Authorization: Bearer <jwt_token>
X-Tenant-ID: <tenant_uuid>

Request Body:
{
  "value": "SECRET",
  "context": {
    "source": "hr_system",
    "confidence": 1.0
  }
}

Response: 200 OK
{
  "valid": true,
  "normalized_value": "SECRET",
  "validation_details": {
    "type_check": {
      "status": "passed",
      "expected_type": "string",
      "actual_type": "string"
    },
    "enum_check": {
      "status": "passed",
      "allowed_values": ["UNCLASSIFIED", "CONFIDENTIAL", "SECRET", "TOP_SECRET"],
      "case_normalized": false
    },
    "format_check": {
      "status": "passed",
      "rules_applied": ["enum_strict", "case_sensitive"]
    },
    "business_rules": {
      "status": "passed",
      "checks": []
    }
  },
  "warnings": [],
  "suggestions": []
}
```

### Bulk Validate Attributes

**Action**: `BulkValidate`

```yaml
POST /definitions/validate-bulk
Content-Type: application/json
Authorization: Bearer <jwt_token>
X-Tenant-ID: <tenant_uuid>

Request Body:
{
  "validations": [
    {
      "definition_id": "attr_550e8400-e29b-41d4-a716-446655440000",
      "value": "SECRET"
    },
    {
      "definition_name": "user.department",
      "value": "finance"
    },
    {
      "definition_id": "attr_550e8400-e29b-41d4-a716-446655440002",
      "value": 7
    }
  ],
  "options": {
    "fail_fast": false,
    "normalize_values": true
  }
}

Response: 200 OK
{
  "results": [
    {
      "definition_id": "attr_550e8400-e29b-41d4-a716-446655440000",
      "valid": true,
      "normalized_value": "SECRET"
    },
    {
      "definition_name": "user.department",
      "valid": true,
      "normalized_value": "finance",
      "definition_id": "attr_550e8400-e29b-41d4-a716-446655440003"
    },
    {
      "definition_id": "attr_550e8400-e29b-41d4-a716-446655440002",
      "valid": false,
      "error": "Value exceeds maximum allowed: 5",
      "error_code": "VALUE_OUT_OF_RANGE"
    }
  ],
  "summary": {
    "total": 3,
    "valid": 2,
    "invalid": 1,
    "warnings": 0
  }
}
```

### Delete Attribute Definition

**Action**: `DeleteDefinition`

```yaml
DELETE /definitions/{definition_id}
Authorization: Bearer <jwt_token>
X-Tenant-ID: <tenant_uuid>

Query Parameters:
- force: boolean (default: false) - bypass usage checks
- cascade: boolean (default: false) - remove from policies

Response: 204 No Content

# If definition is in use and force=false:
Response: 409 Conflict
{
  "error": "DEFINITION_IN_USE",
  "message": "Cannot delete attribute definition: currently in use",
  "details": {
    "policies_using": 12,
    "evaluations_last_30d": 45678,
    "last_used": "2025-01-15T10:45:00Z"
  },
  "suggestions": [
    "Set force=true to override usage check",
    "Deactivate instead of deleting",
    "Remove from policies first"
  ]
}
```

### Attribute Usage Analytics

**Action**: `GetUsageAnalytics`

```yaml
GET /definitions/{definition_id}/analytics?period=30d
Authorization: Bearer <jwt_token>
X-Tenant-ID: <tenant_uuid>

Query Parameters:
- period: 1d, 7d, 30d, 90d (default: 30d)
- include_trends: boolean (default: true)
- include_values: boolean (default: true)

Response: 200 OK
{
  "definition_id": "attr_550e8400-e29b-41d4-a716-446655440000",
  "period": "30d",
  "usage_summary": {
    "total_evaluations": 45678,
    "unique_subjects": 1234,
    "policies_using": 12,
    "avg_evaluations_per_day": 1522
  },
  "value_distribution": [
    {
      "value": "SECRET",
      "count": 20654,
      "percentage": 45.2,
      "trend": "increasing"
    },
    {
      "value": "CONFIDENTIAL", 
      "count": 15234,
      "percentage": 33.3,
      "trend": "stable"
    },
    {
      "value": "TOP_SECRET",
      "count": 6543,
      "percentage": 14.3,
      "trend": "decreasing"
    },
    {
      "value": "UNCLASSIFIED",
      "count": 3247,
      "percentage": 7.1,
      "trend": "stable"
    }
  ],
  "usage_trends": {
    "daily_evaluations": [
      {"date": "2025-01-15", "count": 1678},
      {"date": "2025-01-14", "count": 1543}
    ],
    "growth_rate": "5.2%",
    "peak_usage_hours": [14, 15, 16]
  },
  "performance_metrics": {
    "avg_validation_time_ms": 2.3,
    "cache_hit_rate": 0.94,
    "error_rate": 0.001
  }
}
```

## Goa DSL Implementation Notes

### Service Definition

```go
var _ = Service("attribute-definitions", func() {
    Description("Attribute Definitions Management Service")
    
    HTTP(func() {
        Path("/api/v1/attributes")
        Header("X-Tenant-ID", String, "Tenant identifier", func() {
            Example("550e8400-e29b-41d4-a716-446655440000")
        })
    })
    
    JWT(func() {
        Description("JWT authentication")
        Scope("attribute:read", "Read attribute definitions")
        Scope("attribute:write", "Create/update attribute definitions")
        Scope("attribute:delete", "Delete attribute definitions")
    })
})
```

### Type Definitions

```go
var AttributeDefinition = Type("AttributeDefinition", func() {
    Attribute("id", String, "Unique identifier", func() {
        Format(FormatUUID)
        Example("550e8400-e29b-41d4-a716-446655440000")
    })
    Attribute("name", String, "Attribute name", func() {
        Pattern("^[a-z][a-z0-9_]*\\.[a-z][a-z0-9_]*$")
        Example("user.security_clearance")
    })
    Attribute("display_name", String, "Human-readable name")
    Attribute("description", String, "Detailed description")
    Attribute("data_type", String, "Data type", func() {
        Enum("STRING", "NUMBER", "BOOLEAN", "DATE", "TIME", "JSON", "ARRAY", "ENUM")
    })
    Attribute("category", String, "Attribute category", func() {
        Enum("USER", "RESOURCE", "ENVIRONMENT", "ACTION", "ENTITY", "SESSION")
    })
    Attribute("is_required", Boolean, "Whether attribute is required")
    Attribute("is_sensitive", Boolean, "Whether attribute contains sensitive data")
    Attribute("allowed_values", ArrayOf(String), "Allowed values for ENUM type")
    Attribute("validation_rules", Any, "Validation rules as JSON")
    Attribute("is_active", Boolean, "Whether definition is active")
    Attribute("created_at", String, "Creation timestamp", func() {
        Format(FormatDateTime)
    })
    Attribute("updated_at", String, "Last update timestamp", func() {
        Format(FormatDateTime)
    })
    
    Required("id", "name", "data_type", "category", "is_active")
})
```

### Error Types

```go
var AttributeDefinitionError = Type("AttributeDefinitionError", func() {
    Attribute("error", String, "Error code")
    Attribute("message", String, "Error message")
    Attribute("details", Any, "Error details")
    Attribute("suggestions", ArrayOf(String), "Suggested actions")
    
    Required("error", "message")
})
```

### Validation Middleware

- **Tenant Isolation**: Ensure all operations are scoped to the tenant
- **Schema Validation**: Validate attribute names follow naming conventions
- **Usage Checking**: Prevent deletion of attributes in active use
- **Performance Monitoring**: Track validation performance and cache metrics

### Caching Strategy

- **Definition Cache**: Cache attribute definitions with 1-hour TTL
- **Validation Cache**: Cache validation results with 15-minute TTL
- **Usage Stats Cache**: Cache analytics data with 5-minute TTL
- **Invalidation**: Implement cache invalidation on definition updates




<!-- ##  1. Attribute Management API ->
<!-- ### 1.1 Attribute Definitions ->
<!-- Business Value: Standardize attribute schemas across the organization for consistent policy evaluation. -->
<!---->
<!-- ```json -->
<!-- # Attribute Management API ->
<!---->
<!-- ## Create Attribute Definition ->
<!-- POST /attributes/definitions -->
<!-- Content-Type: application/json -->
<!---->
<!-- Request Body: -->
<!-- { -->
<!--   "name": "user.security_clearance", -->
<!--   "display_name": "Security Clearance Level", -->
<!--   "description": "User's security clearance level for classified information access", -->
<!--   "data_type": "ENUM", -->
<!--   "category": "USER", -->
<!--   "is_required": true, -->
<!--   "is_sensitive": true, -->
<!--   "allowed_values": [ -->
<!--     "UNCLASSIFIED", -->
<!--     "CONFIDENTIAL",  -->
<!--     "SECRET", -->
<!--     "TOP_SECRET" -->
<!--   ], -->
<!--   "validation_rules": { -->
<!--     "enum_strict": true, -->
<!--     "case_sensitive": false -->
<!--   }, -->
<!--   "encryption_required": false, -->
<!--   "default_value": "UNCLASSIFIED" -->
<!-- } -->
<!---->
<!-- Response: 201 Created -->
<!-- { -->
<!--   "id": "550e8400-e29b-41d4-a716-446655440000", -->
<!--   "name": "user.security_clearance", -->
<!--   "display_name": "Security Clearance Level", -->
<!--   "category": "USER", -->
<!--   "data_type": "ENUM", -->
<!--   "is_active": true, -->
<!--   "created_at": "2025-01-15T10:30:00Z", -->
<!--   "updated_at": "2025-01-15T10:30:00Z" -->
<!-- } -->
<!---->
<!-- ## List Attribute Definitions ->
<!-- GET /attributes/definitions?category=USER&is_active=true&page=1&limit=50 -->
<!---->
<!-- Response: 200 OK -->
<!-- { -->
<!--   "attributes": [ -->
<!--     { -->
<!--       "id": "550e8400-e29b-41d4-a716-446655440000", -->
<!--       "name": "user.security_clearance", -->
<!--       "display_name": "Security Clearance Level", -->
<!--       "category": "USER", -->
<!--       "data_type": "ENUM", -->
<!--       "is_required": true, -->
<!--       "is_sensitive": true, -->
<!--       "is_active": true -->
<!--     } -->
<!--   ], -->
<!--   "pagination": { -->
<!--     "page": 1, -->
<!--     "limit": 50, -->
<!--     "total": 1, -->
<!--     "total_pages": 1 -->
<!--   } -->
<!-- } -->
<!---->
<!-- ## Update Attribute Definition ->
<!-- PUT /attributes/definitions/{attribute_id} -->
<!-- Content-Type: application/json -->
<!---->
<!-- Request Body: -->
<!-- { -->
<!--   "display_name": "Updated Security Clearance Level", -->
<!--   "description": "Updated description", -->
<!--   "allowed_values": [ -->
<!--     "UNCLASSIFIED", -->
<!--     "CONFIDENTIAL", -->
<!--     "SECRET",  -->
<!--     "TOP_SECRET", -->
<!--     "ABOVE_TOP_SECRET" -->
<!--   ] -->
<!-- } -->
<!---->
<!-- Response: 200 OK -->
<!-- { -->
<!--   "id": "550e8400-e29b-41d4-a716-446655440000", -->
<!--   "name": "user.security_clearance", -->
<!--   "display_name": "Updated Security Clearance Level", -->
<!--   "updated_at": "2025-01-15T11:00:00Z" -->
<!-- } -->
<!---->
<!-- ## Get Attribute Definition ->
<!-- GET /attributes/definitions/{attribute_id} -->
<!---->
<!-- Response: 200 OK -->
<!-- { -->
<!--   "id": "550e8400-e29b-41d4-a716-446655440000", -->
<!--   "name": "user.security_clearance", -->
<!--   "display_name": "Security Clearance Level", -->
<!--   "description": "User's security clearance level for classified information access", -->
<!--   "data_type": "ENUM", -->
<!--   "category": "USER", -->
<!--   "is_required": true, -->
<!--   "is_sensitive": true, -->
<!--   "allowed_values": ["UNCLASSIFIED", "CONFIDENTIAL", "SECRET", "TOP_SECRET"], -->
<!--   "validation_rules": { -->
<!--     "enum_strict": true, -->
<!--     "case_sensitive": false -->
<!--   }, -->
<!--   "usage_stats": { -->
<!--     "policies_using": 12, -->
<!--     "last_evaluated": "2025-01-15T10:45:00Z" -->
<!--   }, -->
<!--   "created_at": "2025-01-15T10:30:00Z", -->
<!--   "updated_at": "2025-01-15T10:30:00Z" -->
<!-- } -->
<!---->
<!-- ## Delete Attribute Definition ->
<!-- DELETE /attributes/definitions/{attribute_id} -->
<!---->
<!-- Response: 204 No Content -->
<!---->
<!-- ## Validate Attribute Value ->
<!-- POST /attributes/definitions/{attribute_id}/validate -->
<!-- Content-Type: application/json -->
<!---->
<!-- Request Body: -->
<!-- { -->
<!--   "value": "SECRET" -->
<!-- } -->
<!---->
<!-- Response: 200 OK -->
<!-- { -->
<!--   "valid": true, -->
<!--   "normalized_value": "SECRET", -->
<!--   "validation_details": { -->
<!--     "type_check": "passed", -->
<!--     "enum_check": "passed", -->
<!--     "format_check": "passed" -->
<!--   } -->
<!-- } -->
<!-- ``` -->
<!---->
<!---->
<!-- ### 1.2 Attribute Values Management ->
<!-- *Business Value*: Manage actual attribute values for subjects, resources, and environments with real-time updates. -->
