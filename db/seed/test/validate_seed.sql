-- Validate seed data
\echo 'Running seed validation tests...'

SELECT 'Tenant count' AS test, COUNT(*) as result FROM tenants;
SELECT 'Expected: 3' AS note;

SELECT 'Entity count for ACME' AS test, COUNT(*) as result
FROM entities WHERE tenant_id = (SELECT id FROM tenants WHERE slug = 'acme-corp');
SELECT 'Expected: 3' AS note;

SELECT 'Person count' AS test, COUNT(*) as result FROM persons;
SELECT 'Expected: 2' AS note;

SELECT 'Employee count' AS test, COUNT(*) as result FROM employees;
SELECT 'Expected: 2' AS note;

SELECT 'User count' AS test, COUNT(*) as result FROM users;
SELECT 'Expected: 2' AS note;

SELECT 'Hierarchy depth' AS test, MAX(depth) as result FROM hierarchy_paths;
SELECT 'Expected: 2' AS note;

SELECT 'Hierarchy path count' AS test, COUNT(*) as result FROM hierarchy_paths;
SELECT 'Expected: 6' AS note;

\echo 'Seed validation completed.'
