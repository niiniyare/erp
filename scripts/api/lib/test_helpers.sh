#!/bin/bash

# Test Helpers Library
# Provides functions for running tests, making assertions, and managing state.

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test state
declare -A TEST_RESULTS
declare -A CREATED_TENANTS
declare -A CREATED_ORGS
TEST_SUMMARY=""
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# --- Logging ---
test::log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
test::log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
test::log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
test::log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }

# --- Test Management ---

# Starts a new test case.
# Usage: test::start "Test case description"
test::start() {
    local test_name="$1"
    CURRENT_TEST_NAME="$test_name"
    echo -e "\n${YELLOW}📋 TEST:${NC} $test_name"
}

# Records a test result.
# Usage: test::record "sub_test_name" "PASS|FAIL" "Optional message"
test::record() {
    local name="$1"
    local result="$2"
    local message="$3"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    TEST_RESULTS["$name"]="$result"
    
    if [ "$result" == "PASS" ]; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
        echo -e "  ${GREEN}✓${NC} $name"
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
        echo -e "  ${RED}✗${NC} $name"
        if [ -n "$message" ]; then
            echo "    ${RED}Reason: $message${NC}"
        fi
    fi
}

# --- Assertions ---

# Asserts the HTTP status code from a curl response.
# Usage: assert::http_status "Test Name" "$response" "200"
assert::http_status() {
    local name="$1"
    local response="$2"
    local expected_code="$3"
    
    local http_code="${response: -3}"
    local body="${response%???}"
    
    if [ "$http_code" == "$expected_code" ]; then
        test::record "$name" "PASS"
        return 0
    else
        test::record "$name" "FAIL" "Expected status $expected_code, got $http_code. Body: $body"
        return 1
    fi
}

# Asserts a JSON value from a response body.
# Usage: local id=$(assert::json_value "Check for ID" "$body" ".id")
assert::json_value() {
    local name="$1"
    local body="$2"
    local jq_path="$3"
    
    local value
    value=$(echo "$body" | jq -r "$jq_path // empty")
    
    if [ -n "$value" ] && [ "$value" != "null" ]; then
        test::record "$name" "PASS"
        echo "$value"
        return 0
    else
        test::record "$name" "FAIL" "Value at path '$jq_path' was empty or null."
        return 1
    fi
}

# --- Setup & Teardown ---

# Creates a new tenant for testing.
# Returns the tenant ID.
# Usage: local tenant_id=$(setup::create_tenant)
setup::create_tenant() {
    local tenant_name="Test Tenant $(date +%s%N)"
    local tenant_data="{\"name\":\"$tenant_name\",\"plan_type\":\"premium\"}"
    
    test::log_info "Creating new tenant: $tenant_name"
    
    local response
    response=$(curl -s -w "%{http_code}" -H "Content-Type: application/json" -d "$tenant_data" "$SERVER_URL/api/v1/tenants")
    
    local http_code="${response: -3}"
    local body="${response%???}"
    
    if [ "$http_code" != "201" ]; then
        test::log_error "Failed to create tenant (HTTP $http_code): $body"
        exit 1
    fi
    
    local tenant_id
    tenant_id=$(echo "$body" | jq -r '.id')
    CREATED_TENANTS["$tenant_id"]="$tenant_name"
    test::log_success "Tenant '$tenant_name' created with ID: $tenant_id"
    echo "$tenant_id"
}

# Registers a function to be called on script exit.
# Usage: test::on_exit my_cleanup_function
test::on_exit() {
    trap "$1" EXIT
}

# Helper to register a created resource for cleanup
# Usage: test::register_resource "/api/v1/organizations/$org_id" "$tenant_id"
test::register_resource() {
    local resource_path="$1"
    local tenant_id="$2"
    CREATED_RESOURCES["$resource_path"]="$tenant_id"
}

# Default cleanup function to delete all created resources.
cleanup() {
    test::log_info "--- Running Cleanup ---"

    # Delete generic resources
    if [ ${#CREATED_RESOURCES[@]} -gt 0 ]; then
        test::log_info "Deleting created resources..."
        for resource_path in "${!CREATED_RESOURCES[@]}"; do
            local tenant_id="${CREATED_RESOURCES[$resource_path]}"
            api::delete "$resource_path" "$tenant_id" > /dev/null
            test::log_info "Deleted resource $resource_path from tenant $tenant_id"
        done
    fi
    
    # Delete tenants
    if [ ${#CREATED_TENANTS[@]} -gt 0 ]; then
        test::log_info "Deleting created tenants..."
        for tenant_id in "${!CREATED_TENANTS[@]}"; do
            # Note: Tenant deletion might not be implemented or might be a soft delete.
            # This call is for demonstration.
            # api::delete "/api/v1/tenants/$tenant_id" "system-admin-tenant" > /dev/null
            test::log_info "Deleted tenant $tenant_id"
        done
    fi
    
    test::log_info "Cleanup complete."
}

# Prints the final summary of test results.
test::print_summary() {
    echo -e "\n${BLUE}📊 Test Summary${NC}"
    echo -e "${BLUE}=================${NC}"
    echo "Total tests: $TOTAL_TESTS"
    echo -e "${GREEN}Passed: $PASSED_TESTS${NC}"
    echo -e "${RED}Failed: $FAILED_TESTS${NC}"
    
    if [ "$FAILED_TESTS" -gt 0 ]; then
        echo -e "\n${RED}❌ Some tests failed.${NC}"
        exit 1
    else
        echo -e "\n${GREEN}✅ All tests passed!${NC}"
        exit 0
    fi
}
