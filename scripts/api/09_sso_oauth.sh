#!/usr/bin/env bash
# =============================================================================
# scripts/api/09_sso_oauth.sh — OAuth / OIDC SSO flow
#
# Covers:
#   GET /api/v1/auth/oauth/:provider           Begin OAuth (returns redirect URL)
#   GET /api/v1/auth/oauth/:provider/callback  Callback (exchange code → session)
#
# Supported providers: google, microsoft
#
# Flow:
#   1. BeginOAuth  → server generates CSRF state, stores in Redis, returns URL
#   2. User visits the URL, authenticates with Google/Microsoft
#   3. Provider redirects to /callback?code=...&state=...
#   4. Server validates state, exchanges code, fetches userinfo, creates session
#
# Note: The callback step requires a real OAuth code from a browser flow.
# This script tests step 1 only and shows the redirect URL.
# For full end-to-end testing, paste the callback URL into a browser.
#
# Prerequisites:
#   - SSOService must be configured (EncryptionKey set, provider upserted in DB)
#   - tenant "acme" must have a google or microsoft SSO provider configured
#
# Usage:
#   PROVIDER=google bash scripts/api/09_sso_oauth.sh
# =============================================================================

source "$(dirname "$0")/00_config.sh"

PROVIDER="${PROVIDER:-google}"
TENANT_ID="${TENANT_ID:-<paste-tenant-uuid>}"

# ─── 1. BEGIN OAUTH ──────────────────────────────────────────────────────────
hdr "1. GET /api/v1/auth/oauth/${PROVIDER} — Begin OAuth flow"
echo "  Query param tenant_id scopes the state to this tenant."
echo "  Server stores CSRF state in Redis (10-min TTL, single-use)."

http "${HTTP_PLAIN[@]}" \
  GET "${BASE_URL}/api/v1/auth/oauth/${PROVIDER}" \
  "Host:${TENANT_HOST}" \
  tenant_id=="${TENANT_ID}"

# Expected: 302 redirect → Location: https://accounts.google.com/o/oauth2/v2/auth?...
# OR: 200 with { "auth_url": "https://..." } depending on handler implementation

echo ""
echo "  ⚠  Copy the Location / auth_url and open it in a browser."
echo "     After authenticating, copy the full callback URL (with code= and state=)."
echo "     Set CODE and STATE below, then run step 2."

# ─── 2. CALLBACK (manual — paste code+state from browser) ────────────────────
hdr "2. GET /api/v1/auth/oauth/${PROVIDER}/callback — Exchange code for session"
echo "  Set CODE and STATE from the browser redirect URL."

CODE="${CODE:-}"
STATE="${STATE:-}"

if [ -n "$CODE" ] && [ -n "$STATE" ]; then
  http "${HTTP_FLAGS[@]}" \
    GET "${BASE_URL}/api/v1/auth/oauth/${PROVIDER}/callback" \
    "Host:${TENANT_HOST}" \
    code=="${CODE}" \
    state=="${STATE}"

  # Expected: 200 OK → ResolvedSession + session cookie set
  # If provider.auto_provision=true and user is new → JIT provisioned as CUSTOMER
else
  echo "  → CODE and STATE not set — browser flow required."
  echo "     After getting them from the redirect URL, run:"
  echo "     CODE=... STATE=... PROVIDER=${PROVIDER} bash scripts/api/09_sso_oauth.sh"
fi

# ─── 3. SIMULATE: Microsoft ──────────────────────────────────────────────────
hdr "3. GET /api/v1/auth/oauth/microsoft — Begin Microsoft OAuth flow"

http "${HTTP_PLAIN[@]}" \
  GET "${BASE_URL}/api/v1/auth/oauth/microsoft" \
  "Host:${TENANT_HOST}" \
  tenant_id=="${TENANT_ID}"

hdr "Done"
echo "  SSO login creates a full session — use the cookie for further requests."
echo "  MFA is skipped for SSO users (IdP is the second factor)."
