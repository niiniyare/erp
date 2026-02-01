-- =====================================================
-- TENANT BULK OPERATIONS TRACKING MIGRATION
-- =====================================================
-- 
-- PURPOSE:
-- This migration creates tables and functions to track bulk tenant management 
-- operations such as mass suspend, reactivate, archive, and configuration updates.
-- 
-- TABLES CREATED:
-- 1. tenant_bulk_operations - Master table tracking bulk operations
-- 2. tenant_bulk_operation_results - Individual results per tenant in each operation
-- 
-- FEATURES:
-- - Complete audit trail for administrative bulk operations
-- - Progress tracking with status updates
-- - Error handling and detailed reporting
-- - Automatic count updates via triggers
-- - Row-level security for multi-tenant isolation
-- - Utility functions for operation monitoring
-- 
-- SECURITY:
-- - RLS enabled with policies for admin, application, and readonly roles
-- - Proper permission grants for different access levels
-- 
-- AUTHOR: ERP System Migration
-- VERSION: 1.0
-- DATE: 2024
-- =====================================================

-- =====================================================
-- MAIN TABLES
-- =====================================================

-- -----------------------------------------------------
-- BULK OPERATIONS MASTER TABLE
-- -----------------------------------------------------
-- Tracks metadata and overall status of bulk tenant operations
CREATE TABLE tenant_bulk_operations (
  -- Primary identification
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  
  -- Operation classification
  operation_type VARCHAR(50) NOT NULL CHECK (operation_type IN (
    'SUSPEND',           -- Mass suspend tenants
    'REACTIVATE',        -- Mass reactivate suspended tenants  
    'ARCHIVE',           -- Mass archive tenants with retention policies
    'UPDATE_LIMITS',     -- Mass update tenant resource limits
    'UPDATE_FEATURES'    -- Mass enable/disable tenant features
  )),
  
  -- Actor information (who initiated the operation)
  actor_id UUID NOT NULL,           -- User ID who started the operation
  actor_name VARCHAR(255),          -- User name for audit display
  
  -- Operation metrics
  total_tenants INT NOT NULL DEFAULT 0,      -- Total tenants in this operation
  successful_count INT NOT NULL DEFAULT 0,   -- Successfully processed tenants
  failed_count INT NOT NULL DEFAULT 0,       -- Failed tenant operations
  
  -- Operation lifecycle status
  status VARCHAR(20) NOT NULL DEFAULT 'IN_PROGRESS' CHECK (status IN (
    'IN_PROGRESS',       -- Operation is currently running
    'COMPLETED',         -- All operations completed successfully
    'FAILED',           -- All operations failed
    'PARTIAL_SUCCESS',  -- Some succeeded, some failed
    'CANCELLED'         -- Operation was cancelled by user
  )),
  
  -- Timing information
  started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  completed_at TIMESTAMPTZ,        -- Set when operation finishes
  
  -- Operation context and parameters
  parameters JSONB DEFAULT '{}'::jsonb,  -- Operation-specific data (reason, limits, etc.)
  error_summary TEXT,                     -- High-level error description if applicable
  
  -- Standard audit fields
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- -----------------------------------------------------
-- INDIVIDUAL OPERATION RESULTS TABLE
-- -----------------------------------------------------
-- Tracks the result of the bulk operation for each individual tenant
CREATE TABLE tenant_bulk_operation_results (
  -- Composite primary key linking to bulk operation and tenant
  operation_id UUID NOT NULL REFERENCES tenant_bulk_operations(id) ON DELETE CASCADE,
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  
  -- Individual operation status
  status VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (status IN (
    'PENDING',          -- Waiting to be processed
    'PROCESSING',       -- Currently being processed
    'COMPLETED',        -- Successfully completed
    'FAILED',          -- Operation failed for this tenant
    'SKIPPED'          -- Skipped (e.g., tenant already in target state)
  )),
  
  -- Result details
  message TEXT,           -- Success or informational message
  error_details TEXT,     -- Detailed error information if failed
  
  -- Individual timing (for performance analysis)
  started_at TIMESTAMPTZ,
  completed_at TIMESTAMPTZ,
  
  -- Standard audit fields
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  
  -- Composite primary key
  PRIMARY KEY (operation_id, tenant_id)
);

-- =====================================================
-- PERFORMANCE INDEXES
-- =====================================================

-- Indexes for bulk operations table
CREATE INDEX idx_tenant_bulk_operations_actor ON tenant_bulk_operations(actor_id);
CREATE INDEX idx_tenant_bulk_operations_status ON tenant_bulk_operations(status);
CREATE INDEX idx_tenant_bulk_operations_type ON tenant_bulk_operations(operation_type);
CREATE INDEX idx_tenant_bulk_operations_created_at ON tenant_bulk_operations(created_at);

-- Indexes for operation results table
CREATE INDEX idx_tenant_bulk_operation_results_tenant ON tenant_bulk_operation_results(tenant_id);
CREATE INDEX idx_tenant_bulk_operation_results_status ON tenant_bulk_operation_results(status);

-- =====================================================
-- TABLE DOCUMENTATION
-- =====================================================

COMMENT ON TABLE tenant_bulk_operations IS 'Master table tracking bulk tenant management operations with progress monitoring and audit trail';
COMMENT ON TABLE tenant_bulk_operation_results IS 'Individual operation results for each tenant within a bulk operation';

-- Column documentation for bulk operations
COMMENT ON COLUMN tenant_bulk_operations.operation_type IS 'Type of bulk operation: SUSPEND, REACTIVATE, ARCHIVE, UPDATE_LIMITS, UPDATE_FEATURES';
COMMENT ON COLUMN tenant_bulk_operations.parameters IS 'JSON parameters specific to operation type (reason, limits, features, retention policies, etc.)';
COMMENT ON COLUMN tenant_bulk_operations.actor_id IS 'UUID of administrator who initiated the bulk operation';
COMMENT ON COLUMN tenant_bulk_operations.error_summary IS 'High-level summary of errors if operation had failures';

-- Column documentation for operation results
COMMENT ON COLUMN tenant_bulk_operation_results.status IS 'Individual tenant operation status: PENDING, PROCESSING, COMPLETED, FAILED, SKIPPED';
COMMENT ON COLUMN tenant_bulk_operation_results.error_details IS 'Detailed error information specific to this tenant if operation failed';
COMMENT ON COLUMN tenant_bulk_operation_results.message IS 'Success message or additional context for this tenant operation';

-- =====================================================
-- ROW LEVEL SECURITY (RLS)
-- =====================================================

-- Enable RLS on both tables for multi-tenant security
ALTER TABLE tenant_bulk_operations ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_bulk_operation_results ENABLE ROW LEVEL SECURITY;

-- -----------------------------------------------------
-- RLS POLICIES
-- -----------------------------------------------------

-- Admin role: Full access to all bulk operations (for system administration)
CREATE POLICY tenant_bulk_operations_admin_policy 
  ON tenant_bulk_operations 
  FOR ALL TO admin_role 
  USING (true);

CREATE POLICY tenant_bulk_operation_results_admin_policy 
  ON tenant_bulk_operation_results 
  FOR ALL TO admin_role 
  USING (true);

-- Application role: Full access (for API operations)
CREATE POLICY tenant_bulk_operations_app_policy 
  ON tenant_bulk_operations 
  FOR ALL TO application_role 
  USING (true);

CREATE POLICY tenant_bulk_operation_results_app_policy 
  ON tenant_bulk_operation_results 
  FOR ALL TO application_role 
  USING (true);

-- Readonly role: Select access only (for monitoring and reporting)
CREATE POLICY tenant_bulk_operations_readonly_policy 
  ON tenant_bulk_operations 
  FOR SELECT TO readonly_role 
  USING (true);

CREATE POLICY tenant_bulk_operation_results_readonly_policy 
  ON tenant_bulk_operation_results 
  FOR SELECT TO readonly_role 
  USING (true);

-- =====================================================
-- ROLE PERMISSIONS
-- =====================================================

-- Admin role: Full CRUD permissions
GRANT SELECT, INSERT, UPDATE, DELETE ON tenant_bulk_operations TO admin_role;
GRANT SELECT, INSERT, UPDATE, DELETE ON tenant_bulk_operation_results TO admin_role;

-- Application role: Full CRUD permissions (for API operations)
GRANT SELECT, INSERT, UPDATE, DELETE ON tenant_bulk_operations TO application_role;
GRANT SELECT, INSERT, UPDATE, DELETE ON tenant_bulk_operation_results TO application_role;

-- Readonly role: Select permissions only (for monitoring dashboards)
GRANT SELECT ON tenant_bulk_operations TO readonly_role;
GRANT SELECT ON tenant_bulk_operation_results TO readonly_role;

-- =====================================================
-- AUTOMATIC TIMESTAMP TRIGGERS
-- =====================================================

-- Trigger to update 'updated_at' timestamp on bulk operations
CREATE TRIGGER update_tenant_bulk_operations_updated_at 
  BEFORE UPDATE ON tenant_bulk_operations 
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Trigger to update 'updated_at' timestamp on operation results
CREATE TRIGGER update_tenant_bulk_operation_results_updated_at 
  BEFORE UPDATE ON tenant_bulk_operation_results 
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =====================================================
-- UTILITY FUNCTIONS
-- =====================================================

-- -----------------------------------------------------
-- BULK OPERATION SUMMARY FUNCTION
-- -----------------------------------------------------
-- Returns comprehensive summary statistics for a bulk operation
CREATE OR REPLACE FUNCTION get_bulk_operation_summary(p_operation_id UUID)
RETURNS TABLE (
  operation_id UUID,
  operation_type VARCHAR(50),
  status VARCHAR(20),
  total_tenants INT,
  successful_count INT,
  failed_count INT,
  in_progress_count INT,
  duration_seconds INT
) AS $$
BEGIN
  /*
   * PURPOSE: Provides real-time summary of bulk operation progress
   * 
   * PARAMETERS:
   *   p_operation_id - UUID of the bulk operation to summarize
   * 
   * RETURNS:
   *   Complete summary including counts, status, and timing information
   * 
   * LOGIC:
   *   1. Join bulk operation with individual results
   *   2. Count results by status (successful, failed, in-progress)
   *   3. Calculate operation duration (completed or current)
   *   4. Return comprehensive summary for monitoring
   */
  
  RETURN QUERY
  SELECT 
    bo.id,                              -- Operation UUID
    bo.operation_type,                  -- Type of operation (SUSPEND, etc.)
    bo.status,                          -- Overall operation status
    bo.total_tenants,                   -- Total tenants targeted
    bo.successful_count,                -- Successfully processed count
    bo.failed_count,                    -- Failed operations count
    -- Count in-progress items dynamically from results table
    COUNT(CASE WHEN br.status IN ('PENDING', 'PROCESSING') THEN 1 END)::INT as in_progress_count,
    -- Calculate duration: completed operations use actual duration, in-progress use current time
    CASE 
      WHEN bo.completed_at IS NOT NULL 
      THEN EXTRACT(EPOCH FROM (bo.completed_at - bo.started_at))::INT
      ELSE EXTRACT(EPOCH FROM (NOW() - bo.started_at))::INT
    END as duration_seconds
  FROM tenant_bulk_operations bo
  LEFT JOIN tenant_bulk_operation_results br ON bo.id = br.operation_id
  WHERE bo.id = p_operation_id
  GROUP BY bo.id, bo.operation_type, bo.status, bo.total_tenants, 
           bo.successful_count, bo.failed_count, bo.started_at, bo.completed_at;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION get_bulk_operation_summary(UUID) IS 'Returns comprehensive real-time summary statistics for a bulk operation including progress and timing';

-- Grant execute permissions to all roles
GRANT EXECUTE ON FUNCTION get_bulk_operation_summary(UUID) TO admin_role;
GRANT EXECUTE ON FUNCTION get_bulk_operation_summary(UUID) TO application_role;
GRANT EXECUTE ON FUNCTION get_bulk_operation_summary(UUID) TO readonly_role;

-- -----------------------------------------------------
-- BULK OPERATION COUNT UPDATE FUNCTION
-- -----------------------------------------------------
-- Automatically updates success/failure counts and overall status
CREATE OR REPLACE FUNCTION update_bulk_operation_counts(p_operation_id UUID)
RETURNS VOID AS $$
DECLARE
  v_successful INT;     -- Count of successful operations
  v_failed INT;         -- Count of failed operations
  v_total INT;          -- Total operations
  v_in_progress INT;    -- Count of pending/processing operations
BEGIN
  /*
   * PURPOSE: Maintains accurate counts and status for bulk operations
   * 
   * PARAMETERS:
   *   p_operation_id - UUID of bulk operation to update
   * 
   * LOGIC:
   *   1. Count individual results by status (completed, failed, in-progress)
   *   2. Update the master record with current counts
   *   3. Determine overall status based on individual results:
   *      - IN_PROGRESS: if any items still pending/processing
   *      - COMPLETED: if all succeeded and none in progress
   *      - FAILED: if all failed and none in progress
   *      - PARTIAL_SUCCESS: if mixed results and none in progress
   *   4. Set completion timestamp when operation finishes
   */
  
  -- Count results by status category
  SELECT 
    COUNT(CASE WHEN status = 'COMPLETED' THEN 1 END),    -- Successful operations
    COUNT(CASE WHEN status = 'FAILED' THEN 1 END),       -- Failed operations
    COUNT(*),                                             -- Total operations
    COUNT(CASE WHEN status IN ('PENDING', 'PROCESSING') THEN 1 END)  -- Still in progress
  INTO v_successful, v_failed, v_total, v_in_progress
  FROM tenant_bulk_operation_results 
  WHERE operation_id = p_operation_id;
  
  -- Update the master bulk operation record with current counts and status
  UPDATE tenant_bulk_operations SET
    successful_count = v_successful,
    failed_count = v_failed,
    -- Determine overall status based on individual results
    status = CASE 
      WHEN v_in_progress > 0 THEN 'IN_PROGRESS'          -- Still processing
      WHEN v_failed = 0 THEN 'COMPLETED'                 -- All succeeded
      WHEN v_successful = 0 THEN 'FAILED'                -- All failed
      ELSE 'PARTIAL_SUCCESS'                             -- Mixed results
    END,
    -- Set completion timestamp when operation finishes (no more in-progress items)
    completed_at = CASE 
      WHEN v_in_progress = 0 AND completed_at IS NULL THEN NOW()
      ELSE completed_at
    END
  WHERE id = p_operation_id;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION update_bulk_operation_counts(UUID) IS 'Automatically updates bulk operation counts and status based on individual result statuses';

-- Grant execute permissions to roles that can modify data
GRANT EXECUTE ON FUNCTION update_bulk_operation_counts(UUID) TO admin_role;
GRANT EXECUTE ON FUNCTION update_bulk_operation_counts(UUID) TO application_role;

-- -----------------------------------------------------
-- AUTOMATIC COUNT UPDATE TRIGGER
-- -----------------------------------------------------
-- Trigger function that calls the count update function
CREATE OR REPLACE FUNCTION trigger_update_bulk_operation_counts()
RETURNS TRIGGER AS $$
BEGIN
  /*
   * PURPOSE: Trigger function to automatically update bulk operation counts
   * 
   * TRIGGER EVENTS: INSERT, UPDATE, DELETE on tenant_bulk_operation_results
   * 
   * LOGIC:
   *   1. Determine which operation_id was affected (from NEW or OLD record)
   *   2. Call update_bulk_operation_counts() to recalculate totals
   *   3. Return appropriate record for trigger chain continuation
   * 
   * NOTE: This ensures counts are always accurate without manual intervention
   */
  
  -- Update counts for the affected operation (handle all trigger events)
  PERFORM update_bulk_operation_counts(COALESCE(NEW.operation_id, OLD.operation_id));
  
  -- Return appropriate record based on trigger event
  RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Create the trigger on the results table
CREATE TRIGGER tenant_bulk_operation_results_update_counts
  AFTER INSERT OR UPDATE OR DELETE ON tenant_bulk_operation_results
  FOR EACH ROW EXECUTE FUNCTION trigger_update_bulk_operation_counts();

COMMENT ON TRIGGER tenant_bulk_operation_results_update_counts ON tenant_bulk_operation_results 
IS 'Automatically updates bulk operation counts and status whenever individual results change';

-- =====================================================
-- MIGRATION COMPLETION
-- =====================================================
-- Log successful migration completion
-- (This will appear in migration logs for debugging)

DO $$
BEGIN
  RAISE NOTICE 'Tenant bulk operations tracking migration completed successfully';
  RAISE NOTICE 'Created tables: tenant_bulk_operations, tenant_bulk_operation_results';
  RAISE NOTICE 'Created functions: get_bulk_operation_summary(), update_bulk_operation_counts()';
  RAISE NOTICE 'Configured RLS policies for admin_role, application_role, readonly_role';
  RAISE NOTICE 'Set up automatic count updating via triggers';
END $$;