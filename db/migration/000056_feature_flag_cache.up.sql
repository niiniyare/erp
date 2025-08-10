-- Creates materialized view for fast feature flag lookups and cache management

-- =====================================================
-- MATERIALIZED VIEW FOR FEATURE FLAG CACHE
-- =====================================================

-- Materialized view for feature flag evaluation cache
CREATE MATERIALIZED VIEW mv_tenant_feature_flags_cache AS
WITH feature_evaluation AS (
    SELECT 
        ff.tenant_id,
        ff.id as feature_flag_id,
        ff.name as feature_flag_name,
        ff.flag_type,
        ff.default_value,
        ff.rollout_percentage,
        ff.target_audience,
        ff.metadata,
        tfo.enabled as override_enabled,
        tfo.value as override_value,
        tfo.reason as override_reason,
        CASE 
            -- Override exists, use it
            WHEN tfo.enabled IS NOT NULL THEN tfo.enabled
            -- Percentage rollout check
            WHEN ff.rollout_percentage IS NOT NULL AND ff.rollout_percentage > 0 THEN
                (hashtext(ff.tenant_id::text || ff.name) % 100) < ff.rollout_percentage
            -- Default value
            ELSE ff.default_value
        END as effective_enabled,
        COALESCE(tfo.value, '{}') as effective_value,
        CASE 
            WHEN tfo.enabled IS NOT NULL THEN 'override'
            WHEN ff.rollout_percentage IS NOT NULL AND ff.rollout_percentage > 0 THEN 'rollout'
            ELSE 'default'
        END as evaluation_source,
        GREATEST(ff.updated_at, COALESCE(tfo.updated_at, ff.updated_at)) as cache_timestamp
    FROM feature_flags ff
    LEFT JOIN tenant_feature_overrides tfo ON ff.id = tfo.feature_flag_id 
        AND tfo.tenant_id = ff.tenant_id
    WHERE ff.deleted_at IS NULL
)
SELECT 
    tenant_id,
    feature_flag_id,
    feature_flag_name,
    flag_type,
    effective_enabled as enabled,
    effective_value as value,
    evaluation_source,
    default_value,
    rollout_percentage,
    target_audience,
    metadata,
    override_enabled,
    override_value,
    override_reason,
    cache_timestamp,
    NOW() as cache_created_at
FROM feature_evaluation;

-- =====================================================
-- PERFORMANCE INDEXES ON MATERIALIZED VIEW
-- =====================================================

-- Primary lookup indexes
CREATE UNIQUE INDEX idx_tenant_feature_cache_pk 
    ON mv_tenant_feature_flags_cache(tenant_id, feature_flag_id);

CREATE UNIQUE INDEX idx_tenant_feature_cache_name_lookup 
    ON mv_tenant_feature_flags_cache(tenant_id, feature_flag_name);

-- Query optimization indexes
CREATE INDEX idx_tenant_feature_cache_tenant 
    ON mv_tenant_feature_flags_cache(tenant_id);

CREATE INDEX idx_tenant_feature_cache_enabled 
    ON mv_tenant_feature_flags_cache(enabled) WHERE enabled = true;

CREATE INDEX idx_tenant_feature_cache_source 
    ON mv_tenant_feature_flags_cache(evaluation_source);

CREATE INDEX idx_tenant_feature_cache_flag_type 
    ON mv_tenant_feature_flags_cache(flag_type);

CREATE INDEX idx_tenant_feature_cache_timestamp 
    ON mv_tenant_feature_flags_cache(cache_timestamp);

CREATE INDEX idx_tenant_feature_cache_rollout 
    ON mv_tenant_feature_flags_cache(rollout_percentage) 
    WHERE rollout_percentage IS NOT NULL;

-- JSON indexes for complex queries
CREATE INDEX idx_tenant_feature_cache_target_audience 
    ON mv_tenant_feature_flags_cache USING GIN (target_audience) 
    WHERE target_audience != '{}';

CREATE INDEX idx_tenant_feature_cache_metadata 
    ON mv_tenant_feature_flags_cache USING GIN (metadata) 
    WHERE metadata != '{}';

CREATE INDEX idx_tenant_feature_cache_value 
    ON mv_tenant_feature_flags_cache USING GIN (value) 
    WHERE value != '{}';

-- =====================================================
-- CACHE MANAGEMENT FUNCTIONS
-- =====================================================

-- Function to refresh the cache
CREATE OR REPLACE FUNCTION refresh_feature_flags_cache()
RETURNS VOID AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_tenant_feature_flags_cache;
    
    -- Log cache refresh
    INSERT INTO audit_log (
        tenant_id,
        event_type,
        event_category,
        severity,
        user_id,
        decision,
        reason,
        context,
        session_id
    ) VALUES (
        NULL, -- System operation
        'FEATURE_FLAGS_CACHE_REFRESHED',
        'SYSTEM',
        'INFO',
        NULLIF(current_setting('app.current_user_id', true), '')::UUID,
        'ALLOW',
        'Feature flags cache materialized view refreshed',
        jsonb_build_object(
            'operation', 'cache_refresh',
            'timestamp', NOW()
        ),
        NULLIF(current_setting('app.current_session_id', true), '')::UUID
    );
END;
$$ LANGUAGE plpgsql;

-- Function to get cache statistics
CREATE OR REPLACE FUNCTION get_feature_flags_cache_stats()
RETURNS TABLE(
    total_entries BIGINT,
    tenants_count BIGINT,
    flags_per_tenant_avg NUMERIC,
    enabled_flags_count BIGINT,
    override_count BIGINT,
    rollout_count BIGINT,
    default_count BIGINT,
    cache_age INTERVAL
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        COUNT(*) as total_entries,
        COUNT(DISTINCT tffc.tenant_id) as tenants_count,
        ROUND(COUNT(*)::NUMERIC / COUNT(DISTINCT tffc.tenant_id), 2) as flags_per_tenant_avg,
        COUNT(*) FILTER (WHERE tffc.enabled = true) as enabled_flags_count,
        COUNT(*) FILTER (WHERE tffc.evaluation_source = 'override') as override_count,
        COUNT(*) FILTER (WHERE tffc.evaluation_source = 'rollout') as rollout_count,
        COUNT(*) FILTER (WHERE tffc.evaluation_source = 'default') as default_count,
        NOW() - MIN(tffc.cache_created_at) as cache_age
    FROM mv_tenant_feature_flags_cache tffc;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Function to check cache freshness for a tenant
CREATE OR REPLACE FUNCTION check_cache_freshness(p_tenant_id UUID)
RETURNS TABLE(
    is_stale BOOLEAN,
    cache_age INTERVAL,
    last_flag_update TIMESTAMPTZ,
    last_override_update TIMESTAMPTZ
) AS $$
BEGIN
    RETURN QUERY
    WITH cache_info AS (
        SELECT MAX(cache_timestamp) as max_cache_ts
        FROM mv_tenant_feature_flags_cache
        WHERE tenant_id = p_tenant_id
    ),
    source_info AS (
        SELECT 
            MAX(ff.updated_at) as last_flag_update,
            MAX(tfo.updated_at) as last_override_update
        FROM feature_flags ff
        LEFT JOIN tenant_feature_overrides tfo ON ff.id = tfo.feature_flag_id
        WHERE ff.tenant_id = p_tenant_id AND ff.deleted_at IS NULL
    )
    SELECT 
        COALESCE(si.last_flag_update > ci.max_cache_ts OR si.last_override_update > ci.max_cache_ts, true) as is_stale,
        NOW() - ci.max_cache_ts as cache_age,
        si.last_flag_update,
        si.last_override_update
    FROM cache_info ci
    CROSS JOIN source_info si;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- =====================================================
-- CACHED EVALUATION FUNCTIONS
-- =====================================================

-- Fast evaluation using cache
CREATE OR REPLACE FUNCTION evaluate_feature_flag_cached(flag_name VARCHAR)
RETURNS TABLE(enabled BOOLEAN, value JSONB) AS $$
DECLARE
    v_tenant_id UUID;
    v_result RECORD;
    v_cache_stale BOOLEAN;
BEGIN
    -- Get current tenant ID
    v_tenant_id := NULLIF(current_setting('app.current_tenant_id', true), '')::UUID;
    
    IF v_tenant_id IS NULL THEN
        RAISE EXCEPTION 'No tenant context set';
    END IF;
    
    -- Check if cache is stale (optional check)
    SELECT is_stale INTO v_cache_stale
    FROM check_cache_freshness(v_tenant_id)
    LIMIT 1;
    
    -- If cache is stale, optionally refresh (comment out for performance)
    -- IF v_cache_stale THEN
    --     PERFORM refresh_feature_flags_cache();
    -- END IF;
    
    -- Get result from cache
    SELECT tffc.enabled, tffc.value
    INTO v_result
    FROM mv_tenant_feature_flags_cache tffc
    WHERE tffc.tenant_id = v_tenant_id 
      AND tffc.feature_flag_name = flag_name;
    
    IF v_result IS NULL THEN
        RAISE EXCEPTION 'Feature flag not found in cache: % for tenant: %', flag_name, v_tenant_id;
    END IF;
    
    RETURN QUERY SELECT v_result.enabled, v_result.value;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Bulk evaluation using cache
CREATE OR REPLACE FUNCTION evaluate_all_feature_flags_cached()
RETURNS TABLE(flag_name VARCHAR, enabled BOOLEAN, value JSONB, flag_type VARCHAR, source TEXT) AS $$
DECLARE
    v_tenant_id UUID;
BEGIN
    -- Get current tenant ID
    v_tenant_id := NULLIF(current_setting('app.current_tenant_id', true), '')::UUID;
    
    IF v_tenant_id IS NULL THEN
        RAISE EXCEPTION 'No tenant context set';
    END IF;
    
    RETURN QUERY
    SELECT 
        tffc.feature_flag_name::VARCHAR,
        tffc.enabled,
        tffc.value,
        tffc.flag_type::VARCHAR,
        tffc.evaluation_source::TEXT
    FROM mv_tenant_feature_flags_cache tffc
    WHERE tffc.tenant_id = v_tenant_id
    ORDER BY tffc.feature_flag_name;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- =====================================================
-- AUTOMATIC CACHE REFRESH TRIGGERS
-- =====================================================

-- Function to trigger cache refresh on data changes
CREATE OR REPLACE FUNCTION trigger_cache_refresh()
RETURNS TRIGGER AS $$
BEGIN
    -- Async refresh (use pg_notify for external refresh or schedule)
    PERFORM pg_notify('feature_flags_cache_refresh', 
        jsonb_build_object(
            'operation', TG_OP,
            'table', TG_TABLE_NAME,
            'tenant_id', COALESCE(NEW.tenant_id, OLD.tenant_id)
        )::text
    );
    
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Add triggers to automatically notify when refresh is needed
CREATE TRIGGER feature_flags_cache_refresh_trigger
    AFTER INSERT OR UPDATE OR DELETE ON feature_flags
    FOR EACH ROW
    EXECUTE FUNCTION trigger_cache_refresh();

CREATE TRIGGER tenant_overrides_cache_refresh_trigger
    AFTER INSERT OR UPDATE OR DELETE ON tenant_feature_overrides
    FOR EACH ROW
    EXECUTE FUNCTION trigger_cache_refresh();

-- =====================================================
-- PERMISSIONS
-- =====================================================

-- Grant access to cache functions
GRANT EXECUTE ON FUNCTION refresh_feature_flags_cache() TO admin_role;
GRANT EXECUTE ON FUNCTION get_feature_flags_cache_stats() TO admin_role, readonly_role;
GRANT EXECUTE ON FUNCTION check_cache_freshness(UUID) TO application_role, admin_role;
GRANT EXECUTE ON FUNCTION evaluate_feature_flag_cached(VARCHAR) TO application_role;
GRANT EXECUTE ON FUNCTION evaluate_all_feature_flags_cached() TO application_role;

-- Grant access to materialized view
GRANT SELECT ON mv_tenant_feature_flags_cache TO application_role, admin_role, readonly_role;

-- =====================================================
-- COMMENTS
-- =====================================================
COMMENT ON MATERIALIZED VIEW mv_tenant_feature_flags_cache IS 'Materialized view for fast feature flag lookups with pre-computed evaluations';
COMMENT ON FUNCTION refresh_feature_flags_cache() IS 'Refreshes the feature flags cache materialized view';
COMMENT ON FUNCTION get_feature_flags_cache_stats() IS 'Returns statistics about the feature flags cache';
COMMENT ON FUNCTION check_cache_freshness(UUID) IS 'Checks if the cache is stale for a specific tenant';
COMMENT ON FUNCTION evaluate_feature_flag_cached(VARCHAR) IS 'Fast feature flag evaluation using materialized view cache';
COMMENT ON FUNCTION evaluate_all_feature_flags_cached() IS 'Fast bulk feature flag evaluation using materialized view cache';
COMMENT ON FUNCTION trigger_cache_refresh() IS 'Triggers cache refresh notification when data changes';
