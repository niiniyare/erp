# ABAC (Attribute-Based Access Control) API

## Overview

The ABAC API provides advanced authorization capabilities using attribute-based policies. This system evaluates permissions based on user attributes, resource attributes, environmental context, and the specific action being performed.

## Key Features

- **Real-time Policy Evaluation**: Sub-millisecond authorization decisions
- **Attribute Collection**: Dynamic attribute gathering from multiple sources
- **Policy Discovery**: Find applicable policies for any context
- **Audit Trail**: Complete decision history and reasoning
- **Bulk Operations**: Evaluate multiple authorization requests efficiently
- **Caching**: Intelligent caching for performance optimization

## Endpoints Overview

| Method | Endpoint | Purpose |
|--------|----------|---------|
| POST | `/abac/authorize` | Simple authorization check |
| POST | `/abac/evaluate` |  policy evaluation with explanation |
| POST | `/abac/evaluate-bulk` | Bulk authorization requests |
| POST | `/abac/collect-attributes` | Gather attributes for context |
| POST | `/abac/discover-policies` | Find applicable policies |
| POST | `/abac/explain` | Get detailed policy explanation |
| GET | `/abac/audit` | Retrieve decision history |
| GET | `/abac/metrics` | Performance and usage metrics |
| GET | `/abac/health` | System health status |
| POST | `/abac/invalidate-cache` | Clear authorization cache |

## Quick Start Example

### Simple Authorization Check
```bash
curl -X POST "http://localhost:8080/abac/authorize" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-123",
    "resource_id": "document-456", 
    "resource_type": "document",
    "action": "read"
  }'

# Response
{
  "allowed": true,
  "decision": "ALLOW",
  "evaluation_time_ms": 15,
  "request_id": "req-789"
}
```

###  Evaluation with Explanation
```bash
curl -X POST "http://localhost:8080/abac/evaluate" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-123",
    "resource_id": "document-456",
    "resource_type": "document", 
    "action": "write",
    "explain_decision": true,
    "include_advice": true,
    "context": {
      "ip_address": "192.168.1.100",
      "user_agent": "API Client 1.0",
      "time_of_day": "business_hours"
    }
  }'
```

## Advanced Features

### Attribute Collection
Dynamically collect attributes from various sources:

```bash
curl -X POST "http://localhost:8080/abac/collect-attributes" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-123",
    "resource_id": "project-789",
    "resource_type": "project",
    "action": "manage",
    "include_expired": false
  }'

# Response includes all relevant attributes
{
  "user_attributes": {
    "department": {"value": "engineering", "source": "ldap"},
    "clearance_level": {"value": "confidential", "source": "hr_system"}
  },
  "resource_attributes": {
    "classification": {"value": "internal", "source": "metadata"},
    "owner": {"value": "user-456", "source": "database"}
  },
  "environment_attributes": {
    "network_zone": {"value": "corporate", "source": "network_detection"},
    "device_trust": {"value": "managed", "source": "device_registry"}
  },
  "collection_time_ms": 45
}
```

### Policy Discovery
Find policies that might apply to a given context:

```bash
curl -X POST "http://localhost:8080/abac/discover-policies" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-123",
    "resource_type": "financial_report",
    "action": "export",
    "context": {
      "department": "finance"
    }
  }'

# Response shows applicable policies
{
  "applicable_policies": 3,
  "policies": [
    {
      "id": "policy-financial-export",
      "name": "Financial Report Export Policy",
      "description": "Controls access to financial report exports",
      "priority": 100,
      "applicable": true,
      "effect": "PERMIT"
    }
  ]
}
```

### Bulk Evaluation
Process multiple authorization requests in a single call:

```bash
curl -X POST "http://localhost:8080/abac/evaluate-bulk" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-123",
    "requests": [
      {
        "resource_id": "doc-1", 
        "resource_type": "document",
        "action": "read"
      },
      {
        "resource_id": "doc-2",
        "resource_type": "document", 
        "action": "write"
      }
    ],
    "use_cache": true,
    "fail_fast": false
  }'
```

## Response Formats

### Authorization Response
```json
{
  "allowed": boolean,
  "decision": "ALLOW|DENY|NOT_APPLICABLE",
  "evaluation_time_ms": number,
  "request_id": "string",
  "cache_hit": boolean,
  "policy_count": number,
  "evaluated_at": "2025-01-01T12:00:00Z"
}
```

### Detailed Evaluation Response
```json
{
  "allowed": boolean,
  "decision": "string",
  "explanation": {
    "reasoning_summary": "string",
    "policy_evaluations": [...],
    "attributes_used": {...},
    "combining_algorithm": "string"
  },
  "advice": [
    {
      "type": "WARNING",
      "description": "Consider additional verification",
      "parameters": {...}
    }
  ],
  "obligations": [
    {
      "type": "LOG_ACCESS",
      "description": "Log this access for audit",
      "parameters": {"audit_level": "high"}
    }
  ],
  "audit_trail": {
    "evaluation_steps": [...],
    "policies_applied": [...],
    "timing": {...}
  }
}
```

## Best Practices

### Performance Optimization
1. **Enable Caching**: Use `"use_cache": true` for repeated requests
2. **Batch Operations**: Use bulk evaluation for multiple checks
3. **Context Reuse**: Reuse attribute collections when possible
4. **Minimal Explanations**: Only request explanations when needed for debugging

### Security Considerations
1. **Request IDs**: Always include unique request IDs for audit tracking
2. **Context Validation**: Ensure context attributes are validated and sanitized
3. **Rate Limiting**: Respect API rate limits for bulk operations
4. **Error Handling**: Implement proper error handling for authorization failures

### Integration Patterns
1. **Middleware Integration**: Embed authorization checks in application middleware
2. **Policy-as-Code**: Define policies in version control alongside application code
3. **Attribute Sources**: Configure multiple attribute sources for  context
4. **Audit Integration**: Stream audit events to security monitoring systems

## Error Handling

Common error scenarios and responses:

### Missing User ID
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "user_id is required",
    "details": {"field": "user_id"}
  }
}
```

### Policy Evaluation Timeout
```json
{
  "error": {
    "code": "EVALUATION_TIMEOUT", 
    "message": "Policy evaluation exceeded timeout",
    "details": {"timeout_ms": 5000}
  }
}
```

### Attribute Collection Failure
```json
{
  "error": {
    "code": "ATTRIBUTE_COLLECTION_FAILED",
    "message": "Failed to collect required attributes",
    "details": {
      "failed_sources": ["ldap", "database"],
      "partial_success": true
    }
  }
}
```

## Monitoring & Metrics

Access  metrics about ABAC system performance:

```bash
curl -X GET "http://localhost:8080/abac/metrics" \
  -H "Authorization: Bearer $JWT_TOKEN"

# Returns detailed performance metrics
{
  "total_evaluations": 1500000,
  "average_evaluation_time_ms": 12.5,
  "cache_hit_rate": 0.75,
  "policy_hit_distribution": {...},
  "error_rate": 0.001,
  "attribute_collection_stats": {...}
}
```

For detailed implementation guidance, see the [ABAC Integration Guide](../../../modules/user/abac/policy_evaluation.md).

---

**Next**: [Authentication API](../auth/) | **Up**: [API Reference](../index.md)