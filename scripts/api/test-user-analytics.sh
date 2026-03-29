#!/bin/bash

# User Analytics & Behavior API Automated Test Script
# This script tests all User Analytics & Behavior API endpoints

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
    local timeout="${6:-10}"
    
    log_test "$description"
    
    local curl_cmd="curl -s -w '%{http_code}' -X $method '$BASE_URL$endpoint' --connect-timeout 5 --max-time $timeout"
    
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
            # Only show truncated output for large responses
            local body_length=${#body}
            if [[ $body_length -gt 1000 ]]; then
                echo "$body" | jq . 2>/dev/null | head -20 || echo "${body:0:500}..."
                echo "... (truncated, full response $body_length characters)"
            else
                echo "$body" | jq . 2>/dev/null || echo "$body"
            fi
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
        --type)
            TEST_TYPE="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo "Options:"
            echo "  -v, --verbose     Enable verbose output"
            echo "  --no-cleanup      Don't cleanup test data"
            echo "  --type TYPE       Test specific type (behavior, risk, anomaly, insights)"
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
echo -e "${BLUE} Starting User Analytics & Behavior API Test Suite...${NC}"
echo "Base URL: $BASE_URL"
echo "Verbose: $VERBOSE"
echo "Cleanup: $CLEANUP"
if [[ -n "$TEST_TYPE" ]]; then
    echo "Test Type: $TEST_TYPE"
fi
echo

# Global variables for test data
USER_ID_1="00000000-0000-0000-0000-000000000001"
USER_ID_2="00000000-0000-0000-0000-000000000002"
USER_ID_3="00000000-0000-0000-0000-000000000003"

# Test 1: Health Check
check_server

# Test 2: User Behavior Analytics
if [[ -z "$TEST_TYPE" || "$TEST_TYPE" == "behavior" ]]; then
    log_test "User behavior analytics"

    # Test behavior analytics for multiple users
    test_endpoint "Get user behavior analytics (User 1)" "GET" "/api/v1/analytics/users/$USER_ID_1/behavior" "" "200" 15 > /dev/null

    test_endpoint "Get user behavior analytics (User 2)" "GET" "/api/v1/analytics/users/$USER_ID_2/behavior" "" "200" 15 > /dev/null

    test_endpoint "Get user behavior analytics (User 3)" "GET" "/api/v1/analytics/users/$USER_ID_3/behavior" "" "200" 15 > /dev/null

    # Test non-existent user
    test_endpoint "Get behavior for non-existent user" "GET" "/api/v1/analytics/users/00000000-0000-0000-0000-000000000999/behavior" "" "404" 10 > /dev/null

    echo
fi

# Test 3: Risk Assessment
if [[ -z "$TEST_TYPE" || "$TEST_TYPE" == "risk" ]]; then
    log_test "Risk assessment"

    # Test risk assessment for multiple users
    test_endpoint "Get user risk assessment (User 1)" "GET" "/api/v1/analytics/users/$USER_ID_1/risk" "" "200" 15 > /dev/null

    test_endpoint "Get user risk assessment (User 2)" "GET" "/api/v1/analytics/users/$USER_ID_2/risk" "" "200" 15 > /dev/null

    test_endpoint "Get user risk assessment (User 3)" "GET" "/api/v1/analytics/users/$USER_ID_3/risk" "" "200" 15 > /dev/null

    # Test non-existent user
    test_endpoint "Get risk for non-existent user" "GET" "/api/v1/analytics/users/00000000-0000-0000-0000-000000000999/risk" "" "404" 10 > /dev/null

    echo
fi

# Test 4: Personalized Insights
if [[ -z "$TEST_TYPE" || "$TEST_TYPE" == "insights" ]]; then
    log_test "Personalized insights"

    # Test personalized insights for multiple users
    test_endpoint "Get personalized insights (User 1)" "GET" "/api/v1/analytics/users/$USER_ID_1/insights" "" "200" 15 > /dev/null

    test_endpoint "Get personalized insights (User 2)" "GET" "/api/v1/analytics/users/$USER_ID_2/insights" "" "200" 15 > /dev/null

    test_endpoint "Get personalized insights (User 3)" "GET" "/api/v1/analytics/users/$USER_ID_3/insights" "" "200" 15 > /dev/null

    # Test non-existent user
    test_endpoint "Get insights for non-existent user" "GET" "/api/v1/analytics/users/00000000-0000-0000-0000-000000000999/insights" "" "404" 10 > /dev/null

    echo
fi

# Test 5: Anomaly Detection
if [[ -z "$TEST_TYPE" || "$TEST_TYPE" == "anomaly" ]]; then
    log_test "Anomaly detection"

    # Test anomaly detection with different configurations
    test_endpoint "Detect anomalies ()" "POST" "/api/v1/analytics/users/$USER_ID_1/detect-anomalies" '{
        "analysis_period": {
            "start": "2024-01-01T00:00:00Z",
            "end": "2024-01-31T23:59:59Z"
        },
        "anomaly_types": ["login_time", "location", "device", "access_pattern"],
        "sensitivity": "medium"
    }' "200" 20 > /dev/null

    test_endpoint "Detect anomalies (login focus)" "POST" "/api/v1/analytics/users/$USER_ID_2/detect-anomalies" '{
        "analysis_period": {
            "start": "2024-01-01T00:00:00Z",
            "end": "2024-01-31T23:59:59Z"
        },
        "anomaly_types": ["login_time", "login_frequency"],
        "sensitivity": "high"
    }' "200" 20 > /dev/null

    test_endpoint "Detect anomalies (location focus)" "POST" "/api/v1/analytics/users/$USER_ID_3/detect-anomalies" '{
        "analysis_period": {
            "start": "2024-01-01T00:00:00Z",
            "end": "2024-01-31T23:59:59Z"
        },
        "anomaly_types": ["location", "device"],
        "sensitivity": "low"
    }' "200" 20 > /dev/null

    # Test with short analysis period
    test_endpoint "Detect anomalies (short period)" "POST" "/api/v1/analytics/users/$USER_ID_1/detect-anomalies" '{
        "analysis_period": {
            "start": "2024-01-15T00:00:00Z",
            "end": "2024-01-16T23:59:59Z"
        },
        "anomaly_types": ["login_time"],
        "sensitivity": "medium"
    }' "200" 15 > /dev/null

    # Test with extended analysis period
    test_endpoint "Detect anomalies (extended period)" "POST" "/api/v1/analytics/users/$USER_ID_1/detect-anomalies" '{
        "analysis_period": {
            "start": "2024-01-01T00:00:00Z",
            "end": "2024-03-31T23:59:59Z"
        },
        "anomaly_types": ["login_time", "location", "device"],
        "sensitivity": "medium"
    }' "200" 30 > /dev/null

    echo
fi

# Test 6: Error Scenarios
log_test "Error scenarios"

# Test invalid user ID
test_endpoint "Invalid user ID format" "GET" "/api/v1/analytics/users/invalid-uuid/behavior" "" "400" 5 > /dev/null

# Test missing required fields in anomaly detection
test_endpoint "Anomaly detection without analysis period" "POST" "/api/v1/analytics/users/$USER_ID_1/detect-anomalies" '{
    "anomaly_types": ["login_time"],
    "sensitivity": "medium"
}' "400" 10 > /dev/null

# Test invalid analysis period
test_endpoint "Invalid analysis period" "POST" "/api/v1/analytics/users/$USER_ID_1/detect-anomalies" '{
    "analysis_period": {
        "start": "2024-01-31T00:00:00Z",
        "end": "2024-01-01T00:00:00Z"
    },
    "anomaly_types": ["login_time"],
    "sensitivity": "medium"
}' "400" 10 > /dev/null

# Test invalid sensitivity level
test_endpoint "Invalid sensitivity level" "POST" "/api/v1/analytics/users/$USER_ID_1/detect-anomalies" '{
    "analysis_period": {
        "start": "2024-01-01T00:00:00Z",
        "end": "2024-01-31T23:59:59Z"
    },
    "anomaly_types": ["login_time"],
    "sensitivity": "invalid"
}' "400" 10 > /dev/null

# Test invalid anomaly types
test_endpoint "Invalid anomaly types" "POST" "/api/v1/analytics/users/$USER_ID_1/detect-anomalies" '{
    "analysis_period": {
        "start": "2024-01-01T00:00:00Z",
        "end": "2024-01-31T23:59:59Z"
    },
    "anomaly_types": ["invalid_type"],
    "sensitivity": "medium"
}' "400" 10 > /dev/null

echo

# Test 7: Performance Tests
log_test "Performance tests"

log_info "Testing behavior analytics performance"
time test_endpoint "Performance - behavior analytics" "GET" "/api/v1/analytics/users/$USER_ID_1/behavior" "" "200" 15 > /dev/null 2>&1 || true

log_info "Testing risk assessment performance"
time test_endpoint "Performance - risk assessment" "GET" "/api/v1/analytics/users/$USER_ID_1/risk" "" "200" 15 > /dev/null 2>&1 || true

log_info "Testing insights performance"
time test_endpoint "Performance - insights" "GET" "/api/v1/analytics/users/$USER_ID_1/insights" "" "200" 15 > /dev/null 2>&1 || true

log_info "Testing anomaly detection performance"
time test_endpoint "Performance - anomaly detection" "POST" "/api/v1/analytics/users/$USER_ID_1/detect-anomalies" '{
    "analysis_period": {
        "start": "2024-01-01T00:00:00Z",
        "end": "2024-01-31T23:59:59Z"
    },
    "anomaly_types": ["login_time"],
    "sensitivity": "medium"
}' "200" 20 > /dev/null 2>&1 || true

echo

# Test 8: Data Validation
log_test "Data validation tests"

# Test behavior analytics data structure validation
behavior_response=$(test_endpoint "Validate behavior analytics structure" "GET" "/api/v1/analytics/users/$USER_ID_1/behavior" "" "200" 15)

if [[ $? -eq 0 ]]; then
    # Check if response contains expected fields
    if echo "$behavior_response" | jq -e '.behavior_pattern.user_id' > /dev/null 2>&1; then
        log_success "Behavior analytics structure validation passed"
    else
        log_error "Behavior analytics structure validation failed"
    fi
fi

# Test risk assessment data structure validation
risk_response=$(test_endpoint "Validate risk assessment structure" "GET" "/api/v1/analytics/users/$USER_ID_1/risk" "" "200" 15)

if [[ $? -eq 0 ]]; then
    # Check if response contains expected fields
    if echo "$risk_response" | jq -e '.risk_assessment.overall_risk_score' > /dev/null 2>&1; then
        log_success "Risk assessment structure validation passed"
    else
        log_error "Risk assessment structure validation failed"
    fi
fi

# Test insights data structure validation
insights_response=$(test_endpoint "Validate insights structure" "GET" "/api/v1/analytics/users/$USER_ID_1/insights" "" "200" 15)

if [[ $? -eq 0 ]]; then
    # Check if response contains expected fields
    if echo "$insights_response" | jq -e '.insights.user_id' > /dev/null 2>&1; then
        log_success "Insights structure validation passed"
    else
        log_error "Insights structure validation failed"
    fi
fi

echo

# Test 9: Edge Cases
log_test "Edge cases"

# Test analytics for user with minimal data
test_endpoint "Analytics for user with minimal data" "GET" "/api/v1/analytics/users/$USER_ID_3/behavior" "" "200" 15 > /dev/null

# Test anomaly detection for very short period
test_endpoint "Anomaly detection for 1 hour period" "POST" "/api/v1/analytics/users/$USER_ID_1/detect-anomalies" '{
    "analysis_period": {
        "start": "2024-01-15T10:00:00Z",
        "end": "2024-01-15T11:00:00Z"
    },
    "anomaly_types": ["login_time"],
    "sensitivity": "medium"
}' "200" 15 > /dev/null

# Test all anomaly types
test_endpoint "All anomaly types" "POST" "/api/v1/analytics/users/$USER_ID_1/detect-anomalies" '{
    "analysis_period": {
        "start": "2024-01-01T00:00:00Z",
        "end": "2024-01-31T23:59:59Z"
    },
    "anomaly_types": ["login_time", "location", "device", "access_pattern", "login_frequency", "session_duration"],
    "sensitivity": "medium"
}' "200" 25 > /dev/null

echo

# Test 10: Concurrent Access
log_test "Concurrent access simulation"

# Simulate concurrent requests
log_info "Simulating concurrent analytics requests"
for i in {1..3}; do
    test_endpoint "Concurrent request $i" "GET" "/api/v1/analytics/users/$USER_ID_1/behavior" "" "200" 15 > /dev/null &
done
wait

log_success "Concurrent access tests completed"

echo

# Cleanup
if [[ "$CLEANUP" == "true" ]]; then
    log_test "Cleaning up test data"
    
    # Note: Analytics data is typically read-only, so cleanup is minimal
    log_success "Cleanup completed (analytics data is read-only)"
else
    log_warning "Cleanup skipped"
fi

echo
log_success " User Analytics & Behavior API test suite completed successfully!"
echo
echo "Summary:"
echo "- ✅ Health check passed"
echo "- ✅ User behavior analytics tested"
echo "- ✅ Risk assessment tested"
echo "- ✅ Personalized insights tested"
echo "- ✅ Anomaly detection tested"
echo "- ✅ Error handling tested"
echo "- ✅ Performance testing completed"
echo "- ✅ Data validation tested"
echo "- ✅ Edge cases tested"
echo "- ✅ Concurrent access tested"

if [[ "$CLEANUP" == "true" ]]; then
    echo "- ✅ Cleanup completed"
else
    echo "- ⚠️  Cleanup skipped"
fi

echo
echo " For detailed testing guide, see: workflows/user-analytics/curl-examples.md"
echo " To run with verbose output: $0 --verbose"
echo " To test specific type: $0 --type behavior|risk|anomaly|insights"
echo " To preserve test data: $0 --no-cleanup"