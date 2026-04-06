#!/usr/bin/env bash
# HTTP client helpers and session management.

# ── HTTP wrappers ─────────────────────────────────────────────────────────────

# api <METHOD> <path> [curl args...]
# Sends request with cookie jar + JSON header + X-Tenant-ID if set.
api() {
  local method="$1" path="$2"; shift 2
  local headers=(-H "Content-Type: application/json")
  [[ -n "${TENANT_ID:-}" ]] && headers+=(-H "X-Tenant-ID: $TENANT_ID")
  curl -s -X "$method" "${BASE_URL}${path}" \
    -c "$COOKIE_JAR" -b "$COOKIE_JAR" \
    "${headers[@]}" "$@"
}

# apiv <METHOD> <path> [curl args...]
# Same as api but appends "\n<http_status_code>" to the output.
apiv() {
  local method="$1" path="$2"; shift 2
  local headers=(-H "Content-Type: application/json")
  [[ -n "${TENANT_ID:-}" ]] && headers+=(-H "X-Tenant-ID: $TENANT_ID")
  curl -s -w "\n%{http_code}" -X "$method" "${BASE_URL}${path}" \
    -c "$COOKIE_JAR" -b "$COOKIE_JAR" \
    "${headers[@]}" "$@"
}

# api_bearer <token> <METHOD> <path>
# Sends a Bearer-token request (no cookie jar).
api_bearer() {
  local token="$1" method="$2" path="$3"
  curl -s -o /dev/null -w "%{http_code}" \
    -H "Authorization: Bearer $token" \
    -X "$method" "${BASE_URL}${path}"
}

# ── Tenant auto-detection ─────────────────────────────────────────────────────

# resolve_tenant_id — populates $TENANT_ID if not already set.
# Priority: env var (already loaded by env.sh) → data/tenant_id cache → API fetch.
resolve_tenant_id() {
  # Already set (env var or cache loaded by env.sh)
  if [[ -n "${TENANT_ID:-}" ]]; then
    info "Tenant: ${TENANT_ID}"
    return 0
  fi

  info "Auto-detecting tenant ID…"

  # Fetch from platform-level endpoint (no X-Tenant-ID header needed)
  local resp body status tid
  resp=$(curl -s -w "\n%{http_code}" -X GET "${BASE_URL}/api/v1/tenants" \
    -H "Content-Type: application/json" \
    -c "$COOKIE_JAR" -b "$COOKIE_JAR")
  body="$(echo "$resp" | head -n -1)"
  status="$(echo "$resp" | tail -n 1)"

  if [[ "$status" == "200" ]]; then
    tid=$(echo "$body" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    data = d.get('data', d) if isinstance(d, dict) else d
    items = data if isinstance(data, list) else []
    # Prefer ACTIVE tenant if any, otherwise first
    active = [t for t in items if t.get('status','').upper() == 'ACTIVE']
    pick = active[0] if active else (items[0] if items else None)
    print(pick['id'] if pick else '')
except Exception:
    print('')
" 2>/dev/null || true)
  fi

  if [[ -n "$tid" && "$tid" != "None" && "$tid" != "null" && "$tid" != "" ]]; then
    TENANT_ID="$tid"
    echo "$TENANT_ID" > "${DATA_DIR}/tenant_id"
    ok "Tenant ID: ${TENANT_ID}"
    return 0
  fi

  # If API returned 401 we may need to login first — handled by caller
  if [[ "$status" == "401" ]]; then
    warn "Tenant list requires auth — will retry after login"
    return 2
  fi

  warn "Could not auto-detect TENANT_ID (HTTP $status)"
  warn "Run 'bash scripts/seed.sh' or set TENANT_ID=<uuid>"
  return 1
}

# resolve_tenant_id_authed — call resolve_tenant_id after login when
# the tenants endpoint requires authentication.
resolve_tenant_id_authed() {
  [[ -n "${TENANT_ID:-}" ]] && return 0
  resolve_tenant_id
  local rc=$?
  if [[ $rc -eq 2 ]]; then
    # Auth required — login without tenant (some servers allow this for platform ops)
    info "Fetching tenant list with credentials…"
    local resp body status tid
    resp=$(curl -s -w "\n%{http_code}" -X GET "${BASE_URL}/api/v1/tenants" \
      -H "Content-Type: application/json" \
      -H "Authorization: Basic $(echo -n "${ADMIN_EMAIL}:${ADMIN_PASSWORD}" | base64)" \
      -c "$COOKIE_JAR" -b "$COOKIE_JAR")
    body="$(echo "$resp" | head -n -1)"
    status="$(echo "$resp" | tail -n 1)"
    tid=$(echo "$body" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    data = d.get('data', d) if isinstance(d, dict) else d
    items = data if isinstance(data, list) else []
    active = [t for t in items if t.get('status','').upper() == 'ACTIVE']
    pick = active[0] if active else (items[0] if items else None)
    print(pick['id'] if pick else '')
except Exception:
    print('')
" 2>/dev/null || true)
    if [[ -n "$tid" && "$tid" != "None" && "$tid" != "null" && "$tid" != "" ]]; then
      TENANT_ID="$tid"
      echo "$TENANT_ID" > "${DATA_DIR}/tenant_id"
      ok "Tenant ID: ${TENANT_ID}"
      return 0
    fi
  fi
  return $rc
}

# ── Server health ─────────────────────────────────────────────────────────────

health_check() {
  local status
  status=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/health/")
  if [[ "$status" == "200" ]]; then
    ok "Server reachable at ${BASE_URL}"
    return 0
  else
    fail "Server unreachable — GET /health/ returned $status"
    return 1
  fi
}

# ── Session management ────────────────────────────────────────────────────────

# login [email] [password]
# Authenticates and sets AUTH_ACTIVE=true on success.
login() {
  local email="${1:-$ADMIN_EMAIL}" pass="${2:-$ADMIN_PASSWORD}"
  local resp body status

  info "Logging in as ${email}…"
  resp=$(apiv POST /api/v1/auth/login -d "{\"email\":\"$email\",\"password\":\"$pass\"}")
  body="$(echo "$resp" | head -n -1)"
  status="$(echo "$resp" | tail -n 1)"

  case "$status" in
    200)
      AUTH_ACTIVE=true
      AUTH_USER_EMAIL="$email"
      ok "Session started (${email})"
      # Auto-populate TENANT_ID post-login if still missing
      if [[ -z "${TENANT_ID:-}" ]]; then
        resolve_tenant_id || true
      fi
      return 0
      ;;
    202)
      local pt; pt="$(jf "$body" "pending_token")"
      warn "MFA required — pending_token: ${pt:0:14}…"
      warn "Complete: POST /api/v1/auth/mfa/complete"
      AUTH_ACTIVE=false
      return 1
      ;;
    401) fail "Login 401 — bad credentials (${email})"; AUTH_ACTIVE=false; return 1 ;;
    400) fail "Login 400 — check TENANT_ID or request format"; AUTH_ACTIVE=false; return 1 ;;
    *)   fail "Login ${status}: ${body}"; AUTH_ACTIVE=false; return 1 ;;
  esac
}

# logout — terminates the server session and clears the cookie jar.
logout() {
  if [[ "${AUTH_ACTIVE:-false}" == "true" ]]; then
    apiv POST /api/v1/auth/logout > /dev/null 2>&1 || true
    >"$COOKIE_JAR"
    AUTH_ACTIVE=false
    AUTH_USER_EMAIL=""
    ok "Logged out"
  fi
}

# require_auth — logs in if not already authenticated.
require_auth() {
  if [[ "${AUTH_ACTIVE:-false}" != "true" ]]; then
    login || return 1
  fi
  return 0
}

# with_anon_session <fn> — runs a function with a blank cookie jar,
# then restores the current session. Useful for testing 401 guards.
with_anon_session() {
  local fn="$1"
  local saved; saved=$(cat "$COOKIE_JAR" 2>/dev/null || true)
  >"$COOKIE_JAR"
  "$fn"
  echo "$saved" >"$COOKIE_JAR"
}
