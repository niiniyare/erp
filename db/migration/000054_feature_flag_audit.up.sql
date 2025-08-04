-- Creates audit logging functions and triggers for feature flag changes

-- =====================================================
-- AUDIT TRIGGER FUNCTIONS
-- =====================================================

-- Function to create audit log entries for feature flag changes
CREATE OR REPLACE FUNCTION audit_feature_flag_changes()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
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
            NEW.tenant_id,
            'FEATURE_FLAG_CREATED',
            'ADMIN',
            'INFO',
            NULLIF(current_setting('app.current_user_id', true), '')::UUID,
            'ALLOW',
            'Feature flag created: ' || NEW.name,
            jsonb_build_object(
                'feature_flag_id', NEW.id,
                'feature_flag_name', NEW.name,
                'flag_type', NEW.flag_type,
                'default_value', NEW.default_value,
                'rollout_percentage', NEW.rollout_percentage,
                'metadata', NEW.metadata,
                'operation', 'CREATE'
            ),
            NULLIF(current_setting('app.current_session_id', true), '')::UUID
        );
        RETURN NEW;
        
    ELSIF TG_OP = 'UPDATE' THEN
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
            NEW.tenant_id,
            'FEATURE_FLAG_UPDATED',
            'ADMIN',
            CASE 
                WHEN OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL THEN 'WARN'
                ELSE 'INFO'
            END,
            NULLIF(current_setting('app.current_user_id', true), '')::UUID,
            'ALLOW',
            'Feature flag updated: ' || NEW.name,
            jsonb_build_object(
                'feature_flag_id', NEW.id,
                'feature_flag_name', NEW.name,
                'old_values', jsonb_build_object(
                    'flag_type', OLD.flag_type,
                    'default_value', OLD.default_value,
                    'rollout_percentage', OLD.rollout_percentage,
                    'deleted_at', OLD.deleted_at,
                    'metadata', OLD.metadata
                ),
                'new_values', jsonb_build_object(
                    'flag_type', NEW.flag_type,
                    'default_value', NEW.default_value,
                    'rollout_percentage', NEW.rollout_percentage,
                    'deleted_at', NEW.deleted_at,
                    'metadata', NEW.metadata
                ),
                'operation', CASE 
                    WHEN OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL THEN 'SOFT_DELETE'
                    ELSE 'UPDATE'
                END
            ),
            NULLIF(current_setting('app.current_session_id', true), '')::UUID
        );
        RETURN NEW;
        
    ELSIF TG_OP = 'DELETE' THEN
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
            OLD.tenant_id,
            'FEATURE_FLAG_DELETED',
            'ADMIN',
            'WARN',
            NULLIF(current_setting('app.current_user_id', true), '')::UUID,
            'ALLOW',
            'Feature flag permanently deleted: ' || OLD.name,
            jsonb_build_object(
                'feature_flag_id', OLD.id,
                'feature_flag_name', OLD.name,
                'deleted_values', to_jsonb(OLD),
                'operation', 'HARD_DELETE'
            ),
            NULLIF(current_setting('app.current_session_id', true), '')::UUID
        );
        RETURN OLD;
    END IF;
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Function to create audit entries for tenant feature override changes
CREATE OR REPLACE FUNCTION audit_tenant_feature_override_changes()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO audit_log (
            tenant_id,
            event_type,
            event_category,
            severity,
            user_id,
            decision,
            reason,
            risk_score,
            context,
            session_id,
            compliance_flags
        ) VALUES (
            NEW.tenant_id,
            'FEATURE_OVERRIDE_CREATED',
            'ADMIN',
            'INFO',
            NULLIF(current_setting('app.current_user_id', true), '')::UUID,
            'ALLOW',
            COALESCE(NEW.reason, 'Feature override created for ' || NEW.feature_flag_name),
            CASE WHEN NEW.enabled THEN 10 ELSE 5 END, -- Higher risk when enabling features
            jsonb_build_object(
                'feature_flag_id', NEW.feature_flag_id,
                'feature_flag_name', NEW.feature_flag_name,
                'enabled', NEW.enabled,
                'value', NEW.value,
                'operation', 'CREATE_OVERRIDE'
            ),
            NULLIF(current_setting('app.current_session_id', true), '')::UUID,
            jsonb_build_object('feature_management', true)
        );
        RETURN NEW;
        
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO audit_log (
            tenant_id,
            event_type,
            event_category,
            severity,
            user_id,
            decision,
            reason,
            risk_score,
            context,
            session_id,
            compliance_flags
        ) VALUES (
            NEW.tenant_id,
            'FEATURE_OVERRIDE_UPDATED',
            'ADMIN',
            CASE 
                WHEN OLD.enabled != NEW.enabled THEN 'WARN'
                ELSE 'INFO'
            END,
            NULLIF(current_setting('app.current_user_id', true), '')::UUID,
            'ALLOW',
            COALESCE(NEW.reason, 'Feature override updated for ' || NEW.feature_flag_name),
            CASE 
                WHEN OLD.enabled != NEW.enabled THEN 15
                ELSE 8
            END,
            jsonb_build_object(
                'feature_flag_id', NEW.feature_flag_id,
                'feature_flag_name', NEW.feature_flag_name,
                'old_values', jsonb_build_object(
                    'enabled', OLD.enabled,
                    'value', OLD.value
                ),
                'new_values', jsonb_build_object(
                    'enabled', NEW.enabled,
                    'value', NEW.value
                ),
                'operation', 'UPDATE_OVERRIDE'
            ),
            NULLIF(current_setting('app.current_session_id', true), '')::UUID,
            jsonb_build_object('feature_management', true)
        );
        RETURN NEW;
        
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO audit_log (
            tenant_id,
            event_type,
            event_category,
            severity,
            user_id,
            decision,
            reason,
            risk_score,
            context,
            session_id,
            compliance_flags
        ) VALUES (
            OLD.tenant_id,
            'FEATURE_OVERRIDE_DELETED',
            'ADMIN',
            'INFO',
            NULLIF(current_setting('app.current_user_id', true), '')::UUID,
            'ALLOW',
            'Feature override deleted for ' || OLD.feature_flag_name,
            5,
            jsonb_build_object(
                'feature_flag_id', OLD.feature_flag_id,
                'feature_flag_name', OLD.feature_flag_name,
                'deleted_values', jsonb_build_object(
                    'enabled', OLD.enabled,
                    'value', OLD.value
                ),
                'operation', 'DELETE_OVERRIDE'
            ),
            NULLIF(current_setting('app.current_session_id', true), '')::UUID,
            jsonb_build_object('feature_management', true)
        );
        RETURN OLD;
    END IF;
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- CREATE AUDIT TRIGGERS
-- =====================================================

-- Audit trigger for feature_flags table
CREATE TRIGGER feature_flags_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON feature_flags
    FOR EACH ROW
    EXECUTE FUNCTION audit_feature_flag_changes();

-- Audit trigger for tenant_feature_overrides table
CREATE TRIGGER tenant_feature_overrides_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON tenant_feature_overrides
    FOR EACH ROW
    EXECUTE FUNCTION audit_tenant_feature_override_changes();

-- =====================================================
-- FUNCTION COMMENTS
-- =====================================================
COMMENT ON FUNCTION audit_feature_flag_changes() IS 'Creates audit log entries for feature flag changes using existing audit_log table';
COMMENT ON FUNCTION audit_tenant_feature_override_changes() IS 'Creates audit entries for tenant feature override changes using existing audit_log table';
COMMENT ON TRIGGER feature_flags_audit_trigger ON feature_flags IS 'Logs all feature flag changes to audit_log table';
COMMENT ON TRIGGER tenant_feature_overrides_audit_trigger ON tenant_feature_overrides IS 'Logs all tenant override changes to audit_log table';
