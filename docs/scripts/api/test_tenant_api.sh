#!/bin/bash

# Test script for Tenant API with native middleware
# This script tests the tenant endpoints to validate our Day 4 native middleware implementation

# Source environment configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/setup_env.sh"

BASE_URL="http://localhost:${SERVER_PORT}"
TENANT_ID="550e8400-e29b-41d4-a716-446655440000"

echo "🧪 Testing Tenant API with Native Middleware"
echo "============================================="
echo "Base URL: $BASE_URL"
echo ""

# Test 1: Health endpoint (should work without tenant)
echo "📋 Test 1: Health endpoint (public)"
echo "GET $BASE_URL/api/v1/tenants/health"
curl -s -w "\nStatus: %{http_code}\n" \
  -H "Content-Type: application/json" \
  "$BASE_URL/api/v1/tenants/health"
echo ""
echo "---"

# Test 2: List tenants without tenant header (should fail)
echo "📋 Test 2: List tenants without tenant header (should fail)"
echo "GET $BASE_URL/api/v1/tenants"
curl -s -w "\nStatus: %{http_code}\n" \
  -H "Content-Type: application/json" \
  "$BASE_URL/api/v1/tenants"
echo ""
echo "---"

# Test 3: List tenants with tenant header
echo "📋 Test 3: List tenants with tenant header"
echo "GET $BASE_URL/api/v1/tenants"
curl -s -w "\nStatus: %{http_code}\n" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  "$BASE_URL/api/v1/tenants"
echo ""
echo "---"

# Test 4: Get specific tenant with header
echo "📋 Test 4: Get specific tenant with header"
echo "GET $BASE_URL/api/v1/tenants/$TENANT_ID"
curl -s -w "\nStatus: %{http_code}\n" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  "$BASE_URL/api/v1/tenants/$TENANT_ID"
echo ""
echo "---"

# Test 5: Create tenant with header
echo "📋 Test 5: Create tenant with header"
echo "POST $BASE_URL/api/v1/tenants"
curl -s -w "\nStatus: %{http_code}\n" \
  -X POST \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "name": "Test Tenant",
    "subdomain": "testco",
    "plan_type": "starter",
    "description": "Test tenant for curl testing"
  }' \
  "$BASE_URL/api/v1/tenants"
echo ""
echo "---"

# Test 6: Test subdomain-based tenant extraction
echo "📋 Test 6: Subdomain-based tenant extraction"
echo "GET http://bo.testco.localhost:8090/api/v1/tenants"
curl -s -w "\nStatus: %{http_code}\n" \
  -H "Content-Type: application/json" \
  --resolve "bo.testco.localhost:8090:127.0.0.1" \
  "http://bo.testco.localhost:8090/api/v1/tenants"
echo ""
echo "---"

echo "✅ Tenant API tests completed!"
echo ""
echo "Expected results:"
echo "- Test 1: 200 (health check should work)"
echo "- Test 2: 400 (no tenant context)"
echo "- Test 3: 200 (valid tenant header)"
echo "- Test 4: 200 or 404 (depending on tenant existence)"
echo "- Test 5: 201 or error (tenant creation)"
echo "- Test 6: 200 (subdomain extraction)"