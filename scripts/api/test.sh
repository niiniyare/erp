#!/bin/bash

# Main Test Entry Point for ERP API Testing
# This script provides organized access to all API test scripts

set -e

SCRIPTS_DIR="$(cd "$(dirname "$0")" && pwd)"

# Source environment configuration if it exists
if [ -f "$SCRIPTS_DIR/setup_env.sh" ]; then
    source "$SCRIPTS_DIR/setup_env.sh"
fi

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# --- Test Runner Functions ---

run_test_script() {
    local script_name="$1"
    local test_title="$2"
    
    echo -e "${BLUE}▶️  Running $test_title${NC}"
    echo -e "${BLUE}================================$(echo "$test_title" | sed 's/./=/g')${NC}"
    
    local script_path="$SCRIPTS_DIR/$script_name"
    if [ ! -f "$script_path" ]; then
        echo -e "${RED}❌ Test script not found: $script_path${NC}"
        return 1
    fi
    
    chmod +x "$script_path"
    if "$script_path"; then
        echo -e "${GREEN}✅ $test_title PASSED${NC}"
        return 0
    else
        echo -e "${RED}❌ $test_title FAILED${NC}"
        return 1
    fi
}

run_all_tests() {
    echo -e "${BLUE}🚀 Running All API Test Suites${NC}"
    local all_passed=true
    
    run_test_script "test_health_api.sh" "Health Check Tests" || all_passed=false
    run_test_script "test_middleware_api.sh" "Middleware Integration Tests" || all_passed=false
    # Assuming test_tenant_api.sh exists and is structured similarly
    if [ -f "$SCRIPTS_DIR/test_tenant_api.sh" ]; then
        run_test_script "test_tenant_api.sh" "Tenant API Tests" || all_passed=false
    fi
    run_test_script "test_organization_api.sh" "Organization API Tests" || all_passed=false
    run_test_script "test_finance_api.sh" "Finance API Tests" || all_passed=false
    
    echo ""
    if $all_passed;
    then
        echo -e "${GREEN}🎉 All test suites passed successfully!${NC}"
        exit 0
    else
        echo -e "${RED}🔥 One or more test suites failed!${NC}"
        exit 1
    fi
}

# --- Interactive Menu ---

show_menu() {
    echo -e "\n${BLUE}🧪 ERP API Test Suite${NC}"
    echo -e "${BLUE}======================${NC}"
    echo -e "\n${YELLOW}Usage:${NC}"
    echo "  ./test.sh [command]"
    echo -e "\n${YELLOW}Commands:${NC}"
    echo "  all          - Run all test suites"
    echo "  tenant       - Run Tenant API tests"
    echo "  organization - Run Organization API tests"
    echo "  finance      - Run Finance API tests"
    echo "  middleware   - Run Middleware integration tests"
    echo "  health       - Run Health Check tests"
    echo "  (no command) - Show this interactive menu"
    echo ""
}

# --- Main Logic ---

main() {
    # Non-interactive mode
    if [ -n "$1" ]; then
        case $1 in
            all)
                run_all_tests
                ;; 
            tenant)
                run_test_script "test_tenant_api.sh" "Tenant API Tests"
                ;; 
            organization)
                run_test_script "test_organization_api.sh" "Organization API Tests"
                ;; 
            finance)
                run_test_script "test_finance_api.sh" "Finance API Tests"
                ;; 
            middleware)
                run_test_script "test_middleware_api.sh" "Middleware Integration Tests"
                ;; 
            health)
                run_test_script "test_health_api.sh" "Health Check Tests"
                ;; 
            *)
                echo -e "${RED}❌ Invalid command: $1${NC}"
                show_menu
                exit 1
                ;; 
        esac
        exit $?
    fi

    # Interactive mode
    while true; do
        show_menu
        echo -n "Select a command to run (or type 'exit'): "
        read -r choice
        echo ""
        
        if [[ "$choice" == "exit" || "$choice" == "0" ]]; then
            echo -e "${GREEN}👋 Goodbye!${NC}"
            exit 0
        fi
        
        # This re-invokes the script with the choice as a command-line argument
        # which is a clean way to handle both modes.
        bash "$0" "$choice"
        local last_status=$?

        # Stop interactive mode if a test fails
        if [ $last_status -ne 0 ]; then
            exit $last_status
        fi
        
        echo -e "\n${YELLOW}Press Enter to return to the menu...${NC}"
        read -r
    done
}

main "$@"
