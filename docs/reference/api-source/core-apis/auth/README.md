# Authentication API

The Authentication API provides comprehensive authentication and authorization functionality for the AWO ERP System. This API manages user authentication, JWT token generation, token validation, and session management.

## 📋 Overview

**Base URL**: `/api/v1/auth`

**Purpose**: Secure authentication and authorization services with JWT token-based authentication, session management, and user validation.

**Key Features**:
- User login with email/password
- JWT token generation and refresh
- Token validation and user session management
- Secure logout with token invalidation
- Multi-factor authentication support (future)

## 🔗 Available Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/auth/login` | Authenticate user and get JWT tokens |
| `POST` | `/api/v1/auth/refresh` | Refresh JWT access token |
| `POST` | `/api/v1/auth/logout` | Logout user and invalidate tokens |
| `GET` | `/api/v1/auth/validate` | Validate JWT token and get user info |

## 📊 Data Models

### Login Request
```json
{
  "email": "string (email format)",
  "password": "string (min 8 characters)"
}
```

### Authentication Response
```json
{
  "access_token": "string (JWT)",
  "token_type": "string (Bearer)",
  "expires_in": "integer (seconds)",
  "user": {
    "id": "string (UUID)",
    "email": "string",
    "first_name": "string",
    "last_name": "string"
  },
  "permissions": ["string"]
}
```

### Refresh Request
```json
{
  "refresh_token": "string (JWT)"
}
```

### Token Validation Response
```json
{
  "valid": "boolean",
  "user": {
    "id": "string (UUID)", 
    "email": "string",
    "first_name": "string",
    "last_name": "string"
  },
  "permissions": ["string"]
}
```

### User Info Object
```json
{
  "id": "string (UUID)",
  "email": "string",
  "first_name": "string",
  "last_name": "string"
}
```

## 🚀 Quick Examples

### User Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePassword123!"
  }' | jq .
```

**Response:**
```json
{
  "access_token": "mock-jwt-token",
  "token_type": "Bearer",
  "expires_in": 3600,
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe"
  },
  "permissions": ["read:users", "write:entities"]
}
```

### Token Refresh
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "mock-refresh-token"
  }' | jq .
```

**Response:**
```json
{
  "access_token": "refreshed-jwt-token",
  "token_type": "Bearer", 
  "expires_in": 3600
}
```

### Token Validation
```bash
curl -X GET http://localhost:8080/api/v1/auth/validate \
  -H "Authorization: Bearer <your-access-token>" | jq .
```

**Response:**
```json
{
  "valid": true,
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com", 
    "first_name": "John",
    "last_name": "Doe"
  },
  "permissions": ["read:users", "write:entities"]
}
```

### User Logout
```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer <your-access-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "token": "<your-access-token>"
  }'
```

## 🔧 Advanced Usage

### Authentication Flow

1. **Login**: POST to `/api/v1/auth/login` with email/password
2. **Store Tokens**: Save access_token and refresh_token securely
3. **API Calls**: Include `Authorization: Bearer <access_token>` header
4. **Token Refresh**: When access_token expires, use refresh_token
5. **Logout**: POST to `/api/v1/auth/logout` to invalidate tokens

### JWT Token Usage

#### Include in Headers
```bash
# Standard Bearer token format
curl -X GET http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer <your-jwt-token>"
```

#### Token Expiration Handling
```bash
# Check token validity before important operations
curl -X GET http://localhost:8080/api/v1/auth/validate \
  -H "Authorization: Bearer <your-jwt-token>" | jq '.valid'
```

### Security Headers

| Header | Required | Description |
|--------|----------|-------------|
| `Content-Type` | Yes | Must be `application/json` for POST requests |
| `Authorization` | Conditional | Required for protected endpoints |

## 🚨 Error Handling

### Common Error Responses

| Status Code | Error Type | Description |
|-------------|------------|-------------|
| `400` | `bad_request` | Invalid email format or missing fields |
| `401` | `unauthorized` | Invalid credentials or expired token |
| `403` | `forbidden` | User account disabled or insufficient permissions |
| `429` | `too_many_requests` | Rate limit exceeded |
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

#### Invalid Credentials
```json
{
  "error": "unauthorized",
  "message": "Invalid email or password"
}
```

#### Expired Token
```json
{
  "error": "unauthorized", 
  "message": "Token has expired"
}
```

#### Rate Limit Exceeded
```json
{
  "error": "too_many_requests",
  "message": "Rate limit exceeded. Try again in 60 seconds"
}
```

## ✅ Validation Rules

### Login Request
- **email**: Required, valid email format, max 255 characters
- **password**: Required, min 8 characters, max 128 characters

### Token Requirements
- **access_token**: Valid JWT format, not expired
- **refresh_token**: Valid JWT format, not expired, not revoked

## 🔒 Security Considerations

### Password Requirements
- Minimum 8 characters
- Must contain uppercase and lowercase letters
- Must contain at least one number
- Must contain at least one special character
- Cannot be a common password

### Token Security
- Access tokens expire in 1 hour (3600 seconds)
- Refresh tokens expire in 30 days
- Tokens are invalidated on logout
- Failed login attempts are rate limited

### Rate Limiting
- Login attempts: 5 per minute per IP
- Token validation: 100 per minute per user
- Password reset: 3 per hour per email

## 🔍 Testing

### Basic Authentication Test
```bash
# Test complete authentication flow
echo "=== Authentication Flow Test ==="

# 1. Login
echo "1. Logging in..."
AUTH_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "password123"}')

ACCESS_TOKEN=$(echo $AUTH_RESPONSE | jq -r '.access_token')
echo "Access token: ${ACCESS_TOKEN:0:20}..."

# 2. Validate token
echo "2. Validating token..."
curl -X GET http://localhost:8080/api/v1/auth/validate \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .

# 3. Use token for API call
echo "3. Making authenticated API call..."
curl -X GET http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .

# 4. Logout
echo "4. Logging out..."
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"token\": \"$ACCESS_TOKEN\"}"

echo "=== Authentication test completed ==="
```

### Error Handling Test
```bash
# Test invalid credentials
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "invalid@example.com", "password": "wrongpassword"}' | jq .

# Test invalid token format
curl -X GET http://localhost:8080/api/v1/auth/validate \
  -H "Authorization: Bearer <invalid-token>" | jq .

# Test missing authorization header
curl -X GET http://localhost:8080/api/v1/auth/validate | jq .
```

## 📊 Performance Considerations

### Response Times
- Login: < 500ms
- Token validation: < 100ms
- Token refresh: < 200ms
- Logout: < 100ms

### Caching Strategy
- User permissions cached for 15 minutes
- Token blacklist cached until expiration
- Failed login attempts tracked for rate limiting

### Optimization Tips
- Validate tokens locally when possible
- Cache user permissions to reduce database queries
- Use refresh tokens to minimize login frequency
- Implement proper token storage in client applications

## 🔗 Related APIs

- **[User Management API](../users/README.md)** - User account management
- **[Tenant Management API](../tenants/README.md)** - Multi-tenant authentication
- **[Organization API](../organizations/README.md)** - Organization-scoped permissions

## 📚 Additional Resources

- **[curl Examples](./curl-examples.md)** - Complete curl command examples
- **[Testing Scripts](../../utilities/scripts/test-auth.sh)** - Automated testing script
- **[Security Guide](../../../../contributing/securityHandbook.md)** - Security best practices

---

**Last Updated**: 2025-07-19  
**API Version**: 1.0.0  
**Maintained by**: AWO ERP System Team