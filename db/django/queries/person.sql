-- ==============================================
-- PERSONS TABLE OPERATIONS
-- ==============================================

-- name: CreatePerson :one
INSERT INTO persons (
    tenant_id, entity_id, person_type, first_name, last_name, middle_name,
    email, phone, birth_date, national_id, tax_id, address, metadata, is_active
) VALUES (
    current_tenant_id(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
) RETURNING *;

-- name: GetPerson :one
SELECT * FROM persons 
WHERE id = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL;

-- name: GetPersonByEmail :one
SELECT * FROM persons 
WHERE email = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL;

-- name: GetPersonByNationalId :one
SELECT * FROM persons 
WHERE national_id = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL;

-- name: UpdatePerson :one
UPDATE persons 
SET 
    entity_id = COALESCE($2, entity_id),
    person_type = COALESCE($3, person_type),
    first_name = COALESCE($4, first_name),
    last_name = COALESCE($5, last_name),
    middle_name = COALESCE($6, middle_name),
    email = COALESCE($7, email),
    phone = COALESCE($8, phone),
    birth_date = COALESCE($9, birth_date),
    national_id = COALESCE($10, national_id),
    tax_id = COALESCE($11, tax_id),
    address = COALESCE($12, address),
    metadata = COALESCE($13, metadata),
    is_active = COALESCE($14, is_active),
    updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeletePerson :exec
UPDATE persons 
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: RestorePerson :exec
UPDATE persons 
SET deleted_at = NULL, updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: ListPersons :many
SELECT * FROM persons 
WHERE tenant_id = current_tenant_id() AND deleted_at IS NULL
ORDER BY last_name, first_name;

-- name: ListPersonsByType :many
SELECT * FROM persons 
WHERE tenant_id = current_tenant_id() AND person_type = $1 AND deleted_at IS NULL
ORDER BY last_name, first_name;

-- name: ListPersonsByEntity :many
SELECT * FROM persons 
WHERE tenant_id = current_tenant_id() AND entity_id = $1 AND deleted_at IS NULL
ORDER BY last_name, first_name;

-- name: SearchPersonsByName :many
SELECT * FROM persons 
WHERE tenant_id = current_tenant_id() 
    AND (
        first_name ILIKE '%' || $1 || '%' OR 
        last_name ILIKE '%' || $1 || '%' OR
        (first_name || ' ' || last_name) ILIKE '%' || $1 || '%'
    )
    AND deleted_at IS NULL
ORDER BY last_name, first_name
LIMIT $2;

-- name: GetPersonFullName :one
SELECT 
    CASE 
        WHEN middle_name IS NOT NULL AND middle_name != '' 
        THEN first_name || ' ' || middle_name || ' ' || last_name
        ELSE first_name || ' ' || last_name
    END as full_name
FROM persons p
WHERE p.id = sqlc.arg('person_id') AND p.tenant_id = current_tenant_id() AND p.deleted_at IS NULL;

-- ==============================================
-- COMPLEX QUERIES AND REPORTS
-- ==============================================

-- name: GetPersonEmployeeUserInfo :one
SELECT 
    p.*,
    e.id as employee_id, e.employee_number, e.position_title, e.employment_status,
    u.id as user_id, u.username, u.email as user_email, u.user_type, u.is_active as user_active,
    (CASE 
        WHEN p.middle_name IS NOT NULL AND p.middle_name != '' 
        THEN p.first_name || ' ' || p.middle_name || ' ' || p.last_name
        ELSE p.first_name || ' ' || p.last_name
    END)::text as full_name
FROM persons p
LEFT JOIN employees e ON p.id = e.person_id AND e.tenant_id = current_tenant_id() AND e.deleted_at IS NULL
LEFT JOIN users u ON p.id = u.person_id AND u.tenant_id = current_tenant_id() AND u.deleted_at IS NULL
WHERE p.id = sqlc.arg('person_id') AND p.tenant_id = current_tenant_id() AND p.deleted_at IS NULL;

