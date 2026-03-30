#!/usr/bin/env bash
# Users service tests — CRUD using per-session random test data.

MENU+=("users:Users — Create · Read · List · Update · Delete:run_users")

run_users() {
  header "Users — CRUD"
  reset_counters
  require_auth || { warn "Skipping Users (not authenticated)"; return 1; }

  # ── List ─────────────────────────────────────────────────────────────────────
  sep; info "List users"
  local resp body status
  resp=$(apiv GET "/api/v1/users?limit=10")
  split_resp "$resp"; body="$RESP_BODY"; status="$RESP_STATUS"
  assert_any "GET /api/v1/users" "$status" "200" "403" || true
  if [[ "$status" == "200" ]]; then
    ok "  → $(json_len "$body" "data") user(s)"
  fi

  # ── Create ───────────────────────────────────────────────────────────────────
  sep; info "Create test user (${TEST_EMAIL})"
  resp=$(apiv POST /api/v1/users -d "{
    \"username\":     \"$TEST_USERNAME\",
    \"email\":        \"$TEST_EMAIL\",
    \"display_name\": \"Test User ${TEST_SUFFIX}\",
    \"password\":     \"$TEST_PASSWORD\",
    \"user_type\":    \"INTERNAL\",
    \"account_status\": \"ACTIVE\"
  }")
  split_resp "$resp"; body="$RESP_BODY"; status="$RESP_STATUS"
  assert_any "POST /api/v1/users" "$status" "201" "403" "409" || true

  local user_id=""
  if [[ "$status" == "201" ]]; then
    user_id="$(jf "$body" "data.id")"
    [[ -z "$user_id" || "$user_id" == "None" ]] && user_id="$(jf "$body" "id")"
    ok "  → id: ${user_id:0:8}…  email: $TEST_EMAIL"
  fi

  # ── Get by ID ────────────────────────────────────────────────────────────────
  if [[ -n "$user_id" && "$user_id" != "None" ]]; then
    sep; info "Get user by ID"
    resp=$(apiv GET "/api/v1/users/$user_id")
    split_resp "$resp"; body="$RESP_BODY"; status="$RESP_STATUS"
    assert_any "GET /api/v1/users/:id" "$status" "200" "403" || true
    if [[ "$status" == "200" ]]; then
      ok "  → email: $(jf "$body" "data.email")"
    fi

    # ── Update ─────────────────────────────────────────────────────────────────
    sep; info "Update user display name"
    resp=$(apiv PUT "/api/v1/users/$user_id" -d "{
      \"display_name\": \"Updated ${TEST_SUFFIX}\"
    }")
    split_resp "$resp"; status="$RESP_STATUS"
    assert_any "PUT /api/v1/users/:id" "$status" "200" "403" || true

    # ── Delete (soft) ──────────────────────────────────────────────────────────
    sep; info "Delete test user"
    resp=$(apiv DELETE "/api/v1/users/$user_id")
    split_resp "$resp"; status="$RESP_STATUS"
    assert_any "DELETE /api/v1/users/:id" "$status" "204" "403" || true

    if [[ "$status" == "204" ]]; then
      # Should 404 or 410 after deletion
      status=$(apiv GET "/api/v1/users/$user_id" | tail -n 1)
      assert_any "GET /api/v1/users/:id (after delete)" "$status" "404" "410" "403" || true
    fi
  fi

  sep
  if [[ $SUITE_FAILURES -eq 0 ]]; then
    ok "Users suite passed"
  else
    fail "Users suite: $SUITE_FAILURES failure(s)"
  fi
}
