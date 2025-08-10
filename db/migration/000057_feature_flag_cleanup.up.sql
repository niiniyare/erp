-- Creates maintenance functions for cleanup and system health

-- =====================================================
-- AUDIT LOG CLEANUP FUNCTIONS
-- =====================================================

-- Function to clean up old feature flag audit logs
CREATE OR REPLACE FUNCTION cleanup_old_feature_flag_audit_logs(retention_days INTEGER DEFAULT 90)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
    cutoff_date TIMESTAMPTZ;
BEGIN
    cutoff_date := NOW() - (retention_days || ' days')::INTERVAL;
    
    DELETE FROM audit_log
    WHERE event_type IN (
        'FEATURE_FLAG_CREATED', 
        'FEATURE_FLAG_UPDATED', 
        'FEATURE_FLAG_DELETED',
        'FEATURE_OVERRIDE_CREATED', 
        'FEATURE_OVERRIDE_UPDATED', 
        'FEATURE_OVERRIDE_DELETED',
        'FEATURE_FLAG_EVALUATED', 
        'FEATURE_FLAGS_BULK_EVALUATED',
        'FEATURE_FLAGS_CACHE_REFRESHED'
    )
    AND created_at < cutoff_date;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    
    -- Log the cleanup operation
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
        'FEATURE_FLAGS_AUDIT_CLEANUP',
        'SYSTEM',
        'INFO',
        NULLIF(current_setting('app.current_user_id', true), '')::UUID,
        'ALLOW',
        'Feature flag audit logs cleaned up',
        jsonb_build_object(
            'deleted_count', deleted_count,
            'retention_days', retention_days,
            'cutoff_date', cutoff_date,
            'operation', 'cleanup'
        ),
        NULLIF(current_setting('app.current_session_id', true), '')::UUID
    );
    
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Function to clean up soft-deleted feature flags
CREATE OR REPLACE FUNCTION cleanup_soft_deleted_feature_flags(retention_days INTEGER DEFAULT 30)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
    cutoff_date TIMESTAMPTZ;
BEGIN
    cutoff_date := NOW() - (retention_days || ' days')::INTERVAL;
    
    -- Hard delete feature flags that have been soft-deleted for the retention period
    DELETE FROM feature_flags
    WHERE deleted_at IS NOT NULL 
      AND deleted_at < cutoff_date;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    
    -- Log the cleanup operation
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
        'FEATURE_FLAGS_HARD_DELETE_CLEANUP',
        'SYSTEM',
        'INFO',
        NULLIF(current_setting('app.current_user_id', true), '')::UUID,
        'ALLOW',
        'Soft-deleted feature flags permanently removed',
        jsonb_build_object(
            'deleted_count', deleted_count,
            'retention_days', retention_days,
            'cutoff_date', cutoff_date,
            'operation', 'hard_delete_cleanup'
        ),
        NULLIF(current_setting('app.current_session_id', true), '')::UUID
    );
    
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- SYSTEM HEALTH AND MONITORING FUNCTIONS
-- =====================================================

-- Function to get feature flag system health metrics
CREATE OR REPLACE FUNCTION get_feature_flags_health_metrics()
RETURNS TABLE(
    metric_name TEXT,
    metric_value NUMERIC,
    metric_unit TEXT,
    metric_status TEXT,
    details JSONB
) AS $$
BEGIN
    RETURN QUERY
    WITH metrics AS (
        -- Total feature flags
        SELECT 
            'total_feature_flags' as name,
            COUNT(*)::NUMERIC as value,
            'count' as unit,
            CASE WHEN COUNT(*) > 0 THEN 'healthy' ELSE 'warning' END as status,
            jsonb_build_object('active_only', COUNT(*) FILTER (WHERE deleted_at IS NULL)) as details
        FROM feature_flags
        
        UNION ALL
        
        -- Total tenant overrides
        SELECT 
            'total_tenant_overrides' as name,
            COUNT(*)::NUMERIC as value,
            'count' as unit,
            'healthy' as status,
            jsonb_build_object('enabled_overrides', COUNT(*) FILTER (WHERE enabled = true)) as details
        FROM tenant_feature_overrides
        
        UNION ALL
        
        -- Average flags per tenant
        SELECT 
            'avg_flags_per_tenant' as name,
            COALESCE(ROUND(COUNT(*)::NUMERIC / NULLIF(COUNT(DISTINCT tenant_id), 0), 2), 0) as value,
            'count' as unit,
            CASE 
                WHEN COUNT(DISTINCT tenant_id) = 0 THEN 'error'
                WHEN COUNT(*)::NUMERIC / COUNT(DISTINCT tenant_id) > 100 THEN 'warning'
                ELSE 'healthy' 
            END as status,
            jsonb_build_object(
                'total_flags', COUNT(*),
                'total_tenants', COUNT(DISTINCT tenant_id)
            ) as details
        FROM feature_flags
        WHERE deleted_at IS NULL
        
        UNION ALL
        
        -- Cache age
        SELECT 
            'cache_age_minutes' as name,
            COALESCE(EXTRACT(EPOCH FROM (NOW() - MIN(cache_created_at)))/60, 0) as value,
            'minutes' as unit,
            CASE 
                WHEN MIN(cache_created_at) IS NULL THEN 'error'
                WHEN EXTRACT(EPOCH FROM (NOW() - MIN(cache_created_at)))/60 > 60 THEN 'warning'
                ELSE 'healthy' 
            END as status,
            jsonb_build_object(
                'cache_entries', COUNT(*),
                'last_refresh', MIN(cache_created_at)
            ) as details
        FROM mv_tenant_feature_flags_cache
        
        UNION ALL
        
        -- Rollout percentage distribution
        SELECT 
            'flags_with_rollout' as name,
            COUNT(*) FILTER (WHERE rollout_percentage IS NOT NULL)::NUMERIC as value,
            'count' as unit,
            'healthy' as status,
            jsonb_build_object(
                'avg_rollout_percentage', ROUND(AVG(rollout_percentage) FILTER (WHERE rollout_percentage IS NOT NULL), 2),
                'max_rollout_percentage', MAX(rollout_percentage),
                'min_rollout_percentage', MIN(rollout_percentage) FILTER (WHERE rollout_percentage IS NOT NULL)
            ) as details
        FROM feature_flags
        WHERE deleted_at IS NULL
    )
    SELECT m.name, m.value, m.unit, m.status, m.details FROM metrics m;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Function to get tenant-specific feature flag statistics
CREATE OR REPLACE FUNCTION get_tenant_feature_flag_stats(p_tenant_id UUID)
RETURNS TABLE(
    total_flags INTEGER,
    enabled_flags INTEGER,
    overridden_flags INTEGER,
    rollout_flags INTEGER,
    flag_types JSONB,
    last_evaluation TIMESTAMPTZ,
    evaluation_count_today INTEGER
) AS $$
BEGIN
    RETURN QUERY
    WITH tenant_stats AS (
        SELECT 
            COUNT(*)::INTEGER as total_flags,
            COUNT(*) FILTER (WHERE tffc.enabled = true)::INTEGER as enabled_flags,
            COUNT(*) FILTER (WHERE tffc.evaluation_source = 'override')::INTEGER as overridden_flags,
            COUNT(*) FILTER (WHERE tffc.evaluation_source = 'rollout')::INTEGER as rollout_flags,
            jsonb_object_agg(tffc.flag_type, COUNT(*)) as flag_types
        FROM mv_tenant_feature_flags_cache tffc
        WHERE tffc.tenant_id = p_tenant_id
    ),
    audit_stats AS (
        SELECT 
            MAX(al.created_at) as last_evaluation,
            COUNT(*) FILTER (WHERE al.created_at >= CURRENT_DATE)::INTEGER as evaluation_count_today
        FROM audit_log al
        WHERE al.tenant_id = p_tenant_id
          AND al.event_type IN ('FEATURE_FLAG_EVALUATED', 'FEATURE_FLAGS_BULK_EVALUATED')
    )
    SELECT 
        ts.total_flags,
        ts.enabled_flags,
        ts.overridden_flags,
        ts.rollout_flags,
        ts.flag_types,
        aus.last_evaluation,
        aus.evaluation_count_today
    FROM tenant_stats ts
    CROSS JOIN audit_stats aus;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- =====================================================
-- DATA INTEGRITY FUNCTIONS
-- =====================================================

-- Function to check data integrity
CREATE OR REPLACE FUNCTION check_feature_flags_integrity()
RETURNS TABLE(
    check_name TEXT,
    status TEXT,
    issue_count INTEGER,
    details JSONB
) AS $$
BEGIN
    RETURN QUERY
    WITH integrity_checks AS (
        -- Check for orphaned overrides
        SELECT 
            'orphaned_overrides' as check_name,
            CASE WHEN COUNT(*) = 0 THEN 'pass' ELSE 'fail' END as status,
            COUNT(*)::INTEGER as issue_count,
            jsonb_agg(
                jsonb_build_object(
                    'override_id', tfo.id,
                    'tenant_id', tfo.tenant_id,
                    'feature_flag_id', tfo.feature_flag_id
                )
            ) as details
        FROM tenant_feature_overrides tfo
        LEFT JOIN feature_flags ff ON tfo.feature_flag_id = ff.id
        WHERE ff.id IS NULL
        
        UNION ALL
        
        -- Check for mismatched tenant IDs
        SELECT 
            'mismatched_tenant_ids' as check_name,
            CASE WHEN COUNT(*) = 0 THEN 'pass' ELSE 'fail' END as status,
            COUNT(*)::INTEGER as issue_count,
            jsonb_agg(
                jsonb_build_object(
                    'override_id', tfo.id,
                    'override_tenant_id', tfo.tenant_id,
                    'flag_tenant_id', ff.tenant_id
                )
            ) as details
        FROM tenant_feature_overrides tfo
        JOIN feature_flags ff ON tfo.feature_flag_id = ff.id
        WHERE tfo.tenant_id != ff.tenant_id
        
        UNION ALL
        
        -- Check for invalid rollout percentages
        SELECT 
            'invalid_rollout_percentages' as check_name,
            CASE WHEN COUNT(*) = 0 THEN 'pass' ELSE 'fail' END as status,
            COUNT(*)::INTEGER as issue_count,
            jsonb_agg(
                jsonb_build_object(
                    'flag_id', ff.id,
                    'flag_name', ff.name,
                    'rollout_percentage', ff.rollout_percentage
                )
            ) as details
        FROM feature_flags ff
        WHERE ff.rollout_percentage IS NOT NULL 
          AND (ff.rollout_percentage < 0 OR ff.rollout_percentage > 100)
        
        UNION ALL
        
        -- Check for duplicate flag names per tenant
        SELECT 
            'duplicate_flag_names' as check_name,
            CASE WHEN COUNT(*) = 0 THEN 'pass' ELSE 'fail' END as status,
            COUNT(*)::INTEGER as issue_count,
            jsonb_agg(
                jsonb_build_object(
                    'tenant_id', tenant_id,
                    'flag_name', name,
                    'count', flag_count
                )
            ) as details
        FROM (
            SELECT tenant_id, name, COUNT(*) as flag_count
            FROM feature_flags
            WHERE deleted_at IS NULL
            GROUP BY tenant_id, name
            HAVING COUNT(*) > 1
        ) duplicates
    )
    SELECT ic.check_name, ic.status, ic.issue_count, ic.details FROM integrity_checks ic;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Function to fix orphaned overrides
CREATE OR REPLACE FUNCTION fix_orphaned_overrides()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    -- Delete orphaned overrides
    DELETE FROM tenant_feature_overrides tfo
    WHERE NOT EXISTS (
        SELECT 1 FROM feature_flags ff 
        WHERE ff.id = tfo.feature_flag_id
    );
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    
    -- Log the fix operation
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
        'FEATURE_FLAGS_ORPHANED_OVERRIDES_FIXED',
        'SYSTEM',
        'WARN',
        NULLIF(current_setting('app.current_user_id', true), '')::UUID,
        'ALLOW',
        'Orphaned feature flag overrides removed',
        jsonb_build_object(
            'deleted_count', deleted_count,
            'operation', 'fix_orphaned_overrides'
        ),
        NULLIF(current_setting('app.current_session_id', true), '')::UUID
    );
    
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- BATCH OPERATIONS
-- =====================================================

-- Function to bulk update rollout percentages
CREATE OR REPLACE FUNCTION bulk_update_rollout_percentage(
    flag_names TEXT[],
    new_percentage INTEGER,
    p_tenant_id UUID DEFAULT NULL
)
RETURNS INTEGER AS $$
DECLARE
    updated_count INTEGER;
    target_tenant_id UUID;
BEGIN
    -- Use provided tenant_id or current context
    target_tenant_id := COALESCE(p_tenant_id, NULLIF(current_setting('app.current_tenant_id', true), '')::UUID);
    
    IF target_tenant_id IS NULL THEN
        RAISE EXCEPTION 'No tenant context provided';
    END IF;
    
    -- Validate percentage
    IF new_percentage < 0 OR new_percentage > 100 THEN
        RAISE EXCEPTION 'Rollout percentage must be between 0 and 100';
    END IF;
    
    -- Update rollout percentages
    UPDATE feature_flags 
    SET 
        rollout_percentage = new_percentage,
        updated_at = NOW()
    WHERE tenant_id = target_tenant_id
      AND name = ANY(flag_names)
      AND deleted_at IS NULL;
    
    GET DIAGNOSTICS updated_count = ROW_COUNT;
    
    -- Log the bulk update
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
        target_tenant_id,
        'FEATURE_FLAGS_BULK_ROLLOUT_UPDATE',
        'ADMIN',
        'INFO',
        NULLIF(current_setting('app.current_user_id', true), '')::UUID,
        'ALLOW',
        'Bulk rollout percentage update',
        jsonb_build_object(
            'flag_names', flag_names,
            'new_percentage', new_percentage,
            'updated_count', updated_count,
            'operation', 'bulk_rollout_update'
        ),
        NULLIF(current_setting('app.current_session_id', true), '')::UUID
    );
    
    RETURN updated_count;
END;
$$ LANGUAGE plpgsql;

-- Function to bulk create feature flags
CREATE OR REPLACE FUNCTION bulk_create_feature_flags(
    flag_definitions JSONB,
    p_tenant_id UUID DEFAULT NULL
)
RETURNS INTEGER AS $$
DECLARE
    created_count INTEGER := 0;
    flag_def JSONB;
    target_tenant_id UUID;
BEGIN
    -- Use provided tenant_id or current context
    target_tenant_id := COALESCE(p_tenant_id, NULLIF(current_setting('app.current_tenant_id', true), '')::UUID);
    
    IF target_tenant_id IS NULL THEN
        RAISE EXCEPTION 'No tenant context provided';
    END IF;
    
    -- Process each flag definition
    FOR flag_def IN SELECT jsonb_array_elements(flag_definitions)
    LOOP
        INSERT INTO feature_flags (
            tenant_id,
            name,
            description,
            flag_type,
            default_value,
            rollout_percentage,
            target_audience,
            metadata
        ) VALUES (
            target_tenant_id,
            flag_def->>'name',
            flag_def->>'description',
            COALESCE(flag_def->>'flag_type', 'boolean'),
            COALESCE((flag_def->>'default_value')::BOOLEAN, false),
            (flag_def->>'rollout_percentage')::INTEGER,
            COALESCE(flag_def->'target_audience', '{}'),
            COALESCE(flag_def->'metadata', '{}')
        )
        ON CONFLICT (tenant_id, name) DO NOTHING;
        
        IF FOUND THEN
            created_count := created_count + 1;
        END IF;
    END LOOP;
    
    -- Log the bulk creation
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
        target_tenant_id,
        'FEATURE_FLAGS_BULK_CREATED',
        'ADMIN',
        'INFO',
        NULLIF(current_setting('app.current_user_id', true), '')::UUID,
        'ALLOW',
        'Bulk feature flags creation',
        jsonb_build_object(
            'definitions', flag_definitions,
            'created_count', created_count,
            'operation', 'bulk_create'
        ),
        NULLIF(current_setting('app.current_session_id', true), '')::UUID
    );
    
    RETURN created_count;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- EXPORT AND IMPORT FUNCTIONS
-- =====================================================

-- Function to export tenant feature flags configuration
CREATE OR REPLACE FUNCTION export_tenant_feature_flags(p_tenant_id UUID)
RETURNS JSONB AS $$
DECLARE
    result JSONB;
BEGIN
    SELECT jsonb_build_object(
        'tenant_id', p_tenant_id,
        'export_timestamp', NOW(),
        'feature_flags', jsonb_agg(
            jsonb_build_object(
                'name', ff.name,
                'description', ff.description,
                'flag_type', ff.flag_type,
                'default_value', ff.default_value,
                'rollout_percentage', ff.rollout_percentage,
                'target_audience', ff.target_audience,
                'metadata', ff.metadata,
                'created_at', ff.created_at,
                'updated_at', ff.updated_at
            )
        ),
        'overrides', (
            SELECT jsonb_agg(
                jsonb_build_object(
                    'feature_flag_name', tfo.feature_flag_name,
                    'enabled', tfo.enabled,
                    'value', tfo.value,
                    'reason', tfo.reason,
                    'created_at', tfo.created_at,
                    'updated_at', tfo.updated_at
                )
            )
            FROM tenant_feature_overrides tfo
            WHERE tfo.tenant_id = p_tenant_id
        )
    ) INTO result
    FROM feature_flags ff
    WHERE ff.tenant_id = p_tenant_id
      AND ff.deleted_at IS NULL;
    
    -- Log the export
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
        p_tenant_id,
        'FEATURE_FLAGS_EXPORTED',
        'ADMIN',
        'INFO',
        NULLIF(current_setting('app.current_user_id', true), '')::UUID,
        'ALLOW',
        'Feature flags configuration exported',
        jsonb_build_object(
            'export_size_bytes', octet_length(result::text),
            'operation', 'export'
        ),
        NULLIF(current_setting('app.current_session_id', true), '')::UUID
    );
    
    RETURN result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- =====================================================
-- PERMISSIONS
-- =====================================================

-- Grant execute permissions to admin role
GRANT EXECUTE ON FUNCTION cleanup_old_feature_flag_audit_logs(INTEGER) TO admin_role;
GRANT EXECUTE ON FUNCTION cleanup_soft_deleted_feature_flags(INTEGER) TO admin_role;
GRANT EXECUTE ON FUNCTION get_feature_flags_health_metrics() TO admin_role, readonly_role;
GRANT EXECUTE ON FUNCTION get_tenant_feature_flag_stats(UUID) TO application_role, admin_role;
GRANT EXECUTE ON FUNCTION check_feature_flags_integrity() TO admin_role;
GRANT EXECUTE ON FUNCTION fix_orphaned_overrides() TO admin_role;
GRANT EXECUTE ON FUNCTION bulk_update_rollout_percentage(TEXT[], INTEGER, UUID) TO admin_role;
GRANT EXECUTE ON FUNCTION bulk_create_feature_flags(JSONB, UUID) TO admin_role;
GRANT EXECUTE ON FUNCTION export_tenant_feature_flags(UUID) TO admin_role;

-- Grant limited permissions to application role
GRANT EXECUTE ON FUNCTION get_tenant_feature_flag_stats(UUID) TO application_role;

-- =====================================================
-- FUNCTION COMMENTS
-- =====================================================
COMMENT ON FUNCTION cleanup_old_feature_flag_audit_logs(INTEGER) IS 'Cleans up feature flag related audit logs older than specified days';
COMMENT ON FUNCTION cleanup_soft_deleted_feature_flags(INTEGER) IS 'Permanently removes feature flags that have been soft-deleted for specified days';
COMMENT ON FUNCTION get_feature_flags_health_metrics() IS 'Returns comprehensive health metrics for the feature flag system';
COMMENT ON FUNCTION get_tenant_feature_flag_stats(UUID) IS 'Returns detailed statistics for a specific tenant''s feature flags';
COMMENT ON FUNCTION check_feature_flags_integrity() IS 'Performs data integrity checks on feature flag tables';
COMMENT ON FUNCTION fix_orphaned_overrides() IS 'Removes orphaned tenant feature overrides that reference non-existent feature flags';
COMMENT ON FUNCTION bulk_update_rollout_percentage(TEXT[], INTEGER, UUID) IS 'Updates rollout percentage for multiple feature flags in bulk';
COMMENT ON FUNCTION bulk_create_feature_flags(JSONB, UUID) IS 'Creates multiple feature flags from JSON definitions';
COMMENT ON FUNCTION export_tenant_feature_flags(UUID) IS 'Exports complete feature flag configuration for a tenant in JSON format';
