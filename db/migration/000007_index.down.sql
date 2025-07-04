-- Down Migration

-- Drop Indexes
DROP INDEX IF EXISTS idx_entities_tenant;
DROP INDEX IF EXISTS idx_entities_parent;
DROP INDEX IF EXISTS idx_entities_type;
DROP INDEX IF EXISTS idx_hierarchy_paths_tenant;
DROP INDEX IF EXISTS idx_hierarchy_paths_ancestor;
DROP INDEX IF EXISTS idx_hierarchy_paths_descendant;
DROP INDEX IF EXISTS idx_hierarchy_paths_depth;

DROP INDEX IF EXISTS idx_persons_tenant;
DROP INDEX IF EXISTS idx_persons_type;
DROP INDEX IF EXISTS idx_persons_email;
DROP INDEX IF EXISTS idx_persons_name;

DROP INDEX IF EXISTS idx_employees_tenant;
DROP INDEX IF EXISTS idx_employees_entity;
DROP INDEX IF EXISTS idx_employees_number;
DROP INDEX IF EXISTS idx_employees_status;

DROP INDEX IF EXISTS idx_users_tenant;
DROP INDEX IF EXISTS idx_users_entity;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_type;

DROP INDEX IF EXISTS idx_projects_tenant;
DROP INDEX IF EXISTS idx_projects_entity;
DROP INDEX IF EXISTS idx_projects_status;

DROP INDEX IF EXISTS idx_budgets_tenant;
DROP INDEX IF EXISTS idx_budgets_entity;
DROP INDEX IF EXISTS idx_budgets_year;

DROP INDEX IF EXISTS idx_audit_logs_tenant;
DROP INDEX IF EXISTS idx_audit_logs_user;
DROP INDEX IF EXISTS idx_audit_logs_resource;
DROP INDEX IF EXISTS idx_audit_logs_created;

DROP INDEX IF EXISTS idx_chart_of_accounts_tenant;
DROP INDEX IF EXISTS idx_chart_of_accounts_entity;
DROP INDEX IF EXISTS idx_chart_of_accounts_code;
DROP INDEX IF EXISTS idx_chart_of_accounts_type;
DROP INDEX IF EXISTS idx_chart_of_accounts_parent;

-- Disable Row Level Security (RLS)
ALTER TABLE tenants DISABLE ROW LEVEL SECURITY;
ALTER TABLE entities DISABLE ROW LEVEL SECURITY;
ALTER TABLE hierarchy_paths DISABLE ROW LEVEL SECURITY;
ALTER TABLE persons DISABLE ROW LEVEL SECURITY;
ALTER TABLE employees DISABLE ROW LEVEL SECURITY;
ALTER TABLE users DISABLE ROW LEVEL SECURITY;
ALTER TABLE roles DISABLE ROW LEVEL SECURITY;
ALTER TABLE user_roles DISABLE ROW LEVEL SECURITY;
ALTER TABLE projects DISABLE ROW LEVEL SECURITY;
ALTER TABLE budgets DISABLE ROW LEVEL SECURITY;
ALTER TABLE audit_logs DISABLE ROW LEVEL SECURITY;
-- ALTER TABLE chart_of_accounts DISABLE ROW LEVEL SECURITY;
