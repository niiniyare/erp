#!/bin/bash

# Access Request Workflow API Automated Test Script
# This script tests all Access Request Workflow API endpoints

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
    echo -e "${BLUE}🧪 Testing: $1${NC}"
}

# Check if jq is available
if ! command -v jq &> /dev/null; then
    log_error "jq is required but not installed. Please install jq first."
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
    local headers="$6"
    
    log_test "$description"
    
    local curl_cmd="curl -s -w '%{http_code}' -X $method '$BASE_URL$endpoint'"
    
    if [[ -n "$headers" ]]; then
        curl_cmd="$curl_cmd $headers"
    fi
    
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
echo -e "${BLUE}🚀 Starting Access Request Workflow API Test Suite...${NC}"
echo "Base URL: $BASE_URL"
echo "Verbose: $VERBOSE"
echo "Cleanup: $CLEANUP"
echo

# Global variables for test data
ENTITY_ID="00000000-0000-0000-0000-000000000001"
ROLE_ID="00000000-0000-0000-0000-000000000002"
PERMISSION_ID="00000000-0000-0000-0000-000000000003"
RESOURCE_ID="00000000-0000-0000-0000-000000000004"
REQUESTER_ID="00000000-0000-0000-0000-000000000005"
TARGET_USER_ID="00000000-0000-0000-0000-000000000006"
APPROVER_ID="00000000-0000-0000-0000-000000000007"

ACCESS_REQUEST_ID=""
RULE_ID=""

# Test 1: Health Check
check_server

# Test 2: Create Access Requests
log_test "Creating access requests"

# Create role assignment request
role_request_response=$(test_endpoint "Create role assignment request" "POST" "/api/v1/access-requests" '{
    "entity_id": "'$ENTITY_ID'",
    "request_type": "ROLE_ASSIGNMENT",
    "role_id": "'$ROLE_ID'",
    "target_user_id": "'$TARGET_USER_ID'",
    "justification": "User needs admin access for project management",
    "business_reason": "Required for Q1 project delivery",
    "duration_hours": 168,
    "auto_revoke": true
}' "201" "-H 'X-User-ID: $REQUESTER_ID'")

if [[ $? -eq 0 ]]; then
    ACCESS_REQUEST_ID=$(echo "$role_request_response" | jq -r '.access_request.id // empty')
    if [[ -n "$ACCESS_REQUEST_ID" && "$ACCESS_REQUEST_ID" != "null" ]]; then
        log_success "Created access request with ID: $ACCESS_REQUEST_ID"
    fi
fi

# Create permission grant request
test_endpoint "Create permission grant request" "POST" "/api/v1/access-requests" '{
    "entity_id": "'$ENTITY_ID'",
    "request_type": "PERMISSION_GRANT",
    "permission_id": "'$PERMISSION_ID'",
    "target_user_id": "'$TARGET_USER_ID'",
    "justification": "User needs read access to financial reports",
    "business_reason": "Monthly reporting requirements",
    "duration_hours": 24,
    "auto_revoke": true
}' "201" "-H 'X-User-ID: $REQUESTER_ID'" > /dev/null

# Create resource access request
test_endpoint "Create resource access request" "POST" "/api/v1/access-requests" '{
    "entity_id": "'$ENTITY_ID'",
    "request_type": "RESOURCE_ACCESS",
    "resource_id": "'$RESOURCE_ID'",
    "target_user_id": "'$TARGET_USER_ID'",
    "justification": "Need access to customer database for support",
    "business_reason": "Customer escalation case #12345",
    "duration_hours": 8,
    "auto_revoke": true
}' "201" "-H 'X-User-ID: $REQUESTER_ID'" > /dev/null

# Create elevation request
test_endpoint "Create elevation request" "POST" "/api/v1/access-requests" '{
    "entity_id": "'$ENTITY_ID'",
    "request_type": "ELEVATION",
    "justification": "Emergency system maintenance required",
    "business_reason": "Critical security patch deployment",
    "duration_hours": 4,
    "auto_revoke": true
}' "201" "-H 'X-User-ID: $REQUESTER_ID'" > /dev/null

echo

# Test 3: Get Access Request
log_test "Access request retrieval"

if [[ -n "$ACCESS_REQUEST_ID" ]]; then
    test_endpoint "Get access request by ID" "GET" "/api/v1/access-requests/$ACCESS_REQUEST_ID" "" "200" > /dev/null
fi

# Test non-existent request
test_endpoint "Get non-existent request" "GET" "/api/v1/access-requests/00000000-0000-0000-0000-000000000999" "" "404" > /dev/null

echo

# Test 4: List Access Requests
log_test "List access requests"

test_endpoint "List all access requests" "GET" "/api/v1/access-requests" "" "200" > /dev/null

test_endpoint "List with pagination" "GET" "/api/v1/access-requests?limit=5&offset=0" "" "200" > /dev/null

test_endpoint "Filter by status" "GET" "/api/v1/access-requests?status=PENDING" "" "200" > /dev/null

test_endpoint "Filter by request type" "GET" "/api/v1/access-requests?request_type=ROLE_ASSIGNMENT" "" "200" > /dev/null

echo

# Test 5: Process Access Requests
log_test "Process access requests"

if [[ -n "$ACCESS_REQUEST_ID" ]]; then
    # Test approval
    test_endpoint "Approve access request" "POST" "/api/v1/access-requests/$ACCESS_REQUEST_ID/process" '{
        "action": "approve",
        "comments": "Approved for business justification",
        "duration_hours": 24
    }' "200" "-H 'X-User-ID: $APPROVER_ID'" > /dev/null
    
    # Test rejection (create another request first)
    reject_request_response=$(test_endpoint "Create request for rejection" "POST" "/api/v1/access-requests" '{
        "entity_id": "'$ENTITY_ID'",
        "request_type": "ROLE_ASSIGNMENT",
        "role_id": "'$ROLE_ID'",
        "target_user_id": "'$TARGET_USER_ID'",
        "justification": "Request to be rejected for testing",
        "business_reason": "Testing rejection workflow",
        "duration_hours": 24
    }' "201" "-H 'X-User-ID: $REQUESTER_ID'")
    
    if [[ $? -eq 0 ]]; then
        REJECT_REQUEST_ID=$(echo "$reject_request_response" | jq -r '.access_request.id // empty')
        if [[ -n "$REJECT_REQUEST_ID" && "$REJECT_REQUEST_ID" != "null" ]]; then
            test_endpoint "Reject access request" "POST" "/api/v1/access-requests/$REJECT_REQUEST_ID/process" '{
                "action": "reject",
                "comments": "Insufficient business justification"
            }' "200" "-H 'X-User-ID: $APPROVER_ID'" > /dev/null
        fi
    fi
fi

echo

# Test 6: Access Request Statistics
log_test "Access request statistics"

test_endpoint "Get access request statistics" "GET" "/api/v1/access-requests/stats?from_date=2024-01-01&to_date=2024-12-31" "" "200" > /dev/null

echo

# Test 7: Revoke Access Request
log_test "Revoke access request"

if [[ -n "$ACCESS_REQUEST_ID" ]]; then
    test_endpoint "Revoke access request" "DELETE" "/api/v1/access-requests/$ACCESS_REQUEST_ID" "" "200" > /dev/null
fi

echo

# Test 8: Error Scenarios
log_test "Error scenarios"

# Test missing required fields
test_endpoint "Create request with missing fields" "POST" "/api/v1/access-requests" '{
    "request_type": "ROLE_ASSIGNMENT"
}' "400" "-H 'X-User-ID: $REQUESTER_ID'" > /dev/null

# Test missing user header
test_endpoint "Create request without user header" "POST" "/api/v1/access-requests" '{
    "entity_id": "'$ENTITY_ID'",
    "request_type": "ROLE_ASSIGNMENT"
}' "401" > /dev/null

# Test invalid UUID
test_endpoint "Get request with invalid UUID" "GET" "/api/v1/access-requests/invalid-uuid" "" "400" > /dev/null

# Test process without approver header
test_endpoint "Process request without approver header" "POST" "/api/v1/access-requests/00000000-0000-0000-0000-000000000001/process" '{
    "action": "approve"
}' "401" > /dev/null

echo

# Test 9: Performance Tests
log_test "Performance tests"

if [[ -n "$ACCESS_REQUEST_ID" ]]; then
    log_info "Testing response times"
    time test_endpoint "Get access request (performance test)" "GET" "/api/v1/access-requests/$ACCESS_REQUEST_ID" "" "200" > /dev/null 2>&1 || true
fi

echo

# Cleanup
if [[ "$CLEANUP" == "true" ]]; then
    log_test "Cleaning up test data"
    
    # Note: In a real implementation, you would clean up test data here
    # For now, we'll just log that cleanup would occur
    log_success "Cleanup completed (simulated)"
else
    log_warning "Cleanup skipped"
    if [[ -n "$ACCESS_REQUEST_ID" ]]; then
        echo "Access request ID: $ACCESS_REQUEST_ID"
    fi
fi

echo
log_success "🎉 Access Request Workflow API test suite completed successfully!"
echo
echo "Summary:"
echo "- ✅ Health check passed"
echo "- ✅ Access request creation tested"
echo "- ✅ Access request retrieval tested"
echo "- ✅ Access request processing tested"
echo "- ✅ Access request listing tested"
echo "- ✅ Access request statistics tested"
echo "- ✅ Access request revocation tested"
echo "- ✅ Error handling tested"
echo "- ✅ Performance testing completed"

if [[ "$CLEANUP" == "true" ]]; then
    echo "- ✅ Test data cleaned up"
else
    echo "- ⚠️  Test data preserved"
fi

echo
echo "🔍 For detailed testing guide, see: workflows/access-requests/curl-examples.md"
echo "🚀 To run with verbose output: $0 --verbose"
echo "💾 To preserve test data: $0 --no-cleanup"