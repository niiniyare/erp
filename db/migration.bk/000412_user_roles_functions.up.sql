CREATE
OR REPLACE FUNCTION assign_user_role(
  p_user_id UUID,
  p_role_id UUID,
  p_entity_id UUID,
  p_assigned_by UUID
) RETURNS VOID AS
$$
BEGIN
INSERT INTO
  user_roles (user_id, role_id, entity_id, assigned_by)
VALUES
  (p_user_id, p_role_id, p_entity_id, p_assigned_by);

END;

$$
LANGUAGE plpgsql SECURITY INVOKER;

CREATE
OR REPLACE FUNCTION revoke_user_role(
  p_user_id UUID,
  p_role_id UUID,
  p_entity_id UUID
) RETURNS VOID AS
$$
BEGIN
DELETE FROM
  user_roles
WHERE
  user_id = p_user_id
  AND role_id = p_role_id
  AND entity_id = p_entity_id;

END;

$$
LANGUAGE plpgsql SECURITY INVOKER;
