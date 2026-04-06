#!/usr/bin/env bash
# =============================================================================
# scripts/api/06_password_reset.sh — Forgot / Reset password flow
#
# Covers:
#   POST /api/v1/auth/forgot-password  Request reset token (always 200)
#   POST /api/v1/auth/reset-password   Use token to set new password
#
# Security notes from the code:
#   - ForgotPassword ALWAYS returns 200 even if email not found (anti-enumeration)
#   - Token is single-use and has a short TTL
#   - In a real integration the token would be emailed; here we log it from
#     the server output for test purposes
#
# Usage:
#   bash scripts/api/06_password_reset.sh
# =============================================================================

source "$(dirname "$0")/00_config.sh"

# ─── 1. REQUEST RESET TOKEN ──────────────────────────────────────────────────
hdr "1. POST /api/v1/auth/forgot-password — Request password reset"
echo "  Always returns 200 regardless of whether email exists (anti-enumeration)."

http "${HTTP_PLAIN[@]}" \
  POST "${BASE_URL}/api/v1/auth/forgot-password" \
  "Host:${TENANT_HOST}" \
  email="bob@acme.example.com"

# Expected: 200 OK → { "message": "If that email exists, a reset link was sent." }
# In development the raw token is typically logged server-side.
# Set RESET_TOKEN from the server log output.

echo ""
echo "  ⚠  Check the server log for the raw reset token and set:"
echo "     export RESET_TOKEN=<token-from-log>"

RESET_TOKEN="${RESET_TOKEN:-<paste-token-from-server-log>}"

# ─── 2. RESET PASSWORD ───────────────────────────────────────────────────────
hdr "2. POST /api/v1/auth/reset-password — Set new password using token"

http "${HTTP_PLAIN[@]}" \
  POST "${BASE_URL}/api/v1/auth/reset-password" \
  "Host:${TENANT_HOST}" \
  token="${RESET_TOKEN}" \
  new_password="N3wStr0ng!Reset#2025"

# Expected responses:
#   200  → password updated successfully
#   400  → ErrPasswordResetTokenNotFound / ErrPasswordResetTokenExpired / ErrPasswordResetTokenUsed
#   422  → ErrPasswordTooWeak / ErrPasswordReused

# ─── 3. VERIFY LOGIN WITH NEW PASSWORD ───────────────────────────────────────
hdr "3. POST /api/v1/auth/login — Login with new password to confirm reset"

http "${HTTP_FLAGS[@]}" \
  POST "${BASE_URL}/api/v1/auth/login" \
  "Host:${TENANT_HOST}" \
  email="bob@acme.example.com" \
  password="N3wStr0ng!Reset#2025"

hdr "Done"
