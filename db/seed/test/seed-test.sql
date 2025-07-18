-- ================================================
-- TEST 1: Basic Functionality Verification
-- ================================================

-- Verify functions exist
SELECT proname, pg_get_functiondef(oid) 
FROM pg_proc 
WHERE proname IN ('set_tenant_context', 'current_tenant_id');

-- ================================================
-- TEST 2: Valid Tenant Context Operations
-- ================================================

-- Get a valid active tenant
SELECT id AS valid_tenant_id FROM tenants 
WHERE status = 'active' AND deleted_at IS NULL LIMIT 1 \gset

-- Set tenant context (should succeed)
SELECT set_tenant_context(:'valid_tenant_id');

-- Verify current tenant
SELECT current_tenant_id() AS current_tenant;

-- Should match the tenant we set
SELECT :'valid_tenant_id' = current_tenant_id()::text AS is_match;

-- ================================================
-- TEST 3: Invalid Tenant Handling
-- ================================================

-- Test with invalid UUID
BEGIN;
DO $$
BEGIN
    PERFORM set_tenant_context('00000000-0000-0000-0000-000000000000');
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'Expected error: %', SQLERRM;
END
$$;
ROLLBACK;

-- Test with suspended tenant
WITH suspended AS (
    UPDATE tenants SET status = 'suspended' 
    WHERE id = :'valid_tenant_id' RETURNING id
)
DO $$
BEGIN
    PERFORM set_tenant_context((SELECT id FROM suspended));
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'Expected error: %', SQLERRM;
END
$$;

-- Test with deleted tenant
WITH deleted AS (
    UPDATE tenants SET deleted_at = NOW() 
    WHERE id = :'valid_tenant_id' RETURNING id
)
DO $$
BEGIN
    PERFORM set_tenant_context((SELECT id FROM deleted));
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'Expected error: %', SQLERRM;
END
$$;

-- ================================================
-- TEST 4: Session Persistence Verification
-- ================================================

-- Set tenant context
SELECT set_tenant_context(:'valid_tenant_id');

-- Verify context persists across queries
SELECT current_tenant_id() AS first_check \gset
SELECT :'first_check' AS stored_value;

-- Verify context survives transaction boundaries
BEGIN;
SELECT current_tenant_id() AS in_transaction;
COMMIT;

-- ================================================
-- TEST 5: Row-Level Security Integration
-- ================================================

-- Create test data across tenants
INSERT INTO persons (tenant_id, first_name, last_name) 
SELECT id, 'Test', 'User' FROM tenants;

-- Without tenant context (should return 0 rows)
SELECT * FROM persons WHERE first_name = 'Test';

-- With tenant context (should return 1 row)
SELECT set_tenant_context(:'valid_tenant_id');
SELECT COUNT(*) FROM persons WHERE first_name = 'Test';

-- Verify tenant isolation
SELECT 
    (SELECT COUNT(*) FROM persons) AS current_tenant_count,
    (SELECT COUNT(*) FROM persons WHERE tenant_id != :'valid_tenant_id') AS other_tenants_count;

-- ================================================
-- TEST 6: Concurrent Session Handling
-- ================================================

-- Simulate different sessions in transaction blocks
-- Session A
BEGIN;
SELECT set_tenant_context(:'valid_tenant_id');
SELECT pg_sleep(1);  -- Simulate work
SELECT current_tenant_id() AS session_a_tenant;

-- Session B (in new connection would be separate, simulated here)
DO $$
DECLARE 
    other_tenant UUID;
BEGIN
    SELECT id INTO other_tenant 
    FROM tenants 
    WHERE id != :'valid_tenant_id'
    LIMIT 1;
    
    PERFORM set_tenant_context(other_tenant);
    RAISE NOTICE 'Session B tenant: %', current_tenant_id();
END
$$;

-- Session A should maintain its context
SELECT current_tenant_id() AS session_a_after_b;
COMMIT;

-- ================================================
-- TEST 7: Edge Case Handling
-- ================================================

-- Null input test
DO $$
BEGIN
    PERFORM set_tenant_context(NULL);
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'Null test error: %', SQLERRM;
END
$$;

-- Invalid UUID format
DO $$
BEGIN
    PERFORM set_tenant_context('not-a-uuid');
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'Format test error: %', SQLERRM;
END
$$;

-- Verify current_tenant_id() without setting
RESET app.current_tenant_id;
SELECT current_tenant_id() IS NULL AS is_null_without_setting;

-- ================================================
-- TEST 8: Security Privilege Verification
-- ================================================

-- Test as non-privileged user
CREATE ROLE test_user NOLOGIN;
GRANT SELECT ON tenants TO test_user;

SET ROLE test_user;

-- Should fail (security definer requires tenant access)
DO $$
BEGIN
    PERFORM set_tenant_context(:'valid_tenant_id');
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'Privilege test error: %', SQLERRM;
END
$$;

RESET ROLE;
DROP ROLE test_user;

-- ================================================
-- TEST 9: Performance Benchmarking
-- ================================================

-- Timing test (should be < 1ms per call)
\timing on
SELECT set_tenant_context(:'valid_tenant_id');
SELECT current_tenant_id();
\timing off

-- Bulk performance test
DO $$
DECLARE
    t_id UUID;
    start_time TIMESTAMP;
    iterations INT := 1000;
BEGIN
    SELECT id INTO t_id FROM tenants LIMIT 1;
    start_time := clock_timestamp();
    
    FOR i IN 1..iterations LOOP
        PERFORM set_tenant_context(t_id);
        PERFORM current_tenant_id();
    END LOOP;
    
    RAISE NOTICE '%, executions: %', 
        clock_timestamp() - start_time, 
        iterations;
END
$$;
