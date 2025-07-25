--- 1. Optimized materialized view for permission evaluations

-- Drop the materialized view
DROP MATERIALIZED VIEW IF EXISTS user_effective_permissions;

--- 2. Session clustering

-- This action (CLUSTER) reorganizes the table. There isn't a direct "down" for CLUSTER itself,
-- as it's an optimization. The table remains.

--- 3. Security Automation Functions:

-- Drop the session risk assessment function
DROP FUNCTION IF EXISTS assess_session_risk(UUID);

-- Drop the automatic session termination function
DROP FUNCTION IF EXISTS terminate_risky_sessions(INT);

--- 4. Compliance Enhancements:

-- Drop the GDPR right-to-forget implementation function
DROP FUNCTION IF EXISTS gdpr_user_deletion(UUID);

-- Drop the data retention policy enforcement function
DROP FUNCTION IF EXISTS enforce_data_retention();

--- 5. Advanced Threat Detection View:

-- Drop the security threat dashboard view
DROP VIEW IF EXISTS security_threat_dashboard;

--- 6. Index Optimizations for Large-Scale Deployments:

-- Drop BRIN Indexes
DROP INDEX IF EXISTS idx_audit_log_time_brin;
DROP INDEX IF EXISTS idx_user_sessions_time_brin;

-- Drop GIN optimizations
DROP INDEX IF EXISTS idx_users_attributes_gin;
DROP INDEX IF EXISTS idx_policies_rule_gin;

-- Drop Partial indexes
DROP INDEX IF EXISTS idx_active_users;
DROP INDEX IF EXISTS idx_active_roles;

