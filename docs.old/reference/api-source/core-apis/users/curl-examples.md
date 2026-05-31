# User API Testing Guide with Curl

This guide provides  curl commands to test the User Management API following the data flow pattern.

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

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/users/auth` | Authenticate user |
| POST | `/api/v1/users` | Create user |
| GET | `/api/v1/users` | List users |
| GET | `/api/v1/users/search` | Search users |
| GET | `/api/v1/users/{id}` | Get user by ID |
| PUT | `/api/v1/users/{id}` | Update user |
| DELETE | `/api/v1/users/{id}` | Delete user |
| PUT | `/api/v1/users/{id}/password` | Update password |
| GET | `/api/v1/users/{id}/roles` | Get user roles |

##  Health Check

First, verify the server is running:

```bash
# Health check
curl -X GET http://localhost:8080/health | jq .

# Expected response:
# {
#   "status": "ok",
#   "timestamp": "2024-01-01T10:00:00Z"
# }
```

##  User Management Tests

### 1. Create User

```bash
# Create a test user
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

# Expected response (201 Created):
# {
#   "id": "uuid-here",
#   "entity_id": "550e8400-e29b-41d4-a716-446655440000",
#   "username": "john.doe",
#   "email": "john.doe@example.com",
#   "user_type": "INTERNAL",
#   "account_status": "ACTIVE",
#   "is_active": true,
#   "mfa_enabled": false,
#   "created_at": "2024-01-01T10:00:00Z",
#   "updated_at": "2024-01-01T10:00:00Z"
# }
```

### 2. Create Additional Test Users

```bash
# Create admin user
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "entity_id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "admin",
    "email": "admin@example.com",
    "password": "AdminPassword123!",
    "user_type": "ADMIN",
    "account_status": "ACTIVE",
    "mfa_enabled": true,
    "session_timeout_minutes": 480
  }' | jq .

# Create customer user
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "entity_id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "customer1",
    "email": "customer@example.com",
    "password": "CustomerPass123!",
    "user_type": "CUSTOMER",
    "account_status": "ACTIVE",
    "mfa_enabled": false
  }' | jq .
```

### 3. Get User by ID

```bash
# Replace USER_ID with actual ID from create response
USER_ID="your-user-id-here"

curl -X GET http://localhost:8080/api/v1/users/$USER_ID | jq .

# Expected response (200 OK):
# {
#   "id": "user-id",
#   "entity_id": "entity-id",
#   "username": "john.doe",
#   "email": "john.doe@example.com",
#   "user_type": "INTERNAL",
#   "account_status": "ACTIVE",
#   "is_active": true,
#   "created_at": "2024-01-01T10:00:00Z",
#   "updated_at": "2024-01-01T10:00:00Z"
# }
```

### 4. List Users

```bash
# List all users
curl -X GET http://localhost:8080/api/v1/users | jq .

# List with pagination
curl -X GET "http://localhost:8080/api/v1/users?limit=10&offset=0" | jq .

# Filter by user type
curl -X GET "http://localhost:8080/api/v1/users?user_type=INTERNAL&limit=5" | jq .

# Filter by account status
curl -X GET "http://localhost:8080/api/v1/users?account_status=ACTIVE&limit=10" | jq .

# Expected response:
# {
#   "users": [
#     {
#       "id": "user-id",
#       "username": "john.doe",
#       "email": "john.doe@example.com",
#       "user_type": "INTERNAL",
#       "is_active": true,
#       "created_at": "2024-01-01T10:00:00Z"
#     }
#   ],
#   "total": 1,
#   "limit": 20,
#   "offset": 0
# }
```

### 5. Update User

```bash
# Update user information
curl -X PUT http://localhost:8080/api/v1/users/$USER_ID \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john.doe.updated",
    "email": "john.doe.updated@example.com",
    "account_status": "ACTIVE",
    "mfa_enabled": true,
    "session_timeout_minutes": 240
  }' | jq .

# Expected response (200 OK):
# Updated user object with new values
```

### 6. Search Users

```bash
# Search by username/email
curl -X GET "http://localhost:8080/api/v1/users/search?q=john&limit=10" | jq .

# Search with pagination
curl -X GET "http://localhost:8080/api/v1/users/search?q=admin&limit=5&offset=0" | jq .

# Expected response:
# {
#   "users": [
#     {
#       "id": "user-id",
#       "username": "john.doe",
#       "email": "john.doe@example.com",
#       "user_type": "INTERNAL"
#     }
#   ],
#   "total": 1,
#   "limit": 10,
#   "offset": 0,
#   "query": "john"
# }
```

##  Authentication Tests

### 7. Authenticate User

```bash
# Authenticate with email
curl -X POST http://localhost:8080/api/v1/users/auth \
  -H "Content-Type: application/json" \
  -d '{
    "identifier": "john.doe@example.com",
    "password": "SecurePassword123!"
  }' | jq .

# Authenticate with username
curl -X POST http://localhost:8080/api/v1/users/auth \
  -H "Content-Type: application/json" \
  -d '{
    "identifier": "john.doe",
    "password": "SecurePassword123!"
  }' | jq .

# Expected response (200 OK):
# {
#   "user": {
#     "id": "user-id",
#     "username": "john.doe",
#     "email": "john.doe@example.com",
#     "user_type": "INTERNAL",
#     "is_active": true
#   },
#   "token": "jwt-token-placeholder",
#   "expires_at": "2024-01-01T18:00:00Z"
# }
```

### 8. Update Password

```bash
# Update user password
curl -X PUT http://localhost:8080/api/v1/users/$USER_ID/password \
  -H "Content-Type: application/json" \
  -d '{
    "current_password": "SecurePassword123!",
    "new_password": "NewSecurePassword456!"
  }' | jq .

# Expected response (200 OK):
# {
#   "message": "Password updated successfully"
# }
```

##  Role Management Tests

### 9. Get User Roles

```bash
# Get roles for a user
curl -X GET http://localhost:8080/api/v1/users/$USER_ID/roles | jq .

# Expected response (200 OK):
# {
#   "roles": [
#     {
#       "id": "role-id",
#       "name": "user",
#       "display_name": "Standard User",
#       "description": "Standard user role",
#       "role_type": "STANDARD",
#       "assigned_at": "2024-01-01T10:00:00Z"
#     }
#   ]
# }
```

### 10. Delete User (Soft Delete)

```bash
# Soft delete user
curl -X DELETE http://localhost:8080/api/v1/users/$USER_ID | jq .

# Expected response (200 OK):
# {
#   "message": "User deleted successfully"
# }

# Verify user is soft deleted (should return 404)
curl -X GET http://localhost:8080/api/v1/users/$USER_ID | jq .
```

## ❌ Error Scenarios Testing

### 11. Invalid Requests

```bash
# Create user with missing required fields
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "incomplete"
  }' | jq .

# Expected response (400 Bad Request):
# {
#   "error": "Invalid request payload"
# }

# Get non-existent user
curl -X GET http://localhost:8080/api/v1/users/invalid-uuid | jq .

# Expected response (400 Bad Request):
# {
#   "error": "Invalid user ID"
# }

# Authenticate with wrong credentials
curl -X POST http://localhost:8080/api/v1/users/auth \
  -H "Content-Type: application/json" \
  -d '{
    "identifier": "john.doe@example.com",
    "password": "wrongpassword"
  }' | jq .

# Expected response (401 Unauthorized):
# {
#   "error": "Invalid credentials"
# }
```

### 12. Duplicate Email/Username

```bash
# Try to create user with existing email
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "entity_id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "duplicate.user",
    "email": "john.doe@example.com",
    "password": "AnotherPassword123!",
    "user_type": "INTERNAL"
  }' | jq .

# Expected response (409 Conflict):
# {
#   "error": "Email already exists"
# }
```

##  Performance and Monitoring

### 13. Test with Verbose Output

```bash
# Test with timing information
curl -w "@-" -o /dev/null -s -X GET http://localhost:8080/api/v1/users <<'EOF'
     time_namelookup:  %{time_namelookup}\n
        time_connect:  %{time_connect}\n
     time_appconnect:  %{time_appconnect}\n
    time_pretransfer:  %{time_pretransfer}\n
       time_redirect:  %{time_redirect}\n
  time_starttransfer:  %{time_starttransfer}\n
                     ----------\n
          time_total:  %{time_total}\n
EOF
```

### 14. Test Cache Performance

```bash
# First request (cache miss)
time curl -X GET http://localhost:8080/api/v1/users/$USER_ID | jq .

# Second request (cache hit - should be faster)
time curl -X GET http://localhost:8080/api/v1/users/$USER_ID | jq .
```

##  Complete Test Sequence

Here's a complete test sequence that exercises all major functionality:

```bash
#!/bin/bash

# Set base URL
BASE_URL="http://localhost:8080"

echo " Starting User API Test Suite..."

# 1. Health check
echo "1. Health check..."
curl -s $BASE_URL/health | jq .

# 2. Create test entity first (required for users)
ENTITY_ID=$(curl -s -X POST $BASE_URL/api/v1/entities \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Company",
    "code": "TEST001",
    "type": "account"
  }' | jq -r '.id')

echo "Created entity: $ENTITY_ID"

# 3. Create users
echo "2. Creating test users..."
USER1_ID=$(curl -s -X POST $BASE_URL/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "entity_id": "'$ENTITY_ID'",
    "username": "test.user1",
    "email": "test1@example.com",
    "password": "TestPassword123!",
    "user_type": "INTERNAL"
  }' | jq -r '.id')

echo "Created user 1: $USER1_ID"

# 4. Test authentication
echo "3. Testing authentication..."
curl -s -X POST $BASE_URL/api/v1/users/auth \
  -H "Content-Type: application/json" \
  -d '{
    "identifier": "test1@example.com",
    "password": "TestPassword123!"
  }' | jq .

# 5. Test user retrieval
echo "4. Testing user retrieval..."
curl -s -X GET $BASE_URL/api/v1/users/$USER1_ID | jq .

# 6. Test user update
echo "5. Testing user update..."
curl -s -X PUT $BASE_URL/api/v1/users/$USER1_ID \
  -H "Content-Type: application/json" \
  -d '{
    "mfa_enabled": true
  }' | jq .

# 7. Test password update
echo "6. Testing password update..."
curl -s -X PUT $BASE_URL/api/v1/users/$USER1_ID/password \
  -H "Content-Type: application/json" \
  -d '{
    "current_password": "TestPassword123!",
    "new_password": "NewTestPassword456!"
  }' | jq .

# 8. Test search
echo "7. Testing user search..."
curl -s -X GET "$BASE_URL/api/v1/users/search?q=test" | jq .

# 9. Test list users
echo "8. Testing list users..."
curl -s -X GET "$BASE_URL/api/v1/users?limit=10" | jq .

echo "✅ User API test suite completed!"
```

Save this as `test-user-api.sh` and run with:
```bash
chmod +x test-user-api.sh
./test-user-api.sh
```

##  Notes

- Replace UUIDs with actual values from your responses
- All passwords must meet security requirements (8+ characters)
- User codes and emails must be unique within a tenant
- The API includes  observability (logging, tracing, metrics)
- All endpoints return appropriate HTTP status codes
- Caching is implemented for performance optimization

##  Troubleshooting

1. **Server not responding**: Check if server is running on port 8080
2. **Database errors**: Verify database connection and migrations
3. **Validation errors**: Check required fields and data formats
4. **UUID errors**: Ensure UUIDs are valid format
5. **Cache issues**: Redis connection might be required for full functionality