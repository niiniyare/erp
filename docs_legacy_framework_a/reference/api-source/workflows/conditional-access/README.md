> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

# Conditional Access Controls API

The Conditional Access Controls API provides dynamic access control based on contextual factors like time, location, device, network, and risk assessment. It implements a powerful rule engine for real-time access decisions.

##  Overview

The Conditional Access Controls API enables:

- **Real-time Access Evaluation**: Dynamic access decisions based on context
- **Multi-Factor Rules**: Time, location, device, network, and risk-based controls
- **Rule Engine**: Powerful rule evaluation with priority handling
- **Policy Enforcement**: Granular access control policies
- **Effect Actions**: Allow, deny, challenge, audit-only responses
- **Integration Ready**: Seamless integration with existing access controls

##  Base URL

```
http://localhost:8080/api/v1/conditional-access
```

##  Rule Types

| Type | Description | Use Case |
|------|-------------|----------|
| `TIME_RESTRICTION` | Time-based access control | Business hours, maintenance windows |
| `LOCATION_RESTRICTION` | Geographic access control | Country/region restrictions |
| `DEVICE_RESTRICTION` | Device-based access control | Trusted devices, mobile restrictions |
| `NETWORK_RESTRICTION` | Network-based access control | VPN requirements, IP restrictions |
| `RISK_BASED` | Risk assessment based control | High-risk user blocking |
| `COMBINED` | Multiple rule types combined | Complex access scenarios |

##  Access Effects

| Effect | Description | Actions |
|--------|-------------|---------|
| `ALLOW` | Grant access | No additional actions |
| `DENY` | Block access | Block, audit, notify |
| `CHALLENGE` | Require additional verification | MFA, step-up auth |
| `AUDIT_ONLY` | Allow but audit | Log for compliance |

##  Quick Start

### 1. Health Check
```bash
curl -X GET http://localhost:8080/health | jq .
```

### 2. Create Time-Based Rule
```bash
curl -X POST http://localhost:8080/api/v1/conditional-access/rules \
  -H "Content-Type: application/json" \
  -d '{
    "entity_id": "00000000-0000-0000-0000-000000000001",
    "name": "Business Hours Access Only",
    "description": "Restrict access to business hours only",
    "rule_type": "TIME_RESTRICTION",
    "priority": 100,
    "is_active": true,
    "target_actions": ["read", "write"],
    "time_restrictions": {
      "allowed_days": [1, 2, 3, 4, 5],
      "allowed_time_ranges": [
        {
          "start_time": "09:00",
          "end_time": "17:00"
        }
      ],
      "timezone": "America/New_York"
    },
    "effect": "DENY",
    "actions": [
      {
        "type": "BLOCK_ACCESS",
        "parameters": {
          "message": "Access denied outside business hours"
        }
      }
    ]
  }' | jq .
```

### 3. Evaluate Access
```bash
curl -X POST http://localhost:8080/api/v1/conditional-access/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "00000000-0000-0000-0000-000000000001",
    "resource_name": "user-management",
    "action": "read",
    "context": {
      "ip_address": "192.168.1.100",
      "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
      "location": {
        "country": "US",
        "region": "CA",
        "city": "San Francisco"
      },
      "device": {
        "type": "laptop",
        "os": "Windows 10",
        "browser": "Chrome"
      }
    }
  }' | jq .
```

##  Documentation Files

- **curl-examples.md** -  curl command examples
- **API Reference** - Detailed API specification
- **Schema Reference** - Request/response schemas
- **Rule Configuration Guide** - Rule setup and management
- **Integration Guide** - Integration with access control systems

##  Testing

### Automated Testing
```bash
# Run conditional access API tests
./utilities/scripts/test-conditional-access.sh

# Test specific rule types
./utilities/scripts/test-conditional-access.sh --rule-type TIME_RESTRICTION
```

### Manual Testing
```bash
# Follow step-by-step examples
cat curl-examples.md
```

##  Common Use Cases

### Business Hours Restriction
```bash
# Create business hours rule
curl -X POST http://localhost:8080/api/v1/conditional-access/rules \
  -H "Content-Type: application/json" \
  -d '{
    "entity_id": "entity-id",
    "name": "Business Hours Only",
    "rule_type": "TIME_RESTRICTION",
    "priority": 100,
    "time_restrictions": {
      "allowed_days": [1, 2, 3, 4, 5],
      "allowed_time_ranges": [
        {"start_time": "09:00", "end_time": "17:00"}
      ],
      "timezone": "America/New_York"
    },
    "effect": "DENY"
  }' | jq .
```

### Geographic Restriction
```bash
# Create location-based rule
curl -X POST http://localhost:8080/api/v1/conditional-access/rules \
  -H "Content-Type: application/json" \
  -d '{
    "entity_id": "entity-id",
    "name": "Geographic Restriction",
    "rule_type": "LOCATION_RESTRICTION",
    "priority": 200,
    "location_rules": {
      "allowed_countries": ["US", "CA", "GB"],
      "blocked_countries": ["CN", "RU"]
    },
    "effect": "DENY",
    "actions": [
      {
        "type": "BLOCK_ACCESS",
        "parameters": {
          "message": "Access denied from unauthorized location"
        }
      },
      {
        "type": "AUDIT_LOG",
        "parameters": {
          "severity": "HIGH"
        }
      }
    ]
  }' | jq .
```

### Device Trust Requirements
```bash
# Create device-based rule
curl -X POST http://localhost:8080/api/v1/conditional-access/rules \
  -H "Content-Type: application/json" \
  -d '{
    "entity_id": "entity-id",
    "name": "Trusted Device Only",
    "rule_type": "DEVICE_RESTRICTION",
    "priority": 150,
    "device_rules": {
      "allowed_device_types": ["laptop", "desktop"],
      "blocked_device_types": ["mobile", "tablet"],
      "require_managed_device": true,
      "require_compliant_device": true
    },
    "effect": "CHALLENGE",
    "actions": [
      {
        "type": "REQUIRE_MFA",
        "parameters": {
          "methods": ["authenticator", "sms"]
        }
      }
    ]
  }' | jq .
```

### Risk-Based Access Control
```bash
# Create risk-based rule
curl -X POST http://localhost:8080/api/v1/conditional-access/rules \
  -H "Content-Type: application/json" \
  -d '{
    "entity_id": "entity-id",
    "name": "High Risk User Block",
    "rule_type": "RISK_BASED",
    "priority": 300,
    "risk_rules": {
      "max_risk_score": 70,
      "risk_factors": ["unusual_location", "new_device", "suspicious_activity"],
      "block_high_risk": true,
      "require_additional_verification": true
    },
    "effect": "CHALLENGE",
    "actions": [
      {
        "type": "STEP_UP_AUTH",
        "parameters": {
          "auth_methods": ["mfa", "manager_approval"]
        }
      }
    ]
  }' | jq .
```

##  Expected Responses

### Rule Creation Response (201 Created)
```json
{
  "message": "Conditional access rule created successfully",
  "rule": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "tenant_id": "550e8400-e29b-41d4-a716-446655440001",
    "entity_id": "550e8400-e29b-41d4-a716-446655440002",
    "name": "Business Hours Access Only",
    "description": "Restrict access to business hours only",
    "rule_type": "TIME_RESTRICTION",
    "priority": 100,
    "is_active": true,
    "target_actions": ["read", "write"],
    "time_restrictions": {
      "allowed_days": [1, 2, 3, 4, 5],
      "allowed_time_ranges": [
        {
          "start_time": "09:00",
          "end_time": "17:00"
        }
      ],
      "timezone": "America/New_York"
    },
    "effect": "DENY",
    "actions": [
      {
        "type": "BLOCK_ACCESS",
        "parameters": {
          "message": "Access denied outside business hours"
        }
      }
    ],
    "created_at": "2024-01-01T10:00:00Z",
    "updated_at": "2024-01-01T10:00:00Z",
    "created_by": "550e8400-e29b-41d4-a716-446655440003"
  }
}
```

### Access Evaluation Response (200 OK)
```json
{
  "evaluation_result": {
    "access_granted": true,
    "effect": "ALLOW",
    "matched_rules": [
      {
        "rule_id": "550e8400-e29b-41d4-a716-446655440000",
        "rule_name": "Business Hours Access",
        "rule_type": "TIME_RESTRICTION",
        "priority": 100,
        "effect": "ALLOW"
      }
    ],
    "required_actions": [],
    "risk_score": 15,
    "evaluation_context": {
      "timestamp": "2024-01-01T14:30:00Z",
      "user_id": "550e8400-e29b-41d4-a716-446655440001",
      "resource_name": "user-management",
      "action": "read"
    },
    "policy_decisions": [
      {
        "rule_id": "550e8400-e29b-41d4-a716-446655440000",
        "decision": "ALLOW",
        "reason": "Access within business hours"
      }
    ]
  }
}
```

### Access Denied Response (200 OK)
```json
{
  "evaluation_result": {
    "access_granted": false,
    "effect": "DENY",
    "matched_rules": [
      {
        "rule_id": "550e8400-e29b-41d4-a716-446655440000",
        "rule_name": "Business Hours Access",
        "rule_type": "TIME_RESTRICTION",
        "priority": 100,
        "effect": "DENY"
      }
    ],
    "required_actions": [
      {
        "type": "BLOCK_ACCESS",
        "parameters": {
          "message": "Access denied outside business hours"
        }
      }
    ],
    "risk_score": 45,
    "denial_reason": "Access attempted outside allowed business hours"
  }
}
```

##  Common Issues

1. **Rule conflicts** - Multiple rules with same priority
2. **Invalid time zones** - Use IANA timezone names
3. **Missing context** - Required context fields not provided
4. **Rule evaluation errors** - Invalid rule configuration
5. **Performance issues** - Too many complex rules

##  Performance Notes

- Rule evaluation is optimized for millisecond response times
- Rules are cached for faster evaluation
- Rule priority determines evaluation order
- Complex rules may impact performance

##  Security Features

- **Context Validation**: All context data is validated
- **Rule Integrity**: Rules are digitally signed
- **Audit Logging**: All evaluations are logged
- **Tamper Detection**: Rule modification detection
- **Secure Defaults**: Secure-by-default rule behavior

##  Related APIs

- **Access Request Workflow**: Rules influence approval decisions
- **User Analytics**: Risk assessment integration
- **User Management**: User context for evaluation
- **Entity Management**: Rules are scoped to entities

##  Rule Configuration Best Practices

### Rule Priority
- **High Priority (1-100)**: Critical security rules
- **Medium Priority (101-500)**: Business logic rules
- **Low Priority (501-1000)**: Convenience rules

### Rule Naming
- Use descriptive names
- Include rule type in name
- Version rules for tracking

### Performance Optimization
- Use specific target actions
- Avoid overly complex rules
- Test rule performance impact

##  Monitoring and Metrics

### Key Metrics
- Rule evaluation time
- Access grant/deny rates
- Rule match frequencies
- Policy violation counts

### Alerts
- High evaluation times
- Frequent rule denials
- Rule configuration errors
- Unusual access patterns

##  Rule Management

### Rule Lifecycle
1. **Draft**: Rule created but not active
2. **Active**: Rule is enforcing access decisions
3. **Deprecated**: Rule scheduled for removal
4. **Archived**: Rule no longer active

### Rule Testing
- Test rules in audit-only mode
- Validate rule logic before activation
- Monitor rule impact after deployment
- Regular rule review and updates