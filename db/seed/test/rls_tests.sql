\echo 'Running Row Level Security tests...'

-- Test 1: Verify tenant isolation
DO $$
DECLARE
    acme_id UUID;
    globex_id UUID;
    acme_user_count INT;
    globex_user_count INT;
BEGIN
    SELECT id INTO acme_id FROM tenants WHERE slug = 'acme-corp';
    SELECT id INTO globex_id FROM tenants WHERE slug = 'globex';
    
    -- Set ACME context
    PERFORM set_tenant_context(acme_id);
    SELECT COUNT(*) INTO acme_user_count FROM users;
    
    -- Switch to Globex context
    PERFORM set_tenant_context(globex_id);
    SELECT COUNT(*) INTO globex_user_count FROM users;
    
    IF globex_user_count = 0 THEN
        RAISE NOTICE 'PASS: RLS shows % users for Globex (correct)', globex_user_count;
    ELSE
        RAISE EXCEPTION 'FAIL: RLS shows % users for Globex, expected 0', globex_user_count;
    END IF;
    
    -- Verify ACME user count maintained
    PERFORM set_tenant_context(acme_id);
    IF (SELECT COUNT(*) FROM users) = acme_user_count THEN
        RAISE NOTICE 'PASS: RLS maintains ACME user count: %', acme_user_count;
    ELSE
        RAISE EXCEPTION 'FAIL: ACME user count changed unexpectedly';
    END IF;
END $$;

-- Test 2: Verify entity isolation
DO $$
DECLARE
    acme_id UUID;
    globex_id UUID;
    acme_entity_count INT;
    globex_entity_count INT;
BEGIN
    SELECT id INTO acme_id FROM tenants WHERE slug = 'acme-corp';
    SELECT id INTO globex_id FROM tenants WHERE slug = 'globex';
    
    -- Test ACME entities
    PERFORM set_tenant_context(acme_id);
    SELECT COUNT(*) INTO acme_entity_count FROM entities;
    
    -- Test Globex entities
    PERFORM set_tenant_context(globex_id);
    SELECT COUNT(*) INTO globex_entity_count FROM entities;
    
    IF acme_entity_count = 3 AND globex_entity_count = 0 THEN
        RAISE NOTICE 'PASS: Entity isolation working correctly (ACME: %, Globex: %)', 
                     acme_entity_count, globex_entity_count;
    ELSE
        RAISE EXCEPTION 'FAIL: Entity isolation failed (ACME: %, Globex: %)', 
                        acme_entity_count, globex_entity_count;
    END IF;
END $$;

-- Test 3: Validate hierarchy isolation
DO $$
DECLARE
    acme_id UUID;
    globex_id UUID;
    acme_hierarchy_count INT;
    globex_hierarchy_count INT;
BEGIN
    SELECT id INTO acme_id FROM tenants WHERE slug = 'acme-corp';
    SELECT id INTO globex_id FROM tenants WHERE slug = 'globex';
    
    -- Test ACME hierarchy
    PERFORM set_tenant_context(acme_id);
    SELECT COUNT(*) INTO acme_hierarchy_count FROM hierarchy_paths;
    
    -- Test Globex hierarchy
    PERFORM set_tenant_context(globex_id);
    SELECT COUNT(*) INTO globex_hierarchy_count FROM hierarchy_paths;
    
    IF acme_hierarchy_count > 0 AND globex_hierarchy_count = 0 THEN
        RAISE NOTICE 'PASS: Hierarchy isolation working correctly (ACME: %, Globex: %)', 
                     acme_hierarchy_count, globex_hierarchy_count;
    ELSE
        RAISE EXCEPTION 'FAIL: Hierarchy isolation failed (ACME: %, Globex: %)', 
                        acme_hierarchy_count, globex_hierarchy_count;
    END IF;
END $$;

\echo 'Row Level Security tests completed.'
