#!/bin/bash

#  API Test Runner
# This script runs all API tests in the correct order

set -e  # Exit on any error

# Configuration
BASE_URL="http://localhost:8080"
VERBOSE=false
CLEANUP=true
SKIP_LONG_TESTS=false
PARALLEL_TESTS=false

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
    echo -e "${PURPLE}🚀 $1${NC}"
}

log_section() {
    echo -e "${CYAN}📋 $1${NC}"
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
        log_success "Server is running at $BASE_URL"
        return 0
    else
        log_error "Server is not running at $BASE_URL"
        log_info "Please start the server first:"
        echo "  go run cmd/server/main.go"
        exit 1
    fi
}

# Function to run a test script
run_test_script() {
    local script_name="$1"
    local description="$2"
    local script_path="./utilities/scripts/$script_name"
    
    log_section "Running $description"
    
    if [[ ! -f "$script_path" ]]; then
        log_error "Test script not found: $script_path"
        return 1
    fi
    
    if [[ ! -x "$script_path" ]]; then
        log_info "Making script executable: $script_path"
        chmod +x "$script_path"
    fi
    
    local test_args=""
    if [[ "$VERBOSE" == "true" ]]; then
        test_args="$test_args --verbose"
    fi
    if [[ "$CLEANUP" == "false" ]]; then
        test_args="$test_args --no-cleanup"
    fi
    
    echo "----------------------------------------"
    if $script_path $test_args; then
        log_success "$description completed successfully"
        return 0
    else
        log_error "$description failed"
        return 1
    fi
}

# Function to run tests in parallel
run_parallel_tests() {
    local pids=()
    local results=()
    
    log_section "Running Core API Tests in Parallel"
    
    # Start core API tests in parallel
    ./utilities/scripts/test-entities.sh $test_args > /tmp/test-entities.log 2>&1 &
    pids+=($!)
    
    ./utilities/scripts/test-users.sh $test_args > /tmp/test-users.log 2>&1 &
    pids+=($!)
    
    # Wait for core tests to complete
    for i in "${!pids[@]}"; do
        if wait "${pids[$i]}"; then
            results+=("success")
        else
            results+=("failure")
        fi
    done
    
    # Show results
    local tests=("Entities API" "Users API")
    for i in "${!results[@]}"; do
        if [[ "${results[$i]}" == "success" ]]; then
            log_success "${tests[$i]} tests completed"
        else
            log_error "${tests[$i]} tests failed"
            echo "See /tmp/test-${tests[$i],,}.log for details"
        fi
    done
    
    # Run workflow tests sequentially (they may depend on core APIs)
    log_section "Running Workflow API Tests Sequentially"
    
    run_test_script "test-access-requests.sh" "Access Request Workflow API Tests"
    run_test_script "test-conditional-access.sh" "Conditional Access Controls API Tests"
    run_test_script "test-user-analytics.sh" "User Analytics & Behavior API Tests"
}

# Function to run tests sequentially
run_sequential_tests() {
    log_section "Running All API Tests Sequentially"
    
    # Core API tests
    run_test_script "test-entities.sh" "Entity Management API Tests"
    run_test_script "test-users.sh" "User Management API Tests"
    
    # Workflow API tests
    run_test_script "test-access-requests.sh" "Access Request Workflow API Tests"
    run_test_script "test-conditional-access.sh" "Conditional Access Controls API Tests"
    run_test_script "test-user-analytics.sh" "User Analytics & Behavior API Tests"
}

# Function to run specific test category
run_category_tests() {
    local category="$1"
    
    case "$category" in
        "core")
            log_section "Running Core API Tests"
            run_test_script "test-entities.sh" "Entity Management API Tests"
            run_test_script "test-users.sh" "User Management API Tests"
            ;;
        "workflow")
            log_section "Running Workflow API Tests"
            run_test_script "test-access-requests.sh" "Access Request Workflow API Tests"
            run_test_script "test-conditional-access.sh" "Conditional Access Controls API Tests"
            run_test_script "test-user-analytics.sh" "User Analytics & Behavior API Tests"
            ;;
        "entities")
            run_test_script "test-entities.sh" "Entity Management API Tests"
            ;;
        "users")
            run_test_script "test-users.sh" "User Management API Tests"
            ;;
        "access-requests")
            run_test_script "test-access-requests.sh" "Access Request Workflow API Tests"
            ;;
        "conditional-access")
            run_test_script "test-conditional-access.sh" "Conditional Access Controls API Tests"
            ;;
        "user-analytics")
            run_test_script "test-user-analytics.sh" "User Analytics & Behavior API Tests"
            ;;
        *)
            log_error "Unknown test category: $category"
            echo "Available categories: core, workflow, entities, users, access-requests, conditional-access, user-analytics"
            exit 1
            ;;
    esac
}

# Function to show test summary
show_test_summary() {
    echo
    log_header "Test Summary"
    echo "----------------------------------------"
    echo "📊 Test Results Summary:"
    echo "  • Server Status: ✅ Running"
    echo "  • Test Mode: $(if [[ "$PARALLEL_TESTS" == "true" ]]; then echo "Parallel"; else echo "Sequential"; fi)"
    echo "  • Verbose Mode: $(if [[ "$VERBOSE" == "true" ]]; then echo "Enabled"; else echo "Disabled"; fi)"
    echo "  • Cleanup Mode: $(if [[ "$CLEANUP" == "true" ]]; then echo "Enabled"; else echo "Disabled"; fi)"
    echo "  • Base URL: $BASE_URL"
    echo
    echo "🔍 Available Documentation:"
    echo "  • Core APIs: docs/api/core-apis/"
    echo "  • Workflow APIs: docs/api/workflows/"
    echo "  • Test Scripts: docs/api/utilities/scripts/"
    echo "  • Manual Tests: docs/api/testing/manual/"
    echo
    echo "🚀 Quick Commands:"
    echo "  • Test specific category: $0 --category core"
    echo "  • Verbose output: $0 --verbose"
    echo "  • Parallel execution: $0 --parallel"
    echo "  • Preserve test data: $0 --no-cleanup"
    echo
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
        --skip-long)
            SKIP_LONG_TESTS=true
            shift
            ;;
        --parallel)
            PARALLEL_TESTS=true
            shift
            ;;
        --category)
            CATEGORY="$2"
            shift 2
            ;;
        --base-url)
            BASE_URL="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo
            echo "Options:"
            echo "  -v, --verbose         Enable verbose output"
            echo "  --no-cleanup          Don't cleanup test data"
            echo "  --skip-long           Skip long-running tests"
            echo "  --parallel            Run tests in parallel where possible"
            echo "  --category CATEGORY   Run specific test category"
            echo "  --base-url URL        Override base URL (default: http://localhost:8080)"
            echo "  -h, --help            Show this help message"
            echo
            echo "Categories:"
            echo "  core                  Core API tests (entities, users)"
            echo "  workflow              Workflow API tests (access-requests, conditional-access, user-analytics)"
            echo "  entities              Entity Management API tests"
            echo "  users                 User Management API tests"
            echo "  access-requests       Access Request Workflow API tests"
            echo "  conditional-access    Conditional Access Controls API tests"
            echo "  user-analytics        User Analytics & Behavior API tests"
            echo
            echo "Examples:"
            echo "  $0                              # Run all tests"
            echo "  $0 --verbose                    # Run with verbose output"
            echo "  $0 --parallel                   # Run tests in parallel"
            echo "  $0 --category core              # Run only core API tests"
            echo "  $0 --category user-analytics    # Run only user analytics tests"
            echo "  $0 --no-cleanup --verbose       # Verbose mode with test data preservation"
            echo
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Start testing
log_header "AWO ERP System API Test Suite"
echo "Starting API testing..."
echo "Base URL: $BASE_URL"
echo "Verbose: $VERBOSE"
echo "Cleanup: $CLEANUP"
echo "Skip Long Tests: $SKIP_LONG_TESTS"
echo "Parallel Tests: $PARALLEL_TESTS"
if [[ -n "$CATEGORY" ]]; then
    echo "Category: $CATEGORY"
fi
echo

# Check server health
check_server

# Build test arguments
test_args=""
if [[ "$VERBOSE" == "true" ]]; then
    test_args="$test_args --verbose"
fi
if [[ "$CLEANUP" == "false" ]]; then
    test_args="$test_args --no-cleanup"
fi

# Run tests based on mode
if [[ -n "$CATEGORY" ]]; then
    # Run specific category
    run_category_tests "$CATEGORY"
elif [[ "$PARALLEL_TESTS" == "true" ]]; then
    # Run tests in parallel
    run_parallel_tests
else
    # Run tests sequentially
    run_sequential_tests
fi

# Show final summary
show_test_summary

log_success "🎉 All API tests completed successfully!"
echo
echo "📈 Next Steps:"
echo "  • Review test results above"
echo "  • Check server logs for any issues"
echo "  • Run specific tests if needed: $0 --category <category>"
echo "  • Use verbose mode for debugging: $0 --verbose"
echo "  • Consider running performance tests separately"
echo
echo "🔗 Resources:"
echo "  • API Documentation: docs/api/"
echo "  • Test Scripts: docs/api/utilities/scripts/"
echo "  • Manual Testing: docs/api/testing/manual/"
echo "  • Troubleshooting: docs/api/README.md"