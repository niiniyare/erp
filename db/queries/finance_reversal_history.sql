-- =====================================================================
-- FINANCE MODULE — REVERSAL HISTORY QUERIES
-- Tracks every reversal event so that:
--   (a) GetReversalHistory returns accurate data rather than fabricated UUIDs
--   (b) A reversal transaction cannot itself be reversed (double-reversal guard)
-- =====================================================================

-- name: InsertReversalHistory :exec
INSERT INTO finance_reversal_history
  (tenant_id, original_transaction_id, reversal_transaction_id, reason, reversed_by)
VALUES
  (current_tenant_id(), $1, $2, $3, $4);

-- name: IsReversalTransaction :one
-- Returns true if the given transaction_id is itself a reversal of another.
SELECT EXISTS (
  SELECT 1
  FROM   finance_reversal_history
  WHERE  tenant_id              = current_tenant_id()
    AND  reversal_transaction_id = $1
) AS is_reversal;

-- name: GetReversalHistoryByOriginal :many
SELECT id, original_transaction_id, reversal_transaction_id, reason, reversed_by, created_at
FROM   finance_reversal_history
WHERE  tenant_id              = current_tenant_id()
  AND  original_transaction_id = $1
ORDER  BY created_at ASC;
