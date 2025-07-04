-- ==============================================
-- EMPLOYEES TABLE OPERATIONS
-- ==============================================

-- name: CreateEmployee :one
INSERT INTO employees (
    tenant_id, person_id, employee_number, entity_id, position_title,
    department_id, manager_id, hire_date, termination_date, salary_info,
    employment_status, work_schedule
) VALUES (
    current_tenant_id(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;

-- name: GetEmployee :one
SELECT * FROM employees 
WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: GetEmployeeByNumber :one
SELECT * FROM employees 
WHERE employee_number = $1 AND tenant_id = current_tenant_id();

-- name: GetEmployeeByPersonId :one
SELECT * FROM employees 
WHERE person_id = $1 AND tenant_id = current_tenant_id();

-- name: UpdateEmployee :one
UPDATE employees 
SET 
    employee_number = COALESCE($2, employee_number),
    entity_id = COALESCE($3, entity_id),
    position_title = COALESCE($4, position_title),
    department_id = COALESCE($5, department_id),
    manager_id = COALESCE($6, manager_id),
    hire_date = COALESCE($7, hire_date),
    termination_date = COALESCE($8, termination_date),
    salary_info = COALESCE($9, salary_info),
    employment_status = COALESCE($10, employment_status),
    work_schedule = COALESCE($11, work_schedule),
    updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id()
RETURNING *;

-- name: ListEmployees :many
SELECT e.*, p.first_name, p.last_name, p.email, p.phone
FROM employees e
JOIN persons p ON e.person_id = p.id
WHERE e.tenant_id = current_tenant_id()
ORDER BY p.last_name, p.first_name;

-- name: ListActiveEmployees :many
SELECT e.*, p.first_name, p.last_name, p.email, p.phone
FROM employees e
JOIN persons p ON e.person_id = p.id
WHERE e.tenant_id = current_tenant_id() AND e.employment_status = 'ACTIVE'
ORDER BY p.last_name, p.first_name;

-- name: ListEmployeesByDepartment :many
SELECT e.*, p.first_name, p.last_name, p.email, p.phone
FROM employees e
JOIN persons p ON e.person_id = p.id
WHERE e.tenant_id = current_tenant_id() AND e.department_id = $1
ORDER BY p.last_name, p.first_name;

-- name: ListEmployeesByManager :many
SELECT e.*, p.first_name, p.last_name, p.email, p.phone
FROM employees e
JOIN persons p ON e.person_id = p.id
WHERE e.tenant_id = current_tenant_id() AND e.manager_id = $1
ORDER BY p.last_name, p.first_name;

-- name: GetEmployeeHierarchy :many
WITH RECURSIVE employee_hierarchy AS (
    SELECT 
        e.id, e.person_id, e.employee_number, e.position_title, e.manager_id,
        p.first_name, p.last_name, 0 as level
    FROM employees e
    JOIN persons p ON e.person_id = p.id
    WHERE e.id = $1 AND e.tenant_id = current_tenant_id()
    
    UNION ALL
    
    SELECT 
        e.id, e.person_id, e.employee_number, e.position_title, e.manager_id,
        p.first_name, p.last_name, eh.level + 1
    FROM employees e
    JOIN persons p ON e.person_id = p.id
    JOIN employee_hierarchy eh ON e.manager_id = eh.id
    WHERE e.tenant_id = current_tenant_id() AND eh.level < 10
)
SELECT * FROM employee_hierarchy
ORDER BY level, last_name, first_name;


-- name: GetEmployeeStats :one
SELECT 
    COUNT(*) as total_employees,
    COUNT(*) FILTER (WHERE employment_status = 'ACTIVE') as active_employees,
    COUNT(*) FILTER (WHERE employment_status = 'TERMINATED') as terminated_employees,
    COUNT(*) FILTER (WHERE employment_status = 'ON_LEAVE') as on_leave_employees,
    COUNT(DISTINCT department_id) as departments_count,
    COUNT(*) FILTER (WHERE manager_id IS NULL) as managers_count
FROM employees
WHERE tenant_id = current_tenant_id();

