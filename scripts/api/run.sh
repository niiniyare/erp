#!/usr/bin/env bash
# AWO API Test Suite — interactive and automated runner.
#
# Usage:
#   bash scripts/api/run.sh              # interactive menu
#   bash scripts/api/run.sh --all        # run all services sequentially
#   bash scripts/api/run.sh --service iam  # run one service by name
#   bash scripts/api/run.sh --list       # list registered services
#   bash scripts/api/run.sh --help
#
# Environment:
#   BASE_URL        default: http://localhost:8080
#   ADMIN_EMAIL     default: admin@platform.local
#   ADMIN_PASSWORD  default: Admin1234!
#   TENANT_ID       required for tenant-scoped requests

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

source "$SCRIPT_DIR/core/env.sh"
source "$SCRIPT_DIR/core/utils.sh"
source "$SCRIPT_DIR/core/auth.sh"

trap 'logout 2>/dev/null || true; cleanup_session' EXIT

# ── Service registry ──────────────────────────────────────────────────────────
# Each services/*.sh appends to MENU: "name:description:runner_fn"
declare -a MENU=()

for _svc_file in "$SCRIPT_DIR/services/"*.sh; do
  [[ -f "$_svc_file" ]] && source "$_svc_file"
done

# ── Helpers ───────────────────────────────────────────────────────────────────

_banner() {
  echo -e "${BOLD}${CYAN}"
  echo   "  ╔══════════════════════════════════════════╗"
  echo   "  ║   AWO  API  Test  Suite                  ║"
  printf "  ║   session  %-30s║\n" "$SESSION_ID"
  printf "  ║   host     %-30s║\n" "$BASE_URL"
  printf "  ║   tenant   %-30s║\n" "${TENANT_ID:-<not set>}"
  echo   "  ╚══════════════════════════════════════════╝"
  echo -e "${NC}"
}

_list_services() {
  echo -e "${BOLD}Registered services:${NC}"
  local i=1
  for entry in "${MENU[@]}"; do
    local name desc
    name="${entry%%:*}"
    desc="${entry#*:}"; desc="${desc%:*}"
    printf "  %2d)  %-12s  %s\n" "$i" "$name" "$desc"
    i=$((i + 1))
  done
}

_run_by_name() {
  local target="$1"
  for entry in "${MENU[@]}"; do
    local name runner
    name="${entry%%:*}"
    runner="${entry##*:}"
    if [[ "$name" == "$target" ]]; then
      "$runner"
      return $?
    fi
  done
  die "Unknown service: '$target'  (use --list to see available services)"
}

_run_all() {
  _banner
  health_check      || die "Server not reachable. Start it first."
  resolve_tenant_id || true   # best-effort before login
  login             || die "Login failed. Run 'bash scripts/seed.sh' first."

  local total=0 passed=0
  for entry in "${MENU[@]}"; do
    local runner="${entry##*:}"
    SUITE_FAILURES=0
    "$runner" || true
    total=$((total + 1))
    [[ $SUITE_FAILURES -eq 0 ]] && passed=$((passed + 1))
  done

  logout
  sep
  echo -e "${BOLD}Results: $passed/$total suite(s) passed${NC}"
}

# ── Interactive menu ──────────────────────────────────────────────────────────

_interactive() {
  clear
  health_check || {
    warn "Server not reachable at $BASE_URL"
    warn "Start the server first, then re-run."
    exit 1
  }
  resolve_tenant_id || true   # best-effort; login() retries post-auth

  while true; do
    _banner

    echo -e "${BOLD}Select a suite:${NC}"
    echo "    0)  Run all"

    local i=1
    declare -a _names=()
    for entry in "${MENU[@]}"; do
      local name desc
      name="${entry%%:*}"; _names+=("$name")
      desc="${entry#*:}"; desc="${desc%:*}"
      printf "  %3d)  %s\n" "$i" "$desc"
      i=$((i + 1))
    done
    echo "    q)  Quit"
    echo ""

    local choice
    read -r -p "  › " choice

    case "$choice" in
      0)
        login || { warn "Login failed"; _pause; continue; }
        for entry in "${MENU[@]}"; do
          local runner="${entry##*:}"
          SUITE_FAILURES=0
          "$runner" || true
        done
        logout
        ;;
      q|Q|quit|exit)
        echo -e "\n${DIM}Bye.${NC}"
        break
        ;;
      ''|*[!0-9]*)
        warn "Invalid choice — enter a number or 'q'"
        ;;
      *)
        local idx=$((choice - 1))
        if [[ $idx -ge 0 && $idx -lt ${#_names[@]} ]]; then
          require_auth || { _pause; continue; }
          SUITE_FAILURES=0
          _run_by_name "${_names[$idx]}" || true
        else
          warn "Invalid choice"
        fi
        ;;
    esac

    _pause
    clear
  done
}

_pause() {
  echo ""
  read -r -p "  Press Enter to continue…" _dummy || true
}

# ── Entry point ───────────────────────────────────────────────────────────────

MODE="interactive"
TARGET=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --all)       MODE="all" ;;
    --service)   MODE="service"; TARGET="${2:-}"; shift ;;
    --list)      MODE="list" ;;
    --help|-h)
      echo "Usage: bash scripts/api/run.sh [--all | --service <name> | --list]"
      exit 0
      ;;
    *) die "Unknown argument: $1" ;;
  esac
  shift
done

case "$MODE" in
  all)
    _run_all
    ;;
  service)
    [[ -n "$TARGET" ]] || die "--service requires a name (use --list to see options)"
    _banner
    health_check      || die "Server not reachable"
    resolve_tenant_id || true
    require_auth      || die "Login failed"
    SUITE_FAILURES=0
    _run_by_name "$TARGET"
    logout
    ;;
  list)
    _list_services
    ;;
  interactive)
    _interactive
    ;;
esac
