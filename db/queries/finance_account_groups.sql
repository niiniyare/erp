-- name: CreateAccountGroup :one
INSERT INTO
  finance_account_groups (
    tenant_id,
    entity_id,
    group_code,
    group_name,
    group_description,
    parent_group_id,
    group_level,
    group_path,
    root_type,
    group_category,
    financial_statement_section,
    created_by
  )
VALUES
  (
    current_tenant_id(),
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10,
    $11
  )
RETURNING
  *;

-- name: GetAccountGroup :one
SELECT
  *
FROM
  finance_account_groups
WHERE
  id = $1
  AND tenant_id = current_tenant_id();

-- name: GetAccountGroupByCode :one
SELECT
  *
FROM
  finance_account_groups
WHERE
  group_code = $1
  AND tenant_id = current_tenant_id();

-- name: UpdateAccountGroup :one
UPDATE
  finance_account_groups
SET
  group_name = COALESCE($2, group_name),
  group_description = COALESCE($3, group_description),
  updated_at = NOW(),
  updated_by = $4,
  version = version + 1
WHERE
  id = $1
  AND tenant_id = current_tenant_id()
RETURNING
  *;

-- name: DeleteAccountGroup :exec
DELETE FROM
  finance_account_groups
WHERE
  id = $1
  AND tenant_id = current_tenant_id();

-- name: ListAccountGroups :many
SELECT
  *
FROM
  finance_account_groups
WHERE
  tenant_id = current_tenant_id()
  AND (
    $1::uuid IS NULL
    OR entity_id = $1
  )
  AND (
    $2::text IS NULL
    OR root_type = $2
  )
  AND (
    $3::text IS NULL
    OR group_category = $3
  )
ORDER BY
  group_level,
  group_name
LIMIT
  $4 OFFSET $5;

-- name: CountAccountGroups :one
SELECT
  count(*)
FROM
  finance_account_groups
WHERE
  tenant_id = current_tenant_id()
  AND (
    $1::uuid IS NULL
    OR entity_id = $1
  )
  AND (
    $2::text IS NULL
    OR root_type = $2
  )
  AND (
    $3::text IS NULL
    OR group_category = $3
  );

-- name: GetAccountGroupsByRootType :many
SELECT
  *
FROM
  finance_account_groups
WHERE
  tenant_id = current_tenant_id()
  AND (
    $1::uuid IS NULL
    OR entity_id = $1
  )
  AND root_type = $2
ORDER BY
  group_level,
  group_name;

-- name: GetAccountGroupHierarchy :many
SELECT
  *
FROM
  finance_account_groups
WHERE
  tenant_id = current_tenant_id()
  AND (
    $1::uuid IS NULL
    OR entity_id = $1
  )
  AND (
    $2::uuid IS NULL
    OR id = $2
    OR group_path LIKE '%/' || $2::text || '/%'
  )
ORDER BY
  group_level,
  group_name;

-- name: GetAccountGroupChildren :many
SELECT
  *
FROM
  finance_account_groups
WHERE
  tenant_id = current_tenant_id()
  AND (
    $1::uuid IS NULL
    OR entity_id = $1
  )
  AND parent_group_id = $2
ORDER BY
  group_name;

-- name: GetGroupsByStatementSection :many
SELECT
  *
FROM
  finance_account_groups
WHERE
  tenant_id = current_tenant_id()
  AND (
    $1::uuid IS NULL
    OR entity_id = $1
  )
  AND financial_statement_section = $2
ORDER BY
  statement_order,
  group_name;

-- name: GetGroupsByCategory :many
SELECT
  *
FROM
  finance_account_groups
WHERE
  tenant_id = current_tenant_id()
  AND (
    $1::uuid IS NULL
    OR entity_id = $1
  )
  AND group_category = $2
ORDER BY
  group_name;

-- name: SearchAccountGroups :many
SELECT
  *
FROM
  finance_account_groups
WHERE
  tenant_id = current_tenant_id()
  AND (
    $1::uuid IS NULL
    OR entity_id = $1
  )
  AND (
    group_code ILIKE '%' || $2 || '%'
    OR group_name ILIKE '%' || $2 || '%'
    OR group_description ILIKE '%' || $2 || '%'
  )
ORDER BY
  group_name
LIMIT
  $3 OFFSET $4;

-- name: UpdateGroupHierarchyPath :one
UPDATE
  finance_account_groups
SET
  group_path = $2,
  updated_at = NOW(),
  version = version + 1
WHERE
  id = $1
  AND tenant_id = current_tenant_id()
RETURNING
  *;
