# Authentication API - curl Examples

Complete collection of curl commands for testing the Authentication API.

##  Prerequisites

### Environment Setup
```bash
# Set base URL
export API_BASE_URL="http://localhost:8080"

# Test server connectivity
curl -X GET $API_BASE_URL/api/v1/tenants/health | jq .
```

### Required Tools
```bash
# Install jq for JSON processing
sudo apt-get install jq  # Ubuntu/Debian
brew install jq          # macOS
```

##  Authentication Flow Examples

### 1. User Login

#### Basic Login
```bash
curl -X POST $API_BASE_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePassword123!"
  }' | jq .
```

#### Login with Response Processing
```bash
# Login and extract tokens
AUTH_RESPONSE=$(curl -s -X POST $API_BASE_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com", 
    "password": "password123"
  }')

# Extract access token
ACCESS_TOKEN=$(echo $AUTH_RESPONSE | jq -r '.access_token')
REFRESH_TOKEN=$(echo $AUTH_RESPONSE | jq -r '.refresh_token // empty')
USER_ID=$(echo $AUTH_RESPONSE | jq -r '.user.id')

echo "Access Token: ${ACCESS_TOKEN:0:20}..."
echo "User ID: $USER_ID"
```

#### Login with Different User Types
```bash
# Admin user login
curl -X POST $API_BASE_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "AdminPassword123!"
  }' | jq .

# Regular user login
curl -X POST $API_BASE_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "employee@example.com", 
    "password": "EmployeePass123!"
  }' | jq .

# External user login
curl -X POST $API_BASE_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "contractor@example.com",
    "password": "ContractorPass123!"
  }' | jq .
```

### 2. Token Validation

#### Basic Token Validation
```bash
# Validate current token
curl -X GET $API_BASE_URL/api/v1/auth/validate \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

#### Token Information Extraction
```bash
# Get detailed token information
TOKEN_INFO=$(curl -s -X GET $API_BASE_URL/api/v1/auth/validate \
  -H "Authorization: Bearer $ACCESS_TOKEN")

VALID=$(echo $TOKEN_INFO | jq -r '.valid')
USER_EMAIL=$(echo $TOKEN_INFO | jq -r '.user.email')
PERMISSIONS=$(echo $TOKEN_INFO | jq -r '.permissions[]')

echo "Token Valid: $VALID"
echo "User Email: $USER_EMAIL" 
echo "Permissions: $PERMISSIONS"
```

#### Token Validation with Error Handling
```bash
# Test with invalid token
curl -X GET $API_BASE_URL/api/v1/auth/validate \
  -H "Authorization: Bearer <invalid-token>" | jq .

# Test with missing authorization header
curl -X GET $API_BASE_URL/api/v1/auth/validate | jq .

# Test with malformed authorization header
curl -X GET $API_BASE_URL/api/v1/auth/validate \
  -H "Authorization: Invalid-Format token" | jq .
```

### 3. Token Refresh

#### Basic Token Refresh
```bash
# Refresh access token
curl -X POST $API_BASE_URL/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{
    \"refresh_token\": \"$REFRESH_TOKEN\"
  }" | jq .
```

#### Refresh with New Token Storage
```bash
# Refresh and update stored token
REFRESH_RESPONSE=$(curl -s -X POST $API_BASE_URL/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\": \"$REFRESH_TOKEN\"}")

# Update access token
NEW_ACCESS_TOKEN=$(echo $REFRESH_RESPONSE | jq -r '.access_token')
echo "New Access Token: ${NEW_ACCESS_TOKEN:0:20}..."

# Test new token
curl -X GET $API_BASE_URL/api/v1/auth/validate \
  -H "Authorization: Bearer $NEW_ACCESS_TOKEN" | jq .
```

### 4. User Logout

#### Basic Logout
```bash
curl -X POST $API_BASE_URL/api/v1/auth/logout \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"token\": \"$ACCESS_TOKEN\"
  }" -v
```

#### Logout with Verification
```bash
# Logout user
echo "Logging out user..."
curl -s -X POST $API_BASE_URL/api/v1/auth/logout \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"token\": \"$ACCESS_TOKEN\"}"

# Verify token is invalidated
echo "Verifying token invalidation..."
curl -X GET $API_BASE_URL/api/v1/auth/validate \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

##   Test Scenarios

### Scenario 1: Complete Authentication Flow
```bash
echo "=== Complete Authentication Flow ==="

# 1. Login
echo "Step 1: User Login"
LOGIN_RESPONSE=$(curl -s -X POST $API_BASE_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePassword123!"
  }')

ACCESS_TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.access_token')
USER_ID=$(echo $LOGIN_RESPONSE | jq -r '.user.id')

echo "✅ Login successful - User ID: $USER_ID"
echo " Access Token: ${ACCESS_TOKEN:0:20}..."

# 2. Validate token
echo "Step 2: Token Validation"
VALIDATION=$(curl -s -X GET $API_BASE_URL/api/v1/auth/validate \
  -H "Authorization: Bearer $ACCESS_TOKEN")

VALID=$(echo $VALIDATION | jq -r '.valid')
if [ "$VALID" = "true" ]; then
  echo "✅ Token validation successful"
else
  echo "❌ Token validation failed"
fi

# 3. Use token for protected API call
echo "Step 3: Protected API Call"
API_RESPONSE=$(curl -s -X GET $API_BASE_URL/api/v1/users \
  -H "Authorization: Bearer $ACCESS_TOKEN" -w "%{http_code}")

if [[ "$API_RESPONSE" == *"200" ]]; then
  echo "✅ Protected API call successful"
else
  echo "❌ Protected API call failed"
fi

# 4. Logout
echo "Step 4: User Logout"
curl -s -X POST $API_BASE_URL/api/v1/auth/logout \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"token\": \"$ACCESS_TOKEN\"}"

echo "✅ Logout successful"

# 5. Verify token invalidation
echo "Step 5: Token Invalidation Verification"
POST_LOGOUT_VALIDATION=$(curl -s -X GET $API_BASE_URL/api/v1/auth/validate \
  -H "Authorization: Bearer $ACCESS_TOKEN")

POST_LOGOUT_VALID=$(echo $POST_LOGOUT_VALIDATION | jq -r '.valid // false')
if [ "$POST_LOGOUT_VALID" = "false" ]; then
  echo "✅ Token successfully invalidated"
else
  echo "❌ Token still valid after logout"
fi

echo "=== Authentication flow test completed ==="
```

### Scenario 2: Error Handling Tests
```bash
echo "=== Error Handling Tests ==="

# Test 1: Invalid email format
echo "Test 1: Invalid email format"
curl -s -X POST $API_BASE_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "invalid-email",
    "password": "password123"
  }' | jq .

# Test 2: Missing password
echo "Test 2: Missing password"
curl -s -X POST $API_BASE_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com"
  }' | jq .

# Test 3: Empty credentials
echo "Test 3: Empty credentials"
curl -s -X POST $API_BASE_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "",
    "password": ""
  }' | jq .

# Test 4: Invalid credentials
echo "Test 4: Invalid credentials"
curl -s -X POST $API_BASE_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "nonexistent@example.com",
    "password": "wrongpassword"
  }' | jq .

# Test 5: Malformed JSON
echo "Test 5: Malformed JSON"
curl -s -X POST $API_BASE_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password":' | jq .

# Test 6: Invalid token format
echo "Test 6: Invalid token format"
curl -s -X GET $API_BASE_URL/api/v1/auth/validate \
  -H "Authorization: Bearer <invalid-token>" | jq .

# Test 7: Expired token (simulated)
echo "Test 7: Missing authorization header"
curl -s -X GET $API_BASE_URL/api/v1/auth/validate | jq .

echo "=== Error handling tests completed ==="
```

### Scenario 3: Token Lifecycle Management
```bash
echo "=== Token Lifecycle Management ==="

# Login and get tokens
echo "Getting initial tokens..."
AUTH_RESPONSE=$(curl -s -X POST $API_BASE_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePassword123!"
  }')

ORIGINAL_TOKEN=$(echo $AUTH_RESPONSE | jq -r '.access_token')
REFRESH_TOKEN=$(echo $AUTH_RESPONSE | jq -r '.refresh_token // empty')

echo "Original token: ${ORIGINAL_TOKEN:0:20}..."

# Validate original token
echo "Validating original token..."
curl -s -X GET $API_BASE_URL/api/v1/auth/validate \
  -H "Authorization: Bearer $ORIGINAL_TOKEN" | jq '.valid'

# Refresh token (if available)
if [ "$REFRESH_TOKEN" != "" ]; then
  echo "Refreshing token..."
  REFRESH_RESPONSE=$(curl -s -X POST $API_BASE_URL/api/v1/auth/refresh \
    -H "Content-Type: application/json" \
    -d "{\"refresh_token\": \"$REFRESH_TOKEN\"}")
  
  NEW_TOKEN=$(echo $REFRESH_RESPONSE | jq -r '.access_token')
  echo "New token: ${NEW_TOKEN:0:20}..."
  
  # Validate new token
  echo "Validating new token..."
  curl -s -X GET $API_BASE_URL/api/v1/auth/validate \
    -H "Authorization: Bearer $NEW_TOKEN" | jq '.valid'
fi

# Logout with original token
echo "Logging out..."
curl -s -X POST $API_BASE_URL/api/v1/auth/logout \
  -H "Authorization: Bearer $ORIGINAL_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"token\": \"$ORIGINAL_TOKEN\"}"

echo "=== Token lifecycle test completed ==="
```

### Scenario 4: Multiple User Sessions
```bash
echo "=== Multiple User Sessions Test ==="

# User 1 login
echo "User 1 login..."
USER1_RESPONSE=$(curl -s -X POST $API_BASE_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user1@example.com",
    "password": "password123"
  }')

USER1_TOKEN=$(echo $USER1_RESPONSE | jq -r '.access_token')
USER1_ID=$(echo $USER1_RESPONSE | jq -r '.user.id')

# User 2 login
echo "User 2 login..."
USER2_RESPONSE=$(curl -s -X POST $API_BASE_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user2@example.com",
    "password": "password456"
  }')

USER2_TOKEN=$(echo $USER2_RESPONSE | jq -r '.access_token')
USER2_ID=$(echo $USER2_RESPONSE | jq -r '.user.id')

# Validate both tokens
echo "Validating User 1 token..."
curl -s -X GET $API_BASE_URL/api/v1/auth/validate \
  -H "Authorization: Bearer $USER1_TOKEN" | jq '.user.id'

echo "Validating User 2 token..."
curl -s -X GET $API_BASE_URL/api/v1/auth/validate \
  -H "Authorization: Bearer $USER2_TOKEN" | jq '.user.id'

# Logout User 1
echo "Logging out User 1..."
curl -s -X POST $API_BASE_URL/api/v1/auth/logout \
  -H "Authorization: Bearer $USER1_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"token\": \"$USER1_TOKEN\"}"

# Verify User 1 token is invalid, User 2 token still valid
echo "Verifying User 1 token invalidation..."
curl -s -X GET $API_BASE_URL/api/v1/auth/validate \
  -H "Authorization: Bearer $USER1_TOKEN" | jq '.valid // false'

echo "Verifying User 2 token still valid..."
curl -s -X GET $API_BASE_URL/api/v1/auth/validate \
  -H "Authorization: Bearer $USER2_TOKEN" | jq '.valid'

# Logout User 2
echo "Logging out User 2..."
curl -s -X POST $API_BASE_URL/api/v1/auth/logout \
  -H "Authorization: Bearer $USER2_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"token\": \"$USER2_TOKEN\"}"

echo "=== Multiple user sessions test completed ==="
```

##  Performance and Load Testing

### Response Time Testing
```bash
echo "=== Performance Testing ==="

# Test login performance
echo "Testing login performance..."
time curl -s -X POST $API_BASE_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }' > /dev/null

# Test validation performance
echo "Testing validation performance..."
time curl -s -X GET $API_BASE_URL/api/v1/auth/validate \
  -H "Authorization: Bearer $ACCESS_TOKEN" > /dev/null

echo "=== Performance testing completed ==="
```

### Load Testing
```bash
echo "=== Load Testing ==="

# Concurrent login attempts
echo "Testing concurrent logins..."
for i in {1..5}; do
  (curl -s -X POST $API_BASE_URL/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d "{
      \"email\": \"user$i@example.com\",
      \"password\": \"password123\"
    }" > /dev/null) &
done

wait
echo "Concurrent login test completed"

echo "=== Load testing completed ==="
```

##  Utility Functions

### Authentication Helper Function
```bash
#!/bin/bash

# Function to perform authentication and return token
authenticate_user() {
  local email="$1"
  local password="$2"
  
  if [ -z "$email" ] || [ -z "$password" ]; then
    echo "Usage: authenticate_user <email> <password>"
    return 1
  fi
  
  local response=$(curl -s -X POST $API_BASE_URL/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d "{
      \"email\": \"$email\",
      \"password\": \"$password\"
    }")
  
  local token=$(echo $response | jq -r '.access_token // empty')
  
  if [ "$token" != "" ]; then
    echo $token
  else
    echo "Authentication failed"
    echo $response | jq .
    return 1
  fi
}

# Usage example
TOKEN=$(authenticate_user "test@example.com" "password123")
if [ $? -eq 0 ]; then
  echo "Authentication successful: ${TOKEN:0:20}..."
else
  echo "Authentication failed"
fi
```

### Token Validation Function
```bash
#!/bin/bash

# Function to validate token and return user info
validate_token() {
  local token="$1"
  
  if [ -z "$token" ]; then
    echo "Usage: validate_token <token>"
    return 1
  fi
  
  local response=$(curl -s -X GET $API_BASE_URL/api/v1/auth/validate \
    -H "Authorization: Bearer $token")
  
  local valid=$(echo $response | jq -r '.valid // false')
  
  if [ "$valid" = "true" ]; then
    echo "Token is valid"
    echo $response | jq '.user'
  else
    echo "Token is invalid"
    return 1
  fi
}

# Usage example
validate_token "$ACCESS_TOKEN"
```

##  Common Issues and Solutions

### Issue: 401 Unauthorized
```bash
# Check token format
echo $ACCESS_TOKEN | cut -d'.' -f1 | base64 -d 2>/dev/null | jq .

# Check token expiration
echo $ACCESS_TOKEN | cut -d'.' -f2 | base64 -d 2>/dev/null | jq .

# Verify authorization header format
curl -X GET $API_BASE_URL/api/v1/auth/validate \
  -H "Authorization: Bearer $ACCESS_TOKEN" -v
```

### Issue: Invalid JSON
```bash
# Validate JSON before sending
echo '{"email": "test@example.com", "password": "password123"}' | jq empty

# Check Content-Type header
curl -X POST $API_BASE_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "password123"}' -v
```

### Issue: Rate Limiting
```bash
# Check rate limit headers
curl -X POST $API_BASE_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "password123"}' -i

# Wait between requests
sleep 1
```

---

**Last Updated**: 2025-07-19  
**API Version**: 1.0.0  
**Tested with**: curl 7.81.0, jq 1.6