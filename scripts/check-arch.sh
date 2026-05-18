#!/usr/bin/env bash
# check-arch.sh — UI pipeline architecture guards
#
# Run from the repository root. Exits 1 if any guard fails.
# Add to CI as a required check before build and test.
#
# Guards:
#   1  No map[string]any in internal/web/ (excluding ast/ and ui/)
#   2  No direct IAM imports in internal/web/dsl/
#   3  No permission checks inside visibleOn expressions
#   4  No schema generated outside the pipeline (no PageFn calls outside stages/)
#   5  Every non-root UI stage declares StageDependsOn
#   6  No sync.Map or third-party in-process caches in internal/web/
#   7  No raw ast node construction in screens/ (block-level concerns must live in blocks/)
#   8  Every screens/ file is under 60 lines

set -euo pipefail

PASS=0
FAIL=1
failures=0

red()   { printf '\033[0;31m%s\033[0m\n' "$*"; }
green() { printf '\033[0;32m%s\033[0m\n' "$*"; }
warn()  { printf '\033[0;33m%s\033[0m\n' "$*"; }

fail() {
  red "FAIL [$1]: $2"
  failures=$((failures + 1))
}

pass() {
  green "PASS [$1]: $2"
}

# ── Guard 1 ──────────────────────────────────────────────────────────────────
# No map[string]any literals in internal/web/ outside of ast/ and ui/.
# ast/ is allowed because Compile() must return map[string]any.
# ui/ is allowed because it defines the M = map[string]any alias.
# ChartNode.Config assignments in dsl/ are the sole intentional escape hatch —
# they are excluded via the `Config:` line filter.
#
guard=1
desc="no map[string]any outside ast/ and ui/"

hits=$(grep -rn --include="*.go" "map\[string\]any" internal/web/ \
  | grep -v "internal/web/ast/" \
  | grep -v "internal/web/ui/" \
  | grep -v "Config:" \
  | grep -v "// arch:allow-map" \
  || true)

if [ -n "$hits" ]; then
  fail $guard "$desc"
  echo "$hits"
else
  pass $guard "$desc"
fi

# ── Guard 2 ──────────────────────────────────────────────────────────────────
# No direct IAM package imports from internal/web/dsl/.
# DSL code interacts with IAM exclusively through ui.UISessionContext.
#
guard=2
desc="no direct IAM imports in dsl/"

hits=$(grep -rn --include="*.go" '"awo.so/internal/core/iam' internal/web/dsl/ \
  || true)

if [ -n "$hits" ]; then
  fail $guard "$desc"
  echo "$hits"
else
  pass $guard "$desc"
fi

# ── Guard 3 ──────────────────────────────────────────────────────────────────
# No permission/role/IAM string checks embedded in visibleOn expressions.
# Permission logic belongs in block Go code via sess.Can(), not in AMIS expressions.
#
guard=3
desc="no permission checks in visibleOn expressions"

hits=$(grep -rn --include="*.go" 'VisibleOn:.*\(perm\|role\|iam\|Can\|permission\)' \
  internal/web/ \
  || true)

if [ -n "$hits" ]; then
  fail $guard "$desc"
  echo "$hits"
else
  pass $guard "$desc"
fi

# ── Guard 4 ──────────────────────────────────────────────────────────────────
# No PageFn or ASTPageFn called directly outside pipeline stages.
# Schemas must be generated through the pipeline, never inline in handlers.
#
guard=4
desc="no PageFn invocations outside stages/"

# Look for sites that call a variable of type PageFn directly (fn(sess)) outside stages/
hits=$(grep -rn --include="*.go" 'PageFn\|ASTPageFn' internal/web/ \
  | grep -v "internal/web/stages/" \
  | grep -v "internal/web/ui/types.go" \
  | grep -v "internal/web/registry/" \
  | grep -v "// " \
  | grep -v "type " \
  | grep -v "func " \
  | grep -v "DataKey" \
  || true)

if [ -n "$hits" ]; then
  fail $guard "$desc"
  echo "$hits"
else
  pass $guard "$desc"
fi

# ── Guard 5 ──────────────────────────────────────────────────────────────────
# Every non-root UI stage must declare StageDependsOn.
# Root = ui.session (Priority 10). All others must declare deps.
#
guard=5
desc="every non-root UI stage declares StageDependsOn"

# Find BaseStage{} constructions in web/stages/ that lack StageDependsOn.
# Strategy: find NewXxxStage functions that embed BaseStage without StageDependsOn.
hits=$(grep -rn --include="*.go" -A 10 "BaseStage{" internal/web/stages/ \
  | awk '
    /BaseStage\{/ { block=""; found_dep=0; file_line=$0; next }
    found_dep==0 && /StageDependsOn/ { found_dep=1 }
    found_dep==0 && /StageName:.*ui\.session/ { found_dep=1 }  # root stage exempt
    found_dep==0 && /\},$/ {
      if (found_dep==0 && block != "") print file_line "\n" block
    }
    { block=block $0 "\n" }
  ' \
  || true)

# Simpler approach: check that every stage file (except session) has StageDependsOn
missing=""
for f in internal/web/stages/*.go; do
  # Skip non-stage files
  case "$f" in
    *instrument* | *normalize* | *validate* ) ;;
    * )
      if grep -q "BaseStage{" "$f" 2>/dev/null; then
        if ! grep -q "ui\.session\|StageDependsOn" "$f" 2>/dev/null; then
          missing="$missing\n  $f"
        fi
      fi
      ;;
  esac
done

if [ -n "$missing" ]; then
  fail $guard "$desc"
  printf "$missing\n"
else
  pass $guard "$desc"
fi

# ── Guard 6 ──────────────────────────────────────────────────────────────────
# No sync.Map or third-party in-process caches in internal/web/.
# All caching must go through cache.Service (Redis + memory via platform layer).
#
guard=6
desc="no sync.Map or third-party caches in internal/web/"

hits=$(grep -rn --include="*.go" \
  -e "sync\.Map" \
  -e '"github.com/dgraph-io/ristretto"' \
  -e '"github.com/allegro/bigcache"' \
  -e '"github.com/coocood/freecache"' \
  -e '"github.com/patrickmn/go-cache"' \
  internal/web/ \
  || true)

if [ -n "$hits" ]; then
  fail $guard "$desc"
  echo "$hits"
else
  pass $guard "$desc"
fi

# ── Guard 7 ──────────────────────────────────────────────────────────────────
# screens/ files must not construct raw AST nodes that belong in blocks/.
# Allowed: PageNode (the screen root), GridNode (layout only), FlexNode (layout only).
# Not allowed: ComboNode, CRUDNode, FilterBarNode, TableNode, StatNode, TimelineNode —
# these are block-level concerns; their construction must live in blocks/.
#
guard=7
desc="no raw block-level AST nodes constructed in screens/"

hits=$(grep -rn --include="*.go" \
  -e "ast\.ComboNode{" \
  -e "ast\.CRUDNode{" \
  -e "ast\.FilterBarNode{" \
  -e "ast\.TableNode{" \
  -e "ast\.StatNode{" \
  -e "ast\.TimelineNode{" \
  -e "ast\.ChartNode{" \
  -e "ast\.CardNode{" \
  -e "ast\.SectionNode{" \
  internal/web/dsl/screens/ \
  || true)

if [ -n "$hits" ]; then
  fail $guard "$desc"
  echo "$hits"
else
  pass $guard "$desc"
fi

# ── Guard 8 ──────────────────────────────────────────────────────────────────
# Every screens/ file must be under 60 lines (blank lines and comments included).
# A file over 60 lines means logic escaped the block layer.
#
guard=8
desc="every screens/ file is under 60 lines"

over60=""
for f in internal/web/dsl/screens/*.go; do
  lines=$(wc -l < "$f")
  if [ "$lines" -gt 60 ]; then
    over60="$over60\n  $f ($lines lines)"
  fi
done

if [ -n "$over60" ]; then
  fail $guard "$desc"
  printf "$over60\n"
else
  pass $guard "$desc"
fi

# ── Result ────────────────────────────────────────────────────────────────────

echo ""
if [ "$failures" -gt 0 ]; then
  red "Architecture check FAILED: $failures guard(s) violated."
  exit 1
else
  green "Architecture check PASSED: all guards clean."
  exit 0
fi
