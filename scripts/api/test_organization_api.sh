#!/bin/bash

# Organization API Test Script
# Tests hierarchical multi-tenant organization structure.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/api_client.sh"
source "$SCRIPT_DIR/lib/test_helpers.sh"

# --- Test Data ---
declare -A TEST_ORGS=(
    ["root_company"]='{"name":"ACME Corporation","entity_type":"COMPANY"}'
    ["subsidiary"]='{"name":"ACME Europe GmbH","entity_type":"SUBSIDIARY","parent_id":"%root_company%"}'
    ["region_west"]='{"name":"West Region","entity_type":"REGION","parent_id":"%root_company%"}'
    ["branch_sf"]='{"name":"San Francisco Branch","entity_type":"BRANCH","parent_id":"%region_west%"}'
    ["department_eng"]='{"name":"Engineering Department","entity_type":"DEPARTMENT","parent_id":"%branch_sf%"}'
)
CREATION_ORDER=("root_company" "subsidiary" "region_west" "branch_sf" "department_eng")
declare -A CREATED_ORG_IDS # Maps logical name (root_company) to UUID

# --- Test Cases ---

test_hierarchical_creation() {
    local tenant_id="$1"
    test::start "Hierarchical Organization Creation"

    for org_key in "${CREATION_ORDER[@]}"; do
        local org_data="${TEST_ORGS[$org_key]}"

        if [[ "$org_data" == *"%"* ]]; then
            local parent_key=$(echo "$org_data" | grep -o '%\w*%' | tr -d '%')
            local parent_id="${CREATED_ORG_IDS[$parent_key]}"
            if [ -z "$parent_id" ]; then
                test::record "create_$org_key" "FAIL" "Parent ID for '$parent_key' not found. Skipping."
                continue
            fi
            org_data=$(echo "$org_data" | jq --arg pid "$parent_id" '.parent_id = $pid')
        fi

        local response=$(api::post "/api/v1/organizations" "$tenant_id" "$org_data")
        local body="${response%???}"

        if assert::http_status "create_$org_key" "$response" "201"; then
            local org_id=$(assert::json_value "get_id_for_$org_key" "$body" ".id")
            if [ -n "$org_id" ]; then
                CREATED_ORG_IDS["$org_key"]="$org_id"
                test::register_resource "/api/v1/organizations/$org_id" "$tenant_id"
            fi
        fi
    done
}

test_retrieval_and_listing() {
    local tenant_id="$1"
    test::start "Organization Retrieval and Listing"

    for org_key in "${!CREATED_ORG_IDS[@]}"; do
        local org_id="${CREATED_ORG_IDS[$org_key]}"
        local response=$(api::get "/api/v1/organizations/$org_id" "$tenant_id")
        
        if assert::http_status "get_$org_key" "$response" "200"; then
            local body="${response%???}"
            local retrieved_id=$(assert::json_value "verify_id_for_$org_key" "$body" ".id")
            if [ "$retrieved_id" != "$org_id" ]; then
                test::record "verify_id_for_$org_key" "FAIL" "ID mismatch: expected $org_id, got $retrieved_id"
            fi
        fi
    done

    local list_response=$(api::get "/api/v1/organizations?page=1&page_size=20" "$tenant_id")
    if assert::http_status "list_orgs" "$list_response" "200"; then
        local body="${list_response%???}"
        local count=$(echo "$body" | jq -r '.data | length')
        local expected_count=${#CREATED_ORG_IDS[@]}
        
        if [ "$count" -ge "$expected_count" ]; then
            test::record "list_count" "PASS"
        else
            test::record "list_count" "FAIL" "Expected at least $expected_count orgs, got $count"
        fi
    fi
}

test_hierarchy_endpoint() {
    local tenant_id="$1"
    test::start "Organization Hierarchy Endpoint"

    local root_id="${CREATED_ORG_IDS[root_company]}"
    if [ -z "$root_id" ]; then
        test::log_warn "Root company not created, skipping hierarchy test."
        return
    fi

    local response=$(api::get "/api/v1/organizations/$root_id/hierarchy?depth=3" "$tenant_id")
    if assert::http_status "get_hierarchy" "$response" "200"; then
        local body="${response%???}"
        local retrieved_root_id=$(assert::json_value "verify_hierarchy_root_id" "$body" ".root.id")
        if [ "$retrieved_root_id" != "$root_id" ]; then
            test::record "verify_hierarchy_root_id" "FAIL" "Hierarchy root ID mismatch"
        fi
    fi
}

test_tenant_isolation() {
    local tenant1_id="$1"
    local tenant2_id="$2"
    test::start "Multi-Tenant Isolation"

    local org_id_t1="${CREATED_ORG_IDS[root_company]}"
    if [ -z "$org_id_t1" ]; then
        test::log_warn "No org from tenant 1 found, skipping isolation test."
        return
    fi

    local response=$(api::get "/api/v1/organizations/$org_id_t1" "$tenant2_id")
    assert::http_status "access_org_from_wrong_tenant" "$response" "404"
}

main() {
    test::on_exit cleanup
    
    test::log_info "Setting up test environment..."
    local tenant1_id=$(setup::create_tenant)
    local tenant2_id=$(setup::create_tenant)
    
    test::log_info "Starting Organization API Test Suite"
    test::log_info "Tenant 1: $tenant1_id"
    test::log_info "Tenant 2: $tenant2_id"

    test_hierarchical_creation "$tenant1_id"
    test_retrieval_and_listing "$tenant1_id"
    test_hierarchy_endpoint "$tenant1_id"
    test_tenant_isolation "$tenant1_id" "$tenant2_id"

    test::print_summary
}

main
