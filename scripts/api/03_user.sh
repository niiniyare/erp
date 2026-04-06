#!/usr/bin/env bash
# =============================================================================
# scripts/api/03_user.sh — User CRUD + password change
#
# Covers:
#   POST   /api/v1/users                    Create (RegisterNewUser)
#   GET    /api/v1/users                    List
#   GET    /api/v1/users/:id                Get (default / detailed / security / public views)
#   PUT    /api/v1/users/:id                Update
#   DELETE /api/v1/users/:id                Delete (soft)
#   POST   /api/v1/users/authenticate       Verify credentials (raw — not login)
#   POST   /api/v1/users/:id/change-password
#
# Prerequisites:
#   - Tenant "acme" must be ACTIVE
#   - An entity must exist (ROOT_ENTITY_ID)
#
# Usage:
#   ROOT_ENTITY_ID=<uuid> bash scripts/api/03_user.sh
# =============================================================================

source "$(dirname "$0")/00_config.sh"

ROOT_ENTITY_ID="${ROOT_ENTITY_ID:-<paste-entity-uuid>}"

# ─── 1. CREATE ADMIN USER ────────────────────────────────────────────────────
hdr "1. POST /api/v1/users — Create admin user"

ADMIN_RESPONSE=$(http "${HTTP_PLAIN[@]}" \
  POST "${BASE_URL}/api/v1/users" \
  "Host:${TENANT_HOST}" \
  entity_id="${ROOT_ENTITY_ID}" \
  email="alice@acme.example.com" \
  username="alice" \
  password="Str0ng!Pass#2024" \
  user_type="INTERNAL" \
  display_name="Alice Admin")

echo "$ADMIN_RESPONSE"

ADMIN_USER_ID=$(echo "$ADMIN_RESPONSE" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    print(d.get('data', {}).get('id', ''))
except:
    pass
" 2>/dev/null)

echo ""
echo "  → ADMIN_USER_ID=${ADMIN_USER_ID}"

# ─── 2. CREATE REGULAR USER ──────────────────────────────────────────────────
hdr "2. POST /api/v1/users — Create regular user"

USER_RESPONSE=$(http "${HTTP_PLAIN[@]}" \
  POST "${BASE_URL}/api/v1/users" \
  "Host:${TENANT_HOST}" \
  entity_id="${ROOT_ENTITY_ID}" \
  email="bob@acme.example.com" \
  username="bob" \
  password="Str0ng!Pass#2024" \
  user_type="INTERNAL" \
  display_name="Bob User")

echo "$USER_RESPONSE"

BOB_USER_ID=$(echo "$USER_RESPONSE" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    print(d.get('data', {}).get('id', ''))
except:
    pass
" 2>/dev/null)

echo ""
echo "  → BOB_USER_ID=${BOB_USER_ID}"

# ─── 3. LIST USERS ───────────────────────────────────────────────────────────
hdr "3. GET /api/v1/users — List all users"

http "${HTTP_PLAIN[@]}" \
  GET "${BASE_URL}/api/v1/users" \
  "Host:${TENANT_HOST}" \
  limit==20 \
  offset==0

# ─── 3b. LIST USERS — filtered by type ───────────────────────────────────────
hdr "3b. GET /api/v1/users — Filter by user_type=INTERNAL"

http "${HTTP_PLAIN[@]}" \
  GET "${BASE_URL}/api/v1/users" \
  "Host:${TENANT_HOST}" \
  user_type==INTERNAL \
  status==ACTIVE

# ─── 4. GET USER — default view ──────────────────────────────────────────────
hdr "4. GET /api/v1/users/:id — Default view"

http "${HTTP_PLAIN[@]}" \
  GET "${BASE_URL}/api/v1/users/${ADMIN_USER_ID}" \
  "Host:${TENANT_HOST}"

# ─── 4b. GET USER — detailed view ────────────────────────────────────────────
hdr "4b. GET /api/v1/users/:id?view=detailed"

http "${HTTP_PLAIN[@]}" \
  GET "${BASE_URL}/api/v1/users/${ADMIN_USER_ID}" \
  "Host:${TENANT_HOST}" \
  view==detailed

# ─── 4c. GET USER — security view ────────────────────────────────────────────
hdr "4c. GET /api/v1/users/:id?view=security"

http "${HTTP_PLAIN[@]}" \
  GET "${BASE_URL}/api/v1/users/${ADMIN_USER_ID}" \
  "Host:${TENANT_HOST}" \
  view==security

# ─── 5. UPDATE USER ──────────────────────────────────────────────────────────
hdr "5. PUT /api/v1/users/:id — Update display name and timezone"

http "${HTTP_PLAIN[@]}" \
  PUT "${BASE_URL}/api/v1/users/${BOB_USER_ID}" \
  "Host:${TENANT_HOST}" \
  display_name="Robert User" \
  timezone="America/Los_Angeles" \
  language="en-US"

# ─── 6. AUTHENTICATE (raw credential check) ──────────────────────────────────
hdr "6. POST /api/v1/users/authenticate — Verify credentials (no session)"

http "${HTTP_PLAIN[@]}" \
  POST "${BASE_URL}/api/v1/users/authenticate" \
  "Host:${TENANT_HOST}" \
  identifier="alice@acme.example.com" \
  password="Str0ng!Pass#2024"

# ─── 7. CHANGE PASSWORD ──────────────────────────────────────────────────────
hdr "7. POST /api/v1/users/:id/change-password"

http "${HTTP_PLAIN[@]}" \
  POST "${BASE_URL}/api/v1/users/${BOB_USER_ID}/change-password" \
  "Host:${TENANT_HOST}" \
  current_password="Str0ng!Pass#2024" \
  new_password="N3wStr0ng!Pass#2025"

# ─── 8. DELETE USER (soft) ───────────────────────────────────────────────────
# WARNING: soft-deletes the user — comment out to keep for auth tests
#
# hdr "8. DELETE /api/v1/users/:id — Soft delete"
# http "${HTTP_PLAIN[@]}" \
#   DELETE "${BASE_URL}/api/v1/users/${BOB_USER_ID}" \
#   "Host:${TENANT_HOST}"

hdr "Done"
echo "  Export for next scripts:"
echo "  export ADMIN_USER_ID=${ADMIN_USER_ID}"
echo "  export BOB_USER_ID=${BOB_USER_ID}"
