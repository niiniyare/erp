-- Provides examples and documentation for using the feature flag system

-- =====================================================
-- EXAMPLE: BASIC FEATURE FLAG SETUP
-- =====================================================

-- Example: Set up context and create feature flags for a tenant
-- Replace 'your-tenant-uuid' with actual tenant ID

/*
-- Set tenant context
SET app.current_tenant_id = 'your-tenant-uuid';
SET app.current_user_id = 'your-user-uuid';
SET app.current_session_id = 'your-session-uuid';

-- Create basic feature flags
INSERT INTO feature_flags (tenant_id, name, description, flag_type, default_value, metadata) VALUES
    ('your-tenant-uuid', 'advanced_reporting', 'Enable advanced reporting features', 'boolean', false, '{"category": "reporting", "priority": "high"}'),
    ('your-tenant-uuid', 'new_dashboard', 'Enable new dashboard UI', 'boolean', false, '{"category": "ui", "priority": "medium"}'),
    ('your-tenant-uuid', 'api_rate_limit', 'API rate limit configuration', 'number', true, '{"category": "performance", "default_limit": 1000}'),
    ('your-tenant-uuid', 'feature_rollout_test', 'Test gradual rollout', 'boolean', false, '{"category": "testing"}');

-- Set rollout percentage for gradual rollout
UPDATE feature_flags 
SET rollout_percentage = 25 
WHERE name = 'feature_rollout_test' 
  AND tenant_id = 'your-tenant-uuid';

-- Create tenant-specific overrides
INSERT INTO tenant_feature_overrides (tenant_id, feature_flag_id, feature_flag_name, enabled, reason) 
SELECT 
    'your-tenant-uuid',
    ff.id,
    ff.name,
    true,
    'Enable advanced reporting for premium tenant'
FROM feature_flags ff 
WHERE ff.name = 'advanced_reporting' 
  AND ff.tenant_id = 'your-tenant-uuid';
*/

-- =====================================================
-- EXAMPLE: FEATURE FLAG EVALUATION
-- =====================================================

-- Example usage queries (run after setting up feature flags above)

/*
-- Evaluate a single feature flag
SELECT * FROM evaluate_feature_flag('advanced_reporting');

-- Fast evaluation without audit logging
SELECT * FROM evaluate_feature_flag_fast('new_dashboard');

-- Evaluate all feature flags for current tenant
SELECT * FROM evaluate_all_feature_flags();

-- Use cached evaluation (faster for frequent calls)
SELECT * FROM evaluate_feature_flag_cached('advanced_reporting');
SELECT * FROM evaluate_all_feature_flags_cached();
*/

-- =====================================================
-- EXAMPLE: MONITORING AND MAINTENANCE
-- =====================================================

-- Example monitoring queries

/*
-- Check system health
SELECT * FROM get_feature_flags_health_metrics();

-- Get tenant-specific statistics
SELECT * FROM get_tenant_feature_flag_stats('your-tenant-uuid');

-- Check cache statistics
SELECT * FROM get_feature_flags_cache_stats();

-- Check cache freshness for a tenant
SELECT * FROM check_cache_freshness('your-tenant-uuid');

-- Check data integrity
SELECT * FROM check_feature_flags_integrity();
*/

-- =====================================================
-- EXAMPLE: BULK OPERATIONS
-- =====================================================

-- Example bulk operations

/*
-- Bulk update rollout percentages
SELECT bulk_update_rollout_percentage(
    ARRAY['new_dashboard', 'feature_rollout_test'], 
    50, 
    'your-tenant-uuid'
);

-- Bulk create feature flags
SELECT bulk_create_feature_flags('[
    {
        "name": "experimental_feature_a",
        "description": "Experimental feature A",
        "flag_type": "boolean",
        "default_value": false,
        "rollout_percentage": 10,
        "metadata": {"category": "experimental"}
    },
    {
        "name": "experimental_feature_b",
        "description": "Experimental feature B",
        "flag_type": "boolean",  
        "default_value": false,
        "metadata": {"category": "experimental"}
    }
]'::jsonb, 'your-tenant-uuid');

-- Export tenant configuration
SELECT export_tenant_feature_flags('your-tenant-uuid');
*/

-- =====================================================
-- EXAMPLE: MAINTENANCE OPERATIONS
-- =====================================================

-- Example maintenance operations (admin only)

/*
-- Refresh the materialized view cache
SELECT refresh_feature_flags_cache();

-- Clean up old audit logs (keep last 90 days)
SELECT cleanup_old_feature_flag_audit_logs(90);

-- Clean up soft-deleted flags (after 30 days)
SELECT cleanup_soft_deleted_feature_flags(30);

-- Fix any orphaned overrides
SELECT fix_orphaned_overrides();
*/

-- =====================================================
-- EXAMPLE: ADVANCED QUERIES
-- =====================================================

-- Example advanced analytics queries

/*
-- Feature flags by type and status
SELECT 
    flag_type,
    evaluation_source,
    COUNT(*) as count,
    ROUND(AVG(CASE WHEN enabled THEN 1 ELSE 0 END) * 100, 2) as enabled_percentage
FROM mv_tenant_feature_flags_cache
WHERE tenant_id = 'your-tenant-uuid'
GROUP BY flag_type, evaluation_source
ORDER BY flag_type, evaluation_source;

-- Most evaluated features (from audit logs)
SELECT 
    context->>'feature_flag_name' as feature_name,
    COUNT(*) as evaluation_count,
    COUNT(*) FILTER (WHERE decision = 'ALLOW') as enabled_count,
    ROUND(COUNT(*) FILTER (WHERE decision = 'ALLOW')::NUMERIC / COUNT(*) * 100, 2) as enabled_percentage
FROM audit_log
WHERE tenant_id = 'your-tenant-uuid'
  AND event_type = 'FEATURE_FLAG_EVALUATED'
  AND created_at >= NOW() - INTERVAL '7 days'
GROUP BY context->>'feature_flag_name'
ORDER BY evaluation_count DESC
LIMIT 10;

-- Rollout effectiveness analysis
SELECT 
    ff.name,
    ff.rollout_percentage,
    COUNT(DISTINCT al.session_id) as unique_evaluations,
    COUNT(*) FILTER (WHERE al.decision = 'ALLOW') as enabled_evaluations,
    ROUND(COUNT(*) FILTER (WHERE al.decision = 'ALLOW')::NUMERIC / COUNT(*) * 100, 2) as actual_enabled_percentage
FROM feature_flags ff
JOIN audit_log al ON al.context->>'feature_flag_name' = ff.name
WHERE ff.tenant_id = 'your-tenant-uuid'
  AND ff.rollout_percentage IS NOT NULL
  AND al.event_type = 'FEATURE_FLAG_EVALUATED'
  AND al.created_at >= NOW() - INTERVAL '24 hours'
GROUP BY ff.name, ff.rollout_percentage
ORDER BY ff.rollout_percentage DESC;
*/

-- =====================================================
-- PERFORMANCE MONITORING QUERIES
-- =====================================================

-- Queries to monitor performance

/*
-- Index usage statistics
SELECT 
    schemaname,
    tablename,
    indexname,
    idx_scan as index_scans,
    idx_tup_read as tuples_read,
    idx_tup_fetch as tuples_fetched
FROM pg_stat_user_indexes 
WHERE tablename IN ('feature_flags', 'tenant_feature_overrides', 'mv_tenant_feature_flags_cache')
ORDER BY idx_scan DESC;

-- Table size statistics
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size,
    n_tup_ins as inserts,
    n_tup_upd as updates,
    n_tup_del as deletes,
    n_live_tup as live_tuples,
    n_dead_tup as dead_tuples
FROM pg_stat_user_tables 
WHERE tablename IN ('feature_flags', 'tenant_feature_overrides', 'audit_log')
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- Slow query analysis for feature flag operations
SELECT 
    query,
    calls,
    total_time,
    mean_time,
    rows,
    100.0 * shared_blks_hit / nullif(shared_blks_hit + shared_blks_read, 0) AS hit_percent
FROM pg_stat_statements 
WHERE query ILIKE '%feature_flag%' 
   OR query ILIKE '%tenant_feature_overrides%'
ORDER BY mean_time DESC
LIMIT 10;
*/

-- =====================================================
-- TROUBLESHOOTING GUIDE
-- =====================================================

/*
TROUBLESHOOTING COMMON ISSUES:

1. "No tenant context set" error:
   - Ensure you set the tenant context before calling functions
   - SET app.current_tenant_id = 'your-tenant-uuid';

2. Feature flag not found:
   - Check if the flag exists and is not soft-deleted
   - Verify tenant_id matches your context
   - SELECT * FROM feature_flags WHERE name = 'flag_name' AND deleted_at IS NULL;

3. Cache is stale:
   - Refresh the materialized view cache
   - SELECT refresh_feature_flags_cache();
   - Check for notifications: LISTEN feature_flags_cache_refresh;

4. Performance issues:
   - Use cached evaluation functions for high-frequency calls
   - Monitor index usage and table statistics
   - Consider partitioning audit_log table for large datasets

5. Data integrity issues:
   - Run integrity checks regularly
   - SELECT * FROM check_feature_flags_integrity();
   - Fix issues: SELECT fix_orphaned_overrides();

6. Audit log growing too large:
   - Regular cleanup of old logs
   - SELECT cleanup_old_feature_flag_audit_logs(90);
   - Consider partitioning by date

7. RLS (Row Level Security) issues:
   - Verify policies are correctly applied
   - Check role permissions
   - Ensure tenant context is properly set
*/

-- =====================================================
-- BEST PRACTICES
-- =====================================================

/*
FEATURE FLAG BEST PRACTICES:

1. Naming Conventions:
   - Use descriptive, hierarchical names: 'ui.new_dashboard', 'api.v2_endpoints'
   - Avoid spaces, use underscores or dots
   - Include the feature area as prefix

2. Rollout Strategy:
   - Start with low percentages (5-10%) for new features
   - Monitor metrics before increasing rollout
   - Use overrides for specific tenants during testing

3. Cleanup:
   - Regularly review and remove unused flags
   - Set expiration dates in metadata
   - Use soft deletes initially, then hard delete after grace period

4. Monitoring:
   - Set up alerts for integrity check failures
   - Monitor evaluation patterns and performance
   - Track feature adoption rates

5. Documentation:
   - Document flag purpose and expected lifespan in description
   - Use metadata to store additional context
   - Maintain changelog of flag modifications

6. Security:
   - Audit all flag modifications
   - Use reason field for overrides
   - Regular review of admin actions

7. Performance:
   - Use cached evaluation for high-frequency checks
   - Refresh cache after bulk operations
   - Monitor query performance and optimize indexes

8. Testing:
   - Test both enabled and disabled states
   - Verify rollout percentages work as expected
   - Test override functionality
*/
