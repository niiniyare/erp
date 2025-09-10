#!/bin/bash

# Organization API Test Script for GOA-converted Entity Service
# Tests hierarchical multi-tenant organization structure with proper isolation

# Source common utilities and environment
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/setup_env.sh"
source "$SCRIPT_DIR/common.sh"

# Configuration - Read from environment or use defaults from config
SERVER_URL="${SERVER_URL:-http://localhost:${SERVER_PORT}}"
TENANT_HEADER="X-Tenant-ID"
DEFAULT_TENANT="01c02990-6d9c-41a4-88e8-c9ae77dcb6ed"

# Test data
declare -A TEST_ORGS=(
    ["root_company"]='{"name":"ACME Corporation","entity_type":"COMPANY","legal_entity_type":"CORPORATION","description":"Root company for ACME Corp","website":"https://acme.com"}'
    ["subsidiary"]='{"name":"ACME Europe GmbH","entity_type":"SUBSIDIARY","legal_entity_type":"LLC","description":"European subsidiary","parent_id":"PARENT_ID_PLACEHOLDER"}'
    ["region_west"]='{"name":"West Region","entity_type":"REGION","description":"Western region operations","parent_id":"PARENT_ID_PLACEHOLDER"}'
    ["branch_sf"]='{"name":"San Francisco Branch","entity_type":"BRANCH","description":"SF office and operations","parent_id":"PARENT_ID_PLACEHOLDER"}'
    ["department_eng"]='{"name":"Engineering Department","entity_type":"DEPARTMENT","description":"Software engineering team","parent_id":"PARENT_ID_PLACEHOLDER"}'
)

# Test results storage
declare -A CREATED_ORGS
declare -A TEST_RESULTS

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Test organization creation
test_create_organization() {
    local org_key="$1"
    local org_data="$2"
    local tenant="$3"
    
    log_info "Creating organization: $org_key"
    
    local response
    response=$(curl -s -w "%{http_code}" \
        -H "Content-Type: application/json" \
        -H "$TENANT_HEADER: $tenant" \
        -d "$org_data" \
        "$SERVER_URL/api/v1/organizations")
    
    local http_code="${response: -3}"
    local body="${response%???}"
    
    if [ "$http_code" = "201" ]; then
        local org_id
        org_id=$(echo "$body" | jq -r '.id // empty')
        if [ -n "$org_id" ] && [ "$org_id" != "null" ]; then
            CREATED_ORGS["$org_key"]="$org_id"
            TEST_RESULTS["create_$org_key"]="PASS"
            log_success "Created organization $org_key with ID: $org_id"
            return 0
        else
            TEST_RESULTS["create_$org_key"]="FAIL"
            log_error "Created organization but no ID returned"
            return 1
        fi
    else
        TEST_RESULTS["create_$org_key"]="FAIL"
        log_error "Failed to create organization $org_key (HTTP $http_code): $body"
        return 1
    fi
}

# Test organization retrieval
test_get_organization() {
    local org_key="$1"
    local tenant="$2"
    
    local org_id="${CREATED_ORGS[$org_key]}"
    if [ -z "$org_id" ]; then
        log_error "No ID found for organization $org_key"
        TEST_RESULTS["get_$org_key"]="SKIP"
        return 1
    fi
    
    log_info "Getting organization: $org_key ($org_id)"
    
    local response
    response=$(curl -s -w "%{http_code}" \
        -H "$TENANT_HEADER: $tenant" \
        "$SERVER_URL/api/v1/organizations/$org_id")
    
    local http_code="${response: -3}"
    local body="${response%???}"
    
    if [ "$http_code" = "200" ]; then
        local retrieved_id
        retrieved_id=$(echo "$body" | jq -r '.id // empty')
        if [ "$retrieved_id" = "$org_id" ]; then
            TEST_RESULTS["get_$org_key"]="PASS"
            log_success "Retrieved organization $org_key successfully"
            return 0
        else
            TEST_RESULTS["get_$org_key"]="FAIL"
            log_error "Retrieved organization but ID mismatch"
            return 1
        fi
    else
        TEST_RESULTS["get_$org_key"]="FAIL"
        log_error "Failed to get organization $org_key (HTTP $http_code): $body"
        return 1
    fi
}

# Test organization listing
test_list_organizations() {
    local tenant="$1"
    local expected_count="$2"
    
    log_info "Listing organizations for tenant: $tenant"
    
    local response
    response=$(curl -s -w "%{http_code}" \
        -H "$TENANT_HEADER: $tenant" \
        "$SERVER_URL/api/v1/organizations?page=1&page_size=20")
    
    local http_code="${response: -3}"
    local body="${response%???}"
    
    if [ "$http_code" = "200" ]; then
        local count
        count=$(echo "$body" | jq -r '.data | length')
        
        if [ "$count" -ge "$expected_count" ]; then
            TEST_RESULTS["list_orgs"]="PASS"
            log_success "Listed $count organizations (expected at least $expected_count)"
            return 0
        else
            TEST_RESULTS["list_orgs"]="FAIL"
            log_error "Listed $count organizations but expected at least $expected_count"
            return 1
        fi
    else
        TEST_RESULTS["list_orgs"]="FAIL"
        log_error "Failed to list organizations (HTTP $http_code): $body"
        return 1
    fi
}

# Test organization hierarchy
test_organization_hierarchy() {
    local parent_key="$1"
    local tenant="$2"
    
    local parent_id="${CREATED_ORGS[$parent_key]}"
    if [ -z "$parent_id" ]; then
        log_error "No ID found for parent organization $parent_key"
        TEST_RESULTS["hierarchy_$parent_key"]="SKIP"
        return 1
    fi
    
    log_info "Getting hierarchy for organization: $parent_key ($parent_id)"
    
    local response
    response=$(curl -s -w "%{http_code}" \
        -H "$TENANT_HEADER: $tenant" \
        "$SERVER_URL/api/v1/organizations/$parent_id/hierarchy?depth=3")
    
    local http_code="${response: -3}"
    local body="${response%???}"
    
    if [ "$http_code" = "200" ]; then
        local root_id
        root_id=$(echo "$body" | jq -r '.root.id // empty')
        if [ "$root_id" = "$parent_id" ]; then
            TEST_RESULTS["hierarchy_$parent_key"]="PASS"
            log_success "Retrieved hierarchy for $parent_key successfully"
            return 0
        else
            TEST_RESULTS["hierarchy_$parent_key"]="FAIL"
            log_error "Hierarchy root ID mismatch"
            return 1
        fi
    else
        TEST_RESULTS["hierarchy_$parent_key"]="FAIL"
        log_error "Failed to get hierarchy for $parent_key (HTTP $http_code): $body"
        return 1
    fi
}

# Test multi-tenant isolation
test_tenant_isolation() {
    local tenant1="$1"
    local tenant2="$2"
    
    log_info "Testing tenant isolation between $tenant1 and $tenant2"
    
    # Create org in tenant1
    local org_data='{"name":"Isolation Test Org","entity_type":"COMPANY","description":"Test organization for isolation"}'
    local response1
    response1=$(curl -s -w "%{http_code}" \
        -H "Content-Type: application/json" \
        -H "$TENANT_HEADER: $tenant1" \
        -d "$org_data" \
        "$SERVER_URL/api/v1/organizations")
    
    local http_code1="${response1: -3}"
    local body1="${response1%???}"
    
    if [ "$http_code1" != "201" ]; then
        TEST_RESULTS["isolation_test"]="FAIL"
        log_error "Failed to create test organization in $tenant1"
        return 1
    fi
    
    local org_id
    org_id=$(echo "$body1" | jq -r '.id // empty')
    
    # Try to access from tenant2
    local response2
    response2=$(curl -s -w "%{http_code}" \
        -H "$TENANT_HEADER: $tenant2" \
        "$SERVER_URL/api/v1/organizations/$org_id")
    
    local http_code2="${response2: -3}"
    
    if [ "$http_code2" = "404" ] || [ "$http_code2" = "403" ]; then
        TEST_RESULTS["isolation_test"]="PASS"
        log_success "Tenant isolation working correctly (HTTP $http_code2)"
        return 0
    else
        TEST_RESULTS["isolation_test"]="FAIL"
        log_error "Tenant isolation failed - organization accessible across tenants (HTTP $http_code2)"
        return 1
    fi
}

# Print test results
print_results() {
    log_info "=== ORGANIZATION API TEST RESULTS ==="
    
    local total=0
    local passed=0
    local failed=0
    local skipped=0
    
    for test in "${!TEST_RESULTS[@]}"; do
        local result="${TEST_RESULTS[$test]}"
        total=$((total + 1))
        
        case $result in
            "PASS")
                echo -e "${GREEN}✓${NC} $test: PASSED"
                passed=$((passed + 1))
                ;;
            "FAIL")
                echo -e "${RED}✗${NC} $test: FAILED"
                failed=$((failed + 1))
                ;;
            "SKIP")
                echo -e "${YELLOW}○${NC} $test: SKIPPED"
                skipped=$((skipped + 1))
                ;;
        esac
    done
    
    echo ""
    log_info "Summary: $passed passed, $failed failed, $skipped skipped out of $total tests"
    
    if [ $failed -eq 0 ]; then
        log_success "All tests passed! 🎉"
        return 0
    else
        log_error "Some tests failed! ❌"
        return 1
    fi
}

# Create test tenant if it doesn't exist
create_test_tenant() {
    log_info "Creating test tenant if it doesn't exist..."
    
    local tenant_data='{"name":"ACME Corp","plan_type":"premium","description":"Test tenant for organization API","settings":{"timezone":"UTC","currency":"USD","date_format":"MM/DD/YYYY","language":"en"}}'
    
    local response
    response=$(curl -s -w "%{http_code}" \
        -H "Content-Type: application/json" \
        -d "$tenant_data" \
        "$SERVER_URL/api/v1/tenants")
    
    local http_code="${response: -3}"
    local body="${response%???}"
    
    if [ "$http_code" = "201" ] || [ "$http_code" = "409" ]; then
        log_success "Test tenant is available"
        return 0
    else
        log_warning "Could not create test tenant (HTTP $http_code): $body"
        log_info "Proceeding with tests - tenant may already exist"
        return 0
    fi
}

# Main test execution
main() {
    log_info "Starting Organization API Tests"
    log_info "Server: $SERVER_URL"
    log_info "Default Tenant: $DEFAULT_TENANT"
    
    # Check server connectivity
    if ! check_server_connectivity "$SERVER_URL"; then
        log_error "Server is not accessible. Please start the server and try again."
        exit 1
    fi
    
    # Create test tenant if needed
    create_test_tenant
    
    log_info "=== TESTING ORGANIZATION CREATION ==="
    
    # Create root company first
    test_create_organization "root_company" "${TEST_ORGS[root_company]}" "$DEFAULT_TENANT"
    
    # Create subsidiary with parent reference
    if [ -n "${CREATED_ORGS[root_company]}" ]; then
        local subsidiary_data="${TEST_ORGS[subsidiary]}"
        subsidiary_data="${subsidiary_data/PARENT_ID_PLACEHOLDER/${CREATED_ORGS[root_company]}}"
        test_create_organization "subsidiary" "$subsidiary_data" "$DEFAULT_TENANT"
        
        # Create region under root company
        local region_data="${TEST_ORGS[region_west]}"
        region_data="${region_data/PARENT_ID_PLACEHOLDER/${CREATED_ORGS[root_company]}}"
        test_create_organization "region_west" "$region_data" "$DEFAULT_TENANT"
        
        # Create branch under region
        if [ -n "${CREATED_ORGS[region_west]}" ]; then
            local branch_data="${TEST_ORGS[branch_sf]}"
            branch_data="${branch_data/PARENT_ID_PLACEHOLDER/${CREATED_ORGS[region_west]}}"
            test_create_organization "branch_sf" "$branch_data" "$DEFAULT_TENANT"
            
            # Create department under branch
            if [ -n "${CREATED_ORGS[branch_sf]}" ]; then
                local dept_data="${TEST_ORGS[department_eng]}"
                dept_data="${dept_data/PARENT_ID_PLACEHOLDER/${CREATED_ORGS[branch_sf]}}"
                test_create_organization "department_eng" "$dept_data" "$DEFAULT_TENANT"
            fi
        fi
    fi
    
    log_info "=== TESTING ORGANIZATION RETRIEVAL ==="
    
    # Test retrieval of created organizations
    for org_key in "${!CREATED_ORGS[@]}"; do
        test_get_organization "$org_key" "$DEFAULT_TENANT"
    done
    
    log_info "=== TESTING ORGANIZATION LISTING ==="
    
    # Test listing organizations
    test_list_organizations "$DEFAULT_TENANT" 3
    
    log_info "=== TESTING ORGANIZATION HIERARCHY ==="
    
    # Test hierarchy retrieval
    test_organization_hierarchy "root_company" "$DEFAULT_TENANT"
    
    log_info "=== TESTING MULTI-TENANT ISOLATION ==="
    
    # Test tenant isolation
    test_tenant_isolation "$DEFAULT_TENANT" "other-tenant"
    
    # Print final results
    print_results
}

# Execute main function if script is run directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi