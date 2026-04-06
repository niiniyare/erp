#!/usr/bin/env bash
# =============================================================================
# scripts/api/11_run_all.sh — Full end-to-end test runner
#
# Runs the entire happy-path sequence in order.
# Stop on first failure (set -e).
#
# Prerequisites:
#   - Server is running: run the server manually first (see CLAUDE.md)
#   - Database is migrated and seeded
#   - Redis is running
#   - httpie is installed: pip install httpie
#
# Usage:
#   bash scripts/api/11_run_all.sh
#   BASE_URL=http://localhost:3000 bash scripts/api/11_run_all.sh
# =============================================================================

set -euo pipefail

SCRIPT_DIR="$(dirname "$0")"
source "${SCRIPT_DIR}/00_config.sh"

# ── Verify httpie is available ────────────────────────────────────────────────
if ! command -v http &>/dev/null; then
  echo "ERROR: httpie not found. Install with: pip install httpie"
  exit 1
fi

# ── Verify server is up ───────────────────────────────────────────────────────
echo ""
echo "Checking server health..."
if ! http --check-status -q GET "${BASE_URL}/health/live" &>/dev/null; then
  echo "ERROR: Server not responding at ${BASE_URL}. Start the server first."
  exit 1
fi
echo "Server is up."
echo ""

# ── Step 1: Health ────────────────────────────────────────────────────────────
hdr "STEP 1 — Health checks"
bash "${SCRIPT_DIR}/10_health.sh"

# ── Step 2: Tenant ────────────────────────────────────────────────────────────
hdr "STEP 2 — Tenant creation & lifecycle"
bash "${SCRIPT_DIR}/01_tenant.sh"

# Capture tenant ID (the script echoes it at the end)
echo ""
read -rp "Paste the TENANT_ID from above: " TENANT_ID
export TENANT_ID

# Re-activate in case it was suspended during 01_tenant.sh
http "${HTTP_PLAIN[@]}" \
  POST "${BASE_URL}/api/v1/tenants/${TENANT_ID}/activate" 2>/dev/null || true

# ── Step 3: Entities ──────────────────────────────────────────────────────────
hdr "STEP 3 — Entity tree"
bash "${SCRIPT_DIR}/02_entity.sh"

echo ""
read -rp "Paste the ROOT_ENTITY_ID from above: " ROOT_ENTITY_ID
export ROOT_ENTITY_ID

# ── Step 4: Users ─────────────────────────────────────────────────────────────
hdr "STEP 4 — User creation"
bash "${SCRIPT_DIR}/03_user.sh"

echo ""
read -rp "Paste the ADMIN_USER_ID from above: " ADMIN_USER_ID
export ADMIN_USER_ID

# ── Step 5: Auth / Session ────────────────────────────────────────────────────
hdr "STEP 5 — Auth (login / logout / session)"
bash "${SCRIPT_DIR}/04_auth_session.sh"

# ── Step 6: Password Reset ────────────────────────────────────────────────────
hdr "STEP 6 — Password reset flow"
bash "${SCRIPT_DIR}/06_password_reset.sh"

# ── Step 7: Authorization check ───────────────────────────────────────────────
hdr "STEP 7 — Authorization & permissions"
bash "${SCRIPT_DIR}/07_authz_roles_policies.sh"

# ── Step 8: API Keys ──────────────────────────────────────────────────────────
hdr "STEP 8 — API key management"
bash "${SCRIPT_DIR}/08_api_keys.sh"

echo ""
echo "┌─────────────────────────────────────────────────┐"
echo "│  All steps completed.                           │"
echo "│                                                 │"
echo "│  Optional manual steps:                         │"
echo "│    05_mfa.sh        — TOTP MFA flow             │"
echo "│    09_sso_oauth.sh  — OAuth browser flow        │"
echo "└─────────────────────────────────────────────────┘"
