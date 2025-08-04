-- Rollback maintenance and cleanup functions

-- =====================================================
-- DROP FUNCTIONS
-- =====================================================
DROP FUNCTION IF EXISTS export_tenant_feature_flags(UUID);
DROP FUNCTION IF EXISTS bulk_create_feature_flags(JSONB, UUID);
DROP FUNCTION IF EXISTS bulk_update_rollout_percentage(TEXT[], INTEGER, UUID);
DROP FUNCTION IF EXISTS fix_orphaned_overrides();
DROP FUNCTION IF EXISTS check_feature_flags_integrity();
DROP FUNCTION IF EXISTS get_tenant_feature_flag_stats(UUID);
DROP FUNCTION IF EXISTS get_feature_flags_health_metrics();
DROP FUNCTION IF EXISTS cleanup_soft_deleted_feature_flags(INTEGER);
DROP FUNCTION IF EXISTS cleanup_old_feature_flag_audit_logs(INTEGER);