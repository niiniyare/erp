# Access Request Workflow API Testing Guide

This guide provides  curl commands to test the Access Request Workflow API including conditional access controls and user analytics.

##  Prerequisites

1. **Start the Server**
```bash
# Set environment variables
export DB_USER=admin
export DB_PASSWORD=admin
export DB_NAME=ledger
export DB_HOST=localhost
export DB_PORT=5432
export REDIS_HOST=localhost
export REDIS_PORT=6379

# Build and run the server
go run cmd/server/main.go
```

2. **Install jq for JSON processing**
```bash
# Ubuntu/Debian
sudo apt-get install jq

# macOS
brew install jq

# Windows (with Chocolatey)
choco install jq
```

##  API Endpoints Overview

### Access Request Workflow
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/access-requests` | Create access request |
| GET | `/api/v1/access-requests/{id}` | Get access request by ID |
| POST | `/api/v1/access-requests/{id}/process` | Process access request |
| DELETE | `/api/v1/access-requests/{id}` | Revoke access request |
| GET | `/api/v1/access-requests` | List access requests |
| GET | `/api/v1/access-requests/stats` | Get access request statistics |

### Conditional Access
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/conditional-access/evaluate` | Evaluate conditional access |
| POST | `/api/v1/conditional-access/rules` | Create conditional access rule |

### User Analytics
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/analytics/users/{id}/behavior` | Get user behavior analytics |
| GET | `/api/v1/analytics/users/{id}/risk` | Get user risk assessment |
| GET | `/api/v1/analytics/users/{id}/insights` | Get personalized insights |
| POST | `/api/v1/analytics/users/{id}/detect-anomalies` | Detect user anomalies |

##  Health Check

First, verify the server is running:

```bash
# Health check
curl -X GET http://localhost:8080/health | jq .

# Expected response:
# {
#   "service": "awo",
#   "status": "ok",
#   "version": "1.0.0"
# }
```

##  Access Request Workflow Tests

### 1. Create Access Request

```bash
# Create a role assignment request
curl -X POST http://localhost:8080/api/v1/access-requests \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 00000000-0000-0000-0000-000000000001" \
  -d '{
    "entity_id": "00000000-0000-0000-0000-000000000001",
    "request_type": "ROLE_ASSIGNMENT",
    "role_id": "00000000-0000-0000-0000-000000000002",
    "target_user_id": "00000000-0000-0000-0000-000000000003",
    "justification": "User needs admin access for project management",
    "business_reason": "Required for Q1 project delivery",
    "duration_hours": 168,
    "auto_revoke": true
  }' | jq .

# Expected response (201 Created):
# {
#   "message": "Access request created successfully",
#   "access_request": {
#     "id": "uuid-here",
#     "entity_id": "00000000-0000-0000-0000-000000000001",
#     "request_type": "ROLE_ASSIGNMENT",
#     "approval_status": "PENDING",
#     "created_at": "2024-01-01T10:00:00Z"
#   }
# }
```

### 2. Create Permission Grant Request

```bash
# Create a permission grant request
curl -X POST http://localhost:8080/api/v1/access-requests \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 00000000-0000-0000-0000-000000000001" \
  -d '{
    "entity_id": "00000000-0000-0000-0000-000000000001",
    "request_type": "PERMISSION_GRANT",
    "permission_id": "00000000-0000-0000-0000-000000000004",
    "target_user_id": "00000000-0000-0000-0000-000000000003",
    "justification": "User needs read access to financial reports",
    "business_reason": "Monthly reporting requirements",
    "duration_hours": 24,
    "auto_revoke": true
  }' | jq .
```

### 3. Create Resource Access Request

```bash
# Create a resource access request
curl -X POST http://localhost:8080/api/v1/access-requests \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 00000000-0000-0000-0000-000000000001" \
  -d '{
    "entity_id": "00000000-0000-0000-0000-000000000001",
    "request_type": "RESOURCE_ACCESS",
    "resource_id": "00000000-0000-0000-0000-000000000005",
    "target_user_id": "00000000-0000-0000-0000-000000000003",
    "justification": "Need access to customer database for support",
    "business_reason": "Customer escalation case #12345",
    "duration_hours": 8,
    "auto_revoke": true
  }' | jq .
```

### 4. Create Privilege Elevation Request

```bash
# Create a privilege elevation request
curl -X POST http://localhost:8080/api/v1/access-requests \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 00000000-0000-0000-0000-000000000001" \
  -d '{
    "entity_id": "00000000-0000-0000-0000-000000000001",
    "request_type": "ELEVATION",
    "justification": "Emergency system maintenance required",
    "business_reason": "Critical security patch deployment",
    "duration_hours": 4,
    "auto_revoke": true
  }' | jq .
```

### 5. Get Access Request by ID

```bash
# Replace REQUEST_ID with actual ID from create response
REQUEST_ID="your-request-id-here"

curl -X GET http://localhost:8080/api/v1/access-requests/$REQUEST_ID \
  -H "Content-Type: application/json" | jq .

# Expected response (200 OK):
# {
#   "access_request": {
#     "id": "request-id",
#     "entity_id": "entity-id",
#     "request_type": "ROLE_ASSIGNMENT",
#     "approval_status": "PENDING",
#     "created_at": "2024-01-01T10:00:00Z",
#     "updated_at": "2024-01-01T10:00:00Z"
#   }
# }
```

### 6. Process Access Request (Approve)

```bash
# Process (approve) an access request
curl -X POST http://localhost:8080/api/v1/access-requests/$REQUEST_ID/process \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 00000000-0000-0000-0000-000000000002" \
  -d '{
    "action": "approve",
    "comments": "Approved for business justification",
    "duration_hours": 24
  }' | jq .

# Expected response (200 OK):
# {
#   "message": "Access request processed successfully",
#   "access_request": {
#     "id": "request-id",
#     "approval_status": "APPROVED",
#     "approved_by": "approver-id",
#     "approved_at": "2024-01-01T10:30:00Z"
#   }
# }
```

### 7. Process Access Request (Reject)

```bash
# Process (reject) an access request
curl -X POST http://localhost:8080/api/v1/access-requests/$REQUEST_ID/process \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 00000000-0000-0000-0000-000000000002" \
  -d '{
    "action": "reject",
    "comments": "Insufficient business justification provided"
  }' | jq .

# Expected response (200 OK):
# {
#   "message": "Access request processed successfully",
#   "access_request": {
#     "id": "request-id",
#     "approval_status": "REJECTED",
#     "approved_by": "approver-id",
#     "approved_at": "2024-01-01T10:30:00Z"
#   }
# }
```

### 8. List Access Requests

```bash
# List all access requests
curl -X GET http://localhost:8080/api/v1/access-requests \
  -H "Content-Type: application/json" | jq .

# List with pagination
curl -X GET "http://localhost:8080/api/v1/access-requests?limit=10&offset=0" \
  -H "Content-Type: application/json" | jq .

# Filter by status
curl -X GET "http://localhost:8080/api/v1/access-requests?status=PENDING&limit=5" \
  -H "Content-Type: application/json" | jq .

# Filter by request type
curl -X GET "http://localhost:8080/api/v1/access-requests?request_type=ROLE_ASSIGNMENT" \
  -H "Content-Type: application/json" | jq .

# Expected response:
# {
#   "access_requests": [
#     {
#       "id": "request-id",
#       "request_type": "ROLE_ASSIGNMENT",
#       "approval_status": "PENDING",
#       "created_at": "2024-01-01T10:00:00Z"
#     }
#   ],
#   "total": 1,
#   "limit": 20,
#   "offset": 0
# }
```

### 9. Get Access Request Statistics

```bash
# Get access request statistics
curl -X GET "http://localhost:8080/api/v1/access-requests/stats?from_date=2024-01-01&to_date=2024-12-31" \
  -H "Content-Type: application/json" | jq .

# Expected response:
# {
#   "statistics": {
#     "total_requests": 150,
#     "pending_requests": 12,
#     "approved_requests": 120,
#     "rejected_requests": 18,
#     "expired_requests": 0,
#     "avg_approval_time_hours": 6,
#     "requests_by_type": {
#       "ROLE_ASSIGNMENT": 45,
#       "PERMISSION_GRANT": 35,
#       "RESOURCE_ACCESS": 50,
#       "ELEVATION": 20
#     }
#   }
# }
```

### 10. Revoke Access Request

```bash
# Revoke an approved access request
curl -X DELETE http://localhost:8080/api/v1/access-requests/$REQUEST_ID \
  -H "Content-Type: application/json" | jq .

# Expected response (200 OK):
# {
#   "message": "Access request revoked successfully",
#   "access_request": {
#     "id": "request-id",
#     "approval_status": "REVOKED",
#     "updated_at": "2024-01-01T11:00:00Z"
#   }
# }
```

##  Conditional Access Tests

### 11. Evaluate Conditional Access

```bash
# Evaluate conditional access for a user action
curl -X POST http://localhost:8080/api/v1/conditional-access/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "00000000-0000-0000-0000-000000000001",
    "resource_name": "user-management",
    "action": "read",
    "context": {
      "ip_address": "192.168.1.100",
      "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
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

# Expected response (200 OK):
# {
#   "evaluation_result": {
#     "access_granted": true,
#     "effect": "ALLOW",
#     "matched_rules": [
#       {
#         "rule_id": "rule-id",
#         "rule_name": "Business Hours Access",
#         "rule_type": "TIME_RESTRICTION"
#       }
#     ],
#     "required_actions": [],
#     "risk_score": 15
#   }
# }
```

### 12. Create Conditional Access Rule

```bash
# Create a time-based conditional access rule
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

# Expected response (201 Created):
# {
#   "message": "Conditional access rule created successfully",
#   "rule": {
#     "id": "rule-id",
#     "name": "Business Hours Access Only",
#     "rule_type": "TIME_RESTRICTION",
#     "priority": 100,
#     "is_active": true,
#     "created_at": "2024-01-01T10:00:00Z"
#   }
# }
```

### 13. Create Location-Based Conditional Access Rule

```bash
# Create a location-based conditional access rule
curl -X POST http://localhost:8080/api/v1/conditional-access/rules \
  -H "Content-Type: application/json" \
  -d '{
    "entity_id": "00000000-0000-0000-0000-000000000001",
    "name": "Geo-Location Restriction",
    "description": "Block access from unauthorized countries",
    "rule_type": "LOCATION_RESTRICTION",
    "priority": 200,
    "is_active": true,
    "target_actions": ["admin", "sensitive"],
    "location_rules": {
      "allowed_countries": ["US", "CA", "GB"],
      "blocked_countries": ["CN", "RU"],
      "allowed_regions": ["CA", "NY", "TX"],
      "blocked_regions": []
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
          "severity": "HIGH",
          "category": "SECURITY_VIOLATION"
        }
      }
    ]
  }' | jq .
```

### 14. Create Device-Based Conditional Access Rule

```bash
# Create a device-based conditional access rule
curl -X POST http://localhost:8080/api/v1/conditional-access/rules \
  -H "Content-Type: application/json" \
  -d '{
    "entity_id": "00000000-0000-0000-0000-000000000001",
    "name": "Trusted Device Only",
    "description": "Allow access only from trusted devices",
    "rule_type": "DEVICE_RESTRICTION",
    "priority": 150,
    "is_active": true,
    "target_actions": ["admin", "financial"],
    "device_rules": {
      "allowed_device_types": ["laptop", "desktop"],
      "blocked_device_types": ["mobile", "tablet"],
      "allowed_os": ["Windows 10", "macOS", "Ubuntu"],
      "blocked_os": ["Android", "iOS"],
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

##  User Analytics Tests

### 15. Get User Behavior Analytics

```bash
# Get  user behavior analytics
curl -X GET http://localhost:8080/api/v1/analytics/users/00000000-0000-0000-0000-000000000001/behavior \
  -H "Content-Type: application/json" | jq .

# Expected response (200 OK):
# {
#   "behavior_pattern": {
#     "user_id": "00000000-0000-0000-0000-000000000001",
#     "analysis_period": {
#       "start": "2024-06-01T00:00:00Z",
#       "end": "2024-07-01T00:00:00Z"
#     },
#     "login_patterns": {
#       "typical_login_hours": [9, 10, 13, 14],
#       "average_session_duration": 14400000000000,
#       "login_time_variance": 2.5,
#       "weekend_login_frequency": 0.15
#     },
#     "activity_patterns": {
#       "peak_activity_hours": [10, 11, 14, 15],
#       "activity_distribution": {
#         "administration": 0.2,
#         "data_entry": 0.3,
#         "document_access": 0.4,
#         "reporting": 0.1
#       }
#     },
#     "security_profile": {
#       "password_change_frequency": 0.25,
#       "mfa_usage_consistency": 0.95,
#       "compliance_score": 0.88
#     }
#   }
# }
```

### 16. Get User Risk Assessment

```bash
# Get user risk assessment
curl -X GET http://localhost:8080/api/v1/analytics/users/00000000-0000-0000-0000-000000000001/risk \
  -H "Content-Type: application/json" | jq .

# Expected response (200 OK):
# {
#   "risk_assessment": {
#     "user_id": "00000000-0000-0000-0000-000000000001",
#     "overall_risk_score": 25,
#     "risk_level": "LOW",
#     "risk_categories": {
#       "access_patterns": 15,
#       "login_behavior": 20,
#       "security_behavior": 30
#     },
#     "risk_factors": [
#       {
#         "factor": "unusual_login_time",
#         "severity": "medium",
#         "description": "Login detected outside normal hours"
#       }
#     ],
#     "recommendations": [
#       {
#         "priority": "medium",
#         "action": "Enable MFA for all accounts",
#         "reason": "Additional security layer recommended"
#       }
#     ],
#     "last_assessment": "2024-01-01T10:00:00Z",
#     "next_assessment": "2024-01-02T10:00:00Z"
#   }
# }
```

### 17. Get User Personalized Insights

```bash
# Get personalized insights for user
curl -X GET http://localhost:8080/api/v1/analytics/users/00000000-0000-0000-0000-000000000001/insights \
  -H "Content-Type: application/json" | jq .

# Expected response (200 OK):
# {
#   "insights": {
#     "user_id": "00000000-0000-0000-0000-000000000001",
#     "generated_at": "2024-01-01T10:00:00Z",
#     "productivity_insights": [
#       {
#         "category": "time_management",
#         "insight": "You're most productive between 10 AM and 2 PM",
#         "recommendation": "Schedule important tasks during peak hours",
#         "impact": "medium"
#       }
#     ],
#     "security_insights": [
#       {
#         "category": "authentication",
#         "insight": "MFA usage is inconsistent",
#         "recommendation": "Enable MFA for all critical applications",
#         "impact": "high"
#       }
#     ],
#     "efficiency_insights": [
#       {
#         "category": "workflow",
#         "insight": "Document access patterns suggest workflow optimization opportunities",
#         "recommendation": "Consider organizing frequently accessed documents",
#         "impact": "medium"
#       }
#     ],
#     "collaboration_insights": [
#       {
#         "category": "team_interaction",
#         "insight": "Higher collaboration during morning hours",
#         "recommendation": "Schedule team meetings in the morning",
#         "impact": "low"
#       }
#     ]
#   }
# }
```

### 18. Detect User Anomalies

```bash
# Detect anomalies in user behavior
curl -X POST http://localhost:8080/api/v1/analytics/users/00000000-0000-0000-0000-000000000001/detect-anomalies \
  -H "Content-Type: application/json" \
  -d '{
    "analysis_period": {
      "start": "2024-01-01T00:00:00Z",
      "end": "2024-01-31T23:59:59Z"
    },
    "anomaly_types": ["login_time", "location", "device", "access_pattern"],
    "sensitivity": "medium"
  }' | jq .

# Expected response (200 OK):
# {
#   "anomaly_detection": {
#     "user_id": "00000000-0000-0000-0000-000000000001",
#     "analysis_period": {
#       "start": "2024-01-01T00:00:00Z",
#       "end": "2024-01-31T23:59:59Z"
#     },
#     "anomalies_detected": [
#       {
#         "type": "unusual_login_time",
#         "severity": "medium",
#         "description": "Login detected at 2:30 AM, outside normal hours (9 AM - 6 PM)",
#         "detected_at": "2024-01-15T02:30:00Z",
#         "deviation_score": 0.85,
#         "baseline_pattern": "Normal login hours: 9 AM - 6 PM",
#         "recommendations": [
#           "Verify if this was authorized access",
#           "Consider implementing time-based access controls"
#         ]
#       }
#     ],
#     "anomaly_summary": {
#       "total_anomalies": 1,
#       "high_severity": 0,
#       "medium_severity": 1,
#       "low_severity": 0
#     },
#     "baseline_confidence": 0.82,
#     "next_analysis": "2024-02-01T00:00:00Z"
#   }
# }
```

## ❌ Error Scenarios Testing

### 19. Invalid Requests

```bash
# Create access request with missing required fields
curl -X POST http://localhost:8080/api/v1/access-requests \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 00000000-0000-0000-0000-000000000001" \
  -d '{
    "request_type": "ROLE_ASSIGNMENT"
  }' | jq .

# Expected response (400 Bad Request):
# {
#   "error": "Invalid request payload",
#   "details": "entity_id is required"
# }

# Get non-existent access request
curl -X GET http://localhost:8080/api/v1/access-requests/invalid-uuid \
  -H "Content-Type: application/json" | jq .

# Expected response (400 Bad Request):
# {
#   "error": "Invalid request ID format"
# }

# Evaluate conditional access with missing context
curl -X POST http://localhost:8080/api/v1/conditional-access/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "00000000-0000-0000-0000-000000000001"
  }' | jq .

# Expected response (400 Bad Request):
# {
#   "error": "Invalid evaluation request",
#   "details": "resource_name and action are required"
# }
```

### 20. Unauthorized Access

```bash
# Try to process access request without approver header
curl -X POST http://localhost:8080/api/v1/access-requests/00000000-0000-0000-0000-000000000001/process \
  -H "Content-Type: application/json" \
  -d '{
    "action": "approve"
  }' | jq .

# Expected response (401 Unauthorized):
# {
#   "error": "User ID required in X-User-ID header"
# }

# Try to create access request without user header
curl -X POST http://localhost:8080/api/v1/access-requests \
  -H "Content-Type: application/json" \
  -d '{
    "entity_id": "00000000-0000-0000-0000-000000000001",
    "request_type": "ROLE_ASSIGNMENT"
  }' | jq .

# Expected response (401 Unauthorized):
# {
#   "error": "User ID required in X-User-ID header"
# }
```

##  Complete Test Sequence

Here's a complete test sequence that exercises all major functionality:

```bash
#!/bin/bash

# Set base URL
BASE_URL="http://localhost:8080"

echo " Starting Access Request Workflow API Test Suite..."

# 1. Health check
echo "1. Health check..."
curl -s $BASE_URL/health | jq .

# 2. Create access request
echo "2. Creating access request..."
ACCESS_REQUEST_ID=$(curl -s -X POST $BASE_URL/api/v1/access-requests \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 00000000-0000-0000-0000-000000000001" \
  -d '{
    "entity_id": "00000000-0000-0000-0000-000000000001",
    "request_type": "ROLE_ASSIGNMENT",
    "role_id": "00000000-0000-0000-0000-000000000002",
    "target_user_id": "00000000-0000-0000-0000-000000000003",
    "justification": "Testing access request workflow",
    "business_reason": "API Testing",
    "duration_hours": 24
  }' | jq -r '.access_request.id // empty')

echo "Created access request: $ACCESS_REQUEST_ID"

# 3. Get access request
echo "3. Getting access request..."
curl -s -X GET $BASE_URL/api/v1/access-requests/$ACCESS_REQUEST_ID | jq .

# 4. Create conditional access rule
echo "4. Creating conditional access rule..."
RULE_ID=$(curl -s -X POST $BASE_URL/api/v1/conditional-access/rules \
  -H "Content-Type: application/json" \
  -d '{
    "entity_id": "00000000-0000-0000-0000-000000000001",
    "name": "Test Business Hours Rule",
    "description": "Test rule for business hours access",
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
  }' | jq -r '.rule.id // empty')

echo "Created conditional access rule: $RULE_ID"

# 5. Test user behavior analytics
echo "5. Testing user behavior analytics..."
curl -s -X GET $BASE_URL/api/v1/analytics/users/00000000-0000-0000-0000-000000000001/behavior | jq .

# 6. Test user risk assessment
echo "6. Testing user risk assessment..."
curl -s -X GET $BASE_URL/api/v1/analytics/users/00000000-0000-0000-0000-000000000001/risk | jq .

# 7. Test personalized insights
echo "7. Testing personalized insights..."
curl -s -X GET $BASE_URL/api/v1/analytics/users/00000000-0000-0000-0000-000000000001/insights | jq .

# 8. Test anomaly detection
echo "8. Testing anomaly detection..."
curl -s -X POST $BASE_URL/api/v1/analytics/users/00000000-0000-0000-0000-000000000001/detect-anomalies \
  -H "Content-Type: application/json" \
  -d '{
    "analysis_period": {
      "start": "2024-01-01T00:00:00Z",
      "end": "2024-01-31T23:59:59Z"
    },
    "anomaly_types": ["login_time", "location", "device"],
    "sensitivity": "medium"
  }' | jq .

echo "✅ Access Request Workflow API test suite completed!"
```

##  Notes

- Replace UUIDs with actual values from your system
- All endpoints require proper authentication headers
- The X-User-ID header is required for user context
- Duration is specified in hours
- All timestamps are in ISO 8601 format
- Analytics endpoints return  behavioral data
- Conditional access rules support multiple restriction types
- Risk assessments are calculated based on behavioral patterns

##  Troubleshooting

1. **Server not responding**: Check if server is running on port 8080
2. **Database errors**: Verify database connection and migrations
3. **Missing user header**: Ensure X-User-ID header is provided
4. **Invalid UUIDs**: Check UUID format and existence
5. **Analytics timeouts**: Large datasets may require longer timeouts
6. **Conditional access evaluation**: Verify rule configuration and context data

## ️ Environment Setup

Before running tests, ensure your environment is set up:

```bash
# Database environment variables
export DB_USER=admin
export DB_PASSWORD=admin
export DB_NAME=ledger
export DB_HOST=localhost
export DB_PORT=5432

# Redis environment variables (for caching)
export REDIS_HOST=localhost
export REDIS_PORT=6379

# Start the server
go run cmd/server/main.go
```

The server should be running on `http://localhost:8080` for all these curl commands to work.