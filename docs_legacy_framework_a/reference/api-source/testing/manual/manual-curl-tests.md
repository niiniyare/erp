> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

# Manual Curl Testing Commands - Step by Step

Follow these commands in order to manually test the User API with curl.

## Prerequisites

1. **Start the server** (in one terminal):
```bash
# Set environment variables
export DB_USER=admin
export DB_PASSWORD=admin
export DB_NAME=ledger
export DB_HOST=localhost
export DB_PORT=5432
export REDIS_HOST=localhost
export REDIS_PORT=6379

# Run the server
go run cmd/server/main.go
```

2. **Open another terminal** for testing

## Step 1: Health Check

```bash
curl -X GET http://localhost:8080/health | jq .
```

Expected response:
```json
{
  "status": "ok",
  "timestamp": "2024-01-01T10:00:00Z"
}
```

## Step 2: Create Entity (Required for Users)

```bash
curl -X POST http://localhost:8080/api/v1/entities \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Company",
    "code": "TEST001",
    "type": "account"
  }' | jq .
```

**⚠️ IMPORTANT: Copy the `id` from the response - you'll need it for user creation!**

Expected response:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Test Company",
  "code": "TEST001",
  "type": "account",
  "is_active": true,
  "created_at": "2024-01-01T10:00:00Z"
}
```

## Step 3: Create a User

**Replace `YOUR_ENTITY_ID` with the actual ID from Step 2:**

```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "entity_id": "YOUR_ENTITY_ID",
    "username": "john.doe",
    "email": "john.doe@example.com",
    "password": "SecurePassword123!",
    "user_type": "INTERNAL",
    "account_status": "ACTIVE",
    "mfa_enabled": false
  }' | jq .
```

**⚠️ IMPORTANT: Copy the `id` from the response - you'll need it for user operations!**

Expected response:
```json
{
  "id": "user-uuid-here",
  "entity_id": "YOUR_ENTITY_ID",
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

## Step 4: Authenticate the User

```bash
curl -X POST http://localhost:8080/api/v1/users/auth \
  -H "Content-Type: application/json" \
  -d '{
    "identifier": "john.doe@example.com",
    "password": "SecurePassword123!"
  }' | jq .
```

Expected response:
```json
{
  "user": {
    "id": "user-uuid-here",
    "username": "john.doe",
    "email": "john.doe@example.com",
    "user_type": "INTERNAL",
    "is_active": true
  },
  "token": "jwt-token-placeholder",
  "expires_at": "2024-01-01T18:00:00Z"
}
```

## Step 5: Get User by ID

**Replace `YOUR_USER_ID` with the actual user ID from Step 3:**

```bash
curl -X GET http://localhost:8080/api/v1/users/YOUR_USER_ID | jq .
```

Expected response:
```json
{
  "id": "YOUR_USER_ID",
  "entity_id": "YOUR_ENTITY_ID",
  "username": "john.doe",
  "email": "john.doe@example.com",
  "user_type": "INTERNAL",
  "account_status": "ACTIVE",
  "is_active": true,
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T10:00:00Z"
}
```

## Step 6: List All Users

```bash
curl -X GET "http://localhost:8080/api/v1/users?limit=10&offset=0" | jq .
```

Expected response:
```json
{
  "users": [
    {
      "id": "YOUR_USER_ID",
      "username": "john.doe",
      "email": "john.doe@example.com",
      "user_type": "INTERNAL",
      "is_active": true,
      "created_at": "2024-01-01T10:00:00Z"
    }
  ],
  "total": 1,
  "limit": 10,
  "offset": 0
}
```

## Step 7: Search Users

```bash
curl -X GET "http://localhost:8080/api/v1/users/search?q=john&limit=5" | jq .
```

Expected response:
```json
{
  "users": [
    {
      "id": "YOUR_USER_ID",
      "username": "john.doe",
      "email": "john.doe@example.com",
      "user_type": "INTERNAL",
      "is_active": true
    }
  ],
  "total": 1,
  "limit": 5,
  "offset": 0,
  "query": "john"
}
```

## Step 8: Update User

**Replace `YOUR_USER_ID` with the actual user ID:**

```bash
curl -X PUT http://localhost:8080/api/v1/users/YOUR_USER_ID \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john.doe.updated",
    "mfa_enabled": true,
    "session_timeout_minutes": 240
  }' | jq .
```

Expected response:
```json
{
  "id": "YOUR_USER_ID",
  "username": "john.doe.updated",
  "email": "john.doe@example.com",
  "user_type": "INTERNAL",
  "mfa_enabled": true,
  "session_timeout_minutes": 240,
  "updated_at": "2024-01-01T10:05:00Z"
}
```

## Step 9: Update Password

```bash
curl -X PUT http://localhost:8080/api/v1/users/YOUR_USER_ID/password \
  -H "Content-Type: application/json" \
  -d '{
    "current_password": "SecurePassword123!",
    "new_password": "NewSecurePassword456!"
  }' | jq .
```

Expected response:
```json
{
  "message": "Password updated successfully"
}
```

## Step 10: Get User Roles

```bash
curl -X GET http://localhost:8080/api/v1/users/YOUR_USER_ID/roles | jq .
```

Expected response:
```json
{
  "roles": [
    {
      "id": "role-id",
      "name": "user",
      "display_name": "Standard User",
      "description": "Standard user role",
      "role_type": "STANDARD",
      "assigned_at": "2024-01-01T10:00:00Z"
    }
  ]
}
```

## Step 11: Test Error Scenarios

### Invalid Authentication:
```bash
curl -X POST http://localhost:8080/api/v1/users/auth \
  -H "Content-Type: application/json" \
  -d '{
    "identifier": "john.doe@example.com",
    "password": "wrongpassword"
  }' | jq .
```

Expected response (401):
```json
{
  "error": "Invalid credentials"
}
```

### Non-existent User:
```bash
curl -X GET http://localhost:8080/api/v1/users/invalid-uuid | jq .
```

Expected response (400):
```json
{
  "error": "Invalid user ID"
}
```

### Duplicate Email:
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "entity_id": "YOUR_ENTITY_ID",
    "username": "another.user",
    "email": "john.doe@example.com",
    "password": "AnotherPassword123!",
    "user_type": "INTERNAL"
  }' | jq .
```

Expected response (409):
```json
{
  "error": "Email already exists"
}
```

## Step 12: Delete User (Soft Delete)

```bash
curl -X DELETE http://localhost:8080/api/v1/users/YOUR_USER_ID | jq .
```

Expected response:
```json
{
  "message": "User deleted successfully"
}
```

### Verify Deletion:
```bash
curl -X GET http://localhost:8080/api/v1/users/YOUR_USER_ID | jq .
```

Expected response (404):
```json
{
  "error": "User not found"
}
```

## Performance Testing

### Test Cache Performance:

**First request (cache miss):**
```bash
time curl -X GET http://localhost:8080/api/v1/users/YOUR_USER_ID
```

**Second request (cache hit - should be faster):**
```bash
time curl -X GET http://localhost:8080/api/v1/users/YOUR_USER_ID
```

### Test with Multiple Concurrent Requests:
```bash
# Run 5 requests in parallel
for i in {1..5}; do
  curl -X GET "http://localhost:8080/api/v1/users?limit=10" &
done
wait
```

## Complete Test Flow Script

Here's a complete script you can save and run:

```bash
#!/bin/bash

# Save this as test-flow.sh and run: chmod +x test-flow.sh && ./test-flow.sh

BASE_URL="http://localhost:8080"

echo " Testing User API Flow..."

# Health check
echo "1. Health check..."
curl -s $BASE_URL/health | jq .

# Create entity
echo "2. Creating entity..."
ENTITY_RESPONSE=$(curl -s -X POST $BASE_URL/api/v1/entities \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Company","code":"TEST001","type":"account"}')

ENTITY_ID=$(echo $ENTITY_RESPONSE | jq -r '.id')
echo "Entity ID: $ENTITY_ID"

# Create user
echo "3. Creating user..."
USER_RESPONSE=$(curl -s -X POST $BASE_URL/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "entity_id": "'$ENTITY_ID'",
    "username": "test.user",
    "email": "test@example.com",
    "password": "TestPassword123!",
    "user_type": "INTERNAL"
  }')

USER_ID=$(echo $USER_RESPONSE | jq -r '.id')
echo "User ID: $USER_ID"

# Authenticate
echo "4. Authenticating user..."
curl -s -X POST $BASE_URL/api/v1/users/auth \
  -H "Content-Type: application/json" \
  -d '{"identifier":"test@example.com","password":"TestPassword123!"}' | jq .

# Get user
echo "5. Getting user..."
curl -s -X GET $BASE_URL/api/v1/users/$USER_ID | jq .

# List users
echo "6. Listing users..."
curl -s -X GET "$BASE_URL/api/v1/users?limit=5" | jq .

echo "✅ Test flow completed!"
```

## Troubleshooting

1. **Server not responding**: Make sure server is running on port 8080
2. **Database errors**: Check database connection and ensure migrations are run
3. **Invalid UUID errors**: Make sure to use actual UUIDs from responses
4. **Cache errors**: Redis might not be running (optional for basic functionality)

## Monitoring

While testing, you can monitor:
- **Server logs**: Check console output for request logs
- **Database**: Monitor database connections and queries
- **Performance**: Use `time` command to measure response times
- **Traces**: If tracing is configured, check trace outputs

The API includes  observability, so all requests are logged, traced, and have metrics collected automatically.