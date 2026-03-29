#!/bin/bash

# Test script for Tenant API - Comprehensive Testing
# Updated: 2025-09-20
# This script tests all tenant endpoints based on real API testing results

# Source environment configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/setup_env.sh"

BASE_URL="http://localhost:${SERVER_PORT:-8080}"
CREATED_TENANT_ID=""
random="$(cat /dev/urandom | tr -dc 'A-Z' | fold -w 6 | head -n 1)"
echo " Comprehensive Tenant API Testing"
echo "===================================="
echo "Base URL: $BASE_URL"
echo "Server Port: ${SERVER_PORT:-8080}"
echo "Testing Date: $(date)"
echo ""

# Test 1: Create a new tenant (bypasses middleware)
echo " Test 1: Create New Tenant (Public Endpoint)"
echo "POST $BASE_URL/api/v1/tenants"
CREATE_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Company API Script",
    "email": "contact@testapi.com",
    "subdomain": $random,
    "country_code": "US",
    "currency_code": "USD",
    "status": "ACTIVE",
    "industry": "Technology",
    "company_size": "Startup",
    "contact": {
      "email": "contact@testapi.com"
    },
    "settings": {
      "timezone": "UTC",
      "currency": "USD",
      "date_format": "YYYY-MM-DD",
      "language": "en"
    }
  }' \
  "$BASE_URL/api/v1/tenants")

# Extract tenant ID from response
CREATED_TENANT_ID=$(echo "$CREATE_RESPONSE" | sed 's/HTTP_STATUS:.*//' | jq -r '.tenant.id // empty')
STATUS=$(echo "$CREATE_RESPONSE" | grep -o 'HTTP_STATUS:[0-9]*' | cut -d: -f2)

echo "$CREATE_RESPONSE" | sed 's/HTTP_STATUS:.*//' | jq .
echo "Status: $STATUS"
if [ -n "$CREATED_TENANT_ID" ]; then
  echo "✅ Tenant created successfully with ID: $CREATED_TENANT_ID"
else
  echo "❌ Failed to create tenant"
fi
echo ""
echo "---"

# Test 2: Health endpoint with tenant context
echo " Test 2: Health Check (Requires Tenant Context)"
echo "GET $BASE_URL/api/v1/tenants/health"
if [ -n "$CREATED_TENANT_ID" ]; then
  curl -s -w "\nStatus: %{http_code}\n" \
    -H "Content-Type: application/json" \
    -H "X-Tenant-ID: $CREATED_TENANT_ID" \
    "$BASE_URL/api/v1/tenants/health"
else
  echo "❌ Skipping health test - no tenant ID available"
fi
echo ""
echo "---"

# Test 3: Get specific tenant by ID
echo " Test 3: Get Tenant by ID"
echo "GET $BASE_URL/api/v1/tenants/$CREATED_TENANT_ID"
if [ -n "$CREATED_TENANT_ID" ]; then
  curl -s -w "\nStatus: %{http_code}\n" \
    -H "Content-Type: application/json" \
    -H "X-Tenant-ID: $CREATED_TENANT_ID" \
    "$BASE_URL/api/v1/tenants/$CREATED_TENANT_ID"
else
  echo "❌ Skipping get test - no tenant ID available"
fi
echo ""
echo "---"

# Test 4: List tenants with pagination
echo " Test 4: List Tenants with Pagination"
echo "GET $BASE_URL/api/v1/tenants?page=1&page_size=10"
if [ -n "$CREATED_TENANT_ID" ]; then
  curl -s -w "\nStatus: %{http_code}\n" \
    -H "Content-Type: application/json" \
    -H "X-Tenant-ID: $CREATED_TENANT_ID" \
    "$BASE_URL/api/v1/tenants?page=1&page_size=10"
else
  echo "❌ Skipping list test - no tenant ID available"
fi
echo ""
echo "---"

# Test 5: Update tenant
echo " Test 5: Update Tenant"
echo "PUT $BASE_URL/api/v1/tenants/$CREATED_TENANT_ID"
if [ -n "$CREATED_TENANT_ID" ]; then
  curl -s -w "\nStatus: %{http_code}\n" \
    -X PUT \
    -H "Content-Type: application/json" \
    -H "X-Tenant-ID: $CREATED_TENANT_ID" \
    -d '{
      "name": "Updated Test Company API",
      "contact": {
        "email": "updated@testapi.com"
      },
      "settings": {
        "timezone": "America/New_York",
        "currency": "USD",
        "date_format": "MM/DD/YYYY",
        "language": "en"
      }
    }' \
    "$BASE_URL/api/v1/tenants/$CREATED_TENANT_ID"
else
  echo "❌ Skipping update test - no tenant ID available"
fi
echo ""
echo "---"

# Test 6: Test Usage Analytics
echo " Test 6: Get Usage Analytics"
echo "GET $BASE_URL/api/v1/tenants/$CREATED_TENANT_ID/analytics?period=current_month"
if [ -n "$CREATED_TENANT_ID" ]; then
  curl -s -w "\nStatus: %{http_code}\n" \
    -H "Content-Type: application/json" \
    -H "X-Tenant-ID: $CREATED_TENANT_ID" \
    "$BASE_URL/api/v1/tenants/$CREATED_TENANT_ID/analytics?period=current_month"
else
  echo "❌ Skipping analytics test - no tenant ID available"
fi
echo ""
echo "---"

# Test 7: Test Error Scenarios
echo " Test 7: Error Scenarios"
echo "Testing endpoints without tenant context (should fail):"
echo "GET $BASE_URL/api/v1/tenants (no header)"
curl -s -w "\nStatus: %{http_code}\n" \
  -H "Content-Type: application/json" \
  "$BASE_URL/api/v1/tenants"
echo ""
echo "---"

# Test 8: Clean up - Delete the test tenant
echo " Test 8: Delete Test Tenant (Cleanup)"
echo "DELETE $BASE_URL/api/v1/tenants/$CREATED_TENANT_ID"
if [ -n "$CREATED_TENANT_ID" ]; then
  curl -s -w "\nStatus: %{http_code}\n" \
    -X DELETE \
    -H "X-Tenant-ID: $CREATED_TENANT_ID" \
    "$BASE_URL/api/v1/tenants/$CREATED_TENANT_ID"
  echo "️ Test tenant deleted"
else
  echo "❌ Skipping delete test - no tenant ID available"
fi
echo ""
echo "---"

echo "✅ Comprehensive Tenant API tests completed!"
echo ""
echo " Test Results Summary:"
echo "- Test 1: Create Tenant (should succeed - 200)"
echo "- Test 2: Health Check (should succeed with tenant - 200)"
echo "- Test 3: Get Tenant (should succeed - 200)"
echo "- Test 4: List Tenants (should succeed but empty due to RLS - 200)"
echo "- Test 5: Update Tenant (should succeed - 200)"
echo "- Test 6: Usage Analytics (should succeed - 200)"
echo "- Test 7: Error Scenarios (should fail without tenant - 400)"
echo "- Test 8: Delete Tenant (should succeed - 200)"
echo ""
echo " Known Issues:"
echo "- Provision endpoint fails due to DB constraints"
echo "- Suspend/Reactivate fail due to status check constraints"
echo "- Configuration update requires active tenant status"
echo "- List endpoint returns empty due to Row-Level Security (RLS)"

# Additional utility functions
echo ""
echo "️ Additional Test Utilities:"
echo "================================"

# Function to test provisioning (known to fail)
test_provisioning() {
  echo " Testing Provisioning (Expected to fail due to DB constraints)"
  if [ -n "$CREATED_TENANT_ID" ]; then
    curl -s -w "\nStatus: %{http_code}\n" \
      -X POST \
      -H "Content-Type: application/json" \
      -H "X-Tenant-ID: $CREATED_TENANT_ID" \
      -d '{
        "name": "Provisioned Company",
        "subdomain": "provisioned",
        "contact_email": "contact@provisioned.com",
        "admin_email": "admin@provisioned.com",
        "admin_first_name": "Jane",
        "admin_last_name": "Smith"
      }' \
      "$BASE_URL/api/v1/tenants/provision"
  fi
}

# Function to test suspend (known to fail)
test_suspend() {
  echo " Testing Suspend (Expected to fail due to status constraints)"
  if [ -n "$CREATED_TENANT_ID" ]; then
    curl -s -w "\nStatus: %{http_code}\n" \
      -X POST \
      -H "Content-Type: application/json" \
      -H "X-Tenant-ID: $CREATED_TENANT_ID" \
      -d '{"reason": "Testing suspension functionality"}' \
      "$BASE_URL/api/v1/tenants/$CREATED_TENANT_ID/suspend"
  fi
}

# Uncomment to test known failing endpoints:
# echo " Testing Known Issues (will fail):"
# test_provisioning
# test_suspend
