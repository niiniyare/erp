#!/usr/bin/env bash
# IAM service tests — auth, MFA, password reset, API keys, audit logs.
# Sourced by run.sh; registers itself into the MENU array.

MENU+=("iam:IAM — Auth · MFA · Password reset · API keys:run_iam")

run_iam() {
  header "IAM — Authentication & Session Management"
  reset_counters
  require_auth || { warn "Skipping IAM (not authenticated)"; return 1; }

  # ── Schema boot ─────────────────────────────────────────────────────────────
  sep; info "Schema boot"
  local resp body status
  resp=$(apiv GET /api/v1/schema/boot)
  split_resp "$resp"; body="$RESP_BODY"; status="$RESP_STATUS"
  if assert_status "GET /api/v1/schema/boot" 200 "$status"; then
    ok "  → pages: $(json_len "$body" "pages")"
  fi

  # ── Audit logs ──────────────────────────────────────────────────────────────
  sep; info "Audit logs"
  resp=$(apiv GET "/api/v1/audit-logs?limit=5&offset=0")
  split_resp "$resp"; body="$RESP_BODY"; status="$RESP_STATUS"
  if assert_status "GET /api/v1/audit-logs" 200 "$status"; then
    ok "  → $(json_len "$body" "data") event(s) returned"
  fi

  # ── 401 guard (anonymous) ───────────────────────────────────────────────────
  sep; info "Unauthenticated guard"
  local _check_anon
  _check_anon() {
    local s; s=$(apiv GET /api/v1/audit-logs | tail -n 1)
    assert_status "GET /api/v1/audit-logs (no session)" 401 "$s" || true
    s=$(apiv GET /api/v1/schema/boot | tail -n 1)
    assert_status "GET /api/v1/schema/boot (no session)" 401 "$s" || true
  }
  with_anon_session _check_anon

  # ── MFA initiate ────────────────────────────────────────────────────────────
  sep; info "MFA setup (initiate)"
  resp=$(apiv POST /api/v1/auth/mfa/initiate)
  split_resp "$resp"; body="$RESP_BODY"; status="$RESP_STATUS"
  if assert_status "POST /api/v1/auth/mfa/initiate" 200 "$status"; then
    local secret; secret="$(jf "$body" "secret")"
    ok "  → secret: ${secret:0:8}…"
    info "  → Confirm with a TOTP code: POST /api/v1/auth/mfa/confirm {\"code\":\"<6-digit>\"}"
  fi

  # ── Password reset ──────────────────────────────────────────────────────────
  sep; info "Password reset flow"
  status=$(apiv POST /api/v1/auth/forgot-password \
    -d '{"email":"nobody_fake@example.com"}' | tail -n 1)
  assert_status "POST /api/v1/auth/forgot-password (unknown)" 200 "$status" || true

  status=$(apiv POST /api/v1/auth/forgot-password \
    -d "{\"email\":\"$ADMIN_EMAIL\"}" | tail -n 1)
  assert_status "POST /api/v1/auth/forgot-password (known)" 200 "$status" || true

  status=$(apiv POST /api/v1/auth/reset-password -d '{
    "token":"deadbeef00000000000000000000000000000000000000000000000000000000",
    "new_password":"NewPass1!"
  }' | tail -n 1)
  assert_status "POST /api/v1/auth/reset-password (invalid token)" 404 "$status" || true

  # ── API key lifecycle ────────────────────────────────────────────────────────
  sep; info "API key lifecycle"
  resp=$(apiv POST /api/v1/auth/api-keys -d "{
    \"name\": \"test-key-${SESSION_ID}\",
    \"scopes\": [\"finance.accounts.read\"],
    \"expires_at\": null
  }")
  split_resp "$resp"; body="$RESP_BODY"; status="$RESP_STATUS"

  if assert_status "POST /api/v1/auth/api-keys" 201 "$status"; then
    local api_key key_id
    api_key="$(jf "$body" "key")"
    key_id="$(jf "$body" "id")"
    ok "  → key: ${api_key:0:14}…  id: ${key_id:0:8}…"

    # List
    status=$(apiv GET /api/v1/auth/api-keys | tail -n 1)
    assert_status "GET /api/v1/auth/api-keys" 200 "$status" || true

    # Bearer auth test
    if [[ -n "$api_key" ]]; then
      status=$(api_bearer "$api_key" GET /api/v1/audit-logs)
      assert_any "GET /api/v1/audit-logs (Bearer key)" "$status" "200" "403" || true
    fi

    # Revoke
    if [[ -n "$key_id" ]]; then
      status=$(apiv DELETE "/api/v1/auth/api-keys/$key_id" | tail -n 1)
      assert_status "DELETE /api/v1/auth/api-keys/:id" 204 "$status" || true

      if [[ -n "$api_key" ]]; then
        status=$(api_bearer "$api_key" GET /api/v1/audit-logs)
        assert_status "GET /api/v1/audit-logs (revoked key)" 401 "$status" || true
      fi
    fi
  fi

  # ── Logout + dead-session check (restores session after) ────────────────────
  sep; info "Logout verification"
  local saved_jar; saved_jar=$(cat "$COOKIE_JAR" 2>/dev/null || true)

  status=$(apiv POST /api/v1/auth/logout | tail -n 1)
  assert_status "POST /api/v1/auth/logout" 204 "$status" || true
  >"$COOKIE_JAR"

  status=$(apiv GET /api/v1/audit-logs | tail -n 1)
  assert_status "GET /api/v1/audit-logs (after logout)" 401 "$status" || true

  # Restore session so other services in --all mode still work
  echo "$saved_jar" >"$COOKIE_JAR"

  sep
  if [[ $SUITE_FAILURES -eq 0 ]]; then
    ok "IAM suite passed"
  else
    fail "IAM suite: $SUITE_FAILURES failure(s)"
  fi
}
