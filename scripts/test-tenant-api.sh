#!/bin/bash
# Tenant API test script (randomized data)
# Usage: bash scripts/test-tenant-api.sh [base_url]

BASE="${1:-http://localhost:8080}"
API="$BASE/api/v1/tenants"
ONBOARD="$BASE/api/v1/orgs/onboard"

GREEN='\033[0;32m'
CYAN='\033[0;36m'
NC='\033[0m'

log() { echo -e "\n${CYAN}=== $1 ===${NC}"; }
hr() { echo "---"; }

# -------------------------------------------------------
# Random helpers
# -------------------------------------------------------
RAND_SUFFIX="$(date +%s)-$RANDOM"

rand_from() {
  local arr=("$@")
  echo "${arr[RANDOM % ${#arr[@]}]}"
}

ORG_NAMES=("Acme" "Globex" "Initech" "Umbrella" "Soylent" "Hooli" "Stark" "Wayne")
INDUSTRIES=("Technology" "Finance" "Energy" "Healthcare" "Retail" "Logistics")
SIZES=("SMALL" "MEDIUM" "LARGE")
COUNTRIES=("US" "KE" "GB")
CURRENCIES=("USD" "KES" "GBP")

ORG_NAME="$(rand_from "${ORG_NAMES[@]}") Corp $RAND_SUFFIX"
SLUG="$(echo "$ORG_NAME" | tr '[:upper:] ' '[:lower:]-' | tr -cd 'a-z0-9-')"
EMAIL="admin+$RAND_SUFFIX@example.test"
INDUSTRY="$(rand_from "${INDUSTRIES[@]}")"
SIZE="$(rand_from "${SIZES[@]}")"
COUNTRY="$(rand_from "${COUNTRIES[@]}")"
CURRENCY="$(rand_from "${CURRENCIES[@]}")"

# -------------------------------------------------------
# 1. Create Organization
# -------------------------------------------------------
log "POST /tenants — Create (random data)"

CREATE_RESP=$(curl -s -w "\n%{http_code}" -X POST "$API" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"$ORG_NAME\",
    \"slug\": \"$SLUG\",
    \"email\": \"$EMAIL\",
    \"country_code\": \"$COUNTRY\",
    \"currency_code\": \"$CURRENCY\",
    \"industry\": \"$INDUSTRY\",
    \"company_size\": \"$SIZE\"
  }")

HTTP_CODE=$(echo "$CREATE_RESP" | tail -1)
BODY=$(echo "$CREATE_RESP" | sed '$d')

echo "Status: $HTTP_CODE"
echo "$BODY" | python3 -m json.tool 2>/dev/null || echo "$BODY"

TENANT_ID=$(echo "$BODY" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null)

if [ -z "$TENANT_ID" ]; then
  echo "Failed to extract tenant ID. Exiting."
  exit 1
fi

echo -e "${GREEN}Tenant ID: $TENANT_ID${NC}"
hr

# -------------------------------------------------------
# 2–6. Read APIs
# -------------------------------------------------------
log "GET /tenants/:id — Default view"
curl -s "$API/$TENANT_ID" | python3 -m json.tool 2>/dev/null
hr

log "GET /tenants/:id?view=detailed"
curl -s "$API/$TENANT_ID?view=detailed" | python3 -m json.tool 2>/dev/null
hr

log "GET /tenants/:id?view=summary"
curl -s "$API/$TENANT_ID?view=summary" | python3 -m json.tool 2>/dev/null
hr

log "GET /tenants — List"
curl -s "$API?limit=10&offset=0" | python3 -m json.tool 2>/dev/null
hr

log "GET /tenants — Filtered"
curl -s "$API?status=ACTIVE&limit=5&sort_by=name" | python3 -m json.tool 2>/dev/null
hr

# -------------------------------------------------------
# 7. Update (PUT)
# -------------------------------------------------------
log "PUT /tenants/:id — Update"

curl -s -X PUT "$API/$TENANT_ID" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"$ORG_NAME Updated\",
    \"email\": \"contact+$RAND_SUFFIX@example.test\"
  }" | python3 -m json.tool 2>/dev/null
hr

# -------------------------------------------------------
# 8. Update (PATCH)
# -------------------------------------------------------
log "PATCH /tenants/:id — Partial Update"

curl -s -X PATCH "$API/$TENANT_ID" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"$ORG_NAME Intl\"
  }" | python3 -m json.tool 2>/dev/null
hr

# -------------------------------------------------------
# 9–11. Lifecycle actions
# -------------------------------------------------------
log "POST /tenants/:id/suspend"
curl -s -X POST "$API/$TENANT_ID/suspend" \
  -H "Content-Type: application/json" \
  -d "{\"reason\": \"Automated test suspend $RAND_SUFFIX\"}" |
  python3 -m json.tool 2>/dev/null
hr

log "POST /tenants/:id/activate"
curl -s -X POST "$API/$TENANT_ID/activate" \
  -H "Content-Type: application/json" |
  python3 -m json.tool 2>/dev/null
hr

log "POST /tenants/:id/archive"
curl -s -X POST "$API/$TENANT_ID/archive" \
  -H "Content-Type: application/json" |
  python3 -m json.tool 2>/dev/null
hr

# -------------------------------------------------------
# 12. Onboard
# -------------------------------------------------------
log "POST /orgs/onboard — Random"

curl -s -w "\nHTTP Status: %{http_code}\n" -X POST "$ONBOARD" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"OnboardCo $RAND_SUFFIX\",
    \"email\": \"onboard+$RAND_SUFFIX@example.test\"
  }"
hr

# -------------------------------------------------------
# 13. Delete throwaway org
# -------------------------------------------------------
log "DELETE /tenants/:id — Throwaway"

DEL_SUFFIX="del-$RAND_SUFFIX"
DEL_RESP=$(curl -s -X POST "$API" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"Throwaway $DEL_SUFFIX\",
    \"slug\": \"throwaway-$DEL_SUFFIX\",
    \"email\": \"del+$DEL_SUFFIX@example.test\",
    \"country_code\": \"US\",
    \"currency_code\": \"USD\"
  }")

DEL_ID=$(echo "$DEL_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null)

if [ -n "$DEL_ID" ]; then
  echo "Deleting tenant: $DEL_ID"
  curl -s -w "HTTP Status: %{http_code}\n" -X DELETE "$API/$DEL_ID"
else
  echo "Skipped — could not create throwaway tenant"
fi
hr

# -------------------------------------------------------
# 14. Error cases
# -------------------------------------------------------
log "Error: Invalid UUID"
curl -s "$API/not-a-uuid" | python3 -m json.tool 2>/dev/null
hr

log "Error: Missing required fields"
curl -s -X POST "$API" \
  -H "Content-Type: application/json" \
  -d '{"name": ""}' | python3 -m json.tool 2>/dev/null
hr

log "Error: Suspend without reason"
curl -s -X POST "$API/$TENANT_ID/suspend" \
  -H "Content-Type: application/json" \
  -d '{}' | python3 -m json.tool 2>/dev/null
hr

log "Error: Activate archived tenant"
curl -s -X POST "$API/$TENANT_ID/activate" \
  -H "Content-Type: application/json" |
  python3 -m json.tool 2>/dev/null
hr

echo -e "\n${GREEN}Done.${NC}"
