# Settings Module - API Reference

**Version**: 1.0  
**Date**: September 2025  
**Status**: Development  
**OpenAPI Version**: 3.0.3

---

## Overview

### API Description
The Settings Module API provides  configuration management capabilities within the AWO ERP system. This API supports three-level configuration inheritance (System → Tenant → Entity), template-based configuration deployment, and enterprise-grade bulk operations for efficient configuration management.

### Base Information
- **Base URL**: `https://api.awo-erp.com/api/v1/settings`
- **Authentication**: Bearer Token (JWT)
- **Content Type**: `application/json`
- **API Version**: `v1`

### Quick Links
- [Interactive API Explorer](../../../reference/api/swagger-ui.md) - Test endpoints directly
- [Authentication Guide](../../../reference/api/auth/index.md) - Get started with API authentication
- SDK Examples - Code examples in multiple languages

---

## Table of Contents

1. [Overview](#overview)
2. [Authentication](#authentication)
3. [Core Endpoints](#core-endpoints)
   - [Configuration Management](#configuration-management)
   - [Template Management](#template-management)
   - [Bulk Operations](#bulk-operations)
4. [Resolution Endpoints](#resolution-endpoints)
5. [Search Endpoints](#search-endpoints)
6. [Data Models](#data-models)
7. [Error Handling](#error-handling)
8. [Rate Limiting](#rate-limiting)
9. [Code Examples](#code-examples)
10. [Testing](#testing)
11. [Changelog](#changelog)

---

## Authentication

### Bearer Token Authentication
All API endpoints require authentication using JWT bearer tokens with proper configuration permissions.

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

### Configuration Permissions
The Settings API uses fine-grained permissions for configuration access:

**Permission Structure**: `settings:{scope}:{operation}`
- `settings:system:read` - View system-level configurations  
- `settings:tenant:write` - Modify tenant-level configurations
- `settings:entity:write` - Modify entity-level configurations
- `settings:templates:apply` - Apply configuration templates
- `settings:bulk:write` - Perform bulk configuration operations

---

## Core Endpoints

### Configuration Management

#### Get Configuration Value
Resolve a configuration value through the inheritance hierarchy.

**Endpoint**: `GET /api/v1/settings/config/{module}/{config_key}`

#### #### **Path Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `module` | string | Yes | Module name (finance, hr, inventory) |
| `config_key` | string | Yes | Configuration key |

#### **Query Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `entity_id` | UUID | No | Entity context for resolution |
| `include_metadata` | boolean | No | Include inheritance metadata (default: false) |

**Response** (200 OK):
```json
{
  "module": "finance",
  "config_key": "invoice_prefix",
  "value": "INV-2025-",
  "data_type": "string",
  "source": "entity",
  "is_inherited": false,
  "can_override": true,
  "inheritance_chain": {
    "system": "DOC-",
    "tenant": "INV-",
    "entity": "INV-2025-"
  },
  "metadata": {
    "last_modified": "2025-09-01T10:30:00Z",
    "modified_by": "user-456",
    "required_permission": "finance.admin",
    "feature_enabled": true
  }
}
```

#### Set Configuration Value
Update a configuration value at the tenant or entity level.

**Endpoint**: `PUT /api/v1/settings/config/{module}/{config_key}`

#### **Path Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `module` | string | Yes | Module name |
| `config_key` | string | Yes | Configuration key |

**Request Body**:
```json
{
  "value": "BRANCH-INV-2025-",
  "target_type": "entity",
  "target_id": "entity-uuid",
  "data_type": "string",
  "override_existing": true
}
```

**Response** (200 OK):
```json
{
  "module": "finance",
  "config_key": "invoice_prefix", 
  "value": "BRANCH-INV-2025-",
  "source": "entity",
  "previous_value": "INV-",
  "previous_source": "tenant",
  "updated_at": "2025-09-01T14:30:00Z",
  "version": 2
}
```

#### List Module Configurations
Get all configurations for a specific module with inheritance resolution.

**Endpoint**: `GET /api/v1/settings/modules/{module}/config`

**Path Parameqters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `module` | string | Yes | Module name |

#### **Query Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `entity_id` | UUID | No | Entity context for resolution |
| `source` | string | No | Filter by source (system, tenant, entity) |
| `include_inherited` | boolean | No | Include inherited values (default: true) |
| `include_defaults` | boolean | No | Include system defaults (default: true) |

**Response** (200 OK):
```json
{
  "module": "finance",
  "entity_id": "entity-uuid",
  "configurations": [
    {
      "config_key": "invoice_prefix",
      "value": "BRANCH-INV-2025-",
      "data_type": "string",
      "source": "entity",
      "is_inherited": false
    },
    {
      "config_key": "auto_approval_limit",
      "value": 1000,
      "data_type": "integer",
      "source": "tenant",
      "is_inherited": true
    },
    {
      "config_key": "default_currency",
      "value": "USD",
      "data_type": "string", 
      "source": "system",
      "is_inherited": true
    }
  ],
  "summary": {
    "total_configs": 25,
    "entity_overrides": 3,
    "tenant_configs": 15,
    "system_defaults": 7
  }
}
```

#### Delete Configuration Override
Reset a configuration to its inherited value.

**Endpoint**: `DELETE /api/v1/settings/config/{module}/{config_key}`

#### **Path Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `module` | string | Yes | Module name |
| `config_key` | string | Yes | Configuration key |

#### **Query Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `target_type` | string | Yes | Target type (tenant, entity) |
| `target_id` | UUID | Yes | Target identifier |

**Response** (200 OK):
```json
{
  "module": "finance",
  "config_key": "invoice_prefix",
  "action": "reset_to_inherited",
  "previous_value": "BRANCH-INV-2025-",
  "inherited_value": "INV-",
  "inherited_source": "tenant",
  "reset_at": "2025-09-01T15:00:00Z"
}
```

### Template Management

#### List Configuration Templates
Get available configuration templates with filtering.

**Endpoint**: `GET /api/v1/settings/templates`

#### **Query Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `category` | string | No | Filter by category (industry, functional, regional) |
| `applicable_to` | string | No | Filter by tenant type |
| `search` | string | No | Search in name and description |

**Response** (200 OK):
```json
{
  "templates": [
    {
      "id": "template-uuid",
      "name": "Manufacturing Standard",
      "category": "industry",
      "description": "Standard configuration for manufacturing companies",
      "version": "1.2",
      "applicable_tenant_types": ["manufacturing", "industrial"],
      "required_feature_flags": ["inventory.lot_tracking"],
      "config_count": 45,
      "created_at": "2025-08-01T10:00:00Z",
      "last_updated": "2025-08-15T14:30:00Z",
      "is_active": true
    }
  ],
  "pagination": {
    "total": 15,
    "page": 1,
    "limit": 20,
    "has_more": false
  }
}
```

#### Get Template Details
Retrieve detailed information about a specific template.

**Endpoint**: `GET /api/v1/settings/templates/{template_id}`

#### **Path Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `template_id` | UUID | Yes | Template identifier |

**Response** (200 OK):
```json
{
  "id": "template-uuid",
  "name": "Manufacturing Standard",
  "category": "industry",
  "description": "Standard configuration for manufacturing companies",
  "version": "1.2",
  "configurations": [
    {
      "module": "finance",
      "config_key": "invoice_prefix",
      "value": "MFG-INV-",
      "data_type": "string",
      "override_policy": "replace",
      "priority": 100
    },
    {
      "module": "inventory",
      "config_key": "valuation_method", 
      "value": "FIFO",
      "data_type": "string",
      "override_policy": "preserve",
      "priority": 200
    }
  ],
  "dependencies": [
    {
      "template_id": "base-erp-uuid",
      "required_version": ">=1.0"
    }
  ],
  "applicable_tenant_types": ["manufacturing", "industrial"],
  "required_feature_flags": ["inventory.lot_tracking"],
  "conflict_resolution": "merge",
  "created_by": "system",
  "created_at": "2025-08-01T10:00:00Z"
}
```

#### Apply Configuration Template
Apply a template to a tenant or entity with conflict resolution.

**Endpoint**: `POST /api/v1/settings/templates/{template_id}/apply`

#### **Path Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `template_id` | UUID | Yes | Template identifier |

**Request Body**:
```json
{
  "target_type": "tenant",
  "target_id": "tenant-uuid",
  "options": {
    "preserve_existing": true,
    "conflict_resolution": "prompt",
    "apply_dependencies": true,
    "dry_run": false
  },
  "selective_application": {
    "modules": ["finance", "inventory"],
    "exclude_keys": ["finance.bank_account"]
  }
}
```

**Response** (200 OK):
```json
{
  "operation_id": "op-uuid-123",
  "template_id": "template-uuid",
  "target_type": "tenant",
  "target_id": "tenant-uuid",
  "status": "completed",
  "summary": {
    "total_configs": 45,
    "applied": 40,
    "skipped": 3,
    "conflicts": 2
  },
  "applied_configurations": [
    {
      "module": "finance",
      "config_key": "invoice_prefix",
      "old_value": "INV-",
      "new_value": "MFG-INV-",
      "action": "updated"
    }
  ],
  "conflicts": [
    {
      "module": "inventory",
      "config_key": "reorder_threshold", 
      "template_value": 10,
      "existing_value": 15,
      "resolution": "kept_existing",
      "reason": "preserve_existing_policy"
    }
  ],
  "applied_at": "2025-09-01T16:00:00Z",
  "duration_ms": 1250
}
```

### Bulk Operations

#### Bulk Configuration Update
Update multiple configurations across multiple entities efficiently.

**Endpoint**: `POST /api/v1/settings/bulk-update`

**Request Body**:
```json
{
  "targets": {
    "type": "entity",
    "filters": {
      "entity_types": ["BRANCH", "LOCATION"],
      "tags": ["retail"],
      "created_after": "2025-01-01"
    },
    "explicit_ids": ["entity-1", "entity-2"]
  },
  "configurations": [
    {
      "module": "finance",
      "config_key": "approval_limit",
      "value": 2500,
      "data_type": "integer"
    },
    {
      "module": "inventory",
      "config_key": "reorder_enabled",
      "value": true,
      "data_type": "boolean"
    }
  ],
  "options": {
    "preserve_explicit_overrides": true,
    "batch_size": 100,
    "dry_run": false,
    "continue_on_error": true
  }
}
```

**Response** (202 Accepted):
```json
{
  "operation_id": "bulk-op-789",
  "status": "in_progress",
  "targets": {
    "total_entities": 156,
    "batches": 2
  },
  "configurations": 2,
  "estimated_duration": "45 seconds",
  "progress_url": "/api/v1/settings/operations/bulk-op-789",
  "started_at": "2025-09-01T17:00:00Z"
}
```

#### Get Bulk Operation Status
Check the status and progress of a bulk operation.

**Endpoint**: `GET /api/v1/settings/operations/{operation_id}`

#### **Path Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `operation_id` | string | Yes | Operation identifier |

**Response** (200 OK):
```json
{
  "operation_id": "bulk-op-789",
  "status": "completed",
  "progress": {
    "total_targets": 156,
    "processed": 156,
    "successful": 154,
    "failed": 2,
    "percentage": 100
  },
  "summary": {
    "total_configurations": 312,
    "updated": 308,
    "skipped": 2,
    "errors": 2
  },
  "results": [
    {
      "target_id": "entity-1",
      "status": "success",
      "configurations_updated": 2
    },
    {
      "target_id": "entity-failed",
      "status": "error",
      "error": "Permission denied for configuration finance.approval_limit"
    }
  ],
  "started_at": "2025-09-01T17:00:00Z",
  "completed_at": "2025-09-01T17:01:15Z",
  "duration_ms": 75000
}
```

---

## Resolution Endpoints

### Resolve Configuration for Context
Get resolved configuration value for a specific context (tenant + entity).

**Endpoint**: `POST /api/v1/settings/resolve`

**Request Body**:
```json
{
  "context": {
    "tenant_id": "tenant-uuid",
    "entity_id": "entity-uuid"
  },
  "configurations": [
    {
      "module": "finance",
      "config_key": "invoice_prefix"
    },
    {
      "module": "inventory", 
      "config_key": "valuation_method"
    }
  ],
  "options": {
    "include_inheritance_chain": true,
    "include_metadata": true,
    "include_permissions": false
  }
}
```

**Response** (200 OK):
```json
{
  "context": {
    "tenant_id": "tenant-uuid",
    "entity_id": "entity-uuid"
  },
  "resolved_configurations": [
    {
      "module": "finance",
      "config_key": "invoice_prefix",
      "resolved_value": "BRANCH-INV-",
      "data_type": "string",
      "source": "entity",
      "inheritance_chain": {
        "system": "DOC-",
        "tenant": "INV-", 
        "entity": "BRANCH-INV-"
      },
      "can_override": true
    },
    {
      "module": "inventory",
      "config_key": "valuation_method", 
      "resolved_value": "FIFO",
      "data_type": "string",
      "source": "tenant",
      "inheritance_chain": {
        "system": "FIFO",
        "tenant": "FIFO"
      },
      "can_override": true
    }
  ],
  "resolved_at": "2025-09-01T18:00:00Z"
}
```

### Get Effective Configuration Set
Retrieve the complete effective configuration set for an entity.

**Endpoint**: `GET /api/v1/settings/effective`

#### **Query Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `entity_id` | UUID | Yes | Entity identifier |
| `modules` | string[] | No | Filter by modules |
| `changed_since` | datetime | No | Only configs changed since date |

**Response** (200 OK):
```json
{
  "entity_id": "entity-uuid",
  "effective_configuration": {
    "finance": {
      "invoice_prefix": "BRANCH-INV-",
      "auto_approval_limit": 1000,
      "default_currency": "USD",
      "payment_terms": "NET30"
    },
    "inventory": {
      "valuation_method": "FIFO",
      "reorder_enabled": true,
      "negative_stock_allowed": false
    },
    "hr": {
      "overtime_threshold": 40,
      "pay_frequency": "biweekly"
    }
  },
  "configuration_sources": {
    "finance.invoice_prefix": "entity",
    "finance.auto_approval_limit": "tenant",
    "finance.default_currency": "system"
  },
  "generated_at": "2025-09-01T18:30:00Z",
  "cache_expires_at": "2025-09-01T18:35:00Z"
}
```

---

## Search Endpoints

### Search Configurations
Perform advanced search across configurations with multiple criteria.

**Endpoint**: `POST /api/v1/settings/search`

**Request Body**:
```json
{
  "criteria": {
    "modules": ["finance", "inventory"],
    "config_keys": ["*_prefix", "approval_*"],
    "value_contains": "INV",
    "sources": ["entity", "tenant"],
    "data_types": ["string"],
    "modified_after": "2025-08-01T00:00:00Z"
  },
  "scope": {
    "tenant_id": "tenant-uuid",
    "entity_ids": ["entity-1", "entity-2"]
  },
  "sort": [
    {
      "field": "last_modified",
      "order": "desc"
    },
    {
      "field": "module",
      "order": "asc"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 50
  }
}
```

**Response** (200 OK):
```json
{
  "results": [
    {
      "module": "finance",
      "config_key": "invoice_prefix",
      "value": "BRANCH-INV-",
      "data_type": "string",
      "source": "entity",
      "entity_id": "entity-1",
      "last_modified": "2025-09-01T14:30:00Z",
      "modified_by": "user-456"
    }
  ],
  "pagination": {
    "total": 25,
    "page": 1,
    "limit": 50,
    "total_pages": 1
  },
  "search_metadata": {
    "query_time_ms": 45,
    "total_scanned": 1250,
    "filters_applied": 6
  }
}
```

### Configuration Usage Analysis
Analyze configuration usage patterns across the tenant.

**Endpoint**: `GET /api/v1/settings/analysis/usage`

#### **Query Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `module` | string | No | Analyze specific module |
| `period` | string | No | Analysis period (1d, 7d, 30d) |

**Response** (200 OK):
```json
{
  "analysis_period": "30d",
  "module_summary": [
    {
      "module": "finance", 
      "total_configs": 45,
      "entity_overrides": 123,
      "most_customized": [
        {
          "config_key": "invoice_prefix",
          "override_percentage": 85,
          "entity_count": 34
        }
      ]
    }
  ],
  "configuration_patterns": {
    "most_overridden": [
      "finance.invoice_prefix",
      "inventory.reorder_threshold"
    ],
    "never_overridden": [
      "finance.default_currency",
      "hr.base_pay_frequency"
    ]
  },
  "template_adoption": {
    "total_applications": 15,
    "most_applied_template": "Manufacturing Standard",
    "recent_applications": 3
  }
}
```

---

## Data Models

### Configuration Model
```json
{
  "type": "object",
  "properties": {
    "module": {
      "type": "string",
      "enum": ["finance", "hr", "inventory", "sales", "purchasing"],
      "description": "Module that owns this configuration"
    },
    "config_key": {
      "type": "string",
      "pattern": "^[a-z0-9_]+$",
      "maxLength": 100,
      "description": "Configuration key identifier"
    },
    "value": {
      "description": "Configuration value (type varies by data_type)"
    },
    "data_type": {
      "type": "string",
      "enum": ["string", "integer", "boolean", "decimal", "json"],
      "description": "Data type of the configuration value"
    },
    "source": {
      "type": "string",
      "enum": ["system", "tenant", "entity", "template"],
      "description": "Source of the configuration value"
    },
    "is_inherited": {
      "type": "boolean",
      "description": "Whether value is inherited from parent level"
    },
    "can_override": {
      "type": "boolean", 
      "description": "Whether this configuration can be overridden"
    },
    "required_permission": {
      "type": "string",
      "description": "Permission required to modify this configuration"
    },
    "last_modified": {
      "type": "string",
      "format": "date-time",
      "description": "Last modification timestamp"
    },
    "version": {
      "type": "integer",
      "description": "Version for optimistic locking"
    }
  },
  "required": ["module", "config_key", "value", "data_type", "source"]
}
```

### Template Model
```json
{
  "type": "object",
  "properties": {
    "id": {
      "type": "string",
      "format": "uuid",
      "description": "Template identifier"
    },
    "name": {
      "type": "string",
      "maxLength": 255,
      "description": "Template display name"
    },
    "category": {
      "type": "string",
      "enum": ["industry", "functional", "regional"],
      "description": "Template category"
    },
    "description": {
      "type": "string",
      "maxLength": 1000,
      "description": "Template description"
    },
    "version": {
      "type": "string",
      "pattern": "^\\d+\\.\\d+(\\.\\d+)?$",
      "description": "Semantic version"
    },
    "configurations": {
      "type": "array",
      "items": {
        "$ref": "#/components/schemas/TemplateConfiguration"
      },
      "description": "Configuration definitions in template"
    },
    "applicable_tenant_types": {
      "type": "array",
      "items": {"type": "string"},
      "description": "Tenant types this template applies to"
    },
    "required_feature_flags": {
      "type": "array", 
      "items": {"type": "string"},
      "description": "Feature flags required for this template"
    },
    "conflict_resolution": {
      "type": "string",
      "enum": ["merge", "replace", "preserve"],
      "description": "How to handle configuration conflicts"
    },
    "is_active": {
      "type": "boolean",
      "description": "Whether template is active and available"
    }
  },
  "required": ["name", "category", "version", "configurations"]
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
    "correlation_id": {
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
| 400 | `invalid_configuration_value` | Configuration value invalid for data type |
| 400 | `invalid_module` | Module name not recognized |
| 401 | `unauthorized` | Authentication required |
| 403 | `forbidden` | Insufficient permissions |
| 403 | `configuration_not_overridable` | Configuration cannot be overridden |
| 403 | `feature_flag_disabled` | Required feature flag not enabled |
| 404 | `configuration_not_found` | Configuration not found |
| 404 | `template_not_found` | Template not found |
| 409 | `template_conflict` | Template application conflicts |
| 409 | `version_conflict` | Optimistic locking conflict |
| 422 | `template_dependency_missing` | Template dependencies not satisfied |
| 422 | `configuration_constraint_violation` | Configuration constraint violated |
| 429 | `rate_limit_exceeded` | API rate limit exceeded |
| 500 | `internal_server_error` | Unexpected server error |

---

## Code Examples

### JavaScript/Node.js Examples

#### Get Configuration with Inheritance
```javascript
const axios = require('axios');

const client = axios.create({
  baseURL: 'https://api.awo-erp.com/api/v1/settings',
  headers: {
    'Authorization': `Bearer ${process.env.JWT_TOKEN}`,
    'X-Tenant-ID': process.env.TENANT_ID,
    'Content-Type': 'application/json'
  }
});

// Get configuration with inheritance metadata
async function getConfiguration(module, configKey, entityId = null) {
  try {
    const params = {
      include_metadata: true
    };
    
    if (entityId) {
      params.entity_id = entityId;
    }
    
    const response = await client.get(`/config/${module}/${configKey}`, { params });
    
    console.log(`Configuration: ${response.data.value}`);
    console.log(`Source: ${response.data.source}`);
    console.log(`Is Inherited: ${response.data.is_inherited}`);
    
    return response.data;
  } catch (error) {
    console.error('Error getting configuration:', error.response?.data);
    throw error;
  }
}

// Apply configuration template
async function applyTemplate(templateId, targetId, targetType = 'tenant') {
  try {
    const payload = {
      target_type: targetType,
      target_id: targetId,
      options: {
        preserve_existing: true,
        conflict_resolution: 'merge',
        dry_run: false
      }
    };
    
    const response = await client.post(`/templates/${templateId}/apply`, payload);
    
    console.log(`Applied ${response.data.summary.applied} configurations`);
    console.log(`Conflicts: ${response.data.conflicts.length}`);
    
    return response.data;
  } catch (error) {
    console.error('Error applying template:', error.response?.data);
    throw error;
  }
}
```

#### Bulk Configuration Updates
```javascript
// Perform bulk configuration update
async function bulkUpdateConfigurations(targets, configurations) {
  try {
    const payload = {
      targets: targets,
      configurations: configurations,
      options: {
        preserve_explicit_overrides: true,
        batch_size: 50,
        dry_run: false
      }
    };
    
    const response = await client.post('/bulk-update', payload);
    const operationId = response.data.operation_id;
    
    console.log(`Bulk update started: ${operationId}`);
    
    // Poll for completion
    return await pollBulkOperation(operationId);
  } catch (error) {
    console.error('Error starting bulk update:', error.response?.data);
    throw error;
  }
}

async function pollBulkOperation(operationId) {
  while (true) {
    const response = await client.get(`/operations/${operationId}`);
    const status = response.data.status;
    
    console.log(`Operation ${operationId}: ${status} (${response.data.progress.percentage}%)`);
    
    if (status === 'completed' || status === 'failed') {
      return response.data;
    }
    
    await new Promise(resolve => setTimeout(resolve, 2000));
  }
}
```

### Go Examples

#### Configuration Resolution
```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
)

type ConfigurationResponse struct {
    Module      string      `json:"module"`
    ConfigKey   string      `json:"config_key"`
    Value       interface{} `json:"value"`
    DataType    string      `json:"data_type"`
    Source      string      `json:"source"`
    IsInherited bool        `json:"is_inherited"`
}

type ResolutionRequest struct {
    Context struct {
        TenantID string `json:"tenant_id"`
        EntityID string `json:"entity_id"`
    } `json:"context"`
    Configurations []struct {
        Module    string `json:"module"`
        ConfigKey string `json:"config_key"`
    } `json:"configurations"`
    Options struct {
        IncludeInheritanceChain bool `json:"include_inheritance_chain"`
        IncludeMetadata         bool `json:"include_metadata"`
    } `json:"options"`
}

func resolveConfigurations(tenantID, entityID string, configs []struct{ Module, ConfigKey string }) ([]ConfigurationResponse, error) {
    request := ResolutionRequest{}
    request.Context.TenantID = tenantID
    request.Context.EntityID = entityID
    request.Configurations = configs
    request.Options.IncludeInheritanceChain = true

    jsonData, err := json.Marshal(request)
    if err != nil {
        return nil, err
    }

    req, err := http.NewRequest("POST", 
        "https://api.awo-erp.com/api/v1/settings/resolve", 
        bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }

    req.Header.Set("Authorization", "Bearer "+os.Getenv("JWT_TOKEN"))
    req.Header.Set("X-Tenant-ID", tenantID)
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result struct {
        ResolvedConfigurations []ConfigurationResponse `json:"resolved_configurations"`
    }
    
    err = json.NewDecoder(resp.Body).Decode(&result)
    return result.ResolvedConfigurations, err
}
```

---

## Testing

### Postman Collection
A  Postman collection is available with:
- Pre-configured environments (dev, staging, production)
- Authentication setup scripts  
- Complete endpoint coverage
- Configuration templates and examples
- Bulk operation workflows

**Download**: Postman Collection

### API Testing Checklist
- [ ] Authentication works correctly
- [ ] Configuration resolution through inheritance hierarchy
- [ ] Template application with conflict resolution
- [ ] Bulk operations handle large datasets
- [ ] Permission enforcement at all levels
- [ ] Cache invalidation works properly
- [ ] Error responses are properly formatted
- [ ] Rate limiting is enforced
- [ ] Multi-tenancy isolation is verified
- [ ] Performance requirements are met

---

## Changelog

### Version 1.0.0 (2025-09-02)
- Initial API design based on Settings PRD
- Complete configuration management endpoints
- Template-based configuration system
- Bulk operation support
- Three-level inheritance resolution
-  error handling
- Multi-tenant security model

---

**Document Control**  
- **Version**: 1.0
- **Last Updated**: September 2, 2025
- **API Status**: Development
- **OpenAPI Spec**: swagger.yaml

**Related Documents**
- [Architecture Guide](architecture-guide.md)
- [Integration Guide](integration-guide.md)  
- [Settings PRD](index.md)
