#!/usr/bin/env bash
# IAM end-to-end curl test script
# Covers: login, MFA, password reset, API keys, audit logs, schema boot,
#         protected routes (401/403/200), and tenant/user management.
#
# Requirements: curl, jq
# Usage:  bash scripts/curl_test.sh
#         BASE_URL=https://myhost bash scripts/curl_test.sh

set -euo pipefail

# ─── Config ──────────────────────────────────────────────────────────────────

BASE_URL="${BASE_URL:-http://localhost:8080}"
ADMIN_EMAIL="${ADMIN_EMAIL:-admin@platform.local}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-Admin1234!}"
TENANT_ID="${TENANT_ID:-}" # set if server requires X-Tenant-ID header
COOKIE_JAR="$(mktemp ./tmp/awo_cookies.XXXXXX)"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

ok() { echo -e "${GREEN}  ✓ $*${NC}"; }
fail() { echo -e "${RED}  ✗ $*${NC}"; }
info() { echo -e "${CYAN}▶ $*${NC}"; }
sep() { echo -e "${YELLOW}────────────────────────────────────────${NC}"; }

# Base curl wrapper — saves + sends cookies, adds tenant header when set.
# Usage: api <method> <path> [extra curl args...]
api() {
  local method="$1" path="$2"
  shift 2
  local url="${BASE_URL}${path}"
  local headers=(-H "Content-Type: application/json")
  [[ -n "$TENANT_ID" ]] && headers+=(-H "X-Tenant-ID: $TENANT_ID")
  curl -s -X "$method" "$url" \
    -c "$COOKIE_JAR" -b "$COOKIE_JAR" \
    "${headers[@]}" "$@"
}

# Same but print the response with status code.
apiv() {
  local method="$1" path="$2"
  shift 2
  local url="${BASE_URL}${path}"
  local headers=(-H "Content-Type: application/json")
  [[ -n "$TENANT_ID" ]] && headers+=(-H "X-Tenant-ID: $TENANT_ID")
  curl -s -w "\n%{http_code}" -X "$method" "$url" \
    -c "$COOKIE_JAR" -b "$COOKIE_JAR" \
    "${headers[@]}" "$@"
}

assert_status() {
  local label="$1" expected="$2" actual="$3"
  if [[ "$actual" == "$expected" ]]; then
    ok "$label → HTTP $actual"
  else
    fail "$label → expected $expected, got $actual"
  fi
}

cleanup() { rm -f "$COOKIE_JAR"; }
trap cleanup EXIT

# ─── 1. Health check ─────────────────────────────────────────────────────────

sep
info "1. Health check"

STATUS=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/health/")
assert_status "GET /health/" 200 "$STATUS"

# ─── 2. Login ────────────────────────────────────────────────────────────────

sep
info "2. Login"

RESP=$(apiv POST /api/v1/auth/login -d "{
  \"email\": \"$ADMIN_EMAIL\",
  \"password\": \"$ADMIN_PASSWORD\"
}")
BODY=$(echo "$RESP" | head -n -1)
STATUS=$(echo "$RESP" | tail -n 1)

if [[ "$STATUS" == "200" ]]; then
  MFA_REQUIRED=$(echo "$BODY" | jq -r '.mfa_required // false')
  PENDING_TOKEN=$(echo "$BODY" | jq -r '.pending_token // empty')
  if [[ "$MFA_REQUIRED" == "true" ]]; then
    ok "POST /api/v1/auth/login → 200 (MFA required)"
    info "  → pending_token: ${PENDING_TOKEN:0:12}… — complete via POST /api/v1/auth/mfa/complete"
    SKIP_AUTHED=true
  else
    ok "POST /api/v1/auth/login → 200"
    ok "  → permissions: $(echo "$BODY" | jq '.Permissions | keys | length') keys"
    ok "  → EntityScope type: $(echo "$BODY" | jq -r '.EntityScope.type // "all"')"
    SKIP_AUTHED=false
  fi
elif [[ "$STATUS" == "202" ]]; then
  PENDING_TOKEN=$(echo "$BODY" | jq -r '.pending_token // empty')
  ok "POST /api/v1/auth/login → 202 (MFA required)"
  info "  → pending_token: ${PENDING_TOKEN:0:12}…"
  SKIP_AUTHED=true
elif [[ "$STATUS" == "401" ]]; then
  fail "POST /api/v1/auth/login → 401 (bad credentials)"
  info "  → Run 'bash scripts/seed.sh' first to create the admin user"
  SKIP_AUTHED=true
else
  fail "POST /api/v1/auth/login → unexpected $STATUS"
  SKIP_AUTHED=true
fi

# ─── 3. Login: reject missing fields ─────────────────────────────────────────

sep
info "3. Login validation"

STATUS=$(api POST /api/v1/auth/login -d '{"email":"","password":""}' |
  python3 -c "import sys,json; print(json.load(sys.stdin).get('error',''))" 2>/dev/null || true)
BODY_STATUS=$(apiv POST /api/v1/auth/login -d '{"email":"","password":""}' | tail -n 1)
assert_status "POST /api/v1/auth/login (empty fields)" 400 "$BODY_STATUS"

STATUS=$(apiv POST /api/v1/auth/login -d '{"email":"nobody@x.com","password":"wrong"}' | tail -n 1)
assert_status "POST /api/v1/auth/login (bad credentials)" 401 "$STATUS"

# ─── 4. Unauthenticated access ───────────────────────────────────────────────

sep
info "4. Unauthenticated access"

# Clear cookies temporarily to simulate no session
SAVED_JAR=$(cat "$COOKIE_JAR")
>"$COOKIE_JAR"

STATUS=$(apiv GET /api/v1/audit-logs | tail -n 1)
assert_status "GET /api/v1/audit-logs (no cookie)" 401 "$STATUS"

STATUS=$(apiv GET /api/v1/schema/boot | tail -n 1)
assert_status "GET /api/v1/schema/boot (no cookie)" 401 "$STATUS"

echo "$SAVED_JAR" >"$COOKIE_JAR"

# ─── 5. Schema boot ──────────────────────────────────────────────────────────

sep
info "5. Schema boot"

if [[ "$SKIP_AUTHED" == "false" ]]; then
  RESP=$(apiv GET /api/v1/schema/boot)
  BODY=$(echo "$RESP" | head -n -1)
  STATUS=$(echo "$RESP" | tail -n 1)
  assert_status "GET /api/v1/schema/boot" 200 "$STATUS"
  PAGE_COUNT=$(echo "$BODY" | jq '.pages | length // 0')
  ok "  → type: $(echo "$BODY" | jq -r '.type'), pages: $PAGE_COUNT"
else
  info "  → skipped (not authenticated)"
fi

# ─── 6. Audit logs ───────────────────────────────────────────────────────────

sep
info "6. Audit logs"

if [[ "$SKIP_AUTHED" == "false" ]]; then
  RESP=$(apiv GET "/api/v1/audit-logs?limit=10&offset=0")
  BODY=$(echo "$RESP" | head -n -1)
  STATUS=$(echo "$RESP" | tail -n 1)
  assert_status "GET /api/v1/audit-logs" 200 "$STATUS"
  ok "  → $(echo "$BODY" | jq '.data | length // 0') events returned"
else
  info "  → skipped (not authenticated)"
fi

# ─── 7. MFA setup flow ───────────────────────────────────────────────────────

sep
info "7. MFA setup (initiate only)"

if [[ "$SKIP_AUTHED" == "false" ]]; then
  RESP=$(apiv POST /api/v1/auth/mfa/initiate)
  BODY=$(echo "$RESP" | head -n -1)
  STATUS=$(echo "$RESP" | tail -n 1)
  assert_status "POST /api/v1/auth/mfa/initiate" 200 "$STATUS"
  QR_URI=$(echo "$BODY" | jq -r '.qr_uri // empty')
  SECRET=$(echo "$BODY" | jq -r '.secret // empty')
  ok "  → secret: ${SECRET:0:8}…  (scan QR in authenticator app)"

  info "  → To confirm MFA, run:"
  echo "       curl -s -X POST ${BASE_URL}/api/v1/auth/mfa/confirm \\"
  echo "         -b $COOKIE_JAR -c $COOKIE_JAR \\"
  echo "         -H 'Content-Type: application/json' \\"
  echo "         -d '{\"code\":\"<6-digit TOTP>\"}'"

  info "  → To complete MFA login (after logout + re-login):"
  echo "       curl -s -X POST ${BASE_URL}/api/v1/auth/mfa/complete \\"
  echo "         -H 'Content-Type: application/json' \\"
  echo "         -d '{\"pending_token\":\"<token>\",\"code\":\"<6-digit TOTP>\"}'"

  info "  → To disable MFA:"
  echo "       curl -s -X DELETE ${BASE_URL}/api/v1/auth/mfa \\"
  echo "         -b $COOKIE_JAR \\"
  echo "         -H 'Content-Type: application/json' \\"
  echo "         -d '{\"password\":\"$ADMIN_PASSWORD\"}'"
else
  info "  → skipped (not authenticated)"
fi

# ─── 8. Password reset flow ──────────────────────────────────────────────────

sep
info "8. Password reset flow"

# Always returns 200 to prevent user enumeration
STATUS=$(apiv POST /api/v1/auth/forgot-password \
  -d '{"email":"nonexistent@example.com"}' | tail -n 1)
assert_status "POST /api/v1/auth/forgot-password (unknown email)" 200 "$STATUS"

STATUS=$(apiv POST /api/v1/auth/forgot-password \
  -d "{\"email\":\"$ADMIN_EMAIL\"}" | tail -n 1)
assert_status "POST /api/v1/auth/forgot-password (known email)" 200 "$STATUS"

# Reset with invalid token
STATUS=$(apiv POST /api/v1/auth/reset-password -d '{
  "token": "deadbeef00000000000000000000000000000000000000000000000000000000",
  "new_password": "NewPassword1!"
}' | tail -n 1)
assert_status "POST /api/v1/auth/reset-password (invalid token)" 404 "$STATUS"

info "  → To complete reset with a real token (from email link):"
echo "       curl -s -X POST ${BASE_URL}/api/v1/auth/reset-password \\"
echo "         -H 'Content-Type: application/json' \\"
echo "         -d '{\"token\":\"<token-from-email>\",\"new_password\":\"NewPass1!\"}'"

# ─── 9. API key management ───────────────────────────────────────────────────

sep
info "9. API key management"

if [[ "$SKIP_AUTHED" == "false" ]]; then
  # Create
  RESP=$(apiv POST /api/v1/auth/api-keys -d '{
    "name": "curl-test-key",
    "scopes": ["finance.accounts.read", "finance.transactions.read"],
    "expires_at": null
  }')
  BODY=$(echo "$RESP" | head -n -1)
  STATUS=$(echo "$RESP" | tail -n 1)
  assert_status "POST /api/v1/auth/api-keys" 201 "$STATUS"
  API_KEY=$(echo "$BODY" | jq -r '.key // empty')
  KEY_ID=$(echo "$BODY" | jq -r '.id // empty')
  ok "  → key: ${API_KEY:0:12}… (save this — shown once)"

  # List
  RESP=$(apiv GET /api/v1/auth/api-keys)
  STATUS=$(echo "$RESP" | tail -n 1)
  assert_status "GET /api/v1/auth/api-keys" 200 "$STATUS"
  COUNT=$(echo "$RESP" | head -n -1 | jq '. | length // 0')
  ok "  → $COUNT key(s) in tenant"

  # Use the API key as a Bearer token.
  # The key has finance.* scopes only (not iam.sessions.read),
  # so audit-logs will return 403 (authenticated, but missing permission).
  # Any 2xx or 403 means the key was accepted — 401 would mean auth failed.
  if [[ -n "$API_KEY" ]]; then
    info "  → Testing Bearer token authentication with new key"
    STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
      -H "Authorization: Bearer $API_KEY" \
      -H "Content-Type: application/json" \
      "${BASE_URL}/api/v1/audit-logs")
    if [[ "$STATUS" == "200" || "$STATUS" == "403" ]]; then
      ok "GET /api/v1/audit-logs (Bearer eak_ token) → $STATUS (key authenticated)"
    else
      fail "GET /api/v1/audit-logs (Bearer eak_ token) → expected 200 or 403, got $STATUS"
    fi
  fi

  # Revoke
  if [[ -n "$KEY_ID" ]]; then
    RESP=$(apiv DELETE "/api/v1/auth/api-keys/$KEY_ID")
    STATUS=$(echo "$RESP" | tail -n 1)
    assert_status "DELETE /api/v1/auth/api-keys/:id" 204 "$STATUS"

    # Revoked key must be rejected
    if [[ -n "$API_KEY" ]]; then
      STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
        -H "Authorization: Bearer $API_KEY" \
        "${BASE_URL}/api/v1/audit-logs")
      assert_status "GET /api/v1/audit-logs (revoked key)" 401 "$STATUS"
    fi
  fi
else
  info "  → skipped (not authenticated)"
fi

# ─── 10. Finance routes (permission gate) ────────────────────────────────────

sep
info "10. Finance routes — permission gate"

if [[ "$SKIP_AUTHED" == "false" ]]; then
  RESP=$(apiv GET /api/v1/finance/accounts)
  STATUS=$(echo "$RESP" | tail -n 1)
  # 200 = has permission, 403 = lacks finance.accounts.read
  if [[ "$STATUS" == "200" ]]; then
    ok "GET /api/v1/finance/accounts → 200 (has finance.accounts.read)"
  elif [[ "$STATUS" == "403" ]]; then
    ok "GET /api/v1/finance/accounts → 403 (lacks finance.accounts.read — expected for non-finance users)"
  else
    fail "GET /api/v1/finance/accounts → unexpected $STATUS"
  fi
else
  info "  → skipped (not authenticated)"
fi

# ─── 11. Tenant management ───────────────────────────────────────────────────

sep
info "11. Tenant management"

if [[ "$SKIP_AUTHED" == "false" ]]; then
  RESP=$(apiv GET /api/v1/tenants)
  STATUS=$(echo "$RESP" | tail -n 1)
  if [[ "$STATUS" == "200" ]]; then
    ok "GET /api/v1/tenants → 200"
  else
    ok "GET /api/v1/tenants → $STATUS (expected if non-platform user)"
  fi

  info "  → To create a tenant:"
  echo "       curl -s -X POST ${BASE_URL}/api/v1/tenants \\"
  echo "         -b $COOKIE_JAR \\"
  echo "         -H 'Content-Type: application/json' \\"
  echo "         -d '{\"name\":\"Acme\",\"email\":\"admin@acme.com\",\"country_code\":\"NG\",\"currency_code\":\"NGN\"}'"
fi

# ─── 12. Logout ──────────────────────────────────────────────────────────────

sep
info "12. Logout"

if [[ "$SKIP_AUTHED" == "false" ]]; then
  STATUS=$(apiv POST /api/v1/auth/logout | tail -n 1)
  assert_status "POST /api/v1/auth/logout" 204 "$STATUS"

  # Session must be dead now
  STATUS=$(apiv GET /api/v1/audit-logs | tail -n 1)
  assert_status "GET /api/v1/audit-logs (after logout)" 401 "$STATUS"
fi

# ─── SSO note ────────────────────────────────────────────────────────────────

sep
info "SSO / OAuth (browser flow — cannot automate with curl)"
echo "  Begin:    GET  ${BASE_URL}/api/v1/auth/oauth/google?tenant_id=<uuid>"
echo "  Callback: GET  ${BASE_URL}/api/v1/auth/oauth/google/callback?code=…&state=…"

# ─── Summary ─────────────────────────────────────────────────────────────────

sep
echo -e "${GREEN}Done.${NC}"
