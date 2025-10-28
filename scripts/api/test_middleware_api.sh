#!/bin/bash

# Middleware and Core API Test Script

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/api_client.sh"
source "$SCRIPT_DIR/lib/test_helpers.sh"

# --- Test Cases ---

test_middleware_and_health() {
    local tenant_id="$1"
    test::start "Middleware and Health Checks"

    # 1. Public endpoint (should work without tenant)
    local health_resp=$(api::get_public "/health")
    assert::http_status "public_health_endpoint" "$health_resp" "200"

    # 2. Protected endpoint without tenant header (should fail)
    # We need to call curl directly to NOT send the tenant header
    local no_tenant_resp=$(curl -s -w "%{http_code}" -H "Content-Type: application/json" "$SERVER_URL/api/v1/tenants")
    assert::http_status "protected_endpoint_no_tenant" "$no_tenant_resp" "400" # Or 401/403 depending on impl.

    # 3. Protected endpoint with invalid tenant ID
    local invalid_tenant_resp=$(api::get "/api/v1/tenants" "invalid-uuid")
    assert::http_status "protected_endpoint_invalid_tenant" "$invalid_tenant_resp" "400" # Bad request

    # 4. Service-specific health endpoints (should be public)
    local tenant_health=$(api::get_public "/api/v1/tenants/health")
    assert::http_status "tenant_service_health" "$tenant_health" "200"
    
    local finance_health=$(api::get_public "/api/v1/finance/health")
    assert::http_status "finance_service_health" "$finance_health" "200"
}

# --- Main Execution ---
main() {
    test::on_exit cleanup

    test::log_info "Setting up test environment..."
    local tenant_id=$(setup::create_tenant)
    
    test::log_info "Starting Middleware API Test Suite"
    
    test_middleware_and_health "$tenant_id"

    test::print_summary
}

main
