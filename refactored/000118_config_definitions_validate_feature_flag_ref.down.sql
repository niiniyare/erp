-- =============================================================================
-- ROLLBACK: Remove feature flag reference validation trigger
-- =============================================================================
DROP TRIGGER IF EXISTS config_def_validate_feature_flag ON config_definitions;
DROP FUNCTION IF EXISTS validate_feature_flag_reference();
