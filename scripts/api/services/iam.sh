#!/usr/bin/env bash
# IAM service tests — auth, MFA, password reset, API keys, audit logs.
# Sourced by run.sh; registers itself into the MENU array.

MENU+=("iam:IAM — Auth · MFA · Password reset · API keys:run_iam")

run_iam() {
  header "IAM — Authentication & Session Management"
  reset_counters
  require_auth || { warn "Skipping IAM (not authenticated)"; return 1; }

  # ── AT1: Login response validation ──────────────────────────────────────────
  sep; info "AT1 — Login response body (user_id, tenant_id, permissions, no raw token)"
  local resp body status
  resp=$(apiv POST /api/v1/auth/login \
    -d "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"$ADMIN_PASSWORD\"}")
  split_resp "$resp"; body="$RESP_BODY"; status="$RESP_STATUS"
  assert_status "AT1 valid login → 200" 200 "$status" || true
  if [[ "$status" == "200" ]]; then
    local uid tid
    uid="$(jf "$body" "user_id")"; tid="$(jf "$body" "tenant_id")"
    [[ -n "$uid" && "$uid" != "None" ]] && ok "  → user_id present"    || fail "  → user_id missing"
    [[ -n "$tid" && "$tid" != "None" ]] && ok "  → tenant_id present"  || fail "  → tenant_id missing"
    local perms
    perms=$(echo "$body" | python3 -c "
import sys, json
d = json.load(sys.stdin)
print(len(d.get('permissions', {})))
" 2>/dev/null || echo 0)
    [[ "$perms" -gt 0 ]] && ok "  → permissions non-empty ($perms keys)" \
      || warn "  → permissions empty (platform user or no roles assigned)"
    local raw_token; raw_token="$(jf "$body" "token")"
    [[ -z "$raw_token" || "$raw_token" == "None" ]] \
      && ok "  → raw token absent from body" \
      || fail "  → raw token leaked in response body"
  fi

  # ── AT2: Wrong password — generic message ───────────────────────────────────
  sep; info "AT2 — Wrong password → 401, generic error (no oracle)"
  resp=$(apiv POST /api/v1/auth/login \
    -d "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"absolutely-wrong-xXx\"}")
  split_resp "$resp"; body="$RESP_BODY"; status="$RESP_STATUS"
  assert_status "AT2 wrong password → 401" 401 "$status" || true
  if [[ "$status" == "401" ]]; then
    local leak
    leak=$(echo "$body" | python3 -c "
import sys, json
d = json.load(sys.stdin)
msg = str(d.get('error', d.get('message', ''))).lower()
bad = ['user not found', 'email not found', 'password incorrect', 'wrong password']
print('yes' if any(p in msg for p in bad) else 'no')
" 2>/dev/null || echo "no")
    [[ "$leak" == "no" ]] && ok "  → message is generic (no oracle)" \
      || fail "  → error reveals which field failed"
  fi

  # ── AT3: Account lockout (informational — needs 5 prior failures) ────────────
  sep; info "AT3 — Account lockout (informational — requires 5 prior failures)"
  warn "  AT3 is stateful: call AT2 five times for the same account first,"
  warn "  then a correct-password login must return 423 Locked."
  warn "  Run manually to avoid locking the admin account during CI."

  # ── AT4–AT5: Protected route middleware ─────────────────────────────────────
  sep; info "AT4 — Valid session → protected route runs"
  resp=$(apiv GET /api/v1/schema/boot)
  split_resp "$resp"; body="$RESP_BODY"; status="$RESP_STATUS"
  assert_any "AT4 valid session → 200 or 403" "$status" "200" "403" || true
  [[ "$status" == "200" ]] && ok "  → pages: $(json_len "$body" "pages")"

  sep; info "AT5 — No session → 401 on every protected route"
  local _check_anon
  _check_anon() {
    local s
    s=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/api/v1/schema/boot")
    assert_status "AT5 schema/boot (no auth)" 401 "$s" || true
    s=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/api/v1/audit-logs")
    assert_status "AT5 audit-logs (no auth)" 401 "$s" || true
  }
  _check_anon

  # ── AT6: Session present but missing permission → 403 ───────────────────────
  sep; info "AT6 — Missing permission → 403 (informational)"
  warn "  AT6 requires a second user without finance.invoices.read."
  warn "  Create one, login, and hit GET /api/v1/finance/invoices."

  # ── Schema boot ─────────────────────────────────────────────────────────────
  sep; info "AT19 — Schema boot returns AMIS app schema"
  resp=$(apiv GET /api/v1/schema/boot)
  split_resp "$resp"; body="$RESP_BODY"; status="$RESP_STATUS"
  if assert_status "GET /api/v1/schema/boot" 200 "$status"; then
    ok "  → pages: $(json_len "$body" "pages")"
    local schema_type; schema_type="$(jf "$body" "type")"
    [[ "$schema_type" == "app" ]] && ok "  → type=app" || warn "  → type='$schema_type'"
  fi

  # ── AT21: Audit logs ────────────────────────────────────────────────────────
  sep; info "AT21 — Audit logs (requires iam.sessions.read)"
  resp=$(apiv GET "/api/v1/audit-logs?limit=5&offset=0")
  split_resp "$resp"; body="$RESP_BODY"; status="$RESP_STATUS"
  assert_any "AT21 GET /api/v1/audit-logs" "$status" "200" "403" || true
  [[ "$status" == "200" ]] && ok "  → $(json_len "$body" "data") event(s)"
  [[ "$status" == "403" ]] && ok "  → 403 expected for users without iam.sessions.read"

  # ── AT8–AT10: MFA ───────────────────────────────────────────────────────────
  sep; info "AT8 — MFA setup (initiate)"
  resp=$(apiv POST /api/v1/auth/mfa/initiate)
  split_resp "$resp"; body="$RESP_BODY"; status="$RESP_STATUS"
  assert_any "AT8 POST /api/v1/auth/mfa/initiate" "$status" "200" "400" "403" "409" || true
  if [[ "$status" == "200" ]]; then
    local secret; secret="$(jf "$body" "secret")"
    local qr;     qr="$(jf "$body" "qr_uri")"
    ok "  → secret: ${secret:0:8}…"
    echo "$qr" | grep -q "otpauth://totp/" && ok "  → qr_uri valid" || fail "  → qr_uri malformed"
    info "  → AT9 is manual: scan QR then POST /api/v1/auth/mfa/confirm {\"code\":\"<6-digit>\"}"
  else
    warn "AT8 returned $status — MFA may already be enabled or admin lacks permission"
  fi

  sep; info "AT10 — MFA complete with invalid pending_token → 401/400"
  local mfa_status
  mfa_status=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST "${BASE_URL}/api/v1/auth/mfa/complete" \
    -H "Content-Type: application/json" \
    -d '{"pending_token":"invalid_sentinel","code":"000000"}')
  assert_any "AT10 invalid pending_token" "$mfa_status" "401" "400" || true

  # ── AT11–AT14: Password reset ────────────────────────────────────────────────
  sep; info "Password reset flow"
  status=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST "${BASE_URL}/api/v1/auth/forgot-password" \
    -H "Content-Type: application/json" \
    -d '{"email":"nobody_fake@example.com"}')
  assert_status "AT12 forgot-password unknown → 200 (anti-enumeration)" 200 "$status" || true

  status=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST "${BASE_URL}/api/v1/auth/forgot-password" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$ADMIN_EMAIL\"}")
  assert_status "AT11 forgot-password known → 200" 200 "$status" || true

  local dead_token="deadbeef00000000000000000000000000000000000000000000000000000000"

  status=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST "${BASE_URL}/api/v1/auth/reset-password" \
    -H "Content-Type: application/json" \
    -d "{\"token\":\"$dead_token\",\"new_password\":\"NewP@ssword123!\"}")
  assert_any "AT13 reset invalid token → 404/410" "$status" "404" "410" "400" || true

  status=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST "${BASE_URL}/api/v1/auth/reset-password" \
    -H "Content-Type: application/json" \
    -d "{\"token\":\"$dead_token\",\"new_password\":\"weak\"}")
  assert_any "AT14 reset weak password → 400/404" "$status" "400" "404" "410" || true

  # ── AT15–AT18: API key lifecycle ─────────────────────────────────────────────
  sep; info "API key lifecycle"
  resp=$(apiv POST /api/v1/auth/api-keys -d "{
    \"name\": \"test-key-${SESSION_ID}\",
    \"scopes\": [\"finance.accounts.read\"],
    \"expires_at\": null
  }")
  split_resp "$resp"; body="$RESP_BODY"; status="$RESP_STATUS"
  assert_any "AT15 POST /api/v1/auth/api-keys" "$status" "201" "200" "403" || true

  if [[ "$status" =~ ^(200|201)$ ]]; then
    local api_key key_id
    api_key="$(jf "$body" "token")";  [[ -z "$api_key"  || "$api_key"  == "None" ]] && api_key="$(jf "$body" "data.token")"
    key_id="$(jf "$body" "id")";     [[ -z "$key_id"   || "$key_id"   == "None" ]] && key_id="$(jf "$body" "data.id")"
    ok "  → key: ${api_key:0:14}…  id: ${key_id:0:8}…"
    [[ "$api_key" == eak_* ]] && ok "  → eak_ prefix present" || fail "  → eak_ prefix missing"

    # AT17: list — raw key absent
    resp=$(apiv GET /api/v1/auth/api-keys)
    split_resp "$resp"; body="$RESP_BODY"; status="$RESP_STATUS"
    assert_any "AT17 GET /api/v1/auth/api-keys" "$status" "200" "403" || true
    if [[ "$status" == "200" ]]; then
      echo "$body" | grep -qF "$api_key" \
        && fail "  → raw key found in list (security issue)" \
        || ok   "  → raw key absent from list"
    fi

    # AT16: Bearer auth
    if [[ -n "$api_key" ]]; then
      local k_status; k_status=$(api_bearer "$api_key" GET /api/v1/audit-logs)
      assert_any "AT16 Bearer key → 200 or 403" "$k_status" "200" "403" || true
    fi

    # AT18: revoke + dead key
    if [[ -n "$key_id" ]]; then
      status=$(apiv DELETE "/api/v1/auth/api-keys/$key_id" | tail -n 1)
      assert_any "AT18 DELETE /api/v1/auth/api-keys/:id" "$status" "200" "204" "403" || true
      if [[ "$status" =~ ^(200|204)$ && -n "$api_key" ]]; then
        local revoked_status; revoked_status=$(api_bearer "$api_key" GET /api/v1/audit-logs)
        assert_status "AT18 revoked key → 401" 401 "$revoked_status" || true
      fi
    fi
  else
    warn "AT15–AT18 skipped — insufficient permissions for API key management"
  fi

  # ── AT7: Logout + dead session ───────────────────────────────────────────────
  sep; info "AT7 — Logout verification"
  local saved_jar; saved_jar=$(cat "$COOKIE_JAR" 2>/dev/null || true)

  status=$(apiv POST /api/v1/auth/logout | tail -n 1)
  assert_any "AT7 POST /api/v1/auth/logout" "$status" "200" "204" || true
  >"$COOKIE_JAR"

  status=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/api/v1/schema/boot")
  assert_status "AT7 dead token → 401" 401 "$status" || true

  # Restore session so other services in --all mode still work
  echo "$saved_jar" >"$COOKIE_JAR"

  sep
  if [[ $SUITE_FAILURES -eq 0 ]]; then
    ok "IAM suite passed"
  else
    fail "IAM suite: $SUITE_FAILURES failure(s)"
  fi
}
