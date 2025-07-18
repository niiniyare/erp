-- Seed entity data for ACME Corp
DO $$
DECLARE
    acme_tenant_id UUID;
    acme_corp_id UUID;
    us_ops_id UUID;
    sales_dept_id UUID;
BEGIN
    -- Get ACME tenant ID
    SELECT id INTO acme_tenant_id FROM tenants WHERE slug = 'acme-corp';
    
    IF acme_tenant_id IS NULL THEN
        RAISE EXCEPTION 'ACME tenant not found';
    END IF;
    
    -- Insert ACME Corporation root entity
    INSERT INTO entities (tenant_id, name, code, type, parent_id)
    VALUES (acme_tenant_id, 'ACME Corporation', 'ACME', 'COMPANY', NULL)
    ON CONFLICT (tenant_id, name) DO NOTHING
    RETURNING uuid INTO acme_corp_id;
    
    -- Get the ID if it already existed
    IF acme_corp_id IS NULL THEN
        SELECT uuid INTO acme_corp_id FROM entities 
        WHERE tenant_id = acme_tenant_id AND name = 'ACME Corporation';
    END IF;
    
    -- Insert US Operations
    INSERT INTO entities (tenant_id, name, code, type, parent_id)
    VALUES (acme_tenant_id, 'US Operations', 'US-OPS', 'REGION', acme_corp_id)
    ON CONFLICT (tenant_id, name) DO NOTHING
    RETURNING uuid INTO us_ops_id;
    
    IF us_ops_id IS NULL THEN
        SELECT uuid INTO us_ops_id FROM entities 
        WHERE tenant_id = acme_tenant_id AND name = 'US Operations';
    END IF;
    
    -- Insert Sales Department
    INSERT INTO entities (tenant_id, name, code, type, parent_id)
    VALUES (acme_tenant_id, 'Sales Department', 'SALES', 'DEPARTMENT', us_ops_id)
    ON CONFLICT (tenant_id, name) DO NOTHING
    RETURNING uuid INTO sales_dept_id;
    
    -- Rebuild hierarchy paths for ACME
    PERFORM rebuild_hierarchy_paths(acme_tenant_id);
END $$;
