-- name: CreateAccountGroup :one
-- Create a new account group with proper entity isolation
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
    consolidation_method,
    cash_flow_category,
    statement_order,
    display_format,
    indent_level,
    show_totals,
    bold_display,
    is_active,
    created_by
  )
VALUES
  (
    current_tenant_id(),
    sqlc.arg('entity_id'),
    sqlc.arg('group_code'),
    sqlc.arg('group_name'),
    sqlc.narg('group_description'),
    sqlc.narg('parent_group_id'),
    sqlc.arg('group_level'),
    sqlc.arg('group_path'),
    sqlc.arg('root_type'),
    sqlc.narg('group_category'),
    sqlc.narg('financial_statement_section'),
    sqlc.narg('consolidation_method'),
    sqlc.narg('cash_flow_category'),
    sqlc.arg('statement_order'),
    sqlc.arg('display_format'),
    sqlc.arg('indent_level'),
    sqlc.arg('show_totals'),
    sqlc.arg('bold_display'),
    sqlc.arg('is_active'),
    sqlc.arg('created_by')
  )
RETURNING
  *;

-- name: GetAccountGroup :one
-- Get account group by ID with proper tenant/entity isolation
SELECT
  *
FROM
  finance_account_groups
WHERE
  id = sqlc.arg('group_id')
  AND tenant_id = current_tenant_id()
  AND (sqlc.narg('entity_id')::uuid IS NULL OR entity_id = sqlc.arg('entity_id'))
  AND deleted_at IS NULL;

-- name: GetAccountGroupByCode :one
-- Get account group by code with proper tenant/entity isolation
SELECT
  *
FROM
  finance_account_groups
WHERE
  group_code = sqlc.arg('group_code')
  AND tenant_id = current_tenant_id()
  AND (sqlc.narg('entity_id')::uuid IS NULL OR entity_id = sqlc.arg('entity_id'))
  AND deleted_at IS NULL;

-- name: UpdateAccountGroup :one
-- Update account group with proper tenant/entity isolation
UPDATE
  finance_account_groups
SET
  group_name = COALESCE(sqlc.narg('group_name'), group_name),
  group_description = COALESCE(sqlc.narg('group_description'), group_description),
  parent_group_id = COALESCE(sqlc.narg('parent_group_id'), parent_group_id),
  financial_statement_section = COALESCE(sqlc.narg('financial_statement_section'), financial_statement_section),
  consolidation_method = COALESCE(sqlc.narg('consolidation_method'), consolidation_method),
  cash_flow_category = COALESCE(sqlc.narg('cash_flow_category'), cash_flow_category),
  statement_order = COALESCE(sqlc.narg('statement_order'), statement_order),
  indent_level = COALESCE(sqlc.narg('indent_level'), indent_level),
  show_totals = COALESCE(sqlc.narg('show_totals'), show_totals),
  bold_display = COALESCE(sqlc.narg('bold_display'), bold_display),
  is_active = COALESCE(sqlc.narg('is_active'), is_active),
  updated_at = NOW(),
  updated_by = sqlc.arg('updated_by')
WHERE
  id = sqlc.arg('group_id')
  AND tenant_id = current_tenant_id()
  AND (sqlc.narg('entity_id')::uuid IS NULL OR entity_id = sqlc.arg('entity_id'))
  AND deleted_at IS NULL
RETURNING
  *;

-- name: SoftDeleteAccountGroup :exec
-- Soft delete account group with proper tenant/entity isolation
UPDATE
  finance_account_groups
SET
  deleted_at = sqlc.arg('deleted_at'),
  updated_at = sqlc.arg('deleted_at'),
  updated_by = sqlc.arg('deleted_by')
WHERE
  id = sqlc.arg('group_id')
  AND tenant_id = current_tenant_id()
  AND (sqlc.narg('entity_id')::uuid IS NULL OR entity_id = sqlc.arg('entity_id'))
  AND deleted_at IS NULL;

-- name: ListAccountGroups :many
-- List account groups with filtering and pagination
SELECT
  *
FROM
  finance_account_groups
WHERE
  tenant_id = current_tenant_id()
  AND (sqlc.narg('entity_id')::uuid IS NULL OR entity_id = sqlc.arg('entity_id'))
  AND (sqlc.narg('root_type')::text IS NULL OR root_type = sqlc.arg('root_type'))
  AND (sqlc.narg('group_category')::text IS NULL OR group_category = sqlc.arg('group_category'))
  AND (sqlc.narg('parent_group_id')::uuid IS NULL OR parent_group_id = sqlc.arg('parent_group_id'))
  AND (sqlc.narg('is_active')::boolean IS NULL OR is_active = sqlc.arg('is_active'))
  AND deleted_at IS NULL
ORDER BY
  CASE WHEN sqlc.narg('sort_by')::text = 'group_level' THEN group_level END,
  CASE WHEN sqlc.narg('sort_by')::text = 'group_name' THEN group_name END,
  CASE WHEN sqlc.narg('sort_by')::text = 'statement_order' THEN statement_order END,
  group_level,
  group_name
LIMIT
  sqlc.arg('limit_count') OFFSET sqlc.arg('offset_count');

-- name: CountAccountGroups :one
-- Count account groups with filtering
SELECT
  count(*)
FROM
  finance_account_groups
WHERE
  tenant_id = current_tenant_id()
  AND (sqlc.narg('entity_id')::uuid IS NULL OR entity_id = sqlc.arg('entity_id'))
  AND (sqlc.narg('root_type')::text IS NULL OR root_type = sqlc.arg('root_type'))
  AND (sqlc.narg('group_category')::text IS NULL OR group_category = sqlc.arg('group_category'))
  AND (sqlc.narg('parent_group_id')::uuid IS NULL OR parent_group_id = sqlc.arg('parent_group_id'))
  AND (sqlc.narg('is_active')::boolean IS NULL OR is_active = sqlc.arg('is_active'))
  AND deleted_at IS NULL;

-- name: GetAccountGroupsByRootType :many
-- Get account groups by root type with entity isolation
SELECT
  *
FROM
  finance_account_groups
WHERE
  tenant_id = current_tenant_id()
  AND (sqlc.narg('entity_id')::uuid IS NULL OR entity_id = sqlc.arg('entity_id'))
  AND root_type = sqlc.arg('root_type')
  AND deleted_at IS NULL
ORDER BY
  group_level,
  statement_order,
  group_name;

-- name: GetAccountGroupHierarchy :many
-- Get account group hierarchy with optional root filtering
SELECT
  *
FROM
  finance_account_groups
WHERE
  tenant_id = current_tenant_id()
  AND (sqlc.narg('entity_id')::uuid IS NULL OR entity_id = sqlc.arg('entity_id'))
  AND (
    sqlc.narg('root_group_id')::uuid IS NULL
    OR id = sqlc.arg('root_group_id')
    OR group_path LIKE '%/' || sqlc.arg('root_group_id')::text || '/%'
  )
  AND deleted_at IS NULL
ORDER BY
  group_level,
  statement_order,
  group_name;

-- name: GetAccountGroupChildren :many
-- Get direct children of an account group
SELECT
  *
FROM
  finance_account_groups
WHERE
  tenant_id = current_tenant_id()
  AND (sqlc.narg('entity_id')::uuid IS NULL OR entity_id = sqlc.arg('entity_id'))
  AND parent_group_id = sqlc.arg('parent_group_id')
  AND deleted_at IS NULL
ORDER BY
  statement_order,
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

-- name: ValidateAccountGroupCode :one
-- Validate if account group code is unique within entity/tenant
SELECT
  EXISTS(
    SELECT 1
    FROM finance_account_groups
    WHERE tenant_id = current_tenant_id()
      AND (sqlc.narg('entity_id')::uuid IS NULL OR entity_id = sqlc.arg('entity_id'))
      AND group_code = sqlc.arg('group_code')
      AND (sqlc.narg('exclude_id')::uuid IS NULL OR id != sqlc.arg('exclude_id'))
      AND deleted_at IS NULL
  ) as code_exists;


-- name: CheckGroupHasChildren :one
-- Check if account group has child groups
SELECT
  EXISTS(
    SELECT 1
    FROM finance_account_groups
    WHERE parent_group_id = sqlc.arg('group_id')
      AND tenant_id = current_tenant_id()
      AND deleted_at IS NULL
  ) as has_children;
