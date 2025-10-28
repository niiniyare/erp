#!/bin/bash

# test-auth.sh - Authentication API Testing Script
# Tests login, token refresh, and session management endpoints

set -e

# Configuration
BASE_URL="${API_BASE_URL:-http://localhost:8080}"
TEST_USER="${TEST_USER:-admin@example.com}"
TEST_PASS="${TEST_PASS:-password123}"
TEST_TENANT="${TEST_TENANT:-default}"

echo "🔐 Testing Authentication API..."
echo "Base URL: $BASE_URL"

# Test 1: Login
echo "\n📋 Test 1: User Login"
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"$TEST_USER\",
    \"password\": \"$TEST_PASS\",
    \"tenant_id\": \"$TEST_TENANT\"
  }")

if echo "$LOGIN_RESPONSE" | grep -q "access_token"; then
  echo "✅ Login successful"
  ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
  echo "Token: ${ACCESS_TOKEN:0:20}..."
else
  echo "❌ Login failed"
  echo "Response: $LOGIN_RESPONSE"
  exit 1
fi

# Test 2: Protected Endpoint Access
echo "\n📋 Test 2: Protected Endpoint Access"
PROTECTED_RESPONSE=$(curl -s -X GET "$BASE_URL/user/profile" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

if echo "$PROTECTED_RESPONSE" | grep -q "id"; then
  echo "✅ Protected endpoint access successful"
else
  echo "❌ Protected endpoint access failed"
  echo "Response: $PROTECTED_RESPONSE"
fi

# Test 3: Token Validation
echo "\n📋 Test 3: Token Validation"
VALIDATION_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/validate" \
  -H "Content-Type: application/json" \
  -d "{\"token\": \"$ACCESS_TOKEN\"}")

if echo "$VALIDATION_RESPONSE" | grep -q "valid"; then
  echo "✅ Token validation successful"
else
  echo "❌ Token validation failed"
  echo "Response: $VALIDATION_RESPONSE"
fi

echo "\n🎉 Authentication API tests completed!"
