#!/usr/bin/env bash
# =============================================================================
# scripts/api/08_api_keys.sh — API Key management
#
# Covers:
#   POST   /api/v1/auth/api-keys       Create key (requires session)
#   GET    /api/v1/auth/api-keys       List keys  (requires session)
#   DELETE /api/v1/auth/api-keys/:id   Revoke key (requires session)
#
# Key format: "eak_" + 64 hex chars  (e.g. eak_a1b2c3...)
# Only the SHA-256 hash is stored — raw token returned exactly once at creation.
# Validate by sending:  Authorization: Bearer eak_...
#
# Prerequisites:
#   - alice is logged in (session from 04_auth_session.sh)
#
# Usage:
#   bash scripts/api/08_api_keys.sh
# =============================================================================

source "$(dirname "$0")/00_config.sh"

# Ensure alice is logged in
hdr "0. POST /api/v1/auth/login — Ensure alice session"

http "${HTTP_FLAGS[@]}" \
  POST "${BASE_URL}/api/v1/auth/login" \
  "Host:${TENANT_HOST}" \
  email="alice@acme.example.com" \
  password="Str0ng!Pass#2024"

# ─── 1. CREATE API KEY ───────────────────────────────────────────────────────
hdr "1. POST /api/v1/auth/api-keys — Create API key"
echo "  The plaintext token (eak_...) is returned ONCE. Save it immediately."

CREATE_RESPONSE=$(http "${HTTP_FLAGS[@]}" \
  POST "${BASE_URL}/api/v1/auth/api-keys" \
  "Host:${TENANT_HOST}" \
  name="CI Pipeline Key" \
  description="Used by GitHub Actions for automated API tests" \
  scopes:='["finance.accounts.read","finance.transactions.read"]' \
  expires_at="2026-12-31T23:59:59Z")

echo "$CREATE_RESPONSE"

API_KEY_ID=$(echo "$CREATE_RESPONSE" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    print(d.get('data', {}).get('id', d.get('id', '')))
except:
    pass
" 2>/dev/null)

API_KEY_TOKEN=$(echo "$CREATE_RESPONSE" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    print(d.get('token', d.get('raw_token', d.get('data', {}).get('token', ''))))
except:
    pass
" 2>/dev/null)

echo ""
echo "  → API_KEY_ID=${API_KEY_ID}"
echo "  → API_KEY_TOKEN=${API_KEY_TOKEN}"
echo ""
echo "  ⚠  Copy API_KEY_TOKEN now — it will never be shown again."

# ─── 2. CREATE SECOND KEY (read-only) ────────────────────────────────────────
hdr "2. POST /api/v1/auth/api-keys — Create read-only key"

http "${HTTP_FLAGS[@]}" \
  POST "${BASE_URL}/api/v1/auth/api-keys" \
  "Host:${TENANT_HOST}" \
  name="Read-Only Reporting Key" \
  scopes:='["finance.accounts.read","finance.transactions.read","audit.logs.read"]'

# ─── 3. LIST API KEYS ────────────────────────────────────────────────────────
hdr "3. GET /api/v1/auth/api-keys — List all keys for this tenant"

http "${HTTP_FLAGS[@]}" \
  GET "${BASE_URL}/api/v1/auth/api-keys" \
  "Host:${TENANT_HOST}"

# ─── 4. USE API KEY AS BEARER TOKEN ──────────────────────────────────────────
hdr "4. GET /api/v1/finance/accounts — Request using Bearer API key (no cookie)"

if [ -n "$API_KEY_TOKEN" ]; then
  http -p hbHB \
    GET "${BASE_URL}/api/v1/finance/accounts" \
    "Host:${TENANT_HOST}" \
    "Authorization:Bearer ${API_KEY_TOKEN}"
else
  echo "  → API_KEY_TOKEN not available (check response above)."
fi

# ─── 5. REVOKE API KEY ───────────────────────────────────────────────────────
hdr "5. DELETE /api/v1/auth/api-keys/:id — Revoke key"
echo "  Note: cached sessions expire within ~5 minutes (known gap)."

if [ -n "$API_KEY_ID" ]; then
  http "${HTTP_FLAGS[@]}" \
    DELETE "${BASE_URL}/api/v1/auth/api-keys/${API_KEY_ID}" \
    "Host:${TENANT_HOST}"
else
  echo "  → API_KEY_ID not available — set manually."
fi

# ─── 6. VERIFY KEY IS REVOKED ────────────────────────────────────────────────
hdr "6. GET /api/v1/finance/accounts — With revoked key (expect 401)"

if [ -n "$API_KEY_TOKEN" ]; then
  http -p hbHB \
    GET "${BASE_URL}/api/v1/finance/accounts" \
    "Host:${TENANT_HOST}" \
    "Authorization:Bearer ${API_KEY_TOKEN}"
  echo "  Note: may still succeed within the 5-min cache TTL."
fi

hdr "Done"
echo "  export API_KEY_TOKEN=${API_KEY_TOKEN}"
