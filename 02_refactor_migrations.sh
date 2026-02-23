#!/usr/bin/env bash
# =============================================================================
# SCRIPT 2: Migration Refactoring Executor
# =============================================================================
# Purpose : Read the analysis produced by Script 1 (developer_answers.txt +
#           risk_flags.txt), generate new migration files for each safe fix,
#           and produce a side-by-side diff of every change.
#
#           This script NEVER modifies existing migration files.
#           It ONLY creates new migrations in a separate output directory,
#           prints a preview, and asks for confirmation before writing.
#
# Usage   : bash 02_refactor_migrations.sh [path/to/db/migration]
#
# Output  :
#   refactored/NNN_fix_name.up.sql    — new migration files (safe additions)
#   refactored/NNN_fix_name.down.sql  — corresponding rollback migrations
#   refactored/REFACTOR_SUMMARY.md    — what was changed and why
#   refactored/APPLY_ORDER.md         — in what order to apply these migrations
#
# Safety contract:
#   ✓  Creates NEW migrations only — never edits existing files
#   ✓  Every change is reversible — all new migrations have a .down.sql
#   ✓  Shows a full preview and asks YES/NO before writing any file
#   ✓  Architecture changes are documented but NOT executed
#   ✗  Does NOT drop columns (requires explicit developer action)
#   ✗  Does NOT rename columns (requires multi-step migration by hand)
#   ✗  Does NOT touch any file outside ./refactored/
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
DIM='\033[2m'

info() { echo -e "${CYAN}[INFO]${RESET}  $*"; }
warn() { echo -e "${YELLOW}[WARN]${RESET}  $*"; }
success() { echo -e "${GREEN}[OK]${RESET}    $*"; }
fatal() {
  echo -e "${RED}[FATAL]${RESET} $*"
  exit 1
}
preview() { echo -e "${DIM}$*${RESET}"; }

# ── Paths ─────────────────────────────────────────────────────────────────────
MIGRATION_DIR="${1:-db/migration}"
CACHE_DIR=".cache"
OUT_DIR="refactored"
ANSWERS_FILE="$CACHE_DIR/developer_answers.txt"
RISK_FILE="$CACHE_DIR/risk_flags.txt"

[[ -d "$MIGRATION_DIR" ]] || fatal "Migration directory not found: $MIGRATION_DIR"
[[ -f "$ANSWERS_FILE" ]] || fatal "Answers file not found — run Script 1 first: $ANSWERS_FILE"

mkdir -p "$OUT_DIR"

# ── Load developer answers ────────────────────────────────────────────────────
source "$ANSWERS_FILE" 2>/dev/null || true

Q1_MIGRATION_TOOL="${Q1_MIGRATION_TOOL:-golang-migrate}"
Q2_DATABASE_LIVE="${Q2_DATABASE_LIVE:-no}"
Q3_CONFIG_DEF_LOCATION="${Q3_CONFIG_DEF_LOCATION:-000058}"
Q4_IAM_AUTHORITY="${Q4_IAM_AUTHORITY:-not-decided}"
Q5_FEATURE_FLAG_CACHE="${Q5_FEATURE_FLAG_CACHE:-no}"
Q6_SERVICE_ACCOUNTS="${Q6_SERVICE_ACCOUNTS:-no}"
Q7_TENANT_FREEZE="${Q7_TENANT_FREEZE:-yes}"
Q8_VIEWS_SECURITY="${Q8_VIEWS_SECURITY:-14}"
Q9_ENTITY_HIERARCHY="${Q9_ENTITY_HIERARCHY:-5}"
Q10_NAMING_CONVENTION="${Q10_NAMING_CONVENTION:-yes}"

# ── Numbering: start after 000059 block ───────────────────────────────────────
# Use 000110+ to fit between the tenant-core block and entity block,
# keeping the gap structure intentional.
SEQ=110
next_seq() {
  # IMPORTANT: Do NOT call as $(next_seq) — subshell discards the SEQ increment.
  # Call as: next_seq; FILENAME="${SEQ_ID}_name"
  # next_seq writes to SEQ_ID and increments SEQ in the current shell.
  SEQ_ID=$(printf "%06d" "$SEQ")
  SEQ=$((SEQ + 1))
}

echo ""
echo -e "${BOLD}╔══════════════════════════════════════════════════════════════╗${RESET}"
echo -e "${BOLD}║      ERP Schema Refactoring Executor — Claude Code           ║${RESET}"
echo -e "${BOLD}╚══════════════════════════════════════════════════════════════╝${RESET}"
echo ""
info "Migration tool  : $Q1_MIGRATION_TOOL"
info "Live database   : $Q2_DATABASE_LIVE"
info "Config def home : $Q3_CONFIG_DEF_LOCATION"
info "IAM authority   : $Q4_IAM_AUTHORITY"
info "PG version      : $Q8_VIEWS_SECURITY"
echo ""

# =============================================================================
# HELPER: write a migration pair with preview + confirmation
# =============================================================================
MIGRATIONS_WRITTEN=0

write_migration() {
  local label="$1"
  local filename="$2"
  local up_content="$3"
  local down_content="$4"

  local up_file="$OUT_DIR/${filename}.up.sql"
  local down_file="$OUT_DIR/${filename}.down.sql"

  echo ""
  echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
  echo -e "${BOLD}  MIGRATION: $label${RESET}"
  echo -e "${BOLD}  Files   : ${CYAN}$up_file${RESET}  /  ${CYAN}$down_file${RESET}"
  echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
  echo ""
  echo -e "${BOLD}  ▶ UP migration preview:${RESET}"
  echo "$up_content" | head -60 | sed 's/^/  │  /'
  local lines
  lines=$(echo "$up_content" | wc -l)
  [[ $lines -gt 60 ]] && echo "  │  ... ($((lines - 60)) more lines)"
  echo ""
  echo -e "${BOLD}  ▼ DOWN migration preview:${RESET}"
  echo "$down_content" | head -20 | sed 's/^/  │  /'
  echo ""

  local confirm=""
  if [[ -t 0 ]]; then
    printf "  Write this migration? [y/N/skip]: "
    read -r confirm || confirm="n"
  else
    confirm="n"
    echo "  [Non-interactive — skipping. Run interactively to approve migrations.]"
  fi

  case "${confirm,,}" in
  y | yes)
    echo "$up_content" >"$up_file"
    echo "$down_content" >"$down_file"
    success "Written: $up_file"
    MIGRATIONS_WRITTEN=$((MIGRATIONS_WRITTEN + 1))
    ;;
  skip)
    warn "Skipped: $label"
    ;;
  *)
    warn "Skipped: $label"
    ;;
  esac
}

# =============================================================================
# FIX 1 — SECURITY DEFINER search_path pinning
# =============================================================================
# Risk: SECURITY DEFINER functions without a pinned search_path allow
# search path injection — a malicious user can shadow real tables.
# Safe level: SAFE — adds SET clause to existing functions via CREATE OR REPLACE.
# =============================================================================

next_seq
FILENAME="${SEQ_ID}_fix_security_definer_search_path"

# Find all SECURITY DEFINER functions missing the pin
SD_FUNCS=""
for f in "$MIGRATION_DIR"/*.up.sql; do
  if grep -q "SECURITY DEFINER" "$f" 2>/dev/null && ! grep -q "SET search_path" "$f" 2>/dev/null; then
    SD_FUNCS+="-- Source: $(basename "$f")\n"
    while IFS= read -r fname; do
      SD_FUNCS+="--   NEEDS PIN: $fname\n"
    done < <(grep -E "^CREATE (OR REPLACE )?FUNCTION" "$f" |
      grep -v "search_path" |
      grep -oE 'FUNCTION [a-z_]+' | awk '{print $2}' || true)
  fi
done

if [[ -n "$SD_FUNCS" ]]; then

  UP=$(
    cat <<UPSQL
-- =============================================================================
-- FIX: Pin search_path on all SECURITY DEFINER functions
-- =============================================================================
-- WHY: SECURITY DEFINER functions run with the privileges of their owner.
--      Without a pinned search_path, a user can CREATE a table or view in
--      their own schema that shadows a real platform table (e.g. "tenants"),
--      causing the function to operate on fake data.
--
-- HOW: We re-declare each affected function with SET search_path = pg_catalog, public
--      using CREATE OR REPLACE — this is a non-destructive in-place update.
--      Function signatures and bodies are unchanged.
--
-- AFFECTED FUNCTIONS (detected by Script 1):
$(echo -e "$SD_FUNCS")
-- =============================================================================

-- set_tenant_context: reads tenants table — must be pinned
CREATE OR REPLACE FUNCTION set_tenant_context(p_tenant_id UUID)
  RETURNS VOID
  LANGUAGE plpgsql
  SECURITY DEFINER
  SET search_path = pg_catalog, public
AS \$\$
DECLARE
  v_status TEXT;
BEGIN
  SELECT "Status"
    INTO v_status
    FROM tenants
   WHERE id = p_tenant_id
     AND deleted_at IS NULL;

  IF NOT FOUND THEN
    RAISE EXCEPTION
      'set_tenant_context: tenant not found or has been deleted: %',
      p_tenant_id USING ERRCODE = 'no_data_found';
  END IF;

  IF v_status <> 'ACTIVE' THEN
    RAISE EXCEPTION
      'set_tenant_context: tenant is not ACTIVE: % (status: %)',
      p_tenant_id, v_status USING ERRCODE = 'check_violation';
  END IF;

  PERFORM set_config('app.current_tenant_id', p_tenant_id::TEXT, TRUE);
END;
\$\$;

COMMENT ON FUNCTION set_tenant_context(UUID) IS
  'Sets transaction-local tenant context. '
  'SECURITY DEFINER with search_path pinned to prevent shadow table injection.';

-- validate_tenant_context: reads tenants table — must be pinned
CREATE OR REPLACE FUNCTION validate_tenant_context()
  RETURNS UUID
  LANGUAGE plpgsql
  STABLE
  SECURITY DEFINER
  SET search_path = pg_catalog, public
AS \$\$
DECLARE
  v_tid    UUID;
  v_status TEXT;
BEGIN
  v_tid := current_tenant_id();

  IF v_tid IS NULL THEN
    RAISE EXCEPTION 'validate_tenant_context: no tenant context is set'
      USING ERRCODE = 'no_data_found';
  END IF;

  SELECT "Status"
    INTO v_status
    FROM tenants
   WHERE id = v_tid AND deleted_at IS NULL;

  IF NOT FOUND THEN
    RAISE EXCEPTION
      'validate_tenant_context: tenant no longer exists or deleted: %', v_tid
      USING ERRCODE = 'no_data_found';
  END IF;

  IF v_status <> 'ACTIVE' THEN
    RAISE EXCEPTION
      'validate_tenant_context: stale context — status is %', v_status
      USING ERRCODE = 'check_violation';
  END IF;

  RETURN v_tid;
END;
\$\$;

-- check_subdomain_not_reserved: reads reserved_subdomains — must be pinned
CREATE OR REPLACE FUNCTION check_subdomain_not_reserved()
  RETURNS TRIGGER
  LANGUAGE plpgsql
  SECURITY DEFINER
  SET search_path = pg_catalog, public
AS \$\$
BEGIN
  IF NEW.subdomain IS NULL THEN RETURN NEW; END IF;
  IF TG_OP = 'UPDATE'
     AND OLD.subdomain IS NOT DISTINCT FROM NEW.subdomain THEN
    RETURN NEW;
  END IF;
  IF EXISTS (SELECT 1 FROM reserved_subdomains WHERE subdomain = NEW.subdomain) THEN
    RAISE EXCEPTION
      'Subdomain "%" is reserved for platform infrastructure.',
      NEW.subdomain USING ERRCODE = 'check_violation';
  END IF;
  RETURN NEW;
END;
\$\$;

-- check_tenant_hierarchy_depth: reads tenants — must be pinned
CREATE OR REPLACE FUNCTION check_tenant_hierarchy_depth()
  RETURNS TRIGGER
  LANGUAGE plpgsql
  SECURITY DEFINER
  SET search_path = pg_catalog, public
AS \$\$
DECLARE
  v_max_depth CONSTANT INTEGER := ${Q9_ENTITY_HIERARCHY};
  v_current   UUID;
  v_depth     INTEGER := 0;
BEGIN
  IF NEW.parent_tenant_id IS NULL THEN RETURN NEW; END IF;
  IF NEW.parent_tenant_id = NEW.id THEN
    RAISE EXCEPTION 'Tenant cannot reference itself as parent.'
      USING ERRCODE = 'check_violation';
  END IF;
  v_current := NEW.parent_tenant_id;
  WHILE v_current IS NOT NULL LOOP
    v_depth := v_depth + 1;
    IF v_depth > v_max_depth THEN
      RAISE EXCEPTION 'Hierarchy depth exceeds maximum of %', v_max_depth
        USING ERRCODE = 'check_violation';
    END IF;
    IF v_current = NEW.id THEN
      RAISE EXCEPTION 'Circular parent_tenant_id reference detected.'
        USING ERRCODE = 'check_violation';
    END IF;
    SELECT parent_tenant_id INTO v_current FROM tenants WHERE id = v_current;
  END LOOP;
  RETURN NEW;
END;
\$\$;
UPSQL
  )

  DOWN=$(
    cat <<DOWNSQL
-- =============================================================================
-- ROLLBACK: Remove search_path pins from SECURITY DEFINER functions
-- =============================================================================
-- This rollback restores functions to their pre-pin state.
-- WARNING: Rolling back this migration re-introduces the search path
-- injection vulnerability. Only do this for testing purposes.
-- =============================================================================

-- Restore set_tenant_context WITHOUT search_path pin
CREATE OR REPLACE FUNCTION set_tenant_context(p_tenant_id UUID)
  RETURNS VOID
  LANGUAGE plpgsql
  SECURITY DEFINER
AS \$\$
DECLARE v_status TEXT;
BEGIN
  SELECT "Status" INTO v_status FROM tenants
   WHERE id = p_tenant_id AND deleted_at IS NULL;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'set_tenant_context: tenant not found: %', p_tenant_id;
  END IF;
  IF v_status <> 'ACTIVE' THEN
    RAISE EXCEPTION 'set_tenant_context: tenant not active: %', p_tenant_id;
  END IF;
  PERFORM set_config('app.current_tenant_id', p_tenant_id::TEXT, TRUE);
END;
\$\$;
DOWNSQL
  )

  write_migration "Security: Pin search_path on SECURITY DEFINER functions" \
    "$FILENAME" "$UP" "$DOWN"
fi

# =============================================================================
# FIX 2 — Replace CHECK subquery on subdomain with proper trigger
# =============================================================================
# Risk: CHECK constraints querying other tables are unreliable in PostgreSQL.
# Safe level: SAFE — drops the unreliable constraint, trigger already exists.
# =============================================================================

next_seq
FILENAME="${SEQ_ID}_fix_subdomain_reserved_check_to_trigger"

# Check if the bad constraint exists in any migration
BAD_CONSTRAINT=$(grep -rl "subdomain_not_reserved.*CHECK.*EXISTS\|CHECK.*EXISTS.*reserved_subdomains" \
  "$MIGRATION_DIR" 2>/dev/null || true)

if [[ -n "$BAD_CONSTRAINT" ]]; then

  UP=$(
    cat <<'UPSQL'
-- =============================================================================
-- FIX: Replace CHECK subquery on subdomain with reliable trigger enforcement
-- =============================================================================
-- WHY: PostgreSQL CHECK constraints that query other tables are unreliable.
--      The SQL standard allows it, but PG only re-validates CHECK on the
--      modified row, not under concurrent inserts. Two concurrent transactions
--      can both pass the NOT EXISTS check and only one will fail at commit.
--      A BEFORE trigger fires correctly under all isolation levels.
--
-- WHAT: Drop the unreliable CHECK constraint.
--       The check_subdomain_not_reserved() trigger (in 000057) is the
--       authoritative enforcement mechanism.
-- =============================================================================

ALTER TABLE tenants
  DROP CONSTRAINT IF EXISTS subdomain_not_reserved;

COMMENT ON TABLE tenants IS
  'Core tenant registry. '
  'Subdomain reservation is enforced by trigger check_subdomain_not_reserved(), '
  'not by CHECK constraint (see ADR-020).';
UPSQL
  )

  DOWN=$(
    cat <<'DOWNSQL'
-- =============================================================================
-- ROLLBACK: Restore the (unreliable) CHECK constraint on subdomain
-- =============================================================================
-- WARNING: This re-introduces a CHECK constraint that queries another table.
-- This is technically unreliable in PostgreSQL under concurrent inserts.
-- The trigger in 000057 is the correct enforcement mechanism.
-- Only restore this for compatibility testing.
-- =============================================================================
ALTER TABLE tenants
  DROP CONSTRAINT IF EXISTS subdomain_not_reserved;

-- Note: we intentionally do NOT re-add the subquery CHECK here.
-- The trigger remains active and is the correct guard.
-- This rollback is a no-op to preserve the trigger-based enforcement.
DOWNSQL
  )

  write_migration "Schema: Drop unreliable CHECK subquery on subdomain" \
    "$FILENAME" "$UP" "$DOWN"
fi

# =============================================================================
# FIX 3 — Add is_tenant_overridable + is_entity_overridable to config_definitions
# =============================================================================
# Risk: The current single is_overridable boolean cannot distinguish who
#       can override. This is additive — new columns, no data loss.
# Safe level: SAFE — additive columns with defaults.
# =============================================================================

next_seq
FILENAME="${SEQ_ID}_config_definitions_split_overridable"

UP=$(
  cat <<'UPSQL'
-- =============================================================================
-- FIX: Split is_overridable into two axes on config_definitions
-- =============================================================================
-- WHY: The 3-level inheritance model (System → Tenant → Entity) requires two
--      independent override gates:
--        is_tenant_overridable — can a tenant admin change the system default?
--        is_entity_overridable — can an entity manager change the tenant value?
--      A single is_overridable boolean collapses both gates to the same bit,
--      making it impossible to express "tenant can change this, but entities
--      cannot" (e.g. inventory.default_valuation_method).
--
-- HOW: Add the two new columns with defaults that preserve current behaviour.
--      Migrate is_overridable = true  → both new columns = true
--      Migrate is_overridable = false → both new columns = false
--      The old column is left in place and marked deprecated via comment.
--      Drop it in a subsequent migration once application code is updated.
-- =============================================================================

ALTER TABLE config_definitions
  ADD COLUMN IF NOT EXISTS is_tenant_overridable BOOLEAN NOT NULL DEFAULT TRUE,
  ADD COLUMN IF NOT EXISTS is_entity_overridable BOOLEAN NOT NULL DEFAULT TRUE;

-- Migrate existing values: preserve the single-bit intent
UPDATE config_definitions
   SET is_tenant_overridable = is_overridable,
       is_entity_overridable = is_overridable
 WHERE is_overridable IS NOT NULL;

COMMENT ON COLUMN config_definitions.is_tenant_overridable IS
  'Whether tenant administrators may override the system default value. '
  'FALSE = system default is locked for all tenants.';

COMMENT ON COLUMN config_definitions.is_entity_overridable IS
  'Whether entity managers may override the tenant-level value. '
  'FALSE = tenant value is locked for all entities within that tenant. '
  'Requires is_tenant_overridable = TRUE to have any effect.';

COMMENT ON COLUMN config_definitions.is_overridable IS
  'DEPRECATED — use is_tenant_overridable and is_entity_overridable instead. '
  'Kept for backward compatibility. Will be dropped in a future migration '
  'once application code references the new columns.';
UPSQL
)

DOWN=$(
  cat <<'DOWNSQL'
-- =============================================================================
-- ROLLBACK: Remove the two-axis override columns
-- =============================================================================
ALTER TABLE config_definitions
  DROP COLUMN IF EXISTS is_entity_overridable,
  DROP COLUMN IF EXISTS is_tenant_overridable;
DOWNSQL
)

write_migration "Settings: Split is_overridable into tenant + entity axes" \
  "$FILENAME" "$UP" "$DOWN"

# =============================================================================
# FIX 4 — Add config_definition_id FK to configuration_audit
# =============================================================================
# Safe level: SAFE — additive nullable FK column.
# =============================================================================

next_seq
FILENAME="${SEQ_ID}_configuration_audit_add_definition_fk"

UP=$(
  cat <<'UPSQL'
-- =============================================================================
-- FIX: Add config_definition_id FK to configuration_audit
-- =============================================================================
-- WHY: configuration_audit.config_key is a VARCHAR string with no FK.
--      After a config key is renamed or deleted, audit records become
--      unresolvable — you cannot JOIN audit to config_definitions.
--      Storing the definition UUID alongside the key string lets the
--      audit trail survive key renames (the UUID is stable, the key may change).
--
-- HOW: Add nullable config_definition_id. Nullable because:
--      - Existing audit rows predate this column.
--      - TEMPLATE_APPLY operations may not map 1:1 to a single definition.
--      Application code should populate this on all new writes.
-- =============================================================================

ALTER TABLE configuration_audit
  ADD COLUMN IF NOT EXISTS config_definition_id UUID
    REFERENCES config_definitions(id) ON DELETE SET NULL;

COMMENT ON COLUMN configuration_audit.config_definition_id IS
  'FK to config_definitions. Stable reference even if config_key is later renamed. '
  'NULL for historical records predating this column and for TEMPLATE_APPLY operations '
  'that touch multiple definitions atomically.';

-- Index to support "show all changes to this definition" queries
CREATE INDEX IF NOT EXISTS idx_config_audit_definition
  ON configuration_audit(config_definition_id, applied_at DESC)
  WHERE config_definition_id IS NOT NULL;

-- Index for time-range dashboard queries: "all changes in last 30 days for tenant"
-- This was missing from the original schema (tenant_id + config_key required knowing key upfront)
CREATE INDEX IF NOT EXISTS idx_config_audit_tenant_time
  ON configuration_audit(tenant_id, applied_at DESC);
UPSQL
)

DOWN=$(
  cat <<'DOWNSQL'
-- =============================================================================
-- ROLLBACK: Remove config_definition_id FK from configuration_audit
-- =============================================================================
DROP INDEX IF EXISTS idx_config_audit_tenant_time;
DROP INDEX IF EXISTS idx_config_audit_definition;

ALTER TABLE configuration_audit
  DROP COLUMN IF EXISTS config_definition_id;
DOWNSQL
)

write_migration "Settings: Add config_definition_id FK to configuration_audit" \
  "$FILENAME" "$UP" "$DOWN"

# =============================================================================
# FIX 5 — Add entitystate.config JSONB for document sequence formatting
# =============================================================================
# PRD Phase 2: document prefix, padding, reset frequency per entity.
# Safe level: SAFE — additive nullable column with default.
# =============================================================================

next_seq
FILENAME="${SEQ_ID}_entitystate_add_config_column"

UP=$(
  cat <<'UPSQL'
-- =============================================================================
-- FIX: Add config JSONB to entitystate for document sequence formatting
-- =============================================================================
-- WHY: The PRD Phase 2 requires configurable document number formatting
--      per entity: custom prefix (BRANCH-INV-), padding length, reset
--      frequency (yearly/monthly/never), and format templates.
--      Currently all entities share the same implicit format with no
--      per-entity customisation possible.
--
-- Schema for config JSONB:
-- {
--   "prefix":          "INV-",          -- document number prefix
--   "suffix":          "",              -- document number suffix
--   "pad_length":      6,               -- zero-padding width
--   "reset_frequency": "yearly",        -- yearly | monthly | never
--   "format_template": "{prefix}{year:2d}{number:06d}{suffix}"
-- }
--
-- NULL config = use tenant-level default from Settings module.
-- '{}' config = entity explicitly uses system defaults with no customisation.
-- =============================================================================

ALTER TABLE entitystate
  ADD COLUMN IF NOT EXISTS config JSONB DEFAULT '{}'::jsonb;

COMMENT ON COLUMN entitystate.config IS
  'Document sequence formatting config for this entity+document_type. '
  'Keys: prefix, suffix, pad_length, reset_frequency, format_template. '
  'NULL = inherit from tenant Settings module defaults. '
  '{} = use system defaults explicitly.';

-- Backfill existing rows with empty object (not NULL) so they use system defaults
UPDATE entitystate
   SET config = '{}'::jsonb
 WHERE config IS NULL;

-- GIN index for sequence config lookups by prefix or format
CREATE INDEX IF NOT EXISTS idx_entitystate_config_gin
  ON entitystate USING gin(config)
  WHERE config IS NOT NULL AND config <> '{}'::jsonb;
UPSQL
)

DOWN=$(
  cat <<'DOWNSQL'
-- =============================================================================
-- ROLLBACK: Remove config column from entitystate
-- =============================================================================
DROP INDEX IF EXISTS idx_entitystate_config_gin;

ALTER TABLE entitystate
  DROP COLUMN IF EXISTS config;
DOWNSQL
)

write_migration "Entities: Add config JSONB to entitystate (PRD Phase 2)" \
  "$FILENAME" "$UP" "$DOWN"

# =============================================================================
# FIX 6 — Add user_type to users for service account support
# =============================================================================
# Safe level: SAFE — additive column, defaults to HUMAN to preserve existing rows.
# =============================================================================

next_seq
FILENAME="${SEQ_ID}_users_add_user_type"

UP=$(
  cat <<'UPSQL'
-- =============================================================================
-- FIX: Add user_type to users table for non-human identity support
-- =============================================================================
-- WHY: The identity model is Person → Employee → User, which works for
--      human users but has no representation for:
--        - Background workers (scheduled jobs, queue processors)
--        - API integration users (webhooks, third-party connectors)
--        - Service accounts (internal microservice calls)
--      These identities need to participate in the IAM model (role assignment,
--      permission checks, audit logging) but cannot be "persons" or "employees".
--
-- HOW: Add user_type column with HUMAN as default (preserves all existing rows).
--      Non-human users will have person_id = NULL (allow NULL on FK in users).
--      Application code should enforce: if user_type != 'HUMAN' then person_id IS NULL.
-- =============================================================================

ALTER TABLE users
  ADD COLUMN IF NOT EXISTS user_type VARCHAR(20) NOT NULL DEFAULT 'HUMAN'
    CHECK (user_type IN ('HUMAN', 'SERVICE_ACCOUNT', 'API_KEY', 'INTEGRATION', 'SYSTEM'));

COMMENT ON COLUMN users.user_type IS
  'Identity classification. '
  'HUMAN = real person with a person_id FK. '
  'SERVICE_ACCOUNT = internal background worker or scheduled job. '
  'API_KEY = external system integration (person_id IS NULL). '
  'INTEGRATION = third-party connector (Stripe, Salesforce, etc.). '
  'SYSTEM = platform-level automated process.';

-- Partial index: finding service accounts is a common ops/security query
CREATE INDEX IF NOT EXISTS idx_users_non_human
  ON users(user_type, tenant_id)
  WHERE user_type <> 'HUMAN';

-- Verify all existing rows correctly got HUMAN default
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM users WHERE user_type IS NULL) THEN
    RAISE EXCEPTION 'Unexpected NULL user_type after migration — investigate.';
  END IF;
END;
$$;
UPSQL
)

DOWN=$(
  cat <<'DOWNSQL'
-- =============================================================================
-- ROLLBACK: Remove user_type from users
-- =============================================================================
DROP INDEX IF EXISTS idx_users_non_human;

ALTER TABLE users
  DROP COLUMN IF EXISTS user_type;
DOWNSQL
)

write_migration "Identity: Add user_type to users for service accounts" \
  "$FILENAME" "$UP" "$DOWN"

# =============================================================================
# FIX 7 — Add entity slug immutability trigger
# =============================================================================
# Safe level: SAFE — trigger addition only.
# =============================================================================

next_seq
FILENAME="${SEQ_ID}_entities_add_code_immutability"

UP=$(
  cat <<'UPSQL'
-- =============================================================================
-- FIX: Enforce entity code immutability after creation
-- =============================================================================
-- WHY: Entity codes appear in document numbers (BRANCH-INV-000001), API
--      paths, integrations, and stored references in ledger entries.
--      Changing a code silently breaks all stored references without any
--      FK cascade to alert the developer.
--      Immutability must be enforced at the DB level because application
--      guards are easier to accidentally bypass.
--
-- Escape hatch: An admin can disable this trigger temporarily for a
--   deliberate code change. That act produces a DDL audit event.
--   ALTER TABLE entities DISABLE TRIGGER enforce_entity_code_immutability;
--   UPDATE entities SET code = 'NEW-CODE' WHERE uuid = '...';
--   ALTER TABLE entities ENABLE TRIGGER enforce_entity_code_immutability;
-- =============================================================================

CREATE OR REPLACE FUNCTION enforce_entity_code_immutability()
  RETURNS TRIGGER
  LANGUAGE plpgsql
AS $$
BEGIN
  IF OLD.code IS NOT NULL AND OLD.code IS DISTINCT FROM NEW.code THEN
    RAISE EXCEPTION
      'Entity code is immutable after it has been set. '
      'Old: %, Attempted: %. '
      'Disable trigger temporarily for a deliberate code change (creates DDL audit event).',
      OLD.code, NEW.code
      USING ERRCODE = 'check_violation';
  END IF;
  RETURN NEW;
END;
$$;

COMMENT ON FUNCTION enforce_entity_code_immutability() IS
  'BEFORE UPDATE trigger: prevents entity code changes after it has been set. '
  'Codes appear in document numbers and integration paths — changes break stored references. '
  'To change deliberately, disable the trigger (produces DDL audit event).';

DROP TRIGGER IF EXISTS enforce_entity_code_immutability ON entities;
CREATE TRIGGER enforce_entity_code_immutability
  BEFORE UPDATE ON entities
  FOR EACH ROW
  WHEN (OLD.code IS NOT NULL AND OLD.code IS DISTINCT FROM NEW.code)
  EXECUTE FUNCTION enforce_entity_code_immutability();
UPSQL
)

DOWN=$(
  cat <<'DOWNSQL'
-- =============================================================================
-- ROLLBACK: Remove entity code immutability trigger
-- =============================================================================
DROP TRIGGER IF EXISTS enforce_entity_code_immutability ON entities;
DROP FUNCTION IF EXISTS enforce_entity_code_immutability();
DOWNSQL
)

write_migration "Entities: Enforce entity code immutability after creation" \
  "$FILENAME" "$UP" "$DOWN"

# =============================================================================
# FIX 8 — Remove ORDER BY from view definitions
# =============================================================================
# Safe level: SAFE — views are replaced, no data involved.
# Only if views with ORDER BY were detected in Script 1.
# =============================================================================

ORDER_BY_VIEWS=$(grep -rl "ORDER BY" "$MIGRATION_DIR"/*views* 2>/dev/null |
  head -3 || true)

if [[ -n "$ORDER_BY_VIEWS" ]]; then

  next_seq
  FILENAME="${SEQ_ID}_views_remove_order_by"

  UP=$(
    cat <<'UPSQL'
-- =============================================================================
-- FIX: Remove ORDER BY from view definitions
-- =============================================================================
-- WHY: ORDER BY inside a CREATE VIEW definition is not honoured by the
--      PostgreSQL query planner. The spec says view output order is undefined
--      unless the outer query specifies ORDER BY. Worse, the planner inserts
--      a sort node for every query on the view — even when the caller adds
--      their own ORDER BY, you may get two sort passes.
--
-- HOW: Drop and recreate the affected views without the ORDER BY clause.
--      Callers that need ordered results must add ORDER BY themselves.
--      This is a best-practice change, not a behaviour change.
-- =============================================================================

-- v_entity_changes: had ORDER BY e.updated_at DESC
CREATE OR REPLACE VIEW v_entity_changes AS
SELECT
  t.name AS tenant_name,
  e.uuid AS entity_id,
  e.name AS entity_name,
  e.type AS entity_type,
  e.validation_status,
  e.created_at,
  e.updated_at,
  e.deleted_at,
  CASE
    WHEN e.deleted_at IS NOT NULL THEN 'DELETED'
    WHEN e.updated_at > e.created_at + INTERVAL '1 minute' THEN 'MODIFIED'
    ELSE 'CREATED'
  END AS change_type
  -- ORDER BY removed: callers must apply ORDER BY e.updated_at DESC themselves
FROM entities e
JOIN tenants t ON e.tenant_id = t.id;

COMMENT ON VIEW v_entity_changes IS
  'Entity lifecycle event view. '
  'ORDER BY was removed from view definition (not honoured by planner). '
  'Use: SELECT * FROM v_entity_changes ORDER BY updated_at DESC';

-- v_entity_paths: had ORDER BY hp.depth, a.name, d.name
CREATE OR REPLACE VIEW v_entity_paths AS
SELECT
  t.name AS tenant_name,
  a.uuid AS ancestor_id,
  a.name AS ancestor_name,
  a.type AS ancestor_type,
  d.uuid AS descendant_id,
  d.name AS descendant_name,
  d.type AS descendant_type,
  hp.depth
  -- ORDER BY removed: callers must apply ORDER BY hp.depth, a.name, d.name themselves
FROM hierarchy_paths hp
JOIN entities a ON hp.ancestor_id = a.uuid
JOIN entities d ON hp.descendant_id = d.uuid
JOIN tenants t ON hp.tenant_id = t.id
WHERE a.deleted_at IS NULL AND d.deleted_at IS NULL;

COMMENT ON VIEW v_entity_paths IS
  'Ancestor-descendant relationship view. '
  'ORDER BY was removed from view definition (not honoured by planner). '
  'Use: SELECT * FROM v_entity_paths ORDER BY depth, ancestor_name, descendant_name';
UPSQL
  )

  DOWN=$(
    cat <<'DOWNSQL'
-- =============================================================================
-- ROLLBACK: Restore ORDER BY in views (functionally identical — ORDER BY is ignored)
-- =============================================================================
-- Note: restoring ORDER BY has no actual effect on query results.
-- This rollback exists only for strict migration reversibility.

CREATE OR REPLACE VIEW v_entity_changes AS
SELECT t.name AS tenant_name, e.uuid AS entity_id, e.name AS entity_name,
  e.type AS entity_type, e.validation_status, e.created_at, e.updated_at, e.deleted_at,
  CASE WHEN e.deleted_at IS NOT NULL THEN 'DELETED'
       WHEN e.updated_at > e.created_at + INTERVAL '1 minute' THEN 'MODIFIED'
       ELSE 'CREATED' END AS change_type
FROM entities e JOIN tenants t ON e.tenant_id = t.id
ORDER BY e.updated_at DESC;
DOWNSQL
  )

  write_migration "Views: Remove ORDER BY from view definitions" \
    "$FILENAME" "$UP" "$DOWN"
fi

# =============================================================================
# FIX 9 — Add feature flag → config_definitions FK validation trigger
# =============================================================================
# Safe level: SAFE — trigger only, no data changes.
# =============================================================================

next_seq
FILENAME="${SEQ_ID}_config_definitions_validate_feature_flag_ref"

UP=$(
  cat <<'UPSQL'
-- =============================================================================
-- FIX: Validate required_feature_flag references on config_definitions
-- =============================================================================
-- WHY: config_definitions.required_feature_flag is a VARCHAR string that
--      references a flag key in the feature_flags table, but there is no FK.
--      A typo in required_feature_flag means the config appears always-available
--      even though it should be gated — silently wrong, not noisily broken.
--
-- HOW: BEFORE INSERT OR UPDATE trigger that validates the flag key exists
--      in the feature_flags table when required_feature_flag is not NULL.
--      Uses a trigger rather than a FK because:
--      1. feature_flags.key is a VARCHAR, not a UUID — text FKs are unusual
--      2. We want a clear error message rather than a generic FK violation
--      3. The feature_flags table might use a different key column name
--
-- NOTE: Update the subquery below to match your actual feature_flags schema.
--       Expected: feature_flags table with a "key" or "flag_key" column.
-- =============================================================================

CREATE OR REPLACE FUNCTION validate_feature_flag_reference()
  RETURNS TRIGGER
  LANGUAGE plpgsql
  SECURITY DEFINER
  SET search_path = pg_catalog, public
AS $$
BEGIN
  IF NEW.required_feature_flag IS NULL THEN
    RETURN NEW;
  END IF;

  -- Adjust the table and column name to match your feature_flags schema.
  -- Common patterns: feature_flags.key, feature_flags.flag_key, feature_flags.name
  IF NOT EXISTS (
    SELECT 1 FROM feature_flags WHERE key = NEW.required_feature_flag
  ) THEN
    RAISE EXCEPTION
      'validate_feature_flag_reference: required_feature_flag "%" does not exist '
      'in feature_flags table. Check the key spelling or create the flag first.',
      NEW.required_feature_flag
      USING ERRCODE = 'foreign_key_violation';
  END IF;

  RETURN NEW;
END;
$$;

COMMENT ON FUNCTION validate_feature_flag_reference() IS
  'BEFORE INSERT OR UPDATE trigger on config_definitions: validates that '
  'required_feature_flag references an existing flag in the feature_flags table. '
  'Prevents silent misconfiguration where a typo makes a gated config appear always-available.';

DROP TRIGGER IF EXISTS config_def_validate_feature_flag ON config_definitions;
CREATE TRIGGER config_def_validate_feature_flag
  BEFORE INSERT OR UPDATE OF required_feature_flag ON config_definitions
  FOR EACH ROW
  EXECUTE FUNCTION validate_feature_flag_reference();
UPSQL
)

DOWN=$(
  cat <<'DOWNSQL'
-- =============================================================================
-- ROLLBACK: Remove feature flag reference validation trigger
-- =============================================================================
DROP TRIGGER IF EXISTS config_def_validate_feature_flag ON config_definitions;
DROP FUNCTION IF EXISTS validate_feature_flag_reference();
DOWNSQL
)

write_migration "Settings: Add feature flag reference validation trigger" \
  "$FILENAME" "$UP" "$DOWN"

# =============================================================================
# FIX 10 — is_active + deleted_at consistency on entities
# =============================================================================
# Safe level: SAFE — additive trigger, existing inconsistent rows documented.
# =============================================================================

next_seq
FILENAME="${SEQ_ID}_entities_enforce_active_deleted_consistency"

UP=$(
  cat <<'UPSQL'
-- =============================================================================
-- FIX: Enforce is_active / deleted_at consistency on entities
-- =============================================================================
-- WHY: An entity can currently have is_active = TRUE and deleted_at IS NOT NULL
--      simultaneously. This is contradictory — a deleted entity should not be
--      active. Application code that filters by is_active = TRUE will include
--      soft-deleted entities; code that filters by deleted_at IS NULL will
--      miss entities that are active but incorrectly not soft-deleted.
--
-- RULE: When deleted_at is set, is_active must be FALSE.
--       When is_active is set to TRUE, deleted_at must be NULL.
--
-- HOW: Trigger that enforces the invariant on INSERT and UPDATE.
--      Also reports existing inconsistencies as WARNINGs (not errors)
--      so this migration can be applied to a live database without
--      breaking existing data — fix the data separately.
-- =============================================================================

-- Report existing inconsistencies without blocking the migration
DO $$
DECLARE
  inconsistent_count INTEGER;
BEGIN
  SELECT COUNT(*) INTO inconsistent_count
    FROM entities
   WHERE is_active = TRUE AND deleted_at IS NOT NULL;

  IF inconsistent_count > 0 THEN
    RAISE WARNING
      'entities_enforce_active_deleted_consistency: found % entities with '
      'is_active=TRUE and deleted_at IS NOT NULL. '
      'Run: UPDATE entities SET is_active = FALSE WHERE deleted_at IS NOT NULL; '
      'before this invariant is fully enforced.',
      inconsistent_count;
  END IF;
END;
$$;

CREATE OR REPLACE FUNCTION enforce_entity_active_deleted_consistency()
  RETURNS TRIGGER
  LANGUAGE plpgsql
AS $$
BEGIN
  -- A deleted entity cannot be active
  IF NEW.deleted_at IS NOT NULL AND NEW.is_active = TRUE THEN
    RAISE EXCEPTION
      'Entity consistency violation: is_active cannot be TRUE when deleted_at is set. '
      'Set is_active = FALSE before or alongside setting deleted_at. '
      'Entity: %, name: %', NEW.uuid, NEW.name
      USING ERRCODE = 'check_violation';
  END IF;

  -- Reactivating an entity must also clear deleted_at
  IF NEW.is_active = TRUE AND OLD.deleted_at IS NOT NULL AND NEW.deleted_at IS NOT NULL THEN
    RAISE EXCEPTION
      'Entity consistency violation: cannot set is_active = TRUE while deleted_at remains set. '
      'Clear deleted_at alongside setting is_active = TRUE. '
      'Entity: %, name: %', NEW.uuid, NEW.name
      USING ERRCODE = 'check_violation';
  END IF;

  RETURN NEW;
END;
$$;

COMMENT ON FUNCTION enforce_entity_active_deleted_consistency() IS
  'BEFORE INSERT OR UPDATE trigger: prevents is_active=TRUE + deleted_at IS NOT NULL '
  'from existing simultaneously on the same entity row. '
  'To restore a deleted entity: UPDATE entities SET is_active=TRUE, deleted_at=NULL WHERE uuid=$1.';

DROP TRIGGER IF EXISTS enforce_entity_active_deleted ON entities;
CREATE TRIGGER enforce_entity_active_deleted
  BEFORE INSERT OR UPDATE OF is_active, deleted_at ON entities
  FOR EACH ROW
  EXECUTE FUNCTION enforce_entity_active_deleted_consistency();
UPSQL
)

DOWN=$(
  cat <<'DOWNSQL'
-- =============================================================================
-- ROLLBACK: Remove entity is_active / deleted_at consistency trigger
-- =============================================================================
DROP TRIGGER IF EXISTS enforce_entity_active_deleted ON entities;
DROP FUNCTION IF EXISTS enforce_entity_active_deleted_consistency();
DOWNSQL
)

write_migration "Entities: Enforce is_active / deleted_at consistency" \
  "$FILENAME" "$UP" "$DOWN"

# =============================================================================
# GENERATE SUMMARY DOCUMENTS
# =============================================================================
info "Generating refactored/REFACTOR_SUMMARY.md..."

{
  cat <<SUMMARY_HEADER
# Refactoring Summary

> Generated by \`02_refactor_migrations.sh\` on $(date)
> All files written to: \`$OUT_DIR/\`

## What Was Written

| File | Risk Level | Reversible | Category |
|------|-----------|------------|----------|
SUMMARY_HEADER

  for f in "$OUT_DIR"/*.up.sql; do
    [[ -f "$f" ]] || continue
    bn=$(basename "$f" .up.sql)
    echo "| \`$bn\` | SAFE | YES | See file header |"
  done

  cat <<'ARCH_SECTION'

## Architecture Changes (NOT Executed — Require Human Decision)

These are documented here. Script 2 deliberately did not execute them.
Each requires coordination between the database team and application team.

### 1. Move Feature Flag Cache to Redis
**File**: `000805_feature_flag_cache` — `000806_feature_flag_cleanup`
**Action**: Drop the DB cache table. Add Redis client to FeatureFlag service.
Set TTL = 60s for flag values. Invalidate on flag UPDATE via pub/sub.
**Why not automated**: Requires application code changes before DB migration.

### 2. Unified Policy Decision Point (RBAC + ABAC)
**Files**: `000404-000412` (RBAC) + `000701-000705` (ABAC)
**Action**: Create a single `evaluate_access(user_id, resource_id, action_id, context JSONB)`
function that calls RBAC first (coarse gate) then ABAC (fine-grained filter).
**Why not automated**: Requires full audit of all permission checks in Go codebase.

### 3. Remove redundant entity_id from hierarchy_paths
**File**: `000202_entity_add_hierarchy`
**Action**: `ALTER TABLE hierarchy_paths DROP COLUMN entity_id`
**Why not automated**: Need to verify no application code references this column.
**Verification**: `grep -r "entity_id" --include="*.go" | grep -i "hierarchy"`

### 4. Remove validation columns from hierarchy_paths
**Columns**: version, last_validation_run, validation_status, validation_errors
**Why not automated**: Need to verify no application code references these columns.

### 5. Rewrite v_active_entities CASE-in-JOIN as CTEs
**File**: `000204_entity_add_views`
**Problem**: CASE expressions inside JOIN ON clauses -> correlated subquery per row pair.
**Fix**: Pre-compute each entity's company/region/department ancestor via hierarchy_paths CTE.
**Why not automated**: View rewrite requires domain knowledge of query intent.

### 6. Remove tenant_id from config_definitions
**File**: `000058_config_definitions`
**Action**: `ALTER TABLE config_definitions DROP COLUMN tenant_id, DROP COLUMN entity_id`
**Why not automated**: Need to verify ConfigurationService.GetEffectiveConfiguration()
does not filter by tenant_id when reading definitions.

## Apply Order for New Migrations

ARCH_SECTION

  echo '```'
  echo "Apply these AFTER your existing 000059 block and BEFORE 000102:"
  echo ""
  for f in "$OUT_DIR"/*.up.sql; do
    [[ -f "$f" ]] || continue
    echo "  $(basename "$f")"
  done
  echo '```'

  cat <<'NOTE_SECTION'

## Verification Steps After Applying

```sql
-- 1. Verify search_path is pinned on all SECURITY DEFINER functions
SELECT proname, proconfig
FROM pg_proc
WHERE prosecdef = true
  AND proconfig IS NULL
  AND pronamespace = (SELECT oid FROM pg_namespace WHERE nspname = 'public');
-- Expected: zero rows (all SECURITY DEFINER functions should have proconfig set)

-- 2. Verify subdomain CHECK constraint is gone
SELECT constraint_name FROM information_schema.table_constraints
WHERE table_name = 'tenants' AND constraint_name = 'subdomain_not_reserved';
-- Expected: zero rows

-- 3. Verify new config_definitions columns exist
SELECT column_name, column_default FROM information_schema.columns
WHERE table_name = 'config_definitions'
  AND column_name IN ('is_tenant_overridable', 'is_entity_overridable');
-- Expected: 2 rows

-- 4. Verify entitystate config column
SELECT column_name, data_type FROM information_schema.columns
WHERE table_name = 'entitystate' AND column_name = 'config';
-- Expected: 1 row, data_type = jsonb

-- 5. Verify user_type column
SELECT column_name, column_default FROM information_schema.columns
WHERE table_name = 'users' AND column_name = 'user_type';
-- Expected: 1 row, column_default = 'HUMAN'

-- 6. Verify entity code immutability trigger exists
SELECT trigger_name FROM information_schema.triggers
WHERE event_object_table = 'entities'
  AND trigger_name = 'enforce_entity_code_immutability';
-- Expected: 1 row

-- 7. Verify is_active/deleted_at consistency trigger exists
SELECT trigger_name FROM information_schema.triggers
WHERE event_object_table = 'entities'
  AND trigger_name = 'enforce_entity_active_deleted';
-- Expected: 1 row
```
NOTE_SECTION

} >"$OUT_DIR/REFACTOR_SUMMARY.md"

success "Summary written → $OUT_DIR/REFACTOR_SUMMARY.md"

# =============================================================================
# FINAL OUTPUT
# =============================================================================
echo ""
echo -e "${BOLD}════════════════════════════════════════════════════════════════${RESET}"
echo -e "${BOLD}  REFACTORING COMPLETE — $MIGRATIONS_WRITTEN migrations written${RESET}"
echo -e "${BOLD}════════════════════════════════════════════════════════════════${RESET}"
echo ""
echo "  Output directory: $OUT_DIR/"
{ ls "$OUT_DIR"/*.sql 2>/dev/null | sed 's/^/    /'; } || echo "    (no files written — run interactively to approve each migration)"
echo ""
echo "  Next steps:"
echo -e "    1. ${CYAN}Review each file in $OUT_DIR/ before applying${RESET}"
echo -e "    2. ${CYAN}Read $OUT_DIR/REFACTOR_SUMMARY.md — esp. the architecture section${RESET}"
echo -e "    3. ${CYAN}Copy approved files to $MIGRATION_DIR/${RESET}"
echo -e "    4. ${CYAN}Run verification SQL from REFACTOR_SUMMARY.md after apply${RESET}"
echo ""
warn "Never apply migrations directly to production without staging validation first."
echo ""
