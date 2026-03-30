#!/usr/bin/env bash
# Output helpers, assertions, and JSON utilities.

# ── Colors (disabled when not a tty) ─────────────────────────────────────────
if [[ -t 1 ]]; then
  GREEN='\033[0;32m'; RED='\033[0;31m'; YELLOW='\033[1;33m'
  CYAN='\033[0;36m'; BOLD='\033[1m'; DIM='\033[2m'; NC='\033[0m'
else
  GREEN=''; RED=''; YELLOW=''; CYAN=''; BOLD=''; DIM=''; NC=''
fi

# ── Output ────────────────────────────────────────────────────────────────────
ok()     { echo -e "${GREEN}  ✓ $*${NC}"; }
fail()   { echo -e "${RED}  ✗ $*${NC}"; SUITE_FAILURES=$((SUITE_FAILURES + 1)); }
info()   { echo -e "${CYAN}  ▶ $*${NC}"; }
warn()   { echo -e "${YELLOW}  ! $*${NC}"; }
sep()    { echo -e "${DIM}────────────────────────────────────────${NC}"; }
header() { echo -e "\n${BOLD}${CYAN}━━  $*${NC}"; }
die()    { echo -e "${RED}FATAL: $*${NC}" >&2; exit 1; }

# Global failure counter — reset at the start of each service run
SUITE_FAILURES=0

reset_counters() { SUITE_FAILURES=0; }

# ── HTTP assertion ────────────────────────────────────────────────────────────
# assert_status <label> <expected_code> <actual_code>
assert_status() {
  local label="$1" expected="$2" actual="$3"
  if [[ "$actual" == "$expected" ]]; then
    ok "$label → HTTP $actual"
    return 0
  else
    fail "$label → expected HTTP $expected, got $actual"
    return 1
  fi
}

# assert_any <label> <actual_code> <code1> [code2 ...]
assert_any() {
  local label="$1" actual="$2"; shift 2
  for s in "$@"; do
    if [[ "$actual" == "$s" ]]; then
      ok "$label → HTTP $actual"
      return 0
    fi
  done
  fail "$label → expected one of ($*), got $actual"
  return 1
}

# ── Response helpers ──────────────────────────────────────────────────────────
# split_resp <apiv_output> → sets RESP_BODY and RESP_STATUS
split_resp() {
  RESP_BODY="$(echo "$1" | head -n -1)"
  RESP_STATUS="$(echo "$1" | tail -n 1)"
}

# ── JSON field extractor ──────────────────────────────────────────────────────
# jf <json_string> <dot.path>   e.g.  jf "$body" "data.id"
jf() {
  echo "$1" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    for k in '$2'.split('.'):
        d = d[k] if isinstance(d, dict) else d[int(k)]
    print('' if d is None else d)
except Exception:
    print('')
" 2>/dev/null || true
}

# json_len <json_string> <dot.path_to_array>
json_len() {
  echo "$1" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    for k in '$2'.split('.'):
        d = d[k]
    print(len(d))
except Exception:
    print(0)
" 2>/dev/null || echo 0
}

# ── Session cleanup ───────────────────────────────────────────────────────────
cleanup_session() {
  [[ -f "${COOKIE_JAR:-}" ]] && rm -f "$COOKIE_JAR"
}
