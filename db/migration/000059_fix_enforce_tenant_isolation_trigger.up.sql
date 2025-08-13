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
    ELSE
        RAISE NOTICE 'enforce_tenant_isolation trigger fired on table %', TG_TABLE_NAME;
    END IF;

    -- Add similar validations for other tables as needed
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
