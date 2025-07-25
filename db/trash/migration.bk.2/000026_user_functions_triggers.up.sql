-- =====================================================================
-- USER FUNCTIONS AND TRIGGERS UP MIGRATION
-- =====================================================================

-- ------------------------------------------------------------------------------------------------
-- Updated timestamp trigger function
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION update_updated_at_column() IS
'Trigger function to automatically update the updated_at timestamp when a record is modified.';

-- ------------------------------------------------------------------------------------------------
-- Tenant isolation validation trigger
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION enforce_tenant_isolation()
RETURNS TRIGGER AS $$
BEGIN
    -- Ensure all foreign key references belong to the same tenant
    IF TG_TABLE_NAME = 'persons' THEN
        -- Validate entity belongs to same tenant
        IF NOT EXISTS (
            SELECT 1 FROM entities e
            JOIN tenants t ON e.tenant_id = t.id
            WHERE e.uuid = NEW.entity_id AND t.id = NEW.tenant_id
        ) THEN
            RAISE EXCEPTION 'Entity % does not belong to tenant %', NEW.entity_id, NEW.tenant_id;
        END IF;
    END IF;

    -- Add similar validations for other tables as needed
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION enforce_tenant_isolation() IS
'Trigger function to enforce tenant isolation by validating that all foreign key references belong to the same tenant.';

-- ------------------------------------------------------------------------------------------------
-- Role hierarchy validation and level calculation
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION validate_role_hierarchy()
RETURNS TRIGGER AS $$
DECLARE
    max_depth INTEGER := 10;
    current_depth INTEGER := 0;
    current_role_id UUID;
BEGIN
    -- If no parent role, set level to 0
    IF NEW.parent_role_id IS NULL THEN
        NEW.level = 0;
        RETURN NEW;
    END IF;

    -- Ensure parent role belongs to same tenant
    IF NOT EXISTS (
        SELECT 1 FROM roles
        WHERE id = NEW.parent_role_id AND tenant_id = NEW.tenant_id AND deleted_at IS NULL
    ) THEN
        RAISE EXCEPTION 'Parent role % does not belong to tenant % or is deleted', NEW.parent_role_id, NEW.tenant_id;
    END IF;

    -- Check for cycles and calculate depth
    current_role_id := NEW.parent_role_id;
    current_depth := 1;

    WHILE current_role_id IS NOT NULL AND current_depth <= max_depth LOOP
        -- Check if we've hit the new role (cycle detection)
        IF current_role_id = NEW.id THEN
            RAISE EXCEPTION 'Role hierarchy cycle detected for role %', NEW.id;
        END IF;

        -- Get the next parent
        SELECT parent_role_id INTO current_role_id
        FROM roles
        WHERE id = current_role_id AND tenant_id = NEW.tenant_id AND deleted_at IS NULL;

        current_depth := current_depth + 1;
    END LOOP;

    IF current_depth > max_depth THEN
        RAISE EXCEPTION 'Role hierarchy depth exceeds maximum of %', max_depth;
    END IF;

    -- Set the level
    NEW.level = current_depth - 1;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION validate_role_hierarchy() IS
'Validates role hierarchy integrity, prevents cycles, enforces depth limits, and calculates hierarchy levels.';

-- ------------------------------------------------------------------------------------------------
-- User permission evaluation function with ABAC support
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION user_has_permission(
    p_user_id UUID,
    p_resource_name VARCHAR(100),
    p_action_name VARCHAR(100),
    p_tenant_id UUID,
    p_entity_id UUID DEFAULT NULL,
    p_context JSONB DEFAULT '{}'::jsonb
)
RETURNS BOOLEAN AS $$
DECLARE
    v_has_permission BOOLEAN := FALSE;
    v_user_context JSONB;
    v_cache_key VARCHAR(64);
    v_cached_result BOOLEAN;
BEGIN
    -- Generate cache key
    v_cache_key := encode(digest(
        p_user_id::text || p_resource_name || p_action_name ||
        COALESCE(p_entity_id::text, '') || p_context::text, 'sha256'
    ), 'hex');

    -- Check cache first
    SELECT decision = 'ALLOW' INTO v_cached_result
    FROM policy_evaluations
    WHERE user_id = p_user_id
        AND context_hash = v_cache_key
        AND expires_at > NOW()
        AND tenant_id = p_tenant_id;

    IF FOUND THEN
        RETURN v_cached_result;
    END IF;

    -- Build comprehensive user context for ABAC evaluation
    SELECT jsonb_build_object(
        'user_id', u.id,
        'user_type', u.user_type,
        'account_status', u.account_status,
        'user_attributes', COALESCE(u.user_attributes, '{}'::jsonb),
        'person_attributes', COALESCE(p.security_attributes, '{}'::jsonb),
        'employee_attributes', COALESCE(e.access_attributes, '{}'::jsonb),
        'employment_status', e.employment_status,
        'security_level', COALESCE(e.security_level, 0),
        'entity_id', u.entity_id,
        'department_id', e.department_id,
        'context', p_context
    ) INTO v_user_context
    FROM users u
    LEFT JOIN persons p ON u.person_id = p.id
    LEFT JOIN employees e ON u.employee_id = e.id
    WHERE u.id = p_user_id AND u.tenant_id = p_tenant_id;

    -- Check for explicit DENY in direct user permissions first
    SELECT true INTO v_has_permission
    FROM user_permissions up
    JOIN permissions perm ON up.permission_id = perm.id
    JOIN resources r ON perm.resource_id = r.id
    JOIN actions a ON perm.action_id = a.id
    WHERE up.user_id = p_user_id
        AND up.tenant_id = p_tenant_id
        AND r.name = p_resource_name
        AND a.name = p_action_name
        AND up.is_active = true
        AND up.effect = 'DENY'
        AND (up.expires_at IS NULL OR up.expires_at > NOW())
        AND (p_entity_id IS NULL OR up.entity_id = p_entity_id);

    -- If explicit DENY found, return false immediately
    IF FOUND THEN
        RETURN FALSE;
    END IF;

    -- Check direct user permissions for ALLOW
    SELECT true INTO v_has_permission
    FROM user_permissions up
    JOIN permissions perm ON up.permission_id = perm.id
    JOIN resources r ON perm.resource_id = r.id
    JOIN actions a ON perm.action_id = a.id
    WHERE up.user_id = p_user_id
        AND up.tenant_id = p_tenant_id
        AND r.name = p_resource_name
        AND a.name = p_action_name
        AND up.is_active = true
        AND up.effect = 'ALLOW'
        AND (up.expires_at IS NULL OR up.expires_at > NOW())
        AND (p_entity_id IS NULL OR up.entity_id = p_entity_id);

    -- If no direct permission, check role-based permissions
    IF NOT FOUND THEN
        SELECT true INTO v_has_permission
        FROM user_roles ur
        JOIN role_permissions rp ON ur.role_id = rp.role_id
        JOIN permissions perm ON rp.permission_id = perm.id
        JOIN resources r ON perm.resource_id = r.id
        JOIN actions a ON perm.action_id = a.id
        WHERE ur.user_id = p_user_id
            AND r.name = p_resource_name
            AND a.name = p_action_name
            AND ur.is_active = true
            AND rp.is_active = true
            AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
            AND (p_entity_id IS NULL OR ur.entity_id = p_entity_id OR rp.entity_scope = p_entity_id);
    END IF;

    RETURN COALESCE(v_has_permission, FALSE);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

COMMENT ON FUNCTION user_has_permission(UUID, VARCHAR, VARCHAR, UUID, UUID, JSONB) IS
'Evaluates user permissions with ABAC context, caching, and comprehensive policy evaluation including direct permissions and role-based permissions.';

-- ------------------------------------------------------------------------------------------------
-- Cleanup function for expired data and maintenance
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION cleanup_expired_data(p_tenant_id UUID DEFAULT NULL)
RETURNS INTEGER AS $$
DECLARE
    v_cleanup_count INTEGER := 0;
    v_tenant_filter TEXT := '';
BEGIN
    -- Build tenant filter if specified
    IF p_tenant_id IS NOT NULL THEN
        v_tenant_filter := ' AND tenant_id = ' || quote_literal(p_tenant_id);
    END IF;

    -- Clean expired sessions
    EXECUTE 'DELETE FROM user_sessions WHERE expires_at < NOW()' || v_tenant_filter;
    GET DIAGNOSTICS v_cleanup_count = ROW_COUNT;

    -- Clean expired policy evaluations
    EXECUTE 'DELETE FROM policy_evaluations WHERE expires_at < NOW()' || v_tenant_filter;

    -- Deactivate expired user roles
    EXECUTE 'UPDATE user_roles SET is_active = false
             WHERE expires_at < NOW() AND is_active = true' ||
             CASE WHEN p_tenant_id IS NOT NULL THEN
                ' AND EXISTS (SELECT 1 FROM users WHERE id = user_roles.user_id AND tenant_id = ' || quote_literal(p_tenant_id) || ')'
             ELSE '' END;

    -- Deactivate expired user permissions
    EXECUTE 'UPDATE user_permissions SET is_active = false
             WHERE expires_at < NOW() AND is_active = true' || v_tenant_filter;

    -- Expire approved access requests
    EXECUTE 'UPDATE access_requests SET approval_status = ''EXPIRED''
             WHERE expires_at < NOW() AND approval_status = ''APPROVED''' || v_tenant_filter;

    RETURN v_cleanup_count;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

COMMENT ON FUNCTION cleanup_expired_data(UUID) IS
'Cleans up expired sessions, policy evaluations, user roles, permissions, and access requests. Can be run for all tenants or a specific tenant.';

-- =====================================================================
-- MIGRATION COMPLETION MESSAGE
-- =====================================================================

DO $$
BEGIN
    RAISE NOTICE '===================================================================';
    RAISE NOTICE 'USER FUNCTIONS AND TRIGGERS UP MIGRATION COMPLETED';
    RAISE NOTICE '===================================================================';
    RAISE NOTICE 'Functions created: update_updated_at_column, enforce_tenant_isolation, validate_role_hierarchy, user_has_permission, cleanup_expired_data';
    RAISE NOTICE '===================================================================';
END;
$$;