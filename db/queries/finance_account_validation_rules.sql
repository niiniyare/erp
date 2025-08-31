-- name: CreateAccountValidationRule :one
INSERT INTO
  finance_account_validation_rules (
    tenant_id,
    rule_name,
    rule_description,
    account_type,
    root_type,
    validation_parameters,
    rule_severity,
    is_active,
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
    $8
  )
RETURNING
  *;

-- name: GetAccountValidationRule :one
SELECT
  *
FROM
  finance_account_validation_rules
WHERE
  id = $1
  AND tenant_id = current_tenant_id();

-- name: UpdateAccountValidationRule :one
UPDATE
  finance_account_validation_rules
SET
  rule_name = COALESCE($2, rule_name),
  rule_description = COALESCE($3, rule_description),
  validation_parameters = COALESCE($4, validation_parameters),
  rule_severity = COALESCE($5, rule_severity),
  is_active = COALESCE($6, is_active),
  updated_at = NOW(),
  updated_by = $7
WHERE
  id = $1
  AND tenant_id = current_tenant_id()
RETURNING
  *;

-- name: DeleteAccountValidationRule :exec
DELETE FROM
  finance_account_validation_rules
WHERE
  id = $1
  AND tenant_id = current_tenant_id();

-- name: ListAccountValidationRules :many
SELECT
  *
FROM
  finance_account_validation_rules
WHERE
  tenant_id = current_tenant_id()
  AND (
    $1::text IS NULL
    OR account_type = $1
  )
  AND (
    $2::text IS NULL
    OR rule_severity = $2
  )
  AND (
    $3::boolean IS NULL
    OR is_active = $3
  )
ORDER BY
  rule_name
LIMIT
  $4 OFFSET $5;

-- name: CountAccountValidationRules :one
SELECT
  count(*)
FROM
  finance_account_validation_rules
WHERE
  tenant_id = current_tenant_id()
  AND (
    $1::text IS NULL
    OR account_type = $1
  )
  AND (
    $2::text IS NULL
    OR rule_severity = $2
  )
  AND (
    $3::boolean IS NULL
    OR is_active = $3
  );

-- name: GetActiveValidationRules :many
SELECT
  *
FROM
  finance_account_validation_rules
WHERE
  tenant_id = current_tenant_id()
  AND is_active = TRUE
ORDER BY
  rule_name;

-- name: GetValidationRulesByAccountType :many
SELECT
  *
FROM
  finance_account_validation_rules
WHERE
  tenant_id = current_tenant_id()
  AND account_type = $1
ORDER BY
  rule_name;

-- name: ActivateAccountValidationRule :one
UPDATE
  finance_account_validation_rules
SET
  is_active = TRUE,
  updated_at = NOW(),
  updated_by = $2
WHERE
  id = $1
  AND tenant_id = current_tenant_id()
RETURNING
  *;

-- name: DeactivateAccountValidationRule :one
UPDATE
  finance_account_validation_rules
SET
  is_active = false,
  updated_at = NOW(),
  updated_by = $2
WHERE
  id = $1
  AND tenant_id = current_tenant_id()
RETURNING
  *;
