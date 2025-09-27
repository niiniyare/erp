# Access Request Workflow API

The Access Request Workflow API provides a  system for managing access requests including role assignments, permission grants, resource access, and privilege elevations with full approval workflows.

## 📋 Overview

The Access Request Workflow API implements a state machine pattern for managing access requests with:

- **Request Types**: Role assignment, permission grants, resource access, privilege elevation
- **Approval Workflows**: Multi-step approval processes with configurable rules
- **State Management**: Pending, approved, rejected, expired, revoked states
- **Audit Trail**: Complete audit logging of all workflow events
- **Notification System**: Automated notifications for approvers and requesters
- **Auto-Revocation**: Automatic access revocation after expiration
- **Risk Assessment**: Integration with risk assessment for approval decisions

## 🔧 Base URL

```
http://localhost:8080/api/v1/access-requests
```

## 📊 Request Types

| Type | Description | Use Case |
|------|-------------|----------|
| `ROLE_ASSIGNMENT` | Assign roles to users | User promotion, role changes |
| `PERMISSION_GRANT` | Grant specific permissions | Temporary access needs |
| `RESOURCE_ACCESS` | Access to specific resources | Document access, system access |
| `ELEVATION` | Privilege elevation | Emergency access, admin tasks |

## 🔄 Approval Status

| Status | Description | Actions Available |
|--------|-------------|------------------|
| `PENDING` | Awaiting approval | Approve, Reject |
| `APPROVED` | Request approved | Revoke, Execute |
| `REJECTED` | Request rejected | None |
| `EXPIRED` | Request expired | None |
| `REVOKED` | Request revoked | None |
| `EXECUTED` | Access granted | Revoke |

## 🚀 Quick Start

### 1. Health Check
```bash
curl -X GET http://localhost:8080/health | jq .
```

### 2. Create Access Request
```bash
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
```

### 3. Process Access Request
```bash
curl -X POST http://localhost:8080/api/v1/access-requests/{id}/process \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 00000000-0000-0000-0000-000000000002" \
  -d '{
    "action": "approve",
    "comments": "Approved for business justification",
    "duration_hours": 24
  }' | jq .
```

### 4. List Access Requests
```bash
curl -X GET http://localhost:8080/api/v1/access-requests | jq .
```

## 📚 Documentation Files

- **[curl-examples.md](curl-examples.md)** -  curl command examples
- **[API Reference](api-reference.md)** - Detailed API specification
- **[Schema Reference](schema-reference.md)** - Request/response schemas
- **[Workflow Guide](workflow-guide.md)** - Approval workflow configuration
- **Integration Guide** - Integration with other systems

## 🧪 Testing

### Automated Testing
```bash
# Run access request API tests
./utilities/scripts/test-access-requests.sh

# Run with verbose output
./utilities/scripts/test-access-requests.sh -v
```

### Manual Testing
```bash
# Follow step-by-step examples
cat curl-examples.md
```

## 🔍 Common Use Cases

### Role Assignment Request
```bash
# Create role assignment request
curl -X POST http://localhost:8080/api/v1/access-requests \
  -H "Content-Type: application/json" \
  -H "X-User-ID: requester-id" \
  -d '{
    "entity_id": "entity-id",
    "request_type": "ROLE_ASSIGNMENT",
    "role_id": "role-id",
    "target_user_id": "target-user-id",
    "justification": "User needs elevated privileges for project work",
    "business_reason": "Critical project delivery",
    "duration_hours": 168
  }' | jq .
```

### Emergency Access Request
```bash
# Create emergency elevation request
curl -X POST http://localhost:8080/api/v1/access-requests \
  -H "Content-Type: application/json" \
  -H "X-User-ID: requester-id" \
  -d '{
    "entity_id": "entity-id",
    "request_type": "ELEVATION",
    "justification": "Emergency system maintenance required",
    "business_reason": "Critical security patch deployment",
    "duration_hours": 4,
    "auto_revoke": true
  }' | jq .
```

### Approval Workflow
```bash
# 1. Create request
REQUEST_ID=$(curl -s -X POST http://localhost:8080/api/v1/access-requests \
  -H "Content-Type: application/json" \
  -H "X-User-ID: requester-id" \
  -d '{...}' | jq -r '.access_request.id')

# 2. Approve request
curl -X POST http://localhost:8080/api/v1/access-requests/$REQUEST_ID/process \
  -H "Content-Type: application/json" \
  -H "X-User-ID: approver-id" \
  -d '{
    "action": "approve",
    "comments": "Approved for business justification"
  }' | jq .

# 3. Check request status
curl -X GET http://localhost:8080/api/v1/access-requests/$REQUEST_ID | jq .
```

## 🎯 Expected Responses

### Success Response (201 Created)
```json
{
  "message": "Access request created successfully",
  "access_request": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "entity_id": "550e8400-e29b-41d4-a716-446655440001",
    "request_type": "ROLE_ASSIGNMENT",
    "approval_status": "PENDING",
    "requester_id": "550e8400-e29b-41d4-a716-446655440002",
    "target_user_id": "550e8400-e29b-41d4-a716-446655440003",
    "justification": "User needs admin access for project management",
    "business_reason": "Required for Q1 project delivery",
    "duration_hours": 168,
    "auto_revoke": true,
    "created_at": "2024-01-01T10:00:00Z",
    "updated_at": "2024-01-01T10:00:00Z"
  }
}
```

### Process Response (200 OK)
```json
{
  "message": "Access request processed successfully",
  "access_request": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "approval_status": "APPROVED",
    "approved_by": "550e8400-e29b-41d4-a716-446655440004",
    "approved_at": "2024-01-01T10:30:00Z",
    "expires_at": "2024-01-02T10:30:00Z",
    "updated_at": "2024-01-01T10:30:00Z"
  }
}
```

### Statistics Response (200 OK)
```json
{
  "statistics": {
    "total_requests": 150,
    "pending_requests": 12,
    "approved_requests": 120,
    "rejected_requests": 18,
    "expired_requests": 0,
    "avg_approval_time_hours": 6,
    "requests_by_type": {
      "ROLE_ASSIGNMENT": 45,
      "PERMISSION_GRANT": 35,
      "RESOURCE_ACCESS": 50,
      "ELEVATION": 20
    }
  }
}
```

## 🚨 Common Issues

1. **Missing X-User-ID header** - Required for user context
2. **Invalid request type** - Must be one of the supported types
3. **Insufficient permissions** - User may not have approval rights
4. **Request already processed** - Cannot modify processed requests
5. **Expired requests** - Cannot approve expired requests

## 📈 Performance Notes

- Workflow state changes are atomic
- Audit events are logged asynchronously
- Notifications are sent via background jobs
- Statistics are cached for performance

## 🔒 Security Features

- **ABAC Integration**: Attribute-based access control
- **Audit Logging**: Complete audit trail
- **Risk Assessment**: Risk-based approval decisions
- **Auto-Revocation**: Automatic access cleanup
- **Approval Validation**: Strict approval rule enforcement

## 🔗 Related APIs

- **User Management**: Users create and approve requests
- **Entity Management**: Requests are scoped to entities
- **Conditional Access**: Rules may affect request approval
- **User Analytics**: Behavioral data influences risk assessment

## 🎯 Workflow Configuration

### Approval Rules
- **Role-Based**: Approvers based on roles
- **Hierarchy-Based**: Approval based on organizational hierarchy
- **Risk-Based**: Approval requirements based on risk assessment
- **Resource-Based**: Approval rules per resource type

### Notification Configuration
- **Email Notifications**: Automated email alerts
- **Real-time Notifications**: WebSocket notifications
- **Escalation Rules**: Automatic escalation for delays
- **Reminder System**: Periodic reminders for pending requests

## 📊 Monitoring and Metrics

### Key Metrics
- Request creation rate
- Approval processing time
- Rejection reasons
- Auto-revocation statistics
- Risk score distributions

### Alerts
- Pending request backlog
- High-risk requests
- Approval SLA violations
- System errors and failures