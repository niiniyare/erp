#!/usr/bin/env bash

# =============================================================================
# Awo ERP — PostgreSQL Schema Security & Tenant Isolation Inspector
#
# Usage:
#   ./inspect.sh
#   DB_URL="postgresql://user:pass@host:5432/db" ./inspect.sh
#   ./inspect.sh | jq '.summary'
#   ./inspect.sh | jq '.tables[] | select(.analysis.rls_disabled)'
#   ./inspect.sh | jq '.tables[] | select(.analysis.severity >= 5)'
# =============================================================================

set -euo pipefail

DB_URL="${DB_URL:-postgresql://admin:admin@localhost:5432/ledger?sslmode=disable}"

echo " Inspecting PostgreSQL schema: $DB_URL" >&2

# Run a query and always return valid JSON.
# psql -At returns empty string (not "null") when SQL yields NULL —
# we normalise that to "[]" so jq never receives a non-JSON value.
query() {
  local sql="$1"
  local default="${2:-[]}"
  local out
  out=$(psql "$DB_URL" -At -c "$sql" 2>/dev/null) || true
  if [[ -z "$out" || "$out" == "" ]]; then
    echo "$default"
  else
    echo "$out"
  fi
}

# ── 1. Tables + Columns ───────────────────────────────────────────────────────
echo "  → tables & columns..." >&2
tables=$(query "
SELECT COALESCE(json_agg(row_to_json(t) ORDER BY t.table_schema, t.table_name), '[]')
FROM (
    SELECT
        c.table_schema,
        c.table_name,
        EXISTS (
            SELECT 1
            FROM information_schema.columns x
            WHERE x.table_schema = c.table_schema
              AND x.table_name   = c.table_name
              AND x.column_name  = 'tenant_id'
        ) AS has_tenant_id,
        COALESCE((
            SELECT json_agg(json_build_object(
                'name',     x.column_name,
                'type',     x.data_type,
                'nullable', x.is_nullable,
                'default',  x.column_default
            ) ORDER BY x.ordinal_position)
            FROM information_schema.columns x
            WHERE x.table_schema = c.table_schema
              AND x.table_name   = c.table_name
        ), '[]') AS columns
    FROM information_schema.tables c
    WHERE c.table_schema NOT IN ('pg_catalog', 'information_schema')
      AND c.table_type = 'BASE TABLE'
) t;
")

# ── 2. Foreign Keys ───────────────────────────────────────────────────────────
echo "  → foreign keys..." >&2
fks=$(query "
SELECT COALESCE(json_agg(json_build_object(
    'constraint',    tc.constraint_name,
    'source_schema', tc.table_schema,
    'source_table',  tc.table_name,
    'source_column', kcu.column_name,
    'target_table',  ccu.table_name,
    'target_column', ccu.column_name,
    'on_delete',     rc.delete_rule,
    'on_update',     rc.update_rule
) ORDER BY tc.table_name, kcu.column_name), '[]')
FROM       information_schema.table_constraints     tc
JOIN       information_schema.key_column_usage      kcu ON kcu.constraint_name = tc.constraint_name
JOIN       information_schema.constraint_column_usage ccu ON ccu.constraint_name = tc.constraint_name
JOIN       information_schema.referential_constraints rc  ON rc.constraint_name  = tc.constraint_name
WHERE tc.constraint_type = 'FOREIGN KEY';
")

# ── 3. RLS Status ─────────────────────────────────────────────────────────────
echo "  → RLS status..." >&2
rls=$(query "
SELECT COALESCE(json_agg(json_build_object(
    'schema',      n.nspname,
    'table',       c.relname,
    'rls_enabled', c.relrowsecurity,
    'rls_forced',  c.relforcerowsecurity
) ORDER BY n.nspname, c.relname), '[]')
FROM pg_class     c
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE c.relkind = 'r'
  AND n.nspname NOT IN ('pg_catalog', 'information_schema');
")

# ── 4. RLS Policies ───────────────────────────────────────────────────────────
echo "  → RLS policies..." >&2
policies=$(query "
SELECT COALESCE(json_agg(json_build_object(
    'schema',     schemaname,
    'table',      tablename,
    'policy',     policyname,
    'permissive', permissive,
    'roles',      roles,
    'command',    cmd,
    'using',      qual,
    'check',      with_check
) ORDER BY tablename, policyname), '[]')
FROM pg_policies;
")

# ── 5. Indexes ────────────────────────────────────────────────────────────────
echo "  → indexes..." >&2
indexes=$(query "
SELECT COALESCE(json_agg(json_build_object(
    'schema',  n.nspname,
    'table',   t.relname,
    'index',   i.relname,
    'unique',  ix.indisunique,
    'primary', ix.indisprimary,
    'columns', COALESCE((
        SELECT json_agg(a.attname ORDER BY k.ord)
        FROM   unnest(ix.indkey) WITH ORDINALITY AS k(attnum, ord)
        JOIN   pg_attribute a ON a.attrelid = t.oid AND a.attnum = k.attnum
        WHERE  k.attnum > 0
    ), '[]')
) ORDER BY t.relname, i.relname), '[]')
FROM pg_index     ix
JOIN pg_class     t  ON t.oid = ix.indrelid
JOIN pg_class     i  ON i.oid = ix.indexrelid
JOIN pg_namespace n  ON n.oid = t.relnamespace
WHERE n.nspname NOT IN ('pg_catalog', 'information_schema');
")

# ── 6. Views (security_invoker status) ───────────────────────────────────────
echo "  → views..." >&2
views=$(query "
SELECT COALESCE(json_agg(json_build_object(
    'schema', n.nspname,
    'view',   c.relname,
    'security_invoker', (
        c.reloptions IS NOT NULL
        AND EXISTS (
            SELECT 1
            FROM unnest(c.reloptions) AS opt
            WHERE opt ILIKE 'security_invoker=true'
        )
    )
) ORDER BY n.nspname, c.relname), '[]')
FROM pg_class     c
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE c.relkind = 'v'
  AND n.nspname NOT IN ('pg_catalog', 'information_schema');
")

# ── 7. Check Constraints ──────────────────────────────────────────────────────
echo "  → check constraints..." >&2
checks=$(query "
SELECT COALESCE(json_agg(json_build_object(
    'table',      tc.table_name,
    'constraint', tc.constraint_name,
    'column',     kcu.column_name,
    'definition', cc.check_clause
) ORDER BY tc.table_name, tc.constraint_name), '[]')
FROM      information_schema.table_constraints     tc
JOIN      information_schema.check_constraints     cc  ON cc.constraint_name  = tc.constraint_name
LEFT JOIN information_schema.key_column_usage      kcu ON kcu.constraint_name = tc.constraint_name
WHERE tc.constraint_type = 'CHECK'
  AND tc.table_schema NOT IN ('pg_catalog', 'information_schema');
")

# ── 8. Analyse ────────────────────────────────────────────────────────────────
echo "  → analysing..." >&2

jq -n \
  --argjson T "$tables" \
  --argjson F "$fks" \
  --argjson R "$rls" \
  --argjson P "$policies" \
  --argjson I "$indexes" \
  --argjson V "$views" \
  --argjson C "$checks" \
  '

# ── Helper functions ──────────────────────────────────────────────────────────

def safe_array: if type == "array" then . else [] end;

def policies_for(tbl):   $P | safe_array | map(select(.table   == tbl));
def rls_for(tbl):        $R | safe_array | map(select(.table   == tbl)) | .[0] // null;
def indexes_for(tbl):    $I | safe_array | map(select(.table   == tbl));
def checks_for(tbl):     $C | safe_array | map(select(.table   == tbl));

def has_tenant_index(tbl):
  indexes_for(tbl)
  | map(select(
      .columns != null
      and (.columns | safe_array | index("tenant_id")) != null
    ))
  | length > 0;

# Check every policy on a table for common mistakes
def policy_problems(tbl):
  policies_for(tbl) | map(
    . as $pol |
    {
      policy:  $pol.policy,
      command: $pol.command,
      # SELECT / ALL / UPDATE need a USING clause
      missing_using: (
        ($pol.command | IN("SELECT","ALL","UPDATE"))
        and (($pol.using  // "") == "")
      ),
      # INSERT / ALL / UPDATE need a WITH CHECK clause
      missing_check: (
        ($pol.command | IN("INSERT","ALL","UPDATE"))
        and (($pol.check // "") == "")
      ),
      # Policy does not reference tenant_id anywhere
      no_tenant_filter: (
        (($pol.using  // "") | test("tenant_id") | not) and
        (($pol.check  // "") | test("tenant_id") | not)
      ),
      # USING (true) — completely open, no filter
      open_using: (
        ($pol.using != null) and
        ($pol.using | test("^\\s*true\\s*$"))
      )
    }
    | select(
        .missing_using or .missing_check or .no_tenant_filter or .open_using
      )
  );

# Compute a severity score per table (higher = more urgent)
def severity_of(tbl; has_tid):
  (if has_tid | not                                             then 3 else 0 end) +
  (if (rls_for(tbl) | . == null or .rls_enabled == false)      then 3 else 0 end) +
  (if (policies_for(tbl) | length) == 0                        then 2 else 0 end) +
  (if (policy_problems(tbl) | length) > 0                      then 2 else 0 end) +
  (if (rls_for(tbl) | . != null and .rls_forced == false)      then 1 else 0 end) +
  (if has_tid and (has_tenant_index(tbl) | not)                then 1 else 0 end);

# ── Main output ───────────────────────────────────────────────────────────────
{
  generated_at: (now | todate),

  tables: (
    $T | safe_array | map(
      . as $t |
      (.table_name) as $n |
      (.has_tenant_id // false) as $htid |
      {
        schema:        $t.table_schema,
        name:          $n,
        has_tenant_id: $htid,
        columns:       ($t.columns  | safe_array),
        rls:           rls_for($n),
        policies:      policies_for($n),
        indexes:       indexes_for($n),
        checks:        checks_for($n),
        analysis: {
          tenant_column_missing:  ($htid | not),
          tenant_id_not_indexed:  ($htid and (has_tenant_index($n) | not)),
          rls_disabled:           (rls_for($n) | . == null or .rls_enabled == false),
          rls_not_forced:         (rls_for($n) | . != null and .rls_enabled == true and .rls_forced == false),
          no_policies:            (policies_for($n) | length == 0),
          policy_problems:        policy_problems($n),
          has_policy_problems:    (policy_problems($n) | length > 0),
          severity:               severity_of($n; $htid)
        }
      }
    )
  ),

  relationships: ($F | safe_array),

  views: (
    $V | safe_array | map(
      . as $v |
      {
        schema:           $v.schema,
        view:             $v.view,
        security_invoker: ($v.security_invoker // false),
        risk: (
          if ($v.security_invoker // false)
          then "ok"
          else "high — no security_invoker=true; view runs as owner and may bypass RLS"
          end
        )
      }
    )
  ),

  summary: (
    ($T | safe_array) as $tables |
    ($R | safe_array) as $rlss   |
    ($P | safe_array) as $pols   |
    ($I | safe_array) as $idxs   |
    ($V | safe_array) as $vws    |

    ($pols | map(.table) | unique) as $tables_with_policy |

    {
      total_tables: ($tables | length),

      tenant_isolation: {
        missing_tenant_id: (
          $tables | map(select((.has_tenant_id // false) == false)) | length
        ),
        missing_tenant_id_names: (
          $tables | map(select((.has_tenant_id // false) == false)) | map(.table_name)
        ),
        missing_tenant_id_index: (
          $tables | map(select(
            (.has_tenant_id // false) == true and
            (
              (.table_name) as $tn |
              ($idxs | map(select(
                .table == $tn and
                (.columns | safe_array | index("tenant_id")) != null
              )) | length) == 0
            )
          )) | length
        )
      },

      rls: {
        disabled:   ($rlss | map(select(.rls_enabled == false))                              | length),
        not_forced: ($rlss | map(select(.rls_enabled == true and .rls_forced == false))      | length),
        no_policy:  ($tables | map(.table_name) | map(select(. as $t | ($tables_with_policy | index($t)) == null)) | length)
      },

      views: {
        total:                   ($vws | length),
        without_security_invoker: ($vws | map(select((.security_invoker // false) == false)) | length)
      },

      critical_tables: (
        $tables | map(
          (.table_name) as $n |
          (.has_tenant_id // false) as $htid |
          {
            name:     $n,
            severity: severity_of($n; $htid)
          }
        )
        | map(select(.severity >= 5))
        | sort_by(-.severity)
      ),

      overall_health: (
        (
          ($tables | map(select((.has_tenant_id // false) == false)) | length) +
          ($rlss   | map(select(.rls_enabled == false))              | length) +
          ($vws    | map(select((.security_invoker // false) == false)) | length)
        ) as $n |
        if   $n == 0  then "✅  clean"
        elif $n <= 3  then "⚠️   minor issues (\($n))"
        elif $n <= 10 then "  moderate issues (\($n))"
        else               "  critical issues (\($n))"
        end
      )
    }
  )
}
'
