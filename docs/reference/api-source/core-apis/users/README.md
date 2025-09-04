# User Management API

The User Management API provides  user authentication, authorization, and management capabilities following a clean architecture pattern.

## 📋 Overview

The User Management API handles all user-related operations including:

- **User CRUD Operations**: Create, read, update, delete users
- **Authentication**: Email/username and password authentication
- **Authorization**: Role-based access control
- **User Types**: Internal, external, admin, customer, supplier users
- **Account Management**: Account status, MFA settings, session management
- **Caching**: Redis-based caching for performance optimization
- **Security**: Password hashing, session management, audit logging

## 🔧 Base URL

```
http://localhost:8080/api/v1/users
```

## 👥 User Types

| Type | Description | Use Case |
|------|-------------|----------|
| `INTERNAL` | Internal employees | Company staff |
| `EXTERNAL` | External contractors | Temporary workers |
| `ADMIN` | System administrators | System management |
| `CUSTOMER` | Customer users | Client access |
| `SUPPLIER` | Supplier users | Vendor access |
| `PARTNER` | Partner users | Partnership access |

## 🔐 Account Status

| Status | Description | Access |
|--------|-------------|--------|
| `ACTIVE` | Active account | Full access |
| `INACTIVE` | Inactive account | No access |
| `SUSPENDED` | Suspended account | Limited access |
| `LOCKED` | Locked account | No access |
| `PENDING` | Pending activation | Limited access |

## 🚀 Quick Start

### 1. Health Check
```bash
curl -X GET http://localhost:8080/health | jq .
```

### 2. Create User
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "entity_id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "john.doe",
    "email": "john.doe@example.com",
    "password": "SecurePassword123!",
    "user_type": "INTERNAL",
    "account_status": "ACTIVE",
    "mfa_enabled": false
  }' | jq .
```

### 3. Authenticate User
```bash
curl -X POST http://localhost:8080/api/v1/users/auth \
  -H "Content-Type: application/json" \
  -d '{
    "identifier": "john.doe@example.com",
    "password": "SecurePassword123!"
  }' | jq .
```

### 4. List Users
```bash
curl -X GET http://localhost:8080/api/v1/users | jq .
```

## 📚 Documentation Files

- **[curl Examples](./curl-examples.md)** - Comprehensive curl command examples

## 🧪 Testing

### Automated Testing
```bash
# Run user API tests
./utilities/scripts/test-users.sh

# Run with verbose output
./utilities/scripts/test-users.sh -v

# Run without cleanup
./utilities/scripts/test-users.sh --no-cleanup
```

### Manual Testing
```bash
# Follow step-by-step examples
cat curl-examples.md
```

## 🔍 Common Use Cases

### User Registration Flow
```bash
# 1. Create user
USER_ID=$(curl -s -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "entity_id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "new.user",
    "email": "new.user@example.com",
    "password": "SecurePassword123!",
    "user_type": "INTERNAL"
  }' | jq -r '.id')

# 2. Authenticate user
curl -X POST http://localhost:8080/api/v1/users/auth \
  -H "Content-Type: application/json" \
  -d '{
    "identifier": "new.user@example.com",
    "password": "SecurePassword123!"
  }' | jq .

# 3. Get user details
curl -X GET http://localhost:8080/api/v1/users/$USER_ID | jq .
```

### Password Management
```bash
# Update password
curl -X PUT http://localhost:8080/api/v1/users/$USER_ID/password \
  -H "Content-Type: application/json" \
  -d '{
    "current_password": "SecurePassword123!",
    "new_password": "NewSecurePassword456!"
  }' | jq .
```

### User Search
```bash
# Search users by query
curl -X GET "http://localhost:8080/api/v1/users/search?q=john&limit=10" | jq .

# Search with filters
curl -X GET "http://localhost:8080/api/v1/users?user_type=INTERNAL&account_status=ACTIVE" | jq .
```

## 🎯 Expected Responses

### Success Response (201 Created)
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "entity_id": "550e8400-e29b-41d4-a716-446655440001",
  "username": "john.doe",
  "email": "john.doe@example.com",
  "user_type": "INTERNAL",
  "account_status": "ACTIVE",
  "is_active": true,
  "mfa_enabled": false,
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T10:00:00Z"
}
```

### Authentication Response (200 OK)
```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "john.doe",
    "email": "john.doe@example.com",
    "user_type": "INTERNAL",
    "is_active": true
  },
  "token": "jwt-token-placeholder",
  "expires_at": "2024-01-01T18:00:00Z"
}
```

### Error Response (401 Unauthorized)
```json
{
  "error": "Invalid credentials",
  "message": "The provided credentials are incorrect"
}
```

## 🚨 Common Issues

1. **Duplicate email/username** - Email and username must be unique
2. **Invalid credentials** - Check email/username and password
3. **Account locked** - Account may be suspended or locked
4. **Password requirements** - Passwords must meet security criteria
5. **Entity not found** - User must belong to valid entity

## 📈 Performance Notes

- User data is cached for 15 minutes
- Authentication responses are optimized
- Database queries use indexes for performance
- Pagination is implemented for large result sets

## 🔒 Security Features

- **Password Hashing**: bcrypt with salt
- **Session Management**: JWT tokens with expiration
- **MFA Support**: Multi-factor authentication
- **Account Lockout**: Automatic lockout after failed attempts
- **Audit Logging**: All user actions are logged
- **Rate Limiting**: API rate limiting for security

## 🔗 Related APIs

- **Entity Management**: Users belong to entities
- **Access Requests**: Users can request access
- **User Analytics**: Behavioral analysis and risk assessment
- **Tenant Management**: Users belong to tenants