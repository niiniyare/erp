-- Down migration

-- Drop user_roles table
DROP TABLE IF EXISTS user_roles;

-- Drop roles table
DROP TABLE IF EXISTS roles;

-- Drop users table and its associated index
DROP INDEX IF EXISTS user_username_unique_idx;
DROP TABLE IF EXISTS users;

-- Drop employees table
DROP TABLE IF EXISTS employees;

-- Drop persons table and its associated indexes
DROP INDEX IF EXISTS person_email_unique_idx;
DROP INDEX IF EXISTS person_national_id_unique_idx;
DROP TABLE IF EXISTS persons;
