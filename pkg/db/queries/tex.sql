-- name: ChartAccTexEntery :exec 
--  Insert tax record with rate $2
INSERT INTO tax (chart_id, rate) VALUES ($1, $2);

-- name: DeleteAccTax :exec
-- Remove tax for this chart id if id exists
DELETE FROM tax WHERE chart_id = $1;
