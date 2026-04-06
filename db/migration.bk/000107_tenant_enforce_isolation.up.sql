-- ------------------------------------------------------------------------------------------------
-- ENFORCE_TENANT_ISOLATION
-- ------------------------------------------------------------------------------------------------
-- Trigger function that validates cross-table foreign key references belong to the same tenant,
-- preventing data leaking between tenants at the DB level.
-- Currently handles the 'persons' table; extend the IF/ELSIF chain for additional tables.
--
-- NOTE: Depends on entities(uuid, tenant_id) and tenants(id) defined in 000201.
--       Attach this trigger to each table that needs cross-tenant FK validation.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION enforce_tenant_isolation()
RETURNS TRIGGER AS $$
BEGIN
  -- Ensure all foreign key references belong to the same tenant
  IF TG_TABLE_NAME = 'persons' THEN
    -- Validate that the referenced entity belongs to the same tenant
    IF NOT EXISTS (
      SELECT 1
        FROM entities e
        JOIN tenants  t ON e.tenant_id = t.id
       WHERE e.uuid      = NEW.entity_id
         AND t.id        = NEW.tenant_id
    ) THEN
      RAISE EXCEPTION 'Entity % does not belong to tenant %', NEW.entity_id, NEW.tenant_id;
    END IF;

  ELSE
    RAISE NOTICE 'enforce_tenant_isolation trigger fired on table %', TG_TABLE_NAME;
  END IF;

  -- Add similar validations for other tables as needed
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
