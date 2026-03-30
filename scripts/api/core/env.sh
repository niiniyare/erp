#!/usr/bin/env bash
# Core environment — sourced by all scripts.
# Sets shared config and generates unique per-session random values.

BASE_URL="${BASE_URL:-http://localhost:8080}"
ADMIN_EMAIL="${ADMIN_EMAIL:-admin@platform.local}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-Admin1234!}"
TENANT_ID="${TENANT_ID:-}"

# Paths (early so the cache check below can reference DATA_DIR)
SCRIPTS_API_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DATA_DIR="${SCRIPTS_API_DIR}/data"
LOG_DIR="${SCRIPTS_API_DIR}/logs"
mkdir -p "$DATA_DIR" "$LOG_DIR"

# If TENANT_ID was supplied via env, persist it for future runs
_TENANT_CACHE="${DATA_DIR}/tenant_id"
if [[ -n "$TENANT_ID" ]]; then
  echo "$TENANT_ID" > "$_TENANT_CACHE" 2>/dev/null || true
elif [[ -f "$_TENANT_CACHE" ]]; then
  # Load previously saved value
  _cached=$(cat "$_TENANT_CACHE" 2>/dev/null | tr -d '[:space:]')
  [[ -n "$_cached" && "$_cached" != "null" ]] && TENANT_ID="$_cached"
fi

# Per-session unique ID — 8 hex chars, e.g. "a3f8b2c1"
SESSION_ID="$(printf '%04x%04x' $RANDOM $RANDOM)"
TEST_SUFFIX="${SESSION_ID:0:6}"

# Test data — unique per session to avoid collisions between runs
TEST_EMAIL="tuser_${TEST_SUFFIX}@test.local"
TEST_PASSWORD="Tst_${TEST_SUFFIX}1!"
TEST_USERNAME="tuser_${TEST_SUFFIX}"
TEST_TENANT_NAME="TestCo_${TEST_SUFFIX}"
TEST_TENANT_EMAIL="admin_${TEST_SUFFIX}@testco.local"

# Cookie jar — per session, removed on EXIT
COOKIE_JAR="${DATA_DIR}/session_${SESSION_ID}.jar"

# Session state (managed by auth.sh)
AUTH_ACTIVE=false
AUTH_USER_EMAIL=""
