#!/usr/bin/env bash
# Finance service tests — accounts, transactions, reports.

MENU+=("finance:Finance — Accounts · Transactions · Trial balance:run_finance")

run_finance() {
  header "Finance — Accounts & Transactions"
  reset_counters
  require_auth || { warn "Skipping Finance (not authenticated)"; return 1; }

  # ── Accounts ─────────────────────────────────────────────────────────────────
  sep; info "List accounts"
  local resp body status
  resp=$(apiv GET /api/v1/finance/accounts)
  split_resp "$resp"; body="$RESP_BODY"; status="$RESP_STATUS"
  assert_any "GET /api/v1/finance/accounts" "$status" "200" "403" || true
  if [[ "$status" == "200" ]]; then
    ok "  → $(json_len "$body" "data") account(s)"
  fi

  # Create a test account (expects finance.accounts.write permission)
  sep; info "Create test account"
  resp=$(apiv POST /api/v1/finance/accounts -d "{
    \"name\":         \"Test Account ${TEST_SUFFIX}\",
    \"code\":         \"TST-${TEST_SUFFIX}\",
    \"account_type\": \"ASSET\",
    \"currency\":     \"NGN\"
  }")
  split_resp "$resp"; body="$RESP_BODY"; status="$RESP_STATUS"
  assert_any "POST /api/v1/finance/accounts" "$status" "201" "403" "422" || true

  local acct_id=""
  if [[ "$status" == "201" ]]; then
    acct_id="$(jf "$body" "data.id")"
    [[ -z "$acct_id" || "$acct_id" == "None" ]] && acct_id="$(jf "$body" "id")"
    ok "  → id: ${acct_id:0:8}…"

    sep; info "Get account balance"
    resp=$(apiv GET "/api/v1/finance/accounts/$acct_id/balance")
    split_resp "$resp"; status="$RESP_STATUS"
    assert_any "GET /api/v1/finance/accounts/:id/balance" "$status" "200" "403" || true
  fi

  # ── Transactions ─────────────────────────────────────────────────────────────
  sep; info "List transactions"
  resp=$(apiv GET /api/v1/finance/transactions)
  split_resp "$resp"; body="$RESP_BODY"; status="$RESP_STATUS"
  assert_any "GET /api/v1/finance/transactions" "$status" "200" "403" || true

  # ── Reports ──────────────────────────────────────────────────────────────────
  sep; info "Trial balance report"
  resp=$(apiv GET /api/v1/finance/reports/trial-balance)
  split_resp "$resp"; status="$RESP_STATUS"
  assert_any "GET /api/v1/finance/reports/trial-balance" "$status" "200" "403" || true

  sep
  if [[ $SUITE_FAILURES -eq 0 ]]; then
    ok "Finance suite passed"
  else
    fail "Finance suite: $SUITE_FAILURES failure(s)"
  fi
}
