

-- name: GetAccountNo :one 
SELECT id FROM public.chart
WHERE 
-- tenent_id = sqlc.Arg(tenent_id)
accno = $1;

-- name: UpdateChartAccNo :exec
UPDATE chart SET
			accno = $1,
			parent_id = $2,
			description = $3,
			charttype = $4,
			gifi_accno = $5,
			category = $6,
			link = $7,
			contra = $8
		WHERE id = $9;

-- name: CreateAccount :one
		INSERT INTO chart
			(accno, parent_id, description, charttype, gifi_accno, category, link, contra)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING *;

-- name: CheckIfCharAccExcist :one
SELECT chart_id FROM tax WHERE chart_id = $1;
