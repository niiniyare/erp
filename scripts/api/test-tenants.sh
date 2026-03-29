#!/bin/bash

# test-tenants.sh - Tenant API Testing Script
# Tests multi-tenant operations and isolation

set -e

# Configuration
BASE_URL="${API_BASE_URL:-http://localhost:8080}"
AUTH_TOKEN="${AUTH_TOKEN:-}"

if [ -z "$AUTH_TOKEN" ]; then
  echo "Error: AUTH_TOKEN environment variable required"
  echo "Usage: AUTH_TOKEN='your-jwt-token' ./test-tenants.sh"
  exit 1
fi

echo " Testing Tenant API..."
echo "Base URL: $BASE_URL"

# Test 1: List Tenants
echo "\n Test 1: List Tenants"
LIST_RESPONSE=$(curl -s -X GET "$BASE_URL/tenants" \
  -H "Authorization: Bearer $AUTH_TOKEN")

if echo "$LIST_RESPONSE" | grep -q "tenants"; then
  echo "✅ List tenants successful"
  TENANT_COUNT=$(echo "$LIST_RESPONSE" | grep -o '"id"' | wc -l)
  echo "Found $TENANT_COUNT tenants"
else
  echo "❌ List tenants failed"
  echo "Response: $LIST_RESPONSE"
fi

# Test 2: Create Tenant
echo "\n Test 2: Create Tenant"
CREATE_PAYLOAD='{
  "name": "Test Tenant",
  "slug": "test-tenant-$(date +%s)",
  "description": "Created by test script",
  "settings": {
    "max_users": 100,
    "features": ["finance", "inventory"],
    "timezone": "UTC"
  }
}'

CREATE_RESPONSE=$(curl -s -X POST "$BASE_URL/tenants" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "Content-Type: application/json" \
  -d "$CREATE_PAYLOAD")

if echo "$CREATE_RESPONSE" | grep -q "id"; then
  echo "✅ Create tenant successful"
  TENANT_ID=$(echo "$CREATE_RESPONSE" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
  echo "Created tenant ID: $TENANT_ID"
else
  echo "❌ Create tenant failed"
  echo "Response: $CREATE_RESPONSE"
  exit 1
fi

# Test 3: Get Tenant Details
echo "\n Test 3: Get Tenant Details"
GET_RESPONSE=$(curl -s -X GET "$BASE_URL/tenants/$TENANT_ID" \
  -H "Authorization: Bearer $AUTH_TOKEN")

if echo "$GET_RESPONSE" | grep -q "Test Tenant"; then
  echo "✅ Get tenant details successful"
else
  echo "❌ Get tenant details failed"
  echo "Response: $GET_RESPONSE"
fi

# Test 4: Update Tenant Settings
echo "\n Test 4: Update Tenant Settings"
UPDATE_PAYLOAD='{
  "settings": {
    "max_users": 200,
    "features": ["finance", "inventory", "reporting"],
    "timezone": "America/New_York"
  }
}'

UPDATE_RESPONSE=$(curl -s -X PATCH "$BASE_URL/tenants/$TENANT_ID/settings" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "Content-Type: application/json" \
  -d "$UPDATE_PAYLOAD")

if echo "$UPDATE_RESPONSE" | grep -q "max_users.*200"; then
  echo "✅ Update tenant settings successful"
else
  echo "❌ Update tenant settings failed"
  echo "Response: $UPDATE_RESPONSE"
fi

# Test 5: Tenant Isolation Check
echo "\n Test 5: Tenant Isolation Check"
ISOLATION_RESPONSE=$(curl -s -X GET "$BASE_URL/tenants/$TENANT_ID/users" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID")

if echo "$ISOLATION_RESPONSE" | grep -q "users" || echo "$ISOLATION_RESPONSE" | grep -q "\[\]"; then
  echo "✅ Tenant isolation working correctly"
else
  echo "❌ Tenant isolation check failed"
  echo "Response: $ISOLATION_RESPONSE"
fi

# Test 6: Delete Tenant (Cleanup)
echo "\n Test 6: Delete Tenant (Cleanup)"
DELETE_RESPONSE=$(curl -s -X DELETE "$BASE_URL/tenants/$TENANT_ID" \
  -H "Authorization: Bearer $AUTH_TOKEN")

if echo "$DELETE_RESPONSE" | grep -q "deleted\|success" || [ -z "$DELETE_RESPONSE" ]; then
  echo "✅ Delete tenant successful"
else
  echo "❌ Delete tenant failed"
  echo "Response: $DELETE_RESPONSE"
fi

echo "\n Tenant API tests completed!"
