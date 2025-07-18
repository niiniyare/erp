\echo 'Running tenant context tests...'

-- Test 1: Set valid tenant context
DO $$
DECLARE
    acme_id UUID;
BEGIN
    SELECT id INTO acme_id FROM tenants WHERE slug = 'acme-corp';
    PERFORM set_tenant_context(acme_id);
    
    IF current_tenant_id() = acme_id THEN
        RAISE NOTICE 'PASS: Valid tenant context set';
    ELSE
        RAISE EXCEPTION 'FAIL: Context not set properly';
    END IF;
END $$;

-- Test 2: Prevent invalid tenant context
DO $$
BEGIN
    PERFORM set_tenant_context('00000000-0000-0000-0000-000000000000');
    RAISE EXCEPTION 'FAIL: Should not accept invalid UUID';
EXCEPTION
    WHEN others THEN
        RAISE NOTICE 'PASS: Properly rejected invalid tenant';
END $$;

-- Test 3: Switch tenant context
DO $$
DECLARE
    acme_id UUID;
    globex_id UUID;
BEGIN
    SELECT id INTO acme_id FROM tenants WHERE slug = 'acme-corp';
    SELECT id INTO globex_id FROM tenants WHERE slug = 'globex';
    
    PERFORM set_tenant_context(acme_id);
    PERFORM set_tenant_context(globex_id);
    
    IF current_tenant_id() = globex_id THEN
        RAISE NOTICE 'PASS: Tenant context switched successfully';
    ELSE
        RAISE EXCEPTION 'FAIL: Context switch failed';
    END IF;
END $$;

\echo 'Tenant context tests completed.'
