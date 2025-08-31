# Users API

## Overview

The Users API provides comprehensive user management capabilities for AWO ERP, including user creation, profile management, attribute handling, and user analytics. This API supports multi-tenant user management with role-based access control and advanced user behavior analytics.

## Key Features

- **User Lifecycle Management**: Create, update, archive, and restore user accounts
- **Profile Management**: Comprehensive user profile with custom attributes
- **Bulk Operations**: Efficient bulk user operations and attribute management
- **User Analytics**: Behavior tracking, anomaly detection, and risk assessment
- **Multi-tenant Support**: Tenant-scoped user management and isolation
- **Attribute System**: Dynamic user attributes with type safety

## Endpoints Overview

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/v1/users` | List users with pagination and filtering |
| POST | `/api/v1/users` | Create new user account |
| GET | `/api/v1/users/{id}` | Get user details by ID |
| PUT | `/api/v1/users/{id}` | Update user information |
| DELETE | `/api/v1/users/{id}` | Archive user account |
| GET | `/api/v1/users/{id}/attributes` | Get user attributes |
| PUT | `/api/v1/users/{id}/attributes` | Update user attributes |
| POST | `/api/v1/users/bulk-attributes` | Bulk attribute operations |
| GET | `/api/v1/users/profile` | Get current user profile |

## User Management

### List Users
```bash
curl -X GET "http://localhost:8080/api/v1/users" \
  -H "Authorization: Bearer $TOKEN" \
  -G \
  -d "page=1" \
  -d "limit=20" \
  -d "status=active" \
  -d "search=john"
```

**Response:**
```json
{
  "data": [
    {
      "id": "user-123",
      "username": "john.doe@company.com",
      "email": "john.doe@company.com", 
      "first_name": "John",
      "last_name": "Doe",
      "status": "active",
      "roles": ["user", "developer"],
      "department": "Engineering",
      "created_at": "2025-01-01T10:00:00Z",
      "updated_at": "2025-01-15T14:30:00Z",
      "last_login": "2025-01-30T09:15:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "pages": 8
  }
}
```

### Create User
```bash
curl -X POST "http://localhost:8080/api/v1/users" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "jane.smith@company.com",
    "email": "jane.smith@company.com",
    "first_name": "Jane",
    "last_name": "Smith", 
    "password": "secure-temp-password",
    "roles": ["user", "analyst"],
    "department": "Finance",
    "attributes": {
      "employee_id": {"value": "EMP-001", "type": "string"},
      "clearance_level": {"value": "confidential", "type": "string"},
      "start_date": {"value": "2025-02-01", "type": "date"}
    },
    "send_welcome_email": true
  }'
```

### Update User
```bash
curl -X PUT "http://localhost:8080/api/v1/users/user-123" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "John",
    "last_name": "Doe-Smith",
    "department": "Platform Engineering",
    "roles": ["user", "developer", "team_lead"]
  }'
```

### Get User Profile
```bash
curl -X GET "http://localhost:8080/api/v1/users/profile" \
  -H "Authorization: Bearer $TOKEN"
```

## User Attributes System

### Get User Attributes
```bash
curl -X GET "http://localhost:8080/api/v1/users/user-123/attributes" \
  -H "Authorization: Bearer $TOKEN"
```

**Response:**
```json
{
  "data": {
    "employee_id": {
      "value": "EMP-001",
      "type": "string",
      "source": "hr_system",
      "updated_at": "2025-01-01T10:00:00Z"
    },
    "clearance_level": {
      "value": "confidential", 
      "type": "string",
      "source": "security_system",
      "updated_at": "2025-01-15T14:30:00Z"
    },
    "cost_center": {
      "value": "CC-ENG-001",
      "type": "string",
      "source": "finance_system",
      "updated_at": "2025-01-20T09:00:00Z"
    },
    "manager_id": {
      "value": "user-456",
      "type": "user_reference",
      "source": "org_chart",
      "updated_at": "2025-01-01T10:00:00Z"
    }
  }
}
```

### Update User Attributes
```bash
curl -X PUT "http://localhost:8080/api/v1/users/user-123/attributes" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "attributes": {
      "clearance_level": {
        "value": "secret",
        "type": "string",
        "source": "security_review"
      },
      "project_assignments": {
        "value": ["proj-alpha", "proj-beta"],
        "type": "string_array",
        "source": "project_management"
      }
    }
  }'
```

### Bulk Attribute Operations
```bash
curl -X POST "http://localhost:8080/api/v1/users/bulk-attributes" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "operation": "update",
    "filter": {
      "department": "Engineering",
      "status": "active"
    },
    "attributes": {
      "security_training_completed": {
        "value": "2025-01-30",
        "type": "date",
        "source": "training_system"
      }
    }
  }'
```

## User Analytics

### User Behavior Analysis
```bash
curl -X GET "http://localhost:8080/api/v1/analytics/users/user-123/behavior" \
  -H "Authorization: Bearer $TOKEN" \
  -G \
  -d "period=30d" \
  -d "include_details=true"
```

**Response:**
```json
{
  "user_id": "user-123",
  "period": "30d",
  "summary": {
    "total_sessions": 45,
    "total_duration_hours": 180,
    "average_session_duration_minutes": 240,
    "unique_resources_accessed": 125,
    "api_calls": 2500
  },
  "patterns": {
    "most_active_hours": [9, 10, 11, 14, 15, 16],
    "most_active_days": ["monday", "tuesday", "wednesday"],
    "primary_activities": ["document_access", "report_generation", "api_usage"],
    "resource_categories": {
      "financial_reports": 40,
      "user_management": 25,
      "system_admin": 35
    }
  },
  "trends": {
    "activity_trend": "increasing",
    "security_score": 85,
    "compliance_score": 92
  }
}
```

### Anomaly Detection
```bash
curl -X GET "http://localhost:8080/api/v1/analytics/users/user-123/detect-anomalies" \
  -H "Authorization: Bearer $TOKEN" \
  -G \
  -d "sensitivity=medium" \
  -d "lookback_days=7"
```

**Response:**
```json
{
  "user_id": "user-123",
  "analysis_period": "7d",
  "anomalies_detected": 2,
  "anomalies": [
    {
      "type": "unusual_access_time",
      "severity": "medium",
      "description": "User accessed system at 2:30 AM, outside normal hours",
      "timestamp": "2025-01-29T02:30:00Z",
      "confidence": 0.85,
      "context": {
        "normal_hours": "08:00-18:00",
        "access_time": "02:30",
        "resource": "financial_reports"
      }
    },
    {
      "type": "elevated_privilege_usage", 
      "severity": "high",
      "description": "Unusual spike in admin operations",
      "timestamp": "2025-01-28T14:15:00Z",
      "confidence": 0.92,
      "context": {
        "normal_admin_ops_per_day": 5,
        "detected_admin_ops": 25,
        "operations": ["user_creation", "role_assignment"]
      }
    }
  ],
  "risk_assessment": {
    "overall_risk_score": 65,
    "risk_factors": ["off_hours_access", "privilege_escalation"],
    "recommended_actions": [
      "Review access logs",
      "Verify legitimacy of admin operations",
      "Consider temporary access restrictions"
    ]
  }
}
```

### User Risk Assessment
```bash
curl -X GET "http://localhost:8080/api/v1/analytics/users/user-123/risk" \
  -H "Authorization: Bearer $TOKEN"
```

### User Insights
```bash
curl -X GET "http://localhost:8080/api/v1/analytics/users/user-123/insights" \
  -H "Authorization: Bearer $TOKEN" \
  -G \
  -d "include_recommendations=true"
```

## Advanced Filtering and Search

### Complex User Queries
```bash
curl -X GET "http://localhost:8080/api/v1/users" \
  -H "Authorization: Bearer $TOKEN" \
  -G \
  -d "filter[department]=Engineering,Finance" \
  -d "filter[roles]=developer" \
  -d "filter[status]=active" \
  -d "filter[created_after]=2025-01-01" \
  -d "sort=last_login:desc" \
  -d "search=john" \
  -d "include_attributes=clearance_level,cost_center"
```

### User Attribute Queries
```bash
curl -X GET "http://localhost:8080/api/v1/users" \
  -H "Authorization: Bearer $TOKEN" \
  -G \
  -d "attribute_filter[clearance_level]=confidential,secret" \
  -d "attribute_filter[security_training_completed]=>2025-01-01" \
  -d "has_attribute=manager_id"
```

## Error Handling

### Common User API Errors

#### User Not Found
```json
{
  "error": {
    "code": "USER_NOT_FOUND",
    "message": "User with ID 'user-123' not found",
    "details": {"user_id": "user-123"}
  }
}
```

#### Duplicate Username
```json
{
  "error": {
    "code": "USERNAME_CONFLICT",
    "message": "Username 'john.doe@company.com' already exists",
    "details": {"field": "username"}
  }
}
```

#### Invalid Attribute Type
```json
{
  "error": {
    "code": "INVALID_ATTRIBUTE_TYPE",
    "message": "Attribute 'start_date' must be of type 'date'",
    "details": {
      "attribute": "start_date",
      "expected_type": "date",
      "provided_type": "string"
    }
  }
}
```

#### Insufficient Permissions
```json
{
  "error": {
    "code": "INSUFFICIENT_PERMISSIONS",
    "message": "User management requires 'admin:users' permission",
    "details": {
      "required_permission": "admin:users",
      "user_permissions": ["read:users"]
    }
  }
}
```

## Best Practices

### User Creation
- **Strong Passwords**: Enforce password complexity requirements
- **Email Verification**: Require email verification for new accounts
- **Role Assignment**: Follow principle of least privilege
- **Audit Trail**: Log all user creation and modification events

### Attribute Management
- **Type Safety**: Use proper attribute types (string, integer, date, etc.)
- **Source Tracking**: Always specify attribute sources for auditing
- **Validation**: Validate attribute values against business rules
- **Bulk Operations**: Use bulk endpoints for efficiency when updating many users

### Security Considerations
- **Data Minimization**: Only collect necessary user information
- **Encryption**: Encrypt sensitive attributes at rest
- **Access Logging**: Log all user data access for audit trails
- **Regular Cleanup**: Archive or delete unused user accounts

## Integration Patterns

### User Synchronization
```python
def sync_users_from_hr_system(api_client, hr_users):
    """Sync users from HR system with proper error handling."""
    results = {
        'created': 0,
        'updated': 0,
        'errors': []
    }
    
    for hr_user in hr_users:
        try:
            # Check if user exists
            existing_user = api_client.get_user_by_email(hr_user['email'])
            
            if existing_user:
                # Update existing user
                api_client.update_user(existing_user['id'], {
                    'first_name': hr_user['first_name'],
                    'last_name': hr_user['last_name'],
                    'department': hr_user['department'],
                    'attributes': {
                        'employee_id': {'value': hr_user['emp_id'], 'type': 'string'},
                        'manager_id': {'value': hr_user['manager_id'], 'type': 'user_reference'}
                    }
                })
                results['updated'] += 1
            else:
                # Create new user
                api_client.create_user({
                    'username': hr_user['email'],
                    'email': hr_user['email'],
                    'first_name': hr_user['first_name'],
                    'last_name': hr_user['last_name'],
                    'department': hr_user['department'],
                    'roles': ['user'],
                    'attributes': {
                        'employee_id': {'value': hr_user['emp_id'], 'type': 'string'},
                        'hire_date': {'value': hr_user['start_date'], 'type': 'date'}
                    },
                    'send_welcome_email': True
                })
                results['created'] += 1
                
        except Exception as e:
            results['errors'].append({
                'user': hr_user['email'],
                'error': str(e)
            })
    
    return results
```

### User Analytics Dashboard
```python
def generate_user_analytics_dashboard(api_client, user_ids):
    """Generate comprehensive user analytics dashboard."""
    dashboard_data = {
        'users': [],
        'summary': {
            'total_users': len(user_ids),
            'high_risk_users': 0,
            'anomalies_detected': 0,
            'avg_activity_score': 0
        }
    }
    
    activity_scores = []
    
    for user_id in user_ids:
        # Get user behavior data
        behavior = api_client.get_user_behavior(user_id, period='30d')
        
        # Get risk assessment
        risk = api_client.get_user_risk(user_id)
        
        # Get recent anomalies
        anomalies = api_client.detect_user_anomalies(user_id, lookback_days=7)
        
        user_data = {
            'user_id': user_id,
            'activity_score': behavior['summary']['total_sessions'],
            'risk_score': risk['overall_risk_score'],
            'anomaly_count': anomalies['anomalies_detected'],
            'last_activity': behavior.get('last_activity_at')
        }
        
        dashboard_data['users'].append(user_data)
        activity_scores.append(user_data['activity_score'])
        
        if risk['overall_risk_score'] >= 70:
            dashboard_data['summary']['high_risk_users'] += 1
            
        dashboard_data['summary']['anomalies_detected'] += user_data['anomaly_count']
    
    dashboard_data['summary']['avg_activity_score'] = sum(activity_scores) / len(activity_scores)
    
    return dashboard_data
```

---

**Next**: [Organizations API](../organizations/) | **Up**: [API Reference](../index.md)