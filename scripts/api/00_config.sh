#!/usr/bin/env bash
# =============================================================================
# scripts/api/00_config.sh — Shared configuration for all API test scripts
# Source this file in every script: source "$(dirname "$0")/00_config.sh"
# =============================================================================

# Base URL of the running server
BASE_URL="${BASE_URL:-http://localhost:8080}"

# HTTPie session file directory (persists cookies across calls)
SESSION_DIR="$(dirname "$0")/data"
mkdir -p "$SESSION_DIR"

# Tenant subdomain used for all tenant-scoped requests.
# The ResolveTenant middleware reads the Host header subdomain.
TENANT_SUBDOMAIN="${TENANT_SUBDOMAIN:-acme}"
TENANT_HOST="${TENANT_SUBDOMAIN}.localhost:8080"

# HTTPie session file — stores the session cookie between calls
SESSION_FILE="${SESSION_DIR}/session_${TENANT_SUBDOMAIN}.jar"

# Pretty-print helpers
hr()  { printf '\n%s\n' "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; }
hdr() { hr; printf '  %s\n' "$1"; hr; }

# Shared HTTPie flags
# --session     persists cookies (session cookie) across calls
# -p hbHB       print request+response headers and body
HTTP_FLAGS=(--session="$SESSION_FILE" -p hbHB)

# No-session variant (for platform-level calls that don't need a cookie)
HTTP_PLAIN=(-p hbHB)

echo "  BASE_URL : $BASE_URL"
echo "  HOST     : $TENANT_HOST"
echo "  SESSION  : $SESSION_FILE"
