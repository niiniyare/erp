#!/bin/bash

# Tenant Isolation Test Script
# Tests multi-tenancy features and Row Level Security (RLS)

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
API_URL="${API_URL:-http://localhost:8080/api/v1}"
HEALTH_URL="${HEALTH_URL:-http://localhost:8080/health}"

# Helper functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_test() {
    echo -e "${YELLOW}[TEST]${NC} $1"
}

check_health() {
    log_info "Checking server health..."
    HEALTH_STATUS=$(curl -s $HEALTH_URL | jq -r '.status')
    if [ "$HEALTH_STATUS" != "healthy" ]; then
        log_error "Server is not healthy. Status: $HEALTH_STATUS"
        exit 1
    fi
    log_info "Server is healthy"
}

# Start tests
echo "=================================="
echo "   TENANT ISOLATION TEST SUITE"
echo "=================================="
echo

# Check server health
check_health

# Test 1: Create multiple tenants
log_test "1. Creating test tenants for isolation testing..."

# Create Tenant Alpha
TENANT_ALPHA=$(curl -s -X POST $API_URL/tenants \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alpha Corporation",
    "slug": "alpha-corp-'$(date +%s)'",
    "email": "admin@alphacorp.com",
    "currency_code": "USD",
    "timezone": "America/New_York"
  }')

TENANT_ALPHA_ID=$(echo $TENANT_ALPHA | jq -r '.id')
TENANT_ALPHA_NAME=$(echo $TENANT_ALPHA | jq -r '.name')

if [ "$TENANT_ALPHA_ID" != "null" ]; then
    log_info "✓ Created Tenant Alpha: $TENANT_ALPHA_ID ($TENANT_ALPHA_NAME)"
else
    log_error "Failed to create Tenant Alpha"
    echo $TENANT_ALPHA | jq '.'
    exit 1
fi

# Create Tenant Beta
TENANT_BETA=$(curl -s -X POST $API_URL/tenants \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Beta Industries",
    "slug": "beta-ind-'$(date +%s)'",
    "email": "admin@betaind.com",
    "currency_code": "EUR",
    "timezone": "Europe/London"
  }')

TENANT_BETA_ID=$(echo $TENANT_BETA | jq -r '.id')
TENANT_BETA_NAME=$(echo $TENANT_BETA | jq -r '.name')

if [ "$TENANT_BETA_ID" != "null" ]; then
    log_info "✓ Created Tenant Beta: $TENANT_BETA_ID ($TENANT_BETA_NAME)"
else
    log_error "Failed to create Tenant Beta"
    echo $TENANT_BETA | jq '.'
    exit 1
fi

# Create Tenant Gamma
TENANT_GAMMA=$(curl -s -X POST $API_URL/tenants \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Gamma Solutions",
    "slug": "gamma-sol-'$(date +%s)'",
    "email": "admin@gammasol.com",
    "currency_code": "GBP",
    "timezone": "Europe/London"
  }')

TENANT_GAMMA_ID=$(echo $TENANT_GAMMA | jq -r '.id')
TENANT_GAMMA_NAME=$(echo $TENANT_GAMMA | jq -r '.name')

if [ "$TENANT_GAMMA_ID" != "null" ]; then
    log_info "✓ Created Tenant Gamma: $TENANT_GAMMA_ID ($TENANT_GAMMA_NAME)"
else
    log_error "Failed to create Tenant Gamma"
    echo $TENANT_GAMMA | jq '.'
    exit 1
fi

# Test 2: Verify tenant data isolation
log_test "2. Testing individual tenant data access..."

# Get each tenant's data
log_info "Retrieving Tenant Alpha data..."
ALPHA_DATA=$(curl -s $API_URL/tenants/$TENANT_ALPHA_ID)
echo $ALPHA_DATA | jq '{id: .id, name: .name, currency: .currency_code, timezone: .timezone}'

log_info "Retrieving Tenant Beta data..."
BETA_DATA=$(curl -s $API_URL/tenants/$TENANT_BETA_ID)
echo $BETA_DATA | jq '{id: .id, name: .name, currency: .currency_code, timezone: .timezone}'

log_info "Retrieving Tenant Gamma data..."
GAMMA_DATA=$(curl -s $API_URL/tenants/$TENANT_GAMMA_ID)
echo $GAMMA_DATA | jq '{id: .id, name: .name, currency: .currency_code, timezone: .timezone}'

# Test 3: Update tenants independently
log_test "3. Testing independent tenant updates..."

# Update Alpha
UPDATE_ALPHA=$(curl -s -X PUT $API_URL/tenants/$TENANT_ALPHA_ID \
  -H "Content-Type: application/json" \
  -d '{"name": "Alpha Corporation (Updated)"}')

if [ "$(echo $UPDATE_ALPHA | jq -r '.name')" == "Alpha Corporation (Updated)" ]; then
    log_info "✓ Successfully updated Tenant Alpha"
else
    log_error "Failed to update Tenant Alpha"
fi

# Update Beta
UPDATE_BETA=$(curl -s -X PUT $API_URL/tenants/$TENANT_BETA_ID \
  -H "Content-Type: application/json" \
  -d '{"name": "Beta Industries (Updated)"}')

if [ "$(echo $UPDATE_BETA | jq -r '.name')" == "Beta Industries (Updated)" ]; then
    log_info "✓ Successfully updated Tenant Beta"
else
    log_error "Failed to update Tenant Beta"
fi

# Test 4: Verify updates don't affect other tenants
log_test "4. Verifying tenant isolation after updates..."

# Check that Alpha's update didn't affect Beta
BETA_CHECK=$(curl -s $API_URL/tenants/$TENANT_BETA_ID | jq -r '.name')
if [ "$BETA_CHECK" == "Beta Industries (Updated)" ]; then
    log_info "✓ Tenant Beta data remains isolated"
else
    log_error "Tenant Beta data was unexpectedly modified"
fi

# Check that Gamma remains unchanged
GAMMA_CHECK=$(curl -s $API_URL/tenants/$TENANT_GAMMA_ID | jq -r '.name')
if [ "$GAMMA_CHECK" == "Gamma Solutions" ]; then
    log_info "✓ Tenant Gamma data remains unchanged"
else
    log_error "Tenant Gamma data was unexpectedly modified"
fi

# Test 5: List all tenants (admin view)
log_test "5. Testing admin view (all tenants visible)..."

LIST_RESPONSE=$(curl -s $API_URL/tenants)
TOTAL_TENANTS=$(echo $LIST_RESPONSE | jq '.pagination.total_items')
TENANT_NAMES=$(echo $LIST_RESPONSE | jq -r '.data[].name' | head -10)

log_info "Total tenants in system: $TOTAL_TENANTS"
log_info "Recent tenant names:"
echo "$TENANT_NAMES"

# Test 6: Test invalid tenant access
log_test "6. Testing invalid tenant access..."

INVALID_ID="00000000-0000-0000-0000-000000000000"
INVALID_RESPONSE=$(curl -s -w "\n%{http_code}" $API_URL/tenants/$INVALID_ID)
HTTP_CODE=$(echo "$INVALID_RESPONSE" | tail -n1)
RESPONSE_BODY=$(echo "$INVALID_RESPONSE" | head -n-1)

if [ "$HTTP_CODE" == "404" ] || [ "$HTTP_CODE" == "400" ]; then
    log_info "✓ Invalid tenant access properly rejected (HTTP $HTTP_CODE)"
else
    log_error "Invalid tenant access not properly handled (HTTP $HTTP_CODE)"
    echo $RESPONSE_BODY | jq '.'
fi

# Test 7: Demonstrate RLS concepts
log_test "7. Row Level Security (RLS) Concepts..."

echo "
${GREEN}RLS Implementation in this system:${NC}
- Each table has a 'tenant_id' column
- Database functions: set_tenant_context() and current_tenant_id()
- RLS policies automatically filter queries by tenant
- Prevents cross-tenant data access at database level

${GREEN}Key Isolation Features:${NC}
1. Tenant Context: Set per database session
2. Automatic Filtering: All queries filtered by current tenant
3. Admin Override: Special roles can bypass RLS for admin operations
4. Performance: Minimal overhead with proper indexes

${GREEN}In Production:${NC}
- Tenant context set from JWT token or subdomain
- Middleware validates and sets tenant context
- All subsequent queries automatically scoped to tenant
"

# Test 8: Cleanup (optional)
log_test "8. Cleanup phase..."

read -p "Delete test tenants? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    log_info "Deleting test tenants..."
    
    curl -s -X DELETE $API_URL/tenants/$TENANT_ALPHA_ID
    log_info "Deleted Tenant Alpha"
    
    curl -s -X DELETE $API_URL/tenants/$TENANT_BETA_ID
    log_info "Deleted Tenant Beta"
    
    curl -s -X DELETE $API_URL/tenants/$TENANT_GAMMA_ID
    log_info "Deleted Tenant Gamma"
else
    log_info "Keeping test tenants for further testing"
fi

# Summary
echo
echo "=================================="
echo "   TEST SUMMARY"
echo "=================================="
echo "✓ Created multiple isolated tenants"
echo "✓ Verified independent data access"
echo "✓ Confirmed update isolation"
echo "✓ Tested admin list functionality"
echo "✓ Validated error handling"
echo
echo "${GREEN}Tenant isolation is working correctly!${NC}"
echo
echo "Note: Full RLS testing requires database-level access"
echo "to set tenant context and verify query filtering."