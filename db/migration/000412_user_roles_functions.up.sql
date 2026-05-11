-- =============================================================================
-- V2.0 RESERVED — ABAC migration. DO NOT DROP. Not active in v1.0.
-- internal/core/access/ is gated with //go:build ignore until v2.0.
-- =============================================================================

-- ------------------------------------------------------------------------------------------------
-- ROLE_ASSIGNMENTS: DEFERRED FK + ROLE ASSIGNMENT HELPERS
-- ------------------------------------------------------------------------------------------------
-- Adds the deferred FK from role_assignments.granted_by to users(id).
-- role_assignments (000064) is created before users (000303), so the FK must be added here.
-- Also provides assign_user_role and revoke_user_role convenience functions.
-- ------------------------------------------------------------------------------------------------

-- ------------------------------------------------------------------------------------------------
-- DEFERRED FOREIGN KEY
-- ------------------------------------------------------------------------------------------------
ALTER TABLE role_assignments
  ADD CONSTRAINT role_assignments_granted_by_fk
    FOREIGN KEY (granted_by) REFERENCES users(id) ON DELETE SET NULL;

-- ------------------------------------------------------------------------------------------------
-- ASSIGN_USER_ROLE
-- ------------------------------------------------------------------------------------------------
-- Inserts a row into user_roles linking a user to a role scoped to an entity.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION assign_user_role(
  p_user_id    UUID,
  p_role_id    UUID,
  p_entity_id  UUID,
  p_assigned_by UUID
) RETURNS VOID AS $$
BEGIN
  INSERT INTO user_roles (user_id, role_id, entity_id, assigned_by)
  VALUES (p_user_id, p_role_id, p_entity_id, p_assigned_by);
END;
$$ LANGUAGE plpgsql SECURITY INVOKER;

-- ------------------------------------------------------------------------------------------------
-- REVOKE_USER_ROLE
-- ------------------------------------------------------------------------------------------------
-- Removes a specific user-role-entity combination from user_roles.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION revoke_user_role(
  p_user_id   UUID,
  p_role_id   UUID,
  p_entity_id UUID
) RETURNS VOID AS $$
BEGIN
  DELETE FROM user_roles
  WHERE user_id = p_user_id
    AND role_id  = p_role_id
    AND entity_id = p_entity_id;
END;
$$ LANGUAGE plpgsql SECURITY INVOKER;
