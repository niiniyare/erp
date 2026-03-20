-- =====================================================================
-- IAM SPEC — setting_definitions + tenant_settings + user_preferences
-- These queries power SettingService.ResolveForTenant and are called
-- once at login; results stored in sessions.configuration.settings/prefs.
-- =====================================================================

-- name: ResolveAllSettingsForTenant :many
-- THE core query. Called once at login. Returns effective value for every setting.
-- Result stored in sessions.configuration.settings for O(1) per-request access.
SELECT
    sd.setting_key,
    sd.value_type,
    sd.is_system,
    COALESCE(ts.value, sd.default_value) AS effective_value
FROM setting_definitions sd
LEFT JOIN tenant_settings ts
    ON ts.setting_id = sd.id AND ts.tenant_id = $1
ORDER BY sd.setting_key;

-- name: GetSettingDefinition :one
SELECT * FROM setting_definitions
WHERE setting_key = $1;

-- name: ListSettingDefinitionsByModule :many
-- Used by settings screen schema builder — returns all settings for a module
-- with current tenant values so the form can pre-fill.
SELECT
    sd.id,
    sd.setting_key,
    sd.label,
    sd.description,
    sd.value_type,
    sd.default_value,
    sd.enum_options,
    sd.min_value,
    sd.max_value,
    sd.is_system,
    sd.module_id,
    sd.resource_id,
    ts.value AS tenant_value,
    COALESCE(ts.value, sd.default_value) AS effective_value
FROM setting_definitions sd
LEFT JOIN tenant_settings ts
    ON ts.setting_id = sd.id AND ts.tenant_id = $1
WHERE sd.module_id = (SELECT id FROM modules WHERE slug = $2 AND scope = 'SYSTEM')
ORDER BY sd.setting_key;

-- name: UpsertTenantSetting :exec
INSERT INTO tenant_settings (tenant_id, setting_id, setting_key, value, set_by)
SELECT $1, sd.id, sd.setting_key, $2, $3
FROM   setting_definitions sd
WHERE  sd.setting_key = $4
ON CONFLICT (tenant_id, setting_id)
DO UPDATE SET value    = EXCLUDED.value,
             set_by   = EXCLUDED.set_by,
             set_at   = NOW();

-- name: DeleteTenantSetting :exec
-- Removes tenant override — setting reverts to default_value.
DELETE FROM tenant_settings
WHERE tenant_id   = $1
  AND setting_key = $2;

-- name: GetUserPreferences :many
-- Called once at login — result stored in sessions.configuration.prefs.
SELECT pref_key, value
FROM   user_preferences
WHERE  user_id = $1
ORDER  BY pref_key;

-- name: SetUserPreference :exec
INSERT INTO user_preferences (user_id, pref_key, value)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, pref_key)
DO UPDATE SET value  = EXCLUDED.value,
             set_at = NOW();

-- name: DeleteUserPreference :exec
DELETE FROM user_preferences
WHERE user_id  = $1
  AND pref_key = $2;

-- =====================================================================
-- LEGACY CONFIG QUERIES (configuration_templates, config_definitions, etc.)
-- =====================================================================

-- This file contains type-safe SQL queries for the Settings Module

-- ==========================================
-- Configuration Definitions Queries
-- ==========================================

-- name: GetConfigDefinition :one
SELECT * FROM config_definitions 
WHERE module_name = sqlc.arg(module_name) AND config_key = sqlc.arg(config_key);

-- name: ListConfigDefinitions :many
SELECT * FROM config_definitions 
WHERE (sqlc.narg(module_name)::TEXT IS NULL OR sqlc.narg(module_name)::TEXT = '' OR module_name = sqlc.narg(module_name))
ORDER BY module_name, config_key;

-- name: CreateConfigDefinition :one
INSERT INTO config_definitions (
    module_name, config_key, config_type, default_value, 
    validation_rules, description, 
  -- required_permission, 
  -- is_overridable
    required_feature_flag 
  ) VALUES (
    sqlc.arg(module_name), sqlc.arg(config_key), sqlc.arg(config_type), sqlc.arg(default_value), 
    sqlc.arg(validation_rules), sqlc.arg(description),
  -- sqlc.arg(required_permission), 
    sqlc.arg(required_feature_flag)
  -- sqlc.arg(is_overridable)
) RETURNING *;

-- name: UpdateConfigDefinition :one
UPDATE config_definitions 
SET 
    config_type = COALESCE(sqlc.narg(config_type), config_type),
    default_value = COALESCE(sqlc.narg(default_value), default_value),
    validation_rules = COALESCE(sqlc.narg(validation_rules), validation_rules),
    description = COALESCE(sqlc.narg(description), description),
    -- required_permission = COALESCE(sqlc.narg(required_permission), required_permission),
    required_feature_flag = COALESCE(sqlc.narg(required_feature_flag), required_feature_flag),
    -- is_overridable = COALESCE(sqlc.narg(is_overridable), is_overridable),
    updated_at = NOW()
WHERE module_name = sqlc.arg(module_name) AND config_key = sqlc.arg(config_key)
RETURNING *;

-- ==========================================
-- Configuration Resolution Queries
-- ==========================================

-- name: GetTenantConfigurations :one
SELECT settings, settings_version FROM tenant_configurations 
WHERE tenant_id = current_tenant_id();

-- name: UpdateTenantConfigurationSettings :exec
UPDATE tenant_configurations 
SET 
    settings = jsonb_set(
        COALESCE(settings, '{}'),
        sqlc.arg(config_path)::text[],
        sqlc.arg(config_value)::jsonb
    ),
    updated_at = NOW(),
    settings_version = settings_version + 1
WHERE tenant_id = current_tenant_id();

-- name: DeleteTenantConfigurationSettings :exec
UPDATE tenant_configurations 
SET 
    settings = settings #- sqlc.arg(config_path)::text[],
    updated_at = NOW(),
    settings_version = settings_version + 1
WHERE tenant_id = current_tenant_id();

-- name: GetEntityConfigurations :one
SELECT settings FROM entities 
WHERE tenant_id = current_tenant_id() AND uuid = sqlc.arg(entity_id) AND deleted_at IS NULL;

-- name: UpdateEntityConfiguration :exec
UPDATE entities 
SET 
    settings = jsonb_set(
        COALESCE(settings, '{}'),
        sqlc.arg(config_path)::text[],
        sqlc.arg(config_value)::jsonb
    ),
    updated_at = NOW()
WHERE tenant_id = current_tenant_id() AND uuid = sqlc.arg(entity_id);

-- name: DeleteEntityConfiguration :exec
UPDATE entities 
SET 
    settings = settings #- sqlc.arg(config_path)::text[],
    updated_at = NOW()
WHERE tenant_id = current_tenant_id() AND uuid = sqlc.arg(entity_id);

-- name: GetEffectiveConfiguration :one
WITH RECURSIVE config_resolution AS (
    -- System default from config_definitions
    SELECT 
        cd.module_name,
        cd.config_key,
        cd.default_value as value,
        'system' as source,
        0 as priority,
        cd.config_type
        -- cd.is_overridable
    FROM config_definitions cd
    WHERE cd.module_name = sqlc.arg(module_name) AND cd.config_key = sqlc.arg(config_key)
    
    UNION ALL
    
    -- Tenant-level configuration
    SELECT 
        sqlc.arg(module_name)::TEXT as module_name,
        sqlc.arg(config_key)::TEXT as config_key,
        (tc.settings -> (sqlc.arg(module_name) || '.' || sqlc.arg(config_key))) as value,
        'tenant' as source,
        1 as priority,
        cd.config_type
        -- cd.is_overridable
    FROM tenant_configurations tc
    JOIN config_definitions cd ON cd.module_name = sqlc.arg(module_name) AND cd.config_key = sqlc.arg(config_key)
    WHERE tc.tenant_id = current_tenant_id()
    AND tc.settings ? (sqlc.arg(module_name) || '.' || sqlc.arg(config_key))
    
    UNION ALL
    
    -- Entity-level configuration
    SELECT 
        sqlc.arg(module_name)::TEXT as module_name,
        sqlc.arg(config_key)::TEXT as config_key,
        (e.settings -> (sqlc.arg(module_name) || '.' || sqlc.arg(config_key))) as value,
        'entity' as source,
        2 as priority,
        cd.config_type
        -- cd.is_overridable
    FROM entities e
    JOIN config_definitions cd ON cd.module_name = sqlc.arg(module_name) AND cd.config_key = sqlc.arg(config_key)
    WHERE e.tenant_id = current_tenant_id()
    AND e.uuid = sqlc.arg(entity_id)
    AND e.deleted_at IS NULL
    AND e.settings ? (sqlc.arg(module_name) || '.' || sqlc.arg(config_key))
)
SELECT 
    module_name,
    config_key,
    value,
    source,
    config_type
    -- is_overridable
FROM config_resolution
ORDER BY priority DESC
LIMIT 1;

-- name: ListTenantEffectiveConfigurations :many
WITH tenant_configs AS (
    SELECT 
        kv.key as config_full_key,
        split_part(kv.key, '.', 1) as module_name,
        split_part(kv.key, '.', 2) as config_key,
        kv.value,
        'tenant' as source
    FROM tenant_configurations tc,
         jsonb_each(tc.settings) as kv(key, value)
    WHERE tc.tenant_id = current_tenant_id()
),
entity_configs AS (
    SELECT 
        kv.key as config_full_key,
        split_part(kv.key, '.', 1) as module_name,
        split_part(kv.key, '.', 2) as config_key,
        kv.value,
        'entity' as source
    FROM entities e,
         jsonb_each(e.settings) as kv(key, value)
    WHERE e.tenant_id = current_tenant_id()
    AND e.uuid = sqlc.arg(entity_id)
    AND e.deleted_at IS NULL
),
system_configs AS (
    SELECT 
        cd.module_name || '.' || cd.config_key as config_full_key,
        cd.module_name,
        cd.config_key,
        cd.default_value as value,
        'system' as source
    FROM config_definitions cd
    WHERE (sqlc.narg(module_filter)::TEXT IS NULL OR sqlc.narg(module_filter)::TEXT = '' OR cd.module_name = sqlc.narg(module_filter))
)
SELECT DISTINCT ON (config_full_key)
    config_full_key,
    module_name,
    config_key,
    value,
    source
FROM (
    SELECT *, 2 as priority FROM entity_configs
    UNION ALL
    SELECT *, 1 as priority FROM tenant_configs
    UNION ALL
    SELECT *, 0 as priority FROM system_configs
) combined
ORDER BY config_full_key, priority DESC;

-- ==========================================
-- Configuration Templates Queries
-- ==========================================

-- name: CreateConfigurationTemplate :one
INSERT INTO configuration_templates (
    tenant_id, scope, name, category, description, version, configurations,
    applicable_tenant_types, required_feature_flags,
    conflict_resolution, created_by
) VALUES (
    sqlc.narg(tenant_id), sqlc.arg(scope),
    sqlc.arg(name), sqlc.arg(category), sqlc.arg(description), sqlc.arg(version), sqlc.arg(configurations),
    sqlc.arg(applicable_tenant_types), sqlc.arg(required_feature_flags),
    sqlc.arg(conflict_resolution), sqlc.arg(created_by)
) RETURNING *;

-- name: GetConfigurationTemplate :one
SELECT * FROM configuration_templates
WHERE id = sqlc.arg(template_id) AND is_active = true;

-- name: ListConfigurationTemplates :many
-- Returns SYSTEM templates + current tenant's TENANT templates (RLS enforces the TENANT filter).
SELECT * FROM configuration_templates
WHERE is_active = true
AND (sqlc.narg(scope_filter)::TEXT IS NULL    OR sqlc.narg(scope_filter)::TEXT    = '' OR scope    = sqlc.narg(scope_filter))
AND (sqlc.narg(category_filter)::TEXT IS NULL OR sqlc.narg(category_filter)::TEXT = '' OR category = sqlc.narg(category_filter))
AND (sqlc.narg(tenant_types_filter)::TEXT[] IS NULL OR applicable_tenant_types && sqlc.narg(tenant_types_filter)::TEXT[])
ORDER BY scope DESC, name;  -- TENANT before SYSTEM so tenant customisations appear first

-- name: UpdateConfigurationTemplate :one
UPDATE configuration_templates 
SET 
    name = COALESCE(sqlc.narg(name), name),
    description = COALESCE(sqlc.narg(description), description),
    configurations = COALESCE(sqlc.narg(configurations), configurations),
    applicable_tenant_types = COALESCE(sqlc.narg(applicable_tenant_types), applicable_tenant_types),
    required_feature_flags = COALESCE(sqlc.narg(required_feature_flags), required_feature_flags),
    conflict_resolution = COALESCE(sqlc.narg(conflict_resolution), conflict_resolution),
    updated_at = NOW()
WHERE id = sqlc.arg(template_id) AND is_active = true
RETURNING *;

-- name: DeactivateConfigurationTemplate :exec
UPDATE configuration_templates 
SET is_active = false, updated_at = NOW()
WHERE id = sqlc.arg(template_id);

-- ==========================================
-- Template Application Queries
-- ==========================================

-- name: CreateTemplateApplication :one
INSERT INTO template_applications (
    template_id, tenant_id, entity_id, target_type,
    applied_configs, skipped_configs, conflict_count,
    application_summary, applied_by, correlation_id
) VALUES (
    sqlc.arg(template_id), current_tenant_id(), sqlc.narg(entity_id), sqlc.arg(target_type),
    sqlc.arg(applied_configs), sqlc.arg(skipped_configs), sqlc.arg(conflict_count),
    sqlc.arg(application_summary), sqlc.arg(applied_by), sqlc.arg(correlation_id)
) RETURNING *;

-- name: GetTemplateApplicationHistory :many
SELECT 
    ta.*,
    ct.name as template_name,
    ct.version as template_version
FROM template_applications ta
JOIN configuration_templates ct ON ct.id = ta.template_id
WHERE ta.tenant_id = current_tenant_id()
AND (sqlc.narg(entity_id_filter)::UUID IS NULL OR ta.entity_id = sqlc.narg(entity_id_filter))
ORDER BY ta.applied_at DESC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: UpdateTenantTemplateInfo :exec
UPDATE tenant_configurations
SET 
    last_template_applied = sqlc.arg(template_id),
    template_applied_at = NOW()
WHERE tenant_id = current_tenant_id();

-- ==========================================
-- Configuration Audit Queries
-- ==========================================

-- name: CreateConfigurationAudit :one
INSERT INTO configuration_audit (
    tenant_id, entity_id, module_name, config_key_name, old_value, new_value,
    source, operation, user_id, session_id, correlation_id
) VALUES (
    current_tenant_id(), sqlc.narg(entity_id),
    sqlc.arg(module_name), sqlc.arg(config_key_name),
    sqlc.narg(old_value), sqlc.arg(new_value),
    sqlc.arg(source), sqlc.arg(operation), sqlc.arg(user_id),
    sqlc.narg(session_id), sqlc.narg(correlation_id)
) RETURNING *;

-- name: GetConfigurationHistory :many
SELECT * FROM configuration_audit
WHERE tenant_id = current_tenant_id()
AND (sqlc.narg(entity_id_filter)::UUID IS NULL OR entity_id = sqlc.narg(entity_id_filter))
AND (sqlc.narg(module_filter)::TEXT IS NULL OR sqlc.narg(module_filter)::TEXT = '' OR module_name = sqlc.narg(module_filter))
AND (sqlc.narg(config_key_filter)::TEXT IS NULL OR sqlc.narg(config_key_filter)::TEXT = '' OR config_key_name = sqlc.narg(config_key_filter))
ORDER BY applied_at DESC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: GetConfigurationAuditByCorrelation :many
SELECT * FROM configuration_audit
WHERE tenant_id = current_tenant_id()
AND correlation_id = sqlc.arg(correlation_id)
ORDER BY applied_at DESC;

-- ==========================================
-- Bulk Operations Queries
-- ==========================================

-- name: BulkUpdateEntityConfiguration :exec
INSERT INTO entities (uuid, tenant_id, settings, updated_at)
VALUES (sqlc.arg(entity_id), current_tenant_id(), sqlc.arg(settings), NOW())
ON CONFLICT (uuid, tenant_id) 
DO UPDATE SET 
    settings = entities.settings || EXCLUDED.settings,
    updated_at = NOW();

-- name: GetEntitiesForBulkUpdate :many
SELECT uuid, name, type, settings
FROM entities 
WHERE tenant_id = current_tenant_id()
AND deleted_at IS NULL
AND (sqlc.narg(entity_types_filter)::TEXT[] IS NULL OR type = ANY(sqlc.narg(entity_types_filter)::TEXT[]))
AND (sqlc.narg(tags_filter)::TEXT[] IS NULL OR EXISTS (
    SELECT 1 FROM jsonb_array_elements_text(tags) AS tag 
    WHERE tag = ANY(sqlc.narg(tags_filter)::TEXT[])
))
AND (sqlc.narg(created_after)::TIMESTAMPTZ IS NULL OR created_at >= sqlc.narg(created_after));

-- ==========================================
-- Configuration Search Queries
-- ==========================================

-- name: SearchConfigurations :many
WITH all_configs AS (
    -- System configurations
    SELECT 
        cd.module_name,
        cd.config_key,
        cd.module_name || '.' || cd.config_key as full_key,
        cd.default_value as value,
        'system' as source,
        NULL::UUID as tenant_id,
        NULL::UUID as entity_id,
        cd.config_type,
        cd.updated_at
    FROM config_definitions cd
    
    UNION ALL
    
    -- Tenant configurations
    SELECT 
        split_part(key, '.', 1) as module_name,
        split_part(key, '.', 2) as config_key,
        key as full_key,
        value,
        'tenant' as source,
        tc.tenant_id,
        NULL::UUID as entity_id,
        cd.config_type,
        tc.updated_at
    FROM tenant_configurations tc,
         jsonb_each(tc.settings) as setting(key, value)
    JOIN config_definitions cd ON cd.module_name = split_part(key, '.', 1) 
                              AND cd.config_key = split_part(key, '.', 2)
    WHERE tc.tenant_id = current_tenant_id()
    
    UNION ALL
    
    -- Entity configurations
    SELECT 
        split_part(key, '.', 1) as module_name,
        split_part(key, '.', 2) as config_key,
        key as full_key,
        value,
        'entity' as source,
        e.tenant_id,
        e.uuid as entity_id,
        cd.config_type,
        e.updated_at
    FROM entities e,
         jsonb_each(e.settings) as setting(key, value)
    JOIN config_definitions cd ON cd.module_name = split_part(key, '.', 1) 
                              AND cd.config_key = split_part(key, '.', 2)
    WHERE e.tenant_id = current_tenant_id()
    AND e.deleted_at IS NULL
    AND (sqlc.narg(entity_id_filter)::UUID IS NULL OR e.uuid = sqlc.narg(entity_id_filter))
)
SELECT *
FROM all_configs
WHERE (sqlc.narg(modules_filter)::TEXT[] IS NULL OR module_name = ANY(sqlc.narg(modules_filter)::TEXT[]))
AND (sqlc.narg(search_term)::TEXT IS NULL OR sqlc.narg(search_term)::TEXT = '' OR full_key ILIKE '%' || sqlc.narg(search_term) || '%')
AND (sqlc.narg(sources_filter)::TEXT[] IS NULL OR source = ANY(sqlc.narg(sources_filter)::TEXT[]))
AND (sqlc.narg(updated_after)::TIMESTAMPTZ IS NULL OR updated_at >= sqlc.narg(updated_after))
ORDER BY 
    CASE WHEN sqlc.narg(sort_by) = 'module' THEN module_name END,
    CASE WHEN sqlc.narg(sort_by) = 'updated_at' THEN updated_at END DESC,
    full_key
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);
