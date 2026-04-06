#!/usr/bin/env bash
# =============================================================================
# scripts/api/01_tenant.sh — Tenant CRUD + lifecycle
#
# Covers:
#   POST   /api/v1/tenants              Create
#   GET    /api/v1/tenants              List
#   GET    /api/v1/tenants/:id          Get by ID
#   PUT    /api/v1/tenants/:id          Update
#   POST   /api/v1/tenants/:id/activate Activate (PENDING → ACTIVE)
#   POST   /api/v1/tenants/:id/suspend  Suspend  (ACTIVE → SUSPENDED)
#   POST   /api/v1/tenants/:id/archive  Archive  (terminal)
#   DELETE /api/v1/tenants/:id          Delete
#   POST   /api/v1/tenants/onboard      Async onboarding (Temporal)
#
# Usage:
#   bash scripts/api/01_tenant.sh
#   TENANT_ID=<uuid> bash scripts/api/01_tenant.sh   # skip create, use existing
# =============================================================================

source "$(dirname "$0")/00_config.sh"

# ─── 1. CREATE ───────────────────────────────────────────────────────────────
hdr "1. POST /api/v1/tenants — Create"

CREATE_RESPONSE=$(http "${HTTP_PLAIN[@]}" \
  POST "${BASE_URL}/api/v1/tenants" \
  name="Acme Corporation" \
  email="admin@acme.example.com" \
  subdomain="acme" \
  country_code="US" \
  currency_code="USD" \
  timezone="America/New_York" \
  industry="technology" \
  company_size="MEDIUM" \
  plan_tier="professional" \
  tax_id="12-3456789" \
  legal_entity_type="LLC")

echo "$CREATE_RESPONSE"

# Extract tenant ID from response for subsequent calls
TENANT_ID=$(echo "$CREATE_RESPONSE" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    print(d.get('data', {}).get('id', d.get('id', '')))
except:
    pass
" 2>/dev/null)

if [ -z "$TENANT_ID" ]; then
  echo "⚠  Could not extract TENANT_ID from response — set it manually."
  TENANT_ID="${TENANT_ID:-<paste-uuid-here>}"
fi

echo ""
echo "  → TENANT_ID=$TENANT_ID"

# ─── 2. LIST ─────────────────────────────────────────────────────────────────
hdr "2. GET /api/v1/tenants — List (page 1)"

http "${HTTP_PLAIN[@]}" \
  GET "${BASE_URL}/api/v1/tenants" \
  limit==20 \
  offset==0

# ─── 3. GET BY ID ────────────────────────────────────────────────────────────
hdr "3. GET /api/v1/tenants/:id — Get by ID"

http "${HTTP_PLAIN[@]}" \
  GET "${BASE_URL}/api/v1/tenants/${TENANT_ID}"

# ─── 4. UPDATE ───────────────────────────────────────────────────────────────
hdr "4. PUT /api/v1/tenants/:id — Update"

http "${HTTP_PLAIN[@]}" \
  PUT "${BASE_URL}/api/v1/tenants/${TENANT_ID}" \
  name="Acme Corp (Updated)" \
  email="admin@acme.example.com" \
  timezone="UTC" \
  company_size="LARGE"

# ─── 5. ACTIVATE ─────────────────────────────────────────────────────────────
hdr "5. POST /api/v1/tenants/:id/activate — PENDING → ACTIVE"

http "${HTTP_PLAIN[@]}" \
  POST "${BASE_URL}/api/v1/tenants/${TENANT_ID}/activate"

# ─── 6. SUSPEND ──────────────────────────────────────────────────────────────
hdr "6. POST /api/v1/tenants/:id/suspend — ACTIVE → SUSPENDED"

http "${HTTP_PLAIN[@]}" \
  POST "${BASE_URL}/api/v1/tenants/${TENANT_ID}/suspend" \
  reason="Pending payment review"

# ─── 7. RE-ACTIVATE ──────────────────────────────────────────────────────────
hdr "7. POST /api/v1/tenants/:id/activate — Re-activate after suspend"

http "${HTTP_PLAIN[@]}" \
  POST "${BASE_URL}/api/v1/tenants/${TENANT_ID}/activate"

# ─── 8. ARCHIVE (terminal) ───────────────────────────────────────────────────
# WARNING: terminal state — comment out if you need the tenant for further tests
#
# hdr "8. POST /api/v1/tenants/:id/archive — ACTIVE → ARCHIVED (terminal)"
# http "${HTTP_PLAIN[@]}" \
#   POST "${BASE_URL}/api/v1/tenants/${TENANT_ID}/archive"

# ─── 9. DELETE ───────────────────────────────────────────────────────────────
# WARNING: destructive — comment out to keep the tenant alive for further tests
#
# hdr "9. DELETE /api/v1/tenants/:id — Soft-delete"
# http "${HTTP_PLAIN[@]}" \
#   DELETE "${BASE_URL}/api/v1/tenants/${TENANT_ID}"

# ─── 10. ASYNC ONBOARD (Temporal) ────────────────────────────────────────────
hdr "10. POST /api/v1/tenants/onboard — Async workflow (requires Temporal)"
echo "  Note: returns 503 if Temporal is not configured."

http "${HTTP_PLAIN[@]}" \
  POST "${BASE_URL}/api/v1/tenants/onboard" \
  name="Beta Corp" \
  email="hello@betacorp.example.com" \
  country_code="GB" \
  currency_code="GBP" \
  subdomain="betacorp" \
  industry="finance" \
  company_size="SMALL"

hdr "Done — TENANT_ID=${TENANT_ID}"
echo "  Export for next scripts:"
echo "  export TENANT_ID=${TENANT_ID}"
echo "  export TENANT_SUBDOMAIN=acme"
