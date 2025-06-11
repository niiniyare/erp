-- Person table (generic person entity)
CREATE TABLE persons (
    id SERIAL PRIMARY KEY,
    tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    person_type VARCHAR(20) NOT NULL 
        CHECK (person_type IN ('INDIVIDUAL', 'EMPLOYEE', 'CONTACT', 'CUSTOMER', 'VENDOR')),
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    middle_name VARCHAR(100),
--    full_name VARCHAR(255) GENERATED ALWAYS AS (
  --      TRIM(CONCAT(first_name, ' ', COALESCE(middle_name || ' ', ''), last_name))
   -- ) STORED,
    email VARCHAR(255),
    phone VARCHAR(20),
    birth_date DATE,
    national_id VARCHAR(50),
    tax_id VARCHAR(50),
    address JSONB, -- Flexible address structure
    metadata JSONB DEFAULT '{}'::jsonb, -- For industry-specific fields
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX person_email_unique_idx
ON persons (tenant_id, email)
WHERE email IS NOT NULL;

CREATE UNIQUE INDEX person_national_id_unique_idx
ON persons (tenant_id, national_id)
WHERE national_id IS NOT NULL;

-- Employee table (extends person)
CREATE TABLE employees (
    id SERIAL PRIMARY KEY,
    tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    person_id INT NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
    employee_number VARCHAR(50) NOT NULL,
    entity_id INT NOT NULL REFERENCES entities(id) ON DELETE RESTRICT,
    position_title VARCHAR(100),
    department_id INT REFERENCES entities(id), -- Direct department reference
    manager_id INT REFERENCES employees(id),
    hire_date DATE NOT NULL,
    termination_date DATE,
    salary_info JSONB, -- Encrypted/sensitive salary data
    employment_status VARCHAR(20) DEFAULT 'ACTIVE'
        CHECK (employment_status IN ('ACTIVE', 'INACTIVE', 'TERMINATED', 'ON_LEAVE')),
    work_schedule JSONB, -- Work hours, days, etc.
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, employee_number),
    UNIQUE (tenant_id, person_id)
);

-- Users table (system access)
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    person_id INT REFERENCES persons(id) ON DELETE SET NULL, -- May not always be linked to person
    employee_id INT REFERENCES employees(id) ON DELETE SET NULL, -- Employee users
    username VARCHAR(100),
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255),
    user_type VARCHAR(20) NOT NULL DEFAULT 'INTERNAL'
        CHECK (user_type IN ('INTERNAL', 'CUSTOMER', 'VENDOR', 'PARTNER', 'API')),
    entity_id INT REFERENCES entities(id) ON DELETE RESTRICT, -- User's primary entity
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_login_at TIMESTAMPTZ,
    password_changed_at TIMESTAMPTZ,
    settings JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (tenant_id, email)
);
CREATE UNIQUE INDEX user_username_unique_idx
ON users (tenant_id, username)
WHERE username IS NOT NULL;

-- Enhanced roles and permissions
CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    description TEXT,
    module VARCHAR(50), -- Which module this role belongs to
    permissions JSONB NOT NULL DEFAULT '{}'::jsonb,
    entity_scope JSONB, -- Which entity levels this role can access
    is_system_role BOOLEAN DEFAULT false, -- System vs custom roles
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, name)
);

CREATE TABLE user_roles (
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id INT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    entity_id INT REFERENCES entities(id) ON DELETE CASCADE, -- Role scope
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    assigned_by INT REFERENCES users(id),
    expires_at TIMESTAMPTZ,
    PRIMARY KEY (user_id, role_id, entity_id)
);


