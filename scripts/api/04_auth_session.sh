#!/usr/bin/env bash
# =============================================================================
# scripts/api/04_auth_session.sh — Login / Logout / Session validation
#
# Covers:
#   POST /api/v1/auth/login    Login → sets HttpOnly session cookie
#   POST /api/v1/auth/logout   Logout → clears cookie
#
# The session cookie is stored in SESSION_FILE by HTTPie's --session flag.
# All subsequent scripts that use --session will automatically send it.
#
# Prerequisites:
#   - Tenant "acme" must be ACTIVE
#   - User "alice@acme.example.com" must exist
#
# Usage:
#   bash scripts/api/04_auth_session.sh
# =============================================================================

source "$(dirname "$0")/00_config.sh"

# ─── 1. LOGIN ────────────────────────────────────────────────────────────────
hdr "1. POST /api/v1/auth/login — Login"

http "${HTTP_FLAGS[@]}" \
  POST "${BASE_URL}/api/v1/auth/login" \
  "Host:${TENANT_HOST}" \
  email="alice@acme.example.com" \
  password="Str0ng!Pass#2024"

# Expected:
#   200 OK  → body = ResolvedSession JSON, cookie "session" is set
#   202     → { "mfa_required": true, "pending_token": "..." }  (if MFA enabled)
#   401     → invalid credentials

# ─── 2. VERIFY SESSION — hit a protected endpoint ────────────────────────────
hdr "2. GET /api/v1/users — Verify session cookie is sent automatically"

http "${HTTP_FLAGS[@]}" \
  GET "${BASE_URL}/api/v1/users" \
  "Host:${TENANT_HOST}" \
  limit==5

# ─── 3. LOGOUT ───────────────────────────────────────────────────────────────
hdr "3. POST /api/v1/auth/logout — Logout"

http "${HTTP_FLAGS[@]}" \
  POST "${BASE_URL}/api/v1/auth/logout" \
  "Host:${TENANT_HOST}"

# Expected: 204 No Content

# ─── 4. CONFIRM SESSION IS GONE ──────────────────────────────────────────────
hdr "4. GET /api/v1/users — After logout (expect 401 on auth-protected routes)"

http "${HTTP_FLAGS[@]}" \
  GET "${BASE_URL}/api/v1/users" \
  "Host:${TENANT_HOST}" \
  limit==5

# ─── 5. RE-LOGIN for further scripts ─────────────────────────────────────────
hdr "5. POST /api/v1/auth/login — Re-login to restore session for next scripts"

http "${HTTP_FLAGS[@]}" \
  POST "${BASE_URL}/api/v1/auth/login" \
  "Host:${TENANT_HOST}" \
  email="alice@acme.example.com" \
  password="Str0ng!Pass#2024"

hdr "Done — session cookie stored in ${SESSION_FILE}"
