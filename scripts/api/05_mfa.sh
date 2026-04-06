#!/usr/bin/env bash
# =============================================================================
# scripts/api/05_mfa.sh — MFA lifecycle (TOTP)
#
# Covers:
#   POST   /api/v1/auth/mfa/initiate  Generate TOTP secret (requires session)
#   POST   /api/v1/auth/mfa/confirm   Confirm first TOTP code (enables MFA)
#   POST   /api/v1/auth/mfa/complete  Exchange pending token + code → full session
#   DELETE /api/v1/auth/mfa           Disable MFA (requires session + password)
#
# Flow:
#   1. Initiate → get secret + QR URI
#   2. User scans QR in authenticator app
#   3. Confirm with first code → MFA enabled
#   4. Next login returns { mfa_required: true, pending_token: "..." }
#   5. Call /mfa/complete with pending_token + live TOTP code
#
# Prerequisites:
#   - Session cookie from 04_auth_session.sh (alice must be logged in)
#   - A TOTP app to generate codes from the secret
#
# Usage:
#   MFA_CODE=123456 bash scripts/api/05_mfa.sh
# =============================================================================

source "$(dirname "$0")/00_config.sh"

ADMIN_USER_ID="${ADMIN_USER_ID:-<paste-uuid>}"

# ─── 1. INITIATE MFA ─────────────────────────────────────────────────────────
hdr "1. POST /api/v1/auth/mfa/initiate — Generate TOTP secret"
echo "  Requires active session (alice must be logged in via 04_auth_session.sh)"

INITIATE_RESPONSE=$(http "${HTTP_FLAGS[@]}" \
  POST "${BASE_URL}/api/v1/auth/mfa/initiate" \
  "Host:${TENANT_HOST}")

echo "$INITIATE_RESPONSE"

# Response contains:
#   { "secret": "BASE32...", "qr_uri": "otpauth://totp/..." }
# Scan the QR URI with an authenticator app (Google Authenticator, Authy, etc.)

MFA_SECRET=$(echo "$INITIATE_RESPONSE" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    print(d.get('data', {}).get('secret', d.get('secret', '')))
except:
    pass
" 2>/dev/null)

echo ""
echo "  → MFA_SECRET=${MFA_SECRET}"
echo "  ⚠  Scan the qr_uri with your authenticator app, then set MFA_CODE below."

# ─── 2. CONFIRM MFA (enable) ─────────────────────────────────────────────────
hdr "2. POST /api/v1/auth/mfa/confirm — Confirm first TOTP code to enable MFA"
echo "  Set MFA_CODE env var with the 6-digit code from your authenticator."

MFA_CODE="${MFA_CODE:-000000}"

http "${HTTP_FLAGS[@]}" \
  POST "${BASE_URL}/api/v1/auth/mfa/confirm" \
  "Host:${TENANT_HOST}" \
  code="${MFA_CODE}"

# Expected: 200 OK — MFA is now enabled on alice's account

# ─── 3. LOGOUT + RE-LOGIN to trigger MFA flow ────────────────────────────────
hdr "3. POST /api/v1/auth/logout — Log out alice"

http "${HTTP_FLAGS[@]}" \
  POST "${BASE_URL}/api/v1/auth/logout" \
  "Host:${TENANT_HOST}"

hdr "4. POST /api/v1/auth/login — Login triggers MFA challenge"

LOGIN_RESPONSE=$(http "${HTTP_FLAGS[@]}" \
  POST "${BASE_URL}/api/v1/auth/login" \
  "Host:${TENANT_HOST}" \
  email="alice@acme.example.com" \
  password="Str0ng!Pass#2024")

echo "$LOGIN_RESPONSE"

# Expected: 202 Accepted
#   { "mfa_required": true, "pending_token": "<short-lived-token>" }

PENDING_TOKEN=$(echo "$LOGIN_RESPONSE" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    print(d.get('pending_token', ''))
except:
    pass
" 2>/dev/null)

echo ""
echo "  → PENDING_TOKEN=${PENDING_TOKEN}"

# ─── 4. COMPLETE MFA LOGIN ────────────────────────────────────────────────────
hdr "5. POST /api/v1/auth/mfa/complete — Exchange pending token + TOTP code"
echo "  Update MFA_CODE2 with a fresh code from your authenticator."

MFA_CODE2="${MFA_CODE2:-${MFA_CODE}}"

http "${HTTP_FLAGS[@]}" \
  POST "${BASE_URL}/api/v1/auth/mfa/complete" \
  "Host:${TENANT_HOST}" \
  pending_token="${PENDING_TOKEN}" \
  code="${MFA_CODE2}"

# Expected: 200 OK → full ResolvedSession JSON + session cookie set

# ─── 5. DISABLE MFA ──────────────────────────────────────────────────────────
hdr "6. DELETE /api/v1/auth/mfa — Disable MFA (requires password re-verification)"

http "${HTTP_FLAGS[@]}" \
  DELETE "${BASE_URL}/api/v1/auth/mfa" \
  "Host:${TENANT_HOST}" \
  password="Str0ng!Pass#2024"

# Expected: 200 OK — MFA disabled

hdr "Done"
