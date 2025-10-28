#!/bin/bash

# Finance API Test Script

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/api_client.sh"
source "$SCRIPT_DIR/lib/test_helpers.sh"

# --- Test Cases ---

test_finance_flow() {
    local tenant_id="$1"
    test::start "Finance API Flow"

    # 1. Create Account
    local account_payload='{
        "account_code": "1010", "account_name": "Operating Cash", "root_type": "ASSET",
        "account_type": "CASH", "normal_balance": "DEBIT", "currency_code": "USD",
        "is_active": true, "allow_manual_entries": true
    }'
    local create_acct_resp=$(api::post "/api/v1/finance/accounts" "$tenant_id" "$account_payload")
    local create_acct_body="${create_acct_resp%???}"
    local account_id=""
    if assert::http_status "create_finance_account" "$create_acct_resp" "201"; then
        account_id=$(assert::json_value "get_account_id" "$create_acct_body" ".id")
        if [ -n "$account_id" ]; then
            test::register_resource "/api/v1/finance/accounts/$account_id" "$tenant_id"
        fi
    fi

    # 2. List Accounts
    local list_accts_resp=$(api::get "/api/v1/finance/accounts" "$tenant_id")
    assert::http_status "list_finance_accounts" "$list_accts_resp" "200"

    # 3. Create another account for transactions
    local account_payload_2='{
        "account_code": "3010", "account_name": "Sales Revenue", "root_type": "REVENUE",
        "account_type": "REVENUE", "normal_balance": "CREDIT", "currency_code": "USD",
        "is_active": true, "allow_manual_entries": true
    }'
    local create_acct_resp_2=$(api::post "/api/v1/finance/accounts" "$tenant_id" "$account_payload_2")
    local create_acct_body_2="${create_acct_resp_2%???}"
    local account_id_2=""
    if assert::http_status "create_revenue_account" "$create_acct_resp_2" "201"; then
        account_id_2=$(assert::json_value "get_revenue_account_id" "$create_acct_body_2" ".id")
        if [ -n "$account_id_2" ]; then
            test::register_resource "/api/v1/finance/accounts/$account_id_2" "$tenant_id"
        fi
    fi

    # 4. Create Transaction
    local trans_payload=$(cat <<EOF
{
    "transaction_type": "JOURNAL", "transaction_date": "$(date +%Y-%m-%d)",
    "description": "Test sale transaction", "currency": "USD",
    "entries": [
        {"account_code": "1010", "debit_amount": "500.00"},
        {"account_code": "3010", "credit_amount": "500.00"}
    ],
    "auto_approve": true
}
EOF
)
    local create_trans_resp=$(api::post "/api/v1/finance/transactions" "$tenant_id" "$trans_payload")
    local create_trans_body="${create_trans_resp%???}"
    local trans_id=""
    if assert::http_status "create_transaction" "$create_trans_resp" "201"; then
        trans_id=$(assert::json_value "get_transaction_id" "$create_trans_body" ".id")
        # Transactions might not be directly deletable via API, so not registering for cleanup
    fi

    # 5. Get Trial Balance
    local trial_bal_resp=$(api::get "/api/v1/finance/reports/trial-balance" "$tenant_id")
    assert::http_status "get_trial_balance" "$trial_bal_resp" "200"
}

# --- Main Execution ---
main() {
    test::on_exit cleanup
    
    test::log_info "Setting up test environment..."
    local tenant_id=$(setup::create_tenant)
    
    test::log_info "Starting Finance API Test Suite"
    test::log_info "Tenant: $tenant_id"

    # Run tests
    test_finance_flow "$tenant_id"

    test::print_summary
}

main
