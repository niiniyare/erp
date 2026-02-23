-- Rollback audit functions and triggers for feature flag changes
-- =====================================================
-- DROP TRIGGERS
-- =====================================================
DROP TRIGGER IF EXISTS tenant_feature_overrides_audit_trigger ON tenant_feature_overrides;

DROP TRIGGER IF EXISTS feature_flags_audit_trigger ON feature_flags;

-- =====================================================
-- DROP FUNCTIONS
-- =====================================================
DROP FUNCTION IF EXISTS audit_tenant_feature_override_changes();

DROP FUNCTION IF EXISTS audit_feature_flag_changes();
