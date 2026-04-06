#!/usr/bin/env bash
# =============================================================================
# scripts/api/07_authz_roles_policies.sh — Authorization: roles & policies
#
# The AuthzService is consumed directly by the IAM internals (session build,
# middleware). There are no dedicated /roles or /policies REST endpoints yet —
# they are managed through the service layer. This script tests authorization
# indirectly through:
#   - The session ResolvedSession (permissions are pre-computed at login)
#   - The finance endpoints (require finance.accounts.read permission)
#   - The audit-log endpoint (requires active session)
#   - Schema boot endpoint (permission-filtered navigation)
#
# If you add admin endpoints for role/policy management later, add them here.
#
# Covers:
#   GET /api/v1/schema/boot        Boot schema (permission-filtered nav)
#   GET /api/v1/audit-logs         Audit log (requires session)
#   GET /api/v1/finance/accounts   Finance (requires session + permission)
#
# Prerequisites:
#   - alice is logged in (session from 04_auth_session.sh)
#
# Usage:
#   bash scripts/api/07_authz_roles_policies.sh
# =============================================================================

source "$(dirname "$0")/00_config.sh"

# Ensure alice is logged in
hdr "0. POST /api/v1/auth/login — Ensure alice session is active"

http "${HTTP_FLAGS[@]}" \
  POST "${BASE_URL}/api/v1/auth/login" \
  "Host:${TENANT_HOST}" \
  email="alice@acme.example.com" \
  password="Str0ng!Pass#2024"

# ─── 1. BOOT SCHEMA ──────────────────────────────────────────────────────────
hdr "1. GET /api/v1/schema/boot — AMIS app boot (permission-filtered navigation)"
echo "  Returns the navigation tree filtered by the session's permissions."

http "${HTTP_FLAGS[@]}" \
  GET "${BASE_URL}/api/v1/schema/boot" \
  "Host:${TENANT_HOST}"

# ─── 2. AUDIT LOG ────────────────────────────────────────────────────────────
hdr "2. GET /api/v1/audit-logs — Audit log (requires active session)"

http "${HTTP_FLAGS[@]}" \
  GET "${BASE_URL}/api/v1/audit-logs" \
  "Host:${TENANT_HOST}" \
  limit==20 \
  offset==0

# ─── 3. FINANCE — AUTHORIZED ACCESS ─────────────────────────────────────────
hdr "3. GET /api/v1/finance/accounts — Authorized (alice has wildcard tenant_admin)"

http "${HTTP_FLAGS[@]}" \
  GET "${BASE_URL}/api/v1/finance/accounts" \
  "Host:${TENANT_HOST}"

# ─── 4. FINANCE — UNAUTHORIZED (bob without roles) ───────────────────────────
hdr "4. GET /api/v1/finance/accounts — With bob's session (no finance permission)"
echo "  First log bob in to a separate session file."

BOB_SESSION="${SESSION_DIR}/session_bob.jar"

http --session="${BOB_SESSION}" -p hbHB \
  POST "${BASE_URL}/api/v1/auth/login" \
  "Host:${TENANT_HOST}" \
  email="bob@acme.example.com" \
  password="N3wStr0ng!Reset#2025"

echo ""
echo "  Now try finance with bob's session:"

http --session="${BOB_SESSION}" -p hbHB \
  GET "${BASE_URL}/api/v1/finance/accounts" \
  "Host:${TENANT_HOST}"

# Expected: 403 Forbidden — bob has no finance.accounts.read permission

# ─── 5. API KEY — BEARER AUTH (alternative to cookie) ────────────────────────
hdr "5. GET /api/v1/users — Using Bearer token instead of cookie"
echo "  Set API_KEY_TOKEN from script 08_api_keys.sh output."

API_KEY_TOKEN="${API_KEY_TOKEN:-}"
if [ -n "$API_KEY_TOKEN" ]; then
  http -p hbHB \
    GET "${BASE_URL}/api/v1/users" \
    "Host:${TENANT_HOST}" \
    "Authorization:Bearer ${API_KEY_TOKEN}"
else
  echo "  → API_KEY_TOKEN not set — run 08_api_keys.sh first."
fi

hdr "Done"
