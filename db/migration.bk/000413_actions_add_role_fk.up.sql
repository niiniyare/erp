-- ------------------------------------------------------------------------------------------------
-- ACTIONS: FK TO ROLES
-- ------------------------------------------------------------------------------------------------
-- The actions.approver_role_id column was created in 000017 without a FK constraint because the
-- roles table didn't exist yet. Now that roles exists (000405), we add the constraint.
-- ------------------------------------------------------------------------------------------------
ALTER TABLE actions
  ADD CONSTRAINT fk_actions_approver_role
    FOREIGN KEY (approver_role_id) REFERENCES roles(id) ON DELETE SET NULL;
