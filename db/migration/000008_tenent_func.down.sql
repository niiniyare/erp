-- Down migration for the given SQL

-- Drop all RLS Policies
DROP POLICY IF EXISTS tenant_isolation_policy ON tenants;
DROP POLICY IF EXISTS tenant_isolation_policy ON entities;
DROP POLICY IF EXISTS tenant_isolation_policy ON hierarchy_paths;
DROP POLICY IF EXISTS tenant_isolation_policy ON persons;
DROP POLICY IF EXISTS tenant_isolation_policy ON employees;
DROP POLICY IF EXISTS tenant_isolation_policy ON users;
DROP POLICY IF EXISTS tenant_isolation_policy ON roles;
DROP POLICY IF EXISTS tenant_isolation_policy ON projects;
DROP POLICY IF EXISTS tenant_isolation_policy ON budgets;
DROP POLICY IF EXISTS tenant_isolation_policy ON audit_logs;
DROP POLICY IF EXISTS tenant_isolation_policy ON chart_of_accounts;
DROP POLICY IF EXISTS user_roles_policy ON user_roles;

-- Drop all Triggers
DROP TRIGGER IF EXISTS update_tenant_timestamps ON tenants;
DROP TRIGGER IF EXISTS update_entity_timestamps ON entities;
DROP TRIGGER IF EXISTS update_person_timestamps ON persons;
DROP TRIGGER IF EXISTS update_employee_timestamps ON employees;
DROP TRIGGER IF EXISTS update_user_timestamps ON users;
DROP TRIGGER IF EXISTS update_project_timestamps ON projects;
DROP TRIGGER IF EXISTS update_budget_timestamps ON budgets;
DROP TRIGGER IF EXISTS trg_maintain_hierarchy_paths ON entities;

-- Drop all Functions
DROP FUNCTION IF EXISTS current_tenant_id();
DROP FUNCTION IF EXISTS update_timestamps();
DROP FUNCTION IF EXISTS maintain_hierarchy_paths();
DROP FUNCTION IF EXISTS get_entity_descendants(INT, INT);
DROP FUNCTION IF EXISTS get_entity_ancestors(INT, INT);
DROP FUNCTION IF EXISTS is_entity_descendant(INT, INT);
DROP FUNCTION IF EXISTS get_entity_path(INT, VARCHAR);
DROP FUNCTION IF EXISTS get_entity_children(INT);
DROP FUNCTION IF EXISTS rebuild_hierarchy_paths(INT);

