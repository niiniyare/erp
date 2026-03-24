-- ------------------------------------------------------------------------------------------------
-- PERSONS TABLE
-- ------------------------------------------------------------------------------------------------
-- Stores generic person entities that can represent individuals, employees, contacts, customers,
-- vendors, or contractors. Supports ABAC via security_attributes and flexible metadata storage.
-- person_type IN ('INDIVIDUAL','EMPLOYEE','CONTACT','CUSTOMER','VENDOR','CONTRACTOR').
--
-- NOTE: Depends on tenants(id) and entities(uuid) existing before this migration runs.
--       update_updated_at_column() trigger function must exist (created in an earlier migration).
-- ------------------------------------------------------------------------------------------------
CREATE TABLE persons (
  id                   UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id            UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id            UUID         NOT NULL REFERENCES entities(uuid) ON DELETE RESTRICT,
  person_type          VARCHAR(20)  NOT NULL CHECK (
                         person_type IN (
                           'INDIVIDUAL',
                           'EMPLOYEE',
                           'CONTACT',
                           'CUSTOMER',
                           'VENDOR',
                           'CONTRACTOR'
                         )
                       ),
  first_name           VARCHAR(100) NOT NULL,
  last_name            VARCHAR(100) NOT NULL,
  middle_name          VARCHAR(100),
  email                VARCHAR(255),
  phone                VARCHAR(20),
  birth_date           DATE,
  national_id          VARCHAR(50),
  tax_id               VARCHAR(50),
  address              JSONB        DEFAULT '{}'::jsonb,
  security_attributes  JSONB        DEFAULT '{}'::jsonb,  -- ABAC attributes: clearance level, department, location, etc.
  metadata             JSONB        DEFAULT '{}'::jsonb,  -- additional flexible data storage
  is_active            BOOLEAN      NOT NULL DEFAULT TRUE,
  -- validation columns
  version              INTEGER      NOT NULL DEFAULT 1,
  last_validation_run  TIMESTAMPTZ,
  validation_status    VARCHAR(20)  DEFAULT 'PENDING' CHECK (
                         validation_status IN ('PENDING', 'VALID', 'WARNING', 'ERROR')
                       ),
  validation_errors    JSONB        DEFAULT '[]'::jsonb,
  created_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  deleted_at           TIMESTAMPTZ,                                     -- soft delete timestamp; NULL = active
  -- email must be unique per tenant among non-deleted rows
  CONSTRAINT persons_email_unique_active EXCLUDE (tenant_id WITH =, email WITH =)
    WHERE (email IS NOT NULL AND deleted_at IS NULL),
  -- national ID must be unique per tenant among non-deleted rows
  CONSTRAINT persons_national_id_unique_active EXCLUDE (tenant_id WITH =, national_id WITH =)
    WHERE (national_id IS NOT NULL AND deleted_at IS NULL)
);

COMMENT ON TABLE  persons                      IS 'Stores person entities with ABAC security attributes. Supports multiple person types including employees, customers, vendors, and contractors. Implements soft delete and tenant isolation.';
COMMENT ON COLUMN persons.id                   IS 'UUID primary key for the person record.';
COMMENT ON COLUMN persons.tenant_id            IS 'Foreign key to tenants table for multi-tenant isolation.';
COMMENT ON COLUMN persons.entity_id            IS 'Foreign key to entities table for hierarchical organization.';
COMMENT ON COLUMN persons.person_type          IS 'Classification of person: INDIVIDUAL, EMPLOYEE, CONTACT, CUSTOMER, VENDOR, CONTRACTOR.';
COMMENT ON COLUMN persons.first_name           IS 'Person''s first/given name.';
COMMENT ON COLUMN persons.last_name            IS 'Person''s last/family name.';
COMMENT ON COLUMN persons.middle_name          IS 'Person''s middle name or initial (optional).';
COMMENT ON COLUMN persons.email                IS 'Person''s email address (must be unique per tenant when not deleted).';
COMMENT ON COLUMN persons.phone                IS 'Person''s primary phone number.';
COMMENT ON COLUMN persons.birth_date           IS 'Person''s date of birth.';
COMMENT ON COLUMN persons.national_id          IS 'Government-issued national ID number (unique per tenant).';
COMMENT ON COLUMN persons.tax_id               IS 'Tax identification number.';
COMMENT ON COLUMN persons.address              IS 'Person''s address stored as JSONB for flexible structure.';
COMMENT ON COLUMN persons.security_attributes  IS 'JSONB containing ABAC attributes like clearance level, department, location for access control.';
COMMENT ON COLUMN persons.metadata             IS 'Flexible JSONB storage for additional person-related data.';
COMMENT ON COLUMN persons.is_active            IS 'Whether the person record is currently active.';
COMMENT ON COLUMN persons.deleted_at           IS 'Soft delete timestamp — NULL means record is active.';

-- ------------------------------------------------------------------------------------------------
-- INDEXES
-- ------------------------------------------------------------------------------------------------
CREATE INDEX idx_persons_tenant                    ON persons(tenant_id);                      -- tenant-scoped queries (most common)
CREATE INDEX idx_persons_type                      ON persons(person_type);                    -- filter by person type
CREATE INDEX idx_persons_entity                    ON persons(entity_id);                      -- entity association lookups
CREATE INDEX idx_persons_active                    ON persons(is_active) WHERE is_active = TRUE;
CREATE INDEX idx_persons_deleted_at                ON persons(deleted_at) WHERE deleted_at IS NOT NULL;
CREATE INDEX idx_persons_email_lower               ON persons(lower(email)) WHERE email IS NOT NULL; -- case-insensitive email search
CREATE INDEX idx_persons_name                      ON persons(last_name, first_name);           -- name search
CREATE INDEX idx_persons_address_gin               ON persons USING gin(address);
CREATE INDEX idx_persons_security_attributes_gin   ON persons USING gin(security_attributes);
CREATE INDEX idx_persons_metadata_gin              ON persons USING gin(metadata);

-- ------------------------------------------------------------------------------------------------
-- ROW LEVEL SECURITY
-- ------------------------------------------------------------------------------------------------
ALTER TABLE persons ENABLE ROW LEVEL SECURITY;

CREATE POLICY persons_tenant_isolation ON persons FOR ALL TO application_role USING (
  current_tenant_id() IS NOT NULL
  AND tenant_id = current_tenant_id()
) WITH CHECK (
  current_tenant_id() IS NOT NULL
  AND tenant_id = current_tenant_id()
);

CREATE POLICY persons_admin_access ON persons FOR ALL TO admin_role USING (TRUE) WITH CHECK (TRUE);

-- ------------------------------------------------------------------------------------------------
-- TRIGGERS
-- ------------------------------------------------------------------------------------------------
CREATE TRIGGER update_persons_updated_at
  BEFORE UPDATE ON persons
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ------------------------------------------------------------------------------------------------
-- PERMISSIONS
-- ------------------------------------------------------------------------------------------------
GRANT SELECT, INSERT, UPDATE, DELETE ON persons TO application_role;
