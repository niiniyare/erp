-- Tenant context function
CREATE OR REPLACE FUNCTION current_tenant_id() RETURNS INT AS $$
BEGIN
    RETURN current_setting('app.current_tenant_id')::INT;
EXCEPTION WHEN OTHERS THEN
    RAISE EXCEPTION 'Tenant context not set';
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

-- RLS Policies (apply tenant isolation to all tables)
CREATE POLICY tenant_isolation_policy ON tenants USING (id = current_tenant_id());
CREATE POLICY tenant_isolation_policy ON entities USING (tenant_id = current_tenant_id());
CREATE POLICY tenant_isolation_policy ON hierarchy_paths USING (tenant_id = current_tenant_id());
CREATE POLICY tenant_isolation_policy ON persons USING (tenant_id = current_tenant_id());
CREATE POLICY tenant_isolation_policy ON employees USING (tenant_id = current_tenant_id());
CREATE POLICY tenant_isolation_policy ON users USING (tenant_id = current_tenant_id());
CREATE POLICY tenant_isolation_policy ON roles USING (tenant_id = current_tenant_id());
CREATE POLICY tenant_isolation_policy ON projects USING (tenant_id = current_tenant_id());
CREATE POLICY tenant_isolation_policy ON budgets USING (tenant_id = current_tenant_id());
CREATE POLICY tenant_isolation_policy ON audit_logs USING (tenant_id = current_tenant_id());
CREATE POLICY tenant_isolation_policy ON chart_of_accounts USING (tenant_id = current_tenant_id());

-- User roles policy (users can only see their own roles)
CREATE POLICY user_roles_policy ON user_roles
    USING (user_id IN (SELECT id FROM users WHERE tenant_id = current_tenant_id()));

-- Triggers for maintaining hierarchy paths and timestamps
CREATE OR REPLACE FUNCTION update_timestamps() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply timestamp triggers to all relevant tables
CREATE TRIGGER update_tenant_timestamps BEFORE UPDATE ON tenants FOR EACH ROW EXECUTE FUNCTION update_timestamps();
CREATE TRIGGER update_entity_timestamps BEFORE UPDATE ON entities FOR EACH ROW EXECUTE FUNCTION update_timestamps();
CREATE TRIGGER update_person_timestamps BEFORE UPDATE ON persons FOR EACH ROW EXECUTE FUNCTION update_timestamps();
CREATE TRIGGER update_employee_timestamps BEFORE UPDATE ON employees FOR EACH ROW EXECUTE FUNCTION update_timestamps();
CREATE TRIGGER update_user_timestamps BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_timestamps();
CREATE TRIGGER update_project_timestamps BEFORE UPDATE ON projects FOR EACH ROW EXECUTE FUNCTION update_timestamps();
CREATE TRIGGER update_budget_timestamps BEFORE UPDATE ON budgets FOR EACH ROW EXECUTE FUNCTION update_timestamps();

-- Enhanced hierarchy path maintenance (without ltree)
CREATE OR REPLACE FUNCTION maintain_hierarchy_paths() RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        -- Remove all paths involving this entity
        DELETE FROM hierarchy_paths
        WHERE tenant_id = OLD.tenant_id
        AND (ancestor_id = OLD.id OR descendant_id = OLD.id);
        RETURN OLD;
    END IF;

    -- Handle UPDATE of parent_id
    IF TG_OP = 'UPDATE' AND OLD.parent_id IS DISTINCT FROM NEW.parent_id THEN
        -- Remove old hierarchy paths for this entity
        DELETE FROM hierarchy_paths
        WHERE tenant_id = NEW.tenant_id AND descendant_id = NEW.id;
    END IF;

    -- Clear old paths for this descendant (for INSERT or parent change)
    IF TG_OP = 'INSERT' OR (TG_OP = 'UPDATE' AND OLD.parent_id IS DISTINCT FROM NEW.parent_id) THEN
        DELETE FROM hierarchy_paths
        WHERE tenant_id = NEW.tenant_id AND descendant_id = NEW.id;
    END IF;

    -- Self-reference (depth 0)
    INSERT INTO hierarchy_paths (tenant_id, ancestor_id, descendant_id, depth)
    VALUES (NEW.tenant_id, NEW.id, NEW.id, 0)
    ON CONFLICT DO NOTHING;

    -- Add parent paths (all ancestors)
    IF NEW.parent_id IS NOT NULL THEN
        INSERT INTO hierarchy_paths (tenant_id, ancestor_id, descendant_id, depth)
        SELECT NEW.tenant_id, p.ancestor_id, NEW.id, p.depth + 1
        FROM hierarchy_paths p
        WHERE p.tenant_id = NEW.tenant_id 
        AND p.descendant_id = NEW.parent_id
        ON CONFLICT DO NOTHING;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_maintain_hierarchy_paths
    BEFORE INSERT OR UPDATE OR DELETE ON entities
    FOR EACH ROW EXECUTE FUNCTION maintain_hierarchy_paths();

-- Function to get entity descendants
CREATE OR REPLACE FUNCTION get_entity_descendants(entity_id INT, max_depth INT DEFAULT NULL)
RETURNS TABLE(id INT, name VARCHAR, type VARCHAR, depth INT) AS $$
BEGIN
    RETURN QUERY
    SELECT e.id, e.name, e.type, hp.depth
    FROM entities e
    JOIN hierarchy_paths hp ON e.id = hp.descendant_id
    WHERE hp.tenant_id = current_tenant_id()
    AND hp.ancestor_id = entity_id
    AND (max_depth IS NULL OR hp.depth <= max_depth)
    ORDER BY hp.depth, e.name;
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

-- Function to get entity ancestors
CREATE OR REPLACE FUNCTION get_entity_ancestors(entity_id INT, max_depth INT DEFAULT NULL)
RETURNS TABLE(id INT, name VARCHAR, type VARCHAR, depth INT) AS $$
BEGIN
    RETURN QUERY
    SELECT e.id, e.name, e.type, hp.depth
    FROM entities e
    JOIN hierarchy_paths hp ON e.id = hp.ancestor_id
    WHERE hp.tenant_id = current_tenant_id()
    AND hp.descendant_id = entity_id
    AND hp.depth > 0  -- Exclude self-reference
    AND (max_depth IS NULL OR hp.depth <= max_depth)
    ORDER BY hp.depth DESC, e.name;
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

-- Function to check if entity is descendant of another
CREATE OR REPLACE FUNCTION is_entity_descendant(descendant_id INT, ancestor_id INT)
RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS(
        SELECT 1 FROM hierarchy_paths
        WHERE tenant_id = current_tenant_id()
        AND ancestor_id = is_entity_descendant.ancestor_id
        AND descendant_id = is_entity_descendant.descendant_id
        AND depth > 0
    );
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

-- Function to get entity path as text
CREATE OR REPLACE FUNCTION get_entity_path(entity_id INT, separator VARCHAR DEFAULT ' > ')
RETURNS TEXT AS $$
DECLARE
    result TEXT;
BEGIN
    SELECT string_agg(e.name, separator ORDER BY hp.depth DESC)
    INTO result
    FROM entities e
    JOIN hierarchy_paths hp ON e.id = hp.ancestor_id
    WHERE hp.tenant_id = current_tenant_id()
    AND hp.descendant_id = entity_id;
    
    RETURN COALESCE(result, '');
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

-- Function to get immediate children
CREATE OR REPLACE FUNCTION get_entity_children(entity_id INT)
RETURNS TABLE(id INT, name VARCHAR, type VARCHAR) AS $$
BEGIN
    RETURN QUERY
    SELECT e.id, e.name, e.type
    FROM entities e
    WHERE e.tenant_id = current_tenant_id()
    AND e.parent_id = entity_id
    ORDER BY e.name;
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

-- Function to rebuild hierarchy paths (useful for data repair)
CREATE OR REPLACE FUNCTION rebuild_hierarchy_paths(tenant_id_param INT DEFAULT NULL)
RETURNS VOID AS $$
DECLARE
    target_tenant_id INT;
BEGIN
    target_tenant_id := COALESCE(tenant_id_param, current_tenant_id());
    
    -- Clear existing paths for the tenant
    DELETE FROM hierarchy_paths WHERE tenant_id = target_tenant_id;
    
    -- Rebuild paths using recursive CTE
    WITH RECURSIVE entity_paths AS (
        -- Root entities (self-references)
        SELECT id, id as ancestor_id, 0 as depth, tenant_id
        FROM entities
        WHERE tenant_id = target_tenant_id
        
        UNION ALL
        
        -- Parent-child relationships
        SELECT e.id, ep.ancestor_id, ep.depth + 1, e.tenant_id
        FROM entities e
        JOIN entity_paths ep ON e.parent_id = ep.id
    )
    INSERT INTO hierarchy_paths (tenant_id, ancestor_id, descendant_id, depth)
    SELECT tenant_id, ancestor_id, id, depth
    FROM entity_paths;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
