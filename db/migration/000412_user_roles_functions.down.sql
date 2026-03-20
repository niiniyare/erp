ALTER TABLE IF EXISTS role_assignments
    DROP CONSTRAINT IF EXISTS role_assignments_granted_by_fk;

DROP FUNCTION IF EXISTS assign_user_role(UUID, UUID, UUID, UUID);

DROP FUNCTION IF EXISTS revoke_user_role(UUID, UUID, UUID);
