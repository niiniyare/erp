#!/usr/bin/env bash
# Tenants service tests — list, create, and lifecycle transitions.

MENU+=("tenants:Tenants — List · Create · Activate · Suspend:run_tenants")

run_tenants() {
  header "Tenants — Management"
  reset_counters
  require_auth || { warn "Skipping Tenants (not authenticated)"; return 1; }

  # ── List ─────────────────────────────────────────────────────────────────────
  sep; info "List tenants"
  local resp body status
  resp=$(apiv GET /api/v1/tenants)
  split_resp "$resp"; body="$RESP_BODY"; status="$RESP_STATUS"
  assert_any "GET /api/v1/tenants" "$status" "200" "403" || true
  if [[ "$status" == "200" ]]; then
    ok "  → $(json_len "$body" "data") tenant(s)"
  fi

  # ── Create ───────────────────────────────────────────────────────────────────
  sep; info "Create test tenant (${TEST_TENANT_NAME})"
  resp=$(apiv POST /api/v1/tenants -d "{
    \"name\":          \"$TEST_TENANT_NAME\",
    \"email\":         \"$TEST_TENANT_EMAIL\",
    \"country_code\":  \"NG\",
    \"currency_code\": \"NGN\"
  }")
  split_resp "$resp"; body="$RESP_BODY"; status="$RESP_STATUS"
  assert_any "POST /api/v1/tenants" "$status" "201" "403" "409" || true

  local tenant_id=""
  if [[ "$status" == "201" ]]; then
    tenant_id="$(jf "$body" "data.id")"
    [[ -z "$tenant_id" || "$tenant_id" == "None" ]] && tenant_id="$(jf "$body" "id")"
    ok "  → id: ${tenant_id:0:8}…  name: $TEST_TENANT_NAME"
  fi

  # ── Lifecycle ────────────────────────────────────────────────────────────────
  if [[ -n "$tenant_id" && "$tenant_id" != "None" ]]; then
    sep; info "Activate tenant"
    resp=$(apiv POST "/api/v1/tenants/$tenant_id/activate")
    split_resp "$resp"; status="$RESP_STATUS"
    assert_any "POST /api/v1/tenants/:id/activate" "$status" \
      "200" "204" "403" "409" "422" || true

    sep; info "Suspend tenant"
    resp=$(apiv POST "/api/v1/tenants/$tenant_id/suspend")
    split_resp "$resp"; status="$RESP_STATUS"
    assert_any "POST /api/v1/tenants/:id/suspend" "$status" \
      "200" "204" "403" "409" "422" || true

    sep; info "Get tenant by ID"
    resp=$(apiv GET "/api/v1/tenants/$tenant_id")
    split_resp "$resp"; body="$RESP_BODY"; status="$RESP_STATUS"
    assert_any "GET /api/v1/tenants/:id" "$status" "200" "403" || true
    if [[ "$status" == "200" ]]; then
      ok "  → status: $(jf "$body" "data.status")"
    fi
  fi

  sep
  if [[ $SUITE_FAILURES -eq 0 ]]; then
    ok "Tenants suite passed"
  else
    fail "Tenants suite: $SUITE_FAILURES failure(s)"
  fi
}
