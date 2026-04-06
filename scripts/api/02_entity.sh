#!/usr/bin/env bash
# =============================================================================
# scripts/api/02_entity.sh — Entity creation and listing
#
# Entities form the org-unit tree that users are attached to.
# All entity endpoints require the tenant context resolved from the Host header.
#
# Covers:
#   POST /api/v1/entities   Create entity
#   GET  /api/v1/entities   List entities
#
# Prerequisites:
#   - Tenant "acme" must exist and be ACTIVE
#   - Server resolves tenant from Host: acme.localhost:8080
#
# Usage:
#   bash scripts/api/02_entity.sh
#   TENANT_SUBDOMAIN=acme bash scripts/api/02_entity.sh
# =============================================================================

source "$(dirname "$0")/00_config.sh"

# ─── 1. CREATE ROOT ENTITY ───────────────────────────────────────────────────
hdr "1. POST /api/v1/entities — Create root entity (company)"

ROOT_RESPONSE=$(http "${HTTP_PLAIN[@]}" \
  POST "${BASE_URL}/api/v1/entities" \
  "Host:${TENANT_HOST}" \
  name="Acme Corporation" \
  code="ACME-ROOT" \
  type="COMPANY" \
  description="Top-level entity for Acme Corp")

echo "$ROOT_RESPONSE"

ROOT_ENTITY_ID=$(echo "$ROOT_RESPONSE" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    print(d.get('data', {}).get('id', ''))
except:
    pass
" 2>/dev/null)

echo ""
echo "  → ROOT_ENTITY_ID=${ROOT_ENTITY_ID}"

# ─── 2. CREATE DEPARTMENT ENTITY ─────────────────────────────────────────────
hdr "2. POST /api/v1/entities — Create department (child of root)"

DEPT_RESPONSE=$(http "${HTTP_PLAIN[@]}" \
  POST "${BASE_URL}/api/v1/entities" \
  "Host:${TENANT_HOST}" \
  name="Engineering" \
  code="ACME-ENG" \
  type="DEPARTMENT" \
  description="Engineering division" \
  parent_id="${ROOT_ENTITY_ID}")

echo "$DEPT_RESPONSE"

DEPT_ENTITY_ID=$(echo "$DEPT_RESPONSE" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    print(d.get('data', {}).get('id', ''))
except:
    pass
" 2>/dev/null)

echo ""
echo "  → DEPT_ENTITY_ID=${DEPT_ENTITY_ID}"

# ─── 3. CREATE TEAM ENTITY ───────────────────────────────────────────────────
hdr "3. POST /api/v1/entities — Create team (child of department)"

TEAM_RESPONSE=$(http "${HTTP_PLAIN[@]}" \
  POST "${BASE_URL}/api/v1/entities" \
  "Host:${TENANT_HOST}" \
  name="Platform Team" \
  code="ACME-ENG-PLATFORM" \
  type="TEAM" \
  description="Platform engineering team" \
  parent_id="${DEPT_ENTITY_ID}")

echo "$TEAM_RESPONSE"

TEAM_ENTITY_ID=$(echo "$TEAM_RESPONSE" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    print(d.get('data', {}).get('id', ''))
except:
    pass
" 2>/dev/null)

echo ""
echo "  → TEAM_ENTITY_ID=${TEAM_ENTITY_ID}"

# ─── 4. LIST ENTITIES ────────────────────────────────────────────────────────
hdr "4. GET /api/v1/entities — List all entities for tenant"

http "${HTTP_PLAIN[@]}" \
  GET "${BASE_URL}/api/v1/entities" \
  "Host:${TENANT_HOST}"

hdr "Done"
echo "  Export for next scripts:"
echo "  export ROOT_ENTITY_ID=${ROOT_ENTITY_ID}"
echo "  export DEPT_ENTITY_ID=${DEPT_ENTITY_ID}"
echo "  export TEAM_ENTITY_ID=${TEAM_ENTITY_ID}"
