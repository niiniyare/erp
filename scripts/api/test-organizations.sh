#!/bin/bash

# test-organizations.sh - Organization API Testing Script
# Tests CRUD operations for organization management

set -e

# Configuration
BASE_URL="${API_BASE_URL:-http://localhost:8080}"
AUTH_TOKEN="${AUTH_TOKEN:-}"

if [ -z "$AUTH_TOKEN" ]; then
  echo "Error: AUTH_TOKEN environment variable required"
  echo "Usage: AUTH_TOKEN='your-jwt-token' ./test-organizations.sh"
  exit 1
fi

echo " Testing Organization API..."
echo "Base URL: $BASE_URL"

# Test 1: List Organizations
echo "\n Test 1: List Organizations"
LIST_RESPONSE=$(curl -s -X GET "$BASE_URL/organizations" \
  -H "Authorization: Bearer $AUTH_TOKEN")

if echo "$LIST_RESPONSE" | grep -q "organizations"; then
  echo "✅ List organizations successful"
  ORG_COUNT=$(echo "$LIST_RESPONSE" | grep -o '"id"' | wc -l)
  echo "Found $ORG_COUNT organizations"
else
  echo "❌ List organizations failed"
  echo "Response: $LIST_RESPONSE"
fi

# Test 2: Create Organization
echo "\n Test 2: Create Organization"
CREATE_PAYLOAD='{
  "name": "Test Organization",
  "description": "Created by test script",
  "type": "business",
  "settings": {
    "timezone": "UTC",
    "currency": "USD"
  }
}'

CREATE_RESPONSE=$(curl -s -X POST "$BASE_URL/organizations" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "Content-Type: application/json" \
  -d "$CREATE_PAYLOAD")

if echo "$CREATE_RESPONSE" | grep -q "id"; then
  echo "✅ Create organization successful"
  ORG_ID=$(echo "$CREATE_RESPONSE" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
  echo "Created organization ID: $ORG_ID"
else
  echo "❌ Create organization failed"
  echo "Response: $CREATE_RESPONSE"
  exit 1
fi

# Test 3: Get Organization Details
echo "\n Test 3: Get Organization Details"
GET_RESPONSE=$(curl -s -X GET "$BASE_URL/organizations/$ORG_ID" \
  -H "Authorization: Bearer $AUTH_TOKEN")

if echo "$GET_RESPONSE" | grep -q "Test Organization"; then
  echo "✅ Get organization details successful"
else
  echo "❌ Get organization details failed"
  echo "Response: $GET_RESPONSE"
fi

# Test 4: Update Organization
echo "\n Test 4: Update Organization"
UPDATE_PAYLOAD='{
  "name": "Updated Test Organization",
  "description": "Updated by test script"
}'

UPDATE_RESPONSE=$(curl -s -X PUT "$BASE_URL/organizations/$ORG_ID" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "Content-Type: application/json" \
  -d "$UPDATE_PAYLOAD")

if echo "$UPDATE_RESPONSE" | grep -q "Updated Test Organization"; then
  echo "✅ Update organization successful"
else
  echo "❌ Update organization failed"
  echo "Response: $UPDATE_RESPONSE"
fi

# Test 5: Delete Organization
echo "\n Test 5: Delete Organization"
DELETE_RESPONSE=$(curl -s -X DELETE "$BASE_URL/organizations/$ORG_ID" \
  -H "Authorization: Bearer $AUTH_TOKEN")

if echo "$DELETE_RESPONSE" | grep -q "deleted\|success" || [ -z "$DELETE_RESPONSE" ]; then
  echo "✅ Delete organization successful"
else
  echo "❌ Delete organization failed"
  echo "Response: $DELETE_RESPONSE"
fi

echo "\n Organization API tests completed!"
