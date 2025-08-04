-- name: CreatePerson :one
INSERT INTO persons (
    entity_id, person_type, first_name, last_name, middle_name, email, phone, birth_date, national_id, tax_id, address, security_attributes, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
) RETURNING *;

-- name: GetPersonByID :one
SELECT * FROM persons WHERE id = $1 AND deleted_at IS NULL AND tenant_id = current_tenant_id();
