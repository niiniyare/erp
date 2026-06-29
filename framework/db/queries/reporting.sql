-- name: CreateReport :one
INSERT INTO reports (id, tenant_id, name, entity, filters, columns, created_by)
VALUES (gen_random_uuid(), current_tenant_id(), $1, $2, $3, $4, $5)
RETURNING *;

-- name: GetReport :one
SELECT * FROM reports
WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: ListReports :many
SELECT * FROM reports
WHERE tenant_id = current_tenant_id()
ORDER BY name;

-- name: UpdateReport :one
UPDATE reports
SET name = $2, filters = $3, columns = $4, updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id()
RETURNING *;

-- name: DeleteReport :exec
DELETE FROM reports
WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: CreateReportRun :one
INSERT INTO report_runs (id, tenant_id, report_id, status, started_by)
VALUES (gen_random_uuid(), current_tenant_id(), $1, 'running', $2)
RETURNING *;

-- name: CompleteReportRun :exec
UPDATE report_runs
SET status = 'done', result_url = $2, finished_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id();
