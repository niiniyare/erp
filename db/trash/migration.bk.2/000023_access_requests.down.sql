-- =====================================================================
-- ACCESS REQUESTS DOWN MIGRATION
-- =====================================================================

-- Drop RLS policies
DROP POLICY IF EXISTS access_requests_tenant_isolation ON access_requests;

-- Disable RLS
ALTER TABLE IF EXISTS access_requests DISABLE ROW LEVEL SECURITY;

-- Drop table
DROP TABLE IF EXISTS access_requests;

-- =====================================================================
-- MIGRATION COMPLETION MESSAGE
-- =====================================================================

DO $$
BEGIN
    RAISE NOTICE '===================================================================';
    RAISE NOTICE 'ACCESS REQUESTS DOWN MIGRATION COMPLETED';
    RAISE NOTICE '===================================================================';
    RAISE NOTICE 'Table dropped: access_requests';
    RAISE NOTICE 'RLS disabled and policy dropped.';
    RAISE NOTICE '===================================================================';
END;
$$;
