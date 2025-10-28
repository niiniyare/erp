#!/bin/bash

# Health API Test Script

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/api_client.sh"
source "$SCRIPT_DIR/lib/test_helpers.sh"

# --- Test Cases ---

test_health_endpoints() {
    test::start "Health Check Endpoints"

    local endpoints=(
        "/health"
        "/api/v1/tenants/health"
        "/api/v1/feature-flags/health"
        "/api/v1/abac/health"
        "/api/v1/finance/health"
    )

    for endpoint in "${endpoints[@]}"; do
        local resp=$(api::get_public "$endpoint")
        local test_name="health_check_for_${endpoint//\//_}"
        assert::http_status "$test_name" "$resp" "200"
    done
}

# --- Main Execution ---
main() {
    test::on_exit test::print_summary
    test::log_info "Starting Health API Test Suite"
    test_health_endpoints
}

main
