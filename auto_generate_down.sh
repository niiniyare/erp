#!/usr/bin/bash
set -euo pipefail

MIGRATION_DIR="db/migration"
mkdir -p "$MIGRATION_DIR"

# ------------------------
# Helper: rename existing migrations
# ------------------------
rename_file() {
  local old=$1
  local new=$2
  for suffix in up down; do
    old_file="$MIGRATION_DIR/${old}.${suffix}.sql"
    new_file="$MIGRATION_DIR/${new}.${suffix}.sql"
    if [[ -f "$old_file" ]]; then
      mv "$old_file" "$new_file"
      echo "✔ $old_file → $new_file"
    fi
  done
}

# ------------------------
# Helper: create empty migrations
# ------------------------
create_empty_migrations() {
  local group=$1
  local start=$2
  local end=$((start + 99))

  for i in $(seq $start $end); do
    num=$(printf "%06d" "$i")
    up_file="$MIGRATION_DIR/${num}_${group}.up.sql"
    down_file="$MIGRATION_DIR/${num}_${group}.down.sql"
    [[ -f "$up_file" ]] || touch "$up_file"
    [[ -f "$down_file" ]] || touch "$down_file"
  done
  echo "✔ Reserved migration range $start-$end for group '$group'"
}

# ------------------------
# Groups and start numbers
# ------------------------
GROUPS=(
  "platform 1"
  "tenant 101"
  "entity 201"
  "identity 301"
  "auth 401"
  "user_mgmt 501"
  "settings 601"
  "policy 701"
  "feature_flag 801"
  "finance 901"
  "audit 1001"
)

# ------------------------
# Platform migrations
# ------------------------
echo "🔹 Adding Platform Tables (tenant-agnostic)"
platform_migrations=(
  "000001_platform_countries"
  "000002_platform_states"
  "000003_platform_cities"
  "000004_platform_locations"
  "000005_platform_key_values"
  "000006_platform_modules"
  "000007_platform_resources"
  "000008_platform_actions"
  "000009_platform_audit_log"
)
for migration in "${platform_migrations[@]}"; do
  touch "$MIGRATION_DIR/${migration}.up.sql"
  touch "$MIGRATION_DIR/${migration}.down.sql"
done

# ------------------------
# Renaming existing migrations
# ------------------------
echo "🔹 Renaming Existing Migrations"

# TENANT
rename_file "000001_tenants_core" "000101_tenant_create_core_tables"
rename_file "000002_tenant_configurations" "000102_tenant_add_configuration_tables"
rename_file "000003_tenant_usage_stats" "000103_tenant_add_usage_tracking"
rename_file "000004_provision_tenant_complete" "000104_tenant_provision_complete"
rename_file "000005_tenant_bulk_operations_tracking" "000105_tenant_add_bulk_operations"
rename_file "000027_validate_and_set_tenant_context" "000106_validate_set_tenant_context"
rename_file "000032_enforce_tenant_isolation" "000107_enforce_tenant_isolation"

# ENTITY
rename_file "000006_entities_core" "000201_entity_create_core_tables"
rename_file "000007_entities_hierarchy" "000202_entity_add_hierarchy"
rename_file "000008_entities_state" "000203_entity_add_state_machine"
rename_file "000009_entities_views" "000204_entity_add_views"

# IDENTITY
rename_file "000010_persons" "000301_identity_create_persons"
rename_file "000011_employees" "000302_identity_create_employees"
rename_file "000012_users" "000303_identity_create_users"
rename_file "000013_user_sessions" "000304_identity_add_user_sessions"

# AUTH / RBAC
rename_file "000014_modules" "000401_auth_define_modules"
rename_file "000015_resources" "000402_auth_define_resources"
rename_file "000016_actions" "000403_auth_define_actions"
rename_file "000017_permissions" "000404_auth_define_permissions"
rename_file "000018_roles" "000405_auth_create_roles"
rename_file "000019_role_permissions" "000406_auth_map_role_permissions"
rename_file "000020_user_roles" "000407_auth_assign_user_roles"
rename_file "000021_policies" "000408_policy_define_rules"
rename_file "000022_user_entity_access" "000409_user_entity_access"
rename_file "000023_access_requests" "000410_access_request_flow"
rename_file "000043_user_permission" "000411_user_permission"
rename_file "000050_user_roles_functions" "000412_user_roles_functions"

# USER MANAGEMENT
rename_file "000025_user_management_views" "000501_user_management_views"
rename_file "000026_user_functions_triggers" "000502_user_functions_triggers"
rename_file "000044_user_add_on" "000503_user_add_on"
rename_file "000051_user_activities" "000504_user_activities"

# SETTINGS
rename_file "000028_settings_core_tables" "000601_settings_core_tables"
rename_file "000042_notification" "000602_notification"

# POLICY / ATTRIBUTE
rename_file "000045_attribute_definitions" "000701_attribute_definitions"
rename_file "000046_attribute_values" "000702_attribute_values"
rename_file "000047_policy_evaluation" "000703_policy_evaluation"
rename_file "000048_update_policy_evaluations" "000704_policy_evaluation_update"
rename_file "000049_policy_decision" "000705_policy_decision"

# FEATURE FLAGS
rename_file "000052_feature_flag" "000801_feature_flag_core"
rename_file "000053_feature_flag_override" "000802_feature_flag_overrides"
rename_file "000054_feature_flag_audit" "000803_feature_flag_audit"
rename_file "000055_feature_flag_funcs" "000804_feature_flag_functions"
rename_file "000056_feature_flag_cache" "000805_feature_flag_cache"
rename_file "000057_feature_flag_cleanup" "000806_feature_flag_cleanup"
rename_file "000058_feature_flag_usage" "000807_feature_flag_usage"

# FINANCE
rename_file "000059_finance_account_groups" "000901_finance_account_groups"
rename_file "000060_finance_chart_of_accounts" "000902_finance_chart_of_accounts"
rename_file "000061_finance_chart_of_accounts_indexes" "000903_finance_coa_indexes"
rename_file "000062_finance_transactions" "000904_finance_transactions"
rename_file "000063_finance_transaction_entries" "000905_finance_transaction_entries"
rename_file "000064_finance_mod" "000906_finance_mod"
rename_file "000065_finance_add_supporting_tables" "000907_finance_supporting_tables"
rename_file "000066_finance_add_constraints_and_indexes" "000908_finance_constraints_indexes"
rename_file "000067_finance_add_functions_and_triggers" "000909_finance_triggers"
rename_file "000068_finance_add_views" "000910_finance_views"

# AUDIT FUNCTIONS
rename_file "000070_audit_funcs" "001001_platform_audit_funcs"

# ------------------------
# Create empty placeholders for remaining numbers
# ------------------------
echo "🔹 Creating empty placeholders for remaining numbers"
for entry in "${GROUPS[@]}"; do
  group=$(echo $entry | awk '{print $1}')
  start=$(echo $entry | awk '{print $2}')
  # create_empty_migrations "$group" "$start"
done

echo "🎉 All migrations renamed and empty migration placeholders created!"
