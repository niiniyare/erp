CREATE TABLE IF NOT EXISTS org (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    hierarchy_level INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS entities (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type ENTITY_TYPE NOT NULL,
    parent_id INT REFERENCES entities(id) ON DELETE CASCADE,
    org_id INT NOT NULL REFERENCES org(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL
);

CREATE TYPE ENTITY_TYPE AS ENUM (
    'COMPANY',
    'REGIONAL',
    'DEPARTMENT',
    'COST_CENTER'
);

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    full_name VARCHAR(255) NOT NULL,
    entity_id INT NOT NULL REFERENCES entities(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMPTZ DEFAULT NULL,
    deleted_at TIMESTAMPTZ DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS hierarchy_paths (
    ancestor_id INT NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    descendant_id INT NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    depth INT NOT NULL,
    PRIMARY KEY (ancestor_id, descendant_id)
);

-- Indexes
CREATE INDEX idx_entities_parent ON entities(parent_id);
CREATE INDEX idx_entities_org ON entities(org_id);
CREATE INDEX idx_hierarchy_ancestor ON hierarchy_paths(ancestor_id);
CREATE INDEX idx_hierarchy_descendant ON hierarchy_paths(descendant_id);
CREATE INDEX idx_users_entity ON users(entity_id);

-- Path enumeration trigger
CREATE OR REPLACE FUNCTION update_entity_paths() RETURNS TRIGGER AS $$
BEGIN
    -- Self-reference
    INSERT INTO hierarchy_paths (ancestor_id, descendant_id, depth)
    VALUES (NEW.id, NEW.id, 0);
    
    -- Parent paths
    IF NEW.parent_id IS NOT NULL THEN
        INSERT INTO hierarchy_paths (ancestor_id, descendant_id, depth)
        SELECT p.ancestor_id, NEW.id, p.depth + 1
        FROM hierarchy_paths p
        WHERE p.descendant_id = NEW.parent_id;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_entity_paths
AFTER INSERT ON entities
FOR EACH ROW EXECUTE FUNCTION update_entity_paths();

-- Timestamp update trigger
CREATE OR REPLACE FUNCTION update_timestamps() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_update_org_timestamp
BEFORE UPDATE ON org
FOR EACH ROW EXECUTE FUNCTION update_timestamps();

CREATE TRIGGER trg_update_entity_timestamp
BEFORE UPDATE ON entities
FOR EACH ROW EXECUTE FUNCTION update_timestamps();

CREATE TRIGGER trg_update_user_timestamp
BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION update_timestamps();


--- views

-- 1. Fixed Full orgal Hierarchy View
CREATE VIEW v_org_hierarchy AS
WITH RECURSIVE org_chart AS (
  SELECT 
    id, 
    name, 
    type, 
    parent_id, 
    org_id,
    name::TEXT AS path,  -- CAST to TEXT to resolve type conflict
    0 AS depth
  FROM entities
  WHERE parent_id IS NULL
  
  UNION ALL
  
  SELECT 
    e.id, 
    e.name, 
    e.type, 
    e.parent_id, 
    e.org_id,
    (oc.path || ' > ' || e.name)::TEXT,  -- Ensure consistent TEXT type
    oc.depth + 1
  FROM entities e
  JOIN org_chart oc ON e.parent_id = oc.id
)
SELECT 
  o.name AS org,
  oc.id AS entity_id,
  oc.name AS entity_name,
  oc.type AS entity_type,
  oc.path AS full_path,
  oc.depth
FROM org_chart oc
JOIN org o ON oc.org_id = o.id  -- Fixed table name
WHERE oc.deleted_at IS NULL;

-- 2. Fixed User Directory View
CREATE VIEW v_user_directory AS
SELECT
  u.id AS user_id,
  u.email,
  u.full_name,
  u.last_login_at,
  cc.id AS cost_center_id,
  cc.name AS cost_center,
  d.id AS department_id,
  d.name AS department,
  r.id AS regional_id,
  r.name AS regional,
  c.id AS company_id,
  c.name AS company,
  o.id AS org_id,
  o.name AS org
FROM users u
JOIN entities cc ON u.entity_id = cc.id
JOIN entities d ON cc.parent_id = d.id
JOIN entities r ON d.parent_id = r.id
JOIN entities c ON r.parent_id = c.id
JOIN org o ON c.org_id = o.id  -- Fixed table name
WHERE u.deleted_at IS NULL;

-- 3. Fixed Cost Center Management View
CREATE VIEW v_cost_center_management AS
SELECT
  cc.id AS cost_center_id,
  cc.name AS cost_center,
  d.id AS department_id,
  d.name AS department,
  r.id AS regional_id,
  r.name AS regional,
  c.id AS company_id,
  c.name AS company,
  o.id AS org_id,
  o.name AS org,
  COUNT(u.id) AS user_count,
  STRING_AGG(u.full_name, ', ') AS users
FROM entities cc
JOIN entities d ON cc.parent_id = d.id
JOIN entities r ON d.parent_id = r.id
JOIN entities c ON r.parent_id = c.id
JOIN org o ON c.org_id = o.id  -- Fixed table name
LEFT JOIN users u ON u.entity_id = cc.id
WHERE cc.type = 'COST_CENTER'
GROUP BY 
  cc.id, cc.name, 
  d.id, d.name,
  r.id, r.name,
  c.id, c.name,
  o.id, o.name;

-- 4. Fixed Department Summary View
CREATE VIEW v_department_summary AS
SELECT
  d.id AS department_id,
  d.name AS department,
  r.id AS regional_id,
  r.name AS regional,
  c.id AS company_id,
  c.name AS company,
  o.id AS org_id,
  o.name AS org,
  COUNT(DISTINCT cc.id) AS cost_center_count,
  COUNT(DISTINCT u.id) AS user_count
FROM entities d
JOIN entities r ON d.parent_id = r.id
JOIN entities c ON r.parent_id = c.id
JOIN org o ON c.org_id = o.id  -- Fixed table name
LEFT JOIN entities cc ON cc.parent_id = d.id AND cc.type = 'COST_CENTER'
LEFT JOIN users u ON u.entity_id = cc.id
WHERE d.type = 'DEPARTMENT'
GROUP BY 
  d.id, d.name,
  r.id, r.name,
  c.id, c.name,
  o.id, o.name;

-- 5. Company Subtree View (No changes needed)
CREATE VIEW v_company_subtree AS
SELECT
  c.id AS company_id,
  c.name AS company,
  e.id AS entity_id,
  e.name AS entity_name,
  e.type AS entity_type,
  hp.depth AS levels_from_company
FROM entities c
JOIN hierarchy_paths hp ON c.id = hp.ancestor_id
JOIN entities e ON hp.descendant_id = e.id
WHERE c.type = 'COMPANY';

-- 6. Fixed org Size Report View
CREATE VIEW v_org_size_report AS
SELECT
  o.id AS org_id,
  o.name AS org,
  COUNT(DISTINCT c.id) FILTER (WHERE c.type = 'COMPANY') AS company_count,
  COUNT(DISTINCT r.id) FILTER (WHERE r.type = 'REGIONAL') AS regional_count,
  COUNT(DISTINCT d.id) FILTER (WHERE d.type = 'DEPARTMENT') AS department_count,
  COUNT(DISTINCT cc.id) FILTER (WHERE cc.type = 'COST_CENTER') AS cost_center_count,
  COUNT(DISTINCT u.id) AS user_count
FROM org o  -- Fixed table name
LEFT JOIN entities c ON c.org_id = o.id
LEFT JOIN entities r ON r.org_id = o.id
LEFT JOIN entities d ON d.org_id = o.id
LEFT JOIN entities cc ON cc.org_id = o.id
LEFT JOIN users u ON u.entity_id = cc.id
GROUP BY o.id, o.name;

-- 7. Fixed Active User Access Report
CREATE VIEW v_active_users AS
SELECT
  u.id AS user_id,
  u.email,
  u.full_name,
  o.name AS org,
  c.name AS company,
  r.name AS regional,
  d.name AS department,
  cc.name AS cost_center,
  u.last_login_at
FROM users u
JOIN entities cc ON u.entity_id = cc.id
JOIN entities d ON cc.parent_id = d.id
JOIN entities r ON d.parent_id = r.id
JOIN entities c ON r.parent_id = c.id
JOIN org o ON c.org_id = o.id  -- Fixed table name
WHERE u.deleted_at IS NULL
AND u.last_login_at > CURRENT_DATE - INTERVAL '90 days';

-- Recommended Indexes for View Performance
CREATE INDEX idx_entities_org_parent ON entities(org_id, parent_id);
CREATE INDEX idx_users_deleted ON users(deleted_at);
CREATE INDEX idx_entities_type ON entities(type);
CREATE INDEX idx_hierarchy_paths_depth ON hierarchy_paths(depth);
CREATE INDEX idx_users_last_login ON users(last_login_at);
