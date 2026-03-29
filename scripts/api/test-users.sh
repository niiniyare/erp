#!/bin/bash

# User API Automated Test Script
# This script tests all User API endpoints with error handling

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

log_test() {
    echo -e "${BLUE} Testing: $1${NC}"
}

# Check if jq is available
if ! command -v jq &> /dev/null; then
    log_error "jq is required but not installed. Please install jq first."
    echo "  Ubuntu/Debian: sudo apt-get install jq"
    echo "  macOS: brew install jq"
    echo "  Windows: choco install jq"
    exit 1
fi

# Check if server is running
check_server() {
    log_info "Checking if server is running..."
    if curl -s -f "$BASE_URL/health" > /dev/null; then
        log_success "Server is running"
        return 0
    else
        log_error "Server is not running at $BASE_URL"
        log_info "Please start the server first:"
        echo "  go run cmd/server/main.go"
        exit 1
    fi
}

# Test function that handles errors gracefully
test_endpoint() {
    local description="$1"
    local method="$2"
    local endpoint="$3"
    local data="$4"
    local expected_status="$5"
    
    log_test "$description"
    
    local curl_cmd="curl -s -w '%{http_code}' -X $method '$BASE_URL$endpoint'"
    
    if [[ -n "$data" ]]; then
        curl_cmd="$curl_cmd -H 'Content-Type: application/json' -d '$data'"
    fi
    
    if [[ "$VERBOSE" == "true" ]]; then
        echo "Command: $curl_cmd"
    fi
    
    local response=$(eval "$curl_cmd")
    local status_code="${response: -3}"
    local body="${response%???}"
    
    if [[ "$status_code" == "$expected_status" ]]; then
        log_success "Status: $status_code (Expected: $expected_status)"
        if [[ -n "$body" ]] && [[ "$body" != "null" ]]; then
            echo "$body" | jq . 2>/dev/null || echo "$body"
        fi
        echo "$body"  # Return response for further processing
    else
        log_error "Status: $status_code (Expected: $expected_status)"
        if [[ -n "$body" ]]; then
            echo "Response: $body" | jq . 2>/dev/null || echo "Response: $body"
        fi
        return 1
    fi
}

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
            echo "Usage: $0 [OPTIONS]"
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

# Start testing
echo -e "${BLUE} Starting User API Test Suite...${NC}"
echo "Base URL: $BASE_URL"
echo "Verbose: $VERBOSE"
echo "Cleanup: $CLEANUP"
echo

# Global variables for test data
ENTITY_ID=""
USER1_ID=""
USER2_ID=""
ADMIN_ID=""

# Test 1: Health Check
check_server

# Test 2: Create test entity (required for users)
log_test "Creating test entity for users"
entity_response=$(test_endpoint "Create test entity" "POST" "/api/v1/entities" '{
    "name": "Test Company",
    "code": "TEST_USER_001",
    "type": "account"
}' "201")

if [[ $? -eq 0 ]]; then
    ENTITY_ID=$(echo "$entity_response" | jq -r '.id // empty')
    if [[ -n "$ENTITY_ID" && "$ENTITY_ID" != "null" ]]; then
        log_success "Created entity with ID: $ENTITY_ID"
    else
        log_error "Failed to extract entity ID from response"
        exit 1
    fi
else
    log_error "Failed to create test entity"
    exit 1
fi

echo

# Test 3: Create Users
log_test "Creating test users"

# Create regular user
user1_response=$(test_endpoint "Create regular user" "POST" "/api/v1/users" '{
    "entity_id": "'$ENTITY_ID'",
    "username": "test.user1",
    "email": "test1@example.com",
    "password": "TestPassword123!",
    "user_type": "INTERNAL",
    "account_status": "ACTIVE",
    "mfa_enabled": false
}' "201")

if [[ $? -eq 0 ]]; then
    USER1_ID=$(echo "$user1_response" | jq -r '.id // empty')
    log_success "Created user 1 with ID: $USER1_ID"
fi

# Create admin user
admin_response=$(test_endpoint "Create admin user" "POST" "/api/v1/users" '{
    "entity_id": "'$ENTITY_ID'",
    "username": "admin.test",
    "email": "admin@example.com",
    "password": "AdminPassword123!",
    "user_type": "ADMIN",
    "account_status": "ACTIVE",
    "mfa_enabled": true,
    "session_timeout_minutes": 480
}' "201")

if [[ $? -eq 0 ]]; then
    ADMIN_ID=$(echo "$admin_response" | jq -r '.id // empty')
    log_success "Created admin user with ID: $ADMIN_ID"
fi

# Create customer user
customer_response=$(test_endpoint "Create customer user" "POST" "/api/v1/users" '{
    "entity_id": "'$ENTITY_ID'",
    "username": "customer1",
    "email": "customer@example.com",
    "password": "CustomerPass123!",
    "user_type": "CUSTOMER",
    "account_status": "ACTIVE"
}' "201")

if [[ $? -eq 0 ]]; then
    USER2_ID=$(echo "$customer_response" | jq -r '.id // empty')
    log_success "Created customer user with ID: $USER2_ID"
fi

echo

# Test 4: Authentication Tests
log_test "Authentication tests"

# Test successful authentication with email
test_endpoint "Authenticate with email" "POST" "/api/v1/users/auth" '{
    "identifier": "test1@example.com",
    "password": "TestPassword123!"
}' "200" > /dev/null

# Test successful authentication with username
test_endpoint "Authenticate with username" "POST" "/api/v1/users/auth" '{
    "identifier": "test.user1",
    "password": "TestPassword123!"
}' "200" > /dev/null

# Test failed authentication
test_endpoint "Authenticate with wrong password" "POST" "/api/v1/users/auth" '{
    "identifier": "test1@example.com",
    "password": "WrongPassword"
}' "401" > /dev/null

echo

# Test 5: User Retrieval Tests
log_test "User retrieval tests"

if [[ -n "$USER1_ID" ]]; then
    test_endpoint "Get user by ID" "GET" "/api/v1/users/$USER1_ID" "" "200" > /dev/null
fi

# Test non-existent user
test_endpoint "Get non-existent user" "GET" "/api/v1/users/550e8400-e29b-41d4-a716-446655440999" "" "404" > /dev/null

echo

# Test 6: List Users Tests
log_test "List users tests"

test_endpoint "List all users" "GET" "/api/v1/users" "" "200" > /dev/null

test_endpoint "List users with pagination" "GET" "/api/v1/users?limit=2&offset=0" "" "200" > /dev/null

test_endpoint "Filter by user type" "GET" "/api/v1/users?user_type=INTERNAL" "" "200" > /dev/null

test_endpoint "Filter by account status" "GET" "/api/v1/users?account_status=ACTIVE" "" "200" > /dev/null

echo

# Test 7: Update User Tests
log_test "Update user tests"

if [[ -n "$USER1_ID" ]]; then
    test_endpoint "Update user information" "PUT" "/api/v1/users/$USER1_ID" '{
        "username": "test.user1.updated",
        "email": "test1.updated@example.com",
        "mfa_enabled": true,
        "session_timeout_minutes": 240
    }' "200" > /dev/null
    
    # Test update with invalid data
    test_endpoint "Update with invalid email" "PUT" "/api/v1/users/$USER1_ID" '{
        "email": "invalid-email"
    }' "400" > /dev/null
fi

echo

# Test 8: Search Users Tests
log_test "Search users tests"

test_endpoint "Search users by query" "GET" "/api/v1/users/search?q=test" "" "200" > /dev/null

test_endpoint "Search with pagination" "GET" "/api/v1/users/search?q=user&limit=5&offset=0" "" "200" > /dev/null

# Test search without query
test_endpoint "Search without query parameter" "GET" "/api/v1/users/search" "" "400" > /dev/null

echo

# Test 9: Password Update Tests
log_test "Password update tests"

if [[ -n "$USER1_ID" ]]; then
    test_endpoint "Update password" "PUT" "/api/v1/users/$USER1_ID/password" '{
        "current_password": "TestPassword123!",
        "new_password": "NewTestPassword456!"
    }' "200" > /dev/null
    
    # Test with wrong current password
    test_endpoint "Update password with wrong current password" "PUT" "/api/v1/users/$USER1_ID/password" '{
        "current_password": "WrongPassword",
        "new_password": "AnotherPassword789!"
    }' "401" > /dev/null
fi

echo

# Test 10: User Roles Tests
log_test "User roles tests"

if [[ -n "$USER1_ID" ]]; then
    test_endpoint "Get user roles" "GET" "/api/v1/users/$USER1_ID/roles" "" "200" > /dev/null
fi

echo

# Test 11: Validation Error Tests
log_test "Validation error tests"

# Test create user with missing required fields
test_endpoint "Create user with missing fields" "POST" "/api/v1/users" '{
    "username": "incomplete"
}' "400" > /dev/null

# Test create user with invalid UUID
test_endpoint "Create user with invalid entity_id" "POST" "/api/v1/users" '{
    "entity_id": "invalid-uuid",
    "username": "test",
    "email": "test@example.com",
    "password": "password123",
    "user_type": "INTERNAL"
}' "400" > /dev/null

# Test duplicate email
test_endpoint "Create user with duplicate email" "POST" "/api/v1/users" '{
    "entity_id": "'$ENTITY_ID'",
    "username": "another.user",
    "email": "test1.updated@example.com",
    "password": "AnotherPassword123!",
    "user_type": "INTERNAL"
}' "409" > /dev/null

echo

# Test 12: Delete User Tests
log_test "Delete user tests"

if [[ -n "$USER2_ID" ]]; then
    test_endpoint "Soft delete user" "DELETE" "/api/v1/users/$USER2_ID" "" "200" > /dev/null
    
    # Verify user is deleted
    test_endpoint "Get deleted user (should fail)" "GET" "/api/v1/users/$USER2_ID" "" "404" > /dev/null
fi

echo

# Test 13: Performance Tests
log_test "Performance tests"

if [[ -n "$USER1_ID" ]]; then
    log_info "Testing cache performance (first request - cache miss)"
    time test_endpoint "Get user (cache miss)" "GET" "/api/v1/users/$USER1_ID" "" "200" > /dev/null 2>&1 || true
    
    log_info "Testing cache performance (second request - cache hit)"
    time test_endpoint "Get user (cache hit)" "GET" "/api/v1/users/$USER1_ID" "" "200" > /dev/null 2>&1 || true
fi

echo

# Cleanup
if [[ "$CLEANUP" == "true" ]]; then
    log_test "Cleaning up test data"
    
    # Delete remaining test users
    if [[ -n "$USER1_ID" ]]; then
        test_endpoint "Delete test user 1" "DELETE" "/api/v1/users/$USER1_ID" "" "200" > /dev/null || true
    fi
    
    if [[ -n "$ADMIN_ID" ]]; then
        test_endpoint "Delete admin user" "DELETE" "/api/v1/users/$ADMIN_ID" "" "200" > /dev/null || true
    fi
    
    # Delete test entity
    if [[ -n "$ENTITY_ID" ]]; then
        test_endpoint "Delete test entity" "DELETE" "/api/v1/entities/$ENTITY_ID" "" "200" > /dev/null || true
    fi
    
    log_success "Cleanup completed"
else
    log_warning "Cleanup skipped"
    if [[ -n "$ENTITY_ID" ]]; then
        echo "Test entity ID: $ENTITY_ID"
    fi
    if [[ -n "$USER1_ID" ]]; then
        echo "Test user 1 ID: $USER1_ID"
    fi
    if [[ -n "$ADMIN_ID" ]]; then
        echo "Admin user ID: $ADMIN_ID"
    fi
fi

echo
log_success " User API test suite completed successfully!"
echo
echo "Summary:"
echo "- ✅ Health check passed"
echo "- ✅ User creation and management tested"
echo "- ✅ Authentication flows tested"
echo "- ✅ CRUD operations tested"
echo "- ✅ Search functionality tested"
echo "- ✅ Password management tested"
echo "- ✅ Role management tested"
echo "- ✅ Error handling tested"
echo "- ✅ Validation testing completed"
echo "- ✅ Performance testing completed"

if [[ "$CLEANUP" == "true" ]]; then
    echo "- ✅ Test data cleaned up"
else
    echo "- ⚠️  Test data preserved (use --no-cleanup to preserve)"
fi

echo
echo " For detailed testing guide, see: user-api-tests.md"
echo " To run with verbose output: $0 --verbose"
echo " To preserve test data: $0 --no-cleanup"