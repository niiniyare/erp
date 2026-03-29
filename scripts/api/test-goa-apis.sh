#!/bin/bash

# GOA API Test Script
# Tests all currently implemented GOA endpoints

set -e  # Exit on any error

# Configuration
BASE_URL="http://localhost:8080"
VERBOSE=false
CLEANUP=true

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

log_header() {
    echo -e "${PURPLE} $1${NC}"
}

log_section() {
    echo -e "${CYAN} $1${NC}"
}

# Check if jq is available
if ! command -v jq &> /dev/null; then
    log_error "jq is required but not installed. Please install jq first."
    exit 1
fi

# Check if curl is available
if ! command -v curl &> /dev/null; then
    log_error "curl is required but not installed. Please install curl first."
    exit 1
fi

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        --no-cleanup)
            CLEANUP=false
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [options]"
            echo "Options:"
            echo "  -v, --verbose     Enable verbose output"
            echo "  --no-cleanup      Don't cleanup test data"
            echo "  -h, --help        Show this help message"
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Test helper functions
make_request() {
    local method="$1"
    local endpoint="$2"
    local data="$3"
    local headers="$4"
    local expected_status="$5"
    
    if [ "$VERBOSE" = true ]; then
        log_info "Making $method request to $endpoint"
    fi
    
    local curl_cmd="curl -s -w '%{http_code}' -X $method $BASE_URL$endpoint"
    
    if [ -n "$headers" ]; then
        curl_cmd="$curl_cmd $headers"
    fi
    
    if [ -n "$data" ]; then
        curl_cmd="$curl_cmd -d '$data'"
    fi
    
    local response=$(eval $curl_cmd)
    local status_code="${response: -3}"
    local body="${response%???}"
    
    # Remove any trailing newlines that might interfere with JSON parsing
    body=$(echo "$body" | tr -d '\n')
    
    if [ "$status_code" != "$expected_status" ]; then
        log_error "Expected status $expected_status, got $status_code"
        if [ "$VERBOSE" = true ]; then
            echo "Response body: $body"
        fi
        return 1
    fi
    
    echo "$body"
}

# Store created resource IDs for cleanup
CREATED_TENANT_IDS=()
CREATED_ORG_IDS=()
CREATED_USER_IDS=()

# Cleanup function
cleanup_resources() {
    if [ "$CLEANUP" = false ]; then
        log_info "Skipping cleanup (--no-cleanup flag used)"
        return
    fi
    
    log_section "Cleaning up test resources"
    
    # Archive organizations
    for ORG_ID in "${CREATED_ORG_IDS[@]}"; do
        if [ "$VERBOSE" = true ]; then
            log_info "Archiving organization $ORG_ID"
        fi
        curl -s -X PATCH "$BASE_URL/api/v1/organizations/$ORG_ID/archive" \
            -H "Content-Type: application/json" \
            -d '{"reason": "Test cleanup"}' > /dev/null || true
    done
    
    # Deactivate users
    for USER_ID in "${CREATED_USER_IDS[@]}"; do
        if [ "$VERBOSE" = true ]; then
            log_info "Deactivating user $USER_ID"
        fi
        curl -s -X PATCH "$BASE_URL/api/v1/users/$USER_ID/deactivate" > /dev/null || true
    done
    
    # Delete tenants
    for TENANT_ID in "${CREATED_TENANT_IDS[@]}"; do
        if [ "$VERBOSE" = true ]; then
            log_info "Deleting tenant $TENANT_ID"
        fi
        curl -s -X DELETE "$BASE_URL/api/v1/tenants/$TENANT_ID" > /dev/null || true
    done
    
    log_success "Cleanup completed"
}

# Set trap for cleanup on exit
trap cleanup_resources EXIT

# Main test functions
test_health_endpoints() {
    log_section "Testing Health & Status Endpoints"
    
    # Test tenant health
    log_info "Testing tenant health endpoint"
    local health_response=$(make_request "GET" "/api/v1/tenants/health" "" "" "200")
    local status=$(echo "$health_response" | jq -r '.status')
    
    if [ "$status" = "healthy" ]; then
        log_success "Tenant health check passed"
    else
        log_error "Tenant health check failed - status: $status"
        return 1
    fi
    
    # Test OpenAPI spec
    log_info "Testing OpenAPI specification endpoint"
    make_request "GET" "/openapi.json" "" "" "200" > /dev/null
    log_success "OpenAPI specification endpoint working"
    
    # Test Swagger UI
    log_info "Testing Swagger UI endpoint"
    make_request "GET" "/swagger-ui" "" "" "200" > /dev/null
    log_success "Swagger UI endpoint working"
}

test_tenant_api() {
    log_section "Testing Tenant Management API"
    
    # Create tenant
    log_info "Creating test tenant"
    local tenant_data='{"name": "Test Tenant Company"}'
    local create_response=$(make_request "POST" "/api/v1/tenants" "$tenant_data" "-H 'Content-Type: application/json'" "200")
    local tenant_id=$(echo "$create_response" | jq -r '.id')
    
    if [ "$tenant_id" != "null" ] && [ -n "$tenant_id" ]; then
        CREATED_TENANT_IDS+=("$tenant_id")
        log_success "Tenant created: $tenant_id"
    else
        log_error "Failed to create tenant"
        return 1
    fi
    
    # Get tenant
    log_info "Getting tenant by ID"
    local get_response=$(make_request "GET" "/api/v1/tenants/$tenant_id" "" "" "200")
    local retrieved_name=$(echo "$get_response" | jq -r '.name')
    
    if [ "$retrieved_name" = "Test Tenant Company" ]; then
        log_success "Tenant retrieved successfully"
    else
        log_error "Failed to retrieve tenant correctly"
        return 1
    fi
    
    # Update tenant
    log_info "Updating tenant"
    local update_data='{"name": "Updated Test Tenant Company"}'
    make_request "PUT" "/api/v1/tenants/$tenant_id" "$update_data" "-H 'Content-Type: application/json'" "200" > /dev/null
    log_success "Tenant updated successfully"
    
    # List tenants
    log_info "Listing tenants"
    local list_response=$(make_request "GET" "/api/v1/tenants" "" "" "200")
    local tenant_count=$(echo "$list_response" | jq '.data | length')
    
    if [ "$tenant_count" -ge 1 ]; then
        log_success "Tenant listing working - found $tenant_count tenants"
    else
        log_error "Failed to list tenants"
        return 1
    fi
}

test_organization_api() {
    log_section "Testing Organization Management API"
    
    # Create corporation
    log_info "Creating test corporation"
    local corp_data='{"name": "Test Corporation", "organization_type": "CORPORATION"}'
    local corp_response=$(make_request "POST" "/api/v1/organizations" "$corp_data" "-H 'Content-Type: application/json'" "200")
    local corp_id=$(echo "$corp_response" | jq -r '.id')
    
    if [ "$corp_id" != "null" ] && [ -n "$corp_id" ]; then
        CREATED_ORG_IDS+=("$corp_id")
        log_success "Corporation created: $corp_id"
    else
        log_error "Failed to create corporation"
        return 1
    fi
    
    # Create department under corporation
    log_info "Creating test department"
    local dept_data='{"name": "Test Department", "organization_type": "DEPARTMENT", "parent_id": "'$corp_id'"}'
    local dept_response=$(make_request "POST" "/api/v1/organizations" "$dept_data" "-H 'Content-Type: application/json'" "200")
    local dept_id=$(echo "$dept_response" | jq -r '.id')
    
    if [ "$dept_id" != "null" ] && [ -n "$dept_id" ]; then
        CREATED_ORG_IDS+=("$dept_id")
        log_success "Department created: $dept_id"
    else
        log_error "Failed to create department"
        return 1
    fi
    
    # Get organization
    log_info "Getting organization by ID"
    make_request "GET" "/api/v1/organizations/$corp_id" "" "" "200" > /dev/null
    log_success "Organization retrieved successfully"
    
    # Update organization
    log_info "Updating organization"
    local update_data='{"name": "Updated Test Corporation"}'
    make_request "PUT" "/api/v1/organizations/$corp_id" "$update_data" "-H 'Content-Type: application/json'" "200" > /dev/null
    log_success "Organization updated successfully"
    
    # Get hierarchy
    log_info "Getting organization hierarchy"
    local hierarchy_response=$(make_request "GET" "/api/v1/organizations/$corp_id/hierarchy" "" "" "200")
    local children_count=$(echo "$hierarchy_response" | jq '.children | length')
    
    if [ "$children_count" -ge 1 ]; then
        log_success "Organization hierarchy working - found $children_count children"
    else
        log_error "Failed to get organization hierarchy"
        return 1
    fi
    
    # List organizations
    log_info "Listing organizations"
    local list_response=$(make_request "GET" "/api/v1/organizations" "" "" "200")
    local org_count=$(echo "$list_response" | jq '.data | length')
    
    if [ "$org_count" -ge 2 ]; then
        log_success "Organization listing working - found $org_count organizations"
    else
        log_error "Failed to list organizations"
        return 1
    fi
}

test_user_api() {
    log_section "Testing User Management API"
    
    # Create user
    log_info "Creating test user"
    local user_data='{"username": "testuser", "email": "test@example.com", "first_name": "Test", "last_name": "User", "user_type": "INTERNAL"}'
    local user_response=$(make_request "POST" "/api/v1/users" "$user_data" "-H 'Content-Type: application/json'" "200")
    local user_id=$(echo "$user_response" | jq -r '.id')
    
    if [ "$user_id" != "null" ] && [ -n "$user_id" ]; then
        CREATED_USER_IDS+=("$user_id")
        log_success "User created: $user_id"
    else
        log_error "Failed to create user"
        return 1
    fi
    
    # Get user
    log_info "Getting user by ID"
    local get_response=$(make_request "GET" "/api/v1/users/$user_id" "" "" "200")
    local retrieved_email=$(echo "$get_response" | jq -r '.email')
    
    if [ "$retrieved_email" = "test@example.com" ]; then
        log_success "User retrieved successfully"
    else
        log_error "Failed to retrieve user correctly"
        return 1
    fi
    
    # Update user
    log_info "Updating user"
    local update_data='{"first_name": "Updated", "last_name": "TestUser"}'
    make_request "PUT" "/api/v1/users/$user_id" "$update_data" "-H 'Content-Type: application/json'" "200" > /dev/null
    log_success "User updated successfully"
    
    # Get user permissions
    log_info "Getting user permissions"
    make_request "GET" "/api/v1/users/$user_id/permissions" "" "" "200" > /dev/null
    log_success "User permissions retrieved successfully"
    
    # List users
    log_info "Listing users"
    local list_response=$(make_request "GET" "/api/v1/users" "" "" "200")
    local user_count=$(echo "$list_response" | jq '.data | length')
    
    if [ "$user_count" -ge 1 ]; then
        log_success "User listing working - found $user_count users"
    else
        log_error "Failed to list users"
        return 1
    fi
}

test_auth_api() {
    log_section "Testing Authentication API"
    
    # Test login (will use mock response)
    log_info "Testing user login"
    local login_data='{"email": "test@example.com", "password": "password123"}'
    local login_response=$(make_request "POST" "/api/v1/auth/login" "$login_data" "-H 'Content-Type: application/json'" "200")
    local access_token=$(echo "$login_response" | jq -r '.access_token')
    
    if [ "$access_token" != "null" ] && [ -n "$access_token" ]; then
        log_success "User login successful - token received"
    else
        log_error "Failed to login user"
        return 1
    fi
    
    # Test token validation
    log_info "Testing token validation"
    make_request "GET" "/api/v1/auth/validate" "" "-H 'Authorization: Bearer $access_token'" "200" > /dev/null
    log_success "Token validation successful"
    
    # Test token refresh
    log_info "Testing token refresh"
    local refresh_data='{"refresh_token": "mock-refresh-token"}'
    make_request "POST" "/api/v1/auth/refresh" "$refresh_data" "-H 'Content-Type: application/json'" "200" > /dev/null
    log_success "Token refresh successful"
    
    # Test logout
    log_info "Testing user logout"
    local logout_data="{\"token\": \"$access_token\"}"
    make_request "POST" "/api/v1/auth/logout" "$logout_data" "-H 'Authorization: Bearer $access_token' -H 'Content-Type: application/json'" "200" > /dev/null
    log_success "User logout successful"
}

# Main execution
main() {
    log_header "Starting GOA API Tests"
    log_info "Base URL: $BASE_URL"
    log_info "Verbose: $VERBOSE"
    log_info "Cleanup: $CLEANUP"
    echo
    
    # Check server connectivity
    log_info "Checking server connectivity..."
    if ! curl -s --connect-timeout 5 "$BASE_URL/api/v1/tenants/health" > /dev/null; then
        log_error "Cannot connect to server at $BASE_URL"
        log_error "Please ensure the server is running: go run cmd/server/main.go"
        exit 1
    fi
    log_success "Server connectivity confirmed"
    echo
    
    # Run tests
    local start_time=$(date +%s)
    local failed_tests=0
    
    # Test each API group
    if ! test_health_endpoints; then
        ((failed_tests++))
    fi
    echo
    
    if ! test_tenant_api; then
        ((failed_tests++))
    fi
    echo
    
    if ! test_organization_api; then
        ((failed_tests++))
    fi
    echo
    
    if ! test_user_api; then
        ((failed_tests++))
    fi
    echo
    
    if ! test_auth_api; then
        ((failed_tests++))
    fi
    echo
    
    # Summary
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    log_header "Test Summary"
    if [ $failed_tests -eq 0 ]; then
        log_success "All tests passed! "
        log_info "Total runtime: ${duration}s"
        log_info "Created resources:"
        log_info "  - Tenants: ${#CREATED_TENANT_IDS[@]}"
        log_info "  - Organizations: ${#CREATED_ORG_IDS[@]}"
        log_info "  - Users: ${#CREATED_USER_IDS[@]}"
    else
        log_error "$failed_tests test group(s) failed"
        log_info "Total runtime: ${duration}s"
        exit 1
    fi
}

# Run main function
main "$@"