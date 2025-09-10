#!/bin/bash

# Main Test Entry Point for ERP API Testing
# This script provides organized access to all API test scripts

set -e

SCRIPTS_DIR="$(dirname "$0")"
PROJECT_ROOT="$(dirname "$(dirname "$(dirname "$SCRIPTS_DIR")")")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🧪 ERP API Test Suite${NC}"
echo -e "${BLUE}======================${NC}"
echo ""

# Function to show available tests
show_menu() {
    echo -e "${YELLOW}Available Test Categories:${NC}"
    echo ""
    echo "1) 🏢 Tenant API Tests"
    echo "   - Test tenant CRUD operations"
    echo "   - Validate tenant middleware (header/subdomain extraction)"
    echo "   - Test public endpoint bypass"
    echo ""
    echo "2) 🏗️  Entity/Organization API Tests (Coming Soon)"
    echo "   - Test entity management"
    echo "   - Validate hierarchical relationships"
    echo ""
    echo "3) 🔐 Authentication API Tests (Coming Soon)"
    echo "   - Test login/logout functionality"
    echo "   - Validate JWT token handling"
    echo ""
    echo "4) 🛡️  Middleware Integration Tests"
    echo "   - Test native middleware chain"
    echo "   - Validate tenant isolation"
    echo "   - Test error handling scenarios"
    echo ""
    echo "5) 🏥 Health Check Tests"
    echo "   - Test all service health endpoints"
    echo "   - Validate system readiness"
    echo ""
    echo "0) Exit"
    echo ""
}

# Function to check if server is running
check_server() {
    local base_url="http://localhost:8090"
    echo -e "${YELLOW}🔍 Checking if server is running...${NC}"
    
    if curl -s -f "$base_url/health" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Server is running at $base_url${NC}"
        return 0
    else
        echo -e "${RED}❌ Server is not running at $base_url${NC}"
        echo -e "${YELLOW}💡 Start the server with: go run ./cmd/server/${NC}"
        return 1
    fi
}

# Function to run tenant API tests
run_tenant_tests() {
    echo -e "${BLUE}🏢 Running Tenant API Tests${NC}"
    echo -e "${BLUE}===========================${NC}"
    
    if check_server; then
        echo ""
        chmod +x "$SCRIPTS_DIR/test_tenant_api.sh"
        "$SCRIPTS_DIR/test_tenant_api.sh"
    else
        echo -e "${RED}⚠️  Cannot run tests without server${NC}"
        return 1
    fi
}

# Function to run middleware tests
run_middleware_tests() {
    echo -e "${BLUE}🛡️  Running Middleware Tests${NC}"
    echo -e "${BLUE}============================${NC}"
    
    if check_server; then
        echo ""
        echo -e "${YELLOW}Testing native middleware chain...${NC}"
        
        # Test 1: Public endpoint (should work without tenant)
        echo "📋 Test: Health endpoint (public)"
        curl -s -w "Status: %{http_code}\n" \
            -H "Content-Type: application/json" \
            "http://localhost:8090/health" || echo "Failed"
        echo ""
        
        # Test 2: Protected endpoint without tenant (should fail)
        echo "📋 Test: Protected endpoint without tenant header"
        curl -s -w "Status: %{http_code}\n" \
            -H "Content-Type: application/json" \
            "http://localhost:8090/api/v1/tenants" || echo "Expected failure"
        echo ""
        
        # Test 3: Protected endpoint with invalid tenant ID
        echo "📋 Test: Protected endpoint with invalid tenant ID"
        curl -s -w "Status: %{http_code}\n" \
            -H "Content-Type: application/json" \
            -H "X-Tenant-ID: invalid-uuid" \
            "http://localhost:8090/api/v1/tenants" || echo "Expected validation error"
        echo ""
        
        echo -e "${GREEN}✅ Middleware tests completed${NC}"
    else
        return 1
    fi
}

# Function to run health check tests
run_health_tests() {
    echo -e "${BLUE}🏥 Running Health Check Tests${NC}"
    echo -e "${BLUE}=============================${NC}"
    
    local base_url="http://localhost:8090"
    
    echo "📋 Testing service health endpoints..."
    
    # Main health endpoint
    echo "- General health:"
    curl -s -w "  Status: %{http_code}\n" "$base_url/health" || echo "  Failed"
    
    # Tenant service health  
    echo "- Tenant service health:"
    curl -s -w "  Status: %{http_code}\n" "$base_url/api/v1/tenants/health" || echo "  Failed"
    
    # Feature flag service health
    echo "- Feature flag service health:"
    curl -s -w "  Status: %{http_code}\n" "$base_url/api/v1/feature-flags/health" || echo "  Failed"
    
    # ABAC service health
    echo "- ABAC service health:"
    curl -s -w "  Status: %{http_code}\n" "$base_url/api/v1/abac/health" || echo "  Failed"
    
    echo ""
    echo -e "${GREEN}✅ Health check tests completed${NC}"
}

# Function to show test results summary
show_summary() {
    echo ""
    echo -e "${BLUE}📊 Test Summary${NC}"
    echo -e "${BLUE}===============${NC}"
    echo "• Tenant API: GOA-converted, ready for testing"
    echo "• Native Middleware: Implemented and integrated"
    echo "• Public Endpoints: Properly whitelisted"
    echo "• Error Handling: JSON responses with proper HTTP codes"
    echo ""
    echo -e "${YELLOW}Next Steps:${NC}"
    echo "• Convert Entity/Organization APIs to GOA"
    echo "• Add more comprehensive test scenarios"
    echo "• Test multi-tenant isolation"
}

# Main script logic
main() {
    while true; do
        show_menu
        echo -n "Select test category (0-5): "
        read -r choice
        echo ""
        
        case $choice in
            1)
                run_tenant_tests
                ;;
            2)
                echo -e "${YELLOW}🚧 Entity/Organization API tests - Coming Soon${NC}"
                echo "This will test entity management once converted to GOA"
                ;;
            3)
                echo -e "${YELLOW}🚧 Authentication API tests - Coming Soon${NC}"
                echo "This will test login/logout functionality"
                ;;
            4)
                run_middleware_tests
                ;;
            5)
                run_health_tests
                ;;
            0)
                echo -e "${GREEN}👋 Goodbye!${NC}"
                exit 0
                ;;
            *)
                echo -e "${RED}❌ Invalid choice. Please select 0-5.${NC}"
                ;;
        esac
        
        echo ""
        echo -e "${YELLOW}Press Enter to continue...${NC}"
        read -r
        echo ""
    done
}

# Show summary if --summary flag is provided
if [[ "$1" == "--summary" ]]; then
    show_summary
    exit 0
fi

# Run main function
main