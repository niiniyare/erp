#!/usr/bin/env bash
# Bootstrap seed script — creates a tenant and admin user so the server is
# usable from scratch.  Run this once after applying migrations.
#
# Requirements: curl, jq
# Usage:
#   bash scripts/seed.sh
#   BASE_URL=https://myhost ADMIN_EMAIL=me@example.com bash scripts/seed.sh

set -euo pipefail

# ─── Config ──────────────────────────────────────────────────────────────────

BASE_URL="${BASE_URL:-http://localhost:8080}"
TENANT_NAME="${TENANT_NAME:-Platform Admin}"
TENANT_EMAIL="${TENANT_EMAIL:-platform@admin.local}"
TENANT_COUNTRY="${TENANT_COUNTRY:-NG}"
TENANT_CURRENCY="${TENANT_CURRENCY:-NGN}"
ADMIN_EMAIL="${ADMIN_EMAIL:-admin@platform.local}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-Admin1234!}"
ADMIN_USERNAME="${ADMIN_USERNAME:-admin}"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

ok()   { echo -e "${GREEN}  ✓ $*${NC}"; }
fail() { echo -e "${RED}  ✗ $*${NC}"; exit 1; }
info() { echo -e "${CYAN}▶ $*${NC}"; }
sep()  { echo -e "${YELLOW}────────────────────────────────────────${NC}"; }

apiv() {
  local method="$1" path="$2"
  shift 2
  curl -s -w "\n%{http_code}" -X "$method" "${BASE_URL}${path}" \
    -H "Content-Type: application/json" "$@"
}

# ─── 1. Health check ─────────────────────────────────────────────────────────

sep
info "1. Health check"

STATUS=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/health/")
[[ "$STATUS" == "200" ]] || fail "Server not reachable — GET /health/ returned $STATUS"
ok "Server is up"

# ─── 2. Create tenant ────────────────────────────────────────────────────────

sep
info "2. Create tenant"

RESP=$(apiv POST /api/v1/tenants -d "{
  \"name\": \"$TENANT_NAME\",
  \"email\": \"$TENANT_EMAIL\",
  \"country_code\": \"$TENANT_COUNTRY\",
  \"currency_code\": \"$TENANT_CURRENCY\"
}")
BODY=$(echo "$RESP" | head -n -1)
STATUS=$(echo "$RESP" | tail -n 1)

case "$STATUS" in
  201)
    # Create response: { "data": { "id": "...", ... }, ... }
    TENANT_ID=$(echo "$BODY" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null || true)
    ok "Tenant created (201) — ID: $TENANT_ID"
    ;;
  409)
    info "  Tenant already exists (409) — listing to get ID"
    LIST_BODY=$(apiv GET /api/v1/tenants | head -n -1)
    TENANT_ID=$(echo "$LIST_BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data'][0]['id'])" 2>/dev/null || true)
    ok "Found existing tenant — ID: $TENANT_ID"
    ;;
  *)
    fail "POST /api/v1/tenants → $STATUS: $BODY"
    ;;
esac

[[ -n "$TENANT_ID" && "$TENANT_ID" != "null" ]] || fail "Could not extract tenant ID"

# ─── 3. Activate tenant ──────────────────────────────────────────────────────

sep
info "3. Activate tenant ($TENANT_ID)"

RESP=$(apiv POST "/api/v1/tenants/$TENANT_ID/activate")
BODY=$(echo "$RESP" | head -n -1)
STATUS=$(echo "$RESP" | tail -n 1)

case "$STATUS" in
  200|204)
    ok "Tenant activated"
    ;;
  409|422)
    ok "Tenant already active (skipping)"
    ;;
  *)
    fail "POST /api/v1/tenants/$TENANT_ID/activate → $STATUS: $BODY"
    ;;
esac

# ─── 3.5. Create root entity ──────────────────────────────────────────────────

sep
info "3.5. Create root entity"

RESP=$(curl -s -w "\n%{http_code}" \
  -X POST "${BASE_URL}/api/v1/entities" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d "{
    \"name\": \"$TENANT_NAME\",
    \"code\": \"ROOT\",
    \"type\": \"COMPANY\",
    \"is_active\": true
  }")
BODY=$(echo "$RESP" | head -n -1)
STATUS=$(echo "$RESP" | tail -n 1)

ENTITY_ID=""
case "$STATUS" in
  201)
    ENTITY_ID=$(echo "$BODY" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null || true)
    ok "Root entity created (201) — ID: $ENTITY_ID"
    ;;
  409)
    info "  Entity already exists (409) — fetching existing"
    ENT_RESP=$(curl -s -w "\n%{http_code}" -X GET "${BASE_URL}/api/v1/entities" \
      -H "X-Tenant-ID: $TENANT_ID")
    ENT_BODY=$(echo "$ENT_RESP" | head -n -1)
    ENTITY_ID=$(echo "$ENT_BODY" | python3 -c "import sys,json; print(json.load(sys.stdin)['data'][0]['id'])" 2>/dev/null || true)
    ok "Found existing entity — ID: $ENTITY_ID"
    ;;
  *)
    fail "POST /api/v1/entities → $STATUS: $BODY"
    ;;
esac

[[ -n "$ENTITY_ID" && "$ENTITY_ID" != "null" ]] || fail "Could not extract entity ID"

# ─── 4. Create admin user ────────────────────────────────────────────────────

sep
info "4. Create admin user ($ADMIN_EMAIL)"

RESP=$(curl -s -w "\n%{http_code}" \
  -X POST "${BASE_URL}/api/v1/users" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d "{
    \"entity_id\": \"$ENTITY_ID\",
    \"username\": \"$ADMIN_USERNAME\",
    \"email\": \"$ADMIN_EMAIL\",
    \"display_name\": \"Platform Admin\",
    \"password\": \"$ADMIN_PASSWORD\",
    \"user_type\": \"INTERNAL\",
    \"account_status\": \"ACTIVE\"
  }")
BODY=$(echo "$RESP" | head -n -1)
STATUS=$(echo "$RESP" | tail -n 1)

case "$STATUS" in
  201)
    USER_ID=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',d).get('id','?'))" 2>/dev/null || true)
    ok "Admin user created (201) — ID: $USER_ID"
    ;;
  409)
    ok "Admin user already exists (409) — skipping"
    ;;
  *)
    fail "POST /api/v1/users → $STATUS: $BODY"
    ;;
esac

# ─── 5. Verify login ─────────────────────────────────────────────────────────

sep
info "5. Verify login"

RESP=$(curl -s -w "\n%{http_code}" \
  -X POST "${BASE_URL}/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d "{\"email\": \"$ADMIN_EMAIL\", \"password\": \"$ADMIN_PASSWORD\"}")
BODY=$(echo "$RESP" | head -n -1)
STATUS=$(echo "$RESP" | tail -n 1)

case "$STATUS" in
  200)
    MFA=$(echo "$BODY" | python3 -c "import sys,json; print(json.load(sys.stdin).get('mfa_required',False))" 2>/dev/null || echo false)
    PERMS=$(echo "$BODY" | python3 -c "import sys,json; print(len(json.load(sys.stdin).get('Permissions',{})))" 2>/dev/null || echo 0)
    ok "Login successful — mfa_required: $MFA, permissions: $PERMS keys"
    ;;
  202)
    PENDING=$(echo "$BODY" | python3 -c "import sys,json; print(json.load(sys.stdin).get('pending_token',''))" 2>/dev/null || true)
    ok "Login OK — MFA required (pending_token: ${PENDING:0:12}…)"
    ;;
  401)
    fail "Login failed 401 — credentials rejected. Check password hash or user status."
    ;;
  *)
    fail "POST /api/v1/auth/login → $STATUS: $BODY"
    ;;
esac

# ─── Summary ─────────────────────────────────────────────────────────────────

sep
echo -e "${GREEN}Seed complete.${NC}"
echo ""
echo "  Tenant  : $TENANT_NAME ($TENANT_ID)"
echo "  Admin   : $ADMIN_EMAIL"
echo "  Password: $ADMIN_PASSWORD"
echo ""
echo "  Run the full test suite:"
echo "    TENANT_ID=$TENANT_ID ADMIN_EMAIL=$ADMIN_EMAIL ADMIN_PASSWORD=$ADMIN_PASSWORD bash scripts/curl_test.sh"
