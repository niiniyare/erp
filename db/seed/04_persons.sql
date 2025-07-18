-- Seed person data for ACME Corp
DO $$
DECLARE
    acme_tenant_id UUID;
    sales_dept_id UUID;
    person_id UUID;
    emp_counter INT := 1;
BEGIN
    -- Get ACME tenant and sales department IDs
    SELECT id INTO acme_tenant_id FROM tenants WHERE slug = 'acme-corp';
    SELECT uuid INTO sales_dept_id FROM entities 
    WHERE tenant_id = acme_tenant_id AND name = 'Sales Department';
    
    -- Insert persons
    FOR person_data IN 
        SELECT * FROM (VALUES 
            ('John', 'Smith', 'john.smith@acme.com'),
            ('Sarah', 'Johnson', 'sarah.j@acme.com')
        ) AS t(first_name, last_name, email)
    LOOP
        INSERT INTO persons (tenant_id, entity_id, person_type, first_name, last_name, email)
        VALUES (acme_tenant_id, sales_dept_id, 'EMPLOYEE', 
                person_data.first_name, person_data.last_name, person_data.email)
        ON CONFLICT (email) DO NOTHING
        RETURNING id INTO person_id;
        
        -- Get person ID if already exists
        IF person_id IS NULL THEN
            SELECT id INTO person_id FROM persons WHERE email = person_data.email;
        END IF;
        
        -- Insert employee record
        INSERT INTO employees (tenant_id, person_id, employee_number, position_title)
        VALUES (acme_tenant_id, person_id, 'EMP-' || LPAD(emp_counter::text, 5, '0'), 'Sales Representative')
        ON CONFLICT (tenant_id, employee_number) DO NOTHING;
        
        -- Insert user record
        INSERT INTO users (tenant_id, person_id, username, email, password_hash)
        VALUES (acme_tenant_id, person_id, 
                SPLIT_PART(person_data.email, '@', 1), 
                person_data.email,
                crypt('Password123!', gen_salt('bf')))
        ON CONFLICT (email) DO NOTHING;
        
        emp_counter := emp_counter + 1;
    END LOOP;
END $$;
