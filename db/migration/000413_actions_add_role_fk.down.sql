-- =============================================================================
-- V2.0 RESERVED — ABAC migration. DO NOT DROP. Not active in v1.0.
-- =============================================================================

ALTER TABLE actions DROP CONSTRAINT IF EXISTS fk_actions_approver_role;
