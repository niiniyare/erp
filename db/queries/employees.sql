-- name: CreateEmployee :one
INSERT INTO employees (
    person_id, employee_number, entity_id, position_title, department_id, manager_id, hire_date, salary_info, employment_status, work_schedule, security_level, access_attributes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
) RETURNING *;

-- name: GetEmployeeByID :one
SELECT * FROM employees WHERE id = $1 AND deleted_at IS NULL;
