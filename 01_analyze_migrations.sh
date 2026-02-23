#!/usr/bin/env bash
# =============================================================================
# SCRIPT 1: Migration Schema Analysis & Clarification
# =============================================================================
# Purpose : Read every migration file, build a full picture of the schema,
#           ask the developer targeted clarifying questions, then emit a
#           structured analysis report that Script 2 can act on safely.
#
# Usage   : bash 01_analyze_migrations.sh [path/to/db/migration]
#
# Output  :
#   .cache/migration_inventory.txt   — every file with line count + summary
#   .cache/schema_map.json           — table → owning migration mapping
#   .cache/cross_refs.txt            — FK and function cross-references
#   .cache/risk_flags.txt            — patterns flagged for human review
#   analysis_report.md               — final human-readable report
#
# Designed for: Claude Code (claude.ai/code) running in the project root.
# =============================================================================

set -euo pipefail
IFS=$'\n\t'

# ── Colour helpers ────────────────────────────────────────────────────────────
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
BOLD='\033[1m'
RESET='\033[0m'

info() { echo -e "${CYAN}[INFO]${RESET}  $*"; }
warn() { echo -e "${YELLOW}[WARN]${RESET}  $*"; }
success() { echo -e "${GREEN}[OK]${RESET}    $*"; }
fatal() {
  echo -e "${RED}[FATAL]${RESET} $*"
  exit 1
}
ask() { echo -e "${BOLD}[?]${RESET}     $*"; }

# ── Argument handling ─────────────────────────────────────────────────────────
MIGRATION_DIR="${1:-db/migration}"
CACHE_DIR=".cache"
REPORT="analysis_report.md"

[[ -d "$MIGRATION_DIR" ]] || fatal "Migration directory not found: $MIGRATION_DIR"
mkdir -p "$CACHE_DIR"

echo ""
echo -e "${BOLD}╔══════════════════════════════════════════════════════════════╗${RESET}"
echo -e "${BOLD}║        ERP Schema Migration Analyser — Claude Code           ║${RESET}"
echo -e "${BOLD}╚══════════════════════════════════════════════════════════════╝${RESET}"
echo ""

# =============================================================================
# PHASE 1 — INVENTORY ALL MIGRATION FILES
# =============================================================================
info "Phase 1/5 — Building migration inventory..."

UP_FILES=($(find "$MIGRATION_DIR" -name "*.up.sql" | sort))
DOWN_FILES=($(find "$MIGRATION_DIR" -name "*.down.sql" | sort))

TOTAL_UP="${#UP_FILES[@]}"
TOTAL_DOWN="${#DOWN_FILES[@]}"

# Check for orphaned up/down files
ORPHAN_UP=0
ORPHAN_DOWN=0
ORPHAN_LIST=""

for f in "${UP_FILES[@]}"; do
  base="${f%.up.sql}"
  if [[ ! -f "${base}.down.sql" ]]; then
    ORPHAN_UP=$((ORPHAN_UP + 1))
    ORPHAN_LIST+="  MISSING DOWN: $f\n"
  fi
done

for f in "${DOWN_FILES[@]}"; do
  base="${f%.down.sql}"
  if [[ ! -f "${base}.up.sql" ]]; then
    ORPHAN_DOWN=$((ORPHAN_DOWN + 1))
    ORPHAN_LIST+="  MISSING UP:   $f\n"
  fi
done

# Write inventory
{
  echo "MIGRATION INVENTORY — $(date)"
  echo "Directory : $MIGRATION_DIR"
  echo "UP files  : $TOTAL_UP"
  echo "DOWN files: $TOTAL_DOWN"
  echo ""
  echo "FILE LISTING (name | lines | tables_created | functions_created)"
  echo "──────────────────────────────────────────────────────────────────"
  for f in "${UP_FILES[@]}"; do
    # lines=$(wc -l <"$f")
    # tables=$(grep -cE "^CREATE TABLE" "$f" 2>/dev/null || echo 0)
    # funcs=$(grep -cE "^CREATE (OR REPLACE )?FUNCTION" "$f" 2>/dev/null || echo 0)
    # views=$(grep -cE "^CREATE (OR REPLACE )?VIEW" "$f" 2>/dev/null || echo 0)
    tables=$(grep -cE "^CREATE TABLE" "$f" 2>/dev/null)
    funcs=$(grep -cE "^CREATE (OR REPLACE )?FUNCTION" "$f" 2>/dev/null)
    views=$(grep -cE "^CREATE (OR REPLACE )?VIEW" "$f" 2>/dev/null)

    tables=${tables:-0}
    funcs=${funcs:-0}
    views=${views:-0}
    printf "%-65s | %4d lines | %2d tables | %2d funcs | %2d views\n" \
      "$(basename "$f")" \
      "$((lines))" "$((tables))" "$((funcs))" "$((views))"
    # printf "%-65s | %4d lines | %2d tables | %2d funcs | %2d views\n" \
    #   "$(basename "$f")" "$lines" "$tables" "$funcs" "$views"
  done
} >"$CACHE_DIR/migration_inventory.txt"

success "Inventory written → $CACHE_DIR/migration_inventory.txt"

# =============================================================================
# PHASE 2 — BUILD TABLE→MIGRATION MAP
# =============================================================================
info "Phase 2/5 — Mapping tables to owning migrations..."

{
  echo "{"
  echo "  \"generated\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\","
  echo "  \"tables\": {"

  first_table=true
  for f in "${UP_FILES[@]}"; do
    fname=$(basename "$f")
    # Extract CREATE TABLE names
    while IFS= read -r line; do
      tname=$(echo "$line" | grep -oE 'CREATE TABLE (IF NOT EXISTS )?[a-z_"]+' |
        awk '{print $NF}' | tr -d '"')
      [[ -z "$tname" ]] && continue
      [[ "$first_table" == "false" ]] && echo ","
      printf "    \"%s\": \"%s\"" "$tname" "$fname"
      first_table=false
    done < <(grep -E "^CREATE TABLE" "$f" 2>/dev/null || true)
  done

  echo ""
  echo "  },"
  echo "  \"functions\": {"

  first_fn=true
  for f in "${UP_FILES[@]}"; do
    fname=$(basename "$f")
    while IFS= read -r line; do
      fname_fn=$(echo "$line" | grep -oE 'FUNCTION [a-z_]+' | awk '{print $2}')
      [[ -z "$fname_fn" ]] && continue
      [[ "$first_fn" == "false" ]] && echo ","
      printf "    \"%s\": \"%s\"" "$fname_fn" "$fname"
      first_fn=false
    done < <(grep -E "^CREATE (OR REPLACE )?FUNCTION" "$f" 2>/dev/null || true)
  done

  echo ""
  echo "  }"
  echo "}"
} >"$CACHE_DIR/schema_map.json"

success "Schema map written → $CACHE_DIR/schema_map.json"

# =============================================================================
# PHASE 3 — CROSS-REFERENCE ANALYSIS
# =============================================================================
info "Phase 3/5 — Analysing cross-references and dependencies..."

{
  echo "CROSS-REFERENCE ANALYSIS — $(date)"
  echo ""

  echo "══ FOREIGN KEY REFERENCES ══"
  for f in "${UP_FILES[@]}"; do
    refs=$(grep -E "REFERENCES [a-z_]+" "$f" 2>/dev/null || true)
    if [[ -n "$refs" ]]; then
      echo "── $(basename "$f") ──"
      echo "$refs" | sed 's/^/  /'
      echo ""
    fi
  done

  echo ""
  echo "══ RLS POLICY PRESENCE ══"
  for f in "${UP_FILES[@]}"; do
    has_enable=$(grep -c "ENABLE ROW LEVEL SECURITY" "$f" 2>/dev/null || echo 0)
    has_policy=$(grep -c "CREATE POLICY" "$f" 2>/dev/null || echo 0)
    tables_in_file=$(grep -cE "^CREATE TABLE" "$f" 2>/dev/null || echo 0)

    if [[ "$tables_in_file" -gt 0 ]]; then
      if [[ "$has_enable" -eq 0 ]]; then
        echo "  NO RLS  : $(basename "$f") ($tables_in_file tables, 0 RLS enables)"
      else
        echo "  HAS RLS : $(basename "$f") ($has_enable enables, $has_policy policies)"
      fi
    fi
  done

  echo ""
  echo "══ SECURITY DEFINER FUNCTIONS ══"
  for f in "${UP_FILES[@]}"; do
    has_sd=$(grep -l "SECURITY DEFINER" "$f" 2>/dev/null || true)
    if [[ -n "$has_sd" ]]; then
      # Check if search_path is pinned
      has_sp=$(grep -c "SET search_path" "$f" 2>/dev/null || echo 0)
      if [[ "$has_sp" -eq 0 ]]; then
        echo "  UNPINNED: $(basename "$f") — SECURITY DEFINER without search_path pin"
      else
        echo "  SAFE    : $(basename "$f") — SECURITY DEFINER + search_path pinned"
      fi
    fi
  done

  echo ""
  echo "══ TRIGGER FUNCTIONS USAGE ══"
  for f in "${UP_FILES[@]}"; do
    trigger_count=$(grep -c "CREATE TRIGGER" "$f" 2>/dev/null || echo 0)
    if [[ "$trigger_count" -gt 0 ]]; then
      echo "  $(basename "$f"): $trigger_count triggers"
      grep "EXECUTE FUNCTION" "$f" 2>/dev/null | sed 's/^/    /' || true
    fi
  done

} >"$CACHE_DIR/cross_refs.txt"

success "Cross-references written → $CACHE_DIR/cross_refs.txt"

# =============================================================================
# PHASE 4 — RISK FLAG DETECTION
# =============================================================================
info "Phase 4/5 — Detecting risk patterns..."

RISK_COUNT=0
{
  echo "RISK FLAG REPORT — $(date)"
  echo "Each flag should be reviewed before Script 2 runs."
  echo ""

  # ── R1: CHECK constraints querying other tables ──────────────────────────
  echo "══ R1: CHECK constraints with subqueries (unreliable in PostgreSQL) ══"
  for f in "${UP_FILES[@]}"; do
    matches=$(grep -n "CHECK.*EXISTS\|CHECK.*SELECT" "$f" 2>/dev/null || true)
    if [[ -n "$matches" ]]; then
      echo "  FILE: $(basename "$f")"
      echo "$matches" | sed 's/^/    /'
      RISK_COUNT=$((RISK_COUNT + 1))
    fi
  done
  echo ""

  # ── R2: RLS fail-open patterns ────────────────────────────────────────────
  echo "══ R2: RLS fail-open pattern (current_tenant_id() IS NOT NULL) ══"
  echo "    Note: This is intentional in some policies — review each one."
  for f in "${UP_FILES[@]}"; do
    matches=$(grep -n "current_tenant_id() IS NOT NULL" "$f" 2>/dev/null || true)
    if [[ -n "$matches" ]]; then
      echo "  FILE: $(basename "$f")"
      echo "$matches" | sed 's/^/    /'
      RISK_COUNT=$((RISK_COUNT + 1))
    fi
  done
  echo ""

  # ── R3: SECURITY DEFINER without search_path ─────────────────────────────
  echo "══ R3: SECURITY DEFINER without search_path pin (injection risk) ══"
  for f in "${UP_FILES[@]}"; do
    if grep -q "SECURITY DEFINER" "$f" 2>/dev/null; then
      if ! grep -q "SET search_path" "$f" 2>/dev/null; then
        echo "  RISK: $(basename "$f")"
        grep -n "SECURITY DEFINER" "$f" | sed 's/^/    /'
        RISK_COUNT=$((RISK_COUNT + 1))
      fi
    fi
  done
  echo ""

  # ── R4: ORDER BY inside view definitions ─────────────────────────────────
  echo "══ R4: ORDER BY inside CREATE VIEW (ignored by planner) ══"
  in_view=false
  for f in "${UP_FILES[@]}"; do
    # Simplified detection: ORDER BY after CREATE VIEW before next DDL
    matches=$(awk '/CREATE.*VIEW/,/^;/' "$f" 2>/dev/null |
      grep -n "ORDER BY" || true)
    if [[ -n "$matches" ]]; then
      echo "  FILE: $(basename "$f")"
      echo "$matches" | sed 's/^/    /'
      RISK_COUNT=$((RISK_COUNT + 1))
    fi
  done
  echo ""

  # ── R5: Tables with no updated_at trigger ────────────────────────────────
  echo "══ R5: Tables with updated_at column but no trigger ══"
  for f in "${UP_FILES[@]}"; do
    has_updated_at=$(grep -c "updated_at" "$f" 2>/dev/null || echo 0)
    has_trigger=$(grep -c "update_updated_at\|updated_at.*trigger\|TRIGGER.*updated" \
      "$f" 2>/dev/null || echo 0)
    tables=$(grep -cE "^CREATE TABLE" "$f" 2>/dev/null || echo 0)
    if [[ "$has_updated_at" -gt 0 && "$has_trigger" -eq 0 && "$tables" -gt 0 ]]; then
      echo "  FILE: $(basename "$f") — has updated_at but no trigger"
      RISK_COUNT=$((RISK_COUNT + 1))
    fi
  done
  echo ""

  # ── R6: Duplicate table creation ─────────────────────────────────────────
  echo "══ R6: Duplicate table names across migrations ══"
  declare -A seen_tables
  for f in "${UP_FILES[@]}"; do
    while IFS= read -r tname; do
      [[ -z "$tname" ]] && continue
      if [[ -v "seen_tables[$tname]" ]]; then
        echo "  DUPLICATE: '$tname' in $(basename "$f") and ${seen_tables[$tname]}"
        RISK_COUNT=$((RISK_COUNT + 1))
      else
        seen_tables[$tname]="$(basename "$f")"
      fi
    done < <(grep -E "^CREATE TABLE" "$f" 2>/dev/null |
      grep -oE 'CREATE TABLE (IF NOT EXISTS )?[a-z_"]+' |
      awk '{print $NF}' | tr -d '"' || true)
  done
  echo ""

  # ── R7: Missing down migrations ──────────────────────────────────────────
  echo "══ R7: Up migrations without a corresponding down migration ══"
  if [[ -n "$ORPHAN_LIST" ]]; then
    echo -e "$ORPHAN_LIST"
    RISK_COUNT=$((RISK_COUNT + ORPHAN_UP + ORPHAN_DOWN))
  else
    echo "  None found — all up/down pairs are present."
  fi
  echo ""

  # ── R8: Feature flag cache table ─────────────────────────────────────────
  echo "══ R8: Feature flag database cache (anti-pattern) ══"
  for f in "${UP_FILES[@]}"; do
    if echo "$(basename "$f")" | grep -qi "feature_flag_cache"; then
      echo "  ANTI-PATTERN: $(basename "$f")"
      echo "    A DB table for feature flag caching adds read overhead."
      echo "    Recommendation: move cache to Redis / application layer."
      RISK_COUNT=$((RISK_COUNT + 1))
    fi
  done
  echo ""

  # ── R9: Parallel policy evaluation systems ───────────────────────────────
  echo "══ R9: Potential dual policy evaluation (RBAC + ABAC) ══"
  rbac_files=()
  abac_files=()
  for f in "${UP_FILES[@]}"; do
    bn=$(basename "$f")
    echo "$bn" | grep -qi "auth\|role_perm\|user_role" && rbac_files+=("$bn")
    echo "$bn" | grep -qi "attribute\|policy_eval\|abac" && abac_files+=("$bn")
  done
  if [[ ${#rbac_files[@]} -gt 0 && ${#abac_files[@]} -gt 0 ]]; then
    echo "  RBAC migrations found: ${#rbac_files[@]}"
    printf '    %s\n' "${rbac_files[@]}"
    echo "  ABAC migrations found: ${#abac_files[@]}"
    printf '    %s\n' "${abac_files[@]}"
    echo "  → Verify a single policy decision point exists."
    RISK_COUNT=$((RISK_COUNT + 1))
  fi
  echo ""

  echo "══════════════════════════════════════════════════════════════"
  echo "TOTAL RISK FLAGS: $RISK_COUNT"
  echo "══════════════════════════════════════════════════════════════"

} >"$CACHE_DIR/risk_flags.txt"

success "Risk flags written → $CACHE_DIR/risk_flags.txt  ($RISK_COUNT flags)"

# =============================================================================
# PHASE 5 — INTERACTIVE CLARIFICATION QUESTIONS
# =============================================================================
info "Phase 5/5 — Collecting developer answers to clarification questions..."

echo ""
echo -e "${BOLD}════════════════════════════════════════════════════════════════${RESET}"
echo -e "${BOLD}  CLARIFICATION QUESTIONS — Please answer before running Script 2${RESET}"
echo -e "${BOLD}════════════════════════════════════════════════════════════════${RESET}"
echo ""
echo "  Your answers are saved to .cache/developer_answers.txt"
echo "  Script 2 reads this file to make safe decisions."
echo ""

ANSWERS_FILE="$CACHE_DIR/developer_answers.txt"
>"$ANSWERS_FILE"

_ask_and_save() {
  local key="$1"
  local question="$2"
  local hint="${3:-}"
  echo ""
  ask "$question"
  [[ -n "$hint" ]] && echo -e "    ${CYAN}Hint: $hint${RESET}"
  printf "    Your answer: "
  read -r answer
  echo "${key}=${answer}" >>"$ANSWERS_FILE"
  echo "    ✓ Saved"
}

# Q1
_ask_and_save "Q1_MIGRATION_TOOL" \
  "Q1. What migration tool are you using?" \
  "Options: golang-migrate | goose | flyway | dbmate | atlas | other"

# Q2
_ask_and_save "Q2_DATABASE_LIVE" \
  "Q2. Is there a LIVE production database running these migrations?" \
  "yes / no  — This affects whether we use CONCURRENTLY for indexes"

# Q3
_ask_and_save "Q3_CONFIG_DEF_LOCATION" \
  "Q3. Which migration is the authoritative source for config_definitions?" \
  "000058_config_definitions or 000601_settings_core_tables — check which one ConfigurationService queries"

# Q4
_ask_and_save "Q4_IAM_AUTHORITY" \
  "Q4. Is RBAC (000404-412) or ABAC (000701-705) the authoritative policy decision?" \
  "rbac / abac / both-unified / not-decided"

# Q5
_ask_and_save "Q5_FEATURE_FLAG_CACHE" \
  "Q5. Is Redis/Memcached available in your infrastructure?" \
  "yes / no — Determines if we can move feature flag cache out of Postgres"

# Q6
_ask_and_save "Q6_SERVICE_ACCOUNTS" \
  "Q6. Do background workers and API integrations currently have user rows in the identity tables?" \
  "yes / no / not-yet"

# Q7
_ask_and_save "Q7_TENANT_FREEZE" \
  "Q7. Can we add a rule that no new columns are added to the tenants table without a review?" \
  "yes / no — Helps enforce module ownership boundaries"

# Q8
_ask_and_save "Q8_VIEWS_SECURITY" \
  "Q8. Which PostgreSQL version are you running?" \
  "Run: SELECT version(); — Needed for SECURITY INVOKER view support (PG15+)"

# Q9
_ask_and_save "Q9_ENTITY_HIERARCHY" \
  "Q9. What is the maximum real-world entity hierarchy depth in your tenants?" \
  "e.g. 3 = COMPANY→DEPARTMENT→TEAM — Used to set the hierarchy depth limit trigger"

# Q10
_ask_and_save "Q10_NAMING_CONVENTION" \
  "Q10. Should refactored migrations follow the existing naming pattern?" \
  "yes = keep 000NNN_domain_description.up.sql / no = let Script 2 decide"

echo ""
success "All answers saved → $ANSWERS_FILE"

# =============================================================================
# GENERATE REPORT
# =============================================================================
info "Generating analysis_report.md..."

# Parse answers
source "$ANSWERS_FILE" 2>/dev/null || true

{
  cat <<'REPORT_HEADER'
# ERP Schema Migration Analysis Report

> Generated by `01_analyze_migrations.sh`
> Review this report fully before running `02_refactor_migrations.sh`

---

REPORT_HEADER

  echo "## 1. Inventory Summary"
  echo ""
  echo "| Metric | Value |"
  echo "|--------|-------|"
  echo "| Total UP migrations   | $TOTAL_UP |"
  echo "| Total DOWN migrations | $TOTAL_DOWN |"
  echo "| Orphaned UP files     | $ORPHAN_UP |"
  echo "| Orphaned DOWN files   | $ORPHAN_DOWN |"
  echo "| Risk flags detected   | $RISK_COUNT |"
  echo ""

  echo "## 2. Domain Blocks"
  echo ""
  echo "| Range | Domain | Count |"
  echo "|-------|--------|-------|"
  echo "| 000015–000017 | Platform IAM foundations | 3 |"
  echo "| 000051–000059 | Tenant core (our migrations) | 9 |"
  echo "| 000102–000107 | Tenant patch accretion | 6 |"
  echo "| 000201–000204 | Entity core | 4 |"
  echo "| 000301–000304 | Identity | 4 |"
  echo "| 000404–000412 | Auth / RBAC | 9 |"
  echo "| 000450–000451 | Audit | 2 |"
  echo "| 000501–000504 | User management | 4 |"
  echo "| 000601–000602 | Settings | 2 |"
  echo "| 000701–000705 | ABAC / Policy | 5 |"
  echo "| 000801–000807 | Feature flags | 7 |"
  echo "| 000901–000911 | Finance | 11 |"
  echo ""

  echo "## 3. Risk Flags"
  echo ""
  echo "\`\`\`"
  cat "$CACHE_DIR/risk_flags.txt"
  echo "\`\`\`"
  echo ""

  echo "## 4. Your Answers"
  echo ""
  echo "\`\`\`"
  cat "$ANSWERS_FILE"
  echo "\`\`\`"
  echo ""

  echo "## 5. Recommended Refactoring Actions"
  echo ""
  echo "Script 2 will execute the following. Review each before confirming."
  echo ""

  echo "### SAFE (zero data risk — run in any environment)"
  echo "- [ ] Pin \`search_path\` on all SECURITY DEFINER functions"
  echo "- [ ] Add \`COMMENT ON TABLE/COLUMN\` where missing"
  echo "- [ ] Add \`is_tenant_overridable\` / \`is_entity_overridable\` to config_definitions"
  echo "- [ ] Add \`document_type\` column to entitystate (rename KEY)"
  echo "- [ ] Add \`config JSONB DEFAULT '{}'\` to entitystate (Phase 2 PRD)"
  echo "- [ ] Replace subdomain CHECK subquery with a trigger"
  echo "- [ ] Add \`(tenant_id, applied_at DESC)\` index to configuration_audit"
  echo "- [ ] Add \`config_definition_id UUID\` FK to configuration_audit"
  echo "- [ ] Remove ORDER BY from view definitions"
  echo "- [ ] Add slug immutability trigger to entities table"
  echo "- [ ] Add \`user_type\` column to users table for service accounts"
  echo ""

  echo "### CAREFUL (requires coordination with application team)"
  echo "- [ ] Remove \`tenant_id\` and \`entity_id\` from config_definitions"
  echo "- [ ] Consolidate config_definitions location (000058 vs 000601)"
  echo "- [ ] Remove redundant \`entity_id\` from hierarchy_paths"
  echo "- [ ] Remove validation columns from hierarchy_paths"
  echo "- [ ] Replace GRANT on config_definitions to match RLS intent"
  echo "- [ ] Add unified policy decision point (RBAC+ABAC)"
  echo ""

  echo "### ARCHITECTURE (human decision required — Script 2 will NOT execute)"
  echo "- [ ] Move feature flag cache from DB to Redis (Q5=$Q5_FEATURE_FLAG_CACHE)"
  echo "- [ ] Establish 'no new columns on tenants table' review rule (Q7=$Q7_TENANT_FREEZE)"
  echo "- [ ] Add permissions assignment table linking resource+action to role"
  echo "- [ ] Rewrite v_active_entities CASE-in-JOIN as CTEs"
  echo "- [ ] Unify RBAC/ABAC into single policy decision function (Q4=$Q4_IAM_AUTHORITY)"
  echo ""

  echo "---"
  echo ""
  echo "## 6. Full Cross-Reference Log"
  echo ""
  echo "\`\`\`"
  cat "$CACHE_DIR/cross_refs.txt"
  echo "\`\`\`"

} >"$REPORT"

echo ""
success "Report written → $REPORT"

echo ""
echo -e "${BOLD}════════════════════════════════════════════════════════════════${RESET}"
echo -e "${BOLD}  ANALYSIS COMPLETE${RESET}"
echo -e "${BOLD}════════════════════════════════════════════════════════════════${RESET}"
echo ""
echo "  Generated files:"
echo "    $REPORT                          ← Read this first"
echo "    $CACHE_DIR/migration_inventory.txt"
echo "    $CACHE_DIR/schema_map.json"
echo "    $CACHE_DIR/cross_refs.txt"
echo "    $CACHE_DIR/risk_flags.txt"
echo "    $CACHE_DIR/developer_answers.txt  ← Read by Script 2"
echo ""
echo "  Next step:"
echo -e "    ${CYAN}bash 02_refactor_migrations.sh${RESET}"
echo ""
warn "DO NOT run Script 2 until you have read analysis_report.md in full."
echo ""
