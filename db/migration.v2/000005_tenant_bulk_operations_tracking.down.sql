-- =====================================================
-- TENANT BULK OPERATIONS TRACKING ROLLBACK
-- =====================================================
-- 
-- PURPOSE:
-- This rollback migration safely removes all tenant bulk operations 
-- tracking infrastructure created by the corresponding up migration.
-- 
-- REMOVAL ORDER:
-- 1. Drop triggers (dependent on functions)
-- 2. Drop functions (dependent on tables)
-- 3. Drop indexes (performance optimization cleanup)
-- 4. Drop tables (in dependency order)
-- 
-- SAFETY:
-- - Uses IF EXISTS to prevent errors if objects don't exist
-- - Removes dependencies in correct order
-- - Preserves referential integrity
-- 
-- AUTHOR: ERP System Migration
-- VERSION: 1.0
-- DATE: 2024
-- =====================================================

-- Log rollback start
DO $$
BEGIN
  RAISE NOTICE 'Starting tenant bulk operations tracking rollback migration';
END $$;

-- =====================================================
-- DROP TRIGGERS
-- =====================================================

-- Drop automatic count update trigger
DROP TRIGGER IF EXISTS tenant_bulk_operation_results_update_counts ON tenant_bulk_operation_results;

-- Drop timestamp update triggers
DROP TRIGGER IF EXISTS update_tenant_bulk_operation_results_updated_at ON tenant_bulk_operation_results;
DROP TRIGGER IF EXISTS update_tenant_bulk_operations_updated_at ON tenant_bulk_operations;

-- =====================================================
-- DROP FUNCTIONS
-- =====================================================

-- Drop trigger function
DROP FUNCTION IF EXISTS trigger_update_bulk_operation_counts();

-- Drop utility functions
DROP FUNCTION IF EXISTS update_bulk_operation_counts(UUID);
DROP FUNCTION IF EXISTS get_bulk_operation_summary(UUID);

-- =====================================================
-- DROP INDEXES
-- =====================================================

-- Drop indexes for tenant_bulk_operation_results table
DROP INDEX IF EXISTS idx_tenant_bulk_operation_results_status;
DROP INDEX IF EXISTS idx_tenant_bulk_operation_results_tenant;

-- Drop indexes for tenant_bulk_operations table
DROP INDEX IF EXISTS idx_tenant_bulk_operations_created_at;
DROP INDEX IF EXISTS idx_tenant_bulk_operations_type;
DROP INDEX IF EXISTS idx_tenant_bulk_operations_status;
DROP INDEX IF EXISTS idx_tenant_bulk_operations_actor;

-- =====================================================
-- DROP TABLES
-- =====================================================

-- Drop tables in dependency order (child tables first)
DROP TABLE IF EXISTS tenant_bulk_operation_results;
DROP TABLE IF EXISTS tenant_bulk_operations;

-- =====================================================
-- ROLLBACK COMPLETION
-- =====================================================

-- Log successful rollback completion
DO $$
BEGIN
  RAISE NOTICE 'Tenant bulk operations tracking rollback completed successfully';
  RAISE NOTICE 'Removed: triggers, functions, indexes, and tables';
  RAISE NOTICE 'Database restored to state before bulk operations tracking migration';
END $$;