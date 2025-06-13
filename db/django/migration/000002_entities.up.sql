-- Root entity/company table with hierarchical structure and accounting preferences
CREATE TABLE entities (
    uuid UUID PRIMARY KEY,
    tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES entities(uuid) ON DELETE CASCADE,
    
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50), -- Internal reference code
    
    type VARCHAR(20) NOT NULL CHECK (
        type IN (
            'COMPANY', 'SUBSIDIARY', 'REGION', 'BRANCH', 'LOCATION',
            'DEPARTMENT', 'DIVISION', 'COST_CENTER', 'PROJECT', 'BUDGET_UNIT'
        )
    ),

    is_active BOOLEAN NOT NULL DEFAULT true,
    hidden BOOLEAN NOT NULL DEFAULT false,
    accrual_method BOOLEAN NOT NULL,                    -- TRUE = Accrual, FALSE = Cash
    fy_start_month INTEGER NOT NULL CHECK (fy_start_month BETWEEN 1 AND 12),

    address JSONB DEFAULT '{}'::jsonb,
    picture VARCHAR(100),
    settings JSONB DEFAULT '{}'::jsonb,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    UNIQUE (tenant_id, name)
);

COMMENT ON TABLE entities IS
'Purpose: Stores business entities/organizations/companies.
 Description: Core table representing different business entities that can
              have their own accounting books, customers, vendors, etc.
              Uses tree structure for hierarchical organization relationships.';

CREATE UNIQUE INDEX tenant_code_unique_idx
    ON entities (tenant_id, code)
    WHERE code IS NOT NULL;

-- Enhanced closure table for entity hierarchy
CREATE TABLE hierarchy_paths (
    tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    ancestor_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    descendant_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    depth INT NOT NULL CHECK (depth >= 0),

    PRIMARY KEY (tenant_id, ancestor_id, descendant_id)
);

-- Entity state tracking for sequence numbers and fiscal periods
CREATE TABLE IF NOT EXISTS entitystate (
    uuid UUID PRIMARY KEY,
    fiscal_year SMALLINT,
    key VARCHAR(10) NOT NULL,                             -- Document type (e.g., invoice, po)
    sequence BIGINT NOT NULL,                             -- Next sequence number
    entity_id UUID NOT NULL REFERENCES entities(uuid) DEFERRABLE INITIALLY DEFERRED,
    entity_unit_id UUID REFERENCES entities(uuid) DEFERRABLE INITIALLY DEFERRED
);

COMMENT ON TABLE entitystate IS
'Manages sequence numbers for document numbering (invoices, POs, etc.)';
COMMENT ON COLUMN entitystate.key IS
'Document type: invoice, po, estimate, bill, etc.';


-- -- Root entity/company table with hierarchical structure and accounting preferences
-- CREATE TABLE entities (
--     uuid UUID PRIMARY KEY,
--     tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
--     -- parent_id CHAR(32) REFERENCES entities(uuid) ON DELETE CASCADE,
--     name VARCHAR(255) NOT NULL,
--     code VARCHAR(50), -- For internal reference codes
--     type VARCHAR(20) NOT NULL 
--         CHECK (type IN ('COMPANY', 'SUBSIDIARY', 'REGION', 'BRANCH', 'LOCATION', 
--                        'DEPARTMENT', 'DIVISION', 'COST_CENTER', 'PROJECT', 'BUDGET_UNIT')),
--     is_active BOOLEAN NOT NULL DEFAULT true,
--     hidden BOOLEAN NOT NULL,                             -- Whether entity is hidden
--     accrual_method BOOLEAN NOT NULL,                     -- True=Accrual, False=Cash accounting
--     fy_start_month INTEGER NOT NULL,                     -- Fiscal year start month (1-12)
--   
--     address JSONB DEFAULT '{}'::jsonb, -- Entity-specific address
--     picture VARCHAR(100) NULL,                           -- Logo/picture file path
--     settings JSONB DEFAULT '{}'::jsonb, -- Entity-specific settings
--     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
--     updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
--     deleted_at TIMESTAMPTZ,
--     -- UNIQUE (tenant_id, parent_id, name)
--  UNIQUE (tenant_id, name)
-- );

-- COMMENT ON TABLE entities IS ' * Purpose: Stores business entities/organizations/companies
--  * Description: Core table representing different business entities that can
--  *              have their own accounting books, customers, vendors, etc.
--  *              Uses tree structure for hierarchical organization relationships.
-- ';

-- CREATE UNIQUE INDEX tenant_code_unique_idx
-- ON entities (tenant_id, code)
-- WHERE code IS NOT NULL;

-- -- Enhanced closure table (keep your existing structure)
-- CREATE TABLE hierarchy_paths (
--     tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
--     -- ancestor_id CHAR(32) NOT NULL REFERENCES  entities (uuid) ON DELETE CASCADE,
--     descendant_id CHAR(32) NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
--     depth INT NOT NULL,
--     -- PRIMARY KEY (tenant_id, ancestor_id, descendant_id)
--     PRIMARY KEY (tenant_id, descendant_id)

-- );


-- -- Entity state tracking for sequence numbers and fiscal periods
-- CREATE TABLE IF NOT EXISTS entitystate (
--   uuid UUID NOT NULL PRIMARY KEY,
--   fiscal_year SMALLINT NULL,                           -- Fiscal year for this state
--   key VARCHAR(10) NOT NULL,                            -- State key (e.g., invoice, po)
--   sequence BIGINT NOT NULL,                            -- Next sequence number for this key
--   entity__id CHAR(32) NOT NULL REFERENCES entities (uuid) DEFERRABLE INITIALLY DEFERRED,
--   entity_unit_id CHAR(32) NULL REFERENCES entities (uuid) DEFERRABLE INITIALLY DEFERRED
-- );
-- COMMENT ON TABLE entitystate IS 'Manages sequence numbers for document numbering (invoices, POs, etc.)';
-- COMMENT ON COLUMN entitystate.key IS 'Document type: invoice, po, estimate, bill, etc.';

